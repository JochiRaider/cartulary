package assertx

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/fixtures"
)

func RequireExactlyOneChangeSet(t testing.TB, before int, after int) {
	t.Helper()
	if after-before != 1 {
		t.Fatalf("expected exactly one change_set delta, before=%d after=%d", before, after)
	}
}

func RequireNoExtraChangeSet(t testing.TB, before int, after int) {
	t.Helper()
	if after != before {
		t.Fatalf("expected no extra change_set delta, before=%d after=%d", before, after)
	}
}

func RequireActorAttribution(t testing.TB, gotActorUserID string, wantActorUserID string, gotSource string, wantSource string) {
	t.Helper()
	if gotActorUserID != wantActorUserID {
		t.Fatalf("unexpected actor attribution: got %q want %q", gotActorUserID, wantActorUserID)
	}
	if gotSource != wantSource {
		t.Fatalf("unexpected mutation source: got %q want %q", gotSource, wantSource)
	}
}

func RequireRowVersionAdvanced(t testing.TB, before int64, after int64) {
	t.Helper()
	if after <= before {
		t.Fatalf("expected row_version to advance, before=%d after=%d", before, after)
	}
}

func RequireRowVersionStable(t testing.TB, before int64, after int64) {
	t.Helper()
	if after != before {
		t.Fatalf("expected row_version to remain stable, before=%d after=%d", before, after)
	}
}

func RequireProjectionRowRecordID(t testing.TB, row map[string]any, wantRecordID string) {
	t.Helper()
	if row["record_id"] != wantRecordID {
		t.Fatalf("unexpected projection record_id: got %v want %s", row["record_id"], wantRecordID)
	}
}

func RequireActiveLink(t testing.TB, link fixtures.LinkFixture, sourceID uuid.UUID, targetID uuid.UUID, linkType string, provenance string, confidence *int) {
	t.Helper()
	if link.DeletedAt != nil {
		t.Fatalf("expected active link, got tombstoned link %#v", link)
	}
	if link.SourceID != sourceID || link.TargetID != targetID {
		t.Fatalf("unexpected link endpoints: got %s -> %s want %s -> %s", link.SourceID, link.TargetID, sourceID, targetID)
	}
	if link.LinkType != linkType {
		t.Fatalf("unexpected link_type: got %q want %q", link.LinkType, linkType)
	}
	if link.Provenance != provenance {
		t.Fatalf("unexpected provenance: got %q want %q", link.Provenance, provenance)
	}
	if !reflect.DeepEqual(link.Confidence, confidence) {
		t.Fatalf("unexpected confidence: got %#v want %#v", link.Confidence, confidence)
	}
}

func RequireNoActiveLink(t testing.TB, link fixtures.LinkFixture) {
	t.Helper()
	if link.DeletedAt == nil {
		t.Fatalf("expected no active link, got %#v", link)
	}
}

func RequireMergeRepointsLiveLink(t testing.TB, link fixtures.LinkFixture, survivorRecordID uuid.UUID) {
	t.Helper()
	if link.TargetID != survivorRecordID {
		t.Fatalf("expected merge to repoint live link to survivor=%s, got %#v", survivorRecordID, link)
	}
}

func RequireRawTextPreserved(t testing.TB, before string, after string) {
	t.Helper()
	if before != after {
		t.Fatalf("expected raw_text to remain unchanged, before=%q after=%q", before, after)
	}
}

func RequireMentionStatus(t testing.TB, mention fixtures.EntityMentionFixture, want string) {
	t.Helper()
	if mention.ResolutionStatus != want {
		t.Fatalf("unexpected mention resolution_status: got %q want %q", mention.ResolutionStatus, want)
	}
}

func RequireDismissedMentionPreservesRowAndText(t testing.TB, before fixtures.EntityMentionFixture, after fixtures.EntityMentionFixture) {
	t.Helper()
	if after.EntityMentionID != before.EntityMentionID {
		t.Fatalf("dismissed mention must preserve stable identity, before=%s after=%s", before.EntityMentionID, after.EntityMentionID)
	}
	if after.SourceRecordID != before.SourceRecordID {
		t.Fatalf("dismissed mention must preserve source row, before=%s after=%s", before.SourceRecordID, after.SourceRecordID)
	}
	if after.RawText != before.RawText || after.NormalizedText != before.NormalizedText {
		t.Fatalf("dismissed mention must preserve raw and normalized text, before=%#v after=%#v", before, after)
	}
}

func RequireDismissedMentionClearsResolution(t testing.TB, mention fixtures.EntityMentionFixture) {
	t.Helper()
	if mention.ResolutionStatus != "dismissed" {
		t.Fatalf("expected dismissed mention, got %#v", mention)
	}
	if mention.ResolvedRecordID != nil || mention.ResolvedByUserID != nil || mention.ResolvedAt != nil || mention.ResolutionMethod != nil {
		t.Fatalf("dismissed mention must clear active resolution metadata, got %#v", mention)
	}
}

func RequireRestoreToUnresolved(t testing.TB, mention fixtures.EntityMentionFixture) {
	t.Helper()
	if mention.ResolutionStatus != "unresolved" {
		t.Fatalf("restore must return mention to unresolved, got %#v", mention)
	}
	if mention.ResolvedRecordID != nil || mention.ResolvedByUserID != nil || mention.ResolvedAt != nil || mention.ResolutionMethod != nil {
		t.Fatalf("ordinary restore must not silently recover resolved metadata, got %#v", mention)
	}
}

func RequireSelectedMentionOnlyResolvedByCreateFromMention(t testing.TB, selected fixtures.EntityMentionFixture, sibling fixtures.EntityMentionFixture, createdRecordID uuid.UUID) {
	t.Helper()
	if selected.ResolutionStatus != "resolved" || selected.ResolvedRecordID == nil || *selected.ResolvedRecordID != createdRecordID {
		t.Fatalf("expected selected mention to resolve to created record %s, got %#v", createdRecordID, selected)
	}
	if sibling.ResolutionStatus != "unresolved" || sibling.ResolvedRecordID != nil {
		t.Fatalf("expected sibling mention to remain unresolved, got %#v", sibling)
	}
}

func RequireDistinctMentionProvenance(t testing.TB, mentions ...fixtures.EntityMentionFixture) {
	t.Helper()
	seenIDs := map[uuid.UUID]struct{}{}
	seenLocators := map[string]struct{}{}
	for _, mention := range mentions {
		if _, exists := seenIDs[mention.EntityMentionID]; exists {
			t.Fatalf("expected distinct entity_mention_id values, got duplicate %#v", mention)
		}
		seenIDs[mention.EntityMentionID] = struct{}{}
		if _, exists := seenLocators[mention.OriginLocator]; exists {
			t.Fatalf("expected distinct source provenance locators, got duplicate %#v", mention)
		}
		seenLocators[mention.OriginLocator] = struct{}{}
	}
}

func RequireExactMatchPrecedence(t testing.TB, got []string, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected exact-match precedence: got %v want %v", got, want)
	}
}

func RequireSuggestionBoundary(t testing.TB, identifiers []fixtures.PreservedIdentifierFixture) {
	t.Helper()
	for _, identifier := range identifiers {
		if identifier.IdentifierClass == "suggestion_only" && identifier.IdentifierType != "alias" {
			t.Fatalf("suggestion_only identifiers must remain non-authoritative alias inputs, got %#v", identifier)
		}
	}
}

func RequireIndicatorObservationSeparation(t testing.TB, observation fixtures.IndicatorObservationFixture, indicator fixtures.IndicatorFixture) {
	t.Helper()
	if observation.ObservationID == indicator.RecordID {
		t.Fatalf("source-bound observation must remain distinct from canonical indicator identity, got observation=%s indicator=%s", observation.ObservationID, indicator.RecordID)
	}
	if observation.SourceRecordID == uuid.Nil {
		t.Fatalf("indicator observation must remain source-bound, got %#v", observation)
	}
}

func RequireIndicatorLifecycleSeparate(t testing.TB, interval fixtures.IndicatorLifecycleFixture, indicator fixtures.IndicatorFixture) {
	t.Helper()
	if interval.IntervalID == indicator.RecordID {
		t.Fatalf("indicator lifecycle interval must remain distinct from canonical indicator record, got %#v %#v", interval, indicator)
	}
	if interval.IndicatorID != indicator.RecordID {
		t.Fatalf("indicator lifecycle interval must point at canonical indicator record, got %#v %#v", interval, indicator)
	}
}
