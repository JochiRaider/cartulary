package canonicaljson_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestStrictAdmissionRejectsLossyJSON_Unit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{"duplicate", `{"secret":1,"secret":2}`, canonicaljson.ErrDuplicateMember},
		{"escaped duplicate", `{"x":1,"\u0078":2}`, canonicaljson.ErrDuplicateMember},
		{"nested duplicate", `[{"x":null,"x":false}]`, canonicaljson.ErrDuplicateMember},
		{"duplicate before malformed value", `{"x":1,"x":2,"y":}`, canonicaljson.ErrInvalidJSON},
		{"duplicate before trailing value", `{"x":1,"x":2} []`, canonicaljson.ErrInvalidJSON},
		{"duplicate before mismatched delimiter", `[{"x":1,"x":2}]}`, canonicaljson.ErrInvalidJSON},
		{"duplicate before invalid Unicode", `{"x":1,"x":2,"y":"\ud800"}`, canonicaljson.ErrInvalidUnicode},
		{"invalid UTF8", "\"\xff\"", canonicaljson.ErrInvalidUnicode},
		{"high surrogate", `"\ud800"`, canonicaljson.ErrInvalidUnicode},
		{"low surrogate", `"\udfff"`, canonicaljson.ErrInvalidUnicode},
		{"wrong pair", `"\ud800\u0041"`, canonicaljson.ErrInvalidUnicode},
		{"key surrogate", `{"\ud800":1}`, canonicaljson.ErrInvalidUnicode},
		{"trailing value", `{} []`, canonicaljson.ErrInvalidJSON},
		{"empty", ``, canonicaljson.ErrInvalidJSON},
		{"BOM", "\xef\xbb\xbf{}", canonicaljson.ErrInvalidJSON},
		{"trailing comma", `{"x":1,}`, canonicaljson.ErrInvalidJSON},
		{"bad escape", `"\x"`, canonicaljson.ErrInvalidJSON},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := canonicaljson.DecodeStrict([]byte(test.input)); !errors.Is(err, test.want) {
				t.Fatalf("admission error = %v, want %v", err, test.want)
			}
			if _, err := canonicaljson.Canonicalize([]byte(test.input)); !errors.Is(err, test.want) {
				t.Fatalf("canonicalization error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestStrictAdmissionPreservesLexicalNumbersAndUnicode_Unit(t *testing.T) {
	t.Parallel()
	input := []byte(`{"number":1e2,"unicode":"\ud83d\ude00","literal":"\\ud800","empty":[],"distinct":{"é":1,"e\u0301":2}}`)
	value, err := canonicaljson.DecodeStrict(input)
	if err != nil {
		t.Fatal(err)
	}
	object := value.(map[string]any)
	if object["number"].(json.Number).String() != "1e2" || object["unicode"] != "😀" || object["literal"] != `\ud800` {
		t.Fatal("admission repaired an input value")
	}
	encoded, err := canonicaljson.Canonicalize(input)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"empty":[]`) || !strings.Contains(string(encoded), `"distinct":{"é":2,"é":1}`) {
		t.Fatalf("canonical bytes lost empty array or distinct Unicode keys: %s", encoded)
	}
	again, err := canonicaljson.Canonicalize(encoded)
	if err != nil || string(again) != string(encoded) {
		t.Fatal("canonicalization not idempotent")
	}
}
