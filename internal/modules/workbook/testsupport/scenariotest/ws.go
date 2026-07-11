package scenariotest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
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

func ConnectViewSocket(t testing.TB, server *httptestx.Server, incidentID string, viewSchemaID string, sessionToken string) *wstest.Client {
	t.Helper()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionToken)
	client := wstest.ConnectWithHeaders(t, server.HTTP.URL, "/ws/v1/incidents/"+incidentID, headers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Send(ctx, platformws.Message{
		Type: "hello",
		Payload: platformws.RawPayload(map[string]any{
			"client_instance_id": "workbook-test-" + incidentID,
			"presence": map[string]any{
				"sheet_ref": map[string]any{
					"kind": "view_schema",
					"id":   viewSchemaID,
				},
				"mode": "viewing",
			},
		}),
	}); err != nil {
		t.Fatalf("send websocket hello: %v", err)
	}
	message, err := client.Receive(ctx)
	if err != nil {
		t.Fatalf("receive websocket hello_ack message: %v", err)
	}
	wstest.RequireMessageType(t, message, "hello_ack")
	message, err = client.Receive(ctx)
	if err != nil {
		t.Fatalf("receive websocket presence_snapshot message: %v", err)
	}
	wstest.RequireMessageType(t, message, "presence_snapshot")
	return client
}

func RequireRecordChanged(t testing.TB, client *wstest.Client, wantRecordID string, wantRowVersion int64) RecordChangeSocketPayload {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lastPayload *RecordChangeSocketPayload
	for {
		message, err := client.Receive(ctx)
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

func ExpectNoSocketMessage(t testing.TB, client *wstest.Client) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	_, err := client.Receive(ctx)
	if err == nil {
		t.Fatal("expected no websocket message")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected websocket receive timeout, got %v", err)
	}
}
