import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  compareExecutionDependencies,
  executionDependencyInfo,
} from "../execution/execution-dependencies.mjs";
import { phaseManifestNames } from "./phase-manifest.mjs";
import { browserStageResource } from "../scheduler/scheduler-resources.mjs";
import {
  phaseSliceDefaultCapacityProfile,
  resolveSchedulerResourceLimits,
  schedulerCapacityProfileLimits,
} from "../scheduler/scheduler-resource-policy.mjs";
import { addGoUnits, goShardTargetPlanRows } from "./phase-slice-planning/backend-work-units.mjs";
import { addBrowserUnit, resolveBrowserStagesByTarget } from "./phase-slice-planning/browser-work-units.mjs";
import { normalizePhaseSliceSchedulerDAG } from "./phase-slice-planning/scheduler-dag.mjs";
import {
  aggregateClaimStatus,
  childTargetsForRows,
  claimStatusCounts,
  disabledFrontendRowAccountingScope,
  executableRows,
  executionDependenciesForTarget,
  phaseRows,
  rowGroups,
  runtimeBinariesForRows,
  serviceRequirementsForRows,
} from "./phase-slice-planning/row-selection.mjs";
import { targetWeight, uniqueSorted } from "./phase-slice-planning/work-unit-common.mjs";
import {
  phaseSlicePlanSchemaID,
  phaseSlicePlanOutput,
  resourceLimitObject,
  serializePhaseSliceWorkUnit,
  validatePhaseSlicePlanContract,
} from "./phase-slice-plan-contract.mjs";
import {
  parsePhaseRowIDs,
  PhaseSliceSelectionError,
  phaseSliceSelection,
} from "./phase-row-selector.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");

function validPhaseName(value) {
  return /^phase[0-9]+$/.test(value);
}

function phaseSliceProfileResourceLimits(label) {
  return schedulerCapacityProfileLimits(
    "phase_slice",
    phaseSliceDefaultCapacityProfile,
    label,
  );
}

function addGeneratedResourceLimit(resourceLimits, resourceLimitSources, resource, limit) {
  if (!resourceLimits.has(resource)) {
    resourceLimits.set(resource, limit);
    resourceLimitSources.set(resource, "generated");
  }
}

function resolvePlanResourceLimits(plan) {
  const resolved = resolveSchedulerResourceLimits({
    scheduler: "phase_slice",
    resourceLimits: plan.resourceLimits,
    resourceLimitSources: plan.resourceLimitSources,
    label: `${plan.target} ${plan.phase} resource_limits`,
    workUnits: plan.workUnits,
    pruneToClaims: true,
  });
  plan.resourceLimits = resolved.resourceLimits;
  plan.resourceLimitSources = resolved.resourceLimitSources;
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

export function buildPhaseSlicePlan(
  phase,
  { mode = "phase", root = repoRoot, taskSurfaceManifest = null, rowIDs = "" } = {},
) {
  if (!validPhaseName(phase)) {
    throw new PhaseSliceSelectionError(`invalid phase ${phase}; expected phaseN`);
  }
  if (!["phase", "service_backed"].includes(mode)) {
    throw new Error(`invalid phase slice mode ${mode}`);
  }
  const requestedRowIDs = parsePhaseRowIDs(rowIDs);
  const rows = phaseRows(
    phase,
    mode,
    root,
    taskSurfaceManifest,
    requestedRowIDs,
  );
  const runnableRows = executableRows(rows);
  const target = mode === "service_backed" ? "service-backed-slice" : "phase-slice";
  const profileLimits = phaseSliceProfileResourceLimits(`${target} phase slice`);
  const claimCounts = claimStatusCounts(rows);
  const plan = {
    schema_id: phaseSlicePlanSchemaID,
    phase_namespace: "base",
    root,
    target,
    phase,
    mode,
    service_backed_only: mode === "service_backed",
    no_op: runnableRows.length === 0,
    phaseClaimStatus: aggregateClaimStatus(claimCounts),
    claimStatusCounts: claimCounts,
    rows,
    selection: phaseSliceSelection({
      phaseNamespace: "base",
      mode,
      requestedRowIDs,
      resolvedRowIDs: rows.map((row) => row.id),
    }),
    goShardRows: goShardTargetPlanRows(phase, runnableRows, root),
    row_groups: rowGroups(rows),
    child_targets: childTargetsForRows(runnableRows, phase, mode, root, taskSurfaceManifest),
    child_target_names: [],
    runtime_binaries: runtimeBinariesForRows(runnableRows),
    service_requirements: serviceRequirementsForRows(runnableRows),
    resourceLimits: profileLimits.limits,
    resourceLimitSources: profileLimits.sources,
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

  normalizePhaseSliceSchedulerDAG(plan, root);

  for (const stage of plan.browserStages) {
    addGeneratedResourceLimit(
      plan.resourceLimits,
      plan.resourceLimitSources,
      browserStageResource(stage),
      1,
    );
  }
  resolvePlanResourceLimits(plan);

  const counted = plan.workUnits.filter((unit) => unit.countInTotal !== false);
  counted.sort(
    (left, right) =>
      right.weightMs - left.weightMs ||
      left.order - right.order ||
      left.label.localeCompare(right.label),
  );
  const finalizers = plan.workUnits.filter((unit) => unit.countInTotal === false);
  plan.workUnits = [...counted, ...finalizers];

  return validatePhaseSlicePlanContract({
    schema_id: plan.schema_id,
    phase_namespace: plan.phase_namespace,
    target: plan.target,
    phase: plan.phase,
    selection: plan.selection,
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
    resource_limits: resourceLimitObject(plan.resourceLimits),
    work_units: plan.workUnits.map(serializePhaseSliceWorkUnit),
    total_work_units: counted.length,
    finalizer_count: finalizers.length,
  });
}

export function validateAllPhaseSlicePlans({ root = repoRoot, taskSurfaceManifest = null } = {}) {
  const known = phaseManifestNames(root);
  for (const phase of known) {
    buildPhaseSlicePlan(phase, { mode: "phase", root, taskSurfaceManifest });
    buildPhaseSlicePlan(phase, { mode: "service_backed", root, taskSurfaceManifest });
  }
}

export function printablePlan(plan) {
  return phaseSlicePlanOutput(plan);
}
