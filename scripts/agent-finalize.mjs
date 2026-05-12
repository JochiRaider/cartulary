#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  collectServiceTimingContamination,
  formatContaminationReasons,
} from "./lib/duration-drift.mjs";
import { validateSchemaSync } from "./lib/harness-contract.mjs";
import { normalizeOutputMode } from "./lib/tool-output.mjs";

const schemaID = "cartulary.agent_finalize_summary.v1";
const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const makeBin = process.env.MAKE || "make";
const target = "agent-finalize";
const resultsDirInput = (process.env.RESULTS_DIR || "").trim();
const warmBudgetMs = process.env.SCHEDULER_WARM_CHECK_BUDGET_MS || "100000";
const warmBalanceRatio =
  process.env.SCHEDULER_WARM_CHECK_BALANCE_RATIO || "1.25";

const stepDefinitions = [
  {
    id: "retained-run-preflight",
    target: null,
    category: "preflight",
    requiresResultsDir: true,
    mutatesRepo: false,
    run: "preflight",
  },
  {
    id: "phase-ledgers",
    target: "phase-ledgers",
    category: "phase_ledger",
    requiresResultsDir: false,
    mutatesRepo: true,
  },
  {
    id: "phase-ledger-drift",
    target: "phase-ledger-drift",
    category: "phase_ledger",
    requiresResultsDir: false,
    mutatesRepo: false,
  },
  {
    id: "phase-schedules",
    target: "phase-schedules",
    category: "phase_schedule",
    requiresResultsDir: false,
    mutatesRepo: true,
  },
  {
    id: "phase-schedule-drift",
    target: "phase-schedule-drift",
    category: "phase_schedule",
    requiresResultsDir: false,
    mutatesRepo: false,
  },
  {
    id: "json-shape-check",
    target: "json-shape-check",
    category: "shape",
    requiresResultsDir: false,
    mutatesRepo: false,
  },
  {
    id: "go-test-duration-baselines",
    target: "go-test-duration-baselines",
    category: "duration_refresh",
    requiresResultsDir: true,
    mutatesRepo: true,
  },
  {
    id: "browser-e2e-duration-baselines",
    target: "browser-e2e-duration-baselines",
    category: "duration_refresh",
    requiresResultsDir: true,
    mutatesRepo: true,
  },
  {
    id: "service-backed-make-target-duration-baselines",
    target: "service-backed-make-target-duration-baselines",
    category: "duration_refresh",
    requiresResultsDir: true,
    mutatesRepo: true,
  },
  {
    id: "harness-smoke-duration-baselines",
    target: "harness-smoke-duration-baselines",
    category: "duration_refresh",
    requiresResultsDir: true,
    mutatesRepo: true,
  },
  {
    id: "go-test-duration-baseline-coverage",
    target: "go-test-duration-baseline-coverage",
    category: "duration_coverage",
    requiresResultsDir: false,
    mutatesRepo: false,
  },
  {
    id: "duration-baseline-drift-suite",
    target: "duration-baseline-drift-suite",
    category: "duration_drift",
    requiresResultsDir: true,
    mutatesRepo: false,
  },
  {
    id: "scheduler-event-order-drift",
    target: "scheduler-event-order-drift",
    category: "scheduler_check",
    requiresResultsDir: true,
    mutatesRepo: false,
  },
  {
    id: "scheduler-summary-timing-drift",
    target: "scheduler-summary-timing-drift",
    category: "scheduler_check",
    requiresResultsDir: true,
    mutatesRepo: false,
    env: {
      TARGET: "check",
      SCHEDULER_WARM_CHECK_BUDGET_MS: warmBudgetMs,
      SCHEDULER_WARM_CHECK_BALANCE_RATIO: warmBalanceRatio,
    },
  },
];

function now() {
  return new Date().toISOString();
}

function durationMs(startMs) {
  return Math.max(0, Math.round(Date.now() - startMs));
}

function relToRepo(file) {
  const relative = path.relative(repoRoot, file).replaceAll("\\", "/");
  return relative.startsWith("../") || path.isAbsolute(relative)
    ? file.replaceAll("\\", "/")
    : relative;
}

function outputMode() {
  return normalizeOutputMode(process.env);
}

function resultRootAbs() {
  const configured =
    process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results";
  return path.isAbsolute(configured)
    ? configured
    : path.join(repoRoot, configured);
}

function runID() {
  return process.env.CARTULARY_TEST_RUN_ID || "agent-finalize-direct";
}

function runRootAbs() {
  return path.join(resultRootAbs(), runID());
}

function targetDirAbs() {
  if (process.env.CARTULARY_PHASE_ARTIFACT_DIR) {
    return path.dirname(path.resolve(process.env.CARTULARY_PHASE_ARTIFACT_DIR));
  }
  return path.join(runRootAbs(), target);
}

function finalizeSummaryPath() {
  return path.join(targetDirAbs(), "finalize-summary.json");
}

function childSummaryPath(childTarget) {
  return path.join(runRootAbs(), childTarget, "tool-run-summary.json");
}

function childPhaseDir(childTarget) {
  return path.join(runRootAbs(), childTarget, childTarget);
}

function readJSON(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function walkFiles(root) {
  const files = [];
  const stack = [root];
  while (stack.length > 0) {
    const current = stack.pop();
    let entries = [];
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const entry of entries) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
      } else if (entry.isFile()) {
        files.push(next);
      }
    }
  }
  return files.sort();
}

function filesNamed(root, name) {
  if (!existsSync(root)) {
    return [];
  }
  return walkFiles(root).filter((file) => path.basename(file) === name);
}

function gitStatusMap() {
  const result = spawnSync(
    "git",
    ["status", "--porcelain=v1", "--untracked-files=no"],
    {
      cwd: repoRoot,
      encoding: "utf8",
    },
  );
  if (result.status !== 0) {
    return null;
  }
  const entries = new Map();
  for (const line of result.stdout.split(/\r?\n/u).filter(Boolean)) {
    const name = line.slice(3).replace(/^"|"$/gu, "");
    entries.set(name, line.slice(0, 2));
  }
  return entries;
}

function changedFilesSince(before) {
  const after = gitStatusMap();
  if (!before || !after) {
    return [];
  }
  const changed = [];
  for (const [file, status] of after.entries()) {
    if (before.get(file) !== status) {
      changed.push(file);
    }
  }
  return changed.sort((left, right) => left.localeCompare(right));
}

function baseStep(definition) {
  return {
    id: definition.id,
    target: definition.target,
    category: definition.category,
    requires_results_dir: definition.requiresResultsDir,
    mutates_repo: definition.mutatesRepo,
    status: "pending",
    started_at: null,
    completed_at: null,
    duration_ms: null,
    exit_code: null,
    summary_json: null,
    stdout_log: null,
    stderr_log: null,
    skipped_reason: null,
  };
}

function collectChildArtifacts(steps) {
  const artifacts = [];
  for (const step of steps) {
    if (step.summary_json) {
      artifacts.push({
        role: `${step.id}_summary`,
        kind: "json",
        path: step.summary_json,
      });
    }
    if (step.stdout_log) {
      artifacts.push({
        role: `${step.id}_stdout`,
        kind: "log",
        path: step.stdout_log,
      });
    }
    if (step.stderr_log) {
      artifacts.push({
        role: `${step.id}_stderr`,
        kind: "log",
        path: step.stderr_log,
      });
    }
  }
  return artifacts.sort((left, right) =>
    `${left.role}\0${left.kind}\0${left.path}`.localeCompare(
      `${right.role}\0${right.kind}\0${right.path}`,
    ),
  );
}

function failureFromChild(step, status, stderr) {
  const summaryFile = step.target ? childSummaryPath(step.target) : "";
  const summary = summaryFile ? readJSON(summaryFile) : null;
  if (summary?.failure_class && summary?.failure_reason) {
    return {
      step_id: step.id,
      target: step.target,
      failure_class: summary.failure_class,
      failure_reason: summary.failure_reason,
      headline:
        summary.failures?.[0]?.headline ||
        `${step.target ?? step.id} failed with ${summary.failure_reason}`,
      summary_json: relToRepo(summaryFile),
    };
  }
  return {
    step_id: step.id,
    target: step.target,
    failure_class: "harness",
    failure_reason: "child_target_failure",
    headline: `${step.target ?? step.id} failed with status ${status}${stderr ? `: ${stderr.split(/\r?\n/u).find(Boolean) ?? ""}` : ""}`,
    summary_json:
      summaryFile && existsSync(summaryFile) ? relToRepo(summaryFile) : null,
  };
}

function preflightFailure(reason, failureClass = "artifact") {
  return {
    step_id: "retained-run-preflight",
    target: null,
    failure_class: failureClass,
    failure_reason:
      failureClass === "config" ? "configuration_error" : "artifact_error",
    headline: reason,
    summary_json: null,
  };
}

function validateRetainedRun(resultsDir) {
  const resolved = path.resolve(resultsDir);
  if (!existsSync(resolved)) {
    return {
      ok: false,
      failure: preflightFailure(
        `RESULTS_DIR does not exist: ${resultsDir}`,
        "config",
      ),
    };
  }
  if (!statSync(resolved).isDirectory()) {
    return {
      ok: false,
      failure: preflightFailure(
        `RESULTS_DIR is not a directory: ${resultsDir}`,
        "config",
      ),
    };
  }

  const checkToolSummary = path.join(
    resolved,
    "check",
    "tool-run-summary.json",
  );
  const checkSchedulerSummary = path.join(
    resolved,
    "check",
    "scheduler-summary.json",
  );
  const checkEvents = path.join(resolved, "check", "scheduler-events.jsonl");
  const checkSummary = readJSON(checkToolSummary);
  if (!checkSummary || checkSummary.status !== "pass") {
    return {
      ok: false,
      failure: preflightFailure(
        `${relToRepo(checkToolSummary)} must identify a passing warm check run`,
      ),
    };
  }
  for (const file of [checkSchedulerSummary, checkEvents]) {
    if (!existsSync(file)) {
      return {
        ok: false,
        failure: preflightFailure(
          `${relToRepo(file)} is required for warm scheduler checks`,
        ),
      };
    }
  }

  const schedulerSummaries = filesNamed(resolved, "scheduler-summary.json");
  const targetSummaries = filesNamed(resolved, "target-summary.json");
  const phaseSummaries = filesNamed(resolved, "phase-summary.json");
  if (
    schedulerSummaries.length === 0 ||
    targetSummaries.length === 0 ||
    phaseSummaries.length === 0
  ) {
    return {
      ok: false,
      failure: preflightFailure(
        `RESULTS_DIR must contain scheduler, target, and phase summary artifact families`,
      ),
    };
  }

  const failedSummary = filesNamed(resolved, "tool-run-summary.json")
    .map((file) => ({ file, summary: readJSON(file) }))
    .find((entry) => entry.summary?.status === "fail");
  if (failedSummary) {
    return {
      ok: false,
      failure: preflightFailure(
        `${relToRepo(failedSummary.file)} records a failed retained target`,
      ),
    };
  }

  const contamination = collectServiceTimingContamination(repoRoot, resolved);
  if (contamination.contaminated) {
    return {
      ok: false,
      failure: preflightFailure(
        `RESULTS_DIR contains contaminated timing evidence: ${formatContaminationReasons(contamination)}`,
      ),
    };
  }
  return { ok: true, resolved };
}

function writeSummary({
  status,
  startedAt,
  startedMs,
  steps,
  failures,
  updatedFiles,
  resultsDirStatus,
}) {
  const refreshed = steps.some(
    (step) => step.category === "duration_refresh" && step.status === "pass",
  );
  const durationChecked = steps.some(
    (step) => step.category === "duration_drift" && step.status === "pass",
  );
  const runChecked = steps.some(
    (step) => step.category === "scheduler_check" && step.status === "pass",
  );
  const failedDuration = steps.some(
    (step) =>
      ["duration_refresh", "duration_drift"].includes(step.category) &&
      step.status === "fail",
  );
  const failedRunCheck = steps.some(
    (step) => step.category === "scheduler_check" && step.status === "fail",
  );
  const generatedStatus =
    updatedFiles.length > 0
      ? "updated"
      : steps.some(
            (step) =>
              ["phase_ledger", "phase_schedule"].includes(step.category) &&
              step.status === "pass",
          )
        ? "unchanged"
        : "unknown";
  const completedAt = now();
  const summary = {
    schema_id: schemaID,
    target,
    status,
    result_root: relToRepo(resultRootAbs()),
    run_id: runID(),
    run_root: relToRepo(runRootAbs()),
    output_mode: outputMode(),
    results_dir: resultsDirInput
      ? relToRepo(path.resolve(resultsDirInput))
      : null,
    results_dir_status: resultsDirStatus,
    started_at: startedAt,
    completed_at: completedAt,
    duration_ms: durationMs(startedMs),
    generated: {
      status: generatedStatus,
      updated_file_count: updatedFiles.length,
    },
    duration: {
      status: resultsDirInput
        ? failedDuration || !refreshed
          ? "failed"
          : "refreshed"
        : "skipped",
      refreshed,
      checked: durationChecked,
    },
    run_checks: {
      status: resultsDirInput
        ? failedRunCheck
          ? "fail"
          : runChecked
            ? "pass"
            : "fail"
        : "skipped",
      checked: runChecked,
    },
    updated_files: updatedFiles,
    steps,
    failures,
    child_artifacts: collectChildArtifacts(steps),
  };
  mkdirSync(path.dirname(finalizeSummaryPath()), { recursive: true });
  validateSchemaSync(schemaID, summary);
  writeFileSync(finalizeSummaryPath(), `${JSON.stringify(summary, null, 2)}\n`);
  return summary;
}

function runMakeStep(definition, step) {
  const startedAt = now();
  const startedMs = Date.now();
  const childEnv = {
    ...process.env,
    ...(definition.env ?? {}),
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  };
  delete childEnv.CARTULARY_TEST_TARGET;
  const result = spawnSync(
    makeBin,
    ["--no-print-directory", definition.target],
    {
      cwd: repoRoot,
      env: childEnv,
      encoding: "utf8",
    },
  );
  const completedAt = now();
  const summaryFile = childSummaryPath(definition.target);
  const stdoutLog = path.join(childPhaseDir(definition.target), "stdout.log");
  const stderrLog = path.join(childPhaseDir(definition.target), "stderr.log");
  step.started_at = startedAt;
  step.completed_at = completedAt;
  step.duration_ms = durationMs(startedMs);
  step.exit_code = result.status ?? 1;
  step.summary_json = existsSync(summaryFile) ? relToRepo(summaryFile) : null;
  step.stdout_log = existsSync(stdoutLog) ? relToRepo(stdoutLog) : null;
  step.stderr_log = existsSync(stderrLog) ? relToRepo(stderrLog) : null;
  step.status = step.exit_code === 0 ? "pass" : "fail";
  if (step.status === "fail") {
    if (result.stdout) {
      process.stderr.write(result.stdout);
    }
    if (result.stderr) {
      process.stderr.write(result.stderr);
    }
  }
  return {
    status: step.exit_code,
    stderr: result.stderr || result.stdout || "",
  };
}

function main() {
  const startedAt = now();
  const startedMs = Date.now();
  const beforeStatus = gitStatusMap();
  const steps = stepDefinitions
    .filter((definition) => resultsDirInput || !definition.requiresResultsDir)
    .map(baseStep);
  const failures = [];
  let failed = false;
  let resultsDirStatus = resultsDirInput ? "valid" : "skipped";

  for (const definition of stepDefinitions) {
    const step = steps.find((candidate) => candidate.id === definition.id);
    if (!step) {
      continue;
    }
    if (failed) {
      step.status = "skipped";
      step.skipped_reason = "skipped-after-failure";
      continue;
    }

    if (definition.run === "preflight") {
      step.started_at = now();
      const stepStartMs = Date.now();
      const result = validateRetainedRun(resultsDirInput);
      step.completed_at = now();
      step.duration_ms = durationMs(stepStartMs);
      step.exit_code = result.ok ? 0 : 1;
      step.status = result.ok ? "pass" : "fail";
      if (!result.ok) {
        resultsDirStatus = "invalid";
        failures.push(result.failure);
        failed = true;
      }
      continue;
    }

    const result = runMakeStep(definition, step);
    if (result.status !== 0) {
      failures.push(failureFromChild(step, result.status, result.stderr));
      failed = true;
    }
  }

  const updatedFiles = changedFilesSince(beforeStatus);
  if (failed) {
    for (const step of steps.slice(
      steps.findIndex((entry) => entry.status === "fail") + 1,
    )) {
      if (step.status === "pending") {
        step.status = "skipped";
        step.skipped_reason = "skipped-after-failure";
      }
    }
  }

  const summary = writeSummary({
    status: failed ? "fail" : "pass",
    startedAt,
    startedMs,
    steps,
    failures,
    updatedFiles,
    resultsDirStatus,
  });

  if (summary.status === "fail") {
    const failure = summary.failures[0];
    process.stderr.write(
      `agent-finalize: failed at ${failure?.target ?? failure?.step_id ?? "unknown"}\n`,
    );
    process.exit(
      summary.steps.find((step) => step.status === "fail")?.exit_code || 1,
    );
  }
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`agent-finalize: ${message}\n`);
  process.exit(1);
}
