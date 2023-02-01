package shon

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// Number holds numeric values that could be integers or floats.
type Number = json.Number

var _stringType = reflect.TypeOf("")

type decodeCtx struct {
	// Whether to use json.Number
	UseNumber bool
}

type decoder interface {
	Decode(decodeCtx, value) (reflect.Value, error)
}

func newDecoder(t reflect.Type) (decoder, error) {
	switch t.Kind() {
	case reflect.Pointer:
		e, err := newDecoder(t.Elem())
		if err != nil {
			return nil, err
		}
		return &ptrDecoder{
			t: t,
			e: e,
		}, nil

	case reflect.Slice:
		e, err := newDecoder(t.Elem())
		if err != nil {
			return nil, err
		}
		return &sliceDecoder{
			t: t,
			e: e,
		}, nil

	case reflect.Array:
		e, err := newDecoder(t.Elem())
		if err != nil {
			return nil, err
		}
		return &arrayDecoder{
			t:   t,
			e:   e,
			len: t.Len(),
		}, nil

	case reflect.Map:
		k, err := newDecoder(t.Key())
		if err != nil {
			return nil, err
		}
		v, err := newDecoder(t.Elem())
		if err != nil {
			return nil, err
		}
		return &mapDecoder{
			t: t,
			k: k,
			v: v,
		}, nil

	case reflect.Interface:
		if t.NumMethod() == 0 {
			return &anyDecoder{t: t}, nil
		}

	case reflect.Struct:
		return newStructDecoder(t)

	case reflect.String:
		return &stringDecoder{t: t}, nil

	case reflect.Bool:
		return &boolDecoder{t: t}, nil

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return &intDecoder{t: t, bits: t.Bits()}, nil

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return &uintDecoder{t: t, bits: t.Bits()}, nil

	case reflect.Float32, reflect.Float64:
		return &floatDecoder{t: t, bits: t.Bits()}, nil

	case reflect.Complex64, reflect.Complex128:
		return &complexDecoder{t: t, bits: t.Bits()}, nil
	}

	return nil, fmt.Errorf("unsupported type %v", t)
}

type ptrDecoder struct {
	t reflect.Type
	e decoder
}

func (d *ptrDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	v, err := d.e.Decode(ctx, t)
	if err != nil {
		return v, err
	}

	p := reflect.New(d.t).Elem()
	p.Set(v.Addr())
	return p, nil
}

type boolDecoder struct {
	t reflect.Type
}

func (d *boolDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	if t.t != boolType {
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	v := reflect.New(d.t).Elem()
	v.SetBool(t.b)
	return v, nil
}

type intDecoder struct {
	t    reflect.Type
	bits int
}

func (d *intDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	if t.t != scalarType {
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	i, err := strconv.ParseInt(t.s, 10, d.bits)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("bad %v: %w", d.t, err)
	}

	v := reflect.New(d.t).Elem()
	v.SetInt(i)
	return v, nil
}

type uintDecoder struct {
	t    reflect.Type
	bits int
}

func (d *uintDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	if t.t != scalarType {
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	u, err := strconv.ParseUint(t.s, 10, d.bits)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("bad %v: %w", d.t, err)
	}

	v := reflect.New(d.t).Elem()
	v.SetUint(u)
	return v, nil
}

type floatDecoder struct {
	t    reflect.Type
	bits int
}

func (d *floatDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	if t.t != scalarType {
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	f, err := strconv.ParseFloat(t.s, d.bits)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("bad %v: %w", d.t, err)
	}

	v := reflect.New(d.t).Elem()
	v.SetFloat(f)
	return v, nil
}

type complexDecoder struct {
	t    reflect.Type
	bits int
}

func (d *complexDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	if t.t != scalarType {
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	c, err := strconv.ParseComplex(t.s, d.bits)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("bad %v: %w", d.t, err)
	}

	v := reflect.New(d.t).Elem()
	v.SetComplex(c)
	return v, nil
}

type stringDecoder struct {
	t reflect.Type
}

func (d *stringDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	switch t.t {
	case scalarType, stringType:
		// ok
	default:
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	v := reflect.New(d.t).Elem()
	v.SetString(t.s)
	return v, nil
}

type sliceDecoder struct {
	t reflect.Type
	e decoder
}

func (d *sliceDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	if t.t != arrayType {
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	v := reflect.MakeSlice(d.t, 0, 0)
	for r := t.i.(reader); r.more(); {
		i, err := r.next()
		if err != nil {
			return v, err
		}

		e, err := d.e.Decode(ctx, i)
		if err != nil {
			return v, err
		}

		v = reflect.Append(v, e)
	}
	return v, nil
}

type arrayDecoder struct {
	t   reflect.Type
	e   decoder
	len int
}

func (d *arrayDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	if t.t != arrayType {
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	v := reflect.New(d.t).Elem()
	for r, idx := t.i.(reader), 0; r.more(); idx++ {
		if idx >= d.len {
			return v, fmt.Errorf("too many values: at most %v expected", d.len)
		}

		i, err := r.next()
		if err != nil {
			return v, err
		}

		e, err := d.e.Decode(ctx, i)
		if err != nil {
			return v, err
		}

		v.Index(idx).Set(e)
	}
	return v, nil
}

type mapDecoder struct {
	t reflect.Type
	k decoder
	v decoder
}

func (d *mapDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	if t.t != objectType {
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	v := reflect.MakeMap(d.t)
	for r := t.i.(objectReader); r.more(); {
		ks, vs, err := r.next()
		if err != nil {
			return v, err
		}

		key, err := d.k.Decode(ctx, value{t: scalarType, s: ks})
		if err != nil {
			return v, err
		}

		val, err := d.v.Decode(ctx, vs)
		if err != nil {
			return v, err
		}

		v.SetMapIndex(key, val)
	}
	return v, nil
}

type structDecoder struct {
	t            reflect.Type
	fields       []structField
	fieldsByName map[string]int // name => index into .fields
}

func newStructDecoder(t reflect.Type) (*structDecoder, error) {
	p := structDecoder{
		t:            t,
		fieldsByName: make(map[string]int),
	}
	for i := 0; i < t.NumField(); i++ {
		sf, ok, err := newStructField(i, t.Field(i))
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		for _, name := range sf.names {
			p.fieldsByName[name] = len(p.fields)
		}
		p.fields = append(p.fields, sf)
	}
	return &p, nil
}

func (d *structDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	if t.t != objectType {
		return reflect.Value{}, fmt.Errorf("expected %v, got %v", d.t, t.t)
	}

	v := reflect.New(d.t).Elem()
	for r := t.i.(objectReader); r.more(); {
		key, value, err := r.next()
		if err != nil {
			return v, err
		}

		fidx, ok := d.fieldsByName[key]
		if !ok {
			return v, fmt.Errorf("unknown field %q", key)
		}

		f := d.fields[fidx]
		fval, err := f.p.Decode(ctx, value)
		if err != nil {
			return v, err
		}

		v.Field(f.idx).Set(fval)
	}
	return v, nil
}

type structField struct {
	t   reflect.Type
	p   decoder
	idx int // index in struct.Field(i)

	// List of names this field accepts.
	// If 'shon:".."' is set, this contains just one.
	// Otherwise it contains the field name
	// and our guess at its kebab case version.
	names []string
}

func newStructField(idx int, f reflect.StructField) (structField, bool, error) {
	if !f.IsExported() {
		return structField{}, false, nil
	}

	fdec, err := newDecoder(f.Type)
	if err != nil {
		return structField{}, false, err
	}

	var names []string
	if name, ok := f.Tag.Lookup("shon"); ok {
		if name == "-" {
			return structField{}, false, nil
		}
		names = []string{name}
	} else {
		names = []string{f.Name, toKebab(f.Name)}
	}

	return structField{
		t:     f.Type,
		p:     fdec,
		idx:   idx,
		names: names,
	}, true, nil
}

type anyDecoder struct {
	t reflect.Type
}

func (d *anyDecoder) Decode(ctx decodeCtx, t value) (reflect.Value, error) {
	v := reflect.New(d.t).Elem()
	switch t.t {
	case nullType:
		v.Set(reflect.Zero(d.t))
	case boolType:
		v.Set(reflect.ValueOf(t.b))
	case stringType:
		v.Set(reflect.ValueOf(t.s))
	case scalarType:
		if t.num {
			if arg := t.s; ctx.UseNumber {
				v.Set(reflect.ValueOf(json.Number(arg)))
			} else {
				if i, err := strconv.ParseInt(arg, 10, 64); err == nil {
					v = reflect.ValueOf(int(i))
				} else if f, err := strconv.ParseFloat(arg, 64); err == nil {
					v = reflect.ValueOf(f)
				} else {
					// This is impossible unless there's a
					// bug in isNumeric.
					return v, fmt.Errorf("bad number %q", arg)
				}
			}
		} else {
			v.Set(reflect.ValueOf(t.s))
		}
	case arrayType:
		items := reflect.MakeSlice(reflect.SliceOf(d.t), 0, 0)
		for r := t.i.(reader); r.more(); {
			i, err := r.next()
			if err != nil {
				return v, err
			}

			e, err := d.Decode(ctx, i)
			if err != nil {
				return v, err
			}

			items = reflect.Append(items, e)
		}
		v.Set(items)

	case objectType:
		m := reflect.MakeMap(reflect.MapOf(_stringType, d.t))
		for r := t.i.(objectReader); r.more(); {
			key, vs, err := r.next()
			if err != nil {
				return v, err
			}

			val, err := d.Decode(ctx, vs)
			if err != nil {
				return v, err
			}

			m.SetMapIndex(reflect.ValueOf(key), val)
		}
		v.Set(m)

	default:
		return v, fmt.Errorf("unexpected %v", t.t)
	}

	return v, nil
}

type emptyArrayReader struct{}

var _emptyArray reader = (*emptyArrayReader)(nil)

func (*emptyArrayReader) more() bool { return false }

func (*emptyArrayReader) next() (value, error) {
	return _invalid, io.EOF
}

type emptyObjectReader struct{}

var _emptyObject objectReader = (*emptyObjectReader)(nil)

func (*emptyObjectReader) more() bool { return false }

func (*emptyObjectReader) next() (string, value, error) {
	return "", _invalid, io.EOF
}
