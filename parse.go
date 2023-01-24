package shon

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	_anyType      = reflect.TypeOf((*any)(nil)).Elem()
	_anySliceType = reflect.TypeOf([]any{})
	_anyMapType   = reflect.TypeOf(map[string]any{})
	_stringType   = reflect.TypeOf("")
)

func Parse(args []string, v any) error {
	dst := reflect.ValueOf(v)
	if dst.Kind() != reflect.Pointer {
		panic("must be a pointer")
	}

	cursor := sliceCursor{args: args}
	p := parse{
		cursor: &cursor,
	}
	return p.value(dst.Elem())
}

// Number is literal numeric value.
type Number = json.Number

// parse holds the state for a single shon parse.
type parse struct {
	cursor

	useNumber bool
}

// value reads the next value into dst.
func (p *parse) value(dst reflect.Value) error {
	arg, ok := p.next()
	if !ok {
		return errors.New("expected a value")
	}
	return p.valueFrom(dst, arg)
}

func (p *parse) valueFrom(dst reflect.Value, arg string) error {
	// Handle the nil pointer case first so that we don't incorrectly
	// assign a non-nil value to the pointer.
	switch arg {
	case "-n", "-u":
		switch dst.Kind() {
		case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		return fmt.Errorf("cannot assign nil to %v", dst.Type())
	}

	if dst.Kind() == reflect.Pointer {
		v := reflect.New(dst.Type().Elem())
		dst.Set(v)
		return p.valueFrom(v.Elem(), arg)
	}

	// The following should assume no pointer types.
	switch arg {
	case "":
		return p.stringFrom(dst, arg)
	case "[":
		return p.arrayOrObject(dst)
	case "]":
		return errors.New("expected a value")
	case "[]":
		switch dst.Kind() {
		case reflect.Slice:
			dst.Set(reflect.MakeSlice(dst.Type(), 0, 0))
			return nil

		case reflect.Array:
			// Zero out all values.
			z := reflect.Zero(dst.Type().Elem())
			for i := 0; i < dst.Len(); i++ {
				dst.Index(i).Set(z)
			}

		case reflect.Interface:
			if dst.NumMethod() == 0 {
				dst.Set(reflect.MakeSlice(_anySliceType, 0, 0))
				return nil
			}
		}

		return fmt.Errorf("cannot assign empty array to %v", dst.Type())

	case "[--]":
		switch dst.Kind() {
		case reflect.Struct:
			dst.Set(reflect.Zero(dst.Type()))
			return nil

		case reflect.Map:
			if t := dst.Type(); isObjectKey(t.Key()) {
				dst.Set(reflect.MakeMap(t))
				return nil
			}

		case reflect.Interface:
			if dst.NumMethod() == 0 {
				dst.Set(reflect.MakeMap(_anyMapType))
				return nil
			}
		}

		return fmt.Errorf("cannot assign empty map to %v", dst.Type())

	case "-t", "-f":
		b := arg == "-t"
		kind := dst.Kind()
		switch {
		case kind == reflect.Bool:
			dst.SetBool(b)
		case kind == reflect.Interface && dst.NumMethod() == 0:
			dst.Set(reflect.ValueOf(b))
		default:
			return fmt.Errorf("cannot assign bool to %v", dst.Type())
		}
		return nil

	case "--":
		return p.string(dst)
	}

	numeric := isNumeric(arg)
	if arg[0] == '-' && !numeric {
		return fmt.Errorf("unexpected flag %q", arg)
	}

	t := dst.Type()
	switch t.Kind() {
	case reflect.String:
		dst.SetString(arg)
		return nil

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		i, err := strconv.ParseInt(arg, 10, t.Bits())
		if err != nil {
			return fmt.Errorf("bad int for %v: %w", t, err)
		}
		dst.SetInt(i)
		return nil

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		i, err := strconv.ParseUint(arg, 10, t.Bits())
		if err != nil {
			return fmt.Errorf("bad uint for %v: %w", t, err)
		}
		dst.SetUint(i)
		return nil

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(arg, t.Bits())
		if err != nil {
			return fmt.Errorf("bad float for %v: %w", t, err)
		}
		dst.SetFloat(f)

	case reflect.Complex64, reflect.Complex128:
		c, err := strconv.ParseComplex(arg, t.Bits())
		if err != nil {
			return fmt.Errorf("bad complex for %v: %w", t, err)
		}
		dst.SetComplex(c)

	case reflect.Interface:
		if dst.NumMethod() == 0 {
			var v reflect.Value
			if numeric {
				if p.useNumber {
					v = reflect.ValueOf(Number(arg))
				} else {
					if i, err := strconv.ParseInt(arg, 10, 64); err == nil {
						v = reflect.ValueOf(int(i))
					} else if f, err := strconv.ParseFloat(arg, 64); err == nil {
						v = reflect.ValueOf(f)
					} else {
						return fmt.Errorf("bad number %q", err)
					}
				}
			} else {
				v = reflect.ValueOf(arg)
			}
			dst.Set(v)
			return nil
		}
	}

	return fmt.Errorf("cannot assign %q to %v", arg, t)
}

func (p *parse) arrayOrObject(dst reflect.Value) error {
	arg, ok := p.peek()
	if !ok {
		return errors.New(`expected an array item, object key, or "]"`)
	}

	if arg == "]" {
		// Treat [ ] the same as []
		return p.array(dst)
	}

	if arg != "--" && strings.HasPrefix(arg, "--") {
		return p.object(dst)
	}

	return p.array(dst)
}

func (p *parse) array(dst reflect.Value) error {
	var (
		elt   reflect.Type
		items reflect.Value
	)
	switch dst.Kind() {
	case reflect.Slice: // TODO: arrays?
		t := dst.Type()
		elt = t.Elem()
		items = reflect.MakeSlice(t, 0, 0)

	case reflect.Interface:
		if t := dst.Type(); t.NumMethod() != 0 {
			return fmt.Errorf("cannot parse array into %v", t)
		}
		elt = _anyType
		items = reflect.MakeSlice(_anySliceType, 0, 0)

	default:
		return fmt.Errorf("cannot parse array into %v", dst.Type())
	}

	for {
		arg, ok := p.peek()
		if !ok {
			return errors.New("expected an array item or ']'")
		}
		if arg == "]" {
			break
		}

		v := reflect.New(elt).Elem()
		if err := p.value(v); err != nil {
			return err
		}
		items = reflect.Append(items, v)
	}

	dst.Set(items)
	return nil
}

type objectReader interface {
	ValueType(key string) (reflect.Type, error)
	Set(key string, v reflect.Value) error
}

type mapReader struct {
	m reflect.Value

	toKey     func(string) (reflect.Value, error)
	valueType reflect.Type
}

func (d *mapReader) ValueType(string) (reflect.Type, error) {
	return d.valueType, nil
}

func (d *mapReader) Set(key string, v reflect.Value) error {
	k, err := d.toKey(key)
	if err != nil {
		return err
	}
	d.m.SetMapIndex(k, v)
	return nil
}

type structReader struct {
	v      reflect.Value
	fields map[string]reflect.Value
}

func (d *structReader) ValueType(k string) (reflect.Type, error) {
	f := d.v.FieldByName(k)
	if !f.IsValid() {
		return nil, fmt.Errorf("unknown field %q", k)
	}
	d.fields[k] = f
	return f.Type(), nil
}

func (d *structReader) Set(k string, v reflect.Value) error {
	f, ok := d.fields[k]
	if !ok {
		f := d.v.FieldByName(k)
		if !f.IsValid() {
			return fmt.Errorf("unknown field %q", k)
		}
	}

	f.Set(v)
	return nil
}

func (p *parse) object(dst reflect.Value) error {
	var dec objectReader
	switch dst.Kind() {
	case reflect.Struct:
		dec = &structReader{
			v:      dst,
			fields: make(map[string]reflect.Value),
		}

	case reflect.Map:
		mt := dst.Type()
		dst.Set(reflect.MakeMap(mt))
		md := mapReader{
			m:         dst,
			valueType: mt.Elem(),
		}
		switch kt := dst.Type().Key(); kt.Kind() {
		case reflect.String:
			// TODO: extract
			md.toKey = func(k string) (reflect.Value, error) {
				return reflect.ValueOf(k).Convert(kt), nil
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bits := kt.Bits()
			md.toKey = func(k string) (reflect.Value, error) {
				i, err := strconv.ParseInt(k, 10, bits)
				if err != nil {
					return reflect.Value{}, err
				}
				return reflect.ValueOf(i).Convert(kt), nil
			}

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bits := kt.Bits()
			md.toKey = func(k string) (reflect.Value, error) {
				i, err := strconv.ParseUint(k, 10, bits)
				if err != nil {
					return reflect.Value{}, err
				}
				return reflect.ValueOf(i).Convert(kt), nil
			}
		}
		dec = &md

	case reflect.Interface:
		if dst.NumMethod() != 0 {
			return fmt.Errorf("cannot parse object into %v", dst.Type())
		}
		m := reflect.MakeMap(_anyMapType)
		dst.Set(m)
		dec = &mapReader{
			m:         m,
			valueType: _anyType,
			toKey: func(k string) (reflect.Value, error) {
				return reflect.ValueOf(k), nil
			},
		}

	default:
		return fmt.Errorf("cannot parse object into %v", dst.Type())
	}

	for {
		arg, ok := p.next()
		if !ok {
			return errors.New("expected an object key")
		}
		if arg == "]" {
			break
		}

		if arg == "--" || !strings.HasPrefix(arg, "--") {
			return fmt.Errorf("expected object key, got %q", arg)
		}

		key := arg[2:]
		if idx := strings.IndexByte(key, '='); idx >= 0 {
			key, arg = key[:idx], key[idx+1:]

			vt, err := dec.ValueType(key)
			if err != nil {
				return err
			}

			v := reflect.New(vt).Elem()
			if err := p.valueFrom(v, arg); err != nil {
				return err
			}
			dec.Set(key, v)
		} else {
			vt, err := dec.ValueType(key)
			if err != nil {
				return err
			}

			v := reflect.New(vt).Elem()
			if err := p.value(v); err != nil {
				return err
			}
			dec.Set(key, v)
		}
	}

	return nil
}

func (p *parse) string(dst reflect.Value) error {
	s, ok := p.next()
	if !ok {
		return errors.New("expected string")
	}

	return p.stringFrom(dst, s)
}

func (p *parse) stringFrom(dst reflect.Value, s string) error {
	kind := dst.Kind()
	switch {
	case kind == reflect.String:
		dst.SetString(s)
	case kind == reflect.Interface && dst.NumMethod() == 0:
		dst.Set(reflect.ValueOf(s))
	default:
		return fmt.Errorf("cannot assign string to %v", dst.Type())
	}
	return nil
}

func isObjectKey(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
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
