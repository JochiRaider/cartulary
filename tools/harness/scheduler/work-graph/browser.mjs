import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";

import { loadBrowserBatchStages, packBrowserCohorts } from "../adapters/browser.mjs";
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
import { planBrowserLanes } from "./browser-lanes.mjs";

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

function resolvedFixtureProfile(root, group, catalog) {
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
  const row = catalog.rowByID.get(rowID);
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
    postgres_isolation: policy.postgres_isolation,
    cache_policy: "none",
    timeout_ms: owner.default_timeout_ms,
    current_run_evidence_outputs: [
      "performance-fixtures/" + fixture.snapshotKey + "/build-diagnostics.json",
      "performance-fixtures/" + fixture.snapshotKey + "/snapshot-build.json",
    ],
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms: owner.lifecycle_estimates_ms.performance_fixture_prepare,
  };
}

function lifecycleKey(stage, group) {
  return `${group.executionKey ?? group.partitionName}-${group.laneID ?? group.name}`;
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
    renderer_requirement: group.resourceProfileID === "browser_measurement_quiet" ? "dedicated" : "pooled",
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
        "--partition",
        group.partitionName,
        "--group",
        group.name,
      ],
      {
        BROWSER_E2E_BATCH_MANIFEST: manifestRelativePath,
        CARTULARY_BROWSER_STAGE: stage.name,
        CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
        CARTULARY_BROWSER_RESOURCE_PROFILE_ID: group.resourceProfileID,
        ...(group.laneID ? { CARTULARY_BROWSER_LANE_ID: group.laneID } : {}),
        CARTULARY_BROWSER_PREPARATION_LABEL: group.name,
        CARTULARY_BROWSER_RESET_DRAIN_TIMEOUT_MS: String(resetPolicy(root).backend_drain_timeout_ms),
        CARTULARY_BROWSER_PREPARATION_TIMEOUT_MS: String(Object.entries(resetPolicy(root))
          .filter(([key]) => key.endsWith("_timeout_ms")).reduce((total, [, value]) => total + value, 0)),
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
    postgres_isolation: group.postgresIsolation,
    cache_policy: "none",
    timeout_ms:
      resourceProfile(root, group).runner_timeout_ms ?? owner.default_timeout_ms,
    current_run_evidence_outputs: [
      ...group.selectedRowIDs.map((rowID) => `rows/${rowID}.json`),
      `${target}/browser-groups/${safeID(group.name)}/browser-group-result.json`,
    ].sort(compareASCII),
    row_bindings: group.selectedRowIDs.map((row_id) => ({ row_id, runner: "playwright" })),
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
    estimated_work_ms: owner.lifecycle_estimates_ms.measurement_finalize,
  };
}

function targetFinalizer(root, stage, target, groupUnits, needs, owner) {
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
        ...stage.groups.map((group) => group.name).sort(compareASCII).flatMap((name) => ["--groups", name]),
        // The finalizer accumulates repeated mappings. Keep each argument
        // bounded as the selected browser group inventory grows.
        ...groupTargets.flatMap((entry) => ["--group-targets", entry]),
        "--children",
        target === stage.target ? stage.summaryChildren.join(",") : "",
        ...stage.groups.flatMap((group) => ["--group-units",
          `${group.name}=browser_group:${safeID(group.stageName)}:${safeID(group.name)}`]),
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
    estimated_work_ms: owner.lifecycle_estimates_ms.browser_target_finalize,
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

export function compileBrowserStageGraph(root, owner, stage, options = {}) {
  return buildWorkGraph(compileBrowserStagesUnits(root, owner, [stage], options), topology(root));
}

export function compileBrowserStagesUnits(root, owner, stages, {
  mode = "validation", catalog = loadTestCatalog(root), maxLanes = topology(root).browser_lane_policy.default,
  needsForRow = () => [],
} = {}) {
  const units = [], partitions = new Map(), seenRows = new Set();
  for (const stage of stages) for (const group of stage.groups) {
    for (const rowID of group.selectedRowIDs) {
      if (seenRows.has(rowID)) throw new Error(`duplicate selected browser row ${rowID}`);
      seenRows.add(rowID);
      const prerequisites = needsForRow(catalog.rowByID.get(rowID));
      const key = JSON.stringify([group.name, prerequisites]);
      const selected = partitions.get(key) ?? { ...group, selectedRowIDs: [], prerequisites, executionKey: createHash("sha256").update(key).digest("hex") };
      selected.selectedRowIDs.push(rowID);
      partitions.set(key, selected);
    }
  }
  if (!partitions.size) return [];
  const policy = stages[0].invocationPolicy;
  if (stages.some((stage) => JSON.stringify(stage.invocationPolicy) !== JSON.stringify(policy))) {
    throw new Error("selected browser closure has conflicting invocation policies");
  }
  const groups = packBrowserCohorts([...partitions.values()], catalog, policy, owner,
    (group) => resourceProfile(root, group).runner_timeout_ms ?? owner.default_timeout_ms);
  const ordinary = groups.filter((group) => group.resourceProfileID === "browser_ordinary");
  const scheduled = ordinary.length ? planBrowserLanes(ordinary, {
    maxLanes, estimateGroup: (group) => group.estimatedWorkMs,
  }).flatMap((lane) => lane.groups.map(({ group }) => ({ ...group, laneID: lane.laneID }))) : [];
  scheduled.push(...groups.filter((group) => group.resourceProfileID !== "browser_ordinary"));
  const previousByLane = new Map(), previousByAffinity = new Map(), builders = new Set();
  const runners = [], summaries = [];
  for (const group of scheduled) {
    const stage = { name: group.stageName };
    const fixture = resolvedFixtureProfile(root, group, catalog);
    if (fixture && !builders.has(fixture.builderUnitID)) {
      const builder = snapshotBuilderUnit(root, group, fixture, owner);
      builder.needs = group.prerequisites;
      units.push(builder); builders.add(fixture.builderUnitID);
    }
    const key = safeID(lifecycleKey(stage, group));
    const previous = previousByLane.get(group.laneID);
    const needs = [...new Set([...group.prerequisites, ...(previous ? [previous] : []),
      ...(fixture ? [fixture.builderUnitID] : [])])].sort(compareASCII);
    const runner = groupUnit(root, stage, group, needs, owner, mode, fixture);
    // This is an estimate, never an assertion of allocation generation/readiness.
    runner.estimated_work_ms += previousByAffinity.has(key)
      ? owner.lifecycle_estimates_ms.browser_reset : owner.lifecycle_estimates_ms.browser_allocation_acquire;
    units.push(runner); runners.push(runner);
    if (fixture) {
      const summary = measurementSummaryUnit(root, stage, group, runner, owner, fixture, mode);
      units.push(summary); summaries.push(summary);
    }
    previousByAffinity.set(key, runner);
    if (group.laneID) previousByLane.set(group.laneID, runner.unit_id);
  }
  for (const runner of previousByAffinity.values()) runner.estimated_work_ms += owner.lifecycle_estimates_ms.browser_allocation_release;
  for (const stage of stages) {
    const selected = new Set(stage.groups.flatMap((group) => group.selectedRowIDs));
    const reportingGroups = groups.filter((group) => group.selectedRowIDs.some((id) => selected.has(id)));
    const reportingRunners = runners.filter((unit) => unit.row_bindings.some(({ row_id }) => selected.has(row_id)));
    const needs = [...reportingRunners.map((unit) => unit.unit_id),
      ...summaries.filter((unit) => reportingRunners.some((runner) => unit.needs.includes(runner.unit_id))).map((unit) => unit.unit_id)];
    units.push(targetFinalizer(root, { ...stage, groups: reportingGroups },
      mode === "snapshot_update" ? "browser-e2e-visual-update" : stage.target, reportingRunners, needs, owner));
  }
  return units;
}

export function selectBrowserRowStages(root, rowIDs) {
  const remaining = new Set(rowIDs);
  const selectedStages = [];
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
    selectedStages.push({ ...stage, groups, summaryChildren: [] });
  }
  if (remaining.size > 0) {
    throw new Error(
      `Playwright rows are missing generated browser groups: ${[
        ...remaining,
      ].join(", ")}`,
    );
  }
  return selectedStages;
}

export function compileBrowserRowSelectionGraph(root, owner, rowIDs, { catalog } = {}) {
  return buildWorkGraph(compileBrowserStagesUnits(root, owner, selectBrowserRowStages(root, rowIDs), { catalog }), topology(root));
}
