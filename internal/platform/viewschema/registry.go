package viewschema

import (
	"encoding/json"
	"strings"
	"sync"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

type Field struct {
	FieldKey                string  `json:"field_key"`
	Label                   string  `json:"label"`
	Writable                bool    `json:"writable"`
	CreateWritable          bool    `json:"create_writable"`
	ConflictResolutionClass string  `json:"conflict_resolution_class"`
	EntityBindingMode       *string `json:"entity_binding_mode"`
	WriteTarget             *string `json:"write_target"`
	WriteAction             *string `json:"write_action"`
}

type inlineCreate struct {
	PermitsZeroFieldCreate bool `json:"permits_zero_field_create"`
}

type schemaDocument struct {
	ViewSchemaID string       `json:"view_schema_id"`
	InlineCreate inlineCreate `json:"inline_create"`
	Fields       []Field      `json:"fields"`
}

type Schema struct {
	ViewSchemaID           string
	PermitsZeroFieldCreate bool
	fields                 map[string]Field
}

func (s Schema) Fields() map[string]Field {
	cloned := make(map[string]Field, len(s.fields))
	for key, field := range s.fields {
		cloned[key] = field
	}
	return cloned
}

var (
	loadOnce sync.Once
	schemas  map[string]Schema
)

func Lookup(viewSchemaID string) (Schema, bool) {
	loadRegistry()
	schema, ok := schemas[viewSchemaID]
	return schema, ok
}

func LookupField(viewSchemaID string, fieldKey string) (Field, bool) {
	schema, ok := Lookup(viewSchemaID)
	if !ok {
		return Field{}, false
	}
	field, ok := schema.fields[fieldKey]
	return field, ok
}

func loadRegistry() {
	loadOnce.Do(func() {
		schemas = make(map[string]Schema)
		for _, artifact := range gencontracts.ViewSchemaArtifacts {
			if !strings.HasSuffix(artifact.Path, ".json") || strings.HasSuffix(artifact.Path, "/index.json") {
				continue
			}

			var document schemaDocument
			if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
				continue
			}
			if strings.TrimSpace(document.ViewSchemaID) == "" {
				continue
			}

			fieldIndex := make(map[string]Field, len(document.Fields))
			for _, field := range document.Fields {
				fieldIndex[field.FieldKey] = field
			}

			schemas[document.ViewSchemaID] = Schema{
				ViewSchemaID:           document.ViewSchemaID,
				PermitsZeroFieldCreate: document.InlineCreate.PermitsZeroFieldCreate,
				fields:                 fieldIndex,
			}
		}
	})
}
