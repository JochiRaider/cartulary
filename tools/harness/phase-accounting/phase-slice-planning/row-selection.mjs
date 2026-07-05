import {
  compareExecutionDependencies,
  executionDependencyInfo,
} from "../../execution/execution-dependencies.mjs";
import { phaseGuidance, phaseSlice as guidancePhaseSlice } from "../../diagnostics/task-guidance.mjs";
import { activePhaseRegistryEntry, phaseRegistryEntry } from "../phase-registry.mjs";

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function uniqueSorted(values) {
  return Array.from(new Set(values.filter(Boolean))).sort(compareStrings);
}

export function phaseRows(phase, mode, root, taskSurfaceManifest = null) {
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

export function childTargetsForRows(rows, phase, mode, root, taskSurfaceManifest = null) {
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

export function executionDependenciesForTarget(rows, target) {
  return uniqueSorted(rows.filter((row) => row.target === target).map((row) => row.execution_dependency))
    .sort(compareExecutionDependencies);
}

export function serviceRequirementsForRows(rows) {
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

export function runtimeBinariesForRows(rows) {
  return uniqueSorted(rows.flatMap((row) => row.runtime_binaries ?? []));
}

export function disabledFrontendRowAccountingScope(phase) {
  return {
    mode: "disabled",
    invocation_kind: "base_phase_slice",
    phase_namespace: "base",
    phase,
    selection_policy: "base_phase_no_frontend_rows",
    selected_row_ids: [],
  };
}

export function claimStatusCounts(rows) {
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

export function aggregateClaimStatus(counts) {
  if (counts.blocked > 0 || counts.unspecified > 0) {
    return "incomplete";
  }
  if (counts.implemented > 0 || counts.not_applicable > 0) {
    return "complete";
  }
  return "not_applicable";
}

export function executableRows(rows) {
  return rows.filter((row) => row.claim_status !== "blocked");
}

export function rowGroups(rows) {
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
