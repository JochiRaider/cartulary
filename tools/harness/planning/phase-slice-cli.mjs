#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { mkdtemp, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  buildPhaseSlicePlan,
  printablePlan,
} from "./phase-slice-plan.mjs";
import {
  buildFrontendPhaseSlicePlan,
  printableFrontendPlan,
} from "../frontend/readiness/index.mjs";
import {
  formatResourceMap,
} from "../scheduler/scheduler-reporting.mjs";
import {
  countVisibleCompletedUnit,
  finalizerRunningDisplayUnits,
  isDryRunFromMakeFlags,
  makeChildEnv,
  runLifecycle,
  runNormalizedSchedule,
  writeSchedulerDryRun,
} from "../scheduler/scheduler-runner.mjs";
import { resourceMapToObject } from "../scheduler/scheduler-resources.mjs";
import {
  createRunnerContext,
  publicExitCodeForSummary,
  runnerEnv,
} from "../core/public-contract.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const validModes = new Set(["phase", "service-backed"]);
const schedulerEventSchemaID = "cartulary.scheduler_event.v6";
const schedulerSummarySchemaID = "cartulary.phase_slice_scheduler_summary.v4";

function usage() {
  process.stderr.write(
    "usage: run-phase-slice.mjs --phase <phaseN|FE-PN> --mode <phase|service-backed> [--phase-namespace <base|frontend>] [--rows <row-id,...>] [--inside-service-wrapper]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = { phase: "", mode: "", phaseNamespace: "base", rows: "", insideServiceWrapper: false, json: false };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--phase") {
      options.phase = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--mode") {
      options.mode = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--phase-namespace") {
      options.phaseNamespace = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--rows") {
      options.rows = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--inside-service-wrapper") {
      options.insideServiceWrapper = true;
      continue;
    }
    if (arg === "--json") {
      options.json = true;
      continue;
    }
    usage();
  }
  if (!options.phase || !validModes.has(options.mode) || !["base", "frontend"].includes(options.phaseNamespace)) {
    usage();
  }
  return options;
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

function runTargetSummary(context, target, status, children = []) {
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

function targetPublicExitCode(context, target, fallbackStatus) {
  const summaryFile = path.join(context.resultsDir, context.runId, target, "tool-run-summary.json");
  if (!existsSync(summaryFile)) {
    return fallbackStatus === 0 ? 0 : 1;
  }
  try {
    return publicExitCodeForSummary(JSON.parse(readFileSync(summaryFile, "utf8")));
  } catch {
    return fallbackStatus === 0 ? 0 : 1;
  }
}

function makeTarget(context, target) {
  return runWithContext(context.makeBin, ["--no-print-directory", target], {
    env: runnerEnv(context, {
      CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
      MAKEFLAGS: "",
    }),
  });
}

function setupTargets(plan) {
  const targets = [];
  const classes = new Set(plan.work_units.map((unit) => unit.class));
  const hasService =
    plan.service_requirements.includes("postgres") || plan.service_requirements.includes("object_store");
  const hasBrowser = classes.has("browser");
  const hasFrontend = classes.has("frontend");
  const hasBackendProcess = plan.work_units.some((unit) => unit.target === "backend-process");
  const runtimeBinaries = new Set(plan.runtime_binaries ?? []);

  if (hasFrontend || hasBrowser) {
    targets.push("frontend-install");
  }
  if (hasBackendProcess || hasBrowser) {
    targets.push("build-server");
  }
  if (runtimeBinaries.has("operator")) {
    targets.push("build-operator");
  }
  if (hasBrowser) {
    targets.push("build-migrate");
  }
  if (hasService) {
    targets.push("test-service-images");
  }
  return Array.from(new Set(targets));
}

function runSetup(context, plan) {
  for (const target of setupTargets(plan)) {
    const status = makeTarget(context, target);
    if (status !== 0) {
      return status;
    }
  }
  return 0;
}

function operatorBinaryPath() {
  return process.env.CARTULARY_OPERATOR_BIN || process.env.OPERATOR_BIN || path.join(repoRoot, "operator");
}

function runtimeBinaryEnvForIDs(ids = []) {
  const env = {};
  if (ids.includes("operator")) {
    env.CARTULARY_OPERATOR_BIN = operatorBinaryPath();
  }
  return env;
}

function runtimeBinaryEnvForPlan(plan) {
  return runtimeBinaryEnvForIDs(plan.runtime_binaries ?? []);
}

function runtimeBinaryEnvForUnit(unit) {
  return runtimeBinaryEnvForIDs(unit.runtime_binaries ?? []);
}

function needsServiceWrapper(plan) {
  return plan.service_requirements.includes("postgres") || plan.service_requirements.includes("object_store");
}

function reexecInsideServiceWrapper(context, options, plan) {
  if (!context.testServicesBin) {
    process.stderr.write("TEST_SERVICES_BIN is required for service-backed phase slices\n");
    return 2;
  }
  const args = [
    "run",
    "--",
    context.nodeBin,
    path.join(repoRoot, "tools", "harness", "planning", "phase-slice-cli.mjs"),
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
  return new Map(Object.entries(plan.resource_limits));
}

function resourceLimitSources(plan) {
  return new Map(Object.keys(plan.resource_limits).map((resource) => [resource, "phase_slice_plan"]));
}

function browserStageForUnit(unit) {
  return unit.browserStage ?? "";
}

function runtimeEnv(context, extra = {}) {
  const env = runnerEnv(context, {
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
    ...extra,
  });
  return makeChildEnv({
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

function attachRuntime(plan, context, metadataDir) {
  const resourceLimits = resourceLimitsMap(plan);
  for (const unit of plan.work_units) {
    unit.resourceClaims = new Map(Object.entries(unit.resource_claims ?? Object.fromEntries(unit.resourceClaims ?? [])));
    unit.weightMs = unit.weight_ms;
    if (unit.kind === "finalizer") {
      unit.resourceClaims = new Map();
      unit.command = () => ({
        command: context.nodeBin,
        args: [
          context.runnerScript,
          "finalize-shards",
          unit.aggregateTarget,
          metadataDir,
          ...unit.shardNames,
        ],
        env: runtimeEnv(context, {
          CARTULARY_TEST_TARGET: unit.aggregateTarget,
          CARTULARY_GO_TARGET_PHASE: plan.phase,
          TEST_OUTPUT_SCRIPT: context.testOutputScript,
        }),
      });
      continue;
    }
    if (unit.kind === "go_shard") {
      unit.command = () => ({
        command: context.nodeBin,
        args: [context.runnerScript, "capture-shard", unit.target, unit.shard, metadataDir],
        env: runtimeEnv(context, {
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_GO_TARGET_PHASE: plan.phase,
          ...runtimeBinaryEnvForUnit(unit),
        }),
      });
      continue;
    }
    if (unit.kind === "go_target") {
      unit.command = () => ({
        command: context.nodeBin,
        args: [context.runnerScript, "go-target", unit.target],
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
        command: path.join(repoRoot, "scripts", "run-frontend-unit.sh"),
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
        command: path.join(repoRoot, "scripts", "run-browser-e2e-target.sh"),
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
      unit.command = () => ({
        command: context.makeBin,
        args: ["--no-print-directory", unit.target],
        env: runtimeEnv(context, {
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES: "1",
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
    resourceScheduler: "service_backed",
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
        needsServiceWrapper(plan) ? "1" : "0",
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

async function runScheduler(plan, context) {
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
    return targetPublicExitCode(context, plan.target, result.status);
  } finally {
    await rm(tempDir, { recursive: true, force: true });
  }
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const context = createRunnerContext({ repoRoot });
  const mode = options.mode === "service-backed" ? "service_backed" : "phase";
  if (options.phaseNamespace === "frontend") {
    const plan = buildFrontendPhaseSlicePlan(options.phase, {
      mode,
      root: repoRoot,
      rowIDs: options.rows,
    });

    if (options.json || process.env.JSON === "1") {
      process.stdout.write(`${JSON.stringify(printableFrontendPlan(plan), null, 2)}\n`);
      return 0;
    }

    if (plan.no_op) {
      process.stdout.write(`[NOOP] ${plan.target} phase=${plan.phase} mode=${options.mode} children=0\n`);
      return runTargetSummary(context, plan.target, "pass", []);
    }

    if (!options.insideServiceWrapper) {
      const setupStatus = runSetup(context, plan);
      if (setupStatus !== 0) {
        const summaryStatus = runTargetSummary(context, plan.target, "fail", plan.child_target_names);
        return targetPublicExitCode(context, plan.target, summaryStatus === 0 ? setupStatus : summaryStatus);
      }
      if (needsServiceWrapper(plan)) {
        return reexecInsideServiceWrapper(context, options, plan);
      }
    }

    return await runScheduler(plan, context);
  }
  if (options.phase.startsWith("FE-P")) {
    throw new Error("frontend phases require --phase-namespace frontend");
  }
  const plan = buildPhaseSlicePlan(options.phase, { mode, root: repoRoot });

  if (options.json || process.env.JSON === "1") {
    process.stdout.write(`${JSON.stringify(printablePlan(plan), null, 2)}\n`);
    return 0;
  }

  if (plan.no_op) {
    process.stdout.write(`[NOOP] ${plan.target} phase=${plan.phase} mode=${options.mode} children=0\n`);
    return runTargetSummary(context, plan.target, "pass", []);
  }

  if (!options.insideServiceWrapper) {
    const setupStatus = runSetup(context, plan);
    if (setupStatus !== 0) {
      const summaryStatus = runTargetSummary(context, plan.target, "fail", plan.child_target_names);
      return targetPublicExitCode(context, plan.target, summaryStatus === 0 ? setupStatus : summaryStatus);
    }
    if (needsServiceWrapper(plan)) {
      return reexecInsideServiceWrapper(context, options, plan);
    }
  }

  return await runScheduler(plan, context);
}

main()
  .then((status) => {
    process.exitCode = status;
  })
  .catch((error) => {
    const message = error instanceof Error ? error.message : String(error);
    const exitCode = Number.isInteger(error?.exitCode) ? error.exitCode : 1;
    process.stderr.write(exitCode === 2 ? `${message}\n` : `phase slice failed: ${message}\n`);
    process.exitCode = exitCode;
  });
