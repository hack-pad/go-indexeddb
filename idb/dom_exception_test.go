//go:build js && wasm
// +build js,wasm

package idb

import (
	"syscall/js"
	"testing"

	"github.com/hack-pad/go-indexeddb/idb/internal/assert"
	"github.com/hack-pad/safejs"
)

var domException safejs.Value

func init() {
	var err error
	domException, err = safejs.Global().Get("DOMException")
	if err != nil {
		panic(err)
	}
}

func TestTryAsDOMException(t *testing.T) {
	t.Parallel()
	exceptionJS, err := domException.New("message", "name")
	assert.NoError(t, err)
	exception := js.Error{Value: safejs.Unsafe(exceptionJS)}
	assert.Equal(t, DOMException{
		name:    "name",
		message: "message",
	}, tryAsDOMException(exception))
}

func TestDOMExceptionAsError(t *testing.T) {
	t.Parallel()
	exceptionJS, err := domException.New("message", "name")
	assert.NoError(t, err)
	exception := domExceptionAsError(exceptionJS)
	assert.Equal(t, DOMException{
		name:    "name",
		message: "message",
	}, exception)

	assert.Equal(t, "name: message", exception.Error())

	assert.ErrorIs(t, exception, DOMException{name: "name"})
	assert.NotErrorIs(t, exception, DOMException{name: "other name"})
}
