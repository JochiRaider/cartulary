import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { normalizeBrowserBatchStages } from "./browser-batch-manifest.mjs";
import {
  compareExecutionDependencies,
  executionDependencyInfo,
} from "./execution-dependencies.mjs";
import {
  loadExecutionTopology,
  renderBrowserBatchManifest,
} from "./execution-topology.mjs";
import { phaseManifestNames } from "./phase-manifest.mjs";
import { activePhaseRegistryEntry, phaseRegistryEntry } from "./phase-registry.mjs";
import { collectGoShardsForTarget } from "./go-shard-plan.mjs";
import { browserStageResource } from "./scheduler-resources.mjs";
import { phaseGuidance, phaseSlice as guidancePhaseSlice } from "./task-guidance.mjs";
import { findTargetDescriptor } from "./target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");
export const phaseSlicePlanSchemaID = "cartulary.phase_slice_plan.v1";

const goCPUResource = "go_cpu";
const goIOResource = "go_io";
const postgresResetResource = "postgres_reset";
const postgresCloneResource = "postgres_clone";
const browserStackResource = "browser_stack";

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function uniqueSorted(values) {
  return Array.from(new Set(values.filter(Boolean))).sort(compareStrings);
}

function validPhaseName(value) {
  return /^phase[0-9]+$/.test(value);
}

const preferredBrowserStageByDependency = new Map([
  ["browser_functional", "webserver-backed"],
  ["browser_stateful", "stateful"],
  ["browser_measurement", "measurement"],
  ["browser_visual", "visual"],
  ["browser_a11y", "a11y"],
  ["browser_a11y_preflight", "a11y-preflight"],
]);

function resolveBrowserStagesByTarget(root = repoRoot) {
  const topology = loadExecutionTopology({
    manifestPath: path.join(root, "tools", "execution_topology_manifest.json"),
  });
  const stages = normalizeBrowserBatchStages(renderBrowserBatchManifest(topology));
  const byTarget = new Map();
  for (const stage of stages.values()) {
    if (!byTarget.has(stage.target)) {
      byTarget.set(stage.target, []);
    }
    byTarget.get(stage.target).push(stage);
  }
  return byTarget;
}

function browserStageDependencies(stage) {
  return uniqueSorted(
    stage.groups
      .filter((group) => group.coverage !== "raw")
      .map((group) => group.executionDependency)
      .filter(Boolean),
  );
}

function resolveBrowserStageForRows(target, rows, stageByTarget) {
  const candidates = stageByTarget.get(target) ?? [];
  if (candidates.length === 0) {
    throw new Error(`phase slice browser target ${target} is not a browser batch stage target`);
  }
  if (candidates.length === 1) {
    return candidates[0];
  }

  const dependencies = executionDependenciesForTarget(rows, target);
  const preferredNames = uniqueSorted(
    dependencies.map((dependency) => preferredBrowserStageByDependency.get(dependency)),
  );
  const preferredCandidates = candidates.filter((stage) => preferredNames.includes(stage.name));
  if (preferredCandidates.length === 1) {
    return preferredCandidates[0];
  }

  const matchingCandidates = candidates.filter((stage) => {
    const stageDependencies = new Set(browserStageDependencies(stage));
    return dependencies.every((dependency) => stageDependencies.has(dependency));
  });
  if (matchingCandidates.length === 1) {
    return matchingCandidates[0];
  }

  throw new Error(
    `phase slice browser target ${target} matches multiple browser batch stages; declare an explicit dependency-to-stage selector`,
  );
}

function phaseRows(phase, mode, root = repoRoot, taskSurfaceManifest = null) {
  const registryEntry = phaseRegistryEntry(root, phase);
  if (!registryEntry) {
    throw new Error(`unknown phase ${phase}; expected one of tools/phase_registry.json`);
  }
  if (!activePhaseRegistryEntry(root, phase)) {
    throw new Error(`phase ${phase} is ${registryEntry.status} and is not executable`);
  }
  const info = phaseGuidance(phase, {
    root,
    includeExecutionMap: false,
    taskSurfaceManifest,
  });
  if (!info) {
    throw new Error(`unknown phase ${phase}; expected one of tools/phase_registry.json`);
  }
  if (mode === "phase") {
    return info.rows;
  }
  return info.rows.filter((row) => executionDependencyInfo(row.execution_dependency)?.service_backed === true);
}

function childTargetsForRows(rows, phase, mode, root = repoRoot, taskSurfaceManifest = null) {
  const rowTargets = new Set(rows.map((row) => row.target));
  return (guidancePhaseSlice(
    phase,
    {
      root,
      serviceBackedOnly: mode === "service_backed",
      includeExecutionMap: false,
      taskSurfaceManifest,
    },
  )?.child_targets ?? [])
    .filter((target) => rowTargets.has(target.target));
}

function executionDependenciesForTarget(rows, target) {
  return uniqueSorted(rows.filter((row) => row.target === target).map((row) => row.execution_dependency))
    .sort(compareExecutionDependencies);
}

function serviceRequirementsForRows(rows) {
  const requirements = new Set();
  for (const row of rows) {
    const info = executionDependencyInfo(row.execution_dependency);
    if (info?.service_backed) {
      if (info.category === "browser") {
        requirements.add("browser_stack");
      }
      requirements.add("postgres");
      requirements.add("object_store");
    }
  }
  return Array.from(requirements).sort(compareStrings);
}

function runtimeBinariesForRows(rows) {
  return uniqueSorted(rows.flatMap((row) => row.runtime_binaries ?? []));
}

function disabledFrontendRowAccountingScope(phase) {
  return {
    mode: "disabled",
    invocation_kind: "base_phase_slice",
    phase_namespace: "base",
    phase,
    selection_policy: "base_phase_no_frontend_rows",
    selected_row_ids: [],
  };
}

function claimStatusCounts(rows) {
  const counts = {
    implemented: 0,
    blocked: 0,
    not_applicable: 0,
    unspecified: 0,
  };
  for (const row of rows) {
    if (row.coverage !== "authoritative") {
      continue;
    }
    const status = row.claim_status ?? "";
    if (Object.hasOwn(counts, status)) {
      counts[status] += 1;
    } else {
      counts.unspecified += 1;
    }
  }
  return counts;
}

function aggregateClaimStatus(counts) {
  if (counts.blocked > 0 || counts.unspecified > 0) {
    return "incomplete";
  }
  if (counts.implemented > 0 || counts.not_applicable > 0) {
    return "complete";
  }
  return "not_applicable";
}

function executableRows(rows) {
  return rows.filter((row) => row.claim_status !== "blocked");
}

function rowGroups(rows) {
  const groups = new Map();
  for (const row of rows) {
    const key = [
      row.runner,
      row.execution_dependency,
      row.target,
      row.execution_family,
      row.coverage,
    ].join("\u001f");
    if (!groups.has(key)) {
      groups.set(key, {
        runner: row.runner,
        execution_dependency: row.execution_dependency,
        target: row.target,
        execution_family: row.execution_family,
        coverage: row.coverage,
        ids: [],
      });
    }
    groups.get(key).ids.push(row.id);
  }
  return Array.from(groups.values())
    .map((group) => ({
      ...group,
      row_count: group.ids.length,
      ids: group.ids.sort(compareStrings),
    }))
    .sort(
      (left, right) =>
        compareExecutionDependencies(left.execution_dependency, right.execution_dependency) ||
        compareStrings(left.runner, right.runner) ||
        compareStrings(left.target, right.target) ||
        compareStrings(left.execution_family, right.execution_family) ||
        compareStrings(left.coverage, right.coverage),
    );
}

function targetWeight(rows) {
  return Math.max(1, rows.length) * 1000;
}

function availableCPUCount() {
  if (typeof os.availableParallelism === "function") {
    return Math.max(1, os.availableParallelism());
  }
  return Math.max(1, os.cpus().length);
}

function defaultGoCPULimit() {
  const configured = Number.parseInt(process.env.CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT ?? "", 10);
  if (Number.isInteger(configured) && configured > 0) {
    return configured;
  }
  const cpus = availableCPUCount();
  return Math.max(2, Math.min(8, cpus <= 4 ? cpus : Math.floor(cpus * 0.75)));
}

function defaultGoIOLimit(goCPULimit) {
  const configured = Number.parseInt(process.env.CARTULARY_SERVICE_BACKED_GO_IO_LIMIT ?? "", 10);
  if (Number.isInteger(configured) && configured > 0) {
    return configured;
  }
  return Math.max(4, Math.min(12, goCPULimit + 4));
}

function defaultBrowserStackLimit() {
  const configured = Number.parseInt(process.env.CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT ?? "", 10);
  if (Number.isInteger(configured) && configured > 0) {
    return configured;
  }
  return availableCPUCount() >= 8 ? 2 : 1;
}

function schedulerClaimsForShard(shard, resourceLimits) {
  switch (shard.scheduler_profile) {
    case "cpu_heavy":
      return new Map([
        [goCPUResource, 2],
        [goIOResource, 1],
      ]);
    case "io_heavy":
      return new Map([
        [goCPUResource, 1],
        [goIOResource, 2],
      ]);
    case "reset_heavy":
      if (!resourceLimits.has(postgresResetResource)) {
        throw new Error(
          `go shard ${shard.name} has reset_heavy profile but phase slice is missing ${postgresResetResource}`,
        );
      }
      return new Map([
        [goCPUResource, 1],
        [goIOResource, 3],
        [postgresResetResource, 1],
      ]);
    case "clone_heavy":
      if (!resourceLimits.has(postgresCloneResource)) {
        throw new Error(
          `go shard ${shard.name} has clone_heavy profile but phase slice is missing ${postgresCloneResource}`,
        );
      }
      return new Map([
        [goCPUResource, 1],
        [goIOResource, 2],
        [postgresCloneResource, 1],
      ]);
    case "transaction_heavy":
      return new Map([
        [goCPUResource, 1],
        [goIOResource, 1],
      ]);
    default:
      return new Map([
        [goCPUResource, 1],
        [goIOResource, 1],
      ]);
  }
}

function mergeClaims(...claimMaps) {
  const merged = new Map();
  for (const claims of claimMaps) {
    for (const [resource, amount] of claims.entries()) {
      merged.set(resource, (merged.get(resource) ?? 0) + amount);
    }
  }
  return merged;
}

function runtimeBinariesForShard(shard) {
  return uniqueSorted((shard.items ?? []).flatMap((item) => item.runtime_binaries ?? []));
}

function backendProcessClaimsForShard(target, _runtimeBinaries, resourceLimits) {
  if (target !== "backend-process") {
    return new Map();
  }
  if (!resourceLimits.has("process")) {
    throw new Error("backend-process Go shards require resource_limits.process");
  }
  return new Map([["process", 1]]);
}

function shardCompletionKey(shardName) {
  return `go_shard:${shardName}`;
}

function addGoUnits(plan, target, rows) {
  const descriptor = findTargetDescriptor(target, plan.root);
  if (!descriptor) {
    throw new Error(`phase slice target ${target} is not in target-plan`);
  }

  if (descriptor.sharding !== "go_shards") {
    const runtimeBinaries = runtimeBinariesForRows(rows);
    plan.workUnits.push({
      id: target,
      label: target,
      kind: "go_target",
      type: "go_target",
      class: "backend",
      target,
      aggregateTarget: target,
      group: target,
      needs: [],
      completionKeys: [target],
      failureKeys: [target],
      weightMs: targetWeight(rows),
      resourceClaims: new Map([[goCPUResource, 1], [goIOResource, 1]]),
      runtime_binaries: runtimeBinaries,
      order: plan.nextOrder++,
    });
    return;
  }

  const shards = collectGoShardsForTarget(plan.root, target, { phase: plan.phase });
  if (shards.length === 0) {
    throw new Error(`phase slice ${plan.phase} selected no Go shards for ${target}`);
  }
  const sourceClaims = new Map([
    ["postgres", 1],
    ["object_store", 1],
  ]);
  plan.workUnits.push({
    id: `finalize:${target}`,
    label: `finalize/${target}`,
    kind: "finalizer",
    type: "finalizer",
    class: "backend",
    target,
    aggregateTarget: target,
    group: target,
    needs: shards.map((shard) => shardCompletionKey(shard.name)),
    completionKeys: [target],
    failureKeys: [target],
    countInTotal: false,
    countsStarted: false,
    resourceClaims: new Map(),
    shardNames: shards.map((shard) => shard.name),
    unblockLabel: target,
    weightMs: 0,
    order: plan.nextOrder++,
  });
  for (const shard of shards) {
    const runtimeBinaries = runtimeBinariesForShard(shard);
    plan.workUnits.push({
      id: `${target}:${shard.name}`,
      label: `${target}/${shard.name}`,
      kind: "go_shard",
      type: "go_shard",
      class: "backend",
      target,
      aggregateTarget: target,
      group: target,
      needs: [],
      completionKeys: [shardCompletionKey(shard.name)],
      failureKeys: [shardCompletionKey(shard.name)],
      runningDependencyKeys: [target],
      completeOnFailure: true,
      shard: shard.name,
      schedulerProfile: shard.scheduler_profile,
      weightMs: shard.weight_ms,
      resourceClaims: mergeClaims(
        sourceClaims,
        schedulerClaimsForShard(shard, plan.resourceLimits),
        backendProcessClaimsForShard(target, runtimeBinaries, plan.resourceLimits),
      ),
      runtime_binaries: runtimeBinaries,
      order: plan.nextOrder++,
    });
  }
}

function addFrontendUnit(plan, target, rows) {
  plan.workUnits.push({
    id: target,
    label: target,
    kind: "frontend_unit",
    type: "make_target",
    class: "frontend",
    target,
    aggregateTarget: target,
    group: target,
    needs: [],
    completionKeys: [target],
    failureKeys: [target],
    weightMs: targetWeight(rows),
    resourceClaims: new Map([["process", 1]]),
    frontend_row_accounting_scope: disabledFrontendRowAccountingScope(plan.phase),
    order: plan.nextOrder++,
  });
}

function addBrowserUnit(plan, target, rows, stageByTarget) {
  const stage = resolveBrowserStageForRows(target, rows, stageByTarget);
  const claims = new Map([
    ["postgres", 1],
    ["object_store", 1],
    ["process", 1],
    [browserStackResource, 1],
    [browserStageResource(stage.name), 1],
  ]);
  plan.browserStages.add(stage.name);
  plan.workUnits.push({
    id: target,
    label: target,
    kind: "browser_target",
    type: "make_target",
    class: "browser",
    target,
    aggregateTarget: target,
    group: target,
    browserStage: stage.name,
    needs: browserNeeds(plan, stage),
    completionKeys: [target],
    failureKeys: [target],
    weightMs: targetWeight(rows),
    resourceClaims: claims,
    frontend_row_accounting_scope: disabledFrontendRowAccountingScope(plan.phase),
    order: plan.nextOrder++,
  });
}

function browserNeeds(plan, stage) {
  const selectedTargets = new Set(plan.child_target_names);
  const needs = stage.schedulerNeeds ?? [];
  for (const need of needs) {
    if (need === stage.target) {
      throw new Error(`phase slice browser target ${stage.target} must not depend on itself`);
    }
    if (!selectedTargets.has(need)) {
      throw new Error(`phase slice browser target ${stage.target} scheduler_needs target ${need} is not selected by the slice`);
    }
  }
  return needs;
}

function resourceLimitObject(resourceLimits) {
  return Object.fromEntries(Array.from(resourceLimits.entries()).sort(([left], [right]) => left.localeCompare(right)));
}

function serializeWorkUnit(unit) {
  const { resourceClaims, weightMs: _weightMs, ...rest } = unit;
  return {
    ...rest,
    weight_ms: unit.weightMs,
    resource_claims: resourceLimitObject(resourceClaims ?? new Map()),
  };
}

function planResourceLimits(rows, root) {
  const hasGo = rows.some((row) => row.runner === "go_test");
  const hasService = rows.some((row) => executionDependencyInfo(row.execution_dependency)?.service_backed === true);
  const hasBrowser = rows.some((row) => executionDependencyInfo(row.execution_dependency)?.category === "browser");
  const hasProcess = rows.some((row) => ["vitest", "playwright"].includes(row.runner) || row.target === "backend-process");
  const goCPU = defaultGoCPULimit();
  const resourceLimits = new Map();
  if (hasGo) {
    resourceLimits.set(goCPUResource, goCPU);
    resourceLimits.set(goIOResource, defaultGoIOLimit(goCPU));
  }
  if (hasService) {
    resourceLimits.set("postgres", 32);
    resourceLimits.set("object_store", 32);
  }
  if (rows.some((row) => row.runner === "go_test" && findTargetDescriptor(row.target, root)?.sharding === "go_shards")) {
    resourceLimits.set(postgresResetResource, 1);
    resourceLimits.set(postgresCloneResource, 4);
  }
  if (hasProcess || hasBrowser) {
    resourceLimits.set("process", 4);
  }
  if (hasBrowser) {
    resourceLimits.set(browserStackResource, defaultBrowserStackLimit());
  }
  return resourceLimits;
}

export function buildPhaseSlicePlan(phase, { mode = "phase", root = repoRoot, taskSurfaceManifest = null } = {}) {
  if (!validPhaseName(phase)) {
    throw new Error(`invalid phase ${phase}; expected phaseN`);
  }
  if (!["phase", "service_backed"].includes(mode)) {
    throw new Error(`invalid phase slice mode ${mode}`);
  }
  const rows = phaseRows(phase, mode, root, taskSurfaceManifest);
  const runnableRows = executableRows(rows);
  const target = mode === "service_backed" ? "service-backed-slice" : "phase-slice";
  const resourceLimits = planResourceLimits(runnableRows, root);
  const claimCounts = claimStatusCounts(rows);
  const plan = {
    schema_id: phaseSlicePlanSchemaID,
    root,
    target,
    phase,
    mode,
    service_backed_only: mode === "service_backed",
    no_op: runnableRows.length === 0,
    phaseClaimStatus: aggregateClaimStatus(claimCounts),
    claimStatusCounts: claimCounts,
    rows,
    row_groups: rowGroups(rows),
    child_targets: childTargetsForRows(runnableRows, phase, mode, root, taskSurfaceManifest),
    child_target_names: [],
    runtime_binaries: runtimeBinariesForRows(runnableRows),
    service_requirements: serviceRequirementsForRows(runnableRows),
    resourceLimits,
    browserStages: new Set(),
    backendTargets: [],
    workUnits: [],
    nextOrder: 0,
  };
  plan.child_target_names = uniqueSorted(plan.child_targets.map((entry) => entry.target));
  plan.backendTargets = plan.child_target_names.filter((name) =>
    executionDependenciesForTarget(rows, name).some((dependency) => executionDependencyInfo(dependency)?.category === "backend"),
  );

  const rowsByTarget = new Map();
  for (const row of runnableRows) {
    if (!rowsByTarget.has(row.target)) {
      rowsByTarget.set(row.target, []);
    }
    rowsByTarget.get(row.target).push(row);
  }
  const stageByTarget = resolveBrowserStagesByTarget(root);
  const orderedTargets = Array.from(rowsByTarget.keys()).sort((left, right) => {
    const leftDeps = executionDependenciesForTarget(rows, left);
    const rightDeps = executionDependenciesForTarget(rows, right);
    return compareExecutionDependencies(leftDeps[0] ?? "", rightDeps[0] ?? "") || left.localeCompare(right);
  });

  for (const targetName of orderedTargets) {
    const targetRows = rowsByTarget.get(targetName);
    const runners = uniqueSorted(targetRows.map((row) => row.runner));
    if (runners.includes("go_test")) {
      addGoUnits(plan, targetName, targetRows.filter((row) => row.runner === "go_test"));
    }
    if (runners.includes("vitest")) {
      addFrontendUnit(plan, targetName, targetRows.filter((row) => row.runner === "vitest"));
    }
    if (runners.includes("playwright")) {
      addBrowserUnit(plan, targetName, targetRows.filter((row) => row.runner === "playwright"), stageByTarget);
    }
  }

  for (const stage of plan.browserStages) {
    resourceLimits.set(browserStageResource(stage), 1);
  }

  const counted = plan.workUnits.filter((unit) => unit.countInTotal !== false);
  counted.sort(
    (left, right) =>
      right.weightMs - left.weightMs ||
      left.order - right.order ||
      left.label.localeCompare(right.label),
  );
  const finalizers = plan.workUnits.filter((unit) => unit.countInTotal === false);
  plan.workUnits = [...counted, ...finalizers];

  return {
    schema_id: plan.schema_id,
    target: plan.target,
    phase: plan.phase,
    mode: plan.mode,
    service_backed_only: plan.service_backed_only,
    no_op: plan.no_op,
    phase_claim_status: plan.phaseClaimStatus,
    claim_status_counts: plan.claimStatusCounts,
    row_groups: plan.row_groups,
    service_requirements: plan.service_requirements,
    child_targets: plan.child_targets,
    child_target_names: plan.child_target_names,
    runtime_binaries: plan.runtime_binaries,
    resource_limits: resourceLimitObject(resourceLimits),
    work_units: plan.workUnits.map(serializeWorkUnit),
    total_work_units: counted.length,
    finalizer_count: finalizers.length,
  };
}

export function validateAllPhaseSlicePlans({ root = repoRoot, taskSurfaceManifest = null } = {}) {
  const known = phaseManifestNames(root);
  for (const phase of known) {
    buildPhaseSlicePlan(phase, { mode: "phase", root, taskSurfaceManifest });
    buildPhaseSlicePlan(phase, { mode: "service_backed", root, taskSurfaceManifest });
  }
}

export function printablePlan(plan) {
  return {
    schema_id: plan.schema_id,
    target: plan.target,
    phase: plan.phase,
    mode: plan.mode,
    no_op: plan.no_op,
    phase_claim_status: plan.phase_claim_status,
    claim_status_counts: plan.claim_status_counts,
    child_targets: plan.child_target_names,
    row_groups: plan.row_groups,
    runtime_binaries: plan.runtime_binaries ?? [],
    service_requirements: plan.service_requirements,
    resource_limits: plan.resource_limits,
    work_units: plan.work_units
      .filter((unit) => unit.countInTotal !== false)
      .map((unit) => ({
        id: unit.id,
        label: unit.label,
        kind: unit.kind,
        target: unit.target,
        needs: unit.needs ?? [],
        resource_claims: unit.resource_claims ?? resourceLimitObject(unit.resourceClaims),
        ...(unit.runtime_binaries?.length > 0 ? { runtime_binaries: unit.runtime_binaries } : {}),
        ...(unit.frontend_row_accounting_scope
          ? { frontend_row_accounting_scope: unit.frontend_row_accounting_scope }
          : {}),
      })),
  };
}
