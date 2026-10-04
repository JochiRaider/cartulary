package packformat

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func testPortableVerificationContext(t *testing.T) {
	if !admittedPortableContextSize(4096) || admittedPortableContextSize(4097) || admittedPortableContextSize(0) {
		t.Fatal("context byte guard")
	}
	value := map[string]any{"schema_id": "cartulary.reference_pack_portable_verification_context.v1", "configuration_sha256": strings.Repeat("a", 64), "clock_trusted": false, "timeout_seconds": 1800}
	check := func(valid bool) {
		t.Helper()
		encoded, err := canonicaljson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		_, err = DecodePortableVerificationContext(encoded)
		if (err == nil) != valid {
			t.Fatalf("context admitted=%v, want=%v", err == nil, valid)
		}
	}
	check(true)
	for key, original := range value {
		delete(value, key)
		check(false)
		value[key] = nil
		check(false)
		value[key] = []any{}
		check(false)
		value[key] = original
	}
	value["unknown"] = true
	check(false)
	delete(value, "unknown")
	for _, valid := range []int{60, 1800, 86400} {
		value["timeout_seconds"] = valid
		check(true)
	}
	for _, invalid := range []any{59, 86401, 60.1, "1800"} {
		value["timeout_seconds"] = invalid
		check(false)
	}
	value["timeout_seconds"] = 1800
	encoded, err := canonicaljson.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodePortableVerificationContext(append(encoded, '\n')); err == nil {
		t.Fatal("noncanonical frozen context admitted")
	}
	if _, err := DecodePortableVerificationContext([]byte(`{"schema_id":"cartulary.reference_pack_portable_verification_context.v1","clock_trusted":false,"clock_trusted":true,"configuration_sha256":"` + strings.Repeat("a", 64) + `","timeout_seconds":1800}`)); err == nil {
		t.Fatal("duplicate frozen context admitted")
	}
}
