import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, rmSync, symlinkSync, readdirSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { readExecutionIndex } from "../execution-reader.mjs";
import { analyzeExecution } from "../execution-analysis.mjs";
import { executionRecordsDigest, executionObservation, resumeExecutionObservation } from "../../runtime/execution-observations.mjs";
import { atomicLocalFile, readLocalFile } from "../../runtime/secure-local-files.mjs";
import { createRuntime } from "../../runtime/session-client.mjs";
import { createLaunchContext } from "../../runtime/launch-context.mjs";
import { openExecutionChannel } from "../../test-support/execution-channel-fixture.mjs";

const digest = `sha256:${"a".repeat(64)}`;
const id = "00000000-0000-4000-8000-000000000001";
const launch = { kind: "launch", ref: `launch:${id}`, run_id: "execution", invocation_id: id, parent_invocation_id: null,
  unit_id: "u", command_id: null, row_ids: ["r"], producer: "go", launched: true, outcome: "passed" };
function fixture(t) {
  const base = mkdtempSync(path.join(os.tmpdir(), "cartulary-execution-")), root = path.join(base, "execution");
  t.after(() => rmSync(base, { recursive: true, force: true }));
  const manifest = { run_id: "execution", source_digest: digest, graph_digest: digest, instrumentation: { mode: "basic", policy_digest: digest } };
  const run = { manifest, registrations: new Map([["u", { row_ids: ["r"] }]]) };
  const index = { schema_id: "cartulary.harness_execution_index.v2", graph_clock: null, run_id: "execution", source_digest: digest, graph_digest: digest,
    policy_digest: digest, status: "complete", omitted_records: 0, omitted_refs: [], records: [structuredClone(launch)] };
  const file = path.join(root, "diagnostics/execution-index.json");
  const save = () => atomicLocalFile(file, JSON.stringify({ ...index, records_digest: executionRecordsDigest(index.records) }), { replace: true });
  save();
  return { root, run, index, file, save };
}

test("execution reader rejects forged identity, duplicate records, cycles and false completeness", (t) => {
  for (const mutate of [
    (v) => { v.run_id = "foreign"; }, (v) => { v.source_digest = `sha256:${"b".repeat(64)}`; },
    (v) => { v.graph_digest = `sha256:${"b".repeat(64)}`; }, (v) => { v.policy_digest = `sha256:${"b".repeat(64)}`; },
    (v) => v.records.push(v.records[0]), (v) => { v.records[0].parent_invocation_id = id; },
    (v) => { v.records[0].unit_id = "foreign"; }, (v) => { v.records[0].row_ids = ["foreign"]; },
    (v) => { v.records[0].launched = false; }, (v) => { v.records[0].outcome = "incomplete"; },
    (v) => { v.records[0].parent_invocation_id = "00000000-0000-4000-8000-000000000002"; },
  ]) {
    const f = fixture(t); mutate(f.index); f.save(); assert.throws(() => readExecutionIndex(f.root, f.run));
  }
});

test("explicit omissions preserve incomplete parentage without manufacturing a launch", (t) => {
  const f = fixture(t), parent = "00000000-0000-4000-8000-000000000002";
  f.index.records[0].parent_invocation_id = parent;
  f.index.status = "partial"; f.index.omitted_records = 1; f.index.omitted_refs = [`launch:${parent}`]; f.save();
  const index = readExecutionIndex(f.root, f.run);
  assert.equal(index.records.length, 1);
  assert.equal(analyzeExecution(index).completeness, "partial");
});

test("execution reader rejects digest corruption, unsafe files, duplicate JSON keys and byte overflow", (t) => {
  for (const transform of [
    (f) => atomicLocalFile(f.file, JSON.stringify({ ...f.index, records_digest: digest }), { replace: true }),
    (f) => atomicLocalFile(f.file, '{"schema_id":1,"schema_id":2}', { replace: true }),
    (f) => atomicLocalFile(f.file, " ".repeat(4 * 1024 ** 2 + 1), { replace: true }),
    (f) => { const target = path.join(f.root, "elsewhere"); atomicLocalFile(target, "{}"); rmSync(f.file); symlinkSync(target, f.file); },
  ]) { const f = fixture(t); transform(f); assert.throws(() => readExecutionIndex(f.root, f.run)); }
});

test("failed acquisition cannot invent a successful lease or measured unavailable activity", (t) => {
  const f = fixture(t);
  f.index.records.push({ kind: "activity", local_start_ms: null, local_end_ms: null, anchors: [], ref: "activity:a", activity: "migration", unit_id: "u", invocation_id: id, allocation_ref: null,
    outcome: "failed", clock: "unavailable", clock_identity: null, resolution_ms: null, duration_ms: 0, availability: "unavailable" });
  f.save(); assert.throws(() => readExecutionIndex(f.root, f.run));
  f.index.records.pop();
  f.index.records.push({ kind: "allocation", ref: "allocation:a", allocation_ref: "allocation:a", unit_id: "u", ownership: null, capability: "browser_stack", outcome: "failed" },
    { kind: "lease", ref: "lease:a", lease_ref: "a", allocation_ref: "allocation:a", unit_id: "u", ownership: "owned", capability: "browser_stack" });
  atomicLocalFile(path.join(f.root, "_shared/fixture-leases/a.json"), JSON.stringify({ schema_id: "cartulary.harness_fixture_lease.v5", lease_id: "a", capability: "browser_stack", postgres_isolation: "exclusive", ownership: "owned", state: "leased", cleanup_outcome: "not_required", cleanup_failure_reason: null, resource_ids: [], created_at: "2026-10-09T00:00:00.000Z" }));
  f.save(); assert.throws(() => readExecutionIndex(f.root, f.run), /successful allocation/u);
});

test("bounded presentation retains outcomes, unavailable durations and exact list omissions", (t) => {
  const f = fixture(t);
  f.index.records = Array.from({ length: 25 }, (_, i) => ({ kind: "activity", local_start_ms: null, local_end_ms: null, anchors: [], ref: `activity:${i}`, activity: "migration", unit_id: "u", invocation_id: id,
    allocation_ref: null, outcome: "incomplete", clock: "unavailable", clock_identity: null, resolution_ms: null, duration_ms: null, availability: "not_observed" }));
  f.index.status = "partial";
  const view = analyzeExecution(f.index);
  assert.equal(view.activities.records.length, 20); assert.equal(view.activities.omitted, 5);
  assert.equal(view.activities.records[0].duration_ms, null);
});

test("private beginnings obey the byte bound and survive invalid replacement", (t) => {
  const f = fixture(t), repoRoot = path.resolve(import.meta.dirname, "../../../..");
  rmSync(f.file);
  const runtime = createRuntime({ repoRoot, runRoot: f.root, runID: "execution" });
  t.after(() => runtime.close());
  const session = openExecutionChannel(runtime);
  const environment = { CARTULARY_TEST_RUN_ID: "execution", CARTULARY_TEST_RESULTS_DIR: path.dirname(f.root),
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "execution", ...session.environment };
  const options = { repoRoot, environment }, root = runtime.privatePath("execution-observations");
  const oversized = { ...launch, row_ids: Array.from({ length: 1000 }, (_, i) => `row-${i}-${"a".repeat(70)}`) };
  assert.equal(executionObservation(options, oversized).token, null);
  assert.equal(readdirSync(root).some((name) => name.startsWith("slot-")), false);
  const initial = { ...launch, launched: false, outcome: "incomplete" };
  const observation = executionObservation(options, initial);
  const file = path.join(root, observation.token.slot);
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  observation.update({ ...initial, outcome: "unknown" });
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  observation.update(oversized);
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  const resumed = resumeExecutionObservation(options, observation.token);
  resumed.update(oversized);
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  observation.update(launch, {});
  assert.deepEqual(JSON.parse(readLocalFile(file)), initial);
  session.close(); observation.update(launch); assert.deepEqual(session.records(), [initial]);
});

test("spawn proof survives a short launch privately without retroactive retained process details", (t) => {
  const f = fixture(t), repoRoot = path.resolve(import.meta.dirname, "../../../..");
  rmSync(f.file);
  const runtime = createRuntime({ repoRoot, runRoot: f.root, runID: "execution" });
  t.after(() => runtime.close());
  const session = openExecutionChannel(runtime);
  const environment = { CARTULARY_TEST_RUN_ID: "execution", CARTULARY_TEST_RESULTS_DIR: path.dirname(f.root),
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "execution", ...session.environment };
  const child = createLaunchContext({ repoRoot, environment, unitID: "u", producer: "go" });
  child.spawned(process.pid); child.settle("passed");
  session.close();
  const [value]=session.records();
  assert.equal(value.process_proof.pid,process.pid);
  assert.equal(value.process_proof.start,value.process_proof.stat.start);
  assert.equal(value.record.invocation_id,child.identity.invocation_id);
  assert.equal(value.record.outcome,"passed");
  assert.equal(JSON.stringify(value.record).includes('"process_proof"'),false);
});
