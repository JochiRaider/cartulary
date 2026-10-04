import "./test-browser-acquisition.mjs";
import assert from "node:assert/strict";
import { existsSync, mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";
import test from "node:test";
import { browserReadiness, coreReadiness, goReadiness, nodeReadiness, publishFrontendInstallation, serviceImageReadiness, validateFrontendInstallation } from "../../readiness/installed-readiness.mjs";
import { createSuiteRuntime } from "../../runtime/suite-runtime.mjs";
import { reviewChild } from "../review-child.mjs";
import { withReviewResources } from "../review-preparation.mjs";
import { prepareReview, cleanupRecords } from "../ui-review/preparation.mjs";
import { failureRecord, preparationFailure, failureFromRecord, ReviewFailure } from "../ui-review/failure.mjs";
import { repoRoot } from "../ui-review/policy.mjs";
import { ReviewSession } from "../ui-review/session.mjs";
import { toolProfile } from "../ui-review/toolchain.mjs";
import { emitResult } from "../ui-review/output.mjs";

const scratch = () => mkdtempSync(path.join(os.tmpdir(), "cartulary-preparation-"));
function write(root, file, value) { mkdirSync(path.dirname(path.join(root, file)), { recursive: true }); writeFileSync(path.join(root, file), typeof value === "string" ? value : JSON.stringify(value)); }
const context = { phase: "build", subject_id: "server_harness", condition: "child_failed", recovery_id: "build_server_harness" };
const diagnosticModule = new URL("../../runtime/command-failure.mjs", import.meta.url).href;

test("shared preparation cleanup alone has cleanup classification and preserves an earlier operational failure", async () => {
  for (const operational of [false, true]) {
    const events = [], primary = preparationFailure({ failure_class: "harness", failure_reason: "fixture_error" }, { phase: "seeding", subject_id: "fixture_seed", condition: "child_failed", recovery_id: "inspect_failure" });
    let caught;
    try {
      await withReviewResources({ acquire: async () => {}, prepare: async () => { if (operational) throw primary; }, hold: async () => {},
        close: async () => { events.push("close"); throw new Error("private cleanup sentinel"); }, finish: async () => { events.push("finish"); } });
    } catch (error) { caught = error; }
    assert.deepEqual(events, ["close", "finish"]); assert.equal(caught.cleanupFailures.length, 1);
    if (operational) assert.equal(caught, primary);
    const projected = preparationFailure(caught), record = failureRecord(projected);
    assert.equal(projected.exitCode, operational ? 3 : 12);
    assert.equal(record.phase, operational ? "seeding" : "cleanup");
    assert.equal(record.failure_reason, operational ? "fixture_error" : "cleanup_error");
    assert.ok(!JSON.stringify(record).includes("sentinel"));
  }
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

test("Make child envelopes preserve classification, reject corruption and never disclose error text", async () => {
  const root = scratch(), runID = path.basename(root);
  const runtime = createSuiteRuntime({ repoRoot, runRoot: root, runID });
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

test("preparation IPC retains validated context and latches malformed outcomes", async () => {
  const root = scratch();
  const runtime = createSuiteRuntime({ repoRoot, runRoot: root, runID: "ipc" });
  try {
    for (const mode of ["classified", "malformed", "death"]) {
      const record = failureRecord(preparationFailure({ failure_class: "harness", failure_reason: "fixture_error" }, { phase: "seeding", subject_id: "fixture_seed", condition: "child_failed", recovery_id: "inspect_failure" }));
      const worker = path.join(root, `${mode}.mjs`);
      write(root, `${mode}.mjs`, `process.once('message',()=>{ process.send({phase:${JSON.stringify(context)}}); ${mode === "death" ? "process.exit(1);" : `process.send({done:true,failure:${JSON.stringify(mode === "malformed" ? { ...record, private_message: "secret-sentinel" } : record)},cleanup_failures:[]},()=>process.disconnect());`} });`);
      const preparation = prepareReview({ input: {}, runtime, runID: mode, signal: new AbortController().signal, workerModule: pathToFileURL(worker).href });
      await assert.rejects(preparation.ready, (error) => {
        assert.equal(error.diagnostic, mode === "classified" ? "fixture_failed" : mode === "malformed" ? "diagnostic_invalid" : "preparation_failed");
        assert.equal(error.context.phase, mode === "classified" ? "seeding" : "build");
        return true;
      });
      await assert.rejects(preparation.done);
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

test("cancellation in each preparation phase rejects late ready publication and drains the IPC owner", async () => {
  const root = scratch(), runtime = createSuiteRuntime({ repoRoot, runRoot: root, runID: "cancel" });
  try {
    for (const phase of ["prerequisites", "build", "service_acquisition", "service_readiness", "seeding", "browser_start"]) {
      const abort = new AbortController(), worker = path.join(root, `${phase}.mjs`);
      write(root, `${phase}.mjs`, `process.once('message',()=>{process.send({phase:${JSON.stringify({ ...context, phase })}});process.on('message',message=>{if(message.stop) {process.send({ready:{}});process.send({done:true,failure:null,cleanup_failures:[]},()=>process.disconnect());}});});`);
      const preparation = prepareReview({ input: {}, runtime, runID: phase, signal: abort.signal, workerModule: pathToFileURL(worker).href });
      const rejection = assert.rejects(preparation.ready, (error) => error.diagnostic === "interrupted");
      await new Promise((resolve) => setTimeout(resolve, 100)); abort.abort(new ReviewFailure("interrupted"));
      await preparation.stop(); await rejection;
    }
  } finally { runtime.close(); rmSync(root, { recursive: true, force: true }); }
});


test("IPC rejects duplicate terminal outcomes and success followed by child failure", async () => {
  const root = scratch(), runtime = createSuiteRuntime({ repoRoot, runRoot: root, runID: "contradiction" });
  try {
    for (const mode of ["duplicate", "nonzero"]) {
      const worker = path.join(root, `${mode}.mjs`);
      write(root, `${mode}.mjs`, `process.once('message',()=>{process.send({phase:${JSON.stringify(context)}});const done={done:true,failure:null,cleanup_failures:[]};process.send(done,()=>{${mode === "duplicate" ? "process.send(done,()=>process.disconnect());" : "process.exit(1);"}});});`);
      const preparation = prepareReview({ input: {}, runtime, runID: mode, signal: new AbortController().signal, workerModule: pathToFileURL(worker).href });
      await assert.rejects(preparation.ready, (error) => error.diagnostic === "diagnostic_invalid");
      await assert.rejects(preparation.done, (error) => error.diagnostic === "diagnostic_invalid");
    }
  } finally { runtime.close(); rmSync(root, { recursive: true, force: true }); }
});

test("preparation cleanup-only IPC outcome retains cleanup context after cancellation", async () => {
  const root = scratch(), runtime = createSuiteRuntime({ repoRoot, runRoot: root, runID: "cleanup-context" }), abort = new AbortController();
  try {
    const worker = path.join(root, "worker.mjs");
    write(root, "worker.mjs", `process.once('message',()=>{process.send({phase:${JSON.stringify(context)}});process.send({ready:{}});process.once('message',()=>process.send({done:true,failure:null,cleanup_failures:${JSON.stringify([failureRecord(new ReviewFailure("cleanup_failed", { context: { ...context, phase: "cleanup", recovery_id: "exact_stop" } }))])}},()=>process.disconnect()));});`);
    const preparation = prepareReview({ input: {}, runtime, runID: "cleanup-context", signal: abort.signal, workerModule: pathToFileURL(worker).href });
    await preparation.ready; abort.abort(new ReviewFailure("interrupted"));
    await assert.rejects(preparation.stop(), (error) => error.exitCode === 12 && error.context.phase === "cleanup" && error.context.recovery_id === "exact_stop");
  } finally { runtime.close(); rmSync(root, { recursive: true, force: true }); }
});

test("a failing Make producer crosses preparation IPC into the terminal receipt and safe human guidance", async () => {
  const root = scratch(), session = new ReviewSession({ UI_MODE: "artifacts" }, toolProfile());
  try {
    session.initialize({ CARTULARY_TEST_RESULTS_DIR: root, CARTULARY_TEST_RUN_ID: "producer-failure" });
    const privateRoot = session.runtime.privatePath("fault-producer");
    write(privateRoot, "child.mjs", `import {publishCommandFailure} from ${JSON.stringify(diagnosticModule)}; console.error("secret-sentinel /private/path"); publishCommandFailure(${JSON.stringify(repoRoot)}, {failure_class:'harness',failure_reason:'tool_diagnostic_failure'}); process.exit(7);`);
    write(privateRoot, "Makefile", `all:\n\t@${process.execPath} ${path.join(privateRoot, "child.mjs")}\n`);
    const worker = path.join(privateRoot, "worker.mjs");
    write(privateRoot, "worker.mjs", `
      import {reviewChild} from ${JSON.stringify(new URL("../review-child.mjs", import.meta.url).href)};
      import {failureRecord,preparationFailure} from ${JSON.stringify(new URL("../ui-review/failure.mjs", import.meta.url).href)};
      process.once('message',async ({runID,runtime})=>{
        const context=${JSON.stringify(context)};
        process.send({phase:context});
        let failure=null;
        try { await reviewChild({root:${JSON.stringify(repoRoot)}, command:'make',args:['-f',${JSON.stringify(path.join(privateRoot, "Makefile"))}],
          environment:{...process.env,CARTULARY_TEST_RESULTS_DIR:${JSON.stringify(root)},CARTULARY_TEST_RUN_ID:runID,CARTULARY_HARNESS_SUITE_RUNTIME_ROOT:runtime.root,CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID:runtime.leaseID,CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID:runID},
          commandID:'cartulary.harness.command.build_server_harness.v1',context}); }
        catch(error) { failure=failureRecord(preparationFailure(error,context)); }
        process.send({done:true,failure,cleanup_failures:[]},()=>process.disconnect());
      });`);
    session.prepareMode = async () => {
      session.preparedOwner = prepareReview({ input: session.input, runtime: session.runtime, runID: session.runID, signal: session.abort.signal, workerModule: pathToFileURL(worker).href });
      session.preparation = session.preparedOwner.done;
      await session.preparedOwner.ready;
    };
    let primary;
    try { await session.prepare(); } catch (error) { primary = error; }
    assert.equal(primary?.diagnostic, "build_failed");
    await session.stop(primary);
    const value = session.terminalValue("ui-review-stop");
    assert.equal(value.exit_code, 1); assert.equal(value.failures.length, 1);
    assert.deepEqual(value.failures[0], { failure_class: "harness", failure_reason: "tool_diagnostic_failure", diagnostic_code: "build_failed", ...context });
    const file = path.join(session.runRoot, value.receipt.path), receipt = readFileSync(file, "utf8");
    assert.equal(JSON.parse(receipt).cleanup, "complete"); assert.deepEqual(JSON.parse(receipt).failures, value.failures);
    let human = ""; emitResult(value, "summary", { write: (text) => { human += text; } });
    assert.match(human, /phase=build subject=server_harness condition=child_failed; inspect make build-server-harness/u);
    assert.ok(!`${human}${receipt}`.includes("sentinel") && !`${human}${receipt}`.includes("/private/path"));
    assert.equal(existsSync(session.runtime.root), false);
    await session.stop(); assert.equal(readFileSync(file, "utf8"), receipt);
  } finally { await session.stop(); rmSync(root, { recursive: true, force: true }); }
});


test("preparation bounds secondary diagnostics without replacing the primary outcome", () => {
  const records = cleanupRecords(Array.from({ length: 80 }, () => new Error("private cleanup detail")), { subject_id: "preparation_child" });
  assert.equal(records.length, 64);
  assert.equal(records.at(-1).diagnostic_code, "diagnostic_invalid");
  assert.ok(records.every((record) => !JSON.stringify(record).includes("private cleanup detail")));
});
