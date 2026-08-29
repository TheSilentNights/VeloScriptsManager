package services

import (
	"sync"

	"github.com/emirpasic/gods/maps/linkedhashmap"

	"github/TheSilentNights/VeloScriptsManager/service/models"
)

// ExecutionManager keeps track of live (and recently finished) executions so
// clients can attach to their stdio pipes after they have been started.
type ExecutionManager struct {
	mu         sync.Mutex
	executions *linkedhashmap.Map
}

func NewExecutionManager() *ExecutionManager {
	return &ExecutionManager{executions: linkedhashmap.New()}
}

func (m *ExecutionManager) add(execution *models.Execution) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions.Put(execution.ID, execution)
}

func (m *ExecutionManager) get(id string) (*models.Execution, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, found := m.executions.Get(id)
	if !found {
		return nil, false
	}
	return value.(*models.Execution), true
}

func (m *ExecutionManager) remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions.Remove(id)
}

// list returns every tracked execution ordered by start time.
func (m *ExecutionManager) list() []*models.Execution {
	m.mu.Lock()
	values := m.executions.Values()
	out := make([]*models.Execution, 0, len(values))
	for _, v := range values {
		out = append(out, v.(*models.Execution))
	}
	m.mu.Unlock()
	return out
}

// killRunning force-terminates every still-running tracked process.
func (m *ExecutionManager) killRunning() {
	m.mu.Lock()
	values := m.executions.Values()
	executions := make([]*models.Execution, 0, len(values))
	for _, v := range values {
		executions = append(executions, v.(*models.Execution))
	}
	m.mu.Unlock()

	for _, e := range executions {
		if p := e.Process(); p != nil {
			_ = p.Kill()
		}
	}
}
