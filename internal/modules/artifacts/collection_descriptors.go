package artifacts

import (
	"slices"

	"github.com/JochiRaider/cartulary/internal/gen/contractartifacts"
)

type CollectionFamily string

const (
	CollectionFamilyPartyRef  CollectionFamily = "party_ref"
	CollectionFamilyRecordRef CollectionFamily = "record_ref"
	CollectionFamilyRecordTag CollectionFamily = "record_tag"
	CollectionFamilyRiskRef   CollectionFamily = "risk_ref"
)

type CollectionPolicy struct {
	FieldKey           string
	Family             CollectionFamily
	LinkType           string
	ExpectedTargetType string
	AllowedOps         []string
}

func (p CollectionPolicy) AllowsRecordRefs() bool { return p.Family == CollectionFamilyRecordRef }
func (p CollectionPolicy) AllowsPartyRefs() bool  { return p.Family == CollectionFamilyPartyRef }
func (p CollectionPolicy) AllowsTags() bool       { return p.Family == CollectionFamilyRecordTag }
func (p CollectionPolicy) AllowsRiskRefs() bool   { return p.Family == CollectionFamilyRiskRef }

func (p CollectionPolicy) AllowsLinksCollectionMutation() bool {
	return p.AllowsRecordRefs() || p.AllowsPartyRefs() || p.AllowsTags()
}

func (p CollectionPolicy) AllowsOp(op string) bool { return slices.Contains(p.AllowedOps, op) }

func LookupCollectionPolicy(fieldKey string) (CollectionPolicy, bool) {
	for _, surface := range contractartifacts.SourceCatalog {
		for _, field := range surface.CollectionFields {
			if field.FieldKey != fieldKey {
				continue
			}
			return CollectionPolicy{
				FieldKey: field.FieldKey, Family: CollectionFamily(field.CollectionFamily),
				LinkType: field.LinkType, ExpectedTargetType: field.ExpectedTargetRecordType,
				AllowedOps: slices.Clone(field.AllowedOperations),
			}, true
		}
	}
	return CollectionPolicy{}, false
}
