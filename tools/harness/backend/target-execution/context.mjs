import { existsSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  resolveOutputMode as resolveHarnessOutputMode,
  resolveRetainedArtifactIdentity,
  secureMkdir,
  targetPolicy,
} from "../../contract/index.mjs";
import { resolvePath, slugifyLabel } from "./util.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "..", "..", "..", "..");

function resolveNodeBin(repoRoot, env) {
  if (env.NODE_BIN && env.NODE_BIN.trim() !== "") {
    return env.NODE_BIN;
  }
  const repoNode = path.join(repoRoot, "tmp", "node-runtime", "bin", "node");
  return existsSync(repoNode) ? repoNode : process.execPath;
}

function resolveResultsRoot(repoRoot, env) {
  const configured = env.CARTULARY_TEST_RESULTS_DIR;
  if (!configured) {
    return path.join(repoRoot, ".cartulary", "test-results");
  }
  return path.isAbsolute(configured)
    ? configured
    : path.join(repoRoot, configured);
}

function resolveRunID(env) {
  return (
    env.CARTULARY_TEST_RUN_ID ||
    `${new Date().toISOString().replace(/[-:]/g, "").replace(/\..+/, "Z")}-p${process.pid}`
  );
}

function resolveGoArtifactIdentity(repoRoot, env) {
  if (targetPolicy(env.CARTULARY_TEST_TARGET)?.target_class === "public") {
    const identity = resolveRetainedArtifactIdentity(env.CARTULARY_TEST_TARGET, env, {
      root: repoRoot,
      allowExistingRunRoot: env.CARTULARY_HARNESS_IDENTITY_PREPARED === "1",
    });
    return {
      resultsRoot: identity.result_root,
      runId: identity.run_id,
    };
  }
  return {
    resultsRoot: resolveResultsRoot(repoRoot, env),
    runId: resolveRunID(env),
  };
}

function resolveOutputMode(env) {
  const mode = resolveHarnessOutputMode(env, env.CARTULARY_TEST_TARGET || "");
  return mode === "verbose" || mode === "debug" ? "normal" : "quiet";
}

export function createGoTargetContext(options = {}) {
  const repoRoot = path.resolve(options.repoRoot ?? defaultRepoRoot);
  const baseEnv = { ...process.env, ...(options.env ?? {}) };
  const { resultsRoot, runId } = resolveGoArtifactIdentity(repoRoot, baseEnv);
  const env = {
    ...baseEnv,
    CARTULARY_TEST_RUN_ID: runId,
    CARTULARY_TEST_RESULTS_DIR: resultsRoot,
  };
  const nodeBin = resolveNodeBin(repoRoot, env);
  const availableParallelism = options.availableParallelism ?? Math.max(1, os.availableParallelism());
  if (!Number.isInteger(availableParallelism) || availableParallelism < 1) {
    throw new Error(`invalid available parallelism ${availableParallelism}`);
  }
  return {
    repoRoot,
    env,
    nodeBin,
    goBin: env.GO || "go",
    goCacheDir: env.GO_CACHE_DIR || "/tmp/cartulary-go-build",
    goModCacheDir: env.GO_MOD_CACHE_DIR || "/tmp/cartulary-go-mod",
    goTestPackageParallelism:
      env.GO_TEST_PACKAGE_PARALLELISM ||
      env.GO_TEST_SERVICE_PACKAGE_PARALLELISM ||
      "1",
    availableParallelism,
    goMaxProcs: availableParallelism,
    resultsRoot,
    runId,
    testTarget: env.CARTULARY_TEST_TARGET || "",
    outputMode: resolveOutputMode(env),
    testOutputScript: resolvePath(
      repoRoot,
      env.TEST_OUTPUT_SCRIPT || path.join("tools", "harness", "output", "test-output.mjs"),
    ),
    ownerSelection: env.CARTULARY_GO_TARGET_OWNER || "",
    selectedRowIDs: String(env.CARTULARY_GO_TARGET_ROW_IDS || "")
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean),
    scheduledScope: String(env.CARTULARY_GO_SCHEDULE_SCOPE || "").trim(),
    scheduledRowIDs: String(env.CARTULARY_GO_SCHEDULED_ROW_IDS || "")
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean),
    invocation: null,
    targetPlanRows: null,
    shardPlan: null,
  };
}

export function targetDir(ctx) {
  const target = ctx.testTarget || "adhoc";
  const dir = path.join(ctx.resultsRoot, ctx.runId, target);
  secureMkdir(dir);
  return dir;
}

export function prepareStepArtifactDir(ctx, label) {
  const slug = slugifyLabel(label) || "step";
  const dir = path.join(targetDir(ctx), slug);
  secureMkdir(dir);
  return dir;
}

export function prepareSharedArtifactDir(ctx, name) {
  if (!name) {
    throw new Error("prepareSharedArtifactDir requires name");
  }
  const dir = path.join(ctx.resultsRoot, ctx.runId, "_shared", name);
  secureMkdir(dir);
  return dir;
}
