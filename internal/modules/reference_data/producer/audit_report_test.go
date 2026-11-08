package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

const nativeToolIssue = "native_source_invalid:tool--22222222-2222-4222-8222-222222222222:enterprise-attack.json:/objects/2"

func testATTACKNativeAuditEnforcement(t *testing.T) {
	for _, tc := range []struct {
		name string
		fail bool
	}{
		{"native_created_valid", false},
		{"native_created_invalid", true},
		{"native_contributors_64", false},
		{"native_contributors_65", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := testRoot(t)
			lock, fetch := sourceFixture(t, tc.name)
			if _, err := Acquire(t.Context(), root, lock, fetch); err != nil {
				t.Fatal(err)
			}
			inputs := syntheticATTACKInputs(t, lock)
			builder := Builder{ID: "unit-fixture", Version: "1", SHA256: strings.Repeat("1", 64), Size: 1, GoVersion: "unit-fixture"}
			var first []byte
			for range 3 {
				report, err := auditATTACKSources(t.Context(), root, inputs, "2026-11-06T00:00:00Z", builder)
				wantStatus := "source_audit_passed"
				var wantIssues []string
				if tc.fail {
					wantStatus = "source_audit_failed"
					wantIssues = []string{nativeToolIssue}
				}
				if (err != nil) != tc.fail || report.Status != wantStatus || !slices.Equal(report.Issues, wantIssues) {
					t.Fatalf("audit status=%s issues=%v error=%v; want %s %v", report.Status, report.Issues, err, wantStatus, wantIssues)
				}
				assertAuditReportContract(t, report, inputs)
				got := encoded(t, report)
				if first != nil && !bytes.Equal(first, got) {
					t.Fatal("native audit report is not byte reproducible")
				}
				first = got
			}
		})
	}
}

func assertAuditReportContract(t *testing.T, report AuditReport, inputs attackInputs) {
	t.Helper()
	var fields map[string]any
	if err := json.Unmarshal(encoded(t, report), &fields); err != nil {
		t.Fatal(err)
	}
	if fields["schema_id"] != "cartulary.attack_source_audit.v2" || fields["source_profile_id"] != inputs.profileID || fields["source_profile_sha256"] != Digest(inputs.profileBytes) {
		t.Fatal("audit provenance contract changed")
	}
	for _, key := range []string{"pack_status", "native_candidate", "proposed_lookup_aliases"} {
		if _, exists := fields[key]; exists {
			t.Fatal("obsolete audit field", key)
		}
	}
	if _, ok := fields["lookup_aliases"].(map[string]any); !ok {
		t.Fatal("lookup aliases missing")
	}
}

func testATTACKPartialAudit(t *testing.T) {
	root := testRoot(t)
	lock, fetch := sourceFixture(t, "late_malformed_json")
	if _, err := Acquire(context.Background(), root, lock, fetch); err != nil {
		t.Fatal(err)
	}
	inputs := syntheticATTACKInputs(t, lock)
	builder := Builder{ID: "unit-fixture", Version: "1", SHA256: strings.Repeat("1", 64), Size: 1, GoVersion: "unit-fixture"}
	first, err := auditATTACKSources(context.Background(), root, inputs, "2026-11-06T00:00:00Z", builder)
	if err == nil || !strings.Contains(err.Error(), "malformed source JSON") || first.Status != "source_audit_failed" || len(first.References) == 0 || first.UniqueRecords != 5 {
		t.Fatal("expected attributable partial failure", err)
	}
	assertAuditReportContract(t, first, inputs)
	for range 3 {
		again, err := auditATTACKSources(context.Background(), root, inputs, first.BuiltAt, builder)
		if err == nil || !bytes.Equal(encoded(t, first), encoded(t, again)) {
			t.Fatal("partial report order changed", err)
		}
	}
}
