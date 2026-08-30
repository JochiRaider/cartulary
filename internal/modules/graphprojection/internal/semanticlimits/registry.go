// Package semanticlimits owns the private executable projection of Graph
// Projection's owner-authored semantic and traversal limits. It is internal so
// the pure engine and its Graph-owned persistence adapters can share facts
// without enlarging the root API.
package semanticlimits

const (
	MaximumResultVerticesV2 = 100000
	MaximumResultEdgesV2    = 250000
)

type Registry struct {
	MaxInputBytes                  int
	MaxSourceEntities              int
	MaxSourceRelationships         int
	MaxEntityFilters               int
	MaxRelationshipFilters         int
	MaxEntityMappings              int
	MaxRelationshipMappings        int
	MaxPropertyDefinitions         int
	MaxMetadataMappings            int
	MaxAggregationRules            int
	MaxArrayItems                  int
	MaxLabelsPerObject             int
	MaxIdentifierBytes             int
	MaxLabelBytes                  int
	MaxStringBytes                 int
	MaxPropertyKeyBytes            int
	MaxFieldPathBytes              int
	MaxPropertyKeys                int
	MaxValidationIssues            int
	MaxValidationMessageBytes      int
	MaxProjectedVertices           int
	MaxProjectedEdges              int
	MaxTraversalSeedVertices       int
	MaxTraversalKindFilters        int
	MaxTraversalDepth              int
	CancellationCheckIntervalItems int
}

var currentV2 = Registry{
	MaxInputBytes:                  268435456,
	MaxSourceEntities:              100000,
	MaxSourceRelationships:         250000,
	MaxEntityFilters:               1000,
	MaxRelationshipFilters:         1000,
	MaxEntityMappings:              10000,
	MaxRelationshipMappings:        10000,
	MaxPropertyDefinitions:         10000,
	MaxMetadataMappings:            10000,
	MaxAggregationRules:            1000,
	MaxArrayItems:                  100000,
	MaxLabelsPerObject:             256,
	MaxIdentifierBytes:             255,
	MaxLabelBytes:                  256,
	MaxStringBytes:                 16384,
	MaxPropertyKeyBytes:            255,
	MaxFieldPathBytes:              511,
	MaxPropertyKeys:                1024,
	MaxValidationIssues:            100000,
	MaxValidationMessageBytes:      1024,
	MaxProjectedVertices:           MaximumResultVerticesV2,
	MaxProjectedEdges:              MaximumResultEdgesV2,
	MaxTraversalSeedVertices:       1024,
	MaxTraversalKindFilters:        1024,
	MaxTraversalDepth:              16,
	CancellationCheckIntervalItems: 1024,
}

func CurrentV2() Registry {
	return currentV2
}
