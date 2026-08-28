package links_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestMergeDeduplicatesOnlyFullFieldAwareTupleCharacterization(t *testing.T) {
	harness := appsupport.StartStore(t, "links-field-aware-merge-characterization")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "links-field-merge@example.test", "Links Field Merge", "LinksFieldMergePass1!", false, true, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-links-field-merge-incident", "IR-LINK-MERGE", "Links field merge")
	ctx := context.Background()
	now := time.Date(2026, time.August, 21, 18, 41, 0, 0, time.UTC)
	survivor := uuid.New()
	loser := uuid.New()
	target := uuid.New()
	for _, recordID := range []uuid.UUID{survivor, loser, target} {
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
	}
	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin merge characterization: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
INSERT INTO record_links (
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES
    ('00000000-0000-0000-0000-000000000101', $1, $2, $4, 'references_record', 'fixture.links.field_a', 'manual', $5, $5, $6, $6),
    ('00000000-0000-0000-0000-000000000102', $1, $3, $4, 'references_record', 'fixture.links.field_a', 'manual', $5, $5, $6, $6),
    ('00000000-0000-0000-0000-000000000103', $1, $3, $4, 'references_record', 'fixture.links.field_b', 'manual', $5, $5, $6, $6)
`, incident.ID, survivor, loser, target, actor.ID, now); err != nil {
		t.Fatalf("seed merge links: %v", err)
	}
	result, err := links.NewStore().RepointMergedLinksTx(ctx, tx, links.RepointMergedLinksCommand{
		IncidentID: incident.ID, SurvivorRecordID: survivor, LoserRecordID: loser,
		ActorUserID: actor.ID, Now: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("repoint merged links: %v", err)
	}
	if result.DedupedCount != 1 || result.RepointedCount != 1 || len(result.Mutations()) != 3 {
		t.Fatalf("merge result = deduped:%d repointed:%d mutations:%d, want 1/1/3", result.DedupedCount, result.RepointedCount, len(result.Mutations()))
	}
	var activeA, activeB, loserActive int
	if err := tx.QueryRow(ctx, `
SELECT
    count(*) FILTER (WHERE src_record_id = $2 AND field_key = 'fixture.links.field_a'),
    count(*) FILTER (WHERE src_record_id = $2 AND field_key = 'fixture.links.field_b'),
    count(*) FILTER (WHERE src_record_id = $3)
  FROM record_links
 WHERE incident_id = $1 AND dst_record_id = $4
   AND link_type = 'references_record' AND deleted_at IS NULL
`, incident.ID, survivor, loser, target).Scan(&activeA, &activeB, &loserActive); err != nil {
		t.Fatalf("query merged field-aware links: %v", err)
	}
	if activeA != 1 || activeB != 1 || loserActive != 0 {
		t.Fatalf("post-merge active counts = fieldA:%d fieldB:%d loser:%d, want 1/1/0", activeA, activeB, loserActive)
	}
}
