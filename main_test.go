package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name string
		json string
		opts options
		want string
	}{
		{
			name: "basic table and array",
			json: `{"a":1,"b":{"c":"x","d":2.5},"e":[1,2,3]}`,
			want: "a = 1\ne = [1, 2, 3]\n\n[b]\nc = 'x'\nd = 2.5\n",
		},
		{
			name: "inline tables",
			json: `{"a":1,"b":{"c":"x","d":2.5}}`,
			opts: options{inlineTable: true},
			want: "a = 1\nb = {c = 'x', d = 2.5}\n",
		},
		{
			name: "integers stay integers and floats stay floats",
			json: `{"i":42,"f":4.2,"big":10000000000}`,
			want: "big = 10000000000\nf = 4.2\ni = 42\n",
		},
		{
			name: "backslash strings use literal quoting",
			json: `{"path":"C:\\Users\\foo","re":"\\d+"}`,
			want: "path = 'C:\\Users\\foo'\nre = '\\d+'\n",
		},
		{
			name: "bool and string",
			json: `{"ok":true,"name":"json"}`,
			want: "name = 'json'\nok = true\n",
		},
		{
			name: "omit-null drops nulls in objects and arrays",
			json: `{"a":1,"b":null,"c":[1,null,2]}`,
			opts: options{omitNull: true},
			want: "a = 1\nc = [1, 2]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			if err := convert(strings.NewReader(tt.json), &buf, tt.opts); err != nil {
				t.Fatalf("convert() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("convert() output mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestConvertErrors(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		opts    options
		wantSub string
	}{
		{name: "array root", json: `[1,2,3]`, wantSub: "top level must be an object"},
		{name: "scalar root", json: `"hi"`, wantSub: "top level must be an object"},
		{name: "number root", json: `42`, wantSub: "top level must be an object"},
		{name: "invalid json", json: `{`, wantSub: "decoding JSON"},
		{name: "null without omit-null", json: `{"a":null}`, wantSub: "no TOML representation"},
		{name: "null in array without omit-null", json: `{"a":[1,null]}`, wantSub: "no TOML representation"},
		{name: "integer out of int64 range", json: `{"big":123456789012345678901234567890}`, wantSub: "out of range for TOML"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := convert(strings.NewReader(tt.json), &buf, tt.opts)
			if err == nil {
				t.Fatalf("convert() expected error, got nil (output %q)", buf.String())
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("convert() error = %q, want substring %q", err.Error(), tt.wantSub)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	// json.Number "5" must become int64, "5.0" must become float64.
	var v any = map[string]any{
		"nested": []any{json.Number("7"), json.Number("7.5")},
	}
	nv, err := normalize(v, options{})
	if err != nil {
		t.Fatalf("normalize() error = %v", err)
	}
	got := nv.(map[string]any)["nested"].([]any)

	if n, ok := got[0].(int64); !ok || n != 7 {
		t.Errorf("normalize integer = %#v, want int64(7)", got[0])
	}
	if f, ok := got[1].(float64); !ok || f != 7.5 {
		t.Errorf("normalize float = %#v, want float64(7.5)", got[1])
	}
}
