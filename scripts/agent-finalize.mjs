#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  unlinkSync,
  writeFileSync,
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
import {
  evaluateActionCache,
  writeActionCacheRecord,
} from "./lib/agent-finalize-action-cache.mjs";
import { normalizeOutputMode } from "./lib/tool-output.mjs";

const schemaID = "cartulary.agent_finalize_summary.v3";
const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const makeBin = process.env.MAKE || "make";
const target = "agent-finalize";
const resultsDirInput = (process.env.RESULTS_DIR || "").trim();
const allowOlderResultsDir =
  (process.env.ALLOW_OLDER_RESULTS_DIR || "").trim() === "1";
const warmBudgetMs = process.env.SCHEDULER_WARM_CHECK_BUDGET_MS || "240000";
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
    cache: {
      eligible: true,
      inputProfileID: "agent_finalize.structure_ledger_refresh.v1",
      actionContractVersion: "v1",
    },
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
    cache: {
      eligible: true,
      inputProfileID: "agent_finalize.schema_shape_validation.v1",
      actionContractVersion: "v1",
    },
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
    cache: {
      eligible: true,
      inputProfileID: "agent_finalize.duration_baseline_refresh.v1",
      actionContractVersion: "v1",
    },
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
    cache: {
      eligible: true,
      inputProfileID: "agent_finalize.duration_baseline_coverage.v1",
      actionContractVersion: "v1",
    },
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
    cache: {
      eligible: true,
      inputProfileID: "agent_finalize.duration_baseline_drift_validation.v1",
      actionContractVersion: "v1",
    },
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
    cache: {
      eligible: true,
      inputProfileID: "agent_finalize.scheduler_drift_validation.v1",
      actionContractVersion: "v1",
    },
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
        makeVars: {
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

function gitTrackedFiles() {
  const result = spawnSync("git", ["ls-files", "-z"], {
    cwd: repoRoot,
    encoding: "buffer",
  });
  if (result.status !== 0) {
    return [];
  }
  return result.stdout
    .toString("utf8")
    .split("\0")
    .filter(Boolean)
    .sort((left, right) => left.localeCompare(right));
}

function snapshotTrackedFiles() {
  const snapshot = new Map();
  for (const file of gitTrackedFiles()) {
    const absolute = path.join(repoRoot, file);
    try {
      snapshot.set(file, readFileSync(absolute));
    } catch {
      snapshot.set(file, null);
    }
  }
  return snapshot;
}

function bufferEqualsSnapshot(file, expected) {
  const absolute = path.join(repoRoot, file);
  if (expected === null) {
    return !existsSync(absolute);
  }
  try {
    return readFileSync(absolute).equals(expected);
  } catch {
    return false;
  }
}

function restoreTrackedSnapshot(snapshot) {
  const changed = [];
  for (const [file, expected] of snapshot.entries()) {
    if (!bufferEqualsSnapshot(file, expected)) {
      changed.push(file);
    }
  }
  if (changed.length === 0) {
    return {
      status: "not_needed",
      restored_files: [],
      error: null,
    };
  }
  const restored = [];
  try {
    for (const file of changed.sort((left, right) => left.localeCompare(right))) {
      const expected = snapshot.get(file);
      const absolute = path.join(repoRoot, file);
      if (expected === null) {
        if (existsSync(absolute)) {
          unlinkSync(absolute);
        }
      } else {
        mkdirSync(path.dirname(absolute), { recursive: true });
        writeFileSync(absolute, expected);
      }
      restored.push(file);
    }
  } catch (error) {
    return {
      status: "failed",
      restored_files: restored,
      error: error instanceof Error ? error.message : String(error),
    };
  }
  return {
    status: "restored",
    restored_files: restored,
    error: null,
  };
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

function baseActionCache(definition, selected) {
  const cache = definition.cache ?? null;
  return {
    enabled: false,
    state: selected
      ? cache?.eligible
        ? "bypass"
        : "ineligible"
      : "bypass",
    cache_schema_id: "cartulary.agent_finalize_action_cache_record.v1",
    action_contract_version: cache?.actionContractVersion ?? null,
    key_sha256: null,
    input_profile_id: cache?.inputProfileID ?? null,
    input_digest_sha256: null,
    output_digest_sha256: null,
    record_path: null,
    reason_code: selected
      ? cache?.eligible
        ? "not_evaluated"
        : "action_ineligible"
      : "action_not_selected",
  };
}

function substepsForAction(definition, includePreflight) {
  const substeps = includePreflight
    ? [preflightSubstep, ...definition.substeps]
    : definition.substeps;
  return substeps.map(baseSubstep);
}

function baseAction(definition, includePreflight) {
  const selected = Boolean(resultsDirInput) || !definition.requiresResultsDir;
  const substeps = substepsForAction(definition, includePreflight);
  if (!selected) {
    for (const substep of substeps) {
      substep.status = "skipped";
      substep.skipped_reason = "results-dir-not-provided";
    }
  }
  return {
    action_id: definition.actionID,
    description: definition.description,
    requires_results_dir: definition.requiresResultsDir,
    mutating: definition.mutating,
    status: selected ? "pending" : "skipped",
    execution_state: selected ? "pending" : "not_selected",
    started_at: null,
    completed_at: null,
    duration_ms: null,
    skipped_reason: selected ? null : "results-dir-not-provided",
    cache: baseActionCache(definition, selected),
    substeps,
  };
}

function selectedActionDefinitions() {
  if (resultsDirInput) {
    const scheduler = actionRegistry.find(
      (action) => action.actionID === "scheduler_drift_validation",
    );
    if (!scheduler) {
      throw new Error("agent-finalize scheduler drift validation action missing");
    }
    return [
      scheduler,
      ...actionRegistry.filter(
        (action) => action.actionID !== "scheduler_drift_validation",
      ),
    ];
  }
  return actionRegistry;
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

function validateRetainedRunArtifacts(resolved, resultsDir, actionID) {
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
  const serviceBackedMarkers = [
    path.join(resolved, "check-service-backed", "tool-run-summary.json"),
    path.join(resolved, "check-service-backed", "target-summary.json"),
    path.join(resolved, "check-service-backed", "scheduler-summary.json"),
  ];
  const checkSchedulerSummary = path.join(
    resolved,
    "check",
    "scheduler-summary.json",
  );
  const checkEvents = path.join(resolved, "check", "scheduler-events.jsonl");
  if (!existsSync(checkToolSummary)) {
    if (serviceBackedMarkers.some((file) => existsSync(file))) {
      return {
        ok: false,
        failure: preflightFailure(
          actionID,
          `${relToRepo(resolved)} contains check-service-backed artifacts but no ${relToRepo(checkToolSummary)}; RESULTS_DIR is a partial service-backed run root and must be a successful full warm make check retained run root`,
          "config",
        ),
      };
    }
    return {
      ok: false,
      failure: preflightFailure(
        actionID,
        `${relToRepo(checkToolSummary)} is required; RESULTS_DIR must be a successful full warm make check retained run root`,
        "config",
      ),
    };
  }
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

function checkCompletedAt(runRoot) {
  const summary = readJSON(path.join(runRoot, "check", "tool-run-summary.json"));
  return summary?.completed_at || summary?.started_at || path.basename(runRoot);
}

function latestSuccessfulCheckRun(parentDir, actionID) {
  if (!existsSync(parentDir) || !statSync(parentDir).isDirectory()) {
    return null;
  }
  const candidates = [];
  for (const entry of readdirSync(parentDir, { withFileTypes: true })) {
    if (!entry.isDirectory()) {
      continue;
    }
    const candidate = path.join(parentDir, entry.name);
    const validation = validateRetainedRunArtifacts(candidate, candidate, actionID);
    if (!validation.ok) {
      continue;
    }
    candidates.push({
      resolved: candidate,
      completed_at: checkCompletedAt(candidate),
    });
  }
  candidates.sort((left, right) => {
    const byTime = String(left.completed_at).localeCompare(String(right.completed_at));
    return byTime || left.resolved.localeCompare(right.resolved);
  });
  return candidates.at(-1) ?? null;
}

function baseRetainedRunSelection() {
  return {
    status: resultsDirInput ? "not_evaluated" : "skipped",
    supplied_results_dir: resultsDirInput
      ? relToRepo(path.resolve(resultsDirInput))
      : null,
    latest_results_dir: null,
    supplied_is_latest: null,
    allow_older_results_dir: allowOlderResultsDir,
  };
}

function validateRetainedRun(resultsDir, actionID) {
  const resolved = path.resolve(resultsDir);
  const validation = validateRetainedRunArtifacts(resolved, resultsDir, actionID);
  if (!validation.ok) {
    return {
      ...validation,
      selection: {
        ...baseRetainedRunSelection(),
        status: "not_evaluated",
      },
    };
  }

  const latest = latestSuccessfulCheckRun(path.dirname(resolved), actionID);
  const latestResolved = latest?.resolved ?? resolved;
  const suppliedIsLatest = path.resolve(latestResolved) === resolved;
  const selection = {
    status: suppliedIsLatest
      ? "latest"
      : allowOlderResultsDir
        ? "older_with_override"
        : "older_rejected",
    supplied_results_dir: relToRepo(resolved),
    latest_results_dir: relToRepo(latestResolved),
    supplied_is_latest: suppliedIsLatest,
    allow_older_results_dir: allowOlderResultsDir,
  };

  if (!suppliedIsLatest && !allowOlderResultsDir) {
    return {
      ok: false,
      selection,
      failure: preflightFailure(
        actionID,
        `RESULTS_DIR is older than the latest successful full warm check retained root ${relToRepo(latestResolved)}; set ALLOW_OLDER_RESULTS_DIR=1 to intentionally use ${relToRepo(resolved)}`,
        "config",
      ),
    };
  }

  return { ok: true, resolved, selection };
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
  mutationRollback,
  resultsDirStatus,
  retainedRunSelection,
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
    retained_run_selection: retainedRunSelection,
    started_at: startedAt,
    completed_at: completedAt,
    duration_ms: durationMs(startedMs),
    generated: {
      status: generatedStatusFor(actions, updatedFiles),
      updated_file_count: updatedFiles.length,
    },
    mutation_rollback: mutationRollback,
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
  if (definition.requiresResultsDir && resultsDirInput) {
    childEnv.CARTULARY_RETAINED_RESULTS_DIR = resultsDirInput;
  } else {
    delete childEnv.CARTULARY_RETAINED_RESULTS_DIR;
  }
  delete childEnv.ALLOW_OLDER_RESULTS_DIR;
  delete childEnv.CARTULARY_MAKE_ORIGIN_ALLOW_OLDER_RESULTS_DIR;
  scrubMakeCommandVariable(childEnv, "ALLOW_OLDER_RESULTS_DIR");
  if (!definition.requiresResultsDir) {
    delete childEnv.RESULTS_DIR;
    delete childEnv.CARTULARY_MAKE_ORIGIN_RESULTS_DIR;
    scrubMakeCommandVariable(childEnv, "RESULTS_DIR");
  }
  delete childEnv.CARTULARY_TEST_TARGET;
  const makeVarArgs = Object.entries(definition.makeVars ?? {}).map(
    ([key, value]) => `${key}=${value}`,
  );
  const result = spawnSync(makeBin, [
    "--no-print-directory",
    definition.target,
    ...makeVarArgs,
  ], {
    cwd: repoRoot,
    env: childEnv,
    encoding: "utf8",
  });
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

function scrubMakeCommandVariable(env, name) {
  for (const key of ["MAKEFLAGS", "MFLAGS", "GNUMAKEFLAGS", "MAKEOVERRIDES"]) {
    if (!env[key]) {
      continue;
    }
    const stripped = stripMakeVariable(env[key], name);
    if (stripped) {
      env[key] = stripped;
    } else {
      delete env[key];
    }
  }
}

function stripMakeVariable(value, name) {
  return value
    .split(/\s+/u)
    .filter((part) => part && part !== name && !part.startsWith(`${name}=`))
    .join(" ");
}

function markSubstepSkipped(substep, reason = "skipped-after-failure") {
  if (substep.status === "pending") {
    substep.status = "skipped";
    substep.skipped_reason = reason;
  }
}

function markActionSkipped(action, reason = "skipped-after-failure") {
  if (action.execution_state === "not_selected") {
    return;
  }
  action.status = "skipped";
  action.execution_state = "skipped_after_failure";
  action.skipped_reason = reason;
  for (const substep of action.substeps) {
    markSubstepSkipped(substep, reason);
  }
}

function finalizeActionStatus(action, actionStartedMs) {
  if (action.execution_state === "not_selected") {
    return;
  }
  const executedSubsteps = action.substeps.filter((substep) =>
    ["pass", "fail"].includes(substep.status),
  );
  if (executedSubsteps.length === 0 && action.status === "pending") {
    action.status = "skipped";
    action.execution_state = "skipped_after_failure";
    action.skipped_reason = "no-selected-substeps";
  } else if (action.substeps.some((substep) => substep.status === "fail")) {
    action.status = "fail";
    if (action.execution_state === "pending") {
      action.execution_state = "executed";
    }
  } else if (action.status === "pending") {
    action.status = "pass";
    action.execution_state = "executed";
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
  const trackedSnapshot = snapshotTrackedFiles();
  const actions = selectedActions();
  const definitions = selectedActionDefinitions();
  const failures = [];
  let failed = false;
  let resultsDirStatus = resultsDirInput ? "valid" : "skipped";
  let retainedRunSelection = baseRetainedRunSelection();
  let mutationRollback = {
    status: "not_needed",
    restored_file_count: 0,
    restored_files: [],
    error: null,
  };

  for (let actionIndex = 0; actionIndex < actions.length; actionIndex += 1) {
    const action = actions[actionIndex];
    const definition = definitions[actionIndex];
    if (action.execution_state === "not_selected") {
      continue;
    }
    if (failed) {
      markActionSkipped(action);
      continue;
    }

    action.started_at = now();
    const actionStartedMs = Date.now();
    let cacheEvaluation = null;
    let substepStartIndex = 0;

    if (action.substeps[0]?.id === preflightSubstep.id) {
      const substep = action.substeps[0];
      const result = runPreflightSubstep(action, substep);
      if (result.selection) {
        retainedRunSelection = result.selection;
      }
      substepStartIndex = 1;
      if (!result.ok) {
        resultsDirStatus = "invalid";
        failures.push(result.failure);
        failed = true;
      }
    }

    if (!failed) {
      cacheEvaluation = evaluateActionCache({
        actionDefinition: definition,
        repoRoot,
        makeBin,
        retainedRunRoot: resultsDirInput ? path.resolve(resultsDirInput) : null,
      });
      action.cache = cacheEvaluation.summary;
      if (cacheEvaluation.reusable) {
        action.status = "pass";
        action.execution_state = "reused";
        for (const substep of action.substeps.slice(substepStartIndex)) {
          markSubstepSkipped(substep, "action-cache-hit");
        }
        finalizeActionStatus(action, actionStartedMs);
        continue;
      }
      action.execution_state = "executed";
    }

    for (const substep of action.substeps.slice(substepStartIndex)) {
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

      const result = runMakeSubstep(substepDefinition, substep);
      if (result.status !== 0) {
        failures.push(failureFromChild(action, substep, result.status, result.stderr));
        failed = true;
      }
    }

    finalizeActionStatus(action, actionStartedMs);
    if (action.status === "pass" && action.execution_state === "executed") {
      writeActionCacheRecord({ actionDefinition: definition, evaluation: cacheEvaluation });
    }
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

  if (failed) {
    const rollback = restoreTrackedSnapshot(trackedSnapshot);
    mutationRollback = {
      status: rollback.status,
      restored_file_count: rollback.restored_files.length,
      restored_files: rollback.restored_files,
      error: rollback.error,
    };
  }

  const updatedFiles = changedFilesSince(beforeStatus);
  const summary = writeSummary({
    status: failed ? "fail" : "pass",
    startedAt,
    startedMs,
    actions,
    failures,
    updatedFiles,
    mutationRollback,
    resultsDirStatus,
    retainedRunSelection,
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
