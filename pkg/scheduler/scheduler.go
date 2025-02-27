package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// TimeSchedule defines the interface for scheduling tasks.
type TimeSchedule interface {
	// NextRun returns a pointer to the next execution time:
	// - nil means "never run again"
	// - A valid time indicates when to run next
	NextRun(afterTime time.Time) *time.Time
	Description() string
}

// Task defines the interface that every task must implement.
type Task interface {
	ID() string
	Schedule() TimeSchedule
	BeforeExecute(ctx context.Context) error
	Run(ctx context.Context) error
	MaxRetries() int
	RetryDelay(attempt int) time.Duration
}

// Scheduler manages the registration and execution of tasks.
type Scheduler struct {
	tasks       map[string]Task
	isRunning   bool
	stopChannel chan struct{}
	waitGroup   sync.WaitGroup
	mutex       sync.Mutex
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks:       make(map[string]Task),
		stopChannel: make(chan struct{}),
	}
}

// RegisterTask registers a new task for execution.
func (schedulerInstance *Scheduler) RegisterTask(newTask Task) error {
	schedulerInstance.mutex.Lock()
	defer schedulerInstance.mutex.Unlock()

	taskIdentifier := newTask.ID()
	if _, exists := schedulerInstance.tasks[taskIdentifier]; exists {
		return fmt.Errorf("task with ID %q already exists", taskIdentifier)
	}

	schedulerInstance.tasks[taskIdentifier] = newTask
	nextRunTime := newTask.Schedule().NextRun(time.Now())
	slog.Info("Task registered", "task_id", taskIdentifier, "schedule", newTask.Schedule().Description(), "next_run", nextRunTime)

	if schedulerInstance.isRunning {
		schedulerInstance.waitGroup.Add(1)
		go schedulerInstance.executeTask(newTask)
	}

	return nil
}

// Start begins execution of all registered tasks.
func (schedulerInstance *Scheduler) Start() {
	schedulerInstance.mutex.Lock()
	defer schedulerInstance.mutex.Unlock()

	if schedulerInstance.isRunning {
		return
	}

	schedulerInstance.isRunning = true
	schedulerInstance.stopChannel = make(chan struct{})

	for _, taskInstance := range schedulerInstance.tasks {
		schedulerInstance.waitGroup.Add(1)
		go schedulerInstance.executeTask(taskInstance)
	}

	slog.Info("Scheduler started", "task_count", len(schedulerInstance.tasks))
}

// Stop signals the scheduler to stop and waits for running tasks to complete.
func (schedulerInstance *Scheduler) Stop() {
	schedulerInstance.mutex.Lock()
	if !schedulerInstance.isRunning {
		schedulerInstance.mutex.Unlock()
		return
	}
	schedulerInstance.isRunning = false
	close(schedulerInstance.stopChannel)
	schedulerInstance.mutex.Unlock()

	schedulerInstance.waitGroup.Wait()
	slog.Info("Scheduler stopped")
}

// executeTask runs the task according to its schedule
func (schedulerInstance *Scheduler) executeTask(taskInstance Task) {
	defer schedulerInstance.waitGroup.Done()

	for {
		currentTime := time.Now()
		nextRunTimePtr := taskInstance.Schedule().NextRun(currentTime)

		// If nil returned, the task should never run again
		if nextRunTimePtr == nil {
			slog.Info("Task will not run again", "task_id", taskInstance.ID())
			return
		}

		nextRunTime := *nextRunTimePtr
		waitDuration := nextRunTime.Sub(currentTime)

		// Use a timer to wait until next execution
		timer := time.NewTimer(waitDuration)

		select {
		case <-timer.C:
			// Time to execute the task
			ctx := context.Background()
			slog.Info("Executing task", "task_id", taskInstance.ID())
			contextBefore, cancelBefore := context.WithTimeout(ctx, 30*time.Minute)
			executionBeforeError := taskInstance.BeforeExecute(contextBefore)
			cancelBefore()

			if executionBeforeError != nil {
				slog.Error("Task BeforeExecute failed", "task_id", taskInstance.ID(), "error", executionBeforeError)
			} else {
				var retryAttempt int
				maximumRetries := taskInstance.MaxRetries()

				for retryAttempt = 0; retryAttempt <= maximumRetries; retryAttempt++ {
					if retryAttempt > 0 {
						slog.Info("Retrying task after failure", "task_id", taskInstance.ID(), "attempt", retryAttempt, "max_retries", maximumRetries)
					}

					contextRun, cancelRun := context.WithTimeout(ctx, 30*time.Minute)
					executionRunError := taskInstance.Run(contextRun)
					cancelRun()

					if executionRunError == nil {
						break
					}

					if retryAttempt < maximumRetries {
						retryDelayDuration := taskInstance.RetryDelay(retryAttempt)
						slog.Warn("Task failed, retrying", "task_id", taskInstance.ID(), "attempt", retryAttempt+1, "max_retries", maximumRetries, "retry_delay", retryDelayDuration, "error", executionRunError)

						select {
						case <-time.After(retryDelayDuration):
						case <-schedulerInstance.stopChannel:
							slog.Info("Task retry cancelled due to scheduler stopping", "task_id", taskInstance.ID())
							return
						}
					} else {
						executionBeforeError = executionRunError
					}
				}

				if executionBeforeError != nil {
					slog.Error("Task failed after retries", "task_id", taskInstance.ID(), "attempts", retryAttempt, "error", executionBeforeError)
				}
			}

			// After executing a task with a OneTimeSchedule
			if oneTimeSchedule, isOneTime := taskInstance.Schedule().(*OneTimeSchedule); isOneTime {
				oneTimeSchedule.SignalExecution()
			}

		case <-schedulerInstance.stopChannel:
			// Scheduler is stopping
			timer.Stop()
			slog.Info("Task stopped", "task_id", taskInstance.ID())
			return
		}
	}
}

// RunTaskNow executes a task immediately.
func (schedulerInstance *Scheduler) RunTaskNow(taskIdentifier string) error {
	schedulerInstance.mutex.Lock()
	taskInstance, exists := schedulerInstance.tasks[taskIdentifier]
	schedulerInstance.mutex.Unlock()

	if !exists {
		return fmt.Errorf("task not found")
	}

	slog.Info("Executing task on demand", "task_id", taskIdentifier)

	contextBefore, cancelBefore := context.WithTimeout(context.Background(), 30*time.Minute)
	executionBeforeError := taskInstance.BeforeExecute(contextBefore)
	cancelBefore()
	if executionBeforeError != nil {
		slog.Error("Task BeforeExecute failed", "task_id", taskIdentifier, "error", executionBeforeError)
		return executionBeforeError
	}

	var retryAttempt int
	maximumRetries := taskInstance.MaxRetries()

	for retryAttempt = 0; retryAttempt <= maximumRetries; retryAttempt++ {
		if retryAttempt > 0 {
			slog.Info("Retrying on-demand task after failure", "task_id", taskIdentifier, "attempt", retryAttempt, "max_retries", maximumRetries)
		}

		contextRun, cancelRun := context.WithTimeout(context.Background(), 30*time.Minute)
		executionRunError := taskInstance.Run(contextRun)
		cancelRun()

		if executionRunError == nil {
			if retryAttempt > 0 {
				slog.Info("On-demand task succeeded after retry", "task_id", taskIdentifier, "attempts", retryAttempt+1)
			} else {
				slog.Info("On-demand task executed successfully", "task_id", taskIdentifier)
			}
			return nil
		}

		if retryAttempt < maximumRetries {
			retryDelayDuration := taskInstance.RetryDelay(retryAttempt)
			slog.Warn("On-demand task failed, retrying", "task_id", taskIdentifier, "attempt", retryAttempt+1, "max_retries", maximumRetries, "retry_delay", retryDelayDuration, "error", executionRunError)
			time.Sleep(retryDelayDuration)
		} else {
			executionBeforeError = executionRunError
		}
	}

	slog.Error("On-demand task execution failed after retries", "task_id", taskIdentifier, "attempts", retryAttempt, "error", executionBeforeError)
	return executionBeforeError
}
