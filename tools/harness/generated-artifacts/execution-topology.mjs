import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { normalizeRuntimeBinaryEntries } from "../runtime-binary-registry.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");
export const executionTopologySchemaID = "cartulary.execution_topology.v6";
export const taskSurfaceOwnerSchemaID = "cartulary.task_surface_owner.v2";
export const taskSurfaceSchemaID = "cartulary.task_surface_manifest.v15";
export const schedulerManifestSchemaID = "cartulary.scheduler_manifest.v3";
export const browserBatchManifestSchemaID = "cartulary.browser_e2e_batch_manifest.v8";
export const defaultExecutionTopologyManifestPath = path.join(
  repoRoot,
  "tools",
  "execution_topology_manifest.json",
);

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function clone(value) {
  return structuredClone(value);
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function resolveRepoPath(root, value) {
  return path.isAbsolute(value) ? value : path.join(root, value);
}

function requireSchema(value, schemaID, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  if (value.schema_id !== schemaID) {
    throw new Error(`${label} must declare schema_id ${schemaID}`);
  }
}

function assertArray(value, label) {
  if (!Array.isArray(value)) throw new Error(`${label} must be an array`);
  return value;
}

function uniqueByID(values, label) {
  const byID = new Map();
  for (const value of assertArray(values, label)) {
    if (!value || typeof value.id !== "string" || value.id === "") {
      throw new Error(`${label} entries must declare id`);
    }
    if (byID.has(value.id)) throw new Error(`${label} contains duplicate id ${value.id}`);
    byID.set(value.id, value);
  }
  return byID;
}

function normalizeExecutionDependencies(raw, targets) {
  const dependencies = assertArray(raw.execution_dependencies, "execution_dependencies")
    .map((entry) => ({
      id: String(entry.id ?? ""),
      target: String(entry.target ?? ""),
      category: String(entry.category ?? ""),
      order: Number(entry.order),
      serviceBacked: entry.service_backed === true,
      supportTarget: entry.support_target === true,
    }))
    .sort((left, right) => left.order - right.order || compareASCII(left.id, right.id));
  const ids = new Set();
  for (const entry of dependencies) {
    if (!/^[a-z][a-z0-9_]*$/u.test(entry.id) || ids.has(entry.id)) {
      throw new Error(`execution_dependencies contains invalid or duplicate id ${entry.id}`);
    }
    ids.add(entry.id);
    if (!targets.has(entry.target)) throw new Error(`execution dependency ${entry.id} has unknown target ${entry.target}`);
    if (!new Set(["backend", "frontend", "browser"]).has(entry.category)) {
      throw new Error(`execution dependency ${entry.id} has unknown category ${entry.category}`);
    }
    if (!Number.isInteger(entry.order) || entry.order < 0) {
      throw new Error(`execution dependency ${entry.id} order must be non-negative`);
    }
  }
  return dependencies;
}

function normalizeGoTargets(raw, runtimeBinaries) {
  const binaryByID = new Map(runtimeBinaries.map((entry) => [entry.id, entry]));
  const targets = assertArray(raw.go_targets?.targets, "go_targets.targets").map((entry) => ({
    name: entry.name,
    serviceBacked: entry.service_backed === true,
    checkHeavySafe: entry.check_heavy_safe === true,
    checkServiceBackedSafe: entry.check_service_backed_safe === true,
    checkIsolatedSafe: entry.check_isolated_safe === true,
    canonicalAuthoritative: entry.canonical_authoritative === true,
    sharding: entry.sharding,
    goTestParallelism: entry.go_test_parallelism,
    executionDependencies: [...(entry.execution_dependencies ?? [])],
    supportTargets: [...(entry.support_targets ?? [])],
  }));
  const byName = new Map(targets.map((entry) => [entry.name, entry]));
  if (byName.size !== targets.length) throw new Error("go_targets.targets contains duplicate names");
  const runtimeBinariesByFamily = new Map();
  for (const entry of raw.go_targets?.family_runtime_binaries ?? []) {
    runtimeBinariesByFamily.set(
      entry.family_id,
      entry.runtime_binary_ids.map((id) => {
        const binary = binaryByID.get(id);
        if (!binary) throw new Error(`unknown runtime binary ${id}`);
        return binary;
      }),
    );
  }
  const rawAggregates = (raw.go_targets?.raw_go_aggregates ?? []).map((entry) => ({
    id: entry.id,
    target: entry.target,
    section: entry.section,
    packages: [...entry.packages],
    selectionPattern: entry.selection_pattern,
    executionFamily: entry.execution_family,
    executionLabel: entry.execution_label,
    fixtureCapability: entry.fixture_capability,
    estimatedWorkMs: entry.estimated_work_ms,
  }));
  for (const entry of rawAggregates) {
    if (!new Set(["none", "postgres_transaction", "postgres_group", "postgres_dedicated", "postgres_migration", "object_store_namespace", "managed_process", "browser_stack"]).has(entry.fixtureCapability)) {
      throw new Error(`raw Go aggregate ${entry.id} has invalid fixture capability`);
    }
    if (!Number.isInteger(entry.estimatedWorkMs) || entry.estimatedWorkMs < 1) {
      throw new Error(`raw Go aggregate ${entry.id} has invalid estimated work`);
    }
  }
  return { targets, byName, runtimeBinariesByFamily, rawAggregates };
}

function validateBrowserOwner(raw, dependencyIDs, targets, runtimeProfileIDs) {
  const profileIDs = uniqueByID(raw.browser_e2e_batch?.runtime_profiles, "browser_e2e_batch.runtime_profiles");
  for (const id of profileIDs.keys()) {
    if (!runtimeProfileIDs.has(id)) throw new Error(`browser runtime profile ${id} is unknown`);
  }
  const stageNames = new Set();
  for (const stage of assertArray(raw.browser_e2e_batch?.stages, "browser_e2e_batch.stages")) {
    if (stageNames.has(stage.name)) throw new Error(`duplicate browser stage ${stage.name}`);
    stageNames.add(stage.name);
    if (!targets.has(stage.target)) throw new Error(`browser stage ${stage.name} has unknown target ${stage.target}`);
    const groups = new Set();
    for (const group of assertArray(stage.groups, `browser stage ${stage.name} groups`)) {
      if (groups.has(group.name)) throw new Error(`browser stage ${stage.name} has duplicate group ${group.name}`);
      groups.add(group.name);
      if (!targets.has(group.target)) throw new Error(`browser group ${group.name} has unknown target ${group.target}`);
      if (!dependencyIDs.has(group.execution_dependency)) {
        throw new Error(`browser group ${group.name} has unknown execution dependency ${group.execution_dependency}`);
      }
      if (group.selected_row_ids !== undefined || group.specs !== undefined) {
        throw new Error(`browser group ${group.name} must derive selectors from the catalog`);
      }
    }
  }
}

export function serviceRequirementForRuntimeProfile(profile) {
  return (profile.managed_service_ids ?? []).length > 0 ? "test-services" : "none";
}

export function loadExecutionTopology(options = {}) {
  const root = path.resolve(options.root ?? repoRoot);
  const manifestPath = resolveRepoPath(
    root,
    options.manifestPath ?? process.env.CARTULARY_EXECUTION_TOPOLOGY_MANIFEST ?? defaultExecutionTopologyManifestPath,
  );
  const raw = readJSON(manifestPath);
  requireSchema(raw, executionTopologySchemaID, manifestPath);
  const allowedKeys = new Set([
    "schema_id", "runtime_profiles", "resource_profiles", "generated_outputs",
    "runtime_binaries", "execution_dependencies", "go_targets", "browser_e2e_batch",
    "task_surface_owner",
  ]);
  for (const key of Object.keys(raw)) {
    if (!allowedKeys.has(key)) throw new Error(`${manifestPath} contains unsupported key ${key}`);
  }
  if (typeof raw.task_surface_owner !== "string" || raw.task_surface_owner.includes(".generated.")) {
    throw new Error("task_surface_owner must reference an authored owner input");
  }
  const taskSurfaceOwnerPath = resolveRepoPath(root, raw.task_surface_owner);
  const taskSurface = readJSON(taskSurfaceOwnerPath);
  requireSchema(taskSurface, taskSurfaceOwnerSchemaID, taskSurfaceOwnerPath);
  const targets = new Set(taskSurface.targets.map((entry) => entry.name));
  const runtimeProfiles = uniqueByID(raw.runtime_profiles, "runtime_profiles");
  uniqueByID(raw.resource_profiles, "resource_profiles");
  const executionDependencies = normalizeExecutionDependencies(raw, targets);
  const executionDependencyByID = new Map(executionDependencies.map((entry) => [entry.id, entry]));
  const runtimeBinaries = normalizeRuntimeBinaryEntries(raw.runtime_binaries, { taskTargets: targets });
  validateBrowserOwner(raw, new Set(executionDependencyByID.keys()), targets, new Set(runtimeProfiles.keys()));
  return {
    root,
    manifestPath,
    raw: clone(raw),
    generatedOutputs: clone(raw.generated_outputs),
    executionDependencies,
    executionDependencyByID,
    runtimeBinaries,
    goTargets: normalizeGoTargets(raw, runtimeBinaries),
    taskSurface: clone(taskSurface),
    taskSurfaceOwnerPath,
    browserBatch: clone(raw.browser_e2e_batch),
  };
}

export function renderTaskSurfaceManifest(topology) {
  const taskSurface = clone(topology.taskSurface);
  delete taskSurface.schema_id;
  return { schema_id: taskSurfaceSchemaID, ...taskSurface };
}

function runtimeProfileByID(raw) {
  return new Map(raw.runtime_profiles.map((entry) => [entry.id, {
    ...entry,
    browserCapable: entry.browser_capable === true,
    serviceRequirement: serviceRequirementForRuntimeProfile(entry),
  }]));
}

export function renderBrowserBatchManifest(topology) {
  const globalRuntimeProfiles = runtimeProfileByID(topology.raw);
  const catalogRows = loadTestCatalog(topology.root).rows;
  const selectorStageByKind = new Map([
    ["webserver-backed", "webserver_backed"],
    ["duration_balanced_specs", "webserver_backed"],
    ["functional", "webserver_backed"],
    ["support", "support"],
    ["stateful", "stateful"],
    ["stateful_partition", "stateful"],
    ["measurement", "measurement"],
    ["a11y", "accessibility"],
    ["visual", "visual"],
  ]);
  const stages = topology.browserBatch.stages.map((stage) => ({
    ...clone(stage),
    groups: stage.groups.flatMap((policyGroup) => {
      const selectorStage = selectorStageByKind.get(policyGroup.kind);
      if (!selectorStage) throw new Error(`browser group ${policyGroup.name} has no selector-stage mapping`);
      const runtimeProfileID = policyGroup.runtime_profile_id ?? "default";
      const runtimeProfile = globalRuntimeProfiles.get(runtimeProfileID);
      if (!runtimeProfile?.browserCapable) {
        throw new Error(`browser group ${policyGroup.name} has non-browser runtime ${runtimeProfileID}`);
      }
      const rows = catalogRows
        .filter((row) =>
          row.runner === "playwright" &&
          row.selector.stage === selectorStage &&
          row.runtime_profile_id === runtimeProfileID,
        )
        .sort((left, right) => compareASCII(left.row_id, right.row_id));
      if (rows.length === 0) throw new Error(`browser group ${policyGroup.name} selects no catalog rows`);
      const byFile = new Map();
      for (const row of rows) {
        const fileRows = byFile.get(row.selector.file) ?? [];
        fileRows.push(row);
        byFile.set(row.selector.file, fileRows);
      }
      return [...byFile.entries()]
        .sort(([left], [right]) => compareASCII(left, right))
        .map(([file, fileRows], fileIndex) => {
          const fileIdentity = file
            .replace(/^apps\/web\/e2e\//u, "")
            .replace(/\.spec\.ts$/u, "")
            .replaceAll(/[^a-zA-Z0-9]+/gu, "-")
            .replaceAll(/^-|-$/gu, "")
            .toLowerCase();
          const group = {
            ...clone(policyGroup),
            name: `${policyGroup.name}-${fileIdentity}`,
            selected_row_ids: fileRows.map((row) => row.row_id),
            specs: [file],
            runtime_profile_id: runtimeProfileID,
            service_requirement: runtimeProfile.serviceRequirement,
            browser_session_group:
              policyGroup.browser_session_group ?? `${stage.target}-${runtimeProfileID}`,
          };
          if (selectorStage === "stateful" && fileIndex > 0 && !group.reset_before) {
            group.reset_before = `${policyGroup.name}-before-${fileIdentity}`;
          }
          return group;
        });
    }),
  }));
  return {
    schema_id: browserBatchManifestSchemaID,
    runtime_profiles: topology.browserBatch.runtime_profiles.map((profile) => ({
      ...clone(profile),
      service_requirement: globalRuntimeProfiles.get(profile.id).serviceRequirement,
    })),
    stages,
  };
}

export function executionDependencyMetadata(root = repoRoot) {
  const topology = loadExecutionTopology({ root });
  return new Map(topology.executionDependencies.map((entry) => [entry.id, entry]));
}

export function serviceBackedGoExecutionDependencies(root = repoRoot) {
  return new Set(loadExecutionTopology({ root }).executionDependencies
    .filter((entry) => entry.category === "backend" && entry.serviceBacked && !entry.supportTarget)
    .map((entry) => entry.id));
}

export function serviceBackedSupportTargets(root = repoRoot) {
  return new Set(loadExecutionTopology({ root }).executionDependencies
    .filter((entry) => entry.serviceBacked && entry.supportTarget)
    .map((entry) => entry.target));
}

export function validExecutionDependencyIDs(root = repoRoot) {
  return new Set(executionDependencyMetadata(root).keys());
}

export function validSupportTargetIDs(root = repoRoot) {
  return serviceBackedSupportTargets(root);
}

export function compareExecutionDependencyIDs(left, right, root = repoRoot) {
  const metadata = executionDependencyMetadata(root);
  return (metadata.get(left)?.order ?? Number.MAX_SAFE_INTEGER) -
    (metadata.get(right)?.order ?? Number.MAX_SAFE_INTEGER) || compareASCII(left, right);
}

export function targetForExecutionDependencyID(id, label = "execution_dependency", root = repoRoot) {
  if (id === "") return "";
  const info = executionDependencyMetadata(root).get(id);
  if (!info) throw new Error(`${label} has no execution dependency metadata for ${id}`);
  return info.target;
}

export function topologySummary(topology) {
  return {
    schema_id: topology.raw.schema_id,
    execution_dependencies: topology.executionDependencies.length,
    go_targets: topology.goTargets.targets.length,
    raw_go_aggregates: topology.goTargets.rawAggregates.length,
    task_targets: topology.taskSurface.targets.length,
    browser_stages: topology.browserBatch.stages.length,
    generated_outputs: clone(topology.generatedOutputs),
  };
}
