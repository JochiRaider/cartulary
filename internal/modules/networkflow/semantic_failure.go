package networkflow

import (
	"errors"
	"time"
)

// semanticFailure carries owner facts, never a transport status or envelope.
type semanticFailure struct {
	kind    failureKind
	reason  string
	details failureDetails
	cause   error
}

type failureKind string

const (
	failureInternal                   failureKind = "internal_error"
	failureInvalidRequest             failureKind = "network_flow_invalid_request"
	failureInvalidFilter              failureKind = "network_flow_invalid_filter"
	failureInvalidSort                failureKind = "network_flow_invalid_sort"
	failureInvalidTableScope          failureKind = "network_flow_invalid_table_scope"
	failureInvalidLimit               failureKind = "network_flow_invalid_limit"
	failureCursorInvalid              failureKind = "network_flow_cursor_invalid"
	failureInvalidLimitOverride       failureKind = "network_flow_invalid_limit_override"
	failureInvalidGraphAggregation    failureKind = "network_flow_invalid_graph_aggregation"
	failureInvalidTimeRange           failureKind = "network_flow_invalid_time_range"
	failureTableNotFound              failureKind = "network_flow_table_not_found"
	failureTableNotActive             failureKind = "network_flow_table_not_active"
	failureGraphQueryStale            failureKind = "network_flow_graph_query_stale"
	failureGraphLimitExceeded         failureKind = "network_flow_graph_limit_exceeded"
	failureCounterSumLimitExceeded    failureKind = "network_flow_counter_sum_limit_exceeded"
	failureGraphProjectionFailed      failureKind = "network_flow_graph_projection_failed"
	failureGraphMaterializationFailed failureKind = "network_flow_graph_materialization_failed"
)

type failureDetails struct {
	Field                           string
	FieldKey, Op                    *string
	FilterIndex                     *int
	LimitKey, Phase, Digest         string
	Limit, Actual, Minimum, Maximum int
	StartUTC, EndUTC                *time.Time
}

func (f *semanticFailure) Error() string { return string(f.kind) }
func (f *semanticFailure) Unwrap() error { return f.cause }

func newSemanticFailure(kind failureKind, field, reason string) *semanticFailure {
	return &semanticFailure{kind: kind, reason: reason, details: failureDetails{Field: field}}
}

func internalSemanticFailure(cause error) *semanticFailure {
	return &semanticFailure{kind: failureInternal, cause: cause}
}

func tableReadFailure(err error) *semanticFailure {
	switch {
	case errors.Is(err, errTableNotFound):
		return newSemanticFailure(failureTableNotFound, "network_flow_table_id", "not_found")
	case errors.Is(err, errTableNotActive):
		return newSemanticFailure(failureTableNotActive, "network_flow_table_id", "soft_deleted")
	default:
		return internalSemanticFailure(err)
	}
}
