import "./test-browser-acquisition.mjs";
import assert from "node:assert/strict";
import { existsSync, mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { browserReadiness, coreReadiness, goReadiness, nodeReadiness, publishFrontendInstallation, serviceImageReadiness, validateFrontendInstallation } from "../../readiness/installed-readiness.mjs";
import { createRuntime } from "../../runtime/session-client.mjs";
import { recipeEnvironment } from "../review-recipes.mjs";
import { reviewChild } from "../review-child.mjs";
import { acquireProcess } from "../acquire-process.mjs";
import { failureRecord, preparationFailure, failureFromRecord, ReviewFailure } from "../ui-review/failure.mjs";
import { repoRoot } from "../ui-review/policy.mjs";
import { resolveHarnessConfig } from "../../contract/harness-contract.mjs";
import { workspaceLayout } from "../../../workspace_layout.generated.mjs";

const scratch = () => mkdtempSync(path.join(os.tmpdir(), "cartulary-preparation-"));
function write(root, file, value) { mkdirSync(path.dirname(path.join(root, file)), { recursive: true }); writeFileSync(path.join(root, file), typeof value === "string" ? value : JSON.stringify(value)); }
const context = { phase: "build", subject_id: "server_harness", condition: "child_failed", recovery_id: "build_server_harness" };
const diagnosticModule = new URL("../../runtime/command-failure.mjs", import.meta.url).href;

test("browser producer callers retire journals only after native settlement and retain cancellation precedence", async t => {
  const root = scratch(), runID = path.basename(root);
  const runtime = createRuntime({ repoRoot, runRoot: root, runID });
  t.after(() => { runtime.close(); rmSync(root, { recursive: true, force: true }); });
  const environment = { ...process.env, CARTULARY_TEST_RESULTS_DIR: path.dirname(root), CARTULARY_TEST_RUN_ID: runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID };
  for (const kind of ["review", "acquisition"]) {
    const invoke = (command, args, options) => kind === "review"
      ? reviewChild({ root: repoRoot, command, args, environment, context, commandID: "cartulary.harness.command.build_server_harness.v1", ...options })
      : acquireProcess(command, args, { cwd: repoRoot, environment, ...options });
    let released = 0, reaped = 0, pid;
    const register = child => { pid = child; return () => { assert.throws(() => readFileSync(`/proc/${pid}/stat`), { code: "ENOENT" }); released++; }; };
    await invoke("sh", ["-c", "exit 0"], { onChildProcess: register, onReaped: () => { reaped++; } });
    assert.equal(released, 1);
    assert.equal(reaped, kind === "acquisition" ? 1 : 0);

    const abort = new AbortController(), reason = new Error("test cancellation");
    const childFile = path.join(root, `${kind}-descendant`);
    const pending = invoke("sh", ["-c", 'sleep 60 </dev/null >/dev/null 2>&1 & echo $! > "$1"; wait', "fixture", childFile], {
      signal: abort.signal, onChildProcess: register, onReaped: () => { reaped++; },
    });
    pending.catch(() => {});
    t.after(async () => { abort.abort(); await pending.catch(() => {}); });
    const deadline = Date.now() + 5000;
    while (!existsSync(childFile)) {
      if (Date.now() >= deadline) assert.fail("descendant fixture did not become ready");
      await Promise.race([new Promise(resolve => setTimeout(resolve, 10)), pending]);
    }
    const descendant = Number(readFileSync(childFile, "utf8").trim());
    abort.abort(reason);
    await assert.rejects(pending, error => error === reason && !error.cleanupFailures);
    assert.equal(released, 2);
    assert.equal(reaped, kind === "acquisition" ? 2 : 0);
    try { assert.match(readFileSync(`/proc/${descendant}/stat`, "utf8").split(") ").at(-1), /^[ZX] /u); }
    catch (error) { if (error.code !== "ENOENT") throw error; }

    const rejected = new Error("registration failed");
    await assert.rejects(invoke("sleep", ["60"], { onChildProcess: child => { pid = child; throw rejected; } }), error => kind === "review" ? error.cause === rejected : error === rejected);
    assert.throws(() => readFileSync(`/proc/${pid}/stat`), { code: "ENOENT" });
    assert.equal(released, 2, "failed registration invented a release callback");
    const preAborted = new AbortController(), preAbortReason = new Error("cancelled before acquisition");
    preAborted.abort(preAbortReason);
    let preAbortReaped = 0;
    await assert.rejects(invoke("sleep", ["60"], { signal: preAborted.signal,
      onChildProcess: () => assert.fail("pre-cancelled command launched"), onReaped: () => { preAbortReaped++; },
    }), error => error === preAbortReason);
    assert.equal(preAbortReaped, kind === "acquisition" ? 1 : 0, "pre-launch acquisition ticket retained without a process");
    await assert.rejects(invoke("/cartulary-absent-executable", [], { onChildProcess: () => assert.fail("missing command registered") }),
      error => kind === "review" ? error.failure_class === "config" && error.failure_reason === "configuration_error" : error.code === "ENOENT");
  }
});

test("browser producer callers preserve unresolved process registrations", async t => {
  const root = scratch(), runID = path.basename(root);
  const runtime = createRuntime({ repoRoot, runRoot: root, runID });
  t.after(() => { runtime.close(); rmSync(root, { recursive: true, force: true }); });
  const binary = path.join(root, "controller");
  const native = process.env.CARTULARY_HARNESS_RUNTIME_BIN || path.join(repoRoot, workspaceLayout.toolbin, "cartulary-harness");
  const ready = { event: "ready", proof: { boot: "11111111-2222-3333-4444-555555555555", pid: 100, start: "1", group: true, pgid: 100, sid: 2 } };
  const failure = { event: "failed", cause: "cleanup", diagnostic: { schema_id: "cartulary.harness_cleanup_diagnostic.v4", stage: "process_signal", cause: "io", errno: 13 } };
  writeFileSync(binary, `#!${process.execPath}\nimport {writeSync} from 'node:fs';import {spawnSync} from ${JSON.stringify(new URL("../../workspace/child-process.mjs", import.meta.url).href)};
if(process.argv[2]!=="owned-command"){const result=spawnSync(${JSON.stringify(native)},process.argv.slice(2),{stdio:"inherit"});process.exit(result.status??12);}
process.stdin.resume();writeSync(3,${JSON.stringify(JSON.stringify(ready) + "\n" + JSON.stringify(failure) + "\n")});process.exit(12);\n`, { mode: 0o700 });
  const environment = { ...process.env, CARTULARY_HARNESS_RUNTIME_BIN: binary, CARTULARY_TEST_RESULTS_DIR: path.dirname(root), CARTULARY_TEST_RUN_ID: runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID };
  let released = 0, reaped = 0;
  const onChildProcess = () => () => { released++; };
  for (const call of [
    () => reviewChild({ root: repoRoot, command: "unused", args: [], environment, context, onChildProcess }),
    () => acquireProcess("unused", [], { cwd: repoRoot, environment, onChildProcess, onReaped: () => { reaped++; } }),
  ]) await assert.rejects(call, error => error.failure_reason === "cleanup_error");
  assert.equal(released, 0);
  assert.equal(reaped, 0);
});


test("installation proof is independent of reuse caches and binds installed lock and workspace inputs", () => {
  const root = scratch();
  try {
    write(root, "tools/toolchain_pins.json", { node_version: process.versions.node, pnpm_version: "fixture" });
    for (const file of ["package.json", "apps/web/package.json", "packages/fixture/package.json"]) write(root, file, {});
    for (const file of ["pnpm-lock.yaml", "pnpm-workspace.yaml", "node_modules/.pnpm/lock.yaml"]) write(root, file, "fixture");
    assert.throws(() => validateFrontendInstallation(root), (error) => error.condition === "missing");
    publishFrontendInstallation(root);
    validateFrontendInstallation(root); // no optimization cache exists in this fixture
    for (const file of ["pnpm-lock.yaml", "pnpm-workspace.yaml", "packages/fixture/package.json", "node_modules/.pnpm/lock.yaml", ".npmrc"]) {
      write(root, file, "changed");
      assert.throws(() => validateFrontendInstallation(root), (error) => error.condition === "stale_installation");
      publishFrontendInstallation(root);
    }
    write(root, "tools/toolchain_pins.json", { node_version: "0.0.0", pnpm_version: "fixture" });
    assert.throws(() => nodeReadiness(root), (error) => error.subject_id === "node" && error.condition === "incompatible");
    assert.throws(() => validateFrontendInstallation(root), (error) => error.condition === "stale_installation");
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test("readiness rejects missing tools, damaged packages, browsers and images without acquisition", () => {
  const root = scratch();
  try {
    write(root, "tools/toolchain_pins.json", JSON.parse(readFileSync(path.join(repoRoot, "tools/toolchain_pins.json"))));
    assert.throws(() => coreReadiness(root), (error) => error.subject_id === "pnpm" && error.condition === "missing");
    for (const file of ["package.json", "apps/web/package.json", "packages/fixture/package.json"]) write(root, file, {});
    for (const file of ["pnpm-lock.yaml", "pnpm-workspace.yaml", "node_modules/.pnpm/lock.yaml"]) write(root, file, "fixture");
    publishFrontendInstallation(root);
    const pins = JSON.parse(readFileSync(path.join(root, "tools/toolchain_pins.json")));
    assert.throws(() => coreReadiness(root, { runner: () => ({ status: 0, stdout: pins.pnpm_version }) }), (error) => error.subject_id === "frontend_dependencies");
    assert.throws(() => browserReadiness(root), (error) => error.subject_id === "chromium");
    assert.throws(() => serviceImageReadiness({}, root, () => ({ status: 1 })), (error) => error.failure_reason === "preflight_error");
    const calls = [];
    assert.throws(() => serviceImageReadiness({}, root, (command, args) => {
      calls.push([command, ...args]);
      return { status: args[0] === "image" ? 1 : 0, stdout: "fixture-image" };
    }), (error) => error.subject_id === "test_service_images" && error.condition === "missing");
    assert.deepEqual(calls.map((call) => call.slice(1)), [["info", "--format", "{{.ServerVersion}}"], ["images"], ["image", "inspect", "fixture-image"]]);
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test("readiness checks the adapter's axe engine without a redundant root dependency", () => {
  const root = scratch();
  try {
    const pins = JSON.parse(readFileSync(path.join(repoRoot, "tools/toolchain_pins.json")));
    write(root, "tools/toolchain_pins.json", pins);
    const devDependencies = Object.fromEntries(Object.entries(pins.ui_review).filter(([name]) => name !== "axe-core"));
    write(root, "package.json", { devDependencies });
    write(root, "apps/web/package.json", { devDependencies: { "@playwright/test": pins.ui_review.playwright } });
    write(root, "packages/fixture/package.json", {});
    for (const file of ["pnpm-lock.yaml", "pnpm-workspace.yaml", "node_modules/.pnpm/lock.yaml"]) write(root, file, "fixture");
    for (const [name, version] of Object.entries({ ...devDependencies, "@playwright/test": pins.ui_review.playwright })) {
      write(root, `node_modules/${name}/package.json`, { name, version, main: "index.cjs" });
      write(root, `node_modules/${name}/index.cjs`, name === "sharp" ? `module.exports = { versions: { sharp: '${version}' } };` : "module.exports = {};");
    }
    write(root, "node_modules/playwright-core/browsers.json", { browsers: [{ name: "chromium", revision: "fixture", browserVersion: "fixture" }] });
    const engine = "node_modules/@axe-core/playwright/node_modules/axe-core/package.json";
    write(root, engine, { name: "axe-core", version: pins.ui_review["axe-core"] });
    publishFrontendInstallation(root);
    const options = { runner: () => ({ status: 0, stdout: pins.pnpm_version }) };
    assert.equal(coreReadiness(root, options).axe_version, pins.ui_review["axe-core"]);
    assert.equal(existsSync(path.join(root, "node_modules/axe-core")), false);
    write(root, engine, { name: "axe-core", version: "0.0.0" });
    assert.throws(() => coreReadiness(root, options), (error) => error.subject_id === "frontend_dependencies" && error.condition === "incompatible");
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test("Make child envelopes preserve classification, reject corruption and never disclose error text", async () => {
  const root = scratch(), runID = path.basename(root);
  const runtime = createRuntime({ repoRoot, runRoot: root, runID });
  const environment = { ...process.env, CARTULARY_TEST_RESULTS_DIR: path.dirname(root), CARTULARY_TEST_RUN_ID: path.basename(root), CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID };
  try {
    for (const [failure_class, failure_reason] of [["config", "configuration_error"], ["infra", "preflight_error"], ["infra", "service_start_error"], ["infra", "service_readiness_timeout"], ["harness", "fixture_error"], ["harness", "tool_diagnostic_failure"], ["artifact", "artifact_error"], ["infra", "resource_conflict"], ["unknown", "unknown_failure"]]) {
      write(root, "child.mjs", `import {publishCommandFailure} from ${JSON.stringify(diagnosticModule)}; publishCommandFailure(${JSON.stringify(repoRoot)},${JSON.stringify({ failure_class, failure_reason })}); console.error("secret-sentinel /private/path"); process.exit(1);`);
      write(root, "Makefile", `all:\n\t@${process.execPath} ${path.join(root, "child.mjs")}\n`);
      let caught;
      try { await reviewChild({ root: repoRoot, command: "make", args: ["-f", path.join(root, "Makefile")], environment, commandID: "cartulary.harness.command.build_server_harness.v1", context, onChildProcess: () => () => { throw new Error("cleanup secret-sentinel"); } }); }
      catch (error) { caught = error; }
      const record = failureRecord(preparationFailure(caught));
      assert.equal(record.failure_class, failure_class); assert.equal(record.failure_reason, failure_reason);
      assert.equal(record.subject_id, "server_harness"); assert.equal(caught.cleanupFailures.length, 1);
      assert.ok(!JSON.stringify(record).includes("sentinel")); assert.ok(!JSON.stringify(record).includes("/private"));
      failureFromRecord(record);
    }
    for (const mode of ["malformed", "cross-invocation", "conflict", "success-plus-failure", "death"]) {
      write(root, "child.mjs", `import {publishCommandFailure} from ${JSON.stringify(diagnosticModule)}; import {writeFileSync, readdirSync} from 'node:fs'; import path from 'node:path';
        const mode=${JSON.stringify(mode)}, root=${JSON.stringify(path.join(runtime.root,"detail"))};
        if(mode==='death') process.exit(1);
        publishCommandFailure(${JSON.stringify(repoRoot)}, {failure_class:'config',failure_reason:'configuration_error'});
        const directory=path.join(root,readdirSync(root).find(name=>name.startsWith('command-failure-')));
        if(mode==='malformed') writeFileSync(path.join(directory,'failure.json'),'{');
        if(mode==='cross-invocation') { const {readFileSync}=await import('node:fs'); const file=path.join(directory,'failure.json'); const record=JSON.parse(readFileSync(file));record.invocation_id='00000000-0000-4000-8000-000000000000';writeFileSync(file,JSON.stringify(record)); }
        if(mode==='conflict') writeFileSync(path.join(directory,'conflict'),'');
        process.exit(mode==='success-plus-failure'?0:1);`);
      await assert.rejects(reviewChild({ root: repoRoot, command: process.execPath, args: [path.join(root, "child.mjs")], environment, commandID: "cartulary.harness.command.build_server_harness.v1", context }), (error) => error.failure_reason === (mode === "death" ? "unknown_failure" : "scheduler_accounting_error"));
    }
  } finally { runtime.close(); rmSync(root, { recursive: true, force: true }); }
});



test("Go readiness validates only review producer dependencies with network and automatic toolchain acquisition disabled", () => {
  const calls = [];
  const runner = (command, args, options) => { calls.push({ command, args, env: options.env }); return { status: 0, stdout: calls.length === 1 ? "/installed/go" : "" }; };
  assert.equal(goReadiness({ GO_CACHE_DIR: "/cache/build", GO_MOD_CACHE_DIR: "/cache/mod", GO_TMP_DIR: "/cache/tmp" }, repoRoot, runner), "/installed/go");
  assert.ok(calls[1].args.includes("./cmd/server") && !calls[1].args.includes("all"));
  for (const [key, value] of Object.entries({ GOTOOLCHAIN: "local", GOPROXY: "off", GONOPROXY: "none", GOSUMDB: "off" })) assert.equal(calls[1].env[key], value);
  let invocation = 0;
  assert.throws(() => goReadiness({}, repoRoot, () => ({ status: 0, stdout: ++invocation === 1 ? "/installed/go" : "MISSING" })), (error) => error.subject_id === "go" && error.condition === "missing");
});


test("review preparation resolves the selected executable without an inherited Go root", () => {
  for (const preparation_policy of ["installed_only", "ensure"]) {
    const environment = recipeEnvironment(repoRoot, "/review/run", { runID: "fixture", root: "/private/runtime", leaseID: "fixture" },
      { preparation_policy, presentation: "design" }, { ...process.env, GO: "/selected/go", GOROOT: "/foreign/toolchain" });
    assert.equal(environment.GO, "/selected/go");
    assert.equal(Object.hasOwn(environment, "GOROOT"), false);
  }
});

test("review recipes isolate a machine parent without weakening public output admission", () => {
  const parent = { ...process.env, CARTULARY_OUTPUT_MODE: "machine" };
  for (const preparation_policy of ["installed_only", "ensure"]) {
    for (const presentation of ["design", "review"]) {
      const environment = recipeEnvironment(repoRoot, "/review/run", { runID: "fixture", root: "/private/runtime", leaseID: "fixture" },
        { preparation_policy, presentation }, parent);
      assert.equal(Object.hasOwn(environment, "CARTULARY_OUTPUT_MODE"), false);
      assert.doesNotThrow(() => resolveHarnessConfig(environment.CARTULARY_TEST_TARGET, environment));
      assert.throws(() => resolveHarnessConfig(environment.CARTULARY_TEST_TARGET, { ...environment, CARTULARY_OUTPUT_MODE: "machine" }),
        /does not accept CARTULARY_OUTPUT_MODE=machine/u);
    }
  }
  assert.equal(parent.CARTULARY_OUTPUT_MODE, "machine");
});
