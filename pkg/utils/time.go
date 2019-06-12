package utils

import "time"

// Today return today start/end time at
func Today() (time.Time, time.Time) {
	end := time.Now()
	start := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.Local) // yyyy-mm-dd 00:00:00
	return start, end
}

// CurrMonth return current month start/end time at
func CurrMonth() (time.Time, time.Time) {
	end := time.Now()
	start := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.Local) // yyyy-mm-01 00:00:00
	return start, end
}
