package utils

import (
	"path"
	"reflect"
	"runtime"
)

// FuncName return given function's runtime name
// f must be non-nil function Kind
func FuncName(f interface{}) string {
	v := reflect.ValueOf(f)
	fname := runtime.FuncForPC(v.Pointer()).Name()
	return path.Base(fname)
}
