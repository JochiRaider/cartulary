import { existsSync } from "node:fs";
import path from "node:path";
import { hasMakeNodeTool } from "../../command-surface/make-node-tools.mjs";
import { collectExplicitSummaryProjectionErrors, loadSummaryTopologyContext, resolveSummaryGroups } from "../../execution/summary-topology.mjs";
import {
  canonicalInternalMakeValues, compactHelpEntries, harnessCheckEntries, harnessTierChecks, helpTiers, makeRecipeEntries, nonCanonicalPublicMakeVariables, readJSON, repoRoot, restrictedInternalMakeVariables, sequenceDefinition, summaryEntryMap, targetEntries, targetEntryMap, taskSurfaceSchemaID,
} from "./model.mjs";

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
  "removed",
]);
const validServiceRequirements = new Set([
  "postgres",
  "object_store",
  "browser_stack",
  "vite",
]);
const validOutputClasses = new Set([
  "aggregate_summary_with_artifacts",
  "destructive_human",
  "human_summary",
  "interactive_raw",
  "machine_stdout_json",
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
const validSummarySchemas = new Set([
  "cartulary.tool_run_summary.v3",
  "cartulary.otel_conformance_summary.v1",
]);
const validRawStreamPolicies = new Set([
  "explicit_detail_only",
  "failure_or_verbose",
  "human_visible_actions",
  "interactive",
  "machine_json_stdout",
  "never_default",
]);
const validInputBindings = new Set(["make_variable"]);
const validInputSources = new Set([
  "make_command_line",
  "environment",
  "makefile_default",
  "internal_default",
  "manifest",
]);
const validInputTypes = new Set([
  "enum",
  "exact_1_bool",
  "phase_row_ids",
  "path",
  "phase_id",
  "phase_namespace",
  "positive_decimal",
  "positive_integer",
  "result_selector",
  "run_id",
  "task_surface_report_args",
  "target_name",
]);
const validEmptyStringRules = new Set(["invalid", "omitted", "false"]);
const validInputNormalizations = new Set([
  "none",
  "trim",
  "trim_lowercase",
  "path_token",
]);
const validInputInvalidReasons = new Set(["usage_error", "configuration_error"]);
const validInputSummaryEmission = new Set([
  "none",
  "value",
  "redacted_value",
  "source_and_value",
]);
const validInputChildForwarding = new Set([
  "none",
  "argv",
  "runtime_env",
  "argv_and_runtime_env",
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
export const validSideEffectClasses = new Set([
  "none",
  "retained_artifacts",
  "generated_artifacts",
  "authored_source_write",
  "build_outputs",
  "tool_install",
  "service_start",
  "service_resource_mutation",
  "destructive_cleanup",
  "runtime_reset",
]);
const sideEffectRequiredFields = Object.freeze({
  none: [],
  retained_artifacts: ["artifact_policy"],
  generated_artifacts: ["file_families"],
  authored_source_write: ["paths"],
  build_outputs: ["output_roots"],
  tool_install: ["install_roots", "cleanup"],
  service_start: ["ownership_mode", "lifecycle_machine"],
  service_resource_mutation: [
    "ownership_mode",
    "resource_families",
    "lifecycle_machine",
  ],
  destructive_cleanup: ["predicates_section"],
  runtime_reset: ["predicates_section"],
});
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
  "machine_stdout_json",
]);
const makeRecipeValidators = Object.freeze({
  artifact_binding: validateArtifactBindingRecipe,
  aggregate: validateAggregateRecipe,
  readiness_projection: validateReadinessProjectionRecipe,
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


export function validateTaskSurfaceManifest(manifest, manifestPath) {
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
    validateCheckProjection(errors, entry);
    if (!validLifecycleStates.has(entry.lifecycle_state)) {
      errors.push(
        `${entry.name} has invalid lifecycle_state ${JSON.stringify(entry.lifecycle_state)}`,
      );
    } else if (
      entry.target_class === "public" &&
      entry.lifecycle_state !== "public_active"
    ) {
      errors.push(
        `${entry.name}.lifecycle_state must be public_active for public targets`,
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
          } else if (!validateRootScriptPathPolicy(errors, script, `${entry.name}.backing_scripts`)) {
            continue;
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
      validatePublicSideEffects(errors, entry);
      validateOutputPolicy(errors, entry);
      validatePublicInputContract(errors, entry);
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
          } else if (!validateRootScriptPathPolicy(errors, script, `${entry.name}.backing_scripts`)) {
            continue;
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
        if (
          step.skip_prerequisites !== undefined &&
          typeof step.skip_prerequisites !== "boolean"
        ) {
          errors.push(`${label}.skip_prerequisites must be a boolean`);
        }
        if (step.type === "parallel" && step.skip_prerequisites === true) {
          errors.push(`${label}.skip_prerequisites is supported only for serial steps`);
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

function validateCheckProjection(errors, entry) {
  if (entry.check_projection === undefined) {
    return;
  }
  if (
    !entry.check_projection ||
    typeof entry.check_projection !== "object" ||
    Array.isArray(entry.check_projection)
  ) {
    errors.push(`${entry.name}.check_projection must be an object`);
    return;
  }
  const allowed = new Set([
    "mode",
    "schedule",
    "stage",
    "evidence",
    "evidence_class",
    "reason_code",
    "full_target",
    "full_target_equivalent",
  ]);
  for (const key of Object.keys(entry.check_projection)) {
    if (!allowed.has(key)) {
      errors.push(`${entry.name}.check_projection has unknown key ${key}`);
    }
  }
  const projection = entry.check_projection;
  if (!["direct", "projection"].includes(projection.mode)) {
    errors.push(`${entry.name}.check_projection.mode must be direct or projection`);
  }
  for (const key of ["schedule", "stage", "evidence", "evidence_class", "reason_code", "full_target"]) {
    if (
      projection[key] !== undefined &&
      (typeof projection[key] !== "string" || projection[key].trim() === "")
    ) {
      errors.push(`${entry.name}.check_projection.${key} must be a non-empty string`);
    }
  }
  if (projection.full_target_equivalent !== undefined && typeof projection.full_target_equivalent !== "boolean") {
    errors.push(`${entry.name}.check_projection.full_target_equivalent must be a boolean`);
  }
  if (
    projection.mode === "projection" &&
    projection.full_target_equivalent !== false
  ) {
    errors.push(`${entry.name}.check_projection projection mode must declare full_target_equivalent=false`);
  }
  if (projection.mode === "projection") {
    for (const key of ["schedule", "stage", "evidence", "evidence_class", "reason_code", "full_target"]) {
      if (typeof projection[key] !== "string" || projection[key].trim() === "") {
        errors.push(`${entry.name}.check_projection projection mode must declare ${key}`);
      }
    }
    if (entry.default_inclusion_sets?.includes("check")) {
      errors.push(
        `${entry.name}.check_projection projection mode must not advertise direct default_inclusion_sets check membership`,
      );
    }
  }
  if (projection.mode === "direct") {
    for (const key of ["schedule", "stage", "evidence", "evidence_class", "reason_code", "full_target"]) {
      if (typeof projection[key] !== "string" || projection[key].trim() === "") {
        errors.push(`${entry.name}.check_projection direct mode must declare ${key}`);
      }
    }
    if (projection.full_target_equivalent !== true) {
      errors.push(`${entry.name}.check_projection direct mode must declare full_target_equivalent=true`);
    }
    if (!entry.default_inclusion_sets?.includes("check")) {
      errors.push(
        `${entry.name}.check_projection direct mode must advertise direct default_inclusion_sets check membership`,
      );
    }
  }
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

function validateStringArrayField(errors, value, label) {
  if (!Array.isArray(value) || value.length === 0) {
    errors.push(`${label} must be a non-empty array`);
    return;
  }
  const seen = new Set();
  for (const [index, item] of value.entries()) {
    const itemLabel = `${label}[${index + 1}]`;
    if (typeof item !== "string" || item.trim() === "") {
      errors.push(`${itemLabel} must be a non-empty string`);
      continue;
    }
    if (seen.has(item)) {
      errors.push(`${label} contains duplicate ${item}`);
    }
    seen.add(item);
  }
}

function validatePublicSideEffects(errors, entry) {
  if (!Array.isArray(entry.side_effects) || entry.side_effects.length === 0) {
    errors.push(`${entry.name}.side_effects must declare at least one side-effect class`);
    return;
  }
  const seen = new Set();
  for (const [index, sideEffect] of entry.side_effects.entries()) {
    const label = `${entry.name}.side_effects[${index + 1}]`;
    if (!sideEffect || typeof sideEffect !== "object" || Array.isArray(sideEffect)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    const sideEffectClass = sideEffect.class;
    if (
      typeof sideEffectClass !== "string" ||
      !validSideEffectClasses.has(sideEffectClass)
    ) {
      errors.push(`${label}.class must be one of ${[...validSideEffectClasses].join("|")}`);
      continue;
    }
    if (seen.has(sideEffectClass)) {
      errors.push(`${entry.name}.side_effects contains duplicate ${sideEffectClass}`);
    }
    seen.add(sideEffectClass);
    if (
      typeof sideEffect.owner_section !== "string" ||
      !ownerSectionPattern.test(sideEffect.owner_section)
    ) {
      errors.push(`${label}.owner_section must be a Section reference`);
    }
    for (const field of sideEffectRequiredFields[sideEffectClass]) {
      if (!Object.hasOwn(sideEffect, field)) {
        errors.push(`${label}.${field} must be declared for ${sideEffectClass}`);
        continue;
      }
      if (
        [
          "file_families",
          "paths",
          "output_roots",
          "install_roots",
          "resource_families",
        ].includes(field)
      ) {
        validateStringArrayField(errors, sideEffect[field], `${label}.${field}`);
      } else if (
        typeof sideEffect[field] !== "string" ||
        sideEffect[field].trim() === ""
      ) {
        errors.push(`${label}.${field} must be a non-empty string`);
      }
    }
  }
  if (seen.has("none") && seen.size > 1) {
    errors.push(`${entry.name}.side_effects none is mutually exclusive with other classes`);
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
  if (policy.summary_schema !== null && !validSummarySchemas.has(policy.summary_schema)) {
    errors.push(
      `${entry.name}.output_policy.summary_schema must be a registered summary schema or null`,
    );
  }
  if (policy.artifact_policy === "none" && policy.summary_schema !== null) {
    errors.push(
      `${entry.name}.output_policy.summary_schema must be null when artifact_policy is none`,
    );
  }
  if (policy.artifact_policy !== "none" && !validSummarySchemas.has(policy.summary_schema)) {
    errors.push(
      `${entry.name}.output_policy.summary_schema must be a registered summary schema when artifact_policy is ${policy.artifact_policy}`,
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

function validatePublicInputContract(errors, entry) {
  const contract = entry.input_contract;
  if (!contract || typeof contract !== "object" || Array.isArray(contract)) {
    errors.push(`${entry.name}.input_contract must be declared for public targets`);
    return;
  }
  const allowedKeys = new Set([
    "undeclared_make_command_line",
    "undeclared_inherited_env",
    "inputs",
  ]);
  for (const key of Object.keys(contract)) {
    if (!allowedKeys.has(key)) {
      errors.push(`${entry.name}.input_contract has unknown key ${key}`);
    }
  }
  if (contract.undeclared_make_command_line !== "usage_error") {
    errors.push(`${entry.name}.input_contract.undeclared_make_command_line must be usage_error`);
  }
  if (contract.undeclared_inherited_env !== "ignore") {
    errors.push(`${entry.name}.input_contract.undeclared_inherited_env must be ignore`);
  }
  if (!Array.isArray(contract.inputs)) {
    errors.push(`${entry.name}.input_contract.inputs must be an array`);
    return;
  }
  const seen = new Set();
  for (const [index, input] of contract.inputs.entries()) {
    const label = `${entry.name}.input_contract.inputs[${index + 1}]`;
    if (!input || typeof input !== "object" || Array.isArray(input)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    const allowedInputKeys = new Set([
      "name",
      "binding",
      "allowed_sources",
      "required",
      "type",
      "values",
      "min",
      "max",
      "default",
      "empty_string",
      "normalization",
      "invalid_reason",
      "summary_emission",
      "child_forwarding",
    ]);
    for (const key of Object.keys(input)) {
      if (!allowedInputKeys.has(key)) {
        errors.push(`${label} has unknown key ${key}`);
      }
    }
    if (typeof input.name !== "string" || !makeVariablePattern.test(input.name)) {
      errors.push(`${label}.name must be a safe Make variable name`);
    } else if (seen.has(input.name)) {
      errors.push(`${entry.name}.input_contract.inputs contains duplicate ${input.name}`);
    } else {
      seen.add(input.name);
    }
    if (!validInputBindings.has(input.binding)) {
      errors.push(`${label}.binding must be make_variable`);
    }
    if (!Array.isArray(input.allowed_sources) || input.allowed_sources.length === 0) {
      errors.push(`${label}.allowed_sources must be a non-empty array`);
    } else {
      for (const source of input.allowed_sources) {
        if (!validInputSources.has(source)) {
          errors.push(`${label}.allowed_sources contains invalid source ${JSON.stringify(source)}`);
        }
      }
    }
    if (typeof input.required !== "boolean") {
      errors.push(`${label}.required must be a boolean`);
    }
    if (!validInputTypes.has(input.type)) {
      errors.push(`${label}.type has invalid value ${JSON.stringify(input.type)}`);
    }
    if (input.type === "enum") {
      if (!Array.isArray(input.values) || input.values.length === 0) {
        errors.push(`${label}.values must be a non-empty array for enum inputs`);
      } else {
        for (const value of input.values) {
          if (typeof value !== "string" || value.trim() === "") {
            errors.push(`${label}.values contains an invalid enum token`);
          }
        }
      }
    }
    if (
      Object.hasOwn(input, "min") &&
      (!Number.isFinite(input.min) || input.min < 0)
    ) {
      errors.push(`${label}.min must be a non-negative finite number`);
    }
    if (
      Object.hasOwn(input, "max") &&
      (!Number.isFinite(input.max) || input.max < 0)
    ) {
      errors.push(`${label}.max must be a non-negative finite number`);
    }
    if (!Object.hasOwn(input, "default")) {
      errors.push(`${label}.default must be declared, using null when omitted`);
    }
    if (!validEmptyStringRules.has(input.empty_string)) {
      errors.push(`${label}.empty_string has invalid value ${JSON.stringify(input.empty_string)}`);
    }
    if (!validInputNormalizations.has(input.normalization)) {
      errors.push(`${label}.normalization has invalid value ${JSON.stringify(input.normalization)}`);
    }
    if (!validInputInvalidReasons.has(input.invalid_reason)) {
      errors.push(`${label}.invalid_reason must be usage_error or configuration_error`);
    }
    if (!validInputSummaryEmission.has(input.summary_emission)) {
      errors.push(`${label}.summary_emission has invalid value ${JSON.stringify(input.summary_emission)}`);
    }
    if (!validInputChildForwarding.has(input.child_forwarding)) {
      errors.push(`${label}.child_forwarding has invalid value ${JSON.stringify(input.child_forwarding)}`);
    }
  }
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
      recipe.type === "artifact_binding" ||
      recipe.type === "aggregate" ||
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
    if (
      recipeRequiresNodeRuntime(recipe) &&
      Array.isArray(recipe.prerequisites) &&
      !recipe.prerequisites.includes("$(NODE_BIN)")
    ) {
      errors.push(`${label}.prerequisites must include $(NODE_BIN)`);
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

function validateArtifactBindingRecipe({ errors, recipe, label }) {
  if (!Array.isArray(recipe.prerequisites) || recipe.prerequisites.length === 0) {
    errors.push(`${label}.prerequisites must name at least one artifact producer`);
  }
}

function validateAggregateRecipe({ errors, recipe, label, targets }) {
  if (!Array.isArray(recipe.prerequisites) || recipe.prerequisites.length === 0) {
    errors.push(`${label}.prerequisites must name at least one aggregate child`);
    return;
  }
  for (const child of recipe.prerequisites) {
    if (!targets.has(child)) {
      errors.push(`${label}.prerequisites references unknown aggregate child ${child}`);
    }
  }
}

function validateReadinessProjectionRecipe({ errors, target, recipe, label, targets }) {
  if (targets.get(target)?.target_class === "public") {
    errors.push(`${label} readiness_projection must remain internal`);
  }
  if (!Array.isArray(recipe.prerequisites) || recipe.prerequisites.length !== 1) {
    errors.push(`${label}.prerequisites must name exactly one readiness producer`);
    return;
  }
  if (!targets.has(recipe.prerequisites[0])) {
    errors.push(`${label}.prerequisites references unknown readiness producer ${recipe.prerequisites[0]}`);
  }
}

function recipeRequiresNodeRuntime(recipe) {
  if (
    [
      "browser_batch",
      "check_schedule",
      "go_target",
      "make_node_tool",
      "service_backed_schedule",
      "service_backed_target",
      "summary_target",
    ].includes(recipe.type)
  ) {
    return true;
  }
  if (recipe.type !== "phase_command") {
    return false;
  }
  if (recipe.mode === "node") {
    return true;
  }
  return (recipe.command ?? []).includes("$(NODE_BIN)");
}

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
    errors.push(`${label} has no tools/harness/command-surface/make-node-tools.mjs registry entry`);
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
  const script = normalizeRootScriptPath(token);
  if (!script) {
    return;
  }
  validateRootScriptPathPolicy(errors, script, label);
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
