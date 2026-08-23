package workbookassembly

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
)

func TestIndicatorMutationResultUsesOnlyConsumerRequiredSignals(t *testing.T) {
	t.Parallel()
	incidentID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	recordID := uuid.MustParse("00000000-0000-4000-8000-000000000002")
	changeSetID := uuid.MustParse("00000000-0000-4000-8000-000000000003")
	base := indicators.CreateResult{
		CanonicalRow: map[string]any{"record_id": recordID.String(), "row_version": int64(7)},
		RecordID:     recordID, ChangeSetID: changeSetID, RowVersion: 7,
	}

	created := base
	created.Created = true
	createdResult := indicatorMutationResult(created, incidentID, "txn-created")
	if createdResult.StatusCode != http.StatusCreated || createdResult.Replayed {
		t.Fatalf("created result = %#v, want 201 and not replayed", createdResult)
	}

	existingResult := indicatorMutationResult(base, incidentID, "txn-existing")
	if existingResult.StatusCode != http.StatusOK || existingResult.Replayed {
		t.Fatalf("existing result = %#v, want 200 and not replayed", existingResult)
	}

	replayed := base
	replayed.Replayed = true
	replayedResult := indicatorMutationResult(replayed, incidentID, "txn-replayed")
	if replayedResult.StatusCode != http.StatusOK || !replayedResult.Replayed {
		t.Fatalf("replayed result = %#v, want 200 and replayed", replayedResult)
	}
	if replayedResult.RecordID != recordID || replayedResult.ChangeSetID != changeSetID || replayedResult.RowVersion != 7 {
		t.Fatalf("consumer signal contraction changed mutation identity: %#v", replayedResult)
	}
}
