package tests

import (
	"context"
	"testing"
	"time"

	"github.com/temirov/scheduler/pkg/scheduler"
)

func TestSchedulerIntegration(testContext *testing.T) {
	// Create a schedule that will run soon
	scheduleTime := time.Now().Add(200 * time.Millisecond)
	testSchedule := scheduler.NewOneTimeSchedule(scheduleTime)

	// Create and configure the test task
	testTask := NewTestTask("integration-test-task", testSchedule)

	// Register the test task with the scheduler
	err := RegisterTestTaskFactory("integration-test-task", testTask)
	if err != nil {
		testContext.Fatalf("Failed to register test task: %v", err)
	}

	// Create a new scheduler instance
	schedulerInstance := scheduler.NewScheduler()

	// Get the task information and register it with the scheduler
	taskInfo, err := scheduler.GetTaskInfo("integration-test-task")
	if err != nil {
		testContext.Fatalf("Failed to get task info for registered task: %v", err)
	}

	taskFactory, exists := scheduler.GetTaskFactory("integration-test-task")
	if !exists {
		testContext.Fatalf("Failed to get task factory for registered task")
	}

	task, err := taskFactory(taskInfo)
	if err != nil {
		testContext.Fatalf("Failed to create task from factory: %v", err)
	}

	err = schedulerInstance.RegisterTask(task)
	if err != nil {
		testContext.Fatalf("Failed to register task with scheduler: %v", err)
	}

	// Start the scheduler
	schedulerInstance.Start()

	// Wait for execution with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if !testSchedule.WaitForExecution(ctx) {
		testContext.Fatalf("Task execution timed out")
	}

	// Stop the scheduler
	schedulerInstance.Stop()

	// Verify task was executed exactly once
	executionCount := testTask.GetExecutionCount()
	if executionCount != 1 {
		testContext.Errorf("Expected task to be executed once, but was executed %d times",
			executionCount)
	}

	beforeCount := testTask.GetBeforeCount()
	if beforeCount != 1 {
		testContext.Errorf("Expected BeforeExecute to be called once, but was called %d times",
			beforeCount)
	}
}
