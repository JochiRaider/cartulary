package graphprojection

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

const ProjectionSchemaIDV2 = "graph_projection.v2"

const (
	MaximumResultVerticesV2 = 100000
	MaximumResultEdgesV2    = 250000
)

var (
	ErrResultV2NotFound         = errors.New("graph projection result not found")
	ErrResultV2BindingMismatch  = errors.New("graph projection result binding mismatch")
	ErrResultV2IdentityConflict = errors.New("graph projection result identity conflict")
	ErrResultV2Invalid          = errors.New("graph projection result invalid")
	ErrResultV2LeaseNotFound    = errors.New("graph projection result lease not found")
	ErrResultV2NotSelected      = errors.New("graph projection result is not selected")
	ErrResultV2SourceStale      = errors.New("graph projection result source is stale")
)

// ResultBindingV2 is the complete immutable identity tuple consumers must
// prove. Operational timestamps and leases are deliberately absent.
type ResultBindingV2 struct {
	ProjectionResultID            string
	GraphViewID                   string
	SourceOwnerID                 string
	SourceSnapshotID              string
	ProjectionSchemaID            string
	ProjectionVersion             string
	NormalizedConfigurationSHA256 string
	NormalizedSourceSHA256        string
	CanonicalOutputSHA256         string
}

type ResultVertexV2 struct {
	VertexID   string
	VertexKind string
	SortKey    string
	JSON       json.RawMessage
}

type ResultEdgeV2 struct {
	EdgeID      string
	EdgeKind    string
	SrcVertexID string
	DstVertexID string
	Direction   string
	SortKey     string
	JSON        json.RawMessage
}

// CompletedResultV2 is immutable semantic output plus the server-owned
// publication time. PublishedAt is storage metadata and never enters identity.
type CompletedResultV2 struct {
	Binding     ResultBindingV2
	ResultJSON  json.RawMessage
	Vertices    []ResultVertexV2
	Edges       []ResultEdgeV2
	PublishedAt time.Time
}

type ResultLeaseV2 struct {
	LeaseID              string
	ProjectionResultID   string
	LeaseOwnerID         string
	LeaseOwnerResourceID string
	LeasePurpose         string
	LeasedUntil          time.Time
	CreatedAt            time.Time
	RenewedAt            time.Time
}

type TraversalDirectionV2 string

const (
	TraversalOutgoingV2 TraversalDirectionV2 = "outgoing"
	TraversalIncomingV2 TraversalDirectionV2 = "incoming"
	TraversalBothV2     TraversalDirectionV2 = "both"
)

type TraversalRequestV2 struct {
	ProjectionResultID string
	SeedVertexIDs      []string
	Direction          TraversalDirectionV2
	MaximumDepth       int
	MaximumVertices    int
	MaximumEdges       int
	VertexKinds        []string
	EdgeKinds          []string
}

type TraversalResultV2 struct {
	Vertices []ResultVertexV2
	Edges    []ResultEdgeV2
}

// ResultPublisherV2 is intentionally transaction-agnostic. A PostgreSQL
// adapter is constructed with the caller's borrowed transaction.
type ResultPublisherV2 interface {
	PublishResult(context.Context, CompletedResultV2) error
}

type ExactResultReaderV2 interface {
	ReadExactResult(context.Context, ResultBindingV2) (CompletedResultV2, error)
	ReadVertices(context.Context, string, int) ([]ResultVertexV2, error)
	ReadEdges(context.Context, string, int) ([]ResultEdgeV2, error)
	Traverse(context.Context, TraversalRequestV2) (TraversalResultV2, error)
}

type ResultLeaseWriterV2 interface {
	AcquireLease(context.Context, ResultLeaseV2) (ResultLeaseV2, error)
	RenewLease(context.Context, string, time.Time, time.Time) (ResultLeaseV2, error)
	ReleaseLease(context.Context, string) error
}

type ResultCleanupCandidateV2 struct {
	ProjectionResultID string
	PublishedAt        time.Time
}

// ResultMaintenanceV2 exposes only borrowed-transaction maintenance
// primitives. Source owners remain responsible for checking their own
// authoritative declarations between candidate locking and deletion.
type ResultMaintenanceV2 interface {
	DeleteExpiredLeases(context.Context, time.Time, int) (int, bool, error)
	LockCleanupCandidate(context.Context, string, *ResultCleanupCandidateV2) (*ResultCleanupCandidateV2, error)
	HasUnexpiredLease(context.Context, string, time.Time) (bool, error)
	DeleteLockedResult(context.Context, string) (bool, error)
}
