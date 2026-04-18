package fixtures

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/golden"
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

func Phase4Incident() IncidentFixture {
	return IncidentFixture{
		IncidentID:  golden.Phase4IncidentID,
		IncidentKey: "IR-PHASE4-001",
		Title:       "Phase 4 readiness incident",
	}
}

func Phase4Users() map[string]UserFixture {
	return map[string]UserFixture{
		"viewer": {
			UserID:      golden.Phase4ViewerUserID,
			Role:        "viewer",
			Email:       "viewer.phase4@example.test",
			DisplayName: "Viewer Phase4",
		},
		"editor": {
			UserID:      golden.Phase4EditorUserID,
			Role:        "editor",
			Email:       "editor.phase4@example.test",
			DisplayName: "Editor Phase4",
		},
		"reviewer": {
			UserID:      golden.Phase4ReviewerUserID,
			Role:        "reviewer",
			Email:       "reviewer.phase4@example.test",
			DisplayName: "Reviewer Phase4",
		},
		"admin": {
			UserID:      golden.Phase4AdminUserID,
			Role:        "admin",
			Email:       "admin.phase4@example.test",
			DisplayName: "Admin Phase4",
		},
		"nonmember": {
			UserID:      golden.Phase4NonMemberUserID,
			Role:        "nonmember",
			Email:       "nonmember.phase4@example.test",
			DisplayName: "Nonmember Phase4",
		},
	}
}

func Phase4Memberships() map[string]MembershipFixture {
	users := Phase4Users()
	return map[string]MembershipFixture{
		"viewer":   {IncidentID: golden.Phase4IncidentID, UserID: users["viewer"].UserID, Role: users["viewer"].Role},
		"editor":   {IncidentID: golden.Phase4IncidentID, UserID: users["editor"].UserID, Role: users["editor"].Role},
		"reviewer": {IncidentID: golden.Phase4IncidentID, UserID: users["reviewer"].UserID, Role: users["reviewer"].Role},
		"admin":    {IncidentID: golden.Phase4IncidentID, UserID: users["admin"].UserID, Role: users["admin"].Role},
	}
}

func Phase4Views() map[string]ViewFixture {
	return map[string]ViewFixture{
		"timeline": {
			ViewSchemaID: golden.Phase4TimelineViewSchemaID,
			Title:        "Timeline",
			FieldBinding: map[string]string{
				golden.Phase4FieldTimelineHostRefs:     golden.Phase4BindingMentionOrigin,
				golden.Phase4FieldTimelineIdentityRefs: golden.Phase4BindingMentionOrigin,
			},
		},
		"hosts": {
			ViewSchemaID: golden.Phase4HostsViewSchemaID,
			Title:        "Hosts",
			FieldBinding: map[string]string{
				"host.display_name": golden.Phase4BindingEntityOrigin,
				"host.hostname":     golden.Phase4BindingEntityOrigin,
				"host.aliases":      golden.Phase4BindingEntityOrigin,
			},
		},
		"identities": {
			ViewSchemaID: golden.Phase4IdentitiesViewSchemaID,
			Title:        "Identities",
			FieldBinding: map[string]string{
				"identity.display_name":     golden.Phase4BindingEntityOrigin,
				"identity.upn":              golden.Phase4BindingEntityOrigin,
				"identity.email":            golden.Phase4BindingEntityOrigin,
				"identity.sam_account_name": golden.Phase4BindingEntityOrigin,
			},
		},
		"indicators": {
			ViewSchemaID: golden.Phase4IndicatorsViewSchemaID,
			Title:        "Indicators",
			FieldBinding: map[string]string{},
		},
	}
}

func Phase4TimelineRows() map[string]TimelineRowFixture {
	return map[string]TimelineRowFixture{
		"rough": {
			RecordID:   golden.Phase4TimelineRecordID,
			RowVersion: 7,
			Summary:    "Rough timeline row",
			SourceText: "host WS-023 and user analyst@example.test",
		},
		"host_only": {
			RecordID:   golden.Phase4TimelineSiblingRecordID,
			RowVersion: 2,
			Summary:    "Host mention row",
			HostTokens: []string{"WS-023"},
		},
		"mixed": {
			RecordID:     golden.Phase4TimelineMixedRecordID,
			RowVersion:   5,
			Summary:      "Mixed mention row",
			HostTokens:   []string{"WS-023", "VPN Gateway"},
			IdentityText: []string{"analyst@example.test"},
		},
	}
}

func Phase4Hosts() map[string]HostFixture {
	mergedInto := golden.Phase4CanonicalHostRecordID
	return map[string]HostFixture{
		"canonical": {
			RecordID:    golden.Phase4CanonicalHostRecordID,
			State:       golden.Phase4RecordStateCanonical,
			DisplayName: "WS-023",
			AADDeviceID: "aad-device-ws-023",
			FQDN:        "ws-023.corp.example.test",
			Hostname:    "WS-023",
			PreservedIdentifiers: []PreservedIdentifierFixture{
				{IdentifierClass: golden.Phase4IdentifierExactMatchReuse, IdentifierType: "fqdn", Value: "ws-023.alt.example.test"},
				{IdentifierClass: golden.Phase4IdentifierSuggestionOnly, IdentifierType: "alias", Value: "ws023"},
				{IdentifierClass: golden.Phase4IdentifierProvenanceOnly, IdentifierType: "cmdb_name", Value: "workstation-23"},
			},
			SuggestionOnlyAliases: []string{"Workstation 23"},
		},
		"stub": {
			RecordID:    golden.Phase4StubHostRecordID,
			State:       golden.Phase4RecordStateStub,
			DisplayName: "VPN Gateway",
			Hostname:    "VPN Gateway",
		},
		"merged": {
			RecordID:           golden.Phase4MergedHostRecordID,
			State:              golden.Phase4RecordStateMerged,
			DisplayName:        "Old WS-023",
			Hostname:           "WS-023-OLD",
			MergedIntoRecordID: &mergedInto,
		},
		"duplicate": {
			RecordID:    golden.Phase4DuplicateHostRecordID,
			State:       golden.Phase4RecordStateCanonical,
			DisplayName: "WS-023 duplicate",
			FQDN:        "ws-023.duplicate.example.test",
			Hostname:    "WS-023",
			PreservedIdentifiers: []PreservedIdentifierFixture{
				{IdentifierClass: golden.Phase4IdentifierExactMatchReuse, IdentifierType: "hostname", Value: "WS-023"},
			},
		},
	}
}

func Phase4Identities() map[string]IdentityFixture {
	mergedInto := golden.Phase4CanonicalIdentityID
	return map[string]IdentityFixture{
		"canonical": {
			RecordID:       golden.Phase4CanonicalIdentityID,
			State:          golden.Phase4RecordStateCanonical,
			DisplayName:    "Alex Analyst",
			AADObjectID:    "aad-object-alex-001",
			SID:            "S-1-5-21-111-222-333-1001",
			UPN:            "alex.analyst@example.test",
			Email:          "alex.analyst@example.test",
			SamAccountName: "ALEXA",
			PreservedIdentifiers: []PreservedIdentifierFixture{
				{IdentifierClass: golden.Phase4IdentifierExactMatchReuse, IdentifierType: "email", Value: "analyst@example.test"},
				{IdentifierClass: golden.Phase4IdentifierSuggestionOnly, IdentifierType: "alias", Value: "Alex"},
				{IdentifierClass: golden.Phase4IdentifierProvenanceOnly, IdentifierType: "ticket_actor", Value: "Case Owner"},
			},
			SuggestionOnlyAliases: []string{"Analyst Alex"},
		},
		"stub": {
			RecordID:       golden.Phase4StubIdentityID,
			State:          golden.Phase4RecordStateStub,
			DisplayName:    "VPN User",
			Email:          "vpn.user@example.test",
			SamAccountName: "VPNUSER",
		},
		"merged": {
			RecordID:           golden.Phase4MergedIdentityID,
			State:              golden.Phase4RecordStateMerged,
			DisplayName:        "Legacy Analyst",
			UPN:                "legacy.analyst@example.test",
			MergedIntoRecordID: &mergedInto,
		},
		"duplicate": {
			RecordID:       golden.Phase4DuplicateIdentityID,
			State:          golden.Phase4RecordStateCanonical,
			DisplayName:    "Alex Analyst duplicate",
			UPN:            "alex.analyst+dup@example.test",
			Email:          "alex.analyst@example.test",
			SamAccountName: "ALEXA-DUP",
			PreservedIdentifiers: []PreservedIdentifierFixture{
				{IdentifierClass: golden.Phase4IdentifierExactMatchReuse, IdentifierType: "email", Value: "alex.analyst@example.test"},
			},
		},
	}
}

func Phase4Mentions() map[string]EntityMentionFixture {
	resolvedAt := golden.Phase4BaseTime
	resolutionMethod := "explicit_resolve_route"
	return map[string]EntityMentionFixture{
		"host_unresolved": {
			EntityMentionID:  golden.Phase4HostMentionID,
			SourceRecordID:   golden.Phase4TimelineRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.Phase4FieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:1/cell:timeline.host_refs/item:1",
			RawText:          "WS-023",
			NormalizedText:   "WS-023",
			ResolutionStatus: golden.Phase4MentionStatusUnresolved,
			RowVersion:       3,
			Ordinal:          1,
		},
		"host_resolved": {
			EntityMentionID:  golden.Phase4ResolvedHostMentionID,
			SourceRecordID:   golden.Phase4TimelineRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.Phase4FieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:1/cell:timeline.host_refs/item:2",
			RawText:          "WS-023",
			NormalizedText:   "WS-023",
			ResolutionStatus: golden.Phase4MentionStatusResolved,
			RowVersion:       4,
			ResolvedRecordID: uuidPointer(golden.Phase4CanonicalHostRecordID),
			ResolvedByUserID: uuidPointer(golden.Phase4ReviewerUserID),
			ResolvedAt:       &resolvedAt,
			ResolutionMethod: stringPointer(resolutionMethod),
			Ordinal:          2,
		},
		"host_dismissed": {
			EntityMentionID:  golden.Phase4DismissedHostMentionID,
			SourceRecordID:   golden.Phase4TimelineMixedRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.Phase4FieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:3/cell:timeline.host_refs/item:1",
			RawText:          "WS-023 maybe",
			NormalizedText:   "WS-023 maybe",
			ResolutionStatus: golden.Phase4MentionStatusDismissed,
			RowVersion:       6,
			Ordinal:          1,
		},
		"identity_unresolved": {
			EntityMentionID:  golden.Phase4IdentityMentionID,
			SourceRecordID:   golden.Phase4TimelineMixedRecordID,
			EntityType:       "identity",
			SourceFieldKey:   golden.Phase4FieldTimelineIdentityRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:3/cell:timeline.identity_refs/item:1",
			RawText:          "analyst@example.test",
			NormalizedText:   "analyst@example.test",
			ResolutionStatus: golden.Phase4MentionStatusUnresolved,
			RowVersion:       2,
			Ordinal:          1,
		},
		"identity_resolved": {
			EntityMentionID:  golden.Phase4ResolvedIdentityID,
			SourceRecordID:   golden.Phase4TimelineMixedRecordID,
			EntityType:       "identity",
			SourceFieldKey:   golden.Phase4FieldTimelineIdentityRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:3/cell:timeline.identity_refs/item:2",
			RawText:          "alex.analyst@example.test",
			NormalizedText:   "alex.analyst@example.test",
			ResolutionStatus: golden.Phase4MentionStatusResolved,
			RowVersion:       5,
			ResolvedRecordID: uuidPointer(golden.Phase4CanonicalIdentityID),
			ResolvedByUserID: uuidPointer(golden.Phase4ReviewerUserID),
			ResolvedAt:       &resolvedAt,
			ResolutionMethod: stringPointer(resolutionMethod),
			Ordinal:          2,
		},
		"identity_dismissed": {
			EntityMentionID:  golden.Phase4DismissedIdentityID,
			SourceRecordID:   golden.Phase4TimelineMixedRecordID,
			EntityType:       "identity",
			SourceFieldKey:   golden.Phase4FieldTimelineIdentityRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:3/cell:timeline.identity_refs/item:3",
			RawText:          "unknown.user@example.test",
			NormalizedText:   "unknown.user@example.test",
			ResolutionStatus: golden.Phase4MentionStatusDismissed,
			RowVersion:       7,
			Ordinal:          3,
		},
		"repeated_distinct_source_rows": {
			EntityMentionID:  uuid.MustParse("40000000-0000-0000-0000-000000000607"),
			SourceRecordID:   golden.Phase4TimelineSiblingRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.Phase4FieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:2/cell:timeline.host_refs/item:1",
			RawText:          "WS-023",
			NormalizedText:   "WS-023",
			ResolutionStatus: golden.Phase4MentionStatusUnresolved,
			RowVersion:       1,
			Ordinal:          1,
		},
		"repeated_distinct_locator": {
			EntityMentionID:  uuid.MustParse("40000000-0000-0000-0000-000000000608"),
			SourceRecordID:   golden.Phase4TimelineRecordID,
			EntityType:       "host",
			SourceFieldKey:   golden.Phase4FieldTimelineHostRefs,
			OriginKind:       "interactive_cell",
			OriginLocator:    "view:timeline/row:1/cell:timeline.host_refs/item:3",
			RawText:          "WS-023",
			NormalizedText:   "WS-023",
			ResolutionStatus: golden.Phase4MentionStatusUnresolved,
			RowVersion:       1,
			Ordinal:          3,
		},
	}
}

func Phase4Links() map[string]LinkFixture {
	return map[string]LinkFixture{
		"timeline_to_host_manual": {
			RecordLinkID: golden.Phase4ManualLinkID,
			IncidentID:   golden.Phase4IncidentID,
			SourceID:     golden.Phase4TimelineRecordID,
			TargetID:     golden.Phase4CanonicalHostRecordID,
			LinkType:     "observed_on_host",
			Provenance:   golden.Phase4LinkProvenanceManual,
		},
		"timeline_to_identity_manual": {
			RecordLinkID: uuid.MustParse("40000000-0000-0000-0000-000000000704"),
			IncidentID:   golden.Phase4IncidentID,
			SourceID:     golden.Phase4TimelineRecordID,
			TargetID:     golden.Phase4CanonicalIdentityID,
			LinkType:     "observed_as_identity",
			Provenance:   golden.Phase4LinkProvenanceManual,
		},
		"timeline_to_host_auto_match": {
			RecordLinkID: golden.Phase4AutoMatchLinkID,
			IncidentID:   golden.Phase4IncidentID,
			SourceID:     golden.Phase4TimelineSiblingRecordID,
			TargetID:     golden.Phase4CanonicalHostRecordID,
			LinkType:     "observed_on_host",
			Provenance:   golden.Phase4LinkProvenanceAutoMatch,
			Confidence:   intPointer(100),
		},
		"duplicate_merge_candidate": {
			RecordLinkID: golden.Phase4DuplicateLinkID,
			IncidentID:   golden.Phase4IncidentID,
			SourceID:     golden.Phase4TimelineMixedRecordID,
			TargetID:     golden.Phase4DuplicateHostRecordID,
			LinkType:     "observed_on_host",
			Provenance:   golden.Phase4LinkProvenanceManual,
		},
	}
}

func Phase4Tags() map[string]TagFixture {
	return map[string]TagFixture{
		"survivor": {TagID: golden.Phase4TagIDSurvivor, RecordID: golden.Phase4CanonicalHostRecordID, TagName: "critical-host"},
		"loser":    {TagID: golden.Phase4TagIDLoser, RecordID: golden.Phase4DuplicateHostRecordID, TagName: "critical-host"},
	}
}

func Phase4Assessments() map[string]AssessmentFixture {
	return map[string]AssessmentFixture{
		"host": {
			AssessmentID: golden.Phase4AssessmentHostID,
			SubjectID:    golden.Phase4CanonicalHostRecordID,
			SubjectType:  "host",
			State:        "confirmed",
			AssessedAt:   golden.Phase4PastTime,
		},
		"identity": {
			AssessmentID: golden.Phase4AssessmentIdentID,
			SubjectID:    golden.Phase4CanonicalIdentityID,
			SubjectType:  "identity",
			State:        "suspected",
			AssessedAt:   golden.Phase4BaseTime,
		},
		"loser": {
			AssessmentID: uuid.MustParse("40000000-0000-0000-0000-000000000903"),
			SubjectID:    golden.Phase4DuplicateIdentityID,
			SubjectType:  "identity",
			State:        "suspected",
			AssessedAt:   golden.Phase4PastTime,
		},
	}
}

func Phase4Indicators() map[string]IndicatorFixture {
	return map[string]IndicatorFixture{
		"canonical": {
			RecordID:         golden.Phase4CanonicalIndicatorID,
			IndicatorType:    golden.Phase4IndicatorExamples[0].IndicatorType,
			ValueKind:        golden.Phase4IndicatorExamples[0].ValueKind,
			DisplayValue:     golden.Phase4IndicatorExamples[0].DisplayValue,
			NormalizedValue:  golden.Phase4IndicatorExamples[0].NormalizedValue,
			DefangedValue:    golden.Phase4IndicatorExamples[0].DefangedValue,
			ObservationCount: 2,
			FirstObservedAt:  golden.Phase4PastTime,
			LastObservedAt:   golden.Phase4BaseTime,
		},
		"stix_pattern": {
			RecordID:         uuid.MustParse("40000000-0000-0000-0000-000000000504"),
			IndicatorType:    golden.Phase4IndicatorExamples[3].IndicatorType,
			ValueKind:        golden.Phase4IndicatorExamples[3].ValueKind,
			DisplayValue:     golden.Phase4IndicatorExamples[3].DisplayValue,
			NormalizedValue:  golden.Phase4IndicatorExamples[3].NormalizedValue,
			STIXPattern:      golden.Phase4IndicatorExamples[3].STIXPattern,
			HashAlgorithm:    golden.Phase4IndicatorExamples[3].HashAlgorithm,
			HashValue:        golden.Phase4IndicatorExamples[3].HashValue,
			ObservationCount: 1,
			FirstObservedAt:  golden.Phase4BaseTime,
			LastObservedAt:   golden.Phase4BaseTime,
		},
	}
}

func Phase4IndicatorObservations() map[string]IndicatorObservationFixture {
	return map[string]IndicatorObservationFixture{
		"source_bound": {
			ObservationID:       golden.Phase4IndicatorObservationID,
			SourceRecordID:      golden.Phase4TimelineRecordID,
			SourceFieldKey:      golden.Phase4FieldTimelineSourceText,
			OriginKind:          "interactive_cell",
			OriginLocator:       "view:timeline/row:1/cell:timeline.source_text/span:12-24",
			ObservedText:        "203[.]0[.]113[.]24",
			ParsedIndicatorType: golden.Phase4IndicatorTypeIPv4,
			NormalizedCandidate: "203.0.113.24",
			ResolutionStatus:    golden.Phase4MentionStatusUnresolved,
		},
		"distinct_source_same_value": {
			ObservationID:             uuid.MustParse("40000000-0000-0000-0000-000000000505"),
			SourceRecordID:            golden.Phase4TimelineSiblingRecordID,
			SourceFieldKey:            golden.Phase4FieldTimelineSummary,
			OriginKind:                "interactive_cell",
			OriginLocator:             "view:timeline/row:2/cell:timeline.summary/span:5-17",
			ObservedText:              "203[.]0[.]113[.]24",
			ParsedIndicatorType:       golden.Phase4IndicatorTypeIPv4,
			NormalizedCandidate:       "203.0.113.24",
			ResolutionStatus:          golden.Phase4MentionStatusResolved,
			ResolvedIndicatorRecordID: uuidPointer(golden.Phase4CanonicalIndicatorID),
		},
	}
}

func Phase4IndicatorIntervals() map[string]IndicatorLifecycleFixture {
	return map[string]IndicatorLifecycleFixture{
		"active": {
			IntervalID:     golden.Phase4IndicatorIntervalID,
			IndicatorID:    golden.Phase4CanonicalIndicatorID,
			LifecycleState: "active",
			StartedAt:      golden.Phase4PastTime,
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
		"view_schema_id":   golden.Phase4TimelineViewSchemaID,
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
		"host.display_name": Phase4Hosts()["stub"].DisplayName,
		"host.hostname":     Phase4Hosts()["stub"].Hostname,
	}
}

func IdentityCreatePayload(clientTxnID string) map[string]any {
	return map[string]any{
		"client_txn_id":             clientTxnID,
		"identity.display_name":     Phase4Identities()["stub"].DisplayName,
		"identity.email":            Phase4Identities()["stub"].Email,
		"identity.sam_account_name": Phase4Identities()["stub"].SamAccountName,
	}
}

func IndicatorCreatePayload(clientTxnID string) map[string]any {
	indicator := golden.Phase4IndicatorExamples[0]
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
