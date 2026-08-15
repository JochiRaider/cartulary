package postgresstore

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

func (s *Store) ListGraphViews(ctx context.Context, options graphprojection.ListGraphViewsOptions) ([]graphprojection.GraphViewSummary, string, error) {
	limit := 100
	if options.Limit != nil {
		limit = *options.Limit
		if limit < 1 || limit > graphprojection.ResourceLimits().MaxListGraphViewsLimit {
			return nil, "", &graphprojection.QueryError{Code: "invalid_argument", ReasonCode: "out_of_bounds", Field: "limit", Details: map[string]any{"field": "limit", "reason_code": "out_of_bounds"}}
		}
	}
	after := ""
	now := s.now().UTC()
	if options.CursorToken != "" {
		if utf8.RuneCountInString(options.CursorToken) > graphprojection.ResourceLimits().MaxCursorTokenLength {
			return nil, "", cursorInvalid("cursor_token_too_long")
		}
		cursor, err := s.cursorCodec.decode(options.CursorToken)
		if err != nil {
			return nil, "", cursorInvalid("malformed")
		}
		if cursor.Operation != "list_graph_views" || cursor.QueryShapeDigest != options.QueryShapeDigest || cursor.VisibilityScopeDigest != options.VisibilityScopeDigest {
			return nil, "", cursorInvalid("wrong_query_shape")
		}
		if !now.Before(cursor.IssuedAt.Add(15 * time.Minute)) {
			return nil, "", cursorInvalid("expired")
		}
		after = cursor.AfterGraphViewID
	}
	rows, err := s.pool.Query(ctx, `
SELECT graph_view_id,
       graph_view_key,
       state,
       COALESCE(latest_projection_run_id, ''),
       COALESCE(latest_source_snapshot_id, ''),
       COALESCE(projection_version, ''),
       updated_at,
       validation_status
  FROM graph_projection_views
 WHERE graph_view_id > $1
 ORDER BY graph_view_id ASC
 LIMIT $2
`, after, limit+1)
	if err != nil {
		return nil, "", fmt.Errorf("list graph views: %w", err)
	}
	defer rows.Close()
	summaries := []graphprojection.GraphViewSummary{}
	for rows.Next() {
		var summary graphprojection.GraphViewSummary
		var state string
		var updatedAt time.Time
		if err := rows.Scan(&summary.GraphViewID, &summary.GraphViewKey, &state, &summary.LatestProjectionRunID, &summary.LatestSourceSnapshotID, &summary.ProjectionVersion, &updatedAt, &summary.ValidationStatus); err != nil {
			return nil, "", fmt.Errorf("scan graph view summary: %w", err)
		}
		summary.State = graphprojection.GraphViewState(state)
		summary.UpdatedAt = graphprojection.FormatLifecycleTimestamp(updatedAt)
		summaries = append(summaries, summary)
	}
	if rows.Err() != nil {
		return nil, "", rows.Err()
	}
	nextCursor := ""
	if len(summaries) > limit {
		summaries = summaries[:limit]
		encoded, err := s.cursorCodec.encode(listCursor{Operation: "list_graph_views", AfterGraphViewID: summaries[len(summaries)-1].GraphViewID, IssuedAt: now, QueryShapeDigest: options.QueryShapeDigest, VisibilityScopeDigest: options.VisibilityScopeDigest})
		if err != nil {
			return nil, "", err
		}
		nextCursor = encoded
	}
	return summaries, nextCursor, nil
}
