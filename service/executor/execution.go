package executor

import (
	"context"
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emirpasic/gods/maps/linkedhashmap"
)

type ScriptInfo struct {
	ScriptID     string    `json:"scriptId"`
	Name         string    `json:"name"`
	WorkDir      string    `json:"workdir"`
	StartedAt    time.Time `json:"startedAt"`
	Command      []string  `json:"command"`      // params used for this run
	Environments []string  `json:"environments"` // environment ids applied for this run
}

// Execution tracks one asynchronously running script process.
type Execution struct {
	executionId string
	scriptInfo  *ScriptInfo

	mu               sync.Mutex
	closeChannelOnce sync.Once
	doneChanSignal   chan struct{}

	cmd      *exec.Cmd
	exitErr  error
	exitCode int
	status   string // running | finished | failed | prepare | killed
}

func NewExecution(scriptID string, name string, command []string, workDir string, environments []string) *Execution {
	return &Execution{
		executionId: utils.GenerateExecutionId(),
		scriptInfo: &ScriptInfo{
			ScriptID:     scriptID,
			Name:         name,
			StartedAt:    time.Now(),
			WorkDir:      workDir,
			Command:      command,
			Environments: environments,
		},
		doneChanSignal: make(chan struct{}),
		status:         "prepare",
		exitCode:       -1,
	}
}

func (execution *Execution) Start(ctx context.Context) error {
	if len(execution.scriptInfo.Command) == 0 {
		return errors.New("empty command")
	}

	//修复将标题识别为可执行程序的问题
	title := strings.ReplaceAll(execution.scriptInfo.Name, `"`, `'`)
	if !strings.ContainsAny(title, " \t") {
		title += " "
	}

	launchArgs := []string{"/c", "start", title, "/wait"}
	launchArgs = append(launchArgs, execution.scriptInfo.Command...)
	cmd := exec.CommandContext(ctx, "cmd.exe", launchArgs...)
	if execution.scriptInfo.WorkDir != "" {
		cmd.Dir = execution.scriptInfo.WorkDir
	}

	if len(execution.scriptInfo.Environments) > 0 {
		cmd.Env = mergeEnviron(os.Environ(), execution.scriptInfo.Environments)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	execution.cmd = cmd
	execution.status = "running"

	go execution.waitForExit()

	return nil
}

func (execution *Execution) waitForExit() {
	err := execution.cmd.Wait()

	execution.finish(err)
}

func (execution *Execution) Kill() error {
	execution.mu.Lock()
	if execution.cmd == nil || execution.cmd.Process == nil {
		return errors.New("process not started")
	}
	err := exec.Command("taskkill", "/PID", strconv.Itoa(execution.cmd.Process.Pid), "/T", "/F").Run()
	execution.mu.Unlock()

	if err == nil {
		execution.mu.Lock()
		defer execution.mu.Unlock()

		execution.closeChannelOnce.Do(func() {
			close(execution.doneChanSignal)

			execution.status = "killed"
			execution.exitCode = -1
		})

	}

	return err
}

func (execution *Execution) finish(err error) {
	execution.mu.Lock()
	defer execution.mu.Unlock()

	execution.exitErr = err
	if err == nil {
		execution.exitCode = 0
	} else {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			execution.exitCode = exitErr.ExitCode()
		}
	}

	execution.closeChannelOnce.Do(func() {
		close(execution.doneChanSignal)

		//set status
		if execution.exitErr != nil {
			execution.status = "failed"
		} else {
			execution.status = "finished"
		}

		execution.cmd = nil
	})

}

type envEntry struct {
	key   string
	value string
}

func mergeEnviron(base, overrides []string) []string {
	entries := linkedhashmap.New()

	put := func(key, value string) {
		lower := strings.ToLower(key)
		if existing, found := entries.Get(lower); found {
			key = existing.(envEntry).key
		}
		entries.Put(lower, envEntry{key: key, value: value})
	}

	for _, kv := range base {
		key, value, _ := strings.Cut(kv, "=")
		put(key, value)
	}

	for _, kv := range overrides {
		key, value, _ := strings.Cut(kv, "=")
		lower := strings.ToLower(key)
		if lower == "path" {
			if existing, found := entries.Get(lower); found {
				entry := existing.(envEntry)
				switch {
				case entry.value == "":
					entry.value = value
				case value == "":
				default:
					entry.value = value + ";" + entry.value
				}
				entries.Put(lower, entry)
				continue
			}
		}
		put(key, value)
	}

	keys := entries.Keys()
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		raw, _ := entries.Get(k)
		entry := raw.(envEntry)
		out = append(out, entry.key+"="+entry.value)
	}
	return out
}

func (execution *Execution) GetExecutionId() string {
	return execution.executionId
}

func (execution *Execution) GetScriptInfo() *ScriptInfo {
	return execution.scriptInfo
}

func (execution *Execution) GetStatus() string {
	execution.mu.Lock()
	defer execution.mu.Unlock()
	return execution.status
}

func (execution *Execution) GetExitCode() int {
	execution.mu.Lock()
	defer execution.mu.Unlock()
	return execution.exitCode
}

func (execution *Execution) GetError() string {
	execution.mu.Lock()
	defer execution.mu.Unlock()
	return execution.exitErr.Error()
}
