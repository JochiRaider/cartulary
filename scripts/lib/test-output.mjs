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

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..");
const resultsRoot = resolveResultsRoot();
const runId = process.env.CARTULARY_TEST_RUN_ID || "adhoc";

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
    case "run-summary":
      process.exit(handleRunSummary(rest));
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
  return `actual_phases=${modes.actual ?? 0} reused_phases=${modes.reused ?? 0} derived_phases=${modes.derived ?? 0}`;
}

function formatDurationFields(wallDurationMs, executedDurationMs, logicalDurationMs = executedDurationMs) {
  const effectiveLogical = clampDurationMs(logicalDurationMs);
  const effectiveExecuted = clampDurationMs(executedDurationMs);
  const effectiveWall = Number.isFinite(wallDurationMs) ? wallDurationMs : effectiveLogical;
  return `wall_duration=${formatDuration(effectiveWall)} executed_duration=${formatDuration(effectiveExecuted)} logical_duration=${formatDuration(effectiveLogical)}`;
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
    startTime: requiredEnv("CARTULARY_PHASE_START_TIME"),
    endTime: requiredEnv("CARTULARY_PHASE_END_TIME"),
    accountingMode,
    durationMs: logicalDurationMs,
    executedDurationMs,
    logicalDurationMs,
    wallDurationMs: clampDurationMs(parseInteger("CARTULARY_PHASE_WALL_DURATION_MS", logicalDurationMs)),
    exitStatus: parseInteger("CARTULARY_PHASE_EXIT_STATUS", 0),
  };
}

function writePhaseArtifacts(context, details) {
  const artifacts = {};
  for (const [key, value] of Object.entries(details.artifacts ?? {})) {
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
    duration_ms: context.durationMs,
    wall_duration_ms: context.wallDurationMs,
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
    duration_ms: context.durationMs,
    wall_duration_ms: context.wallDurationMs,
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
  let durationMs = 0;
  let executedDurationMs = 0;
  let logicalDurationMs = 0;
  let wallDurationMs = 0;
  const accountingModes = createAccountingModes();
  let failed = false;

  for (const summary of summaries) {
    const accountingMode = normalizeAccountingMode(summary.accounting_mode);
    const summaryLogicalDurationMs = clampDurationMs(
      summary.logical_duration_ms ?? summary.duration_ms ?? 0,
    );
    const summaryExecutedDurationMs = clampDurationMs(
      summary.executed_duration_ms ??
        (accountingMode === "actual" ? summaryLogicalDurationMs : 0),
    );
    if (startTime === "" || summary.start_time < startTime) {
      startTime = summary.start_time;
    }
    if (endTime === "" || summary.end_time > endTime) {
      endTime = summary.end_time;
    }
    logicalDurationMs += summaryLogicalDurationMs;
    executedDurationMs += summaryExecutedDurationMs;
    durationMs += summaryLogicalDurationMs;
    wallDurationMs += clampDurationMs(summary.wall_duration_ms ?? summaryLogicalDurationMs);
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

  return {
    target,
    targetDir,
    summaries,
    counts,
    startTime,
    endTime,
    durationMs,
    executedDurationMs,
    logicalDurationMs,
    wallDurationMs,
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
    throw new Error("usage: test-output.mjs target-summary <target> [pass|fail] [--children <target,target,...>]");
  }

  let requestedStatus = "pass";
  const remaining = [...rest];
  if (remaining.length > 0 && !remaining[0].startsWith("--")) {
    requestedStatus = remaining.shift();
  }

  const childTargetNames = [];
  while (remaining.length > 0) {
    const option = remaining.shift();
    if (option !== "--children") {
      throw new Error(`unknown target-summary option ${option}`);
    }
    const value = remaining.shift();
    if (value === undefined) {
      throw new Error("--children requires a comma-separated target list");
    }
    childTargetNames.push(...parseTargetList(value));
  }

  return { target, requestedStatus, childTargetNames };
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

function toTargetSummaryReference(summary, fallbackTarget) {
  return {
    target: summary.target ?? fallbackTarget,
    status: summary.status ?? "",
    start_time: summary.start_time ?? "",
    end_time: summary.end_time ?? "",
    executed_duration_ms: clampDurationMs(
      summary.executed_duration_ms ?? summary.logical_duration_ms ?? summary.duration_ms ?? 0,
    ),
    logical_duration_ms: clampDurationMs(summary.logical_duration_ms ?? summary.duration_ms ?? 0),
    duration_ms: clampDurationMs(summary.duration_ms ?? summary.logical_duration_ms ?? 0),
    wall_duration_ms: clampDurationMs(
      summary.wall_duration_ms ?? summary.logical_duration_ms ?? summary.duration_ms ?? 0,
    ),
    counts: summary.counts ?? { phases: 0, ...createCounts() },
    accounting_modes: resolveAccountingModes(
      summary.accounting_modes,
      summary.counts?.phases ?? 0,
    ),
    artifacts: {
      dir: summary.artifacts?.dir ?? relToRepo(path.join(resultsRoot, runId, fallbackTarget)),
    },
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

function writeTargetLine(stream, label, target, summary, targetDir) {
  stream.write(
    `${label} ${target} phases=${summary.counts.phases} tests=${summary.counts.tests} authoritative=${summary.counts.authoritative} support=${summary.counts.support} unmapped=${summary.counts.unmapped} packages=${summary.counts.packages} ${formatDurationFields(summary.wallDurationMs, summary.executedDurationMs, summary.logicalDurationMs)} ${formatAccountingModeFields(summary.accountingModes)} artifacts=${relToRepo(targetDir)}\n`,
  );
}

function writeChildTargetLines(stream, parentTarget, childTargets, missingChildTargetSummaries) {
  for (const child of childTargets) {
    stream.write(
      `[CHILD] ${parentTarget} ${child.target} status=${child.status} phases=${child.counts?.phases ?? 0} tests=${child.counts?.tests ?? 0} ${formatDurationFields(child.wall_duration_ms, child.executed_duration_ms, child.logical_duration_ms)} ${formatAccountingModeFields(child.accounting_modes)} artifacts=${child.artifacts?.dir ?? ""}\n`,
    );
  }
  for (const childTarget of missingChildTargetSummaries) {
    stream.write(
      `[CHILD-MISSING] ${parentTarget} ${childTarget} artifacts=${relToRepo(targetSummaryPath(childTarget))}\n`,
    );
  }
}

function handleTargetSummary(args) {
  const { target, requestedStatus, childTargetNames } = parseTargetSummaryArgs(args);
  const summary = summarizeTargetDir(target);
  const { childTargets, missingChildTargetSummaries } =
    loadChildTargetSummaries(childTargetNames);
  const status =
    summary.failed || missingChildTargetSummaries.length > 0 || requestedStatus === "fail"
      ? "FAIL"
      : "PASS";
  const targetSummary = {
    target,
    status: status.toLowerCase(),
    start_time: summary.startTime,
    end_time: summary.endTime,
    executed_duration_ms: summary.executedDurationMs,
    logical_duration_ms: summary.logicalDurationMs,
    duration_ms: summary.durationMs,
    wall_duration_ms: summary.wallDurationMs,
    accounting_modes: summary.accountingModes,
    counts: summary.counts,
    artifacts: {
      dir: relToRepo(summary.targetDir),
    },
    child_targets: childTargets,
    missing_child_target_summaries: missingChildTargetSummaries,
  };
  writeJson(path.join(summary.targetDir, "target-summary.json"), targetSummary);

  if (status === "PASS") {
    writeTargetLine(process.stdout, "[PASS]", target, summary, summary.targetDir);
    writeChildTargetLines(process.stdout, target, childTargets, missingChildTargetSummaries);
    printInventory(summary);
    return 0;
  }

  process.stderr.write(
    `[FAIL] ${target} phases=${summary.counts.phases} tests=${summary.counts.tests} failed=${summary.counts.failed} authoritative_failed=${summary.counts.authoritative_failed} support_failed=${summary.counts.support_failed} unmapped_failed=${summary.counts.unmapped_failed} non_test_failed=${summary.counts.non_test_failed} ${formatDurationFields(summary.wallDurationMs, summary.executedDurationMs, summary.logicalDurationMs)} ${formatAccountingModeFields(summary.accountingModes)} artifacts=${relToRepo(summary.targetDir)}\n`,
  );
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
    throw new Error("usage: test-output.mjs run-summary <label> <pass|fail> <completed> <total> <aborted_after|-> [--summary-groups <name=a,b;name=c>] [targets...]");
  }
  const targets = [];
  const summaryGroups = [];
  while (remaining.length > 0) {
    const value = remaining.shift();
    if (value === "--summary-groups") {
      const spec = remaining.shift();
      if (spec === undefined) {
        throw new Error("--summary-groups requires <name=a,b;name=c>");
      }
      summaryGroups.push(...parseSummaryGroupsSpec(spec));
      continue;
    }
    targets.push(value);
  }
  return { label, requestedStatus, completedText, totalText, abortedAfter, targets, summaryGroups };
}

function createDurationAggregate() {
  return {
    phases: 0,
    ...createCounts(),
    duration_ms: 0,
    executed_duration_ms: 0,
    logical_duration_ms: 0,
    wall_duration_ms: 0,
  };
}

function addSummaryToAggregate(aggregate, accountingModes, summary) {
  aggregate.phases += summary.counts?.phases ?? 0;
  aggregate.tests += summary.counts?.tests ?? 0;
  aggregate.failed += summary.counts?.failed ?? 0;
  aggregate.authoritative += summary.counts?.authoritative ?? 0;
  aggregate.support += summary.counts?.support ?? 0;
  aggregate.unmapped += summary.counts?.unmapped ?? 0;
  aggregate.non_test += summary.counts?.non_test ?? 0;
  aggregate.authoritative_failed += summary.counts?.authoritative_failed ?? 0;
  aggregate.support_failed += summary.counts?.support_failed ?? 0;
  aggregate.unmapped_failed += summary.counts?.unmapped_failed ?? 0;
  aggregate.non_test_failed += summary.counts?.non_test_failed ?? 0;
  const summaryLogicalDurationMs = clampDurationMs(
    summary.logical_duration_ms ?? summary.duration_ms ?? 0,
  );
  aggregate.logical_duration_ms += summaryLogicalDurationMs;
  aggregate.duration_ms += summaryLogicalDurationMs;
  aggregate.executed_duration_ms += clampDurationMs(
    summary.executed_duration_ms ?? summaryLogicalDurationMs,
  );
  aggregate.wall_duration_ms += clampDurationMs(
    summary.wall_duration_ms ?? summaryLogicalDurationMs,
  );
  mergeAccountingModes(
    accountingModes,
    resolveAccountingModes(summary.accounting_modes, summary.counts?.phases ?? 0),
  );
}

function summarizeTargetSummaries(summaries, missingTargetSummaries, requestedStatus = "pass") {
  const aggregate = createDurationAggregate();
  const accountingModes = createAccountingModes();
  let failed = requestedStatus === "fail" || missingTargetSummaries.length > 0;
  let startTime = "";
  let endTime = "";

  for (const summary of summaries) {
    addSummaryToAggregate(aggregate, accountingModes, summary);
    if (startTime === "" || (summary.start_time && summary.start_time < startTime)) {
      startTime = summary.start_time ?? "";
    }
    if (endTime === "" || (summary.end_time && summary.end_time > endTime)) {
      endTime = summary.end_time ?? "";
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
  return { aggregate, accountingModes, failed, startTime, endTime, wallDurationMs };
}

function buildSummaryGroups(summaryGroups) {
  return summaryGroups.map((group) => {
    const groupSummaries = [];
    const missingTargetSummaries = [];
    for (const target of group.targets) {
      const summary = loadTargetSummary(target);
      if (!summary) {
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
      start_time: summarized.startTime,
      end_time: summarized.endTime,
      executed_duration_ms: summarized.aggregate.executed_duration_ms,
      logical_duration_ms: summarized.aggregate.logical_duration_ms,
      duration_ms: summarized.aggregate.duration_ms,
      wall_duration_ms: summarized.wallDurationMs,
      accounting_modes: summarized.accountingModes,
      counts: summarized.aggregate,
    };
  });
}

function writeSummaryGroupLines(stream, label, summaryGroups) {
  for (const group of summaryGroups) {
    const missing =
      group.missing_target_summaries.length > 0
        ? ` missing=${group.missing_target_summaries.join(",")}`
        : "";
    stream.write(
      `[GROUP] ${label} ${group.name} targets=${group.targets.join(",")} status=${group.status} ${formatDurationFields(group.wall_duration_ms, group.executed_duration_ms, group.logical_duration_ms)} ${formatAccountingModeFields(group.accounting_modes)}${missing}\n`,
    );
  }
}

function handleRunSummary(args) {
  const { label, requestedStatus, completedText, totalText, abortedAfter, targets, summaryGroups } =
    parseRunSummaryArgs(args);
  const completedTargets = Number.parseInt(completedText, 10) || 0;
  const totalTargets = Number.parseInt(totalText, 10) || 0;
  const missingTargetSummaries = [];
  const targetSummaries = [];

  for (const target of targets) {
    const summary = loadTargetSummary(target);
    if (!summary) {
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
  const renderedSummaryGroups = buildSummaryGroups(summaryGroups);
  const failed = summarized.failed || renderedSummaryGroups.some((group) => group.status !== "pass");

  const runSummary = {
    label,
    status: failed ? "fail" : "pass",
    completed_targets: `${completedTargets}/${totalTargets}`,
    aborted_after: abortedAfter === "-" ? "" : abortedAfter,
    start_time: summarized.startTime,
    end_time: summarized.endTime,
    executed_duration_ms: aggregate.executed_duration_ms,
    logical_duration_ms: aggregate.logical_duration_ms,
    duration_ms: aggregate.duration_ms,
    wall_duration_ms: wallDurationMs,
    accounting_modes: accountingModes,
    counts: aggregate,
    artifacts: {
      dir: relToRepo(path.join(resultsRoot, runId)),
    },
    targets,
    target_summaries: targetSummaries,
    missing_target_summaries: missingTargetSummaries,
    summary_groups: renderedSummaryGroups,
  };
  writeJson(path.join(resultsRoot, runId, "run-summary.json"), runSummary);

  if (!failed) {
    process.stdout.write(
      `[PASS] ${label} completed_targets=${completedTargets}/${totalTargets} phases=${aggregate.phases} tests=${aggregate.tests} authoritative=${aggregate.authoritative} support=${aggregate.support} unmapped=${aggregate.unmapped} ${formatDurationFields(wallDurationMs, aggregate.executed_duration_ms, aggregate.logical_duration_ms)} ${formatAccountingModeFields(accountingModes)} artifacts=${relToRepo(path.join(resultsRoot, runId))}\n`,
    );
    writeSummaryGroupLines(process.stdout, label, renderedSummaryGroups);
    return 0;
  }

  process.stderr.write(
    `[FAIL] ${label} completed_targets=${completedTargets}/${totalTargets} aborted_after=${abortedAfter === "-" ? "-" : abortedAfter} phases=${aggregate.phases} tests=${aggregate.tests} failed=${aggregate.failed} authoritative_failed=${aggregate.authoritative_failed} support_failed=${aggregate.support_failed} unmapped_failed=${aggregate.unmapped_failed} non_test_failed=${aggregate.non_test_failed} ${formatDurationFields(wallDurationMs, aggregate.executed_duration_ms, aggregate.logical_duration_ms)} ${formatAccountingModeFields(accountingModes)} artifacts=${relToRepo(path.join(resultsRoot, runId))}\n`,
  );
  writeSummaryGroupLines(process.stderr, label, renderedSummaryGroups);
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
    const entries = selectGoManifestEntries(
      phase,
      section,
      coverage,
      executionDependency,
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

function selectGoManifestEntries(phase, section, coverage, executionDependency, packagePatterns) {
  const { manifest } = loadManifest(repoRoot, phase);
  return collectEntries(manifest).filter((entry) => {
    if (entry.runner !== "go_test" || entry.section !== section || entry.coverage !== coverage) {
      return false;
    }
    if (executionDependency && entry.execution_dependency !== executionDependency) {
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
  const packagePatterns = optionalLines("CARTULARY_GO_PACKAGE_PATTERNS");
  const entries = selectGoManifestEntries(
    phase,
    section,
    coverage,
    executionDependency,
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
  removeEmptyArtifact(stderrLog);
  removeEmptyArtifact(stdoutLog);

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

function handlePlaywrightPhase({ manifestAware }) {
  const context = createBasePhaseContext("playwright");
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
  let status = context.exitStatus === 0 && summary.dossiers.length === 0 ? "pass" : "fail";
  let manifestSummary = null;
  let manifestMismatch = null;

  if (manifestAware && context.exitStatus === 0 && summary.dossiers.length === 0) {
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
