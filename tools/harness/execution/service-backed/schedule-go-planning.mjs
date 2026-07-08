import { collectGoShardsForTarget } from "../../backend/backend-shard-plan.mjs";
import {
  runtimeBinaryDefaultEnvForIDs,
  runtimeBinaryProducerTargetsForIDs,
  runtimeBinaryRegistry,
} from "../../runtime-binary-registry.mjs";
import { readinessAttributionForMakeTarget } from "../../scheduler/scheduler-manifest.mjs";
import { directRuntimeProducerClaims } from "./schedule-resource-claims.mjs";
import { command, sortedUnique } from "./schedule-utils.mjs";

export function shardCompletionKey(shardName) {
  return `go_shard:${shardName}`;
}

function runtimeBinaryIDsForShard(shard) {
  return sortedUnique((shard.items ?? []).flatMap((item) => item.runtime_binaries ?? []));
}

function runtimeBinaryEnvForIDs(source, ids) {
  return runtimeBinaryDefaultEnvForIDs(
    runtimeBinaryRegistry(source.runtime_binary_records ?? []),
    ids,
    `${source.target} shard`,
  );
}

function runtimeBinaryNeedsForIDs(source, ids) {
  return runtimeBinaryProducerTargetsForIDs(
    runtimeBinaryRegistry(source.runtime_binary_records ?? []),
    ids,
    `${source.target} shard`,
  );
}

export function shardRuntimeConfig(source, shard) {
  const runtimeBinaries = runtimeBinaryIDsForShard(shard);
  return {
    runtimeBinaries,
    needs: runtimeBinaryNeedsForIDs(source, runtimeBinaries),
    env: runtimeBinaryEnvForIDs(source, runtimeBinaries),
  };
}

export function collectServiceBackedGoShards(repoRoot, source, scheduleTarget) {
  const shards = collectGoShardsForTarget(repoRoot, source.target, {
    defaultCheckOnly: source.default_check_required === true,
  });
  if (shards.length === 0) {
    throw new Error(`${scheduleTarget} go_shards source ${source.target} selected no shards`);
  }
  return shards;
}

export function addDirectRuntimeProducerUnits(unitsByTarget, runtime, source, sourceIndex, priority) {
  for (const target of runtime.needs) {
    if (unitsByTarget.has(target)) {
      continue;
    }
    const readinessAttribution = readinessAttributionForMakeTarget(target);
    unitsByTarget.set(target, {
      id: target,
      kind: "make_target",
      class: "backend",
      target,
      label: target,
      aggregate_target: target,
      priority: priority(source.priority),
      weight_ms: 1,
      needs: [],
      completion_keys: [target],
      failure_keys: [target],
      make_prerequisite_policy: "run",
      resource_claims: directRuntimeProducerClaims(),
      command: command("make_target", { target }),
      ...(readinessAttribution ? { readiness_attribution: readinessAttribution } : {}),
      order: sourceIndex - 0.5,
    });
  }
}
