import assert from "node:assert/strict";
import test from "node:test";
import { lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync, symlinkSync, chmodSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { createSuiteRuntime } from "../runtime/suite-runtime.mjs";
import { createCommandFailureContext, publishCommandFailure, readCommandFailure } from "../runtime/command-failure.mjs";
import { executeUnitProcess } from "../scheduler/work-graph/executor.mjs";
import { WorkGraphCompiler, buildWorkGraph, runWorkGraph } from "../scheduler/work-graph/index.mjs";
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
  test("Govulncheck causes survive Make and aggregate scheduling without cache reuse", { timeout: 60000 }, async (t) => {
    const { scratch, runtime, environment, execute } = setup(t);
    const compiler = new WorkGraphCompiler(root);
    const canonical = compiler.compile({ kind: "target", target: "go-vulncheck" }).units.find((unit) => unit.unit_id === "target:go-vulncheck");
    assert.deepEqual(compiler.compile({ kind: "aggregate", target: "check" }).units.find((unit) => unit.unit_id === canonical.unit_id), canonical, "direct and aggregate selection use the same scanner unit");
    const fakeGo = path.join(scratch, "go");
    const scanner = path.join(scratch, "govulncheck");
    const payload = path.join(scratch, "scanner-output");
    const invoked = path.join(scratch, "scanner-invoked");
    const makefile = path.join(scratch, "Makefile");
    writeFileSync(fakeGo, '#!/bin/sh\nprintf "%s\\n" github.com/JochiRaider/cartulary/cmd/server\nexit "${FAKE_GO_EXIT:-0}"\n', { mode: 0o700 });
    writeFileSync(scanner, '#!/bin/sh\nprintf invoked >"$FAKE_SCANNER_INVOKED"\nif [ "$FAKE_SCANNER_CONFLICT" = 1 ]; then "$NODE_BIN" "$FAKE_DIAGNOSTIC_FIXTURE" conflict >&2; fi\ncat "$FAKE_SCANNER_OUTPUT"\nexit "$FAKE_SCANNER_EXIT"\n', { mode: 0o700 });
    writeFileSync(makefile, `go-vulncheck:\n\t@bash "${root}/tools/harness/static-analysis/go-govulncheck.sh"\n`);
    const config = { config: { protocol_version: "v1.0.0", scanner_name: "govulncheck", scanner_version: "v1.3.0", scan_level: "symbol", scan_mode: "source" } };
    const finding = (reachability) => ({ finding: { osv: "GO-2099-0001", fixed_version: "v1.2.3", trace: [{ module: "example.com/vulnerable", version: "v1.2.2", ...(reachability !== "module" ? { package: "example.com/vulnerable" } : {}), ...(reachability === "symbol" ? { function: "Explode" } : {}) }] } });
    const events = (...values) => values.map((value) => JSON.stringify(value)).join("\n");
    const scenarios = [
      { name: "clean", output: events(config) },
      { name: "diagnostic", output: events(config, finding("module"), finding("package")), diagnostic: true },
      { name: "blocking", output: events(config, finding("symbol")), failure: ["security", "security_finding", 1] },
      { name: "blocking-nonzero", output: events(config, finding("symbol")), scannerExit: 3, failure: ["security", "security_finding", 1] },
      { name: "malformed", output: "not-json", failure: ["artifact", "artifact_error", 11] },
      { name: "missing-findings", output: "", failure: ["artifact", "artifact_error", 11] },
      { name: "scanner-failure", output: events(config), scannerExit: 9, failure: ["harness", "tool_diagnostic_failure", 1] },
      { name: "scanner-failure-no-output", output: "", scannerExit: 9, failure: ["harness", "tool_diagnostic_failure", 1] },
      { name: "missing-tool", output: "", missingTool: true, failure: ["config", "configuration_error", 2] },
      { name: "package-discovery-failure", output: "", goExit: 9, failure: ["harness", "tool_diagnostic_failure", 1] },
      { name: "conflicting-diagnostics", output: events(config, finding("symbol")), conflict: true, failure: ["harness", "scheduler_accounting_error", 11] },
    ];
    for (const scenario of scenarios) {
      writeFileSync(payload, scenario.output);
      rmSync(invoked, { force: true });
      const artifacts = path.join(scratch, scenario.name);
      const unit = {
        ...canonical, needs: [], cache_policy: "none", resource_claims: { cpu: 1 }, timeout_ms: 10000,
        command: { executable: "make", args: ["--silent", "--no-print-directory", "-f", makefile, "go-vulncheck"], environment: {
          CARTULARY_TEST_TARGET: "go-vulncheck", CARTULARY_HARNESS_CACHE_MODE: "off", NODE_BIN: process.execPath,
          GO: fakeGo, GOVULNCHECK_BIN: scenario.missingTool ? path.join(scratch, "absent") : scanner,
          GO_CACHE_DIR: path.join(scratch, "cache"), GO_MOD_CACHE_DIR: path.join(scratch, "modules"), GO_TMP_DIR: scratch,
          CARTULARY_STEP_ARTIFACT_DIR: artifacts, FAKE_SCANNER_OUTPUT: payload, FAKE_SCANNER_INVOKED: invoked,
          FAKE_SCANNER_EXIT: String(scenario.scannerExit ?? 0), FAKE_GO_EXIT: String(scenario.goExit ?? 0),
          FAKE_SCANNER_CONFLICT: scenario.conflict ? "1" : "0", FAKE_DIAGNOSTIC_FIXTURE: fixture,
        } },
      };
      const independent = { ...unit, unit_id: "target:independent", command: { executable: "true", args: [], environment: {} }, current_run_evidence_outputs: ["unit-results/target-independent.json"] };
      const result = await runWorkGraph({ graph: buildWorkGraph([unit, independent]), capacities: new Map([["cpu", 1]]), cwd: root, environment, executeUnit: execute });
      const outcome = result.unit_results[unit.unit_id];
      assert.equal(result.unit_results[independent.unit_id].status, "passed", "independent aggregate work completes");
      if (scenario.failure) {
        assert.deepEqual([outcome.failure_class, outcome.failure_reason, outcome.exit_code], scenario.failure, `${scenario.name}: ${JSON.stringify(outcome)}`);
        const failed = result.events.find((event) => event.unit_id === unit.unit_id && event.event === "failed");
        assert.deepEqual([failed.failure_class, failed.failure_reason], scenario.failure.slice(0, 2), "aggregate projection journal preserves the cause");
        assert.match(outcome.stderr, /Error (1|2|11)/u, "GNU Make wrapped the underlying failure");
      } else assert.equal(outcome.status, "passed", `${scenario.name}: ${JSON.stringify(outcome)}`);
      assert.equal(Boolean(lstatSync(invoked, { throwIfNoEntry: false })), !scenario.missingTool && !scenario.goExit, "fixtures execute freshly; discovery failures stop before scanning");
      if (scenario.diagnostic) {
        const findings = JSON.parse(readFileSync(path.join(artifacts, "govulncheck-findings.json")));
        assert.equal(findings.counts.blocking_count, 0);
        assert.deepEqual(findings.counts.reachability, { module: 1, package: 1, symbol: 0 });
      }
      assert.equal(readdirSync(runtime.root).some((name) => name.startsWith("command-failure-")), false, "each scanner invocation releases its private diagnostic channel");
    }
  });
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
