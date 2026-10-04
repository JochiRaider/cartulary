package reporting

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type CutoverRequiredError struct{ categories []reportingCutoverCategory }
type reportingCutoverCategory struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

func (*CutoverRequiredError) Error() string      { return "reporting_cutover_required" }
func (*CutoverRequiredError) ReasonCode() string { return "reporting_cutover_required" }
func (e *CutoverRequiredError) RemediationReportJSON() string {
	b, _ := json.Marshal(struct {
		SchemaID       string                     `json:"schema_id"`
		ReasonCode     string                     `json:"reason_code"`
		Categories     []reportingCutoverCategory `json:"categories"`
		RequiredAction string                     `json:"required_action"`
	}{"cartulary.reporting_cutover_report.v1", e.ReasonCode(), e.categories, "Preserve a full backup and incident exports using the matching historical application. For disposable development data only, explicitly run make db-reset CARTULARY_DESTRUCTIVE_CONFIRM=db-reset. Historical backups remain tied to that application; reset does not convert them."})
	return string(b)
}

// PreflightCutover runs before any schema mutation. The migration repeats the
// same inventory under table locks; neither path invents render identities.
func PreflightCutover(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("reporting: cutover preflight database unavailable")
	}
	var current bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid=to_regclass('public.reporting_snapshots') AND conname='reporting_snapshot_identity_v1_ck')`).Scan(&current); err != nil {
		return err
	}
	if current {
		return nil
	}
	categories := []reportingCutoverCategory{}
	for _, item := range []struct{ category, table, query string }{
		{"historical_snapshots", "reporting_snapshots", `SELECT count(*) FROM public.reporting_snapshots`},
		{"retained_reporting_jobs", "reporting_job_payloads", `SELECT count(*) FROM public.reporting_job_payloads`},
		{"imported_reporting_artifacts", "reporting_imported_artifact_files", `SELECT count(*) FROM public.reporting_imported_artifact_files`},
		{"retained_composition_previews", "report_composition_preview_attempts", `SELECT count(*) FROM public.report_composition_preview_attempts`},
	} {
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, "public."+item.table).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			continue
		}
		var count int64
		if err := db.QueryRowContext(ctx, item.query).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			categories = append(categories, reportingCutoverCategory{item.category, count})
		}
	}
	if len(categories) > 0 {
		return &CutoverRequiredError{categories: categories}
	}
	return nil
}
