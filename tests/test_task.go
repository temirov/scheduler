package tests

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/temirov/scheduler/pkg/scheduler"
)

// TestTask implements the scheduler.Task interface for testing
type TestTask struct {
	taskIdentifier   string
	taskSchedule     scheduler.TimeSchedule
	executionCount   atomic.Int32
	beforeCount      atomic.Int32
	maximumRetries   int
	retryDelayAmount time.Duration
}

// NewTestTask creates a new test task for integration testing
func NewTestTask(identifier string, schedule scheduler.TimeSchedule) *TestTask {
	return &TestTask{
		taskIdentifier:   identifier,
		taskSchedule:     schedule,
		maximumRetries:   1,
		retryDelayAmount: time.Millisecond * 100,
	}
}

// ID returns the task identifier
func (testTask *TestTask) ID() string {
	return testTask.taskIdentifier
}

// Schedule returns the task schedule
func (testTask *TestTask) Schedule() scheduler.TimeSchedule {
	return testTask.taskSchedule
}

// BeforeExecute implements the pre-execution hook
func (testTask *TestTask) BeforeExecute(ctx context.Context) error {
	testTask.beforeCount.Add(1)
	return nil
}

// Run executes the task
func (testTask *TestTask) Run(ctx context.Context) error {
	testTask.executionCount.Add(1)
	return nil
}

// MaxRetries returns the maximum number of retry attempts
func (testTask *TestTask) MaxRetries() int {
	return testTask.maximumRetries
}

// RetryDelay returns the delay between retry attempts
func (testTask *TestTask) RetryDelay(attempt int) time.Duration {
	return testTask.retryDelayAmount
}

// GetExecutionCount returns the number of times the task has been executed
func (testTask *TestTask) GetExecutionCount() int32 {
	return testTask.executionCount.Load()
}

// GetBeforeCount returns the number of times BeforeExecute has been called
func (testTask *TestTask) GetBeforeCount() int32 {
	return testTask.beforeCount.Load()
}

// WaitForExecution waits until the task is executed or until the context expires
func (testTask *TestTask) WaitForExecution(ctx context.Context) bool {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if testTask.executionCount.Load() > 0 {
				return true
			}
		}
	}
}
