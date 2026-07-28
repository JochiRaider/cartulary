package incidentwstest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration"
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
		ViewSchemaID string `json:"view_schema_id"`
		ChangeKind   string `json:"change_kind"`
	} `json:"affected_views"`
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

func RequireRecordChanged(
	t testing.TB,
	client *Client,
	wantRecordID string,
	wantRowVersion int64,
) RecordChangeSocketPayload {
	t.Helper()

	var lastPayload *RecordChangeSocketPayload
	for {
		message, err := client.AwaitNextMessage(5 * time.Second)
		if err != nil {
			if lastPayload != nil {
				t.Fatalf("timed out waiting for websocket record_changed record=%s version=%d after seeing %#v: %v", wantRecordID, wantRowVersion, *lastPayload, err)
			}
			t.Fatalf("receive websocket record_changed: %v", err)
		}
		if message.Type != "record_changed" {
			continue
		}

		var payload RecordChangeSocketPayload
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			t.Fatalf("decode record_changed payload: %v", err)
		}
		lastPayload = &payload
		if payload.RecordID == wantRecordID && payload.RowVersion == float64(wantRowVersion) {
			return payload
		}
	}
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
