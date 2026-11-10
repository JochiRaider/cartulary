// Package referenceadmission inspects retained Reporting bindings for current
// Reference Pack contract admission, independently of database migration.
package referenceadmission

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

type ReferenceReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

// Reporting owns these paths, including the Composition preview delegated to
// Reporting. Reference Data owns the callback and the referenced set catalog.
var referenceQueries = []sqlshape.Fragment{
	sqlshape.Literal(`SELECT CASE WHEN job_kind='snapshot_create' THEN snapshot_admission_json->'reference_packs' ELSE request_json->'reference_packs' END FROM public.reporting_job_payloads`),
	sqlshape.Literal(`SELECT request_json->'source_capture'->'reference_packs' FROM public.reporting_job_payloads WHERE job_kind='snapshot_create' AND request_json IS NOT NULL`),
	sqlshape.Literal(`SELECT bundle_manifest_json->'reference_packs' FROM public.reporting_render_bundles`),
	sqlshape.Literal(`SELECT bundle_manifest_json->'reference_packs' FROM public.reporting_composition_preview_outputs`),
	sqlshape.Literal(`SELECT convert_from(string_agg(c.content,''::bytea ORDER BY c.chunk_ordinal),'UTF8')::jsonb->'reference_packs' FROM public.reporting_imported_artifact_files f LEFT JOIN public.reporting_imported_artifact_chunks c USING(incident_id,bundle_path) WHERE f.bundle_path LIKE 'ext/snapshots/releases/%/manifest.json' OR f.bundle_path LIKE 'ext/snapshots/%/source-capture.json' GROUP BY f.incident_id,f.bundle_path`),
}

func ValidateReferenceBindings(ctx context.Context, db ReferenceReader, admit func([]byte) error) error {
	bindings := map[string]bool{}
	for _, q := range referenceQueries {
		rows, err := db.Query(ctx, sqlshape.Text(q))
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
