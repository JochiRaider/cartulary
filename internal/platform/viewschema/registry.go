package viewschema

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

type Field struct {
	FieldKey                  string   `json:"field_key"`
	Label                     string   `json:"label"`
	DefaultHidden             bool     `json:"default_hidden"`
	Sortable                  bool     `json:"sortable"`
	Groupable                 bool     `json:"groupable"`
	ReadKind                  string   `json:"read_kind"`
	WriteKind                 string   `json:"write_kind"`
	Writable                  bool     `json:"writable"`
	CreateWritable            bool     `json:"create_writable"`
	HeaderSortFieldKey        *string  `json:"header_sort_field_key"`
	FilterOps                 []string `json:"filter_ops"`
	ConflictResolutionClass   string   `json:"conflict_resolution_class"`
	EntityBindingMode         *string  `json:"entity_binding_mode"`
	StringContractID          *string  `json:"string_contract_id"`
	DirectScalarContractID    *string  `json:"direct_scalar_contract_id"`
	DirectReferenceContractID *string  `json:"direct_reference_contract_id"`
	WriteTarget               *string  `json:"write_target"`
	WriteAction               *string  `json:"write_action"`
	Clearable                 bool     `json:"clearable"`
	EnumValues                []string `json:"enum_values"`
}

type SortEntry struct {
	FieldKey  string `json:"field_key"`
	Direction string `json:"direction"`
}

type inlineCreate struct {
	PermitsZeroFieldCreate bool `json:"permits_zero_field_create"`
}

type SyntheticFilterPredicate struct {
	FieldKey  string   `json:"field_key"`
	Label     string   `json:"label"`
	FilterOps []string `json:"filter_ops"`
}

type schemaDocument struct {
	ViewSchemaID              string                     `json:"view_schema_id"`
	Title                     string                     `json:"title"`
	SurfaceKind               string                     `json:"surface_kind"`
	SourceRecordTypes         []string                   `json:"source_record_types"`
	TechnicalFields           []string                   `json:"technical_fields"`
	RequiredReferencePackKeys []string                   `json:"required_reference_pack_keys"`
	DefaultSort               []SortEntry                `json:"default_sort"`
	SortFields                []string                   `json:"sort_fields"`
	FilterFields              []string                   `json:"filter_fields"`
	SyntheticFilterPredicates []SyntheticFilterPredicate `json:"synthetic_filter_predicates"`
	GroupingFields            []string                   `json:"grouping_fields"`
	InlineCreate              inlineCreate               `json:"inline_create"`
	Fields                    []Field                    `json:"fields"`
}

type registryIndex struct {
	ViewSchemas []registryIndexEntry `json:"view_schemas"`
}

type registryIndexEntry struct {
	ViewSchemaID string `json:"view_schema_id"`
	ArtifactPath string `json:"artifact_path"`
}

type Schema struct {
	ViewSchemaID           string
	PermitsZeroFieldCreate bool
	defaultSort            []SortEntry
	sortFields             []string
	filterFields           []string
	groupingFields         []string
	fields                 map[string]Field
	public                 ViewSchemaResource
}

type ViewFieldEntry struct {
	FieldKey                  string   `json:"field_key"`
	Label                     string   `json:"label"`
	DefaultHidden             bool     `json:"default_hidden"`
	Sortable                  bool     `json:"sortable"`
	HeaderSortFieldKey        *string  `json:"header_sort_field_key"`
	FilterOps                 []string `json:"filter_ops"`
	Groupable                 bool     `json:"groupable"`
	ReadKind                  string   `json:"read_kind"`
	WriteKind                 string   `json:"write_kind"`
	ConflictResolutionClass   *string  `json:"conflict_resolution_class"`
	EntityBindingMode         *string  `json:"entity_binding_mode"`
	StringContractID          *string  `json:"string_contract_id"`
	DirectScalarContractID    *string  `json:"direct_scalar_contract_id"`
	DirectReferenceContractID *string  `json:"direct_reference_contract_id"`
	Clearable                 bool     `json:"clearable"`
	EnumValues                []string `json:"enum_values"`
}

type ViewSchemaResource struct {
	ViewSchemaID              string                     `json:"view_schema_id"`
	SurfaceKind               string                     `json:"surface_kind"`
	Title                     string                     `json:"title"`
	SourceRecordTypes         []string                   `json:"source_record_types"`
	TechnicalFields           []string                   `json:"technical_fields"`
	RequiredReferencePackKeys []string                   `json:"required_reference_pack_keys"`
	DefaultSort               []SortEntry                `json:"default_sort"`
	SortFields                []string                   `json:"sort_fields"`
	FilterFields              []string                   `json:"filter_fields"`
	SyntheticFilterPredicates []SyntheticFilterPredicate `json:"synthetic_filter_predicates"`
	GroupingFields            []string                   `json:"grouping_fields"`
	Fields                    []ViewFieldEntry           `json:"fields"`
}

func (s Schema) Fields() map[string]Field {
	cloned := make(map[string]Field, len(s.fields))
	for key, field := range s.fields {
		cloned[key] = field
	}
	return cloned
}

func (s Schema) DefaultSort() []SortEntry {
	cloned := make([]SortEntry, len(s.defaultSort))
	copy(cloned, s.defaultSort)
	return cloned
}

func (s Schema) SortFields() []string {
	cloned := make([]string, len(s.sortFields))
	copy(cloned, s.sortFields)
	return cloned
}

func (s Schema) FilterFields() []string {
	cloned := make([]string, len(s.filterFields))
	copy(cloned, s.filterFields)
	return cloned
}

func (s Schema) GroupingFields() []string {
	cloned := make([]string, len(s.groupingFields))
	copy(cloned, s.groupingFields)
	return cloned
}

var (
	loadOnce sync.Once
	schemas  map[string]Schema
	ordered  []Schema
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

func ListPublicResources() []ViewSchemaResource {
	loadRegistry()
	resources := make([]ViewSchemaResource, 0, len(ordered))
	for _, schema := range ordered {
		resources = append(resources, cloneResource(schema.public))
	}
	return resources
}

func LookupPublicResource(viewSchemaID string) (ViewSchemaResource, bool) {
	schema, ok := Lookup(viewSchemaID)
	if !ok {
		return ViewSchemaResource{}, false
	}
	return cloneResource(schema.public), true
}

func loadRegistry() {
	loadOnce.Do(func() {
		schemas = make(map[string]Schema)
		artifactsByPath := make(map[string]string, len(gencontracts.ViewSchemaArtifacts))
		var index registryIndex
		indexLoaded := false
		for _, artifact := range gencontracts.ViewSchemaArtifacts {
			if !strings.HasSuffix(artifact.Path, ".json") {
				continue
			}
			if strings.HasSuffix(artifact.Path, "/index.json") {
				if err := json.Unmarshal([]byte(artifact.JSON), &index); err != nil {
					panic(fmt.Sprintf("viewschema: load registry index %s: %v", artifact.Path, err))
				}
				indexLoaded = true
				continue
			}
			artifactsByPath[artifact.Path] = artifact.JSON
		}
		if !indexLoaded {
			panic("viewschema: missing registry index")
		}

		for _, entry := range index.ViewSchemas {
			payload, ok := artifactsByPath[entry.ArtifactPath]
			if !ok {
				panic(fmt.Sprintf("viewschema: indexed artifact %s not embedded", entry.ArtifactPath))
			}
			var document schemaDocument
			if err := json.Unmarshal([]byte(payload), &document); err != nil {
				panic(fmt.Sprintf("viewschema: load %s: %v", entry.ArtifactPath, err))
			}
			if document.ViewSchemaID != entry.ViewSchemaID {
				panic(fmt.Sprintf("viewschema: index id %s does not match artifact id %s", entry.ViewSchemaID, document.ViewSchemaID))
			}
			if strings.TrimSpace(document.ViewSchemaID) == "" {
				panic(fmt.Sprintf("viewschema: %s has empty view_schema_id", entry.ArtifactPath))
			}
			if _, exists := schemas[document.ViewSchemaID]; exists {
				panic(fmt.Sprintf("viewschema: duplicate view_schema_id %s", document.ViewSchemaID))
			}

			fieldIndex := make(map[string]Field, len(document.Fields))
			for _, field := range document.Fields {
				if field.FilterOps == nil {
					field.FilterOps = []string{}
				}
				fieldIndex[field.FieldKey] = field
			}

			schemas[document.ViewSchemaID] = Schema{
				ViewSchemaID:           document.ViewSchemaID,
				PermitsZeroFieldCreate: document.InlineCreate.PermitsZeroFieldCreate,
				defaultSort:            append([]SortEntry(nil), document.DefaultSort...),
				sortFields:             append([]string(nil), document.SortFields...),
				filterFields:           append([]string(nil), document.FilterFields...),
				groupingFields:         append([]string(nil), document.GroupingFields...),
				fields:                 fieldIndex,
				public:                 buildPublicResource(document),
			}
		}

		ordered = make([]Schema, 0, len(schemas))
		for _, schema := range schemas {
			ordered = append(ordered, schema)
		}
		slices.SortFunc(ordered, func(left Schema, right Schema) int {
			return strings.Compare(left.ViewSchemaID, right.ViewSchemaID)
		})
	})
}

func buildPublicResource(document schemaDocument) ViewSchemaResource {
	fields := make([]ViewFieldEntry, 0, len(document.Fields))
	for _, field := range document.Fields {
		fields = append(fields, ViewFieldEntry{
			FieldKey:                  field.FieldKey,
			Label:                     field.Label,
			DefaultHidden:             field.DefaultHidden,
			Sortable:                  field.Sortable,
			HeaderSortFieldKey:        field.HeaderSortFieldKey,
			FilterOps:                 cloneStrings(field.FilterOps),
			Groupable:                 field.Groupable,
			ReadKind:                  field.ReadKind,
			WriteKind:                 field.WriteKind,
			ConflictResolutionClass:   nullableString(field.ConflictResolutionClass),
			EntityBindingMode:         cloneStringPointer(field.EntityBindingMode),
			StringContractID:          cloneStringPointer(field.StringContractID),
			DirectScalarContractID:    cloneStringPointer(field.DirectScalarContractID),
			DirectReferenceContractID: cloneStringPointer(field.DirectReferenceContractID),
			Clearable:                 field.Clearable,
			EnumValues:                cloneNullableStrings(field.EnumValues),
		})
	}
	return ViewSchemaResource{
		ViewSchemaID:              document.ViewSchemaID,
		SurfaceKind:               document.SurfaceKind,
		Title:                     document.Title,
		SourceRecordTypes:         cloneStrings(document.SourceRecordTypes),
		TechnicalFields:           cloneStrings(document.TechnicalFields),
		RequiredReferencePackKeys: cloneStrings(document.RequiredReferencePackKeys),
		DefaultSort:               cloneSortEntries(document.DefaultSort),
		SortFields:                cloneStrings(document.SortFields),
		FilterFields:              cloneStrings(document.FilterFields),
		SyntheticFilterPredicates: cloneSyntheticFilterPredicates(document.SyntheticFilterPredicates),
		GroupingFields:            cloneStrings(document.GroupingFields),
		Fields:                    fields,
	}
}

func cloneResource(resource ViewSchemaResource) ViewSchemaResource {
	resource.SourceRecordTypes = cloneStrings(resource.SourceRecordTypes)
	resource.TechnicalFields = cloneStrings(resource.TechnicalFields)
	resource.RequiredReferencePackKeys = cloneStrings(resource.RequiredReferencePackKeys)
	resource.DefaultSort = cloneSortEntries(resource.DefaultSort)
	resource.SortFields = cloneStrings(resource.SortFields)
	resource.FilterFields = cloneStrings(resource.FilterFields)
	resource.SyntheticFilterPredicates = cloneSyntheticFilterPredicates(resource.SyntheticFilterPredicates)
	resource.GroupingFields = cloneStrings(resource.GroupingFields)
	resource.Fields = cloneViewFieldEntries(resource.Fields)
	return resource
}

func cloneViewFieldEntries(fields []ViewFieldEntry) []ViewFieldEntry {
	cloned := make([]ViewFieldEntry, len(fields))
	for index, field := range fields {
		cloned[index] = field
		cloned[index].HeaderSortFieldKey = cloneStringPointer(field.HeaderSortFieldKey)
		cloned[index].FilterOps = cloneStrings(field.FilterOps)
		cloned[index].ConflictResolutionClass = cloneStringPointer(field.ConflictResolutionClass)
		cloned[index].EntityBindingMode = cloneStringPointer(field.EntityBindingMode)
		cloned[index].StringContractID = cloneStringPointer(field.StringContractID)
		cloned[index].DirectScalarContractID = cloneStringPointer(field.DirectScalarContractID)
		cloned[index].DirectReferenceContractID = cloneStringPointer(field.DirectReferenceContractID)
		cloned[index].EnumValues = cloneNullableStrings(field.EnumValues)
	}
	return cloned
}

func cloneSyntheticFilterPredicates(predicates []SyntheticFilterPredicate) []SyntheticFilterPredicate {
	cloned := make([]SyntheticFilterPredicate, len(predicates))
	for index, predicate := range predicates {
		cloned[index] = predicate
		cloned[index].FilterOps = cloneStrings(predicate.FilterOps)
	}
	return cloned
}

func cloneSortEntries(entries []SortEntry) []SortEntry {
	cloned := make([]SortEntry, len(entries))
	copy(cloned, entries)
	return cloned
}

func cloneStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func cloneNullableStrings(values []string) []string {
	if values == nil {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
