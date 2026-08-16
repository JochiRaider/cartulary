package graphprojection

type projectionRequest struct {
	ProjectionSchemaID   string
	SourceSnapshotID     string
	projectionConfig     projectionConfig
	SourceEntities       []sourceEntity
	SourceRelationships  []sourceRelationship
	SourceMetadata       map[string]any
	filters              filters
	RelationshipMappings []relationshipMapping
	PropertyDefinitions  []propertyDefinition
}

type projectionConfig struct {
	GraphViewKey                    string
	ProjectionVersion               string
	DeclaredSourceEntityKinds       []string
	DeclaredSourceRelationshipKinds []string
	EntityMappings                  []entityMapping
	RelationshipMappings            []relationshipMapping
	MetadataMappings                []metadataMapping
	AggregationRules                []aggregationRule
	DefaultVertexLabels             []string
	DefaultEdgeLabels               []string
	AllowEmptyKindRegistry          bool
}

type entityMapping struct {
	MappingRuleID         string
	SourceEntityKind      string
	ProjectedVertexKind   string
	InclusionPredicate    string
	InclusionFilter       *filterPredicate
	LabelPolicy           string
	MappingLabels         []string
	RequiredPropertyKeys  []string
	OptionalPropertyKeys  []string
	MappingIdentityDigest string
}

type relationshipMapping struct {
	MappingRuleID           string
	SourceRelationshipKind  string
	ProjectedEdgeKind       string
	InclusionPredicate      string
	InclusionFilter         *filterPredicate
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

type metadataMapping struct {
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

type aggregationRule struct {
	AggregationRuleID          string
	TargetScope                string
	InputScope                 string
	InputKind                  string
	ProjectedKind              string
	GroupingKeys               []string
	MissingGroupingKeyBehavior string
	PropertyMergeBehavior      map[string]string
	EdgeDirection              string
	endpointGrouping           *endpointGrouping
	AggregationIdentityDigest  string
}

type endpointGrouping struct {
	SourceVertexAggregationRuleID      string
	SourceGroupingKeys                 []string
	DestinationVertexAggregationRuleID string
	DestinationGroupingKeys            []string
	MissingEndpointBehavior            string
}

type filters struct {
	EntityFilters       []filterPredicate
	RelationshipFilters []filterPredicate
	Logic               string
}

type filterPredicate struct {
	FieldPath        string
	Operator         string
	Value            any
	HasValue         bool
	IncludeIfMissing bool
}

type sourceEntity struct {
	SourceEntityID   string
	SourceEntityKind string
	Properties       map[string]any
	Metadata         map[string]any
	Labels           []string
}

type sourceRelationship struct {
	SourceRelationshipID   string
	SourceRelationshipKind string
	SrcSourceEntityID      string
	DstSourceEntityID      string
	Direction              string
	Properties             map[string]any
	Metadata               map[string]any
	Labels                 []string
}

type propertyDefinition struct {
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
