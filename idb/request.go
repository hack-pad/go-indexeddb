//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"errors"
	"log"
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
	"github.com/hack-pad/safejs"
)

var (
	// ErrCursorStopIter stops iteration when returned from a CursorRequest.Iter() handler
	ErrCursorStopIter = errors.New("stop cursor iteration")
)

var (
	jsIDBRequest = safejs.Must(safejs.Global().Get("IDBRequest"))
	jsIDBIndex   = safejs.Must(safejs.Global().Get("IDBIndex"))
)

// Request provides access to results of asynchronous requests to databases and database objects
// using event listeners. Each reading and writing operation on a database is done using a request.
type Request struct {
	txn       *Transaction
	jsRequest safejs.Value
}

func wrapRequest(txn *Transaction, jsRequest safejs.Value) *Request {
	if isInstance, err := jsRequest.InstanceOf(jsIDBRequest); !isInstance || err != nil {
		panic("Invalid JS request type")
	}
	if txn == nil {
		txn = (*Transaction)(nil)
	}
	return &Request{
		txn:       txn,
		jsRequest: jsRequest,
	}
}

// Source returns the source of the request, such as an Index or an ObjectStore. If no source exists (such as when calling Factory.Open), it returns nil for both.
func (r *Request) Source() (objectStore *ObjectStore, index *Index, err error) {
	jsSource, err := r.jsRequest.Get("source")
	if err != nil {
		return
	}
	if isInstance, _ := jsSource.InstanceOf(jsObjectStore); isInstance {
		objectStore = wrapObjectStore(r.txn, jsSource)
	} else if isInstance, _ := jsSource.InstanceOf(jsIDBIndex); isInstance {
		index = wrapIndex(r.txn, jsSource)
	}
	return
}

func (r *Request) safeResult() (safejs.Value, error) {
	return r.jsRequest.Get("result")
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (r *Request) Result() (js.Value, error) {
	value, err := r.safeResult()
	return safejs.Unsafe(value), err
}

// Err returns an error in the event of an unsuccessful request, indicating what went wrong.
func (r *Request) Err() (err error) {
	jsErr, err := r.jsRequest.Get("error")
	if err != nil {
		return err
	}
	if truthy, err := jsErr.Truthy(); err != nil {
		return err
	} else if truthy {
		return js.Error{Value: safejs.Unsafe(jsErr)}
	}
	return nil
}

func (r *Request) safeAwait(ctx context.Context) (result safejs.Value, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	r.Listen(ctx, func() {
		result, err = r.safeResult()
		cancel()
	}, func() {
		err = r.Err()
		cancel()
	})
	<-ctx.Done()
	return
}

// Await waits for success or failure, then returns the results.
func (r *Request) Await(ctx context.Context) (js.Value, error) {
	result, err := r.safeAwait(ctx)
	return safejs.Unsafe(result), err
}

// ReadyState returns the state of the request. Every request starts in the pending state. The state changes to done when the request completes successfully or when an error occurs.
func (r *Request) ReadyState() (string, error) {
	readyState, err := r.jsRequest.Get("readyState")
	if err != nil {
		return "", err
	}
	return readyState.String()
}

// Transaction returns the transaction for the request. This can return nil for certain requests, for example those returned from Factory.Open unless an upgrade is needed. (You're just connecting to a database, so there is no transaction to return).
func (r *Request) Transaction() (*Transaction, error) {
	if r.txn == (*Transaction)(nil) {
		return nil, errNotInTransaction
	}
	return r.txn, nil
}

// ListenSuccess invokes the callback when the request succeeds
func (r *Request) ListenSuccess(ctx context.Context, success func()) {
	r.Listen(ctx, success, nil)
}

// ListenError invokes the callback when the request fails
func (r *Request) ListenError(ctx context.Context, failed func()) {
	r.Listen(ctx, nil, failed)
}

// Listen invokes the success callback when the request succeeds and failed when it fails.
func (r *Request) Listen(ctx context.Context, success, failed func()) {
	if success != nil {
		// by default, only listen for 1 value
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		originalSuccess := success
		success = func() {
			defer cancel()
			originalSuccess()
		}
	}
	r.listen(ctx, success, failed)
}

// listen is like Listen, but doesn't cancel the context after success is called
func (r *Request) listen(ctx context.Context, success, failed func()) {
	ctx, cancel := context.WithCancel(ctx)
	panicHandler := func(err error) {
		log.Println("Failed resolving request results:", err)
		txn, err := r.Transaction()
		if err == nil {
			_ = txn.Abort()
		}
		cancel()
		ignorePanic(failed) // helps the listener to cancel the outer context
	}

	if failed != nil {
		errFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			defer exception.CatchHandler(panicHandler)
			failed()
			cancel()
			return nil
		})
		_, err := r.jsRequest.Call(addEventListener, "error", errFunc)
		if err != nil {
			panic(err)
		}
		go func() {
			<-ctx.Done()
			_, err := r.jsRequest.Call(removeEventListener, "error", errFunc)
			if err != nil {
				panic(err)
			}
			errFunc.Release()
		}()
	}
	if success != nil {
		successFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			defer exception.CatchHandler(panicHandler)
			success()
			// don't cancel ctx here, need to allow multiple values for cursors
			return nil
		})
		_, err := r.jsRequest.Call(addEventListener, "success", successFunc)
		if err != nil {
			panic(err)
		}
		go func() {
			<-ctx.Done()
			_, err := r.jsRequest.Call(removeEventListener, "success", successFunc)
			if err != nil {
				panic(err)
			}
			successFunc.Release()
		}()
	}
}

func ignorePanic(fn func()) {
	defer func() {
		_ = recover()
	}()
	fn()
}

// UintRequest is a Request that retrieves a uint result
type UintRequest struct {
	*Request
}

func newUintRequest(req *Request) *UintRequest {
	return &UintRequest{req}
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (u *UintRequest) Result() (uint, error) {
	result, err := u.Request.safeResult()
	if err != nil {
		return 0, err
	}
	value, err := result.Int()
	return uint(value), err
}

// Await waits for success or failure, then returns the results.
func (u *UintRequest) Await(ctx context.Context) (uint, error) {
	result, err := u.Request.safeAwait(ctx)
	if err != nil {
		return 0, err
	}
	value, err := result.Int()
	return uint(value), err
}

// ArrayRequest is a Request that retrieves an array of js.Values
type ArrayRequest struct {
	*Request
}

func newArrayRequest(req *Request) *ArrayRequest {
	return &ArrayRequest{req}
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (a *ArrayRequest) Result() ([]js.Value, error) {
	result, err := a.Request.safeResult()
	if err != nil {
		return nil, err
	}
	var values []js.Value
	err = iterArray(result, func(i int, value safejs.Value) (bool, error) {
		values = append(values, safejs.Unsafe(value))
		return true, nil
	})
	return values, err
}

// Await waits for success or failure, then returns the results.
func (a *ArrayRequest) Await(ctx context.Context) ([]js.Value, error) {
	result, err := a.Request.safeAwait(ctx)
	if err != nil {
		return nil, err
	}
	var values []js.Value
	err = iterArray(result, func(i int, value safejs.Value) (bool, error) {
		values = append(values, safejs.Unsafe(value))
		return true, nil
	})
	return values, err
}

// AckRequest is a Request that doesn't retrieve a value, only used to detect errors.
type AckRequest struct {
	*Request
}

func newAckRequest(req *Request) *AckRequest {
	return &AckRequest{req}
}

// Result is a no-op. This kind of request does not retrieve any data in the result.
func (a *AckRequest) Result() {} // no-op

// Await waits for success or failure, then returns the results.
func (a *AckRequest) Await(ctx context.Context) error {
	_, err := a.Request.safeAwait(ctx)
	return err
}

func cursorIter(ctx context.Context, req *Request, iter func(*Cursor) error) error {
	ctx, cancel := context.WithCancel(ctx)
	var returnErr error
	req.listen(ctx, func() {
		jsCursor, err := req.safeResult()
		if err != nil {
			returnErr = err
			cancel()
			return
		}
		if jsCursor.IsNull() {
			cancel()
			return
		}
		cursor := wrapCursor(req.txn, jsCursor)
		err = iter(cursor)
		if err != nil {
			if err != ErrCursorStopIter {
				returnErr = err
			}
			cancel()
			return
		}
		if !cursor.iterated {
			err := cursor.Continue()
			if err != nil {
				returnErr = err
				cancel()
				return
			}
		}
	}, func() {
		returnErr = req.Err()
		if returnErr == nil {
			returnErr = errors.New("Failed to handle panic in JS callback")
		}
		cancel()
	})
	<-ctx.Done()
	return returnErr
}

// CursorRequest is a Request that retrieves a Cursor
type CursorRequest struct {
	*Request
}

func newCursorRequest(req *Request) *CursorRequest {
	return &CursorRequest{req}
}

// Iter invokes the callback when the request succeeds for each cursor iteration
func (c *CursorRequest) Iter(ctx context.Context, iter func(*Cursor) error) error {
	return cursorIter(ctx, c.Request, func(cursor *Cursor) error {
		return iter(cursor)
	})
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (c *CursorRequest) Result() (*Cursor, error) {
	result, err := c.Request.safeResult()
	if err != nil {
		return nil, err
	}
	return wrapCursor(c.txn, result), nil
}

// Await waits for success or failure, then returns the results.
func (c *CursorRequest) Await(ctx context.Context) (*Cursor, error) {
	result, err := c.Request.safeAwait(ctx)
	if err != nil {
		return nil, err
	}
	return wrapCursor(c.txn, result), nil
}

// CursorWithValueRequest is a Request that retrieves a CursorWithValue
type CursorWithValueRequest struct {
	*Request
}

func newCursorWithValueRequest(req *Request) *CursorWithValueRequest {
	return &CursorWithValueRequest{req}
}

// Iter invokes the callback when the request succeeds for each cursor iteration
func (c *CursorWithValueRequest) Iter(ctx context.Context, iter func(*CursorWithValue) error) error {
	return cursorIter(ctx, c.Request, func(cursor *Cursor) error {
		return iter(newCursorWithValue(cursor))
	})
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (c *CursorWithValueRequest) Result() (*CursorWithValue, error) {
	result, err := c.Request.safeResult()
	if err != nil {
		return nil, err
	}
	return wrapCursorWithValue(c.txn, result), nil
}

// Await waits for success or failure, then returns the results.
func (c *CursorWithValueRequest) Await(ctx context.Context) (*CursorWithValue, error) {
	result, err := c.Request.safeAwait(ctx)
	if err != nil {
		return nil, err
	}
	return wrapCursorWithValue(c.txn, result), nil
}
