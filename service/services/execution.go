package services

import (
	"sync"
	"time"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
)

// Execution tracks one asynchronously running script process.
type Execution struct {
	ID        string            `json:"executionId"`
	ScriptID  string            `json:"scriptId"`
	Name      string            `json:"name"`
	Status    string            `json:"status"` // running | finished | failed
	ExitCode  int               `json:"exitCode"`
	Error     string            `json:"error,omitempty"`
	StartedAt time.Time         `json:"startedAt"`
	Process   *executor.Process `json:"-"`
}

// executionManager keeps track of live (and recently finished) executions so
// clients can attach to their stdio pipes after they have been started.
type executionManager struct {
	mu         sync.Mutex
	executions map[string]*Execution
}

func newExecutionManager() *executionManager {
	return &executionManager{executions: make(map[string]*Execution)}
}

func (m *executionManager) add(execution *Execution) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions[execution.ID] = execution
}

func (m *executionManager) get(id string) (*Execution, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	execution, ok := m.executions[id]
	return execution, ok
}

func (m *executionManager) remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.executions, id)
}
