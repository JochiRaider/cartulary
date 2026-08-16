package networkflow

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	graphTelemetryOperationMaterialization = "graph_materialization"
	graphTelemetryOperationCleanup         = "cleanup_sweep"
	graphTelemetryPhaseSourceValidation    = "source_validation"
	graphTelemetryPhaseSourceScan          = "source_scan"
	graphTelemetryPhaseProjection          = "projection"
	graphTelemetryPhasePublication         = "publication"
)

// GraphPhaseTelemetryObservation is an owner-safe phase boundary. It contains
// only closed vocabulary and aggregate timing; identifiers and raw errors are
// intentionally not representable.
type GraphPhaseTelemetryObservation struct {
	Operation  string
	Phase      string
	GraphMode  string
	Result     string
	ErrorClass string
	Duration   time.Duration
}

// GraphResultTelemetryObservation reports only bounded result cardinality.
type GraphResultTelemetryObservation struct {
	GraphMode        string
	Result           string
	ContributingRows int
	Vertices         int
	Edges            int
	TimeBuckets      int
}

// GraphCleanupTelemetryObservation contains a post-sweep health snapshot from
// the same source-owner eligibility predicate used by cleanup. Oldest age is
// absent when no eligible result exists.
type GraphCleanupTelemetryObservation struct {
	Operation               string
	Result                  string
	ErrorClass              string
	Duration                time.Duration
	DeletedLeases           int
	DeletedResults          int
	HealthSnapshotValid     bool
	EligibleResultBacklog   int64
	OldestEligibleResultAge *time.Duration
}

// GraphTelemetryObserver is supplied by application assembly. Network Flow
// owns semantic boundaries and the platform adapter owns OpenTelemetry APIs.
type GraphTelemetryObserver interface {
	ObserveGraphPhase(context.Context, GraphPhaseTelemetryObservation)
	ObserveGraphResult(context.Context, GraphResultTelemetryObservation)
	ObserveGraphCleanup(context.Context, GraphCleanupTelemetryObservation)
}

func observeGraphPhaseSafely(ctx context.Context, observer GraphTelemetryObserver, observation GraphPhaseTelemetryObservation) {
	if observer == nil {
		return
	}
	defer func() { _ = recover() }()
	observer.ObserveGraphPhase(ctx, observation)
}

func observeGraphResultSafely(ctx context.Context, observer GraphTelemetryObserver, observation GraphResultTelemetryObservation) {
	if observer == nil {
		return
	}
	defer func() { _ = recover() }()
	observer.ObserveGraphResult(ctx, observation)
}

func observeGraphCleanupSafely(ctx context.Context, observer GraphTelemetryObserver, observation GraphCleanupTelemetryObservation) {
	if observer == nil {
		return
	}
	defer func() { _ = recover() }()
	observer.ObserveGraphCleanup(ctx, observation)
}

func graphTelemetryOutcomeForAPIError(apiErr *httpapi.APIError) (string, string) {
	if apiErr == nil {
		return "success", ""
	}
	reason, _ := apiErr.Details["reason_code"].(string)
	switch reason {
	case "projection_cancelled", "cancelled":
		return "canceled", ""
	case "projection_timeout", "timeout":
		return "timeout", "timeout"
	}
	switch apiErr.Status {
	case http.StatusConflict:
		return "conflict", "lifecycle_conflict"
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return "failed", "dependency_unavailable"
	}
	if apiErr.Status >= 400 && apiErr.Status < 500 {
		switch apiErr.Code {
		case "network_flow_table_not_found", "network_flow_graph_view_not_found":
			return "rejected", "not_found"
		case "network_flow_graph_limit_exceeded", "network_flow_counter_sum_limit_exceeded":
			return "rejected", "policy_rejected"
		default:
			return "rejected", "request_invalid"
		}
	}
	return "failed", "internal_error"
}

func graphTelemetryOutcomeForError(err error) (string, string) {
	switch {
	case err == nil:
		return "success", ""
	case errors.Is(err, context.Canceled):
		return "canceled", ""
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout", "timeout"
	case errors.Is(err, ErrGraphViewPublicationStale):
		return "conflict", "lifecycle_conflict"
	case errors.Is(err, graphprojection.ErrResultV2IdentityConflict):
		return "conflict", "invariant_violation"
	default:
		return "failed", "internal_error"
	}
}

func (s *Service) observeGraphPhase(
	ctx context.Context,
	phase string,
	graphMode string,
	started time.Time,
	apiErr *httpapi.APIError,
) {
	if s == nil {
		return
	}
	result, errorClass := graphTelemetryOutcomeForAPIError(apiErr)
	observeGraphPhaseSafely(ctx, s.graphTelemetry, GraphPhaseTelemetryObservation{
		Operation: graphTelemetryOperationMaterialization,
		Phase:     phase, GraphMode: graphMode, Result: result, ErrorClass: errorClass,
		Duration: nonnegativeGraphTelemetryDuration(started),
	})
}

func (m *Module) observeGraphPhase(
	ctx context.Context,
	phase string,
	graphMode string,
	started time.Time,
	err error,
) {
	if m == nil {
		return
	}
	result, errorClass := graphTelemetryOutcomeForError(err)
	observeGraphPhaseSafely(ctx, m.graphTelemetry, GraphPhaseTelemetryObservation{
		Operation: graphTelemetryOperationMaterialization,
		Phase:     phase, GraphMode: graphMode, Result: result, ErrorClass: errorClass,
		Duration: nonnegativeGraphTelemetryDuration(started),
	})
}

func (s *Service) observeGraphComposition(ctx context.Context, composition graphComposition) {
	if s == nil {
		return
	}
	observeGraphResultSafely(ctx, s.graphTelemetry, GraphResultTelemetryObservation{
		GraphMode: compositionTelemetryMode(composition), Result: "success",
		ContributingRows: composition.ContributingRows,
		Vertices:         len(composition.Vertices), Edges: len(composition.Edges),
		TimeBuckets: len(composition.TimeBuckets),
	})
}

func (m *Module) observeGraphComposition(ctx context.Context, composition graphComposition) {
	if m == nil {
		return
	}
	observeGraphResultSafely(ctx, m.graphTelemetry, GraphResultTelemetryObservation{
		GraphMode: compositionTelemetryMode(composition), Result: "success",
		ContributingRows: composition.ContributingRows,
		Vertices:         len(composition.Vertices), Edges: len(composition.Edges),
		TimeBuckets: len(composition.TimeBuckets),
	})
}

func compositionTelemetryMode(composition graphComposition) string {
	if mode, ok := composition.SemanticQuery["aggregation"].(map[string]any); ok {
		if value, ok := mode["mode"].(string); ok {
			return value
		}
	}
	return "default_flow_edge_v1"
}

func nonnegativeGraphTelemetryDuration(started time.Time) time.Duration {
	duration := time.Since(started)
	if duration < 0 {
		return 0
	}
	return duration
}
