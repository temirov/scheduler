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
func RegisterTaskInfo(taskID, description string, schedule TimeSchedule) {
	registryLock.Lock()
	defer registryLock.Unlock()

	taskRegistry[taskID] = TaskInfo{
		ID:          taskID,
		Description: description,
		Schedule:    schedule,
	}
}

// GetTaskInfo returns information about a registered task.
func GetTaskInfo(taskID string) (TaskInfo, bool) {
	registryLock.RLock()
	defer registryLock.RUnlock()

	taskInformation, exists := taskRegistry[taskID]
	return taskInformation, exists
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
