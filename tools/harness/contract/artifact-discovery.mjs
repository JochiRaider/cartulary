import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  loadTaskSurfaceManifest,
  makeRecipeEntries,
} from "../generated-artifacts/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const aggregateRunSummaryTargets = new Set(["test", "test-fast", "check", "ci", "release-check"]);
const commandArtifactDefinitions = new Map([
  [
    "release-readiness-evidence",
    [
      {
        kind: "release_readiness_evidence",
        relativePath: "release-readiness-evidence.json",
        schemaID: "cartulary.release_readiness_evidence.v2",
      },
    ],
  ],
]);

export function relToRepo(value, root = repoRoot) {
  const relative = path.relative(root, value).replaceAll("\\", "/");
  if (!relative.startsWith("../") && relative !== "..") {
    return relative === "" ? "." : relative;
  }
  return value.replaceAll("\\", "/");
}

export function resolveResultsRoot(root = repoRoot, env = process.env) {
  const configured = env.CARTULARY_TEST_RESULTS_DIR ?? "";
  if (configured) {
    return path.isAbsolute(configured) ? configured : path.join(root, configured);
  }
  return path.join(root, ".cartulary", "test-results");
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function safeReadJSON(file) {
  try {
    return readJSON(file);
  } catch {
    return null;
  }
}

function artifactRecord(file, kind, runID, root) {
  const stats = statSync(file);
  return {
    kind,
    path: relToRepo(file, root),
    run_id: runID,
    mtime_ms: stats.mtimeMs,
  };
}

function walkPhaseSummaries(targetDir) {
  if (!existsSync(targetDir)) {
    return [];
  }
  const files = [];
  const stack = [targetDir];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (entry.isFile() && entry.name === "phase-summary.json") {
        files.push(next);
      }
    }
  }
  return files.sort((left, right) => left.localeCompare(right));
}

function targetRecipe(target, root = repoRoot) {
  const manifestPath =
    process.env.CARTULARY_TASK_SURFACE_MANIFEST ?? path.join(root, "tools", "task_surface_manifest.json");
  try {
    const { manifest } = loadTaskSurfaceManifest(manifestPath);
    return makeRecipeEntries(manifest).find((recipe) => recipe.target === target) ?? null;
  } catch {
    return null;
  }
}

function targetHasRunSummary(target, recipe) {
  return recipe?.type === "sequence" || recipe?.type === "check_schedule" || aggregateRunSummaryTargets.has(target);
}

export function expectedTargetArtifacts(target, { root = repoRoot } = {}) {
  const resultsRoot = resolveResultsRoot(root);
  const recipe = targetRecipe(target, root);
  const recipeType = recipe?.type ?? "";
  const expected = [
    relToRepo(path.join(resultsRoot, "<run-id>", target, "target-summary.json"), root),
  ];
  if (recipeType === "check_schedule" || recipeType === "service_backed_schedule") {
    expected.push(
      relToRepo(path.join(resultsRoot, "<run-id>", target, "scheduler-summary.json"), root),
      relToRepo(path.join(resultsRoot, "<run-id>", target, "scheduler-events.jsonl"), root),
      relToRepo(path.join(resultsRoot, "<run-id>", target, "pressure-summary.json"), root),
      relToRepo(path.join(resultsRoot, "<run-id>", target, "progress-summary.log"), root),
    );
  }
  if (!["sequence", "check_schedule", "service_backed_schedule"].includes(recipeType)) {
    expected.push(relToRepo(path.join(resultsRoot, "<run-id>", target, "<phase-label>", "phase-summary.json"), root));
  }
  if (targetHasRunSummary(target, recipe)) {
    expected.push(relToRepo(path.join(resultsRoot, "<run-id>", "run-summary.json"), root));
  }
  for (const artifact of commandArtifactDefinitions.get(target) ?? []) {
    expected.push(relToRepo(path.join(resultsRoot, "<run-id>", target, artifact.relativePath), root));
  }
  return expected;
}

export function targetArtifactCandidates(target, { root = repoRoot } = {}) {
  const resultsRoot = resolveResultsRoot(root);
  if (!existsSync(resultsRoot)) {
    return [];
  }
  const recipe = targetRecipe(target, root);
  const hasRunSummary = targetHasRunSummary(target, recipe);
  const candidates = [];
  for (const entry of readdirSync(resultsRoot, { withFileTypes: true })) {
    if (!entry.isDirectory()) {
      continue;
    }
    const runID = entry.name;
    const runDir = path.join(resultsRoot, runID);
    const targetDir = path.join(runDir, target);
    const targetSummary = path.join(targetDir, "target-summary.json");
    if (existsSync(targetSummary) && safeReadJSON(targetSummary)?.target === target) {
      candidates.push(artifactRecord(targetSummary, "target_summary", runID, root));
    }
    const schedulerSummary = path.join(targetDir, "scheduler-summary.json");
    if (existsSync(schedulerSummary) && safeReadJSON(schedulerSummary)?.target === target) {
      candidates.push(artifactRecord(schedulerSummary, "scheduler_summary", runID, root));
    }
    const schedulerEvents = path.join(targetDir, "scheduler-events.jsonl");
    if (existsSync(schedulerEvents)) {
      candidates.push(artifactRecord(schedulerEvents, "scheduler_events", runID, root));
    }
    const pressureSummary = path.join(targetDir, "pressure-summary.json");
    if (existsSync(pressureSummary) && safeReadJSON(pressureSummary)?.target === target) {
      candidates.push(artifactRecord(pressureSummary, "pressure_summary", runID, root));
    }
    if (hasRunSummary) {
      const runSummary = path.join(runDir, "run-summary.json");
      if (existsSync(runSummary) && safeReadJSON(runSummary)?.label === target) {
        candidates.push(artifactRecord(runSummary, "run_summary", runID, root));
      }
    }
    for (const phaseSummary of walkPhaseSummaries(targetDir)) {
      const summary = safeReadJSON(phaseSummary);
      if (summary?.target === target) {
        candidates.push({
          ...artifactRecord(phaseSummary, "phase_summary", runID, root),
          label: summary.label ?? "",
          status: summary.status ?? "",
        });
      }
    }
    for (const artifact of commandArtifactDefinitions.get(target) ?? []) {
      const artifactFile = path.join(targetDir, artifact.relativePath);
      if (existsSync(artifactFile) && safeReadJSON(artifactFile)?.schema_id === artifact.schemaID) {
        candidates.push(artifactRecord(artifactFile, artifact.kind, runID, root));
      }
    }
  }
  return candidates.sort((left, right) => right.mtime_ms - left.mtime_ms || left.path.localeCompare(right.path));
}

export function newestTargetArtifact(target, { root = repoRoot } = {}) {
  const candidates = targetArtifactCandidates(target, { root });
  return {
    latest: candidates[0] ?? null,
    candidates,
    expected: expectedTargetArtifacts(target, { root }),
  };
}

export function helperArtifactReferences(helperTargets, { root = repoRoot, runId = "" } = {}) {
  const resultsRoot = resolveResultsRoot(root);
  const currentRunID = runId || process.env.CARTULARY_TEST_RUN_ID || "adhoc";
  return helperTargets
    .map((target) => {
      const targetDir = path.join(resultsRoot, currentRunID, target);
      const phaseSummaries = walkPhaseSummaries(targetDir)
        .map((file) => {
          const summary = safeReadJSON(file);
          if (summary?.target !== target) {
            return null;
          }
          return {
            label: summary.label ?? "",
            status: summary.status ?? "",
            artifact: relToRepo(file, root),
            runner_json: summary.artifacts?.runner_json ?? "",
            stdout_log: summary.artifacts?.stdout_log ?? "",
            stderr_log: summary.artifacts?.stderr_log ?? "",
          };
        })
        .filter(Boolean)
        .sort((left, right) => left.artifact.localeCompare(right.artifact));
      return {
        target,
        latest: phaseSummaries.at(-1)?.artifact ?? "",
        phase_summaries: phaseSummaries,
      };
    })
    .filter((entry) => entry.phase_summaries.length > 0);
}
