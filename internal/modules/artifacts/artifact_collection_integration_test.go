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
)

func TestArtifactCollectionMutationContractMatrix(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-collection-contract")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-collection-owner@example.test",
		"Artifact Collection Owner",
		"ArtifactCollectionOwner1!",
		false,
		false,
		true,
	)
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
INSERT INTO parties (record_id, incident_id, display_name, party_kind, created_at, updated_at)
VALUES ($1, $2, 'Synthetic responder', 'person', $3, $3)
`, partyRecordID, incident.ID, now); err != nil {
		t.Fatalf("seed party subtype: %v", err)
	}

	cases := []struct {
		name         string
		viewSchemaID string
		values       map[string]artifacts.FieldValue
		collections  map[string]artifacts.CollectionActionPayload
		assert       func(artifacts.MutationResult)
	}{
		{
			name:         "record_tag",
			viewSchemaID: artifacts.NotesViewSchemaID,
			values:       artifactTextValues("note.title", "Tagged note"),
			collections: map[string]artifacts.CollectionActionPayload{
				"note.tags": {Actions: []artifacts.CollectionAction{
					{Op: "add_tag", RawText: "Synthetic", NormalizedText: "synthetic"},
					{Op: "add_tag", RawText: "synthetic", NormalizedText: "synthetic"},
				}},
			},
			assert: func(result artifacts.MutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM record_tags WHERE record_id = $1 AND normalized_tag_name = 'synthetic' AND deleted_at IS NULL`, result.RecordID, 1)
			},
		},
		{
			name:         "record_ref",
			viewSchemaID: artifacts.FindingsViewSchemaID,
			values:       artifactTextValues("finding.statement", "Synthetic relationship"),
			collections: map[string]artifacts.CollectionActionPayload{
				"finding.supporting_refs": {Actions: []artifacts.CollectionAction{
					{Op: "add_record_ref", LinkedRecordID: &targetRecordID},
					{Op: "add_record_ref", LinkedRecordID: &targetRecordID},
				}},
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
			collections: map[string]artifacts.CollectionActionPayload{
				"comm_log.audience_party_ids": {Actions: []artifacts.CollectionAction{
					{Op: "add_party_ref", PartyID: &partyRecordID},
					{Op: "add_party_ref", PartyID: &partyRecordID},
				}},
			},
			assert: func(result artifacts.MutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM record_links WHERE src_record_id = $1 AND dst_record_id = $2 AND field_key = 'comm_log.audience_party_ids' AND deleted_at IS NULL`, result.RecordID, partyRecordID, 1)
			},
		},
		{
			name:         "risk_ref",
			viewSchemaID: artifacts.HandoffViewSchemaID,
			values: map[string]artifacts.FieldValue{
				"handoff.incoming_owner_user_id": {UUID: &actor.ID},
				"handoff.current_state_summary":  artifactTextValue("Synthetic handoff"),
			},
			collections: map[string]artifacts.CollectionActionPayload{
				"handoff.open_risk_refs": {Actions: []artifacts.CollectionAction{
					{Op: "add_risk_ref", RiskRefText: "Synthetic risk", NormalizedText: "synthetic risk"},
					{Op: "add_risk_ref", RiskRefText: "synthetic risk", NormalizedText: "synthetic risk"},
				}},
			},
			assert: func(result artifacts.MutationResult) {
				requireCount(t, harness, `SELECT count(*) FROM handoff_risk_refs WHERE handoff_record_id = $1 AND normalized_risk_ref_text = 'synthetic risk' AND deleted_at IS NULL`, result.RecordID, 1)
			},
		},
	}

	for index, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			clientTxnID := fmt.Sprintf("txn-artifacts-collection-%02d", index)
			command := artifacts.CreateCommand{
				ActorUserID: actor.ID, IncidentID: incident.ID,
				Request: artifacts.CreateRequest{
					ViewSchemaID: tc.viewSchemaID,
					ClientTxnID:  clientTxnID,
					Values:       tc.values,
					Collections:  tc.collections,
				},
				RequestHash: []byte("hash-" + clientTxnID),
				RequestID:   "req-" + clientTxnID,
				OperationID: artifacts.OperationCreate,
				Now:         now.Add(time.Duration(index) * time.Minute),
			}
			result, err := facade.Create(ctx, command)
			if err != nil {
				t.Fatalf("create collection family %s: %v", tc.name, err)
			}
			tc.assert(result)
			replayed, err := facade.Create(ctx, command)
			if err != nil || !replayed.Replayed || replayed.RecordID != result.RecordID {
				t.Fatalf("collection replay = %#v, %v; want original result", replayed, err)
			}
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, result.RecordID, 1)

			conflicting := command
			conflicting.RequestHash = []byte("changed-" + clientTxnID)
			if _, err := facade.Create(ctx, conflicting); !errors.Is(err, artifacts.ErrClientTxnConflict) {
				t.Fatalf("changed collection replay error = %v, want client transaction conflict", err)
			}
		})
	}

	beforeRecords := artifactContractCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID)
	_, err := facade.Create(ctx, artifacts.CreateCommand{
		ActorUserID: actor.ID, IncidentID: incident.ID,
		Request: artifacts.CreateRequest{
			ViewSchemaID: artifacts.NotesViewSchemaID,
			ClientTxnID:  "txn-artifacts-collection-invalid",
			Values:       artifactTextValues("note.title", "Rejected collection"),
			Collections: map[string]artifacts.CollectionActionPayload{
				"note.tags": {Actions: []artifacts.CollectionAction{{Op: "replace", RawText: "invalid"}}},
			},
		},
		RequestHash: []byte("hash-artifacts-collection-invalid"),
		RequestID:   "req-artifacts-collection-invalid",
		OperationID: artifacts.OperationCreate,
		Now:         now.Add(time.Hour),
	})
	if err == nil {
		t.Fatal("invalid collection operation unexpectedly succeeded")
	}
	if got := artifactContractCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID); got != beforeRecords {
		t.Fatalf("invalid collection changed record count: got %d want %d", got, beforeRecords)
	}
}
