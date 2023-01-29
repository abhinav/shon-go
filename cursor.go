package shon

// cursor points to a position in the input stream.
type cursor interface {
	// Reports whether there are more items
	// without moving the cursor.
	more() (ok bool)

	// Returns the next item or false,
	// without moving the cursor.
	//
	// Must return true if the prior 'more' call returned true.
	peek() (v string, ok bool)

	// Returns the next item or false,
	// and moves the cursor to the next position.
	//
	// Must return true if the prior 'more' call returned true.
	next() (v string, ok bool)
}

// sliceCursor implements cursor around a slice of values.
type sliceCursor struct {
	args []string
	pos  int
}

var _ cursor = (*sliceCursor)(nil)

func (c *sliceCursor) more() bool {
	return c.pos < len(c.args)
}

func (c *sliceCursor) next() (s string, ok bool) {
	if !c.more() {
		return "", false
	}
	arg := c.args[c.pos]
	c.pos++
	return arg, true
}

func (c *sliceCursor) peek() (s string, ok bool) {
	if c.more() {
		return c.args[c.pos], true
	}
	return "", false
}
