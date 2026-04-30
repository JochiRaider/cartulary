import { readFileSync } from "node:fs";
import path from "node:path";
import {
  defaultShardTargetMsByTargetEntries,
  normalizePositiveInteger,
  rawAggregateBaselineKey,
  rawPackageBaselineKey,
  readGoDurationBaselineMaps,
  testBaselineKey,
} from "./go-duration-baselines.mjs";
import { collectTargetPlanRows } from "./target-plan.mjs";

const cpuHeavyShardWeightMs = 12_000;
const ioHeavyFixturePolicies = new Set(["group_clone", "migration_scratch"]);
const shardTargets = new Set(["backend-store", "backend-integration", "backend-integration-support"]);
const executionTargets = new Set(["backend-store", "backend-integration", "backend-integration-support"]);
const defaultShardTargetMsByTarget = new Map(defaultShardTargetMsByTargetEntries);

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function exactRegex(values) {
  const escaped = values.map((value) => value.replace(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`));
  if (escaped.length === 0) {
    throw new Error("cannot create shard regex from empty symbol list");
  }
  if (escaped.length === 1) {
    return `^${escaped[0]}$`;
  }
  return `^(${escaped.join("|")})$`;
}

function loadGoModulePath(root) {
  const goMod = readFileSync(path.join(root, "go.mod"), "utf8");
  const match = goMod.match(/^module\s+(\S+)$/m);
  if (!match) {
    throw new Error("unable to determine Go module path from go.mod");
  }
  return match[1];
}

function toGoImportPath(modulePath, repoRelativePackage) {
  if (!repoRelativePackage.startsWith("./")) {
    throw new Error(`Go package must be repo-relative: ${repoRelativePackage}`);
  }
  const suffix = repoRelativePackage.slice(2);
  return suffix === "" ? modulePath : `${modulePath}/${suffix}`;
}

function rowPackages(row) {
  if (row.package) {
    return [row.package];
  }
  return [...row.packages];
}

function rowPackageImportPaths(modulePath, row) {
  return rowPackages(row).map((pkg) => toGoImportPath(modulePath, pkg));
}

function rawItemWeight(baselines, key, aggregateKey, rawPackageCount) {
  const packageWeight = baselines.rawAggregates.get(key);
  if (normalizePositiveInteger(packageWeight, 0) > 0) {
    return {
      weightMs: packageWeight,
      weightSource: "baseline",
    };
  }
  const aggregateWeight = baselines.rawAggregates.get(aggregateKey);
  if (normalizePositiveInteger(aggregateWeight, 0) > 0 && rawPackageCount > 0) {
    return {
      weightMs: Math.max(1, Math.ceil(aggregateWeight / rawPackageCount)),
      weightSource: "aggregate_baseline",
    };
  }
  return {
    weightMs: baselines.defaultItemWeightMs,
    weightSource: "default",
  };
}

function unionRegex(regexes) {
  const values = Array.from(regexes).sort(compareStrings);
  if (values.length === 0) {
    throw new Error("cannot build raw shard regex from an empty selection");
  }
  if (values.length === 1) {
    return values[0];
  }
  return values.map((regex) => `(${regex})`).join("|");
}

function addAggregate(aggregates, row, mode) {
  const key = `${row.target}\u001f${row.execution_family}`;
  if (!aggregates.has(key)) {
    aggregates.set(key, {
      target: row.target,
      name: row.execution_family,
      mode,
      label: row.execution_label ?? row.label ?? row.execution_family,
      phase: row.manifest_phase,
      section: row.section,
      coverage: row.coverage,
      execution_dependency: row.execution_dependency,
      raw_selector: row.raw_selector ?? "",
      packages: new Set(),
      shards: new Set(),
      weight_ms: 0,
    });
  }
  const aggregate = aggregates.get(key);
  for (const pkg of row.packages ?? rowPackages(row)) {
    aggregate.packages.add(pkg);
  }
  return aggregate;
}

function normalizePostgresFixturePolicy(value) {
  return value === "template_clone" ||
    value === "package_reset" ||
    value === "migration_scratch" ||
    value === "transaction" ||
    value === "group_clone"
    ? value
    : "";
}

function buildExecutionItems(root) {
  const modulePath = loadGoModulePath(root);
  const baselines = readGoDurationBaselineMaps(root, "", { allowMissing: true });
  const rows = collectTargetPlanRows(root);
  const aggregates = new Map();
  const executableItems = [];

  for (const row of rows) {
    if (!executionTargets.has(row.target) || row.runner_family !== "go_test") {
      continue;
    }
    if (row.target === "backend-integration" && row.coverage === "raw") {
      addAggregate(aggregates, row, "raw");
      const aggregateKey = rawAggregateBaselineKey(row.target, row.execution_family);
      for (const pkg of row.packages) {
        const importPath = toGoImportPath(modulePath, pkg);
        const key = rawPackageBaselineKey(row.target, row.execution_family, importPath);
        const { weightMs, weightSource } = rawItemWeight(
          baselines,
          key,
          aggregateKey,
          row.packages.length,
        );
        executableItems.push({
          target: row.target,
          aggregate_name: row.execution_family,
          kind: "raw",
          id: `${row.id}:${pkg}`,
          packages: [pkg],
          regex: row.raw_selector,
          symbol: "",
          import_path: importPath,
          package_import_paths: [importPath],
          weight_ms: weightMs,
          weight_source: weightSource,
          baseline_key: key,
          legacy_baseline_key: aggregateKey,
          shard_isolation: row.shard_isolation === true,
          postgres_fixture_policy: normalizePostgresFixturePolicy(row.fixture_policy?.postgres),
          postgres_fixture_budget: row.fixture_budget?.postgres ?? {},
        });
      }
      continue;
    }
    if (
      (row.target === "backend-integration" || row.target === "backend-store") &&
      row.coverage === "authoritative"
    ) {
      addAggregate(aggregates, row, "manifest");
      for (const symbol of row.symbols) {
        const importPath = toGoImportPath(modulePath, row.package);
        const key = testBaselineKey(importPath, symbol);
        const weightMs = normalizePositiveInteger(
          baselines.tests.get(key),
          baselines.defaultItemWeightMs,
        );
        executableItems.push({
          target: row.target,
          aggregate_name: row.execution_family,
          kind: "authoritative",
          id: row.id,
          packages: rowPackages(row),
          regex: "",
          symbol,
          import_path: importPath,
          package_import_paths: rowPackageImportPaths(modulePath, row),
          weight_ms: weightMs,
          weight_source: baselines.tests.has(key) ? "baseline" : "default",
          baseline_key: key,
          shard_isolation: row.shard_isolation === true,
          postgres_fixture_policy: normalizePostgresFixturePolicy(row.fixture_policy?.postgres),
          postgres_fixture_budget: row.fixture_budget?.postgres ?? {},
        });
      }
      continue;
    }
    if (row.target === "backend-integration-support" && row.support_only) {
      addAggregate(aggregates, row, "support");
      for (const symbol of row.symbols) {
        const importPath = toGoImportPath(modulePath, row.package);
        const key = testBaselineKey(importPath, symbol);
        const weightMs = normalizePositiveInteger(
          baselines.tests.get(key),
          baselines.defaultItemWeightMs,
        );
        executableItems.push({
          target: row.target,
          aggregate_name: row.execution_family,
          kind: "support",
          id: row.id,
          packages: rowPackages(row),
          regex: "",
          symbol,
          import_path: importPath,
          package_import_paths: rowPackageImportPaths(modulePath, row),
          weight_ms: weightMs,
          weight_source: baselines.tests.has(key) ? "baseline" : "default",
          baseline_key: key,
          shard_isolation: row.shard_isolation === true,
          postgres_fixture_policy: normalizePostgresFixturePolicy(row.fixture_policy?.postgres),
          postgres_fixture_budget: row.fixture_budget?.postgres ?? {},
        });
      }
    }
  }

  return { baselines, aggregates, executableItems };
}

function shardWeightMs(items, baselines) {
  if (items.length > 0 && items.every((item) => item.kind === "raw")) {
    return items.reduce((sum, item) => sum + item.weight_ms, 0);
  }
  let weightMs = items.reduce((sum, item) => sum + item.weight_ms, 0);
  const packageKeys = new Set();
  const targets = new Set();
  for (const item of items) {
    targets.add(item.target);
    for (const importPath of item.package_import_paths) {
      packageKeys.add(`${item.target}::${importPath}`);
    }
  }
  for (const key of packageKeys) {
    weightMs += normalizePositiveInteger(baselines.packageOverheads.get(key), 0);
  }
  for (const target of targets) {
    weightMs += normalizePositiveInteger(baselines.commandOverheadsByTarget.get(target), 0);
  }
  return weightMs;
}

function packShardLane(aggregateName, items, targetMs, baselines) {
  if (items.length === 0) {
    return [];
  }
  const sorted = [...items].sort(
    (left, right) =>
      right.weight_ms - left.weight_ms ||
      compareStrings(left.symbol || left.id, right.symbol || right.id),
  );
  const bins = [];
  for (const item of sorted) {
    if (item.shard_isolation) {
      const isolatedItems = [item];
      bins.push({ aggregateName, items: isolatedItems, weight_ms: shardWeightMs(isolatedItems, baselines), isolated: true });
      continue;
    }
    let selected = null;
    for (const bin of bins) {
      if (bin.isolated) {
        continue;
      }
      const nextWeightMs = shardWeightMs([...bin.items, item], baselines);
      if (nextWeightMs <= targetMs) {
        selected = bin;
        break;
      }
    }
    if (!selected) {
      selected = { aggregateName, items: [], weight_ms: 0 };
      bins.push(selected);
    }
    selected.items.push(item);
    selected.weight_ms = shardWeightMs(selected.items, baselines);
  }
  return bins;
}

function packAggregateItems(aggregateName, items, targetMs, baselines) {
  if (items.length === 0) {
    return [];
  }
  const lanes = new Map();
  for (const item of items) {
    if (!lanes.has(item.kind)) {
      lanes.set(item.kind, []);
    }
    lanes.get(item.kind).push(item);
  }
  const laneOrder = ["raw", "authoritative", "support"];
  const bins = [];
  for (const kind of laneOrder) {
    bins.push(...packShardLane(aggregateName, lanes.get(kind) ?? [], targetMs, baselines));
  }
  return bins;
}

function shardTargetMsForItems(items, baselines, overrideTargetMs) {
  if (Number.isInteger(overrideTargetMs)) {
    return overrideTargetMs;
  }
  const targets = new Set(items.map((item) => item.target));
  let targetMs = Number.POSITIVE_INFINITY;
  for (const target of targets) {
    targetMs = Math.min(
      targetMs,
      normalizePositiveInteger(
        baselines.shardTargetMsByTarget.get(target),
        baselines.defaultShardTargetMs,
      ),
    );
  }
  if (Number.isFinite(targetMs)) {
    return targetMs;
  }
  return baselines.defaultShardTargetMs;
}

function targetOwnsShard(target, shard) {
  if (target === "backend-integration") {
    return shard.items.some((item) => item.kind === "authoritative" || item.kind === "raw");
  }
  if (target === "backend-integration-support") {
    return shard.items.some((item) => item.kind === "support");
  }
  if (target === "backend-store") {
    return shard.items.some((item) => item.kind === "authoritative");
  }
  return false;
}

function targetOwnsAggregate(target, aggregate) {
  return aggregate.target === target;
}

function publicAggregateTargetForMode(mode) {
  return mode === "support" ? "backend-integration-support" : "backend-integration";
}

function publicAggregateTarget(aggregate) {
  if (aggregate.target === "backend-store") {
    return "backend-store";
  }
  return publicAggregateTargetForMode(aggregate.mode);
}

function shardName(aggregateName, index) {
  return `${aggregateName}-shard-${String(index + 1).padStart(2, "0")}`;
}

function schedulerProfileForShard(items, weightMs) {
  const hasResetHeavyFixture = items.some(
    (item) =>
      item.postgres_fixture_policy === "package_reset" &&
      ((item.postgres_fixture_budget?.max_package_resets ?? 0) > 0 ||
        (item.postgres_fixture_budget?.max_reset_duration_ms ?? 0) > 0),
  );
  if (hasResetHeavyFixture) {
    return "reset_heavy";
  }
  const hasIOHeavyFixture = items.some((item) =>
    ioHeavyFixturePolicies.has(item.postgres_fixture_policy),
  );
  if (hasIOHeavyFixture) {
    return "io_heavy";
  }
  if (weightMs >= cpuHeavyShardWeightMs) {
    return "cpu_heavy";
  }
  return "balanced";
}

export function collectGoShardPlan(root = process.cwd(), options = {}) {
  const requestedTargetMs = normalizePositiveInteger(
    options.targetMs,
    Number.NaN,
  );
  const { baselines, aggregates, executableItems } = buildExecutionItems(root);
  const itemsByAggregate = new Map();
  for (const item of executableItems) {
    if (!itemsByAggregate.has(item.aggregate_name)) {
      itemsByAggregate.set(item.aggregate_name, []);
    }
    itemsByAggregate.get(item.aggregate_name).push(item);
  }

  const shards = [];
  for (const [aggregateName, items] of [...itemsByAggregate.entries()].sort()) {
    const targetMs = shardTargetMsForItems(items, baselines, requestedTargetMs);
    const bins = packAggregateItems(aggregateName, items, targetMs, baselines);
    bins.forEach((bin, index) => {
      const packages = new Set();
      const symbols = [];
      const targets = new Set();
      let rawRegex = "";
      const rawRegexes = new Set();
      let hasAuthoritative = false;
      let hasSupport = false;
      let hasRaw = false;
      for (const item of bin.items) {
        targets.add(item.target);
        for (const pkg of item.packages) {
          packages.add(pkg);
        }
        if (item.kind === "raw") {
          rawRegexes.add(item.regex);
          hasRaw = true;
        } else {
          symbols.push(item.symbol);
          hasAuthoritative ||= item.kind === "authoritative";
          hasSupport ||= item.kind === "support";
        }
      }
      shards.push({
        name: shardName(aggregateName, index),
        target: targets.size === 1 ? Array.from(targets)[0] : Array.from(targets).sort(compareStrings).join(","),
        aggregate_name: aggregateName,
        shard_target_ms: targetMs,
        scheduler_profile: schedulerProfileForShard(bin.items, bin.weight_ms),
        regex: hasRaw ? unionRegex(rawRegexes) : exactRegex(symbols.sort(compareStrings)),
        packages: Array.from(packages).sort(compareStrings),
        weight_ms: bin.weight_ms,
        has_authoritative: hasAuthoritative,
        has_support: hasSupport,
        has_raw: hasRaw,
        shared_across_targets: false,
        item_count: bin.items.length,
        items: bin.items
          .map((item) => ({
            kind: item.kind,
            target: item.target,
            id: item.id,
            symbol: item.symbol,
            import_path: item.import_path,
            package_import_paths: item.package_import_paths,
            packages: item.packages,
            weight_ms: item.weight_ms,
            weight_source: item.weight_source,
            baseline_key: item.baseline_key,
            legacy_baseline_key: item.legacy_baseline_key ?? "",
            shard_isolation: item.shard_isolation,
            postgres_fixture_policy: item.postgres_fixture_policy,
            postgres_fixture_budget: item.postgres_fixture_budget,
          }))
          .sort(
            (left, right) =>
              compareStrings(left.kind, right.kind) ||
              compareStrings(left.id, right.id) ||
              compareStrings(left.symbol, right.symbol),
          ),
      });
    });
  }

  for (const aggregate of aggregates.values()) {
    aggregate.target = publicAggregateTarget(aggregate);
  }
  for (const shard of shards) {
    for (const aggregate of aggregates.values()) {
      if (aggregate.name !== shard.aggregate_name) {
        continue;
      }
      if (
        (aggregate.mode === "raw" && shard.has_raw) ||
        (aggregate.mode === "manifest" && shard.has_authoritative) ||
        (aggregate.mode === "support" && shard.has_support)
      ) {
        aggregate.shards.add(shard.name);
        aggregate.weight_ms += shard.weight_ms;
      }
    }
  }

  const aggregateList = Array.from(aggregates.values())
    .filter((aggregate) => aggregate.shards.size > 0)
    .map((aggregate) => ({
      ...aggregate,
      packages: Array.from(aggregate.packages).sort(compareStrings),
      shards: Array.from(aggregate.shards).sort(compareStrings),
    }))
    .sort(
      (left, right) =>
        compareStrings(left.target, right.target) ||
        compareStrings(left.phase, right.phase) ||
        compareStrings(left.name, right.name),
    );

  const shardList = shards.sort(
    (left, right) =>
      right.weight_ms - left.weight_ms ||
      compareStrings(left.aggregate_name, right.aggregate_name) ||
      compareStrings(left.name, right.name),
  );

  return {
    schema_id: "cartulary.go_shard_plan.v3",
    default_shard_target_ms: baselines.defaultShardTargetMs,
    shard_target_ms_by_target: Object.fromEntries(
      [...baselines.shardTargetMsByTarget.entries()].sort(([left], [right]) =>
        compareStrings(left, right),
      ),
    ),
    default_item_weight_ms: baselines.defaultItemWeightMs,
    targets: Array.from(shardTargets).sort(compareStrings),
    aggregates: aggregateList,
    shards: shardList,
  };
}

function fixturePolicyAssignmentsForShard(shard, mode) {
  const assignments = [];
  for (const item of shard.items) {
    if (!item.postgres_fixture_policy) {
      continue;
    }
    if (item.postgres_fixture_policy === "migration_scratch") {
      continue;
    }
    if (mode === "tests" && item.symbol) {
      assignments.push(`${item.symbol}=${item.postgres_fixture_policy}`);
      continue;
    }
    if (mode === "packages" && item.kind === "raw") {
      for (const pkg of item.packages) {
        assignments.push(`${pkg}=${item.postgres_fixture_policy}`);
      }
    }
  }
  return assignments.sort();
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
      continue;
    }
    if (mode === "packages" && item.kind === "raw") {
      for (const pkg of item.packages) {
        assignments.push(`${pkg}=${dirtyTables.join("|")}`);
      }
    }
  }
  return assignments.sort();
}

function targetShards(plan, target) {
  if (!shardTargets.has(target)) {
    throw new Error(`unsupported Go shard target ${target}`);
  }
  const aggregateNames = new Set(
    plan.aggregates
      .filter((aggregate) => targetOwnsAggregate(target, aggregate))
      .map((aggregate) => aggregate.name),
  );
  return plan.shards.filter(
    (shard) => aggregateNames.has(shard.aggregate_name) && targetOwnsShard(target, shard),
  );
}

function targetAggregates(plan, target) {
  if (!shardTargets.has(target)) {
    throw new Error(`unsupported Go shard target ${target}`);
  }
  return plan.aggregates.filter((aggregate) => targetOwnsAggregate(target, aggregate));
}

export function collectGoShardsForTarget(root = process.cwd(), target) {
  const plan = collectGoShardPlan(root);
  return targetShards(plan, target);
}

function findShard(plan, target, name) {
  const shard = targetShards(plan, target).find((candidate) => candidate.name === name);
  if (!shard) {
    throw new Error(`unknown shard ${name} for ${target}`);
  }
  return shard;
}

function findAggregate(plan, target, name) {
  const aggregate = targetAggregates(plan, target).find((candidate) => candidate.name === name);
  if (!aggregate) {
    throw new Error(`unknown aggregate ${name} for ${target}`);
  }
  return aggregate;
}

function printLines(lines) {
  process.stdout.write(`${lines.join("\n")}\n`);
}

function main(argv) {
  const [command, target, name] = argv;
  const plan = collectGoShardPlan(process.cwd());
  switch (command) {
    case "json":
      process.stdout.write(`${JSON.stringify(plan, null, 2)}\n`);
      return;
    case "list-shards":
      printLines(targetShards(plan, target).map((shard) => shard.name));
      return;
    case "shard-spec": {
      const shard = findShard(plan, target, name);
      printLines([shard.regex, ...shard.packages]);
      return;
    }
    case "shard-postgres-fixture-policy-tests": {
      const shard = findShard(plan, target, name);
      printLines([fixturePolicyAssignmentsForShard(shard, "tests").join(",")]);
      return;
    }
    case "shard-postgres-fixture-policy-packages": {
      const shard = findShard(plan, target, name);
      printLines([fixturePolicyAssignmentsForShard(shard, "packages").join(",")]);
      return;
    }
    case "shard-postgres-reset-table-tests": {
      const shard = findShard(plan, target, name);
      printLines([resetTableAssignmentsForShard(shard, "tests").join(",")]);
      return;
    }
    case "shard-postgres-reset-table-packages": {
      const shard = findShard(plan, target, name);
      printLines([resetTableAssignmentsForShard(shard, "packages").join(",")]);
      return;
    }
    case "shard-field": {
      const field = argv[3];
      const shard = findShard(plan, target, name);
      const value = shard[field] ?? "";
      process.stdout.write(`${String(value)}\n`);
      return;
    }
    case "list-aggregates":
      printLines(targetAggregates(plan, target).map((aggregate) => aggregate.name));
      return;
    case "aggregate-shards": {
      const aggregate = findAggregate(plan, target, name);
      printLines(aggregate.shards);
      return;
    }
    case "aggregate-packages": {
      const aggregate = findAggregate(plan, target, name);
      printLines(aggregate.packages);
      return;
    }
    case "aggregate-field": {
      const field = argv[3];
      const aggregate = findAggregate(plan, target, name);
      const value = aggregate[field] ?? "";
      process.stdout.write(`${String(value)}\n`);
      return;
    }
    default:
      throw new Error(
        "usage: go-shard-plan.mjs <json|list-shards|shard-spec|shard-field|list-aggregates|aggregate-shards|aggregate-packages|aggregate-field> [target] [name] [field]",
      );
  }
}

if (import.meta.url === `file://${process.argv[1]}`) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    process.stderr.write(`${error.message}\n`);
    process.exit(1);
  }
}
