import { createRuntime } from "../../runtime/session-client.mjs";
import { closeReviewBrowserFixture } from "./ui-review-browser-fixture.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { createServer } from "node:http";
import { existsSync, mkdtempSync, rmSync, writeFileSync, truncateSync, symlinkSync, readFileSync, lstatSync } from "node:fs";
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

function fixtureChild(children, args, options) {
  const child = spawn(process.execPath, args, options);
  let error;
  child.once("error", (failure) => { error = failure; });
  const closed = new Promise(resolve => child.once("close", (code, signal) => resolve({ code, signal, error })));
  const owned = { child, closed, stop: async () => {
    if (child.exitCode === null && child.signalCode === null) child.kill("SIGKILL");
    await closed;
  } };
  children.push(owned);
  return owned;
}

test("bundled CLI is locally available and diagnostic output is bounded and strict", () => {
  const cli = path.join(resolvePlaywrightPackages(repoRoot).playwrightPath, "cli.js");
  const result = spawnSync(process.execPath, [cli, "cli", "--help"], { env: { PATH: "/usr/bin:/bin" }, encoding: "utf8", timeout: 10000 });
  assert.equal(result.status, 0); assert.match(result.stdout, /snapshot/u);
  assert.ok(!Object.keys(diagnosticEnvironment("/private", "token")).some((key) => key.startsWith("PLAYWRIGHT_")));
  for (const bytes of [Buffer.from("{"), Buffer.from('{"isError":true}'), Buffer.from('{"x":1,"x":2}'), Buffer.alloc(1048577)]) assert.throws(() => parseDiagnosticOutput(bytes));
});

test("native diagnostic recovery admits only persisted scopes after producer death", async (t) => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-diagnostic-authority-"));
  const children = [], scopes = [];
  let scratchRoot, runtime;
  t.after(async () => {
    const outcomes = await Promise.allSettled(children.map(child => child.stop()));
    outcomes.push(...await Promise.allSettled(scopes.map(async target => {
      await stopDiagnosticProcesses(runtime, target);
      runtime.record({ kind: "diagnostic_scope", target, state: "released" });
    })));
    const failures = outcomes.filter(result => result.status === "rejected").map(result => result.reason);
    if (failures.length) throw new AggregateError(failures, "diagnostic fixture cleanup failed");
    runtime?.close();
    if (scratchRoot) rmSync(scratchRoot, { recursive: true, force: true });
    rmSync(root, { recursive: true, force: true });
  });
  scratchRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-diagnostic-private-"));
  runtime = createRuntime({ repoRoot, runRoot: root, runID: "scope-authority", scratchRoot });
  const scope = diagnosticProcessScope(runtime, []); scopes.push(scope);
  const otherScope = diagnosticProcessScope(runtime, []); scopes.push(otherScope);
  const unrelated = fixtureChild(children, ["-e", "setInterval(()=>{},1000)"], { detached: true, stdio: "ignore", env: { CARTULARY_DIAGNOSTIC_TOKEN: otherScope.token } }).child;
  const producer = fixtureChild(children, ["-e", `const {spawn}=require('node:child_process');spawn(process.execPath,['-e','setInterval(()=>{},1000)'],{detached:true,stdio:'ignore',env:process.env}).unref();`],
    { stdio: "ignore", env: { CARTULARY_DIAGNOSTIC_TOKEN: scope.token } });
  const ended = await producer.closed;
  assert.equal(ended.error, undefined); assert.equal(ended.code, 0);
  // Every inspection and settlement is a different native process. The
  // producer has exited without publishing the detached daemon PID.
  const orphans = diagnosticProcesses(runtime, scope); assert.equal(orphans.length, 1);
  assert.throws(() => diagnosticProcesses(runtime, { ...scope, token: "a".repeat(64) }));
  assert.throws(() => diagnosticProcesses(runtime, { ...scope, boot: "previous-boot" }));
  await stopDiagnosticProcesses(runtime, scope); await stopDiagnosticProcesses(runtime, scope);
  assert.deepEqual(diagnosticProcesses(runtime, scope), []);
  assert.ok(process.kill(unrelated.pid, 0));
  assert.deepEqual(diagnosticProcesses(runtime, otherScope).map((entry) => entry.pid), [unrelated.pid]);
});

test("diagnostic scratch bounds opaque daemon files and rejects symlink escapes", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-diagnostic-storage-"));
  try {
    const file = path.join(root, "daemon.log"); writeFileSync(file, "private"); checkDiagnosticScratch(root);
    truncateSync(file, diagnosticScratchLimit + 1); assert.throws(() => checkDiagnosticScratch(root), /observation_limit/u);
    rmSync(file); symlinkSync(os.tmpdir(), file); assert.throws(() => checkDiagnosticScratch(root), /unsafe_artifact/u);
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test("diagnostic admission ignores expired exact guards but propagates live bind failures", async (t) => {
  const children = [];
  t.after(async () => { await children[0]?.stop(); });
  const child = fixtureChild(children, ["-e", "setInterval(()=>{},1000)"], { detached: true, stdio: "ignore" }).child;
  const proof = ownedProcess(child.pid), failure = new Error("injected bind failure");
  const lease = { bind: async (guard) => { assert.deepEqual(guard, { boot: proof.boot, pid: proof.pid, start: proof.start }); throw failure; } };
  try { await assert.rejects(bindDiagnosticProcess(lease, proof), (error) => error === failure); }
  finally { await stopOwnedProcess(proof); }
  assert.equal(await bindDiagnosticProcess(lease, proof), false);
});

test("CLI diagnostic shares the exact owned page and preserves its state", async (t) => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-diagnostic-test-"));
  const server = createServer((_request, response) => { response.setHeader("Content-Type", "text/html"); response.end('<h1>private diagnostic fixture</h1><input aria-label="field"><div style="height:3000px">content</div>'); });
  let owner, scratchRoot, runtime;
  t.after(async () => {
    await closeReviewBrowserFixture({ browser: owner, server });
    runtime?.close();
    if (scratchRoot && !existsSync(path.join(scratchRoot, "diagnostic-failure.json"))) rmSync(scratchRoot, { recursive: true, force: true });
    rmSync(root, { recursive: true, force: true });
  });
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  owner = new ReviewBrowser({ origin: `http://127.0.0.1:${server.address().port}`, mode: "dev" });
  const resources = new Map();
  scratchRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-diagnostic-runtime-"));
  runtime = createRuntime({ repoRoot, runRoot: root, runID: "diagnostic-browser", scratchRoot });
  await owner.start();
  await owner.page.setViewportSize({ width: 1024, height: 768 });
  await owner.page.evaluate(() => { document.documentElement.style.zoom = "1.25"; });
  await owner.page.locator("input").fill("private-field-state");
  await owner.page.locator("input").focus();
  await owner.page.evaluate(() => { localStorage.setItem("test", "unchanged"); window.scrollTo(0, 100); });
  await owner.snapshot();
  const refs = [...owner.references.keys()], epoch = owner.epoch;
  const assertPreserved = async () => {
    assert.equal(owner.epoch, epoch); assert.deepEqual([...owner.references.keys()], refs);
    assert.deepEqual(await owner.page.evaluate(() => {
      const field = document.querySelector("input");
      return { url: location.href, viewport: [innerWidth, innerHeight], zoom: document.documentElement.style.zoom,
        focused: document.activeElement === field, scroll: [scrollX, scrollY],
        value: field.value, stored: localStorage.getItem("test") };
    }), { url: `${owner.origin}/`, viewport: [1024, 768], zoom: "1.25",
      focused: true, scroll: [0, 100], value: "private-field-state", stored: "unchanged" });
  };
  await assertPreserved();
  let observations;
  try {
    observations = await diagnosticSnapshot({ owner, runtime, dependencies: [], operationID: 1,
      signal: new AbortController().signal, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))), parentLease: inheritedHostLease(),
      onResource: (resource) => { runtime.record(resource); const key = JSON.stringify([resource.kind, resource.target]); if (resource.state === "released") resources.delete(key); else resources.set(key, resource); } });
  } catch (error) {
    // Keep failure detail private for causal inspection instead of deleting it
    // with the runtime in the test finalizer. Never print its contents.
    const detail = runtime.privatePath("artifacts", "diagnostic-failure.json");
    if (existsSync(detail)) writeFileSync(path.join(scratchRoot, "diagnostic-failure.json"), readFileSync(detail), { mode: 0o600 });
    throw new Error(`diagnostic failed at ${error.diagnosticStage ?? "cleanup"}: ${error.diagnostic}; cause code ${error.cause?.code ?? "none"}`, { cause: undefined });
  }
  assert.match(observations.accessibility_snapshot, /private diagnostic fixture/u);
  await assertPreserved();
  for (const failure of ["spawn", "publication"]) {
    let helperPublished = false;
    await assert.rejects(() => diagnosticSnapshot({ owner, runtime, dependencies: [], operationID: failure === "spawn" ? 4 : 5,
      signal: new AbortController().signal, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))), parentLease: inheritedHostLease(),
      ...(failure === "spawn" ? { spawnProcess: () => {
        const child = new EventEmitter(); child.stdout = new PassThrough(); child.stderr = new PassThrough();
        child.kill = () => { throw new Error("must not signal a child without a PID"); };
        queueMicrotask(() => { child.emit("error", Object.assign(new Error("injected spawn failure"), { code: "EAGAIN" })); child.stdout.end(); child.stderr.end(); child.emit("close", -1); });
        return child;
      } } : {}),
      onResource: (resource) => {
        runtime.record(resource); const key = JSON.stringify([resource.kind, resource.target]);
        if (resource.state === "released") resources.delete(key); else resources.set(key, resource);
        if (resource.kind === "helper_process" && resource.state !== "released") { helperPublished = true; throw new Error("injected publication failure"); }
      } }), (error) => error.diagnostic === "target_unavailable");
    assert.equal(helperPublished, failure === "publication");
    const detailFile = runtime.privatePath("artifacts", "diagnostic-failure.json");
    const detail = JSON.parse(readFileSync(detailFile, "utf8"));
    assert.equal(detail.operation_id, failure === "spawn" ? 4 : 5); assert.equal(detail.stage, "attach");
    assert.equal(detail.cause_code, failure === "spawn" ? "EAGAIN" : null);
    assert.equal(lstatSync(detailFile).mode & 0o777, 0o600); assert.ok(lstatSync(detailFile).size <= 65536);
    assert.equal(resources.size, 0); assert.ok(owner.browser.isConnected());
    await assertPreserved();
  }
  let acquired = 0;
  await assert.rejects(() => diagnosticSnapshot({ owner, runtime, dependencies: [], operationID: 3,
    signal: new AbortController().signal, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))), parentLease: inheritedHostLease(),
    onResource: (resource) => {
      runtime.record(resource); const key = JSON.stringify([resource.kind, resource.target]);
      if (resource.state === "released") resources.delete(key); else resources.set(key, resource);
      if (resource.kind === "helper_process" && resource.state !== "released" && ++acquired === 2) process.kill(resource.target.pid, "SIGKILL");
    } }));
  assert.equal(resources.size, 0); assert.ok(owner.browser.isConnected());
  await assertPreserved();
  assert.equal(resources.size, 0); assert.ok(owner.browser.isConnected());
  const cancelled = new AbortController();
  await assert.rejects(() => diagnosticSnapshot({ owner, runtime, dependencies: [], operationID: 2,
    signal: cancelled.signal, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))), parentLease: inheritedHostLease(),
    onResource: (resource) => {
      runtime.record(resource); const key = JSON.stringify([resource.kind, resource.target]);
      if (resource.state === "released") resources.delete(key); else resources.set(key, resource);
      if (resource.kind === "helper_process" && resource.state !== "released") cancelled.abort();
    } }));
  assert.equal(resources.size, 0); assert.ok(owner.browser.isConnected());
  await assertPreserved();
});
