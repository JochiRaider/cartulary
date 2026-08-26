package sourcecatalog

import (
	"fmt"
	"slices"
	"sync"

	"github.com/JochiRaider/cartulary/internal/gen/contracttasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type FieldKind string

const (
	FieldKindDirect     FieldKind = "direct"
	FieldKindCollection FieldKind = "collection"
)

type Surface struct {
	ViewSchemaID             string
	RecordType               string
	BaseProjection           string
	SourceTable              string
	RevisionSnapshotSchemaID string
}

type ViewFieldFacts struct {
	ReadKind            string
	WriteKind           string
	Writable            bool
	CreateWritable      bool
	Clearable           bool
	ReferenceContractID string
	EnumValues          []string
}

type StorageMapping struct {
	Table  string
	Column string
}

type ReferencePolicy struct {
	Role                     string
	ExpectedTargetRecordType string
	MirrorLinkType           string
}

type CollectionPolicy struct {
	Family                   string
	AllowedOperations        []string
	LinkType                 string
	ExpectedTargetRecordType string
}

type Field struct {
	ViewSchemaID string
	FieldKey     string
	Kind         FieldKind
	View         ViewFieldFacts
	Storage      StorageMapping
	Reference    ReferencePolicy
	Collection   CollectionPolicy
}

type Catalog struct {
	surfaces         []Surface
	surfacesByViewID map[string]Surface
	surfacesByRecord map[string]Surface
	fieldsByKey      map[string]Field
}

var (
	loadOnce    sync.Once
	loadCatalog *Catalog
	loadErr     error
)

// Load validates the generated Tasks/Decisions owner facts once and caches
// both the immutable catalog and any construction error.
func Load() (*Catalog, error) {
	loadOnce.Do(func() {
		loadCatalog, loadErr = build(contracttasksdecisions.SourceCatalog)
	})
	return loadCatalog, loadErr
}

func (c *Catalog) Surfaces() []Surface {
	if c == nil {
		return nil
	}
	return slices.Clone(c.surfaces)
}

func (c *Catalog) SurfaceByViewID(viewSchemaID string) (Surface, bool) {
	if c == nil {
		return Surface{}, false
	}
	surface, ok := c.surfacesByViewID[viewSchemaID]
	return surface, ok
}

func (c *Catalog) SurfaceByRecordType(recordType string) (Surface, bool) {
	if c == nil {
		return Surface{}, false
	}
	surface, ok := c.surfacesByRecord[recordType]
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
	result := make(map[string]StorageMapping, 20)
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

func (c *Catalog) ConflictFieldSourceKeys(viewSchemaID string) map[string]string {
	result := make(map[string]string)
	if c == nil {
		return result
	}
	for fieldKey, field := range c.fieldsByKey {
		if field.ViewSchemaID == viewSchemaID && field.Kind == FieldKindDirect {
			result[fieldKey] = field.Storage.Column
		}
	}
	return result
}

func build(generated []contracttasksdecisions.SourceSurface) (*Catalog, error) {
	if len(generated) != 2 {
		return nil, fmt.Errorf("tasksdecisions: source catalog must contain exactly 2 surfaces")
	}
	catalog := &Catalog{
		surfaces:         make([]Surface, 0, len(generated)),
		surfacesByViewID: make(map[string]Surface, len(generated)),
		surfacesByRecord: make(map[string]Surface, len(generated)),
		fieldsByKey:      make(map[string]Field, 23),
	}
	directCount := 0
	collectionCount := 0
	for _, registered := range generated {
		if err := validateSurfaceFacts(registered); err != nil {
			return nil, err
		}
		if _, duplicate := catalog.surfacesByViewID[registered.ViewSchemaID]; duplicate {
			return nil, fmt.Errorf("tasksdecisions: duplicate source surface %s", registered.ViewSchemaID)
		}
		if _, duplicate := catalog.surfacesByRecord[registered.RecordType]; duplicate {
			return nil, fmt.Errorf("tasksdecisions: duplicate record type %s", registered.RecordType)
		}
		schema, ok := viewschema.Lookup(registered.ViewSchemaID)
		if !ok || schema.BaseProjection != registered.BaseProjection {
			return nil, fmt.Errorf("tasksdecisions: source surface %s mismatches its view schema", registered.ViewSchemaID)
		}
		surface := Surface{
			ViewSchemaID: registered.ViewSchemaID, RecordType: registered.RecordType,
			BaseProjection: registered.BaseProjection, SourceTable: registered.SourceTable,
			RevisionSnapshotSchemaID: registered.RevisionSnapshotSchemaID,
		}
		ownedFields := make(map[string]struct{}, len(registered.DirectFields)+len(registered.CollectionFields))
		viewFields := schema.Fields()
		for _, generatedField := range registered.DirectFields {
			field, ok := viewFields[generatedField.FieldKey]
			if !ok || !field.Writable || field.WriteKind != "direct_value" {
				return nil, fmt.Errorf("tasksdecisions: direct field %s is missing, read-only, or mismatched", generatedField.FieldKey)
			}
			if !viewFieldFactsMatch(generatedField.View, field) {
				return nil, fmt.Errorf("tasksdecisions: generated view facts for direct field %s are stale", generatedField.FieldKey)
			}
			if !validStorageMapping(registered.SourceTable, generatedField.Column) {
				return nil, fmt.Errorf("tasksdecisions: direct field %s has an unsafe storage mapping", generatedField.FieldKey)
			}
			reference := ReferencePolicy{
				Role: generatedField.ReferenceRole, ExpectedTargetRecordType: generatedField.ExpectedTargetRecordType,
				MirrorLinkType: generatedField.MirrorLinkType,
			}
			if !validReferencePolicy(generatedField.FieldKey, generatedField.View.DirectReferenceContractID, reference) {
				return nil, fmt.Errorf("tasksdecisions: direct field %s has invalid reference policy", generatedField.FieldKey)
			}
			catalogField := Field{
				ViewSchemaID: registered.ViewSchemaID, FieldKey: generatedField.FieldKey, Kind: FieldKindDirect,
				View: viewFacts(generatedField.View), Storage: StorageMapping{Table: registered.SourceTable, Column: generatedField.Column},
				Reference: reference,
			}
			if err := catalog.registerField(catalogField, ownedFields); err != nil {
				return nil, err
			}
			directCount++
		}
		for _, generatedField := range registered.CollectionFields {
			field, ok := viewFields[generatedField.FieldKey]
			if !ok || !field.Writable || field.WriteKind != "action_payload" {
				return nil, fmt.Errorf("tasksdecisions: collection field %s is missing, read-only, or mismatched", generatedField.FieldKey)
			}
			if !viewFieldFactsMatch(generatedField.View, field) {
				return nil, fmt.Errorf("tasksdecisions: generated view facts for collection field %s are stale", generatedField.FieldKey)
			}
			collection := CollectionPolicy{
				Family: generatedField.CollectionFamily, AllowedOperations: slices.Clone(generatedField.AllowedOperations),
				LinkType: generatedField.LinkType, ExpectedTargetRecordType: generatedField.ExpectedTargetRecordType,
			}
			if !validCollectionPolicy(collection) {
				return nil, fmt.Errorf("tasksdecisions: collection field %s has invalid generated policy", generatedField.FieldKey)
			}
			catalogField := Field{
				ViewSchemaID: registered.ViewSchemaID, FieldKey: generatedField.FieldKey, Kind: FieldKindCollection,
				View: viewFacts(generatedField.View), Collection: collection,
			}
			if err := catalog.registerField(catalogField, ownedFields); err != nil {
				return nil, err
			}
			collectionCount++
		}
		for fieldKey, field := range viewFields {
			if !field.Writable {
				continue
			}
			if _, present := ownedFields[fieldKey]; !present {
				return nil, fmt.Errorf("tasksdecisions: writable field %s is missing from the source catalog", fieldKey)
			}
		}
		catalog.surfaces = append(catalog.surfaces, surface)
		catalog.surfacesByViewID[surface.ViewSchemaID] = surface
		catalog.surfacesByRecord[surface.RecordType] = surface
	}
	if directCount != 20 || collectionCount != 3 {
		return nil, fmt.Errorf("tasksdecisions: source catalog counts are %d direct and %d collection fields; want 20 and 3", directCount, collectionCount)
	}
	return catalog, nil
}

func (c *Catalog) registerField(field Field, ownedFields map[string]struct{}) error {
	if _, duplicate := c.fieldsByKey[field.FieldKey]; duplicate {
		return fmt.Errorf("tasksdecisions: source field %s is owned by multiple surfaces", field.FieldKey)
	}
	ownedFields[field.FieldKey] = struct{}{}
	c.fieldsByKey[field.FieldKey] = field
	return nil
}

func validateSurfaceFacts(surface contracttasksdecisions.SourceSurface) error {
	type expected struct {
		recordType, baseProjection, table, snapshot string
	}
	wantByView := map[string]expected{
		contracttasksdecisions.DecisionViewSchemaID: {
			recordType: "decision", baseProjection: "decision_grid_projection", table: "decisions",
			snapshot: "cartulary.revisions.snapshot.decision.v1",
		},
		contracttasksdecisions.TaskRequestViewSchemaID: {
			recordType: "task_request", baseProjection: "task_request_grid_projection", table: "task_requests",
			snapshot: "cartulary.revisions.snapshot.task_request.v1",
		},
	}
	want, ok := wantByView[surface.ViewSchemaID]
	if !ok || surface.RecordType != want.recordType || surface.BaseProjection != want.baseProjection ||
		surface.SourceTable != want.table || surface.RevisionSnapshotSchemaID != want.snapshot {
		return fmt.Errorf("tasksdecisions: source surface %s has invalid routing facts", surface.ViewSchemaID)
	}
	return nil
}

func viewFacts(generated contracttasksdecisions.ViewFieldFacts) ViewFieldFacts {
	return ViewFieldFacts{
		ReadKind: generated.ReadKind, WriteKind: generated.WriteKind, Writable: generated.Writable,
		CreateWritable: generated.CreateWritable, Clearable: generated.Clearable,
		ReferenceContractID: generated.DirectReferenceContractID, EnumValues: slices.Clone(generated.EnumValues),
	}
}

func viewFieldFactsMatch(generated contracttasksdecisions.ViewFieldFacts, field viewschema.Field) bool {
	reference := ""
	if field.DirectReferenceContractID != nil {
		reference = *field.DirectReferenceContractID
	}
	return generated.ReadKind == field.ReadKind && generated.WriteKind == field.WriteKind &&
		generated.Writable == field.Writable && generated.CreateWritable == field.CreateWritable &&
		generated.Clearable == field.Clearable && generated.DirectReferenceContractID == reference &&
		slices.Equal(generated.EnumValues, field.EnumValues)
}

func validStorageMapping(table, column string) bool {
	if table != "task_requests" && table != "decisions" {
		return false
	}
	return validSQLIdentifier(column)
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

func validReferencePolicy(fieldKey, contractID string, policy ReferencePolicy) bool {
	switch fieldKey {
	case "decision.owner_user_id", "task.owner_user_id":
		return contractID == "incident_member_user_ref_v1" && policy.Role == "incident_member_user" &&
			policy.ExpectedTargetRecordType == "" && policy.MirrorLinkType == ""
	case "task.requester_party_id":
		return contractID == "same_incident_party_ref_v1" && policy.Role == "same_incident_record" &&
			policy.ExpectedTargetRecordType == "party" && policy.MirrorLinkType == ""
	case "task.decision_record_id":
		return contractID == "same_incident_decision_ref_v1" && policy.Role == "same_incident_record" &&
			policy.ExpectedTargetRecordType == "decision" && policy.MirrorLinkType == "references_record"
	default:
		return contractID == "" && policy == (ReferencePolicy{})
	}
}

func validCollectionPolicy(policy CollectionPolicy) bool {
	return policy.Family == "record_ref" && policy.ExpectedTargetRecordType == "" &&
		(policy.LinkType == "references_record" || policy.LinkType == "supported_by") &&
		slices.Equal(policy.AllowedOperations, []string{"add_record_ref", "remove_record_ref"})
}

func cloneField(field Field) Field {
	field.View.EnumValues = slices.Clone(field.View.EnumValues)
	field.Collection.AllowedOperations = slices.Clone(field.Collection.AllowedOperations)
	return field
}
