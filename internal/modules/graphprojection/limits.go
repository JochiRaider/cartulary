package graphprojection

// resourceLimits is the single implementation registry for the public limits
// in Graph Projection NLSpec section 4.12. These values are conformance
// constants, not deployment configuration.
type resourceLimits struct {
	MaxInputBytes                      int
	MaxSourceEntities                  int
	MaxSourceRelationships             int
	MaxEntityFilters                   int
	MaxRelationshipFilters             int
	MaxDeclaredSourceEntityKinds       int
	MaxDeclaredSourceRelationshipKinds int
	MaxEntityMappings                  int
	MaxRelationshipMappings            int
	MaxPropertyDefinitions             int
	MaxMetadataMappings                int
	MaxAggregationRules                int
	MaxDefaultVertexLabels             int
	MaxDefaultEdgeLabels               int
	MaxMappingLabelsPerRule            int
	MaxMappingPropertyKeyRefs          int
	MaxLabelsPerSourceItem             int
	MaxLabelLength                     int
	MaxStringPropertyValueLength       int
	MaxMetadataKeysPerObject           int
	MaxPropertiesPerObject             int
	MaxValidationIssues                int
	MaxValidationMessageLength         int
	MaxFailureReasonLength             int
	MaxCursorTokenLength               int
	MaxProjectedVertices               int
	MaxProjectedEdges                  int
	MaxTraversalSeedVertices           int
	MaxTraversalKindFilters            int
	MaxTraversalDepth                  int
	MaxListGraphViewsLimit             int
}

var graphProjectionLimits = resourceLimits{
	MaxInputBytes:                      268435456,
	MaxSourceEntities:                  100000,
	MaxSourceRelationships:             250000,
	MaxEntityFilters:                   1000,
	MaxRelationshipFilters:             1000,
	MaxDeclaredSourceEntityKinds:       10000,
	MaxDeclaredSourceRelationshipKinds: 10000,
	MaxEntityMappings:                  10000,
	MaxRelationshipMappings:            10000,
	MaxPropertyDefinitions:             10000,
	MaxMetadataMappings:                10000,
	MaxAggregationRules:                1000,
	MaxDefaultVertexLabels:             256,
	MaxDefaultEdgeLabels:               256,
	MaxMappingLabelsPerRule:            256,
	MaxMappingPropertyKeyRefs:          1024,
	MaxLabelsPerSourceItem:             256,
	MaxLabelLength:                     256,
	MaxStringPropertyValueLength:       16384,
	MaxMetadataKeysPerObject:           1024,
	MaxPropertiesPerObject:             1024,
	MaxValidationIssues:                100000,
	MaxValidationMessageLength:         1024,
	MaxFailureReasonLength:             4096,
	MaxCursorTokenLength:               4096,
	MaxProjectedVertices:               500000,
	MaxProjectedEdges:                  1000000,
	MaxTraversalSeedVertices:           1024,
	MaxTraversalKindFilters:            1024,
	MaxTraversalDepth:                  16,
	MaxListGraphViewsLimit:             1000,
}

// ResourceLimits exposes the closed NLSpec §4.12 registry to Graph Projection
// adapters without letting callers configure conformance behavior.
func ResourceLimits() resourceLimits {
	return graphProjectionLimits
}
