package shon

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_nonPointer(t *testing.T) {
	t.Parallel()

	err := Parse([]string{"foo"}, 42)
	assert.ErrorContains(t, err, "must be a pointer")
}

func ptrOf[T any](v T) *T {
	return &v
}

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give []string
		want any
	}{
		{
			desc: "slice/string",
			give: []string{"[", "foo", "bar", "]"},
			want: []string{"foo", "bar"},
		},
		{
			desc: "slice/int-as-string",
			give: []string{"[", "42", "100", "]"},
			want: []string{"42", "100"},
		},
		{
			desc: "array/full",
			give: []string{"[", "1", "2", "3", "]"},
			want: [3]int{1, 2, 3},
		},
		{
			desc: "array/partial",
			give: []string{"[", "1", "2", "]"},
			want: [3]int{1, 2, 0},
		},
		{
			desc: "nested pointer",
			give: []string{"foo"},
			want: ptrOf(ptrOf(ptrOf(ptrOf("foo")))),
		},
		{
			desc: "struct with array",
			give: []string{
				"[",
				"--name", "foo",
				"--items", "[", "1", "2", "3", "]",
				"]",
			},
			want: struct {
				Name  string
				Items []int
			}{
				Name:  "foo",
				Items: []int{1, 2, 3},
			},
		},
		{
			desc: "map/string",
			give: []string{
				"[",
				"--foo", "bar",
				"--baz", "qux",
				"]",
			},
			want: map[string]string{"foo": "bar", "baz": "qux"},
		},
		{
			desc: "map/int",
			give: []string{
				"[",
				"--1", "bar",
				"--2", "qux",
				"]",
			},
			want: map[int]string{1: "bar", 2: "qux"},
		},
		{
			desc: "map/with-equals",
			give: []string{
				"[",
				"--foo=bar",
				"--baz=--", "qux",
				"]",
			},
			want: map[string]string{"foo": "bar", "baz": "qux"},
		},
		{
			desc: "bool/true",
			give: []string{"-t"},
			want: true,
		},
		{
			desc: "bool/false",
			give: []string{"-f"},
			want: false,
		},
		{
			desc: "uint",
			give: []string{"42"},
			want: uint(42),
		},
		{
			desc: "float",
			give: []string{"42"},
			want: float64(42),
		},
		{
			desc: "complex",
			give: []string{"1+2i"},
			want: complex(1, 2),
		},
		{
			desc: "struct/default field names",
			give: []string{"[", "--foo", "bar", "--baz-qux", "quux", "]"},
			want: struct {
				Foo    string
				BazQux string
			}{
				Foo:    "bar",
				BazQux: "quux",
			},
		},
		{
			desc: "struct/override field names",
			give: []string{"[", "--foo", "bar", "--bax", "quux", "]"},
			want: struct {
				Foo    string
				BazQux string `shon:"bax"`
			}{
				Foo:    "bar",
				BazQux: "quux",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			got := reflect.New(reflect.TypeOf(tt.want))
			require.NoError(t, Parse(tt.give, got.Interface()))
			assert.Equal(t, tt.want, got.Elem().Interface())
		})
	}
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
			give: []string{"4.2"},
			want: 4.2,
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

func TestParseObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give []string
		want any
	}{
		{
			desc: "no arguments/empty struct",
			give: []string{},
			want: struct{}{},
		},
		{
			desc: "no arguments/non-empty struct",
			give: []string{},
			want: struct{ Foo string }{},
		},
		{
			desc: "simple arguments",
			give: []string{
				"--foo", "42",
				"--bar", "--", "-t",
			},
			want: struct {
				Foo int8
				Bar string
			}{
				Foo: 42,
				Bar: "-t",
			},
		},
		{
			desc: "nested list and object",
			give: []string{
				"--items", "[", "1", "2", "3", "]",
				"--pairs", "[", "--foo", "bar", "--baz", "qux", "]",
			},
			want: struct {
				Items []int
				Pairs map[string]string
			}{
				Items: []int{1, 2, 3},
				Pairs: map[string]string{
					"foo": "bar",
					"baz": "qux",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			got := reflect.New(reflect.TypeOf(tt.want))
			require.NoError(t, ParseObject(tt.give, got.Interface()))
			assert.Equal(t, tt.want, got.Elem().Interface())
		})
	}
}

func TestParse_decodeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc    string
		give    []string
		into    any
		wantErr string
	}{
		{
			desc:    "pointer to bad int",
			give:    []string{"foo"},
			into:    ptrOf(42),
			wantErr: "bad int",
		},
		{
			desc:    "unexpected bool",
			give:    []string{"foo"},
			into:    true,
			wantErr: "expected bool, got scalar",
		},
		{
			desc:    "unexpected int",
			give:    []string{"--", "42"},
			into:    int(0),
			wantErr: "expected int, got string",
		},
		{
			desc:    "bad int",
			give:    []string{"foo"},
			into:    int(0),
			wantErr: "bad int: ",
		},
		{
			desc:    "int overflow",
			give:    []string{"139831983198318"},
			into:    int8(0),
			wantErr: "value out of range",
		},
		{
			desc:    "unexpected uint",
			give:    []string{"--", "42"},
			into:    uint(0),
			wantErr: "expected uint, got string",
		},
		{
			desc:    "bad uint",
			give:    []string{"foo"},
			into:    uint(0),
			wantErr: "bad uint: ",
		},
		{
			desc:    "uint overflow",
			give:    []string{"139831983198318"},
			into:    uint8(0),
			wantErr: "value out of range",
		},
		{
			desc:    "unexpected float",
			give:    []string{"--", "42"},
			into:    float32(0),
			wantErr: "expected float32, got string",
		},
		{
			desc:    "bad float",
			give:    []string{"foo"},
			into:    float64(0),
			wantErr: "bad float",
		},
		{
			desc:    "unexpected complex",
			give:    []string{"--", "42"},
			into:    complex64(0),
			wantErr: "expected complex64, got string",
		},
		{
			desc:    "bad complex",
			give:    []string{"foo"},
			into:    complex128(0),
			wantErr: "bad complex",
		},
		{
			desc:    "unexpected string",
			give:    []string{"[", "]"},
			into:    "",
			wantErr: "expected string, got array",
		},
		{
			desc:    "unexpected slice",
			give:    []string{"[--]"},
			into:    []string{},
			wantErr: "expected []string, got object",
		},
		{
			desc:    "incorrect slice item",
			give:    []string{"[", "--", "]"},
			into:    []int{},
			wantErr: "expected int, got string",
		},
		{
			desc:    "bad slice item",
			give:    []string{"[", "foo", "]"},
			into:    []int{},
			wantErr: "bad int",
		},
		{
			desc:    "unexpected array",
			give:    []string{"[--]"},
			into:    [1]string{},
			wantErr: "expected [1]string, got object",
		},
		{
			desc:    "array/too many items",
			give:    []string{"[", "foo", "bar", "baz", "]"},
			into:    [2]string{},
			wantErr: "too many values",
		},
		{
			desc:    "unexpected map",
			give:    []string{"[]"},
			into:    map[string]string{},
			wantErr: "expected map[string]string, got array",
		},
		{
			desc:    "bad key",
			give:    []string{"[", "--foo", "bar", "]"},
			into:    map[int]string{},
			wantErr: "bad int",
		},
		{
			desc:    "bad value",
			give:    []string{"[", "--foo", "bar", "]"},
			into:    map[string]int{},
			wantErr: "bad int",
		},
		{
			desc:    "unexpected struct",
			give:    []string{"[]"},
			into:    struct{}{},
			wantErr: "expected struct {}, got array",
		},
		{
			desc:    "unexpected field",
			give:    []string{"[", "--foo", "42", "]"},
			into:    struct{}{},
			wantErr: `unknown field "foo"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			got := reflect.New(reflect.TypeOf(tt.into))
			err := Parse(tt.give, got.Interface())
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestParse_parseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc    string
		give    []string
		wantErr string
	}{
		{
			desc:    "empty",
			wantErr: "expected a value",
		},
		{
			desc:    "too many arguments",
			give:    []string{"42", "1"},
			wantErr: `unexpected arguments: ["1"]`,
		},
		{
			desc:    "unexpected ']'",
			give:    []string{"]"},
			wantErr: `expected a value, got "]"`,
		},
		{
			desc:    "end after string",
			give:    []string{"--"},
			wantErr: "unexpected end of input",
		},
		{
			desc:    "unexpected flag",
			give:    []string{"[", "42", "--foo", "]"},
			wantErr: `unexpected flag "--foo"`,
		},
		{
			desc:    "unclosed array",
			give:    []string{"["},
			wantErr: "expected an array item, an object key",
		},
		{
			desc:    "missing object key",
			give:    []string{"[", "--foo", "bar", "baz"},
			wantErr: `expected object key, got "baz"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			var v any
			err := Parse(tt.give, &v)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestIsNumeric(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want bool
	}{
		{"", false},
		{"x42", false},
		{"42", true},
		{"+42", true},
		{"-42", true},
		{"42x", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, isNumeric(tt.give))
		})
	}
}

func TestToKebab(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want string
	}{
		{"", ""},
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
