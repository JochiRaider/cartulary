package entities_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/assertx"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

// U-4-01 / REQ-02-028..REQ-02-036 / AC-019, AC-020, AC-022.
func TestPhase4_BindingMode_U_4_01_Red(t *testing.T) {
	views := fixtures.Phase4Views()
	if got := views["timeline"].FieldBinding[golden.Phase4FieldTimelineHostRefs]; got != golden.Phase4BindingMentionOrigin {
		t.Fatalf("expected fixture binding mention_origin for %s, got %q", golden.Phase4FieldTimelineHostRefs, got)
	}
	if got := views["hosts"].FieldBinding["host.display_name"]; got != golden.Phase4BindingEntityOrigin {
		t.Fatalf("expected fixture binding entity_origin for host.display_name, got %q", got)
	}
	if got := views["identities"].FieldBinding["identity.display_name"]; got != golden.Phase4BindingEntityOrigin {
		t.Fatalf("expected fixture binding entity_origin for identity.display_name, got %q", got)
	}

	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4TimelineViewSchemaID, golden.Phase4FieldTimelineHostRefs, golden.Phase4BindingMentionOrigin)
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4TimelineViewSchemaID, golden.Phase4FieldTimelineIdentityRefs, golden.Phase4BindingMentionOrigin)
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4HostsViewSchemaID, "host.display_name", golden.Phase4BindingEntityOrigin)
	phase4test.RequireViewFieldBindingMode(t, "U-4-01", golden.Phase4IdentitiesViewSchemaID, "identity.display_name", golden.Phase4BindingEntityOrigin)
}

// U-4-02 / REQ-02-031..REQ-02-032, REQ-02-058 / AC-019, AC-021.
func TestPhase4_DuplicateMentionProvenance_U_4_02_Red(t *testing.T) {
	mentions := fixtures.Phase4Mentions()
	assertx.RequireDistinctMentionProvenance(
		t,
		mentions["host_unresolved"],
		mentions["repeated_distinct_source_rows"],
		mentions["repeated_distinct_locator"],
	)
	if mentions["host_unresolved"].RawText != mentions["repeated_distinct_source_rows"].RawText {
		t.Fatalf("expected identical repeated raw mention text, got %q and %q", mentions["host_unresolved"].RawText, mentions["repeated_distinct_source_rows"].RawText)
	}

	phase4test.RequireMigrationTables(t, "U-4-02", "entity_mentions")
}

// U-4-03 / REQ-02-034, REQ-02-038, REQ-02-054..REQ-02-055 / AC-020, AC-021, AC-186.
func TestPhase4_CreateFromMention_U_4_03_Red(t *testing.T) {
	mentions := fixtures.Phase4Mentions()
	selected := mentions["host_unresolved"]
	sibling := mentions["repeated_distinct_locator"]
	selected.ResolutionStatus = golden.Phase4MentionStatusResolved
	selected.ResolvedRecordID = &golden.Phase4StubHostRecordID

	assertx.RequireSelectedMentionOnlyResolvedByCreateFromMention(t, selected, sibling, golden.Phase4StubHostRecordID)
	assertx.RequireRawTextPreserved(t, mentions["host_unresolved"].RawText, selected.RawText)

	phase4test.RequireMigrationTables(t, "U-4-03", "hosts", "entity_mentions")
}

// U-4-04 / REQ-02-039..REQ-02-041 / AC-188..AC-190, AC-224, AC-225.
func TestPhase4_DismissRestoreMentionLifecycle_U_4_04_Red(t *testing.T) {
	mentions := fixtures.Phase4Mentions()
	before := mentions["host_resolved"]
	dismissed := before
	dismissed.ResolutionStatus = golden.Phase4MentionStatusDismissed
	dismissed.ResolvedRecordID = nil
	dismissed.ResolvedByUserID = nil
	dismissed.ResolvedAt = nil
	dismissed.ResolutionMethod = nil

	assertx.RequireDismissedMentionPreservesRowAndText(t, before, dismissed)
	assertx.RequireDismissedMentionClearsResolution(t, dismissed)

	restored := dismissed
	restored.ResolutionStatus = golden.Phase4MentionStatusUnresolved
	assertx.RequireRestoreToUnresolved(t, restored)

	phase4test.RequireMigrationTables(t, "U-4-04", "entity_mentions")
}

// U-4-05 / REQ-02-059..REQ-02-063 / AC-021, AC-022.
func TestPhase4_ExactMatchPrecedence_U_4_05_Red(t *testing.T) {
	hosts := fixtures.Phase4Hosts()
	identities := fixtures.Phase4Identities()

	assertx.RequireExactMatchPrecedence(t, golden.Phase4HostExactMatchPrecedence, []string{"aad_device_id", "fqdn", "hostname"})
	assertx.RequireExactMatchPrecedence(t, golden.Phase4IdentityExactMatchPrecedence, []string{"aad_object_id", "sid", "upn", "email", "sam_account_name"})
	assertx.RequireSuggestionBoundary(t, hosts["canonical"].PreservedIdentifiers)
	assertx.RequireSuggestionBoundary(t, identities["canonical"].PreservedIdentifiers)

	phase4test.RequireMigrationTables(t, "U-4-05", "hosts", "identities", "entity_aliases")
}

// U-4-06 / REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
func TestPhase4_ExplicitEntityMerge_U_4_06_Red(t *testing.T) {
	links := fixtures.Phase4Links()
	before := links["duplicate_merge_candidate"]
	after := before
	after.TargetID = golden.Phase4CanonicalHostRecordID
	assertx.RequireMergeRepointsLiveLink(t, after, golden.Phase4CanonicalHostRecordID)

	assessments := fixtures.Phase4Assessments()
	if assessments["loser"].SubjectID != golden.Phase4DuplicateIdentityID {
		t.Fatalf("expected loser assessment fixture to point at duplicate identity, got %#v", assessments["loser"])
	}
	assertx.RequireRawTextPreserved(t, fixtures.Phase4Mentions()["host_unresolved"].RawText, fixtures.Phase4Mentions()["host_unresolved"].RawText)

	phase4test.RequireMigrationTables(t, "U-4-06", "hosts", "identities", "record_tags", "compromise_assessments", "entity_mentions")
}

// U-4-07 / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestPhase4_IndicatorObservationSeparation_U_4_07_Red(t *testing.T) {
	indicators := fixtures.Phase4Indicators()
	observations := fixtures.Phase4IndicatorObservations()
	intervals := fixtures.Phase4IndicatorIntervals()

	assertx.RequireIndicatorObservationSeparation(t, observations["source_bound"], indicators["canonical"])
	assertx.RequireIndicatorLifecycleSeparate(t, intervals["active"], indicators["canonical"])

	phase4test.RequireMigrationTables(t, "U-4-07", "indicators", "indicator_observations", "indicator_state_intervals")
}
