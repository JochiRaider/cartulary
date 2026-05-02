#!/usr/bin/env node
import { existsSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  defaultExecutionTopologyManifestPath,
  loadExecutionTopology,
  renderServiceBackedScheduleProfile,
} from "./lib/execution-topology.mjs";
import {
  collectServiceTimingContamination,
  durationDriftDescription,
  durationDriftKind,
  formatContaminationReasons,
  printContaminationReasons,
} from "./lib/duration-drift.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const baselineSchemaID = "cartulary.service_backed_make_target_duration_baselines.v1";
const scheduleSchemaID = "cartulary.service_backed_schedule.v8";
const defaultBaselineFile = path.join(
  repoRoot,
  "tools",
  "service_backed_make_target_duration_baselines.json",
);
const defaultTopologyFile = defaultExecutionTopologyManifestPath;
const defaultScheduleManifestFile = path.join(repoRoot, "tools", "service_backed_schedule_manifest.json");
const baselineNote =
  "Service-backed scheduler make-target duration weights generated from successful scheduler artifacts. Refresh with make service-backed-make-target-duration-baselines RESULTS_DIR=<dir>.";
const defaultMakeTargetWeightMs = 10000;

function usage() {
  process.stderr.write(
    [
      "usage:",
      "  service-backed-make-target-durations.mjs update [--baseline-file <path>] <results-dir>",
      "  service-backed-make-target-durations.mjs check-drift [--baseline-file <path>] [--topology <path>] [--schedule-manifest <path>] <results-dir>",
    ].join("\n") + "\n",
  );
  process.exit(2);
}

function resolvePath(file) {
  return path.isAbsolute(file) ? file : path.join(repoRoot, file);
}

function rel(file) {
  return path.relative(repoRoot, file);
}

function readJSON(file) {
  return JSON.parse(readFileSync(resolvePath(file), "utf8"));
}

function sortedObject(entries) {
  return Object.fromEntries([...entries].sort(([left], [right]) => left.localeCompare(right)));
}

function positiveInteger(value, fallback) {
  return Number.isInteger(value) && value > 0 ? value : fallback;
}

function parseCommonArgs(argv, { includeTopology = false } = {}) {
  const options = {
    baselineFile: defaultBaselineFile,
    topologyFile: defaultTopologyFile,
    scheduleManifestFile: defaultScheduleManifestFile,
    resultsDir: "",
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
    if (includeTopology && arg === "--topology") {
      options.topologyFile = resolvePath(argv[index + 1] ?? "");
      index += 1;
      if (!options.topologyFile) {
        usage();
      }
      continue;
    }
    if (includeTopology && arg === "--schedule-manifest") {
      options.scheduleManifestFile = resolvePath(argv[index + 1] ?? "");
      index += 1;
      if (!options.scheduleManifestFile) {
        usage();
      }
      continue;
    }
    if (arg.startsWith("--")) {
      usage();
    }
    if (options.resultsDir) {
      usage();
    }
    options.resultsDir = resolvePath(arg);
  }
  if (!options.resultsDir) {
    usage();
  }
  return options;
}

function schedulerSummaryFiles(root) {
  const files = [];
  const stack = [resolvePath(root)];
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
      if (entry.isFile() && entry.name === "scheduler-summary.json") {
        files.push(next);
      }
    }
  }
  return files.sort();
}

function readSchedulerEvents(eventsFile) {
  if (!existsSync(eventsFile)) {
    return [];
  }
  return readFileSync(eventsFile, "utf8")
    .trim()
    .split(/\n/)
    .filter(Boolean)
    .map((line) => JSON.parse(line));
}

function collectObservedMakeTargetDurations(resultsDir) {
  const observed = new Map();
  let passedSchedulerCount = 0;
  for (const summaryFile of schedulerSummaryFiles(resultsDir)) {
    const summary = readJSON(summaryFile);
    if (summary.scheduler_kind !== "service-backed" || summary.status !== "pass") {
      continue;
    }
    const events = readSchedulerEvents(path.join(path.dirname(summaryFile), "scheduler-events.jsonl"));
    if (events.length === 0) {
      continue;
    }
    passedSchedulerCount += 1;
    const starts = new Map();
    for (const event of events) {
      if (event.event === "start" && typeof event.work_unit_id === "string") {
        starts.set(event.work_unit_id, event);
      }
    }
    for (const event of events) {
      if (event.event !== "finish" || event.status !== 0) {
        continue;
      }
      const start = starts.get(event.work_unit_id);
      if (start?.work_unit_type !== "make_target") {
        continue;
      }
      const target = start.aggregate_target || start.work_unit || event.work_unit;
      const durationMs = Math.max(1, Math.round(Number(event.duration_ms ?? 0)));
      if (typeof target === "string" && target !== "" && durationMs > 0) {
        observed.set(target, Math.max(observed.get(target) ?? 0, durationMs));
      }
    }
  }
  return { observed, passedSchedulerCount };
}

function readBaseline(file, { allowMissing = false } = {}) {
  const baselineFile = resolvePath(file);
  if (!existsSync(baselineFile)) {
    if (allowMissing) {
      return {
        schema_id: baselineSchemaID,
        note: baselineNote,
        default_make_target_weight_ms: defaultMakeTargetWeightMs,
        targets: {},
      };
    }
    throw new Error(`${rel(baselineFile)} is missing`);
  }
  const baseline = readJSON(baselineFile);
  if (baseline.schema_id !== baselineSchemaID) {
    throw new Error(`${rel(baselineFile)} must declare schema_id ${baselineSchemaID}`);
  }
  if (!baseline.targets || typeof baseline.targets !== "object" || Array.isArray(baseline.targets)) {
    throw new Error(`${rel(baselineFile)} targets must be an object`);
  }
  for (const [target, weight] of Object.entries(baseline.targets)) {
    if (!Number.isInteger(weight) || weight <= 0) {
      throw new Error(`${rel(baselineFile)} targets.${target} must be positive integer weight ms`);
    }
  }
  return baseline;
}

function readTopologyProfile(file) {
  const profile = renderServiceBackedScheduleProfile(loadExecutionTopology({ manifestPath: resolvePath(file) }));
  if (!profile.defaults || typeof profile.defaults !== "object" || Array.isArray(profile.defaults)) {
    throw new Error(`${rel(resolvePath(file))} service_backed_schedules.defaults must be an object`);
  }
  return profile;
}

function readScheduleMakeTargets(file) {
  const scheduleFile = resolvePath(file);
  const manifest = readJSON(scheduleFile);
  if (manifest.schema_id !== scheduleSchemaID) {
    throw new Error(`${rel(scheduleFile)} must declare schema_id ${scheduleSchemaID}`);
  }
  const targets = new Set();
  for (const schedule of manifest.schedules ?? []) {
    for (const source of schedule.work_unit_sources ?? []) {
      if (source?.type === "make_target" && typeof source.target === "string") {
        targets.add(source.target);
      }
    }
  }
  return targets;
}

function validatedOverrides(profile, scheduledTargets) {
  const raw = profile.defaults.make_target_weight_overrides ?? {};
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw new Error("defaults.make_target_weight_overrides must be an object when present");
  }
  const errors = [];
  const overrides = new Map();
  const now = Date.now();
  for (const [target, override] of Object.entries(raw)) {
    if (!scheduledTargets.has(target)) {
      errors.push(`override target=${target} is not present in generated service-backed schedules`);
      continue;
    }
    if (!override || typeof override !== "object" || Array.isArray(override)) {
      errors.push(`override target=${target} must be an object`);
      continue;
    }
    if (!Number.isInteger(override.weight_ms) || override.weight_ms <= 0) {
      errors.push(`override target=${target} must declare positive integer weight_ms`);
    }
    if (typeof override.reason !== "string" || override.reason.trim() === "") {
      errors.push(`override target=${target} must declare reason`);
    }
    if (typeof override.expires_at !== "string" || Number.isNaN(Date.parse(override.expires_at))) {
      errors.push(`override target=${target} must declare ISO expires_at`);
    } else if (Date.parse(override.expires_at) <= now) {
      errors.push(`override target=${target} expired at ${override.expires_at}`);
    }
    if (errors.length === 0 || !errors.some((error) => error.includes(`target=${target}`))) {
      overrides.set(target, override.weight_ms);
    }
  }
  return { overrides, errors };
}

function plannedWeight(target, baseline, overrides) {
  if (overrides.has(target)) {
    return overrides.get(target);
  }
  const weight = baseline.targets[target];
  if (Number.isInteger(weight) && weight > 0) {
    return weight;
  }
  return null;
}

function update(argv) {
  const options = parseCommonArgs(argv);
  const serviceContamination = collectServiceTimingContamination(repoRoot, options.resultsDir);
  if (serviceContamination.contaminated) {
    process.stderr.write("Refusing to refresh service-backed make-target duration baselines from contaminated service timing evidence:\n");
    printContaminationReasons(process.stderr, serviceContamination);
    process.stderr.write(`Inspect fixture costs with: make fixture-report RESULTS_DIR=${options.resultsDir}\n`);
    process.stderr.write("Rerun check-shaped evidence with: make check\n");
    process.exit(1);
  }
  const baseline = readBaseline(options.baselineFile, { allowMissing: true });
  const { observed, passedSchedulerCount } = collectObservedMakeTargetDurations(options.resultsDir);
  baseline.schema_id = baselineSchemaID;
  baseline.note = baselineNote;
  baseline.default_make_target_weight_ms = positiveInteger(
    baseline.default_make_target_weight_ms,
    defaultMakeTargetWeightMs,
  );
  baseline.targets ??= {};
  for (const [target, durationMs] of observed.entries()) {
    baseline.targets[target] = durationMs;
  }
  baseline.updated_at = new Date().toISOString();
  baseline.targets = sortedObject(Object.entries(baseline.targets));
  writeFileSync(options.baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  process.stdout.write(
    `updated ${observed.size} service-backed make-target duration baselines from ${passedSchedulerCount} successful scheduler artifact(s)\n`,
  );
}

function checkDrift(argv) {
  const options = parseCommonArgs(argv, { includeTopology: true });
  const baseline = readBaseline(options.baselineFile);
  const profile = readTopologyProfile(options.topologyFile);
  const scheduledTargets = readScheduleMakeTargets(options.scheduleManifestFile);
  const { overrides, errors } = validatedOverrides(profile, scheduledTargets);
  const missingBaselines = [];
  const driftErrors = [];
  const warnings = [];
  const serviceContamination = collectServiceTimingContamination(repoRoot, options.resultsDir);
  const { observed, passedSchedulerCount } = collectObservedMakeTargetDurations(options.resultsDir);

  for (const [target, actual] of observed.entries()) {
    const planned = plannedWeight(target, baseline, overrides);
    if (!planned) {
      missingBaselines.push(`missing make-target baseline target=${target} actual_ms=${actual}`);
      continue;
    }
    const kind = durationDriftKind(actual, planned);
    if (!kind) {
      continue;
    }
    const description = durationDriftDescription(kind, {
      subject: `target=${target}`,
      plannedMs: planned,
      actualMs: actual,
    });
    if (kind === "underplanned" && serviceContamination.contaminated) {
      warnings.push(
        `ignored ${description.replace(`${kind} `, `${kind} contaminated `)} service_timing_contamination=[${formatContaminationReasons(serviceContamination)}]`,
      );
      continue;
    }
    driftErrors.push(description);
  }

  if (warnings.length > 0) {
    process.stderr.write("Service-backed make-target duration baseline drift ignored contaminated timing evidence:\n");
    for (const warning of warnings) {
      process.stderr.write(`- ${warning}\n`);
    }
    process.stderr.write("Service timing contamination detected:\n");
    printContaminationReasons(process.stderr, serviceContamination);
    process.stderr.write("Rerun after clean service timing evidence; do not refresh baselines from contaminated timing evidence.\n");
  }

  if (errors.length > 0 || missingBaselines.length > 0 || driftErrors.length > 0) {
    process.stderr.write("Service-backed make-target duration baseline drift detected:\n");
    if (errors.length > 0) {
      process.stderr.write("Configuration errors:\n");
      for (const error of errors) {
        process.stderr.write(`- ${error}\n`);
      }
    }
    if (missingBaselines.length > 0) {
      process.stderr.write("Missing baseline components:\n");
      for (const error of missingBaselines) {
        process.stderr.write(`- ${error}\n`);
      }
    }
    if (driftErrors.length > 0) {
      process.stderr.write("Observed timing drift:\n");
      for (const error of driftErrors) {
        process.stderr.write(`- ${error}\n`);
      }
    }
    process.stderr.write(`Inspect fixture costs with: make fixture-report RESULTS_DIR=${options.resultsDir}\n`);
    process.stderr.write("Rerun check-shaped evidence with: make check\n");
    process.stderr.write(
      `Refresh from a successful service-backed run with: make service-backed-make-target-duration-baselines RESULTS_DIR=${options.resultsDir}\n`,
    );
    process.exit(1);
  }
  process.stdout.write(
    `Service-backed make-target duration baselines match ${observed.size} observed make target(s) from ${passedSchedulerCount} successful scheduler artifact(s)\n`,
  );
}

try {
  const [command, ...rest] = process.argv.slice(2);
  if (command === "update") {
    update(rest);
  } else if (command === "check-drift") {
    checkDrift(rest);
  } else {
    usage();
  }
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
