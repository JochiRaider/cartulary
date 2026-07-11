package timeline_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/fixtures"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/golden"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
)

// U-4-08 / REQ-01-057..REQ-01-088, REQ-01-228..REQ-01-239, REQ-01-315..REQ-01-316, REQ-01-568, REQ-02-163..REQ-02-185, REQ-03-205..REQ-03-216, REQ-03-276..REQ-03-279 / AC-205, AC-388..AC-392.
func TestPhase4_AutoResolutionEligibility_U_4_08(t *testing.T) {
	t.Run("mention token contract preserves raw text and collapses whitespace for comparison", func(t *testing.T) {
		payload := fixtures.TimelineCollectionPatchPayload(
			golden.RecordFieldTimelineHostRefs,
			7,
			"txn-phase4-u-4-08-normalize",
			fixtures.CollectionActions(
				fixtures.AddTokenAction(" vpn   gateway "),
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
		if request.CanonicalChange[0].FieldKey != golden.RecordFieldTimelineHostRefs {
			t.Fatalf("unexpected decoded field_key: %#v", request.CanonicalChange)
		}
		action := request.CanonicalChange[0].ActionPayload.Actions[0]
		if action.RawText != " vpn   gateway " {
			t.Fatalf("expected raw token text to remain authoritative, got %#v", action)
		}
		contractassert.RequireWritableStringNormalization(t, action.NormalizedText, "vpn gateway")
	})

	t.Run("suppressor and forbidden rewrite tokens remain valid submitted tokens", func(t *testing.T) {
		tokenCases := append([]string{}, golden.RecordAutoResolutionSuppressedTokens...)
		for _, rawText := range tokenCases {
			t.Run(rawText, func(t *testing.T) {
				payload := fixtures.TimelineCollectionPatchPayload(
					golden.RecordFieldTimelineHostRefs,
					7,
					"txn-phase4-u-4-08-"+rawText,
					fixtures.CollectionActions(
						fixtures.AddTokenAction(rawText),
					),
				)
				data, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("marshal Phase 4 collection patch payload: %v", err)
				}

				request, apiErr := timeline.DecodeTimelinePatchRequest(bytes.NewReader(data))
				if apiErr != nil {
					t.Fatalf("expected suppressor/forbidden rewrite token %q to decode, got %#v", rawText, apiErr)
				}
				action := request.CanonicalChange[0].ActionPayload.Actions[0]
				if action.RawText != rawText {
					t.Fatalf("expected raw_text %q to remain authoritative, got %#v", rawText, action)
				}
			})
		}
	})
}

// U-4-09 / REQ-01-311, REQ-01-314..REQ-01-320, REQ-02-248, REQ-03-280 / AC-394, AC-396, AC-397.
func TestPhase4_ManualTimelineConfidenceNull_U_4_09(t *testing.T) {
	t.Run("manual relationship mutation omits confidence and should decode", func(t *testing.T) {
		payload := fixtures.TimelineCollectionPatchPayload(
			golden.RecordFieldTimelineIdentityRefs,
			4,
			"txn-phase4-u-4-09-manual",
			fixtures.CollectionActions(
				fixtures.AddResolvedRefAction("alex.analyst@example.test", golden.RecordCanonicalIdentityID),
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
		if request.CanonicalChange[0].FieldKey != golden.RecordFieldTimelineIdentityRefs {
			t.Fatalf("unexpected decoded field_key: %#v", request.CanonicalChange)
		}
	})

	t.Run("manual create-time add_resolved_ref omits confidence and should decode", func(t *testing.T) {
		payload := map[string]any{
			"client_txn_id": "txn-phase4-u-4-09-create-manual",
			golden.RecordFieldTimelineHostRefs: fixtures.CollectionActions(
				fixtures.AddResolvedRefAction("WS-023", golden.RecordCanonicalHostRecordID),
			),
		}
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal create manual relationship payload: %v", err)
		}

		request, apiErr := timeline.DecodeTimelineCreateRequest(bytes.NewReader(data))
		if apiErr != nil {
			t.Fatalf("expected create-time manual relationship mutation without confidence to decode, got %#v", apiErr)
		}
		if request.HostRefs == nil || len(request.HostRefs.Actions) != 1 || request.HostRefs.Actions[0].Op != "add_resolved_ref" {
			t.Fatalf("unexpected decoded create host_refs payload: %#v", request.HostRefs)
		}
	})

	t.Run("client supplied forbidden metadata should fail closed", func(t *testing.T) {
		cases := []struct {
			name   string
			field  string
			action map[string]any
		}{
			{
				name:  "add_resolved_ref confidence number",
				field: golden.RecordFieldTimelineHostRefs,
				action: map[string]any{
					"op":                 "add_resolved_ref",
					"raw_text":           "WS-023",
					"resolved_record_id": golden.RecordCanonicalHostRecordID.String(),
					"confidence":         80,
				},
			},
			{
				name:  "add_resolved_ref confidence null",
				field: golden.RecordFieldTimelineIdentityRefs,
				action: map[string]any{
					"op":                 "add_resolved_ref",
					"raw_text":           "alex.analyst@example.test",
					"resolved_record_id": golden.RecordCanonicalIdentityID.String(),
					"confidence":         nil,
				},
			},
			{
				name:  "resolve_item confidence string",
				field: golden.RecordFieldTimelineHostRefs,
				action: map[string]any{
					"op":                 "resolve_item",
					"item_ref":           fixtures.MentionItemRef(golden.RecordHostMentionID),
					"resolved_record_id": golden.RecordCanonicalHostRecordID.String(),
					"confidence":         "80",
				},
			},
			{
				name:  "resolve_item provenance override",
				field: golden.RecordFieldTimelineIdentityRefs,
				action: map[string]any{
					"op":                 "resolve_item",
					"item_ref":           fixtures.MentionItemRef(golden.RecordIdentityMentionID),
					"resolved_record_id": golden.RecordCanonicalIdentityID.String(),
					"provenance":         "auto_match",
				},
			},
			{
				name:  "add_resolved_ref link_type override",
				field: golden.RecordFieldTimelineHostRefs,
				action: map[string]any{
					"op":                 "add_resolved_ref",
					"raw_text":           "WS-023",
					"resolved_record_id": golden.RecordCanonicalHostRecordID.String(),
					"link_type":          "observed_as_identity",
				},
			},
			{
				name:  "resolve_item source routing metadata",
				field: golden.RecordFieldTimelineHostRefs,
				action: map[string]any{
					"op":                 "resolve_item",
					"item_ref":           fixtures.MentionItemRef(golden.RecordHostMentionID),
					"resolved_record_id": golden.RecordCanonicalHostRecordID.String(),
					"source_record_id":   golden.RecordTimelineRecordID.String(),
				},
			},
			{
				name:  "add_resolved_ref target routing metadata",
				field: golden.RecordFieldTimelineIdentityRefs,
				action: map[string]any{
					"op":                 "add_resolved_ref",
					"raw_text":           "alex.analyst@example.test",
					"resolved_record_id": golden.RecordCanonicalIdentityID.String(),
					"target_record_id":   golden.RecordCanonicalIdentityID.String(),
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				payload := fixtures.TimelineCollectionPatchPayload(
					tc.field,
					4,
					"txn-phase4-u-4-09-"+tc.name,
					fixtures.CollectionActions(tc.action),
				)
				data, err := json.Marshal(payload)
				if err != nil {
					t.Fatalf("marshal forbidden metadata payload: %v", err)
				}

				_, apiErr := timeline.DecodeTimelinePatchRequest(bytes.NewReader(data))
				if apiErr == nil {
					t.Fatal("expected client-supplied metadata to fail closed")
				}
				contractassert.RequireClosedVocabularyRejected(t, apiErr.Code, apiErr.Details, tc.field, "invalid_value")
			})
		}
	})
}
