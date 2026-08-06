import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  resolveMachineStatePaths,
  resolveOutputMode as resolveHarnessOutputMode,
} from "./harness-contract.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "..", "..", "..");

function repoPath(repoRoot, value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function envPath(name, fallback, repoRoot) {
  const value = process.env[name];
  if (value && value.trim() !== "") {
    return repoPath(repoRoot, value);
  }
  return repoPath(repoRoot, fallback);
}

function resolveNodeBin(repoRoot) {
  const configured = process.env.NODE_BIN;
  if (configured && configured.trim() !== "") {
    return configured;
  }
  const repoNode = path.join(repoRoot, "tmp", "node-runtime", "bin", "node");
  if (existsSync(repoNode)) {
    return repoNode;
  }
  return process.execPath;
}

export function createRunnerContext(options = {}) {
  const repoRoot = options.repoRoot ? path.resolve(options.repoRoot) : defaultRepoRoot;
  const nodeRuntimeDir = envPath("NODE_RUNTIME_DIR", "tmp/node-runtime", repoRoot);
  const nodeBin = resolveNodeBin(repoRoot);
  const machineState = resolveMachineStatePaths(process.env, { root: repoRoot });

  return {
    repoRoot,
    nodeRuntimeDir,
    nodeBin,
    goBin: process.env.GO || "go",
    machineCacheDir: machineState.CARTULARY_MACHINE_CACHE_DIR.value,
    goCacheDir: machineState.GO_CACHE_DIR.value,
    goModCacheDir: machineState.GO_MOD_CACHE_DIR.value,
    goTmpDir: machineState.GO_TMP_DIR.value,
    makeBin: process.env.MAKE_BIN || process.env.MAKE || "make",
    runnerScript: envPath("CARTULARY_RUNNER_SCRIPT", "tools/harness/execution/cartulary-runner-cli.mjs", repoRoot),
    runStepScript: envPath("RUN_STEP_SCRIPT", "tools/harness/execution/run-step.sh", repoRoot),
    schedulerManifest: envPath(
      "SCHEDULER_MANIFEST",
      "tools/scheduler_manifest.json",
      repoRoot,
    ),
    taskSurfaceManifest: envPath(
      "TASK_SURFACE_MANIFEST",
      "tools/task_surface_manifest.json",
      repoRoot,
    ),
    testOutputScript: envPath("TEST_OUTPUT_SCRIPT", "tools/harness/output/test-output.mjs", repoRoot),
    testServicesBin: process.env.TEST_SERVICES_BIN || "",
    outputMode: ["verbose", "debug"].includes(
      resolveHarnessOutputMode(process.env, process.env.CARTULARY_TEST_TARGET || ""),
    )
      ? "normal"
      : "quiet",
    resultsDir: envPath("CARTULARY_TEST_RESULTS_DIR", ".cartulary/test-results", repoRoot),
    runId: process.env.CARTULARY_TEST_RUN_ID || "adhoc",
  };
}

export function runnerEnv(context, extra = {}) {
  return {
    ...process.env,
    MAKE: context.makeBin,
    NODE_BIN: context.nodeBin,
    GO: context.goBin,
    CARTULARY_MACHINE_CACHE_DIR: context.machineCacheDir,
    GO_CACHE_DIR: context.goCacheDir,
    GO_MOD_CACHE_DIR: context.goModCacheDir,
    GO_TMP_DIR: context.goTmpDir,
    TEST_OUTPUT_SCRIPT: context.testOutputScript,
    TASK_SURFACE_MANIFEST: context.taskSurfaceManifest,
    CARTULARY_RUNNER_SCRIPT: context.runnerScript,
    ...extra,
  };
}
