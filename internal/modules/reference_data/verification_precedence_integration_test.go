package reference_data_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestLiveVerificationIdentityPrecedesContentAndRuntimeChecks_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-precedence")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	importReferencePack(t, harness, admin, "type_registry.host", "2", "precedence-existing")
	for _, tc := range []struct {
		name, version, code, check string
		manifest                   func(map[string]any)
	}{
		{"rollback-before-schema", "1", "pack_release_sequence_rollback", "sequence_rollback", nil},
		{"sequence-collision-before-schema", "collision", "pack_release_sequence_collision", "sequence_collision", func(m map[string]any) { m["pack_release_sequence"] = 3 }},
		{"logical-collision-before-schema", "2", "pack_version_collision", "logical_collision", func(m map[string]any) { m["pack_release_sequence"] = 4 }},
		{"schema-before-runtime", "runtime", "content_schema_invalid", "content_schema", func(m map[string]any) { m["compatibility"].(map[string]any)["cartulary_contract_majors"] = []int{2} }},
		{"schema-before-dependency", "dependency", "content_schema_invalid", "content_schema", func(m map[string]any) {
			m["dependencies"] = []any{map[string]any{"pack_key": "type_registry.evidence", "pack_version": "absent", "payload_sha256": "0000000000000000000000000000000000000000000000000000000000000000"}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bundle := referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackVersion: tc.version, MetadataVersion: 2, ManifestTransform: tc.manifest, PayloadTransform: func([]byte) []byte { return []byte("{}\n") }})
			response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"precedence-`+tc.name+`"}`, bundle, "precedence.zip", "application/zip")
			jobID := requireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
			done := requireJob(t, harness, admin, jobID)
			if done["status"] != "failed" {
				t.Fatalf("combined faults did not reject: %#v", done)
			}
			if got := done["error_summary"].(map[string]any)["details"].(map[string]any)["reason_code"]; got != tc.code {
				t.Fatalf("first failing check reason = %v, want %s", got, tc.code)
			}
			var check string
			if err := harness.DB.QueryRow(`SELECT m.check_id FROM reference_pack_attempt_members m JOIN reference_pack_attempts a USING(attempt_id) JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1`, jobID).Scan(&check); err != nil {
				t.Fatal(err)
			}
			if check != tc.check {
				t.Fatalf("winning check = %s, want %s", check, tc.check)
			}
			details := done["error_summary"].(map[string]any)["details"].(map[string]any)
			ref, ok := details["validation_summary_ref"].(string)
			if !ok || !strings.HasPrefix(ref, "rpvs_") || details["primary_issue_id"] == nil {
				t.Fatalf("missing stable diagnostic reference: %#v", details)
			}
			response = httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/validation-summaries/"+ref, nil, httptestx.WithCookies(admin.SessionCookie))
			summary := requireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
			if summary["primary_issue_id"] != details["primary_issue_id"] || summary["total_issue_count"] != details["total_issue_count"] {
				t.Fatal("public diagnostic summary disagrees with terminal Job")
			}
			var stored []byte
			if err := harness.DB.QueryRow(`SELECT m.canonical_validation_summary FROM reference_pack_attempt_members m JOIN reference_pack_attempts a USING(attempt_id) JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1`, jobID).Scan(&stored); err != nil {
				t.Fatal(err)
			}
			var persisted map[string]any
			if err := json.Unmarshal(stored, &persisted); err != nil || !reflect.DeepEqual(persisted, summary) {
				t.Fatal("committed diagnostic changed during projection")
			}
			if tc.check == "content_schema" && summary["total_issue_count"] != float64(11) {
				t.Fatalf("missing complete empty-row schema findings: %#v", summary)
			}
			if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_events WHERE operation_id=(SELECT operation_id FROM reference_pack_operations WHERE job_id=$1) AND convert_from(canonical_attestation,'UTF8')::jsonb->>'validation_summary_ref'=$2`, jobID, ref); got != 1 {
				t.Fatal("failure attestation omitted exact summary binding")
			}
			if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_index_generations JOIN reference_pack_operations USING(operation_id) WHERE job_id=$1`, jobID); got != 0 {
				t.Fatal("content rejection created an index generation")
			}
		})
	}
	t.Run("complete-semantics-before-runtime", func(t *testing.T) {
		bundle := referencePackBundle(t, bundleOptions{
			PackKey: "type_registry.host", PackVersion: "semantics", MetadataVersion: 2,
			ManifestTransform: func(m map[string]any) {
				m["license"].(map[string]any)["expression"] = "UnknownLicense"
				m["compatibility"].(map[string]any)["cartulary_contract_majors"] = []int{2}
			},
			PayloadTransform: func(data []byte) []byte {
				var row map[string]any
				if err := json.Unmarshal(data, &row); err != nil {
					t.Fatal(err)
				}
				row["description"], row["display_label"] = " private secret ", " private secret "
				row["aliases"] = []string{"same", "same"}
				return append(integrationCanonical(t, row), '\n')
			},
		})
		response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"complete-semantic-findings"}`, bundle, "semantics.zip", "application/zip")
		jobID := requireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
		done := requireJob(t, harness, admin, jobID)
		if done["status"] != "failed" {
			t.Fatal(done)
		}
		details := done["error_summary"].(map[string]any)["details"].(map[string]any)
		ref := details["validation_summary_ref"].(string)
		response = httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/validation-summaries/"+ref, nil, httptestx.WithCookies(admin.SessionCookie))
		summary := requireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if summary["total_issue_count"] != float64(4) {
			t.Fatalf("semantic findings: %#v", summary)
		}
		for _, value := range summary["issues"].([]any) {
			if value.(map[string]any)["check_id"] != "content_semantics" {
				t.Fatal("later runtime fault changed winning check", value)
			}
		}
		encoded, err := json.Marshal(summary)
		if err != nil || strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "UnknownLicense") {
			t.Fatal("unsafe semantic diagnostics", err)
		}
		if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_index_generations JOIN reference_pack_operations USING(operation_id) WHERE job_id=$1`, jobID) != 0 {
			t.Fatal("semantic rejection constructed index")
		}
	})
}

func TestLiveVerificationResolvesExactDependenciesBeforeIndexing_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-dependencies")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	importReferencePackWithKind(t, harness, admin, "enrichment.tor", "enrichment", "1", "dependency-seed")
	requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/enrichment.tor/1/disable", "dependency-disable", ""), http.StatusOK)
	var digest string
	if err := harness.DB.QueryRow(`SELECT payload_sha256 FROM reference_pack_versions WHERE pack_key='enrichment.tor' AND pack_version='1'`).Scan(&digest); err != nil {
		t.Fatal(err)
	}
	dependency := func(key, version, digest string) map[string]any {
		return map[string]any{"pack_key": key, "pack_version": version, "payload_sha256": digest}
	}
	for index, test := range []struct {
		name, check string
		count       int
		mutate      func(map[string]any)
	}{
		{"missing", "dependencies", 2, func(m map[string]any) {
			m["dependencies"] = []any{dependency("framework.attack", "absent", strings.Repeat("0", 64)), dependency("framework.d3fend", "absent", strings.Repeat("0", 64))}
		}},
		{"wrong-digest", "dependencies", 1, func(m map[string]any) {
			m["dependencies"] = []any{dependency("enrichment.tor", "1", strings.Repeat("0", 64))}
		}},
		{"self-conflict", "conflicts", 1, func(m map[string]any) {
			m["conflicts"] = []any{map[string]any{"pack_key": "enrichment.lolbas", "pack_version": nil}}
		}},
		{"dependency-conflict", "conflicts", 1, func(m map[string]any) {
			m["dependencies"] = []any{dependency("enrichment.tor", "1", digest)}
			m["conflicts"] = []any{map[string]any{"pack_key": "enrichment.tor", "pack_version": "1"}}
		}},
		{"disabled-healthy-dependency", "", 0, func(m map[string]any) { m["dependencies"] = []any{dependency("enrichment.tor", "1", digest)} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			bundle := referencePackBundle(t, bundleOptions{PackKey: "enrichment.lolbas", PackVersion: strconv.Itoa(index + 1), ManifestTransform: test.mutate})
			response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"dependency-`+test.name+`"}`, bundle, "dependency.zip", "application/zip")
			job := requireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
			done := requireJob(t, harness, admin, job)
			if test.check == "" {
				if done["status"] != "succeeded" {
					t.Fatal(done)
				}
				return
			}
			if done["status"] != "failed" {
				t.Fatal(done)
			}
			var check string
			var total int
			if err := harness.DB.QueryRow(`SELECT m.check_id,(convert_from(m.canonical_validation_summary,'UTF8')::jsonb->>'total_issue_count')::int FROM reference_pack_attempt_members m JOIN reference_pack_attempts a USING(attempt_id) JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1`, job).Scan(&check, &total); err != nil {
				t.Fatal(err)
			}
			if check != test.check || total != test.count {
				t.Fatalf("got %s/%d want %s/%d", check, total, test.check, test.count)
			}
			if count := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_index_generations JOIN reference_pack_operations USING(operation_id) WHERE job_id=$1`, job); count != 0 {
				t.Fatal("dependency rejection produced an index")
			}
		})
	}
}
