import { spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { mkdir, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  createRunnerContext,
  publicExitCodeForSummary,
  runnerEnv,
  validateSchemaSync,
} from "../contract/index.mjs";
import { phaseSlicePlanOutput } from "../phase-accounting/index.mjs";
import {
  loadRuntimeBinaryRegistry,
  runtimeBinaryAbsoluteEnvForIDs,
} from "../runtime-binary-registry.mjs";
import {
  formatResourceMap,
  relToRepo,
} from "./scheduler-reporting.mjs";
import {
  countVisibleCompletedUnit,
  finalizerRunningDisplayUnits,
  isDryRunFromMakeFlags,
  runLifecycle,
  runNormalizedSchedule,
  writeSchedulerDryRun,
} from "./scheduler-runner.mjs";
import {
  goFinalizerRuntimeCommand,
  goShardRuntimeCommand,
  goTargetRuntimeCommand,
  makeTargetRuntimeCommand,
  schedulerChildEnv,
} from "./scheduler-runtime.mjs";
import {
  normalizeResourceLimits,
  resourceMapToObject,
} from "./scheduler-resources.mjs";
import {
  createServiceSessionRuntime,
  serviceSessionTarget,
} from "./scheduler/service-session-runtime.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const schedulerEventSchemaID = "cartulary.scheduler_event.v6";
const schedulerSummarySchemaID = "cartulary.phase_slice_scheduler_summary.v4";
let runtimeBinaryRegistryCache = null;

function runtimeBinaryRegistryForRepo() {
  if (!runtimeBinaryRegistryCache) {
    runtimeBinaryRegistryCache = loadRuntimeBinaryRegistry({ repoRoot });
  }
  return runtimeBinaryRegistryCache;
}

export function createPhaseSliceRunnerContext(options = {}) {
  return createRunnerContext({ repoRoot, ...options });
}

function runWithContext(command, args, { env = process.env, stdio = "inherit" } = {}) {
  const result = spawnSync(command, args, {
    cwd: repoRoot,
    env,
    stdio,
  });
  if (result.error) {
    throw result.error;
  }
  return result.status ?? 1;
}

export function runPhaseSliceTargetSummary(context, target, status, children = []) {
  const command = context.testOutputScript.endsWith(".mjs")
    ? context.nodeBin
    : context.testOutputScript;
  const args = context.testOutputScript.endsWith(".mjs")
    ? [context.testOutputScript, "target-summary", target, status]
    : ["target-summary", target, status];
  if (children.length > 0) {
    args.push("--children", children.join(","));
  }
  return runWithContext(command, args, { env: runnerEnv(context) });
}

export function phaseSliceTargetPublicExitCode(context, target, fallbackStatus) {
  const summaryFile = path.join(
    context.resultsDir,
    context.runId,
    target,
    "tool-run-summary.json",
  );
  if (!existsSync(summaryFile)) {
    return fallbackStatus === 0 ? 0 : 1;
  }
  try {
    return publicExitCodeForSummary(JSON.parse(readFileSync(summaryFile, "utf8")));
  } catch {
    return fallbackStatus === 0 ? 0 : 1;
  }
}

function runtimeBinaryEnvForIDs(ids = []) {
  return runtimeBinaryAbsoluteEnvForIDs(runtimeBinaryRegistryForRepo(), ids, {
    repoRoot,
    label: "phase-slice",
  });
}

function runtimeBinaryEnvForUnit(unit) {
  return runtimeBinaryEnvForIDs(unit.runtime_binaries ?? []);
}

export function phaseSliceNeedsService(plan) {
  return (
    plan.service_requirements.includes("postgres") ||
    plan.service_requirements.includes("object_store")
  );
}

function resourceLimitsMap(plan) {
  return normalizeResourceLimits(plan.resource_limits, `${plan.target} phase slice`, {
    scheduler: "phase_slice",
  }).limits;
}

function resourceLimitSources(plan) {
  return normalizeResourceLimits(plan.resource_limits, `${plan.target} phase slice`, {
    scheduler: "phase_slice",
  }).sources;
}

function browserStageForUnit(unit) {
  return unit.browserStage ?? "";
}

function runtimeEnv(context, extra = {}) {
  const env = runnerEnv(context, {
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
    ...extra,
  });
  return schedulerChildEnv({
    ...env,
    NODE_RUNTIME_DIR: process.env.NODE_RUNTIME_DIR || path.join(repoRoot, "tmp", "node-runtime"),
    PNPM: process.env.PNPM || path.join(repoRoot, "tmp", "node-runtime", "bin", "pnpm"),
    CARTULARY_SERVER_HARNESS_BIN:
      path.resolve(repoRoot, process.env.SERVER_HARNESS_BIN || "server-harness"),
    CARTULARY_MIGRATE_BIN: path.resolve(repoRoot, process.env.MIGRATE_BIN || "migrate"),
    CARTULARY_TEST_SERVICES_BIN: context.testServicesBin,
  });
}

async function runtimeEnvForUnit(context, serviceRuntime, unit, extra = {}) {
  const serviceEnv = await serviceRuntime.serviceEnvFor(
    unit,
    serviceSessionTarget(unit),
  );
  return runtimeEnv(context, {
    ...serviceEnv,
    ...extra,
  });
}

function frontendRowAccountingEnv(unit) {
  const scope = unit.frontend_row_accounting_scope;
  if (!scope) {
    return {};
  }
  return {
    CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE: scope.mode,
    CARTULARY_FRONTEND_ROW_ACCOUNTING_TARGET: unit.target,
    CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE_NAMESPACE:
      scope.phase_namespace ?? "",
    CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE: scope.phase ?? "",
    CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS:
      (scope.selected_row_ids ?? []).join(","),
  };
}

function exactBaseRowSelectionEnv(plan) {
  if (
    plan.phase_namespace !== "base" ||
    plan.selection?.mode !== "exact_rows"
  ) {
    return {};
  }
  return {
    CARTULARY_GO_TARGET_ROW_IDS:
      (plan.selection.resolved_row_ids ?? []).join(","),
  };
}

function exactBaseManifestSelectionEnv(plan) {
  const selection = exactBaseRowSelectionEnv(plan);
  const rowIDs = selection.CARTULARY_GO_TARGET_ROW_IDS;
  if (!rowIDs) {
    return {};
  }
  return {
    CARTULARY_MANIFEST_SELECTED_IDS: rowIDs.split(",").join("\n"),
  };
}

function exactBaseBrowserSelectionEnv(plan) {
  const selection = exactBaseRowSelectionEnv(plan);
  const rowIDs = selection.CARTULARY_GO_TARGET_ROW_IDS;
  if (!rowIDs) {
    return {};
  }
  return {
    CARTULARY_BROWSER_SELECTED_PHASE: plan.phase,
    CARTULARY_BROWSER_SELECTED_ROW_IDS: rowIDs,
  };
}

function frontendRowAccountingSummaryArgs(unit) {
  const scope = unit.frontend_row_accounting_scope;
  if (!scope) {
    return [];
  }
  const args = [
    "--frontend-row-accounting-scope",
    scope.mode,
    "--frontend-row-accounting-phase-namespace",
    scope.phase_namespace ?? "",
    "--frontend-row-accounting-phase",
    scope.phase ?? "",
  ];
  const selectedRows = (scope.selected_row_ids ?? []).join(",");
  if (selectedRows) {
    args.push("--frontend-row-accounting-row-ids", selectedRows);
  }
  return args;
}

async function emitMakeTargetUnitSummary(context, unit, status) {
  if (unit.kind !== "make_target") {
    return;
  }
  const summaryStatus = status === 0 ? "pass" : "fail";
  const args = [
    "target-summary",
    unit.target,
    summaryStatus,
    "--quiet-success",
    "--quiet-failure",
    "--suppress-machine-output",
    "--preserve-existing-tool-summary",
    ...frontendRowAccountingSummaryArgs(unit),
  ];
  await runLifecycle(
    repoRoot,
    context.testOutputScript,
    args,
    summaryStatus === "pass" ? process.stdout : process.stderr,
    runtimeEnv(context, {
      CARTULARY_TEST_TARGET: unit.target,
      ...frontendRowAccountingEnv(unit),
    }),
  );
}

function attachRuntime(plan, context, metadataDir, serviceRuntime) {
  const resourceLimits = resourceLimitsMap(plan);
  for (const unit of plan.work_units) {
    unit.resourceClaims = new Map(Object.entries(unit.resource_claims ?? Object.fromEntries(unit.resourceClaims ?? [])));
    unit.weightMs = unit.weight_ms;
    unit.readinessAttribution = unit.readiness_attribution ?? null;
    if (["service_session", "service_complete"].includes(unit.kind)) {
      continue;
    }
    if (unit.kind === "finalizer") {
      unit.resourceClaims = new Map();
      unit.command = async () =>
        goFinalizerRuntimeCommand({
          command: context.nodeBin,
          commandPrefix: [context.runnerScript, "go-target"],
          aggregateTarget: unit.aggregateTarget,
          metadataDir,
          shardNames: unit.shardNames,
          env: await runtimeEnvForUnit(context, serviceRuntime, unit, {
            CARTULARY_TEST_TARGET: unit.aggregateTarget,
            CARTULARY_GO_TARGET_PHASE: plan.phase,
            ...exactBaseRowSelectionEnv(plan),
            TEST_OUTPUT_SCRIPT: context.testOutputScript,
          }),
        });
      continue;
    }
    if (unit.kind === "go_shard") {
      unit.command = async () =>
        goShardRuntimeCommand({
          command: context.nodeBin,
          commandPrefix: [context.runnerScript, "go-target"],
          target: unit.target,
          shard: unit.shard,
          metadataDir,
          env: await runtimeEnvForUnit(context, serviceRuntime, unit, {
            CARTULARY_TEST_TARGET: unit.target,
            CARTULARY_GO_TARGET_PHASE: plan.phase,
            ...exactBaseRowSelectionEnv(plan),
            ...runtimeBinaryEnvForUnit(unit),
          }),
        });
      continue;
    }
    if (unit.kind === "go_target") {
      unit.command = async () =>
        goTargetRuntimeCommand({
          command: context.nodeBin,
          commandPrefix: [context.runnerScript],
          target: unit.target,
          env: await runtimeEnvForUnit(context, serviceRuntime, unit, {
            CARTULARY_TEST_TARGET: unit.target,
            CARTULARY_GO_TARGET_PHASE: plan.phase,
            ...exactBaseRowSelectionEnv(plan),
            ...runtimeBinaryEnvForUnit(unit),
          }),
        });
      continue;
    }
    if (unit.kind === "frontend_unit") {
      unit.command = async () => ({
        command: path.join(repoRoot, "tools", "harness", "execution", "run-frontend-unit.sh"),
        args: [],
        env: await runtimeEnvForUnit(context, serviceRuntime, unit, {
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_PHASE_SLICE_PHASE: plan.phase,
          ...exactBaseManifestSelectionEnv(plan),
          ...frontendRowAccountingEnv(unit),
        }),
      });
      continue;
    }
    if (unit.kind === "browser_target") {
      unit.command = async () => ({
        command: path.join(repoRoot, "tools", "harness", "browser", "run-browser-e2e-target.sh"),
        args: [browserStageForUnit(unit)],
        env: await runtimeEnvForUnit(context, serviceRuntime, unit, {
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_PHASE_SLICE_PHASE: plan.phase,
          ...exactBaseBrowserSelectionEnv(plan),
          ...frontendRowAccountingEnv(unit),
        }),
      });
      continue;
    }
    if (unit.kind === "make_target") {
      unit.command = async () =>
        makeTargetRuntimeCommand({
          makeBin: context.makeBin,
          target: unit.target,
          skipPrerequisites: unit.make_prerequisite_policy !== "run",
          env: await runtimeEnvForUnit(context, serviceRuntime, unit, {
            CARTULARY_TEST_TARGET: unit.target,
            ...frontendRowAccountingEnv(unit),
            MAKEFLAGS: "",
          }),
        });
      continue;
    }
    throw new Error(`unsupported phase slice work unit kind ${unit.kind}`);
  }
  serviceRuntime.attachCommands();

  return {
    target: plan.target,
    kind: "phase_slice",
    prefix: "PHASE-SCHEDULER",
    eventSchemaID: schedulerEventSchemaID,
    summarySchemaID: schedulerSummarySchemaID,
    resourceScheduler: "phase_slice",
    stopOnFirstFailure: false,
    showFinalizing: true,
    resourceLimits,
    resourceLimitSources: resourceLimitSources(plan),
    workUnits: plan.work_units,
    totalWorkUnits: plan.total_work_units,
    finalizerCount: plan.finalizer_count,
    childTargets: plan.child_target_names,
    countCompletedUnit: countVisibleCompletedUnit,
    runningDisplayUnits: finalizerRunningDisplayUnits,
    beforeRun: async () => {
      await runLifecycle(repoRoot, context.testOutputScript, [
        "target-start",
        plan.target,
        "--children",
        plan.child_target_names.join(","),
        "--service-backed",
        phaseSliceNeedsService(plan) ? "1" : "0",
      ]).catch(() => {});
    },
    shouldReplayLog: ({ result, reporter }) => result.status !== 0 || reporter.verbose,
    beforeUnitStart: async ({ unit }) => {
      await serviceRuntime.beforeUnitStart(unit);
    },
    afterWorkComplete: async () => serviceRuntime.cleanup(),
    summaryExtra: ({ reporter }) => ({
      phase: plan.phase,
      mode: plan.mode,
      phase_claim_status: plan.phase_claim_status,
      claim_status_counts: plan.claim_status_counts,
      child_targets: plan.child_target_names,
      resource_limits: resourceMapToObject(resourceLimits),
      service_sessions: serviceRuntime.summary(
        reporter,
        (value) => relToRepo(repoRoot, value),
      ),
    }),
    afterUnitFinish: async ({ unit, result }) => {
      await serviceRuntime.afterUnitFinish(unit);
      await emitMakeTargetUnitSummary(context, unit, result.status);
    },
    afterSummary: async ({ requestedStatus }) => {
      await runLifecycle(
        repoRoot,
        context.testOutputScript,
        ["target-summary", plan.target, requestedStatus, "--children", plan.child_target_names.join(",")],
        requestedStatus === "pass" ? process.stdout : process.stderr,
      );
    },
  };
}

export async function runPhaseSliceScheduler(plan, context) {
  const tempDir = path.join(
    context.resultsDir,
    context.runId,
    plan.target,
    "service-sessions",
  );
  const metadataDir = path.join(tempDir, "go-shard-metadata");
  await rm(tempDir, { recursive: true, force: true });
  await mkdir(tempDir, { recursive: true });
  const serviceRuntime = createServiceSessionRuntime({
    repoRoot,
    workUnits: plan.work_units,
    tempDir,
    testServicesBin: context.testServicesBin,
    resultsDir: context.resultsDir,
    runId: context.runId,
  });
  const runtimeSchedule = attachRuntime(
    plan,
    context,
    metadataDir,
    serviceRuntime,
  );
  if (isDryRunFromMakeFlags()) {
    writeSchedulerDryRun({
      repoRoot,
      schedule: runtimeSchedule,
      manifestPath: path.join(
        repoRoot,
        "tools",
        plan.phase_namespace === "frontend"
          ? "frontend_phase_registry.json"
          : "phase_registry.json",
      ),
      verboseUnitLine(unit) {
        if (unit.countInTotal === false) {
          return "";
        }
        const needs = unit.needs?.length > 0 ? ` needs=${unit.needs.join(",")}` : "";
        return `[DRY-RUN] ${plan.target} unit ${unit.label} kind=${unit.kind}${needs} claims=${formatResourceMap(unit.resourceClaims)}\n`;
      },
    });
    return 0;
  }
  const result = await runNormalizedSchedule({
    repoRoot,
    schedule: runtimeSchedule,
    testOutputScript: context.testOutputScript,
  });
  return phaseSliceTargetPublicExitCode(context, plan.target, result.status);
}

export async function writePhaseSlicePlan(plan, context) {
  const targetDir = path.join(context.resultsDir, context.runId, plan.target);
  await mkdir(targetDir, { recursive: true });
  const output = phaseSlicePlanOutput(plan);
  validateSchemaSync(plan.schema_id, output);
  await writeFile(
    path.join(targetDir, "phase-slice-plan.json"),
    `${JSON.stringify(output, null, 2)}\n`,
    "utf8",
  );
}

export async function runPhaseSliceExecution(
  plan,
  context,
  options,
) {
  await writePhaseSlicePlan(plan, context);
  if (plan.no_op) {
    process.stdout.write(
      `[NOOP] ${plan.target} phase=${plan.phase} mode=${options.mode} children=0 phase_claim_status=${plan.phase_claim_status} blocked=${plan.claim_status_counts.blocked}\n`,
    );
    return runPhaseSliceTargetSummary(context, plan.target, "pass", []);
  }

  return await runPhaseSliceScheduler(plan, context);
}
