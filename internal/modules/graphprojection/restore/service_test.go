package restore_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	. "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
)

type restoreTestSourceState struct{}

func (restoreTestSourceState) GraphProjectionRestoreSourceState() {}

type recordingRestorePublisher struct {
	plans []RestorePublicationPlan
	err   error
	proof func(RestorePublicationPlan) RestorePublicationProof
}

func (publisher *recordingRestorePublisher) ReplaceAll(_ context.Context, plan RestorePublicationPlan) (RestorePublicationProof, error) {
	publisher.plans = append(publisher.plans, plan)
	if publisher.err != nil {
		return RestorePublicationProof{}, publisher.err
	}
	if publisher.proof != nil {
		return publisher.proof(plan), nil
	}
	return RestorePublicationProof{RebuiltViews: append([]RestoreRebuiltView{}, plan.RebuiltViews...), PostconditionSHA256: plan.PostconditionSHA256}, nil
}

func TestGraphRestoreServiceRequiresPublisherAndRegistry_Unit(t *testing.T) {
	publisher := &recordingRestorePublisher{}
	if service, err := NewRestoreService(nil, CurrentRestoreSourceRegistry()); err == nil || service != nil {
		t.Fatalf("nil publisher constructor result = %#v, %v", service, err)
	}
	if service, err := NewRestoreService(publisher, nil); err == nil || service != nil {
		t.Fatalf("nil registry constructor result = %#v, %v", service, err)
	}
}

func TestGraphRestoreAcceptanceGPRA02Through05ExactCandidateSemantics(t *testing.T) {
	registry, _, expected := currentRestoreFixture(t)
	publisher := &recordingRestorePublisher{}
	service, err := NewRestoreService(publisher, registry)
	if err != nil {
		t.Fatalf("construct restore service: %v", err)
	}
	request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
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
	registry, _, _ := currentRestoreFixture(t)
	base := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
	tests := []struct {
		name   string
		mutate func(*RestoreRebuildRequest)
		code   RestoreErrorCode
	}{
		{"registry digest", func(r *RestoreRebuildRequest) { r.SourceRegistry.SHA256 = strings.Repeat("a", 64) }, RestoreErrorSourceRegistryMismatch},
		{"binding unavailable", func(r *RestoreRebuildRequest) { r.ImplementationBinding.SHA256 = strings.Repeat("b", 64) }, RestoreErrorUnsupportedGeneration},
		{"catalog digest", func(r *RestoreRebuildRequest) { r.RecoveryStateCatalog.DigestSHA256 = strings.Repeat("c", 64) }, RestoreErrorRecoveryCatalogMismatch},
		{"algorithm", func(r *RestoreRebuildRequest) { r.RecoveryStateCatalog.AlgorithmID = "unknown" }, RestoreErrorRecoveryCatalogMismatch},
		{"tables", func(r *RestoreRebuildRequest) {
			r.RecoveryStateCatalog.GraphTableIDs = r.RecoveryStateCatalog.GraphTableIDs[:3]
		}, RestoreErrorRecoveryCatalogMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publisher := &recordingRestorePublisher{}
			service, err := NewRestoreService(publisher, registry)
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
	_, registration, expected := currentRestoreFixture(t)
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
			service, err := NewRestoreService(publisher, registry)
			if err != nil {
				t.Fatalf("construct restore service: %v", err)
			}
			request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
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
	registry, registration, _ := currentRestoreFixture(t)
	_ = registry
	registration.Enumerate = func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
		return make([]RestoreCandidate, RestoreMaximumCandidates+1), nil
	}
	registry, err := NewCurrentRestoreSourceRegistry(registration)
	if err != nil {
		t.Fatalf("construct registry: %v", err)
	}
	publisher := &recordingRestorePublisher{}
	service, _ := NewRestoreService(publisher, registry)
	request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
	result, err := service.Rebuild(request.Context, request)
	if code, ok := RestoreErrorCodeOf(err); !ok || code != RestoreErrorResourceOverflow || len(publisher.plans) != 0 || result.ReadinessOutcome != RestoreReadinessIncomplete {
		t.Fatalf("overflow crossed publication: code=%q result=%#v", code, result)
	}
}

func TestGraphRestoreAcceptanceGPRA15RetryPreservesDeterministicOutput(t *testing.T) {
	registry, _, expected := currentRestoreFixture(t)
	publisher := &recordingRestorePublisher{}
	service, _ := NewRestoreService(publisher, registry)
	request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
	first, firstErr := service.Rebuild(request.Context, request)
	second, secondErr := service.Rebuild(request.Context, request)
	if firstErr != nil || secondErr != nil || first.RebuiltViews[0] != second.RebuiltViews[0] || first.RebuiltViews[0].ProjectionResultID != expected.ProjectionResultID {
		t.Fatalf("retry changed immutable result: first=%#v second=%#v errs=%v/%v", first, second, firstErr, secondErr)
	}
}

func TestGraphRestoreReadinessRequiresExactReconciliationProof_Unit(t *testing.T) {
	registry, _, _ := currentRestoreFixture(t)
	tests := []struct {
		name  string
		proof func(RestorePublicationPlan) RestorePublicationProof
	}{
		{
			name: "postcondition digest mismatch",
			proof: func(plan RestorePublicationPlan) RestorePublicationProof {
				return RestorePublicationProof{RebuiltViews: plan.RebuiltViews, PostconditionSHA256: strings.Repeat("f", 64)}
			},
		},
		{
			name: "rebuilt binding mismatch",
			proof: func(plan RestorePublicationPlan) RestorePublicationProof {
				views := append([]RestoreRebuiltView{}, plan.RebuiltViews...)
				views[0].ProjectionResultID = "gpres_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				return RestorePublicationProof{RebuiltViews: views, PostconditionSHA256: plan.PostconditionSHA256}
			},
		},
		{
			name: "negative job count",
			proof: func(plan RestorePublicationPlan) RestorePublicationProof {
				return RestorePublicationProof{RebuiltViews: plan.RebuiltViews, ReconciledNonterminalJobCount: -1, PostconditionSHA256: plan.PostconditionSHA256}
			},
		},
		{
			name: "negative lease count",
			proof: func(plan RestorePublicationPlan) RestorePublicationProof {
				return RestorePublicationProof{RebuiltViews: plan.RebuiltViews, ReconciledLeaseCount: -1, PostconditionSHA256: plan.PostconditionSHA256}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publisher := &recordingRestorePublisher{proof: test.proof}
			service, err := NewRestoreService(publisher, registry)
			if err != nil {
				t.Fatal(err)
			}
			request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
			result, err := service.Rebuild(request.Context, request)
			if code, ok := RestoreErrorCodeOf(err); !ok || code != RestoreErrorPostconditionFailed || result.ReadinessSatisfied() || result.ReadinessOutcome != RestoreReadinessIncomplete || len(publisher.plans) != 1 {
				t.Fatalf("invalid reconciliation proof became ready: result=%#v code=%q ok=%t err=%v", result, code, ok, err)
			}
		})
	}
}

func TestGraphRestoreAcceptanceGPRA13And17IndeterminateFailureIsClosed(t *testing.T) {
	registry, _, _ := currentRestoreFixture(t)
	publisher := &recordingRestorePublisher{err: &RestorePublicationError{Indeterminate: true, Cause: errors.New("database SECRET")}}
	service, _ := NewRestoreService(publisher, registry)
	request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
	result, err := service.Rebuild(request.Context, request)
	if code, ok := RestoreErrorCodeOf(err); !ok || code != RestoreErrorOutcomeIndeterminate || result.ReadinessOutcome != RestoreReadinessIncomplete {
		t.Fatalf("indeterminate got %q %#v", code, result)
	}
	body, _ := json.Marshal(result)
	if strings.Contains(string(body), "SECRET") || strings.Contains(err.Error(), "SECRET") {
		t.Fatal("closed failure leaked cause")
	}
}

func currentRestoreFixture(t *testing.T) (*RestoreSourceRegistry, RestoreSourceRegistration, graphprojection.ResultBindingV2) {
	t.Helper()
	input := twoEntityProjectionInputV2(false)
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	const graphViewID = "nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	result, err := graphprojection.ProjectV2(context.Background(), graphprojection.InvocationContextV2{GraphViewID: graphViewID, SourceOwnerID: "network_flow_activity"}, body)
	if err != nil {
		t.Fatalf("project fixture: %v", err)
	}
	expected := result.ResultBindingV2()
	registration := RestoreSourceRegistration{
		Entry: RestoreSourceRegistryEntry{
			SourceRegistrationID: "network_flow_activity.graph_views.v1", SourceOwnerID: "network_flow_activity",
			AuthoritativeFamilyID: "network_flow_activity.graph_views", EnumeratorBindingID: "network_flow_activity.graph_view_restore_enumerator_v2",
			SemanticQuerySchemaIDs:     []string{"cartulary.network_flow.graph_semantic_query.v2"},
			ProjectionInputContractID:  graphprojection.ProjectionSchemaIDV2,
			ProjectionResultContractID: "graph_projection_result.v2", Status: "active",
		},
		Enumerate: func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error) {
			return []RestoreCandidate{restoreCandidate(t, expected)}, nil
		},
	}
	registry, err := NewCurrentRestoreSourceRegistry(registration)
	if err != nil {
		t.Fatalf("construct exact current registry: %v", err)
	}
	return registry, registration, expected
}

func restoreCandidate(t *testing.T, expected graphprojection.ResultBindingV2) RestoreCandidate {
	t.Helper()
	body, err := json.Marshal(twoEntityProjectionInputV2(false))
	if err != nil {
		t.Fatal(err)
	}
	return RestoreCandidate{
		CandidateID: expected.GraphViewID, GraphViewID: expected.GraphViewID,
		SemanticQuerySchemaID: "cartulary.network_flow.graph_semantic_query.v2",
		SemanticInput:         body, ExpectedBinding: expected,
	}
}

func currentRestoreRequest(ctx context.Context, registry *RestoreSourceRegistry, binding RestoreImplementationBindingRef) RestoreRebuildRequest {
	return RestoreRebuildRequest{
		Context: ctx, RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000007001"), RestoredSourceState: restoreTestSourceState{},
		BackupSetID: uuid.MustParse("00000000-0000-0000-0000-000000007002"), ConsistencyPointAt: time.Date(2026, 8, 15, 20, 0, 0, 0, time.UTC),
		TargetGenerationID:   uuid.MustParse("00000000-0000-0000-0000-000000007003"),
		RecoveryStateCatalog: RestoreRecoveryCatalogRef{DigestSHA256: CurrentRestoreImplementationBinding().Binding.RecoveryStateCatalogSHA256, AlgorithmID: RestoreAlgorithmID, GraphTableIDs: RestoreGraphTableIDs()},
		SourceRegistry:       RestoreSourceRegistryRef{Registry: registry, SHA256: registry.DigestSHA256()}, ImplementationBinding: binding,
	}
}

func twoEntityProjectionInputV2(reverse bool) map[string]any {
	kinds := []any{"kind-a", "kind-b"}
	mappings := []any{
		map[string]any{"mapping_rule_id": "map-a", "source_entity_kind": "kind-a", "projected_vertex_kind": "vertex", "inclusion_predicate": "always", "label_policy": "mapping_only", "mapping_labels": []any{}, "required_property_keys": []any{}, "optional_property_keys": []any{}},
		map[string]any{"mapping_rule_id": "map-b", "source_entity_kind": "kind-b", "projected_vertex_kind": "vertex", "inclusion_predicate": "always", "label_policy": "mapping_only", "mapping_labels": []any{}, "required_property_keys": []any{}, "optional_property_keys": []any{}},
	}
	entities := []any{
		map[string]any{"source_entity_id": "entity-a", "source_entity_kind": "kind-a", "properties": map[string]any{}, "metadata": map[string]any{"number": json.Number("1.0")}, "labels": []any{}},
		map[string]any{"source_entity_id": "entity-b", "source_entity_kind": "kind-b", "properties": map[string]any{}, "metadata": map[string]any{"number": json.Number("1e0")}, "labels": []any{}},
	}
	if reverse {
		kinds[0], kinds[1] = kinds[1], kinds[0]
		mappings[0], mappings[1] = mappings[1], mappings[0]
		entities[0], entities[1] = entities[1], entities[0]
	}
	return map[string]any{
		"projection_schema_id": graphprojection.ProjectionSchemaIDV2,
		"source_snapshot_id":   "snapshot-ordering",
		"projection_config": map[string]any{
			"projection_version": "test-v1", "declared_source_entity_kinds": kinds, "declared_source_relationship_kinds": []any{},
			"entity_mappings": mappings, "relationship_mappings": []any{}, "metadata_mappings": []any{}, "aggregation_rules": []any{},
			"default_vertex_labels": []any{}, "default_edge_labels": []any{}, "allow_empty_kind_registry": false,
		},
		"source_entities": entities, "source_relationships": []any{}, "source_metadata": map[string]any{},
		"filters": map[string]any{"entity_filters": []any{}, "relationship_filters": []any{}, "logic": "and"}, "property_definitions": []any{},
	}
}
