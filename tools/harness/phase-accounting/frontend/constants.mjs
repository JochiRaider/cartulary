import { frontendVisualFixtureIDPattern } from "./phase-ids.mjs";

export const frontendPhaseNamespace = "frontend";
export const frontendPhaseRegistrySchemaID =
  "cartulary.frontend_phase_registry.v5";
export const frontendPhaseTestMapSchemaID =
  "cartulary.frontend_phase_test_map.v4";
export const frontendVisualFixtureRegistrySchemaID =
  "cartulary.frontend_visual_fixture_registry.v3";

export const registryKeys = new Set([
  "schema_id",
  "phase_namespace",
  "guide_path",
  "guide_digest",
  "schema_version",
  "phases",
]);
export const registryEntryKeys = new Set([
  "phase_id",
  "status",
  "row_rollup_state",
  "base_phase_join",
  "manifest_path",
  "manifest_digest",
  "ledger_path",
  "ledger_digest",
  "owner_refs",
  "depends_on",
  "activation_blockers",
  "evidence_freshness_digest",
]);
export const mapKeys = new Set([
  "schema_id",
  "schema_version",
  "phase_namespace",
  "phase_id",
  "guide_digest",
  "rows",
]);
export const rowKeys = new Set([
  "id",
  "phase_id",
  "layer",
  "evidence_class",
  "default_check_required",
  "default_check_kind",
  "default_check_reason_code",
  "default_check_reason",
  "primary_evidence_owner",
  "duplicate_of",
  "evidence_delta",
  "warm_local_cost_class",
  "runtime_profile_id",
  "owner_refs",
  "core_req_ids",
  "core_ac_ids",
  "support_or_design_ac_ids",
  "targets",
  "scenario_titles",
  "claim_status",
  "claim",
  "blockers",
  "out_of_scope",
]);
export const ownerRefKeys = new Set([
  "source_key",
  "path",
  "section_ref",
  "heading_text",
  "req_ids",
  "ac_ids",
  "role",
  "resolution_status",
]);
export const targetRefKeys = new Set([
  "target_name",
  "command_id",
  "evidence_role",
  "required_for_closure",
  "frontend_row_accounting_required",
  "scenario_title_required",
]);
export const claimKeys = new Set([
  "statement",
  "claim_publication_intent",
  "closure_scope",
]);
export const blockerKeys = new Set([
  "blocker_id",
  "reason_code",
  "description",
  "resolution_owner",
]);

export const validStatuses = new Set(["active"]);
export const validRowRollupStates = new Set(["active_green"]);
export const validEvidenceClasses = new Set([
  "product_conformance",
  "design_direction",
  "implementation_support",
  "claim_publication_boundary",
]);
export const validClaimStatuses = new Set(["implemented"]);
export const validLayers = new Set([
  "unit",
  "integration",
  "browser_integration",
  "e2e",
  "visual",
  "accessibility",
  "support",
]);
export const validOwnerSourceKeys = new Set([
  "core00",
  "core01",
  "core02",
  "core03",
  "core04",
  "core05",
  "testing_harness_nlspec",
  "design_md",
  "ui_ux_guide",
  "visual_golden_guide",
  "dev_guide",
  "implementation_testing_guide",
  "research_rationale",
  "guide_local",
]);
export const validOwnerRoles = new Set([
  "product_owner",
  "harness_owner",
  "design_owner",
  "support_owner",
  "rationale_only",
  "claim_publication_owner",
]);
export const validOwnerResolutionStatuses = new Set([
  "resolved",
  "owner_lookup_required",
  "contradiction_detected",
  "suspected_owner_drift",
]);
export const validEvidenceRoles = new Set([
  "primary",
  "supporting",
  "drift",
  "diagnostic_only",
]);
export const validClaimPublicationIntents = new Set([
  "none",
  "informative_engineering_measurement",
  "claim_bearing_publication",
]);
export const validClosureScopes = new Set(["scenario", "target_level"]);
export const validDefaultCheckKinds = new Set([
  "primary_local_evidence",
  "default_local_cross_stack_conformance",
  "full_target_equivalent",
  "bounded_readiness",
  "explicit_only",
  "duplicate_regression",
]);
export const validDefaultCheckReasonCodes = new Set([
  "cheapest_authoritative_layer",
  "lower_layer_gap",
  "full_target_equivalent_stateful",
  "bounded_readiness",
  "explicit_full_target",
  "explicit_readiness",
  "explicit_measurement",
  "implementation_support_explicit_only",
  "design_direction_explicit_only",
  "claim_publication_boundary",
  "duplicate_of_primary_owner",
]);
export const validWarmLocalCostClasses = new Set([
  "none",
  "low",
  "medium",
  "service_backed",
  "browser",
  "explicit_heavy",
]);

export const phaseIDPattern = /^FE-P(?:0|[1-9]\d*)$/;
export const phaseMapFilenamePattern = /^fe_p(0|[1-9]\d*)_test_map\.json$/;
export const phaseLedgerFilenamePattern =
  /^fe_p(0|[1-9]\d*)_coverage_ledger\.md$/;
export const rowIDPattern =
  /^FE-(?:U|I|B|E|V|A11Y|S)-P(?:0|[1-9]\d*)-\d{2}$/;
export const commandIDPattern =
  /^cartulary\.harness\.command\.[a-z0-9_]+\.v1$/;

export const visualFixtureRegistryKeys = new Set([
  "schema_id",
  "guide_path",
  "fixtures",
]);
export const visualFixtureKeys = new Set([
  "fixture_id",
  "fixture_title",
  "status",
  "owner_phase_ids",
  "owner_row_ids",
  "playwright_scenario_title",
  "golden_filename",
  "golden_artifacts",
  "seed_id",
  "viewport_css_px",
  "device_scale_factor",
  "browser_zoom_percent",
  "theme_id",
  "density_id",
  "scroll_normalization",
  "capture_scope",
  "focus_state",
  "editor_state",
  "inspector_state",
  "dynamic_masks",
  "no_dynamic_regions",
  "blocked_reason",
  "replacement_fixture_id",
]);
export const visualFixtureIDPattern = frontendVisualFixtureIDPattern;
export const validVisualCaptureScopeKinds = new Set([
  "full_viewport",
  "selector",
  "region",
]);
export const visualFullViewportCaptureScopeKeys = new Set(["kind"]);
export const visualSelectorCaptureScopeKeys = new Set(["kind", "selector"]);
export const visualRegionCaptureScopeKeys = new Set([
  "kind",
  "x",
  "y",
  "width",
  "height",
]);
export const stableVisualCaptureSelectorPattern =
  /^\[[a-z0-9_-]*data-[a-z0-9_-]+=(?:"[^"]+"|'[^']+')\]$/u;
export const exposedThemeCaptureSelector =
  "[data-design-fixture='exposed-theme']";

export const core03SortingFilteringGroupingReqIDs = new Set(
  Array.from({ length: 13 }, (_, index) => `REQ-03-${223 + index}`),
);
export const core03Section14OwnerRefPattern =
  /^Core 03 Section(?:s)?\b.*(?:^|[^\d.])14(?:[^\d.]|$)/;
export const core03Section48OwnerRefPattern =
  /^Core 03 Section(?:s)?\b.*(?:^|[^\d.])4\.8(?:[^\d.]|$)/;
