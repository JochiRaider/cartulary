#!/usr/bin/env node

import assert from "node:assert/strict";
import path from "node:path";

import {
  WorkGraphCompiler,
  executeUnitProcess,
} from "../../scheduler/work-graph/index.mjs";
import { adaptGoInvocation } from "../../execution/runners/go.mjs";

const invocation = {
  rows: [{ row_id: "row-a", selectors: ["TestA"] }],
};
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

const unitFailure = await executeUnitProcess({
  command: {
    executable: process.execPath,
    args: [
      "--input-type=module",
      "--eval",
      "process.stderr.write('[FAIL] row=row-a failure_class=infra failure_reason=service_readiness_timeout\\n'); process.exit(3);",
    ],
    environment: {},
  },
  timeout_ms: 1000,
});
assert.equal(unitFailure.exit_code, 3);
assert.equal(unitFailure.failure_class, "infra");
assert.equal(unitFailure.failure_reason, "service_readiness_timeout");

const root = path.resolve(import.meta.dirname, "../../../..");
const graph = new WorkGraphCompiler(root).compile({ kind: "target", target: "backend-unit" });
const goUnits = graph.units.filter((unit) => unit.unit_id.startsWith("go:"));
assert.ok(goUnits.length > 0);
assert.ok(goUnits.every((unit) => unit.evidence_outputs.some((output) => output.startsWith("rows/"))));
assert.ok(
  goUnits.every((unit) =>
    unit.evidence_outputs.every(
      (output) => output.startsWith("rows/") || output.startsWith("unit-results/"),
    ),
  ),
);
