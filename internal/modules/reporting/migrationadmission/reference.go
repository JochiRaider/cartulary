package migrationadmission

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

type ReferenceReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
type ReferenceSQLReader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// Reporting owns these paths, including the Composition preview delegated to
// Reporting. Reference Data owns the callback and the referenced set catalog.
var referenceQueries = []struct{ table, query string }{
	{"reporting_job_payloads", `SELECT CASE WHEN job_kind='snapshot_create' THEN snapshot_admission_json->'reference_packs' ELSE request_json->'reference_packs' END FROM public.reporting_job_payloads`},
	{"reporting_job_payloads", `SELECT request_json->'source_capture'->'reference_packs' FROM public.reporting_job_payloads WHERE job_kind='snapshot_create' AND request_json IS NOT NULL`},
	{"reporting_render_bundles", `SELECT bundle_manifest_json->'reference_packs' FROM public.reporting_render_bundles`},
	{"reporting_composition_preview_outputs", `SELECT bundle_manifest_json->'reference_packs' FROM public.reporting_composition_preview_outputs`},
	{"reporting_imported_artifact_chunks", `SELECT convert_from(string_agg(c.content,''::bytea ORDER BY c.chunk_ordinal),'UTF8')::jsonb->'reference_packs' FROM public.reporting_imported_artifact_files f LEFT JOIN public.reporting_imported_artifact_chunks c USING(incident_id,bundle_path) WHERE f.bundle_path LIKE 'ext/snapshots/releases/%/manifest.json' OR f.bundle_path LIKE 'ext/snapshots/%/source-capture.json' GROUP BY f.incident_id,f.bundle_path`},
}

func ValidateReferenceBindingsMigration(ctx context.Context, conn ReferenceSQLReader, admit func([]byte) error) error {
	bindings := map[string]bool{}
	for _, q := range referenceQueries {
		var present bool
		if err := conn.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+q.table).Scan(&present); err != nil {
			return err
		}
		if !present {
			continue
		}
		rows, err := conn.QueryContext(ctx, q.query)
		if err != nil {
			return err
		}
		for rows.Next() {
			var b []byte
			if err = rows.Scan(&b); err != nil {
				rows.Close()
				return err
			}
			bindings[string(b)] = true
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	return admitBindings(bindings, admit)
}
func ValidateReferenceBindings(ctx context.Context, db ReferenceReader, admit func([]byte) error) error {
	bindings := map[string]bool{}
	for _, q := range referenceQueries {
		var present bool
		if err := db.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+q.table).Scan(&present); err != nil {
			return err
		}
		if !present {
			continue
		}
		rows, err := db.Query(ctx, q.query)
		if err != nil {
			return err
		}
		for rows.Next() {
			var b []byte
			if err = rows.Scan(&b); err != nil {
				rows.Close()
				return err
			}
			bindings[string(b)] = true
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	return admitBindings(bindings, admit)
}
func admitBindings(bindings map[string]bool, admit func([]byte) error) error {
	if admit == nil {
		return errors.New("reporting Reference Data admission callback required")
	}
	for b := range bindings {
		if err := admit([]byte(b)); err != nil {
			return err
		}
	}
	return nil
}
