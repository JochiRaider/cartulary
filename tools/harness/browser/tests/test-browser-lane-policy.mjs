import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import { WorkGraphCompiler } from "../../scheduler/work-graph/compiler.mjs";
import { planBrowserLanes } from "../../scheduler/work-graph/browser-lanes.mjs";
import { resolveCommandInputs, validateSchemaSync } from "../../contract/index.mjs";
import { buildGoInvocations } from "../../execution/runners/go.mjs";
import { automaticCapacities, cpuCapacityWithSafetyMargin, resourceCapacities } from "../../scheduler/work-graph/capability.mjs";
import { loadSchedulerResourceRegistry } from "../../scheduler/scheduler-resources.mjs";

const root = new URL("../../../../", import.meta.url).pathname;
const rowIDs = (graph) => graph.units.flatMap((unit) => unit.row_bindings ?? []).map((row) => row.row_id).sort();

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
  const compiler = new WorkGraphCompiler(root);
  const stage = stages.get("support");
  const partition = stage.groups.find((group) => new Set(group.selectedRowIDs.map((id) => compiler.catalog.rowByID.get(id).owner_id)).size > 1);
  assert.ok(partition);
  assert.equal(Object.hasOwn(partition, "ownerID"), false);
  const all = partition.selectedRowIDs.map((id) => compiler.catalog.rowByID.get(id));
  const selected = [all[0], all.find((row) => row.owner_id !== all[0].owner_id)];
  const cohort = browserCohort(partition, selected, stage.invocationPolicy);
  assert.deepEqual(cohort.selectedRowIDs, selected.map((row) => row.row_id).sort());
  assert.equal(cohort.name, browserCohort(partition, [...selected].reverse(), stage.invocationPolicy).name);
  const graph = compiler.compile({ kind: "rows", row_ids: cohort.selectedRowIDs });
  assert.deepEqual(rowIDs(graph), cohort.selectedRowIDs);
  const units = graph.units.filter((unit) => unit.row_bindings?.length);
  assert.equal(units.length, 1);
  assert.equal(units[0].owner_id, "harness.browser");
  const alone = compiler.compile({ kind: "rows", row_ids: [selected[0].row_id] });
  assert.deepEqual(rowIDs(alone), [selected[0].row_id]);
  assert.equal(alone.units.some((unit) => unit.unit_id.startsWith("browser_reset:")), false);
  for (const group of stages.values()) for (const partition of group.groups) {
    const rows = partition.selectedRowIDs.map((id) => compiler.catalog.rowByID.get(id));
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

test("host budgets charge each exclusive browser cluster and one shared object store", () => {
  const registry = loadSchedulerResourceRegistry();
  const capacities = (cores) => automaticCapacities({ registry,
    cpuTokens: cpuCapacityWithSafetyMargin(cores, registry.capacity_policies.cpu_tokens),
    detectedMemory: 32 * 1024 ** 3, detectedProcessLimit: 4096 });
  const large = capacities(24);
  assert.deepEqual([large.postgres_lanes, large.object_store_lanes, large.port_lanes], [16, 8, 32]);
  const small = capacities(4);
  assert.deepEqual([small.postgres_lanes, small.object_store_lanes, small.port_lanes], [3, 3, 6]);
  const smallest = capacities(1);
  assert.deepEqual([smallest.postgres_lanes, smallest.object_store_lanes, smallest.port_lanes], [1, 1, 2]);

  const compiler = new WorkGraphCompiler(root);
  const graph = compiler.compileCommand("browser-e2e", { SUITE: "functional", BROWSER_LANES: "6" }).graph;
  const lanes = new Map();
  for (const unit of graph.units.filter((item) => item.unit_id.startsWith("browser_group:"))) {
    if (!lanes.has(unit.command.environment.CARTULARY_BROWSER_LANE_ID)) lanes.set(unit.command.environment.CARTULARY_BROWSER_LANE_ID, unit);
  }
  assert.equal(lanes.size, 6);
  const requiredFor = (count) => {
    const required = {}, services = new Set();
    const add = (claims) => {
      for (const [resource, amount] of Object.entries(claims)) {
        required[resource] = resource === "volume"
          ? Math.max(required[resource] ?? 0, amount)
          : (required[resource] ?? 0) + amount;
      }
    };
    for (const unit of [...lanes.values()].slice(0, count)) {
      add(unit.resource_claims);
      for (const [kind, claims] of Object.entries(unit.physical_allocations)) {
        if (kind === "object_store") {
          if (services.has(kind)) continue;
          services.add(kind);
        }
        add(claims);
      }
    }
    return required;
  };
  const four = requiredFor(4), six = requiredFor(6);
  assert.deepEqual([four.memory_mb, four.process, four.port_lane, four.postgres, four.object_store], [7424, 74, 17, 12, 5]);
  assert.deepEqual([six.memory_mb, six.process, six.port_lane, six.postgres, six.object_store], [11008, 110, 25, 18, 7]);
  const available = resourceCapacities({ ...large, writable_volume: true, services: {} });
  // Four active allocations plus a cleanup operation fit. Six authored lanes
  // are logical work queues; admission cannot run all six exclusive clusters.
  for (const [resource, amount] of Object.entries({ ...four, memory_mb: four.memory_mb + 64, process: four.process + 1 })) {
    assert.ok(amount <= available.get(resource), resource);
  }
  assert.ok(six.process > available.get("process"));
  assert.ok(six.postgres > available.get("postgres"));
  assert.ok(four.postgres > small.postgres_lanes);
  assert.ok(four.port_lane > small.port_lanes);
});

test("persistent lifecycle borrowing does not reserve the owner's physical capacity twice", () => {
  const compiler = new WorkGraphCompiler(root);
  const rowID = "harness.browser.integration.native_borrowed_fixture_lifecycle";
  const graph = compiler.compileCommand("test-slice", { OWNER: "harness.browser", ROWS: rowID }).graph;
  const unit = graph.units.find((candidate) => candidate.row_bindings?.some((row) => row.row_id === rowID));
  assert.ok(unit);
  assert.deepEqual(unit.service_dependencies, []);
  assert.ok(unit.resource_claims.cpu > 0);
  assert.ok(unit.resource_claims.memory_mb > 0);
  for (const resource of ["browser_stack", "postgres", "object_store", "port_lane"]) {
    assert.equal(unit.resource_claims[resource] ?? 0, 0, resource);
  }
});

test("topology resolves four or six lanes and rejects invalid public controls", () => {
  const compiler = new WorkGraphCompiler(root);
  for (const target of ["test-fast", "check", "ci", "release-check", "test-slice", "browser-e2e", "browser-e2e-visual-update"]) {
    const scope = target === "test-slice" ? { OWNER: "harness.browser" } : target === "browser-e2e" ? { SUITE: "functional" } : {};
    assert.equal(resolveCommandInputs(target, scope, compiler.taskSurface).BROWSER_LANES, "4");
    assert.equal(resolveCommandInputs(target, { ...scope, BROWSER_LANES: "6" }, compiler.taskSurface).BROWSER_LANES, "6");
    for (const value of ["0", "7", "1.5", "six", ""]) assert.throws(() => resolveCommandInputs(target, { ...scope, BROWSER_LANES: value }, compiler.taskSurface));
    assert.throws(() => resolveCommandInputs(target, { ...scope, PLAYWRIGHT_WORKERS: "1" }, compiler.taskSurface));
  }
  const groups = Array.from({ length: 12 }, (_, index) => ({ name: `group-${index}`, runtimeProfileID: "default", resourceProfileID: "browser_ordinary" }));
  for (const maxLanes of [4, 6]) {
    const plan = planBrowserLanes(groups, { maxLanes, estimateGroup: () => 1 });
    assert.equal(plan.length, maxLanes);
    assert.deepEqual(plan, planBrowserLanes([...groups].reverse(), { maxLanes, estimateGroup: () => 1 }));
    assert.equal(new Set(plan.flatMap((lane) => lane.groups.map((item) => item.group.name))).size, 12);
  }
  assert.equal(planBrowserLanes(groups.slice(0, 2).map((group, index) => ({ ...group, runtimeProfileID: `profile-${index}` })), { maxLanes: 1, estimateGroup: () => 1 })[0].groups.length, 2);
  for (const stage of compiler.topology.browser_e2e_batch.stages) for (const group of stage.groups) assert.equal(group.workers, "1");
  for (const limit of [4, 6]) {
    const graph = compiler.compileCommand("browser-e2e", { SUITE: "functional", BROWSER_LANES: String(limit) }).graph;
    const groups = graph.units.filter((unit) => unit.unit_id.startsWith("browser_group:"));
    assert.equal(new Set(groups.map((unit) => unit.command.environment.CARTULARY_BROWSER_LANE_ID)).size, limit);
    assert.equal(compiler.browserLaneLimit, limit);
  }
});

test("explicit-only rows stay outside aggregates and broad selections", () => {
  const compiler = new WorkGraphCompiler(root);
  const excluded = compiler.catalog.rows.filter((row) => row.aggregate_placement === "explicit_only");
  const characterization = "harness.browser.integration.reset_checkpoint_characterization";
  assert.ok(excluded.some((row) => row.row_id === characterization));
  assert.deepEqual(rowIDs(compiler.compileCommand("test-slice", { OWNER: "harness.browser", ROWS: characterization }).graph), [characterization]);
  const legacy = { ...compiler.catalog.rows[0], minimum_tier: "fast" };
  delete legacy.aggregate_placement;
  assert.throws(() => validateSchemaSync("cartulary.test_family_manifest.v13", { schema_id: "cartulary.test_family_manifest.v13", owner_id: legacy.owner_id, rows: [legacy] }));
  for (const target of ["test-fast", "check", "ci", "release-check"]) {
    const selected = rowIDs(compiler.compileAggregate(target));
    assert.ok(excluded.every((row) => !selected.includes(row.row_id)), target);
  }
  for (const owner of new Set(excluded.map((row) => row.owner_id))) {
    assert.ok(rowIDs(compiler.compileCommand("test-slice", { OWNER: owner }).graph).every((id) => !excluded.some((row) => row.row_id === id)));
  }
  assert.throws(() => compiler.compileCommand("browser-e2e", { SUITE: "all" }));
  const visual = rowIDs(compiler.compileCommand("browser-e2e", { SUITE: "visual" }).graph);
  const registry = "web.design.static.visual_fixture_registry";
  assert.equal(visual.length, 48);
  assert.equal(new Set(visual).size, 48);
  assert.equal(visual.filter((id) => id === registry).length, 1);
  const snapshots = rowIDs(compiler.compileCommand("browser-e2e-visual-update", {}).graph);
  assert.equal(snapshots.length, 43);
  assert.ok(snapshots.every((id) => id === registry || compiler.catalog.rowByID.get(id).selector.stage === "visual"));
  assert.equal(snapshots.filter((id) => id === registry).length, 1);
  const review = compiler.compileCommand("harness-contract", { SUITE: "ui-review" }).graph;
  assert.equal(rowIDs(review).filter((id) => excluded.some((row) => row.row_id === id)).length, 9);
  const native = review.units.find((unit) => unit.row_bindings?.some((row) => row.row_id === "harness.browser.unit.native_review"));
  assert.equal(native.command.environment.CARTULARY_GO_RACE, "1");
  const row = compiler.catalog.rowByID.get("harness.browser.unit.native_review");
  const invocation = buildGoInvocations([row], 1, "go", { environment: { CGO_ENABLED: "0" }, goRace: native.command.environment.CARTULARY_GO_RACE === "1" })[0];
  assert.ok(invocation.args.includes("-race"));
  assert.equal(invocation.environment.CGO_ENABLED, "1");
  assert.ok(rowIDs(compiler.compileCommand("test-slice", { OWNER: row.owner_id, ROWS: row.row_id }).graph).includes(row.row_id));
  compiler.owner.policy_units["harness.frontend_install"].required_rows = [{ row_id: row.row_id, rationale: "invalid aggregate dependency" }];
  assert.throws(() => compiler.compileAggregate("ci"), /explicit-only/);
});

test("browser cohorts bind exact cases and execution compatibility independently of reporting ownership", async () => {
  const { browserCohort, packBrowserCohorts } = await import("../browser-cohorts.mjs");
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
  const catalog = { rowByID: new Map(rows.map((row) => [row.row_id, row])) };
  const policy = { startup_ms: 20, cleanup_ms: 10 }, owner = {
    evidence_estimates_ms: { browser: 40 },
    lifecycle_estimates_ms: { browser_allocation_acquire: 15, browser_allocation_release: 5, browser_reset: 10 },
  };
  const packed = packBrowserCohorts([partition], catalog, policy, owner, () => 130);
  assert.deepEqual(packed.map((cohort) => cohort.estimatedWorkMs), [110, 70]);
  assert.deepEqual(packed.map((cohort) => cohort.selectedRowIDs.length), [2, 1]);
  assert.deepEqual(packBrowserCohorts([{ ...partition, selectedRowIDs: [...partition.selectedRowIDs].reverse() }], catalog, policy, owner, () => 130), packed);
  assert.deepEqual(packBrowserCohorts([partition], catalog, policy, owner, () => 110)
    .map((cohort) => cohort.selectedRowIDs.length), [1, 1, 1], "allocation headroom cannot enlarge the existing envelope");
  assert.deepEqual(packBrowserCohorts([], catalog, policy, owner, () => 130), []);
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
  assert.notEqual(subset.name, packed[0].name);
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
