package ws

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestWebSocketTelemetrySafeVocabulary(t *testing.T) {
	if got := safeWebSocketEventType("record_changed"); got != "record_changed" {
		t.Fatalf("unexpected event type: %q", got)
	}
	if got := safeWebSocketEventType("incident/10000000"); got != "other" {
		t.Fatalf("unexpected unsafe event type mapping: %q", got)
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
	hub := NewHub()
	hub.ConfigureTelemetry("0.0.0+unknown")
	incidentID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	hub.BroadcastPresenceDelta(incidentID, "updated", PresenceRecord{
		ConnectionID: uuid.NewString(),
		UserID:       uuid.NewString(),
		DisplayName:  "Operator",
		SheetRef:     map[string]string{"kind": "view_schema", "id": "cartulary.view.timeline.v2"},
		Mode:         "viewing",
		ObservedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		ExpiresAt:    time.Now().UTC().Add(PresenceTTL).Format(time.RFC3339Nano),
	}, time.Now().UTC())
}

func TestActiveConnectionTelemetryGaugeNoSDK(t *testing.T) {
	hub := NewHub()
	hub.ConfigureTelemetry("0.0.0+unknown")
	hub.ConfigureTelemetry("0.0.0+unknown")
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
