package artifacts

import "slices"

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

func (p CollectionPolicy) AllowsRecordRefs() bool {
	return p.Family == CollectionFamilyRecordRef
}

func (p CollectionPolicy) AllowsPartyRefs() bool {
	return p.Family == CollectionFamilyPartyRef
}

func (p CollectionPolicy) AllowsTags() bool {
	return p.Family == CollectionFamilyRecordTag
}

func (p CollectionPolicy) AllowsRiskRefs() bool {
	return p.Family == CollectionFamilyRiskRef
}

func (p CollectionPolicy) AllowsLinksCollectionMutation() bool {
	return p.AllowsRecordRefs() || p.AllowsPartyRefs() || p.AllowsTags()
}

func (p CollectionPolicy) AllowsOp(op string) bool {
	return slices.Contains(p.AllowedOps, op)
}

func LookupCollectionPolicy(fieldKey string) (CollectionPolicy, bool) {
	policy, ok := artifactCollectionPolicies[fieldKey]
	if !ok {
		return CollectionPolicy{}, false
	}
	return cloneCollectionPolicy(policy), true
}

func cloneCollectionPolicy(policy CollectionPolicy) CollectionPolicy {
	policy.AllowedOps = slices.Clone(policy.AllowedOps)
	return policy
}

var artifactCollectionPolicies = map[string]CollectionPolicy{
	"comm_log.action_task_ids": {
		FieldKey:           "comm_log.action_task_ids",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "task_request",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"comm_log.attendee_party_ids": {
		FieldKey:           "comm_log.attendee_party_ids",
		Family:             CollectionFamilyPartyRef,
		LinkType:           "references_record",
		ExpectedTargetType: "party",
		AllowedOps:         []string{"add_party_ref", "remove_party_ref"},
	},
	"comm_log.audience_party_ids": {
		FieldKey:           "comm_log.audience_party_ids",
		Family:             CollectionFamilyPartyRef,
		LinkType:           "references_record",
		ExpectedTargetType: "party",
		AllowedOps:         []string{"add_party_ref", "remove_party_ref"},
	},
	"comm_log.decision_ids": {
		FieldKey:           "comm_log.decision_ids",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "decision",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"finding.contradictory_refs": {
		FieldKey:   "finding.contradictory_refs",
		Family:     CollectionFamilyRecordRef,
		LinkType:   "references_record",
		AllowedOps: []string{"add_record_ref", "remove_record_ref"},
	},
	"finding.supporting_refs": {
		FieldKey:   "finding.supporting_refs",
		Family:     CollectionFamilyRecordRef,
		LinkType:   "supported_by",
		AllowedOps: []string{"add_record_ref", "remove_record_ref"},
	},
	"handoff.open_decision_ids": {
		FieldKey:           "handoff.open_decision_ids",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "decision",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"handoff.open_risk_refs": {
		FieldKey:   HandoffOpenRiskRefsFieldKey,
		Family:     CollectionFamilyRiskRef,
		AllowedOps: []string{"add_risk_ref", "remove_risk_ref"},
	},
	"handoff.open_task_ids": {
		FieldKey:           "handoff.open_task_ids",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "task_request",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"lesson.evidence_refs": {
		FieldKey:           "lesson.evidence_refs",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "evidence",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"lesson.follow_up_task_ids": {
		FieldKey:           "lesson.follow_up_task_ids",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "task_request",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"note.tags": {
		FieldKey:   "note.tags",
		Family:     CollectionFamilyRecordTag,
		AllowedOps: []string{"add_tag", "remove_tag"},
	},
	"status_review.blocked_task_ids": {
		FieldKey:           "status_review.blocked_task_ids",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "task_request",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"status_review.open_decision_ids": {
		FieldKey:           "status_review.open_decision_ids",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "decision",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
	"status_review.pending_evidence_ids": {
		FieldKey:           "status_review.pending_evidence_ids",
		Family:             CollectionFamilyRecordRef,
		LinkType:           "references_record",
		ExpectedTargetType: "evidence",
		AllowedOps:         []string{"add_record_ref", "remove_record_ref"},
	},
}
