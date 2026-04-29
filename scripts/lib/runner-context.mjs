import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "..", "..");

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

  return {
    repoRoot,
    nodeRuntimeDir,
    nodeBin,
    goBin: process.env.GO || "go",
    goCacheDir: process.env.GO_CACHE_DIR || "/tmp/cartulary-go-build",
    goModCacheDir: process.env.GO_MOD_CACHE_DIR || "/tmp/cartulary-go-mod",
    makeBin: process.env.MAKE_BIN || process.env.MAKE || "make",
    runnerScript: envPath("CARTULARY_RUNNER_SCRIPT", "scripts/cartulary-runner.mjs", repoRoot),
    runPhaseScript: envPath("RUN_PHASE_SCRIPT", "scripts/lib/run-phase.sh", repoRoot),
    runGoTargetScript: envPath("RUN_GO_TARGET_SCRIPT", "scripts/run-go-target.mjs", repoRoot),
    serviceBackedScheduleScript: envPath(
      "RUN_SERVICE_BACKED_SCHEDULE_SCRIPT",
      "scripts/run-service-backed-schedule.mjs",
      repoRoot,
    ),
    serviceBackedScheduleManifest: envPath(
      "SERVICE_BACKED_SCHEDULE_MANIFEST",
      "tools/service_backed_schedule_manifest.json",
      repoRoot,
    ),
    taskSurfaceManifest: envPath(
      "TASK_SURFACE_MANIFEST",
      "tools/task_surface_manifest.json",
      repoRoot,
    ),
    testOutputScript: envPath("TEST_OUTPUT_SCRIPT", "scripts/lib/test-output.mjs", repoRoot),
    testServicesBin: process.env.TEST_SERVICES_BIN || "",
    outputMode:
      process.env.VERBOSE === "1" || process.env.CI_VERBOSE === "1"
        ? "normal"
        : process.env.CARTULARY_OUTPUT_MODE || "quiet",
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
    GO_CACHE_DIR: context.goCacheDir,
    GO_MOD_CACHE_DIR: context.goModCacheDir,
    TEST_OUTPUT_SCRIPT: context.testOutputScript,
    TASK_SURFACE_MANIFEST: context.taskSurfaceManifest,
    SERVICE_BACKED_SCHEDULE_MANIFEST: context.serviceBackedScheduleManifest,
    CARTULARY_RUNNER_SCRIPT: context.runnerScript,
    ...extra,
  };
}
