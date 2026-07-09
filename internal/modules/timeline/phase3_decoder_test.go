package timeline

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPhase3_PatchPayloadValidation_U_3_06(t *testing.T) {
	request, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(`{
		"view_schema_id": "cartulary.view.timeline.v2",
		"base_row_version": 2,
		"client_txn_id": "txn-u-3-06",
		"changes": [
			{ "field_key": "timeline.activity_synopsis_text", "value": "B" },
			{ "field_key": "timeline.raw_activity_text", "value": "A" }
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
		"timeline.raw_activity_text",
		"timeline.host_refs",
		"timeline.identity_refs",
		"timeline.activity_utc_text",
		"timeline.raw_activity_text",
		"timeline.activity_synopsis_text",
		"timeline.tags",
	})

	t.Run("thirty two raw changes do not trip the count ceiling", func(t *testing.T) {
		rawChanges := make([]map[string]any, 0, maxPatchChanges)
		for range maxPatchChanges {
			rawChanges = append(rawChanges, map[string]any{
				"field_key": "timeline.activity_synopsis_text",
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
			"view_schema_id": "cartulary.view.timeline.v2",
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
			"view_schema_id": "cartulary.view.timeline.v2",
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

	t.Run("registry-disallowed collection ops are rejected by family", func(t *testing.T) {
		stableID := "00000000-0000-0000-0000-000000000001"
		cases := []struct {
			name     string
			fieldKey string
			action   string
		}{
			{
				name:     "host refs reject tag op",
				fieldKey: "timeline.host_refs",
				action:   `{"op":"add_tag","tag_name":"wrong-family"}`,
			},
			{
				name:     "tags reject mention token op",
				fieldKey: "timeline.tags",
				action:   `{"op":"add_token","raw_text":"host01"}`,
			},
			{
				name:     "attached evidence rejects mention token op",
				fieldKey: "timeline.attached_evidence_ids",
				action:   `{"op":"add_token","raw_text":"artifact"}`,
			},
			{
				name:     "tags reject attached evidence op",
				fieldKey: "timeline.tags",
				action:   `{"op":"add_record_ref","linked_record_id":"` + stableID + `"}`,
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				body := `{
					"view_schema_id": "cartulary.view.timeline.v2",
					"base_row_version": 1,
					"client_txn_id": "txn-registry-disallowed",
					"changes": [
						{
							"field_key": "` + tc.fieldKey + `",
							"action_payload": {
								"kind": "collection_actions_v1",
								"actions": [` + tc.action + `]
							}
						}
					]
				}`
				_, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(body))
				if apiErr == nil {
					t.Fatalf("expected registry-disallowed op to fail for %s", tc.fieldKey)
				}
				requireClosedVocabularyRejected(t, apiErr.Code, apiErr.Details, tc.fieldKey, "invalid_value")
			})
		}
	})

	t.Run("strict JSON object envelope rejects ambiguous mutation bodies", func(t *testing.T) {
		cases := []struct {
			name string
			body string
		}{
			{
				name: "top-level non-object",
				body: `[]`,
			},
			{
				name: "trailing object",
				body: `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn-strict-trailing-object","changes":[{"field_key":"timeline.activity_synopsis_text","value":"x"}]} {"extra":true}`,
			},
			{
				name: "trailing scalar",
				body: `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn-strict-trailing-scalar","changes":[{"field_key":"timeline.activity_synopsis_text","value":"x"}]} true`,
			},
			{
				name: "duplicate top-level key",
				body: `{"view_schema_id":"cartulary.view.timeline.v2","view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn-strict-duplicate-top","changes":[{"field_key":"timeline.activity_synopsis_text","value":"x"}]}`,
			},
			{
				name: "duplicate nested action key",
				body: `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn-strict-duplicate-nested","changes":[{"field_key":"timeline.host_refs","action_payload":{"kind":"collection_actions_v1","actions":[{"op":"add_token","op":"add_token","raw_text":"WS-023"}]}}]}`,
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(tc.body))
				if apiErr == nil {
					t.Fatalf("expected strict JSON rejection for %s", tc.name)
				}
				if apiErr.Code != "invalid_mutation_payload" {
					t.Fatalf("unexpected rejection code: %q", apiErr.Code)
				}
				if _, ok := apiErr.Details["field"]; ok {
					t.Fatalf("unexpected field detail for envelope rejection: %v", apiErr.Details["field"])
				}
				if apiErr.Details["reason_code"] != "request_not_object" {
					t.Fatalf("unexpected rejection reason_code: got %v want %q", apiErr.Details["reason_code"], "request_not_object")
				}
			})
		}
	})

	t.Run("strict JSON object envelope permits whitespace suffix", func(t *testing.T) {
		body := `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn-strict-whitespace","changes":[{"field_key":"timeline.activity_synopsis_text","value":"x"}]}` + " \n\t "
		_, apiErr := DecodeTimelinePatchRequest(bytes.NewBufferString(body))
		if apiErr != nil {
			t.Fatalf("expected whitespace suffix to decode, got %#v", apiErr)
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
			body:   `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.activity_synopsis_text","value":"x"}],"unknown":"value"}`,
			field:  "unknown",
			reason: "unknown_field",
		},
		{
			name:   "missing view schema",
			body:   `{"base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.activity_synopsis_text","value":"x"}]}`,
			field:  "view_schema_id",
			reason: "missing_required_field",
		},
		{
			name:   "missing base row version",
			body:   `{"view_schema_id":"cartulary.view.timeline.v2","client_txn_id":"txn","changes":[{"field_key":"timeline.activity_synopsis_text","value":"x"}]}`,
			field:  "base_row_version",
			reason: "missing_required_field",
		},
		{
			name:   "missing client txn",
			body:   `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"changes":[{"field_key":"timeline.activity_synopsis_text","value":"x"}]}`,
			field:  "client_txn_id",
			reason: "missing_required_field",
		},
		{
			name:   "empty changes",
			body:   `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn","changes":[]}`,
			field:  "changes",
			reason: "empty_changes",
		},
		{
			name:   "duplicate field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.activity_synopsis_text","value":"x"},{"field_key":"timeline.activity_synopsis_text","value":"y"}]}`,
			field:  "changes",
			reason: "duplicate_field_key",
		},
		{
			name:   "readonly system field",
			body:   `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline.capture_state","value":"reviewed"}]}`,
			field:  "field_key",
			reason: "unsupported_field_key",
		},
		{
			name:   "visible label is not a field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"Summary","value":"x"}]}`,
			field:  "field_key",
			reason: "unsupported_field_key",
		},
		{
			name:   "storage alias is not a field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"summary","value":"x"}]}`,
			field:  "field_key",
			reason: "unsupported_field_key",
		},
		{
			name:   "storage table path is not a field key",
			body:   `{"view_schema_id":"cartulary.view.timeline.v2","base_row_version":1,"client_txn_id":"txn","changes":[{"field_key":"timeline_events.summary","value":"x"}]}`,
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
				"field_key": "timeline.activity_synopsis_text",
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

func TestSupportPhase3Unit_TimelineVisibleTextContract(t *testing.T) {
	t.Run("create preserves nullable empty and source-like text", func(t *testing.T) {
		sourceLike := `=HYPERLINK("https://example.test","click") <script>alert(1)</script> **bold** [link](https://example.test)`
		exactWhitespace := " \tTabbed\nLine\rCarriage "
		maxLength := strings.Repeat("a", 32768)
		payload, err := json.Marshal(map[string]any{
			"client_txn_id":                   "txn-u-3-12-create-visible-text",
			"timeline.activity_synopsis_text": "",
			"timeline.raw_activity_text":      exactWhitespace,
			"timeline.data_source_text":       sourceLike,
			"timeline.analyst_text":           nil,
			"timeline.device_object_text":     maxLength,
		})
		if err != nil {
			t.Fatalf("marshal visible-text create payload: %v", err)
		}

		request, apiErr := DecodeTimelineCreateRequest(bytes.NewReader(payload))
		if apiErr != nil {
			t.Fatalf("expected visible-text create payload to decode, got %#v", apiErr)
		}
		if request.ActivitySynopsisText == nil || *request.ActivitySynopsisText != "" {
			t.Fatalf("expected empty string to remain distinct from null, got %#v", request.ActivitySynopsisText)
		}
		if request.RawActivityText == nil || *request.RawActivityText != exactWhitespace {
			t.Fatalf("expected exact whitespace preservation, got %#v", request.RawActivityText)
		}
		if request.DataSourceText == nil || *request.DataSourceText != sourceLike {
			t.Fatalf("expected source-like text preservation, got %#v", request.DataSourceText)
		}
		if request.AnalystText != nil {
			t.Fatalf("expected explicit null to remain nil, got %#v", request.AnalystText)
		}
		if request.DeviceObjectText == nil || *request.DeviceObjectText != maxLength {
			t.Fatalf("expected max-length visible text to decode, got %#v", request.DeviceObjectText)
		}
	})

	t.Run("patch preserves source-like text", func(t *testing.T) {
		sourceLike := `=HYPERLINK("https://example.test","click") <b>host</b> _markdown_`
		payload, err := json.Marshal(map[string]any{
			"view_schema_id":   TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-u-3-12-patch-visible-text",
			"changes": []map[string]any{
				{"field_key": "timeline.raw_activity_text", "value": sourceLike},
			},
		})
		if err != nil {
			t.Fatalf("marshal visible-text patch payload: %v", err)
		}

		request, apiErr := DecodeTimelinePatchRequest(bytes.NewReader(payload))
		if apiErr != nil {
			t.Fatalf("expected visible-text patch payload to decode, got %#v", apiErr)
		}
		if len(request.CanonicalChange) != 1 || request.CanonicalChange[0].TextValue == nil || *request.CanonicalChange[0].TextValue != sourceLike {
			t.Fatalf("expected source-like patch text preservation, got %#v", request.CanonicalChange)
		}
	})

	t.Run("invalid controls and oversized values fail closed", func(t *testing.T) {
		cases := []struct {
			name  string
			value string
		}{
			{name: "nul", value: "bad\x00value"},
			{name: "control", value: "bad\x01value"},
			{name: "c1 control", value: "bad\u0085value"},
			{name: "too long", value: strings.Repeat("a", 32769)},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				payload, err := json.Marshal(map[string]any{
					"client_txn_id":                   "txn-u-3-12-invalid-" + tc.name,
					"timeline.activity_synopsis_text": tc.value,
				})
				if err != nil {
					t.Fatalf("marshal invalid visible-text payload: %v", err)
				}

				_, apiErr := DecodeTimelineCreateRequest(bytes.NewReader(payload))
				if apiErr == nil {
					t.Fatal("expected invalid visible text to fail closed")
				}
				requireClosedVocabularyRejected(
					t,
					apiErr.Code,
					apiErr.Details,
					"timeline.activity_synopsis_text",
					"invalid_value",
				)
			})
		}
	})
}
