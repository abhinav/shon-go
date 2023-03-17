// playground implements a WASM module
// that can be used to experiment with SHON.
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/google/shlex"
	"go.abhg.dev/shon"
)

func main() {
	js.Global().Set("shon2json", js.FuncOf(
		func(this js.Value, args []js.Value) any {
			if len(args) != 1 {
				return js.ValueOf(fmt.Errorf("expected 1 argument, got %d", len(args)))
			}

			prompt := args[0].String()
			res, err := shon2json(prompt)
			if err != nil {
				return (&result{Err: err}).Encode()
			}
			return (&result{JSON: res}).Encode()
		}))
	select {}
}

type result struct {
	JSON string
	Err  error
}

func (r *result) Encode() js.Value {
	result := make(map[string]any)
	if r.Err != nil {
		result["error"] = r.Err.Error()
	}
	if r.JSON != "" {
		result["json"] = r.JSON
	}
	return js.ValueOf(result)
}

func shon2json(prompt string) (string, error) {
	args, err := shlex.Split(prompt)
	if err != nil {
		return "", fmt.Errorf("split shell: %w", err)
	}

	var result any
	if err := shon.Parse(args, &result); err != nil {
		return "", fmt.Errorf("parse SHON: %w", err)
	}

	bs, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}

	return string(bs), nil
}
