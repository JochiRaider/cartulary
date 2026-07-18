import path from "node:path";
import { fileURLToPath } from "node:url";

import { newestTargetArtifact } from "../contract/index.mjs";
import { collectTargetNames, collectTargetPlanRows } from "../backend/backend-target-plan.mjs";
import {
  helpTiers,
  readJSON,
  targetEntryMap,
} from "../generated-artifacts/task-surface/model.mjs";
import {
  collectExecutionOwnerRows,
  executionSummary,
  serviceRequirementsForTarget,
  summarizeExecutionRows,
  targetExecutionMap,
} from "./task-execution-map.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function loadTaskSurface(root = repoRoot) {
  const manifestFile =
    process.env.CARTULARY_TASK_SURFACE_MANIFEST ??
    path.join(root, "tools", "task_surface_manifest.json");
  const manifest = readJSON(manifestFile);
  return { manifest, targets: targetEntryMap(manifest) };
}

function helpTierByTarget(manifest) {
  const tiers = new Map();
  for (const tier of helpTiers(manifest)) {
    for (const entry of tier.entries ?? []) {
      tiers.set(entry.target, tier.name);
    }
  }
  return tiers;
}

function rowsForTarget(target, rows) {
  if (target === "test-fast") {
    return rows.filter(
      (row) => row.coverage !== "raw" && !row.target.startsWith("browser-e2e"),
    );
  }
  if (["test", "check", "ci", "release-check"].includes(target)) {
    return rows.filter((row) => row.coverage !== "raw");
  }
  if (target === "browser-e2e") {
    return rows.filter((row) =>
      [
        "browser-e2e-stateful",
        "browser-e2e-measurement",
        "browser-e2e-visual",
      ].includes(row.target),
    );
  }
  return rows.filter((row) => row.target === target);
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
  const rows = rowsForTarget(target, collectExecutionOwnerRows(root));
  const goRows = collectTargetNames(root).includes(target)
    ? collectTargetPlanRows(root).filter((row) => row.target === target)
    : [];
  const artifact = includeArtifacts
    ? newestTargetArtifact(target, { root })
    : { latest: null, candidates: [], expected: [] };
  const executionMap = includeExecutionMap
    ? targetExecutionMap(target, { root, includeArtifacts })
    : null;
  return {
    schema_id: "cartulary.target_guidance.v2",
    target,
    target_class: entry.target_class,
    default_inclusion_sets: entry.default_inclusion_sets ?? [],
    check_projection: entry.check_projection ?? null,
    input_contract: entry.input_contract ?? null,
    help_tier: helpTierByTarget(manifest).get(target) ?? null,
    backing_scripts: entry.backing_scripts ?? [],
    sequence,
    execution_map: executionMap,
    execution_summary: executionMap ? executionSummary(executionMap) : "not expanded",
    service_requirements: serviceRequirementsForTarget(
      target,
      rows,
      entry.service_requirements ?? [],
      root,
    ),
    artifact,
    step_coverage: summarizeExecutionRows(rows),
    rows,
    go_rows: goRows,
  };
}

export function allTargetNames({ root = repoRoot } = {}) {
  return [...loadTaskSurface(root).targets.keys()].sort(compareStrings);
}

export function formatRequirements(values) {
  return values.length === 0 ? "none" : values.join(",");
}

export function formatStepCoverage(summary) {
  if (!summary || summary.total === 0) {
    return "none";
  }
  return [
    `owners=${summary.owners.join(",") || "none"}`,
    `authoritative=${summary.authoritative}`,
    `support=${summary.support}`,
    `supplemental=${summary.supplemental}`,
    `raw=${summary.raw}`,
    `dependencies=${summary.execution_dependencies.join(",") || "none"}`,
  ].join(" ");
}
