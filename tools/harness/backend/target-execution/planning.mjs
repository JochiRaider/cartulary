import {
  aggregatePackages,
  aggregateRegex,
  fixturePolicyAssignments,
  resetTableAssignments,
} from "../go-target-aggregate.mjs";
import { collectGoShardPlanFromRows } from "../go-shard-plan.mjs";
import {
  collectTargetPlanRows,
  findTargetDescriptor,
} from "../backend-target-plan.mjs";
import { renderGoTestCommand } from "./command.mjs";
import { compareStrings } from "./util.mjs";

function targetRows(ctx) {
  if (!ctx.targetPlanRows) {
    const allRows = collectTargetPlanRows(ctx.repoRoot);
    if (ctx.scheduledScope !== "") {
      const supportedScopes = new Set(["all", "default_check", "rows"]);
      if (!supportedScopes.has(ctx.scheduledScope)) {
        throw new Error(`unknown scheduled Go selection scope ${ctx.scheduledScope}`);
      }
      if (ctx.scheduledScope !== "rows" && ctx.scheduledRowIDs.length > 0) {
        throw new Error(
          `scheduled Go selection scope ${ctx.scheduledScope} must not declare row IDs`,
        );
      }
      if (ctx.scheduledScope === "rows" && ctx.scheduledRowIDs.length === 0) {
        throw new Error("scheduled Go row selection scope requires row IDs");
      }
      if (ctx.scheduledScope === "all") {
        ctx.targetPlanRows = allRows;
        return ctx.targetPlanRows;
      }
      if (ctx.scheduledScope === "default_check") {
        ctx.targetPlanRows = allRows.filter(
          (row) => row.default_check_required === true,
        );
        return ctx.targetPlanRows;
      }
      const scheduledIDs = new Set(ctx.scheduledRowIDs);
      if (scheduledIDs.size !== ctx.scheduledRowIDs.length) {
        throw new Error("scheduled Go row selection contains duplicate row IDs");
      }
      const sortedScheduledIDs = [...ctx.scheduledRowIDs].sort(compareStrings);
      if (
        sortedScheduledIDs.some(
          (rowID, index) => rowID !== ctx.scheduledRowIDs[index],
        )
      ) {
        throw new Error("scheduled Go row selection must be sorted");
      }
      const knownIDs = new Set(allRows.map((row) => row.id));
      const missingIDs = [...scheduledIDs].filter((rowID) => !knownIDs.has(rowID));
      if (missingIDs.length > 0) {
        throw new Error(`scheduled Go row selection contains unknown row ${missingIDs.sort(compareStrings)[0]}`);
      }
      ctx.targetPlanRows = allRows.filter((row) => scheduledIDs.has(row.id));
      return ctx.targetPlanRows;
    }
    if (ctx.scheduledRowIDs.length > 0) {
      throw new Error("scheduled Go row IDs require a selection scope");
    }
    const ownerRows = allRows.filter((row) => {
      if (!ctx.ownerSelection) {
        return true;
      }
      return row.owner_id === ctx.ownerSelection;
    });
    if (ctx.selectedRowIDs.length === 0) {
      ctx.targetPlanRows = ownerRows;
    } else {
      const selectedIDs = new Set(ctx.selectedRowIDs);
      const selectedTargets = new Set(
        ownerRows
          .filter((row) => selectedIDs.has(row.id))
          .map((row) => row.target),
      );
      ctx.targetPlanRows = ownerRows.filter(
        (row) =>
          selectedIDs.has(row.id) ||
          (row.support_only === true && selectedTargets.has(row.target)),
      );
    }
  }
  return ctx.targetPlanRows;
}

function shardPlan(ctx) {
  if (!ctx.shardPlan) {
    ctx.shardPlan = collectGoShardPlanFromRows(ctx.repoRoot, targetRows(ctx), {
      owner: ctx.ownerSelection,
    });
  }
  return ctx.shardPlan;
}

export function rowsForAggregate(ctx, target, family) {
  const rows = targetRows(ctx).filter(
    (row) => row.target === target && row.execution_family === family,
  );
  if (rows.length === 0) {
    throw new Error(`unknown execution family ${family} for ${target}`);
  }
  return rows;
}

function rowPackages(row) {
  if (row.package) {
    return [row.package];
  }
  return [...(row.packages ?? [])];
}

function rowMatchesShardItems(row, items) {
  const rowIDs = new Set(
    items.map((item) => String(item.id ?? "").split(":")[0]),
  );
  if (rowIDs.has(row.id)) {
    return true;
  }
  if (row.coverage === "raw") {
    const packages = new Set(items.flatMap((item) => item.packages ?? []));
    return rowPackages(row).some((pkg) => packages.has(pkg));
  }
  const symbols = new Set(
    items
      .map((item) => item.symbol ?? "")
      .filter((symbol) => symbol !== ""),
  );
  return (row.symbols ?? []).some((symbol) => symbols.has(symbol));
}

export function rowsForScheduledAggregate(ctx, target, aggregateName, shardNames) {
  const rows = rowsForAggregate(ctx, target, aggregateName);
  if (shardNames.length === 0) {
    return rows;
  }
  const selectedItems = shardNames.flatMap(
    (shardName) => findShard(ctx, target, shardName).items ?? [],
  );
  const filtered = rows.filter((row) => rowMatchesShardItems(row, selectedItems));
  if (filtered.length === 0) {
    throw new Error(
      `scheduled shards for ${target}:${aggregateName} matched zero manifest rows`,
    );
  }
  return filtered;
}

export function aggregateNames(ctx, target) {
  return Array.from(
    new Set(
      targetRows(ctx)
        .filter((row) => row.target === target)
        .map((row) => row.execution_family),
    ),
  ).sort(compareStrings);
}

function targetOwnsShard(target, shard) {
  if (target === "backend-integration") {
    return shard.items.some(
      (item) => item.kind === "authoritative" || item.kind === "raw",
    );
  }
  if (target === "backend-integration-support") {
    return shard.items.some((item) => item.kind === "support");
  }
  if (target === "backend-store") {
    return shard.items.some((item) => item.kind === "authoritative");
  }
  if (target === "backend-process") {
    return shard.items.some((item) => item.target === "backend-process");
  }
  return false;
}

export function targetShards(ctx, target) {
  const plan = shardPlan(ctx);
  const aggregateSet = new Set(
    plan.aggregates
      .filter((aggregate) => aggregate.target === target)
      .map((aggregate) => aggregate.name),
  );
  return plan.shards.filter(
    (shard) =>
      aggregateSet.has(shard.aggregate_name) && targetOwnsShard(target, shard),
  );
}

export function targetAggregates(ctx, target) {
  return shardPlan(ctx).aggregates.filter(
    (aggregate) => aggregate.target === target,
  );
}

function findShard(ctx, target, name) {
  const shard = findShardOrNull(ctx, target, name);
  if (!shard) {
    throw new Error(`unknown shard ${name} for ${target}`);
  }
  return shard;
}

function findShardOrNull(ctx, target, name) {
  return (
    targetShards(ctx, target).find(
      (candidate) => candidate.name === name,
    ) ?? null
  );
}

function fixturePolicyAssignmentsForShard(shard, mode) {
  const assignments = [];
  for (const item of shard.items) {
    if (
      !item.postgres_fixture_policy ||
      item.postgres_fixture_policy === "migration_scratch"
    ) {
      continue;
    }
    if (mode === "tests" && item.symbol) {
      assignments.push(`${item.symbol}=${item.postgres_fixture_policy}`);
    }
    if (mode === "packages" && item.kind === "raw") {
      for (const pkg of item.packages) {
        assignments.push(`${pkg}=${item.postgres_fixture_policy}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

function resetTableAssignmentsForShard(shard, mode) {
  const assignments = [];
  for (const item of shard.items) {
    const dirtyTables = item.postgres_fixture_budget?.dirty_tables ?? [];
    if (dirtyTables.length === 0) {
      continue;
    }
    if (mode === "tests" && item.symbol) {
      assignments.push(`${item.symbol}=${dirtyTables.join("|")}`);
    }
    if (mode === "packages" && item.kind === "raw") {
      for (const pkg of item.packages) {
        assignments.push(`${pkg}=${dirtyTables.join("|")}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

function targetGoTestArgs(ctx, target) {
  const descriptor = findTargetDescriptor(target, ctx.repoRoot);
  if (!descriptor) {
    throw new Error(`unknown target ${target}`);
  }
  switch (descriptor.goTestParallelism) {
    case "none":
      return [];
    case "package":
      return ["-p", String(ctx.goTestPackageParallelism)];
    case "process":
      return ["-parallel", "4"];
    default:
      throw new Error(
        `unsupported go_test_parallelism ${descriptor.goTestParallelism} for ${target}`,
      );
  }
}

function aggregateSpec(ctx, target, family) {
  const rows = rowsForAggregate(ctx, target, family);
  return {
    regex: aggregateRegex(rows),
    args: [...targetGoTestArgs(ctx, target), ...aggregatePackages(rows)],
  };
}

function shardSpec(ctx, target, name) {
  const shard = findShard(ctx, target, name);
  return {
    regex: shard.regex,
    args: [...targetGoTestArgs(ctx, target), ...shard.packages],
  };
}

export function resolveExecutionFamilySpec(ctx, target, familyOrShard) {
  if (findShardOrNull(ctx, target, familyOrShard)) {
    return shardSpec(ctx, target, familyOrShard);
  }
  return aggregateSpec(ctx, target, familyOrShard);
}

export function resolveExecutionFamilyPostgresFixturePolicy(
  ctx,
  target,
  familyOrShard,
) {
  const shard = findShardOrNull(ctx, target, familyOrShard);
  if (shard) {
    return {
      tests: fixturePolicyAssignmentsForShard(shard, "tests").join(","),
      packages: fixturePolicyAssignmentsForShard(shard, "packages").join(","),
      resetTests: resetTableAssignmentsForShard(shard, "tests").join(","),
      resetPackages: resetTableAssignmentsForShard(shard, "packages").join(","),
    };
  }
  const rows = rowsForAggregate(ctx, target, familyOrShard);
  return {
    tests: fixturePolicyAssignments(rows, "tests").join(","),
    packages: fixturePolicyAssignments(rows, "packages").join(","),
    resetTests: resetTableAssignments(rows, "tests").join(","),
    resetPackages: resetTableAssignments(rows, "packages").join(","),
  };
}

export function runtimeRowsForExecution(ctx, target, familyOrShard) {
  const shard = findShardOrNull(ctx, target, familyOrShard);
  if (!shard) {
    return rowsForAggregate(ctx, target, familyOrShard);
  }
  const rowIDs = new Set(
    (shard.items ?? []).map((item) => String(item.id ?? "").split(":")[0]),
  );
  return targetRows(ctx).filter((row) => row.target === target && rowIDs.has(row.id));
}

export function isCrossTargetSharedReport(ctx, target, sharedName) {
  return findShardOrNull(ctx, target, sharedName)?.shared_across_targets === true;
}

export function inspectAggregateCommand(ctx, target, familyOrShard) {
  const spec = resolveExecutionFamilySpec(ctx, target, familyOrShard);
  const policy = resolveExecutionFamilyPostgresFixturePolicy(
    ctx,
    target,
    familyOrShard,
  );
  return renderGoTestCommand(ctx, spec.regex, spec.args, policy);
}
