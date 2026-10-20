import assert from "node:assert/strict";
import test from "node:test";
import { mkdtempSync, rmSync, symlinkSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { buildMakeNodeToolChildEnv, buildMakeNodeToolInvocation } from "../../command-surface/make-node-tools.mjs";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { auditOwnerEvidence } from "../owner-evidence-audit.mjs";
import { reduceCanonicalUnitIntervals } from "../../evidence-accounting/index.mjs";

const digest = `sha256:${"a".repeat(64)}`;
const foreign = `sha256:${"b".repeat(64)}`;
const owner = "module.demo";

test("owner audit CLI retains explicit inputs after the Make dispatcher consumes them", (t) => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-owner-audit-cli-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  const manifestFile = path.join(root, "manifest.json");
  atomicLocalFile(manifestFile, JSON.stringify({ schema_id: "cartulary.harness_evidence_root_manifest.v1", owner_id: owner, entries: [{ target: "test-slice", run_root: "unused" }] }));
  const env = { ...process.env, OWNER: owner, EVIDENCE_ROOTS_FILE: manifestFile,
    CARTULARY_MAKE_INPUT_SOURCES: "OWNER=cli EVIDENCE_ROOTS_FILE=cli",
    CARTULARY_TEST_RESULTS_DIR: path.join(root, "results"), CARTULARY_TEST_RUN_ID: "explicit-arguments" };
  delete env.CARTULARY_HARNESS_IDENTITY_PREPARED;
  const invocation = buildMakeNodeToolInvocation("test-evidence-audit", env);
  const childEnv = buildMakeNodeToolChildEnv("test-evidence-audit", env);
  assert.equal(childEnv.OWNER, undefined);
  const child = spawnSync(process.execPath, [invocation.script, ...invocation.args], { env: childEnv, encoding: "utf8" });
  assert.equal(child.status, 2, child.stderr);
  assert.match(child.stderr, /unknown active owner module\.demo/u);
});
function fixture(t, target = "test-slice") {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-owner-audit-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  const rows = ["first", "second"].map((name) => ({ row_id: `${owner}.unit.${name}`, owner_id: owner, runner: "go", evidence_class: "unit", runtime_profile_id: "none" }));
  const units = rows.map((row, i) => ({ unit_id: `u${i}`, kind: "runner", needs: [], resource_claims: { cpu: 1 }, service_dependencies: [], semantic_digest: digest, row_bindings: [{ row_id: row.row_id, runner: row.runner }], current_run_evidence_outputs: [`rows/${row.row_id}.json`, `unit-results/u${i}.json`] }));
  const rowResult = (row) => ({ schema_id: "cartulary.harness_row_result.v6", row_id: row.row_id, runner: "go", terminal_state: "passed", duration_ms: 1, exit_code: 0,
    failure_class: null, failure_reason: null, failure_diagnostic: null, started_at: "2026-10-13T00:00:00.000Z", finished_at: "2026-10-13T00:00:00.001Z", wall_duration_ms: 1 });
  const unitResult = (unit) => ({ schema_id: "cartulary.harness_unit_result.v1", unit_id: unit.unit_id, semantic_digest: digest, status: "passed", exit_code: 0, signal: null,
    failure_class: null, failure_reason: null, missing_outputs: [], evidence_outputs: unit.current_run_evidence_outputs });
  const cleanup = { schema_id: "cartulary.harness_cleanup_results.v2", run_id: "synthetic", results: units.map((unit, i) => ({ sequence: i + 1, operation: "host_release", unit_id: unit.unit_id,
    fixture_lease_id: null, outcome: "completed", failure_class: null, failure_reason: null, failures: [], artifact_refs: [], diagnostic: null })) };
  const save = (relative, value) => atomicLocalFile(path.join(root, "run", relative), JSON.stringify(value), { replace: true });
  rows.forEach((row) => save(`rows/${row.row_id}.json`, rowResult(row)));
  units.forEach((unit) => save(`unit-results/${unit.unit_id}.json`, unitResult(unit)));
  save("cleanup-results.json", cleanup);
  const commandID = (name) => `cartulary.harness.command.${name.replaceAll("-", "_")}.v1`;
  const projection = { target, command_id: commandID(target), status: "pass", failure_class: null, failure_reason: null, unit_ids: units.map((unit) => unit.unit_id), evidence_refs: units.flatMap((unit) => unit.current_run_evidence_outputs) };
  const run = {
    manifest: { run_id: "synthetic", target, source_digest: digest, graph_digest: digest, declared_inputs: target === "test-slice" ? { OWNER: owner } : {} },
    summary: { status: "pass", failure_class: null, failure_reason: null },
    registrations: new Map(units.map((unit, i) => [unit.unit_id, { ...unit, row_ids: [rows[i].row_id] }])),
    started: new Map(units.map((unit) => [unit.unit_id, 0])),
    terminal: new Map(units.map((unit) => [unit.unit_id, { status: "passed" }])),
    targetSummaries: new Map([[target, projection]]),
  };
  const plan = { graph: { graph_digest: digest, units }, projections: { [target]: units.map((unit) => unit.unit_id) } };
  const options = {
    owner, manifestFile: path.join(root, "manifest.json"), sourceDigest: digest,
    manifest: { schema_id: "cartulary.harness_evidence_root_manifest.v1", owner_id: owner, entries: [{ target, run_root: "run" }] },
    surface: { targets: ["test-slice", "service-backed-test-slice", "backend-unit", "browser-e2e-measurement"].map((name) => ({ name, command_id: commandID(name), target_class: "public" })) },
    catalog: { rows, test_catalog_digest: digest, verification: { routing_digest: digest } },
    compileRun: () => plan, loadRun: async (requested) => { assert.equal(requested, path.join(root, "run")); return run; },
  };
  return { root, options, run, plan, rows, units, projection, cleanup, rowResult, unitResult, save };
}

test("owner audit accepts a complete owner or explicit target partition without opening unused roots", async (t) => {
  const f = fixture(t);
  f.options.manifest.entries.unshift({ target: "browser-e2e-measurement", run_root: "unavailable" });
  const report = await auditOwnerEvidence(f.options);
  assert.equal(report.status, "pass");
  assert.deepEqual(report.accepted[0].row_ids, f.rows.map((row) => row.row_id));
  assert.deepEqual(report.unused_inputs, [f.options.manifest.entries[0]]);
  const leaf = fixture(t, "backend-unit");
  assert.equal((await auditOwnerEvidence(leaf.options)).selection, "target_partitions");
});

test("owner audit admits required support projections from a public provider", async (t) => {
  for (const target of ["backend-integration-support", "browser-e2e-support"]) {
    const f = fixture(t, target);
    const provider = { name: "release-check", command_id: "cartulary.harness.command.release_check.v2", target_class: "public" };
    const support = { name: target, target_class: target === "backend-integration-support" ? "check_internal" : "internal_helper",
      ...(target === "browser-e2e-support" ? { command_id: "cartulary.harness.command.browser_e2e_support.v2" } : {}) };
    f.options.surface.targets.push(provider, support);
    Object.assign(f.run.manifest, { target: provider.name, command_id: provider.command_id, declared_inputs: {} });
    f.projection.command_id = support.command_id ?? provider.command_id;
    for (const row of f.rows) {
      if (target === "backend-integration-support") Object.assign(row, { family_id: `${owner}.support_integration`, evidence_class: "integration", runtime_profile_id: "default" });
      else Object.assign(row, { runner: "playwright", selector: { stage: "support" }, evidence_class: "browser", runtime_profile_id: "default" });
      f.units.find((unit) => unit.row_bindings[0].row_id === row.row_id).row_bindings[0].runner = row.runner;
      f.save(`rows/${row.row_id}.json`, { ...f.rowResult(row), runner: row.runner });
    }
    assert.equal((await auditOwnerEvidence(f.options)).status, "pass");
    f.projection.command_id = "cartulary.harness.command.unrelated.v1";
    await assert.rejects(auditOwnerEvidence(f.options), /target projection/u);
    f.projection.command_id = support.command_id ?? provider.command_id;
    f.run.manifest.target = target;
    await assert.rejects(auditOwnerEvidence(f.options), /public invocation/u);
    f.options.surface.targets.push({ name: "unrelated-private", target_class: "internal_helper", command_id: "cartulary.harness.command.unrelated.v1" });
    f.options.manifest.entries.push({ target: "unrelated-private", run_root: "unread" });
    f.options.loadRun = () => assert.fail("unowned private target must reject before evidence reads");
    await assert.rejects(auditOwnerEvidence(f.options), (error) => error.exit_code === 2);
  }
});

test("owner audit accepts cache reuse from canonical events without inventing a child execution", async (t) => {
  const f = fixture(t);
  const events = [];
  for (const [index, unit] of f.units.entries()) {
    const kinds = ["queued", "eligible", "wait_started", "wait_ended",
      ...(index === 0 ? ["cache_hit"] : ["cache_miss", "admitted", "started"]), "completed"];
    for (const kind of kinds) {
      events.push({ schema_id: "cartulary.harness_unit_event.v4", seq: events.length + 1,
        monotonic_ms: index, event: kind, unit_id: unit.unit_id,
        status: kind === "completed" ? "passed" : "pending", needs: [], resource_claims: {}, service_dependencies: [],
        ...(kind === "queued" ? { unit_kind: "runner", row_ids: [f.rows[index].row_id] } : {}),
        ...(kind.startsWith("wait_") ? { wait_reason: "rank", blocking_resources: [], blocking_unit_ids: [] } : {}),
        ...(kind.startsWith("cache_") ? { cache_profile_id: "fixture", cache_reason: "fixture" } : {}),
      });
    }
  }
  const eventFile = path.join(f.root, "run/unit-events.ndjson");
  atomicLocalFile(eventFile, `${events.map((event) => JSON.stringify(event)).join("\n")}\n`);
  const facts = await reduceCanonicalUnitIntervals(eventFile);
  f.run.terminal = facts.terminals;
  f.run.started = facts.starts;
  const unit = f.units[0];
  f.save(`unit-results/${unit.unit_id}.json`, { ...f.unitResult(unit), exit_code: null });
  assert.equal(f.run.started.has(unit.unit_id), false);
  assert.equal((await auditOwnerEvidence(f.options)).status, "pass");
});

test("owner audit rejects ambiguous and incomplete input partitions before reading roots", async (t) => {
  for (const entries of [
    [{ target: "unknown", run_root: "run" }],
    [{ target: "test-slice", run_root: "run" }, { target: "test-slice", run_root: "other" }],
    [{ target: "test-slice", run_root: "run" }, { target: "backend-unit", run_root: "other" }],
    [{ target: "service-backed-test-slice", run_root: "run" }],
  ]) {
    const f = fixture(t); f.options.manifest.entries = entries;
    f.options.loadRun = () => { assert.fail("invalid manifest must not read evidence"); };
    await assert.rejects(auditOwnerEvidence(f.options), (error) => error.exit_code === 2);
  }
});

test("owner audit rejects incompatible source, selection, results, cleanup and profiles", async (t) => {
  const mutations = [
    (f) => { f.run.manifest.source_digest = foreign; },
    (f) => { f.run.manifest.graph_digest = foreign; },
    (f) => { f.run.manifest.declared_inputs.ROWS = f.rows[0].row_id; },
    (f) => { f.run.manifest.declared_inputs.SERVICE_BACKED_ONLY = "1"; },
    (f) => { f.run.summary.status = "fail"; },
    (f) => { f.projection.status = "fail"; },
    (f) => { f.run.terminal.get("u0").status = "failed"; },
    (f) => { f.run.registrations.get("u0").row_ids = []; },
    (f) => { f.run.registrations.get("u0").kind = "policy"; },
    (f) => { f.run.registrations.get("u0").needs = ["u1"]; },
    (f) => { f.run.registrations.get("u0").resource_claims = { cpu: 2 }; },
    (f) => { f.run.registrations.get("u0").service_dependencies = ["postgres"]; },
    (f) => { f.units[0].row_bindings[0].runner = "vitest"; },
    (f) => { f.projection.unit_ids = ["u0"]; },
    (f) => { f.projection.evidence_refs = []; },
    (f) => { f.save("cleanup-results.json", { ...f.cleanup, run_id: "foreign" }); },
    (f) => { f.save("cleanup-results.json", { ...f.cleanup, results: [] }); },
    (f) => { f.cleanup.results[0].outcome = "blocked"; f.save("cleanup-results.json", f.cleanup); },
    (f) => { f.save("unit-results/u0.json", { ...f.unitResult(f.units[0]), semantic_digest: foreign }); },
    (f) => { f.save("unit-results/u0.json", { ...f.unitResult(f.units[0]), exit_code: null }); },
    (f) => { f.save("unit-results/u0.json", { ...f.unitResult(f.units[0]), evidence_outputs: [] }); },
    (f) => { f.save(`rows/${f.rows[0].row_id}.json`, { ...f.rowResult(f.rows[0]), runner: "vitest" }); },
    (f) => { f.save(`rows/${f.rows[0].row_id}.json`, { ...f.rowResult(f.rows[0]), terminal_state: "failed", exit_code: 1, failure_class: "product", failure_reason: "test_assertion_failure" }); },
    (f) => { f.save(`rows/${f.rows[0].row_id}.json`, { ...f.rowResult(f.rows[0]), schema_id: "cartulary.harness_row_result.v2" }); },
    (f) => { rmSync(path.join(f.root, "run/rows", `${f.rows[0].row_id}.json`)); },
    (f) => { const file = path.join(f.root, "run/rows", `${f.rows[0].row_id}.json`); rmSync(file); symlinkSync("/dev/null", file); },
  ];
  for (const mutate of mutations) {
    const f = fixture(t); mutate(f);
    await assert.rejects(auditOwnerEvidence(f.options), (error) => error.exit_code === 11 && error.auditReport.status === "fail" && error.auditReport.rejected.length === 1);
  }
});
