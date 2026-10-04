package reference_data

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

func TestSuccessfulEnvelopeClosedPersistence_Unit(t *testing.T) {
	storage, attempt := canonicalEngineFixture(t)
	content, err := verifyEngineFixture(storage, attempt)
	if err != nil {
		t.Fatal(err)
	}
	defer content.Close()
	ref, digest, until := "objects/fixture", content.ContainerSHA256, content.Trust.ValidUntil
	envelope := successfulEnvelope{
		SchemaID: "cartulary.reference_pack_successful_envelope.v1", OperationID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		PackKey: content.Manifest.Key, PackVersion: content.Manifest.Version, DistributionKind: "operator_imported",
		ManifestSHA256: content.ManifestSHA256, PayloadSHA256: content.PayloadSHA256, ContainerSHA256: &digest, ContainerRef: &ref,
		VerifiedAt: content.VerifiedAt, TrustValidUntil: &until, TrustSnapshot: &content.TrustSnapshot, TrustProposal: &content.Trust,
	}
	data, err := canonicaljson.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeSuccessfulEnvelope(data)
	if err != nil || !reflect.DeepEqual(decoded, envelope) {
		t.Fatal("valid retained envelope rejected", err)
	}
	clone := func() map[string]any {
		v, err := canonicaljson.DecodeStrict(data)
		if err != nil {
			t.Fatal(err)
		}
		return v.(map[string]any)
	}
	reject := func(t *testing.T, object map[string]any) {
		t.Helper()
		encoded, err := canonicaljson.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decodeSuccessfulEnvelope(encoded); err == nil || strings.Contains(err.Error(), "hostile-secret") {
			t.Fatal("invalid or unsafe retained envelope", err)
		}
	}
	paths := [][]string{
		{}, {"trust_snapshot"}, {"trust_proposal"}, {"trust_proposal", "binding"},
		{"trust_proposal", "signers"}, {"trust_proposal", "metadata"},
		{"trust_proposal", "metadata", "targets"}, {"trust_proposal", "metadata", "snapshot"}, {"trust_proposal", "metadata", "timestamp"},
	}
	objectAt := func(object map[string]any, path []string) map[string]any {
		for _, key := range path {
			object = object[key].(map[string]any)
		}
		return object
	}
	for _, path := range paths {
		for field := range objectAt(clone(), path) {
			for _, mutation := range []string{"omitted", "null", "type"} {
				t.Run(strings.Join(append(append([]string{}, path...), field, mutation), "/"), func(t *testing.T) {
					object := clone()
					target := objectAt(object, path)
					switch mutation {
					case "omitted":
						delete(target, field)
					case "null":
						target[field] = nil
					case "type":
						target[field] = false
					}
					reject(t, object)
				})
			}
		}
		object := clone()
		objectAt(object, path)["hostile-secret"] = "hostile-secret"
		reject(t, object)
	}
	for name, mutate := range map[string]func(map[string]any){
		"expiry equality": func(m map[string]any) {
			m["trust_valid_until"] = m["verified_at"]
			objectAt(m, []string{"trust_proposal"})["valid_until"] = m["verified_at"]
		},
		"binding mismatch": func(m map[string]any) {
			objectAt(m, []string{"trust_proposal", "binding"})["payload_sha256"] = strings.Repeat("0", 64)
		},
		"root alias": func(m map[string]any) {
			objectAt(m, []string{"trust_snapshot", "root_history"})["01"] = m["trust_snapshot"].(map[string]any)["root"]
		},
		"missing admitted root": func(m map[string]any) { objectAt(m, []string{"trust_snapshot"})["root_history"] = map[string]any{} },
		"unknown role": func(m map[string]any) {
			objectAt(m, []string{"trust_snapshot", "highest"})["hostile-secret"] = map[string]any{}
		},
		"unknown signer role": func(m map[string]any) { objectAt(m, []string{"trust_proposal", "signers"})["hostile-secret"] = []any{} },
		"unknown transition": func(m map[string]any) {
			objectAt(m, []string{"trust_proposal"})["root_transitions"] = []any{map[string]any{"version": json.Number("2")}}
		},
		"unsafe reference": func(m map[string]any) { m["container_ref"] = "../hostile-secret" },
		"retired Go field name": func(m map[string]any) {
			snapshot := objectAt(m, []string{"trust_snapshot"})
			snapshot["Root"] = snapshot["root"]
			delete(snapshot, "root")
		},
	} {
		t.Run(name, func(t *testing.T) { object := clone(); mutate(object); reject(t, object) })
	}
	if _, err := decodeSuccessfulEnvelope(append(data, '\n')); err == nil {
		t.Fatal("noncanonical envelope repaired")
	}
	duplicate := append([]byte(`{"schema_id":"hostile-secret",`), data[1:]...)
	if _, err := decodeSuccessfulEnvelope(duplicate); err == nil {
		t.Fatal("duplicate envelope member erased")
	}
	builtin := clone()
	builtin["distribution_kind"] = "packaged_builtin"
	for _, key := range []string{"container_sha256", "container_ref", "trust_valid_until", "trust_snapshot", "trust_proposal"} {
		builtin[key] = nil
	}
	encoded, err := canonicaljson.Marshal(builtin)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeSuccessfulEnvelope(encoded); err != nil {
		t.Fatal("explicit nullable built-in fields rejected", err)
	}
}
