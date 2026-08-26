package models

import (
	"sync"
	"time"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
)

// Execution tracks one asynchronously running script process.
type Execution struct {
	ID        string    `json:"executionId"`
	ScriptID  string    `json:"scriptId"`
	Name      string    `json:"name"`
	StartedAt time.Time `json:"startedAt"`

	mu       sync.Mutex
	status   string // running | finished | failed
	exitCode int
	err      string
	process  *executor.Process
}

func NewExecution(id, scriptID, name string, process *executor.Process) *Execution {
	return &Execution{
		ID:        id,
		ScriptID:  scriptID,
		Name:      name,
		StartedAt: time.Now(),
		status:    "running",
		exitCode:  -1,
		process:   process,
	}
}

func (e *Execution) Status() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.status
}

func (e *Execution) ExitCode() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.exitCode
}

func (e *Execution) Error() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.err
}

func (e *Execution) Process() *executor.Process {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.process
}

func (e *Execution) Finish(exitCode int, runErr error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.exitCode = exitCode
	if runErr != nil {
		e.status = "failed"
		e.err = runErr.Error()
	} else {
		e.status = "finished"
	}
	// 抛弃 process 句柄，让 GC 回收其输出缓冲，防止内存膨胀。
	e.process = nil
}

// ExecutionStatusInfo is the API projection of a tracked execution returned by
// the executions listing endpoint.
type ExecutionStatusInfo struct {
	ExecutionId string    `json:"executionId"`
	ScriptId    string    `json:"scriptId"`
	Name        string    `json:"name"`
	StartedAt   time.Time `json:"startedAt"`
	Status      string    `json:"status"`   // running | finished | failed
	ExitCode    int       `json:"exitCode"` // -1 while still running
	Error       string    `json:"error"`
}
