package services

import (
	"sort"
	"sync"

	"github/TheSilentNights/VeloScriptsManager/service/models"
)

// executionManager keeps track of live (and recently finished) executions so
// clients can attach to their stdio pipes after they have been started.
type executionManager struct {
	mu         sync.Mutex
	executions map[string]*models.Execution
}

func newExecutionManager() *executionManager {
	return &executionManager{executions: make(map[string]*models.Execution)}
}

func (m *executionManager) add(execution *models.Execution) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions[execution.ID] = execution
}

func (m *executionManager) get(id string) (*models.Execution, bool) {
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

// list returns every tracked execution ordered by start time.
func (m *executionManager) list() []*models.Execution {
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
func (m *executionManager) killRunning() {
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
