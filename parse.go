package shon

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func Parse(args []string, v any) error {
	dst := reflect.ValueOf(v)
	if dst.Kind() != reflect.Pointer {
		panic("must be a pointer")
	}
	dec, err := newDecoder(dst.Type().Elem())
	if err != nil {
		return err
	}

	cursor := sliceCursor{args: args}
	p := parser{
		cursor: &cursor,
	}

	val, err := p.value()
	if err != nil {
		return err
	}

	res, err := dec.Decode(decodeCtx{
		// TODO: Allow setting this.
		UseNumber: false,
	}, val)
	if err != nil {
		return err
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
		return _invalid, errors.New("expected value")
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
		return _invalid, errors.New("expected value")
	case "[]":
		return arrayValue(_emptyArray), nil
	case "[--]":
		return objectValue(_emptyObject), nil
	case "-t", "-f":
		return boolValue(arg == "-t"), nil
	case "-n":
		return _null, nil
	case "-u":
		return _undefined, nil
	case "--":
		return p.string()
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

func (p *parser) string() (value, error) {
	v, ok := p.next()
	if !ok {
		return _invalid, errors.New("expected string")
	}
	return stringValue(v), nil
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
