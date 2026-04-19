package phase4test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

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
	client := wstest.ConnectWithHeaders(t, server.HTTP.URL, "/ws/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/changes", headers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message, err := client.Receive(ctx)
	if err != nil {
		t.Fatalf("receive websocket connected message: %v", err)
	}
	wstest.RequireMessageType(t, message, "connected")
	return client
}

func RequireRecordChanged(t testing.TB, client *wstest.Client, wantRecordID string, wantRowVersion int64) RecordChangeSocketPayload {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message, err := client.Receive(ctx)
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
	return payload
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
