#!/usr/bin/env node
import { collectGoShardPlan } from "./go-shard-plan.mjs";

process.stdout.on("error", (error) => {
  if (error.code === "EPIPE") {
    process.exit(0);
  }
  throw error;
});

function usage() {
  process.stderr.write("usage: print-go-shard-plan.mjs [--json] [--target <target>]\n");
  process.exit(2);
}

function parseArgs(argv) {
  const options = { json: false, target: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--json") {
      options.json = true;
      continue;
    }
    if (arg === "--target") {
      const target = argv[index + 1];
      if (!target) {
        usage();
      }
      options.target = target;
      index += 1;
      continue;
    }
    usage();
  }
  return options;
}

function filterPlan(plan, target) {
  if (!target) {
    return plan;
  }
  return {
    ...plan,
    targets: plan.targets.filter((candidate) => candidate === target),
    aggregates: plan.aggregates.filter((aggregate) => aggregate.target === target),
    shards: plan.shards.filter((shard) => shard.target === target),
  };
}

function renderHuman(plan) {
  const lines = [];
  for (const target of plan.targets) {
    const shards = plan.shards.filter((shard) => shard.target === target);
    lines.push(`${target} shards=${shards.length}`);
    for (const shard of shards) {
      lines.push(
        `  - ${shard.name} weight_ms=${shard.weight_ms} shard_target_ms=${shard.shard_target_ms} scheduler_profile=${shard.scheduler_profile} aggregate=${shard.aggregate_name} items=${shard.item_count}`,
      );
    }
  }
  return lines.join("\n");
}

try {
  const options = parseArgs(process.argv.slice(2));
  const plan = filterPlan(collectGoShardPlan(process.cwd()), options.target);
  if (options.json) {
    process.stdout.write(`${JSON.stringify(plan, null, 2)}\n`);
  } else {
    process.stdout.write(`${renderHuman(plan)}\n`);
  }
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
