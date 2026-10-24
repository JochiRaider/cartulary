import { packInvocationCohorts } from "../scheduler/work-graph/cohorts.mjs";
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

export function planGoLPTShards(items, { availableGoLanes,
  invocationPolicy = { startup_ms: 5000, cleanup_ms: 250 } } = {}) {
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
  const ordered = [...groups.values()].sort(
    (left, right) =>
      compareASCII(left.digest, right.digest) ||
      compareASCII(left.items[0].id, right.items[0].id),
  );
  const shards = [];
  for (const group of ordered) {
    const contract = group.items[0].compatibility;
    const cpu = Math.min(availableGoLanes, contract.execution_context?.cpu_tokens ?? contract.resource_claims?.cpu ?? 1);
    const cohorts = packInvocationCohorts(
      group.items.map((item) => ({ ...item, cost: item.estimated_work_ms })),
      invocationPolicy, contract.timeout_ms ?? 600000,
    );
    for (const cohort of cohorts) {
      const itemIDs = cohort.items.map((item) => item.id).sort(compareASCII);
      const itemDigest = semanticJSONDigest({ itemIDs, invocationPolicy, packing: "bounded_first_fit" }).slice("sha256:".length, "sha256:".length + 12);
      shards.push({
        shard_id: `go:${group.digest.slice("sha256:".length, "sha256:".length + 16)}:${itemDigest}`,
        compatibility_digest: group.digest,
        item_ids: itemIDs,
        estimated_work_ms: cohort.items.reduce((total, item) => total + item.cost, 0),
        isolated: group.isolated,
        cpu_tokens: cpu,
      });
    }
  }
  shards.sort((left, right) => compareASCII(left.shard_id, right.shard_id));
  const plan = {
    schema_id: "cartulary.harness_go_lpt_plan.v3",
    available_parallelism: availableGoLanes,
    shards,
    plan_digest: "",
  };
  const semantic = { ...plan };
  delete semantic.plan_digest;
  plan.plan_digest = semanticJSONDigest(semantic);
  validateSchemaSync(plan.schema_id, plan);
  return plan;
}
