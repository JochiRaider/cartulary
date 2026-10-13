#!/usr/bin/env node

import assert from "node:assert/strict";
import path from "node:path";
import os from "node:os";
import { mkdtempSync, writeFileSync, rmSync } from "node:fs";

import {
  WorkGraphCompiler,
} from "../../scheduler/work-graph/index.mjs";
import { adaptGoInvocation, adaptGoInvocationFile } from "../../execution/runners/go.mjs";

const invocation = {
  rows: [{ row_id: "row-a", selectors: ["TestA"] }],
};
const scratch = mkdtempSync(path.join(os.tmpdir(), "go-failure-details-"));
try {
  const file = path.join(scratch, "events");
  const events = [];
  for (let i = 0; i < 12; i++) {
    const name = `TestA/case${i}`;
    events.push({ Action: "output", Test: name, Output: `${"x".repeat(5000)}early assertion ${i}\n` });
    events.push({ Action: "fail", Test: name });
  }
  events.push({ Action: "output", Test: "TestOther", Output: "unselected failure detail" }, { Action: "fail", Test: "TestOther" });
  for (let i = 0; i < 100; i++) events.push({ Action: "output", Test: "TestLater", Output: "passing noise".repeat(100) });
  writeFileSync(file, events.map((event) => JSON.stringify(event)).join("\n"), { mode: 0o600 });
  let output;
  const rows = await adaptGoInvocationFile(invocation, { stdoutPath: file }, (value) => { output = value; });
  assert.equal(rows[0].terminal_state, "infrastructure_failed", "diagnostic detail cannot invent missing parent accounting");
  assert.match(output, /early assertion 0/);
  assert.match(output, /selected=8 omitted=4/);
  assert.ok(!output.includes("passing noise") && !output.includes("unselected failure detail"));
  assert.ok(Buffer.byteLength(output) < 40 * 1024);
  writeFileSync(file, JSON.stringify({ Action: "pass", Test: "TestA" }));
  await adaptGoInvocationFile(invocation, { stdoutPath: file }, (value) => { output = value; });
  assert.equal(output, "");
} finally { rmSync(scratch, { recursive: true, force: true }); }
const passed = adaptGoInvocation(invocation, {
  status: 0,
  stdout: `${JSON.stringify({ Action: "pass", Test: "TestA", Elapsed: 0.01 })}\n`,
});
assert.deepEqual(passed.map((entry) => entry.terminal_state), ["passed"]);
const missing = adaptGoInvocation(invocation, { status: 1, stdout: "" });
assert.deepEqual(missing.map((entry) => entry.terminal_state), ["infrastructure_failed"]);
assert.deepEqual(missing.map((entry) => entry.exit_code), [3]);
const failed = adaptGoInvocation(invocation, {
  status: 1,
  stdout: `${JSON.stringify({ Action: "fail", Test: "TestA", Elapsed: 0.01 })}\n`,
});
assert.deepEqual(failed.map((entry) => entry.terminal_state), ["failed"]);
assert.deepEqual(failed.map((entry) => entry.exit_code), [10]);
const readinessEnvelope = {
  schema_id: "cartulary.harness_test_failure.v1",
  failure_class: "infra",
  failure_reason: "service_readiness_timeout",
  setup_source: "s3test",
  service: "object_store",
  readiness_stage: "put",
  attempt_count: 25,
  cleanup_outcome: "completed",
};
const readiness = adaptGoInvocation(invocation, {
  status: 1,
  stdout: [
    JSON.stringify({
      Action: "output",
      Test: "TestA/readiness",
      Output: `CARTULARY_HARNESS_TEST_FAILURE=${JSON.stringify(readinessEnvelope)}\n`,
    }),
    JSON.stringify({ Action: "fail", Test: "TestA", Elapsed: 120 }),
  ].join("\n"),
});
assert.equal(readiness[0].terminal_state, "infrastructure_failed");
assert.equal(readiness[0].failure_class, "infra");
assert.equal(readiness[0].failure_reason, "service_readiness_timeout");
assert.equal(readiness[0].exit_code, 3);
assert.deepEqual(readiness[0].failure_diagnostic, readinessEnvelope);

const duplicateReadiness = adaptGoInvocation(invocation, {
  status: 1,
  stdout: [
    ...["TestA", "TestA/readiness"].map((test) => JSON.stringify({
      Action: "output",
      Test: test,
      Output: `CARTULARY_HARNESS_TEST_FAILURE=${JSON.stringify(readinessEnvelope)}\n`,
    })),
    JSON.stringify({ Action: "fail", Test: "TestA", Elapsed: 120 }),
  ].join("\n"),
});
assert.equal(duplicateReadiness[0].failure_class, "infra");
assert.equal(duplicateReadiness[0].failure_reason, "service_readiness_timeout");
assert.deepEqual(duplicateReadiness[0].failure_diagnostic, readinessEnvelope);

const conflictingReadiness = adaptGoInvocation(invocation, {
  status: 1,
  stdout: [
    JSON.stringify({
      Action: "output",
      Test: "TestA",
      Output: `CARTULARY_HARNESS_TEST_FAILURE=${JSON.stringify(readinessEnvelope)}\n`,
    }),
    JSON.stringify({
      Action: "output",
      Test: "TestA/readiness",
      Output: `CARTULARY_HARNESS_TEST_FAILURE=${JSON.stringify({ ...readinessEnvelope, cleanup_outcome: "failed" })}\n`,
    }),
    JSON.stringify({ Action: "fail", Test: "TestA", Elapsed: 120 }),
  ].join("\n"),
});
assert.equal(conflictingReadiness[0].failure_class, "harness");
assert.equal(conflictingReadiness[0].failure_reason, "scheduler_accounting_error");
assert.equal(conflictingReadiness[0].exit_code, 11);
assert.equal(conflictingReadiness[0].failure_diagnostic, null);

const unrelatedReadiness = adaptGoInvocation(invocation, {
  status: 1,
  stdout: [
    JSON.stringify({
      Action: "output",
      Test: "TestOther/readiness",
      Output: `CARTULARY_HARNESS_TEST_FAILURE=${JSON.stringify(readinessEnvelope)}\n`,
    }),
    JSON.stringify({ Action: "fail", Test: "TestA", Elapsed: 0.01 }),
  ].join("\n"),
});
assert.equal(unrelatedReadiness[0].failure_class, "product");
assert.equal(unrelatedReadiness[0].failure_reason, "test_assertion_failure");
assert.equal(unrelatedReadiness[0].exit_code, 10);
assert.equal(unrelatedReadiness[0].failure_diagnostic, null);

const cancelledEnvelope = {
  ...readinessEnvelope,
  failure_class: "interrupted",
  failure_reason: "cancelled_or_interrupted",
};
const cancelled = adaptGoInvocation(invocation, {
  status: 1,
  stdout: [
    JSON.stringify({
      Action: "output",
      Test: "TestA/readiness",
      Output: `CARTULARY_HARNESS_TEST_FAILURE=${JSON.stringify(cancelledEnvelope)}\n`,
    }),
    JSON.stringify({ Action: "fail", Test: "TestA", Elapsed: 0.01 }),
  ].join("\n"),
});
assert.equal(cancelled[0].terminal_state, "cancelled");
assert.equal(cancelled[0].failure_class, "interrupted");
assert.equal(cancelled[0].failure_reason, "cancelled_or_interrupted");
assert.equal(cancelled[0].exit_code, 130);

const malformed = adaptGoInvocation(invocation, {
  status: 1,
  stdout: [
    JSON.stringify({
      Action: "output",
      Test: "TestA",
      Output: "CARTULARY_HARNESS_TEST_FAILURE={not-json}\n",
    }),
    JSON.stringify({ Action: "fail", Test: "TestA", Elapsed: 0.01 }),
  ].join("\n"),
});
assert.equal(malformed[0].failure_class, "harness");
assert.equal(malformed[0].failure_reason, "scheduler_accounting_error");
assert.equal(malformed[0].exit_code, 11);

const root = path.resolve(import.meta.dirname, "../../../..");
const graph = new WorkGraphCompiler(root).compile({ kind: "target", target: "backend-unit" });
const goUnits = graph.units.filter((unit) => unit.unit_id.startsWith("go:"));
assert.ok(goUnits.length > 0);
assert.ok(
  goUnits.every((unit) =>
    unit.current_run_evidence_outputs.some((output) => output.startsWith("rows/")),
  ),
);
assert.ok(
  goUnits.every((unit) =>
    unit.current_run_evidence_outputs.every(
      (output) => output.startsWith("rows/") || output.startsWith("unit-results/"),
    ),
  ),
);
