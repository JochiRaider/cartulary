package artifacts

import (
	"fmt"
	"slices"
	"sync"

	"github.com/JochiRaider/cartulary/internal/gen/contractartifacts"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type sourceFieldKind string

const (
	sourceFieldDirect     sourceFieldKind = "direct"
	sourceFieldCollection sourceFieldKind = "collection"
)

type sourceStorageMapping struct {
	table  string
	column string
}

type sourceFieldPolicy struct {
	viewSchemaID string
	fieldKey     string
	kind         sourceFieldKind
	readKind     string
	writable     bool
	clearable    bool
	reference    bool
	storage      sourceStorageMapping
	collection   CollectionPolicy
}

type sourceSurfacePolicy struct {
	viewSchemaID string
	artifactType string
	fields       map[string]sourceFieldPolicy
}

type sourcePolicyCatalog struct {
	surfaces map[string]sourceSurfacePolicy
	fields   map[string]sourceFieldPolicy
}

var (
	artifactSourcePoliciesOnce sync.Once
	artifactSourcePolicies     sourcePolicyCatalog
)

func artifactSourcePolicyCatalog() sourcePolicyCatalog {
	artifactSourcePoliciesOnce.Do(func() {
		artifactSourcePolicies = mustBuildArtifactSourcePolicyCatalog()
	})
	return artifactSourcePolicies
}

func lookupArtifactSourceSurface(viewSchemaID string) (sourceSurfacePolicy, bool) {
	policy, ok := artifactSourcePolicyCatalog().surfaces[viewSchemaID]
	return policy, ok
}

func lookupArtifactSourceField(fieldKey string) (sourceFieldPolicy, bool) {
	policy, ok := artifactSourcePolicyCatalog().fields[fieldKey]
	return policy, ok
}

func mustBuildArtifactSourcePolicyCatalog() sourcePolicyCatalog {
	catalog, err := buildArtifactSourcePolicyCatalog(contractartifacts.SourceCatalog)
	if err != nil {
		panic(err)
	}
	return catalog
}

func buildArtifactSourcePolicyCatalog(generated []contractartifacts.SourceSurface) (sourcePolicyCatalog, error) {
	if len(generated) != 8 {
		return sourcePolicyCatalog{}, fmt.Errorf("artifacts: source catalog must contain exactly 8 surfaces")
	}
	catalog := sourcePolicyCatalog{
		surfaces: make(map[string]sourceSurfacePolicy, len(generated)),
		fields:   make(map[string]sourceFieldPolicy),
	}
	seenTypes := map[string]struct{}{}
	directCount := 0
	collectionCount := 0
	for _, registered := range generated {
		viewSchemaID := registered.ViewSchemaID
		if viewSchemaID == "" || registered.ArtifactType == "" {
			return sourcePolicyCatalog{}, fmt.Errorf("artifacts: source catalog contains an incomplete surface")
		}
		if _, duplicate := catalog.surfaces[viewSchemaID]; duplicate {
			return sourcePolicyCatalog{}, fmt.Errorf("artifacts: duplicate source surface %s", viewSchemaID)
		}
		if _, duplicate := seenTypes[registered.ArtifactType]; duplicate {
			return sourcePolicyCatalog{}, fmt.Errorf("artifacts: duplicate artifact type %s", registered.ArtifactType)
		}
		seenTypes[registered.ArtifactType] = struct{}{}
		schema, ok := viewschema.Lookup(viewSchemaID)
		if !ok {
			return sourcePolicyCatalog{}, fmt.Errorf("artifacts: source policy surface %s is missing its view schema", viewSchemaID)
		}
		filter, ok := schema.CanonicalSourceFilter()
		if schema.BaseProjection != "artifact_grid_projection" || !ok || filter.Kind != "artifact_type" || filter.Field != "artifact_type" || filter.Value != registered.ArtifactType {
			return sourcePolicyCatalog{}, fmt.Errorf("artifacts: source policy surface %s mismatches its view schema source filter", viewSchemaID)
		}
		surface := sourceSurfacePolicy{
			viewSchemaID: viewSchemaID,
			artifactType: registered.ArtifactType,
			fields:       make(map[string]sourceFieldPolicy),
		}
		viewFields := schema.Fields()
		for _, generatedField := range registered.DirectFields {
			fieldKey := generatedField.FieldKey
			field, ok := viewFields[fieldKey]
			if !ok || (!field.Writable && !field.CreateWritable) || field.WriteKind != "direct_value" {
				return sourcePolicyCatalog{}, fmt.Errorf("artifacts: catalog direct field %s is missing, read-only, or mismatched", fieldKey)
			}
			if !viewFieldFactsMatch(generatedField.View, field) {
				return sourcePolicyCatalog{}, fmt.Errorf("artifacts: generated view facts for direct field %s are stale", fieldKey)
			}
			if !validArtifactStorageMapping(generatedField.Table, generatedField.Column) {
				return sourcePolicyCatalog{}, fmt.Errorf("artifacts: direct field %s has an incomplete storage mapping", fieldKey)
			}
			policy := sourceFieldPolicy{
				viewSchemaID: viewSchemaID,
				fieldKey:     fieldKey,
				readKind:     field.ReadKind,
				writable:     field.Writable || field.CreateWritable,
				clearable:    field.Clearable,
				reference:    field.DirectReferenceContractID != nil,
				kind:         sourceFieldDirect,
				storage: sourceStorageMapping{
					table: generatedField.Table, column: generatedField.Column,
				},
			}
			if _, duplicate := catalog.fields[fieldKey]; duplicate {
				return sourcePolicyCatalog{}, fmt.Errorf("artifacts: source field %s is owned by multiple surfaces", fieldKey)
			}
			surface.fields[fieldKey] = policy
			catalog.fields[fieldKey] = policy
			directCount++
		}
		for _, generatedField := range registered.CollectionFields {
			fieldKey := generatedField.FieldKey
			field, ok := viewFields[fieldKey]
			if !ok || (!field.Writable && !field.CreateWritable) || field.WriteKind != "action_payload" {
				return sourcePolicyCatalog{}, fmt.Errorf("artifacts: catalog collection field %s is missing, read-only, or mismatched", fieldKey)
			}
			if !viewFieldFactsMatch(generatedField.View, field) {
				return sourcePolicyCatalog{}, fmt.Errorf("artifacts: generated view facts for collection field %s are stale", fieldKey)
			}
			collection := CollectionPolicy{
				FieldKey: generatedField.FieldKey, Family: CollectionFamily(generatedField.CollectionFamily),
				LinkType: generatedField.LinkType, ExpectedTargetType: generatedField.ExpectedTargetRecordType,
				AllowedOps: append([]string(nil), generatedField.AllowedOperations...),
			}
			if !validCollectionPolicy(collection) {
				return sourcePolicyCatalog{}, fmt.Errorf("artifacts: collection field %s has invalid generated policy", fieldKey)
			}
			policy := sourceFieldPolicy{
				viewSchemaID: viewSchemaID, fieldKey: fieldKey, kind: sourceFieldCollection,
				readKind: field.ReadKind, writable: field.Writable || field.CreateWritable,
				clearable: field.Clearable, reference: field.DirectReferenceContractID != nil,
				collection: collection,
			}
			if _, duplicate := catalog.fields[fieldKey]; duplicate {
				return sourcePolicyCatalog{}, fmt.Errorf("artifacts: source field %s is owned by multiple surfaces", fieldKey)
			}
			surface.fields[fieldKey] = policy
			catalog.fields[fieldKey] = policy
			collectionCount++
		}
		for fieldKey, field := range viewFields {
			if !field.Writable && !field.CreateWritable {
				continue
			}
			if _, present := surface.fields[fieldKey]; !present {
				return sourcePolicyCatalog{}, fmt.Errorf("artifacts: writable field %s is missing from generated source catalog", fieldKey)
			}
		}
		catalog.surfaces[viewSchemaID] = surface
	}
	if directCount != 36 || collectionCount != 15 {
		return sourcePolicyCatalog{}, fmt.Errorf("artifacts: source catalog counts are %d direct and %d collection fields; want 36 and 15", directCount, collectionCount)
	}
	return catalog, nil
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

func validCollectionPolicy(policy CollectionPolicy) bool {
	switch policy.Family {
	case CollectionFamilyPartyRef:
		return policy.LinkType == "references_record" && policy.ExpectedTargetType == "party" &&
			slices.Equal(policy.AllowedOps, []string{"add_party_ref", "remove_party_ref"})
	case CollectionFamilyRecordRef:
		return (policy.LinkType == "references_record" || policy.LinkType == "supported_by") &&
			slices.Equal(policy.AllowedOps, []string{"add_record_ref", "remove_record_ref"})
	case CollectionFamilyRecordTag:
		return policy.LinkType == "" && policy.ExpectedTargetType == "" &&
			slices.Equal(policy.AllowedOps, []string{"add_tag", "remove_tag"})
	case CollectionFamilyRiskRef:
		return policy.LinkType == "" && policy.ExpectedTargetType == "" &&
			slices.Equal(policy.AllowedOps, []string{"add_risk_ref", "remove_risk_ref"})
	default:
		return false
	}
}

func validArtifactStorageMapping(table, column string) bool {
	switch table {
	case "artifacts", "artifact_findings", "artifact_forensic_keywords", "artifact_investigative_queries":
		return validArtifactSQLIdentifier(column)
	default:
		return false
	}
}

func validArtifactSQLIdentifier(value string) bool {
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

func validateArtifactDirectValue(policy sourceFieldPolicy, value FieldValue) error {
	count := 0
	if value.Text != nil {
		count++
	}
	if value.Timestamp != nil {
		count++
	}
	if value.UUID != nil {
		count++
	}
	if value.Number != nil {
		count++
	}
	if value.Bool != nil {
		count++
	}
	if count == 0 {
		if policy.clearable {
			return nil
		}
		return &ValidationError{Field: policy.fieldKey, ReasonCode: "field_not_nullable"}
	}
	if count != 1 {
		return &ValidationError{Field: policy.fieldKey, ReasonCode: "invalid_value"}
	}
	validKind := false
	switch {
	case policy.reference:
		validKind = value.UUID != nil
	case policy.readKind == "timestamp":
		validKind = value.Timestamp != nil
	case policy.readKind == "number":
		validKind = value.Number != nil
	case policy.readKind == "boolean":
		validKind = value.Bool != nil
	case policy.readKind == "text" || policy.readKind == "enum":
		validKind = value.Text != nil
	}
	if !validKind {
		return &ValidationError{Field: policy.fieldKey, ReasonCode: "invalid_value"}
	}
	return nil
}
