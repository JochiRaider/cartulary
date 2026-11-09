package links_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func TestMergeDeduplicatesOnlyFullFieldAwareTupleCharacterization(t *testing.T) {
	harness := appsupport.StartStore(t, "links-field-aware-merge-characterization")
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "links-field-merge@example.test", DisplayName: "Links Field Merge", DeploymentAdmin: true})
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
    ('00000000-0000-0000-0000-000000000103', $1, $3, $4, 'references_record', 'fixture.links.field_b', 'manual', $5, $5, $6, $6),
    ('00000000-0000-0000-0000-000000000104', $1, $2, $4, 'references_record', NULL, 'manual', $5, $5, $6, $6),
    ('00000000-0000-0000-0000-000000000105', $1, $3, $4, 'references_record', NULL, 'manual', $5, $5, $6, $6),
    ('00000000-0000-0000-0000-000000000106', $1, $3, $2, 'references_record', NULL, 'manual', $5, $5, $6, $6)
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
	if result.DedupedCount != 3 || result.RepointedCount != 1 || len(result.Mutations()) != 5 {
		t.Fatalf("merge result = deduped:%d repointed:%d mutations:%d, want 3/1/5", result.DedupedCount, result.RepointedCount, len(result.Mutations()))
	}
	var activeA, activeB, activeNull, loserActive int
	if err := tx.QueryRow(ctx, `
SELECT
    count(*) FILTER (WHERE src_record_id = $2 AND field_key = 'fixture.links.field_a'),
    count(*) FILTER (WHERE src_record_id = $2 AND field_key = 'fixture.links.field_b'),
    count(*) FILTER (WHERE src_record_id = $2 AND field_key IS NULL),
    count(*) FILTER (WHERE src_record_id = $3)
  FROM record_links
 WHERE incident_id = $1 AND dst_record_id = $4
   AND link_type = 'references_record' AND deleted_at IS NULL
`, incident.ID, survivor, loser, target).Scan(&activeA, &activeB, &activeNull, &loserActive); err != nil {
		t.Fatalf("query merged field-aware links: %v", err)
	}
	if activeA != 1 || activeB != 1 || activeNull != 1 || loserActive != 0 {
		t.Fatalf("post-merge active counts = fieldA:%d fieldB:%d loser:%d, want 1/1/0", activeA, activeB, loserActive)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	creator := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "links-original-creator@example.test", DisplayName: "Original creator", DeploymentAdmin: true})
	for _, provenance := range []string{"manual", "auto_match", "import", "system", "rollback"} {
		confidences := []any{nil}
		if provenance == "auto_match" {
			confidences = []any{100}
		}
		if provenance == "import" || provenance == "system" || provenance == "rollback" {
			confidences = []any{nil, 0, 37, 100}
		}
		for _, confidence := range confidences {
			t.Run(fmt.Sprintf("retained %s confidence %v", provenance, confidence), func(t *testing.T) {
				port, err := links.NewIncidentBundleSourcePort()
				if err != nil {
					t.Fatal(err)
				}
				linkID := uuid.New()
				linkType := "references_record"
				if provenance == "auto_match" {
					linkType = "observed_on_host"
				}
				row := map[string]any{
					"record_link_id": linkID.String(), "incident_id": incident.ID.String(),
					"src_record_id": target.String(), "dst_record_id": loser.String(),
					"link_type": linkType, "field_key": nil, "provenance": provenance, "confidence": confidence,
					"owner_user_id": actor.ID.String(), "created_by_user_id": creator.ID.String(),
					"decided_at": now.Add(time.Hour).Format("2006-01-02T15:04:05.999999+00:00"), "created_at": now.Format("2006-01-02T15:04:05.999999+00:00"),
					"deleted_at": nil, "deleted_by_user_id": nil,
				}
				payload, err := json.Marshal(row)
				if err != nil {
					t.Fatal(err)
				}
				input := sourceport.ImportContext{IncidentID: incident.ID, ActorUserID: actor.ID, BundleVersion: 8,
					OperationID: "merge-" + provenance, Attributions: &linksPortableAttributionRecorder{}}
				prepared, err := port.PrepareImport(ctx, portabilitysupport.MapBundle{
					"data/record_links.ndjson": append(payload, '\n'), "data/tags.ndjson": {}, "data/record_tags.ndjson": {},
				}, input)
				if err != nil {
					t.Fatal(err)
				}
				tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
					t.Fatal(err)
				}
				if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
					t.Fatal(err)
				}
				// Import rebinds deployment actors. Exercise a valid later ownership change
				// so creation attribution differs from the current owner.
				if _, err := tx.Exec(ctx, `UPDATE record_links SET owner_user_id=$2 WHERE record_link_id=$1`, linkID, creator.ID); err != nil {
					t.Fatal(err)
				}
				result, err := links.NewStore().RepointMergedLinksTx(ctx, tx, links.RepointMergedLinksCommand{
					IncidentID: incident.ID, SurvivorRecordID: survivor, LoserRecordID: loser,
					ActorUserID: actor.ID, Now: now.Add(2 * time.Hour),
				})
				if err != nil {
					t.Fatalf("merge admitted retained %s: %v", provenance, err)
				}
				mutations := result.Mutations()
				if result.RepointedCount != 1 || result.DedupedCount != 0 || len(mutations) != 2 {
					t.Fatalf("unexpected merge result: %#v", result)
				}
				before, after := mutationMap(t, mutations[0].BeforeValue()), mutationMap(t, mutations[1].AfterValue())
				for _, key := range []string{"field_key", "provenance", "confidence", "owner_user_id", "created_by_user_id", "decided_at", "created_at"} {
					if !reflect.DeepEqual(before[key], after[key]) {
						t.Errorf("merge changed retained %s: %v -> %v", key, before[key], after[key])
					}
				}
				if after["dst_record_id"] != survivor.String() {
					t.Fatal("merge did not repoint endpoint")
				}
				var persisted []byte
				if err := tx.QueryRow(ctx, `SELECT to_jsonb(l) FROM record_links l WHERE record_link_id=$1`, after["record_link_id"]).Scan(&persisted); err != nil {
					t.Fatal(err)
				}
				var stored map[string]any
				if err := json.Unmarshal(persisted, &stored); err != nil {
					t.Fatal(err)
				}
				for _, key := range []string{"owner_user_id", "created_by_user_id", "provenance", "confidence", "field_key"} {
					want, _ := json.Marshal(after[key])
					got, _ := json.Marshal(stored[key])
					if string(want) != string(got) {
						t.Fatalf("persisted %s differs: %s != %s", key, got, want)
					}
				}
				for _, key := range []string{"decided_at", "created_at"} {
					got, err := time.Parse(time.RFC3339Nano, stored[key].(string))
					want, wantErr := time.Parse(time.RFC3339Nano, after[key].(string))
					if err != nil || wantErr != nil || !got.Equal(want) {
						t.Fatalf("persisted %s differs", key)
					}
				}
				provider := links.RevisionProviderContribution().NonRowTargets[0]
				for i := len(mutations) - 1; i >= 0; i-- {
					mutation := mutations[i]
					result, err := provider.RollbackProvider.ApplyInverseTx(ctx, tx, rollbackcontract.ApplyInverseRequest{
						Target:      rollbackcontract.NonRowTarget{IncidentID: incident.ID, TargetKind: mutation.TargetKind(), TargetID: mutation.TargetID(), OperationKind: mutation.OperationKind(), BeforeValue: mutationMap(t, mutation.BeforeValue()), AfterValue: mutationMap(t, mutation.AfterValue())},
						ActorUserID: actor.ID, Now: now.Add(3 * time.Hour),
					})
					if err != nil {
						t.Fatal(err)
					}
					if i == 0 && !reflect.DeepEqual(before, result.AfterValue) {
						t.Fatalf("inverse lost original source: %#v", result.AfterValue)
					}
				}
			})
		}
	}
	for _, invalid := range []struct {
		provenance, linkType string
		confidence           any
	}{
		{"manual", "references_record", 37}, {"auto_match", "observed_on_host", nil},
		{"auto_match", "observed_on_host", 99}, {"auto_match", "references_record", 100},
	} {
		t.Run(fmt.Sprintf("reject %s %s %v", invalid.provenance, invalid.linkType, invalid.confidence), func(t *testing.T) {
			tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `INSERT INTO record_links(incident_id,src_record_id,dst_record_id,link_type,provenance,confidence,owner_user_id,created_by_user_id,decided_at,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$7,$8,$8)`, incident.ID, target, loser, invalid.linkType, invalid.provenance, invalid.confidence, actor.ID, now); err != nil {
				t.Fatal(err)
			}
			result, err := links.NewStore().RepointMergedLinksTx(ctx, tx, links.RepointMergedLinksCommand{IncidentID: incident.ID, SurvivorRecordID: survivor, LoserRecordID: loser, ActorUserID: actor.ID, Now: now.Add(time.Hour)})
			if err == nil || len(result.Mutations()) != 0 {
				t.Fatal("invalid retained tuple was transformed")
			}
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE incident_id=$1 AND dst_record_id=$2 AND deleted_at IS NULL`, incident.ID, loser).Scan(&count); err != nil || count != 1 {
				t.Fatal("invalid input changed source", err)
			}
		})
	}
}
