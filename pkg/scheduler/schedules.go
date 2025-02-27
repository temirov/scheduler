package scheduler

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// DailySchedule runs a task at the same time every day.
type DailySchedule struct {
	Hour   int
	Minute int
}

// NextRun returns the next run time after the provided time.
func (daily DailySchedule) NextRun(after time.Time) *time.Time {
	nextRunTime := time.Date(
		after.Year(), after.Month(), after.Day(),
		daily.Hour, daily.Minute, 0, 0, after.Location(),
	)
	if nextRunTime.Before(after) {
		nextRunTime = nextRunTime.AddDate(0, 0, 1)
	}
	return &nextRunTime
}

// Description returns a description of the daily schedule.
func (daily DailySchedule) Description() string {
	return fmt.Sprintf("Daily at %02d:%02d", daily.Hour, daily.Minute)
}

// WeekdaySchedule runs a task on specified weekdays at a given time.
type WeekdaySchedule struct {
	Weekdays []time.Weekday
	Hour     int
	Minute   int
}

// NextRun returns the next run time after the provided time.
func (weekday WeekdaySchedule) NextRun(after time.Time) *time.Time {
	nextRunTime := time.Date(
		after.Year(), after.Month(), after.Day(),
		weekday.Hour, weekday.Minute, 0, 0, after.Location(),
	)
	if nextRunTime.Before(after) {
		nextRunTime = nextRunTime.AddDate(0, 0, 1)
	}
	for !containsWeekday(weekday.Weekdays, nextRunTime.Weekday()) {
		nextRunTime = nextRunTime.AddDate(0, 0, 1)
	}
	return &nextRunTime
}

// Description returns a description of the weekday schedule.
func (weekday WeekdaySchedule) Description() string {
	return fmt.Sprintf("At %02d:%02d on %v", weekday.Hour, weekday.Minute, weekday.Weekdays)
}

// IntervalSchedule runs a task repeatedly at a fixed time interval.
type IntervalSchedule struct {
	Interval  time.Duration
	StartTime time.Time // Optional start time.
}

// NextRun returns the next run time after the provided time.
func (interval IntervalSchedule) NextRun(after time.Time) *time.Time {
	startRunTime := interval.StartTime
	if startRunTime.IsZero() {
		startRunTime = after
	}
	if startRunTime.After(after) {
		return &startRunTime
	}
	elapsedTime := after.Sub(startRunTime)
	numberOfIntervals := elapsedTime / interval.Interval
	nextRunTime := startRunTime.Add(interval.Interval * (numberOfIntervals + 1))
	return &nextRunTime
}

// Description returns a description of the interval schedule.
func (interval IntervalSchedule) Description() string {
	return fmt.Sprintf("Every %s", interval.Interval)
}

// containsWeekday checks if targetWeekday is in the list of weekdays.
func containsWeekday(weekdays []time.Weekday, targetWeekday time.Weekday) bool {
	for _, weekday := range weekdays {
		if weekday == targetWeekday {
			return true
		}
	}
	return false
}

// OneTimeSchedule implements a schedule that runs once and then never again
type OneTimeSchedule struct {
	runTime          time.Time
	hasExecuted      atomic.Bool
	executionChannel chan struct{}
}

// NewOneTimeSchedule creates a new schedule that runs once at the specified time
func NewOneTimeSchedule(executeAt time.Time) *OneTimeSchedule {
	return &OneTimeSchedule{
		runTime:          executeAt,
		executionChannel: make(chan struct{}, 1),
	}
}

// NextRun returns the next execution time or nil if already executed
func (schedule *OneTimeSchedule) NextRun(afterTime time.Time) *time.Time {
	// First check if already executed
	if schedule.hasExecuted.Load() {
		return nil // Return nil to indicate no more runs
	}

	// If we're past the scheduled time but haven't executed yet
	if afterTime.After(schedule.runTime) {
		// Mark as executed atomically
		schedule.hasExecuted.CompareAndSwap(false, true)

		// Return current time to run now
		now := afterTime.Add(time.Millisecond)
		return &now
	}

	// Return the scheduled time
	result := schedule.runTime
	return &result
}

// Description returns a human-readable description of the schedule
func (schedule *OneTimeSchedule) Description() string {
	return "One-time execution at " + schedule.runTime.Format("15:04:05")
}

// WaitForExecution waits for the execution to complete
func (schedule *OneTimeSchedule) WaitForExecution(ctx context.Context) bool {
	select {
	case <-schedule.executionChannel:
		return true
	case <-ctx.Done():
		return false
	}
}

// SignalExecution signals that the task has completed execution
func (schedule *OneTimeSchedule) SignalExecution() {
	// Non-blocking send to prevent hanging if WaitForExecution isn't called
	select {
	case schedule.executionChannel <- struct{}{}:
	default:
	}
}
