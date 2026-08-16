package networkflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestNetworkFlowGraphTelemetryBoundaryContainsObserverFailure_Unit(t *testing.T) {
	observer := panicGraphTelemetryObserver{}
	observeGraphPhaseSafely(context.Background(), observer, GraphPhaseTelemetryObservation{Phase: "source_scan"})
	observeGraphResultSafely(context.Background(), observer, GraphResultTelemetryObservation{Vertices: 1})
	observeGraphCleanupSafely(context.Background(), observer, GraphCleanupTelemetryObservation{DeletedResults: 1})

	sweeper := &graphResultCleanupDispatcherTestSweeper{
		calls: make(chan *graphprojection.ResultCleanupCandidateV2, 1),
		responses: []graphResultCleanupDispatcherTestResponse{{result: graphResultCleanupSweepResult{
			DeletedLeases: 2, DeletedResults: 1, HealthSnapshotValid: true,
		}}},
	}
	dispatcher, err := newGraphResultCleanupDispatcher(sweeper, time.Now, func() {})
	if err != nil {
		t.Fatal(err)
	}
	dispatcher.telemetry = observer
	result, err := dispatcher.runOnce(context.Background(), nil)
	if err != nil || result.DeletedLeases != 2 || result.DeletedResults != 1 {
		t.Fatalf("telemetry panic changed cleanup product result = %#v err=%v", result, err)
	}
}

func TestNetworkFlowGraphTelemetryOutcomeVocabulary_Unit(t *testing.T) {
	tests := []struct {
		name       string
		apiErr     *httpapi.APIError
		wantResult string
		wantClass  string
	}{
		{name: "success", wantResult: "success"},
		{name: "cancel", apiErr: graphProjectionFailed("projection_cancelled"), wantResult: "canceled"},
		{name: "timeout", apiErr: graphProjectionFailed("projection_timeout"), wantResult: "timeout", wantClass: "timeout"},
		{name: "limit", apiErr: graphLimitExceeded("vertex_limit_exceeded", "limit", 1, 2), wantResult: "rejected", wantClass: "policy_rejected"},
		{name: "dependency", apiErr: graphProjectionFailed("projection_unavailable"), wantResult: "failed", wantClass: "dependency_unavailable"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, class := graphTelemetryOutcomeForAPIError(test.apiErr)
			if result != test.wantResult || class != test.wantClass {
				t.Fatalf("outcome = %q/%q want %q/%q", result, class, test.wantResult, test.wantClass)
			}
		})
	}
	if result, class := graphTelemetryOutcomeForError(context.Canceled); result != "canceled" || class != "" {
		t.Fatalf("canceled error outcome = %q/%q", result, class)
	}
	if result, class := graphTelemetryOutcomeForError(context.DeadlineExceeded); result != "timeout" || class != "timeout" {
		t.Fatalf("timeout error outcome = %q/%q", result, class)
	}
	if result, class := graphTelemetryOutcomeForError(errors.New("SENTINEL raw SQL secret")); result != "failed" || class != "internal_error" {
		t.Fatalf("raw error outcome = %q/%q", result, class)
	}
}

func TestNetworkFlowGraphTelemetrySemanticBoundaries_Unit(t *testing.T) {
	observer := &recordingGraphTelemetryObserver{}
	service := &Service{graphTelemetry: observer}
	service.observeGraphPhase(context.Background(), graphTelemetryPhaseSourceValidation, "default_flow_edge_v1", time.Now(), nil)
	service.observeGraphPhase(
		context.Background(), graphTelemetryPhaseSourceScan, "default_flow_edge_v1", time.Now(),
		graphProjectionFailedForContext(context.Canceled),
	)
	service.observeGraphPhase(
		context.Background(), graphTelemetryPhaseProjection, "default_flow_edge_v1", time.Now(),
		graphProjectionFailedForContext(context.DeadlineExceeded),
	)
	service.observeGraphComposition(context.Background(), graphComposition{
		SemanticQuery:    map[string]any{"aggregation": map[string]any{"mode": "default_flow_edge_v1"}},
		ContributingRows: 4,
		Vertices:         map[string]*graphVertex{"one": {}},
		Edges:            map[string]*graphEdge{"one": {}, "two": {}},
	})

	module := &Module{graphTelemetry: observer}
	module.observeGraphPhase(
		context.Background(), graphTelemetryPhasePublication, "default_flow_edge_v1", time.Now(),
		errors.New("SENTINEL publication error"),
	)

	if len(observer.phases) != 4 {
		t.Fatalf("phase observations = %d want 4: %#v", len(observer.phases), observer.phases)
	}
	wantOutcomes := [][2]string{{"success", ""}, {"canceled", ""}, {"timeout", "timeout"}, {"failed", "internal_error"}}
	for index, observation := range observer.phases {
		if observation.Operation != graphTelemetryOperationMaterialization || observation.GraphMode != "default_flow_edge_v1" || observation.Duration < 0 {
			t.Fatalf("phase %d has unsafe or invalid semantics: %#v", index, observation)
		}
		if observation.Result != wantOutcomes[index][0] || observation.ErrorClass != wantOutcomes[index][1] {
			t.Fatalf("phase %d outcome = %q/%q want %q/%q", index, observation.Result, observation.ErrorClass, wantOutcomes[index][0], wantOutcomes[index][1])
		}
	}
	if len(observer.results) != 1 || observer.results[0].ContributingRows != 4 ||
		observer.results[0].Vertices != 1 || observer.results[0].Edges != 2 || observer.results[0].Result != "success" {
		t.Fatalf("result observations = %#v", observer.results)
	}

	sweeper := &graphResultCleanupDispatcherTestSweeper{
		calls:     make(chan *graphprojection.ResultCleanupCandidateV2, 1),
		responses: []graphResultCleanupDispatcherTestResponse{{err: errors.New("SENTINEL cleanup storage error")}},
	}
	dispatcher, err := newGraphResultCleanupDispatcher(sweeper, time.Now, func() {})
	if err != nil {
		t.Fatal(err)
	}
	dispatcher.telemetry = observer
	if _, err := dispatcher.runOnce(context.Background(), nil); err == nil {
		t.Fatal("cleanup failure was not returned to the dispatcher")
	}
	if len(observer.cleanups) != 1 || observer.cleanups[0].Operation != graphTelemetryOperationCleanup ||
		observer.cleanups[0].Result != "failed" || observer.cleanups[0].ErrorClass != "internal_error" || observer.cleanups[0].Duration < 0 {
		t.Fatalf("cleanup observations = %#v", observer.cleanups)
	}
}

type panicGraphTelemetryObserver struct{}

func (panicGraphTelemetryObserver) ObserveGraphPhase(context.Context, GraphPhaseTelemetryObservation) {
	panic("SENTINEL phase observer failure")
}

func (panicGraphTelemetryObserver) ObserveGraphResult(context.Context, GraphResultTelemetryObservation) {
	panic("SENTINEL result observer failure")
}

func (panicGraphTelemetryObserver) ObserveGraphCleanup(context.Context, GraphCleanupTelemetryObservation) {
	panic("SENTINEL cleanup observer failure")
}

type recordingGraphTelemetryObserver struct {
	phases   []GraphPhaseTelemetryObservation
	results  []GraphResultTelemetryObservation
	cleanups []GraphCleanupTelemetryObservation
}

func (observer *recordingGraphTelemetryObserver) ObserveGraphPhase(_ context.Context, observation GraphPhaseTelemetryObservation) {
	observer.phases = append(observer.phases, observation)
}

func (observer *recordingGraphTelemetryObserver) ObserveGraphResult(_ context.Context, observation GraphResultTelemetryObservation) {
	observer.results = append(observer.results, observation)
}

func (observer *recordingGraphTelemetryObserver) ObserveGraphCleanup(_ context.Context, observation GraphCleanupTelemetryObservation) {
	observer.cleanups = append(observer.cleanups, observation)
}
