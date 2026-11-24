#!/usr/bin/env node

import assert from "node:assert/strict";
import path from "node:path";
import os from "node:os";
import { chmodSync, existsSync, mkdtempSync, readFileSync, writeFileSync, rmSync } from "node:fs";

import { redactString } from "../../contract/index.mjs";
import {
  adaptGoInvocation, adaptGoInvocationFile, buildGoInvocations, goExecutionEnvironment,
  goTestDurationsDocument, goTestDurationsPath, goTestDurationsSchemaID, retainGoTestDurations,
} from "../../execution/runners/go.mjs";
import { goRunnerBuildContext } from "../../test-catalog/index.mjs";
import { foundationSchemaValidators } from "../../contract/generated/foundation-schema-validators.cjs";

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
  schema_id: "cartulary.harness_test_failure.v4",
  failure_class: "infra",
  failure_reason: "service_readiness_timeout",
  setup_source: "s3test",
  service: "object_store",
  readiness_stage: "put",
  attempt_count: 25,
  cleanup_outcome: "completed",
  phase: "setup",
  prior_failure: false,
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
assert.deepEqual(readiness[0].failure_diagnostic, { schema_id: "cartulary.harness_go_failure_diagnostic.v3", setup: readinessEnvelope, cleanup: [] });

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
assert.deepEqual(duplicateReadiness[0].failure_diagnostic, readiness[0].failure_diagnostic);

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

const cleanupEnvelope = {
  ...readinessEnvelope, failure_class: "harness", failure_reason: "cleanup_error",
  setup_source: "pgtest", service: "postgres", readiness_stage: "delete",
  attempt_count: 1, cleanup_outcome: "failed", phase: "cleanup",
  cleanup_diagnostics: [{ stage: "forced_drop", cause: "deadline_exceeded" }],
};
function withCleanup(envelopes, selected = invocation) {
  return adaptGoInvocation(selected, { status: 1, stdout: [
    ...envelopes.map(({ test = "TestA", envelope }) => JSON.stringify({ Action: "output", Test: test,
      Output: `CARTULARY_HARNESS_TEST_FAILURE=${JSON.stringify(envelope)}\n` })),
    ...selected.rows.flatMap((row) => row.selectors).map((Test) => JSON.stringify({ Action: "fail", Test })),
  ].join("\n") })[0];
}
const cleanupOnly = withCleanup([{ envelope: cleanupEnvelope }]);
assert.equal(cleanupOnly.failure_class, "harness");
assert.equal(cleanupOnly.failure_reason, "cleanup_error");
assert.equal(cleanupOnly.exit_code, 12);
assert.deepEqual(cleanupOnly.failure_diagnostic.cleanup, [cleanupEnvelope]);
const waitSnapshot = {
  type: "IPC", event: "CheckpointDone", blockers: 0,
  checkpointer_type: "IO", checkpointer_event: "DataFileSync", checkpointer_blockers: 0,
  requested: 7, done: 6, buffers: 200, write_ms: 1234, sync_ms: 56,
  barrier: { operation: "select", database_scope: "other", operation_age_ms: 123, status: "sampled", backend: "client backend", state: "idle", type: "Client", event: "ClientRead", blockers: 0 },
};
const dropAttempt = { duration_ms: 5000, admission_ms: 1, service_generation: null, admission_budget_ms: 15000, operation_ms: 4999, operation_budget_ms: 5000, admission: "acquired", outcome: "deadline_exceeded" };
const observedCleanup = { ...cleanupEnvelope, cleanup_diagnostics: [{ stage: "forced_drop", cause: "deadline_exceeded",
  postgres: { ...dropAttempt, wait: { status: "sampled", first: waitSnapshot, last: waitSnapshot } } }] };
const retainedWait = withCleanup([{ envelope: observedCleanup }]);
assert.equal(retainedWait.exit_code, 12);
assert.deepEqual(retainedWait.failure_diagnostic.cleanup, [observedCleanup]);

assert.equal(Object.hasOwn(foundationSchemaValidators, "cartulary.postgres_cleanup_observation.v2"), false);
assert.equal(typeof foundationSchemaValidators["cartulary.postgres_cleanup_observation.v3"], "function");
for (const [label, changes, valid] of [
  ["remaining owner budget", { operation_budget_ms: 25000 }, true],
  ["maximum owner budget", { operation_budget_ms: 35000 }, true],
  ["negative operation budget", { operation_budget_ms: -1 }, false],
  ["oversized operation budget", { operation_budget_ms: 35001 }, false],
  ["oversized admission budget", { admission_budget_ms: 15001 }, false],
  ["negative admission budget", { admission_budget_ms: -1 }, false],
  ["unadmitted operation", { admission: "rejected" }, false],
  ["unadmitted budget", { admission: "rejected", operation_ms: 0 }, false],
  ["unadmitted sampled operation", { admission: "rejected", operation_ms: 0, operation_budget_ms: 0 }, false],
  ["unadmitted zero values", { admission: "rejected", operation_ms: 0, operation_budget_ms: 0, wait: { status: "not_observed", first: null, last: null } }, true],
]) {
  const envelope = { ...observedCleanup, cleanup_diagnostics: [{ ...observedCleanup.cleanup_diagnostics[0],
    postgres: { ...observedCleanup.cleanup_diagnostics[0].postgres, ...changes } }] };
  const result = withCleanup([{ envelope }]);
  assert.equal(result.exit_code, valid ? 12 : 11, label);
  assert.deepEqual(result.failure_diagnostic, valid ? { schema_id: "cartulary.harness_go_failure_diagnostic.v3", setup: null, cleanup: [envelope] } : null, label);
}

// go test -json can split one logical diagnostic line across Output events.
const diagnosticLine = `    pgtest.go:676: CARTULARY_HARNESS_TEST_FAILURE=${JSON.stringify(observedCleanup)}\n`;
const fragmentEvents = (parts, test = "TestA") => parts.flatMap((Output, index) => [
  { Action: "output", Test: test, Output, OutputType: index === 0 ? "error" : "error-continue" },
  { Action: "output", Test: "TestOther", Output: "unrelated interleaved output\n" },
]);
const checkFraming = async (events, expected, label) => {
  const stdout = [...events, { Action: "fail", Test: "TestA", Elapsed: 0.01 }].map((event) => JSON.stringify(event)).join("\n");
  const memory = adaptGoInvocation(invocation, { status: 1, stdout });
  assert.equal(memory[0].exit_code, expected, label);
  const directory = mkdtempSync(path.join(os.tmpdir(), "go-envelope-framing-"));
  try {
    const file = path.join(directory, "events");
    writeFileSync(file, stdout);
    assert.deepEqual(await adaptGoInvocationFile(invocation, { status: 1, stdoutPath: file }), memory, `${label}: streaming parity`);
  } finally { rmSync(directory, { recursive: true, force: true }); }
  return memory[0];
};
const split = await checkFraming(fragmentEvents([diagnosticLine.slice(0, 1024), diagnosticLine.slice(1024)]), 12, "fragmented cleanup retains its failure class");
assert.deepEqual(split.failure_diagnostic.cleanup, [observedCleanup]);
await checkFraming([null, [], "process diagnostic", ...fragmentEvents([diagnosticLine.slice(0, 1024), diagnosticLine.slice(1024)])], 12, "non-event JSON does not alter framing");
await checkFraming([
  { Action: "output", Package: "first", Test: "TestA", Output: diagnosticLine.slice(0, 1024) },
  { Action: "output", Package: "second", Test: "TestA", Output: diagnosticLine.slice(1024) },
], 11, "another package cannot complete an envelope at stream end");

const markerCut = diagnosticLine.indexOf("CARTULARY_HARNESS_TEST_FAILURE") + 11;
await checkFraming(fragmentEvents([diagnosticLine.slice(0, markerCut), diagnosticLine.slice(markerCut, 1024), diagnosticLine.slice(1024)]), 12, "split marker is reassembled within its test");
await checkFraming(fragmentEvents([diagnosticLine.trimEnd()]), 11, "unterminated envelope fails closed");
await checkFraming(fragmentEvents([diagnosticLine.slice(0, 1024)]), 11, "truncated envelope fails closed");
await checkFraming([
  { Action: "output", Test: "TestA", Output: diagnosticLine.slice(0, 1024) },
  { Action: "output", Test: "TestOther", Output: diagnosticLine.slice(1024) },
], 11, "another test cannot complete an envelope");
await checkFraming(fragmentEvents([diagnosticLine.trimEnd(), " ".repeat(65536), "\n"]), 11, "oversized envelope fails closed");
await checkFraming([{ Action: "output", Test: "TestA", Output: diagnosticLine + diagnosticLine }], 12, "multiple diagnostic lines retain logical framing");
await checkFraming([
  ...Array.from({ length: 128 }, (_, index) => ({ Action: "output", Test: `TestOther${index}`, Output: "partial" })),
  { Action: "output", Test: "TestA", Output: diagnosticLine.slice(0, 1024) },
], 11, "unfinished-line bound cannot silently discard a diagnostic");

for (const invalidWait of [
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, barrier: { ...waitSnapshot.barrier, operation: "PrivateCredentialAlpha123" } } },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, barrier: { ...waitSnapshot.barrier, database_scope: "private_database" } } },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, barrier: { ...waitSnapshot.barrier, operation_age_ms: -1 } } },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, barrier: { ...waitSnapshot.barrier, operation_age_ms: 86400001 } } },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, barrier: { ...waitSnapshot.barrier, operation: "" } } },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, barrier: { ...waitSnapshot.barrier, database_scope: "" } } },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, barrier: { ...waitSnapshot.barrier, backend: "PrivateCredentialAlpha123" } } },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, barrier: { ...waitSnapshot.barrier, pid: 123 } } },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, barrier: { ...waitSnapshot.barrier, status: "unavailable" } } },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, event: "PrivateCredentialAlpha123" } },
  { status: "sampled", first: null, last: waitSnapshot },
  { status: "not_observed", first: waitSnapshot, last: null },
  { status: "sampled", first: waitSnapshot, last: { ...waitSnapshot, buffers: -1 } },
]) {
  const rejected = withCleanup([{ envelope: { ...cleanupEnvelope, cleanup_diagnostics: [{ stage: "forced_drop", cause: "deadline_exceeded", postgres: { ...dropAttempt, wait: invalidWait } }] } }]);
  assert.equal(rejected.failure_reason, "scheduler_accounting_error", "invalid drop observations must fail before evidence attribution");
  assert.equal(rejected.failure_diagnostic, null);
}
const misplacedWait = withCleanup([{ envelope: { ...observedCleanup, cleanup_diagnostics: [{ ...observedCleanup.cleanup_diagnostics[0], stage: "catalog_unlock" }] } }]);
assert.equal(misplacedWait.failure_reason, "cleanup_error", "attempt evidence survives independent unlock failure");
const afterAssertion = withCleanup([{ envelope: { ...cleanupEnvelope, prior_failure: true } }]);
assert.equal(afterAssertion.failure_reason, "test_assertion_failure");
assert.equal(afterAssertion.exit_code, 10);
assert.equal(afterAssertion.failure_diagnostic.cleanup[0].failure_reason, "cleanup_error");
const afterSetup = withCleanup([{ envelope: readinessEnvelope }, { envelope: { ...cleanupEnvelope, prior_failure: true } }]);
assert.equal(afterSetup.failure_reason, "service_readiness_timeout");
assert.equal(afterSetup.exit_code, 3);
assert.deepEqual(afterSetup.failure_diagnostic.setup, readinessEnvelope);
assert.equal(afterSetup.failure_diagnostic.cleanup.length, 1);
const repeatedCleanup = withCleanup([{ envelope: cleanupEnvelope }, { envelope: cleanupEnvelope },
  { envelope: { ...cleanupEnvelope, prior_failure: true } }]);
assert.equal(repeatedCleanup.failure_reason, "cleanup_error", "later cleanup must not manufacture an earlier product failure");
assert.equal(repeatedCleanup.exit_code, 12);
const afterChildCleanup = withCleanup([{ test: "TestA/child", envelope: cleanupEnvelope },
  { envelope: { ...cleanupEnvelope, prior_failure: true } }]);
assert.equal(afterChildCleanup.exit_code, 12);
const separateAssertion = withCleanup([{ envelope: cleanupEnvelope }], { rows: [{ row_id: "row-a", selectors: ["TestA", "TestB"] }] });
assert.equal(separateAssertion.exit_code, 10, "a separate assertion survives another selector's cleanup failure");
const assertionAndSetup = withCleanup([{ envelope: readinessEnvelope }, { envelope: { ...cleanupEnvelope, prior_failure: true } }], { rows: [{ row_id: "row-a", selectors: ["TestA", "TestB"] }] });
assert.equal(assertionAndSetup.exit_code, 10, "product precedence survives another selector's typed setup and cleanup");
assert.equal(assertionAndSetup.failure_diagnostic.setup.failure_reason, "service_readiness_timeout");
const childAssertionAndSetup = adaptGoInvocation(invocation, { status: 1, stdout: [
  JSON.stringify({ Action: "output", Test: "TestA/setup", Output: `CARTULARY_HARNESS_TEST_FAILURE=${JSON.stringify(readinessEnvelope)}\n` }),
  JSON.stringify({ Action: "fail", Test: "TestA/setup" }),
  JSON.stringify({ Action: "fail", Test: "TestA/assertion" }),
  JSON.stringify({ Action: "fail", Test: "TestA" }),
].join("\n") })[0];
assert.equal(childAssertionAndSetup.exit_code, 10, "a failing sibling subtest retains product precedence");
const unrelatedCleanup = withCleanup([{ test: "TestOther", envelope: cleanupEnvelope }]);
assert.equal(unrelatedCleanup.exit_code, 10);
assert.equal(unrelatedCleanup.failure_diagnostic, null);
const malformedCleanup = withCleanup([{ envelope: { ...cleanupEnvelope, failure_class: "product" } }]);
assert.equal(malformedCleanup.failure_reason, "scheduler_accounting_error");
const unsafeCleanup = withCleanup([{ envelope: { ...cleanupEnvelope,
  cleanup_diagnostics: [{ stage: "forced_drop", cause: "private-transport-text" }] } }]);
assert.equal(unsafeCleanup.failure_reason, "scheduler_accounting_error");
assert.equal(unsafeCleanup.failure_diagnostic, null);

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
const buildContext = goRunnerBuildContext(root);
const machineEnvironment = {
  HOME: "/runner-home", XDG_CACHE_HOME: "/runner-cache",
  GO_CACHE_DIR: "/declared/build", GO_MOD_CACHE_DIR: "/declared/mod", GO_TMP_DIR: "/declared/tmp",
  GOCACHE: "/foreign/build", GOMODCACHE: "/foreign/mod", GOTMPDIR: "/foreign/tmp", GOROOT: "/foreign/toolchain",
};
assert.deepEqual(goExecutionEnvironment(root, machineEnvironment), {
  GOCACHE: "/declared/build", GOMODCACHE: "/declared/mod", GOTMPDIR: "/declared/tmp", GOROOT: "",
});
assert.deepEqual(goExecutionEnvironment(root, { HOME: "/runner-home" }), {
  GOCACHE: "/runner-home/.cache/cartulary/go/build", GOMODCACHE: "/runner-home/.cache/cartulary/go/mod",
  GOTMPDIR: "/runner-home/.cache/cartulary/go/tmp", GOROOT: "",
});
for (const invalid of [{ GO_MOD_CACHE_DIR: root }, { GO_TMP_DIR: "/declared/mod/overlap" }, { GO_CACHE_DIR: "relative" }, { GO_MOD_CACHE_DIR: "" }]) {
  assert.throws(() => goExecutionEnvironment(root, { ...machineEnvironment, ...invalid }));
}

const [ordinaryInvocation] = buildGoInvocations([
  { row_id: "fixture", selector: { package: "./internal/fixture", tests: ["TestActive"] } },
], 2, "pinned-go", buildContext);
const contaminatedHost = {
  GOOS: "windows", GOARCH: "arm64", GOAMD64: "v3", CGO_ENABLED: "1",
  GOFLAGS: "-tags=private -race -overlay=/external/source.json", GOENV: "/external/goenv",
  GOEXPERIMENT: "jsonv2", GOTOOLCHAIN: "auto", GOWORK: "/external/go.work", GO111MODULE: "off",
};
const childEnvironment = { ...contaminatedHost, ...ordinaryInvocation.environment };
assert.deepEqual(childEnvironment, buildContext.environment);
assert.equal(childEnvironment.CGO_ENABLED, "0");
assert.equal(childEnvironment.GOFLAGS, "");
assert.equal(childEnvironment.GOENV, "off");
assert.equal(childEnvironment.GOWORK, "off");
assert.equal(childEnvironment.GOEXPERIMENT, "");
assert.equal(childEnvironment.GOOS, buildContext.goos);
assert.equal(childEnvironment.GOARCH, buildContext.goarch);
assert.ok(buildContext.tags.has(`${childEnvironment.GOARCH}.${childEnvironment.GOAMD64}`));
assert.ok(!buildContext.tags.has("cgo") && !buildContext.tags.has("race"));
assert.deepEqual(ordinaryInvocation.args, ["test", "-json", "-count=1", "-p=2", "-run", "^(?:TestActive)$", "./internal/fixture"]);
const [moduleInvocation] = buildGoInvocations([
  { row_id: "independent", selector: { package: "./referencepacks/format/dataset", tests: ["TestGenericAdmissionAndLookup"] } },
], 2, "pinned-go", buildContext);
assert.deepEqual(moduleInvocation.args, ["-C", "referencepacks/format", "test", "-mod=readonly", "-json", "-count=1", "-p=2", "-run", "^(?:TestGenericAdmissionAndLookup)$", "./dataset"]);
assert.equal(moduleInvocation.environment.GOWORK, "off");
assert.throws(() => buildGoInvocations([
  { row_id: "unknown", selector: { package: "./referencepacks/undeclared/data", tests: ["TestUnknown"] } },
], 1, "go", buildContext), /unregistered independent Go module/);
// A shuffle seed only orders the exact symbol selection.
const [shuffledInvocation] = buildGoInvocations([
  { row_id: "fixture", selector: { package: "./internal/fixture", tests: ["TestActive"] } },
], 2, "pinned-go", { ...buildContext, goShuffleSeed: "1" });
assert.deepEqual(shuffledInvocation.args.filter((arg) => !arg.startsWith("-shuffle=")), ordinaryInvocation.args);
assert.ok(shuffledInvocation.args.includes("-shuffle=1"));
assert.equal(shuffledInvocation.environment.GOFLAGS, "");
assert.throws(() => buildGoInvocations([], 1, "go", { ...buildContext, goShuffleSeed: "on" }), /shuffle seed/);
const [racedInvocation] = buildGoInvocations([
  { row_id: "fixture", selector: { package: "./internal/fixture", tests: ["TestActive"] } },
], 2, "pinned-go", { ...buildContext, goRace: true });
assert.deepEqual(racedInvocation.args.filter((arg) => arg !== "-race"), ordinaryInvocation.args);
assert.ok(racedInvocation.args.includes("-race"));
assert.equal(racedInvocation.environment.CGO_ENABLED, "1");

// TH-HARNESS-AC-165: Go test durations come from the outcome parse, keep
// terminal-event order, redact and bound names, and never change outcomes.
{
  const durationScratch = mkdtempSync(path.join(os.tmpdir(), "go-test-durations-"));
  try {
    const durationInvocation = { args: ["test", "-json", "./internal/fixture"], rows: [
      { row_id: "row-b", selectors: ["TestB"] }, { row_id: "row-a", selectors: ["TestA"] },
    ] };
    const longName = `TestA/${"x".repeat(600)}`;
    const durationEvents = [
      { Action: "run", Package: "example.test/fixture", Test: "TestA" },
      { Action: "pass", Package: "example.test/fixture", Test: "TestA/one", Elapsed: 1.25 },
      { Action: "pass", Package: "example.test/fixture", Test: "TestA/two/nested", Elapsed: 0.004 },
      { Action: "fail", Package: "example.test/fixture", Test: "TestA/two", Elapsed: 0.5 },
      { Action: "skip", Package: "example.test/fixture", Test: "TestA/password=hunter2", Elapsed: 0 },
      { Action: "pass", Package: "example.test/fixture", Test: longName, Elapsed: 0.1 },
      { Action: "fail", Package: "example.test/fixture", Test: "TestA", Elapsed: 2 },
      { Action: "pass", Package: "example.test/fixture", Test: "TestB", Elapsed: 0.25 },
      { Action: "fail", Package: "example.test/fixture", Elapsed: 2.6 },
    ];
    const eventsFile = path.join(durationScratch, "events");
    writeFileSync(eventsFile, durationEvents.map((event) => JSON.stringify(event)).join("\n"), { mode: 0o600 });
    const records = [];
    const withDurations = await adaptGoInvocationFile(durationInvocation, { stdoutPath: eventsFile }, () => {}, (record) => records.push(record));
    const withoutDurations = await adaptGoInvocationFile(durationInvocation, { stdoutPath: eventsFile });
    assert.deepEqual(withDurations, withoutDurations, "duration diagnostics cannot change row outcomes");
    assert.equal(records.length, 1);
    assert.deepEqual(records[0], {
      package: "example.test/fixture",
      rows: [{ row_id: "row-b", selectors: ["TestB"] }, { row_id: "row-a", selectors: ["TestA"] }],
      package_action: "fail", package_elapsed_ms: 2600,
      tests: [
        { name: "TestA/one", action: "pass", duration_ms: 1250 },
        { name: "TestA/two/nested", action: "pass", duration_ms: 4 },
        { name: "TestA/two", action: "fail", duration_ms: 500 },
        { name: "TestA/password=hunter2", action: "skip", duration_ms: 0 },
        { name: longName, action: "pass", duration_ms: 100 },
        { name: "TestA", action: "fail", duration_ms: 2000 },
        { name: "TestB", action: "pass", duration_ms: 250 },
      ],
    });
    const document = goTestDurationsDocument("go:unit", records, redactString);
    assert.equal(typeof foundationSchemaValidators[goTestDurationsSchemaID], "function");
    assert.equal(document.omitted, 1, "an over-length name is counted, not truncated");
    assert.deepEqual(document.invocations[0].tests.map((test) => test.name),
      ["TestA/one", "TestA/two/nested", "TestA/two", "TestA/password=<redacted>", "TestA", "TestB"]);

    const many = { ...records[0], tests: Array.from({ length: 4100 }, (_, index) => ({ name: `TestA/case${index}`, action: "pass", duration_ms: index })) };
    const bounded = goTestDurationsDocument(null, [many, { ...records[0] }], redactString);
    assert.equal(bounded.invocations[0].tests.length, 4096);
    assert.equal(bounded.invocations[1].tests.length, 0);
    assert.equal(bounded.omitted, 4 + records[0].tests.length, "the bound spans the unit's invocations");
    assert.equal(bounded.invocations[0].tests.at(-1).name, "TestA/case4095");

    const goRows = durationInvocation.rows.map((row) => ({ ...row, runner: "go" }));
    const reports = [];
    const report = (message) => reports.push(message);
    const runRoot = path.join(durationScratch, "run");
    const destination = goTestDurationsPath(runRoot, goRows);
    assert.equal(destination, path.join(runRoot, "unit-logs", "row-row-a", "go-test-durations.json"));
    assert.equal(retainGoTestDurations({ environment: {}, runRoot, rows: goRows, records, redact: redactString, report }), null);
    assert.equal(existsSync(destination), false, "off mode writes no durations");
    assert.equal(retainGoTestDurations({ environment: { CARTULARY_HARNESS_EXECUTION_ENABLED: "1" }, runRoot,
      rows: goRows.map((row) => ({ ...row, runner: "vitest" })), records, redact: redactString, report }), null);
    const basic = { CARTULARY_HARNESS_EXECUTION_ENABLED: "1", CARTULARY_WORK_UNIT_ID: "go:unit" };
    assert.equal(retainGoTestDurations({ environment: basic, runRoot, rows: goRows, records, redact: redactString, report }), destination);
    const retained = JSON.parse(readFileSync(destination, "utf8"));
    assert.deepEqual(retained, document);
    assert.deepEqual(reports, []);

    const blocked = path.join(durationScratch, "blocked");
    writeFileSync(blocked, "not a directory", { mode: 0o600 });
    chmodSync(blocked, 0o400);
    assert.equal(retainGoTestDurations({ environment: basic, runRoot: blocked, rows: goRows, records, redact: redactString, report }), null);
    assert.deepEqual(reports, ["go test durations unavailable (diagnostic_write_failed)\n"]);
  } finally { rmSync(durationScratch, { recursive: true, force: true }); }
}
