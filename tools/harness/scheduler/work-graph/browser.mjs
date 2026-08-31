import { readFileSync } from "node:fs";
import path from "node:path";

import { loadBrowserBatchStages } from "../adapters/browser.mjs";
import {
  activePerformanceFixtureProfile,
  loadPerformanceFixtureBuilderPolicy,
  loadPerformanceFixtureSnapshotRegistry,
  performanceFixtureBindingsForRows,
  postgresMigrationDigest,
  snapshotKey,
} from "../../performance-fixture/index.mjs";
import { loadTestCatalog } from "../../test-catalog/index.mjs";
import { buildWorkGraph } from "./model.mjs";
import {
  assertFixtureServiceDependencies,
  topologyResourceClaims,
} from "./resource-claims.mjs";
import { planBrowserFunctionalLanes } from "./browser-functional-lanes.mjs";

const manifestRelativePath = "tools/browser_e2e_batch_manifest.json";

function topology(root) {
  return JSON.parse(
    readFileSync(path.join(root, "tools/execution_topology_manifest.json"), "utf8"),
  );
}

function resourceProfile(root, group) {
  const topology = JSON.parse(
    readFileSync(path.join(root, "tools/execution_topology_manifest.json"), "utf8"),
  );
  const profile = topology.resource_profiles.find(
    (entry) => entry.id === group.resourceProfileID,
  );
  if (!profile) {
    throw new Error(`browser group ${group.name} has unknown resource profile ${group.resourceProfileID}`);
  }
  if (
    profile.runner_timeout_ms !== undefined &&
    (!Number.isSafeInteger(profile.runner_timeout_ms) || profile.runner_timeout_ms < 1)
  ) {
    throw new Error(`browser resource profile ${profile.id} has invalid runner_timeout_ms`);
  }
  return profile;
}

function resourceClaims(root, group) {
  assertFixtureServiceDependencies("browser_stack", group.serviceDependencies, `browser group ${group.name}`);
  return topologyResourceClaims(
    topology(root),
    group.resourceProfileID,
    group.serviceDependencies,
    `browser group ${group.name}`,
  );
}

function hostLocks(group) {
  return group.resourceProfileID === "browser_measurement_quiet"
    ? { shared: [], exclusive: ["host_activity"] }
    : { shared: ["host_activity"], exclusive: [] };
}

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function safeID(value) {
  return value.replaceAll(/[^A-Za-z0-9_.+-]+/gu, "-");
}

function requiredFailurePolicy() {
  return {
    block_descendants: true,
    continue_independent: true,
    aggregate_effect: "required",
  };
}

function finalizableFailurePolicy() {
  return {
    block_descendants: false,
    continue_independent: true,
    aggregate_effect: "required",
  };
}

function command(executable, args, environment = {}) {
  return { executable, args, environment };
}

function resolvedFixtureProfile(root, group) {
  if (!group.fixtureProfileID) return null;
  const registry = loadPerformanceFixtureSnapshotRegistry(root);
  const profile = activePerformanceFixtureProfile(registry, group.fixtureProfileID);
  const builderPolicy = loadPerformanceFixtureBuilderPolicy(root, { registry })
    .byFixtureProfileID.get(group.fixtureProfileID);
  if (!builderPolicy) {
    throw new Error("profiled browser group " + group.name + " has no builder policy");
  }
  const migrationDigest = postgresMigrationDigest(root);
  const key = snapshotKey(profile, migrationDigest);
  if (group.selectedRowIDs.length !== 1) {
    throw new Error("profiled browser group " + group.name + " must select exactly one row");
  }
  const rowID = group.selectedRowIDs[0];
  const row = loadTestCatalog(root).rowByID.get(rowID);
  const bindings = row === undefined
    ? []
    : performanceFixtureBindingsForRows(root, [row], { registry });
  if (bindings.length !== 1) {
    throw new Error("profiled browser row " + rowID + " must resolve exactly one predicate binding");
  }
  return {
    profile,
    builderPolicy,
    migrationDigest,
    snapshotKey: key,
    builderUnitID:
      "fixture_snapshot:" + builderPolicy.runtime_profile_id + ":" +
      group.fixtureProfileID + ":" + key,
    rowID,
    predicateID: bindings[0].predicate_id,
  };
}

function snapshotBuilderUnit(root, group, fixture, owner) {
  const policy = fixture.builderPolicy;
  assertFixtureServiceDependencies(
    policy.fixture_capability,
    policy.service_dependencies,
    `performance fixture builder ${fixture.profile.fixture_profile_id}`,
  );
  return {
    unit_id: fixture.builderUnitID,
    owner_id: "harness.browser",
    kind: "fixture_builder",
    command: command(
      "node",
      [
        "tools/harness/performance-fixture/snapshot-builder-cli.mjs",
        "--fixture-profile",
        group.fixtureProfileID,
        "--snapshot-key",
        fixture.snapshotKey,
      ],
      {
        CARTULARY_BROWSER_RUNTIME_PROFILE_ID: policy.runtime_profile_id,
        CARTULARY_FIXTURE_BUILDER_RESOURCE_PROFILE_ID: policy.resource_profile_id,
        CARTULARY_FIXTURE_PROFILE_ID: group.fixtureProfileID,
        CARTULARY_FIXTURE_SNAPSHOT_KEY: fixture.snapshotKey,
        CARTULARY_FIXTURE_MIGRATION_DIGEST: fixture.migrationDigest,
        CARTULARY_FIXTURE_SOURCE_CONTRACT_DIGEST:
          fixture.profile.source_contract_digest,
        CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID: fixture.builderUnitID,
      },
    ),
    needs: [],
    resource_claims: topologyResourceClaims(
      topology(root),
      policy.resource_profile_id,
      policy.service_dependencies,
      `performance fixture builder ${fixture.profile.fixture_profile_id}`,
    ),
    shared_locks: ["host_activity"],
    exclusive_locks: [],
    fixture_profile_id: group.fixtureProfileID,
    snapshot_key: fixture.snapshotKey,
    fixture_lease: policy.fixture_capability,
    service_dependencies: policy.service_dependencies,
    cache_policy: "none",
    timeout_ms: owner.default_timeout_ms,
    current_run_evidence_outputs: [
      "performance-fixtures/" + fixture.snapshotKey + "/build-diagnostics.json",
      "performance-fixtures/" + fixture.snapshotKey + "/snapshot-build.json",
    ],
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms: 120000,
  };
}

function lifecycleKey(stage, group) {
  return [
    stage.name,
    group.browserSessionGroup,
    group.runtimeProfileID,
    group.resourceProfileID,
    group.fixtureProfileID ?? "mutable",
  ].join("-");
}

function resetPolicy(root) {
  const policy = topology(root).browser_reset_policy;
  if (
    policy?.max_attempts !== 1 ||
    ![
      policy.backend_drain_timeout_ms,
      policy.database_reset_timeout_ms,
      policy.backend_readiness_timeout_ms,
      policy.evidence_overhead_timeout_ms,
    ].every((value) => Number.isSafeInteger(value) && value > 0)
  ) {
    throw new Error("browser reset policy must declare one attempt and positive stage deadlines");
  }
  return policy;
}

function resetUnit(root, stage, group, dependencies, fixture, resetLabel) {
  const key = safeID(lifecycleKey(stage, group));
  const locks = hostLocks(group);
  const policy = resetPolicy(root);
  return {
    unit_id: `browser_reset:${safeID(stage.name)}:${safeID(resetLabel)}`,
    owner_id: "harness.browser",
    kind: "lifecycle",
    command: command("tools/harness/browser/reset-web-e2e-stack.sh", [
      "--label",
      resetLabel,
    ], {
      CARTULARY_TEST_TARGET: group.target,
      CARTULARY_MAKE_INPUT_SOURCES: "",
	  OWNER: "",
	  ROWS: "",
	  SERVICE_BACKED_ONLY: "",
      CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
      CARTULARY_BROWSER_RESOURCE_PROFILE_ID: group.resourceProfileID,
      ...(group.functionalLaneID
        ? {
            CARTULARY_BROWSER_FUNCTIONAL_LANE_ID: group.functionalLaneID,
            CARTULARY_BROWSER_GROUP_GENERATION: String(group.functionalGeneration ?? 1),
          }
        : {}),
      CARTULARY_BROWSER_SERVICE_REQUIREMENT: group.serviceRequirement,
      CARTULARY_HARNESS_SERVICE_DEPENDENCIES: group.serviceDependencies.join(","),
      CARTULARY_BROWSER_SESSION_CONTRACT: group.browserSessionGroup,
      CARTULARY_BROWSER_RESET_DRAIN_TIMEOUT_MS: String(policy.backend_drain_timeout_ms),
      ...(fixture
        ? {
            CARTULARY_FIXTURE_PROFILE_ID: group.fixtureProfileID,
            CARTULARY_FIXTURE_SNAPSHOT_KEY: fixture.snapshotKey,
            CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID: fixture.builderUnitID,
            CARTULARY_FIXTURE_ROW_ID: fixture.rowID,
            CARTULARY_FIXTURE_PREDICATE_ID: fixture.predicateID,
          }
        : {}),
    }),
    needs: dependencies,
    resource_claims: resourceClaims(root, group),
    shared_locks: locks.shared,
    exclusive_locks: [...locks.exclusive, `browser_session:${key}`].sort(compareASCII),
    affinity_key: key,
    ...(fixture
      ? {
          fixture_profile_id: group.fixtureProfileID,
          snapshot_key: fixture.snapshotKey,
        }
      : {}),
    fixture_lease: "browser_stack",
    service_dependencies: group.serviceDependencies,
    cache_policy: "none",
    timeout_ms:
      policy.backend_drain_timeout_ms +
      policy.database_reset_timeout_ms +
      policy.backend_readiness_timeout_ms +
      policy.evidence_overhead_timeout_ms,
    current_run_evidence_outputs: [
      `${group.target}/reset-boundary/${safeID(resetLabel)}.attempt.json`,
    ],
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms: 1000,
  };
}

function groupLocks(group, mode) {
  const shared = [];
  const exclusive = [];
  const host = hostLocks(group);
  shared.push(...host.shared);
  exclusive.push(...host.exclusive);
  if (group.kind === "visual" && mode === "snapshot_update") {
    exclusive.push("browser_snapshots");
  } else if (group.kind === "visual") {
    shared.push("browser_snapshots");
  }
  return {
    shared: shared.sort(compareASCII),
    exclusive: exclusive.sort(compareASCII),
  };
}

function groupUnit(root, stage, group, dependencies, owner, mode, fixture) {
  const key = safeID(lifecycleKey(stage, group));
  const locks = groupLocks(group, mode);
  const target = mode === "snapshot_update" ? "browser-e2e-visual-update" : group.target;
  return {
    unit_id: `browser_group:${safeID(stage.name)}:${safeID(group.name)}`,
    owner_id: "harness.browser",
    kind: "runner",
    command: command(
      "node",
      [
        "tools/harness/browser/browser-catalog-group-cli.mjs",
        "--manifest",
        manifestRelativePath,
        "--stage",
        stage.name,
        "--group",
        group.name,
      ],
      {
        BROWSER_E2E_BATCH_MANIFEST: manifestRelativePath,
        CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
        CARTULARY_BROWSER_RESOURCE_PROFILE_ID: group.resourceProfileID,
        CARTULARY_BROWSER_SELECTED_ROW_IDS: group.selectedRowIDs.join(","),
        ...(group.functionalLaneID
          ? {
              CARTULARY_BROWSER_FUNCTIONAL_LANE_ID: group.functionalLaneID,
              CARTULARY_BROWSER_GROUP_GENERATION: String(group.functionalGeneration),
            }
          : {}),
        CARTULARY_BROWSER_SERVICE_REQUIREMENT: group.serviceRequirement,
        CARTULARY_HARNESS_SERVICE_DEPENDENCIES: group.serviceDependencies.join(","),
        CARTULARY_BROWSER_SESSION_CONTRACT: group.browserSessionGroup,
        CARTULARY_TEST_TARGET: target,
        ...(fixture
          ? {
              CARTULARY_FIXTURE_PROFILE_ID: group.fixtureProfileID,
              CARTULARY_FIXTURE_SNAPSHOT_KEY: fixture.snapshotKey,
              CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID: fixture.builderUnitID,
              CARTULARY_FIXTURE_ROW_ID: fixture.rowID,
              CARTULARY_FIXTURE_PREDICATE_ID: fixture.predicateID,
            }
          : {}),
        ...(mode === "snapshot_update"
          ? {
              CARTULARY_BROWSER_MAINTENANCE_MODE: "snapshot_update",
              CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS: "1",
            }
          : {}),
      },
    ),
    needs: dependencies,
    resource_claims: resourceClaims(root, group),
    shared_locks: locks.shared,
    exclusive_locks: [...locks.exclusive, `browser_session:${key}`].sort(compareASCII),
    affinity_key: key,
    ...(fixture
      ? {
          fixture_profile_id: group.fixtureProfileID,
          snapshot_key: fixture.snapshotKey,
        }
      : {}),
    fixture_lease: "browser_stack",
    service_dependencies: group.serviceDependencies,
    cache_policy: "none",
    timeout_ms:
      resourceProfile(root, group).runner_timeout_ms ?? owner.default_timeout_ms,
    current_run_evidence_outputs: [
      ...group.selectedRowIDs.map((rowID) => `rows/${rowID}.json`),
      `${target}/browser-groups/${safeID(group.name)}/browser-group-result.json`,
      ...(group.kind === "visual"
        ? [`${target}/browser-groups/${safeID(group.name)}/renderer-profile-attestation.json`]
        : []),
    ].sort(compareASCII),
    failure_policy: finalizableFailurePolicy(),
    estimated_work_ms: group.estimatedWorkMs ??
      owner.evidence_estimates_ms[group.kind === "a11y" ? "accessibility" : group.kind] ??
      owner.evidence_estimates_ms.browser,
  };
}

function measurementSummaryUnit(root, stage, group, runner, owner, fixture, mode) {
  const target = mode === "snapshot_update" ? "browser-e2e-visual-update" : group.target;
  return {
    unit_id: `browser_measurement_summary:${safeID(stage.name)}:${safeID(group.name)}`,
    owner_id: "harness.browser",
    kind: "finalizer",
    command: command(
      "node",
      [
        "tools/harness/browser/browser-measurement-finalizer-cli.mjs",
        "--target",
        target,
        "--stage",
        stage.name,
        "--group",
        group.name,
        "--row",
        fixture.rowID,
        "--predicate",
        fixture.predicateID,
      ],
    ),
    needs: [runner.unit_id],
    resource_claims: topologyResourceClaims(
      topology(root),
      "standard",
      [],
      `browser measurement row finalizer ${fixture.rowID}`,
    ),
    shared_locks: ["host_activity"],
    exclusive_locks: [],
    fixture_profile_id: group.fixtureProfileID,
    snapshot_key: fixture.snapshotKey,
    fixture_lease: "none",
    service_dependencies: [],
    cache_policy: "none",
    timeout_ms: owner.default_timeout_ms,
    current_run_evidence_outputs: [
      `${target}/browser-groups/${safeID(group.name)}/frontend-measurement-summary.${fixture.profile.artifact_policy.summary_schema_id.split(".").at(-1)}.json`,
    ],
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms: 500,
  };
}

function targetFinalizer(root, stage, target, groupUnits, resetUnits, needs, owner) {
  const groupTargets = stage.groups
    .map((group) => `${group.name}=${target === "browser-e2e-visual-update" ? target : group.target}`)
    .sort(compareASCII);
  const quiet = stage.groups.every(
    (group) => group.resourceProfileID === "browser_measurement_quiet",
  );
  const includesFixtureMeasurements = stage.groups.some((group) => group.fixtureProfileID);
  return {
    unit_id: `browser_target_summary:${safeID(target)}`,
    owner_id: "harness.browser",
    kind: "finalizer",
    command: command(
      "node",
      [
        "tools/harness/browser/browser-target-finalizer-cli.mjs",
        "--target",
        target,
        "--groups",
        stage.groups.map((group) => group.name).sort(compareASCII).join(","),
        "--group-targets",
        groupTargets.join(","),
        "--children",
        target === stage.target ? stage.summaryChildren.join(",") : "",
        "--reset-prefix",
        `browser_reset:${safeID(stage.name)}:`,
      ],
      { CARTULARY_DEFER_OBSERVABILITY_FINALIZE: "1" },
    ),
    needs: needs.sort(compareASCII),
    resource_claims: topologyResourceClaims(
      topology(root),
      "standard",
      [],
      `browser target finalizer ${target}`,
    ),
    shared_locks: quiet ? [] : ["host_activity"],
    exclusive_locks: quiet ? ["host_activity"] : [],
    fixture_lease: "none",
    service_dependencies: [],
    cache_policy: "none",
    timeout_ms: owner.default_timeout_ms,
    current_run_evidence_outputs: [
      `${target}/browser-target-result.json`,
      ...(target === "browser-e2e-measurement" && includesFixtureMeasurements
        ? [`${target}/frontend-measurement-aggregate.json`]
        : []),
      ...(target === "browser-e2e-visual" || target === "browser-e2e-visual-update"
        ? [`${target}/frontend-visual-reconciliation.json`]
        : []),
    ],
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms: 500,
  };
}

export function browserStages(root) {
  return loadBrowserBatchStages(path.join(root, manifestRelativePath));
}

export function browserTargetStage(root, target) {
  if (target === "browser-e2e-visual-update") {
    return { stage: browserStages(root).get("visual"), mode: "snapshot_update" };
  }
  const stage = [...browserStages(root).values()].find((entry) => entry.target === target);
  return stage ? { stage, mode: "validation" } : null;
}

export function compileBrowserStageGraph(root, owner, stage, { mode = "validation" } = {}) {
  const units = [];
  const previousByLifecycle = new Map();
  const previousGroupByLifecycle = new Map();
  const previousResetByLifecycle = new Map();
  const snapshotBuilderIDs = new Set();
  const groupUnits = [];
  const resetUnits = [];
  const measurementSummaryUnits = [];
  let scheduledGroups = stage.groups;
  if (stage.groups.every(
    (group) =>
      group.kind === "duration_balanced_specs" &&
      group.resourceProfileID === "browser_functional",
  )) {
    const catalog = loadTestCatalog(root);
    const lanes = planBrowserFunctionalLanes(stage.groups, {
      lanePrefix: stage.name,
      maxLanes: 4,
      estimateGroup(group) {
        return group.selectedRowIDs.reduce((total, rowID) => {
          const row = catalog.rowByID.get(rowID);
          if (!row) {
            throw new Error(`browser functional group ${group.name} references unknown row ${rowID}`);
          }
          const estimate = owner.evidence_estimates_ms[row.evidence_class];
          if (!Number.isSafeInteger(estimate) || estimate < 1) {
            throw new Error(
              `browser functional row ${rowID} has no positive evidence estimate`,
            );
          }
          return total + estimate;
        }, 0);
      },
    });
    scheduledGroups = lanes.flatMap((lane) =>
      lane.groups.map((item) => ({
        ...item.group,
        estimatedWorkMs: item.estimatedWorkMs,
        functionalLaneID: lane.laneID,
        functionalGeneration: item.generation,
      })),
    );
  }
  for (const group of scheduledGroups) {
    const fixture = resolvedFixtureProfile(root, group);
    if (fixture && !snapshotBuilderIDs.has(fixture.builderUnitID)) {
      units.push(snapshotBuilderUnit(root, group, fixture, owner));
      snapshotBuilderIDs.add(fixture.builderUnitID);
    }
    const key = safeID(lifecycleKey(stage, group));
    const previousID = previousByLifecycle.get(key);
    const previousGroupID = previousGroupByLifecycle.get(key);
    let dependencies = previousID
      ? [previousID]
      : fixture
        ? [fixture.builderUnitID]
        : [];
    const resetLabel = previousID
      ? `${safeID(previousGroupID)}--before-${safeID(group.name)}`
      : "";
    if (previousID) {
      const previousResetID = previousResetByLifecycle.get(key);
      const resetDependencies = [previousID, ...(previousResetID ? [previousResetID] : [])]
        .sort(compareASCII);
      const reset = resetUnit(
        root,
        stage,
        group,
        resetDependencies,
        fixture,
        resetLabel,
      );
      units.push(reset);
      resetUnits.push(reset);
      previousResetByLifecycle.set(key, reset.unit_id);
      dependencies = [reset.unit_id];
    }
    const runner = groupUnit(root, stage, group, dependencies, owner, mode, fixture);
    units.push(runner);
    groupUnits.push(runner);
    if (fixture) {
      const summary = measurementSummaryUnit(
        root,
        stage,
        group,
        runner,
        owner,
        fixture,
        mode,
      );
      units.push(summary);
      measurementSummaryUnits.push(summary);
    }
    previousByLifecycle.set(key, runner.unit_id);
    previousGroupByLifecycle.set(key, group.name);
  }

  for (const terminalUnitID of previousByLifecycle.values()) {
    const terminalUnit = units.find((unit) => unit.unit_id === terminalUnitID);
    terminalUnit.command.environment.CARTULARY_BROWSER_RELEASE_AFFINITY = "1";
  }

  const target = mode === "snapshot_update" ? "browser-e2e-visual-update" : stage.target;
  units.push(
    targetFinalizer(
      root,
      stage,
      target,
      groupUnits,
      resetUnits,
      [
        ...groupUnits.map((unit) => unit.unit_id),
        ...resetUnits.map((unit) => unit.unit_id),
        ...measurementSummaryUnits.map((unit) => unit.unit_id),
      ],
      owner,
    ),
  );
  return buildWorkGraph(units);
}

export function compileBrowserRowSelectionGraph(root, owner, rowIDs) {
  const remaining = new Set(rowIDs);
  const units = [];
  for (const stage of browserStages(root).values()) {
    const groups = stage.groups.flatMap((group) => {
      const selectedRowIDs = group.selectedRowIDs.filter((rowID) =>
        remaining.has(rowID),
      );
      if (selectedRowIDs.length === 0) return [];
      for (const rowID of selectedRowIDs) remaining.delete(rowID);
      return [{ ...group, selectedRowIDs }];
    });
    if (groups.length === 0) continue;
    units.push(
      ...compileBrowserStageGraph(
        root,
        owner,
        { ...stage, groups, summaryChildren: [] },
        { mode: "validation" },
      ).units,
    );
  }
  if (remaining.size > 0) {
    throw new Error(
      `Playwright rows are missing generated browser groups: ${[
        ...remaining,
      ].join(", ")}`,
    );
  }
  return buildWorkGraph(units);
}
