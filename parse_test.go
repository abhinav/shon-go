package shon

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	var x struct {
		Name  string
		Items []int
	}
	require.NoError(t,
		Parse([]string{
			"[",
			"--name", "foo",
			"--items", "[", "1", "2", "3", "]",
			"]",
		}, &x))
	assert.Equal(t, "foo", x.Name)
	assert.Equal(t, []int{1, 2, 3}, x.Items)
}

func TestParseAny(t *testing.T) {
	t.Parallel()

	type array = []any
	type object = map[string]any

	tests := []struct {
		give []string
		want interface{}
	}{
		{
			give: []string{"[", "--hello", "World", "]"},
			want: object{"hello": "World"},
		},
		{
			give: []string{"[", "beep", "boop", "]"},
			want: array{"beep", "boop"},
		},
		{
			give: []string{"[", "1", "2", "3", "]"},
			want: array{1, 2, 3},
		},
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

			var got any
			err := Parse(tt.give, &got)
			require.NoError(t, err, "Parse(%q)", tt.give)
			assert.Equal(t, tt.want, got, "Parse(%q)", tt.give)
		})
	}
}

func TestToKebab(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want string
	}{
		{"Foo", "foo"},
		{"FooBar", "foo-bar"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, toKebab(tt.give))
		})
	}
}
