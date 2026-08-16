package graphprojection

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type restoreTestSourceState struct{}

func (restoreTestSourceState) GraphProjectionRestoreSourceState() {}

type recordingRestorePublisher struct {
	plans []RestorePublicationPlan
	err   error
}

func (publisher *recordingRestorePublisher) ReplaceAll(_ context.Context, plan RestorePublicationPlan) (RestorePublicationProof, error) {
	publisher.plans = append(publisher.plans, plan)
	if publisher.err != nil {
		return RestorePublicationProof{}, publisher.err
	}
	return RestorePublicationProof{RebuiltViews: append([]RestoreRebuiltView{}, plan.RebuiltViews...), PostconditionSHA256: plan.PostconditionSHA256}, nil
}

func TestGraphRestoreAcceptanceGPRA02Through05ExactCandidateSemantics(t *testing.T) {
	registry, _, expected := restoreV2Fixture(t)
	publisher := &recordingRestorePublisher{}
	service, err := NewRestoreService(publisher, registry, RestoreServiceOptions{})
	if err != nil {
		t.Fatalf("construct restore service: %v", err)
	}
	request := restoreV2Request(context.Background(), registry, CurrentRestoreImplementationBinding())
	result, err := service.Rebuild(request.Context, request)
	if err != nil || !result.ReadinessSatisfied() || len(result.RebuiltViews) != 1 || len(publisher.plans) != 1 {
		t.Fatalf("rebuild exact selected result: result=%#v err=%v", result, err)
	}
	if got := result.RebuiltViews[0]; got.ProjectionResultID != expected.ProjectionResultID || got.GraphViewID != expected.GraphViewID || got.SourceSnapshotID != expected.SourceSnapshotID {
		t.Fatalf("rebuilt exact identity drifted: got=%#v want=%#v", got, expected)
	}
	if publisher.plans[0].Projections[0].Result.Binding != expected {
		t.Fatalf("published binding drifted: %#v", publisher.plans[0].Projections[0].Result.Binding)
	}
}

func TestGraphRestoreAcceptanceGPRA06Through11PreflightFailsBeforeMutation(t *testing.T) {
	registry, _, _ := restoreV2Fixture(t)
	base := restoreV2Request(context.Background(), registry, CurrentRestoreImplementationBinding())
	tests := []struct {
		name   string
		mutate func(*RestoreRebuildRequest)
		code   RestoreErrorCode
	}{
		{"registry digest", func(r *RestoreRebuildRequest) { r.SourceRegistry.SHA256 = strings.Repeat("a", 64) }, RestoreErrorSourceRegistryMismatch},
		{"binding unavailable", func(r *RestoreRebuildRequest) { r.ImplementationBinding.SHA256 = strings.Repeat("b", 64) }, RestoreErrorBindingUnavailable},
		{"catalog digest", func(r *RestoreRebuildRequest) { r.RecoveryStateCatalog.DigestSHA256 = strings.Repeat("c", 64) }, RestoreErrorRecoveryCatalogMismatch},
		{"algorithm", func(r *RestoreRebuildRequest) { r.RecoveryStateCatalog.AlgorithmID = "unknown" }, RestoreErrorRecoveryCatalogMismatch},
		{"tables", func(r *RestoreRebuildRequest) {
			r.RecoveryStateCatalog.GraphTableIDs = r.RecoveryStateCatalog.GraphTableIDs[:3]
		}, RestoreErrorRecoveryCatalogMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publisher := &recordingRestorePublisher{}
			service, err := NewRestoreService(publisher, registry, RestoreServiceOptions{})
			if err != nil {
				t.Fatalf("construct restore service: %v", err)
			}
			request := base
			request.RecoveryStateCatalog.GraphTableIDs = append([]string(nil), base.RecoveryStateCatalog.GraphTableIDs...)
			test.mutate(&request)
			result, err := service.Rebuild(request.Context, request)
			if code, ok := RestoreErrorCodeOf(err); !ok || code != test.code || result.ReadinessOutcome != RestoreReadinessIncomplete || len(publisher.plans) != 0 {
				t.Fatalf("preflight code/result got %q,%v %#v plans=%d", code, ok, result, len(publisher.plans))
			}
		})
	}
}

func TestGraphRestoreAcceptanceGPRA06And10SourceFailuresBeforeMutation(t *testing.T) {
	_, registration, expected := restoreV2Fixture(t)
	tests := []struct {
		name      string
		enumerate RestoreCandidateEnumerator
		code      RestoreErrorCode
	}{
		{"source unavailable", func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
			return nil, errors.New("SECRET source")
		}, RestoreErrorSourceEnumeration},
		{"selected binding mismatch", func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
			candidate := restoreCandidate(t, expected)
			candidate.ExpectedBinding.CanonicalOutputSHA256 = strings.Repeat("f", 64)
			return []RestoreCandidate{candidate}, nil
		}, RestoreErrorInvalidCandidate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registration.Enumerate = test.enumerate
			registry, err := NewCurrentRestoreSourceRegistry(registration)
			if err != nil {
				t.Fatalf("construct registry: %v", err)
			}
			publisher := &recordingRestorePublisher{}
			service, err := NewRestoreService(publisher, registry, RestoreServiceOptions{})
			if err != nil {
				t.Fatalf("construct restore service: %v", err)
			}
			request := restoreV2Request(context.Background(), registry, CurrentRestoreImplementationBinding())
			result, err := service.Rebuild(request.Context, request)
			if code, ok := RestoreErrorCodeOf(err); !ok || code != test.code || len(publisher.plans) != 0 {
				t.Fatalf("source failure got %q,%v result=%#v", code, ok, result)
			}
			body, _ := json.Marshal(result)
			if strings.Contains(string(body), "SECRET") || strings.Contains(err.Error(), "SECRET") {
				t.Fatal("closed error leaked source detail")
			}
		})
	}
}

func TestGraphRestoreAcceptanceGPRA11OperationWideResourceLimits(t *testing.T) {
	usage := restoreResourceUsage{sourceRegistrations: RestoreMaximumSourceRegistrations}
	if !usage.addCandidates(RestoreMaximumCandidates) || usage.addCandidates(1) {
		t.Fatal("candidate limit drifted")
	}
	usage = restoreResourceUsage{}
	if !usage.addNormalizedInputBytes(RestoreMaximumNormalizedInputBytes) || usage.addNormalizedInputBytes(1) {
		t.Fatal("input limit drifted")
	}
	usage = restoreResourceUsage{}
	if !usage.addGraphSize(RestoreMaximumVertices, RestoreMaximumEdges) || usage.addGraphSize(1, 0) {
		t.Fatal("graph limit drifted")
	}

	registry, registration, _ := restoreV2Fixture(t)
	_ = registry
	registration.Enumerate = func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
		return make([]RestoreCandidate, RestoreMaximumCandidates+1), nil
	}
	registry, err := NewCurrentRestoreSourceRegistry(registration)
	if err != nil {
		t.Fatalf("construct registry: %v", err)
	}
	publisher := &recordingRestorePublisher{}
	service, _ := NewRestoreService(publisher, registry, RestoreServiceOptions{})
	request := restoreV2Request(context.Background(), registry, CurrentRestoreImplementationBinding())
	result, err := service.Rebuild(request.Context, request)
	if code, ok := RestoreErrorCodeOf(err); !ok || code != RestoreErrorResourceOverflow || len(publisher.plans) != 0 || result.ReadinessOutcome != RestoreReadinessIncomplete {
		t.Fatalf("overflow crossed publication: code=%q result=%#v", code, result)
	}
}

func TestGraphRestoreAcceptanceGPRA15RetryPreservesDeterministicOutput(t *testing.T) {
	registry, _, expected := restoreV2Fixture(t)
	publisher := &recordingRestorePublisher{}
	service, _ := NewRestoreService(publisher, registry, RestoreServiceOptions{Now: func() time.Time { return time.Now().UTC() }})
	request := restoreV2Request(context.Background(), registry, CurrentRestoreImplementationBinding())
	first, firstErr := service.Rebuild(request.Context, request)
	second, secondErr := service.Rebuild(request.Context, request)
	if firstErr != nil || secondErr != nil || first.RebuiltViews[0] != second.RebuiltViews[0] || first.RebuiltViews[0].ProjectionResultID != expected.ProjectionResultID {
		t.Fatalf("retry changed immutable result: first=%#v second=%#v errs=%v/%v", first, second, firstErr, secondErr)
	}
}

func TestGraphRestoreAcceptanceGPRA13And17IndeterminateFailureIsClosed(t *testing.T) {
	registry, _, _ := restoreV2Fixture(t)
	publisher := &recordingRestorePublisher{err: &RestorePublicationError{Indeterminate: true, Cause: errors.New("database SECRET")}}
	service, _ := NewRestoreService(publisher, registry, RestoreServiceOptions{})
	request := restoreV2Request(context.Background(), registry, CurrentRestoreImplementationBinding())
	result, err := service.Rebuild(request.Context, request)
	if code, ok := RestoreErrorCodeOf(err); !ok || code != RestoreErrorOutcomeIndeterminate || result.ReadinessOutcome != RestoreReadinessIncomplete {
		t.Fatalf("indeterminate got %q %#v", code, result)
	}
	body, _ := json.Marshal(result)
	if strings.Contains(string(body), "SECRET") || strings.Contains(err.Error(), "SECRET") {
		t.Fatal("closed failure leaked cause")
	}
}

func restoreV2Fixture(t *testing.T) (*RestoreSourceRegistry, RestoreSourceRegistration, ResultBindingV2) {
	t.Helper()
	input := twoEntityProjectionInputV2(false)
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	const graphViewID = "nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	result, err := NewEngineV2().Project(context.Background(), InvocationContextV2{GraphViewID: graphViewID, SourceOwnerID: "network_flow_activity"}, body)
	if err != nil {
		t.Fatalf("project fixture: %v", err)
	}
	expected := result.ResultBindingV2()
	registration := RestoreSourceRegistration{
		Entry: RestoreSourceRegistryEntry{
			SourceRegistrationID: "network_flow_activity.graph_views.v1", SourceOwnerID: "network_flow_activity",
			AuthoritativeFamilyID: "network_flow_activity.graph_views", EnumeratorBindingID: "network_flow_activity.graph_view_restore_enumerator_v1",
			ValidityBindingID: "network_flow_activity.graph_view_restore_validity_v1", ProjectionInputContractID: ProjectionSchemaIDV2,
			ProjectionResultContractID: "graph_projection_result.v2", Status: "active",
		},
		Enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
			return []RestoreCandidate{restoreCandidate(t, expected)}, nil
		},
		EvaluateValidity: func(context.Context, RestoreCandidate, time.Time) (RestoreCandidateValidity, error) {
			return RestoreCandidateValidity{Eligible: true}, nil
		},
	}
	registry, err := NewCurrentRestoreSourceRegistry(registration)
	if err != nil {
		t.Fatalf("construct exact current registry: %v", err)
	}
	return registry, registration, expected
}

func restoreCandidate(t *testing.T, expected ResultBindingV2) RestoreCandidate {
	t.Helper()
	body, err := json.Marshal(twoEntityProjectionInputV2(false))
	if err != nil {
		t.Fatal(err)
	}
	return RestoreCandidate{CandidateID: expected.GraphViewID, GraphViewID: expected.GraphViewID, SemanticInput: body, ExpectedBinding: expected}
}

func restoreV2Request(ctx context.Context, registry *RestoreSourceRegistry, binding RestoreImplementationBindingRef) RestoreRebuildRequest {
	return RestoreRebuildRequest{
		Context: ctx, RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000007001"), RestoredSourceState: restoreTestSourceState{},
		BackupSetID: uuid.MustParse("00000000-0000-0000-0000-000000007002"), ConsistencyPointAt: time.Date(2026, 8, 15, 20, 0, 0, 0, time.UTC),
		TargetGenerationID:   uuid.MustParse("00000000-0000-0000-0000-000000007003"),
		RecoveryStateCatalog: RestoreRecoveryCatalogRef{DigestSHA256: CurrentRestoreImplementationBinding().Binding.RecoveryStateCatalogSHA256, AlgorithmID: RestoreAlgorithmID, GraphTableIDs: RestoreGraphTableIDs()},
		SourceRegistry:       RestoreSourceRegistryRef{Registry: registry, SHA256: registry.DigestSHA256()}, ImplementationBinding: binding,
	}
}
