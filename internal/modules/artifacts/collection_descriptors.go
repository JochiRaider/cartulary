package artifacts

import (
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
)

type collectionFamily string

const (
	collectionFamilyPartyRef  collectionFamily = "party_ref"
	collectionFamilyRecordRef collectionFamily = "record_ref"
	collectionFamilyRecordTag collectionFamily = "record_tag"
	collectionFamilyRiskRef   collectionFamily = "risk_ref"
)

type collectionPolicy struct {
	FieldKey           string
	Family             collectionFamily
	LinkType           string
	ExpectedTargetType string
	AllowedOps         []string
}

func (p collectionPolicy) allowsRecordRefs() bool { return p.Family == collectionFamilyRecordRef }
func (p collectionPolicy) allowsPartyRefs() bool  { return p.Family == collectionFamilyPartyRef }
func (p collectionPolicy) allowsTags() bool       { return p.Family == collectionFamilyRecordTag }
func (p collectionPolicy) allowsRiskRefs() bool   { return p.Family == collectionFamilyRiskRef }

func (p collectionPolicy) allowsOp(op string) bool { return slices.Contains(p.AllowedOps, op) }

func lookupCollectionPolicy(fieldKey string) (collectionPolicy, bool) {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return collectionPolicy{}, false
	}
	field, ok := catalog.Field(fieldKey)
	if !ok || field.Kind != sourcecatalog.FieldKindCollection {
		return collectionPolicy{}, false
	}
	return collectionPolicyFromCatalogField(field), true
}

func collectionPolicyFromCatalogField(field sourcecatalog.Field) collectionPolicy {
	return collectionPolicy{
		FieldKey: field.FieldKey, Family: collectionFamily(field.Collection.Family),
		LinkType: field.Collection.LinkType, ExpectedTargetType: field.Collection.ExpectedTargetType,
		AllowedOps: slices.Clone(field.Collection.AllowedOperations),
	}
}
