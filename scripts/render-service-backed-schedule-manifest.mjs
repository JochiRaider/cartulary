#!/usr/bin/env node
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { normalizeBrowserBatchStages } from "./lib/browser-batch-manifest.mjs";
import { createPlan as createBrowserShardPlan } from "./lib/browser-shard-plan.mjs";
import {
  defaultExecutionTopologyManifestPath,
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderServiceBackedScheduleProfile,
} from "./lib/execution-topology.mjs";
import {
  compareExecutionDependencies,
  executionDependencyInfo,
} from "./lib/execution-dependencies.mjs";
import {
  collectEntries,
  loadManifest,
  phaseManifestNames,
} from "./lib/phase-manifest.mjs";
import {
  browserStageResource,
  resourceLimitsForCapacityProfile,
} from "./lib/scheduler-resources.mjs";
import { collectTargetPlanRows, findTargetDescriptor } from "./lib/target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const scheduleSchemaID = "cartulary.service_backed_schedule.v10";
const defaultOutputPath = path.join(repoRoot, "tools", "service_backed_schedule_manifest.json");
const makeTargetBaselineSchemaID = "cartulary.scheduler_work_unit_duration_baselines.v1";
const defaultBrowserFunctionalMinShards = 2;
const defaultBrowserFunctionalMaxShards = 4;
const browserCriticalPathPriority = 100;
const measurementIsolationStages = new Set(["webserver-backed", "stateful", "visual"]);

function usage() {
  throw new Error(
    "usage: render-service-backed-schedule-manifest.mjs [--check] [--topology <path>] [--output <path>]",
  );
}

function parseArgs(argv) {
  const options = {
    check: false,
    topology: defaultExecutionTopologyManifestPath,
    output: defaultOutputPath,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--check") {
      options.check = true;
      continue;
    }
    if (arg === "--topology") {
      options.topology = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--output") {
      options.output = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.topology || !options.output) {
    usage();
  }
  return options;
}

function resolvePath(file) {
  return path.isAbsolute(file) ? file : path.join(repoRoot, file);
}

function readJSON(file) {
  return JSON.parse(readFileSync(resolvePath(file), "utf8"));
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

function requireStringArray(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  const seen = new Set();
  const result = [];
  for (const [index, item] of value.entries()) {
    const normalized = requireString(item, `${label}[${index}]`);
    if (seen.has(normalized)) {
      throw new Error(`${label} must not contain duplicate ${normalized}`);
    }
    seen.add(normalized);
    result.push(normalized);
  }
  return result;
}

function cloneObject(value) {
  return JSON.parse(JSON.stringify(value));
}

function repoRelativeOrResolved(file) {
  const resolved = resolvePath(file);
  return path.relative(repoRoot, resolved);
}

function loadMakeTargetDurationBaselines(profile, topologyPath) {
  const baselinePath = requireString(
    profile.defaults.make_target_duration_baseline,
    "defaults.make_target_duration_baseline",
  );
  const resolved = path.isAbsolute(baselinePath)
    ? baselinePath
    : path.join(path.dirname(resolvePath(topologyPath)), baselinePath);
  const baseline = readJSON(resolved);
  if (baseline.schema_id !== makeTargetBaselineSchemaID) {
    throw new Error(
      `${path.relative(repoRoot, resolved)} must declare schema_id ${makeTargetBaselineSchemaID}`,
    );
  }
  if (!Number.isInteger(baseline.default_work_unit_weight_ms) || baseline.default_work_unit_weight_ms <= 0) {
    throw new Error(
      `${path.relative(repoRoot, resolved)} must declare positive integer default_work_unit_weight_ms`,
    );
  }
  if (!baseline.work_units || typeof baseline.work_units !== "object" || Array.isArray(baseline.work_units)) {
    throw new Error(`${path.relative(repoRoot, resolved)} work_units must be an object`);
  }
  const workUnits = new Map();
  for (const [key, entry] of Object.entries(baseline.work_units)) {
    if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
      throw new Error(`${path.relative(repoRoot, resolved)} work_units.${key} must be an object`);
    }
    const expectedKey = [
      entry.scheduler_kind,
      entry.schedule_target,
      entry.work_unit_id,
      entry.aggregate_target,
    ].join("|");
    if (key !== expectedKey) {
      throw new Error(`${path.relative(repoRoot, resolved)} work_units.${key} must match scheduler context key ${expectedKey}`);
    }
    if (!Number.isInteger(entry.duration_ms) || entry.duration_ms <= 0) {
      throw new Error(`${path.relative(repoRoot, resolved)} work_units.${key}.duration_ms must be positive integer weight ms`);
    }
    workUnits.set(key, entry.duration_ms);
  }
  return {
    path: resolved,
    defaultWeightMs: baseline.default_work_unit_weight_ms,
    workUnits,
  };
}

function loadMakeTargetWeightOverrides(profile) {
  const raw = profile.defaults.make_target_weight_overrides ?? {};
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw new Error("defaults.make_target_weight_overrides must be an object when present");
  }
  const now = Date.now();
  const overrides = new Map();
  for (const [target, override] of Object.entries(raw)) {
    if (!override || typeof override !== "object" || Array.isArray(override)) {
      throw new Error(`defaults.make_target_weight_overrides.${target} must be an object`);
    }
    if (!Number.isInteger(override.weight_ms) || override.weight_ms <= 0) {
      throw new Error(`defaults.make_target_weight_overrides.${target}.weight_ms must be positive integer`);
    }
    if (typeof override.reason !== "string" || override.reason.trim() === "") {
      throw new Error(`defaults.make_target_weight_overrides.${target}.reason must be non-empty string`);
    }
    if (typeof override.expires_at !== "string" || Number.isNaN(Date.parse(override.expires_at))) {
      throw new Error(`defaults.make_target_weight_overrides.${target}.expires_at must be an ISO timestamp`);
    }
    if (Date.parse(override.expires_at) <= now) {
      throw new Error(`defaults.make_target_weight_overrides.${target} expired at ${override.expires_at}`);
    }
    overrides.set(target, override.weight_ms);
  }
  return overrides;
}

function workUnitBaselineKey(scheduleTarget, target) {
  return ["service-backed", scheduleTarget, target, target].join("|");
}

function makeTargetWeight(timing, scheduleTarget, target) {
  if (timing.overrides.has(target)) {
    return timing.overrides.get(target);
  }
  const baselineWeight = timing.baseline.workUnits.get(
    workUnitBaselineKey(scheduleTarget, target),
  );
  if (Number.isInteger(baselineWeight) && baselineWeight > 0) {
    return baselineWeight;
  }
  return timing.baseline.defaultWeightMs;
}

function browserGroupBaselineKey(scheduleTarget, groupID, aggregateTarget) {
  return ["service-backed", scheduleTarget, groupID, aggregateTarget].join("|");
}

function browserGroupWeight(timing, scheduleTarget, groupID, aggregateTarget, fallback = 0) {
  const baselineWeight =
    timing.baseline.workUnits.get(browserGroupBaselineKey(scheduleTarget, groupID, aggregateTarget)) ??
    timing.baseline.workUnits.get(
      ["check", "check", `${scheduleTarget}:${groupID}`, aggregateTarget].join("|"),
    );
  if (Number.isInteger(baselineWeight) && baselineWeight > 0) {
    return baselineWeight;
  }
  return Number.isInteger(fallback) && fallback > 0 ? fallback : timing.baseline.defaultWeightMs;
}

function minExecutionDependency(target) {
  const descriptor = findTargetDescriptor(target, repoRoot);
  const dependencies = [
    ...(descriptor?.executionDependencies ?? []),
    ...(descriptor?.supportTargets ?? []),
  ].filter((dependency) => dependency !== "");
  if (dependencies.length === 0) {
    return "";
  }
  return dependencies.sort(compareExecutionDependencies)[0];
}

function compareBackendTargets(left, right) {
  const leftDependency = minExecutionDependency(left);
  const rightDependency = minExecutionDependency(right);
  return (
    compareExecutionDependencies(leftDependency, rightDependency) ||
    String(left).localeCompare(String(right))
  );
}

function backendSelector(scheduleProfile) {
  const selector = scheduleProfile.selectors?.backend ?? {};
  if (!selector || typeof selector !== "object" || Array.isArray(selector)) {
    throw new Error(`${scheduleProfile.target}.selectors.backend must be an object when present`);
  }
  return {
    serviceBacked:
      selector.service_backed === undefined
        ? true
        : requireBoolean(selector.service_backed, `${scheduleProfile.target}.selectors.backend.service_backed`),
    checkServiceBackedSafe:
      selector.check_service_backed_safe === undefined
        ? true
        : requireBoolean(
            selector.check_service_backed_safe,
            `${scheduleProfile.target}.selectors.backend.check_service_backed_safe`,
          ),
  };
}

function browserSelector(scheduleProfile) {
  const selector = scheduleProfile.selectors?.browser;
  if (selector === undefined) {
    return null;
  }
  if (!selector || typeof selector !== "object" || Array.isArray(selector)) {
    throw new Error(`${scheduleProfile.target}.selectors.browser must be an object when present`);
  }
  const tags = requireStringArray(selector.schedule_tags, `${scheduleProfile.target}.selectors.browser.schedule_tags`);
  if (tags.length === 0) {
    throw new Error(`${scheduleProfile.target}.selectors.browser.schedule_tags must not be empty`);
  }
  return { scheduleTags: tags };
}

function orderedServiceBackedBackendTargets(scheduleProfile) {
  const selector = backendSelector(scheduleProfile);
  const targetsWithRows = new Set(
    collectTargetPlanRows(repoRoot)
      .filter((row) => {
        if (row.runner_family !== "go_test") {
          return false;
        }
        if (selector.serviceBacked && row.service_backed !== true) {
          return false;
        }
        if (selector.checkServiceBackedSafe && row.check_service_backed_safe !== true) {
          return false;
        }
        return true;
      })
      .map((row) => row.target),
  );
  return Array.from(targetsWithRows)
    .filter((target) => {
      const descriptor = findTargetDescriptor(target, repoRoot);
      return (
        descriptor?.serviceBacked === selector.serviceBacked &&
        (!selector.checkServiceBackedSafe || descriptor?.checkServiceBackedSafe === true)
      );
    })
    .sort(compareBackendTargets);
}

function backendSource(profile, timing, scheduleTarget, target) {
  const descriptor = findTargetDescriptor(target, repoRoot);
  if (!descriptor) {
    throw new Error(`unknown backend target ${target}`);
  }
  if (!descriptor.serviceBacked) {
    throw new Error(`backend target ${target} is not service-backed`);
  }
  if (descriptor.sharding === "go_shards") {
    return {
      type: "go_shards",
      class: "backend",
      target,
      resource_claims: cloneObject(profile.defaults.go_shards_resource_claims),
    };
  }
  const claims = requireObject(
    profile.defaults.backend_make_target_resource_claims,
    "defaults.backend_make_target_resource_claims",
  );
  if (!claims[target]) {
    throw new Error(`defaults.backend_make_target_resource_claims must declare ${target}`);
  }
  return {
    type: "make_target",
    class: "backend",
    target,
    weight: makeTargetWeight(timing, scheduleTarget, target),
    resource_claims: cloneObject(claims[target]),
  };
}

function browserStageNeeds(stage, selectedTargets, scheduleTarget) {
  const needs = stage.schedulerNeeds ?? [];
  for (const need of needs) {
    if (need === stage.target) {
      throw new Error(`${scheduleTarget} browser stage ${stage.name} must not depend on itself`);
    }
    if (!selectedTargets.has(need)) {
      throw new Error(
        `${scheduleTarget} browser stage ${stage.name} scheduler_needs target ${need} is not selected by the schedule`,
      );
    }
  }
  return needs;
}

function browserStageResourceClaims(profile, stageName) {
  const raw = profile.defaults.browser_stage_resource_claims ?? {};
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw new Error("defaults.browser_stage_resource_claims must be an object when present");
  }
  const stageClaims = raw[stageName] ?? {};
  if (!stageClaims || typeof stageClaims !== "object" || Array.isArray(stageClaims)) {
    throw new Error(`defaults.browser_stage_resource_claims.${stageName} must be an object when present`);
  }
  for (const [resource, amount] of Object.entries(stageClaims)) {
    if (amount !== "limit" && (!Number.isInteger(amount) || amount < 1)) {
      throw new Error(`defaults.browser_stage_resource_claims.${stageName}.${resource} must be a positive integer or "limit"`);
    }
  }
  return cloneObject(stageClaims);
}

function browserFunctionalSharding(profile) {
  const raw = profile.defaults.browser_functional_sharding ?? {};
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw new Error("defaults.browser_functional_sharding must be an object when present");
  }
  const minShards = raw.min_shards ?? defaultBrowserFunctionalMinShards;
  const maxShards = raw.max_shards ?? defaultBrowserFunctionalMaxShards;
  if (!Number.isInteger(minShards) || minShards < 1) {
    throw new Error("defaults.browser_functional_sharding.min_shards must be a positive integer when present");
  }
  if (!Number.isInteger(maxShards) || maxShards < 1) {
    throw new Error("defaults.browser_functional_sharding.max_shards must be a positive integer when present");
  }
  if (minShards > maxShards) {
    throw new Error("defaults.browser_functional_sharding.min_shards must be less than or equal to max_shards");
  }
  return { minShards, maxShards };
}

function browserGroupSources(profile, timing, scheduleTarget, stage) {
  const groups = [];
  const functionalSharding = browserFunctionalSharding(profile);
  for (const group of stage.groups) {
    if (stage.name === "webserver-backed" && group.kind === "duration_balanced_specs") {
      const plan = createBrowserShardPlan({
        baselineFile: path.join(repoRoot, "tools", "browser_e2e_duration_baselines.json"),
        minShards: functionalSharding.minShards,
        maxShards: functionalSharding.maxShards,
      });
      for (const [index, shard] of plan.shards.entries()) {
        const id = `${stage.target}:${shard.name}`;
        groups.push({
          id,
          name: shard.name,
          kind: "functional_shard",
          target: stage.target,
          aggregate_target: stage.target,
          coverage: group.coverage,
          execution_dependency: group.executionDependency,
          shard_name: shard.name,
          shard_index: index,
          shard_count: plan.shard_count,
          phases: shard.phases,
          entry_ids: shard.entries.map((entry) => entry.id),
          scheduler_priority: browserCriticalPathPriority,
          weight: shard.weight_ms,
          resource_claims: {
            go_cpu: 1,
            go_io: 1,
          },
        });
      }
      groups.push({
        id: `${stage.target}:support`,
        name: "support",
        kind: "support",
        target: stage.target,
        aggregate_target: stage.target,
        coverage: "supplemental",
        execution_dependency: "browser_support",
        scheduler_priority: browserCriticalPathPriority,
        weight: browserGroupWeight(timing, scheduleTarget, `${stage.target}:support`, stage.target),
        resource_claims: {
          go_cpu: 1,
          go_io: 1,
        },
      });
      continue;
    }

    const id = `${stage.target}:${group.name}`;
    groups.push({
      id,
      name: group.name,
      kind: group.kind,
      target: group.target,
      aggregate_target: stage.target,
      coverage: group.coverage,
      execution_dependency: group.executionDependency,
      scheduler_priority: browserCriticalPathPriority,
      weight: browserGroupWeight(timing, scheduleTarget, id, stage.target, makeTargetWeight(timing, scheduleTarget, group.target)),
      resource_claims: {
        go_cpu: 1,
        go_io: 1,
        ...browserStageResourceClaims(profile, stage.name),
      },
    });
  }
  return groups;
}

function browserSource(profile, timing, scheduleTarget, stage, selectedTargets, generatedNeeds = []) {
  const stageName = stage.name;
  const laneResource = browserStageResource(stageName);
  const claims = {
    ...cloneObject(profile.defaults.browser_make_target_resource_claims),
    ...browserStageResourceClaims(profile, stageName),
    [laneResource]: 1,
  };
  const needs = Array.from(
    new Set([...browserStageNeeds(stage, selectedTargets, scheduleTarget), ...generatedNeeds]),
  );
  return {
    type: "browser_stage",
    class: "browser",
    target: stage.target,
    browser_stage: stageName,
    ...(needs.length > 0 ? { needs } : {}),
    scheduler_priority: browserCriticalPathPriority,
    weight: browserGroupSources(profile, timing, scheduleTarget, stage)
      .reduce((sum, group) => sum + group.weight, 0),
    resource_claims: claims,
    groups: browserGroupSources(profile, timing, scheduleTarget, stage),
  };
}

function stageHasRequiredTag(stage, requiredTags) {
  const tags = new Set(stage.scheduleTags ?? []);
  return requiredTags.every((tag) => tags.has(tag));
}

function stageNonRawExecutionDependencies(stage) {
  return Array.from(
    new Set(
      stage.groups
        .filter((group) => group.coverage !== "raw")
        .map((group) => group.executionDependency)
        .filter((dependency) => dependency !== ""),
    ),
  );
}

function validateStageHasServiceBackedEvidence(stage, scheduleTarget) {
  const dependencies = stageNonRawExecutionDependencies(stage);
  const rawOnlyVisual =
    stage.target === "browser-e2e-visual" &&
    stage.groups.every((group) => group.coverage === "raw" && group.kind === "visual");
  if (dependencies.length === 0 && !rawOnlyVisual) {
    throw new Error(`${scheduleTarget} browser stage ${stage.name} has no non-raw execution dependencies`);
  }
  for (const group of stage.groups.filter((candidate) => candidate.coverage !== "raw")) {
    const dependency = group.executionDependency;
    const info = executionDependencyInfo(dependency);
    if (!info || info.category !== "browser" || info.service_backed !== true) {
      throw new Error(
        `${scheduleTarget} browser stage ${stage.name} dependency ${dependency} is not service-backed browser evidence`,
      );
    }
    if (!hasPlaywrightRows(group.coverage, dependency)) {
      throw new Error(
        `${scheduleTarget} browser stage ${stage.name} has no phase-map Playwright rows for ${group.coverage} ${dependency}`,
      );
    }
  }
}

function hasPlaywrightRows(coverage, executionDependency) {
  for (const phase of phaseManifestNames(repoRoot)) {
    const { manifest } = loadManifest(repoRoot, phase);
    if (
      collectEntries(manifest).some(
        (entry) =>
          entry.runner === "playwright" &&
          entry.coverage === coverage &&
          entry.execution_dependency === executionDependency,
      )
    ) {
      return true;
    }
  }
  return false;
}

function selectedBrowserStages(scheduleProfile, browserStages) {
  const selector = browserSelector(scheduleProfile);
  if (!selector) {
    return [];
  }
  const stages = Array.from(browserStages.values())
    .filter((stage) => stageHasRequiredTag(stage, selector.scheduleTags))
    .sort((left, right) => {
      const leftDependency = stageNonRawExecutionDependencies(left).sort(compareExecutionDependencies)[0] ?? "";
      const rightDependency = stageNonRawExecutionDependencies(right).sort(compareExecutionDependencies)[0] ?? "";
      return (
        compareExecutionDependencies(leftDependency, rightDependency) ||
        left.name.localeCompare(right.name)
      );
    });
  for (const stage of stages) {
    validateStageHasServiceBackedEvidence(stage, scheduleProfile.target);
  }
  return stages;
}

function renderSchedule(profile, timing, scheduleProfile, browserStages) {
  const target = requireString(scheduleProfile.target, "schedules[].target");
  if (scheduleProfile.resource_limits !== undefined) {
    throw new Error(`${target}.resource_limits is obsolete; use capacity_profile`);
  }
  const capacityProfile = requireString(scheduleProfile.capacity_profile, `${target}.capacity_profile`);
  const profileLimits = resourceLimitsForCapacityProfile(capacityProfile, `${target}.capacity_profile`, {
    scheduler: "service_backed",
    allowAuto: true,
  });
  const resourceLimits = Object.fromEntries(profileLimits.limits.entries());
  const sources = [];
  for (const backendTarget of orderedServiceBackedBackendTargets(scheduleProfile)) {
    sources.push(backendSource(profile, timing, target, backendTarget));
  }
  const backendTargets = sources
    .filter((source) => source.class === "backend")
    .map((source) => source.target);
  const stages = selectedBrowserStages(scheduleProfile, browserStages);
  const selectedTargets = new Set([
    ...backendTargets,
    ...stages.map((stage) => stage.target),
  ]);
  for (const stage of stages) {
    resourceLimits[browserStageResource(stage.name)] = 1;
    const generatedNeeds = stage.name === "measurement"
      ? measurementGeneratedNeeds(stages, scheduleProfile.target)
      : [];
    const source = browserSource(
      profile,
      timing,
      target,
      stage,
      selectedTargets,
      generatedNeeds,
    );
    sources.push(source);
  }
  return {
    target,
    capacity_profile: capacityProfile,
    resource_limits: resourceLimits,
    work_unit_sources: sources,
  };
}

function measurementGeneratedNeeds(stages, scheduleTarget) {
  const dependencies = [];
  for (const stage of stages) {
    if (stage.name === "measurement") {
      continue;
    }
    if (!measurementIsolationStages.has(stage.name)) {
      throw new Error(
        `${scheduleTarget} browser measurement isolation must explicitly account for newly selected stage ${stage.name}`,
      );
    }
    dependencies.push(stage.target);
  }
  return dependencies;
}

export function renderServiceBackedScheduleManifest(options = {}) {
  const topologyPath = options.topology ?? defaultExecutionTopologyManifestPath;
  const topology = options.topologyObject ?? loadExecutionTopology({ manifestPath: topologyPath });
  const profile = renderServiceBackedScheduleProfile(topology);
  requireObject(profile.defaults, "defaults");
  if (profile.defaults.backend_make_target_weights !== undefined) {
    throw new Error("defaults.backend_make_target_weights is obsolete; use make_target_duration_baseline");
  }
  if (profile.defaults.browser_stage_weights !== undefined) {
    throw new Error("defaults.browser_stage_weights is obsolete; use make_target_duration_baseline");
  }
  const timing = {
    baseline: loadMakeTargetDurationBaselines(profile, topologyPath),
    overrides: loadMakeTargetWeightOverrides(profile),
  };
  const browserStages = normalizeBrowserBatchStages(renderBrowserBatchManifest(topology));
  return {
    schema_id: scheduleSchemaID,
    generated: {
      generator: "scripts/render-service-backed-schedule-manifest.mjs",
      topology: path.relative(repoRoot, resolvePath(topologyPath)),
      browser_batch_manifest: topology.generatedOutputs.browser_e2e_batch_manifest,
      make_target_duration_baseline: repoRelativeOrResolved(timing.baseline.path),
    },
    schedules: requireArray(profile.schedules, "schedules").map((schedule) =>
      renderSchedule(profile, timing, schedule, browserStages),
    ),
  };
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const rendered = `${JSON.stringify(renderServiceBackedScheduleManifest(options), null, 2)}\n`;
  const outputPath = resolvePath(options.output);
  if (options.check) {
    const existing = readFileSync(outputPath, "utf8");
    if (existing !== rendered) {
      throw new Error(`${path.relative(repoRoot, outputPath)} is stale; run make phase-schedules`);
    }
    return;
  }
  writeFileSync(outputPath, rendered);
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    main();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.error(`service-backed schedule render failed: ${message}`);
    process.exit(1);
  }
}
