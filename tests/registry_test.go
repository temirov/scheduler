package tests

import (
	"testing"
	"time"

	"github.com/temirov/scheduler/pkg/scheduler"
)

func TestTaskRegistry(testContext *testing.T) {
	// Clear the registry before testing
	clearRegistry()

	// Create some sample schedules
	dailySchedule := scheduler.DailySchedule{Hour: 8, Minute: 30}
	intervalSchedule := scheduler.IntervalSchedule{Interval: 2 * time.Hour}

	// Test 1: Register tasks and verify they're added
	scheduler.RegisterTaskInfo("task1", "First test task", dailySchedule)
	scheduler.RegisterTaskInfo("task2", "Second test task", intervalSchedule)

	// Test 2: Get task info for a specific task
	taskInfo, err := scheduler.GetTaskInfo("task1")
	if err != nil {
		testContext.Errorf("Expected task1 to exist in registry, but got error: %v", err)
	}
	if taskInfo.ID != "task1" || taskInfo.Description != "First test task" {
		testContext.Errorf("Task info mismatch. Got %+v", taskInfo)
	}

	// Test 3: Verify non-existent task returns error
	_, err = scheduler.GetTaskInfo("non-existent-task")
	if err != scheduler.ErrTaskNotFound {
		testContext.Errorf("Expected ErrTaskNotFound for non-existent task, got: %v", err)
	}

	// Test 4: Get all task info
	allTasks := scheduler.GetAllTaskInfo()
	if len(allTasks) != 2 {
		testContext.Errorf("Expected 2 tasks, got %d", len(allTasks))
	}

	// Test 5: Get all task IDs
	taskIDs := scheduler.GetRegisteredTaskIDs()
	if len(taskIDs) != 2 {
		testContext.Errorf("Expected 2 task IDs, got %d", len(taskIDs))
	}

	// Verify that all expected task IDs are present
	foundTask1 := false
	foundTask2 := false
	for _, identifier := range taskIDs {
		if identifier == "task1" {
			foundTask1 = true
		} else if identifier == "task2" {
			foundTask2 = true
		}
	}

	if !foundTask1 || !foundTask2 {
		testContext.Errorf("Did not find all expected task IDs. Found task1: %v, Found task2: %v",
			foundTask1, foundTask2)
	}
}

// Helper function to clear the registry between test runs
func clearRegistry() {
	scheduler.ClearRegistryForTesting()
}

func TestRegisterTaskInfo_ErrorHandling(t *testing.T) {
	// Clear the registry before testing
	clearRegistry()

	// First registration should succeed
	taskID := "test-task"
	description := "Test task description"
	schedule := scheduler.DailySchedule{Hour: 12, Minute: 0}

	err := scheduler.RegisterTaskInfo(taskID, description, schedule)
	if err != nil {
		t.Errorf("First task registration should succeed, got error: %v", err)
	}

	// Attempting to register same ID again should fail
	err = scheduler.RegisterTaskInfo(taskID, "Different description", schedule)
	if err != scheduler.ErrTaskAlreadyExists {
		t.Errorf("Expected ErrTaskAlreadyExists but got: %v", err)
	}
}

func TestGetTaskInfo_ErrorHandling(t *testing.T) {
	// Clear the registry before testing
	clearRegistry()

	// Add a test task
	taskID := "existing-task"
	description := "Existing task description"
	schedule := scheduler.DailySchedule{Hour: 15, Minute: 30}

	err := scheduler.RegisterTaskInfo(taskID, description, schedule)
	if err != nil {
		t.Fatalf("Failed to register test task: %v", err)
	}

	// Test successful retrieval
	taskInfo, err := scheduler.GetTaskInfo(taskID)
	if err != nil {
		t.Errorf("Expected no error for existing task, got: %v", err)
	}
	if taskInfo.ID != taskID || taskInfo.Description != description {
		t.Errorf("Retrieved task info doesn't match registered task")
	}

	// Test retrieval of non-existent task
	nonExistentID := "non-existent-task"
	_, err = scheduler.GetTaskInfo(nonExistentID)
	if err != scheduler.ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound for non-existent task, got: %v", err)
	}
}

func TestRegisterTask(t *testing.T) {
	// Clear the registry before testing
	clearRegistry()

	// Define test variables
	taskID := "combined-registration-task"
	description := "Task registered with combined function"
	schedule := scheduler.DailySchedule{Hour: 10, Minute: 15}

	// Create a test task using the existing TestTask implementation
	testTask := NewTestTask(taskID, schedule)

	// Register using the new combined function
	err := scheduler.RegisterTask(taskID, description, schedule, func() scheduler.Task {
		return testTask
	})
	if err != nil {
		t.Errorf("Task registration with combined function failed: %v", err)
	}

	// Verify task info was registered correctly
	taskInfo, err := scheduler.GetTaskInfo(taskID)
	if err != nil {
		t.Errorf("Failed to get task info after combined registration: %v", err)
	}
	if taskInfo.ID != taskID || taskInfo.Description != description {
		t.Errorf("Task info mismatch after combined registration. Got %+v", taskInfo)
	}

	// Verify factory was registered correctly
	factory, exists := scheduler.GetTaskFactory(taskID)
	if !exists {
		t.Errorf("Task factory not found after combined registration")
	}

	// Verify factory creates the correct task
	createdTask, err := factory(taskInfo)
	if err != nil {
		t.Errorf("Error creating task from factory: %v", err)
	}
	if createdTask.ID() != taskID {
		t.Errorf("Created task has wrong ID. Expected %s, got %s", taskID, createdTask.ID())
	}

	// Attempting to register same ID again should fail
	err = scheduler.RegisterTask(taskID, "Different description", schedule, func() scheduler.Task {
		return testTask
	})
	if err != scheduler.ErrTaskAlreadyExists {
		t.Errorf("Expected ErrTaskAlreadyExists but got: %v", err)
	}
}
