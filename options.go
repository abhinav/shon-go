package shon

import "fmt"

// ParseOption customizes the behavior of [Parse].
type ParseOption interface{ applyParseOption(*parseOptions) }

type parseOptions struct {
	useNumber      bool
	implicitObject bool
}

func buildParseOptions(opts ...ParseOption) parseOptions {
	var po parseOptions
	for _, o := range opts {
		o.applyParseOption(&po)
	}
	return po
}

// UseNumber specifies whether the decoder should read numeric values
// as [Number] objects for fields of the type 'any'.
//
// If false, the decoder will attempt to parse numeric values as
// an int64 or float64, and place that there instead.
//
// Defaults to false.
func UseNumber(b bool) ParseOption {
	return useNumberOption(b)
}

type useNumberOption bool

func (o useNumberOption) String() string {
	return fmt.Sprintf("UseNumber(%v)", bool(o))
}

func (o useNumberOption) applyParseOption(opts *parseOptions) {
	opts.useNumber = bool(o)
}

// implicitObject specifies that Parse should assume it's inside an object
// at the top level.
// With this,
//
//	foo --bar baz --qux -t
//
// Is treated as if it was:
//
//	foo [ --bar baz --qux -t ]
//
// This option is not public -- users should use the ParseObject function.
func implicitObject(b bool) ParseOption {
	return implicitObjectOption(b)
}

type implicitObjectOption bool

func (o implicitObjectOption) String() string {
	return fmt.Sprintf("implicitObject(%v)", bool(o))
}

func (o implicitObjectOption) applyParseOption(opts *parseOptions) {
	opts.implicitObject = bool(o)
}
