package protocol_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
)

func TestReplayValidationCoversEveryFamilyAndAcceptsAdditiveMembers(t *testing.T) {
	incidentID := uuid.New()
	now := time.Date(2026, time.August, 28, 18, 0, 0, 0, time.UTC)
	tests := []struct {
		family  string
		payload any
	}{
		{
			family: "record_changed",
			payload: map[string]any{
				"record_id": uuid.NewString(), "row_version": 2,
				"change_set_id": uuid.NewString(), "client_txn_id": "txn-additive",
				"actor_user_id": uuid.NewString(), "changed_field_keys": []string{"timeline.title"},
				"affected_views": []map[string]any{{
					"view_schema_id": "cartulary.view.timeline.v2", "change_kind": "invalidate",
					"future_view_member": true,
				}},
				"future_payload_member": map[string]any{"accepted": true},
			},
		},
		{
			family: "job_progress",
			payload: map[string]any{
				"job_id": "job-additive",
				"scope": map[string]any{
					"kind": protocol.JobScopeKindIncident, "incident_id": incidentID.String(),
					"future_scope_member": true,
				},
				"status":     protocol.JobStatusRunning,
				"progress":   map[string]any{"completed": 1, "total": 2, "future_progress_member": true},
				"updated_at": now, "future_payload_member": true,
			},
		},
		{
			family: "extension_resource_changed",
			payload: map[string]any{
				"extension_profile_id": "network_flow_activity",
				"resource_kind":        "network_flow_table", "resource_id": "nft_additive",
				"change_kind":           protocol.ExtensionResourceChangeKindInvalidate,
				"reason_code":           protocol.ExtensionResourceReasonRenamed,
				"future_payload_member": true,
			},
		},
	}

	for _, reason := range []string{"created", "refresh_requested", "materialized", "materialization_failed", "source_invalidated"} {
		payload := protocol.ExtensionResourceChangePayload{ExtensionProfileID: "network_flow_activity", ResourceKind: "network_flow_graph_view", ResourceID: "nfgv_additive", ChangeKind: protocol.ExtensionResourceChangeKindInvalidate, ReasonCode: reason}
		if err := protocol.ValidateExtensionResourceChangePayload(payload); err != nil {
			t.Fatalf("additive invalidation %s: %v", reason, err)
		}
		payload.ChangeKind = protocol.ExtensionResourceChangeKindRemove
		if err := protocol.ValidateExtensionResourceChangePayload(payload); err == nil {
			t.Fatalf("additive reason %s admitted removal", reason)
		}
	}
	for _, test := range tests {
		t.Run(test.family, func(t *testing.T) {
			payload := collabtestprotocol.RawPayload(t, test.payload)
			if err := protocol.ValidateReplayablePayload(incidentID, test.family, payload); err != nil {
				t.Fatalf("validate payload: %v", err)
			}
			message := collabtestprotocol.SequencedMessage(t,
				test.family,
				incidentID,
				"opaque:event/identity",
				1,
				now,
				test.payload,
			)
			for _, timestamp := range []string{"2000-02-29T23:59:59Z", "2026-08-28T18:00:00.123456789+00:00", "2026-08-28T18:00:00-00:00"} {
				message.EmittedAt = timestamp
				if err := protocol.ValidateSequencedReplayableMessage(message); err != nil {
					t.Fatalf("validate sequenced timestamp %s: %v", timestamp, err)
				}
			}
		})
	}
}

func TestReplayValidationFailsClosed(t *testing.T) {
	incidentID := uuid.New()
	now := time.Date(2026, time.August, 28, 18, 0, 0, 0, time.UTC)
	t.Run("durable wire shape and exact semantic names", func(t *testing.T) {
		recordID := uuid.NewString()
		fixtures := []struct {
			family   string
			payload  map[string]any
			required []string
		}{
			{"record_changed", map[string]any{
				"record_id": recordID, "row_version": 2, "change_set_id": uuid.NewString(), "client_txn_id": "txn-shape", "actor_user_id": uuid.NewString(),
				"changed_field_keys": []any{"timeline.title"}, "affected_views": []any{map[string]any{"view_schema_id": "cartulary.view.timeline.v2", "change_kind": "patch", "patch_cells": map[string]any{"record_id": recordID, "row_version": 2, "cells": map[string]any{"timeline.title": map[string]any{"value": nil}}, "group_values": map[string]any{"timeline.title": "group"}}}},
			}, []string{"record_id", "row_version", "change_set_id", "client_txn_id", "actor_user_id", "changed_field_keys", "affected_views"}},
			{"job_progress", map[string]any{
				"job_id": "job-shape", "scope": map[string]any{"kind": "incident", "incident_id": incidentID.String()}, "status": "running", "progress": map[string]any{"completed": 0, "total": nil}, "updated_at": now.Format(time.RFC3339Nano),
				"result_summary": map[string]any{"code": "result", "message": "", "resource_refs": []any{map[string]any{"kind": "report", "id": "report-1", "route": "/api/v1/reports/report-1"}}}, "error_summary": map[string]any{"code": "error", "message": "", "retryable": false, "details": map[string]any{}},
			}, []string{"job_id", "scope", "status", "progress", "updated_at"}},
			{"extension_resource_changed", map[string]any{
				"extension_profile_id": "network_flow_activity", "resource_kind": "network_flow_table", "resource_id": "nft-shape", "change_kind": "invalidate", "reason_code": "renamed", "workspace_refs": []any{map[string]any{"kind": "extension_workspace", "extension_profile_id": "network_flow_activity", "workspace_key": "network_analysis"}},
			}, []string{"extension_profile_id", "resource_kind", "resource_id", "change_kind", "reason_code"}},
		}
		for _, fixture := range fixtures {
			t.Run(fixture.family, func(t *testing.T) {
				copyFixture := func() map[string]any {
					var copy map[string]any
					if err := json.Unmarshal(collabtestprotocol.RawPayload(t, fixture.payload), &copy); err != nil {
						t.Fatal(err)
					}
					return copy
				}
				check := func(label string, payload any, valid bool) {
					t.Helper()
					raw := collabtestprotocol.RawPayload(t, payload)
					seq := int64(1)
					message := protocol.Message{Type: fixture.family, IncidentID: incidentID.String(), EventID: "event-shape", StreamSeq: &seq, EmittedAt: now.Format(time.RFC3339Nano), Payload: raw}
					for boundary, err := range map[string]error{"admission": protocol.ValidateReplayablePayload(incidentID, fixture.family, raw), "replay": protocol.ValidateSequencedReplayableMessage(message)} {
						if (err == nil) != valid {
							t.Fatalf("%s/%s validity = %v want %v: %v", label, boundary, err == nil, valid, err)
						}
						if err != nil && strings.Contains(err.Error(), "private-invalid-value") {
							t.Fatalf("%s leaked payload in diagnostic", boundary)
						}
					}
				}
				check("valid", fixture.payload, true)
				for _, key := range fixture.required {
					for _, mode := range []string{"absent", "null", "case-only", "case-shadow"} {
						copy := copyFixture()
						switch mode {
						case "absent":
							delete(copy, key)
						case "null":
							copy[key] = nil
						case "case-only":
							copy[strings.ToUpper(key)] = copy[key]
							delete(copy, key)
						case "case-shadow":
							copy[strings.ToUpper(key)] = map[string]any{"private-invalid-value": true}
						}
						check(key+"/"+mode, copy, mode == "case-shadow")
					}
				}
				type nestedObject struct {
					label    string
					path     []any
					required []string
					nullable map[string]bool
				}
				var nested []nestedObject
				switch fixture.family {
				case "record_changed":
					nested = append(nested, nestedObject{"view", []any{"affected_views", 0}, []string{"view_schema_id", "change_kind", "patch_cells"}, nil})
					nested = append(nested, nestedObject{"patch", []any{"affected_views", 0, "patch_cells"}, []string{"record_id", "row_version", "cells"}, nil})
					nested = append(nested, nestedObject{"cell", []any{"affected_views", 0, "patch_cells", "cells", "timeline.title"}, []string{"value"}, map[string]bool{"value": true}})
				case "job_progress":
					for label, required := range map[string][]string{"scope": {"kind", "incident_id"}, "progress": {"completed", "total"}, "result_summary": {"code", "message"}, "error_summary": {"code", "message", "retryable"}} {
						nested = append(nested, nestedObject{label, []any{label}, required, map[string]bool{"total": true}})
					}
					nested = append(nested, nestedObject{"resource", []any{"result_summary", "resource_refs", 0}, []string{"kind", "id"}, nil})
				case "extension_resource_changed":
					nested = append(nested, nestedObject{"workspace", []any{"workspace_refs", 0}, []string{"kind", "extension_profile_id", "workspace_key"}, nil})
				}
				for _, target := range nested {
					for _, key := range target.required {
						for _, mode := range []string{"absent", "null", "case-only", "case-shadow"} {
							copy := copyFixture()
							var object any = copy
							for _, segment := range target.path {
								switch segment := segment.(type) {
								case string:
									object = object.(map[string]any)[segment]
								case int:
									object = object.([]any)[segment]
								}
							}
							members := object.(map[string]any)
							switch mode {
							case "absent":
								delete(members, key)
							case "null":
								members[key] = nil
							case "case-only":
								members[strings.ToUpper(key)] = members[key]
								delete(members, key)
							case "case-shadow":
								members[strings.ToUpper(key)] = map[string]any{"private-invalid-value": true}
							}
							check(target.label+"/"+key+"/"+mode, copy, mode == "case-shadow" || (mode == "null" && target.nullable[key]))
						}
					}
				}
				for key, value := range map[string]any{"future_object": map[string]any{"nested": []any{nil, false}}, "future_scalar": 1} {
					copy := copyFixture()
					copy[key] = value
					check("additive/"+key, copy, true)
				}
				raw := collabtestprotocol.RawPayload(t, fixture.payload)
				duplicate := append([]byte(`{"future":1,"future":2,`), raw[1:]...)
				if err := protocol.ValidateReplayablePayload(incidentID, fixture.family, duplicate); err == nil {
					t.Fatal("duplicate durable object member accepted")
				}
				if fixture.family == "job_progress" {
					for key, value := range map[string]any{"cancelable": nil, "message": 1, "result_summary": "private-invalid-value", "error_summary": false, "retained_until": "2026-08-28T19:00:00+01:00", "updated_at": "2026-02-30T00:00:00Z"} {
						copy := copyFixture()
						copy[key] = value
						check("optional-or-date/"+key, copy, false)
					}
					copy := copyFixture()
					copy["result_summary"] = nil
					copy["error_summary"] = nil
					copy["retained_until"] = nil
					check("nullable summaries and retention", copy, true)
					copy = copyFixture()
					copy["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["route"] = "https://private-invalid-value.test/"
					check("foreign result route", copy, false)
					copy = copyFixture()
					copy["error_summary"].(map[string]any)["details"] = []any{}
					check("error details object", copy, false)
				}
				if fixture.family == "extension_resource_changed" {
					copy := copyFixture()
					copy["workspace_refs"] = nil
					check("null workspace refs", copy, false)
				}
				if fixture.family == "record_changed" {
					for _, mode := range []string{"other-record", "other-version", "unchanged-cell", "unchanged-group", "nonscalar-group"} {
						copy := copyFixture()
						patch := copy["affected_views"].([]any)[0].(map[string]any)["patch_cells"].(map[string]any)
						switch mode {
						case "other-record":
							patch["record_id"] = uuid.NewString()
						case "other-version":
							patch["row_version"] = 3
						case "unchanged-cell":
							patch["cells"].(map[string]any)["timeline.other"] = map[string]any{"value": "unexpected"}
						case "unchanged-group":
							patch["group_values"].(map[string]any)["timeline.other"] = "unexpected"
						case "nonscalar-group":
							patch["group_values"].(map[string]any)["timeline.title"] = []any{"unexpected"}
						}
						check(mode, copy, false)
					}
				}
			})
		}
	})
	validJob := collabtestprotocol.NewIncidentJobProgressPayload(
		"job-validation",
		incidentID,
		protocol.JobStatusRunning,
		protocol.JobProgress{},
		now,
	)
	validMessage := collabtestprotocol.SequencedMessage(t,
		"job_progress", incidentID, "opaque-event", 1, now, validJob,
	)
	tests := []struct {
		name   string
		mutate func(*protocol.Message)
	}{
		{name: "unknown family", mutate: func(message *protocol.Message) { message.Type = "future_family" }},
		{name: "invalid incident", mutate: func(message *protocol.Message) { message.IncidentID = "invalid" }},
		{name: "missing event identity", mutate: func(message *protocol.Message) { message.EventID = "" }},
		{name: "missing sequence", mutate: func(message *protocol.Message) { message.StreamSeq = nil }},
		{name: "nonpositive sequence", mutate: func(message *protocol.Message) { sequence := int64(0); message.StreamSeq = &sequence }},
		{name: "comma timestamp fraction", mutate: func(message *protocol.Message) { message.EmittedAt = "2026-08-28T18:00:00,123Z" }},
		{name: "one digit timestamp hour", mutate: func(message *protocol.Message) { message.EmittedAt = "2026-08-28T1:00:00Z" }},
		{name: "invalid calendar date", mutate: func(message *protocol.Message) { message.EmittedAt = "2026-02-29T00:00:00Z" }},
		{name: "zero timestamp", mutate: func(message *protocol.Message) { message.EmittedAt = "0001-01-01T00:00:00Z" }},
		{name: "invalid timestamp", mutate: func(message *protocol.Message) { message.EmittedAt = "invalid" }},
		{name: "non UTC timestamp", mutate: func(message *protocol.Message) { message.EmittedAt = "2026-08-28T14:00:00-04:00" }},
		{name: "nonobject payload", mutate: func(message *protocol.Message) { message.Payload = []byte(`[]`) }},
		{name: "invalid family payload", mutate: func(message *protocol.Message) { message.Payload = []byte(`{"job_id":"incomplete"}`) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := validMessage
			test.mutate(&message)
			if err := protocol.ValidateSequencedReplayableMessage(message); err == nil {
				t.Fatalf("invalid sequenced message was accepted: %#v", message)
			}
		})
	}
}
