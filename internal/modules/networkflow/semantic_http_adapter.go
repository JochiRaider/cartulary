package networkflow

import (
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
	"io"
)

func networkFlowAPIError(status int, code string, field string, reason string) *httpapi.APIError {
	details := map[string]any{"reason_code": reason}
	if field != "" {
		details["field"] = field
	}
	return &httpapi.APIError{Status: status, Code: code, Message: code, Details: details}
}

func bindGraphV2ResponseMetadataHTTP(composition *graphComposition) *httpapi.APIError {
	value0 := bindGraphV2ResponseMetadata(composition)
	return semanticHTTPError(value0)
}

func deriveTimeBucketIndexFromExactResultHTTP(
	timeRange graphTimeRange,
	width int64,
	limit int,
	projection map[string]any,
) ([]graphTimeBucket, *httpapi.APIError) {
	value0, value1 := deriveTimeBucketIndexFromExactResult(timeRange, width, limit, projection)
	return value0, semanticHTTPError(value1)
}

func malformedStoredGraphResultHTTP() *httpapi.APIError {
	value0 := malformedStoredGraphResult()
	return semanticHTTPError(value0)
}

func decodeGraphTimeRangeV2HTTP(raw json.RawMessage) (graphTimeRange, *httpapi.APIError) {
	value0, value1 := decodeGraphTimeRangeV2(raw)
	return value0, semanticHTTPError(value1)
}

func decodeGraphAggregationV2HTTP(raw json.RawMessage) (graphAggregation, *httpapi.APIError) {
	value0, value1 := decodeGraphAggregationV2(raw)
	return value0, semanticHTTPError(value1)
}

func validateAggregationTimeRangeHTTP(aggregation graphAggregation, timeRange graphTimeRange) *httpapi.APIError {
	value0 := validateAggregationTimeRange(aggregation, timeRange)
	return semanticHTTPError(value0)
}

func decodeGraphResultLimitsHTTP(raw json.RawMessage, limits EffectiveLimits) (graphResultLimits, *httpapi.APIError) {
	value0, value1 := decodeGraphResultLimits(raw, limits)
	return value0, semanticHTTPError(value1)
}

func decodeGraphSelectorHTTP(raw json.RawMessage) (graphSelector, *httpapi.APIError) {
	value0, value1 := decodeGraphSelector(raw)
	return value0, semanticHTTPError(value1)
}

func decodeGraphSemanticRequestHTTP(raw json.RawMessage, limits EffectiveLimits) (graphSemanticRequest, *httpapi.APIError) {
	value0, value1 := decodeGraphSemanticRequest(raw, limits)
	return value0, semanticHTTPError(value1)
}

func (s *routeService) composeGraphHTTP(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, request graphQueryRequest) (graphComposition, *httpapi.APIError) {
	value0, value1 := s.graphComposer.composeGraph(ctx, incidentID, actorUserID, request)
	return value0, semanticHTTPError(value1)
}

func (s *routeService) composeGraphSourceFromSemanticHTTP(ctx context.Context, incidentID uuid.UUID, semantic graphSemanticRequest) (graphComposition, *httpapi.APIError) {
	value0, value1 := s.graphComposer.composeGraphSourceFromSemantic(ctx, incidentID, semantic)
	return value0, semanticHTTPError(value1)
}

func rowMatchesGraphQueryHTTP(row flowRow, filters []queryFilter, timeRange graphTimeRange, aggregation graphAggregation) (bool, *httpapi.APIError) {
	value0, value1 := rowMatchesGraphQuery(row, filters, timeRange, aggregation)
	return value0, semanticHTTPError(value1)
}

func (s *routeService) queryGraphContributorPageHTTP(ctx context.Context, incidentID uuid.UUID, semantic graphSemanticRequest, expectedDigest string, selector graphSelector, position *contributorCursorPosition, limit int) ([]flowRow, bool, map[string]int, *httpapi.APIError) {
	value0, value1, value2, value3 := s.graphComposer.queryGraphContributorPage(ctx, incidentID, semantic, expectedDigest, selector, position, limit)
	return value0, value1, value2, semanticHTTPError(value3)
}

func graphProjectionFailedForContextHTTP(err error) *httpapi.APIError {
	value0 := graphProjectionFailedForContext(err)
	return semanticHTTPError(value0)
}

func graphQueryStaleHTTP(reason string, digest string) *httpapi.APIError {
	value0 := graphQueryStale(reason, digest)
	return semanticHTTPError(value0)
}

func decodeAcceptedRowQueryRequestHTTP(reader io.Reader, expectedSchemaID string, continuationSchemaID string, limits EffectiveLimits) (rowQueryRequest, *httpapi.APIError) {
	value0, value1 := decodeAcceptedRowQueryRequest(reader, expectedSchemaID, continuationSchemaID, limits)
	return value0, semanticHTTPError(value1)
}

func decodeRejectedRowsQueryRequestHTTP(reader io.Reader, limits EffectiveLimits) (rejectedRowsQueryRequest, *httpapi.APIError) {
	value0, value1 := decodeRejectedRowsQueryRequest(reader, limits)
	return value0, semanticHTTPError(value1)
}

func decodeNetworkFlowObjectHTTP(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	value0, value1 := decodeNetworkFlowObject(reader)
	return value0, semanticHTTPError(value1)
}

func requiredJSONStringHTTP(raw map[string]json.RawMessage, field string) (string, *httpapi.APIError) {
	value0, value1 := requiredJSONString(raw, field)
	return value0, semanticHTTPError(value1)
}

func ensureAllowedMembersHTTP(raw map[string]json.RawMessage, allowed ...string) *httpapi.APIError {
	value0 := ensureAllowedMembers(raw, allowed...)
	return semanticHTTPError(value0)
}

func requiredTableScopeHTTP(raw json.RawMessage, limits EffectiveLimits) (tableScope, *httpapi.APIError) {
	value0, value1 := requiredTableScope(raw, limits)
	return value0, semanticHTTPError(value1)
}

func decodeFiltersHTTP(raw json.RawMessage, limits EffectiveLimits) ([]queryFilter, *httpapi.APIError) {
	value0, value1 := decodeFilters(raw, limits)
	return value0, semanticHTTPError(value1)
}

func decodePositiveIntHTTP(raw json.RawMessage, field string) (int, *httpapi.APIError) {
	value0, value1 := decodePositiveInt(raw, field)
	return value0, semanticHTTPError(value1)
}

func decodeIntegerRangeHTTP(raw json.RawMessage) (*int64, *int64, *httpapi.APIError) {
	value0, value1, value2 := decodeIntegerRange(raw)
	return value0, value1, semanticHTTPError(value2)
}

func invalidNetworkFlowRequestHTTP(field string, reason string) *httpapi.APIError {
	value0 := invalidNetworkFlowRequest(field, reason)
	return semanticHTTPError(value0)
}

func invalidTableScopeHTTP(field string, reason string) *httpapi.APIError {
	value0 := invalidTableScope(field, reason)
	return semanticHTTPError(value0)
}

func invalidLimitHTTP(field string, reason string) *httpapi.APIError {
	value0 := invalidLimit(field, reason)
	return semanticHTTPError(value0)
}

func cursorInvalidHTTP(reason string) *httpapi.APIError {
	value0 := cursorInvalid(reason)
	return semanticHTTPError(value0)
}
