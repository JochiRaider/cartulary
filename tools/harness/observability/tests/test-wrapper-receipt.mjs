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
import { writeCanonicalFixture } from "./retained-fixture.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
const nativeBinary = path.join(root, ".cache/cartulary/toolbin/cartulary-harness");
function publishTerminalReceipts(directory) {
  const result = spawnSync(nativeBinary, ["diagnostics", "--terminal-receipt", directory], { cwd: root, encoding: "utf8" });
  if (result.status !== 0) throw Error("native receipt rejected");
  return readdirSync(directory).filter((name) => /^graph-[a-f0-9]{64}\.json$/u.test(name)).length;
}
const wrapper = path.join(root, "tools/harness/workspace/admission.sh");
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
    const runRoot = path.join(dir, id), { manifest } = writeCanonicalFixture(runRoot);
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
