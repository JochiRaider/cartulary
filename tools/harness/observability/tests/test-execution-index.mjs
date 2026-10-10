import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, rmSync, symlinkSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { readExecutionIndex } from "../execution-reader.mjs";
import { analyzeExecution } from "../execution-analysis.mjs";
import { executionRecordsDigest, createExecutionSession, executionObservation } from "../../runtime/execution-observations.mjs";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { createSuiteRuntime } from "../../runtime/suite-runtime.mjs";
import { createLaunchContext } from "../../runtime/launch-context.mjs";
import { instrumentationPolicy } from "../resource-collector.mjs";

const digest = `sha256:${"a".repeat(64)}`;
const id = "00000000-0000-4000-8000-000000000001";
const launch = { kind: "launch", ref: `launch:${id}`, run_id: "execution", invocation_id: id, parent_invocation_id: null,
  unit_id: "u", command_id: null, row_ids: ["r"], producer: "go", launched: true, outcome: "passed" };
function fixture(t) {
  const base = mkdtempSync(path.join(os.tmpdir(), "cartulary-execution-")), root = path.join(base, "execution");
  t.after(() => rmSync(base, { recursive: true, force: true }));
  const manifest = { run_id: "execution", source_digest: digest, graph_digest: digest, instrumentation: { mode: "basic", policy_digest: digest } };
  const run = { manifest, registrations: new Map([["u", { row_ids: ["r"] }]]) };
  const index = { schema_id: "cartulary.harness_execution_index.v1", run_id: "execution", source_digest: digest, graph_digest: digest,
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
  f.index.records.push({ kind: "activity", ref: "activity:a", activity: "migration", unit_id: "u", invocation_id: id, allocation_ref: null,
    outcome: "failed", clock: "unavailable", clock_identity: null, resolution_ms: null, duration_ms: 0, availability: "unavailable" });
  f.save(); assert.throws(() => readExecutionIndex(f.root, f.run));
  f.index.records.pop();
  f.index.records.push({ kind: "allocation", ref: "allocation:a", allocation_ref: "allocation:a", unit_id: "u", ownership: null, capability: "browser_stack", outcome: "failed" },
    { kind: "lease", ref: "lease:a", lease_ref: "a", allocation_ref: "allocation:a", unit_id: "u", ownership: "owned", capability: "browser_stack" });
  atomicLocalFile(path.join(f.root, "_shared/fixture-leases/a.json"), JSON.stringify({ schema_id: "cartulary.harness_fixture_lease.v4", lease_id: "a", capability: "browser_stack", ownership: "owned", state: "leased", cleanup_outcome: "not_required", cleanup_failure_reason: null, resource_ids: [], created_at: "2026-10-09T00:00:00.000Z" }));
  f.save(); assert.throws(() => readExecutionIndex(f.root, f.run), /successful allocation/u);
});

test("bounded presentation retains outcomes, unavailable durations and exact list omissions", (t) => {
  const f = fixture(t);
  f.index.records = Array.from({ length: 25 }, (_, i) => ({ kind: "activity", ref: `activity:${i}`, activity: "migration", unit_id: "u", invocation_id: id,
    allocation_ref: null, outcome: "incomplete", clock: "unavailable", clock_identity: null, resolution_ms: null, duration_ms: null, availability: "not_observed" }));
  f.index.status = "partial";
  const view = analyzeExecution(f.index);
  assert.equal(view.activities.records.length, 20); assert.equal(view.activities.omitted, 5);
  assert.equal(view.activities.records[0].duration_ms, null);
});

test("slot cap exhaustion publish truthful truncation without changing execution", (t) => {
  const f = fixture(t), repoRoot = path.resolve(import.meta.dirname, "../../../..");
  rmSync(f.file);
  const runtime = createSuiteRuntime({ repoRoot, runRoot: f.root, runID: "execution" });
  t.after(() => runtime.close());
  const { policy } = instrumentationPolicy();
  const session = createExecutionSession({ runtime, runRoot: f.root, manifest: f.run.manifest, policy: { ...policy, maximum_execution_records: 1 } });
  const environment = { CARTULARY_TEST_RUN_ID: "execution", CARTULARY_TEST_RESULTS_DIR: path.dirname(f.root),
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "execution", ...session.environment };
  const first = createLaunchContext({ repoRoot, environment, unitID: "u" }); first.spawned(); first.settle("passed");
  const child = createLaunchContext({ repoRoot, environment: { ...environment, ...first.environment }, unitID: "u" }); child.settle("spawn_failed");
  session.capture(); session.publish();
  const index = readExecutionIndex(f.root, f.run);
  assert.equal(index.status, "truncated"); assert.equal(index.records.length, 1);
});

test("corrupt private beginnings become explicit omissions and shutdown rejects late writes", (t) => {
  const f = fixture(t), repoRoot = path.resolve(import.meta.dirname, "../../../..");
  rmSync(f.file);
  const runtime = createSuiteRuntime({ repoRoot, runRoot: f.root, runID: "execution" });
  t.after(() => runtime.close());
  const { policy } = instrumentationPolicy();
  const session = createExecutionSession({ runtime, runRoot: f.root, manifest: f.run.manifest, policy });
  const environment = { CARTULARY_TEST_RUN_ID: "execution", CARTULARY_TEST_RESULTS_DIR: path.dirname(f.root),
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "execution", ...session.environment };
  const options = { repoRoot, environment };
  const observation = executionObservation(options, launch);
  atomicLocalFile(runtime.privatePath("execution-observations", observation.token.slot), "broken", { replace: true });
  const childID = "00000000-0000-4000-8000-000000000002";
  executionObservation(options, { ...launch, ref: `launch:${childID}`, invocation_id: childID, parent_invocation_id: id });
  session.capture(); observation.update(launch); session.publish();
  const index = readExecutionIndex(f.root, f.run);
  assert.equal(index.status, "partial"); assert.deepEqual(index.omitted_refs, [`launch:${id}`]);
  assert.equal(index.records.length, 1);
});


test("spawn proof survives a short launch privately without retroactive retained process details", (t) => {
  const f = fixture(t), repoRoot = path.resolve(import.meta.dirname, "../../../..");
  rmSync(f.file);
  const runtime = createSuiteRuntime({ repoRoot, runRoot: f.root, runID: "execution" });
  t.after(() => runtime.close());
  const proofs = [];
  const session = createExecutionSession({ runtime, runRoot: f.root, manifest: f.run.manifest,
    policy: instrumentationPolicy().policy, onProcessProof: (proof, record) => proofs.push({ proof, record }) });
  const environment = { CARTULARY_TEST_RUN_ID: "execution", CARTULARY_TEST_RESULTS_DIR: path.dirname(f.root),
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "execution", ...session.environment };
  const child = createLaunchContext({ repoRoot, environment, unitID: "u", producer: "go" });
  child.spawned(process.pid); child.settle("passed");
  session.capture(); session.publish();
  const index = readExecutionIndex(f.root, f.run);
  assert.equal(proofs.length, 1);
  assert.equal(proofs[0].proof.pid, process.pid);
  assert.equal(proofs[0].proof.start, proofs[0].proof.stat.start);
  assert.equal(proofs[0].record.invocation_id, child.identity.invocation_id);
  assert.equal(index.records[0].outcome, "passed");
  assert.equal(JSON.stringify(index).includes('"process_proof"'), false);
  assert.equal(JSON.stringify(index).includes('"pid"'), false);
});
