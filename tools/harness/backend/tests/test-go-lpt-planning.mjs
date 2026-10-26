import assert from "node:assert/strict";
import test from "node:test";
import { planGoLPTShards } from "../go-lpt-shards.mjs";

test("go_lpt_cpu_budget: Go LPT planning preserves row closure and CPU budgets", async () => {
  const plan = planGoLPTShards(Array.from({ length: 17 }, (_, itemIndex) => ({ id: `row-${itemIndex}`, estimated_work_ms: itemIndex + 1, compatibility: { runtime: "none" } })), { availableGoLanes: 4 });
  assert.equal(plan.shards.flatMap((shard) => shard.item_ids).length, 17);
  assert.equal(plan.shards.length, 1);
  assert.ok(plan.shards.every((shard) => shard.cpu_tokens === 1));
  const hostPlan = planGoLPTShards(Array.from({ length: 10 }, (_, itemIndex) => ({
    id: `group-${itemIndex}`,
    estimated_work_ms: itemIndex + 1,
    compatibility: { package: `package-${itemIndex}` },
  })), { availableGoLanes: 24 });
  assert.equal(hostPlan.shards.length, 10);
  assert.ok(hostPlan.shards.every((shard) => shard.cpu_tokens === 1));
  const isolatedPlan = planGoLPTShards(Array.from({ length: 3 }, (_, itemIndex) => ({
    id: `isolated-${itemIndex}`,
    estimated_work_ms: itemIndex + 1,
    compatibility: { package: "same-package" },
    isolated: true,
  })), { availableGoLanes: 4 });
  assert.equal(isolatedPlan.shards.length, 3);
  assert.ok(isolatedPlan.shards.every((shard) => shard.isolated));
  const boundedItems = [80, 80, 80, 80, 80, 80, 200].map((cost, index) => ({
    id: `bounded-${index}`, estimated_work_ms: cost,
    compatibility: { package: "same-package", timeout_ms: 100, resource_claims: { cpu: 2 } },
  }));
  const options = { availableGoLanes: 2, invocationPolicy: { startup_ms: 10, cleanup_ms: 5 } };
  const bounded = planGoLPTShards(boundedItems, options);
  assert.deepEqual(planGoLPTShards([...boundedItems].reverse(), options), bounded);
  assert.deepEqual(bounded.shards.flatMap((shard) => shard.item_ids).sort(), boundedItems.map((item) => item.id));
  assert.equal(bounded.shards.length, 7, "timeout packing has no speculative two-wave limit");
  assert.ok(bounded.shards.every((shard) => shard.cpu_tokens === 2));
  assert.equal(bounded.shards.filter((shard) => shard.estimated_work_ms > 100).length, 1, "an indivisible oversized row stays explicit without changing its timeout or selection");
});
