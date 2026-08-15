import { readFileSync } from "node:fs";
import path from "node:path";

import { loadBrowserBatchStages } from "../adapters/browser.mjs";
import {
  activePerformanceFixtureProfile,
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
    migrationDigest,
    snapshotKey: key,
    builderUnitID:
      "fixture_snapshot:" + group.runtimeProfileID + ":" +
      group.fixtureProfileID + ":" + key,
    rowID,
    predicateID: bindings[0].predicate_id,
  };
}

function snapshotBuilderUnit(root, group, fixture, owner) {
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
        CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
        CARTULARY_FIXTURE_PROFILE_ID: group.fixtureProfileID,
        CARTULARY_FIXTURE_SNAPSHOT_KEY: fixture.snapshotKey,
        CARTULARY_FIXTURE_MIGRATION_DIGEST: fixture.migrationDigest,
        CARTULARY_FIXTURE_SOURCE_CONTRACT_DIGEST:
          fixture.profile.source_contract_digest,
        CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID: fixture.builderUnitID,
      },
    ),
    needs: [],
    resource_claims: resourceClaims(root, group),
    shared_locks: ["host_activity"],
    exclusive_locks: [],
    fixture_profile_id: group.fixtureProfileID,
    snapshot_key: fixture.snapshotKey,
    fixture_lease: "postgres_dedicated",
    service_dependencies: group.serviceDependencies,
    cache_policy: "none",
    timeout_ms: owner.default_timeout_ms,
    evidence_outputs: [
      "performance-fixtures/" + fixture.snapshotKey + "/snapshot-build.json",
    ],
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms: 120000,
  };
}

function lifecycleKey(stage, group) {
  if (group.kind === "stateful_partition") {
    return `${stage.name}-${group.browserSessionGroup}-${group.runtimeProfileID}`;
  }
  return `${stage.name}-${group.name}-${group.runtimeProfileID}`;
}

function lifecycleUnit(root, stage, group, owner, fixture) {
  const key = safeID(lifecycleKey(stage, group));
  const locks = hostLocks(group);
  return {
    unit_id: `browser_lifecycle:${key}`,
    owner_id: "harness.browser",
    kind: "readiness",
    command: command("true", [], {
      CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
      CARTULARY_BROWSER_RESOURCE_PROFILE_ID: group.resourceProfileID,
      CARTULARY_BROWSER_SERVICE_REQUIREMENT: group.serviceRequirement,
      CARTULARY_HARNESS_SERVICE_DEPENDENCIES: group.serviceDependencies.join(","),
      CARTULARY_BROWSER_SESSION_CONTRACT: group.browserSessionGroup,
      CARTULARY_BROWSER_SESSION_GROUP: key,
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
    needs: fixture ? [fixture.builderUnitID] : [],
    resource_claims: resourceClaims(root, group),
    shared_locks: locks.shared,
    exclusive_locks: locks.exclusive,
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
    timeout_ms: owner.default_timeout_ms,
    evidence_outputs: [],
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms: 1000,
  };
}

function resetUnit(root, stage, group, previousID, owner, fixture) {
  const key = safeID(lifecycleKey(stage, group));
  const locks = hostLocks(group);
  return {
    unit_id: `browser_reset:${safeID(stage.name)}:${safeID(group.resetBefore)}`,
    owner_id: "harness.browser",
    kind: "lifecycle",
    command: command("tools/harness/browser/reset-web-e2e-stack.sh", ["--label", group.resetBefore], {
      CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
      CARTULARY_BROWSER_RESOURCE_PROFILE_ID: group.resourceProfileID,
      CARTULARY_BROWSER_SERVICE_REQUIREMENT: group.serviceRequirement,
      CARTULARY_HARNESS_SERVICE_DEPENDENCIES: group.serviceDependencies.join(","),
      CARTULARY_BROWSER_SESSION_CONTRACT: group.browserSessionGroup,
      CARTULARY_BROWSER_SESSION_GROUP: key,
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
    needs: [previousID],
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
    timeout_ms: owner.default_timeout_ms,
    evidence_outputs: [],
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

function groupUnit(root, stage, group, dependencyID, owner, mode, fixture) {
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
        CARTULARY_BROWSER_SERVICE_REQUIREMENT: group.serviceRequirement,
        CARTULARY_HARNESS_SERVICE_DEPENDENCIES: group.serviceDependencies.join(","),
        CARTULARY_BROWSER_SESSION_CONTRACT: group.browserSessionGroup,
        CARTULARY_BROWSER_SESSION_GROUP: key,
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
    needs: [dependencyID],
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
    evidence_outputs: [
      ...group.selectedRowIDs.map((rowID) => `rows/${rowID}.json`),
      `${target}/browser-groups/${safeID(group.name)}/browser-group-result.json`,
    ].sort(compareASCII),
    failure_policy: finalizableFailurePolicy(),
    estimated_work_ms:
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
    evidence_outputs: [
      `${target}/browser-groups/${safeID(group.name)}/frontend-measurement-summary.${fixture.profile.artifact_policy.summary_schema_id.split(".").at(-1)}.json`,
    ],
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms: 500,
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
        "--groups",
        stage.groups.map((group) => group.name).sort(compareASCII).join(","),
        "--group-targets",
        groupTargets.join(","),
        "--children",
        target === stage.target ? stage.summaryChildren.join(",") : "",
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
    evidence_outputs: [
      `${target}/browser-target-result.json`,
      ...(target === "browser-e2e-measurement" && includesFixtureMeasurements
        ? [`${target}/frontend-measurement-aggregate.json`]
        : []),
      ...(target === "browser-e2e-visual"
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
  const lifecycleIDs = new Map();
  const previousByLifecycle = new Map();
  const snapshotBuilderIDs = new Set();
  const groupUnits = [];
  const measurementSummaryUnits = [];
  for (const group of stage.groups) {
    const fixture = resolvedFixtureProfile(root, group);
    if (fixture && !snapshotBuilderIDs.has(fixture.builderUnitID)) {
      units.push(snapshotBuilderUnit(root, group, fixture, owner));
      snapshotBuilderIDs.add(fixture.builderUnitID);
    }
    const key = safeID(lifecycleKey(stage, group));
    let lifecycleID = lifecycleIDs.get(key);
    if (!lifecycleID) {
      const lifecycle = lifecycleUnit(root, stage, group, owner, fixture);
      lifecycleID = lifecycle.unit_id;
      lifecycleIDs.set(key, lifecycleID);
      previousByLifecycle.set(key, lifecycleID);
      units.push(lifecycle);
    }
    let dependencyID = previousByLifecycle.get(key);
    if (group.resetBefore) {
      const reset = resetUnit(root, stage, group, dependencyID, owner, fixture);
      units.push(reset);
      dependencyID = reset.unit_id;
    }
    const runner = groupUnit(root, stage, group, dependencyID, owner, mode, fixture);
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
      [
        ...groupUnits.map((unit) => unit.unit_id),
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
