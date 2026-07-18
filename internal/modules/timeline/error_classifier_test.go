package timeline

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestTimelineMutationAPIErrorClassifier_Unit(t *testing.T) {
	recordID := uuid.New()
	conflict := map[string]any{"field_key": "timeline.activity_synopsis_text"}

	tests := []struct {
		name     string
		err      error
		context  MutationAPIErrorContext
		status   int
		code     string
		message  string
		details  map[string]any
		conflict any
	}{
		{
			name:    "client txn conflict",
			err:     authn.ErrClientTxnConflict,
			context: MutationAPIErrorContext{ClientTxnID: "txn-1"},
			status:  http.StatusConflict,
			code:    "client_txn_conflict",
			message: "client transaction conflicts with an existing request",
			details: map[string]any{"client_txn_id": "txn-1"},
		},
		{
			name:    "incident closed",
			err:     incidents.ErrIncidentClosed,
			status:  http.StatusConflict,
			code:    "incident_closed",
			message: "incident closed",
			details: map[string]any{},
		},
		{
			name:    "record not found",
			err:     ErrRecordNotFound,
			status:  http.StatusNotFound,
			code:    "incident_not_found",
			details: map[string]any{},
		},
		{
			name:    "deleted record use restore",
			err:     revisions.ErrRecordDeletedUseRestore,
			status:  http.StatusConflict,
			code:    "record_deleted_use_restore",
			message: "record deleted use restore",
			details: map[string]any{},
		},
		{
			name:     "same field conflict",
			err:      &SameFieldConflictError{Conflict: conflict},
			status:   http.StatusConflict,
			code:     "same_field_conflict",
			message:  "same field conflict",
			details:  map[string]any{},
			conflict: conflict,
		},
		{
			name: "typed row version conflict",
			err: &RowVersionConflictError{
				RecordID:          recordID,
				BaseRowVersion:    3,
				CurrentRowVersion: 4,
			},
			status: http.StatusConflict,
			code:   "row_version_conflict",
			details: map[string]any{
				"record_id":           recordID.String(),
				"base_row_version":    int64(3),
				"current_row_version": int64(4),
			},
		},
		{
			name:    "sentinel row version conflict",
			err:     ErrRowVersionConflict,
			status:  http.StatusConflict,
			code:    "row_version_conflict",
			details: map[string]any{},
		},
		{
			name:    "typed illegal transition",
			err:     newIllegalTransitionError("typed_reason", "rough", "reviewed", "guard_a"),
			context: MutationAPIErrorContext{IllegalTransitionReasonCode: "fallback_reason"},
			status:  http.StatusConflict,
			code:    "illegal_transition",
			message: "illegal transition",
			details: map[string]any{
				"reason_code":     "typed_reason",
				"from_status":     "rough",
				"to_status":       "reviewed",
				"violated_guards": []string{"guard_a"},
			},
		},
		{
			name:    "sentinel illegal transition",
			err:     ErrIllegalTransition,
			context: MutationAPIErrorContext{IllegalTransitionReasonCode: "superseded_terminal"},
			status:  http.StatusConflict,
			code:    "illegal_transition",
			message: "illegal transition",
			details: map[string]any{"reason_code": "superseded_terminal"},
		},
		{
			name:    "no effective change",
			err:     ErrNoEffectiveChange,
			context: MutationAPIErrorContext{NoEffectiveChangeField: "action_payload"},
			status:  http.StatusBadRequest,
			code:    "invalid_mutation_payload",
			message: "invalid mutation payload",
			details: map[string]any{"field": "action_payload", "reason_code": "no_effective_change"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := fmt.Errorf("wrapped: %w", tc.err)
			apiErr, ok := ClassifyMutationAPIError(wrapped, tc.context)
			if !ok {
				t.Fatalf("expected %s to be classified", tc.name)
			}
			if apiErr.Status != tc.status || apiErr.Code != tc.code || apiErr.Message != tc.message {
				t.Fatalf("unexpected envelope: status=%d code=%s message=%q", apiErr.Status, apiErr.Code, apiErr.Message)
			}
			if !reflect.DeepEqual(apiErr.Details, tc.details) {
				t.Fatalf("unexpected details:\nwant %#v\ngot  %#v", tc.details, apiErr.Details)
			}
			if !reflect.DeepEqual(apiErr.Conflict, tc.conflict) {
				t.Fatalf("unexpected conflict:\nwant %#v\ngot  %#v", tc.conflict, apiErr.Conflict)
			}
		})
	}

	if apiErr, ok := ClassifyMutationAPIError(errors.New("other"), MutationAPIErrorContext{}); ok || apiErr != nil {
		t.Fatalf("unclassified error returned ok=%v apiErr=%#v", ok, apiErr)
	}
}
