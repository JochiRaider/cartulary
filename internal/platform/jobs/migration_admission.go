package jobs

import (
	"context"
	"database/sql"
)

// RetainedExtensionStatePresent probes Jobs-owned shells, expired tombstones
// and independent commit proofs before an owner's fresh-state schema boundary.
// Older installations may predate individual columns; no payload is decoded.
func RetainedExtensionStatePresent(ctx context.Context, conn *sql.Conn, profile, kind, handler string) (bool, error) {
	for _, probe := range []struct{ table, column, value, query string }{
		{"jobs", "extension_owner_profile_id", profile, "SELECT EXISTS (SELECT 1 FROM public.jobs WHERE extension_owner_profile_id=$1)"},
		{"jobs", "job_kind", kind, "SELECT EXISTS (SELECT 1 FROM public.jobs WHERE job_kind=$1)"},
		{"jobs", "handler_name", handler, "SELECT EXISTS (SELECT 1 FROM public.jobs WHERE handler_name=$1)"},
		{"extension_job_commit_proofs", "owner_profile_id", profile, "SELECT EXISTS (SELECT 1 FROM public.extension_job_commit_proofs WHERE owner_profile_id=$1)"},
	} {
		var exists bool
		if err := conn.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=$1 AND column_name=$2)`, probe.table, probe.column).Scan(&exists); err != nil {
			return false, err
		}
		if !exists {
			continue
		}
		var retained bool
		if err := conn.QueryRowContext(ctx, probe.query, probe.value).Scan(&retained); err != nil {
			return false, err
		}
		if retained {
			return true, nil
		}
	}
	return false, nil
}
