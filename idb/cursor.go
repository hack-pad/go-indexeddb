// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
	"github.com/hack-pad/go-indexeddb/idb/internal/jscache"
)

var (
	jsObjectStore        = js.Global().Get("IDBObjectStore")
	cursorDirectionCache jscache.Strings
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
	case "prev":
		return CursorPrevious
	case "prevunique":
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
		return "prev"
	case CursorPreviousUnique:
		return "prevunique"
	default:
		return "next"
	}
}

func (d CursorDirection) jsValue() js.Value {
	return cursorDirectionCache.Value(d.String())
}

// Cursor represents a cursor for traversing or iterating over multiple records in a Database
type Cursor struct {
	txn      *Transaction
	jsCursor js.Value
	iterated bool // set to true when an iteration method is called, like Continue
}

func wrapCursor(txn *Transaction, jsCursor js.Value) *Cursor {
	if txn == nil {
		txn = (*Transaction)(nil)
	}
	return &Cursor{
		txn:      txn,
		jsCursor: jsCursor,
	}
}

// Source returns the ObjectStore or Index that the cursor is iterating
func (c *Cursor) Source() (objectStore *ObjectStore, index *Index, err error) {
	defer exception.Catch(&err)
	jsSource := c.jsCursor.Get("source")
	if jsSource.InstanceOf(jsObjectStore) {
		objectStore = wrapObjectStore(c.txn, jsSource)
	} else if jsSource.InstanceOf(jsIDBIndex) {
		index = wrapIndex(c.txn, jsSource)
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
	return wrapRequest(c.txn, c.jsCursor.Get("request")), nil
}

// Advance sets the number of times a cursor should move its position forward.
func (c *Cursor) Advance(count uint) (err error) {
	defer exception.Catch(&err)
	c.iterated = true
	c.jsCursor.Call("advance", count)
	return nil
}

// Continue advances the cursor to the next position along its direction.
func (c *Cursor) Continue() (err error) {
	defer exception.Catch(&err)
	c.iterated = true
	c.jsCursor.Call("continue")
	return nil
}

// ContinueKey advances the cursor to the next position along its direction.
func (c *Cursor) ContinueKey(key js.Value) (err error) {
	defer exception.Catch(&err)
	c.iterated = true
	c.jsCursor.Call("continue", key)
	return nil
}

// ContinuePrimaryKey sets the cursor to the given index key and primary key given as arguments. Returns an error if the source is not an index.
func (c *Cursor) ContinuePrimaryKey(key, primaryKey js.Value) (err error) {
	defer exception.Catch(&err)
	c.iterated = true
	c.jsCursor.Call("continuePrimaryKey", key, primaryKey)
	return nil
}

// Delete returns an AckRequest, and, in a separate thread, deletes the record at the cursor's position, without changing the cursor's position. This can be used to delete specific records.
func (c *Cursor) Delete() (_ *AckRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(c.txn, c.jsCursor.Call("delete"))
	return newAckRequest(req), nil
}

// Update returns a Request, and, in a separate thread, updates the value at the current position of the cursor in the object store. This can be used to update specific records.
func (c *Cursor) Update(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(c.txn, c.jsCursor.Call("update", value)), nil
}

// CursorWithValue represents a cursor for traversing or iterating over multiple records in a database. It is the same as the Cursor, except that it includes the value property.
type CursorWithValue struct {
	*Cursor
}

func newCursorWithValue(cursor *Cursor) *CursorWithValue {
	return &CursorWithValue{cursor}
}

func wrapCursorWithValue(txn *Transaction, jsCursor js.Value) *CursorWithValue {
	return newCursorWithValue(wrapCursor(txn, jsCursor))
}

// Value returns the value of the current cursor
func (c *CursorWithValue) Value() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return c.jsCursor.Get("value"), nil
}
