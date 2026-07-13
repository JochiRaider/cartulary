package graphprojection

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/fixturetest"
)

type unitFixtureExecutor struct{}

func (unitFixtureExecutor) ExecuteFixtureStep(manifest fixturetest.Manifest, step fixturetest.Step, input []byte) (fixturetest.StepExecution, error) {
	var artifact []byte
	var err error
	switch step.Operation {
	case "project_ephemeral":
		artifact, err = executeEphemeralFixture(manifest, input)
	case "admit_projection":
		artifact, err = executeAdmissionFixture(manifest, input)
	case "canonical_json":
		var value any
		if err = json.Unmarshal(input, &value); err == nil {
			artifact, err = canonicalJSON(value)
		}
	case "classify_integers":
		artifact, err = classifyStrings(input, validFiniteInteger)
	case "classify_timestamps":
		artifact, err = classifyStrings(input, func(value string) bool {
			_, parseErr := parseTimestamp(value)
			return parseErr == nil
		})
	case "classify_identifiers":
		artifact, err = classifyStrings(input, validIdentifier)
	case "classify_field_paths":
		artifact, err = classifyStrings(input, validFieldPath)
	case "sort_canonical_values":
		var values []string
		if err = json.Unmarshal(input, &values); err == nil {
			for index := range values {
				values[index] = canonicalValueKey(values[index])
			}
			sort.Strings(values)
			artifact, err = json.Marshal(values)
		}
	case "summarize_validation_issues":
		artifact, err = executeValidationSummaryFixture(input)
	case "evaluate_property_candidates":
		artifact, err = executeCandidateFixture(input)
	case "merge_property_values":
		artifact, err = executeMergeFixture(input)
	case "project_ephemeral_with_limits":
		artifact, err = executeLimitedEphemeralFixture(manifest, input)
	default:
		return fixturetest.StepExecution{}, errors.New("unsupported unit fixture operation: " + step.Operation)
	}
	if err != nil {
		return fixturetest.StepExecution{}, err
	}
	return fixturetest.StepExecution{Artifact: artifact, StateEffectMode: "no_retained_state_change"}, nil
}

func executeLimitedEphemeralFixture(manifest fixturetest.Manifest, input []byte) ([]byte, error) {
	var probe struct {
		ProjectionInput json.RawMessage `json:"projection_input"`
		Limits          map[string]int  `json:"limits"`
	}
	if err := json.Unmarshal(input, &probe); err != nil {
		return nil, err
	}
	previous := graphProjectionLimits
	defer func() { graphProjectionLimits = previous }()
	for key, value := range probe.Limits {
		if err := setFixtureResourceLimit(key, value); err != nil {
			return nil, err
		}
	}
	return executeEphemeralFixture(manifest, probe.ProjectionInput)
}

func setFixtureResourceLimit(key string, value int) error {
	switch key {
	case "max_source_entities":
		graphProjectionLimits.MaxSourceEntities = value
	case "max_source_relationships":
		graphProjectionLimits.MaxSourceRelationships = value
	case "max_entity_filters":
		graphProjectionLimits.MaxEntityFilters = value
	case "max_relationship_filters":
		graphProjectionLimits.MaxRelationshipFilters = value
	case "max_declared_source_entity_kinds":
		graphProjectionLimits.MaxDeclaredSourceEntityKinds = value
	case "max_declared_source_relationship_kinds":
		graphProjectionLimits.MaxDeclaredSourceRelationshipKinds = value
	case "max_entity_mappings":
		graphProjectionLimits.MaxEntityMappings = value
	case "max_relationship_mappings":
		graphProjectionLimits.MaxRelationshipMappings = value
	case "max_property_definitions":
		graphProjectionLimits.MaxPropertyDefinitions = value
	case "max_metadata_mappings":
		graphProjectionLimits.MaxMetadataMappings = value
	case "max_aggregation_rules":
		graphProjectionLimits.MaxAggregationRules = value
	case "max_default_vertex_labels":
		graphProjectionLimits.MaxDefaultVertexLabels = value
	case "max_default_edge_labels":
		graphProjectionLimits.MaxDefaultEdgeLabels = value
	case "max_mapping_labels_per_rule":
		graphProjectionLimits.MaxMappingLabelsPerRule = value
	case "max_mapping_property_key_refs":
		graphProjectionLimits.MaxMappingPropertyKeyRefs = value
	case "max_labels_per_source_item":
		graphProjectionLimits.MaxLabelsPerSourceItem = value
	case "max_string_property_value_length":
		graphProjectionLimits.MaxStringPropertyValueLength = value
	case "max_metadata_keys_per_object":
		graphProjectionLimits.MaxMetadataKeysPerObject = value
	case "max_properties_per_object":
		graphProjectionLimits.MaxPropertiesPerObject = value
	case "max_custom_config_keys":
		graphProjectionLimits.MaxCustomConfigKeys = value
	default:
		return errors.New("unknown fixture resource limit: " + key)
	}
	return nil
}

type fixtureIssueProbe struct {
	Severity   string         `json:"severity"`
	Code       string         `json:"code"`
	TargetKind string         `json:"target_kind"`
	TargetID   string         `json:"target_id"`
	Field      *string        `json:"field"`
	Details    map[string]any `json:"details"`
}

func executeValidationSummaryFixture(input []byte) ([]byte, error) {
	var probe struct {
		GraphViewID         string              `json:"graph_view_id"`
		ProjectionRunID     string              `json:"projection_run_id"`
		MaxValidationIssues int                 `json:"max_validation_issues"`
		Issues              []fixtureIssueProbe `json:"issues"`
	}
	if err := json.Unmarshal(input, &probe); err != nil {
		return nil, err
	}
	previous := graphProjectionLimits
	graphProjectionLimits.MaxValidationIssues = probe.MaxValidationIssues
	defer func() { graphProjectionLimits = previous }()
	run := ProjectionRun{GraphViewID: probe.GraphViewID, ProjectionRunID: probe.ProjectionRunID}
	issues := make([]ValidationIssue, 0, len(probe.Issues))
	for _, issue := range probe.Issues {
		var field any
		if issue.Field != nil {
			field = *issue.Field
		}
		issues = append(issues, run.issue(issue.Severity, issue.Code, issue.TargetKind, issue.TargetID, field, issue.Details))
	}
	return json.Marshal(validationSummaryResource(validationSummary(run, issues)))
}

func executeCandidateFixture(input []byte) ([]byte, error) {
	var probes []struct {
		Definition struct {
			ProjectedType      string `json:"projected_type"`
			DefaultValue       any    `json:"default_value"`
			HasDefaultValue    bool   `json:"has_default_value"`
			MissingBehavior    string `json:"missing_behavior"`
			SourceNullBehavior string `json:"source_null_behavior"`
			NullOutputPolicy   string `json:"null_output_policy"`
		} `json:"definition"`
		Value   any  `json:"value"`
		Present bool `json:"present"`
	}
	if err := json.Unmarshal(input, &probes); err != nil {
		return nil, err
	}
	results := make([]map[string]any, 0, len(probes))
	for _, probe := range probes {
		definition := candidateDefinition{ProjectedType: probe.Definition.ProjectedType, DefaultValue: probe.Definition.DefaultValue, HasDefaultValue: probe.Definition.HasDefaultValue, MissingBehavior: probe.Definition.MissingBehavior, SourceNullBehavior: probe.Definition.SourceNullBehavior, NullOutputPolicy: probe.Definition.NullOutputPolicy}
		value, include, code := evaluateCandidate(definition, probe.Value, probe.Present)
		results = append(results, map[string]any{"value": value, "include": include, "code": code})
	}
	return json.Marshal(results)
}

func executeMergeFixture(input []byte) ([]byte, error) {
	var probe struct {
		Behavior string `json:"behavior"`
		Values   []any  `json:"values"`
	}
	if err := json.Unmarshal(input, &probe); err != nil {
		return nil, err
	}
	value, include, conflict := mergeValues(probe.Behavior, probe.Values)
	return json.Marshal(map[string]any{"value": value, "include": include, "conflict": conflict})
}

func executeEphemeralFixture(manifest fixturetest.Manifest, input []byte) ([]byte, error) {
	now, err := time.Parse(time.RFC3339Nano, manifest.Determinism.Clock)
	if err != nil {
		return nil, err
	}
	service := NewService(ServiceOptions{
		Now:      func() time.Time { return now },
		NewNonce: func() (string, error) { return manifest.Determinism.Nonce, nil },
	})
	result, err := service.ProjectEphemeral(context.Background(), EphemeralProjectionRequest{ProjectionInput: input})
	if err != nil {
		return fixtureErrorArtifact(err)
	}
	return json.Marshal(result.Resource())
}

func executeAdmissionFixture(manifest fixturetest.Manifest, input []byte) ([]byte, error) {
	now, err := time.Parse(time.RFC3339Nano, manifest.Determinism.Clock)
	if err != nil {
		return nil, err
	}
	run, err := admitProjectionInput(input, admitOptions{Operation: "project_ephemeral", ProjectionRunNonce: manifest.Determinism.Nonce, AcceptedAt: now})
	if err != nil {
		return fixtureErrorArtifact(err)
	}
	return json.Marshal(map[string]any{
		"graph_view_id":            run.GraphViewID,
		"projection_run_id":        run.ProjectionRunID,
		"projection_config_digest": run.ProjectionConfigDigest,
		"projection_source_digest": run.ProjectionSourceDigest,
		"state":                    run.State,
	})
}

func fixtureErrorArtifact(err error) ([]byte, error) {
	var lifecycleError *LifecycleError
	if errors.As(err, &lifecycleError) {
		return lifecycleError.EnvelopeJSON()
	}
	var queryError *QueryError
	if errors.As(err, &queryError) {
		return queryError.EnvelopeJSON()
	}
	return nil, err
}

func classifyStrings(input []byte, valid func(string) bool) ([]byte, error) {
	var values []string
	if err := json.Unmarshal(input, &values); err != nil {
		return nil, err
	}
	accepted := make([]string, 0, len(values))
	rejected := make([]string, 0, len(values))
	for _, value := range values {
		if valid(value) {
			accepted = append(accepted, value)
		} else {
			rejected = append(rejected, value)
		}
	}
	return json.Marshal(map[string]any{"accepted": accepted, "rejected": rejected})
}

func verifyUnitFixture(t testingT, fixtureID string) {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := fixturetest.RepoRoot(workingDirectory)
	if err != nil {
		t.Fatal(err)
	}
	if err := fixturetest.Verify(root, fixtureID, unitFixtureExecutor{}); err != nil {
		t.Fatalf("verify %s: %v", fixtureID, err)
	}
}

type testingT interface {
	Helper()
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}
