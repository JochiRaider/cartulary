package graphprojection

import (
	"context"
	"time"
)

// RetainedProjectionOptions are server-assigned values for a retained
// projection invocation. Callers use Service; adapters receive this value only
// after the service has established the operation boundary.
type RetainedProjectionOptions struct {
	ProjectionRunNonce string
	AcceptedAt         time.Time
	GeneratedAt        time.Time
	IdempotencyKey     string
}

// RetainedInvalidation is the persistence-facing form of an already validated
// invalidation request. It intentionally contains no transport or database
// transaction types.
type RetainedInvalidation struct {
	GraphViewID     string
	ProjectionRunID string
	ReasonCode      string
	RequestedAt     string
	RequestedBy     string
	IdempotencyKey  string
	InvalidatedAt   time.Time
}

// RetainedRepository is the Graph Projection persistence port. Implementations
// own graph-table mapping and transactions; Service remains the application
// facade consumed by sibling modules.
type RetainedRepository interface {
	CreateProjection(context.Context, []byte, RetainedProjectionOptions) (ProjectionRun, error)
	RefreshProjection(context.Context, []byte, RetainedProjectionOptions) (ProjectionRun, error)
	GetProjectionRun(context.Context, string) (ProjectionRun, error)
	GetGraphView(context.Context, string, string) (GraphView, error)
	GetVertex(context.Context, string, string, string) (Vertex, error)
	GetEdge(context.Context, string, string, string) (Edge, error)
	ListGraphViews(context.Context, ListGraphViewsOptions) ([]GraphViewSummary, string, error)
	Traverse(context.Context, TraverseRequest) (TraverseResult, error)
	InvalidateGraphView(context.Context, RetainedInvalidation) (InvalidationSummary, error)
	InvalidateProjectionRun(context.Context, RetainedInvalidation) (InvalidationSummary, error)
}
