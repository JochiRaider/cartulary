// Package migrationadmission owns Reporting's read-only retained-input cutover
// check. Application assembly invokes it before the migration provider writes.
package migrationadmission

import (
	"context"
	"database/sql"
)

func Validate(ctx context.Context, conn *sql.Conn) error {
	var present bool
	if err := conn.QueryRowContext(ctx, `SELECT to_regclass('public.reporting_job_payloads') IS NOT NULL`).Scan(&present); err != nil || !present {
		return err
	}
	var incompatible bool
	if err := conn.QueryRowContext(ctx, `SELECT EXISTS (
        SELECT 1 FROM public.reporting_job_payloads
         WHERE job_kind <> 'snapshot_create'
           AND request_json->'graph_sources'->>'schema_id' IS DISTINCT FROM 'cartulary.reporting_graph_input.v2'
    )`).Scan(&incompatible); err != nil {
		return err
	}
	if incompatible {
		return incompatibleInput{}
	}
	return nil
}

type incompatibleInput struct{}

func (incompatibleInput) Error() string {
	return "orientation cutover requires fresh Reporting inputs; retain historical reports in the previous compatible environment"
}

func (incompatibleInput) ReasonCode() string { return "reporting_direction_state_incompatible" }
