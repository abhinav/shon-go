package shon

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAny(t *testing.T) {
	t.Parallel()

	type array = []any
	type object = map[string]any

	tests := []struct {
		give []string
		want interface{}
	}{
		{
			give: []string{""},
			want: "",
		},
		{
			give: []string{"+10"},
			want: 10,
		},
		{
			give: []string{"-10"},
			want: -10,
		},
		{
			give: []string{"10"},
			want: 10,
		},
		{
			give: []string{"--", "10"},
			want: "10",
		},
		{
			give: []string{"a"},
			want: "a",
		},
		{
			give: []string{"-t"},
			want: true,
		},
		{
			give: []string{"-f"},
			want: false,
		},
		{
			give: []string{"-n"},
			want: nil,
		},
		{
			give: []string{"[", "]"},
			want: array{},
		},
		{
			give: []string{"[]"},
			want: array{},
		},
		{
			give: []string{"[", "[]", "]"},
			want: array{array{}},
		},
		{
			give: []string{"[", "[", "hello", "world", "]", "]"},
			want: array{array{"hello", "world"}},
		},
		{
			give: []string{"[", "[", "[--]", "[--]", "]", "]"},
			want: array{array{object{}, object{}}},
		},
		{
			give: []string{"[", "[", "--key", "10", "]", "]"},
			want: array{object{"key": 10}},
		},
		{
			give: []string{"[", "[", "--key=10", "]", "]"},
			want: array{object{"key": 10}},
		},
		{
			give: []string{"[", "[", "--key=-10", "]", "]"},
			want: array{object{"key": -10}},
		},
		{
			give: []string{"[", "[", "--key=--", "--", "]", "]"},
			want: array{object{"key": "--"}},
		},
		{
			give: []string{"[--]"},
			want: object{},
		},
		{
			give: []string{"[", "--", "--", "]"},
			want: array{"--"},
		},
		{
			give: []string{"[", "+10", "]"},
			want: array{10},
		},
		{
			give: []string{"[", "--a", "+10", "]"},
			want: object{"a": 10},
		},
		{
			give: []string{"[", "--foo", "+10", "--bar", "20", "]"},
			want: object{"foo": 10, "bar": 20},
		},
		{
			give: []string{"[", "--foo=+10", "--bar=20", "]"},
			want: object{"foo": 10, "bar": 20},
		},
		{
			give: []string{"[", "--foo=+10", "--bar=--", "20", "]"},
			want: object{"foo": 10, "bar": "20"},
		},
		{
			give: []string{"[", "--foo=+10", "--bar", "--", "20", "]"},
			want: object{"foo": 10, "bar": "20"},
		},
		{
			give: []string{"[", "[", "hi", "]", "]"},
			want: array{array{"hi"}},
		},
		{
			give: []string{"[", "--xs", "[", "--ys", "y", "]", "]"},
			want: object{"xs": object{"ys": "y"}},
		},
		{
			give: []string{"--", "["},
			want: "[",
		},
		{
			give: []string{"--", "]"},
			want: "]",
		},
		{
			give: []string{"--", "[a"},
			want: "[a",
		},
		{
			give: []string{"--", "]b"},
			want: "]b",
		},
	}

	for idx, tt := range tests {
		tt := tt
		t.Run(fmt.Sprint(idx), func(t *testing.T) {
			t.Parallel()

			got, err := ParseAny(tt.give)
			require.NoError(t, err, "Parse(%q)", tt.give)
			assert.Equal(t, tt.want, got, "Parse(%q)", tt.give)
		})
	}
}
