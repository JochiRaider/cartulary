package phase2test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func ConnectTimelineSocket(t testing.TB, serverURL string, incidentID string, sessionToken string) *wstest.Client {
	t.Helper()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionToken)
	client := wstest.ConnectWithHeaders(t, serverURL, "/ws/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/changes", headers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message, err := client.Receive(ctx)
	if err != nil {
		t.Fatalf("receive phase2 timeline websocket connected message: %v", err)
	}
	wstest.RequireMessageType(t, message, "connected")

	var payload map[string]any
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("decode phase2 timeline websocket payload: %v", err)
	}
	if payload["incident_id"] != incidentID {
		t.Fatalf("unexpected phase2 timeline websocket incident payload: %#v", payload)
	}
	return client
}
