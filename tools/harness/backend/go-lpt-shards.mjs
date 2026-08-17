import { validateSchemaSync } from "../contract/index.mjs";
import { semanticJSONDigest } from "../contract/index.mjs";

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function validateItems(items) {
  if (!Array.isArray(items) || items.length === 0) {
    throw new Error("Go LPT planning requires at least one item");
  }
  const ids = items.map((item) => item.id);
  if (ids.some((id) => typeof id !== "string" || id.length === 0)) {
    throw new Error("Go LPT items require stable IDs");
  }
  if (new Set(ids).size !== ids.length) throw new Error("Go LPT items contain duplicate IDs");
  for (const item of items) {
    if (!Number.isInteger(item.estimated_work_ms) || item.estimated_work_ms < 1) {
      throw new Error(`${item.id} has invalid estimated work`);
    }
    if (!item.compatibility || typeof item.compatibility !== "object" || Array.isArray(item.compatibility)) {
      throw new Error(`${item.id} has no compatibility declaration`);
    }
  }
}

export function planGoLPTShards(items, { availableGoLanes } = {}) {
  validateItems(items);
  if (!Number.isInteger(availableGoLanes) || availableGoLanes < 1) {
    throw new Error("availableGoLanes must be a positive integer");
  }
  const groups = new Map();
  for (const item of items) {
    const digest = semanticJSONDigest(item.compatibility);
    const key = `${digest}:${item.isolated === true ? item.id : "shared"}`;
    const group = groups.get(key) ?? { digest, isolated: item.isolated === true, items: [] };
    group.items.push(item);
    groups.set(key, group);
  }
  const workerCount = Math.min(
    groups.size,
    Math.min(8, Math.max(1, Math.floor(availableGoLanes / 4))),
  );
  const goMaxProcs = Math.max(1, Math.floor(availableGoLanes / workerCount));
  const shards = [];
  for (const group of [...groups.values()].sort(
    (left, right) =>
      compareASCII(left.digest, right.digest) ||
      compareASCII(left.items[0].id, right.items[0].id),
  )) {
    const sorted = [...group.items].sort(
      (left, right) =>
        right.estimated_work_ms - left.estimated_work_ms || compareASCII(left.id, right.id),
    );
    const itemIDs = sorted.map((item) => item.id).sort(compareASCII);
    const itemDigest = semanticJSONDigest(itemIDs).slice(
      "sha256:".length,
      "sha256:".length + 12,
    );
    shards.push({
      shard_id: `go:${group.digest.slice("sha256:".length, "sha256:".length + 16)}:${itemDigest}`,
      compatibility_digest: group.digest,
      item_ids: itemIDs,
      estimated_work_ms: sorted.reduce(
        (total, item) => total + item.estimated_work_ms,
        0,
      ),
      isolated: group.isolated,
      cpu_tokens: goMaxProcs,
    });
  }
  shards.sort((left, right) => compareASCII(left.shard_id, right.shard_id));
  const plan = {
    schema_id: "cartulary.harness_go_lpt_plan.v2",
    available_parallelism: availableGoLanes,
    worker_count: workerCount,
    gomaxprocs: goMaxProcs,
    shards,
    plan_digest: "",
  };
  const semantic = { ...plan };
  delete semantic.plan_digest;
  plan.plan_digest = semanticJSONDigest(semantic);
  validateSchemaSync(plan.schema_id, plan);
  return plan;
}
