import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "./frontend-phase-manifest.mjs";
import {
  loadTaskSurfaceManifest,
  targetEntryMap,
} from "../generated-artifacts/task-surface.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");
export const frontendPhaseSlicePlanSchemaID = "cartulary.phase_slice_plan.v1";

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

function uniqueSorted(values) {
  return Array.from(new Set(values.filter(Boolean))).sort(compareStrings);
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
  return /^FE-P(?:0|[1-9]\d*)$/.test(value);
}

const frontendRowIDPattern =
  /^FE-(?:U|I|B|E|V|A11Y|S)-P(?:0|[1-9][0-9]*)-[0-9]{2}$/;

function parseSelectedRowIDs(value) {
  const rowIDs = uniqueSorted(
    (Array.isArray(value) ? value : String(value ?? "").split(","))
      .map((item) => String(item).trim())
      .filter(Boolean),
  );
  for (const rowID of rowIDs) {
    if (!frontendRowIDPattern.test(rowID)) {
      throw new Error(`invalid selected frontend row id ${rowID}`);
    }
  }
  return rowIDs;
}

function frontendPhaseNumber(value) {
  const match = String(value).match(/^FE-P([0-9]+)$/);
  return match ? Number.parseInt(match[1], 10) : Number.NaN;
}

function availableCPUCount() {
  if (typeof os.availableParallelism === "function") {
    return Math.max(1, os.availableParallelism());
  }
  return Math.max(1, os.cpus().length);
}

function defaultBrowserStackLimit() {
  const configured = Number.parseInt(
    process.env.CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT ?? "",
    10,
  );
  if (Number.isInteger(configured) && configured > 0) {
    return configured;
  }
  return availableCPUCount() >= 8 ? 2 : 1;
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

function selectedRowAccountingScope(phase, rowIDs) {
  return {
    mode: "selected_rows",
    invocation_kind: "frontend_phase_slice",
    phase_namespace: "frontend",
    phase,
    selection_policy: "frontend_rows_through_selected_phase",
    selected_row_ids: uniqueSorted(rowIDs),
  };
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
    .map((entry) => ({
      target: entry.target,
      row_count: uniqueSorted(entry.ids).length,
      ids: uniqueSorted(entry.ids),
      frontend_row_accounting_scope: selectedRowAccountingScope(
        phase,
        entry.accountingIDs,
      ),
    }))
    .sort(
      (left, right) =>
        targetOrder(left.target) - targetOrder(right.target) ||
        compareStrings(left.target, right.target),
    );
}

function targetWeight(rowCount) {
  return Math.max(1, rowCount) * 1000;
}

function resourceLimitObject(resourceLimits) {
  return Object.fromEntries(
    Array.from(resourceLimits.entries()).sort(([left], [right]) =>
      left.localeCompare(right),
    ),
  );
}

function resourceLimitsForTargets(targets) {
  const resourceLimits = new Map();
  if (targets.length > 0) {
    resourceLimits.set("process", 4);
  }
  if (targets.some((target) => isBrowserTarget(target))) {
    resourceLimits.set("postgres", 32);
    resourceLimits.set("object_store", 32);
    resourceLimits.set(browserStackResource, defaultBrowserStackLimit());
  }
  return resourceLimits;
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

function serializeWorkUnit(unit) {
  const { resourceClaims, weightMs: _weightMs, ...rest } = unit;
  return {
    ...rest,
    weight_ms: unit.weightMs,
    resource_claims: resourceLimitObject(resourceClaims ?? new Map()),
  };
}

function serviceRequirementsForTargets(targets) {
  if (!targets.some((target) => isBrowserTarget(target))) {
    return [];
  }
  return ["browser_stack", "object_store", "postgres"];
}

function knownTaskTargets(root) {
  const { manifest } = loadTaskSurfaceManifest(
    path.join(root, "tools", "task_surface_manifest.json"),
  );
  return targetEntryMap(manifest);
}

function rowsThroughSelectedActiveFrontendPhase(root, registry, selectedPhase) {
  const selectedOrder = frontendPhaseNumber(selectedPhase);
  const rows = [];
  for (const entry of registry.phases) {
    const order = frontendPhaseNumber(entry.phase_id);
    if (!Number.isFinite(order) || order > selectedOrder) {
      continue;
    }
    if (entry.status !== "active") {
      continue;
    }
    const { manifest } = loadFrontendPhaseMap(root, entry.phase_id);
    rows.push(...manifest.rows);
  }
  return rows;
}

function selectedFrontendRows(root, registry, selectedPhase, selectedRowIDs) {
  const selectedOrder = frontendPhaseNumber(selectedPhase);
  const selectedIDSet = new Set(selectedRowIDs);
  const found = new Map();
  for (const entry of registry.phases) {
    const order = frontendPhaseNumber(entry.phase_id);
    if (!Number.isFinite(order) || order > selectedOrder) {
      continue;
    }
    const { manifest } = loadFrontendPhaseMap(root, entry.phase_id);
    for (const row of manifest.rows) {
      if (!selectedIDSet.has(row.id)) {
        continue;
      }
      if (!["implemented", "stale"].includes(row.claim_status)) {
        throw new Error(
          `selected frontend row ${row.id} is ${row.claim_status} and is not executable`,
        );
      }
      found.set(row.id, row);
    }
  }
  const missing = selectedRowIDs.filter((rowID) => !found.has(rowID));
  if (missing.length > 0) {
    throw new Error(
      `selected frontend row id(s) not found through ${selectedPhase}: ${missing.join(",")}`,
    );
  }
  return selectedRowIDs.map((rowID) => found.get(rowID));
}

export function buildFrontendPhaseSlicePlan(
  phase,
  { mode = "phase", root = repoRoot, rowIDs = "" } = {},
) {
  if (!validFrontendPhaseName(phase)) {
    throw new Error(`invalid frontend phase ${phase}; expected FE-P<N>`);
  }
  if (!["phase", "service_backed"].includes(mode)) {
    throw new Error(`invalid frontend phase slice mode ${mode}`);
  }
  const registry = loadFrontendPhaseRegistry(root);
  const registryEntry = registry.phases.find((entry) => entry.phase_id === phase);
  if (!registryEntry) {
    throw new Error(`unknown frontend phase ${phase}; expected FE-P0 through FE-P11`);
  }
  const selectedRowIDs = parseSelectedRowIDs(rowIDs);
  if (registryEntry.status !== "active" && selectedRowIDs.length === 0) {
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
  const taskTargets = knownTaskTargets(root);
  for (const target of targetNames) {
    if (!taskTargets.has(target)) {
      throw new Error(
        `frontend phase slice target ${target} is not in task surface`,
      );
    }
  }
  const resourceLimits = resourceLimitsForTargets(targetNames);
  const selectedExecutableRows = uniqueSorted(entries.map((entry) => entry.row.id)).map(
    (id) => selectedRows.find((row) => row.id === id),
  ).filter(Boolean);
  const claimCounts = claimStatusCounts(selectedExecutableRows);
  const workUnits = children.map((child, index) =>
    serializeWorkUnit({
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
      resourceClaims: resourceClaimsForTarget(child.target),
      frontend_row_accounting_scope: child.frontend_row_accounting_scope,
      order: index,
    }),
  );

  return {
    schema_id: frontendPhaseSlicePlanSchemaID,
    phase_namespace: "frontend",
    target: mode === "service_backed" ? "service-backed-slice" : "phase-slice",
    phase,
    selected_row_ids: selectedRowIDs,
    mode,
    service_backed_only: mode === "service_backed",
    no_op: workUnits.length === 0,
    phase_claim_status: aggregateClaimStatus(claimCounts),
    claim_status_counts: claimCounts,
    row_groups: rowGroups(entries),
    service_requirements: serviceRequirementsForTargets(targetNames),
    child_targets: children,
    child_target_names: targetNames,
    resource_limits: resourceLimitObject(resourceLimits),
    work_units: workUnits,
    total_work_units: workUnits.length,
    finalizer_count: 0,
  };
}

export function printableFrontendPlan(plan) {
  return {
    schema_id: plan.schema_id,
    phase_namespace: plan.phase_namespace,
    target: plan.target,
    phase: plan.phase,
    selected_row_ids: plan.selected_row_ids,
    mode: plan.mode,
    no_op: plan.no_op,
    phase_claim_status: plan.phase_claim_status,
    claim_status_counts: plan.claim_status_counts,
    child_targets: plan.child_target_names,
    row_groups: plan.row_groups,
    service_requirements: plan.service_requirements,
    resource_limits: plan.resource_limits,
    work_units: plan.work_units.map((unit) => ({
      id: unit.id,
      label: unit.label,
      kind: unit.kind,
        target: unit.target,
        needs: unit.needs ?? [],
        resource_claims: unit.resource_claims,
        frontend_row_accounting_scope: unit.frontend_row_accounting_scope,
      })),
  };
}
