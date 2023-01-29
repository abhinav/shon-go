package main

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var got bytes.Buffer
		require.NoError(t, run(&got, []string{"foo"}))
		assert.JSONEq(t, `"foo"`, got.String())
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		err := run(io.Discard, []string{"]"})
		assert.ErrorContains(t, err, "expected a value")
	})
}
