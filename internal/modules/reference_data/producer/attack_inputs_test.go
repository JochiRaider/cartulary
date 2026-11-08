package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testATTACKCurrentAdmission(t *testing.T) {
	profile := fixtureBytes(t, "source-profile.native.v2.r2.json")
	policy := fixtureBytes(t, "field-policy.native.v1.json")
	lock := fixtureBytes(t, "source-lock.v20.1.json")
	for _, name := range []string{"source-lock.v20.0.json", "source-lock.v20.1.json"} {
		if _, err := admitATTACKInputs(fixtureBytes(t, name), profile, policy); err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range []string{"source_profile_id", "status", "accepted_source_lock_sha256", "field_policy_sha256", "adopted_schema_sha256"} {
		t.Run(key, func(t *testing.T) {
			var p map[string]any
			if err := json.Unmarshal(profile, &p); err != nil {
				t.Fatal(err)
			}
			p[key] = "unbound"
			if _, err := admitATTACKInputs(lock, encoded(t, p), policy); err == nil {
				t.Fatal("altered profile admitted")
			}
		})
	}
	for _, tc := range []struct {
		name                  string
		lock, profile, policy []byte
	}{
		{"profile_byte_change", lock, append(bytes.Clone(profile), '\n'), policy},
		{"policy_byte_change", lock, profile, append(bytes.Clone(policy), '\n')},
		{"unbound_lock", append(bytes.Clone(lock), '\n'), profile, policy},
		{"duplicate_json", []byte(`{"schema_id":"a","schema_id":"b"}`), profile, policy},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := admitATTACKInputs(tc.lock, tc.profile, tc.policy); err == nil {
				t.Fatal("unbound input admitted")
			}
		})
	}
	var p Policy
	if err := DecodeClosed(policy, &p); err != nil {
		t.Fatal(err)
	}
	p.Status = "proposal_only"
	if _, err := ReadPolicy(encoded(t, p)); err == nil {
		t.Fatal("proposal policy admitted")
	}
	// Neither public entry point admits the unpinned synthetic source used by
	// the private semantic/projection tests, even with a valid acquisition.
	root := testRoot(t)
	synthetic, fetch := sourceFixture(t, "valid")
	if _, err := Acquire(context.Background(), root, synthetic, fetch); err != nil {
		t.Fatal(err)
	}
	builder := Builder{ID: "unit-fixture", Version: "1", SHA256: strings.Repeat("1", 64), Size: 1, GoVersion: "unit-fixture"}
	report, err := Audit(context.Background(), root, synthetic, profile, policy, "2026-11-06T00:00:00Z", builder)
	if err == nil || report.SchemaID != "" {
		t.Fatal("public audit admitted synthetic source")
	}
	dir := t.TempDir()
	for name, b := range map[string][]byte{"lock.json": synthetic, "profile.json": profile, "policy.json": policy} {
		if err := os.WriteFile(filepath.Join(dir, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	input := PackageInput{SchemaID: "cartulary.attack_package_input.v1", SourceRoot: root.Name(), SourceLock: filepath.Join(dir, "lock.json"), SourceProfile: filepath.Join(dir, "profile.json"), FieldPolicy: filepath.Join(dir, "policy.json"), OutputRoot: filepath.Join(dir, "output"), BuiltAt: "2026-11-06T00:00:00Z", PackVersion: "unit", LedgerRoot: filepath.Join(dir, "ledger"), BuilderSHA256: builder.SHA256, Repository: FixtureRepository, Redistribution: "restricted"}
	if _, err := Assemble(context.Background(), encoded(t, input), builder); err == nil {
		t.Fatal("public assembly admitted synthetic source")
	}
	for _, path := range []string{input.OutputRoot, input.LedgerRoot, filepath.Join(root.Name(), "audits")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("rejected input mutated publication or audit state", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if report, err := auditATTACKSources(ctx, root, syntheticATTACKInputs(t, synthetic), input.BuiltAt, builder); err != context.Canceled || report.Status == "source_audit_passed" {
		t.Fatal("cancellation lost", err)
	}
}
