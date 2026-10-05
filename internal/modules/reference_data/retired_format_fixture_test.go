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
// can promote their claims to trust. Signed does not mean a cryptographic signature.
func retiredReferencePackBundle(t testing.TB, options retiredBundleOptions) []byte {
	t.Helper()
	if options.ContractVersion == "" {
		options.ContractVersion = "cartulary.reference_pack.v1"
	}
	if options.PayloadPath == "" {
		options.PayloadPath = "payload/data.json"
	}
	payload := []byte(`{"items":[{"key":"host","label":"Host"}]}`)
	payloadSHABytes := sha256.Sum256(payload)
	payloadSHA := hex.EncodeToString(payloadSHABytes[:])
	if options.BadPayloadSHA {
		payloadSHA = "0000000000000000000000000000000000000000000000000000000000000000"
	}
	canonicalPayloadSHA := payloadSHA
	manifest := map[string]any{
		"pack_key":              options.PackKey,
		"pack_kind":             options.PackKind,
		"pack_version":          options.PackVersion,
		"pack_contract_version": options.ContractVersion,
		"source_identifier":     "https://offline.invalid/reference-packs/" + options.PackKey,
		"verification_method":   "manifest_sha256_v1",
		"payloads": []map[string]any{
			{"path": options.PayloadPath, "sha256": payloadSHA},
		},
	}
	if options.Signed {
		signatureSHA := canonicalPayloadSHA
		if options.BadSignature {
			signatureSHA = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
		}
		manifest["verification_method"] = "signed_manifest_v1"
		manifest["signer_key_id"] = "fixture-key"
		manifest["signature"] = map[string]any{"payload_sha256": signatureSHA}
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	addZipFile(t, writer, "manifest.json", manifestBytes)
	if !options.OmitPayload {
		addZipFile(t, writer, options.PayloadPath, payload)
	}
	if options.ExtraPath != "" {
		addZipFile(t, writer, options.ExtraPath, []byte("{}"))
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buffer.Bytes()
}

type retiredBundleOptions struct {
	PackKey         string
	PackKind        string
	PackVersion     string
	ContractVersion string
	PayloadPath     string
	BadPayloadSHA   bool
	Signed          bool
	BadSignature    bool
	ExtraPath       string
	OmitPayload     bool
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
