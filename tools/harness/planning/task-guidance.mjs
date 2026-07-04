import path from "node:path";
import { fileURLToPath } from "node:url";

import { newestTargetArtifact as discoverNewestTargetArtifact } from "../contract/index.mjs";
import {
  compareExecutionDependencies,
  executionDependencyInfo,
} from "../scheduler/execution-dependencies.mjs";
import { collectEntries, loadManifest } from "./phase-manifest.mjs";
import {
  activePhaseStatus,
  phaseManifestRoot,
  phaseRegistryEntries,
  phaseRegistryEntry,
} from "./phase-registry.mjs";
import {
  collectExecutionPhaseRows,
  executionSummary,
  phaseExecutionMap,
  phaseSliceExecutionMap,
  sectionCounts,
  serviceRequirementsForTarget,
  summarizeExecutionRows,
  targetExecutionMap,
  targetExecutionSummary,
} from "./task-execution-map.mjs";
import {
  collectTargetNames,
  collectTargetPlanRows,
} from "./target-plan.mjs";
import {
  helpTiers,
  loadTaskSurfaceManifest,
  targetEntryMap,
} from "../generated-artifacts/task-surface.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const taskSurfaceCache = new Map();
const helpTierCache = new WeakMap();

const roleDefinitions = [
  {
    role: "local-dev",
    summary: "set up tools, local services, and the development loop",
    targets: ["doctor", "bootstrap", "db-up", "dev"],
  },
  {
    role: "feature-dev",
    summary: "verify ordinary product changes before the full gate",
    targets: ["agent-finalize", "test-fast", "backend-unit", "frontend-unit", "lint"],
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

function collectPhaseRows(root = repoRoot) {
  return collectExecutionPhaseRows(root);
}

function loadTaskSurface(root = repoRoot, taskSurfaceManifest = null) {
  if (taskSurfaceManifest) {
    return {
      manifest: taskSurfaceManifest,
      targets: targetEntryMap(taskSurfaceManifest),
    };
  }
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

function helpTierByTarget(manifest) {
  const cached = helpTierCache.get(manifest);
  if (cached) {
    return cached;
  }
  const result = new Map();
  for (const tier of helpTiers(manifest)) {
    for (const entry of tier.entries ?? []) {
      result.set(entry.target, tier.name);
    }
  }
  helpTierCache.set(manifest, result);
  return result;
}

function phaseRowsForTarget(target, rows) {
  if (target === "test-fast") {
    return rows.filter(
      (row) =>
        row.coverage !== "raw" &&
        !row.target.startsWith("browser-e2e"),
    );
  }
  if (target === "check") {
    return rows.filter(
      (row) =>
        row.coverage !== "raw" &&
        row.target !== "browser-e2e-measurement" &&
        row.default_check_required === true,
    );
  }
  if (["test", "ci", "release-check"].includes(target)) {
    return rows.filter((row) => row.coverage !== "raw");
  }
  if (target === "browser-e2e") {
    return rows.filter((row) =>
      ["browser-e2e-stateful", "browser-e2e-measurement", "browser-e2e-visual"].includes(row.target),
    );
  }
  return rows.filter((row) => row.target === target);
}

function phaseRowsForTargets(targets, rows) {
  const targetSet = new Set(targets);
  return rows.filter((row) => targetSet.has(row.target));
}

const summarizeRows = summarizeExecutionRows;

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

export function targetGuidance(
  target,
  { root = repoRoot, includeExecutionMap = true, includeArtifacts = true } = {},
) {
  const { manifest, targets } = loadTaskSurface(root);
  const entry = targets.get(target);
  if (!entry) {
    return null;
  }
  const recipe = manifest.make_recipes?.[target] ?? null;
  const sequence =
    recipe?.type === "sequence"
      ? (manifest.sequences?.[recipe.sequence] ?? null)
      : null;
  const tiers = helpTierByTarget(manifest);
  const allPhaseRows = collectPhaseRows(root);
  const targetRows = phaseRowsForTarget(target, allPhaseRows);
  const goRows = collectTargetNames(root).includes(target)
    ? collectTargetPlanRows(root).filter((row) => row.target === target)
    : [];
  const artifacts = includeArtifacts
    ? discoverNewestTargetArtifact(target, { root })
    : { latest: null, candidates: [], expected: [] };
  const executionMap = includeExecutionMap ? targetExecutionMap(target, { root }) : null;
  return {
    schema_id: "cartulary.target_guidance.v1",
    target,
    target_class: entry.target_class,
    default_inclusion_sets: entry.default_inclusion_sets ?? [],
    check_projection: entry.check_projection ?? null,
    input_contract: entry.input_contract ?? null,
    help_tier: tiers.get(target) ?? null,
    backing_scripts: entry.backing_scripts ?? [],
    sequence,
    execution_map: executionMap,
    execution_summary: executionMap ? executionSummary(executionMap) : "not expanded",
    service_requirements: serviceRequirementsForTarget(target, targetRows, entry.service_requirements, root),
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

export function phaseGuidance(phase, { root = repoRoot, includeExecutionMap = true, taskSurfaceManifest = null } = {}) {
  const registryEntry = phaseRegistryEntry(root, phase);
  if (!registryEntry) {
    return null;
  }
  const known = phaseRegistryEntries(root).map((entry) => entry.phase);
  if (registryEntry.status !== activePhaseStatus) {
    let manifestRows = [];
    try {
      const { manifest } = loadManifest(root, phase, { allowPlanned: true });
      manifestRows = collectEntries(manifest).map((row) => ({
        ...row,
        phase,
      }));
    } catch {
      manifestRows = [];
    }
    const counts = sectionCounts(manifestRows);
    return {
      schema_id: "cartulary.phase_guidance.v1",
      phase,
      status: registryEntry.status,
      label: registryEntry.label,
      manifest_path: registryEntry.manifest_path,
      ledger_path: registryEntry.ledger_path,
      scope: registryEntry.scope,
      normative_owners: registryEntry.normative_owners,
      ...counts,
      counts,
      phases: [phase],
      known_phases: known,
      execution_dependencies: uniqueSorted(
        manifestRows.map((row) => row.execution_dependency),
      ).sort(compareExecutionDependencies),
      targets: [],
      rows: manifestRows,
      execution_map: includeExecutionMap ? phaseExecutionMap(phase, { root }) : null,
    };
  }
  const { manifestPath } = loadManifest(root, phase);
  const rows = collectPhaseRows(root).filter((row) => row.phase === phase);
  const counts = sectionCounts(rows);
  const byTarget = uniqueSorted(rows.map((row) => row.target));
  const { targets } = loadTaskSurface(root, taskSurfaceManifest);
  const targetSummaries = byTarget.map((target) => {
    const entry = targets.get(target);
    const rowsForTarget = rows.filter((row) => row.target === target);
    const executionDependencies = uniqueSorted(
      rowsForTarget.map((row) => row.execution_dependency),
    ).sort(compareExecutionDependencies);
    return {
      target,
      service_requirements: serviceRequirementsForTarget(
        target,
        rowsForTarget,
        entry?.service_requirements ?? [],
        root,
      ),
      execution_summary: "not expanded",
      target_class: entry?.target_class ?? null,
      default_inclusion_sets: entry?.default_inclusion_sets ?? [],
      counts: sectionCounts(rowsForTarget),
      execution_dependencies: executionDependencies,
      execution_categories: executionDependencyCategories(executionDependencies),
      service_backed: executionDependencyServiceBacked(executionDependencies),
    };
  }).sort(comparePhaseTargets);
  return {
    schema_id: "cartulary.phase_guidance.v1",
    phase,
    status: registryEntry.status,
    label: registryEntry.label,
    manifest_path: relToRepo(manifestPath, phaseManifestRoot(root)),
    ledger_path: registryEntry.ledger_path,
    scope: registryEntry.scope,
    normative_owners: registryEntry.normative_owners,
    ...counts,
    counts,
    phases: [phase],
    known_phases: known,
    execution_dependencies: uniqueSorted(rows.map((row) => row.execution_dependency)).sort(
      compareExecutionDependencies,
    ),
    targets: targetSummaries,
    rows,
    execution_map: includeExecutionMap ? phaseExecutionMap(phase, { root }) : null,
  };
}

export function phaseSlice(
  phase,
  { root = repoRoot, serviceBackedOnly = false, includeExecutionMap = true, taskSurfaceManifest = null } = {},
) {
  const phaseInfo = phaseGuidance(phase, { root, includeExecutionMap: false, taskSurfaceManifest });
  if (!phaseInfo) {
    return null;
  }
  const childTargets = phaseInfo.targets
    .filter((target) => !serviceBackedOnly || target.service_requirements.length > 0 || target.service_backed)
    .map((target) => ({
      target: target.target,
      target_class: target.target_class,
      default_inclusion_sets: target.default_inclusion_sets,
      service_requirements: target.service_requirements,
      execution_summary: target.execution_summary,
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
  const executionMap = includeExecutionMap
    ? phaseSliceExecutionMap(phase, { root, serviceBackedOnly })
    : null;
  return {
    target: serviceBackedOnly ? "service-backed-slice" : "phase-slice",
    phase: phaseInfo.phase,
    mode: serviceBackedOnly ? "service_backed" : "phase",
    service_backed_only: serviceBackedOnly,
    child_targets: childTargets,
    service_requirements: requirements,
    phase_coverage: summarizeRows(rows),
    execution_map: executionMap,
    execution_summary: executionMap ? executionSummary(executionMap) : "not expanded",
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
  const guidance = targetGuidance(target, { root, includeExecutionMap: false });
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
    execution_summary: guidance ? targetExecutionSummary(target, { root }) : "none",
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
  const guidance = targetGuidance(slice.target, { root, includeExecutionMap: false });
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
    execution_summary: slice.execution_summary ?? guidance?.execution_summary ?? "none",
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
      summaryOverride: "selected phase manifest-row slice",
    }),
  ];
  const serviceBackedRecommendations = [
    recommendationForPhaseSlice(phaseSlice(phaseInfo.phase, { root, serviceBackedOnly: true }), {
      root,
      phaseRelevance: "service_backed_slice",
      summaryOverride: "selected phase service-backed manifest-row slice",
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
      summary: `manifest rows that cover ${phaseInfo.phase}`,
      recommendations: phaseRecommendations,
    },
    {
      name: "service-backed slice",
      summary: `service-backed manifest rows that cover ${phaseInfo.phase}`,
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
  const phaseGuidanceForOutput = phaseInfo ? { ...phaseInfo, execution_map: undefined } : null;
  const roles = role ? roleDefinitions.filter((entry) => entry.role === role) : roleDefinitions;
  return {
    schema_id: "cartulary.task_guide.v1",
    role: role || "all",
    phase: phase || "",
    phase_guidance: phaseGuidanceForOutput,
    execution_map: phaseInfo?.execution_map ?? null,
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
  const includes =
    guidance.default_inclusion_sets.length > 0
      ? `default inclusion sets ${guidance.default_inclusion_sets.join(",")}`
      : "no default inclusion set";
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
