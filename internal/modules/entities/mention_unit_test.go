package entities_test

import (
	"context"
	"errors"
	"testing"
	"time"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	timeline "github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestCreateFromMention_Unit(t *testing.T) {
	t.Run("explicit host resolve links the selected mention to the canonical record", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-03-host-reuse")
		mentionStore := newEntityTestTimelineBundle(t, harness.DB).EntityMentionStore
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u403-host@example.test", "U403 Host", "U403HostEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-03-host-incident", "IR-U403-H", "Record relationships entity-storage host")

		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.RecordID)
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.SiblingRecordID)
		entitytest.SeedMention(t, harness.DB, actor.ID, entitytest.HostMentionID, timelinetest.RecordID, timelinetest.FieldHostRefs, "host", "WS-023", "unresolved", nil, nil)
		entitytest.SeedMention(t, harness.DB, actor.ID, entitytest.ResolvedHostMentionID, timelinetest.SiblingRecordID, timelinetest.FieldHostRefs, "host", "WS-023", "unresolved", nil, nil)

		tx, err := harness.DB.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		targetID := entitytest.CanonicalHostRecordID
		result, err := mentionStore.ResolveExistingFromMentionTx(context.Background(), tx, actor, timelinetest.RecordID, timelinetest.FieldHostRefs, entitytest.HostMentionID, &targetID, entitytest.BaseTime)
		if err != nil {
			t.Fatalf("resolve mention: %v", err)
		}
		if err := tx.Commit(context.Background()); err != nil {
			t.Fatalf("commit tx: %v", err)
		}

		if result.RecordID != entitytest.CanonicalHostRecordID || result.EntityType != "host" {
			t.Fatalf("expected selected mention to reuse canonical host, got %#v", result)
		}
		selected := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, selected, entitytest.MentionStatusResolved)
		if selected.ResolvedRecordID == nil || *selected.ResolvedRecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected selected mention to resolve to canonical host, got %#v", selected)
		}
		if selected.RawText != "WS-023" {
			t.Fatalf("expected selected raw text preservation, got %#v", selected)
		}

		sibling := entitytest.LookupMention(t, harness.DB, entitytest.ResolvedHostMentionID)
		entitytest.RequireMentionStatus(t, sibling, entitytest.MentionStatusUnresolved)
		if sibling.ResolvedRecordID != nil {
			t.Fatalf("expected sibling mention to remain unresolved, got %#v", sibling)
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM hosts WHERE incident_id = $1`, incident.ID); got != 1 {
			t.Fatalf("expected explicit resolve to avoid creating a second host, got %d rows", got)
		}
		var (
			projectedDisplayName string
			projectedHostname    string
			projectedState       string
		)
		if err := harness.DB.QueryRow(context.Background(), `
SELECT display_name, hostname, host_state
  FROM host_grid_projection
 WHERE record_id = $1
`, entitytest.CanonicalHostRecordID).Scan(&projectedDisplayName, &projectedHostname, &projectedState); err != nil {
			t.Fatalf("lookup host projection after explicit resolve: %v", err)
		}
		if projectedDisplayName != "WS-023" || projectedHostname != "WS-023" || projectedState != "canonical" {
			t.Fatalf("unexpected host projection after explicit resolve: display=%q hostname=%q state=%q", projectedDisplayName, projectedHostname, projectedState)
		}
	})

	t.Run("identity resolve without an explicit target is rejected without creating a stub", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-03-identity-create")
		mentionStore := newEntityTestTimelineBundle(t, harness.DB).EntityMentionStore
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u403-identity@example.test", "U403 Identity", "U403IdentityEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-03-identity-incident", "IR-U403-I", "Record relationships entity-storage identity")

		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.MixedRecordID)
		entitytest.SeedMention(t, harness.DB, actor.ID, entitytest.IdentityMentionID, timelinetest.MixedRecordID, timelinetest.FieldIdentityRefs, "identity", "alex.analyst@example.test", "unresolved", nil, nil)

		tx, err := harness.DB.BeginTx(context.Background(), pgxTxOptions())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		_, err = mentionStore.ResolveExistingFromMentionTx(context.Background(), tx, actor, timelinetest.MixedRecordID, timelinetest.FieldIdentityRefs, entitytest.IdentityMentionID, nil, entitytest.BaseTime)
		if !errors.Is(err, mentions.ErrInvalidMentionResolution) {
			t.Fatalf("expected missing target rejection, got %v", err)
		}

		mention := entitytest.LookupMention(t, harness.DB, entitytest.IdentityMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusUnresolved)
		if mention.ResolvedRecordID != nil {
			t.Fatalf("expected mention to remain unresolved, got %#v", mention)
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM identities WHERE incident_id = $1`, incident.ID); got != 0 {
			t.Fatalf("expected missing-target resolve to create no identity rows, got %d", got)
		}
	})
}

// entity-storage / REQ-02-039..REQ-02-041 / AC-188..AC-190, AC-224, AC-225.
func TestDismissRestoreMentionLifecycle_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "entity_linking-u-4-04")
	timelineBundle, projections := newEntityTestTimelineComposition(t, harness.DB)
	mentionStore := timelineBundle.EntityMentionStore
	timelineFacade := timelineBundle.Facade
	timelineProjectionStore := projections.RestoreProbeQuery()
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u404@example.test", "U404", "U404EntityLinkingPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-04-incident", "IR-U404", "Record relationships entity-storage")

	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
	normalizedHostToken, ok := fieldnorm.NormalizeMentionToken("WS-023")
	if !ok {
		t.Fatal("normalize mention token")
	}
	summary := "dismiss and restore relationship row"
	createRequest := timeline.CreateRequest{
		ClientTxnID:          "txn-entity_linking-u-4-04-row",
		ActivitySynopsisText: &summary,
		HostRefs: &timeline.CollectionActionPayload{
			Actions: []timeline.CollectionAction{
				{
					Op:             "add_resolved_ref",
					RawText:        "WS-023",
					NormalizedText: normalizedHostToken,
					ResolvedRecord: &entitytest.CanonicalHostRecordID,
				},
			},
		},
	}
	created, err := timelineFacade.CreateRow(context.Background(), timeline.CreateRowCommand{
		Actor:       actor,
		IncidentID:  incident.ID,
		Request:     createRequest,
		RequestHash: []byte("txn-entity_linking-u-4-04-row"),
		RequestID:   "req-entity_linking-u-4-04-row",
		Now:         entitytest.BaseTime,
	})
	if err != nil {
		t.Fatalf("create resolved relationship row: %v", err)
	}
	initialRows, err := timelineProjectionStore.QueryRows(context.Background(), incident.ID, timeline.TimelineViewSchemaID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query initial timeline rows: %v", err)
	}
	initialRow := workbookscenariotest.FindRow(t, initialRows, created.RecordID.String())
	initialItem := workbookscenariotest.RequireSingleCollectionItem(t, initialRow, timelinetest.FieldHostRefs)
	mentionID := entitytest.MentionIDFromItemRef(t, initialItem["item_ref"].(string))

	initialMention := entitytest.LookupMention(t, harness.DB, mentionID)
	dismissRequest := mentions.MentionActionRequest{
		BaseMentionRowVersion: initialMention.RowVersion,
		ClientTxnID:           "txn-entity_linking-u-4-04-dismiss",
		Action:                "dismiss_item",
	}
	dismissResult, err := mentionStore.ApplyMentionAction(context.Background(), actor, mentionID, dismissRequest, mentions.MentionActionRequestHash(dismissRequest), "req-entity_linking-u-4-04-dismiss", entitytest.BaseTime)
	if err != nil {
		t.Fatalf("dismiss mention: %v", err)
	}

	dismissed := entitytest.LookupMention(t, harness.DB, mentionID)
	entitytest.RequireMentionStatus(t, dismissed, entitytest.MentionStatusDismissed)
	if dismissed.ResolvedRecordID != nil || dismissed.ResolvedAt != nil || dismissed.ResolutionMethod != nil {
		t.Fatalf("expected dismissed mention to clear resolution metadata, got %#v", dismissed)
	}
	if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incident.ID, created.RecordID, entitytest.CanonicalHostRecordID); got != 0 {
		t.Fatalf("expected dismiss to remove active derived link, got %d rows", got)
	}
	dismissedRows, err := timelineProjectionStore.QueryRows(context.Background(), incident.ID, timeline.TimelineViewSchemaID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query timeline rows after dismiss: %v", err)
	}
	dismissedRow := workbookscenariotest.FindRow(t, dismissedRows, created.RecordID.String())
	if got := workbookscenariotest.CollectionItems(t, dismissedRow, timelinetest.FieldHostRefs); len(got) != 0 {
		t.Fatalf("dismissed mention must be excluded from current relationship-cell values, got %#v", got)
	}
	requireTimelineMutationAfterSnapshotMatchesRecord(t, harness.DB, dismissResult.ChangeSetID, dismissedRow)

	restoreRequest := mentions.MentionActionRequest{
		BaseMentionRowVersion: dismissed.RowVersion,
		ClientTxnID:           "txn-entity_linking-u-4-04-restore",
		Action:                "revert_to_unresolved",
	}
	restoreResult, err := mentionStore.ApplyMentionAction(context.Background(), actor, mentionID, restoreRequest, mentions.MentionActionRequestHash(restoreRequest), "req-entity_linking-u-4-04-restore", entitytest.BaseTime.Add(time.Minute))
	if err != nil {
		t.Fatalf("restore mention: %v", err)
	}

	restored := entitytest.LookupMention(t, harness.DB, mentionID)
	entitytest.RequireMentionStatus(t, restored, entitytest.MentionStatusUnresolved)
	if restored.RawText != "WS-023" || restored.ResolvedRecordID != nil || restored.ResolutionMethod != nil {
		t.Fatalf("expected durable restore-to-unresolved semantics, got %#v", restored)
	}
	restoredRows, err := timelineProjectionStore.QueryRows(context.Background(), incident.ID, timeline.TimelineViewSchemaID, mustDefaultQueryMeta(t, timeline.TimelineViewSchemaID))
	if err != nil {
		t.Fatalf("query timeline rows after restore: %v", err)
	}
	restoredRow := workbookscenariotest.FindRow(t, restoredRows, created.RecordID.String())
	restoredItem := workbookscenariotest.RequireSingleCollectionItem(t, restoredRow, timelinetest.FieldHostRefs)
	if restoredItem["item_kind"] != "unresolved_mention" || restoredItem["raw_text"] != "WS-023" {
		t.Fatalf("restore must surface the unresolved mention in current-state reads, got %#v", restoredItem)
	}
	requireTimelineMutationAfterSnapshotMatchesRecord(t, harness.DB, restoreResult.ChangeSetID, restoredRow)
	if _, ok := restoredItem["resolved_record_id"]; ok {
		t.Fatalf("restore must not silently relink the historical target, got %#v", restoredItem)
	}
}

// entity-storage / REQ-02-059..REQ-02-063 / AC-021, AC-022.
