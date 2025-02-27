package tests

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/temirov/scheduler/pkg/scheduler"
)

func TestCLIRunCommand(t *testing.T) {
	// Save original command line arguments and restore them after the test
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	// Create a test task with a future schedule
	taskID := "test-cli-task"
	futureSchedule := scheduler.DailySchedule{Hour: 23, Minute: 59}
	testTask := NewTestTask(taskID, futureSchedule)

	// Register the task in the scheduler
	err := RegisterTestTaskFactory(taskID, testTask)
	if err != nil {
		t.Fatalf("Failed to register test task: %v", err)
	}

	// Configure CLI args to run the task immediately
	os.Args = []string{"scheduler", "--run", taskID}

	// Execute the CLI handler - we need to run this in a goroutine
	// because it calls os.Exit on success
	exitChannel := make(chan struct{})
	go func() {
		defer func() {
			// Recover from any potential panic or exit
			if r := recover(); r != nil {
				t.Logf("CLI execution recovered: %v", r)
			}
			exitChannel <- struct{}{}
		}()

		// Call the Execute function that handles CLI commands
		// Note: This would normally exit the process on success
		scheduler.Execute()
	}()

	// Wait for the task to execute or timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	success := testTask.WaitForExecution(ctx)
	if !success {
		t.Fatal("Task execution timed out")
	}

	// Wait for the CLI execution to complete
	select {
	case <-exitChannel:
		// CLI execution completed
	case <-time.After(1 * time.Second):
		t.Fatal("CLI execution did not complete in time")
	}

	// Verify the task was executed exactly once
	executionCount := testTask.GetExecutionCount()
	if executionCount != 1 {
		t.Errorf("Expected task to be executed once, but was executed %d times", executionCount)
	}

	beforeCount := testTask.GetBeforeCount()
	if beforeCount != 1 {
		t.Errorf("Expected BeforeExecute to be called once, but was called %d times", beforeCount)
	}
}
