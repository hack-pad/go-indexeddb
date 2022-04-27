// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

var (
	jsIDBKeyRange = js.Global().Get("IDBKeyRange")
)

// KeyRange represents a continuous interval over some data type that is used for keys. Records can be retrieved from ObjectStore and Index objects using keys or a range of keys.
type KeyRange struct {
	js.Value
}

func wrapKeyRange(jsKeyRange js.Value) *KeyRange {
	return &KeyRange{jsKeyRange}
}

// NewKeyRangeBound creates a new key range with the specified upper and lower bounds.
// The bounds can be open (that is, the bounds exclude the endpoint values) or closed (that is, the bounds include the endpoint values).
func NewKeyRangeBound(lower, upper js.Value, lowerOpen, upperOpen bool) (_ *KeyRange, err error) {
	defer exception.Catch(&err)
	return wrapKeyRange(jsIDBKeyRange.Call("bound", lower, upper, lowerOpen, upperOpen)), nil
}

// NewKeyRangeLowerBound creates a new key range with only a lower bound.
func NewKeyRangeLowerBound(lower js.Value, open bool) (_ *KeyRange, err error) {
	defer exception.Catch(&err)
	return wrapKeyRange(jsIDBKeyRange.Call("lowerBound", lower, open)), nil
}

// NewKeyRangeUpperBound creates a new key range with only an upper bound.
func NewKeyRangeUpperBound(upper js.Value, open bool) (_ *KeyRange, err error) {
	defer exception.Catch(&err)
	return wrapKeyRange(jsIDBKeyRange.Call("upperBound", upper, open)), nil
}

// NewKeyRangeOnly creates a new key range containing a single value.
func NewKeyRangeOnly(only js.Value) (_ *KeyRange, err error) {
	defer exception.Catch(&err)
	return wrapKeyRange(jsIDBKeyRange.Call("only", only)), nil
}

// Lower returns the lower bound of the key range.
func (k *KeyRange) Lower() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return k.Get("lower"), nil
}

// Upper returns the upper bound of the key range.
func (k *KeyRange) Upper() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return k.Get("upper"), nil
}

// LowerOpen returns false if the lower-bound value is included in the key range.
func (k *KeyRange) LowerOpen() (_ bool, err error) {
	defer exception.Catch(&err)
	return k.Get("lowerOpen").Bool(), nil
}

// UpperOpen returns false if the upper-bound value is included in the key range.
func (k *KeyRange) UpperOpen() (_ bool, err error) {
	defer exception.Catch(&err)
	return k.Get("upperOpen").Bool(), nil
}

// Includes returns a boolean indicating whether a specified key is inside the key range.
func (k *KeyRange) Includes(key js.Value) (_ bool, err error) {
	defer exception.Catch(&err)
	return k.Call("includes", key).Bool(), nil
}

// JSValue implements js.Wrapper
// removed see : https://github.com/golang/go/issues/44006
