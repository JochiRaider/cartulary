package collectionpolicy

import (
	"slices"
	"sort"
)

type OwnerFamily string

const (
	OwnerArtifactsRiskRefs OwnerFamily = "artifacts_risk_refs"
	OwnerEntities          OwnerFamily = "entities"
	OwnerLinks             OwnerFamily = "links"
	OwnerTimelineMentions  OwnerFamily = "timeline_mentions"
)

type ItemFamily string

const (
	ItemFamilyAliasText     ItemFamily = "alias_text"
	ItemFamilyMentionOrigin ItemFamily = "mention_origin"
	ItemFamilyPartyRef      ItemFamily = "party_ref"
	ItemFamilyRecordRef     ItemFamily = "record_ref"
	ItemFamilyRecordTag     ItemFamily = "record_tag"
	ItemFamilyRiskRef       ItemFamily = "risk_ref"
)

type Policy struct {
	FieldKey           string
	Owner              OwnerFamily
	ItemFamily         ItemFamily
	Ordered            bool
	LinkType           string
	ExpectedTargetType string
	ChangedFieldKeys   []string
	AllowedOps         []string
}

func (p Policy) AllowsLinksCollectionMutation() bool {
	return p.Owner == OwnerLinks && (p.ItemFamily == ItemFamilyRecordRef || p.ItemFamily == ItemFamilyPartyRef || p.ItemFamily == ItemFamilyRecordTag)
}

func (p Policy) AllowsRecordRefs() bool {
	return p.Owner == OwnerLinks && p.ItemFamily == ItemFamilyRecordRef
}

func (p Policy) AllowsPartyRefs() bool {
	return p.Owner == OwnerLinks && p.ItemFamily == ItemFamilyPartyRef
}

func (p Policy) AllowsTags() bool {
	return p.Owner == OwnerLinks && p.ItemFamily == ItemFamilyRecordTag
}

func (p Policy) EffectiveChangedFieldKeys() []string {
	if len(p.ChangedFieldKeys) == 0 {
		return []string{p.FieldKey}
	}
	return slices.Clone(p.ChangedFieldKeys)
}

func Lookup(fieldKey string) (Policy, bool) {
	policy, ok := policies[fieldKey]
	if !ok {
		return Policy{}, false
	}
	return clonePolicy(policy), true
}

func MustLookup(fieldKey string) Policy {
	policy, ok := Lookup(fieldKey)
	if !ok {
		panic("missing collection policy for " + fieldKey)
	}
	return policy
}

func All() []Policy {
	result := make([]Policy, 0, len(policies))
	for _, policy := range policies {
		result = append(result, clonePolicy(policy))
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left].FieldKey < result[right].FieldKey
	})
	return result
}

func FieldKeys() []string {
	keys := make([]string, 0, len(policies))
	for key := range policies {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func clonePolicy(policy Policy) Policy {
	policy.ChangedFieldKeys = slices.Clone(policy.ChangedFieldKeys)
	policy.AllowedOps = slices.Clone(policy.AllowedOps)
	return policy
}

var policies = map[string]Policy{
	"comm_log.action_task_ids": {
		FieldKey:           "comm_log.action_task_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "task_request",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"comm_log.attendee_party_ids": {
		FieldKey:           "comm_log.attendee_party_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyPartyRef,
		LinkType:           "references_record",
		ExpectedTargetType: "party",
		AllowedOps:         []string{"add_party_ref", "remove_party_ref"},
	},
	"comm_log.audience_party_ids": {
		FieldKey:           "comm_log.audience_party_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyPartyRef,
		LinkType:           "references_record",
		ExpectedTargetType: "party",
		AllowedOps:         []string{"add_party_ref", "remove_party_ref"},
	},
	"comm_log.decision_ids": {
		FieldKey:           "comm_log.decision_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "decision",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"decision.affected_record_ids": {
		FieldKey:   "decision.affected_record_ids",
		Owner:      OwnerLinks,
		ItemFamily: ItemFamilyRecordRef,
		LinkType:   "references_record",
		AllowedOps: []string{"add_record_ref", "remove_record_ref"},
	},
	"decision.support_refs": {
		FieldKey:   "decision.support_refs",
		Owner:      OwnerLinks,
		ItemFamily: ItemFamilyRecordRef,
		LinkType:   "supported_by",
		AllowedOps: []string{"add_record_ref", "remove_record_ref"},
	},
	"finding.contradictory_refs": {
		FieldKey:   "finding.contradictory_refs",
		Owner:      OwnerLinks,
		ItemFamily: ItemFamilyRecordRef,
		LinkType:   "references_record",
		AllowedOps: []string{"add_record_ref", "remove_record_ref"},
	},
	"finding.supporting_refs": {
		FieldKey:   "finding.supporting_refs",
		Owner:      OwnerLinks,
		ItemFamily: ItemFamilyRecordRef,
		LinkType:   "supported_by",
		AllowedOps: []string{"add_record_ref", "remove_record_ref"},
	},
	"handoff.open_decision_ids": {
		FieldKey:           "handoff.open_decision_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "decision",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"handoff.open_risk_refs": {
		FieldKey:   "handoff.open_risk_refs",
		Owner:      OwnerArtifactsRiskRefs,
		ItemFamily: ItemFamilyRiskRef,
		AllowedOps: []string{"add_risk_ref", "remove_risk_ref"},
	},
	"handoff.open_task_ids": {
		FieldKey:           "handoff.open_task_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "task_request",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"host.aliases": {
		FieldKey:   "host.aliases",
		Owner:      OwnerEntities,
		ItemFamily: ItemFamilyAliasText,
		AllowedOps: []string{"add_alias", "remove_alias"},
	},
	"identity.aliases": {
		FieldKey:   "identity.aliases",
		Owner:      OwnerEntities,
		ItemFamily: ItemFamilyAliasText,
		AllowedOps: []string{"add_alias", "remove_alias"},
	},
	"lesson.evidence_refs": {
		FieldKey:           "lesson.evidence_refs",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "evidence",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"lesson.follow_up_task_ids": {
		FieldKey:           "lesson.follow_up_task_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "task_request",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"note.tags": {
		FieldKey:   "note.tags",
		Owner:      OwnerLinks,
		ItemFamily: ItemFamilyRecordTag,
		AllowedOps: []string{"add_tag", "remove_tag"},
	},
	"status_review.blocked_task_ids": {
		FieldKey:           "status_review.blocked_task_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "task_request",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"status_review.open_decision_ids": {
		FieldKey:           "status_review.open_decision_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "decision",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"status_review.pending_evidence_ids": {
		FieldKey:           "status_review.pending_evidence_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "evidence",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"task.linked_record_ids": {
		FieldKey:   "task.linked_record_ids",
		Owner:      OwnerLinks,
		ItemFamily: ItemFamilyRecordRef,
		LinkType:   "references_record",
		AllowedOps: []string{"add_record_ref", "remove_record_ref"},
	},
	"timeline.attached_evidence_ids": {
		FieldKey:           "timeline.attached_evidence_ids",
		Owner:              OwnerLinks,
		ItemFamily:         ItemFamilyRecordRef,
		LinkType:           "attached_evidence",
		ExpectedTargetType: "evidence",
		ChangedFieldKeys:   []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"},
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"timeline.host_refs": {
		FieldKey:           "timeline.host_refs",
		Owner:              OwnerTimelineMentions,
		ItemFamily:         ItemFamilyMentionOrigin,
		Ordered:            true,
		LinkType:           "observed_on_host",
		ExpectedTargetType: "host",
		AllowedOps:         []string{"add_token", "add_resolved_ref", "resolve_item", "dismiss_item", "revert_to_unresolved"},
	},
	"timeline.identity_refs": {
		FieldKey:           "timeline.identity_refs",
		Owner:              OwnerTimelineMentions,
		ItemFamily:         ItemFamilyMentionOrigin,
		Ordered:            true,
		LinkType:           "observed_as_identity",
		ExpectedTargetType: "identity",
		AllowedOps:         []string{"add_token", "add_resolved_ref", "resolve_item", "dismiss_item", "revert_to_unresolved"},
	},
	"timeline.tags": {
		FieldKey:         "timeline.tags",
		Owner:            OwnerLinks,
		ItemFamily:       ItemFamilyRecordTag,
		ChangedFieldKeys: []string{"timeline.tags"},
		AllowedOps:       []string{"add_tag", "remove_tag"},
	},
}
