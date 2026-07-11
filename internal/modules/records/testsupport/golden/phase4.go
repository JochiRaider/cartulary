package golden

import (
	"time"

	"github.com/google/uuid"
)

const (
	Phase4TimelineViewSchemaID   = "cartulary.view.timeline.v2"
	Phase4HostsViewSchemaID      = "cartulary.view.hosts.v1"
	Phase4IdentitiesViewSchemaID = "cartulary.view.identities.v1"
	Phase4IndicatorsViewSchemaID = "cartulary.view.indicators.v1"

	Phase4FieldTimelineHostRefs     = "timeline.host_refs"
	Phase4FieldTimelineIdentityRefs = "timeline.identity_refs"
	Phase4FieldTimelineSummary      = "timeline.activity_synopsis_text"
	Phase4FieldTimelineSourceText   = "timeline.raw_activity_text"

	Phase4BindingMentionOrigin = "mention_origin"
	Phase4BindingEntityOrigin  = "entity_origin"

	Phase4MentionStatusUnresolved = "unresolved"
	Phase4MentionStatusResolved   = "resolved"
	Phase4MentionStatusDismissed  = "dismissed"

	Phase4MentionActionResolve = "resolve_item"
	Phase4MentionActionDismiss = "dismiss_item"
	Phase4MentionActionRevert  = "revert_to_unresolved"

	Phase4IdentifierExactMatchReuse = "exact_match_reuse"
	Phase4IdentifierSuggestionOnly  = "suggestion_only"
	Phase4IdentifierProvenanceOnly  = "provenance_only"

	Phase4RecordStateStub      = "stub"
	Phase4RecordStateCanonical = "canonical"
	Phase4RecordStateMerged    = "merged"

	Phase4LinkProvenanceManual    = "manual"
	Phase4LinkProvenanceAutoMatch = "auto_match"

	Phase4IndicatorTypeIPv4   = "ipv4_addr"
	Phase4IndicatorTypeDomain = "domain_name"
	Phase4IndicatorTypeURL    = "url"
	Phase4IndicatorTypeHash   = "sha256"

	Phase4IndicatorValueKindAtomic    = "atomic"
	Phase4IndicatorValueKindPattern   = "pattern"
	Phase4IndicatorValueKindReference = "reference"
)

var (
	Phase4BaseTime = time.Date(2026, time.April, 18, 14, 30, 0, 0, time.UTC)
	Phase4PastTime = time.Date(2026, time.April, 17, 9, 15, 0, 0, time.UTC)

	Phase4IncidentID = uuid.MustParse("40000000-0000-0000-0000-000000000001")

	Phase4ViewerUserID    = uuid.MustParse("40000000-0000-0000-0000-000000000101")
	Phase4EditorUserID    = uuid.MustParse("40000000-0000-0000-0000-000000000102")
	Phase4ReviewerUserID  = uuid.MustParse("40000000-0000-0000-0000-000000000103")
	Phase4AdminUserID     = uuid.MustParse("40000000-0000-0000-0000-000000000104")
	Phase4NonMemberUserID = uuid.MustParse("40000000-0000-0000-0000-000000000105")

	Phase4TimelineRecordID        = uuid.MustParse("40000000-0000-0000-0000-000000000201")
	Phase4TimelineSiblingRecordID = uuid.MustParse("40000000-0000-0000-0000-000000000202")
	Phase4TimelineMixedRecordID   = uuid.MustParse("40000000-0000-0000-0000-000000000203")

	Phase4CanonicalHostRecordID  = uuid.MustParse("40000000-0000-0000-0000-000000000301")
	Phase4StubHostRecordID       = uuid.MustParse("40000000-0000-0000-0000-000000000302")
	Phase4MergedHostRecordID     = uuid.MustParse("40000000-0000-0000-0000-000000000303")
	Phase4DuplicateHostRecordID  = uuid.MustParse("40000000-0000-0000-0000-000000000304")
	Phase4CanonicalIdentityID    = uuid.MustParse("40000000-0000-0000-0000-000000000401")
	Phase4StubIdentityID         = uuid.MustParse("40000000-0000-0000-0000-000000000402")
	Phase4MergedIdentityID       = uuid.MustParse("40000000-0000-0000-0000-000000000403")
	Phase4DuplicateIdentityID    = uuid.MustParse("40000000-0000-0000-0000-000000000404")
	Phase4CanonicalIndicatorID   = uuid.MustParse("40000000-0000-0000-0000-000000000501")
	Phase4IndicatorObservationID = uuid.MustParse("40000000-0000-0000-0000-000000000502")
	Phase4IndicatorIntervalID    = uuid.MustParse("40000000-0000-0000-0000-000000000503")

	Phase4HostMentionID          = uuid.MustParse("40000000-0000-0000-0000-000000000601")
	Phase4ResolvedHostMentionID  = uuid.MustParse("40000000-0000-0000-0000-000000000602")
	Phase4DismissedHostMentionID = uuid.MustParse("40000000-0000-0000-0000-000000000603")
	Phase4IdentityMentionID      = uuid.MustParse("40000000-0000-0000-0000-000000000604")
	Phase4ResolvedIdentityID     = uuid.MustParse("40000000-0000-0000-0000-000000000605")
	Phase4DismissedIdentityID    = uuid.MustParse("40000000-0000-0000-0000-000000000606")

	Phase4ManualLinkID      = uuid.MustParse("40000000-0000-0000-0000-000000000701")
	Phase4AutoMatchLinkID   = uuid.MustParse("40000000-0000-0000-0000-000000000702")
	Phase4DuplicateLinkID   = uuid.MustParse("40000000-0000-0000-0000-000000000703")
	Phase4TagIDSurvivor     = uuid.MustParse("40000000-0000-0000-0000-000000000801")
	Phase4TagIDLoser        = uuid.MustParse("40000000-0000-0000-0000-000000000802")
	Phase4AssessmentHostID  = uuid.MustParse("40000000-0000-0000-0000-000000000901")
	Phase4AssessmentIdentID = uuid.MustParse("40000000-0000-0000-0000-000000000902")
)

var (
	Phase4HostExactMatchPrecedence = []string{
		"aad_device_id",
		"fqdn",
		"hostname",
	}
	Phase4IdentityExactMatchPrecedence = []string{
		"aad_object_id",
		"sid",
		"upn",
		"email",
		"sam_account_name",
	}
	Phase4MentionLifecycle = []string{
		Phase4MentionStatusUnresolved,
		Phase4MentionStatusResolved,
		Phase4MentionStatusDismissed,
	}
	Phase4MentionActions = []string{
		Phase4MentionActionResolve,
		Phase4MentionActionDismiss,
		Phase4MentionActionRevert,
	}
	Phase4IdentifierClasses = []string{
		Phase4IdentifierExactMatchReuse,
		Phase4IdentifierSuggestionOnly,
		Phase4IdentifierProvenanceOnly,
	}
	Phase4AutoResolutionEligibleTokens = []string{
		"WS-023",
		"vpn   gateway",
	}
	Phase4AutoResolutionSuppressedTokens = []string{
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

type Phase4LinkExpectation struct {
	Provenance string
	Confidence *int
}

type Phase4IndicatorExample struct {
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
	Phase4ManualLinkExpectation = Phase4LinkExpectation{
		Provenance: Phase4LinkProvenanceManual,
		Confidence: nil,
	}
	Phase4AutoMatchLinkExpectation = Phase4LinkExpectation{
		Provenance: Phase4LinkProvenanceAutoMatch,
		Confidence: intPointer(100),
	}
	Phase4IndicatorExamples = []Phase4IndicatorExample{
		{
			IndicatorType:   Phase4IndicatorTypeIPv4,
			ValueKind:       Phase4IndicatorValueKindAtomic,
			DisplayValue:    "203.0.113.24",
			NormalizedValue: "203.0.113.24",
			DefangedValue:   "203[.]0[.]113[.]24",
		},
		{
			IndicatorType:   Phase4IndicatorTypeDomain,
			ValueKind:       Phase4IndicatorValueKindAtomic,
			DisplayValue:    "vpn-gateway.example.test",
			NormalizedValue: "vpn-gateway.example.test",
			DefangedValue:   "vpn-gateway[.]example[.]test",
		},
		{
			IndicatorType:   Phase4IndicatorTypeURL,
			ValueKind:       Phase4IndicatorValueKindAtomic,
			DisplayValue:    "https://portal.example.test/login",
			NormalizedValue: "https://portal.example.test/login",
			DefangedValue:   "hxxps://portal[.]example[.]test/login",
		},
		{
			IndicatorType:   Phase4IndicatorTypeHash,
			ValueKind:       Phase4IndicatorValueKindPattern,
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
