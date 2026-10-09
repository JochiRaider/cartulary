import assert from "node:assert/strict";
import { spawn, spawnSync } from "node:child_process";
import { Worker } from "node:worker_threads";
import { pathToFileURL } from "node:url";
import { chmodSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, statSync, symlinkSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { observeRun, renderObservation } from "../run-observation.mjs";
import { liveSnapshotBytes, validateLiveSnapshot } from "../../contract/live-observation.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { createLiveStatusPublisher } from "../../scheduler/work-graph/live-status.mjs";
import { buildWorkGraph } from "../../scheduler/work-graph/model.mjs";
import { runWorkGraph } from "../../scheduler/work-graph/scheduler.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const digest = `sha256:${"a".repeat(64)}`;
const zeroTiming = Object.fromEntries(["setup_ms", "fixture_ms", "execution_ms", "collation_ms", "wrapper_ms", "unattributed_ms", "resource_blocking_ms", "process_count"].map((key) => [key, 0]));
const graphUnit = (id, needs = [], overrides = {}) => ({
  unit_id: id, owner_id: "harness.evidence_accounting", kind: "runner",
  command: { executable: "true", args: [], environment: {} }, needs,
  resource_claims: { cpu: 1, process: 1 }, fixture_lease: "none", service_dependencies: [],
  cache_policy: "none", timeout_ms: 1000, current_run_evidence_outputs: [],
  failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" },
  estimated_work_ms: 10, ...overrides,
});

function fixture(t, count = 2) {
  const parent = mkdtempSync(path.join(os.tmpdir(), "cartulary-observation-"));
  const runRoot = path.join(parent, "selected");
  mkdirSync(runRoot, { mode: 0o700 });
  const graph = buildWorkGraph(Array.from({ length: count }, (_, index) => graphUnit(`unit-${String(index).padStart(5, "0")}`)));
  const manifest = {
    schema_id: "cartulary.harness_run_manifest.v1", run_id: "selected", target: "test-slice",
    command_id: "cartulary.harness.command.test_slice.v1", declared_inputs: {},
    source_commit: "a".repeat(40), source_state: "dirty", source_digest: digest,
    toolchain_digest: digest, system_digest: digest, graph_digest: graph.graph_digest,
    cache_mode: "normal", started_at: "2026-10-09T00:00:00.000Z",
    capability_snapshot: {
      schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 2, memory_bytes: 1024,
      process_slots: 2, io_tokens: 1, postgres_lanes: 1, object_store_lanes: 1,
      writable_volume: true, port_lanes: 1, services: {}, sources: {}, snapshot_digest: digest,
    },
  };
  validateSchemaSync(manifest.schema_id, manifest);
  atomicLocalFile(path.join(runRoot, "run-manifest.json"), `${JSON.stringify(manifest)}\n`);
  const publishers = [];
  t.after(() => { for (const publisher of publishers) publisher.stop(); rmSync(parent, { recursive: true, force: true }); });
  const publisher = (options = {}) => {
    const value = createLiveStatusPublisher({ runRoot, manifest, graph, projections: { "test-slice": graph.units.map((unit) => unit.unit_id), subset: [graph.units[0].unit_id] }, ...options });
    publishers.push(value);
    assert.equal(value.publicationError, null);
    return value;
  };
  const snapshotFile = path.join(runRoot, "diagnostics/live-status.json");
  const snapshot = () => JSON.parse(readFileSync(snapshotFile, "utf8"));
  const replace = (value) => atomicLocalFile(snapshotFile, `${JSON.stringify(value)}\n`, { replace: true, maximumReplacementBytes: liveSnapshotBytes });
  return { parent, runRoot, graph, manifest, publisher, snapshotFile, snapshot, replace };
}

function terminal(f, status = "fail") {
  const summary = {
    schema_id: "cartulary.harness_run_summary.v1", run_id: f.manifest.run_id, target: f.manifest.target,
    status, failure_class: status === "fail" ? "product" : null,
    failure_reason: status === "fail" ? "test_assertion_failure" : null,
    unit_counts: { total: 2, passed: status === "pass" ? 2 : 1, failed: status === "fail" ? 1 : 0, skipped: 0, cancelled: 0 },
    wall_duration_ms: 0, critical_path: [], actual_dependency_critical_path_ms: 0,
    timing_accounting: zeroTiming, resource_pressure: {}, cache: {}, artifact_refs: [],
  };
  atomicLocalFile(path.join(f.runRoot, "run-summary.json"), `${JSON.stringify(summary)}\n`, { replace: true });
}

export function registerRunObservationTests() {
  test("live observation binds exact runs, counts unique units and ignores staging", async (t) => {
    const f = fixture(t); f.publisher();
    writeFileSync(path.join(f.runRoot, "unit-events.ndjson.private"), "not a public input", { mode: 0 });
    const result = await observeRun({ resultsDir: f.parent, runId: "selected" });
    assert.equal(result.observation, "live"); assert.equal(result.live.unit_counts.total, 2);
    assert.equal((await observeRun({ resultsDir: f.runRoot, target: "subset" })).live.unit_counts.total, 1);
    assert.equal((await observeRun({ resultsDir: path.join(f.parent, "newer-missing") })).observation, "not_published");
    assert.equal(readdirSync(f.parent).includes("newer-missing"), false);
    assert.equal((await observeRun({ resultsDir: f.runRoot, target: "wrong" })).failure_reason, "usage_error");
  });
  test("heartbeat differs from progress and a paused matching cursor never polls", async (t) => {
    const f = fixture(t); let clock = 0; let timer;
    const timers = { setTimeout: (fn, ms) => { timer = { fn, due: clock + ms }; return timer; }, clearTimeout: () => { timer = null; } };
    const publisher = f.publisher({ now: () => clock, timers });
    const first = f.snapshot(); clock = 5000; timer.fn();
    assert.equal(f.snapshot().revision, first.revision + 1);
    assert.equal(f.snapshot().last_transition_seq, first.last_transition_seq);
    publisher.update({ units: first.units, phase: "executing", sampled_elapsed_ms: 0, last_transition_seq: 1 });
    publisher.pause(); const paused = f.snapshot(); assert.equal(timer, null);
    assert.equal(paused.sampled_elapsed_ms, 5000, "late scheduler admission preserves publisher elapsed time");
    const result = await observeRun({ resultsDir: f.runRoot, afterRevision: paused.revision, waitSeconds: 30 }, { sleep: () => { throw new Error("paused reader polled"); } });
    assert.equal(result.wait_result, "unavailable"); assert.ok(result.limitations.includes("measurement_pause"));
    publisher.resume(); assert.equal(f.snapshot().publication_mode, "active");
    publisher.stop(); const final = readFileSync(f.snapshotFile); assert.equal(timer, null);
    publisher.update({}); publisher.resume(); publisher.stop(); assert.deepEqual(readFileSync(f.snapshotFile), final);
  });
  test("observation expiry and cancellation preserve invocation bytes", async (t) => {
    const f = fixture(t); f.publisher(); let clock = 0; let polls = 0;
    const before = readFileSync(f.snapshotFile);
    const result = await observeRun({ resultsDir: f.runRoot, afterRevision: 1, waitSeconds: 3 }, { now: () => clock, sleep: async (ms) => { polls++; clock += ms; } });
    assert.equal(result.wait_result, "wait_expired"); assert.equal(result.operation_exit_code, 0); assert.equal(polls, 3);
    const controller = new AbortController();
    const cancelled = await observeRun({ resultsDir: f.runRoot, afterRevision: 1, waitSeconds: 30 }, { signal: controller.signal, sleep: async () => { controller.abort(); } });
    assert.equal(cancelled.failure_class, "interrupted"); assert.equal(cancelled.operation_exit_code, 15);
    assert.deepEqual(readFileSync(f.snapshotFile), before);
    assert.equal((await observeRun({ resultsDir: f.runRoot, afterRevision: 2 })).failure_reason, "usage_error");
    assert.equal((await observeRun({ resultsDir: f.runRoot, waitSeconds: 1 })).failure_reason, "usage_error");
  });
  test("terminal evidence takes precedence without asserting outer exit or acceptance", async (t) => {
    const f = fixture(t); f.publisher(); terminal(f);
    const result = await observeRun({ resultsDir: f.runRoot, afterRevision: 900, waitSeconds: 30 });
    assert.equal(result.operation_exit_code, 0); assert.equal(result.terminal.reported_status, "fail");
    assert.equal(result.terminal.failure_class, "product"); assert.equal(result.terminal.normalized_exit_code, null);
    terminal(f, "pass"); const pass = await observeRun({ resultsDir: f.runRoot });
    assert.equal(pass.terminal.evidence_audit, "not_run"); assert.ok(pass.limitations.includes("outer_exit_not_observed"));
    writeFileSync(path.join(f.runRoot, "run-summary.json"), "{broken", { mode: 0o600 });
    assert.equal((await observeRun({ resultsDir: f.runRoot })).failure_reason, "artifact_error");
  });
  test("tool-only failures retain normalized exits without synthetic graph evidence", async (t) => {
    const f = fixture(t); rmSync(path.join(f.runRoot, "run-manifest.json"));
    assert.equal((await observeRun({ resultsDir: f.runRoot })).observation, "not_published");
    const summary = {
      schema_id: "cartulary.tool_run_summary.v5", target: "lint-scripts",
      command: { cwd: root, argv: ["make", "lint-scripts"], make_target: "lint-scripts", env: {} },
      status: "fail", exit_code: 13, started_at: "2026-10-09T00:00:00Z", completed_at: "2026-10-09T00:00:01Z",
      duration_ms: 1000, output_mode: "summary", result_root: f.parent, run_id: "selected", run_root: f.runRoot,
      summary_artifacts: [], log_artifacts: [], work_units: [], evidence_targets: [], helper_units: [],
      counts: { steps: 0, tests: 0, failed: 0, non_test: 0, non_test_failed: 0, packages: 0 },
      step_accounting: Object.fromEntries(["authoritative", "support", "raw", "tooling_support", "unowned_regression", "unmapped", "authoritative_failed", "support_failed", "raw_failed", "tooling_support_failed", "unowned_regression_failed", "unmapped_failed", "missing"].map((key) => [key, 0])),
      failure_class: "timing", failure_reason: "timeout_failure", failures: [], slowest: [], warnings: [], rerun_commands: [], scheduler_timing: null, extensions: {},
    };
    const file = path.join(f.runRoot, "lint-scripts/tool-run-summary.json");
    atomicLocalFile(file, `${JSON.stringify(summary)}\n`);
    const result = await observeRun({ resultsDir: f.runRoot });
    assert.equal(result.operation_exit_code, 0); assert.equal(result.terminal.normalized_exit_code, 13);
    atomicLocalFile(file, `${JSON.stringify({ ...summary, run_id: "other" })}\n`, { replace: true });
    assert.equal((await observeRun({ resultsDir: f.runRoot })).failure_reason, "artifact_error");
  });
  test("unsafe paths, unsupported versions, oversize and identity conflicts fail closed", async (t) => {
    const f = fixture(t); f.publisher(); const valid = f.snapshot();
    for (const change of [{ schema_id: "cartulary.harness_live_snapshot.v99" }, { run_id: "sibling" }, { graph_digest: `sha256:${"b".repeat(64)}` }, { revision: -1 }]) {
      f.replace({ ...valid, ...change }); assert.equal((await observeRun({ resultsDir: f.runRoot })).failure_reason, "artifact_error");
    }
    f.replace(valid); const linked = path.join(f.parent, "linked"); symlinkSync(f.runRoot, linked);
    assert.equal((await observeRun({ resultsDir: linked })).failure_reason, "artifact_error");
    chmodSync(f.snapshotFile, 0o644); assert.equal((await observeRun({ resultsDir: f.runRoot })).failure_reason, "artifact_error");
    chmodSync(f.snapshotFile, 0o600); writeFileSync(f.snapshotFile, " ".repeat(liveSnapshotBytes + 1));
    assert.equal((await observeRun({ resultsDir: f.runRoot })).failure_reason, "artifact_error"); rmSync(f.snapshotFile);
  });
  test("large repeated atomic replacement and unavailable publication keep truthful bounds", async (t) => {
    const f = fixture(t, 3000); const publisher = f.publisher(); const large = f.snapshot();
    const units = large.units.map((unit) => ({ ...unit, status: "running", activity: "waiting", wait_reason: "resources", blocking_resources: Array.from({ length: 12 }, (_, i) => `resource-${String(i).padStart(2, "0")}`), blocking_unit_ids: large.units.slice(0, 12).map((other) => other.unit_id) }));
    publisher.update({ units, phase: "executing", last_transition_seq: 5, sampled_elapsed_ms: 0 }); publisher.pause(); publisher.resume();
    assert.equal(publisher.publicationError, null); assert.ok(statSync(f.snapshotFile).size > 1024 ** 2);
    assert.throws(() => atomicLocalFile(f.snapshotFile, "{}", { replace: true }), /unsafe local file/u);
    const result = await observeRun({ resultsDir: f.runRoot });
    assert.equal(result.live.unit_counts.total, 3000); assert.equal(result.live.waiting.length, 16); assert.equal(result.live.waiting_omitted, 2984);
    assert.equal(result.live.waiting[0].blocking_unit_ids_omitted, 4);
    assert.ok(Buffer.byteLength(JSON.stringify(result)) < 65536); assert.ok(Buffer.byteLength(renderObservation(result)) <= 8192);
    // Cross the actual production cap, not just an injected low limit.
    const blockers = Array.from({ length: 24 }, (_, i) => `resource-${String(i).padStart(2, "0")}-${"x".repeat(240)}`);
    const oversized = units.map((unit) => ({ ...unit, blocking_resources: blockers }));
    assert.ok(Buffer.byteLength(JSON.stringify({ ...large, units: oversized })) > liveSnapshotBytes);
    publisher.update({ units: oversized, phase: "executing", last_transition_seq: 6, sampled_elapsed_ms: 0 }); publisher.pause();
    assert.equal(f.snapshot().availability, "unavailable"); assert.equal(f.snapshot().units, undefined);
    assert.equal(f.snapshot().last_transition_seq, 6);
    assert.equal((await observeRun({ resultsDir: f.runRoot })).unavailable_reason, "snapshot_too_large"); publisher.stop();
  });
  test("concurrent readers admit only complete atomic replacements", async (t) => {
    const f = fixture(t, 3000); const publisher = f.publisher(); publisher.stop();
    const worker = (role) => new Promise((resolve, reject) => {
      const child = new Worker(`
        const { workerData, parentPort } = require('node:worker_threads');
        (async () => {
          const { atomicLocalFile } = await import(workerData.secure);
          const { observeRun } = await import(workerData.reader);
          for (let i = 0; i < 24; i++) {
            if (workerData.role === 'writer') {
              atomicLocalFile(workerData.file, JSON.stringify({ ...workerData.snapshot, revision: i + 3 }) + '\\n', { replace: true, maximumReplacementBytes: 16777216 });
            } else {
              const observed = await observeRun({ resultsDir: workerData.root });
              if (observed.operation_exit_code !== 0 || observed.live.unit_counts.total !== 3000) throw new Error(JSON.stringify(observed));
            }
            await new Promise(resolve => setTimeout(resolve, 2));
          }
          parentPort.postMessage('complete');
        })().catch(error => { throw error; });
      `, { eval: true, workerData: {
        role, root: f.runRoot, file: f.snapshotFile, snapshot: f.snapshot(),
        secure: pathToFileURL(path.join(root, "tools/harness/runtime/secure-local-files.mjs")).href,
        reader: pathToFileURL(path.join(root, "tools/harness/diagnostics/run-observation.mjs")).href,
      } });
      t.after(() => child.terminate());
      child.once("error", reject); child.once("exit", (code) => code ? reject(new Error(`worker exit ${code}`)) : resolve());
    });
    await Promise.all([worker("writer"), worker("reader"), worker("reader"), worker("reader"), worker("reader")]);
  });
  test("reader process interruption leaves observed invocation and leases untouched", async (t) => {
    const f = fixture(t); const publisher = f.publisher(); publisher.stop();
    const before = readFileSync(f.snapshotFile);
    const lease = path.join(f.runRoot, "owned-lease"); writeFileSync(lease, "not owned by observer", { mode: 0o600 });
    const child = spawn(process.execPath, [path.join(root, "tools/harness/diagnostics/test-run-status-cli.mjs"), "--results-dir", f.runRoot, "--after-revision", String(f.snapshot().revision), "--wait-seconds", "30", "--json"], { stdio: ["ignore", "pipe", "pipe"] });
    t.after(() => { if (child.exitCode === null) child.kill(); });
    let output = ""; child.stdout.on("data", (bytes) => { output += bytes; });
    const timer = setTimeout(() => child.kill("SIGTERM"), 1500);
    t.after(() => clearTimeout(timer));
    const code = await new Promise((resolve, reject) => { child.once("error", reject); child.once("exit", resolve); });
    assert.equal(code, 143); assert.equal(JSON.parse(output).operation_exit_code, 143);
    assert.deepEqual(readFileSync(f.snapshotFile), before); assert.equal(readFileSync(lease, "utf8"), "not owned by observer");
  });
  test("scheduler observation separates host, fixture, execution and cleanup boundaries", async () => {
    const graph = buildWorkGraph([graphUnit("a", [], { fixture_lease: "managed_process", shared_locks: [], exclusive_locks: ["host_activity"] })]);
    const views = []; const boundaries = [];
    const result = await runWorkGraph({ graph, capacities: new Map([["cpu", 1], ["process", 1]]), cwd: root, environment: {},
      observation: { update: (view) => views.push(view), pause: () => boundaries.push("pause"), resume: () => boundaries.push("resume") },
      hostAdmission: async () => { boundaries.push("host"); assert.equal(views.at(-1).units[0].wait_reason, "host_admission"); return { token: "unused", release: async () => boundaries.push("release") }; },
      fixtureBroker: { acquire: async () => { assert.equal(views.at(-1).units[0].activity, "preparing"); return { record: { lease_id: "fixture-a" }, release: async () => { assert.equal(views.at(-1).units[0].activity, "finalizing"); return { retained: false }; } }; }, close: async () => {} },
      executeUnit: async () => { assert.equal(views.at(-1).units[0].activity, "executing"); return { status: "passed", exit_code: 0 }; },
    });
    assert.equal(result.states.a, "passed"); assert.deepEqual(boundaries, ["pause", "host", "release", "resume"]);
    assert.equal(views.at(-1).phase, "finalizing"); assert.ok(views.every((view, i) => !i || view.last_transition_seq > views[i - 1].last_transition_seq));
  });
  test("scheduler failures retain independent completion and dependency skips despite sink failure", async () => {
    const graph = buildWorkGraph([graphUnit("a"), graphUnit("b"), graphUnit("c", ["a"])]); const views = [];
    const result = await runWorkGraph({ graph, capacities: new Map([["cpu", 2], ["process", 2]]), cwd: root, environment: {},
      observation: { update: (view) => { views.push(view); throw new Error("unavailable diagnostic sink"); } },
      executeUnit: async (unit) => unit.unit_id === "a" ? { status: "failed", failure_class: "product", failure_reason: "test_assertion_failure", exit_code: 1 } : { status: "passed", exit_code: 0 },
    });
    assert.deepEqual(result.states, { a: "failed", b: "passed", c: "skipped" }); assert.ok(views.some((view) => view.phase === "draining"));
    assert.equal(views.at(-1).units.find((unit) => unit.unit_id === "c").failure_class, null);
  });
  test("snapshot validation rejects duplicate or foreign rosters", (t) => {
    const f = fixture(t); f.publisher(); const snapshot = f.snapshot(); snapshot.targets[0].unit_ids.push("foreign");
    assert.throws(() => validateLiveSnapshot(snapshot, f.manifest));
    const duplicate = f.snapshot(); duplicate.units.push(duplicate.units[0]); assert.throws(() => validateLiveSnapshot(duplicate, f.manifest));
  });
}

export function registerRunStatusCommandTests() {
  test("public status JSON ignores inherited selection and never provisions tooling", (t) => {
    const f = fixture(t); f.publisher();
    const environment = Object.fromEntries(Object.entries(process.env).filter(([key]) => !/^(?:CARTULARY_|MAKE|MFLAGS$|RESULTS_DIR$|RUN_ID$|TARGET$|OWNER$|ROWS$|JSON$|AFTER_REVISION$|WAIT_SECONDS$)/u.test(key)));
    const run = (args, extra = {}) => spawnSync("make", ["--no-print-directory", "test-run-status", `RESULTS_DIR=${f.runRoot}`, ...args], { cwd: root, env: { ...environment, ...extra }, encoding: "utf8", timeout: 15000, maxBuffer: 131072 });
    const result = run(["JSON=1"], { CARTULARY_TEST_RUN_ID: "sibling", OWNER: "invalid-owner", ROWS: "not-a-row", CARTULARY_HARNESS_LIVE_UNIT_EVENTS_FILE: "/do-not-open" });
    assert.equal(result.status, 0, result.stderr); assert.equal(JSON.parse(result.stdout).run.run_id, "selected");
    assert.equal(result.stdout.trim().split("\n").length, 1);
    for (const args of [["JSON=2"], ["WAIT_SECONDS=31"], ["WAIT_SECONDS=1"], ["JSON=1", "CARTULARY_OUTPUT_MODE=machine"]]) assert.notEqual(run(args).status, 0);
    const missing = run(["NODE_BIN=/does-not-exist", "JSON=1"]);
    assert.notEqual(missing.status, 0); assert.match(missing.stderr, /installed Node runtime/u); assert.equal(missing.stdout, "");
    const mixed = run(["help", "JSON=1"]);
    assert.notEqual(mixed.status, 0); assert.match(mixed.stderr, /sole Make goal/u);
  });
}
