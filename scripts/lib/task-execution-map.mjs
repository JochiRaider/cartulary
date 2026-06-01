import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { newestTargetArtifact } from "./artifact-discovery.mjs";
import { normalizeBrowserBatchStages } from "./browser-batch-manifest.mjs";
import {
  compareExecutionDependencies,
  executionDependencyInfo,
  targetForExecutionDependency,
} from "./execution-dependencies.mjs";
import {
  loadExecutionTopology,
  renderBrowserBatchManifest,
} from "./execution-topology.mjs";
import {
  collectEntries,
  collectSupportGoEntries,
  entryClaimStatus,
  loadManifest,
  phaseManifestNames,
  playwrightEntryTitles,
  vitestEntryTitles,
} from "./phase-manifest.mjs";
import {
  activePhaseStatus,
  phaseManifestRoot,
  phaseRegistryEntry,
} from "./phase-registry.mjs";
import { findTargetDescriptor } from "./target-plan.mjs";
import {
  loadTaskSurfaceManifest,
  makeRecipeEntries,
  targetEntryMap,
} from "./task-surface.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");
export const taskExecutionMapSchemaID = "cartulary.task_execution_map.v1";
const executionPhaseRowsCache = new Map();
const taskSurfaceCache = new Map();
const schedulerManifestCache = new Map();
const browserStagesCache = new Map();

const serviceRequirementDisplayNames = new Map([
  ["postgres", "Postgres"],
  ["minio", "MinIO"],
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

function relToRepo(value, root = repoRoot) {
  const relative = path.relative(root, value).replaceAll("\\", "/");
  if (!relative.startsWith("../") && relative !== "..") {
    return relative === "" ? "." : relative;
  }
  return value.replaceAll("\\", "/");
}

function targetForEntry(entry) {
  return targetForExecutionDependency(
    entry.execution_dependency ?? "",
    `manifest entry ${entry.id} execution_dependency`,
  );
}

function targetForSupportEntry(entry) {
  return targetForExecutionDependency(
    entry.target ?? "",
    `support_go_target ${entry.file ?? "(missing file)"} target`,
  );
}

export function collectExecutionPhaseRows(root = repoRoot) {
  const cacheKey = cacheKeyForRoot(root);
  const cached = executionPhaseRowsCache.get(cacheKey);
  if (cached) {
    return cached;
  }
  const rows = [];
  for (const phase of phaseManifestNames(root)) {
    const { manifest, manifestPath } = loadManifest(root, phase);
    for (const entry of collectEntries(manifest)) {
      rows.push({
        id: entry.id,
        phase,
        section: entry.section,
        coverage: entry.coverage,
        claim_status: entryClaimStatus(entry),
        runner: entry.runner,
        execution_dependency: entry.execution_dependency ?? "",
        evidence_class: entry.evidence_class ?? "",
        layer: entry.layer ?? "",
        default_check_required: entry.default_check_required === true,
        target: targetForEntry(entry),
        file: entry.file ?? "",
        package: entry.package ?? "",
        title:
          entry.runner === "vitest"
            ? vitestEntryTitles(entry).join(" | ")
            : entry.runner === "playwright"
              ? playwrightEntryTitles(entry).join(" | ")
              : entry.title ?? "",
        manifest_path: relToRepo(manifestPath, phaseManifestRoot(root)),
      });
    }
    for (const entry of collectSupportGoEntries(manifest)) {
      rows.push({
        id: `support:${entry.file}`,
        phase,
        section: entry.section ?? "",
        coverage: "support",
        runner: "go_test",
        execution_dependency: entry.target ?? "",
        evidence_class: entry.evidence_class ?? "",
        layer: entry.layer ?? "",
        default_check_required: entry.default_check_required === true,
        target: targetForSupportEntry(entry),
        file: entry.file ?? "",
        package: entry.package ?? "",
        title: entry.selection_pattern ?? "",
        manifest_path: relToRepo(manifestPath, phaseManifestRoot(root)),
      });
    }
  }
  const result = rows.sort((left, right) => (
    compareStrings(left.phase, right.phase) ||
    compareStrings(left.target, right.target) ||
    compareStrings(left.section, right.section) ||
    compareStrings(left.id, right.id)
  ));
  executionPhaseRowsCache.set(cacheKey, result);
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
    phases: uniqueSorted(rows.map((row) => row.phase)),
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
  const { manifest } = loadTaskSurfaceManifest(manifestFile);
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
  addServiceRequirement(requirements, "minio");
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
    addServiceRequirement(requirements, "minio");
  }
  if (target.startsWith("browser-e2e")) {
    addServiceRequirement(requirements, "postgres");
    addServiceRequirement(requirements, "minio");
    addServiceRequirement(requirements, "browser_stack");
  }
  if (target === "db-up" || target === "services-up" || target === "minio-init") {
    addServiceRequirement(requirements, "postgres");
    addServiceRequirement(requirements, "minio");
  }
  if (target === "dev") {
    addServiceRequirement(requirements, "postgres");
    addServiceRequirement(requirements, "minio");
    addServiceRequirement(requirements, "vite");
  }
  if (["test", "test-fast", "check", "ci", "release-check"].includes(target)) {
    addServiceRequirement(requirements, "postgres");
    addServiceRequirement(requirements, "minio");
  }
  if (targetRows.some((row) => row.target.startsWith("browser-e2e"))) {
    addServiceRequirement(requirements, "browser_stack");
  }
  return Array.from(requirements);
}

function phaseRowsForTarget(target, rows) {
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

function phaseSchedulerSummary(target) {
  if (target !== "phase-slice" && target !== "service-backed-slice") {
    return [];
  }
  return [{
    kind: "phase_scheduler",
    id: `phase-scheduler:${target}`,
    label: "phase scheduler",
    target,
    scheduler: "phase-scheduler",
    detail: target === "service-backed-slice"
      ? "selected phase service-backed manifest-row slice"
      : "selected phase manifest-row slice",
  }];
}

function targetWorkUnitSummary(target, manifest, root = repoRoot) {
  const summaries = [
    ...phaseSchedulerSummary(target),
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
  const targetRows = rows ?? phaseRowsForTarget(target, collectExecutionPhaseRows(root));
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
    phases: uniqueSorted(summaries.flatMap((summary) => summary.phases ?? [])),
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

function comparePhaseTargetNodes(left, right) {
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

export function phaseExecutionMap(phase, { root = repoRoot } = {}) {
  const registryEntry = phaseRegistryEntry(root, phase);
  if (!registryEntry) {
    return null;
  }
  const executable = registryEntry.status === activePhaseStatus;
  const rows = executable
    ? collectExecutionPhaseRows(root).filter((row) => row.phase === phase)
    : [];
  const rowsByTarget = new Map();
  for (const row of rows) {
    if (!rowsByTarget.has(row.target)) {
      rowsByTarget.set(row.target, []);
    }
    rowsByTarget.get(row.target).push(row);
  }

  const publicChildren = [];
  const supportChildren = [];
  const { targets } = loadTaskSurface(root);
  for (const [target, targetRows] of rowsByTarget.entries()) {
    const classification = targets.get(target)?.target_class ?? null;
    const publicRows = targetRows.filter((row) => classification === "public" && row.coverage !== "support");
    const supportRows = targetRows.filter((row) => classification !== "public" || row.coverage === "support");
    if (publicRows.length > 0) {
      publicChildren.push(targetNodeForRows(target, publicRows, { root }));
    }
    if (supportRows.length > 0) {
      supportChildren.push(targetNodeForRows(target, supportRows, { root }));
    }
  }
  publicChildren.sort(comparePhaseTargetNodes);
  supportChildren.sort(comparePhaseTargetNodes);

  const sections = [
    makeSectionNode("public-evidence", "public evidence", publicChildren),
    makeSectionNode("support-internal-evidence", "support/internal evidence", supportChildren),
  ].filter((section) => section.children.length > 0);

  return {
    schema_id: taskExecutionMapSchemaID,
    kind: "phase",
    id: phase,
    label: phase,
    target_class: registryEntry.status,
    coverage: summarizeExecutionRows(rows),
    services: uniqueSorted(sections.flatMap((section) => section.services ?? [])),
    execution_dependencies: uniqueSorted(sections.flatMap((section) => section.execution_dependencies ?? []))
      .sort(compareExecutionDependencies),
    work_unit_summary: [],
    artifacts: {
      manifest: registryEntry.manifest_path,
      ledger: registryEntry.ledger_path,
    },
    children: sections,
  };
}

export function phaseSliceExecutionMap(phase, { root = repoRoot, serviceBackedOnly = false } = {}) {
  const map = phaseExecutionMap(phase, { root });
  if (!map || !serviceBackedOnly) {
    return map;
  }
  const serviceSections = map.children
    .map((section) => ({
      ...section,
      children: section.children.filter((child) =>
        child.execution_dependencies.some((dependency) => executionDependencyInfo(dependency)?.service_backed === true),
      ),
    }))
    .filter((section) => section.children.length > 0)
    .map((section) => makeSectionNode(section.id, section.label, section.children));
  return {
    ...map,
    id: `${phase}:service-backed`,
    label: `${phase} service-backed`,
    coverage: sectionCoverage(serviceSections),
    services: uniqueSorted(serviceSections.flatMap((section) => section.services ?? [])),
    execution_dependencies: uniqueSorted(serviceSections.flatMap((section) => section.execution_dependencies ?? []))
      .sort(compareExecutionDependencies),
    children: serviceSections,
  };
}

export function executionSummary(executionMap) {
  const units = executionMap?.work_unit_summary ?? [];
  if (executionMap?.kind === "phase") {
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
