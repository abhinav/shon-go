# shon

shon is a Go library for parsing the SHON format.

## What is SHON?

SHON (**Sh**ell **O**bject **N**otation) is a notation
for expressing complex objects at the command line.

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

## Credits

The SHON format was invented by [Kris Kowal](https://github.com/kriskowal).

## License

This project is made available under the MIT License.
