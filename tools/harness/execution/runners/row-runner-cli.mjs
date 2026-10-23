#!/usr/bin/env node

import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { goRunnerBuildContext, loadSelectedTestRows, targetForCatalogRow } from "../../test-catalog/index.mjs";
import { primaryPublicFailure, publicExitCodeForFailure, publicExitCodeForFailures, redactString, validateSchemaSync } from "../../contract/index.mjs";
import { createLaunchContext } from "../../runtime/launch-context.mjs";
import { beginActivity } from "../../runtime/local-activity.mjs";
import { createCommandFailureContext, reportCommandFailure } from "../../runtime/command-failure.mjs";
import { runPrivateCapturedProcess } from "../../runtime/private-child-process.mjs";
import { adaptGoInvocationFile, buildGoInvocations, goExecutionEnvironment } from "./go.mjs";
import { adaptShellInvocation, buildShellInvocations } from "./shell.mjs";
import { publishRowResult } from "./row-results.mjs";
import { readUnitSelection } from "./unit-selection.mjs";
import { adaptVitestInvocationFile, buildVitestInvocations, vitestInvocationFailure } from "./vitest.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../../..");

function usage() {
  return "usage: row-runner-cli.mjs --row-id <row-id> | --row-ids <row-id,...> | --unit-rows";
}

function parseArgs(argv) {
  if (argv.length === 1 && argv[0] === "--unit-rows") return readUnitSelection(runRoot(), process.env.CARTULARY_TEST_RUN_ID, process.env.CARTULARY_WORK_UNIT_ID);
  if (argv.length !== 2 || !new Set(["--row-id", "--row-ids"]).has(argv[0]) || !argv[1]) {
    throw new Error(usage());
  }
  const rowIDs = argv[1].split(",");
  if (rowIDs.some((rowID) => !rowID) || new Set(rowIDs).size !== rowIDs.length) {
    throw new Error(usage());
  }
  return rowIDs.sort();
}



function runRoot() {
  const resultsDir = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!resultsDir || !runID) {
    throw new Error("row runner requires CARTULARY_TEST_RESULTS_DIR and CARTULARY_TEST_RUN_ID");
  }
  return path.resolve(root, resultsDir, runID);
}

async function execute(invocation, producer) {
  const launch = createLaunchContext({ repoRoot: root, environment: process.env,
    unitID: process.env.CARTULARY_WORK_UNIT_ID || `row:${invocation.rows[0].row_id}`,
    rowIDs: invocation.rows.map((row) => row.row_id), producer });
  const context = launch.identity.command_id ? createCommandFailureContext({ repoRoot: root, environment: process.env, launch }) : null;
  let result;
  try {
    result = await runPrivateCapturedProcess(invocation.command, invocation.args, {
      launch, dependencies: JSON.parse(process.env.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES),
      cwd: root,
      env: { ...process.env, ...invocation.environment, ...context?.environment },
      repoRoot: root,
      runRoot: runRoot(),
      tailBytes: 1024 * 1024,
    });
  } catch (error) {
    launch.settle("spawn_failed");
    try { context?.close(); }
    catch { process.stderr.write("row command context cleanup failed (cleanup_error)\n"); }
    throw error;
  }
  const commandFailure = context?.read() ?? null;
  launch.settle(result.signal ? "cancelled" : result.status === 0 && !commandFailure ? "passed" : "failed");
  return { ...result, commandFailure, cleanup() {
    let failure;
    for (const release of [() => result.cleanup(), () => context?.close()]) {
      try { release(); }
      catch (error) { failure ??= error; }
    }
    if (failure) throw failure;
  } };
}

function invocationsForRows(rows) {
  const runners = new Set(rows.map((row) => row.runner));
  if (runners.size !== 1) throw new Error("row runner unit must contain one runner kind");
  const runner = rows[0].runner;
  const workers = Number.parseInt(
    runner === "vitest"
      ? process.env.VITEST_MAX_WORKERS ?? "1"
      : process.env.CARTULARY_UNIT_CPU_TOKENS ?? "1",
    10,
  );
  if (!Number.isInteger(workers) || workers < 1 || workers > 64) {
    throw new Error("row runner worker count is invalid");
  }
  if (runner === "go") {
    return {
      invocations: buildGoInvocations(rows, workers, process.env.GO || "go", { ...goRunnerBuildContext(root), goRace: process.env.CARTULARY_GO_RACE === "1", goShuffleSeed: process.env.CARTULARY_GO_SHUFFLE_SEED || undefined }).map((invocation) => ({
        ...invocation,
        // Explicit test credentials: production key loading has no default.
        environment: { ...goExecutionEnvironment(root), ...invocation.environment, CARTULARY_AUTH_MASTER_KEY: readFileSync(path.join(root, "internal/testutil/fixtures/auth/master-key.base64"), "utf8").trim() },
      })),
      adapt: adaptGoInvocationFile,
    };
  }
  if (runner === "vitest") {
    return {
      invocations: buildVitestInvocations(root, rows, runRoot()),
      adapt: adaptVitestInvocationFile,
    };
  }
  if (runner === "shell") {
    if (rows.length !== 1) throw new Error("shell units execute exactly one row");
    const row = rows[0];
    const targetName = targetForCatalogRow(row);
    const shellRow = { ...row, target_name: targetName };
    return {
      invocations: buildShellInvocations([shellRow]),
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
      exit_code: publicExitCodeForFailure(result),
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

async function main() {
  const rowIDs = parseArgs(process.argv.slice(2));
  const rows = loadSelectedTestRows(root, rowIDs);
  if (rows.some((row) => row.runner === "playwright")) {
    throw new Error("Playwright rows must execute through a browser group unit");
  }
  const { invocations, adapt } = invocationsForRows(rows);
  const startedAt = new Date().toISOString();
  const started = Date.now();
  const results = [], invocationFailures = [];
  let cleanupFailed = false;
  for (const invocation of invocations) {
    const execution = await execute(invocation, rows[0].runner);
    try {
      let failureOutput = "";
      const parsing = beginActivity({ repoRoot: root, environment: process.env, launch: execution.launch, activity: "report_parse" });
      try {
        const adapted = await adapt(invocation, execution, (output) => { failureOutput = output; });
        parsing.finish(adapted.some((row) => row.failure_class === "artifact" || row.failure_class === "harness") ? "failed" : "passed");
        results.push(...adapted);
        if (rows[0].runner === "vitest") {
          const failure = vitestInvocationFailure(execution, adapted);
          if (failure) invocationFailures.push(failure);
        }
      } catch (error) { parsing.finish(error.name === "AbortError" ? "cancelled" : "failed"); throw error; }
      if (execution.status !== 0) {
        if (execution.stdout) process.stderr.write(redactString(execution.stdout));
        if (execution.stderr) process.stderr.write(redactString(execution.stderr));
        if (failureOutput) process.stderr.write(redactString(failureOutput));
      }
    } finally {
      try { execution.cleanup(); }
      catch {
        cleanupFailed = true;
        process.stderr.write("row capture cleanup failed (cleanup_error)\n");
      }
    }
  }
  for (const result of results) {
    publishRowResult(runRoot(), {
      schema_id: "cartulary.harness_row_result.v6",
      ...result,
      runner: rows[0].runner,
      started_at: startedAt,
      finished_at: new Date().toISOString(),
      wall_duration_ms: Date.now() - started,
    }, { preserveEquivalent: rows[0].runner === "vitest" });
    if (result.terminal_state !== "passed") {
      const failure = canonicalFailure(result);
      process.stderr.write(
        `[FAIL] row=${result.row_id} runner=${rows[0].runner} failure_class=${failure.failure_class} failure_reason=${failure.failure_reason} row_reason=${result.failure_reason ?? "unknown"}\n`,
      );
    }
  }
  if (invocationFailures.length) {
    const failure = primaryPublicFailure(invocationFailures);
    reportCommandFailure(root, new Error("Vitest invocation did not complete successfully"), failure);
    process.stderr.write(`[FAIL] runner=vitest failure_class=${failure.failure_class} failure_reason=${failure.failure_reason}\n`);
  }
  const failures = [...results.filter((result) => result.terminal_state !== "passed").map(canonicalFailure), ...invocationFailures];
  if (!failures.length) {
    if (cleanupFailed) process.stderr.write("failure_class=harness failure_reason=cleanup_error\n");
    return cleanupFailed ? 12 : 0;
  }
  return publicExitCodeForFailures(failures);
}

try {
  process.exitCode = await main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  if (error.message === usage()) process.exitCode = 2;
  else process.exitCode = publicExitCodeForFailure(reportCommandFailure(root, error, { failure_class: "artifact", failure_reason: "artifact_error" }));
}
