//go:build js && wasm
// +build js,wasm

package jscache

import (
	"github.com/hack-pad/safejs"
)

var (
	jsReflectGet = safejs.Must(safejs.Must(safejs.Global().Get("Reflect")).Get("get"))
)

// Strings caches encoding strings as safejs.Value's.
// String encoding today is quite CPU intensive, so caching commonly used strings helps with performance.
type Strings struct {
	cacher
}

// Value retrieves the safejs.Value for the given string
func (c *Strings) Value(s string) safejs.Value {
	return c.value(s, identityStringGetter{s}.value)
}

// GetProperty retrieves the given object's property, using a cached string value if available. Saves on the performance cost of 2 round trips to JS.
func (c *Strings) GetProperty(obj safejs.Value, key string) (safejs.Value, error) {
	jsKey := c.Value(key)
	return jsReflectGet.Invoke(obj, jsKey)
}

type identityStringGetter struct {
	s string
}

func (i identityStringGetter) value() safejs.Value {
	value, err := safejs.ValueOf(i.s)
	if err != nil {
		panic(err)
	}
	return value
}
