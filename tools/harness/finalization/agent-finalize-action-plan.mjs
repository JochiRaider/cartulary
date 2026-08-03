export const preflightSubstep = {
  id: "retained-run-preflight",
  target: null,
  commandKind: "retained_run_preflight",
  requiresResultsDir: true,
  mutatesRepo: false,
  run: "preflight",
};

function baseSubstep(definition) {
  return {
    id: definition.id,
    target: definition.target,
    command_kind: definition.commandKind,
    requires_results_dir: definition.requiresResultsDir,
    mutates_repo: definition.mutatesRepo,
    status: "pending",
    started_at: null,
    completed_at: null,
    duration_ms: null,
    exit_code: null,
    summary_json: null,
    stdout_log: null,
    stderr_log: null,
    skipped_reason: null,
  };
}

function substepsForAction(definition, includePreflight) {
  const substeps = includePreflight
    ? [preflightSubstep, ...definition.substeps]
    : definition.substeps;
  return substeps.map(baseSubstep);
}

function baseAction(definition, includePreflight, resultsDirInput) {
  const selected = Boolean(resultsDirInput) || !definition.requiresResultsDir;
  const substeps = substepsForAction(definition, includePreflight);
  if (!selected) {
    for (const substep of substeps) {
      substep.status = "skipped";
      substep.skipped_reason = "results-dir-not-provided";
    }
  }
  return {
    action_id: definition.actionID,
    description: definition.description,
    requires_results_dir: definition.requiresResultsDir,
    mutating: definition.mutating,
    status: selected ? "pending" : "skipped",
    execution_state: selected ? "pending" : "not_selected",
    started_at: null,
    completed_at: null,
    duration_ms: null,
    skipped_reason: selected ? null : "results-dir-not-provided",
    substeps,
  };
}

export function selectedActionDefinitions(actionRegistry, resultsDirInput) {
  const actionByID = new Map(
    actionRegistry.map((action) => [action.actionID, action]),
  );
  const order = resultsDirInput
    ? [
      "scheduler_drift_validation",
      "schema_shape_validation",
      "tier_coverage_validation",
      "generated_structure_refresh",
      "canonical_evidence_validation",
    ]
    : [
      "schema_shape_validation",
      "tier_coverage_validation",
      "generated_structure_refresh",
      "canonical_evidence_validation",
      "scheduler_drift_validation",
    ];
  const selected = order.map((actionID) => actionByID.get(actionID));
  if (
    actionByID.size !== actionRegistry.length ||
    actionByID.size !== order.length ||
    selected.some((action) => action === undefined)
  ) {
    throw new Error("agent-finalize action registry is incomplete");
  }
  return selected;
}

export function selectedActions(actionRegistry, resultsDirInput) {
  const definitions = selectedActionDefinitions(actionRegistry, resultsDirInput);
  return definitions.map((definition, index) =>
    baseAction(definition, Boolean(resultsDirInput) && index === 0, resultsDirInput),
  );
}

export function flattenSubsteps(actions) {
  return actions.flatMap((action) => action.substeps);
}

export function collectChildArtifacts(actions) {
  const artifacts = [];
  for (const action of actions) {
    for (const substep of action.substeps) {
      if (substep.summary_json) {
        artifacts.push({
          role: `${action.action_id}_${substep.id}_summary`,
          kind: "json",
          path: substep.summary_json,
        });
      }
      if (substep.stdout_log) {
        artifacts.push({
          role: `${action.action_id}_${substep.id}_stdout`,
          kind: "log",
          path: substep.stdout_log,
        });
      }
      if (substep.stderr_log) {
        artifacts.push({
          role: `${action.action_id}_${substep.id}_stderr`,
          kind: "log",
          path: substep.stderr_log,
        });
      }
    }
  }
  return artifacts.sort((left, right) =>
    `${left.role}\0${left.path_kind}\0${left.format ?? ""}\0${left.path}`.localeCompare(
      `${right.role}\0${right.path_kind}\0${right.format ?? ""}\0${right.path}`,
    ),
  );
}
