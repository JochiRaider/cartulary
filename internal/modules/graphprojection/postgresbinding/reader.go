package postgresbinding

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

type Reader struct {
	tx pgx.Tx
}

func NewReader(tx pgx.Tx) *Reader {
	return &Reader{tx: tx}
}

func (r *Reader) LookupProjectionBinding(ctx context.Context, projectionRunID string) (graphprojection.ProjectionBinding, error) {
	var binding graphprojection.ProjectionBinding
	var state string
	var outputDigest pgtype.Text
	err := r.tx.QueryRow(ctx, `
SELECT projection_run_id, graph_view_id, source_snapshot_id, projection_version, state,
       projection_config_digest, projection_source_digest, projection_output_digest
  FROM graph_projection_runs
 WHERE projection_run_id = $1
`, projectionRunID).Scan(
		&binding.ProjectionRunID,
		&binding.GraphViewID,
		&binding.SourceSnapshotID,
		&binding.ProjectionVersion,
		&state,
		&binding.ProjectionConfigDigest,
		&binding.ProjectionSourceDigest,
		&outputDigest,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return graphprojection.ProjectionBinding{}, graphprojection.ErrProjectionRunNotFound
	}
	if err != nil {
		return graphprojection.ProjectionBinding{}, err
	}
	binding.State = graphprojection.RunState(state)
	if outputDigest.Valid {
		binding.ProjectionOutputDigest = outputDigest.String
	}
	return binding, nil
}
