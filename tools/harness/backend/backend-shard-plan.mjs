import {
  collectGoShardPlanFromRows,
  collectGoShardsForTargetFromRows,
} from "./go-shard-plan.mjs";
import { collectTargetPlanRows } from "./target-plan.mjs";

export { collectGoShardPlanFromRows, collectGoShardsForTargetFromRows };

export function targetPlanRowsForGoShards(root = process.cwd()) {
  return collectTargetPlanRows(root);
}

export function collectGoShardPlan(root = process.cwd(), options = {}) {
  return collectGoShardPlanFromRows(root, targetPlanRowsForGoShards(root), options);
}

export function collectGoShardsForTarget(root = process.cwd(), target, options = {}) {
  return collectGoShardsForTargetFromRows(
    root,
    targetPlanRowsForGoShards(root),
    target,
    options,
  );
}
