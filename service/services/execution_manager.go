package services

import (
	"sort"
	"sync"

	"github/TheSilentNights/VeloScriptsManager/service/models"
)

// ExecutionManager keeps track of live (and recently finished) executions so
// clients can attach to their stdio pipes after they have been started.
type ExecutionManager struct {
	mu         sync.Mutex
	executions map[string]*models.Execution
}

func NewExecutionManager() *ExecutionManager {
	return &ExecutionManager{executions: make(map[string]*models.Execution)}
}

func (m *ExecutionManager) add(execution *models.Execution) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions[execution.ID] = execution
}

func (m *ExecutionManager) get(id string) (*models.Execution, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	execution, ok := m.executions[id]
	return execution, ok
}

func (m *ExecutionManager) remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.executions, id)
}

// list returns every tracked execution ordered by start time.
func (m *ExecutionManager) list() []*models.Execution {
	m.mu.Lock()
	out := make([]*models.Execution, 0, len(m.executions))
	for _, e := range m.executions {
		out = append(out, e)
	}
	m.mu.Unlock()

	sort.Slice(out, func(i, j int) bool {
		return out[i].StartedAt.Before(out[j].StartedAt)
	})
	return out
}

// killRunning force-terminates every still-running tracked process.
func (m *ExecutionManager) killRunning() {
	m.mu.Lock()
	executions := make([]*models.Execution, 0, len(m.executions))
	for _, e := range m.executions {
		executions = append(executions, e)
	}
	m.mu.Unlock()

	for _, e := range executions {
		if p := e.Process(); p != nil {
			_ = p.Kill()
		}
	}
}
