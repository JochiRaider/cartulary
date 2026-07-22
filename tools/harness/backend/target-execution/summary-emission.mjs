import { spawn } from "node:child_process";
import path from "node:path";
import { Worker } from "node:worker_threads";

import { collectAggregateEmissions } from "../go-target-aggregate.mjs";
import {
  createFailureClassCounts,
  createFailureReasonCounts,
  secureWriteFile,
} from "../../contract/index.mjs";
import { testCoverageBuckets } from "../../contract/test-output-context.mjs";
import {
  prepareStepArtifactDir,
  targetDir,
} from "./context.mjs";
import { renderCommand } from "./command.mjs";
import { rowsForAggregate } from "./planning.mjs";
import { loadStepWindow } from "./reports.mjs";
import {
  captureStart,
  captureFinish,
  nowUTC,
  relToRepo,
  slugifyLabel,
} from "./util.mjs";

export async function runHelper(ctx, args, env = {}) {
  const command = ctx.testOutputScript.endsWith(".mjs")
    ? ctx.nodeBin
    : ctx.testOutputScript;
  const commandArgs = ctx.testOutputScript.endsWith(".mjs")
    ? [ctx.testOutputScript, ...args]
    : args;
  return await new Promise((resolve, reject) => {
    const child = spawn(command, commandArgs, {
      cwd: ctx.repoRoot,
      env: {
        ...ctx.env,
        NODE_BIN: ctx.nodeBin,
        ...env,
      },
      stdio: "inherit",
    });
    child.on("error", reject);
    child.on("close", (status) => resolve(status ?? 1));
  });
}

async function emitTargetTimingSpan(
  ctx,
  bucket,
  label,
  window,
  status,
  exitStatus,
) {
  if (!ctx.testTarget) {
    return;
  }
  await runHelper(ctx, ["timing-span"], {
    CARTULARY_TEST_TARGET: ctx.testTarget,
    CARTULARY_TIMING_BUCKET: bucket,
    CARTULARY_TIMING_LABEL: label,
    CARTULARY_TIMING_START_TIME: window.startTime,
    CARTULARY_TIMING_END_TIME: window.endTime,
    CARTULARY_TIMING_DURATION_MS: String(window.durationMs),
    CARTULARY_TIMING_STATUS: status,
    CARTULARY_STEP_EXIT_STATUS: String(exitStatus),
  });
}

function timingSpanArtifactPath(ctx, label) {
  const dir = path.join(targetDir(ctx), "timing-spans");
  const slug = slugifyLabel(label) || "timing-span";
  const timestamp = nowUTC().replace(/[:.]/g, "-");
  return path.join(dir, `${timestamp}-${process.pid}-${slug}.json`);
}

export function writeTargetTimingSpan(ctx, bucket, label, window, status) {
  if (!ctx.testTarget) {
    return;
  }
  secureWriteFile(
    timingSpanArtifactPath(ctx, label),
    `${JSON.stringify({
      source: "target",
      bucket,
      label,
      start_time: window.startTime,
      end_time: window.endTime,
      duration_ms: window.durationMs,
      status,
    })}\n`,
  );
}

function createNonTestFailureCounts() {
  const counts = {
    tests: 0,
    failed: 1,
    non_test: 1,
    non_test_failed: 1,
    packages: 0,
  };
  for (const coverage of testCoverageBuckets) {
    counts[coverage] = 0;
    counts[`${coverage}_failed`] = 0;
  }
  return counts;
}

function finalizerErrorClassification(error) {
  const message = String(error?.message ?? error ?? "");
  if (message.startsWith("unknown scheduled shard ")) {
    return {
      failureClass: "harness",
      failureReason: "scheduler_accounting_error",
    };
  }
  return {
    failureClass: "artifact",
    failureReason: "artifact_error",
  };
}

export function writeFinalizerFailureStep(
  ctx,
  {
    target,
    label,
    commandArgs,
    window,
    exitStatus = 1,
    error,
    metadataDir,
    aggregateReportDir = "",
    shardNames = [],
  },
) {
  if (!ctx.testTarget) {
    return;
  }
  const { failureClass, failureReason } = finalizerErrorClassification(error);
  const stepDir = prepareStepArtifactDir(ctx, label);
  const counts = createNonTestFailureCounts();
  const failureClasses = createFailureClassCounts();
  failureClasses[failureClass] = 1;
  const failureReasons = createFailureReasonCounts();
  failureReasons[failureReason] = 1;
  const artifacts = {
    metadata_dir: relToRepo(ctx, metadataDir),
  };
  if (aggregateReportDir) {
    artifacts.aggregate_report_dir = relToRepo(ctx, aggregateReportDir);
  }
  const message = String(error?.message ?? error ?? "go shard finalizer failed");
  const failure = {
    failure_class: failureClass,
    failure_reason: failureReason,
    kind: failureClass === "artifact" ? "artifact" : "failure",
    source: "go-shard-finalizer",
    target,
    label,
    message,
    artifact: artifacts.aggregate_report_dir || artifacts.metadata_dir,
    shard_names: shardNames,
  };
  secureWriteFile(
    path.join(stepDir, "step-summary.json"),
    `${JSON.stringify(
      {
        schema_id: "cartulary.test_step_summary.v1",
        label,
        target: ctx.testTarget,
        runner: "go-shard-finalizer",
        status: "fail",
        step: "go-shard-finalize",
        command: renderCommand(commandArgs),
        start_time: window.startTime,
        end_time: window.endTime,
        accounting_mode: "actual",
        executed_duration_ms: window.durationMs,
        logical_duration_ms: window.durationMs,
        reused_duration_ms: 0,
        derived_duration_ms: 0,
        wall_duration_ms: window.durationMs,
        critical_path_wall_duration_ms: window.durationMs,
        teardown_duration_ms: 0,
        timing_bucket: "report_collation",
        exit_status: exitStatus,
        counting_mode: "counted",
        artifacts,
        counts,
        failure_class: failureClass,
        failure_reason: failureReason,
        failure_classes: failureClasses,
        failure_reasons: failureReasons,
        failures: [failure],
        failure_headline: `${failureClass} reason=${failureReason} ${message}`,
        owners: [],
        inventory: [],
        dossiers: [],
        manifest_mismatch: null,
      },
      null,
      2,
    )}\n`,
  );
}

export async function runBounded(items, jobs, worker) {
  if (items.length === 0) {
    return [];
  }
  const workerCount = Math.min(items.length, Math.max(1, jobs));
  const results = new Array(items.length);
  let next = 0;
  await Promise.all(
    Array.from({ length: workerCount }, async () => {
      while (next < items.length) {
        const index = next;
        next += 1;
        results[index] = await worker(items[index], index);
      }
    }),
  );
  return results;
}

export async function runSettledBounded(items, jobs, worker) {
  return await runBounded(items, jobs, async (item, index) => {
    try {
      return Object.freeze({
        value: await worker(item, index),
        error: null,
      });
    } catch (error) {
      return Object.freeze({ value: null, error });
    }
  });
}

async function emitTargetSummary(ctx, status) {
  if (!ctx.testTarget) {
    return 0;
  }
  return await runHelper(
    ctx,
    ["target-summary", ctx.testTarget, status],
    { CARTULARY_TARGET_EVIDENCE_FINALIZE: "1" },
  );
}

async function emitGoTargetInvocationSpan(ctx, status) {
  if (!ctx.invocation || ctx.invocation.emitted) {
    return;
  }
  const window = captureFinish(ctx.invocation);
  ctx.invocation.emitted = true;
  await emitTargetTimingSpan(
    ctx,
    "test_command",
    `run-go-target ${ctx.testTarget || "unknown"}`,
    window,
    status === 0 ? "pass" : "fail",
    status,
  );
}

export async function finishTarget(ctx, status) {
  await emitGoTargetInvocationSpan(ctx, status);
  if (status === 0) {
    return await emitTargetSummary(ctx, "pass");
  }
  await emitTargetSummary(ctx, "fail").catch(() => {});
  return status;
}

function reportStepRequest(
  ctx,
  helperCommand,
  label,
  reportDir,
  mode,
  extraEnv = {},
) {
  const step = loadStepWindow(reportDir, mode);
  const stepDir = prepareStepArtifactDir(ctx, label);
  return Object.freeze({
    helperCommand,
    catalogAware: helperCommand === "go-catalog-step",
    env: Object.freeze({
      CARTULARY_TEST_TARGET: ctx.testTarget,
      CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
      CARTULARY_STEP_LABEL: label,
      CARTULARY_STEP_DIR: stepDir,
      CARTULARY_STEP_COMMAND: step.command,
      CARTULARY_STEP_START_TIME: step.startTime,
      CARTULARY_STEP_END_TIME: step.endTime,
      CARTULARY_STEP_LOGICAL_DURATION_MS: String(step.durationMs),
      CARTULARY_STEP_EXECUTED_DURATION_MS: String(step.durationMs),
      CARTULARY_STEP_WALL_DURATION_MS: String(step.wallDurationMs),
      CARTULARY_STEP_EXIT_STATUS: String(step.exitStatus),
      CARTULARY_REPORT_SLICE: "1",
      CARTULARY_STEP_ACCOUNTING_MODE: mode,
      CARTULARY_STEP_RUNNER_LOG: path.join(reportDir, "runner.jsonl"),
      CARTULARY_STEP_STDERR_LOG: path.join(reportDir, "stderr.log"),
      CARTULARY_CATALOG_OWNER_ID: "",
      CARTULARY_MANIFEST_SECTION: "",
      CARTULARY_MANIFEST_COVERAGE: "",
      CARTULARY_MANIFEST_EXECUTION_DEPENDENCY: "",
      CARTULARY_EXECUTION_FAMILY: "",
      CARTULARY_GO_PACKAGE_PATTERNS: "",
      CARTULARY_MANIFEST_SELECTED_IDS: "",
      CARTULARY_GO_TEST_REGEX: "",
      CARTULARY_ACCOUNTING_COVERAGE: "",
      ...extraEnv,
    }),
  });
}

function packagePatternsEnv(packages) {
  return packages.join("\n");
}

function goRawStepRequest(
  ctx,
  label,
  mode,
  reportDir,
  regex,
  packages,
  coverage,
) {
  return reportStepRequest(ctx, "go-step", label, reportDir, mode, {
    CARTULARY_GO_TEST_REGEX: regex,
    CARTULARY_ACCOUNTING_COVERAGE: coverage,
    CARTULARY_GO_PACKAGE_PATTERNS: packagePatternsEnv(packages),
  });
}

function goCatalogStepRequest(
  ctx,
  label,
  mode,
  reportDir,
  ownerID,
  section,
  coverage,
  executionDependency,
  executionFamily,
  packages,
  selectedIDs = [],
) {
  return reportStepRequest(
    ctx,
    "go-catalog-step",
    label,
    reportDir,
    mode,
    {
      CARTULARY_CATALOG_OWNER_ID: ownerID,
      CARTULARY_MANIFEST_SECTION: section,
      CARTULARY_MANIFEST_COVERAGE: coverage,
      CARTULARY_MANIFEST_EXECUTION_DEPENDENCY: executionDependency,
      CARTULARY_EXECUTION_FAMILY: executionFamily,
      CARTULARY_GO_PACKAGE_PATTERNS: packagePatternsEnv(packages),
      ...(selectedIDs.length > 0
        ? { CARTULARY_MANIFEST_SELECTED_IDS: selectedIDs.join("\n") }
        : {}),
    },
  );
}

export async function emitExecutionFamily(
  ctx,
  target,
  family,
  usage,
  reportDir,
  rows = null,
) {
  const request = createExecutionFamilyRequest(
    ctx,
    target,
    family,
    usage,
    reportDir,
    rows,
  );
  return await emitExecutionFamilyRequest(ctx, request);
}

export function createExecutionFamilyRequest(
  ctx,
  target,
  family,
  usage,
  reportDir,
  rows = null,
) {
  const requests = [];
  const emissions = collectAggregateEmissions(
    rows ?? rowsForAggregate(ctx, target, family),
  );
  for (const [index, emission] of emissions.entries()) {
    const emissionUsage = index === 0 ? usage : "derived";
    if (emission.mode === "manifest") {
      requests.push(goCatalogStepRequest(
        ctx,
        emission.label,
        emissionUsage,
        reportDir,
        emission.owner_id,
        emission.section,
        emission.coverage,
        emission.execution_dependency,
        family,
        emission.packages,
        emission.ids ?? [],
      ));
    } else if (emission.mode === "support") {
      requests.push(goRawStepRequest(
        ctx,
        emission.label,
        emissionUsage,
        reportDir,
        emission.regex,
        emission.packages,
        "support",
      ));
    } else if (emission.mode === "raw") {
      requests.push(goRawStepRequest(
        ctx,
        emission.label,
        emissionUsage,
        reportDir,
        emission.regex,
        emission.packages,
        "raw",
      ));
    } else {
      throw new Error(
        `unsupported execution family emission mode ${emission.mode}`,
      );
    }
  }
  return Object.freeze({
    target,
    family,
    emissions: Object.freeze(requests),
  });
}

export async function emitExecutionFamilyRequest(ctx, request) {
  let status = 0;
  for (const emission of request.emissions) {
    const result = await runHelper(ctx, [emission.helperCommand], emission.env);
    if (result !== 0) status = result;
  }
  return status;
}

function defaultTestOutputScript(ctx) {
  return path.resolve(ctx.testOutputScript) === path.join(
    path.resolve(ctx.repoRoot),
    "tools",
    "harness",
    "output",
    "test-output.mjs",
  );
}

function workerError(value) {
  if (!value) return null;
  const error = new Error(value.message ?? "report emission worker failed");
  if (Number.isInteger(value.exitCode)) error.exitCode = value.exitCode;
  if (value.stack) error.stack = value.stack;
  return error;
}

async function runEmissionWorker(ctx, entries) {
  return await new Promise((resolve) => {
    let settled = false;
    const worker = new Worker(
      new URL("./report-emission-worker.mjs", import.meta.url),
      {
        workerData: { entries },
        env: {
          ...ctx.env,
          NODE_BIN: ctx.nodeBin,
        },
      },
    );
    const finish = (results) => {
      if (settled) return;
      settled = true;
      resolve(results);
    };
    worker.once("message", (message) => finish(message.results));
    worker.once("error", (error) => finish(entries.map(({ index }) => ({
      index,
      error: { message: error.message, stack: error.stack },
      window: null,
      status: null,
    }))));
    worker.once("exit", (status) => {
      if (status !== 0) {
        finish(entries.map(({ index }) => ({
          index,
          error: { message: `report emission worker exited ${status}`, exitCode: status },
          window: null,
          status: null,
        })));
      }
    });
  });
}

export function partitionEmissionRequests(requests, jobs) {
  if (!Number.isInteger(jobs) || jobs < 1) {
    throw new Error(`invalid report emission worker count ${jobs}`);
  }
  if (requests.length === 0) return [];
  const workerCount = Math.min(requests.length, jobs);
  const batches = Array.from({ length: workerCount }, () => []);
  requests.forEach((request, index) => {
    batches[index % workerCount].push(Object.freeze({ index, request }));
  });
  return batches.map((batch) => Object.freeze(batch));
}

export async function emitExecutionFamilyRequests(ctx, requests, jobs) {
  if (!defaultTestOutputScript(ctx)) {
    return await runSettledBounded(requests, jobs, async (request) => {
      const started = captureStart();
      try {
        const status = await emitExecutionFamilyRequest(ctx, request);
        return Object.freeze({ status, window: captureFinish(started) });
      } catch (error) {
        error.window = captureFinish(started);
        throw error;
      }
    });
  }
  if (requests.length === 0) return [];
  const batches = partitionEmissionRequests(requests, jobs);
  const workerResults = (await Promise.all(
    batches.map((entries) => runEmissionWorker(ctx, entries)),
  )).flat();
  const results = new Array(requests.length);
  for (const result of workerResults) {
    const error = workerError(result.error);
    results[result.index] = Object.freeze({
      value: error ? null : Object.freeze({ status: result.status, window: result.window }),
      error,
    });
    if (error && result.window) error.window = result.window;
  }
  return results;
}
