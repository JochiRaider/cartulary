package main

import (
	"crypto/sha256"
	"encoding/hex"
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
		"base_projection",
		"canonical_source_filter",
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
		"inspector_config",
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
		"grid_editable",
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
	viewSchemaIndexKeys       = stringSet("$schema", "registry_id", "note", "view_schemas")
	viewSchemaIndexEntryKeys  = stringSet("view_schema_id", "title", "surface_kind", "surface_status", "source_record_types", "required_reference_pack_keys", "artifact_path")
	syntheticPredicateKeys    = stringSet("field_key", "label", "filter_ops")
	canonicalSourceFilterKeys = stringSet("kind", "field", "value")
	inspectorConfigKeys       = stringSet("inspector_config_schema_id", "view_schema_id", "default_open", "subject_binding", "no_row_state", "unsupported_feature_behavior", "panels", "feature_groups")
	inspectorSubjectKeys      = stringSet("kind")
	inspectorPanelKeys        = stringSet("panel_id", "label")
	inspectorFeatureKeys      = stringSet("feature_group_key", "panel_id", "label", "minimum_incident_role", "mutates", "requires_confirmation", "route_binding", "seed_bindings", "disabled_when", "success_result_behavior", "failure_result_behavior")
	inspectorRouteBindingKeys = stringSet("kind", "owner", "target_view_schema_id", "action_key")
	inspectorSeedBindingKeys  = stringSet("target_field_key", "source")
	inspectorSeedSourceKeys   = stringSet("kind", "source_field_key", "value")
	inspectorDisabledTokens   = stringSet("no_row_selected", "incident_closed", "authorization_lost", "row_version_changed", "record_deleted", "record_merged", "evidence_preview_unavailable", "merge_target_unavailable", "record_not_deleted", "rollback_target_unavailable", "party_text_unavailable", "pivot_target_unavailable")
	inspectorRouteOwners      = stringSet("current_row_projection", "view_query_route", "view_row_create_route", "record_patch_route", "record_mark_reviewed_route", "record_supersede_route", "record_delete_route", "record_restore_route", "record_history_route", "record_rollback_route", "record_merge_route", "entity_mention_resolve_route", "evidence_attach_blob_route", "evidence_preview_handle_route", "evidence_download_handle_route")
	inspectorSuccessBehaviors = stringSet("preserve_selected_row", "retarget_selected_row", "clear_to_no_row_selected", "surface_pivot")
	inspectorFailureBehaviors = stringSet("show_same_shell_error_preserve_selection", "show_same_shell_error_invalidate_pending_action", "show_same_shell_error_clear_subject")
	errorRegistryKeys         = stringSet("$schema", "registry_id", "note", "errors", "reason_registries")
	errorEntryKeys            = stringSet("code", "http_status", "summary")
	reasonRegistryEntryKeys   = stringSet("error_code", "reason_codes")
	reasonCodeEntryKeys       = stringSet("code", "summary")
	extensionInputCatalogKeys = stringSet("$schema", "schema_id", "extensions_document_version", "extensions_document_sha256", "artifacts")
	extensionInputEntryKeys   = stringSet("path", "schema_id", "owner_id", "artifact_class")
	extensionArtifactClasses  = stringSet("owner_contract_manifest", "owner_fragment", "profile_contract", "shared_owner_resolution", "specification_input", "build_input", "validation_input", "traceability_input")
	wsIndexKeys               = stringSet("$schema", "$id", "title", "description", "type", "additionalProperties", "properties", "required")
)

var inspectorFeatureRegistry = map[string][]string{
	"cartulary.view.timeline.v2":              {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "entity_mentions.resolve", "entity_mentions.create_host", "entity_mentions.create_identity", "entity_mentions.dismiss", "entity_mentions.restore", "indicator.observations.manage", "relationships.manage", "evidence.attach_blob", "evidence.preview_handle", "evidence.download_handle", "timeline.mark_reviewed", "timeline.supersede", "create_related.note", "create_related.task_request", "create_related.decision", "create_related.evidence", "create_related.comm_log", "create_related.handoff", "create_related.status_review", "create_related.lesson"},
	"cartulary.view.hosts.v1":                 {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "entity.aliases.read", "entity.relationships.manage", "entity.merge", "surface_pivot.timeline", "surface_pivot.evidence", "surface_pivot.assessments", "create_related.note", "create_related.task_request", "create_related.decision"},
	"cartulary.view.identities.v1":            {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "entity.aliases.read", "entity.relationships.manage", "entity.merge", "surface_pivot.timeline", "surface_pivot.evidence", "surface_pivot.assessments", "create_related.note", "create_related.task_request", "create_related.decision"},
	"cartulary.view.evidence.v1":              {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "evidence.preview_handle", "evidence.download_handle", "evidence.attach_blob", "party.collector.link", "party.source.link", "party.reference.clear", "relationships.manage", "surface_pivot.linked_records", "surface_pivot.timeline", "create_related.note", "create_related.task_request", "create_related.decision"},
	"cartulary.view.notes.v1":                 {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "artifact.source_links.manage", "artifact.evidence_refs.manage", "artifact.tags.manage", "artifact.related_notes.manage", "surface_pivot.source_records", "create_related.task_request", "create_related.decision"},
	"cartulary.view.indicators.v1":            {"details.read", "relationships.read", "history.read", "record.delete", "record.restore", "history.rollback", "indicator.observations.pivot", "indicator.lifecycle.read", "relationships.manage", "create_related.task_request", "create_related.decision"},
	"cartulary.view.assessments.v1":           {"details.read", "relationships.read", "history.read", "record.delete", "record.restore", "history.rollback", "assessment.subject_pivot", "assessment.prior_history", "assessment.support_refs.manage", "evidence.refs.manage", "create_related.task_request", "create_related.decision"},
	"cartulary.view.task_requests.v1":         {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "task.links.manage", "task.requester_party.link", "task.requester_party.clear", "task.decision.link", "task.decision.clear", "task.status.transition", "create_related.comm_log", "create_related.status_review", "create_related.lesson"},
	"cartulary.view.decisions.v1":             {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "decision.support_refs.manage", "decision.affected_records.manage", "decision.status.transition", "decision.supersede", "create_related.task_request", "create_related.comm_log", "create_related.status_review"},
	"cartulary.view.parties.v1":               {"details.read", "relationships.read", "history.read", "record.delete", "record.restore", "history.rollback", "party.usage_pivot.requester", "party.usage_pivot.collector_source", "party.usage_pivot.audience_attendee", "party.usage_pivot.owner_stakeholder", "party.reference.link", "party.reference.clear", "party.reference.clear_both"},
	"cartulary.view.comm_log.v1":              {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "comm.decisions.link", "comm.action_tasks.link", "comm.parties.manage", "comm.next_report.manage", "create_related.task_request", "create_related.status_review"},
	"cartulary.view.handoff.v1":               {"details.read", "relationships.read", "history.read", "record.delete", "record.restore", "history.rollback", "handoff.acknowledge", "handoff.open_tasks.review", "handoff.open_decisions.review", "handoff.risks.review", "handoff.next_checks.manage", "create_related.task_request", "create_related.status_review"},
	"cartulary.view.status_review.v1":         {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "status_review.blocked_tasks.review", "status_review.pending_evidence.review", "status_review.open_decisions.review", "status_review.risks.review", "status_review.next_report.manage", "create_related.task_request", "create_related.comm_log"},
	"cartulary.view.lesson.v1":                {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "lesson.followup_tasks.manage", "lesson.evidence_refs.manage", "lesson.owner.manage", "lesson.close_or_reopen", "create_related.task_request"},
	"cartulary.view.findings.v1":              {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "finding.support_refs.manage", "finding.contradictory_refs.manage", "finding.evidence_refs.manage", "finding.owner.manage", "finding.close_or_reopen", "create_related.task_request", "create_related.decision"},
	"cartulary.view.investigative_queries.v1": {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "query.source.link", "query.result.link", "query.evidence_refs.manage", "query.findings.link", "create_related.task_request"},
	"cartulary.view.forensic_keywords.v1":     {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "keyword.evidence_refs.manage", "keyword.timeline_rows.link", "keyword.findings.link", "create_related.task_request"},
}

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
		if relativePath == "index.json" {
			return validateExtensionInputCatalog(value)
		}
		return validateExtensionAuthoredArtifact(value, relativePath)
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
	if familyDir == "extensions" {
		return validateExtensionContractFamily(root)
	}
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
		indexPackKeys, err := stringArray(entry["required_reference_pack_keys"], label+".required_reference_pack_keys", false)
		if err != nil {
			return err
		}
		schemaPackKeys, err := stringArray(schema["required_reference_pack_keys"], artifactPath+".required_reference_pack_keys", false)
		if err != nil {
			return err
		}
		if strings.Join(indexPackKeys, "\x00") != strings.Join(schemaPackKeys, "\x00") {
			return fmt.Errorf("%s.required_reference_pack_keys must match %s.required_reference_pack_keys", label, artifactPath)
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
		surfaceKind, err := requireEnumString(entry, "surface_kind", label, "built_in_sheet", "system_view")
		if err != nil {
			return err
		}
		surfaceStatus, err := requireEnumString(
			entry,
			"surface_status",
			label,
			"required_built_in_sheet",
			"required_system_view",
			"standardized_optional_workbook_surface",
		)
		if err != nil {
			return err
		}
		expectedKind := "system_view"
		if surfaceStatus == "required_built_in_sheet" {
			expectedKind = "built_in_sheet"
		}
		if surfaceKind != expectedKind {
			return fmt.Errorf("%s.surface_kind must be %s for surface_status %s", label, expectedKind, surfaceStatus)
		}
		if _, err := stringArray(entry["source_record_types"], label+".source_record_types", true); err != nil {
			return err
		}
		if _, err := stringArray(entry["required_reference_pack_keys"], label+".required_reference_pack_keys", false); err != nil {
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
	if _, ok := object["base_projection"]; ok {
		if _, err := requiredString(object, "base_projection", relativePath); err != nil {
			return err
		}
	}
	if _, ok := object["canonical_source_filter"]; ok {
		if err := validateCanonicalSourceFilter(object["canonical_source_filter"], relativePath+".canonical_source_filter"); err != nil {
			return err
		}
	}
	inlineCreate, err := asObject(object["inline_create"], relativePath+".inline_create")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(inlineCreate, stringSet("permits_zero_field_create", "minimum_create_field_sets"), relativePath+".inline_create"); err != nil {
		return err
	}
	if _, err := requiredBool(inlineCreate, "permits_zero_field_create", relativePath+".inline_create"); err != nil {
		return err
	}
	if _, ok := inlineCreate["minimum_create_field_sets"]; ok {
		if _, err := stringMatrix(inlineCreate["minimum_create_field_sets"], relativePath+".inline_create.minimum_create_field_sets"); err != nil {
			return err
		}
	}
	if err := validateInspectorConfig(object["inspector_config"], viewSchemaID, relativePath+".inspector_config"); err != nil {
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
	if rawSets, ok := inlineCreate["minimum_create_field_sets"]; ok {
		fieldSets, err := stringMatrix(rawSets, relativePath+".inline_create.minimum_create_field_sets")
		if err != nil {
			return err
		}
		for setIndex, fieldSet := range fieldSets {
			if len(fieldSet) == 0 {
				return fmt.Errorf("%s.inline_create.minimum_create_field_sets[%d] must not be empty", relativePath, setIndex+1)
			}
			for itemIndex, fieldKey := range fieldSet {
				if _, ok := fieldKeys[fieldKey]; !ok {
					return fmt.Errorf("%s.inline_create.minimum_create_field_sets[%d][%d] references unknown field_key %s", relativePath, setIndex+1, itemIndex+1, fieldKey)
				}
			}
		}
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

func validateInspectorConfig(value any, viewSchemaID string, label string) error {
	object, err := asObject(value, label)
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, inspectorConfigKeys, label); err != nil {
		return err
	}
	if schemaID, err := requireEnumString(object, "inspector_config_schema_id", label, "cartulary.inspector_config.v1"); err != nil {
		return err
	} else if schemaID == "" {
		return fmt.Errorf("%s.inspector_config_schema_id is required", label)
	}
	if got, err := requiredString(object, "view_schema_id", label); err != nil {
		return err
	} else if got != viewSchemaID {
		return fmt.Errorf("%s.view_schema_id must match containing view_schema_id %s", label, viewSchemaID)
	}
	defaultOpen, err := requiredBool(object, "default_open", label)
	if err != nil {
		return err
	}
	if defaultOpen {
		return fmt.Errorf("%s.default_open must be false", label)
	}
	subject, err := asObject(object["subject_binding"], label+".subject_binding")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(subject, inspectorSubjectKeys, label+".subject_binding"); err != nil {
		return err
	}
	if _, err := requireEnumString(subject, "kind", label+".subject_binding", "selected_record"); err != nil {
		return err
	}
	if _, err := requireEnumString(object, "no_row_state", label, "no_row_selected"); err != nil {
		return err
	}
	if _, err := requireEnumString(object, "unsupported_feature_behavior", label, "omit_feature"); err != nil {
		return err
	}

	panels, err := objectArray(object["panels"], label+".panels")
	if err != nil {
		return err
	}
	if len(panels) > 5 {
		return fmt.Errorf("%s.panels must contain at most 5 entries", label)
	}
	declaredPanels := map[string]struct{}{}
	for index, panel := range panels {
		panelLabel := fmt.Sprintf("%s.panels[%d]", label, index+1)
		if err := requireAllowedKeys(panel, inspectorPanelKeys, panelLabel); err != nil {
			return err
		}
		panelID, err := requireEnumString(panel, "panel_id", panelLabel, "details", "relationships", "evidence", "history", "workflow")
		if err != nil {
			return err
		}
		if _, exists := declaredPanels[panelID]; exists {
			return fmt.Errorf("%s duplicate panel_id %s", label, panelID)
		}
		declaredPanels[panelID] = struct{}{}
		if _, err := requiredString(panel, "label", panelLabel); err != nil {
			return err
		}
	}

	featureGroups, err := objectArrayAllowEmpty(object["feature_groups"], label+".feature_groups")
	if err != nil {
		return err
	}
	if len(featureGroups) > 64 {
		return fmt.Errorf("%s.feature_groups must contain at most 64 entries", label)
	}
	seenFeatures := map[string]struct{}{}
	for index, group := range featureGroups {
		groupLabel := fmt.Sprintf("%s.feature_groups[%d]", label, index+1)
		if err := validateInspectorFeatureGroup(group, groupLabel, declaredPanels); err != nil {
			return err
		}
		key, _ := requiredString(group, "feature_group_key", groupLabel)
		if !isInspectorFeatureKey(key) {
			return fmt.Errorf("%s.feature_group_key must be ASCII lower snake or dotted key", groupLabel)
		}
		if _, exists := seenFeatures[key]; exists {
			return fmt.Errorf("%s duplicate feature_group_key %s", label, key)
		}
		seenFeatures[key] = struct{}{}
	}
	if expected, ok := inspectorFeatureRegistry[viewSchemaID]; ok {
		if len(seenFeatures) != len(expected) {
			return fmt.Errorf("%s.feature_groups must contain exactly %d declared feature groups for %s, got %d", label, len(expected), viewSchemaID, len(seenFeatures))
		}
		for _, key := range expected {
			if _, ok := seenFeatures[key]; !ok {
				return fmt.Errorf("%s.feature_groups missing required feature_group_key %s for %s", label, key, viewSchemaID)
			}
		}
		expectedSet := stringSet(expected...)
		for key := range seenFeatures {
			if _, ok := expectedSet[key]; !ok {
				return fmt.Errorf("%s.feature_groups contains undeclared feature_group_key %s for %s", label, key, viewSchemaID)
			}
		}
	}
	return nil
}

func validateInspectorFeatureGroup(group map[string]any, label string, declaredPanels map[string]struct{}) error {
	if err := requireAllowedKeys(group, inspectorFeatureKeys, label); err != nil {
		return err
	}
	if _, err := requiredString(group, "feature_group_key", label); err != nil {
		return err
	}
	panelID, err := requiredString(group, "panel_id", label)
	if err != nil {
		return err
	}
	if _, ok := declaredPanels[panelID]; !ok {
		return fmt.Errorf("%s.panel_id references unknown panel_id %s", label, panelID)
	}
	if _, err := requiredString(group, "label", label); err != nil {
		return err
	}
	if err := validateNullableIncidentRole(group, "minimum_incident_role", label); err != nil {
		return err
	}
	if _, err := requiredBool(group, "mutates", label); err != nil {
		return err
	}
	if _, err := requiredBool(group, "requires_confirmation", label); err != nil {
		return err
	}
	routeBinding, err := asObject(group["route_binding"], label+".route_binding")
	if err != nil {
		return err
	}
	if err := validateInspectorRouteBinding(routeBinding, label+".route_binding"); err != nil {
		return err
	}
	if err := validateInspectorSeedBindings(group["seed_bindings"], label+".seed_bindings"); err != nil {
		return err
	}
	conditions, err := stringArray(group["disabled_when"], label+".disabled_when", false)
	if err != nil {
		return err
	}
	if len(conditions) > 16 {
		return fmt.Errorf("%s.disabled_when must contain at most 16 entries", label)
	}
	for _, condition := range conditions {
		if _, ok := inspectorDisabledTokens[condition]; !ok {
			return fmt.Errorf("%s.disabled_when references unknown condition %s", label, condition)
		}
	}
	if _, err := requireEnumStringFromSet(group, "success_result_behavior", label, inspectorSuccessBehaviors); err != nil {
		return err
	}
	if _, err := requireEnumStringFromSet(group, "failure_result_behavior", label, inspectorFailureBehaviors); err != nil {
		return err
	}
	return nil
}

func validateNullableIncidentRole(object map[string]any, key, label string) error {
	value, ok := object[key]
	if !ok {
		return fmt.Errorf("%s.%s is required", label, key)
	}
	if value == nil {
		return nil
	}
	role, ok := value.(string)
	if !ok || strings.TrimSpace(role) == "" {
		return fmt.Errorf("%s.%s must be null or a non-empty string", label, key)
	}
	for _, allowed := range []string{"viewer", "editor", "reviewer", "admin"} {
		if role == allowed {
			return nil
		}
	}
	return fmt.Errorf("%s.%s must be one of viewer|editor|reviewer|admin or null", label, key)
}

func validateInspectorRouteBinding(object map[string]any, label string) error {
	if err := requireAllowedKeys(object, inspectorRouteBindingKeys, label); err != nil {
		return err
	}
	if _, err := requireEnumString(object, "kind", label, "panel_read", "view_row_create", "record_patch", "record_action", "entity_mention_action", "evidence_access", "surface_pivot"); err != nil {
		return err
	}
	if _, err := requireEnumStringFromSet(object, "owner", label, inspectorRouteOwners); err != nil {
		return err
	}
	if value, ok := object["target_view_schema_id"]; ok {
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return fmt.Errorf("%s.target_view_schema_id must be a non-empty string", label)
		}
	}
	if value, ok := object["action_key"]; ok {
		text, ok := value.(string)
		if !ok || !isInspectorFeatureKey(text) {
			return fmt.Errorf("%s.action_key must be ASCII lower snake or dotted key", label)
		}
	}
	return nil
}

func validateInspectorSeedBindings(value any, label string) error {
	bindings, err := objectArrayAllowEmpty(value, label)
	if err != nil {
		return err
	}
	if len(bindings) > 16 {
		return fmt.Errorf("%s must contain at most 16 entries", label)
	}
	for index, binding := range bindings {
		bindingLabel := fmt.Sprintf("%s[%d]", label, index+1)
		if err := requireAllowedKeys(binding, inspectorSeedBindingKeys, bindingLabel); err != nil {
			return err
		}
		targetFieldKey, err := requiredString(binding, "target_field_key", bindingLabel)
		if err != nil {
			return err
		}
		if !isStableFieldKey(targetFieldKey) {
			return fmt.Errorf("%s.target_field_key must be a stable field_key", bindingLabel)
		}
		source, err := asObject(binding["source"], bindingLabel+".source")
		if err != nil {
			return err
		}
		if err := requireAllowedKeys(source, inspectorSeedSourceKeys, bindingLabel+".source"); err != nil {
			return err
		}
		kind, err := requireEnumString(source, "kind", bindingLabel+".source", "selected_record_id", "selected_field_value", "literal")
		if err != nil {
			return err
		}
		if kind == "selected_field_value" {
			sourceFieldKey, err := requiredString(source, "source_field_key", bindingLabel+".source")
			if err != nil {
				return err
			}
			if !isStableFieldKey(sourceFieldKey) {
				return fmt.Errorf("%s.source.source_field_key must be a stable field_key", bindingLabel)
			}
		}
		if kind == "literal" {
			if _, ok := source["value"]; !ok {
				return fmt.Errorf("%s.source.value is required for literal seed source", bindingLabel)
			}
		}
	}
	return nil
}

func isStableFieldKey(value string) bool {
	return isInspectorFeatureKey(value)
}

func isInspectorFeatureKey(value string) bool {
	if value == "" {
		return false
	}
	segmentStart := true
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			segmentStart = false
		case r >= '0' && r <= '9':
			if segmentStart {
				return false
			}
		case r == '_':
			if segmentStart {
				return false
			}
		case r == '.':
			if segmentStart {
				return false
			}
			segmentStart = true
		default:
			return false
		}
	}
	return !segmentStart
}

func validateCanonicalSourceFilter(value any, label string) error {
	object, err := asObject(value, label)
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, canonicalSourceFilterKeys, label); err != nil {
		return err
	}
	if _, err := requireEnumString(object, "kind", label, "artifact_type"); err != nil {
		return err
	}
	if _, err := requireEnumString(object, "field", label, "artifact_type"); err != nil {
		return err
	}
	if _, err := requiredString(object, "value", label); err != nil {
		return err
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
	readKind, _ := field["read_kind"].(string)
	for _, key := range []string{"default_hidden", "sortable", "groupable", "clearable", "writable", "grid_editable"} {
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
		if readKind != "enum" {
			return "", fmt.Errorf("%s enum_values requires read_kind=enum", label)
		}
	} else if readKind == "enum" {
		return "", fmt.Errorf("%s read_kind=enum requires enum_values", label)
	}
	if field["create_writable"] != nil {
		if _, err := requiredBool(field, "create_writable", label); err != nil {
			return "", err
		}
	}
	writeKind, _ := field["write_kind"].(string)
	gridEditable, _ := field["grid_editable"].(bool)
	writable, _ := field["writable"].(bool)
	if gridEditable && (writeKind != "direct_value" || !writable) {
		return "", fmt.Errorf("%s.grid_editable=true requires an existing-row writable direct_value field", label)
	}
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

func validateExtensionInputCatalog(value any) error {
	object, err := asObject(value, "contracts/extensions/index.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, extensionInputCatalogKeys, "contracts/extensions/index.json"); err != nil {
		return err
	}
	if err := requireDraftSchema(object, "contracts/extensions/index.json"); err != nil {
		return err
	}
	if schemaID, err := requiredString(object, "schema_id", "contracts/extensions/index.json"); err != nil {
		return err
	} else if schemaID != "cartulary.extension_authored_input_catalog.v1" {
		return fmt.Errorf("contracts/extensions/index.json.schema_id must be cartulary.extension_authored_input_catalog.v1")
	}
	if version, err := requiredString(object, "extensions_document_version", "contracts/extensions/index.json"); err != nil {
		return err
	} else if version != "0.6.1" {
		return fmt.Errorf("contracts/extensions/index.json.extensions_document_version must be 0.6.1")
	}
	if digest, err := requiredString(object, "extensions_document_sha256", "contracts/extensions/index.json"); err != nil {
		return err
	} else if !isLowerSHA256(digest) {
		return fmt.Errorf("contracts/extensions/index.json.extensions_document_sha256 must be lowercase SHA-256")
	}
	artifacts, err := objectArray(object["artifacts"], "artifacts")
	if err != nil {
		return err
	}
	if len(artifacts) == 0 {
		return fmt.Errorf("contracts/extensions/index.json.artifacts must not be empty")
	}
	previousPath := ""
	seenPaths := map[string]struct{}{}
	for index, entry := range artifacts {
		label := fmt.Sprintf("artifacts[%d]", index+1)
		if err := requireAllowedKeys(entry, extensionInputEntryKeys, label); err != nil {
			return err
		}
		path, err := requiredString(entry, "path", label)
		if err != nil {
			return err
		}
		if !validExtensionCatalogPath(path) {
			return fmt.Errorf("%s.path must be a normalized relative JSON path", label)
		}
		if previousPath != "" && previousPath >= path {
			return fmt.Errorf("contracts/extensions/index.json.artifacts must be unique and sorted by path")
		}
		previousPath = path
		if _, duplicate := seenPaths[path]; duplicate {
			return fmt.Errorf("duplicate extension artifact path %s", path)
		}
		seenPaths[path] = struct{}{}
		if _, err := requiredString(entry, "schema_id", label); err != nil {
			return err
		}
		if _, err := requiredString(entry, "owner_id", label); err != nil {
			return err
		}
		artifactClass, err := requiredString(entry, "artifact_class", label)
		if err != nil {
			return err
		}
		if _, ok := extensionArtifactClasses[artifactClass]; !ok {
			return fmt.Errorf("%s.artifact_class is not recognized", label)
		}
	}
	return nil
}

func validateExtensionAuthoredArtifact(value any, relativePath string) error {
	object, err := asObject(value, "contracts/extensions/"+relativePath)
	if err != nil {
		return err
	}
	if _, err := requiredString(object, "schema_id", "contracts/extensions/"+relativePath); err != nil {
		return err
	}
	return validateExtensionArtifactShape(value, relativePath)
}

func validateExtensionContractFamily(root string) error {
	baseDir := filepath.Join(root, "contracts", "extensions")
	contractRoot, err := os.OpenRoot(baseDir)
	if err != nil {
		return fmt.Errorf("open extensions contract root: %w", err)
	}
	defer contractRoot.Close()
	rawIndex, err := contractRoot.ReadFile("index.json")
	if err != nil {
		return fmt.Errorf("read extensions input catalog: %w", err)
	}
	decoded, err := decodeContract(rawIndex)
	if err != nil {
		return fmt.Errorf("decode extensions input catalog: %w", err)
	}
	if err := validateExtensionInputCatalog(decoded); err != nil {
		return err
	}
	index, err := asObject(decoded, "contracts/extensions/index.json")
	if err != nil {
		return err
	}
	documentDigest, _ := requiredString(index, "extensions_document_sha256", "contracts/extensions/index.json")
	documentBytes, err := os.ReadFile(filepath.Join(root, "docs", "extension-subsystem-nlspec.md"))
	if err != nil {
		return fmt.Errorf("read Extensions NLSpec for authored-input binding: %w", err)
	}
	actualDocumentDigest := sha256.Sum256(documentBytes)
	if hex.EncodeToString(actualDocumentDigest[:]) != documentDigest {
		return fmt.Errorf("contracts/extensions/index.json.extensions_document_sha256 is stale")
	}

	entries, _ := objectArray(index["artifacts"], "artifacts")
	indexed := make(map[string]string, len(entries))
	indexedObjects := make(map[string]map[string]any, len(entries))
	for entryIndex, entry := range entries {
		label := fmt.Sprintf("artifacts[%d]", entryIndex+1)
		path, _ := requiredString(entry, "path", label)
		schemaID, _ := requiredString(entry, "schema_id", label)
		raw, err := contractRoot.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%s.path references missing artifact %s", label, path)
		}
		value, err := decodeContract(raw)
		if err != nil {
			return fmt.Errorf("decode indexed extension artifact %s: %w", path, err)
		}
		object, err := asObject(value, "contracts/extensions/"+path)
		if err != nil {
			return err
		}
		actualSchemaID, err := requiredString(object, "schema_id", "contracts/extensions/"+path)
		if err != nil {
			return err
		}
		if actualSchemaID != schemaID {
			return fmt.Errorf("%s.schema_id does not match indexed artifact %s", label, path)
		}
		catalogOwnerID, _ := requiredString(entry, "owner_id", label)
		if actualOwnerID, hasOwnerID := object["owner_id"].(string); hasOwnerID && actualOwnerID != catalogOwnerID {
			return fmt.Errorf("%s.owner_id does not match indexed artifact %s", label, path)
		}
		indexed[path] = schemaID
		indexedObjects[path] = object
	}

	discovered := map[string]struct{}{}
	if err := filepath.WalkDir(baseDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() == "index.json" || !strings.HasSuffix(entry.Name(), ".json") {
			return nil
		}
		relative, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		discovered[relative] = struct{}{}
		if _, ok := indexed[relative]; !ok {
			return fmt.Errorf("unexpected extensions artifact %s", relative)
		}
		return nil
	}); err != nil {
		return err
	}
	for path := range indexed {
		if _, ok := discovered[path]; !ok {
			return fmt.Errorf("indexed extensions artifact %s is missing", path)
		}
	}
	if err := validateExtensionOwnerBindings(root, indexedObjects); err != nil {
		return err
	}
	if err := validateExtensionTraceabilityRanges(root, indexedObjects); err != nil {
		return err
	}
	return nil
}

func validExtensionCatalogPath(path string) bool {
	if path == "" || strings.Contains(path, "\\") || strings.HasPrefix(path, "/") || !strings.HasSuffix(path, ".json") {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean == path && path != "." && path != ".." && !strings.HasPrefix(path, "../") && !strings.Contains(path, "/../")
}

func isLowerSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && hex.EncodeToString(decoded) == value
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

func stringMatrix(value any, label string) ([][]string, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array", label)
	}
	values := make([][]string, 0, len(items))
	for index, item := range items {
		nested, err := stringArray(item, fmt.Sprintf("%s[%d]", label, index+1), true)
		if err != nil {
			return nil, err
		}
		values = append(values, nested)
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

func requireEnumStringFromSet(object map[string]any, key, label string, allowed map[string]struct{}) (string, error) {
	value, err := requiredString(object, key, label)
	if err != nil {
		return "", err
	}
	if _, ok := allowed[value]; ok {
		return value, nil
	}
	values := make([]string, 0, len(allowed))
	for candidate := range allowed {
		values = append(values, candidate)
	}
	sort.Strings(values)
	return "", fmt.Errorf("%s.%s must be one of %s", label, key, strings.Join(values, "|"))
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
