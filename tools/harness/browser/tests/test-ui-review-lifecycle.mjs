import { closeReviewBrowser } from "./ui-review-browser-fixture.mjs";
import assert from "node:assert/strict";
import { chmodSync, existsSync, linkSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { createServer } from "node:http";
import os from "node:os";
import path from "node:path";
import { spawn, spawnSync } from "node:child_process";
import test from "node:test";
import { acquireHostAdmission, inheritedHostLease, processIdentityAlive } from "../../runtime/host-admission.mjs";
import { runPrivateCapturedProcess } from "../../runtime/private-child-process.mjs";
import { captureCapabilitySnapshot, resourceCapacities } from "../../scheduler/work-graph/capability.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile } from "../../runtime/secure-local-files.mjs";
import { ReviewBrowser } from "../ui-review/browser.mjs";
import { schemaID, validate, ReviewFailure, failureMappings, limits } from "../ui-review/contract.mjs";
import { inputPath } from "../ui-review/input-files.mjs";
import { repoRoot } from "../ui-review/policy.mjs";
import { ownedProcess, stopOwnedProcess } from "../ui-review/ownership.mjs";
import { createRuntime } from "../../runtime/session-client.mjs";

const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
test("failed runner acquisition publication reaps the child and preserves the primary error", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-capture-publication-"));
  const scratchRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-capture-scratch-"));
  const runtime = createRuntime({ repoRoot, runRoot: root, runID: "capture-publication", scratchRoot });
  const options = { dependencies: [], cwd: repoRoot, repoRoot, runRoot: root, detached: true, env: { ...process.env,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runtime.runID } };
  let proof;
  try {
    await assert.rejects(runPrivateCapturedProcess(process.execPath, ["-e", "setInterval(()=>{},1000)"], { ...options, onSpawn(child) { proof = ownedProcess(child.pid); throw new Error("injected publication failure"); } }), (error) => error.cause?.message === "injected publication failure");
    assert.ok(proof); assert.equal(processIdentityAlive(proof), false);
    await assert.rejects(runPrivateCapturedProcess(path.join(root, "missing-executable"), [], { ...options, onSpawn() { assert.fail("failed spawn must not publish an owned process"); } }), (error) => error.cause?.code === "ENOENT");
  } finally { if (proof) await stopOwnedProcess(proof); runtime.close(); rmSync(scratchRoot, { recursive: true, force: true }); rmSync(root, { recursive: true, force: true }); }
});
test("process recovery rejects PID reuse and preserves ambiguous leaderless ownership", async () => {
  const child = spawn(process.execPath, ["--input-type=module", "--eval", `import {spawn} from 'node:child_process'; const child=spawn(process.execPath,['-e','process.on("SIGTERM",()=>{});setInterval(()=>{},1000)'],{stdio:'ignore'});process.stdout.write(String(child.pid));setTimeout(()=>process.exit(),200);`], { detached: true, stdio: ["ignore", "pipe", "ignore"] });
  const proof = ownedProcess(child.pid); let descendant = "";
  child.stdout.on("data", (bytes) => { descendant += bytes; });
  const ended = new Promise((resolve) => child.once("exit", resolve));
  try {
    await assert.rejects(stopOwnedProcess({ ...proof, start: String(BigInt(proof.start) + 1n) }));
    assert.equal(child.exitCode, null);
    await ended; assert.ok(Number(descendant) > 1);
    await assert.rejects(stopOwnedProcess(proof, { graceMS: 40 }));
    const stat = existsSync(`/proc/${descendant}/stat`) ? readFileSync(`/proc/${descendant}/stat`, "utf8") : "";
    assert.ok(stat && !stat.split(") ").at(-1).startsWith("Z "));
  } finally {
    // The fixture owns this exact child; an ambiguous production proof must not.
    if (Number(descendant) > 1) process.kill(Number(descendant), "SIGKILL");
    if (child.exitCode === null) child.kill("SIGKILL");
  }
});
test("private local reads reject links, permissive files and changed publication targets", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-private-files-"));
  try {
    const file = path.join(root, "value.json"); atomicLocalFile(file, "first");
    assert.equal(readLocalFile(file).toString(), "first");
    assert.throws(() => atomicLocalFile(file, "overwrite"));
    atomicLocalFile(file, "next", { replace: true }); assert.equal(readLocalFile(file).toString(), "next");
    chmodSync(file, 0o644); assert.throws(() => readLocalFile(file)); chmodSync(file, 0o600);
    const alias = path.join(root, "alias.json"); linkSync(file, alias); assert.throws(() => readLocalFile(file)); rmSync(alias);
    symlinkSync(file, alias); assert.throws(() => readLocalFile(alias)); assert.throws(() => atomicLocalFile(alias, "bad", { replace: true }));
    const linkedDirectory = path.join(root, "linked"); symlinkSync(root, linkedDirectory); assert.throws(() => readLocalFile(path.join(linkedDirectory, "value.json")));
    assert.throws(() => readLocalFile(file, { maximum: 3 }));
    assert.throws(() => inputPath("README.md", ".json"));
    assert.throws(() => inputPath("neutral.markdown", ".json"));
  } finally { rmSync(root, { recursive: true, force: true }); }
});
test("host admission preserves exclusive waiter priority and bounded browser capacity", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-admission-")); const leases = [];
  try {
    const first = await acquireHostAdmission({ root, claims: { browser_stack: 1 }, capacities: { browser_stack: 1 } }); leases.push(first);
    await assert.rejects(acquireHostAdmission({ root, claims: { browser_stack: 1 }, capacities: { browser_stack: 1 }, timeoutMs: 50 }), /admission/u);
    const order = [];
    const quiet = acquireHostAdmission({ root, mode: "exclusive", capacities: { browser_stack: 1 } }).then((lease) => { leases.push(lease); order.push("quiet"); return lease; });
    await pause(100);
    const shared = acquireHostAdmission({ root, capacities: { browser_stack: 1 } }).then((lease) => { leases.push(lease); order.push("shared"); return lease; });
    await first.release(); const exclusive = await quiet;
    await pause(100); assert.deepEqual(order, ["quiet"]);
    await exclusive.release(); await shared; assert.deepEqual(order, ["quiet", "shared"]);
  } finally { for (const lease of leases) await lease.release(); rmSync(root, { recursive: true, force: true }); }
});
test("host admission mutex is released even while the scheduler parent blocks its event loop", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-admission-parent-"));
  let lease;
  try {
    const first = acquireHostAdmission({ root, capacities: { browser_stack: 1 } });
    const source = `import { acquireHostAdmission, inheritedHostLease } from ${JSON.stringify(new URL("../../runtime/host-admission.mjs", import.meta.url).href)}; const start=performance.now();const lease=await acquireHostAdmission({root:${JSON.stringify(root)},capacities:{browser_stack:1}});await lease.release();process.stdout.write(String(performance.now()-start));`;
    const child = spawn(process.execPath, ["--input-type=module", "--eval", source], { stdio: ["ignore", "pipe", "pipe"] });
    let output = "", error = ""; child.stdout.on("data", (part) => { output += part; }); child.stderr.on("data", (part) => { error += part; });
    const ended = new Promise((resolve) => child.once("close", resolve));
    Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 6000);
    lease = await first; assert.equal(await ended, 0, error); assert.ok(Number(output) < 4000);
  } finally { await lease?.release(); rmSync(root, { recursive: true, force: true }); }
});
test("browser actions use exact handles, epochs and isolated contexts without modifying borrowed service ownership", async () => {
  let requests = 0;
  const server = createServer((_request, response) => { requests++; response.setHeader("content-type", "text/html"); response.end('<!doctype html><html lang="en"><title>Review fixture</title><body><h1>Fixture</h1><input data-testid="entry" aria-label="Entry"><button data-testid="go" onclick="document.querySelector(\'h1\').textContent=\'Clicked\'">Go</button><button>Duplicate</button><button>Duplicate</button><button disabled>Disabled</button></body></html>'); });
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  const origin = `http://127.0.0.1:${server.address().port}`;
  const browser = new ReviewBrowser({ origin, mode: "dev" });
  let admission;
  const action = (name, parameters = {}, expected_epoch = browser.epoch) => ({ schema_id: schemaID("action"), expected_epoch, action: name, parameters });
  try {
    admission = await acquireHostAdmission({ parent: inheritedHostLease(), claims: { browser_stack: 1 }, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))) });
    await browser.start(); const snapshot = await browser.action(action("snapshot")); validate("observations", snapshot);
    const reference = snapshot.elements.find((element) => element.name === "Entry").resolved_ref;
    await browser.action(action("fill", { target: { kind: "element_ref", value: reference, epoch: 0 }, text: 'literal $(touch /no-shell)' }));
    assert.equal(await browser.page.getByTestId("entry").inputValue(), 'literal $(touch /no-shell)');
    await assert.rejects(browser.action(action("click", { target: { kind: "test_id", value: "go" } }, 0)), /session_mismatch/u);
    await assert.rejects(browser.action(action("click", { target: { kind: "role", role: "button", name: "Duplicate" } })), /target_unavailable/u);
    await assert.rejects(browser.action(action("click", { target: { kind: "role", role: "button", name: "Disabled" } })), /target_unavailable/u);
    assert.equal(await browser.page.locator("h1").textContent(), "Fixture");
    const next = await browser.action(action("snapshot"));
    const entry = next.elements.find((element) => element.name === "Entry");
    await browser.page.getByTestId("entry").evaluate((node) => node.replaceWith(node.cloneNode()));
    await assert.rejects(browser.action(action("fill", { target: { kind: "element_ref", value: entry.resolved_ref, epoch: browser.epoch }, text: "replacement" })), /target_unavailable/u);
    await assert.rejects(browser.action(action("authenticate", { actor: "admin" })), /invalid_request/u);
    await browser.action(action("click", { target: { kind: "test_id", value: "go" } })); assert.equal(await browser.page.locator("h1").textContent(), "Clicked");
    await browser.page.evaluate(() => {
      document.body.style.height = "3000px";
      window.addEventListener("scroll", () => { if (scrollY > 50 && !document.querySelector('[data-testid="virtual-row"]')) { const row = document.createElement("button"); row.dataset.testid = "virtual-row"; row.textContent = "Rendered row"; row.style.cssText = "position:fixed;top:20px"; document.body.append(row); } });
    });
    await assert.rejects(browser.resolve({ kind: "test_id", value: "virtual-row" }), /target_unavailable/u);
    const beforeZoom = browser.epoch;
    await browser.action(action("zoom", { percent: 125 }));
    assert.equal(browser.epoch, beforeZoom + 1);
    assert.equal(await browser.page.evaluate(() => getComputedStyle(document.documentElement).zoom), "1.25");
    await browser.action(action("zoom", { percent: 100 }));
    await browser.action(action("scroll", { x: 0, y: 100 }));
    await browser.page.getByTestId("virtual-row").waitFor({ state: "visible" });
    const row = await browser.resolve({ kind: "test_id", value: "virtual-row" }); await row.dispose();
    await closeReviewBrowser(browser); assert.equal((await fetch(origin)).status, 200); assert.ok(requests > 0);
  } finally { await closeReviewBrowser(browser); await admission?.release(); await new Promise((resolve) => server.close(resolve)); }
});





test("browser timeout requires a fresh snapshot, redirects stay at the borrowed origin, and browser loss is reported", async () => {
  const server = createServer((request, response) => {
    if (request.url === "/redirect") { response.writeHead(302, { location: "https://example.invalid/private" }); response.end(); return; }
    response.setHeader("content-type", "text/html"); response.end('<html><body><button data-testid="covered">Covered</button><div style="position:fixed;inset:0;z-index:9"></div></body></html>');
  });
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  let lost;
  const browser = new ReviewBrowser({ origin: `http://127.0.0.1:${server.address().port}`, mode: "dev", onLost: (error) => { lost = error; } });
  let admission;
  const action = (name, parameters = {}) => ({ schema_id: schemaID("action"), expected_epoch: browser.epoch, action: name, parameters });
  try {
    admission = await acquireHostAdmission({ parent: inheritedHostLease(), claims: { browser_stack: 1 }, capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))) });
    await browser.start();
    await assert.rejects(browser.action(action("click", { target: { kind: "test_id", value: "covered" } })), /operation_expired/u);
    await assert.rejects(browser.action(action("resize", { width: 800, height: 600 })), /session_mismatch/u);
    await browser.action(action("snapshot"));
    await assert.rejects(browser.action(action("navigate", { path: "/redirect" })), /navigation_boundary/u);
    await browser.browser.close(); assert.equal(lost.diagnostic, "session_lost");
    assert.equal((await fetch(browser.origin)).status, 200);
  } finally { await closeReviewBrowser(browser); await admission?.release(); await new Promise((resolve) => server.close(resolve)); }
});
function cleanEnvironment() {
  const env = { ...process.env };
  for (const name of Object.keys(env)) if (/^(?:MAKE|MFLAGS|CARTULARY_(?:HARNESS|MAKE|TEST|OUTPUT)|OWNER$|ROWS$|UI_)/u.test(name)) delete env[name];
  return env;
}
test("public artifact sessions retain only receipts and repeated stop preserves the exact terminal digest", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-review-public-")); privateDirectory(root);
  const child = spawn("make", ["--silent", "ui-review", "UI_MODE=artifacts"], { cwd: repoRoot, env: { ...cleanEnvironment(), CARTULARY_TEST_RESULTS_DIR: root, CARTULARY_TEST_RUN_ID: "session" }, stdio: ["ignore", "pipe", "pipe"] });
  let output = ""; child.stdout.on("data", (part) => { output += part; }); let stderr = ""; child.stderr.on("data", (part) => { stderr += part; });
  const ended = new Promise((resolve) => child.once("exit", resolve));
  const locator = path.join(root, "session/ui-review/session.json");
  const invoke = (command) => {
    const run = spawnSync("make", ["--silent", command, `UI_SESSION=${locator}`], { cwd: repoRoot, env: { ...cleanEnvironment(), CARTULARY_OUTPUT_MODE: "machine" }, encoding: "utf8", timeout: 30000 });
    assert.equal(run.status, 0, `${command} failed`); const value = JSON.parse(run.stdout); validate("command_result", value); return value;
  };
  try {
    for (let index = 0; index < 150 && !output.includes("UI review ready"); index++) { if (child.exitCode !== null) break; await pause(100); }
    assert.ok(output.includes("UI review ready"), `review failed to start: ${stderr}`);
    const status = invoke("ui-review-status"); assert.equal(status.state, "ready"); assert.equal(status.epoch, null);
    const stopped = invoke("ui-review-stop"); assert.equal(stopped.state, "closed"); assert.deepEqual(invoke("ui-review-stop").receipt, stopped.receipt);
    assert.equal(await ended, 0); assert.equal(invoke("ui-review-status").state, "closed");
    assert.ok(existsSync(locator)); assert.equal(readFileSync(path.join(root, "session/ui-review/terminal.json"), "utf8").includes("private_refs"), false);
  } finally { if (child.exitCode === null) child.kill("SIGTERM"); await ended; rmSync(root, { recursive: true, force: true }); }
});

test("controller death is recovered from exact ownership and retains a failed terminal outcome", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-review-death-"));
  const child = spawn("make", ["--silent", "ui-review", "UI_MODE=artifacts"], { cwd: repoRoot, env: { ...cleanEnvironment(), CARTULARY_TEST_RESULTS_DIR: root, CARTULARY_TEST_RUN_ID: "death" }, stdio: ["ignore", "pipe", "pipe"] });
  let output = ""; child.stdout.on("data", (part) => { output += part; }); child.stderr.resume();
  const ended = new Promise((resolve) => child.once("exit", resolve));
  try {
    for (let index = 0; index < 100 && !output.includes("UI review ready"); index++) await pause(100);
    assert.ok(output.includes("UI review ready"));
    const locator = path.join(root, "death/ui-review/session.json"); const identity = JSON.parse(readFileSync(locator));
    const registry = path.join(os.tmpdir(), `cartulary-ui-review-${process.getuid()}`, `${identity.session_id.slice(9)}.json`);
    const record = JSON.parse(readLocalFile(registry));
    process.kill(record.process.pid, "SIGKILL");
    assert.equal(await ended, 2);
    const final = JSON.parse(readFileSync(locator)); assert.equal(final.state, "failed"); assert.ok(final.terminal_receipt);
    const receipt = JSON.parse(readFileSync(path.join(root, "death", final.terminal_receipt.path)));
    assert.equal(receipt.failures[0].diagnostic_code, "session_lost"); assert.equal(receipt.cleanup, "complete");
    assert.equal(existsSync(record.runtime.root), false); assert.equal(existsSync(registry), false);
  } finally { if (child.exitCode === null) child.kill("SIGTERM"); await ended; rmSync(root, { recursive: true, force: true }); }
});

for (const death of ["parent", "both"]) test(`${death} supervisor death permits exact recovery without engine loading`, async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-review-supervisors-"));
  const server = createServer((_request, response) => response.end("<html><body>Borrowed origin</body></html>"));
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  const origin = `http://127.0.0.1:${server.address().port}`;
  const mode = death === "both" ? ["UI_MODE=dev", `UI_ORIGIN=${origin}`] : ["UI_MODE=artifacts"];
  const child = spawn("make", ["--silent", "ui-review", ...mode], { cwd: repoRoot, env: { ...cleanEnvironment(), CARTULARY_TEST_RESULTS_DIR: root, CARTULARY_TEST_RUN_ID: death }, stdio: ["ignore", "pipe", "pipe"] });
  let output = ""; child.stdout.on("data", (part) => { output += part; }); child.stderr.resume();
  const ended = new Promise((resolve) => child.once("exit", resolve));
  const locator = path.join(root, death, "ui-review/session.json"); let record;
  // The loader fails closed if either control path imports an engine, simulating
  // broken installations without changing packages used by concurrent tasks.
  const control = (command) => spawnSync(process.execPath, ["--input-type=module", "--eval", `import {registerHooks} from 'node:module'; registerHooks({resolve(specifier,context,next){if(['sharp','playwright','playwright-core','@axe-core/playwright','axe-core'].some(name=>specifier===name||specifier.startsWith(name+'/')||specifier.includes('/node_modules/'+name+'/')))throw Error('engine unavailable');return next(specifier,context);}});process.argv=[process.execPath,'cli',${JSON.stringify(command)}];await import(${JSON.stringify(new URL("../ui-review/cli.mjs", import.meta.url).href)});`], { cwd: repoRoot, env: { ...cleanEnvironment(), UI_SESSION: locator, CARTULARY_MAKE_INPUT_SOURCES: "UI_SESSION=cli", CARTULARY_OUTPUT_MODE: "machine" }, encoding: "utf8", timeout: 30000 });
  try {
    for (let index = 0; index < 150 && !output.includes("UI review ready"); index++) { if (child.exitCode !== null) break; await pause(100); }
    assert.ok(output.includes("UI review ready"));
    assert.equal(JSON.parse(control("ui-review-status").stdout).state, "ready");
    const identity = JSON.parse(readFileSync(locator));
    const registry = path.join(os.tmpdir(), `cartulary-ui-review-${process.getuid()}`, `${identity.session_id.slice(9)}.json`);
    record = JSON.parse(readLocalFile(registry));
    const parent = Number(readFileSync(`/proc/${record.process.pid}/stat`, "utf8").split(") ").at(-1).split(" ")[1]);
    if (death === "both") { process.kill(parent, "SIGSTOP"); process.kill(record.process.pid, "SIGKILL"); }
    process.kill(parent, "SIGKILL"); await ended;
    if (death === "both") {
      assert.notEqual(control("ui-review-status").status, 0);
      assert.equal(existsSync(record.runtime.root), true);
    } else for (let index = 0; index < 100 && JSON.parse(readFileSync(locator)).state !== "failed"; index++) await pause(20);
    const stopped = control("ui-review-stop"); const value = JSON.parse(stopped.stdout);
    assert.equal(value.state, "failed", stopped.stdout); assert.equal(value.exit_code, 3);
    assert.ok(value.receipt); assert.equal(existsSync(record.runtime.root), false); assert.equal(existsSync(registry), false);
    assert.deepEqual(JSON.parse(control("ui-review-stop").stdout), value);
    assert.equal((await fetch(origin)).status, 200);
  } finally {
    if (child.exitCode === null) child.kill("SIGTERM");
    if (record && existsSync(record.runtime.root)) control("ui-review-stop");
    await ended; await new Promise((resolve) => server.close(resolve)); rmSync(root, { recursive: true, force: true });
  }
});

test("stop interrupts a busy public dev action and leaves the borrowed origin alive", async () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-review-cancel-"));
  const navigationStarted = Promise.withResolvers();
  const server = createServer((request, response) => { response.setHeader("content-type", "text/html"); if (request.url === "/slow") { navigationStarted.resolve(); const timer = setTimeout(() => response.end("late"), 30000); response.once("close", () => clearTimeout(timer)); } else response.end("<html><body>Ready</body></html>"); });
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  const origin = `http://127.0.0.1:${server.address().port}`;
  const child = spawn("make", ["--silent", "ui-review", "UI_MODE=dev", `UI_ORIGIN=${origin}`], { cwd: repoRoot, env: { ...cleanEnvironment(), CARTULARY_TEST_RESULTS_DIR: root, CARTULARY_TEST_RUN_ID: "dev" }, stdio: ["ignore", "pipe", "pipe"] });
  let output = ""; child.stdout.on("data", (part) => { output += part; }); child.stderr.resume();
  const ended = new Promise((resolve) => child.once("exit", resolve));
  const locator = path.join(root, "dev/ui-review/session.json");
  const env = { ...cleanEnvironment(), CARTULARY_OUTPUT_MODE: "machine" };
  let actionChild;
  try {
    for (let index = 0; index < 150 && !output.includes("UI review ready"); index++) { if (child.exitCode !== null) break; await pause(100); }
    assert.ok(output.includes("UI review ready"));
    const file = path.join(root, "action.json"); atomicLocalFile(file, JSON.stringify({ schema_id: schemaID("action"), expected_epoch: 0, action: "navigate", parameters: { path: "/slow" } }));
    actionChild = spawn("make", ["--silent", "ui-browser", `UI_SESSION=${locator}`, `UI_REQUEST=${file}`], { cwd: repoRoot, env, stdio: ["ignore", "pipe", "pipe"] });
    let actionOutput = ""; actionChild.stdout.on("data", (part) => { actionOutput += part; }); actionChild.stderr.resume();
    const actionEnded = new Promise((resolve) => actionChild.once("exit", resolve));
    // Synchronize on the actual navigation, not Make/Node startup taking less
    // than two seconds under aggregate load. The stop-latency assertion below
    // starts only once the browser has begun the action being interrupted.
    let admissionTimer;
    try {
      await Promise.race([
        navigationStarted.promise,
        actionEnded.then(() => { throw new Error("navigation ended before the borrowed origin received it"); }),
        new Promise((_, reject) => { admissionTimer = setTimeout(() => reject(new Error("navigation did not reach the borrowed origin")), 15000); }),
      ]);
    } finally { clearTimeout(admissionTimer); }
    assert.equal(JSON.parse(readFileSync(locator)).state, "busy");
    const tick = performance.now();
    const stopped = spawnSync("make", ["--silent", "ui-review-stop", `UI_SESSION=${locator}`], { cwd: repoRoot, env, encoding: "utf8", timeout: 10000 });
    assert.equal(stopped.status, 0); assert.equal(JSON.parse(stopped.stdout).state, "closed"); assert.ok(performance.now() - tick < 5000);
    assert.equal(await actionEnded, 2); const response = JSON.parse(actionOutput); validate("command_result", response); assert.equal(response.failures[0].diagnostic_code, "interrupted"); assert.deepEqual(response.private_refs, []);
    assert.equal(await ended, 0); assert.equal((await fetch(origin)).status, 200);
  } finally { if (child.exitCode === null) child.kill("SIGTERM"); if (actionChild?.exitCode === null) actionChild.kill("SIGTERM"); await ended; await new Promise((resolve) => server.close(resolve)); rmSync(root, { recursive: true, force: true }); }
});
