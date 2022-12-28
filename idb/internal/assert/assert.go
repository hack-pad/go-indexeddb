//go:build js && wasm
// +build js,wasm

// Package assert contains small assertion test functions to assist in writing clean tests.
package assert

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
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
		tb.Errorf("%+v != %+v\nExpected: %#v\nActual:    %#v", expected, actual, expected, actual)
		return false
	}
	return true
}

// NotEqual asserts actual is not equal to expected
func NotEqual(tb testing.TB, expected, actual interface{}) bool {
	tb.Helper()
	if reflect.DeepEqual(expected, actual) {
		tb.Errorf("Should not be equal: %+v\nExpected: %#v\nActual:    %#v", actual, expected, actual)
		return false
	}
	return true
}

func contains(tb testing.TB, collection, item interface{}) bool {
	collectionVal := reflect.ValueOf(collection)
	switch collectionVal.Kind() {
	case reflect.Slice:
		length := collectionVal.Len()
		for i := 0; i < length; i++ {
			candidateItem := collectionVal.Index(i).Interface()
			if reflect.DeepEqual(candidateItem, item) {
				return true
			}
		}
		return false
	case reflect.String:
		itemVal := reflect.ValueOf(item)
		if itemVal.Kind() != reflect.String {
			tb.Errorf("Invalid item type for string collection. Expected string, got: %T", item)
			return false
		}
		return strings.Contains(collection.(string), item.(string))
	default:
		tb.Errorf("Invalid collection type. Expected slice, got: %T", collection)
		return false
	}
}

// Contains asserts item is contained by collection
func Contains(tb testing.TB, collection, item interface{}) bool {
	tb.Helper()

	if !contains(tb, collection, item) {
		tb.Errorf("Collection does not contain expected item:\nCollection: %#v\nExpected item: %#v", collection, item)
		return false
	}
	return true
}

// NotContains asserts item is not contained by collection
func NotContains(tb testing.TB, collection, item interface{}) bool {
	tb.Helper()

	if contains(tb, collection, item) {
		tb.Errorf("Collection contains unexpected item:\nCollection: %#v\nUnexpected item: %#v", collection, item)
		return false
	}
	return true
}

// Eventually asserts fn() returns true within totalWait time, checking at the given interval
func Eventually(tb testing.TB, fn func(context.Context) bool, totalWait time.Duration, checkInterval time.Duration) bool {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), totalWait)
	defer cancel()
	for {
		success := fn(ctx)
		if success {
			return true
		}
		timer := time.NewTimer(checkInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			tb.Errorf("Condition did not become true within %s", totalWait)
			return false
		case <-timer.C:
		}
	}
}

// Panics asserts fn() panics
func Panics(tb testing.TB, fn func()) (panicked bool) {
	tb.Helper()
	defer func() {
		val := recover()
		if val != nil {
			panicked = true
		} else {
			tb.Error("Function should panic")
		}
	}()
	fn()
	return false
}

// NotPanics asserts fn() does not panic
func NotPanics(tb testing.TB, fn func()) bool {
	tb.Helper()
	defer func() {
		val := recover()
		if val != nil {
			tb.Errorf("Function should not panic, got: %#v", val)
		}
	}()
	fn()
	return true
}
