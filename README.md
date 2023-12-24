# shon

[![Go Reference](https://pkg.go.dev/badge/go.abhg.dev/shon.svg)](https://pkg.go.dev/go.abhg.dev/shon)
[![Go](https://github.com/abhinav/shon-go/actions/workflows/ci.yml/badge.svg)](https://github.com/abhinav/shon-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/abhinav/shon-go/branch/main/graph/badge.svg?token=SG8rMD57rN)](https://codecov.io/gh/abhinav/shon-go)

shon is a Go library for parsing the SHON format.
Read more about SHON below.

## Playground

A web-based playground to experiment with SHON is available at:
<https://abhinav.github.io/shon-go/playground>.

## Installation

```bash
go get go.abhg.dev/shon@latest
```

## Usage

Use the `shon.Parse` function
to decode a list of arguments into a Go structure.

```go
var cfg struct {
  OutDir string   `shon:"out"`
  Inputs []string `shon:"input"`
}

err := shon.Parse(os.Args[1:], &cfg)
// ...
```

The above will parse input like:

```bash
[ --out mydir/ --input [ foo "bar baz" qux ] ]
```

## What is SHON?

SHON (pronounced 'shawn') is short for **Sh**ell **O**bject **N**otation.
It is a notation
for expressing complex objects at the command line.
Because it is intended to be used on the command line,
it aims to reduce extraneous commas and brackets.

All JSON objects can be expressed via SHON,
typically in a format that is easier to specify on the command line.

| JSON                 | SHON                |
|----------------------|---------------------|
| `{"hello": "World"}` | `[ --hello World ]` |
| `["beep", "boop"]`   | `[ beep boop ]`     |
| `[1, 2, 3]`          | `[ 1 2 3 ]`         |
| `[]`                 | `[ ]` or `[]`       |
| `{"a": 10, b: 20}`   | `[ --a 10 --b 20 ]` |
| `{}`                 | `[--]`              |
| `1`                  | `1`                 |
| `-1`                 | `-1`                |
| `1e3`                | `1e3`               |
| `"hello"`            | `hello`             |
| `"hello world"`      | `'hello world'`     |
| `"10"`               | `-- 10`             |
| `"-10"`              | `-- -10`            |
| `"-"`                | `-- -`              |
| `"--"`               | `-- --`             |
| `true`               | `-t`                |
| `false`              | `-f`                |
| `null`               | `-n`                |

### Implementations

- An implementation of SHON for Javascript is available at
  <https://github.com/borkshop/shon>.

## SHON Reference

- Scalars (strings and numbers) are expressed as-is.
  If the scalar is numeric, it is read as a number.

    ```bash
    foo  # == "foo"
    42   # == 42
    ```

- Strings with spaces in them must be quoted.

    ```bash
    "foo bar"  # == "foo bar"
    ```

- Strings may be preceded by a `--` to escape them
  and interpret them verbatim.

    ```bash
    -- foo        # == "foo"
    -- "foo bar"  # == "foo bar"
    -- --         # == "--"
    ```

  The idiomatic way to express an arbitrary string
  stored in Bash variable `VAR` in SHON is:

    ```bash
    -- "$VAR"
    ```

- Booleans are expressed by `-t` for true, and `-f` for false.

    ```bash
    -t  # true
    -f  # false
    ```

- `null` is expressed as `-n`.

    ```bash
    -n  # null
    ```

- Arrays of values are expressed by wrapping the with `[`, `]`.


    ```bash
    [ 1 2 3 ]            # == [1, 2, 3]
    [ "foo bar" 42 -t ]  # == ["foo bar", 42, true]
    ```

- An empty array is a pair of `[`, `]` with nothing between them
  or the string `[]`.

    ```bash
    []   # == []
    [ ]  # == []
    ```

- Objects are expressed as key-value pairs, with keys prefixed with `--`.
  Keys are typically kebab-case.

    ```bash
    [ --id 42 --disable -t ]  # == {"id": 42, "disable": true}
    ```

- An empty object is expressed by the string `[--]`:

    ```bash
    [--]   # {}
    ```

## Credits

The SHON format was invented by [Kris Kowal](https://github.com/kriskowal).

## FAQ

### Why?

The project is only half-serious.
The other half is whimsy.

So, I guess because it was fun to write,
and to possibly infect others with the idea

## License

This project is made available under the BSD3 License.
