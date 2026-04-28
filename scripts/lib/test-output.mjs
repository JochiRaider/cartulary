#!/usr/bin/env node

import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { collectEntries, loadManifest } from "./phase-manifest.mjs";
import { collectTargetPlanRows, findTargetDescriptor } from "./target-plan.mjs";
import {
  combineFixtureSummaries,
  emptyFixtureSummary,
  fixtureSummaryLine,
  normalizeFixtureSummary,
  summarizeFixtureActivities,
} from "./fixture-reporting.mjs";
import {
  defaultTaskSurfaceManifestPath,
  loadTaskSurfaceManifest,
  projectionChildren,
} from "./task-surface.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..");
const resultsRoot = resolveResultsRoot();
const runId = process.env.CARTULARY_TEST_RUN_ID || "adhoc";
const phaseSummarySchemaID = "cartulary.test_phase_summary.v2";
const targetTimingSchemaID = "cartulary.test_target_timing.v1";
const targetSummarySchemaID = "cartulary.test_target_summary.v3";
const runSummarySchemaID = "cartulary.test_run_summary.v3";
const sharedExecutionGroupSchemaID = "cartulary.test_shared_execution_group.v1";
const timingBucketOrder = [
  "setup",
  "service_wait",
  "migration",
  "server_startup",
  "frontend_startup",
  "test_command",
  "teardown",
  "report_collation",
];
const timingBucketSet = new Set(timingBucketOrder);

let cachedGoModulePath;
let cachedManifestIndex;

function main() {
  const [command, ...rest] = process.argv.slice(2);
  switch (command) {
    case "shell-phase":
      process.exit(handleShellPhase());
      break;
    case "go-phase":
      process.exit(handleGoPhase({ manifestAware: false }));
      break;
    case "go-manifest-phase":
      process.exit(handleGoPhase({ manifestAware: true }));
      break;
    case "vitest-phase":
      process.exit(handleVitestPhase({ manifestAware: false }));
      break;
    case "vitest-manifest-phase":
      process.exit(handleVitestPhase({ manifestAware: true }));
      break;
    case "playwright-phase":
      process.exit(handlePlaywrightPhase({ manifestAware: false }));
      break;
    case "playwright-manifest-phase":
      process.exit(handlePlaywrightPhase({ manifestAware: true }));
      break;
    case "go-json-stream":
      process.exit(handleGoJSONStream());
      break;
    case "target-summary":
      process.exit(handleTargetSummary(rest));
      break;
    case "timing-span":
      process.exit(handleTimingSpan());
      break;
    case "shared-execution":
      process.exit(handleSharedExecution(rest));
      break;
    case "run-summary":
      process.exit(handleRunSummary(rest));
      break;
    case "run-start":
      process.exit(handleRunStart(rest));
      break;
    case "step-start":
      process.exit(handleStepStart(rest));
      break;
    case "target-start":
      process.exit(handleTargetStart(rest));
      break;
    default:
      throw new Error(`unknown test-output command ${command}`);
  }
}

function resolveResultsRoot() {
  const configured = process.env.CARTULARY_TEST_RESULTS_DIR;
  if (!configured) {
    return path.join(repoRoot, ".cartulary", "test-results");
  }
  return path.isAbsolute(configured) ? configured : path.join(repoRoot, configured);
}

function resolveGoModulePath() {
  if (cachedGoModulePath !== undefined) {
    return cachedGoModulePath;
  }
  const goMod = readFileSync(path.join(repoRoot, "go.mod"), "utf8");
  const match = goMod.match(/^module\s+(\S+)$/m);
  if (!match) {
    throw new Error("unable to determine Go module path from go.mod");
  }
  cachedGoModulePath = match[1];
  return cachedGoModulePath;
}

function toGoImportPath(repoRelativePackage) {
  if (!repoRelativePackage.startsWith("./")) {
    return repoRelativePackage;
  }
  const suffix = repoRelativePackage.slice(2);
  if (suffix === "") {
    return resolveGoModulePath();
  }
  return `${resolveGoModulePath()}/${suffix}`;
}

function toRepoRelativePackage(importPath) {
  const modulePath = resolveGoModulePath();
  if (importPath === modulePath) {
    return "./";
  }
  if (importPath.startsWith(`${modulePath}/`)) {
    return `./${importPath.slice(modulePath.length + 1)}`;
  }
  return importPath;
}

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function relToRepo(value) {
  if (!value) {
    return "";
  }
  const normalized = normalizePath(value);
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(repoRoot, value));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

function ensureDir(dir) {
  mkdirSync(dir, { recursive: true });
}

function writeJson(file, value) {
  ensureDir(path.dirname(file));
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function createCounts() {
  return {
    tests: 0,
    failed: 0,
    authoritative: 0,
    support: 0,
    unmapped: 0,
    non_test: 0,
    authoritative_failed: 0,
    support_failed: 0,
    unmapped_failed: 0,
    non_test_failed: 0,
    packages: 0,
  };
}

function clampDurationMs(value) {
  if (!Number.isFinite(value) || value < 0) {
    return 0;
  }
  return value;
}

function formatDuration(durationMs) {
  if (!Number.isFinite(durationMs) || durationMs < 0) {
    return "0ms";
  }
  if (durationMs < 1000) {
    return `${Math.round(durationMs)}ms`;
  }
  const seconds = durationMs / 1000;
  if (seconds < 60) {
    return `${seconds.toFixed(seconds >= 10 ? 1 : 2)}s`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainder = seconds - minutes * 60;
  return `${minutes}m${remainder.toFixed(1)}s`;
}

function normalizeAccountingMode(value) {
  if (value === "actual" || value === "reused" || value === "derived") {
    return value;
  }
  return "actual";
}

function createAccountingModes() {
  return {
    actual: 0,
    reused: 0,
    derived: 0,
  };
}

function mergeAccountingModes(target, source) {
  for (const mode of Object.keys(target)) {
    target[mode] += clampDurationMs(source?.[mode] ?? 0);
  }
}

function resolveAccountingModes(accountingModes, fallbackActualPhases = 0) {
  const modes = createAccountingModes();
  if (!accountingModes) {
    modes.actual = clampDurationMs(fallbackActualPhases);
    return modes;
  }
  for (const mode of Object.keys(modes)) {
    modes[mode] = clampDurationMs(accountingModes[mode] ?? 0);
  }
  return modes;
}

function formatAccountingModeFields(accountingModes) {
  const modes = resolveAccountingModes(accountingModes);
  return `actual=${modes.actual ?? 0} reused=${modes.reused ?? 0} derived=${modes.derived ?? 0}`;
}

function formatDurationFields(
  wallDurationMs,
  executedDurationMs,
  logicalDurationMs = executedDurationMs,
  criticalPathWallDurationMs = wallDurationMs,
  teardownDurationMs = 0,
) {
  const effectiveLogical = clampDurationMs(logicalDurationMs);
  const effectiveExecuted = clampDurationMs(executedDurationMs);
  const effectiveWall = Number.isFinite(wallDurationMs) ? wallDurationMs : effectiveLogical;
  const effectiveCriticalPath = Number.isFinite(criticalPathWallDurationMs)
    ? criticalPathWallDurationMs
    : effectiveWall;
  const effectiveTeardown = clampDurationMs(teardownDurationMs);
  return `wall=${formatDuration(effectiveWall)} critical=${formatDuration(effectiveCriticalPath)} exec=${formatDuration(effectiveExecuted)} logical=${formatDuration(effectiveLogical)} teardown=${formatDuration(effectiveTeardown)}`;
}

function resolveOutputMode() {
  if (process.env.VERBOSE === "1" || process.env.CI_VERBOSE === "1") {
    return "normal";
  }
  return process.env.CARTULARY_OUTPUT_MODE || "quiet";
}

function quietOutputMode() {
  return resolveOutputMode() === "quiet";
}

function parseLifecycleOptions(args) {
  const options = { positional: [], force: false };
  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];
    if (arg === "--force") {
      options.force = true;
      continue;
    }
    if (arg.startsWith("--")) {
      const name = arg.slice(2).replaceAll("-", "_");
      const value = args[index + 1];
      if (value === undefined) {
        throw new Error(`${arg} requires a value`);
      }
      options[name] = value;
      index += 1;
      continue;
    }
    options.positional.push(arg);
  }
  return options;
}

function shouldEmitLifecycle(options) {
  return options.force || resolveOutputMode() === "quiet";
}

function parseNonNegativeInteger(value, label) {
  const parsed = Number.parseInt(value ?? "", 10);
  if (!Number.isFinite(parsed) || parsed < 0) {
    throw new Error(`${label} must be a non-negative integer`);
  }
  return parsed;
}

function parseTargetListValue(value) {
  if (!value) {
    return [];
  }
  return value
    .split(",")
    .map((entry) => entry.trim())
    .filter((entry) => entry.length > 0);
}

function handleRunStart(args) {
  const options = parseLifecycleOptions(args);
  const [label] = options.positional;
  if (!label || options.steps === undefined || options.targets === undefined || options.jobs === undefined) {
    throw new Error("usage: test-output.mjs run-start <label> --steps <n> --targets <n> --jobs <n> [--force]");
  }
  if (!shouldEmitLifecycle(options)) {
    return 0;
  }
  const steps = parseNonNegativeInteger(options.steps, "steps");
  const targets = parseNonNegativeInteger(options.targets, "targets");
  const jobs = parseNonNegativeInteger(options.jobs, "jobs");
  process.stdout.write(`[RUN] ${label} steps=${steps} targets=${targets} jobs=${jobs} run_id=${runId}\n`);
  return 0;
}

function handleStepStart(args) {
  const options = parseLifecycleOptions(args);
  const [label, indexText, totalText, target] = options.positional;
  const mode = options.mode ?? options.positional[4] ?? "serial";
  const jobsText = options.jobs ?? options.positional[5] ?? "1";
  if (!label || !indexText || !totalText || !target) {
    throw new Error("usage: test-output.mjs step-start <label> <index> <total> <target> [--mode <mode>] [--jobs <n>] [--force]");
  }
  if (!shouldEmitLifecycle(options)) {
    return 0;
  }
  const index = parseNonNegativeInteger(indexText, "index");
  const total = parseNonNegativeInteger(totalText, "total");
  const jobs = parseNonNegativeInteger(jobsText, "jobs");
  process.stdout.write(`[STEP] ${label} ${index}/${total} ${target} mode=${mode} jobs=${jobs}\n`);
  return 0;
}

function targetStartStats(target, children) {
  const childSet = new Set(children);
  const rows = collectTargetPlanRows(repoRoot).filter((row) => {
    if (childSet.size > 0) {
      return childSet.has(row.target);
    }
    return row.target === target;
  });
  const descriptor = findTargetDescriptor(target);
  const serviceBacked = descriptor?.serviceBacked ?? rows.some((row) => row.service_backed);
  const manifestPhases = new Set(rows.map((row) => row.manifest_phase).filter(Boolean));
  const rawRows = rows.filter((row) => row.manifest_phase === "").length;
  return {
    serviceBacked,
    expectedPhases: manifestPhases.size + rawRows,
    expectedTests: rows.length,
  };
}

function handleTargetStart(args) {
  const options = parseLifecycleOptions(args);
  const [target] = options.positional;
  if (!target) {
    throw new Error("usage: test-output.mjs target-start <target> [--children <a,b>] [--service-backed <0|1>] [--expected-phases <n>] [--expected-tests <n>] [--force]");
  }
  if (!shouldEmitLifecycle(options)) {
    return 0;
  }
  const children = parseTargetListValue(options.children);
  const stats = targetStartStats(target, children);
  const serviceBacked =
    options.service_backed === undefined
      ? stats.serviceBacked
      : options.service_backed === "1" || options.service_backed === "true";
  const expectedPhases =
    options.expected_phases === undefined
      ? stats.expectedPhases
      : parseNonNegativeInteger(options.expected_phases, "expected-phases");
  const expectedTests =
    options.expected_tests === undefined
      ? stats.expectedTests
      : parseNonNegativeInteger(options.expected_tests, "expected-tests");
  const childField = children.length > 0 ? ` children=${children.join(",")}` : "";
  process.stdout.write(
    `[TARGET] start ${target} service_backed=${serviceBacked ? 1 : 0} expected_phases=${expectedPhases} expected_tests=${expectedTests}${childField}\n`,
  );
  return 0;
}

function normalizeTimingBucket(value, runner = "") {
  if (value && timingBucketSet.has(value)) {
    return value;
  }
  if (runner === "go_test" || runner === "vitest" || runner === "playwright") {
    return "test_command";
  }
  return "test_command";
}

function formatBucketSummary(bucket) {
  if (!bucket) {
    return "none";
  }
  return `${bucket.name}(${formatDuration(bucket.duration_ms)})`;
}

function formatTargetBucketSummary(bucket) {
  if (!bucket) {
    return "none";
  }
  return `${bucket.target}:${bucket.name}(${formatDuration(bucket.duration_ms)})`;
}

function computeWindowDurationMs(startTime, endTime) {
  if (!startTime || !endTime) {
    return 0;
  }
  const startMs = Date.parse(startTime);
  const endMs = Date.parse(endTime);
  if (!Number.isFinite(startMs) || !Number.isFinite(endMs) || endMs < startMs) {
    return 0;
  }
  return endMs - startMs;
}

function requiredEnv(name) {
  const value = process.env[name];
  if (value === undefined || value === "") {
    throw new Error(`missing required environment variable ${name}`);
  }
  return value;
}

function optionalEnv(name, fallback = "") {
  return process.env[name] ?? fallback;
}

function optionalLines(name) {
  const value = optionalEnv(name);
  if (value === "") {
    return [];
  }
  return value
    .split("\n")
    .map((entry) => entry.trim())
    .filter(Boolean);
}

function optionalSetFromLines(name) {
  return new Set(optionalLines(name));
}

function parseInteger(name, fallback = 0) {
  const value = process.env[name];
  if (value === undefined || value === "") {
    return fallback;
  }
  const parsed = Number.parseInt(value, 10);
  if (Number.isNaN(parsed)) {
    throw new Error(`invalid integer ${name}=${value}`);
  }
  return parsed;
}

function slugifyLabel(label) {
  return label
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .replace(/--+/g, "-");
}

function splitLogLines(file) {
  if (!file || !existsSync(file)) {
    return [];
  }
  return readFileSync(file, "utf8").split(/\r?\n/);
}

function readNonEmptyFile(file) {
  if (!file || !existsSync(file)) {
    return "";
  }
  const value = readFileSync(file, "utf8");
  return value;
}

function removeEmptyArtifact(file) {
  if (!file || !existsSync(file)) {
    return;
  }
  if (statSync(file).size === 0) {
    rmSync(file, { force: true });
  }
}

function loadManifestIndex() {
  if (cachedManifestIndex) {
    return cachedManifestIndex;
  }

  const index = {
    authoritativeGo: new Map(),
    authoritativeVitest: new Map(),
    authoritativePlaywright: new Map(),
    manifestPlaywright: new Map(),
    forbiddenFilesByPhase: new Map(),
  };

  const toolsDir = path.join(repoRoot, "tools");
  const files = readdirSync(toolsDir)
    .filter((entry) => /^phase\d+_test_map\.json$/.test(entry))
    .sort();

  for (const filename of files) {
    const phase = filename.replace(/_test_map\.json$/, "");
    const { manifest } = loadManifest(repoRoot, phase);
    const entries = collectEntries(manifest);
    for (const forbidden of manifest.forbidden_id_files ?? []) {
      if (!index.forbiddenFilesByPhase.has(phase)) {
        index.forbiddenFilesByPhase.set(phase, new Set());
      }
      index.forbiddenFilesByPhase.get(phase).add(normalizePath(forbidden));
    }
    for (const entry of entries) {
      if (entry.runner === "playwright") {
        index.manifestPlaywright.set(
          `${normalizePath(entry.file)}::${entry.title}`,
          { ...entry, phase },
        );
      }
      if (entry.coverage !== "authoritative") {
        continue;
      }
      if (entry.runner === "go_test") {
        const symbols = entry.symbol !== undefined ? [entry.symbol] : entry.symbols;
        for (const symbol of symbols) {
          index.authoritativeGo.set(
            `${toGoImportPath(entry.package)}::${symbol}`,
            { ...entry, phase },
          );
        }
        continue;
      }
      if (entry.runner === "vitest") {
        index.authoritativeVitest.set(
          `${normalizePath(entry.file)}::${entry.title}`,
          { ...entry, phase },
        );
        continue;
      }
      if (entry.runner === "playwright") {
        index.authoritativePlaywright.set(
          `${normalizePath(entry.file)}::${entry.title}`,
          { ...entry, phase },
        );
      }
    }
  }

  cachedManifestIndex = index;
  return index;
}

function inferPhaseFromText(value) {
  if (!value) {
    return "";
  }
  const patterns = [
    /\bphase(?:\s|_|-)?(\d+)\b/i,
    /\b[UIE][-_](\d+)-\d+\b/,
    /\b[UIE]_(\d+)_\d+\b/,
  ];
  for (const pattern of patterns) {
    const match = value.match(pattern);
    if (match) {
      return `phase${match[1]}`;
    }
  }
  return "";
}

function supportNamedTitle(value) {
  return /^Phase\s+\d+\s+support\b/i.test(value);
}

function firstActionableLine(lines) {
  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    if (
      line.startsWith("=== RUN") ||
      line.startsWith("--- PASS") ||
      line.startsWith("--- FAIL") ||
      line.startsWith("--- SKIP") ||
      line === "PASS" ||
      line === "FAIL" ||
      /^ok\s/.test(line) ||
      /^\?\s/.test(line)
    ) {
      continue;
    }
    return line;
  }
  return "";
}

function firstGoActionableLine(lines) {
  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    if (
      line.startsWith("=== RUN") ||
      line.startsWith("=== PAUSE") ||
      line.startsWith("=== CONT") ||
      line.startsWith("=== NAME") ||
      line.startsWith("--- PASS") ||
      line.startsWith("--- FAIL") ||
      line.startsWith("--- SKIP") ||
      line === "PASS" ||
      line === "FAIL" ||
      /^ok\s/.test(line) ||
      /^\?\s/.test(line)
    ) {
      continue;
    }
    return line;
  }
  return "";
}

function renderList(values) {
  if (values.length === 0) {
    return "none";
  }
  return values.join(",");
}

function printBlock(header, fields) {
  const lines = [header];
  for (const [key, value] of Object.entries(fields)) {
    lines.push(`${key}=${value === "" ? "-" : value}`);
  }
  process.stderr.write(`${lines.join("\n")}\n`);
}

function createBasePhaseContext(runner) {
  const phaseDir = requiredEnv("CARTULARY_PHASE_DIR");
  ensureDir(phaseDir);
  const accountingMode =
    optionalEnv("CARTULARY_REPORT_SLICE") === "1"
      ? normalizeAccountingMode(optionalEnv("CARTULARY_PHASE_ACCOUNTING_MODE", "actual"))
      : "actual";
  const legacyDurationMs = parseInteger("CARTULARY_PHASE_DURATION_MS", 0);
  const logicalDurationMs = clampDurationMs(
    parseInteger("CARTULARY_PHASE_LOGICAL_DURATION_MS", legacyDurationMs),
  );
  const executedDurationMs = clampDurationMs(
    parseInteger(
      "CARTULARY_PHASE_EXECUTED_DURATION_MS",
      accountingMode === "actual" ? logicalDurationMs : 0,
    ),
  );
  return {
    label: requiredEnv("CARTULARY_PHASE_LABEL"),
    phaseDir,
    target: optionalEnv("CARTULARY_TEST_TARGET", "adhoc"),
    command: requiredEnv("CARTULARY_PHASE_COMMAND"),
    runner,
    timingBucket: normalizeTimingBucket(optionalEnv("CARTULARY_PHASE_TIMING_BUCKET"), runner),
    startTime: requiredEnv("CARTULARY_PHASE_START_TIME"),
    endTime: requiredEnv("CARTULARY_PHASE_END_TIME"),
    accountingMode,
    executedDurationMs,
    logicalDurationMs,
    reusedDurationMs: accountingMode === "reused" ? logicalDurationMs : 0,
    derivedDurationMs: accountingMode === "derived" ? logicalDurationMs : 0,
    wallDurationMs: clampDurationMs(parseInteger("CARTULARY_PHASE_WALL_DURATION_MS", logicalDurationMs)),
    exitStatus: parseInteger("CARTULARY_PHASE_EXIT_STATUS", 0),
  };
}

function writePhaseArtifacts(context, details) {
  const playwrightTimingPath = details.playwrightTiming
    ? path.join(context.phaseDir, "playwright-timing.json")
    : "";
  if (details.playwrightTiming) {
    writeJson(playwrightTimingPath, details.playwrightTiming);
  }

  const artifacts = {};
  for (const [key, value] of Object.entries({
    ...(details.artifacts ?? {}),
    playwright_timing: playwrightTimingPath,
  })) {
    if (!value) {
      continue;
    }
    artifacts[key] = relToRepo(value);
  }

  const meta = {
    label: context.label,
    runner: context.runner,
    command: context.command,
    start_time: context.startTime,
    end_time: context.endTime,
    exit_status: context.exitStatus,
    accounting_mode: context.accountingMode,
    executed_duration_ms: context.executedDurationMs,
    logical_duration_ms: context.logicalDurationMs,
    reused_duration_ms: context.reusedDurationMs,
    derived_duration_ms: context.derivedDurationMs,
    wall_duration_ms: context.wallDurationMs,
    critical_path_wall_duration_ms: context.wallDurationMs,
    teardown_duration_ms: context.timingBucket === "teardown" ? context.wallDurationMs : 0,
    timing_bucket: context.timingBucket,
    status: details.status,
    counts: details.counts,
  };

  writeJson(path.join(context.phaseDir, "meta.json"), meta);

  if (details.manifestSummary) {
    writeJson(path.join(context.phaseDir, "manifest-summary.json"), details.manifestSummary);
  }
  if (details.manifestMismatch) {
    writeJson(path.join(context.phaseDir, "manifest-mismatch.json"), details.manifestMismatch);
  }

  const summary = {
    schema_id: phaseSummarySchemaID,
    label: context.label,
    target: context.target,
    runner: context.runner,
    status: details.status,
    phase: details.phase,
    command: context.command,
    start_time: context.startTime,
    end_time: context.endTime,
    accounting_mode: context.accountingMode,
    executed_duration_ms: context.executedDurationMs,
    logical_duration_ms: context.logicalDurationMs,
    reused_duration_ms: context.reusedDurationMs,
    derived_duration_ms: context.derivedDurationMs,
    wall_duration_ms: context.wallDurationMs,
    critical_path_wall_duration_ms: context.wallDurationMs,
    teardown_duration_ms: context.timingBucket === "teardown" ? context.wallDurationMs : 0,
    timing_bucket: context.timingBucket,
    exit_status: context.exitStatus,
    artifacts,
    counts: details.counts,
    owners: details.owners,
    inventory: details.inventory,
    dossiers: details.dossiers,
    manifest_mismatch: details.manifestMismatch,
  };
  writeJson(path.join(context.phaseDir, "phase-summary.json"), summary);
}

function timingSpanPath(target) {
  const targetDir = path.join(resultsRoot, runId, target);
  const label = optionalEnv("CARTULARY_TIMING_LABEL", "timing-span");
  const slug = slugifyLabel(label) || "timing-span";
  const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
  return path.join(targetDir, "timing-spans", `${timestamp}-${process.pid}-${slug}.json`);
}

function createTargetOwnedTimingSpan(source = "target") {
  const target = optionalEnv("CARTULARY_TEST_TARGET", "");
  if (target === "") {
    return null;
  }
  const bucket = normalizeTimingBucket(optionalEnv("CARTULARY_TIMING_BUCKET"));
  const label = requiredEnv("CARTULARY_TIMING_LABEL");
  const durationMs = clampDurationMs(parseInteger("CARTULARY_TIMING_DURATION_MS", 0));
  const startTime = optionalEnv("CARTULARY_TIMING_START_TIME");
  const endTime = optionalEnv("CARTULARY_TIMING_END_TIME");
  return {
    source,
    bucket,
    label,
    start_time: startTime,
    end_time: endTime,
    duration_ms: durationMs,
    status: optionalEnv("CARTULARY_TIMING_STATUS", "pass"),
  };
}

function handleTimingSpan() {
  const span = createTargetOwnedTimingSpan();
  if (!span) {
    return 0;
  }
  writeJson(timingSpanPath(optionalEnv("CARTULARY_TEST_TARGET")), span);
  return 0;
}

function handleSharedExecution(args) {
  const [group, sharedReport, status, startTime, endTime, durationText, exitStatusText, outputPath] =
    args;
  if (
    !group ||
    !sharedReport ||
    !status ||
    !startTime ||
    !endTime ||
    durationText === undefined ||
    exitStatusText === undefined ||
    !outputPath
  ) {
    throw new Error(
      "usage: test-output.mjs shared-execution <group> <shared-report> <status> <start-time> <end-time> <duration-ms> <exit-status> <output-path>",
    );
  }
  const durationMs = clampDurationMs(Number.parseInt(durationText, 10));
  writeJson(outputPath, {
    schema_id: sharedExecutionGroupSchemaID,
    execution_group: group,
    shared_report: sharedReport,
    status,
    start_time: startTime,
    end_time: endTime,
    duration_ms: durationMs,
    wall_duration_ms: durationMs,
    executed_duration_ms: durationMs,
    exit_status: Number.parseInt(exitStatusText, 10) || 0,
    artifact: relToRepo(path.dirname(outputPath)),
  });
  return 0;
}

function loadTargetOwnedTimingSpans(targetDir) {
  const spansDir = path.join(targetDir, "timing-spans");
  if (!existsSync(spansDir)) {
    return [];
  }
  const spans = [];
  for (const entry of readdirSync(spansDir, { withFileTypes: true })) {
    if (!entry.isFile() || !entry.name.endsWith(".json")) {
      continue;
    }
    const span = JSON.parse(readFileSync(path.join(spansDir, entry.name), "utf8"));
    if (!span?.bucket || !timingBucketSet.has(span.bucket)) {
      continue;
    }
    spans.push({
      source: span.source ?? "target",
      bucket: span.bucket,
      label: span.label ?? "",
      start_time: span.start_time ?? "",
      end_time: span.end_time ?? "",
      duration_ms: clampDurationMs(span.duration_ms ?? 0),
      status: span.status ?? "",
    });
  }
  return spans;
}

function loadServiceTimingSpans(target) {
  const servicesRoot = path.join(resultsRoot, runId, "_shared", "test-services");
  if (!existsSync(servicesRoot)) {
    return [];
  }
  const spans = [];
  const stack = [servicesRoot];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (!entry.isFile() || !entry.name.endsWith(".json")) {
        continue;
      }
      let event;
      try {
        event = JSON.parse(readFileSync(next, "utf8"));
      } catch {
        continue;
      }
      if (event.type !== "timing-span") {
        continue;
      }
      const details = event.details ?? {};
      if (details.target !== target) {
        continue;
      }
      const bucket = normalizeTimingBucket(details.bucket);
      spans.push({
        source: "test_services",
        bucket,
        label: details.label ?? event.name ?? "test-services timing",
        start_time: details.start_time ?? "",
        end_time: details.end_time ?? event.timestamp ?? "",
        duration_ms: clampDurationMs(details.duration_ms ?? 0),
        status: details.status ?? event.status ?? "",
        janitorial: details.janitorial === true,
        artifact: relToRepo(path.dirname(path.dirname(next))),
      });
    }
  }
  return spans;
}

function phaseSummaryTimingSpan(summary) {
  return {
    source: "phase",
    bucket: normalizeTimingBucket(summary.timing_bucket, summary.runner),
    label: summary.label,
    runner: summary.runner,
    status: summary.status,
    accounting_mode: normalizeAccountingMode(summary.accounting_mode),
    start_time: summary.start_time ?? "",
    end_time: summary.end_time ?? "",
    duration_ms: clampDurationMs(
      summary.wall_duration_ms ?? summary.logical_duration_ms ?? summary.duration_ms ?? 0,
    ),
    logical_duration_ms: clampDurationMs(summary.logical_duration_ms ?? summary.duration_ms ?? 0),
    executed_duration_ms: clampDurationMs(summary.executed_duration_ms ?? 0),
    artifacts: summary.artifacts ?? {},
  };
}

function addTimingSpanToBuckets(buckets, span) {
  if (!span || !timingBucketSet.has(span.bucket)) {
    return;
  }
  if (!buckets.has(span.bucket)) {
    buckets.set(span.bucket, {
      name: span.bucket,
      duration_ms: 0,
      spans: [],
    });
  }
  const bucket = buckets.get(span.bucket);
  const durationMs = clampDurationMs(span.duration_ms ?? 0);
  bucket.spans.push({
    ...span,
    duration_ms: durationMs,
  });
}

function disjointSpanDurationMs(spans) {
  if (spans.length === 1) {
    return clampDurationMs(spans[0]?.duration_ms ?? 0);
  }

  const intervals = [];
  let fallbackDurationMs = 0;
  for (const span of spans) {
    const durationMs = clampDurationMs(span.duration_ms ?? 0);
    const startMs = Date.parse(span.start_time ?? "");
    const endMs = Date.parse(span.end_time ?? "");
    if (Number.isFinite(startMs) && Number.isFinite(endMs) && endMs >= startMs) {
      intervals.push([startMs, endMs]);
      continue;
    }
    fallbackDurationMs += durationMs;
  }
  intervals.sort((left, right) => left[0] - right[0] || left[1] - right[1]);
  let total = fallbackDurationMs;
  let currentStart = undefined;
  let currentEnd = undefined;
  for (const [startMs, endMs] of intervals) {
    if (currentStart === undefined) {
      currentStart = startMs;
      currentEnd = endMs;
      continue;
    }
    if (startMs <= currentEnd) {
      currentEnd = Math.max(currentEnd, endMs);
      continue;
    }
    total += currentEnd - currentStart;
    currentStart = startMs;
    currentEnd = endMs;
  }
  if (currentStart !== undefined) {
    total += currentEnd - currentStart;
  }
  return clampDurationMs(total);
}

function lifecycleTimingSpans(target, targetDir) {
  return [...loadTargetOwnedTimingSpans(targetDir), ...loadServiceTimingSpans(target)].filter(
    (span) => span.janitorial !== true,
  );
}

function janitorialTimingSpans(target) {
  return loadServiceTimingSpans(target).filter((span) => span.janitorial === true);
}

function timingStatusFailed(status) {
  const normalized = String(status ?? "").trim().toLowerCase();
  if (normalized === "" || normalized === "pass" || normalized === "succeeded") {
    return false;
  }
  return true;
}

function createDurationFields() {
  return {
    wall_duration_ms: 0,
    critical_path_wall_duration_ms: 0,
    executed_duration_ms: 0,
    logical_duration_ms: 0,
    reused_duration_ms: 0,
    derived_duration_ms: 0,
    teardown_duration_ms: 0,
  };
}

function readSummaryDurationFields(summary, accountingMode = normalizeAccountingMode(summary?.accounting_mode)) {
  const logicalDurationMs = clampDurationMs(
    summary?.logical_duration_ms ?? summary?.duration_ms ?? 0,
  );
  const wallDurationMs = clampDurationMs(
    summary?.wall_duration_ms ?? (accountingMode === "actual" ? logicalDurationMs : 0),
  );
  return {
    wall_duration_ms: wallDurationMs,
    critical_path_wall_duration_ms: clampDurationMs(
      summary?.critical_path_wall_duration_ms ?? wallDurationMs,
    ),
    executed_duration_ms: clampDurationMs(
      summary?.executed_duration_ms ?? (accountingMode === "actual" ? logicalDurationMs : 0),
    ),
    logical_duration_ms: logicalDurationMs,
    reused_duration_ms: clampDurationMs(
      summary?.reused_duration_ms ?? (accountingMode === "reused" ? logicalDurationMs : 0),
    ),
    derived_duration_ms: clampDurationMs(
      summary?.derived_duration_ms ?? (accountingMode === "derived" ? logicalDurationMs : 0),
    ),
    teardown_duration_ms: clampDurationMs(
      summary?.teardown_duration_ms ?? (summary?.timing_bucket === "teardown" ? wallDurationMs : 0),
    ),
  };
}

function addDurationFields(target, fields) {
  for (const key of Object.keys(createDurationFields())) {
    target[key] += clampDurationMs(fields?.[key] ?? 0);
  }
}

function durationFieldsForJSON(fields, overrides = {}) {
  return {
    wall_duration_ms: clampDurationMs(overrides.wall_duration_ms ?? fields.wall_duration_ms),
    critical_path_wall_duration_ms: clampDurationMs(
      overrides.critical_path_wall_duration_ms ?? fields.critical_path_wall_duration_ms,
    ),
    executed_duration_ms: clampDurationMs(
      overrides.executed_duration_ms ?? fields.executed_duration_ms,
    ),
    logical_duration_ms: clampDurationMs(overrides.logical_duration_ms ?? fields.logical_duration_ms),
    reused_duration_ms: clampDurationMs(overrides.reused_duration_ms ?? fields.reused_duration_ms),
    derived_duration_ms: clampDurationMs(overrides.derived_duration_ms ?? fields.derived_duration_ms),
    teardown_duration_ms: clampDurationMs(
      overrides.teardown_duration_ms ?? fields.teardown_duration_ms,
    ),
  };
}

function timingFailureReference(span) {
  return {
    source: span.source ?? "",
    bucket: span.bucket ?? "",
    label: span.label ?? "",
    status: span.status ?? "",
    start_time: span.start_time ?? "",
    end_time: span.end_time ?? "",
    wall_duration_ms: clampDurationMs(span.duration_ms ?? 0),
    artifact: span.artifact ?? "",
  };
}

function timingFailuresFromSpans(spans) {
  return spans.filter((span) => timingStatusFailed(span.status)).map(timingFailureReference);
}

function teardownStatus(teardownDurationMs, teardownFailures) {
  if (teardownFailures.length > 0) {
    return "fail";
  }
  if (teardownDurationMs > 0) {
    return "pass";
  }
  return "none";
}

function loadSharedExecutionRecords() {
  const sharedRoot = path.join(resultsRoot, runId, "_shared");
  if (!existsSync(sharedRoot)) {
    return [];
  }
  const records = [];
  const stack = [sharedRoot];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (!entry.isFile() || entry.name !== "shared-execution.json") {
        continue;
      }
      let record;
      try {
        record = JSON.parse(readFileSync(next, "utf8"));
      } catch {
        continue;
      }
      if (record?.schema_id !== sharedExecutionGroupSchemaID || !record.execution_group) {
        continue;
      }
      records.push({
        execution_group: record.execution_group,
        shared_report: record.shared_report ?? "",
        status: record.status ?? "",
        start_time: record.start_time ?? "",
        end_time: record.end_time ?? "",
        duration_ms: clampDurationMs(record.duration_ms ?? 0),
        wall_duration_ms: clampDurationMs(record.wall_duration_ms ?? record.duration_ms ?? 0),
        executed_duration_ms: clampDurationMs(
          record.executed_duration_ms ?? record.duration_ms ?? 0,
        ),
        exit_status: clampDurationMs(record.exit_status ?? 0),
        artifact: record.artifact ?? relToRepo(path.dirname(next)),
      });
    }
  }
  return records.sort((left, right) =>
    `${left.execution_group}:${left.shared_report}`.localeCompare(
      `${right.execution_group}:${right.shared_report}`,
    ),
  );
}

function buildSharedExecutionGroups() {
  const byGroup = new Map();
  for (const record of loadSharedExecutionRecords()) {
    if (!byGroup.has(record.execution_group)) {
      byGroup.set(record.execution_group, []);
    }
    byGroup.get(record.execution_group).push(record);
  }

  return [...byGroup.entries()]
    .sort((left, right) => left[0].localeCompare(right[0]))
    .map(([name, records]) => {
      const startTime = records
        .map((record) => record.start_time)
        .filter((value) => Number.isFinite(Date.parse(value)))
        .sort((left, right) => Date.parse(left) - Date.parse(right))[0] ?? "";
      const endTimes = records
        .map((record) => record.end_time)
        .filter((value) => Number.isFinite(Date.parse(value)))
        .sort((left, right) => Date.parse(left) - Date.parse(right));
      const endTime = endTimes[endTimes.length - 1] ?? "";
      const wallDurationMs = disjointSpanDurationMs(
        records.map((record) => ({
          start_time: record.start_time,
          end_time: record.end_time,
          duration_ms: record.wall_duration_ms,
        })),
      );
      const executedDurationMs = records.reduce(
        (total, record) => total + clampDurationMs(record.executed_duration_ms),
        0,
      );
      const failed = records.some((record) => timingStatusFailed(record.status));
      return {
        schema_id: sharedExecutionGroupSchemaID,
        name,
        status: failed ? "fail" : "pass",
        start_time: startTime,
        end_time: endTime,
        wall_duration_ms: wallDurationMs,
        critical_path_wall_duration_ms: wallDurationMs,
        executed_duration_ms: executedDurationMs,
        shared_reports: records.map((record) => record.shared_report).filter(Boolean).sort(),
        reports: records.length,
      };
    });
}

function accountableTargetWallSpan(span) {
  if (!span || span.bucket === "report_collation") {
    return false;
  }
  if (span.source === "phase") {
    return normalizeAccountingMode(span.accounting_mode) === "actual";
  }
  return true;
}

function summarizeAccountableTargetWindow(spans) {
  const accountableSpans = spans.filter(accountableTargetWallSpan);
  const summedDurationMs = accountableSpans.reduce(
    (total, span) => total + clampDurationMs(span.duration_ms ?? 0),
    0,
  );
  const startTimes = accountableSpans
    .map((span) => span.start_time)
    .filter((value) => Number.isFinite(Date.parse(value)))
    .sort((left, right) => Date.parse(left) - Date.parse(right));
  const endTimes = accountableSpans
    .map((span) => span.end_time)
    .filter((value) => Number.isFinite(Date.parse(value)))
    .sort((left, right) => Date.parse(left) - Date.parse(right));
  const startTime = startTimes[0] ?? "";
  const endTime = endTimes[endTimes.length - 1] ?? "";
  const windowDurationMs = computeWindowDurationMs(startTime, endTime);
  const wallDurationMs =
    accountableSpans.length === 1
      ? summedDurationMs
      : windowDurationMs > 0
        ? windowDurationMs
        : summedDurationMs;
  return {
    startTime,
    endTime,
    wallDurationMs,
  };
}

function summarizeTargetTiming(
  target,
  targetDir,
  phaseSummaries,
  status,
  reportCollationSpan,
  lifecycleSpans = lifecycleTimingSpans(target, targetDir),
) {
  const buckets = new Map();
  const phaseSpans = phaseSummaries.map((summary) => phaseSummaryTimingSpan(summary));
  for (const span of phaseSpans) {
    addTimingSpanToBuckets(buckets, span);
  }
  for (const span of lifecycleSpans) {
    addTimingSpanToBuckets(buckets, span);
  }
  addTimingSpanToBuckets(buckets, reportCollationSpan);
  const accountableWindow = summarizeAccountableTargetWindow([...phaseSpans, ...lifecycleSpans]);

  const bucketList = timingBucketOrder
    .map((name) => buckets.get(name))
    .filter(Boolean)
    .map((bucket) => ({
      ...bucket,
      duration_ms: disjointSpanDurationMs(bucket.spans),
      spans: bucket.spans.sort((left, right) =>
        `${left.start_time ?? ""}:${left.label ?? ""}`.localeCompare(
          `${right.start_time ?? ""}:${right.label ?? ""}`,
        ),
      ),
    }));
  const slowest = bucketList.reduce((current, bucket) => {
    if (!current || bucket.duration_ms > current.duration_ms) {
      return { name: bucket.name, duration_ms: bucket.duration_ms };
    }
    return current;
  }, null);
  const timing = {
    schema_id: targetTimingSchemaID,
    target,
    status,
    generated_at: new Date().toISOString(),
    start_time: accountableWindow.startTime,
    end_time: accountableWindow.endTime,
    buckets: bucketList,
    slowest_lifecycle_bucket: slowest,
  };
  const timingPath = path.join(targetDir, "target-timing.json");
  writeJson(timingPath, timing);
  return { timing, timingPath, accountableWindow };
}

function summarizeTargetDir(target) {
  const targetDir = path.join(resultsRoot, runId, target);
  const summaries = [];
  if (existsSync(targetDir)) {
    const stack = [targetDir];
    while (stack.length > 0) {
      const current = stack.pop();
      for (const entry of readdirSync(current, { withFileTypes: true })) {
        const next = path.join(current, entry.name);
        if (entry.isDirectory()) {
          stack.push(next);
          continue;
        }
        if (entry.isFile() && entry.name === "phase-summary.json") {
          summaries.push(JSON.parse(readFileSync(next, "utf8")));
        }
      }
    }
  }
  summaries.sort((left, right) => left.start_time.localeCompare(right.start_time));

  const owners = new Set();
  const authoritativeInventory = [];
  const supportInventory = [];
  const counts = {
    phases: summaries.length,
    ...createCounts(),
  };
  let startTime = "";
  let endTime = "";
  let actualStartTime = "";
  let actualEndTime = "";
  const durations = createDurationFields();
  const accountingModes = createAccountingModes();
  let failed = false;

  for (const summary of summaries) {
    const accountingMode = normalizeAccountingMode(summary.accounting_mode);
    const summaryDurations = readSummaryDurationFields(summary, accountingMode);
    if (startTime === "" || summary.start_time < startTime) {
      startTime = summary.start_time;
    }
    if (endTime === "" || summary.end_time > endTime) {
      endTime = summary.end_time;
    }
    if (accountingMode === "actual") {
      if (actualStartTime === "" || summary.start_time < actualStartTime) {
        actualStartTime = summary.start_time;
      }
      if (actualEndTime === "" || summary.end_time > actualEndTime) {
        actualEndTime = summary.end_time;
      }
    }
    addDurationFields(durations, summaryDurations);
    accountingModes[accountingMode] += 1;
    counts.tests += summary.counts?.tests ?? 0;
    counts.failed += summary.counts?.failed ?? 0;
    counts.authoritative += summary.counts?.authoritative ?? 0;
    counts.support += summary.counts?.support ?? 0;
    counts.unmapped += summary.counts?.unmapped ?? 0;
    counts.non_test += summary.counts?.non_test ?? 0;
    counts.authoritative_failed += summary.counts?.authoritative_failed ?? 0;
    counts.support_failed += summary.counts?.support_failed ?? 0;
    counts.unmapped_failed += summary.counts?.unmapped_failed ?? 0;
    counts.non_test_failed += summary.counts?.non_test_failed ?? 0;
    for (const owner of summary.owners ?? []) {
      owners.add(owner);
    }
    for (const item of summary.inventory ?? []) {
      if (item.coverage === "authoritative") {
        authoritativeInventory.push(item);
      } else if (item.coverage === "support") {
        supportInventory.push(item);
      }
    }
    if (summary.status !== "pass") {
      failed = true;
    }
  }
  counts.packages = owners.size;

  const actualWindowWallDurationMs = computeWindowDurationMs(actualStartTime, actualEndTime);
  const wallDurationMs =
    actualWindowWallDurationMs > 0 ? actualWindowWallDurationMs : durations.wall_duration_ms;

  return {
    target,
    targetDir,
    summaries,
    counts,
    startTime,
    endTime,
    durations: durationFieldsForJSON(durations, {
      wall_duration_ms: wallDurationMs,
      critical_path_wall_duration_ms: wallDurationMs,
    }),
    executedDurationMs: durations.executed_duration_ms,
    logicalDurationMs: durations.logical_duration_ms,
    wallDurationMs,
    criticalPathWallDurationMs: wallDurationMs,
    reusedDurationMs: durations.reused_duration_ms,
    derivedDurationMs: durations.derived_duration_ms,
    teardownDurationMs: durations.teardown_duration_ms,
    accountingModes,
    failed,
    authoritativeInventory,
    supportInventory,
  };
}

function printInventory(targetSummary) {
  if (process.env.CARTULARY_TEST_INVENTORY !== "1") {
    return;
  }
  const sections = [
    ["authoritative", targetSummary.authoritativeInventory],
    ["support", targetSummary.supportInventory],
  ];
  for (const [coverage, items] of sections) {
    if (items.length === 0) {
      continue;
    }
    const uniqueItems = Array.from(
      new Map(
        items.map((item) => [
          `${item.coverage}::${item.phase}::${item.id}::${item.package_or_file}::${item.symbol_or_title}`,
          item,
        ]),
      ).values(),
    );
    process.stdout.write(`[INVENTORY] ${targetSummary.target} ${coverage}=${uniqueItems.length}\n`);
    const sorted = [...uniqueItems].sort((left, right) => {
      const leftKey = `${left.phase}::${left.id}::${left.package_or_file}::${left.symbol_or_title}`;
      const rightKey = `${right.phase}::${right.id}::${right.package_or_file}::${right.symbol_or_title}`;
      return leftKey.localeCompare(rightKey);
    });
    for (const item of sorted) {
      process.stdout.write(
        `${coverage} phase=${item.phase || "-"} id=${item.id || "-"} owner=${item.package_or_file} name=${item.symbol_or_title}\n`,
      );
    }
  }
}

function parseTargetList(value) {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter((item) => item.length > 0);
}

function parseTargetSummaryArgs(args) {
  const [target, ...rest] = args;
  if (!target) {
    throw new Error("usage: test-output.mjs target-summary <target> [pass|fail] [--children <target,target,...>] [--projection <target>] [--quiet-success]");
  }

  let requestedStatus = "pass";
  let projectionTarget = "";
  let quietSuccess = false;
  const remaining = [...rest];
  if (remaining.length > 0 && !remaining[0].startsWith("--")) {
    requestedStatus = remaining.shift();
  }

  const childTargetNames = [];
  while (remaining.length > 0) {
    const option = remaining.shift();
    if (option === "--quiet-success") {
      quietSuccess = true;
      continue;
    }
    if (option === "--projection") {
      projectionTarget = remaining.shift() ?? "";
      if (projectionTarget === "") {
        throw new Error("--projection requires a target name");
      }
      continue;
    }
    if (option !== "--children") {
      throw new Error(`unknown target-summary option ${option}`);
    }
    const value = remaining.shift();
    if (value === undefined) {
      throw new Error("--children requires a comma-separated target list");
    }
    childTargetNames.push(...parseTargetList(value));
  }

  if (projectionTarget && childTargetNames.length === 0) {
    const { manifest } = loadTaskSurfaceManifest(
      process.env.TASK_SURFACE_MANIFEST ?? defaultTaskSurfaceManifestPath,
    );
    childTargetNames.push(...projectionChildren(manifest, projectionTarget));
  }

  return { target, requestedStatus, childTargetNames, quietSuccess };
}

function targetSummaryPath(target) {
  return path.join(resultsRoot, runId, target, "target-summary.json");
}

function loadTargetSummary(target) {
  const file = targetSummaryPath(target);
  if (!existsSync(file)) {
    return undefined;
  }
  return JSON.parse(readFileSync(file, "utf8"));
}

function normalizeCounts(counts = {}) {
  return {
    phases: clampDurationMs(counts.phases ?? 0),
    tests: clampDurationMs(counts.tests ?? 0),
    failed: clampDurationMs(counts.failed ?? 0),
    authoritative: clampDurationMs(counts.authoritative ?? 0),
    support: clampDurationMs(counts.support ?? 0),
    unmapped: clampDurationMs(counts.unmapped ?? 0),
    non_test: clampDurationMs(counts.non_test ?? 0),
    authoritative_failed: clampDurationMs(counts.authoritative_failed ?? 0),
    support_failed: clampDurationMs(counts.support_failed ?? 0),
    unmapped_failed: clampDurationMs(counts.unmapped_failed ?? 0),
    non_test_failed: clampDurationMs(counts.non_test_failed ?? 0),
    packages: clampDurationMs(counts.packages ?? 0),
  };
}

function addCounts(target, source) {
  const normalized = normalizeCounts(source);
  for (const key of Object.keys(normalized)) {
    target[key] += normalized[key];
  }
}

function sectionFromFlatSummary(summary, fallbackTarget) {
  const durations = readSummaryDurationFields(summary);
  const counts = normalizeCounts(summary?.counts ?? {});
  return {
    target: summary?.target ?? fallbackTarget,
    status: summary?.status ?? "",
    start_time: summary?.start_time ?? "",
    end_time: summary?.end_time ?? "",
    ...durations,
    accounting_modes: resolveAccountingModes(summary?.accounting_modes, counts.phases),
    counts,
    slowest_lifecycle_bucket: summary?.slowest_lifecycle_bucket ?? null,
    timing_failures: summary?.timing_failures ?? [],
    teardown_status: summary?.teardown_status ?? teardownStatus(durations.teardown_duration_ms, []),
    teardown_failures: summary?.teardown_failures ?? [],
    fixture: normalizeFixtureSummary(summary?.target ?? fallbackTarget, summary?.fixture),
    artifacts: {
      dir: summary?.artifacts?.dir ?? relToRepo(path.join(resultsRoot, runId, fallbackTarget)),
      timing_json: summary?.artifacts?.timing_json ?? "",
    },
  };
}

function targetSummarySection(summary, sectionName, fallbackTarget) {
  if (summary?.[sectionName]) {
    return sectionFromFlatSummary(summary[sectionName], summary.target ?? fallbackTarget);
  }
  return sectionFromFlatSummary(summary, fallbackTarget);
}

function targetSummaryAccountingView(summary, fallbackTarget = summary?.target ?? "") {
  return targetSummarySection(summary, "totals", fallbackTarget);
}

function toTargetSummaryReference(summary, fallbackTarget) {
  const own = targetSummarySection(summary, "own", fallbackTarget);
  const totals = targetSummaryAccountingView(summary, fallbackTarget);
  return {
    schema_id: summary.schema_id ?? "",
    kind: summary.kind ?? "leaf",
    target: summary.target ?? fallbackTarget,
    status: summary.status ?? "",
    start_time: totals.start_time,
    end_time: totals.end_time,
    ...durationFieldsForJSON(totals),
    counts: totals.counts,
    accounting_modes: totals.accounting_modes,
    slowest_lifecycle_bucket: totals.slowest_lifecycle_bucket,
    timing_failures: totals.timing_failures,
    teardown_status: totals.teardown_status,
    teardown_failures: totals.teardown_failures,
    fixture: totals.fixture,
    artifacts: own.artifacts,
    own,
    children: summary.children ?? {
      expected: [],
      present: [],
      missing: [],
      status: "pass",
      ...durationFieldsForJSON(createDurationFields()),
      accounting_modes: createAccountingModes(),
      counts: normalizeCounts(),
      fixture: emptyFixtureSummary(summary.target ?? fallbackTarget),
      failed_targets: [],
    },
    totals,
  };
}

function loadChildTargetSummaries(childTargetNames) {
  const childTargets = [];
  const missingChildTargetSummaries = [];
  for (const childTarget of childTargetNames) {
    const summary = loadTargetSummary(childTarget);
    if (!summary) {
      missingChildTargetSummaries.push(childTarget);
      continue;
    }
    childTargets.push(toTargetSummaryReference(summary, childTarget));
  }
  return { childTargets, missingChildTargetSummaries };
}

function combineSummarySections(target, sections, status = "pass") {
  const aggregate = createDurationAggregate();
  const accountingModes = createAccountingModes();
  const timingFailures = [];
  const teardownFailures = [];
  let startTime = "";
  let endTime = "";
  let failed = status !== "pass";

  for (const section of sections) {
    const view = targetSummaryAccountingView(section, section?.target ?? target);
    addCounts(aggregate, view.counts);
    addDurationFields(aggregate, view);
    mergeAccountingModes(accountingModes, view.accounting_modes);
    timingFailures.push(...(view.timing_failures ?? []));
    teardownFailures.push(...(view.teardown_failures ?? []));
    if (startTime === "" || (view.start_time && view.start_time < startTime)) {
      startTime = view.start_time ?? "";
    }
    if (endTime === "" || (view.end_time && view.end_time > endTime)) {
      endTime = view.end_time ?? "";
    }
    if (view.status && view.status !== "pass") {
      failed = true;
    }
  }

  const windowWallDurationMs = computeWindowDurationMs(startTime, endTime);
  const wallDurationMs = windowWallDurationMs > 0 ? windowWallDurationMs : aggregate.wall_duration_ms;
  const criticalPathWallDurationMs = wallDurationMs;
  return {
    target,
    status: failed ? "fail" : "pass",
    start_time: startTime,
    end_time: endTime,
    ...durationFieldsForJSON(aggregate, {
      wall_duration_ms: wallDurationMs,
      critical_path_wall_duration_ms: criticalPathWallDurationMs,
    }),
    accounting_modes: accountingModes,
    counts: countsForJSON(aggregate),
    slowest_lifecycle_bucket: findSlowestLifecycleBucket(sections),
    timing_failures: timingFailures,
    teardown_status: teardownStatus(aggregate.teardown_duration_ms, teardownFailures),
    teardown_failures: teardownFailures,
  };
}

function writeTargetLine(stream, label, targetSummary) {
  const target = targetSummary.target;
  if (targetSummary.kind === "aggregate") {
    const own = targetSummary.own;
    const children = targetSummary.children;
    const totals = targetSummary.totals;
    const slowestChild = findSlowestTarget(children.present ?? []);
    const slowestChildField = slowestChild
      ? `${slowestChild.target}(${formatDuration(slowestChild.critical_path_wall_duration_ms)})`
      : "none";
    const failedChildren =
      (children.failed_targets ?? []).length > 0 ? children.failed_targets.join(",") : "none";
    stream.write(
      `${label} ${target} kind=aggregate children=${children.present.length}/${children.expected.length} child_tests=${children.counts.tests} child_failed=${children.counts.failed} failed_children=${failedChildren} slowest_child=${slowestChildField} own_phases=${own.counts.phases} own_tests=${own.counts.tests} own_failed=${own.counts.failed} total_tests=${totals.counts.tests} total_failed=${totals.counts.failed} ${formatDurationFields(totals.wall_duration_ms, totals.executed_duration_ms, totals.logical_duration_ms, totals.critical_path_wall_duration_ms, totals.teardown_duration_ms)} ${formatAccountingModeFields(totals.accounting_modes)} own_fixture_count=${own.fixture.total_count} own_fixture_duration=${formatDuration(own.fixture.total_duration_ms)} child_fixture_count=${children.fixture.total_count} child_fixture_duration=${formatDuration(children.fixture.total_duration_ms)} total_fixture_count=${totals.fixture.total_count} total_fixture_duration=${formatDuration(totals.fixture.total_duration_ms)} slowest_lifecycle_bucket=${formatBucketSummary(totals.slowest_lifecycle_bucket)} artifacts=${targetSummary.own.artifacts.dir}\n`,
    );
    return;
  }

  const totals = targetSummary.totals;
  stream.write(
    `${label} ${target} kind=leaf phases=${totals.counts.phases} tests=${totals.counts.tests} failed=${totals.counts.failed} authoritative=${totals.counts.authoritative} support=${totals.counts.support} unmapped=${totals.counts.unmapped} packages=${totals.counts.packages} ${formatDurationFields(totals.wall_duration_ms, totals.executed_duration_ms, totals.logical_duration_ms, totals.critical_path_wall_duration_ms, totals.teardown_duration_ms)} ${formatAccountingModeFields(totals.accounting_modes)} fixture_count=${totals.fixture.total_count} fixture_duration=${formatDuration(totals.fixture.total_duration_ms)} slowest_lifecycle_bucket=${formatBucketSummary(totals.slowest_lifecycle_bucket)} artifacts=${targetSummary.own.artifacts.dir}\n`,
  );
}

function fixtureLineOptions() {
  return {
    thresholdMs:
      process.env.FIXTURE_THRESHOLD_MS ?? process.env.CARTULARY_FIXTURE_THRESHOLD_MS,
    top: process.env.FIXTURE_TOP ?? process.env.CARTULARY_FIXTURE_TOP,
  };
}

function writeFixtureLine(stream, fixture) {
  const line = fixtureSummaryLine(fixture, fixtureLineOptions());
  if (line) {
    stream.write(`${line}\n`);
  }
}

function writeChildTargetLines(stream, parentTarget, childTargets, missingChildTargetSummaries) {
  for (const child of childTargets) {
    const totals = child.totals ?? targetSummaryAccountingView(child, child.target);
    stream.write(
      `[CHILD] ${parentTarget} ${child.target} status=${child.status} phases=${totals.counts?.phases ?? 0} tests=${totals.counts?.tests ?? 0} failed=${totals.counts?.failed ?? 0} ${formatDurationFields(totals.wall_duration_ms, totals.executed_duration_ms, totals.logical_duration_ms, totals.critical_path_wall_duration_ms, totals.teardown_duration_ms)} ${formatAccountingModeFields(totals.accounting_modes)} artifacts=${child.artifacts?.dir ?? ""}\n`,
    );
  }
  for (const childTarget of missingChildTargetSummaries) {
    stream.write(
      `[CHILD-MISSING] ${parentTarget} ${childTarget} artifacts=${relToRepo(targetSummaryPath(childTarget))}\n`,
    );
  }
}

function handleTargetSummary(args) {
  const reportCollationStartMs = Date.now();
  const reportCollationStartTime = new Date(reportCollationStartMs).toISOString();
  const { target, requestedStatus, childTargetNames, quietSuccess } = parseTargetSummaryArgs(args);
  const summary = summarizeTargetDir(target);
  const lifecycleSpans = lifecycleTimingSpans(target, summary.targetDir);
  const timingFailures = timingFailuresFromSpans(lifecycleSpans);
  const teardownFailures = timingFailures.filter((failure) => failure.bucket === "teardown");
  const { childTargets, missingChildTargetSummaries } =
    loadChildTargetSummaries(childTargetNames);
  const failedChildTargets = childTargets
    .filter((child) => child.status !== "pass")
    .map((child) => child.target);
  if (timingFailures.length > 0) {
    summary.counts.failed += timingFailures.length;
    summary.counts.non_test += timingFailures.length;
    summary.counts.non_test_failed += timingFailures.length;
  }
  if (missingChildTargetSummaries.length > 0) {
    summary.counts.failed += missingChildTargetSummaries.length;
    summary.counts.non_test += missingChildTargetSummaries.length;
    summary.counts.non_test_failed += missingChildTargetSummaries.length;
  }
  if (
    requestedStatus === "fail" &&
    summary.failed === false &&
    timingFailures.length === 0 &&
    missingChildTargetSummaries.length === 0 &&
    failedChildTargets.length === 0
  ) {
    summary.counts.failed += 1;
    summary.counts.non_test += 1;
    summary.counts.non_test_failed += 1;
  }
  const ownFailed =
    summary.failed ||
    timingFailures.length > 0 ||
    missingChildTargetSummaries.length > 0 ||
    (requestedStatus === "fail" && failedChildTargets.length === 0);
  const status =
    ownFailed ||
    failedChildTargets.length > 0 ||
    requestedStatus === "fail"
      ? "FAIL"
      : "PASS";
  const reportCollationEndMs = Date.now();
  const reportCollationEndTime = new Date(reportCollationEndMs).toISOString();
  const { timing, timingPath, accountableWindow } = summarizeTargetTiming(
    target,
    summary.targetDir,
    summary.summaries,
    status.toLowerCase(),
    {
      source: "target",
      bucket: "report_collation",
      label: "target summary collation",
      start_time: reportCollationStartTime,
      end_time: reportCollationEndTime,
      duration_ms: clampDurationMs(reportCollationEndMs - reportCollationStartMs),
      status: status.toLowerCase(),
    },
    lifecycleSpans,
  );
  summary.wallDurationMs = accountableWindow.wallDurationMs;
  summary.criticalPathWallDurationMs = accountableWindow.wallDurationMs;
  summary.startTime = accountableWindow.startTime;
  summary.endTime = accountableWindow.endTime;
  summary.slowestLifecycleBucket = timing.slowest_lifecycle_bucket;
  const ownFixture = summarizeFixtureActivities(target, { resultsRoot, runId, repoRoot });
  const childFixture = combineFixtureSummaries(target, null, childTargets);
  const totalFixture = combineFixtureSummaries(target, ownFixture, childTargets);
  summary.teardownDurationMs = clampDurationMs(
    timing.buckets.find((bucket) => bucket.name === "teardown")?.duration_ms ?? 0,
  );
  summary.durations = durationFieldsForJSON(summary.durations, {
    wall_duration_ms: summary.wallDurationMs,
    critical_path_wall_duration_ms: summary.criticalPathWallDurationMs,
    teardown_duration_ms: summary.teardownDurationMs,
  });
  const ownSection = {
    target,
    status: ownFailed ? "fail" : "pass",
    start_time: summary.startTime,
    end_time: summary.endTime,
    ...summary.durations,
    accounting_modes: summary.accountingModes,
    counts: normalizeCounts(summary.counts),
    slowest_lifecycle_bucket: timing.slowest_lifecycle_bucket,
    timing_failures: timingFailures,
    janitorial_timing: janitorialTimingSpans(target),
    teardown_status: teardownStatus(summary.teardownDurationMs, teardownFailures),
    teardown_failures: teardownFailures,
    fixture: ownFixture,
    artifacts: {
      dir: relToRepo(summary.targetDir),
      timing_json: relToRepo(timingPath),
    },
  };
  const childrenRollup = combineSummarySections(target, childTargets);
  const childrenSection = {
    target,
    status:
      failedChildTargets.length > 0 || missingChildTargetSummaries.length > 0 ? "fail" : "pass",
    expected: childTargetNames,
    present: childTargets,
    missing: missingChildTargetSummaries,
    failed_targets: failedChildTargets,
    start_time: childrenRollup.start_time,
    end_time: childrenRollup.end_time,
    ...durationFieldsForJSON(childrenRollup),
    accounting_modes: childrenRollup.accounting_modes,
    counts: childrenRollup.counts,
    slowest_lifecycle_bucket: childrenRollup.slowest_lifecycle_bucket,
    timing_failures: childrenRollup.timing_failures,
    teardown_status: childrenRollup.teardown_status,
    teardown_failures: childrenRollup.teardown_failures,
    fixture: childFixture,
  };
  const totalRollup =
    childTargetNames.length === 0
      ? ownSection
      : combineSummarySections(target, [ownSection, childrenSection], status.toLowerCase());
  const totalsSection =
    childTargetNames.length === 0
      ? { ...ownSection, fixture: ownFixture }
      : {
          ...totalRollup,
          status: status.toLowerCase(),
          fixture: totalFixture,
        };
  const targetSummary = {
    schema_id: targetSummarySchemaID,
    target,
    kind: childTargetNames.length > 0 ? "aggregate" : "leaf",
    status: status.toLowerCase(),
    start_time: summary.startTime,
    end_time: summary.endTime,
    ...durationFieldsForJSON(totalsSection),
    accounting_modes: totalsSection.accounting_modes,
    artifacts: ownSection.artifacts,
    own: ownSection,
    children: childrenSection,
    totals: totalsSection,
  };
  writeJson(path.join(summary.targetDir, "target-summary.json"), targetSummary);

  if (status === "PASS") {
    if (quietSuccess && quietOutputMode()) {
      return 0;
    }
    writeTargetLine(process.stdout, "[PASS]", targetSummary);
    writeFixtureLine(process.stdout, targetSummary.totals.fixture);
    if (!quietOutputMode()) {
      writeChildTargetLines(process.stdout, target, childTargets, missingChildTargetSummaries);
    }
    printInventory(summary);
    return 0;
  }

  writeTargetLine(process.stderr, "[FAIL]", targetSummary);
  writeFixtureLine(process.stderr, targetSummary.totals.fixture);
  writeChildTargetLines(process.stderr, target, childTargets, missingChildTargetSummaries);
  return 0;
}

function parseSummaryGroupsSpec(value) {
  if (!value) {
    return [];
  }
  return value
    .split(";")
    .map((group) => group.trim())
    .filter((group) => group.length > 0)
    .map((group) => {
      const separator = group.indexOf("=");
      if (separator <= 0) {
        throw new Error(`invalid summary group ${group}; expected <name>=<target,target>`);
      }
      const name = group.slice(0, separator).trim();
      const targets = parseTargetList(group.slice(separator + 1));
      if (targets.length === 0) {
        throw new Error(`invalid summary group ${name}; expected at least one target`);
      }
      return { name, targets };
    });
}

function parseRunSummaryArgs(args) {
  const [label, requestedStatus = "pass", completedText = "0", totalText = "0", abortedAfter = "", ...remaining] = args;
  if (!label) {
    throw new Error("usage: test-output.mjs run-summary <label> <pass|fail> <completed> <total> <aborted_after|-> [--summary-groups <name=a,b;name=c>] [--skipped-after-failure <target,target>] [--quiet-success] [targets...]");
  }
  const targets = [];
  const summaryGroups = [];
  const skippedAfterFailure = [];
  let quietSuccess = false;
  while (remaining.length > 0) {
    const value = remaining.shift();
    if (value === "--quiet-success") {
      quietSuccess = true;
      continue;
    }
    if (value === "--summary-groups") {
      const spec = remaining.shift();
      if (spec === undefined) {
        throw new Error("--summary-groups requires <name=a,b;name=c>");
      }
      summaryGroups.push(...parseSummaryGroupsSpec(spec));
      continue;
    }
    if (value === "--skipped-after-failure") {
      const spec = remaining.shift();
      if (spec === undefined) {
        throw new Error("--skipped-after-failure requires <target,target>");
      }
      skippedAfterFailure.push(...parseTargetList(spec));
      continue;
    }
    targets.push(value);
  }
  return { label, requestedStatus, completedText, totalText, abortedAfter, targets, summaryGroups, skippedAfterFailure, quietSuccess };
}

function createDurationAggregate() {
  return {
    phases: 0,
    ...createCounts(),
    ...createDurationFields(),
  };
}

function addSummaryToAggregate(aggregate, accountingModes, summary) {
  const view = targetSummaryAccountingView(summary);
  addCounts(aggregate, view.counts);
  addDurationFields(aggregate, view);
  mergeAccountingModes(
    accountingModes,
    resolveAccountingModes(view.accounting_modes, view.counts?.phases ?? 0),
  );
}

function countsForJSON(aggregate) {
  return {
    phases: aggregate.phases,
    tests: aggregate.tests,
    failed: aggregate.failed,
    authoritative: aggregate.authoritative,
    support: aggregate.support,
    unmapped: aggregate.unmapped,
    non_test: aggregate.non_test,
    authoritative_failed: aggregate.authoritative_failed,
    support_failed: aggregate.support_failed,
    unmapped_failed: aggregate.unmapped_failed,
    non_test_failed: aggregate.non_test_failed,
    packages: aggregate.packages,
  };
}

function summarizeTargetSummaries(summaries, missingTargetSummaries, requestedStatus = "pass") {
  const aggregate = createDurationAggregate();
  const accountingModes = createAccountingModes();
  let failed = requestedStatus === "fail" || missingTargetSummaries.length > 0;
  let startTime = "";
  let endTime = "";
  const timingFailures = [];
  const teardownFailures = [];

  for (const summary of summaries) {
    const view = targetSummaryAccountingView(summary);
    addSummaryToAggregate(aggregate, accountingModes, summary);
    timingFailures.push(...(view.timing_failures ?? []));
    teardownFailures.push(...(view.teardown_failures ?? []));
    if (startTime === "" || (view.start_time && view.start_time < startTime)) {
      startTime = view.start_time ?? "";
    }
    if (endTime === "" || (view.end_time && view.end_time > endTime)) {
      endTime = view.end_time ?? "";
    }
    if (summary.status !== "pass") {
      failed = true;
    }
  }

  if (requestedStatus === "fail" && aggregate.failed === 0) {
    aggregate.phases += 1;
    aggregate.failed += 1;
    aggregate.non_test += 1;
    aggregate.non_test_failed += 1;
  }

  const windowWallDurationMs = computeWindowDurationMs(startTime, endTime);
  const wallDurationMs = windowWallDurationMs > 0 ? windowWallDurationMs : aggregate.wall_duration_ms;
  const criticalPathWallDurationMs = wallDurationMs;
  return {
    aggregate,
    accountingModes,
    failed,
    startTime,
    endTime,
    wallDurationMs,
    criticalPathWallDurationMs,
    timingFailures,
    teardownFailures,
  };
}

function buildSummaryGroups(summaryGroups, skippedAfterFailureSet = new Set()) {
  return summaryGroups.map((group) => {
    const groupSummaries = [];
    const missingTargetSummaries = [];
    const skippedAfterFailure = [];
    for (const target of group.targets) {
      const summary = loadTargetSummary(target);
      if (!summary) {
        if (skippedAfterFailureSet.has(target)) {
          skippedAfterFailure.push(target);
          continue;
        }
        missingTargetSummaries.push(target);
        continue;
      }
      groupSummaries.push(summary);
    }
    const summarized = summarizeTargetSummaries(groupSummaries, missingTargetSummaries);
    return {
      name: group.name,
      status: summarized.failed ? "fail" : "pass",
      targets: group.targets,
      missing_target_summaries: missingTargetSummaries,
      skipped_after_failure: skippedAfterFailure,
      start_time: summarized.startTime,
      end_time: summarized.endTime,
      ...durationFieldsForJSON(summarized.aggregate, {
        wall_duration_ms: summarized.wallDurationMs,
        critical_path_wall_duration_ms: summarized.criticalPathWallDurationMs,
      }),
      accounting_modes: summarized.accountingModes,
      counts: countsForJSON(summarized.aggregate),
      timing_failures: summarized.timingFailures,
      teardown_status: teardownStatus(
        summarized.aggregate.teardown_duration_ms,
        summarized.teardownFailures,
      ),
      teardown_failures: summarized.teardownFailures,
    };
  });
}

function writeSummaryGroupLines(stream, label, summaryGroups) {
  for (const group of summaryGroups) {
    const missing =
      group.missing_target_summaries.length > 0
        ? ` missing=${group.missing_target_summaries.join(",")}`
        : "";
    const skipped =
      group.skipped_after_failure.length > 0
        ? ` skipped_after_failure=${group.skipped_after_failure.join(",")}`
        : "";
    stream.write(
      `[GROUP] ${label} ${group.name} targets=${group.targets.join(",")} status=${group.status} ${formatDurationFields(group.wall_duration_ms, group.executed_duration_ms, group.logical_duration_ms, group.critical_path_wall_duration_ms, group.teardown_duration_ms)} ${formatAccountingModeFields(group.accounting_modes)}${missing}${skipped}\n`,
    );
  }
}

function writeSharedExecutionGroupLines(stream, label, sharedExecutionGroups) {
  for (const group of sharedExecutionGroups) {
    stream.write(
      `[SHARED] ${label} ${group.name} status=${group.status} wall=${formatDuration(group.wall_duration_ms)} exec=${formatDuration(group.executed_duration_ms)} reports=${group.reports}\n`,
    );
  }
}

function findSlowestTarget(targetSummaries) {
  return targetSummaries.reduce((current, summary) => {
    const view = targetSummaryAccountingView(summary);
    const durationMs = clampDurationMs(
      view.critical_path_wall_duration_ms ??
        view.wall_duration_ms ??
        view.logical_duration_ms ??
        0,
    );
    if (!current || durationMs > current.critical_path_wall_duration_ms) {
      return {
        target: summary.target,
        critical_path_wall_duration_ms: durationMs,
        basis: "critical_path_wall_duration_ms",
      };
    }
    return current;
  }, null);
}

function findSlowestLifecycleBucket(targetSummaries) {
  return targetSummaries.reduce((current, summary) => {
    const bucket = targetSummaryAccountingView(summary).slowest_lifecycle_bucket;
    if (!bucket) {
      return current;
    }
    const candidate = {
      target: summary.target,
      name: bucket.name,
      duration_ms: clampDurationMs(bucket.duration_ms ?? 0),
    };
    if (!current || candidate.duration_ms > current.duration_ms) {
      return candidate;
    }
    return current;
  }, null);
}

function handleRunSummary(args) {
  const { label, requestedStatus, completedText, totalText, abortedAfter, targets, summaryGroups, skippedAfterFailure, quietSuccess } =
    parseRunSummaryArgs(args);
  const completedTargets = Number.parseInt(completedText, 10) || 0;
  const totalTargets = Number.parseInt(totalText, 10) || 0;
  const skippedAfterFailureSet = new Set(skippedAfterFailure);
  const missingTargetSummaries = [];
  const targetSummaries = [];

  for (const target of targets) {
    const summary = loadTargetSummary(target);
    if (!summary) {
      if (skippedAfterFailureSet.has(target)) {
        continue;
      }
      missingTargetSummaries.push(target);
      continue;
    }
    targetSummaries.push(summary);
  }
  const summarized = summarizeTargetSummaries(
    targetSummaries,
    missingTargetSummaries,
    requestedStatus,
  );
  const aggregate = summarized.aggregate;
  const accountingModes = summarized.accountingModes;
  const wallDurationMs = summarized.wallDurationMs;
  const criticalPathWallDurationMs = summarized.criticalPathWallDurationMs;
  const renderedSummaryGroups = buildSummaryGroups(summaryGroups, skippedAfterFailureSet);
  const sharedExecutionGroups = buildSharedExecutionGroups();
  const failed =
    summarized.failed ||
    renderedSummaryGroups.some((group) => group.status !== "pass") ||
    sharedExecutionGroups.some((group) => group.status !== "pass");
  const slowestTarget = findSlowestTarget(targetSummaries);
  const slowestLifecycleBucket = findSlowestLifecycleBucket(targetSummaries);
  const runFixture = combineFixtureSummaries(
    label,
    null,
    targetSummaries.map((summary) => ({
      fixture: targetSummaryAccountingView(summary, summary.target).fixture,
    })),
  );

  const runSummary = {
    schema_id: runSummarySchemaID,
    label,
    status: failed ? "fail" : "pass",
    completed_targets: `${completedTargets}/${totalTargets}`,
    aborted_after: abortedAfter === "-" ? "" : abortedAfter,
    start_time: summarized.startTime,
    end_time: summarized.endTime,
    ...durationFieldsForJSON(aggregate, {
      wall_duration_ms: wallDurationMs,
      critical_path_wall_duration_ms: criticalPathWallDurationMs,
    }),
    accounting_modes: accountingModes,
    counts: countsForJSON(aggregate),
    slowest_target: slowestTarget,
    slowest_lifecycle_bucket: slowestLifecycleBucket,
    timing_failures: summarized.timingFailures,
    teardown_status: teardownStatus(aggregate.teardown_duration_ms, summarized.teardownFailures),
    teardown_failures: summarized.teardownFailures,
    fixture: runFixture,
    artifacts: {
      dir: relToRepo(path.join(resultsRoot, runId)),
    },
    targets,
    target_summaries: targetSummaries,
    missing_target_summaries: missingTargetSummaries,
    skipped_after_failure: skippedAfterFailure,
    summary_groups: renderedSummaryGroups,
    shared_execution_groups: sharedExecutionGroups,
  };
  writeJson(path.join(resultsRoot, runId, "run-summary.json"), runSummary);

  if (!failed) {
    if (quietSuccess && quietOutputMode()) {
      return 0;
    }
    process.stdout.write(
      `[PASS] ${label} completed_targets=${completedTargets}/${totalTargets} phases=${aggregate.phases} tests=${aggregate.tests} authoritative=${aggregate.authoritative} support=${aggregate.support} unmapped=${aggregate.unmapped} ${formatDurationFields(wallDurationMs, aggregate.executed_duration_ms, aggregate.logical_duration_ms, criticalPathWallDurationMs, aggregate.teardown_duration_ms)} ${formatAccountingModeFields(accountingModes)} slowest_target=${slowestTarget ? `${slowestTarget.target}(${formatDuration(slowestTarget.critical_path_wall_duration_ms)})` : "none"} slowest_lifecycle_bucket=${formatTargetBucketSummary(slowestLifecycleBucket)} artifacts=${relToRepo(path.join(resultsRoot, runId))}\n`,
    );
    writeFixtureLine(process.stdout, runFixture);
    writeSummaryGroupLines(process.stdout, label, renderedSummaryGroups);
    writeSharedExecutionGroupLines(process.stdout, label, sharedExecutionGroups);
    return 0;
  }

  process.stderr.write(
    `[FAIL] ${label} completed_targets=${completedTargets}/${totalTargets} aborted_after=${abortedAfter === "-" ? "-" : abortedAfter} phases=${aggregate.phases} tests=${aggregate.tests} failed=${aggregate.failed} authoritative_failed=${aggregate.authoritative_failed} support_failed=${aggregate.support_failed} unmapped_failed=${aggregate.unmapped_failed} non_test_failed=${aggregate.non_test_failed} ${formatDurationFields(wallDurationMs, aggregate.executed_duration_ms, aggregate.logical_duration_ms, criticalPathWallDurationMs, aggregate.teardown_duration_ms)} ${formatAccountingModeFields(accountingModes)} slowest_target=${slowestTarget ? `${slowestTarget.target}(${formatDuration(slowestTarget.critical_path_wall_duration_ms)})` : "none"} slowest_lifecycle_bucket=${formatTargetBucketSummary(slowestLifecycleBucket)} artifacts=${relToRepo(path.join(resultsRoot, runId))}\n`,
  );
  writeFixtureLine(process.stderr, runFixture);
  writeSummaryGroupLines(process.stderr, label, renderedSummaryGroups);
  writeSharedExecutionGroupLines(process.stderr, label, sharedExecutionGroups);
  return 1;
}

function handleGoJSONStream() {
  flushGoJSONStream(readFileSync(0, "utf8"), true);
  return 0;
}

function flushGoJSONStream(buffer, flushAll) {
  const lines = buffer.split(/\r?\n/);
  const pending = flushAll ? "" : lines.pop() ?? "";
  for (const line of lines) {
    if (!line.trim()) {
      continue;
    }
    try {
      const entry = JSON.parse(line);
      if (typeof entry.Output === "string") {
        process.stdout.write(entry.Output);
      }
    } catch {
      // Go's -json stream is expected on stdout; ignore malformed lines.
    }
  }
  if (flushAll && pending.trim()) {
    try {
      const entry = JSON.parse(pending);
      if (typeof entry.Output === "string") {
        process.stdout.write(entry.Output);
      }
    } catch {
      // Ignore trailing malformed output.
    }
  }
  return pending;
}

function createInventoryItem({ coverage, phase, id, owner, name }) {
  return {
    coverage,
    phase,
    id: id ?? "",
    package_or_file: owner,
    symbol_or_title: name,
  };
}

function finalizeShellPhase(context, stdoutLog, stderrLog, details) {
  removeEmptyArtifact(stdoutLog);
  removeEmptyArtifact(stderrLog);

  writePhaseArtifacts(context, {
    ...details,
    artifacts: {
      stdout_log: existsSync(stdoutLog) ? stdoutLog : "",
      stderr_log: existsSync(stderrLog) ? stderrLog : "",
    },
  });

  if (details.status === "pass") {
    return 0;
  }

  for (const dossier of details.dossiers) {
    printBlock(`failure: ${context.label}`, dossier);
  }
  return 1;
}

function handleShellPhase() {
  const context = createBasePhaseContext("shell");
  const stdoutLog = requiredEnv("CARTULARY_PHASE_STDOUT_LOG");
  const stderrLog = requiredEnv("CARTULARY_PHASE_STDERR_LOG");

  if (context.exitStatus === 0) {
    return finalizeShellPhase(context, stdoutLog, stderrLog, {
      status: "pass",
      phase: inferPhaseFromText(context.label),
      counts: {
        ...createCounts(),
      },
      owners: [],
      inventory: [],
      dossiers: [],
    });
  }

  const messageBase =
    firstActionableLine(splitLogLines(stderrLog)) ||
    firstActionableLine(splitLogLines(stdoutLog)) ||
    `command exited with status ${context.exitStatus}`;
  const failureNote = optionalEnv("CARTULARY_PHASE_FAILURE_NOTE");
  const message =
    failureNote === ""
      ? messageBase
      : `${messageBase} | remediation: ${failureNote}`;
  return finalizeShellPhase(context, stdoutLog, stderrLog, {
    status: "fail",
    phase: inferPhaseFromText(context.label),
    counts: {
      ...createCounts(),
      failed: 1,
      non_test: 1,
      non_test_failed: 1,
    },
    owners: [],
    inventory: [],
    dossiers: [
      {
        coverage: "non_test",
        phase: inferPhaseFromText(context.label),
        id: "",
        runner: "shell",
        package_or_file: "(shell command)",
        symbol_or_title: "(shell command)",
        message,
        reproduce: context.command,
        raw: renderRawList([stdoutLog, stderrLog]),
      },
    ],
  });
}

function readGoEvents(logFile) {
  const events = [];
  for (const rawLine of splitLogLines(logFile)) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    events.push(JSON.parse(line));
  }
  return events;
}

function classifyGoTest(importPath, testName, phaseLabel) {
  const manifestIndex = loadManifestIndex();
  const authoritative = manifestIndex.authoritativeGo.get(`${importPath}::${testName}`);
  if (authoritative) {
    return {
      coverage: "authoritative",
      phase: authoritative.phase,
      id: authoritative.id,
      owner: authoritative.package,
    };
  }

  const inferredPhase = inferPhaseFromText(testName) || inferPhaseFromText(phaseLabel);
  const support =
    /^TestSupportPhase\d+_/.test(testName) ||
    /ProcessSmoke/.test(testName) ||
    /\bsupport\b/i.test(phaseLabel) ||
    /\bsmoke\b/i.test(phaseLabel);

  return {
    coverage: support ? "support" : "unmapped",
    phase: inferredPhase,
    id: "",
    owner: toRepoRelativePackage(importPath),
  };
}

function classifyGoPackageFailure(importPath, phaseLabel) {
  const manifestPhase = optionalEnv("CARTULARY_MANIFEST_PHASE");
  const manifestCoverage = optionalEnv("CARTULARY_MANIFEST_COVERAGE");
  if (manifestPhase !== "" && manifestCoverage !== "") {
    return {
      coverage: manifestCoverage,
      phase: manifestPhase,
      id: "",
      owner: toRepoRelativePackage(importPath),
    };
  }

  const inferredPhase = inferPhaseFromText(phaseLabel);
  const support = /\bsupport\b/i.test(phaseLabel) || /\bsmoke\b/i.test(phaseLabel);
  return {
    coverage: support ? "support" : "unmapped",
    phase: inferredPhase,
    id: "",
    owner: toRepoRelativePackage(importPath),
  };
}

function createGoSelection({ manifestAware }) {
  const packagePatterns = optionalLines("CARTULARY_GO_PACKAGE_PATTERNS");
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";

  if (manifestAware && reportSlice) {
    const phase = requiredEnv("CARTULARY_MANIFEST_PHASE");
    const section = requiredEnv("CARTULARY_MANIFEST_SECTION");
    const coverage = requiredEnv("CARTULARY_MANIFEST_COVERAGE");
    const executionDependency = optionalEnv("CARTULARY_MANIFEST_EXECUTION_DEPENDENCY");
    const executionFamily = optionalEnv("CARTULARY_EXECUTION_FAMILY");
    const entries = selectGoManifestEntries(
      phase,
      section,
      coverage,
      executionDependency,
      executionFamily,
      packagePatterns,
    );
    const selectedTests = new Set();
    const selectedPackages = new Set();
    for (const entry of entries) {
      const symbols = entry.symbol !== undefined ? [entry.symbol] : entry.symbols;
      selectedPackages.add(toGoImportPath(entry.package));
      for (const symbol of symbols) {
        selectedTests.add(`${toGoImportPath(entry.package)}::${symbol}`);
      }
    }
    return {
      matchesTest(importPath, testName) {
        return selectedTests.has(`${importPath}::${testName}`);
      },
      matchesPackage(importPath) {
        return selectedPackages.has(importPath);
      },
    };
  }

  const testRegexSource = optionalEnv("CARTULARY_GO_TEST_REGEX");
  if (testRegexSource === "" && packagePatterns.length === 0) {
    return null;
  }
  const testRegex = testRegexSource === "" ? null : new RegExp(testRegexSource);
  return {
    matchesTest(importPath, testName) {
      if (
        packagePatterns.length > 0 &&
        !packagePatterns.some((pattern) => packageMatchesPattern(toRepoRelativePackage(importPath), pattern))
      ) {
        return false;
      }
      return testRegex === null ? true : testRegex.test(testName);
    },
    matchesPackage(importPath) {
      if (packagePatterns.length === 0) {
        return true;
      }
      return packagePatterns.some((pattern) =>
        packageMatchesPattern(toRepoRelativePackage(importPath), pattern),
      );
    },
  };
}

function summarizeGoRun(logFile, phaseLabel, exitStatus, selection = null) {
  const events = readGoEvents(logFile);
  const topLevel = new Map();
  const packageOutputs = new Map();
  const packageFailures = new Map();

  for (const entry of events) {
    const pkg = typeof entry.Package === "string" ? entry.Package : "";
    const testName = typeof entry.Test === "string" ? entry.Test : "";
    if (pkg !== "" && typeof entry.Output === "string") {
      if (testName === "" || testName.includes("/")) {
        if (!packageOutputs.has(pkg)) {
          packageOutputs.set(pkg, []);
        }
        packageOutputs.get(pkg).push(entry.Output);
      }
    }
    if (testName === "" || testName.includes("/")) {
      if (pkg !== "" && entry.Action === "fail") {
        packageFailures.set(pkg, true);
      }
      continue;
    }
    if (
      !testName.startsWith("Test") &&
      !testName.startsWith("Benchmark") &&
      !testName.startsWith("Fuzz")
    ) {
      continue;
    }
    const key = `${pkg}::${testName}`;
    if (!topLevel.has(key)) {
      topLevel.set(key, {
        package: pkg,
        test: testName,
        status: "",
        outputs: [],
      });
    }
    const current = topLevel.get(key);
    if (typeof entry.Output === "string") {
      current.outputs.push(entry.Output);
    }
    if (entry.Action === "run") {
      if (current.status === "") {
        current.status = "run";
      }
      continue;
    }
    if (["pass", "fail", "skip"].includes(entry.Action)) {
      current.status = entry.Action;
    }
  }

  const cases = [];
  const dossiers = [];
  const owners = new Set();
  const counts = createCounts();
  let passedCount = 0;
  let skippedCount = 0;
  let incompleteCount = 0;

  for (const testCase of topLevel.values()) {
    if (selection && !selection.matchesTest(testCase.package, testCase.test)) {
      continue;
    }
    const classification = classifyGoTest(testCase.package, testCase.test, phaseLabel);
    const owner = classification.owner;
    owners.add(owner);
    if (testCase.status !== "skip") {
      counts.tests += 1;
      counts[classification.coverage] += 1;
    }
    if (testCase.status === "pass") {
      passedCount += 1;
      cases.push({
        status: "pass",
        classification,
        owner,
        name: testCase.test,
      });
      continue;
    }
    if (testCase.status === "skip") {
      skippedCount += 1;
      continue;
    }
    if (testCase.status === "fail") {
      counts.failed += 1;
      counts[`${classification.coverage}_failed`] += 1;
      dossiers.push({
        coverage: classification.coverage,
        phase: classification.phase,
        id: classification.id,
        runner: "go_test",
        package_or_file: owner,
        symbol_or_title: testCase.test,
        message:
          firstGoActionableLine(testCase.outputs) || `${testCase.test} failed`,
        reproduce: `go test ${owner} -run '^${escapeRegex(testCase.test)}$'`,
        raw: relToRepo(logFile),
      });
      continue;
    }
    incompleteCount += 1;
  }

  for (const pkg of packageFailures.keys()) {
    if (selection && !selection.matchesPackage(pkg)) {
      continue;
    }
    const classification = classifyGoPackageFailure(pkg, phaseLabel);
    const owner = classification.owner;
    owners.add(owner);
    if ([...topLevel.values()].some((entry) => entry.package === pkg && entry.status === "fail")) {
      continue;
    }
    counts.failed += 1;
    counts[`${classification.coverage}_failed`] += 1;
    dossiers.push({
      coverage: classification.coverage,
      phase: classification.phase,
      id: "",
      runner: "go_test",
      package_or_file: owner,
      symbol_or_title: "(package setup)",
      message:
        firstGoActionableLine(packageOutputs.get(pkg) ?? []) ||
        `package ${owner} failed before a top-level test was attributed`,
      reproduce: `go test ${owner}`,
      raw: relToRepo(logFile),
    });
  }

  if (exitStatus === 0 && (passedCount === 0 || skippedCount > 0 || incompleteCount > 0)) {
    const coverage = /\bsupport\b/i.test(phaseLabel) ? "support" : "unmapped";
    const message =
      passedCount === 0 && skippedCount === 0 && incompleteCount === 0
        ? coverage === "support"
          ? "support phase matched zero tests"
          : "phase matched zero tests"
        : `go test inventory requires top-level pass: skipped=${skippedCount} incomplete=${incompleteCount}`;
    dossiers.push({
      coverage,
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "go_test",
      package_or_file: "(phase selection)",
      symbol_or_title: "(top-level selection)",
      message,
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
      raw: relToRepo(logFile),
    });
    counts.failed += 1;
    counts[`${coverage}_failed`] += 1;
  }

  counts.packages = owners.size;

  return {
    counts,
    owners: Array.from(owners).sort(),
    inventory: cases
      .filter((entry) => entry.classification.coverage !== "unmapped" || entry.classification.id)
      .map((entry) =>
        createInventoryItem({
          coverage: entry.classification.coverage,
          phase: entry.classification.phase,
          id: entry.classification.id,
          owner: entry.owner,
          name: entry.name,
        }),
      ),
    dossiers,
  };
}

function selectGoManifestEntries(
  phase,
  section,
  coverage,
  executionDependency,
  executionFamily,
  packagePatterns,
) {
  const { manifest } = loadManifest(repoRoot, phase);
  return collectEntries(manifest).filter((entry) => {
    if (entry.runner !== "go_test" || entry.section !== section || entry.coverage !== coverage) {
      return false;
    }
    if (executionDependency && entry.execution_dependency !== executionDependency) {
      return false;
    }
    if (executionFamily && entry.execution_family !== executionFamily) {
      return false;
    }
    return packagePatterns.some((pattern) => packageMatchesPattern(entry.package, pattern));
  });
}

function packageMatchesPattern(pkg, pattern) {
  if (pattern.endsWith("/...")) {
    const prefix = pattern.slice(0, -4);
    return pkg === prefix || pkg.startsWith(`${prefix}/`);
  }
  return pkg === pattern;
}

function evaluateGoManifest(summary) {
  const phase = requiredEnv("CARTULARY_MANIFEST_PHASE");
  const section = requiredEnv("CARTULARY_MANIFEST_SECTION");
  const coverage = requiredEnv("CARTULARY_MANIFEST_COVERAGE");
  const executionDependency = optionalEnv("CARTULARY_MANIFEST_EXECUTION_DEPENDENCY");
  const executionFamily = optionalEnv("CARTULARY_EXECUTION_FAMILY");
  const packagePatterns = optionalLines("CARTULARY_GO_PACKAGE_PATTERNS");
  const entries = selectGoManifestEntries(
    phase,
    section,
    coverage,
    executionDependency,
    executionFamily,
    packagePatterns,
  );
  const expectedByID = new Map();
  for (const entry of entries) {
    const symbols = entry.symbol !== undefined ? [entry.symbol] : entry.symbols;
    expectedByID.set(entry.id, {
      id: entry.id,
      symbols,
    });
  }
  const passedTests = new Set(
    summary.inventory.map((item) => `${toGoImportPath(item.package_or_file)}::${item.symbol_or_title}`),
  );
  const missingIDs = [];
  for (const entry of entries) {
    const symbols = entry.symbol !== undefined ? [entry.symbol] : entry.symbols;
    const missing = symbols.some((symbol) => !passedTests.has(`${toGoImportPath(entry.package)}::${symbol}`));
    if (missing) {
      missingIDs.push(entry.id);
    }
  }
  const expectedIDs = new Set(entries.map((entry) => entry.id));
  const unexpectedIDs = [];
  for (const item of summary.inventory) {
    if (item.coverage !== "authoritative" || item.id === "" || expectedIDs.has(item.id)) {
      continue;
    }
    unexpectedIDs.push(item.id);
  }
  return {
    phase,
    missingIDs: [...new Set(missingIDs)].sort(),
    unexpectedIDs: [...new Set(unexpectedIDs)].sort(),
    forbiddenIDFiles: Array.from(loadManifestIndex().forbiddenFilesByPhase.get(phase) ?? []).sort(),
  };
}

function handleGoPhase({ manifestAware }) {
  const context = createBasePhaseContext("go_test");
  const runnerLog = requiredEnv("CARTULARY_PHASE_RUNNER_LOG");
  const stderrLog = optionalEnv("CARTULARY_PHASE_STDERR_LOG");
  removeEmptyArtifact(stderrLog);

  const summary = summarizeGoRun(
    runnerLog,
    context.label,
    context.exitStatus,
    createGoSelection({ manifestAware }),
  );
  let status = context.exitStatus === 0 && summary.dossiers.length === 0 ? "pass" : "fail";
  let manifestMismatch = null;
  let manifestSummary = null;

  if (manifestAware && context.exitStatus === 0 && summary.dossiers.length === 0) {
    const verification = evaluateGoManifest(summary);
    manifestSummary = {
      missing_ids: verification.missingIDs,
      unexpected_ids: verification.unexpectedIDs,
    };
    if (verification.missingIDs.length > 0 || verification.unexpectedIDs.length > 0) {
      status = "mismatch";
      manifestMismatch = {
        missing_ids: verification.missingIDs,
        unexpected_ids: verification.unexpectedIDs,
        forbidden_id_files: verification.forbiddenIDFiles,
        raw: relToRepo(runnerLog),
      };
    }
  }

  writePhaseArtifacts(context, {
    status,
    phase: inferPhaseFromText(context.label),
    counts: summary.counts,
    owners: summary.owners,
    inventory: summary.inventory,
    dossiers: summary.dossiers,
    manifestSummary,
    manifestMismatch,
    artifacts: {
      runner_jsonl: runnerLog,
      stderr_log: existsSync(stderrLog) ? stderrLog : "",
    },
  });

  if (status === "pass") {
    return 0;
  }

  if (status === "mismatch") {
    printBlock(`manifest mismatch: ${context.label}`, {
      missing_ids: renderList(manifestMismatch.missing_ids),
      unexpected_ids: renderList(manifestMismatch.unexpected_ids),
      forbidden_id_files: renderList(manifestMismatch.forbidden_id_files),
      raw: manifestMismatch.raw,
    });
    return 1;
  }

  for (const dossier of summary.dossiers) {
    printBlock(`failure: ${context.label}`, {
      ...dossier,
      raw: renderRawList([runnerLog, stderrLog]),
    });
  }
  return 1;
}

function classifyVitestCase(ownerPath, title, phaseLabel) {
  const manifestFile = vitestOwnerToSelectionFile(ownerPath);
  const authoritative = loadManifestIndex().authoritativeVitest.get(`${manifestFile}::${title}`);
  if (authoritative) {
    return {
      coverage: "authoritative",
      phase: authoritative.phase,
      id: authoritative.id,
      owner: ownerPath,
    };
  }
  const support =
    ownerPath.includes(".support.") ||
    supportNamedTitle(title) ||
    isForbiddenFile(ownerPath, inferPhaseFromText(ownerPath) || inferPhaseFromText(title)) ||
    /\bsupport\b/i.test(phaseLabel);
  return {
    coverage: support ? "support" : "unmapped",
    phase: inferPhaseFromText(ownerPath) || inferPhaseFromText(title) || inferPhaseFromText(phaseLabel),
    id: "",
    owner: ownerPath,
  };
}

function normalizeVitestOwnerPath(filePath) {
  return relToRepo(filePath);
}

function normalizeVitestSelectionFile(filePath) {
  const relative = relToRepo(filePath);
  if (relative === "") {
    return "";
  }
  if (relative.startsWith("apps/web/")) {
    return relative;
  }
  return normalizePath(path.join("apps/web", relative));
}

function vitestOwnerToSelectionFile(ownerPath) {
  if (ownerPath === "") {
    return "";
  }
  if (ownerPath.startsWith("apps/web/")) {
    return ownerPath;
  }
  return normalizePath(path.join("apps/web", ownerPath));
}

function vitestOwnerToReproducePath(ownerPath) {
  if (ownerPath.startsWith("apps/web/")) {
    return ownerPath.slice("apps/web/".length);
  }
  if (ownerPath.startsWith("packages/")) {
    return `../../${ownerPath}`;
  }
  return ownerPath;
}

function renderVitestReproduceCommand(ownerPath, title = "") {
  const reproducePath = vitestOwnerToReproducePath(ownerPath);
  if (title === "") {
    return `pnpm --dir apps/web exec vitest run ${reproducePath}`;
  }
  return `pnpm --dir apps/web exec vitest run ${reproducePath} -t '${escapeSingleQuotes(title)}$'`;
}

function createVitestSelection({ manifestAware }) {
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";

  if (manifestAware && reportSlice) {
    const phase = requiredEnv("CARTULARY_MANIFEST_PHASE");
    const coverage = requiredEnv("CARTULARY_MANIFEST_COVERAGE");
    const executionDependency = optionalEnv("CARTULARY_MANIFEST_EXECUTION_DEPENDENCY");
    const entries = selectVitestManifestEntries(phase, coverage, executionDependency);
    const selected = new Set(
      entries.map((entry) => `${entry.file}::${entry.title}`),
    );
    const selectedFiles = new Set(entries.map((entry) => normalizePath(entry.file)));
    return {
      matches(ownerPath, title) {
        return selected.has(`${vitestOwnerToSelectionFile(ownerPath)}::${title}`);
      },
      matchesFile(ownerPath) {
        return selectedFiles.has(vitestOwnerToSelectionFile(ownerPath));
      },
      classifyFileFailure(ownerPath) {
        if (!selectedFiles.has(vitestOwnerToSelectionFile(ownerPath))) {
          return null;
        }
        return {
          coverage,
          phase,
          id: "",
          owner: ownerPath,
        };
      },
    };
  }

  const selectedFiles = new Set(
    optionalLines("CARTULARY_VITEST_FILES").map((value) => normalizeVitestSelectionFile(value)),
  );
  const selectedTitles = optionalSetFromLines("CARTULARY_VITEST_TITLES");
  if (selectedFiles.size === 0 && selectedTitles.size === 0) {
    return null;
  }
  return {
    matches(ownerPath, title) {
      if (
        selectedFiles.size > 0 &&
        !selectedFiles.has(vitestOwnerToSelectionFile(ownerPath))
      ) {
        return false;
      }
      if (selectedTitles.size > 0 && !selectedTitles.has(title)) {
        return false;
      }
      return true;
    },
    matchesFile(ownerPath) {
      if (selectedFiles.size === 0) {
        return selectedTitles.size === 0;
      }
      return selectedFiles.has(vitestOwnerToSelectionFile(ownerPath));
    },
    classifyFileFailure() {
      return null;
    },
  };
}

function isVitestFileResult(value) {
  return Boolean(
    value &&
      typeof value === "object" &&
      typeof value.name === "string" &&
      Array.isArray(value.assertionResults),
  );
}

function collectVitestFileResults(report) {
  const fileResults = [];
  const visited = new Set();
  appendVitestFileResults(report, fileResults, visited);
  return fileResults;
}

function appendVitestFileResults(value, fileResults, visited) {
  if (!value || typeof value !== "object") {
    return;
  }
  if (isVitestFileResult(value)) {
    fileResults.push(value);
    return;
  }
  if (visited.has(value)) {
    return;
  }
  visited.add(value);
  if (Array.isArray(value)) {
    for (const entry of value) {
      appendVitestFileResults(entry, fileResults, visited);
    }
    return;
  }
  for (const key of ["testResults", "projectResults", "projects", "results"]) {
    if (!Array.isArray(value[key])) {
      continue;
    }
    for (const entry of value[key]) {
      appendVitestFileResults(entry, fileResults, visited);
    }
  }
}

function findVitestAuthoritativeFileEntry(ownerPath, phaseLabel) {
  const manifestFile = vitestOwnerToSelectionFile(ownerPath);
  const inferredPhase = inferPhaseFromText(ownerPath) || inferPhaseFromText(phaseLabel);
  let fallback = null;

  for (const entry of loadManifestIndex().authoritativeVitest.values()) {
    if (normalizePath(entry.file) !== manifestFile) {
      continue;
    }
    if (!fallback) {
      fallback = entry;
    }
    if (inferredPhase && entry.phase === inferredPhase) {
      return entry;
    }
  }

  return fallback;
}

function classifyVitestFileFailure(ownerPath, phaseLabel, selection = null) {
  const selected = selection?.classifyFileFailure?.(ownerPath);
  if (selected) {
    return selected;
  }

  const authoritative = findVitestAuthoritativeFileEntry(ownerPath, phaseLabel);
  if (authoritative) {
    return {
      coverage: "authoritative",
      phase: authoritative.phase,
      id: "",
      owner: ownerPath,
    };
  }

  const inferredPhase = inferPhaseFromText(ownerPath) || inferPhaseFromText(phaseLabel);
  const support =
    ownerPath.includes(".support.") ||
    isForbiddenFile(ownerPath, inferredPhase) ||
    /\bsupport\b/i.test(phaseLabel);
  return {
    coverage: support ? "support" : "unmapped",
    phase: inferredPhase,
    id: "",
    owner: ownerPath,
  };
}

function summarizeVitestRun(reportFile, phaseLabel, selection = null) {
  const report = JSON.parse(readFileSync(reportFile, "utf8"));
  const owners = new Set();
  const inventory = [];
  const dossiers = [];
  const counts = createCounts();

  for (const fileResult of collectVitestFileResults(report)) {
    const ownerPath = normalizeVitestOwnerPath(fileResult.name ?? "");
    const assertions = fileResult.assertionResults ?? [];
    const executedAssertions = assertions.filter(
      (assertion) => assertion.status !== "skipped",
    );
    if (executedAssertions.length === 0 && fileResult.status === "failed") {
      if (selection && !selection.matchesFile(ownerPath)) {
        continue;
      }
      const classification = classifyVitestFileFailure(
        ownerPath,
        phaseLabel,
        selection,
      );
      owners.add(classification.owner);
      counts.failed += 1;
      counts[`${classification.coverage}_failed`] += 1;
      dossiers.push({
        coverage: classification.coverage,
        phase: classification.phase,
        id: classification.id,
        runner: "vitest",
        package_or_file: classification.owner,
        symbol_or_title: "(suite load)",
        message:
          fileResult.message?.split("\n")[0]?.trim() ||
          `test file ${classification.owner} failed before a top-level test was attributed`,
        reproduce: renderVitestReproduceCommand(classification.owner),
        raw: relToRepo(reportFile),
      });
      continue;
    }
    for (const assertion of assertions) {
      if (assertion.status === "skipped") {
        continue;
      }
      if (selection && !selection.matches(ownerPath, assertion.title ?? "")) {
        continue;
      }
      const classification = classifyVitestCase(ownerPath, assertion.title ?? "", phaseLabel);
      owners.add(classification.owner);
      counts.tests += 1;
      counts[classification.coverage] += 1;
      if (assertion.status === "passed") {
        inventory.push(
          createInventoryItem({
            coverage: classification.coverage,
            phase: classification.phase,
            id: classification.id,
            owner: classification.owner,
            name: assertion.title ?? "(missing title)",
          }),
        );
        continue;
      }
      counts.failed += 1;
      counts[`${classification.coverage}_failed`] += 1;
      const failureMessage = Array.isArray(assertion.failureMessages) ? assertion.failureMessages[0] ?? "" : "";
      dossiers.push({
        coverage: classification.coverage,
        phase: classification.phase,
        id: classification.id,
        runner: "vitest",
        package_or_file: classification.owner,
        symbol_or_title: assertion.title ?? "(missing title)",
        message: failureMessage.split("\n")[0] || `${assertion.title ?? "vitest assertion"} failed`,
        reproduce: renderVitestReproduceCommand(
          classification.owner,
          (assertion.title ?? "").trim(),
        ),
        raw: relToRepo(reportFile),
      });
    }
  }

  if (counts.tests === 0 && dossiers.length === 0) {
    dossiers.push({
      coverage: "unmapped",
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "vitest",
      package_or_file: "(vitest selection)",
      symbol_or_title: "(vitest selection)",
      message: "phase matched zero tests",
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
      raw: relToRepo(reportFile),
    });
    counts.failed += 1;
    counts.unmapped_failed += 1;
  }

  counts.packages = owners.size;

  return {
    report,
    counts,
    owners: Array.from(owners).sort(),
    inventory,
    dossiers,
  };
}

function selectVitestManifestEntries(phase, coverage, executionDependency) {
  const { manifest } = loadManifest(repoRoot, phase);
  return collectEntries(manifest).filter((entry) => {
    if (entry.runner !== "vitest" || entry.section !== "unit" || entry.coverage !== coverage) {
      return false;
    }
    if (executionDependency && entry.execution_dependency !== executionDependency) {
      return false;
    }
    return true;
  });
}

function evaluateVitestManifest(summary) {
  const phase = requiredEnv("CARTULARY_MANIFEST_PHASE");
  const coverage = requiredEnv("CARTULARY_MANIFEST_COVERAGE");
  const executionDependency = optionalEnv("CARTULARY_MANIFEST_EXECUTION_DEPENDENCY");
  const entries = selectVitestManifestEntries(phase, coverage, executionDependency);
  const executedKeys = new Set(
    summary.inventory
      .filter((item) => item.coverage === "authoritative")
      .map((item) => `${item.package_or_file}::${item.symbol_or_title}`),
  );

  const missingIDs = entries
    .filter((entry) => !executedKeys.has(`${entry.file}::${entry.title}`))
    .map((entry) => entry.id)
    .sort();
  const expectedIDs = new Set(entries.map((entry) => entry.id));
  const unexpectedIDs = summary.inventory
    .filter((item) => item.coverage === "authoritative" && item.id && !expectedIDs.has(item.id))
    .map((item) => item.id)
    .sort();

  return {
    phase,
    missingIDs,
    unexpectedIDs,
    forbiddenIDFiles: Array.from(loadManifestIndex().forbiddenFilesByPhase.get(phase) ?? []).sort(),
  };
}

function handleVitestPhase({ manifestAware }) {
  const context = createBasePhaseContext("vitest");
  const reportFile = requiredEnv("CARTULARY_PHASE_RUNNER_LOG");
  const stderrLog = optionalEnv("CARTULARY_PHASE_STDERR_LOG");
  const stdoutLog = optionalEnv("CARTULARY_PHASE_STDOUT_LOG");
  const watchdogLog = optionalEnv("CARTULARY_PHASE_WATCHDOG_LOG");
  removeEmptyArtifact(stderrLog);
  removeEmptyArtifact(stdoutLog);

  if (!existsSync(reportFile)) {
    const counts = createCounts();
    counts.failed += 1;
    counts.non_test += 1;
    counts.non_test_failed += 1;
    const message = existsSync(watchdogLog)
      ? "vitest watchdog timed out before runner.json was written"
      : "vitest runner.json was not written";
    const dossier = {
      coverage: "non_test",
      phase: inferPhaseFromText(context.label),
      id: "",
      runner: "vitest",
      package_or_file: "(vitest runner)",
      symbol_or_title: "(runner.json)",
      message,
      reproduce: context.command,
      raw: renderRawList([watchdogLog, stdoutLog, stderrLog]),
    };
    writePhaseArtifacts(context, {
      status: "fail",
      phase: inferPhaseFromText(context.label),
      counts,
      owners: [],
      inventory: [],
      dossiers: [dossier],
      artifacts: {
        runner_json: reportFile,
        stdout_log: existsSync(stdoutLog) ? stdoutLog : "",
        stderr_log: existsSync(stderrLog) ? stderrLog : "",
        watchdog_json: existsSync(watchdogLog) ? watchdogLog : "",
      },
    });
    printBlock(`failure: ${context.label}`, dossier);
    return 1;
  }

  const summary = summarizeVitestRun(
    reportFile,
    context.label,
    createVitestSelection({ manifestAware }),
  );
  let status = context.exitStatus === 0 && summary.dossiers.length === 0 ? "pass" : "fail";
  let manifestSummary = null;
  let manifestMismatch = null;

  if (manifestAware && context.exitStatus === 0 && summary.dossiers.length === 0) {
    const verification = evaluateVitestManifest(summary);
    manifestSummary = {
      missing_ids: verification.missingIDs,
      unexpected_ids: verification.unexpectedIDs,
    };
    if (verification.missingIDs.length > 0 || verification.unexpectedIDs.length > 0) {
      status = "mismatch";
      manifestMismatch = {
        missing_ids: verification.missingIDs,
        unexpected_ids: verification.unexpectedIDs,
        forbidden_id_files: verification.forbiddenIDFiles,
        raw: relToRepo(reportFile),
      };
    }
  }

  writePhaseArtifacts(context, {
    status,
    phase: inferPhaseFromText(context.label),
    counts: summary.counts,
    owners: summary.owners,
    inventory: summary.inventory,
    dossiers: summary.dossiers,
    manifestSummary,
    manifestMismatch,
    artifacts: {
      runner_json: reportFile,
      stdout_log: existsSync(stdoutLog) ? stdoutLog : "",
      stderr_log: existsSync(stderrLog) ? stderrLog : "",
      watchdog_json: existsSync(watchdogLog) ? watchdogLog : "",
    },
  });

  if (status === "pass") {
    return 0;
  }
  if (status === "mismatch") {
    printBlock(`manifest mismatch: ${context.label}`, {
      missing_ids: renderList(manifestMismatch.missing_ids),
      unexpected_ids: renderList(manifestMismatch.unexpected_ids),
      forbidden_id_files: renderList(manifestMismatch.forbidden_id_files),
      raw: manifestMismatch.raw,
    });
    return 1;
  }
  for (const dossier of summary.dossiers) {
    printBlock(`failure: ${context.label}`, {
      ...dossier,
      raw: renderRawList([reportFile, stdoutLog, stderrLog]),
    });
  }
  return 1;
}

function normalizePlaywrightFile(file) {
  const normalized = normalizePath(file);
  if (normalized.startsWith("apps/web/")) {
    return normalized;
  }
  return normalizePath(path.join("apps/web", "e2e", normalized));
}

function isForbiddenFile(file, phase) {
  if (!phase) {
    return false;
  }
  const files = loadManifestIndex().forbiddenFilesByPhase.get(phase);
  return files ? files.has(file) : false;
}

function classifyPlaywrightCase(file, title, phaseLabel) {
  const normalizedFile = normalizePlaywrightFile(file);
  const manifested = loadManifestIndex().manifestPlaywright.get(`${normalizedFile}::${title}`);
  if (manifested && manifested.coverage !== "authoritative") {
    return {
      coverage: "support",
      phase: manifested.phase,
      id: manifested.id,
      owner: normalizedFile,
    };
  }
  const authoritative = loadManifestIndex().authoritativePlaywright.get(`${normalizedFile}::${title}`);
  if (authoritative) {
    return {
      coverage: "authoritative",
      phase: authoritative.phase,
      id: authoritative.id,
      owner: normalizedFile,
    };
  }
  const inferredPhase =
    inferPhaseFromText(normalizedFile) || inferPhaseFromText(title) || inferPhaseFromText(phaseLabel);
  const support =
    normalizedFile.includes(".support.") ||
    isForbiddenFile(normalizedFile, inferredPhase) ||
    /\bsupport\b/i.test(phaseLabel) ||
    /\bsmoke\b/i.test(phaseLabel);
  return {
    coverage: support ? "support" : "unmapped",
    phase: inferredPhase,
    id: "",
    owner: normalizedFile,
  };
}

function flattenPlaywrightSuites(suites, specs = []) {
  for (const suite of suites ?? []) {
    flattenPlaywrightSuites(suite.suites, specs);
    for (const spec of suite.specs ?? []) {
      specs.push(spec);
    }
  }
  return specs;
}

function parsePlaywrightStartTime(value) {
  const parsed = Date.parse(value ?? "");
  return Number.isFinite(parsed) ? parsed : null;
}

function updateTimingWindow(window, startMs, durationMs) {
  if (startMs === null) {
    return;
  }
  const endMs = startMs + clampDurationMs(durationMs);
  if (window.startMs === null || startMs < window.startMs) {
    window.startMs = startMs;
  }
  if (window.endMs === null || endMs > window.endMs) {
    window.endMs = endMs;
  }
}

function timingWindowDurationMs(window) {
  if (window.startMs === null || window.endMs === null || window.endMs < window.startMs) {
    return 0;
  }
  return window.endMs - window.startMs;
}

function summarizePlaywrightTiming(specs, phase) {
  const files = new Map();
  const totalWindow = { startMs: null, endMs: null };
  let tests = 0;
  let executedDurationMs = 0;
  let hasDuration = false;
  let hasTimestamp = false;

  for (const spec of specs) {
    const file = normalizePlaywrightFile(spec.file ?? "");
    if (!files.has(file)) {
      files.set(file, {
        file,
        tests: 0,
        executed_duration_ms: 0,
        wall_duration_ms: 0,
        window: { startMs: null, endMs: null },
        hasTimestamp: false,
      });
    }
    const fileTiming = files.get(file);
    let executed = false;

    for (const test of spec.tests ?? []) {
      for (const result of test.results ?? []) {
        if (!result.status || result.status === "skipped") {
          continue;
        }
        executed = true;
        const durationMs = clampDurationMs(result.duration ?? 0);
        if (durationMs > 0) {
          hasDuration = true;
        }
        executedDurationMs += durationMs;
        fileTiming.executed_duration_ms += durationMs;

        const startMs = parsePlaywrightStartTime(result.startTime);
        if (startMs !== null) {
          hasTimestamp = true;
          fileTiming.hasTimestamp = true;
          updateTimingWindow(totalWindow, startMs, durationMs);
          updateTimingWindow(fileTiming.window, startMs, durationMs);
        }
      }
    }

    if (executed) {
      tests += 1;
      fileTiming.tests += 1;
    }
  }

  const fileSummaries = Array.from(files.values())
    .map((fileTiming) => ({
      file: fileTiming.file,
      tests: fileTiming.tests,
      executed_duration_ms: fileTiming.executed_duration_ms,
      wall_duration_ms: fileTiming.hasTimestamp
        ? timingWindowDurationMs(fileTiming.window)
        : fileTiming.executed_duration_ms,
    }))
    .sort((left, right) => left.file.localeCompare(right.file));

  return {
    phase,
    files: fileSummaries,
    tests,
    executed_duration_ms: executedDurationMs,
    wall_duration_ms: hasTimestamp ? timingWindowDurationMs(totalWindow) : executedDurationMs,
    start_time:
      hasTimestamp && totalWindow.startMs !== null ? new Date(totalWindow.startMs).toISOString() : "",
    end_time: hasTimestamp && totalWindow.endMs !== null ? new Date(totalWindow.endMs).toISOString() : "",
    source: hasTimestamp
      ? "playwright_result_timestamps"
      : hasDuration
        ? "playwright_result_durations"
        : "phase_window_fallback",
  };
}

function summarizePlaywrightErrors(report) {
  const messages = (report.errors ?? [])
    .map((error) => error?.message)
    .filter((message) => typeof message === "string" && message.trim() !== "");
  return messages.join("; ");
}

function createPlaywrightSelection({ manifestAware }) {
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";

  if (manifestAware && reportSlice) {
    const phase = requiredEnv("CARTULARY_MANIFEST_PHASE");
    const coverage = requiredEnv("CARTULARY_MANIFEST_COVERAGE");
    const executionDependency = optionalEnv("CARTULARY_MANIFEST_EXECUTION_DEPENDENCY");
    const selected = new Set(
      selectPlaywrightManifestEntries(phase, coverage, executionDependency).map(
        (entry) => `${entry.file}::${entry.title}`,
      ),
    );
    return {
      matches(normalizedFile, title) {
        return selected.has(`${normalizedFile}::${title}`);
      },
    };
  }

  const selectedFiles = new Set(
    optionalLines("CARTULARY_PLAYWRIGHT_FILES").map((value) => normalizePlaywrightFile(value)),
  );
  const selectedTitles = optionalSetFromLines("CARTULARY_PLAYWRIGHT_TITLES");
  if (selectedFiles.size === 0 && selectedTitles.size === 0) {
    return null;
  }
  return {
    matches(normalizedFile, title) {
      if (selectedFiles.size > 0 && !selectedFiles.has(normalizedFile)) {
        return false;
      }
      if (selectedTitles.size > 0 && !selectedTitles.has(title)) {
        return false;
      }
      return true;
    },
  };
}

function summarizePlaywrightRun(reportFile, phaseLabel, selection = null) {
  const report = JSON.parse(readFileSync(reportFile, "utf8"));
  const owners = new Set();
  const inventory = [];
  const dossiers = [];
  const counts = createCounts();

  const specs = flattenPlaywrightSuites(report.suites).filter((spec) => {
    if (!selection) {
      return true;
    }
    return selection.matches(
      normalizePlaywrightFile(spec.file ?? ""),
      spec.title ?? "",
    );
  });
  if (specs.length === 0 && (report.errors ?? []).length > 0) {
    const coverage = /\bsupport\b/i.test(phaseLabel) ? "support" : "unmapped";
    const message = (report.errors ?? [])
      .map((error) => error?.message)
      .find((entry) => typeof entry === "string" && entry.trim() !== "");
    dossiers.push({
      coverage,
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "playwright",
      package_or_file: "(playwright setup)",
      symbol_or_title: "(playwright setup)",
      message: (message ?? "playwright setup failure").split("\n")[0],
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
      raw: relToRepo(reportFile),
    });
    counts.failed += 1;
    counts[`${coverage}_failed`] += 1;
    return {
      report,
      counts,
      owners: [],
      inventory,
      dossiers,
    };
  }

  const topLevelErrorSummary = summarizePlaywrightErrors(report);
  if (topLevelErrorSummary !== "") {
    dossiers.push({
      coverage: "non_test",
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "playwright",
      package_or_file: "(playwright runner)",
      symbol_or_title: "(playwright runner)",
      message: topLevelErrorSummary.split("\n")[0],
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
      raw: relToRepo(reportFile),
    });
    counts.failed += 1;
    counts.non_test_failed += 1;
  }

  for (const spec of specs) {
    const classification = classifyPlaywrightCase(spec.file ?? "", spec.title ?? "", phaseLabel);
    owners.add(classification.owner);
    const executedResults = [];
    for (const test of spec.tests ?? []) {
      for (const result of test.results ?? []) {
        if (result.status && result.status !== "skipped") {
          executedResults.push({
            retry: result.retry ?? 0,
            status: result.status,
            result,
          });
        }
      }
    }
    if (executedResults.length === 0) {
      continue;
    }

    counts.tests += 1;
    counts[classification.coverage] += 1;
    const successful = executedResults.some(
      (entry) => entry.status === "passed" || entry.status === "flaky",
    );
    if (successful) {
      inventory.push(
        createInventoryItem({
          coverage: classification.coverage,
          phase: classification.phase,
          id: classification.id,
          owner: classification.owner,
          name: spec.title ?? "(missing title)",
        }),
      );
    }

    if (successful && executedResults.every((entry) => entry.status === "passed" || entry.status === "flaky")) {
      continue;
    }

    counts.failed += 1;
    counts[`${classification.coverage}_failed`] += 1;
    for (const attempt of executedResults.filter((entry) => entry.status !== "passed" && entry.status !== "flaky")) {
      const attachments = (attempt.result.attachments ?? [])
        .map((attachment) => attachment?.path)
        .filter((entry) => typeof entry === "string" && entry !== "")
        .map((entry) => relToRepo(entry));
      const message =
        attempt.result.error?.message?.split("\n")[0] ||
        (attempt.result.errors ?? [])
          .map((error) => error?.message)
          .find((entry) => typeof entry === "string" && entry.trim() !== "")
          ?.split("\n")[0] ||
        `${spec.title ?? "playwright spec"} failed`;
      dossiers.push({
        coverage: classification.coverage,
        phase: classification.phase,
        id: classification.id,
        runner: "playwright",
        package_or_file: classification.owner,
        symbol_or_title: spec.title ?? "(missing title)",
        message: message.trim(),
        reproduce: `pnpm --dir apps/web exec playwright test ${classification.owner.replace(/^apps\/web\//, "")} -g '^${escapeSingleQuotes(spec.title ?? "")}$'`,
        raw: attachments.length > 0 ? `${relToRepo(reportFile)};${attachments.join(";")}` : relToRepo(reportFile),
        retry: String(attempt.retry ?? 0),
      });
    }
  }

  counts.packages = owners.size;
  if (counts.tests === 0 && dossiers.length === 0) {
    const coverage = /\bsupport\b/i.test(phaseLabel) ? "support" : "unmapped";
    dossiers.push({
      coverage,
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "playwright",
      package_or_file: "(playwright selection)",
      symbol_or_title: "(playwright selection)",
      message: "phase matched zero tests",
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
      raw: relToRepo(reportFile),
    });
    counts.failed += 1;
    counts[`${coverage}_failed`] += 1;
  }
  return {
    report,
    counts,
    owners: Array.from(owners).sort(),
    inventory,
    dossiers,
    playwrightTiming: summarizePlaywrightTiming(specs, inferPhaseFromText(phaseLabel)),
  };
}

function selectPlaywrightManifestEntries(phase, coverage, executionDependency) {
  const { manifest } = loadManifest(repoRoot, phase);
  return collectEntries(manifest).filter((entry) => {
    if (entry.runner !== "playwright" || entry.section !== "e2e" || entry.coverage !== coverage) {
      return false;
    }
    if (executionDependency && entry.execution_dependency !== executionDependency) {
      return false;
    }
    return true;
  });
}

function evaluatePlaywrightManifest(summary) {
  const phase = requiredEnv("CARTULARY_MANIFEST_PHASE");
  const coverage = requiredEnv("CARTULARY_MANIFEST_COVERAGE");
  const executionDependency = optionalEnv("CARTULARY_MANIFEST_EXECUTION_DEPENDENCY");
  const entries = selectPlaywrightManifestEntries(phase, coverage, executionDependency);
  const expectedKeys = new Set(entries.map((entry) => `${entry.file}::${entry.title}`));
  const expectedCoverage = coverage === "authoritative" ? "authoritative" : "support";
  const executedKeys = new Set(
    summary.inventory
      .filter((item) => item.coverage === expectedCoverage)
      .map((item) => `${item.package_or_file}::${item.symbol_or_title}`),
  );
  const missingIDs = entries
    .filter((entry) => !executedKeys.has(`${entry.file}::${entry.title}`))
    .map((entry) => entry.id)
    .sort();
  const expectedIDs = new Set(entries.map((entry) => entry.id));
  const unexpectedIDs = summary.inventory
    .filter((item) => item.coverage === expectedCoverage && item.id && !expectedIDs.has(item.id))
    .map((item) => item.id)
    .sort();
  return {
    phase,
    missingIDs,
    unexpectedIDs,
    forbiddenIDFiles: Array.from(loadManifestIndex().forbiddenFilesByPhase.get(phase) ?? []).sort(),
  };
}

function handlePlaywrightPhase({ manifestAware }) {
  const context = createBasePhaseContext("playwright");
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";
  const reportFile = requiredEnv("CARTULARY_PHASE_RUNNER_LOG");
  const selectionReport = optionalEnv("CARTULARY_PLAYWRIGHT_SELECTION_REPORT");
  const stdoutLog = optionalEnv("CARTULARY_PHASE_STDOUT_LOG");
  const stderrLog = optionalEnv("CARTULARY_PHASE_STDERR_LOG");
  const outputDir = optionalEnv("CARTULARY_PLAYWRIGHT_OUTPUT_DIR");
  const serverLog = optionalEnv("CARTULARY_WEB_E2E_SERVER_LOG");
  const webLog = optionalEnv("CARTULARY_WEB_E2E_WEB_LOG");
  removeEmptyArtifact(stdoutLog);
  removeEmptyArtifact(stderrLog);

  const summary = summarizePlaywrightRun(
    reportFile,
    context.label,
    createPlaywrightSelection({ manifestAware }),
  );
  if (reportSlice && summary.playwrightTiming) {
    const timing = summary.playwrightTiming;
    const executedDurationMs = clampDurationMs(timing.executed_duration_ms ?? 0);
    const wallDurationMs = clampDurationMs(timing.wall_duration_ms ?? 0);
    if (executedDurationMs > 0) {
      context.executedDurationMs = executedDurationMs;
      context.logicalDurationMs = executedDurationMs;
      context.reusedDurationMs = context.accountingMode === "reused" ? executedDurationMs : 0;
      context.derivedDurationMs = context.accountingMode === "derived" ? executedDurationMs : 0;
    }
    if (wallDurationMs > 0) {
      context.wallDurationMs = wallDurationMs;
    }
    if (timing.start_time && timing.end_time) {
      context.startTime = timing.start_time;
      context.endTime = timing.end_time;
    }
  }
  const selectedSlicePassed =
    summary.dossiers.length === 0 && (context.exitStatus === 0 || reportSlice);
  let status = selectedSlicePassed ? "pass" : "fail";
  let manifestSummary = null;
  let manifestMismatch = null;

  if (manifestAware && selectedSlicePassed) {
    const verification = evaluatePlaywrightManifest(summary);
    manifestSummary = {
      missing_ids: verification.missingIDs,
      unexpected_ids: verification.unexpectedIDs,
    };
    if (verification.missingIDs.length > 0 || verification.unexpectedIDs.length > 0) {
      status = "mismatch";
      manifestMismatch = {
        missing_ids: verification.missingIDs,
        unexpected_ids: verification.unexpectedIDs,
        forbidden_id_files: verification.forbiddenIDFiles,
        selection: selectionReport && existsSync(selectionReport) ? relToRepo(selectionReport) : "",
        runner: relToRepo(reportFile),
      };
    }
  }

  writePhaseArtifacts(context, {
    status,
    phase: inferPhaseFromText(context.label),
    counts: summary.counts,
    owners: summary.owners,
    inventory: summary.inventory,
    dossiers: summary.dossiers,
    playwrightTiming: summary.playwrightTiming,
    manifestSummary,
    manifestMismatch,
    artifacts: {
      selection_json: selectionReport,
      runner_json: reportFile,
      stdout_log: existsSync(stdoutLog) ? stdoutLog : "",
      stderr_log: existsSync(stderrLog) ? stderrLog : "",
      playwright_output: outputDir,
      server_log: serverLog,
      web_log: webLog,
    },
  });

  if (status === "pass") {
    return 0;
  }
  if (status === "mismatch") {
    printBlock(`manifest mismatch: ${context.label}`, {
      missing_ids: renderList(manifestMismatch.missing_ids),
      unexpected_ids: renderList(manifestMismatch.unexpected_ids),
      forbidden_id_files: renderList(manifestMismatch.forbidden_id_files),
      selection: manifestMismatch.selection,
      runner: manifestMismatch.runner,
    });
    return 1;
  }
  for (const dossier of summary.dossiers) {
    const { runner, raw, ...rest } = dossier;
    printBlock(`failure: ${context.label}`, {
      ...rest,
      test_runner: runner,
      selection: selectionReport && existsSync(selectionReport) ? relToRepo(selectionReport) : "",
      runner: relToRepo(reportFile),
      raw: mergeRawPaths(raw, renderRawList([reportFile, outputDir, serverLog, webLog])),
    });
  }
  return 1;
}

function renderRawList(paths) {
  return paths
    .filter((entry) => entry && existsSync(entry))
    .map((entry) => relToRepo(entry))
    .join(";");
}

function mergeRawPaths(...values) {
  return values
    .filter(Boolean)
    .join(";")
    .split(";")
    .map((entry) => entry.trim())
    .filter(Boolean)
    .filter((entry, index, array) => array.indexOf(entry) === index)
    .join(";");
}

function escapeSingleQuotes(value) {
  return value.replaceAll("'", "'\"'\"'");
}

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
