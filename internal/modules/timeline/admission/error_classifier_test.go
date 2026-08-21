package admission

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestTimelineMutationAPIErrorClassifier_Unit(t *testing.T) {
	recordID := uuid.New()
	actorID := uuid.New()
	updatedAt := time.Date(2026, 7, 25, 12, 30, 0, 0, time.UTC)
	conflict := timeline.SameFieldConflict{
		ConflictToken: "opaque-token", RecordID: recordID,
		FieldKey: "timeline.activity_synopsis_text", ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion: 3, CurrentRowVersion: 4,
		ClientValue: "client", ServerValue: "server",
		BaseValue:       timeline.OptionalConflictValue{Present: true, Value: "base"},
		ServerUpdatedBy: actorID, ServerUpdatedAt: updatedAt,
	}
	conflictPayload := map[string]any{
		"conflict_token": "opaque-token", "record_id": recordID.String(),
		"field_key": "timeline.activity_synopsis_text", "conflict_resolution_class": "text_compare_merge",
		"base_row_version": int64(3), "current_row_version": int64(4),
		"client_value": "client", "server_value": "server", "base_value": "base",
		"server_updated_by": actorID.String(), "server_updated_at": updatedAt.Format(time.RFC3339Nano),
	}

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
			err:     timeline.ErrIncidentClosed,
			status:  http.StatusConflict,
			code:    "incident_closed",
			message: "incident closed",
			details: map[string]any{},
		},
		{
			name:    "record not found",
			err:     timeline.ErrRecordNotFound,
			status:  http.StatusNotFound,
			code:    "incident_not_found",
			details: map[string]any{},
		},
		{
			name:    "deleted record use restore",
			err:     timeline.ErrRecordDeleted,
			status:  http.StatusConflict,
			code:    "record_deleted_use_restore",
			message: "record deleted use restore",
			details: map[string]any{},
		},
		{
			name:     "same field conflict",
			err:      &timeline.SameFieldConflictError{Conflict: conflict},
			status:   http.StatusConflict,
			code:     "same_field_conflict",
			message:  "same field conflict",
			details:  map[string]any{},
			conflict: conflictPayload,
		},
		{
			name: "typed row version conflict",
			err: &timeline.RowVersionConflictError{
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
			err:     timeline.ErrRowVersionConflict,
			status:  http.StatusConflict,
			code:    "row_version_conflict",
			details: map[string]any{},
		},
		{
			name: "typed illegal transition",
			err: &timeline.IllegalTransitionError{
				ReasonCode:     "typed_reason",
				FromStatus:     "rough",
				ToStatus:       "reviewed",
				ViolatedGuards: []string{"guard_a"},
			},
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
			err:     timeline.ErrIllegalTransition,
			context: MutationAPIErrorContext{IllegalTransitionReasonCode: "superseded_terminal"},
			status:  http.StatusConflict,
			code:    "illegal_transition",
			message: "illegal transition",
			details: map[string]any{"reason_code": "superseded_terminal"},
		},
		{
			name:    "no effective change",
			err:     timeline.ErrNoEffectiveChange,
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
