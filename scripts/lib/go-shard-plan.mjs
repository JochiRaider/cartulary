import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { collectTargetPlanRows } from "./target-plan.mjs";

const defaultShardTargetMs = 30_000;
const defaultBackendIntegrationShardTargetMs = 18_000;
const defaultIntegrationWeightMs = 10_000;
const baselinePath = path.join("tools", "go_test_duration_baselines.json");
const shardTargets = new Set(["backend-store", "backend-integration", "backend-integration-support"]);
const executionTargets = new Set(["backend-store", "backend-integration", "backend-integration-support"]);
const defaultShardTargetMsByTarget = new Map([
  ["backend-store", defaultShardTargetMs],
  ["backend-integration", defaultBackendIntegrationShardTargetMs],
  ["backend-integration-support", defaultBackendIntegrationShardTargetMs],
]);

const aggregateLabelOverrides = new Map([
  ["backend-integration-testutil", "backend-integration testutil"],
  ["backend-integration-phase0-platform", "backend-integration phase0 authoritative platform"],
  ["backend-integration-phase0-app", "backend-integration phase0 authoritative app"],
  ["backend-integration-auth", "backend-integration phase1 authoritative"],
  ["backend-integration-phase2-incidents", "backend-integration phase2 authoritative"],
  ["backend-integration-phase3-timeline", "backend-integration phase3 authoritative"],
  ["backend-integration-phase4-entities", "backend-integration phase4 authoritative entities"],
  ["backend-integration-phase4-timeline", "backend-integration phase4 authoritative timeline"],
  ["backend-store-shared", "backend-store authoritative"],
]);

const supportLabelOverrides = new Map([
  ["backend-integration-phase0-platform", "backend-integration support phase0 platform"],
  ["backend-integration-auth", "backend-integration support phase1"],
  ["backend-integration-phase2-incidents", "backend-integration support phase2"],
  ["backend-integration-phase3-timeline", "backend-integration support phase3"],
  ["backend-integration-phase4-entities", "backend-integration support phase4 entities"],
]);

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

function loadDurationBaselines(root) {
  const file = path.join(root, baselinePath);
  if (!existsSync(file)) {
    return {
      defaultShardTargetMs,
      shardTargetMsByTarget: new Map(defaultShardTargetMsByTarget),
      defaultIntegrationWeightMs,
      tests: new Map(),
    };
  }
  const raw = JSON.parse(readFileSync(file, "utf8"));
  const tests = new Map(Object.entries(raw.tests ?? raw));
  const shardTargetMsByTarget = new Map(defaultShardTargetMsByTarget);
  for (const [target, targetMs] of Object.entries(raw.shard_target_ms_by_target ?? {})) {
    if (shardTargets.has(target)) {
      shardTargetMsByTarget.set(
        target,
        normalizePositiveInteger(targetMs, shardTargetMsByTarget.get(target)),
      );
    }
  }
  return {
    defaultShardTargetMs: normalizePositiveInteger(
      raw.default_shard_target_ms,
      defaultShardTargetMs,
    ),
    shardTargetMsByTarget,
    defaultIntegrationWeightMs: normalizePositiveInteger(
      raw.default_integration_weight_ms,
      defaultIntegrationWeightMs,
    ),
    tests,
  };
}

function normalizePositiveInteger(value, fallback) {
  if (Number.isInteger(value) && value > 0) {
    return value;
  }
  return fallback;
}

function rowPackages(row) {
  if (row.package) {
    return [row.package];
  }
  return [...row.packages];
}

function addAggregate(aggregates, row, mode) {
  const key = `${row.target}\u001f${row.shared_report}`;
  if (!aggregates.has(key)) {
    aggregates.set(key, {
      target: row.target,
      name: row.shared_report,
      mode,
      label:
        mode === "support"
          ? supportLabelOverrides.get(row.shared_report) ?? `${row.shared_report} support`
          : aggregateLabelOverrides.get(row.shared_report) ?? row.label ?? row.shared_report,
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
  const baselines = loadDurationBaselines(root);
  const rows = collectTargetPlanRows(root);
  const aggregates = new Map();
  const executableItems = [];

  for (const row of rows) {
    if (!executionTargets.has(row.target) || row.runner_family !== "go_test") {
      continue;
    }
    if (row.target === "backend-integration" && row.coverage === "raw") {
      addAggregate(aggregates, row, "raw");
      executableItems.push({
        target: row.target,
        aggregate_name: row.shared_report,
        kind: "raw",
        id: row.id,
        packages: [...row.packages],
        regex: row.raw_selector,
        symbol: "",
        import_path: "",
        weight_ms: baselines.defaultIntegrationWeightMs,
        weight_source: "default",
        shard_isolation: row.shard_isolation === true,
        postgres_fixture_policy: normalizePostgresFixturePolicy(row.fixture_policy?.postgres),
        postgres_fixture_budget: row.fixture_budget?.postgres ?? {},
      });
      continue;
    }
    if (
      (row.target === "backend-integration" || row.target === "backend-store") &&
      row.coverage === "authoritative"
    ) {
      addAggregate(aggregates, row, "manifest");
      for (const symbol of row.symbols) {
        const importPath = toGoImportPath(modulePath, row.package);
        const key = `${importPath}::${symbol}`;
        const weightMs = normalizePositiveInteger(
          baselines.tests.get(key),
          baselines.defaultIntegrationWeightMs,
        );
        executableItems.push({
          target: row.target,
          aggregate_name: row.shared_report,
          kind: "authoritative",
          id: row.id,
          packages: rowPackages(row),
          regex: "",
          symbol,
          import_path: importPath,
          weight_ms: weightMs,
          weight_source: baselines.tests.has(key) ? "baseline" : "default",
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
        const key = `${importPath}::${symbol}`;
        const weightMs = normalizePositiveInteger(
          baselines.tests.get(key),
          baselines.defaultIntegrationWeightMs,
        );
        executableItems.push({
          target: row.target,
          aggregate_name: row.shared_report,
          kind: "support",
          id: row.id,
          packages: rowPackages(row),
          regex: "",
          symbol,
          import_path: importPath,
          weight_ms: weightMs,
          weight_source: baselines.tests.has(key) ? "baseline" : "default",
          shard_isolation: row.shard_isolation === true,
          postgres_fixture_policy: normalizePostgresFixturePolicy(row.fixture_policy?.postgres),
          postgres_fixture_budget: row.fixture_budget?.postgres ?? {},
        });
      }
    }
  }

  return { baselines, aggregates, executableItems };
}

function packAggregateItems(aggregateName, items, targetMs) {
  if (items.length === 0) {
    return [];
  }
  if (items.length === 1 && items[0].kind === "raw") {
    return [{ aggregateName, items: [...items], weight_ms: items[0].weight_ms }];
  }
  const sorted = [...items].sort(
    (left, right) =>
      right.weight_ms - left.weight_ms ||
      compareStrings(left.symbol || left.id, right.symbol || right.id),
  );
  const bins = [];
  for (const item of sorted) {
    if (item.shard_isolation) {
      bins.push({ aggregateName, items: [item], weight_ms: item.weight_ms, isolated: true });
      continue;
    }
    let selected = null;
    for (const bin of bins) {
      if (bin.isolated) {
        continue;
      }
      if (bin.weight_ms + item.weight_ms <= targetMs) {
        selected = bin;
        break;
      }
    }
    if (!selected) {
      selected = { aggregateName, items: [], weight_ms: 0 };
      bins.push(selected);
    }
    selected.items.push(item);
    selected.weight_ms += item.weight_ms;
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
    const bins = packAggregateItems(aggregateName, items, targetMs);
    bins.forEach((bin, index) => {
      const packages = new Set();
      const symbols = [];
      let rawRegex = "";
      let hasAuthoritative = false;
      let hasSupport = false;
      let hasRaw = false;
      for (const item of bin.items) {
        for (const pkg of item.packages) {
          packages.add(pkg);
        }
        if (item.kind === "raw") {
          rawRegex = item.regex;
          hasRaw = true;
        } else {
          symbols.push(item.symbol);
          hasAuthoritative ||= item.kind === "authoritative";
          hasSupport ||= item.kind === "support";
        }
      }
      shards.push({
        name: shardName(aggregateName, index),
        aggregate_name: aggregateName,
        shard_target_ms: targetMs,
        regex: hasRaw ? rawRegex : exactRegex(symbols.sort(compareStrings)),
        packages: Array.from(packages).sort(compareStrings),
        weight_ms: bin.weight_ms,
        has_authoritative: hasAuthoritative,
        has_support: hasSupport,
        has_raw: hasRaw,
        shared_across_targets: hasAuthoritative && hasSupport,
        item_count: bin.items.length,
        items: bin.items
          .map((item) => ({
            kind: item.kind,
            id: item.id,
            symbol: item.symbol,
            import_path: item.import_path,
            packages: item.packages,
            weight_ms: item.weight_ms,
            weight_source: item.weight_source,
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
    schema_id: "cartulary.go_shard_plan.v2",
    default_shard_target_ms: baselines.defaultShardTargetMs,
    shard_target_ms_by_target: Object.fromEntries(
      [...baselines.shardTargetMsByTarget.entries()].sort(([left], [right]) =>
        compareStrings(left, right),
      ),
    ),
    default_integration_weight_ms: baselines.defaultIntegrationWeightMs,
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
