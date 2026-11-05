// Package extensioncodec implements logical row encoding and binding proofs
// from admitted declarations. It has no profile callbacks or database access.
package extensioncodec

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const (
	PostgresMappingSchemaID = "cartulary.extension_postgres_binding_mapping.v1"
	NetworkFlowRowsSchemaID = "cartulary.network_flow_activity.postgres_rows.v2"
)

var (
	ErrInvalidContract   = errors.New("extension logical codec contract is invalid")
	ErrInvalidEntry      = errors.New("extension logical entry is invalid")
	ErrLimitExceeded     = errors.New("extension binding proof limit exceeded")
	ErrDuplicateIdentity = errors.New("extension binding contains duplicate logical identities")
	identifierPattern    = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)
	logicalIDPattern     = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)
)

type Definition struct {
	SchemaID  string   `json:"schema_id"`
	ProfileID string   `json:"profile_id"`
	Families  []Family `json:"families"`
}

type Family struct {
	LogicalFamilyID  string   `json:"logical_family_id"`
	IdentityFieldIDs []string `json:"identity_field_ids"`
	Fields           []Field  `json:"fields"`
}

type Field struct {
	FieldID   string `json:"field_id"`
	ValueKind string `json:"value_kind"`
	Nullable  bool   `json:"nullable"`
}

type Mapping struct {
	SchemaID         string   `json:"schema_id"`
	BindingID        string   `json:"binding_id"`
	CodecContractRef string   `json:"codec_contract_ref"`
	SchemaName       string   `json:"schema_name"`
	TableName        string   `json:"table_name"`
	Columns          []Column `json:"columns"`
}

type Column struct {
	FieldID    string `json:"field_id"`
	ColumnName string `json:"column_name"`
}

func ParseDefinition(raw []byte) (Definition, error) {
	var definition Definition
	value, err := parseDocument(raw)
	if err != nil {
		return definition, err
	}
	root, ok := closedObject(value, "schema_id", "profile_id", "families")
	if !ok {
		return definition, ErrInvalidContract
	}
	families, ok := root["families"].([]any)
	if !ok {
		return definition, ErrInvalidContract
	}
	for _, item := range families {
		family, ok := closedObject(item, "logical_family_id", "identity_field_ids", "fields")
		if !ok {
			return definition, ErrInvalidContract
		}
		fields, ok := family["fields"].([]any)
		if !ok {
			return definition, ErrInvalidContract
		}
		for _, item := range fields {
			field, ok := closedObject(item, "field_id", "value_kind", "nullable")
			if !ok {
				return definition, ErrInvalidContract
			}
			if _, ok := field["nullable"].(bool); !ok {
				return definition, ErrInvalidContract
			}
		}
	}
	if json.Unmarshal(raw, &definition) != nil || definition.validate() != nil {
		return Definition{}, ErrInvalidContract
	}
	return definition, nil
}

func ParseMapping(raw []byte) (Mapping, error) {
	var mapping Mapping
	value, err := parseDocument(raw)
	if err != nil {
		return mapping, err
	}
	root, ok := closedObject(value, "schema_id", "binding_id", "codec_contract_ref", "schema_name", "table_name", "columns")
	if !ok {
		return mapping, ErrInvalidContract
	}
	columns, ok := root["columns"].([]any)
	if !ok {
		return mapping, ErrInvalidContract
	}
	for _, item := range columns {
		if _, ok := closedObject(item, "field_id", "column_name"); !ok {
			return mapping, ErrInvalidContract
		}
	}
	if json.Unmarshal(raw, &mapping) != nil || mapping.validate() != nil {
		return Mapping{}, ErrInvalidContract
	}
	return mapping, nil
}

func parseDocument(raw []byte) (any, error) {
	if len(raw) == 0 || len(raw) > 1048576 {
		return nil, ErrInvalidContract
	}
	value, err := canonicaljson.DecodeStrict(raw)
	if err != nil {
		return nil, ErrInvalidContract
	}
	return value, nil
}

func closedObject(value any, keys ...string) (map[string]any, bool) {
	object, ok := value.(map[string]any)
	if !ok || len(object) != len(keys) {
		return nil, false
	}
	for _, key := range keys {
		if _, present := object[key]; !present {
			return nil, false
		}
	}
	return object, true
}

func validLogicalID(value string) bool {
	return len(value) <= 191 && logicalIDPattern.MatchString(value)
}

func (definition Definition) validate() error {
	if !validLogicalID(definition.SchemaID) || !validLogicalID(definition.ProfileID) ||
		strings.Contains(definition.ProfileID, ".") || len(definition.Families) == 0 || len(definition.Families) > 64 {
		return ErrInvalidContract
	}
	previous := ""
	for _, family := range definition.Families {
		if !validLogicalID(family.LogicalFamilyID) || !strings.HasPrefix(family.LogicalFamilyID, definition.ProfileID+".") ||
			family.LogicalFamilyID <= previous || len(family.Fields) == 0 || len(family.Fields) > 4096 ||
			len(family.IdentityFieldIDs) == 0 || len(family.IdentityFieldIDs) > len(family.Fields) {
			return ErrInvalidContract
		}
		previous = family.LogicalFamilyID
		fields := make(map[string]Field, len(family.Fields))
		previousField := ""
		for _, field := range family.Fields {
			if !validLogicalID(field.FieldID) || field.FieldID <= previousField {
				return ErrInvalidContract
			}
			switch field.ValueKind {
			case "text", "uuid", "int64", "boolean", "timestamp", "text_array", "json":
			default:
				return ErrInvalidContract
			}
			previousField = field.FieldID
			fields[field.FieldID] = field
		}
		seen := make(map[string]bool, len(family.IdentityFieldIDs))
		for _, identity := range family.IdentityFieldIDs {
			field, present := fields[identity]
			if !present || field.Nullable || seen[identity] {
				return ErrInvalidContract
			}
			seen[identity] = true
		}
	}
	return nil
}

func (mapping Mapping) validate() error {
	if mapping.SchemaID != PostgresMappingSchemaID || !validLogicalID(mapping.BindingID) ||
		!validLogicalID(mapping.CodecContractRef) || !identifierPattern.MatchString(mapping.SchemaName) ||
		!identifierPattern.MatchString(mapping.TableName) || len(mapping.Columns) == 0 || len(mapping.Columns) > 4096 {
		return ErrInvalidContract
	}
	previous := ""
	seen := make(map[string]bool, len(mapping.Columns))
	for _, column := range mapping.Columns {
		if !validLogicalID(column.FieldID) || column.FieldID <= previous || !identifierPattern.MatchString(column.ColumnName) || seen[column.ColumnName] {
			return ErrInvalidContract
		}
		previous = column.FieldID
		seen[column.ColumnName] = true
	}
	return nil
}
