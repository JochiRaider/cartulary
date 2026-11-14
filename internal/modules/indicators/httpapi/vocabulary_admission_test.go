package httpapi

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestIndicatorHTTPVocabularyAdmissionIsExact(t *testing.T) {
	t.Parallel()
	t.Run("composition rejects missing dependencies", testIndicatorHTTPCompositionRejectsMissingDependencies)
	for _, value := range []any{nil, "", " \t", "réviewed"} {
		text, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		body := fmt.Sprintf(`{"client_txn_id":"txn","base_row_version":1,"lifecycle_state":"active","valid_from":"2026-08-22T12:00:00Z","valid_to":null,"confidence":null,"rationale":%s,"support_refs":[],"assessor":%s}`, text, text)
		request, apiErr := decodeLifecycleAppend(strings.NewReader(body))
		if apiErr != nil {
			t.Fatalf("nullable interval text rejected: %v", apiErr)
		}
		for _, got := range []*string{request.Rationale, request.Assessor} {
			if value == nil && got != nil || value != nil && (got == nil || *got != value) {
				t.Fatalf("nullable interval text changed: got=%v want=%#v", got, value)
			}
		}
	}
	for _, field := range []string{"rationale", "assessor"} {
		body := map[string]any{"client_txn_id": "txn", "base_row_version": 1, "lifecycle_state": "active", "valid_from": "2026-08-22T12:00:00Z", "valid_to": nil, "confidence": nil, "rationale": nil, "support_refs": []string{}, "assessor": nil}
		body[field] = "bad\x00text"
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		if _, apiErr := decodeLifecycleAppend(strings.NewReader(string(payload))); apiErr == nil {
			t.Fatalf("NUL in %s accepted", field)
		}
	}

	for _, value := range []string{"DOMAIN_NAME", " domain_name", "domain_name ", "domain", "unknown"} {
		value := value
		t.Run("Indicator type/"+value, func(t *testing.T) {
			t.Parallel()
			body := fmt.Sprintf(`{"client_txn_id":"txn","base_row_version":1,"source_field_key":"source_text","span_start_byte":0,"span_end_byte":1,"parsed_indicator_type":%q}`, value)
			if _, apiErr := decodeObservationCreate(strings.NewReader(body)); apiErr == nil {
				t.Fatalf("parsed_indicator_type %q was accepted", value)
			}
		})
	}
	for _, value := range []string{"ACTIVE", " active", "active ", "inactive", "unknown"} {
		value := value
		t.Run("lifecycle/"+value, func(t *testing.T) {
			t.Parallel()
			body := fmt.Sprintf(`{"client_txn_id":"txn","base_row_version":1,"lifecycle_state":%q,"valid_from":"2026-08-22T12:00:00Z","valid_to":null,"confidence":null,"rationale":null,"support_refs":[],"assessor":null}`, value)
			if _, apiErr := decodeLifecycleAppend(strings.NewReader(body)); apiErr == nil {
				t.Fatalf("lifecycle_state %q was accepted", value)
			}
		})
	}
}
