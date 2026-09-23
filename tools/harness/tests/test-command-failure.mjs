import assert from "node:assert/strict";
import { lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync, symlinkSync, chmodSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createSuiteRuntime } from "../runtime/suite-runtime.mjs";
import { createCommandFailureContext, publishCommandFailure, readCommandFailure } from "../runtime/command-failure.mjs";
import { executeUnitProcess } from "../scheduler/work-graph/executor.mjs";
import { adaptShellInvocation } from "../execution/runners/shell.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const scratch = mkdtempSync(path.join(os.tmpdir(), "command-failure-test-"));
const runRoot = path.join(scratch, "run");
mkdirSync(runRoot, { mode: 0o700 });
const runtime = createSuiteRuntime({ repoRoot: root, runRoot, runID: "run", scratchRoot: path.join(scratch, "private") });
const environment = { CARTULARY_TEST_RESULTS_DIR: scratch, CARTULARY_TEST_RUN_ID: "run", CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: "run" };
const artifactFailure = { failure_class: "artifact", failure_reason: "artifact_error" };
const accounting = { failure_class: "harness", failure_reason: "scheduler_accounting_error" };
const create = () => createCommandFailureContext({ repoRoot: root, environment, unitID: "target:build-web", commandID: "cartulary.harness.command.build_web.v2" });
try {
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
  const invocation = { rows: [{ row_id: "example" }] };
  assert.equal(adaptShellInvocation(invocation, { status: 2, commandFailure: artifactFailure })[0].failure_reason, "artifact_error");
  assert.equal(adaptShellInvocation(invocation, { status: 0, commandFailure: artifactFailure })[0].failure_reason, "scheduler_accounting_error");
  assert.equal(adaptShellInvocation(invocation, { status: 1 })[0].failure_reason, "test_assertion_failure");

  const fixture = path.join(root, "tools/harness/tests/frontend-producer-fixture.mjs");
  const makefile = path.join(scratch, "Makefile");
  writeFileSync(makefile, `diagnostic contradictory assertion:\n\t@${process.execPath} ${fixture} $@\n`);
  for (const [mode, expected, exit] of [["diagnostic", "artifact_error", 11], ["contradictory", "scheduler_accounting_error", 11], ["assertion", "test_assertion_failure", 10]]) {
    const outcome = await executeUnitProcess({ unit_id: "target:build-web", kind: "artifact", command: { executable: mode === "assertion" ? process.execPath : "make", args: mode === "assertion" ? [fixture, mode] : ["--silent", "-f", makefile, mode], environment: { CARTULARY_TEST_TARGET: "build-web" } }, timeout_ms: 10000 }, { cwd: root, environment });
    assert.equal(outcome.failure_reason, expected, JSON.stringify(outcome));
    assert.equal(outcome.exit_code, exit);
  }
  const shim = path.join(scratch, "make-probe");
  writeFileSync(shim, `#!/bin/sh\ncase "$*" in\n  *protocol-ts-browser-artifact-reachability*) exec ${process.execPath} ${fixture} diagnostic ;;\n  *) exec env MAKE=make make "$@" ;;\nesac\n`, { mode: 0o700 });
  const rowID = "package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b";
  const rowOutcome = await executeUnitProcess({ unit_id: `row:${rowID}`, kind: "runner", command: { executable: process.execPath, args: ["tools/harness/execution/runners/row-runner-cli.mjs", "--row-id", rowID], environment: { CARTULARY_TEST_TARGET: "protocol-ts-browser-artifact-reachability", MAKE: shim } }, timeout_ms: 10000 }, { cwd: root, environment });
  assert.equal(rowOutcome.failure_reason, "artifact_error", JSON.stringify(rowOutcome));
  assert.ok(lstatSync(path.join(runRoot, "rows", `${rowID}.json`), { throwIfNoEntry: false }), JSON.stringify(rowOutcome));
  assert.equal(JSON.parse(readFileSync(path.join(runRoot, "rows", `${rowID}.json`))).failure_reason, "artifact_error", "canonical shell row preserves the cause before the executor writes its result");
  assert.deepEqual(readdirSync(runtime.privatePath("child-captures")), [], "failed row execution releases its private captures");
  assert.equal(readdirSync(runtime.root).some((name) => name.startsWith("command-failure-")), false, "row execution releases command failure contexts");
  const waiting = { unit_id: "target:build-web", kind: "artifact", command: { executable: process.execPath, args: [fixture, "wait"], environment: { CARTULARY_TEST_TARGET: "build-web" } }, timeout_ms: 500 };
  const timedOut = await executeUnitProcess(waiting, { cwd: root, environment });
  assert.equal(timedOut.failure_reason, "timeout_failure");
  assert.equal(timedOut.signal, "SIGKILL", "unresponsive child is bounded by the existing process group boundary");
  const cooperative = await executeUnitProcess({ ...waiting, command: { ...waiting.command, args: [fixture, "cooperative"] } }, { cwd: root, environment });
  assert.equal(cooperative.status, "failed");
  assert.equal(cooperative.failure_reason, "timeout_failure", "cooperative zero exit cannot turn a timeout into success");
  assert.equal(cooperative.exit_code, 13);
  const cancellation = new AbortController();
  cancellation.abort();
  const cancelled = await executeUnitProcess(waiting, { cwd: root, environment, signal: cancellation.signal });
  assert.equal(cancelled.failure_reason, "cancelled_or_interrupted");
} finally { runtime.close(); rmSync(scratch, { recursive: true, force: true }); }
