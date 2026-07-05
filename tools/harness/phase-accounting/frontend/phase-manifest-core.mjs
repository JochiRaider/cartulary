import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";

import {
  loadTaskSurfaceManifest,
  targetEntryMap,
} from "../../generated-artifacts/task-surface.mjs";
import {
  assertObjectKeys,
  assertUnique,
  readJsonObject,
  requireBoolean,
  requireEnum,
  requireInteger,
  requireObject,
  requireObjectArray,
  requireRepoRelativePath,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "../../contract/json-shape.mjs";
import { frontendEvidenceAuditInputForTarget } from "./audit-routing.mjs";
import {
  frontendEvidenceFreshnessDigest,
  sha256File,
} from "./freshness.mjs";
import { collectFrontendGuideTargetRestatementErrors } from "./guide-restatements.mjs";
import { frontendVisualFixtureIDPattern } from "./phase-ids.mjs";

export const frontendPhaseNamespace = "frontend";
export const frontendPhaseRegistrySchemaID =
  "cartulary.frontend_phase_registry.v3";
export const frontendPhaseTestMapSchemaID =
  "cartulary.frontend_phase_test_map.v3";

const registryKeys = new Set([
  "schema_id",
  "phase_namespace",
  "guide_path",
  "guide_digest",
  "schema_version",
  "phases",
]);
const registryEntryKeys = new Set([
  "phase_id",
  "status",
  "row_rollup_state",
  "manifest_path",
  "manifest_digest",
  "ledger_path",
  "ledger_digest",
  "owner_refs",
  "depends_on",
  "activation_blockers",
  "evidence_freshness_digest",
]);
const mapKeys = new Set([
  "schema_id",
  "schema_version",
  "phase_namespace",
  "phase_id",
  "guide_digest",
  "rows",
]);
const rowKeys = new Set([
  "id",
  "phase_id",
  "layer",
  "evidence_class",
  "default_check_required",
  "default_check_kind",
  "default_check_reason_code",
  "future_default_check_candidate",
  "default_check_reason",
  "primary_evidence_owner",
  "duplicate_of",
  "evidence_delta",
  "warm_local_cost_class",
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
const ownerRefKeys = new Set([
  "source_key",
  "path",
  "section_ref",
  "heading_text",
  "req_ids",
  "ac_ids",
  "role",
  "resolution_status",
]);
const targetRefKeys = new Set([
  "target_name",
  "command_id",
  "evidence_role",
  "required_for_closure",
  "frontend_row_accounting_required",
  "scenario_title_required",
]);
const claimKeys = new Set([
  "statement",
  "claim_publication_intent",
  "closure_scope",
]);
const blockerKeys = new Set([
  "blocker_id",
  "reason_code",
  "description",
  "resolution_owner",
]);
const validStatuses = new Set(["planned", "active", "retired"]);
const validRowRollupStates = new Set([
  "no_rows_implemented",
  "partially_implemented",
  "implemented_dependency_blocked",
  "activation_ready",
  "active_green",
  "stale",
  "retired",
]);
const validEvidenceClasses = new Set([
  "product_conformance",
  "design_direction",
  "implementation_support",
  "claim_publication_boundary",
  "TODO_owner_lookup",
]);
const validClaimStatuses = new Set([
  "not_implemented",
  "implemented",
  "blocked",
  "stale",
  "retired",
]);
const validLayers = new Set([
  "unit",
  "integration",
  "browser_integration",
  "e2e",
  "visual",
  "accessibility",
  "support",
]);
const validOwnerSourceKeys = new Set([
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
const validOwnerRoles = new Set([
  "product_owner",
  "harness_owner",
  "design_owner",
  "support_owner",
  "rationale_only",
  "claim_publication_owner",
]);
const validOwnerResolutionStatuses = new Set([
  "resolved",
  "owner_lookup_required",
  "contradiction_detected",
  "suspected_owner_drift",
]);
const validEvidenceRoles = new Set([
  "primary",
  "supporting",
  "drift",
  "diagnostic_only",
]);
const validClaimPublicationIntents = new Set([
  "none",
  "informative_engineering_measurement",
  "claim_bearing_publication",
]);
const validClosureScopes = new Set([
  "scenario",
  "target_level",
  "blocked",
  "stale",
  "retired",
]);
const validDefaultCheckKinds = new Set([
  "primary_local_evidence",
  "default_local_cross_stack_conformance",
  "full_target_equivalent",
  "bounded_readiness",
  "explicit_only",
  "duplicate_regression",
  "future_candidate",
]);
const validDefaultCheckReasonCodes = new Set([
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
  "blocked_future_candidate",
]);
const validWarmLocalCostClasses = new Set([
  "none",
  "low",
  "medium",
  "service_backed",
  "browser",
  "explicit_heavy",
]);
const phaseIDPattern = /^FE-P(?:0|[1-9]\d*)$/;
const phaseMapFilenamePattern = /^fe_p(0|[1-9]\d*)_test_map\.json$/;
const phaseLedgerFilenamePattern = /^fe_p(0|[1-9]\d*)_coverage_ledger\.md$/;
const rowIDPattern = /^FE-(?:U|I|B|E|V|A11Y|S)-P(?:0|[1-9]\d*)-\d{2}$/;
const commandIDPattern =
  /^cartulary\.harness\.command\.[a-z0-9_]+\.v1$/;
const frontendVisualFixtureRegistrySchemaID =
  "cartulary.frontend_visual_fixture_registry.v3";
const visualFixtureRegistryKeys = new Set([
  "schema_id",
  "guide_path",
  "fixtures",
]);
const visualFixtureKeys = new Set([
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
const visualFixtureIDPattern = frontendVisualFixtureIDPattern;
const validVisualCaptureScopeKinds = new Set([
  "full_viewport",
  "selector",
  "region",
]);
const visualFullViewportCaptureScopeKeys = new Set(["kind"]);
const visualSelectorCaptureScopeKeys = new Set(["kind", "selector"]);
const visualRegionCaptureScopeKeys = new Set([
  "kind",
  "x",
  "y",
  "width",
  "height",
]);
const stableVisualCaptureSelectorPattern =
  /^\[[a-z0-9_-]*data-[a-z0-9_-]+=(?:"[^"]+"|'[^']+')\]$/u;
const exposedThemeCaptureSelector =
  "[data-design-fixture='exposed-theme']";
const core03SortingFilteringGroupingReqIDs = new Set(
  Array.from({ length: 13 }, (_, index) => `REQ-03-${223 + index}`),
);
const core03Section14OwnerRefPattern =
  /^Core 03 Section(?:s)?\b.*(?:^|[^\d.])14(?:[^\d.]|$)/;
const core03Section48OwnerRefPattern =
  /^Core 03 Section(?:s)?\b.*(?:^|[^\d.])4\.8(?:[^\d.]|$)/;
let cachedBaseAuthoritativePlaywrightTitles = null;

function repoPath(root, relativePath) {
  return path.join(root, relativePath);
}

function phaseNumber(phaseID) {
  const match = /^FE-P(0|[1-9]\d*)$/.exec(phaseID);
  if (!match) {
    throw new Error(`frontend phase id ${phaseID} must match FE-P<N>`);
  }
  return match[1];
}

function phaseFromMapPath(manifestPath, label) {
  const match = phaseMapFilenamePattern.exec(path.posix.basename(manifestPath));
  if (!match) {
    throw new Error(`${label} must end with fe_p<N>_test_map.json`);
  }
  return `FE-P${match[1]}`;
}

function phaseFromLedgerPath(ledgerPath, label) {
  const match = phaseLedgerFilenamePattern.exec(
    path.posix.basename(ledgerPath),
  );
  if (!match) {
    throw new Error(`${label} must end with fe_p<N>_coverage_ledger.md`);
  }
  return `FE-P${match[1]}`;
}

function requirePhaseID(value, label) {
  return requireString(value, label, { pattern: phaseIDPattern });
}

function entryTitles(entry) {
  if (Array.isArray(entry?.titles)) {
    return entry.titles.filter((title) => typeof title === "string");
  }
  return typeof entry?.title === "string" ? [entry.title] : [];
}

function baseAuthoritativePlaywrightTitleIndex(root = process.cwd()) {
  if (cachedBaseAuthoritativePlaywrightTitles) {
    return cachedBaseAuthoritativePlaywrightTitles;
  }
  const index = new Map();
  const toolsDir = repoPath(root, "tools");
  for (const filename of readdirSync(toolsDir).filter((name) =>
    /^phase[0-9]+_test_map\.json$/u.test(name),
  )) {
    const file = path.posix.join("tools", filename);
    const manifest = readJsonObject(repoPath(root, file), file);
    for (const value of Object.values(manifest)) {
      if (!Array.isArray(value)) {
        continue;
      }
      for (const entry of value) {
        if (
          entry?.runner !== "playwright" ||
          entry?.coverage !== "authoritative"
        ) {
          continue;
        }
        for (const title of entryTitles(entry)) {
          index.set(title, entry.id ?? file);
        }
      }
    }
  }
  cachedBaseAuthoritativePlaywrightTitles = index;
  return index;
}

function requiresBrowserScenarioTitleOwnership(row) {
  return (
    row.claim_status === "implemented" &&
    row.targets.some(
      (target) =>
        target.target_name.startsWith("browser-e2e") &&
        target.scenario_title_required,
    )
  );
}

function validateFrontendBrowserScenarioTitleOwnership(
  row,
  label,
  titleOwners,
) {
  if (!requiresBrowserScenarioTitleOwnership(row)) {
    return;
  }
  const baseTitles = baseAuthoritativePlaywrightTitleIndex();
  for (const title of row.scenario_titles) {
    const baseOwner = baseTitles.get(title);
    if (baseOwner) {
      throw new Error(
        `${label}.scenario_titles reuses base authoritative Playwright title owned by ${baseOwner}`,
      );
    }
    const existingOwner = titleOwners.get(title);
    if (existingOwner && existingOwner !== row.id) {
      throw new Error(
        `${label}.scenario_titles duplicates frontend browser title owned by ${existingOwner}`,
      );
    }
    if (!title.startsWith(`${row.id} `)) {
      throw new Error(
        `${label}.scenario_titles must use FE-owned titles prefixed by ${row.id}`,
      );
    }
    titleOwners.set(title, row.id);
  }
}

function requireFiniteNumber(value, label, { positive = false } = {}) {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw new Error(`${label} must be a finite number`);
  }
  if (positive && value <= 0) {
    throw new Error(`${label} must be greater than 0`);
  }
  return value;
}

function requireExactStringArray(value, expected, label) {
  const actual = requireStringArray(value, label);
  if (
    actual.length !== expected.length ||
    actual.some((entry, index) => entry !== expected[index])
  ) {
    throw new Error(`${label} must be exactly ${expected.join(", ")}`);
  }
  return actual;
}

function validateVisualCaptureScope(scopeValue, label) {
  const scope = requireObject(scopeValue, label);
  const kind = requireEnum(scope.kind, `${label}.kind`, validVisualCaptureScopeKinds);
  if (kind === "full_viewport") {
    assertObjectKeys(scope, visualFullViewportCaptureScopeKeys, label);
    return scope;
  }
  if (kind === "selector") {
    assertObjectKeys(scope, visualSelectorCaptureScopeKeys, label);
    const selector = requireString(scope.selector, `${label}.selector`);
    if (!stableVisualCaptureSelectorPattern.test(selector)) {
      throw new Error(
        `${label}.selector must be a stable data-attribute selector`,
      );
    }
    return scope;
  }
  assertObjectKeys(scope, visualRegionCaptureScopeKeys, label);
  requireFiniteNumber(scope.x, `${label}.x`);
  requireFiniteNumber(scope.y, `${label}.y`);
  requireFiniteNumber(scope.width, `${label}.width`, { positive: true });
  requireFiniteNumber(scope.height, `${label}.height`, { positive: true });
  return scope;
}

function validateVisualScrollNormalization(scroll, captureScope, label) {
  const kind = requireString(scroll.kind, `${label}.kind`);
  if (kind === "top_left") {
    requireString(scroll.anchor, `${label}.anchor`);
  }
  if (kind === "not_applicable") {
    requireString(scroll.reason, `${label}.reason`);
  }
  if (captureScope.kind === "selector" && scroll.anchor === "workbook-grid") {
    throw new Error(
      `${label}.anchor must not be workbook-grid for selector-only non-grid fixtures`,
    );
  }
}

function validateExposedThemeFixtureContract(fixture, label, captureScope) {
  if (fixture.status !== "current") {
    throw new Error(`${label}.status must be current for FE-VFIX-14`);
  }
  requireExactStringArray(fixture.owner_phase_ids, ["FE-P11"], `${label}.owner_phase_ids`);
  requireExactStringArray(fixture.owner_row_ids, ["FE-V-P11-03"], `${label}.owner_row_ids`);
  if (fixture.viewport_css_px !== "1280x720") {
    throw new Error(`${label}.viewport_css_px must be 1280x720 for FE-VFIX-14`);
  }
  if (fixture.golden_filename !== "apps/web/e2e/workbook.visual.spec.ts-snapshots/fe-v-p11-03-exposed-theme-states-linux.png") {
    throw new Error(`${label}.golden_filename must match FE-V-P11-03 exposed-theme golden`);
  }
  if (
    captureScope.kind !== "selector" ||
    captureScope.selector !== exposedThemeCaptureSelector
  ) {
    throw new Error(
      `${label}.capture_scope must select ${exposedThemeCaptureSelector}`,
    );
  }
  assertObjectKeys(
    fixture.scroll_normalization,
    new Set(["kind", "reason"]),
    `${label}.scroll_normalization`,
  );
  if (fixture.scroll_normalization.kind !== "not_applicable") {
    throw new Error(
      `${label}.scroll_normalization.kind must be not_applicable for FE-VFIX-14`,
    );
  }
  if (
    fixture.scroll_normalization.reason !==
    "selector-only exposed-theme specimen; no workbook-grid scroll state"
  ) {
    throw new Error(
      `${label}.scroll_normalization.reason must describe the selector-only exposed-theme specimen`,
    );
  }
  if (fixture.dynamic_masks.length !== 0) {
    throw new Error(`${label}.dynamic_masks must be empty for FE-VFIX-14`);
  }
  if (fixture.no_dynamic_regions !== true) {
    throw new Error(`${label}.no_dynamic_regions must be true for FE-VFIX-14`);
  }
}

function targetDisplayName(target) {
  return `make ${target.target_name}`;
}

function targetRefMatches(target, normalizedTarget) {
  return targetDisplayName(target) === normalizedTarget;
}

function validateFrontendGuideTargetRestatements(root, registry, rowTargetNames) {
  const absoluteGuidePath = repoPath(root, registry.guide_path);
  if (!existsSync(absoluteGuidePath)) {
    throw new Error(`frontend guide missing: ${registry.guide_path}`);
  }
  const errors = collectFrontendGuideTargetRestatementErrors(
    readFileSync(absoluteGuidePath, "utf8"),
    rowTargetNames,
    registry.guide_path,
  );
  if (errors.length > 0) {
    throw new Error(
      `frontend guide target restatement drift:\n${errors.join("\n")}`,
    );
  }
}

function sectionNumbersFromOwnerRef(sectionRef) {
  const normalized = sectionRef
    .replace(/^Core\s+\d+\s+Sections?\s+/i, "")
    .replace(/^Sections?\s+/i, "");
  if (/\bthrough\b/.test(normalized)) {
    throw new Error(`section_ref must list exact sections, got ${sectionRef}`);
  }
  return [...normalized.matchAll(/\d+(?:\.\d+)*[A-Z]?/g)].map(
    (match) => match[0],
  );
}

function coreHeadingMap(text) {
  const headings = new Map();
  for (const match of text.matchAll(/^#{2,6}\s+([0-9]+(?:\.[0-9]+)*[A-Z]?)\.?\s+(.+)$/gm)) {
    headings.set(match[1], `${match[1]} ${match[2].trim()}`);
  }
  return headings;
}

function validateOwnerRef(ownerRef, label, row) {
  assertObjectKeys(ownerRef, ownerRefKeys, label);
  const sourceKey = requireEnum(
    ownerRef.source_key,
    `${label}.source_key`,
    validOwnerSourceKeys,
  );
  const ownerPath = requireRepoRelativePath(ownerRef.path, `${label}.path`);
  requireString(ownerRef.section_ref, `${label}.section_ref`);
  requireString(ownerRef.heading_text, `${label}.heading_text`);
  requireStringArray(ownerRef.req_ids, `${label}.req_ids`);
  requireStringArray(ownerRef.ac_ids, `${label}.ac_ids`);
  requireEnum(ownerRef.role, `${label}.role`, validOwnerRoles);
  const resolutionStatus = requireEnum(
    ownerRef.resolution_status,
    `${label}.resolution_status`,
    validOwnerResolutionStatuses,
  );
  if (resolutionStatus !== "resolved" && row.claim_status !== "blocked") {
    throw new Error(
      `${label} unresolved owner refs are valid only on blocked rows`,
    );
  }
  if (sourceKey.startsWith("core") && resolutionStatus === "resolved") {
    const absolute = repoPath(process.cwd(), ownerPath);
    if (!existsSync(absolute)) {
      throw new Error(`${label}.path does not exist: ${ownerPath}`);
    }
    const text = readFileSync(absolute, "utf8");
    const sectionNumbers = sectionNumbersFromOwnerRef(ownerRef.section_ref);
    if (sectionNumbers.length === 0) {
      throw new Error(`${label}.section_ref must name at least one exact section`);
    }
    const headings = coreHeadingMap(text);
    const expectedHeadingText = sectionNumbers
      .map((sectionNumber) => {
        const heading = headings.get(sectionNumber);
        if (!heading) {
          throw new Error(
            `${label}.section_ref references missing section ${sectionNumber}`,
          );
        }
        return heading;
      })
      .join("; ");
    if (ownerRef.heading_text !== expectedHeadingText) {
      throw new Error(
        `${label}.heading_text must match resolved headings for ${ownerRef.section_ref}`,
      );
    }
    for (const reqID of ownerRef.req_ids) {
      if (!text.includes(`**${reqID}**`)) {
        throw new Error(`${label}.req_ids references missing ${reqID}`);
      }
    }
    for (const acID of ownerRef.ac_ids) {
      if (!text.includes(acID)) {
        throw new Error(`${label}.ac_ids references missing ${acID}`);
      }
    }
  }
  return ownerRef;
}

function validateTargetRef(target, label, row, targetEntriesByName = null) {
  assertObjectKeys(target, targetRefKeys, label);
  requireString(target.target_name, `${label}.target_name`);
  requireString(target.command_id, `${label}.command_id`, {
    pattern: commandIDPattern,
  });
  requireEnum(target.evidence_role, `${label}.evidence_role`, validEvidenceRoles);
  requireBoolean(target.required_for_closure, `${label}.required_for_closure`);
  requireBoolean(
    target.frontend_row_accounting_required,
    `${label}.frontend_row_accounting_required`,
  );
  requireBoolean(
    target.scenario_title_required,
    `${label}.scenario_title_required`,
  );
  if (targetEntriesByName) {
    const targetEntry = targetEntriesByName.get(target.target_name);
    if (!targetEntry) {
      throw new Error(
        `${label}.target_name must reference a task-surface target: ${target.target_name}`,
      );
    }
    if (targetEntry.command_id !== target.command_id) {
      throw new Error(
        `${label}.command_id must match task-surface ${target.target_name} command_id ${targetEntry.command_id}`,
      );
    }
  }
  if (
    row.claim_status === "implemented" &&
    target.required_for_closure &&
    !frontendEvidenceAuditInputForTarget(target.target_name)
  ) {
    throw new Error(
      `${label}.target_name ${target.target_name} is required for implemented row closure but has no frontend-evidence-audit retained-root route`,
    );
  }
  if (
    row.claim_status === "implemented" &&
    !["implementation_support", "claim_publication_boundary"].includes(
      row.evidence_class,
    ) &&
    target.required_for_closure &&
    target.scenario_title_required !== true
  ) {
    throw new Error(
      `${label}.scenario_title_required must be true for implemented non-support rows`,
    );
  }
  return target;
}

function validateClaim(claim, label, row) {
  assertObjectKeys(claim, claimKeys, label);
  requireString(claim.statement, `${label}.statement`);
  requireEnum(
    claim.claim_publication_intent,
    `${label}.claim_publication_intent`,
    validClaimPublicationIntents,
  );
  requireEnum(claim.closure_scope, `${label}.closure_scope`, validClosureScopes);
  if (
    row.claim_status === "implemented" &&
    !["scenario", "target_level"].includes(claim.closure_scope)
  ) {
    throw new Error(`${label}.closure_scope is incompatible with implemented`);
  }
  if (row.claim_status === "blocked" && claim.closure_scope !== "blocked") {
    throw new Error(`${label}.closure_scope must be blocked for blocked rows`);
  }
  return claim;
}

function validateBlocker(blocker, label) {
  assertObjectKeys(blocker, blockerKeys, label);
  requireString(blocker.blocker_id, `${label}.blocker_id`);
  requireString(blocker.reason_code, `${label}.reason_code`);
  requireString(blocker.description, `${label}.description`);
  requireString(blocker.resolution_owner, `${label}.resolution_owner`);
  return blocker;
}

function validateRowMetadata(row, label) {
  const evidenceClass = row.evidence_class;
  if (evidenceClass === "product_conformance") {
    if (row.core_req_ids.length === 0 || row.core_ac_ids.length === 0) {
      throw new Error(
        `${label} product_conformance rows must declare core_req_ids[] and core_ac_ids[]`,
      );
    }
    return;
  }
  if (
    evidenceClass === "design_direction" ||
    evidenceClass === "implementation_support"
  ) {
    if (row.core_req_ids.length !== 0 || row.core_ac_ids.length !== 0) {
      throw new Error(
        `${label} ${evidenceClass} rows must not declare Core requirement or acceptance IDs`,
      );
    }
    if (row.support_or_design_ac_ids.length === 0) {
      throw new Error(
        `${label} ${evidenceClass} rows must declare support_or_design_ac_ids[]`,
      );
    }
    return;
  }
  if (evidenceClass === "claim_publication_boundary") {
    for (const id of row.core_req_ids) {
      if (!id.startsWith("REQ-05-")) {
        throw new Error(
          `${label} claim_publication_boundary Core IDs must be Core 05 IDs`,
        );
      }
    }
    return;
  }
  if (evidenceClass === "TODO_owner_lookup") {
    if (row.core_req_ids.length !== 0 || row.core_ac_ids.length !== 0) {
      throw new Error(
        `${label} TODO_owner_lookup rows must not declare Core IDs`,
      );
    }
    if (row.claim_status !== "blocked") {
      throw new Error(
        `${label} TODO_owner_lookup rows must declare claim_status=blocked`,
      );
    }
  }
}

function validateVisualAccessibilityEvidenceBoundary(row, label) {
  const visualOrAccessibility =
    row.layer === "visual" ||
    row.layer === "accessibility" ||
    row.id.startsWith("FE-V-") ||
    row.id.startsWith("FE-A11Y-");
  if (!visualOrAccessibility) {
    return;
  }
  if (row.evidence_class === "product_conformance") {
    throw new Error(
      `${label} FE visual and accessibility rows must not use product_conformance; use design_direction, implementation_support, or an explicit Core 05 claim_publication_boundary route`,
    );
  }
  if (row.evidence_class !== "claim_publication_boundary") {
    return;
  }
  const hasCore05Owner = row.owner_refs.some(
    (ownerRef) =>
      ownerRef.source_key === "core05" &&
      ownerRef.role === "claim_publication_owner" &&
      ownerRef.resolution_status === "resolved",
  );
  if (!hasCore05Owner) {
    throw new Error(
      `${label} FE visual and accessibility claim_publication_boundary rows must cite a resolved Core 05 claim_publication_owner`,
    );
  }
  if (row.claim.claim_publication_intent !== "claim_bearing_publication") {
    throw new Error(
      `${label} FE visual and accessibility claim_publication_boundary rows must declare claim_publication_intent=claim_bearing_publication`,
    );
  }
}

function validateCore03SortingFilteringGroupingOwnerRefs(row, label) {
  const coversSortingFilteringGrouping = row.core_req_ids.some((id) =>
    core03SortingFilteringGroupingReqIDs.has(id),
  );
  if (!coversSortingFilteringGrouping) {
    return;
  }

  const citesCore03Section14 = row.owner_refs.some((ownerRef) =>
    ownerRef.source_key === "core03" &&
    (sectionNumbersFromOwnerRef(ownerRef.section_ref).includes("14") ||
      core03Section14OwnerRefPattern.test(ownerRef.heading_text)),
  );
  const citesCore03Section48 = row.owner_refs.some((ownerRef) =>
    ownerRef.source_key === "core03" &&
    (sectionNumbersFromOwnerRef(ownerRef.section_ref).includes("4.8") ||
      core03Section48OwnerRefPattern.test(ownerRef.heading_text)),
  );
  if (!citesCore03Section14 || citesCore03Section48) {
    throw new Error(
      `${label} rows covering REQ-03-223..REQ-03-235 must cite Core 03 Section 14 and must not cite Core 03 Section 4.8`,
    );
  }
}

export function frontendRegistryPath(root = process.cwd()) {
  return repoPath(root, "tools/frontend_phase_registry.json");
}

export function frontendVisualFixtureRegistryPath(root = process.cwd()) {
  return repoPath(root, "tools/frontend_visual_fixture_registry.json");
}

export function loadFrontendPhaseRegistry(root = process.cwd()) {
  const file = frontendRegistryPath(root);
  const registry = readJsonObject(file, file);
  assertObjectKeys(registry, registryKeys, file);
  requireSchemaID(registry, frontendPhaseRegistrySchemaID, file);
  if (registry.phase_namespace !== frontendPhaseNamespace) {
    throw new Error(
      `${file}.phase_namespace must be ${frontendPhaseNamespace}`,
    );
  }
  requireInteger(registry.schema_version, `${file}.schema_version`, { min: 3 });
  requireRepoRelativePath(registry.guide_path, `${file}.guide_path`, {
    extension: ".md",
  });
  requireString(registry.guide_digest, `${file}.guide_digest`);

  const rawPhases = requireObjectArray(registry.phases, `${file}.phases`, {
    nonEmpty: true,
  });

  const phases = rawPhases.map((entry, index) => {
    const label = `${file}.phases[${index + 1}]`;
    assertObjectKeys(entry, registryEntryKeys, label);
    const phaseID = requirePhaseID(entry.phase_id, `${label}.phase_id`);
    const status = requireEnum(entry.status, `${label}.status`, validStatuses);
    const rowRollupState = requireEnum(
      entry.row_rollup_state,
      `${label}.row_rollup_state`,
      validRowRollupStates,
    );
    const manifestPath = requireRepoRelativePath(
      entry.manifest_path,
      `${label}.manifest_path`,
      { extension: ".json" },
    );
    const ledgerPath = requireRepoRelativePath(
      entry.ledger_path,
      `${label}.ledger_path`,
      { extension: ".md" },
    );
    if (phaseFromMapPath(manifestPath, `${label}.manifest_path`) !== phaseID) {
      throw new Error(`${label}.manifest_path must match ${phaseID}`);
    }
    if (phaseFromLedgerPath(ledgerPath, `${label}.ledger_path`) !== phaseID) {
      throw new Error(`${label}.ledger_path must match ${phaseID}`);
    }
    return {
      phase_id: phaseID,
      status,
      row_rollup_state: rowRollupState,
      manifest_path: manifestPath,
      manifest_digest: requireString(
        entry.manifest_digest,
        `${label}.manifest_digest`,
      ),
      ledger_path: ledgerPath,
      ledger_digest: requireString(entry.ledger_digest, `${label}.ledger_digest`),
      owner_refs: requireObjectArray(entry.owner_refs, `${label}.owner_refs`, {
        nonEmpty: true,
      }).map((ownerRef, ownerIndex) =>
        validateOwnerRef(ownerRef, `${label}.owner_refs[${ownerIndex + 1}]`, {
          claim_status: "implemented",
        }),
      ),
      depends_on: requireStringArray(entry.depends_on, `${label}.depends_on`),
      activation_blockers: requireObjectArray(
        entry.activation_blockers,
        `${label}.activation_blockers`,
      ).map((blocker, blockerIndex) =>
        validateBlocker(
          blocker,
          `${label}.activation_blockers[${blockerIndex + 1}]`,
        ),
      ),
      evidence_freshness_digest: requireString(
        entry.evidence_freshness_digest,
        `${label}.evidence_freshness_digest`,
      ),
    };
  });

  assertUnique(
    phases.map((entry) => entry.phase_id),
    `${file}.phases.phase_id`,
  );
  const expected = Array.from(
    { length: phases.length },
    (_, index) => `FE-P${index}`,
  );
  const actual = phases
    .map((entry) => entry.phase_id)
    .sort(
      (left, right) =>
        Number(phaseNumber(left)) - Number(phaseNumber(right)),
    );
  if (actual.join(",") !== expected.join(",")) {
    throw new Error(
      `${file}.phases must contain contiguous frontend phases ${expected.join(", ")}`,
    );
  }
  const phaseIDs = new Set(actual);
  for (const entry of phases) {
    for (const dependency of entry.depends_on) {
      if (!phaseIDs.has(dependency)) {
        throw new Error(
          `${file} ${entry.phase_id}.depends_on references unknown ${dependency}`,
        );
      }
      if (
        Number(phaseNumber(dependency)) >= Number(phaseNumber(entry.phase_id))
      ) {
        throw new Error(
          `${file} ${entry.phase_id}.depends_on must reference earlier phases`,
        );
      }
    }
  }

  return {
    path: file,
    phase_namespace: registry.phase_namespace,
    guide_path: registry.guide_path,
    guide_digest: registry.guide_digest,
    phases: phases.sort(
      (left, right) =>
        Number(phaseNumber(left.phase_id)) -
        Number(phaseNumber(right.phase_id)),
    ),
  };
}

export function loadFrontendPhaseMap(root, phaseID) {
  const registry = loadFrontendPhaseRegistry(root);
  const entry = registry.phases.find(
    (candidate) => candidate.phase_id === phaseID,
  );
  if (!entry) {
    throw new Error(`unknown frontend phase ${phaseID}`);
  }
  const file = repoPath(root, entry.manifest_path);
  const manifest = readJsonObject(file, file);
  validateFrontendPhaseMap(manifest, file, phaseID);
  return { path: file, registryEntry: entry, manifest };
}

export function validateFrontendPhaseMap(
  manifest,
  label,
  expectedPhaseID = "",
  options = {},
) {
  assertObjectKeys(manifest, mapKeys, label);
  requireSchemaID(manifest, frontendPhaseTestMapSchemaID, label);
  requireInteger(manifest.schema_version, `${label}.schema_version`, { min: 3 });
  if (manifest.phase_namespace !== frontendPhaseNamespace) {
    throw new Error(
      `${label}.phase_namespace must be ${frontendPhaseNamespace}`,
    );
  }
  const phaseID = requirePhaseID(manifest.phase_id, `${label}.phase_id`);
  if (expectedPhaseID && phaseID !== expectedPhaseID) {
    throw new Error(`${label}.phase_id must be ${expectedPhaseID}`);
  }
  requireString(manifest.guide_digest, `${label}.guide_digest`);
  const rows = requireObjectArray(manifest.rows, `${label}.rows`, {
    nonEmpty: true,
  });
  const targetEntriesByName = options.targetEntriesByName ?? null;
  const ids = [];
  const frontendBrowserTitleOwners = new Map();
  for (const [index, row] of rows.entries()) {
    const rowLabel = `${label}.rows[${index + 1}]`;
    assertObjectKeys(row, rowKeys, rowLabel);
    ids.push(
      requireString(row.id, `${rowLabel}.id`, { pattern: rowIDPattern }),
    );
    if (requirePhaseID(row.phase_id, `${rowLabel}.phase_id`) !== phaseID) {
      throw new Error(`${rowLabel}.phase_id must be ${phaseID}`);
    }
    requireEnum(row.layer, `${rowLabel}.layer`, validLayers);
    requireEnum(
      row.evidence_class,
      `${rowLabel}.evidence_class`,
      validEvidenceClasses,
    );
    requireBoolean(row.default_check_required, `${rowLabel}.default_check_required`);
    requireEnum(row.default_check_kind, `${rowLabel}.default_check_kind`, validDefaultCheckKinds);
    requireEnum(
      row.default_check_reason_code,
      `${rowLabel}.default_check_reason_code`,
      validDefaultCheckReasonCodes,
    );
    requireString(row.primary_evidence_owner, `${rowLabel}.primary_evidence_owner`);
    requireString(row.duplicate_of, `${rowLabel}.duplicate_of`);
    requireString(row.evidence_delta, `${rowLabel}.evidence_delta`);
    requireEnum(row.warm_local_cost_class, `${rowLabel}.warm_local_cost_class`, validWarmLocalCostClasses);
    if (row.future_default_check_candidate !== undefined) {
      requireBoolean(
        row.future_default_check_candidate,
        `${rowLabel}.future_default_check_candidate`,
      );
    }
    if (row.default_check_reason !== undefined) {
      requireString(row.default_check_reason, `${rowLabel}.default_check_reason`);
    }
    requireObjectArray(row.owner_refs, `${rowLabel}.owner_refs`, {
      nonEmpty: true,
    }).forEach((ownerRef, ownerIndex) => {
      validateOwnerRef(
        ownerRef,
        `${rowLabel}.owner_refs[${ownerIndex + 1}]`,
        row,
      );
    });
    requireStringArray(row.core_req_ids, `${rowLabel}.core_req_ids`);
    requireStringArray(row.core_ac_ids, `${rowLabel}.core_ac_ids`);
    requireStringArray(
      row.support_or_design_ac_ids,
      `${rowLabel}.support_or_design_ac_ids`,
    );
    requireObjectArray(row.targets, `${rowLabel}.targets`, {
      nonEmpty: true,
    }).forEach((targetRef, targetIndex) => {
      validateTargetRef(
        targetRef,
        `${rowLabel}.targets[${targetIndex + 1}]`,
        row,
        targetEntriesByName,
      );
    });
    requireStringArray(row.scenario_titles, `${rowLabel}.scenario_titles`);
    const claimStatus = requireEnum(
      row.claim_status,
      `${rowLabel}.claim_status`,
      validClaimStatuses,
    );
    if (claimStatus === "blocked" && row.default_check_required === true) {
      throw new Error(
        `${rowLabel} blocked rows must not declare current default_check_required=true; use future_default_check_candidate for planned check placement`,
      );
    }
    const blockers = requireObjectArray(row.blockers, `${rowLabel}.blockers`).map(
      (blocker, blockerIndex) =>
        validateBlocker(blocker, `${rowLabel}.blockers[${blockerIndex + 1}]`),
    );
    if (claimStatus === "blocked" && blockers.length === 0) {
      throw new Error(`${rowLabel} blocked rows must declare blockers[]`);
    }
    if (claimStatus === "implemented" && blockers.length !== 0) {
      throw new Error(`${rowLabel} implemented rows must not declare blockers[]`);
    }
    if (
      row.future_default_check_candidate === true &&
      claimStatus !== "blocked"
    ) {
      throw new Error(
        `${rowLabel}.future_default_check_candidate is only valid for blocked rows`,
      );
    }
    if (
      row.default_check_required === true &&
      row.default_check_kind === "explicit_only"
    ) {
      throw new Error(`${rowLabel} default_check_required=true cannot use default_check_kind=explicit_only`);
    }
    if (
      row.default_check_required === false &&
      row.default_check_kind === "primary_local_evidence"
    ) {
      throw new Error(`${rowLabel} default_check_required=false cannot use primary_local_evidence`);
    }
    if (
      row.default_check_required === true &&
      (typeof row.default_check_reason !== "string" ||
        row.default_check_reason.trim() === "")
    ) {
      throw new Error(
        `${rowLabel} default-check frontend rows must declare default_check_reason`,
      );
    }
    validateClaim(row.claim, `${rowLabel}.claim`, row);
    requireStringArray(row.out_of_scope, `${rowLabel}.out_of_scope`);
    if (!row.id.includes(`-${phaseID.replace("FE-", "")}-`)) {
      throw new Error(`${rowLabel}.id must belong to ${phaseID}`);
    }
    if (
      row.targets.some(
        (target) =>
          target.target_name.startsWith("browser-e2e") &&
          target.scenario_title_required,
      ) &&
      row.scenario_titles.length === 0
    ) {
      throw new Error(
        `${rowLabel}.scenario_titles must be non-empty for scenario-backed browser rows`,
      );
    }
    if (
      row.targets.some((target) => target.scenario_title_required) &&
      claimStatus === "implemented" &&
      row.scenario_titles.length === 0
    ) {
      throw new Error(
        `${rowLabel}.scenario_titles must be non-empty when scenario_title_required=true`,
      );
    }
    if (
      claimStatus === "implemented" &&
      !row.targets.some((target) => target.required_for_closure)
    ) {
      throw new Error(
        `${rowLabel} implemented rows must have at least one required closure target`,
      );
    }
    if (
      row.layer === "accessibility" &&
      claimStatus === "implemented" &&
      row.targets.some(
        (target) =>
          target.target_name === "browser-e2e-a11y-preflight" &&
          (!target.required_for_closure ||
            !target.frontend_row_accounting_required ||
            !target.scenario_title_required),
      )
    ) {
      throw new Error(
        `${rowLabel} implemented accessibility preflight rows must require v3 row accounting and exact scenario closure`,
      );
    }
    validateRowMetadata(row, rowLabel);
    validateVisualAccessibilityEvidenceBoundary(row, rowLabel);
    validateCore03SortingFilteringGroupingOwnerRefs(row, rowLabel);
    validateFrontendBrowserScenarioTitleOwnership(
      row,
      rowLabel,
      frontendBrowserTitleOwners,
    );
  }
  assertUnique(ids, `${label}.rows.id`);
}

export function validateFrontendPhaseArtifacts(root = process.cwd(), options = {}) {
  const checkFreshness = options.checkFreshness !== false;
  const registry = loadFrontendPhaseRegistry(root);
  const targetEntriesByName =
    options.targetEntriesByName ??
    targetEntryMap(
      loadTaskSurfaceManifest(
        path.join(root, "tools", "task_surface_manifest.json"),
      ).manifest,
    );
  const phaseStates = new Map();
  const rowTargetNames = new Map();
  const frontendBrowserTitleOwners = new Map();
  const expectedGuideDigest = sha256File(root, registry.guide_path);
  if (
    checkFreshness &&
    expectedGuideDigest &&
    registry.guide_digest !== expectedGuideDigest
  ) {
    throw new Error(
      `${registry.path}.guide_digest must match ${registry.guide_path}`,
    );
  }
  for (const entry of registry.phases) {
    if (!existsSync(repoPath(root, entry.manifest_path))) {
      throw new Error(`frontend phase map missing: ${entry.manifest_path}`);
    }
    const manifest = readJsonObject(
      repoPath(root, entry.manifest_path),
      entry.manifest_path,
    );
    validateFrontendPhaseMap(manifest, entry.manifest_path, entry.phase_id, {
      targetEntriesByName,
    });
    for (const row of manifest.rows) {
      validateFrontendBrowserScenarioTitleOwnership(
        row,
        `${entry.manifest_path}.rows.${row.id}`,
        frontendBrowserTitleOwners,
      );
      rowTargetNames.set(
        row.id,
        new Set(row.targets.map((target) => target.target_name)),
      );
    }
    if (
      checkFreshness &&
      expectedGuideDigest &&
      manifest.guide_digest !== expectedGuideDigest
    ) {
      throw new Error(
        `${entry.manifest_path}.guide_digest must match ${registry.guide_path}`,
      );
    }
    const expectedManifestDigest = sha256File(root, entry.manifest_path);
    if (
      checkFreshness &&
      expectedManifestDigest &&
      entry.manifest_digest !== expectedManifestDigest
    ) {
      throw new Error(
        `${entry.phase_id}.manifest_digest must match ${entry.manifest_path}`,
      );
    }
    const expectedLedgerDigest = sha256File(root, entry.ledger_path);
    if (
      checkFreshness &&
      expectedLedgerDigest &&
      entry.ledger_digest !== expectedLedgerDigest
    ) {
      throw new Error(
        `${entry.phase_id}.ledger_digest must match ${entry.ledger_path}`,
      );
    }
    const expectedFreshnessDigest = frontendEvidenceFreshnessDigest(
      root,
      registry,
      entry,
    );
    if (
      checkFreshness &&
      entry.evidence_freshness_digest !== expectedFreshnessDigest
    ) {
      throw new Error(
        `${entry.phase_id}.evidence_freshness_digest must match frontend freshness inputs`,
      );
    }
    const rowRollupState = computeRowRollupState(entry, manifest, phaseStates);
    phaseStates.set(entry.phase_id, rowRollupState);
    if (entry.row_rollup_state !== rowRollupState) {
      throw new Error(
        `${entry.phase_id}.row_rollup_state must be ${rowRollupState}, got ${entry.row_rollup_state}`,
      );
    }
    if (entry.status === "active" && rowRollupState !== "active_green") {
      throw new Error(
        `${entry.phase_id} active phases must have row_rollup_state=active_green`,
      );
    }
    if (
      entry.status === "planned" &&
      rowRollupState === "activation_ready" &&
      entry.activation_blockers.length === 0
    ) {
      throw new Error(
        `${entry.phase_id} activation-ready planned phases must declare activation_blockers[] or be promoted active`,
      );
    }
    if (
      entry.status === "active" &&
      manifest.rows.some((row) => row.claim_status === "blocked")
    ) {
      throw new Error(
        `${entry.phase_id} is active but contains blocked frontend rows`,
      );
    }
  }
  const mapDir = repoPath(root, "tools/frontend_phase_maps");
  for (const filename of readdirSync(mapDir).filter((name) =>
    name.endsWith(".json"),
  )) {
    const file = path.posix.join("tools/frontend_phase_maps", filename);
    const phaseID = phaseFromMapPath(file, file);
    if (!registry.phases.some((entry) => entry.phase_id === phaseID)) {
      throw new Error(`unregistered frontend phase map: ${file}`);
    }
  }
  validateFrontendGuideTargetRestatements(root, registry, rowTargetNames);
  validateFrontendVisualFixtureRegistry(root);
}

function computeRowRollupState(entry, manifest, priorPhaseStates) {
  if (entry.status === "retired") {
    return "retired";
  }
  const nonRetiredRows = manifest.rows.filter(
    (row) => row.claim_status !== "retired",
  );
  if (nonRetiredRows.length === 0) {
    return "retired";
  }
  if (nonRetiredRows.some((row) => row.claim_status === "stale")) {
    return "stale";
  }
  const implementedRows = nonRetiredRows.filter(
    (row) => row.claim_status === "implemented",
  );
  if (implementedRows.length === 0) {
    return "no_rows_implemented";
  }
  if (implementedRows.length !== nonRetiredRows.length) {
    return "partially_implemented";
  }
  const dependenciesGreen = entry.depends_on.every((phaseID) =>
    ["active_green", "activation_ready"].includes(priorPhaseStates.get(phaseID)),
  );
  if (!dependenciesGreen) {
    return "implemented_dependency_blocked";
  }
  return entry.status === "active" ? "active_green" : "activation_ready";
}

export function validateFrontendVisualFixtureRegistry(root = process.cwd()) {
  const file = frontendVisualFixtureRegistryPath(root);
  if (!existsSync(file)) {
    throw new Error("frontend visual fixture registry missing: tools/frontend_visual_fixture_registry.json");
  }
  const registry = readJsonObject(file, file);
  assertObjectKeys(registry, visualFixtureRegistryKeys, file);
  requireSchemaID(registry, frontendVisualFixtureRegistrySchemaID, file);
  if (registry.guide_path !== "docs/guides/cartulary_frontend_implementation_testing_guide.md") {
    throw new Error(`${file}.guide_path must point to the frontend guide`);
  }
  const fixtures = requireObjectArray(registry.fixtures, `${file}.fixtures`, {
    nonEmpty: true,
  });
  const fixtureIDs = [];
  const rowIDs = new Set();
  for (const phase of loadFrontendPhaseRegistry(root).phases) {
    const { manifest } = loadFrontendPhaseMap(root, phase.phase_id);
    for (const row of manifest.rows) {
      rowIDs.add(row.id);
    }
  }
  for (const [index, fixture] of fixtures.entries()) {
    const label = `${file}.fixtures[${index + 1}]`;
    assertObjectKeys(fixture, visualFixtureKeys, label);
    const fixtureID = requireString(fixture.fixture_id, `${label}.fixture_id`, {
      pattern: visualFixtureIDPattern,
    });
    fixtureIDs.push(fixtureID);
    requireString(fixture.fixture_title, `${label}.fixture_title`);
    const status = requireEnum(
      fixture.status,
      `${label}.status`,
      new Set(["current", "missing", "retired"]),
    );
    requireStringArray(fixture.owner_phase_ids, `${label}.owner_phase_ids`, {
      nonEmpty: true,
      pattern: phaseIDPattern,
    });
    const ownerRowIDs = requireStringArray(
      fixture.owner_row_ids,
      `${label}.owner_row_ids`,
      { nonEmpty: true, pattern: rowIDPattern },
    );
    for (const rowID of ownerRowIDs) {
      if (!rowIDs.has(rowID)) {
        throw new Error(`${label}.owner_row_ids references unknown ${rowID}`);
      }
    }
    if (status === "current") {
      requireString(fixture.playwright_scenario_title, `${label}.playwright_scenario_title`);
      const golden = requireRepoRelativePath(
        fixture.golden_filename,
        `${label}.golden_filename`,
        { extension: ".png" },
      );
      if (!existsSync(repoPath(root, golden))) {
        throw new Error(`${label}.golden_filename does not exist: ${golden}`);
      }
      const goldenArtifacts = requireStringArray(
        fixture.golden_artifacts,
        `${label}.golden_artifacts`,
        { nonEmpty: true },
      );
      if (!goldenArtifacts.includes(fixture.golden_filename)) {
        throw new Error(`${label}.golden_artifacts must include golden_filename`);
      }
      for (const [artifactIndex, artifact] of goldenArtifacts.entries()) {
        const artifactLabel = `${label}.golden_artifacts[${artifactIndex + 1}]`;
        const artifactPath = requireRepoRelativePath(artifact, artifactLabel, {
          extension: ".png",
        });
        if (!existsSync(repoPath(root, artifactPath))) {
          throw new Error(`${artifactLabel} does not exist: ${artifactPath}`);
        }
      }
      requireString(fixture.seed_id, `${label}.seed_id`);
    } else {
      requireString(fixture.blocked_reason, `${label}.blocked_reason`);
    }
    requireString(fixture.viewport_css_px, `${label}.viewport_css_px`);
    requireInteger(fixture.device_scale_factor, `${label}.device_scale_factor`, { min: 1 });
    requireInteger(fixture.browser_zoom_percent, `${label}.browser_zoom_percent`, { min: 1 });
    requireString(fixture.theme_id, `${label}.theme_id`);
    requireString(fixture.density_id, `${label}.density_id`);
    const scrollNormalization = requireObject(
      fixture.scroll_normalization,
      `${label}.scroll_normalization`,
    );
    const captureScope = validateVisualCaptureScope(
      fixture.capture_scope,
      `${label}.capture_scope`,
    );
    validateVisualScrollNormalization(
      scrollNormalization,
      captureScope,
      `${label}.scroll_normalization`,
    );
    requireObject(fixture.focus_state, `${label}.focus_state`);
    requireObject(fixture.editor_state, `${label}.editor_state`);
    requireObject(fixture.inspector_state, `${label}.inspector_state`);
    const dynamicMasks = requireStringArray(
      fixture.dynamic_masks,
      `${label}.dynamic_masks`,
    );
    const noDynamicRegions = requireBoolean(
      fixture.no_dynamic_regions,
      `${label}.no_dynamic_regions`,
    );
    if (noDynamicRegions && dynamicMasks.length > 0) {
      throw new Error(
        `${label}.no_dynamic_regions=true requires empty dynamic_masks`,
      );
    }
    if (fixtureID === "FE-VFIX-14") {
      validateExposedThemeFixtureContract(fixture, label, captureScope);
    }
    requireString(fixture.replacement_fixture_id || "none", `${label}.replacement_fixture_id`);
  }
  assertUnique(fixtureIDs, `${file}.fixtures.fixture_id`);
  const expected = Array.from({ length: fixtures.length }, (_, index) =>
    `FE-VFIX-${String(index + 1).padStart(2, "0")}`,
  );
  if (fixtureIDs.sort().join(",") !== expected.join(",")) {
    throw new Error(`${file}.fixtures must contain contiguous visual fixture IDs ${expected.join(", ")}`);
  }
}
