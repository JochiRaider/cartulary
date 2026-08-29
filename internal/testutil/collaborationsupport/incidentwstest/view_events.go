package incidentwstest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

type RecordChangeSocketPayload struct {
	RecordID         string   `json:"record_id"`
	RowVersion       float64  `json:"row_version"`
	ChangeSetID      string   `json:"change_set_id"`
	ClientTxnID      string   `json:"client_txn_id"`
	ActorUserID      string   `json:"actor_user_id"`
	ChangedFieldKeys []string `json:"changed_field_keys"`
	AffectedViews    []struct {
		ViewSchemaID string         `json:"view_schema_id"`
		ChangeKind   string         `json:"change_kind"`
		PatchCells   map[string]any `json:"patch_cells,omitempty"`
	} `json:"affected_views"`
}

type ExtensionResourceChangeExpectation struct {
	IncidentID         string
	ExtensionProfileID string
	ResourceKind       string
	ResourceID         string
	ChangeKind         string
	ReasonCode         string
	ForbiddenKeys      []string
}

func ConnectViewSocket(
	t testing.TB,
	server *httptestx.Server,
	incidentID string,
	viewSchemaID string,
	sessionToken string,
) *Client {
	t.Helper()

	return ConnectAndHello(
		t,
		server.HTTP.URL,
		incidentID,
		ConnectOptions{
			SessionToken:     sessionToken,
			ClientInstanceID: "view-event-test-" + incidentID,
			Presence: platformws.PresenceInput{
				SheetRef: map[string]string{
					"kind": "view_schema",
					"id":   viewSchemaID,
				},
				Mode: "viewing",
			},
		},
	)
}

func ConnectExtensionWorkspaceSocket(
	t testing.TB,
	server *httptestx.Server,
	incidentID string,
	extensionProfileID string,
	workspaceKey string,
	sessionToken string,
) *Client {
	t.Helper()

	return ConnectAndHello(
		t,
		server.HTTP.URL,
		incidentID,
		ConnectOptions{
			SessionToken:     sessionToken,
			ClientInstanceID: "extension-event-test-" + incidentID,
			Presence: platformws.PresenceInput{
				SheetRef: map[string]string{
					"kind":                 "extension_workspace",
					"extension_profile_id": extensionProfileID,
					"workspace_key":        workspaceKey,
				},
				Mode: "viewing",
			},
		},
	)
}

func RequireRecordChangedEvent(
	t testing.TB,
	client *Client,
	wantRecordID uuid.UUID,
	wantRowVersion int64,
) collabtestprotocol.RecordChangedEvent {
	t.Helper()

	var lastChange *collabtestprotocol.RecordChangedEvent
	deadline := time.Now().Add(5 * time.Second)
	for {
		message, err := client.AwaitNextMessage(time.Until(deadline))
		if err != nil {
			if lastChange != nil {
				t.Fatalf("timed out waiting for websocket record_changed record=%s version=%d after seeing %#v: %v", wantRecordID, wantRowVersion, *lastChange, err)
			}
			t.Fatalf("receive websocket record_changed: %v", err)
		}
		if message.Type != "record_changed" {
			continue
		}
		change, err := collabtestprotocol.RecordChangeFromSequencedMessage(message)
		if err != nil {
			t.Fatalf("decode record_changed message: %v", err)
		}
		lastChange = &change
		if change.RecordID == wantRecordID && change.RowVersion == wantRowVersion {
			return change
		}
	}
}

func RequireRecordChangedEvents(
	t testing.TB,
	client *Client,
	want map[uuid.UUID]int64,
) []collabtestprotocol.RecordChangedEvent {
	t.Helper()

	remaining := make(map[uuid.UUID]int64, len(want))
	for recordID, rowVersion := range want {
		remaining[recordID] = rowVersion
	}
	result := make([]collabtestprotocol.RecordChangedEvent, 0, len(want))
	deadline := time.Now().Add(5 * time.Second)
	for len(remaining) > 0 {
		message, err := client.AwaitNextMessage(time.Until(deadline))
		if err != nil {
			t.Fatalf("receive websocket record_changed set; remaining=%v: %v", remaining, err)
		}
		if message.Type != "record_changed" {
			continue
		}
		change, err := collabtestprotocol.RecordChangeFromSequencedMessage(message)
		if err != nil {
			t.Fatalf("decode record_changed message: %v", err)
		}
		rowVersion, expected := remaining[change.RecordID]
		if !expected {
			continue
		}
		if change.RowVersion != rowVersion {
			t.Fatalf("record_changed record=%s row_version=%d want %d", change.RecordID, change.RowVersion, rowVersion)
		}
		delete(remaining, change.RecordID)
		result = append(result, change)
	}
	return result
}

func RequireRecordChanged(
	t testing.TB,
	client *Client,
	wantRecordID string,
	wantRowVersion int64,
) RecordChangeSocketPayload {
	t.Helper()

	recordID, err := uuid.Parse(wantRecordID)
	if err != nil {
		t.Fatalf("parse expected record_changed record_id: %v", err)
	}
	change := RequireRecordChangedEvent(t, client, recordID, wantRowVersion)
	payloadJSON, err := json.Marshal(collabtestprotocol.RecordChangePayload(change))
	if err != nil {
		t.Fatalf("encode semantic record_changed payload: %v", err)
	}
	var payload RecordChangeSocketPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		t.Fatalf("decode semantic record_changed payload: %v", err)
	}
	return payload
}

func RequireExtensionResourceChanged(
	t testing.TB,
	client *Client,
	want ExtensionResourceChangeExpectation,
) platformws.ExtensionResourceChangePayload {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for {
		message, err := client.AwaitNextMessage(time.Until(deadline))
		if err != nil {
			t.Fatalf("receive websocket extension_resource_changed: %v", err)
		}
		if message.Type != "extension_resource_changed" {
			continue
		}
		var payload platformws.ExtensionResourceChangePayload
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			t.Fatalf("decode extension_resource_changed payload: %v", err)
		}
		if payload.ResourceID != want.ResourceID {
			continue
		}
		if message.IncidentID != want.IncidentID || payload.ExtensionProfileID != want.ExtensionProfileID ||
			payload.ResourceKind != want.ResourceKind || payload.ChangeKind != want.ChangeKind || payload.ReasonCode != want.ReasonCode {
			t.Fatalf("unexpected extension_resource_changed message: message=%#v payload=%#v want=%#v", message, payload, want)
		}
		var raw map[string]any
		if err := json.Unmarshal(message.Payload, &raw); err != nil {
			t.Fatalf("decode extension_resource_changed member map: %v", err)
		}
		for _, key := range want.ForbiddenKeys {
			if _, exists := raw[key]; exists {
				t.Fatalf("extension_resource_changed payload disclosed forbidden member %q: %#v", key, raw)
			}
		}
		return payload
	}
}

func ExpectNoMatchingMessage(
	t testing.TB,
	client *Client,
	timeout time.Duration,
	matches func(platformws.Message) bool,
) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return
		}
		message, err := client.AwaitNextMessage(remaining)
		if errors.Is(err, context.DeadlineExceeded) {
			return
		}
		if err != nil {
			t.Fatalf("receive filtered websocket message: %v", err)
		}
		if matches(message) {
			t.Fatalf("unexpected matching websocket message: %#v", message)
		}
	}
}

func ExpectNoRecordChanged(t testing.TB, client *Client, recordID uuid.UUID) {
	t.Helper()
	ExpectNoMatchingMessage(t, client, 300*time.Millisecond, func(message platformws.Message) bool {
		if message.Type != "record_changed" {
			return false
		}
		change, err := collabtestprotocol.RecordChangeFromSequencedMessage(message)
		return err == nil && change.RecordID == recordID
	})
}

func ExpectNoRecordChangedMessage(t testing.TB, client *Client) {
	t.Helper()
	ExpectNoMatchingMessage(t, client, 300*time.Millisecond, func(message platformws.Message) bool {
		return message.Type == "record_changed"
	})
}

func ExpectNoExtensionResourceChanged(t testing.TB, client *Client, resourceID string) {
	t.Helper()
	ExpectNoMatchingMessage(t, client, 300*time.Millisecond, func(message platformws.Message) bool {
		if message.Type != "extension_resource_changed" {
			return false
		}
		var payload platformws.ExtensionResourceChangePayload
		return json.Unmarshal(message.Payload, &payload) == nil && payload.ResourceID == resourceID
	})
}

func ExpectNoSocketMessage(t testing.TB, client *Client) {
	t.Helper()

	_, err := client.AwaitNextMessage(300 * time.Millisecond)
	if err == nil {
		t.Fatal("expected no websocket message")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected websocket receive timeout, got %v", err)
	}
}
