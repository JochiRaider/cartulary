import { validateSchemaSync } from "../contract/index.mjs";
import { semanticJSONDigest } from "../test-catalog/index.mjs";

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
  const targetShardCount = availableGoLanes * 2;
  const groups = new Map();
  for (const item of items) {
    const digest = semanticJSONDigest(item.compatibility);
    const key = `${digest}:${item.isolated === true ? item.id : "shared"}`;
    const group = groups.get(key) ?? { digest, isolated: item.isolated === true, items: [] };
    group.items.push(item);
    groups.set(key, group);
  }
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
    const binCount = group.isolated ? 1 : Math.min(sorted.length, targetShardCount);
    const bins = Array.from({ length: binCount }, () => ({ items: [], weight: 0 }));
    for (const item of sorted) {
      const bin = bins
        .map((entry, index) => ({ entry, index }))
        .sort(
          (left, right) => left.entry.weight - right.entry.weight || left.index - right.index,
        )[0].entry;
      bin.items.push(item);
      bin.weight += item.estimated_work_ms;
    }
    bins.forEach((bin, index) => {
      const itemIDs = bin.items.map((item) => item.id).sort(compareASCII);
      const itemDigest = semanticJSONDigest(itemIDs).slice("sha256:".length, "sha256:".length + 12);
      shards.push({
        shard_id: `go:${group.digest.slice("sha256:".length, "sha256:".length + 16)}:${itemDigest}:${String(index + 1).padStart(3, "0")}`,
        compatibility_digest: group.digest,
        item_ids: itemIDs,
        estimated_work_ms: bin.weight,
        isolated: group.isolated,
      });
    });
  }
  shards.sort((left, right) => compareASCII(left.shard_id, right.shard_id));
  const plan = {
    schema_id: "cartulary.harness_go_lpt_plan.v1",
    available_go_lanes: availableGoLanes,
    target_shard_count: targetShardCount,
    shards,
    plan_digest: "",
  };
  const semantic = { ...plan };
  delete semantic.plan_digest;
  plan.plan_digest = semanticJSONDigest(semantic);
  validateSchemaSync(plan.schema_id, plan);
  return plan;
}
