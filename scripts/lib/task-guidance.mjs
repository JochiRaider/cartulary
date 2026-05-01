import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadBrowserBatchStages } from "./browser-batch-manifest.mjs";
import {
  compareExecutionDependencies,
  executionDependencyInfo,
  targetForExecutionDependency,
} from "./execution-dependencies.mjs";
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
  helpTiers,
  loadTaskSurfaceManifest,
  makeRecipeEntries,
  targetEntryMap,
} from "./task-surface.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");

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
  const manifestFile =
    process.env.CARTULARY_TASK_SURFACE_MANIFEST ?? path.join(root, "tools", "task_surface_manifest.json");
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
  const matchingArtifact = (file, field) => {
    try {
      return readJSON(file)?.[field] === target;
    } catch {
      return false;
    }
  };
  for (const entry of readdirSync(resultsRoot, { withFileTypes: true })) {
    if (!entry.isDirectory()) {
      continue;
    }
    const runDir = path.join(resultsRoot, entry.name);
    const artifactFiles = [
      { file: path.join(runDir, target, "target-summary.json"), field: "target" },
      { file: path.join(runDir, target, "scheduler-summary.json"), field: "target" },
    ];
    if (["test", "test-fast", "check", "ci", "release-check"].includes(target)) {
      artifactFiles.push({ file: path.join(runDir, "run-summary.json"), field: "label" });
    }
    for (const { file, field } of artifactFiles) {
      if (!existsSync(file)) {
        continue;
      }
      if (!matchingArtifact(file, field)) {
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

function phaseRowsForTargets(targets, rows) {
  const targetSet = new Set(targets);
  return rows.filter((row) => targetSet.has(row.target));
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

function executionDependencyCategories(values) {
  return uniqueSorted(
    values.map((value) => executionDependencyInfo(value)?.category ?? ""),
  );
}

function executionDependencyServiceBacked(values) {
  return values.some((value) => executionDependencyInfo(value)?.service_backed === true);
}

function firstExecutionDependencyOrder(values) {
  const dependencies = uniqueSorted(values).sort(compareExecutionDependencies);
  const first = dependencies[0] ?? "";
  return executionDependencyInfo(first)?.order ?? Number.MAX_SAFE_INTEGER;
}

function comparePhaseTargets(left, right) {
  return (
    firstExecutionDependencyOrder(left.execution_dependencies) -
      firstExecutionDependencyOrder(right.execution_dependencies) ||
    compareStrings(left.target, right.target)
  );
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
    const executionDependencies = uniqueSorted(
      rowsForTarget.map((row) => row.execution_dependency),
    ).sort(compareExecutionDependencies);
    return {
      target,
      service_requirements: guidance?.service_requirements ?? [],
      scheduler_owner: guidance?.scheduler_owner ?? ["direct Make target"],
      classification: guidance?.classification ?? null,
      included_in: guidance?.included_in ?? [],
      counts: sectionCounts(rowsForTarget),
      execution_dependencies: executionDependencies,
      execution_categories: executionDependencyCategories(executionDependencies),
      service_backed: executionDependencyServiceBacked(executionDependencies),
    };
  }).sort(comparePhaseTargets);
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
    execution_dependencies: uniqueSorted(rows.map((row) => row.execution_dependency)).sort(
      compareExecutionDependencies,
    ),
    targets: targetSummaries,
    rows,
  };
}

export function phaseSlice(phase, { root = repoRoot, serviceBackedOnly = false } = {}) {
  const phaseInfo = phaseGuidance(phase, { root });
  if (!phaseInfo) {
    return null;
  }
  const childTargets = phaseInfo.targets
    .filter((target) => !serviceBackedOnly || target.service_requirements.length > 0 || target.service_backed)
    .map((target) => ({
      target: target.target,
      classification: target.classification,
      included_in: target.included_in,
      service_requirements: target.service_requirements,
      scheduler_owner: target.scheduler_owner,
      counts: target.counts,
      execution_dependencies: target.execution_dependencies,
      execution_categories: target.execution_categories,
      service_backed: target.service_backed,
    }));
  const childTargetNames = childTargets.map((target) => target.target);
  const rows = phaseRowsForTargets(childTargetNames, phaseInfo.rows);
  const requirements = Array.from(
    new Set(childTargets.flatMap((target) => target.service_requirements)),
  );
  return {
    target: serviceBackedOnly ? "service-backed-slice" : "phase-slice",
    phase: phaseInfo.phase,
    mode: serviceBackedOnly ? "service_backed" : "phase",
    service_backed_only: serviceBackedOnly,
    child_targets: childTargets,
    service_requirements: requirements,
    phase_coverage: summarizeRows(rows),
    no_op: childTargets.length === 0,
  };
}

function recommendationForTarget(
  target,
  {
    root = repoRoot,
    actualTarget = target,
    phaseRows = null,
    phaseRelevance = "general",
    summaryOverride = "",
  } = {},
) {
  const guidance = targetGuidance(target, { root });
  const targetRows = phaseRows ?? null;
  const coverage = targetRows ? summarizeRows(targetRows) : (guidance?.phase_coverage ?? null);
  const executionDependencies = coverage?.execution_dependencies ?? [];
  return {
    target: actualTarget,
    summary: summaryOverride || (guidance ? guidanceSummary(guidance) : "inspect phase evidence"),
    phase_relevance: phaseRelevance,
    execution_dependencies: executionDependencies,
    execution_categories: executionDependencyCategories(executionDependencies),
    service_backed: executionDependencyServiceBacked(executionDependencies),
    service_requirements: guidance?.service_requirements ?? [],
    scheduler_owner: guidance?.scheduler_owner ?? ["direct Make target"],
    latest_artifact: guidance?.artifact.latest?.path ?? "none",
    expected_artifacts: guidance?.artifact.expected ?? [],
    phase_coverage: coverage,
  };
}

function recommendationForPhaseSlice(
  slice,
  {
    root = repoRoot,
    phaseRelevance,
    summaryOverride,
  } = {},
) {
  const guidance = targetGuidance(slice.target, { root });
  const coverage = slice.phase_coverage;
  const executionDependencies = coverage?.execution_dependencies ?? [];
  return {
    target: `${slice.target} PHASE=${slice.phase}`,
    summary: summaryOverride,
    phase_relevance: phaseRelevance,
    execution_dependencies: executionDependencies,
    execution_categories: executionDependencyCategories(executionDependencies),
    service_backed: executionDependencyServiceBacked(executionDependencies),
    service_requirements: slice.service_requirements,
    scheduler_owner: guidance?.scheduler_owner ?? ["direct Make target"],
    latest_artifact: guidance?.artifact.latest?.path ?? "none",
    expected_artifacts: guidance?.artifact.expected ?? [],
    phase_coverage: coverage,
    child_targets: slice.child_targets,
  };
}

function defaultRecommendationTiers(definition, { phase = "", root = repoRoot } = {}) {
  return [
    {
      name: "recommended targets",
      summary: "role-oriented targets from the stable task surface",
      recommendations: definition.targets.map((target) => {
        const actualTarget = target === "explain-phase" && phase ? `explain-phase PHASE=${phase}` : target;
        return recommendationForTarget(target, {
          root,
          actualTarget,
          phaseRelevance: phase ? "phase_context" : "general",
        });
      }),
    },
  ];
}

function featureDevPhaseRecommendationTiers(definition, phaseInfo, { root = repoRoot } = {}) {
  const phaseTargets = [...phaseInfo.targets].sort(comparePhaseTargets);
  const phaseTargetNames = new Set(phaseTargets.map((target) => target.target));
  const fullLocalGateTargets = ["test-fast", "check"];
  const hygieneTargets = definition.targets.filter(
    (target) => !phaseTargetNames.has(target) && !fullLocalGateTargets.includes(target),
  );

  const phaseRecommendations = [
    recommendationForPhaseSlice(phaseSlice(phaseInfo.phase, { root }), {
      root,
      phaseRelevance: "phase_slice",
      summaryOverride: "selected phase target slice",
    }),
  ];
  const serviceBackedRecommendations = [
    recommendationForPhaseSlice(phaseSlice(phaseInfo.phase, { root, serviceBackedOnly: true }), {
      root,
      phaseRelevance: "service_backed_slice",
      summaryOverride: "selected phase service-backed target slice",
    }),
  ];
  const fullLocalGateRecommendations = fullLocalGateTargets.map((target) =>
    recommendationForTarget(target, {
      root,
      phaseRelevance: "full_local_gate",
      summaryOverride: "aggregate local verification gate",
    }),
  );
  const hygieneRecommendations = hygieneTargets.map((target) =>
    recommendationForTarget(target, {
      root,
      phaseRelevance: "general_hygiene",
      summaryOverride: "general hygiene outside the selected phase slice",
    }),
  );

  return [
    {
      name: "minimal phase slice",
      summary: `direct targets that cover ${phaseInfo.phase}`,
      recommendations: phaseRecommendations,
    },
    {
      name: "service-backed slice",
      summary: `service-backed targets that cover ${phaseInfo.phase}`,
      recommendations: serviceBackedRecommendations,
    },
    {
      name: "full local gate",
      summary: "aggregate verification before handoff or review",
      recommendations: fullLocalGateRecommendations,
    },
    {
      name: "general hygiene",
      summary: "useful non-phase checks that do not claim selected phase evidence",
      recommendations: hygieneRecommendations,
    },
  ].filter((tier) => tier.recommendations.length > 0);
}

function recommendationTiersForRole(definition, { phase = "", phaseInfo = null, root = repoRoot } = {}) {
  if (definition.role === "feature-dev" && phaseInfo) {
    return featureDevPhaseRecommendationTiers(definition, phaseInfo, { root });
  }
  return defaultRecommendationTiers(definition, { phase, root });
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
      role: definition.role,
      summary: definition.summary,
      recommendation_tiers: recommendationTiersForRole(definition, {
        phase,
        phaseInfo,
        root,
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
