#!/usr/bin/env node
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

import {
  defaultExecutionTopologyManifestPath,
  loadExecutionTopology,
  renderServiceBackedScheduleProfile,
} from "./lib/execution-topology.mjs";
import {
  durationBaselineCliContext,
  parseDurationBaselineResultsArgs,
} from "./lib/duration-baseline-cli.mjs";
import {
  collectServiceTimingContamination,
  durationDriftDescription,
  durationDriftKind,
  formatContaminationReasons,
  printContaminationReasons,
} from "./lib/duration-drift.mjs";
import { findFilesNamed } from "./lib/result-artifacts.mjs";
import {
  readJSON,
  sortedObjectByKey,
} from "./lib/target-duration-baselines.mjs";

const { repoRoot, resolvePath, rel } = durationBaselineCliContext(import.meta.url);
const baselineSchemaID = "cartulary.scheduler_work_unit_duration_baselines.v1";
const scheduleSchemaID = "cartulary.service_backed_schedule.v10";
const defaultBaselineFile = path.join(
  repoRoot,
  "tools",
  "service_backed_make_target_duration_baselines.json",
);
const defaultTopologyFile = defaultExecutionTopologyManifestPath;
const defaultScheduleManifestFile = path.join(repoRoot, "tools", "service_backed_schedule_manifest.json");
const baselineNote =
  "Scheduler work-unit duration weights generated from successful scheduler artifacts. Refresh with make service-backed-make-target-duration-baselines RESULTS_DIR=<dir>.";
const defaultWorkUnitWeightMs = 10000;

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

function positiveInteger(value, fallback) {
  return Number.isInteger(value) && value > 0 ? value : fallback;
}

function parseCommonArgs(argv, { includeTopology = false } = {}) {
  return parseDurationBaselineResultsArgs(argv, {
    usage,
    resolvePath,
    baselineFile: defaultBaselineFile,
    flags: includeTopology
      ? [
          {
            flag: "--topology",
            name: "topologyFile",
            defaultValue: defaultTopologyFile,
          },
          {
            flag: "--schedule-manifest",
            name: "scheduleManifestFile",
            defaultValue: defaultScheduleManifestFile,
          },
        ]
      : [],
  });
}

function schedulerSummaryFiles(root) {
  return findFilesNamed(root, "scheduler-summary.json", { repoRoot });
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

function baselineKey({ schedulerKind, scheduleTarget, workUnitID, aggregateTarget }) {
  return [schedulerKind, scheduleTarget, workUnitID, aggregateTarget].join("|");
}

function normalizeSchedulerKind(value) {
  return String(value ?? "").trim();
}

function normalizeWorkUnitType(value) {
  return String(value ?? "").trim();
}

function observedEntryFromEvents(summary, start, finish) {
  const schedulerKind = normalizeSchedulerKind(summary.scheduler_kind);
  const scheduleTarget = String(summary.target ?? "").trim();
  const workUnitID = String(finish.work_unit_id ?? start.work_unit_id ?? "").trim();
  const aggregateTarget = String(
    start.aggregate_target ?? finish.aggregate_target ?? start.work_unit ?? finish.work_unit ?? "",
  ).trim();
  const durationMs = Math.max(1, Math.round(Number(finish.duration_ms ?? 0)));
  if (
    schedulerKind === "" ||
    scheduleTarget === "" ||
    workUnitID === "" ||
    aggregateTarget === "" ||
    durationMs <= 0
  ) {
    return null;
  }
  return {
    scheduler_kind: schedulerKind,
    schedule_target: scheduleTarget,
    work_unit_id: workUnitID,
    aggregate_target: aggregateTarget,
    duration_ms: durationMs,
  };
}

function collectObservedWorkUnitDurations(resultsDir) {
  const observed = new Map();
  let passedSchedulerCount = 0;
  for (const summaryFile of schedulerSummaryFiles(resultsDir)) {
    const summary = readJSON(repoRoot, summaryFile);
    const schedulerKind = normalizeSchedulerKind(summary.scheduler_kind);
    if (!["service-backed", "check"].includes(schedulerKind) || summary.status !== "pass") {
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
      if (!["make_target", "service_make_target", "browser_group"].includes(normalizeWorkUnitType(start?.work_unit_type))) {
        continue;
      }
      const entry = observedEntryFromEvents(summary, start, event);
      if (!entry) {
        continue;
      }
      const key = baselineKey({
        schedulerKind: entry.scheduler_kind,
        scheduleTarget: entry.schedule_target,
        workUnitID: entry.work_unit_id,
        aggregateTarget: entry.aggregate_target,
      });
      const previous = observed.get(key);
      if (!previous || entry.duration_ms > previous.duration_ms) {
        observed.set(key, entry);
      }
    }
  }
  return { observed, passedSchedulerCount };
}

function validateBaselineDocument(baseline, label) {
  if (!Number.isInteger(baseline.default_work_unit_weight_ms) || baseline.default_work_unit_weight_ms <= 0) {
    throw new Error(`${label} default_work_unit_weight_ms must be a positive integer`);
  }
  if (!baseline.work_units || typeof baseline.work_units !== "object" || Array.isArray(baseline.work_units)) {
    throw new Error(`${label} work_units must be an object`);
  }
  for (const [key, entry] of Object.entries(baseline.work_units)) {
    if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
      throw new Error(`${label} work_units.${key} must be an object`);
    }
    const expectedKey = baselineKey({
      schedulerKind: entry.scheduler_kind,
      scheduleTarget: entry.schedule_target,
      workUnitID: entry.work_unit_id,
      aggregateTarget: entry.aggregate_target,
    });
    if (key !== expectedKey) {
      throw new Error(`${label} work_units.${key} must match scheduler context key ${expectedKey}`);
    }
    if (!Number.isInteger(entry.duration_ms) || entry.duration_ms <= 0) {
      throw new Error(`${label} work_units.${key}.duration_ms must be positive integer weight ms`);
    }
  }
}

function readBaseline(file, { allowMissing = false } = {}) {
  const baselineFile = resolvePath(file);
  if (!existsSync(baselineFile)) {
    if (allowMissing) {
      return {
        schema_id: baselineSchemaID,
        note: baselineNote,
        default_work_unit_weight_ms: defaultWorkUnitWeightMs,
        work_units: {},
      };
    }
    throw new Error(`${rel(baselineFile)} is missing`);
  }
  const baseline = readJSON(repoRoot, baselineFile);
  if (baseline.schema_id !== baselineSchemaID) {
    throw new Error(`${rel(baselineFile)} must declare schema_id ${baselineSchemaID}`);
  }
  validateBaselineDocument(baseline, rel(baselineFile));
  return baseline;
}

function readTopologyProfile(file) {
  const profile = renderServiceBackedScheduleProfile(loadExecutionTopology({ manifestPath: resolvePath(file) }));
  if (!profile.defaults || typeof profile.defaults !== "object" || Array.isArray(profile.defaults)) {
    throw new Error(`${rel(resolvePath(file))} service_backed_schedules.defaults must be an object`);
  }
  return profile;
}

function readScheduledWorkUnits(topologyFile, scheduleManifestFile) {
  const scheduled = new Map();
  const add = (entry) => {
    scheduled.set(baselineKey({
      schedulerKind: entry.scheduler_kind,
      scheduleTarget: entry.schedule_target,
      workUnitID: entry.work_unit_id,
      aggregateTarget: entry.aggregate_target,
    }), entry);
  };
  const scheduleFile = resolvePath(scheduleManifestFile);
  const manifest = readJSON(repoRoot, scheduleFile);
  if (manifest.schema_id !== scheduleSchemaID) {
    throw new Error(`${rel(scheduleFile)} must declare schema_id ${scheduleSchemaID}`);
  }
  for (const schedule of manifest.schedules ?? []) {
    for (const source of schedule.work_unit_sources ?? []) {
      if (source?.type === "make_target" && typeof source.target === "string") {
        add({
          scheduler_kind: "service-backed",
          schedule_target: schedule.target,
          work_unit_id: source.target,
          aggregate_target: source.target,
        });
      } else if (source?.type === "browser_stage" && Array.isArray(source.groups)) {
        for (const group of source.groups) {
          if (typeof group.id !== "string" || typeof source.target !== "string") {
            continue;
          }
          add({
            scheduler_kind: "service-backed",
            schedule_target: schedule.target,
            work_unit_id: group.id,
            aggregate_target: source.target,
          });
        }
      }
    }
  }
  const topology = loadExecutionTopology({ manifestPath: resolvePath(topologyFile) });
  const checkScheduleFile = resolvePath(topology.generatedOutputs.check_schedule_manifest);
  const checkManifest = existsSync(checkScheduleFile)
    ? readJSON(repoRoot, checkScheduleFile)
    : { schedules: topology.checkSchedules ?? [] };
  for (const schedule of checkManifest.schedules ?? []) {
    for (const unit of schedule.work_units ?? []) {
      if (
        ["service_make_target", "browser_group"].includes(unit?.kind) &&
        typeof unit.id === "string" &&
        typeof unit.aggregate_target === "string"
      ) {
        add({
          scheduler_kind: "check",
          schedule_target: schedule.target,
          work_unit_id: unit.id,
          aggregate_target: unit.aggregate_target,
        });
      }
    }
  }
  return scheduled;
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

function baselineEntryForKey(baseline, key) {
  const entry = baseline.work_units[key];
  return entry && typeof entry === "object" ? entry : null;
}

function plannedWeight(entry, baseline, overrides) {
  if (overrides.has(entry.aggregate_target)) {
    return overrides.get(entry.aggregate_target);
  }
  const key = baselineKey({
    schedulerKind: entry.scheduler_kind,
    scheduleTarget: entry.schedule_target,
    workUnitID: entry.work_unit_id,
    aggregateTarget: entry.aggregate_target,
  });
  const baselineEntry = baselineEntryForKey(baseline, key);
  if (baselineEntry && Number.isInteger(baselineEntry.duration_ms) && baselineEntry.duration_ms > 0) {
    return baselineEntry.duration_ms;
  }
  return null;
}

function formatSubject(entry) {
  return `scheduler=${entry.scheduler_kind} schedule=${entry.schedule_target} work_unit=${entry.work_unit_id} aggregate=${entry.aggregate_target}`;
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
  const { observed, passedSchedulerCount } = collectObservedWorkUnitDurations(options.resultsDir);
  baseline.schema_id = baselineSchemaID;
  baseline.note = baselineNote;
  baseline.default_work_unit_weight_ms = positiveInteger(
    baseline.default_work_unit_weight_ms,
    defaultWorkUnitWeightMs,
  );
  baseline.work_units ??= {};
  delete baseline.default_make_target_weight_ms;
  delete baseline.targets;
  for (const [key, entry] of observed.entries()) {
    baseline.work_units[key] = entry;
  }
  baseline.updated_at = new Date().toISOString();
  baseline.work_units = sortedObjectByKey(Object.entries(baseline.work_units));
  writeFileSync(options.baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  process.stdout.write(
    `updated ${observed.size} scheduler work-unit duration baselines from ${passedSchedulerCount} successful scheduler artifact(s)\n`,
  );
}

function checkDrift(argv) {
  const options = parseCommonArgs(argv, { includeTopology: true });
  const baseline = readBaseline(options.baselineFile);
  const profile = readTopologyProfile(options.topologyFile);
  const scheduledWorkUnits = readScheduledWorkUnits(options.topologyFile, options.scheduleManifestFile);
  const scheduledTargets = new Set(
    Array.from(scheduledWorkUnits.values()).map((entry) => entry.aggregate_target),
  );
  const { overrides, errors } = validatedOverrides(profile, scheduledTargets);
  const missingBaselines = [];
  const driftErrors = [];
  const warnings = [];
  const serviceContamination = collectServiceTimingContamination(repoRoot, options.resultsDir);
  const { observed, passedSchedulerCount } = collectObservedWorkUnitDurations(options.resultsDir);

  for (const [key, entry] of observed.entries()) {
    if (!scheduledWorkUnits.has(key)) {
      continue;
    }
    const actual = entry.duration_ms;
    const planned = plannedWeight(entry, baseline, overrides);
    if (!planned) {
      missingBaselines.push(`missing scheduler work-unit baseline ${formatSubject(entry)} actual_ms=${actual}`);
      continue;
    }
    const kind = durationDriftKind(actual, planned);
    if (!kind) {
      continue;
    }
    const description = durationDriftDescription(kind, {
      subject: formatSubject(entry),
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
    process.stderr.write("Scheduler work-unit duration baseline drift ignored contaminated timing evidence:\n");
    for (const warning of warnings) {
      process.stderr.write(`- ${warning}\n`);
    }
    process.stderr.write("Service timing contamination detected:\n");
    printContaminationReasons(process.stderr, serviceContamination);
    process.stderr.write("Rerun after clean service timing evidence; do not refresh baselines from contaminated timing evidence.\n");
  }

  if (errors.length > 0 || missingBaselines.length > 0 || driftErrors.length > 0) {
    process.stderr.write("Scheduler work-unit duration baseline drift detected:\n");
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
      `Refresh from a successful scheduler run with: make service-backed-make-target-duration-baselines RESULTS_DIR=${options.resultsDir}\n`,
    );
    process.exit(1);
  }
  process.stdout.write(
    `Scheduler work-unit duration baselines match ${Array.from(observed.keys()).filter((key) => scheduledWorkUnits.has(key)).length} observed scheduled work unit(s) from ${passedSchedulerCount} successful scheduler artifact(s)\n`,
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
