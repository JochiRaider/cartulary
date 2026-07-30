package testsupport

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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

type ProjectionRow struct {
	RecordID            uuid.UUID
	RowVersion          int64
	IndicatorType       string
	ValueKind           string
	DisplayValue        string
	NormalizedValue     *string
	DefangedValue       *string
	HashAlgorithm       *string
	HashValue           *string
	STIXPattern         *string
	FirstObservedAt     *time.Time
	LastObservedAt      *time.Time
	ObservationCount    int
	LifecycleSummary    *string
	SupportingLinkCount int
}

func LookupProjection(t testing.TB, db postgres.DB, recordID uuid.UUID) ProjectionRow {
	t.Helper()
	var (
		row              ProjectionRow
		recordIDRaw      string
		normalizedValue  sql.NullString
		defangedValue    sql.NullString
		hashAlgorithm    sql.NullString
		hashValue        sql.NullString
		stixPattern      sql.NullString
		firstObservedAt  sql.NullTime
		lastObservedAt   sql.NullTime
		lifecycleSummary sql.NullString
	)
	if err := db.QueryRow(context.Background(), `
SELECT record_id::text, row_version, indicator_type, value_kind, display_value,
       normalized_value, defanged_value, hash_algorithm, hash_value, stix_pattern,
       first_observed_at, last_observed_at, observation_count, lifecycle_summary,
       supporting_link_count
  FROM indicator_grid_projection
 WHERE record_id = $1
`, recordID).Scan(
		&recordIDRaw,
		&row.RowVersion,
		&row.IndicatorType,
		&row.ValueKind,
		&row.DisplayValue,
		&normalizedValue,
		&defangedValue,
		&hashAlgorithm,
		&hashValue,
		&stixPattern,
		&firstObservedAt,
		&lastObservedAt,
		&row.ObservationCount,
		&lifecycleSummary,
		&row.SupportingLinkCount,
	); err != nil {
		t.Fatalf("lookup indicator projection: %v", err)
	}
	row.RecordID = uuid.MustParse(recordIDRaw)
	row.NormalizedValue = nullStringPointer(normalizedValue)
	row.DefangedValue = nullStringPointer(defangedValue)
	row.HashAlgorithm = nullStringPointer(hashAlgorithm)
	row.HashValue = nullStringPointer(hashValue)
	row.STIXPattern = nullStringPointer(stixPattern)
	row.FirstObservedAt = nullTimePointer(firstObservedAt)
	row.LastObservedAt = nullTimePointer(lastObservedAt)
	row.LifecycleSummary = nullStringPointer(lifecycleSummary)
	return row
}

func nullStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func nullTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}
