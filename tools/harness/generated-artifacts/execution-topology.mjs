import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { normalizeRuntimeBinaryEntries } from "../runtime-binary-registry.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";
import { validateSchemaSync } from "../contract/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
export const executionTopologySchemaID = "cartulary.execution_topology.v15";
import { projectOperationBindings, readOperationOwner } from "../scheduler/work-graph/operations.mjs";
export const taskSurfaceOwnerSchemaID = "cartulary.task_surface_owner.v4";
export const taskSurfaceSchemaID = "cartulary.task_surface_manifest.v16";
const schedulerManifestSchemaID = "cartulary.scheduler_manifest.v3";
const browserBatchManifestSchemaID = "cartulary.browser_e2e_batch_manifest.v14";
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
    if (!/^[a-z][a-z0-9-]*$/u.test(entry.target)) throw new Error(`execution dependency ${entry.id} has unknown target ${entry.target}`);
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
  return { targets, byName, runtimeBinariesByFamily };
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
    if (!/^[a-z][a-z0-9-]*$/u.test(stage.target)) throw new Error(`browser stage ${stage.name} has unknown target ${stage.target}`);
    const artifact = raw.browser_e2e_batch.frontend_artifacts?.[stage.frontend_artifact_id];
    if (!artifact || !targets.has(artifact.producer_operation) || !/^[a-z][a-z0-9-]*$/u.test(artifact.producer_report_key) ||
        !/^apps\/web\/dist(?:-measurement)?$/u.test(artifact.path) ||
        !Array.isArray(artifact.entries) || artifact.entries.length === 0 ||
        artifact.entries.some((entry) => !/^[a-z]+\.html$/u.test(entry))) {
      throw new Error(`browser stage ${stage.name} has invalid frontend artifact`);
    }
    const groups = new Set();
    for (const group of assertArray(stage.groups, `browser stage ${stage.name} groups`)) {
      if (groups.has(group.name)) throw new Error(`browser stage ${stage.name} has duplicate group ${group.name}`);
      groups.add(group.name);
      if (group.workers !== "1") throw new Error(`browser group ${group.name} requires exactly one worker`);
      if (!/^[a-z][a-z0-9-]*$/u.test(group.target)) throw new Error(`browser group ${group.name} has unknown target ${group.target}`);
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
  validateSchemaSync(executionTopologySchemaID, raw);
  if (typeof raw.task_surface_owner !== "string" || raw.task_surface_owner.includes(".generated.")) {
    throw new Error("task_surface_owner must reference an authored owner input");
  }
  const taskSurfaceOwnerPath = resolveRepoPath(root, raw.task_surface_owner);
  const taskSurface = readJSON(taskSurfaceOwnerPath);
  requireSchema(taskSurface, taskSurfaceOwnerSchemaID, taskSurfaceOwnerPath);
  if (raw.browser_lane_policy.default > raw.browser_lane_policy.maximum) throw new Error("invalid browser lane policy");
  for (const target of taskSurface.targets) {
    const input = target.input_contract?.inputs.find((entry) => entry.name === "BROWSER_LANES");
    if (input) Object.assign(input, { default: raw.browser_lane_policy.default, max: raw.browser_lane_policy.maximum });
  }
  const operationOwner = readOperationOwner(root);
  const targets = new Set(Object.keys(operationOwner.operations));
  const runtimeProfiles = uniqueByID(raw.runtime_profiles, "runtime_profiles");
  uniqueByID(raw.resource_profiles, "resource_profiles");
  const executionDependencies = normalizeExecutionDependencies(raw, targets);
  const executionDependencyByID = new Map(executionDependencies.map((entry) => [entry.id, entry]));
  const runtimeBinaries = normalizeRuntimeBinaryEntries(raw.runtime_binaries, { operationIDs: targets });
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
    taskSurface: projectOperationBindings(taskSurface, readOperationOwner(root)),
    taskSurfaceOwnerPath,
    browserBatch: clone(raw.browser_e2e_batch),
  };
}

export function resolveBrowserFrontendArtifact(root, stageName, rawTopology) {
  const browser = rawTopology?.browser_e2e_batch ?? loadExecutionTopology({ root }).browserBatch;
  const stage = browser.stages.find((entry) => entry.name === stageName);
  if (!stage) throw new Error(`unknown browser frontend stage ${stageName}`);
  return browser.frontend_artifacts[stage.frontend_artifact_id];
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
      const partitions = new Map();
      for (const row of rows) {
        const key = JSON.stringify([
          row.owner_id, row.selector.project_id, row.runtime_profile_id,
          row.resource_profile_id, row.fixture_capability, row.postgres_isolation, row.fixture_profile_id ?? null,
          row.service_dependencies, row.preparation_operations ?? [], row.browser_state_scope,
          row.browser_state_scope === "test_owned" ? null : row.row_id,
        ]);
        const partition = partitions.get(key) ?? [];
        partition.push(row);
        partitions.set(key, partition);
      }
      return [...partitions.entries()].sort(([a], [b]) => compareASCII(a, b)).map(([key, partition]) => {
        const first = partition[0];
        const identity = createHash("sha256").update(key).digest("hex");
        return {
          ...clone(policyGroup),
          name: `${policyGroup.name}-${identity}`,
          owner_id: first.owner_id,
          state_scope: first.browser_state_scope,
          project: first.selector.project_id,
          config: "playwright.config.ts",
          selected_row_ids: partition.map((row) => row.row_id),
          specs: [...new Set(partition.map((row) => row.selector.file))].sort(compareASCII),
          runtime_profile_id: runtimeProfileID,
          resource_profile_id: first.resource_profile_id,
          service_requirement: runtimeProfile.serviceRequirement,
          service_dependencies: [...first.service_dependencies],
          postgres_isolation: first.postgres_isolation,
          preparation_operations: [...(first.preparation_operations ?? [])],
          browser_session_group: [
            policyGroup.browser_session_group ?? `${stage.target}-${runtimeProfileID}`,
            first.browser_state_scope === "measurement" ? identity : "",
          ].filter(Boolean).join("-"),
          ...(first.fixture_profile_id ? { fixture_profile_id: first.fixture_profile_id } : {}),
        };
      });
    }),
  }));
  return {
    schema_id: browserBatchManifestSchemaID,
    invocation_policy: clone(topology.browserBatch.invocation_policy),
    runtime_profiles: topology.browserBatch.runtime_profiles.map((profile) => ({
      ...clone(profile),
      service_requirement: globalRuntimeProfiles.get(profile.id).serviceRequirement,
    })),
    stages,
  };
}
