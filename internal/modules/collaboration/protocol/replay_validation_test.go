package protocol_test

import (
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

	for _, test := range tests {
		t.Run(test.family, func(t *testing.T) {
			payload := collabtestprotocol.RawPayload(test.payload)
			if err := protocol.ValidateReplayablePayload(incidentID, test.family, payload); err != nil {
				t.Fatalf("validate payload: %v", err)
			}
			message := collabtestprotocol.SequencedMessage(
				test.family,
				incidentID,
				"opaque:event/identity",
				1,
				now,
				test.payload,
			)
			if err := protocol.ValidateSequencedReplayableMessage(message); err != nil {
				t.Fatalf("validate sequenced message: %v", err)
			}
		})
	}
}

func TestReplayValidationFailsClosed(t *testing.T) {
	incidentID := uuid.New()
	now := time.Date(2026, time.August, 28, 18, 0, 0, 0, time.UTC)
	validJob := collabtestprotocol.NewIncidentJobProgressPayload(
		"job-validation",
		incidentID,
		protocol.JobStatusRunning,
		protocol.JobProgress{},
		now,
	)
	validMessage := collabtestprotocol.SequencedMessage(
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
