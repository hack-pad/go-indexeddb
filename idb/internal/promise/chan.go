// +build js,wasm

package promise

import (
	"context"
	"fmt"
)

// Chan is a Go channel-based promise
type Chan struct {
	ctx                     context.Context
	resolveChan, rejectChan <-chan interface{}
}

// NewChan creates a new Chan
func NewChan(ctx context.Context) (resolve, reject Resolver, promise Chan) {
	resolversCtx, done := context.WithCancel(context.Background())
	resolveChan, rejectChan := make(chan interface{}, 1), make(chan interface{}, 1)
	c := Chan{
		ctx:         ctx,
		resolveChan: resolveChan,
		rejectChan:  rejectChan,
	}

	resolve = func(result interface{}) {
		select {
		case <-ctx.Done():
		case <-resolversCtx.Done():
		default:
			resolveChan <- result
		}
		done()
	}
	reject = func(result interface{}) {
		select {
		case <-ctx.Done():
		case <-resolversCtx.Done():
		default:
			rejectChan <- result
		}
		done()
	}
	go func() {
		select {
		case <-ctx.Done():
			rejectChan <- ctx.Err()
		case <-resolversCtx.Done():
		}
		close(resolveChan)
		close(rejectChan)
	}()
	return resolve, reject, c
}

// Then implements Promise
func (c Chan) Then(fn func(value interface{}) interface{}) Promise {
	// TODO support failing a Then call
	resolve, _, prom := NewChan(c.ctx)
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
	_, reject, prom := NewChan(c.ctx)
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
		return value, fmt.Errorf("%v", err)
	}
}
