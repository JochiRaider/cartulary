package reporting

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"
)

func testReportingCanonicalBytes(t *testing.T) {
	t.Helper()
	// The UTF-8 byte order here differs from RFC 8785's UTF-16 order.
	raw := []byte("{\"𐀀\":1,\"\":2,\"a\":\"\\u003c\\u003e\\u0026\\u2028\\u000b\\n\"}")
	want := []byte("{\"a\":\"<>&\u2028\\u000b\\n\",\"\":2,\"𐀀\":1}")
	got, err := canonicalReportingBytes(raw)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("independent canonical bytes: %q %v", got, err)
	}
	again, err := canonicalReportingBytes(got)
	if err != nil || !bytes.Equal(got, again) {
		t.Fatal("canonicalization is not idempotent", err)
	}
	for _, invalid := range []string{`{"n":1.0}`, `{"n":1e0}`, `{"n":-0}`, `{"n":9007199254740992}`, `{"n":-9007199254740992}`, `{"n":1,"n":2}`, `{"a":"\ud800"}`, "{\"a\":\"\xff\"}"} {
		if _, err := canonicalReportingBytes([]byte(invalid)); err == nil {
			t.Fatalf("admitted invalid boundary %q", invalid)
		}
	}
	for _, valid := range []string{`{"n":0}`, `{"n":9007199254740991}`, `{"n":-9007199254740991}`} {
		if _, err := canonicalReportingBytes([]byte(valid)); err != nil {
			t.Fatal("rejected inclusive number bound", err)
		}
	}
	for _, invalid := range []any{[]byte("opaque"), map[string]any{"x": "\xff"}, map[string]any{"\xff": true}, math.Copysign(0, -1), math.Inf(1), math.NaN(), 1.5, json.RawMessage(`{"n":1e0}`), make(chan int)} {
		if _, err := canonicalJSON(invalid); err == nil {
			t.Fatalf("admitted invalid producer value %T", invalid)
		}
	}
	data := []byte("{\"a\":\"<>&\u2028\",\"n\":9007199254740991,\"schema_id\":\"cartulary.reporting.test.v1\"}")
	// Expected domain-separated digest calculated independently with Python
	// hashlib over the literal UTF-8 vector, never from this implementation.
	if digest := reportingObjectDigest("cartulary.reporting.test.v1", data); digest != "6294ddcedd6488f31276613cddb2b9aa74d396767922b7f1cba1ab7e0f672c16" {
		t.Fatal("object digest", digest)
	}
	if hashHex(data) == reportingObjectDigest("cartulary.reporting.test.v1", data) {
		t.Fatal("file and object identity conflated")
	}
}
