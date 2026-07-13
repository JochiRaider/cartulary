package graphprojection

import (
	"errors"
	"time"
)

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
	GraphViewID            string
	GraphViewKey           string
	State                  GraphViewState
	LatestProjectionRunID  string
	LatestSourceSnapshotID string
	ProjectionVersion      string
	UpdatedAt              time.Time
	ValidationStatus       string
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
