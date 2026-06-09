// Command json2toml converts JSON read from stdin into TOML written to stdout.
//
// With -inline-table, nested tables are emitted as inline tables
// ({ key = value }) instead of the default [table] header style.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "json2toml:", err)
		os.Exit(1)
	}
}

// options controls how convert renders TOML.
type options struct {
	inlineTable bool // emit nested tables as inline tables
	omitNull    bool // drop JSON null values instead of erroring
}

func run() error {
	var opts options
	flag.BoolVar(&opts.inlineTable, "inline-table", false, "emit nested tables as inline tables")
	flag.BoolVar(&opts.omitNull, "omit-null", false, "drop JSON null values instead of erroring")
	flag.Parse()
	return convert(os.Stdin, os.Stdout, opts)
}

// convert reads JSON from r and writes the equivalent TOML to w.
func convert(r io.Reader, w io.Writer, opts options) error {
	dec := json.NewDecoder(r)
	dec.UseNumber() // keep numbers as json.Number to preserve int vs float

	var v any
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}

	if _, ok := v.(map[string]any); !ok {
		return fmt.Errorf("the JSON top level must be an object, got %T", v)
	}
	v, err := normalize(v, opts)
	if err != nil {
		return err
	}

	enc := toml.NewEncoder(w)
	if opts.inlineTable {
		enc.SetTablesInline(true)
	}
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding TOML: %w", err)
	}
	return nil
}

// normalize prepares a decoded JSON value for TOML encoding: it converts
// json.Number to int64/float64 and rejects (or drops) values TOML cannot
// represent, such as null.
func normalize(v any, opts options) (any, error) {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if val == nil {
				if opts.omitNull {
					delete(x, k)
					continue
				}
				return nil, fmt.Errorf("null value at key %q has no TOML representation (use -omit-null to drop nulls)", k)
			}
			nv, err := normalize(val, opts)
			if err != nil {
				return nil, err
			}
			x[k] = nv
		}
		return x, nil
	case []any:
		out := make([]any, 0, len(x))
		for _, val := range x {
			if val == nil {
				if opts.omitNull {
					continue
				}
				return nil, fmt.Errorf("null value in array has no TOML representation (use -omit-null to drop nulls)")
			}
			nv, err := normalize(val, opts)
			if err != nil {
				return nil, err
			}
			out = append(out, nv)
		}
		return out, nil
	case json.Number:
		return normalizeNumber(x)
	default:
		return v, nil
	}
}

// normalizeNumber converts a json.Number to int64 or float64. TOML integers are
// 64-bit signed, so an integer literal outside that range is an error rather
// than a lossy float conversion.
func normalizeNumber(n json.Number) (any, error) {
	s := n.String()
	if !strings.ContainsAny(s, ".eE") {
		i, err := n.Int64()
		if err != nil {
			return nil, fmt.Errorf("integer %s is out of range for TOML (64-bit signed integers only)", s)
		}
		return i, nil
	}
	f, err := n.Float64()
	if err != nil {
		return nil, fmt.Errorf("invalid number %s: %w", s, err)
	}
	return f, nil
}
