import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import { resolveCommandInputs, validateSchemaSync } from "../../contract/index.mjs";
import { loadExecutionTopology } from "../../generated-artifacts/execution-topology.mjs";
import { loadTestCatalog } from "../../test-catalog/index.mjs";

const root = new URL("../../../../", import.meta.url).pathname;

test("v16 rejects ownership partitions and preserves exact cross-owner selection", async () => {
  const { normalizeBrowserBatchStages } = await import("../browser-batch-manifest.mjs");
  const { browserCohort } = await import("../browser-cohorts.mjs");
  const manifest = JSON.parse(readFileSync(new URL("../../../browser_e2e_batch_manifest.json", import.meta.url), "utf8"));
  assert.equal(manifest.schema_id, "cartulary.browser_e2e_batch_manifest.v16");
  validateSchemaSync(manifest.schema_id, manifest);
  assert.throws(() => normalizeBrowserBatchStages({ ...manifest, schema_id: "cartulary.browser_e2e_batch_manifest.v14" }), /schema/);
  const legacyOwner = structuredClone(manifest);
  legacyOwner.stages[0].groups[0].owner_id = "harness.browser";
  assert.throws(() => normalizeBrowserBatchStages(legacyOwner), /additional properties/);
  assert.throws(() => validateSchemaSync(manifest.schema_id, legacyOwner));
  const stages = normalizeBrowserBatchStages(manifest);
  const { rowByID } = loadTestCatalog(root);
  const stage = stages.get("support");
  const partition = stage.groups.find((group) => new Set(group.selectedRowIDs.map((id) => rowByID.get(id).owner_id)).size > 1);
  assert.ok(partition);
  assert.equal(Object.hasOwn(partition, "ownerID"), false);
  const all = partition.selectedRowIDs.map((id) => rowByID.get(id));
  const selected = [all[0], all.find((row) => row.owner_id !== all[0].owner_id)];
  const cohort = browserCohort(partition, selected, stage.invocationPolicy);
  assert.deepEqual(cohort.selectedRowIDs, selected.map((row) => row.row_id).sort());
  assert.equal(cohort.name, browserCohort(partition, [...selected].reverse(), stage.invocationPolicy).name);
  for (const group of stages.values()) for (const partition of group.groups) {
    const rows = partition.selectedRowIDs.map((id) => rowByID.get(id));
    // Admission uses the exact same tuple as the generator for every catalog row.
    for (const row of rows) browserCohort(partition, [row], group.invocationPolicy);
    if (partition.stateScope !== "test_owned") assert.equal(rows.length, 1);
  }
});

test("browser manifest admission rejects missing and mistyped execution metadata", async () => {
  const { normalizeBrowserBatchStages, validateBrowserBatchManifestShape } = await import("../browser-batch-manifest.mjs");
  const manifest = JSON.parse(readFileSync(new URL("../../../browser_e2e_batch_manifest.json", import.meta.url), "utf8"));
  for (const field of ["runtime_profile_id", "resource_profile_id", "fixture_capability", "project", "config",
    "state_scope", "postgres_isolation", "workers", "preparation_operations", "selected_row_ids", "specs",
    "service_requirement", "service_dependencies", "browser_session_group"]) {
    const missing = structuredClone(manifest);
    delete missing.partitions[0][field];
    assert.throws(() => validateBrowserBatchManifestShape(missing, "missing execution metadata"), /validation failed/, field);
    assert.throws(() => normalizeBrowserBatchStages(missing), /validation failed/, field);
  }
  for (const [field, value] of [["runtime_profile_id", null], ["resource_profile_id", 12],
    ["project", ["chromium"]], ["workers", 1], ["preparation_operations", {}],
    ["service_dependencies", "postgres"], ["selected_row_ids", []], ["specs", []]]) {
    const malformed = structuredClone(manifest);
    malformed.partitions[0][field] = value;
    assert.throws(() => validateBrowserBatchManifestShape(malformed, "mistyped execution metadata"), /validation failed/, field);
  }
});

test("topology resolves four or six lanes and rejects invalid public controls", () => {
  const { raw, taskSurface } = loadExecutionTopology({ root });
  for (const target of ["test-fast", "check", "ci", "test-slice", "browser-e2e", "browser-e2e-visual-update"]) {
    const scope = target === "test-slice" ? { OWNER: "harness.browser" } : target === "browser-e2e" ? { SUITE: "functional" } : {};
    assert.equal(resolveCommandInputs(target, scope, taskSurface).BROWSER_LANES, "4");
    assert.equal(resolveCommandInputs(target, { ...scope, BROWSER_LANES: "6" }, taskSurface).BROWSER_LANES, "6");
    for (const value of ["0", "7", "1.5", "six", ""]) assert.throws(() => resolveCommandInputs(target, { ...scope, BROWSER_LANES: value }, taskSurface));
    assert.throws(() => resolveCommandInputs(target, { ...scope, PLAYWRIGHT_WORKERS: "1" }, taskSurface));
  }
  for (const stage of raw.browser_e2e_batch.stages) for (const group of stage.groups) assert.equal(group.workers, "1");
});

test("catalog rows declare an aggregate placement", () => {
  const [row] = loadTestCatalog(root).rows;
  const family = (rows) => ({ schema_id: "cartulary.test_family_manifest.v15", owner_id: row.owner_id, rows });
  validateSchemaSync("cartulary.test_family_manifest.v15", family([row]));
  const legacy = { ...row, minimum_tier: "fast" };
  delete legacy.aggregate_placement;
  assert.throws(() => validateSchemaSync("cartulary.test_family_manifest.v15", family([legacy])));
});

test("browser cohorts bind exact cases and execution compatibility independently of reporting ownership", async () => {
  const { browserCohort } = await import("../browser-cohorts.mjs");
  const rows = ["a", "b", "c"].map((id) => ({ row_id: `fixture.${id}`, owner_id: `fixture.owner_${id}`,
    runner: "playwright", browser_state_scope: "test_owned", runtime_profile_id: "default",
    fixture_capability: "browser_stack", postgres_isolation: "exclusive",
    resource_profile_id: "browser_ordinary", service_dependencies: ["postgres"], evidence_class: "browser",
    selector: { file: `apps/web/e2e/${id}.spec.ts`, project_id: "chromium", stage: "webserver_backed", titles: ["repeated title"] } }));
  const partition = { name: "fixture", stateScope: "test_owned", stageName: "webserver-backed",
    stageTarget: "browser-e2e-webserver-backed", frontendArtifactID: "production", stageScheduleTags: [], stageSchedulerNeeds: [],
    target: "browser-e2e-webserver-backed", kind: "duration_balanced_specs", coverage: "authoritative",
    executionDependency: "browser_ordinary", dependencyTarget: "", workers: "1", config: "playwright.config.ts",
    project: "chromium", runtimeProfileID: "default", runtimeKind: "default", keyRingManifestPath: "",
    resourceProfileID: "browser_ordinary", fixtureCapability: "browser_stack", postgresIsolation: "exclusive",
    fixtureProfileID: "", serviceRequirement: "test-services", serviceDependencies: ["postgres"], preparationOperations: [],
    browserSessionGroup: "fixture-session", selectedRowIDs: rows.map((row) => row.row_id),
    specs: rows.map((row) => row.selector.file) };
  const policy = { startup_ms: 20, cleanup_ms: 10 };
  const subset = browserCohort(partition, [rows[1]], policy);
  for (const key of ["frontendArtifactID", "config", "runtimeKind", "keyRingManifestPath", "serviceRequirement"]) {
    assert.notEqual(subset.name, browserCohort({ ...partition, [key]: "changed" }, [rows[1]], policy).name, key);
  }
  for (const key of ["stageName", "stageTarget", "target", "coverage", "browserSessionGroup", "executionDependency", "dependencyTarget"]) {
    assert.equal(subset.name, browserCohort({ ...partition, [key]: "changed" }, [rows[1]], policy).name, key);
  }
  assert.equal(subset.name, browserCohort({ ...partition, stageScheduleTags: ["changed"] }, [rows[1]], policy).name);
  assert.notEqual(subset.name, browserCohort({ ...partition, stageSchedulerNeeds: ["changed"] }, [rows[1]], policy).name);
  assert.equal(subset.name, browserCohort(partition, [{ ...rows[1], selector: { ...rows[1].selector, stage: "stateful" } }], policy).name);
  assert.deepEqual(subset.specs, [rows[1].selector.file]);
  assert.notEqual(subset.name, browserCohort(partition, rows.slice(0, 2), policy).name);
  for (const key of ["startup_ms", "cleanup_ms"]) {
    assert.notEqual(subset.name, browserCohort(partition, [rows[1]], { ...policy, [key]: 21 }).name);
  }
  assert.equal(subset.name, browserCohort({ ...partition, name: "another-label", selectedRowIDs: ["unselected", ...partition.selectedRowIDs].reverse(),
    specs: ["apps/web/e2e/unselected.spec.ts", ...partition.specs].reverse() }, [{ ...rows[1], owner_id: "changed.owner" }], policy).name);
  for (const changed of [
    { ...rows[0], preparation_operations: ["harness.frontend_install"] },
    { ...rows[0], fixture_capability: "postgres_dedicated" }, { ...rows[0], fixture_profile_id: "changed_v1" },
    { ...rows[0], postgres_isolation: "shared" }, { ...rows[0], browser_state_scope: "isolated" },
    { ...rows[0], selector: { ...rows[0].selector, project_id: "other" } },
    { ...rows[0], service_dependencies: [] }, { ...rows[0], resource_profile_id: "browser_measurement_quiet" },
    { ...rows[0], runtime_profile_id: "network_flow_claimed" },
  ]) assert.throws(() => browserCohort(partition, [changed], policy), /compatibility/);
  for (const scope of ["isolated", "measurement"]) {
    assert.throws(() => browserCohort({ ...partition, stateScope: scope }, rows, policy), /compatibility/);
  }
  assert.throws(() => browserCohort(partition, [rows[0], rows[0]], policy), /compatibility/);
  assert.throws(() => browserCohort(partition, [], policy), /compatibility/);
  const paired = { ...rows[0], selector: { ...rows[0].selector, titles: ["z", "a"], scenario_ids: ["scenario.z", "scenario.a"] } };
  const identity = browserCohort(partition, [paired], policy).name;
  assert.equal(identity, browserCohort(partition, [{ ...paired, selector: { ...paired.selector,
    titles: [...paired.selector.titles].reverse(), scenario_ids: [...paired.selector.scenario_ids].reverse() } }], policy).name);
  assert.notEqual(identity, browserCohort(partition, [{ ...paired, selector: { ...paired.selector,
    scenario_ids: [...paired.selector.scenario_ids].reverse() } }], policy).name, "scenario/title associations remain bound");
});

// Testing Harness NLSpec §12.5: the reset drain exceeds the shutdown deadline
// configured for every started backend by the declared margin. The start
// script reads that deadline from the same authored policy.
test("browser reset drain keeps a declared margin above the configured backend shutdown deadline", () => {
  const topology = JSON.parse(readFileSync(new URL("../../../execution_topology_manifest.json", import.meta.url), "utf8"));
  const policy = topology.browser_reset_policy;
  assert.equal(policy.max_attempts, 1);
  assert.ok(policy.backend_drain_margin_ms > 0);
  assert.ok(policy.backend_drain_timeout_ms > policy.backend_shutdown_drain_seconds * 1000 + policy.backend_drain_margin_ms);
});
