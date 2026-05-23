#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import {
  existsSync,
  readdirSync,
  readFileSync,
  statSync,
} from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  collectServiceTimingContamination,
  formatContaminationReasons,
} from "./lib/duration-drift.mjs";
import {
  prettyJSONString,
  secureWriteFile,
  validateSchemaSync,
} from "./lib/harness-contract.mjs";
import { normalizeOutputMode } from "./lib/tool-output.mjs";

const schemaID = "cartulary.agent_finalize_summary.v2";
const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const makeBin = process.env.MAKE || "make";
const target = "agent-finalize";
const resultsDirInput = (process.env.RESULTS_DIR || "").trim();
const warmBudgetMs = process.env.SCHEDULER_WARM_CHECK_BUDGET_MS || "120000";
const warmBalanceRatio =
  process.env.SCHEDULER_WARM_CHECK_BALANCE_RATIO || "1.25";

const deniedTargets = new Set([
  "format",
  "generate",
  "generate-drift",
  "migration-drift",
  "test-fast",
  "test",
  "check",
  "ci",
  "release-check",
  "browser-e2e",
  "browser-e2e-webserver-backed",
  "browser-e2e-functional",
  "browser-e2e-support",
  "browser-e2e-stateful",
  "browser-e2e-resettable",
  "browser-e2e-measurement",
  "browser-e2e-visual",
  "go-vulncheck",
  "go-gosec-targeted",
  "go-gosec-audit",
  "build",
  "build-server",
  "build-migrate",
  "build-operator",
  "build-web",
  "clean",
  "distclean",
  "benchmark-claim-check",
]);

const preflightSubstep = {
  id: "retained-run-preflight",
  target: null,
  commandKind: "retained_run_preflight",
  requiresResultsDir: true,
  mutatesRepo: false,
  run: "preflight",
};

const actionRegistry = [
  {
    actionID: "structure_ledger_refresh",
    description:
      "Refresh phase-ledger and phase-schedule generated artifacts, then verify no unsupported drift remains.",
    requiresResultsDir: false,
    mutating: true,
    substeps: [
      {
        id: "phase-ledgers",
        target: "phase-ledgers",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: true,
      },
      {
        id: "phase-ledger-drift",
        target: "phase-ledger-drift",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: false,
      },
      {
        id: "phase-schedules",
        target: "phase-schedules",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: true,
      },
      {
        id: "phase-schedule-drift",
        target: "phase-schedule-drift",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: false,
      },
    ],
  },
  {
    actionID: "schema_shape_validation",
    description:
      "Validate harness-owned JSON shape and schema attachments needed by the finalizer path.",
    requiresResultsDir: false,
    mutating: false,
    substeps: [
      {
        id: "json-shape-check",
        target: "json-shape-check",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: false,
      },
    ],
  },
  {
    actionID: "duration_baseline_refresh",
    description:
      "Refresh advisory harness duration-baseline artifacts from a successful, uncontaminated retained run, then refresh schedule artifacts that consume those baselines.",
    requiresResultsDir: true,
    mutating: true,
    substeps: [
      {
        id: "go-test-duration-baselines",
        target: "go-test-duration-baselines",
        commandKind: "make_target",
        requiresResultsDir: true,
        mutatesRepo: true,
      },
      {
        id: "browser-e2e-duration-baselines",
        target: "browser-e2e-duration-baselines",
        commandKind: "make_target",
        requiresResultsDir: true,
        mutatesRepo: true,
      },
      {
        id: "service-backed-make-target-duration-baselines",
        target: "service-backed-make-target-duration-baselines",
        commandKind: "make_target",
        requiresResultsDir: true,
        mutatesRepo: true,
      },
      {
        id: "harness-smoke-duration-baselines",
        target: "harness-smoke-duration-baselines",
        commandKind: "make_target",
        requiresResultsDir: true,
        mutatesRepo: true,
      },
      {
        id: "phase-schedules-after-duration-baselines",
        target: "phase-schedules",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: true,
      },
      {
        id: "phase-schedule-drift-after-duration-baselines",
        target: "phase-schedule-drift",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: false,
      },
    ],
  },
  {
    actionID: "duration_baseline_coverage",
    description:
      "Verify that required advisory duration-baseline entries exist or are explicitly defaulted.",
    requiresResultsDir: false,
    mutating: false,
    substeps: [
      {
        id: "go-test-duration-baseline-coverage",
        target: "go-test-duration-baseline-coverage",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: false,
      },
    ],
  },
  {
    actionID: "duration_baseline_drift_validation",
    description:
      "Validate advisory duration-baseline freshness against the retained run.",
    requiresResultsDir: true,
    mutating: false,
    substeps: [
      {
        id: "duration-baseline-drift-suite",
        target: "duration-baseline-drift-suite",
        commandKind: "make_target",
        requiresResultsDir: true,
        mutatesRepo: false,
      },
    ],
  },
  {
    actionID: "scheduler_drift_validation",
    description:
      "Validate scheduler event ordering and warm-check timing health against the retained run.",
    requiresResultsDir: true,
    mutating: false,
    substeps: [
      {
        id: "scheduler-event-order-drift",
        target: "scheduler-event-order-drift",
        commandKind: "make_target",
        requiresResultsDir: true,
        mutatesRepo: false,
      },
      {
        id: "scheduler-summary-timing-drift",
        target: "scheduler-summary-timing-drift",
        commandKind: "make_target",
        requiresResultsDir: true,
        mutatesRepo: false,
        env: {
          TARGET: "check",
          SCHEDULER_WARM_CHECK_BUDGET_MS: warmBudgetMs,
          SCHEDULER_WARM_CHECK_BALANCE_RATIO: warmBalanceRatio,
        },
      },
    ],
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

function baseSubstep(definition) {
  return {
    id: definition.id,
    target: definition.target,
    command_kind: definition.commandKind,
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

function substepsForAction(definition, includePreflight) {
  const substeps = includePreflight
    ? [preflightSubstep, ...definition.substeps]
    : definition.substeps;
  return substeps.map(baseSubstep);
}

function baseAction(definition, includePreflight) {
  return {
    action_id: definition.actionID,
    description: definition.description,
    requires_results_dir: definition.requiresResultsDir,
    mutating: definition.mutating,
    status: "pending",
    started_at: null,
    completed_at: null,
    duration_ms: null,
    skipped_reason: null,
    substeps: substepsForAction(definition, includePreflight),
  };
}

function selectedActionDefinitions() {
  return actionRegistry.filter(
    (definition) => resultsDirInput || !definition.requiresResultsDir,
  );
}

function selectedActions() {
  const definitions = selectedActionDefinitions();
  return definitions.map((definition, index) =>
    baseAction(definition, Boolean(resultsDirInput) && index === 0),
  );
}

function flattenSubsteps(actions) {
  return actions.flatMap((action) => action.substeps);
}

function collectChildArtifacts(actions) {
  const artifacts = [];
  for (const action of actions) {
    for (const substep of action.substeps) {
      if (substep.summary_json) {
        artifacts.push({
          role: `${action.action_id}_${substep.id}_summary`,
          kind: "json",
          path: substep.summary_json,
        });
      }
      if (substep.stdout_log) {
        artifacts.push({
          role: `${action.action_id}_${substep.id}_stdout`,
          kind: "log",
          path: substep.stdout_log,
        });
      }
      if (substep.stderr_log) {
        artifacts.push({
          role: `${action.action_id}_${substep.id}_stderr`,
          kind: "log",
          path: substep.stderr_log,
        });
      }
    }
  }
  return artifacts.sort((left, right) =>
    `${left.role}\0${left.kind}\0${left.path}`.localeCompare(
      `${right.role}\0${right.kind}\0${right.path}`,
    ),
  );
}

function failureFromChild(action, substep, status, stderr) {
  const summaryFile = substep.target ? childSummaryPath(substep.target) : "";
  const summary = summaryFile ? readJSON(summaryFile) : null;
  if (summary?.failure_class && summary?.failure_reason) {
    return {
      action_id: action.action_id,
      substep_id: substep.id,
      target: substep.target,
      failure_class: summary.failure_class,
      failure_reason: summary.failure_reason,
      headline:
        summary.failures?.[0]?.headline ||
        `${substep.target ?? substep.id} failed with ${summary.failure_reason}`,
      summary_json: relToRepo(summaryFile),
    };
  }
  return {
    action_id: action.action_id,
    substep_id: substep.id,
    target: substep.target,
    failure_class: "harness",
    failure_reason: "child_target_failure",
    headline: `${substep.target ?? substep.id} failed with status ${status}${stderr ? `: ${stderr.split(/\r?\n/u).find(Boolean) ?? ""}` : ""}`,
    summary_json:
      summaryFile && existsSync(summaryFile) ? relToRepo(summaryFile) : null,
  };
}

function preflightFailure(actionID, reason, failureClass = "artifact") {
  return {
    action_id: actionID,
    substep_id: "retained-run-preflight",
    target: null,
    failure_class: failureClass,
    failure_reason:
      failureClass === "config" ? "configuration_error" : "artifact_error",
    headline: reason,
    summary_json: null,
  };
}

function validateRetainedRun(resultsDir, actionID) {
  const resolved = path.resolve(resultsDir);
  if (!existsSync(resolved)) {
    return {
      ok: false,
      failure: preflightFailure(
        actionID,
        `RESULTS_DIR does not exist: ${resultsDir}`,
        "config",
      ),
    };
  }
  if (!statSync(resolved).isDirectory()) {
    return {
      ok: false,
      failure: preflightFailure(
        actionID,
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
        actionID,
        `${relToRepo(checkToolSummary)} must identify a passing warm check run`,
      ),
    };
  }
  for (const file of [checkSchedulerSummary, checkEvents]) {
    if (!existsSync(file)) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
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
        actionID,
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
        actionID,
        `${relToRepo(failedSummary.file)} records a failed retained target`,
      ),
    };
  }

  const contamination = collectServiceTimingContamination(repoRoot, resolved);
  if (contamination.contaminated) {
    return {
      ok: false,
      failure: preflightFailure(
        actionID,
        `RESULTS_DIR contains contaminated timing evidence: ${formatContaminationReasons(contamination)}`,
      ),
    };
  }
  return { ok: true, resolved };
}

function generatedStatusFor(actions, updatedFiles) {
  if (updatedFiles.length > 0) {
    return "updated";
  }
  return actions.some(
    (action) =>
      action.action_id === "structure_ledger_refresh" &&
      action.status === "pass",
  )
    ? "unchanged"
    : "unknown";
}

function actionPassed(actions, actionID) {
  return actions.some(
    (action) => action.action_id === actionID && action.status === "pass",
  );
}

function actionFailed(actions, actionIDs) {
  const ids = new Set(actionIDs);
  return actions.some(
    (action) => ids.has(action.action_id) && action.status === "fail",
  );
}

function writeSummary({
  status,
  startedAt,
  startedMs,
  actions,
  failures,
  updatedFiles,
  resultsDirStatus,
}) {
  const refreshed = actionPassed(actions, "duration_baseline_refresh");
  const durationChecked = actionPassed(
    actions,
    "duration_baseline_drift_validation",
  );
  const runChecked = actionPassed(actions, "scheduler_drift_validation");
  const failedDuration = actionFailed(actions, [
    "duration_baseline_refresh",
    "duration_baseline_drift_validation",
  ]);
  const failedRunCheck = actionFailed(actions, ["scheduler_drift_validation"]);
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
      status: generatedStatusFor(actions, updatedFiles),
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
    actions,
    failures,
    child_artifacts: collectChildArtifacts(actions),
  };
  validateSchemaSync(schemaID, summary);
  secureWriteFile(finalizeSummaryPath(), prettyJSONString(summary));
  return summary;
}

function runMakeSubstep(definition, substep) {
  if (deniedTargets.has(definition.target)) {
    throw new Error(`agent-finalize action substep uses denied target ${definition.target}`);
  }
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
  substep.started_at = startedAt;
  substep.completed_at = completedAt;
  substep.duration_ms = durationMs(startedMs);
  substep.exit_code = result.status ?? 1;
  substep.summary_json = existsSync(summaryFile) ? relToRepo(summaryFile) : null;
  substep.stdout_log = existsSync(stdoutLog) ? relToRepo(stdoutLog) : null;
  substep.stderr_log = existsSync(stderrLog) ? relToRepo(stderrLog) : null;
  substep.status = substep.exit_code === 0 ? "pass" : "fail";
  if (substep.status === "fail") {
    if (result.stdout) {
      process.stderr.write(result.stdout);
    }
    if (result.stderr) {
      process.stderr.write(result.stderr);
    }
  }
  return {
    status: substep.exit_code,
    stderr: result.stderr || result.stdout || "",
  };
}

function markSubstepSkipped(substep, reason = "skipped-after-failure") {
  if (substep.status === "pending") {
    substep.status = "skipped";
    substep.skipped_reason = reason;
  }
}

function markActionSkipped(action, reason = "skipped-after-failure") {
  action.status = "skipped";
  action.skipped_reason = reason;
  for (const substep of action.substeps) {
    markSubstepSkipped(substep, reason);
  }
}

function finalizeActionStatus(action, actionStartedMs) {
  const executedSubsteps = action.substeps.filter((substep) =>
    ["pass", "fail"].includes(substep.status),
  );
  if (executedSubsteps.length === 0 && action.status === "pending") {
    action.status = "skipped";
    action.skipped_reason = "no-selected-substeps";
  } else if (action.substeps.some((substep) => substep.status === "fail")) {
    action.status = "fail";
  } else if (action.status === "pending") {
    action.status = "pass";
  }
  action.completed_at = now();
  action.duration_ms = durationMs(actionStartedMs);
}

function runPreflightSubstep(action, substep) {
  substep.started_at = now();
  const stepStartMs = Date.now();
  const result = validateRetainedRun(resultsDirInput, action.action_id);
  substep.completed_at = now();
  substep.duration_ms = durationMs(stepStartMs);
  substep.exit_code = result.ok ? 0 : 1;
  substep.status = result.ok ? "pass" : "fail";
  return result;
}

function main() {
  const startedAt = now();
  const startedMs = Date.now();
  const beforeStatus = gitStatusMap();
  const actions = selectedActions();
  const definitions = selectedActionDefinitions();
  const failures = [];
  let failed = false;
  let resultsDirStatus = resultsDirInput ? "valid" : "skipped";

  for (let actionIndex = 0; actionIndex < actions.length; actionIndex += 1) {
    const action = actions[actionIndex];
    const definition = definitions[actionIndex];
    if (failed) {
      markActionSkipped(action);
      continue;
    }

    action.started_at = now();
    const actionStartedMs = Date.now();

    for (const substep of action.substeps) {
      if (failed) {
        markSubstepSkipped(substep);
        continue;
      }

      const substepDefinition =
        substep.id === preflightSubstep.id
          ? preflightSubstep
          : definition.substeps.find((entry) => entry.id === substep.id);
      if (!substepDefinition) {
        throw new Error(`missing substep definition for ${action.action_id}:${substep.id}`);
      }

      if (substepDefinition.run === "preflight") {
        const result = runPreflightSubstep(action, substep);
        if (!result.ok) {
          resultsDirStatus = "invalid";
          failures.push(result.failure);
          failed = true;
        }
        continue;
      }

      const result = runMakeSubstep(substepDefinition, substep);
      if (result.status !== 0) {
        failures.push(failureFromChild(action, substep, result.status, result.stderr));
        failed = true;
      }
    }

    finalizeActionStatus(action, actionStartedMs);
  }

  if (failed) {
    for (const action of actions) {
      if (action.status === "pending") {
        markActionSkipped(action);
      }
      for (const substep of action.substeps) {
        markSubstepSkipped(substep);
      }
    }
  }

  const updatedFiles = changedFilesSince(beforeStatus);
  const summary = writeSummary({
    status: failed ? "fail" : "pass",
    startedAt,
    startedMs,
    actions,
    failures,
    updatedFiles,
    resultsDirStatus,
  });

  if (summary.status === "fail") {
    const failure = summary.failures[0];
    process.stderr.write(
      `agent-finalize: failed at ${failure?.target ?? failure?.substep_id ?? failure?.action_id ?? "unknown"}\n`,
    );
    process.exit(
      flattenSubsteps(summary.actions).find((substep) => substep.status === "fail")
        ?.exit_code || 1,
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
