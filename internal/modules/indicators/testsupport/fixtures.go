package testsupport

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
)

type Example struct {
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue string
}

func PrimaryExample() Example {
	return Example{
		IndicatorType:   "ipv4_addr",
		ValueKind:       "atomic",
		DisplayValue:    "203.0.113.24",
		NormalizedValue: "203.0.113.24",
	}
}

func BaseTime() time.Time {
	return time.Date(2026, time.April, 18, 14, 30, 0, 0, time.UTC)
}

func PastTime() time.Time {
	return time.Date(2026, time.April, 17, 9, 15, 0, 0, time.UTC)
}

func CreatePayload(clientTxnID string) map[string]any {
	return map[string]any{
		"client_txn_id":              clientTxnID,
		"indicator.indicator_type":   "ipv4_addr",
		"indicator.value_kind":       "atomic",
		"indicator.display_value":    "203.0.113.24",
		"indicator.normalized_value": "203.0.113.24",
		"indicator.defanged_value":   "203[.]0[.]113[.]24",
	}
}

func CanonicalDedupeKey(t testing.TB, indicatorType string, valueKind string, displayValue string) string {
	t.Helper()
	canonical, err := identity.Canonicalize(identity.Input{
		IndicatorType: indicatorType,
		ValueKind:     valueKind,
		DisplayValue:  displayValue,
	})
	if err != nil {
		t.Fatalf("canonicalize Indicator fixture: %v", err)
	}
	return canonical.DedupeKey
}

// SeedRecord creates a Records envelope and its Indicator-owned subtype row.
// Cross-owner tests use this owner fixture instead of depending on the physical
// Indicator schema.
func SeedRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, indicatorType string, valueKind string, displayValue string) {
	t.Helper()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, recordID, "indicator")
	SeedSubtype(t, db, incidentID, recordID, indicatorType, valueKind, displayValue)
}

// SeedSubtype creates only the Indicator-owned row. It is reserved for tests
// that deliberately control the Records envelope, such as portability fixtures.
func SeedSubtype(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, indicatorType string, valueKind string, displayValue string) {
	t.Helper()
	canonical, err := identity.Canonicalize(identity.Input{
		IndicatorType: indicatorType,
		ValueKind:     valueKind,
		DisplayValue:  displayValue,
	})
	if err != nil {
		t.Fatalf("canonicalize Indicator fixture: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO indicators (
    record_id,
    incident_id,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    dedupe_key
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, recordID, incidentID, canonical.IndicatorType, canonical.ValueKind, canonical.DisplayValue, canonical.NormalizedValue, canonical.DedupeKey); err != nil {
		t.Fatalf("seed Indicator subtype: %v", err)
	}
}
