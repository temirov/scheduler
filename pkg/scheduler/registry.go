package scheduler

import (
	"sync"
)

// TaskInfo contains metadata about a registered task.
type TaskInfo struct {
	ID          string
	Description string
	Schedule    TimeSchedule
}

var (
	taskRegistry = make(map[string]TaskInfo)
	registryLock = sync.RWMutex{}
)

// RegisterTaskInfo registers metadata about a task in the registry.
// Returns ErrTaskAlreadyExists if a task with the same ID already exists.
func RegisterTaskInfo(taskID, description string, schedule TimeSchedule) error {
	registryLock.Lock()
	defer registryLock.Unlock()

	if _, exists := taskRegistry[taskID]; exists {
		return ErrTaskAlreadyExists
	}

	taskRegistry[taskID] = TaskInfo{
		ID:          taskID,
		Description: description,
		Schedule:    schedule,
	}

	return nil
}

// GetTaskInfo returns information about a registered task or an error if not found.
func GetTaskInfo(taskID string) (TaskInfo, error) {
	registryLock.RLock()
	defer registryLock.RUnlock()

	taskInformation, exists := taskRegistry[taskID]
	if !exists {
		return TaskInfo{}, ErrTaskNotFound
	}
	return taskInformation, nil
}

// GetAllTaskInfo returns information about all registered tasks.
func GetAllTaskInfo() []TaskInfo {
	registryLock.RLock()
	defer registryLock.RUnlock()

	var allTaskInfos []TaskInfo
	for _, taskInformation := range taskRegistry {
		allTaskInfos = append(allTaskInfos, taskInformation)
	}
	return allTaskInfos
}

// GetRegisteredTaskIDs returns all registered task IDs.
func GetRegisteredTaskIDs() []string {
	registryLock.RLock()
	defer registryLock.RUnlock()

	var taskIDs []string
	for taskID := range taskRegistry {
		taskIDs = append(taskIDs, taskID)
	}
	return taskIDs
}

// ClearRegistryForTesting resets the task registry (only for testing purposes)
func ClearRegistryForTesting() {
	registryLock.Lock()
	defer registryLock.Unlock()

	taskRegistry = make(map[string]TaskInfo)
}
