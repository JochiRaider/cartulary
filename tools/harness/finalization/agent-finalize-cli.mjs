#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  unlinkSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  prettyJSONString,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  evaluateActionCache,
  writeActionCacheRecord,
} from "./agent-finalize-action-cache.mjs";
import {
  collectChildArtifacts,
  flattenSubsteps,
  preflightSubstep,
  selectedActionDefinitions as plannedActionDefinitions,
  selectedActions as plannedActions,
} from "./agent-finalize-action-plan.mjs";
import { createRetainedRunPreflight } from "./agent-finalize-retained-run.mjs";
import { normalizeOutputMode } from "../output/index.mjs";

const schemaID = "cartulary.agent_finalize_summary.v3";
const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const makeBin = process.env.MAKE || "make";
const target = "agent-finalize";
const resultsDirInput = (process.env.RESULTS_DIR || "").trim();
const allowOlderResultsDir =
  (process.env.ALLOW_OLDER_RESULTS_DIR || "").trim() === "1";
const warmBudgetMs = process.env.SCHEDULER_WARM_CHECK_BUDGET_MS || "155000";
const warmBalanceRatio =
  process.env.SCHEDULER_WARM_CHECK_BALANCE_RATIO || "1.25";

const deniedTargets = new Set([
  "format",
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

const actionRegistry = [
  {
    actionID: "generated_structure_refresh",
    description:
      "Refresh catalog-derived task-surface, scheduler, browser, and topology artifacts, then verify no unsupported drift remains.",
    requiresResultsDir: false,
    mutating: true,
    cache: {
      eligible: true,
      inputProfileID: "agent_finalize.generated_structure_refresh.v2",
      actionContractVersion: "v2",
    },
    substeps: [
      {
        id: "generate",
        target: "generate",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: true,
      },
      {
        id: "generate-drift",
        target: "generate-drift",
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
      "Refresh advisory harness duration-baseline artifacts from compatible successful owner evidence, then refresh generated scheduling artifacts that consume those baselines.",
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
        id: "generate-after-duration-baselines",
        target: "generate",
        commandKind: "make_target",
        requiresResultsDir: false,
        mutatesRepo: true,
      },
      {
        id: "generate-drift-after-duration-baselines",
        target: "generate-drift",
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
  if (process.env.CARTULARY_STEP_ARTIFACT_DIR) {
    return path.dirname(path.resolve(process.env.CARTULARY_STEP_ARTIFACT_DIR));
  }
  return path.join(runRootAbs(), target);
}

function finalizeSummaryPath() {
  return path.join(targetDirAbs(), "finalize-summary.json");
}

function childSummaryPath(childTarget) {
  return path.join(runRootAbs(), childTarget, "tool-run-summary.json");
}

function childStepDir(childTarget) {
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

const retainedRunPreflight = createRetainedRunPreflight({
  allowOlderResultsDir,
  filesNamed,
  readJSON,
  relToRepo,
  repoRoot,
  resultsDirInput,
});

function selectedActionDefinitions() {
  return plannedActionDefinitions(actionRegistry, resultsDirInput);
}

function selectedActions() {
  return plannedActions(actionRegistry, resultsDirInput);
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

function generatedStatusFor(actions, updatedFiles) {
  if (updatedFiles.length > 0) {
    return "updated";
  }
  return actions.some(
    (action) =>
      action.action_id === "generated_structure_refresh" &&
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
  removeMakeInputSources(childEnv, ["ALLOW_OLDER_RESULTS_DIR"]);
  scrubMakeCommandVariable(childEnv, "ALLOW_OLDER_RESULTS_DIR");
  if (!definition.requiresResultsDir) {
    delete childEnv.RESULTS_DIR;
    removeMakeInputSources(childEnv, ["RESULTS_DIR"]);
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
  const stdoutLog = path.join(childStepDir(definition.target), "stdout.log");
  const stderrLog = path.join(childStepDir(definition.target), "stderr.log");
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

function removeMakeInputSources(env, removedNames) {
  const removed = new Set(removedNames);
  const tokens = String(env.CARTULARY_MAKE_INPUT_SOURCES ?? "")
    .trim()
    .split(/\s+/u)
    .filter(Boolean)
    .filter((token) => !removed.has(token.split("=", 1)[0]));
  if (tokens.length > 0) {
    env.CARTULARY_MAKE_INPUT_SOURCES = tokens.join(" ");
  } else {
    delete env.CARTULARY_MAKE_INPUT_SOURCES;
  }
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
  const result = retainedRunPreflight.validate(resultsDirInput, action.action_id);
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
  let retainedRunSelection = retainedRunPreflight.baseSelection();
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
