// +build js,wasm

package jscache

import "syscall/js"

type cacher struct {
	cache map[string]js.Value
}

func (c *cacher) value(key string, valueFn func() interface{}) js.Value {
	if val, ok := c.cache[key]; ok {
		return val
	}
	if c.cache == nil {
		c.cache = make(map[string]js.Value)
	}
	val := js.ValueOf(valueFn())
	c.cache[key] = val
	return val
}
