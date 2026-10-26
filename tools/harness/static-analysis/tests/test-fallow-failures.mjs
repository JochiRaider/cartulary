import assert from "node:assert/strict";
import { existsSync, rmSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { createRunnerContext, validateSchemaSync } from "../../contract/index.mjs";
import { runFallowStatic } from "../fallow-static.mjs";
import { createReachabilityFixture } from "./fallow-fixtures.mjs";

const names = ["dead-code", "dupes", "health", "package-surface-dead-code"];
function report(name) {
  if (name === "dupes") return { kind: name, clone_groups: [], clone_families: [], stats: {} };
  if (name === "health") return { kind: name, findings: [], summary: {} };
  return { kind: "dead-code", unused_files: [], unused_exports: [], unused_types: [] };
}

function setup(t) {
  const fixture = createReachabilityFixture(t);
  fixture.write("node_modules/fallow/bin/fallow", "fixture executor owns child execution\n");
  fixture.write("packages/example/src/index.ts", "export const dead = 1;\n");
  const owner = fixture.readJSON("tools/fallow/reachability_owner.json");
  owner.blocking_package_surfaces.packages[0].entrypoint = "packages/example/src/index.ts";
  fixture.write("tools/fallow/reachability_owner.json", owner);
  const context = createRunnerContext({ repoRoot: fixture.root });
  const env = { CARTULARY_MACHINE_CACHE_DIR: context.machineCacheDir, CARTULARY_HARNESS_SCRATCH_ROOT: context.scratchRoot, CARTULARY_TEST_RUN_ID: "tr-fallow-fixture", CARTULARY_OUTPUT_MODE: "summary" };
  const targetRoot = ".cartulary/test-results/tr-fallow-fixture/frontend-fallow-static";
  const started = [];
  async function run(overrides = {}) {
    const code = await runFallowStatic({ root: fixture.root, env, execute: async (reportRoot, name, args, outputFile) => {
      started.push(name);
      const scenario = overrides[name] ?? {};
      if (scenario.gate) await scenario.gate;
      if (scenario.reject) throw new Error("fixture launch failed");
      const stdoutFile = path.join(reportRoot, `${name}.stdout.log`);
      const stderrFile = path.join(reportRoot, `${name}.stderr.log`);
      fixture.write(path.relative(fixture.root, stdoutFile), "");
      fixture.write(path.relative(fixture.root, stderrFile), scenario.stderr ?? "");
      if (!scenario.missing) fixture.write(path.relative(fixture.root, outputFile), scenario.data ?? report(name));
      if (name === "dead-code" && !scenario.missingSARIF) fixture.write(path.relative(fixture.root, path.join(reportRoot, "dead-code.sarif")), scenario.sarif ?? { version: "2.1.0", runs: [] });
      const exitCode = scenario.exitCode ?? 0;
      return { name, command: scenario.command ?? ["fallow", ...args], outputFile, stdoutFile, stderrFile, status: exitCode === 0 ? "pass" : "fail", exitCode };
    } });
    const summary = fixture.readJSON(`${targetRoot}/tool-run-summary.json`);
    validateSchemaSync("cartulary.tool_run_summary.v5", summary);
    assert.equal(code, summary.exit_code);
    return { code, summary };
  }
  return { fixture, run, started, targetRoot };
}

function expectFailure(result, failureClass, reason, code) {
  assert.equal(result.code, code);
  assert.equal(result.summary.status, "fail");
  assert.equal(result.summary.failure_class, failureClass);
  assert.equal(result.summary.failure_reason, reason);
}

for (const [name, data] of [
  ["dead-code", { ...report("dead-code"), unused_files: [{ path: "apps/web/src/dead.ts" }] }],
  ["package-surface-dead-code", { ...report("dead-code"), unused_exports: [{ path: "packages/example/src/index.ts", export_name: "dead" }] }],
]) {
  test(`Fallow publication classifies blocking ${name} as a tool diagnostic`, async (t) => {
    const { run, fixture, targetRoot } = setup(t);
    const result = await run({ [name]: { data, exitCode: 1 } });
    expectFailure(result, "harness", "tool_diagnostic_failure", 1);
    assert.equal(result.summary.failures.length, 1);
    assert.ok(result.summary.summary_artifacts.some((entry) => entry.role === "fallow_static_summary"));
    validateSchemaSync("cartulary.fallow_static_summary.v2", fixture.readJSON(`${targetRoot}/fallow-static-summary.json`));
  });
}

test("Fallow publication retains advisory reports without failing", async (t) => {
  const { run } = setup(t);
  const result = await run({ "dead-code": { data: { ...report("dead-code"), unused_files: [{ path: "tools/advisory.ts" }] }, exitCode: 1 } });
  assert.equal(result.code, 0);
  assert.equal(result.summary.status, "pass");
  assert.equal(result.summary.failure_class, null);
  assert.ok(result.summary.warnings.some((warning) => warning.kind === "fallow_findings"));
});

test("Fallow publication rejects bounded E2E findings despite analyzer success", async (t) => {
  const { run, fixture, targetRoot } = setup(t);
  const result = await run({ "dead-code": { data: {
    ...report("dead-code"),
    unused_files: [{ path: "apps/web/e2e/support/orphan.ts" }],
    unused_exports: [{ path: "apps/web/e2e/support/client.ts", export_name: "unusedValue" }],
    unused_types: [{ path: "apps/web/e2e/support/client.ts", export_name: "UnusedType" }],
  }, exitCode: 0 } });
  expectFailure(result, "harness", "tool_diagnostic_failure", 1);
  const summary = fixture.readJSON(`${targetRoot}/fallow-static-summary.json`);
  assert.equal(summary.extensions["cartulary.fallow_reachability"].web_surface.finding_count, 3);
  assert.equal(result.summary.failures.length, 3);
});

for (const [label, scenario] of [
  ["malformed JSON", { data: "{broken" }],
  ["missing report", { missing: true }],
  ["wrong report kind", { data: report("health") }],
  ["missing package rule", { data: { kind: "dead-code" } }],
  ["unsafe package path", { data: { ...report("dead-code"), unused_exports: [{ path: "../escape.ts", export_name: "dead" }] } }],
  ["missing package symbol", { data: { ...report("dead-code"), unused_types: [{ path: "packages/example/src/index.ts" }] } }],
]) {
  test(`Fallow publication rejects ${label} as an artifact failure`, async (t) => {
    const { run } = setup(t);
    expectFailure(await run({ "package-surface-dead-code": scenario }), "artifact", "artifact_error", 11);
  });
}

test("Fallow publication uses process status, never diagnostic substrings", async (t) => {
  const { run } = setup(t);
  const result = await run({ health: { exitCode: 127, stderr: "config is merely a diagnostic word" }, dupes: { exitCode: 2, stderr: "invalid input" } });
  expectFailure(result, "config", "configuration_error", 2);
  assert.deepEqual(result.summary.failures.map((failure) => [failure.label, failure.failure_class]).sort(), [["dupes", "config"], ["health", "harness"]]);
});

test("Fallow publication drains siblings and retains artifacts and secondary failures", async (t) => {
  const { run, started, fixture, targetRoot } = setup(t);
  const gate = Promise.withResolvers();
  let complete = false;
  const pending = run({ "dead-code": { data: "invalid" }, health: { reject: true }, dupes: { gate: gate.promise } }).then((result) => { complete = true; return result; });
  await Promise.resolve();
  assert.deepEqual(started, names);
  await Promise.resolve();
  assert.equal(complete, false);
  gate.resolve();
  const result = await pending;
  expectFailure(result, "harness", "tool_diagnostic_failure", 1);
  assert.deepEqual(result.summary.failures.map((failure) => [failure.label, failure.failure_class]).sort(), [["dead-code", "artifact"], ["health", "harness"]]);
  for (const role of ["resolved_fallow_config", "package_surface_config", "dupes_json", "package_surface_dead_code", "dead_code_json", "dead_code_sarif"]) {
    assert.ok(result.summary.summary_artifacts.some((entry) => entry.role === role), role);
  }
  assert.ok(existsSync(path.join(fixture.root, targetRoot, "fallow/dupes.json")));
});

for (const file of [".fallowrc.json", "tools/fallow/reachability_owner.json", "node_modules/fallow/bin/fallow", "apps/web/src/testing/testSetup.ts"]) {
  test(`Fallow publication rejects missing prerequisite ${file} before launch`, async (t) => {
    const { fixture, run, started } = setup(t);
    rmSync(path.join(fixture.root, file));
    expectFailure(await run(), "config", "configuration_error", 2);
    assert.deepEqual(started, []);
  });
}

for (const [name, scenario] of [
  ["missing SARIF", { missingSARIF: true }],
  ["malformed SARIF", { sarif: "invalid" }],
  ["invalid SARIF shape", { sarif: { version: "2.1.0" } }],
  ["invalid normalized command", { command: ["fallow", null] }],
]) {
  test(`Fallow publication rejects ${name}`, async (t) => {
    const { run } = setup(t);
    expectFailure(await run({ "dead-code": scenario }), "artifact", "artifact_error", 11);
  });
}

test("Fallow publication preserves blocking findings when a sibling report fails", async (t) => {
  const { run } = setup(t);
  const result = await run({ "dead-code": { data: { ...report("dead-code"), unused_files: [{ path: "apps/web/dead.ts" }] } }, health: { data: "invalid" } });
  expectFailure(result, "harness", "tool_diagnostic_failure", 1);
  assert.deepEqual(result.summary.failures.map((failure) => failure.failure_class).sort(), ["artifact", "harness"]);
});
