package shon

import (
	"fmt"
)

type valueType int

const (
	invalidType valueType = iota
	nullType              // -n
	boolType              // -t or -f
	stringType            // explicitly a string
	scalarType            // string, int, float, complex
	arrayType             // [], [ ... ]
	objectType            // [--], [ --k v ... ]
)

func (t valueType) String() string {
	switch t {
	case invalidType:
		return "invalid"
	case nullType:
		return "null"
	case boolType:
		return "bool"
	case stringType:
		return "string"
	case scalarType:
		return "scalar"
	case arrayType:
		return "array"
	case objectType:
		return "object"
	default:
		return fmt.Sprintf("valueType(%d)", int(t))
	}
}

type reader interface {
	more() bool
	next() (value, error)
}

type objectReader interface {
	more() bool
	next() (string, value, error)
}

type value struct {
	t valueType
	b bool   // set if boolType
	s string // set if scalarType
	i any    // reader if arrayType, objectReader if objectType

	num bool // whether numeric if scalarType
}

var (
	_invalid = value{t: invalidType}
	_null    = value{t: nullType}
)

func stringValue(s string) value {
	return value{t: stringType, s: s}
}

func boolValue(b bool) value {
	return value{t: boolType, b: b}
}

func arrayValue(r reader) value {
	return value{t: arrayType, i: r}
}

func objectValue(r objectReader) value {
	return value{t: objectType, i: r}
}
