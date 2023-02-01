package shon

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// Parse decodes args and stores the result into the value pointed to by v.
// If v is not a pointer, Parse returns an error.
//
// The type and structure of v informs how Parse will decode values.
// The following provides a listing of the type of v
// and what Parse will accept for it.
//
//   - bool: -t or -f
//   - any int, uint, float, or complex type:
//     a numeric value parsed to that type
//   - string: a value picked verbatim, or preceded by a '--' argument
//   - slice: a collection of values surrounded by '[', ']'
//   - array: up to as many values as the array has room for,
//     surrounded by '[', ']'
//   - map: key-value pairs where the key is prefixed with '--',
//     surrounded by '[', ']'
//   - struct: key-value where the key is an exported field name
//     in kebab-case and prefixed with '--',
//     and the whole object is surrounded by '[', ']'
//   - pointer types: parsed as the target type
//   - any or interface{}: accepts anything, see below for more
//
// # Parsing structs
//
// Structs are parsed by matching their field names to object keys.
// For example, given the struct:
//
//	type User struct {
//		FirstName string
//		LastName  string
//	}
//
// We can pass the following arguments:
//
//	[ --first-name Jack --last-name Sparrow ]
//
// Although these names are guessed implicitly,
// it is recommended to expliclty annotate fields with
// the shon:".." tag to specify names for those fields.
//
//	type User struct {
//		FirstName string `shon:"first-name"`
//		LastName  string `shon:"last-name"`
//	}
//
// # Parsing any value
//
// As a special case, a field of type any (interface{})
// will accept a complete value decoded into one of the following types:
//
//	bool
//	string
//	int64
//	float64
//	[]any
//	map[string]any
//
// If the [UseNumber] option is used, int64 and float64 above
// will be replaced with [Number].
func Parse(args []string, v any, opts ...ParseOption) error {
	dst := reflect.ValueOf(v)
	if dst.Kind() != reflect.Pointer {
		return errors.New("must be a pointer")
	}

	options := buildParseOptions(opts...)

	cur := sliceCursor{args: args}
	val, err := (&parser{cursor: &cur}).value()
	if err != nil {
		return err
	}

	dec, err := newDecoder(dst.Type().Elem())
	if err != nil {
		return err
	}

	res, err := dec.Decode(decodeCtx{UseNumber: options.useNumber}, val)
	if err != nil {
		return err
	}

	if cur.more() {
		return fmt.Errorf("unexpected arguments: %q", cur.args[cur.pos:])
	}

	dst.Elem().Set(res)
	return nil
}

type parser struct {
	cursor
}

func (p *parser) value() (value, error) {
	arg, ok := p.next()
	if !ok {
		return _invalid, errors.New("expected a value")
	}
	return p.valueFrom(arg)
}

func (p *parser) valueFrom(arg string) (value, error) {
	switch arg {
	case "":
		return stringValue(arg), nil
	case "[":
		return p.arrayOrObject()
	case "]":
		return _invalid, fmt.Errorf("expected a value, got %q", arg)
	case "[]":
		return arrayValue(_emptyArray), nil
	case "[--]":
		return objectValue(_emptyObject), nil
	case "-t", "-f":
		return boolValue(arg == "-t"), nil
	case "-n":
		return _null, nil
	case "--":
		v, ok := p.next()
		if !ok {
			return _invalid, errors.New("unexpected end of input, expected a string")
		}
		return stringValue(v), nil
	}

	numeric := isNumeric(arg)
	if arg[0] == '-' && !numeric {
		return _invalid, fmt.Errorf("unexpected flag %q", arg)
	}

	return value{
		t:   scalarType,
		s:   arg,
		num: numeric,
	}, nil
}

func (p *parser) arrayOrObject() (value, error) {
	arg, ok := p.peek()
	if !ok {
		return _invalid, errors.New("expected an array item, an object key, or ']'")
	}

	if arg == "]" {
		// Treat [ ] the same as []
		_, _ = p.next() // drop the value
		return arrayValue(_emptyArray), nil
	}

	if arg != "--" && strings.HasPrefix(arg, "--") {
		return objectValue(&cursorObjectReader{p: p}), nil
	}
	return arrayValue(&cursorArrayReader{p: p}), nil
}

type cursorArrayReader struct {
	p *parser

	last struct {
		arg string
		ok  bool
	}
}

func (r *cursorArrayReader) more() bool {
	r.last.arg, r.last.ok = r.p.next()
	return r.last.ok && r.last.arg != "]"
}

func (r *cursorArrayReader) next() (value, error) {
	if r.last.ok {
		return r.p.valueFrom(r.last.arg)
	}
	return r.p.value()
}

type cursorObjectReader struct {
	p *parser

	last struct {
		arg string
		ok  bool
	}
}

func (r *cursorObjectReader) more() bool {
	r.last.arg, r.last.ok = r.p.next()
	return r.last.ok && r.last.arg != "]"
}

func (r *cursorObjectReader) next() (string, value, error) {
	var (
		arg string
		ok  bool
	)
	if r.last.ok {
		arg, ok = r.last.arg, r.last.ok
	} else {
		arg, ok = r.p.next()
	}
	if !ok {
		return "", _invalid, errors.New("expected object key")
	}

	if arg == "--" || !strings.HasPrefix(arg, "--") {
		return "", _invalid, fmt.Errorf("expected object key, got %q", arg)
	}

	key := arg[2:]
	var (
		value value
		err   error
	)
	if idx := strings.IndexByte(key, '='); idx >= 0 {
		value, err = r.p.valueFrom(key[idx+1:])
		key = key[:idx]
	} else {
		value, err = r.p.value()
	}
	return key, value, err
}

func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}

	switch s[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+':
		// ok
	default:
		return false
	}

	for _, r := range s[1:] {
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+', '.', 'e', 'E':
			// ok
		default:
			return false
		}
	}

	return true
}

func toKebab(name string) string {
	if len(name) == 0 {
		return name
	}

	var tokens []string
	for len(name) > 0 {
		idx := strings.IndexFunc(name[1:], unicode.IsUpper) + 1
		if idx <= 0 {
			tokens = append(tokens, strings.ToLower(name))
			break
		}

		tokens = append(tokens, strings.ToLower(name[:idx]))
		name = name[idx:]
	}

	return strings.Join(tokens, "-")
}
