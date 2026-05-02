#!/usr/bin/env node
import { existsSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  collectEntries,
  loadManifest,
  phaseManifestNames,
} from "./phase-manifest.mjs";
import {
  durationDriftDescription,
  durationDriftKind,
} from "./duration-drift.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..");
const baselineSchemaID = "cartulary.browser_e2e_duration_baselines.v1";
const defaultBaselineFile = path.join(repoRoot, "tools", "browser_e2e_duration_baselines.json");
const baselineNote =
  "Advisory browser functional spec weights for duration-balanced Playwright sharding. Refresh with make browser-e2e-duration-baselines RESULTS_DIR=<dir>.";
const defaultSpecWeightMs = 10000;
const defaultShardTargetMs = 12000;

function usage() {
  process.stderr.write(
    [
      "usage:",
      "  browser-shard-plan.mjs plan [--baseline-file <path>] [--max-shards <n>]",
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
    default_spec_weight_ms: defaultSpecWeightMs,
    shard_target_ms: defaultShardTargetMs,
    specs: {},
  };
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
  const rawSpecs = baseline.specs ?? {};
  if (!rawSpecs || typeof rawSpecs !== "object" || Array.isArray(rawSpecs)) {
    throw new Error(`${path.relative(repoRoot, file)} specs must be an object`);
  }
  return baseline;
}

function readBaseline(file) {
  const baseline = readBaselineDocument(file);
  const rawSpecs = baseline.specs ?? {};
  return {
    defaultSpecWeightMs: positiveIntegerOrDefault(
      baseline.default_spec_weight_ms,
      defaultSpecWeightMs,
    ),
    shardTargetMs: positiveIntegerOrDefault(baseline.shard_target_ms, defaultShardTargetMs),
    specs: new Map(
      Object.entries(rawSpecs).map(([specFile, weight]) => [
        normalizeManifestFile(specFile),
        positiveIntegerOrDefault(weight, defaultSpecWeightMs),
      ]),
    ),
  };
}

function positiveIntegerOrDefault(value, fallback) {
  return Number.isInteger(value) && value > 0 ? value : fallback;
}

function browserFunctionalEntries(root, { phase: phaseFilter = "" } = {}) {
  const entries = [];
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
        entry.execution_dependency === "browser_functional"
      ) {
        entries.push({
          id: entry.id,
          phase,
          file: normalizeManifestFile(entry.file),
          title: entry.title,
        });
      }
    }
  }
  return entries.sort((left, right) => {
    if (left.phase !== right.phase) {
      return left.phase.localeCompare(right.phase, undefined, { numeric: true });
    }
    if (left.file !== right.file) {
      return left.file.localeCompare(right.file);
    }
    return left.title.localeCompare(right.title);
  });
}

function collectSpecRows(root, baseline, { phase = "" } = {}) {
  const specs = new Map();
  for (const entry of browserFunctionalEntries(root, { phase })) {
    if (!specs.has(entry.file)) {
      specs.set(entry.file, {
        file: entry.file,
        weight_ms: baseline.specs.get(entry.file) ?? baseline.defaultSpecWeightMs,
        phases: new Set(),
        entries: [],
      });
    }
    const spec = specs.get(entry.file);
    spec.phases.add(entry.phase);
    spec.entries.push(entry);
  }
  return [...specs.values()]
    .map((spec) => ({
      ...spec,
      phases: [...spec.phases].sort((left, right) =>
        left.localeCompare(right, undefined, { numeric: true }),
      ),
      entries: spec.entries.sort((left, right) => left.title.localeCompare(right.title)),
    }))
    .sort((left, right) => left.file.localeCompare(right.file));
}

function planShardCount(specs, baseline, maxShards) {
  if (specs.length === 0) {
    return 0;
  }
  const totalWeight = specs.reduce((total, spec) => total + spec.weight_ms, 0);
  const targetShards = Math.max(1, Math.ceil(totalWeight / baseline.shardTargetMs));
  return Math.max(1, Math.min(specs.length, maxShards, targetShards));
}

function createPlan({ baselineFile, maxShards, phase = "" }) {
  const baseline = readBaseline(baselineFile);
  const specs = collectSpecRows(repoRoot, baseline, { phase });
  if (specs.length === 0) {
    throw new Error(
      phase
        ? `no authoritative browser_functional Playwright rows found for ${phase}`
        : "no authoritative browser_functional Playwright rows found",
    );
  }
  const shardCount = planShardCount(specs, baseline, maxShards);
  const shards = Array.from({ length: shardCount }, (_, index) => ({
    name: `browser-functional-shard-${String(index + 1).padStart(2, "0")}`,
    weight_ms: 0,
    files: [],
    phases: new Set(),
    entries: [],
  }));

  const weightedSpecs = [...specs].sort(
    (left, right) => right.weight_ms - left.weight_ms || left.file.localeCompare(right.file),
  );
  for (const spec of weightedSpecs) {
    const shard = shards
      .slice()
      .sort((left, right) => left.weight_ms - right.weight_ms || left.name.localeCompare(right.name))[0];
    shard.weight_ms += spec.weight_ms;
    shard.files.push(spec.file);
    for (const phase of spec.phases) {
      shard.phases.add(phase);
    }
    shard.entries.push(...spec.entries);
  }

  return {
    schema_id: "cartulary.browser_e2e_shard_plan.v1",
    phase,
    generated_at: new Date().toISOString(),
    baseline_file: path.relative(repoRoot, baselineFile),
    max_shards: maxShards,
    shard_count: shardCount,
    default_spec_weight_ms: baseline.defaultSpecWeightMs,
    shard_target_ms: baseline.shardTargetMs,
    spec_count: specs.length,
    specs,
    shards: shards.map((shard) => ({
      name: shard.name,
      weight_ms: shard.weight_ms,
      files: shard.files.sort(),
      phases: [...shard.phases].sort((left, right) =>
        left.localeCompare(right, undefined, { numeric: true }),
      ),
      grep: alternationRegex(shard.entries.map((entry) => entry.title)),
      entries: shard.entries
        .sort((left, right) => {
          if (left.file !== right.file) {
            return left.file.localeCompare(right.file);
          }
          return left.title.localeCompare(right.title);
        })
        .map((entry) => ({
          id: entry.id,
          phase: entry.phase,
          file: entry.file,
          title: entry.title,
        })),
    })),
  };
}

function escapeRegex(value) {
  return value.replace(/[\\^$.*+?()[\]{}|]/g, "\\$&");
}

function alternationRegex(values) {
  if (values.length === 0) {
    return "(?!)";
  }
  if (values.length === 1) {
    return escapeRegex(values[0]);
  }
  return `(${values.map(escapeRegex).join("|")})`;
}

function parsePlanArgs(argv) {
  const options = {
    baselineFile: defaultBaselineFile,
    maxShards: 1,
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

function collectObservedBrowserSpecDurations(resultsDir, { requirePassingPhaseSummary = true } = {}) {
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
      for (const fileTiming of timing.files ?? []) {
        const file = normalizePlaywrightReportFile(fileTiming.file);
        const durationMs = positiveIntegerOrDefault(
          fileTiming.wall_duration_ms || fileTiming.executed_duration_ms,
          0,
        );
        if (durationMs > 0) {
          observed.set(file, Math.max(observed.get(file) ?? 0, durationMs));
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

function updateBaselines(argv) {
  const { baselineFile, resultsDir } = parseBaselineResultsArgs(argv);
  const baseline = readBaselineDocument(baselineFile, { allowMissing: true });
  const authoritativeSpecs = collectSpecRows(repoRoot, readBaseline(baselineFile)).map((spec) => spec.file);
  const observed = collectObservedBrowserSpecDurations(resultsDir);
  const missingObserved = authoritativeSpecs.filter((specFile) => !observed.has(specFile));
  if (missingObserved.length > 0) {
    throw new Error(
      `missing observed browser spec timings: ${missingObserved.join(", ")}`,
    );
  }

  baseline.schema_id = baselineSchemaID;
  baseline.note = baselineNote;
  baseline.default_spec_weight_ms = positiveIntegerOrDefault(
    baseline.default_spec_weight_ms,
    defaultSpecWeightMs,
  );
  baseline.shard_target_ms = positiveIntegerOrDefault(baseline.shard_target_ms, defaultShardTargetMs);
  baseline.updated_at = new Date().toISOString();
  baseline.specs = sortedObject(authoritativeSpecs.map((specFile) => [specFile, observed.get(specFile)]));

  writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  process.stdout.write(
    `updated ${authoritativeSpecs.length} browser E2E duration baselines from ${path.relative(repoRoot, resultsDir)}\n`,
  );
}

function checkBaselineDrift(argv) {
  const { baselineFile, resultsDir } = parseBaselineResultsArgs(argv);

  const baseline = readBaseline(baselineFile);
  const observed = collectObservedBrowserSpecDurations(resultsDir);
  const authoritativeSpecs = new Set(collectSpecRows(repoRoot, baseline).map((spec) => spec.file));
  const errors = [];

  for (const specFile of authoritativeSpecs) {
    const actual = observed.get(specFile);
    if (!actual) {
      continue;
    }
    const planned = baseline.specs.get(specFile);
    if (!Number.isInteger(planned) || planned <= 0) {
      errors.push(`missing browser spec baseline file=${specFile}`);
      continue;
    }
    const kind = durationDriftKind(actual, planned);
    if (kind) {
      errors.push(
        durationDriftDescription(kind, {
          subject: `file=${specFile}`,
          plannedMs: planned,
          actualMs: actual,
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
  process.stdout.write(`Browser E2E duration baselines match ${observed.size} observed spec timings\n`);
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

try {
  main(process.argv.slice(2));
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
