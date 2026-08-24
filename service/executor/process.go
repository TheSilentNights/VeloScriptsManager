package executor

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"sync"
)

const (
	// subscriberBufferSize is the per-subscriber outbound channel buffer.
	subscriberBufferSize = 512

	// maxBufferedOutput caps how much output history a Process keeps in memory.
	maxBufferedOutput = 10 << 20 // 10 MiB
)

// Process is a running command with exposed stdio pipes. stdout and stderr are
// merged into a single output stream, in arrival order, which is accumulated
// (capped) and broadcast to every subscriber, so multiple clients can attach
// to the same process simultaneously.
type Process struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser

	mu         sync.Mutex
	chunks     [][]byte
	totalBytes int
	subs       []chan []byte
	closed     bool

	done     chan struct{}
	exitErr  error
	exitCode int
}

// Start launches the command described by runner+params inside workDir without
// blocking and returns a Process handle for interacting with it.
//
// The runner/params resolution follows Exec: when runner is empty the first
// param is treated as the executable and the rest as its arguments.
func Start(ctx context.Context, runner string, params []string, workDir string) (*Process, error) {
	if len(params) == 0 {
		return nil, errors.New("empty command")
	}

	exe, args := runner, params
	if exe == "" {
		exe = params[0]
		args = params[1:]
	}

	cmd := exec.CommandContext(ctx, exe, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	p := &Process{
		cmd:      cmd,
		stdin:    stdin,
		done:     make(chan struct{}),
		exitCode: -1,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go p.pump(stdout, &wg)
	go p.pump(stderr, &wg)
	go p.wait(&wg)

	return p, nil
}

// Subscribe registers a new output consumer. Every output chunk produced so
// far is replayed first (in order), followed by live chunks. The returned
// channel is closed once the process has exited. Replay happens while holding
// the internal lock, so no chunk can be missed or duplicated.
func (p *Process) Subscribe() <-chan []byte {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan []byte, subscriberBufferSize)
	if p.closed {
		close(ch)
		return ch
	}

	for _, chunk := range p.chunks {
		select {
		case ch <- chunk:
		default:
			// History bigger than the channel buffer: drop the oldest replay
			// chunks rather than blocking the process output pump.
		}
	}

	p.subs = append(p.subs, ch)
	return ch
}

// WriteStdin writes data into the process stdin.
func (p *Process) WriteStdin(data []byte) (int, error) {
	return p.stdin.Write(data)
}

// CloseStdin closes the process stdin, letting a reader detect EOF.
func (p *Process) CloseStdin() error {
	return p.stdin.Close()
}

// Kill force-terminates the process.
func (p *Process) Kill() error {
	return p.cmd.Process.Kill()
}

// Done is closed once the process has exited and all output has been flushed.
func (p *Process) Done() <-chan struct{} {
	return p.done
}

// ExitCode returns the process exit code, or -1 while it is still running.
func (p *Process) ExitCode() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.exitCode
}

// Err returns the error reported by cmd.Wait, or nil on a clean exit.
func (p *Process) Err() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.exitErr
}

// FullOutput returns all output captured so far as a string.
func (p *Process) FullOutput() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var b strings.Builder
	for _, chunk := range p.chunks {
		b.Write(chunk)
	}
	return b.String()
}

func (p *Process) pump(r io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			p.broadcast(chunk)
		}
		if err != nil {
			return
		}
	}
}

func (p *Process) broadcast(chunk []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.chunks = append(p.chunks, chunk)
	p.totalBytes += len(chunk)
	for len(p.chunks) > 0 && p.totalBytes > maxBufferedOutput {
		p.totalBytes -= len(p.chunks[0])
		p.chunks = p.chunks[1:]
	}

	for _, ch := range p.subs {
		select {
		case ch <- chunk:
		default:
			// Slow subscriber: drop the chunk for that subscriber only.
		}
	}
}

func (p *Process) wait(wg *sync.WaitGroup) {
	wg.Wait()

	err := p.cmd.Wait()

	p.mu.Lock()
	p.exitErr = err
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		p.exitCode = exitErr.ExitCode()
	}
	p.closed = true
	subs := p.subs
	p.subs = nil
	p.mu.Unlock()

	for _, ch := range subs {
		close(ch)
	}
	close(p.done)
}
