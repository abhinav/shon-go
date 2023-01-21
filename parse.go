package shon

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func ParseAny(args []string) (any, error) {
	p := parser{args: args}
	return p.value()
}

type parser struct {
	args []string
	pos  int
}

func (p *parser) more() bool {
	return p.pos < len(p.args)
}

func (p *parser) next() (s string, ok bool) {
	if !p.more() {
		return "", false
	}
	arg := p.args[p.pos]
	p.pos++
	return arg, true
}

func (p *parser) peek() (s string, ok bool) {
	if p.more() {
		return p.args[p.pos], true
	}
	return "", false
}

func (p *parser) value( /* TODO type hint */ ) (any, error) {
	arg, ok := p.next()
	if !ok {
		return nil, errors.New("expected value")
	}
	return p.valueFrom(arg)
}

func (p *parser) valueFrom(arg string) (any, error) {
	switch arg {
	case "":
		return arg, nil
	case "[":
		return p.arrayOrObject()
	case "]":
		return "", errors.New("expected value")
	case "[]":
		return []any{}, nil
	case "[--]":
		return map[string]any{}, nil
	case "-t":
		return true, nil
	case "-f":
		return false, nil
	case "-n", "-u":
		return nil, nil
	case "--":
		return p.string()
	}

	if n, err := strconv.Atoi(arg); err == nil {
		return n, nil
	}

	if arg[0] == '-' {
		return nil, fmt.Errorf("unexpected flag %q", arg)
	}

	return arg, nil
}

func (p *parser) arrayOrObject() (any, error) {
	arg, ok := p.peek()
	if !ok {
		return nil, errors.New("expected an array item, an object key, or ']'")
	}

	if arg == "]" {
		_, _ = p.next() // drop the values
		return []any{}, nil
	}

	if arg != "--" && strings.HasPrefix(arg, "--") {
		return p.object()
	}

	return p.array()
}

func (p *parser) array() (any, error) {
	var items []any
	for {
		arg, ok := p.peek()
		if !ok {
			return "", errors.New("expected an array item or ']'")
		}
		if arg == "]" {
			break
		}

		v, err := p.value()
		if err != nil {
			return items, err
		}
		items = append(items, v)
	}
	return items, nil
}

func (p *parser) object() (any, error) {
	fields := make(map[string]any)
	for {
		arg, ok := p.next()
		if !ok {
			return nil, errors.New("expected an object key")
		}
		if arg == "]" {
			break
		}

		if arg == "--" || !strings.HasPrefix(arg, "--") {
			return nil, fmt.Errorf("expected object key, got %q", arg)
		}

		key := arg[2:]
		var (
			value any
			err   error
		)
		if idx := strings.IndexByte(key, '='); idx >= 0 {
			value, err = p.valueFrom(key[idx+1:])
			key = key[:idx]
		} else {
			value, err = p.value()
		}
		if err != nil {
			return "", err // TODO: more context
		}
		fields[key] = value

	}
	return fields, nil
}

func (p *parser) string() (string, error) {
	v, ok := p.next()
	if !ok {
		return "", errors.New("expected string")
	}
	return v, nil
}
