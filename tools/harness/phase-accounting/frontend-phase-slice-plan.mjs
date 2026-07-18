import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadFrontendPhaseRegistry } from "./frontend/registry-loader.mjs";
import {
  frontendPhaseIDPattern,
  frontendPhaseRangeLabel,
  frontendRowIDPattern,
} from "./frontend/phase-ids.mjs";
import {
  frontendRowsForAccountingTarget,
  frontendTargetHasClosureRows,
  rowsThroughSelectedActiveFrontendPhase,
  selectedFrontendRowAccountingScope,
  selectedFrontendRows,
  uniqueSorted,
  validateFrontendRowIDs,
} from "./frontend/row-scope.mjs";
import {
  loadTaskSurfaceManifest,
  targetEntryMap,
} from "../generated-artifacts/index.mjs";
import {
  testSliceDefaultCapacityProfile,
  resolveSchedulerResourceLimits,
  schedulerCapacityProfileLimits,
} from "../scheduler/scheduler-resource-policy.mjs";
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
import { normalizePhaseSliceSchedulerDAG } from "./phase-slice-planning/scheduler-dag.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");
export const frontendPhaseSlicePlanSchemaID = phaseSlicePlanSchemaID;

const browserStackResource = "browser_stack";

export class FrontendPhaseNotExecutableError extends Error {
  constructor(message) {
    super(message);
    this.name = "FrontendPhaseNotExecutableError";
    this.exitCode = 2;
  }
}

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function normalizeTarget(target) {
  if (target && typeof target === "object") {
    return target.target_name;
  }
  return target.startsWith("make ") ? target.slice("make ".length) : target;
}

function isBrowserTarget(target) {
  return target.startsWith("browser-e2e");
}

function validFrontendPhaseName(value) {
  return frontendPhaseIDPattern.test(String(value));
}

function parseSelectedRowIDs(value) {
  const rowIDs = parsePhaseRowIDs(value);
  for (const rowID of rowIDs) {
    if (!frontendRowIDPattern.test(rowID)) {
      throw new PhaseSliceSelectionError(
        `selected row ${rowID} does not belong to the frontend namespace`,
      );
    }
  }
  validateFrontendRowIDs(
    rowIDs,
    (rowID) => `invalid selected frontend row id ${rowID}`,
  );
  return rowIDs;
}

function claimStatusCounts(rows) {
  const counts = {
    implemented: 0,
    blocked: 0,
    not_applicable: 0,
    unspecified: 0,
  };
  for (const row of rows) {
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

function rowTargetEntries(rows, mode) {
  const entries = [];
  for (const row of rows) {
    if (row.claim_status === "blocked") {
      continue;
    }
    for (const rawTarget of row.targets) {
      const target = normalizeTarget(rawTarget);
      if (mode === "service_backed" && !isBrowserTarget(target)) {
        continue;
      }
      const frontendRowAccountingRequired =
        rawTarget && typeof rawTarget === "object"
          ? rawTarget.required_for_closure ||
            rawTarget.frontend_row_accounting_required
          : true;
      entries.push({ row, target, frontendRowAccountingRequired });
    }
  }
  return entries;
}

function targetOrder(target) {
  if (target === "frontend-typecheck") {
    return 10;
  }
  if (target === "frontend-unit") {
    return 20;
  }
  if (target === "frontend-import-boundary-check") {
    return 30;
  }
  if (target === "lint-biome") {
    return 40;
  }
  if (target.startsWith("browser-e2e")) {
    return 50;
  }
  return 60;
}

function rowGroups(entries) {
  const groups = new Map();
  for (const { row, target } of entries) {
    const key = [
      target,
      row.layer,
      row.evidence_class,
      row.claim_status,
    ].join("\u001f");
    if (!groups.has(key)) {
      groups.set(key, {
        target,
        layer: row.layer,
        evidence_class: row.evidence_class,
        claim_status: row.claim_status,
        ids: [],
      });
    }
    groups.get(key).ids.push(row.id);
  }
  return Array.from(groups.values())
    .map((group) => ({
      ...group,
      row_count: group.ids.length,
      ids: uniqueSorted(group.ids),
    }))
    .sort(
      (left, right) =>
        targetOrder(left.target) - targetOrder(right.target) ||
        compareStrings(left.target, right.target) ||
        compareStrings(left.layer, right.layer) ||
        compareStrings(left.evidence_class, right.evidence_class),
    );
}

function childTargets(entries, phase) {
  const byTarget = new Map();
  for (const { row, target, frontendRowAccountingRequired } of entries) {
    if (!byTarget.has(target)) {
      byTarget.set(target, { target, ids: [], accountingIDs: [] });
    }
    byTarget.get(target).ids.push(row.id);
    if (frontendRowAccountingRequired) {
      byTarget.get(target).accountingIDs.push(row.id);
    }
  }
  return Array.from(byTarget.values())
    .map((entry) => {
      const accountingIDs = uniqueSorted(entry.accountingIDs);
      return {
        target: entry.target,
        row_count: uniqueSorted(entry.ids).length,
        ids: uniqueSorted(entry.ids),
        ...(accountingIDs.length > 0
          ? {
              frontend_row_accounting_scope:
                selectedFrontendRowAccountingScope(phase, accountingIDs),
            }
          : {}),
      };
    })
    .sort(
      (left, right) =>
        targetOrder(left.target) - targetOrder(right.target) ||
        compareStrings(left.target, right.target),
    );
}

function targetWeight(rowCount) {
  return Math.max(1, rowCount) * 1000;
}

function serializeCompositeGateWorkUnits(workUnits, taskSurfaceManifest) {
  const compositeUnits = workUnits.filter((unit) => {
    const type = taskSurfaceManifest.make_recipes?.[unit.target]?.type;
    return type === "check_schedule" || type === "sequence";
  });
  if (compositeUnits.length === 0) {
    return compositeUnits;
  }

  const leafCompletionKeys = workUnits
    .filter((unit) => !compositeUnits.includes(unit))
    .flatMap((unit) => unit.completionKeys ?? [unit.id]);
  let precedingComposite = null;
  for (const unit of compositeUnits) {
    unit.needs = uniqueSorted([
      ...(unit.needs ?? []),
      ...leafCompletionKeys,
      ...(precedingComposite === null
        ? []
        : (precedingComposite.completionKeys ?? [precedingComposite.id])),
    ]);
    precedingComposite = unit;
  }
  return compositeUnits;
}

function scheduleCompositeGatesAfterServiceCompletion(plan, compositeUnits) {
  if (compositeUnits.length === 0) {
    return;
  }
  const serviceComplete = plan.workUnits.find(
    (unit) => unit.kind === "service_complete",
  );
  if (!serviceComplete) {
    return;
  }

  const serviceSessionKey = `service_session:${plan.target}`;
  const serviceCompleteKey = `service_complete:${plan.target}`;
  const compositeCompletionKeys = new Set(
    compositeUnits.flatMap((unit) => unit.completionKeys ?? [unit.id]),
  );
  serviceComplete.needs = (serviceComplete.needs ?? []).filter(
    (need) => !compositeCompletionKeys.has(need),
  );
  for (const unit of compositeUnits) {
    unit.needs = uniqueSorted([
      ...(unit.needs ?? []).filter((need) => need !== serviceSessionKey),
      serviceCompleteKey,
    ]);
    delete unit.serviceSession;
    delete unit.service_session;
    unit.isolatedRetainedRun = true;
  }
}

function resourceLimitsForWorkUnits(workUnits, label) {
  const profileLimits = schedulerCapacityProfileLimits(
    "test_slice",
    testSliceDefaultCapacityProfile,
    label,
  );
  const resolved = resolveSchedulerResourceLimits({
    scheduler: "test_slice",
    resourceLimits: profileLimits.limits,
    resourceLimitSources: profileLimits.sources,
    label,
    workUnits,
    pruneToClaims: true,
  });
  return resolved.resourceLimits;
}

function resourceClaimsForTarget(target) {
  if (isBrowserTarget(target)) {
    return new Map([
      ["postgres", 1],
      ["object_store", 1],
      ["process", 1],
      [browserStackResource, 1],
    ]);
  }
  return new Map([["process", 1]]);
}

function serviceRequirementsForTargets(targets) {
  if (!targets.some((target) => isBrowserTarget(target))) {
    return [];
  }
  return ["browser_stack", "object_store", "postgres"];
}

function knownTaskSurface(root) {
  const { manifest } = loadTaskSurfaceManifest(
    path.join(root, "tools", "task_surface_manifest.json"),
  );
  return { manifest, targets: targetEntryMap(manifest) };
}

function declaredTargetPrerequisites(manifest, target, knownTargets) {
  const found = new Set();
  const pending = [target];
  while (pending.length > 0) {
    const current = pending.pop();
    const recipe = manifest.make_recipes?.[current];
    for (const prerequisite of recipe?.prerequisites ?? []) {
      if (!knownTargets.has(prerequisite) || found.has(prerequisite)) {
        continue;
      }
      found.add(prerequisite);
      pending.push(prerequisite);
    }
  }
  return [...found];
}

function validateFrontendAccountingBoundaries({
  root,
  registry,
  taskSurface,
  children,
}) {
  for (const child of children) {
    if (child.frontend_row_accounting_scope !== undefined) {
      frontendRowsForAccountingTarget({
        root,
        registry,
        target: child.target,
        scope: child.frontend_row_accounting_scope,
      });
    }
    for (const prerequisite of declaredTargetPrerequisites(
      taskSurface.manifest,
      child.target,
      taskSurface.targets,
    )) {
      if (
        frontendTargetHasClosureRows({
          root,
          registry,
          target: prerequisite,
        })
      ) {
        throw new Error(
          `frontend phase slice evidence target ${child.target} has closure-bearing prerequisite ${prerequisite}; selected row accounting requires an explicit target boundary`,
        );
      }
    }
  }
}

export function buildFrontendPhaseSlicePlan(
  phase,
  { mode = "phase", root = repoRoot, rowIDs = "" } = {},
) {
  if (!validFrontendPhaseName(phase)) {
    throw new PhaseSliceSelectionError(
      `invalid frontend phase ${phase}; expected FE-P<N>`,
    );
  }
  if (!["phase", "service_backed"].includes(mode)) {
    throw new Error(`invalid frontend phase slice mode ${mode}`);
  }
  const registry = loadFrontendPhaseRegistry(root);
  const registryEntry = registry.phases.find((entry) => entry.phase_id === phase);
  if (!registryEntry) {
    throw new PhaseSliceSelectionError(
      `unknown frontend phase ${phase}; expected ${frontendPhaseRangeLabel(registry)}`,
    );
  }
  const selectedRowIDs = parseSelectedRowIDs(rowIDs);
  if (registryEntry.status !== "active") {
    throw new FrontendPhaseNotExecutableError(
      `planned/non-executable frontend phase ${phase}; frontend phase is ${registryEntry.status} and must be promoted to active before phase-slice execution`,
    );
  }
  const selectedRows =
    selectedRowIDs.length > 0
      ? selectedFrontendRows(root, registry, phase, selectedRowIDs)
      : rowsThroughSelectedActiveFrontendPhase(root, registry, phase);
  const entries = rowTargetEntries(selectedRows, mode);
  const children = childTargets(entries, phase);
  const targetNames = children.map((entry) => entry.target);
  const taskSurface = knownTaskSurface(root);
  const taskTargets = taskSurface.targets;
  for (const target of targetNames) {
    if (!taskTargets.has(target)) {
      throw new Error(
        `frontend phase slice target ${target} is not in task surface`,
      );
    }
  }
  validateFrontendAccountingBoundaries({
    root,
    registry,
    taskSurface,
    children,
  });
  const selectedExecutableRows = uniqueSorted(entries.map((entry) => entry.row.id)).map(
    (id) => selectedRows.find((row) => row.id === id),
  ).filter(Boolean);
  const claimCounts = claimStatusCounts(selectedExecutableRows);
  const workUnitModels = children.map((child, index) => ({
    id: child.target,
    label: child.target,
    kind: "make_target",
    type: "make_target",
    class: isBrowserTarget(child.target) ? "browser" : "frontend",
    target: child.target,
    aggregateTarget: child.target,
    group: child.target,
    needs: [],
    completionKeys: [child.target],
    failureKeys: [child.target],
    weightMs: targetWeight(child.row_count),
    make_prerequisite_policy: "skip",
    resourceClaims: resourceClaimsForTarget(child.target),
    ...(child.frontend_row_accounting_scope === undefined
      ? {}
      : {
          frontend_row_accounting_scope:
            child.frontend_row_accounting_scope,
        }),
    order: index,
  }));
  const compositeUnits = serializeCompositeGateWorkUnits(
    workUnitModels,
    taskSurface.manifest,
  );
  const target = mode === "service_backed" ? "service-backed-slice" : "phase-slice";
  const serviceRequirements = serviceRequirementsForTargets(targetNames);
  const schedulerPlan = {
    target,
    workUnits: workUnitModels,
    service_requirements: serviceRequirements,
    runtime_binaries: [],
    nextOrder: workUnitModels.length,
  };
  normalizePhaseSliceSchedulerDAG(schedulerPlan, root);
  scheduleCompositeGatesAfterServiceCompletion(schedulerPlan, compositeUnits);
  const resourceLimits = resourceLimitsForWorkUnits(
    schedulerPlan.workUnits,
    `${target} ${phase} frontend resource_limits`,
  );
  const workUnits = schedulerPlan.workUnits.map(serializePhaseSliceWorkUnit);
  const countedWorkUnits = schedulerPlan.workUnits.filter(
    (unit) => unit.countInTotal !== false,
  );
  const finalizerWorkUnits = schedulerPlan.workUnits.filter(
    (unit) => unit.countInTotal === false,
  );

  return validatePhaseSlicePlanContract({
    schema_id: frontendPhaseSlicePlanSchemaID,
    phase_namespace: "frontend",
    target,
    phase,
    selection: phaseSliceSelection({
      phaseNamespace: "frontend",
      mode,
      requestedRowIDs: selectedRowIDs,
      resolvedRowIDs: selectedExecutableRows.map((row) => row.id),
    }),
    mode,
    service_backed_only: mode === "service_backed",
    no_op: workUnits.length === 0,
    phase_claim_status: aggregateClaimStatus(claimCounts),
    claim_status_counts: claimCounts,
    row_groups: rowGroups(entries),
    service_requirements: serviceRequirements,
    child_targets: children,
    child_target_names: targetNames,
    resource_limits: resourceLimitObject(resourceLimits),
    work_units: workUnits,
    total_work_units: countedWorkUnits.length,
    finalizer_count: finalizerWorkUnits.length,
  });
}

export function printableFrontendPlan(plan) {
  return phaseSlicePlanOutput(plan);
}
