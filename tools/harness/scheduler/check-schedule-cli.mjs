#!/usr/bin/env node
import { existsSync } from "node:fs";
import { mkdir, rm } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  createRunnerContext,
  publicExitCodeForSummary,
} from "../contract/index.mjs";
import {
  attachSchedulerRuntimeCommands,
  createSchedulerRuntimeAttachment,
  stopSchedulerBrowserSessionLeases,
  loadSchedulerRunnerManifest,
} from "./scheduler-runtime.mjs";
import {
  createServiceSessionRuntime,
  serviceSessionTarget,
} from "./scheduler/service-session-runtime.mjs";
import {
  normalizeSchedulerSchedule,
  parseResourceLimitOverride,
} from "./scheduler-manifest.mjs";
import {
  relToRepo as relToRepoPath,
} from "./scheduler-reporting.mjs";
import { formatResourceMap } from "./scheduler-resources.mjs";
import { schedulerAutoLimitResolvers } from "./scheduler-resource-policy.mjs";
import {
  isDryRunFromMakeFlags,
  runLifecycle,
  runNormalizedSchedule,
  writeSchedulerDryRun,
} from "./scheduler-runner.mjs";
import {
  loadSummaryTopologyContext,
  resolveSummaryGroups,
  serviceBackedScheduleChildren,
  summaryGroupsSpec,
} from "../execution/summary-topology.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const defaultManifestPath = path.join(
  repoRoot,
  "tools",
  "scheduler_manifest.json",
);
const supportedSchemaID = "cartulary.scheduler_manifest.v2";
const schedulerEventSchemaID = "cartulary.scheduler_event.v6";
const schedulerSummarySchemaID = "cartulary.check_scheduler_summary.v10";
const goTargetRunnerEnv = "CARTULARY_TEST_GO_TARGET_RUNNER";
const packageReadinessTarget = "check-frontend-install";

function attachRuntime(
  schedule,
  {
    makeBin,
    testOutputScript,
    summaryTargets,
    summaryGroups,
    testServicesBin,
    goTargetRunner,
    goTargetRunnerPrefix,
    tempDir,
    serviceSummaryChildren,
    resultsDir,
    runId,
  },
) {
  const summaryTargetSet = new Set(summaryTargets);
  const targetSummaryFile = (target) =>
    path.join(resultsDir, runId, target, "target-summary.json");
  const serviceTargetStatus = (requestedStatus, children) =>
    requestedStatus === "pass" ||
    children.every((childTarget) => existsSync(targetSummaryFile(childTarget)))
      ? "pass"
      : "fail";
  const serviceSessionRuntime = createServiceSessionRuntime({
    repoRoot,
    workUnits: schedule.workUnits,
    tempDir,
    testServicesBin,
    resultsDir,
    runId,
  });
  const serviceSessionTargets = serviceSessionRuntime.targets;
  const runtimeAttachment = createSchedulerRuntimeAttachment({
    repoRoot,
    workUnits: schedule.workUnits,
    tempDir,
    testOutputScript,
    testServicesBin,
  });
  const {
    browserSessionFiles,
    browserSessionKeys,
    browserSessionUnitByKey,
  } = runtimeAttachment;
  const helperUnitNames = schedule.workUnits
    .filter((unit) => !summaryTargetSet.has(unit.target))
    .map((unit) => unit.target);
  const countedWorkUnitCount = schedule.workUnits.filter(
    (unit) => unit.countInTotal !== false,
  ).length;
  const deferSchemaValidationForPackageReadiness = schedule.workUnits.some(
    (unit) =>
      unit.target === packageReadinessTarget && (unit.needs ?? []).length === 0,
  );
  let runStartEmitted = false;
  const emitRunStart = async () => {
    if (runStartEmitted) {
      return;
    }
    runStartEmitted = true;
    const capacityDisplay =
      schedule.resourceLimits.get("host_cpu") ??
      Math.max(...schedule.resourceLimits.values());
    await runLifecycle(repoRoot, testOutputScript, [
      "run-start",
      schedule.target,
      "--steps",
      String(countedWorkUnitCount),
      "--summary-targets",
      String(summaryTargets.length),
      "--helper-units",
      String(helperUnitNames.length),
      "--jobs",
      String(capacityDisplay),
    ]);
  };
  for (const unit of schedule.workUnits) {
    unit.startDetail = {};
  }
  serviceSessionRuntime.attachCommands();
  attachSchedulerRuntimeCommands(schedule, {
    runtime: runtimeAttachment,
    makeBin,
    goTargetRunner,
    goTargetRunnerPrefix,
    serviceTargetForUnit: serviceSessionTarget,
    serviceEnvFor: serviceSessionRuntime.serviceEnvFor,
    metadataDirForUnit: serviceSessionRuntime.metadataDirForUnit,
    aggregateMetadataDirForUnit:
      serviceSessionRuntime.aggregateMetadataDirForUnit,
    makeTargetSkipPrerequisites: (unit) =>
      unit.makePrerequisitePolicy === "skip",
    skipKinds: ["service_session", "service_complete"],
  });
  return {
    ...schedule,
    kind: "check",
    prefix: "CHECK-SCHEDULER",
    eventSchemaID: schedulerEventSchemaID,
    summarySchemaID: schedulerSummarySchemaID,
    resourceScheduler: "check",
    stopOnFirstFailure: true,
    summaryTotalWallTime: true,
    schemaValidationEnabled: !deferSchemaValidationForPackageReadiness,
    countCompletedUnit: (unit, result) =>
      unit.countInTotal !== false && result.status === 0,
    shouldReplayLog: ({ result, reporter }) =>
      result.status !== 0 || reporter.verbose,
    afterUnitFinish: async (context) => {
      if (
        deferSchemaValidationForPackageReadiness &&
        context.unit.target === packageReadinessTarget &&
        context.result.status === 0
      ) {
        context.reporter.setSchemaValidationEnabled(true);
        await emitRunStart();
      }
      await serviceSessionRuntime.afterUnitFinish(context.unit);
    },
    beforeUnitStart: async ({ unit, started, total, reporter }) => {
      await serviceSessionRuntime.beforeUnitStart(unit);
      if (!reporter.verbose || unit.countInTotal === false) {
        return;
      }
      await runLifecycle(repoRoot, testOutputScript, [
        "step-start",
        schedule.target,
        String(started),
        String(total),
        unit.label,
        "--mode",
        "scheduler",
        "--jobs",
        String(unit.makeJobs),
      ]);
    },
    afterWorkComplete: async () => {
      let cleanupFailure = null;
      await stopSchedulerBrowserSessionLeases(runtimeAttachment);
      cleanupFailure = await serviceSessionRuntime.cleanup();
      return cleanupFailure;
    },
    summaryExtra: ({ reporter }) => ({
      service_sessions: serviceSessionRuntime.summary(
        reporter,
        (value) => relToRepoPath(repoRoot, value),
      ),
      browser_stage_sessions: browserSessionKeys.map((sessionKey) => {
        const unit = browserSessionUnitByKey.get(sessionKey);
        const files = browserSessionFiles.get(sessionKey);
        return {
          target: unit?.target ?? sessionKey,
          session_group: sessionKey,
          aggregate_target: unit?.aggregateTarget ?? unit?.target ?? sessionKey,
          browser_stage: unit?.browserStage ?? "",
          ...(unit?.browserSessionIsolationReason
            ? { isolation_reason: unit.browserSessionIsolationReason }
            : {}),
          env_file: relToRepoPath(repoRoot, files.envFile),
          lease_file: relToRepoPath(repoRoot, files.leaseFile),
        };
      }),
    }),
    beforeRun: async () => {
      if (deferSchemaValidationForPackageReadiness) {
        return;
      }
      await emitRunStart();
    },
    nestedSchedulerLimits: () => [],
    nestedSchedulerObservations: () => [],
    afterSummary: async ({
      reporter,
      requestedStatus,
      completedKeys,
      firstFailureLabel,
    }) => {
      for (const target of serviceSessionTargets) {
        const children = serviceSummaryChildren.get(target) ?? [];
        if (children.length === 0) {
          continue;
        }
        const serviceStatus = serviceTargetStatus(requestedStatus, children);
        await runLifecycle(repoRoot, testOutputScript, [
          "target-summary",
          target,
          serviceStatus,
          "--children",
          children.join(","),
          "--skipped-from-scheduler",
          schedule.target,
          "--suppress-machine-output",
          serviceStatus === "pass" ? "--quiet-success" : "--quiet-failure",
        ]);
      }
      const summaryArgs = [
        "run-summary",
        schedule.target,
        requestedStatus,
        String(reporter.completedCount),
        String(countedWorkUnitCount),
        firstFailureLabel ?? "-",
        "--suppress-machine-output",
        "--quiet-failure",
      ];
      if (summaryGroups) {
        summaryArgs.push("--summary-groups", summaryGroups);
      }
      if (helperUnitNames.length > 0) {
        summaryArgs.push("--helper-units", helperUnitNames.join(","));
        summaryArgs.push(
          "--completed-helper-units",
          helperUnitNames.filter((name) => completedKeys.has(name)).join(","),
        );
      }
      const unitsById = new Map(
        schedule.workUnits.map((unit) => [unit.id, unit]),
      );
      const skippedSummaryTargets = new Set();
      for (const skipped of reporter.skippedWork) {
        const skippedUnit = unitsById.get(skipped.id);
        if (!skippedUnit) {
          continue;
        }
        if (summaryTargetSet.has(skippedUnit.target)) {
          skippedSummaryTargets.add(skippedUnit.target);
        }
        for (const target of skippedUnit.producesSummaryTargets) {
          skippedSummaryTargets.add(target);
        }
      }
      const skippedSummaryTargetsList = summaryTargets.filter((target) =>
        skippedSummaryTargets.has(target),
      );
      if (skippedSummaryTargetsList.length > 0) {
        summaryArgs.push(
          "--skipped-after-failure",
          skippedSummaryTargetsList.join(","),
        );
      }
      summaryArgs.push(...summaryTargets);
      await runLifecycle(
        repoRoot,
        testOutputScript,
        summaryArgs,
        requestedStatus === "pass" ? process.stdout : process.stderr,
      ).catch((error) => {
        if (requestedStatus === "pass") {
          throw error;
        }
      });
      await runLifecycle(repoRoot, testOutputScript, [
        "target-summary",
        schedule.target,
        requestedStatus,
        "--children",
        summaryTargets.join(","),
        "--skipped-from-scheduler",
        schedule.target,
        "--quiet-success",
      ]);
    },
  };
}

async function main() {
  const context = createRunnerContext({ repoRoot });
  const { manifest, manifestPath, options } = await loadSchedulerRunnerManifest(
    process.argv.slice(2),
    {
      defaultManifestPath,
      parseResourceLimitOverride,
      repoRoot,
      schemaID: supportedSchemaID,
      usageText:
        "usage: run-check-schedule.mjs --target <target> [--manifest <path>] [--resource-limit <name=value>...]\n",
    },
  );
  const schedule = normalizeSchedulerSchedule(manifest, options.target, {
    scheduler: "check",
    resourceLimitOverrides: options.resourceLimitOverrides,
    label: "scheduler schedule",
    autoLimitResolvers: (provisionalUnits) =>
      schedulerAutoLimitResolvers("check", provisionalUnits),
  });
  schedule.summaryTargets = schedule.workUnits.flatMap(
    (unit) => unit.producesSummaryTargets,
  );
  const topologyContext = loadSummaryTopologyContext({
    taskSurfaceManifestPath:
      process.env.TASK_SURFACE_MANIFEST ??
      path.join(repoRoot, "tools", "task_surface_manifest.json"),
    schedulerManifestPath: process.env.SCHEDULER_MANIFEST ?? options.manifest,
    browserBatchManifestPath: process.env.BROWSER_E2E_BATCH_MANIFEST,
  });
  const summaryTargets = schedule.summaryTargets;
  const summaryGroups = summaryGroupsSpec(
    resolveSummaryGroups(topologyContext, schedule.summaryGroups),
  );
  if (summaryTargets.length === 0) {
    throw new Error("check schedule must produce at least one summary target");
  }
  const makeBin = process.env.MAKE || "make";
  const testOutputScript =
    process.env.TEST_OUTPUT_SCRIPT ||
    path.join(repoRoot, "tools", "harness", "core", "test-output.mjs");
  const serviceSummaryChildren = new Map();
  for (const unit of schedule.workUnits) {
    const target = serviceSessionTarget(unit);
    if (target && !serviceSummaryChildren.has(target)) {
      serviceSummaryChildren.set(
        target,
        serviceBackedScheduleChildren(topologyContext, target),
      );
    }
  }
  const tempDir = path.join(
    context.resultsDir,
    context.runId,
    options.target,
    "service-sessions",
  );
  await rm(tempDir, { recursive: true, force: true });
  await mkdir(tempDir, { recursive: true });
  const runtimeSchedule = attachRuntime(schedule, {
    makeBin,
    testOutputScript,
    summaryTargets,
    summaryGroups,
    testServicesBin: process.env.TEST_SERVICES_BIN || context.testServicesBin,
    goTargetRunner: process.env[goTargetRunnerEnv] || context.runnerScript,
    goTargetRunnerPrefix: process.env[goTargetRunnerEnv] ? [] : ["go-target"],
    tempDir,
    serviceSummaryChildren,
    resultsDir: context.resultsDir,
    runId: context.runId,
  });

  if (isDryRunFromMakeFlags()) {
    writeSchedulerDryRun({
      repoRoot,
      schedule: runtimeSchedule,
      manifestPath,
      verboseUnitLine(unit) {
        return `[DRY-RUN] ${runtimeSchedule.target} unit ${unit.label} needs=${unit.needs.length === 0 ? "none" : unit.needs.join(",")} claims=${formatResourceMap(unit.resourceClaims)} make_jobs=${unit.makeJobs}\n`;
      },
    });
    return;
  }

  const result = await runNormalizedSchedule({
    repoRoot,
    schedule: runtimeSchedule,
    testOutputScript,
  });
  process.exitCode = publicExitCodeForSummary(result.summary, {
    status: result.status,
  });
}

main().catch((error) => {
  const exitCode = Number.isInteger(error?.exitCode) ? error.exitCode : 2;
  const reason =
    typeof error?.failure_class === "string" &&
    typeof error?.failure_reason === "string"
      ? `failure_class=${error.failure_class} reason=${error.failure_reason}\n`
      : "";
  process.stderr.write(`${error.message}\n${reason}`);
  process.exitCode = exitCode;
});
