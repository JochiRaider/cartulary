import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  collectTargetNames as collectBackendTargetNames,
  collectTargetPlanRows as collectBackendTargetPlanRows,
} from "../backend/backend-target-plan.mjs";
import {
  loadTaskSurfaceManifest,
  targetEntryMap,
} from "../generated-artifacts/index.mjs";
import { loadTestCatalog, targetForCatalogRow } from "../test-catalog/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function compareStrings(left, right) {
  return String(left ?? "").localeCompare(String(right ?? ""));
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

function projectionFields(taskTargets, target) {
  const projection = taskTargets.get(target)?.check_projection ?? null;
  return {
    check_projection_mode: projection?.mode ?? "",
    check_projection_full_target: projection?.full_target ?? "",
    check_projection_full_target_equivalent:
      projection?.full_target_equivalent ?? false,
  };
}

function defaultCheckWorkTargets(root) {
  const manifestPath = path.join(root, "tools", "scheduler_manifest.json");
  if (!existsSync(manifestPath)) return new Set();
  const targets = new Set();
  for (const schedule of readJSON(manifestPath).schedules ?? []) {
    if (schedule.target !== "check") continue;
    for (const unit of schedule.work_units ?? []) {
      if (unit.command?.type === "make_target" && unit.target) targets.add(unit.target);
    }
  }
  return targets;
}

function defaultBrowserSelectionIndex(root) {
  const manifestPath = path.join(root, "tools", "scheduler_manifest.json");
  const byRowID = new Map();
  if (!existsSync(manifestPath)) return byRowID;
  for (const schedule of readJSON(manifestPath).schedules ?? []) {
    if (
      schedule.target !== "check-service-backed" ||
      schedule.scheduler_kind !== "service_backed"
    ) continue;
    for (const unit of schedule.work_units ?? []) {
      if (unit.command?.type !== "browser_group") continue;
      const groupID = unit.command.group_id ?? "";
      const group = groupID.includes(":") ? groupID.slice(groupID.indexOf(":") + 1) : groupID;
      const browserGroup = unit.browser_group ?? {};
      const selectedRowIDs = uniqueSorted([
        ...parseCSV(unit.env?.CARTULARY_BROWSER_SELECTED_ROW_IDS ?? ""),
        ...(browserGroup.entry_ids ?? []),
      ]);
      const selection = {
        schedule: schedule.target,
        target: unit.target ?? "",
        browser_stage: unit.browser_stage ?? "",
        browser_group: browserGroup.name ?? group,
        browser_group_kind: browserGroup.kind ?? unit.kind ?? "browser_group",
        selected_row_ids: selectedRowIDs,
      };
      for (const rowID of selectedRowIDs) {
        const entries = byRowID.get(rowID) ?? [];
        entries.push(selection);
        byRowID.set(rowID, entries);
      }
    }
  }
  return byRowID;
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

function commandTargetIndex(taskTargets) {
  return new Map(
    [...taskTargets.values()].map((entry) => [entry.command_id, entry.name]),
  );
}

function profileByID(catalog, kind, profileID) {
  const profile = catalog.profiles.semantic[kind].find((entry) => entry.id === profileID);
  if (!profile) throw new Error(`unresolved ${kind} profile ${profileID}`);
  return profile;
}

function supportRow(row) {
  const family = row.family_id.split(".").at(-1);
  return family.startsWith("support_") ||
    (row.runner === "playwright" && row.selector.stage === "support");
}

function catalogPlanRow({ row, target, taskTargets, catalog, selection, directCheckTargets }) {
  const runtime = profileByID(catalog, "runtime_profiles", row.runtime_profile_id);
  const supportOnly = supportRow(row);
  const directSelection =
    !selection && row.default_check && directCheckTargets.has(target)
      ? {
          schedule: "check",
          target,
          browser_stage: "",
          browser_group: "",
          browser_group_kind: "",
          selected_row_ids: [row.row_id],
        }
      : null;
  const selected = selection ?? directSelection;
  const file = row.selector.file ?? "";
  const titles = [...(row.selector.titles ?? [])];
  return {
    source_family:
      row.runner === "playwright" ? "browser" : row.runner === "vitest" ? "frontend" : "command",
    target,
    service_backed: runtime.managed_service_ids.length > 0,
    runner_family: row.runner,
    id: row.row_id,
    owner_id: row.owner_id,
    family_id: row.family_id,
    section: row.family_id.split(".").at(-1),
    coverage: supportOnly ? "support" : "authoritative",
    execution_dependency: target.replaceAll("-", "_"),
    evidence_class: row.evidence_class,
    layer: target,
    default_check_required: row.default_check,
    default_check_kind: row.default_check ? "catalog_default" : "explicit_only",
    default_check_reason_code: row.default_check ? "catalog_selected" : "catalog_explicit_only",
    primary_evidence_owner: row.row_id,
    duplicate_of: null,
    evidence_delta: "Catalog-owned exact selector evidence.",
    warm_local_cost_class: runtime.managed_service_ids.length > 0 ? "service_backed" : "low",
    execution_family: row.family_id,
    execution_label: row.family_id,
    packages: [],
    support_only: supportOnly,
    support_selector: null,
    raw_selector: null,
    file,
    package: "",
    symbols: [],
    titles,
    scenario_titles: titles,
    command_id: row.selector.command_id ?? "",
    runtime_profile_id: row.runtime_profile_id,
    resource_profile_id: row.resource_profile_id,
    fixture_profile_id: row.fixture_profile_id,
    evidence_layer: target,
    ...projectionFields(taskTargets, target),
    ...selectionFields(selected),
  };
}

function compareRows(left, right) {
  return (
    compareStrings(left.source_family, right.source_family) ||
    compareStrings(left.target, right.target) ||
    compareStrings(left.owner_id, right.owner_id) ||
    compareStrings(left.family_id, right.family_id) ||
    compareStrings(left.id, right.id)
  );
}

export function collectTargetNames(root = repoRoot) {
  const names = new Set(collectBackendTargetNames(root));
  try {
    for (const target of loadTaskSurfaceTargets(root).keys()) names.add(target);
  } catch {
    // Synthetic backend-only roots may intentionally omit the task surface.
  }
  return [...names].sort(compareStrings);
}

export function collectHarnessTargetPlanRows(root = repoRoot) {
  const taskTargets = loadTaskSurfaceTargets(root);
  const catalog = loadTestCatalog(root);
  const commandTargets = commandTargetIndex(taskTargets);
  const selectionIndex = defaultBrowserSelectionIndex(root);
  const directCheckTargets = defaultCheckWorkTargets(root);
  const nonGoRows = catalog.rows
    .filter((row) => row.runner !== "go")
    .map((row) => {
      const target = targetForCatalogRow(row, { commandTargetByID: commandTargets });
      const selection = row.runner === "playwright"
        ? (selectionIndex.get(row.row_id) ?? []).find((entry) => entry.target === target) ?? null
        : null;
      return catalogPlanRow({
        row,
        target,
        taskTargets,
        catalog,
        selection,
        directCheckTargets,
      });
    });
  return [
    ...collectBackendTargetPlanRows(root).map(({ manifest_phase: _retiredPhase, ...row }) => ({
      source_family: "backend",
      ...row,
      ...projectionFields(taskTargets, row.target),
      ...selectionFields(null),
    })),
    ...nonGoRows,
  ].sort(compareRows);
}
