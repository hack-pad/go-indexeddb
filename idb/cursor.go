// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

type CursorDirection int

const (
	CursorNext CursorDirection = iota
	CursorNextUnique
	CursorPrevious
	CursorPreviousUnique
)

func (d CursorDirection) String() string {
	switch d {
	case CursorNextUnique:
		return "nextunique"
	case CursorPrevious:
		return "previous"
	case CursorPreviousUnique:
		return "previousunique"
	default:
		return "next"
	}
}

type Cursor struct {
	jsCursor js.Value
}

func (c *Cursor) Advance(count uint) (err error) {
	defer exception.Catch(&err)
	c.jsCursor.Call("advance", count)
	return nil
}

func (c *Cursor) Continue() (err error) {
	defer exception.Catch(&err)
	c.jsCursor.Call("continue")
	return nil
}

func (c *Cursor) ContinuePrimaryKey(key, primaryKey js.Value) (err error) {
	defer exception.Catch(&err)
	c.jsCursor.Call("continuePrimaryKey", key, primaryKey)
	return nil
}

func (c *Cursor) Delete() (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(c.jsCursor.Call("delete")), nil
}

func (c *Cursor) Update(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(c.jsCursor.Call("update", value)), nil
}
