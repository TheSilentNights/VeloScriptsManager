package executor

import (
	"sync"

	"github.com/emirpasic/gods/maps/linkedhashmap"
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

func (m *ExecutionManager) Add(execution *Execution) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions.Put(execution.executionId, execution)
}

func (m *ExecutionManager) Get(id string) (*Execution, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, found := m.executions.Get(id)
	if !found {
		return nil, false
	}
	return value.(*Execution), true
}

func (m *ExecutionManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions.Remove(id)
}

// List returns every tracked execution ordered by start time.
func (m *ExecutionManager) List() []*Execution {
	m.mu.Lock()
	values := m.executions.Values()
	out := make([]*Execution, 0, len(values))
	for _, v := range values {
		out = append(out, v.(*Execution))
	}
	m.mu.Unlock()
	return out
}

// KillRunning force-terminates every still-running tracked process.
func (m *ExecutionManager) KillRunning() {
	m.mu.Lock()
	values := m.executions.Values()
	executions := make([]*Execution, 0, len(values))
	for _, v := range values {
		executions = append(executions, v.(*Execution))
	}
	m.mu.Unlock()

	for _, e := range executions {
		_ = e.Kill()
	}
}
