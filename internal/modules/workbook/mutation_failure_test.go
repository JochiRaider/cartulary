package workbook

import (
	"reflect"
	"testing"
)

func TestInvalidPayloadFailureDetailMapping_Unit(t *testing.T) {
	tests := []struct {
		name        string
		failure     *MutationFailure
		wantDetails map[string]any
	}{
		{
			name:    "plain",
			failure: InvalidPayloadFailure("view_schema_id", "unknown_view_schema"),
			wantDetails: map[string]any{
				"field": "view_schema_id", "reason_code": "unknown_view_schema",
			},
		},
		{
			name: "collection context without counts",
			failure: InvalidPayloadCollectionFailure(
				"decision.support_refs.actions", "empty_collection_actions", "decision.support_refs",
			),
			wantDetails: map[string]any{
				"field": "decision.support_refs.actions", "reason_code": "empty_collection_actions",
				"field_key": "decision.support_refs",
			},
		},
		{
			name:    "patch count limit",
			failure: InvalidPayloadLimitFailure("changes", "change_count_exceeded", 33, 32, ""),
			wantDetails: map[string]any{
				"field": "changes", "reason_code": "change_count_exceeded",
				"requested_count": 33, "max_count": 32,
			},
		},
		{
			name: "collection count limit",
			failure: InvalidPayloadLimitFailure(
				"decision.support_refs.actions", "collection_action_count_exceeded", 65, 64, "decision.support_refs",
			),
			wantDetails: map[string]any{
				"field": "decision.support_refs.actions", "reason_code": "collection_action_count_exceeded",
				"field_key": "decision.support_refs", "requested_count": 65, "max_count": 64,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			apiError := mutationFailureAPIError(test.failure)
			if apiError.Status != 400 || apiError.Code != "invalid_mutation_payload" ||
				apiError.Message != "invalid mutation payload" {
				t.Fatalf("public mapping = %#v", apiError)
			}
			if !reflect.DeepEqual(apiError.Details, test.wantDetails) {
				t.Fatalf("details = %#v, want %#v", apiError.Details, test.wantDetails)
			}
		})
	}
}
