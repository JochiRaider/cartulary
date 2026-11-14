package artifacts_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/google/uuid"
)

func TestArtifactCollectionMutationContractMatrix(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-collection-contract")
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "artifact-collection-owner@example.test", DisplayName: "Artifact Collection Owner"})
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-collection-incident",
		"IR-ARTIFACTS-COLLECTION",
		"Artifact collection contract",
	)
	now := time.Date(2026, 7, 30, 15, 0, 0, 0, time.UTC)
	facade := mustArtifactMutationFacade(
		t,
		harness.DB,
		conflicttest.NewCodec("artifacts-collection"),
	)

	targetRecordID := seedArtifactContractRecord(t, harness, incident.ID, actor.ID, "evidence", now)
	partyRecordID := seedArtifactContractRecord(t, harness, incident.ID, actor.ID, "party", now)
	if _, err := harness.DB.Exec(ctx, `
INSERT INTO parties (record_id, incident_id, display_name, party_kind)
VALUES ($1, $2, 'Synthetic responder', 'person')
`, partyRecordID, incident.ID); err != nil {
		t.Fatalf("seed party subtype: %v", err)
	}

	cases := []struct {
		name         string
		viewSchemaID string
		values       map[string]any
		collections  map[string]any
		assert       func(artifacts.MutationResult)
	}{
		{
			name:         "record_tag",
			viewSchemaID: artifacts.NotesViewSchemaID,
			values:       artifactTextValues("note.title", "Tagged note"),
			collections: map[string]any{
				"note.tags": artifactCollectionPayload(
					map[string]any{"op": "add_tag", "tag_name": "Synthetic"},
					map[string]any{"op": "add_tag", "tag_name": "synthetic"},
				),
			},
			assert: func(result artifacts.MutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM record_tags WHERE record_id = $1 AND normalized_tag_name = 'synthetic' AND deleted_at IS NULL`, result.RecordID, 1)
			},
		},
		{
			name:         "record_ref",
			viewSchemaID: artifacts.FindingsViewSchemaID,
			values:       artifactTextValues("finding.statement", "Synthetic relationship"),
			collections: map[string]any{
				"finding.supporting_refs": artifactCollectionPayload(
					map[string]any{"op": "add_record_ref", "linked_record_id": targetRecordID.String()},
					map[string]any{"op": "add_record_ref", "linked_record_id": targetRecordID.String()},
				),
			},
			assert: func(result artifacts.MutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM record_links WHERE src_record_id = $1 AND dst_record_id = $2 AND link_type = 'supported_by' AND deleted_at IS NULL`, result.RecordID, targetRecordID, 1)
			},
		},
		{
			name:         "party_ref",
			viewSchemaID: artifacts.CommLogViewSchemaID,
			values: artifactTextValues(
				"comm_log.comm_type", "briefing",
				"comm_log.audience", "responders",
				"comm_log.channel_or_meeting", "bridge",
				"comm_log.summary", "Synthetic party reference",
			),
			collections: map[string]any{
				"comm_log.audience_party_ids": artifactCollectionPayload(
					map[string]any{"op": "add_party_ref", "party_id": partyRecordID.String()},
					map[string]any{"op": "add_party_ref", "party_id": partyRecordID.String()},
				),
			},
			assert: func(result artifacts.MutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM record_links WHERE src_record_id = $1 AND dst_record_id = $2 AND field_key = 'comm_log.audience_party_ids' AND deleted_at IS NULL`, result.RecordID, partyRecordID, 1)
			},
		},
		{
			name:         "risk_ref",
			viewSchemaID: artifacts.HandoffViewSchemaID,
			values: map[string]any{
				"handoff.incoming_owner_user_id": actor.ID.String(),
				"handoff.current_state_summary":  "Synthetic handoff",
			},
			collections: map[string]any{
				"handoff.open_risk_refs": artifactCollectionPayload(
					map[string]any{"op": "add_risk_ref", "risk_ref_text": " Synthetic risk\u200d "},
					map[string]any{"op": "add_risk_ref", "risk_ref_text": "Synthetic risk\u200d"},
					map[string]any{"op": "add_risk_ref", "risk_ref_text": "synthetic risk\u200d"},
				),
			},
			assert: func(result artifacts.MutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM handoff_risk_refs WHERE handoff_record_id = $1 AND deleted_at IS NULL`, result.RecordID, 2)
				for _, text := range []string{"Synthetic risk\u200d", "synthetic risk\u200d"} {
					requireCount(t, harness, `SELECT count(*) FROM handoff_risk_refs WHERE handoff_record_id = $1 AND risk_ref_text = $2 AND deleted_at IS NULL`, result.RecordID, text, 1)
				}
			},
		},
	}

	for index, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			clientTxnID := fmt.Sprintf("txn-artifacts-collection-%02d", index)
			command := artifacts.CreateCommand{
				ActorUserID: actor.ID, IncidentID: incident.ID,
				Admission: mustArtifactCreateAdmission(t, tc.viewSchemaID, clientTxnID, tc.values, tc.collections),
				RequestID: "req-" + clientTxnID,
				Now:       now.Add(time.Duration(index) * time.Minute),
			}
			result, err := facade.Create(ctx, command)
			if err != nil {
				t.Fatalf("create collection family %s: %v", tc.name, err)
			}
			tc.assert(result)
			var createdRiskIDs []uuid.UUID
			if tc.name == "risk_ref" {
				if err := harness.DB.QueryRow(ctx, `SELECT array_agg(risk_ref_id ORDER BY risk_ref_id) FROM handoff_risk_refs WHERE handoff_record_id=$1 AND deleted_at IS NULL`, result.RecordID).Scan(&createdRiskIDs); err != nil {
					t.Fatal(err)
				}
			}
			replayed, err := facade.Create(ctx, command)
			if err != nil || replayed.Outcome != artifacts.MutationOutcomeReplayed || replayed.RecordID != result.RecordID {
				t.Fatalf("collection replay = %#v, %v; want original result", replayed, err)
			}
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, result.RecordID, 1)
			if tc.name == "risk_ref" {
				readIDs := func(parent uuid.UUID) map[uuid.UUID]string {
					rows, err := harness.DB.Query(ctx, `SELECT risk_ref_id, risk_ref_text FROM handoff_risk_refs WHERE handoff_record_id=$1 AND deleted_at IS NULL`, parent)
					if err != nil {
						t.Fatal(err)
					}
					defer rows.Close()
					ids := map[uuid.UUID]string{}
					for rows.Next() {
						var id uuid.UUID
						var text string
						if err := rows.Scan(&id, &text); err != nil {
							t.Fatal(err)
						}
						if id == uuid.Nil || ids[id] != "" {
							t.Fatal("missing or duplicate risk identity")
						}
						ids[id] = text
					}
					if rows.Err() != nil || len(ids) != 2 {
						t.Fatalf("risk identities = %v, %v", ids, rows.Err())
					}
					return ids
				}
				firstIDs := readIDs(result.RecordID)
				for _, id := range createdRiskIDs {
					if _, exists := firstIDs[id]; !exists {
						t.Fatal("replay replaced a stable risk identity")
					}
				}
				other := command
				other.Admission = mustArtifactCreateAdmission(t, tc.viewSchemaID, "second-risk-parent", tc.values, tc.collections)
				other.RequestID = "second-risk-parent"
				second, err := facade.Create(ctx, other)
				if err != nil || second.RecordID == result.RecordID {
					t.Fatalf("second risk parent = %#v, %v", second, err)
				}
				tc.assert(second)
				for id := range readIDs(second.RecordID) {
					if _, exists := firstIDs[id]; exists {
						t.Fatal("risk identity escaped parent scope")
					}
				}
				for id, text := range readIDs(result.RecordID) {
					if firstIDs[id] != text {
						t.Fatal("existing risk identity changed")
					}
				}
			}

			conflicting := command
			conflicting.Admission = mustArtifactCreateAdmission(
				t, tc.viewSchemaID, clientTxnID, tc.values, duplicateFirstCollectionAction(t, tc.collections),
			)
			if _, err := facade.Create(ctx, conflicting); !errors.Is(err, artifacts.ErrClientTxnConflict) {
				t.Fatalf("changed collection replay error = %v, want client transaction conflict", err)
			}
		})
	}

}
