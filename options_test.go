package shon

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildParseOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give []ParseOption
		want parseOptions
	}{
		{
			desc: "none",
			want: parseOptions{useNumber: false},
		},
		{
			desc: "use number",
			give: []ParseOption{
				UseNumber(true),
			},
			want: parseOptions{useNumber: true},
		},
		{
			desc: "use number override",
			give: []ParseOption{
				UseNumber(true),
				UseNumber(false),
			},
			want: parseOptions{useNumber: false},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			got := buildParseOptions(tt.give...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseOption_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give ParseOption
		want string
	}{
		{UseNumber(false), "UseNumber(false)"},
		{UseNumber(true), "UseNumber(true)"},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, fmt.Sprint(tt.give))
		})
	}
}
