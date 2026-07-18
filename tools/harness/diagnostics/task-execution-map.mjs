import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { newestTargetArtifact } from "../contract/index.mjs";
import { normalizeBrowserBatchStages } from "../browser/browser-batch-manifest.mjs";
import {
  compareExecutionDependencies,
  executionDependencyInfo,
} from "../execution/execution-dependencies.mjs";
import {
  loadExecutionTopology,
  renderBrowserBatchManifest,
} from "../generated-artifacts/execution-topology.mjs";
import { findTargetDescriptor } from "../backend/backend-target-plan.mjs";
import { loadTestCatalog, targetForCatalogRow } from "../test-catalog/index.mjs";
import {
  makeRecipeEntries,
  targetEntryMap,
} from "../generated-artifacts/task-surface/model.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const taskExecutionMapSchemaID = "cartulary.task_execution_map.v1";
const executionOwnerRowsCache = new Map();
const taskSurfaceCache = new Map();
const schedulerManifestCache = new Map();
const browserStagesCache = new Map();

const serviceRequirementDisplayNames = new Map([
  ["postgres", "Postgres"],
  ["object_store", "object store"],
  ["browser_stack", "browser stack"],
  ["vite", "Vite"],
]);

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function cacheKeyForRoot(root) {
  return path.resolve(root);
}

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function uniqueSorted(values) {
  return Array.from(new Set(values.filter((value) => value !== ""))).sort(compareStrings);
}

export function collectExecutionOwnerRows(root = repoRoot) {
  const cacheKey = cacheKeyForRoot(root);
  const cached = executionOwnerRowsCache.get(cacheKey);
  if (cached) {
    return cached;
  }
  const taskSurface = loadTaskSurface(root).manifest;
  const commandTargetByID = new Map(
    taskSurface.targets.map((entry) => [entry.command_id, entry.name]),
  );
  const rows = loadTestCatalog(root).rows
    .filter((entry) => entry.status === "active")
    .map((entry) => {
      const selector = entry.selector ?? {};
      const target = targetForCatalogRow(entry, { commandTargetByID });
      return {
        id: entry.row_id,
        owner: entry.owner_id,
        section: entry.family_id,
        coverage: "authoritative",
        claim_status: entry.status,
        runner: entry.runner,
        execution_dependency: target.replaceAll("-", "_"),
        evidence_class: entry.evidence_class,
        layer: entry.family_id,
        default_check_required: entry.default_check === true,
        runtime_binaries: [],
        target,
        file: selector.file ?? "",
        package: selector.package ?? "",
        title: [...(selector.tests ?? []), ...(selector.titles ?? [])].join(" | "),
        manifest_path: `tools/test_families/${entry.owner_id}.json`,
      };
    });
  const result = rows.sort((left, right) => (
    compareStrings(left.owner, right.owner) ||
    compareStrings(left.target, right.target) ||
    compareStrings(left.section, right.section) ||
    compareStrings(left.id, right.id)
  ));
  executionOwnerRowsCache.set(cacheKey, result);
  return result;
}

export function sectionCounts(rows) {
  const counts = {
    authoritative: 0,
    supplemental: 0,
    support: 0,
    raw: 0,
    total: rows.length,
  };
  for (const row of rows) {
    if (row.coverage === "authoritative") {
      counts.authoritative += 1;
    } else if (row.coverage === "supplemental") {
      counts.supplemental += 1;
    } else if (row.coverage === "support") {
      counts.support += 1;
    } else if (row.coverage === "raw") {
      counts.raw += 1;
    }
  }
  return counts;
}

export function summarizeExecutionRows(rows) {
  const counts = sectionCounts(rows);
  return {
    ...counts,
    owners: uniqueSorted(rows.map((row) => row.owner)),
    execution_dependencies: uniqueSorted(rows.map((row) => row.execution_dependency))
      .sort(compareExecutionDependencies),
    runners: uniqueSorted(rows.map((row) => row.runner)),
    targets: uniqueSorted(rows.map((row) => row.target)),
  };
}

function loadTaskSurface(root = repoRoot) {
  const manifestFile =
    process.env.CARTULARY_TASK_SURFACE_MANIFEST ?? path.join(root, "tools", "task_surface_manifest.json");
  const cacheKey = path.resolve(manifestFile);
  const cached = taskSurfaceCache.get(cacheKey);
  if (cached) {
    return cached;
  }
  const manifest = readJSON(manifestFile);
  const result = {
    manifest,
    targets: targetEntryMap(manifest),
  };
  taskSurfaceCache.set(cacheKey, result);
  return result;
}

function addServiceRequirement(requirements, value) {
  requirements.add(serviceRequirementDisplayNames.get(value) ?? value);
}

function addServiceRequirements(requirements, values = []) {
  for (const value of values) {
    addServiceRequirement(requirements, value);
  }
}

function serviceBackedScheduleForTarget(target, root = repoRoot) {
  const manifestPath = path.join(root, "tools", "scheduler_manifest.json");
  if (!existsSync(manifestPath)) {
    return null;
  }
  const manifest = loadSchedulerManifest(manifestPath);
  return (manifest.schedules ?? []).find(
    (schedule) => schedule?.target === target && schedule?.scheduler_kind === "service_backed",
  ) ?? null;
}

function loadSchedulerManifest(manifestPath) {
  const cacheKey = path.resolve(manifestPath);
  const cached = schedulerManifestCache.get(cacheKey);
  if (cached) {
    return cached;
  }
  const manifest = readJSON(manifestPath);
  schedulerManifestCache.set(cacheKey, manifest);
  return manifest;
}

function addServiceBackedScheduleRequirements(requirements, schedule) {
  if (!schedule) {
    return;
  }
  addServiceRequirement(requirements, "postgres");
  addServiceRequirement(requirements, "object_store");
  const sources = Array.isArray(schedule.work_unit_sources) ? schedule.work_unit_sources : [];
  if (
    sources.some((source) => source?.class === "browser" || source?.browser_stage) ||
    (schedule.work_units ?? []).some((unit) => unit?.class === "browser" || unit?.browser_stage) ||
    Object.hasOwn(schedule.resource_limits ?? {}, "browser_stack")
  ) {
    addServiceRequirement(requirements, "browser_stack");
  }
}

export function serviceRequirementsForTarget(target, targetRows = [], declaredRequirements = [], root = repoRoot) {
  const requirements = new Set();
  addServiceRequirements(requirements, declaredRequirements);
  addServiceBackedScheduleRequirements(requirements, serviceBackedScheduleForTarget(target, root));
  const goDescriptor = findTargetDescriptor(target, root);
  if (goDescriptor?.serviceBacked) {
    addServiceRequirement(requirements, "postgres");
    addServiceRequirement(requirements, "object_store");
  }
  if (target.startsWith("browser-e2e")) {
    addServiceRequirement(requirements, "postgres");
    addServiceRequirement(requirements, "object_store");
    addServiceRequirement(requirements, "browser_stack");
  }
  if (target === "db-up" || target === "services-up" || target === "object-store-init") {
    addServiceRequirement(requirements, "postgres");
    addServiceRequirement(requirements, "object_store");
  }
  if (target === "dev") {
    addServiceRequirement(requirements, "postgres");
    addServiceRequirement(requirements, "object_store");
    addServiceRequirement(requirements, "vite");
  }
  if (["test", "test-fast", "check", "ci", "release-check"].includes(target)) {
    addServiceRequirement(requirements, "postgres");
    addServiceRequirement(requirements, "object_store");
  }
  if (targetRows.some((row) => row.target.startsWith("browser-e2e"))) {
    addServiceRequirement(requirements, "browser_stack");
  }
  return Array.from(requirements);
}

function ownerRowsForTarget(target, rows) {
  if (target === "test-fast") {
    return rows.filter((row) => row.coverage !== "raw" && !row.target.startsWith("browser-e2e"));
  }
  if (["test", "check", "ci", "release-check"].includes(target)) {
    return rows.filter((row) => row.coverage !== "raw");
  }
  if (target === "browser-e2e") {
    return rows.filter((row) =>
      ["browser-e2e-stateful", "browser-e2e-measurement", "browser-e2e-visual"].includes(row.target),
    );
  }
  return rows.filter((row) => row.target === target);
}

function browserStagesByTarget(root = repoRoot) {
  const cacheKey = cacheKeyForRoot(root);
  const cached = browserStagesCache.get(cacheKey);
  if (cached) {
    return cached;
  }
  const result = new Map();
  const manifestPath = path.join(root, "tools", "execution_topology_manifest.json");
  if (!existsSync(manifestPath)) {
    browserStagesCache.set(cacheKey, result);
    return result;
  }
  const topology = loadExecutionTopology({ manifestPath });
  const stages = normalizeBrowserBatchStages(renderBrowserBatchManifest(topology));
  for (const stage of stages.values()) {
    const stageEntry = {
      stage: stage.name,
      target: stage.target,
      groups: stage.groups.map((group) => ({
        name: group.name,
        target: group.target,
        kind: group.kind,
        coverage: group.coverage,
        execution_dependency: group.executionDependency ?? group.execution_dependency ?? "",
      })),
    };
    const stageValues = result.get(stage.target) ?? [];
    stageValues.push(stageEntry);
    result.set(stage.target, stageValues);
    for (const group of stage.groups) {
      const groupValues = result.get(group.target) ?? [];
      groupValues.push(stageEntry);
      result.set(group.target, groupValues);
    }
  }
  browserStagesCache.set(cacheKey, result);
  return result;
}

function directRecipeSummary(target, manifest) {
  const recipes = new Map(makeRecipeEntries(manifest).map((recipe) => [recipe.target, recipe]));
  const recipe = recipes.get(target);
  if (!recipe) {
    return [];
  }
  if (recipe.type === "sequence") {
    return [{
      kind: "make_sequence",
      id: `make:${recipe.sequence}`,
      label: `Make sequence ${recipe.sequence}`,
      target,
      scheduler: "",
      detail: "runs child targets in sequence",
    }];
  }
  if (recipe.type === "check_schedule") {
    return [{
      kind: "check_scheduler",
      id: `check:${recipe.target}`,
      label: `${recipe.target} check scheduler`,
      target,
      scheduler: recipe.target,
      detail: recipe.schedule_manifest ?? "",
    }];
  }
  if (recipe.type === "service_backed_schedule") {
    return [{
      kind: "service_backed_scheduler",
      id: `service-backed:${target}`,
      label: `${target} service-backed scheduler`,
      target,
      scheduler: target,
      detail: recipe.schedule_manifest ?? "",
    }];
  }
  if (recipe.type === "browser_batch") {
    return [{
      kind: "browser_batch",
      id: `browser:${recipe.stage}`,
      label: `${recipe.stage} browser batch`,
      target,
      scheduler: recipe.stage ?? "",
      stage: recipe.stage ?? "",
      detail: "browser batch stage",
    }];
  }
  return [{
    kind: "make_target",
    id: `make:${target}`,
    label: `Make ${recipe.type}`,
    target,
    scheduler: "",
    detail: "direct Make target recipe",
  }];
}

function checkScheduleSummary(target, root = repoRoot) {
  const result = [];
  const manifestPath = path.join(root, "tools", "scheduler_manifest.json");
  if (!existsSync(manifestPath)) {
    return result;
  }
  const manifest = loadSchedulerManifest(manifestPath);
  for (const schedule of (manifest.schedules ?? []).filter((entry) => entry?.scheduler_kind === "check")) {
    for (const unit of schedule.work_units ?? []) {
      if (unit.target !== target) {
        continue;
      }
      result.push({
        kind: "check_work_unit",
        id: `${schedule.target}:${unit.target}`,
        label: `${schedule.target} work unit ${unit.target}`,
        target,
        scheduler: schedule.target,
        detail: unit.nested_scheduler
          ? `nested ${unit.nested_scheduler.type} scheduler ${unit.nested_scheduler.target}`
          : "check scheduler work unit",
        nested_scheduler: unit.nested_scheduler ?? null,
        resource_claims: unit.resource_claims ?? {},
      });
    }
  }
  return result;
}

function serviceBackedScheduleSummary(target, root = repoRoot) {
  const result = [];
  const manifestPath = path.join(root, "tools", "scheduler_manifest.json");
  if (!existsSync(manifestPath)) {
    return result;
  }
  const manifest = loadSchedulerManifest(manifestPath);
  for (const schedule of (manifest.schedules ?? []).filter((entry) => entry?.scheduler_kind === "service_backed")) {
    const seen = new Set();
    for (const unit of schedule.work_units ?? []) {
      const aggregateTarget = unit.aggregate_target ?? unit.target;
      if (aggregateTarget !== target || seen.has(aggregateTarget)) {
        continue;
      }
      seen.add(aggregateTarget);
      result.push({
        kind: "service_backed_work_unit",
        id: `${schedule.target}:${aggregateTarget}`,
        label: `${schedule.target} normalized work ${aggregateTarget}`,
        target,
        scheduler: schedule.target,
        source_type: unit.kind ?? "",
        source_class: unit.class ?? "",
        stage: unit.browser_stage ?? "",
        detail: unit.browser_stage ? `browser stage ${unit.browser_stage}` : `${unit.kind} work unit`,
        resource_claims: unit.resource_claims ?? {},
      });
    }
  }
  return result;
}

function browserBatchSummary(target, root = repoRoot) {
  const result = [];
  for (const stage of browserStagesByTarget(root).get(target) ?? []) {
    result.push({
      kind: "browser_batch_stage",
      id: `browser:${stage.stage}:${target}`,
      label: `${stage.stage} browser batch`,
      target,
      scheduler: stage.stage,
      stage: stage.stage,
      detail: stage.groups
        .map((group) => `${group.name}:${group.kind}`)
        .join(","),
      groups: stage.groups,
    });
  }
  return result;
}

function ownerSchedulerSummary(target) {
  if (target !== "owner-slice" && target !== "service-backed-slice") {
    return [];
  }
  return [{
    kind: "owner_scheduler",
    id: `owner-scheduler:${target}`,
    label: "owner scheduler",
    target,
    scheduler: "owner-scheduler",
    detail: target === "service-backed-slice"
      ? "selected owner service-backed manifest-row slice"
      : "selected owner manifest-row slice",
  }];
}

function targetWorkUnitSummary(target, manifest, root = repoRoot) {
  const summaries = [
    ...ownerSchedulerSummary(target),
    ...directRecipeSummary(target, manifest),
    ...checkScheduleSummary(target, root),
    ...serviceBackedScheduleSummary(target, root),
    ...browserBatchSummary(target, root),
  ];
  if (summaries.length === 0) {
    summaries.push({
      kind: "direct_make_target",
      id: `make:${target}`,
      label: "direct Make target",
      target,
      scheduler: "",
      detail: "no generated scheduler ownership",
    });
  }
  const seen = new Set();
  return summaries.filter((summary) => {
    const key = `${summary.kind}\u001f${summary.id}\u001f${summary.label}`;
    if (seen.has(key)) {
      return false;
    }
    seen.add(key);
    return true;
  }).sort((left, right) => (
    compareStrings(left.kind, right.kind) ||
    compareStrings(left.id, right.id) ||
    compareStrings(left.label, right.label)
  ));
}

function targetNode(target, {
  rows = null,
  root = repoRoot,
  coverageOverride = null,
  includeArtifacts = true,
} = {}) {
  const { manifest, targets } = loadTaskSurface(root);
  const targetRows = rows ?? ownerRowsForTarget(target, collectExecutionOwnerRows(root));
  const entry = targets.get(target);
  const coverage = coverageOverride ?? summarizeExecutionRows(targetRows);
  const artifacts = includeArtifacts
    ? newestTargetArtifact(target, { root })
    : { latest: null, expected: [], candidates: [] };
  return {
    kind: "target",
    id: target,
    label: target,
    target_class: entry?.target_class ?? null,
    coverage,
    services: serviceRequirementsForTarget(target, targetRows, entry?.service_requirements ?? [], root),
    execution_dependencies: coverage.execution_dependencies ?? [],
    work_unit_summary: targetWorkUnitSummary(target, manifest, root),
    artifacts: {
      latest: artifacts.latest?.path ?? "none",
      expected: artifacts.expected ?? [],
      discovered: artifacts.candidates ?? [],
    },
    children: [],
  };
}

function targetNodeForRows(target, rows, { root = repoRoot } = {}) {
  return targetNode(target, {
    rows,
    root,
    coverageOverride: summarizeExecutionRows(rows),
  });
}

function sectionCoverage(children) {
  const summaries = children.map((child) => child.coverage).filter(Boolean);
  const counts = {
    authoritative: 0,
    supplemental: 0,
    support: 0,
    raw: 0,
    total: 0,
  };
  for (const summary of summaries) {
    counts.authoritative += summary.authoritative ?? 0;
    counts.supplemental += summary.supplemental ?? 0;
    counts.support += summary.support ?? 0;
    counts.raw += summary.raw ?? 0;
    counts.total += summary.total ?? 0;
  }
  return {
    ...counts,
    owners: uniqueSorted(summaries.flatMap((summary) => summary.owners ?? [])),
    execution_dependencies: uniqueSorted(
      summaries.flatMap((summary) => summary.execution_dependencies ?? []),
    ).sort(compareExecutionDependencies),
    runners: uniqueSorted(summaries.flatMap((summary) => summary.runners ?? [])),
    targets: uniqueSorted(summaries.flatMap((summary) => summary.targets ?? [])),
  };
}

function makeSectionNode(id, label, children) {
  return {
    kind: "section",
    id,
    label,
    target_class: id,
    coverage: sectionCoverage(children),
    services: uniqueSorted(children.flatMap((child) => child.services ?? [])),
    execution_dependencies: uniqueSorted(children.flatMap((child) => child.execution_dependencies ?? []))
      .sort(compareExecutionDependencies),
    work_unit_summary: [],
    artifacts: {},
    children,
  };
}

function compareOwnerTargetNodes(left, right) {
  const leftDependency = left.execution_dependencies[0] ?? "";
  const rightDependency = right.execution_dependencies[0] ?? "";
  return compareExecutionDependencies(leftDependency, rightDependency) || compareStrings(left.id, right.id);
}

export function targetExecutionMap(target, { root = repoRoot, includeArtifacts = true } = {}) {
  return {
    schema_id: taskExecutionMapSchemaID,
    ...targetNode(target, { root, includeArtifacts }),
  };
}

export function executionSummary(executionMap) {
  const units = executionMap?.work_unit_summary ?? [];
  if (executionMap?.kind === "owner") {
    const sections = (executionMap.children ?? []).map((section) => {
      const count = section.children?.length ?? 0;
      const noun = count === 1 ? "target" : "targets";
      return `${section.label}: ${count} ${noun}`;
    });
    return sections.join(", ") || "none";
  }
  if (units.length === 0) {
    const childUnits = (executionMap?.children ?? [])
      .flatMap((child) => child.children ?? [])
      .flatMap((child) => child.work_unit_summary ?? []);
    return uniqueSorted(childUnits.map((unit) => unit.label)).join(", ") || "none";
  }
  return uniqueSorted(units.map((unit) => unit.label)).join(", ") || "none";
}

export function targetExecutionSummary(target, { root = repoRoot } = {}) {
  return executionSummary(targetExecutionMap(target, { root, includeArtifacts: false }));
}
