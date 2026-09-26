import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { randomUUID } from "node:crypto";
import { existsSync, mkdirSync, readFileSync, rmSync, statSync, symlinkSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { primaryPublicFailure, publicExitCodeForFailures, validateSchemaSync } from "../../contract/index.mjs";
import { normalizeVitestObservations, readVitestJSON, reconcileVitestReport } from "../../diagnostics/vitest-failure-details.mjs";
import { adaptVitestInvocation } from "../runners/vitest.mjs";
import { createCommandFailureContext } from "../../runtime/command-failure.mjs";
import { createSuiteRuntime } from "../../runtime/suite-runtime.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../../..");
const fixture = path.join(root, "tmp", `vitest-diagnostics-${randomUUID()}`);
mkdirSync(fixture, { recursive: true, mode: 0o700 });
const runRoot = path.resolve(root, process.env.CARTULARY_TEST_RESULTS_DIR, process.env.CARTULARY_TEST_RUN_ID);
const output = path.join(runRoot, "vitest-diagnostics-contract", randomUUID());
mkdirSync(output, { recursive: true, mode: 0o700 });
const ownedSuiteRuntime = process.env.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT
  ? null
  : createSuiteRuntime({ repoRoot: root, runRoot, runID: process.env.CARTULARY_TEST_RUN_ID });
const suiteRuntime = ownedSuiteRuntime?.root ?? process.env.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT;
const fixtureEnvironment = {
  ...process.env,
  ...(ownedSuiteRuntime ? {
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: ownedSuiteRuntime.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: ownedSuiteRuntime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: ownedSuiteRuntime.runID,
  } : {}),
};
fixtureEnvironment.PATH = [path.dirname(process.execPath), fixtureEnvironment.PATH]
  .filter(Boolean).join(path.delimiter);
// Deliberate child failures must not publish into the enclosing successful test command.
delete fixtureEnvironment.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
const reportFile = path.join(output, "runner.json");
const detailsFile = path.join(output, "vitest-failure-details.json");
const api = pathToFileURL(path.join(root, "apps/web/node_modules/vitest/dist/index.js")).href;
const testFile = path.join(fixture, "diagnostics.test.mjs");
const configFile = path.join(fixture, "vitest.config.mjs");
try {
  writeFileSync(configFile, `export default ${JSON.stringify({ root: fixture, test: { include: ["**/*.test.mjs"], environment: "node", testTimeout: 500, hookTimeout: 100, retry: 0 } })};`);
  writeFileSync(testFile, `import { test, describe, beforeEach, expect } from ${JSON.stringify(api)};
test('deadline', async () => new Promise(() => {}), 30);
describe('hook', () => { beforeEach(async () => new Promise(() => {}), 30); test('deadline', () => {}); });
test('original cause', () => { const e = new Error('precise original assertion'); e.name = 'AssertionError'; e.stack = 'Error: STACK_TRACE_ERROR\\n    at registration'; throw e; });
test('marker only', () => { throw new Error('STACK_TRACE_ERROR'); });
test('redaction', () => { throw new Error('password=vitest-fixture-sensitive-value'); });
describe('left', () => { test('same name', () => expect(1).toBe(1)); });
describe('right', () => { test('same name', () => expect(2).toBe(2)); });
`);
  const command = [path.join(root, "tools/harness/execution/vitest-invocation-cli.mjs"), reportFile, detailsFile, "--",
    process.env.PNPM || path.join(root, "tmp/node-runtime/bin/pnpm"), "--dir", "apps/web", "exec", "vitest", "run", `--config=${configFile}`, "--maxWorkers=1"];
  const result = spawnSync(process.execPath, command, { cwd: root, env: fixtureEnvironment, encoding: "utf8", timeout: 30000 });
  assert.ok(existsSync(detailsFile), `real collector did not retain diagnostics: ${result.stderr}`);
  const details = readVitestJSON(detailsFile);
  const report = readVitestJSON(reportFile);
  validateSchemaSync(details.schema_id, details);
  assert.equal(result.status, publicExitCodeForFailures(details.failures));
  assert.equal(details.failures.find((item) => item.title === "deadline").kind, "test_timeout");
  assert.equal(details.failures.find((item) => item.title === "hook deadline").kind, "hook_timeout");
  assert.equal(details.failures.find((item) => item.title === "original cause").message, "precise original assertion");
  assert.match(details.failures.find((item) => item.title === "original cause").original_error.stack, /STACK_TRACE_ERROR/u);
  assert.equal(details.failures.find((item) => item.title === "marker only").failure_class, "unknown");
  assert.ok(!readFileSync(detailsFile, "utf8").includes("vitest-fixture-sensitive-value"));
  assert.ok(!readFileSync(reportFile, "utf8").includes("vitest-fixture-sensitive-value"));
  assert.equal(statSync(detailsFile).mode & 0o777, 0o600);
  const file = path.relative(root, testFile);
  const invocation = { root, file, rows: [
    { row_id: "test.timeout", selectors: ["deadline"] },
    { row_id: "test.assertion", selectors: ["original cause"] },
    { row_id: "test.unknown", selectors: ["marker only"] },
    { row_id: "test.names", selectors: ["left same name", "right same name"] },
    { row_id: "test.hook", selectors: ["hook deadline"] },
  ] };
  const rows = adaptVitestInvocation(invocation, { details, report, status: result.status });
  assert.equal(rows[0].exit_code, 13);
  assert.equal(rows[0].failure_class, "timing");
  assert.ok(rows[0].duration_ms >= 30);
  assert.equal(rows[1].exit_code, 10);
  assert.equal(rows[2].failure_class, "unknown");
  assert.equal(rows[3].terminal_state, "passed");
  assert.equal(rows[4].exit_code, 13);
  assert.ok(rows[4].duration_ms >= 30);
  assert.equal(adaptVitestInvocation(invocation, { details: {}, report })[0].failure_reason, "artifact_error");
  assert.equal(adaptVitestInvocation({ ...invocation, rows: [{ row_id: "ambiguous", selectors: ["same name"] }] }, { details, report })[0].failure_reason, "scheduler_accounting_error");
  const raw = structuredClone(details);
  raw.failures = [];
  const context = { invocationID: raw.invocation_id, runID: raw.run_id, runnerJSON: reportFile, stdoutLog: "", stderrLog: "" };
  assert.throws(() => normalizeVitestObservations(raw, { ...context, invocationID: "wrong" }), /different invocation/u);
  assert.throws(() => normalizeVitestObservations({ ...raw, observations: [...raw.observations, raw.observations[0]] }, context), /Duplicate/u);
  assert.throws(() => normalizeVitestObservations({ ...raw, schema_id: "cartulary.vitest_failure_details.v1" }, context), /schema/u);
  const contradictory = structuredClone(report);
  contradictory.testResults[0].assertionResults[0].status = "passed";
  assert.throws(() => reconcileVitestReport(details, contradictory, root), /contradict/u);
  assert.equal(adaptVitestInvocation(invocation, { details, report: contradictory })[0].failure_reason, "scheduler_accounting_error");
  const normalized = normalizeVitestObservations(raw, context);
  assert.equal(primaryPublicFailure(normalized.failures).failure_class, primaryPublicFailure(details.failures).failure_class);
  assert.throws(() => readVitestJSON(path.join(fixture, "absent.json")), /missing/u);
  writeFileSync(path.join(fixture, "malformed.json"), "{");
  assert.throws(() => readVitestJSON(path.join(fixture, "malformed.json")), /malformed/u);
  symlinkSync(reportFile, path.join(fixture, "linked.json"));
  assert.throws(() => readVitestJSON(path.join(fixture, "linked.json")), /unsafe/u);
  assert.equal(adaptVitestInvocation(invocation, { signal: "SIGTERM", status: null })[0].exit_code, 143);
  assert.equal(existsSync(path.join(suiteRuntime, `vitest-${details.invocation_id}`)), false);
  const label = `vitest-diagnostic-${randomUUID()}`;
  const fullEnvironment = { ...fixtureEnvironment, CARTULARY_OUTPUT_MODE: "quiet", NODE_BIN: process.execPath };
  delete fullEnvironment.CARTULARY_HARNESS_IDENTITY_PREPARED;
  delete fullEnvironment.CARTULARY_TEST_TARGET;
  const full = spawnSync("bash", [path.join(root, "tools/harness/execution/run-vitest-step.sh"), label, "--", ...command.slice(4)], {
    cwd: root, env: fullEnvironment,
    encoding: "utf8", timeout: 30000,
  });
  assert.equal(full.status, publicExitCodeForFailures(details.failures), full.stderr);
  const fullSummaryFile = path.join(runRoot, "adhoc", label, "step-summary.json");
  assert.ok(existsSync(fullSummaryFile), `missing full summary: ${full.stdout}\n${full.stderr}`);
  const fullSummary = readVitestJSON(fullSummaryFile);
  assert.equal(fullSummary.failure_reason, primaryPublicFailure(details.failures).failure_reason);
  assert.ok(fullSummary.dossiers.some((item) => item.failure_class === "timing" && item.message.includes("timed out")));
  assert.ok(fullSummary.dossiers.some((item) => item.failure_class === "product" && item.message.includes("precise original assertion")));
  assert.ok(fullSummary.dossiers.some((item) => item.failure_class === "unknown"));
  const timeoutLabel = `vitest-timeout-${randomUUID()}`;
  const timeoutOnly = spawnSync("bash", [path.join(root, "tools/harness/execution/run-vitest-step.sh"), timeoutLabel,
    "--", ...command.slice(4), "-t", "^deadline$"], { cwd: root, env: fullEnvironment, encoding: "utf8", timeout: 30000 });
  assert.equal(timeoutOnly.status, 13, timeoutOnly.stderr);
  const timeoutSummary = readVitestJSON(path.join(runRoot, "adhoc", timeoutLabel, "step-summary.json"));
  assert.equal(timeoutSummary.failure_reason, "timeout_failure");
  const fakeRunner = path.join(fixture, "invalid-runner.mjs");
  writeFileSync(fakeRunner, `import { readFileSync, writeFileSync } from 'node:fs';
const report = ${JSON.stringify(report)};
const capture = ${JSON.stringify(raw)};
capture.invocation_id = process.env.CARTULARY_VITEST_INVOCATION_ID;
const mode = process.env.VITEST_DIAGNOSTIC_FIXTURE_MODE;
if (mode === 'contradictory') capture.observations[0].status = 'skipped';
if (mode === 'interrupted') {
  capture.run_status = 'interrupted';
  for (const observation of capture.observations) { observation.status = 'skipped'; observation.errors = []; }
  for (const suite of report.testResults) for (const assertion of suite.assertionResults) assertion.status = 'skipped';
}
writeFileSync(process.argv.find(arg => arg.startsWith('--outputFile.json=')).slice('--outputFile.json='.length), JSON.stringify(report));
if (mode !== 'missing') writeFileSync(process.env.CARTULARY_VITEST_CAPTURE_FILE, mode === 'malformed' ? '{' : JSON.stringify(capture));
`);
  for (const mode of ["missing", "malformed", "contradictory", "interrupted"]) {
    const invalidLabel = `vitest-invalid-${randomUUID()}`;
    const failureContext = createCommandFailureContext({ repoRoot: root, environment: fullEnvironment,
      unitID: invalidLabel, commandID: "cartulary.harness.command.frontend_unit.v1" });
    try {
      const invalid = spawnSync("bash", [path.join(root, "tools/harness/execution/run-vitest-step.sh"), invalidLabel,
        "--", process.execPath, fakeRunner], { cwd: root,
        env: { ...fullEnvironment, ...failureContext.environment, VITEST_DIAGNOSTIC_FIXTURE_MODE: mode }, encoding: "utf8", timeout: 30000 });
      assert.equal(invalid.status, mode === "interrupted" ? 130 : 11, invalid.stderr);
      const invalidSummary = readVitestJSON(path.join(runRoot, "adhoc", invalidLabel, "step-summary.json"));
      const expectedReason = mode === "interrupted" ? "cancelled_or_interrupted"
        : mode === "contradictory" ? "scheduler_accounting_error" : "artifact_error";
      assert.equal(invalidSummary.failure_reason, expectedReason);
      assert.equal(adaptVitestInvocation(invocation, { status: invalid.status, commandFailure: failureContext.read() })[0].failure_reason, expectedReason);
    } finally { failureContext.close(); }
  }
  const loadFile = path.join(fixture, "load.test.mjs");
  writeFileSync(loadFile, "throw new Error('suite collection failed');");
  const loadCommand = [...command, loadFile];
  loadCommand[1] = path.join(output, "load-runner.json");
  loadCommand[2] = path.join(output, "load-details.json");
  const loadResult = spawnSync(process.execPath, loadCommand, { cwd: root, env: fixtureEnvironment, encoding: "utf8", timeout: 30000 });
  assert.equal(loadResult.status, 3, loadResult.stderr);
  const loadDetails = readVitestJSON(loadCommand[2]);
  assert.equal(loadDetails.failures[0].kind, "suite_setup");
  const loadRows = adaptVitestInvocation({ root, file: path.relative(root, loadFile), rows: [{ row_id: "load", selectors: ["not collected"] }] },
    { details: loadDetails, report: readVitestJSON(loadCommand[1]), status: loadResult.status });
  assert.equal(loadRows[0].failure_class, "infra");
  const smoke = spawnSync("bash", [path.join(root, "tools/harness/execution/tests/test-run-vitest-step.sh")], {
    cwd: root, env: { ...fixtureEnvironment, NODE_BIN: process.execPath }, encoding: "utf8", timeout: 120000,
  });
  assert.equal(smoke.status, 0, `full Vitest wrapper smoke: ${smoke.stdout}\n${smoke.stderr}`);
  process.stdout.write("Vitest real collector, classification, attribution, redaction, and validation contracts passed\n");
} finally {
  rmSync(fixture, { recursive: true, force: true });
  ownedSuiteRuntime?.close();
}
