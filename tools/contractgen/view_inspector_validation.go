package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	viewInspectorRegistrySchemaID = "cartulary.view_inspector_registry.v1"
	viewInspectorRegistryID       = "cartulary.view_inspector.base.v1"
)

type viewInspectorRegistry struct {
	Schema              string                            `json:"$schema"`
	SchemaID            string                            `json:"schema_id"`
	RegistryID          string                            `json:"registry_id"`
	Note                string                            `json:"note"`
	Vocabularies        viewInspectorVocabularies         `json:"vocabularies"`
	ViewFeatureKeys     map[string][]string               `json:"view_feature_keys"`
	SpecializedFeatures []viewInspectorSpecializedFeature `json:"specialized_features"`
	ResolutionPolicy    viewInspectorResolutionPolicy     `json:"resolution_policy"`
}

type viewInspectorVocabularies struct {
	Panels                 []string `json:"panels"`
	RouteKinds             []string `json:"route_kinds"`
	RouteOwners            []string `json:"route_owners"`
	IncidentRoles          []string `json:"incident_roles"`
	DisabledConditions     []string `json:"disabled_conditions"`
	SuccessResultBehaviors []string `json:"success_result_behaviors"`
	FailureResultBehaviors []string `json:"failure_result_behaviors"`
	SeedSourceKinds        []string `json:"seed_source_kinds"`
}

type viewInspectorSpecializedFeature struct {
	ViewSchemaID          string   `json:"view_schema_id"`
	FeatureGroupKey       string   `json:"feature_group_key"`
	PanelID               string   `json:"panel_id"`
	RouteBindingKind      string   `json:"route_binding_kind"`
	RouteBindingOwner     string   `json:"route_binding_owner"`
	ActionKey             string   `json:"action_key"`
	MinimumIncidentRole   *string  `json:"minimum_incident_role"`
	Mutates               bool     `json:"mutates"`
	RequiresConfirmation  bool     `json:"requires_confirmation"`
	SeedBindings          []any    `json:"seed_bindings"`
	DisabledWhen          []string `json:"disabled_when"`
	SuccessResultBehavior string   `json:"success_result_behavior"`
	FailureResultBehavior string   `json:"failure_result_behavior"`
}

type viewInspectorResolutionPolicy struct {
	ExactPrecedesWildcard          bool     `json:"exact_precedes_wildcard"`
	RecordPatchExcludedFeatureKeys []string `json:"record_patch_excluded_feature_keys"`
}

func validateViewInspectorRegistryShape(value any, relativePath string) error {
	object, err := asObject(value, relativePath)
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, stringSet(
		"$schema",
		"schema_id",
		"registry_id",
		"note",
		"vocabularies",
		"view_feature_keys",
		"specialized_features",
		"resolution_policy",
	), relativePath); err != nil {
		return err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s: %w", relativePath, err)
	}
	var registry viewInspectorRegistry
	if err := json.Unmarshal(encoded, &registry); err != nil {
		return fmt.Errorf("decode %s: %w", relativePath, err)
	}
	return validateViewInspectorRegistry(registry, relativePath)
}

func loadViewInspectorRegistry(root string) (viewInspectorRegistry, error) {
	path := filepath.Join(root, "contracts", "view-inspector", "index.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return viewInspectorRegistry{}, fmt.Errorf("read view Inspector registry: %w", err)
	}
	value, err := decodeContract(raw)
	if err != nil {
		return viewInspectorRegistry{}, fmt.Errorf("decode view Inspector registry: %w", err)
	}
	if err := validateViewInspectorRegistryShape(value, "contracts/view-inspector/index.json"); err != nil {
		return viewInspectorRegistry{}, err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return viewInspectorRegistry{}, fmt.Errorf("encode view Inspector registry: %w", err)
	}
	var registry viewInspectorRegistry
	if err := json.Unmarshal(encoded, &registry); err != nil {
		return viewInspectorRegistry{}, fmt.Errorf("materialize view Inspector registry: %w", err)
	}
	return registry, nil
}

func validateViewInspectorRegistry(registry viewInspectorRegistry, label string) error {
	if registry.Schema != contractDraft202012Schema {
		return fmt.Errorf("%s.$schema must be %s", label, contractDraft202012Schema)
	}
	if registry.SchemaID != viewInspectorRegistrySchemaID {
		return fmt.Errorf("%s.schema_id must be %s", label, viewInspectorRegistrySchemaID)
	}
	if registry.RegistryID != viewInspectorRegistryID {
		return fmt.Errorf("%s.registry_id must be %s", label, viewInspectorRegistryID)
	}
	if strings.TrimSpace(registry.Note) == "" {
		return fmt.Errorf("%s.note is required", label)
	}
	vocabularies := map[string][]string{
		"panels":                   registry.Vocabularies.Panels,
		"route_kinds":              registry.Vocabularies.RouteKinds,
		"route_owners":             registry.Vocabularies.RouteOwners,
		"incident_roles":           registry.Vocabularies.IncidentRoles,
		"disabled_conditions":      registry.Vocabularies.DisabledConditions,
		"success_result_behaviors": registry.Vocabularies.SuccessResultBehaviors,
		"failure_result_behaviors": registry.Vocabularies.FailureResultBehaviors,
		"seed_source_kinds":        registry.Vocabularies.SeedSourceKinds,
	}
	for name, values := range vocabularies {
		if err := validateInspectorTokenList(values, label+".vocabularies."+name, false); err != nil {
			return err
		}
	}
	if len(registry.ViewFeatureKeys) == 0 {
		return fmt.Errorf("%s.view_feature_keys must not be empty", label)
	}
	for viewSchemaID, featureKeys := range registry.ViewFeatureKeys {
		if !strings.HasPrefix(viewSchemaID, "cartulary.view.") || !isInspectorFeatureKey(viewSchemaID) {
			return fmt.Errorf("%s.view_feature_keys contains invalid view_schema_id %s", label, viewSchemaID)
		}
		if err := validateInspectorTokenList(featureKeys, label+".view_feature_keys."+viewSchemaID, true); err != nil {
			return err
		}
	}
	if !registry.ResolutionPolicy.ExactPrecedesWildcard {
		return fmt.Errorf("%s.resolution_policy.exact_precedes_wildcard must be true", label)
	}
	if err := validateInspectorTokenList(
		registry.ResolutionPolicy.RecordPatchExcludedFeatureKeys,
		label+".resolution_policy.record_patch_excluded_feature_keys",
		true,
	); err != nil {
		return err
	}
	excluded := stringSet(registry.ResolutionPolicy.RecordPatchExcludedFeatureKeys...)
	specializedKeys := map[string]struct{}{}
	panels := stringSet(registry.Vocabularies.Panels...)
	kinds := stringSet(registry.Vocabularies.RouteKinds...)
	owners := stringSet(registry.Vocabularies.RouteOwners...)
	roles := stringSet(registry.Vocabularies.IncidentRoles...)
	disabled := stringSet(registry.Vocabularies.DisabledConditions...)
	success := stringSet(registry.Vocabularies.SuccessResultBehaviors...)
	failure := stringSet(registry.Vocabularies.FailureResultBehaviors...)
	for index, feature := range registry.SpecializedFeatures {
		featureLabel := fmt.Sprintf("%s.specialized_features[%d]", label, index+1)
		identity := feature.ViewSchemaID + "\x00" + feature.FeatureGroupKey
		if _, duplicate := specializedKeys[identity]; duplicate {
			return fmt.Errorf("%s duplicates specialized feature %s for %s", featureLabel, feature.FeatureGroupKey, feature.ViewSchemaID)
		}
		specializedKeys[identity] = struct{}{}
		keys, ok := registry.ViewFeatureKeys[feature.ViewSchemaID]
		if !ok || !containsString(keys, feature.FeatureGroupKey) {
			return fmt.Errorf("%s is absent from view_feature_keys", featureLabel)
		}
		if feature.ActionKey != feature.FeatureGroupKey {
			return fmt.Errorf("%s.action_key must equal feature_group_key", featureLabel)
		}
		if _, ok := excluded[feature.FeatureGroupKey]; !ok {
			return fmt.Errorf("%s must be excluded from record patch", featureLabel)
		}
		if _, ok := panels[feature.PanelID]; !ok {
			return fmt.Errorf("%s.panel_id references unknown panel %s", featureLabel, feature.PanelID)
		}
		if _, ok := kinds[feature.RouteBindingKind]; !ok || feature.RouteBindingKind == "record_patch" {
			return fmt.Errorf("%s.route_binding_kind must be a declared specialized kind", featureLabel)
		}
		if _, ok := owners[feature.RouteBindingOwner]; !ok || feature.RouteBindingOwner == "record_patch_route" {
			return fmt.Errorf("%s.route_binding_owner must be a declared specialized owner", featureLabel)
		}
		if feature.MinimumIncidentRole != nil {
			if _, ok := roles[*feature.MinimumIncidentRole]; !ok {
				return fmt.Errorf("%s.minimum_incident_role references unknown role %s", featureLabel, *feature.MinimumIncidentRole)
			}
		}
		if len(feature.SeedBindings) != 0 {
			return fmt.Errorf("%s.seed_bindings must be empty", featureLabel)
		}
		if err := requireKnownStrings(feature.DisabledWhen, disabled, featureLabel+".disabled_when"); err != nil {
			return err
		}
		if _, ok := success[feature.SuccessResultBehavior]; !ok {
			return fmt.Errorf("%s.success_result_behavior is unknown", featureLabel)
		}
		if _, ok := failure[feature.FailureResultBehavior]; !ok {
			return fmt.Errorf("%s.failure_result_behavior is unknown", featureLabel)
		}
	}
	for featureKey := range excluded {
		found := false
		for _, feature := range registry.SpecializedFeatures {
			if feature.FeatureGroupKey == featureKey {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s excludes %s from record patch without a specialized binding", label, featureKey)
		}
	}
	return nil
}

func validateInspectorTokenList(values []string, label string, dotted bool) error {
	if len(values) == 0 {
		return fmt.Errorf("%s must not be empty", label)
	}
	seen := map[string]struct{}{}
	for index, value := range values {
		valid := isInspectorFeatureKey(value)
		if !dotted {
			valid = valid && !strings.Contains(value, ".")
		}
		if !valid {
			return fmt.Errorf("%s[%d] contains invalid token %s", label, index+1, value)
		}
		if _, duplicate := seen[value]; duplicate {
			return fmt.Errorf("%s contains duplicate token %s", label, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateViewSchemasAgainstInspectorRegistry(
	discovered map[string]map[string]any,
	registry viewInspectorRegistry,
) error {
	byID := map[string]map[string]any{}
	for path, schema := range discovered {
		viewSchemaID, _ := requiredString(schema, "view_schema_id", path)
		byID[viewSchemaID] = schema
	}
	if len(byID) != len(registry.ViewFeatureKeys) {
		return fmt.Errorf("view Inspector registry contains %d view entries, expected %d", len(registry.ViewFeatureKeys), len(byID))
	}
	viewIDs := make([]string, 0, len(byID))
	for viewSchemaID := range byID {
		viewIDs = append(viewIDs, viewSchemaID)
	}
	sort.Strings(viewIDs)
	for _, viewSchemaID := range viewIDs {
		schema := byID[viewSchemaID]
		expected, ok := registry.ViewFeatureKeys[viewSchemaID]
		if !ok {
			return fmt.Errorf("view Inspector registry missing %s", viewSchemaID)
		}
		config, _ := asObject(schema["inspector_config"], viewSchemaID+".inspector_config")
		if err := validateInspectorConfigAgainstRegistry(config, viewSchemaID, expected, registry); err != nil {
			return err
		}
	}
	return nil
}

func validateInspectorConfigAgainstRegistry(
	config map[string]any,
	viewSchemaID string,
	expectedFeatureKeys []string,
	registry viewInspectorRegistry,
) error {
	label := viewSchemaID + ".inspector_config"
	panels := stringSet(registry.Vocabularies.Panels...)
	panelRows, _ := objectArray(config["panels"], label+".panels")
	for index, panel := range panelRows {
		panelID, _ := requiredString(panel, "panel_id", fmt.Sprintf("%s.panels[%d]", label, index+1))
		if _, ok := panels[panelID]; !ok {
			return fmt.Errorf("%s.panels[%d].panel_id references unknown panel %s", label, index+1, panelID)
		}
	}
	groups, _ := objectArrayAllowEmpty(config["feature_groups"], label+".feature_groups")
	if len(groups) != len(expectedFeatureKeys) {
		return fmt.Errorf("%s.feature_groups must contain exactly %d ordered feature groups, got %d", label, len(expectedFeatureKeys), len(groups))
	}
	specialized := map[string]viewInspectorSpecializedFeature{}
	for _, feature := range registry.SpecializedFeatures {
		if feature.ViewSchemaID == viewSchemaID {
			specialized[feature.FeatureGroupKey] = feature
		}
	}
	for index, group := range groups {
		groupLabel := fmt.Sprintf("%s.feature_groups[%d]", label, index+1)
		featureKey, _ := requiredString(group, "feature_group_key", groupLabel)
		if featureKey != expectedFeatureKeys[index] {
			return fmt.Errorf("%s.feature_group_key must be %s, got %s", groupLabel, expectedFeatureKeys[index], featureKey)
		}
		if err := validateInspectorGroupVocabulary(group, groupLabel, registry); err != nil {
			return err
		}
		if exact, ok := specialized[featureKey]; ok {
			if err := validateSpecializedInspectorGroup(group, groupLabel, exact); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateInspectorGroupVocabulary(group map[string]any, label string, registry viewInspectorRegistry) error {
	panelID, _ := requiredString(group, "panel_id", label)
	if _, ok := stringSet(registry.Vocabularies.Panels...)[panelID]; !ok {
		return fmt.Errorf("%s.panel_id references unknown panel %s", label, panelID)
	}
	if role := group["minimum_incident_role"]; role != nil {
		roleText, _ := role.(string)
		if _, ok := stringSet(registry.Vocabularies.IncidentRoles...)[roleText]; !ok {
			return fmt.Errorf("%s.minimum_incident_role references unknown role %s", label, roleText)
		}
	}
	route, _ := asObject(group["route_binding"], label+".route_binding")
	kind, _ := requiredString(route, "kind", label+".route_binding")
	if _, ok := stringSet(registry.Vocabularies.RouteKinds...)[kind]; !ok {
		return fmt.Errorf("%s.route_binding.kind references unknown kind %s", label, kind)
	}
	owner, _ := requiredString(route, "owner", label+".route_binding")
	if _, ok := stringSet(registry.Vocabularies.RouteOwners...)[owner]; !ok {
		return fmt.Errorf("%s.route_binding.owner references unknown owner %s", label, owner)
	}
	conditions, _ := stringArray(group["disabled_when"], label+".disabled_when", false)
	if err := requireKnownStrings(conditions, stringSet(registry.Vocabularies.DisabledConditions...), label+".disabled_when"); err != nil {
		return err
	}
	success, _ := requiredString(group, "success_result_behavior", label)
	if _, ok := stringSet(registry.Vocabularies.SuccessResultBehaviors...)[success]; !ok {
		return fmt.Errorf("%s.success_result_behavior references unknown behavior %s", label, success)
	}
	failure, _ := requiredString(group, "failure_result_behavior", label)
	if _, ok := stringSet(registry.Vocabularies.FailureResultBehaviors...)[failure]; !ok {
		return fmt.Errorf("%s.failure_result_behavior references unknown behavior %s", label, failure)
	}
	bindings, _ := objectArrayAllowEmpty(group["seed_bindings"], label+".seed_bindings")
	for index, binding := range bindings {
		source, _ := asObject(binding["source"], fmt.Sprintf("%s.seed_bindings[%d].source", label, index+1))
		kind, _ := requiredString(source, "kind", label)
		if _, ok := stringSet(registry.Vocabularies.SeedSourceKinds...)[kind]; !ok {
			return fmt.Errorf("%s.seed_bindings[%d].source.kind references unknown kind %s", label, index+1, kind)
		}
		if kind == "selected_field_value" {
			if _, err := requiredString(source, "source_field_key", fmt.Sprintf("%s.seed_bindings[%d].source", label, index+1)); err != nil {
				return err
			}
		}
		if kind == "literal" {
			if _, ok := source["value"]; !ok {
				return fmt.Errorf("%s.seed_bindings[%d].source.value is required for literal", label, index+1)
			}
		}
	}
	return nil
}

func validateSpecializedInspectorGroup(group map[string]any, label string, expected viewInspectorSpecializedFeature) error {
	checks := map[string]string{
		"panel_id":                expected.PanelID,
		"success_result_behavior": expected.SuccessResultBehavior,
		"failure_result_behavior": expected.FailureResultBehavior,
	}
	for key, want := range checks {
		got, _ := requiredString(group, key, label)
		if got != want {
			return fmt.Errorf("%s.%s must be %s, got %s", label, key, want, got)
		}
	}
	if got, _ := requiredBool(group, "mutates", label); got != expected.Mutates {
		return fmt.Errorf("%s.mutates must be %t", label, expected.Mutates)
	}
	if got, _ := requiredBool(group, "requires_confirmation", label); got != expected.RequiresConfirmation {
		return fmt.Errorf("%s.requires_confirmation must be %t", label, expected.RequiresConfirmation)
	}
	if expected.MinimumIncidentRole == nil {
		if group["minimum_incident_role"] != nil {
			return fmt.Errorf("%s.minimum_incident_role must be null", label)
		}
	} else if group["minimum_incident_role"] != *expected.MinimumIncidentRole {
		return fmt.Errorf("%s.minimum_incident_role must be %s", label, *expected.MinimumIncidentRole)
	}
	route, _ := asObject(group["route_binding"], label+".route_binding")
	for key, want := range map[string]string{
		"kind":       expected.RouteBindingKind,
		"owner":      expected.RouteBindingOwner,
		"action_key": expected.ActionKey,
	} {
		got, _ := requiredString(route, key, label+".route_binding")
		if got != want {
			return fmt.Errorf("%s.route_binding.%s must be %s, got %s", label, key, want, got)
		}
	}
	if _, present := route["target_view_schema_id"]; present {
		return fmt.Errorf("%s.route_binding.target_view_schema_id must be omitted", label)
	}
	seedBindings, _ := objectArrayAllowEmpty(group["seed_bindings"], label+".seed_bindings")
	if len(seedBindings) != 0 {
		return fmt.Errorf("%s.seed_bindings must be empty", label)
	}
	disabled, _ := stringArray(group["disabled_when"], label+".disabled_when", false)
	if strings.Join(disabled, "\x00") != strings.Join(expected.DisabledWhen, "\x00") {
		return fmt.Errorf("%s.disabled_when does not match the specialized registry", label)
	}
	return nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
