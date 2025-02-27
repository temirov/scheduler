package utils

import "time"

// IsSameDay checks if two time.Time values represent the same calendar day,
// regardless of the time component. It compares year, month and day components.
func IsSameDay(firstTime, secondTime time.Time) bool {
	return firstTime.Year() == secondTime.Year() && 
	       firstTime.Month() == secondTime.Month() && 
	       firstTime.Day() == secondTime.Day()
} 