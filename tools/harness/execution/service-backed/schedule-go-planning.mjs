import { collectGoShardsForTarget } from "../../backend/backend-shard-plan.mjs";
import { directRuntimeProducerClaims } from "./schedule-resource-claims.mjs";
import { command, sortedUnique } from "./schedule-utils.mjs";

export function shardCompletionKey(shardName) {
  return `go_shard:${shardName}`;
}

function runtimeBinaryIDsForShard(shard) {
  return sortedUnique((shard.items ?? []).flatMap((item) => item.runtime_binaries ?? []));
}

function runtimeBinaryRegistry(source) {
  return new Map((source.runtime_binary_records ?? []).map((entry) => [entry.id, entry]));
}

function runtimeBinaryEnvForIDs(source, ids) {
  const registry = runtimeBinaryRegistry(source);
  const env = {};
  for (const id of ids) {
    const entry = registry.get(id);
    if (!entry) {
      throw new Error(`${source.target} shard runtime binary ${id} is missing from runtime_binary_records`);
    }
    if (id !== "operator") {
      throw new Error(`${source.target} shard runtime binary ${id} is missing default output path wiring`);
    }
    env[entry.consumer_env] = "operator";
  }
  return env;
}

function runtimeBinaryNeedsForIDs(source, ids) {
  const registry = runtimeBinaryRegistry(source);
  return ids.map((id) => {
    const entry = registry.get(id);
    if (!entry) {
      throw new Error(`${source.target} shard runtime binary ${id} is missing from runtime_binary_records`);
    }
    return entry.producer_target;
  });
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
      order: sourceIndex - 0.5,
    });
  }
}
