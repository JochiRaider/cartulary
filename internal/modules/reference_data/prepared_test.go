package reference_data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

func preparedFixture(t *testing.T) []byte {
	t.Helper()
	storage, attempt := canonicalEngineFixture(t)
	content, err := verifyEngineFixture(storage, attempt)
	if err != nil {
		t.Fatal(err)
	}
	defer content.Close()
	ref, _ := ParseStorageRef("objects/container")
	expiry := content.Trust.ValidUntil
	digest, rawRef := content.ContainerSHA256, ref.String()
	p := preparedVersion{
		Content: content, IndexID: uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Envelope: successfulEnvelope{
			SchemaID:    "cartulary.reference_pack_successful_envelope.v1",
			OperationID: uuid.MustParse("22222222-2222-4222-8222-222222222222"),
			PackKey:     content.Manifest.Key, PackVersion: content.Manifest.Version, DistributionKind: "operator_imported",
			ManifestSHA256: content.ManifestSHA256, PayloadSHA256: content.PayloadSHA256,
			ContainerSHA256: &digest, ContainerRef: &rawRef, VerifiedAt: content.VerifiedAt,
			TrustValidUntil: &expiry, TrustSnapshot: &content.TrustSnapshot, TrustProposal: &content.Trust,
		},
	}
	paths := []string{"manifest.json"}
	for _, file := range content.Manifest.Files {
		paths = append(paths, file.Path)
	}
	for i, path := range paths {
		member := content.Inventory[path]
		ref, _ := ParseStorageRef(fmt.Sprintf("objects/member-%d", i))
		p.Objects = append(p.Objects, preparedObject{ID: uuid.New(), Path: path, Digest: member.SHA256, Size: member.Size, Reference: ref})
	}
	p.Objects = append(p.Objects, preparedObject{ID: uuid.New(), Path: "container", Digest: digest, Size: content.ContainerBytes, Reference: ref})
	data, err := encodePrepared(p)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestPreparedResultClosedShapeAndBindings_Unit(t *testing.T) {
	data := preparedFixture(t)
	if _, err := decodePrepared(data); err != nil {
		t.Fatal(err)
	}
	mutate := func(t *testing.T, change func(map[string]any)) {
		t.Helper()
		value, err := canonicaljson.DecodeStrict(data)
		if err != nil {
			t.Fatal(err)
		}
		change(value.(map[string]any))
		encoded, err := canonicaljson.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decodePrepared(encoded); err == nil {
			t.Fatal("invalid preparation admitted")
		}
	}
	for _, boundary := range []string{"prepared", "object"} {
		selectObject := func(value map[string]any) map[string]any {
			if boundary == "object" {
				return value["objects"].([]any)[1].(map[string]any)
			}
			return value
		}
		members := []string{"schema_id", "manifest", "envelope", "objects", "index_id"}
		if boundary == "object" {
			members = []string{"id", "path", "sha256", "size", "reference"}
		}
		for _, member := range members {
			for _, kind := range []string{"omitted", "null", "wrong type"} {
				t.Run(boundary+"/"+member+"/"+kind, func(t *testing.T) {
					mutate(t, func(value map[string]any) {
						object := selectObject(value)
						switch kind {
						case "omitted":
							delete(object, member)
						case "null":
							object[member] = nil
						case "wrong type":
							object[member] = true
						}
					})
				})
			}
		}
		t.Run(boundary+"/unknown", func(t *testing.T) {
			mutate(t, func(value map[string]any) { selectObject(value)["unknown"] = "untrusted" })
		})
	}
	for _, tc := range []struct {
		name   string
		change func(map[string]any)
	}{
		{"empty inventory", func(v map[string]any) { v["objects"] = []any{} }},
		{"missing member", func(v map[string]any) { v["objects"] = v["objects"].([]any)[:1] }},
		{"duplicate path", func(v map[string]any) { rows := v["objects"].([]any); rows[1] = rows[0] }},
		{"out of order", func(v map[string]any) { rows := v["objects"].([]any); rows[0], rows[1] = rows[1], rows[0] }},
		{"wrong digest", func(v map[string]any) {
			v["objects"].([]any)[1].(map[string]any)["sha256"] = string(bytes.Repeat([]byte("0"), 64))
		}},
		{"wrong length", func(v map[string]any) { v["objects"].([]any)[1].(map[string]any)["size"] = 0 }},
		{"fractional length", func(v map[string]any) { v["objects"].([]any)[1].(map[string]any)["size"] = json.Number("0.5") }},
		{"container reference substitution", func(v map[string]any) { v["objects"].([]any)[0].(map[string]any)["reference"] = "objects/other" }},
		{"object ID collision", func(v map[string]any) {
			rows := v["objects"].([]any)
			rows[1].(map[string]any)["id"] = rows[0].(map[string]any)["id"]
		}},
		{"reference collision", func(v map[string]any) {
			rows := v["objects"].([]any)
			rows[1].(map[string]any)["reference"] = rows[0].(map[string]any)["reference"]
		}},
		{"path escape", func(v map[string]any) { v["objects"].([]any)[1].(map[string]any)["reference"] = "../outside" }},
		{"empty index identity", func(v map[string]any) { v["index_id"] = uuid.Nil.String() }},
		{"omitted success envelope member", func(v map[string]any) { delete(v["envelope"].(map[string]any), "trust_snapshot") }},
	} {
		t.Run(tc.name, func(t *testing.T) { mutate(t, tc.change) })
	}
	for _, invalid := range [][]byte{append(bytes.Clone(data), '\n'), append([]byte(`{"index_id":null,`), data[1:]...)} {
		if _, err := decodePrepared(invalid); err == nil {
			t.Fatal("noncanonical or duplicate member admitted")
		}
	}
}
