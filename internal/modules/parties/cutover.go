package parties

import (
	"context"
	"database/sql"
	"errors"
)

type ClockCutoverRequiredError struct{}

func (*ClockCutoverRequiredError) Error() string      { return "party_clock_cutover_required" }
func (*ClockCutoverRequiredError) ReasonCode() string { return "party_clock_cutover_required" }
func (*ClockCutoverRequiredError) RemediationReportJSON() string {
	return `{"schema_id":"cartulary.party_clock_cutover_report.v1","reason_code":"party_clock_cutover_required","required_action":"Preserve a matching application/data backup. Explicitly initialize a fresh disposable environment; retained Party history is not converted."}`
}

// PreflightClockCutover is read-only and runs before the migration ledger or
// source schema changes. Migration 73 repeats this guard under table locks.
func PreflightClockCutover(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("parties: cutover preflight database unavailable")
	}
	var retiredClock bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_attribute a JOIN pg_class t ON t.oid=a.attrelid JOIN pg_namespace n ON n.oid=t.relnamespace WHERE n.nspname='public' AND t.relname='parties' AND a.attname='created_at' AND NOT a.attisdropped)`).Scan(&retiredClock); err != nil {
		return err
	}
	if !retiredClock {
		return nil
	}
	for _, item := range []struct{ table, query string }{
		{"parties", `SELECT EXISTS (SELECT 1 FROM public.parties)`},
		{"records", `SELECT EXISTS (SELECT 1 FROM public.records WHERE record_type='party')`},
		{"change_set_mutations", `SELECT EXISTS (SELECT 1 FROM public.change_set_mutations WHERE before_value->>'snapshot_schema_id'='cartulary.revisions.snapshot.party.v1' OR after_value->>'snapshot_schema_id'='cartulary.revisions.snapshot.party.v1')`},
		{"record_revisions", `SELECT EXISTS (SELECT 1 FROM public.record_revisions WHERE before_json->>'snapshot_schema_id'='cartulary.revisions.snapshot.party.v1' OR after_json->>'snapshot_schema_id'='cartulary.revisions.snapshot.party.v1')`},
	} {
		var present bool
		if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, "public."+item.table).Scan(&present); err != nil {
			return err
		}
		if !present {
			continue
		}
		var retained bool
		if err := db.QueryRowContext(ctx, item.query).Scan(&retained); err != nil {
			return err
		}
		if retained {
			return &ClockCutoverRequiredError{}
		}
	}
	return nil
}
