import "./test-browser-acquisition.mjs";
import assert from "node:assert/strict";
import { existsSync, mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { browserReadiness, coreReadiness, goReadiness, nodeReadiness, publishFrontendInstallation, serviceImageReadiness, validateFrontendInstallation } from "../../readiness/installed-readiness.mjs";
import { createRuntime } from "../../runtime/session-client.mjs";
import { reviewChild } from "../review-child.mjs";
import { failureRecord, preparationFailure, failureFromRecord, ReviewFailure } from "../ui-review/failure.mjs";
import { repoRoot } from "../ui-review/policy.mjs";

const scratch = () => mkdtempSync(path.join(os.tmpdir(), "cartulary-preparation-"));
function write(root, file, value) { mkdirSync(path.dirname(path.join(root, file)), { recursive: true }); writeFileSync(path.join(root, file), typeof value === "string" ? value : JSON.stringify(value)); }
const context = { phase: "build", subject_id: "server_harness", condition: "child_failed", recovery_id: "build_server_harness" };
const diagnosticModule = new URL("../../runtime/command-failure.mjs", import.meta.url).href;


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
        const mode=${JSON.stringify(mode)}, root=${JSON.stringify(runtime.root)};
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
