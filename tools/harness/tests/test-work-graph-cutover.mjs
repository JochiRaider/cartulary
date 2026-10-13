#!/usr/bin/env node

import assert from "node:assert/strict";
import { existsSync, mkdtempSync, readFileSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import {
  WorkGraphCompiler,
  assertFixtureServiceDependencies,
  assertServiceDependencies,
  buildWorkGraph,
} from "../scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const mode = process.argv[2] ?? "matrix";
assert.ok(["fast", "matrix"].includes(mode), `unknown graph compiler test mode ${mode}`);

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

function expectedCurrentRunOutputs(unitValue, rowIDs) {
  const safeUnitID = unitValue.unit_id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
  return [...rowIDs.map((rowID) => `rows/${rowID}.json`), `unit-results/${safeUnitID}.json`].sort();
}

function assertCanonicalCurrentRunOutputs(unitValue, rowIDs, label) {
  assert.deepEqual(
    unitValue.current_run_evidence_outputs,
    expectedCurrentRunOutputs(unitValue, rowIDs),
    `${label} must publish only selected row evidence and one canonical unit result`,
  );
  assert.equal(
    unitValue.current_run_evidence_outputs.filter((output) => output.startsWith("unit-results/"))
      .length,
    1,
    `${label} must publish exactly one canonical unit result`,
  );
  assert.equal(
    new Set(unitValue.current_run_evidence_outputs).size,
    unitValue.current_run_evidence_outputs.length,
    `${label} outputs must be unique`,
  );
}

function assertGraphCutover() {
  const compiler = new WorkGraphCompiler(root);
  const frontendRows = compiler.catalog.rows.filter((row) => row.runner === "vitest");
  const firstFrontend = frontendRows.find((row) => frontendRows.filter((other) => other.selector.file === row.selector.file &&
    other.owner_id === row.owner_id && other.evidence_class === row.evidence_class && other.resource_profile_id === row.resource_profile_id).length > 1);
  const pair = frontendRows.filter((row) => row.selector.file === firstFrontend.selector.file && row.owner_id === firstFrontend.owner_id &&
    row.evidence_class === firstFrontend.evidence_class && row.resource_profile_id === firstFrontend.resource_profile_id).slice(0, 2);
  const ids = pair.map((row) => row.row_id).sort();
  const consumers = (graph) => graph.units.filter((unit) => unit.current_run_evidence_outputs.some((output) => ids.some((id) => output === `rows/${id}.json`)));
  const sharedGraph = compiler.compileRows(ids), shared = consumers(sharedGraph);
  assert.equal(shared.length, 1, "compatible file rows must share one fresh process");
  assert.equal(shared[0].command.environment.VITEST_MAX_WORKERS, "1");
  assert.equal(shared[0].resource_claims.cpu, 1);
  assertCanonicalCurrentRunOutputs(shared[0], ids, "Vitest file group");
  assert.deepEqual(compiler.compileRows([...ids].reverse()), sharedGraph);
  assert.deepEqual(compiler.compileOwner(pair[0].owner_id, ids), sharedGraph);
  const single = consumers(compiler.compileRows([ids[0]]));
  assertCanonicalCurrentRunOutputs(single[0], [ids[0]], "Vitest subset");
  assert.notEqual(single[0].unit_id, shared[0].unit_id, "selected membership owns semantic identity");
  for (const mutation of [
    { process_isolation: "exclusive" }, { evidence_class: "security" },
    { runtime_profile_id: "default" }, { resource_profile_id: "harness_contract" },
    { owner_id: pair[1].owner_id === "web.architecture" ? "harness.browser" : "web.architecture" },
    { selector: { ...pair[1].selector, project_id: pair[1].selector.project_id === "browser-unit" ? "harness-node" : "browser-unit" } },
  ]) {
    const prior = compiler.catalog.rowByID.get(pair[1].row_id);
    compiler.catalog.rowByID.set(prior.row_id, { ...prior, ...mutation });
    try { assert.equal(consumers(compiler.compileRows(ids)).length, 2, JSON.stringify(mutation)); }
    finally { compiler.catalog.rowByID.set(prior.row_id, prior); }
  }
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
  assertCanonicalCurrentRunOutputs(dedicatedUnits[0], dedicatedRows, "dedicated Go batch");
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
  assertCanonicalCurrentRunOutputs(migrationUnits[0], migrationRows, "migration Go batch");

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

  const catalogPostgresUnit = compiler
    .compile({ kind: "target", target: "backend-integration" })
    .units.find((entry) => entry.current_run_evidence_outputs.includes("rows/harness.browser.integration.wstest_helper_contract.json"));
  assert.equal(catalogPostgresUnit.resource_claims.cpu, 1);
  assert.equal(catalogPostgresUnit.resource_claims.postgres, 1);
  assert.equal(catalogPostgresUnit.command.environment.GOMAXPROCS, "1");

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

assertGraphCutover();
