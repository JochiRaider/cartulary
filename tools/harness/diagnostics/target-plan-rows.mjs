import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  collectTargetNames as collectBackendTargetNames,
  collectTargetPlanRows as collectBackendTargetPlanRows,
} from "../backend/backend-target-plan.mjs";
import { targetForExecutionDependency } from "../execution/execution-dependencies.mjs";
import {
  loadTaskSurfaceManifest,
  targetEntryMap,
} from "../generated-artifacts/index.mjs";
import {
  collectEntries,
  entryIsExecutable,
  frontendPhaseBaseJoin,
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
  loadManifest,
  phaseManifestNames,
  playwrightEntryTitles,
} from "../phase-accounting/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function compareStrings(left, right) {
  return String(left ?? "").localeCompare(String(right ?? ""), undefined, {
    numeric: true,
  });
}

function uniqueSorted(values) {
  return [...new Set(values.filter(Boolean))].sort(compareStrings);
}

function parseCSV(value) {
  return uniqueSorted(
    String(value ?? "")
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean),
  );
}

function loadTaskSurfaceTargets(root) {
  const manifestPath = path.join(root, "tools", "task_surface_manifest.json");
  const { manifest } = loadTaskSurfaceManifest(manifestPath);
  return targetEntryMap(manifest);
}

function checkProjectionForTarget(taskTargets, target) {
  return taskTargets.get(target)?.check_projection ?? null;
}

function projectionFields(taskTargets, target) {
  const projection = checkProjectionForTarget(taskTargets, target);
  return {
    check_projection_mode: projection?.mode ?? "",
    check_projection_full_target: projection?.full_target ?? "",
    check_projection_full_target_equivalent:
      projection?.full_target_equivalent ?? false,
  };
}

function defaultCheckMetadata(row) {
  return {
    default_check_required: row.default_check_required === true,
    default_check_kind: row.default_check_kind ?? "",
    default_check_reason_code: row.default_check_reason_code ?? "",
    ...(row.default_check_reason
      ? { default_check_reason: row.default_check_reason }
      : {}),
    primary_evidence_owner: row.primary_evidence_owner ?? "",
    duplicate_of: row.duplicate_of ?? null,
    evidence_delta: row.evidence_delta ?? "",
    warm_local_cost_class: row.warm_local_cost_class ?? "",
  };
}

function coverageForFrontendEvidenceClass(evidenceClass) {
  return evidenceClass === "product_conformance" ? "authoritative" : "support";
}

function defaultCheckWorkTargets(root) {
  const manifestPath = path.join(root, "tools", "scheduler_manifest.json");
  if (!existsSync(manifestPath)) {
    return new Set();
  }
  const manifest = readJSON(manifestPath);
  const targets = new Set();
  for (const schedule of manifest.schedules ?? []) {
    if (schedule.target !== "check") {
      continue;
    }
    for (const unit of schedule.work_units ?? []) {
      if (unit.command?.type !== "make_target") {
        continue;
      }
      if (unit.target) {
        targets.add(unit.target);
      }
    }
  }
  return targets;
}

function defaultBrowserSelectionIndex(root) {
  const manifestPath = path.join(root, "tools", "scheduler_manifest.json");
  const byRowID = new Map();
  if (!existsSync(manifestPath)) {
    return byRowID;
  }

  const add = (rowID, selection) => {
    if (!rowID) {
      return;
    }
    const current = byRowID.get(rowID) ?? [];
    current.push(selection);
    byRowID.set(rowID, current);
  };

  const manifest = readJSON(manifestPath);
  for (const schedule of manifest.schedules ?? []) {
    if (
      schedule.target !== "check-service-backed" ||
      schedule.scheduler_kind !== "service_backed"
    ) {
      continue;
    }
    for (const unit of schedule.work_units ?? []) {
      if (unit.command?.type !== "browser_group") {
        continue;
      }
      const groupID = unit.command.group_id ?? "";
      const group = groupID.includes(":")
        ? groupID.slice(groupID.indexOf(":") + 1)
        : groupID;
      const browserGroup = unit.browser_group ?? {};
      const selectedRowIDs = uniqueSorted([
        ...parseCSV(unit.env?.CARTULARY_BROWSER_SELECTED_ROW_IDS ?? ""),
        ...(Array.isArray(browserGroup.entry_ids) ? browserGroup.entry_ids : []),
      ]);
      const selection = {
        schedule: schedule.target,
        target: unit.target ?? "",
        browser_stage: unit.browser_stage ?? "",
        browser_group: browserGroup.name ?? group,
        browser_group_kind: browserGroup.kind ?? unit.kind ?? "browser_group",
        selected_row_ids: selectedRowIDs,
      };
      for (const rowID of selection.selected_row_ids) {
        add(rowID, selection);
      }
    }
  }

  return byRowID;
}

function firstSelection(selectionIndex, rowID) {
  return selectionIndex.get(rowID)?.[0] ?? null;
}

function firstSelectionForTarget(selectionIndex, rowID, target) {
  return (
    (selectionIndex.get(rowID) ?? []).find(
      (selection) => selection.target === target,
    ) ?? null
  );
}

function selectionFields(selection) {
  return {
    scheduled_by_default_check: selection !== null,
    default_check_schedule: selection?.schedule ?? "",
    browser_stage: selection?.browser_stage ?? "",
    browser_group: selection?.browser_group ?? "",
    browser_group_kind: selection?.browser_group_kind ?? "",
    default_check_selected_row_ids: selection?.selected_row_ids ?? [],
  };
}

function collectBrowserRows(root, taskTargets, selectionIndex) {
  const rows = [];
  for (const phase of phaseManifestNames(root)) {
    const { manifest } = loadManifest(root, phase);
    for (const entry of collectEntries(manifest)) {
      if (
        entry.runner !== "playwright" ||
        !entryIsExecutable(entry) ||
        !entry.execution_dependency
      ) {
        continue;
      }
      const target = targetForExecutionDependency(
        entry.execution_dependency,
        `manifest entry ${entry.id} execution_dependency`,
      );
      const selection = firstSelection(selectionIndex, entry.id);
      rows.push({
        source_family: "browser",
        target,
        service_backed: true,
        runner_family: "playwright",
        id: entry.id,
        manifest_phase: phase,
        section: entry.section ?? "",
        coverage: entry.coverage ?? "",
        execution_dependency: entry.execution_dependency,
        evidence_class: entry.evidence_class ?? "",
        layer: entry.layer ?? "",
        ...defaultCheckMetadata(entry),
        execution_family: entry.execution_family ?? "",
        execution_label: entry.execution_label ?? "",
        packages: [],
        support_only: false,
        support_selector: null,
        raw_selector: null,
        file: entry.file ?? "",
        package: "",
        symbols: [],
        titles: playwrightEntryTitles(entry),
        shard_isolation: entry.shard_isolation === true,
        evidence_layer: entry.evidence_layer ?? "",
        ...projectionFields(taskTargets, target),
        ...selectionFields(selection),
      });
    }
  }
  return rows;
}

function collectFrontendRows(root, taskTargets, selectionIndex, directCheckTargets) {
  let registry;
  try {
    registry = loadFrontendPhaseRegistry(root);
  } catch {
    return [];
  }

  const rows = [];
  for (const phase of registry.phases ?? []) {
    if (phase.status !== "active") {
      continue;
    }
    const basePhase = frontendPhaseBaseJoin(phase);
    const { manifest } = loadFrontendPhaseMap(root, phase.phase_id);
    for (const row of manifest.rows ?? []) {
      if (row.claim_status !== "implemented") {
        continue;
      }
      for (const targetRef of row.targets ?? []) {
        const target = targetRef.target_name ?? "";
        const browserSelection = target.startsWith("browser-e2e")
          ? firstSelectionForTarget(selectionIndex, row.id, target)
          : null;
        const directSelection =
          !browserSelection && directCheckTargets.has(target)
            ? {
                schedule: "check",
                target,
                browser_stage: "",
                browser_group: "",
                browser_group_kind: "",
                selected_row_ids: [row.id],
              }
            : null;
        const selection = browserSelection ?? directSelection;
        rows.push({
          source_family: "frontend",
          target,
          service_backed: target.startsWith("browser-e2e"),
          runner_family: target.startsWith("browser-e2e")
            ? "playwright"
            : target === "frontend-unit"
              ? "vitest"
              : "frontend",
          id: row.id,
          manifest_phase: basePhase,
          phase_id: phase.phase_id,
          phase_namespace: "frontend",
          base_phase_join: basePhase,
          section: row.layer ?? "",
          coverage: coverageForFrontendEvidenceClass(row.evidence_class),
          execution_dependency: target.startsWith("browser-e2e")
            ? `frontend:${target}`
            : "",
          evidence_class: row.evidence_class ?? "",
          layer: row.layer ?? "",
          claim_status: row.claim_status,
          target_evidence_role: targetRef.evidence_role ?? "",
          target_required_for_closure:
            targetRef.required_for_closure === true,
          frontend_row_accounting_required:
            targetRef.frontend_row_accounting_required === true,
          ...defaultCheckMetadata(row),
          execution_family: "frontend_phase_map",
          execution_label: row.claim?.statement ?? row.id,
          packages: [],
          support_only: row.evidence_class !== "product_conformance",
          support_selector: null,
          raw_selector: null,
          file: "",
          package: "",
          symbols: [],
          scenario_titles: [...(row.scenario_titles ?? [])],
          evidence_layer: row.layer ?? "",
          ...projectionFields(taskTargets, target),
          ...selectionFields(selection),
        });
      }
    }
  }
  return rows;
}

function compareRows(left, right) {
  return (
    compareStrings(left.source_family, right.source_family) ||
    compareStrings(left.target, right.target) ||
    compareStrings(left.manifest_phase ?? left.phase_id, right.manifest_phase ?? right.phase_id) ||
    compareStrings(left.browser_stage, right.browser_stage) ||
    compareStrings(left.browser_group, right.browser_group) ||
    compareStrings(left.id, right.id)
  );
}

export function collectTargetNames(root = repoRoot) {
  const names = new Set(collectBackendTargetNames(root));
  try {
    for (const target of loadTaskSurfaceTargets(root).keys()) {
      names.add(target);
    }
  } catch {
    // Keep backend names available for synthetic test roots without task surface data.
  }
  return [...names].sort(compareStrings);
}

export function collectHarnessTargetPlanRows(root = repoRoot) {
  const taskTargets = loadTaskSurfaceTargets(root);
  const selectionIndex = defaultBrowserSelectionIndex(root);
  const directCheckTargets = defaultCheckWorkTargets(root);
  return [
    ...collectBackendTargetPlanRows(root).map((row) => ({
      source_family: "backend",
      ...row,
    })),
    ...collectBrowserRows(root, taskTargets, selectionIndex),
    ...collectFrontendRows(root, taskTargets, selectionIndex, directCheckTargets),
  ].sort(compareRows);
}
