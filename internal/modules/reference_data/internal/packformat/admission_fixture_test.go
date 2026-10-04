package packformat

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// These static containers were produced with Python's standard archive tools.
// Their digests and diagnostic identities were independently computed with
// hashlib over declared tuples. Expected bytes never come from the verifier.
func runAdmissionManifest(t *testing.T, input []byte, expected map[string]any) {
	t.Helper()
	value, err := canonicaljson.DecodeStrict(input)
	if err != nil || !compileProjection("admission_fixture.v1.schema.json").matches(value) {
		t.Fatal("invalid admission fixture", err)
	}
	var fixture struct {
		Kind     string            `json:"kind"`
		Bytes    string            `json:"bytes_base64"`
		Members  map[string]string `json:"expected_members"`
		Operator bool              `json:"operator"`
	}
	if err := json.Unmarshal(input, &fixture); err != nil {
		t.Fatal(err)
	}
	data, err := base64.StdEncoding.Strict().DecodeString(fixture.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	var verdict error
	switch fixture.Kind {
	case "archive":
		if Digest(data) != expected["expected_container_sha256"] {
			t.Fatal("independent container digest differs")
		}
		destination := memoryDestination{}
		inventory, err := ExtractWithDiagnostics(context.Background(), bytes.NewReader(data), int64(len(data)), DefaultArchiveLimits(), destination, newDiagnosticTestScratch())
		verdict = err
		if len(destination) != len(fixture.Members) || err == nil && len(inventory) != len(fixture.Members) {
			t.Fatal("unexpected extracted bytes or directory in logical inventory", len(destination), len(inventory))
		}
		for path, want := range fixture.Members {
			if got, ok := destination[path]; !ok || got.String() != want || inventory[path].SHA256 != Digest([]byte(want)) {
				t.Fatal("logical member bytes changed")
			}
		}
		if expected["expected_manifest_sha256"] != nil {
			manifest, err := DecodeManifest([]byte(fixture.Members["manifest.json"]), fixture.Operator)
			if err != nil || Digest([]byte(fixture.Members["manifest.json"])) != expected["expected_manifest_sha256"] {
				t.Fatal("manifest identity changed", err)
			}
			payload, err := PayloadDigest(manifest.Files)
			if err != nil || payload != expected["expected_payload_sha256"] {
				t.Fatal("payload identity changed", err)
			}
			if err := ValidateInventory(manifest, inventory, fixture.Operator); err != nil {
				t.Fatal("complete canonical inventory rejected", err)
			}
			var rows contentRows
			if err := ValidateContent(context.Background(), manifest, contentMemory(fixture.Members), &rows); err != nil || len(rows) == 0 {
				t.Fatal("complete canonical payload rejected", err)
			}
		}
	case "manifest", "license":
		var admitted Manifest
		admitted, verdict = DecodeManifestWithDiagnostics(context.Background(), data, fixture.Operator, newDiagnosticTestScratch())
		if verdict == nil && fixture.Kind == "license" {
			verdict = runCheckRange(context.Background(), "content_semantics", "content_semantics", newDiagnosticTestScratch(), map[string]CheckFunc{
				"content_semantics": func(_ context.Context, emit FindingSink) error { return ManifestLicenseFindings(admitted, emit) },
			}, true)
		}
		if verdict == nil && expected["expected_manifest_sha256"] != nil {
			payload, err := PayloadDigest(admitted.Files)
			if err != nil || Digest(data) != expected["expected_manifest_sha256"] || payload != expected["expected_payload_sha256"] {
				t.Fatal("independent manifest/payload identity changed", err)
			}
		}
	default:
		t.Fatal("unknown fixture operation")
	}
	assertManifestVerdict(t, expected, verdict)
}

func assertManifestVerdict(t *testing.T, expected map[string]any, verdict error) {
	t.Helper()
	if expected["expected_public_error"] == nil {
		if verdict != nil || len(expected["expected_issues"].([]any)) != 0 {
			t.Fatal("valid fixture rejected", verdict)
		}
		return
	}
	public := expected["expected_public_error"].(map[string]any)
	var failure *Failure
	if !errors.As(verdict, &failure) || failure.Summary == nil || public["code"] != "reference_pack_verification_failed" || failure.Code != public["reason_code"] {
		t.Fatal("wrong rejection", verdict)
	}
	got, err := canonicaljson.Marshal(failure.Summary.Issues)
	if err != nil {
		t.Fatal(err)
	}
	want, err := canonicaljson.Marshal(expected["expected_issues"])
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("diagnostic tuple changed: %v\ngot %s\nwant %s", err, got, want)
	}
}
