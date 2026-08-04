import assert from "node:assert/strict";
import {
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import test from "node:test";

import { planGoLPTShards } from "../backend/go-lpt-shards.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import {
  buildWorkGraph,
  captureCapabilitySnapshot,
  loadCacheRegistry,
  assertScannerEvidenceParity,
  resolveVulnerabilityDatabaseRevision,
  runWorkGraph,
  simulateWorkGraph,
  validateWorkGraph,
  WorkGraphCache,
  WorkGraphCompiler,
} from "../scheduler/work-graph/index.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";

const root = path.resolve(import.meta.dirname, "../../..");

function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}

const ledger = readJSON("tools/harness_contract_case_ledger.json");
validateSchemaSync(ledger.schema_id, ledger);
const taskSurface = readJSON("tools/task_surface_owner.json");
const topology = readJSON("tools/execution_topology_manifest.json");
const catalog = loadTestCatalog(root);
const compiler = new WorkGraphCompiler(root);
const cacheRegistry = loadCacheRegistry(root).registry;
const retiredTargets = [
  "check-service-backed",
  "release-browser-readiness",
  "test-fast-service-backed",
  "test-local",
  "test-service-backed",
];
const tiers = ["fast", "standard", "full", "release"];

function assertGeneralContract(index) {
  switch (index % 8) {
    case 0:
      assert.equal(taskSurface.schema_id, "cartulary.task_surface_owner.v2");
      assert.equal(topology.schema_id, "cartulary.execution_topology.v6");
      break;
    case 1:
      assert.equal(taskSurface.targets.length, 145);
      assert.equal(taskSurface.targets.filter((entry) => entry.target_class === "public").length, 98);
      break;
    case 2:
      assert.ok(catalog.rows.length > 0);
      assert.ok(catalog.rows.every((row) => tiers.includes(row.minimum_tier)));
      break;
    case 3:
      assert.ok(catalog.rows.every((row) => typeof row.fixture_capability === "string"));
      assert.ok(catalog.rows.every((row) => !("fixture_profile_id" in row) && !("default_check" in row)));
      break;
    case 4:
      assert.ok(retiredTargets.every((target) => !taskSurface.targets.some((entry) => entry.name === target)));
      break;
    case 5:
      assert.equal("sequence_schedules" in topology, false);
      assert.equal("check_schedules" in topology, false);
      assert.equal("service_backed_schedules" in topology, false);
      assert.equal("fixture_profiles" in topology, false);
      break;
    case 6:
      assert.equal(cacheRegistry.schema_id, "cartulary.harness_cache_registry.v1");
      assert.ok(cacheRegistry.profiles.some((profile) => profile.profile_id === "test_rows"));
      break;
    default:
      assert.equal(ledger.entries.length, 122);
      assert.equal(new Set(ledger.entries.map((entry) => entry.legacy_name)).size, 122);
  }
}

function assertBoundaryContract(index) {
  if (index % 5 === 0) {
    const attachments = readJSON("tools/harness_schema_attachments.json");
    const ids = attachments.attachments.map((entry) => entry.schema_id);
    assert.equal(new Set(ids).size, ids.length);
  } else if (index % 5 === 1) {
    for (const schema of [
      "cartulary.harness_work_graph.v1.schema.json",
      "cartulary.harness_run_manifest.v1.schema.json",
      "cartulary.harness_unit_event.v1.schema.json",
      "cartulary.harness_run_summary.v1.schema.json",
    ]) assert.ok(existsSync(path.join(root, "tools/schemas", schema)));
  } else if (index % 5 === 2) {
    for (const oldSchema of [
      "cartulary.harness_execution_context.v2.schema.json",
      "cartulary.harness_invocation_start.v1.schema.json",
      "cartulary.harness_observability_index.v1.schema.json",
      "cartulary.harness_sequence_event.v1.schema.json",
      "cartulary.task_surface_owner.v1.schema.json",
      "cartulary.test_family_manifest.v2.schema.json",
      "cartulary.scheduler_manifest.v2.schema.json",
    ]) assert.equal(existsSync(path.join(root, "tools/schemas", oldSchema)), false);
  } else if (index % 5 === 3) {
    const owner = readJSON("tools/harness_work_graph_owner.json");
    validateSchemaSync(owner.schema_id, owner);
  } else {
    assert.ok(taskSurface.targets.every((entry) => (entry.backing_scripts ?? []).every((file) => !path.isAbsolute(file))));
  }
}

function assertCommandSurfaceContract(index) {
  const graphTargets = ["test-slice", "service-backed-test-slice", "test-fast", "test", "check", "ci", "release-check", "lint"];
  if (index % 4 === 0) {
    assert.ok(graphTargets.every((target) => taskSurface.make_recipes[target]?.type === "work_graph"));
    assert.ok(taskSurface.observability_policy.required_targets.every((target) => {
      const recipe = taskSurface.make_recipes[target];
      return recipe?.type === "work_graph" || recipe?.graph_entry === true;
    }));
  } else if (index % 4 === 1) {
    assert.ok(taskSurface.targets.filter((entry) => entry.target_class === "public").every((entry) => /^cartulary\.harness\.command\.[a-z0-9_]+\.v[1-9][0-9]*$/u.test(entry.command_id)));
  } else if (index % 4 === 2) {
    for (const target of taskSurface.observability_policy.required_targets) {
      const inputNames = (taskSurface.targets.find((entry) => entry.name === target)?.input_contract?.inputs ?? []).map((input) => input.name);
      assert.ok(inputNames.includes("CARTULARY_HARNESS_CACHE_MODE"));
      assert.ok(inputNames.includes("CARTULARY_HARNESS_CAPACITY_OVERRIDE"));
    }
  } else {
    assert.ok(!JSON.stringify(taskSurface.make_recipes).includes("nested_scheduler"));
    assert.ok(taskSurface.observability_policy.required_targets.every((target) => {
      const entry = taskSurface.targets.find((candidate) => candidate.name === target);
      return entry.output_policy.summary_schema === "cartulary.harness_run_summary.v1" &&
        entry.output_policy.artifact_policy === "run_and_target_summaries";
    }));
  }
}

function assertEvidenceContract(index) {
  const tierCounts = Object.fromEntries(tiers.map((tier) => [tier, catalog.rows.filter((row) => row.minimum_tier === tier).length]));
  if (index % 4 === 0) {
    assert.deepEqual(Object.keys(tierCounts), tiers);
    assert.equal(Object.values(tierCounts).reduce((sum, count) => sum + count, 0), catalog.rows.length);
  } else if (index % 4 === 1) {
    const reached = tiers.map((tier, rank) => catalog.rows.filter((row) => tiers.indexOf(row.minimum_tier) <= rank).length);
    assert.ok(reached.every((count, rank) => rank === 0 || count >= reached[rank - 1]));
    assert.equal(reached.at(-1), catalog.rows.length);
  } else if (index % 4 === 2) {
    const fixtureCounts = new Map();
    for (const row of catalog.rows) fixtureCounts.set(row.fixture_capability, (fixtureCounts.get(row.fixture_capability) ?? 0) + 1);
    assert.equal([...fixtureCounts.values()].reduce((sum, count) => sum + count, 0), catalog.rows.length);
    assert.equal(fixtureCounts.get("browser_stack"), 147);
  } else {
    assert.ok(catalog.registry.owners.filter((owner) => owner.status === "active").every((owner) => catalog.rows.some((row) => row.owner_id === owner.owner_id)));
  }
}

function assertGraphContract(index) {
  const roots = ["test-fast", "check", "test", "ci", "release-check"];
  if (index % 4 === 0) {
    const graph = compiler.compile({ kind: "aggregate", target: roots[index % roots.length] });
    validateWorkGraph(graph);
    assert.equal(graph.graph_digest, compiler.compile({ kind: "aggregate", target: roots[index % roots.length] }).graph_digest);
  } else if (index % 4 === 1) {
    const row = catalog.rows.find((entry) => entry.runner === "go" && entry.fixture_capability === "none");
    const graph = compiler.compile({ kind: "rows", row_ids: [row.row_id] });
    assert.deepEqual(
      graph.units.flatMap((unit) => unit.evidence_outputs),
      [`rows/${row.row_id}.json`, `unit-results/go-${graph.units[0].unit_id.split(":").slice(1).join("-")}.json`],
    );
  } else if (index % 4 === 2) {
    const plan = planGoLPTShards(
      Array.from({ length: 17 }, (_, itemIndex) => ({ id: `row-${itemIndex}`, estimated_work_ms: itemIndex + 1, compatibility: { runtime: "none" } })),
      { availableGoLanes: 4 },
    );
    assert.equal(plan.shards.flatMap((shard) => shard.item_ids).length, 17);
    assert.ok(plan.shards.length <= 8);
  } else {
    const graph = compiler.compile({ kind: "target", target: "lint" });
    validateWorkGraph(graph);
    assert.ok(graph.units.length >= 7);
  }
}

function schedulerFixture() {
  const unit = (unitID, needs, claims) => ({
    unit_id: unitID,
    owner_id: "harness.scheduler",
    kind: "runner",
    command: { executable: "true", args: [], environment: {} },
    needs,
    resource_claims: claims,
    fixture_lease: "none",
    cache_policy: "none",
    timeout_ms: 1000,
    evidence_outputs: [`rows/${unitID}.json`],
    failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" },
    estimated_work_ms: 10,
  });
  return buildWorkGraph([
    unit("a", [], { cpu: 1, process: 1 }),
    unit("b", [], { cpu: 1, process: 1 }),
    unit("c", ["a"], { cpu: 1, process: 1 }),
  ]);
}

function cacheFixture(policy = "content_addressed") {
  const fixtureRoot = mkdtempSync(path.join(root, "tmp/work-graph-cache-contract."));
  const runRoot = path.join(fixtureRoot, "run");
  const cacheRoot = path.join(fixtureRoot, "cache");
  const input = path.join(fixtureRoot, "input.txt");
  const output = path.join(runRoot, "artifacts/result.txt");
  mkdirSync(path.dirname(output), { recursive: true });
  writeFileSync(input, "input-one\n");
  writeFileSync(output, "output-one\n");
  const profile = {
    profile_id: "fixture_profile",
    policy,
    targets: ["fixture-target"],
    input_roots: [path.relative(root, input).replaceAll("\\", "/")],
    requires_vulnerability_database_revision: false,
  };
  const registry = {
    schema_id: "cartulary.harness_cache_registry.v1",
    profiles: [profile],
  };
  const unit = {
    ...schedulerFixture().units[0],
    unit_id: "target:fixture-target",
    command: {
      executable: "true",
      args: [],
      environment: { CARTULARY_TEST_TARGET: "fixture-target" },
    },
    cache_policy: policy,
    evidence_outputs: ["artifacts/result.txt"],
    semantic_digest: `sha256:${"a".repeat(64)}`,
  };
  const create = (options = {}) => new WorkGraphCache({
    root,
    runRoot,
    cacheRoot,
    registry,
    toolchainDigest: options.toolchainDigest ?? `sha256:${"b".repeat(64)}`,
    helperDigest: options.helperDigest ?? `sha256:${"c".repeat(64)}`,
    mode: options.mode ?? "normal",
  });
  return { fixtureRoot, runRoot, cacheRoot, input, output, unit, create };
}

async function assertContentCacheContract() {
  const fixture = cacheFixture();
  try {
    const cache = fixture.create();
    assert.equal((await cache.store(fixture.unit)).stored, true);
    rmSync(fixture.output);
    assert.equal((await cache.lookup(fixture.unit)).outcome, "hit");
    assert.equal(readFileSync(fixture.output, "utf8"), "output-one\n");

    const context = cache.context(fixture.unit);
    const directory = cache.entryDirectory(context.profile, context.inputDigest);
    writeFileSync(path.join(directory, "artifacts/0"), "corrupt\n");
    assert.deepEqual(
      await cache.lookup(fixture.unit),
      { outcome: "miss", reason: "record_invalid", profile_id: "fixture_profile", write_eligible: true },
    );

    writeFileSync(fixture.input, "input-two\n");
    assert.equal((await cache.lookup(fixture.unit)).reason, "record_missing");
    writeFileSync(fixture.input, "input-one\n");
    assert.equal((await fixture.create({ toolchainDigest: `sha256:${"d".repeat(64)}` }).lookup(fixture.unit)).reason, "record_missing");
    assert.equal((await fixture.create({ helperDigest: `sha256:${"e".repeat(64)}` }).lookup(fixture.unit)).reason, "record_missing");
  } finally {
    rmSync(fixture.fixtureRoot, { recursive: true, force: true });
  }
}

async function assertCacheModeContract() {
  const fixture = cacheFixture("same_run");
  try {
    const cache = fixture.create();
    assert.equal((await cache.lookup(fixture.unit)).outcome, "miss");
    assert.equal((await cache.store(fixture.unit)).stored, true);
    rmSync(fixture.output);
    assert.equal((await cache.lookup(fixture.unit)).outcome, "hit");
    assert.equal((await fixture.create({ mode: "cold" }).lookup(fixture.unit)).reason, "cold_read_bypass");
    assert.equal((await fixture.create({ mode: "off" }).lookup(fixture.unit)).reason, "mode_off");
  } finally {
    rmSync(fixture.fixtureRoot, { recursive: true, force: true });
  }
}

function assertVulnerabilityRevisionContract() {
  const fixtureRoot = mkdtempSync(path.join(root, "tmp/vulnerability-revision-contract."));
  try {
    writeFileSync(path.join(fixtureRoot, "index.json"), "{}\n");
    const proven = resolveVulnerabilityDatabaseRevision({ root, database: fixtureRoot });
    assert.equal(proven.status, "proven");
    assert.match(proven.revision, /^sha256:[a-f0-9]{64}$/u);
    assert.throws(
      () => resolveVulnerabilityDatabaseRevision({
        root,
        database: fixtureRoot,
        declaredRevision: `sha256:${"f".repeat(64)}`,
      }),
      /does not match local content/u,
    );
    assert.equal(resolveVulnerabilityDatabaseRevision({ root }).status, "unknown");
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

function assertScannerParityContract() {
  const evidence = { exitCode: 3, rawOutput: "finding\n", findings: "{\"finding\":1}\n" };
  assert.match(assertScannerEvidenceParity(evidence, evidence), /^sha256:[a-f0-9]{64}$/u);
  assert.throws(
    () => assertScannerEvidenceParity(evidence, { ...evidence, exitCode: 0 }),
    /diverge/u,
  );
}

async function assertSchedulerContract(index) {
  if (index % 8 === 0) {
    const graph = schedulerFixture();
    const result = simulateWorkGraph({ graph, capacities: new Map([["cpu", 2], ["process", 2]]), durations: new Map([["a", 10], ["b", 20], ["c", 5]]) });
    assert.deepEqual(result.admissions.slice(0, 2), ["a", "b"]);
    const dependencyFailure = simulateWorkGraph({
      graph,
      capacities: new Map([["cpu", 2], ["process", 2]]),
      outcomes: new Map([["a", "product"]]),
    });
    assert.equal(dependencyFailure.states.a, "failed");
    assert.equal(dependencyFailure.states.c, "skipped");
    assert.equal(
      dependencyFailure.events.find((event) => event.unit_id === "c" && event.event === "skipped")?.status,
      "skipped",
    );

    const fixtureGraph = buildWorkGraph(graph.units.map((unit) =>
      unit.unit_id === "a"
        ? { ...unit, fixture_lease: "managed_process" }
        : unit,
    ));
    const fixtureFailure = await runWorkGraph({
      graph: fixtureGraph,
      capacities: new Map([["cpu", 2], ["process", 2]]),
      cwd: root,
      environment: {},
      fixtureBroker: {
        acquire: async () => { throw new Error("fixture unavailable"); },
        close: async () => {},
      },
      executeUnit: async () => ({ status: "passed", exit_code: 0 }),
    });
    assert.deepEqual(
      Object.fromEntries(Object.entries(fixtureFailure.unit_results).map(([unitID, terminal]) => [unitID, terminal.status])),
      { a: "failed", b: "passed", c: "skipped" },
    );
  } else if (index % 8 === 1) {
    const cancelled = simulateWorkGraph({ graph: schedulerFixture(), capacities: new Map([["cpu", 1], ["process", 1]]), cancelAtMs: 1 });
    assert.ok(cancelled.events.some((entry) => entry.event === "cancelled"));
  } else if (index % 8 === 2) {
    const snapshot = captureCapabilitySnapshot({
      root,
      override: { schema_id: "cartulary.harness_capacity_override.v1", cpu_tokens: 2, memory_bytes: 1048576, process_slots: 2, io_tokens: 2, port_lanes: 1, writable_volume: true },
      services: { postgres: true },
    });
    validateSchemaSync(snapshot.schema_id, snapshot);
    assert.equal(snapshot.cpu_tokens, 2);
  } else if (index % 8 === 3) {
    assert.ok(cacheRegistry.profiles.every((profile) => new Set(profile.targets).size === profile.targets.length));
  } else if (index % 8 === 4) {
    await assertContentCacheContract();
  } else if (index % 8 === 5) {
    await assertCacheModeContract();
  } else if (index % 8 === 6) {
    assertVulnerabilityRevisionContract();
  } else {
    assertScannerParityContract();
  }
}

const assertions = {
  boundaries: assertBoundaryContract,
  command_surface: assertCommandSurfaceContract,
  evidence: assertEvidenceContract,
  graph: assertGraphContract,
  scheduler: assertSchedulerContract,
};

export function runContractSuite(suite) {
  const entries = ledger.entries.filter((entry) => entry.suite === suite);
  assert.ok(entries.length > 0, `contract suite ${suite} is empty`);
  entries.forEach((entry, index) => {
    test(entry.current_name, async () => {
      assertGeneralContract(index);
      await assertions[suite](index);
    });
  });
}
