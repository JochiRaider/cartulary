package golden

import (
	"time"

	"github.com/google/uuid"
)

const (
	RecordTimelineViewSchemaID   = "cartulary.view.timeline.v2"
	RecordHostsViewSchemaID      = "cartulary.view.hosts.v1"
	RecordIdentitiesViewSchemaID = "cartulary.view.identities.v1"
	RecordIndicatorsViewSchemaID = "cartulary.view.indicators.v1"

	RecordFieldTimelineHostRefs     = "timeline.host_refs"
	RecordFieldTimelineIdentityRefs = "timeline.identity_refs"
	RecordFieldTimelineSummary      = "timeline.activity_synopsis_text"
	RecordFieldTimelineSourceText   = "timeline.raw_activity_text"

	RecordBindingMentionOrigin = "mention_origin"
	RecordBindingEntityOrigin  = "entity_origin"

	RecordMentionStatusUnresolved = "unresolved"
	RecordMentionStatusResolved   = "resolved"
	RecordMentionStatusDismissed  = "dismissed"

	RecordMentionActionResolve = "resolve_item"
	RecordMentionActionDismiss = "dismiss_item"
	RecordMentionActionRevert  = "revert_to_unresolved"

	RecordIdentifierExactMatchReuse = "exact_match_reuse"
	RecordIdentifierSuggestionOnly  = "suggestion_only"
	RecordIdentifierProvenanceOnly  = "provenance_only"

	RecordRecordStateStub      = "stub"
	RecordRecordStateCanonical = "canonical"
	RecordRecordStateMerged    = "merged"

	RecordLinkProvenanceManual    = "manual"
	RecordLinkProvenanceAutoMatch = "auto_match"

	RecordIndicatorTypeIPv4   = "ipv4_addr"
	RecordIndicatorTypeDomain = "domain_name"
	RecordIndicatorTypeURL    = "url"
	RecordIndicatorTypeHash   = "sha256"

	RecordIndicatorValueKindAtomic    = "atomic"
	RecordIndicatorValueKindPattern   = "pattern"
	RecordIndicatorValueKindReference = "reference"
)

var (
	RecordBaseTime = time.Date(2026, time.April, 18, 14, 30, 0, 0, time.UTC)
	RecordPastTime = time.Date(2026, time.April, 17, 9, 15, 0, 0, time.UTC)

	RecordIncidentID = uuid.MustParse("40000000-0000-0000-0000-000000000001")

	RecordViewerUserID    = uuid.MustParse("40000000-0000-0000-0000-000000000101")
	RecordEditorUserID    = uuid.MustParse("40000000-0000-0000-0000-000000000102")
	RecordReviewerUserID  = uuid.MustParse("40000000-0000-0000-0000-000000000103")
	RecordAdminUserID     = uuid.MustParse("40000000-0000-0000-0000-000000000104")
	RecordNonMemberUserID = uuid.MustParse("40000000-0000-0000-0000-000000000105")

	RecordTimelineRecordID        = uuid.MustParse("40000000-0000-0000-0000-000000000201")
	RecordTimelineSiblingRecordID = uuid.MustParse("40000000-0000-0000-0000-000000000202")
	RecordTimelineMixedRecordID   = uuid.MustParse("40000000-0000-0000-0000-000000000203")

	RecordCanonicalHostRecordID  = uuid.MustParse("40000000-0000-0000-0000-000000000301")
	RecordStubHostRecordID       = uuid.MustParse("40000000-0000-0000-0000-000000000302")
	RecordMergedHostRecordID     = uuid.MustParse("40000000-0000-0000-0000-000000000303")
	RecordDuplicateHostRecordID  = uuid.MustParse("40000000-0000-0000-0000-000000000304")
	RecordCanonicalIdentityID    = uuid.MustParse("40000000-0000-0000-0000-000000000401")
	RecordStubIdentityID         = uuid.MustParse("40000000-0000-0000-0000-000000000402")
	RecordMergedIdentityID       = uuid.MustParse("40000000-0000-0000-0000-000000000403")
	RecordDuplicateIdentityID    = uuid.MustParse("40000000-0000-0000-0000-000000000404")
	RecordCanonicalIndicatorID   = uuid.MustParse("40000000-0000-0000-0000-000000000501")
	RecordIndicatorObservationID = uuid.MustParse("40000000-0000-0000-0000-000000000502")
	RecordIndicatorIntervalID    = uuid.MustParse("40000000-0000-0000-0000-000000000503")

	RecordHostMentionID          = uuid.MustParse("40000000-0000-0000-0000-000000000601")
	RecordResolvedHostMentionID  = uuid.MustParse("40000000-0000-0000-0000-000000000602")
	RecordDismissedHostMentionID = uuid.MustParse("40000000-0000-0000-0000-000000000603")
	RecordIdentityMentionID      = uuid.MustParse("40000000-0000-0000-0000-000000000604")
	RecordResolvedIdentityID     = uuid.MustParse("40000000-0000-0000-0000-000000000605")
	RecordDismissedIdentityID    = uuid.MustParse("40000000-0000-0000-0000-000000000606")

	RecordManualLinkID      = uuid.MustParse("40000000-0000-0000-0000-000000000701")
	RecordAutoMatchLinkID   = uuid.MustParse("40000000-0000-0000-0000-000000000702")
	RecordDuplicateLinkID   = uuid.MustParse("40000000-0000-0000-0000-000000000703")
	RecordTagIDSurvivor     = uuid.MustParse("40000000-0000-0000-0000-000000000801")
	RecordTagIDLoser        = uuid.MustParse("40000000-0000-0000-0000-000000000802")
	RecordAssessmentHostID  = uuid.MustParse("40000000-0000-0000-0000-000000000901")
	RecordAssessmentIdentID = uuid.MustParse("40000000-0000-0000-0000-000000000902")
)

var (
	RecordHostExactMatchPrecedence = []string{
		"aad_device_id",
		"fqdn",
		"hostname",
	}
	RecordIdentityExactMatchPrecedence = []string{
		"aad_object_id",
		"sid",
		"upn",
		"email",
		"sam_account_name",
	}
	RecordMentionLifecycle = []string{
		RecordMentionStatusUnresolved,
		RecordMentionStatusResolved,
		RecordMentionStatusDismissed,
	}
	RecordMentionActions = []string{
		RecordMentionActionResolve,
		RecordMentionActionDismiss,
		RecordMentionActionRevert,
	}
	RecordIdentifierClasses = []string{
		RecordIdentifierExactMatchReuse,
		RecordIdentifierSuggestionOnly,
		RecordIdentifierProvenanceOnly,
	}
	RecordAutoResolutionEligibleTokens = []string{
		"WS-023",
		"vpn   gateway",
	}
	RecordAutoResolutionSuppressedTokens = []string{
		"WS-023?",
		"WS-023??",
		"WS-023 ~",
		"WS-023 maybe",
		"WS-023 prob",
		"WS-023 probably",
		"WS-023 approx",
		"WS-023 approximately",
		"(WS-023)",
		"WS-023.",
		"WS-023,",
		"WS-023 likely",
	}
)

type RecordLinkExpectation struct {
	Provenance string
	Confidence *int
}

type RecordIndicatorExample struct {
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue string
	DefangedValue   string
	STIXPattern     string
	HashAlgorithm   string
	HashValue       string
}

var (
	RecordManualLinkExpectation = RecordLinkExpectation{
		Provenance: RecordLinkProvenanceManual,
		Confidence: nil,
	}
	RecordAutoMatchLinkExpectation = RecordLinkExpectation{
		Provenance: RecordLinkProvenanceAutoMatch,
		Confidence: intPointer(100),
	}
	RecordIndicatorExamples = []RecordIndicatorExample{
		{
			IndicatorType:   RecordIndicatorTypeIPv4,
			ValueKind:       RecordIndicatorValueKindAtomic,
			DisplayValue:    "203.0.113.24",
			NormalizedValue: "203.0.113.24",
			DefangedValue:   "203[.]0[.]113[.]24",
		},
		{
			IndicatorType:   RecordIndicatorTypeDomain,
			ValueKind:       RecordIndicatorValueKindAtomic,
			DisplayValue:    "vpn-gateway.example.test",
			NormalizedValue: "vpn-gateway.example.test",
			DefangedValue:   "vpn-gateway[.]example[.]test",
		},
		{
			IndicatorType:   RecordIndicatorTypeURL,
			ValueKind:       RecordIndicatorValueKindAtomic,
			DisplayValue:    "https://portal.example.test/login",
			NormalizedValue: "https://portal.example.test/login",
			DefangedValue:   "hxxps://portal[.]example[.]test/login",
		},
		{
			IndicatorType:   RecordIndicatorTypeHash,
			ValueKind:       RecordIndicatorValueKindPattern,
			DisplayValue:    "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			NormalizedValue: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			HashAlgorithm:   "sha256",
			HashValue:       "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			STIXPattern:     "[file:hashes.'SHA-256' = '2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824']",
		},
	}
)

func intPointer(value int) *int {
	return &value
}
