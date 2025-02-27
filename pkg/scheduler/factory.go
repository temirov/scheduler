package scheduler

import (
	"sync"
)

// TaskFactory is a function type that creates a Task instance from a TaskInfo
type TaskFactory func(taskInfo TaskInfo) (Task, error)

var (
	taskFactories = make(map[string]TaskFactory)
	factoryLock   = sync.RWMutex{}
)

// RegisterTaskFactory registers a factory function for a specific task ID
// Returns ErrTaskAlreadyExists if a factory for this task ID is already registered
func RegisterTaskFactory(taskID string, factory TaskFactory) error {
	factoryLock.Lock()
	defer factoryLock.Unlock()

	if _, exists := taskFactories[taskID]; exists {
		return ErrTaskAlreadyExists
	}

	taskFactories[taskID] = factory
	return nil
}

// GetTaskFactory returns the factory function for a specific task ID
func GetTaskFactory(taskID string) (TaskFactory, bool) {
	factoryLock.RLock()
	defer factoryLock.RUnlock()

	factory, exists := taskFactories[taskID]
	return factory, exists
}
