export function browserStageSessionKey(target) {
  return `browser_stage_session:${target}`;
}

export function browserGroupCompletionKey(groupID) {
  return `browser_group:${groupID}`;
}

export function browserGroupNeeds(stageSessionKey) {
  return [stageSessionKey];
}

export function browserStageCompletionNeeds(groups) {
  return (groups ?? []).map((group) => browserGroupCompletionKey(group.id));
}

function integerField(value, ...fields) {
  for (const field of fields) {
    if (Number.isInteger(value?.[field])) {
      return value[field];
    }
  }
  return 0;
}

export function browserGroupWorkerEnv(groups, group) {
  const hasSupportGroup = (groups ?? []).some((candidate) => candidate?.kind === "support");
  const functionalShardCount = Math.max(
    0,
    ...(groups ?? [])
      .filter((candidate) => candidate?.kind === "functional_shard")
      .map((candidate) => integerField(candidate, "shardCount", "shard_count")),
  );
  if (group?.kind === "functional_shard") {
    return {
      CARTULARY_PLAYWRIGHT_WORKER_COUNT: String(
        (functionalShardCount || integerField(group, "shardCount", "shard_count")) +
          (hasSupportGroup ? 1 : 0),
      ),
      CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET: String(
        integerField(group, "shardIndex", "shard_index"),
      ),
    };
  }
  if (group?.kind === "support") {
    return {
      CARTULARY_PLAYWRIGHT_WORKER_COUNT: String(functionalShardCount + 1),
      CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET: String(functionalShardCount),
      PLAYWRIGHT_WORKERS: "1",
    };
  }
  return {};
}
