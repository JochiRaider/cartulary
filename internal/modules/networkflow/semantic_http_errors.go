package networkflow

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// semanticHTTPError is the sole semantic-to-HTTP policy. Raw causes never
// enter the response; unrecognized kinds fail closed.
func semanticHTTPError(f *semanticFailure) *httpapi.APIError {
	if f == nil {
		return nil
	}
	status := http.StatusBadRequest
	switch f.kind {
	case failureInvalidRequest, failureInvalidFilter, failureInvalidSort,
		failureInvalidTableScope, failureInvalidLimit, failureCursorInvalid,
		failureInvalidLimitOverride, failureInvalidGraphAggregation, failureInvalidTimeRange:
	case failureTableNotFound:
		status = http.StatusNotFound
	case failureTableNotActive, failureGraphQueryStale:
		status = http.StatusConflict
	case failureGraphLimitExceeded, failureCounterSumLimitExceeded:
		status = http.StatusRequestEntityTooLarge
	case failureGraphProjectionFailed:
		status = http.StatusBadGateway
	case failureGraphMaterializationFailed:
		status = http.StatusInternalServerError
	default:
		return httpapi.InternalAPIError(f)
	}
	d := f.details
	details := map[string]any{"reason_code": f.reason}
	if d.Field != "" {
		details["field"] = d.Field
	}
	switch f.kind {
	case failureInvalidFilter:
		details["field_key"], details["op"], details["filter_index"] = nil, nil, nil
		if d.FieldKey != nil {
			details["field_key"] = *d.FieldKey
		}
		if d.Op != nil {
			details["op"] = *d.Op
		}
		if d.FilterIndex != nil {
			details["filter_index"] = *d.FilterIndex
		}
		details["retry_action"] = "correct_request"
	case failureCursorInvalid:
		details["retry_action"] = "restart_query"
	case failureInvalidGraphAggregation:
		details["retry_action"] = "correct_request"
	case failureInvalidTimeRange:
		details["start_utc"], details["end_utc"] = nullableTimestamp(d.StartUTC), nullableTimestamp(d.EndUTC)
		details["retry_action"] = "correct_request"
	case failureInvalidLimitOverride:
		if d.LimitKey != "" {
			details["limit_key"], details["limit"] = d.LimitKey, d.Limit
			details["minimum"], details["maximum"] = d.Minimum, d.Maximum
		}
	case failureGraphLimitExceeded, failureCounterSumLimitExceeded:
		details["limit_key"], details["limit"], details["actual"] = d.LimitKey, d.Limit, d.Actual
		details["phase"], details["retry_action"] = d.Phase, "reduce_scope_or_limits"
	case failureGraphProjectionFailed:
		details["projection_contract_version"], details["retry_action"] = graphProjectionSchemaID, "do_not_retry"
	case failureGraphMaterializationFailed:
		details["retry_action"] = "do_not_retry"
	case failureGraphQueryStale:
		details["graph_query_digest"], details["retry_action"] = d.Digest, "refresh_resource"
	}
	return &httpapi.APIError{Status: status, Code: string(f.kind), Message: string(f.kind), Details: details}
}
