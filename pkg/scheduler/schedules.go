package scheduler

import (
	"fmt"
	"time"
)

// DailySchedule runs a task at the same time every day.
type DailySchedule struct {
	Hour   int
	Minute int
}

// NextRun returns the next run time after the provided time.
func (daily DailySchedule) NextRun(afterTime time.Time) time.Time {
	nextRunTime := time.Date(
		afterTime.Year(), afterTime.Month(), afterTime.Day(),
		daily.Hour, daily.Minute, 0, 0, afterTime.Location(),
	)
	if nextRunTime.Before(afterTime) {
		nextRunTime = nextRunTime.Add(24 * time.Hour)
	}
	return nextRunTime
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
func (weekday WeekdaySchedule) NextRun(afterTime time.Time) time.Time {
	nextRunTime := time.Date(
		afterTime.Year(), afterTime.Month(), afterTime.Day(),
		weekday.Hour, weekday.Minute, 0, 0, afterTime.Location(),
	)
	if nextRunTime.Before(afterTime) {
		nextRunTime = nextRunTime.Add(24 * time.Hour)
	}
	for !containsWeekday(weekday.Weekdays, nextRunTime.Weekday()) {
		nextRunTime = nextRunTime.Add(24 * time.Hour)
	}
	return nextRunTime
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
func (interval IntervalSchedule) NextRun(afterTime time.Time) time.Time {
	startRunTime := interval.StartTime
	if startRunTime.IsZero() {
		startRunTime = afterTime
	}
	if startRunTime.After(afterTime) {
		return startRunTime
	}
	elapsedTime := afterTime.Sub(startRunTime)
	numberOfIntervals := elapsedTime / interval.Interval
	return startRunTime.Add(interval.Interval * (numberOfIntervals + 1))
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
