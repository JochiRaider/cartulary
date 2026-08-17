import { hasMakeNodeTool } from "../../command-surface/make-node-tools.mjs";
import { resolveSummaryGroups } from "../../execution/summary-topology.mjs";
import { readJSON, repoRoot } from "./model.mjs";
const validGateSmokeRoles = new Set([
  "public_make_wrapper", "work_graph_scheduler_semantic", "fixture_broker_semantic",
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
const ownerCommandTargets = new Set([
  "explain-test-owner",
  "service-backed-test-slice",
  "task-guide",
  "test-evidence-audit",
  "test-slice",
]);

const makeRecipeValidators = Object.freeze({
  artifact_binding: validateArtifactBindingRecipe,
  aggregate: validateAggregateRecipe,
  readiness_projection: validateReadinessProjectionRecipe,
  cleanup: validateCleanupRecipe,
  print_help: validatePrintHelpRecipe,
  work_graph: validateWorkGraphRecipe,
  step_command: validateStepCommandRecipe,
  owner_command: validateOwnerCommandRecipe,
  summary_target: validateSummaryTargetRecipe,
  node_tool: validateNodeToolRecipe,
});
const validMakeRecipeTypes = new Set(Object.keys(makeRecipeValidators));
function recipeCanProduceArtifactPolicy(target, recipe, artifactPolicy) {
  if (artifactPolicy === "none") {
    return true;
  }
  if (!recipe || typeof recipe !== "object") {
    return false;
  }
  if (artifactPolicy === "run_and_target_summaries") {
    return recipe.type === "work_graph" || recipe.graph_entry === true;
  }
  if (artifactPolicy === "scheduler_and_tool_run_summaries") {
    return (
      recipe.type === "work_graph" ||
      (recipe.type === "owner_command" &&
        ["test-slice", "service-backed-test-slice"].includes(target))
    );
  }
  if (artifactPolicy === "tool_run_summary") {
    return (
      recipe.type === "artifact_binding" ||
      recipe.type === "aggregate" ||
      recipe.type === "node_tool" ||
      recipe.type === "work_graph" ||
      recipe.type === "summary_target" ||
      recipe.type === "owner_command" ||
      recipe.type === "step_command"
    );
  }
  return false;
}
export function validateOutputPolicyRouting(errors, targets, recipes) {
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
export function validateMakeRecipes(errors, targets, sequences, recipes) {
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
    validatePrerequisiteJobs({
      errors,
      target,
      recipe,
      label,
      targets,
      sequences,
    });
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
    if (
      recipe.graph_child_skips_prerequisites !== undefined &&
      typeof recipe.graph_child_skips_prerequisites !== "boolean"
    ) {
      errors.push(`${label}.graph_child_skips_prerequisites must be a boolean`);
    }
    if (
      recipe.graph_child_skips_prerequisites === true &&
      (!Array.isArray(recipe.prerequisites) || recipe.prerequisites.length === 0)
    ) {
      errors.push(`${label}.graph_child_skips_prerequisites requires prerequisites`);
    }
    if (recipe.graph_entry !== undefined) {
      if (recipe.graph_entry !== true) {
        errors.push(`${label}.graph_entry must be true when present`);
      }
      if (recipe.type === "work_graph") {
        errors.push(`${label}.graph_entry is redundant for work_graph recipes`);
      }
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

function validatePrerequisiteJobs({
  errors,
  target,
  recipe,
  label,
  targets,
  sequences,
}) {
  if (recipe.prerequisite_jobs === undefined) {
    return;
  }
  const jobs = recipe.prerequisite_jobs;
  const runnablePrerequisites = Array.isArray(recipe.prerequisites)
    ? recipe.prerequisites.filter((prerequisite) => prerequisite !== "$(NODE_BIN)")
    : [];
  if (!Number.isInteger(jobs) || jobs < 2 || jobs > 8) {
    errors.push(`${label}.prerequisite_jobs must be an integer from 2 through 8`);
    return;
  }
  if (jobs > runnablePrerequisites.length) {
    errors.push(
      `${label}.prerequisite_jobs must not exceed its non-Node prerequisite count`,
    );
  }
  const sequenceTarget = Object.values(sequences ?? {}).some((sequence) =>
    (sequence.steps ?? []).some((step) => step.target === target),
  );
  if (targets.get(target)?.target_class !== "public" && !sequenceTarget) {
    errors.push(
      `${label}.prerequisite_jobs requires a public or sequence-step recipe`,
    );
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
      "make_node_tool",
      "owner_command",
      "summary_target",
      "work_graph",
    ].includes(recipe.type)
  ) {
    return true;
  }
  if (recipe.type !== "step_command") {
    return false;
  }
  if (recipe.mode === "node") {
    return true;
  }
  return (recipe.command ?? []).includes("$(NODE_BIN)");
}

function validateWorkGraphRecipe({ errors, target, recipe, label }) {
  if (!new Set(["target", "aggregate", "owner"]).has(recipe.selection)) {
    errors.push(`${label}.selection must be target, aggregate, or owner`);
  }
  if (recipe.selection === "aggregate" && !new Set(["test-fast", "check", "test", "ci", "release-check"]).has(target)) {
    errors.push(`${label}.selection=aggregate is limited to aggregate roots`);
  }
  if (recipe.selection === "owner" && !new Set(["test-slice", "service-backed-test-slice"]).has(target)) {
    errors.push(`${label}.selection=owner is limited to owner slice targets`);
  }
  if (recipe.service_backed_only !== undefined && typeof recipe.service_backed_only !== "boolean") {
    errors.push(`${label}.service_backed_only must be a boolean`);
  }
  if (recipe.service_backed_only === true && target !== "service-backed-test-slice") {
    errors.push(`${label}.service_backed_only is limited to service-backed-test-slice`);
  }
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

function validateStepCommandRecipe({ errors, recipe, label, helpers }) {
  if (!["run_step", "node", "command"].includes(recipe.mode)) {
    errors.push(`${label}.mode must be run_step, node, or command`);
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
  if (recipe.service_resolution !== undefined) {
    if (recipe.service_resolution !== "runtime-profile") {
      errors.push(`${label}.service_resolution must be runtime-profile`);
    }
    if (
      typeof recipe.browser_stage !== "string" ||
      !makeTargetPattern.test(recipe.browser_stage)
    ) {
      errors.push(
        `${label}.browser_stage must be a safe browser stage when service_resolution is runtime-profile`,
      );
    }
  } else if (recipe.browser_stage !== undefined) {
    errors.push(
      `${label}.browser_stage requires service_resolution=runtime-profile`,
    );
  }
  helpers.validateCommandTokens(errors, recipe.command, `${label}.command`, {
    required: recipe.mode !== "node",
  });
  helpers.validateCommandTokens(errors, recipe.args, `${label}.args`, {
    required: false,
  });
  if (
    recipe.mode === "run_step" &&
    (typeof recipe.step_label !== "string" || recipe.step_label.trim() === "")
  ) {
    errors.push(`${label}.step_label must be a non-empty string for run_step`);
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
    recipe.step_label !== undefined &&
    (typeof recipe.step_label !== "string" ||
      !makeValuePattern.test(recipe.step_label) ||
      recipe.step_label.trim() === "")
  ) {
    errors.push(`${label}.step_label must be a safe non-empty Make value`);
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

function validateOwnerCommandRecipe({ errors, target, label, helpers }) {
  if (!ownerCommandTargets.has(target)) {
    errors.push(`${label} target is not in the closed owner command family`);
  }
  validateNodeToolRecipe({ errors, target, label, helpers });
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

export function validateCommandTokens(
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

export function validateRootScriptPathPolicy(errors, token, label) {
  const script = normalizeRootScriptPath(token);
  if (!script) {
    return true;
  }
  errors.push(
    `${label} must not reference retired root scripts/ path ${script}; use an owner path under tools/** or a deployment package path under deploy/**/scripts/**`,
  );
  return false;
}

export function validateNamedTargetList(
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

export function validateSummaryGroups(
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

export function validateHarnessTiers(errors, harnessChecks, tiers) {
  if (!tiers || typeof tiers !== "object" || Array.isArray(tiers)) {
    errors.push("harness_tiers must be an object");
    return;
  }
  const fastChecks = new Set();
  const selectedChecks = new Set();
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
      selectedChecks.add(check);
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
  for (const check of harnessChecks.keys()) {
    if (!selectedChecks.has(check)) {
      errors.push(`harness check ${check} is not reachable from any harness tier`);
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

export function validateCompactHelp(errors, targets, compactHelp) {
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

export function validateHelpTiers(errors, targets, tiers) {
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
