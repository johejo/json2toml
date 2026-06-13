# json2toml

A tiny CLI that converts JSON (stdin) to TOML (stdout).

## Install

```sh
go install github.com/johejo/json2toml@latest
```

## Usage

```sh
json2toml [-inline-table] [-omit-null] [--version] < input.json > output.toml
```

| Flag | Description |
| --- | --- |
| `-inline-table` | Emit nested tables as inline tables (`{ key = value }`) instead of `[table]` headers. |
| `-omit-null` | Drop JSON `null` values instead of erroring. |
| `--version` | Print the version and exit. |

## Examples

```sh
$ echo '{"a":1,"b":{"c":"x","d":2.5}}' | json2toml
a = 1

[b]
c = 'x'
d = 2.5

$ echo '{"a":1,"b":{"c":"x","d":2.5}}' | json2toml -inline-table
a = 1
b = {c = 'x', d = 2.5}
```

## Notes

- Numbers keep their JSON form: `42` stays an integer, `42.0` stays a float.
- TOML integers are 64-bit signed; an integer outside that range is an error.
- The JSON top level must be an object (TOML documents are tables).
- TOML has no `null`, so `null` is an error unless `-omit-null` is given.

## License

[MIT](LICENSE)
