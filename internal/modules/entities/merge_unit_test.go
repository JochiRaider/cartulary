package entities_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	assessmenttest "github.com/JochiRaider/cartulary/internal/modules/assessments/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestExplicitEntityMerge_Unit(t *testing.T) {
	t.Run("host merge preserves raw mentions, loser lineage, and survivor reuse", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-06-host")
		store := newEntityTestStore(t, harness.DB)
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u406@example.test", "U406", "U406EntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-incident", "IR-U406", "Record relationships entity-storage")

		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, entitytest.DuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "ws-023.corp.example.test", "")
		entitytest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, entitytest.DuplicateHostRecordID, "host", "Workstation 23")
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.RecordID)
		outgoingTargetID := uuid.New()
		outgoingLinkID := uuid.New()
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, outgoingTargetID)
		entitytest.SeedResolvedMention(t, harness.DB, actor.ID, entitytest.HostMentionID, timelinetest.RecordID, entitytest.DuplicateHostRecordID, timelinetest.FieldHostRefs, "host", "WS-023")
		linktest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, linktest.DuplicateLinkID, timelinetest.RecordID, entitytest.DuplicateHostRecordID, "observed_on_host", "manual", nil)
		linktest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, outgoingLinkID, entitytest.DuplicateHostRecordID, outgoingTargetID, "references_record", "manual", nil)
		linktest.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, linktest.TagIDSurvivor, entitytest.CanonicalHostRecordID, "critical-host")
		linktest.SeedRecordTag(t, harness.DB, incident.ID, actor.ID, linktest.TagIDLoser, entitytest.DuplicateHostRecordID, "critical-host")
		mergeFixtureCreatedAt := entitytest.BaseTime.Add(-time.Minute)
		if _, err := harness.DB.Exec(context.Background(), `
UPDATE record_links
   SET decided_at = $1,
       created_at = $1
 WHERE record_link_id IN ($2, $3)
`, mergeFixtureCreatedAt, linktest.DuplicateLinkID, outgoingLinkID); err != nil {
			t.Fatalf("normalize merge link fixture timestamps: %v", err)
		}
		if _, err := harness.DB.Exec(context.Background(), `
UPDATE record_tags
   SET created_at = $1,
       updated_at = $1
 WHERE record_tag_id IN ($2, $3)
`, mergeFixtureCreatedAt, linktest.TagIDSurvivor, linktest.TagIDLoser); err != nil {
			t.Fatalf("normalize merge tag fixture timestamps: %v", err)
		}
		assessmenttest.SeedAssessment(t, harness.DB, incident.ID, actor.ID, assessmenttest.HostAssessmentID, entitytest.DuplicateHostRecordID, "host", "confirmed")
		beforeMention := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)

		result, err := newEntityTestTimelineBundle(t, harness.DB).EntityMergeStore.MergeEntity(context.Background(), actor, entitytest.CanonicalHostRecordID, merge.MergeRequest{
			LoserRecordID:          entitytest.DuplicateHostRecordID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-merge",
		}, []byte("txn-entity_linking-u-4-06-merge"), "req-merge", entitytest.BaseTime)
		if err != nil {
			t.Fatalf("merge entity: %v", err)
		}
		if result.SurvivorRecordID != entitytest.CanonicalHostRecordID || result.LoserRecordID != entitytest.DuplicateHostRecordID {
			t.Fatalf("unexpected merge result: %#v", result)
		}
		if result.MergeSummary.RepointedLinkCount != 2 {
			t.Fatalf("expected merge to repoint incoming and outgoing links, got summary %#v", result.MergeSummary)
		}

		mention := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected merge to repoint mention resolution to survivor, got %#v", mention)
		}
		entitytest.RequireRawTextPreserved(t, beforeMention.RawText, mention.RawText)
		link := linktest.LookupActiveLink(t, harness.DB, incident.ID, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host")
		linktest.RequireActiveLink(t, link, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host", "manual", nil)
		outgoing := linktest.LookupActiveLink(t, harness.DB, incident.ID, entitytest.CanonicalHostRecordID, outgoingTargetID, "references_record")
		linktest.RequireActiveLink(t, outgoing, entitytest.CanonicalHostRecordID, outgoingTargetID, "references_record", "manual", nil)
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, outgoingLinkID); got != 0 {
			t.Fatalf("expected loser-sourced active link to disappear, got %d rows", got)
		}
		if got := assessmenttest.LookupAssessmentSubject(t, harness.DB, assessmenttest.HostAssessmentID); got != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected assessment repoint to survivor, got %s", got)
		}
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'critical-host'
   AND deleted_at IS NULL
`, incident.ID, entitytest.CanonicalHostRecordID); got != 1 {
			t.Fatalf("expected one active deduped survivor tag, got %d", got)
		}
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, linktest.DuplicateLinkID); got != 0 {
			t.Fatalf("expected loser-targeted active link to disappear, got %d rows", got)
		}
		state, mergedInto, rowVersion, _ := entitytest.LookupHostState(t, harness.DB, entitytest.DuplicateHostRecordID)
		if state != "merged" || mergedInto == nil || *mergedInto != entitytest.CanonicalHostRecordID || rowVersion != 2 {
			t.Fatalf("expected loser host lineage state after merge, got state=%s merged_into=%v row_version=%d", state, mergedInto, rowVersion)
		}

		reuse, err := store.CreateHostRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-reuse",
			Values: map[string]string{
				"host.fqdn": "ws-023.corp.example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-reuse"), "req-reuse", entitytest.BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge exact-match reuse: %v", err)
		}
		if reuse.RecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected exact-match reuse to carry forward to survivor, got %#v", reuse)
		}
	})

	t.Run("identity merge preserves raw mentions, loser lineage, and survivor reuse", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-06-identity")
		store := newEntityTestStore(t, harness.DB)
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u406-identity@example.test", "U406 Identity", "U406IdentityEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-identity-incident", "IR-U406-I", "Record relationships entity-storage identity")

		entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, entitytest.CanonicalIdentityRecordID, "Alex Analyst", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, entitytest.DuplicateIdentityRecordID, "Alex Duplicate", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		entitytest.SeedEntityAlias(t, harness.DB, incident.ID, actor.ID, entitytest.DuplicateIdentityRecordID, "identity", "Case Owner")
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.RecordID)
		entitytest.SeedResolvedMention(t, harness.DB, actor.ID, entitytest.IdentityMentionID, timelinetest.RecordID, entitytest.DuplicateIdentityRecordID, timelinetest.FieldIdentityRefs, "identity", "Case Owner")
		linktest.SeedRecordLink(t, harness.DB, incident.ID, actor.ID, linktest.DuplicateLinkID, timelinetest.RecordID, entitytest.DuplicateIdentityRecordID, "observed_as_identity", "manual", nil)
		mergeFixtureCreatedAt := entitytest.BaseTime.Add(-time.Minute)
		if _, err := harness.DB.Exec(context.Background(), `
UPDATE record_links
   SET decided_at = $1,
       created_at = $1
 WHERE record_link_id = $2
`, mergeFixtureCreatedAt, linktest.DuplicateLinkID); err != nil {
			t.Fatalf("normalize merge link fixture timestamps: %v", err)
		}
		assessmenttest.SeedAssessment(t, harness.DB, incident.ID, actor.ID, assessmenttest.IdentityAssessmentID, entitytest.DuplicateIdentityRecordID, "identity", "confirmed")
		beforeMention := entitytest.LookupMention(t, harness.DB, entitytest.IdentityMentionID)

		result, err := newEntityTestTimelineBundle(t, harness.DB).EntityMergeStore.MergeEntity(context.Background(), actor, entitytest.CanonicalIdentityRecordID, merge.MergeRequest{
			LoserRecordID:          entitytest.DuplicateIdentityRecordID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-identity-merge",
		}, []byte("txn-entity_linking-u-4-06-identity-merge"), "req-identity-merge", entitytest.BaseTime)
		if err != nil {
			t.Fatalf("merge identity: %v", err)
		}
		if result.SurvivorRecordID != entitytest.CanonicalIdentityRecordID || result.LoserRecordID != entitytest.DuplicateIdentityRecordID {
			t.Fatalf("unexpected identity merge result: %#v", result)
		}

		mention := entitytest.LookupMention(t, harness.DB, entitytest.IdentityMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != entitytest.CanonicalIdentityRecordID {
			t.Fatalf("expected identity merge to repoint mention resolution to survivor, got %#v", mention)
		}
		entitytest.RequireRawTextPreserved(t, beforeMention.RawText, mention.RawText)
		link := linktest.LookupActiveLink(t, harness.DB, incident.ID, timelinetest.RecordID, entitytest.CanonicalIdentityRecordID, "observed_as_identity")
		linktest.RequireActiveLink(t, link, timelinetest.RecordID, entitytest.CanonicalIdentityRecordID, "observed_as_identity", "manual", nil)
		if got := assessmenttest.LookupAssessmentSubject(t, harness.DB, assessmenttest.IdentityAssessmentID); got != entitytest.CanonicalIdentityRecordID {
			t.Fatalf("expected identity assessment repoint to survivor, got %s", got)
		}

		var (
			state         string
			mergedIntoRaw sql.NullString
			rowVersion    int64
		)
		if err := harness.DB.QueryRow(context.Background(), `
SELECT identity_state, merged_into_record_id::text, row_version
  FROM identities
 WHERE record_id = $1
`, entitytest.DuplicateIdentityRecordID).Scan(&state, &mergedIntoRaw, &rowVersion); err != nil {
			t.Fatalf("lookup loser identity state: %v", err)
		}
		if state != "merged" || !mergedIntoRaw.Valid || mergedIntoRaw.String != entitytest.CanonicalIdentityRecordID.String() || rowVersion != 2 {
			t.Fatalf("expected loser identity lineage after merge, got state=%s merged_into=%v row_version=%d", state, mergedIntoRaw, rowVersion)
		}

		reuse, err := store.CreateIdentityRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-identity-reuse",
			Values: map[string]string{
				"identity.email": "alex.analyst@example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-identity-reuse"), "req-identity-reuse", entitytest.BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge identity exact-match reuse: %v", err)
		}
		if reuse.RecordID != entitytest.CanonicalIdentityRecordID {
			t.Fatalf("expected identity exact-match reuse to carry forward to survivor, got %#v", reuse)
		}
	})

	t.Run("host merge exposes carried secondary reusable identifiers", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-06-host-reusable-row")
		store := newEntityTestStore(t, harness.DB)
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u406-host-reusable@example.test", "U406 Host Reusable", "U406HostReusableEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-host-reusable-incident", "IR-U406-HR", "Record relationships entity-storage host reusable rows")
		survivorID := uuid.New()
		loserID := uuid.New()

		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, survivorID, "WS-023", "WS-023", "ws-023.current.example.test", "")
		entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, loserID, "Legacy WS-023", "LEGACY-WS-023", "legacy-ws-023.example.test", "")

		if _, err := newEntityTestTimelineBundle(t, harness.DB).EntityMergeStore.MergeEntity(context.Background(), actor, survivorID, merge.MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-host-reusable-merge",
		}, []byte("txn-entity_linking-u-4-06-host-reusable-merge"), "req-host-reusable-merge", entitytest.BaseTime); err != nil {
			t.Fatalf("merge host reusable identifiers: %v", err)
		}

		page, err := store.QueryHostRowsPage(context.Background(), incident.ID, mustDefaultQueryMeta(t, entitycontract.HostsViewSchemaID), querypage.Window{Limit: 100})
		if err != nil {
			t.Fatalf("query host rows after reusable merge: %v", err)
		}
		survivorRow := requireEntityRow(t, page.Rows, survivorID)
		requireReusableIdentifierItem(t, survivorRow, "host.reusable_identifiers", "fqdn", "legacy-ws-023.example.test")
		requireNoReusableIdentifierItem(t, survivorRow, "host.reusable_identifiers", "fqdn", "ws-023.current.example.test")

		reuse, err := store.CreateHostRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-host-reusable-create",
			Values: map[string]string{
				"host.fqdn": "legacy-ws-023.example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-host-reusable-create"), "req-host-reusable-create", entitytest.BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge host reusable exact match: %v", err)
		}
		if reuse.RecordID != survivorID {
			t.Fatalf("expected reusable identifier to match survivor, got %#v", reuse)
		}
		reuseRow := reuse.Payload["row"].(map[string]any)
		requireReusableIdentifierItem(t, reuseRow, "host.reusable_identifiers", "fqdn", "legacy-ws-023.example.test")
	})

	t.Run("identity merge exposes carried secondary reusable identifiers", func(t *testing.T) {
		harness := appsupport.StartStore(t, "entity_linking-u-4-06-identity-reusable-row")
		store := newEntityTestStore(t, harness.DB)
		actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u406-identity-reusable@example.test", "U406 Identity Reusable", "U406IdentityReusableEntityLinkingPass1!", false, false, true)
		incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-06-identity-reusable-incident", "IR-U406-IR", "Record relationships entity-storage identity reusable rows")
		survivorID := uuid.New()
		loserID := uuid.New()

		entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, survivorID, "Alex Survivor", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, loserID, "Alex Analyst Legacy", "alex.legacy@example.test", "alex.legacy@example.test", "ALEXLEGACY")

		if _, err := newEntityTestTimelineBundle(t, harness.DB).EntityMergeStore.MergeEntity(context.Background(), actor, survivorID, merge.MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-entity_linking-u-4-06-identity-reusable-merge",
		}, []byte("txn-entity_linking-u-4-06-identity-reusable-merge"), "req-identity-reusable-merge", entitytest.BaseTime); err != nil {
			t.Fatalf("merge identity reusable identifiers: %v", err)
		}

		page, err := store.QueryIdentityRowsPage(context.Background(), incident.ID, mustDefaultQueryMeta(t, entitycontract.IdentitiesViewSchemaID), querypage.Window{Limit: 100})
		if err != nil {
			t.Fatalf("query identity rows after reusable merge: %v", err)
		}
		survivorRow := requireEntityRow(t, page.Rows, survivorID)
		requireReusableIdentifierItem(t, survivorRow, "identity.reusable_identifiers", "email", "alex.legacy@example.test")
		requireNoReusableIdentifierItem(t, survivorRow, "identity.reusable_identifiers", "email", "alex.survivor@example.test")

		reuse, err := store.CreateIdentityRow(context.Background(), actor, incident.ID, hostidentity.CreateRequest{
			ClientTxnID: "txn-entity_linking-u-4-06-identity-reusable-create",
			Values: map[string]string{
				"identity.email": "alex.legacy@example.test",
			},
		}, []byte("txn-entity_linking-u-4-06-identity-reusable-create"), "req-identity-reusable-create", entitytest.BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("post-merge identity reusable exact match: %v", err)
		}
		if reuse.RecordID != survivorID {
			t.Fatalf("expected reusable identifier to match survivor, got %#v", reuse)
		}
		reuseRow := reuse.Payload["row"].(map[string]any)
		requireReusableIdentifierItem(t, reuseRow, "identity.reusable_identifiers", "email", "alex.legacy@example.test")
	})
}
