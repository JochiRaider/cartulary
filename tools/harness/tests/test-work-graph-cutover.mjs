#!/usr/bin/env node

import assert from "node:assert/strict";
import { existsSync, mkdtempSync, readFileSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { FixtureBroker } from "../scheduler/fixture-broker/index.mjs";
import {
  WorkGraphCompiler,
  assertFixtureServiceDependencies,
  assertServiceDependencies,
  buildWorkGraph,
  createAtomicNDJSONWriter,
  simulateWorkGraph,
  workGraphCacheRootRelative,
  writeAtomicNDJSON,
} from "../scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const mode = process.argv[2] ?? "matrix";
const validModes = new Set([
  "fast",
  "matrix",
  "scheduler-smoke",
  "scheduler-matrix",
  "fixture-smoke",
  "fixture-matrix",
  "service-backed",
]);
assert.ok(validModes.has(mode), `unknown work-graph cutover test mode ${mode}`);

function unit(unitID, claims, work, needs = []) {
  return {
    unit_id: unitID,
    owner_id: "harness.command_surface",
    kind: "runner",
    command: { executable: "true", args: [], environment: {} },
    needs,
    resource_claims: claims,
    fixture_lease: "none",
    service_dependencies: [],
    cache_policy: "none",
    timeout_ms: 1000,
    current_run_evidence_outputs: [`rows/${unitID}.json`],
    failure_policy: {
      block_descendants: true,
      continue_independent: true,
      aggregate_effect: "required",
    },
    estimated_work_ms: work,
  };
}

function assertGraphCutover() {
  assert.equal(workGraphCacheRootRelative, ".cache/cartulary/graph-v2");
  const compiler = new WorkGraphCompiler(root);
  const topology = JSON.parse(
    readFileSync(path.join(root, "tools/execution_topology_manifest.json"), "utf8"),
  );
  assert.throws(
    () => assertServiceDependencies(topology, "default", ["postgres", "postgres"], "duplicate"),
    /duplicate/u,
  );
  assert.throws(
    () => assertServiceDependencies(topology, "default", ["postgres", "object_store"], "unsorted"),
    /sorted/u,
  );
  assert.throws(
    () => assertServiceDependencies(topology, "default", ["message_bus"], "unknown"),
    /unknown/u,
  );
  assert.throws(
    () => assertServiceDependencies(topology, "none", ["postgres"], "runtime-incompatible"),
    /unavailable/u,
  );
  assert.throws(
    () => assertFixtureServiceDependencies("postgres_dedicated", [], "omitted"),
    /requires service dependency postgres/u,
  );
  assert.throws(
    () => buildWorkGraph([{ ...unit("omitted", { cpu: 1 }, 1), service_dependencies: undefined }]),
    /service_dependencies is required/u,
  );
  assert.throws(
    () => buildWorkGraph([unit("unbounded", {}, 1)]),
    /resource_claims must bound executable work/u,
  );
  const affectedRowID = "module.entities.integration.the_mention_resolve_route_persists_durable_resol_d68c1befd6";
  const affectedDirect = compiler.compile({ kind: "rows", row_ids: [affectedRowID] });
  const affectedOwner = compiler.compile({
    kind: "owner",
    owner_id: "module.entities",
    row_ids: [affectedRowID],
  });
  assert.deepEqual(affectedDirect, affectedOwner, "equivalent row selectors must compile identically");
  const affectedUnit = affectedDirect.units.find((entry) =>
    entry.current_run_evidence_outputs.includes(`rows/${affectedRowID}.json`),
  );
  assert.equal(affectedUnit.fixture_lease, "postgres_dedicated");
  assert.deepEqual(affectedUnit.service_dependencies, ["object_store", "postgres"]);
  assert.equal(affectedUnit.resource_claims.object_store, 1);
  assert.equal(affectedUnit.resource_claims.cpu, 1);
  assert.equal(affectedUnit.resource_claims.postgres, 1);
  assert.equal(affectedUnit.command.environment.CARTULARY_UNIT_CPU_TOKENS, "1");
  assert.equal(affectedUnit.command.environment.GOMAXPROCS, "1");

  const dedicatedRows = [
    "module.entities.integration.host_and_identity_create_routes_reuse_exact_matc_f443e2591f",
    "module.entities.integration.the_explicit_merge_route_repoints_live_fan_out_p_0ec76e8044",
  ];
  const dedicatedGraph = compiler.compile({ kind: "rows", row_ids: dedicatedRows });
  const dedicatedUnits = dedicatedGraph.units.filter((entry) =>
    dedicatedRows.some((rowID) =>
      entry.current_run_evidence_outputs.includes(`rows/${rowID}.json`),
    ),
  );
  assert.equal(dedicatedUnits.length, 1, "compatible dedicated rows must share one Go process");
  assert.equal(dedicatedUnits[0].fixture_lease, "postgres_dedicated");
  assert.deepEqual(
    dedicatedUnits[0].current_run_evidence_outputs,
    dedicatedRows.map((rowID) => `rows/${rowID}.json`),
  );
  assert.equal(dedicatedUnits[0].command.environment.GOMAXPROCS, "1");

  const migrationRows = [
    "module.database_migrations.integration.production_ddl_v2_recurrence",
    "module.database_migrations.integration.production_preflight_state_matrix",
  ];
  const migrationGraph = compiler.compile({ kind: "rows", row_ids: migrationRows });
  const migrationUnits = migrationGraph.units.filter((entry) =>
    migrationRows.some((rowID) =>
      entry.current_run_evidence_outputs.includes(`rows/${rowID}.json`),
    ),
  );
  assert.equal(migrationUnits.length, 1, "compatible migration rows must share one Go process");
  assert.equal(migrationUnits[0].fixture_lease, "postgres_migration");
  assert.deepEqual(
    migrationUnits[0].current_run_evidence_outputs,
    migrationRows.map((rowID) => `rows/${rowID}.json`),
  );

  const processGlobalRows = [
    "platform.jobs.integration.claim_recovery_and_publication",
    "platform.jobs.integration.operational_telemetry",
  ];
  const processGlobalGraph = compiler.compile({ kind: "rows", row_ids: processGlobalRows });
  const processGlobalUnits = processGlobalGraph.units.filter((entry) =>
    processGlobalRows.some((rowID) =>
      entry.current_run_evidence_outputs.includes(`rows/${rowID}.json`),
    ),
  );
  assert.equal(
    processGlobalUnits.length,
    2,
    "an explicit process-global assertion must not share a Go process",
  );

  const managedProcessRows = [
    "module.extensions.integration.bc011_deadline_precedence_ef23af86ac",
    "module.extensions.integration.bc015_browser_availability_e0a71bee5d",
  ];
  const managedProcessGraph = compiler.compile({
    kind: "rows",
    row_ids: managedProcessRows,
  });
  const managedProcessUnits = managedProcessGraph.units.filter((entry) =>
    managedProcessRows.some((rowID) =>
      entry.current_run_evidence_outputs.includes(`rows/${rowID}.json`),
    ),
  );
  assert.equal(managedProcessUnits.length, 2, "managed-process rows must remain process-isolated");
  assert.ok(managedProcessUnits.every((entry) => entry.fixture_lease === "managed_process"));

  const rawPostgresUnit = compiler
    .compile({ kind: "target", target: "backend-integration" })
    .units.find((entry) => entry.unit_id === "raw_go:backend-integration-testutil");
  assert.equal(rawPostgresUnit.resource_claims.cpu, 1);
  assert.equal(rawPostgresUnit.resource_claims.postgres, 1);
  assert.equal(rawPostgresUnit.command.environment.GOMAXPROCS, "1");

  const ioHeavy = compiler.compile({
    kind: "rows",
    row_ids: ["platform.audit.integration.transactional_immutable_persistence"],
  }).units.find((entry) =>
    entry.current_run_evidence_outputs.includes(
      "rows/platform.audit.integration.transactional_immutable_persistence.json",
    ),
  );
  assert.equal(ioHeavy.resource_claims.io, 2, "io_heavy must be topology-authoritative");

  const first = compiler.compileAggregatePlan("test-fast");
  const second = compiler.compileAggregatePlan("test-fast");
  assert.deepEqual(first, second, "same-source graph compilation must be byte-stable");
  assert.ok(first.graph.units.length > 0);
  assert.ok(first.projections["test-fast"].length > 0);
  for (const graphTarget of ["test-fast", "test", "check", "ci", "release-check"]) {
    const graph = compiler.compileAggregatePlan(graphTarget).graph;
    assert.equal(
      graph.units.some((entry) =>
        ["test-fast", "test", "check", "ci", "release-check"].includes(
          entry.command.args.at(-1),
        )),
      false,
      `${graphTarget} must not invoke a nested aggregate`,
    );
  }
  const releaseGraph = compiler.compileAggregatePlan("release-check").graph;
  for (const entry of releaseGraph.units) {
    if (Object.keys(entry.resource_claims).length === 0) continue;
    assert.equal(
      entry.shared_locks.includes("host_activity") +
        entry.exclusive_locks.includes("host_activity"),
      1,
      `${entry.unit_id} must have exactly one host_activity lock mode`,
    );
  }
  const compatibility = releaseGraph.units.find(
    (entry) => entry.unit_id === "target:seaweedfs-compatibility",
  );
  assert.equal(compatibility?.fixture_lease, "object_store_namespace");
  assert.equal(compatibility?.resource_claims.object_store, 1);
  assert.equal(
    releaseGraph.units.some((entry) => entry.unit_id === "target:object-store-init"),
    false,
    "release compatibility must not start the local-development object store",
  );
}

function assertScheduler() {
  const graph = buildWorkGraph([
    unit("critical", { cpu: 2 }, 20),
    unit("backfill", { cpu: 1 }, 5),
    unit("tail", { cpu: 1 }, 4, ["critical"]),
  ]);
  const result = simulateWorkGraph({
    graph,
    capacities: new Map([["cpu", 2]]),
  });
  assert.equal(result.status, "passed");
  assert.deepEqual(result.states, {
    backfill: "passed",
    critical: "passed",
    tail: "passed",
  });
  assert.equal(result.events.at(-1).event, "completed");

  const blocker = unit("shared-blocker", { cpu: 1 }, 10);
  blocker.shared_locks = ["host_activity"];
  const gate = unit("exclusive-gate", { cpu: 1 }, 1);
  const exclusive = unit("quiet-measurement", { cpu: 1 }, 1, ["exclusive-gate"]);
  exclusive.exclusive_locks = ["host_activity"];
  const laterShared = unit("later-shared", { cpu: 1 }, 1, ["exclusive-gate"]);
  laterShared.shared_locks = ["host_activity"];
  const fairness = simulateWorkGraph({
    graph: buildWorkGraph([blocker, gate, exclusive, laterShared]),
    capacities: new Map([["cpu", 2]]),
    durations: new Map([
      ["shared-blocker", 10],
      ["exclusive-gate", 1],
      ["quiet-measurement", 1],
      ["later-shared", 1],
    ]),
  });
  assert.ok(
    fairness.admissions.indexOf("quiet-measurement") <
      fairness.admissions.indexOf("later-shared"),
    "a ready exclusive measurement waiter must prevent later shared backfill",
  );

  const failedGroup = unit("failed-browser-group", { cpu: 1 }, 1);
  failedGroup.failure_policy.block_descendants = false;
  const evidenceFinalizer = unit(
    "browser-target-finalizer",
    { cpu: 1 },
    1,
    [failedGroup.unit_id],
  );
  const reporting = simulateWorkGraph({
    graph: buildWorkGraph([failedGroup, evidenceFinalizer]),
    capacities: new Map([["cpu", 1]]),
    outcomes: new Map([[failedGroup.unit_id, "test_failure"]]),
  });
  assert.equal(reporting.status, "failed", "a reporting descendant must not hide its dependency failure");
  assert.deepEqual(reporting.states, {
    "browser-target-finalizer": "passed",
    "failed-browser-group": "failed",
  });
}

async function assertAtomicNDJSON() {
  const directory = mkdtempSync(path.join(os.tmpdir(), "cartulary-work-graph-ndjson-"));
  try {
    const output = path.join(directory, "unit-events.ndjson");
    const events = Array.from({ length: 10_000 }, (_, index) => ({
      event: "test_event",
      index,
      detail: `event-${index}`,
    }));
    writeAtomicNDJSON(output, events);
    const lines = readFileSync(output, "utf8").trimEnd().split("\n");
    assert.equal(lines.length, events.length);
    assert.deepEqual(JSON.parse(lines.at(0)), events.at(0));
    assert.deepEqual(JSON.parse(lines.at(-1)), events.at(-1));

    const streamed = path.join(directory, "streamed-events.ndjson");
    const writer = createAtomicNDJSONWriter(streamed);
    for (const event of events.slice(0, 1000)) await writer.write(event);
    assert.equal(existsSync(streamed), false, "a live writer must not publish the canonical path");
    assert.equal(existsSync(writer.stagingFile), true, "a live writer must expose its private staging stream");
    assert.equal(readFileSync(writer.stagingFile, "utf8").trimEnd().split("\n").length, 1000);
    await writer.close();
    assert.equal(existsSync(writer.stagingFile), false, "publishing must consume the staging path");
    assert.equal(readFileSync(streamed, "utf8").trimEnd().split("\n").length, 1000);

    const aborted = path.join(directory, "aborted-events.ndjson");
    const abortedWriter = createAtomicNDJSONWriter(aborted);
    await abortedWriter.write(events[0]);
    await abortedWriter.abort();
    assert.equal(existsSync(aborted), false, "an aborted writer must not publish partial evidence");
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}

async function assertFixtures() {
  const released = [];
  let browserAllocation = 0;
  const providers = {
    postgres_transaction: {
      acquire: async ({ unitID }) => ({
        ownership: "borrowed",
        resource_ids: [`postgres:${unitID}`],
        resource: { unitID },
        detach: async () => released.push(unitID),
      }),
    },
    postgres_dedicated: {
      acquire: async ({ unitID }) => ({
        ownership: "owned",
        resource_ids: [`postgres:${unitID}`],
        resource: { unitID },
        release: async () => released.push(`healthy:${unitID}`),
        quarantine: async () => released.push(`quarantined:${unitID}`),
      }),
    },
    browser_stack: {
      acquire: async ({ affinityKey, fixtureProfileID, snapshotKey, builderUnitID }) => {
        const allocation = ++browserAllocation;
        return {
          ownership: "owned",
          resource_ids: [`browser:${affinityKey}:allocation-${allocation}`],
          resource: { affinityKey, allocation },
          ...(fixtureProfileID
            ? {
                fixture_profile_id: fixtureProfileID,
                snapshot_key: snapshotKey,
                builder_unit_id: builderUnitID,
                clone_ordinal: 1,
              }
            : {}),
          release: async () => released.push(affinityKey),
          quarantine: async () => released.push(`quarantined:${affinityKey}`),
        };
      },
    },
  };
  const broker = new FixtureBroker({ providers });
  const transaction = await broker.acquire("postgres_transaction", { unitID: "row-a" });
  assert.equal(transaction.record.ownership, "borrowed");
  await transaction.release();
  const failedDedicated = await broker.acquire("postgres_dedicated", { unitID: "row-failed" });
  await failedDedicated.quarantine();
  assert.equal(failedDedicated.record.state, "quarantined");
  const first = await broker.acquire("browser_stack", { unitID: "group-a", affinityKey: "chain-a" });
  const second = await broker.acquire("browser_stack", { unitID: "group-b", affinityKey: "chain-a" });
  assert.equal(first.resource, second.resource, "browser affinity must reuse one stack lease");
  await first.release();
  await second.release();
  const warm = await broker.acquire("browser_stack", {
    unitID: "warm-a",
    affinityKey: "warm-lane",
  });
  assert.deepEqual(await warm.release({ healthy: true, retainWarm: true }), {
    retained: true,
  });
  const warmSuccessor = await broker.acquire("browser_stack", {
    unitID: "warm-b",
    affinityKey: "warm-lane",
  });
  assert.equal(warm.resource, warmSuccessor.resource, "healthy lane must retain its allocation");
  await warmSuccessor.release({ healthy: false, retainWarm: true });
  const replacement = await broker.acquire("browser_stack", {
    unitID: "warm-c",
    affinityKey: "warm-lane",
  });
  assert.notEqual(
    replacement.resource,
    warm.resource,
    "failed lane must be quarantined before a successor obtains a fresh allocation",
  );
  assert.notDeepEqual(
    replacement.record.resource_ids,
    warm.record.resource_ids,
    "replacement browser allocations must expose a fresh concrete resource identity",
  );
  await replacement.release();
  const profile = {
    fixtureProfileID: "ac043_large_grid_snapshot_v1",
    snapshotKey: "a".repeat(64),
    builderUnitID: `fixture_snapshot:default:ac043_large_grid_snapshot_v1:${"a".repeat(64)}`,
    rowID: "module.timeline.measurement.row",
    predicateID: "perf.typing_ack.v1",
  };
  const profiled = await broker.acquire("browser_stack", {
    unitID: "profiled-a",
    affinityKey: "profiled-a",
    ...profile,
  });
  assert.equal(profiled.record.fixture_profile_id, profile.fixtureProfileID);
  assert.equal(profiled.record.snapshot_key, profile.snapshotKey);
  assert.equal(profiled.record.builder_unit_id, profile.builderUnitID);
  assert.equal(profiled.record.clone_ordinal, 1);
  const sameProfile = await broker.acquire("browser_stack", {
    unitID: "profiled-b",
    affinityKey: "profiled-a",
    ...profile,
  });
  assert.equal(profiled.resource, sameProfile.resource, "same key and affinity must join one allocation");
  const otherKey = "b".repeat(64);
  const differentProfile = await broker.acquire("browser_stack", {
    unitID: "profiled-c",
    affinityKey: "profiled-a",
    ...profile,
    snapshotKey: otherKey,
    builderUnitID: `fixture_snapshot:default:ac043_large_grid_snapshot_v1:${otherKey}`,
  });
  assert.notEqual(profiled.resource, differentProfile.resource, "different keys must never share a browser allocation");
  await profiled.release();
  await sameProfile.release();
  await differentProfile.release();
  await broker.close();
  assert.deepEqual(released.sort(), [
    "chain-a",
    "profiled-a",
    "profiled-a",
    "quarantined:row-failed",
    "quarantined:warm-lane",
    "row-a",
    "warm-lane",
  ]);

}

if (["fast", "matrix"].includes(mode)) {
  assertGraphCutover();
  await assertAtomicNDJSON();
}
if (["scheduler-smoke", "scheduler-matrix", "matrix"].includes(mode)) assertScheduler();
if (["fixture-smoke", "fixture-matrix", "service-backed", "matrix"].includes(mode)) {
  await assertFixtures();
}
