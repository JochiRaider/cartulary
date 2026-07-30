package artifacts

import (
	"fmt"
	"sync"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/sourcecontract"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/surfacecatalog"
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
	registeredSurfaces := surfacecatalog.All()
	storage := artifactDirectStorageMappings()
	catalog := sourcePolicyCatalog{
		surfaces: make(map[string]sourceSurfacePolicy, len(registeredSurfaces)),
		fields:   make(map[string]sourceFieldPolicy),
	}
	for _, registered := range registeredSurfaces {
		viewSchemaID := registered.ViewSchemaID
		schema, ok := viewschema.Lookup(viewSchemaID)
		if !ok {
			panic(fmt.Sprintf("artifacts: source policy surface %s is missing its view schema", viewSchemaID))
		}
		surface := sourceSurfacePolicy{
			viewSchemaID: viewSchemaID,
			artifactType: registered.ArtifactType,
			fields:       make(map[string]sourceFieldPolicy),
		}
		for fieldKey, field := range schema.Fields() {
			if !field.Writable && !field.CreateWritable {
				continue
			}
			policy := sourceFieldPolicy{
				viewSchemaID: viewSchemaID,
				fieldKey:     fieldKey,
				readKind:     field.ReadKind,
				writable:     field.Writable || field.CreateWritable,
				clearable:    field.Clearable,
				reference:    field.DirectReferenceContractID != nil,
			}
			switch field.WriteKind {
			case "direct_value":
				mapping, mapped := storage[fieldKey]
				if !mapped {
					panic(fmt.Sprintf("artifacts: writable direct field %s has no exact storage mapping", fieldKey))
				}
				policy.kind = sourceFieldDirect
				policy.storage = mapping
			case "action_payload":
				collection, mapped := artifactCollectionPolicies[fieldKey]
				if !mapped {
					panic(fmt.Sprintf("artifacts: writable collection field %s has no collection policy", fieldKey))
				}
				policy.kind = sourceFieldCollection
				policy.collection = cloneCollectionPolicy(collection)
			default:
				panic(fmt.Sprintf("artifacts: writable field %s has unsupported write kind %q", fieldKey, field.WriteKind))
			}
			if _, duplicate := catalog.fields[fieldKey]; duplicate {
				panic(fmt.Sprintf("artifacts: source field %s is owned by multiple surfaces", fieldKey))
			}
			surface.fields[fieldKey] = policy
			catalog.fields[fieldKey] = policy
		}
		catalog.surfaces[viewSchemaID] = surface
	}
	for fieldKey := range storage {
		policy, ok := catalog.fields[fieldKey]
		if !ok || policy.kind != sourceFieldDirect {
			panic(fmt.Sprintf("artifacts: storage mapping %s is not an authored writable direct field", fieldKey))
		}
	}
	for fieldKey := range artifactCollectionPolicies {
		policy, ok := catalog.fields[fieldKey]
		if !ok || policy.kind != sourceFieldCollection {
			panic(fmt.Sprintf("artifacts: collection policy %s is not an authored writable collection field", fieldKey))
		}
	}
	return catalog
}

func artifactDirectStorageMappings() map[string]sourceStorageMapping {
	owned := sourcecontract.WritableDirectStorageMappings()
	result := make(map[string]sourceStorageMapping, len(owned))
	for fieldKey, mapping := range owned {
		result[fieldKey] = sourceStorageMapping{
			table:  mapping.Table,
			column: mapping.Column,
		}
	}
	return result
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
