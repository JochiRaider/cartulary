package indicators

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestIndicatorCreateReplayCodecPreservesStoredBytes(t *testing.T) {
	t.Parallel()
	changeSetID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	row := map[string]any{
		"record_id":   "00000000-0000-4000-8000-000000000002",
		"incident_id": "00000000-0000-4000-8000-000000000003",
		"row_version": int64(4),
	}
	encoded, err := json.Marshal(buildStoredCreateResponse(changeSetID, row))
	if err != nil {
		t.Fatalf("encode stored Indicator create response: %v", err)
	}
	want := `{"change_set_id":"00000000-0000-4000-8000-000000000001","row":{"incident_id":"00000000-0000-4000-8000-000000000003","record_id":"00000000-0000-4000-8000-000000000002","row_version":4},"view_schema_id":"cartulary.view.indicators.v1"}`
	if string(encoded) != want {
		t.Fatalf("stored Indicator create bytes changed:\ngot:  %s\nwant: %s", encoded, want)
	}
	decoded, err := decodeStoredResponse(encoded)
	if err != nil {
		t.Fatalf("decode stored Indicator create response: %v", err)
	}
	if got, err := extractUUIDFromPayload(decoded, "change_set_id"); err != nil || got != changeSetID {
		t.Fatalf("decoded change set = %s, err=%v", got, err)
	}
	if got, err := extractInt64FromPayload(decoded, "row", "row_version"); err != nil || got != 4 {
		t.Fatalf("decoded row version = %d, err=%v", got, err)
	}
}
