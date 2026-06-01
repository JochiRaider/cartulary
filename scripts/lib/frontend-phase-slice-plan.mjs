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
} from "./task-surface.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");
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
      entries.push({ row, target });
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

function childTargets(entries) {
  const byTarget = new Map();
  for (const { row, target } of entries) {
    if (!byTarget.has(target)) {
      byTarget.set(target, { target, ids: [] });
    }
    byTarget.get(target).ids.push(row.id);
  }
  return Array.from(byTarget.values())
    .map((entry) => ({
      target: entry.target,
      row_count: uniqueSorted(entry.ids).length,
      ids: uniqueSorted(entry.ids),
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
    resourceLimits.set("minio", 32);
    resourceLimits.set(browserStackResource, defaultBrowserStackLimit());
  }
  return resourceLimits;
}

function resourceClaimsForTarget(target) {
  if (isBrowserTarget(target)) {
    return new Map([
      ["postgres", 1],
      ["minio", 1],
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
  return ["browser_stack", "minio", "postgres"];
}

function knownTaskTargets(root) {
  const { manifest } = loadTaskSurfaceManifest(
    path.join(root, "tools", "task_surface_manifest.json"),
  );
  return targetEntryMap(manifest);
}

export function buildFrontendPhaseSlicePlan(
  phase,
  { mode = "phase", root = repoRoot } = {},
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
  if (registryEntry.status !== "active") {
    throw new FrontendPhaseNotExecutableError(
      `planned/non-executable frontend phase ${phase}; frontend phase is ${registryEntry.status} and must be promoted to active before phase-slice execution`,
    );
  }
  const { manifest } = loadFrontendPhaseMap(root, phase);
  const entries = rowTargetEntries(manifest.rows, mode);
  const children = childTargets(entries);
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
  const selectedRows = uniqueSorted(entries.map((entry) => entry.row.id)).map(
    (id) => manifest.rows.find((row) => row.id === id),
  ).filter(Boolean);
  const claimCounts = claimStatusCounts(selectedRows);
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
      order: index,
    }),
  );

  return {
    schema_id: frontendPhaseSlicePlanSchemaID,
    phase_namespace: "frontend",
    target: mode === "service_backed" ? "service-backed-slice" : "phase-slice",
    phase,
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
    })),
  };
}
