package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	attackinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/attack"
)

const contractRoot = "../../../../contracts/reference-pack-producers/attack/"

func fixtureBytes(t *testing.T, name string) []byte {
	t.Helper()
	b, err := attackinputs.Files.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

type fixtureFetcher struct {
	values map[string][]byte
	calls  int
}

func (f *fixtureFetcher) Fetch(_ context.Context, url string, _ int64) ([]byte, error) {
	f.calls++
	v, ok := f.values[url]
	if !ok {
		return nil, errors.New("unexpected network locator")
	}
	return v, nil
}

func sourceFixture(t *testing.T, scenario string) ([]byte, *fixtureFetcher) {
	t.Helper()
	lock, err := ReadLock(fixtureBytes(t, "source-lock.v20.1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var input struct {
		Schema  string                    `json:"schema_id"`
		Purpose string                    `json:"purpose"`
		Bundles map[string]map[string]any `json:"bundles"`
	}
	if err := DecodeClosed(fixtureBytes(t, "fixtures/inputs.v1.json"), &input); err != nil {
		t.Fatal(err)
	}
	enterprise := input.Bundles["enterprise-attack.json"]["objects"].([]any)
	malware := enterprise[1].(map[string]any)
	tool := enterprise[2].(map[string]any)
	relation := enterprise[3].(map[string]any)
	switch scenario {
	case "valid", "hash_mismatch", "malformed_json", "late_malformed_json", "release_mismatch", "tag_moved":
	case "native_created_valid":
		tool["created"] = "2026-11-05T00:00:00Z"
	case "native_created_invalid":
		tool["created"] = "not-a-timestamp"
	case "native_contributors_64", "native_contributors_65":
		count := 64
		if scenario == "native_contributors_65" {
			count++
		}
		contributors := make([]string, count)
		for i := range contributors {
			contributors[i] = fmt.Sprintf("Synthetic contributor %02d", i)
		}
		tool["x_mitre_contributors"] = contributors
	case "unknown_relationship_type":
		relation["relationship_type"] = "inferred-transitive-path"
	case "unknown_property":
		malware["unclassified"] = "not silently discarded"
	case "wrong_property_type":
		malware["revoked"] = "false"
	case "duplicate_conflict":
		input.Bundles["mobile-attack.json"]["objects"].([]any)[1].(map[string]any)["name"] = "different subject"
	case "missing_endpoint":
		relation["target_ref"] = "malware--99999999-9999-4999-8999-999999999999"
	case "replacement_cycle":
		relation["relationship_type"] = "revoked-by"
		input.Bundles["enterprise-attack.json"]["objects"] = append(enterprise, map[string]any{"id": "relationship--55555555-5555-4555-8555-555555555555", "type": "relationship", "relationship_type": "revoked-by", "source_ref": tool["id"], "target_ref": malware["id"]})
	}
	f := &fixtureFetcher{values: map[string][]byte{}}
	assets := []any{}
	for i, a := range lock.Artifacts {
		b := []byte("Synthetic producer unit license; not MITRE redistribution permission.\n")
		if a.Domain != "license" {
			b = encoded(t, input.Bundles[a.Name])
		}
		if (scenario == "malformed_json" && a.Domain == "enterprise-attack") || (scenario == "late_malformed_json" && a.Domain == "mobile-attack") {
			b = []byte("{")
		}
		lock.Artifacts[i].SHA256 = Digest(b)
		lock.Artifacts[i].Size = int64(len(b))
		f.values[a.URL] = b
		if a.Domain != "license" {
			assets = append(assets, map[string]any{"name": a.Name, "size": len(b), "digest": "sha256:" + Digest(b), "browser_download_url": a.URL})
		}
	}
	release := map[string]any{"tag_name": lock.Tag, "published_at": lock.PublishedAt, "draft": false, "prerelease": false, "assets": assets}
	if scenario == "release_mismatch" {
		release["published_at"] = "2026-11-04T00:00:00Z"
	}
	f.values[lock.ReleaseURL] = encoded(t, release)
	f.values[lock.TagURL] = encoded(t, map[string]any{"sha": lock.TagObject, "tag": lock.Tag, "object": map[string]any{"sha": lock.Commit, "type": "commit"}})
	refSHA := lock.TagObject
	if scenario == "tag_moved" {
		refSHA = strings.Repeat("1", 40)
	}
	f.values["https://api.github.com/repos/mitre-attack/attack-stix-data/git/ref/tags/"+lock.Tag] = encoded(t, map[string]any{"ref": "refs/tags/" + lock.Tag, "object": map[string]any{"sha": refSHA, "type": "tag"}})
	if scenario == "hash_mismatch" {
		f.values[lock.Artifacts[1].URL] = []byte("{")
	}
	return encoded(t, lock), f
}

func TestATTACKSourceFixtures(t *testing.T) {
	var cases struct {
		Schema string `json:"schema_id"`
		Cases  []struct {
			Name      string  `json:"name"`
			Operation string  `json:"operation"`
			Expected  *string `json:"expected_error"`
		} `json:"cases"`
	}
	if err := DecodeClosed(fixtureBytes(t, "fixtures/cases.v1.json"), &cases); err != nil {
		t.Fatal(err)
	}
	if cases.Schema != "cartulary.attack_source_cases.v1" || len(cases.Cases) == 0 {
		t.Fatal("source fixtures missing")
	}
	for _, tc := range cases.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			root := testRoot(t)
			lock, fetch := sourceFixture(t, tc.Name)
			_, err := Acquire(context.Background(), root, lock, fetch)
			if tc.Operation == "acquire" {
				if err == nil || tc.Expected == nil || !strings.Contains(err.Error(), *tc.Expected) {
					t.Fatalf("acquisition outcome %v", err)
				}
				if _, err := root.Stat("acquisition.json"); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("failed source admitted")
				}
				return
			}
			if tc.Operation != "audit" {
				t.Fatal("fixture has no expectation runner")
			}
			if err != nil {
				t.Fatal(err)
			}
			report, err := auditATTACKSources(context.Background(), root, syntheticATTACKInputs(t, lock), "2026-11-06T00:00:00Z", Builder{ID: "unit-fixture", Version: "1", SHA256: strings.Repeat("1", 64), Size: 1, GoVersion: "unit-fixture"})
			if tc.Expected != nil {
				if report.SchemaID != "" && report.Status != "source_audit_failed" {
					t.Fatal("failed audit claimed success")
				}
				if err == nil || !strings.Contains(strings.Join(report.Issues, "\n")+"\n"+err.Error(), *tc.Expected) {
					t.Fatalf("expected %s: %v %v", *tc.Expected, err, report.Issues)
				}
				again, againErr := auditATTACKSources(context.Background(), root, syntheticATTACKInputs(t, lock), "2026-11-06T00:00:00Z", Builder{ID: "unit-fixture", Version: "1", SHA256: strings.Repeat("1", 64), Size: 1, GoVersion: "unit-fixture"})
				if againErr == nil || !bytes.Equal(encoded(t, report), encoded(t, again)) {
					t.Fatal("failed audit is not byte reproducible")
				}
				return
			}
			if err != nil {
				t.Fatal(err, report.Issues)
			}
			assertAuditReportContract(t, report, syntheticATTACKInputs(t, lock))
			// Independently authored witness for the small, explicitly synthetic
			// source parser fixture. The real release is audited by Make separately.
			if report.LookupAliases["malware--11111111-1111-4111-8111-111111111111"][0] != "unique" {
				t.Fatal("ambiguous alias survived")
			}
			if report.UniqueRecords != 6 || report.DuplicateIDs != 1 || len(report.Collisions) != 1 || report.Collisions[0].Key != "shared" || len(report.Collisions[0].ObjectIDs) != 2 {
				t.Fatal("source reconciliation witness changed")
			}
			before := fetch.calls
			if _, err := Acquire(context.Background(), root, lock, fetch); err != nil || fetch.calls != before {
				t.Fatal("retained replay fetched or failed", err)
			}
			if _, err := Check(root, lock); err != nil {
				t.Fatal(err)
			}
			again, err := auditATTACKSources(context.Background(), root, syntheticATTACKInputs(t, lock), report.BuiltAt, report.Builder)
			if err != nil || !bytes.Equal(encoded(t, report), encoded(t, again)) {
				t.Fatal("audit is not byte reproducible", err)
			}
		})
	}
}

func TestATTACKSourceRetentionBoundaries(t *testing.T) {
	t.Run("retained_acquisition_inventory_shape", func(t *testing.T) {
		root := testRoot(t)
		lock, fetch := sourceFixture(t, "valid")
		if _, err := Acquire(context.Background(), root, lock, fetch); err != nil {
			t.Fatal(err)
		}
		if _, err := Check(root, lock); err != nil {
			t.Fatal(err)
		}
		schema, err := compileProducerSchema(fixtureBytes(t, "acquisition.v1.schema.json"), "")
		if err != nil {
			t.Fatal(err)
		}
		b, err := root.ReadFile("acquisition.json")
		if err != nil {
			t.Fatal(err)
		}
		var acquired map[string]any
		if err := json.Unmarshal(b, &acquired); err != nil {
			t.Fatal(err)
		}
		if err := schema.Validate(acquired); err != nil {
			t.Fatal(err)
		}
		t.Run("acquisition_requires_tag_reference", func(t *testing.T) {
			b, err := root.ReadFile("acquisition.json")
			if err != nil {
				t.Fatal(err)
			}
			var receipt map[string]any
			if err := json.Unmarshal(b, &receipt); err != nil {
				t.Fatal(err)
			}
			delete(receipt, "tag_reference")
			if err := schema.Validate(receipt); err == nil {
				t.Fatal("acquisition schema accepted absent tag reference")
			}
			if err := root.WriteFile("acquisition.json", encoded(t, receipt), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Check(root, lock); err == nil {
				t.Fatal("missing tag reference admitted")
			}
		})
	})
	root := testRoot(t)
	lock, fetch := sourceFixture(t, "valid")
	a, err := Acquire(context.Background(), root, lock, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if err := Retain(root, a.Sources[0].Path, []byte("changed"), 0600); err == nil {
		t.Fatal("immutable source overwritten")
	}
	if err := Retain(root, "../escape", []byte("x"), 0600); err == nil {
		t.Fatal("path escaped")
	}
	if err := root.WriteFile(a.Sources[0].Path, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Check(root, lock); err == nil {
		t.Fatal("retained source tampering admitted")
	}
	if err := root.Symlink("sources", "alias"); err != nil {
		t.Fatal(err)
	}
	if err := Retain(root, "alias/new", []byte("x"), 0600); err == nil {
		t.Fatal("symlinked evidence admitted")
	}
}

func TestATTACKSourceContractClosure(t *testing.T) {
	t.Run("schema compiler", testProducerSchemaCompiler)
	t.Run("current admission", testATTACKCurrentAdmission)
	t.Run("partial audit report", testATTACKPartialAudit)
	t.Run("native audit enforcement", testATTACKNativeAuditEnforcement)

	lock := fixtureBytes(t, "source-lock.v20.1.json")
	var value map[string]any
	if err := json.Unmarshal(lock, &value); err != nil {
		t.Fatal(err)
	}
	for key := range value {
		copy := map[string]any{}
		for k, v := range value {
			if k != key {
				copy[k] = v
			}
		}
		if _, err := ReadLock(encoded(t, copy)); err == nil {
			t.Fatalf("omitted lock member %s accepted", key)
		}
	}
	value["unknown"] = true
	if _, err := ReadLock(encoded(t, value)); err == nil {
		t.Fatal("unknown lock member admitted")
	}
	if _, err := ReadLock([]byte(`{"schema_id":"a","schema_id":"b"}`)); err == nil {
		t.Fatal("duplicate JSON member admitted")
	}
	assertATTACKProducerExcludedFromRuntime(t)

	// A source-profile input is a regular exact file, never executable
	// runtime transformation or a symlink into human documentation.
	for _, name := range []string{"source-lock.v20.1.json", "field-policy.native.v1.json"} {
		st, err := os.Lstat(filepath.Join(contractRoot, name))
		if err != nil || !st.Mode().IsRegular() {
			t.Fatal("producer input not regular", err)
		}
	}
}

// Synthetic input admission is test-only. The exported producer entry points
// must reject these locks; source and projection tests use this private seam.
func syntheticATTACKInputs(t *testing.T, lockBytes []byte) attackInputs {
	t.Helper()
	inputs, err := readSyntheticATTACKInputs(lockBytes)
	if err != nil {
		t.Fatal(err)
	}
	return inputs
}

func readSyntheticATTACKInputs(lockBytes []byte) (attackInputs, error) {
	lock, err := ReadLock(lockBytes)
	if err != nil {
		return attackInputs{}, err
	}
	policyBytes, err := attackinputs.Files.ReadFile("field-policy.native.v1.json")
	if err != nil {
		return attackInputs{}, err
	}
	policy, err := ReadPolicy(policyBytes)
	if err != nil {
		return attackInputs{}, err
	}
	return attackInputs{lock: lock, policy: policy, profileID: "synthetic-unit-source", lockBytes: lockBytes, policyBytes: policyBytes, profileBytes: []byte(`{"purpose":"synthetic private seam only"}`)}, nil
}
