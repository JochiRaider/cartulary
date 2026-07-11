package phase3storetest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"

	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

type RecordChangeSocketPayload struct {
	RecordID         string   `json:"record_id"`
	RowVersion       float64  `json:"row_version"`
	ChangeSetID      string   `json:"change_set_id"`
	ClientTxnID      string   `json:"client_txn_id"`
	ChangedFieldKeys []string `json:"changed_field_keys"`
}

type TimelineSocketClient struct {
	raw      *wstest.Client
	messages chan platformws.Message
	errors   chan error
}

func (c *TimelineSocketClient) Close(code int, reason string) {
	if c == nil || c.raw == nil {
		return
	}
	c.raw.Close(websocket.StatusCode(code), reason)
}

func ConnectTimelineSocket(t testing.TB, server *httptestx.Server, incidentID string, sessionToken string) *TimelineSocketClient {
	t.Helper()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionToken)
	rawClient := wstest.ConnectWithHeaders(t, server.HTTP.URL, "/ws/v1/incidents/"+incidentID, headers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rawClient.Send(ctx, platformws.Message{
		Type: "hello",
		Payload: platformws.RawPayload(map[string]any{
			"client_instance_id": "phase3-test-" + incidentID,
			"presence": map[string]any{
				"sheet_ref": map[string]any{
					"kind": "view_schema",
					"id":   timelineViewSchemaID,
				},
				"mode": "viewing",
			},
		}),
	}); err != nil {
		t.Fatalf("send websocket hello: %v", err)
	}
	message, err := rawClient.Receive(ctx)
	if err != nil {
		t.Fatalf("receive websocket hello_ack message: %v", err)
	}
	wstest.RequireMessageType(t, message, "hello_ack")
	message, err = rawClient.Receive(ctx)
	if err != nil {
		t.Fatalf("receive websocket presence_snapshot message: %v", err)
	}
	wstest.RequireMessageType(t, message, "presence_snapshot")

	client := &TimelineSocketClient{
		raw:      rawClient,
		messages: make(chan platformws.Message, 32),
		errors:   make(chan error, 1),
	}
	go func() {
		for {
			message, err := rawClient.Receive(context.Background())
			if err != nil {
				select {
				case client.errors <- err:
				default:
				}
				return
			}
			select {
			case client.messages <- message:
			default:
				select {
				case client.errors <- fmt.Errorf("timeline websocket buffer overflow"):
				default:
				}
				return
			}
		}
	}()
	return client
}

func RequireTimelineSocketChange(t testing.TB, client *TimelineSocketClient, wantRecordID string, wantRowVersion int64) RecordChangeSocketPayload {
	t.Helper()

	var message platformws.Message
	select {
	case message = <-client.messages:
	case err := <-client.errors:
		t.Fatalf("receive websocket record_changed: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for websocket record_changed")
	}
	wstest.RequireMessageType(t, message, "record_changed")

	var payload RecordChangeSocketPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("decode record_changed payload: %v", err)
	}
	if payload.RecordID != wantRecordID || payload.RowVersion != float64(wantRowVersion) {
		t.Fatalf("unexpected record_changed payload: %#v", payload)
	}
	if payload.ClientTxnID == "" {
		t.Fatalf("expected websocket payload client_txn_id, got %#v", payload)
	}
	return payload
}

func ExpectNoTimelineSocketMessage(t testing.TB, client *TimelineSocketClient) {
	t.Helper()

	select {
	case message := <-client.messages:
		t.Fatalf("expected no websocket message, got %#v", message)
	case err := <-client.errors:
		t.Fatalf("expected no websocket message, got read error %v", err)
	case <-time.After(300 * time.Millisecond):
	}
}
