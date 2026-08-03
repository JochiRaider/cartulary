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
	DefangedValue   string
	STIXPattern     string
	HashAlgorithm   string
	HashValue       string
}

var (
	BaseTime = time.Date(2026, time.April, 18, 14, 30, 0, 0, time.UTC)
	PastTime = time.Date(2026, time.April, 17, 9, 15, 0, 0, time.UTC)
	Examples = []Example{
		{
			IndicatorType:   "ipv4_addr",
			ValueKind:       "atomic",
			DisplayValue:    "203.0.113.24",
			NormalizedValue: "203.0.113.24",
			DefangedValue:   "203[.]0[.]113[.]24",
		},
		{
			IndicatorType:   "domain_name",
			ValueKind:       "atomic",
			DisplayValue:    "vpn-gateway.example.test",
			NormalizedValue: "vpn-gateway.example.test",
			DefangedValue:   "vpn-gateway[.]example[.]test",
		},
		{
			IndicatorType:   "url",
			ValueKind:       "atomic",
			DisplayValue:    "https://portal.example.test/login",
			NormalizedValue: "https://portal.example.test/login",
			DefangedValue:   "hxxps://portal[.]example[.]test/login",
		},
		{
			IndicatorType:   "sha256",
			ValueKind:       "pattern",
			DisplayValue:    "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			NormalizedValue: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			HashAlgorithm:   "sha256",
			HashValue:       "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			STIXPattern:     "[file:hashes.'SHA-256' = '2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824']",
		},
	}
)

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
