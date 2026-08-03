import path from "node:path";

import { loadBrowserBatchStages } from "../adapters/browser.mjs";
import { buildWorkGraph } from "./model.mjs";

const manifestRelativePath = "tools/browser_e2e_batch_manifest.json";
// One browser stack owns a server pool plus setup/reset clients. Four logical
// Postgres lanes prevent the scheduler from admitting enough simultaneous
// stacks to exhaust the shared service's connection limit.
const browserPostgresLanes = 4;

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

function command(executable, args, environment = {}) {
  return { executable, args, environment };
}

function lifecycleKey(stage, group) {
  if (group.kind === "stateful_partition") {
    return `${stage.name}-${group.browserSessionGroup}-${group.runtimeProfileID}`;
  }
  return `${stage.name}-${group.name}-${group.runtimeProfileID}`;
}

function lifecycleUnit(stage, group, owner) {
  const key = safeID(lifecycleKey(stage, group));
  return {
    unit_id: `browser_lifecycle:${key}`,
    owner_id: "harness.browser",
    kind: "readiness",
    command: command("true", [], {
      CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
      CARTULARY_BROWSER_SERVICE_REQUIREMENT: group.serviceRequirement,
      CARTULARY_BROWSER_SESSION_CONTRACT: group.browserSessionGroup,
      CARTULARY_BROWSER_SESSION_GROUP: key,
    }),
    needs: [],
    resource_claims: { browser_stack: 1, io: 1, memory_mb: 128, port_lane: 1, postgres: browserPostgresLanes, process: 1 },
    shared_locks: ["host_activity"],
    exclusive_locks: [],
    affinity_key: key,
    fixture_lease: "browser_stack",
    cache_policy: "none",
    timeout_ms: owner.default_timeout_ms,
    evidence_outputs: [],
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms: 1000,
  };
}

function resetUnit(stage, group, previousID, owner) {
  const key = safeID(lifecycleKey(stage, group));
  return {
    unit_id: `browser_reset:${safeID(stage.name)}:${safeID(group.resetBefore)}`,
    owner_id: "harness.browser",
    kind: "lifecycle",
    command: command("tools/harness/browser/reset-web-e2e-stack.sh", ["--label", group.resetBefore], {
      CARTULARY_BROWSER_RUNTIME_PROFILE_ID: group.runtimeProfileID,
      CARTULARY_BROWSER_SERVICE_REQUIREMENT: group.serviceRequirement,
      CARTULARY_BROWSER_SESSION_CONTRACT: group.browserSessionGroup,
      CARTULARY_BROWSER_SESSION_GROUP: key,
    }),
    needs: [previousID],
    resource_claims: { browser_stack: 1, io: 1, memory_mb: 64, postgres: browserPostgresLanes, process: 1 },
    shared_locks: ["host_activity"],
    exclusive_locks: [`browser_session:${key}`],
    affinity_key: key,
    fixture_lease: "browser_stack",
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
  if (group.kind === "measurement") exclusive.push("host_activity");
  else shared.push("host_activity");
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

function groupUnit(stage, group, dependencyID, owner, mode) {
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
        CARTULARY_BROWSER_SERVICE_REQUIREMENT: group.serviceRequirement,
        CARTULARY_BROWSER_SESSION_CONTRACT: group.browserSessionGroup,
        CARTULARY_BROWSER_SESSION_GROUP: key,
        CARTULARY_TEST_TARGET: target,
        ...(mode === "snapshot_update"
          ? {
              CARTULARY_BROWSER_MAINTENANCE_MODE: "snapshot_update",
              CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS: "1",
            }
          : {}),
      },
    ),
    needs: [dependencyID],
    resource_claims: { browser_stack: 1, cpu: 1, io: 1, memory_mb: 512, port_lane: 1, postgres: browserPostgresLanes, process: 1 },
    shared_locks: locks.shared,
    exclusive_locks: [...locks.exclusive, `browser_session:${key}`].sort(compareASCII),
    affinity_key: key,
    fixture_lease: "browser_stack",
    cache_policy: "none",
    timeout_ms: owner.default_timeout_ms,
    evidence_outputs: [
      ...group.selectedRowIDs.map((rowID) => `rows/${rowID}.json`),
      `${target}/browser-groups/${safeID(group.name)}/browser-group-result.json`,
    ].sort(compareASCII),
    failure_policy: requiredFailurePolicy(),
    estimated_work_ms:
      owner.evidence_estimates_ms[group.kind === "a11y" ? "accessibility" : group.kind] ??
      owner.evidence_estimates_ms.browser,
  };
}

function targetFinalizer(stage, target, groupUnits, needs, owner) {
  const groupTargets = stage.groups
    .map((group) => `${group.name}=${target === "browser-e2e-visual-update" ? target : group.target}`)
    .sort(compareASCII);
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
    resource_claims: { cpu: 1, io: 1, memory_mb: 128, process: 1 },
    shared_locks: ["host_activity"],
    exclusive_locks: [],
    fixture_lease: "none",
    cache_policy: "none",
    timeout_ms: owner.default_timeout_ms,
    evidence_outputs: [`${target}/browser-target-result.json`],
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
  const groupUnits = [];
  for (const group of stage.groups) {
    const key = safeID(lifecycleKey(stage, group));
    let lifecycleID = lifecycleIDs.get(key);
    if (!lifecycleID) {
      const lifecycle = lifecycleUnit(stage, group, owner);
      lifecycleID = lifecycle.unit_id;
      lifecycleIDs.set(key, lifecycleID);
      previousByLifecycle.set(key, lifecycleID);
      units.push(lifecycle);
    }
    let dependencyID = previousByLifecycle.get(key);
    if (group.resetBefore) {
      const reset = resetUnit(stage, group, dependencyID, owner);
      units.push(reset);
      dependencyID = reset.unit_id;
    }
    const runner = groupUnit(stage, group, dependencyID, owner, mode);
    units.push(runner);
    groupUnits.push(runner);
    previousByLifecycle.set(key, runner.unit_id);
  }

  for (const terminalUnitID of previousByLifecycle.values()) {
    const terminalUnit = units.find((unit) => unit.unit_id === terminalUnitID);
    terminalUnit.command.environment.CARTULARY_BROWSER_RELEASE_AFFINITY = "1";
  }

  const target = mode === "snapshot_update" ? "browser-e2e-visual-update" : stage.target;
  units.push(
    targetFinalizer(
      stage,
      target,
      groupUnits,
      groupUnits.map((unit) => unit.unit_id),
      owner,
    ),
  );
  return buildWorkGraph(units);
}
