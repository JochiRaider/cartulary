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
	artifactSourceCatalogSchemaID = "cartulary.artifacts.source_catalog.v1"
	artifactSourceCatalogID       = "cartulary.artifacts.sources.v1"
	artifactSourceCatalogPath     = "contracts/artifacts/source-catalog.v1.json"
)

var (
	artifactCatalogKeys         = stringSet("$schema", "schema_id", "catalog_id", "surfaces")
	artifactSurfaceKeys         = stringSet("view_schema_id", "artifact_type", "direct_fields", "collection_fields")
	artifactDirectFieldKeys     = stringSet("field_key", "table", "column")
	artifactCollectionFieldKeys = stringSet("field_key", "collection_family", "allowed_operations", "link_type", "expected_target_record_type")
	artifactOwnedRelations      = stringSet("artifact_findings", "artifact_forensic_keywords", "artifact_investigative_queries", "artifacts")
	artifactCollectionFamilies  = stringSet("party_ref", "record_ref", "record_tag", "risk_ref")
	artifactLinkTypes           = stringSet("references_record", "supported_by")
	artifactTargetRecordTypes   = stringSet("decision", "evidence", "party", "task_request")
	artifactIdentifierPattern   = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	artifactFieldKeyPattern     = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)
	artifactOperationsByFamily  = map[string][]string{
		"party_ref":  {"add_party_ref", "remove_party_ref"},
		"record_ref": {"add_record_ref", "remove_record_ref"},
		"record_tag": {"add_tag", "remove_tag"},
		"risk_ref":   {"add_risk_ref", "remove_risk_ref"},
	}
)

type artifactSourceCatalog struct {
	Surfaces []artifactSourceSurface
}

type artifactSourceSurface struct {
	ViewSchemaID     string
	ArtifactType     string
	DirectFields     []artifactDirectField
	CollectionFields []artifactCollectionField
}

type artifactDirectField struct {
	FieldKey string
	Table    string
	Column   string
	View     artifactViewFieldFacts
}

type artifactCollectionField struct {
	FieldKey                 string
	CollectionFamily         string
	AllowedOperations        []string
	LinkType                 string
	ExpectedTargetRecordType string
	View                     artifactViewFieldFacts
}

type artifactViewFieldFacts struct {
	ReadKind            string
	WriteKind           string
	Writable            bool
	CreateWritable      bool
	Clearable           bool
	ReferenceContractID string
}

func validateArtifactContractInput(relativePath string, value any) error {
	switch relativePath {
	case "source-catalog.v1.json":
		_, err := parseArtifactSourceCatalog(value)
		return err
	case "source-catalog.v1.schema.json":
		return validateArtifactSourceCatalogSchema(value)
	default:
		return fmt.Errorf("unexpected Artifacts contract artifact %s", relativePath)
	}
}

func validateArtifactSourceCatalogSchema(value any) error {
	object, err := asObject(value, "contracts/artifacts/source-catalog.v1.schema.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, stringSet("$schema", "$id", "title", "type", "additionalProperties", "required", "properties", "$defs"), "contracts/artifacts/source-catalog.v1.schema.json"); err != nil {
		return err
	}
	if err := requireDraftSchema(object, "contracts/artifacts/source-catalog.v1.schema.json"); err != nil {
		return err
	}
	if object["additionalProperties"] != false {
		return fmt.Errorf("artifacts source catalog schema must be closed")
	}
	return nil
}

func parseArtifactSourceCatalog(value any) (artifactSourceCatalog, error) {
	object, err := asObject(value, artifactSourceCatalogPath)
	if err != nil {
		return artifactSourceCatalog{}, err
	}
	if err := requireAllowedKeys(object, artifactCatalogKeys, artifactSourceCatalogPath); err != nil {
		return artifactSourceCatalog{}, err
	}
	if err := requireDraftSchema(object, artifactSourceCatalogPath); err != nil {
		return artifactSourceCatalog{}, err
	}
	if schemaID, err := requiredString(object, "schema_id", artifactSourceCatalogPath); err != nil {
		return artifactSourceCatalog{}, err
	} else if schemaID != artifactSourceCatalogSchemaID {
		return artifactSourceCatalog{}, fmt.Errorf("%s.schema_id must be %s", artifactSourceCatalogPath, artifactSourceCatalogSchemaID)
	}
	if catalogID, err := requiredString(object, "catalog_id", artifactSourceCatalogPath); err != nil {
		return artifactSourceCatalog{}, err
	} else if catalogID != artifactSourceCatalogID {
		return artifactSourceCatalog{}, fmt.Errorf("%s.catalog_id must be %s", artifactSourceCatalogPath, artifactSourceCatalogID)
	}
	surfaceObjects, err := objectArray(object["surfaces"], artifactSourceCatalogPath+".surfaces")
	if err != nil {
		return artifactSourceCatalog{}, err
	}
	if len(surfaceObjects) != 8 {
		return artifactSourceCatalog{}, fmt.Errorf("artifacts source catalog must contain exactly 8 surfaces")
	}
	catalog := artifactSourceCatalog{Surfaces: make([]artifactSourceSurface, 0, len(surfaceObjects))}
	seenViews := map[string]struct{}{}
	seenTypes := map[string]struct{}{}
	seenFields := map[string]string{}
	previousView := ""
	directCount := 0
	collectionCount := 0
	for surfaceIndex, surfaceObject := range surfaceObjects {
		label := fmt.Sprintf("%s.surfaces[%d]", artifactSourceCatalogPath, surfaceIndex+1)
		if err := requireAllowedKeys(surfaceObject, artifactSurfaceKeys, label); err != nil {
			return artifactSourceCatalog{}, err
		}
		viewSchemaID, err := requiredString(surfaceObject, "view_schema_id", label)
		if err != nil {
			return artifactSourceCatalog{}, err
		}
		if previousView != "" && previousView >= viewSchemaID {
			return artifactSourceCatalog{}, fmt.Errorf("artifacts source catalog surfaces must be unique and sorted by view_schema_id")
		}
		previousView = viewSchemaID
		if _, duplicate := seenViews[viewSchemaID]; duplicate {
			return artifactSourceCatalog{}, fmt.Errorf("duplicate artifact view schema %s", viewSchemaID)
		}
		seenViews[viewSchemaID] = struct{}{}
		artifactType, err := requiredString(surfaceObject, "artifact_type", label)
		if err != nil {
			return artifactSourceCatalog{}, err
		}
		if !artifactIdentifierPattern.MatchString(artifactType) {
			return artifactSourceCatalog{}, fmt.Errorf("%s.artifact_type is not a safe identifier", label)
		}
		if _, duplicate := seenTypes[artifactType]; duplicate {
			return artifactSourceCatalog{}, fmt.Errorf("duplicate artifact type %s", artifactType)
		}
		seenTypes[artifactType] = struct{}{}
		directObjects, err := objectArrayAllowEmpty(surfaceObject["direct_fields"], label+".direct_fields")
		if err != nil {
			return artifactSourceCatalog{}, err
		}
		collectionObjects, err := objectArrayAllowEmpty(surfaceObject["collection_fields"], label+".collection_fields")
		if err != nil {
			return artifactSourceCatalog{}, err
		}
		surface := artifactSourceSurface{ViewSchemaID: viewSchemaID, ArtifactType: artifactType}
		previousField := ""
		for fieldIndex, fieldObject := range directObjects {
			fieldLabel := fmt.Sprintf("%s.direct_fields[%d]", label, fieldIndex+1)
			if err := requireAllowedKeys(fieldObject, artifactDirectFieldKeys, fieldLabel); err != nil {
				return artifactSourceCatalog{}, err
			}
			fieldKey, err := requiredString(fieldObject, "field_key", fieldLabel)
			if err != nil {
				return artifactSourceCatalog{}, err
			}
			if previousField != "" && previousField >= fieldKey {
				return artifactSourceCatalog{}, fmt.Errorf("%s.direct_fields must be unique and sorted by field_key", label)
			}
			previousField = fieldKey
			if err := registerArtifactField(seenFields, fieldKey, viewSchemaID); err != nil {
				return artifactSourceCatalog{}, err
			}
			table, err := requiredString(fieldObject, "table", fieldLabel)
			if err != nil {
				return artifactSourceCatalog{}, err
			}
			if _, owned := artifactOwnedRelations[table]; !owned {
				return artifactSourceCatalog{}, fmt.Errorf("%s.table references unowned relation %s", fieldLabel, table)
			}
			column, err := requiredString(fieldObject, "column", fieldLabel)
			if err != nil {
				return artifactSourceCatalog{}, err
			}
			if !artifactIdentifierPattern.MatchString(column) {
				return artifactSourceCatalog{}, fmt.Errorf("%s.column is not a safe identifier", fieldLabel)
			}
			surface.DirectFields = append(surface.DirectFields, artifactDirectField{FieldKey: fieldKey, Table: table, Column: column})
			directCount++
		}
		previousField = ""
		for fieldIndex, fieldObject := range collectionObjects {
			fieldLabel := fmt.Sprintf("%s.collection_fields[%d]", label, fieldIndex+1)
			if err := requireAllowedKeys(fieldObject, artifactCollectionFieldKeys, fieldLabel); err != nil {
				return artifactSourceCatalog{}, err
			}
			fieldKey, err := requiredString(fieldObject, "field_key", fieldLabel)
			if err != nil {
				return artifactSourceCatalog{}, err
			}
			if previousField != "" && previousField >= fieldKey {
				return artifactSourceCatalog{}, fmt.Errorf("%s.collection_fields must be unique and sorted by field_key", label)
			}
			previousField = fieldKey
			if err := registerArtifactField(seenFields, fieldKey, viewSchemaID); err != nil {
				return artifactSourceCatalog{}, err
			}
			family, err := requiredString(fieldObject, "collection_family", fieldLabel)
			if err != nil {
				return artifactSourceCatalog{}, err
			}
			if _, known := artifactCollectionFamilies[family]; !known {
				return artifactSourceCatalog{}, fmt.Errorf("%s.collection_family is unknown", fieldLabel)
			}
			operations, err := stringArray(fieldObject["allowed_operations"], fieldLabel+".allowed_operations", true)
			if err != nil {
				return artifactSourceCatalog{}, err
			}
			if strings.Join(operations, "\x00") != strings.Join(artifactOperationsByFamily[family], "\x00") {
				return artifactSourceCatalog{}, fmt.Errorf("%s.allowed_operations do not match collection family %s", fieldLabel, family)
			}
			linkType, err := optionalArtifactString(fieldObject, "link_type", fieldLabel)
			if err != nil {
				return artifactSourceCatalog{}, err
			}
			targetType, err := optionalArtifactString(fieldObject, "expected_target_record_type", fieldLabel)
			if err != nil {
				return artifactSourceCatalog{}, err
			}
			if err := validateArtifactCollectionSemantics(fieldLabel, family, linkType, targetType); err != nil {
				return artifactSourceCatalog{}, err
			}
			surface.CollectionFields = append(surface.CollectionFields, artifactCollectionField{
				FieldKey: fieldKey, CollectionFamily: family, AllowedOperations: operations,
				LinkType: linkType, ExpectedTargetRecordType: targetType,
			})
			collectionCount++
		}
		catalog.Surfaces = append(catalog.Surfaces, surface)
	}
	if directCount != 36 || collectionCount != 15 {
		return artifactSourceCatalog{}, fmt.Errorf("artifacts source catalog counts are %d direct and %d collection fields; want 36 and 15", directCount, collectionCount)
	}
	return catalog, nil
}

func registerArtifactField(seen map[string]string, fieldKey, viewSchemaID string) error {
	if !artifactFieldKeyPattern.MatchString(fieldKey) {
		return fmt.Errorf("artifact field key %s is invalid", fieldKey)
	}
	if previous, duplicate := seen[fieldKey]; duplicate {
		return fmt.Errorf("artifact field %s is owned by both %s and %s", fieldKey, previous, viewSchemaID)
	}
	seen[fieldKey] = viewSchemaID
	return nil
}

func optionalArtifactString(object map[string]any, key, label string) (string, error) {
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

func validateArtifactCollectionSemantics(label, family, linkType, targetType string) error {
	switch family {
	case "record_ref":
		if _, known := artifactLinkTypes[linkType]; !known {
			return fmt.Errorf("%s.link_type is required for record_ref", label)
		}
		if targetType != "" {
			if _, known := artifactTargetRecordTypes[targetType]; !known || targetType == "party" {
				return fmt.Errorf("%s.expected_target_record_type is invalid for record_ref", label)
			}
		}
	case "party_ref":
		if linkType != "references_record" || targetType != "party" {
			return fmt.Errorf("%s party_ref requires references_record and party", label)
		}
	case "record_tag", "risk_ref":
		if linkType != "" || targetType != "" {
			return fmt.Errorf("%s %s must not declare link or target record facts", label, family)
		}
	}
	return nil
}

func validateArtifactSourceCatalogFamily(root string) error {
	base := filepath.Join(root, "contracts", "artifacts")
	entries, err := os.ReadDir(base)
	if err != nil {
		return fmt.Errorf("read Artifacts contract family: %w", err)
	}
	expected := stringSet("source-catalog.v1.json", "source-catalog.v1.schema.json")
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if _, ok := expected[entry.Name()]; !ok {
			return fmt.Errorf("unexpected Artifacts contract artifact %s", entry.Name())
		}
		delete(expected, entry.Name())
	}
	if len(expected) != 0 {
		return fmt.Errorf("artifacts contract family is incomplete")
	}
	_, err = loadArtifactSourceCatalog(root)
	return err
}

func loadArtifactSourceCatalog(root string) (artifactSourceCatalog, error) {
	data, err := os.ReadFile(filepath.Join(root, artifactSourceCatalogPath))
	if err != nil {
		return artifactSourceCatalog{}, err
	}
	value, err := decodeContract(data)
	if err != nil {
		return artifactSourceCatalog{}, err
	}
	catalog, err := parseArtifactSourceCatalog(value)
	if err != nil {
		return artifactSourceCatalog{}, err
	}
	for surfaceIndex := range catalog.Surfaces {
		if err := enrichArtifactSurfaceFromViewSchema(root, &catalog.Surfaces[surfaceIndex]); err != nil {
			return artifactSourceCatalog{}, err
		}
	}
	return catalog, nil
}

func enrichArtifactSurfaceFromViewSchema(root string, surface *artifactSourceSurface) error {
	path := filepath.Join(root, "contracts", "view-schemas", surface.ViewSchemaID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("artifact surface %s view schema is missing: %w", surface.ViewSchemaID, err)
	}
	value, err := decodeContract(data)
	if err != nil {
		return err
	}
	document, err := asObject(value, surface.ViewSchemaID)
	if err != nil {
		return err
	}
	if base, err := requiredString(document, "base_projection", surface.ViewSchemaID); err != nil {
		return err
	} else if base != "artifact_grid_projection" {
		return fmt.Errorf("artifact surface %s base projection is %s", surface.ViewSchemaID, base)
	}
	sourceTypes, err := stringArray(document["source_record_types"], surface.ViewSchemaID+".source_record_types", true)
	if err != nil || len(sourceTypes) != 1 || sourceTypes[0] != "artifact" {
		return fmt.Errorf("artifact surface %s source record types must be exactly artifact", surface.ViewSchemaID)
	}
	filter, err := asObject(document["canonical_source_filter"], surface.ViewSchemaID+".canonical_source_filter")
	if err != nil {
		return err
	}
	for key, want := range map[string]string{"kind": "artifact_type", "field": "artifact_type", "value": surface.ArtifactType} {
		got, err := requiredString(filter, key, surface.ViewSchemaID+".canonical_source_filter")
		if err != nil || got != want {
			return fmt.Errorf("artifact surface %s source filter %s must be %s", surface.ViewSchemaID, key, want)
		}
	}
	fieldObjects, err := objectArray(document["fields"], surface.ViewSchemaID+".fields")
	if err != nil {
		return err
	}
	viewFields := map[string]artifactViewFieldFacts{}
	for index, object := range fieldObjects {
		label := fmt.Sprintf("%s.fields[%d]", surface.ViewSchemaID, index+1)
		fieldKey, err := requiredString(object, "field_key", label)
		if err != nil {
			return err
		}
		writable, err := artifactOptionalBool(object, "writable", label)
		if err != nil {
			return err
		}
		createWritable, err := artifactOptionalBool(object, "create_writable", label)
		if err != nil {
			return err
		}
		if !writable && !createWritable {
			continue
		}
		readKind, err := requiredString(object, "read_kind", label)
		if err != nil {
			return err
		}
		writeKind, err := requiredString(object, "write_kind", label)
		if err != nil {
			return err
		}
		clearable, err := requiredBool(object, "clearable", label)
		if err != nil {
			return err
		}
		reference, err := nullableString(object, "direct_reference_contract_id", label)
		if err != nil {
			return err
		}
		viewFields[fieldKey] = artifactViewFieldFacts{
			ReadKind: readKind, WriteKind: writeKind, Writable: writable,
			CreateWritable: createWritable, Clearable: clearable, ReferenceContractID: reference,
		}
	}
	seen := map[string]struct{}{}
	for index := range surface.DirectFields {
		field := &surface.DirectFields[index]
		facts, ok := viewFields[field.FieldKey]
		if !ok {
			return fmt.Errorf("artifact catalog direct field %s is missing or read-only in %s", field.FieldKey, surface.ViewSchemaID)
		}
		if facts.WriteKind != "direct_value" {
			return fmt.Errorf("artifact catalog direct field %s has write kind %s", field.FieldKey, facts.WriteKind)
		}
		if strings.HasSuffix(field.Column, "_user_id") {
			if facts.ReferenceContractID != "incident_member_user_ref_v1" || facts.Clearable {
				return fmt.Errorf("artifact member reference %s has mismatched reference or nullability contract", field.FieldKey)
			}
		} else if facts.ReferenceContractID != "" {
			return fmt.Errorf("artifact scalar %s has unexpected reference contract %s", field.FieldKey, facts.ReferenceContractID)
		}
		field.View = facts
		seen[field.FieldKey] = struct{}{}
	}
	for index := range surface.CollectionFields {
		field := &surface.CollectionFields[index]
		facts, ok := viewFields[field.FieldKey]
		if !ok {
			return fmt.Errorf("artifact catalog collection field %s is missing or read-only in %s", field.FieldKey, surface.ViewSchemaID)
		}
		if facts.WriteKind != "action_payload" || facts.ReadKind != "collection" || facts.ReferenceContractID != "" || facts.Clearable {
			return fmt.Errorf("artifact collection field %s has mismatched write, reference, or nullability contract", field.FieldKey)
		}
		field.View = facts
		seen[field.FieldKey] = struct{}{}
	}
	for fieldKey := range viewFields {
		if _, ok := seen[fieldKey]; !ok {
			return fmt.Errorf("artifact writable field %s is missing from source catalog surface %s", fieldKey, surface.ViewSchemaID)
		}
	}
	return nil
}

func artifactOptionalBool(object map[string]any, key, label string) (bool, error) {
	value, present := object[key]
	if !present {
		return false, nil
	}
	boolean, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("%s.%s must be a boolean when present", label, key)
	}
	return boolean, nil
}

func writeArtifactSourceCatalogGo(root string) error {
	catalog, err := loadArtifactSourceCatalog(root)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	buffer.WriteString("// Code generated by tools/contractgen; DO NOT EDIT.\n\npackage contractartifacts\n\n")
	buffer.WriteString("type ViewFieldFacts struct {\n\tReadKind string\n\tWriteKind string\n\tWritable bool\n\tCreateWritable bool\n\tClearable bool\n\tReferenceContractID string\n}\n\n")
	buffer.WriteString("type DirectField struct {\n\tFieldKey string\n\tTable string\n\tColumn string\n\tView ViewFieldFacts\n}\n\n")
	buffer.WriteString("type CollectionField struct {\n\tFieldKey string\n\tCollectionFamily string\n\tAllowedOperations []string\n\tLinkType string\n\tExpectedTargetRecordType string\n\tView ViewFieldFacts\n}\n\n")
	buffer.WriteString("type SourceSurface struct {\n\tViewSchemaID string\n\tArtifactType string\n\tDirectFields []DirectField\n\tCollectionFields []CollectionField\n}\n\n")
	buffer.WriteString("const (\n")
	for _, surface := range catalog.Surfaces {
		fmt.Fprintf(&buffer, "\t%sViewSchemaID = %s\n", artifactGoName(surface.ArtifactType), strconv.Quote(surface.ViewSchemaID))
	}
	buffer.WriteString(")\n\nvar SourceCatalog = []SourceSurface{\n")
	for _, surface := range catalog.Surfaces {
		fmt.Fprintf(&buffer, "\t{ViewSchemaID: %s, ArtifactType: %s,\n", strconv.Quote(surface.ViewSchemaID), strconv.Quote(surface.ArtifactType))
		buffer.WriteString("\t\tDirectFields: []DirectField{\n")
		for _, field := range surface.DirectFields {
			fmt.Fprintf(&buffer, "\t\t\t{FieldKey: %s, Table: %s, Column: %s, View: ", strconv.Quote(field.FieldKey), strconv.Quote(field.Table), strconv.Quote(field.Column))
			writeArtifactViewFacts(&buffer, field.View)
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
			writeArtifactViewFacts(&buffer, field.View)
			buffer.WriteString("},\n")
		}
		buffer.WriteString("\t\t},\n\t},\n")
	}
	buffer.WriteString("}\n")
	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("format generated Artifacts source catalog: %w", err)
	}
	return stageGeneratedFile(filepath.Join(root, "internal", "gen", "contractartifacts", "source_catalog_gen.go"), formatted, 0o644)
}

func writeArtifactViewFacts(buffer *bytes.Buffer, facts artifactViewFieldFacts) {
	fmt.Fprintf(buffer, "ViewFieldFacts{ReadKind: %s, WriteKind: %s, Writable: %t, CreateWritable: %t, Clearable: %t, ReferenceContractID: %s}", strconv.Quote(facts.ReadKind), strconv.Quote(facts.WriteKind), facts.Writable, facts.CreateWritable, facts.Clearable, strconv.Quote(facts.ReferenceContractID))
}

func artifactGoName(value string) string {
	parts := strings.Split(value, "_")
	for index := range parts {
		parts[index] = strings.ToUpper(parts[index][:1]) + parts[index][1:]
	}
	return strings.Join(parts, "")
}
