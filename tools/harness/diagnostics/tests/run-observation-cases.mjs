import assert from "node:assert/strict";
import { spawn, spawnSync } from "node:child_process";
import { Worker } from "node:worker_threads";
import { pathToFileURL } from "node:url";
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, statSync, symlinkSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { observeRun, renderObservation, nativeBinary } from "./native-observation.mjs";
import { liveSnapshotBytes, validateLiveSnapshot } from "../../contract/live-observation.mjs";
import { validateSchemaSync, validateResultRoot } from "../../contract/index.mjs";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { buildWorkGraph } from "../../scheduler/work-graph/model.mjs";

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
    schema_id: "cartulary.harness_run_manifest.v3", run_id: "selected", target: "test-slice",
    command_id: "cartulary.harness.command.test_slice.v1", declared_inputs: {},
    source_commit: "a".repeat(40), source_state: "dirty", source_digest: digest,
    toolchain_digest: digest, system_digest: digest, graph_digest: graph.graph_digest,
    instrumentation: { mode: "off", policy_digest: digest },
    cache_mode: "normal", started_at: "2026-10-09T00:00:00.000Z",
    go_worker_budget: 2, capability_snapshot: {
      schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 2, memory_bytes: 1024,
      process_slots: 2, io_tokens: 1, postgres_lanes: 1, object_store_lanes: 1,
      writable_volume: true, port_lanes: 1, services: {}, sources: {}, snapshot_digest: digest,
    },
  };
  validateSchemaSync(manifest.schema_id, manifest);
  atomicLocalFile(path.join(runRoot, "run-manifest.json"), `${JSON.stringify(manifest)}\n`);
  t.after(() => rmSync(parent, { recursive: true, force: true }));
  // Authored reader input, independent of either runtime publisher.
  const publisher = () => {
    replace({ schema_id: "cartulary.harness_live_snapshot.v3", run_id: manifest.run_id,
      public_target: manifest.target, command_id: manifest.command_id, source_digest: digest,
      graph_digest: graph.graph_digest, revision: 1, sampled_at: "2026-10-09T00:00:00.000Z",
      sampled_elapsed_ms: 0, last_transition_seq: 0, publication_mode: "active",
      phase: "executing", availability: "available", unavailable_reason: null,
      targets: [{target:"subset",unit_ids:[graph.units[0].unit_id]}, {target:"test-slice",unit_ids:graph.units.map(u=>u.unit_id)}],
      units:graph.units.map(u=>({unit_id:u.unit_id,status:"pending",activity:"pending",started_elapsed_ms:null,
        wait_reason:null,blocking_resources:[],blocking_unit_ids:[],cache_disposition:"not_observed",failure_class:null,failure_reason:null})) });
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
  test("paused cursor is unavailable without polling or confusing heartbeat with progress", async (t) => {
    const f=fixture(t);f.publisher();const first=f.snapshot();
    f.replace({...first,revision:2,sampled_elapsed_ms:5000});
    assert.equal(f.snapshot().last_transition_seq,first.last_transition_seq);
    f.replace({...f.snapshot(),revision:3,publication_mode:"paused_measurement",last_transition_seq:1});
    const result=await observeRun({resultsDir:f.runRoot,afterRevision:3,waitSeconds:30});
    assert.equal(result.wait_result,"unavailable");assert.ok(result.waited_ms<1000);assert.ok(result.limitations.includes("measurement_pause"));
  });
  test("observation expiry and cancellation preserve invocation bytes", async (t) => {
    const f = fixture(t); f.publisher();
    const before = readFileSync(f.snapshotFile);
    const result = await observeRun({ resultsDir: f.runRoot, afterRevision: 1, waitSeconds: 1 });
    assert.equal(result.wait_result, "wait_expired"); assert.equal(result.operation_exit_code, 0); assert.ok(result.waited_ms >= 1000);
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
    const f = fixture(t, 3000); f.publisher(); const large = f.snapshot();
    const units = large.units.map((unit) => ({ ...unit, status: "running", activity: "waiting", wait_reason: "resources", blocking_resources: Array.from({ length: 12 }, (_, i) => `resource-${String(i).padStart(2, "0")}`), blocking_unit_ids: large.units.slice(0, 12).map((other) => other.unit_id) }));
    f.replace({...large,units,last_transition_seq:5}); assert.ok(statSync(f.snapshotFile).size > 1024 ** 2);
    assert.throws(() => atomicLocalFile(f.snapshotFile, "{}", { replace: true }), /unsafe local file/u);
    const result = await observeRun({ resultsDir: f.runRoot });
    assert.equal(result.live.unit_counts.total, 3000); assert.equal(result.live.waiting.length, 16); assert.equal(result.live.waiting_omitted, 2984);
    assert.equal(result.live.waiting[0].blocking_unit_ids_omitted, 4);
    assert.ok(Buffer.byteLength(JSON.stringify(result)) < 65536); assert.ok(Buffer.byteLength(renderObservation(result)) <= 8192);
    // Cross the actual production cap, not just an injected low limit.
    const blockers = Array.from({ length: 24 }, (_, i) => `resource-${String(i).padStart(2, "0")}-${"x".repeat(240)}`);
    const oversized = units.map((unit) => ({ ...unit, blocking_resources: blockers }));
    assert.ok(Buffer.byteLength(JSON.stringify({ ...large, units: oversized })) > liveSnapshotBytes);
    const unavailable={...large,last_transition_seq:6,availability:"unavailable",unavailable_reason:"snapshot_too_large"};delete unavailable.units;delete unavailable.targets;f.replace(unavailable);
    assert.equal(f.snapshot().availability, "unavailable"); assert.equal(f.snapshot().units, undefined);
    assert.equal(f.snapshot().last_transition_seq, 6);
    assert.equal((await observeRun({ resultsDir: f.runRoot })).unavailable_reason, "snapshot_too_large");
  });
  test("concurrent readers admit only complete atomic replacements", async (t) => {
    const f = fixture(t, 3000); f.publisher();
    const workers = ["writer", "reader", "reader", "reader", "reader"].map((role) => new Worker(`
      const { workerData, parentPort } = require('node:worker_threads');
      (async () => {
        const { atomicLocalFile } = await import(workerData.secure);
        const { observeRun } = await import(workerData.reader);
        parentPort.on('message', async (revision) => {
          try {
            if (workerData.role === 'writer') {
              atomicLocalFile(workerData.file, JSON.stringify({ ...workerData.snapshot, revision }) + '\\n', { replace: true, maximumReplacementBytes: 16777216 });
            } else {
              const observed = await observeRun({ resultsDir: workerData.root });
              if (observed.operation_exit_code !== 0 || observed.live.unit_counts.total !== 3000) throw new Error(JSON.stringify(observed));
            }
            parentPort.postMessage('complete');
          } catch (error) { throw error; }
        });
        parentPort.postMessage('ready');
      })().catch(error => { throw error; });
    `, { eval: true, workerData: {
      role, root: f.runRoot, file: f.snapshotFile, snapshot: f.snapshot(),
      secure: pathToFileURL(path.join(root, "tools/harness/runtime/secure-local-files.mjs")).href,
      reader: pathToFileURL(path.join(root, "tools/harness/diagnostics/tests/native-observation.mjs")).href,
    } }));
    const acknowledgement = (child) => new Promise((resolve, reject) => {
      const cleanup = () => { child.off("message", complete); child.off("error", failed); child.off("exit", exited); };
      const complete = () => { cleanup(); resolve(); };
      const failed = (error) => { cleanup(); reject(error); };
      const exited = () => failed(new Error("reader fixture worker exited before acknowledgement"));
      child.once("message", complete); child.once("error", failed); child.once("exit", exited);
    });
    try {
      await Promise.all(workers.map(acknowledgement));
      for (let revision = 3; revision < 27; revision++) {
        const completed = workers.map(acknowledgement);
        for (const child of workers) child.postMessage(revision);
        await Promise.all(completed);
      }
    } finally {
      // Fixture deletion is registered by fixture(t); every worker must settle
      // before this test returns, including the first reader or writer failure.
      await Promise.all(workers.map((child) => child.terminate()));
    }
  });
  test("reader process interruption leaves observed invocation and leases untouched", async (t) => {
    const f = fixture(t); f.publisher();
    const before = readFileSync(f.snapshotFile);
    const lease = path.join(f.runRoot, "owned-lease"); writeFileSync(lease, "not owned by observer", { mode: 0o600 });
    const child = spawn(nativeBinary, ["diagnostics", "--results-dir", f.runRoot, "--after-revision", String(f.snapshot().revision), "--wait-seconds", "30", "--json"], { stdio: ["ignore", "pipe", "pipe"] });
    t.after(() => { if (child.exitCode === null) child.kill(); });
    let output = ""; child.stdout.on("data", (bytes) => { output += bytes; });
    const timer = setTimeout(() => child.kill("SIGTERM"), 1500);
    t.after(() => clearTimeout(timer));
    const code = await new Promise((resolve, reject) => { child.once("error", reject); child.once("exit", resolve); });
    assert.equal(code, 143); assert.equal(JSON.parse(output).operation_exit_code, 143);
    assert.deepEqual(readFileSync(f.snapshotFile), before); assert.equal(readFileSync(lease, "utf8"), "not owned by observer");
  });
  test("snapshot validation rejects duplicate or foreign rosters", (t) => {
    const f = fixture(t); f.publisher(); const snapshot = f.snapshot(); snapshot.targets[0].unit_ids.push("foreign");
    assert.throws(() => validateLiveSnapshot(snapshot, f.manifest));
    const duplicate = f.snapshot(); duplicate.units.push(duplicate.units[0]); assert.throws(() => validateLiveSnapshot(duplicate, f.manifest));
  });
}

export function registerRunStatusCommandTests() {
  test("result-root inspection preserves metadata and rejects links without creating directories", (t) => {
    const f = fixture(t);
    chmodSync(f.runRoot, 0o750);
    const before = statSync(f.runRoot, { bigint: true });
    assert.equal(validateResultRoot("selected", { root: f.parent }), f.runRoot);
    const after = statSync(f.runRoot, { bigint: true });
    for (const key of ["mode", "ctimeNs", "mtimeNs", "ino", "uid", "gid"])
      assert.equal(after[key], before[key], key);
    validateResultRoot("absent/deep", { root: f.parent });
    assert.deepEqual(readdirSync(f.parent), ["selected"]);
    for (const [name, destination] of [["link", f.runRoot], ["dangling", path.join(f.parent, "absent")]]) {
      symlinkSync(destination, path.join(f.parent, name));
      for (const candidate of [name, `${name}/child`, path.join(f.parent, name)])
        assert.throws(() => validateResultRoot(candidate, { root: f.parent }), /symlink/u);
    }
    validateResultRoot("selected", { root: f.parent, create: true });
    assert.equal(statSync(f.runRoot).mode & 0o777, 0o700);
    validateResultRoot("created", { root: f.parent, create: true });
    assert.equal(statSync(path.join(f.parent, "created")).mode & 0o777, 0o700);
  });
  test("entrypoint preflight cannot be assigned to a writer or an unregistered reader", async () => {
    const { validateMakeRecipes } = await import("../../generated-artifacts/task-surface/recipe-validation.mjs");
    const { loadExecutionTopology } = await import("../../generated-artifacts/execution-topology.mjs");
    const owner = loadExecutionTopology({ root }).taskSurface;
    const reject = (target, mutate) => {
      const targets = new Map(structuredClone(owner.targets).map((entry) => [entry.name, entry]));
      const recipes = structuredClone(owner.make_recipes);
      recipes[target].preflight_owner = "entrypoint";
      mutate?.(targets.get(target));
      const errors = [];
      validateMakeRecipes(errors, targets, owner.sequences, recipes);
      assert.ok(errors.some((error) => error.includes(`make_recipes.${target} requires an explicitly capable public read-only entrypoint`)), errors.join("\n"));
    };
    reject("test-run-status", (entry) => { entry.output_policy.artifact_policy = "tool_run_summary"; });
    reject("test-run-status", (entry) => { entry.side_effects = [{ class: "filesystem" }]; });
    reject("explain-run");
  });
  test("public status JSON ignores inherited selection and never provisions tooling", (t) => {
    const f = fixture(t); f.publisher();
    const environment = Object.fromEntries(Object.entries(process.env).filter(([key]) => !/^(?:CARTULARY_|MAKE|MFLAGS$|RESULTS_DIR$|RUN_ID$|TARGET$|OWNER$|ROWS$|JSON$|AFTER_REVISION$|WAIT_SECONDS$)/u.test(key)));
    const run = (args, extra = {}) => spawnSync("make", ["--no-print-directory", "test-run-status", `RESULTS_DIR=${f.runRoot}`, ...args], { cwd: root, env: { ...environment, ...extra }, encoding: "utf8", timeout: 15000, maxBuffer: 131072 });
    const result = run(["JSON=1"], { CARTULARY_TEST_RUN_ID: "sibling", TASK_SURFACE_MANIFEST: "/do-not-open", CARTULARY_TASK_SURFACE_MANIFEST: "/do-not-open", OWNER: "invalid-owner", ROWS: "not-a-row", CARTULARY_HARNESS_UNIT_EVENT_PREFIX: "invalid-private-prefix" });
    assert.equal(result.status, 0, result.stderr); assert.equal(JSON.parse(result.stdout).run.run_id, "selected");
    assert.equal(result.stdout.trim().split("\n").length, 1);
    for (const args of [["OWNER=invalid-owner"], ["UNDECLARED=1"], ["JSON=2"], ["WAIT_SECONDS=31"], ["WAIT_SECONDS=1"], ["JSON=1", "CARTULARY_OUTPUT_MODE=machine"], ["GO_CACHE_DIR=relative"], [`GO_CACHE_DIR=${root}/cache`], ["GO_CACHE_DIR=/tmp/same", "GO_MOD_CACHE_DIR=/tmp/same"]]) assert.notEqual(run(args).status, 0);
    const missing = run(["HARNESS_RUNTIME_BIN=/does-not-exist", "JSON=1"]);
    assert.notEqual(missing.status, 0); assert.match(missing.stderr, /installed diagnostic tool/u); assert.equal(missing.stdout, "");
    const marker = path.join(f.runRoot, "unexpected-command");
    for (const suffix of [`$(shell touch ${marker})`, `';touch ${marker};'`, `\`touch ${marker}\``]) {
      const literal = run([`RESULTS_DIR=${f.runRoot}${suffix}`, "JSON=1"]);
      assert.notEqual(literal.status, 0);
      assert.equal(existsSync(marker), false, "native exports executed an input value");
    }
    const mixed = run(["help", "JSON=1"]);
    assert.notEqual(mixed.status, 0); assert.match(mixed.stderr, /sole Make goal/u);
  });
}
