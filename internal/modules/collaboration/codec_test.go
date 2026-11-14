package collaboration

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
)

func TestCollaborationCodecSemanticContract(t *testing.T) {
	t.Run("client payload projects exact required names and ignores additive objects", func(t *testing.T) {
		for _, kind := range []string{"view_schema", "saved_view", "extension_workspace"} {
			sheet := map[string]any{"kind": kind, "future": map[string]any{"nested": []any{1, nil}}, "KIND": false}
			wantSheet := map[string]string{"kind": kind}
			if kind == "extension_workspace" {
				sheet["extension_profile_id"], sheet["workspace_key"] = "network_flow_activity", "network_analysis"
				wantSheet["extension_profile_id"], wantSheet["workspace_key"] = "network_flow_activity", "network_analysis"
			} else {
				sheet["id"], wantSheet["id"] = "surface-1", "surface-1"
			}
			for _, family := range []string{"hello", "resume", "presence_update", "pong"} {
				payload := map[string]any{
					"client_instance_id": "client", "CLIENT_INSTANCE_ID": false,
					"resume_token": "token", "last_seen_stream_seq": 0,
					"presence": map[string]any{"mode": "viewing", "sheet_ref": sheet, "MODE": false, "future": []any{false}},
					"future":   map[string]any{"nested": true},
				}
				got, err := collabprotocol.DecodeClientPayload(collabprotocol.Message{Type: family, Payload: collabtestprotocol.RawPayload(t, payload)})
				if err != nil {
					t.Fatalf("%s/%s: %v", family, kind, err)
				}
				if family != "pong" && (got.Presence.Mode != "viewing" || !reflect.DeepEqual(got.Presence.SheetRef, wantSheet)) {
					t.Fatalf("known presence projection = %#v", got.Presence)
				}
			}
		}
		for _, payload := range []string{
			`{"client_instance_id":"client","resume_token":"token","presence":{"sheet_ref":{"kind":"view_schema","id":"surface"},"mode":"viewing"}}`,
			`{"client_instance_id":"client","resume_token":"token","last_seen_stream_seq":null,"presence":{"sheet_ref":{"kind":"view_schema","id":"surface"},"mode":"viewing"}}`,
			`{"client_instance_id":"client","resume_token":"token","last_seen_stream_seq":-1,"presence":{"sheet_ref":{"kind":"view_schema","id":"surface"},"mode":"viewing"}}`,
			`{"CLIENT_INSTANCE_ID":"client","resume_token":"token","last_seen_stream_seq":0,"presence":{"sheet_ref":{"kind":"view_schema","id":"surface"},"mode":"viewing"}}`,
			`{"client_instance_id":"client","resume_token":"token","last_seen_stream_seq":0,"presence":{"sheet_ref":{"kind":"view_schema","id":"surface"},"mode":"viewing","record_id":null}}`,
			`{"client_instance_id":"client","resume_token":"token","last_seen_stream_seq":0,"presence":{"sheet_ref":{"kind":"view_schema","id":"surface"},"mode":"viewing","field_key":"title"}}`,
			`{"client_instance_id":"client","resume_token":"token","last_seen_stream_seq":0,"presence":{"sheet_ref":{"kind":"extension_workspace","extension_profile_id":"network_flow_activity","workspace_key":"network_analysis","id":null},"mode":"viewing"}}`,
		} {
			if _, err := collabprotocol.DecodeClientPayload(collabprotocol.Message{Type: "resume", Payload: []byte(payload)}); err == nil {
				t.Fatalf("invalid client payload accepted: %s", payload)
			}
		}
	})
	t.Run("field presence requires a materialized row in every client family", func(t *testing.T) {
		for _, family := range []string{"hello", "resume", "presence_update"} {
			for _, kind := range []string{"view_schema", "saved_view"} {
				for _, anchor := range []any{nil, "", "row-1"} {
					presence := map[string]any{"sheet_ref": map[string]string{"kind": kind, "id": "surface"}, "mode": "editing", "field_key": "title"}
					if anchor != nil {
						presence["record_id"] = anchor
					}
					payload := map[string]any{"client_instance_id": "client", "resume_token": "token", "last_seen_stream_seq": 0, "presence": presence}
					got, err := collabprotocol.DecodeClientPayload(collabprotocol.Message{Type: family, Payload: collabtestprotocol.RawPayload(t, payload)})
					if anchor != "row-1" {
						if err == nil {
							t.Fatalf("%s/%s accepted unanchored field: %#v", family, kind, anchor)
						}
					} else if err != nil || got.Presence.RecordID == nil || *got.Presence.RecordID != "row-1" || got.Presence.FieldKey == nil || *got.Presence.FieldKey != "title" {
						t.Fatalf("%s/%s lost valid field anchor: %#v %v", family, kind, got, err)
					}
				}
			}
		}
	})
	incidentID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	t.Run("client envelope admits UTC and additive fields", func(t *testing.T) {
		for _, timestamp := range []string{"2000-02-29T23:59:59Z", "2026-07-13T12:00:00Z", "2026-07-13T12:00:00.123456789+00:00", "2026-07-13T12:00:00-00:00"} {
			for _, messageType := range []string{"hello", "resume", "presence_update", "pong"} {
				data, err := json.Marshal(map[string]any{
					"type": messageType, "incident_id": incidentID.String(), "emitted_at": timestamp,
					"payload": map[string]any{}, "future": map[string]any{"nested": true},
					"TYPE": 42, "INCIDENT_ID": "other", "EMITTED_AT": false, "PAYLOAD": []any{},
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := collabprotocol.DecodeClientMessage(collabprotocol.MessageText, data, incidentID.String())
				if err != nil || got.Type != messageType || got.IncidentID != incidentID.String() || got.EmittedAt != timestamp {
					t.Fatalf("valid client envelope: %#v %v", got, err)
				}
			}
		}
	})
	t.Run("client envelope rejects non RFC3339 UTC grammar and calendar", func(t *testing.T) {
		for _, timestamp := range []string{"2026-07-13T1:00:00Z", "2026-07-13T12:00:00,123Z", "2026-02-29T00:00:00Z", "2026-04-31T00:00:00Z", "2026-07-13T24:00:00Z", "2026-07-13T12:60:00Z", "2026-07-13T12:00:61Z", "2026-07-13T12:00:00+01:00", "2026-07-13T12:00:00", "2026-07-13 12:00:00Z", ""} {
			for _, family := range []string{"hello", "resume", "presence_update", "pong"} {
				raw := collabtestprotocol.RawPayload(t, map[string]any{"type": family, "incident_id": incidentID.String(), "emitted_at": timestamp, "payload": map[string]any{}})
				_, err := collabprotocol.DecodeClientMessage(collabprotocol.MessageText, raw, incidentID.String())
				var failure *collabprotocol.DecodeFailure
				if !errors.As(err, &failure) || failure.Kind != collabprotocol.DecodeFailureInvalidEnvelope {
					t.Fatalf("%s/%s accepted or misclassified: %v", family, timestamp, err)
				}
			}
		}
	})
	message := collabprotocol.EphemeralMessage(incidentID, "resume_ack", map[string]any{
		"status":       collabprotocol.ResumeStatusReplayed,
		"resume_token": "token",
	}, time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC))

	encoded, err := collabprotocol.EncodeMessage(message)
	if err != nil {
		t.Fatalf("encode message: %v", err)
	}
	if bytes.HasSuffix(encoded, []byte("\n")) {
		t.Fatalf("encoded message has a trailing LF: %q", encoded)
	}
	decoded, err := collabprotocol.DecodeMessage(collabprotocol.MessageText, appendUnknownMembers(encoded))
	if err != nil {
		t.Fatalf("decode message with additive members: %v", err)
	}
	if decoded.Type != "resume_ack" {
		t.Fatalf("decoded type = %q", decoded.Type)
	}
	const prefix = `{"type":"pong","payload":{},"future":"`
	const suffix = `"}`
	maximum := append([]byte(prefix), bytes.Repeat([]byte("x"), 32768-len(prefix)-len(suffix))...)
	maximum = append(maximum, suffix...)
	decoded, err = collabprotocol.DecodeMessage(collabprotocol.MessageText, maximum)
	if err != nil || decoded.Type != "pong" {
		t.Fatalf("32768-byte additive message rejected: type=%q error=%v", decoded.Type, err)
	}

	for name, testCase := range map[string]struct {
		kind     collabprotocol.MessageKind
		payload  []byte
		wantKind collabprotocol.DecodeFailureKind
		wantSize bool
	}{
		"binary": {
			kind:     collabprotocol.MessageBinary,
			payload:  []byte(`{"type":"pong","payload":{}}`),
			wantKind: collabprotocol.DecodeFailureBinaryMessage,
		},
		"malformed": {
			kind:     collabprotocol.MessageText,
			payload:  []byte(`{"type":`),
			wantKind: collabprotocol.DecodeFailureInvalidJSON,
		},
		"invalid UTF-8": {
			kind:     collabprotocol.MessageText,
			payload:  []byte{0xff},
			wantKind: collabprotocol.DecodeFailureInvalidJSON,
		},
		"duplicate envelope member": {
			kind:     collabprotocol.MessageText,
			payload:  []byte(`{"type":"pong","type":"pong","payload":{}}`),
			wantKind: collabprotocol.DecodeFailureDuplicateMember,
		},
		"duplicate nested member": {
			kind:     collabprotocol.MessageText,
			payload:  []byte(`{"type":"pong","payload":{"value":1,"value":2}}`),
			wantKind: collabprotocol.DecodeFailureDuplicateMember,
		},
		"valid JSON array": {
			kind: collabprotocol.MessageText, payload: []byte(`[]`), wantKind: collabprotocol.DecodeFailureInvalidEnvelope,
		},
		"valid JSON scalar": {
			kind: collabprotocol.MessageText, payload: []byte(`42`), wantKind: collabprotocol.DecodeFailureInvalidEnvelope,
		},
		"invalid type member": {
			kind: collabprotocol.MessageText, payload: []byte(`{"type":42,"payload":{}}`), wantKind: collabprotocol.DecodeFailureInvalidEnvelope,
		},
		"invalid incident member": {
			kind: collabprotocol.MessageText, payload: []byte(`{"type":"hello","incident_id":false,"payload":{}}`), wantKind: collabprotocol.DecodeFailureInvalidEnvelope,
		},
		"oversized": {
			kind:     collabprotocol.MessageText,
			payload:  append(maximum, ' '),
			wantSize: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := collabprotocol.DecodeMessage(testCase.kind, testCase.payload)
			if testCase.wantSize {
				if !errors.Is(err, collabprotocol.ErrMessageTooLarge) {
					t.Fatalf("error = %v want collabprotocol.ErrMessageTooLarge", err)
				}
				return
			}
			var failure *collabprotocol.DecodeFailure
			if !errors.As(err, &failure) || failure.Kind != testCase.wantKind {
				t.Fatalf("failure = %#v / %v want kind %d", failure, err, testCase.wantKind)
			}
		})
	}

	if got := safeWebSocketEventType("extension_resource_changed"); got != "extension_resource_changed" {
		t.Fatalf("extension event classification = %q", got)
	}
	if got := safeWebSocketEventType("resume_result"); got != "other" {
		t.Fatalf("non-public resume_result classification = %q", got)
	}
}

func appendUnknownMembers(encoded []byte) []byte {
	if len(encoded) == 0 || encoded[len(encoded)-1] != '}' {
		return encoded
	}
	output := append([]byte(nil), encoded[:len(encoded)-1]...)
	output = append(output, []byte(`,"future_envelope_member":{"nested":true}}`)...)
	return output
}
