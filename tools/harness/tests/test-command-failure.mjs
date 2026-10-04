import assert from "node:assert/strict";
import test from "node:test";
import { lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync, symlinkSync, chmodSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createSuiteRuntime } from "../runtime/suite-runtime.mjs";
import { createCommandFailureContext, publishCommandFailure, readCommandFailure } from "../runtime/command-failure.mjs";
import { executeUnitProcess } from "../scheduler/work-graph/executor.mjs";
import { adaptShellInvocation } from "../execution/runners/shell.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const fixture = path.join(root, "tools/harness/tests/command-failure-fixture.mjs");
const artifactFailure = { failure_class: "artifact", failure_reason: "artifact_error" };
const accounting = { failure_class: "harness", failure_reason: "scheduler_accounting_error" };
function setup(t) {
  const scratch = mkdtempSync(path.join(os.tmpdir(), "command-failure-test-"));
  const runRoot = path.join(scratch, "run"); mkdirSync(runRoot, { mode: 0o700 });
  const runtime = createSuiteRuntime({ repoRoot: root, runRoot, runID: "run", scratchRoot: path.join(scratch, "private") });
  const environment = { CARTULARY_TEST_RESULTS_DIR: scratch, CARTULARY_TEST_RUN_ID: "run", CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "run" };
  const abort = new AbortController(); const pending = new Set(); const timerAdapters = new Set(); const contexts = [];
  t.after(async () => {
    abort.abort();
    for (const timers of timerAdapters) timers.expireAll();
    await Promise.all(pending);
    for (const context of contexts) context.close();
    runtime.close(); rmSync(scratch, { recursive: true, force: true });
  });
  const execute = (unit, options = {}) => {
    const start = performance.timeOrigin + performance.now();
    if (options.timers) timerAdapters.add(options.timers);
    const result = executeUnitProcess(unit, { cwd: root, environment, ...options, signal: options.signal ? AbortSignal.any([abort.signal, options.signal]) : abort.signal });
    pending.add(result);
    return result.then((value) => {
      const ended = performance.timeOrigin + performance.now();
      const phases = new Map(value.stdout.split("\n").filter((line) => line.startsWith('{"phase":')).map((line) => { const phase = JSON.parse(line); return [phase.phase, phase.at]; }));
      if (phases.has("published")) t.diagnostic(JSON.stringify({ launch_ms: phases.get("started") - start, publication_ms: phases.get("published") - phases.get("started"), execution_ms: phases.get("completed") - phases.get("published"), settlement_ms: ended - phases.get("completed") }));
      pending.delete(result); t.diagnostic(`child ${unit.unit_id}: ${Math.round(performance.timeOrigin + performance.now() - start)}ms including launch, execution and settlement`); return value; });
  };
  const create = () => {
    const context = createCommandFailureContext({ repoRoot: root, environment, unitID: "target:build-web", commandID: "cartulary.harness.command.build_web.v2" });
    contexts.push(context); return context;
  };
  return { scratch, runRoot, runtime, environment, execute, create };
}
function manualTimers() {
  const active = new Map();
  return {
    setTimeout(callback, ms) { const handle = {}; active.set(handle, { callback, ms }); return handle; },
    clearTimeout(handle) { active.delete(handle); },
    fire(ms) { const found = [...active].find(([, timer]) => timer.ms === ms); assert.ok(found, `timer ${ms} armed`); active.delete(found[0]); found[1].callback(); },
    count() { return active.size; },
    expireAll() { while (active.size) { const [handle, { callback }] = active.entries().next().value; active.delete(handle); callback(); } },
  };
}
async function ready(file) {
  const deadline = Date.now() + 10000;
  while (!lstatSync(file, { throwIfNoEntry: false })) {
    assert.ok(Date.now() < deadline, "child installed signal handlers within watchdog");
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  return Number(readFileSync(file, "utf8"));
}
export function registerCommandFailureTests() {
  test("command failure envelopes bind identity and reject corruption", (t) => {
    const { scratch, runtime, environment, create } = setup(t);
  for (const mutation of ["valid", "identity", "malformed", "conflict", "permission", "symlink", "partial"]) {
    const context = create();
    const env = { ...environment, ...context.environment };
    assert.equal(context.read(), null);
    publishCommandFailure(root, artifactFailure, env);
    const identity = JSON.parse(env.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT);
    const file = runtime.privatePath(`command-failure-${identity.invocation_id}`, "failure.json");
    if (mutation === "identity") writeFileSync(file, JSON.stringify({ ...JSON.parse(readFileSync(file)), unit_id: "another-unit" }));
    if (mutation === "malformed") writeFileSync(file, '{"schema_id":');
    if (mutation === "conflict") assert.throws(() => publishCommandFailure(root, accounting, env), /conflicting/u);
    if (mutation === "permission") chmodSync(file, 0o644);
    if (mutation === "symlink") { rmSync(file); symlinkSync(path.join(scratch, "absent"), file); }
    if (mutation === "partial") rmSync(file);
    assert.deepEqual(context.read(), mutation === "valid" ? artifactFailure : accounting, mutation);
    assert.deepEqual(readCommandFailure(root, env), context.read());
    if (mutation === "valid") {
      publishCommandFailure(root, artifactFailure, env);
      assert.deepEqual(context.read(), artifactFailure, "identical diagnostic is idempotent");
      assert.equal(lstatSync(file).mode & 0o777, 0o600);
    }
    context.close();
  }
  });
  test("shell classification preserves diagnostics and accounts for contradictory success", () => {
  const invocation = { rows: [{ row_id: "example" }] };
  assert.equal(adaptShellInvocation(invocation, { status: 2, commandFailure: artifactFailure })[0].failure_reason, "artifact_error");
  assert.equal(adaptShellInvocation(invocation, { status: 0, commandFailure: artifactFailure })[0].failure_reason, "scheduler_accounting_error");
  assert.equal(adaptShellInvocation(invocation, { status: 1 })[0].failure_reason, "test_assertion_failure");

  });
  for (const [mode, expected, exit] of [["diagnostic", "artifact_error", 11], ["contradictory", "scheduler_accounting_error", 11], ["assertion", "test_assertion_failure", 10]]) {
    test(`command child ${mode} propagates through Make and executor`, { timeout: 15000 }, async (t) => {
      const { scratch, execute } = setup(t);
      const makefile = path.join(scratch, "Makefile");
      writeFileSync(makefile, `diagnostic contradictory assertion:\n\t@${process.execPath} ${fixture} $@\n`);
      const outcome = await execute({ unit_id: "target:build-web", kind: "artifact", command: { executable: mode === "assertion" ? process.execPath : "make", args: mode === "assertion" ? [fixture, mode] : ["--silent", "-f", makefile, mode], environment: { CARTULARY_TEST_TARGET: "build-web" } }, timeout_ms: 10000 });
      assert.equal(outcome.failure_reason, expected, JSON.stringify(outcome)); assert.equal(outcome.exit_code, exit);
    });
  }
  test("artifact diagnostic survives canonical row execution and removes private captures", { timeout: 15000 }, async (t) => {
    const { scratch, runtime, runRoot, execute } = setup(t);
  const shim = path.join(scratch, "make-probe");
  writeFileSync(shim, `#!/bin/sh\ncase "$*" in\n  *protocol-ts-browser-artifact-reachability*) exec ${process.execPath} ${fixture} diagnostic ;;\n  *) exec env MAKE=make make "$@" ;;\nesac\n`, { mode: 0o700 });
  const rowID = "package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b";
  const rowOutcome = await execute({ unit_id: `row:${rowID}`, kind: "runner", command: { executable: process.execPath, args: ["tools/harness/execution/runners/row-runner-cli.mjs", "--row-id", rowID], environment: { CARTULARY_TEST_TARGET: "protocol-ts-browser-artifact-reachability", MAKE: shim } }, timeout_ms: 10000 });
  assert.equal(rowOutcome.failure_reason, "artifact_error", JSON.stringify(rowOutcome));
  assert.ok(lstatSync(path.join(runRoot, "rows", `${rowID}.json`), { throwIfNoEntry: false }), JSON.stringify(rowOutcome));
  assert.equal(JSON.parse(readFileSync(path.join(runRoot, "rows", `${rowID}.json`))).failure_reason, "artifact_error", "canonical shell row preserves the cause before the executor writes its result");
  assert.deepEqual(readdirSync(runtime.privatePath("child-captures")), [], "failed row execution releases its private captures");
  assert.equal(readdirSync(runtime.root).some((name) => name.startsWith("command-failure-")), false, "row execution releases command failure contexts");
  });
  for (const mode of ["wait", "cooperative", "cancelled"]) {
    test(`executor ${mode} after handler readiness preserves deadline and cancellation precedence`, { timeout: 15000 }, async (t) => {
      const { scratch, execute } = setup(t); const timers = manualTimers();
      const readyFile = path.join(scratch, "ready"); const controller = new AbortController();
      t.after(() => controller.abort());
      const result = execute({ unit_id: "target:build-web", kind: "artifact", command: { executable: process.execPath, args: [fixture, mode === "cancelled" ? "cooperative" : mode, readyFile], environment: { CARTULARY_TEST_TARGET: "build-web" } }, timeout_ms: 500 }, { timers, signal: controller.signal });
      const pid = await ready(readyFile);
      if (mode === "cancelled") controller.abort(); else timers.fire(500);
      if (mode === "wait") timers.fire(2000);
      const outcome = await result;
      assert.equal(outcome.failure_reason, mode === "cancelled" ? "cancelled_or_interrupted" : "timeout_failure", JSON.stringify(outcome));
      if (mode !== "cancelled") assert.equal(outcome.exit_code, 13);
      if (mode === "wait") assert.equal(outcome.signal, "SIGKILL");
      assert.throws(() => process.kill(pid, 0), { code: "ESRCH" }, "child is reaped before completion");
      assert.equal(timers.count(), 0, "all deadline callbacks removed");
    });
  }
}
