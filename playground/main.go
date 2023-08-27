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

			var req request
			req.Decode(args[0])
			res, err := shon2json(req)
			if err != nil {
				return (&result{Err: err}).Encode()
			}
			return (&result{JSON: res}).Encode()
		}))
	select {}
}

type request struct {
	Prompt string
	Object bool
}

func (r *request) Decode(v js.Value) {
	r.Prompt = v.Get("prompt").String()
	r.Object = v.Get("object").Bool()
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

func shon2json(req request) (string, error) {
	args, err := shlex.Split(req.Prompt)
	if err != nil {
		return "", fmt.Errorf("split shell: %w", err)
	}

	parser := shon.Parse
	if req.Object {
		parser = shon.ParseObject
	}

	var result any
	if err := parser(args, &result); err != nil {
		return "", fmt.Errorf("parse SHON: %w", err)
	}

	bs, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}

	return string(bs), nil
}
