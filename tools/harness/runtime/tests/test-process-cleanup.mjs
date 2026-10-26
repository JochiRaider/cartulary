import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { validateSchemaSync } from "../../contract/index.mjs";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import { spawn, spawnSync } from "node:child_process";
test("retired private process dispatch rejects without signaling a live owned process", async (t) => {
  const { once } = await import("node:events");
  const { createRuntime } = await import("../session-client.mjs");
  const { ownedProcess, stopOwnedProcess } = await import("../owned-process.mjs");
  const repository = path.resolve(import.meta.dirname, "../../../..");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "retired-process-dispatch-"));
  const runRoot = path.join(scratch, "results", "retired-dispatch");
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  const runtime = createRuntime({ repoRoot: repository, runRoot, runID: "retired-dispatch", scratchRoot: path.join(scratch, "private") });
  const child = spawn(process.execPath, ["-e", "process.stdout.write('ready');setInterval(()=>{},1000)"], { detached: true, stdio: ["ignore", "pipe", "ignore"] });
  const closed = once(child, "close");
  t.after(async () => { if (child.exitCode === null) {
    child.kill("SIGKILL");
    await closed;
  } });
  await once(child.stdout, "data");
  const proof = ownedProcess(child.pid);
  const resource = { kind: "helper_process", target: proof, state: "acquired", run_id: runtime.runID, lease_id: runtime.leaseID, depends_on: [], recovery_files: [] };
  validateSchemaSync("cartulary.harness_runtime_resource.v2", resource);
  runtime.record(resource);
  try {
    assert.throws(() => validateSchemaSync("cartulary.harness_runtime_resource.v2", { ...resource, kind: "preparation_process" }));
    const result = spawnSync(process.execPath, [path.join(repository, "tools/harness/runtime/resource-stop-private.mjs")], {
      cwd: repository, env: { ...process.env, ...runtime.producerEnvironment(), CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT: runRoot }, input: JSON.stringify(resource), timeout: 10000, encoding: "utf8"
    });
    assert.equal(result.error, undefined);
    assert.equal(result.status, 12);
    assert.equal(child.exitCode, null);
    assert.doesNotThrow(() => process.kill(child.pid, 0), "retired dispatch signaled a process");
  }
  finally {
    await stopOwnedProcess(proof);
    await closed;
    runtime.record({ ...resource, state: "released" });
    runtime.close();
    rmSync(scratch, { recursive: true, force: true });
  }
});
test("physical cleanup diagnostics remain bounded and cannot authorize success", async (t) => {
  const { stopOwnedProcess } = await import("../owned-process.mjs");
  const { findCleanupDiagnostic, withCleanupDiagnostic } = await import("../cleanup-diagnostic.mjs");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "cleanup-diagnostic-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  const valid = { schema_id: "cartulary.harness_cleanup_diagnostic.v4", stage: "process_signal", cause: "io", errno: 13 };
  const cases = [
    [JSON.stringify(valid), "process_signal"],
    [JSON.stringify({ ...valid, process_id: 123 }), "process_protocol"],
    [JSON.stringify({ ...valid, cause: "credential-sentinel" }), "process_protocol"],
    [JSON.stringify(valid) + JSON.stringify(valid), "process_protocol"],
    [" ".repeat(4097), "process_protocol"],
  ];
  for (const [index, [stderr, stage]] of cases.entries()) {
    const binary = path.join(scratch, `physical-${index}`);
    writeFileSync(binary, `#!${process.execPath}\nprocess.stdin.resume();process.stdin.on("end",()=>{process.stderr.write(${JSON.stringify(stderr)});process.exitCode=12;});\n`, { mode: 0o700 });
    await assert.rejects(stopOwnedProcess({}, { environment: { ...process.env, CARTULARY_HARNESS_RUNTIME_BIN: binary } }), error => {
      const diagnostic = findCleanupDiagnostic(new AggregateError([error], "private-path-sentinel"));
      assert.equal(diagnostic.stage, stage);
      assert.equal(JSON.stringify(diagnostic).includes("sentinel"), false);
      return true;
    });
  }
  const privateFailure = new Error("private-path-sentinel");
  Object.defineProperty(privateFailure, "privateDiagnostic", { value: { reason: "unsafe", errno: null, process_id: 123, proof_file: "environ" } });
  const projected = findCleanupDiagnostic(withCleanupDiagnostic(privateFailure, "browser_acknowledgement"));
  assert.deepEqual(projected, { schema_id: valid.schema_id, stage: "browser_acknowledgement", cause: "unsafe", errno: 0 });
  const wrapped = new AggregateError([new Error("private wrapper", { cause: privateFailure })], "private outer");
  assert.deepEqual(findCleanupDiagnostic(withCleanupDiagnostic(wrapped, "recovery_proof")), projected);
  const unprojected = new Error("private-path-sentinel");
  Object.defineProperty(unprojected, "privateDiagnostic", { value: privateFailure.privateDiagnostic });
  const nested = new AggregateError([new Error("private wrapper", { cause: unprojected })], "private outer");
  assert.deepEqual(findCleanupDiagnostic(withCleanupDiagnostic(nested, "recovery_proof")), { schema_id: valid.schema_id, stage: "recovery_proof", cause: "unsafe", errno: 0 });
});
test("cleanup evidence retains each failure without leaking private errors", async (t) => {
  const { CleanupResults } = await import("../cleanup-results.mjs");
  const { withCleanupDiagnostic } = await import("../cleanup-diagnostic.mjs");
  const root = mkdtempSync(path.join(os.tmpdir(), "cleanup-results-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  const results = new CleanupResults();
  const cause = new Error("credential-sentinel");
  Object.defineProperty(cause, "privateDiagnostic", { value: { reason: "io", errno: 13, proof_file: "credential-sentinel" } });
  results.record("fixture_release", { error: withCleanupDiagnostic(cause, "browser_namespace") });
  results.record("host_release", { error: new Error("another-private-path") });
  results.record("services_close", { blocked: true });
  results.record("runtime_close");
  results.publish(root, "cleanup-evidence");
  const bytes = readFileSync(path.join(root, "cleanup-results.json"), "utf8");
  assert.equal(bytes.includes("sentinel"), false);
  assert.equal(bytes.includes("private-path"), false);
  const record = JSON.parse(bytes);
  validateSchemaSync("cartulary.harness_cleanup_results.v2", record);
  assert.deepEqual(record.results.map(x => x.diagnostic?.stage ?? null), ["browser_namespace", "host_release", null, null]);
  assert.equal(record.results[0].diagnostic.errno, 13);
  for (const mutate of [
    r => { r.schema_id = "cartulary.harness_cleanup_results.v1"; },
    r => { r.results[0].diagnostic = null; },
    r => { r.results[0].diagnostic.schema_id = "cartulary.harness_cleanup_diagnostic.v2"; },
    r => { r.results[0].diagnostic.secret = "credential-sentinel"; },
    r => { r.results[0].diagnostic.stage = "arbitrary-private-path"; },
    r => { r.results[3].diagnostic = r.results[0].diagnostic; },
  ]) {
    const candidate = structuredClone(record);
    mutate(candidate);
    assert.throws(() => validateSchemaSync("cartulary.harness_cleanup_results.v2", candidate));
  }
});
