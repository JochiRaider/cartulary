package packformat

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func testLimitDiagnostics(t *testing.T) {
	ctx := context.Background()
	assertLimit := func(summary *ValidationSummary, id, expected, actual string) {
		t.Helper()
		for _, issue := range summary.Issues {
			d := issue.Details
			if d.LimitID != nil && *d.LimitID == id && d.ExpectedToken != nil && *d.ExpectedToken == expected && d.ActualToken != nil && *d.ActualToken == actual {
				return
			}
		}
		t.Fatalf("missing numeric limit finding %s %s %s: %#v", id, expected, actual, summary)
	}
	for _, tc := range []struct {
		id, check        string
		size             int64
		expected, actual string
	}{
		{"max_manifest_json_bytes", "manifest_schema", 1048577, "1048576", "1048577"},
		{"max_bundle_json_bytes", "bundle_shape", 16385, "16384", "16385"},
	} {
		program, _ := ManifestChecks(nil, tc.size, true)
		if tc.check == "bundle_shape" {
			program, _ = HintChecks(nil, tc.size)
		}
		summary, err := CheckSummary(ctx, tc.check, nil, program[tc.check])
		if err != nil {
			t.Fatal(err)
		}
		assertLimit(summary, tc.id, tc.expected, tc.actual)
	}
	value, err := canonicaljson.DecodeStrict([]byte(manifestFixture(t).Canonical))
	if err != nil {
		t.Fatal(err)
	}
	manifest := value.(map[string]any)
	for _, tc := range []struct {
		field, id string
		count     int
		maximum   string
	}{
		{"files", "max_manifest_files", 67, "66"},
		{"source_artifacts", "max_source_artifacts", 65, "64"},
		{"dependencies", "max_dependencies", 65, "64"},
		{"conflicts", "max_conflicts", 65, "64"},
	} {
		original := manifest[tc.field]
		manifest[tc.field] = make([]any, tc.count)
		raw, err := canonicaljson.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		program, _ := ManifestChecks(raw, int64(len(raw)), true)
		summary, err := checkProgramRange(ctx, "manifest_encoding", "manifest_canonical", program)
		if err != nil {
			t.Fatal(err)
		}
		actual := "65"
		if tc.count == 67 {
			actual = "67"
		}
		assertLimit(summary, tc.id, tc.maximum, actual)
		manifest[tc.field] = original
	}
	for _, tc := range []struct {
		kind, id, maximum, actual string
		count                     int
	}{
		{"root", "max_root_update_signatures", "128", "129", 129},
		{"timestamp", "max_signatures", "64", "65", 65},
	} {
		data := []byte(trustVectors(t)[0].Bootstrap)
		if tc.kind == "timestamp" {
			data = vectorMetadata(trustVectors(t)[0])["metadata/timestamp.json"]
		}
		value, err := canonicaljson.DecodeStrict(data)
		if err != nil {
			t.Fatal(err)
		}
		object := value.(map[string]any)
		first := object["signatures"].([]any)[0]
		signatures := make([]any, tc.count)
		for i := range signatures {
			signatures[i] = first
		}
		object["signatures"] = signatures
		data, err = canonicaljson.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		summary, err := CheckSummary(ctx, "tuf_schema", nil, func(_ context.Context, emit FindingSink) error {
			maxSignatures := 64
			if tc.kind == "root" {
				maxSignatures = 128
			}
			_, err := metadataSchemaFindings(data, tc.kind, maxSignatures, "$.metadata[0]", emit)
			var failure *Failure
			if errors.As(err, &failure) {
				return nil
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		assertLimit(summary, tc.id, tc.maximum, tc.actual)
	}
	archive := bytes.NewReader(make([]byte, 2048)) // A complete empty ustar stream.
	limits := DefaultArchiveLimits()
	limits.ContainerBytes = 1024
	_, err = ExtractWithDiagnostics(ctx, archive, archive.Size(), limits, memoryDestination{}, nil)
	var failure *Failure
	if !errors.As(err, &failure) || failure.Summary == nil {
		t.Fatal(err)
	}
	assertLimit(failure.Summary, "max_container_bytes", "1024", "2048")
	summary, err := CheckSummary(ctx, "archive_structure", nil, func(_ context.Context, emit FindingSink) error {
		return structuralPathFindings(strings.Repeat("a", 1025), false, "$.archive_members[0]", emit)
	})
	if err != nil {
		t.Fatal(err)
	}
	assertLimit(summary, "max_path_bytes", "1024", "1025")
	assertLimit(summary, "max_segment_bytes", "255", "1025")
	summary, err = CheckSummary(ctx, "content_schema", nil, func(ctx context.Context, emit FindingSink) error {
		_, err := enumerateContentRows(ctx, strings.NewReader(strings.Repeat("a", 1048577)), nil, "$.entries", "entry", 1, emit)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	assertLimit(summary, "max_line_bytes", "1048576", "1048577")
	if !diagnosticMeasurement("18446744073709551615") || diagnosticMeasurement("18446744073709551616") || diagnosticMeasurement("01") || diagnosticDecimal("18446744073709551615") {
		t.Fatal("unsigned count strings widened path indexes or lost their bound")
	}
}

func checkProgramRange(ctx context.Context, first, last string, program map[string]CheckFunc) (*ValidationSummary, error) {
	active := false
	for _, check := range Checks() {
		if check.ID == first {
			active = true
		}
		if active {
			summary, err := CheckSummary(ctx, check.ID, nil, program[check.ID])
			if err != nil || summary.Result == "failed" {
				return summary, err
			}
		}
		if check.ID == last {
			break
		}
	}
	return successSummary(), nil
}
