#!/usr/bin/env node

import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../contract/index.mjs";
import { executeOwnerSlicePlan, retainOwnerSliceUnitLogs } from "./execution.mjs";
import { buildOwnerSlicePlan, OwnerSliceUsageError } from "./plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function usage() {
  return "usage: owner-slice --target <test-slice|service-backed-test-slice> --owner <owner-id> [--rows <row-id,...>] [--vitest-workers <1..16>] [--playwright-workers <1..16>] [--json]";
}

function parseArgs(argv) {
  const options = {
    target: "",
    ownerID: "",
    rows: "",
    rowsProvided: false,
    vitestWorkers: undefined,
    playwrightWorkers: undefined,
    jsonValue: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") options.target = argv[++index] ?? "";
    else if (arg === "--owner") options.ownerID = argv[++index] ?? "";
    else if (arg === "--rows") {
      options.rows = argv[++index] ?? "";
      options.rowsProvided = true;
    } else if (arg === "--vitest-workers") options.vitestWorkers = argv[++index] ?? "";
    else if (arg === "--playwright-workers") options.playwrightWorkers = argv[++index] ?? "";
    else if (arg === "--json") options.jsonValue = "1";
    else if (arg === "--json-value") options.jsonValue = argv[++index] ?? "";
    else throw new OwnerSliceUsageError(usage());
  }
  if (!new Set(["test-slice", "service-backed-test-slice"]).has(options.target)) {
    throw new OwnerSliceUsageError(usage());
  }
  return options;
}

function runID() {
  return process.env.CARTULARY_TEST_RUN_ID || `${new Date().toISOString().replaceAll(/[-:]/gu, "").replace(/\.\d{3}Z$/u, "Z")}-p${process.pid}`;
}

function resultsRoot() {
  return path.resolve(root, process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results");
}

function identityFields(plan) {
  return {
    command_id: plan.command_id,
    run_id: plan.run_id,
    owner_id: plan.owner_id,
    selected_rows: plan.selected_rows,
    source_snapshot_digest: plan.source_snapshot_digest,
    catalog_semantic_digest: plan.catalog_semantic_digest,
    verification_semantic_digest: plan.verification_semantic_digest,
    runtime_profile_digest: plan.runtime_profile_digest,
    resource_profile_digest: plan.resource_profile_digest,
    fixture_profile_digest: plan.fixture_profile_digest,
  };
}

function schedulerSummary(plan, execution, logs, startedAt, finishedAt) {
  const failedRows = execution.unit_results
    .filter((result) => result.status !== "passed")
    .flatMap((result) => result.row_ids)
    .sort();
  return {
    schema_id: "cartulary.test_slice_scheduler_summary.v1",
    ...identityFields(plan),
    target: plan.target,
    status: execution.status,
    started_at: startedAt,
    finished_at: finishedAt,
    duration_ms: execution.duration_ms,
    selection: plan.selection,
    counts: {
      selected: plan.selected_rows.length,
      passed: plan.selected_rows.length - failedRows.length,
      failed: failedRows.length,
      infrastructure_failed: 0,
      skipped_dependency: 0,
      cancelled: 0,
      skipped_authorized: 0,
    },
    unused_inputs: plan.unused_inputs,
    work_units: execution.unit_results.map((result) => ({
      work_unit_id: result.work_unit_id,
      runner: result.runner,
      target_name: result.target_name,
      row_ids: result.row_ids,
      terminal_state: result.status,
      exit_code: result.exit_code,
      duration_ms: result.duration_ms,
      failure_reason: result.failure_reason,
    })),
    artifacts: {
      plan: `${plan.target}/test-slice-plan.json`,
      scheduler_summary: `${plan.target}/test-slice-scheduler-summary.json`,
      work_unit_logs: logs,
    },
  };
}

function retainPlan(plan, targetDir) {
  validateSchemaSync(plan.schema_id, plan);
  writeFileSync(path.join(targetDir, "test-slice-plan.json"), `${JSON.stringify(plan, null, 2)}\n`, "utf8");
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const id = runID();
  const commandID = options.target === "test-slice"
    ? "cartulary.harness.command.test_slice.v1"
    : "cartulary.harness.command.service_backed_test_slice.v1";
  const startedAt = new Date().toISOString();
  const plan = buildOwnerSlicePlan(root, {
    ...options,
    dependencyScope: options.target === "test-slice" ? "all" : "service_backed",
    target: options.target,
    commandID,
    runID: id,
    timestamp: startedAt,
  });

  const targetDir = path.join(resultsRoot(), id, options.target);
  mkdirSync(targetDir, { recursive: true });
  retainPlan(plan, targetDir);
  const execution = executeOwnerSlicePlan(root, plan);
  const logs = retainOwnerSliceUnitLogs(targetDir, execution);
  const finishedAt = new Date().toISOString();
  const summary = schedulerSummary(plan, execution, logs, startedAt, finishedAt);
  validateSchemaSync(summary.schema_id, summary);
  writeFileSync(
    path.join(targetDir, "test-slice-scheduler-summary.json"),
    `${JSON.stringify(summary, null, 2)}\n`,
    "utf8",
  );
  if (options.jsonValue === "1") {
    process.stdout.write(`${JSON.stringify(summary)}\n`);
  } else if (summary.status === "pass") {
    process.stdout.write(
      `[RESULT] target=${summary.target} status=pass owner=${summary.owner_id} rows=${summary.counts.passed}/${summary.counts.selected} run_root=${path.relative(root, path.dirname(targetDir)).replaceAll("\\", "/")} scheduler_json=${summary.target}/test-slice-scheduler-summary.json\n`,
    );
  } else {
    process.stderr.write(
      `[FAIL] target=${summary.target} exit_code=10 failure_class=product reason=test_assertion_failure owner=${summary.owner_id} failed_rows=${summary.counts.failed}\n`,
    );
  }
  return summary.status === "pass" ? 0 : 10;
}

main()
  .then((status) => {
    process.exitCode = status;
  })
  .catch((error) => {
    if (error instanceof OwnerSliceUsageError || error?.exitCode === 2) {
      process.stderr.write(`${error.message}\n`);
      process.exitCode = 2;
      return;
    }
    process.stderr.write(`owner slice failed: ${error.message}\n`);
    process.exitCode = 11;
  });
