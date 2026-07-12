import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..", "..");
export const defaultTaskSurfaceManifestPath = path.join(repoRoot, "tools", "task_surface_manifest.json");
export const defaultGeneratedMakePath = path.join(repoRoot, "tools", "task_surface.generated.mk");
export const defaultGeneratedMakeRuntimePath = path.join(
  repoRoot,
  "tools",
  "task_surface.runtime.generated.mk",
);
export const taskSurfaceSchemaID = "cartulary.task_surface_manifest.v15";

export const restrictedInternalMakeVariables = Object.freeze([
  "CARTULARY_OPERATOR_BIN",
  "CARTULARY_EXECUTION_TOPOLOGY_MANIFEST",
  "CARTULARY_TASK_SURFACE_MANIFEST",
  "EXECUTION_TOPOLOGY_MANIFEST",
  "SCHEDULER_MANIFEST",
  "TASK_SURFACE_MANIFEST",
]);
export const canonicalInternalMakeValues = Object.freeze({
  EXECUTION_TOPOLOGY_MANIFEST: "$(TASK_SURFACE_CANONICAL_EXECUTION_TOPOLOGY_MANIFEST)",
  SCHEDULER_MANIFEST: "$(TASK_SURFACE_CANONICAL_SCHEDULER_MANIFEST)",
  TASK_SURFACE_MANIFEST: "$(TASK_SURFACE_CANONICAL_TASK_SURFACE_MANIFEST)",
});
export const nonCanonicalPublicMakeVariables = Object.freeze([
  "GOVULNCHECK_FLAGS",
  "GOVULNCHECK_PATTERNS",
  "GOSEC_AUDIT_RUNTIME_FLAGS",
  "GOSEC_AUDIT_RUNTIME_PATTERNS",
  "GOSEC_AUDIT_RUNTIME_RULES",
  "GOSEC_AUDIT_SUPPORT_FLAGS",
  "GOSEC_AUDIT_SUPPORT_PATTERNS",
  "GOSEC_AUDIT_SUPPORT_RULES",
  "GOSEC_FLAGS",
  "GOSEC_PATTERNS",
  "GOSEC_RULES",
  "GOSEC_TARGETED_RUNTIME_FLAGS",
  "GOSEC_TARGETED_RUNTIME_PATTERNS",
  "GOSEC_TARGETED_RUNTIME_RULES",
  "STATICCHECK_CHECKS",
  "VITEST_FLAGS",
]);

export function resolveRepoPath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

export function readJSON(file) {
  return JSON.parse(readFileSync(resolveRepoPath(file), "utf8"));
}

export function targetEntries(manifest) {
  return manifest.targets ?? [];
}

export function targetEntryMap(manifest) {
  return new Map(targetEntries(manifest).map((entry) => [entry.name, entry]));
}

export function harnessCheckEntries(manifest) {
  return manifest.harness_checks ?? [];
}

export function harnessCheckEntryMap(manifest) {
  return new Map(
    harnessCheckEntries(manifest).map((entry) => [entry.name, entry]),
  );
}

export function helpTiers(manifest) {
  return manifest.help_tiers ?? [];
}

export function compactHelpEntries(manifest) {
  return manifest.compact_help?.entries ?? [];
}

export function summaryEntryMap(manifest) {
  return new Map([
    ...targetEntryMap(manifest),
    ...harnessCheckEntryMap(manifest),
  ]);
}

export function harnessTierChecks(manifest, name) {
  const tier = manifest.harness_tiers?.[name];
  if (!tier) {
    throw new Error(`unknown harness tier ${name}`);
  }
  return [...tier.checks];
}

export function harnessCheck(manifest, name) {
  const check = harnessCheckEntryMap(manifest).get(name);
  if (!check) {
    throw new Error(`unknown harness check ${name}`);
  }
  return {
    name: check.name,
    backing_scripts: [...check.backing_scripts],
    command: check.command === undefined ? null : [...check.command],
  };
}

export function sequenceDefinition(manifest, name) {
  const sequence = manifest.sequences?.[name];
  if (!sequence) {
    throw new Error(`unknown task-surface sequence ${name}`);
  }
  return {
    name,
    summaryGroups: sequence.summary_groups ?? [],
    steps: sequence.steps.map((step) => ({
      type: step.type,
      target: step.target,
      jobs: step.jobs,
      jobsVariable: step.jobs_variable,
      skipPrerequisites: step.skip_prerequisites === true,
      producesSummaryTargets: [...(step.produces_summary_targets ?? [])],
    })),
  };
}

export function makeRecipeEntries(manifest) {
  return Object.entries(manifest.make_recipes ?? {}).map(
    ([target, recipe]) => ({ target, ...recipe }),
  );
}

export function makeIdentifier(value) {
  return value.replace(/[^A-Za-z0-9_]/g, "_").toUpperCase();
}
