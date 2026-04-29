import { spawn, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import {
  appendFileSync,
  createWriteStream,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { collectGoShardPlan } from "./go-shard-plan.mjs";
import { collectTargetPlanRows, findTargetDescriptor } from "./target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "..", "..");
const postgresFixturePolicyEnvAssignable = new Set([
  "template_clone",
  "package_reset",
  "transaction",
  "group_clone",
]);

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
  return path.isAbsolute(configured) ? configured : path.join(repoRoot, configured);
}

function resolveRunID(env) {
  return env.CARTULARY_TEST_RUN_ID || `${new Date().toISOString().replace(/[-:]/g, "").replace(/\..+/, "Z")}-p${process.pid}`;
}

function resolveOutputMode(env) {
  if (env.VERBOSE === "1" || env.CI_VERBOSE === "1") {
    return "normal";
  }
  return env.CARTULARY_OUTPUT_MODE || "quiet";
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
  const runId = resolveRunID(baseEnv);
  const resultsRoot = resolveResultsRoot(repoRoot, baseEnv);
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
      env.GO_TEST_PACKAGE_PARALLELISM || env.GO_TEST_SERVICE_PACKAGE_PARALLELISM || "1",
    backendIntegrationShardJobs: Number.parseInt(env.BACKEND_INTEGRATION_SHARD_JOBS || "4", 10) || 4,
    resultsRoot,
    runId,
    testTarget: env.CARTULARY_TEST_TARGET || "",
    outputMode: resolveOutputMode(env),
    testOutputScript: resolvePath(
      repoRoot,
      env.TEST_OUTPUT_SCRIPT || path.join("scripts", "lib", "test-output.mjs"),
    ),
    invocation: null,
    targetPlanRows: null,
    shardPlan: null,
  };
}

function targetDir(ctx) {
  const target = ctx.testTarget || "adhoc";
  const dir = path.join(ctx.resultsRoot, ctx.runId, target);
  mkdirSync(dir, { recursive: true });
  return dir;
}

function preparePhaseArtifactDir(ctx, label) {
  const slug = slugifyLabel(label) || "phase";
  const dir = path.join(targetDir(ctx), slug);
  mkdirSync(dir, { recursive: true });
  return dir;
}

export function prepareSharedArtifactDir(ctx, name) {
  if (!name) {
    throw new Error("prepareSharedArtifactDir requires name");
  }
  const dir = path.join(ctx.resultsRoot, ctx.runId, "_shared", name);
  mkdirSync(dir, { recursive: true });
  return dir;
}

function targetRows(ctx) {
  if (!ctx.targetPlanRows) {
    ctx.targetPlanRows = collectTargetPlanRows(ctx.repoRoot);
  }
  return ctx.targetPlanRows;
}

function shardPlan(ctx) {
  if (!ctx.shardPlan) {
    ctx.shardPlan = collectGoShardPlan(ctx.repoRoot);
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

function escapeRegex(value) {
  return String(value).replace(/[.*+?^${}()|[\]\\]/gu, String.raw`\$&`);
}

function exactRegex(values) {
  const escaped = values.map(escapeRegex);
  if (escaped.length === 0) {
    throw new Error("cannot build an exact regex from an empty value list");
  }
  if (escaped.length === 1) {
    return `^${escaped[0]}$`;
  }
  return `^(${escaped.join("|")})$`;
}

function buildUnionRegex(components) {
  const values = components.filter((component) => component !== "");
  if (values.length === 0) {
    throw new Error("cannot build aggregate regex from an empty selection");
  }
  if (values.length === 1) {
    return values[0];
  }
  return values.map((component) => `(${component})`).join("|");
}

function aggregateRegex(rows) {
  const symbols = rows.flatMap((row) => row.symbols ?? []);
  const components = [];
  if (symbols.length > 0) {
    components.push(exactRegex(symbols.sort(compareStrings)));
  }
  for (const row of rows) {
    if (row.raw_selector) {
      components.push(row.raw_selector);
    }
  }
  return buildUnionRegex(components);
}

function aggregatePackages(rows) {
  return Array.from(new Set(rows.flatMap(rowPackages))).sort(compareStrings);
}

function fixturePolicyAssignments(rows, mode) {
  const assignments = [];
  for (const row of rows) {
    const policy = row.fixture_policy?.postgres ?? "";
    if (!postgresFixturePolicyEnvAssignable.has(policy)) {
      continue;
    }
    if (mode === "tests" && row.coverage !== "raw") {
      for (const symbol of row.symbols ?? []) {
        assignments.push(`${symbol}=${policy}`);
      }
    }
    if (mode === "packages" && row.coverage === "raw") {
      for (const pkg of row.packages ?? []) {
        assignments.push(`${pkg}=${policy}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

function resetTableAssignments(rows, mode) {
  const assignments = [];
  for (const row of rows) {
    const dirtyTables = row.fixture_budget?.postgres?.dirty_tables ?? [];
    if (dirtyTables.length === 0) {
      continue;
    }
    if (mode === "tests" && row.coverage !== "raw") {
      for (const symbol of row.symbols ?? []) {
        assignments.push(`${symbol}=${dirtyTables.join("|")}`);
      }
    }
    if (mode === "packages" && row.coverage === "raw") {
      for (const pkg of row.packages ?? []) {
        assignments.push(`${pkg}=${dirtyTables.join("|")}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

function aggregateKey(row) {
  if (row.coverage === "raw") {
    return `raw:${row.id}`;
  }
  if (row.support_only) {
    return [
      "support",
      row.manifest_phase,
      row.execution_dependency,
      row.execution_family,
      row.execution_label,
    ].join("\u001f");
  }
  return [
    "manifest",
    row.manifest_phase,
    row.section,
    row.coverage,
    row.execution_dependency,
    row.execution_family,
    row.execution_label,
  ].join("\u001f");
}

function collectAggregateEmissions(rows) {
  const groups = new Map();
  for (const row of rows) {
    const key = aggregateKey(row);
    if (!groups.has(key)) {
      groups.set(key, {
        mode: row.coverage === "raw" ? "raw" : row.support_only ? "support" : "manifest",
        label: row.execution_label,
        phase: row.manifest_phase,
        section: row.section,
        coverage: row.coverage,
        execution_dependency: row.execution_dependency,
        execution_family: row.execution_family,
        support_target: row.support_only ? row.execution_dependency : "",
        regex: row.raw_selector ?? "",
        packages: new Set(),
        symbols: [],
      });
    }
    const group = groups.get(key);
    for (const pkg of rowPackages(row)) {
      group.packages.add(pkg);
    }
    if (row.support_only) {
      group.symbols.push(...(row.symbols ?? []));
    }
  }
  return Array.from(groups.values()).map((group) => {
    const symbols = group.symbols.sort(compareStrings);
    return {
      ...group,
      regex: group.mode === "support" ? exactRegex(symbols) : group.regex,
      packages: Array.from(group.packages).sort(compareStrings),
      symbols,
    };
  });
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
    return shard.items.some((item) => item.kind === "authoritative" || item.kind === "raw");
  }
  if (target === "backend-integration-support") {
    return shard.items.some((item) => item.kind === "support");
  }
  if (target === "backend-store") {
    return shard.items.some((item) => item.kind === "authoritative");
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
    (shard) => aggregateSet.has(shard.aggregate_name) && targetOwnsShard(target, shard),
  );
}

function targetAggregates(ctx, target) {
  return shardPlan(ctx).aggregates.filter((aggregate) => aggregate.target === target);
}

function findShard(ctx, target, name) {
  const shard = targetShards(ctx, target).find((candidate) => candidate.name === name);
  if (!shard) {
    throw new Error(`unknown shard ${name} for ${target}`);
  }
  return shard;
}

function findPlannedAggregate(ctx, target, name) {
  const aggregate = targetAggregates(ctx, target).find((candidate) => candidate.name === name);
  if (!aggregate) {
    throw new Error(`unknown aggregate ${name} for ${target}`);
  }
  return aggregate;
}

function fixturePolicyAssignmentsForShard(shard, mode) {
  const assignments = [];
  for (const item of shard.items) {
    if (!item.postgres_fixture_policy || item.postgres_fixture_policy === "migration_scratch") {
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
      throw new Error(`unsupported go_test_parallelism ${descriptor.goTestParallelism} for ${target}`);
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
  if (familyOrShard.includes("-shard-")) {
    return shardSpec(ctx, target, familyOrShard);
  }
  return aggregateSpec(ctx, target, familyOrShard);
}

function resolveExecutionFamilyPostgresFixturePolicy(ctx, target, familyOrShard) {
  if (familyOrShard.includes("-shard-")) {
    const shard = findShard(ctx, target, familyOrShard);
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
      mkdirSync(lockDir, { recursive: false });
      for (const [name, value] of Object.entries(metadata)) {
        writeFileSync(path.join(lockDir, name), `${value}\n`);
      }
      return;
    } catch (error) {
      if (error?.code !== "EEXIST") {
        throw error;
      }
    }

    const ownerPid = Number.parseInt(
      existsSync(path.join(lockDir, "pid")) ? readFileSync(path.join(lockDir, "pid"), "utf8") : "",
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
  const timeoutSeconds = Number.parseInt(ctx.env.CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS || "300", 10);
  if (!Number.isInteger(timeoutSeconds) || timeoutSeconds < 1) {
    throw new Error(`invalid CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS=${ctx.env.CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS}`);
  }
  mkdirSync(warmRoot, { recursive: true });
  const warmKey = hashGoTestDependencyInputs(ctx);
  const stampFile = path.join(warmRoot, `${warmKey}.stamp`);
  if (existsSync(stampFile)) {
    return;
  }
  await acquireDirectoryLock(lockDir, "go_test_dependency_warm_lock", timeoutSeconds, {
    pid: process.pid,
    acquired_at: nowUTC(),
  });
  try {
    if (existsSync(stampFile)) {
      return;
    }
    for (const args of [["mod", "download"], ["list", "-deps", "-test", "./..."]]) {
      const result = spawnSync(ctx.goBin, args, {
        cwd: ctx.repoRoot,
        env: goChildEnv(ctx),
        stdio: "inherit",
      });
      if ((result.status ?? 1) !== 0) {
        process.exitCode = result.status ?? 1;
        throw new Error(`${ctx.goBin} ${args.join(" ")} exited ${result.status ?? 1}`);
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
  const command = ctx.testOutputScript.endsWith(".mjs") ? ctx.nodeBin : ctx.testOutputScript;
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

async function emitTargetTimingSpan(ctx, bucket, label, window, status, exitStatus) {
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
  const stdoutStream = createWriteStream(runnerLog, { flags: "w" });
  const stderrStream = createWriteStream(stderrLog, { flags: "w" });
  let stdoutBuffer = "";

  const child = spawn(ctx.goBin, ["test", "-json", "-run", regex, ...args], {
    cwd: ctx.repoRoot,
    env: goChildEnv(ctx, policy),
    stdio: ["ignore", "pipe", "pipe"],
  });
  child.stdout.on("data", (chunk) => {
    stdoutStream.write(chunk);
    if (ctx.outputMode === "quiet") {
      return;
    }
    stdoutBuffer += chunk.toString("utf8");
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
    stderrStream.write(chunk);
    if (ctx.outputMode !== "quiet") {
      process.stderr.write(chunk);
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
  const timeoutSeconds = Number.parseInt(ctx.env.CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS || "300", 10);
  if (!Number.isInteger(timeoutSeconds) || timeoutSeconds < 1) {
    throw new Error(`invalid CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS=${ctx.env.CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS}`);
  }
  try {
    await acquireDirectoryLock(path.join(sharedDir, "capture.lock"), "shared_go_report_lock", timeoutSeconds, {
      pid: process.pid,
      shared_report: sharedName,
      acquired_at: nowUTC(),
    });
  } catch (error) {
    if (String(error.message).startsWith("shared_go_report_lock_timeout")) {
      process.stderr.write(`shared_go_report_lock_timeout report=${sharedName} lock=${path.join(sharedDir, "capture.lock")}\n`);
    }
    throw error;
  }
}

export function releaseSharedReportLock(sharedDir) {
  rmSync(path.join(sharedDir, "capture.lock"), { recursive: true, force: true });
}

function isCrossTargetSharedReport(ctx, target, sharedName) {
  if (!sharedName.includes("-shard-")) {
    return false;
  }
  try {
    return findShard(ctx, target, sharedName).shared_across_targets === true;
  } catch {
    return false;
  }
}

async function writeCrossTargetSharedExecutionMetadata(ctx, sharedDir, sharedName, window, status) {
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

export async function captureGoReportLocked(ctx, sharedDir, sharedName, regex, args, policy = {}) {
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
  writeFileSync(path.join(sharedDir, "command.txt"), `${commandText}\n`);
  writeFileSync(path.join(sharedDir, "start_time.txt"), `${window.startTime}\n`);
  writeFileSync(path.join(sharedDir, "end_time.txt"), `${window.endTime}\n`);
  writeFileSync(path.join(sharedDir, "duration_ms.txt"), `${window.durationMs}\n`);
  writeFileSync(path.join(sharedDir, "exit_status.txt"), `${status}\n`);
  await writeCrossTargetSharedExecutionMetadata(ctx, sharedDir, sharedName, window, status);
  writeFileSync(completeFile, "");
  return { reportDir: sharedDir, usage: "actual" };
}

export async function captureGoReport(ctx, sharedName, regex, args, policy = {}) {
  const sharedDir = prepareSharedArtifactDir(ctx, sharedName);
  await acquireSharedReportLock(ctx, sharedDir, sharedName);
  try {
    return await captureGoReportLocked(ctx, sharedDir, sharedName, regex, args, policy);
  } finally {
    releaseSharedReportLock(sharedDir);
  }
}

export async function assignExecutionFamily(ctx, target, familyOrShard) {
  const spec = resolveExecutionFamilySpec(ctx, target, familyOrShard);
  const policy = resolveExecutionFamilyPostgresFixturePolicy(ctx, target, familyOrShard);
  const captured = await captureGoReport(ctx, familyOrShard, spec.regex, spec.args, policy);
  if (captured.usage === "actual" && isCrossTargetSharedReport(ctx, target, familyOrShard)) {
    return { ...captured, usage: "reused" };
  }
  return captured;
}

function writeShardMetadata(metadataDir, sharedName, captured) {
  mkdirSync(metadataDir, { recursive: true });
  const file = path.join(metadataDir, `${sharedName}.meta`);
  writeFileSync(`${file}.${process.pid}`, `${captured.reportDir}\n${captured.usage}\n`);
  rmSync(file, { force: true });
  writeFileSync(file, readFileSync(`${file}.${process.pid}`));
  rmSync(`${file}.${process.pid}`, { force: true });
}

export async function captureScheduledShard(ctx, target, sharedName, metadataDir) {
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

function isoWindowDurationMs(start, end) {
  const duration = Date.parse(end) - Date.parse(start);
  return Number.isFinite(duration) && duration > 0 ? duration : 0;
}

export function createAggregateReport(ctx, metadataDir, aggregateName, target, shardNames) {
  const aggregateRoot = path.join(metadataDir, "aggregate-reports", target);
  const outputDir = path.join(aggregateRoot, aggregateName);
  const runnerLog = path.join(outputDir, "runner.jsonl");
  const stderrLog = path.join(outputDir, "stderr.log");
  const commandFile = path.join(outputDir, "command.txt");
  mkdirSync(outputDir, { recursive: true });
  writeFileSync(runnerLog, "");
  writeFileSync(stderrLog, "");
  writeFileSync(commandFile, "");

  let startTime = "";
  let endTime = "";
  let durationMs = 0;
  let actualStartTime = "";
  let actualEndTime = "";
  let actualDurationMs = 0;
  let wallDurationMs = 0;
  let exitStatus = 0;
  let hasActual = false;

  for (const shardName of shardNames) {
    const metadata = readSharedReportMetadata(metadataDir, shardName);
    let usage = metadata.usage;
    if (usage === "actual" && target === "backend-integration-support" && isCrossTargetSharedReport(ctx, target, shardName)) {
      usage = "reused";
    }
    if (existsSync(path.join(metadata.reportDir, "runner.jsonl"))) {
      appendFileSync(runnerLog, readFileSync(path.join(metadata.reportDir, "runner.jsonl")));
    }
    if (existsSync(path.join(metadata.reportDir, "stderr.log"))) {
      appendFileSync(stderrLog, readFileSync(path.join(metadata.reportDir, "stderr.log")));
    }
    if (readFileSync(commandFile, "utf8").length > 0) {
      appendFileSync(commandFile, "\n");
    }
    appendFileSync(commandFile, `${shardName}: ${readFileSync(path.join(metadata.reportDir, "command.txt"), "utf8").trimEnd()}\n`);

    const shardDuration = clampDurationMs(readFileSync(path.join(metadata.reportDir, "duration_ms.txt"), "utf8"));
    durationMs += shardDuration;
    const shardStatus = Number.parseInt(readFileSync(path.join(metadata.reportDir, "exit_status.txt"), "utf8"), 10) || 0;
    if (shardStatus !== 0) {
      exitStatus = shardStatus;
    }
    const shardStart = readFileSync(path.join(metadata.reportDir, "start_time.txt"), "utf8").trim();
    const shardEnd = readFileSync(path.join(metadata.reportDir, "end_time.txt"), "utf8").trim();
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
  writeFileSync(path.join(outputDir, "start_time.txt"), `${startTime}\n`);
  writeFileSync(path.join(outputDir, "end_time.txt"), `${endTime}\n`);
  writeFileSync(path.join(outputDir, "duration_ms.txt"), `${clampDurationMs(durationMs)}\n`);
  writeFileSync(path.join(outputDir, "wall_duration_ms.txt"), `${clampDurationMs(wallDurationMs)}\n`);
  writeFileSync(path.join(outputDir, "exit_status.txt"), `${exitStatus}\n`);
  writeFileSync(path.join(outputDir, "aggregate.txt"), `${target}:${aggregateName}\n`);
  return { reportDir: outputDir, usage };
}

function loadPhaseWindow(reportDir, mode) {
  const command = readFileSync(path.join(reportDir, "command.txt"), "utf8").trimEnd();
  const exitStatus = Number.parseInt(readFileSync(path.join(reportDir, "exit_status.txt"), "utf8"), 10) || 0;
  const storedDurationMs = clampDurationMs(readFileSync(path.join(reportDir, "duration_ms.txt"), "utf8"));
  const storedWallDurationMs = existsSync(path.join(reportDir, "wall_duration_ms.txt"))
    ? clampDurationMs(readFileSync(path.join(reportDir, "wall_duration_ms.txt"), "utf8"))
    : storedDurationMs;
  if (mode === "actual") {
    return {
      command,
      exitStatus,
      startTime: readFileSync(path.join(reportDir, "start_time.txt"), "utf8").trim(),
      endTime: readFileSync(path.join(reportDir, "end_time.txt"), "utf8").trim(),
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

async function emitReportPhaseSummary(ctx, helperCommand, label, reportDir, mode, extraEnv = {}) {
  const phase = loadPhaseWindow(reportDir, mode);
  const phaseDir = preparePhaseArtifactDir(ctx, label);
  return await runHelper(ctx, [helperCommand], {
    CARTULARY_TEST_TARGET: ctx.testTarget,
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

async function emitGoRawPhase(ctx, label, mode, reportDir, regex, packages, coverage) {
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
) {
  return await emitReportPhaseSummary(ctx, "go-manifest-phase", label, reportDir, mode, {
    CARTULARY_MANIFEST_PHASE: manifestPhase,
    CARTULARY_MANIFEST_SECTION: section,
    CARTULARY_MANIFEST_COVERAGE: coverage,
    CARTULARY_MANIFEST_EXECUTION_DEPENDENCY: executionDependency,
    CARTULARY_EXECUTION_FAMILY: executionFamily,
    CARTULARY_GO_PACKAGE_PATTERNS: packagePatternsEnv(packages),
  });
}

export async function emitExecutionFamily(ctx, target, family, usage, reportDir) {
  let status = 0;
  const emissions = collectAggregateEmissions(rowsForAggregate(ctx, target, family));
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
      );
    } else if (emission.mode === "support") {
      result = await emitGoRawPhase(ctx, emission.label, emissionUsage, reportDir, emission.regex, emission.packages, "support");
    } else if (emission.mode === "raw") {
      result = await emitGoRawPhase(ctx, emission.label, emissionUsage, reportDir, emission.regex, emission.packages, "raw");
    } else {
      throw new Error(`unsupported execution family emission mode ${emission.mode}`);
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
        status = await emitExecutionFamily(ctx, target, aggregateName, captured.usage, captured.reportDir);
      }
    } catch (error) {
      process.stderr.write(`${error.message}\n`);
      status = 1;
    }
  }
  return await finishTarget(ctx, status);
}

export async function finalizeScheduledShards(ctx, target, metadataDir) {
  let status = 0;
  for (const aggregate of targetAggregates(ctx, target)) {
    try {
      const report = createAggregateReport(ctx, metadataDir, aggregate.name, target, aggregate.shards);
      if (status === 0) {
        status = await emitExecutionFamily(ctx, target, aggregate.name, report.usage, report.reportDir);
      }
    } catch (error) {
      process.stderr.write(`${error.message}\n`);
      status = 1;
    }
  }
  return await finishTarget(ctx, status);
}

export async function captureNamedSharedReportsParallel(ctx, target, jobs, metadataDir, shardNames) {
  if (!Number.isInteger(jobs) || jobs < 1) {
    throw new Error(`invalid shard job count: ${jobs}`);
  }
  mkdirSync(metadataDir, { recursive: true });
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
  const metadataDir = mkdtempSync(path.join(os.tmpdir(), `cartulary-${target}-shards.`));
  const shardNames = targetShards(ctx, target).map((shard) => shard.name);
  let status = await captureNamedSharedReportsParallel(
    ctx,
    target,
    ctx.backendIntegrationShardJobs,
    metadataDir,
    shardNames,
  );
  if (status === 0) {
    status = await finalizeScheduledShards(ctx, target, metadataDir);
  } else {
    status = await finishTarget(ctx, status);
  }
  rmSync(metadataDir, { recursive: true, force: true });
  return status;
}

export function inspectAggregateCommand(ctx, target, familyOrShard) {
  const spec = resolveExecutionFamilySpec(ctx, target, familyOrShard);
  const policy = resolveExecutionFamilyPostgresFixturePolicy(ctx, target, familyOrShard);
  return renderGoTestCommand(ctx, spec.regex, spec.args, policy);
}

function usage() {
  return [
    "usage: run-go-target.mjs <backend-unit|backend-store|backend-integration|backend-integration-support|backend-process>",
    "       run-go-target.mjs inspect-aggregate-command <target> <execution-family-or-shard>",
    "       run-go-target.mjs capture-shard <target> <shard-name> <metadata-dir>",
    "       run-go-target.mjs finalize-shards <target> <metadata-dir>",
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
        process.stdout.write(`${inspectAggregateCommand(ctx, rest[0], rest[1])}\n`);
        return 0;
      case "capture-shard":
        if (rest.length !== 3) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        await captureScheduledShard(ctx, rest[0], rest[1], rest[2]);
        return 0;
      case "finalize-shards":
        if (rest.length !== 2) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await finalizeScheduledShards(ctx, rest[0], rest[1]);
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
        return await runUnshardedTarget(ctx, "backend-process");
      default:
        process.stderr.write(`${usage()}\n`);
        return 2;
    }
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    return process.exitCode && process.exitCode !== 0 ? process.exitCode : 1;
  }
}
