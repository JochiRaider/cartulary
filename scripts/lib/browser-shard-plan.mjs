#!/usr/bin/env node
import { existsSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import {
  collectEntries,
  entryIsExecutable,
  loadManifest,
  phaseManifestNames,
  playwrightEntryTitles,
} from "./phase-manifest.mjs";
import {
  durationDriftDescription,
  durationDriftKind,
} from "./duration-drift.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..");
const baselineSchemaID = "cartulary.browser_e2e_duration_baselines.v2";
const shardPlanSchemaID = "cartulary.browser_e2e_shard_plan.v2";
const defaultBaselineFile = path.join(repoRoot, "tools", "browser_e2e_duration_baselines.json");
const baselineNote =
  "Advisory browser functional manifest-entry weights for duration-balanced Playwright sharding. Refresh with make browser-e2e-duration-baselines RESULTS_DIR=<dir>.";
const defaultEntryWeightMs = 10000;
const defaultShardTargetMs = 12000;
const browserManifestIDPattern = /^E-[0-9]+-[A-Z0-9]+(?:-[A-Z0-9]+)*$/u;

function usage() {
  process.stderr.write(
    [
      "usage:",
      "  browser-shard-plan.mjs plan [--baseline-file <path>] [--min-shards <n>] [--max-shards <n>]",
      "  browser-shard-plan.mjs merge-reports <output-report> <input-report...>",
      "  browser-shard-plan.mjs update-baselines [--baseline-file <path>] <results-dir>",
      "  browser-shard-plan.mjs check-baseline-drift [--baseline-file <path>] <results-dir>",
    ].join("\n") + "\n",
  );
  process.exit(2);
}

function resolvePath(file) {
  return path.isAbsolute(file) ? file : path.join(repoRoot, file);
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function sortedObject(entries) {
  return Object.fromEntries([...entries].sort(([left], [right]) => left.localeCompare(right)));
}

function normalizeManifestFile(file) {
  const normalized = String(file ?? "").replaceAll("\\", "/");
  if (normalized.startsWith("apps/web/e2e/")) {
    return normalized;
  }
  if (normalized.startsWith("e2e/")) {
    return `apps/web/${normalized}`;
  }
  return normalized;
}

function normalizePlaywrightReportFile(file) {
  const normalized = String(file ?? "").replaceAll("\\", "/");
  if (normalized.startsWith("apps/web/e2e/")) {
    return normalized;
  }
  if (normalized.startsWith("e2e/")) {
    return `apps/web/${normalized}`;
  }
  return `apps/web/e2e/${normalized.replace(/^\/+/, "")}`;
}

function defaultBaselineDocument() {
  return {
    schema_id: baselineSchemaID,
    note: baselineNote,
    default_entry_weight_ms: defaultEntryWeightMs,
    shard_target_ms: defaultShardTargetMs,
    entries: {},
  };
}

function positiveIntegerOrDefault(value, fallback) {
  return Number.isInteger(value) && value > 0 ? value : fallback;
}

function readBaselineDocument(file, { allowMissing = true } = {}) {
  if (!existsSync(file)) {
    if (allowMissing) {
      return defaultBaselineDocument();
    }
    throw new Error(`${path.relative(repoRoot, file)} is missing`);
  }
  const baseline = readJSON(file);
  if (baseline.schema_id !== baselineSchemaID) {
    throw new Error(`${path.relative(repoRoot, file)} must declare schema_id ${baselineSchemaID}`);
  }
  const rawEntries = baseline.entries ?? {};
  if (!rawEntries || typeof rawEntries !== "object" || Array.isArray(rawEntries)) {
    throw new Error(`${path.relative(repoRoot, file)} entries must be an object`);
  }
  for (const [id, entry] of Object.entries(rawEntries)) {
    if (!browserManifestIDPattern.test(id)) {
      throw new Error(`${path.relative(repoRoot, file)} entries key ${id} must be an E-* manifest ID`);
    }
    if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
      throw new Error(`${path.relative(repoRoot, file)} entries.${id} must be an object`);
    }
    if (normalizeManifestFile(entry.file) === "") {
      throw new Error(`${path.relative(repoRoot, file)} entries.${id}.file must be non-empty`);
    }
    if (typeof entry.title !== "string" || entry.title.trim() === "") {
      throw new Error(`${path.relative(repoRoot, file)} entries.${id}.title must be non-empty`);
    }
    if (!Number.isInteger(entry.weight_ms) || entry.weight_ms <= 0) {
      throw new Error(`${path.relative(repoRoot, file)} entries.${id}.weight_ms must be a positive integer`);
    }
  }
  return baseline;
}

function browserFunctionalEntries(root, { phase: phaseFilter = "" } = {}) {
  const entries = [];
  const seenIDs = new Set();
  for (const phase of phaseManifestNames(root)) {
    if (phaseFilter && phase !== phaseFilter) {
      continue;
    }
    const { manifest } = loadManifest(root, phase);
    for (const entry of collectEntries(manifest)) {
      if (
        entry.section === "e2e" &&
        entry.runner === "playwright" &&
        entry.coverage === "authoritative" &&
        entry.execution_dependency === "browser_functional" &&
        entryIsExecutable(entry)
      ) {
        if (seenIDs.has(entry.id)) {
          throw new Error(`duplicate browser functional manifest ID ${entry.id}`);
        }
        seenIDs.add(entry.id);
        const titles = playwrightEntryTitles(entry);
        entries.push({
          id: entry.id,
          phase,
          file: normalizeManifestFile(entry.file),
          title: titles[0],
          titles,
        });
      }
    }
  }
  return entries.sort(compareEntries);
}

function compareEntries(left, right) {
  if (left.phase !== right.phase) {
    return left.phase.localeCompare(right.phase, undefined, { numeric: true });
  }
  if (left.file !== right.file) {
    return left.file.localeCompare(right.file);
  }
  if (left.title !== right.title) {
    return left.title.localeCompare(right.title);
  }
  return left.id.localeCompare(right.id, undefined, { numeric: true });
}

function baselineEntryMap(rawEntries) {
  return new Map(
    Object.entries(rawEntries).map(([id, entry]) => [
      id,
      {
        file: normalizeManifestFile(entry.file),
        title: entry.title,
        weight_ms: entry.weight_ms,
      },
    ]),
  );
}

function readBaseline(file, activeEntries) {
  const baseline = readBaselineDocument(file);
  const activeByID = new Map(activeEntries.map((entry) => [entry.id, entry]));
  const entries = baselineEntryMap(baseline.entries ?? {});
  for (const [id, baselineEntry] of entries) {
    const active = activeByID.get(id);
    if (!active) {
      throw new Error(`${path.relative(repoRoot, file)} contains retired browser entry baseline id=${id}`);
    }
    if (active.file !== baselineEntry.file || active.title !== baselineEntry.title) {
      throw new Error(
        `${path.relative(repoRoot, file)} entries.${id} must match active manifest file/title`,
      );
    }
  }
  return {
    defaultEntryWeightMs: positiveIntegerOrDefault(
      baseline.default_entry_weight_ms,
      defaultEntryWeightMs,
    ),
    shardTargetMs: positiveIntegerOrDefault(baseline.shard_target_ms, defaultShardTargetMs),
    entries,
  };
}

function collectEntryRows(root, baseline, { phase = "" } = {}) {
  return browserFunctionalEntries(root, { phase }).map((entry) => ({
    ...entry,
    weight_ms: baseline.entries.get(entry.id)?.weight_ms ?? baseline.defaultEntryWeightMs,
  }));
}

function uniqueSortedFiles(entries) {
  return [...new Set(entries.map((entry) => entry.file))].sort();
}

function planShardCount(entries, { minShards, maxShards, shardTargetMs }) {
  if (entries.length === 0) {
    return 0;
  }
  const totalWeight = entries.reduce((sum, entry) => sum + entry.weight_ms, 0);
  const targetCount = Math.ceil(totalWeight / Math.max(1, shardTargetMs));
  return Math.max(1, Math.min(entries.length, maxShards, Math.max(minShards, targetCount)));
}

export function createPlan({ baselineFile, minShards = 1, maxShards, phase = "" }) {
  const activeEntries = browserFunctionalEntries(repoRoot);
  const baseline = readBaseline(baselineFile, activeEntries);
  const entries = collectEntryRows(repoRoot, baseline, { phase });
  if (entries.length === 0) {
    throw new Error(
      phase
        ? `no authoritative browser_functional Playwright rows found for ${phase}`
        : "no authoritative browser_functional Playwright rows found",
    );
  }
  const shardCount = planShardCount(entries, {
    minShards,
    maxShards,
    shardTargetMs: baseline.shardTargetMs,
  });
  const shards = Array.from({ length: shardCount }, (_, index) => ({
    name: `browser-functional-shard-${String(index + 1).padStart(2, "0")}`,
    weight_ms: 0,
    files: new Set(),
    phases: new Set(),
    entries: [],
  }));

  const weightedEntries = [...entries].sort(
    (left, right) =>
      right.weight_ms - left.weight_ms ||
      left.id.localeCompare(right.id, undefined, { numeric: true }),
  );
  for (const entry of weightedEntries) {
    const shard = shards
      .slice()
      .sort((left, right) => left.weight_ms - right.weight_ms || left.name.localeCompare(right.name))[0];
    shard.weight_ms += entry.weight_ms;
    shard.files.add(entry.file);
    shard.phases.add(entry.phase);
    shard.entries.push(entry);
  }

  return {
    schema_id: shardPlanSchemaID,
    phase,
    generated_at: new Date().toISOString(),
    baseline_file: path.relative(repoRoot, baselineFile),
    min_shards: minShards,
    max_shards: maxShards,
    shard_count: shardCount,
    default_entry_weight_ms: baseline.defaultEntryWeightMs,
    shard_target_ms: baseline.shardTargetMs,
    entry_count: entries.length,
    file_count: uniqueSortedFiles(entries).length,
    files: uniqueSortedFiles(entries),
    entries,
    shards: shards.map((shard) => {
      const shardEntries = shard.entries.sort(compareEntries);
      return {
        name: shard.name,
        weight_ms: shard.weight_ms,
        entry_count: shardEntries.length,
        file_count: shard.files.size,
        files: [...shard.files].sort(),
        phases: [...shard.phases].sort((left, right) =>
          left.localeCompare(right, undefined, { numeric: true }),
        ),
        grep: exactAlternationRegex(shardEntries.flatMap((entry) => entry.titles ?? [entry.title])),
        entries: shardEntries.map((entry) => ({
          id: entry.id,
          phase: entry.phase,
          file: entry.file,
          title: entry.title,
          titles: entry.titles ?? [entry.title],
          weight_ms: entry.weight_ms,
        })),
      };
    }),
  };
}

function escapeRegex(value) {
  return value.replace(/[\\^$.*+?()[\]{}|]/g, "\\$&");
}

function exactAlternationRegex(values) {
  if (values.length === 0) {
    return "(?!)";
  }
  return `(?:${values.map(escapeRegex).join("|")})`;
}

function parsePlanArgs(argv) {
  const options = {
    baselineFile: defaultBaselineFile,
    maxShards: 1,
    minShards: 1,
    phase: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--baseline-file") {
      options.baselineFile = resolvePath(argv[index + 1] ?? "");
      index += 1;
      if (!options.baselineFile) {
        usage();
      }
      continue;
    }
    if (arg === "--max-shards") {
      options.maxShards = Number.parseInt(argv[index + 1] ?? "", 10);
      index += 1;
      if (!Number.isInteger(options.maxShards) || options.maxShards < 1) {
        usage();
      }
      continue;
    }
    if (arg === "--min-shards") {
      options.minShards = Number.parseInt(argv[index + 1] ?? "", 10);
      index += 1;
      if (!Number.isInteger(options.minShards) || options.minShards < 1) {
        usage();
      }
      continue;
    }
    if (arg === "--phase") {
      options.phase = argv[index + 1] ?? "";
      index += 1;
      if (!/^phase[0-9]+$/.test(options.phase)) {
        usage();
      }
      continue;
    }
    usage();
  }
  if (options.minShards > options.maxShards) {
    usage();
  }
  return options;
}

function flattenSuites(suites, specs = []) {
  for (const suite of suites ?? []) {
    flattenSuites(suite.suites, specs);
    for (const spec of suite.specs ?? []) {
      specs.push(spec);
    }
  }
  return specs;
}

function mergeReports(outputFile, inputFiles) {
  const specs = [];
  const errors = [];
  for (const inputFile of inputFiles) {
    if (!existsSync(inputFile)) {
      errors.push({ message: `missing Playwright shard report: ${inputFile}` });
      continue;
    }
    const report = readJSON(inputFile);
    specs.push(...flattenSuites(report.suites));
    errors.push(...(report.errors ?? []));
  }
  writeFileSync(
    outputFile,
    `${JSON.stringify({ suites: [{ specs, suites: [] }], errors }, null, 2)}\n`,
  );
}

function passingPlaywrightPhaseSummary(timingFile) {
  const summaryFile = path.join(path.dirname(timingFile), "phase-summary.json");
  if (!existsSync(summaryFile)) {
    return false;
  }
  try {
    const summary = readJSON(summaryFile);
    return summary.status === "pass" && summary.runner === "playwright";
  } catch {
    return false;
  }
}

function observedDurationMs(entryTiming) {
  return positiveIntegerOrDefault(
    entryTiming.wall_duration_ms || entryTiming.executed_duration_ms,
    0,
  );
}

function collectObservedBrowserEntryDurations(resultsDir, { requirePassingPhaseSummary = true } = {}) {
  const observed = new Map();
  const stack = [resultsDir];
  while (stack.length > 0) {
    const current = stack.pop();
    let entries = [];
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const entry of entries) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (!entry.isFile() || entry.name !== "playwright-timing.json") {
        continue;
      }
      if (requirePassingPhaseSummary && !passingPlaywrightPhaseSummary(next)) {
        continue;
      }
      const timing = readJSON(next);
      for (const entryTiming of timing.entries ?? []) {
        const id = String(entryTiming.id ?? "");
        const durationMs = observedDurationMs(entryTiming);
        if (browserManifestIDPattern.test(id) && durationMs > 0) {
          const currentObserved = observed.get(id);
          const normalized = {
            id,
            phase: String(entryTiming.phase ?? ""),
            file: normalizePlaywrightReportFile(entryTiming.file),
            title: String(entryTiming.title ?? ""),
            duration_ms: durationMs,
          };
          if (!currentObserved || durationMs > currentObserved.duration_ms) {
            observed.set(id, normalized);
          }
        }
      }
    }
  }
  return observed;
}

function parseBaselineResultsArgs(argv) {
  let baselineFile = defaultBaselineFile;
  let resultsDir = "";
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--baseline-file") {
      baselineFile = resolvePath(argv[index + 1] ?? "");
      index += 1;
      if (!baselineFile) {
        usage();
      }
      continue;
    }
    if (!resultsDir) {
      resultsDir = resolvePath(arg);
      continue;
    }
    usage();
  }
  if (!resultsDir) {
    usage();
  }
  return { baselineFile, resultsDir };
}

function activeEntryRowsForBaseline(baselineFile) {
  const activeEntries = browserFunctionalEntries(repoRoot);
  const baseline = readBaseline(baselineFile, activeEntries);
  return {
    baseline,
    entries: collectEntryRows(repoRoot, baseline),
  };
}

function updateBaselines(argv) {
  const { baselineFile, resultsDir } = parseBaselineResultsArgs(argv);
  const baseline = readBaselineDocument(baselineFile, { allowMissing: true });
  const authoritativeEntries = browserFunctionalEntries(repoRoot);
  const observed = collectObservedBrowserEntryDurations(resultsDir);
  const missingObserved = authoritativeEntries.filter((entry) => !observed.has(entry.id));
  if (missingObserved.length > 0) {
    throw new Error(
      `missing observed browser entry timings: ${missingObserved.map((entry) => entry.id).join(", ")}`,
    );
  }

  baseline.schema_id = baselineSchemaID;
  baseline.note = baselineNote;
  baseline.default_entry_weight_ms = positiveIntegerOrDefault(
    baseline.default_entry_weight_ms,
    defaultEntryWeightMs,
  );
  delete baseline.default_spec_weight_ms;
  delete baseline.specs;
  baseline.shard_target_ms = positiveIntegerOrDefault(baseline.shard_target_ms, defaultShardTargetMs);
  baseline.updated_at = new Date().toISOString();
  baseline.entries = sortedObject(
    authoritativeEntries.map((entry) => [
      entry.id,
      {
        file: entry.file,
        title: entry.title,
        weight_ms: observed.get(entry.id).duration_ms,
      },
    ]),
  );

  writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  process.stdout.write(
    `updated ${authoritativeEntries.length} browser E2E entry duration baselines from ${path.relative(repoRoot, resultsDir)}\n`,
  );
}

function driftSubject(entry) {
  return `id=${entry.id} file=${entry.file} title=${JSON.stringify(entry.title)}`;
}

function checkBaselineDrift(argv) {
  const { baselineFile, resultsDir } = parseBaselineResultsArgs(argv);
  const { baseline, entries } = activeEntryRowsForBaseline(baselineFile);
  const observed = collectObservedBrowserEntryDurations(resultsDir);
  const errors = [];

  for (const entry of entries) {
    const actual = observed.get(entry.id);
    if (!actual) {
      continue;
    }
    const planned = baseline.entries.get(entry.id)?.weight_ms;
    if (!Number.isInteger(planned) || planned <= 0) {
      errors.push(`missing browser entry baseline ${driftSubject(entry)}`);
      continue;
    }
    const kind = durationDriftKind(actual.duration_ms, planned);
    if (kind) {
      errors.push(
        durationDriftDescription(kind, {
          subject: driftSubject(entry),
          plannedMs: planned,
          actualMs: actual.duration_ms,
        }),
      );
    }
  }

  if (errors.length > 0) {
    process.stderr.write("Browser E2E duration baseline drift detected:\n");
    for (const error of errors) {
      process.stderr.write(`- ${error}\n`);
    }
    process.stderr.write(
      `Refresh from a successful browser run with: make browser-e2e-duration-baselines RESULTS_DIR=${resultsDir}\n`,
    );
    process.exit(1);
  }
  process.stdout.write(`Browser E2E duration baselines match ${observed.size} observed entry timings\n`);
}

function main(argv) {
  const [command, ...rest] = argv;
  switch (command) {
    case "plan": {
      const options = parsePlanArgs(rest);
      process.stdout.write(`${JSON.stringify(createPlan(options), null, 2)}\n`);
      return;
    }
    case "merge-reports": {
      const [outputFile, ...inputFiles] = rest;
      if (!outputFile || inputFiles.length === 0) {
        usage();
      }
      mergeReports(resolvePath(outputFile), inputFiles.map(resolvePath));
      return;
    }
    case "update-baselines":
      updateBaselines(rest);
      return;
    case "check-baseline-drift":
      checkBaselineDrift(rest);
      return;
    default:
      usage();
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exit(1);
  }
}
