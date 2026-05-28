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

function browserGroupID(group) {
  if (typeof group?.id !== "string" || group.id.trim() === "") {
    throw new Error("scheduled browser group must declare id before worker-slot allocation");
  }
  return group.id.trim();
}

function positiveIntegerText(value, label) {
  const text = String(value).trim();
  const parsed = Number.parseInt(text, 10);
  if (!Number.isInteger(parsed) || parsed < 1 || String(parsed) !== text) {
    throw new Error(`${label} must be a positive integer`);
  }
  return parsed;
}

export function browserGroupWorkerSlotCount(group) {
  if (group?.kind === "functional_shard" || group?.kind === "support") {
    return 1;
  }
  const workers = group?.workers ?? group?.worker_count ?? group?.workerCount;
  if (workers === undefined || workers === null || String(workers).trim() === "") {
    return 1;
  }
  if (String(workers).trim() === "default") {
    return 1;
  }
  return positiveIntegerText(workers, `scheduled browser group ${browserGroupID(group)} workers`);
}

function browserGroupSlotSortKey(entry) {
  const kind = entry.group?.kind ?? "";
  if (kind === "functional_shard") {
    return [
      0,
      integerField(entry.group, "shardIndex", "shard_index"),
      entry.sourceIndex,
      entry.groupIndex,
    ];
  }
  if (kind === "support") {
    return [1, entry.sourceIndex, entry.groupIndex, 0];
  }
  return [2, entry.sourceIndex, entry.groupIndex, 0];
}

function compareSlotEntries(left, right) {
  const leftKey = browserGroupSlotSortKey(left);
  const rightKey = browserGroupSlotSortKey(right);
  for (let index = 0; index < Math.max(leftKey.length, rightKey.length); index += 1) {
    const comparison = (leftKey[index] ?? 0) - (rightKey[index] ?? 0);
    if (comparison !== 0) {
      return comparison;
    }
  }
  return browserGroupID(left.group).localeCompare(browserGroupID(right.group));
}

export function browserGroupWorkerSlotPlan(sources) {
  const entries = [];
  const seen = new Set();
  for (const [sourceIndex, source] of (sources ?? []).entries()) {
    if (source?.type !== "browser_stage") {
      continue;
    }
    for (const [groupIndex, group] of (source.groups ?? []).entries()) {
      const id = browserGroupID(group);
      if (seen.has(id)) {
        throw new Error(`scheduled browser group ${id} appears more than once`);
      }
      seen.add(id);
      entries.push({
        group,
        groupIndex,
        id,
        slots: browserGroupWorkerSlotCount(group),
        sourceIndex,
      });
    }
  }

  const ordered = [...entries].sort(compareSlotEntries);
  const totalWorkerCount = ordered.reduce((sum, entry) => sum + entry.slots, 0);
  const envByGroupID = new Map();
  let offset = 0;
  for (const entry of ordered) {
    envByGroupID.set(entry.id, {
      CARTULARY_PLAYWRIGHT_WORKER_COUNT: String(totalWorkerCount),
      CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET: String(offset),
      ...(entry.group?.kind === "support" ? { PLAYWRIGHT_WORKERS: "1" } : {}),
    });
    offset += entry.slots;
  }
  return envByGroupID;
}

export function browserGroupWorkerEnv(groups, group) {
  return browserGroupWorkerSlotPlan([{ type: "browser_stage", groups }]).get(browserGroupID(group)) ?? {};
}

export function browserGroupWorkerEnvFromPlan(plan, group) {
  const id = browserGroupID(group);
  const env = plan.get(id);
  if (!env) {
    throw new Error(`scheduled browser group ${id} has no worker-slot allocation`);
  }
  return { ...env };
}
