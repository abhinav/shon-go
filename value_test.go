package shon

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give valueType
		want string
	}{
		{invalidType, "invalid"},
		{nullType, "null"},
		{undefinedType, "undefined"},
		{boolType, "bool"},
		{stringType, "string"},
		{scalarType, "scalar"},
		{arrayType, "array"},
		{objectType, "object"},
		{valueType(-42), "valueType(-42)"},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}
