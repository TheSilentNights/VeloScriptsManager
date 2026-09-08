package executor

import (
	"context"
	"errors"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/logs"
	"github/TheSilentNights/VeloScriptsManager/service/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emirpasic/gods/maps/linkedhashmap"
	"go.uber.org/zap"
)

type ScriptInfo struct {
	ScriptID              string    `json:"scriptId"`
	Name                  string    `json:"name"`
	WorkDir               string    `json:"workdir"`
	StartedAt             time.Time `json:"startedAt"`
	Command               []string  `json:"command"`               // params used for this run
	EnvironmentsFlattened []string  `json:"environmentsFlattened"` // environment variables applied for this run
}

// Execution tracks one asynchronously running script process.
type Execution struct {
	executionId string
	scriptInfo  *ScriptInfo

	doFinishOnce sync.Once
	isKilled     bool

	mu sync.Mutex

	cmd      *exec.Cmd
	exitErr  string
	exitCode int
	status   string // running | finished | failed | prepare | killed
}

func NewExecution(scriptID string, name string, command []string, workDir string, environments []string) *Execution {
	return &Execution{
		executionId: utils.GenerateExecutionId(),
		scriptInfo: &ScriptInfo{
			ScriptID:              scriptID,
			Name:                  name,
			StartedAt:             time.Now(),
			WorkDir:               workDir,
			Command:               command,
			EnvironmentsFlattened: environments,
		},
		status:   "prepare",
		exitCode: -1,
	}
}

func (execution *Execution) Start(ctx context.Context) error {
	logs.Logger.Info("launched execution",
		zap.String("scriptId", execution.scriptInfo.ScriptID),
		zap.String("name", execution.scriptInfo.Name),
		zap.String("workdir", execution.scriptInfo.WorkDir),
		zap.Time("startedAt", execution.scriptInfo.StartedAt),
		zap.Strings("command", execution.scriptInfo.Command),
		zap.Strings("environments", execution.scriptInfo.EnvironmentsFlattened),
	)

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

	if len(execution.scriptInfo.EnvironmentsFlattened) > 0 {
		cmd.Env = mergeEnviron(os.Environ(), execution.scriptInfo.EnvironmentsFlattened)
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

func (execution *Execution) doFinish(exitCode int, status string, err error, fromKill bool) {
	execution.mu.Lock()
	defer execution.mu.Unlock()

	execution.doFinishOnce.Do(func() {
		if !fromKill && execution.isKilled {
			return
		}
		execution.status = status
		execution.exitCode = exitCode
		if err != nil {
			execution.exitErr = err.Error()
		}
		execution.cmd = nil

		logs.Logger.Info("execution finished",
			zap.String("scriptId", execution.scriptInfo.ScriptID),
			zap.String("name", execution.scriptInfo.Name),
			zap.Int("exitCode", exitCode),
			zap.String("exitErr", execution.exitErr),
			zap.String("status", status),
		)
	})

}

func (execution *Execution) Kill() error {
	execution.mu.Lock()

	if execution.status != "running" {
		execution.mu.Unlock()
		return ierrors.ExecutionNotRunningError
	}

	if execution.cmd == nil || execution.cmd.Process == nil {
		execution.mu.Unlock()
		return errors.New("process not started")
	}

	err := exec.Command("taskkill", "/PID", strconv.Itoa(execution.cmd.Process.Pid), "/T", "/F").Run()
	execution.mu.Unlock()

	if err != nil {
		execution.doFinish(-1, "failed", err, true)
		return err
	}

	execution.doFinish(-1, "killed", err, true)

	return err
}

func (execution *Execution) finish(err error) {
	var exitCode int
	var status string

	if err == nil {
		exitCode = 0
		status = "finished"
	} else {
		var exitErr *exec.ExitError
		exitCode = -1
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		status = "failed"
	}

	execution.doFinish(exitCode, status, err, false)

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
	return execution.exitErr
}
