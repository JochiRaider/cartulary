package reporting_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	reportingjson "github.com/JochiRaider/cartulary/internal/modules/reporting/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Extends the owner-routed real preview scenario, including tokenized output.
func assertNativePreviewReadiness(t *testing.T, harness *appsupport.ServerHarness, preview string) {
	t.Helper()
	ctx := context.Background()
	id := uuid.MustParse(preview)
	references := reportingReadinessReferences(t, harness)
	ready := func() error { return reporting.ValidateRetainedPreviews(ctx, harness.Pool, references) }
	if err := ready(); err != nil {
		t.Fatal("legitimate previews rejected", err)
	}
	var jobID uuid.UUID
	var input, summary, raw []byte
	var admittedAt time.Time
	var rendered reporting.RenderedRelease
	b := &rendered.RenderBundle
	if err := harness.DB.QueryRow(`SELECT o.render_attempt_id,o.output_sha256,o.output_media_type,o.redaction_profile_sha256,o.redaction_manifest_sha256,o.redaction_manifest_json,o.bundle_manifest_sha256,o.bundle_manifest_json,o.primary_bundle_path,o.primary_media_type,j.request_json,j.created_at,p.result_summary_json FROM reporting_composition_preview_outputs o JOIN reporting_job_payloads j ON j.job_id=o.render_attempt_id JOIN jobs p ON p.job_id=o.render_attempt_id WHERE o.preview_attempt_id=$1`, id).Scan(&jobID, &rendered.OutputSHA256, &rendered.OutputMediaType, &rendered.ProfileSHA256, &rendered.RedactionManifestSHA256, &raw, &b.ManifestSHA256, &b.ManifestJSON, &b.PrimaryPath, &b.PrimaryMedia, &input, &admittedAt, &summary); err != nil {
		t.Fatal(err)
	}
	var err error
	rendered.RedactionManifestJSON, err = reportingjson.Canonicalize(raw)
	if err != nil {
		t.Fatal(err)
	}
	b.ManifestJSON, err = reportingjson.Canonicalize(b.ManifestJSON)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b.ManifestJSON, &b.Manifest); err != nil {
		t.Fatal(err)
	}
	var admitted struct {
		Profile reporting.RedactionProfile `json:"redaction_profile"`
	}
	if err := json.Unmarshal(input, &admitted); err != nil {
		t.Fatal(err)
	}
	rendered.Profile = requirePreparedRenderProfile(t, harness, jobID)
	for _, descriptor := range b.Manifest.Files {
		f := reporting.RenderBundleFile{RenderBundleManifestFile: descriptor, StorageKind: "database_inline"}
		if err := harness.DB.QueryRow(`SELECT inline_bytes FROM reporting_composition_preview_output_files WHERE preview_attempt_id=$1 AND bundle_path=$2`, id, f.Path).Scan(&f.Bytes); err != nil {
			t.Fatal(err)
		}
		b.Files = append(b.Files, f)
		if f.Role == "token_manifest" {
			rendered.Redaction.TokenManifestJSON = f.Bytes
		}
	}
	rendered.Redaction.TokenManifestSHA256 = *b.Manifest.TokenManifestSHA256
	if err := harness.DB.QueryRow(`SELECT canonical_bytes,reveal_map_sha256 FROM reporting_preview_reveal_maps WHERE preview_attempt_id=$1`, id).Scan(&rendered.Redaction.RevealMapJSON, &rendered.Redaction.RevealMapSHA256); err != nil {
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
		return (&reporting.Store{}).CompleteCompositionPreviewJobTx(ctx, tx, jobID, target, proof, time.Now().UTC())
	}
	if err := complete(rendered, id); err != nil {
		t.Fatal("identical preview completion", err)
	}
	assertPreparedCandidateIsolation(t, harness, jobID, rendered, func(proof reporting.RenderCandidate) error {
		tx, err := harness.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		return (&reporting.Store{}).CompleteCompositionPreviewJobTx(ctx, tx, jobID, id, proof, time.Now().UTC())
	})
	if complete(rendered, uuid.New()) == nil {
		t.Fatal("preview completion changed identity")
	}
	changed := rendered
	changed.OutputSHA256 = string(bytes.Repeat([]byte("0"), 64))
	if complete(changed, id) == nil {
		t.Fatal("preview completion changed output")
	}
	exec := func(t *testing.T, query string, args ...any) {
		t.Helper()
		if _, err := harness.DB.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	var validationBytes []byte
	if err := harness.DB.QueryRow(`SELECT inline_bytes FROM reporting_composition_preview_output_files WHERE preview_attempt_id=$1 AND bundle_path='validation/summary.json'`, id).Scan(&validationBytes); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct {
		name, query, restore   string
		key, changed, original any
	}{
		{"unknown input", `UPDATE reporting_job_payloads SET request_json=request_json || $2::jsonb WHERE job_id=$1`, `UPDATE reporting_job_payloads SET request_json=$2 WHERE job_id=$1`, jobID, `{"unadmitted":true}`, input},
		{"normalized request", `UPDATE reporting_job_payloads SET request_json=jsonb_set(request_json,'{normalized_request}',$2::jsonb) WHERE job_id=$1`, `UPDATE reporting_job_payloads SET request_json=$2 WHERE job_id=$1`, jobID, `"e30="`, input},
		{"source admission time", `UPDATE report_composition_preview_attempts SET created_at=$2 WHERE preview_attempt_id=$1`, `UPDATE report_composition_preview_attempts SET created_at=$2 WHERE preview_attempt_id=$1`, id, admittedAt.Add(time.Second), admittedAt},
		{"row profile", `UPDATE reporting_composition_preview_outputs SET redaction_profile_sha256=$2 WHERE preview_attempt_id=$1`, `UPDATE reporting_composition_preview_outputs SET redaction_profile_sha256=$2 WHERE preview_attempt_id=$1`, id, string(bytes.Repeat([]byte("0"), 64)), rendered.ProfileSHA256},
		{"member bytes", `UPDATE reporting_composition_preview_output_files SET inline_bytes=$2 WHERE preview_attempt_id=$1 AND bundle_path='validation/summary.json'`, `UPDATE reporting_composition_preview_output_files SET inline_bytes=$2 WHERE preview_attempt_id=$1 AND bundle_path='validation/summary.json'`, id, []byte(`{}`), validationBytes},
		{"Jobs outcome", `UPDATE jobs SET result_summary_json=jsonb_set(result_summary_json,'{code}',$2::jsonb) WHERE job_id=$1`, `UPDATE jobs SET result_summary_json=$2 WHERE job_id=$1`, jobID, `"snapshot_created"`, summary},
	} {
		t.Run("retained preview "+mutation.name, func(t *testing.T) {
			exec(t, mutation.query, mutation.key, mutation.changed)
			defer exec(t, mutation.restore, mutation.key, mutation.original)
			if ready() == nil {
				t.Fatal("recovery accepted substituted preview")
			}
			if complete(rendered, id) == nil {
				t.Fatal("repeat completion bypassed preview integrity")
			}
		})
	}
	t.Run("missing retained input", func(t *testing.T) {
		// A private input can be absent even though the Composition source and Job
		// survive. The owner enumeration must keep this defect visible.
		var retainedRow []byte
		if err := harness.DB.QueryRow(`SELECT row_to_json(p) FROM reporting_job_payloads p WHERE job_id=$1`, jobID).Scan(&retainedRow); err != nil {
			t.Fatal(err)
		}
		exec(t, `DELETE FROM reporting_job_payloads WHERE job_id=$1`, jobID)
		defer exec(t, `INSERT INTO reporting_job_payloads SELECT * FROM json_populate_record(NULL::reporting_job_payloads,$1::json)`, retainedRow)

		if ready() == nil {
			t.Fatal("missing preview input inferred from current source")
		}
	})
	if err := ready(); err != nil {
		t.Fatal("restored preview history rejected", err)
	}
	compactNativeReleaseTestJob(t, harness, jobID)
	if err := ready(); err != nil {
		t.Fatal("expired preview admission rejected", err)
	}
	if err := complete(rendered, id); err != nil {
		t.Fatal("expired preview completion rejected", err)
	}
}

// Stop only this harness's runner after the live rendering/restore assertions.
// New requests then retain genuine queued inputs without racing execution.
func assertUnfinishedPreviewReadiness(t *testing.T, harness *appsupport.ServerHarness, manager *jobs.Manager, transactions *jobs.TransactionService, runner *jobs.Runner, actor string, create func(string) map[string]any) {
	t.Helper()
	ctx := context.Background()
	closeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := runner.Close(closeCtx); err != nil {
		t.Fatal(err)
	}
	references := reportingReadinessReferences(t, harness)
	ready := func() error { return reporting.ValidateRetainedPreviews(ctx, harness.Pool, references) }
	finalizer := reportingFixtureFinalizer(t, harness, transactions)
	for _, status := range []string{jobs.StatusQueued, jobs.StatusRunning, jobs.StatusFailed, jobs.StatusCanceled} {
		admitted := create("txn-preview-retained-" + status)
		id := uuid.MustParse(admitted["render_attempt_id"].(string))
		if status != jobs.StatusQueued {
			execution, claimed, err := manager.Claim(ctx, id)
			if err != nil || !claimed {
				t.Fatal("claim pending preview", claimed, err)
			}
			reason, failureCode, stage := "invalid_type", "export_model_invalid", "export_model_materialization"
			if status == jobs.StatusCanceled {
				reason, failureCode, stage = "render_canceled", "render_canceled", "release_state"
			}
			mutate := func(ctx context.Context, tx pgx.Tx) error {
				_, err := (&reporting.Store{}).CompleteRenderFailureJobTx(ctx, tx, id, reporting.RenderFailure{Stage: stage, FailureCode: failureCode, ReasonCode: reason, SafeDetails: map[string]any{}}, time.Now().UTC())
				return err
			}
			switch status {
			case jobs.StatusFailed:
				_, err = finalizer.FinalizeReportingJobFailure(ctx, reporting.JobFailureFinalization{Execution: execution, Mutate: mutate, Completion: jobs.FailureCompletion{Progress: jobs.Progress{Completed: 0}, ErrorSummary: jobs.ErrorSummary{Code: "release_render_failed", Message: "Release render failed.", Details: map[string]any{"reason_code": reason}}}})
			case jobs.StatusCanceled:
				_, err = manager.Cancel(ctx, jobs.CancelParams{JobID: id, ActorUserID: uuid.MustParse(actor), ClientTxnID: "cancel-retained-preview", NormalizedRequest: []byte(`{"client_txn_id":"cancel-retained-preview","reason":null}`)})
				if err == nil {
					_, err = finalizer.FinalizeReportingJobCancellation(ctx, reporting.JobCancellationFinalization{Execution: execution, Mutate: mutate, Completion: jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 0}}})
				}
			}
			if err != nil {
				t.Fatal(status, err)
			}
		}
		actual, err := manager.Get(ctx, id)
		if err != nil || actual.Status != status {
			t.Fatal("retained state", actual.Status, status, err)
		}
		if err := ready(); err != nil {
			t.Fatal("legitimate unfinished preview rejected", status, err)
		}
		var raw []byte
		if err := harness.DB.QueryRow(`SELECT request_json FROM reporting_job_payloads WHERE job_id=$1`, id).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET request_json=request_json || '{"unadmitted":true}'::jsonb WHERE job_id=$1`, id); err != nil {
			t.Fatal(err)
		}
		rejected := ready()
		if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET request_json=$2 WHERE job_id=$1`, id, raw); err != nil {
			t.Fatal(err)
		}
		if rejected == nil {
			t.Fatal("unfinished preview corruption accepted", status)
		}
		var outputs int
		if err := harness.DB.QueryRow(`SELECT count(*) FROM reporting_composition_preview_outputs WHERE render_attempt_id=$1`, id).Scan(&outputs); err != nil || outputs != 0 {
			t.Fatal("unfinished preview manufactured output", outputs, err)
		}
		if status == jobs.StatusFailed || status == jobs.StatusCanceled {
			assertRetainedFailureIntegrity(t, harness, id, ready, nil)
			compactNativeReleaseTestJob(t, harness, id)
			if err := ready(); err != nil {
				t.Fatal("expired unfinished preview rejected", status, err)
			}
			assertRetainedFailureIntegrity(t, harness, id, ready, nil)
		}
	}
}

func reportingFixtureFinalizer(t *testing.T, harness *appsupport.ServerHarness, transactions *jobs.TransactionService) reporting.JobSuccessFinalizer {
	t.Helper()
	state, err := extensionstore.New(harness.Pool, nil)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := extensionstore.NewOwnerFinalizer(state, transactions, extensionassembly.ReportingReceiptReconciler{}, time.Now, func(err error) { t.Errorf("indeterminate fixture finalization: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	return extensionassembly.NewReportingJobSuccessFinalizer(owner)
}
