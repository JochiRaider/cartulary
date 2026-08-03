#!/usr/bin/env node

import assert from "node:assert/strict";
import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { FixtureBroker } from "../scheduler/fixture-broker/index.mjs";
import {
  WorkGraphCompiler,
  buildWorkGraph,
  simulateWorkGraph,
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
    cache_policy: "none",
    timeout_ms: 1000,
    evidence_outputs: [`rows/${unitID}.json`],
    failure_policy: {
      block_descendants: true,
      continue_independent: true,
      aggregate_effect: "required",
    },
    estimated_work_ms: work,
  };
}

function assertGraphCutover() {
  const compiler = new WorkGraphCompiler(root);
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
}

function assertAtomicNDJSON() {
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
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}

async function assertFixtures() {
  const released = [];
  const providers = {
    postgres_transaction: {
      acquire: async ({ unitID }) => ({
        ownership: "borrowed",
        resource_ids: [`postgres:${unitID}`],
        resource: { unitID },
        detach: async () => released.push(unitID),
      }),
    },
    browser_stack: {
      acquire: async ({ affinityKey }) => ({
        ownership: "owned",
        resource_ids: [`browser:${affinityKey}`],
        resource: { affinityKey },
        release: async () => released.push(affinityKey),
      }),
    },
  };
  const broker = new FixtureBroker({ providers });
  const transaction = await broker.acquire("postgres_transaction", { unitID: "row-a" });
  assert.equal(transaction.record.ownership, "borrowed");
  await transaction.release();
  const first = await broker.acquire("browser_stack", { unitID: "group-a", affinityKey: "chain-a" });
  const second = await broker.acquire("browser_stack", { unitID: "group-b", affinityKey: "chain-a" });
  assert.equal(first.resource, second.resource, "browser affinity must reuse one stack lease");
  await first.release();
  await second.release();
  await broker.close();
  assert.deepEqual(released.sort(), ["chain-a", "row-a"]);

}

if (["fast", "matrix"].includes(mode)) {
  assertGraphCutover();
  assertAtomicNDJSON();
}
if (["scheduler-smoke", "scheduler-matrix", "matrix"].includes(mode)) assertScheduler();
if (["fixture-smoke", "fixture-matrix", "service-backed", "matrix"].includes(mode)) {
  await assertFixtures();
}
