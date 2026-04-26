#!/usr/bin/env node
import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { collectGoShardPlan } from "./lib/go-shard-plan.mjs";
import { collectTargetPlanRows } from "./lib/target-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultTargets = [
  "backend-store",
  "backend-integration",
  "backend-integration-support",
  "backend-process",
];

function usage() {
  process.stderr.write("usage: check-postgres-fixture-budget.mjs [--targets <csv>]\n");
  process.exit(2);
}

function parseArgs(argv) {
  const options = { targets: defaultTargets };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--targets") {
      const raw = argv[index + 1] ?? "";
      index += 1;
      options.targets = raw
        .split(",")
        .map((value) => value.trim())
        .filter(Boolean);
      continue;
    }
    usage();
  }
  return options;
}

function eventDirs() {
  const resultsDir = path.resolve(
    repoRoot,
    process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results",
  );
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!runID) {
    throw new Error("CARTULARY_TEST_RUN_ID is required");
  }
  const servicesRoot = path.join(resultsDir, runID, "_shared", "test-services");
  if (!existsSync(servicesRoot)) {
    return [];
  }
  return readdirSync(servicesRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(servicesRoot, entry.name, "events"))
    .filter((dir) => existsSync(dir));
}

function loadEvents() {
  const events = [];
  const stack = eventDirs();
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (!entry.isFile() || !entry.name.endsWith(".json")) {
        continue;
      }
      events.push(JSON.parse(readFileSync(next, "utf8")));
    }
  }
  events.sort(
    (left, right) =>
      String(left.timestamp ?? "").localeCompare(String(right.timestamp ?? "")) ||
      String(left.name ?? "").localeCompare(String(right.name ?? "")),
  );
  return events;
}

function topLevelTestName(testName) {
  return String(testName ?? "").split("/", 1)[0];
}

function intDetail(details, key) {
  const value = details?.[key];
  return Number.isInteger(value) ? value : Number.isFinite(value) ? Math.trunc(value) : 0;
}

function normalizePackage(value) {
  return String(value ?? "")
    .trim()
    .replace(/^\.\//, "")
    .replace(/^github\.com\/JochiRaider\/cartulary\//, "");
}

function emptyBudget() {
  return {
    templateClones: 0,
    groupClones: 0,
    packageResetCreates: 0,
    packageResetEvents: 0,
    packageResetDurationMS: 0,
    transactions: 0,
    migrationScratchCreates: 0,
    packageResetTests: new Set(),
    templateCloneTests: new Set(),
    groupCloneTests: new Set(),
    transactionTests: new Set(),
    migrationScratchTests: new Set(),
    packageResetPackages: new Set(),
  };
}

function addBudgetValue(targetBudget, policy, budget, item) {
  const symbol = item.symbol ?? "";
  const pkg = normalizePackage(item.import_path || item.package || item.packages?.[0] || "");
  switch (policy) {
    case "template_clone":
      targetBudget.templateClones += budget.max_template_clones ?? 0;
      if (symbol) targetBudget.templateCloneTests.add(symbol);
      break;
    case "group_clone":
      targetBudget.groupClones += budget.max_group_clones ?? 0;
      if (symbol) targetBudget.groupCloneTests.add(symbol);
      break;
    case "package_reset":
      targetBudget.packageResetEvents += budget.max_package_resets ?? 0;
      targetBudget.packageResetDurationMS += budget.max_reset_duration_ms ?? 0;
      if (symbol) targetBudget.packageResetTests.add(symbol);
      if (pkg) targetBudget.packageResetPackages.add(pkg);
      break;
    case "transaction":
      targetBudget.transactions += budget.max_transactions ?? 0;
      if (symbol) targetBudget.transactionTests.add(symbol);
      break;
    case "migration_scratch":
      targetBudget.migrationScratchCreates += budget.max_migration_scratch ?? 0;
      if (symbol) targetBudget.migrationScratchTests.add(symbol);
      break;
  }
}

function plannedBudget(target) {
  const budget = emptyBudget();
  if (["backend-store", "backend-integration", "backend-integration-support"].includes(target)) {
    const targetPlan = collectGoShardPlan(repoRoot);
    const aggregateNames = new Set(
      targetPlan.aggregates
        .filter((aggregate) => aggregate.target === target)
        .map((aggregate) => aggregate.name),
    );
    const shards = targetPlan.shards.filter((shard) => aggregateNames.has(shard.aggregate_name));
    for (const shard of shards) {
      const shardPackageResetPackages = new Set();
      for (const item of shard.items) {
        addBudgetValue(budget, item.postgres_fixture_policy, item.postgres_fixture_budget ?? {}, item);
        if (item.postgres_fixture_policy !== "package_reset") {
          continue;
        }
        if (item.kind === "raw") {
          for (const pkg of item.packages ?? []) {
            shardPackageResetPackages.add(normalizePackage(pkg));
          }
          continue;
        }
        const pkg = normalizePackage(item.import_path);
        if (pkg) {
          shardPackageResetPackages.add(pkg);
        }
      }
      budget.packageResetCreates += shardPackageResetPackages.size;
    }
    return budget;
  }

  for (const row of collectTargetPlanRows(repoRoot).filter((candidate) => candidate.target === target)) {
    const policy = row.fixture_policy?.postgres ?? "";
    const rowBudget = row.fixture_budget?.postgres ?? {};
    const symbols = row.symbols?.length ? row.symbols : [""];
    for (const symbol of symbols) {
      addBudgetValue(budget, policy, rowBudget, {
        symbol,
        package: row.package,
        packages: row.packages,
      });
    }
  }
  budget.packageResetCreates = budget.packageResetPackages.size;
  return budget;
}

function actualStats(events, target) {
  const stats = {
    templateClones: 0,
    groupClones: 0,
    packageResetCreates: 0,
    packageResetEvents: 0,
    packageResetDurationMS: 0,
    transactions: 0,
    migrationScratchCreates: 0,
    migrationScratchDetails: [],
    unplannedTemplateClones: [],
    unplannedGroupClones: [],
    unplannedTransactions: [],
  };

  for (const event of events) {
    const details = event.details ?? {};
    if (details.target !== target) {
      continue;
    }
    const policy = details.fixture_policy ?? "";
    const reuseScope = details.reuse_scope ?? "per-test";
    const testName = topLevelTestName(details.test_name);
    if (event.type === "postgres-db-created" && event.kind === "template-clone") {
      if (reuseScope === "group-reused" || policy === "group_clone") {
        stats.groupClones += 1;
      } else if (reuseScope === "per-test" && policy === "template_clone") {
        stats.templateClones += 1;
      }
    }
    if (
      event.type === "postgres-db-created" &&
      event.kind === "template-clone" &&
      reuseScope === "package-reused" &&
      policy === "package_reset"
    ) {
      stats.packageResetCreates += 1;
    }
    if (
      event.type === "postgres-db-created" &&
      event.kind === "scratch" &&
      reuseScope === "migration-scratch"
    ) {
      stats.migrationScratchCreates += 1;
      stats.migrationScratchDetails.push(
        `${testName || "(unknown test)"} ${details.caller_file ?? ""}`.trim(),
      );
    }
    if (
      event.type === "postgres-db-reset" &&
      reuseScope === "package-reused" &&
      policy === "package_reset"
    ) {
      stats.packageResetEvents += 1;
      stats.packageResetDurationMS += intDetail(details, "duration_ms");
    }
    if (event.type === "postgres-transaction") {
      stats.transactions += 1;
    }
    if (
      event.type === "postgres-db-created" &&
      event.kind === "template-clone" &&
      reuseScope === "per-test" &&
      policy !== "template_clone"
    ) {
      if (testName || details.caller_file || details.caller_package) {
        stats.unplannedTemplateClones.push(`${testName || "(unknown test)"} ${details.caller_file ?? ""}`.trim());
      }
    }
    if (
      event.type === "postgres-db-created" &&
      event.kind === "template-clone" &&
      reuseScope === "group-reused" &&
      policy !== "group_clone"
    ) {
      stats.unplannedGroupClones.push(`${testName || "(unknown test)"} ${details.caller_file ?? ""}`.trim());
    }
    if (event.type === "postgres-transaction" && policy !== "transaction") {
      stats.unplannedTransactions.push(`${testName || "(unknown test)"} ${details.caller_file ?? ""}`.trim());
    }
  }
  return stats;
}

function failIfOver(target, name, actual, budget) {
  if (actual > budget) {
    throw new Error(`${target} exceeded postgres ${name} budget: got ${actual}, budget ${budget}`);
  }
}

function failIfMigrationScratchOver(target, stats, budget) {
  if (stats.migrationScratchCreates <= budget.migrationScratchCreates) {
    return;
  }
  const actual =
    stats.migrationScratchDetails.length > 0
      ? stats.migrationScratchDetails.join("; ")
      : "none";
  const planned =
    budget.migrationScratchTests.size > 0
      ? [...budget.migrationScratchTests].sort().join(",")
      : "none";
  throw new Error(
    `${target} exceeded postgres migration scratch create budget: got ${stats.migrationScratchCreates}, budget ${budget.migrationScratchCreates}; actual=${actual}; planned_manifest_symbols=${planned}`,
  );
}

function checkTarget(events, target) {
  const budget = plannedBudget(target);
  const stats = actualStats(events, target);
  failIfOver(target, "template clone", stats.templateClones, budget.templateClones);
  failIfOver(target, "group clone", stats.groupClones, budget.groupClones);
  failIfOver(target, "package database create", stats.packageResetCreates, budget.packageResetCreates);
  failIfOver(target, "package reset event", stats.packageResetEvents, budget.packageResetEvents);
  failIfOver(target, "package reset duration", stats.packageResetDurationMS, budget.packageResetDurationMS);
  failIfOver(target, "transaction", stats.transactions, budget.transactions);
  failIfMigrationScratchOver(target, stats, budget);
  if (stats.unplannedTemplateClones.length > 0) {
    throw new Error(`${target} used unplanned per-test postgres template clones: ${stats.unplannedTemplateClones.join("; ")}`);
  }
  if (stats.unplannedGroupClones.length > 0) {
    throw new Error(`${target} used unplanned postgres group clones: ${stats.unplannedGroupClones.join("; ")}`);
  }
  if (stats.unplannedTransactions.length > 0) {
    throw new Error(`${target} used unplanned postgres transactions: ${stats.unplannedTransactions.join("; ")}`);
  }
}

try {
  const options = parseArgs(process.argv.slice(2));
  if (options.targets.length === 0) {
    process.exit(0);
  }
  const events = loadEvents();
  for (const target of options.targets) {
    checkTarget(events, target);
  }
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
}
