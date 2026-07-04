import { spawn } from "node:child_process";
import path from "node:path";

import { collectAggregateEmissions } from "../go-target-aggregate.mjs";
import {
  createFailureClassCounts,
  createFailureReasonCounts,
  secureWriteFile,
} from "../../contract/index.mjs";
import { testCoverageBuckets } from "../../contract/test-output-context.mjs";
import {
  preparePhaseArtifactDir,
  targetDir,
} from "./context.mjs";
import { renderCommand } from "./command.mjs";
import { rowsForAggregate } from "./planning.mjs";
import { loadPhaseWindow } from "./reports.mjs";
import {
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

export async function emitTargetTimingSpan(
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
    CARTULARY_PHASE_EXIT_STATUS: String(exitStatus),
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

export function writeFinalizerFailurePhase(
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
  const phaseDir = preparePhaseArtifactDir(ctx, label);
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
    path.join(phaseDir, "phase-summary.json"),
    `${JSON.stringify(
      {
        schema_id: "cartulary.test_phase_summary.v3",
        label,
        target: ctx.testTarget,
        runner: "go-shard-finalizer",
        status: "fail",
        phase: "go-shard-finalize",
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

export function resolveFinalizerEmitJobs(ctx, count) {
  if (count <= 0) {
    return 0;
  }
  const configured = ctx.env.CARTULARY_GO_TARGET_FINALIZER_EMIT_JOBS;
  if (configured) {
    const parsed = Number.parseInt(configured, 10);
    if (!Number.isInteger(parsed) || parsed < 1) {
      throw new Error(
        `invalid CARTULARY_GO_TARGET_FINALIZER_EMIT_JOBS=${configured}`,
      );
    }
    return Math.min(count, parsed);
  }
  return Math.min(count, 4);
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

async function emitTargetSummary(ctx, status) {
  if (!ctx.testTarget) {
    return 0;
  }
  return await runHelper(ctx, ["target-summary", ctx.testTarget, status]);
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

async function emitReportPhaseSummary(
  ctx,
  helperCommand,
  label,
  reportDir,
  mode,
  extraEnv = {},
) {
  const phase = loadPhaseWindow(reportDir, mode);
  const phaseDir = preparePhaseArtifactDir(ctx, label);
  return await runHelper(ctx, [helperCommand], {
    CARTULARY_TEST_TARGET: ctx.testTarget,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
    CARTULARY_PHASE_LABEL: label,
    CARTULARY_PHASE_DIR: phaseDir,
    CARTULARY_PHASE_COMMAND: phase.command,
    CARTULARY_PHASE_START_TIME: phase.startTime,
    CARTULARY_PHASE_END_TIME: phase.endTime,
    CARTULARY_PHASE_DURATION_MS: String(phase.durationMs),
    CARTULARY_PHASE_WALL_DURATION_MS: String(phase.wallDurationMs),
    CARTULARY_PHASE_EXIT_STATUS: String(phase.exitStatus),
    CARTULARY_REPORT_SLICE: "1",
    CARTULARY_PHASE_ACCOUNTING_MODE: mode,
    CARTULARY_PHASE_RUNNER_LOG: path.join(reportDir, "runner.jsonl"),
    CARTULARY_PHASE_STDERR_LOG: path.join(reportDir, "stderr.log"),
    ...extraEnv,
  });
}

function packagePatternsEnv(packages) {
  return packages.join("\n");
}

async function emitGoRawPhase(
  ctx,
  label,
  mode,
  reportDir,
  regex,
  packages,
  coverage,
) {
  return await emitReportPhaseSummary(ctx, "go-phase", label, reportDir, mode, {
    CARTULARY_GO_TEST_REGEX: regex,
    CARTULARY_ACCOUNTING_COVERAGE: coverage,
    CARTULARY_GO_PACKAGE_PATTERNS: packagePatternsEnv(packages),
  });
}

async function emitGoManifestPhase(
  ctx,
  label,
  mode,
  reportDir,
  manifestPhase,
  section,
  coverage,
  executionDependency,
  executionFamily,
  packages,
  selectedIDs = [],
) {
  return await emitReportPhaseSummary(
    ctx,
    "go-manifest-phase",
    label,
    reportDir,
    mode,
    {
      CARTULARY_MANIFEST_PHASE: manifestPhase,
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
  let status = 0;
  const emissions = collectAggregateEmissions(
    rows ?? rowsForAggregate(ctx, target, family),
  );
  for (const [index, emission] of emissions.entries()) {
    const emissionUsage = index === 0 ? usage : "derived";
    let result = 0;
    if (emission.mode === "manifest") {
      result = await emitGoManifestPhase(
        ctx,
        emission.label,
        emissionUsage,
        reportDir,
        emission.phase,
        emission.section,
        emission.coverage,
        emission.execution_dependency,
        family,
        emission.packages,
        emission.ids ?? [],
      );
    } else if (emission.mode === "support") {
      result = await emitGoRawPhase(
        ctx,
        emission.label,
        emissionUsage,
        reportDir,
        emission.regex,
        emission.packages,
        "support",
      );
    } else if (emission.mode === "raw") {
      result = await emitGoRawPhase(
        ctx,
        emission.label,
        emissionUsage,
        reportDir,
        emission.regex,
        emission.packages,
        "raw",
      );
    } else {
      throw new Error(
        `unsupported execution family emission mode ${emission.mode}`,
      );
    }
    if (result !== 0) {
      status = result;
    }
  }
  return status;
}
