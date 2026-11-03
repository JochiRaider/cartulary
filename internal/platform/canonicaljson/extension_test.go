package canonicaljson_test

import (
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestExtensionCanonicalEncoding(t *testing.T) {
	for _, vector := range []struct {
		name  string
		value any
		want  string
	}{
		{"null", nil, "null\n"},
		{"empty object", map[string]any{}, "{}\n"},
		{"empty array", []any{}, "[]\n"},
		{"array order", []any{true, false, nil, "é"}, "[true,false,null,\"é\"]\n"},
		{"integer bounds", []any{int64(-9223372036854775808), int64(9223372036854775807), uint64(18446744073709551615)}, "[-9223372036854775808,9223372036854775807,18446744073709551615]\n"},
		{"UTF-8 order", map[string]any{"😀": 2, "\ue000": 1, "a": 3}, "{\"a\":3,\"\ue000\":1,\"😀\":2}\n"},
		{"scalar escaping", "<>&/\u2028\u2029\x00\x01\b\t\n\f\r\"\\", "\"<>&/\u2028\u2029\\u0000\\u0001\\b\\t\\n\\f\\r\\\"\\\\\"\n"},
		{"typed projection", struct {
			Value string `json:"value"`
		}{"<é>"}, "{\"value\":\"<é>\"}\n"},
	} {
		t.Run(vector.name, func(t *testing.T) {
			got, err := canonicaljson.MarshalExtension(vector.value)
			if err != nil || string(got) != vector.want {
				t.Fatalf("bytes = %q, error %v; want %q", got, err, vector.want)
			}
			if _, err := canonicaljson.DecodeExtension(got); err != nil {
				t.Fatalf("canonical bytes rejected: %v", err)
			}
		})
	}
	for _, invalid := range []any{
		json.Number("1.0"), json.Number("1e0"), json.Number("-0"), json.Number("+1"),
		float64(0.5), "\xff", []string{"\xff"}, map[string]string{"\xff": "value"},
		struct{ Value string }{"\xff"}, json.RawMessage(`{"duplicate":1,"duplicate":2}`),
		json.RawMessage(`"\ud800"`),
	} {
		if _, err := canonicaljson.MarshalExtension(invalid); err == nil {
			t.Errorf("invalid typed input %T accepted", invalid)
		}
	}
	for _, invalid := range []string{
		"{}", "{}\n\n", "{}\n!", " {}\n", "{ }\n", "\xef\xbb\xbf{}\n",
		"{\"b\":0,\"a\":0}\n", "{\"a\":0,\"a\":1}\n", "1.5\n", "1e0\n", "-0\n",
		"\"\\u0061\"\n", "\"\\ud800\"\n", "\"\xff\"\n", "\"\\u2028\"\n",
	} {
		if _, err := canonicaljson.DecodeExtension([]byte(invalid)); err == nil {
			t.Errorf("invalid original bytes %q accepted", invalid)
		}
	}
	got, err := canonicaljson.CanonicalizeExtension([]byte(`{"z":"\u003c","a":9223372036854775807}`))
	if err != nil || string(got) != "{\"a\":9223372036854775807,\"z\":\"<\"}\n" {
		t.Fatalf("canonicalization = %q, %v", got, err)
	}
}
