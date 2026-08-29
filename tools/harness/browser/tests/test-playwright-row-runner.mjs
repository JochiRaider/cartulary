import assert from "node:assert/strict";
import process from "node:process";
import test from "node:test";

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
