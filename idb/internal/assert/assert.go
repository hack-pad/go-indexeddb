// +build js,wasm

package assert

import (
	"reflect"
	"testing"
)

// Error asserts err is not nil
func Error(tb testing.TB, err error) bool {
	tb.Helper()
	if err == nil {
		tb.Error("Expected error, got nil")
		return false
	}
	return true
}

// NoError asserts err is nil
func NoError(tb testing.TB, err error) bool {
	tb.Helper()
	if err != nil {
		tb.Errorf("Unexpected error: %+v", err)
		return false
	}
	return true
}

// Zero asserts value is the zero value
func Zero(tb testing.TB, value interface{}) bool {
	tb.Helper()
	if !reflect.ValueOf(value).IsZero() {
		tb.Errorf("Value should be zero, got: %#v", value)
		return false
	}
	return true
}

// NotZero asserts value is not the zero value
func NotZero(tb testing.TB, value interface{}) bool {
	tb.Helper()
	if reflect.ValueOf(value).IsZero() {
		tb.Error("Value should not be zero")
		return false
	}
	return true
}

// Equal asserts actual is equal to expected
func Equal(tb testing.TB, expected, actual interface{}) bool {
	tb.Helper()
	if !reflect.DeepEqual(expected, actual) {
		tb.Errorf("Expected: %#v\nActual:    %#v", expected, actual)
		return false
	}
	return true
}

// NotEqual asserts actual is not equal to expected
func NotEqual(tb testing.TB, expected, actual interface{}) bool {
	tb.Helper()
	if reflect.DeepEqual(expected, actual) {
		tb.Errorf("Should not be equal.\nExpected: %#v\nActual:    %#v", expected, actual)
		return false
	}
	return true
}
