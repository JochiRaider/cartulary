package scenariotest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/coder/websocket"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
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
	incident *incidentwstest.Client
}

func (c *TimelineSocketClient) Close(code int, reason string) {
	if c == nil || c.incident == nil {
		return
	}
	c.incident.Close(websocket.StatusCode(code), reason)
}

func ConnectTimelineSocket(t testing.TB, server *httptestx.Server, incidentID string, sessionToken string) *TimelineSocketClient {
	t.Helper()

	client := incidentwstest.ConnectAndHello(t, server.HTTP.URL, incidentID, incidentwstest.ConnectOptions{
		SessionToken:     sessionToken,
		ClientInstanceID: "timeline-mutation-test-" + incidentID,
		Presence: platformws.PresenceInput{
			SheetRef: map[string]string{
				"kind": "view_schema",
				"id":   timeline.TimelineViewSchemaID,
			},
			Mode: "viewing",
		},
	})
	return &TimelineSocketClient{incident: client}
}

func RequireTimelineSocketChange(t testing.TB, client *TimelineSocketClient, wantRecordID string, wantRowVersion int64) RecordChangeSocketPayload {
	t.Helper()

	message, err := client.incident.AwaitNextMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("receive websocket record_changed: %v", err)
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

	message, err := client.incident.AwaitNextMessage(300 * time.Millisecond)
	if err == nil {
		t.Fatalf("expected no websocket message, got %#v", message)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected no websocket message, got read error %v", err)
	}
}
