package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	tasksDecisionsSourceCatalogSchemaID = "cartulary.tasksdecisions.source_catalog.v1"
	tasksDecisionsSourceCatalogID       = "cartulary.tasksdecisions.sources.v1"
	tasksDecisionsSourceCatalogPath     = "contracts/tasksdecisions/source-catalog.v1.json"
)

var (
	tasksDecisionsCatalogKeys     = stringSet("$schema", "schema_id", "catalog_id", "surfaces")
	tasksDecisionsSurfaceKeys     = stringSet("view_schema_id", "record_type", "base_projection", "source_table", "revision_snapshot_schema_id", "direct_fields", "collection_fields")
	tasksDecisionsDirectFieldKeys = stringSet("field_key", "column", "reference_role", "expected_target_record_type", "mirror_link_type")
	tasksDecisionsCollectionKeys  = stringSet("field_key", "collection_family", "allowed_operations", "link_type", "expected_target_record_type")
	tasksDecisionsIdentifier      = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	tasksDecisionsFieldKey        = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)
	tasksDecisionsSurfaceFacts    = map[string]struct {
		recordType     string
		baseProjection string
		sourceTable    string
		snapshotSchema string
	}{
		"cartulary.view.decisions.v1": {
			recordType: "decision", baseProjection: "decision_grid_projection",
			sourceTable: "decisions", snapshotSchema: "cartulary.revisions.snapshot.decision.v1",
		},
		"cartulary.view.task_requests.v1": {
			recordType: "task_request", baseProjection: "task_request_grid_projection",
			sourceTable: "task_requests", snapshotSchema: "cartulary.revisions.snapshot.task_request.v1",
		},
	}
)

type tasksDecisionsSourceCatalog struct {
	Surfaces []tasksDecisionsSourceSurface
}

type tasksDecisionsSourceSurface struct {
	ViewSchemaID             string
	RecordType               string
	BaseProjection           string
	SourceTable              string
	RevisionSnapshotSchemaID string
	DirectFields             []tasksDecisionsDirectField
	CollectionFields         []tasksDecisionsCollectionField
}

type tasksDecisionsDirectField struct {
	FieldKey                 string
	Column                   string
	ReferenceRole            string
	ExpectedTargetRecordType string
	MirrorLinkType           string
	View                     tasksDecisionsViewFieldFacts
}

type tasksDecisionsCollectionField struct {
	FieldKey                 string
	CollectionFamily         string
	AllowedOperations        []string
	LinkType                 string
	ExpectedTargetRecordType string
	View                     tasksDecisionsViewFieldFacts
}

type tasksDecisionsViewFieldFacts struct {
	ReadKind                string
	WriteKind               string
	Writable                bool
	CreateWritable          bool
	Clearable               bool
	DirectReferenceContract string
	EnumValues              []string
}

func validateTasksDecisionsContractInput(relativePath string, value any) error {
	switch relativePath {
	case "source-catalog.v1.json":
		_, err := parseTasksDecisionsSourceCatalog(value)
		return err
	case "source-catalog.v1.schema.json":
		return validateTasksDecisionsSourceCatalogSchema(value)
	default:
		return fmt.Errorf("unexpected Tasks/Decisions contract artifact %s", relativePath)
	}
}

func validateTasksDecisionsSourceCatalogSchema(value any) error {
	const label = "contracts/tasksdecisions/source-catalog.v1.schema.json"
	object, err := asObject(value, label)
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, stringSet("$schema", "$id", "title", "type", "additionalProperties", "required", "properties", "$defs"), label); err != nil {
		return err
	}
	if err := requireDraftSchema(object, label); err != nil {
		return err
	}
	if object["additionalProperties"] != false {
		return fmt.Errorf("tasksdecisions source catalog schema must be closed")
	}
	return nil
}

func parseTasksDecisionsSourceCatalog(value any) (tasksDecisionsSourceCatalog, error) {
	object, err := asObject(value, tasksDecisionsSourceCatalogPath)
	if err != nil {
		return tasksDecisionsSourceCatalog{}, err
	}
	if err := requireAllowedKeys(object, tasksDecisionsCatalogKeys, tasksDecisionsSourceCatalogPath); err != nil {
		return tasksDecisionsSourceCatalog{}, err
	}
	if err := requireDraftSchema(object, tasksDecisionsSourceCatalogPath); err != nil {
		return tasksDecisionsSourceCatalog{}, err
	}
	if schemaID, err := requiredString(object, "schema_id", tasksDecisionsSourceCatalogPath); err != nil {
		return tasksDecisionsSourceCatalog{}, err
	} else if schemaID != tasksDecisionsSourceCatalogSchemaID {
		return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.schema_id must be %s", tasksDecisionsSourceCatalogPath, tasksDecisionsSourceCatalogSchemaID)
	}
	if catalogID, err := requiredString(object, "catalog_id", tasksDecisionsSourceCatalogPath); err != nil {
		return tasksDecisionsSourceCatalog{}, err
	} else if catalogID != tasksDecisionsSourceCatalogID {
		return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.catalog_id must be %s", tasksDecisionsSourceCatalogPath, tasksDecisionsSourceCatalogID)
	}
	surfaceObjects, err := objectArray(object["surfaces"], tasksDecisionsSourceCatalogPath+".surfaces")
	if err != nil {
		return tasksDecisionsSourceCatalog{}, err
	}
	if len(surfaceObjects) != 2 {
		return tasksDecisionsSourceCatalog{}, fmt.Errorf("tasksdecisions source catalog must contain exactly 2 surfaces")
	}
	catalog := tasksDecisionsSourceCatalog{Surfaces: make([]tasksDecisionsSourceSurface, 0, 2)}
	seenFields := map[string]string{}
	seenRecords := map[string]struct{}{}
	previousView := ""
	directCount := 0
	collectionCount := 0
	for surfaceIndex, surfaceObject := range surfaceObjects {
		label := fmt.Sprintf("%s.surfaces[%d]", tasksDecisionsSourceCatalogPath, surfaceIndex+1)
		if err := requireAllowedKeys(surfaceObject, tasksDecisionsSurfaceKeys, label); err != nil {
			return tasksDecisionsSourceCatalog{}, err
		}
		viewSchemaID, err := requiredString(surfaceObject, "view_schema_id", label)
		if err != nil {
			return tasksDecisionsSourceCatalog{}, err
		}
		if previousView != "" && previousView >= viewSchemaID {
			return tasksDecisionsSourceCatalog{}, fmt.Errorf("tasksdecisions surfaces must be unique and sorted by view_schema_id")
		}
		previousView = viewSchemaID
		want, known := tasksDecisionsSurfaceFacts[viewSchemaID]
		if !known {
			return tasksDecisionsSourceCatalog{}, fmt.Errorf("unknown tasksdecisions view schema %s", viewSchemaID)
		}
		recordType, err := requiredString(surfaceObject, "record_type", label)
		if err != nil {
			return tasksDecisionsSourceCatalog{}, err
		}
		baseProjection, err := requiredString(surfaceObject, "base_projection", label)
		if err != nil {
			return tasksDecisionsSourceCatalog{}, err
		}
		sourceTable, err := requiredString(surfaceObject, "source_table", label)
		if err != nil {
			return tasksDecisionsSourceCatalog{}, err
		}
		snapshotSchema, err := requiredString(surfaceObject, "revision_snapshot_schema_id", label)
		if err != nil {
			return tasksDecisionsSourceCatalog{}, err
		}
		if recordType != want.recordType || baseProjection != want.baseProjection || sourceTable != want.sourceTable || snapshotSchema != want.snapshotSchema {
			return tasksDecisionsSourceCatalog{}, fmt.Errorf("tasksdecisions surface %s has mismatched owner routing facts", viewSchemaID)
		}
		if !tasksDecisionsIdentifier.MatchString(sourceTable) {
			return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.source_table is not a safe identifier", label)
		}
		if _, duplicate := seenRecords[recordType]; duplicate {
			return tasksDecisionsSourceCatalog{}, fmt.Errorf("duplicate tasksdecisions record type %s", recordType)
		}
		seenRecords[recordType] = struct{}{}
		directObjects, err := objectArray(surfaceObject["direct_fields"], label+".direct_fields")
		if err != nil {
			return tasksDecisionsSourceCatalog{}, err
		}
		collectionObjects, err := objectArray(surfaceObject["collection_fields"], label+".collection_fields")
		if err != nil {
			return tasksDecisionsSourceCatalog{}, err
		}
		surface := tasksDecisionsSourceSurface{
			ViewSchemaID: viewSchemaID, RecordType: recordType, BaseProjection: baseProjection,
			SourceTable: sourceTable, RevisionSnapshotSchemaID: snapshotSchema,
		}
		previousField := ""
		for fieldIndex, fieldObject := range directObjects {
			fieldLabel := fmt.Sprintf("%s.direct_fields[%d]", label, fieldIndex+1)
			if err := requireAllowedKeys(fieldObject, tasksDecisionsDirectFieldKeys, fieldLabel); err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			fieldKey, err := requiredString(fieldObject, "field_key", fieldLabel)
			if err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			if previousField != "" && previousField >= fieldKey {
				return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.direct_fields must be unique and sorted by field_key", label)
			}
			previousField = fieldKey
			if err := registerTasksDecisionsField(seenFields, fieldKey, viewSchemaID); err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			column, err := requiredString(fieldObject, "column", fieldLabel)
			if err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			if !tasksDecisionsIdentifier.MatchString(column) {
				return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.column is not a safe identifier", fieldLabel)
			}
			referenceRole, err := optionalTasksDecisionsString(fieldObject, "reference_role", fieldLabel)
			if err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			targetType, err := optionalTasksDecisionsString(fieldObject, "expected_target_record_type", fieldLabel)
			if err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			mirrorLinkType, err := optionalTasksDecisionsString(fieldObject, "mirror_link_type", fieldLabel)
			if err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			if err := validateTasksDecisionsDirectReference(fieldLabel, fieldKey, referenceRole, targetType, mirrorLinkType); err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			surface.DirectFields = append(surface.DirectFields, tasksDecisionsDirectField{
				FieldKey: fieldKey, Column: column, ReferenceRole: referenceRole,
				ExpectedTargetRecordType: targetType, MirrorLinkType: mirrorLinkType,
			})
			directCount++
		}
		previousField = ""
		for fieldIndex, fieldObject := range collectionObjects {
			fieldLabel := fmt.Sprintf("%s.collection_fields[%d]", label, fieldIndex+1)
			if err := requireAllowedKeys(fieldObject, tasksDecisionsCollectionKeys, fieldLabel); err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			fieldKey, err := requiredString(fieldObject, "field_key", fieldLabel)
			if err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			if previousField != "" && previousField >= fieldKey {
				return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.collection_fields must be unique and sorted by field_key", label)
			}
			previousField = fieldKey
			if err := registerTasksDecisionsField(seenFields, fieldKey, viewSchemaID); err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			family, err := requiredString(fieldObject, "collection_family", fieldLabel)
			if err != nil || family != "record_ref" {
				return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.collection_family must be record_ref", fieldLabel)
			}
			operations, err := stringArray(fieldObject["allowed_operations"], fieldLabel+".allowed_operations", true)
			if err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			if strings.Join(operations, "\x00") != "add_record_ref\x00remove_record_ref" {
				return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.allowed_operations must be add_record_ref then remove_record_ref", fieldLabel)
			}
			linkType, err := requiredString(fieldObject, "link_type", fieldLabel)
			if err != nil || (linkType != "references_record" && linkType != "supported_by") {
				return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.link_type is invalid", fieldLabel)
			}
			targetType, err := optionalTasksDecisionsString(fieldObject, "expected_target_record_type", fieldLabel)
			if err != nil {
				return tasksDecisionsSourceCatalog{}, err
			}
			if targetType != "" && targetType != "decision" && targetType != "evidence" && targetType != "party" && targetType != "task_request" {
				return tasksDecisionsSourceCatalog{}, fmt.Errorf("%s.expected_target_record_type is invalid", fieldLabel)
			}
			surface.CollectionFields = append(surface.CollectionFields, tasksDecisionsCollectionField{
				FieldKey: fieldKey, CollectionFamily: family, AllowedOperations: operations,
				LinkType: linkType, ExpectedTargetRecordType: targetType,
			})
			collectionCount++
		}
		catalog.Surfaces = append(catalog.Surfaces, surface)
	}
	if directCount != 20 || collectionCount != 3 {
		return tasksDecisionsSourceCatalog{}, fmt.Errorf("tasksdecisions source catalog counts are %d direct and %d collection fields; want 20 and 3", directCount, collectionCount)
	}
	return catalog, nil
}

func registerTasksDecisionsField(seen map[string]string, fieldKey, viewSchemaID string) error {
	if !tasksDecisionsFieldKey.MatchString(fieldKey) {
		return fmt.Errorf("tasksdecisions field key %s is invalid", fieldKey)
	}
	if previous, duplicate := seen[fieldKey]; duplicate {
		return fmt.Errorf("tasksdecisions field %s is owned by both %s and %s", fieldKey, previous, viewSchemaID)
	}
	seen[fieldKey] = viewSchemaID
	return nil
}

func optionalTasksDecisionsString(object map[string]any, key, label string) (string, error) {
	value, present := object[key]
	if !present {
		return "", nil
	}
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("%s.%s must be a non-empty string when present", label, key)
	}
	return text, nil
}

func validateTasksDecisionsDirectReference(label, fieldKey, role, targetType, mirrorLinkType string) error {
	switch role {
	case "":
		if targetType != "" || mirrorLinkType != "" {
			return fmt.Errorf("%s reference details require reference_role", label)
		}
	case "incident_member_user":
		if targetType != "" || mirrorLinkType != "" {
			return fmt.Errorf("%s incident member reference must not declare record or link facts", label)
		}
	case "same_incident_record":
		if targetType != "decision" && targetType != "party" {
			return fmt.Errorf("%s same-incident reference requires a supported target record type", label)
		}
		if mirrorLinkType != "" && (fieldKey != "task.decision_record_id" || mirrorLinkType != "references_record") {
			return fmt.Errorf("%s declares an unsupported mirrored link", label)
		}
	default:
		return fmt.Errorf("%s.reference_role is unknown", label)
	}
	return nil
}

func validateTasksDecisionsSourceCatalogFamily(root string) error {
	base := filepath.Join(root, "contracts", "tasksdecisions")
	entries, err := os.ReadDir(base)
	if err != nil {
		return fmt.Errorf("read Tasks/Decisions contract family: %w", err)
	}
	expected := stringSet("source-catalog.v1.json", "source-catalog.v1.schema.json")
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if _, ok := expected[entry.Name()]; !ok {
			return fmt.Errorf("unexpected Tasks/Decisions contract artifact %s", entry.Name())
		}
		delete(expected, entry.Name())
	}
	if len(expected) != 0 {
		return fmt.Errorf("tasksdecisions contract family is incomplete")
	}
	_, err = loadTasksDecisionsSourceCatalog(root)
	return err
}

func loadTasksDecisionsSourceCatalog(root string) (tasksDecisionsSourceCatalog, error) {
	data, err := os.ReadFile(filepath.Join(root, tasksDecisionsSourceCatalogPath))
	if err != nil {
		return tasksDecisionsSourceCatalog{}, err
	}
	value, err := decodeContract(data)
	if err != nil {
		return tasksDecisionsSourceCatalog{}, err
	}
	catalog, err := parseTasksDecisionsSourceCatalog(value)
	if err != nil {
		return tasksDecisionsSourceCatalog{}, err
	}
	for index := range catalog.Surfaces {
		if err := enrichTasksDecisionsSurfaceFromViewSchema(root, &catalog.Surfaces[index]); err != nil {
			return tasksDecisionsSourceCatalog{}, err
		}
	}
	return catalog, nil
}

func enrichTasksDecisionsSurfaceFromViewSchema(root string, surface *tasksDecisionsSourceSurface) error {
	data, err := os.ReadFile(filepath.Join(root, "contracts", "view-schemas", surface.ViewSchemaID+".json"))
	if err != nil {
		return fmt.Errorf("tasksdecisions surface %s view schema is missing: %w", surface.ViewSchemaID, err)
	}
	value, err := decodeContract(data)
	if err != nil {
		return err
	}
	document, err := asObject(value, surface.ViewSchemaID)
	if err != nil {
		return err
	}
	baseProjection, err := requiredString(document, "base_projection", surface.ViewSchemaID)
	if err != nil || baseProjection != surface.BaseProjection {
		return fmt.Errorf("tasksdecisions surface %s base projection mismatches source catalog", surface.ViewSchemaID)
	}
	sourceTypes, err := stringArray(document["source_record_types"], surface.ViewSchemaID+".source_record_types", true)
	if err != nil || len(sourceTypes) != 1 || sourceTypes[0] != surface.RecordType {
		return fmt.Errorf("tasksdecisions surface %s source record type mismatches source catalog", surface.ViewSchemaID)
	}
	fieldObjects, err := objectArray(document["fields"], surface.ViewSchemaID+".fields")
	if err != nil {
		return err
	}
	viewFields := map[string]tasksDecisionsViewFieldFacts{}
	for fieldIndex, fieldObject := range fieldObjects {
		label := fmt.Sprintf("%s.fields[%d]", surface.ViewSchemaID, fieldIndex+1)
		fieldKey, err := requiredString(fieldObject, "field_key", label)
		if err != nil {
			return err
		}
		writable, err := artifactOptionalBool(fieldObject, "writable", label)
		if err != nil {
			return err
		}
		createWritable, err := artifactOptionalBool(fieldObject, "create_writable", label)
		if err != nil {
			return err
		}
		if !writable && !createWritable {
			continue
		}
		readKind, err := requiredString(fieldObject, "read_kind", label)
		if err != nil {
			return err
		}
		writeKind, err := requiredString(fieldObject, "write_kind", label)
		if err != nil {
			return err
		}
		clearable, err := requiredBool(fieldObject, "clearable", label)
		if err != nil {
			return err
		}
		reference, err := nullableString(fieldObject, "direct_reference_contract_id", label)
		if err != nil {
			return err
		}
		enumValues, err := tasksDecisionsOptionalStringArray(fieldObject, "enum_values", label)
		if err != nil {
			return err
		}
		viewFields[fieldKey] = tasksDecisionsViewFieldFacts{
			ReadKind: readKind, WriteKind: writeKind, Writable: writable,
			CreateWritable: createWritable, Clearable: clearable,
			DirectReferenceContract: reference, EnumValues: enumValues,
		}
	}
	seen := map[string]struct{}{}
	for index := range surface.DirectFields {
		field := &surface.DirectFields[index]
		facts, ok := viewFields[field.FieldKey]
		if !ok || facts.WriteKind != "direct_value" {
			return fmt.Errorf("tasksdecisions catalog direct field %s is missing, read-only, or mismatched", field.FieldKey)
		}
		if err := validateTasksDecisionsViewReference(field, facts); err != nil {
			return err
		}
		field.View = facts
		seen[field.FieldKey] = struct{}{}
	}
	for index := range surface.CollectionFields {
		field := &surface.CollectionFields[index]
		facts, ok := viewFields[field.FieldKey]
		if !ok || facts.WriteKind != "action_payload" || facts.ReadKind != "collection" || facts.DirectReferenceContract != "" || facts.Clearable || len(facts.EnumValues) != 0 {
			return fmt.Errorf("tasksdecisions catalog collection field %s is missing, read-only, or mismatched", field.FieldKey)
		}
		field.View = facts
		seen[field.FieldKey] = struct{}{}
	}
	for fieldKey := range viewFields {
		if _, ok := seen[fieldKey]; !ok {
			return fmt.Errorf("tasksdecisions writable field %s is missing from source catalog surface %s", fieldKey, surface.ViewSchemaID)
		}
	}
	return nil
}

func validateTasksDecisionsViewReference(field *tasksDecisionsDirectField, facts tasksDecisionsViewFieldFacts) error {
	wantContract := ""
	switch field.ReferenceRole {
	case "incident_member_user":
		wantContract = "incident_member_user_ref_v1"
	case "same_incident_record":
		if field.ExpectedTargetRecordType == "party" {
			wantContract = "same_incident_party_ref_v1"
		} else {
			wantContract = "same_incident_decision_ref_v1"
		}
	}
	if facts.DirectReferenceContract != wantContract {
		return fmt.Errorf("tasksdecisions direct field %s has mismatched reference contract", field.FieldKey)
	}
	return nil
}

func tasksDecisionsOptionalStringArray(object map[string]any, key, label string) ([]string, error) {
	value, present := object[key]
	if !present || value == nil {
		return nil, nil
	}
	return stringArray(value, label+"."+key, true)
}

func writeTasksDecisionsSourceCatalogGo(root string) error {
	catalog, err := loadTasksDecisionsSourceCatalog(root)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	buffer.WriteString("// Code generated by tools/contractgen; DO NOT EDIT.\n\npackage contracttasksdecisions\n\n")
	buffer.WriteString("type ViewFieldFacts struct {\n\tReadKind string\n\tWriteKind string\n\tWritable bool\n\tCreateWritable bool\n\tClearable bool\n\tDirectReferenceContractID string\n\tEnumValues []string\n}\n\n")
	buffer.WriteString("type DirectField struct {\n\tFieldKey string\n\tColumn string\n\tReferenceRole string\n\tExpectedTargetRecordType string\n\tMirrorLinkType string\n\tView ViewFieldFacts\n}\n\n")
	buffer.WriteString("type CollectionField struct {\n\tFieldKey string\n\tCollectionFamily string\n\tAllowedOperations []string\n\tLinkType string\n\tExpectedTargetRecordType string\n\tView ViewFieldFacts\n}\n\n")
	buffer.WriteString("type SourceSurface struct {\n\tViewSchemaID string\n\tRecordType string\n\tBaseProjection string\n\tSourceTable string\n\tRevisionSnapshotSchemaID string\n\tDirectFields []DirectField\n\tCollectionFields []CollectionField\n}\n\n")
	buffer.WriteString("const (\n")
	for _, surface := range catalog.Surfaces {
		fmt.Fprintf(&buffer, "\t%sViewSchemaID = %s\n", tasksDecisionsGoName(surface.RecordType), strconv.Quote(surface.ViewSchemaID))
	}
	buffer.WriteString(")\n\nvar SourceCatalog = []SourceSurface{\n")
	for _, surface := range catalog.Surfaces {
		fmt.Fprintf(&buffer, "\t{ViewSchemaID: %s, RecordType: %s, BaseProjection: %s, SourceTable: %s, RevisionSnapshotSchemaID: %s,\n", strconv.Quote(surface.ViewSchemaID), strconv.Quote(surface.RecordType), strconv.Quote(surface.BaseProjection), strconv.Quote(surface.SourceTable), strconv.Quote(surface.RevisionSnapshotSchemaID))
		buffer.WriteString("\t\tDirectFields: []DirectField{\n")
		for _, field := range surface.DirectFields {
			fmt.Fprintf(&buffer, "\t\t\t{FieldKey: %s, Column: %s, ReferenceRole: %s, ExpectedTargetRecordType: %s, MirrorLinkType: %s, View: ", strconv.Quote(field.FieldKey), strconv.Quote(field.Column), strconv.Quote(field.ReferenceRole), strconv.Quote(field.ExpectedTargetRecordType), strconv.Quote(field.MirrorLinkType))
			writeTasksDecisionsViewFacts(&buffer, field.View)
			buffer.WriteString("},\n")
		}
		buffer.WriteString("\t\t},\n\t\tCollectionFields: []CollectionField{\n")
		for _, field := range surface.CollectionFields {
			fmt.Fprintf(&buffer, "\t\t\t{FieldKey: %s, CollectionFamily: %s, AllowedOperations: []string{", strconv.Quote(field.FieldKey), strconv.Quote(field.CollectionFamily))
			for index, operation := range field.AllowedOperations {
				if index > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(strconv.Quote(operation))
			}
			fmt.Fprintf(&buffer, "}, LinkType: %s, ExpectedTargetRecordType: %s, View: ", strconv.Quote(field.LinkType), strconv.Quote(field.ExpectedTargetRecordType))
			writeTasksDecisionsViewFacts(&buffer, field.View)
			buffer.WriteString("},\n")
		}
		buffer.WriteString("\t\t},\n\t},\n")
	}
	buffer.WriteString("}\n")
	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("format generated Tasks/Decisions source catalog: %w", err)
	}
	return stageGeneratedFile(filepath.Join(root, "internal", "gen", "contracttasksdecisions", "source_catalog_gen.go"), formatted, 0o644)
}

func writeTasksDecisionsViewFacts(buffer *bytes.Buffer, facts tasksDecisionsViewFieldFacts) {
	fmt.Fprintf(buffer, "ViewFieldFacts{ReadKind: %s, WriteKind: %s, Writable: %t, CreateWritable: %t, Clearable: %t, DirectReferenceContractID: %s, EnumValues: []string{", strconv.Quote(facts.ReadKind), strconv.Quote(facts.WriteKind), facts.Writable, facts.CreateWritable, facts.Clearable, strconv.Quote(facts.DirectReferenceContract))
	for index, value := range facts.EnumValues {
		if index > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(strconv.Quote(value))
	}
	buffer.WriteString("}}")
}

func tasksDecisionsGoName(value string) string {
	parts := strings.Split(value, "_")
	for index := range parts {
		parts[index] = strings.ToUpper(parts[index][:1]) + parts[index][1:]
	}
	return strings.Join(parts, "")
}
