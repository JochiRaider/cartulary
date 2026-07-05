import { existsSync } from "node:fs";

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
import { repoPath } from "./common.mjs";
import {
  exposedThemeCaptureSelector,
  frontendVisualFixtureRegistrySchemaID,
  phaseIDPattern,
  rowIDPattern,
  stableVisualCaptureSelectorPattern,
  validVisualCaptureScopeKinds,
  visualFixtureIDPattern,
  visualFixtureKeys,
  visualFixtureRegistryKeys,
  visualFullViewportCaptureScopeKeys,
  visualRegionCaptureScopeKeys,
  visualSelectorCaptureScopeKeys,
} from "./constants.mjs";
import {
  frontendVisualFixtureRegistryPath,
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "./registry-loader.mjs";

function requireFinite(value, label, { positive = false } = {}) {
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
  requireFinite(scope.x, `${label}.x`);
  requireFinite(scope.y, `${label}.y`);
  requireFinite(scope.width, `${label}.width`, { positive: true });
  requireFinite(scope.height, `${label}.height`, { positive: true });
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
