package graphprojection

import "errors"

var (
	ErrProjectionRunNotFound = errors.New("graphprojection: projection run not found")
	ErrGraphViewNotFound     = errors.New("graphprojection: graph view not found")
	ErrCursorInvalid         = errors.New("graphprojection: cursor invalid")
	ErrGraphViewUnavailable  = errors.New("graphprojection: graph view unavailable")
	ErrVertexNotFound        = errors.New("graphprojection: vertex not found")
	ErrEdgeNotFound          = errors.New("graphprojection: edge not found")
)

type GetGraphViewRequest struct {
	GraphViewID     string
	ProjectionRunID Optional[string]
}

type GetProjectionRunRequest struct {
	GraphViewID     string
	ProjectionRunID string
}

type GetVertexRequest struct {
	GraphViewID     string
	ProjectionRunID Optional[string]
	VertexID        string
}

type GetEdgeRequest struct {
	GraphViewID     string
	ProjectionRunID Optional[string]
	EdgeID          string
}

type ListGraphViewsOptions struct {
	Limit                 *int
	CursorToken           string
	QueryShapeDigest      string
	VisibilityScopeDigest string
}

type GraphViewSummary struct {
	GraphViewID            string         `json:"graph_view_id"`
	GraphViewKey           string         `json:"graph_view_key"`
	State                  GraphViewState `json:"state"`
	LatestProjectionRunID  string         `json:"latest_projection_run_id"`
	LatestSourceSnapshotID string         `json:"latest_source_snapshot_id"`
	ProjectionVersion      string         `json:"projection_version"`
	UpdatedAt              string         `json:"updated_at"`
	ValidationStatus       string         `json:"validation_status"`
}

type ListGraphViewsResult struct {
	GraphViews      []GraphViewSummary `json:"graph_views"`
	NextCursorToken *string            `json:"next_cursor_token"`
}

type ProjectionRunInspection struct {
	GraphViewID            string             `json:"graph_view_id"`
	ProjectionRunID        string             `json:"projection_run_id"`
	SourceSnapshotID       string             `json:"source_snapshot_id"`
	ProjectionVersion      string             `json:"projection_version"`
	State                  RunState           `json:"state"`
	StartedAt              *string            `json:"started_at"`
	CompletedAt            *string            `json:"completed_at"`
	ValidationSummary      *ValidationSummary `json:"validation_summary"`
	FailureReason          *string            `json:"failure_reason"`
	HasConsumableGraphView bool               `json:"has_consumable_graph_view"`
	Invalidation           *Invalidation      `json:"invalidation"`
	RetentionExpiresAt     *string            `json:"retention_expires_at"`
}

type TraverseRequest struct {
	GraphViewID     string
	ProjectionRunID string
	SeedVertexIDs   []string
	MaxDepth        *int
	Direction       string
	VertexKinds     []string
	EdgeKinds       []string
}

type TraverseResult struct {
	GraphViewID          string
	ProjectionRunID      string
	SeedVertexIDs        []string
	OmittedSeedVertexIDs []string
	Vertices             []Vertex
	Edges                []Edge
	Metadata             map[string]any
}
