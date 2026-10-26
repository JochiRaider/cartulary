import assert from "node:assert/strict";
import test from "node:test";
import path from "node:path";
import os from "node:os";
import { randomUUID, createHash } from "node:crypto";
import { mkdtempSync, rmSync, readFileSync, readlinkSync, readdirSync, symlinkSync } from "node:fs";
import { spawn, spawnSync } from "node:child_process";
import { once } from "node:events";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { readBootClock } from "../../runtime/boot-clock.mjs";
import { loadRetainedObservability } from "../observability.mjs";
import { readInvocationAccounting } from "../invocation-accounting.mjs";
import { receiptDigest } from "../terminal-receipt.mjs";
import { analyzePerformance } from "../performance-analysis.mjs";
import { buildReaderEvidence, loadFixtureIdentity, persistEvidence } from "./retained-fixture.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
const nativeBinary = path.join(root, ".cache/cartulary/toolbin/cartulary-harness");
function publishTerminalReceipts(directory) {
  const result = spawnSync(nativeBinary, ["diagnostics", "--terminal-receipt", directory], { cwd: root, encoding: "utf8" });
  if (result.status !== 0) throw Error("native receipt rejected");
  return readdirSync(directory).filter((name) => /^graph-[a-f0-9]{64}\.json$/u.test(name)).length;
}
const wrapper = path.join(root, "tools/harness/workspace/admission.sh");

test("accounting rejects missing, foreign, malformed and rebound whole-invocation facts", async (t) => {
  const f = fixture(t), a = f.graph(); f.complete(); publishTerminalReceipts(f.privateRoot);
  const retained = await loadRetainedObservability(a.runRoot);
  assert.equal(retained.run.cleanupStarted, 9);
  assert.equal(retained.run.cleanupCompleted, 10);
  const receiptBytes = readFileSync(path.join(a.runRoot, "diagnostics/terminal-receipt.json"));
  const receipt = retained.terminalReceipt, terminalDigest = receiptDigest(receiptBytes);
  const file = path.join(a.runRoot, "diagnostics/invocation-accounting.json");
  const value = { schema_id: "cartulary.harness_invocation_accounting.v1", invocation_digest: receipt.invocation_digest, source_digest: receipt.source_digest,
    graph_memberships: [{ run_digest: receiptDigest(retained.run.manifest.run_id), graph_digest: retained.run.manifest.graph_digest, receipt_digest: terminalDigest }],
    metrics: { duration_ms: 10, cpu_usage_us: 25000, memory_peak_bytes: 32768, initial_memory_peak_bytes: 4096, controller_setup_ms: 2, controller_publication_prefix_ms: 1,
      child_exit_code: 0, signal: 0, scope: "invocation_cgroup", memory_semantics: "peak_cgroup_charge", settlement: "empty" } };
  const read = (run = retained.run) => readInvocationAccounting(a.runRoot, run, receipt, terminalDigest);
  assert.equal(read(), null);
  assert.throws(() => read({ manifest: { ...retained.run.manifest, declared_inputs: { HARNESS_ACCOUNTING: "required" } } }));
  f.write(file, value); assert.deepEqual(read(), value);
  for (const mutate of [
    (v) => { v.invocation_digest = `sha256:${"a".repeat(64)}`; },
    (v) => { v.graph_memberships.push(v.graph_memberships[0]); },
    (v) => { v.graph_memberships[0].receipt_digest = `sha256:${"a".repeat(64)}`; },
    (v) => { v.metrics.memory_peak_bytes = -1; },
    (v) => { v.metrics.initial_memory_peak_bytes = v.metrics.memory_peak_bytes + 1; },
    (v) => { v.metrics.raw_path = "/private/sentinel"; },
    (v) => { v.metrics.child_exit_code = 1; },
    (v) => { v.metrics.duration_ms = 0; },
    (v) => { v.metrics.duration_ms = receipt.outer_duration_ms + 100; },
    (v) => { v.metrics.signal = 15; },
  ]) {
    const broken = structuredClone(value); mutate(broken); f.write(file, broken); assert.throws(read);
  }
});

test("queue and cleanup aggregates use all selected units before display truncation", () => {
  const ids = Array.from({ length: 50 }, (_, i) => `unit:${i}`);
  const run = { manifest: { target: "fixture", run_id: "fixture" }, summary: { critical_path: [], wall_duration_ms: 200, resource_pressure: { blocked_units: [], resource_holders: [] } },
    targetSummaries: new Map([["fixture", { target: "fixture", unit_ids: ids }]]),
    registrations: new Map(ids.map((id) => [id, { kind: "runner" }])),
    intervals: new Map(ids.map((id, i) => [id, { queue_ms: i, start: 50, end: 60 }])),
    waits: new Map(ids.map((id) => [id, { by_reason: {}, count: 1 }])),
    phases: new Map([["a", { unit_id: ids[0], phase: "fixture_release", start_ms: 10, end_ms: 30 }], ["b", { unit_id: ids[1], phase: "fixture_release", start_ms: 20, end_ms: 40 }]]),
    cleanupStarted: 100, cleanupCompleted: 130 };
  const value = analyzePerformance(run, "", { comparison: null, resources: {}, envelope: null });
  assert.equal(value.units.length, 20); assert.equal(value.units_omitted, 30);
  assert.deepEqual(value.selected_aggregates, { unit_count: 50, ready_to_start_p95_ms: 47, cleanup_union_ms: 60 });
});
function fixture(t) {
  const dir = mkdtempSync(path.join(os.tmpdir(), "cartulary-receipt-"));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  const owner = { version: 1, invocation_id: randomUUID(), pid: process.pid,
    start_ticks: readFileSync(`/proc/${process.pid}/stat`, "utf8").split(") ").at(-1).trim().split(/\s+/u)[19],
    start_ms: readBootClock().monotonic_ms, boot_id: readFileSync("/proc/sys/kernel/random/boot_id", "utf8").trim(), namespace: readlinkSync(`/proc/${process.pid}/ns/time`) };
  atomicLocalFile(path.join(dir, "private/owner.json"), JSON.stringify(owner));
  const privateRoot = path.join(dir, "private");
  const write = (file, value) => atomicLocalFile(file, JSON.stringify(value), { replace: true });
  const graph = (id = "graph1", scan = true) => {
    const runRoot = path.join(dir, id), evidence = buildReaderEvidence({ runID: id, ...loadFixtureIdentity() });
    persistEvidence(runRoot, evidence);
    const { manifest } = evidence;
    write(path.join(runRoot, "cleanup-results.json"), { schema_id: "cartulary.harness_cleanup_results.v2", run_id: manifest.run_id, results: [] });
    // Closed protocol data, not a replica of the registration lifecycle owner.
    const invocationDigest = `sha256:${createHash("sha256").update(owner.invocation_id).digest("hex")}`;
    const scanAt = readBootClock().monotonic_ms;
    const membership = { version: 1, invocation_digest: invocationDigest,
      run_root: runRoot, run_id: manifest.run_id, source_digest: manifest.source_digest,
      graph_digest: manifest.graph_digest, phase: "complete",
      scan_start_ms: scan === null ? null : scanAt, scan_end_ms: scan === null ? null : scanAt,
      scan_outcome: scan === null ? null : scan ? "passed" : "failed" };
    write(path.join(privateRoot, `graph-${createHash("sha256").update(runRoot).digest("hex")}.json`), membership);
    const envelope = { schema_id: "cartulary.harness_invocation_envelope.v2", run_id: id, source_digest: manifest.source_digest, graph_digest: manifest.graph_digest,
      invocation_digest: invocationDigest, boundary: "graph_main_entry_to_pre_scan", clock: "graph_process_monotonic", canonical_ref: "unit-events.ndjson",
      intervals: [], exclusions: ["make_and_preflight", "module_imports", "retained_secret_scan", "command_return", "receipt_publication"] };
    write(path.join(runRoot, "diagnostics/invocation-envelope.json"), envelope);
    if (scan) write(path.join(runRoot, "retained-secret-scan.json"), { schema_id: "cartulary.harness_retained_secret_scan.v1", status: "pass", scanned_files: 5, policy_digest: manifest.instrumentation.policy_digest });
    return { runRoot, envelope };
  };
  return { dir, owner, privateRoot, graph, write, complete(code = 0, signal = 0) { write(path.join(dir, "private/completion.json"), { end_ms: readBootClock().monotonic_ms, exit_code: code, signal, boot_id: owner.boot_id, namespace: owner.namespace }); } };
}
test("terminal receipt binds one outer duration to multiple graphs and detects later writes", async (t) => {
  const f = fixture(t), a = f.graph(), b = f.graph("graph2"); f.complete();
  assert.equal(publishTerminalReceipts(f.privateRoot), 2);
  const first = (await loadRetainedObservability(a.runRoot)).terminalReceipt;
  const second = (await loadRetainedObservability(b.runRoot)).terminalReceipt;
  assert.deepEqual(first, second); assert.equal(first.graph_memberships.length, 2); assert.equal(first.outcome, "passed");
  assert.throws(() => publishTerminalReceipts(f.privateRoot), /rejected/u);
  f.write(path.join(a.runRoot, "diagnostics/invocation-envelope.json"), { ...a.envelope, intervals: [{ phase: "collector_shutdown", start_ms: 0, end_ms: 1 }] });
  await assert.rejects(loadRetainedObservability(a.runRoot), /receipt envelope mismatch/u);
  f.write(path.join(a.runRoot, "diagnostics/invocation-envelope.json"), a.envelope);
  f.write(path.join(a.runRoot, "retained-secret-scan.json"), { schema_id: "cartulary.harness_retained_secret_scan.v1", status: "pass", scanned_files: 6, policy_digest: a.envelope.source_digest });
  await assert.rejects(loadRetainedObservability(a.runRoot), /receipt scan mismatch/u);
});
test("terminal receipt preserves failure and cancellation, while missing or failed scans cannot publish", async (t) => {
  for (const [code, signal, outcome] of [[7, 0, "failed"], [130, 2, "cancelled"]]) {
    const f = fixture(t), a = f.graph(); f.complete(code, signal); publishTerminalReceipts(f.privateRoot);
    const receipt = (await loadRetainedObservability(a.runRoot)).terminalReceipt;
    assert.equal(receipt.child_exit_code, code); assert.equal(receipt.outcome, outcome);
  }
  for (const scan of [false, null]) { const f = fixture(t); f.graph("graph1", scan); f.complete(); assert.throws(() => publishTerminalReceipts(f.privateRoot), /rejected/u); }
  const empty = fixture(t); empty.complete(2); assert.equal(publishTerminalReceipts(empty.privateRoot), 0);
});
test("wrapper publication rejects forged identities, unsafe files, overflow, and interrupted publication", async (t) => {
  const f = fixture(t), a = f.graph(); f.complete();
  const membership = path.join(f.privateRoot, readdirSync(f.privateRoot).find((name) => name.startsWith("graph-")));
  const saved = JSON.parse(readFileSync(membership, "utf8"));
  f.write(membership, { ...saved, invocation_digest: `sha256:${"0".repeat(64)}` }); assert.throws(() => publishTerminalReceipts(f.privateRoot), /rejected/u);
  f.write(membership, saved);
  f.write(path.join(a.runRoot, "diagnostics/invocation-envelope.json"), { ...a.envelope, source_digest: `sha256:${"0".repeat(64)}` }); assert.throws(() => publishTerminalReceipts(f.privateRoot), /rejected/u);
  f.write(path.join(a.runRoot, "diagnostics/invocation-envelope.json"), a.envelope);
  const receiptFile = path.join(a.runRoot, "diagnostics/terminal-receipt.json");
  symlinkSync("/dev/null", receiptFile); assert.throws(() => publishTerminalReceipts(f.privateRoot)); rmSync(receiptFile);
  atomicLocalFile(receiptFile, '{"schema_id":'); await assert.rejects(loadRetainedObservability(a.runRoot)); rmSync(receiptFile);
  for (let index = 0; index < 65; index++) f.write(path.join(f.privateRoot, `graph-${String(index).padStart(64, "0")}.json`), saved);
  assert.throws(() => publishTerminalReceipts(f.privateRoot), /rejected/u);
});
test("supervised admission preserves descriptor inheritance, nesting, early failure, and primary exit", () => {
  for (const mode of ["off", "basic"]) {
    const result = spawnSync("bash", [wrapper, "run", "--", "bash", "-c", 'bash "$1" verify && bash "$1" run -- bash -c \'test -r /proc/self/fd/9; exit 7\'', "fixture", wrapper], { cwd: root, env: { ...process.env, HARNESS_DIAGNOSTICS: mode }, encoding: "utf8" });
    assert.equal(result.status, 7, result.stderr); assert.equal(result.stdout, "shared\n");
  }
});
test("supervised admission forwards cancellation and waits for child cleanup", async (t) => {
  for (const signal of ["SIGINT", "SIGTERM"]) {
    const child = spawn("bash", [wrapper, "run", "--", process.execPath, "--input-type=module", "-e", 'process.on("SIGINT",()=>{console.log("cleaned");process.exit(130)});process.on("SIGTERM",()=>{console.log("cleaned");process.exit(143)});console.log("ready");setInterval(()=>{},1000)'], { cwd: root, env: { ...process.env, HARNESS_DIAGNOSTICS: "basic" }, stdio: ["ignore", "pipe", "pipe"] });
    t.after(() => { if (child.exitCode === null) child.kill("SIGKILL"); });
    let stdout = ""; const ready = new Promise((resolve) => child.stdout.on("data", (data) => { stdout += data; if (stdout.includes("ready")) resolve(); }));
    const completion = once(child, "close"); await ready; child.kill(signal);
    const [code] = await completion; assert.equal(code, signal === "SIGINT" ? 130 : 143); assert.match(stdout, /cleaned/u);
  }
});
