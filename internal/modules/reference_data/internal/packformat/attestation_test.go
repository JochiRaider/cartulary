package packformat

import (
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestAttestationIndependentIdentityAndClosedShape_Unit(t *testing.T) {
	// The ASCII fixture's hash was authored using sorted JSON and sha256sum,
	// independently of the Go JCS and attestation implementation.
	raw := fixtureProjection("fixtures/attestation.v1.json")
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil || ValidateAttestation(canonical) != nil {
		t.Fatal("independent attestation rejected", err)
	}
	decode := func() map[string]any {
		var object map[string]any
		if err := json.Unmarshal(raw, &object); err != nil {
			t.Fatal(err)
		}
		return object
	}
	rehash := func(object map[string]any) []byte {
		delete(object, "attestation_id")
		preimage, err := canonicaljson.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		object["attestation_id"] = "rpa_" + Digest(preimage)
		encoded, err := canonicaljson.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	for name := range decode() {
		if name == "attestation_id" {
			continue
		}
		t.Run("omitted "+name, func(t *testing.T) {
			object := decode()
			delete(object, name)
			if ValidateAttestation(rehash(object)) == nil {
				t.Fatal("omission accepted")
			}
		})
		t.Run("wrong type "+name, func(t *testing.T) {
			object := decode()
			object[name] = true
			if ValidateAttestation(rehash(object)) == nil {
				t.Fatal("wrong type accepted")
			}
		})
	}
	for _, name := range []string{"root", "targets", "snapshot", "timestamp"} {
		object := decode()
		delete(object["trusted_metadata_versions"].(map[string]any), name)
		if ValidateAttestation(rehash(object)) == nil {
			t.Fatal("omitted nullable metadata version", name)
		}
	}
	for _, mutate := range []func(map[string]any){
		func(o map[string]any) { o["unexpected_payload_value"] = "hostile" },
		func(o map[string]any) { o["trusted_metadata_versions"].(map[string]any)["unexpected"] = 1 },
		func(o map[string]any) { o["event_kind"] = "verification" },
		func(o map[string]any) { o["actor_kind"] = "user" },
		func(o map[string]any) { o["distribution_kind"] = "packaged_builtin" },
		func(o map[string]any) { o["container_sha256"] = o["manifest_sha256"] },
		func(o map[string]any) { o["occurred_at"] = "2026-02-30T00:00:00Z" },
		func(o map[string]any) {
			ids := o["verified_signer_key_ids"].([]any)
			o["verified_signer_key_ids"] = []any{ids[0], ids[0]}
		},
		func(o map[string]any) {
			ids := o["verified_signer_key_ids"].([]any)
			o["verified_signer_key_ids"] = []any{ids[1], ids[0]}
		},
	} {
		object := decode()
		mutate(object)
		if ValidateAttestation(rehash(object)) == nil {
			t.Fatal("invalid attestation admitted")
		}
	}
	object := decode()
	object["attestation_id"] = "rpa_" + Digest([]byte("wrong identity"))
	altered, err := canonicaljson.Marshal(object)
	if err != nil || ValidateAttestation(altered) == nil {
		t.Fatal("incorrect attestation identity admitted", err)
	}
}
