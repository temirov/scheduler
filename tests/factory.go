package tests

import (
	"github.com/temirov/scheduler/pkg/scheduler"
)

// RegisterTestTaskFactory registers a factory for the test task
func RegisterTestTaskFactory(taskID string, testTask *TestTask) error {
	// Register task info
	err := scheduler.RegisterTaskInfo(
		taskID,
		"Test task for CLI integration testing",
		testTask.Schedule(),
	)
	if err != nil {
		return err
	}

	// Register factory for the task
	return scheduler.RegisterTaskFactory(taskID, func(taskInfo scheduler.TaskInfo) (scheduler.Task, error) {
		return testTask, nil
	})
} 