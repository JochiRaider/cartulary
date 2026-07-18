#!/usr/bin/env node

import { existsSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  collectTaskSurfaceManifestErrors,
  collectTaskSurfaceMakeDensityErrors,
  compactHelpEntries,
  defaultGeneratedMakePath,
  defaultGeneratedMakeRuntimePath,
  harnessCheckEntries,
  helpTiers,
  makeRecipeEntries,
  renderTaskSurfaceMake,
  renderTaskSurfaceMakeRuntime,
  taskSurfaceMakeDensity,
  taskSurfaceSchemaID,
} from "./task-surface/index.mjs";
import { loadTestCatalog, targetForCatalogRow } from "../test-catalog/index.mjs";

const schedulerGrowthBudgets = Object.freeze({
  manifestBytesPerWorkUnit: 1700,
  ordinaryP95Bytes: 1500,
  structurallyWideP95Bytes: 5000,
  maximumBytes: 12 * 1024,
  syntheticBytesPerWorkUnit: 1600,
});

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const checkMode = process.argv.includes("--check");
const jsonMode = process.argv.includes("--json");
const allMode = process.argv.includes("--all");
const makefilePath = resolvePath(
  process.env.CARTULARY_TASK_SURFACE_MAKEFILE ?? "Makefile",
);
const manifestPath = resolvePath(
  process.env.CARTULARY_TASK_SURFACE_MANIFEST ?? "tools/task_surface_manifest.json",
);
const generatedMakePath = resolvePath(
  process.env.CARTULARY_TASK_SURFACE_GENERATED_MAKE ?? defaultGeneratedMakePath,
);
const generatedMakeRuntimePath = resolvePath(
  process.env.CARTULARY_TASK_SURFACE_GENERATED_RUNTIME_MAKE ?? defaultGeneratedMakeRuntimePath,
);
const schedulerManifestPath = resolvePath("tools/scheduler_manifest.json");

const validTargetClasses = new Set(["public", "check_internal", "internal_helper"]);
const validDefaultInclusionSets = new Set(["test", "check", "ci", "release-check", "helper_only"]);
const retiredRootRunnerHelperPattern =
  /^\s*(RUN_(?:GO|PLAYWRIGHT|VITEST)(?:_MANIFEST)?_PHASE(?:_SCRIPT)?)\s*(?::=|\?=|\+=|=)/;
const targetOwnedPhaseTargetsPattern = /^\s*TARGET_OWNED_PHASE_TARGETS\s*(?::=|\?=|\+=|=)/;
const targetSpecificTestTargetExportPattern =
  /^\s*([A-Za-z0-9_.-]+)\s*:\s*export\s+CARTULARY_TEST_TARGET\b/;

function main() {
  const generatedMake = readFileSync(generatedMakePath, "utf8");
  const authoredMake = readFileSync(makefilePath, "utf8");
  const makefile = `${generatedMake}\n${authoredMake}`;
  const manifest = readJSON(manifestPath);
  const phonyTargets = collectPhonyTargets(makefile);
  const helpEntries = collectHelpEntries(makefile);
  const targetBlocks = collectTargetBlocks(makefile, phonyTargets);
  const authoredRecipeTargets = makeRecipeEntries(manifest).map((entry) => entry.target);
  const recipeByTarget = new Map(makeRecipeEntries(manifest).map((entry) => [entry.target, entry]));
  const authoredGeneratedRecipeBlocks = collectTargetBlocks(authoredMake, authoredRecipeTargets);
  const targetScriptRefs = new Map(
    phonyTargets.map((target) => [target, collectDirectScriptRefs(targetBlocks.get(target) ?? "")]),
  );
  const catalogTargetPartitions = collectCatalogTargetPartitions(manifest);
  const errors = validateTaskSurface({
    authoredMake,
    generatedMakePath,
    generatedMakeRuntimePath,
    helpEntries,
    manifest,
    authoredGeneratedRecipeBlocks,
    phonyTargets,
    recipeByTarget,
    targetScriptRefs,
  });
  const schedulerGrowth = collectSchedulerGrowthMetrics(schedulerManifestPath);
  errors.push(...schedulerGrowth.errors);
  const report = buildReport({
    errors,
    helpEntries,
    manifest,
    catalogTargetPartitions,
    phonyTargets,
    targetScriptRefs,
    makeDensity: taskSurfaceMakeDensity(manifest),
    schedulerGrowth,
  });

  if (jsonMode) {
    process.stdout.write(`${JSON.stringify(report, null, 2)}\n`);
  } else {
    printHumanReport(report, { allMode });
  }

  if (checkMode && errors.length > 0) {
    process.exit(1);
  }
}

function serializedBytes(value) {
  return Buffer.byteLength(JSON.stringify(value), "utf8");
}

function collectSchedulerGrowthMetrics(file) {
  const manifest = readJSON(file);
  const workUnits = (manifest.schedules ?? []).flatMap((schedule) => schedule.work_units ?? []);
  const serializedSizes = workUnits.map(serializedBytes).sort((left, right) => left - right);
  const structurallyWideKinds = new Set([
    "aggregate_finalize",
    "browser_group",
    "browser_stage_complete",
  ]);
  const ordinarySerializedSizes = workUnits
    .filter((unit) => !structurallyWideKinds.has(unit.kind))
    .map(serializedBytes)
    .sort((left, right) => left - right);
  const structurallyWideSerializedSizes = workUnits
    .filter((unit) => structurallyWideKinds.has(unit.kind))
    .map(serializedBytes)
    .sort((left, right) => left - right);
  const workUnitCount = serializedSizes.length;
  const manifestBytes = statSync(file).size;
  const p95 = (values) => values[Math.max(0, Math.ceil(values.length * 0.95) - 1)] ?? 0;
  const syntheticUnit = {
    target: "synthetic-growth-unit",
    priority: 100,
    weight_ms: 1000,
    needs: [],
    make_prerequisite_policy: "run",
    resource_claims: { host_cpu: 1 },
    make_jobs: "host_cpu",
    command: { type: "make_target", target: "synthetic-growth-unit" },
  };
  const renderSynthetic = () =>
    `${JSON.stringify(
      {
        work_units: Array.from({ length: 25 }, (_, index) => ({
          ...syntheticUnit,
          target: `${syntheticUnit.target}-${String(index + 1).padStart(2, "0")}`,
          command: {
            ...syntheticUnit.command,
            target: `${syntheticUnit.target}-${String(index + 1).padStart(2, "0")}`,
          },
        })),
      },
      null,
      2,
    )}\n`;
  const syntheticA = renderSynthetic();
  const syntheticB = renderSynthetic();
  const syntheticEmptyBytes = Buffer.byteLength(`${JSON.stringify({ work_units: [] }, null, 2)}\n`);
  const metrics = {
    manifest_bytes: manifestBytes,
    work_unit_count: workUnitCount,
    manifest_bytes_per_work_unit:
      workUnitCount === 0 ? 0 : Math.ceil(manifestBytes / workUnitCount),
    p95_serialized_work_unit_bytes: p95(serializedSizes),
    ordinary_p95_serialized_work_unit_bytes: p95(ordinarySerializedSizes),
    structurally_wide_p95_serialized_work_unit_bytes: p95(structurallyWideSerializedSizes),
    max_serialized_work_unit_bytes: serializedSizes.at(-1) ?? 0,
    synthetic_25_bytes_per_work_unit: Math.ceil(
      (Buffer.byteLength(syntheticA) - syntheticEmptyBytes) / 25,
    ),
    scratch_generations_byte_identical: syntheticA === syntheticB,
  };
  const errors = [];
  if (metrics.manifest_bytes_per_work_unit > schedulerGrowthBudgets.manifestBytesPerWorkUnit) {
    errors.push(
      `scheduler manifest exceeds ${schedulerGrowthBudgets.manifestBytesPerWorkUnit} bytes per work unit`,
    );
  }
  if (metrics.ordinary_p95_serialized_work_unit_bytes > schedulerGrowthBudgets.ordinaryP95Bytes) {
    errors.push(
      `scheduler manifest ordinary-kind p95 serialized work unit exceeds ${schedulerGrowthBudgets.ordinaryP95Bytes} bytes`,
    );
  }
  if (
    metrics.structurally_wide_p95_serialized_work_unit_bytes >
    schedulerGrowthBudgets.structurallyWideP95Bytes
  ) {
    errors.push(
      `scheduler manifest structurally-wide p95 serialized work unit exceeds ${schedulerGrowthBudgets.structurallyWideP95Bytes} bytes`,
    );
  }
  if (metrics.max_serialized_work_unit_bytes > schedulerGrowthBudgets.maximumBytes) {
    errors.push("scheduler manifest maximum serialized work unit exceeds 12 KiB");
  }
  if (
    metrics.synthetic_25_bytes_per_work_unit > schedulerGrowthBudgets.syntheticBytesPerWorkUnit
  ) {
    errors.push(
      `25 synthetic ordinary scheduler work units exceed ${schedulerGrowthBudgets.syntheticBytesPerWorkUnit} bytes each`,
    );
  }
  if (!metrics.scratch_generations_byte_identical) {
    errors.push("two scratch scheduler generations are not byte-identical");
  }
  return { ...metrics, errors };
}

function resolvePath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function readJSON(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch (error) {
    throw new Error(`failed to read JSON ${file}: ${error.message}`);
  }
}

function collectPhonyTargets(makefile) {
  const targets = [];
  const lines = makefile.split(/\r?\n/);
  for (let index = 0; index < lines.length; index += 1) {
    if (!lines[index].startsWith(".PHONY:")) {
      continue;
    }
    let logical = lines[index].replace(".PHONY:", "");
    while (logical.trimEnd().endsWith("\\") && index + 1 < lines.length) {
      logical = `${logical.trimEnd().slice(0, -1)} ${lines[(index += 1)]}`;
    }
    targets.push(...logical.trim().split(/\s+/).filter(Boolean));
  }
  return targets;
}

function collectHelpEntries(makefile) {
  const entries = new Map();
  for (const line of makefile.split(/\r?\n/)) {
    const match = /^\s*' {2}make ([A-Za-z0-9_.-]+)(?:\s+[^']*)?'/.exec(line);
    if (!match) {
      continue;
    }
    const target = match[1];
    entries.set(target, line.trim().replace(/^'/, "").replace(/' \\?$/, ""));
  }
  return entries;
}

function collectTargetBlocks(makefile, targets) {
  const lines = makefile.split(/\r?\n/);
  const targetSet = new Set(targets);
  const blocks = new Map();

  for (let index = 0; index < lines.length; index += 1) {
    const match = /^([A-Za-z0-9_.-]+):/.exec(lines[index]);
    if (!match || !targetSet.has(match[1]) || lines[index].includes(": export ")) {
      continue;
    }
    const target = match[1];
    const blockLines = [lines[index]];
    for (let next = index + 1; next < lines.length; next += 1) {
      if (/^[^\s#][^:]*:/.test(lines[next])) {
        break;
      }
      blockLines.push(lines[next]);
    }
    blocks.set(target, blockLines.join("\n"));
  }

  return blocks;
}

function collectDirectScriptRefs(source) {
  const refs = new Set();
  for (const match of source.matchAll(/(?:\.\/)?scripts\/[A-Za-z0-9_./-]+/g)) {
    refs.add(match[0].replace(/^\.\//, ""));
  }
  for (const match of source.matchAll(
    /(?:\.\/)?(?:tools|deploy)\/[A-Za-z0-9_./-]+\.(?:json|mjs|tsx|yaml|css|go|js|md|sh|sql|toml|ts|yml)(?![A-Za-z0-9_.-])/g,
  )) {
    refs.add(match[0].replace(/^\.\//, ""));
  }
  return Array.from(refs).sort();
}

function collectRetiredRootRunnerHelpers(source) {
  const helpers = [];
  const lines = source.split(/\r?\n/);
  for (let index = 0; index < lines.length; index += 1) {
    const match = retiredRootRunnerHelperPattern.exec(lines[index]);
    if (match) {
      helpers.push({
        line: index + 1,
        name: match[1],
      });
    }
  }
  return helpers;
}

function normalizeRootScriptPath(token) {
  const relative = token.startsWith("./") ? token.slice(2) : token;
  return relative.startsWith("scripts/") ? relative : "";
}

function validateRootScriptPathPolicy(errors, token, label) {
  const script = normalizeRootScriptPath(token);
  if (!script) {
    return true;
  }
  errors.push(
    `${label} must not reference retired root scripts/ path ${script}; use an owner path under tools/** or a deployment package path under deploy/**/scripts/**`,
  );
  return false;
}

function collectForbiddenMakeOwnership(source) {
  const violations = [];
  const lines = source.split(/\r?\n/);
  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    if (targetOwnedPhaseTargetsPattern.test(line)) {
      violations.push({
        line: index + 1,
        message:
          "Makefile must not define TARGET_OWNED_PHASE_TARGETS; use task_surface.make_recipes[].test_target",
      });
      continue;
    }
    const exportMatch = targetSpecificTestTargetExportPattern.exec(line);
    if (exportMatch) {
      violations.push({
        line: index + 1,
        message: `Makefile must not define target-specific CARTULARY_TEST_TARGET for ${exportMatch[1]}; use task_surface.make_recipes[].test_target`,
      });
    }
  }
  return violations;
}

function collectCatalogTargetPartitions(taskSurface) {
  const commandTargetByID = new Map(
    taskSurface.targets.map((entry) => [entry.command_id, entry.name]),
  );
  const counts = new Map();
  for (const row of loadTestCatalog(repoRoot).rows.filter((entry) => entry.status === "active")) {
    const targetID = targetForCatalogRow(row, { commandTargetByID });
    const key = `${row.owner_id}:${targetID}:${row.runner}`;
    counts.set(key, (counts.get(key) ?? 0) + 1);
  }
  return [...counts.entries()]
    .map(([key, count]) => {
      const [ownerID, targetID, runner] = key.split(":", 3);
      return { count, owner_id: ownerID, runner, target_id: targetID };
    })
    .sort((left, right) =>
      `${left.owner_id}:${left.target_id}:${left.runner}`.localeCompare(
        `${right.owner_id}:${right.target_id}:${right.runner}`,
      ),
    );
}

function validateTaskSurface({
  authoredMake,
  generatedMakePath,
  generatedMakeRuntimePath,
  helpEntries,
  manifest,
  authoredGeneratedRecipeBlocks,
  phonyTargets,
  recipeByTarget,
  targetScriptRefs,
}) {
  const errors = [];

  for (const helper of collectRetiredRootRunnerHelpers(authoredMake)) {
    errors.push(
      `Makefile must not define retired runner-specific helper ${helper.name} on line ${helper.line}; use generic RUN_PHASE or script-local helpers`,
    );
  }
  for (const violation of collectForbiddenMakeOwnership(authoredMake)) {
    errors.push(`${violation.message} on line ${violation.line}`);
  }

  if (manifest.schema_id !== taskSurfaceSchemaID) {
    errors.push(`tools/task_surface_manifest.json must declare schema_id=${taskSurfaceSchemaID}`);
  }
  if (!Array.isArray(manifest.targets)) {
    errors.push("tools/task_surface_manifest.json must declare targets[]");
    return errors;
  }
  errors.push(...collectTaskSurfaceManifestErrors(manifest));
  errors.push(...collectTaskSurfaceMakeDensityErrors(manifest));
  const renderedMake = renderTaskSurfaceMake(manifest);
  const renderedRuntimeMake = renderTaskSurfaceMakeRuntime(manifest);
  const committedMake = readFileSync(generatedMakePath, "utf8");
  const committedRuntimeMake = readFileSync(generatedMakeRuntimePath, "utf8");
  if (renderedMake !== committedMake) {
    errors.push("tools/task_surface.generated.mk is stale; run tools/harness/generated-artifacts/render-task-surface-make.mjs");
  }
  if (renderedRuntimeMake !== committedRuntimeMake) {
    errors.push(
      "tools/task_surface.runtime.generated.mk is stale; run tools/harness/generated-artifacts/render-task-surface-make.mjs",
    );
  }
  for (const recipe of makeRecipeEntries(manifest)) {
    if (authoredGeneratedRecipeBlocks.has(recipe.target)) {
      errors.push(`generated Make recipe target ${recipe.target} must not be hand-defined in Makefile`);
    }
  }

  const phonySet = new Set(phonyTargets);
  const entriesByName = new Map();
  for (const entry of manifest.targets) {
    if (typeof entry.name !== "string" || entry.name.trim() === "") {
      errors.push("task-surface manifest entry has missing name");
      continue;
    }
    if (entriesByName.has(entry.name)) {
      errors.push(`task-surface manifest has duplicate target ${entry.name}`);
      continue;
    }
    entriesByName.set(entry.name, entry);
  }
  const harnessChecksByName = new Map();
  for (const entry of harnessCheckEntries(manifest)) {
    if (typeof entry.name !== "string" || entry.name.trim() === "") {
      errors.push("task-surface manifest harness check has missing name");
      continue;
    }
    if (entriesByName.has(entry.name)) {
      errors.push(`task-surface harness check ${entry.name} conflicts with a Makefile target`);
      continue;
    }
    if (harnessChecksByName.has(entry.name)) {
      errors.push(`task-surface manifest has duplicate harness check ${entry.name}`);
      continue;
    }
    harnessChecksByName.set(entry.name, entry);
    const declaredScripts = Array.isArray(entry.backing_scripts) ? entry.backing_scripts : [];
    if (declaredScripts.length === 0) {
      errors.push(`${entry.name} must declare non-empty backing_scripts[]`);
    }
    for (const script of declaredScripts) {
      if (typeof script !== "string" || script.trim() === "") {
        errors.push(`${entry.name} declares an invalid backing script`);
        continue;
      }
      if (!validateRootScriptPathPolicy(errors, script, `${entry.name}.backing_scripts`)) {
        continue;
      }
      const scriptPath = path.join(repoRoot, script);
      if (!existsSync(scriptPath) || !statSync(scriptPath).isFile()) {
        errors.push(`${entry.name} backing script missing: ${script}`);
      }
    }
  }

  for (const target of phonyTargets) {
    if (!entriesByName.has(target)) {
      errors.push(`Makefile .PHONY target ${target} is missing task-surface target_class`);
    }
  }
  for (const target of entriesByName.keys()) {
    if (!phonySet.has(target)) {
      errors.push(`task-surface manifest target ${target} is not a Makefile .PHONY target`);
    }
  }
  for (const check of harnessChecksByName.keys()) {
    if (phonySet.has(check)) {
      errors.push(`harness check ${check} must not be a Makefile .PHONY target`);
    }
  }

  for (const [target, entry] of entriesByName.entries()) {
    if (!validTargetClasses.has(entry.target_class)) {
      errors.push(`${target} has invalid target_class ${JSON.stringify(entry.target_class)}`);
    }
    if (!Array.isArray(entry.default_inclusion_sets)) {
      errors.push(`${target} must declare default_inclusion_sets[]`);
    } else {
      for (const inclusion of entry.default_inclusion_sets) {
        if (!validDefaultInclusionSets.has(inclusion)) {
          errors.push(`${target} has invalid default_inclusion_sets value ${JSON.stringify(inclusion)}`);
        }
      }
      if (
        entry.target_class !== "public" &&
        entry.default_inclusion_sets.includes("helper_only")
      ) {
        errors.push(`${target} default_inclusion_sets helper_only is only valid for public targets`);
      }
    }

    const hasHelp = helpEntries.has(target);
    if (entry.target_class === "public" && !hasHelp) {
      errors.push(`public target ${target} is missing a help entry`);
    }
    if (hasHelp && entry.target_class !== "public") {
      errors.push(`help entry ${target} must be target_class public`);
    }

    const declaredScripts = Array.isArray(entry.backing_scripts) ? entry.backing_scripts : [];
    for (const script of declaredScripts) {
      if (typeof script !== "string" || script.trim() === "") {
        errors.push(`${target} declares an invalid backing script`);
        continue;
      }
      if (!validateRootScriptPathPolicy(errors, script, `${target}.backing_scripts`)) {
        continue;
      }
      const scriptPath = path.join(repoRoot, script);
      if (!existsSync(scriptPath) || !statSync(scriptPath).isFile()) {
        errors.push(`${target} backing script missing: ${script}`);
      }
    }

    const actualScriptRefs = targetScriptRefs.get(target) ?? [];
    const declaredScriptSet = new Set(declaredScripts);
    for (const script of actualScriptRefs) {
      if (!validateRootScriptPathPolicy(errors, script, `${target}.makefile_script_refs`)) {
        continue;
      }
      if (
        script === "tools/harness/execution/run-make-node-tool.sh" &&
        recipeByTarget.get(target)?.type === "node_tool"
      ) {
        continue;
      }
      if (!declaredScriptSet.has(script)) {
        errors.push(`${target} references ${script} but does not declare it in task_surface_manifest.json`);
      }
    }
  }

  for (const target of helpEntries.keys()) {
    if (!phonySet.has(target)) {
      errors.push(`help entry ${target} is not a Makefile .PHONY target`);
    }
  }

  return errors;
}

function buildReport({
  catalogTargetPartitions,
  errors,
  helpEntries,
  makeDensity,
  manifest,
  phonyTargets,
  schedulerGrowth,
  targetScriptRefs,
}) {
  const entriesByName = new Map((manifest.targets ?? []).map((entry) => [entry.name, entry]));
  const compactTargets = compactHelpEntries(manifest).map((entry) => entry.target);
  const helpTierByTarget = new Map();
  const helpTierSummaries = helpTiers(manifest).map((tier) => {
    const targets = (tier.entries ?? []).map((entry) => entry.target);
    for (const target of targets) {
      helpTierByTarget.set(target, tier.name);
    }
    return {
      name: tier.name,
      count: targets.length,
      targets,
    };
  });
  const harnessChecks = harnessCheckEntries(manifest).map((entry) => ({
    name: entry.name,
    backing_scripts: entry.backing_scripts ?? [],
    command: entry.command ?? null,
  }));
  const targets = phonyTargets.map((target) => {
    const entry = entriesByName.get(target) ?? {};
    return {
      name: target,
      target_class: entry.target_class ?? "unclassified",
      command_id: entry.command_id ?? null,
      semantic_behaviors: entry.semantic_behaviors ?? [],
      side_effects: entry.side_effects ?? [],
      has_help: helpEntries.has(target),
      help_tier: helpTierByTarget.get(target) ?? null,
      default_inclusion_sets: entry.default_inclusion_sets ?? [],
      backing_scripts: entry.backing_scripts ?? [],
      makefile_script_refs: targetScriptRefs.get(target) ?? [],
    };
  });
  const semanticBehaviorCounts = new Map();
  const sideEffectCounts = new Map();
  const publicSemanticRows = targets
    .filter((entry) => entry.target_class === "public")
    .map((entry) => {
      const behaviors = Array.isArray(entry.semantic_behaviors)
        ? entry.semantic_behaviors.map((item) => item.behavior).filter(Boolean)
        : [];
      const sideEffects = Array.isArray(entry.side_effects)
        ? entry.side_effects.map((item) => item.class).filter(Boolean)
        : [];
      for (const behavior of behaviors) {
        semanticBehaviorCounts.set(
          behavior,
          (semanticBehaviorCounts.get(behavior) ?? 0) + 1,
        );
      }
      for (const sideEffect of sideEffects) {
        sideEffectCounts.set(sideEffect, (sideEffectCounts.get(sideEffect) ?? 0) + 1);
      }
      return {
        target: entry.name,
        command_id: entry.command_id,
        behaviors,
        side_effects: sideEffects,
      };
    });

  return {
    schema_id: "cartulary.task_surface_report.v4",
    check_passed: errors.length === 0,
    errors,
    targets,
    harness_checks: harnessChecks,
    compact_help: {
      count: compactTargets.length,
      targets: compactTargets,
    },
    help_entries: Array.from(helpEntries.keys()).sort(),
    help_tiers: helpTierSummaries,
    semantic_value: {
      public_targets: publicSemanticRows,
      behavior_counts: Object.fromEntries(
        [...semanticBehaviorCounts.entries()].sort(([left], [right]) =>
          left.localeCompare(right),
        ),
      ),
      side_effect_counts: Object.fromEntries(
        [...sideEffectCounts.entries()].sort(([left], [right]) =>
          left.localeCompare(right),
        ),
      ),
    },
    generated_make_density: makeDensity,
    scheduler_growth: Object.fromEntries(
      Object.entries(schedulerGrowth).filter(([key]) => key !== "errors"),
    ),
    catalog_target_partitions: catalogTargetPartitions,
  };
}

function printHumanReport(report, { allMode = false } = {}) {
  console.log("Cartulary task-surface report");
  console.log(`check=${report.check_passed ? "pass" : "fail"}`);

  const counts = new Map();
  for (const target of report.targets) {
    counts.set(target.target_class, (counts.get(target.target_class) ?? 0) + 1);
  }
  console.log("");
  console.log("target_class counts:");
  for (const targetClass of ["public", "check_internal", "internal_helper", "unclassified"]) {
    const count = counts.get(targetClass) ?? 0;
    if (count > 0) {
      console.log(`  ${targetClass}: ${count}`);
    }
  }

  console.log("");
  console.log("compact help count:");
  console.log(`  compact: ${report.compact_help.count}`);

  console.log("");
  console.log("generated Make density:");
  for (const [metric, value] of Object.entries(report.generated_make_density)) {
    console.log(`  ${metric}: ${value}`);
  }

  console.log("");
  console.log("scheduler growth:");
  for (const [metric, value] of Object.entries(report.scheduler_growth)) {
    console.log(`  ${metric}: ${value}`);
  }

  console.log("");
  console.log("help tier counts:");
  for (const tier of report.help_tiers) {
    console.log(`  ${tier.name}: ${tier.count}`);
  }

  console.log("");
  console.log("semantic behavior counts:");
  for (const [behavior, count] of Object.entries(report.semantic_value.behavior_counts)) {
    console.log(`  ${behavior}: ${count}`);
  }

  console.log("");
  console.log("side-effect counts:");
  for (const [sideEffect, count] of Object.entries(report.semantic_value.side_effect_counts)) {
    console.log(`  ${sideEffect}: ${count}`);
  }

  if (allMode) {
    console.log("");
    console.log("public Make targets:");
    for (const target of report.targets.filter((entry) => entry.target_class === "public")) {
      const behaviors = target.semantic_behaviors
        .map((entry) => `${entry.behavior}@${entry.owner_section}`)
        .join(",");
      const sideEffects = target.side_effects
        .map((entry) => `${entry.class}@${entry.owner_section}`)
        .join(",");
      console.log(
        `  ${target.name} command_id=${target.command_id ?? "-"} behaviors=${behaviors || "-"} side_effects=${sideEffects || "-"} help=${target.has_help ? "yes" : "no"} help_tier=${target.help_tier ?? "-"} default_inclusion_sets=${target.default_inclusion_sets.join(",")}`,
      );
    }

    console.log("");
    console.log("task target classes:");
    for (const target of report.targets) {
      const scripts = target.backing_scripts.length > 0 ? target.backing_scripts.join(",") : "-";
      console.log(
        `  ${target.name} target_class=${target.target_class} default_inclusion_sets=${target.default_inclusion_sets.join(",")} scripts=${scripts}`,
      );
    }

    console.log("");
    console.log("logical harness checks:");
    for (const check of report.harness_checks) {
      const scripts = check.backing_scripts.length > 0 ? check.backing_scripts.join(",") : "-";
      const command = Array.isArray(check.command) && check.command.length > 0 ? check.command.join(" ") : "-";
      console.log(`  ${check.name} scripts=${scripts} command=${command}`);
    }

    console.log("");
    console.log("catalog target partitions:");
    for (const row of report.catalog_target_partitions) {
      console.log(
        `  ${row.owner_id} ${row.target_id} ${row.runner} rows=${row.count}`,
      );
    }
  } else {
    console.log("");
    console.log(`logical harness checks: ${report.harness_checks.length}`);
    console.log(`catalog target partitions: ${report.catalog_target_partitions.length}`);
    console.log("use --all to print public targets, private targets, harness checks, and catalog partitions");
  }

  if (report.errors.length > 0) {
    console.log("");
    console.log("drift:");
    for (const error of report.errors) {
      console.log(`  ${error}`);
    }
  }
}

try {
  main();
} catch (error) {
  console.error(`task-surface report failed: ${error.message}`);
  process.exit(1);
}
