package workbookassembly

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
)

func TestIndicatorValidationFailuresHaveOneWorkbookTranslation(t *testing.T) {
	t.Parallel()

	provider, err := newIndicatorCreateProvider(&indicators.Application{})
	if err != nil {
		t.Fatalf("compose Indicator create provider: %v", err)
	}
	_, decoderFailure, err := provider.DecodeCreate(strings.NewReader(`{"client_txn_id":"txn","unknown":"value"}`))
	if err != nil {
		t.Fatalf("decode Indicator create: %v", err)
	}
	assertIndicatorFailureDetail(t, decoderFailure, "unknown", "unknown_field")

	ownerFailure, safe := indicatorCreateFailure(&indicators.IndicatorCreateValidationError{
		Field: "indicator.indicator_type", ReasonCode: "invalid_value",
	}, "txn")
	if !safe {
		t.Fatal("owner validation was not classified as a safe Workbook failure")
	}
	assertIndicatorFailureDetail(t, ownerFailure, "indicator.indicator_type", "invalid_value")
}

func assertIndicatorFailureDetail(t testing.TB, failure interface {
	InvalidPayloadDetail() (string, string, bool)
}, wantField string, wantReason string) {
	t.Helper()
	if failure == nil {
		t.Fatal("Workbook failure is nil")
	}
	field, reason, ok := failure.InvalidPayloadDetail()
	if !ok || field != wantField || reason != wantReason {
		t.Fatalf("Workbook failure detail = (%q, %q, %t), want (%q, %q, true)", field, reason, ok, wantField, wantReason)
	}
}

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
