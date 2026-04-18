package timeline_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
)

// U-4-08 / REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 / AC-205, AC-388..AC-392.
func TestPhase4_AutoResolutionEligibility_U_4_08_Red(t *testing.T) {
	payload := fixtures.TimelineCollectionPatchPayload(
		golden.Phase4FieldTimelineHostRefs,
		7,
		"txn-phase4-u-4-08",
		fixtures.CollectionActions(
			fixtures.AddTokenAction(golden.Phase4AutoResolutionEligibleTokens[0]),
			fixtures.AddTokenAction(golden.Phase4AutoResolutionSuppressedTokens[0]),
		),
	)
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal Phase 4 collection patch payload: %v", err)
	}

	request, apiErr := timeline.DecodeTimelinePatchRequest(bytes.NewReader(data))
	if apiErr != nil {
		t.Fatalf("expected Phase 4 Timeline relationship collection patch to decode for auto-resolution eligibility assertions, got %#v", apiErr)
	}
	if request.CanonicalChange[0].FieldKey != golden.Phase4FieldTimelineHostRefs {
		t.Fatalf("unexpected decoded field_key: %#v", request.CanonicalChange)
	}
}

// U-4-09 / REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-248, REQ-03-280 / AC-394, AC-396, AC-397.
func TestPhase4_ManualTimelineConfidenceNull_U_4_09_Red(t *testing.T) {
	t.Run("manual relationship mutation omits confidence and should decode", func(t *testing.T) {
		payload := fixtures.TimelineCollectionPatchPayload(
			golden.Phase4FieldTimelineIdentityRefs,
			4,
			"txn-phase4-u-4-09-manual",
			fixtures.CollectionActions(
				fixtures.AddResolvedRefAction("alex.analyst@example.test", golden.Phase4CanonicalIdentityID),
			),
		)
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal manual relationship payload: %v", err)
		}

		request, apiErr := timeline.DecodeTimelinePatchRequest(bytes.NewReader(data))
		if apiErr != nil {
			t.Fatalf("expected manual relationship mutation without confidence to decode, got %#v", apiErr)
		}
		if request.CanonicalChange[0].FieldKey != golden.Phase4FieldTimelineIdentityRefs {
			t.Fatalf("unexpected decoded field_key: %#v", request.CanonicalChange)
		}
	})

	t.Run("client supplied confidence should fail closed", func(t *testing.T) {
		payload := fixtures.TimelineCollectionPatchPayload(
			golden.Phase4FieldTimelineHostRefs,
			4,
			"txn-phase4-u-4-09-confidence",
			fixtures.CollectionActions(
				map[string]any{
					"op":                 "add_resolved_ref",
					"raw_text":           "WS-023",
					"resolved_record_id": golden.Phase4CanonicalHostRecordID.String(),
					"confidence":         80,
				},
			),
		)
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal confidence-bearing payload: %v", err)
		}

		_, apiErr := timeline.DecodeTimelinePatchRequest(bytes.NewReader(data))
		if apiErr == nil {
			t.Fatal("expected client-supplied confidence to fail closed")
		}
		if apiErr.Code != "invalid_mutation_payload" {
			t.Fatalf("unexpected error code: %#v", apiErr)
		}
	})
}
