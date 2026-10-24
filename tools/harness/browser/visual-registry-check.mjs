#!/usr/bin/env node
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { atomicLocalFile } from "../runtime/secure-local-files.mjs";

const expectedFixtureIDs = [
  "visual.fixture.claimed_network_analysis_workspace_states",
  "visual.fixture.claimed_network_analysis_narrow_workspace",
  "visual.fixture.claimed_network_analysis_compact_workspace",
  "visual.fixture.base_inspector",
  "visual.fixture.inspector_narrow_technical_details",
  "visual.fixture.inspector_compact_actions",
  "visual.fixture.default_timeline_workbook_shell",
  "visual.fixture.compact_desktop_workbook_shell",
  "visual.fixture.destructive_actions",
  "visual.fixture.delayed_initial_loading",
  "visual.fixture.error_presentation_loci",
  "visual.fixture.drag_fill_handle",
  "visual.fixture.edit_cell",
  "visual.fixture.empty_successful_query",
  "visual.fixture.evidence_affordance",
  "visual.fixture.component_state_matrix",
  "visual.fixture.frozen_column",
  "visual.fixture.mention_chip_state_matrix",
  "visual.fixture.narrow_desktop_workbook_shell",
  "visual.fixture.resize_handle",
  "visual.fixture.presence_overflow",
  "visual.fixture.same_field_conflict",
  "visual.fixture.save_state_strip",
  "visual.fixture.saved_view_query_controls_and_grouped_result",
  "visual.fixture.task_requests_or_decisions",
  "visual.fixture.tree_group_row",
  "visual.fixture.indicator_lifecycle_authoring",
  "visual.fixture.indicator_observations_authoring",
  "visual.fixture.contextual_coordination_creation",
];
const expectedDesignIDs = Array.from({ length: 14 }, (_, index) => `D-VFIX-${String(index + 1).padStart(3, "0")}`);
const root = path.resolve(import.meta.dirname, "../../..");
const bytes = readFileSync(path.join(root, "tools/frontend_visual_fixture_registry.json"));
const registry = JSON.parse(bytes);
assert.equal(registry.schema_id, "cartulary.frontend_visual_fixture_registry.v6");
assert.equal(registry.owner_id, "harness.visual");
assert.equal(registry.verification_id, "harness.visual.verification.stable_fixture_identity");
const fixtureIDs = registry.fixtures.map((fixture) => fixture.fixture_id);
assert.deepEqual(fixtureIDs, expectedFixtureIDs);
const byID = new Map(registry.fixtures.map((fixture) => [fixture.fixture_id, fixture]));
assert.equal(byID.size, registry.fixtures.length, "fixture identities must be unique");
for (const fixture of registry.fixtures) {
  assert.equal(fixture.status, "current");
  for (const key of ["fixture_title", "playwright_scenario_title", "seed_id"]) assert.ok(fixture[key].length > 0, `${fixture.fixture_id}.${key}`);
  assert.ok(fixture.catalog_row_ids.length > 0);
  assert.deepEqual(Object.keys(fixture.capture_profiles).sort(), [...fixture.golden_artifacts].sort());
  for (const profile of Object.values(fixture.capture_profiles)) {
    assert.match(profile.viewport_css_px, /^[0-9]+x[0-9]+$/u);
    assert.ok(profile.device_scale_factor >= 1);
    assert.ok(profile.browser_zoom_percent >= 100);
    assert.equal(profile.theme_id, "dark_graphite");
    assert.ok(["compact", "default", "comfortable"].includes(profile.density_id));
  }
  assert.match(fixture.capture_scope.kind, /^(full_viewport|selector|region)$/u);
  assert.ok(fixture.scroll_normalization.kind.length > 0);
  assert.ok(Array.isArray(fixture.dynamic_masks));
  if (fixture.no_dynamic_regions) assert.deepEqual(fixture.dynamic_masks, []);
  assert.equal(fixture.blocked_reason, "");
  assert.equal(fixture.replacement_fixture_id, "");
  for (const artifact of fixture.golden_artifacts) assert.ok(existsSync(path.join(root, artifact)), `${artifact} should exist`);
  assert.ok(fixture.golden_artifacts.includes(fixture.golden_filename));
}
assert.deepEqual(registry.fixtures.map((fixture) => fixture.design_contract_id).filter(Boolean).sort(), expectedDesignIDs);
const shell = byID.get("visual.fixture.default_timeline_workbook_shell");
assert.equal(shell.fixture_title, "Default Timeline workbook shell");
assert.equal(shell.capture_scope.kind, "full_viewport");
const componentsRow = "web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c";
assert.ok(!shell.catalog_row_ids.includes(componentsRow));
const components = byID.get("visual.fixture.component_state_matrix");
assert.equal(components.fixture_title, "Component state matrix");
assert.equal(components.design_contract_id, "D-VFIX-009");
assert.deepEqual(components.catalog_row_ids, [componentsRow]);
assert.deepEqual(components.capture_scope, { kind: "selector", selector: "[data-design-fixture='components']" });
assert.equal(components.scroll_normalization.kind, "not_applicable");
assert.equal(Object.values(components.capture_profiles)[0].viewport_css_px, "1280x720");
assert.deepEqual(components.dynamic_masks, []);
assert.equal(components.no_dynamic_regions, true);
const registry_sha256 = createHash("sha256").update(bytes).digest("hex");
const results = process.env.CARTULARY_TEST_RESULTS_DIR, runID = process.env.CARTULARY_TEST_RUN_ID;
assert.ok(results && runID, "registry evidence requires the owning harness run");
const directory = path.resolve(root, results, runID, "visual-registry");
const payloads = {
  "visual-registry-metadata.json": {
    dynamic_masked_fixture_count: registry.fixtures.filter((fixture) => !fixture.no_dynamic_regions).length,
    fixture_count: registry.fixtures.length, fixture_ids: fixtureIDs, registry_sha256,
    scroll_anchor_fixture_count: registry.fixtures.filter((fixture) => fixture.scroll_normalization.anchor).length,
    selector_fixture_ids: registry.fixtures.filter((fixture) => fixture.capture_scope.kind === "selector").map((fixture) => fixture.fixture_id),
  },
  "visual-fixture-matrix.json": {
    fixture_ids: fixtureIDs, matrix_titles: fixtureIDs.map((id) => byID.get(id).fixture_title), registry_sha256,
    non_claim_boundaries: [
      "component-state evidence does not satisfy shell geometry fixtures",
      "current fixture metadata does not close browser.design-readiness without row accounting",
      "visual evidence remains non-publication evidence",
      "registry metadata validation does not execute a browser or establish rendering",
    ],
  },
};
for (const [file, payload] of Object.entries(payloads)) atomicLocalFile(path.join(directory, file), `${JSON.stringify({ evidence_kind: "static_metadata", ...payload }, null, 2)}\n`);
process.stdout.write(`Validated ${fixtureIDs.length} visual fixture declarations without services.\n`);
