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
  return {
    label: requiredEnv("CARTULARY_PHASE_LABEL"),
    phaseDir,
    target: optionalEnv("CARTULARY_TEST_TARGET", "adhoc"),
    command: requiredEnv("CARTULARY_PHASE_COMMAND"),
    runner,
    startTime: requiredEnv("CARTULARY_PHASE_START_TIME"),
    endTime: requiredEnv("CARTULARY_PHASE_END_TIME"),
    durationMs: parseInteger("CARTULARY_PHASE_DURATION_MS", 0),
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
    duration_ms: context.durationMs,
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
    duration_ms: context.durationMs,
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
    tests: 0,
    failed: 0,
    authoritative: 0,
    support: 0,
    unmapped: 0,
    authoritative_failed: 0,
    support_failed: 0,
    unmapped_failed: 0,
    packages: 0,
  };
  let startTime = "";
  let endTime = "";
  let durationMs = 0;
  let failed = false;

  for (const summary of summaries) {
    if (startTime === "" || summary.start_time < startTime) {
      startTime = summary.start_time;
    }
    if (endTime === "" || summary.end_time > endTime) {
      endTime = summary.end_time;
    }
    durationMs += summary.duration_ms ?? 0;
    counts.tests += summary.counts?.tests ?? 0;
    counts.failed += summary.counts?.failed ?? 0;
    counts.authoritative += summary.counts?.authoritative ?? 0;
    counts.support += summary.counts?.support ?? 0;
    counts.unmapped += summary.counts?.unmapped ?? 0;
    counts.authoritative_failed += summary.counts?.authoritative_failed ?? 0;
    counts.support_failed += summary.counts?.support_failed ?? 0;
    counts.unmapped_failed += summary.counts?.unmapped_failed ?? 0;
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

function handleTargetSummary(args) {
  const [target, requestedStatus = "pass"] = args;
  if (!target) {
    throw new Error("usage: test-output.mjs target-summary <target> [pass|fail]");
  }
  const summary = summarizeTargetDir(target);
  const status = summary.failed || requestedStatus === "fail" ? "FAIL" : "PASS";
  const targetSummary = {
    target,
    status: status.toLowerCase(),
    start_time: summary.startTime,
    end_time: summary.endTime,
    duration_ms: summary.durationMs,
    counts: summary.counts,
    artifacts: {
      dir: relToRepo(summary.targetDir),
    },
  };
  writeJson(path.join(summary.targetDir, "target-summary.json"), targetSummary);

  if (status === "PASS") {
    process.stdout.write(
      `[PASS] ${target} phases=${summary.counts.phases} tests=${summary.counts.tests} authoritative=${summary.counts.authoritative} support=${summary.counts.support} unmapped=${summary.counts.unmapped} packages=${summary.counts.packages} duration=${formatDuration(summary.durationMs)} artifacts=${relToRepo(summary.targetDir)}\n`,
    );
    printInventory(summary);
    return 0;
  }

  process.stderr.write(
    `[FAIL] ${target} phases=${summary.counts.phases} tests=${summary.counts.tests} failed=${summary.counts.failed} authoritative_failed=${summary.counts.authoritative_failed} support_failed=${summary.counts.support_failed} unmapped_failed=${summary.counts.unmapped_failed} duration=${formatDuration(summary.durationMs)} artifacts=${relToRepo(summary.targetDir)}\n`,
  );
  return 0;
}

function handleRunSummary(args) {
  const [label, requestedStatus = "pass", completedText = "0", totalText = "0", abortedAfter = "", ...targets] = args;
  if (!label) {
    throw new Error("usage: test-output.mjs run-summary <label> <pass|fail> <completed> <total> <aborted_after|-> [targets...]");
  }
  const completedTargets = Number.parseInt(completedText, 10) || 0;
  const totalTargets = Number.parseInt(totalText, 10) || 0;
  const aggregate = {
    phases: 0,
    tests: 0,
    failed: 0,
    authoritative: 0,
    support: 0,
    unmapped: 0,
    authoritative_failed: 0,
    support_failed: 0,
    unmapped_failed: 0,
    duration_ms: 0,
  };
  let failed = requestedStatus === "fail";

  for (const target of targets) {
    const file = path.join(resultsRoot, runId, target, "target-summary.json");
    if (!existsSync(file)) {
      continue;
    }
    const summary = JSON.parse(readFileSync(file, "utf8"));
    aggregate.phases += summary.counts?.phases ?? 0;
    aggregate.tests += summary.counts?.tests ?? 0;
    aggregate.failed += summary.counts?.failed ?? 0;
    aggregate.authoritative += summary.counts?.authoritative ?? 0;
    aggregate.support += summary.counts?.support ?? 0;
    aggregate.unmapped += summary.counts?.unmapped ?? 0;
    aggregate.authoritative_failed += summary.counts?.authoritative_failed ?? 0;
    aggregate.support_failed += summary.counts?.support_failed ?? 0;
    aggregate.unmapped_failed += summary.counts?.unmapped_failed ?? 0;
    aggregate.duration_ms += summary.duration_ms ?? 0;
    if (summary.status !== "pass") {
      failed = true;
    }
  }

  const runSummary = {
    label,
    status: failed ? "fail" : "pass",
    completed_targets: `${completedTargets}/${totalTargets}`,
    aborted_after: abortedAfter === "-" ? "" : abortedAfter,
    counts: aggregate,
    artifacts: {
      dir: relToRepo(path.join(resultsRoot, runId)),
    },
    targets,
  };
  writeJson(path.join(resultsRoot, runId, "run-summary.json"), runSummary);

  if (!failed) {
    process.stdout.write(
      `[PASS] ${label} completed_targets=${completedTargets}/${totalTargets} phases=${aggregate.phases} tests=${aggregate.tests} authoritative=${aggregate.authoritative} support=${aggregate.support} unmapped=${aggregate.unmapped} duration=${formatDuration(aggregate.duration_ms)} artifacts=${relToRepo(path.join(resultsRoot, runId))}\n`,
    );
    return 0;
  }

  process.stderr.write(
    `[FAIL] ${label} completed_targets=${completedTargets}/${totalTargets} aborted_after=${abortedAfter === "-" ? "-" : abortedAfter} phases=${aggregate.phases} tests=${aggregate.tests} failed=${aggregate.failed} authoritative_failed=${aggregate.authoritative_failed} support_failed=${aggregate.support_failed} unmapped_failed=${aggregate.unmapped_failed} duration=${formatDuration(aggregate.duration_ms)} artifacts=${relToRepo(path.join(resultsRoot, runId))}\n`,
  );
  return 0;
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
        tests: 0,
        failed: 0,
        authoritative: 0,
        support: 0,
        unmapped: 0,
        authoritative_failed: 0,
        support_failed: 0,
        unmapped_failed: 0,
        packages: 0,
      },
      owners: [],
      inventory: [],
      dossiers: [],
    });
  }

  const message =
    firstActionableLine(splitLogLines(stderrLog)) ||
    firstActionableLine(splitLogLines(stdoutLog)) ||
    `command exited with status ${context.exitStatus}`;
  return finalizeShellPhase(context, stdoutLog, stderrLog, {
    status: "fail",
    phase: inferPhaseFromText(context.label),
    counts: {
      tests: 0,
      failed: 1,
      authoritative: 0,
      support: 0,
      unmapped: 0,
      authoritative_failed: 0,
      support_failed: 0,
      unmapped_failed: 1,
      packages: 0,
    },
    owners: [],
    inventory: [],
    dossiers: [
      {
        coverage: "unmapped",
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

function summarizeGoRun(logFile, phaseLabel, exitStatus) {
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
  const counts = {
    tests: 0,
    failed: 0,
    authoritative: 0,
    support: 0,
    unmapped: 0,
    authoritative_failed: 0,
    support_failed: 0,
    unmapped_failed: 0,
    packages: 0,
  };
  let passedCount = 0;
  let skippedCount = 0;
  let incompleteCount = 0;

  for (const testCase of topLevel.values()) {
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
    const message =
      passedCount === 0 && skippedCount === 0 && incompleteCount === 0
        ? "phase matched zero tests"
        : `go test inventory requires top-level pass: skipped=${skippedCount} incomplete=${incompleteCount}`;
    dossiers.push({
      coverage: "unmapped",
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
    counts.unmapped_failed += 1;
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

  const summary = summarizeGoRun(runnerLog, context.label, context.exitStatus);
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

function classifyVitestCase(filePath, title, phaseLabel) {
  const normalizedFile = normalizeVitestFile(filePath);
  const authoritative = loadManifestIndex().authoritativeVitest.get(`${normalizedFile}::${title}`);
  if (authoritative) {
    return {
      coverage: "authoritative",
      phase: authoritative.phase,
      id: authoritative.id,
      owner: normalizedFile,
    };
  }
  const support =
    normalizedFile.includes(".support.") ||
    supportNamedTitle(title) ||
    isForbiddenFile(normalizedFile, inferPhaseFromText(normalizedFile) || inferPhaseFromText(title)) ||
    /\bsupport\b/i.test(phaseLabel);
  return {
    coverage: support ? "support" : "unmapped",
    phase: inferPhaseFromText(normalizedFile) || inferPhaseFromText(title) || inferPhaseFromText(phaseLabel),
    id: "",
    owner: normalizedFile,
  };
}

function normalizeVitestFile(filePath) {
  const relative = relToRepo(filePath);
  if (relative.startsWith("apps/web/")) {
    return relative;
  }
  return normalizePath(path.join("apps/web", relative));
}

function summarizeVitestRun(reportFile, phaseLabel) {
  const report = JSON.parse(readFileSync(reportFile, "utf8"));
  const owners = new Set();
  const inventory = [];
  const dossiers = [];
  const counts = {
    tests: 0,
    failed: 0,
    authoritative: 0,
    support: 0,
    unmapped: 0,
    authoritative_failed: 0,
    support_failed: 0,
    unmapped_failed: 0,
    packages: 0,
  };

  for (const fileResult of report.testResults ?? []) {
    const normalizedFile = normalizeVitestFile(fileResult.name ?? "");
    for (const assertion of fileResult.assertionResults ?? []) {
      if (assertion.status === "skipped") {
        continue;
      }
      const classification = classifyVitestCase(normalizedFile, assertion.title ?? "", phaseLabel);
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
        reproduce: `pnpm --dir apps/web exec vitest run ${classification.owner.replace(/^apps\/web\//, "")} -t '${escapeSingleQuotes((assertion.title ?? "").trim())}$'`,
        raw: relToRepo(reportFile),
      });
    }
  }

  if (counts.tests === 0) {
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

  const summary = summarizeVitestRun(reportFile, context.label);
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

function summarizePlaywrightRun(reportFile, phaseLabel) {
  const report = JSON.parse(readFileSync(reportFile, "utf8"));
  const owners = new Set();
  const inventory = [];
  const dossiers = [];
  const counts = {
    tests: 0,
    failed: 0,
    authoritative: 0,
    support: 0,
    unmapped: 0,
    authoritative_failed: 0,
    support_failed: 0,
    unmapped_failed: 0,
    packages: 0,
  };

  const specs = flattenPlaywrightSuites(report.suites);
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
  const listReport = optionalEnv("CARTULARY_PLAYWRIGHT_LIST_REPORT");
  const stdoutLog = optionalEnv("CARTULARY_PHASE_STDOUT_LOG");
  const stderrLog = optionalEnv("CARTULARY_PHASE_STDERR_LOG");
  const outputDir = optionalEnv("CARTULARY_PLAYWRIGHT_OUTPUT_DIR");
  const serverLog = optionalEnv("CARTULARY_WEB_E2E_SERVER_LOG");
  const webLog = optionalEnv("CARTULARY_WEB_E2E_WEB_LOG");
  removeEmptyArtifact(stdoutLog);
  removeEmptyArtifact(stderrLog);

  const summary = summarizePlaywrightRun(reportFile, context.label);
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
        raw: renderRawList([listReport, reportFile]),
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
      list_json: listReport,
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
      raw: manifestMismatch.raw,
    });
    return 1;
  }
  for (const dossier of summary.dossiers) {
    printBlock(`failure: ${context.label}`, {
      ...dossier,
      raw: mergeRawPaths(dossier.raw, renderRawList([reportFile, outputDir, serverLog, webLog])),
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
