import { spawn, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import {
  appendFileSync,
  accessSync,
  constants,
  existsSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import {
  aggregatePackages,
  aggregateRegex,
  collectAggregateEmissions,
  fixturePolicyAssignments,
  resetTableAssignments,
} from "./go-target-aggregate.mjs";
import { collectGoShardPlanFromRows } from "./go-shard-plan.mjs";
import {
  createFailureClassCounts,
  createFailureReasonCounts,
} from "../contract/index.mjs";
import {
  createSecureWriteStream,
  redactString,
  resolveOutputMode as resolveHarnessOutputMode,
  resolveRetainedArtifactIdentity,
  secureMkdir,
  secureWriteFile,
  targetPolicy,
} from "../contract/index.mjs";
import { testCoverageBuckets } from "../contract/test-output-context.mjs";
import {
  collectTargetPlanRows,
  findTargetDescriptor,
} from "./backend-target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "..", "..", "..");

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

function nowUTC() {
  return new Date().toISOString();
}

function monotonicMs() {
  return Number(process.hrtime.bigint() / 1_000_000n);
}

function clampDurationMs(value) {
  const numeric = Number.parseInt(String(value ?? "0"), 10);
  if (!Number.isInteger(numeric) || numeric < 0) {
    return 0;
  }
  return numeric;
}

function captureStart() {
  return {
    startTime: nowUTC(),
    startMs: monotonicMs(),
  };
}

function captureFinish(started) {
  return {
    startTime: started.startTime,
    endTime: nowUTC(),
    durationMs: clampDurationMs(monotonicMs() - started.startMs),
  };
}

function resolvePath(repoRoot, value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function relToRepo(ctx, value) {
  if (!value) {
    return "";
  }
  const normalized = String(value).replaceAll("\\", "/");
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = path.relative(ctx.repoRoot, value).replaceAll("\\", "/");
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

const runtimeBinaryRegistry = Object.freeze({
  operator: Object.freeze({
    id: "operator",
    producerTarget: "build-operator",
    consumerEnv: "CARTULARY_OPERATOR_BIN",
  }),
});

class RuntimeBinaryError extends Error {
  constructor(message, { exitCode, reason }) {
    super(message);
    this.name = "RuntimeBinaryError";
    this.exitCode = exitCode;
    this.reason = reason;
  }
}

function runtimeBinaryIDsForRows(rows) {
  return Array.from(new Set(rows.flatMap((row) => row.runtime_binaries ?? []))).sort(compareStrings);
}

function fileSha256(file) {
  return `sha256:${createHash("sha256").update(readFileSync(file)).digest("hex")}`;
}

function buildArtifactOutputDigest(ctx, file) {
  const display = relToRepo(ctx, file);
  const material = `output\t${display}\t${fileSha256(file)}\n`;
  return `sha256:${createHash("sha256").update(material).digest("hex")}`;
}

function runtimeBinaryPath(ctx, record) {
  const raw = ctx.env[record.consumerEnv] ?? "";
  if (String(raw).trim() === "") {
    throw new RuntimeBinaryError(
      `${record.consumerEnv} is required for runtime binary ${record.id}`,
      { exitCode: 2, reason: "configuration_error" },
    );
  }
  if (String(raw).includes("\0")) {
    throw new RuntimeBinaryError(`${record.consumerEnv} must not contain NUL`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  return resolvePath(ctx.repoRoot, String(raw).trim());
}

function readBuildArtifact(ctx, record) {
  const file = path.join(
    ctx.resultsRoot,
    ctx.runId,
    record.producerTarget,
    `build-artifact-cache-${record.producerTarget}.json`,
  );
  if (!existsSync(file)) {
    throw new RuntimeBinaryError(
      `missing ${record.producerTarget} build-artifact cache reference for runtime binary ${record.id}`,
      { exitCode: 11, reason: "artifact_error" },
    );
  }
  try {
    return { file, artifact: JSON.parse(readFileSync(file, "utf8")) };
  } catch (error) {
    throw new RuntimeBinaryError(
      `invalid ${record.producerTarget} build-artifact cache reference: ${error.message}`,
      { exitCode: 11, reason: "artifact_error" },
    );
  }
}

function validateRuntimeBinary(ctx, id) {
  const record = runtimeBinaryRegistry[id];
  if (!record) {
    throw new RuntimeBinaryError(`unknown runtime binary ${id}`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  const binaryPath = runtimeBinaryPath(ctx, record);
  let info;
  try {
    info = lstatSync(binaryPath);
  } catch {
    throw new RuntimeBinaryError(`${record.consumerEnv} does not exist: ${relToRepo(ctx, binaryPath)}`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  if (info.isSymbolicLink() || !info.isFile()) {
    throw new RuntimeBinaryError(`${record.consumerEnv} must name a regular executable file`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  try {
    accessSync(binaryPath, constants.X_OK);
  } catch {
    throw new RuntimeBinaryError(`${record.consumerEnv} is not executable: ${relToRepo(ctx, binaryPath)}`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  const stat = statSync(binaryPath);
  if (!stat.isFile()) {
    throw new RuntimeBinaryError(`${record.consumerEnv} must name a regular executable file`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  const { file: artifactFile, artifact } = readBuildArtifact(ctx, record);
  const expectedDigest = buildArtifactOutputDigest(ctx, binaryPath);
  if (artifact.output_digest_sha256 !== expectedDigest) {
    throw new RuntimeBinaryError(
      `${record.producerTarget} artifact digest does not match ${record.consumerEnv}`,
      { exitCode: 11, reason: "artifact_error" },
    );
  }
  return {
    id,
    producer_target: record.producerTarget,
    consumer_env: record.consumerEnv,
    source: "scheduler-produced",
    path: relToRepo(ctx, binaryPath),
    sha256: fileSha256(binaryPath),
    build_artifact_ref: relToRepo(ctx, artifactFile),
    build_artifact_output_digest: artifact.output_digest_sha256,
  };
}

function validateRuntimeBinaries(ctx, rows, reportDir) {
  const ids = runtimeBinaryIDsForRows(rows);
  if (ids.length === 0) {
    return [];
  }
  const records = ids.map((id) => validateRuntimeBinary(ctx, id));
  secureWriteFile(
    path.join(reportDir, "runtime-binaries.json"),
    `${JSON.stringify({ runtime_binaries: records }, null, 2)}\n`,
  );
  return records;
}

function runtimeRowsForExecution(ctx, target, familyOrShard) {
  const shard = findShardOrNull(ctx, target, familyOrShard);
  if (!shard) {
    return rowsForAggregate(ctx, target, familyOrShard);
  }
  const rowIDs = new Set(
    (shard.items ?? []).map((item) => String(item.id ?? "").split(":")[0]),
  );
  return targetRows(ctx).filter((row) => row.target === target && rowIDs.has(row.id));
}

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

function shellQuote(value) {
  const text = String(value);
  if (/^[A-Za-z0-9_@%+=:,./-]+$/u.test(text)) {
    return text;
  }
  return `'${text.replaceAll("'", "'\"'\"'")}'`;
}

export function renderCommand(args) {
  return args.map(shellQuote).join(" ");
}

function slugifyLabel(label) {
  return String(label)
    .toLowerCase()
    .replace(/[^a-z0-9]+/gu, "-")
    .replace(/^-+|-+$/gu, "")
    .replace(/--+/gu, "-");
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
    backendIntegrationShardJobs:
      Number.parseInt(env.BACKEND_INTEGRATION_SHARD_JOBS || "4", 10) || 4,
    resultsRoot,
    runId,
    testTarget: env.CARTULARY_TEST_TARGET || "",
    outputMode: resolveOutputMode(env),
    testOutputScript: resolvePath(
      repoRoot,
      env.TEST_OUTPUT_SCRIPT || path.join("tools", "harness", "core", "test-output.mjs"),
    ),
    phaseSelection: env.CARTULARY_GO_TARGET_PHASE || "",
    invocation: null,
    targetPlanRows: null,
    shardPlan: null,
  };
}

function targetDir(ctx) {
  const target = ctx.testTarget || "adhoc";
  const dir = path.join(ctx.resultsRoot, ctx.runId, target);
  secureMkdir(dir);
  return dir;
}

function preparePhaseArtifactDir(ctx, label) {
  const slug = slugifyLabel(label) || "phase";
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

function targetRows(ctx) {
  if (!ctx.targetPlanRows) {
    ctx.targetPlanRows = collectTargetPlanRows(ctx.repoRoot).filter((row) => {
      if (!ctx.phaseSelection) {
        return true;
      }
      return row.manifest_phase === ctx.phaseSelection;
    });
  }
  return ctx.targetPlanRows;
}

function shardPlan(ctx) {
  if (!ctx.shardPlan) {
    ctx.shardPlan = collectGoShardPlanFromRows(ctx.repoRoot, targetRows(ctx), {
      phase: ctx.phaseSelection,
    });
  }
  return ctx.shardPlan;
}

function rowsForAggregate(ctx, target, family) {
  const rows = targetRows(ctx).filter(
    (row) => row.target === target && row.execution_family === family,
  );
  if (rows.length === 0) {
    throw new Error(`unknown execution family ${family} for ${target}`);
  }
  return rows;
}

function rowPackages(row) {
  if (row.package) {
    return [row.package];
  }
  return [...(row.packages ?? [])];
}

function rowMatchesShardItems(row, items) {
  const rowIDs = new Set(
    items.map((item) => String(item.id ?? "").split(":")[0]),
  );
  if (rowIDs.has(row.id)) {
    return true;
  }
  if (row.coverage === "raw") {
    const packages = new Set(items.flatMap((item) => item.packages ?? []));
    return rowPackages(row).some((pkg) => packages.has(pkg));
  }
  const symbols = new Set(
    items
      .map((item) => item.symbol ?? "")
      .filter((symbol) => symbol !== ""),
  );
  return (row.symbols ?? []).some((symbol) => symbols.has(symbol));
}

function rowsForScheduledAggregate(ctx, target, aggregateName, shardNames) {
  const rows = rowsForAggregate(ctx, target, aggregateName);
  if (shardNames.length === 0) {
    return rows;
  }
  const selectedItems = shardNames.flatMap(
    (shardName) => findShard(ctx, target, shardName).items ?? [],
  );
  const filtered = rows.filter((row) => rowMatchesShardItems(row, selectedItems));
  if (filtered.length === 0) {
    throw new Error(
      `scheduled shards for ${target}:${aggregateName} matched zero manifest rows`,
    );
  }
  return filtered;
}

function aggregateNames(ctx, target) {
  return Array.from(
    new Set(
      targetRows(ctx)
        .filter((row) => row.target === target)
        .map((row) => row.execution_family),
    ),
  ).sort(compareStrings);
}

function targetOwnsShard(target, shard) {
  if (target === "backend-integration") {
    return shard.items.some(
      (item) => item.kind === "authoritative" || item.kind === "raw",
    );
  }
  if (target === "backend-integration-support") {
    return shard.items.some((item) => item.kind === "support");
  }
  if (target === "backend-store") {
    return shard.items.some((item) => item.kind === "authoritative");
  }
  if (target === "backend-process") {
    return shard.items.some((item) => item.target === "backend-process");
  }
  return false;
}

function targetShards(ctx, target) {
  const plan = shardPlan(ctx);
  const aggregateSet = new Set(
    plan.aggregates
      .filter((aggregate) => aggregate.target === target)
      .map((aggregate) => aggregate.name),
  );
  return plan.shards.filter(
    (shard) =>
      aggregateSet.has(shard.aggregate_name) && targetOwnsShard(target, shard),
  );
}

function targetAggregates(ctx, target) {
  return shardPlan(ctx).aggregates.filter(
    (aggregate) => aggregate.target === target,
  );
}

function findShard(ctx, target, name) {
  const shard = findShardOrNull(ctx, target, name);
  if (!shard) {
    throw new Error(`unknown shard ${name} for ${target}`);
  }
  return shard;
}

function findShardOrNull(ctx, target, name) {
  return (
    targetShards(ctx, target).find(
      (candidate) => candidate.name === name,
    ) ?? null
  );
}

function fixturePolicyAssignmentsForShard(shard, mode) {
  const assignments = [];
  for (const item of shard.items) {
    if (
      !item.postgres_fixture_policy ||
      item.postgres_fixture_policy === "migration_scratch"
    ) {
      continue;
    }
    if (mode === "tests" && item.symbol) {
      assignments.push(`${item.symbol}=${item.postgres_fixture_policy}`);
    }
    if (mode === "packages" && item.kind === "raw") {
      for (const pkg of item.packages) {
        assignments.push(`${pkg}=${item.postgres_fixture_policy}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

function resetTableAssignmentsForShard(shard, mode) {
  const assignments = [];
  for (const item of shard.items) {
    const dirtyTables = item.postgres_fixture_budget?.dirty_tables ?? [];
    if (dirtyTables.length === 0) {
      continue;
    }
    if (mode === "tests" && item.symbol) {
      assignments.push(`${item.symbol}=${dirtyTables.join("|")}`);
    }
    if (mode === "packages" && item.kind === "raw") {
      for (const pkg of item.packages) {
        assignments.push(`${pkg}=${dirtyTables.join("|")}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

function targetGoTestArgs(ctx, target) {
  const descriptor = findTargetDescriptor(target, ctx.repoRoot);
  if (!descriptor) {
    throw new Error(`unknown target ${target}`);
  }
  switch (descriptor.goTestParallelism) {
    case "none":
      return [];
    case "package":
      return ["-p", String(ctx.goTestPackageParallelism)];
    case "process":
      return ["-parallel", "4"];
    default:
      throw new Error(
        `unsupported go_test_parallelism ${descriptor.goTestParallelism} for ${target}`,
      );
  }
}

function aggregateSpec(ctx, target, family) {
  const rows = rowsForAggregate(ctx, target, family);
  return {
    regex: aggregateRegex(rows),
    args: [...targetGoTestArgs(ctx, target), ...aggregatePackages(rows)],
  };
}

function shardSpec(ctx, target, name) {
  const shard = findShard(ctx, target, name);
  return {
    regex: shard.regex,
    args: [...targetGoTestArgs(ctx, target), ...shard.packages],
  };
}

function resolveExecutionFamilySpec(ctx, target, familyOrShard) {
  if (findShardOrNull(ctx, target, familyOrShard)) {
    return shardSpec(ctx, target, familyOrShard);
  }
  return aggregateSpec(ctx, target, familyOrShard);
}

function resolveExecutionFamilyPostgresFixturePolicy(
  ctx,
  target,
  familyOrShard,
) {
  const shard = findShardOrNull(ctx, target, familyOrShard);
  if (shard) {
    return {
      tests: fixturePolicyAssignmentsForShard(shard, "tests").join(","),
      packages: fixturePolicyAssignmentsForShard(shard, "packages").join(","),
      resetTests: resetTableAssignmentsForShard(shard, "tests").join(","),
      resetPackages: resetTableAssignmentsForShard(shard, "packages").join(","),
    };
  }
  const rows = rowsForAggregate(ctx, target, familyOrShard);
  return {
    tests: fixturePolicyAssignments(rows, "tests").join(","),
    packages: fixturePolicyAssignments(rows, "packages").join(","),
    resetTests: resetTableAssignments(rows, "tests").join(","),
    resetPackages: resetTableAssignments(rows, "packages").join(","),
  };
}

function fixtureEnv(policy = {}) {
  return {
    CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS: policy.tests ?? "",
    CARTULARY_POSTGRES_FIXTURE_POLICY_PACKAGES: policy.packages ?? "",
    CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT: policy.defaultPolicy ?? "",
    CARTULARY_POSTGRES_RESET_TABLES_TESTS: policy.resetTests ?? "",
    CARTULARY_POSTGRES_RESET_TABLES_PACKAGES: policy.resetPackages ?? "",
  };
}

function goTestEnvAssignments(ctx, policy = {}) {
  const assignments = [
    `GOCACHE=${ctx.goCacheDir}`,
    `GOMODCACHE=${ctx.goModCacheDir}`,
  ];
  const values = fixtureEnv(policy);
  for (const [name, value] of Object.entries(values)) {
    if (value) {
      assignments.push(`${name}=${value}`);
    }
  }
  return assignments;
}

export function renderGoTestCommand(ctx, regex, args, policy = {}) {
  return renderCommand([
    "env",
    ...goTestEnvAssignments(ctx, policy),
    ctx.goBin,
    "test",
    "-json",
    "-run",
    regex,
    ...args,
  ]);
}

function goChildEnv(ctx, policy = {}) {
  return {
    ...ctx.env,
    GOCACHE: ctx.goCacheDir,
    GOMODCACHE: ctx.goModCacheDir,
    ...fixtureEnv(policy),
  };
}

function hashGoTestDependencyInputs(ctx) {
  const hash = createHash("sha256");
  const version = spawnSync(ctx.goBin, ["version"], {
    cwd: ctx.repoRoot,
    encoding: "utf8",
  });
  hash.update(version.stdout || "");
  hash.update("\n-- go.mod --\n");
  hash.update(readFileSync(path.join(ctx.repoRoot, "go.mod")));
  hash.update("\n-- go.sum --\n");
  hash.update(readFileSync(path.join(ctx.repoRoot, "go.sum")));
  return hash.digest("hex");
}

async function acquireDirectoryLock(lockDir, label, timeoutSeconds, metadata) {
  const started = monotonicMs();
  while (true) {
    try {
      mkdirSync(lockDir, { recursive: false, mode: 0o700 });
      for (const [name, value] of Object.entries(metadata)) {
        secureWriteFile(path.join(lockDir, name), `${value}\n`);
      }
      return;
    } catch (error) {
      if (error?.code !== "EEXIST") {
        throw error;
      }
    }

    const ownerPid = Number.parseInt(
      existsSync(path.join(lockDir, "pid"))
        ? readFileSync(path.join(lockDir, "pid"), "utf8")
        : "",
      10,
    );
    if (Number.isInteger(ownerPid)) {
      try {
        process.kill(ownerPid, 0);
      } catch (error) {
        if (error?.code === "ESRCH") {
          rmSync(lockDir, { recursive: true, force: true });
          continue;
        }
      }
    }

    if (monotonicMs() - started >= timeoutSeconds * 1000) {
      throw new Error(`${label}_timeout lock=${lockDir}`);
    }
    await sleep(100);
  }
}

async function warmGoTestDependencies(ctx) {
  const warmRoot = path.join(ctx.goModCacheDir, ".cartulary-go-test-warm");
  const lockDir = path.join(warmRoot, "lock");
  const timeoutSeconds = Number.parseInt(
    ctx.env.CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS || "300",
    10,
  );
  if (!Number.isInteger(timeoutSeconds) || timeoutSeconds < 1) {
    throw new Error(
      `invalid CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS=${ctx.env.CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS}`,
    );
  }
  mkdirSync(warmRoot, { recursive: true });
  const warmKey = hashGoTestDependencyInputs(ctx);
  const stampFile = path.join(warmRoot, `${warmKey}.stamp`);
  if (existsSync(stampFile)) {
    return;
  }
  await acquireDirectoryLock(
    lockDir,
    "go_test_dependency_warm_lock",
    timeoutSeconds,
    {
      pid: process.pid,
      acquired_at: nowUTC(),
    },
  );
  try {
    if (existsSync(stampFile)) {
      return;
    }
    for (const args of [
      ["mod", "download"],
      ["list", "-deps", "-test", "./..."],
    ]) {
      const result = spawnSync(ctx.goBin, args, {
        cwd: ctx.repoRoot,
        env: goChildEnv(ctx),
        stdio: "inherit",
      });
      if ((result.status ?? 1) !== 0) {
        process.exitCode = result.status ?? 1;
        throw new Error(
          `${ctx.goBin} ${args.join(" ")} exited ${result.status ?? 1}`,
        );
      }
    }
    writeFileSync(
      `${stampFile}.${process.pid}`,
      `warmed_at=${nowUTC()}\ngo=${spawnSync(ctx.goBin, ["version"], { encoding: "utf8" }).stdout || ""}`,
    );
    rmSync(stampFile, { force: true });
    writeFileSync(stampFile, readFileSync(`${stampFile}.${process.pid}`));
    rmSync(`${stampFile}.${process.pid}`, { force: true });
  } finally {
    rmSync(lockDir, { recursive: true, force: true });
  }
}

async function runHelper(ctx, args, env = {}) {
  const command = ctx.testOutputScript.endsWith(".mjs")
    ? ctx.nodeBin
    : ctx.testOutputScript;
  const commandArgs = ctx.testOutputScript.endsWith(".mjs")
    ? [ctx.testOutputScript, ...args]
    : args;
  return await new Promise((resolve, reject) => {
    const child = spawn(command, commandArgs, {
      cwd: ctx.repoRoot,
      env: {
        ...ctx.env,
        NODE_BIN: ctx.nodeBin,
        ...env,
      },
      stdio: "inherit",
    });
    child.on("error", reject);
    child.on("close", (status) => resolve(status ?? 1));
  });
}

async function emitTargetTimingSpan(
  ctx,
  bucket,
  label,
  window,
  status,
  exitStatus,
) {
  if (!ctx.testTarget) {
    return;
  }
  await runHelper(ctx, ["timing-span"], {
    CARTULARY_TEST_TARGET: ctx.testTarget,
    CARTULARY_TIMING_BUCKET: bucket,
    CARTULARY_TIMING_LABEL: label,
    CARTULARY_TIMING_START_TIME: window.startTime,
    CARTULARY_TIMING_END_TIME: window.endTime,
    CARTULARY_TIMING_DURATION_MS: String(window.durationMs),
    CARTULARY_TIMING_STATUS: status,
    CARTULARY_PHASE_EXIT_STATUS: String(exitStatus),
  });
}

function timingSpanArtifactPath(ctx, label) {
  const dir = path.join(targetDir(ctx), "timing-spans");
  secureMkdir(dir);
  const slug = slugifyLabel(label) || "timing-span";
  const timestamp = nowUTC().replace(/[:.]/g, "-");
  return path.join(dir, `${timestamp}-${process.pid}-${slug}.json`);
}

function writeTargetTimingSpan(ctx, bucket, label, window, status) {
  if (!ctx.testTarget) {
    return;
  }
  secureWriteFile(
    timingSpanArtifactPath(ctx, label),
    `${JSON.stringify({
      source: "target",
      bucket,
      label,
      start_time: window.startTime,
      end_time: window.endTime,
      duration_ms: window.durationMs,
      status,
    })}\n`,
  );
}

function createNonTestFailureCounts() {
  const counts = {
    tests: 0,
    failed: 1,
    non_test: 1,
    non_test_failed: 1,
    packages: 0,
  };
  for (const coverage of testCoverageBuckets) {
    counts[coverage] = 0;
    counts[`${coverage}_failed`] = 0;
  }
  return counts;
}

function finalizerErrorClassification(error) {
  const message = String(error?.message ?? error ?? "");
  if (message.startsWith("unknown scheduled shard ")) {
    return {
      failureClass: "harness",
      failureReason: "scheduler_accounting_error",
    };
  }
  return {
    failureClass: "artifact",
    failureReason: "artifact_error",
  };
}

function writeFinalizerFailurePhase(
  ctx,
  {
    target,
    label,
    commandArgs,
    window,
    exitStatus = 1,
    error,
    metadataDir,
    aggregateReportDir = "",
    shardNames = [],
  },
) {
  if (!ctx.testTarget) {
    return;
  }
  const { failureClass, failureReason } = finalizerErrorClassification(error);
  const phaseDir = preparePhaseArtifactDir(ctx, label);
  const counts = createNonTestFailureCounts();
  const failureClasses = createFailureClassCounts();
  failureClasses[failureClass] = 1;
  const failureReasons = createFailureReasonCounts();
  failureReasons[failureReason] = 1;
  const artifacts = {
    metadata_dir: relToRepo(ctx, metadataDir),
  };
  if (aggregateReportDir) {
    artifacts.aggregate_report_dir = relToRepo(ctx, aggregateReportDir);
  }
  const message = String(error?.message ?? error ?? "go shard finalizer failed");
  const failure = {
    failure_class: failureClass,
    failure_reason: failureReason,
    kind: failureClass === "artifact" ? "artifact" : "failure",
    source: "go-shard-finalizer",
    target,
    label,
    message,
    artifact: artifacts.aggregate_report_dir || artifacts.metadata_dir,
    shard_names: shardNames,
  };
  secureWriteFile(
    path.join(phaseDir, "phase-summary.json"),
    `${JSON.stringify(
      {
        schema_id: "cartulary.test_phase_summary.v3",
        label,
        target: ctx.testTarget,
        runner: "go-shard-finalizer",
        status: "fail",
        phase: "go-shard-finalize",
        command: renderCommand(commandArgs),
        start_time: window.startTime,
        end_time: window.endTime,
        accounting_mode: "actual",
        executed_duration_ms: window.durationMs,
        logical_duration_ms: window.durationMs,
        reused_duration_ms: 0,
        derived_duration_ms: 0,
        wall_duration_ms: window.durationMs,
        critical_path_wall_duration_ms: window.durationMs,
        teardown_duration_ms: 0,
        timing_bucket: "report_collation",
        exit_status: exitStatus,
        counting_mode: "counted",
        artifacts,
        counts,
        failure_class: failureClass,
        failure_reason: failureReason,
        failure_classes: failureClasses,
        failure_reasons: failureReasons,
        failures: [failure],
        failure_headline: `${failureClass} reason=${failureReason} ${message}`,
        owners: [],
        inventory: [],
        dossiers: [],
        manifest_mismatch: null,
      },
      null,
      2,
    )}\n`,
  );
}

function resolveFinalizerEmitJobs(ctx, count) {
  if (count <= 0) {
    return 0;
  }
  const configured = ctx.env.CARTULARY_GO_TARGET_FINALIZER_EMIT_JOBS;
  if (configured) {
    const parsed = Number.parseInt(configured, 10);
    if (!Number.isInteger(parsed) || parsed < 1) {
      throw new Error(
        `invalid CARTULARY_GO_TARGET_FINALIZER_EMIT_JOBS=${configured}`,
      );
    }
    return Math.min(count, parsed);
  }
  return Math.min(count, 4);
}

async function runBounded(items, jobs, worker) {
  if (items.length === 0) {
    return [];
  }
  const workerCount = Math.min(items.length, Math.max(1, jobs));
  const results = new Array(items.length);
  let next = 0;
  await Promise.all(
    Array.from({ length: workerCount }, async () => {
      while (next < items.length) {
        const index = next;
        next += 1;
        results[index] = await worker(items[index], index);
      }
    }),
  );
  return results;
}

async function emitTargetSummary(ctx, status) {
  if (!ctx.testTarget) {
    return 0;
  }
  return await runHelper(ctx, ["target-summary", ctx.testTarget, status]);
}

async function emitGoTargetInvocationSpan(ctx, status) {
  if (!ctx.invocation || ctx.invocation.emitted) {
    return;
  }
  const window = captureFinish(ctx.invocation);
  ctx.invocation.emitted = true;
  await emitTargetTimingSpan(
    ctx,
    "test_command",
    `run-go-target ${ctx.testTarget || "unknown"}`,
    window,
    status === 0 ? "pass" : "fail",
    status,
  );
}

async function finishTarget(ctx, status) {
  await emitGoTargetInvocationSpan(ctx, status);
  if (status === 0) {
    return await emitTargetSummary(ctx, "pass");
  }
  await emitTargetSummary(ctx, "fail").catch(() => {});
  return status;
}

async function runGoTestCapture(ctx, regex, args, reportDir, policy = {}) {
  await warmGoTestDependencies(ctx);
  const runnerLog = path.join(reportDir, "runner.jsonl");
  const stderrLog = path.join(reportDir, "stderr.log");
  const stdoutStream = createSecureWriteStream(runnerLog);
  const stderrStream = createSecureWriteStream(stderrLog);
  let stdoutBuffer = "";

  const child = spawn(ctx.goBin, ["test", "-json", "-run", regex, ...args], {
    cwd: ctx.repoRoot,
    env: goChildEnv(ctx, policy),
    stdio: ["ignore", "pipe", "pipe"],
  });
  child.stdout.on("data", (chunk) => {
    const redactedChunk = redactString(chunk.toString("utf8"));
    stdoutStream.write(redactedChunk);
    if (ctx.outputMode === "quiet") {
      return;
    }
    stdoutBuffer += redactedChunk;
    const lines = stdoutBuffer.split(/\r?\n/u);
    stdoutBuffer = lines.pop() ?? "";
    for (const line of lines) {
      if (!line.trim()) {
        continue;
      }
      try {
        const entry = JSON.parse(line);
        if (typeof entry.Output === "string") {
          process.stderr.write(entry.Output);
        }
      } catch {
        // Preserve raw JSON in the log; malformed streaming lines are not lifecycle output.
      }
    }
  });
  child.stderr.on("data", (chunk) => {
    const redactedChunk = redactString(chunk.toString("utf8"));
    stderrStream.write(redactedChunk);
    if (ctx.outputMode !== "quiet") {
      process.stderr.write(redactedChunk);
    }
  });
  return await new Promise((resolve, reject) => {
    child.on("error", reject);
    child.on("close", (status) => {
      stdoutStream.end();
      stderrStream.end();
      resolve(status ?? 1);
    });
  });
}

export async function acquireSharedReportLock(ctx, sharedDir, sharedName) {
  const timeoutSeconds = Number.parseInt(
    ctx.env.CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS || "300",
    10,
  );
  if (!Number.isInteger(timeoutSeconds) || timeoutSeconds < 1) {
    throw new Error(
      `invalid CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS=${ctx.env.CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS}`,
    );
  }
  try {
    await acquireDirectoryLock(
      path.join(sharedDir, "capture.lock"),
      "shared_go_report_lock",
      timeoutSeconds,
      {
        pid: process.pid,
        shared_report: sharedName,
        acquired_at: nowUTC(),
      },
    );
  } catch (error) {
    if (String(error.message).startsWith("shared_go_report_lock_timeout")) {
      process.stderr.write(
        `shared_go_report_lock_timeout report=${sharedName} lock=${path.join(sharedDir, "capture.lock")}\n`,
      );
    }
    throw error;
  }
}

export function releaseSharedReportLock(sharedDir) {
  rmSync(path.join(sharedDir, "capture.lock"), {
    recursive: true,
    force: true,
  });
}

function isCrossTargetSharedReport(ctx, target, sharedName) {
  return findShardOrNull(ctx, target, sharedName)?.shared_across_targets === true;
}

async function writeCrossTargetSharedExecutionMetadata(
  ctx,
  sharedDir,
  sharedName,
  window,
  status,
) {
  if (
    !isCrossTargetSharedReport(ctx, "backend-integration", sharedName) &&
    !isCrossTargetSharedReport(ctx, "backend-integration-support", sharedName)
  ) {
    return;
  }
  await runHelper(ctx, [
    "shared-execution",
    "backend-integration-shards",
    sharedName,
    status === 0 ? "pass" : "fail",
    window.startTime,
    window.endTime,
    String(window.durationMs),
    String(status),
    path.join(sharedDir, "shared-execution.json"),
  ]);
}

export async function captureGoReportLocked(
  ctx,
  sharedDir,
  sharedName,
  regex,
  args,
  policy = {},
) {
  const commandText = renderGoTestCommand(ctx, regex, args, policy);
  const completeFile = path.join(sharedDir, "complete");
  if (existsSync(completeFile)) {
    const existingCommand = existsSync(path.join(sharedDir, "command.txt"))
      ? readFileSync(path.join(sharedDir, "command.txt"), "utf8").trimEnd()
      : "";
    if (existingCommand !== commandText) {
      throw new Error(
        [
          `shared_go_report_command_mismatch report=${sharedName}`,
          `shared go report ${sharedName} was created with a different command`,
          `existing: ${existingCommand}`,
          `current:  ${commandText}`,
        ].join("\n"),
      );
    }
    return { reportDir: sharedDir, usage: "reused" };
  }

  const started = captureStart();
  const status = await runGoTestCapture(ctx, regex, args, sharedDir, policy);
  const window = captureFinish(started);
  secureWriteFile(path.join(sharedDir, "command.txt"), `${commandText}\n`);
  secureWriteFile(
    path.join(sharedDir, "start_time.txt"),
    `${window.startTime}\n`,
  );
  secureWriteFile(path.join(sharedDir, "end_time.txt"), `${window.endTime}\n`);
  secureWriteFile(
    path.join(sharedDir, "duration_ms.txt"),
    `${window.durationMs}\n`,
  );
  secureWriteFile(path.join(sharedDir, "exit_status.txt"), `${status}\n`);
  await writeCrossTargetSharedExecutionMetadata(
    ctx,
    sharedDir,
    sharedName,
    window,
    status,
  );
  secureWriteFile(completeFile, "");
  return { reportDir: sharedDir, usage: "actual" };
}

export async function captureGoReport(
  ctx,
  sharedName,
  regex,
  args,
  policy = {},
) {
  const sharedDir = prepareSharedArtifactDir(ctx, sharedName);
  await acquireSharedReportLock(ctx, sharedDir, sharedName);
  try {
    return await captureGoReportLocked(
      ctx,
      sharedDir,
      sharedName,
      regex,
      args,
      policy,
    );
  } finally {
    releaseSharedReportLock(sharedDir);
  }
}

export async function assignExecutionFamily(ctx, target, familyOrShard) {
  const spec = resolveExecutionFamilySpec(ctx, target, familyOrShard);
  const policy = resolveExecutionFamilyPostgresFixturePolicy(
    ctx,
    target,
    familyOrShard,
  );
  const runtimeRows = runtimeRowsForExecution(ctx, target, familyOrShard);
  if (runtimeBinaryIDsForRows(runtimeRows).length > 0) {
    validateRuntimeBinaries(
      ctx,
      runtimeRows,
      prepareSharedArtifactDir(ctx, familyOrShard),
    );
  }
  const captured = await captureGoReport(
    ctx,
    familyOrShard,
    spec.regex,
    spec.args,
    policy,
  );
  if (
    captured.usage === "actual" &&
    isCrossTargetSharedReport(ctx, target, familyOrShard)
  ) {
    return { ...captured, usage: "reused" };
  }
  return captured;
}

function writeShardMetadata(metadataDir, sharedName, captured) {
  secureMkdir(metadataDir);
  const file = path.join(metadataDir, `${sharedName}.meta`);
  secureWriteFile(
    `${file}.${process.pid}`,
    `${captured.reportDir}\n${captured.usage}\n`,
  );
  rmSync(file, { force: true });
  secureWriteFile(file, readFileSync(`${file}.${process.pid}`));
  rmSync(`${file}.${process.pid}`, { force: true });
}

export async function captureScheduledShard(
  ctx,
  target,
  sharedName,
  metadataDir,
) {
  const captured = await assignExecutionFamily(ctx, target, sharedName);
  writeShardMetadata(metadataDir, sharedName, captured);
}

function readSharedReportMetadata(metadataDir, sharedName) {
  const file = path.join(metadataDir, `${sharedName}.meta`);
  if (!existsSync(file)) {
    throw new Error(`missing shared report metadata for ${sharedName}`);
  }
  const metadata = readFileSync(file, "utf8").trimEnd().split(/\r?\n/u);
  if (metadata.length !== 2) {
    throw new Error(`incomplete shared report metadata for ${sharedName}`);
  }
  return { reportDir: metadata[0], usage: metadata[1] };
}

function sharedReportMetadata(metadataDir, sharedName, metadataByShard) {
  if (!metadataByShard) {
    return readSharedReportMetadata(metadataDir, sharedName);
  }
  if (!metadataByShard.has(sharedName)) {
    metadataByShard.set(sharedName, readSharedReportMetadata(metadataDir, sharedName));
  }
  return metadataByShard.get(sharedName);
}

function isoWindowDurationMs(start, end) {
  const duration = Date.parse(end) - Date.parse(start);
  return Number.isFinite(duration) && duration > 0 ? duration : 0;
}

export function createAggregateReport(
  ctx,
  metadataDir,
  aggregateName,
  target,
  shardNames,
  metadataByShard = null,
) {
  const aggregateRoot = path.join(metadataDir, "aggregate-reports", target);
  const outputDir = path.join(aggregateRoot, aggregateName);
  const runnerLog = path.join(outputDir, "runner.jsonl");
  const stderrLog = path.join(outputDir, "stderr.log");
  const commandFile = path.join(outputDir, "command.txt");
  secureMkdir(outputDir);
  secureWriteFile(runnerLog, "");
  secureWriteFile(stderrLog, "");
  secureWriteFile(commandFile, "");

  let startTime = "";
  let endTime = "";
  let durationMs = 0;
  let actualStartTime = "";
  let actualEndTime = "";
  let actualDurationMs = 0;
  let wallDurationMs = 0;
  let exitStatus = 0;
  let hasActual = false;
  let wroteCommand = false;

  for (const shardName of shardNames) {
    const metadata = sharedReportMetadata(metadataDir, shardName, metadataByShard);
    let usage = metadata.usage;
    if (
      usage === "actual" &&
      target === "backend-integration-support" &&
      isCrossTargetSharedReport(ctx, target, shardName)
    ) {
      usage = "reused";
    }
    if (existsSync(path.join(metadata.reportDir, "runner.jsonl"))) {
      appendFileSync(
        runnerLog,
        readFileSync(path.join(metadata.reportDir, "runner.jsonl")),
      );
    }
    if (existsSync(path.join(metadata.reportDir, "stderr.log"))) {
      appendFileSync(
        stderrLog,
        readFileSync(path.join(metadata.reportDir, "stderr.log")),
      );
    }
    if (wroteCommand) {
      appendFileSync(commandFile, "\n");
    }
    appendFileSync(
      commandFile,
      `${shardName}: ${readFileSync(path.join(metadata.reportDir, "command.txt"), "utf8").trimEnd()}\n`,
    );
    wroteCommand = true;

    const shardDuration = clampDurationMs(
      readFileSync(path.join(metadata.reportDir, "duration_ms.txt"), "utf8"),
    );
    durationMs += shardDuration;
    const shardStatus =
      Number.parseInt(
        readFileSync(path.join(metadata.reportDir, "exit_status.txt"), "utf8"),
        10,
      ) || 0;
    if (shardStatus !== 0) {
      exitStatus = shardStatus;
    }
    const shardStart = readFileSync(
      path.join(metadata.reportDir, "start_time.txt"),
      "utf8",
    ).trim();
    const shardEnd = readFileSync(
      path.join(metadata.reportDir, "end_time.txt"),
      "utf8",
    ).trim();
    if (startTime === "" || shardStart < startTime) {
      startTime = shardStart;
    }
    if (endTime === "" || shardEnd > endTime) {
      endTime = shardEnd;
    }
    if (usage === "actual") {
      hasActual = true;
      actualDurationMs += shardDuration;
      if (actualStartTime === "" || shardStart < actualStartTime) {
        actualStartTime = shardStart;
      }
      if (actualEndTime === "" || shardEnd > actualEndTime) {
        actualEndTime = shardEnd;
      }
    }
  }

  const usage = hasActual ? "actual" : "reused";
  if (hasActual) {
    startTime = actualStartTime;
    endTime = actualEndTime;
    durationMs = actualDurationMs;
    wallDurationMs = isoWindowDurationMs(startTime, endTime);
  }
  secureWriteFile(path.join(outputDir, "start_time.txt"), `${startTime}\n`);
  secureWriteFile(path.join(outputDir, "end_time.txt"), `${endTime}\n`);
  secureWriteFile(
    path.join(outputDir, "duration_ms.txt"),
    `${clampDurationMs(durationMs)}\n`,
  );
  secureWriteFile(
    path.join(outputDir, "wall_duration_ms.txt"),
    `${clampDurationMs(wallDurationMs)}\n`,
  );
  secureWriteFile(path.join(outputDir, "exit_status.txt"), `${exitStatus}\n`);
  secureWriteFile(
    path.join(outputDir, "aggregate.txt"),
    `${target}:${aggregateName}\n`,
  );
  return { reportDir: outputDir, usage };
}

function loadPhaseWindow(reportDir, mode) {
  const command = readFileSync(
    path.join(reportDir, "command.txt"),
    "utf8",
  ).trimEnd();
  const exitStatus =
    Number.parseInt(
      readFileSync(path.join(reportDir, "exit_status.txt"), "utf8"),
      10,
    ) || 0;
  const storedDurationMs = clampDurationMs(
    readFileSync(path.join(reportDir, "duration_ms.txt"), "utf8"),
  );
  const storedWallDurationMs = existsSync(
    path.join(reportDir, "wall_duration_ms.txt"),
  )
    ? clampDurationMs(
        readFileSync(path.join(reportDir, "wall_duration_ms.txt"), "utf8"),
      )
    : storedDurationMs;
  if (mode === "actual") {
    return {
      command,
      exitStatus,
      startTime: readFileSync(
        path.join(reportDir, "start_time.txt"),
        "utf8",
      ).trim(),
      endTime: readFileSync(
        path.join(reportDir, "end_time.txt"),
        "utf8",
      ).trim(),
      durationMs: storedDurationMs,
      wallDurationMs: storedWallDurationMs,
    };
  }
  const timestamp = nowUTC();
  return {
    command,
    exitStatus,
    startTime: timestamp,
    endTime: timestamp,
    durationMs: mode === "reused" ? storedDurationMs : 0,
    wallDurationMs: 0,
  };
}

async function emitReportPhaseSummary(
  ctx,
  helperCommand,
  label,
  reportDir,
  mode,
  extraEnv = {},
) {
  const phase = loadPhaseWindow(reportDir, mode);
  const phaseDir = preparePhaseArtifactDir(ctx, label);
  return await runHelper(ctx, [helperCommand], {
    CARTULARY_TEST_TARGET: ctx.testTarget,
    CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
    CARTULARY_PHASE_LABEL: label,
    CARTULARY_PHASE_DIR: phaseDir,
    CARTULARY_PHASE_COMMAND: phase.command,
    CARTULARY_PHASE_START_TIME: phase.startTime,
    CARTULARY_PHASE_END_TIME: phase.endTime,
    CARTULARY_PHASE_DURATION_MS: String(phase.durationMs),
    CARTULARY_PHASE_WALL_DURATION_MS: String(phase.wallDurationMs),
    CARTULARY_PHASE_EXIT_STATUS: String(phase.exitStatus),
    CARTULARY_REPORT_SLICE: "1",
    CARTULARY_PHASE_ACCOUNTING_MODE: mode,
    CARTULARY_PHASE_RUNNER_LOG: path.join(reportDir, "runner.jsonl"),
    CARTULARY_PHASE_STDERR_LOG: path.join(reportDir, "stderr.log"),
    ...extraEnv,
  });
}

function packagePatternsEnv(packages) {
  return packages.join("\n");
}

async function emitGoRawPhase(
  ctx,
  label,
  mode,
  reportDir,
  regex,
  packages,
  coverage,
) {
  return await emitReportPhaseSummary(ctx, "go-phase", label, reportDir, mode, {
    CARTULARY_GO_TEST_REGEX: regex,
    CARTULARY_ACCOUNTING_COVERAGE: coverage,
    CARTULARY_GO_PACKAGE_PATTERNS: packagePatternsEnv(packages),
  });
}

async function emitGoManifestPhase(
  ctx,
  label,
  mode,
  reportDir,
  manifestPhase,
  section,
  coverage,
  executionDependency,
  executionFamily,
  packages,
  selectedIDs = [],
) {
  return await emitReportPhaseSummary(
    ctx,
    "go-manifest-phase",
    label,
    reportDir,
    mode,
    {
      CARTULARY_MANIFEST_PHASE: manifestPhase,
      CARTULARY_MANIFEST_SECTION: section,
      CARTULARY_MANIFEST_COVERAGE: coverage,
      CARTULARY_MANIFEST_EXECUTION_DEPENDENCY: executionDependency,
      CARTULARY_EXECUTION_FAMILY: executionFamily,
      CARTULARY_GO_PACKAGE_PATTERNS: packagePatternsEnv(packages),
      ...(selectedIDs.length > 0
        ? { CARTULARY_MANIFEST_SELECTED_IDS: selectedIDs.join("\n") }
        : {}),
    },
  );
}

export async function emitExecutionFamily(
  ctx,
  target,
  family,
  usage,
  reportDir,
  rows = null,
) {
  let status = 0;
  const emissions = collectAggregateEmissions(
    rows ?? rowsForAggregate(ctx, target, family),
  );
  for (const [index, emission] of emissions.entries()) {
    const emissionUsage = index === 0 ? usage : "derived";
    let result = 0;
    if (emission.mode === "manifest") {
      result = await emitGoManifestPhase(
        ctx,
        emission.label,
        emissionUsage,
        reportDir,
        emission.phase,
        emission.section,
        emission.coverage,
        emission.execution_dependency,
        family,
        emission.packages,
        emission.ids ?? [],
      );
    } else if (emission.mode === "support") {
      result = await emitGoRawPhase(
        ctx,
        emission.label,
        emissionUsage,
        reportDir,
        emission.regex,
        emission.packages,
        "support",
      );
    } else if (emission.mode === "raw") {
      result = await emitGoRawPhase(
        ctx,
        emission.label,
        emissionUsage,
        reportDir,
        emission.regex,
        emission.packages,
        "raw",
      );
    } else {
      throw new Error(
        `unsupported execution family emission mode ${emission.mode}`,
      );
    }
    if (result !== 0) {
      status = result;
    }
  }
  return status;
}

export async function runUnshardedTarget(ctx, target) {
  let status = 0;
  for (const aggregateName of aggregateNames(ctx, target)) {
    try {
      const captured = await assignExecutionFamily(ctx, target, aggregateName);
      if (status === 0) {
        status = await emitExecutionFamily(
          ctx,
          target,
          aggregateName,
          captured.usage,
          captured.reportDir,
        );
      }
    } catch (error) {
      process.stderr.write(`${error.message}\n`);
      status = Number.isInteger(error.exitCode) ? error.exitCode : 1;
    }
  }
  return await finishTarget(ctx, status);
}

export async function finalizeScheduledShards(
  ctx,
  target,
  metadataDir,
  selectedShardNames = [],
) {
  let status = 0;
  const metadataByShard = new Map();
  const aggregateReports = [];
  const selectedShardSet =
    selectedShardNames.length > 0 ? new Set(selectedShardNames) : null;
  const finalizerCommandArgs = [
    "run-go-target.mjs",
    "finalize-shards",
    target,
    metadataDir,
    ...selectedShardNames,
  ];
  if (selectedShardSet) {
    const validationStarted = captureStart();
    const knownShards = new Set(targetShards(ctx, target).map((shard) => shard.name));
    for (const shardName of selectedShardSet) {
      if (!knownShards.has(shardName)) {
        const error = new Error(`unknown scheduled shard ${shardName} for ${target}`);
        const validationWindow = captureFinish(validationStarted);
        writeTargetTimingSpan(
          ctx,
          "report_collation",
          `validate ${target}:selected-shards`,
          validationWindow,
          "fail",
        );
        writeFinalizerFailurePhase(ctx, {
          target,
          label: `validate ${target}:selected-shards`,
          commandArgs: finalizerCommandArgs,
          window: validationWindow,
          error,
          metadataDir,
          shardNames: selectedShardNames,
        });
        process.stderr.write(`${error.message}\n`);
        return await finishTarget(ctx, 1);
      }
    }
  }
  for (const aggregate of targetAggregates(ctx, target)) {
    const shardNames = selectedShardSet
      ? aggregate.shards.filter((shardName) => selectedShardSet.has(shardName))
      : aggregate.shards;
    if (selectedShardSet && shardNames.length === 0) {
      continue;
    }
    const aggregateStarted = captureStart();
    let report = null;
    let rows = null;
    const aggregateRoot = path.join(metadataDir, "aggregate-reports", target);
    const aggregateReportDir = path.join(aggregateRoot, aggregate.name);
    try {
      rows = rowsForScheduledAggregate(ctx, target, aggregate.name, shardNames);
      report = createAggregateReport(
        ctx,
        metadataDir,
        aggregate.name,
        target,
        shardNames,
        metadataByShard,
      );
      const aggregateWindow = captureFinish(aggregateStarted);
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `collate ${target}:${aggregate.name}`,
        aggregateWindow,
        "pass",
      );
    } catch (error) {
      const aggregateWindow = captureFinish(aggregateStarted);
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `collate ${target}:${aggregate.name}`,
        aggregateWindow,
        "fail",
      );
      writeFinalizerFailurePhase(ctx, {
        target,
        label: `collate ${target}:${aggregate.name}`,
        commandArgs: finalizerCommandArgs,
        window: aggregateWindow,
        error,
        metadataDir,
        aggregateReportDir,
        shardNames,
      });
      process.stderr.write(`${error.message}\n`);
      status = 1;
      continue;
    }
    aggregateReports.push({ aggregate, report, rows });
  }
  if (status === 0) {
    const emitStatuses = await runBounded(
      aggregateReports,
      resolveFinalizerEmitJobs(ctx, aggregateReports.length),
      async ({ aggregate, report, rows }) => {
        let emitStatus = 0;
        const emitStarted = captureStart();
        try {
          emitStatus = await emitExecutionFamily(
            ctx,
            target,
            aggregate.name,
            report.usage,
            report.reportDir,
            rows,
          );
        } catch (error) {
          process.stderr.write(`${error.message}\n`);
          emitStatus = 1;
        }
        const emitWindow = captureFinish(emitStarted);
        writeTargetTimingSpan(
          ctx,
          "report_collation",
          `emit ${target}:${aggregate.name}`,
          emitWindow,
          emitStatus === 0 ? "pass" : "fail",
        );
        return emitStatus;
      },
    );
    for (const emitStatus of emitStatuses) {
      if (emitStatus !== 0) {
        status = emitStatus;
        break;
      }
    }
  }
  return await finishTarget(ctx, status);
}

export async function captureNamedSharedReportsParallel(
  ctx,
  target,
  jobs,
  metadataDir,
  shardNames,
) {
  if (!Number.isInteger(jobs) || jobs < 1) {
    throw new Error(`invalid shard job count: ${jobs}`);
  }
  secureMkdir(metadataDir);
  let status = 0;
  let next = 0;
  const workerCount = Math.min(jobs, shardNames.length);
  await Promise.all(
    Array.from({ length: workerCount }, async () => {
      while (next < shardNames.length) {
        const shardName = shardNames[next];
        next += 1;
        try {
          await captureScheduledShard(ctx, target, shardName, metadataDir);
        } catch (error) {
          status = 1;
          process.stderr.write(`${error.message}\n`);
        }
      }
    }),
  );
  return status;
}

export async function runShardedTarget(ctx, target) {
  const metadataDir = mkdtempSync(
    path.join(os.tmpdir(), `cartulary-${target}-shards.`),
  );
  const shardNames = targetShards(ctx, target).map((shard) => shard.name);
  let status = await captureNamedSharedReportsParallel(
    ctx,
    target,
    ctx.backendIntegrationShardJobs,
    metadataDir,
    shardNames,
  );
  if (status === 0) {
    status = await finalizeScheduledShards(ctx, target, metadataDir, shardNames);
  } else {
    status = await finishTarget(ctx, status);
  }
  rmSync(metadataDir, { recursive: true, force: true });
  return status;
}

export function inspectAggregateCommand(ctx, target, familyOrShard) {
  const spec = resolveExecutionFamilySpec(ctx, target, familyOrShard);
  const policy = resolveExecutionFamilyPostgresFixturePolicy(
    ctx,
    target,
    familyOrShard,
  );
  return renderGoTestCommand(ctx, spec.regex, spec.args, policy);
}

function usage() {
  return [
    "usage: run-go-target.mjs <backend-unit|backend-store|backend-integration|backend-integration-support|backend-process>",
    "       run-go-target.mjs inspect-aggregate-command <target> <execution-family-or-shard>",
    "       run-go-target.mjs capture-shard <target> <shard-name> <metadata-dir>",
    "       run-go-target.mjs finalize-shards <target> <metadata-dir> [shard-name...]",
  ].join("\n");
}

export async function runGoTargetCLI(argv, options = {}) {
  if (argv.length === 0) {
    process.stderr.write(`${usage()}\n`);
    return 2;
  }
  const ctx = createGoTargetContext(options);
  const [command, ...rest] = argv;
  try {
    switch (command) {
      case "inspect-aggregate-command":
        if (rest.length !== 2) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        process.stdout.write(
          `${inspectAggregateCommand(ctx, rest[0], rest[1])}\n`,
        );
        return 0;
      case "capture-shard":
        if (rest.length !== 3) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        await captureScheduledShard(ctx, rest[0], rest[1], rest[2]);
        return 0;
      case "finalize-shards":
        if (rest.length < 2) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await finalizeScheduledShards(ctx, rest[0], rest[1], rest.slice(2));
      case "backend-unit":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runUnshardedTarget(ctx, "backend-unit");
      case "backend-store":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runShardedTarget(ctx, "backend-store");
      case "backend-integration":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runShardedTarget(ctx, "backend-integration");
      case "backend-integration-support":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runShardedTarget(ctx, "backend-integration-support");
      case "backend-process":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runShardedTarget(ctx, "backend-process");
      default:
        process.stderr.write(`${usage()}\n`);
        return 2;
    }
  } catch (error) {
    process.stderr.write(
      `${error instanceof Error ? error.message : String(error)}\n`,
    );
    return process.exitCode && process.exitCode !== 0 ? process.exitCode : 1;
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const status = await runGoTargetCLI(process.argv.slice(2));
  process.exit(status);
}
