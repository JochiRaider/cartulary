package fixtures

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/golden"
)

type IncidentFixture struct {
	IncidentID  uuid.UUID
	IncidentKey string
	Title       string
}

type UserFixture struct {
	UserID      uuid.UUID
	Role        string
	Email       string
	DisplayName string
}

type MembershipFixture struct {
	IncidentID uuid.UUID
	UserID     uuid.UUID
	Role       string
}

type ViewFixture struct {
	ViewSchemaID string
	Title        string
	FieldBinding map[string]string
}

type TimelineRowFixture struct {
	RecordID     uuid.UUID
	RowVersion   int64
	Summary      string
	SourceText   string
	HostTokens   []string
	IdentityText []string
}

type PreservedIdentifierFixture struct {
	IdentifierClass string
	IdentifierType  string
	Value           string
}

type HostFixture struct {
	RecordID              uuid.UUID
	State                 string
	DisplayName           string
	AADDeviceID           string
	FQDN                  string
	Hostname              string
	MergedIntoRecordID    *uuid.UUID
	PreservedIdentifiers  []PreservedIdentifierFixture
	SuggestionOnlyAliases []string
}

type IdentityFixture struct {
	RecordID              uuid.UUID
	State                 string
	DisplayName           string
	AADObjectID           string
	SID                   string
	UPN                   string
	Email                 string
	SamAccountName        string
	MergedIntoRecordID    *uuid.UUID
	PreservedIdentifiers  []PreservedIdentifierFixture
	SuggestionOnlyAliases []string
}

type EntityMentionFixture struct {
	EntityMentionID  uuid.UUID
	SourceRecordID   uuid.UUID
	EntityType       string
	SourceFieldKey   string
	OriginKind       string
	OriginLocator    string
	RawText          string
	NormalizedText   string
	ResolutionStatus string
	RowVersion       int64
	ResolvedRecordID *uuid.UUID
	ResolvedByUserID *uuid.UUID
	ResolvedAt       *time.Time
	ResolutionMethod *string
	Ordinal          int
}

type LinkFixture struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SourceID     uuid.UUID
	TargetID     uuid.UUID
	LinkType     string
	Provenance   string
	Confidence   *int
	DeletedAt    *time.Time
}

type TagFixture struct {
	TagID     uuid.UUID
	RecordID  uuid.UUID
	TagName   string
	DeletedAt *time.Time
}

type AssessmentFixture struct {
	AssessmentID uuid.UUID
	SubjectID    uuid.UUID
	SubjectType  string
	State        string
	AssessedAt   time.Time
}

type IndicatorFixture struct {
	RecordID         uuid.UUID
	IndicatorType    string
	ValueKind        string
	DisplayValue     string
	NormalizedValue  string
	DefangedValue    string
	STIXPattern      string
	HashAlgorithm    string
	HashValue        string
	ObservationCount int
	FirstObservedAt  time.Time
	LastObservedAt   time.Time
}

type IndicatorObservationFixture struct {
	ObservationID             uuid.UUID
	SourceRecordID            uuid.UUID
	SourceFieldKey            string
	OriginKind                string
	OriginLocator             string
	ObservedText              string
	ParsedIndicatorType       string
	NormalizedCandidate       string
	ResolutionStatus          string
	ResolvedIndicatorRecordID *uuid.UUID
}

type IndicatorLifecycleFixture struct {
	IntervalID     uuid.UUID
	IndicatorID    uuid.UUID
	LifecycleState string
	StartedAt      time.Time
	EndedAt        *time.Time
}

func RecordIncident() IncidentFixture {
	return IncidentFixture{
		IncidentID:  golden.RecordIncidentID,
		IncidentKey: "IR-ENTITY-LINKING-001",
		Title:       "Record relationships readiness incident",
	}
}

func RecordUsers() map[string]UserFixture {
	return map[string]UserFixture{
		"viewer": {
			UserID:      golden.RecordViewerUserID,
			Role:        "viewer",
			Email:       "viewer.records@example.test",
			DisplayName: "Viewer Record",
		},
		"editor": {
			UserID:      golden.RecordEditorUserID,
			Role:        "editor",
			Email:       "editor.records@example.test",
			DisplayName: "Editor Record",
		},
		"reviewer": {
			UserID:      golden.RecordReviewerUserID,
			Role:        "reviewer",
			Email:       "reviewer.records@example.test",
			DisplayName: "Reviewer Record",
		},
		"admin": {
			UserID:      golden.RecordAdminUserID,
			Role:        "admin",
			Email:       "admin.records@example.test",
			DisplayName: "Admin Record",
		},
		"nonmember": {
			UserID:      golden.RecordNonMemberUserID,
			Role:        "nonmember",
			Email:       "nonmember.records@example.test",
			DisplayName: "Nonmember Record",
		},
	}
}

func RecordMemberships() map[string]MembershipFixture {
	users := RecordUsers()
	return map[string]MembershipFixture{
		"viewer":   {IncidentID: golden.RecordIncidentID, UserID: users["viewer"].UserID, Role: users["viewer"].Role},
		"editor":   {IncidentID: golden.RecordIncidentID, UserID: users["editor"].UserID, Role: users["editor"].Role},
		"reviewer": {IncidentID: golden.RecordIncidentID, UserID: users["reviewer"].UserID, Role: users["reviewer"].Role},
		"admin":    {IncidentID: golden.RecordIncidentID, UserID: users["admin"].UserID, Role: users["admin"].Role},
	}
}

func RecordViews() map[string]ViewFixture {
	return map[string]ViewFixture{
		"timeline": {
			ViewSchemaID: golden.RecordTimelineViewSchemaID,
			Title:        "Timeline",
			FieldBinding: map[string]string{
				golden.RecordFieldTimelineHostRefs:     golden.RecordBindingMentionOrigin,
				golden.RecordFieldTimelineIdentityRefs: golden.RecordBindingMentionOrigin,
			},
		},
		"hosts": {
			ViewSchemaID: golden.RecordHostsViewSchemaID,
			Title:        "Hosts",
			FieldBinding: map[string]string{
				"host.display_name": golden.RecordBindingEntityOrigin,
				"host.hostname":     golden.RecordBindingEntityOrigin,
				"host.aliases":      golden.RecordBindingEntityOrigin,
			},
		},
		"identities": {
			ViewSchemaID: golden.RecordIdentitiesViewSchemaID,
			Title:        "Identities",
			FieldBinding: map[string]string{
				"identity.display_name":     golden.RecordBindingEntityOrigin,
				"identity.upn":              golden.RecordBindingEntityOrigin,
				"identity.email":            golden.RecordBindingEntityOrigin,
				"identity.sam_account_name": golden.RecordBindingEntityOrigin,
			},
		},
		"indicators": {
			ViewSchemaID: golden.RecordIndicatorsViewSchemaID,
			Title:        "Indicators",
			FieldBinding: map[string]string{},
		},
	}
}

func RecordTimelineRows() map[string]TimelineRowFixture {
	return map[string]TimelineRowFixture{
		"rough": {
			RecordID:   golden.RecordTimelineRecordID,
			RowVersion: 7,
			Summary:    "Rough timeline row",
			SourceText: "host WS-023 and user analyst@example.test",
		},
		"host_only": {
			RecordID:   golden.RecordTimelineSiblingRecordID,
			RowVersion: 2,
			Summary:    "Host mention row",
			HostTokens: []string{"WS-023"},
		},
		"mixed": {
			RecordID:     golden.RecordTimelineMixedRecordID,
			RowVersion:   5,
			Summary:      "Mixed mention row",
			HostTokens:   []string{"WS-023", "VPN Gateway"},
			IdentityText: []string{"analyst@example.test"},
		},
	}
}

func RecordHosts() map[string]HostFixture {
	mergedInto := golden.RecordCanonicalHostRecordID
	return map[string]HostFixture{
		"canonical": {
			RecordID:    golden.RecordCanonicalHostRecordID,
			State:       golden.RecordRecordStateCanonical,
			DisplayName: "WS-023",
			AADDeviceID: "aad-device-ws-023",
			FQDN:        "ws-023.corp.example.test",
			Hostname:    "WS-023",
			PreservedIdentifiers: []PreservedIdentifierFixture{
				{IdentifierClass: golden.RecordIdentifierExactMatchReuse, IdentifierType: "fqdn", Value: "ws-023.alt.example.test"},
				{IdentifierClass: golden.RecordIdentifierSuggestionOnly, IdentifierType: "alias", Value: "ws023"},
				{IdentifierClass: golden.RecordIdentifierProvenanceOnly, IdentifierType: "cmdb_name", Value: "workstation-23"},
			},
			SuggestionOnlyAliases: []string{"Workstation 23"},
		},
		"stub": {
			RecordID:    golden.RecordStubHostRecordID,
			State:       golden.RecordRecordStateStub,
			DisplayName: "VPN Gateway",
			Hostname:    "VPN Gateway",
		},
		"merged": {
			RecordID:           golden.RecordMergedHostRecordID,
			State:              golden.RecordRecordStateMerged,
			DisplayName:        "Old WS-023",
			Hostname:           "WS-023-OLD",
			MergedIntoRecordID: &mergedInto,
		},
		"duplicate": {
			RecordID:    golden.RecordDuplicateHostRecordID,
			State:       golden.RecordRecordStateCanonical,
			DisplayName: "WS-023 duplicate",
			FQDN:        "ws-023.duplicate.example.test",
			Hostname:    "WS-023",
			PreservedIdentifiers: []PreservedIdentifierFixture{
				{IdentifierClass: golden.RecordIdentifierExactMatchReuse, IdentifierType: "hostname", Value: "WS-023"},
			},
		},
	}
}

func RecordIdentities() map[string]IdentityFixture {
	mergedInto := golden.RecordCanonicalIdentityID
	return map[string]IdentityFixture{
		"canonical": {
			RecordID:       golden.RecordCanonicalIdentityID,
			State:          golden.RecordRecordStateCanonical,
			DisplayName:    "Alex Analyst",
			AADObjectID:    "aad-object-alex-001",
			SID:            "S-1-5-21-111-222-333-1001",
			UPN:            "alex.analyst@example.test",
			Email:          "alex.analyst@example.test",
			SamAccountName: "ALEXA",
			PreservedIdentifiers: []PreservedIdentifierFixture{
				{IdentifierClass: golden.RecordIdentifierExactMatchReuse, IdentifierType: "email", Value: "analyst@example.test"},
				{IdentifierClass: golden.RecordIdentifierSuggestionOnly, IdentifierType: "alias", Value: "Alex"},
				{IdentifierClass: golden.RecordIdentifierProvenanceOnly, IdentifierType: "ticket_actor", Value: "Case Owner"},
			},
			SuggestionOnlyAliases: []string{"Analyst Alex"},
		},
		"stub": {
			RecordID:       golden.RecordStubIdentityID,
			State:          golden.RecordRecordStateStub,
			DisplayName:    "VPN User",
			Email:          "vpn.user@example.test",
			SamAccountName: "VPNUSER",
		},
		"merged": {
			RecordID:           golden.RecordMergedIdentityID,
			State:              golden.RecordRecordStateMerged,
			DisplayName:        "Legacy Analyst",
			UPN:                "legacy.analyst@example.test",
			MergedIntoRecordID: &mergedInto,
		},
		"duplicate": {
			RecordID:       golden.RecordDuplicateIdentityID,
			State:          golden.RecordRecordStateCanonical,
			DisplayName:    "Alex Analyst duplicate",
			UPN:            "alex.analyst+dup@example.test",
			Email:          "alex.analyst@example.test",
			SamAccountName: "ALEXA-DUP",
			PreservedIdentifiers: []PreservedIdentifierFixture{
				{IdentifierClass: golden.RecordIdentifierExactMatchReuse, IdentifierType: "email", Value: "alex.analyst@example.test"},
			},
		},
	}
}

func RecordMentions() map[string]EntityMentionFixture {
	resolvedAt := golden.RecordBaseTime
	resolutionMethod := "explicit_resolve_route"
	return map[string]EntityMentionFixture{
		"host_unresolved": {
			EntityMentionID:  golden.RecordHostMentionID,
			SourceRecordID:   golden.RecordTimelineRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.RecordFieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:1/cell:timeline.host_refs/item:1",
			RawText:          "WS-023",
			NormalizedText:   "WS-023",
			ResolutionStatus: golden.RecordMentionStatusUnresolved,
			RowVersion:       3,
			Ordinal:          1,
		},
		"host_resolved": {
			EntityMentionID:  golden.RecordResolvedHostMentionID,
			SourceRecordID:   golden.RecordTimelineRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.RecordFieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:1/cell:timeline.host_refs/item:2",
			RawText:          "WS-023",
			NormalizedText:   "WS-023",
			ResolutionStatus: golden.RecordMentionStatusResolved,
			RowVersion:       4,
			ResolvedRecordID: uuidPointer(golden.RecordCanonicalHostRecordID),
			ResolvedByUserID: uuidPointer(golden.RecordReviewerUserID),
			ResolvedAt:       &resolvedAt,
			ResolutionMethod: stringPointer(resolutionMethod),
			Ordinal:          2,
		},
		"host_dismissed": {
			EntityMentionID:  golden.RecordDismissedHostMentionID,
			SourceRecordID:   golden.RecordTimelineMixedRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.RecordFieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:3/cell:timeline.host_refs/item:1",
			RawText:          "WS-023 maybe",
			NormalizedText:   "WS-023 maybe",
			ResolutionStatus: golden.RecordMentionStatusDismissed,
			RowVersion:       6,
			Ordinal:          1,
		},
		"identity_unresolved": {
			EntityMentionID:  golden.RecordIdentityMentionID,
			SourceRecordID:   golden.RecordTimelineMixedRecordID,
			EntityType:       "identity",
			SourceFieldKey:   golden.RecordFieldTimelineIdentityRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:3/cell:timeline.identity_refs/item:1",
			RawText:          "analyst@example.test",
			NormalizedText:   "analyst@example.test",
			ResolutionStatus: golden.RecordMentionStatusUnresolved,
			RowVersion:       2,
			Ordinal:          1,
		},
		"identity_resolved": {
			EntityMentionID:  golden.RecordResolvedIdentityID,
			SourceRecordID:   golden.RecordTimelineMixedRecordID,
			EntityType:       "identity",
			SourceFieldKey:   golden.RecordFieldTimelineIdentityRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:3/cell:timeline.identity_refs/item:2",
			RawText:          "alex.analyst@example.test",
			NormalizedText:   "alex.analyst@example.test",
			ResolutionStatus: golden.RecordMentionStatusResolved,
			RowVersion:       5,
			ResolvedRecordID: uuidPointer(golden.RecordCanonicalIdentityID),
			ResolvedByUserID: uuidPointer(golden.RecordReviewerUserID),
			ResolvedAt:       &resolvedAt,
			ResolutionMethod: stringPointer(resolutionMethod),
			Ordinal:          2,
		},
		"identity_dismissed": {
			EntityMentionID:  golden.RecordDismissedIdentityID,
			SourceRecordID:   golden.RecordTimelineMixedRecordID,
			EntityType:       "identity",
			SourceFieldKey:   golden.RecordFieldTimelineIdentityRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:3/cell:timeline.identity_refs/item:3",
			RawText:          "unknown.user@example.test",
			NormalizedText:   "unknown.user@example.test",
			ResolutionStatus: golden.RecordMentionStatusDismissed,
			RowVersion:       7,
			Ordinal:          3,
		},
		"repeated_distinct_source_rows": {
			EntityMentionID:  uuid.MustParse("40000000-0000-0000-0000-000000000607"),
			SourceRecordID:   golden.RecordTimelineSiblingRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.RecordFieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:2/cell:timeline.host_refs/item:1",
			RawText:          "WS-023",
			NormalizedText:   "WS-023",
			ResolutionStatus: golden.RecordMentionStatusUnresolved,
			RowVersion:       1,
			Ordinal:          1,
		},
		"repeated_distinct_locator": {
			EntityMentionID:  uuid.MustParse("40000000-0000-0000-0000-000000000608"),
			SourceRecordID:   golden.RecordTimelineRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.RecordFieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:1/cell:timeline.host_refs/item:3",
			RawText:          "WS-023",
			NormalizedText:   "WS-023",
			ResolutionStatus: golden.RecordMentionStatusUnresolved,
			RowVersion:       1,
			Ordinal:          3,
		},
	}
}

func RecordLinks() map[string]LinkFixture {
	return map[string]LinkFixture{
		"timeline_to_host_manual": {
			RecordLinkID: golden.RecordManualLinkID,
			IncidentID:   golden.RecordIncidentID,
			SourceID:     golden.RecordTimelineRecordID,
			TargetID:     golden.RecordCanonicalHostRecordID,
			LinkType:     "observed_on_host",
			Provenance:   golden.RecordLinkProvenanceManual,
		},
		"timeline_to_identity_manual": {
			RecordLinkID: uuid.MustParse("40000000-0000-0000-0000-000000000704"),
			IncidentID:   golden.RecordIncidentID,
			SourceID:     golden.RecordTimelineRecordID,
			TargetID:     golden.RecordCanonicalIdentityID,
			LinkType:     "observed_as_identity",
			Provenance:   golden.RecordLinkProvenanceManual,
		},
		"timeline_to_host_auto_match": {
			RecordLinkID: golden.RecordAutoMatchLinkID,
			IncidentID:   golden.RecordIncidentID,
			SourceID:     golden.RecordTimelineSiblingRecordID,
			TargetID:     golden.RecordCanonicalHostRecordID,
			LinkType:     "observed_on_host",
			Provenance:   golden.RecordLinkProvenanceAutoMatch,
			Confidence:   intPointer(100),
		},
		"duplicate_merge_candidate": {
			RecordLinkID: golden.RecordDuplicateLinkID,
			IncidentID:   golden.RecordIncidentID,
			SourceID:     golden.RecordTimelineMixedRecordID,
			TargetID:     golden.RecordDuplicateHostRecordID,
			LinkType:     "observed_on_host",
			Provenance:   golden.RecordLinkProvenanceManual,
		},
	}
}

func RecordTags() map[string]TagFixture {
	return map[string]TagFixture{
		"survivor": {TagID: golden.RecordTagIDSurvivor, RecordID: golden.RecordCanonicalHostRecordID, TagName: "critical-host"},
		"loser":    {TagID: golden.RecordTagIDLoser, RecordID: golden.RecordDuplicateHostRecordID, TagName: "critical-host"},
	}
}

func RecordAssessments() map[string]AssessmentFixture {
	return map[string]AssessmentFixture{
		"host": {
			AssessmentID: golden.RecordAssessmentHostID,
			SubjectID:    golden.RecordCanonicalHostRecordID,
			SubjectType:  "host",
			State:        "confirmed",
			AssessedAt:   golden.RecordPastTime,
		},
		"identity": {
			AssessmentID: golden.RecordAssessmentIdentID,
			SubjectID:    golden.RecordCanonicalIdentityID,
			SubjectType:  "identity",
			State:        "suspected",
			AssessedAt:   golden.RecordBaseTime,
		},
		"loser": {
			AssessmentID: uuid.MustParse("40000000-0000-0000-0000-000000000903"),
			SubjectID:    golden.RecordDuplicateIdentityID,
			SubjectType:  "identity",
			State:        "suspected",
			AssessedAt:   golden.RecordPastTime,
		},
	}
}

func RecordIndicators() map[string]IndicatorFixture {
	return map[string]IndicatorFixture{
		"canonical": {
			RecordID:         golden.RecordCanonicalIndicatorID,
			IndicatorType:    golden.RecordIndicatorExamples[0].IndicatorType,
			ValueKind:        golden.RecordIndicatorExamples[0].ValueKind,
			DisplayValue:     golden.RecordIndicatorExamples[0].DisplayValue,
			NormalizedValue:  golden.RecordIndicatorExamples[0].NormalizedValue,
			DefangedValue:    golden.RecordIndicatorExamples[0].DefangedValue,
			ObservationCount: 2,
			FirstObservedAt:  golden.RecordPastTime,
			LastObservedAt:   golden.RecordBaseTime,
		},
		"stix_pattern": {
			RecordID:         uuid.MustParse("40000000-0000-0000-0000-000000000504"),
			IndicatorType:    golden.RecordIndicatorExamples[3].IndicatorType,
			ValueKind:        golden.RecordIndicatorExamples[3].ValueKind,
			DisplayValue:     golden.RecordIndicatorExamples[3].DisplayValue,
			NormalizedValue:  golden.RecordIndicatorExamples[3].NormalizedValue,
			STIXPattern:      golden.RecordIndicatorExamples[3].STIXPattern,
			HashAlgorithm:    golden.RecordIndicatorExamples[3].HashAlgorithm,
			HashValue:        golden.RecordIndicatorExamples[3].HashValue,
			ObservationCount: 1,
			FirstObservedAt:  golden.RecordBaseTime,
			LastObservedAt:   golden.RecordBaseTime,
		},
	}
}

func RecordIndicatorObservations() map[string]IndicatorObservationFixture {
	return map[string]IndicatorObservationFixture{
		"source_bound": {
			ObservationID:       golden.RecordIndicatorObservationID,
			SourceRecordID:      golden.RecordTimelineRecordID,
			SourceFieldKey:      golden.RecordFieldTimelineSourceText,
			OriginKind:          "interactive_cell",
			OriginLocator:       "view:timeline/row:1/cell:timeline.raw_activity_text/span:12-24",
			ObservedText:        "203[.]0[.]113[.]24",
			ParsedIndicatorType: golden.RecordIndicatorTypeIPv4,
			NormalizedCandidate: "203.0.113.24",
			ResolutionStatus:    golden.RecordMentionStatusUnresolved,
		},
		"distinct_source_same_value": {
			ObservationID:             uuid.MustParse("40000000-0000-0000-0000-000000000505"),
			SourceRecordID:            golden.RecordTimelineSiblingRecordID,
			SourceFieldKey:            golden.RecordFieldTimelineSummary,
			OriginKind:                "interactive_cell",
			OriginLocator:             "view:timeline/row:2/cell:timeline.activity_synopsis_text/span:5-17",
			ObservedText:              "203[.]0[.]113[.]24",
			ParsedIndicatorType:       golden.RecordIndicatorTypeIPv4,
			NormalizedCandidate:       "203.0.113.24",
			ResolutionStatus:          golden.RecordMentionStatusResolved,
			ResolvedIndicatorRecordID: uuidPointer(golden.RecordCanonicalIndicatorID),
		},
	}
}

func RecordIndicatorIntervals() map[string]IndicatorLifecycleFixture {
	return map[string]IndicatorLifecycleFixture{
		"active": {
			IntervalID:     golden.RecordIndicatorIntervalID,
			IndicatorID:    golden.RecordCanonicalIndicatorID,
			LifecycleState: "active",
			StartedAt:      golden.RecordPastTime,
		},
	}
}

func CollectionActions(actions ...map[string]any) map[string]any {
	return map[string]any{
		"kind":    "collection_actions_v1",
		"actions": append([]map[string]any(nil), actions...),
	}
}

func AddTokenAction(rawText string) map[string]any {
	return map[string]any{
		"op":       "add_token",
		"raw_text": rawText,
	}
}

func AddResolvedRefAction(rawText string, resolvedRecordID uuid.UUID) map[string]any {
	return map[string]any{
		"op":                 "add_resolved_ref",
		"raw_text":           rawText,
		"resolved_record_id": resolvedRecordID.String(),
	}
}

func ResolveItemAction(itemRef string, resolvedRecordID uuid.UUID) map[string]any {
	return map[string]any{
		"op":                 "resolve_item",
		"item_ref":           itemRef,
		"resolved_record_id": resolvedRecordID.String(),
	}
}

func RevertToUnresolvedAction(itemRef string) map[string]any {
	return map[string]any{
		"op":       "revert_to_unresolved",
		"item_ref": itemRef,
	}
}

func DismissItemAction(itemRef string) map[string]any {
	return map[string]any{
		"op":       "dismiss_item",
		"item_ref": itemRef,
	}
}

func TimelineCollectionPatchPayload(fieldKey string, baseRowVersion int64, clientTxnID string, actionPayload map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id":   golden.RecordTimelineViewSchemaID,
		"base_row_version": baseRowVersion,
		"client_txn_id":    clientTxnID,
		"changes": []map[string]any{
			{
				"field_key":      fieldKey,
				"action_payload": actionPayload,
			},
		},
	}
}

func MentionResolveRoutePayload(baseMentionRowVersion int64, clientTxnID string, action string, resolvedRecordID *uuid.UUID, reason *string) map[string]any {
	payload := map[string]any{
		"base_mention_row_version": baseMentionRowVersion,
		"client_txn_id":            clientTxnID,
		"action":                   action,
	}
	if resolvedRecordID != nil {
		payload["resolved_record_id"] = resolvedRecordID.String()
	}
	if reason != nil {
		payload["reason"] = *reason
	}
	return payload
}

func HostCreatePayload(clientTxnID string) map[string]any {
	return map[string]any{
		"client_txn_id":     clientTxnID,
		"host.display_name": RecordHosts()["stub"].DisplayName,
		"host.hostname":     RecordHosts()["stub"].Hostname,
	}
}

func IdentityCreatePayload(clientTxnID string) map[string]any {
	return map[string]any{
		"client_txn_id":             clientTxnID,
		"identity.display_name":     RecordIdentities()["stub"].DisplayName,
		"identity.email":            RecordIdentities()["stub"].Email,
		"identity.sam_account_name": RecordIdentities()["stub"].SamAccountName,
	}
}

func IndicatorCreatePayload(clientTxnID string) map[string]any {
	indicator := golden.RecordIndicatorExamples[0]
	return map[string]any{
		"client_txn_id":              clientTxnID,
		"indicator.indicator_type":   indicator.IndicatorType,
		"indicator.value_kind":       indicator.ValueKind,
		"indicator.display_value":    indicator.DisplayValue,
		"indicator.normalized_value": indicator.NormalizedValue,
		"indicator.defanged_value":   indicator.DefangedValue,
	}
}

func MentionItemRef(mentionID uuid.UUID) string {
	return fmt.Sprintf("entity_mention:%s", mentionID.String())
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

func intPointer(value int) *int {
	return &value
}
