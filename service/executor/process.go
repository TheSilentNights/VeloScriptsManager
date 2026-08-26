package executor

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const (
	// subscriberBufferSize is the per-subscriber outbound channel buffer.
	subscriberBufferSize = 512

	// replayHeadroom is the fraction of the subscriber buffer reserved for
	// live output after a mid-stream attach replays history into it.
	replayHeadroom = 2 // replay fills at most buffer/replayHeadroom slots

	// maxBufferedOutput caps how much output history a Process keeps in memory.
	maxBufferedOutput = 10 << 20 // 10 MiB
)

// Chunk is one unit of process output delivered to subscribers. Dropped > 0
// signals that Dropped earlier chunks were omitted right before Data because
// the subscriber fell behind; consumers should surface that gap to clients.
type Chunk struct {
	Data    []byte
	Dropped int
}

type subscriber struct {
	ch     chan Chunk
	missed int
}

// Process is a running command with exposed stdio pipes. stdout and stderr are
// merged into a single output stream, in arrival order, which is accumulated
// (capped) and broadcast to every subscriber, so multiple clients can attach
// to the same process simultaneously.
type Process struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser

	stdinMu    sync.Mutex
	mu         sync.Mutex
	chunks     [][]byte
	totalBytes int
	subs       []*subscriber
	closed     bool

	done     chan struct{}
	exitErr  error
	exitCode int
}

// Start launches the command
func Start(ctx context.Context, runner string, params []string, workDir string, env []string) (*Process, error) {
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
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	//init pipe
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

// Subscribe registers a new output consumer.
func (p *Process) Subscribe() <-chan Chunk {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan Chunk, subscriberBufferSize)
	if p.closed {
		close(ch)
		return ch
	}

	// Replay at most half of the buffer as history so freshly attached
	// consumers keep headroom for live output before they start reading.
	start := 0
	history := len(p.chunks)
	if limit := subscriberBufferSize / replayHeadroom; history > limit {
		start = history - limit
	}

	for i, chunk := range p.chunks[start:] {
		item := Chunk{Data: chunk}
		if i == 0 && start > 0 {
			item.Dropped = start
		}
		ch <- item
	}

	p.subs = append(p.subs, &subscriber{ch: ch})
	return ch
}

// WriteStdin writes data into the process stdin.
func (p *Process) WriteStdin(data []byte) (int, error) {
	p.stdinMu.Lock()
	defer p.stdinMu.Unlock()
	return p.stdin.Write(data)
}

// CloseStdin closes the process stdin, letting a reader detect EOF.
func (p *Process) CloseStdin() error {
	p.stdinMu.Lock()
	defer p.stdinMu.Unlock()
	return p.stdin.Close()
}

// Kill force-terminates the process.
func (p *Process) Kill() error {
	if p.cmd.Process == nil {
		return errors.New("process not started")
	}
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

// broadcast appends the chunk to the capped history and fans it out to every
// subscriber. A subscriber whose buffer is full misses chunks; the next chunk
// it accepts is preceded by a gap marker carrying how many were skipped.
func (p *Process) broadcast(chunk []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.chunks = append(p.chunks, chunk)
	p.totalBytes += len(chunk)
	for len(p.chunks) > 0 && p.totalBytes > maxBufferedOutput {
		//remove the earliest chunk
		p.totalBytes -= len(p.chunks[0])
		p.chunks = p.chunks[1:]
	}

	for _, s := range p.subs {
		if s.missed > 0 {
			select {
			case s.ch <- Chunk{Dropped: s.missed}:
				s.missed = 0
			default:
				// Marker did not fit; keep accumulating until it does.
			}
		}
		select {
		case s.ch <- Chunk{Data: chunk}:
		default:
			s.missed++
		}
	}
}

func (p *Process) wait(wg *sync.WaitGroup) {
	wg.Wait()

	err := p.cmd.Wait()

	p.mu.Lock()
	p.exitErr = err
	if err == nil {
		p.exitCode = 0
	} else {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			p.exitCode = exitErr.ExitCode()
		}
	}
	p.closed = true
	subs := p.subs
	p.subs = nil
	p.mu.Unlock()

	for _, s := range subs {
		close(s.ch)
	}
	close(p.done)
}
