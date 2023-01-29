package shon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceCursor(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		sc := sliceCursor{args: []string{"foo", "bar"}}

		if assert.True(t, sc.more(), "expected more") {
			s, ok := sc.peek()
			assert.True(t, ok)
			assert.Equal(t, "foo", s)

			s, ok = sc.next()
			assert.True(t, ok)
			assert.Equal(t, "foo", s)
		}

		if assert.True(t, sc.more(), "expected more") {
			s, ok := sc.peek()
			assert.True(t, ok)
			assert.Equal(t, "bar", s)

			s, ok = sc.next()
			assert.True(t, ok)
			assert.Equal(t, "bar", s)
		}

		if assert.False(t, sc.more(), "expected no more") {
			_, ok := sc.peek()
			assert.False(t, ok)

			_, ok = sc.next()
			assert.False(t, ok)
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		var c sliceCursor
		assert.False(t, c.more())

		_, ok := c.peek()
		assert.False(t, ok)

		_, ok = c.next()
		assert.False(t, ok)
	})
}
