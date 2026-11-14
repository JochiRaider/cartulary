import assert from "node:assert/strict";
import { linkSync, mkdirSync, readFileSync, renameSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { createScratch } from "../test-support/scratch.mjs";
import { assertInvalidTargetOutput, assertRealCommandOutput } from "./real-command-output.mjs";

const repoRoot = path.resolve(import.meta.dirname, "../../..");
const digest = `sha256:${"a".repeat(64)}`;
const timing = { setup_ms: 0, fixture_ms: 0, execution_ms: 0, collation_ms: 0,
  wrapper_ms: 0, unattributed_ms: 0, resource_blocking_ms: 0, process_count: 0 };

// Authored reader inputs, independent of the publisher and assertion helper.
function fixture(t, mode = "machine") {
  const scratch = createScratch(t, "real-command-output-");
  const runID = "requested";
  const runRoot = path.join(scratch, runID);
  mkdirSync(path.join(runRoot, "target-summaries"), { recursive: true, mode: 0o700 });
  const manifest = {
    schema_id: "cartulary.harness_run_manifest.v3", run_id: runID,
    command_id: "cartulary.harness.command.source_check.v1", target: "source-check", declared_inputs: { CHECK: "schema" },
    source_commit: "a".repeat(40), source_state: "dirty", source_digest: digest,
    toolchain_digest: digest, system_digest: digest, graph_digest: digest,
    instrumentation: { mode: "off", policy_digest: digest }, cache_mode: "normal",
    started_at: "2026-11-11T00:00:00.000Z", go_worker_budget: 1,
    capability_snapshot: { schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 1,
      memory_bytes: 1024, process_slots: 1, io_tokens: 1, postgres_lanes: 1,
      object_store_lanes: 1, writable_volume: true, port_lanes: 1,
      services: {}, sources: {}, snapshot_digest: digest },
  };
  const run = {
    schema_id: "cartulary.harness_run_summary.v1", run_id: runID, target: "source-check",
    status: "pass", failure_class: null, failure_reason: null,
    unit_counts: { total: 1, passed: 1, failed: 0, skipped: 0, cancelled: 0 },
    wall_duration_ms: 1, critical_path: [], actual_dependency_critical_path_ms: 0,
    timing_accounting: timing, resource_pressure: {}, cache: {}, artifact_refs: [],
  };
  const target = {
    schema_id: "cartulary.harness_target_summary.v1", target: "source-check",
    command_id: "cartulary.harness.command.source_check.v1", status: "pass",
    failure_class: null, failure_reason: null, workload_digest: digest, unit_ids: [],
    inclusive_wall_ms: 1, exclusive_wall_ms: 1, actual_dependency_critical_path_ms: 0,
    timing_accounting: timing, children: [], evidence_refs: [],
  };
  const put = (file, value) => writeFileSync(path.join(runRoot, file), `${JSON.stringify(value)}\n`, { mode: 0o600 });
  put("run-manifest.json", manifest); put("run-summary.json", run);
  put("target-summaries/source-check.json", target);
  writeFileSync(path.join(runRoot, "unit-events.ndjson"), [
    { seq: 1, monotonic_ms: 0, event: "run_started", status: "running" },
    { seq: 2, monotonic_ms: 1, event: "run_completed", status: "passed" },
  ].map((e) => JSON.stringify({ schema_id: "cartulary.harness_unit_event.v4", unit_id: "run",
    needs: [], resource_claims: {}, service_dependencies: [], ...e })).join("\n") + "\n", { mode: 0o600 });
  const options = { repoRoot, targetName: "source-check", mode, runRoot, runID,
    stdout: Buffer.from(mode === "machine" ? `${JSON.stringify(run)}\n` :
      `[GRAPH] target=source-check status=pass run_root=${runRoot}\n`), stderr: Buffer.alloc(0) };
  return { scratch, manifest, run, target, put, options, check: () => assertRealCommandOutput(options) };
}

for (const mode of ["summary", "machine"]) {
  test(`real command ${mode}: valid exact publication`, async (t) => { await fixture(t, mode).check(); });
  for (const [name, mutate] of [
    ["foreign manifest run", (f) => f.put("run-manifest.json", { ...f.manifest, run_id: "foreign" })],
    ["foreign summary run", (f) => f.put("run-summary.json", { ...f.run, run_id: "foreign" })],
    ["foreign manifest command", (f) => f.put("run-manifest.json", { ...f.manifest, command_id: "cartulary.harness.command.build.v3" })],
    ["foreign target command", (f) => f.put("target-summaries/source-check.json", { ...f.target, command_id: "cartulary.harness.command.build.v3" })],
    ["missing canonical field", (f) => { const value = { ...f.run }; delete value.unit_counts; f.put("run-summary.json", value); }],
    ["unknown canonical field", (f) => f.put("run-manifest.json", { ...f.manifest, unexpected: true })],
    ["contradictory passing summary", (f) => f.put("run-summary.json", { ...f.run, failure_reason: "usage_error" })],
    ["failed target", (f) => f.put("target-summaries/source-check.json", { ...f.target, status: "fail" })],
    ["foreign target", (f) => f.put("target-summaries/source-check.json", { ...f.target, target: "build" })],
    ["stderr", (f) => { f.options.stderr = Buffer.from("unexpected\n"); }],
    ["invalid UTF-8 output", (f) => { f.options.stdout = Buffer.from([0xff]); }],
    ["missing events", (f) => rmSync(path.join(f.options.runRoot, "unit-events.ndjson"))],
    ["malformed events", (f) => writeFileSync(path.join(f.options.runRoot, "unit-events.ndjson"), "{}\n")],
    ["duplicate retained key", (f) => writeFileSync(path.join(f.options.runRoot, "run-summary.json"), JSON.stringify(f.run).replace('{', '{"run_id":"foreign",'))],
    ["invalid retained UTF-8", (f) => writeFileSync(path.join(f.options.runRoot, "run-summary.json"), Buffer.from([0xff]))],
    ["linked retained file", (f) => linkSync(path.join(f.options.runRoot, "run-summary.json"), path.join(f.scratch, "alias.json"))],
    ["symlink retained file", (f) => { const file = path.join(f.options.runRoot, "run-summary.json"); renameSync(file, `${file}.original`); symlinkSync(`${file}.original`, file); }],
    ["missing manifest", (f) => rmSync(path.join(f.options.runRoot, "run-manifest.json"))],
  ]) {
    test(`real command ${mode}: rejects ${name}`, async (t) => {
      const f = fixture(t, mode); mutate(f); await assert.rejects(f.check());
    });
  }
}

for (const [name, mutate] of [
  ["missing LF", (s) => s.slice(0, -1)],
  ["extra blank line", (s) => `${s}\n`],
  ["CRLF", (s) => s.slice(0, -1) + "\r\n"],
  ["human output", (s) => `[PROGRESS] running\n${s}`],
  ["two objects", (s) => s + s],
  ["duplicate key", (s) => s.replace('{', '{"run_id":"foreign",')],
  ["schema impostor", () => '{"schema_id":"cartulary.harness_run_summary.v1","target":"source-check","status":"pass"}\n'],
  ["foreign identity", (s) => s.replace('"requested"', '"foreign"')],
  ["different retained result", (s) => s.replace('"wall_duration_ms":1', '"wall_duration_ms":2')],
]) {
  test(`real command machine: rejects ${name}`, async (t) => {
    const f = fixture(t); f.options.stdout = Buffer.from(mutate(f.options.stdout.toString()));
    await assert.rejects(f.check());
  });
}

for (const [name, mutate] of [
  ["foreign root", (s) => s.replace("/requested", "/foreign")],
  ["missing root", (s) => s.replace(/ run_root=.+/u, "")],
  ["foreign target", (s) => s.replace("target=source-check", "target=build")],
  ["failed status", (s) => s.replace("status=pass", "status=fail")],
  ["duplicate summary", (s) => s + s],
  ["duplicate identity field", (s) => s.replace("status=pass", "status=pass target=source-check")],
  ["line budget", (s) => `${s}extra\nextra\n`],
  ["byte budget", (s) => `${s}${"x".repeat(601)}\n`],
]) {
  test(`real command summary: rejects ${name}`, async (t) => {
    const f = fixture(t, "summary"); f.options.stdout = Buffer.from(mutate(f.options.stdout.toString()));
    await assert.rejects(f.check());
  });
}

test("real command output assertion leaves all evidence unchanged", async (t) => {
  const f = fixture(t); const file = path.join(f.options.runRoot, "run-summary.json");
  const before = readFileSync(file); await f.check(); assert.deepEqual(readFileSync(file), before);
});

test("invalid target assertion requires exact exit, no stdout and bounded typed diagnostic", () => {
  const valid = { status: 2, stdout: Buffer.alloc(0),
    stderr: Buffer.from("failure_reason=usage_error TARGET must name a declared target\n") };
  assertInvalidTargetOutput(valid);
  for (const changes of [ { status: 0 }, { status: 1 }, { status: 127 },
    { stdout: Buffer.from("{}\n") }, { stderr: Buffer.alloc(0) }, { stderr: Buffer.from("x".repeat(8193)) },
    { stderr: Buffer.from("failure_reason=usage_error") }, { stderr: Buffer.from("TARGET must name a declared target") } ]) {
    assert.throws(() => assertInvalidTargetOutput({ ...valid, ...changes }));
  }
});
