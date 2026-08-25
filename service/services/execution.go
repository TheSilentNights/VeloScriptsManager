package services

import (
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