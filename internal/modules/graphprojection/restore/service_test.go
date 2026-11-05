package restore_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	. "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
)

type recordingRestorePublisher struct {
	events     []string
	plans      []RestorePublicationPlan
	clearErr   error
	afterClear func()
	err        error
	proof      func(RestorePublicationPlan) RestorePublicationProof
}

func (publisher *recordingRestorePublisher) Clear(context.Context) error {
	publisher.events = append(publisher.events, "clear")
	if publisher.clearErr != nil {
		return publisher.clearErr
	}
	if publisher.afterClear != nil {
		publisher.afterClear()
	}
	return nil
}

func (publisher *recordingRestorePublisher) ReplaceAll(_ context.Context, plan RestorePublicationPlan) (RestorePublicationProof, error) {
	publisher.events = append(publisher.events, "publish")
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
	t.Run("mixed dependency reconstruction", checkMixedDependencyReconstruction)
	_, registration, expected := currentRestoreFixture(t)
	publisher := &recordingRestorePublisher{}
	registry := recordingRestoreRegistry(t, registration, publisher)
	service, err := NewRestoreService(publisher, registry)
	if err != nil {
		t.Fatalf("construct restore service: %v", err)
	}
	request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
	result, err := service.Rebuild(request.Context, request)
	assertRestoreEvents(t, publisher, "clear", "enumerate", "publish")
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

func checkMixedDependencyReconstruction(t *testing.T) {
	t.Run("identity grouping", func(t *testing.T) { checkAggregationReconstruction(t, "identity") })
	t.Run("numeric grouping", func(t *testing.T) { checkAggregationReconstruction(t, "numeric") })
	t.Run("ordered contributor reconstruction", func(t *testing.T) { checkAggregationReconstruction(t, "ordered") })
}

func checkAggregationReconstruction(t *testing.T, mode string) {
	_, registration, initial := currentRestoreFixture(t)
	input := twoEntityProjectionInputV2(false)
	rule := func(id, target, scope, inputKind, outputKind, key string) map[string]any {
		return map[string]any{"aggregation_rule_id": id, "target_scope": target, "input_scope": scope, "input_kind": inputKind, "projected_kind": outputKind,
			"grouping_keys": []any{key}, "missing_grouping_key_behavior": "error", "property_merge_behavior": map[string]any{}}
	}
	root := rule("root", "vertex", "source_entity", "kind-a", "endpoint", "source_entity_id")
	edge := rule("edge", "edge", "source_entity", "kind-a", "bridge", "source_entity_id")
	edge["edge_direction"] = "directed"
	edge["endpoint_grouping"] = map[string]any{"src_vertex_aggregation_rule_id": "root", "dst_vertex_aggregation_rule_id": "root", "src_grouping_keys": []any{"source_entity_id"}, "dst_grouping_keys": []any{"source_entity_id"}, "missing_endpoint_behavior": "error"}
	leaf := rule("leaf", "vertex", "projected_edge", "bridge", "leaf", "projected.edge_id")
	wantVertices, wantEdges := 4, 1
	if mode != "identity" {
		for i, raw := range input["source_entities"].([]any) {
			entity := raw.(map[string]any)
			entity["source_entity_kind"] = "kind-a"
			entity["properties"] = map[string]any{"group": []json.Number{"1.5", "2.5"}[i]}
		}
		root["grouping_keys"], edge["grouping_keys"] = []any{"properties.group"}, []any{"properties.group"}
		endpoint := edge["endpoint_grouping"].(map[string]any)
		endpoint["src_grouping_keys"], endpoint["dst_grouping_keys"] = []any{"properties.group"}, []any{"properties.group"}
		wantVertices, wantEdges = 6, 2
		if mode == "ordered" {
			for _, raw := range input["source_entities"].([]any) {
				raw.(map[string]any)["properties"] = map[string]any{"group": json.Number("1.5")}
			}
			definitions := []any{}
			for _, merge := range []string{"first", "last"} {
				definitions = append(definitions, map[string]any{"property_definition_id": merge, "target_scope": "vertex", "target_kind": "endpoint", "source_field_path": "source_entity_id", "projected_key": merge, "projected_type": "string", "required": true, "missing_behavior": "error", "source_null_behavior": "error", "null_output_policy": "omit", "merge_behavior": merge})
			}
			input["property_definitions"] = definitions
			wantVertices, wantEdges = 4, 1
		}
	}
	input["projection_config"].(map[string]any)["aggregation_rules"] = []any{leaf, edge, root}
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := graphprojection.ProjectV2(context.Background(), graphprojection.InvocationContextV2{GraphViewID: initial.GraphViewID, SourceOwnerID: initial.SourceOwnerID}, body)
	if err != nil || len(projection.Vertices) != wantVertices || len(projection.Edges) != wantEdges {
		t.Fatalf("mixed restore fixture incomplete: %d/%d %v", len(projection.Vertices), len(projection.Edges), err)
	}
	if mode == "ordered" {
		for _, vertex := range projection.Vertices {
			if vertex.VertexKind == "endpoint" && (vertex.Properties["first"] != "entity-a" || vertex.Properties["last"] != "entity-b") {
				t.Fatal("restore fixture has wrong contributor merge order")
			}
		}
	}
	// Reconstruct from a different source/declaration order with the same binding.
	entities := input["source_entities"].([]any)
	entities[0], entities[1] = entities[1], entities[0]
	input["projection_config"].(map[string]any)["aggregation_rules"] = []any{root, edge, leaf}
	body, err = json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	completed, err := projection.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	for _, mismatch := range []bool{false, true} {
		selected := projection.ResultBindingV2()
		if mismatch {
			selected.CanonicalOutputSHA256 = strings.Repeat("f", 64)
		}
		registration.Enumerate = func(context.Context) ([]RestoreCandidate, error) {
			return []RestoreCandidate{{CandidateID: initial.GraphViewID, GraphViewID: initial.GraphViewID, SemanticQuerySchemaID: "cartulary.network_flow.graph_semantic_query.v3", SemanticInput: body, ExpectedBinding: selected}}, nil
		}
		publisher := &recordingRestorePublisher{}
		registry := recordingRestoreRegistry(t, registration, publisher)
		service, err := NewRestoreService(publisher, registry)
		if err != nil {
			t.Fatal(err)
		}
		request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
		result, err := service.Rebuild(request.Context, request)
		if mismatch {
			assertRestoreEvents(t, publisher, "clear", "enumerate")
			assertRestoreFailedWithoutCommittedFacts(t, result, err, RestoreErrorInvalidCandidate)
			if len(publisher.plans) != 0 {
				t.Fatal("mismatched binding was published")
			}
			continue
		}
		assertRestoreEvents(t, publisher, "clear", "enumerate", "publish")
		completed.PublishedAt = request.PublishedAt.UTC()
		if err != nil || !result.ReadinessSatisfied() || len(publisher.plans) != 1 || !reflect.DeepEqual(publisher.plans[0].Projections[0].Result, completed) {
			t.Fatalf("mixed reconstruction changed exact bytes: %v", err)
		}
	}
}

func TestGraphRestoreServiceRejectsMetadataBeforeClear_Unit(t *testing.T) {
	registry, registration, _ := currentRestoreFixture(t)
	base := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
	tests := []struct {
		name   string
		mutate func(*RestoreRebuildRequest)
		code   RestoreErrorCode
	}{
		{"missing publication time", func(r *RestoreRebuildRequest) { r.PublishedAt = time.Time{} }, RestoreErrorInvalidRequest},
		{"missing request context", func(r *RestoreRebuildRequest) { r.Context = nil }, RestoreErrorInvalidRequest},
		{"registry digest", func(r *RestoreRebuildRequest) { r.SourceRegistry.SHA256 = strings.Repeat("a", 64) }, RestoreErrorSourceRegistryMismatch},
		{"binding unavailable", func(r *RestoreRebuildRequest) { r.ImplementationBinding.SHA256 = strings.Repeat("b", 64) }, RestoreErrorUnsupportedGeneration},
		{"binding contents with original digest", func(r *RestoreRebuildRequest) { r.ImplementationBinding.Binding.BindingID = "altered" }, RestoreErrorUnsupportedGeneration},
		{"binding contents with recomputed digest", func(r *RestoreRebuildRequest) {
			r.ImplementationBinding.Binding.SemanticQuerySchemaIDs[0] = "unsupported.v999"
			body, err := json.Marshal(r.ImplementationBinding.Binding)
			if err != nil {
				t.Fatal(err)
			}
			r.ImplementationBinding.SHA256 = fmt.Sprintf("%x", sha256.Sum256(body))
		}, RestoreErrorUnsupportedGeneration},
		{"retired binding", func(r *RestoreRebuildRequest) {
			r.ImplementationBinding.Binding.SchemaID = "cartulary.graph_projection_restore_implementation_binding.v6"
		}, RestoreErrorUnsupportedGeneration},
		{"unsupported binding algorithm", func(r *RestoreRebuildRequest) { r.ImplementationBinding.Binding.AlgorithmID = "unsupported.v999" }, RestoreErrorUnsupportedGeneration},
		{"unsupported projection schema", func(r *RestoreRebuildRequest) {
			r.ImplementationBinding.Binding.ProjectionSchemaID = "unsupported.v999"
		}, RestoreErrorUnsupportedGeneration},
		{"malformed frozen binding", func(r *RestoreRebuildRequest) {
			// Preserve the selected request identity while supplying the rejected
			// definition; a decoder failure is not a new selected digest.
			r.ImplementationBinding.Binding = FrozenRestoreImplementationBinding([]byte(`{"schema_id":`), r.ImplementationBinding.SHA256).Binding
		}, RestoreErrorUnsupportedGeneration},
		{"catalog digest", func(r *RestoreRebuildRequest) { r.RecoveryStateCatalog.DigestSHA256 = strings.Repeat("c", 64) }, RestoreErrorRecoveryCatalogMismatch},
		{"algorithm", func(r *RestoreRebuildRequest) { r.RecoveryStateCatalog.AlgorithmID = "unknown" }, RestoreErrorRecoveryCatalogMismatch},
		{"tables", func(r *RestoreRebuildRequest) {
			r.RecoveryStateCatalog.GraphTableIDs = r.RecoveryStateCatalog.GraphTableIDs[:len(r.RecoveryStateCatalog.GraphTableIDs)-1]
		}, RestoreErrorRecoveryCatalogMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publisher := &recordingRestorePublisher{}
			registry := recordingRestoreRegistry(t, registration, publisher)
			service, err := NewRestoreService(publisher, registry)
			if err != nil {
				t.Fatalf("construct restore service: %v", err)
			}
			request := base
			request.RecoveryStateCatalog.GraphTableIDs = append([]string(nil), base.RecoveryStateCatalog.GraphTableIDs...)
			request.ImplementationBinding.Binding.SemanticQuerySchemaIDs = append([]string(nil), base.ImplementationBinding.Binding.SemanticQuerySchemaIDs...)
			request.ImplementationBinding.Binding.GraphTableIDs = append([]string(nil), base.ImplementationBinding.Binding.GraphTableIDs...)
			test.mutate(&request)
			result, err := service.Rebuild(context.Background(), request)
			assertRestoreEvents(t, publisher)
			assertRestoreFailedWithoutCommittedFacts(t, result, err, test.code)
		})
	}
}

func TestGraphRestoreServiceSourceFailuresAfterClearSkipPublication_Unit(t *testing.T) {
	_, registration, expected := currentRestoreFixture(t)
	tests := []struct {
		name      string
		enumerate RestoreCandidateEnumerator
		code      RestoreErrorCode
	}{
		{"source unavailable", func(context.Context) ([]RestoreCandidate, error) {
			return nil, errors.New("SECRET source")
		}, RestoreErrorSourceEnumeration},
		{"invalid semantic input", func(context.Context) ([]RestoreCandidate, error) {
			candidate := restoreCandidate(t, expected)
			candidate.SemanticInput = []byte(`{"SECRET":"invalid semantic input"}`)
			return []RestoreCandidate{candidate}, nil
		}, RestoreErrorInvalidCandidate},
		{"selected binding mismatch", func(context.Context) ([]RestoreCandidate, error) {
			candidate := restoreCandidate(t, expected)
			candidate.ExpectedBinding.CanonicalOutputSHA256 = strings.Repeat("f", 64)
			return []RestoreCandidate{candidate}, nil
		}, RestoreErrorInvalidCandidate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registration.Enumerate = test.enumerate
			publisher := &recordingRestorePublisher{}
			registry := recordingRestoreRegistry(t, registration, publisher)
			service, err := NewRestoreService(publisher, registry)
			if err != nil {
				t.Fatalf("construct restore service: %v", err)
			}
			request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
			result, err := service.Rebuild(request.Context, request)
			assertRestoreEvents(t, publisher, "clear", "enumerate")
			assertRestoreFailedWithoutCommittedFacts(t, result, err, test.code)
		})
	}
}

func TestGraphRestoreAcceptanceGPRA11OperationWideResourceLimits(t *testing.T) {
	_, registration, _ := currentRestoreFixture(t)
	registration.Enumerate = func(context.Context) ([]RestoreCandidate, error) {
		return make([]RestoreCandidate, RestoreMaximumCandidates+1), nil
	}
	publisher := &recordingRestorePublisher{}
	registry := recordingRestoreRegistry(t, registration, publisher)
	service, _ := NewRestoreService(publisher, registry)
	request := currentRestoreRequest(context.Background(), registry, CurrentRestoreImplementationBinding())
	result, err := service.Rebuild(request.Context, request)
	assertRestoreEvents(t, publisher, "clear", "enumerate")
	assertRestoreFailedWithoutCommittedFacts(t, result, err, RestoreErrorResourceOverflow)
}

func TestGraphRestoreServiceClearAndCancellationBoundaries_Unit(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		code   RestoreErrorCode
		events []string
	}{
		{"clear failure", RestoreErrorPublicationFailed, []string{"clear"}},
		{"cancel before admission", RestoreErrorInvalidRequest, nil},
		{"cancel caller after clear", RestoreErrorSourceEnumeration, []string{"clear"}},
		{"cancel request after clear", RestoreErrorSourceEnumeration, []string{"clear"}},
		{"cancel during enumeration", RestoreErrorInvalidCandidate, []string{"clear", "enumerate"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			_, registration, _ := currentRestoreFixture(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			requestContext, cancelRequest := context.WithCancel(t.Context())
			defer cancelRequest()
			publisher := &recordingRestorePublisher{}
			switch scenario.name {
			case "clear failure":
				publisher.clearErr = errors.New("SECRET clear failure")
			case "cancel before admission":
				cancel()
			case "cancel caller after clear":
				publisher.afterClear = cancel
			case "cancel request after clear":
				publisher.afterClear = cancelRequest
			case "cancel during enumeration":
				enumerate := registration.Enumerate
				registration.Enumerate = func(ctx context.Context) ([]RestoreCandidate, error) {
					candidates, err := enumerate(ctx)
					cancel()
					return candidates, err
				}
			}
			registry := recordingRestoreRegistry(t, registration, publisher)
			service, err := NewRestoreService(publisher, registry)
			if err != nil {
				t.Fatal(err)
			}
			request := currentRestoreRequest(requestContext, registry, CurrentRestoreImplementationBinding())
			result, err := service.Rebuild(ctx, request)
			assertRestoreEvents(t, publisher, scenario.events...)
			assertRestoreFailedWithoutCommittedFacts(t, result, err, scenario.code)
		})
	}
}

func recordingRestoreRegistry(t *testing.T, registration RestoreSourceRegistration, publisher *recordingRestorePublisher) *RestoreSourceRegistry {
	t.Helper()
	enumerate := registration.Enumerate
	registration.Enumerate = func(ctx context.Context) ([]RestoreCandidate, error) {
		publisher.events = append(publisher.events, "enumerate")
		return enumerate(ctx)
	}
	registry, err := NewCurrentRestoreSourceRegistry(registration)
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func assertRestoreEvents(t *testing.T, publisher *recordingRestorePublisher, want ...string) {
	t.Helper()
	if !slices.Equal(publisher.events, want) {
		t.Fatalf("restore phase order = %v, want %v", publisher.events, want)
	}
	if !slices.Contains(want, "publish") && len(publisher.plans) != 0 {
		t.Fatal("failed preparation crossed publication")
	}
}

func assertRestoreFailedWithoutCommittedFacts(t *testing.T, result RestoreRebuildResult, err error, want RestoreErrorCode) {
	t.Helper()
	if code, ok := RestoreErrorCodeOf(err); !ok || code != want {
		t.Fatalf("restore error = %q, %v; want %q", code, err, want)
	}
	if validationErr := result.Validate(); validationErr != nil {
		t.Fatalf("failed result violates the closed contract: %v", validationErr)
	}
	if result.Status != RestoreStatusFailed || result.ReadinessOutcome != RestoreReadinessIncomplete || result.ReadinessSatisfied() ||
		len(result.ClearedTableIDs) != 0 || len(result.RebuiltViews) != 0 || result.ReconciledNonterminalJobCount != 0 ||
		result.PostconditionSHA256 != nil || len(result.Errors) != 1 || result.Errors[0].Code != want {
		t.Fatalf("failed restore claimed committed facts or an invalid error: %#v", result)
	}
	body, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	if strings.Contains(string(body), "SECRET") || strings.Contains(err.Error(), "SECRET") {
		t.Fatal("closed failure leaked source detail")
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
			assertRestoreFailedWithoutCommittedFacts(t, result, err, RestoreErrorPostconditionFailed)
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
	assertRestoreFailedWithoutCommittedFacts(t, result, err, RestoreErrorOutcomeIndeterminate)
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
			SemanticQuerySchemaIDs:     []string{"cartulary.network_flow.graph_semantic_query.v3"},
			ProjectionInputContractID:  graphprojection.ProjectionSchemaIDV2,
			ProjectionResultContractID: "graph_projection_result.v2", Status: "active",
		},
		Enumerate: func(context.Context) ([]RestoreCandidate, error) {
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
		SemanticQuerySchemaID: "cartulary.network_flow.graph_semantic_query.v3",
		SemanticInput:         body, ExpectedBinding: expected,
	}
}

func currentRestoreRequest(ctx context.Context, registry *RestoreSourceRegistry, binding RestoreImplementationBindingRef) RestoreRebuildRequest {
	return RestoreRebuildRequest{
		Context: ctx, RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000007001"),
		PublishedAt:          time.Date(2026, 8, 15, 20, 0, 0, 0, time.UTC),
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
