package graphprojection

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	ReportingProjectionReasonNotBound       = "graph_projection_not_bound"
	ReportingProjectionReasonNotCompleted   = "graph_projection_not_completed"
	ReportingProjectionReasonStale          = "graph_projection_stale"
	ReportingProjectionReasonDigestMismatch = "graph_projection_digest_mismatch"
)

type ReportingProjectionRef struct {
	ProjectionSchemaID     string
	GraphViewID            string
	SourceSnapshotID       string
	ProjectionRunID        string
	ProjectionVersion      string
	ProjectionConfigDigest string
	ProjectionSourceDigest string
	ProjectionOutputDigest string
}

type ReportingProjectionRefError struct {
	Field      string
	ReasonCode string
}

func (e *ReportingProjectionRefError) Error() string {
	if e == nil || e.ReasonCode == "" {
		return "graphprojection: invalid reporting projection ref"
	}
	return "graphprojection: invalid reporting projection ref: " + e.ReasonCode
}

func ValidateReportingProjectionRefsTx(ctx context.Context, tx pgx.Tx, expectedSnapshotID string, refs []ReportingProjectionRef) error {
	for _, ref := range refs {
		if ref.ProjectionSchemaID != ProjectionSchemaID {
			return &ReportingProjectionRefError{Field: "graph_projection_refs", ReasonCode: ReportingProjectionReasonDigestMismatch}
		}
		var graphViewID string
		var sourceSnapshotID string
		var projectionVersion string
		var state string
		var projectionConfigDigest string
		var projectionSourceDigest string
		var projectionOutputDigest pgtype.Text
		err := tx.QueryRow(ctx, `
SELECT graph_view_id, source_snapshot_id, projection_version, state,
       projection_config_digest, projection_source_digest, projection_output_digest
  FROM graph_projection_runs
 WHERE projection_run_id = $1
`, ref.ProjectionRunID).Scan(&graphViewID, &sourceSnapshotID, &projectionVersion, &state, &projectionConfigDigest, &projectionSourceDigest, &projectionOutputDigest)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return &ReportingProjectionRefError{Field: "graph_projection_refs", ReasonCode: ReportingProjectionReasonNotBound}
			}
			return err
		}
		if state != string(RunStateAvailable) && state != string(RunStateReplaced) {
			return &ReportingProjectionRefError{Field: "graph_projection_refs", ReasonCode: ReportingProjectionReasonNotCompleted}
		}
		if sourceSnapshotID != expectedSnapshotID {
			return &ReportingProjectionRefError{Field: "graph_projection_refs", ReasonCode: ReportingProjectionReasonStale}
		}
		if graphViewID != ref.GraphViewID ||
			projectionVersion != ref.ProjectionVersion ||
			projectionConfigDigest != ref.ProjectionConfigDigest ||
			projectionSourceDigest != ref.ProjectionSourceDigest ||
			!projectionOutputDigest.Valid ||
			projectionOutputDigest.String != ref.ProjectionOutputDigest {
			return &ReportingProjectionRefError{Field: "graph_projection_refs", ReasonCode: ReportingProjectionReasonDigestMismatch}
		}
	}
	return nil
}
