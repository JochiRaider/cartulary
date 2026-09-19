package networkflow

import (
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
	"time"
)

func graphTimeBucketsHTTP(timeRange graphTimeRange, width int64, limit int) ([]graphTimeBucket, *httpapi.APIError) {
	value0, value1 := graphTimeBuckets(timeRange, width, limit)
	return value0, semanticHTTPError(value1)
}
func decodeOptionalTimestampHTTP(raw json.RawMessage, field string) (*time.Time, *httpapi.APIError) {
	value0, value1 := decodeOptionalTimestamp(raw, field)
	return value0, semanticHTTPError(value1)
}
func decodeLowerableGraphLimitHTTP(raw json.RawMessage, key string, minimum int, maximum int) (int, *httpapi.APIError) {
	value0, value1 := decodeLowerableGraphLimit(raw, key, minimum, maximum)
	return value0, semanticHTTPError(value1)
}
func composeGraphRowHTTP(incidentID uuid.UUID, row flowRow, tableByID map[string]tableRecord, composition *graphComposition) *httpapi.APIError {
	value0 := composeGraphRow(incidentID, row, tableByID, composition)
	return semanticHTTPError(value0)
}
func validateGraphLimitsHTTP(composition graphComposition) *httpapi.APIError {
	value0 := validateGraphLimits(composition)
	return semanticHTTPError(value0)
}
func (s *graphSourceComposer) projectNetworkFlowGraphHTTP(ctx context.Context, actorUserID uuid.UUID, sourceSnapshotID string, composition graphComposition, requestedAt time.Time) (map[string]any, *httpapi.APIError) {
	value0, value1 := s.projectNetworkFlowGraph(ctx, actorUserID, sourceSnapshotID, composition, requestedAt)
	return value0, semanticHTTPError(value1)
}
func canonicalGraphContributorPredicateHTTP(incidentID uuid.UUID, selector graphSelector) (graphContributorPredicate, *httpapi.APIError) {
	value0, value1 := canonicalGraphContributorPredicate(incidentID, selector)
	return value0, semanticHTTPError(value1)
}
func rowMatchesFilterHTTP(row flowRow, filter queryFilter) (bool, *httpapi.APIError) {
	value0, value1 := rowMatchesFilter(row, filter)
	return value0, semanticHTTPError(value1)
}
func invalidFilterHTTP(field string, reason string) *httpapi.APIError {
	value0 := invalidFilter(field, reason)
	return semanticHTTPError(value0)
}
func graphProjectionFailedForProjectionErrorHTTP(err error) *httpapi.APIError {
	value0 := graphProjectionFailedForProjectionError(err)
	return semanticHTTPError(value0)
}
