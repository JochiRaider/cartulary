import assert from "node:assert/strict";
import { spawn, spawnSync } from "node:child_process";
import { randomUUID } from "node:crypto";
import { existsSync, mkdirSync, readFileSync, rmSync, statSync, symlinkSync, writeFileSync } from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";
import { primaryPublicFailure, publicExitCodeForFailures, validateSchemaSync } from "../../contract/index.mjs";
import { createVitestEvidence, normalizeVitestObservations, readVitestEvidence, readVitestJSON, reconcileVitestReport } from "../../diagnostics/vitest-failure-details.mjs";
import { adaptVitestInvocation, adaptVitestInvocationFile, buildVitestInvocations, vitestInvocation, vitestInvocationFailure } from "../runners/vitest.mjs";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { publishRowResult } from "../runners/row-results.mjs";
import { createRuntime } from "../../runtime/session-client.mjs";
import { openExecutionChannel } from "../../test-support/execution-channel-fixture.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const fixture = path.join(root, "tmp", `vitest-diagnostics-${randomUUID()}`);
const groupedFixture = path.join(root, "apps/web/.cartulary", `vitest-group-${randomUUID()}`);
mkdirSync(fixture, { recursive: true, mode: 0o700 });
mkdirSync(groupedFixture, { recursive: true, mode: 0o700 });
const runRoot = path.resolve(root, process.env.CARTULARY_TEST_RESULTS_DIR, process.env.CARTULARY_TEST_RUN_ID);
const output = path.join(runRoot, "vitest-diagnostics-contract", randomUUID());
mkdirSync(output, { recursive: true, mode: 0o700 });
const ownedRuntime = process.env.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT ? null : createRuntime({ repoRoot: root, runRoot, runID: process.env.CARTULARY_TEST_RUN_ID });
const fixtureEnvironment = { ...process.env, ...(ownedRuntime ? {
  CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: ownedRuntime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: ownedRuntime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: ownedRuntime.runID,
} : {}) };
fixtureEnvironment.VITEST_MAX_WORKERS = "1";
fixtureEnvironment.PATH = [path.dirname(process.execPath), fixtureEnvironment.PATH].filter(Boolean).join(path.delimiter);
delete fixtureEnvironment.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
const api = pathToFileURL(path.join(root, "apps/web/node_modules/vitest/dist/index.js")).href;
const cli = path.join(root, "tools/harness/execution/vitest-invocation-cli.mjs");
const configFile = path.join(fixture, "vitest.config.mjs");
const relative = (file) => path.relative(root, file);
const save = (file, value) => { mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 }); writeFileSync(file, `${JSON.stringify(value)}\n`, { mode: 0o600 }); };
function config(extra = {}) { writeFileSync(configFile, `export default ${JSON.stringify({ root: fixture, test: { name: "fixture", include: ["**/*.test.mjs"], environment: "node", testTimeout: 500, hookTimeout: 100, retry: 0, ...extra } })};`); }
function request(files) { return { schema_id: "cartulary.vitest_invocation.v1", project: "fixture", config: relative(configFile), publish_rows: false,
  files: files.map(([file, titles]) => ({ file: relative(file), rows: titles.map((title, index) => ({ row_id: `${path.basename(file)}.${index}`, selectors: [title] })) })) }; }
function prepare(name, selection) {
  const directory = path.join(output, name);
  const requestFile = path.join(directory, "invocation.json"), runnerJSON = path.join(directory, "runner.json"), detailsFile = path.join(directory, "vitest-failure-details.json");
  save(requestFile, selection);
  return { ...vitestInvocation(root, selection), requestFile, runnerJSON, detailsFile, args: [cli, runnerJSON, detailsFile, requestFile] };
}
function run(name, selection) {
  const invocation = prepare(name, selection);
  const result = spawnSync(process.execPath, invocation.args, { cwd: root, env: fixtureEnvironment, encoding: "utf8", timeout: 30000 });
  assert.ok(existsSync(invocation.detailsFile), `missing diagnostics (${name}): ${result.stderr}`);
  const details = readVitestJSON(invocation.detailsFile), report = readVitestJSON(path.resolve(root, details.runner_json));
  validateSchemaSync(details.schema_id, details);
  return { invocation, result, details, report, rows: adaptVitestInvocation(invocation, { details, report, status: result.status }) };
}
function runObserved(name, selection) {
  const runtime = createRuntime({ repoRoot: root, runRoot, runID: process.env.CARTULARY_TEST_RUN_ID });
  const channel = openExecutionChannel(runtime);
  const before = { ...fixtureEnvironment };
  Object.assign(fixtureEnvironment, channel.environment, {
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runtime.runID,
  });
  try {
    const result = run(name, selection);
    channel.close();
    return { ...result, activities: channel.records().map((entry) => entry.record ?? entry).filter((entry) => entry.kind === "activity") };
  } finally {
    for (const key of Object.keys(fixtureEnvironment)) delete fixtureEnvironment[key];
    Object.assign(fixtureEnvironment, before);
    runtime.close();
  }
}
async function waitForFile(file, child) {
  const end = Date.now() + 15000;
  while (!existsSync(file) && Date.now() < end && child.exitCode === null && child.signalCode === null) await new Promise((resolve) => setTimeout(resolve, 20));
  assert.ok(existsSync(file), `child did not reach ${path.basename(file)}`);
}
try {
  config();
  const testFile = path.join(fixture, "diagnostics.test.mjs");
  const titles = ["deadline", "hook deadline", "original cause", "marker only", "redaction", "left same name", "right same name"];
  writeFileSync(testFile, `import { test, describe, beforeEach, expect } from ${JSON.stringify(api)};
test('deadline', async () => new Promise(() => {}), 30);
describe('hook', () => { beforeEach(async () => new Promise(() => {}), 30); test('deadline', () => {}); });
test('original cause', () => { const e = new Error('precise original assertion'); e.name = 'AssertionError'; e.stack = 'Error: STACK_TRACE_ERROR\\n    at registration'; throw e; });
test('marker only', () => { throw new Error('STACK_TRACE_ERROR'); });
test('redaction', () => { throw new Error('password=vitest-fixture-sensitive-value'); });
describe('left', () => { test('same name', () => expect(1).toBe(1)); });
describe('right', () => { test('same name', () => expect(2).toBe(2)); });`);
  const selected = request([[testFile, titles]]);
  const full = run("causes", selected);
  const { details, report, invocation } = full;
  assert.equal(full.result.status, publicExitCodeForFailures(details.failures), full.result.stderr);
  assert.equal(details.failures.find((item) => item.title === "deadline").kind, "test_timeout");
  assert.equal(details.failures.find((item) => item.title === "hook deadline").kind, "hook_timeout");
  assert.equal(details.failures.find((item) => item.title === "original cause").message, "precise original assertion");
  assert.match(details.failures.find((item) => item.title === "original cause").original_error.stack, /STACK_TRACE_ERROR/u);
  assert.equal(details.failures.find((item) => item.title === "marker only").failure_class, "unknown");
  for (const file of [invocation.detailsFile, invocation.runnerJSON, path.join(path.dirname(invocation.runnerJSON), "vitest-evidence/file-0.json")]) assert.ok(!readFileSync(file, "utf8").includes("vitest-fixture-sensitive-value"));
  assert.equal(statSync(invocation.detailsFile).mode & 0o777, 0o600);
  assert.deepEqual(full.rows.map((row) => row.failure_class), ["timing", "timing", "product", "unknown", "product", null, null]);
  assert.ok(full.rows[0].duration_ms >= 30 && full.rows[1].duration_ms >= 30);
  assert.equal(adaptVitestInvocation(invocation, { details: {}, report })[0].failure_reason, "artifact_error");
  const missing = { ...invocation, rows: [...invocation.rows, { ...invocation.rows[0], row_id: "missing", selectors: ["absent"] }] };
  const partial = adaptVitestInvocation(missing, { details, report, status: full.result.status });
  assert.equal(partial[5].terminal_state, "passed");
  assert.equal(partial.at(-1).failure_reason, "scheduler_accounting_error");
  assert.ok(adaptVitestInvocation({ ...invocation, rows: invocation.rows.slice(0, -1) }, { details, report }).every((row) => row.failure_reason === "scheduler_accounting_error"));
  const raw = { ...structuredClone(details), failures: [] };
  const context = { invocationID: raw.invocation_id, runID: raw.run_id, runnerJSON: invocation.runnerJSON, stdoutLog: "", stderrLog: "" };
  assert.throws(() => normalizeVitestObservations(raw, { ...context, invocationID: "wrong" }), /different invocation/u);
  assert.throws(() => normalizeVitestObservations({ ...raw, observations: [...raw.observations, raw.observations[0]] }, context), /Duplicate/u);
  assert.throws(() => normalizeVitestObservations({ ...raw, schema_id: "cartulary.vitest_failure_details.v2" }, context), /schema/u);
  const contradictory = structuredClone(report);
  contradictory.reports[0].report.testResults[0].assertionResults[0].status = "passed";
  assert.throws(() => reconcileVitestReport(details, contradictory, root), /contradict/u);
  assert.equal(adaptVitestInvocation(invocation, { details, report: contradictory })[0].failure_reason, "scheduler_accounting_error");
  for (const mutation of [
    (value) => { value.reports.push(value.reports[0]); },
    (value) => { value.reports = []; },
    (value) => { value.reports[0].project = "foreign-project"; },
  ]) {
    const invalid = structuredClone(report); mutation(invalid);
    assert.ok(adaptVitestInvocation(invocation, { details, report: invalid }).every((row) => row.terminal_state !== "passed"));
  }
  assert.equal(primaryPublicFailure(normalizeVitestObservations(raw, context).failures).failure_class, primaryPublicFailure(details.failures).failure_class);
  assert.throws(() => readVitestJSON(path.join(fixture, "absent.json")), /missing/u);
  writeFileSync(path.join(fixture, "malformed.json"), "{");
  assert.throws(() => readVitestJSON(path.join(fixture, "malformed.json")), /malformed/u);
  writeFileSync(path.join(fixture, "duplicate.json"), '{"same":1,"same":2}');
  assert.throws(() => readVitestJSON(path.join(fixture, "duplicate.json")), /malformed/u);
  symlinkSync(invocation.runnerJSON, path.join(fixture, "linked.json"));
  assert.throws(() => readVitestJSON(path.join(fixture, "linked.json")), /unsafe/u);
  assert.equal(adaptVitestInvocation(invocation, { signal: "SIGTERM", status: null })[0].exit_code, 143);

  // Exercise crash boundaries without launching workers: the durable protocol
  // must recover independently of row publication or terminal-envelope commit.
  const synthetic = (name, count, interruptPublish = () => false) => {
    const selection = request(Array.from({ length: count }, (_, index) => [path.join(fixture, `file-${index}.test.mjs`), ["same title"]]));
    const invocation = prepare(name, selection), invocationID = randomUUID();
    const writes = [];
    const settings = { root, request: selection, reportFile: invocation.runnerJSON, detailsFile: invocation.detailsFile,
      runID: process.env.CARTULARY_TEST_RUN_ID, invocationID, startedAt: "2026-01-01T00:00:00.000Z" };
    const evidence = createVitestEvidence({ ...settings, publish(file, bytes, options) {
      if (interruptPublish(file)) throw new Error("publication interrupted");
      writes.push({ file, bytes: Buffer.byteLength(bytes) }); atomicLocalFile(file, bytes, options);
    } });
    const settle = (index) => {
      const file = selection.files[index].file;
      const capture = { ...raw, invocation_id: invocationID, run_id: settings.runID, run_status: "passed", final_complete: false,
        completed_files: [{ project: selection.project, owner_path: file }],
        observations: [{ project: selection.project, owner_path: file, title: "same title", test_name: "same title",
          scope: "test", status: "passed", duration_ms: 1, timeout_ms: null, errors: [] }], failures: [] };
      evidence.settleFile(index, capture, { success: true, testResults: [{ name: path.resolve(root, file),
        assertionResults: [{ fullName: "same title", status: "passed" }] }] });
    };
    return { ...settings, invocation, evidence, writes, settle,
      read: () => readVitestEvidence(settings),
      packet: (index) => path.join(path.dirname(invocation.runnerJSON), "vitest-evidence", `file-${index}.json`) };
  };
  const growth = [];
  for (const count of [4, 8, 16]) {
    const current = synthetic(`growth-${count}`, count);
    current.settle(0);
    const first = statSync(current.packet(0));
    for (let index = 1; index < count; index++) current.settle(index);
    assert.equal(statSync(current.packet(0)).mtimeMs, first.mtimeMs);
    assert.equal(statSync(current.packet(0)).ino, first.ino);
    assert.equal(current.writes.length, count + 1, "one immutable publication per file plus the initial disposition");
    assert.ok(adaptVitestInvocation(current.invocation, current.read()).every((row) => row.terminal_state === "passed"));
    current.evidence.finish("");
    assert.equal(current.writes.length, count + 5, "terminal envelopes are each published once");
    assert.equal(current.read().details.final_complete, true);
    growth.push(current.writes.reduce((sum, item) => sum + item.bytes, 0));
  }
  assert.ok(growth[1] < growth[0] * 2.1 && growth[2] < growth[1] * 2.1, `serialized publication growth: ${growth}`);
  const partialPublication = synthetic("partial-publication", 2, (file) => file.endsWith("vitest-failure-details.json"));
  partialPublication.settle(0); partialPublication.settle(1);
  assert.throws(() => partialPublication.evidence.finish(""), /publication interrupted/u);
  assert.equal(partialPublication.read().details.final_complete, false);
  assert.ok(adaptVitestInvocation(partialPublication.invocation, partialPublication.read()).every((row) => row.terminal_state === "passed"));
  partialPublication.evidence.recordRunFailures([{ project: "", owner_path: "", title: "global close", test_name: "",
    scope: "run", status: "failed", duration_ms: 0, timeout_ms: null, errors: [{ name: "Error", message: "global close failed", stack: "" }] }]);
  assert.ok(adaptVitestInvocation(partialPublication.invocation, partialPublication.read()).every((row) => row.failure_class === "infra"));
  partialPublication.evidence.fail({ failure_class: "harness", failure_reason: "scheduler_accounting_error" });
  assert.ok(adaptVitestInvocation(partialPublication.invocation, partialPublication.read()).every((row) => row.failure_reason === "scheduler_accounting_error"));
  const corrupt = synthetic("invalid-evidence", 2);
  corrupt.settle(0); corrupt.settle(1);
  const original = readVitestJSON(corrupt.packet(0));
  for (const mutate of [
    (packet) => { packet.run_id = "foreign"; },
    (packet) => { packet.invocation_id = "foreign"; },
    (packet) => { packet.request_digest = `sha256:${"0".repeat(64)}`; },
    (packet) => { packet.artifact_root = "foreign"; },
    (packet) => { packet.file_index = 1; },
    (packet) => { packet.details.completed_files[0].project = "foreign"; },
    (packet) => { packet.report.reports[0].report.testResults[0].assertionResults[0].status = "failed"; },
    (packet) => { packet.details.observations.push(packet.details.observations[0]); },
    (packet) => { packet.details.observations = []; },
    (packet) => { packet.details.completed_files.push(packet.details.completed_files[0]); },
  ]) {
    const packet = structuredClone(original); mutate(packet); save(corrupt.packet(0), packet);
    assert.throws(corrupt.read, undefined, "invalid checkpoint cannot establish success");
  }
  const skipped = structuredClone(original);
  skipped.details.observations[0].status = "skipped";
  skipped.report.reports[0].report.testResults[0].assertionResults[0].status = "skipped";
  save(corrupt.packet(0), skipped);
  assert.equal(adaptVitestInvocation(corrupt.invocation, corrupt.read())[0].failure_reason, "scheduler_accounting_error");
  save(corrupt.packet(0), original);
  assert.throws(() => readVitestEvidence({ ...corrupt, request: { ...corrupt.request, project: "foreign" } }), /different invocation/u);
  rmSync(corrupt.packet(0));
  assert.throws(corrupt.read, /gap/u);
  symlinkSync(corrupt.packet(1), corrupt.packet(0));
  assert.throws(corrupt.read, /unsafe/u);
  rmSync(corrupt.packet(0)); writeFileSync(corrupt.packet(0), "{", { mode: 0o600 });
  assert.throws(corrupt.read, /malformed/u);
  save(corrupt.packet(0), original); save(corrupt.packet(2), original);
  assert.throws(corrupt.read, /Unexpected/u); rmSync(corrupt.packet(2));
  writeFileSync(corrupt.packet(0), " ".repeat(32 * 1024 * 1024 + 1));
  assert.throws(corrupt.read, /oversized/u);
  save(corrupt.packet(0), original);
  assert.ok(adaptVitestInvocationFile(corrupt.invocation, { status: 0 }).every((row) => row.failure_reason === "scheduler_accounting_error"), "zero exit without final settlement is contradictory");
  assert.ok(adaptVitestInvocationFile(corrupt.invocation, { status: 11,
    commandFailure: { failure_class: "artifact", failure_reason: "artifact_error" } }).every((row) => row.failure_reason === "artifact_error"), "private command failure must not be erased by a null disposition failure");
  const rowRoot = path.join(output, "stable-row");
  const row = { schema_id: "cartulary.harness_row_result.v6", ...full.rows[5], row_id: "harness.command_surface.behavior.stable_fixture", runner: "vitest",
    started_at: "2026-01-01T00:00:00.000Z", finished_at: "2026-01-01T00:00:01.000Z", wall_duration_ms: 1000 };
  publishRowResult(rowRoot, row, { preserveEquivalent: true });
  const rowFile = path.join(rowRoot, "rows", `${row.row_id}.json`), before = statSync(rowFile);
  publishRowResult(rowRoot, { ...row, finished_at: "2026-01-01T00:00:02.000Z", wall_duration_ms: 2000 }, { preserveEquivalent: true });
  assert.equal(statSync(rowFile).ino, before.ino);
  assert.deepEqual(readVitestJSON(rowFile), row);
  publishRowResult(rowRoot, { ...row, ...full.rows[0], row_id: row.row_id }, { preserveEquivalent: true });
  assert.equal(readVitestJSON(rowFile).failure_class, "timing", "changed outcomes supersede earlier rows");

  const firstFile = path.join(root, "apps/web/src/testing/fixtures/vitestSelectionTestSupport.ts");
  const secondFile = path.join(root, "apps/web/src/testing/fixtures/vitestCohortTestSupport.ts");
  const groupedConfig = path.join(groupedFixture, "vite.config.mjs");
  writeFileSync(groupedConfig, `export default ${JSON.stringify({ root: groupedFixture, test: { name: "harness-node", include: [path.relative(groupedFixture, firstFile), path.relative(groupedFixture, secondFile)], environment: "node" } })};`);
  const groupedRows = [
    ...["one", "two"].map((title) => ({ row_id: `fixture.${title}`, runner: "vitest", fixture_capability: "none", selector: { file: relative(firstFile), project_id: "harness-node", titles: [title] } })),
    { row_id: "fixture.independent", runner: "vitest", fixture_capability: "none", selector: { file: relative(secondFile), project_id: "harness-node", titles: ["independent"] } },
  ];
  for (const reverse of [false, true]) {
    const [grouped] = buildVitestInvocations(root, groupedRows, path.join(output, `cohort-${reverse}`));
    const selected = readVitestJSON(grouped.requestFile);
    selected.config = relative(groupedConfig); selected.publish_rows = false;
    if (reverse) selected.files.reverse();
    save(grouped.requestFile, selected);
    const result = spawnSync(grouped.command, grouped.args, { cwd: root, env: fixtureEnvironment, encoding: "utf8", timeout: 30000 });
    assert.equal(result.status, 0, result.stderr);
    assert.deepEqual(adaptVitestInvocationFile(grouped, result).map((row) => row.terminal_state), ["passed", "passed", "passed"]);
    const observed = readVitestJSON(grouped.detailsFile);
    assert.equal(observed.completed_files.length, 2);
    assert.equal(observed.observations.filter((item) => item.status === "passed").length, 3);
  }
  assert.throws(() => buildVitestInvocations(root, [groupedRows[0], { ...groupedRows[1], process_isolation: "exclusive" }], output), /compatible project/u);
  assert.throws(() => buildVitestInvocations(root, [groupedRows[0], { ...groupedRows[1], selector: { ...groupedRows[1].selector, project_id: "browser-unit" } }], output), /compatible project/u);
  assert.throws(() => buildVitestInvocations(root, [groupedRows[0], { ...groupedRows[0], row_id: "overlap" }], output), /overlapping/u);

  const passFile = path.join(fixture, "pass.test.mjs"), secondPassFile = path.join(fixture, "second-pass.test.mjs"), loadFile = path.join(fixture, "load.test.mjs"), teardownFile = path.join(fixture, "teardown.test.mjs"), unhandledFile = path.join(fixture, "unhandled.test.mjs");
  writeFileSync(passFile, `import { test } from ${JSON.stringify(api)}; test('ok', () => {});`);
  writeFileSync(secondPassFile, `import { test } from ${JSON.stringify(api)}; test('ok', () => {});`);
  const timed = runObserved("timed-files", request([[passFile, ["ok"]], [secondPassFile, ["ok"]]]));
  assert.equal(timed.result.status, 0, timed.result.stderr);
  for (const [activity, count] of [["runner_startup", 1], ["runner_cleanup", 1], ["file_setup", 2], ["file_execution", 2], ["file_teardown", 2]]) {
    const records = timed.activities.filter((entry) => entry.activity === activity);
    assert.equal(records.length, count, activity);
    assert.ok(records.every((entry) => entry.outcome === "passed" && entry.duration_ms >= 0), activity);
    assert.ok(records.every((entry) => !JSON.stringify(entry).includes("pass.test")), "diagnostics contain no file paths");
  }
  writeFileSync(loadFile, "throw new Error('suite collection failed');");
  writeFileSync(teardownFile, `import { test, afterAll } from ${JSON.stringify(api)}; test('ok', () => {}); afterAll(() => { throw new Error('file teardown failed'); });`);
  writeFileSync(unhandledFile, `import { test } from ${JSON.stringify(api)}; test('ok', async () => { void Promise.reject(new Error('unhandled runner error')); await new Promise(resolve => setTimeout(resolve, 25)); });`);
  const skippedFile = path.join(fixture, "skipped.test.mjs");
  writeFileSync(skippedFile, `import { test } from ${JSON.stringify(api)}; test.skip('selected', () => {});`);
  assert.equal(run("selected-skipped", request([[skippedFile, ["selected"]]])).rows[0].failure_reason, "scheduler_accounting_error");
  const excludedFile = path.join(fixture, "excluded.mjs");
  writeFileSync(excludedFile, "throw new Error('excluded file must never execute');");
  const excluded = prepare("excluded", request([[excludedFile, ["excluded"]]]));
  const excludedResult = spawnSync(process.execPath, excluded.args, { cwd: root, env: fixtureEnvironment, encoding: "utf8", timeout: 30000 });
  assert.equal(excludedResult.status, 11); assert.match(excludedResult.stderr, /excluded from its registered project/u);
  for (const file of [loadFile, teardownFile]) {
    const runResult = run(path.basename(file), request([[file, ["ok"]], [passFile, ["ok"]]]));
    assert.equal(runResult.result.status, 3, runResult.result.stderr);
    assert.equal(runResult.rows[0].failure_class, "infra"); assert.equal(runResult.rows[1].terminal_state, "passed");
  }
  const unhandled = run("unhandled", request([[passFile, ["ok"]], [unhandledFile, ["ok"]]]));
  assert.ok(unhandled.rows.every((row) => row.failure_class === "infra"));
  const globalSetup = path.join(fixture, "global-setup.mjs");
  const globalMarker = path.join(fixture, "global-setup-ran");
  writeFileSync(globalSetup, `import { writeFileSync } from 'node:fs'; export default function () { writeFileSync(${JSON.stringify(globalMarker)}, 'setup'); return () => { writeFileSync(${JSON.stringify(globalMarker)}, 'teardown'); throw new Error('global teardown failed'); }; }`);
  config({ globalSetup });
  const globalFailure = run("global-teardown", request([[passFile, ["ok"]], [secondPassFile, ["ok"]]]));
  assert.equal(readFileSync(globalMarker, "utf8"), "teardown", globalFailure.result.stderr);
  assert.equal(globalFailure.rows[0].failure_class, "infra", globalFailure.result.stderr + globalFailure.result.stdout);
  assert.equal(globalFailure.rows[1].failure_class, "infra", "global teardown supersedes every settled file");
  assert.ok(globalFailure.details.failures.some((failure) => failure.message.includes("global teardown failed")));
  config();

  const marker = path.join(fixture, "running"), hangFile = path.join(fixture, "hang.test.mjs");
  writeFileSync(hangFile, `import { test } from ${JSON.stringify(api)}; import { writeFileSync } from 'node:fs'; test('hang', async () => { console.log('interruption stdout'); console.error('interruption stderr'); writeFileSync(${JSON.stringify(marker)}, 'running'); await new Promise(() => {}); }, 10000);`);
  for (const signal of ["SIGTERM", "SIGKILL"]) {
    rmSync(marker, { force: true });
    const interrupted = prepare(`interrupted-${signal}`, request([[passFile, ["ok"]], [secondPassFile, ["ok"]], [hangFile, ["hang"]]]));
    const child = spawn(process.execPath, interrupted.args, { cwd: root, env: fixtureEnvironment, stdio: "ignore" });
    const done = new Promise((resolve) => child.once("close", (status, signal) => resolve({ status, signal })));
    try { await waitForFile(marker, child); child.kill(signal); }
    finally { const timer = setTimeout(() => child.kill("SIGKILL"), 1500); await done; clearTimeout(timer); }
    const rows = adaptVitestInvocationFile(interrupted, { ...(await done), signal });
    assert.equal(rows[0].terminal_state, "passed", "settled file survives invocation interruption/owner loss");
    assert.equal(rows[1].terminal_state, "passed", "a second settled file also survives interruption");
    assert.equal(rows[2].terminal_state, "cancelled");
    const checkpoint = readVitestEvidence({ root, request: interrupted.request, reportFile: interrupted.runnerJSON,
      detailsFile: interrupted.detailsFile, runID: process.env.CARTULARY_TEST_RUN_ID }).details;
    assert.equal(checkpoint.final_complete, false);
    assert.equal(checkpoint.completed_files.length, 2);
  }

  // Loss during global close, after every file settled, preserves row evidence
  // while the enclosing invocation still reports interruption.
  const closeMarker = path.join(fixture, "closing"), hangingClose = path.join(fixture, "hanging-close.mjs");
  writeFileSync(hangingClose, `import { writeFileSync } from 'node:fs'; export default function() { return async () => { writeFileSync(${JSON.stringify(closeMarker)}, 'closing'); await new Promise(() => { setInterval(() => {}, 1000); }); }; }`);
  config({ globalSetup: hangingClose });
  const completedFiles = prepare("interrupted-after-files", request([[passFile, ["ok"]], [secondPassFile, ["ok"]]]));
  const closingChild = spawn(process.execPath, completedFiles.args, { cwd: root, env: fixtureEnvironment, stdio: "ignore" });
  const closingDone = new Promise((resolve) => closingChild.once("close", (status, signal) => resolve({ status, signal })));
  try { await waitForFile(closeMarker, closingChild); }
  finally { closingChild.kill("SIGKILL"); await closingDone; }
  const closingExecution = await closingDone;
  const settledRows = adaptVitestInvocationFile(completedFiles, closingExecution);
  assert.ok(settledRows.every((row) => row.terminal_state === "passed"));
  assert.equal(closingExecution.signal, "SIGKILL", JSON.stringify(closingExecution));
  assert.equal(vitestInvocationFailure(closingExecution, settledRows).failure_class, "interrupted");
  assert.equal(vitestInvocationFailure({ status: 11 }, settledRows).failure_reason, "scheduler_accounting_error");
  assert.equal(vitestInvocationFailure({ status: 0 }, settledRows), null);
  config();

  // A passed assertion is insufficient while environment teardown remains open.
  const environment = path.join(fixture, "environment.mjs"), teardownMarker = path.join(fixture, "environment-teardown");
  const environmentFailure = path.join(fixture, "environment-failure.test.mjs");
  writeFileSync(environmentFailure, `import { test } from ${JSON.stringify(api)}; test('ok', () => { globalThis.failEnvironmentTeardown = true; });`);
  writeFileSync(environment, "export default { name: 'fixture', viteEnvironment: 'ssr', setup(global) { return { teardown() { if (global.failEnvironmentTeardown) throw new Error('environment teardown failed'); } }; } };");
  config({ environment });
  const failedEnvironment = runObserved("failed-environment", request([[environmentFailure, ["ok"]], [passFile, ["ok"]]]));
  assert.deepEqual(failedEnvironment.activities.filter((entry) => entry.activity === "file_teardown").map((entry) => entry.outcome).sort(), ["failed", "passed"]);
  assert.equal(failedEnvironment.rows[0].failure_class, "infra", failedEnvironment.result.stderr);
  assert.equal(failedEnvironment.rows[1].terminal_state, "passed");
  writeFileSync(environment, `import { writeFileSync } from 'node:fs'; export default { name: 'fixture', viteEnvironment: 'ssr', setup() { return { async teardown() { writeFileSync(${JSON.stringify(teardownMarker)}, 'pending'); await new Promise(() => {}); } }; } };`);
  config({ environment });
  const unsettled = prepare("unsettled", request([[passFile, ["ok"]]]));
  const child = spawn(process.execPath, unsettled.args, { cwd: root, env: fixtureEnvironment, stdio: "ignore" });
  const done = new Promise((resolve) => child.once("close", resolve));
  try { await waitForFile(teardownMarker, child); }
  finally { child.kill("SIGKILL"); await done; }
  assert.equal(adaptVitestInvocationFile(unsettled, { signal: "SIGKILL" })[0].terminal_state, "cancelled");
  assert.equal(existsSync(unsettled.detailsFile), false);
  config();

  const wrapperEnvironment = { ...fixtureEnvironment, NODE_BIN: process.execPath, CARTULARY_OUTPUT_MODE: "quiet" };
  delete wrapperEnvironment.CARTULARY_HARNESS_IDENTITY_PREPARED;
  delete wrapperEnvironment.CARTULARY_TEST_TARGET;
  const wrapperPath = path.join(root, "tools/harness/execution/run-vitest-step.sh");
  const label = `vitest-causes-${randomUUID()}`;
  const wrapper = spawnSync("bash", [wrapperPath, label, "--", invocation.requestFile],
    { cwd: root, env: wrapperEnvironment, encoding: "utf8", timeout: 30000 });
  assert.equal(wrapper.status, full.result.status, wrapper.stderr);
  const summary = readVitestJSON(path.join(runRoot, "adhoc", label, "step-summary.json"));
  for (const kind of ["timing", "product", "unknown"]) assert.ok(summary.dossiers.some((item) => item.failure_class === kind));
  assert.ok(summary.dossiers.some((item) => item.message.includes("precise original assertion")));
  const pass = prepare("wrapper-pass", request([[passFile, ["ok"]]]));
  const success = spawnSync("bash", [wrapperPath, `vitest-pass-${randomUUID()}`, "--", pass.requestFile],
    { cwd: root, env: { ...wrapperEnvironment, CARTULARY_SUPPRESS_CHILD_SUCCESS: "1" }, encoding: "utf8", timeout: 30000 });
  assert.equal(success.status, 0, success.stderr);
  assert.equal(success.stdout, "");
  const timeoutLabel = `vitest-watchdog-${randomUUID()}`;
  const hung = prepare("wrapper-hang", request([[passFile, ["ok"]], [hangFile, ["hang"]]]));
  const timeout = spawnSync("bash", [wrapperPath, timeoutLabel, "--", hung.requestFile],
    { cwd: root, env: { ...wrapperEnvironment, CARTULARY_VITEST_WATCHDOG_SECONDS: "1", CARTULARY_WATCHDOG_KILL_GRACE_SECONDS: "1" }, encoding: "utf8", timeout: 15000 });
  assert.notEqual(timeout.status, 0, timeout.stderr);
  const timeoutSummary = readVitestJSON(path.join(runRoot, "adhoc", timeoutLabel, "step-summary.json"));
  assert.equal(timeoutSummary.failure_reason, "timeout_failure");
  assert.ok(timeoutSummary.artifacts.watchdog_json);
  rmSync(marker, { force: true });
  const interruptLabel = `vitest-wrapper-interrupt-${randomUUID()}`;
  const wrapperChild = spawn("bash", [wrapperPath, interruptLabel, "--", hung.requestFile],
    { cwd: root, env: { ...wrapperEnvironment, CARTULARY_VITEST_WATCHDOG_SECONDS: "30", CARTULARY_WATCHDOG_KILL_GRACE_SECONDS: "1" }, stdio: "ignore" });
  const wrapperDone = new Promise((resolve) => wrapperChild.once("close", (status, signal) => resolve({ status, signal })));
  try { await waitForFile(marker, wrapperChild); wrapperChild.kill("SIGTERM"); }
  finally { const timer = setTimeout(() => wrapperChild.kill("SIGKILL"), 10000); await wrapperDone; clearTimeout(timer); }
  assert.equal((await wrapperDone).status, 143);
  const interruption = readVitestJSON(path.join(runRoot, "adhoc", interruptLabel, "step-summary.json"));
  assert.equal(interruption.failure_reason, "cancelled_or_interrupted");
  assert.ok(interruption.inventory.some((item) => item.symbol_or_title === "ok"), "output consumer retains settled inventory");
  assert.match(readFileSync(path.resolve(root, interruption.artifacts.stdout_log), "utf8"), /interruption stdout/u);
  assert.match(readFileSync(path.resolve(root, interruption.artifacts.stderr_log), "utf8"), /interruption stderr/u);
  process.stdout.write("Vitest exact cohorts, file settlement, diagnostic attribution, interruption and privacy contracts passed\n");
} finally {
  rmSync(groupedFixture, { recursive: true, force: true });
  rmSync(fixture, { recursive: true, force: true });
  ownedRuntime?.close();
}
