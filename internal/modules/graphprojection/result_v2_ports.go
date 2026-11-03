package graphprojection

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/internal/semanticlimits"
)

const ProjectionSchemaIDV2 = "graph_projection.v2"

const (
	MaximumResultVerticesV2 = semanticlimits.MaximumResultVerticesV2
	MaximumResultEdgesV2    = semanticlimits.MaximumResultEdgesV2
)

var (
	ErrResultV2NotFound         = errors.New("graph projection result not found")
	ErrResultV2BindingMismatch  = errors.New("graph projection result binding mismatch")
	ErrResultV2IdentityConflict = errors.New("graph projection result identity conflict")
	ErrResultV2Invalid          = errors.New("graph projection result invalid")
	ErrResultV2NotSelected      = errors.New("graph projection result is not selected")
	ErrResultV2SourceStale      = errors.New("graph projection result source is stale")
)

// ResultBindingV2 is the complete immutable identity tuple consumers must
// prove. Operational timestamps are deliberately absent.
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

// ValidateIdentity checks the canonical identifier of the supplied immutable
// tuple without depending on derived rows or operational timestamps.
func (binding ResultBindingV2) ValidateIdentity() error {
	for _, identifier := range []string{binding.GraphViewID, binding.SourceOwnerID, binding.SourceSnapshotID, binding.ProjectionVersion} {
		if !validIdentifierV2(identifier) {
			return ErrResultV2Invalid
		}
	}
	for _, digest := range []string{binding.NormalizedConfigurationSHA256, binding.NormalizedSourceSHA256, binding.CanonicalOutputSHA256} {
		if !validResultSHA256V2(digest) {
			return ErrResultV2Invalid
		}
	}
	expected, err := deriveProjectionResultIDV2(binding)
	if err != nil || binding.ProjectionSchemaID != ProjectionSchemaIDV2 || binding.ProjectionResultID != expected {
		return ErrResultV2Invalid
	}
	return nil
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

type ResultCleanupCandidateV2 struct {
	ProjectionResultID string
	PublishedAt        time.Time
}
