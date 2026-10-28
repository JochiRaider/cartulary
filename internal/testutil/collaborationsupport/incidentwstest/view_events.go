package incidentwstest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

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

// awaitReplayable consumes only declared ephemeral traffic between mutation events.
// Every replayable observation must belong to this incident and advance the stream.
func (c *Client) awaitReplayable(timeout time.Duration) (platformws.Message, error) {
	deadline := time.Now().Add(timeout)
	for {
		message, err := c.AwaitNextMessage(time.Until(deadline))
		if err != nil {
			return platformws.Message{}, err
		}
		if message.Type == "presence_delta" || message.Type == "ping" {
			if message.StreamSeq != nil {
				return message, fmt.Errorf("ephemeral %s has stream_seq", message.Type)
			}
			continue
		}
		if err := platformws.ValidateSequencedReplayableMessage(message); err != nil {
			return message, fmt.Errorf("invalid mutation event: %w", err)
		}
		if message.IncidentID != c.incidentID || *message.StreamSeq <= c.lastStreamSeq {
			return message, fmt.Errorf("unexpected incident/sequence: got %s/%d after %s/%d", message.IncidentID, *message.StreamSeq, c.incidentID, c.lastStreamSeq)
		}
		if message.Type == "record_changed" {
			if _, err := collabtestprotocol.RecordChangeFromSequencedMessage(message); err != nil {
				return message, err
			}
		}
		c.lastStreamSeq = *message.StreamSeq
		return message, nil
	}
}

func RequireRecordChangedEvent(t testing.TB, client *Client, wantRecordID uuid.UUID, wantRowVersion int64) collabtestprotocol.RecordChangedEvent {
	t.Helper()
	changes := RequireRecordChangedEvents(t, client, map[uuid.UUID]int64{wantRecordID: wantRowVersion})
	return changes[wantRecordID]
}

func (c *Client) awaitRecordChangedEvents(want map[uuid.UUID]int64) (map[uuid.UUID]collabtestprotocol.RecordChangedEvent, error) {
	result := make(map[uuid.UUID]collabtestprotocol.RecordChangedEvent, len(want))
	deadline := time.Now().Add(waitTimeout)
	for len(result) < len(want) {
		message, err := c.awaitReplayable(time.Until(deadline))
		if err != nil {
			return nil, err
		}
		change, err := collabtestprotocol.RecordChangeFromSequencedMessage(message)
		if err != nil {
			return nil, err
		}
		version, expected := want[change.RecordID]
		_, duplicate := result[change.RecordID]
		if !expected || duplicate || version != change.RowVersion {
			return nil, fmt.Errorf("unexpected record_changed record=%s version=%d duplicate=%t; want %v", change.RecordID, change.RowVersion, duplicate, want)
		}
		result[change.RecordID] = change
	}
	return result, nil
}

func RequireRecordChangedEvents(t testing.TB, client *Client, want map[uuid.UUID]int64) map[uuid.UUID]collabtestprotocol.RecordChangedEvent {
	t.Helper()
	result, err := client.awaitRecordChangedEvents(want)
	if err != nil {
		t.Fatalf("receive expected record_changed events: %v", err)
	}
	return result
}

func RequireRecordChanged(
	t testing.TB,
	client *Client,
	wantRecordID string,
	wantRowVersion int64,
) collabtestprotocol.RecordChangedEvent {
	t.Helper()

	recordID, err := uuid.Parse(wantRecordID)
	if err != nil {
		t.Fatalf("parse expected record_changed record_id: %v", err)
	}
	return RequireRecordChangedEvent(t, client, recordID, wantRowVersion)
}

func RequireExtensionResourceChanged(t testing.TB, client *Client, want ExtensionResourceChangeExpectation) platformws.ExtensionResourceChangePayload {
	t.Helper()
	message, err := client.awaitReplayable(waitTimeout)
	if err != nil {
		t.Fatalf("receive extension_resource_changed: %v", err)
	}
	if message.Type != "extension_resource_changed" {
		t.Fatalf("unexpected event type %q", message.Type)
	}
	var payload platformws.ExtensionResourceChangePayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if message.IncidentID != want.IncidentID || payload.ResourceID != want.ResourceID || payload.ExtensionProfileID != want.ExtensionProfileID || payload.ResourceKind != want.ResourceKind || payload.ChangeKind != want.ChangeKind || payload.ReasonCode != want.ReasonCode {
		t.Fatalf("unexpected extension_resource_changed: message=%#v payload=%#v want=%#v", message, payload, want)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(message.Payload, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range want.ForbiddenKeys {
		if _, exists := raw[key]; exists {
			t.Fatalf("extension event disclosed %q", key)
		}
	}
	return payload
}

// ExpectNoReplayableMessage is a bounded observation, not proof of no durable effect.
func ExpectNoReplayableMessage(t testing.TB, client *Client) {
	t.Helper()
	message, err := client.awaitReplayable(300 * time.Millisecond)
	if errors.Is(err, context.DeadlineExceeded) {
		return
	}
	t.Fatalf("unexpected mutation observation: message=%#v error=%v", message, err)
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
