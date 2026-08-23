package httpapi

import (
	"fmt"
	"strings"
	"testing"
)

func TestIndicatorHTTPVocabularyAdmissionIsExact(t *testing.T) {
	t.Parallel()
	t.Run("composition rejects missing dependencies", testIndicatorHTTPCompositionRejectsMissingDependencies)
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
