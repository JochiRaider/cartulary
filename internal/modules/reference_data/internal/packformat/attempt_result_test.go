package packformat

import (
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"testing"
)

func testAttemptResultBoundary(t *testing.T) {
	if !admittedAttemptResultSize(4096) || admittedAttemptResultSize(4097) || admittedAttemptResultSize(0) {
		t.Fatal("attempt result byte guard")
	}
	t.Helper()
	for _, outcome := range []string{"succeeded", "content_rejected", "canceled", "timed_out", "stale_state", "execution_failed", "interrupted"} {
		members, failed := int64(3), int64(0)
		if outcome == "content_rejected" {
			failed = 2
		}
		if outcome == "interrupted" {
			members = 0
		}
		data, err := EncodeAttemptResult(outcome, members, failed)
		if err != nil {
			t.Fatal(err)
		}
		value, err := canonicaljson.DecodeStrict(data)
		if err != nil {
			t.Fatal(err)
		}
		object := value.(map[string]any)
		check := func() {
			t.Helper()
			raw, err := canonicaljson.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			if ValidateAttemptResult(raw) == nil {
				t.Fatal("invalid result admitted", outcome)
			}
		}
		for key, original := range object {
			delete(object, key)
			check()
			object[key] = nil
			check()
			object[key] = []any{}
			check()
			object[key] = original
		}
		object["unexpected"] = false
		check()
		delete(object, "unexpected")
		if ValidateAttemptResult(append(data, '\n')) == nil {
			t.Fatal("noncanonical result admitted")
		}
	}
	for _, v := range []struct {
		outcome         string
		members, failed int64
	}{
		{"unknown", 1, 0}, {"succeeded", 1, 1}, {"execution_failed", 1, 1}, {"content_rejected", 1, 0}, {"content_rejected", 1, 2}, {"succeeded", -1, 0}, {"interrupted", 1, 0}, {"succeeded", 9007199254740992, 0},
	} {
		if _, err := EncodeAttemptResult(v.outcome, v.members, v.failed); err == nil {
			t.Fatal("invalid count/outcome relation admitted", v)
		}
	}
	for _, members := range []int64{0, 1, 9007199254740991} {
		if _, err := EncodeAttemptResult("succeeded", members, 0); err != nil {
			t.Fatal("valid count boundary", err)
		}
	}
	if ValidateAttemptResult([]byte(`{"outcome":"interrupted","outcome":"interrupted","schema_id":"cartulary.reference_pack_attempt_result.v1"}`)) == nil {
		t.Fatal("duplicate result key admitted")
	}
}
