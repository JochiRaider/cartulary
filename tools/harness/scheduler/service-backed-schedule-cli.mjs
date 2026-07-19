#!/usr/bin/env node
import { spawn } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  createRunnerContext,
  publicExitCodeForSummary,
} from "../contract/index.mjs";
import {
  loadBrowserBatchStages as loadBrowserBatchStagesFromManifest,
} from "./adapters/browser.mjs";
import {
  attachSchedulerRuntimeCommands,
  createSchedulerRuntimeAttachment,
  loadSchedulerRunnerManifest,
  readStringEnvFile,
  stopSchedulerBrowserSessionLeases,
} from "./scheduler-runtime.mjs";
import {
  normalizeSchedulerSchedule,
  parseResourceLimitOverride,
} from "./scheduler-manifest.mjs";
import { formatResourceMap } from "./scheduler-resources.mjs";
import { schedulerAutoLimitResolvers } from "./scheduler-resource-policy.mjs";
import {
  countVisibleCompletedUnit,
  finalizer,
  finalizerRunningDisplayUnits,
  isDryRunFromMakeFlags,
  replayFailedAggregateLogsBeforeFinalizer,
  runLifecycle,
  runNormalizedSchedule,
  writeSchedulerDryRun,
} from "./scheduler-runner.mjs";
import { findTargetDescriptor } from "../backend/backend-target-plan.mjs";
import { loadRuntimeBinaryRegistry } from "../runtime-binary-registry.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const defaultManifestPath = path.join(
  repoRoot,
  "tools",
  "scheduler_manifest.json",
);
const defaultBrowserBatchManifestPath = path.join(
  repoRoot,
  "tools",
  "browser_e2e_batch_manifest.json",
);
const supportedSchemaID = "cartulary.scheduler_manifest.v2";
const schedulerEventSchemaID = "cartulary.scheduler_event.v6";
const schedulerSummarySchemaID =
  "cartulary.service_backed_scheduler_summary.v10";
const goCPUResource = "go_cpu";
const goTargetRunnerEnv = "CARTULARY_TEST_GO_TARGET_RUNNER";
const runtimeProducerTargets = new Set(
  Array.from(loadRuntimeBinaryRegistry({ repoRoot }).values()).map(
    (record) => record.producerTarget,
  ),
);

async function _loadBrowserBatchStages() {
  return loadBrowserBatchStagesFromManifest(defaultBrowserBatchManifestPath);
}

function validateBackendTarget(scheduleTarget, target, label) {
  const descriptor = findTargetDescriptor(target);
  if (!descriptor) {
    throw new Error(`${label} backend target ${target} is not in target-plan`);
  }
  if (!descriptor.serviceBacked) {
    throw new Error(`${label} backend target ${target} is not service-backed`);
  }
  if (
    scheduleTarget === "check-service-backed" &&
    descriptor.checkServiceBackedSafe !== true
  ) {
    throw new Error(
      `${label} backend target ${target} is not check-service-backed safe`,
    );
  }
}

function validateBrowserTarget(source, target, label, browserStages) {
  if (source.type !== "browser_stage") {
    throw new Error(
      `${label} browser target ${target} must use type browser_stage`,
    );
  }
  if (
    typeof source.browser_stage !== "string" ||
    source.browser_stage.trim() === ""
  ) {
    throw new Error(
      `${label} browser target ${target} must declare browser_stage`,
    );
  }
  const browserStage = source.browser_stage.trim();
  const stage = browserStages.get(browserStage);
  if (!stage) {
    throw new Error(
      `${label} browser target ${target} declares unknown browser_stage ${browserStage}`,
    );
  }
  if (stage.target !== target) {
    throw new Error(
      `${label} browser target ${target} must match browser_stage ${browserStage} aggregate target ${stage.target}`,
    );
  }
}

function validateNormalizedServiceBackedSchedule(schedule, browserStages) {
  for (const [index, unit] of schedule.workUnits.entries()) {
    const label = `scheduler schedule ${schedule.target} work_units ${index + 1}`;
    if (unit.class === "browser" || unit.kind.startsWith("browser_")) {
      validateBrowserTarget(
        {
          type: "browser_stage",
          browser_stage: unit.browserStage,
        },
        unit.aggregateTarget || unit.target,
        label,
        browserStages,
      );
      continue;
    }
    if (unit.kind === "make_target" && runtimeProducerTargets.has(unit.target)) {
      continue;
    }
    validateBackendTarget(
      schedule.target,
      unit.aggregateTarget || unit.target,
      label,
    );
  }
}

async function readJSONEnvFile(file) {
  return readStringEnvFile(
    file,
    `${file} must contain a JSON environment object`,
  );
}

function runPostgresFixtureBudgetCheck(targets) {
  return new Promise((resolve, reject) => {
    let stderr = "";
    const child = spawn(
      process.execPath,
      [
        path.join(repoRoot, "tools", "harness", "backend", "postgres-fixture-budget-cli.mjs"),
        "--targets",
        targets.join(","),
      ],
      {
        cwd: repoRoot,
        env: process.env,
        stdio: ["ignore", "pipe", "pipe"],
      },
    );
    child.stdout.pipe(process.stdout, { end: false });
    child.stderr.on("data", (chunk) => {
      stderr += chunk.toString("utf8");
    });
    child.stderr.pipe(process.stderr, { end: false });
    child.on("error", reject);
    child.on("close", (status) => {
      if (status === 0) {
        resolve({ status: 0, stderr });
        return;
      }
      resolve({ status: status ?? 1, stderr });
    });
  });
}

function displayCapacity(schedule) {
  return (
    schedule.resourceLimits.get(goCPUResource) ??
    Math.max(...schedule.resourceLimits.values())
  );
}

function attachRuntime(
  schedule,
  { makeBin, testOutputScript, deferSummary, goTargetRunner, goTargetRunnerPrefix, metadataDir },
) {
  const capacityDisplay = displayCapacity(schedule);
  const unitBrowserSessionKey = (unit) =>
    unit.browserSessionGroup || unit.aggregateTarget || unit.target;
  const runtimeAttachment = createSchedulerRuntimeAttachment({
    repoRoot,
    workUnits: schedule.workUnits,
    tempDir: metadataDir,
    testOutputScript,
    browserEnvReader: readJSONEnvFile,
  });
  attachSchedulerRuntimeCommands(schedule, {
    runtime: runtimeAttachment,
    makeBin,
    goTargetRunner,
    goTargetRunnerPrefix,
    serviceEnvFor: async () => process.env,
    metadataDirForUnit: () => metadataDir,
    aggregateMetadataDirForUnit: () => metadataDir,
    makeTargetJobs: () => 1,
    makeTargetSkipPrerequisites: () => null,
    browserStageSessionEnv: ({ unit, runtime }) => ({
      ...process.env,
      ...(unit.env ?? {}),
      CARTULARY_TEST_SERVICES_BIN: runtime.cartularyTestServicesBin,
      CARTULARY_TEST_TARGET: unit.target,
      CARTULARY_BROWSER_SESSION_GROUP: unitBrowserSessionKey(unit),
      CARTULARY_BROWSER_STAGE: unit.browserStage,
      CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
    }),
    browserGroupEnv: ({ unit, group, sessionEnv, runtime }) => ({
      ...process.env,
      ...sessionEnv,
      ...(unit.env ?? unit.browserWorkerEnv ?? {}),
      CARTULARY_TEST_SERVICES_BIN: runtime.cartularyTestServicesBin,
      CARTULARY_TEST_TARGET: unit.aggregateTarget,
      CARTULARY_BROWSER_SESSION_GROUP: unitBrowserSessionKey(unit),
      CARTULARY_BROWSER_STAGE: unit.browserStage,
      CARTULARY_BROWSER_GROUP_KIND: group.kind,
      CARTULARY_BROWSER_GROUP_NAME: group.name,
      CARTULARY_BROWSER_GROUP_TARGET: unit.target,
      CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
    }),
    browserGroupScriptEnv: ({ unit }) => ({
      CARTULARY_TEST_TARGET: unit.target,
      PLAYWRIGHT_WORKERS: "1",
    }),
    browserStageCompleteEnv: ({ unit, runtime }) => ({
      ...process.env,
      ...(unit.env ?? {}),
      CARTULARY_TEST_SERVICES_BIN: runtime.cartularyTestServicesBin,
      CARTULARY_TEST_TARGET: unit.target,
      CARTULARY_BROWSER_SESSION_GROUP: unitBrowserSessionKey(unit),
      CARTULARY_BROWSER_STAGE: unit.browserStage,
      TEST_OUTPUT_SCRIPT: testOutputScript,
    }),
  });

  return {
    ...schedule,
    kind: "service_backed",
    prefix: "SCHEDULER",
    eventSchemaID: schedulerEventSchemaID,
    summarySchemaID: schedulerSummarySchemaID,
    resourceScheduler: "service_backed",
    showFinalizing: true,
    deferInitialProgress: true,
    validateSummaryTiming: !deferSummary,
    stopOnFirstFailure: false,
    runningDisplayUnits: finalizerRunningDisplayUnits,
    countCompletedUnit: countVisibleCompletedUnit,
    beforeRun: async ({ reporter }) => {
      if (!reporter.verbose) {
        return;
      }
      await runLifecycle(repoRoot, testOutputScript, [
        "target-start",
        schedule.target,
        "--children",
        schedule.children.join(","),
        "--service-backed",
        "1",
      ]);
    },
    beforeUnitStart: async ({ unit, started, total, reporter }) => {
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
        String(capacityDisplay),
      ]);
    },
    beforeReplayLog: replayFailedAggregateLogsBeforeFinalizer,
    shouldReplayLog: ({ result, reporter }) =>
      result.status !== 0 || reporter.verbose,
    afterWorkComplete: async ({ firstFailure }) => {
      if (firstFailure !== 0 || schedule.backendChildren.length === 0) {
        await stopSchedulerBrowserSessionLeases(runtimeAttachment);
        return null;
      }
      await stopSchedulerBrowserSessionLeases(runtimeAttachment);
      const fixtureCheck = await runPostgresFixtureBudgetCheck(
        schedule.backendChildren,
      );
      return fixtureCheck.status === 0
        ? null
        : {
            status: fixtureCheck.status,
            label: "postgres-fixture-shape",
            failure_class: "harness",
            failure_reason: "fixture_error",
            message:
              fixtureCheck.stderr.trim().split(/\r?\n/u).find(Boolean) ??
              "postgres fixture shape check failed",
          };
    },
    summaryExtra: ({ started }) => ({
      started_count: started,
    }),
    afterSummary: async ({ requestedStatus }) => {
      if (deferSummary) {
        return;
      }
      await runLifecycle(
        repoRoot,
        testOutputScript,
        [
          "target-summary",
          schedule.target,
          requestedStatus,
          "--children",
          schedule.children.join(","),
        ],
        requestedStatus === "pass" ? process.stdout : process.stderr,
      );
    },
  };
}

async function runSchedule({
  schedule,
  makeBin,
  testOutputScript,
  deferSummary,
}) {
  const context = createRunnerContext({ repoRoot });
  const tempDir = await mkdtemp(
    path.join(os.tmpdir(), "cartulary-service-backed-schedule-"),
  );
  const metadataDir = path.join(tempDir, "go-shard-metadata");
  const goTargetRunner = process.env[goTargetRunnerEnv] || context.runnerScript;
  const goTargetRunnerPrefix = process.env[goTargetRunnerEnv] ? [] : ["go-target"];
  try {
    const runtimeSchedule = attachRuntime(schedule, {
      makeBin,
      testOutputScript,
      deferSummary,
      goTargetRunner,
      goTargetRunnerPrefix,
      metadataDir,
    });
    const result = await runNormalizedSchedule({
      repoRoot,
      schedule: runtimeSchedule,
      testOutputScript,
    });
    return publicExitCodeForSummary(result.summary, { status: result.status });
  } finally {
    await rm(tempDir, { recursive: true, force: true });
  }
}

async function main() {
  const context = createRunnerContext({ repoRoot });
  const { manifest, manifestPath, options } = await loadSchedulerRunnerManifest(
    process.argv.slice(2),
    {
      allowDeferSummary: true,
      defaultManifestPath,
      parseResourceLimitOverride,
      repoRoot,
      schemaID: supportedSchemaID,
      usageText:
        "usage: run-service-backed-schedule.mjs --target <target> [--manifest <path>] [--defer-summary] [--resource-limit <name=value>...]\n",
    },
  );
  const schedule = normalizeSchedulerSchedule(manifest, options.target, {
    scheduler: "service_backed",
    resourceLimitOverrides: options.resourceLimitOverrides,
    label: "scheduler schedule",
    autoLimitResolvers: (provisionalUnits) =>
      schedulerAutoLimitResolvers("service_backed", provisionalUnits),
  });
  validateNormalizedServiceBackedSchedule(
    schedule,
    await _loadBrowserBatchStages(context),
  );
  schedule.totalWorkUnits = schedule.workUnits.filter(
    (unit) => unit.countInTotal !== false,
  ).length;
  schedule.finalizerCount = schedule.workUnits.filter(finalizer).length;
  schedule.children = Array.from(
    new Set(
      schedule.workUnits
        .map((unit) => unit.aggregateTarget || unit.target)
        .filter(Boolean),
    ),
  ).sort((left, right) => left.localeCompare(right));
  schedule.backendChildren = Array.from(
    new Set(
      schedule.workUnits
        .filter((unit) => unit.class === "backend")
        .map((unit) => unit.aggregateTarget || unit.target)
        .filter(Boolean),
    ),
  ).sort((left, right) => left.localeCompare(right));
  const makeBin = process.env.MAKE || context.makeBin;
  const testOutputScript =
    process.env.TEST_OUTPUT_SCRIPT || context.testOutputScript;

  if (isDryRunFromMakeFlags()) {
    writeSchedulerDryRun({
      repoRoot,
      schedule: {
        ...schedule,
        kind: "service_backed",
        resourceScheduler: "service_backed",
      },
      manifestPath,
      verboseUnitLine(unit) {
        if (unit.countInTotal === false) {
          return "";
        }
        const profile = unit.schedulerProfile
          ? ` profile=${unit.schedulerProfile}`
          : "";
        const needs =
          unit.needs.length > 0 ? ` needs=${unit.needs.join(",")}` : "";
        return `[DRY-RUN] ${schedule.target} unit ${unit.label} type=${unit.type} class=${unit.class}${profile}${needs} claims=${formatResourceMap(unit.resourceClaims)}\n`;
      },
    });
    return;
  }

  const status = await runSchedule({
    schedule,
    makeBin,
    testOutputScript,
    deferSummary: options.deferSummary,
  });
  process.exitCode = status;
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
