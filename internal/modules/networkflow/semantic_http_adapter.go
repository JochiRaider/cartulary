package networkflow

import (
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"io"
)

func networkFlowAPIError(status int, code string, field string, reason string) *httpapi.APIError {
	details := map[string]any{"reason_code": reason}
	if field != "" {
		details["field"] = field
	}
	return &httpapi.APIError{Status: status, Code: code, Message: code, Details: details}
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
