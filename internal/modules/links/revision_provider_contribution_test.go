package links_test

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestRevisionProviderContributionOwnsLinkAndTagHistorySemantics_Unit(t *testing.T) {
	contribution := links.RevisionProviderContribution()
	if contribution.SourceOwnerModule != revisions.SourceOwnerLinks || len(contribution.NonRowTargets) != 2 {
		t.Fatalf("unexpected Links revision contribution: %#v", contribution)
	}

	sourceID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	destinationID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	recordID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	for _, target := range contribution.NonRowTargets {
		if target.RollbackProvider == nil || target.HistorySemantics == nil {
			t.Fatalf("target %q has incomplete semantics", target.TargetKind)
		}
		var mutation revisions.StoredMutation
		var want []uuid.UUID
		switch target.TargetKind {
		case "record_link":
			mutation = revisions.StoredMutation{
				TargetKind: "record_link",
				TargetID:   uuid.NewString(),
				BeforeValue: map[string]any{
					"src_record_id": sourceID.String(),
					"dst_record_id": destinationID.String(),
				},
				AfterValue: map[string]any{
					"src_record_id": destinationID.String(),
					"dst_record_id": sourceID.String(),
				},
			}
			want = []uuid.UUID{sourceID, destinationID}
		case "record_tag":
			mutation = revisions.StoredMutation{
				TargetKind: "record_tag",
				TargetID:   "record_tag:" + recordID.String() + ":" + uuid.NewString(),
				AfterValue: map[string]any{"record_id": recordID.String()},
			}
			want = []uuid.UUID{recordID}
		default:
			t.Fatalf("unexpected Links target %q", target.TargetKind)
		}
		description, err := target.HistorySemantics.DescribeMutation(mutation)
		if err != nil {
			t.Fatalf("describe %s: %v", target.TargetKind, err)
		}
		if !slices.Equal(description.HistoryRecordIDs, want) || !slices.Equal(description.HistoryEntryRecordIDs, want) {
			t.Fatalf("unexpected %s associations: records=%v entries=%v want=%v", target.TargetKind, description.HistoryRecordIDs, description.HistoryEntryRecordIDs, want)
		}
	}
}
