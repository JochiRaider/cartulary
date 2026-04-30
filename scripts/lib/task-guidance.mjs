import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadBrowserBatchStages } from "./browser-batch-manifest.mjs";
import {
  collectEntries,
  collectSupportGoEntries,
  loadManifest,
  phaseManifestNames,
} from "./phase-manifest.mjs";
import {
  collectTargetNames,
  collectTargetPlanRows,
  findTargetDescriptor,
} from "./target-plan.mjs";
import {
  defaultTaskSurfaceManifestPath,
  helpTiers,
  loadTaskSurfaceManifest,
  makeRecipeEntries,
  targetEntryMap,
} from "./task-surface.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");

const executionDependencyTargets = new Map([
  ["backend_unit", "backend-unit"],
  ["backend_store", "backend-store"],
  ["backend_integration", "backend-integration"],
  ["backend_process", "backend-process"],
  ["backend_integration_support", "backend-integration-support"],
  ["frontend_unit", "frontend-unit"],
  ["browser_functional", "browser-e2e-webserver-backed"],
  ["browser_stateful", "browser-e2e-stateful"],
  ["browser_measurement", "browser-e2e-measurement"],
  ["browser_support", "browser-e2e-support"],
]);

const roleDefinitions = [
  {
    role: "local-dev",
    summary: "set up tools, local services, and the development loop",
    targets: ["doctor", "bootstrap", "db-up", "dev"],
  },
  {
    role: "feature-dev",
    summary: "verify ordinary product changes before the full gate",
    targets: ["test-fast", "backend-unit", "frontend-unit", "lint"],
  },
  {
    role: "phase-author",
    summary: "inspect and maintain phase-owned evidence",
    targets: ["explain-phase", "phase-ledger-drift", "phase-schedule-drift", "test-fast"],
  },
  {
    role: "ci-investigator",
    summary: "triage failed or slow runs from retained artifacts",
    targets: ["explain-run", "explain-target", "fixture-report", "task-surface-report"],
  },
  {
    role: "release",
    summary: "run release and provider-neutral gates",
    targets: ["release-check", "ci", "build"],
  },
];

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
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

function resolveResultsRoot(root = repoRoot) {
  const configured = process.env.CARTULARY_TEST_RESULTS_DIR ?? "";
  if (configured) {
    return path.isAbsolute(configured) ? configured : path.join(root, configured);
  }
  return path.join(root, ".cartulary", "test-results");
}

function targetForEntry(entry) {
  return executionDependencyTargets.get(entry.execution_dependency ?? "") ?? "";
}

function targetForSupportEntry(entry) {
  return executionDependencyTargets.get(entry.target ?? "") ?? "";
}

function sectionCounts(rows) {
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

function collectPhaseRows(root = repoRoot) {
  const rows = [];
  for (const phase of phaseManifestNames(root)) {
    const { manifest, manifestPath } = loadManifest(root, phase);
    for (const entry of collectEntries(manifest)) {
      rows.push({
        id: entry.id,
        phase,
        section: entry.section,
        coverage: entry.coverage,
        runner: entry.runner,
        execution_dependency: entry.execution_dependency ?? "",
        target: targetForEntry(entry),
        file: entry.file ?? "",
        package: entry.package ?? "",
        title: entry.title ?? "",
        manifest_path: relToRepo(manifestPath, root),
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
        target: targetForSupportEntry(entry),
        file: entry.file ?? "",
        package: entry.package ?? "",
        title: entry.selection_pattern ?? "",
        manifest_path: relToRepo(manifestPath, root),
      });
    }
  }
  return rows.sort((left, right) => {
    return (
      compareStrings(left.phase, right.phase) ||
      compareStrings(left.target, right.target) ||
      compareStrings(left.section, right.section) ||
      compareStrings(left.id, right.id)
    );
  });
}

function loadTaskSurface(root = repoRoot) {
  const manifestFile = process.env.CARTULARY_TASK_SURFACE_MANIFEST ?? defaultTaskSurfaceManifestPath;
  const { manifest } = loadTaskSurfaceManifest(manifestFile);
  return {
    manifest,
    targets: targetEntryMap(manifest),
  };
}

function helpTierByTarget(manifest) {
  const result = new Map();
  for (const tier of helpTiers(manifest)) {
    for (const entry of tier.entries ?? []) {
      result.set(entry.target, tier.name);
    }
  }
  return result;
}

function serviceRequirementsForTarget(target, targetRows = []) {
  const requirements = new Set();
  const goDescriptor = findTargetDescriptor(target, repoRoot);
  if (goDescriptor?.serviceBacked) {
    requirements.add("Postgres");
    requirements.add("MinIO");
  }
  if (target.startsWith("browser-e2e")) {
    requirements.add("Postgres");
    requirements.add("MinIO");
    requirements.add("browser stack");
  }
  if (target === "db-up" || target === "services-up" || target === "minio-init") {
    requirements.add("Postgres");
    requirements.add("MinIO");
  }
  if (target === "dev") {
    requirements.add("Postgres");
    requirements.add("MinIO");
    requirements.add("Vite");
  }
  if (["test", "test-fast", "check", "ci", "release-check"].includes(target)) {
    requirements.add("Postgres");
    requirements.add("MinIO");
  }
  if (targetRows.some((row) => row.target.startsWith("browser-e2e"))) {
    requirements.add("browser stack");
  }
  return Array.from(requirements);
}

function loadScheduleOwners(root = repoRoot) {
  const owners = new Map();
  const serviceManifestPath = path.join(root, "tools", "service_backed_schedule_manifest.json");
  if (existsSync(serviceManifestPath)) {
    const serviceManifest = readJSON(serviceManifestPath);
    for (const schedule of serviceManifest.schedules ?? []) {
      for (const source of schedule.work_unit_sources ?? []) {
        if (!source.target) {
          continue;
        }
        const values = owners.get(source.target) ?? [];
        values.push(`${schedule.target} service-backed scheduler`);
        owners.set(source.target, values);
      }
    }
  }
  const checkManifestPath = path.join(root, "tools", "check_schedule_manifest.json");
  if (existsSync(checkManifestPath)) {
    const checkManifest = readJSON(checkManifestPath);
    for (const schedule of checkManifest.schedules ?? []) {
      for (const unit of schedule.work_units ?? []) {
        if (!unit.target) {
          continue;
        }
        const values = owners.get(unit.target) ?? [];
        values.push(`${schedule.target} check scheduler`);
        owners.set(unit.target, values);
      }
    }
  }
  return owners;
}

function browserStageOwners(root = repoRoot) {
  const owners = new Map();
  const manifestPath = path.join(root, "tools", "browser_e2e_batch_manifest.json");
  if (!existsSync(manifestPath)) {
    return owners;
  }
  for (const stage of loadBrowserBatchStages(manifestPath).values()) {
    const values = owners.get(stage.target) ?? [];
    values.push(`${stage.name} browser batch`);
    owners.set(stage.target, values);
    for (const group of stage.groups) {
      const groupValues = owners.get(group.target) ?? [];
      groupValues.push(`${stage.name} browser batch`);
      owners.set(group.target, groupValues);
    }
  }
  return owners;
}

function schedulerOwnerForTarget(target, manifest, root = repoRoot) {
  const values = [];
  const recipes = new Map(makeRecipeEntries(manifest).map((recipe) => [recipe.target, recipe]));
  const recipe = recipes.get(target);
  if (recipe?.type === "sequence") {
    values.push(`Make sequence ${recipe.sequence}`);
  } else if (recipe?.type === "check_schedule") {
    values.push(`${recipe.target} check scheduler`);
  } else if (recipe?.type === "service_backed_schedule") {
    values.push(`${target} service-backed scheduler`);
  } else if (recipe?.type === "browser_batch") {
    values.push(`${recipe.stage} browser batch`);
  } else if (recipe) {
    values.push(`Make ${recipe.type}`);
  }
  for (const owner of loadScheduleOwners(root).get(target) ?? []) {
    values.push(owner);
  }
  for (const owner of browserStageOwners(root).get(target) ?? []) {
    values.push(owner);
  }
  if (values.length === 0) {
    values.push("direct Make target");
  }
  return uniqueSorted(values);
}

function newestTargetArtifact(target, root = repoRoot) {
  const resultsRoot = resolveResultsRoot(root);
  const expected = [
    relToRepo(path.join(resultsRoot, "<run-id>", target, "target-summary.json"), root),
    relToRepo(path.join(resultsRoot, "<run-id>", target, "scheduler-summary.json"), root),
  ];
  if (["test", "test-fast", "check", "ci", "release-check"].includes(target)) {
    expected.push(relToRepo(path.join(resultsRoot, "<run-id>", "run-summary.json"), root));
  }
  if (!existsSync(resultsRoot)) {
    return { latest: null, expected };
  }
  const candidates = [];
  for (const entry of readdirSync(resultsRoot, { withFileTypes: true })) {
    if (!entry.isDirectory()) {
      continue;
    }
    const runDir = path.join(resultsRoot, entry.name);
    const artifactFiles = [
      path.join(runDir, target, "target-summary.json"),
      path.join(runDir, target, "scheduler-summary.json"),
    ];
    if (["test", "test-fast", "check", "ci", "release-check"].includes(target)) {
      artifactFiles.push(path.join(runDir, "run-summary.json"));
    }
    for (const file of artifactFiles) {
      if (!existsSync(file)) {
        continue;
      }
      const stats = statSync(file);
      candidates.push({
        path: relToRepo(file, root),
        mtime_ms: stats.mtimeMs,
      });
    }
  }
  candidates.sort((left, right) => right.mtime_ms - left.mtime_ms || left.path.localeCompare(right.path));
  return {
    latest: candidates[0] ?? null,
    expected,
  };
}

function phaseRowsForTarget(target, rows) {
  if (target === "test-fast") {
    return rows.filter(
      (row) =>
        row.coverage !== "raw" &&
        !row.target.startsWith("browser-e2e"),
    );
  }
  if (["test", "check", "ci", "release-check"].includes(target)) {
    return rows.filter((row) => row.coverage !== "raw");
  }
  if (target === "browser-e2e") {
    return rows.filter((row) => ["browser-e2e-stateful", "browser-e2e-measurement"].includes(row.target));
  }
  return rows.filter((row) => row.target === target);
}

function summarizeRows(rows) {
  const counts = sectionCounts(rows);
  return {
    ...counts,
    phases: uniqueSorted(rows.map((row) => row.phase)),
    execution_dependencies: uniqueSorted(rows.map((row) => row.execution_dependency)),
    runners: uniqueSorted(rows.map((row) => row.runner)),
    targets: uniqueSorted(rows.map((row) => row.target)),
  };
}

export function knownRoles() {
  return roleDefinitions.map((role) => role.role);
}

export function targetGuidance(target, { root = repoRoot } = {}) {
  const { manifest, targets } = loadTaskSurface(root);
  const entry = targets.get(target);
  if (!entry) {
    return null;
  }
  const tiers = helpTierByTarget(manifest);
  const allPhaseRows = collectPhaseRows(root);
  const targetRows = phaseRowsForTarget(target, allPhaseRows);
  const goRows = collectTargetNames(root).includes(target)
    ? collectTargetPlanRows(root).filter((row) => row.target === target)
    : [];
  const artifacts = newestTargetArtifact(target, root);
  return {
    target,
    classification: entry.classification,
    included_in: entry.included_in ?? [],
    help_tier: tiers.get(target) ?? null,
    backing_scripts: entry.backing_scripts ?? [],
    scheduler_owner: schedulerOwnerForTarget(target, manifest, root),
    service_requirements: serviceRequirementsForTarget(target, targetRows),
    artifact: artifacts,
    phase_coverage: summarizeRows(targetRows),
    rows: targetRows,
    go_rows: goRows,
  };
}

export function allTargetNames({ root = repoRoot } = {}) {
  const { targets } = loadTaskSurface(root);
  return Array.from(targets.keys()).sort(compareStrings);
}

export function phaseGuidance(phase, { root = repoRoot } = {}) {
  const known = phaseManifestNames(root);
  if (!known.includes(phase)) {
    return null;
  }
  const { manifest, manifestPath } = loadManifest(root, phase);
  const rows = collectPhaseRows(root).filter((row) => row.phase === phase);
  const counts = sectionCounts(rows);
  const byTarget = uniqueSorted(rows.map((row) => row.target));
  const targetSummaries = byTarget.map((target) => {
    const guidance = targetGuidance(target, { root });
    const rowsForTarget = rows.filter((row) => row.target === target);
    return {
      target,
      service_requirements: guidance?.service_requirements ?? [],
      scheduler_owner: guidance?.scheduler_owner ?? ["direct Make target"],
      counts: sectionCounts(rowsForTarget),
      execution_dependencies: uniqueSorted(rowsForTarget.map((row) => row.execution_dependency)),
    };
  });
  return {
    phase,
    manifest_path: relToRepo(manifestPath, root),
    ledger_path: `docs/testing/${phase}_coverage_ledger.md`,
    scope: manifest.ledger?.scope ?? "",
    normative_owners: manifest.ledger?.normative_owners ?? "",
    ...counts,
    counts,
    phases: [phase],
    known_phases: known,
    execution_dependencies: uniqueSorted(rows.map((row) => row.execution_dependency)),
    targets: targetSummaries,
    rows,
  };
}

export function taskGuide({ role = "", phase = "", root = repoRoot } = {}) {
  if (role && !knownRoles().includes(role)) {
    return null;
  }
  const phaseInfo = phase ? phaseGuidance(phase, { root }) : null;
  if (phase && !phaseInfo) {
    return null;
  }
  const roles = role ? roleDefinitions.filter((entry) => entry.role === role) : roleDefinitions;
  return {
    role: role || "all",
    phase: phase || "",
    phase_guidance: phaseInfo,
    roles: roles.map((definition) => ({
      ...definition,
      recommendations: definition.targets.map((target) => {
        const actualTarget = target === "explain-phase" && phase ? `explain-phase PHASE=${phase}` : target;
        const guidance = targetGuidance(target, { root });
        return {
          target: actualTarget,
          summary: guidance ? guidanceSummary(guidance) : "inspect phase evidence",
          service_requirements: guidance?.service_requirements ?? [],
          scheduler_owner: guidance?.scheduler_owner ?? ["direct Make target"],
          latest_artifact: guidance?.artifact.latest?.path ?? "none",
          expected_artifacts: guidance?.artifact.expected ?? [],
          phase_coverage: guidance?.phase_coverage ?? null,
        };
      }),
    })),
  };
}

function guidanceSummary(guidance) {
  const tier = guidance.help_tier ? `${guidance.help_tier} target` : "task-surface target";
  const includes = guidance.included_in.length > 0 ? `included in ${guidance.included_in.join(",")}` : "helper";
  return `${tier}; ${includes}`;
}

export function formatRequirements(values) {
  return values.length === 0 ? "none" : values.join(",");
}

export function formatPhaseCoverage(summary) {
  if (!summary || summary.total === 0) {
    return "none";
  }
  return [
    `phases=${summary.phases.join(",") || "none"}`,
    `authoritative=${summary.authoritative}`,
    `support=${summary.support}`,
    `supplemental=${summary.supplemental}`,
    `raw=${summary.raw}`,
    `dependencies=${summary.execution_dependencies.join(",") || "none"}`,
  ].join(" ");
}
