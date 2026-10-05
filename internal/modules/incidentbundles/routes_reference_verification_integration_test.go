package incidentbundles_test

import (
	"bytes"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func testEmbeddedReferenceDestinationVerification(t *testing.T, runtime *appsupport.Runtime, source *appsupport.ServerHarness) {
	t.Helper()
	expected := referencetest.OwnerFixture(t, "portable_destination")
	vector, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/portable-input.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	references, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/portable-references.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	fixture := referencetest.PortableFixture(t, time.Now().UTC(), vector, references)
	bootstrap := filepath.Join(t.TempDir(), "portable-bootstrap.json")
	if err := os.WriteFile(bootstrap, fixture.Bootstrap, 0600); err != nil {
		t.Fatal(err)
	}
	target := startIsolatedIncidentBundleServerWithEnv(t, runtime, "embedded-reference-target", map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"})
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, source.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, target.Server.HTTP.URL)
	var active string
	if err := target.DB.QueryRow(`SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`).Scan(&active); err != nil || active != "rpset_"+*expected.Set {
		t.Fatal(err)
	}
	create := func(key string) (string, []byte) {
		t.Helper()
		incident := scenariotest.CreateIncident(t, source.Server, sourceAdmin, map[string]any{"client_txn_id": key + "-create", "incident_key": key, "title": "Embedded destination verification"})
		id := incident["incident_id"].(string)
		bundle := exportBundleBytes(t, source, sourceAdmin, id, key+"-export")
		bundle = replaceStructuredBundleMember(t, bundle, "data/reference_pack_refs.json", fixture.References)
		bundle = replaceStructuredBundleMember(t, bundle, "ext/reference_packs/content.json", fixture.Content)
		var manifest incidentBundleManifestMirror
		if err := json.Unmarshal(zipMemberBytes(t, bundle, "manifest.json"), &manifest); err != nil {
			t.Fatal(err)
		}
		manifest.ReferencePackMode = "embedded"
		raw, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		bundle = replaceStructuredBundleMember(t, bundle, "manifest.json", raw)
		return id, bundle
	}
	t.Run("untrusted clock rejects fresh verification", func(t *testing.T) {
		untrusted := startIsolatedIncidentBundleServerWithEnv(t, runtime, "embedded-clock-untrusted", map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "false"})
		admin, _ := flowtest.ProvisionBootstrapAdmin(t, untrusted.Server.HTTP.URL)
		id, bundle := create("EMBEDDED-CLOCK-UNTRUSTED")
		bundle = replaceStructuredBundleMember(t, bundle, fixture.Path, fixture.Container)
		response := postImport(t, untrusted.Server, admin, `{"client_txn_id":"embedded-clock-import"}`, bundle, "bundle.zip")
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		terminal := waitFailedJob(t, untrusted.Server, admin, job["job_id"].(string))
		requireFailedJobReason(t, terminal, "reference_pack_operation_rejected", "clock_untrusted")
		var incidents, candidates int
		if err := untrusted.DB.QueryRow(`SELECT (SELECT count(*) FROM incidents WHERE id=$1),(SELECT count(*) FROM reference_pack_candidates WHERE pack_key=$2 AND pack_version=$3)`, id, fixture.Key, fixture.Version).Scan(&incidents, &candidates); err != nil || incidents != 0 || candidates != 0 {
			t.Fatal("untrusted clock published state", incidents, candidates, err)
		}
	})
	t.Run("missing required bytes abort without a candidate", func(t *testing.T) {
		id, bundle := create("EMBEDDED-REQUIRED-MISSING")
		response := postImport(t, target.Server, targetAdmin, `{"client_txn_id":"embedded-missing-import"}`, bundle, "bundle.zip")
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		terminal := waitFailedJob(t, target.Server, targetAdmin, job["job_id"].(string))
		requireFailedJobReason(t, terminal, "incident_bundle_import_rejected", "source_family_invalid")
		var incidents, candidates int
		if err := target.DB.QueryRow(`SELECT (SELECT count(*) FROM incidents WHERE id=$1),(SELECT count(*) FROM reference_pack_candidates WHERE pack_key=$2 AND pack_version=$3)`, id, fixture.Key, fixture.Version).Scan(&incidents, &candidates); err != nil || incidents != 0 || candidates != 0 {
			t.Fatal("missing required bytes published state", incidents, candidates, err)
		}
	})
	files := zipMemberMap(t, fixture.Container)
	var targets map[string]any
	if err := json.Unmarshal(files["metadata/targets.json"], &targets); err != nil {
		t.Fatal(err)
	}
	signature := targets["signatures"].([]any)[0].(map[string]any)
	value := signature["sig"].(string)
	replacement := "0"
	if value[0] == '0' {
		replacement = "1"
	}
	signature["sig"] = replacement + value[1:]
	files["metadata/targets.json"], err = canonicaljson.Marshal(targets)
	if err != nil {
		t.Fatal(err)
	}
	invalid := writeZipMemberMap(t, files)
	for _, required := range []bool{true, false} {
		name := "optional"
		key := "EMBEDDED-INVALID-OPTIONAL"
		if required {
			name = "required"
			key = "EMBEDDED-INVALID-REQUIRED"
		}
		t.Run("invalid "+name+" signatures", func(t *testing.T) {
			id, bundle := create(key)
			var content map[string]any
			if err := json.Unmarshal(fixture.Content, &content); err != nil {
				t.Fatal(err)
			}
			descriptor := content["containers"].([]any)[0].(map[string]any)
			descriptor["container_sha256"] = hashHexBytes(invalid)
			descriptor["size_bytes"] = len(invalid)
			if !required {
				content["required_members"] = []any{}
			}
			encoded, err := json.Marshal(content)
			if err != nil {
				t.Fatal(err)
			}
			bundle = replaceStructuredBundleMember(t, bundle, "ext/reference_packs/content.json", encoded)
			bundle = replaceStructuredBundleMember(t, bundle, fixture.Path, invalid)
			response := postImport(t, target.Server, targetAdmin, `{"client_txn_id":"embedded-invalid-`+name+`"}`, bundle, "bundle.zip")
			job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
			if required {
				requireFailedJobReason(t, waitFailedJob(t, target.Server, targetAdmin, job["job_id"].(string)), "incident_bundle_import_rejected", "source_family_invalid")
			} else {
				waitJob(t, target.Server, targetAdmin, job["job_id"].(string))
			}
			var incidents, versions int
			var health string
			var failure *string
			if err := target.DB.QueryRow(`SELECT (SELECT count(*) FROM incidents WHERE id=$1),(SELECT count(*) FROM reference_pack_versions WHERE pack_key=$2 AND pack_version=$3),health,last_failure_code FROM reference_pack_candidates WHERE pack_key=$2 AND pack_version=$3`, id, fixture.Key, fixture.Version).Scan(&incidents, &versions, &health, &failure); err != nil {
				t.Fatal(err)
			}
			expected := 1
			if required {
				expected = 0
			}
			if incidents != expected || versions != 0 || health != "failed" || failure == nil || *failure != "signature_threshold_not_met" {
				t.Fatal("failed signature semantics", incidents, versions, health, failure)
			}
			var outcome string
			var catalogs, trustChanges int
			if err := target.DB.QueryRow(`SELECT a.outcome,
 (SELECT count(*) FROM reference_pack_portable_catalogs c WHERE c.operation_id=o.operation_id),
 (SELECT count(*) FROM reference_pack_events t WHERE t.operation_id=o.operation_id AND t.event_kind='trust_root_update')
 FROM reference_pack_operations o JOIN reference_pack_attempts a USING(operation_id) WHERE o.job_id=$1 AND a.completed_at IS NOT NULL`, job["job_id"].(string)).Scan(&outcome, &catalogs, &trustChanges); err != nil {
				t.Fatal(err)
			}
			if outcome != "content_rejected" || catalogs != expected || trustChanges != 0 {
				t.Fatal("rejection publication", outcome, catalogs, trustChanges)
			}
		})
	}
	id, bundle := create("EMBEDDED-VERIFIED")
	bundle = replaceStructuredBundleMember(t, bundle, fixture.Path, fixture.Container)
	terminal := importBundleAndWait(t, target.Server, targetAdmin, bundle, "embedded-verified-import")
	if terminal["status"] != "succeeded" {
		data, _ := json.Marshal(terminal)
		t.Fatalf("destination verification: %s", data)
	}
	var nowActive, health, operationKind string
	var envelopes, pins, attempts int
	if err := target.DB.QueryRow(`SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`).Scan(&nowActive); err != nil || nowActive != active {
		t.Fatal("import changed activation", nowActive, err)
	}
	if err := target.DB.QueryRow(`SELECT c.health,o.kind,(SELECT count(*) FROM reference_pack_envelopes e WHERE e.operation_id=o.operation_id),(SELECT count(*) FROM reference_pack_version_pins p WHERE p.operation_id=o.operation_id),(SELECT count(*) FROM reference_pack_attempts a WHERE a.operation_id=o.operation_id AND a.outcome='succeeded') FROM reference_pack_candidates c JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id JOIN reference_pack_operations o USING(operation_id) WHERE c.pack_key=$1 AND c.pack_version=$2`, fixture.Key, fixture.Version).Scan(&health, &operationKind, &envelopes, &pins, &attempts); err != nil || health != "verified_available" || operationKind != "portable_retention" || envelopes != 1 || pins != 1 || attempts != 1 {
		t.Fatal("destination evidence", health, operationKind, envelopes, pins, attempts, err)
	}
	var manifestDigest, payloadDigest string
	if err := target.DB.QueryRow(`SELECT manifest_sha256,payload_sha256 FROM reference_pack_versions WHERE pack_key=$1 AND pack_version=$2`, fixture.Key, fixture.Version).Scan(&manifestDigest, &payloadDigest); err != nil || manifestDigest != *expected.Manifest || payloadDigest != *expected.Payload || health != *expected.Condition {
		t.Fatal("canonical destination expectation differs", err)
	}
	var attestations, indexRows int
	if err := target.DB.QueryRow(`SELECT (SELECT count(*) FROM reference_pack_events WHERE pack_key=$1 AND pack_version=$2),(SELECT count(*) FROM reference_pack_indexes i JOIN reference_pack_candidates c ON c.current_index_id=i.index_id WHERE c.pack_key=$1 AND c.pack_version=$2)`, fixture.Key, fixture.Version).Scan(&attestations, &indexRows); err != nil || attestations == 0 || indexRows == 0 {
		t.Fatal("missing canonical destination side effect", err)
	}
	exported := exportBundleBytes(t, target, targetAdmin, id, "embedded-verified-reexport")
	if !bytes.Equal(zipMemberBytes(t, exported, "data/reference_pack_refs.json"), fixture.References) {
		t.Fatal("source references changed")
	}
	t.Run("exact reuse ignores redundant ineligible bytes", func(t *testing.T) {
		_, bundle := create("EMBEDDED-REUSE")
		bundle = replaceStructuredBundleMember(t, bundle, fixture.Path, []byte("redundant invalid container"))
		terminal := importBundleAndWait(t, target.Server, targetAdmin, bundle, "embedded-reuse-import")
		if terminal["status"] != "succeeded" {
			data, _ := json.Marshal(terminal)
			t.Fatalf("exact reuse: %s", data)
		}
		var count int
		if err := target.DB.QueryRow(`SELECT count(*) FROM reference_pack_envelopes WHERE pack_key=$1 AND pack_version=$2`, fixture.Key, fixture.Version).Scan(&count); err != nil || count != 1 {
			t.Fatal("reuse renewed verification", count, err)
		}
	})
	for _, name := range []string{"compatible", "conflicting", "publication-failure", "lost-acknowledgement", "commit-rollback"} {
		conflicting := name == "conflicting"
		publicationFailure := name == "publication-failure"
		commitRollback := name == "commit-rollback"
		commitFault := commitRollback || name == "lost-acknowledgement"
		t.Run("fresh cohort with "+name+" valid root proposals", func(t *testing.T) {
			bootstrapBytes, refs, content, containers := referencetest.PortableCohortFixture(t, time.Now().UTC(), vector, references, conflicting)
			rootPath := filepath.Join(t.TempDir(), "cohort-bootstrap.json")
			if err := os.WriteFile(rootPath, bootstrapBytes, 0600); err != nil {
				t.Fatal(err)
			}
			var fault *portableCommitFaultPool
			configure := func(options *server.Options) {
				if commitFault {
					fault = &portableCommitFaultPool{AdmittedPool: options.Postgres, rollback: commitRollback}
					options.Postgres = fault
				}
			}
			destination := startIsolatedIncidentBundleServerConfigured(t, runtime, "embedded-cohort-"+name, map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": rootPath, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}, configure)
			admin, _ := flowtest.ProvisionBootstrapAdmin(t, destination.Server.HTTP.URL)
			if publicationFailure {
				// The sequence is deliberate nontransactional test evidence:
				// value 63 proves all earlier owner writes were visible before
				// a late parent-transaction error rolled them back.
				if _, err := destination.DB.Exec(`CREATE SEQUENCE rp_fixture_abort_progress;
GRANT USAGE, UPDATE ON SEQUENCE rp_fixture_abort_progress TO cartulary_runtime;
CREATE FUNCTION rp_fixture_abort_catalog() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
 PERFORM setval('rp_fixture_abort_progress', 1
 + 2 * ((SELECT count(*) FROM reference_pack_envelopes WHERE operation_id=NEW.operation_id)=2)::int
 + 4 * ((SELECT count(*) FROM reference_pack_events WHERE operation_id=NEW.operation_id)=4)::int
 + 8 * ((SELECT count(*) FROM reference_pack_version_pins WHERE operation_id=NEW.operation_id)=2)::int
 + 16 * ((SELECT root_version FROM reference_pack_repositories WHERE repository_id='fixture.repo')=2)::int
 + 32 * (EXISTS(SELECT 1 FROM incidents WHERE id=NEW.incident_id))::int);
 RAISE EXCEPTION 'injected final catalog publication failure'; END $$;
CREATE TRIGGER rp_fixture_abort_catalog BEFORE INSERT ON reference_pack_portable_catalogs FOR EACH ROW EXECUTE FUNCTION rp_fixture_abort_catalog();`); err != nil {
					t.Fatal(err)
				}
			}

			var beforeSet string
			var beforeRevision int
			if err := destination.DB.QueryRow(`SELECT c.pack_set_id,r.revision FROM reference_pack_current_set c CROSS JOIN reference_pack_repositories r WHERE c.singleton AND r.repository_id='fixture.repo'`).Scan(&beforeSet, &beforeRevision); err != nil {
				t.Fatal(err)
			}
			incidentID, bundle := create("EMBEDDED-COHORT-" + name)
			bundle = replaceStructuredBundleMember(t, bundle, "data/reference_pack_refs.json", refs)
			bundle = replaceStructuredBundleMember(t, bundle, "ext/reference_packs/content.json", content)
			for path, data := range containers {
				bundle = replaceStructuredBundleMember(t, bundle, path, data)
			}
			metadata := `{"client_txn_id":"cohort-` + name + `"}`
			response := postImport(t, destination.Server, admin, metadata, bundle, "bundle.zip")
			jobID := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
			if conflicting {
				requireFailedJobReason(t, waitFailedJob(t, destination.Server, admin, jobID), "reference_pack_operation_rejected", "stale_admission_state")
			} else if publicationFailure || commitRollback {
				waitFailedJob(t, destination.Server, admin, jobID)
			} else {
				waitJob(t, destination.Server, admin, jobID)
			}
			if commitFault && (fault == nil || fault.injected.Load() != 1) {
				t.Fatal("parent commit fault did not execute exactly once")
			}
			var afterSet, outcome string
			var rootVersion, revision, roots, incidents, versions, envelopes, pins, events, catalogs, verified int
			if err := destination.DB.QueryRow(`SELECT c.pack_set_id,r.root_version,r.revision,(SELECT count(*) FROM reference_pack_roots WHERE repository_id='fixture.repo') FROM reference_pack_current_set c CROSS JOIN reference_pack_repositories r WHERE c.singleton AND r.repository_id='fixture.repo'`).Scan(&afterSet, &rootVersion, &revision, &roots); err != nil {
				t.Fatal(err)
			}
			if err := destination.DB.QueryRow(`SELECT a.outcome,
 (SELECT count(*) FROM incidents WHERE id=$2),
 (SELECT count(*) FROM reference_pack_versions WHERE pack_key='type_registry.host' AND pack_version IN ('z-older','a-newer')),
 (SELECT count(*) FROM reference_pack_envelopes WHERE operation_id=o.operation_id),
 (SELECT count(*) FROM reference_pack_version_pins WHERE operation_id=o.operation_id),
 (SELECT count(*) FROM reference_pack_events WHERE operation_id=o.operation_id),
 (SELECT count(*) FROM reference_pack_portable_catalogs WHERE operation_id=o.operation_id),
 (SELECT count(*) FROM reference_pack_attempt_members WHERE attempt_id=a.attempt_id AND verdict='succeeded')
 FROM reference_pack_operations o JOIN reference_pack_attempts a USING(operation_id) WHERE o.job_id=$1 AND a.completed_at IS NOT NULL`, jobID, incidentID).Scan(&outcome, &incidents, &versions, &envelopes, &pins, &events, &catalogs, &verified); err != nil {
				t.Fatal(err)
			}
			if afterSet != beforeSet || verified != 2 {
				t.Fatal("cohort changed activation or did not verify both proposals", verified)
			}
			if conflicting || publicationFailure || commitRollback {
				expectedOutcome := "stale_state"
				if publicationFailure || commitRollback {
					expectedOutcome = "execution_failed"
				}
				if publicationFailure {
					var progress int
					if err := destination.DB.QueryRow(`SELECT last_value FROM rp_fixture_abort_progress`).Scan(&progress); err != nil || progress != 63 {
						t.Fatal("late fault did not observe complete parent mutation", progress, err)
					}
				}
				if outcome != expectedOutcome || rootVersion != 1 || revision != beforeRevision || roots != 1 || incidents != 0 || versions != 0 || envelopes != 0 || pins != 0 || events != 0 || catalogs != 0 {
					t.Fatal("conflicting proposals published a partial delta", outcome, rootVersion, revision, roots, incidents, versions, envelopes, pins, events, catalogs)
				}
				var condemned int
				if err := destination.DB.QueryRow(`SELECT count(*) FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version IN ('z-older','a-newer') AND (health<>'failed' OR last_failure_code IS NOT NULL OR current_envelope_id IS NOT NULL)`).Scan(&condemned); err != nil || condemned != 0 {
					t.Fatal("operational abort fabricated a content verdict", condemned, err)
				}
			} else {
				if outcome != "succeeded" || rootVersion != 2 || revision != beforeRevision+1 || roots != 2 || incidents != 1 || versions != 2 || envelopes != 2 || pins != 2 || events != 4 || catalogs != 1 {
					t.Fatal("complete cohort publication", outcome, rootVersion, revision, roots, incidents, versions, envelopes, pins, events, catalogs)
				}
				var instants int
				if err := destination.DB.QueryRow(`SELECT count(DISTINCT verified_at) FROM reference_pack_envelopes e JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1`, jobID).Scan(&instants); err != nil || instants != 1 {
					t.Fatal("cohort used multiple freshness instants", instants, err)
				}
				exported := exportBundleBytes(t, destination, admin, incidentID, "cohort-reexport")
				if !bytes.Equal(zipMemberBytes(t, exported, "data/reference_pack_refs.json"), refs) {
					t.Fatal("multiple historical sets changed on re-export")
				}
			}
			replayed := httptestx.RequireSuccessEnvelope(t, postImport(t, destination.Server, admin, metadata, bundle, "bundle.zip"), http.StatusAccepted)["data"].(map[string]any)
			if replayed["job_id"] != jobID {
				t.Fatal("cohort replay admitted a new operation")
			}
			var attempts int
			if err := destination.DB.QueryRow(`SELECT count(*) FROM reference_pack_attempts a JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1`, jobID).Scan(&attempts); err != nil || attempts != 1 {
				t.Fatal("cohort replay created new attempt", attempts, err)
			}
		})
	}

}
