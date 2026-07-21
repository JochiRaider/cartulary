import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { collectCompatibleCaptureGroups } from "../go-target-aggregate.mjs";
import { collectGoShardPlanFromRows } from "../go-shard-plan.mjs";
import { collectTargetPlanRows } from "../target-plan.mjs";
import { resolveBackendWorkerPool } from "../target-execution/worker-policy.mjs";

const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-service-go-batching-"));

function symbol(index) {
  return `TestSyntheticBatch_${String(index).padStart(3, "0")}`;
}

function row(index, overrides = {}) {
  const selectedSymbol = overrides.symbol ?? symbol(index);
  return {
    target: "backend-store",
    service_backed: true,
    runner_family: "go_test",
    id: `U-SYN-${String(index).padStart(3, "0")}`,
    owner_id: "module.recovery",
    section: "unit",
    coverage: "authoritative",
    execution_dependency: "backend_store",
    evidence_class: "product_conformance",
    layer: "backend_store",
    default_check_required: true,
    primary_evidence_owner: `U-SYN-${String(index).padStart(3, "0")}`,
    execution_family: "backend-store-synthetic-batch",
    execution_label: "backend-store synthetic batch",
    packages: [overrides.package ?? "./internal/modules/synthetic"],
    support_only: false,
    package: overrides.package ?? "./internal/modules/synthetic",
    symbols: [selectedSymbol],
    scenario_symbols: { [`SCN-${String(index).padStart(3, "0")}`]: selectedSymbol },
    runtime_binaries: overrides.runtime_binaries ?? [],
    shard_isolation: overrides.shard_isolation ?? false,
    fixture_policy: { postgres: overrides.fixture_policy ?? "transaction" },
    fixture_budget: {
      postgres: overrides.fixture_budget ?? { max_transactions: 8 },
    },
    ...overrides.row,
  };
}

function plan(rows, weights = {}) {
  const tests = {};
  for (const selectedRow of rows) {
    for (const selectedSymbol of selectedRow.symbols) {
      const importPath = `example.test/batching/${selectedRow.package.slice(2)}`;
      tests[`${importPath}::${selectedSymbol}`] = weights[selectedSymbol] ?? 1_000;
    }
  }
  writeFileSync(
    path.join(root, "tools/go_test_duration_baselines.json"),
    `${JSON.stringify({
      schema_id: "cartulary.go_test_duration_baselines.v4",
      default_shard_target_ms: 9_000,
      shard_target_ms_by_target: { "backend-store": 9_000 },
      default_item_weight_ms: 10_000,
      command_overheads_by_target: { "backend-store": 20_000 },
      package_overheads: {},
      fixture_overheads_by_package: {},
      fixture_overheads_by_test: {},
      raw_aggregates: {},
      tests,
    })}\n`,
  );
  return collectGoShardPlanFromRows(root, rows, { defaultCheckOnly: true });
}

function assertBoundedGrowth(count, expectedShards) {
  const rows = Array.from({ length: count }, (_, index) => row(index + 1));
  const first = plan(rows);
  const second = plan([...rows].reverse());
  assert.equal(first.shards.length, expectedShards);
  assert.equal(first.shards.reduce((sum, shard) => sum + shard.item_count, 0), count);
  assert.ok(first.shards.every((shard) => shard.item_count <= 8));
  assert.ok(first.shards.every((shard) => shard.work_weight_ms <= 12_000));
  assert.ok(first.shards.every((shard) => shard.shard_target_ms === 12_000));
  assert.equal(JSON.stringify(first), JSON.stringify(second));
  return Buffer.byteLength(JSON.stringify(first));
}

try {
  mkdirSync(path.join(root, "tools"), { recursive: true });
  writeFileSync(path.join(root, "go.mod"), "module example.test/batching\n\ngo 1.25\n");

  const bytes25 = assertBoundedGrowth(25, 4);
  const bytes100 = assertBoundedGrowth(100, 13);
  assert.ok(bytes100 - bytes25 < 1_500 * 75, "plan growth must remain linear and bounded");

  const compatible = [
    row(201),
    row(202),
    row(203, { package: "./internal/modules/other" }),
    row(204, { runtime_binaries: ["server"] }),
    row(205, { fixture_budget: { max_transactions: 9 } }),
    row(206, { fixture_policy: "template_clone", fixture_budget: { max_template_clones: 1 } }),
    row(207, { shard_isolation: true }),
  ];
  const compatibilityPlan = plan(compatible);
  assert.equal(compatibilityPlan.shards.length, 6);
  const paired = compatibilityPlan.shards.find((shard) => shard.item_count === 2);
  assert.deepEqual(
    paired.items.map((item) => item.id).sort(),
    ["U-SYN-201", "U-SYN-202"],
  );
  assert.ok(compatibilityPlan.shards.every((shard) => shard.regex.startsWith("^") && shard.regex.endsWith("$")));
  assert.ok(
    compatibilityPlan.shards.flatMap((shard) => shard.items).every(
      (item) => item.id && item.scenario_id && item.primary_evidence_owner && item.baseline_key,
    ),
  );
  const compatibleCaptureGroups = collectCompatibleCaptureGroups(compatible);
  assert.equal(compatibleCaptureGroups.length, 6);
  assert.equal(compatibleCaptureGroups.filter((group) => group.rows.length === 2).length, 1);

  const currentBackendRows = collectTargetPlanRows(process.cwd())
    .filter((selectedRow) => selectedRow.target === "backend-unit");
  const currentCaptureGroups = collectCompatibleCaptureGroups(currentBackendRows);
  assert.equal(currentCaptureGroups.length, 30, "backend-unit physical process plan");
  assert.equal(currentCaptureGroups.reduce((sum, group) => sum + group.exact_symbol_count, 0), 251);
  assert.equal(currentCaptureGroups.filter((group) => group.raw).length, 1);
  assert.equal(new Set(currentCaptureGroups.flatMap((group) => group.families)).size, 34);
  assert.deepEqual(resolveBackendWorkerPool(24, currentCaptureGroups.length), {
    workers: 6,
    goMaxProcs: 4,
  });
  assert.deepEqual(resolveBackendWorkerPool(3, 30), { workers: 1, goMaxProcs: 3 });
  assert.deepEqual(resolveBackendWorkerPool(64, 3), { workers: 3, goMaxProcs: 21 });
  assert.throws(() => resolveBackendWorkerPool(0, 1), /invalid available parallelism/u);
  assert.throws(
    () => collectCompatibleCaptureGroups([
      row(401, { symbol: "TestDuplicateSymbol" }),
      row(402, { symbol: "TestDuplicateSymbol" }),
    ]),
    /duplicates an exact symbol/u,
  );

  const oversizedRows = [row(301), row(302)];
  const oversizedPlan = plan(oversizedRows, { [symbol(301)]: 13_000, [symbol(302)]: 1_000 });
  const oversized = oversizedPlan.shards.find((shard) => shard.items.some((item) => item.symbol === symbol(301)));
  assert.equal(oversized.item_count, 1);
  assert.equal(oversized.work_weight_ms, 13_000);

  process.stdout.write(
    `service Go batching smoke passed: 25_rows=4_shards 100_rows=13_shards backend_unit_processes=30 backend_unit_exact_symbols=251 workers=6 gomaxprocs=4 bytes25=${bytes25} bytes100=${bytes100}\n`,
  );
} finally {
  rmSync(root, { recursive: true, force: true });
}
