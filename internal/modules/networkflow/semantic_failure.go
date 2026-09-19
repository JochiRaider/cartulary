package networkflow

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
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
	failureInvalidIndicatorSelector failureKind = "network_flow_invalid_indicator_selector"
	failureInvalidIndicatorTarget   failureKind = "network_flow_invalid_indicator_target"
	failureIndicatorLinkForbidden   failureKind = "network_flow_indicator_link_forbidden"
	failureIndicatorLinkAmbiguous   failureKind = "network_flow_indicator_link_ambiguous"
	failureResourceLimit            failureKind = "network_flow_resource_limit_exceeded"
	failureTransactionConflict      failureKind = "transaction_conflict"
	failureTransactionTimeout       failureKind = "service_unavailable"
	failureClientTxnConflict        failureKind = "client_txn_conflict"
	failureAdmission                failureKind = "incident_admission"

	failureInternal                   failureKind = "internal_error"
	failureInvalidDisplayName         failureKind = "network_flow_invalid_display_name"
	failureInvalidRequest             failureKind = "network_flow_invalid_request"
	failureInvalidFilter              failureKind = "network_flow_invalid_filter"
	failureInvalidSort                failureKind = "network_flow_invalid_sort"
	failureInvalidTableScope          failureKind = "network_flow_invalid_table_scope"
	failureInvalidLimit               failureKind = "network_flow_invalid_limit"
	failureCursorInvalid              failureKind = "network_flow_cursor_invalid"
	failureInvalidLimitOverride       failureKind = "network_flow_invalid_limit_override"
	failureInvalidGraphAggregation    failureKind = "network_flow_invalid_graph_aggregation"
	failureInvalidTimeRange           failureKind = "network_flow_invalid_time_range"
	failureGraphViewNotFound          failureKind = "network_flow_graph_view_not_found"
	failureGraphViewNotMaterialized   failureKind = "network_flow_graph_view_not_materialized"
	failureTableNotFound              failureKind = "network_flow_table_not_found"
	failureTableNotActive             failureKind = "network_flow_table_not_active"
	failureGraphQueryStale            failureKind = "network_flow_graph_query_stale"
	failureGraphLimitExceeded         failureKind = "network_flow_graph_limit_exceeded"
	failureCounterSumLimitExceeded    failureKind = "network_flow_counter_sum_limit_exceeded"
	failureGraphProjectionFailed      failureKind = "network_flow_graph_projection_failed"
	failureGraphMaterializationFailed failureKind = "network_flow_graph_materialization_failed"
)

type failureDetails struct {
	LinkContext, LinkGraph                                            bool
	SelectorKind, LinkFieldKey, TargetMode, Candidate                 string
	ActualKind, SourceTableID, RequiredRole, ClientTxnID, OperationID string
	TableIDs                                                          []string
	TimeoutSeconds                                                    float64

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

func clientTxnFailure(txn string) *semanticFailure {
	return &semanticFailure{kind: failureClientTxnConflict, details: failureDetails{ClientTxnID: txn}}
}
func applicationAdmissionFailure(err error, role string) *semanticFailure {
	var denied *admission.Denied
	if errors.As(err, &denied) {
		return &semanticFailure{kind: failureAdmission, cause: err, details: failureDetails{RequiredRole: role}}
	}
	return internalSemanticFailure(err)
}

func tableNameFailure(err error) *semanticFailure {
	var invalid *invalidDisplayNameError
	if errors.As(err, &invalid) {
		return &semanticFailure{kind: failureInvalidDisplayName, reason: invalid.ReasonCode, details: failureDetails{Field: "display_name", Limit: 64, Actual: invalid.NormalizedLength}, cause: err}
	}
	return internalSemanticFailure(err)
}
