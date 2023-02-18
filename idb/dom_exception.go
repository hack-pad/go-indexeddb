//go:build js && wasm
// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/safejs"
)

type DOMException struct {
	name    string
	message string
}

func tryAsDOMException(err error) error {
	switch err := err.(type) {
	case js.Error:
		return domExceptionAsError(safejs.Safe(err.Value))
	default:
		return err
	}
}

func domExceptionAsError(jsDOMException safejs.Value) error {
	truthy, err := jsDOMException.Truthy()
	if err != nil || !truthy {
		return err
	}
	domException, err := parseJSDOMException(jsDOMException)
	if err != nil {
		return err
	}
	return domException
}

func parseJSDOMException(jsDOMException safejs.Value) (DOMException, error) {
	name, err := jsDOMException.Get("name")
	if err != nil {
		return DOMException{}, err
	}
	nameStr, err := name.String()
	if err != nil {
		return DOMException{}, err
	}
	message, err := jsDOMException.Get("message")
	if err != nil {
		return DOMException{}, err
	}
	messageStr, err := message.String()
	if err != nil {
		return DOMException{}, err
	}
	return DOMException{
		name:    nameStr,
		message: messageStr,
	}, nil
}

func (e DOMException) Error() string {
	return e.name + ": " + e.message
}

func (e DOMException) Is(target error) bool {
	targetDOMException, ok := target.(DOMException)
	return ok && targetDOMException.name == e.name
}
