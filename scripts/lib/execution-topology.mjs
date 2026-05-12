import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  assertKnownResource,
  normalizeResourceClaims,
  normalizeResourceLimits,
  resourceOverrideEnvVariablesForScheduler,
  resourceLimitsForCapacityProfile,
} from "./scheduler-resources.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");
export const executionTopologySchemaID = "cartulary.execution_topology.v3";
export const defaultExecutionTopologyManifestPath = path.join(
  repoRoot,
  "tools",
  "execution_topology_manifest.json",
);
export const taskSurfaceSchemaID = "cartulary.task_surface_manifest.v12";
export const schedulerManifestSchemaID = "cartulary.scheduler_manifest.v1";
export const checkScheduleSchemaID = "cartulary.check_schedule_sources.v1";
export const serviceBackedScheduleSchemaID = "cartulary.service_backed_schedule_sources.v1";
export const browserBatchManifestSchemaID = "cartulary.browser_e2e_batch_manifest.v5";
export const makeTargetBaselineSchemaID =
  "cartulary.scheduler_work_unit_duration_baselines.v2";
const defaultCheckWorkUnitWeightMs = 10_000;

const validDependencyCategories = new Set(["backend", "frontend", "browser"]);
const validShardModes = new Set(["none", "go_shards"]);
const validParallelismModes = new Set(["none", "package", "process"]);
const validBrowserCoverage = new Set(["authoritative", "supplemental", "raw"]);
const serviceRequirementsRequiringCheckServiceStack = new Set(["postgres", "minio", "browser_stack"]);
const checkScheduleProfileKeys = new Set(["resource_claims", "make_jobs"]);
const checkScheduleEnvNamePattern = /^[A-Z][A-Z0-9_]*$/;
const checkScheduleOwnedEnvNames = new Set([
  "CARTULARY_TEST_TARGET",
  "MAKEFLAGS",
  "MFLAGS",
  ...resourceOverrideEnvVariablesForScheduler("check"),
  ...resourceOverrideEnvVariablesForScheduler("service_backed"),
]);
const checkScheduleTargetKeys = new Set([
  "schedules",
  "profile",
  "needs",
  "expanded_needs",
  "local_input_stamp",
  "priority_band",
  "order",
  "produces_summary_targets",
  "service_backed_schedule",
  "env",
]);

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function resolveRepoPath(root, value) {
  return path.isAbsolute(value) ? value : path.join(root, value);
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
}

function requireArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value;
}

function requireNonEmptyArray(value, label) {
  const array = requireArray(value, label);
  if (array.length === 0) {
    throw new Error(`${label} must not be empty`);
  }
  return array;
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function requireBoolean(value, label) {
  if (typeof value !== "boolean") {
    throw new Error(`${label} must be a boolean`);
  }
  return value;
}

function requireStringArray(value, label, { required = true } = {}) {
  if (value === undefined && !required) {
    return [];
  }
  const result = [];
  const seen = new Set();
  for (const [index, item] of requireArray(value, label).entries()) {
    const normalized = requireString(item, `${label}[${index + 1}]`);
    if (seen.has(normalized)) {
      throw new Error(`${label} contains duplicate ${normalized}`);
    }
    seen.add(normalized);
    result.push(normalized);
  }
  return result;
}

function requireNonNegativeInteger(value, label) {
  if (!Number.isInteger(value) || value < 0) {
    throw new Error(`${label} must be a non-negative integer`);
  }
  return value;
}

function requirePositiveInteger(value, label) {
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} must be a positive integer`);
  }
  return value;
}

function requireSchema(manifest, schemaID, label) {
  if (manifest.schema_id !== schemaID) {
    throw new Error(`${label} must declare schema_id ${schemaID}`);
  }
}

function validateAllowedKeys(value, allowed, label) {
  for (const key of Object.keys(value)) {
    if (!allowed.has(key)) {
      throw new Error(`${label} has unknown key ${key}`);
    }
  }
}

function objectFromEntries(entries) {
  return Object.fromEntries([...entries].sort(([left], [right]) => left.localeCompare(right)));
}

function validateOutputPaths(root, outputs) {
  requireObject(outputs, "generated_outputs");
  for (const key of [
    "task_surface_manifest",
    "task_surface_make",
    "scheduler_manifest",
    "browser_e2e_batch_manifest",
    "execution_topology_render_index",
  ]) {
    const output = requireString(outputs[key], `generated_outputs.${key}`);
    if (output.includes("..") || path.isAbsolute(output)) {
      throw new Error(`generated_outputs.${key} must be a repo-local path`);
    }
    const resolved = resolveRepoPath(root, output);
    if (!resolved.startsWith(root + path.sep)) {
      throw new Error(`generated_outputs.${key} must stay under the repository root`);
    }
  }
}

function normalizeExecutionDependencies(topology) {
  const dependencies = [];
  const byID = new Map();
  const orders = new Set();
  for (const [index, entry] of requireNonEmptyArray(
    topology.execution_dependencies,
    "execution_dependencies",
  ).entries()) {
    const label = `execution_dependencies[${index + 1}]`;
    const dependency = {
      id: requireString(entry?.id, `${label}.id`),
      target: requireString(entry.target, `${label}.target`),
      category: requireString(entry.category, `${label}.category`),
      order: entry.order,
      service_backed: requireBoolean(entry.service_backed, `${label}.service_backed`),
      support_target: entry.support_target === true,
    };
    if (!/^[a-z][a-z0-9_]*$/.test(dependency.id)) {
      throw new Error(`${label}.id must be a snake_case identifier`);
    }
    if (!validDependencyCategories.has(dependency.category)) {
      throw new Error(`${label}.category must be backend|frontend|browser`);
    }
    if (!Number.isInteger(dependency.order) || dependency.order < 0) {
      throw new Error(`${label}.order must be a non-negative integer`);
    }
    if (byID.has(dependency.id)) {
      throw new Error(`duplicate execution dependency ${dependency.id}`);
    }
    if (orders.has(dependency.order)) {
      throw new Error(`duplicate execution dependency order ${dependency.order}`);
    }
    byID.set(dependency.id, dependency);
    orders.add(dependency.order);
    dependencies.push(dependency);
  }
  return { dependencies, byID };
}

function normalizeGoTargets(topology, dependencyByID) {
  const raw = requireObject(topology.go_targets, "go_targets");
  const targets = [];
  const byName = new Map();
  const dependencyTargets = new Map();
  const supportTargets = new Map();
  for (const [index, entry] of requireNonEmptyArray(raw.targets, "go_targets.targets").entries()) {
    const label = `go_targets.targets[${index + 1}]`;
    const descriptor = {
      name: requireString(entry?.name, `${label}.name`),
      serviceBacked: requireBoolean(entry.service_backed, `${label}.service_backed`),
      checkHeavySafe: requireBoolean(entry.check_heavy_safe, `${label}.check_heavy_safe`),
      checkServiceBackedSafe: requireBoolean(
        entry.check_service_backed_safe,
        `${label}.check_service_backed_safe`,
      ),
      checkIsolatedSafe: requireBoolean(entry.check_isolated_safe, `${label}.check_isolated_safe`),
      canonicalAuthoritative: requireBoolean(
        entry.canonical_authoritative,
        `${label}.canonical_authoritative`,
      ),
      sharding: requireString(entry.sharding, `${label}.sharding`),
      goTestParallelism: requireString(entry.go_test_parallelism, `${label}.go_test_parallelism`),
      executionDependencies: requireStringArray(
        entry.execution_dependencies ?? [],
        `${label}.execution_dependencies`,
      ),
      supportTargets: requireStringArray(entry.support_targets ?? [], `${label}.support_targets`),
    };
    if (!validShardModes.has(descriptor.sharding)) {
      throw new Error(`${label}.sharding must be none|go_shards`);
    }
    if (!validParallelismModes.has(descriptor.goTestParallelism)) {
      throw new Error(`${label}.go_test_parallelism must be none|package|process`);
    }
    if (byName.has(descriptor.name)) {
      throw new Error(`duplicate Go execution target ${descriptor.name}`);
    }
    byName.set(descriptor.name, descriptor);
    targets.push(descriptor);
    for (const dependency of descriptor.executionDependencies) {
      const metadata = dependencyByID.get(dependency);
      if (!metadata) {
        throw new Error(`${label}.execution_dependencies references unknown ${dependency}`);
      }
      if (metadata.category !== "backend") {
        throw new Error(`${label}.execution_dependencies ${dependency} is not a backend dependency`);
      }
      if (dependencyTargets.has(dependency)) {
        throw new Error(`execution dependency ${dependency} maps to more than one Go target`);
      }
      dependencyTargets.set(dependency, descriptor);
    }
    for (const supportTarget of descriptor.supportTargets) {
      const metadata = dependencyByID.get(supportTarget);
      if (!metadata) {
        throw new Error(`${label}.support_targets references unknown ${supportTarget}`);
      }
      if (!metadata.support_target) {
        throw new Error(`${label}.support_targets ${supportTarget} is not marked support_target`);
      }
      if (supportTargets.has(supportTarget)) {
        throw new Error(`support target ${supportTarget} maps to more than one Go target`);
      }
      supportTargets.set(supportTarget, descriptor);
    }
  }

  const rawAggregates = [];
  for (const [index, aggregate] of (raw.raw_go_aggregates ?? []).entries()) {
    const label = `go_targets.raw_go_aggregates[${index + 1}]`;
    const target = requireString(aggregate.target, `${label}.target`);
    if (!byName.has(target)) {
      throw new Error(`${label}.target references unknown Go target ${target}`);
    }
    rawAggregates.push({
      id: requireString(aggregate.id, `${label}.id`),
      target,
      section: requireString(aggregate.section, `${label}.section`),
      packages: requireNonEmptyArray(aggregate.packages, `${label}.packages`).map((item, itemIndex) =>
        requireString(item, `${label}.packages[${itemIndex + 1}]`),
      ),
      selectionPattern: requireString(aggregate.selection_pattern, `${label}.selection_pattern`),
      executionFamily: requireString(aggregate.execution_family, `${label}.execution_family`),
      executionLabel: requireString(aggregate.execution_label, `${label}.execution_label`),
      fixturePolicy: clone(aggregate.fixture_policy ?? {}),
      fixtureBudget: clone(aggregate.fixture_budget ?? {}),
    });
  }

  return { targets, byName, dependencyTargets, supportTargets, rawAggregates };
}

function targetNamesFromTaskSurface(topology) {
  const targets = new Set();
  for (const entry of requireNonEmptyArray(topology.task_surface?.targets, "task_surface.targets")) {
    targets.add(requireString(entry?.name, "task_surface.targets[].name"));
  }
  return targets;
}

function targetEntriesFromTaskSurface(topology) {
  const targets = new Map();
  for (const entry of requireNonEmptyArray(topology.task_surface?.targets, "task_surface.targets")) {
    targets.set(requireString(entry?.name, "task_surface.targets[].name"), entry);
  }
  return targets;
}

function validateExecutionDependencyTargets(dependencies, taskTargets) {
  for (const dependency of dependencies) {
    if (!taskTargets.has(dependency.target)) {
      throw new Error(`execution dependency ${dependency.id} target ${dependency.target} is missing from task_surface.targets`);
    }
  }
}

function validateBrowserBatch(topology, dependencyByID, taskTargets) {
  const batch = requireObject(topology.browser_e2e_batch, "browser_e2e_batch");
  const stages = requireNonEmptyArray(batch.stages, "browser_e2e_batch.stages");
  const seenStages = new Set();
  for (const [index, stage] of stages.entries()) {
    const label = `browser_e2e_batch.stages[${index + 1}]`;
    const name = requireString(stage?.name, `${label}.name`);
    const target = requireString(stage.target, `${label}.target`);
    if (seenStages.has(name)) {
      throw new Error(`duplicate browser stage ${name}`);
    }
    seenStages.add(name);
    if (!taskTargets.has(target)) {
      throw new Error(`${label}.target ${target} is missing from task_surface.targets`);
    }
    if (stage.scheduler_dependency_policy !== undefined) {
      throw new Error(`${label}.scheduler_dependency_policy is obsolete; use scheduler_needs[]`);
    }
    if (stage.scheduler_needs !== undefined) {
      const seenNeeds = new Set();
      for (const [needIndex, need] of requireArray(stage.scheduler_needs, `${label}.scheduler_needs`).entries()) {
        const targetNeed = requireString(need, `${label}.scheduler_needs[${needIndex + 1}]`);
        if (seenNeeds.has(targetNeed)) {
          throw new Error(`${label}.scheduler_needs contains duplicate ${targetNeed}`);
        }
        seenNeeds.add(targetNeed);
        if (!taskTargets.has(targetNeed)) {
          throw new Error(`${label}.scheduler_needs target ${targetNeed} is missing from task_surface.targets`);
        }
      }
    }
    for (const group of requireNonEmptyArray(stage.groups, `${label}.groups`)) {
      const groupLabel = `${label}.groups.${group?.name ?? "(missing)"}`;
      const groupTarget = requireString(group.target, `${groupLabel}.target`);
      if (!taskTargets.has(groupTarget)) {
        throw new Error(`${groupLabel}.target ${groupTarget} is missing from task_surface.targets`);
      }
      const coverage = group.coverage === undefined ? "" : requireString(group.coverage, `${groupLabel}.coverage`);
      if (coverage && !validBrowserCoverage.has(coverage)) {
        throw new Error(`${groupLabel}.coverage must be authoritative|supplemental|raw`);
      }
      const dependency = group.execution_dependency === undefined ? "" : String(group.execution_dependency).trim();
      if (dependency !== "") {
        const metadata = dependencyByID.get(dependency);
        if (!metadata) {
          throw new Error(`${groupLabel}.execution_dependency references unknown ${dependency}`);
        }
        if (metadata.category !== "browser") {
          throw new Error(`${groupLabel}.execution_dependency ${dependency} is not a browser dependency`);
        }
      }
    }
  }
}

function serviceRequirementsForTaskEntry(entry) {
  if (!Array.isArray(entry?.service_requirements)) {
    return [];
  }
  return entry.service_requirements.map((value) => String(value).trim()).filter(Boolean);
}

function requiresCheckServiceStack(entry) {
  return serviceRequirementsForTaskEntry(entry).some((requirement) =>
    serviceRequirementsRequiringCheckServiceStack.has(requirement),
  );
}

function normalizeCheckScheduleRoot(topology, taskTargets) {
  if (Array.isArray(topology.check_schedules)) {
    throw new Error(
      "check_schedules must be a profile object; flat check_schedules[] work_units are no longer supported",
    );
  }
  const root = requireObject(topology.check_schedules, "check_schedules");
  const defaults = requireObject(root.defaults, "check_schedules.defaults");
  const rawProfiles = requireObject(
    defaults.resource_profiles,
    "check_schedules.defaults.resource_profiles",
  );
  const rawPriorityBands = requireObject(
    defaults.priority_bands,
    "check_schedules.defaults.priority_bands",
  );
  const schedules = requireNonEmptyArray(root.schedules, "check_schedules.schedules");
  if (Object.keys(rawProfiles).length === 0) {
    throw new Error("check_schedules.defaults.resource_profiles must not be empty");
  }
  if (Object.keys(rawPriorityBands).length === 0) {
    throw new Error("check_schedules.defaults.priority_bands must not be empty");
  }

  const resourceProfiles = new Map();
  for (const [name, value] of Object.entries(rawProfiles)) {
    const profileName = requireString(name, "check_schedules.defaults.resource_profiles key");
    const label = `check_schedules.defaults.resource_profiles.${profileName}`;
    const profile = requireObject(value, label);
    validateAllowedKeys(profile, checkScheduleProfileKeys, label);
    resourceProfiles.set(profileName, {
      name: profileName,
      resourceClaims: clone(requireObject(profile.resource_claims, `${label}.resource_claims`)),
      makeJobs: profile.make_jobs,
    });
    if (profile.make_jobs === undefined) {
      throw new Error(`${label}.make_jobs must be declared by the reusable profile`);
    }
  }

  const priorityBands = new Map();
  for (const [name, priority] of Object.entries(rawPriorityBands)) {
    const bandName = requireString(name, "check_schedules.defaults.priority_bands key");
    priorityBands.set(
      bandName,
      requirePositiveInteger(priority, `check_schedules.defaults.priority_bands.${bandName}`),
    );
  }

  const normalizedSchedules = [];
  const seenScheduleTargets = new Set();
  for (const [index, schedule] of schedules.entries()) {
    const label = `check_schedules.schedules[${index + 1}]`;
    const target = requireString(schedule?.target, `${label}.target`);
    if (seenScheduleTargets.has(target)) {
      throw new Error(`check_schedules.schedules contains duplicate schedule ${target}`);
    }
    seenScheduleTargets.add(target);
    if (!taskTargets.has(target)) {
      throw new Error(`${label}.target ${target} is missing from task_surface.targets`);
    }
    if (schedule.work_units !== undefined) {
      throw new Error(`${label}.work_units is obsolete; add per-target check_schedule metadata`);
    }
    if (schedule.resource_limits !== undefined) {
      throw new Error(`${label}.resource_limits is obsolete; use capacity_profile`);
    }
    const capacityProfile = requireString(schedule.capacity_profile, `${label}.capacity_profile`);
    const profileLimits = resourceLimitsForCapacityProfile(capacityProfile, label, {
      scheduler: "check",
      allowAuto: true,
    });
    normalizedSchedules.push({
      target,
      capacityProfile,
      resourceLimits: Object.fromEntries(profileLimits.limits.entries()),
      summaryGroups: clone(schedule.summary_groups ?? []),
    });
  }

  return {
    resourceProfiles,
    priorityBands,
    schedules: normalizedSchedules,
    scheduleTargets: seenScheduleTargets,
  };
}

function normalizeCheckScheduleMetadata(entry, label, scheduleTargets) {
  if (entry.check_schedule === undefined) {
    return null;
  }
  const ownerTarget = requireString(entry.name, `${label}.name`);
  const raw = requireObject(entry.check_schedule, `${label}.check_schedule`);
  validateAllowedKeys(raw, checkScheduleTargetKeys, `${label}.check_schedule`);
  const schedules = requireStringArray(raw.schedules, `${label}.check_schedule.schedules`);
  if (schedules.length === 0) {
    throw new Error(`${label}.check_schedule.schedules must not be empty`);
  }
  for (const schedule of schedules) {
    if (!scheduleTargets.has(schedule)) {
      throw new Error(`${label}.check_schedule.schedules references unknown check schedule ${schedule}`);
    }
    if (
      !Array.isArray(entry.default_inclusion_sets) ||
      !entry.default_inclusion_sets.includes(schedule)
    ) {
      throw new Error(
        `${label}.check_schedule includes ${schedule} but target is not default_inclusion_sets ${schedule}`,
      );
    }
  }
  const producesSummaryTargets = requireStringArray(
    raw.produces_summary_targets ?? [],
    `${label}.check_schedule.produces_summary_targets`,
  );
  if (producesSummaryTargets.length > 0 && !producesSummaryTargets.includes(ownerTarget)) {
    throw new Error(
      `${label}.check_schedule.produces_summary_targets must include owning target ${ownerTarget}`,
    );
  }
  return {
    schedules,
    profile: requireString(raw.profile, `${label}.check_schedule.profile`),
    needs: requireStringArray(raw.needs ?? [], `${label}.check_schedule.needs`),
    expandedNeeds: requireStringArray(raw.expanded_needs ?? [], `${label}.check_schedule.expanded_needs`),
    localInputStamp: normalizeLocalInputStamp(raw.local_input_stamp, `${label}.check_schedule.local_input_stamp`),
    priorityBand: requireString(raw.priority_band, `${label}.check_schedule.priority_band`),
    order: requireNonNegativeInteger(raw.order, `${label}.check_schedule.order`),
    producesSummaryTargets,
    serviceBackedSchedule: raw.service_backed_schedule === undefined
      ? null
      : requireString(raw.service_backed_schedule, `${label}.check_schedule.service_backed_schedule`),
    env: normalizeCheckScheduleEnv(raw.env, `${label}.check_schedule.env`),
  };
}

function normalizeLocalInputStamp(value, label) {
  if (value === undefined) {
    return null;
  }
  const raw = requireObject(value, label);
  validateAllowedKeys(raw, new Set(["profile"]), label);
  return {
    profile: requireString(raw.profile, `${label}.profile`),
  };
}

function normalizeCheckScheduleEnv(value, label) {
  if (value === undefined) {
    return {};
  }
  const env = requireObject(value, label);
  const entries = [];
  for (const [name, rawValue] of Object.entries(env)) {
    const envName = requireString(name, `${label} key`);
    if (!checkScheduleEnvNamePattern.test(envName)) {
      throw new Error(`${label}.${envName} must be a safe environment variable name`);
    }
    if (checkScheduleOwnedEnvNames.has(envName)) {
      throw new Error(`${label}.${envName} is scheduler-owned and cannot be overridden`);
    }
    if (typeof rawValue !== "string") {
      throw new Error(`${label}.${envName} must be a string`);
    }
    if (rawValue.includes("\0") || rawValue.includes("\n") || rawValue.includes("\r")) {
      throw new Error(`${label}.${envName} must be a single-line string`);
    }
    entries.push([envName, rawValue]);
  }
  return objectFromEntries(entries);
}

function normalizeCheckMakeJobs(value, label, claims) {
  if (typeof value === "string") {
    const resource = assertKnownResource(value, `${label}.make_jobs`, { scheduler: "check" });
    if (!claims.has(resource)) {
      throw new Error(`${label}.make_jobs resource ${resource} must be claimed by the profile`);
    }
    return resource;
  }
  return requirePositiveInteger(value, `${label}.make_jobs`);
}

function claimsCheckServiceBoundaryResource(claims) {
  return claims.has("suite_service_stack") || claims.has("migration_scratch_postgres");
}

function renderCheckSchedulesFromTopology(topology, taskTargets, taskTargetEntries) {
  const root = normalizeCheckScheduleRoot(topology, taskTargets);
  const targetMetadata = [];
  for (const [target, entry] of taskTargetEntries.entries()) {
    const metadata = normalizeCheckScheduleMetadata(entry, `task_surface.targets.${target}`, root.scheduleTargets);
    if (metadata) {
      targetMetadata.push({ target, entry, metadata });
    }
  }

  return root.schedules.map((schedule) => {
    const label = `check_schedules.schedules.${schedule.target}`;
    const resourceLimits = normalizeResourceLimits(schedule.resourceLimits, label, {
      scheduler: "check",
      capacityProfile: schedule.capacityProfile,
      allowAuto: true,
    }).limits;
    const usedTargets = new Set();
    const usedOrders = new Map();
    const units = [];
    for (const { target, entry, metadata } of targetMetadata) {
      if (!metadata.schedules.includes(schedule.target)) {
        continue;
      }
      if (usedTargets.has(target)) {
        throw new Error(`${label} contains duplicate generated work unit ${target}`);
      }
      usedTargets.add(target);
      const profile = root.resourceProfiles.get(metadata.profile);
      if (!profile) {
        throw new Error(`${target}.check_schedule.profile references unknown profile ${metadata.profile}`);
      }
      const priorityBase = root.priorityBands.get(metadata.priorityBand);
      if (priorityBase === undefined) {
        throw new Error(
          `${target}.check_schedule.priority_band references unknown priority band ${metadata.priorityBand}`,
        );
      }
      const orderKey = `${metadata.priorityBand}:${metadata.order}`;
      if (usedOrders.has(orderKey)) {
        throw new Error(
          `${label} has duplicate priority order ${orderKey} for ${usedOrders.get(orderKey)} and ${target}`,
        );
      }
      usedOrders.set(orderKey, target);
      const priority = priorityBase - metadata.order;
      if (priority < 1) {
        throw new Error(`${target}.check_schedule order ${metadata.order} exhausts ${metadata.priorityBand} priority`);
      }
      const claims = normalizeResourceClaims(profile.resourceClaims, `${target}.check_schedule profile ${profile.name}`, resourceLimits, {
        scheduler: "check",
        allowBounded: true,
      });
      if (
        requiresCheckServiceStack(entry) &&
        !claimsCheckServiceBoundaryResource(claims) &&
        !metadata.serviceBackedSchedule
      ) {
        throw new Error(
          `${target}.check_schedule target declares service_requirements and must claim a check service boundary resource or use a service-backed schedule`,
        );
      }
      const unit = {
        target,
        priority,
        weight_ms: defaultCheckWorkUnitWeightMs,
        needs: clone(metadata.needs),
        ...(metadata.expandedNeeds.length > 0 ? { expanded_needs: clone(metadata.expandedNeeds) } : {}),
        ...(metadata.localInputStamp ? { local_input_stamp: clone(metadata.localInputStamp) } : {}),
        ...(metadata.producesSummaryTargets.length > 0
          ? { produces_summary_targets: clone(metadata.producesSummaryTargets) }
          : {}),
        resource_claims: clone(profile.resourceClaims),
        make_jobs: normalizeCheckMakeJobs(profile.makeJobs, `${target}.check_schedule profile ${profile.name}`, claims),
        command: { type: "make_target", target },
        ...(Object.keys(metadata.env).length > 0 ? { env: clone(metadata.env) } : {}),
        ...(metadata.serviceBackedSchedule ? { service_backed_schedule: metadata.serviceBackedSchedule } : {}),
      };
      units.push({ ...unit, order: metadata.order });
    }
    if (units.length === 0) {
      throw new Error(`${label} must produce at least one work unit`);
    }
    assertAcyclicUnits(schedule.target, units);
    units.sort(
      (left, right) =>
        right.priority - left.priority ||
        right.weight_ms - left.weight_ms ||
        left.order - right.order ||
        left.target.localeCompare(right.target),
    );
    return {
      target: schedule.target,
      scheduler_kind: "check",
      capacity_profile: schedule.capacityProfile,
      resource_limits: clone(schedule.resourceLimits),
      summary_groups: clone(schedule.summaryGroups),
      work_units: units.map(({ order: _order, ...unit }) => unit),
    };
  });
}

function assertAcyclicUnits(scheduleTarget, units) {
  const byTarget = new Map(units.map((unit) => [unit.target, unit]));
  const visiting = new Set();
  const visited = new Set();
  const visit = (target) => {
    if (visited.has(target)) {
      return;
    }
    if (visiting.has(target)) {
      throw new Error(`check schedule ${scheduleTarget} has a dependency cycle at ${target}`);
    }
    const unit = byTarget.get(target);
    if (!unit) {
      throw new Error(`check schedule ${scheduleTarget} references unknown dependency ${target}`);
    }
    visiting.add(target);
    for (const need of unit.needs ?? []) {
      if (!byTarget.has(need)) {
        throw new Error(`check schedule ${scheduleTarget} work unit ${target} depends on unknown ${need}`);
      }
      visit(need);
    }
    visiting.delete(target);
    visited.add(target);
  };
  for (const unit of units) {
    visit(unit.target);
  }
}

function validateServiceBackedSchedules(manifestPath, topology, taskTargets) {
  const service = requireObject(topology.service_backed_schedules, "service_backed_schedules");
  const defaults = requireObject(service.defaults, "service_backed_schedules.defaults");
  const baselinePath = requireString(
    defaults.make_target_duration_baseline,
    "service_backed_schedules.defaults.make_target_duration_baseline",
  );
  const resolvedBaseline = path.isAbsolute(baselinePath)
    ? baselinePath
    : path.join(path.dirname(manifestPath), baselinePath);
  if (!existsSync(resolvedBaseline)) {
    throw new Error(`scheduler work-unit duration baseline is missing: ${baselinePath}`);
  }
  requireSchema(readJSON(resolvedBaseline), makeTargetBaselineSchemaID, baselinePath);

  for (const [scheduleIndex, schedule] of requireNonEmptyArray(
    service.schedules,
    "service_backed_schedules.schedules",
  ).entries()) {
    const label = `service_backed_schedules.schedules[${scheduleIndex + 1}]`;
    const target = requireString(schedule?.target, `${label}.target`);
    if (!taskTargets.has(target)) {
      throw new Error(`${label}.target ${target} is missing from task_surface.targets`);
    }
    if (schedule.resource_limits !== undefined) {
      throw new Error(`${label}.resource_limits is obsolete; use capacity_profile`);
    }
    normalizeResourceLimits(undefined, label, {
      scheduler: "service_backed",
      capacityProfile: requireString(schedule.capacity_profile, `${label}.capacity_profile`),
      allowAuto: true,
    });
  }
}

function normalizeTopology(raw, root, manifestPath) {
  requireSchema(raw, executionTopologySchemaID, manifestPath);
  validateOutputPaths(root, raw.generated_outputs);
  requireObject(raw.task_surface, "task_surface");
  const taskTargets = targetNamesFromTaskSurface(raw);
  const taskTargetEntries = targetEntriesFromTaskSurface(raw);
  const { dependencies, byID: dependencyByID } = normalizeExecutionDependencies(raw);
  validateExecutionDependencyTargets(dependencies, taskTargets);
  const goTargets = normalizeGoTargets(raw, dependencyByID);
  validateBrowserBatch(raw, dependencyByID, taskTargets);
  const checkSchedules = renderCheckSchedulesFromTopology(raw, taskTargets, taskTargetEntries);
  validateServiceBackedSchedules(manifestPath, raw, taskTargets);
  return {
    root,
    manifestPath,
    raw: clone(raw),
    generatedOutputs: clone(raw.generated_outputs),
    executionDependencies: dependencies,
    executionDependencyByID: dependencyByID,
    goTargets,
    taskSurface: clone(raw.task_surface),
    checkScheduleProfile: clone(raw.check_schedules),
    checkSchedules,
    serviceBackedSchedules: clone(raw.service_backed_schedules),
    browserBatch: clone(raw.browser_e2e_batch),
  };
}

export function loadExecutionTopology(options = {}) {
  const root = path.resolve(options.root ?? repoRoot);
  const configured =
    options.manifestPath ??
    process.env.CARTULARY_EXECUTION_TOPOLOGY_MANIFEST ??
    defaultExecutionTopologyManifestPath;
  const manifestPath = resolveRepoPath(root, configured);
  return normalizeTopology(readJSON(manifestPath), root, manifestPath);
}

export function renderTaskSurfaceManifest(topology) {
  const taskSurface = clone(topology.taskSurface);
  taskSurface.targets = taskSurface.targets.map(({ check_schedule: _checkSchedule, ...target }) => target);
  return {
    schema_id: taskSurfaceSchemaID,
    ...taskSurface,
  };
}

export function renderBrowserBatchManifest(topology) {
  return {
    schema_id: browserBatchManifestSchemaID,
    ...clone(topology.browserBatch),
  };
}

function serviceBackedScheduleByTarget(manifest) {
  return new Map((manifest?.schedules ?? []).map((schedule) => [schedule.target, schedule]));
}

function serviceBackedCheckResourceLimits(serviceSchedule) {
  const limits = {};
  for (const [resource, limit] of Object.entries(serviceSchedule.resource_limits ?? {})) {
    if (resource === "go_cpu" || resource === "go_io") {
      continue;
    }
    limits[resource] = limit;
  }
  return limits;
}

function expandCheckScheduleServiceBackedUnits(schedule, serviceBackedManifest, expandServiceBackedScheduleForCheck) {
  const schedulesByTarget = serviceBackedScheduleByTarget(serviceBackedManifest);
  const workUnits = [];
  let resourceLimits = clone(schedule.resource_limits);
  for (const unit of schedule.work_units) {
    if (!unit.service_backed_schedule) {
      const { expanded_needs: expandedNeeds = [], ...rest } = unit;
      workUnits.push({
        ...rest,
        needs: [...(unit.needs ?? []), ...expandedNeeds],
      });
      continue;
    }
    const serviceSchedule = schedulesByTarget.get(unit.service_backed_schedule);
    if (!serviceSchedule) {
      throw new Error(
        `check schedule ${schedule.target} references missing service-backed schedule ${unit.service_backed_schedule}`,
      );
    }
    resourceLimits = {
      ...resourceLimits,
      ...serviceBackedCheckResourceLimits(serviceSchedule),
    };
    workUnits.push(
      ...expandServiceBackedScheduleForCheck({
        repoRoot,
        serviceSchedule,
        parentUnit: unit,
      }),
    );
  }
  return {
    ...schedule,
    resource_limits: resourceLimits,
    work_units: workUnits,
  };
}

export function renderCheckScheduleManifest(topology, options = {}) {
  const serviceBackedManifest = options.serviceBackedScheduleManifest ?? null;
  const expandServiceBackedScheduleForCheck = options.expandServiceBackedScheduleForCheck ?? null;
  if (serviceBackedManifest && typeof expandServiceBackedScheduleForCheck !== "function") {
    throw new Error("renderCheckScheduleManifest requires expandServiceBackedScheduleForCheck with serviceBackedScheduleManifest");
  }
  const schedules = serviceBackedManifest
    ? topology.checkSchedules.map((schedule) =>
        expandCheckScheduleServiceBackedUnits(
          schedule,
          serviceBackedManifest,
          expandServiceBackedScheduleForCheck,
        ),
      )
    : topology.checkSchedules;
  return {
    schema_id: checkScheduleSchemaID,
    schedules: clone(schedules),
  };
}

export function renderServiceBackedScheduleProfile(topology) {
  return clone(topology.serviceBackedSchedules);
}

export function executionDependencyMetadata(root = repoRoot) {
  const topology = loadExecutionTopology({ root });
  return new Map(topology.executionDependencies.map((entry) => [entry.id, entry]));
}

export function serviceBackedGoExecutionDependencies(root = repoRoot) {
  const topology = loadExecutionTopology({ root });
  return new Set(
    topology.executionDependencies
      .filter((entry) => entry.category === "backend" && entry.service_backed && !entry.support_target)
      .map((entry) => entry.id),
  );
}

export function serviceBackedSupportTargets(root = repoRoot) {
  const topology = loadExecutionTopology({ root });
  return new Set(
    topology.executionDependencies
      .filter((entry) => entry.support_target && entry.service_backed)
      .map((entry) => entry.id),
  );
}

export function validExecutionDependencyIDs(root = repoRoot) {
  return new Set(executionDependencyMetadata(root).keys());
}

export function validSupportTargetIDs(root = repoRoot) {
  const topology = loadExecutionTopology({ root });
  return new Set(
    topology.executionDependencies.filter((entry) => entry.support_target).map((entry) => entry.id),
  );
}

export function compareExecutionDependencyIDs(left, right, root = repoRoot) {
  const metadata = executionDependencyMetadata(root);
  const leftInfo = metadata.get(left);
  const rightInfo = metadata.get(right);
  return (
    (leftInfo?.order ?? Number.MAX_SAFE_INTEGER) -
      (rightInfo?.order ?? Number.MAX_SAFE_INTEGER) ||
    String(left).localeCompare(String(right))
  );
}

export function targetForExecutionDependencyID(id, label = "execution_dependency", root = repoRoot) {
  if (id === "") {
    return "";
  }
  const info = executionDependencyMetadata(root).get(id);
  if (!info) {
    throw new Error(`${label} has no execution dependency metadata for ${id}`);
  }
  if (!info.target) {
    throw new Error(`${label} ${id} has no Make target mapping`);
  }
  return info.target;
}

export function topologySummary(topology) {
  return {
    schema_id: topology.raw.schema_id,
    execution_dependencies: topology.executionDependencies.length,
    go_targets: topology.goTargets.targets.length,
    raw_go_aggregates: topology.goTargets.rawAggregates.length,
    task_targets: topology.taskSurface.targets.length,
    check_schedules: topology.checkSchedules.length,
    service_backed_schedules: topology.serviceBackedSchedules.schedules.length,
    browser_stages: topology.browserBatch.stages.length,
    generated_outputs: objectFromEntries(Object.entries(topology.generatedOutputs)),
  };
}
