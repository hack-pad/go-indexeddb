// +build js,wasm

package exception

import (
	"syscall/js"

	"github.com/pkg/errors"
)

func Catch(err *error) {
	recoverErr := handleRecovery(recover())
	if recoverErr != nil {
		*err = recoverErr
	}
}

func CatchHandler(fn func(err error)) {
	err := handleRecovery(recover())
	if err != nil {
		fn(err)
	}
}

func handleRecovery(r interface{}) error {
	if r == nil {
		return nil
	}
	switch val := r.(type) {
	case error:
		return val
	case js.Value:
		return js.Error{Value: val}
	default:
		return errors.Errorf("%+v", val)
	}
}
