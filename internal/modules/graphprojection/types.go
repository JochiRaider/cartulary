package graphprojection

import "time"

const ProjectionSchemaID = "graph_projection.v1"

type Optional[T any] struct {
	Present bool
	Null    bool
	Value   T
}

func ValueOf[T any](value T) Optional[T] { return Optional[T]{Present: true, Value: value} }
func ExplicitNull[T any]() Optional[T]   { return Optional[T]{Present: true, Null: true} }

type RunState string

const (
	RunStateAccepted    RunState = "accepted"
	RunStateComputing   RunState = "computing"
	RunStateAvailable   RunState = "available"
	RunStateFailed      RunState = "failed"
	RunStateReplaced    RunState = "replaced"
	RunStateInvalidated RunState = "invalidated"
)

type GraphViewState string

const (
	GraphViewStateCreating    GraphViewState = "creating"
	GraphViewStateAvailable   GraphViewState = "available"
	GraphViewStateRefreshing  GraphViewState = "refreshing"
	GraphViewStateFailed      GraphViewState = "failed"
	GraphViewStateInvalidated GraphViewState = "invalidated"
)

type ProjectionRequest struct {
	ProjectionSchemaID                string
	GraphViewID                       string
	SourceSnapshotID                  string
	ProjectionConfig                  ProjectionConfig
	SourceEntities                    []SourceEntity
	SourceRelationships               []SourceRelationship
	SourceMetadata                    map[string]any
	Filters                           Filters
	RelationshipMappings              []RelationshipMapping
	PropertyDefinitions               []PropertyDefinition
	RequestedAt                       string
	RequestedBy                       string
	Normalized                        map[string]any
	RelationshipMappingSourceConflict bool
}

type ProjectionConfig struct {
	GraphViewKey                    string
	ProjectionVersion               string
	DeclaredSourceEntityKinds       []string
	DeclaredSourceRelationshipKinds []string
	EntityMappings                  []EntityMapping
	RelationshipMappings            []RelationshipMapping
	MetadataMappings                []MetadataMapping
	AggregationRules                []AggregationRule
	DefaultVertexLabels             []string
	DefaultEdgeLabels               []string
	AllowEmptyKindRegistry          bool
	RetentionPolicy                 RetentionPolicy
	CustomConfig                    map[string]any
}

type RetentionPolicy struct {
	RetainReplacedResults       bool              `json:"retain_replaced_results"`
	RetentionCount              int               `json:"retention_count"`
	RetentionDurationSeconds    int               `json:"retention_duration_seconds"`
	RetainFailedResults         bool              `json:"retain_failed_results"`
	FailedRetentionCount        int               `json:"failed_retention_count"`
	FailedRetentionDurationSecs int               `json:"failed_retention_duration_seconds"`
	RawIntegerLexemes           map[string]string `json:"-"`
}

type EntityMapping struct {
	MappingRuleID         string
	SourceEntityKind      string
	ProjectedVertexKind   string
	InclusionPredicate    string
	InclusionFilter       *FilterPredicate
	LabelPolicy           string
	MappingLabels         []string
	RequiredPropertyKeys  []string
	OptionalPropertyKeys  []string
	MappingIdentityDigest string
}

type RelationshipMapping struct {
	MappingRuleID           string
	SourceRelationshipKind  string
	ProjectedEdgeKind       string
	InclusionPredicate      string
	InclusionFilter         *FilterPredicate
	DirectionPolicy         string
	EmitReverseEdge         bool
	ReverseEdgeKind         string
	ReverseEdgeKindSupplied bool
	LabelPolicy             string
	MappingLabels           []string
	RequiredPropertyKeys    []string
	OptionalPropertyKeys    []string
	MappingIdentityDigest   string
}

type MetadataMapping struct {
	MetadataMappingID    string
	TargetScope          string
	TargetKind           string
	SourceFieldPath      string
	ProjectedMetadataKey string
	ProjectedType        string
	Required             bool
	DefaultValue         any
	HasDefaultValue      bool
	MissingBehavior      string
	SourceNullBehavior   string
	NullOutputPolicy     string
	MergeBehavior        string
}

type AggregationRule struct {
	AggregationRuleID          string
	TargetScope                string
	InputScope                 string
	InputKind                  string
	ProjectedKind              string
	GroupingKeys               []string
	MissingGroupingKeyBehavior string
	PropertyMergeBehavior      map[string]string
	EdgeDirection              string
	EndpointGrouping           *EndpointGrouping
	AggregationIdentityDigest  string
}

type EndpointGrouping struct {
	SourceVertexAggregationRuleID      string
	SourceGroupingKeys                 []string
	DestinationVertexAggregationRuleID string
	DestinationGroupingKeys            []string
	MissingEndpointBehavior            string
}

type Filters struct {
	EntityFilters       []FilterPredicate
	RelationshipFilters []FilterPredicate
	Logic               string
}

type FilterPredicate struct {
	FieldPath        string
	Operator         string
	Value            any
	HasValue         bool
	IncludeIfMissing bool
}

type SourceEntity struct {
	SourceEntityID   string
	SourceEntityKind string
	Properties       map[string]any
	Metadata         map[string]any
	Labels           []string
}

type SourceRelationship struct {
	SourceRelationshipID   string
	SourceRelationshipKind string
	SrcSourceEntityID      string
	DstSourceEntityID      string
	Direction              string
	Properties             map[string]any
	Metadata               map[string]any
	Labels                 []string
}

type PropertyDefinition struct {
	PropertyDefinitionID string
	TargetScope          string
	TargetKind           string
	SourceFieldPath      string
	ProjectedKey         string
	ProjectedType        string
	Required             bool
	DefaultValue         any
	HasDefaultValue      bool
	MissingBehavior      string
	SourceNullBehavior   string
	NullOutputPolicy     string
	MergeBehavior        string
}

type ProjectionRun struct {
	Request                 ProjectionRequest
	GraphViewID             string
	ProjectionRunID         string
	ProjectionRunNonce      string
	ProjectionConfigDigest  string
	ProjectionSourceDigest  string
	ProjectionOutputDigest  string
	AcceptedAt              time.Time
	StartedAt               *time.Time
	GeneratedAt             *time.Time
	CompletedAt             *time.Time
	ReplacedAt              *time.Time
	InvalidatedAt           *time.Time
	Invalidation            *Invalidation
	AcceptedReplay          *AcceptedRunSummary
	State                   RunState
	GraphView               *GraphView
	ValidationSummary       ValidationSummary
	FailureReason           string
	PreviousProjectionRunID *string
	RetentionExpiresAt      *time.Time
}

type GraphView struct {
	ProjectionSchemaID   string
	GraphViewID          string
	GraphViewKey         string
	ProjectionRunID      string
	SourceSnapshotID     string
	ProjectionVersion    string
	GeneratedAt          string
	State                RunState
	Properties           map[string]any
	Metadata             GraphMetadata
	SchemaRegistry       SchemaRegistry
	Vertices             []Vertex
	Edges                []Edge
	ValidationSummary    ValidationSummary
	ConsumerCapabilities ConsumerCapabilities
}

type GraphMetadata struct {
	ProjectionConfigDigest  string
	ProjectionSourceDigest  string
	PreviousProjectionRunID *string
	Invalidation            *Invalidation
	MappedMetadata          map[string]any
}

type SchemaRegistry struct {
	VertexKinds  []VertexKindSchema
	EdgeKinds    []EdgeKindSchema
	PropertyKeys []PropertySchema
	MetadataKeys []MetadataSchema
}

type VertexKindSchema struct {
	VertexKind            string
	SourceEntityKinds     []string
	AggregationRuleIDs    []string
	Labels                []string
	SourceLabelsPreserved bool
	Properties            []PropertySchemaReference
}

type EdgeKindSchema struct {
	EdgeKind                string
	SourceRelationshipKinds []string
	AggregationRuleIDs      []string
	Directions              []string
	Labels                  []string
	SourceLabelsPreserved   bool
	Properties              []PropertySchemaReference
}

type PropertySchemaReference struct {
	ProjectedKey   string
	ProjectedType  string
	Required       bool
	NullableOutput bool
}

type PropertySchema struct {
	TargetScope        string
	TargetKind         string
	ProjectedKey       string
	ProjectedType      string
	Required           bool
	NullableOutput     bool
	MissingBehavior    string
	SourceNullBehavior string
}

type MetadataSchema struct {
	TargetScope          string
	TargetKind           string
	ProjectedMetadataKey string
	ProjectedType        string
	Required             bool
	NullableOutput       bool
	MissingBehavior      string
	SourceNullBehavior   string
}

type Vertex struct {
	VertexID        string
	VertexKind      string
	VertexFamily    string
	Labels          []string
	Properties      map[string]any
	Metadata        VertexMetadata
	SourceEntityRef *SourceEntityRef
	SortKey         string
}

type VertexMetadata struct {
	MappingRuleID         *string
	AggregationRuleID     *string
	AggregationSourceRefs []SourceRef
	MappedMetadata        map[string]any
}

type Edge struct {
	EdgeID                string
	EdgeKind              string
	EdgeFamily            string
	SrcVertexID           string
	DstVertexID           string
	Direction             string
	Labels                []string
	Properties            map[string]any
	Metadata              EdgeMetadata
	SourceRelationshipRef *SourceRelationshipRef
	SortKey               string
}

type EdgeMetadata struct {
	MappingRuleID         *string
	AggregationRuleID     *string
	IsReverseEdge         bool
	ReverseOfEdgeID       *string
	AggregationSourceRefs []SourceRef
	MappedMetadata        map[string]any
}

type SourceEntityRef struct {
	SourceEntityID   string
	SourceEntityKind string
	MappingRuleID    string
}

type SourceRelationshipRef struct {
	SourceRelationshipID   string
	SourceRelationshipKind string
	MappingRuleID          string
}

type SourceRef struct {
	RefKind            string
	RefID              string
	RefKindName        string
	ContributorSortKey string
}

type ValidationSummary struct {
	Status     string
	IssueCount int
	Issues     []ValidationIssue
}

type ValidationIssue struct {
	IssueID    string
	Severity   string
	Code       string
	TargetKind string
	TargetID   string
	Field      *string
	Message    string
	Details    map[string]any
}

type ConsumerCapabilities struct {
	QueryShapes                     []string
	SupportsDirectVertexLookup      bool
	SupportsDirectEdgeLookup        bool
	SupportsBreadthFirstTraversal   bool
	SupportsAlternateTraversalOrder []string
	MaxTraversalDepth               int
	MaxTraversalSeedVertices        int
	MaxKindFilters                  int
}

type InvalidationSummary struct {
	GraphViewID           string         `json:"graph_view_id"`
	TargetScope           string         `json:"target_scope"`
	TargetProjectionRunID *string        `json:"target_projection_run_id"`
	InvalidatedRunIDs     []string       `json:"invalidated_projection_run_ids"`
	GraphViewStateAfter   GraphViewState `json:"graph_view_state_after"`
	InvalidatedAt         string         `json:"invalidated_at"`
	ReasonCode            string         `json:"reason_code"`
	RequestedAt           string         `json:"requested_at,omitempty"`
	RequestedBy           string         `json:"requested_by"`
	IdempotencyExpiresAt  *string        `json:"idempotency_expires_at"`
}

type Invalidation struct {
	InvalidatedAt         string  `json:"invalidated_at"`
	ReasonCode            string  `json:"reason_code"`
	RequestedBy           string  `json:"requested_by"`
	TargetScope           string  `json:"target_scope"`
	TargetProjectionRunID *string `json:"target_projection_run_id"`
}
