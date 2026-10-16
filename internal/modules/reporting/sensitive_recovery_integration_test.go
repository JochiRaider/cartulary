package reporting_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

// Called from the existing release and preview owner-routed integration paths.
func assertPrivateRevealReadiness(t *testing.T, harness *appsupport.ServerHarness, id string, preview bool) {
	t.Helper()
	ctx := context.Background()
	if err := reporting.ValidateRetainedSensitiveArtifacts(ctx, harness.Pool); err != nil {
		t.Fatal("valid private state rejected", err)
	}
	selectSQL := `SELECT canonical_bytes,reveal_map_sha256 FROM reporting_release_reveal_maps WHERE release_id=$1`
	updateSQL := `UPDATE reporting_release_reveal_maps SET canonical_bytes=$2,reveal_map_sha256=$3 WHERE release_id=$1`
	if preview {
		selectSQL = `SELECT canonical_bytes,reveal_map_sha256 FROM reporting_preview_reveal_maps WHERE preview_attempt_id=$1`
		updateSQL = `UPDATE reporting_preview_reveal_maps SET canonical_bytes=$2,reveal_map_sha256=$3 WHERE preview_attempt_id=$1`
	}
	var original []byte
	var digest string
	if err := harness.DB.QueryRow(selectSQL, id).Scan(&original, &digest); err != nil {
		t.Fatal("private map not retained", err)
	}
	restore := func() {
		if _, err := harness.DB.Exec(updateSQL, id, original, digest); err != nil {
			t.Fatal(err)
		}
	}
	defer restore()
	var unchanged map[string]any
	decoder := json.NewDecoder(bytes.NewReader(original))
	decoder.UseNumber()
	if err := decoder.Decode(&unchanged); err != nil {
		t.Fatal(err)
	}
	fixtureEncoding, err := json.Marshal(unchanged)
	if err != nil || !bytes.Equal(fixtureEncoding, original) {
		t.Fatal("tamper fixture must preserve canonical owner bytes", err)
	}

	for _, mutation := range []string{"foreign identity", "missing entries", "changed token", "changed rule", "changed action", "changed disclosure", "unknown member"} {
		var model map[string]any
		decoder := json.NewDecoder(bytes.NewReader(original))
		decoder.UseNumber()
		if err := decoder.Decode(&model); err != nil {
			t.Fatal(err)
		}
		switch mutation {
		case "foreign identity":
			field := "release_id"
			if preview {
				field = "preview_attempt_id"
			}
			model[field] = "69000000-0000-4000-8000-000000000001"
		case "missing entries":
			model["entries"] = []any{}
		case "changed token":
			model["entries"].([]any)[0].(map[string]any)["display_token"] = "OTHER-0001"
		case "changed rule":
			model["entries"].([]any)[0].(map[string]any)["rule_id"] = "unselected-rule"
		case "changed action":
			model["entries"].([]any)[0].(map[string]any)["action"] = "stub"
		case "changed disclosure":
			model["entries"].([]any)[0].(map[string]any)["disclosure_partition_refs"] = []any{"party:unselected"}
		case "unknown member":
			model["hostile"] = true
		}
		// Simple fixture values make encoding/json bytes identical to the owner
		// canonical form here. Digest consistency alone must never admit a map.
		changed, err := json.Marshal(model)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(append([]byte(reporting.RedactionRevealMapSchemaID+"\n"), changed...))
		if _, err := harness.DB.Exec(updateSQL, id, changed, hex.EncodeToString(sum[:])); err != nil {
			t.Fatal(err)
		}
		if err := reporting.ValidateRetainedSensitiveArtifacts(ctx, harness.Pool); err == nil {
			t.Fatal("readiness accepted altered private artifact", mutation)
		}
		restore()
	}
	// A recomputed private display digest must fail both the public subject
	// binding and the independent frozen-input check for release and preview.
	entry := unchanged["entries"].([]any)[0].(map[string]any)
	entry["canonical_display_value"] = "invented recovery value"
	value, err := json.Marshal(map[string]any{"schema_id": "cartulary.tokenizable_subject.v1", "stable_subject_ref": entry["stable_subject_ref"], "value": entry["canonical_display_value"]})
	if err != nil {
		t.Fatal(err)
	}
	valueSum := sha256.Sum256(value)
	entry["canonical_display_value_sha256"] = hex.EncodeToString(valueSum[:])
	changed, err := json.Marshal(unchanged)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(append([]byte(reporting.RedactionRevealMapSchemaID+"\n"), changed...))
	if _, err := harness.DB.Exec(updateSQL, id, changed, hex.EncodeToString(sum[:])); err != nil {
		t.Fatal(err)
	}
	if err := reporting.ValidateRetainedSensitiveArtifacts(ctx, harness.Pool); err == nil {
		t.Fatal("recomputed private display digest escaped the independent public digest")
	}
	references := reportingReadinessReferences(t, harness)
	if preview {
		err = reporting.ValidateRetainedPreviews(ctx, harness.Pool, references)
	} else {
		err = reporting.ValidateRetainedReleases(ctx, harness.Pool, references)
	}
	if err == nil {
		t.Fatal("recovery accepted invented private values with recomputed digests")
	}
	restore()
	if err := reporting.ValidateRetainedSensitiveArtifacts(ctx, harness.Pool); err != nil {
		t.Fatal("restored original private artifact rejected", err)
	}
}
