#!/usr/bin/env node
import { existsSync, readdirSync, readFileSync, statSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import {
  durationDriftDescription,
  durationDriftKind,
} from "../duration-accounting/duration-drift.mjs";
import {
  browserDurationBaselineEntries,
  selectedEntriesForPlan,
} from "./browser-duration-accounting.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import { buildSourceSnapshot } from "../owner-slice/source-snapshot.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const baselineSchemaID = "cartulary.browser_e2e_duration_baselines.v3";
const shardPlanSchemaID = "cartulary.browser_e2e_shard_plan.v2";
const defaultBaselineFile = path.join(
  repoRoot,
  "tools",
  "browser_e2e_duration_baselines.json",
);
const baselineNote =
  "Advisory browser functional manifest-entry and file-overhead weights for duration-balanced Playwright sharding. Refresh with make browser-e2e-duration-baselines RESULTS_DIR=<dir>.";
const defaultEntryWeightMs = 10000;
const defaultFileOverheadMs = 2500;
const defaultShardTargetMs = 12000;
const legacyBrowserSelectionIDPattern =
  /^(?:E-[0-9]+-[A-Z0-9]+(?:-[A-Z0-9]+)*|FE-(?:U|I|B|E|V|A11Y|S)-P(?:0|[1-9][0-9]*)-[0-9]{2})$/u;
const semanticBrowserRowIDPattern =
  /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*){2,}$/u;

function usage() {
  process.stderr.write(
    [
      "usage:",
      "  browser-shard-plan.mjs plan [--baseline-file <path>] [--min-shards <n>] [--max-shards <n>] [--frontend-row-ids <ids>] [--entry-ids <ids>] [--runtime-profile-id <id>] [--single-shard-name <name>]",
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
  return Object.fromEntries(
    [...entries].sort(([left], [right]) => left.localeCompare(right)),
  );
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

function defaultBaselineDocument() {
  return {
    schema_id: baselineSchemaID,
    note: baselineNote,
    default_entry_weight_ms: defaultEntryWeightMs,
    file_overhead_ms: defaultFileOverheadMs,
    shard_target_ms: defaultShardTargetMs,
    entries: {},
  };
}

function positiveIntegerOrDefault(value, fallback) {
  return Number.isInteger(value) && value > 0 ? value : fallback;
}

function nonNegativeIntegerOrDefault(value, fallback) {
  return Number.isInteger(value) && value >= 0 ? value : fallback;
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
    throw new Error(
      `${path.relative(repoRoot, file)} must declare schema_id ${baselineSchemaID}`,
    );
  }
  const rawEntries = baseline.entries ?? {};
  if (
    !rawEntries ||
    typeof rawEntries !== "object" ||
    Array.isArray(rawEntries)
  ) {
    throw new Error(
      `${path.relative(repoRoot, file)} entries must be an object`,
    );
  }
  for (const [id, entry] of Object.entries(rawEntries)) {
    if (!semanticBrowserRowIDPattern.test(id)) {
      throw new Error(
        `${path.relative(repoRoot, file)} entries key ${id} must be a semantic catalog row ID`,
      );
    }
    if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
      throw new Error(
        `${path.relative(repoRoot, file)} entries.${id} must be an object`,
      );
    }
    if (normalizeManifestFile(entry.file) === "") {
      throw new Error(
        `${path.relative(repoRoot, file)} entries.${id}.file must be non-empty`,
      );
    }
    if (typeof entry.title !== "string" || entry.title.trim() === "") {
      throw new Error(
        `${path.relative(repoRoot, file)} entries.${id}.title must be non-empty`,
      );
    }
    if (!Number.isInteger(entry.weight_ms) || entry.weight_ms <= 0) {
      throw new Error(
        `${path.relative(repoRoot, file)} entries.${id}.weight_ms must be a positive integer`,
      );
    }
  }
  return baseline;
}

function compareEntries(left, right) {
  if (left.stage !== right.stage) {
    return left.stage.localeCompare(right.stage, undefined, { numeric: true });
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

export function parseFrontendRowIDs(value) {
  return new Set(
    String(value ?? "")
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean),
  );
}

export function parseBrowserEntryIDs(value) {
  return new Set(
    String(value ?? "")
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean),
  );
}

function readBaseline(file, activeEntries) {
  const baseline = readBaselineDocument(file);
  const activeByID = new Map(activeEntries.map((entry) => [entry.id, entry]));
  const entries = baselineEntryMap(baseline.entries ?? {});
  for (const [id, baselineEntry] of entries) {
    const active = activeByID.get(id);
    if (!active) {
      throw new Error(
        `${path.relative(repoRoot, file)} contains retired browser entry baseline id=${id}`,
      );
    }
    if (
      active.file !== baselineEntry.file ||
      active.title !== baselineEntry.title
    ) {
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
    fileOverheadMs: nonNegativeIntegerOrDefault(
      baseline.file_overhead_ms,
      defaultFileOverheadMs,
    ),
    shardTargetMs: positiveIntegerOrDefault(
      baseline.shard_target_ms,
      defaultShardTargetMs,
    ),
    entries,
  };
}

function uniqueSortedFiles(entries) {
  return [...new Set(entries.map((entry) => entry.file))].sort();
}

function planShardCount(entries, { minShards, maxShards, fileOverheadMs, shardTargetMs }) {
  if (entries.length === 0) {
    return 0;
  }
  const totalWeight =
    entries.reduce((sum, entry) => sum + entry.weight_ms, 0) +
    uniqueSortedFiles(entries).length * fileOverheadMs;
  const targetCount = Math.ceil(totalWeight / Math.max(1, shardTargetMs));
  return Math.max(
    1,
    Math.min(entries.length, maxShards, Math.max(minShards, targetCount)),
  );
}

export function createPlanFromEntries({
  baselineFile,
  minShards = 1,
  maxShards,
  stage = "",
  frontendRowIDs = new Set(),
  selectedEntryIDs = new Set(),
  runtimeProfileID = "default",
  shardNamePrefix = runtimeProfileID === "default"
    ? "browser-functional"
    : `browser-functional-${runtimeProfileID.replaceAll("_", "-")}`,
  singleShardName = "",
  baselineEntries = [],
  selectedEntries = [],
}) {
  const baseline = readBaseline(
    baselineFile,
    baselineEntries,
  );
  const entries = selectedEntries
    .filter((entry) => (entry.runtime_profile_id ?? "default") === runtimeProfileID)
    .filter((entry) => frontendRowIDs.size === 0 || frontendRowIDs.has(entry.id))
    .filter((entry) => selectedEntryIDs.size === 0 || selectedEntryIDs.has(entry.id))
    .map((entry) => ({
      ...entry,
      weight_ms:
        baseline.entries.get(entry.id)?.weight_ms ??
        baseline.defaultEntryWeightMs,
    }))
    .sort(compareEntries);
  if (entries.length === 0) {
    let message = stage
      ? `no authoritative browser_functional Playwright rows found for ${stage}`
      : "no authoritative browser_functional Playwright rows found";
    if (frontendRowIDs.size > 0) {
      message = `no browser_functional Playwright rows found for selected frontend row id(s): ${[...frontendRowIDs].sort().join(",")}`;
    } else if (selectedEntryIDs.size > 0) {
      message = `no browser_functional Playwright rows found for selected entry id(s): ${[...selectedEntryIDs].sort().join(",")}`;
    }
    throw new Error(message);
  }
  if (selectedEntryIDs.size > 0) {
    const found = new Set(selectedEntries.map((entry) => entry.id));
    const unknown = [...selectedEntryIDs]
      .filter((id) => !found.has(id))
      .sort();
    if (unknown.length > 0) {
      throw new Error(`selected browser entry id(s) not found: ${unknown.join(",")}`);
    }
  }
  if (singleShardName) {
    const singleShardEntries = [...entries].sort(compareEntries);
    const files = uniqueSortedFiles(singleShardEntries);
    const weightMs =
      singleShardEntries.reduce((sum, entry) => sum + entry.weight_ms, 0) +
      files.length * baseline.fileOverheadMs;
    return {
      schema_id: shardPlanSchemaID,
      stage,
      generated_at: new Date().toISOString(),
      baseline_file: path.relative(repoRoot, baselineFile),
      min_shards: minShards,
      max_shards: maxShards,
      shard_count: 1,
      default_entry_weight_ms: baseline.defaultEntryWeightMs,
      file_overhead_ms: baseline.fileOverheadMs,
      shard_target_ms: baseline.shardTargetMs,
      entry_count: singleShardEntries.length,
      file_count: files.length,
      files,
      entries: singleShardEntries,
      shards: [
        {
          name: singleShardName,
          weight_ms: weightMs,
          entry_count: singleShardEntries.length,
          file_count: files.length,
          files,
          stages: [...new Set(singleShardEntries.map((entry) => entry.stage))].sort((left, right) =>
            left.localeCompare(right, undefined, { numeric: true }),
          ),
          grep: exactAlternationRegex(
            singleShardEntries.flatMap((entry) => entry.titles ?? [entry.title]),
          ),
          entries: singleShardEntries.map((entry) => ({
            id: entry.id,
            stage: entry.stage,
            file: entry.file,
            title: entry.title,
            titles: entry.titles ?? [entry.title],
            weight_ms: entry.weight_ms,
          })),
        },
      ],
    };
  }
  const shardCount = planShardCount(entries, {
    minShards,
    maxShards,
    fileOverheadMs: baseline.fileOverheadMs,
    shardTargetMs: baseline.shardTargetMs,
  });
  const shards = Array.from({ length: shardCount }, (_, index) => ({
    name: `${shardNamePrefix}-shard-${String(index + 1).padStart(2, "0")}`,
    weight_ms: 0,
    files: new Set(),
    stages: new Set(),
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
      .sort(
        (left, right) => {
          const leftProjected =
            left.weight_ms + entry.weight_ms + (left.files.has(entry.file) ? 0 : baseline.fileOverheadMs);
          const rightProjected =
            right.weight_ms + entry.weight_ms + (right.files.has(entry.file) ? 0 : baseline.fileOverheadMs);
          return leftProjected - rightProjected || left.name.localeCompare(right.name);
        },
      )[0];
    shard.weight_ms += entry.weight_ms + (shard.files.has(entry.file) ? 0 : baseline.fileOverheadMs);
    shard.files.add(entry.file);
    shard.stages.add(entry.stage);
    shard.entries.push(entry);
  }

  return {
    schema_id: shardPlanSchemaID,
    stage,
    generated_at: new Date().toISOString(),
    baseline_file: path.relative(repoRoot, baselineFile),
    min_shards: minShards,
    max_shards: maxShards,
    shard_count: shardCount,
    default_entry_weight_ms: baseline.defaultEntryWeightMs,
    file_overhead_ms: baseline.fileOverheadMs,
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
        stages: [...shard.stages].sort((left, right) =>
          left.localeCompare(right, undefined, { numeric: true }),
        ),
        grep: exactAlternationRegex(
          shardEntries.flatMap((entry) => entry.titles ?? [entry.title]),
        ),
        entries: shardEntries.map((entry) => ({
          id: entry.id,
          stage: entry.stage,
          file: entry.file,
          title: entry.title,
          titles: entry.titles ?? [entry.title],
          weight_ms: entry.weight_ms,
        })),
      };
    }),
  };
}

export function createPlan(options) {
  if (Array.isArray(options?.baselineEntries) && Array.isArray(options?.selectedEntries)) {
    return createPlanFromEntries(options);
  }
  throw new Error(
    "createPlan requires explicit baselineEntries and selectedEntries; use tools/harness/browser/browser-duration-accounting.mjs when catalog discovery is needed",
  );
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
    frontendRowIDs: new Set(),
    selectedEntryIDs: new Set(),
    singleShardName: "",
    runtimeProfileID: "default",
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
    if (arg === "--frontend-row-ids") {
      options.frontendRowIDs = parseFrontendRowIDs(argv[index + 1] ?? "");
      index += 1;
      if (options.frontendRowIDs.size === 0) {
        usage();
      }
      continue;
    }
    if (arg === "--entry-ids") {
      options.selectedEntryIDs = parseBrowserEntryIDs(argv[index + 1] ?? "");
      index += 1;
      if (options.selectedEntryIDs.size === 0) {
        usage();
      }
      for (const id of options.selectedEntryIDs) {
        if (
          !legacyBrowserSelectionIDPattern.test(id) &&
          !semanticBrowserRowIDPattern.test(id)
        ) {
          usage();
        }
      }
      continue;
    }
    if (arg === "--single-shard-name") {
      options.singleShardName = argv[index + 1] ?? "";
      index += 1;
      if (!/^browser-functional(?:-[a-z0-9-]+)?-shard-[0-9]{2}$/.test(options.singleShardName)) {
        usage();
      }
      continue;
    }
    if (arg === "--runtime-profile-id") {
      options.runtimeProfileID = argv[index + 1] ?? "";
      index += 1;
      if (!/^[a-z][a-z0-9_]*$/u.test(options.runtimeProfileID)) {
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

function collectObservedBrowserEntryDurations(
  resultsDir,
) {
  const catalog = loadTestCatalog(repoRoot);
  const currentIdentity = {
    source_snapshot_digest: buildSourceSnapshot(repoRoot).digest,
    catalog_semantic_digest: catalog.summary.catalog_semantic_digest,
    verification_semantic_digest: catalog.summary.verification_semantic_digest,
  };
  const authoritative = new Map(
    browserDurationBaselineEntries(repoRoot).map((entry) => [entry.id, entry]),
  );
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
      if (!entry.isFile() || entry.name !== "test-evidence-accounting.json") {
        continue;
      }
      let accounting;
      try {
        accounting = readJSON(next);
        validateSchemaSync("cartulary.test_evidence_accounting.v1", accounting);
      } catch {
        continue;
      }
      if (
        accounting.status !== "pass" ||
        accounting.target_id !== "browser-e2e-webserver-backed" ||
        Object.entries(currentIdentity).some(
          ([field, expected]) => accounting[field] !== expected,
        )
      ) {
        continue;
      }
      const expectedByID = new Map(
        accounting.expected_rows.map((row) => [row.row_id, row]),
      );
      for (const row of accounting.observed_rows) {
        const active = authoritative.get(row.row_id);
        const expected = expectedByID.get(row.row_id);
        const durationMs = positiveIntegerOrDefault(row.executed_duration_ms, 0);
        if (
          !active ||
          expected?.evidence_target_id !== "browser-e2e-webserver-backed" ||
          !["passed", "skipped_authorized"].includes(row.terminal_state) ||
          durationMs <= 0
        ) {
          continue;
        }
        const currentObserved = observed.get(row.row_id);
        const normalized = {
          id: row.row_id,
          stage: active.stage,
          file: active.file,
          title: active.title,
          duration_ms: durationMs,
        };
        if (!currentObserved || durationMs > currentObserved.duration_ms) {
          observed.set(row.row_id, normalized);
        }
      }
    }
  }
  return observed;
}

function relativeResultsPath(resultsDir, file = resultsDir) {
  return path.relative(resultsDir, file) || ".";
}

function summaryStatusIsFailure(status) {
  return ["fail", "failed", "error"].includes(String(status ?? "").toLowerCase());
}

function browserRetainedRunSafetyErrors(resultsDir) {
  const errors = [];
  try {
    if (!statSync(resultsDir).isDirectory()) {
      return [`retained results path is not a directory: ${resultsDir}`];
    }
  } catch {
    return [`retained results directory is missing: ${resultsDir}`];
  }

  const stack = [resultsDir];
  while (stack.length > 0) {
    const current = stack.pop();
    let entries = [];
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch (error) {
      errors.push(`cannot read retained results path ${relativeResultsPath(resultsDir, current)}: ${error.message}`);
      continue;
    }
    for (const entry of entries) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (!entry.isFile()) {
        continue;
      }
      if (entry.name === "stack.tainted" || entry.name.endsWith(".tainted")) {
        errors.push(`tainted browser stack marker: ${relativeResultsPath(resultsDir, next)}`);
        continue;
      }
      if (entry.name !== "target-summary.json" && entry.name !== "scheduler-summary.json") {
        continue;
      }
      let summary;
      try {
        summary = readJSON(next);
      } catch (error) {
        errors.push(`malformed retained summary ${relativeResultsPath(resultsDir, next)}: ${error.message}`);
        continue;
      }
      if (summaryStatusIsFailure(summary.status)) {
        errors.push(
          `failed retained summary ${relativeResultsPath(resultsDir, next)} status=${summary.status}`,
        );
      }
    }
  }
  return errors.sort();
}

function assertBrowserRetainedRunSafe(resultsDir) {
  const errors = browserRetainedRunSafetyErrors(resultsDir);
  if (errors.length > 0) {
    throw new Error(
      `unsafe retained browser duration evidence:\n${errors.map((error) => `- ${error}`).join("\n")}`,
    );
  }
}

function retainedRunIsDefaultCheck(resultsDir) {
  for (const summaryFile of [
    path.join(resultsDir, "tool-run-summary.json"),
    path.join(resultsDir, "check", "tool-run-summary.json"),
  ]) {
    if (!existsSync(summaryFile)) continue;
    try {
      const summary = readJSON(summaryFile);
      if (summary.target === "check" && summary.status === "pass") return true;
    } catch {
      continue;
    }
  }
  return false;
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

function activeEntryRowsForBaseline(baselineFile, activeEntries) {
  const baseline = readBaseline(baselineFile, activeEntries);
  return {
    baseline,
    entries: activeEntries
      .map((entry) => ({
        ...entry,
        weight_ms:
          baseline.entries.get(entry.id)?.weight_ms ??
          baseline.defaultEntryWeightMs,
      }))
      .sort(compareEntries),
  };
}

export function updateBaselinesFromEntries(argv, authoritativeEntries) {
  const { baselineFile, resultsDir } = parseBaselineResultsArgs(argv);
  assertBrowserRetainedRunSafe(resultsDir);
  const baseline = readBaselineDocument(baselineFile, { allowMissing: true });
  const existingEntries = baselineEntryMap(baseline.entries ?? {});
  const observed = collectObservedBrowserEntryDurations(resultsDir);
  const allowPartialRefresh = retainedRunIsDefaultCheck(resultsDir);
  const missingObserved = authoritativeEntries.filter(
    (entry) => !observed.has(entry.id),
  );
  if (missingObserved.length > 0 && !allowPartialRefresh) {
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
  baseline.file_overhead_ms = nonNegativeIntegerOrDefault(
    baseline.file_overhead_ms,
    defaultFileOverheadMs,
  );
  delete baseline.default_spec_weight_ms;
  delete baseline.specs;
  baseline.shard_target_ms = positiveIntegerOrDefault(
    baseline.shard_target_ms,
    defaultShardTargetMs,
  );
  baseline.updated_at = new Date().toISOString();
  const nextEntries = [];
  let observedEntryCount = 0;
  for (const entry of authoritativeEntries) {
    const actual = observed.get(entry.id);
    if (actual) {
      observedEntryCount += 1;
      nextEntries.push([
        entry.id,
        {
          file: entry.file,
          title: entry.title,
          weight_ms: actual.duration_ms,
        },
      ]);
      continue;
    }
    const existing = existingEntries.get(entry.id);
    if (allowPartialRefresh && existing) {
      nextEntries.push([
        entry.id,
        {
          file: entry.file,
          title: entry.title,
          weight_ms: existing.weight_ms,
        },
      ]);
    }
  }
  baseline.entries = sortedObject(nextEntries);

  writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  const preservedCount = nextEntries.length - observedEntryCount;
  const partialNote = allowPartialRefresh
    ? `; preserved ${preservedCount} active unobserved baseline entries from the prior baseline`
    : "";
  process.stdout.write(
    `updated ${observedEntryCount} browser E2E row duration baselines from ${path.relative(repoRoot, resultsDir)}${partialNote}\n`,
  );
}

function driftSubject(entry) {
  return `id=${entry.id} file=${entry.file} title=${JSON.stringify(entry.title)}`;
}

export function checkBaselineDriftFromEntries(argv, activeEntries) {
  const { baselineFile, resultsDir } = parseBaselineResultsArgs(argv);
  assertBrowserRetainedRunSafe(resultsDir);
  const { baseline, entries } = activeEntryRowsForBaseline(baselineFile, activeEntries);
  const observed = collectObservedBrowserEntryDurations(resultsDir);
  const errors = [];

  if (observed.size === 0) {
    throw new Error("retained browser duration evidence has no observed Playwright timings");
  }

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
  process.stdout.write(
    `Browser E2E duration baselines match ${observed.size} observed entry timings\n`,
  );
}

function createDiscoveredPlan(options) {
  return createPlanFromEntries({
    ...options,
    baselineEntries: browserDurationBaselineEntries(repoRoot),
    selectedEntries: selectedEntriesForPlan(repoRoot, options),
  });
}

function updateDiscoveredBaselines(argv) {
  updateBaselinesFromEntries(argv, browserDurationBaselineEntries(repoRoot));
}

function checkDiscoveredBaselineDrift(argv) {
  checkBaselineDriftFromEntries(argv, browserDurationBaselineEntries(repoRoot));
}

async function main(argv) {
  const [command, ...rest] = argv;
  switch (command) {
    case "plan": {
      const options = parsePlanArgs(rest);
      process.stdout.write(
        `${JSON.stringify(createDiscoveredPlan(options), null, 2)}\n`,
      );
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
      updateDiscoveredBaselines(rest);
      return;
    case "check-baseline-drift":
      checkDiscoveredBaselineDrift(rest);
      return;
    default:
      usage();
  }
}

if (
  process.argv[1] &&
  import.meta.url === pathToFileURL(process.argv[1]).href
) {
  try {
    await main(process.argv.slice(2));
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exit(1);
  }
}
