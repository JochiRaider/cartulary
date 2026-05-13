package timeline

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPhase3_PatchPayloadValidation_U_3_06(t *testing.T) {
	request, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v1",
		"base_row_version": 2,
		"client_txn_id": "txn-u-3-06",
		"changes": [
			{ "field_key": "timeline.summary", "value": "B" },
			{ "field_key": "timeline.details", "value": "A" }
		]
	}`))
	if apiErr != nil {
		t.Fatalf("expected valid patch request, got %#v", apiErr)
	}
	fieldKeys := []string{
		request.CanonicalChange[0].FieldKey,
		request.CanonicalChange[1].FieldKey,
	}
	requireFieldKeyConformance(t, fieldKeys, []string{
		"timeline.details",
		"timeline.host_refs",
		"timeline.identity_refs",
		"timeline.occurred_at",
		"timeline.source_text",
		"timeline.summary",
		"timeline.tags",
	})

	t.Run("thirty two raw changes do not trip the count ceiling", func(t *testing.T) {
		rawChanges := make([]map[string]any, 0, maxPatchChanges)
		for range maxPatchChanges {
			rawChanges = append(rawChanges, map[string]any{
				"field_key": "timeline.summary",
				"value":     "summary-value",
			})
		}
		payload, err := json.Marshal(map[string]any{
			"view_schema_id":   TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-u-3-06-max-raw",
			"changes":          rawChanges,
		})
		if err != nil {
			t.Fatalf("marshal patch payload: %v", err)
		}

		_, apiErr := DecodeTimelinePatchRequest(bytes.NewReader(payload))
		if apiErr == nil {
			t.Fatal("expected duplicate field rejection at the ceiling")
		}
		requireClosedVocabularyRejected(
			t,
			apiErr.Code,
			apiErr.Details,
			"changes",
			"duplicate_field_key",
		)
	})

	t.Run("sixty four collection actions remain valid", func(t *testing.T) {
		actions := make([]map[string]any, 0, maxCollectionActions)
		for index := range maxCollectionActions {
			actions = append(actions, map[string]any{
				"op":       "add_token",
				"raw_text": "host-token-" + string(rune('a'+(index%26))),
			})
		}
		payload, err := json.Marshal(map[string]any{
			"view_schema_id":   TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-u-3-06-max-actions",
			"changes": []map[string]any{
				{
					"field_key": "timeline.host_refs",
					"action_payload": map[string]any{
						"kind":    "collection_actions_v1",
						"actions": actions,
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("marshal collection patch payload: %v", err)
		}

		request, apiErr := DecodeTimelinePatchRequest(bytes.NewReader(payload))
		if apiErr != nil {
			t.Fatalf("expected max collection action payload to decode, got %#v", apiErr)
		}
		if got := len(request.CanonicalChange[0].ActionPayload.Actions); got != maxCollectionActions {
			t.Fatalf("unexpected decoded action count: got %d want %d", got, maxCollectionActions)
		}
	})

	t.Run("timeline tags collection actions remain valid", func(t *testing.T) {
		request, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
			"view_schema_id": "cartulary.view.timeline.v1",
			"base_row_version": 1,
			"client_txn_id": "txn-u-3-06-tags",
			"changes": [
				{
					"field_key": "timeline.tags",
					"action_payload": {
						"kind": "collection_actions_v1",
						"actions": [
							{ "op": "add_tag", "tag_name": "critical-host" }
						]
					}
				}
			]
		}`))
		if apiErr != nil {
			t.Fatalf("expected timeline.tags payload to decode, got %#v", apiErr)
		}
		if got := request.CanonicalChange[0].FieldKey; got != "timeline.tags" {
			t.Fatalf("unexpected field key: got %q", got)
		}
		if got := request.CanonicalChange[0].ActionPayload.Actions[0].NormalizedText; got != "critical-host" {
			t.Fatalf("unexpected normalized tag label: got %q", got)
		}
	})

	t.Run("timeline attached evidence accepts record ref actions", func(t *testing.T) {
		request, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
			"view_schema_id": "cartulary.view.timeline.v1",
			"base_row_version": 1,
			"client_txn_id": "txn-u-5-attached-evidence",
			"changes": [
				{
					"field_key": "timeline.attached_evidence_ids",
					"action_payload": {
						"kind": "collection_actions_v1",
						"actions": [
							{ "op": "add_record_ref", "linked_record_id": "00000000-0000-0000-0000-000000000001" }
						]
					}
				}
			]
		}`))
		if apiErr != nil {
			t.Fatalf("expected timeline.attached_evidence_ids payload to decode, got %#v", apiErr)
		}
		action := request.CanonicalChange[0].ActionPayload.Actions[0]
		if action.Op != "add_record_ref" || action.LinkedRecordID == nil {
			t.Fatalf("unexpected attached evidence action: %#v", action)
		}
	})

	cases := []struct {
		name   string
		body   string
		field  string
		reason string
	}{
		{
			name:   "unknown top level member",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.summary","value":"x"}],"unknown":"value"}`,
			field:  "unknown",
			reason: "unknown_field",
		},
		{
			name:   "missing view schema",
			body:   `{"base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.summary","value":"x"}]}`,
			field:  "view_schema_id",
			reason: "missing_required_field",
		},
		{
			name:   "missing base row version",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","client_txn_id":"txn","changes":[{"field_key":"timeline.summary","value":"x"}]}`,
			field:  "base_row_version",
			reason: "missing_required_field",
		},
		{
			name:   "missing client txn",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"changes":[{"field_key":"timeline.summary","value":"x"}]}`,
			field:  "client_txn_id",
			reason: "missing_required_field",
		},
		{
			name:   "empty changes",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[]}`,
			field:  "changes",
			reason: "empty_changes",
		},
		{
			name:   "duplicate field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.summary","value":"x"},{"field_key":"timeline.summary","value":"y"}]}`,
			field:  "changes",
			reason: "duplicate_field_key",
		},
		{
			name:   "readonly system field",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.capture_state","value":"reviewed"}]}`,
			field:  "field_key",
			reason: "unsupported_field_key",
		},
		{
			name:   "visible label is not a field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"Summary","value":"x"}]}`,
			field:  "field_key",
			reason: "unsupported_field_key",
		},
		{
			name:   "storage alias is not a field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"summary","value":"x"}]}`,
			field:  "field_key",
			reason: "unsupported_field_key",
		},
		{
			name:   "storage table path is not a field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v1","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline_events.summary","value":"x"}]}`,
			field:  "field_key",
			reason: "unsupported_field_key",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(tc.body))
			if apiErr == nil {
				t.Fatalf("expected validation error for %s", tc.name)
			}
			requireClosedVocabularyRejected(
				t,
				apiErr.Code,
				apiErr.Details,
				tc.field,
				tc.reason,
			)
		})
	}

	t.Run("thirty three raw changes fail closed", func(t *testing.T) {
		rawChanges := make([]map[string]any, 0, maxPatchChanges+1)
		for range maxPatchChanges + 1 {
			rawChanges = append(rawChanges, map[string]any{
				"field_key": "timeline.summary",
				"value":     "summary-value",
			})
		}
		payload, err := json.Marshal(map[string]any{
			"view_schema_id":   TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-u-3-06-too-many",
			"changes":          rawChanges,
		})
		if err != nil {
			t.Fatalf("marshal oversize patch payload: %v", err)
		}

		_, apiErr := DecodeTimelinePatchRequest(bytes.NewReader(payload))
		if apiErr == nil {
			t.Fatal("expected change count rejection")
		}
		requireClosedVocabularyRejected(
			t,
			apiErr.Code,
			apiErr.Details,
			"changes",
			"change_count_exceeded",
		)
		requireErrorDetail(t, apiErr.Details, "requested_count", maxPatchChanges+1)
		requireErrorDetail(t, apiErr.Details, "max_count", maxPatchChanges)
	})

	t.Run("empty patch collection actions fail with collection count detail", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"view_schema_id":   TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-u-3-06-empty-actions",
			"changes": []map[string]any{
				{
					"field_key": "timeline.host_refs",
					"action_payload": map[string]any{
						"kind":    "collection_actions_v1",
						"actions": []map[string]any{},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("marshal empty collection patch payload: %v", err)
		}

		_, apiErr := DecodeTimelinePatchRequest(bytes.NewReader(payload))
		if apiErr == nil {
			t.Fatal("expected empty collection action rejection")
		}
		requireClosedVocabularyRejected(
			t,
			apiErr.Code,
			apiErr.Details,
			"changes.action_payload.actions",
			"empty_collection_actions",
		)
		requireErrorDetail(t, apiErr.Details, "field_key", "timeline.host_refs")
	})

	t.Run("sixty five collection actions fail closed", func(t *testing.T) {
		actions := make([]map[string]any, 0, maxCollectionActions+1)
		for index := range maxCollectionActions + 1 {
			actions = append(actions, map[string]any{
				"op":       "add_token",
				"raw_text": "host-token-" + string(rune('a'+(index%26))),
			})
		}
		payload, err := json.Marshal(map[string]any{
			"view_schema_id":   TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-u-3-06-too-many-actions",
			"changes": []map[string]any{
				{
					"field_key": "timeline.host_refs",
					"action_payload": map[string]any{
						"kind":    "collection_actions_v1",
						"actions": actions,
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("marshal oversized collection patch payload: %v", err)
		}

		_, apiErr := DecodeTimelinePatchRequest(bytes.NewReader(payload))
		if apiErr == nil {
			t.Fatal("expected oversized collection action rejection")
		}
		requireClosedVocabularyRejected(
			t,
			apiErr.Code,
			apiErr.Details,
			"changes.action_payload.actions",
			"collection_action_count_exceeded",
		)
		requireErrorDetail(t, apiErr.Details, "field_key", "timeline.host_refs")
		requireErrorDetail(t, apiErr.Details, "requested_count", maxCollectionActions+1)
		requireErrorDetail(t, apiErr.Details, "max_count", maxCollectionActions)
	})

	t.Run("empty create collection actions fail with collection count detail", func(t *testing.T) {
		payload, err := json.Marshal(map[string]any{
			"client_txn_id": "txn-u-3-06-create-empty-actions",
			"timeline.host_refs": map[string]any{
				"kind":    "collection_actions_v1",
				"actions": []map[string]any{},
			},
		})
		if err != nil {
			t.Fatalf("marshal empty create collection payload: %v", err)
		}

		_, apiErr := DecodeTimelineCreateRequest(bytes.NewReader(payload))
		if apiErr == nil {
			t.Fatal("expected empty create collection action rejection")
		}
		requireClosedVocabularyRejected(
			t,
			apiErr.Code,
			apiErr.Details,
			"timeline.host_refs.actions",
			"empty_collection_actions",
		)
		requireErrorDetail(t, apiErr.Details, "field_key", "timeline.host_refs")
	})

	t.Run("oversized create collection actions fail with collection count detail", func(t *testing.T) {
		actions := make([]map[string]any, 0, maxCollectionActions+1)
		for index := range maxCollectionActions + 1 {
			actions = append(actions, map[string]any{
				"op":       "add_token",
				"raw_text": "host-token-" + string(rune('a'+(index%26))),
			})
		}
		payload, err := json.Marshal(map[string]any{
			"client_txn_id": "txn-u-3-06-create-too-many-actions",
			"timeline.host_refs": map[string]any{
				"kind":    "collection_actions_v1",
				"actions": actions,
			},
		})
		if err != nil {
			t.Fatalf("marshal oversized create collection payload: %v", err)
		}

		_, apiErr := DecodeTimelineCreateRequest(bytes.NewReader(payload))
		if apiErr == nil {
			t.Fatal("expected oversized create collection action rejection")
		}
		requireClosedVocabularyRejected(
			t,
			apiErr.Code,
			apiErr.Details,
			"timeline.host_refs.actions",
			"collection_action_count_exceeded",
		)
		requireErrorDetail(t, apiErr.Details, "field_key", "timeline.host_refs")
		requireErrorDetail(t, apiErr.Details, "requested_count", maxCollectionActions+1)
		requireErrorDetail(t, apiErr.Details, "max_count", maxCollectionActions)
	})
}
