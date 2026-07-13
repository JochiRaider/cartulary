package postgresbinding

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestReaderSeesProjectionBindingInsideCallerTransaction(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "graphprojection-binding-reader")
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin caller transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	now := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
	digest := strings.Repeat("a", 64)
	if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_views (
    graph_view_id, graph_view_key, state, latest_projection_run_id,
    latest_source_snapshot_id, projection_version, updated_at, validation_status
) VALUES ('gv_binding', 'binding', 'available', 'gpr_binding', 'snapshot', 'v1', $1, 'valid')
`, now); err != nil {
		t.Fatalf("seed transaction graph view: %v", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_runs (
    projection_run_id, graph_view_id, source_snapshot_id, projection_version,
    state, projection_run_nonce, projection_config_digest, projection_source_digest,
    projection_output_digest, accepted_at, completed_at, validation_summary_json
) VALUES ('gpr_binding', 'gv_binding', 'snapshot', 'v1', 'available', 'nonce', $1, $1, $1, $2, $2, '{}'::jsonb)
`, digest, now); err != nil {
		t.Fatalf("seed transaction projection run: %v", err)
	}
	binding, err := NewReader(tx).LookupProjectionBinding(ctx, "gpr_binding")
	if err != nil {
		t.Fatalf("lookup same-transaction projection binding: %v", err)
	}
	if binding.GraphViewID != "gv_binding" || binding.SourceSnapshotID != "snapshot" || binding.ProjectionOutputDigest != digest {
		t.Fatalf("unexpected projection binding: %#v", binding)
	}
}
