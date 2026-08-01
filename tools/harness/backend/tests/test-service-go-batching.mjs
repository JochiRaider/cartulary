import assert from "node:assert/strict";
import {
  mkdtempSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";
import { Worker } from "node:worker_threads";

import { collectCompatibleCaptureGroups } from "../go-target-aggregate.mjs";
import { collectGoShardPlanFromRows } from "../go-shard-plan.mjs";
import { serviceExactShardProfile } from "../go-shard-policy.mjs";
import { collectTargetPlanRows } from "../target-plan.mjs";
import {
  resolveBackendCapturePool,
  resolveBackendWorkerPool,
} from "../target-execution/worker-policy.mjs";
import {
  createUnshardedFamilyReport,
  parsePhysicalReport,
} from "../target-execution/reports.mjs";
import {
  partitionEmissionRequests,
  runSettledBounded,
} from "../target-execution/summary-emission.mjs";

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
      schema_id: "cartulary.go_test_duration_baselines.v5",
      default_shard_target_ms: 9_000,
      shard_target_ms_by_target: { "backend-store": 9_000 },
      default_item_weight_ms: 10_000,
      default_package_overhead_ms: 100,
      default_command_overhead_ms: 70_000,
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
  assert.deepEqual(serviceExactShardProfile("backend-integration"), {
    max_symbols: 8,
    max_estimated_test_work_ms: 6_000,
  });
  assert.deepEqual(serviceExactShardProfile("backend-store"), {
    max_symbols: 8,
    max_estimated_test_work_ms: 12_000,
  });

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
  assert.deepEqual(resolveBackendCapturePool("backend-unit", 24, 30), {
    workers: 6,
    goMaxProcs: 4,
  });
  assert.deepEqual(resolveBackendCapturePool("backend-process", 24, 6), {
    workers: 6,
    goMaxProcs: 24,
  });
  const currentProcessRows = collectTargetPlanRows(process.cwd())
    .filter((selectedRow) => selectedRow.target === "backend-process");
  const currentProcessPlan = collectGoShardPlanFromRows(process.cwd(), currentProcessRows);
  const currentProcessShards = currentProcessPlan.shards.filter((shard) =>
    shard.items.some((item) => item.target === "backend-process"));
  assert.equal(currentProcessShards.length, 4, "backend-process physical process plan");
  assert.equal(currentProcessShards.reduce((sum, shard) => sum + shard.item_count, 0), 36);
  assert.deepEqual(
    currentProcessShards.map((shard) => [shard.aggregate_name, shard.item_count]),
    [
      ["app.server.process", 16],
      ["module.recovery.process", 6],
      ["app.server.process", 12],
      ["module.extensions.process", 2],
    ],
  );
  assert.ok(currentProcessShards.every((shard) => shard.item_count <= 16));
  assert.ok(currentProcessShards.every((shard) => shard.work_weight_ms <= 24_000));
  assert.ok(currentProcessShards.every((shard) => shard.shard_target_ms === 24_000));
  assert.deepEqual(resolveBackendCapturePool("backend-process", 24, currentProcessShards.length), {
    workers: 4,
    goMaxProcs: 24,
  });
  assert.throws(() => resolveBackendWorkerPool(0, 1), /invalid available parallelism/u);
  assert.throws(
    () => collectCompatibleCaptureGroups([
      row(401, { symbol: "TestDuplicateSymbol" }),
      row(402, { symbol: "TestDuplicateSymbol" }),
    ]),
    /duplicates an exact symbol/u,
  );

  const physicalReportDir = path.join(root, "physical-report");
  mkdirSync(physicalReportDir, { recursive: true });
  const physicalFiles = {
    "runner.jsonl": `${JSON.stringify({ Action: "pass", Test: "TestSynthetic" })}\n`,
    "stderr.log": "",
    "command.txt": "env go test -json -run '^TestSynthetic$' ./internal/modules/synthetic\n",
    "start_time.txt": "2026-01-01T00:00:00Z\n",
    "end_time.txt": "2026-01-01T00:00:01Z\n",
    "duration_ms.txt": "1000\n",
    "exit_status.txt": "0\n",
  };
  for (const [name, value] of Object.entries(physicalFiles)) {
    writeFileSync(path.join(physicalReportDir, name), value);
  }
  const parsedPhysical = parsePhysicalReport(physicalReportDir);
  assert.ok(Object.isFrozen(parsedPhysical));
  writeFileSync(path.join(physicalReportDir, "runner.jsonl"), "not-json\n");
  const reportContext = {
    resultsRoot: path.join(root, "results"),
    runId: "parse-once",
  };
  const firstProjection = createUnshardedFamilyReport(reportContext, "family-one", [
    Object.freeze({ name: "physical", usage: "actual", report: parsedPhysical }),
  ]);
  const secondProjection = createUnshardedFamilyReport(reportContext, "family-two", [
    Object.freeze({ name: "physical", usage: "actual", report: parsedPhysical }),
  ]);
  const projectedFiles = [
    "runner.jsonl",
    "stderr.log",
    "command.txt",
    "start_time.txt",
    "end_time.txt",
    "duration_ms.txt",
    "wall_duration_ms.txt",
    "exit_status.txt",
  ];
  for (const name of projectedFiles) {
    assert.equal(
      readFileSync(path.join(firstProjection.reportDir, name), "utf8"),
      readFileSync(path.join(secondProjection.reportDir, name), "utf8"),
      `${name} projection bytes`,
    );
  }
  assert.match(
    readFileSync(path.join(firstProjection.reportDir, "runner.jsonl"), "utf8"),
    /TestSynthetic/u,
  );
  assert.throws(() => parsePhysicalReport(physicalReportDir), /malformed JSON/u);

  let activeWorkers = 0;
  let maximumWorkers = 0;
  const settled = await runSettledBounded([0, 1, 2, 3, 4], 2, async (value) => {
    activeWorkers += 1;
    maximumWorkers = Math.max(maximumWorkers, activeWorkers);
    await new Promise((resolve) => setTimeout(resolve, value % 2 === 0 ? 4 : 1));
    activeWorkers -= 1;
    if (value === 1 || value === 3) throw new Error(`worker-${value}`);
    return `value-${value}`;
  });
  assert.equal(maximumWorkers, 2);
  assert.deepEqual(
    settled.map((result) => result.error?.message ?? result.value),
    ["value-0", "worker-1", "value-2", "worker-3", "value-4"],
  );
  assert.ok(settled.every(Object.isFrozen));

  const emissionBatches = partitionEmissionRequests(
    Array.from({ length: 34 }, (_, index) => `family-${index}`),
    6,
  );
  assert.deepEqual(emissionBatches.map((batch) => batch.length), [6, 6, 6, 6, 5, 5]);
  assert.deepEqual(
    emissionBatches.flat().map((entry) => entry.index).sort((left, right) => left - right),
    Array.from({ length: 34 }, (_, index) => index),
  );
  assert.throws(() => partitionEmissionRequests(["family"], 0), /invalid report emission worker count/u);

  const workerFixture = await new Promise((resolve, reject) => {
    const worker = new Worker(
      new URL("../target-execution/report-emission-worker.mjs", import.meta.url),
      {
        workerData: {
          entries: [
            { index: 0, request: { emissions: [{ catalogAware: false, env: {} }] } },
            { index: 1, request: { emissions: [] } },
          ],
        },
      },
    );
    worker.once("message", resolve);
    worker.once("error", reject);
  });
  assert.match(workerFixture.results[0].error.message, /missing required environment variable/u);
  assert.equal(workerFixture.results[1].error, null);
  assert.equal(workerFixture.results[1].status, 0);

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
