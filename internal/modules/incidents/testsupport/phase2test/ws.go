package phase2test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func ConnectTimelineSocket(t testing.TB, serverURL string, incidentID string, sessionToken string) *wstest.Client {
	t.Helper()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionToken)
	client := wstest.ConnectWithHeaders(t, serverURL, "/ws/v1/incidents/"+incidentID, headers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Send(ctx, platformws.Message{
		Type: "hello",
		Payload: platformws.RawPayload(map[string]any{
			"client_instance_id": "phase2-test-" + incidentID,
			"presence": map[string]any{
				"sheet_ref": map[string]any{
					"kind": "view_schema",
					"id":   timeline.TimelineViewSchemaID,
				},
				"mode": "viewing",
			},
		}),
	}); err != nil {
		t.Fatalf("send phase2 timeline websocket hello: %v", err)
	}
	message, err := client.Receive(ctx)
	if err != nil {
		t.Fatalf("receive phase2 timeline websocket hello_ack message: %v", err)
	}
	wstest.RequireMessageType(t, message, "hello_ack")

	var payload map[string]any
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("decode phase2 timeline websocket payload: %v", err)
	}
	if payload["connection_id"] == "" {
		t.Fatalf("unexpected phase2 timeline websocket hello payload: %#v", payload)
	}
	message, err = client.Receive(ctx)
	if err != nil {
		t.Fatalf("receive phase2 timeline websocket presence_snapshot message: %v", err)
	}
	wstest.RequireMessageType(t, message, "presence_snapshot")
	return client
}
