import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  hasMakeNodeTool,
  makeNodeToolMakeEnvVars,
} from "./make-node-tools.mjs";
import { resourceOverrideEnvVariablesForScheduler } from "./scheduler-resources.mjs";
import {
  collectExplicitSummaryProjectionErrors,
  loadSummaryTopologyContext,
  resolveSummaryGroups,
} from "./summary-topology.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..");
export const defaultTaskSurfaceManifestPath = path.join(
  repoRoot,
  "tools",
  "task_surface_manifest.json",
);
export const defaultGeneratedMakePath = path.join(
  repoRoot,
  "tools",
  "task_surface.generated.mk",
);
export const taskSurfaceSchemaID = "cartulary.task_surface_manifest.v12";

const validTargetClasses = new Set([
  "public",
  "check_internal",
  "internal_helper",
]);
const validDefaultInclusionSets = new Set([
  "test",
  "check",
  "ci",
  "release-check",
  "helper_only",
]);
const validFamilyIDs = new Set([
  "help_discovery",
  "bootstrap_toolchain",
  "local_services_dev",
  "generated_drift",
  "phase_service_slices",
  "backend_frontend_leaf_tests",
  "browser_e2e",
  "aggregates_gates",
  "static_analysis_security",
  "builds",
  "cleanup",
  "formatting",
]);
const validLifecycleStates = new Set([
  "candidate_child",
  "public_active",
  "public_deprecated",
]);
const validServiceRequirements = new Set([
  "postgres",
  "minio",
  "browser_stack",
  "vite",
]);
const validOutputClasses = new Set([
  "aggregate_summary_with_artifacts",
  "destructive_human",
  "human_summary",
  "interactive_raw",
  "machine_stdout",
  "scheduler_summary_with_artifacts",
  "service_summary",
  "summary_with_artifacts",
]);
const validArtifactPolicies = new Set([
  "none",
  "run_and_target_summaries",
  "scheduler_and_tool_run_summaries",
  "service_logs_and_summary",
  "tool_run_summary",
]);
const validRawStreamPolicies = new Set([
  "explicit_detail_only",
  "failure_or_verbose",
  "human_visible_actions",
  "interactive",
  "machine_json_stdout",
  "never_default",
]);
export const validSemanticBehaviors = new Set([
  "configuration_resolution",
  "evidence_normalization",
  "failure_normalization",
  "service_lifecycle",
  "scheduler_orchestration",
  "destructive_safety",
  "security_boundary",
  "diagnostic_synthesis",
]);
const validGateSmokeRoles = new Set([
  "public_make_wrapper",
  "check_scheduler_semantic",
  "service_backed_scheduler_semantic",
]);
const standardSuccessBudgetKeys = new Set([
  "stdout_lines",
  "stdout_bytes",
  "stderr_lines",
  "stderr_bytes",
]);
const standardFailureBudgetKeys = new Set([
  "stderr_lines",
  "stderr_bytes",
  "excerpt_lines",
  "excerpt_bytes",
]);
const artifactPolicyNoneOutputClasses = new Set([
  "destructive_human",
  "human_summary",
  "interactive_raw",
  "machine_stdout",
]);
const makeRecipeValidators = Object.freeze({
  alias: validateAliasRecipe,
  cleanup: validateCleanupRecipe,
  print_help: validatePrintHelpRecipe,
  sequence: validateSequenceRecipe,
  check_schedule: validateCheckScheduleRecipe,
  go_target: validateGoTargetRecipe,
  service_backed_target: validateServiceBackedTargetRecipe,
  service_backed_schedule: validateServiceBackedScheduleRecipe,
  browser_batch: validateBrowserBatchRecipe,
  phase_command: validatePhaseCommandRecipe,
  summary_target: validateSummaryTargetRecipe,
  node_tool: validateNodeToolRecipe,
});
const validMakeRecipeTypes = new Set(Object.keys(makeRecipeValidators));
const compactHelpMaxEntries = 12;
const helpTargetColumnWidth = 30;
const helpDescriptionIndent = " ".repeat(
  2 + "make ".length + helpTargetColumnWidth + 1,
);
const makeVariablePattern = /^[A-Z][A-Z0-9_]*$/;
const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const commandIDPattern = /^cartulary\.harness\.command\.[a-z][a-z0-9_]*\.v1$/;
const ownerSectionPattern = /^Section (?:[1-9]|1[0-9])(?:\.[0-9]+)?$/;
const makePrerequisitePattern = /^[A-Za-z0-9_.$()/:,/-]+$/;
const makeValuePattern = /^[A-Za-z0-9_.$()/:,;="' -]+$/;
const makeTokenPattern = /^[A-Za-z0-9_.$()/:,="./ -]+$/;
const repoJSONPathPattern = /^[A-Za-z0-9_./-]+\.json$/;

export function resolveRepoPath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

export function readJSON(file) {
  return JSON.parse(readFileSync(resolveRepoPath(file), "utf8"));
}

export function loadTaskSurfaceManifest(file = defaultTaskSurfaceManifestPath) {
  const manifestPath = resolveRepoPath(file);
  const manifest = readJSON(manifestPath);
  validateTaskSurfaceManifest(manifest, manifestPath);
  return { manifest, manifestPath };
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

function validateTaskSurfaceManifest(manifest, manifestPath) {
  const errors = collectTaskSurfaceManifestErrors(manifest);
  if (errors.length > 0) {
    throw new Error(
      `${manifestPath} is invalid:\n${errors.map((error) => `  - ${error}`).join("\n")}`,
    );
  }
}

export function collectTaskSurfaceManifestErrors(manifest, options = {}) {
  const errors = [];
  if (manifest.schema_id !== taskSurfaceSchemaID) {
    errors.push(`schema_id must be ${taskSurfaceSchemaID}`);
  }
  if (!Array.isArray(manifest.targets) || manifest.targets.length === 0) {
    errors.push("targets[] must be a non-empty array");
    return errors;
  }

  const targets = new Map();
  const commandIDs = new Map();
  for (const [index, entry] of manifest.targets.entries()) {
    const label = `targets[${index + 1}]`;
    if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    if (typeof entry.name !== "string" || entry.name.trim() === "") {
      errors.push(`${label}.name must be a non-empty string`);
      continue;
    }
    if (targets.has(entry.name)) {
      errors.push(`duplicate target ${entry.name}`);
      continue;
    }
    targets.set(entry.name, entry);
    if (Object.hasOwn(entry, "classification")) {
      errors.push(`${entry.name}.classification is obsolete; use target_class`);
    }
    if (Object.hasOwn(entry, "included_in")) {
      errors.push(
        `${entry.name}.included_in is obsolete; use default_inclusion_sets`,
      );
    }
    if (!validTargetClasses.has(entry.target_class)) {
      errors.push(
        `${entry.name} has invalid target_class ${JSON.stringify(entry.target_class)}`,
      );
    }
    if (!Array.isArray(entry.default_inclusion_sets)) {
      errors.push(`${entry.name} must declare default_inclusion_sets[]`);
    } else {
      for (const inclusion of entry.default_inclusion_sets) {
        if (!validDefaultInclusionSets.has(inclusion)) {
          errors.push(
            `${entry.name} has invalid default_inclusion_sets value ${JSON.stringify(inclusion)}`,
          );
        }
      }
      if (
        entry.target_class !== "public" &&
        entry.default_inclusion_sets.includes("helper_only")
      ) {
        errors.push(
          `${entry.name}.default_inclusion_sets helper_only is only valid for public direct-invocation targets`,
        );
      }
    }
    if (!validLifecycleStates.has(entry.lifecycle_state)) {
      errors.push(
        `${entry.name} has invalid lifecycle_state ${JSON.stringify(entry.lifecycle_state)}`,
      );
    } else if (
      entry.target_class === "public" &&
      !["public_active", "public_deprecated"].includes(entry.lifecycle_state)
    ) {
      errors.push(
        `${entry.name}.lifecycle_state must be public_active or public_deprecated for public targets`,
      );
    } else if (
      entry.target_class !== "public" &&
      entry.lifecycle_state !== "candidate_child"
    ) {
      errors.push(
        `${entry.name}.lifecycle_state must be candidate_child for non-public targets`,
      );
    }
    if (entry.target_class === "public") {
      if (typeof entry.family_id !== "string" || entry.family_id.trim() === "") {
        errors.push(`${entry.name}.family_id must be declared for public targets`);
      } else if (!validFamilyIDs.has(entry.family_id)) {
        errors.push(
          `${entry.name}.family_id has invalid value ${JSON.stringify(entry.family_id)}`,
        );
      }
    } else if (entry.family_id !== undefined) {
      errors.push(`${entry.name}.family_id is only valid for public targets`);
    }
    if (entry.backing_scripts !== undefined) {
      if (!Array.isArray(entry.backing_scripts)) {
        errors.push(`${entry.name}.backing_scripts must be an array`);
      } else {
        for (const script of entry.backing_scripts) {
          if (typeof script !== "string" || script.trim() === "") {
            errors.push(`${entry.name} declares an invalid backing script`);
          } else if (!existsSync(path.join(repoRoot, script))) {
            errors.push(`${entry.name} backing script missing: ${script}`);
          }
        }
      }
    }
    if (entry.service_requirements !== undefined) {
      if (!Array.isArray(entry.service_requirements)) {
        errors.push(`${entry.name}.service_requirements must be an array`);
      } else {
        const seenRequirements = new Set();
        for (const [
          requirementIndex,
          requirement,
        ] of entry.service_requirements.entries()) {
          const label = `${entry.name}.service_requirements[${requirementIndex + 1}]`;
          if (typeof requirement !== "string" || requirement.trim() === "") {
            errors.push(`${label} must be a non-empty string`);
            continue;
          }
          if (!validServiceRequirements.has(requirement)) {
            errors.push(
              `${label} has invalid service requirement ${JSON.stringify(requirement)}`,
            );
          }
          if (seenRequirements.has(requirement)) {
            errors.push(
              `${entry.name}.service_requirements contains duplicate ${requirement}`,
            );
          }
          seenRequirements.add(requirement);
        }
      }
    }
    if (entry.target_class === "public") {
      validatePublicCommandIdentity(errors, entry, commandIDs);
      validatePublicSemanticBehaviors(errors, entry);
      validateOutputPolicy(errors, entry);
    }
  }

  const harnessChecks = new Map();
  if (!Array.isArray(manifest.harness_checks)) {
    errors.push("harness_checks[] must be an array");
  } else {
    for (const [index, entry] of manifest.harness_checks.entries()) {
      const label = `harness_checks[${index + 1}]`;
      if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
        errors.push(`${label} must be an object`);
        continue;
      }
      if (typeof entry.name !== "string" || entry.name.trim() === "") {
        errors.push(`${label}.name must be a non-empty string`);
        continue;
      }
      if (targets.has(entry.name)) {
        errors.push(`harness check ${entry.name} conflicts with a Make target`);
        continue;
      }
      if (harnessChecks.has(entry.name)) {
        errors.push(`duplicate harness check ${entry.name}`);
        continue;
      }
      harnessChecks.set(entry.name, entry);
      if (
        !Array.isArray(entry.backing_scripts) ||
        entry.backing_scripts.length === 0
      ) {
        errors.push(`${entry.name}.backing_scripts must be a non-empty array`);
      } else {
        for (const script of entry.backing_scripts) {
          if (typeof script !== "string" || script.trim() === "") {
            errors.push(`${entry.name} declares an invalid backing script`);
          } else if (!existsSync(path.join(repoRoot, script))) {
            errors.push(`${entry.name} backing script missing: ${script}`);
          }
        }
      }
      validateCommandTokens(errors, entry.command, `${entry.name}.command`, {
        required: false,
      });
      if (entry.gate_smoke_role !== undefined) {
        if (
          typeof entry.gate_smoke_role !== "string" ||
          !validGateSmokeRoles.has(entry.gate_smoke_role)
        ) {
          errors.push(
            `${entry.name}.gate_smoke_role must be one of ${[...validGateSmokeRoles].join("|")}`,
          );
        }
      }
    }
  }

  const summaryEntries = new Map([...targets, ...harnessChecks]);
  const topologyContext =
    options.summaryTopologyContext ??
    loadSummaryTopologyContext({
      taskSurfaceManifest: manifest,
      browserBatchManifest: options.browserBatchManifest,
      browserStages: options.browserStages,
      serviceBackedScheduleManifest: options.serviceBackedScheduleManifest,
    });
  errors.push(
    ...collectExplicitSummaryProjectionErrors(manifest, topologyContext),
  );

  validateCompactHelp(errors, targets, manifest.compact_help);

  validateHelpTiers(errors, targets, manifest.help_tiers);

  validateHarnessTiers(errors, harnessChecks, manifest.harness_tiers);

  if (
    !manifest.sequences ||
    typeof manifest.sequences !== "object" ||
    Array.isArray(manifest.sequences)
  ) {
    errors.push("sequences must be an object");
  } else {
    for (const [name, sequence] of Object.entries(manifest.sequences)) {
      if (!targets.has(name)) {
        errors.push(`sequence ${name} does not match a declared target`);
      }
      if (sequence.summary_profile !== undefined) {
        errors.push(
          `sequence ${name} must use summary_groups and step produces_summary_targets, not summary_profile`,
        );
      }
      validateSummaryGroups(
        errors,
        summaryEntries,
        topologyContext,
        sequence.summary_groups,
        `sequence ${name}`,
      );
      if (
        !sequence.steps?.some(
          (step) =>
            Array.isArray(step.produces_summary_targets) &&
            step.produces_summary_targets.length > 0,
        )
      ) {
        errors.push(
          `sequence ${name} must declare step produces_summary_targets`,
        );
      }
      if (!Array.isArray(sequence.steps) || sequence.steps.length === 0) {
        errors.push(`sequence ${name} must declare steps[]`);
        continue;
      }
      for (const [index, step] of sequence.steps.entries()) {
        const label = `sequence ${name} steps[${index + 1}]`;
        if (!["step", "parallel"].includes(step?.type)) {
          errors.push(`${label}.type must be step or parallel`);
        }
        if (typeof step?.target !== "string" || !targets.has(step.target)) {
          errors.push(
            `${label}.target references unknown target ${JSON.stringify(step?.target)}`,
          );
        }
        if (
          step.type === "parallel" &&
          typeof step.jobs_variable !== "string" &&
          !Number.isInteger(step.jobs)
        ) {
          errors.push(
            `${label} parallel step must declare jobs or jobs_variable`,
          );
        }
        validateNamedTargetList(
          errors,
          summaryEntries,
          step.produces_summary_targets,
          `${label}.produces_summary_targets`,
          {
            required: false,
          },
        );
      }
    }
  }

  validateMakeRecipes(
    errors,
    targets,
    manifest.sequences,
    manifest.make_recipes,
  );
  validateOutputPolicyRouting(errors, targets, manifest.make_recipes);

  return errors;
}

function validatePublicCommandIdentity(errors, entry, commandIDs) {
  if (typeof entry.command_id !== "string" || entry.command_id.trim() === "") {
    errors.push(`${entry.name}.command_id must be declared for public targets`);
    return;
  }
  if (!commandIDPattern.test(entry.command_id)) {
    errors.push(
      `${entry.name}.command_id must match cartulary.harness.command.<name>.v1`,
    );
  }
  const previousTarget = commandIDs.get(entry.command_id);
  if (previousTarget) {
    errors.push(
      `${entry.name}.command_id duplicates ${previousTarget}: ${entry.command_id}`,
    );
  } else {
    commandIDs.set(entry.command_id, entry.name);
  }
}

function validatePublicSemanticBehaviors(errors, entry) {
  if (
    !Array.isArray(entry.semantic_behaviors) ||
    entry.semantic_behaviors.length === 0
  ) {
    errors.push(
      `${entry.name}.semantic_behaviors must declare at least one semantic behavior`,
    );
    return;
  }
  const seen = new Set();
  for (const [index, behaviorEntry] of entry.semantic_behaviors.entries()) {
    const label = `${entry.name}.semantic_behaviors[${index + 1}]`;
    if (
      !behaviorEntry ||
      typeof behaviorEntry !== "object" ||
      Array.isArray(behaviorEntry)
    ) {
      errors.push(`${label} must be an object`);
      continue;
    }
    const behavior = behaviorEntry.behavior;
    if (typeof behavior !== "string" || !validSemanticBehaviors.has(behavior)) {
      errors.push(`${label}.behavior must be one of ${[...validSemanticBehaviors].join("|")}`);
    } else if (seen.has(behavior)) {
      errors.push(`${entry.name}.semantic_behaviors contains duplicate ${behavior}`);
    } else {
      seen.add(behavior);
    }
    if (
      typeof behaviorEntry.owner_section !== "string" ||
      !ownerSectionPattern.test(behaviorEntry.owner_section)
    ) {
      errors.push(`${label}.owner_section must be a Section reference`);
    }
  }
}

function validateBudget(errors, budget, label, requiredKeys) {
  if (!budget || typeof budget !== "object" || Array.isArray(budget)) {
    errors.push(`${label} must be an object`);
    return;
  }
  for (const key of requiredKeys) {
    if (!Object.hasOwn(budget, key)) {
      errors.push(`${label}.${key} must be declared`);
    }
  }
  for (const key of Object.keys(budget)) {
    if (!requiredKeys.has(key)) {
      errors.push(`${label}.${key} is not a standard budget key`);
      continue;
    }
    if (!/^[a-z_]+$/u.test(key)) {
      errors.push(`${label}.${key} has invalid key syntax`);
      continue;
    }
    if (!Number.isInteger(budget[key]) || budget[key] < 0) {
      errors.push(`${label}.${key} must be a non-negative integer`);
    }
  }
}

function validateOutputPolicy(errors, entry) {
  const policy = entry.output_policy;
  if (!policy || typeof policy !== "object" || Array.isArray(policy)) {
    errors.push(
      `${entry.name}.output_policy must be declared for public targets`,
    );
    return;
  }
  if (!validOutputClasses.has(policy.output_class)) {
    errors.push(
      `${entry.name}.output_policy.output_class has invalid value ${JSON.stringify(policy.output_class)}`,
    );
  }
  if (!validArtifactPolicies.has(policy.artifact_policy)) {
    errors.push(
      `${entry.name}.output_policy.artifact_policy has invalid value ${JSON.stringify(policy.artifact_policy)}`,
    );
  }
  if (!validRawStreamPolicies.has(policy.raw_stream_policy)) {
    errors.push(
      `${entry.name}.output_policy.raw_stream_policy has invalid value ${JSON.stringify(policy.raw_stream_policy)}`,
    );
  }
  if (
    policy.artifact_policy === "none" &&
    !artifactPolicyNoneOutputClasses.has(policy.output_class)
  ) {
    errors.push(
      `${entry.name}.output_policy.artifact_policy=none is only allowed for help/investigation, interactive, machine-stdout, or destructive targets`,
    );
  }
  if (
    policy.summary_schema !== null &&
    policy.summary_schema !== "cartulary.tool_run_summary.v3"
  ) {
    errors.push(
      `${entry.name}.output_policy.summary_schema must be cartulary.tool_run_summary.v3 or null`,
    );
  }
  if (policy.artifact_policy === "none" && policy.summary_schema !== null) {
    errors.push(
      `${entry.name}.output_policy.summary_schema must be null when artifact_policy is none`,
    );
  }
  if (
    policy.artifact_policy !== "none" &&
    policy.summary_schema !== "cartulary.tool_run_summary.v3"
  ) {
    errors.push(
      `${entry.name}.output_policy.summary_schema must be cartulary.tool_run_summary.v3 when artifact_policy is ${policy.artifact_policy}`,
    );
  }
  validateBudget(
    errors,
    policy.success_budget,
    `${entry.name}.output_policy.success_budget`,
    standardSuccessBudgetKeys,
  );
  validateBudget(
    errors,
    policy.failure_budget,
    `${entry.name}.output_policy.failure_budget`,
    standardFailureBudgetKeys,
  );
}

function recipeCanProduceArtifactPolicy(target, recipe, artifactPolicy) {
  if (artifactPolicy === "none") {
    return true;
  }
  if (!recipe || typeof recipe !== "object") {
    return false;
  }
  if (artifactPolicy === "run_and_target_summaries") {
    return recipe.type === "sequence";
  }
  if (artifactPolicy === "scheduler_and_tool_run_summaries") {
    return (
      recipe.type === "check_schedule" ||
      recipe.type === "service_backed_schedule" ||
      (recipe.type === "node_tool" &&
        ["phase-slice", "service-backed-slice"].includes(target))
    );
  }
  if (artifactPolicy === "tool_run_summary") {
    return (
      recipe.type === "alias" ||
      recipe.type === "node_tool" ||
      recipe.type === "go_target" ||
      recipe.type === "service_backed_target" ||
      recipe.type === "browser_batch" ||
      recipe.type === "summary_target" ||
      recipe.type === "phase_command"
    );
  }
  return false;
}

function validateOutputPolicyRouting(errors, targets, recipes) {
  for (const [target, entry] of targets.entries()) {
    if (entry.target_class !== "public") {
      continue;
    }
    const artifactPolicy = entry.output_policy?.artifact_policy ?? "none";
    if (
      !recipeCanProduceArtifactPolicy(target, recipes?.[target], artifactPolicy)
    ) {
      errors.push(
        `${target}.output_policy.artifact_policy=${artifactPolicy} requires a centralized summary-producing recipe`,
      );
    }
  }
}

function validateMakeRecipes(errors, targets, sequences, recipes) {
  if (!recipes || typeof recipes !== "object" || Array.isArray(recipes)) {
    errors.push("make_recipes must be an object");
    return;
  }

  for (const target of targets.keys()) {
    if (!Object.hasOwn(recipes, target)) {
      errors.push(`make_recipes is missing target ${target}`);
    }
  }

  for (const [target, recipe] of Object.entries(recipes)) {
    const label = `make_recipes.${target}`;
    if (!makeTargetPattern.test(target)) {
      errors.push(`${label} target name must be a Make target identifier`);
    }
    if (!targets.has(target)) {
      errors.push(`${label} references unknown target`);
    }
    if (!recipe || typeof recipe !== "object" || Array.isArray(recipe)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    if (!Array.isArray(recipe.prerequisites)) {
      errors.push(`${label}.prerequisites must be an array`);
    } else {
      for (const prerequisite of recipe.prerequisites) {
        if (
          typeof prerequisite !== "string" ||
          !makePrerequisitePattern.test(prerequisite)
        ) {
          errors.push(
            `${label}.prerequisites contains unsafe prerequisite ${JSON.stringify(prerequisite)}`,
          );
        }
      }
    }
    if (!validMakeRecipeTypes.has(recipe.type)) {
      errors.push(
        `${label}.type must be one of ${Array.from(validMakeRecipeTypes).join(", ")}`,
      );
      continue;
    }
    if (recipe.test_target !== undefined && recipe.test_target !== "self") {
      errors.push(`${label}.test_target must be self when present`);
    }
    if (
      recipe.success_summary !== undefined &&
      typeof recipe.success_summary !== "boolean"
    ) {
      errors.push(`${label}.success_summary must be a boolean`);
    }
    validateMakeExports(errors, recipe.exports, `${label}.exports`);
    if (recipe.exports?.CARTULARY_TEST_TARGET !== undefined) {
      errors.push(
        `${label}.exports.CARTULARY_TEST_TARGET is obsolete; use test_target: "self"`,
      );
    }
    validateMakeComments(errors, recipe.comments, `${label}.comments`);

    makeRecipeValidators[recipe.type]({
      errors,
      target,
      recipe,
      label,
      targets,
      sequences,
      repoRoot,
      readJSON,
      helpers: {
        validateEnvEntries,
        validateCommandTokens,
        validateScriptTokenExists,
        hasMakeNodeTool,
      },
    });
  }
}

function validateAliasRecipe() {}

function validateCleanupRecipe({ errors, recipe, label }) {
  if (!["clean", "distclean"].includes(recipe.scope)) {
    errors.push(`${label}.scope must be clean or distclean`);
  }
}

function validatePrintHelpRecipe({ errors, recipe, label }) {
  if (!["compact", "all"].includes(recipe.scope)) {
    errors.push(`${label}.scope must be compact or all`);
  }
}

function validateSequenceRecipe({ errors, recipe, label, sequences }) {
  if (typeof recipe.sequence !== "string" || !sequences?.[recipe.sequence]) {
    errors.push(
      `${label}.sequence references unknown sequence ${JSON.stringify(recipe.sequence)}`,
    );
  }
}

function validateCheckScheduleRecipe({
  errors,
  recipe,
  label,
  targets,
  repoRoot,
  readJSON,
}) {
  if (typeof recipe.target !== "string" || !targets.has(recipe.target)) {
    errors.push(
      `${label}.target references unknown schedule target ${JSON.stringify(recipe.target)}`,
    );
  }
  if (recipe.summary_profile !== undefined) {
    errors.push(
      `${label}.summary_profile is obsolete; summary targets derive from the check schedule`,
    );
  }
  if (
    typeof recipe.manifest_variable !== "string" ||
    !makeVariablePattern.test(recipe.manifest_variable)
  ) {
    errors.push(`${label}.manifest_variable must be a safe Make variable name`);
  }
  if (
    typeof recipe.schedule_manifest !== "string" ||
    !repoJSONPathPattern.test(recipe.schedule_manifest) ||
    recipe.schedule_manifest.includes("..")
  ) {
    errors.push(`${label}.schedule_manifest must be a safe repo-local JSON path`);
  } else {
    const scheduleManifestPath = path.join(repoRoot, recipe.schedule_manifest);
    if (!existsSync(scheduleManifestPath)) {
      if (recipe.schedule_manifest !== "tools/scheduler_manifest.json") {
        errors.push(
          `${label}.schedule_manifest missing: ${recipe.schedule_manifest}`,
        );
      }
    } else {
      const scheduleManifest = readJSON(recipe.schedule_manifest);
      const scheduleTargets = new Set(
        (scheduleManifest.schedules ?? []).map((schedule) => schedule.target),
      );
      if (!scheduleTargets.has(recipe.target)) {
        errors.push(
          `${label}.target ${recipe.target} is missing from ${recipe.schedule_manifest}`,
        );
      }
    }
  }
  if (recipe.resource_limits !== undefined) {
    errors.push(
      `${label}.resource_limits is obsolete; scheduler capacity overrides come from the resource registry`,
    );
  }
}

function validateGoTargetRecipe({ errors, recipe, label, helpers }) {
  helpers.validateEnvEntries(errors, recipe.env, `${label}.env`);
}

function validateServiceBackedTargetRecipe({ errors, recipe, label, helpers }) {
  helpers.validateEnvEntries(errors, recipe.env, `${label}.env`);
}

function validateServiceBackedScheduleRecipe({ errors, recipe, label }) {
  if (
    typeof recipe.phase_label !== "string" ||
    recipe.phase_label.trim() === ""
  ) {
    errors.push(`${label}.phase_label must be a non-empty string`);
  }
  if (!["test-services"].includes(recipe.service_wrapper)) {
    errors.push(`${label}.service_wrapper must be test-services`);
  }
}

function validateBrowserBatchRecipe({ errors, recipe, label }) {
  if (
    typeof recipe.stage !== "string" ||
    !makeTargetPattern.test(recipe.stage)
  ) {
    errors.push(`${label}.stage must be a safe browser stage name`);
  }
  if (
    typeof recipe.workers !== "string" ||
    !makeValuePattern.test(recipe.workers)
  ) {
    errors.push(`${label}.workers must be a safe Make value`);
  }
  if (!["direct", "test-services"].includes(recipe.service_wrapper)) {
    errors.push(`${label}.service_wrapper must be direct or test-services`);
  }
}

function validatePhaseCommandRecipe({ errors, recipe, label, helpers }) {
  if (!["run_phase", "node", "command"].includes(recipe.mode)) {
    errors.push(`${label}.mode must be run_phase, node, or command`);
  }
  if (recipe.allow_success_log !== undefined) {
    errors.push(`${label}.allow_success_log is obsolete`);
  }
  if (
    recipe.failure_note !== undefined &&
    (typeof recipe.failure_note !== "string" ||
      !makeValuePattern.test(recipe.failure_note) ||
      recipe.failure_note.trim() === "")
  ) {
    errors.push(`${label}.failure_note must be a safe non-empty Make value`);
  }
  helpers.validateEnvEntries(errors, recipe.env, `${label}.env`);
  helpers.validateCommandTokens(errors, recipe.command, `${label}.command`, {
    required: recipe.mode !== "node",
  });
  helpers.validateCommandTokens(errors, recipe.args, `${label}.args`, {
    required: false,
  });
  if (
    recipe.mode === "run_phase" &&
    (typeof recipe.phase_label !== "string" || recipe.phase_label.trim() === "")
  ) {
    errors.push(`${label}.phase_label must be a non-empty string for run_phase`);
  }
  if (recipe.mode === "node") {
    if (
      typeof recipe.script !== "string" ||
      !makeTokenPattern.test(recipe.script)
    ) {
      errors.push(`${label}.script must be a safe node script token`);
    } else {
      helpers.validateScriptTokenExists(
        errors,
        recipe.script,
        `${label}.script`,
      );
    }
  }
}

function validateSummaryTargetRecipe({ errors, recipe, label, targets }) {
  if (typeof recipe.child_target !== "string" || !targets.has(recipe.child_target)) {
    errors.push(
      `${label}.child_target references unknown target ${JSON.stringify(recipe.child_target)}`,
    );
  }
  if (recipe.status !== undefined && !["pass", "fail"].includes(recipe.status)) {
    errors.push(`${label}.status must be pass or fail`);
  }
  if (
    recipe.phase_label !== undefined &&
    (typeof recipe.phase_label !== "string" ||
      !makeValuePattern.test(recipe.phase_label) ||
      recipe.phase_label.trim() === "")
  ) {
    errors.push(`${label}.phase_label must be a safe non-empty Make value`);
  }
  if (
    recipe.projection !== undefined &&
    (typeof recipe.projection !== "string" ||
      !makeTargetPattern.test(recipe.projection))
  ) {
    errors.push(`${label}.projection must be a safe target token`);
  }
}

function validateNodeToolRecipe({ errors, target, label, helpers }) {
  if (!helpers.hasMakeNodeTool(target)) {
    errors.push(`${label} has no scripts/lib/make-node-tools.mjs registry entry`);
  }
}

function validateMakeAssignmentMap(
  errors,
  assignments,
  label,
  { keyDescription },
) {
  if (assignments === undefined) {
    return;
  }
  if (
    !assignments ||
    typeof assignments !== "object" ||
    Array.isArray(assignments)
  ) {
    errors.push(`${label} must be an object`);
    return;
  }
  for (const [name, value] of Object.entries(assignments)) {
    if (!makeVariablePattern.test(name)) {
      errors.push(`${label}.${name} must be a safe ${keyDescription}`);
    }
    if (typeof value !== "string" || !makeValuePattern.test(value)) {
      errors.push(`${label}.${name} must be a safe Make value`);
    }
  }
}

function validateMakeExports(errors, exports, label) {
  validateMakeAssignmentMap(errors, exports, label, {
    keyDescription: "Make variable name",
  });
}

function validateMakeComments(errors, comments, label) {
  if (comments === undefined) {
    return;
  }
  if (!Array.isArray(comments)) {
    errors.push(`${label} must be an array`);
    return;
  }
  for (const [index, comment] of comments.entries()) {
    if (
      typeof comment !== "string" ||
      comment.trim() === "" ||
      comment.includes("\n")
    ) {
      errors.push(`${label}[${index + 1}] must be a single-line comment`);
    }
  }
}

function validateEnvEntries(errors, env, label) {
  validateMakeAssignmentMap(errors, env, label, {
    keyDescription: "environment variable name",
  });
}

function validateCommandTokens(
  errors,
  tokens,
  label,
  { required = true } = {},
) {
  if (tokens === undefined && !required) {
    return;
  }
  if (!Array.isArray(tokens) || (required && tokens.length === 0)) {
    errors.push(`${label} must be a ${required ? "non-empty " : ""}array`);
    return;
  }
  for (const [index, token] of tokens.entries()) {
    if (typeof token !== "string" || !makeTokenPattern.test(token)) {
      errors.push(`${label}[${index + 1}] must be a safe command token`);
      continue;
    }
    validateScriptTokenExists(errors, token, `${label}[${index + 1}]`);
  }
}

function validateScriptTokenExists(errors, token, label) {
  const script = token.startsWith("./scripts/")
    ? token.slice(2)
    : token.startsWith("scripts/")
      ? token
      : "";
  if (script && !existsSync(path.join(repoRoot, script))) {
    errors.push(`${label} references missing script ${script}`);
  }
}

function validateNamedTargetList(
  errors,
  targets,
  targetList,
  label,
  { required = true } = {},
) {
  if (targetList === undefined && !required) {
    return;
  }
  if (!Array.isArray(targetList) || (required && targetList.length === 0)) {
    errors.push(`${label} must be a ${required ? "non-empty " : ""}array`);
    return;
  }
  const seen = new Set();
  for (const target of targetList) {
    if (typeof target !== "string" || target.trim() === "") {
      errors.push(`${label} contains an invalid target`);
      continue;
    }
    if (seen.has(target)) {
      errors.push(`${label} contains duplicate target ${target}`);
    }
    seen.add(target);
    if (!targets.has(target)) {
      errors.push(`${label} references unknown target ${target}`);
    }
  }
}

function validateSummaryGroups(
  errors,
  targets,
  topologyContext,
  groups,
  label,
) {
  if (!Array.isArray(groups)) {
    errors.push(`${label}.summary_groups must be an array`);
    return;
  }
  try {
    for (const group of resolveSummaryGroups(topologyContext, groups)) {
      validateNamedTargetList(
        errors,
        targets,
        group.summaryTargets,
        `${label}.summary_groups.${group.name}.summary_targets`,
      );
    }
  } catch (error) {
    errors.push(`${label}.summary_groups invalid: ${error.message}`);
  }
}

function validateHarnessTiers(errors, harnessChecks, tiers) {
  if (!tiers || typeof tiers !== "object" || Array.isArray(tiers)) {
    errors.push("harness_tiers must be an object");
    return;
  }
  const fastChecks = new Set();
  for (const [name, tier] of Object.entries(tiers)) {
    const checks = tier?.checks;
    if (!Array.isArray(checks) || checks.length === 0) {
      errors.push(`harness_tiers.${name}.checks must be a non-empty array`);
      continue;
    }
    const seen = new Set();
    for (const check of checks) {
      if (typeof check !== "string" || check.trim() === "") {
        errors.push(`harness_tiers.${name}.checks contains an invalid check`);
        continue;
      }
      if (seen.has(check)) {
        errors.push(
          `harness_tiers.${name}.checks contains duplicate check ${check}`,
        );
      }
      seen.add(check);
      if (!harnessChecks.has(check)) {
        errors.push(
          `harness_tiers.${name}.checks references unknown harness check ${check}`,
        );
      }
      if (name === "fast") {
        fastChecks.add(check);
      }
    }
  }
  validateFastHarnessGateSmokeRoles(errors, harnessChecks, fastChecks);
}

function validateFastHarnessGateSmokeRoles(errors, harnessChecks, fastChecks) {
  if (fastChecks.size === 0) {
    errors.push("harness_tiers.fast.checks must declare gate smoke checks");
    return;
  }
  if (fastChecks.size !== validGateSmokeRoles.size) {
    errors.push(
      `harness_tiers.fast.checks must contain exactly ${validGateSmokeRoles.size} gate smoke checks`,
    );
  }
  const roles = new Map();
  for (const check of fastChecks) {
    const entry = harnessChecks.get(check);
    if (!entry) {
      continue;
    }
    const role = entry.gate_smoke_role;
    if (role === undefined) {
      errors.push(`${check}.gate_smoke_role is required for fast harness smoke`);
      continue;
    }
    if (!validGateSmokeRoles.has(role)) {
      continue;
    }
    if (roles.has(role)) {
      errors.push(
        `harness_tiers.fast.checks has duplicate gate smoke role ${role}`,
      );
      continue;
    }
    roles.set(role, check);
  }
  for (const role of validGateSmokeRoles) {
    if (!roles.has(role)) {
      errors.push(`harness_tiers.fast.checks missing gate smoke role ${role}`);
    }
  }
  for (const [name, entry] of harnessChecks) {
    if (entry.gate_smoke_role !== undefined && !fastChecks.has(name)) {
      errors.push(`${name}.gate_smoke_role is only allowed on fast harness smoke checks`);
    }
  }
}

function validateCompactHelp(errors, targets, compactHelp) {
  if (
    !compactHelp ||
    typeof compactHelp !== "object" ||
    Array.isArray(compactHelp)
  ) {
    errors.push("compact_help must be an object");
    return;
  }
  if (!Array.isArray(compactHelp.entries) || compactHelp.entries.length === 0) {
    errors.push("compact_help.entries must be a non-empty array");
    return;
  }
  if (compactHelp.entries.length > compactHelpMaxEntries) {
    errors.push(
      `compact_help.entries must not exceed ${compactHelpMaxEntries} entries`,
    );
  }

  const compactTargets = new Set();
  for (const [entryIndex, helpEntry] of compactHelp.entries.entries()) {
    const label = `compact_help.entries[${entryIndex + 1}]`;
    if (
      typeof helpEntry?.target !== "string" ||
      helpEntry.target.trim() === ""
    ) {
      errors.push(`${label}.target must be a non-empty string`);
      continue;
    }
    if (compactTargets.has(helpEntry.target)) {
      errors.push(`${label} contains duplicate target ${helpEntry.target}`);
    }
    compactTargets.add(helpEntry.target);

    const target = targets.get(helpEntry.target);
    if (!target) {
      errors.push(`${label} references unknown target ${helpEntry.target}`);
      continue;
    }
    if (target.target_class !== "public") {
      errors.push(
        `${helpEntry.target} appears in compact_help but is not target_class public`,
      );
    }
    validateHelpEntryText(errors, helpEntry, label);
  }
}

function validateHelpTiers(errors, targets, tiers) {
  if (!Array.isArray(tiers) || tiers.length === 0) {
    errors.push("help_tiers[] must be a non-empty array");
    return;
  }

  const tierNames = new Set();
  const placements = new Map();

  for (const [tierIndex, tier] of tiers.entries()) {
    const tierLabel = `help_tiers[${tierIndex + 1}]`;
    const tierName = tier?.name;
    if (typeof tierName !== "string" || tierName.trim() === "") {
      errors.push(`${tierLabel}.name must be a non-empty string`);
    } else {
      if (tierNames.has(tierName)) {
        errors.push(`help_tiers contains duplicate tier ${tierName}`);
      }
      tierNames.add(tierName);
    }

    if (!Array.isArray(tier?.entries) || tier.entries.length === 0) {
      errors.push(`${tierLabel}.entries must be a non-empty array`);
      continue;
    }

    const tierTargets = new Set();
    for (const [entryIndex, helpEntry] of tier.entries.entries()) {
      const label = `${tierLabel}.entries[${entryIndex + 1}]`;
      if (
        typeof helpEntry?.target !== "string" ||
        helpEntry.target.trim() === ""
      ) {
        errors.push(`${label}.target must be a non-empty string`);
        continue;
      }
      if (tierTargets.has(helpEntry.target)) {
        errors.push(`${label} contains duplicate target ${helpEntry.target}`);
      }
      tierTargets.add(helpEntry.target);

      const target = targets.get(helpEntry.target);
      if (!target) {
        errors.push(`${label} references unknown target ${helpEntry.target}`);
        continue;
      }
      if (target.target_class !== "public") {
        errors.push(
          `${helpEntry.target} appears in help tier ${tierName ?? "unknown"} but is not target_class public`,
        );
      }
      const targetPlacements = placements.get(helpEntry.target) ?? [];
      targetPlacements.push(tierName ?? "unknown");
      placements.set(helpEntry.target, targetPlacements);

      validateHelpEntryText(errors, helpEntry, label);
    }
  }

  for (const target of targets.values()) {
    if (target.target_class !== "public") {
      continue;
    }
    const targetPlacements = placements.get(target.name) ?? [];
    if (targetPlacements.length === 0) {
      errors.push(
        `public target ${target.name} is missing help tier placement`,
      );
    } else if (targetPlacements.length > 1) {
      errors.push(
        `public target ${target.name} appears in multiple help tiers: ${targetPlacements.join(",")}`,
      );
    }
  }
}

function validateHelpEntryText(errors, helpEntry, label) {
  if (
    typeof helpEntry.description !== "string" ||
    helpEntry.description.trim() === ""
  ) {
    errors.push(`${label}.description must be a non-empty string`);
  } else if (helpEntry.description.includes("\n")) {
    errors.push(`${label}.description must be a single line`);
  }
  if (helpEntry.usage !== undefined) {
    if (typeof helpEntry.usage !== "string" || helpEntry.usage.trim() === "") {
      errors.push(`${label}.usage must be a non-empty string when present`);
    } else if (helpEntry.usage.includes("\n")) {
      errors.push(`${label}.usage must be a single line`);
    }
  }
}

export function renderTaskSurfaceMake(manifest) {
  const lines = [
    "# Code generated by scripts/render-task-surface-make.mjs; DO NOT EDIT.",
    "",
  ];
  lines.push(
    `.PHONY: ${targetEntries(manifest)
      .map((entry) => entry.name)
      .join(" ")}`,
  );
  lines.push("");
  lines.push("TASK_SURFACE_HELP_LINES := \\");
  for (const line of helpLines(manifest)) {
    lines.push(`\t'${escapeMakeSingleQuoted(line)}' \\`);
  }
  lines.push("\t''");
  lines.push("");
  lines.push("TASK_SURFACE_HELP_ALL_LINES := \\");
  for (const line of helpAllLines(manifest)) {
    lines.push(`\t'${escapeMakeSingleQuoted(line)}' \\`);
  }
  lines.push("\t''");
  lines.push("");
  lines.push(
    'TASK_SURFACE_RUN_ENV = NODE_BIN="$(NODE_BIN)" TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)" TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)"',
  );
  lines.push(
    'TASK_SURFACE_GO_ENV = GO="$(GO)" GO_CACHE_DIR="$(GO_CACHE_DIR)" GO_MOD_CACHE_DIR="$(GO_MOD_CACHE_DIR)" NODE_BIN="$(NODE_BIN)"',
  );
  lines.push(
    'TASK_SURFACE_SERVICE_SCHEDULE_ENV = $(TASK_SURFACE_RUN_ENV) TEST_SERVICES_BIN="$(TEST_SERVICES_BIN)" RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)" RUN_SERVICE_BACKED_SCHEDULE_SCRIPT="$(RUN_SERVICE_BACKED_SCHEDULE_SCRIPT)" SCHEDULER_MANIFEST="$(SCHEDULER_MANIFEST)" CARTULARY_RUNNER_SCRIPT="$(CARTULARY_RUNNER_SCRIPT)"',
  );
  lines.push(
    `TASK_SURFACE_CHECK_SCHEDULER_OVERRIDE_ENV = ${checkSchedulerOverrideEnvExpression()}`,
  );
  lines.push(
    "CHECK_FRONTEND_INSTALL_TARGET = $(if $(filter 1,$(CI)),frontend-install-ci,frontend-install)",
  );
  lines.push(
    'RUN_MAKE_NODE_TOOL = env NODE_BIN="$(NODE_BIN)" $(2) ./scripts/run-make-node-tool.sh $(1)',
  );
  lines.push("RUN_PUBLIC_PREFLIGHT = $(RUN_HARNESS_PREFLIGHT) $(1)");
  lines.push(
    "RUN_TARGET_SUMMARY = $(Q)env $(TASK_SURFACE_RUN_ENV) $(NODE_BIN) $(CARTULARY_RUNNER_SCRIPT) target-summary $(1) $(2)",
  );
  lines.push("");
  for (const recipe of makeRecipeEntries(manifest)) {
    lines.push(...renderMakeRecipe(recipe, manifest));
    lines.push("");
  }
  return `${lines.join("\n")}\n`;
}

function checkSchedulerOverrideEnvExpression() {
  return resourceOverrideEnvVariablesForScheduler("check")
    .map(
      (name) =>
        `$(if $(filter undefined,$(origin ${name})),,${name}="$(${name})")`,
    )
    .join(" ");
}

function renderMakeRecipe(recipe, manifest) {
  const entry = targetEntryMap(manifest).get(recipe.target);
  const prerequisitePrelude = renderPrerequisitePrelude(recipe, entry);
  const preflightPrelude = renderPreflightPrelude(recipe, entry);
  const prerequisites =
    prerequisitePrelude.length > 0 ? "" : (recipe.prerequisites ?? []).join(" ");
  const header = prerequisites
    ? `${recipe.target}: ${prerequisites}`
    : `${recipe.target}:`;
  const prefix = renderRecipePrefix(recipe, entry);
  if (recipe.type === "alias") {
    const lines = [...prefix, header, ...preflightPrelude, ...prerequisitePrelude];
    if (entry?.output_policy?.summary_schema === "cartulary.tool_run_summary.v3") {
      lines.push(`\t$(call RUN_TARGET_SUMMARY,${recipe.target},pass)`);
    }
    return lines;
  }
  if (recipe.type === "cleanup") {
    if (recipe.scope === "distclean") {
      return [
        ...prefix,
        header,
        ...preflightPrelude,
        "\t$(RUN_HARNESS_CLEANUP) distclean $(DISTCLEAN_PATHS)",
      ];
    }
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      "\t$(RUN_HARNESS_CLEANUP) clean $(CLEAN_PATHS)",
    ];
  }
  if (recipe.type === "print_help") {
    const variable =
      recipe.scope === "all"
        ? "TASK_SURFACE_HELP_ALL_LINES"
        : "TASK_SURFACE_HELP_LINES";
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      `\t$(Q)printf '%s\\n' $(${variable})`,
    ];
  }
  if (recipe.type === "sequence") {
    const sequence = sequenceDefinition(manifest, recipe.sequence);
    const env = [
      'MAKE="$(MAKE)"',
      'NODE_BIN="$(NODE_BIN)"',
      'TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)"',
      'TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)"',
    ];
    for (const step of sequence.steps) {
      if (step.jobsVariable) {
        env.push(`${step.jobsVariable}="$(${step.jobsVariable})"`);
      }
    }
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      ...prerequisitePrelude,
      `\t$(Q)${env.join(" ")} $(RUN_MAKE_SEQUENCE_SCRIPT) --sequence ${recipe.sequence}`,
    ];
  }
  if (recipe.type === "check_schedule") {
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      ...prerequisitePrelude,
      `\t$(Q)env MAKE="$(MAKE)" TEST_SERVICES_BIN="$(TEST_SERVICES_BIN)" $(TASK_SURFACE_RUN_ENV) $(TASK_SURFACE_CHECK_SCHEDULER_OVERRIDE_ENV) $(NODE_BIN) $(RUN_CHECK_SCHEDULE_SCRIPT) --target ${recipe.target} --manifest "$(${recipe.manifest_variable})"`,
    ];
  }
  if (recipe.type === "go_target") {
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      ...prerequisitePrelude,
      `\t$(Q)env ${goTargetEnv(recipe).join(" ")} $(NODE_BIN) $(CARTULARY_RUNNER_SCRIPT) go-target ${recipe.target}`,
    ];
  }
  if (recipe.type === "service_backed_target") {
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      ...prerequisitePrelude,
      `\t$(Q)env ${goTargetEnv(recipe).join(" ")} $(TEST_SERVICES_BIN) run -- $(NODE_BIN) $(CARTULARY_RUNNER_SCRIPT) go-target ${recipe.target}`,
    ];
  }
  if (recipe.type === "service_backed_schedule") {
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      ...prerequisitePrelude,
      `\t$(Q)env $(TASK_SURFACE_SERVICE_SCHEDULE_ENV) $(NODE_BIN) $(CARTULARY_RUNNER_SCRIPT) service-backed-target --target ${recipe.target} --phase-label "${recipe.phase_label}" --service-wrapper ${recipe.service_wrapper}`,
    ];
  }
  if (recipe.type === "browser_batch") {
    const wrapper =
      recipe.service_wrapper === "test-services"
        ? "$(TEST_SERVICES_BIN) run -- "
        : "";
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      ...prerequisitePrelude,
      `\t$(Q)env $(BROWSER_E2E_OWNED_STACK_ENV) TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)" PLAYWRIGHT_WORKERS=${recipe.workers} BROWSER_E2E_FUNCTIONAL_SHARDS="$(BROWSER_E2E_FUNCTIONAL_SHARDS)" ${wrapper}./scripts/run-browser-e2e-target.sh ${recipe.stage}`,
    ];
  }
  if (recipe.type === "phase_command") {
    const lines = [
      ...prefix,
      header,
      ...preflightPrelude,
      ...prerequisitePrelude,
      ...renderPhaseCommandRecipe(recipe, entry),
    ];
    if (recipe.success_summary === true) {
      lines.push(`\t$(call RUN_TARGET_SUMMARY,${recipe.target},pass)`);
    }
    return lines;
  }
  if (recipe.type === "summary_target") {
    const status = recipe.status ?? "pass";
    const phaseLabel =
      recipe.phase_label ?? `${recipe.target} child ${recipe.child_target}`;
    const projection = recipe.projection
      ? ` --projection ${recipe.projection}`
      : "";
    const env = [
      'MAKE_BIN="$(MAKE)"',
      'NODE_BIN="$(NODE_BIN)"',
      'TEST_OUTPUT_SCRIPT="$(TEST_OUTPUT_SCRIPT)"',
      'TASK_SURFACE_MANIFEST="$(TASK_SURFACE_MANIFEST)"',
      'RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)"',
    ];
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      ...prerequisitePrelude,
      `\t$(Q)env ${env.join(" ")} $(NODE_BIN) $(CARTULARY_RUNNER_SCRIPT) summary-target --target ${recipe.target} --child-target ${recipe.child_target} --status ${status} --phase-label "${phaseLabel}"${projection}`,
    ];
  }
  if (recipe.type === "node_tool") {
    const env = makeNodeToolMakeEnvVars(recipe.target)
      .map((name) => `${name}="$(${name})"`)
      .join(" ");
    return [
      ...prefix,
      header,
      ...preflightPrelude,
      ...prerequisitePrelude,
      `\t$(Q)$(call RUN_MAKE_NODE_TOOL,${recipe.target},${env})`,
    ];
  }
  throw new Error(`unsupported Make recipe type ${recipe.type}`);
}

function shouldCentralizePrerequisiteOutput(recipe, entry = null) {
  return (
    entry?.target_class === "public" &&
    (recipe.prerequisites ?? []).length > 0 &&
    !["cleanup", "print_help"].includes(recipe.type)
  );
}

function renderPreflightPrelude(recipe, entry = null) {
  if (entry?.target_class !== "public") {
    return [];
  }
  return [`\t$(call RUN_PUBLIC_PREFLIGHT,${recipe.target})`];
}

function renderPrerequisitePrelude(recipe, entry = null) {
  if (!shouldCentralizePrerequisiteOutput(recipe, entry)) {
    return [];
  }
  return [
    "\t$(Q)if [ \"$${CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES:-0}\" != \"1\" ]; then env CARTULARY_SUPPRESS_CHILD_SUCCESS=1 $(MAKE) --silent --no-print-directory " +
      `${(recipe.prerequisites ?? []).join(" ")}; fi`,
  ];
}

function renderRecipePrefix(recipe, entry = null) {
  const lines = [];
  for (const comment of recipe.comments ?? []) {
    lines.push(`# ${comment}`);
  }
  if (recipe.test_target === "self") {
    lines.push(
      `${recipe.target}: export CARTULARY_TEST_TARGET ?= ${recipe.target}`,
    );
  }
  if (entry?.target_class !== "public") {
    lines.push(
      `${recipe.target}: export CARTULARY_SUPPRESS_CHILD_SUCCESS ?= 1`,
    );
  }
  for (const [name, value] of Object.entries(recipe.exports ?? {})) {
    lines.push(`${recipe.target}: export ${name} := ${value}`);
  }
  return lines;
}

function goTargetEnv(recipe) {
  return [
    "$(TASK_SURFACE_GO_ENV)",
    "CARTULARY_HARNESS_IDENTITY_PREPARED=1",
    ...Object.entries(recipe.env ?? {}).map(
      ([name, value]) => `${name}="${value}"`,
    ),
    'GO_TEST_SERVICE_PACKAGE_PARALLELISM="$(GO_TEST_SERVICE_PACKAGE_PARALLELISM)"',
  ];
}

function renderPhaseCommandRecipe(recipe, entry = null) {
  const env = Object.entries(recipe.env ?? {}).map(
    ([name, value]) => `${name}="${value}"`,
  );
  const envPrefix = env.length > 0 ? `${env.join(" ")} ` : "";
  const args = (recipe.args ?? []).join(" ");
  const argsSuffix = args ? ` ${args}` : "";
  if (recipe.mode === "run_phase") {
    const runnerEnv = [];
    if (recipe.success_summary === true) {
      runnerEnv.push("CARTULARY_SUPPRESS_CHILD_SUCCESS=1");
    }
    if (recipe.failure_note) {
      runnerEnv.push(`CARTULARY_PHASE_FAILURE_NOTE="${recipe.failure_note}"`);
    }
    const runnerPrefix = runnerEnv.length > 0 ? `${runnerEnv.join(" ")} ` : "";
    const childPrefix = env.length > 0 ? `env ${env.join(" ")} ` : "";
    return [
      `\t$(Q)${runnerPrefix}$(RUN_PHASE_SCRIPT) "${recipe.phase_label}" -- ${childPrefix}${recipe.command.join(" ")}${argsSuffix}`,
    ];
  }
  if (recipe.mode === "node") {
    return [`\t$(Q)${envPrefix}$(NODE_BIN) ${recipe.script}${argsSuffix}`];
  }
  if (recipe.mode === "command") {
    if (entry?.output_policy?.summary_schema === "cartulary.tool_run_summary.v3") {
      const childPrefix = env.length > 0 ? `env ${env.join(" ")} ` : "";
      const testTarget = `CARTULARY_TEST_TARGET="$\${CARTULARY_TEST_TARGET:-${recipe.target}}"`;
      return [
        `\t$(Q)${testTarget} $(RUN_PHASE_SCRIPT) "${recipe.target}" -- ${childPrefix}${recipe.command.join(" ")}${argsSuffix}`,
      ];
    }
    return [`\t$(Q)${envPrefix}${recipe.command.join(" ")}${argsSuffix}`];
  }
  throw new Error(`unsupported phase command mode ${recipe.mode}`);
}

export function helpLines(manifest) {
  const lines = ["Cartulary compact workflow task surface", ""];
  appendHelpTierLines(lines, {
    name: "compact",
    entries: compactHelpEntries(manifest),
  });
  lines.push("");
  lines.push("For all public targets, run: make help-all");
  lines.push(
    "For private/check internals, run: make task-surface-report TASK_SURFACE_REPORT_ARGS=--all",
  );
  return trimTrailingBlank(lines);
}

export function helpAllLines(manifest) {
  const lines = ["Cartulary public task surface", ""];
  lines.push("How to read task evidence:");
  lines.push("  phase -> target -> scheduler work unit -> artifact");
  lines.push(
    "  public evidence is runnable from this surface; support/internal evidence is shown by task-guide and explain-*.",
  );
  lines.push("");
  for (const tier of helpTiers(manifest)) {
    appendHelpTierLines(lines, tier);
    lines.push("");
  }
  return trimTrailingBlank(lines);
}

function appendHelpTierLines(lines, tier) {
  if (!tier) {
    return;
  }
  lines.push(`${tier.name}:`);
  for (const entry of tier.entries) {
    lines.push(...renderHelpEntryLines(entry));
  }
}

function renderHelpEntryLines(entry) {
  const command = `make ${entry.target}`;
  const description = entry.description.trim();
  const usage = typeof entry.usage === "string" ? entry.usage.trim() : "";
  if (!usage && entry.target.length <= helpTargetColumnWidth) {
    return [
      `  make ${entry.target.padEnd(helpTargetColumnWidth)} ${description}`,
    ];
  }
  const detail = usage ? `${usage} ${description}` : description;
  return [`  ${command}`, `${helpDescriptionIndent}${detail}`];
}

function trimTrailingBlank(lines) {
  if (lines[lines.length - 1] === "") {
    lines.pop();
  }
  return lines;
}

function escapeMakeSingleQuoted(value) {
  return value.replaceAll("'", "'\"'\"'");
}
