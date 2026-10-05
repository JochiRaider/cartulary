package reference_data

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestRetainedProvenanceClosedAndBoundToExactSet_Unit(t *testing.T) {
	_, builtins, err := loadBuiltinRelease(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	storage, attempt := canonicalEngineFixture(t)
	operator, err := verifyEngineFixture(storage, attempt)
	if err != nil {
		t.Fatal(err)
	}
	defer operator.Close()
	contents := map[string]*verifiedContent{}
	members := []PackSetMember{}
	for _, builtin := range builtins {
		content := builtin.Content
		if content.Manifest.Key == operator.Manifest.Key {
			content = operator
		}
		contents[content.Manifest.Key] = content
		members = append(members, memberFor(content))
	}
	set, err := buildPackSet(members)
	if err != nil {
		t.Fatal(err)
	}
	anchors := []PackProvenance{}
	for _, member := range set.Members {
		content := contents[member.Key]
		envelope := successfulEnvelope{ManifestSHA256: content.ManifestSHA256, PayloadSHA256: content.PayloadSHA256, VerifiedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
		if content == operator {
			envelope.VerifiedAt = content.VerifiedAt
			envelope.TrustProposal = &content.Trust
			envelope.TrustValidUntil = &content.Trust.ValidUntil
		}
		anchors = append(anchors, provenanceFor(set.ID, content.Manifest, envelope))
	}
	data, err := canonicaljson.Marshal(anchors)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeRetainedProvenance(data, set)
	if err != nil || !reflect.DeepEqual(decoded, anchors) {
		t.Fatal("valid retained anchors rejected", err)
	}
	clone := func() []any {
		value, err := canonicaljson.DecodeStrict(data)
		if err != nil {
			t.Fatal(err)
		}
		return value.([]any)
	}
	reject := func(t *testing.T, value any) {
		t.Helper()
		encoded, err := canonicaljson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decodeRetainedProvenance(encoded, set); err == nil || err.Error() != "reference pack consumer: pack_unavailable" {
			t.Fatal("invalid anchor admitted or unsafe diagnostic", err)
		}
	}
	// Exercise every populated object boundary, including the manifest-owned
	// source-artifact and license shapes, without deriving tests from prose.
	var walk func([]any, func([]any) map[string]any, string)
	walk = func(value []any, selectObject func([]any) map[string]any, path string) {
		object := selectObject(value)
		for field, original := range object {
			for _, mutation := range []string{"omitted", "null", "type"} {
				if mutation == "null" && original == nil {
					continue
				}
				t.Run(path+"/"+field+"/"+mutation, func(t *testing.T) {
					copy := clone()
					target := selectObject(copy)
					switch mutation {
					case "omitted":
						delete(target, field)
					case "null":
						target[field] = nil
					case "type":
						target[field] = false
					}
					if mutation == "null" && field == "source_as_of" {
						encoded, err := canonicaljson.Marshal(copy)
						if err != nil {
							t.Fatal(err)
						}
						if _, err := decodeRetainedProvenance(encoded, set); err != nil {
							t.Fatal("required nullable source time rejected", err)
						}
					} else {
						reject(t, copy)
					}
				})
			}
			if _, ok := original.(map[string]any); ok {
				walk(value, func(v []any) map[string]any { return selectObject(v)[field].(map[string]any) }, path+"/"+field)
			}
			if array, ok := original.([]any); ok {
				for index, child := range array {
					if _, ok := child.(map[string]any); ok {
						walk(value, func(v []any) map[string]any { return selectObject(v)[field].([]any)[index].(map[string]any) }, path+"/"+field)
					}
				}
			}
		}
		copy := clone()
		selectObject(copy)["hostile-secret"] = "hostile-secret"
		reject(t, copy)
	}
	for i := range anchors {
		walk(clone(), func(v []any) map[string]any { return v[i].(map[string]any) }, anchors[i].PackKey)
	}
	for _, field := range []string{"pack_set_id", "pack_key", "pack_version", "manifest_sha256", "payload_sha256", "authority_class"} {
		copy := clone()
		copy[0].(map[string]any)[field] = "hostile-secret"
		reject(t, copy)
	}
	copy := clone()
	copy[0], copy[1] = copy[1], copy[0]
	reject(t, copy)
	reject(t, clone()[:len(anchors)-1])
	for i, anchor := range anchors {
		copy := clone()
		object := copy[i].(map[string]any)
		if anchor.TrustValidUntil != nil {
			object["trust_valid_until"] = object["last_verified_at"]
			reject(t, copy)
			object["trust_valid_until"] = anchor.TrustValidUntil.Format(time.RFC3339Nano)
			object["verified_signer_key_ids"] = append(object["verified_signer_key_ids"].([]any), anchor.VerifiedSignerKeyIDs[0])
			reject(t, copy)
		} else {
			object["verified_signer_key_ids"] = []string{strings.Repeat("a", 64)}
			reject(t, copy)
		}
	}
	for _, malformed := range [][]byte{append(append([]byte{}, data...), '\n'), []byte(strings.Replace(string(data), `"pack_key":`, `"pack_key":"hostile-secret","pack_key":`, 1))} {
		if _, err := decodeRetainedProvenance(malformed, set); err == nil {
			t.Fatal("noncanonical or duplicate members admitted")
		}
	}
	var fields map[string]json.RawMessage
	encoded, _ := canonicaljson.Marshal(anchors[0])
	if err := json.Unmarshal(encoded, &fields); err != nil || len(fields) != 20 {
		t.Fatal("provenance projection changed without review", err)
	}
}
