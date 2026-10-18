import assert from "node:assert/strict";
import test from "node:test";
import { WorkGraphCompiler } from "../../scheduler/work-graph/compiler.mjs";
import { planBrowserFunctionalLanes } from "../../scheduler/work-graph/browser-functional-lanes.mjs";
import { resolveCommandInputs, validateSchemaSync } from "../../contract/index.mjs";
import { buildGoInvocations } from "../../execution/runners/go.mjs";
import { automaticCapacities, cpuCapacityWithSafetyMargin, resourceCapacities } from "../../scheduler/work-graph/capability.mjs";
import { loadSchedulerResourceRegistry } from "../../scheduler/scheduler-resources.mjs";

const root = new URL("../../../../", import.meta.url).pathname;
const rowIDs = (graph) => graph.units.flatMap((unit) => unit.row_bindings ?? []).map((row) => row.row_id).sort();

test("host-bounded defaults admit six functional lanes plus shared resident services", () => {
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
    if (!lanes.has(unit.affinity_key)) lanes.set(unit.affinity_key, unit);
  }
  assert.equal(lanes.size, 6);
  const required = {};
  const add = (claims) => {
    for (const [resource, amount] of Object.entries(claims)) required[resource] = (required[resource] ?? 0) + amount;
  };
  const services = new Set();
  for (const unit of lanes.values()) {
    add(unit.resource_claims);
    for (const [kind, claims] of Object.entries(unit.physical_allocations)) {
      if (["postgres", "object_store"].includes(kind)) {
        if (services.has(kind)) continue;
        services.add(kind);
      }
      add(claims);
    }
  }
  assert.deepEqual([required.memory_mb, required.process, required.port_lane, required.postgres, required.object_store], [8960, 52, 20, 13, 7]);
  // Include a cleanup operation as well as all six active groups. Native
  // admission remains responsible for actual occupancy and exclusive waiters.
  add({ memory_mb: 64, process: 1 });
  const available = resourceCapacities({ ...large, writable_volume: true, services: {} });
  for (const [resource, amount] of Object.entries(required)) assert.ok(amount <= available.get(resource), resource);
  assert.ok(required.postgres > small.postgres_lanes);
  assert.ok(required.port_lane > small.port_lanes);
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
  const groups = Array.from({ length: 12 }, (_, index) => ({ name: `group-${index}`, runtimeProfileID: "default", resourceProfileID: "browser_functional" }));
  for (const maxLanes of [4, 6]) {
    const plan = planBrowserFunctionalLanes(groups, { maxLanes, estimateGroup: () => 1 });
    assert.equal(plan.length, maxLanes);
    assert.deepEqual(plan, planBrowserFunctionalLanes([...groups].reverse(), { maxLanes, estimateGroup: () => 1 }));
    assert.equal(new Set(plan.flatMap((lane) => lane.groups.map((item) => item.group.name))).size, 12);
  }
  assert.throws(() => planBrowserFunctionalLanes(groups.slice(0, 2).map((group, index) => ({ ...group, runtimeProfileID: `profile-${index}` })), { maxLanes: 1, estimateGroup: () => 1 }), /runtime profiles/);
  for (const stage of compiler.topology.browser_e2e_batch.stages) for (const group of stage.groups) assert.equal(group.workers, "1");
  for (const limit of [4, 6]) {
    const graph = compiler.compileCommand("browser-e2e", { SUITE: "functional", BROWSER_LANES: String(limit) }).graph;
    const groups = graph.units.filter((unit) => unit.unit_id.startsWith("browser_group:"));
    assert.equal(new Set(groups.map((unit) => unit.affinity_key)).size, limit);
    assert.equal(compiler.browserLaneLimit, limit);
  }
});

test("explicit-only rows stay outside aggregates and broad selections", () => {
  const compiler = new WorkGraphCompiler(root);
  const excluded = compiler.catalog.rows.filter((row) => row.aggregate_placement === "explicit_only");
  assert.equal(excluded.length, 58);
  const legacy = { ...compiler.catalog.rows[0], minimum_tier: "fast" };
  delete legacy.aggregate_placement;
  assert.throws(() => validateSchemaSync("cartulary.test_family_manifest.v10", { schema_id: "cartulary.test_family_manifest.v10", owner_id: legacy.owner_id, rows: [legacy] }));
  for (const target of ["test-fast", "check", "ci", "release-check"]) {
    const selected = rowIDs(compiler.compileAggregate(target));
    assert.ok(excluded.every((row) => !selected.includes(row.row_id)), target);
  }
  for (const owner of new Set(excluded.map((row) => row.owner_id))) {
    assert.ok(rowIDs(compiler.compileCommand("test-slice", { OWNER: owner }).graph).every((id) => !excluded.some((row) => row.row_id === id)));
  }
  assert.throws(() => compiler.compileCommand("browser-e2e", { SUITE: "all" }));
  const visual = rowIDs(compiler.compileCommand("browser-e2e", { SUITE: "visual" }).graph);
  assert.equal(visual.length, 49);
  assert.equal(new Set(visual).size, 49);
  const snapshots = rowIDs(compiler.compileCommand("browser-e2e-visual-update", {}).graph);
  assert.equal(snapshots.length, 44);
  assert.ok(snapshots.every((id) => compiler.catalog.rowByID.get(id).selector.stage === "visual"));
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
