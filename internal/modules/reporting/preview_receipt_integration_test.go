package reporting_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func assertPreviewReceiptIntegrity(t *testing.T, harness *appsupport.ServerHarness, jobText, clientTxn string) {
	t.Helper()
	ctx := context.Background()
	tx, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	source, err := reportcomposition.PreviewSourceForRenderTx(ctx, tx, uuid.MustParse(jobText))
	if err != nil {
		t.Fatal(err)
	}
	key := jobs.NewRouteIdempotencyKey("report_compositions.preview", source.CreatedByUserID, source.CompositionID.String(), clientTxn)
	actor := source.CreatedByUserID.String()
	job := jobs.Resource{JobID: jobText, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &source.IncidentID}, SubmittedByUserID: &actor}
	var original, digest []byte
	if err := tx.QueryRow(ctx, `SELECT response_json,request_hash FROM route_idempotency WHERE route_key=$1 AND actor_user_id=$2 AND scope_key=$3 AND client_txn_id=$4`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID).Scan(&original, &digest); err != nil {
		t.Fatal(err)
	}
	reconcile := func() (bool, error) {
		return (reportcomposition.PreviewReceiptReconciler{}).ReconcileFinalIdempotencyOutcomeTx(ctx, tx, key, digest, job)
	}
	if ok, err := reconcile(); !ok || err != nil {
		t.Fatal("original preview receipt rejected", err)
	}
	for _, mutate := range []func(map[string]any){
		func(v map[string]any) { v["render_attempt_id"] = uuid.NewString() },
		func(v map[string]any) { delete(v, "composition_sha256") },
		func(v map[string]any) { v["snapshot_id"] = "replacement" },
		func(v map[string]any) { v["job_id"] = jobText },
		func(v map[string]any) { v["output_options"] = map[string]any{} },
	} {
		var payload map[string]any
		if err := json.Unmarshal(original, &payload); err != nil {
			t.Fatal(err)
		}
		mutate(payload)
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE route_idempotency SET response_json=$5 WHERE route_key=$1 AND actor_user_id=$2 AND scope_key=$3 AND client_txn_id=$4`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID, raw); err != nil {
			t.Fatal(err)
		}
		if ok, err := reconcile(); ok || err == nil {
			t.Fatal("altered preview receipt accepted")
		}
	}
}
