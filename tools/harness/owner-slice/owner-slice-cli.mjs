#!/usr/bin/env node

import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../contract/index.mjs";
import {
  buildTestEvidenceAccounting,
  buildTestOwnerSummary,
} from "../evidence-accounting/index.mjs";
import {
  buildToolRunSummary,
  fileArtifactRef,
  normalizeOutputMode,
} from "../output/index.mjs";
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
  const counts = {
    selected: execution.row_results.length,
    passed: 0,
    failed: 0,
    infrastructure_failed: 0,
    skipped_dependency: 0,
    cancelled: 0,
    skipped_authorized: 0,
  };
  for (const result of execution.row_results) counts[result.terminal_state] += 1;
  const ownerPrefix = `${plan.target}/owners/${plan.owner_id}`;
  return {
    schema_id: "cartulary.test_slice_scheduler_summary.v1",
    ...identityFields(plan),
    target: plan.target,
    status: execution.status,
    started_at: startedAt,
    finished_at: finishedAt,
    duration_ms: execution.duration_ms,
    selection: plan.selection,
    counts,
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
      evidence_accounting: `${ownerPrefix}/test-evidence-accounting.json`,
      owner_summary: `${ownerPrefix}/test-owner-summary.json`,
      plan: `${plan.target}/test-slice-plan.json`,
      scheduler_summary: `${plan.target}/test-slice-scheduler-summary.json`,
      tool_run_summary: `${plan.target}/tool-run-summary.json`,
      work_unit_logs: logs,
    },
  };
}

function ownerArtifactPaths(plan) {
  const prefix = `${plan.target}/owners/${plan.owner_id}`;
  return {
    evidence_accounting: `${prefix}/test-evidence-accounting.json`,
    owner_summary: `${prefix}/test-owner-summary.json`,
    plan: `${plan.target}/test-slice-plan.json`,
    scheduler_summary: `${plan.target}/test-slice-scheduler-summary.json`,
    tool_run_summary: `${plan.target}/tool-run-summary.json`,
  };
}

function retainJSON(file, value) {
  validateSchemaSync(value.schema_id, value);
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`, "utf8");
}

function toolSummary(plan, execution, paths, startedAt, finishedAt, resultRoot, runRoot) {
  const status = execution.status;
  const failure = execution.row_results.find((row) => row.terminal_state !== "passed");
  return buildToolRunSummary({
    target: plan.target,
    command: ["make", plan.target],
    status,
    exitCode: status === "pass" ? 0 : failure?.terminal_state === "failed" ? 10 : 11,
    startedAt,
    completedAt: finishedAt,
    durationMs: execution.duration_ms,
    outputMode: normalizeOutputMode(),
    resultRoot,
    runId: plan.run_id,
    runRoot,
    summaryArtifacts: [
      fileArtifactRef("test_evidence_accounting", paths.evidence_accounting),
      fileArtifactRef("test_owner_summary", paths.owner_summary),
      fileArtifactRef("test_slice_plan", paths.plan),
      fileArtifactRef("test_slice_scheduler_summary", paths.scheduler_summary),
      fileArtifactRef("tool_run_summary", paths.tool_run_summary),
    ],
    workUnits: execution.unit_results.map((result) => ({
      id: result.work_unit_id,
      completed: result.row_results.length,
      total: result.row_ids.length,
      status: result.status,
    })),
    counts: {
      tests: execution.row_results.length,
      failed: execution.row_results.filter((row) => row.terminal_state !== "passed").length,
    },
    failureClass: status === "pass" ? null : failure?.terminal_state === "failed" ? "product" : "artifact",
    failureReason: status === "pass" ? null : failure?.terminal_state === "failed" ? "test_assertion_failure" : "scheduler_accounting_error",
    failures: status === "pass" ? [] : [{
      target: plan.target,
      work_unit: plan.work_units.find((unit) => unit.row_ids.includes(failure.row_id))?.work_unit_id ?? "",
      failure_class: failure?.terminal_state === "failed" ? "product" : "artifact",
      failure_reason: failure?.terminal_state === "failed" ? "test_assertion_failure" : "scheduler_accounting_error",
      headline: failure?.failure_reason ?? "owner evidence accounting failed",
    }],
    rerunCommands: [`make ${plan.target} OWNER=${plan.owner_id}`],
  });
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
  const ownerDir = path.join(targetDir, "owners", plan.owner_id);
  mkdirSync(ownerDir, { recursive: true });
  const paths = ownerArtifactPaths(plan);
  const accounting = buildTestEvidenceAccounting(plan, execution, logs, startedAt, finishedAt);
  retainJSON(path.join(ownerDir, "test-evidence-accounting.json"), accounting);
  const ownerSummary = buildTestOwnerSummary(plan, accounting, paths);
  retainJSON(path.join(ownerDir, "test-owner-summary.json"), ownerSummary);
  const summary = schedulerSummary(plan, execution, logs, startedAt, finishedAt);
  retainJSON(
    path.join(targetDir, "test-slice-scheduler-summary.json"),
    summary,
  );
  const runRoot = path.dirname(targetDir);
  const retainedToolSummary = toolSummary(
    plan,
    execution,
    paths,
    startedAt,
    finishedAt,
    path.relative(root, resultsRoot()).replaceAll("\\", "/"),
    path.relative(root, runRoot).replaceAll("\\", "/"),
  );
  retainJSON(path.join(targetDir, "tool-run-summary.json"), retainedToolSummary);
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
