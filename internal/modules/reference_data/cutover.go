package reference_data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type cutoverCategory struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}
type CutoverRequiredError struct{ categories []cutoverCategory }

func (e *CutoverRequiredError) Error() string      { return "reference_pack_cutover_required" }
func (e *CutoverRequiredError) ReasonCode() string { return "reference_pack_cutover_required" }
func (e *CutoverRequiredError) RemediationReportJSON() string {
	data, _ := json.Marshal(struct {
		SchemaID       string            `json:"schema_id"`
		Reason         string            `json:"reason_code"`
		Categories     []cutoverCategory `json:"categories"`
		RequiredAction string            `json:"required_action"`
	}{"cartulary.reference_pack_cutover_report.v1", e.ReasonCode(), e.categories, "Preserve a full backup and incident exports using the matching historical application. For disposable development data only, explicitly run make db-reset CARTULARY_DESTRUCTIVE_CONFIRM=db-reset. Historical backups remain tied to that application; reset does not convert them."})
	return string(data)
}

// PreflightCutover reports only category counts, before the migration facade
// mutates schema. The SQL migration repeats admission under table locks; this
// read-only report is not a substitute for that transactional guard.
func PreflightCutover(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("reference pack: cutover preflight database unavailable")
	}
	var packsCurrent, indicatorsCurrent, portabilityCurrent, diagnosticsCurrent, operatorCurrent, dependenciesCurrent, requiredContentCurrent, legacyPresent, fallbackCurrent bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.reference_pack_candidates') IS NOT NULL,
        EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid=to_regclass('public.indicator_active_identities') AND attname='dedupe_sha256' AND NOT attisdropped),
        to_regclass('public.reference_pack_portable_catalogs') IS NOT NULL,
        EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid=to_regclass('public.reference_pack_attempt_members') AND attname='canonical_validation_summary' AND NOT attisdropped),
        EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid=to_regclass('public.jobs') AND conname='jobs_submitting_actor_ck'),
 to_regclass('public.reference_pack_operation_dependency_keys') IS NOT NULL,
 EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid=to_regclass('public.reference_pack_operations') AND conname='rp_portable_content_input_ck'),
 to_regclass('public.reference_packs') IS NOT NULL OR to_regclass('public.reference_pack_job_payloads') IS NOT NULL OR to_regclass('public.reference_pack_attestations') IS NOT NULL OR to_regclass('public.reference_pack_activation_state') IS NOT NULL,
 EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid=to_regclass('public.reference_pack_candidates') AND attname='fallback_from_version' AND NOT attisdropped)`).Scan(&packsCurrent, &indicatorsCurrent, &portabilityCurrent, &diagnosticsCurrent, &operatorCurrent, &dependenciesCurrent, &requiredContentCurrent, &legacyPresent, &fallbackCurrent); err != nil {
		return err
	}
	if packsCurrent && indicatorsCurrent && portabilityCurrent && diagnosticsCurrent && operatorCurrent && dependenciesCurrent && requiredContentCurrent && !legacyPresent && fallbackCurrent {
		return nil
	}
	categories := []cutoverCategory{}
	for _, item := range []struct{ category, table, query string }{
		{"unclassified_safety_fallbacks", "reference_pack_events", `SELECT count(*) FROM public.reference_pack_events WHERE event_kind='safety_fallback'`},
		{"unclassified_portable_content", "reference_pack_operations", `SELECT count(*) FROM public.reference_pack_operations WHERE kind='portable_retention'`},
		{"pending_dependency_captures", "reference_pack_operations", `SELECT count(*) FROM public.reference_pack_operations WHERE kind IN ('import','reverify','refresh') AND terminal_at IS NULL`},
		{"unproven_dependency_history", "reference_pack_versions", `SELECT count(*) FROM public.reference_pack_versions WHERE jsonb_array_length(convert_from(manifest_bytes,'UTF8')::jsonb->'dependencies')>0`},
		{"legacy_packs", "reference_packs", `SELECT count(*) FROM public.reference_packs`},
		{"legacy_pack_jobs", "reference_pack_job_payloads", `SELECT count(*) FROM public.reference_pack_job_payloads`},
		{"legacy_pack_attestations", "reference_pack_attestations", `SELECT count(*) FROM public.reference_pack_attestations`},
		{"legacy_pack_activations", "reference_pack_activation_state", `SELECT count(*) FROM public.reference_pack_activation_state`},
		{"retained_pack_jobs", "jobs", `SELECT count(*) FROM public.jobs WHERE job_kind LIKE 'reference_pack.%'`},
		{"indicator_identities", "indicators", `SELECT count(*) FROM public.indicators`},
		{"indicator_observations", "indicator_observations", `SELECT count(*) FROM public.indicator_observations`},
		{"retained_reporting_jobs", "jobs", `SELECT count(*) FROM public.jobs WHERE job_kind LIKE 'snapshot_reporting.%'`},
		{"retained_portability_jobs", "jobs", `SELECT count(*) FROM public.jobs WHERE job_kind LIKE 'incident_portability.%'`},
		{"historical_incident_bundles", "incident_bundle_exports", `SELECT count(*) FROM public.incident_bundle_exports`},
		{"historical_snapshots", "reporting_snapshots", `SELECT count(*) FROM public.reporting_snapshots`},
		{"historical_incident_reference_bindings", "incident_bundle_imported_actors", `SELECT count(*) FROM (
SELECT incident_id FROM public.incident_bundle_imported_actors
UNION SELECT incident_id FROM public.incident_bundle_imported_attributions
UNION SELECT imported_incident_id FROM public.incident_bundle_job_payloads WHERE imported_incident_id IS NOT NULL) imported`},
		{"retired_import_attribution", "jobs", `SELECT count(*) FROM public.jobs WHERE job_kind='reference_pack.import_v2' AND extension_idempotency_identity->>'schema_id' IS DISTINCT FROM 'cartulary.route_scoped_idempotency_identity.v2'`},
		{"historical_validation_summaries", "reference_pack_attempt_members", `SELECT count(*) FROM public.reference_pack_attempt_members WHERE verdict='content_rejected'`},
	} {
		// The cutover spans multiple append-only migrations. An intermediate
		// schema must still inventory identities before migration 48 runs;
		// completion of pack persistence alone does not establish that fact.
		retiredCategory := item.category == "legacy_packs" || item.category == "legacy_pack_jobs" || item.category == "legacy_pack_attestations" || item.category == "legacy_pack_activations"
		indicatorCategory := item.category == "indicator_identities" || item.category == "indicator_observations"
		portabilityCategory := item.category == "historical_incident_reference_bindings"
		diagnosticCategory := item.category == "historical_validation_summaries"
		operatorCategory := item.category == "retired_import_attribution"
		fallbackCategory := item.category == "unclassified_safety_fallbacks"
		contentCategory := item.category == "unclassified_portable_content"
		dependencyCategory := item.category == "pending_dependency_captures" || item.category == "unproven_dependency_history"
		if (fallbackCategory && fallbackCurrent) || (contentCategory && requiredContentCurrent) || (dependencyCategory && dependenciesCurrent) || (operatorCategory && operatorCurrent) || (indicatorCategory && indicatorsCurrent) || (portabilityCategory && portabilityCurrent) || (diagnosticCategory && diagnosticsCurrent) || (!retiredCategory && !fallbackCategory && !contentCategory && !dependencyCategory && !operatorCategory && !indicatorCategory && !portabilityCategory && !diagnosticCategory && packsCurrent) {
			continue
		}
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
			categories = append(categories, cutoverCategory{item.category, count})
		}
	}
	if len(categories) > 0 {
		return &CutoverRequiredError{categories: categories}
	}
	return nil
}
