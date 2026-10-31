package networkflow

import "github.com/google/uuid"

// Successful-value assertions only: policy remains in the checked production functions.
func canonicalTestValue[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func mustCanonicalJSON(value any) []byte { return canonicalTestValue(canonicalJSON(value)) }

func mustMappingFingerprint(mapping approvedMapping, sourceContentSHA256 string) string {
	return canonicalTestValue(mappingFingerprint(mapping, sourceContentSHA256))
}

func mustNormalizedRowDigest(mappingFingerprint string, values map[string]any, unmappedRaw retainedRawValues) string {
	return canonicalTestValue(normalizedRowDigest(mappingFingerprint, values, unmappedRaw))
}

func mustGraphQueryDigestV2(incidentID uuid.UUID, tableIDs []string, filters []queryFilter, timeRange graphTimeRange, aggregation graphAggregation) string {
	return canonicalTestValue(graphQueryDigestV2(incidentID, tableIDs, filters, timeRange, aggregation))
}

func mustGraphViewMutationBytes(route, pathIdentity string, body map[string]any) []byte {
	return canonicalTestValue(graphViewMutationBytes(route, pathIdentity, body))
}

func mustTableRenameRequestHash(tableID string, request tableRenameRequest) []byte {
	return canonicalTestValue(tableRenameRequestHash(tableID, request))
}

func mustTableSoftDeleteRequestHash(tableID string, request tableSoftDeleteRequest) []byte {
	return canonicalTestValue(tableSoftDeleteRequestHash(tableID, request))
}

func mustIndicatorLinkRequestHash(request indicatorLinkRequest) []byte {
	return canonicalTestValue(indicatorLinkRequestHash(request))
}
