package timeline

import (
	"slices"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

type CollectionFamily string

const (
	CollectionFamilyMentionOrigin CollectionFamily = "mention_origin"
	CollectionFamilyRecordRef     CollectionFamily = "record_ref"
	CollectionFamilyRecordTag     CollectionFamily = "record_tag"
)

type CollectionPolicy struct {
	FieldKey           string
	Family             CollectionFamily
	Ordered            bool
	LinkType           string
	ExpectedTargetType string
	ChangedFieldKeys   []string
	AllowedOps         []string
}

func (p CollectionPolicy) AllowsOp(op string) bool {
	return slices.Contains(p.AllowedOps, op)
}

func (p CollectionPolicy) AllowsLinksCollectionMutation() bool {
	return p.Family == CollectionFamilyRecordRef || p.Family == CollectionFamilyRecordTag
}

func (p CollectionPolicy) IsAttachedEvidence() bool {
	return p.Family == CollectionFamilyRecordRef && p.LinkType == "attached_evidence"
}

// NormalizeText returns the source value and comparison key for a textual action.
func (p CollectionPolicy) NormalizeText(raw string) (source, comparison string, ok bool) {
	switch p.Family {
	case CollectionFamilyMentionOrigin:
		token, valid := fieldnorm.NormalizeMentionToken(raw)
		return raw, token, valid
	case CollectionFamilyRecordTag:
		return fieldnorm.NormalizeTagLabel(raw)
	default:
		return "", "", false
	}
}

// mentionEntityType returns the eligible entity type for a mention-origin field.
func mentionEntityType(fieldKey string) string {
	policy := timelineCollectionPolicies[fieldKey]
	if policy.Family != CollectionFamilyMentionOrigin {
		return ""
	}
	return policy.ExpectedTargetType
}

func LookupCollectionPolicy(fieldKey string) (CollectionPolicy, bool) {
	policy, ok := timelineCollectionPolicies[fieldKey]
	if !ok {
		return CollectionPolicy{}, false
	}
	return cloneCollectionPolicy(policy), true
}

func cloneCollectionPolicy(policy CollectionPolicy) CollectionPolicy {
	policy.ChangedFieldKeys = slices.Clone(policy.ChangedFieldKeys)
	policy.AllowedOps = slices.Clone(policy.AllowedOps)
	return policy
}

var timelineCollectionPolicies = map[string]CollectionPolicy{
	"timeline.attached_evidence_ids": {
		FieldKey:           "timeline.attached_evidence_ids",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "attached_evidence",
		ExpectedTargetType: "evidence",
		ChangedFieldKeys:   []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"},
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"timeline.host_refs": {
		FieldKey:           "timeline.host_refs",
		Family:             CollectionFamilyMentionOrigin,
		Ordered:            true,
		LinkType:           "observed_on_host",
		ExpectedTargetType: "host",
		AllowedOps:         []string{"add_token", "add_resolved_ref", "resolve_item", "dismiss_item", "revert_to_unresolved"},
	},
	"timeline.identity_refs": {
		FieldKey:           "timeline.identity_refs",
		Family:             CollectionFamilyMentionOrigin,
		Ordered:            true,
		LinkType:           "observed_as_identity",
		ExpectedTargetType: "identity",
		AllowedOps:         []string{"add_token", "add_resolved_ref", "resolve_item", "dismiss_item", "revert_to_unresolved"},
	},
	"timeline.tags": {
		FieldKey:         "timeline.tags",
		Family:           CollectionFamilyRecordTag,
		ChangedFieldKeys: []string{"timeline.tags"},
		AllowedOps:       []string{"add_tag", "remove_tag"},
	},
}
