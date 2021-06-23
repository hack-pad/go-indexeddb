// +build js,wasm

package promise

import (
	"github.com/pkg/errors"
)

// Chan is a Go channel-based promise
type Chan struct {
	resolveChan, rejectChan <-chan interface{}
}

// NewChan creates a new Chan
func NewChan() (resolve, reject Resolver, promise Chan) {
	resolveChan, rejectChan := make(chan interface{}, 1), make(chan interface{}, 1)
	var c Chan
	c.resolveChan, c.rejectChan = resolveChan, rejectChan

	resolve = func(result interface{}) {
		resolveChan <- result
		close(resolveChan)
		close(rejectChan)
	}
	reject = func(result interface{}) {
		rejectChan <- result
		close(resolveChan)
		close(rejectChan)
	}
	return resolve, reject, c
}

// Then implements Promise
func (c Chan) Then(fn func(value interface{}) interface{}) Promise {
	// TODO support failing a Then call
	resolve, _, prom := NewChan()
	go func() {
		value, ok := <-c.resolveChan
		if ok {
			newValue := fn(value)
			resolve(newValue)
		}
	}()
	return prom
}

// Catch implements Promise
func (c Chan) Catch(fn func(rejectedReason interface{}) interface{}) Promise {
	_, reject, prom := NewChan()
	go func() {
		reason, ok := <-c.rejectChan
		if ok {
			newReason := fn(reason)
			reject(newReason)
		}
	}()
	return prom
}

// Await implements Promise
func (c Chan) Await() (interface{}, error) {
	// TODO support error handling inside promise functions instead
	value := <-c.resolveChan
	switch err := (<-c.rejectChan).(type) {
	case nil:
		return value, nil
	case error:
		return value, err
	default:
		return value, errors.Errorf("%v", err)
	}
}
