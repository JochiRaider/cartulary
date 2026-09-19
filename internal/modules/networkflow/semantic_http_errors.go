package networkflow

import (
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// semanticHTTPError is the sole semantic-to-HTTP policy. Raw causes never
// enter the response; unrecognized kinds fail closed.
func semanticHTTPError(f *semanticFailure) *httpapi.APIError {
	if f == nil {
		return nil
	}
	if f.kind == failureAdmission {
		return savedGraphAdmissionError(f.cause, f.details.RequiredRole)
	}
	if f.kind == failureClientTxnConflict {
		return httpapi.ClientTxnConflictError(f.details.ClientTxnID)
	}
	status := http.StatusBadRequest
	switch f.kind {
	case failureInvalidDisplayName, failureInvalidRequest, failureInvalidFilter, failureInvalidSort,
		failureInvalidTableScope, failureInvalidLimit, failureCursorInvalid,
		failureInvalidLimitOverride, failureInvalidGraphAggregation, failureInvalidTimeRange:
	case failureInvalidIndicatorSelector, failureInvalidIndicatorTarget, failureIndicatorLinkAmbiguous:
	case failureIndicatorLinkForbidden:
		status = http.StatusForbidden
	case failureTransactionConflict:
		status = http.StatusConflict
	case failureTransactionTimeout:
		status = http.StatusServiceUnavailable
	case failureTableNotFound, failureGraphViewNotFound:
		status = http.StatusNotFound
	case failureTableNotActive, failureGraphQueryStale, failureGraphViewNotMaterialized:
		status = http.StatusConflict
	case failureGraphLimitExceeded, failureCounterSumLimitExceeded, failureResourceLimit:
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
	case failureInvalidDisplayName:
		details["max_length"], details["normalized_length"], details["retry_action"] = d.Limit, d.Actual, "correct_request"
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
	case failureGraphLimitExceeded, failureCounterSumLimitExceeded, failureResourceLimit:
		details["limit_key"], details["limit"], details["actual"] = d.LimitKey, d.Limit, d.Actual
		details["phase"], details["retry_action"] = d.Phase, "reduce_scope_or_limits"
	case failureGraphProjectionFailed:
		details["projection_contract_version"], details["retry_action"] = graphProjectionSchemaID, "do_not_retry"
	case failureGraphMaterializationFailed:
		details["retry_action"] = "do_not_retry"
	case failureGraphQueryStale:
		details["graph_query_digest"], details["retry_action"] = d.Digest, "refresh_resource"
	}
	if d.LinkContext || d.LinkGraph {
		if strings.HasPrefix(string(f.kind), "network_flow_") {
			if _, ok := details["retry_action"]; !ok {
				details["retry_action"] = "correct_request"
			}
		}
	}
	if d.LinkGraph && f.kind == failureInvalidTableScope {
		tables := d.TableIDs
		if tables == nil {
			tables = []string{}
		}
		details["mode"], details["table_ids"], details["limit_key"] = "selected_tables", tables, linkNullable(d.LimitKey)
	}
	if d.SourceTableID != "" {
		details["network_flow_table_id"], details["table_status"], details["allowed_states"], details["retry_action"] = d.SourceTableID, nil, []string{"active"}, "refresh_resource"
		if f.kind == failureTableNotActive {
			details["table_status"] = tableStatusSoftDeleted
		}
	}
	if d.LinkContext && f.kind == failureInvalidRequest {
		details["field"], details["actual_kind"], details["expected_contract"], details["retry_action"] = linkNullable(d.Field), linkNullable(d.ActualKind), schemaIndicatorLinkRequest, "correct_request"
	}
	switch f.kind {
	case failureInvalidIndicatorSelector, failureInvalidIndicatorTarget, failureIndicatorLinkForbidden, failureIndicatorLinkAmbiguous:
		if d.LinkContext {
			details["selector_kind"], details["field_key"], details["target_mode"], details["resolved_candidate_value"] = linkNullable(d.SelectorKind), linkNullable(d.LinkFieldKey), linkNullable(d.TargetMode), linkNullable(d.Candidate)
		} else {
			if d.SelectorKind != "" {
				details["selector_kind"] = d.SelectorKind
			}
			if d.LinkFieldKey != "" {
				details["field_key"] = d.LinkFieldKey
			}
			if d.TargetMode != "" {
				details["target_mode"] = d.TargetMode
			}
			if d.Candidate != "" {
				details["resolved_candidate_value"] = d.Candidate
			}
		}
		details["retry_action"] = "correct_request"
		if f.kind == failureIndicatorLinkForbidden {
			details["retry_action"] = "do_not_retry"
		}
	case failureTransactionTimeout:
		details["operation_id"], details["timeout_seconds"] = d.OperationID, d.TimeoutSeconds
	}
	return &httpapi.APIError{Status: status, Code: string(f.kind), Message: string(f.kind), Details: details, Retryable: f.kind == failureTransactionConflict || f.kind == failureTransactionTimeout}
}
