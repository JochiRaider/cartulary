import assert from "node:assert/strict";
import test from "node:test";
import { createServer } from "node:http";
import { mkdtempSync, rmSync, writeFileSync, truncateSync, symlinkSync, readFileSync, lstatSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { spawn, spawnSync } from "node:child_process";
import { EventEmitter } from "node:events";
import { PassThrough } from "node:stream";
import { ReviewBrowser } from "../ui-review/browser.mjs";
import { diagnosticSnapshot, diagnosticEnvironment, parseDiagnosticOutput, checkDiagnosticScratch, diagnosticScratchLimit } from "../ui-review/diagnostic.mjs";
import { resolvePlaywrightPackages } from "../../readiness/playwright-packages.mjs";
import { captureCapabilitySnapshot, resourceCapacities } from "../../scheduler/work-graph/capability.mjs";
import { inheritedHostLease } from "../../runtime/host-admission.mjs";
import { repoRoot } from "../ui-review/policy.mjs";
import { diagnosticProcessScope, diagnosticProcesses, stopDiagnosticProcesses, bindDiagnosticProcess } from "../ui-review/diagnostic-processes.mjs";
import { ownedProcess, stopOwnedProcess } from "../ui-review/ownership.mjs";

test("bundled CLI is locally available and diagnostic output is bounded and strict", () => {
  const cli = path.join(resolvePlaywrightPackages(repoRoot).playwrightPath, "cli.js");
  const result = spawnSync(process.execPath, [cli, "cli", "--help"], { env: { PATH: "/usr/bin:/bin" }, encoding: "utf8", timeout: 10000 });
  assert.equal(result.status, 0); assert.match(result.stdout, /snapshot/u);
  assert.ok(!Object.keys(diagnosticEnvironment("/private", "token")).some((key) => key.startsWith("PLAYWRIGHT_")));
  for (const bytes of [Buffer.from("{"), Buffer.from('{"isError":true}'), Buffer.from('{"x":1,"x":2}'), Buffer.alloc(1048577)]) assert.throws(() => parseDiagnosticOutput(bytes));
});

test("diagnostic recovery reaps only its marked processes before daemon PID publication", async () => {
  const scope = diagnosticProcessScope();
  const otherScope = diagnosticProcessScope();
  const unrelated = spawn(process.execPath, ["-e", "setInterval(()=>{},1000)"], { detached: true, stdio: "ignore", env: { CARTULARY_DIAGNOSTIC_TOKEN: otherScope.token } });
  const otherProof = ownedProcess(unrelated.pid);
  const child = spawn(process.execPath, ["-e", "setInterval(()=>{},1000)"], { detached: true, stdio: "ignore", env: { CARTULARY_DIAGNOSTIC_TOKEN: scope.token } });
  try {
    assert.deepEqual(diagnosticProcesses(scope).map((entry) => entry.pid), [child.pid]);
    await stopDiagnosticProcesses(scope); await stopDiagnosticProcesses(scope);
    assert.deepEqual(diagnosticProcesses(scope), []);
    assert.ok(process.kill(unrelated.pid, 0));
    assert.deepEqual(diagnosticProcesses(otherScope).map((entry) => entry.pid), [unrelated.pid]);
    assert.deepEqual(diagnosticProcesses({ ...scope, boot: "previous-boot" }), []);
  } finally { await stopOwnedProcess(otherProof); child.kill("SIGKILL"); }
});

test("diagnostic scratch bounds opaque daemon files and rejects symlink escapes", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-diagnostic-storage-"));
  try {
    const file = path.join(root, "daemon.log"); writeFileSync(file, "private"); checkDiagnosticScratch(root);
    truncateSync(file, diagnosticScratchLimit + 1); assert.throws(() => checkDiagnosticScratch(root), /observation_limit/u);
    rmSync(file); symlinkSync(os.tmpdir(), file); assert.throws(() => checkDiagnosticScratch(root), /unsafe_artifact/u);
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test("diagnostic admission ignores expired exact guards but propagates live bind failures", async () => {
  const child = spawn(process.execPath, ["-e", "setInterval(()=>{},1000)"], { detached: true, stdio: "ignore" });
  const proof = ownedProcess(child.pid), failure = new Error("injected bind failure");
  const lease = { bind: async (guard) => { assert.deepEqual(guard, { boot: proof.boot, pid: proof.pid, start: proof.start }); throw failure; } };
  try { await assert.rejects(bindDiagnosticProcess(lease, proof), (error) => error === failure); }
  finally { await stopOwnedProcess(proof); }
  assert.equal(await bindDiagnosticProcess(lease, proof), false);
});

test("CLI diagnostic shares the exact owned page and preserves its state", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-diagnostic-test-"));
  const server = createServer((_request, response) => { response.setHeader("Content-Type", "text/html"); response.end('<h1>private diagnostic fixture</h1><input aria-label="field"><div style="height:3000px">content</div>'); });
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  const owner = new ReviewBrowser({ origin: `http://127.0.0.1:${server.address().port}`, mode: "dev" });
  const resources = new Map();
  try {
    await owner.start();
    await owner.page.locator("input").fill("private-field-state");
    await owner.page.locator("input").focus();
    await owner.page.evaluate(() => { localStorage.setItem("test", "unchanged"); window.scrollTo(0, 100); });
    await owner.snapshot();
    const refs = [...owner.references.keys()], epoch = owner.epoch;
    let observations;
    try {
      observations = await diagnosticSnapshot({ owner, runtime: { privatePath: (...parts) => path.join(root, ...parts) }, operationID: 1,
        signal: new AbortController().signal, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))), parentLease: inheritedHostLease(),
        onResource: (resource) => { const key = JSON.stringify([resource.kind, resource.target]); if (resource.state === "released") resources.delete(key); else resources.set(key, resource); } });
    } catch (error) { throw new Error(`diagnostic failed at ${error.diagnosticStage ?? "cleanup"}: ${error.diagnostic}; cause code ${error.cause?.code ?? "none"}`, { cause: undefined }); }
    assert.match(observations.accessibility_snapshot, /private diagnostic fixture/u);
    assert.equal(owner.epoch, epoch); assert.deepEqual([...owner.references.keys()], refs);
    assert.equal(await owner.page.locator("input").inputValue(), "private-field-state");
    for (const failure of ["spawn", "publication"]) {
      let helperPublished = false;
      await assert.rejects(() => diagnosticSnapshot({ owner, runtime: { privatePath: (...parts) => path.join(root, ...parts) }, operationID: failure === "spawn" ? 4 : 5,
        signal: new AbortController().signal, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))), parentLease: inheritedHostLease(),
        ...(failure === "spawn" ? { spawnProcess: () => {
          const child = new EventEmitter(); child.stdout = new PassThrough(); child.stderr = new PassThrough();
          child.kill = () => { throw new Error("must not signal a child without a PID"); };
          queueMicrotask(() => { child.emit("error", Object.assign(new Error("injected spawn failure"), { code: "EAGAIN" })); child.stdout.end(); child.stderr.end(); child.emit("close", -1); });
          return child;
        } } : {}),
        onResource: (resource) => {
          const key = JSON.stringify([resource.kind, resource.target]);
          if (resource.state === "released") resources.delete(key); else resources.set(key, resource);
          if (resource.kind === "helper_process" && resource.state !== "released") { helperPublished = true; throw new Error("injected publication failure"); }
        } }), (error) => error.diagnostic === "target_unavailable");
      assert.equal(helperPublished, failure === "publication");
      const detailFile = path.join(root, "artifacts", "diagnostic-failure.json");
      const detail = JSON.parse(readFileSync(detailFile, "utf8"));
      assert.equal(detail.operation_id, failure === "spawn" ? 4 : 5); assert.equal(detail.stage, "attach");
      assert.equal(detail.cause_code, failure === "spawn" ? "EAGAIN" : null);
      assert.equal(lstatSync(detailFile).mode & 0o777, 0o600); assert.ok(lstatSync(detailFile).size <= 65536);
      assert.equal(resources.size, 0); assert.ok(owner.browser.isConnected());
      assert.equal(await owner.page.locator("input").inputValue(), "private-field-state");
    }
    let acquired = 0;
    await assert.rejects(() => diagnosticSnapshot({ owner, runtime: { privatePath: (...parts) => path.join(root, ...parts) }, operationID: 3,
      signal: new AbortController().signal, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))), parentLease: inheritedHostLease(),
      onResource: (resource) => {
        const key = JSON.stringify([resource.kind, resource.target]);
        if (resource.state === "released") resources.delete(key); else resources.set(key, resource);
        if (resource.kind === "helper_process" && resource.state !== "released" && ++acquired === 2) process.kill(resource.target.pid, "SIGKILL");
      } }));
    assert.equal(resources.size, 0); assert.ok(owner.browser.isConnected());
    assert.equal(await owner.page.evaluate(() => localStorage.getItem("test")), "unchanged");
    assert.equal(resources.size, 0); assert.ok(owner.browser.isConnected());
    const cancelled = new AbortController();
    await assert.rejects(() => diagnosticSnapshot({ owner, runtime: { privatePath: (...parts) => path.join(root, ...parts) }, operationID: 2,
      signal: cancelled.signal, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))), parentLease: inheritedHostLease(),
      onResource: (resource) => {
        const key = JSON.stringify([resource.kind, resource.target]);
        if (resource.state === "released") resources.delete(key); else resources.set(key, resource);
        if (resource.kind === "helper_process" && resource.state !== "released") cancelled.abort();
      } }));
    assert.equal(resources.size, 0); assert.ok(owner.browser.isConnected());
    assert.equal(await owner.page.locator("input").inputValue(), "private-field-state");
  } finally { await owner.close(); await new Promise((resolve) => server.close(resolve)); rmSync(root, { recursive: true, force: true }); }
});
