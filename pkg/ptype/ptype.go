package ptype

import (
	"time"
)

// String returns a pointer to the given string value.
func String(v string) *string {
	return &v
}

// StringV returns the value of the given string pointer or
// "" if the pointer is nil.
func StringV(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// StringSlice converts a slice of string values into a slice of
// string pointers
func StringSlice(src []string) []*string {
	dst := make([]*string, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// StringSliceV converts a slice of string pointers into a slice of
// string values
func StringSliceV(src []*string) []string {
	dst := make([]string, len(src))
	for i := 0; i < len(src); i++ {
		if src[i] != nil {
			dst[i] = *(src[i])
		}
	}
	return dst
}

// StringMap converts a string map of string values into a string
// map of string pointers
func StringMap(src map[string]string) map[string]*string {
	dst := make(map[string]*string)
	for k, val := range src {
		v := val
		dst[k] = &v
	}
	return dst
}

// StringMapV converts a string map of string pointers into a string
// map of string values
func StringMapV(src map[string]*string) map[string]string {
	dst := make(map[string]string)
	for k, val := range src {
		if val != nil {
			dst[k] = *val
		}
	}
	return dst
}

// Bool returns a pointer to the given bool value.
func Bool(v bool) *bool {
	return &v
}

// BoolV returns the value of the given bool pointer or
// false if the pointer is nil.
func BoolV(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}

// BoolSlice converts a slice of bool values into a slice of
// bool pointers
func BoolSlice(src []bool) []*bool {
	dst := make([]*bool, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// BoolSliceV converts a slice of bool pointers into a slice of
// bool values
func BoolSliceV(src []*bool) []bool {
	dst := make([]bool, len(src))
	for i := 0; i < len(src); i++ {
		if src[i] != nil {
			dst[i] = *(src[i])
		}
	}
	return dst
}

// BoolMap converts a string map of bool values into a string
// map of bool pointers
func BoolMap(src map[string]bool) map[string]*bool {
	dst := make(map[string]*bool)
	for k, val := range src {
		v := val
		dst[k] = &v
	}
	return dst
}

// BoolMapV converts a string map of bool pointers into a string
// map of bool values
func BoolMapV(src map[string]*bool) map[string]bool {
	dst := make(map[string]bool)
	for k, val := range src {
		if val != nil {
			dst[k] = *val
		}
	}
	return dst
}

// Int64 returns a pointer to the given int64 value.
func Int64(v int64) *int64 {
	return &v
}

// Int64V returns the value of the given int64 pointer or
// 0 if the pointer is nil.
func Int64V(v *int64) int64 {
	if v != nil {
		return *v
	}
	return int64(0)
}

// Int returns a pointer to the given int value.
func Int(v int) *int {
	return &v
}

// IntV returns the value of the given int pointer or
// 0 if the pointer is nil.
func IntV(v *int) int {
	if v != nil {
		return *v
	}
	return 0
}

// IntSlice converts a slice of int values into a slice of
// int pointers
func IntSlice(src []int) []*int {
	dst := make([]*int, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// IntSliceV converts a slice of int pointers into a slice of
// int values
func IntSliceV(src []*int) []int {
	dst := make([]int, len(src))
	for i := 0; i < len(src); i++ {
		if src[i] != nil {
			dst[i] = *(src[i])
		}
	}
	return dst
}

// IntMap converts a string map of int values into a string
// map of int pointers
func IntMap(src map[string]int) map[string]*int {
	dst := make(map[string]*int)
	for k, val := range src {
		v := val
		dst[k] = &v
	}
	return dst
}

// IntMapV converts a string map of int pointers into a string
// map of int values
func IntMapV(src map[string]*int) map[string]int {
	dst := make(map[string]int)
	for k, val := range src {
		if val != nil {
			dst[k] = *val
		}
	}
	return dst
}

// Time returns a pointer to the given time.Time value.
func Time(v time.Time) *time.Time {
	return &v
}

// TimeV returns the value of the given time.Time pointer or
// time.Time{} if the pointer is nil.
func TimeV(v *time.Time) time.Time {
	if v != nil {
		return *v
	}
	return time.Time{}
}

// TimeDuration returns a pointer to the given time.Duration value.
func TimeDuration(dur time.Duration) *time.Duration {
	return &dur
}

// TimeDurationV returns the value of the given time.Duration pointer or
// time.Duration(0) if the pointer is nil.
func TimeDurationV(v *time.Duration) time.Duration {
	if v != nil {
		return *v
	}
	return time.Duration(0)
}

// TimeUnixMilli returns a Unix timestamp in milliseconds from "January 1, 1970 UTC".
// The result is undefined if the Unix time cannot be represented by an int64.
// Which includes calling TimeUnixMilli on a zero Time is undefined.
//
// See Go stdlib https://golang.org/pkg/time/#Time.UnixNano for more information.
func TimeUnixMilli(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond/time.Nanosecond)
}

// TimeSlice converts a slice of time.Time values into a slice of
// time.Time pointers
func TimeSlice(src []time.Time) []*time.Time {
	dst := make([]*time.Time, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// TimeSliceV converts a slice of time.Time pointers into a slice of
// time.Time values
func TimeSliceV(src []*time.Time) []time.Time {
	dst := make([]time.Time, len(src))
	for i := 0; i < len(src); i++ {
		if src[i] != nil {
			dst[i] = *(src[i])
		}
	}
	return dst
}

// TimeMap converts a string map of time.Time values into a string
// map of time.Time pointers
func TimeMap(src map[string]time.Time) map[string]*time.Time {
	dst := make(map[string]*time.Time)
	for k, val := range src {
		v := val
		dst[k] = &v
	}
	return dst
}

// TimeMapV converts a string map of time.Time pointers into a string
// map of time.Time values
func TimeMapV(src map[string]*time.Time) map[string]time.Time {
	dst := make(map[string]time.Time)
	for k, val := range src {
		if val != nil {
			dst[k] = *val
		}
	}
	return dst
}
