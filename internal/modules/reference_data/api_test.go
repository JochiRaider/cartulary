package reference_data

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestPhase11_U_11_REFERENCE_PACK_01_RequestValidationNormalizationAndClosedRegistries(t *testing.T) {
	envelope := httpapi.UploadEnvelope{
		Metadata: map[string]json.RawMessage{
			"client_txn_id": json.RawMessage(`"txn-import"`),
		},
		FileSHA256Hex: "abc123",
	}
	request, apiErr := DecodeImportMetadata(envelope)
	if apiErr != nil {
		t.Fatalf("DecodeImportMetadata: %v", apiErr)
	}
	if request.ActivationPolicy != "staged_only" {
		t.Fatalf("default activation policy = %q", request.ActivationPolicy)
	}
	explicit := envelope
	explicit.Metadata = map[string]json.RawMessage{
		"client_txn_id":     json.RawMessage(`"txn-import"`),
		"activation_policy": json.RawMessage(`"staged_only"`),
	}
	explicitRequest, apiErr := DecodeImportMetadata(explicit)
	if apiErr != nil {
		t.Fatalf("explicit staged_only: %v", apiErr)
	}
	if !bytes.Equal(request.Normalized, explicitRequest.Normalized) {
		t.Fatalf("omitted and explicit staged_only must normalize equally: %s vs %s", request.Normalized, explicitRequest.Normalized)
	}
	auto := envelope
	auto.Metadata = map[string]json.RawMessage{
		"client_txn_id":     json.RawMessage(`"txn-import"`),
		"activation_policy": json.RawMessage(`"activate"`),
	}
	if _, apiErr := DecodeImportMetadata(auto); apiErr == nil || apiErr.Details["reason_code"] != "auto_activation_not_supported" {
		t.Fatalf("auto activation rejection = %#v", apiErr)
	}

	action, apiErr := DecodeActionRequest(bytes.NewBufferString(`{"client_txn_id":"txn-action","reason":" \r\n "}`))
	if apiErr != nil {
		t.Fatalf("DecodeActionRequest: %v", apiErr)
	}
	nullAction, apiErr := DecodeActionRequest(bytes.NewBufferString(`{"client_txn_id":"txn-action","reason":null}`))
	if apiErr != nil {
		t.Fatalf("DecodeActionRequest null: %v", apiErr)
	}
	omittedAction, apiErr := DecodeActionRequest(bytes.NewBufferString(`{"client_txn_id":"txn-action"}`))
	if apiErr != nil {
		t.Fatalf("DecodeActionRequest omitted: %v", apiErr)
	}
	if !bytes.Equal(action.Normalized, nullAction.Normalized) || !bytes.Equal(action.Normalized, omittedAction.Normalized) {
		t.Fatalf("empty, null, and omitted reason must normalize equally: %s %s %s", action.Normalized, nullAction.Normalized, omittedAction.Normalized)
	}

	refresh, apiErr := DecodeRefreshRequest(bytes.NewBufferString(`{"client_txn_id":"txn-refresh","pack_keys":["z","a","z"]}`))
	if apiErr != nil {
		t.Fatalf("DecodeRefreshRequest: %v", apiErr)
	}
	if got := refresh.PackKeys; len(got) != 2 || got[0] != "a" || got[1] != "z" {
		t.Fatalf("pack_keys must coalesce and sort, got %#v", got)
	}
	if _, apiErr := DecodeRefreshRequest(bytes.NewBufferString(`{"client_txn_id":"txn-refresh","pack_keys":[]}`)); apiErr == nil || apiErr.Details["reason_code"] != "empty_pack_keys" {
		t.Fatalf("empty pack_keys rejection = %#v", apiErr)
	}

	for _, reason := range []string{
		"checksum_mismatch",
		"signature_mismatch",
		"missing_integrity_metadata",
		"contract_incompatible",
		"path_traversal",
		"disallowed_content",
		"payload_missing",
		"archive_extracted_bytes_exceeded",
		"archive_compression_ratio_exceeded",
		"archive_member_count_exceeded",
	} {
		if !isValidVerificationFailureReason(reason) {
			t.Fatalf("reason %q missing from closed verification registry", reason)
		}
	}
	if err := referencePackVerificationFailed("checksum_mismatch"); err.Status != http.StatusConflict || err.Code != "reference_pack_verification_failed" {
		t.Fatalf("verification error shape = %#v", err)
	}
}

func TestPhase11_U_11_REFERENCE_PACK_03_VerifierRejectsArchiveLimits(t *testing.T) {
	valid := referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.host",
		PackKind:    "type_registry",
		PackVersion: "1",
	})
	cases := []struct {
		name       string
		bundle     []byte
		input      VerificationInput
		wantReason string
	}{
		{
			name:   "member-count",
			bundle: valid,
			input: VerificationInput{
				ContentType:   MediaTypeZip,
				ArchiveLimits: config.ArchiveLimits{MaxMembers: 1},
			},
			wantReason: "archive_member_count_exceeded",
		},
		{
			name:   "extracted-bytes",
			bundle: valid,
			input: VerificationInput{
				ContentType:     MediaTypeZip,
				ReferenceLimits: config.ReferencePackLimits{MaxExtractedBytes: 1},
			},
			wantReason: "archive_extracted_bytes_exceeded",
		},
		{
			name:   "compression-ratio",
			bundle: compressibleReferencePackBundle(t),
			input: VerificationInput{
				ContentType:   MediaTypeZip,
				ArchiveLimits: config.ArchiveLimits{MaxCompressionRatio: 1},
			},
			wantReason: "archive_compression_ratio_exceeded",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.input.Bundle = tc.bundle
			_, err := VerifyBundle(tc.input)
			var verificationErr *VerificationError
			if !asVerificationError(err, &verificationErr) || verificationErr.ReasonCode != tc.wantReason {
				t.Fatalf("VerifyBundle limit error = %v, want %s", err, tc.wantReason)
			}
		})
	}
}

func compressibleReferencePackBundle(t testing.TB) []byte {
	t.Helper()
	payload := bytes.Repeat([]byte("a"), 64*1024)
	payloadSHABytes := sha256.Sum256(payload)
	payloadSHA := hex.EncodeToString(payloadSHABytes[:])
	manifestBytes, err := json.Marshal(map[string]any{
		"pack_key":              "type_registry.compression",
		"pack_kind":             "type_registry",
		"pack_version":          "1",
		"pack_contract_version": PackContractVersionV1,
		"verification_method":   "manifest_sha256_v1",
		"payloads": []map[string]any{
			{"path": "payload/data.json", "sha256": payloadSHA},
		},
	})
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	addZipFile(t, writer, "manifest.json", manifestBytes)
	addZipFile(t, writer, "payload/data.json", payload)
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buffer.Bytes()
}

func referencePackBundle(t testing.TB, options bundleOptions) []byte {
	t.Helper()
	if options.ContractVersion == "" {
		options.ContractVersion = PackContractVersionV1
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
	payloads := []manifestPayload{{Path: options.PayloadPath, SHA256: payloadSHA}}
	canonicalPayloadSHA, err := canonicalPayloadSHA256(payloads)
	if err != nil {
		t.Fatalf("canonical payload sha: %v", err)
	}
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

type bundleOptions struct {
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

func asVerificationError(err error, target **VerificationError) bool {
	return errors.As(err, target)
}

func TestPhase11_U_11_REFERENCE_PACK_02_VerifierAcceptsLocalBundleAndRejectsFailures(t *testing.T) {
	valid := referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.host",
		PackKind:    "type_registry",
		PackVersion: "1",
	})
	result, err := VerifyBundle(VerificationInput{Bundle: valid, ContentType: MediaTypeZip})
	if err != nil {
		t.Fatalf("VerifyBundle valid: %v", err)
	}
	if result.PackKey != "type_registry.host" || result.PackVersion != "1" || result.PayloadSHA256 == "" || result.ManifestSHA256 == "" {
		t.Fatalf("unexpected verification result: %#v", result)
	}

	cases := []struct {
		name       string
		bundle     []byte
		wantReason string
	}{
		{name: "checksum", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "2", BadPayloadSHA: true}), wantReason: "checksum_mismatch"},
		{name: "signature", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "3", Signed: true, BadSignature: true}), wantReason: "signature_mismatch"},
		{name: "path", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "4", ExtraPath: "../escape.json"}), wantReason: "path_traversal"},
		{name: "active-content", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "5", PayloadPath: "payload/run.js"}), wantReason: "disallowed_content"},
		{name: "missing-payload", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "6", OmitPayload: true}), wantReason: "payload_missing"},
		{name: "contract", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "7", ContractVersion: "other"}), wantReason: "contract_incompatible"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := VerifyBundle(VerificationInput{Bundle: tc.bundle, ContentType: MediaTypeZip})
			var verificationErr *VerificationError
			if !asVerificationError(err, &verificationErr) || verificationErr.ReasonCode != tc.wantReason {
				t.Fatalf("VerifyBundle error = %v, want %s", err, tc.wantReason)
			}
		})
	}
}
