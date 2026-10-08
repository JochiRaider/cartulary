import assert from "node:assert/strict";
import process from "node:process";
import test from "node:test";
import { validateSchemaSync } from "../../contract/index.mjs";

import {
  publicExitCodeForFailure,
} from "../../contract/failure-taxonomy.mjs";
import {
  adaptPlaywrightReport,
  playwrightGroupExitCode,
} from "../../execution/runners/playwright.mjs";
import { executeUnitProcess } from "../../scheduler/work-graph/executor.mjs";

const file = "apps/web/e2e/keyboard.spec.ts";

function row(rowID, ...titles) {
  return {
    row_id: rowID,
    selector: { file, titles },
  };
}

function spec(title, status, duration = 7) {
  return {
    file,
    title,
    tests: [
      {
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

test("scheduler watchdog timeout remains a timing failure with exit 13", async () => {
  const result = await executeUnitProcess(
    {
      command: {
        args: ["-e", "setInterval(() => {}, 1000)"],
        environment: {},
        executable: process.execPath,
      },
      kind: "test",
      timeout_ms: 25,
    },
    { cwd: process.cwd(), inheritProcessEnvironment: false },
  );
  assert.equal(result.status, "failed");
  assert.equal(result.failure_class, "timing");
  assert.equal(result.failure_reason, "timeout_failure");
  assert.equal(publicExitCodeForFailure(result, { signal: result.signal }), 13);
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

test("browser fixture retirement waits for inherited streams and kills the owned process group", async () => {
  const { createBrowserFixtureProcess } = await import("../index.mjs");
  const descendant = "process.on('SIGTERM', () => {}); process.stdout.write(JSON.stringify({}) + String.fromCharCode(10)); setInterval(() => {}, 1000);";
  const fixture = createBrowserFixtureProcess({ command: process.execPath,
    args: ["-e", `
      const { spawn } = require('node:child_process');
      spawn(process.execPath, ['-e', ${JSON.stringify(descendant)}], { stdio: ['ignore', 'inherit', 'inherit'] });
      process.stdin.resume(); process.stdin.on('end', () => process.exit(0));
    `, "--"], startupMs: 5000, cleanupMs: 30, signalMs: 100,
  });
  try {
    await fixture.ready;
    await assert.rejects(fixture.stop(), /forced=true/);
    assert.equal(fixture.report().process.closed, true);
  } finally {
    await fixture.stop().catch(() => {});
  }
});
