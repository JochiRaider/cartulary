export const browserBatchManifestSchemaID = "cartulary.browser_e2e_batch_manifest.v15";

// Execution compatibility is independent of catalog reporting ownership. Keep
// this projection shared by generation, compilation and runtime admission.
const selectorStages = new Map([
  ["webserver-backed", "webserver_backed"],
  ["duration_balanced_specs", "webserver_backed"],
  ["support", "support"],
  ["stateful", "stateful"],
  ["stateful_partition", "stateful"],
  ["measurement", "measurement"],
  ["a11y", "accessibility"],
  ["visual", "visual"],
]);

export function browserSelectorStage(kind) {
  const stage = selectorStages.get(kind);
  if (!stage) throw new Error(`browser kind ${kind} has no selector-stage mapping`);
  return stage;
}

export function browserPartitionContract(stage, group, runtimeProfile) {
  return {
    stageName: stage.name,
    stageTarget: stage.target,
    frontendArtifactID: stage.frontend_artifact_id ?? "",
    stageScheduleTags: [...(stage.schedule_tags ?? [])].sort(),
    stageSchedulerNeeds: [...(stage.scheduler_needs ?? [])].sort(),
    target: group.target,
    kind: group.kind,
    coverage: group.coverage ?? (group.kind === "support" ? "supplemental" : ""),
    executionDependency: group.execution_dependency ?? "",
    dependencyTarget: group.dependency_target ?? "",
    workers: group.workers,
    stateScope: group.state_scope,
    project: group.project,
    config: group.config,
    runtimeProfileID: group.runtime_profile_id,
    runtimeKind: runtimeProfile.kind,
    keyRingManifestPath: runtimeProfile.key_ring_manifest_path ?? "",
    resourceProfileID: group.resource_profile_id,
    fixtureCapability: group.fixture_capability,
    fixtureProfileID: group.fixture_profile_id ?? "",
    serviceRequirement: group.service_requirement,
    serviceDependencies: [...group.service_dependencies],
    postgresIsolation: group.postgres_isolation,
    preparationOperations: [...group.preparation_operations],
    browserSessionGroup: group.browser_session_group,
  };
}

export function browserCompatibility(partition, policy, row) {
  return {
    stage: partition.stageName,
    stage_target: partition.stageTarget,
    frontend_artifact: partition.frontendArtifactID,
    schedule_tags: [...partition.stageScheduleTags].sort(),
    scheduler_needs: [...partition.stageSchedulerNeeds].sort(),
    target: partition.target,
    kind: partition.kind,
    coverage: partition.coverage,
    execution_dependency: partition.executionDependency,
    dependency_target: partition.dependencyTarget,
    selector_stage: row ? row.selector.stage : browserSelectorStage(partition.kind),
    project: row ? row.selector.project_id : partition.project,
    config: partition.config,
    runtime_profile: row ? row.runtime_profile_id : partition.runtimeProfileID,
    runtime_kind: partition.runtimeKind,
    key_ring_manifest: partition.keyRingManifestPath,
    resource_profile: row ? row.resource_profile_id : partition.resourceProfileID,
    fixture_capability: row ? row.fixture_capability : partition.fixtureCapability,
    fixture_profile: row ? (row.fixture_profile_id ?? "") : partition.fixtureProfileID,
    service_requirement: partition.serviceRequirement,
    service_dependencies: [...(row ? row.service_dependencies : partition.serviceDependencies)].sort(),
    postgres_isolation: row ? row.postgres_isolation : partition.postgresIsolation,
    preparation_operations: [...(row ? (row.preparation_operations ?? []) : partition.preparationOperations)].sort(),
    state_scope: row ? row.browser_state_scope : partition.stateScope,
    singleton_row: partition.stateScope === "test_owned" ? null : (row?.row_id ?? partition.selectedRowIDs[0]),
    browser_session_group: partition.browserSessionGroup,
    invocation_policy: { startup_ms: policy.startup_ms, cleanup_ms: policy.cleanup_ms },
    workers: partition.workers,
    max_failures: 1,
    retries: 0,
  };
}
