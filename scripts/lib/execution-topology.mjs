import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  assertKnownResource,
  normalizeResourceClaims,
  normalizeResourceLimits,
} from "./scheduler-resources.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");
export const executionTopologySchemaID = "cartulary.execution_topology.v1";
export const defaultExecutionTopologyManifestPath = path.join(
  repoRoot,
  "tools",
  "execution_topology_manifest.json",
);
export const taskSurfaceSchemaID = "cartulary.task_surface_manifest.v9";
export const checkScheduleSchemaID = "cartulary.check_schedule.v6";
export const serviceBackedScheduleSchemaID = "cartulary.service_backed_schedule.v8";
export const browserBatchManifestSchemaID = "cartulary.browser_e2e_batch_manifest.v4";
export const makeTargetBaselineSchemaID =
  "cartulary.service_backed_make_target_duration_baselines.v1";

const validDependencyCategories = new Set(["backend", "frontend", "browser"]);
const validShardModes = new Set(["none", "go_shards"]);
const validParallelismModes = new Set(["none", "package", "process"]);
const validBrowserCoverage = new Set(["authoritative", "supplemental", "raw"]);
const validBrowserDependencyPolicies = new Set([
  "parallel",
  "after_backend",
  "after_prior_browser",
  "after_backend_and_prior_browser",
]);
const serviceRequirementsRequiringCheckServiceStack = new Set(["postgres", "minio", "browser_stack"]);

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

function requireSchema(manifest, schemaID, label) {
  if (manifest.schema_id !== schemaID) {
    throw new Error(`${label} must declare schema_id ${schemaID}`);
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
    "check_schedule_manifest",
    "service_backed_schedule_manifest",
    "browser_e2e_batch_manifest",
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
      const policy = requireString(stage.scheduler_dependency_policy, `${label}.scheduler_dependency_policy`);
      if (!validBrowserDependencyPolicies.has(policy)) {
        throw new Error(`${label}.scheduler_dependency_policy is invalid`);
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

function validateCheckSchedules(topology, taskTargets, taskTargetEntries) {
  for (const [scheduleIndex, schedule] of requireNonEmptyArray(
    topology.check_schedules,
    "check_schedules",
  ).entries()) {
    const label = `check_schedules[${scheduleIndex + 1}]`;
    const target = requireString(schedule?.target, `${label}.target`);
    if (!taskTargets.has(target)) {
      throw new Error(`${label}.target ${target} is missing from task_surface.targets`);
    }
    const resourceLimits = normalizeResourceLimits(schedule.resource_limits, label, {
      scheduler: "check",
    }).limits;
    const units = requireNonEmptyArray(schedule.work_units, `${label}.work_units`);
    const unitTargets = new Set();
    for (const [unitIndex, unit] of units.entries()) {
      const unitLabel = `${label}.work_units[${unitIndex + 1}]`;
      const unitTarget = requireString(unit?.target, `${unitLabel}.target`);
      if (!taskTargets.has(unitTarget)) {
        throw new Error(`${unitLabel}.target ${unitTarget} is missing from task_surface.targets`);
      }
      if (unitTargets.has(unitTarget)) {
        throw new Error(`${label} contains duplicate work unit ${unitTarget}`);
      }
      unitTargets.add(unitTarget);
      const claims = normalizeResourceClaims(unit.resource_claims ?? {}, unitLabel, resourceLimits, {
        scheduler: "check",
        allowBounded: true,
      });
      if (unit.nested_scheduler?.forwarding) {
        assertKnownResource("service_stack", `${unitLabel}.resource_claims.service_stack`, {
          scheduler: "check",
        });
        if (!claims.has("service_stack")) {
          throw new Error(`${unitLabel}.nested_scheduler forwarding must claim service_stack`);
        }
      }
      if (
        requiresCheckServiceStack(taskTargetEntries.get(unitTarget)) &&
        !claims.has("service_stack") &&
        unit.nested_scheduler?.type !== "service_backed"
      ) {
        throw new Error(
          `${unitLabel}.target ${unitTarget} declares service_requirements and must claim service_stack or use a nested service-backed scheduler`,
        );
      }
    }
    assertAcyclicUnits(target, units);
  }
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
    throw new Error(`service-backed make-target duration baseline is missing: ${baselinePath}`);
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
    normalizeResourceLimits(schedule.resource_limits, label, {
      scheduler: "service_backed",
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
  validateCheckSchedules(raw, taskTargets, taskTargetEntries);
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
    checkSchedules: clone(raw.check_schedules),
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
  return {
    schema_id: taskSurfaceSchemaID,
    ...clone(topology.taskSurface),
  };
}

export function renderBrowserBatchManifest(topology) {
  return {
    schema_id: browserBatchManifestSchemaID,
    ...clone(topology.browserBatch),
  };
}

export function renderCheckScheduleManifest(topology) {
  return {
    schema_id: checkScheduleSchemaID,
    schedules: clone(topology.checkSchedules),
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
