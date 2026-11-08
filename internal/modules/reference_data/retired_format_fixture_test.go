package reference_data

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

// Retired checksum-labelled formats exist only to prove that no legacy reader
// can promote their claims to trust. The signed variant is not a cryptographic signature.
func retiredReferencePackBundle(t testing.TB, signed bool) []byte {
	t.Helper()
	const payloadPath = "payload/data.json"
	payload := []byte(`{"items":[{"key":"host","label":"Host"}]}`)
	payloadSHABytes := sha256.Sum256(payload)
	payloadSHA := hex.EncodeToString(payloadSHABytes[:])
	manifest := map[string]any{
		"pack_key":              "type_registry.host",
		"pack_kind":             "type_registry",
		"pack_version":          "1",
		"pack_contract_version": "cartulary.reference_pack.v1",
		"source_identifier":     "https://offline.invalid/reference-packs/type_registry.host",
		"verification_method":   "manifest_sha256_v1",
		"payloads": []map[string]any{
			{"path": payloadPath, "sha256": payloadSHA},
		},
	}
	if signed {
		manifest["verification_method"] = "signed_manifest_v1"
		manifest["signer_key_id"] = "fixture-key"
		manifest["signature"] = map[string]any{"payload_sha256": payloadSHA}
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	addZipFile(t, writer, "manifest.json", manifestBytes)
	addZipFile(t, writer, payloadPath, payload)
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buffer.Bytes()
}

func addZipFile(t testing.TB, writer *zip.Writer, name string, data []byte) {
	t.Helper()
	file, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create zip member %s: %v", name, err)
	}
	if _, err := file.Write(data); err != nil {
		t.Fatalf("write zip member %s: %v", name, err)
	}
}
