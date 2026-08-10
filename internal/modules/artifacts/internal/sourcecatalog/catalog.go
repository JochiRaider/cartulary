package sourcecatalog

import (
	"fmt"
	"slices"
	"sync"

	"github.com/JochiRaider/cartulary/internal/gen/contractartifacts"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type FieldKind string

const (
	FieldKindDirect     FieldKind = "direct"
	FieldKindCollection FieldKind = "collection"
)

type Surface struct {
	ViewSchemaID string
	ArtifactType string
}

type ViewFieldFacts struct {
	ReadKind            string
	WriteKind           string
	Writable            bool
	CreateWritable      bool
	Clearable           bool
	ReferenceContractID string
}

type StorageMapping struct {
	Table  string
	Column string
}

type collectionPolicy struct {
	Family             string
	LinkType           string
	ExpectedTargetType string
	AllowedOperations  []string
}

type Field struct {
	ViewSchemaID string
	FieldKey     string
	Kind         FieldKind
	View         ViewFieldFacts
	Storage      StorageMapping
	Collection   collectionPolicy
}

type Catalog struct {
	surfaces               []Surface
	projectionSurfaces     []Surface
	surfacesByViewID       map[string]Surface
	surfacesByArtifactType map[string]Surface
	fieldsByKey            map[string]Field
}

var (
	loadOnce    sync.Once
	loadCatalog *Catalog
	loadErr     error
)

// Load validates the generated Artifacts owner facts once and caches both the
// immutable catalog and any construction error. Invalid production inputs are
// reported to composition callers instead of panicking during initialization.
func Load() (*Catalog, error) {
	loadOnce.Do(func() {
		loadCatalog, loadErr = build(contractartifacts.SourceCatalog)
	})
	return loadCatalog, loadErr
}

func (c *Catalog) Surfaces() []Surface {
	if c == nil {
		return nil
	}
	return slices.Clone(c.surfaces)
}

func (c *Catalog) ProjectionSurfaces() []Surface {
	if c == nil {
		return nil
	}
	return slices.Clone(c.projectionSurfaces)
}

func (c *Catalog) SurfaceByViewID(viewSchemaID string) (Surface, bool) {
	if c == nil {
		return Surface{}, false
	}
	surface, ok := c.surfacesByViewID[viewSchemaID]
	return surface, ok
}

func (c *Catalog) SurfaceByArtifactType(artifactType string) (Surface, bool) {
	if c == nil {
		return Surface{}, false
	}
	surface, ok := c.surfacesByArtifactType[artifactType]
	return surface, ok
}

func (c *Catalog) Field(fieldKey string) (Field, bool) {
	if c == nil {
		return Field{}, false
	}
	field, ok := c.fieldsByKey[fieldKey]
	return cloneField(field), ok
}

func (c *Catalog) Fields() map[string]Field {
	if c == nil {
		return nil
	}
	result := make(map[string]Field, len(c.fieldsByKey))
	for fieldKey, field := range c.fieldsByKey {
		result[fieldKey] = cloneField(field)
	}
	return result
}

func (c *Catalog) WritableDirectStorageMappings() map[string]StorageMapping {
	result := make(map[string]StorageMapping, 36)
	if c == nil {
		return result
	}
	for fieldKey, field := range c.fieldsByKey {
		if field.Kind == FieldKindDirect {
			result[fieldKey] = field.Storage
		}
	}
	return result
}

func (c *Catalog) ConflictFieldSourceKeys() map[string]string {
	result := make(map[string]string, 36)
	for fieldKey, storage := range c.WritableDirectStorageMappings() {
		result[fieldKey] = storage.Column
	}
	return result
}

func build(generated []contractartifacts.SourceSurface) (*Catalog, error) {
	if len(generated) != 8 {
		return nil, fmt.Errorf("artifacts: source catalog must contain exactly 8 surfaces")
	}
	catalog := &Catalog{
		surfaces:               make([]Surface, 0, len(generated)),
		surfacesByViewID:       make(map[string]Surface, len(generated)),
		surfacesByArtifactType: make(map[string]Surface, len(generated)),
		fieldsByKey:            make(map[string]Field),
	}
	directCount := 0
	collectionCount := 0
	for _, registered := range generated {
		viewSchemaID := registered.ViewSchemaID
		if viewSchemaID == "" || registered.ArtifactType == "" {
			return nil, fmt.Errorf("artifacts: source catalog contains an incomplete surface")
		}
		if _, duplicate := catalog.surfacesByViewID[viewSchemaID]; duplicate {
			return nil, fmt.Errorf("artifacts: duplicate source surface %s", viewSchemaID)
		}
		if _, duplicate := catalog.surfacesByArtifactType[registered.ArtifactType]; duplicate {
			return nil, fmt.Errorf("artifacts: duplicate artifact type %s", registered.ArtifactType)
		}
		schema, ok := viewschema.Lookup(viewSchemaID)
		if !ok {
			return nil, fmt.Errorf("artifacts: source policy surface %s is missing its view schema", viewSchemaID)
		}
		filter, ok := schema.CanonicalSourceFilter()
		if schema.BaseProjection != "artifact_grid_projection" || !ok || filter.Kind != "artifact_type" || filter.Field != "artifact_type" || filter.Value != registered.ArtifactType {
			return nil, fmt.Errorf("artifacts: source policy surface %s mismatches its view schema source filter", viewSchemaID)
		}
		surface := Surface{ViewSchemaID: viewSchemaID, ArtifactType: registered.ArtifactType}
		viewFields := schema.Fields()
		ownedFields := make(map[string]struct{}, len(registered.DirectFields)+len(registered.CollectionFields))
		for _, generatedField := range registered.DirectFields {
			fieldKey := generatedField.FieldKey
			field, ok := viewFields[fieldKey]
			if !ok || (!field.Writable && !field.CreateWritable) || field.WriteKind != "direct_value" {
				return nil, fmt.Errorf("artifacts: catalog direct field %s is missing, read-only, or mismatched", fieldKey)
			}
			if !viewFieldFactsMatch(generatedField.View, field) {
				return nil, fmt.Errorf("artifacts: generated view facts for direct field %s are stale", fieldKey)
			}
			if !validStorageMapping(generatedField.Table, generatedField.Column) {
				return nil, fmt.Errorf("artifacts: direct field %s has an incomplete storage mapping", fieldKey)
			}
			catalogField := Field{
				ViewSchemaID: viewSchemaID,
				FieldKey:     fieldKey,
				Kind:         FieldKindDirect,
				View:         viewFacts(generatedField.View),
				Storage:      StorageMapping{Table: generatedField.Table, Column: generatedField.Column},
			}
			if _, duplicate := catalog.fieldsByKey[fieldKey]; duplicate {
				return nil, fmt.Errorf("artifacts: source field %s is owned by multiple surfaces", fieldKey)
			}
			ownedFields[fieldKey] = struct{}{}
			catalog.fieldsByKey[fieldKey] = catalogField
			directCount++
		}
		for _, generatedField := range registered.CollectionFields {
			fieldKey := generatedField.FieldKey
			field, ok := viewFields[fieldKey]
			if !ok || (!field.Writable && !field.CreateWritable) || field.WriteKind != "action_payload" {
				return nil, fmt.Errorf("artifacts: catalog collection field %s is missing, read-only, or mismatched", fieldKey)
			}
			if !viewFieldFactsMatch(generatedField.View, field) {
				return nil, fmt.Errorf("artifacts: generated view facts for collection field %s are stale", fieldKey)
			}
			collection := collectionPolicy{
				Family:             generatedField.CollectionFamily,
				LinkType:           generatedField.LinkType,
				ExpectedTargetType: generatedField.ExpectedTargetRecordType,
				AllowedOperations:  slices.Clone(generatedField.AllowedOperations),
			}
			if !validCollectionPolicy(collection) {
				return nil, fmt.Errorf("artifacts: collection field %s has invalid generated policy", fieldKey)
			}
			catalogField := Field{
				ViewSchemaID: viewSchemaID,
				FieldKey:     fieldKey,
				Kind:         FieldKindCollection,
				View:         viewFacts(generatedField.View),
				Collection:   collection,
			}
			if _, duplicate := catalog.fieldsByKey[fieldKey]; duplicate {
				return nil, fmt.Errorf("artifacts: source field %s is owned by multiple surfaces", fieldKey)
			}
			ownedFields[fieldKey] = struct{}{}
			catalog.fieldsByKey[fieldKey] = catalogField
			collectionCount++
		}
		for fieldKey, field := range viewFields {
			if !field.Writable && !field.CreateWritable {
				continue
			}
			if _, present := ownedFields[fieldKey]; !present {
				return nil, fmt.Errorf("artifacts: writable field %s is missing from generated source catalog", fieldKey)
			}
		}
		catalog.surfaces = append(catalog.surfaces, surface)
		catalog.surfacesByViewID[viewSchemaID] = surface
		catalog.surfacesByArtifactType[registered.ArtifactType] = surface
	}
	if directCount != 36 || collectionCount != 15 {
		return nil, fmt.Errorf("artifacts: source catalog counts are %d direct and %d collection fields; want 36 and 15", directCount, collectionCount)
	}
	for _, viewSchemaID := range []string{
		contractartifacts.NoteViewSchemaID,
		contractartifacts.CommLogViewSchemaID,
		contractartifacts.HandoffViewSchemaID,
		contractartifacts.StatusReviewViewSchemaID,
		contractartifacts.LessonViewSchemaID,
		contractartifacts.FindingViewSchemaID,
		contractartifacts.InvestigativeQueryViewSchemaID,
		contractartifacts.ForensicKeywordViewSchemaID,
	} {
		surface, ok := catalog.surfacesByViewID[viewSchemaID]
		if !ok {
			return nil, fmt.Errorf("artifacts: projection surface order contains unknown view %s", viewSchemaID)
		}
		catalog.projectionSurfaces = append(catalog.projectionSurfaces, surface)
	}
	if len(catalog.projectionSurfaces) != len(catalog.surfaces) {
		return nil, fmt.Errorf("artifacts: projection surface order does not cover the source catalog")
	}
	return catalog, nil
}

func viewFacts(generated contractartifacts.ViewFieldFacts) ViewFieldFacts {
	return ViewFieldFacts{
		ReadKind:            generated.ReadKind,
		WriteKind:           generated.WriteKind,
		Writable:            generated.Writable,
		CreateWritable:      generated.CreateWritable,
		Clearable:           generated.Clearable,
		ReferenceContractID: generated.ReferenceContractID,
	}
}

func viewFieldFactsMatch(generated contractartifacts.ViewFieldFacts, field viewschema.Field) bool {
	reference := ""
	if field.DirectReferenceContractID != nil {
		reference = *field.DirectReferenceContractID
	}
	return generated.ReadKind == field.ReadKind && generated.WriteKind == field.WriteKind &&
		generated.Writable == field.Writable && generated.CreateWritable == field.CreateWritable &&
		generated.Clearable == field.Clearable && generated.ReferenceContractID == reference
}

func validCollectionPolicy(policy collectionPolicy) bool {
	switch policy.Family {
	case "party_ref":
		return policy.LinkType == "references_record" && policy.ExpectedTargetType == "party" &&
			slices.Equal(policy.AllowedOperations, []string{"add_party_ref", "remove_party_ref"})
	case "record_ref":
		return (policy.LinkType == "references_record" || policy.LinkType == "supported_by") &&
			slices.Equal(policy.AllowedOperations, []string{"add_record_ref", "remove_record_ref"})
	case "record_tag":
		return policy.LinkType == "" && policy.ExpectedTargetType == "" &&
			slices.Equal(policy.AllowedOperations, []string{"add_tag", "remove_tag"})
	case "risk_ref":
		return policy.LinkType == "" && policy.ExpectedTargetType == "" &&
			slices.Equal(policy.AllowedOperations, []string{"add_risk_ref", "remove_risk_ref"})
	default:
		return false
	}
}

func validStorageMapping(table, column string) bool {
	switch table {
	case "artifacts", "artifact_findings", "artifact_forensic_keywords", "artifact_investigative_queries":
		return validSQLIdentifier(column)
	default:
		return false
	}
}

func validSQLIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for index, char := range value {
		if char == '_' || char >= 'a' && char <= 'z' || index > 0 && char >= '0' && char <= '9' {
			continue
		}
		return false
	}
	return true
}

func cloneField(field Field) Field {
	field.Collection.AllowedOperations = slices.Clone(field.Collection.AllowedOperations)
	return field
}
