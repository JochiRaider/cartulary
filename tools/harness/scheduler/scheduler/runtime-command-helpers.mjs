import { existsSync } from "node:fs";
import { readFile } from "node:fs/promises";
import path from "node:path";

import { browserGroupCommand } from "../adapters/browser.mjs";
import { browserStageCompleteRuntimeCommand } from "../scheduler-browser-runtime.mjs";
import { parseSchedulerRunnerArgs } from "../scheduler-cli.mjs";
import { loadSchedulerManifest } from "../scheduler-manifest.mjs";
import { makeChildEnv, runLifecycle } from "../scheduler-runner.mjs";

export function browserSessionKeyFor(unit) {
  return unit.browserSessionGroup || unit.aggregateTarget || unit.target;
}

export async function readStringEnvFile(file, invalidObjectMessage) {
  const parsed = JSON.parse(await readFile(file, "utf8"));
  if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
    throw new Error(invalidObjectMessage);
  }
  return Object.fromEntries(
    Object.entries(parsed).filter((entry) => typeof entry[1] === "string"),
  );
}

export async function loadSchedulerRunnerManifest(
  argv,
  {
    allowDeferSummary = false,
    defaultManifestPath,
    parseResourceLimitOverride,
    repoRoot,
    schemaID,
    usageText,
  },
) {
  const options = parseSchedulerRunnerArgs(argv, {
    allowDeferSummary,
    defaultManifestPath,
    parseResourceLimitOverride,
    usageText,
  });
  const { manifest, manifestPath } = await loadSchedulerManifest(
    options.manifest,
    {
      repoRoot,
      schemaID,
    },
  );
  return { manifest, manifestPath, options };
}

function browserSessionFileStem(sessionKey) {
  return sessionKey.replaceAll(/[^A-Za-z0-9_.-]/g, "_");
}

export function browserSessionFilesFor(workUnits, rootDir) {
  const sessionUnits = workUnits
    .filter((unit) => unit.kind === "browser_stage_session")
    .sort((left, right) =>
      browserSessionKeyFor(left).localeCompare(browserSessionKeyFor(right)),
    );
  const sessionKeys = Array.from(
    new Set(
      sessionUnits
        .map((unit) => browserSessionKeyFor(unit))
        .filter((target) => target !== ""),
    ),
  ).sort((left, right) => left.localeCompare(right));
  const sessionUnitByKey = new Map(
    sessionUnits.map((unit) => [browserSessionKeyFor(unit), unit]),
  );
  const sessionFiles = new Map(
    sessionKeys.map((sessionKey) => {
      const fileStem = browserSessionFileStem(sessionKey);
      return [
        sessionKey,
        {
          envFile: path.join(rootDir, `${fileStem}-browser-env.json`),
          leaseFile: path.join(rootDir, `${fileStem}-browser-lease.json`),
        },
      ];
    }),
  );
  return { sessionFiles, sessionKeys, sessionUnitByKey, sessionUnits };
}

export function testOutputRuntimeCommand(testOutputScript) {
  return testOutputScript.endsWith(".mjs")
    ? `${JSON.stringify(process.env.NODE_BIN || process.execPath)} ${JSON.stringify(testOutputScript)}`
    : JSON.stringify(testOutputScript);
}

export function defaultBrowserSessionScript(repoRoot) {
  return (
    process.env.CARTULARY_BROWSER_E2E_SESSION_SCRIPT ||
    path.join(repoRoot, "tools", "harness", "browser", "start-web-e2e.sh")
  );
}

export function defaultBrowserGroupRunner() {
  return process.env.CARTULARY_BROWSER_E2E_GROUP_RUNNER || "";
}

export function defaultPnpmBin(repoRoot) {
  return process.env.PNPM || path.join(repoRoot, "tmp", "node-runtime", "bin", "pnpm");
}

export function resolveTestServicesBin(...candidates) {
  return (
    process.env.CARTULARY_TEST_SERVICES_BIN ||
    candidates.find((candidate) => typeof candidate === "string" && candidate.trim() !== "") ||
    process.env.TEST_SERVICES_BIN ||
    ""
  );
}

export function schedulerChildEnv(...parts) {
  return makeChildEnv(Object.assign({}, ...parts.filter(Boolean)));
}

export function makeTargetRuntimeCommand({
  makeBin,
  target,
  env,
  jobs = null,
  outputSync = false,
  skipPrerequisites = null,
}) {
  const args = ["--no-print-directory"];
  if (outputSync) {
    args.push("--output-sync=target");
  }
  if (jobs !== null && jobs !== undefined) {
    args.push(`-j${jobs}`);
  }
  args.push(target);

  const childEnv = { ...(env ?? {}) };
  if (skipPrerequisites !== null) {
    delete childEnv.CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES;
    if (skipPrerequisites) {
      childEnv.CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES = "1";
    }
  }
  return { command: makeBin, args, env: makeChildEnv(childEnv) };
}

export function goShardRuntimeCommand({
  command,
  commandPrefix = [],
  target,
  shard,
  metadataDir,
  env,
}) {
  return {
    command,
    args: [...commandPrefix, "capture-shard", target, shard, metadataDir],
    env,
  };
}

export function goTargetRuntimeCommand({ command, commandPrefix = [], target, env }) {
  return {
    command,
    args: [...commandPrefix, "go-target", target],
    env,
  };
}

export function goFinalizerRuntimeCommand({
  command,
  commandPrefix = [],
  aggregateTarget,
  metadataDir,
  shardNames = [],
  env,
}) {
  return {
    command,
    args: [...commandPrefix, "finalize-shards", aggregateTarget, metadataDir, ...shardNames],
    env,
  };
}

function browserSessionStartCommand({
  browserSessionScript,
  env,
  envFile,
  leaseFile,
}) {
  return {
    command: browserSessionScript,
    args: ["--session-start", "--env-file", envFile, "--lease-file", leaseFile],
    env,
  };
}

export function browserStageSessionRuntimeCommand({
  browserSessionScript,
  env,
  envFile,
  leaseFile,
}) {
  return browserSessionStartCommand({
    browserSessionScript,
    env,
    envFile,
    leaseFile,
  });
}

export function browserGroupRuntimeCommand({
  browserGroupRunner,
  env,
  group,
  pnpmBin,
  repoRoot,
  scriptEnv = {},
}) {
  return browserGroupCommand({
    browserGroupRunner,
    env,
    group,
    pnpmBin,
    repoRoot,
    scriptEnv,
  });
}

export function browserStageCompleteCommand({
  browserSessionScript,
  env,
  leaseFile,
  shouldStopSession,
  target,
  testOutputCommand,
}) {
  return browserStageCompleteRuntimeCommand({
    browserSessionScript,
    env,
    leaseFile,
    shouldStopSession,
    target,
    testOutputCommand,
  });
}

export function browserSessionFinalizerCommand({
  browserSessionScript,
  env,
  leaseFile,
}) {
  return {
    command: browserSessionScript,
    args: ["--session-stop", "--lease-file", leaseFile],
    env,
  };
}

export async function stopBrowserSessionLease({
  repoRoot,
  browserSessionScript,
  leaseFile,
  ignoreErrors = true,
}) {
  if (!leaseFile || !existsSync(leaseFile)) {
    return false;
  }
  const stop = runLifecycle(repoRoot, browserSessionScript, [
    "--session-stop",
    "--lease-file",
    leaseFile,
  ]);
  if (ignoreErrors) {
    await stop.catch(() => {});
  } else {
    await stop;
  }
  return true;
}
