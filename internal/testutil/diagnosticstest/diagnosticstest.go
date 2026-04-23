package diagnosticstest

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/golden"
)

func RequireJSONMatchesGolden(t testing.TB, gotJSON string, goldenParts []string) {
	t.Helper()

	got := []byte(gotJSON)
	want := golden.MustRead(goldenParts...)

	var gotValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("unmarshal got diagnostics: %v", err)
	}

	var wantValue any
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("unmarshal golden diagnostics: %v", err)
	}

	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("diagnostics mismatch\nwant:\n%s\n\ngot:\n%s", string(want), string(got))
	}
}
