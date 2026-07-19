import { existsSync } from "node:fs";

import {
  browserGroupRuntimeCommand,
  browserSessionFilesFor,
  browserSessionFinalizerCommand,
  browserSessionKeyFor,
  browserStageCompleteCommand,
  browserStageSessionRuntimeCommand,
  defaultBrowserGroupRunner,
  defaultBrowserSessionScript,
  defaultPnpmBin,
  goFinalizerRuntimeCommand,
  goShardRuntimeCommand,
  makeTargetRuntimeCommand,
  readStringEnvFile,
  resolveTestServicesBin,
  schedulerChildEnv,
  stopBrowserSessionLease,
  testOutputRuntimeCommand,
} from "./runtime-command-helpers.mjs";
import {
  loadRuntimeBinaryRegistry,
  runtimeBinaryAbsoluteEnvForIDs,
} from "../../runtime-binary-registry.mjs";

async function defaultBrowserEnvReader(file) {
  return readStringEnvFile(file, `${file} must contain a JSON environment object`);
}

async function defaultProcessEnv() {
  return process.env;
}

function noOpRuntimeCommand() {
  return {
    command: process.execPath,
    args: ["-e", ""],
    env: process.env,
  };
}

function browserSessionCompletionKey(unit) {
  return `browser_stage_session:${browserSessionKeyFor(unit)}`;
}

function goShardIdentity(aggregateTarget, shard) {
  return `${aggregateTarget}\u0000${shard}`;
}

function defaultServiceTargetForUnit(unit) {
  return typeof unit.serviceSession?.target === "string"
    ? unit.serviceSession.target.trim()
    : "";
}

function defaultMakeTargetEnv(unit) {
  return {
    ...process.env,
    ...unit.env,
    CARTULARY_TEST_TARGET: unit.target,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  };
}

function defaultBrowserStageSessionEnv({ unit, serviceEnv, runtime }) {
  return {
    ...serviceEnv,
    ...unit.env,
    CARTULARY_TEST_SERVICES_BIN: runtime.cartularyTestServicesBin,
    CARTULARY_TEST_TARGET: unit.target,
    CARTULARY_BROWSER_SESSION_GROUP: browserSessionKeyFor(unit),
    CARTULARY_BROWSER_STAGE: unit.browserStage,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  };
}

function defaultBrowserGroupEnv({ unit, group, serviceEnv, sessionEnv, runtime }) {
  return {
    ...serviceEnv,
    ...sessionEnv,
    ...unit.env,
    CARTULARY_TEST_SERVICES_BIN: runtime.cartularyTestServicesBin,
    CARTULARY_TEST_TARGET: unit.aggregateTarget,
    CARTULARY_BROWSER_SESSION_GROUP: browserSessionKeyFor(unit),
    CARTULARY_BROWSER_STAGE: unit.browserStage,
    CARTULARY_BROWSER_GROUP_KIND: group.kind,
    CARTULARY_BROWSER_GROUP_NAME: group.name,
    CARTULARY_BROWSER_GROUP_TARGET: unit.target,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  };
}

function defaultBrowserStageCompleteEnv({ unit, runtime }) {
  return {
    ...process.env,
    ...unit.env,
    CARTULARY_TEST_SERVICES_BIN: runtime.cartularyTestServicesBin,
    CARTULARY_TEST_TARGET: unit.target,
    CARTULARY_BROWSER_SESSION_GROUP: browserSessionKeyFor(unit),
    CARTULARY_BROWSER_STAGE: unit.browserStage,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  };
}

function defaultBrowserSessionFinalizerEnv({ unit, runtime }) {
  return {
    ...process.env,
    ...unit.env,
    CARTULARY_TEST_SERVICES_BIN: runtime.cartularyTestServicesBin,
    CARTULARY_TEST_TARGET: unit.target,
    CARTULARY_BROWSER_SESSION_GROUP: browserSessionKeyFor(unit),
    CARTULARY_BROWSER_STAGE: unit.browserStage,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  };
}

function defaultGoShardEnv({ unit, serviceEnv }) {
  return {
    ...serviceEnv,
    ...unit.env,
    CARTULARY_TEST_TARGET: unit.target,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  };
}

function runtimeBinaryEnv(runtime, ids = []) {
  return runtimeBinaryAbsoluteEnvForIDs(
    loadRuntimeBinaryRegistry({ repoRoot: runtime.repoRoot }),
    ids,
    { repoRoot: runtime.repoRoot, label: "scheduler runtime" },
  );
}

function runtimeBinaryIDsForUnit(unit) {
  return unit.runtimeBinaries ?? unit.runtime_binaries ?? [];
}

function defaultGoFinalizerEnv({ unit, testOutputScript }) {
  return {
    ...process.env,
    CARTULARY_TEST_TARGET: unit.aggregateTarget,
    TEST_OUTPUT_SCRIPT: testOutputScript,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
  };
}

export function createSchedulerRuntimeAttachment({
  repoRoot,
  workUnits,
  tempDir,
  testOutputScript,
  testServicesBin = "",
  browserEnvReader = defaultBrowserEnvReader,
}) {
  const {
    sessionFiles: browserSessionFiles,
    sessionKeys: browserSessionKeys,
    sessionUnitByKey: browserSessionUnitByKey,
    sessionUnits: browserSessionUnits,
  } = browserSessionFilesFor(workUnits, tempDir);
  return {
    repoRoot,
    tempDir,
    testOutputScript,
    testOutputCommand: testOutputRuntimeCommand(testOutputScript),
    testServicesBin,
    cartularyTestServicesBin: resolveTestServicesBin(testServicesBin),
    browserSessionScript: defaultBrowserSessionScript(repoRoot),
    browserGroupRunner: defaultBrowserGroupRunner(),
    browserSessionFiles,
    browserSessionKeys,
    browserSessionUnitByKey,
    browserSessionUnits,
    browserEnvReader,
  };
}

async function browserSessionEnvFor(runtime, target) {
  const files = runtime.browserSessionFiles.get(target);
  return files ? runtime.browserEnvReader(files.envFile) : {};
}

export async function stopSchedulerBrowserSessionLeases(runtime) {
  for (const sessionKey of runtime.browserSessionKeys) {
    const files = runtime.browserSessionFiles.get(sessionKey);
    await stopBrowserSessionLease({
      repoRoot: runtime.repoRoot,
      browserSessionScript: runtime.browserSessionScript,
      leaseFile: files?.leaseFile,
    });
  }
}

export function attachSchedulerRuntimeCommands(
  schedule,
  {
    runtime,
    makeBin,
    goTargetRunner,
    goTargetRunnerPrefix = [],
    serviceTargetForUnit = defaultServiceTargetForUnit,
    serviceEnvFor = defaultProcessEnv,
    metadataDirForUnit = () => runtime.tempDir,
    aggregateMetadataDirForUnit = metadataDirForUnit,
    skipKinds = [],
    makeTargetJobs = (unit) => unit.makeJobs,
    makeTargetOutputSync = () => true,
    makeTargetSkipPrerequisites = (unit) =>
      unit.makePrerequisitePolicy === "skip" ? true : null,
    makeTargetEnv = defaultMakeTargetEnv,
    serviceMakeTargetEnv = async ({ unit, serviceEnv }) => ({
      ...serviceEnv,
      ...unit.env,
      CARTULARY_TEST_TARGET: unit.target,
      CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
    }),
    browserStageSessionEnv = defaultBrowserStageSessionEnv,
    browserGroupEnv = defaultBrowserGroupEnv,
    browserGroupScriptEnv = () => ({ PLAYWRIGHT_WORKERS: "1" }),
    browserStageCompleteEnv = defaultBrowserStageCompleteEnv,
    browserSessionFinalizerEnv = defaultBrowserSessionFinalizerEnv,
    goShardEnv = defaultGoShardEnv,
    goFinalizerEnv = defaultGoFinalizerEnv,
  },
) {
  const skipped = new Set(skipKinds);
  const serviceTarget = (unit) => serviceTargetForUnit(unit);
  const unitServiceEnv = async (unit) => serviceEnvFor(unit, serviceTarget(unit));
  const goShardUnitIDs = new Map(
    schedule.workUnits
      .filter((unit) => unit.kind === "go_shard")
      .map((unit) => [goShardIdentity(unit.aggregateTarget, unit.shard), unit.id]),
  );

  for (const unit of schedule.workUnits) {
    if (skipped.has(unit.kind)) {
      continue;
    }
    if (unit.kind === "make_target") {
      unit.command = () =>
        makeTargetRuntimeCommand({
          makeBin,
          target: unit.target,
          jobs: makeTargetJobs(unit),
          outputSync: makeTargetOutputSync(unit),
          skipPrerequisites: makeTargetSkipPrerequisites(unit),
          env: makeTargetEnv(unit),
        });
      continue;
    }
    if (unit.kind === "service_make_target") {
      unit.command = async () =>
        makeTargetRuntimeCommand({
          makeBin,
          target: unit.target,
          jobs: 1,
          outputSync: true,
          env: await serviceMakeTargetEnv({
            unit,
            runtime,
            serviceEnv: await unitServiceEnv(unit),
          }),
        });
      continue;
    }
    if (unit.kind === "browser_stage_session") {
      const files = runtime.browserSessionFiles.get(browserSessionKeyFor(unit));
      unit.command = async () =>
        browserStageSessionRuntimeCommand({
          browserSessionScript: runtime.browserSessionScript,
          env: schedulerChildEnv(
            await browserStageSessionEnv({
              unit,
              runtime,
              serviceEnv: await unitServiceEnv(unit),
            }),
            runtimeBinaryEnv(runtime, ["server-harness", "migrate"]),
          ),
          envFile: files.envFile,
          leaseFile: files.leaseFile,
        });
      continue;
    }
    if (unit.kind === "browser_group") {
      unit.command = async () => {
        const group = unit.browserGroup;
        const sessionEnv = await browserSessionEnvFor(
          runtime,
          browserSessionKeyFor(unit),
        );
        return browserGroupRuntimeCommand({
          browserGroupRunner: runtime.browserGroupRunner,
          env: schedulerChildEnv(
            await browserGroupEnv({
              unit,
              group,
              runtime,
              serviceEnv: await unitServiceEnv(unit),
              sessionEnv,
            }),
            runtimeBinaryEnv(runtime, ["server-harness", "migrate"]),
          ),
          group,
          pnpmBin: defaultPnpmBin(runtime.repoRoot),
          repoRoot: runtime.repoRoot,
          scriptEnv: browserGroupScriptEnv({ unit, group, runtime }),
        });
      };
      continue;
    }
    if (unit.kind === "browser_stage_complete") {
      const files = runtime.browserSessionFiles.get(browserSessionKeyFor(unit));
      const shouldStopSession = unit.browserSessionFinalizer !== false;
      unit.command = async ({ completedKeys = new Set() } = {}) => {
        const sessionCompleted = completedKeys.has(browserSessionCompletionKey(unit));
        const leaseExists = existsSync(files.leaseFile);
        if (!sessionCompleted && !leaseExists) {
          return noOpRuntimeCommand();
        }
        return browserStageCompleteCommand({
          browserSessionScript: runtime.browserSessionScript,
          env: schedulerChildEnv(
            await browserStageCompleteEnv({
              unit,
              runtime,
              serviceEnv: await unitServiceEnv(unit),
            }),
            runtimeBinaryEnv(runtime, ["server-harness", "migrate"]),
          ),
          emitPassSummary: unit.needs.every((need) => completedKeys.has(need)),
          leaseFile: files.leaseFile,
          shouldStopSession,
          target: unit.target,
          testOutputCommand: runtime.testOutputCommand,
        });
      };
      continue;
    }
    if (unit.kind === "browser_session_finalizer") {
      const files = runtime.browserSessionFiles.get(browserSessionKeyFor(unit));
      unit.command = async ({ completedKeys = new Set() } = {}) => {
        const sessionCompleted = completedKeys.has(browserSessionCompletionKey(unit));
        if (!sessionCompleted && !existsSync(files.leaseFile)) {
          return noOpRuntimeCommand();
        }
        return browserSessionFinalizerCommand({
          browserSessionScript: runtime.browserSessionScript,
          env: schedulerChildEnv(
            await browserSessionFinalizerEnv({
              unit,
              runtime,
              serviceEnv: await unitServiceEnv(unit),
            }),
            runtimeBinaryEnv(runtime, ["server-harness", "migrate"]),
          ),
          leaseFile: files.leaseFile,
        });
      };
      continue;
    }
    if (unit.kind === "go_shard") {
      unit.command = async () =>
        goShardRuntimeCommand({
          command: goTargetRunner,
          commandPrefix: goTargetRunnerPrefix,
          target: unit.target,
          shard: unit.shard,
          metadataDir: metadataDirForUnit(unit),
          env: schedulerChildEnv(
            await goShardEnv({
              unit,
              runtime,
              serviceEnv: await unitServiceEnv(unit),
            }),
            runtimeBinaryEnv(runtime, runtimeBinaryIDsForUnit(unit)),
          ),
        });
      continue;
    }
    if (unit.kind === "aggregate_finalize") {
      unit.command = async ({ startedUnitIDs = new Set() } = {}) => {
        const startedShardNames = unit.shardNames.filter((shard) => {
          const shardUnitID = goShardUnitIDs.get(
            goShardIdentity(unit.aggregateTarget, shard),
          );
          return shardUnitID !== undefined && startedUnitIDs.has(shardUnitID);
        });
        if (startedShardNames.length === 0) {
          return noOpRuntimeCommand();
        }
        return goFinalizerRuntimeCommand({
          command: goTargetRunner,
          commandPrefix: goTargetRunnerPrefix,
          aggregateTarget: unit.aggregateTarget,
          metadataDir: aggregateMetadataDirForUnit(unit),
          shardNames: startedShardNames,
          env: schedulerChildEnv(
            await goFinalizerEnv({
              unit,
              runtime,
              serviceEnv: await unitServiceEnv(unit),
              testOutputScript: runtime.testOutputScript,
            }),
          ),
        });
      };
    }
  }

  return schedule;
}
