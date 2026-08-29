package evidence

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEvidenceCreateValidation_AdoptedSignalAndLifecycleMatrix(t *testing.T) {
	text := func(value string) FieldValue {
		return FieldValue{Text: &value}
	}
	timestamp := func(value time.Time) FieldValue {
		return FieldValue{Timestamp: &value}
	}
	identifier := func(value uuid.UUID) FieldValue {
		return FieldValue{UUID: &value}
	}

	now := time.Date(2026, time.July, 30, 18, 0, 0, 0, time.UTC)
	partyID := uuid.MustParse("ac905af9-b062-4ba4-970c-19df150ca91e")
	cases := []struct {
		name          string
		values        map[string]FieldValue
		finalizedBlob bool
		wantReason    string
		wantLifecycle bool
	}{
		{name: "blank", values: map[string]FieldValue{}, wantReason: "minimum_create_signal_missing"},
		{name: "whitespace title", values: map[string]FieldValue{"evidence.title": text(" \t\r\n ")}, wantReason: "minimum_create_signal_missing"},
		{name: "evidence.title", values: map[string]FieldValue{"evidence.title": text("evidence.title")}},
		{name: "storage ref", values: map[string]FieldValue{"evidence.storage_ref": text("s3://case/object")}},
		{name: "reserved object ref is owner rejected", values: map[string]FieldValue{"evidence.storage_ref": text("object://00000000-0000-0000-0000-000000210004")}, wantReason: "reserved_server_managed_ref"},
		{name: "malformed reserved object ref is owner rejected", values: map[string]FieldValue{"evidence.storage_ref": text("object://not-a-uuid")}, wantReason: "reserved_server_managed_ref"},
		{name: "collector text", values: map[string]FieldValue{"evidence.collector_party_text": text("collector")}},
		{name: "source text", values: map[string]FieldValue{"evidence.source_party_text": text("source")}},
		{name: "requested lifecycle alone", values: map[string]FieldValue{"evidence.lifecycle_state": text("requested")}},
		{name: "pending lifecycle alone", values: map[string]FieldValue{"evidence.lifecycle_state": text("pending_receipt")}},
		{name: "received lifecycle alone", values: map[string]FieldValue{"evidence.lifecycle_state": text("received")}},
		{name: "available lifecycle requires blob", values: map[string]FieldValue{"evidence.lifecycle_state": text("available")}, wantLifecycle: true},
		{name: "quarantined lifecycle without blob", values: map[string]FieldValue{"evidence.lifecycle_state": text("quarantined")}},
		{name: "released lifecycle is not initial", values: map[string]FieldValue{"evidence.lifecycle_state": text("released")}, wantLifecycle: true},
		{name: "requested timestamp alone", values: map[string]FieldValue{"evidence.requested_at": timestamp(now)}},
		{name: "received timestamp alone", values: map[string]FieldValue{"evidence.received_at": timestamp(now)}},
		{name: "collector party id alone", values: map[string]FieldValue{"evidence.collector_party_id": identifier(partyID)}, wantReason: "minimum_create_signal_missing"},
		{name: "source party id alone", values: map[string]FieldValue{"evidence.source_party_id": identifier(partyID)}, wantReason: "minimum_create_signal_missing"},
		{name: "invalid lifecycle after qualifying signal", values: map[string]FieldValue{"evidence.title": text("evidence.title"), "evidence.lifecycle_state": text("unknown")}, wantReason: "invalid_value"},
		{name: "finalized blob alone", values: map[string]FieldValue{}, finalizedBlob: true},
		{name: "finalized blob requested", values: map[string]FieldValue{"evidence.lifecycle_state": text("requested")}, finalizedBlob: true},
		{name: "finalized blob pending", values: map[string]FieldValue{"evidence.lifecycle_state": text("pending_receipt")}, finalizedBlob: true},
		{name: "finalized blob received", values: map[string]FieldValue{"evidence.lifecycle_state": text("received")}, finalizedBlob: true},
		{name: "finalized blob available", values: map[string]FieldValue{"evidence.lifecycle_state": text("available")}, finalizedBlob: true},
		{name: "finalized blob quarantined", values: map[string]FieldValue{"evidence.lifecycle_state": text("quarantined")}, finalizedBlob: true, wantLifecycle: true},
		{name: "finalized blob released", values: map[string]FieldValue{"evidence.lifecycle_state": text("released")}, finalizedBlob: true, wantLifecycle: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCreateParams(createParams{
				Values:                 tc.values,
				InitialBlobFinalized:   tc.finalizedBlob,
				InitialBlobWasSupplied: tc.finalizedBlob,
			})
			if tc.wantReason == "" {
				if tc.wantLifecycle {
					var lifecycleErr *LifecycleValidationError
					if !errors.As(err, &lifecycleErr) {
						t.Fatalf("validateCreateParams() error = %v, want LifecycleValidationError", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("validateCreateParams() error = %v", err)
				}
				return
			}
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("validateCreateParams() error = %v, want ValidationError", err)
			}
			if validationErr.ReasonCode != tc.wantReason {
				t.Fatalf("reason = %q, want %q", validationErr.ReasonCode, tc.wantReason)
			}
		})
	}
}
