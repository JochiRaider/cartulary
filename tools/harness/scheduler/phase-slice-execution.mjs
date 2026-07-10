import { spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  createRunnerContext,
  publicExitCodeForSummary,
  runnerEnv,
} from "../contract/index.mjs";
import {
  loadRuntimeBinaryRegistry,
  runtimeBinaryDefaultEnvForIDs,
  runtimeBinaryProducerTargetsForIDs,
} from "../runtime-binary-registry.mjs";
import {
  formatResourceMap,
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

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const defaultPhaseSliceCliPath = path.join(
  repoRoot,
  "tools",
  "harness",
  "phase-accounting",
  "phase-slice-cli.mjs",
);
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

export function phaseSliceSetupTargets(plan) {
  const targets = [];
  const classes = new Set(plan.work_units.map((unit) => unit.class));
  const hasService =
    plan.service_requirements.includes("postgres") ||
    plan.service_requirements.includes("object_store");
  const hasBrowser = classes.has("browser");
  const hasFrontend = classes.has("frontend");
  const hasBackendProcess = plan.work_units.some((unit) => unit.target === "backend-process");

  if (hasFrontend || hasBrowser) {
    targets.push("frontend-install");
  }
  if (hasBackendProcess || hasBrowser) {
    targets.push("build-server");
  }
  targets.push(
    ...runtimeBinaryProducerTargetsForIDs(
      runtimeBinaryRegistryForRepo(),
      plan.runtime_binaries ?? [],
      "phase-slice",
    ),
  );
  if (hasBrowser) {
    targets.push("build-migrate");
  }
  if (hasService) {
    targets.push("test-service-images");
  }
  return Array.from(new Set(targets));
}

function makeTarget(context, target) {
  return runWithContext(context.makeBin, ["--no-print-directory", target], {
    env: runnerEnv(context, {
      CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
      MAKEFLAGS: "",
    }),
  });
}

export function runPhaseSliceSetup(context, plan) {
  for (const target of phaseSliceSetupTargets(plan)) {
    const status = makeTarget(context, target);
    if (status !== 0) {
      return status;
    }
  }
  return 0;
}

function runtimeBinaryEnvForIDs(ids = []) {
  return runtimeBinaryDefaultEnvForIDs(runtimeBinaryRegistryForRepo(), ids, "phase-slice");
}

function runtimeBinaryEnvForPlan(plan) {
  return runtimeBinaryEnvForIDs(plan.runtime_binaries ?? []);
}

function runtimeBinaryEnvForUnit(unit) {
  return runtimeBinaryEnvForIDs(unit.runtime_binaries ?? []);
}

export function phaseSliceNeedsServiceWrapper(plan) {
  return (
    plan.service_requirements.includes("postgres") ||
    plan.service_requirements.includes("object_store")
  );
}

export function reexecPhaseSliceInsideServiceWrapper(
  context,
  options,
  plan,
  { phaseSliceCliPath = defaultPhaseSliceCliPath } = {},
) {
  if (!context.testServicesBin) {
    process.stderr.write("TEST_SERVICES_BIN is required for service-backed phase slices\n");
    return 2;
  }
  const args = [
    "run",
    "--",
    context.nodeBin,
    phaseSliceCliPath,
    "--phase",
    options.phase,
    "--mode",
    options.mode,
    "--phase-namespace",
    options.phaseNamespace,
  ];
  if (options.rows) {
    args.push("--rows", options.rows);
  }
  args.push("--inside-service-wrapper");
  return runWithContext(context.testServicesBin, args, {
    env: runnerEnv(context, {
      CARTULARY_TEST_SERVICES_BIN: context.testServicesBin,
      ...runtimeBinaryEnvForPlan(plan),
    }),
  });
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
    CARTULARY_SERVER_BIN: process.env.SERVER_BIN || path.join(repoRoot, "server"),
    CARTULARY_MIGRATE_BIN: process.env.MIGRATE_BIN || path.join(repoRoot, "migrate"),
    CARTULARY_TEST_SERVICES_BIN: context.testServicesBin,
    CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES: "1",
  });
}

function frontendRowAccountingEnv(unit) {
  const scope = unit.frontend_row_accounting_scope;
  if (!scope) {
    return {};
  }
  return {
    CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE: scope.mode,
    CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE_NAMESPACE:
      scope.phase_namespace ?? "",
    CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE: scope.phase ?? "",
    CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS:
      (scope.selected_row_ids ?? []).join(","),
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

function attachRuntime(plan, context, metadataDir) {
  const resourceLimits = resourceLimitsMap(plan);
  for (const unit of plan.work_units) {
    unit.resourceClaims = new Map(Object.entries(unit.resource_claims ?? Object.fromEntries(unit.resourceClaims ?? [])));
    unit.weightMs = unit.weight_ms;
    if (unit.kind === "finalizer") {
      unit.resourceClaims = new Map();
      unit.command = () =>
        goFinalizerRuntimeCommand({
          command: context.nodeBin,
          commandPrefix: [context.runnerScript, "go-target"],
          aggregateTarget: unit.aggregateTarget,
          metadataDir,
          shardNames: unit.shardNames,
          env: runtimeEnv(context, {
            CARTULARY_TEST_TARGET: unit.aggregateTarget,
            CARTULARY_GO_TARGET_PHASE: plan.phase,
            TEST_OUTPUT_SCRIPT: context.testOutputScript,
          }),
        });
      continue;
    }
    if (unit.kind === "go_shard") {
      unit.command = () =>
        goShardRuntimeCommand({
          command: context.nodeBin,
          commandPrefix: [context.runnerScript, "go-target"],
          target: unit.target,
          shard: unit.shard,
          metadataDir,
          env: runtimeEnv(context, {
            CARTULARY_TEST_TARGET: unit.target,
            CARTULARY_GO_TARGET_PHASE: plan.phase,
            ...runtimeBinaryEnvForUnit(unit),
          }),
        });
      continue;
    }
    if (unit.kind === "go_target") {
      unit.command = () =>
        goTargetRuntimeCommand({
          command: context.nodeBin,
          commandPrefix: [context.runnerScript],
          target: unit.target,
          env: runtimeEnv(context, {
            CARTULARY_TEST_TARGET: unit.target,
            CARTULARY_GO_TARGET_PHASE: plan.phase,
            ...runtimeBinaryEnvForUnit(unit),
          }),
        });
      continue;
    }
    if (unit.kind === "frontend_unit") {
      unit.command = () => ({
        command: path.join(repoRoot, "tools", "harness", "execution", "run-frontend-unit.sh"),
        args: [],
        env: runtimeEnv(context, {
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_PHASE_SLICE_PHASE: plan.phase,
          ...frontendRowAccountingEnv(unit),
        }),
      });
      continue;
    }
    if (unit.kind === "browser_target") {
      unit.command = () => ({
        command: path.join(repoRoot, "tools", "harness", "browser", "run-browser-e2e-target.sh"),
        args: [browserStageForUnit(unit)],
        env: runtimeEnv(context, {
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_PHASE_SLICE_PHASE: plan.phase,
          ...frontendRowAccountingEnv(unit),
        }),
      });
      continue;
    }
    if (unit.kind === "make_target") {
      unit.command = () =>
        makeTargetRuntimeCommand({
          makeBin: context.makeBin,
          target: unit.target,
          skipPrerequisites: true,
          env: runtimeEnv(context, {
            CARTULARY_TEST_TARGET: unit.target,
            ...frontendRowAccountingEnv(unit),
            MAKEFLAGS: "",
          }),
        });
      continue;
    }
    throw new Error(`unsupported phase slice work unit kind ${unit.kind}`);
  }

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
        phaseSliceNeedsServiceWrapper(plan) ? "1" : "0",
      ]).catch(() => {});
    },
    shouldReplayLog: ({ result, reporter }) => result.status !== 0 || reporter.verbose,
    summaryExtra: () => ({
      phase: plan.phase,
      mode: plan.mode,
      phase_claim_status: plan.phase_claim_status,
      claim_status_counts: plan.claim_status_counts,
      child_targets: plan.child_target_names,
      resource_limits: resourceMapToObject(resourceLimits),
    }),
    afterUnitFinish: async ({ unit, result }) => {
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
  const tempDir = await mkdtemp(path.join(os.tmpdir(), "cartulary-phase-slice-"));
  const metadataDir = path.join(tempDir, "go-shard-metadata");
  try {
    const runtimeSchedule = attachRuntime(plan, context, metadataDir);
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
  } finally {
    await rm(tempDir, { recursive: true, force: true });
  }
}

async function writeNoOpPhaseSlicePlan(plan, context) {
  const targetDir = path.join(context.resultsDir, context.runId, plan.target);
  await mkdir(targetDir, { recursive: true });
  await writeFile(
    path.join(targetDir, "phase-slice-plan.json"),
    `${JSON.stringify({
      schema_id: plan.schema_id,
      target: plan.target,
      phase: plan.phase,
      mode: plan.mode,
      no_op: plan.no_op,
      phase_claim_status: plan.phase_claim_status,
      claim_status_counts: plan.claim_status_counts,
      row_groups: plan.row_groups,
      child_targets: plan.child_targets,
      child_target_names: plan.child_target_names,
      total_work_units: plan.total_work_units,
      finalizer_count: plan.finalizer_count,
    }, null, 2)}\n`,
    "utf8",
  );
}

export async function runPhaseSliceExecution(
  plan,
  context,
  options,
  { phaseSliceCliPath = defaultPhaseSliceCliPath } = {},
) {
  if (plan.no_op) {
    await writeNoOpPhaseSlicePlan(plan, context);
    process.stdout.write(
      `[NOOP] ${plan.target} phase=${plan.phase} mode=${options.mode} children=0 phase_claim_status=${plan.phase_claim_status} blocked=${plan.claim_status_counts.blocked}\n`,
    );
    return runPhaseSliceTargetSummary(context, plan.target, "pass", []);
  }

  if (!options.insideServiceWrapper) {
    const setupStatus = runPhaseSliceSetup(context, plan);
    if (setupStatus !== 0) {
      const summaryStatus = runPhaseSliceTargetSummary(
        context,
        plan.target,
        "fail",
        plan.child_target_names,
      );
      return phaseSliceTargetPublicExitCode(
        context,
        plan.target,
        summaryStatus === 0 ? setupStatus : summaryStatus,
      );
    }
    if (phaseSliceNeedsServiceWrapper(plan)) {
      return reexecPhaseSliceInsideServiceWrapper(context, options, plan, {
        phaseSliceCliPath,
      });
    }
  }

  return await runPhaseSliceScheduler(plan, context);
}
