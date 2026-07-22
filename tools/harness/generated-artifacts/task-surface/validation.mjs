import { existsSync } from "node:fs";
import path from "node:path";
import { hasMakeNodeTool } from "../../command-surface/make-node-tools.mjs";
import { collectExplicitSummaryProjectionErrors, loadSummaryTopologyContext, resolveSummaryGroups } from "../../execution/summary-topology.mjs";
import {
  canonicalInternalMakeValues, compactHelpEntries, harnessCheckEntries, harnessTierChecks, helpTiers, makeRecipeEntries, nonCanonicalPublicMakeVariables, readJSON, repoRoot, restrictedInternalMakeVariables, sequenceDefinition, summaryEntryMap, targetEntries, targetEntryMap, taskSurfaceSchemaID,
} from "./model.mjs";
import {
  validateCompactHelp,
  validateCommandTokens,
  validateHarnessTiers,
  validateHelpTiers,
  validateMakeRecipes,
  validateNamedTargetList,
  validateOutputPolicyRouting,
  validateRootScriptPathPolicy,
  validateSummaryGroups,
} from "./recipe-validation.mjs";
import { validatePublicInputContract } from "./input-validation.mjs";

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
  "test_owner_slices",
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
  "cartulary.tool_run_summary.v5",
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
  "external_network",
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
  external_network: [
    "endpoint_input",
    "protocol",
    "privacy_contract",
    "redirect_policy",
    "timeout_contract",
    "failure_contract",
  ],
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
const compactHelpMaxEntries = 12;
const helpTargetColumnWidth = 30;
const helpDescriptionIndent = " ".repeat(
  2 + "make ".length + helpTargetColumnWidth + 1,
);
const makeVariablePattern = /^[A-Z][A-Z0-9_]*$/;
const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const commandIDPattern = /^cartulary\.harness\.command\.[a-z][a-z0-9_]*\.v[1-9][0-9]*$/;
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
  const requireSequenceTopology = options.requireSequenceTopology !== false;
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

  const observabilityPolicy = manifest.observability_policy;
  if (!observabilityPolicy || typeof observabilityPolicy !== "object" || Array.isArray(observabilityPolicy)) {
    errors.push("observability_policy must be declared");
  } else {
    if (observabilityPolicy.scope !== "cartulary.harness.execution") {
      errors.push("observability_policy.scope must be cartulary.harness.execution");
    }
    const requiredObservabilityTargets = new Set();
    if (!Array.isArray(observabilityPolicy.required_targets) || observabilityPolicy.required_targets.length === 0) {
      errors.push("observability_policy.required_targets must be a non-empty array");
    } else {
      for (const [index, targetName] of observabilityPolicy.required_targets.entries()) {
        const label = `observability_policy.required_targets[${index + 1}]`;
        const target = targets.get(targetName);
        if (!target || target.target_class !== "public") errors.push(`${label} must name a public target`);
        if (requiredObservabilityTargets.has(targetName)) errors.push(`observability_policy contains duplicate required target ${targetName}`);
        requiredObservabilityTargets.add(targetName);
      }
    }
    const dispositions = new Map(
      [...requiredObservabilityTargets].map((target) => [target, "required"]),
    );
    for (const [property, disposition] of [
      ["excluded_targets", "excluded"],
      ["out_of_scope_targets", "out_of_scope"],
    ]) {
      const entries = observabilityPolicy[property];
      if (!Array.isArray(entries)) {
        errors.push(`observability_policy.${property} must be an array`);
        continue;
      }
      const localTargets = new Set();
      for (const [index, entry] of entries.entries()) {
        const label = `observability_policy.${property}[${index + 1}]`;
        if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
          errors.push(`${label} must be an object`);
          continue;
        }
        const target = targets.get(entry.target);
        if (!target || target.target_class !== "public") {
          errors.push(`${label}.target must name a public target`);
        }
        if (localTargets.has(entry.target)) {
          errors.push(`${label}.target duplicates ${entry.target}`);
        }
        localTargets.add(entry.target);
        if (dispositions.has(entry.target)) {
          errors.push(
            `${label}.target overlaps ${dispositions.get(entry.target)} disposition`,
          );
        } else {
          dispositions.set(entry.target, disposition);
        }
        if (typeof entry.owner_section !== "string" || !ownerSectionPattern.test(entry.owner_section)) {
          errors.push(`${label}.owner_section must be a Section reference`);
        }
        if (typeof entry.reason !== "string" || entry.reason.trim() === "") {
          errors.push(`${label}.reason must be non-empty`);
        }
      }
    }
    for (const target of targets.values()) {
      if (target.target_class === "public" && !dispositions.has(target.name)) {
        errors.push(`observability_policy omits public target ${target.name}`);
      }
    }
    for (const target of dispositions.keys()) {
      if (!targets.has(target)) {
        errors.push(`observability_policy contains unknown target ${target}`);
      }
    }

    const profileIDs = new Set();
    if (!Array.isArray(observabilityPolicy.measurement_profiles) || observabilityPolicy.measurement_profiles.length === 0) {
      errors.push("observability_policy.measurement_profiles must be a non-empty array");
    } else {
      for (const [index, profile] of observabilityPolicy.measurement_profiles.entries()) {
        const label = `observability_policy.measurement_profiles[${index + 1}]`;
        if (!profile || typeof profile !== "object" || Array.isArray(profile)) {
          errors.push(`${label} must be an object`);
          continue;
        }
        if (profileIDs.has(profile.profile_id)) {
          errors.push(`${label}.profile_id duplicates ${profile.profile_id}`);
        }
        profileIDs.add(profile.profile_id);
      }
    }
    const measurementTargets = new Set();
    const usedProfiles = new Set();
    if (!Array.isArray(observabilityPolicy.target_measurement_profiles)) {
      errors.push("observability_policy.target_measurement_profiles must be an array");
    } else {
      for (const [index, binding] of observabilityPolicy.target_measurement_profiles.entries()) {
        const label = `observability_policy.target_measurement_profiles[${index + 1}]`;
        const target = targets.get(binding?.target);
        if (!target || !new Set(["public", "check_internal"]).has(target.target_class)) {
          errors.push(`${label}.target must name a public or check-internal target`);
        }
        if (target?.target_class === "check_internal") {
          if (typeof target.command_id !== "string" || !commandIDPattern.test(target.command_id)) {
            errors.push(`${label}.target must declare a valid command_id`);
          } else if (commandIDs.has(target.command_id)) {
            errors.push(`${label}.target command_id duplicates ${commandIDs.get(target.command_id)}`);
          } else {
            commandIDs.set(target.command_id, target.name);
          }
        }
        if (measurementTargets.has(binding?.target)) {
          errors.push(`${label}.target duplicates ${binding?.target}`);
        }
        measurementTargets.add(binding?.target);
        if (!profileIDs.has(binding?.profile_id)) {
          errors.push(`${label}.profile_id references unknown profile ${binding?.profile_id}`);
        }
        if (binding?.allowed_policy_transition !== undefined && (
          binding.target !== "browser-e2e-webserver-backed" ||
          binding.allowed_policy_transition !== "browser_webserver_stage_capacity_1_to_2"
        )) {
          errors.push(`${label}.allowed_policy_transition is not valid for ${binding?.target}`);
        }
        usedProfiles.add(binding?.profile_id);
      }
    }
    for (const target of requiredObservabilityTargets) {
      if (!measurementTargets.has(target)) {
        errors.push(`observability_policy required target ${target} lacks a measurement profile`);
      }
    }
    for (const target of measurementTargets) {
      if (!requiredObservabilityTargets.has(target) && target !== "release-browser-readiness") {
        errors.push(`observability_policy measurement target ${target} is not required`);
      }
    }
    for (const profileID of profileIDs) {
      if (!usedProfiles.has(profileID)) {
        errors.push(`observability_policy measurement profile ${profileID} is unused`);
      }
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
      const resourceLimits = sequence.resource_limits;
      if (requireSequenceTopology) {
        const executionMode = sequence.execution_mode;
        if (!new Set(["serial", "dag"]).has(executionMode)) {
          errors.push(`sequence ${name}.execution_mode must be serial or dag`);
        }
        if (!Number.isInteger(sequence.max_jobs) || sequence.max_jobs < 1 || sequence.max_jobs > 16) {
          errors.push(`sequence ${name}.max_jobs must be an integer from 1 through 16`);
        }
        if (executionMode === "serial" && sequence.max_jobs !== 1) {
          errors.push(`sequence ${name}.max_jobs must be 1 for serial execution`);
        }
        if (!resourceLimits || typeof resourceLimits !== "object" || Array.isArray(resourceLimits)) {
          errors.push(`sequence ${name}.resource_limits must be an object`);
        } else {
          for (const [resource, limit] of Object.entries(resourceLimits)) {
            if ((limit !== "auto" && (!Number.isInteger(limit) || limit < 1))) {
              errors.push(`sequence ${name}.resource_limits.${resource} must be positive or auto`);
            }
          }
        }
        if (
          sequence.capacity_profile !== undefined &&
          (typeof sequence.capacity_profile !== "string" || sequence.capacity_profile === "")
        ) {
          errors.push(`sequence ${name}.capacity_profile must be a non-empty string`);
        }
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
      const sequenceTargets = new Set();
      for (const [index, step] of sequence.steps.entries()) {
        if (sequenceTargets.has(step?.target)) {
          errors.push(
            `sequence ${name} steps[${index + 1}].target duplicates ${step?.target}; occurrence aliases are unsupported`,
          );
        }
        sequenceTargets.add(step?.target);
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
        if (requireSequenceTopology && !Number.isInteger(step.priority)) {
          errors.push(`${label}.priority must be an integer`);
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
        if (requireSequenceTopology && !Array.isArray(step.needs ?? [])) {
          errors.push(`${label}.needs must be an array`);
        } else if (requireSequenceTopology) {
          for (const dependency of step.needs ?? []) {
            if (!sequence.steps.some((candidate) => candidate.target === dependency)) {
              errors.push(`${label}.needs references non-sequence target ${dependency}`);
            }
          }
        }
        if (requireSequenceTopology && step.resource_claims !== undefined) {
          if (!step.resource_claims || typeof step.resource_claims !== "object" || Array.isArray(step.resource_claims)) {
            errors.push(`${label}.resource_claims must be an object`);
          } else {
            for (const [resource, claim] of Object.entries(step.resource_claims)) {
              const limit = resourceLimits?.[resource];
              const validClaim = claim === "limit" || (Number.isInteger(claim) && claim >= 1);
              const validLimit = limit === "auto" || (Number.isInteger(limit) && limit >= 1);
              if (
                !validClaim ||
                !validLimit ||
                (Number.isInteger(claim) && Number.isInteger(limit) && claim > limit)
              ) {
                errors.push(`${label}.resource_claims.${resource} must fit the sequence resource limit`);
              }
            }
          }
        }
        if (
          requireSequenceTopology &&
          step.resource_profile !== undefined &&
          (typeof step.resource_profile !== "string" || step.resource_profile === "")
        ) {
          errors.push(`${label}.resource_profile must be a non-empty string`);
        }
        if (requireSequenceTopology) {
          const makeJobs = step.make_jobs;
          if (makeJobs !== undefined &&
            !(Number.isInteger(makeJobs) && makeJobs >= 1) &&
            !(typeof makeJobs === "string" && Object.hasOwn(step.resource_claims ?? {}, makeJobs))
          ) {
            errors.push(`${label}.make_jobs must be positive or reference a claimed resource`);
          }
          if (step.forwarding !== undefined && typeof step.forwarding !== "string") {
            errors.push(`${label}.forwarding must be a string`);
          }
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
      if (requireSequenceTopology) {
        const dependencyMap = new Map(
          sequence.steps.map((step) => [step.target, step.needs ?? []]),
        );
      const visiting = new Set();
      const visited = new Set();
      const visit = (target) => {
        if (visiting.has(target)) {
          errors.push(`sequence ${name} contains a dependency cycle at ${target}`);
          return;
        }
        if (visited.has(target)) return;
        visiting.add(target);
        for (const dependency of dependencyMap.get(target) ?? []) {
          if (dependencyMap.has(dependency)) visit(dependency);
        }
        visiting.delete(target);
        visited.add(target);
      };
        for (const target of dependencyMap.keys()) visit(target);
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
      `${entry.name}.command_id must match cartulary.harness.command.<name>.vN`,
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
