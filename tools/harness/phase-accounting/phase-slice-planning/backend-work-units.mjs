import { collectGoShardsForTargetFromRows } from "../../backend/backend-shard-plan.mjs";
import {
  collectTargetPlanRows,
  findTargetDescriptor,
} from "../../backend/backend-target-plan.mjs";
import { goShardSchedulerProfileClaims } from "../../scheduler/scheduler-resource-policy.mjs";
import { runtimeBinariesForRows } from "./row-selection.mjs";
import {
  goCPUResource,
  goIOResource,
  mergeClaims,
  targetWeight,
  uniqueSorted,
} from "./work-unit-common.mjs";

export function goShardTargetPlanRows(phase, rows, root) {
  const selectedRowIDs = new Set(rows.map((row) => row.id));
  const selectedGoShardTargets = new Set(
    rows
      .filter(
        (row) =>
          row.runner === "go_test" &&
          findTargetDescriptor(row.target, root)?.sharding === "go_shards",
      )
      .map((row) => row.target),
  );
  if (selectedGoShardTargets.size === 0) {
    return [];
  }
  return collectTargetPlanRows(root).filter(
    (row) =>
      row.manifest_phase === phase &&
      selectedGoShardTargets.has(row.target) &&
      (selectedRowIDs.has(row.id) || row.support_only === true),
  );
}

function schedulerClaimsForShard(shard, resourceLimits) {
  return new Map(
    Object.entries(
      goShardSchedulerProfileClaims(shard.scheduler_profile, {
        scheduler: "phase_slice",
        resourceLimits,
        shardName: shard.name,
      }),
    ),
  );
}

function runtimeBinariesForShard(shard) {
  return uniqueSorted((shard.items ?? []).flatMap((item) => item.runtime_binaries ?? []));
}

function backendProcessClaimsForShard(target, _runtimeBinaries, resourceLimits) {
  if (target !== "backend-process") {
    return new Map();
  }
  if (!resourceLimits.has("process")) {
    throw new Error("backend-process Go shards require resource_limits.process");
  }
  return new Map([["process", 1]]);
}

function shardCompletionKey(shardName) {
  return `go_shard:${shardName}`;
}

export function addGoUnits(plan, target, rows) {
  const descriptor = findTargetDescriptor(target, plan.root);
  if (!descriptor) {
    throw new Error(`phase slice target ${target} is not in target-plan`);
  }

  if (descriptor.sharding !== "go_shards") {
    const runtimeBinaries = runtimeBinariesForRows(rows);
    plan.workUnits.push({
      id: target,
      label: target,
      kind: "go_target",
      type: "go_target",
      class: "backend",
      target,
      aggregateTarget: target,
      group: target,
      needs: [],
      completionKeys: [target],
      failureKeys: [target],
      weightMs: targetWeight(rows),
      resourceClaims: new Map([[goCPUResource, 1], [goIOResource, 1]]),
      runtime_binaries: runtimeBinaries,
      order: plan.nextOrder++,
    });
    return;
  }

  const shards = collectGoShardsForTargetFromRows(plan.root, plan.goShardRows, target, {
    phase: plan.phase,
  });
  if (shards.length === 0) {
    throw new Error(`phase slice ${plan.phase} selected no Go shards for ${target}`);
  }
  const sourceClaims = new Map([
    ["postgres", 1],
    ["object_store", 1],
  ]);
  plan.workUnits.push({
    id: `finalize:${target}`,
    label: `finalize/${target}`,
    kind: "finalizer",
    type: "finalizer",
    class: "backend",
    target,
    aggregateTarget: target,
    group: target,
    needs: shards.map((shard) => shardCompletionKey(shard.name)),
    completionKeys: [target],
    failureKeys: [target],
    countInTotal: false,
    countsStarted: false,
    resourceClaims: new Map(),
    shardNames: shards.map((shard) => shard.name),
    unblockLabel: target,
    weightMs: 0,
    order: plan.nextOrder++,
  });
  for (const shard of shards) {
    const runtimeBinaries = runtimeBinariesForShard(shard);
    plan.workUnits.push({
      id: `${target}:${shard.name}`,
      label: `${target}/${shard.name}`,
      kind: "go_shard",
      type: "go_shard",
      class: "backend",
      target,
      aggregateTarget: target,
      group: target,
      needs: [],
      completionKeys: [shardCompletionKey(shard.name)],
      failureKeys: [shardCompletionKey(shard.name)],
      runningDependencyKeys: [target],
      completeOnFailure: true,
      shard: shard.name,
      schedulerProfile: shard.scheduler_profile,
      weightMs: shard.weight_ms,
      resourceClaims: mergeClaims(
        sourceClaims,
        schedulerClaimsForShard(shard, plan.resourceLimits),
        backendProcessClaimsForShard(target, runtimeBinaries, plan.resourceLimits),
      ),
      runtime_binaries: runtimeBinaries,
      order: plan.nextOrder++,
    });
  }
}
