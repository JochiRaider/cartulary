#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadTestCatalog, targetForCatalogRow } from "../../test-catalog/index.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";
import { adaptGoInvocation, buildGoInvocations } from "./go.mjs";
import { adaptShellInvocation, buildShellInvocations } from "./shell.mjs";
import { adaptVitestInvocation, buildVitestInvocations } from "./vitest.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../../..");

function usage() {
  return "usage: row-runner-cli.mjs --row-id <row-id> | --row-ids <row-id,...>";
}

function parseArgs(argv) {
  if (argv.length !== 2 || !new Set(["--row-id", "--row-ids"]).has(argv[0]) || !argv[1]) {
    throw new Error(usage());
  }
  const rowIDs = argv[1].split(",");
  if (rowIDs.some((rowID) => !rowID) || new Set(rowIDs).size !== rowIDs.length) {
    throw new Error(usage());
  }
  return rowIDs.sort();
}

function readTaskCommandTargets() {
  const taskSurface = JSON.parse(
    readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"),
  );
  const targets = new Map(
    taskSurface.targets
      .filter((target) => target.command_id)
      .map((target) => [target.command_id, target.name]),
  );
  return targets;
}

function runRoot() {
  const resultsDir = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!resultsDir || !runID) {
    throw new Error("row runner requires CARTULARY_TEST_RESULTS_DIR and CARTULARY_TEST_RUN_ID");
  }
  return path.resolve(root, resultsDir, runID);
}

function writeResult(result) {
  validateSchemaSync(result.schema_id, result);
  const output = path.join(runRoot(), "rows", `${result.row_id}.json`);
  mkdirSync(path.dirname(output), { recursive: true, mode: 0o700 });
  writeFileSync(output, `${JSON.stringify(result, null, 2)}\n`, { mode: 0o600 });
}

function execute(invocation) {
  const child = spawnSync(invocation.command, invocation.args, {
    cwd: root,
    env: process.env,
    encoding: "utf8",
    maxBuffer: 128 * 1024 * 1024,
  });
  if (child.error) throw child.error;
  return {
    status: child.status ?? 11,
    signal: child.signal,
    stdout: child.stdout ?? "",
    stderr: child.stderr ?? "",
  };
}

function invocationsForRows(rows) {
  const runners = new Set(rows.map((row) => row.runner));
  if (runners.size !== 1) throw new Error("row runner unit must contain one runner kind");
  const runner = rows[0].runner;
  const workers = Number.parseInt(
    runner === "vitest"
      ? process.env.VITEST_MAX_WORKERS ?? "4"
      : process.env.CARTULARY_UNIT_CPU_TOKENS ?? "1",
    10,
  );
  if (!Number.isInteger(workers) || workers < 1 || workers > 64) {
    throw new Error("row runner worker count is invalid");
  }
  if (runner === "go") {
    return {
      invocations: buildGoInvocations(rows, workers, process.env.GO || "go"),
      adapt: adaptGoInvocation,
    };
  }
  if (runner === "vitest") {
    const command = process.env.PNPM || path.join(root, "tmp/node-runtime/bin/pnpm");
    return {
      invocations: buildVitestInvocations(root, rows, workers, command),
      adapt: adaptVitestInvocation,
    };
  }
  if (runner === "shell") {
    if (rows.length !== 1) throw new Error("shell units execute exactly one row");
    const row = rows[0];
    const commandTargets = readTaskCommandTargets();
    const targetName = targetForCatalogRow(row, { commandTargetByID: commandTargets });
    const shellRow = { ...row, target_name: targetName };
    return {
      invocations: buildShellInvocations([shellRow], process.env.MAKE || "make"),
      adapt: adaptShellInvocation,
    };
  }
  throw new Error(`row runner does not execute ${row.runner} rows`);
}

function canonicalFailure(result) {
  if (result.failure_class && result.failure_reason) {
    return {
      failure_class: result.failure_class,
      failure_reason: result.failure_reason,
      exit_code: result.failure_class === "product"
        ? 10
        : result.failure_class === "infra"
          ? 3
          : result.failure_class === "interrupted"
            ? 130
            : result.failure_reason === "fixture_error"
              ? 3
              : 11,
    };
  }
  switch (result.terminal_state) {
    case "failed":
      return {
        failure_class: "product",
        failure_reason: "test_assertion_failure",
        exit_code: 10,
      };
    case "infrastructure_failed":
      return {
        failure_class: "infra",
        failure_reason: "preflight_error",
        exit_code: 3,
      };
    case "cancelled":
      return {
        failure_class: "interrupted",
        failure_reason: "cancelled_or_interrupted",
        exit_code: 130,
      };
    default:
      return {
        failure_class: "harness",
        failure_reason: "test_accounting_unmapped",
        exit_code: 11,
      };
  }
}

function main() {
  const rowIDs = parseArgs(process.argv.slice(2));
  const catalog = loadTestCatalog(root);
  const rows = rowIDs.map((rowID) => {
    const row = catalog.rowByID.get(rowID);
    if (!row) throw new Error(`unknown active row ${rowID}`);
    return row;
  });
  if (rows.some((row) => row.runner === "playwright")) {
    throw new Error("Playwright rows must execute through a browser group unit");
  }
  const { invocations, adapt } = invocationsForRows(rows);
  const startedAt = new Date().toISOString();
  const started = Date.now();
  const results = [];
  for (const invocation of invocations) {
    const execution = execute(invocation);
    results.push(...adapt(invocation, execution));
    if (execution.status !== 0) {
      if (execution.stdout) process.stderr.write(execution.stdout);
      if (execution.stderr) process.stderr.write(execution.stderr);
    }
  }
  for (const result of results) {
    writeResult({
      schema_id: "cartulary.harness_row_result.v2",
      ...result,
      runner: rows[0].runner,
      started_at: startedAt,
      finished_at: new Date().toISOString(),
      wall_duration_ms: Date.now() - started,
    });
    if (result.terminal_state !== "passed") {
      const failure = canonicalFailure(result);
      process.stderr.write(
        `[FAIL] row=${result.row_id} runner=${rows[0].runner} failure_class=${failure.failure_class} failure_reason=${failure.failure_reason} row_reason=${result.failure_reason ?? "unknown"}\n`,
      );
    }
  }
  if (results.every((result) => result.terminal_state === "passed")) return 0;
  const failures = results
    .filter((result) => result.terminal_state !== "passed")
    .map(canonicalFailure);
  if (failures.some((failure) => failure.failure_class === "product")) return 10;
  if (failures.some((failure) => failure.failure_class === "infra")) return 3;
  if (failures.some((failure) => failure.failure_class === "interrupted")) return 130;
  if (failures.some((failure) => failure.failure_reason === "fixture_error")) return 3;
  return 11;
}

try {
  process.exitCode = main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = error.message === usage() ? 2 : 11;
}
