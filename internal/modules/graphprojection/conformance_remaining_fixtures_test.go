package graphprojection

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestGPFIX001Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-001")
	_, err := admitProjectionInput([]byte(`{"projection_schema_id":`), admitOptions{Operation: "project_ephemeral"})
	assertOperationError(t, err, "invalid_projection_request", "invalid_json_syntax")
}

func TestGPFIX002Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-002")
	_, err := admitProjectionInput([]byte(`{"projection_schema_id":"graph_projection.v1","projection_schema_id":"graph_projection.v1"}`), admitOptions{Operation: "project_ephemeral"})
	var operationError *OperationError
	if !errors.As(err, &operationError) || operationError.ReasonCode != "duplicate_object_member" || operationError.Field != "$.projection_schema_id" {
		t.Fatalf("duplicate member error = %#v", err)
	}
}

func TestGPFIX003Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-003")
	input := minimalInput(t, "fix003")
	input["graph_view_id"] = "gv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_, err := admitProjectionInput(mustJSON(t, input), admitOptions{Operation: "create_projection"})
	var operationError *OperationError
	if !errors.As(err, &operationError) || operationError.ReasonCode != "invalid_graph_view_id" {
		t.Fatalf("graph view id error = %#v", err)
	}
	if strings.Contains(mustJSONText(t, operationError.Details), "fix003") || strings.Contains(mustJSONText(t, operationError.Details), "expected") {
		t.Fatalf("graph view id error leaked private details: %#v", operationError.Details)
	}
}

func TestGPFIX005Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-005")
	previousLimits := graphProjectionLimits
	graphProjectionLimits.MaxValidationIssues = 3
	defer func() { graphProjectionLimits = previousLimits }()
	run := ProjectionRun{GraphViewID: "gv_test", ProjectionRunID: "gpr_test"}
	discovered := []ValidationIssue{
		run.issue("warning", "invalid_filter", "filter", "c", nil, nil),
		run.issue("error", "invalid_mapping_rule", "mapping_rule", "b", nil, nil),
		run.issue("fatal", "invalid_projection_config", "projection_input", "a", nil, nil),
		run.issue("error", "invalid_property_definition", "property_definition", "d", nil, nil),
	}
	summary := validationSummary(run, discovered)
	if len(summary.Issues) != 3 || summary.Issues[2].Code != "validation_issue_limit_exceeded" || summary.Status != "failed" {
		t.Fatalf("capped summary = %#v", summary)
	}
}

func TestGPFIX006Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-006")
	definition := candidateDefinition{ProjectedType: "string", DefaultValue: "defaulted", HasDefaultValue: true, MissingBehavior: "default", SourceNullBehavior: "default", NullOutputPolicy: "omit"}
	value, include, code := evaluateCandidate(definition, nil, false)
	if value != "defaulted" || !include || code != "" {
		t.Fatalf("missing/default candidate = value:%#v include:%v code:%q", value, include, code)
	}
	value, include, code = evaluateCandidate(definition, nil, true)
	if value != "defaulted" || !include || code != "" {
		t.Fatalf("explicit-null/default candidate = value:%#v include:%v code:%q", value, include, code)
	}
	definition.SourceNullBehavior = "emit_null"
	definition.NullOutputPolicy = "emit_null"
	value, include, code = evaluateCandidate(definition, nil, true)
	if value != nil || !include || code != "" {
		t.Fatalf("explicit-null emit candidate = value:%#v include:%v code:%q", value, include, code)
	}
}

func TestGPFIX007Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-007")
	_, ok, conflict := mergeValues("single_value", []any{"defaulted", "present"})
	if ok || !conflict {
		t.Fatalf("single_value conflict ok=%v conflict=%v", ok, conflict)
	}
}

func TestGPFIX008Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-008")
	input := incidentGraphInput(t)
	definitions := input["property_definitions"].([]any)
	definitions[1].(map[string]any)["source_field_path"] = "properties.unused_missing"
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "fix008", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	vertex := findVertexKind(t, *run.GraphView, "site_vertex")
	if got := vertex.Properties["host_count"]; got != 2 {
		t.Fatalf("count aggregate = %#v", got)
	}
}

func TestGPFIX009Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-009")
	input := incidentGraphInput(t)
	relationship := input["source_relationships"].([]any)[0].(map[string]any)
	delete(relationship["properties"].(map[string]any), "src_site")
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "fix009", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if run.State != RunStateAvailable || !hasIssueCode(run.ValidationSummary.Issues, "aggregation_endpoint_missing") || hasEdgeKind(*run.GraphView, "site_logon_edge") {
		t.Fatalf("missing endpoint result state=%s issues=%#v graph=%#v", run.State, run.ValidationSummary.Issues, run.GraphView)
	}
}

func TestGPFIX010Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-010")
	input := incidentGraphInput(t)
	relationship := input["source_relationships"].([]any)[0].(map[string]any)
	delete(relationship["properties"].(map[string]any), "src_site")
	aggregation := input["projection_config"].(map[string]any)["aggregation_rules"].([]any)[1].(map[string]any)
	aggregation["endpoint_grouping"].(map[string]any)["missing_endpoint_behavior"] = "exclude"
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "fix010", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if run.State != RunStateAvailable || hasIssueCode(run.ValidationSummary.Issues, "aggregation_endpoint_missing") {
		t.Fatalf("exclude endpoint result state=%s issues=%#v", run.State, run.ValidationSummary.Issues)
	}
}

func TestGPFIX011Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-011")
	input := incidentGraphInput(t)
	relationship := input["source_relationships"].([]any)[0].(map[string]any)
	relationship["properties"].(map[string]any)["src_site"] = "remote"
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "fix011", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if run.State != RunStateAvailable || !hasIssueReasonCode(run.ValidationSummary.Issues, "endpoint_vertex_not_found") || hasEdgeKind(*run.GraphView, "site_logon_edge") {
		t.Fatalf("unmatched endpoint result state=%s issues=%#v graph=%#v", run.State, run.ValidationSummary.Issues, run.GraphView)
	}
}

func TestGPFIX012Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-012")
	input := incidentGraphInput(t)
	entityMapping := input["projection_config"].(map[string]any)["entity_mappings"].([]any)[0].(map[string]any)
	entityMapping["label_policy"] = "mapping_then_source"
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "fix012", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	schema := findVertexSchema(t, run.GraphView.SchemaRegistry, "host_vertex")
	if !schema.SourceLabelsPreserved || !containsString(schema.Labels, "asset") || !containsString(schema.Labels, "host") {
		t.Fatalf("host schema labels = %#v preserved=%v", schema.Labels, schema.SourceLabelsPreserved)
	}
}

func TestGPFIX013Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-013")
	input := incidentGraphInput(t)
	definitions := input["property_definitions"].([]any)
	definitions = append(definitions, map[string]any{"property_definition_id": "pd_wild", "target_scope": "vertex", "target_kind": "*", "source_field_path": "properties.hostname", "projected_key": "wild_name", "projected_type": "string"})
	input["property_definitions"] = definitions
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "fix013", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if !hasPropertySchema(run.GraphView.SchemaRegistry, "vertex", "host_vertex", "wild_name") || !hasPropertySchema(run.GraphView.SchemaRegistry, "vertex", "site_vertex", "wild_name") {
		t.Fatalf("wildcard property schema = %#v", run.GraphView.SchemaRegistry.PropertyKeys)
	}
}

func TestGPFIX024Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-024")
	input := minimalInput(t, "fix024")
	input["projection_config"].(map[string]any)["retention_policy"] = map[string]any{"retention_count": 1, "unexpected": true}
	_, err := admitProjectionInput(mustJSON(t, input), admitOptions{Operation: "create_projection"})
	var operationError *OperationError
	if !errors.As(err, &operationError) || operationError.ReasonCode != "unknown_member" || operationError.Field != "$.projection_config.retention_policy.unexpected" {
		t.Fatalf("nested unknown error = %#v", err)
	}
}

func TestGPFIX025Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-025")
	input := minimalInput(t, "fix025")
	input["source_snapshot_id"] = " invalid"
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "fix025", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if run.ProjectionRunID == "" || run.State != RunStateFailed || !hasIssueCode(run.ValidationSummary.Issues, "invalid_input_shape") {
		t.Fatalf("scalar violation run = %#v", run)
	}
}

func TestGPFIX026Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-026")
	input := incidentGraphInput(t)
	input["source_entities"] = append(input["source_entities"].([]any), map[string]any{"source_entity_id": "host1", "source_entity_kind": "host"})
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "fix026", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	if run.State != RunStateFailed || !hasIssueCode(run.ValidationSummary.Issues, "duplicate_identifier") {
		t.Fatalf("duplicate identifier run = %#v", run)
	}
}

func TestGPFIX027Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-027")
	input := incidentGraphInput(t)
	config := input["projection_config"].(map[string]any)
	config["metadata_mappings"] = []any{map[string]any{"metadata_mapping_id": "mm_site", "target_scope": "vertex", "target_kind": "host_vertex", "source_field_path": "properties.site", "projected_metadata_key": "site_meta", "projected_type": "string"}}
	config["aggregation_rules"] = append(config["aggregation_rules"].([]any), map[string]any{"aggregation_rule_id": "agg_projected_by_site", "target_scope": "vertex", "input_scope": "projected_vertex", "input_kind": "host_vertex", "projected_kind": "site_from_projected", "grouping_keys": []any{"projected.metadata.site_meta"}, "missing_grouping_key_behavior": "error"})
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "fix027", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatal(err)
	}
	findVertexKind(t, *run.GraphView, "site_from_projected")
}

func TestGPFIX028Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-028")
	if validFieldPath("projected.metadata.mapping_rule_id") {
		t.Fatal("system metadata field path accepted")
	}
}

func TestGPFIX034Remediation(t *testing.T) {
	assertFixtureLoaded(t, "GP-FIX-034")
	previousLimits := graphProjectionLimits
	graphProjectionLimits.MaxSourceEntities = 0
	graphProjectionLimits.MaxSourceRelationships = 0
	graphProjectionLimits.MaxEntityFilters = 0
	graphProjectionLimits.MaxRelationshipFilters = 0
	graphProjectionLimits.MaxDeclaredSourceEntityKinds = 0
	graphProjectionLimits.MaxDeclaredSourceRelationshipKinds = 0
	graphProjectionLimits.MaxEntityMappings = 0
	graphProjectionLimits.MaxRelationshipMappings = 0
	graphProjectionLimits.MaxPropertyDefinitions = 0
	graphProjectionLimits.MaxMetadataMappings = 0
	graphProjectionLimits.MaxAggregationRules = 0
	graphProjectionLimits.MaxDefaultVertexLabels = 0
	graphProjectionLimits.MaxDefaultEdgeLabels = 0
	graphProjectionLimits.MaxMappingLabelsPerRule = 0
	graphProjectionLimits.MaxMappingPropertyKeyRefs = 0
	graphProjectionLimits.MaxLabelsPerSourceItem = 0
	graphProjectionLimits.MaxStringPropertyValueLength = 1
	graphProjectionLimits.MaxMetadataKeysPerObject = 0
	graphProjectionLimits.MaxPropertiesPerObject = 0
	graphProjectionLimits.MaxCustomConfigKeys = 0
	defer func() { graphProjectionLimits = previousLimits }()

	run := ProjectionRun{GraphViewID: "gv_test", ProjectionRunID: "gpr_test"}
	run.Request.SourceEntities = []SourceEntity{{SourceEntityID: "host1", SourceEntityKind: "host", Labels: []string{"xx"}, Properties: map[string]any{"name": "xx"}, Metadata: map[string]any{"m": "xx"}}}
	run.Request.SourceRelationships = []SourceRelationship{{SourceRelationshipID: "rel1", SourceRelationshipKind: "rel", Labels: []string{"xx"}, Properties: map[string]any{"name": "xx"}, Metadata: map[string]any{"m": "xx"}}}
	run.Request.Filters.EntityFilters = []FilterPredicate{{FieldPath: "kind", Operator: "equals", Value: "host", HasValue: true}}
	run.Request.Filters.RelationshipFilters = []FilterPredicate{{FieldPath: "kind", Operator: "equals", Value: "rel", HasValue: true}}
	run.Request.ProjectionConfig.DeclaredSourceEntityKinds = []string{"host"}
	run.Request.ProjectionConfig.DeclaredSourceRelationshipKinds = []string{"rel"}
	run.Request.ProjectionConfig.EntityMappings = []EntityMapping{{MappingRuleID: "map_host", MappingLabels: []string{"label"}, RequiredPropertyKeys: []string{"name"}, OptionalPropertyKeys: []string{"site"}}}
	run.Request.RelationshipMappings = []RelationshipMapping{{MappingRuleID: "map_rel", MappingLabels: []string{"label"}, RequiredPropertyKeys: []string{"name"}, OptionalPropertyKeys: []string{"site"}}}
	run.Request.PropertyDefinitions = []PropertyDefinition{{PropertyDefinitionID: "pd", ProjectedKey: "name"}}
	run.Request.ProjectionConfig.MetadataMappings = []MetadataMapping{{MetadataMappingID: "mm", ProjectedMetadataKey: "m"}}
	run.Request.ProjectionConfig.AggregationRules = []AggregationRule{{AggregationRuleID: "agg"}}
	run.Request.ProjectionConfig.DefaultVertexLabels = []string{"v"}
	run.Request.ProjectionConfig.DefaultEdgeLabels = []string{"e"}
	run.Request.SourceMetadata = map[string]any{"case": "xx"}
	run.Request.ProjectionConfig.CustomConfig = map[string]any{"k": "v"}

	issues := admittedResourceLimitIssues(run)
	for _, key := range []string{"max_source_entities", "max_source_relationships", "max_entity_filters", "max_relationship_filters", "max_declared_source_entity_kinds", "max_declared_source_relationship_kinds", "max_entity_mappings", "max_relationship_mappings", "max_property_definitions", "max_metadata_mappings", "max_aggregation_rules", "max_default_vertex_labels", "max_default_edge_labels", "max_mapping_labels_per_rule", "max_mapping_property_key_refs", "max_labels_per_source_item", "max_string_property_value_length", "max_metadata_keys_per_object", "max_properties_per_object", "max_custom_config_keys"} {
		if !hasIssueLimitKey(issues, key) {
			t.Fatalf("missing resource limit key %s in %#v", key, issues)
		}
	}
}

func assertOperationError(t *testing.T, err error, code, reason string) {
	t.Helper()
	var operationError *OperationError
	if !errors.As(err, &operationError) || operationError.Code != code || operationError.ReasonCode != reason {
		t.Fatalf("operation error = %#v want %s/%s", err, code, reason)
	}
}

func hasIssueCode(issues []ValidationIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func hasIssueLimitKey(issues []ValidationIssue, key string) bool {
	for _, issue := range issues {
		if issue.Details["limit_key"] == key {
			return true
		}
	}
	return false
}

func findVertexKind(t *testing.T, graph GraphView, kind string) Vertex {
	t.Helper()
	for _, vertex := range graph.Vertices {
		if vertex.VertexKind == kind {
			return vertex
		}
	}
	t.Fatalf("missing vertex kind %s in %#v", kind, graph.Vertices)
	return Vertex{}
}

func hasEdgeKind(graph GraphView, kind string) bool {
	for _, edge := range graph.Edges {
		if edge.EdgeKind == kind {
			return true
		}
	}
	return false
}

func hasIssueReasonCode(issues []ValidationIssue, reasonCode string) bool {
	for _, issue := range issues {
		if issue.Details["reason_code"] == reasonCode {
			return true
		}
	}
	return false
}

func findVertexSchema(t *testing.T, registry SchemaRegistry, kind string) VertexKindSchema {
	t.Helper()
	for _, schema := range registry.VertexKinds {
		if schema.VertexKind == kind {
			return schema
		}
	}
	t.Fatalf("missing vertex schema %s in %#v", kind, registry.VertexKinds)
	return VertexKindSchema{}
}

func hasPropertySchema(registry SchemaRegistry, scope, kind, key string) bool {
	for _, schema := range registry.PropertyKeys {
		if schema.TargetScope == scope && schema.TargetKind == kind && schema.ProjectedKey == key {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func mustJSONText(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
