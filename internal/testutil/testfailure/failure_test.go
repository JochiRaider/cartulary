package testfailure

import (
	"strings"
	"testing"
)

func TestEncodeProducesBoundedMarkerWithoutDiagnosticText(t *testing.T) {
	envelope := NewEnvelope(
		"infra",
		"service_readiness_timeout",
		"s3test",
		"object_store",
		"put",
		25,
		"completed",
	)
	encoded, err := Encode(envelope)
	if err != nil {
		t.Fatalf("encode envelope: %v", err)
	}
	if !strings.HasPrefix(encoded, Marker) {
		t.Fatalf("marker prefix missing: %q", encoded)
	}
	if strings.Contains(encoded, "transport") || strings.Contains(encoded, "credential") {
		t.Fatalf("marker leaked diagnostic text: %q", encoded)
	}
}

func TestEnvelopeRejectsUnregisteredValues(t *testing.T) {
	envelope := NewEnvelope("infra", "other", "s3test", "object_store", "put", 1, "completed")
	if err := envelope.Validate(); err == nil {
		t.Fatal("expected unregistered reason to fail")
	}
}
