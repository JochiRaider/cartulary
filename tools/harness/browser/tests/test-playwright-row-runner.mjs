import assert from "node:assert/strict";
import process from "node:process";
import { spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from "node:fs";
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
    schema_id: "cartulary.browser_group_result.v7", target_id: "browser-e2e-visual",
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
      process.stdout.write(JSON.stringify({ user_password: 'private-fixture-value' }) + '\\n');
      process.stdin.resume();
      process.stdin.on('end', () => {
        process.stderr.write('late private-fixture-');
        process.stderr.write('value postgres://user:secret@localhost/db\\n');
        process.stderr.write(JSON.stringify({ schema_id: 'cartulary.browser_fixture_event.v1', attempt_id: attempt,
          stage: 'target.database', phase: 'cleanup', outcome: 'failed', elapsed_ms: 1, deadline_ms: 100,
          message: 'private-fixture-value: catalog retirement failed' }) + '\\n');
        process.exitCode = 1;
      });
    `, "--"],
  });
  await fixture.ready;
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
        process.stderr.write(JSON.stringify({ schema_id: 'cartulary.browser_fixture_event.v1', attempt_id: process.argv.at(-1),
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
          schema_id: "cartulary.browser_group_result.v7", target_id: target,
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
        validateSchemaSync(group.schema_id, group);
        writeFileSync(path.join(directory, "browser-group-result.json"), JSON.stringify(group), { mode: 0o600 });
        return groupID;
      });
      const result = spawnSync(process.execPath, [path.join(root, "tools/harness/browser/browser-target-finalizer-cli.mjs"),
        "--target", target, "--groups", groups.join(","), "--group-targets", groups.map(group => `${group}=${target}`).join(","),
        "--reset-prefix", `browser_reset:${target}:`], {
        cwd: root, encoding: "utf8", env: { ...process.env, CARTULARY_TEST_RESULTS_DIR: scratch, CARTULARY_TEST_RUN_ID: id },
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
