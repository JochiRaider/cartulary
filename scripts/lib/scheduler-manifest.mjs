import { readFile } from "node:fs/promises";
import path from "node:path";

import {
  normalizeResourceClaims as normalizeSchedulerResourceClaims,
  normalizeResourceLimits as normalizeSchedulerResourceLimits,
  provisionalResourceLimitsForClaims,
  resolveAutoResourceLimits,
} from "./scheduler-resources.mjs";

export const schedulerManifestSchemaID = "cartulary.scheduler_manifest.v1";
const envNamePattern = /^[A-Z][A-Z0-9_]*$/;
const commandTypes = new Set([
  "make_target",
  "service_session_start",
  "browser_stage_session_start",
  "browser_group",
  "browser_stage_complete",
  "go_shard",
  "go_shard_finalize",
  "service_complete",
]);

export function parseResourceLimitOverride(value) {
  const [resource, amountText, extra] = value.split("=");
  if (!resource || !amountText || extra !== undefined) {
    throw new Error(`--resource-limit expects <name=value>, got ${value}`);
  }
  const amount = Number.parseInt(amountText, 10);
  if (!Number.isInteger(amount) || amount < 1) {
    throw new Error(`--resource-limit ${resource} must be a positive integer`);
  }
  return [resource.trim(), amount];
}

export async function loadSchedulerManifest(file, { repoRoot, schemaID = schedulerManifestSchemaID } = {}) {
  const manifestPath = path.isAbsolute(file) ? file : path.join(repoRoot, file);
  const manifest = JSON.parse(await readFile(manifestPath, "utf8"));
  if (manifest.schema_id !== schemaID) {
    throw new Error(`${manifestPath} must declare schema_id ${schemaID}; run make phase-schedules to regenerate the normalized scheduler manifest`);
  }
  if (!Array.isArray(manifest.schedules)) {
    throw new Error(`${manifestPath} must declare schedules[]`);
  }
  return { manifest, manifestPath };
}

export const loadScheduleManifest = loadSchedulerManifest;

export function selectSingleSchedule(manifest, target, { label = "schedule" } = {}) {
  const matches = manifest.schedules.filter((schedule) => schedule?.target === target);
  if (matches.length !== 1) {
    throw new Error(`expected exactly one ${label} for ${target}, found ${matches.length}`);
  }
  return matches[0];
}

function normalizeResourceClaims(value, label, resourceLimits, scheduler) {
  return normalizeSchedulerResourceClaims(value ?? {}, label, resourceLimits, {
    scheduler,
    allowBounded: scheduler === "check",
  });
}

function normalizePriority(value, label) {
  if (value === undefined) {
    return 0;
  }
  if (!Number.isInteger(value) || value < 0) {
    throw new Error(`${label} priority must be a non-negative integer`);
  }
  return value;
}

function normalizePositiveWeight(value, label) {
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} weight_ms must be a positive integer`);
  }
  return value;
}

function normalizeBoolean(value, label, fallback) {
  if (value === undefined) {
    return fallback;
  }
  if (typeof value !== "boolean") {
    throw new Error(`${label} must be a boolean`);
  }
  return value;
}

function normalizeProgressTickSeconds(value, label) {
  if (value === undefined) {
    return 30;
  }
  if (!Number.isInteger(value) || value < 5 || value > 300) {
    throw new Error(`${label} must be an integer between 5 and 300`);
  }
  return value;
}

function normalizeEnv(value, label) {
  if (value === undefined) {
    return {};
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} env must be an object`);
  }
  const entries = [];
  for (const [name, rawValue] of Object.entries(value)) {
    if (!envNamePattern.test(name)) {
      throw new Error(`${label} env.${name} must be a safe environment variable name`);
    }
    if (typeof rawValue !== "string" || rawValue.includes("\0") || rawValue.includes("\n") || rawValue.includes("\r")) {
      throw new Error(`${label} env.${name} must be a single-line string`);
    }
    entries.push([name, rawValue]);
  }
  return Object.fromEntries(entries.sort(([left], [right]) => left.localeCompare(right)));
}

function normalizeCommand(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} command must be an object`);
  }
  if (typeof value.type !== "string" || !commandTypes.has(value.type)) {
    throw new Error(`${label} command.type must be one of ${Array.from(commandTypes).join(", ")}`);
  }
  return JSON.parse(JSON.stringify(value));
}

function normalizeRetainedResourceClaims(value, label, resourceClaims, resourceLimits, scheduler) {
  if (value === undefined) {
    return new Map();
  }
  const retained = normalizeResourceClaims(value, label, resourceLimits, scheduler);
  for (const [resource, amount] of retained.entries()) {
    if ((resourceClaims.get(resource) ?? 0) < amount) {
      throw new Error(`${label}.${resource} exceeds resource_claims`);
    }
  }
  return retained;
}

function normalizeMakeJobs(value, label, resourceClaims, scheduler) {
  if (value === undefined) {
    return 1;
  }
  if (typeof value === "string") {
    if (scheduler !== "check") {
      throw new Error(`${label} make_jobs resource references are only supported by check schedules`);
    }
    return resourceClaims.get(value) ?? 1;
  }
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} make_jobs must be a positive integer or claimed resource name`);
  }
  return value;
}

function normalizeWorkUnit(unit, index, scheduleLabel, scheduler, resourceLimits) {
  const label = `${scheduleLabel} work_units ${index + 1}`;
  if (!unit || typeof unit !== "object" || Array.isArray(unit)) {
    throw new Error(`${label} must be an object`);
  }
  if (typeof unit.target !== "string" || unit.target.trim() === "") {
    throw new Error(`${label} must declare target`);
  }
  const target = unit.target.trim();
  const id = typeof unit.id === "string" && unit.id.trim() !== "" ? unit.id.trim() : target;
  const kind = typeof unit.kind === "string" && unit.kind.trim() !== "" ? unit.kind.trim() : "make_target";
  const aggregateTarget =
    typeof unit.aggregate_target === "string" && unit.aggregate_target.trim() !== ""
      ? unit.aggregate_target.trim()
      : target;
  const claims = normalizeResourceClaims(unit.resource_claims, `${label} ${target}`, resourceLimits, scheduler);
  const retainedResourceClaims = normalizeRetainedResourceClaims(
    unit.retained_resource_claims,
    `${label} ${target} retained_resource_claims`,
    claims,
    resourceLimits,
    scheduler,
  );
  return {
    id,
    label: typeof unit.label === "string" && unit.label.trim() !== "" ? unit.label.trim() : target,
    kind,
    type: typeof unit.type === "string" && unit.type.trim() !== "" ? unit.type.trim() : kind,
    class: typeof unit.class === "string" ? unit.class.trim() : "",
    target,
    aggregateTarget,
    completionKeys: normalizeStringList(unit.completion_keys, `${label} ${target} completion_keys`).length > 0
      ? normalizeStringList(unit.completion_keys, `${label} ${target} completion_keys`)
      : [id],
    failureKeys: normalizeStringList(unit.failure_keys, `${label} ${target} failure_keys`).length > 0
      ? normalizeStringList(unit.failure_keys, `${label} ${target} failure_keys`)
      : undefined,
    runningDependencyKeys: normalizeStringList(
      unit.running_dependency_keys,
      `${label} ${target} running_dependency_keys`,
    ),
    priority: normalizePriority(unit.priority, `${label} ${target}`),
    weightMs: normalizePositiveWeight(unit.weight_ms, `${label} ${target}`),
    needs: normalizeNeeds(unit.needs, `${label} ${target}`),
    producesSummaryTargets: normalizeStringList(
      unit.produces_summary_targets,
      `${label} ${target} produces_summary_targets`,
    ),
    resourceClaims: claims,
    retainedResourceClaims,
    releaseRetainedResourceClaims: normalizeResourceClaims(
      unit.release_retained_resource_claims ?? {},
      `${label} ${target} release_retained_resource_claims`,
      resourceLimits,
      scheduler,
    ),
    makeJobs: normalizeMakeJobs(unit.make_jobs, `${label} ${target}`, claims, scheduler),
    env: normalizeEnv(unit.env, `${label} ${target}`),
    commandSpec: normalizeCommand(unit.command, `${label} ${target}`),
    serviceSession: unit.service_session && typeof unit.service_session === "object" && !Array.isArray(unit.service_session)
      ? JSON.parse(JSON.stringify(unit.service_session))
      : null,
    browserStage: typeof unit.browser_stage === "string" ? unit.browser_stage : "",
    browserGroup: unit.browser_group && typeof unit.browser_group === "object" && !Array.isArray(unit.browser_group)
      ? JSON.parse(JSON.stringify(unit.browser_group))
      : null,
    shard: typeof unit.shard === "string" ? unit.shard : "",
    shardNames: Array.isArray(unit.shard_names)
      ? unit.shard_names.filter((entry) => typeof entry === "string" && entry !== "")
      : [],
    schedulerProfile: typeof unit.scheduler_profile === "string" ? unit.scheduler_profile : "",
    countInTotal: unit.count_in_total === false ? false : undefined,
    countsStarted: unit.counts_started === false ? false : undefined,
    completeOnFailure: unit.complete_on_failure === true,
    unblockLabel: typeof unit.unblock_label === "string" && unit.unblock_label.trim() !== ""
      ? unit.unblock_label.trim()
      : undefined,
    startDetail: {},
    order: index,
  };
}

function validateWorkUnitDependencyGraph(units, scheduleLabel) {
  const ids = new Set();
  const completionKeys = new Map();
  for (const unit of units) {
    if (ids.has(unit.id)) {
      throw new Error(`${scheduleLabel} contains duplicate work unit id ${unit.id}`);
    }
    ids.add(unit.id);
    for (const key of unit.completionKeys) {
      if (completionKeys.has(key)) {
        throw new Error(`${scheduleLabel} completion key ${key} is produced by both ${completionKeys.get(key)} and ${unit.id}`);
      }
      completionKeys.set(key, unit.id);
    }
  }
  for (const unit of units) {
    unit.failureKeys = unit.failureKeys?.length ? unit.failureKeys : unit.completionKeys;
    for (const need of unit.needs) {
      if (!completionKeys.has(need)) {
        throw new Error(`${scheduleLabel} work unit ${unit.id} depends on unknown completion key ${need}`);
      }
    }
  }
  const byID = new Map(units.map((unit) => [unit.id, unit]));
  const visiting = new Set();
  const visited = new Set();
  const visit = (unit) => {
    if (visited.has(unit.id)) {
      return;
    }
    if (visiting.has(unit.id)) {
      throw new Error(`${scheduleLabel} has a dependency cycle at ${unit.id}`);
    }
    visiting.add(unit.id);
    for (const need of unit.needs) {
      const producer = byID.get(completionKeys.get(need));
      if (producer) {
        visit(producer);
      }
    }
    visiting.delete(unit.id);
    visited.add(unit.id);
  };
  for (const unit of units) {
    visit(unit);
  }
}

function validateRetainedClaimReleases(units, scheduleLabel) {
  const retainedTotals = new Map();
  const releaseTotals = new Map();
  for (const unit of units) {
    for (const [resource, amount] of unit.retainedResourceClaims.entries()) {
      retainedTotals.set(resource, (retainedTotals.get(resource) ?? 0) + amount);
    }
    for (const [resource, amount] of unit.releaseRetainedResourceClaims.entries()) {
      releaseTotals.set(resource, (releaseTotals.get(resource) ?? 0) + amount);
    }
  }
  for (const [resource, amount] of releaseTotals.entries()) {
    if ((retainedTotals.get(resource) ?? 0) < amount) {
      throw new Error(`${scheduleLabel} releases retained ${resource} claims that were never retained`);
    }
  }
}

export function normalizeSchedulerSchedule(manifest, target, {
  scheduler = null,
  resourceLimitOverrides = new Map(),
  autoLimitResolvers = () => ({}),
  env = process.env,
  label = "scheduler schedule",
} = {}) {
  const schedule = selectSingleSchedule(manifest, target, { label });
  const scheduleLabel = `${label} ${target}`;
  if (scheduler && schedule.scheduler_kind !== scheduler) {
    throw new Error(`${scheduleLabel} scheduler_kind must be ${scheduler}`);
  }
  const schedulerKind = schedule.scheduler_kind;
  if (!["check", "service_backed", "phase_slice"].includes(schedulerKind)) {
    throw new Error(`${scheduleLabel} scheduler_kind is unsupported: ${schedulerKind}`);
  }
  if (!Array.isArray(schedule.work_units) || schedule.work_units.length === 0) {
    throw new Error(`${scheduleLabel} must declare work_units[]`);
  }
  if (Array.isArray(schedule.work_unit_sources)) {
    throw new Error(`${scheduleLabel} must use normalized work_units[], not work_unit_sources[]`);
  }
  const normalizedLimits = normalizeSchedulerResourceLimits(schedule.resource_limits, scheduleLabel, {
    scheduler: schedulerKind,
    capacityProfile: schedule.capacity_profile ?? null,
    overrides: resourceLimitOverrides,
    allowAuto: true,
    env,
  });
  const provisionalLimits = provisionalResourceLimitsForClaims(normalizedLimits.limits);
  const provisionalUnits = schedule.work_units.map((unit, index) =>
    normalizeWorkUnit(unit, index, scheduleLabel, schedulerKind, provisionalLimits),
  );
  const resolvedLimits = resolveAutoResourceLimits(
    normalizedLimits.limits,
    normalizedLimits.sources,
    scheduleLabel,
    autoLimitResolvers(provisionalUnits),
  );
  const units = schedule.work_units.map((unit, index) =>
    normalizeWorkUnit(unit, index, scheduleLabel, schedulerKind, resolvedLimits.resourceLimits),
  );
  validateWorkUnitDependencyGraph(units, scheduleLabel);
  validateRetainedClaimReleases(units, scheduleLabel);
  const sortedUnits = units.sort(
    (left, right) =>
      right.priority - left.priority ||
      right.weightMs - left.weightMs ||
      left.order - right.order ||
      left.id.localeCompare(right.id),
  );
  return {
    target,
    schedulerKind,
    capacityProfile: schedule.capacity_profile ?? "",
    stopOnFirstFailure: normalizeBoolean(
      schedule.stop_on_first_failure,
      `${scheduleLabel}.stop_on_first_failure`,
      schedulerKind === "check",
    ),
    progressTickSeconds: normalizeProgressTickSeconds(
      schedule.progress_tick_seconds,
      `${scheduleLabel}.progress_tick_seconds`,
    ),
    validateTiming: normalizeBoolean(schedule.validate_timing, `${scheduleLabel}.validate_timing`, true),
    summaryGroups: schedule.summary_groups ?? [],
    finalizers: Array.isArray(schedule.finalizers) ? schedule.finalizers : [],
    resourceLimits: resolvedLimits.resourceLimits,
    resourceLimitSources: resolvedLimits.resourceLimitSources,
    workUnits: sortedUnits,
  };
}

export function normalizeStringList(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label} entries must be non-empty strings`);
    }
    return entry.trim();
  });
}

export function normalizeNeeds(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} needs must be an array`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label} needs entries must be non-empty strings`);
    }
    return entry.trim();
  });
}

export function validateTargetDependencyGraph(
  nodes,
  { scheduleLabel, nodeKind, duplicateTargetsMessage = null },
) {
  const targets = new Set();
  const duplicateTargets = [];
  for (const node of nodes) {
    if (targets.has(node.target)) {
      duplicateTargets.push(node.target);
    }
    targets.add(node.target);
  }
  if (duplicateTargets.length > 0) {
    if (duplicateTargetsMessage) {
      throw new Error(duplicateTargetsMessage(duplicateTargets));
    }
    throw new Error(`${scheduleLabel} contains duplicate ${nodeKind} target ${duplicateTargets[0]}`);
  }

  for (const node of nodes) {
    for (const need of node.needs) {
      if (!targets.has(need)) {
        throw new Error(`${scheduleLabel} ${nodeKind} ${node.target} depends on unknown target ${need}`);
      }
      if (need === node.target) {
        throw new Error(`${scheduleLabel} ${nodeKind} ${node.target} cannot depend on itself`);
      }
    }
  }

  const byTarget = new Map(nodes.map((node) => [node.target, node]));
  const visiting = new Set();
  const visited = new Set();
  const visit = (node) => {
    if (visited.has(node.target)) {
      return;
    }
    if (visiting.has(node.target)) {
      throw new Error(`${scheduleLabel} has a dependency cycle at ${node.target}`);
    }
    visiting.add(node.target);
    for (const need of node.needs) {
      visit(byTarget.get(need));
    }
    visiting.delete(node.target);
    visited.add(node.target);
  };
  for (const node of nodes) {
    visit(node);
  }
}
