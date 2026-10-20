import { publishRendererFixture } from "./renderer-evidence-fixture.mjs";
import assert from "node:assert/strict";
import process from "node:process";
import { spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, writeFileSync, rmSync, existsSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { pathToFileURL } from "node:url";
import { resolvePlaywrightPackages } from "../../readiness/playwright-packages.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";

import {
  publicExitCodeForFailure,
} from "../../contract/failure-taxonomy.mjs";
import {
  adaptPlaywrightReport,
  playwrightTestList,
  playwrightGroupExitCode,
  playwrightReportMaximumBytes,
} from "../../execution/runners/playwright.mjs";

const file = "apps/web/e2e/keyboard.spec.ts";

function row(rowID, ...titles) {
  return {
    row_id: rowID,
    selector: { file, titles, project_id: "chromium" },
  };
}

function spec(title, status, duration = 7) {
  return {
    file,
    title,
    tests: [
      {
        projectName: "chromium",
        results: [{ duration, status }],
      },
    ],
  };
}

function report(...specs) {
  return {
    suites: [{ file, specs }],
  };
}

function adapt(singleRow, playwrightReport, processStatus = 0, processSignal = null) {
  return adaptPlaywrightReport(
    [singleRow],
    playwrightReport,
    processStatus,
    processSignal,
    singleRow.selector.titles.map(title => ({ row_id: singleRow.row_id, file, project_id: "chromium", title_path: [title] })),
  )[0];
}

test("Playwright row adapter distinguishes primary asset failure from secondary cleanup and product failure", () => {
  const broken = spec("assets", "failed");
  const result = broken.tests[0].results[0];
  result.errors = [
    { message: "CartularyFrontendArtifactError: Required frontend assets failed" },
    { message: "Error: browser context has closed during cleanup" },
  ];
  const assetResult = adapt(row("harness.browser.boundary_support.asset_probe", "assets"), report(broken));
  assert.equal(assetResult.failure_class, "artifact");
  const group = {
    schema_id: "cartulary.browser_group_result.v11", renderer_borrow: { unit_id:"fixture:group", borrow_id:"11111111-1111-4111-8111-111111111111", allocation_id:"22222222-2222-4222-8222-222222222222", generation:"33333333-3333-4333-8333-333333333333", attestation_ref:"_shared/renderers/22222222-2222-4222-8222-222222222222/renderer-profile-attestation.json", trust_ref:"_shared/renderers/22222222-2222-4222-8222-222222222222/tls-trust-attestation.json" }, target_id: "browser-e2e-visual",
    stage_id: "visual", group_id: "asset-probe", browser_session_id: "session",
    runtime_profile_id: "default", service_requirement: "test-services",
    fixture_capabilities: ["browser_stack"], service_dependencies: ["postgres"],
    resource_profile_ids: ["browser_isolated"], selected_rows: ["harness.browser.boundary_support.asset_probe"],
    started_at: "2026-09-09T00:00:00Z", finished_at: "2026-09-09T00:00:01Z",
    duration_ms: 1000, status: "fail", exit_code: 11, row_results: [assetResult],
    session_artifacts: ["stack_v7", "startup_diagnostics_v2"].map((kind) => ({
      kind, ref: `${kind}.json`, sha256: `sha256:${"0".repeat(64)}`,
    })),
    artifacts: { playwright_report: "report.json", stdout: "stdout.log", stderr: "stderr.log" },
  };
  validateSchemaSync(group.schema_id, group);
  assert.throws(() => validateSchemaSync(group.schema_id, {
    ...group, session_artifacts: group.session_artifacts.map((entry) => ({ ...entry, kind: "stack_v5" })),
  }));
  result.errors.unshift({ message: "Error: expected product behavior" });
  assert.equal(adapt(row("harness.browser.boundary_support.asset_probe", "assets"), report(broken)).failure_class, "product");
});

test("Playwright row adapter preserves the closed status and exit matrix", () => {
  const cases = [
    {
      status: "passed",
      expected: ["passed", 0, null, null],
    },
    {
      status: "failed",
      expected: ["failed", 10, "product", "test_assertion_failure"],
    },
    {
      status: "timedOut",
      expected: ["failed", 10, "product", "test_assertion_failure"],
    },
    {
      status: "skipped",
      expected: [
        "infrastructure_failed",
        11,
        "harness",
        "scheduler_accounting_error",
      ],
    },
  ];

  for (const fixture of cases) {
    const result = adapt(
      row(`harness.test_catalog.unit.playwright_${fixture.status}`, fixture.status),
      report(spec(fixture.status, fixture.status)),
    );
    assert.deepEqual(
      [
        result.terminal_state,
        result.exit_code,
        result.failure_class,
        result.failure_reason,
      ],
      fixture.expected,
      fixture.status,
    );
    assert.equal(result.duration_ms, 7, fixture.status);
  }
});

test("Playwright row adapter treats runner interruption as cancellation", () => {
  const interrupted = adapt(
    row("harness.test_catalog.unit.playwright_interrupted", "interrupted"),
    report(spec("interrupted", "interrupted")),
    130,
  );
  assert.deepEqual(
    [
      interrupted.terminal_state,
      interrupted.exit_code,
      interrupted.failure_class,
      interrupted.failure_reason,
    ],
    ["cancelled", 130, "interrupted", "cancelled_or_interrupted"],
  );
});

test("Playwright row adapter classifies missing and ambiguous selectors as accounting failures", () => {
  const missing = adapt(
    row("harness.test_catalog.unit.playwright_missing", "missing"),
    report(),
  );
  assert.deepEqual(
    [missing.terminal_state, missing.exit_code, missing.failure_class, missing.failure_reason],
    ["infrastructure_failed", 11, "harness", "scheduler_accounting_error"],
  );

  const ambiguous = adapt(
    row("harness.test_catalog.unit.playwright_ambiguous", "ambiguous"),
    report(spec("ambiguous", "passed"), spec("ambiguous", "passed")),
  );
  assert.deepEqual(
    [
      ambiguous.terminal_state,
      ambiguous.exit_code,
      ambiguous.failure_class,
      ambiguous.failure_reason,
    ],
    ["infrastructure_failed", 11, "harness", "scheduler_accounting_error"],
  );
});

test("Playwright row adapter applies primary-failure precedence within a row", () => {
  const mixed = adapt(
    row(
      "harness.test_catalog.unit.playwright_mixed",
      "product failure",
      "missing observation",
    ),
    report(spec("product failure", "failed")),
    1,
  );
  assert.deepEqual(
    [mixed.terminal_state, mixed.exit_code, mixed.failure_class, mixed.failure_reason],
    ["failed", 10, "product", "test_assertion_failure"],
  );
});

test("Playwright row adapter rejects a passing report from a nonzero child", () => {
  const mismatch = adapt(
    row("harness.test_catalog.unit.playwright_child_mismatch", "passed"),
    report(spec("passed", "passed")),
    1,
  );
  assert.deepEqual(
    [mismatch.terminal_state, mismatch.exit_code, mismatch.failure_class, mismatch.failure_reason],
    ["infrastructure_failed", 11, "harness", "scheduler_accounting_error"],
  );
});

test("Playwright group exit uses canonical product-first failure precedence", () => {
  const product = adapt(
    row("harness.test_catalog.unit.playwright_group_product", "failed"),
    report(spec("failed", "failed")),
    1,
  );
  const accounting = adapt(
    row("harness.test_catalog.unit.playwright_group_accounting", "missing"),
    report(),
    1,
  );
  assert.equal(
    playwrightGroupExitCode([accounting, product], { signal: null, status: 1 }),
    10,
  );
});

test("browser fixture process retains redacted late diagnostics and preserves cleanup classification", async () => {
  const { createBrowserFixtureProcess, fixtureLifecycleAttachment } = await import("../index.mjs");
  const fixture = createBrowserFixtureProcess({
    command: process.execPath,
    args: ["-e", `
      const attempt = process.argv.at(-1);
      const cleanupBudget = Number(process.argv[process.argv.indexOf('--cleanup-budget-ms')+1]);
      process.stdout.write(JSON.stringify({ cleanup_budget_ms: cleanupBudget, user_password: 'private-fixture-value' }) + '\\n');
      process.stdin.resume();
      process.stdin.on('end', () => {
        process.stderr.write('late private-fixture-');
        process.stderr.write('value postgres://user:secret@localhost/db\\n');
        process.stderr.write(JSON.stringify({ schema_id: 'cartulary.browser_fixture_event.v3', attempt_id: attempt,
          stage: 'target.database', phase: 'cleanup', outcome: 'failed', elapsed_ms: 1, deadline_ms: 100,
          postgres: { duration_ms: 100, admission_ms: 1, service_generation: null, admission_budget_ms: 100, operation_ms: 99, operation_budget_ms: 99, admission: 'acquired', outcome: 'deadline_exceeded', wait: { status: 'not_observed', first: null, last: null } },
          message: 'private-fixture-value: catalog retirement failed' }) + '\\n');
        process.exitCode = 1;
      });
    `, "--"],
  });
  assert.equal((await fixture.ready).cleanup_budget_ms, 30_000);
  const stopped = fixture.stop();
  assert.equal(fixture.stop(), stopped);
  let cleanup;
  try { await stopped; } catch (error) { cleanup = error; }
  assert.match(cleanup.message, /target.database/);
  assert.match(cleanup.message, /late/);
  assert.doesNotMatch(cleanup.message, /private-fixture-value|user:secret/);
  const failures = [{ phase: "cleanup", failure_class: "harness", failure_reason: "cleanup_error", error: cleanup }];
  const attach = () => ({ name: fixtureLifecycleAttachment, contentType: "application/json", body: Buffer.from(JSON.stringify(fixture.report(failures))).toString("base64") });
  const broken = spec("cleanup", "failed");
  broken.tests[0].results[0].attachments = [attach()];
  let rowResult = adapt(row("harness.browser.unit.fixture_cleanup", "cleanup"), report(broken), 1);
  assert.equal(rowResult.failure_class, "harness");
  assert.equal(rowResult.failure_reason, "cleanup_error");
  assert.equal(rowResult.exit_code, 12);
  assert.equal(rowResult.failure_diagnostic.process.closed, true);
  assert.equal(rowResult.failure_diagnostic.events[0].postgres.operation_budget_ms, 99);
  failures.unshift({ phase: "body", failure_class: "product", failure_reason: "test_assertion_failure", error: new Error("original assertion") });
  broken.tests[0].results[0].attachments = [attach()];
  rowResult = adapt(row("harness.browser.unit.fixture_cleanup", "cleanup"), report(broken), 1);
  assert.equal(rowResult.failure_diagnostic.failures.length, 2);
  assert.equal(rowResult.failure_diagnostic.failures[0].message, "original assertion");
  assert.equal(rowResult.exit_code, 10);
  failures[0] = { phase: "body", failure_class: "interrupted", failure_reason: "cancelled_or_interrupted", error: new Error("test interrupted") };
  broken.tests[0].results[0].status = "interrupted";
  broken.tests[0].results[0].attachments = [attach()];
  rowResult = adapt(row("harness.browser.unit.fixture_cleanup", "cleanup"), report(broken), 130);
  assert.equal(rowResult.failure_diagnostic.failures[0].failure_class, "interrupted");
  assert.equal(rowResult.failure_diagnostic.failures[1].failure_reason, "cleanup_error");
  const invalid = attach(); invalid.body = Buffer.from('{}').toString('base64');
  broken.tests[0].results[0].attachments = [invalid];
  assert.equal(adapt(row("harness.browser.unit.fixture_cleanup", "cleanup"), report(broken)).failure_class, "artifact");
});

test("browser fixture forced retirement remains a failure even when SIGTERM exits successfully", async () => {
  const { createBrowserFixtureProcess } = await import("../index.mjs");
  const fixture = createBrowserFixtureProcess({
    command: process.execPath,
    args: ["-e", `process.on('SIGTERM', () => process.exit(0)); process.stdout.write('{}\\n'); setInterval(() => {}, 1000);`, "--"],
    cleanupMs: 20, signalMs: 1000,
  });
  await fixture.ready;
  await assert.rejects(fixture.stop(), /forced=true/);
  assert.equal(fixture.report().process.forced, true);
  assert.equal(fixture.report().process.closed, true);
});

test("browser fixture stderr truncation cannot expose a multiline credential suffix", async () => {
  const { createBrowserFixtureProcess } = await import("../index.mjs");
  const fixture = createBrowserFixtureProcess({ command: process.execPath,
    env: { ...process.env, FIXTURE_PASSWORD: `${"private-head-".repeat(900)}\nprivate-tail-value` },
    args: ["-e", `
      process.stdout.write('{}\\n'); process.stdin.resume();
      process.stdin.on('end', () => {
        process.stderr.write(process.env.FIXTURE_PASSWORD + '\\n' + 'safe'.repeat(7000) + '\\n');
        process.stderr.write(JSON.stringify({ schema_id: 'cartulary.browser_fixture_event.v3', attempt_id: process.argv.at(-1),
          stage: 'fixture', phase: 'terminal', outcome: 'succeeded', elapsed_ms: 1, deadline_ms: 0, message: '' }) + '\\n');
      });
    `, "--"],
  });
  await fixture.ready;
  await fixture.stop();
  assert.equal(/private-head-|private-tail-value/u.test(fixture.report().stderr), false);
  assert.match(fixture.report().stderr, /safe/);
});

test("browser fixture startup failure and absent terminal evidence fail with redacted diagnostics", async () => {
  const { createBrowserFixtureProcess } = await import("../index.mjs");
  const startup = createBrowserFixtureProcess({
    command: process.execPath,
    args: ["-e", `process.stderr.write(process.env.FIXTURE_PASSWORD + ': startup failed\\n'); process.exitCode = 1;`, "--"],
    env: { ...process.env, FIXTURE_PASSWORD: "startup-private-value" },
  });
  await assert.rejects(startup.ready, (error) => {
    assert.match(error.message, /startup failed/);
    assert.doesNotMatch(error.message, /startup-private-value/);
    return true;
  });
  await assert.rejects(startup.stop(), /startup failed/);
  const incomplete = createBrowserFixtureProcess({ command: process.execPath,
    args: ["-e", `process.stdout.write('{}\\n'); process.stdin.resume();`, "--"] });
  await incomplete.ready;
  await assert.rejects(incomplete.stop(), /Missing terminal fixture evidence/);
});

test("browser fixture retirement rejects an unanchored surviving group", async () => {
  const { createBrowserFixtureProcess } = await import("../index.mjs");
  const descendant = "process.on('SIGTERM', () => {}); process.stdout.write(JSON.stringify({pid:process.pid}) + String.fromCharCode(10)); setInterval(() => {}, 1000);";
  const fixture = createBrowserFixtureProcess({ command: process.execPath,
    args: ["-e", `
      const { spawn } = require('node:child_process');
      spawn(process.execPath, ['-e', ${JSON.stringify(descendant)}], { stdio: ['ignore', 'inherit', 'inherit'] });
      process.stdin.resume(); process.stdin.on('end', () => process.exit(0));
    `, "--"], startupMs: 5000, cleanupMs: 30, signalMs: 100,
  });
  let descendantPID;
  try {
    descendantPID = (await fixture.ready).pid;
    await assert.rejects(fixture.stop(), /Unresolved browser fixture process settlement/);
    assert.equal(fixture.report().process.closed, false);
    assert.ok(process.kill(descendantPID, 0), "unanchored identity must not be signaled");
  } finally {
    // This fixture published its exact live child; dispose it independently of
    // the deliberately unavailable production group anchor.
    if (descendantPID) process.kill(descendantPID, "SIGKILL");
    await fixture.stop().catch(() => {});
  }
});


test("browser target finalization preserves typed row failures and cleanup precedence", () => {
  const scratch = mkdtempSync(path.join(os.tmpdir(), "browser-target-classification-"));
  const root = path.resolve(import.meta.dirname, "../../../..");
  const target = "browser-e2e-webserver-backed";
  const cleanup = { failure_class: "harness", failure_reason: "cleanup_error", exit_code: 12 };
  const product = { failure_class: "product", failure_reason: "test_assertion_failure", exit_code: 10 };
  try {
    for (const [id, failures, expected] of [
      ["cleanup", [cleanup], 12],
      ["product_and_cleanup", [product, cleanup], 10],
      ["artifact", [{ failure_class: "artifact", failure_reason: "artifact_error", exit_code: 11 }], 11],
      ["infra", [{ failure_class: "infra", failure_reason: "service_start_error", exit_code: 3 }], 3],
      ["group_cleanup", [null], 12],
    ]) {
      const groups = failures.map((failure, index) => {
        const groupID = `group-${index}`;
        const rowID = `harness.browser.boundary_support.probe_${index}`;
        const directory = path.join(scratch, id, target, "browser-groups", groupID);
        mkdirSync(directory, { recursive: true, mode: 0o700 });
        const group = {
          schema_id: "cartulary.browser_group_result.v11", renderer_borrow: { unit_id:"fixture:group", borrow_id:"11111111-1111-4111-8111-111111111111", allocation_id:"22222222-2222-4222-8222-222222222222", generation:"33333333-3333-4333-8333-333333333333", attestation_ref:"_shared/renderers/22222222-2222-4222-8222-222222222222/renderer-profile-attestation.json", trust_ref:"_shared/renderers/22222222-2222-4222-8222-222222222222/tls-trust-attestation.json" }, target_id: target,
          stage_id: "webserver-backed", group_id: groupID, browser_session_id: `session-${index}`,
          runtime_profile_id: "default", service_requirement: "test-services",
          fixture_capabilities: ["browser_stack"], service_dependencies: ["postgres"],
          resource_profile_ids: ["browser_isolated"], selected_rows: [rowID],
          started_at: "2026-10-13T00:00:00Z", finished_at: "2026-10-13T00:00:01Z",
          duration_ms: 1000, status: "fail", exit_code: failure?.exit_code ?? 12,
          row_results: [{ row_id: rowID, terminal_state: failure ? "failed" : "passed", duration_ms: 1000,
            exit_code: failure?.exit_code ?? 0, failure_class: failure?.failure_class ?? null,
            failure_reason: failure?.failure_reason ?? null, failure_diagnostic: null }],
          session_artifacts: ["stack_v7", "startup_diagnostics_v2"].map(kind => ({
            kind, ref: `${kind}.json`, sha256: `sha256:${"0".repeat(64)}`,
          })), artifacts: { playwright_report: "report.json", stdout: "stdout.log", stderr: "stderr.log" },
        };
        group.renderer_borrow.borrow_id = `11111111-1111-4111-8111-${String(index).padStart(12,"0")}`;
        group.renderer_borrow.unit_id = `fixture:${groupID}`;
        publishRendererFixture(path.join(scratch,id),group,id);
        validateSchemaSync(group.schema_id, group);
        writeFileSync(path.join(directory, "browser-group-result.json"), JSON.stringify(group), { mode: 0o600 });
        return groupID;
      });
      const finalizerEnv = { ...process.env, CARTULARY_TEST_RESULTS_DIR: scratch, CARTULARY_TEST_RUN_ID: id };
      delete finalizerEnv.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
      delete finalizerEnv.CARTULARY_HARNESS_UNIT_EVENT_PREFIX;
      const result = spawnSync(process.execPath, [path.join(root, "tools/harness/browser/browser-target-finalizer-cli.mjs"),
        "--target", target, "--groups", groups.join(","), "--group-targets", groups.map(group => `${group}=${target}`).join(","),
        "--reset-prefix", `browser_reset:${target}:`], {
        cwd: root, encoding: "utf8", env: finalizerEnv,
      });
      assert.equal(result.status, expected, `${id}: ${result.stderr}`);
    }
  } finally { rmSync(scratch, { recursive: true, force: true }); }
});


test("Playwright selection and accounting bind exact project, file and suite path", () => {
  const selected = row("harness.browser.unit.identity", "same");
  const selection = [{ row_id: selected.row_id, file, project_id: "chromium", title_path: ["outer", "same"] }];
  assert.equal(playwrightTestList(selection), "[chromium] › keyboard.spec.ts › outer › same\n");
  const value = { suites: [{ file, suites: [{ title: "outer", specs: [spec("same", "passed")] }] }] };
  const run = input => adaptPlaywrightReport([selected], input, 0, null, selection)[0];
  assert.equal(run(value).terminal_state, "passed");
  for (const mutate of [
    input => { input.suites[0].suites[0].title = "another"; },
    input => { input.suites[0].suites[0].specs[0].tests[0].projectName = "other"; },
    input => { input.suites[0].suites[0].specs[0].file = "apps/web/e2e/other.spec.ts"; },
    input => { input.suites[0].suites[0].specs.push(spec("same suffix", "passed")); },
    input => { input.suites[0].suites.push({ title: "other", specs: [spec("same", "passed")] }); },
  ]) {
    const invalid = structuredClone(value); mutate(invalid);
    assert.equal(run(invalid).failure_reason, "scheduler_accounting_error");
  }
  assert.throws(() => adaptPlaywrightReport([selected], value), /resolved selection/u);
  assert.throws(() => playwrightTestList([{ ...selection[0], title_path: ["bad\nname"] }]), /test-list/u);
});


test("pinned Playwright test-list executes only an exact nested tagged identity", () => {
  const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-playwright-selection-"));
  try {
    const root = path.resolve(import.meta.dirname, "../../../..");
    const packages = resolvePlaywrightPackages(root);
    const testModule = pathToFileURL(path.join(packages.playwrightPath, "test.mjs")).href;
    writeFileSync(path.join(scratch, "selection.spec.mjs"), `
      import { test } from ${JSON.stringify(testModule)};
      test("target", () => { throw new Error("unselected root test executed"); });
      test.describe("outer", () => {
        test("target", { tag: "@fast" }, () => {});
        test("target suffix", () => { throw new Error("unselected suffix test executed"); });
      });
    `);
    writeFileSync(path.join(scratch, "config.mjs"), `export default ${JSON.stringify({ testDir: scratch, projects: [{ name: "chromium" }], retries: 0, workers: 1 })};`);
    const selected = { row_id: "fixture.selection", selector: { file: "apps/web/e2e/selection.spec.mjs", project_id: "chromium", titles: ["outer target"] } };
    const selection = [{ row_id: selected.row_id, project_id: "chromium", file: selected.selector.file, title_path: ["outer", "target"] }];
    const selectionFile = path.join(scratch, "selection.txt");
    writeFileSync(selectionFile, playwrightTestList(selection));
    const child = spawnSync(process.execPath, [path.join(packages.playwrightPath, "cli.js"), "test", "--config", path.join(scratch, "config.mjs"), "--test-list", selectionFile, "--reporter=json"], { cwd: scratch, encoding: "utf8", timeout: 30000, maxBuffer: 2 * 1024 * 1024 });
    assert.equal(child.status, 0, child.stderr);
    const result = JSON.parse(child.stdout);
    assert.equal(result.stats.expected, 1);
    assert.equal(result.stats.unexpected, 0);
    assert.equal(adaptPlaywrightReport([selected], result, child.status, child.signal, selection)[0].terminal_state, "passed");
  } finally { rmSync(scratch, { recursive: true, force: true }); }
});

test("cohort accounting preserves settled siblings through interruption and invocation cleanup failure", () => {
  const rows = [row("harness.browser.unit.first", "first"), row("harness.browser.unit.second", "second")];
  const selection = rows.map((row) => ({ row_id: row.row_id, file, project_id: "chromium", title_path: row.selector.titles }));
  const partial = report(spec("first", "passed"));
  assert.deepEqual(adaptPlaywrightReport(rows, partial, null, "SIGTERM", selection).map((row) => row.terminal_state), ["passed", "cancelled"]);
  const failed = report(spec("first", "failed"), spec("second", "skipped"));
  assert.deepEqual(adaptPlaywrightReport(rows, failed, 1, null, selection).map((row) => row.terminal_state), ["failed", "skipped_dependency"]);
  const globalFailure = { ...report(spec("first", "passed"), spec("second", "passed")), errors: [{ message: "worker teardown failed" }] };
  const settled = adaptPlaywrightReport(rows, globalFailure, 1, null, selection);
  assert.ok(settled.every((row) => row.terminal_state === "passed"));
  assert.equal(playwrightGroupExitCode(settled, { status: 1 }, globalFailure), 1);
});

test("public Playwright reporter replaces large checkpoints after cross-file test cleanup", async () => {
  const { readFileSync } = await import("node:fs");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-playwright-cohort-"));
  try {
    const root = path.resolve(import.meta.dirname, "../../../..");
    const packages = resolvePlaywrightPackages(root);
    const testModule = pathToFileURL(path.join(packages.playwrightPath, "test.mjs")).href;
    const testDir = path.join(scratch, "apps/web/e2e"); mkdirSync(testDir, { recursive: true });
    const rows = ["a", "b"].map((name) => ({ row_id: `harness.browser.unit.${name}`, selector: {
      file: `apps/web/e2e/${name}.spec.mjs`, project_id: "chromium", titles: ["outer nested same"] } }));
    const selection = rows.map((row) => ({ row_id: row.row_id, file: row.selector.file, project_id: "chromium", title_path: ["outer", "nested", "same"] }));
    const checkpointFile = path.join(scratch, "checkpoint.json");
    const reporterFile = path.join(scratch, "reporter.mjs");
    writeFileSync(reporterFile, `import { PlaywrightCohortCollector } from ${JSON.stringify(pathToFileURL(path.join(root, "tools/harness/browser/playwright-cohort-reporter.mjs")).href)};
      import { existsSync } from 'node:fs';
      export default class extends PlaywrightCohortCollector {
        constructor() { super(${JSON.stringify({ rows, selection, runRoot: scratch, checkpointFile })}); }
        onTestEnd(test, result) {
          if (!existsSync(test.location.file + '.settled')) throw new Error('test teardown did not settle');
          super.onTestEnd(test, result);
        }
      }`);
    for (const name of ["a", "b"]) writeFileSync(path.join(testDir, `${name}.spec.mjs`), `
      import { test as base } from ${JSON.stringify(testModule)};
      import { writeFileSync } from 'node:fs';
      const test = base.extend({ owned: async ({}, use, info) => {
        await use({ id: info.testId });
        await info.attach('large-diagnostic', { body: Buffer.alloc(1100 * 1024, 1), contentType: 'application/octet-stream' });
        writeFileSync(info.file + '.settled', 'yes');
      } });
      test('same', () => { throw new Error('excluded root'); });
      test.describe('outer', () => test.describe('nested', () => {
        test('same', async ({ owned }) => { if (!owned.id) throw new Error('missing owned fixture'); });
        test('other', () => { throw new Error('excluded sibling'); });
      }));
    `);
    writeFileSync(path.join(scratch, "config.mjs"), `export default ${JSON.stringify({ testDir, projects: [{ name: "chromium" }, { name: "excluded" }], retries: 0, workers: 1, maxFailures: 1 })};`);
    const selectionFile = path.join(scratch, "selection.txt"); writeFileSync(selectionFile, playwrightTestList(selection));
    const child = spawnSync(process.execPath, [path.join(packages.playwrightPath, "cli.js"), "test", "--config", path.join(scratch, "config.mjs"), "--test-list", selectionFile, `--reporter=json,${reporterFile}`], { cwd: scratch, encoding: "utf8", timeout: 30000, maxBuffer: playwrightReportMaximumBytes });
    assert.equal(child.status, 0, child.stderr);
    assert.equal(JSON.parse(child.stdout).stats.expected, 2);
    const checkpointBytes = readFileSync(checkpointFile);
    assert.ok(checkpointBytes.length > 1024 * 1024);
    assert.ok(checkpointBytes.length < playwrightReportMaximumBytes);
    const checkpoint = JSON.parse(checkpointBytes.toString("utf8"));
    assert.deepEqual(checkpoint.errors, []);
    assert.deepEqual(adaptPlaywrightReport(rows, checkpoint, 0, null, selection).map((row) => row.terminal_state), ["passed", "passed"]);
    for (const row of rows) assert.equal(JSON.parse(readFileSync(path.join(scratch, "rows", `${row.row_id}.json`), "utf8")).terminal_state, "passed");
  } finally { rmSync(scratch, { recursive: true, force: true }); }
});

test("public Playwright checkpoints survive interruption and worker/global teardown failures", async () => {
  const { existsSync, readFileSync } = await import("node:fs");
  const { spawn } = await import("node:child_process");
  const { once } = await import("node:events");
  const root = path.resolve(import.meta.dirname, "../../../..");
  const packages = resolvePlaywrightPackages(root);
  const testModule = pathToFileURL(path.join(packages.playwrightPath, "test.mjs")).href;
  for (const mode of ["SIGTERM", "SIGKILL", "test_cleanup", "worker_cleanup", "global_cleanup"]) {
    const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-playwright-settlement-"));
    let child;
    try {
      const testDir = path.join(scratch, "apps/web/e2e"); mkdirSync(testDir, { recursive: true });
      const rows = ["first", "second"].map((title) => ({ row_id: `harness.browser.unit.${title}`, selector: {
        file: "apps/web/e2e/settlement.spec.mjs", project_id: "chromium", titles: [title] } }));
      const selection = rows.map((row) => ({ row_id: row.row_id, file: row.selector.file, project_id: "chromium", title_path: row.selector.titles }));
      const checkpointFile = path.join(scratch, "checkpoint.json"), marker = path.join(scratch, "second-started");
      const reporterFile = path.join(scratch, "reporter.mjs");
      writeFileSync(reporterFile, `import { PlaywrightCohortCollector } from ${JSON.stringify(pathToFileURL(path.join(root, "tools/harness/browser/playwright-cohort-reporter.mjs")).href)};
        export default class extends PlaywrightCohortCollector { constructor() { super(${JSON.stringify({ rows, selection, runRoot: scratch, checkpointFile })}); } }`);
      writeFileSync(path.join(testDir, "settlement.spec.mjs"), `import { test as base } from ${JSON.stringify(testModule)};
        import { writeFileSync } from 'node:fs';
        const test = base.extend({
          owned: async ({}, use, info) => { await use(true); if (${JSON.stringify(mode)} === 'test_cleanup' && info.title === 'second') throw new Error('test cleanup failed'); },
          provisioner: [async ({}, use) => { await use(true); if (${JSON.stringify(mode)} === 'worker_cleanup') throw new Error('worker cleanup failed'); }, { scope: 'worker' }],
        });
        test('first', async ({ owned, provisioner }) => {});
        test('second', async ({ owned, provisioner }) => { writeFileSync(${JSON.stringify(marker)}, 'started');
          if (${JSON.stringify(mode)}.startsWith('SIG')) await new Promise(resolve => setTimeout(resolve, 20000)); });`);
      const teardown = path.join(scratch, "teardown.mjs");
      writeFileSync(teardown, `export default () => { if (${JSON.stringify(mode)} === 'global_cleanup') throw new Error('global cleanup failed'); };`);
      const config = path.join(scratch, "config.mjs");
      writeFileSync(config, `export default ${JSON.stringify({ testDir, projects: [{ name: "chromium" }], retries: 0, workers: 1, maxFailures: 1, globalTeardown: teardown, reporter: [[reporterFile]] })};`);
      child = spawn(process.execPath, [path.join(packages.playwrightPath, "cli.js"), "test", "--config", config], { cwd: scratch, detached: true, stdio: ["ignore", "pipe", "pipe"] });
      let output = ""; child.stdout.on("data", (data) => { output += data; }); child.stderr.on("data", (data) => { output += data; });
      const finished = once(child, "close");
      if (mode.startsWith("SIG")) {
        const deadline = Date.now() + 15000;
        while ((!existsSync(marker) || !existsSync(path.join(scratch, "rows", `${rows[0].row_id}.json`))) && child.exitCode === null && Date.now() < deadline) await new Promise(resolve => setTimeout(resolve, 25));
        assert.ok(existsSync(marker) && existsSync(path.join(scratch, "rows", `${rows[0].row_id}.json`)), output);
        // Both coordinator and worker are test-owned members of this group.
        process.kill(-child.pid, mode);
      }
      const timer = setTimeout(() => { try { process.kill(-child.pid, "SIGKILL"); } catch {} }, 20000);
      const [status, signal] = await finished; clearTimeout(timer);
      const report = JSON.parse(readFileSync(checkpointFile, "utf8"));
      const results = adaptPlaywrightReport(rows, report, status, signal, selection);
      if (mode.startsWith("SIG")) {
        assert.equal(JSON.parse(readFileSync(path.join(scratch, "rows", `${rows[0].row_id}.json`))).terminal_state, "passed", output);
        assert.equal(results[0].terminal_state, "passed", output);
        assert.equal(results[1].terminal_state, "cancelled", output);
      } else if (mode === "test_cleanup") {
        assert.deepEqual(results.map((row) => row.terminal_state), ["passed", "failed"], output);
      } else {
        assert.ok(results.every((row) => row.terminal_state === "passed"), output);
        assert.equal(playwrightGroupExitCode(results, { status, signal }, report), 1);
        assert.equal(JSON.parse(readFileSync(path.join(scratch, "rows", `${rows[0].row_id}.json`))).terminal_state, "passed");
      }
    } finally {
      if (child?.pid) { try { process.kill(-child.pid, "SIGKILL"); } catch {} }
      rmSync(scratch, { recursive: true, force: true });
    }
  }
});


test("browser target distinguishes canonical non-launch from missing required runner output", async (t) => {
  const { writeCanonicalFixture } = await import("../../observability/tests/retained-fixture.mjs");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "browser-prelaunch-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  const root = path.resolve(import.meta.dirname, "../../../..");
  for (const target of ["browser-e2e-stateful", "browser-e2e-visual-update", "browser-e2e-measurement"]) {
    for (const mode of ["admission", "fixture", "cancelled", "skipped", "launched", "missing_unit", "malformed_unit", "contradictory_unit", "duplicate_terminal", "unproved"]) {
      const id = `${target}-${mode}`, directory = path.join(scratch, id), group = "cohort-probe", unit = `browser_group:probe:${group}`;
      mkdirSync(directory, { mode: 0o700 });
      const fixture = writeCanonicalFixture(directory);
      const events = structuredClone(fixture.events).map(event => ({ ...event, unit_id: event.unit_id === "a" ? unit : event.unit_id }));
      const terminal = events.find(event => event.event === "completed");
      const failure = mode === "cancelled" ? { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" } :
        { failure_class: "infra", failure_reason: "resource_conflict" };
      Object.assign(terminal, { event: mode === "cancelled" ? "cancelled" : "failed", status: mode === "cancelled" ? "cancelled" : "failed", ...failure });
      const phase = { ...events.find(event => event.event === "started"), event: "phase_started", interval_id: "phase:1",
        phase: mode === "launched" ? "runner" : mode === "fixture" ? "fixture_acquire" : "host_admission" };
      if (!["unproved", "skipped"].includes(mode)) events.splice(7, 0, phase, { ...phase, event: "phase_finished", monotonic_ms: 8, phase_outcome: mode === "cancelled" ? "cancelled" : "failed" });
      if (mode === "skipped") {
        for (let i = events.length-1; i>=0; i--) if (["admitted","started"].includes(events[i].event)) events.splice(i,1);
        Object.assign(terminal,{ event:"skipped",status:"skipped",monotonic_ms:2,failure_reason:"dependency_failure" });
        delete terminal.failure_class;
        for (const event of events) if(event.unit_id===unit) event.needs=["setup"];
        const ancestor = structuredClone(fixture.events).filter(event=>event.unit_id==="a").map(event=>({ ...event,unit_id:"setup",monotonic_ms:event.monotonic_ms ? 1 : 0 }));
        Object.assign(ancestor.at(-1),{event:"failed",status:"failed",...failure});
        events.push(...ancestor);events.sort((a,b)=>a.monotonic_ms-b.monotonic_ms);
      }
      if (mode === "duplicate_terminal") events.splice(events.indexOf(terminal)+1, 0, { ...terminal });
      writeFileSync(path.join(directory, "unit-events.ndjson"), events.map((event, index) => JSON.stringify({ ...event, seq: index+1 })).join("\n")+"\n", { mode: 0o600 });
      mkdirSync(path.join(directory, "unit-results"), { mode: 0o700 });
      const result = { schema_id: "cartulary.harness_unit_result.v1", unit_id: unit, semantic_digest: `sha256:${"a".repeat(64)}`,
        status: terminal.status, exit_code: null, signal: null, ...(mode === "skipped" ? { failure_class:null,failure_reason:null } : failure),
        evidence_outputs: [`${target}/browser-groups/${group}/browser-group-result.json`], missing_outputs: [] };
      if (mode === "malformed_unit") result.unknown = true;
      if (mode === "contradictory_unit") result.failure_reason = "service_start_error";
      if (mode !== "missing_unit") writeFileSync(path.join(directory, "unit-results", `${unit.replaceAll(/[^A-Za-z0-9_.+-]+/gu,"-")}.json`), JSON.stringify(result), { mode: 0o600 });
      if (!["duplicate_terminal", "unproved", "launched"].includes(mode)) {
        const { reduceCanonicalUnitIntervals, notLaunchedUnitIDs } = await import("../../evidence-accounting/index.mjs");
        const facts = await reduceCanonicalUnitIntervals(path.join(directory,"unit-events.ndjson"), {unitIDs:mode === "skipped" ? [unit,"setup"] : [unit]});
        assert.equal(notLaunchedUnitIDs(facts).has(unit),true,"fixture must prove canonical non-launch");
      }
      if (!["malformed_unit"].includes(mode)) validateSchemaSync(result.schema_id,result);
      if (["admission","fixture","cancelled","skipped"].includes(mode)) {
        const { readUnlaunchedBrowserFailures } = await import("../browser-target-outcomes.mjs");
        const causes = await readUnlaunchedBrowserFailures(directory,{resetPrefix:"browser_reset:probe:",groups:[group],groupTargets:new Map([[group,target]])},[],{file:path.join(directory,"unit-events.ndjson")});
        assert.equal(causes.some(cause=>cause.failure_reason===failure.failure_reason),true);
        const { finalizeBrowserTarget } = await import("../browser-target-finalizer-cli.mjs");
        await assert.rejects(finalizeBrowserTarget({target,resetPrefix:"browser_reset:probe:",groups:[group],groupTargets:new Map([[group,target]])},directory,{file:path.join(directory,"unit-events.ndjson")}),error=>{
          assert.equal(error.failure_reason,failure.failure_reason,error.stack);return true;
        });
      }
      const env = { ...process.env, CARTULARY_TEST_RESULTS_DIR: scratch, CARTULARY_TEST_RUN_ID: id };
      delete env.CARTULARY_HARNESS_UNIT_EVENT_PREFIX;
      delete env.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
      const child = spawnSync(process.execPath, [path.join(root,"tools/harness/browser/browser-target-finalizer-cli.mjs"),
        "--target",target,"--groups",group,"--group-targets",`${group}=${target}`,"--reset-prefix","browser_reset:probe:"], { cwd:root,encoding:"utf8",env });
      const expected = mode === "cancelled" ? 15 : ["admission","fixture","skipped"].includes(mode) ? 4 : 11;
      assert.equal(child.status, expected, `${id}: ${child.stderr}`);
      assert.equal(existsSync(path.join(directory,target)),false,"unlaunched target published aggregate, reconciliation or promotion");
    }
  }
});
