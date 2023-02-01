// shon is a program that accepts SHON input on the command line
// and prints an equivalent JSON object to stdout.
//
// Install it by running:
//
//	go install go.abhg.dev/shon/cmd/shon@latest
package main

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"go.abhg.dev/shon"
)

func main() {
	log.SetFlags(0)
	if err := run(os.Stdout, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(stdout io.Writer, args []string) error {
	var x any
	if err := shon.Parse(args, &x); err != nil {
		return err
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(x)
}
