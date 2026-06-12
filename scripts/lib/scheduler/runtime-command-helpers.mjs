import { readFile } from "node:fs/promises";
import path from "node:path";

import { browserStageCompleteRuntimeCommand } from "../scheduler-browser-runtime.mjs";
import { parseSchedulerRunnerArgs } from "../scheduler-cli.mjs";
import { loadSchedulerManifest } from "../scheduler-manifest.mjs";

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

export function browserSessionStartCommand({
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
