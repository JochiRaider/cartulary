package entities_test

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestRevisionProviderContributionOwnsEntityHistorySemantics_Unit(t *testing.T) {
	contribution := entities.RevisionProviderContribution()
	if contribution.SourceOwnerModule != revisions.SourceOwnerEntities || len(contribution.Records) != 2 || len(contribution.NonRowTargets) != 3 {
		t.Fatalf("unexpected Entities revision contribution: %#v", contribution)
	}
	wantRecords := map[string]string{
		"host":     "cartulary.revisions.snapshot.host.v1",
		"identity": "cartulary.revisions.snapshot.identity.v1",
	}
	for _, record := range contribution.Records {
		if record.SnapshotSchemaID != wantRecords[record.RecordType] || !slices.Equal(record.HistoryTargetKinds, []string{record.RecordType}) {
			t.Fatalf("unexpected %s record semantics: %#v", record.RecordType, record)
		}
	}

	sourceID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	entityID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	for _, target := range contribution.NonRowTargets {
		if target.RollbackProvider == nil || target.HistorySemantics == nil {
			t.Fatalf("target %q has incomplete semantics", target.TargetKind)
		}
		mutation := revisions.StoredMutation{TargetKind: target.TargetKind, TargetID: uuid.NewString()}
		wantRecordID := entityID
		wantAddressable := false
		switch target.TargetKind {
		case "entity_mention":
			mutation.AfterValue = map[string]any{
				"source_record_id":   sourceID.String(),
				"resolved_record_id": entityID.String(),
			}
			wantRecordID = sourceID
			wantAddressable = true
		case "entity_alias", "entity_preserved_identifier":
			mutation.AfterValue = map[string]any{"record_id": entityID.String()}
		default:
			t.Fatalf("unexpected Entities target %q", target.TargetKind)
		}
		description, err := target.HistorySemantics.DescribeMutation(mutation)
		if err != nil {
			t.Fatalf("describe %s: %v", target.TargetKind, err)
		}
		if !slices.Equal(description.HistoryRecordIDs, []uuid.UUID{wantRecordID}) {
			t.Fatalf("unexpected %s record associations: %v", target.TargetKind, description.HistoryRecordIDs)
		}
		if wantAddressable != slices.Equal(description.HistoryEntryRecordIDs, []uuid.UUID{wantRecordID}) {
			t.Fatalf("unexpected %s addressability: %v", target.TargetKind, description.HistoryEntryRecordIDs)
		}
	}
}
