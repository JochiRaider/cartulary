package reporting_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	reportingjson "github.com/JochiRaider/cartulary/internal/modules/reporting/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func assertNativeReleaseReadiness(t *testing.T, harness *appsupport.ServerHarness, release string) {
	t.Helper()
	ctx := context.Background()
	references := reportingReadinessReferences(t, harness)
	ready := func() error { return reporting.ValidateRetainedReleases(ctx, harness.Pool, references) }
	if err := ready(); err != nil {
		t.Fatal("valid native releases rejected", err)
	}
	id := uuid.MustParse(release)
	var jobID uuid.UUID
	var raw, input, summary []byte
	var rendered reporting.RenderedRelease
	var txn string
	var graphDigest string
	if err := harness.DB.QueryRow(`SELECT p.admitted_graph_sources_sha256 FROM reporting_job_payloads p JOIN reporting_releases r ON r.create_job_id=p.job_id WHERE r.release_id=$1`, id).Scan(&graphDigest); err != nil {
		t.Fatal(err)
	}
	if err := harness.DB.QueryRow(`SELECT r.create_job_id,r.client_txn_id,r.output_sha256,r.output_media_type,r.redaction_profile_sha256,r.redaction_manifest_sha256,r.redaction_manifest_json,b.bundle_manifest_sha256,b.bundle_manifest_json,b.primary_bundle_path,b.primary_media_type,j.request_json,p.result_summary_json FROM reporting_releases r JOIN reporting_render_bundles b USING(release_id) JOIN reporting_job_payloads j ON j.job_id=r.create_job_id JOIN jobs p ON p.job_id=r.create_job_id WHERE r.release_id=$1`, id).Scan(
		&jobID, &txn, &rendered.OutputSHA256, &rendered.OutputMediaType, &rendered.ProfileSHA256, &rendered.RedactionManifestSHA256, &raw,
		&rendered.RenderBundle.ManifestSHA256, &rendered.RenderBundle.ManifestJSON, &rendered.RenderBundle.PrimaryPath, &rendered.RenderBundle.PrimaryMedia, &input, &summary); err != nil {
		t.Fatal(err)
	}
	var err error
	rendered.RedactionManifestJSON, err = reportingjson.Canonicalize(raw)
	if err != nil {
		t.Fatal(err)
	}
	rendered.RenderBundle.ManifestJSON, err = reportingjson.Canonicalize(rendered.RenderBundle.ManifestJSON)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(rendered.RenderBundle.ManifestJSON, &rendered.RenderBundle.Manifest); err != nil {
		t.Fatal(err)
	}
	var admitted struct {
		Profile reporting.RedactionProfile `json:"redaction_profile"`
	}
	if err := json.Unmarshal(input, &admitted); err != nil {
		t.Fatal(err)
	}
	rendered.Profile = requirePreparedRenderProfile(t, harness, jobID)
	for _, descriptor := range rendered.RenderBundle.Manifest.Files {
		f := reporting.RenderBundleFile{RenderBundleManifestFile: descriptor, StorageKind: "database_inline"}
		if err := harness.DB.QueryRow(`SELECT inline_bytes FROM reporting_render_bundle_files WHERE release_id=$1 AND bundle_path=$2`, id, f.Path).Scan(&f.Bytes); err != nil {
			t.Fatal(err)
		}
		rendered.RenderBundle.Files = append(rendered.RenderBundle.Files, f)
		if f.Role == "token_manifest" {
			rendered.Redaction.TokenManifestJSON = f.Bytes
		}
	}
	rendered.Redaction.TokenManifestSHA256 = *rendered.RenderBundle.Manifest.TokenManifestSHA256
	if err := harness.DB.QueryRow(`SELECT canonical_bytes,reveal_map_sha256 FROM reporting_release_reveal_maps WHERE release_id=$1`, id).Scan(&rendered.Redaction.RevealMapJSON, &rendered.Redaction.RevealMapSHA256); err != nil {
		t.Fatal(err)
	}
	complete := func(candidate reporting.RenderedRelease, target uuid.UUID) error {
		proof, err := reporting.ValidateRenderCandidate(ctx, harness.Pool, jobID, candidate)
		if err != nil {
			return err
		}
		tx, err := harness.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		return (&reporting.Store{}).CompleteReleaseCreateJobTx(ctx, tx, jobID, target, proof, time.Now().UTC())
	}
	if err := complete(rendered, id); err != nil {
		t.Fatal("identical completion rejected", err)
	}
	assertPreparedCandidateIsolation(t, harness, jobID, rendered, func(proof reporting.RenderCandidate) error {
		tx, err := harness.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		return (&reporting.Store{}).CompleteReleaseCreateJobTx(ctx, tx, jobID, id, proof, time.Now().UTC())
	})
	if complete(rendered, uuid.New()) == nil {
		t.Fatal("completion changed release identity")
	}
	changed := rendered
	changed.OutputMediaType = "application/octet-stream"
	if complete(changed, id) == nil {
		t.Fatal("completion changed output binding")
	}
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := harness.DB.Exec(sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, mutation := range []struct {
		name, change, restore string
		changed, original     any
	}{
		{"transaction", `UPDATE reporting_releases SET client_txn_id=$2 WHERE release_id=$1`, `UPDATE reporting_releases SET client_txn_id=$2 WHERE release_id=$1`, "substituted", txn},
		{"profile digest", `UPDATE reporting_releases SET redaction_profile_sha256=$2 WHERE release_id=$1`, `UPDATE reporting_releases SET redaction_profile_sha256=$2 WHERE release_id=$1`, string(bytes.Repeat([]byte("0"), 64)), rendered.ProfileSHA256},
		{"output digest", `UPDATE reporting_releases SET output_sha256=$2 WHERE release_id=$1`, `UPDATE reporting_releases SET output_sha256=$2 WHERE release_id=$1`, string(bytes.Repeat([]byte("0"), 64)), rendered.OutputSHA256},
		{"payload unknown", `UPDATE reporting_job_payloads SET request_json=request_json || $2::jsonb WHERE job_id=(SELECT create_job_id FROM reporting_releases WHERE release_id=$1)`, `UPDATE reporting_job_payloads SET request_json=$2 WHERE job_id=(SELECT create_job_id FROM reporting_releases WHERE release_id=$1)`, `{"unadmitted":true}`, input},
		{"independent graph digest", `UPDATE reporting_job_payloads SET admitted_graph_sources_sha256=$2 WHERE job_id=(SELECT create_job_id FROM reporting_releases WHERE release_id=$1)`, `UPDATE reporting_job_payloads SET admitted_graph_sources_sha256=$2 WHERE job_id=(SELECT create_job_id FROM reporting_releases WHERE release_id=$1)`, string(bytes.Repeat([]byte("0"), 64)), graphDigest},
		{"member bytes", `UPDATE reporting_render_bundle_files SET inline_bytes=$2 WHERE release_id=$1 AND bundle_path='validation/summary.json'`, `UPDATE reporting_render_bundle_files SET inline_bytes=$2 WHERE release_id=$1 AND bundle_path='validation/summary.json'`, []byte(`{}`), nil},
	} {
		t.Run("retained release "+mutation.name, func(t *testing.T) {
			if mutation.original == nil {
				var original []byte
				if err := harness.DB.QueryRow(`SELECT inline_bytes FROM reporting_render_bundle_files WHERE release_id=$1 AND bundle_path='validation/summary.json'`, id).Scan(&original); err != nil {
					t.Fatal(err)
				}
				mutation.original = original
			}
			if _, err := harness.DB.Exec(mutation.change, id, mutation.changed); err != nil {
				var constraint *pgconn.PgError
				if errors.As(err, &constraint) && constraint.Code == "23514" {
					return
				}
				t.Fatal(err)
			}
			defer exec(mutation.restore, id, mutation.original)
			if ready() == nil {
				t.Fatal("recovery admitted substituted release")
			}
			if complete(rendered, id) == nil {
				t.Fatal("repeat completion bypassed retained integrity")
			}
		})
	}
	t.Run("independent Jobs outcome", func(t *testing.T) {
		exec(`UPDATE jobs SET result_summary_json=jsonb_set(result_summary_json,'{resource_refs,0,id}','"69000000-0000-4000-8000-000000000999"') WHERE job_id=$1`, jobID)
		defer exec(`UPDATE jobs SET result_summary_json=$2 WHERE job_id=$1`, jobID, summary)
		if ready() == nil {
			t.Fatal("recovery accepted another Job result")
		}
		if complete(rendered, id) == nil {
			t.Fatal("repeated completion accepted another Job result")
		}
	})
	if err := ready(); err != nil {
		t.Fatal("restored native releases rejected", err)
	}
	compactNativeReleaseTestJob(t, harness, jobID)
	if err := ready(); err != nil {
		t.Fatal("public response expiry invalidated native release", err)
	}
	if err := complete(rendered, id); err != nil {
		t.Fatal("expired native completion rejected", err)
	}
	assertRehashedPresentationReadiness(t, harness, id, rendered, ready)
}

func assertRehashedPresentationReadiness(t *testing.T, harness *appsupport.ServerHarness, id uuid.UUID, original reporting.RenderedRelease, ready func() error) {
	t.Helper()
	for index, file := range original.RenderBundle.Files {
		if file.Role != "deck_model" && file.Role != "source_slidev" {
			continue
		}
		t.Run("rehashed retained "+file.Role, func(t *testing.T) {
			changed := append([]byte{}, file.Bytes...)
			if file.Role == "source_slidev" {
				changed = append(changed, []byte("\nSubstituted presentation.\n")...)
			} else {
				var deck map[string]any
				if err := json.Unmarshal(changed, &deck); err != nil {
					t.Fatal(err)
				}
				deck["title"] = "Substituted presentation"
				var err error
				changed, err = reportingjson.Marshal(deck)
				if err != nil {
					t.Fatal(err)
				}
			}
			manifest := original.RenderBundle.Manifest
			manifest.Files = append([]reporting.RenderBundleManifestFile{}, manifest.Files...)
			changedHash := fmt.Sprintf("%x", sha256.Sum256(changed))
			manifest.Files[index].SHA256, manifest.Files[index].SizeBytes = changedHash, int64(len(changed))
			manifestJSON, err := reportingjson.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			manifestHash := fmt.Sprintf("%x", sha256.Sum256(append([]byte(manifest.SchemaID+"\n"), manifestJSON...)))
			write := func(content []byte, contentHash string, root []byte, rootHash string) {
				t.Helper()
				ctx := context.Background()
				tx, err := harness.Pool.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback(ctx)
				if _, err := tx.Exec(ctx, `UPDATE reporting_render_bundle_files SET inline_bytes=$3,file_sha256=$4,size_bytes=$5 WHERE release_id=$1 AND bundle_path=$2`, id, file.Path, content, contentHash, len(content)); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `UPDATE reporting_render_bundles SET bundle_manifest_json=$2,bundle_manifest_sha256=$3 WHERE release_id=$1`, id, root, rootHash); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `UPDATE reporting_releases SET output_sha256=$2 WHERE release_id=$1`, id, rootHash); err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(ctx); err != nil {
					t.Fatal(err)
				}
			}
			write(changed, changedHash, manifestJSON, manifestHash)
			defer write(file.Bytes, file.SHA256, original.RenderBundle.ManifestJSON, original.RenderBundle.ManifestSHA256)
			if err := ready(); err == nil || err.Error() != "reporting: retained presentation differs from admission" {
				t.Fatal("recovery did not reject rehashed presentation against frozen input", err)
			}
		})
	}
	if err := ready(); err != nil {
		t.Fatal("restored presentation rejected", err)
	}
}

func assertFailedReleaseReadiness(t *testing.T, harness *appsupport.ServerHarness, release string) {
	t.Helper()
	ctx := context.Background()
	references := reportingReadinessReferences(t, harness)
	if err := reporting.ValidateRetainedReleases(ctx, harness.Pool, references); err != nil {
		t.Fatal("legitimate failed history rejected", err)
	}
	var jobID uuid.UUID
	var reason string
	var diagnostic []byte
	if err := harness.DB.QueryRow(`SELECT r.create_job_id,r.render_failed_reason_code,p.render_failure_summary FROM reporting_releases r JOIN reporting_job_payloads p ON p.job_id=r.create_job_id WHERE r.release_id=$1`, release).Scan(&jobID, &reason, &diagnostic); err != nil {
		t.Fatal(err)
	}
	var summary struct {
		Stage     string         `json:"terminal_stage"`
		Code      string         `json:"failure_code"`
		Details   map[string]any `json:"safe_details"`
		ReleaseID string         `json:"release_id"`
		Result    string         `json:"result"`
	}
	if err := json.Unmarshal(diagnostic, &summary); err != nil || summary.ReleaseID != release || summary.Result != "failed" {
		t.Fatal("missing durable failure summary", err)
	}
	for _, altered := range []bool{false, true} {
		tx, err := harness.Pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		candidate := reason
		if altered {
			candidate = "different_failure"
		}
		id, err := (&reporting.Store{}).CompleteRenderFailureJobTx(ctx, tx, jobID, reporting.RenderFailure{Stage: summary.Stage, FailureCode: summary.Code, ReasonCode: candidate, SafeDetails: summary.Details}, time.Now().UTC())
		_ = tx.Rollback(ctx)
		if altered && err == nil {
			t.Fatal("failure replay replaced original reason")
		}
		if !altered && (err != nil || id == nil || id.String() != release) {
			t.Fatal("identical failure replay rejected", err)
		}
	}
	compactNativeReleaseTestJob(t, harness, jobID)
	if err := reporting.ValidateRetainedReleases(ctx, harness.Pool, references); err != nil {
		t.Fatal("expired failure history rejected", err)
	}
	replay := func() error {
		tx, err := harness.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		_, err = (&reporting.Store{}).CompleteRenderFailureJobTx(ctx, tx, jobID, reporting.RenderFailure{Stage: summary.Stage, FailureCode: summary.Code, ReasonCode: reason, SafeDetails: summary.Details}, time.Now().UTC())
		return err
	}
	t.Run("missing failed release outcome", func(t *testing.T) {
		var row []byte
		if err := harness.DB.QueryRow(`SELECT row_to_json(r) FROM reporting_releases r WHERE release_id=$1`, release).Scan(&row); err != nil {
			t.Fatal(err)
		}
		if _, err := harness.DB.Exec(`DELETE FROM reporting_releases WHERE release_id=$1`, release); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if _, err := harness.DB.Exec(`INSERT INTO reporting_releases SELECT (json_populate_record(NULL::reporting_releases,$1)).*`, row); err != nil {
				t.Fatal(err)
			}
		}()
		if reporting.ValidateRetainedReleases(ctx, harness.Pool, references) == nil || replay() == nil {
			t.Fatal("terminal outcome disappeared from recovery or was reconstructed")
		}
	})
	assertRetainedFailureIntegrity(t, harness, jobID, func() error { return reporting.ValidateRetainedReleases(ctx, harness.Pool, references) }, replay)
}

// Exercise storage integrity independently of the producer, including a
// correctly rehashed substitution and missing bytes after public Jobs expiry.
func assertRetainedFailureIntegrity(t *testing.T, harness *appsupport.ServerHarness, jobID uuid.UUID, ready, replay func() error) {
	t.Helper()
	var original []byte
	var digest string
	if err := harness.DB.QueryRow(`SELECT render_failure_summary,render_failure_sha256 FROM reporting_job_payloads WHERE job_id=$1`, jobID).Scan(&original, &digest); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"missing", "digest", "rehashed admission time", "rehashed identity", "rehashed issue", "rehashed raw detail"} {
		t.Run("retained failure "+mutation, func(t *testing.T) {
			var changed map[string]any
			if err := json.Unmarshal(original, &changed); err != nil {
				t.Fatal(err)
			}
			switch mutation {
			case "rehashed admission time":
				changed["created_at"] = "2026-01-01T00:00:00.000000Z"
			case "rehashed identity":
				changed["release_id"] = uuid.NewString()
			case "rehashed issue":
				changed["first_failure"].(map[string]any)["reason_code"] = "render_timeout"
			case "rehashed raw detail":
				changed["issues"].([]any)[0].(map[string]any)["safe_details"].(map[string]any)["raw_error"] = "private-error-detail"
			}
			encoded, err := reportingjson.Marshal(changed)
			if err != nil {
				t.Fatal(err)
			}
			candidateDigest := fmt.Sprintf("%x", sha256.Sum256(append([]byte("cartulary.reporting_render_validation_summary.v1\n"), encoded...)))
			var raw, hash any = encoded, candidateDigest
			if mutation == "missing" {
				raw, hash = nil, nil
			} else if mutation == "digest" {
				hash = string(bytes.Repeat([]byte("0"), 64))
			}
			if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET render_failure_summary=$2,render_failure_sha256=$3 WHERE job_id=$1`, jobID, raw, hash); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET render_failure_summary=$2,render_failure_sha256=$3 WHERE job_id=$1`, jobID, original, digest); err != nil {
					t.Fatal(err)
				}
			}()
			if ready() == nil {
				t.Fatal("recovery accepted missing or substituted failure diagnostics")
			}
			if replay != nil && replay() == nil {
				t.Fatal("terminal replay repaired or replaced retained diagnostics")
			}
		})
	}
	if err := ready(); err != nil {
		t.Fatal("restored failure diagnostics rejected", err)
	}
}

func compactNativeReleaseTestJob(t *testing.T, harness *appsupport.ServerHarness, jobID uuid.UUID) {
	t.Helper()
	// Model a legitimate compacted Jobs tombstone; this does not exercise the
	// Jobs expiry scheduler or change application clocks.
	if _, err := harness.DB.Exec(`UPDATE jobs SET retained_until=finished_at,expired_at=now(),handler_payload_json=NULL,handler_attempt_id=NULL,handler_lease_expires_at=NULL,handler_failure_count=0,handler_next_attempt_at=NULL,handler_last_attempted_at=NULL,handler_last_error=NULL,result_summary_json=NULL,error_summary_json=NULL,message=NULL,extension_idempotency_identity=NULL,extension_idempotency_route_key=NULL,extension_idempotency_scope_key=NULL,extension_normalized_request_sha256=NULL WHERE job_id=$1`, jobID); err != nil {
		t.Fatal(err)
	}
}

// The calling scenario has drained its runner. Admissions remain real HTTP
// requests; claimed cancellation and owner finalization use production ports.
func assertUnfinishedReleaseReadiness(t *testing.T, harness *appsupport.ServerHarness, manager *jobs.Manager, transactions *jobs.TransactionService, actor string, create func(string) map[string]any) {
	t.Helper()
	ctx := context.Background()
	refs := reportingReadinessReferences(t, harness)
	ready := func() error { return reporting.ValidateRetainedReleases(ctx, harness.Pool, refs) }
	finalizer := reportingFixtureFinalizer(t, harness, transactions)
	for _, state := range []string{"queued", "running", "canceled", "cleanup failure"} {
		admitted := create("txn-retained-release-" + state)
		id := uuid.MustParse(admitted["job_id"].(string))
		if err := ready(); err != nil {
			t.Fatal("queued admission rejected", err)
		}
		if state == "queued" {
			continue
		}
		canceled := state == "canceled" || state == "cleanup failure"
		if canceled {
			if _, err := manager.Cancel(ctx, jobs.CancelParams{JobID: id, ActorUserID: uuid.MustParse(actor), ClientTxnID: "cancel-queued-release", NormalizedRequest: []byte(`{"client_txn_id":"cancel-queued-release","reason":null}`)}); err != nil {
				t.Fatal(err)
			}
		}
		execution, claimed, err := manager.Claim(ctx, id)
		if err != nil || !claimed {
			t.Fatal("claim pending release", claimed, err)
		}
		if canceled {
			failure := reporting.RenderFailure{Stage: "release_state", FailureCode: "render_canceled", ReasonCode: "render_canceled", SafeDetails: map[string]any{}}
			if state == "cleanup failure" {
				failure.Stage, failure.FailureCode, failure.ReasonCode = "security_sandbox", "security_sandbox_violation", "partial_output_cleanup_failed"
			}
			var releaseID *uuid.UUID
			_, err = finalizer.FinalizeReportingJobCancellation(ctx, reporting.JobCancellationFinalization{Execution: execution, Completion: jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 0}}, Mutate: func(ctx context.Context, tx pgx.Tx) error {
				var err error
				releaseID, err = (&reporting.Store{}).CompleteRenderFailureJobTx(ctx, tx, id, failure, time.Now().UTC())
				return err
			}})
			if err != nil || releaseID != nil {
				t.Fatal("unprepared cancellation manufactured a release", releaseID, err)
			}
			var outcomes int
			if err := harness.DB.QueryRow(`SELECT count(*) FROM reporting_releases WHERE create_job_id=$1`, id).Scan(&outcomes); err != nil || outcomes != 0 {
				t.Fatal("unprepared cancellation retained a release", outcomes, err)
			}
			assertRetainedFailureIntegrity(t, harness, id, ready, nil)
			compactNativeReleaseTestJob(t, harness, id)

		}
		if err := ready(); err != nil {
			t.Fatal("unfinished release rejected", state, err)
		}
	}
}

func assertPreparedCandidateIsolation(t *testing.T, harness *appsupport.ServerHarness, jobID uuid.UUID, rendered reporting.RenderedRelease, complete func(reporting.RenderCandidate) error) {
	t.Helper()
	if complete(reporting.RenderCandidate{}) == nil {
		t.Fatal("zero publication proof accepted")
	}
	raw, err := json.Marshal(rendered)
	if err != nil {
		t.Fatal(err)
	}
	var supplied reporting.RenderedRelease
	if err := json.Unmarshal(raw, &supplied); err != nil {
		t.Fatal(err)
	}
	proof, err := reporting.ValidateRenderCandidate(context.Background(), harness.Pool, jobID, supplied)
	if err != nil {
		t.Fatal(err)
	}
	if len(supplied.RenderBundle.Files) == 0 || len(supplied.RenderBundle.Files[0].Bytes) == 0 {
		t.Fatal("publication fixture lacks member bytes")
	}
	supplied.RenderBundle.Files[0].Bytes[0] ^= 1
	if err := complete(proof); err != nil {
		t.Fatal("caller mutation changed detached publication proof", err)
	}
	var digest string
	if err := harness.DB.QueryRow(`SELECT render_preparation_sha256 FROM reporting_job_payloads WHERE job_id=$1`, jobID).Scan(&digest); err != nil {
		t.Fatal(err)
	}
	if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET render_preparation_sha256=$2 WHERE job_id=$1`, jobID, string(bytes.Repeat([]byte("0"), 64))); err != nil {
		t.Fatal(err)
	}
	rejected := complete(proof)
	if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET render_preparation_sha256=$2 WHERE job_id=$1`, jobID, digest); err != nil {
		t.Fatal(err)
	}
	if rejected == nil {
		t.Fatal("publication accepted changed preparation after validation")
	}
	if err := complete(proof); err != nil {
		t.Fatal("restored preparation rejected", err)
	}
}
