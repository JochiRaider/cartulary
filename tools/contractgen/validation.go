package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const contractDraft202012Schema = "https://json-schema.org/draft/2020-12/schema"

var (
	viewSchemaTopLevelKeys = stringSet(
		"$schema",
		"view_schema_id",
		"title",
		"surface_kind",
		"source_record_types",
		"technical_fields",
		"required_reference_pack_keys",
		"default_visible_fields",
		"default_hidden_fields",
		"default_sort",
		"sort_fields",
		"sort_null_order",
		"filter_fields",
		"synthetic_filter_predicates",
		"grouping_fields",
		"inline_create",
		"fields",
	)
	viewSchemaFieldKeys = stringSet(
		"field_key",
		"label",
		"default_hidden",
		"sortable",
		"header_sort_field_key",
		"filter_ops",
		"groupable",
		"read_kind",
		"write_kind",
		"conflict_resolution_class",
		"entity_binding_mode",
		"string_contract_id",
		"direct_scalar_contract_id",
		"direct_reference_contract_id",
		"clearable",
		"enum_values",
		"writable",
		"read_model",
		"write_target",
		"write_action",
		"create_writable",
	)
	viewSchemaIndexKeys      = stringSet("$schema", "registry_id", "note", "view_schemas")
	viewSchemaIndexEntryKeys = stringSet("view_schema_id", "title", "surface_kind", "source_record_types", "artifact_path")
	syntheticPredicateKeys   = stringSet("field_key", "label", "filter_ops")
	errorRegistryKeys        = stringSet("$schema", "registry_id", "note", "errors", "reason_registries")
	errorEntryKeys           = stringSet("code", "http_status", "summary")
	reasonRegistryEntryKeys  = stringSet("error_code", "reason_codes")
	reasonCodeEntryKeys      = stringSet("code", "summary")
	extensionRegistryKeys    = stringSet("$schema", "registry_id", "note", "profiles")
	extensionProfileKeys     = stringSet("profile_id", "route_families")
	wsIndexKeys              = stringSet("$schema", "$id", "title", "description", "type", "additionalProperties", "properties", "required")
)

func validateContractInput(familyDir, relativePath string, value any) error {
	switch familyDir {
	case "view-schemas":
		if relativePath == "index.json" {
			return validateViewSchemaIndexShape(value, relativePath)
		}
		return validateViewSchemaShape(value, relativePath)
	case "errors":
		if relativePath != "index.json" {
			return fmt.Errorf("unexpected errors artifact %s", relativePath)
		}
		return validateErrorRegistry(value)
	case "extensions":
		if relativePath != "index.json" {
			return fmt.Errorf("unexpected extensions artifact %s", relativePath)
		}
		return validateExtensionRegistry(value)
	case "ws":
		if relativePath != "index.schema.json" {
			return fmt.Errorf("unexpected websocket artifact %s", relativePath)
		}
		return validateWSIndex(value)
	default:
		return nil
	}
}

func validateContractFamily(root, familyDir string) error {
	if familyDir != "view-schemas" {
		return nil
	}
	baseDir := filepath.Join(root, "contracts", "view-schemas")
	contractRoot, err := os.OpenRoot(baseDir)
	if err != nil {
		return fmt.Errorf("open view schema root: %w", err)
	}
	defer contractRoot.Close()

	rawIndex, err := contractRoot.ReadFile("index.json")
	if err != nil {
		return fmt.Errorf("read view schema index: %w", err)
	}
	indexValue, err := decodeContract(rawIndex)
	if err != nil {
		return fmt.Errorf("decode view schema index: %w", err)
	}
	index, err := asObject(indexValue, "contracts/view-schemas/index.json")
	if err != nil {
		return err
	}
	entries, err := objectArray(index["view_schemas"], "view_schemas")
	if err != nil {
		return err
	}

	discovered := map[string]map[string]any{}
	if err := filepath.WalkDir(baseDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() == "index.json" || !strings.HasSuffix(entry.Name(), ".json") {
			return nil
		}
		relToBase, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		raw, err := contractRoot.ReadFile(relToBase)
		if err != nil {
			return err
		}
		decoded, err := decodeContract(raw)
		if err != nil {
			return err
		}
		object, err := asObject(decoded, filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		discovered[filepath.ToSlash(rel)] = object
		return nil
	}); err != nil {
		return fmt.Errorf("collect view schemas: %w", err)
	}

	indexed := map[string]string{}
	for entryIndex, entry := range entries {
		label := fmt.Sprintf("view_schemas[%d]", entryIndex+1)
		artifactPath, err := requiredString(entry, "artifact_path", label)
		if err != nil {
			return err
		}
		schema, ok := discovered[artifactPath]
		if !ok {
			return fmt.Errorf("%s.artifact_path references missing artifact %s", label, artifactPath)
		}
		viewSchemaID, err := requiredString(entry, "view_schema_id", label)
		if err != nil {
			return err
		}
		if previous, exists := indexed[viewSchemaID]; exists {
			return fmt.Errorf("duplicate view schema index id %s in %s and %s", viewSchemaID, previous, artifactPath)
		}
		indexed[viewSchemaID] = artifactPath
		if schemaID, err := requiredString(schema, "view_schema_id", artifactPath); err != nil {
			return err
		} else if schemaID != viewSchemaID {
			return fmt.Errorf("%s.view_schema_id must match %s", label, artifactPath)
		}
		for _, key := range []string{"title", "surface_kind"} {
			indexValue, err := requiredString(entry, key, label)
			if err != nil {
				return err
			}
			schemaValue, err := requiredString(schema, key, artifactPath)
			if err != nil {
				return err
			}
			if indexValue != schemaValue {
				return fmt.Errorf("%s.%s must match %s.%s", label, key, artifactPath, key)
			}
		}
		indexTypes, err := stringArray(entry["source_record_types"], label+".source_record_types", true)
		if err != nil {
			return err
		}
		schemaTypes, err := stringArray(schema["source_record_types"], artifactPath+".source_record_types", true)
		if err != nil {
			return err
		}
		if strings.Join(indexTypes, "\x00") != strings.Join(schemaTypes, "\x00") {
			return fmt.Errorf("%s.source_record_types must match %s.source_record_types", label, artifactPath)
		}
	}
	if len(indexed) != len(discovered) {
		missing := make([]string, 0, len(discovered))
		for path, schema := range discovered {
			id, _ := requiredString(schema, "view_schema_id", path)
			if _, ok := indexed[id]; !ok {
				missing = append(missing, path)
			}
		}
		sort.Strings(missing)
		return fmt.Errorf("view schema index missing artifacts: %s", strings.Join(missing, ", "))
	}
	return nil
}

func validateViewSchemaIndexShape(value any, relativePath string) error {
	object, err := asObject(value, relativePath)
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, viewSchemaIndexKeys, relativePath); err != nil {
		return err
	}
	if err := requireDraftSchema(object, relativePath); err != nil {
		return err
	}
	if _, err := requiredString(object, "registry_id", relativePath); err != nil {
		return err
	}
	if _, err := requiredString(object, "note", relativePath); err != nil {
		return err
	}
	entries, err := objectArray(object["view_schemas"], "view_schemas")
	if err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for index, entry := range entries {
		label := fmt.Sprintf("view_schemas[%d]", index+1)
		if err := requireAllowedKeys(entry, viewSchemaIndexEntryKeys, label); err != nil {
			return err
		}
		id, err := requiredString(entry, "view_schema_id", label)
		if err != nil {
			return err
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate view schema id %s", id)
		}
		seen[id] = struct{}{}
		if _, err := requiredString(entry, "title", label); err != nil {
			return err
		}
		if _, err := requireEnumString(entry, "surface_kind", label, "built_in_sheet", "system_view"); err != nil {
			return err
		}
		if _, err := stringArray(entry["source_record_types"], label+".source_record_types", true); err != nil {
			return err
		}
		artifactPath, err := requiredString(entry, "artifact_path", label)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(artifactPath, "contracts/view-schemas/") || !strings.HasSuffix(artifactPath, ".json") {
			return fmt.Errorf("%s.artifact_path must point to contracts/view-schemas/*.json", label)
		}
	}
	return nil
}

func validateViewSchemaShape(value any, relativePath string) error {
	object, err := asObject(value, relativePath)
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, viewSchemaTopLevelKeys, relativePath); err != nil {
		return err
	}
	if err := requireDraftSchema(object, relativePath); err != nil {
		return err
	}
	viewSchemaID, err := requiredString(object, "view_schema_id", relativePath)
	if err != nil {
		return err
	}
	if expected := strings.TrimSuffix(filepath.Base(relativePath), ".json"); expected != viewSchemaID {
		return fmt.Errorf("%s view_schema_id must match filename", relativePath)
	}
	if _, err := requiredString(object, "title", relativePath); err != nil {
		return err
	}
	if _, err := requireEnumString(object, "surface_kind", relativePath, "built_in_sheet", "system_view"); err != nil {
		return err
	}
	for _, key := range []string{
		"source_record_types",
		"technical_fields",
		"required_reference_pack_keys",
		"default_visible_fields",
		"default_hidden_fields",
		"sort_fields",
		"filter_fields",
		"grouping_fields",
	} {
		if _, err := stringArray(object[key], relativePath+"."+key, false); err != nil {
			return err
		}
	}
	if _, err := requireEnumString(object, "sort_null_order", relativePath, "last"); err != nil {
		return err
	}
	inlineCreate, err := asObject(object["inline_create"], relativePath+".inline_create")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(inlineCreate, stringSet("permits_zero_field_create"), relativePath+".inline_create"); err != nil {
		return err
	}
	if _, err := requiredBool(inlineCreate, "permits_zero_field_create", relativePath+".inline_create"); err != nil {
		return err
	}
	defaultSort, err := objectArray(object["default_sort"], relativePath+".default_sort")
	if err != nil {
		return err
	}
	fields, err := objectArray(object["fields"], relativePath+".fields")
	if err != nil {
		return err
	}
	fieldKeys := map[string]struct{}{}
	for index, field := range fields {
		label := fmt.Sprintf("%s.fields[%d]", relativePath, index+1)
		fieldKey, err := validateViewSchemaField(field, label)
		if err != nil {
			return err
		}
		if _, exists := fieldKeys[fieldKey]; exists {
			return fmt.Errorf("%s duplicate field_key %s", relativePath, fieldKey)
		}
		fieldKeys[fieldKey] = struct{}{}
	}
	technicalFields, err := stringArray(object["technical_fields"], relativePath+".technical_fields", false)
	if err != nil {
		return err
	}
	referenceableFields := make(map[string]struct{}, len(fieldKeys)+len(technicalFields))
	for key := range fieldKeys {
		referenceableFields[key] = struct{}{}
	}
	for _, key := range technicalFields {
		referenceableFields[key] = struct{}{}
	}
	for _, key := range []string{"default_visible_fields", "default_hidden_fields"} {
		values, err := stringArray(object[key], relativePath+"."+key, false)
		if err != nil {
			return err
		}
		if err := requireKnownStrings(values, referenceableFields, relativePath+"."+key); err != nil {
			return err
		}
	}
	for _, key := range []string{"sort_fields", "filter_fields", "grouping_fields"} {
		values, err := stringArray(object[key], relativePath+"."+key, false)
		if err != nil {
			return err
		}
		if err := requireKnownStrings(values, fieldKeys, relativePath+"."+key); err != nil {
			return err
		}
	}
	for index, entry := range defaultSort {
		label := fmt.Sprintf("%s.default_sort[%d]", relativePath, index+1)
		if err := requireAllowedKeys(entry, stringSet("field_key", "direction"), label); err != nil {
			return err
		}
		fieldKey, err := requiredString(entry, "field_key", label)
		if err != nil {
			return err
		}
		if _, ok := referenceableFields[fieldKey]; !ok {
			return fmt.Errorf("%s.field_key references unknown field %s", label, fieldKey)
		}
		if _, err := requireEnumString(entry, "direction", label, "asc", "desc"); err != nil {
			return err
		}
	}
	syntheticPredicates, err := objectArrayAllowEmpty(object["synthetic_filter_predicates"], relativePath+".synthetic_filter_predicates")
	if err != nil {
		return err
	}
	seenSyntheticPredicates := map[string]struct{}{}
	for index, predicate := range syntheticPredicates {
		label := fmt.Sprintf("%s.synthetic_filter_predicates[%d]", relativePath, index+1)
		if err := requireAllowedKeys(predicate, syntheticPredicateKeys, label); err != nil {
			return err
		}
		fieldKey, err := requiredString(predicate, "field_key", label)
		if err != nil {
			return err
		}
		if _, exists := seenSyntheticPredicates[fieldKey]; exists {
			return fmt.Errorf("%s duplicate synthetic predicate field_key %s", relativePath, fieldKey)
		}
		seenSyntheticPredicates[fieldKey] = struct{}{}
		if _, err := requiredString(predicate, "label", label); err != nil {
			return err
		}
		if _, err := stringArray(predicate["filter_ops"], label+".filter_ops", true); err != nil {
			return err
		}
	}
	return nil
}

func validateViewSchemaField(field map[string]any, label string) (string, error) {
	if err := requireAllowedKeys(field, viewSchemaFieldKeys, label); err != nil {
		return "", err
	}
	fieldKey, err := requiredString(field, "field_key", label)
	if err != nil {
		return "", err
	}
	for _, key := range []string{"label", "read_kind", "write_kind", "read_model"} {
		if _, err := requiredString(field, key, label); err != nil {
			return "", err
		}
	}
	for _, key := range []string{"default_hidden", "sortable", "groupable", "clearable", "writable"} {
		if _, err := requiredBool(field, key, label); err != nil {
			return "", err
		}
	}
	if _, err := nullableString(field, "header_sort_field_key", label); err != nil {
		return "", err
	}
	if _, err := stringArray(field["filter_ops"], label+".filter_ops", false); err != nil {
		return "", err
	}
	if _, err := nullableString(field, "conflict_resolution_class", label); err != nil {
		return "", err
	}
	if _, err := nullableString(field, "entity_binding_mode", label); err != nil {
		return "", err
	}
	if _, err := nullableString(field, "string_contract_id", label); err != nil {
		return "", err
	}
	if _, err := nullableString(field, "direct_scalar_contract_id", label); err != nil {
		return "", err
	}
	if _, err := nullableString(field, "direct_reference_contract_id", label); err != nil {
		return "", err
	}
	if field["enum_values"] != nil {
		if _, err := stringArray(field["enum_values"], label+".enum_values", true); err != nil {
			return "", err
		}
	}
	if field["create_writable"] != nil {
		if _, err := requiredBool(field, "create_writable", label); err != nil {
			return "", err
		}
	}
	writeKind, _ := field["write_kind"].(string)
	switch writeKind {
	case "direct_value":
		if _, err := requiredString(field, "write_target", label); err != nil {
			return "", err
		}
		if _, ok := field["write_action"]; ok {
			return "", fmt.Errorf("%s must not declare write_action for write_kind=direct_value", label)
		}
	case "action_payload":
		if _, err := requiredString(field, "write_action", label); err != nil {
			return "", err
		}
		if _, ok := field["write_target"]; ok {
			return "", fmt.Errorf("%s must not declare write_target for write_kind=action_payload", label)
		}
	case "read_only":
		if _, ok := field["write_target"]; ok {
			return "", fmt.Errorf("%s must not declare write_target for write_kind=read_only", label)
		}
		if _, ok := field["write_action"]; ok {
			return "", fmt.Errorf("%s must not declare write_action for write_kind=read_only", label)
		}
	default:
		return "", fmt.Errorf("%s.write_kind has unsupported value %q", label, writeKind)
	}
	return fieldKey, nil
}

func validateErrorRegistry(value any) error {
	object, err := asObject(value, "contracts/errors/index.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, errorRegistryKeys, "contracts/errors/index.json"); err != nil {
		return err
	}
	if err := requireDraftSchema(object, "contracts/errors/index.json"); err != nil {
		return err
	}
	if _, err := requiredString(object, "registry_id", "contracts/errors/index.json"); err != nil {
		return err
	}
	if _, err := requiredString(object, "note", "contracts/errors/index.json"); err != nil {
		return err
	}
	errors, err := objectArray(object["errors"], "errors")
	if err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for index, entry := range errors {
		label := fmt.Sprintf("errors[%d]", index+1)
		if err := requireAllowedKeys(entry, errorEntryKeys, label); err != nil {
			return err
		}
		code, err := requiredString(entry, "code", label)
		if err != nil {
			return err
		}
		if _, exists := seen[code]; exists {
			return fmt.Errorf("duplicate error code %s", code)
		}
		seen[code] = struct{}{}
		status, err := requiredInt(entry, "http_status", label)
		if err != nil {
			return err
		}
		if status < 400 || status > 599 {
			return fmt.Errorf("%s.http_status must be an HTTP error status", label)
		}
		if _, err := requiredString(entry, "summary", label); err != nil {
			return err
		}
	}
	reasonRegistriesValue, ok := object["reason_registries"]
	if !ok {
		return nil
	}
	reasonRegistries, err := objectArray(reasonRegistriesValue, "reason_registries")
	if err != nil {
		return err
	}
	seenReasonRegistries := map[string]struct{}{}
	for index, entry := range reasonRegistries {
		label := fmt.Sprintf("reason_registries[%d]", index+1)
		if err := requireAllowedKeys(entry, reasonRegistryEntryKeys, label); err != nil {
			return err
		}
		errorCode, err := requiredString(entry, "error_code", label)
		if err != nil {
			return err
		}
		if _, ok := seen[errorCode]; !ok {
			return fmt.Errorf("%s.error_code references unknown error code %s", label, errorCode)
		}
		if _, exists := seenReasonRegistries[errorCode]; exists {
			return fmt.Errorf("duplicate reason registry for error code %s", errorCode)
		}
		seenReasonRegistries[errorCode] = struct{}{}
		reasonCodes, err := objectArray(entry["reason_codes"], label+".reason_codes")
		if err != nil {
			return err
		}
		seenReasons := map[string]struct{}{}
		for reasonIndex, reasonEntry := range reasonCodes {
			reasonLabel := fmt.Sprintf("%s.reason_codes[%d]", label, reasonIndex+1)
			if err := requireAllowedKeys(reasonEntry, reasonCodeEntryKeys, reasonLabel); err != nil {
				return err
			}
			reasonCode, err := requiredString(reasonEntry, "code", reasonLabel)
			if err != nil {
				return err
			}
			if _, exists := seenReasons[reasonCode]; exists {
				return fmt.Errorf("%s contains duplicate reason code %s", label, reasonCode)
			}
			seenReasons[reasonCode] = struct{}{}
			if _, err := requiredString(reasonEntry, "summary", reasonLabel); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateExtensionRegistry(value any) error {
	object, err := asObject(value, "contracts/extensions/index.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, extensionRegistryKeys, "contracts/extensions/index.json"); err != nil {
		return err
	}
	if err := requireDraftSchema(object, "contracts/extensions/index.json"); err != nil {
		return err
	}
	if _, err := requiredString(object, "registry_id", "contracts/extensions/index.json"); err != nil {
		return err
	}
	if _, err := requiredString(object, "note", "contracts/extensions/index.json"); err != nil {
		return err
	}
	profiles, err := objectArray(object["profiles"], "profiles")
	if err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for index, profile := range profiles {
		label := fmt.Sprintf("profiles[%d]", index+1)
		if err := requireAllowedKeys(profile, extensionProfileKeys, label); err != nil {
			return err
		}
		profileID, err := requiredString(profile, "profile_id", label)
		if err != nil {
			return err
		}
		if _, exists := seen[profileID]; exists {
			return fmt.Errorf("duplicate extension profile %s", profileID)
		}
		seen[profileID] = struct{}{}
		routes, err := stringArray(profile["route_families"], label+".route_families", true)
		if err != nil {
			return err
		}
		for _, route := range routes {
			if !strings.HasPrefix(route, "/") {
				return fmt.Errorf("%s.route_families must contain route paths", label)
			}
		}
	}
	return nil
}

func validateWSIndex(value any) error {
	object, err := asObject(value, "contracts/ws/index.schema.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, wsIndexKeys, "contracts/ws/index.schema.json"); err != nil {
		return err
	}
	if err := requireDraftSchema(object, "contracts/ws/index.schema.json"); err != nil {
		return err
	}
	if _, err := requiredString(object, "$id", "contracts/ws/index.schema.json"); err != nil {
		return err
	}
	if _, err := requiredString(object, "title", "contracts/ws/index.schema.json"); err != nil {
		return err
	}
	if _, err := requiredString(object, "description", "contracts/ws/index.schema.json"); err != nil {
		return err
	}
	if typ, err := requiredString(object, "type", "contracts/ws/index.schema.json"); err != nil {
		return err
	} else if typ != "object" {
		return fmt.Errorf("contracts/ws/index.schema.json.type must be object")
	}
	additionalProperties, err := requiredBool(object, "additionalProperties", "contracts/ws/index.schema.json")
	if err != nil {
		return err
	}
	if additionalProperties {
		return fmt.Errorf("contracts/ws/index.schema.json.additionalProperties must be false")
	}
	properties, err := asObject(object["properties"], "contracts/ws/index.schema.json.properties")
	if err != nil {
		return err
	}
	for _, key := range []string{"route", "messages"} {
		if _, ok := properties[key]; !ok {
			return fmt.Errorf("contracts/ws/index.schema.json.properties must declare %s", key)
		}
	}
	required, err := stringArray(object["required"], "contracts/ws/index.schema.json.required", true)
	if err != nil {
		return err
	}
	for _, key := range []string{"route", "messages"} {
		if !contains(required, key) {
			return fmt.Errorf("contracts/ws/index.schema.json.required must include %s", key)
		}
	}
	return nil
}

func requireDraftSchema(object map[string]any, label string) error {
	schema, err := requiredString(object, "$schema", label)
	if err != nil {
		return err
	}
	if schema != contractDraft202012Schema {
		return fmt.Errorf("%s.$schema must be %s", label, contractDraft202012Schema)
	}
	return nil
}

func asObject(value any, label string) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok || object == nil {
		return nil, fmt.Errorf("%s must be an object", label)
	}
	return object, nil
}

func objectArray(value any, label string) ([]map[string]any, error) {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("%s must be a non-empty array", label)
	}
	objects := make([]map[string]any, 0, len(items))
	for index, item := range items {
		object, err := asObject(item, fmt.Sprintf("%s[%d]", label, index+1))
		if err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	return objects, nil
}

func objectArrayAllowEmpty(value any, label string) ([]map[string]any, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array", label)
	}
	objects := make([]map[string]any, 0, len(items))
	for index, item := range items {
		object, err := asObject(item, fmt.Sprintf("%s[%d]", label, index+1))
		if err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	return objects, nil
}

func stringArray(value any, label string, requireNonEmpty bool) ([]string, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array", label)
	}
	if requireNonEmpty && len(items) == 0 {
		return nil, fmt.Errorf("%s must be a non-empty array", label)
	}
	values := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for index, item := range items {
		text, ok := item.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return nil, fmt.Errorf("%s[%d] must be a non-empty string", label, index+1)
		}
		if _, exists := seen[text]; exists {
			return nil, fmt.Errorf("%s contains duplicate %s", label, text)
		}
		seen[text] = struct{}{}
		values = append(values, text)
	}
	return values, nil
}

func requiredString(object map[string]any, key, label string) (string, error) {
	value, ok := object[key]
	if !ok {
		return "", fmt.Errorf("%s.%s is required", label, key)
	}
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("%s.%s must be a non-empty string", label, key)
	}
	return text, nil
}

func nullableString(object map[string]any, key, label string) (string, error) {
	value, ok := object[key]
	if !ok {
		return "", fmt.Errorf("%s.%s is required", label, key)
	}
	if value == nil {
		return "", nil
	}
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("%s.%s must be null or a non-empty string", label, key)
	}
	return text, nil
}

func requireEnumString(object map[string]any, key, label string, allowed ...string) (string, error) {
	value, err := requiredString(object, key, label)
	if err != nil {
		return "", err
	}
	for _, candidate := range allowed {
		if value == candidate {
			return value, nil
		}
	}
	return "", fmt.Errorf("%s.%s must be one of %s", label, key, strings.Join(allowed, "|"))
}

func requiredBool(object map[string]any, key, label string) (bool, error) {
	value, ok := object[key]
	if !ok {
		return false, fmt.Errorf("%s.%s is required", label, key)
	}
	boolean, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("%s.%s must be a boolean", label, key)
	}
	return boolean, nil
}

func requiredInt(object map[string]any, key, label string) (int64, error) {
	value, ok := object[key]
	if !ok {
		return 0, fmt.Errorf("%s.%s is required", label, key)
	}
	number, ok := value.(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s.%s must be an integer", label, key)
	}
	integer, err := number.Int64()
	if err != nil {
		return 0, fmt.Errorf("%s.%s must be an integer", label, key)
	}
	return integer, nil
}

func requireAllowedKeys(object map[string]any, allowed map[string]struct{}, label string) error {
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("%s has unknown key %s", label, key)
		}
	}
	return nil
}

func requireKnownStrings(values []string, known map[string]struct{}, label string) error {
	for _, value := range values {
		if _, ok := known[value]; !ok {
			return fmt.Errorf("%s references unknown field %s", label, value)
		}
	}
	return nil
}

func stringSet(values ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
