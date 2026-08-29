package collaboration

import (
	"testing"
	"time"

	"github.com/google/uuid"

	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

func TestWebSocketTelemetrySafeVocabulary(t *testing.T) {
	for _, eventType := range []string{
		"hello_ack",
		"resume_ack",
		"presence_snapshot",
		"presence_delta",
		"record_changed",
		"extension_resource_changed",
		"job_progress",
		"ping",
		"error",
		"session_revoked",
	} {
		if got := safeWebSocketEventType(eventType); got != eventType {
			t.Fatalf("event type %q mapped to %q", eventType, got)
		}
	}
	for _, unsafe := range []string{"resume_result", "incident/10000000"} {
		if got := safeWebSocketEventType(unsafe); got != "other" {
			t.Fatalf("unsafe event type %q mapped to %q", unsafe, got)
		}
	}
	if got := safeWebSocketResult("success"); got != "success" {
		t.Fatalf("unexpected result: %q", got)
	}
	if got := safeWebSocketResult("raw failure"); got != "failed" {
		t.Fatalf("unexpected unsafe result mapping: %q", got)
	}
	if got := safeDropReason("queue_full"); got != "queue_full" {
		t.Fatalf("unexpected drop reason: %q", got)
	}
	if got := safeDropReason("payload overflow"); got != "redaction_rejected" {
		t.Fatalf("unexpected unsafe drop reason mapping: %q", got)
	}
}

func TestWebSocketEventSendTelemetryNoSDK(t *testing.T) {
	hub := newHub("0.0.0+unknown")
	incidentID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	hub.BroadcastPresenceDelta(incidentID, "updated", collabprotocol.PresenceRecord{
		ConnectionID: uuid.NewString(),
		UserID:       uuid.NewString(),
		DisplayName:  "Operator",
		SheetRef:     map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"},
		Mode:         "viewing",
		ObservedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		ExpiresAt:    time.Now().UTC().Add(collabprotocol.PresenceTTL).Format(time.RFC3339Nano),
	}, time.Now().UTC())
}

func TestActiveConnectionTelemetryGaugeNoSDK(t *testing.T) {
	hub := newHub("0.0.0+unknown")
	if got := hub.ActiveConnections(); got != 0 {
		t.Fatalf("initial active connections = %d", got)
	}
	firstDone := hub.TrackActiveConnection()
	secondDone := hub.TrackActiveConnection()
	if got := hub.ActiveConnections(); got != 2 {
		t.Fatalf("active connections after tracking = %d", got)
	}
	firstDone()
	if got := hub.ActiveConnections(); got != 1 {
		t.Fatalf("active connections after first close = %d", got)
	}
	secondDone()
	if got := hub.ActiveConnections(); got != 0 {
		t.Fatalf("active connections after second close = %d", got)
	}
}
