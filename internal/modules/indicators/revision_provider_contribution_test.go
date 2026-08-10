package indicators_test

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestRevisionProviderContributionOwnsIndicatorHistorySemantics_Unit(t *testing.T) {
	contribution := indicators.NewRevisionContribution()
	if contribution.SourceOwnerModule != revisions.SourceOwnerIndicators || len(contribution.Records) != 1 || len(contribution.NonRowTargets) != 2 {
		t.Fatalf("unexpected Indicators revision contribution: %#v", contribution)
	}
	record := contribution.Records[0]
	if record.RecordType != "indicator" || record.SnapshotSchemaID != "cartulary.revisions.snapshot.indicator.v1" ||
		!slices.Equal(record.HistoryTargetKinds, []string{"indicator"}) {
		t.Fatalf("unexpected Indicator record semantics: %#v", record)
	}

	sourceID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	indicatorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	for _, target := range contribution.NonRowTargets {
		if target.RollbackProvider == nil || target.HistorySemantics == nil {
			t.Fatalf("target %q has incomplete semantics", target.TargetKind)
		}
		var mutation revisions.StoredMutation
		var want []uuid.UUID
		switch target.TargetKind {
		case "indicator_observation":
			mutation = revisions.StoredMutation{
				TargetKind: "indicator_observation",
				TargetID:   uuid.NewString(),
				BeforeValue: map[string]any{
					"source_record_id":             sourceID.String(),
					"resolved_indicator_record_id": indicatorID.String(),
				},
				AfterValue: map[string]any{
					"source_record_id": sourceID.String(),
				},
			}
			want = []uuid.UUID{sourceID, indicatorID}
		case "indicator_state_interval":
			mutation = revisions.StoredMutation{
				TargetKind: "indicator_state_interval",
				TargetID:   uuid.NewString(),
				AfterValue: map[string]any{"indicator_record_id": indicatorID.String()},
			}
			want = []uuid.UUID{indicatorID}
		default:
			t.Fatalf("unexpected Indicators target %q", target.TargetKind)
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
