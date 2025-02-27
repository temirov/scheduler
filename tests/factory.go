package tests

import (
	"github.com/temirov/scheduler/pkg/scheduler"
)

// RegisterTestTaskFactory registers a factory for the test task
func RegisterTestTaskFactory(taskID string, testTask *TestTask) error {
	// Register the task using the new simplified API
	return scheduler.RegisterTask(
		taskID,
		"Test task for CLI integration testing",
		testTask.Schedule(),
		func() scheduler.Task {
			return testTask
		},
	)
}
