import assert from "node:assert/strict";
import test from "node:test";
import path from "node:path";
import { rmSync, symlinkSync } from "node:fs";
import { semanticJSONDigest, validateSchemaSync } from "../../contract/index.mjs";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { buildWorkGraph } from "../../scheduler/work-graph/model.mjs";
import { createScratch } from "../../test-support/scratch.mjs";
import { readUnitSelection } from "../runners/unit-selection.mjs";

// This fixture owns its complete published inputs; no enclosing-run state is read.
function publishedSelection(t, runner) {
  const root = createScratch(t, "published-selection-");
  const unit = (id, rows) => ({ unit_id: id, owner_id: "harness.command_surface", kind: "runner",
    command: { executable: "node", args: rows.length ? ["tools/harness/execution/runners/row-runner-cli.mjs", "--unit-rows"] : [], environment: {} },
    needs: [], resource_claims: { cpu: 1, process: 1 }, fixture_lease: "none", service_dependencies: [],
    cache_policy: "none", timeout_ms: 1000, current_run_evidence_outputs: rows.map(({ row_id }) => `rows/${row_id}.json`),
    row_bindings: rows, failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" }, estimated_work_ms: 1 });
  const rows = [{ row_id: "harness.command_surface.unit.first", runner }, { row_id: "harness.command_surface.unit.second", runner }];
  const graph = buildWorkGraph([unit("selected", rows), unit("preparation", [])]);
  const digest = `sha256:${"a".repeat(64)}`;
  const manifest = { schema_id: "cartulary.harness_run_manifest.v3", run_id: "selection-fixture",
    target: "test-slice", command_id: "cartulary.harness.command.test_slice.v5", declared_inputs: {},
    source_commit: "a".repeat(40), source_state: "dirty", source_digest: digest, toolchain_digest: digest, system_digest: digest,
    capability_snapshot: { schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 2, memory_bytes: 1024,
      process_slots: 2, io_tokens: 1, postgres_lanes: 1, object_store_lanes: 1, writable_volume: true,
      port_lanes: 1, services: {}, sources: {}, snapshot_digest: digest },
    go_worker_budget: 2, graph_digest: graph.graph_digest, cache_mode: "off", started_at: "2026-10-26T00:00:00.000Z",
    instrumentation: { mode: "off", policy_digest: digest } };
  const context = { schema_id: "cartulary.harness_execution_context.v3", manifest_digest: semanticJSONDigest(manifest), graph,
    projections: { "test-slice": { command_id: manifest.command_id, unit_ids: ["preparation", "selected"] } },
    source_inputs: [], host_profile: { platform: "linux", architecture: "x64", kernel_release: "fixture", cpu_models: ["fixture"] },
    instrumentation_policy_bytes: "{}" };
  validateSchemaSync(manifest.schema_id, manifest);
  validateSchemaSync(context.schema_id, context);
  const save = (name, value) => atomicLocalFile(path.join(root, name), typeof value === "string" ? value : JSON.stringify(value), { replace: true });
  save("run-manifest.json", manifest);
  save("execution-context.json", context);
  const read = (run = manifest.run_id, id = "selected") => readUnitSelection(root, run, id);
  assert.deepEqual(read(), ["harness.command_surface.unit.first", "harness.command_surface.unit.second"], "positive control before corruption");
  return { root, manifest, context, save, read };
}

for (const runner of ["go", "vitest"]) {
  test(`${runner} published selection rejects foreign runs, absent units and preparation`, (t) => {
    const fixture = publishedSelection(t, runner);
    assert.throws(() => fixture.read("foreign-run"), /bind/u);
    assert.throws(() => fixture.read(fixture.manifest.run_id, "absent"), /published work unit/u);
    assert.throws(() => fixture.read(fixture.manifest.run_id, "preparation"), /exact Go or Vitest/u);
  });
}
for (const [name, corrupt] of [
  ["row mutation", ({ context, save }) => { context.graph.units.find((unit) => unit.unit_id === "selected").row_bindings.pop(); save("execution-context.json", context); }],
  ["manifest mismatch", ({ manifest, save }) => { manifest.graph_digest = `sha256:${"b".repeat(64)}`; save("run-manifest.json", manifest); }],
  ["truncated manifest", ({ save }) => save("run-manifest.json", "{")],
  ["missing manifest", ({ root }) => rmSync(path.join(root, "run-manifest.json"))],
  ["symlink manifest", ({ root, manifest, save }) => { save("another-manifest.json", manifest); rmSync(path.join(root, "run-manifest.json")); symlinkSync(path.join(root, "another-manifest.json"), path.join(root, "run-manifest.json")); }],
]) {
  test(`published selection rejects ${name}`, (t) => {
    const fixture = publishedSelection(t, "vitest");
    corrupt(fixture);
    assert.throws(() => fixture.read());
  });
}
