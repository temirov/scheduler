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

// RegisterTask is a convenience function that registers both task information and a factory function in a single call.
// This is the recommended way to register tasks as it ensures both components are properly registered.
// The taskFactory parameter should be a function that creates a new instance of your task.
func RegisterTask(taskID string, description string, schedule TimeSchedule, taskFactory func() Task) error {
	// Register task info first
	err := RegisterTaskInfo(taskID, description, schedule)
	if err != nil {
		return err
	}

	// Then register factory function that uses the provided factory
	return RegisterTaskFactory(taskID, func(taskInfo TaskInfo) (Task, error) {
		return taskFactory(), nil
	})
}

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
