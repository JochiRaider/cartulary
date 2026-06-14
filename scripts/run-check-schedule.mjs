#!/usr/bin/env node
import { existsSync } from "node:fs";
import { mkdir, rm } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { publicExitCodeForSummary } from "./lib/failure-taxonomy.mjs";
import { browserGroupCommand } from "./lib/browser-scheduler-dependencies.mjs";
import { createRunnerContext } from "./lib/runner-context.mjs";
import {
  browserSessionFilesFor,
  browserSessionFinalizerCommand,
  browserSessionKeyFor,
  browserSessionStartCommand,
  browserStageCompleteCommand,
  loadSchedulerRunnerManifest,
  readStringEnvFile,
  testOutputRuntimeCommand,
} from "./lib/scheduler/runtime-command-helpers.mjs";
import {
  normalizeSchedulerSchedule,
  parseResourceLimitOverride,
} from "./lib/scheduler-manifest.mjs";
import {
  formatResourceMap,
  relToRepo as relToRepoPath,
} from "./lib/scheduler-reporting.mjs";
import {
  estimateBrowserStackAutoLimit,
  estimateCheckHostCPULimit,
  estimateCheckHostIOLimit,
  estimatePostgresCloneAutoLimit,
  estimatePostgresResetAutoLimit,
} from "./lib/scheduler-resources.mjs";
import {
  isDryRunFromMakeFlags,
  makeChildEnv,
  runLifecycle,
  runNormalizedSchedule,
  writeSchedulerDryRun,
} from "./lib/scheduler-runner.mjs";
import {
  loadSummaryTopologyContext,
  resolveSummaryGroups,
  serviceBackedScheduleChildren,
  summaryGroupsSpec,
} from "./lib/summary-topology.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultManifestPath = path.join(
  repoRoot,
  "tools",
  "scheduler_manifest.json",
);
const supportedSchemaID = "cartulary.scheduler_manifest.v1";
const schedulerEventSchemaID = "cartulary.scheduler_event.v6";
const schedulerSummarySchemaID = "cartulary.check_scheduler_summary.v10";
const goTargetRunnerEnv = "CARTULARY_TEST_GO_TARGET_RUNNER";
const packageReadinessTarget = "check-frontend-install";

function maxResourceClaim(units, resource) {
  return units.reduce(
    (max, unit) => Math.max(max, unit.resourceClaims.get(resource) ?? 0),
    1,
  );
}

async function readServiceSessionEnv(envFile) {
  return readStringEnvFile(
    envFile,
    `service session env file ${envFile} must contain an object`,
  );
}

function serviceSessionTarget(unit) {
  return typeof unit.serviceSession?.target === "string" &&
    unit.serviceSession.target.trim() !== ""
    ? unit.serviceSession.target.trim()
    : "";
}

function attachRuntime(
  schedule,
  {
    makeBin,
    testOutputScript,
    summaryTargets,
    summaryGroups,
    testServicesBin,
    goTargetRunner,
    tempDir,
    serviceSummaryChildren,
    resultsDir,
    runId,
  },
) {
  const summaryTargetSet = new Set(summaryTargets);
  const browserSessionScript =
    process.env.CARTULARY_BROWSER_E2E_SESSION_SCRIPT ||
    path.join(repoRoot, "scripts", "start-web-e2e.sh");
  const browserGroupRunner =
    process.env.CARTULARY_BROWSER_E2E_GROUP_RUNNER || "";
  const testOutputCommand = testOutputRuntimeCommand(testOutputScript);
  const cartularyTestServicesBin =
    process.env.CARTULARY_TEST_SERVICES_BIN ||
    testServicesBin ||
    process.env.TEST_SERVICES_BIN ||
    "";
  const serviceSessionTargets = Array.from(
    new Set(
      schedule.workUnits
        .map(serviceSessionTarget)
        .filter((target) => target !== ""),
    ),
  ).sort((left, right) => left.localeCompare(right));
  const serviceSessionFiles = new Map(
    serviceSessionTargets.map((target) => [
      target,
      {
        envFile: path.join(tempDir, `${target}-env.json`),
        leaseFile: path.join(tempDir, `${target}-lease.json`),
        metadataDir: path.join(tempDir, `${target}-go-shard-metadata`),
      },
    ]),
  );
  const targetSummaryFile = (target) =>
    path.join(resultsDir, runId, target, "target-summary.json");
  const serviceTargetStatus = (requestedStatus, children) =>
    requestedStatus === "pass" ||
    children.every((childTarget) => existsSync(targetSummaryFile(childTarget)))
      ? "pass"
      : "fail";
  const serviceSessionCleanupStatus = new Map(
    serviceSessionTargets.map((target) => [target, "not_started"]),
  );
  const serviceSessionCleanupDurationMs = new Map(
    serviceSessionTargets.map((target) => [target, null]),
  );
  const {
    sessionFiles: browserSessionFiles,
    sessionKeys: browserSessionKeys,
    sessionUnitByKey: browserSessionUnitByKey,
  } = browserSessionFilesFor(schedule.workUnits, tempDir);
  const serviceEnvFor = async (target) => {
    const files = serviceSessionFiles.get(target);
    if (!files) {
      return process.env;
    }
    return {
      ...process.env,
      ...(await readServiceSessionEnv(files.envFile)),
    };
  };
  const recordServiceChildLifecycle = async (unit, event) => {
    if (!unit.serviceSession?.target) {
      return;
    }
    if (unit.kind === "service_session" || unit.kind === "service_complete") {
      return;
    }
    const files = serviceSessionFiles.get(serviceSessionTarget(unit));
    if (!files?.envFile || !existsSync(files.envFile)) {
      return;
    }
    if (!testServicesBin) {
      throw new Error("TEST_SERVICES_BIN is required for service lifecycle accounting");
    }
    await runLifecycle(repoRoot, testServicesBin, [
      "record-lifecycle",
      "--env-file",
      files.envFile,
      "--event",
      event,
      "--child-key",
      unit.id,
    ]);
  };
  const browserEnvFor = async (target) => {
    const files = browserSessionFiles.get(target);
    if (!files) {
      return {};
    }
    return readServiceSessionEnv(files.envFile);
  };
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
    if (unit.kind === "service_session") {
      const files = serviceSessionFiles.get(serviceSessionTarget(unit));
      unit.command = () => {
        if (!testServicesBin) {
          throw new Error(
            "TEST_SERVICES_BIN is required for check service sessions",
          );
        }
        return {
          command: testServicesBin,
          args: [
            "start-suite",
            "--env-file",
            files.envFile,
            "--lease-file",
            files.leaseFile,
          ],
          env: makeChildEnv({
            ...process.env,
            ...unit.env,
            CARTULARY_TEST_RESULTS_DIR: resultsDir,
            CARTULARY_TEST_RUN_ID: runId,
            CARTULARY_TEST_TARGET: unit.target,
            CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
          }),
        };
      };
      continue;
    }
    if (unit.kind === "browser_stage_session") {
      const files = browserSessionFiles.get(browserSessionKeyFor(unit));
      unit.command = async () =>
        browserSessionStartCommand({
          browserSessionScript,
          env: makeChildEnv({
            ...(await serviceEnvFor(serviceSessionTarget(unit))),
            ...unit.env,
            CARTULARY_TEST_SERVICES_BIN: cartularyTestServicesBin,
            CARTULARY_TEST_TARGET: unit.target,
            CARTULARY_BROWSER_SESSION_GROUP: browserSessionKeyFor(unit),
            CARTULARY_BROWSER_STAGE: unit.browserStage,
            CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
          }),
          envFile: files.envFile,
          leaseFile: files.leaseFile,
        });
      continue;
    }
    if (unit.kind === "browser_group") {
      unit.command = async () => {
        const sessionEnv = await browserEnvFor(browserSessionKeyFor(unit));
        const serviceEnv = await serviceEnvFor(serviceSessionTarget(unit));
        const group = unit.browserGroup;
        const pnpmBin =
          process.env.PNPM ||
          path.join(repoRoot, "tmp", "node-runtime", "bin", "pnpm");
        const commonEnv = makeChildEnv({
          ...serviceEnv,
          ...sessionEnv,
          ...unit.env,
          CARTULARY_TEST_SERVICES_BIN: cartularyTestServicesBin,
          CARTULARY_TEST_TARGET: unit.aggregateTarget,
          CARTULARY_BROWSER_SESSION_GROUP: browserSessionKeyFor(unit),
          CARTULARY_BROWSER_STAGE: unit.browserStage,
          CARTULARY_BROWSER_GROUP_KIND: group.kind,
          CARTULARY_BROWSER_GROUP_NAME: group.name,
          CARTULARY_BROWSER_GROUP_TARGET: unit.target,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        });
        return browserGroupCommand({
          browserGroupRunner,
          env: commonEnv,
          group,
          pnpmBin,
          repoRoot,
          scriptEnv: {
            PLAYWRIGHT_WORKERS: "1",
          },
        });
      };
      continue;
    }
    if (unit.kind === "browser_stage_complete") {
      const files = browserSessionFiles.get(browserSessionKeyFor(unit));
      const shouldStopSession = unit.browserSessionFinalizer !== false;
      unit.command = () =>
        browserStageCompleteCommand({
          browserSessionScript,
          env: makeChildEnv({
            ...process.env,
            ...unit.env,
            CARTULARY_TEST_SERVICES_BIN: cartularyTestServicesBin,
            CARTULARY_TEST_TARGET: unit.target,
            CARTULARY_BROWSER_SESSION_GROUP: browserSessionKeyFor(unit),
            CARTULARY_BROWSER_STAGE: unit.browserStage,
            CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
          }),
          leaseFile: files.leaseFile,
          shouldStopSession,
          target: unit.target,
          testOutputCommand,
        });
      continue;
    }
    if (unit.kind === "browser_session_finalizer") {
      const files = browserSessionFiles.get(browserSessionKeyFor(unit));
      unit.command = () =>
        browserSessionFinalizerCommand({
          browserSessionScript,
          env: makeChildEnv({
            ...process.env,
            ...unit.env,
            CARTULARY_TEST_SERVICES_BIN: cartularyTestServicesBin,
            CARTULARY_TEST_TARGET: unit.target,
            CARTULARY_BROWSER_SESSION_GROUP: browserSessionKeyFor(unit),
            CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
          }),
          leaseFile: files.leaseFile,
        });
      continue;
    }
    if (unit.kind === "go_shard") {
      const files = serviceSessionFiles.get(serviceSessionTarget(unit));
      unit.command = async () => ({
        command: goTargetRunner,
        args: ["capture-shard", unit.target, unit.shard, files.metadataDir],
        env: {
          ...(await serviceEnvFor(serviceSessionTarget(unit))),
          ...unit.env,
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        },
      });
      continue;
    }
    if (unit.kind === "aggregate_finalize") {
      const files = serviceSessionFiles.get(
        serviceSessionTarget(unit) || serviceSessionTargets[0],
      );
      unit.command = () => ({
        command: goTargetRunner,
        args: [
          "finalize-shards",
          unit.aggregateTarget,
          files?.metadataDir ?? tempDir,
          ...unit.shardNames,
        ],
        env: {
          ...process.env,
          CARTULARY_TEST_TARGET: unit.aggregateTarget,
          TEST_OUTPUT_SCRIPT: testOutputScript,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        },
      });
      continue;
    }
    if (unit.kind === "service_make_target") {
      unit.command = async () => ({
        command: makeBin,
        args: [
          "--no-print-directory",
          "--output-sync=target",
          "-j1",
          unit.target,
        ],
        env: makeChildEnv({
          ...(await serviceEnvFor(serviceSessionTarget(unit))),
          ...unit.env,
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        }),
      });
      continue;
    }
    if (unit.kind === "service_complete") {
      unit.command = () => ({
        command: process.execPath,
        args: ["-e", ""],
        env: process.env,
      });
      continue;
    }
    unit.command = () => {
      const args = [
        "--no-print-directory",
        "--output-sync=target",
        `-j${unit.makeJobs}`,
        unit.target,
      ];
      const childEnv = {
        ...process.env,
        ...unit.env,
        CARTULARY_TEST_TARGET: unit.target,
        CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
      };
      delete childEnv.CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES;
      if (unit.makePrerequisitePolicy === "skip") {
        childEnv.CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES = "1";
      }
      const env = makeChildEnv(childEnv);
      return { command: makeBin, args, env };
    };
  }
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
      await recordServiceChildLifecycle(context.unit, "child_finished");
    },
    beforeUnitStart: async ({ unit, started, total, reporter }) => {
      await recordServiceChildLifecycle(unit, "child_started");
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
      for (const sessionKey of browserSessionKeys) {
        const files = browserSessionFiles.get(sessionKey);
        if (!files?.leaseFile) {
          continue;
        }
        if (!existsSync(files.leaseFile)) {
          continue;
        }
        await runLifecycle(repoRoot, browserSessionScript, [
          "--session-stop",
          "--lease-file",
          files.leaseFile,
        ]).catch(() => {});
      }
      for (const target of serviceSessionTargets) {
        const files = serviceSessionFiles.get(target);
        if (!files?.leaseFile) {
          continue;
        }
        if (!existsSync(files.leaseFile)) {
          serviceSessionCleanupStatus.set(target, "skipped_no_lease");
          continue;
        }
        serviceSessionCleanupStatus.set(target, "running");
        const cleanupStartedAt = Date.now();
        const result = await runLifecycle(repoRoot, testServicesBin, [
          "terminate-suite",
          "--lease",
          files.leaseFile,
        ]).then(
          () => 0,
          () => 1,
        );
        serviceSessionCleanupDurationMs.set(
          target,
          Math.max(0, Date.now() - cleanupStartedAt),
        );
        if (result !== 0 && !cleanupFailure) {
          serviceSessionCleanupStatus.set(target, "failed");
          cleanupFailure = {
            status: result,
            label: `${target}:terminate-suite`,
          };
        } else if (result === 0) {
          serviceSessionCleanupStatus.set(target, "pass");
        }
      }
      return cleanupFailure;
    },
    summaryExtra: ({ reporter }) => ({
      service_sessions: serviceSessionTargets.map((target) => {
        const files = serviceSessionFiles.get(target);
        const setupRecord = reporter.completedWork.find(
          (record) =>
            record.service_session_target === target &&
            record.work_unit_type === "service_session",
        );
        const childWork = reporter.completedWork.filter(
          (record) =>
            record.service_session_target === target &&
            !["service_session", "service_complete"].includes(
              record.work_unit_type,
            ),
        );
        const childWorkStartedAt =
          childWork.length > 0
            ? Math.min(
                ...childWork.map((record) => record.started_monotonic_ms),
              )
            : null;
        return {
          target,
          env_file: relToRepoPath(repoRoot, files.envFile),
          lease_file: relToRepoPath(repoRoot, files.leaseFile),
          metadata_dir: relToRepoPath(repoRoot, files.metadataDir),
          cleanup_status: serviceSessionCleanupStatus.get(target) ?? "unknown",
          setup_duration_ms: setupRecord?.duration_ms ?? null,
          ready_at_monotonic_ms:
            setupRecord?.status === 0
              ? setupRecord.finished_monotonic_ms
              : null,
          child_work_started_at_monotonic_ms: childWorkStartedAt,
          cleanup_duration_ms:
            serviceSessionCleanupDurationMs.get(target) ?? null,
        };
      }),
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
    autoLimitResolvers: (provisionalUnits) => ({
      check_host_cpu: () => estimateCheckHostCPULimit(),
      check_host_io: ({ resourceLimits: currentLimits }) =>
        Math.max(
          estimateCheckHostIOLimit(currentLimits),
          maxResourceClaim(provisionalUnits, "host_io"),
        ),
      service_backed_browser_stack: ({ resourceLimits: currentLimits }) =>
        estimateBrowserStackAutoLimit(provisionalUnits, currentLimits, {
          cpuResources: ["host_cpu"],
        }),
      service_backed_postgres_clone: ({ resourceLimits: currentLimits }) =>
        estimatePostgresCloneAutoLimit(currentLimits, {
          cpuResources: ["host_cpu"],
          ioResources: ["host_io"],
        }),
      service_backed_postgres_reset: ({ resourceLimits: currentLimits }) =>
        estimatePostgresResetAutoLimit(currentLimits, {
          ioResources: ["host_io"],
        }),
    }),
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
    path.join(repoRoot, "scripts", "lib", "test-output.mjs");
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
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 2;
});
