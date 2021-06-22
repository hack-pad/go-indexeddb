// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

var (
	jsObjectStore = js.Global().Get("IDBObjectStore")
)

type CursorDirection int

const (
	CursorNext CursorDirection = iota
	CursorNextUnique
	CursorPrevious
	CursorPreviousUnique
)

func parseCursorDirection(s string) CursorDirection {
	switch s {
	case "nextunique":
		return CursorNextUnique
	case "previous":
		return CursorPrevious
	case "previousunique":
		return CursorPreviousUnique
	default:
		return CursorNext
	}
}

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

func wrapCursor(jsCursor js.Value) *Cursor {
	return &Cursor{jsCursor}
}

func (c *Cursor) Source() (_ interface{}, err error) {
	defer exception.Catch(&err)
	source := c.jsCursor.Get("source")
	if source.InstanceOf(jsObjectStore) {
		return wrapObjectStore(source), nil
	}
	return wrapIndex(source), nil
}

func (c *Cursor) Direction() (_ CursorDirection, err error) {
	defer exception.Catch(&err)
	direction := c.jsCursor.Get("direction")
	return parseCursorDirection(direction.String()), nil
}

func (c *Cursor) Key() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return c.jsCursor.Get("key"), nil
}

func (c *Cursor) PrimaryKey() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return c.jsCursor.Get("primaryKey"), nil
}

func (c *Cursor) Request() (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(c.jsCursor.Get("request")), nil
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

func (c *Cursor) ContinueKey(key js.Value) (err error) {
	defer exception.Catch(&err)
	c.jsCursor.Call("continue", key)
	return nil
}

func (c *Cursor) ContinuePrimaryKey(key, primaryKey js.Value) (err error) {
	defer exception.Catch(&err)
	c.jsCursor.Call("continuePrimaryKey", key, primaryKey)
	return nil
}

func (c *Cursor) Delete() (_ *AckRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(c.jsCursor.Call("delete"))
	return newAckRequest(req), nil
}

func (c *Cursor) Update(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(c.jsCursor.Call("update", value)), nil
}

type CursorWithValue struct {
	*Cursor
}

func wrapCursorWithValue(jsCursor js.Value) *CursorWithValue {
	return &CursorWithValue{wrapCursor(jsCursor)}
}

func (c *CursorWithValue) Value() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return c.jsCursor.Get("value"), nil
}
