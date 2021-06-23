// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

var (
	jsObjectStore = js.Global().Get("IDBObjectStore")
)

// CursorDirection is the direction of traversal of the cursor
type CursorDirection int

const (
	// CursorNext direction causes the cursor to be opened at the start of the source.
	CursorNext CursorDirection = iota
	// CursorNextUnique direction causes the cursor to be opened at the start of the source. For every key with duplicate values, only the first record is yielded.
	CursorNextUnique
	// CursorPrevious direction causes the cursor to be opened at the end of the source.
	CursorPrevious
	// CursorPreviousUnique direction causes the cursor to be opened at the end of the source. For every key with duplicate values, only the first record is yielded.
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

// Cursor represents a cursor for traversing or iterating over multiple records in a Database
type Cursor struct {
	jsCursor js.Value
}

func wrapCursor(jsCursor js.Value) *Cursor {
	return &Cursor{jsCursor}
}

// Source returns the ObjectStore or Index that the cursor is iterating
func (c *Cursor) Source() (objectStore *ObjectStore, index *Index, err error) {
	defer exception.Catch(&err)
	jsSource := c.jsCursor.Get("source")
	if jsSource.InstanceOf(jsObjectStore) {
		objectStore = wrapObjectStore(jsSource)
	} else if jsSource.InstanceOf(jsIDBIndex) {
		index = wrapIndex(jsSource)
	}
	return
}

// Direction returns the direction of traversal of the cursor
func (c *Cursor) Direction() (_ CursorDirection, err error) {
	defer exception.Catch(&err)
	direction := c.jsCursor.Get("direction")
	return parseCursorDirection(direction.String()), nil
}

// Key returns the key for the record at the cursor's position. If the cursor is outside its range, this is set to undefined.
func (c *Cursor) Key() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return c.jsCursor.Get("key"), nil
}

// PrimaryKey returns the cursor's current effective primary key. If the cursor is currently being iterated or has iterated outside its range, this is set to undefined.
func (c *Cursor) PrimaryKey() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return c.jsCursor.Get("primaryKey"), nil
}

// Request returns the Request that was used to obtain the cursor.
func (c *Cursor) Request() (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(c.jsCursor.Get("request")), nil
}

// Advance sets the number of times a cursor should move its position forward.
func (c *Cursor) Advance(count uint) (err error) {
	defer exception.Catch(&err)
	c.jsCursor.Call("advance", count)
	return nil
}

// Continue advances the cursor to the next position along its direction.
func (c *Cursor) Continue() (err error) {
	defer exception.Catch(&err)
	c.jsCursor.Call("continue")
	return nil
}

// ContinueKey advances the cursor to the next position along its direction.
func (c *Cursor) ContinueKey(key js.Value) (err error) {
	defer exception.Catch(&err)
	c.jsCursor.Call("continue", key)
	return nil
}

// ContinuePrimaryKey sets the cursor to the given index key and primary key given as arguments.
func (c *Cursor) ContinuePrimaryKey(key, primaryKey js.Value) (err error) {
	defer exception.Catch(&err)
	c.jsCursor.Call("continuePrimaryKey", key, primaryKey)
	return nil
}

// Delete returns an AckRequest, and, in a separate thread, deletes the record at the cursor's position, without changing the cursor's position. This can be used to delete specific records.
func (c *Cursor) Delete() (_ *AckRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(c.jsCursor.Call("delete"))
	return newAckRequest(req), nil
}

// Update returns a Request, and, in a separate thread, updates the value at the current position of the cursor in the object store. This can be used to update specific records.
func (c *Cursor) Update(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(c.jsCursor.Call("update", value)), nil
}

// CursorWithValue represents a cursor for traversing or iterating over multiple records in a database. It is the same as the Cursor, except that it includes the value property.
type CursorWithValue struct {
	*Cursor
}

func wrapCursorWithValue(jsCursor js.Value) *CursorWithValue {
	return &CursorWithValue{wrapCursor(jsCursor)}
}

// Value returns the value of the current cursor
func (c *CursorWithValue) Value() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return c.jsCursor.Get("value"), nil
}
