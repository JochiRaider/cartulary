#!/usr/bin/env node
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { normalizeBrowserBatchStages } from "../browser/browser-batch-manifest.mjs";
import {
  browserStageGeneratedNeedsPolicyForStage,
  defaultExecutionTopologyManifestPath,
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderServiceBackedScheduleProfile,
} from "./execution-topology.mjs";
import {
  compareExecutionDependencies,
  executionDependencyInfo,
} from "../execution/execution-dependencies.mjs";
import {
  browserStageResource,
  normalizeResourceLimits,
} from "../scheduler/scheduler-resources.mjs";
import {
  runtimeBinaryDefaultEnvForIDs,
  runtimeBinaryProducerTargetsForIDs,
  runtimeBinaryRecordsForIDs,
  runtimeBinaryRegistry,
} from "../runtime-binary-registry.mjs";
import { collectTargetPlanRows, findTargetDescriptor } from "../backend/backend-target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const scheduleSchemaID = "cartulary.service_backed_schedule_sources.v1";
const makeTargetBaselineSchemaID = "cartulary.scheduler_work_unit_duration_baselines.v2";

class UsageError extends Error {}

function usage() {
  throw new UsageError(
    "usage: render-service-backed-schedule-manifest.mjs [--check] [--topology <path>] [--output <path>]",
  );
}

function parseArgs(argv) {
  const options = {
    check: false,
    topology: defaultExecutionTopologyManifestPath,
    output: "",
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

function catalogBrowserRowIndex() {
  const targetByStage = new Map([
    ["webserver_backed", "browser-e2e-webserver-backed"],
    ["support", "browser-e2e-support"],
    ["stateful", "browser-e2e-stateful"],
    ["measurement", "browser-e2e-measurement"],
    ["accessibility", "browser-e2e-a11y"],
    ["visual", "browser-e2e-visual"],
  ]);
  const dependencyByStage = new Map([
    ["webserver_backed", "browser_functional"],
    ["support", "browser_support"],
    ["stateful", "browser_stateful"],
    ["measurement", "browser_measurement"],
    ["accessibility", "browser_a11y"],
    ["visual", "browser_visual"],
  ]);
  const registry = readJSON("tools/test_catalog_owner.json");
  const rows = registry.owners.flatMap((owner) => readJSON(owner.manifest_path).rows);
  return new Map(
    rows
      .filter((row) => row.runner === "playwright")
      .map((row) => [row.row_id, {
        admissible: row.default_check === true,
        implemented: true,
        targets: [targetByStage.get(row.selector.stage)],
        execution_dependency: dependencyByStage.get(row.selector.stage),
      }]),
  );
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

function requirePositiveInteger(value, label) {
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} must be a positive integer`);
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
    if (!Number.isInteger(entry.weight_ms) || entry.weight_ms <= 0) {
      throw new Error(`${path.relative(repoRoot, resolved)} work_units.${key}.weight_ms must be positive integer`);
    }
    workUnits.set(key, entry.weight_ms);
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

function serviceBackedPriorityBands(profile) {
  const raw = requireObject(
    profile.defaults.priority_bands,
    "defaults.priority_bands",
  );
  return {
    browserCriticalPath: requirePositiveInteger(
      raw.browser_critical_path,
      "defaults.priority_bands.browser_critical_path",
    ),
    backendCriticalPath: requirePositiveInteger(
      raw.backend_critical_path,
      "defaults.priority_bands.backend_critical_path",
    ),
    serviceComplete: requirePositiveInteger(
      raw.service_complete,
      "defaults.priority_bands.service_complete",
    ),
  };
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

function uniqueSorted(values) {
  return Array.from(new Set(values.filter(Boolean))).sort((left, right) =>
    String(left).localeCompare(String(right)),
  );
}

function backendSelector(scheduleProfile) {
  const selector = scheduleProfile.selectors?.backend ?? {};
  if (!selector || typeof selector !== "object" || Array.isArray(selector)) {
    throw new Error(`${scheduleProfile.target}.selectors.backend must be an object when present`);
  }
  return {
    enabled:
      selector.enabled === undefined
        ? true
        : requireBoolean(selector.enabled, `${scheduleProfile.target}.selectors.backend.enabled`),
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
    defaultCheckRequired:
      selector.default_check_required === undefined
        ? false
        : requireBoolean(
            selector.default_check_required,
            `${scheduleProfile.target}.selectors.backend.default_check_required`,
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
  return {
    scheduleTags: tags,
    defaultCheckRequired:
      selector.default_check_required === undefined
        ? false
        : requireBoolean(
            selector.default_check_required,
            `${scheduleProfile.target}.selectors.browser.default_check_required`,
          ),
    sessionGroups: browserSessionGroups(
      selector.session_groups,
      `${scheduleProfile.target}.selectors.browser.session_groups`,
    ),
  };
}

function browserSessionGroups(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array when present`);
  }
  const groups = [];
  const seenNames = new Set();
  const seenStages = new Set();
  for (const [index, raw] of value.entries()) {
    const groupLabel = `${label}[${index}]`;
    const group = requireObject(raw, groupLabel);
    const name = requireString(group.name, `${groupLabel}.name`);
    if (seenNames.has(name)) {
      throw new Error(`${label} contains duplicate session group ${name}`);
    }
    seenNames.add(name);
    const stages = requireStringArray(group.stages, `${groupLabel}.stages`);
    if (stages.length === 0) {
      throw new Error(`${groupLabel}.stages must not be empty`);
    }
    for (const stage of stages) {
      if (seenStages.has(stage)) {
        throw new Error(`${label} assigns browser stage ${stage} to multiple session groups`);
      }
      seenStages.add(stage);
    }
    const isolationReason =
      group.isolation_reason === undefined
        ? ""
        : requireString(group.isolation_reason, `${groupLabel}.isolation_reason`);
    groups.push({ name, stages, isolationReason });
  }
  return groups;
}

function browserSessionGroupByStage(groups) {
  const byStage = new Map();
  for (const group of groups ?? []) {
    for (const stage of group.stages) {
      byStage.set(stage, group);
    }
  }
  return byStage;
}

function runtimeBinariesForBackendTarget(scheduleProfile, target) {
  const selector = backendSelector(scheduleProfile);
  return uniqueSorted(
    collectTargetPlanRows(repoRoot)
      .filter((row) => {
        if (row.target !== target || row.runner_family !== "go_test") {
          return false;
        }
        if (selector.serviceBacked && row.service_backed !== true) {
          return false;
        }
        if (selector.checkServiceBackedSafe && row.check_service_backed_safe !== true) {
          return false;
        }
        if (selector.defaultCheckRequired && row.default_check_required !== true) {
          return false;
        }
        return true;
      })
      .flatMap((row) => row.runtime_binaries ?? []),
  );
}

function runtimeBinaryNeeds(profile, ids) {
  return uniqueSorted(
    runtimeBinaryProducerTargetsForIDs(
      runtimeBinaryRegistry(profile.runtime_binaries ?? []),
      ids,
    ),
  );
}

function runtimeBinaryEnv(profile, ids) {
  return runtimeBinaryDefaultEnvForIDs(
    runtimeBinaryRegistry(profile.runtime_binaries ?? []),
    ids,
  );
}

function runtimeBinaryRecords(profile, ids) {
  return runtimeBinaryRecordsForIDs(
    runtimeBinaryRegistry(profile.runtime_binaries ?? []),
    ids,
  );
}

function goShardResourceClaims(profile, target) {
  const claims = {
    ...cloneObject(profile.defaults.go_shards_resource_claims),
  };
  const byTarget = profile.defaults.go_shards_resource_claims_by_target ?? {};
  if (byTarget[target]) {
    Object.assign(claims, cloneObject(byTarget[target]));
  }
  return claims;
}

function goShardResourceClaimsByExecutionFamily(profile) {
  return cloneObject(profile.defaults.go_shards_resource_claims_by_execution_family ?? {});
}

function orderedServiceBackedBackendTargets(scheduleProfile) {
  const selector = backendSelector(scheduleProfile);
  if (!selector.enabled) {
    return [];
  }
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
        if (selector.defaultCheckRequired && row.default_check_required !== true) {
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

function backendSource(profile, timing, scheduleProfile, target, priorities, { defaultCheckOnly = false } = {}) {
  const scheduleTarget = scheduleProfile.target;
  const descriptor = findTargetDescriptor(target, repoRoot);
  if (!descriptor) {
    throw new Error(`unknown backend target ${target}`);
  }
  if (!descriptor.serviceBacked) {
    throw new Error(`backend target ${target} is not service-backed`);
  }
  const runtimeBinaries = runtimeBinariesForBackendTarget(scheduleProfile, target);
  if (descriptor.sharding === "go_shards") {
    return {
      type: "go_shards",
      class: "backend",
      target,
      ...(runtimeBinaries.length > 0 ? { runtime_binary_records: runtimeBinaryRecords(profile, runtimeBinaries) } : {}),
      priority: priorities.backendCriticalPath,
      resource_claims: goShardResourceClaims(profile, target),
      resource_claims_by_execution_family: goShardResourceClaimsByExecutionFamily(profile),
      default_check_required: defaultCheckOnly,
    };
  }
  const runtimeNeeds = runtimeBinaryNeeds(profile, runtimeBinaries);
  const runtimeEnv = runtimeBinaryEnv(profile, runtimeBinaries);
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
    needs: runtimeNeeds,
    ...(runtimeBinaries.length > 0 ? { runtime_binaries: runtimeBinaries } : {}),
    ...(Object.keys(runtimeEnv).length > 0 ? { env: runtimeEnv } : {}),
    priority: priorities.backendCriticalPath,
    weight_ms: makeTargetWeight(timing, scheduleTarget, target),
    resource_claims: cloneObject(claims[target]),
    default_check_required: defaultCheckOnly,
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

function isRetainedBrowserStageResource(resource) {
  return resource === "browser_stack" || resource === "process" || resource.startsWith("browser_stage_");
}

function browserGroupResourceClaims(profile, stageName) {
  const claims = {
    go_cpu: 1,
    go_io: 1,
  };
  if (stageName !== "measurement") {
    return claims;
  }
  for (const [resource, amount] of Object.entries(browserStageResourceClaims(profile, stageName))) {
    if (isRetainedBrowserStageResource(resource)) {
      continue;
    }
    claims[resource] = amount;
  }
  return claims;
}

function browserStageResourceLimit(profile, stageName) {
  const raw = profile.defaults.browser_stage_resource_limits ?? {};
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw new Error("defaults.browser_stage_resource_limits must be an object when present");
  }
  const value = raw[stageName] ?? 1;
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`defaults.browser_stage_resource_limits.${stageName} must be a positive integer`);
  }
  return value;
}

function selectedGroupFields(group) {
  const fields = {};
  if (Array.isArray(group.selectedRowIDs) && group.selectedRowIDs.length > 0) {
    fields.selected_row_ids = [...group.selectedRowIDs];
  }
  return fields;
}

function projectedBrowserRowIDs(rowIDs, rowIndex, { target, executionDependency, label }) {
  const projected = [];
  for (const rowID of rowIDs) {
    const record = rowIndex.get(rowID);
    if (!record) {
      throw new Error(`${label} selected row id ${rowID} is unknown to browser row metadata`);
    }
    if (!record.implemented) {
      throw new Error(`${label} selected row id ${rowID} is not implemented/executable`);
    }
    if (
      target &&
      Array.isArray(record.targets) &&
      record.targets.length > 0 &&
      !record.targets.includes(target)
    ) {
      throw new Error(`${label} selected row id ${rowID} is not mapped to ${target}`);
    }
    if (
      executionDependency &&
      record.execution_dependency &&
      record.execution_dependency !== executionDependency
    ) {
      throw new Error(
        `${label} selected row id ${rowID} has execution_dependency ${record.execution_dependency}, expected ${executionDependency}`,
      );
    }
    if (record.admissible) {
      projected.push(rowID);
    }
  }
  return projected;
}

function projectBrowserGroupForDefaultCheck(group, rowIndex, label) {
  if (group.runtimeProfileID !== "default") {
    return null;
  }
  if (!Array.isArray(group.selectedRowIDs) || group.selectedRowIDs.length === 0) {
    return group;
  }
  const selectedRowIDs = projectedBrowserRowIDs(group.selectedRowIDs, rowIndex, {
    target: group.target,
    executionDependency: group.executionDependency,
    label,
  });
  if (selectedRowIDs.length === 0) {
    return null;
  }
  return { ...group, selectedRowIDs };
}

function browserGroupSessionFields(stage, group, sessionGroup) {
  const fields = { runtime_profile_id: group.runtimeProfileID };
  if (sessionGroup?.name) {
    fields.browser_session_group = sessionGroup.name;
  } else if (group.browserSessionGroup) {
    fields.browser_session_group = group.browserSessionGroup;
  } else if (group.kind === "stateful_partition") {
    const base = sessionGroup?.name ?? stage.target;
    fields.browser_session_group = `${base}-${group.name}`;
  }
  const isolationReason = group.browserSessionIsolationReason || sessionGroup?.isolationReason || "";
  if (fields.browser_session_group && isolationReason) {
    fields.browser_session_isolation_reason = isolationReason;
  }
  return fields;
}

function browserGroupSources(
  profile,
  timing,
  scheduleTarget,
  stage,
  priorities,
  sessionGroup = null,
  { defaultCheckOnly = false, rowIndex = new Map() } = {},
) {
  const groups = [];
  for (const group of stage.groups) {
    const projectedGroup = defaultCheckOnly
      ? projectBrowserGroupForDefaultCheck(
          group,
          rowIndex,
          `${scheduleTarget} browser stage ${stage.name} group ${group.name}`,
        )
      : group;
    if (!projectedGroup) {
      continue;
    }
    const id = `${stage.target}:${group.name}`;
    groups.push({
      id,
      name: projectedGroup.name,
      kind: projectedGroup.kind,
      target: projectedGroup.target,
      aggregate_target: stage.target,
      coverage: projectedGroup.coverage,
      execution_dependency: projectedGroup.executionDependency,
      workers: projectedGroup.workers,
      runtime_profile_id: projectedGroup.runtimeProfileID,
      ...selectedGroupFields(projectedGroup),
      ...browserGroupSessionFields(stage, projectedGroup, sessionGroup),
      priority: priorities.browserCriticalPath,
      weight_ms: browserGroupWeight(
        timing,
        scheduleTarget,
        id,
        stage.target,
        projectedGroup.kind === "stateful_partition"
          ? 0
          : makeTargetWeight(timing, scheduleTarget, projectedGroup.target),
      ),
      resource_claims: browserGroupResourceClaims(profile, stage.name),
    });
  }
  return groups;
}

function browserSource(
  profile,
  timing,
  scheduleTarget,
  stage,
  selectedTargets,
  priorities,
  generatedNeeds = [],
  sessionGroup = null,
  options = {},
) {
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
  const groups = browserGroupSources(
    profile,
    timing,
    scheduleTarget,
    stage,
    priorities,
    sessionGroup,
    options,
  );
  return {
    type: "browser_stage",
    class: "browser",
    target: stage.target,
    browser_stage: stageName,
    default_check_required: options.defaultCheckOnly === true,
    ...(sessionGroup
      ? {
          browser_session_group: sessionGroup.name,
          ...(sessionGroup.isolationReason
            ? { browser_session_isolation_reason: sessionGroup.isolationReason }
            : {}),
        }
      : {}),
    ...(needs.length > 0 ? { needs } : {}),
    priority: priorities.browserCriticalPath,
    weight_ms: groups.reduce((sum, group) => sum + group.weight_ms, 0),
    resource_claims: claims,
    groups,
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
    if (!Array.isArray(group.selectedRowIDs) || group.selectedRowIDs.length === 0) {
      throw new Error(
        `${scheduleTarget} browser stage ${stage.name} has no catalog Playwright rows for ${group.coverage} ${dependency}`,
      );
    }
  }
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
  const capacityProfile = requireString(scheduleProfile.capacity_profile, `${target}.capacity_profile`);
  const profileLimits = normalizeResourceLimits(
    scheduleProfile.resource_limits,
    `${target}.resource_limits`,
    {
    scheduler: "service_backed",
    capacityProfile,
    allowAuto: true,
    },
  );
  const resourceLimits = Object.fromEntries(profileLimits.limits.entries());
  const priorities = serviceBackedPriorityBands(profile);
  const backend = backendSelector(scheduleProfile);
  const sources = [];
  for (const backendTarget of orderedServiceBackedBackendTargets(scheduleProfile)) {
    sources.push(backendSource(profile, timing, scheduleProfile, backendTarget, priorities, {
      defaultCheckOnly: backend.defaultCheckRequired,
    }));
  }
  const backendTargets = sources
    .filter((source) => source.class === "backend")
    .map((source) => source.target);
  const stages = selectedBrowserStages(scheduleProfile, browserStages);
  const selector = browserSelector(scheduleProfile);
  const browserDefaultCheckOnly = selector?.defaultCheckRequired === true;
  const browserRowIndex = browserDefaultCheckOnly
    ? catalogBrowserRowIndex()
    : new Map();
  const sessionGroupsByStage = browserSessionGroupByStage(selector?.sessionGroups ?? []);
  if (selector?.sessionGroups?.length > 0) {
    const selectedStageNames = new Set(stages.map((stage) => stage.name));
    for (const [stageName, group] of sessionGroupsByStage.entries()) {
      if (!selectedStageNames.has(stageName)) {
        throw new Error(
          `${target}.selectors.browser.session_groups.${group.name} references unselected browser stage ${stageName}`,
        );
      }
    }
    for (const stage of stages) {
      if (!sessionGroupsByStage.has(stage.name)) {
        throw new Error(
          `${target}.selectors.browser.session_groups must assign selected browser stage ${stage.name}`,
        );
      }
    }
  }
  const selectedTargets = new Set([
    ...backendTargets,
    ...stages.map((stage) => stage.target),
  ]);
  for (const stage of stages) {
    resourceLimits[browserStageResource(stage.name)] = browserStageResourceLimit(profile, stage.name);
    const generatedNeeds = stage.name === "measurement"
      ? measurementGeneratedNeeds(profile, stages, scheduleProfile.target)
      : [];
    const source = browserSource(
      profile,
      timing,
      target,
      stage,
      selectedTargets,
      priorities,
      generatedNeeds,
      sessionGroupsByStage.get(stage.name) ?? null,
      {
        defaultCheckOnly: browserDefaultCheckOnly,
        rowIndex: browserRowIndex,
      },
    );
    if (source.groups.length === 0) {
      continue;
    }
    sources.push(source);
  }
  return {
    target,
    scheduler_kind: "service_backed",
    capacity_profile: capacityProfile,
    service_complete_priority: priorities.serviceComplete,
    resource_limits: resourceLimits,
    work_unit_sources: sources,
  };
}

function measurementGeneratedNeeds(profile, stages, scheduleTarget) {
  const policy = browserStageGeneratedNeedsPolicyForStage(
    profile,
    "measurement",
    "defaults.browser_stage_generated_needs",
  );
  const selectedPeerStages = new Set(policy.selectedPeerStages);
  const dependencies = [];
  for (const stage of stages) {
    if (stage.name === "measurement") {
      continue;
    }
    if (!selectedPeerStages.has(stage.name)) {
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
      generator: "tools/harness/generated-artifacts/render-service-backed-schedule-manifest.mjs",
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
      throw new Error(`${path.relative(repoRoot, outputPath)} is stale; run make generate`);
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
    process.exit(error instanceof UsageError ? 2 : 1);
  }
}
