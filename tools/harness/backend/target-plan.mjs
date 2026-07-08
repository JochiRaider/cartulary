import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadExecutionTopology } from "../generated-artifacts/execution-topology.mjs";
import {
  aggregatePackages,
  aggregateRegex,
  collectAggregateEmissions,
  fixturePolicyAssignments,
  resetTableAssignments,
} from "./go-target-aggregate.mjs";
import {
  collectEntries,
  collectSupportGoEntries,
  entryIsExecutable,
  goEntryScenarioSymbols,
  goEntryPostgresFixtureBudget,
  goEntryPostgresFixturePolicy,
  goEntrySymbolFixtureDetails,
  goEntrySymbols,
  loadManifest,
  phaseManifestNames,
  supportGoEntryPostgresFixtureBudget,
  supportGoEntryPostgresFixturePolicy,
  supportGoEntrySymbolFixtureDetails,
  supportGoEntrySymbols,
} from "../phase-accounting/index.mjs";

const validShardModes = new Set(["none", "go_shards"]);
const validParallelismModes = new Set(["none", "package", "process"]);
const executionTargetsCache = new Map();
const targetPlanRowsCache = new Map();

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function compareRows(left, right) {
  return (
    compareStrings(left.target, right.target) ||
    compareStrings(left.execution_family, right.execution_family) ||
    compareStrings(left.manifest_phase, right.manifest_phase) ||
    compareStrings(left.section, right.section) ||
    compareStrings(left.id, right.id)
  );
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function loadExecutionTargets(root) {
  const cacheKey = path.resolve(root);
  const cached = executionTargetsCache.get(cacheKey);
  if (cached) {
    return cached;
  }
  const topology = loadExecutionTopology({ root });
  for (const descriptor of topology.goTargets.targets) {
    const label = `execution topology go target ${descriptor.name}`;
    if (!validShardModes.has(descriptor.sharding)) {
      throw new Error(`${label}.sharding must be none|go_shards`);
    }
    if (!validParallelismModes.has(descriptor.goTestParallelism)) {
      throw new Error(`${label}.go_test_parallelism must be none|package|process`);
    }
  }
  const result = {
    descriptors: topology.goTargets.targets,
    byName: topology.goTargets.byName,
    dependencyTargets: topology.goTargets.dependencyTargets,
    supportTargets: topology.goTargets.supportTargets,
    rawAggregates: topology.goTargets.rawAggregates,
  };
  executionTargetsCache.set(cacheKey, result);
  return result;
}

function rowBase(descriptor) {
  return {
    target: descriptor.name,
    service_backed: descriptor.serviceBacked,
    runner_family: "go_test",
    check_heavy_safe: descriptor.checkHeavySafe,
    check_service_backed_safe: descriptor.checkServiceBackedSafe,
    check_isolated_safe: descriptor.checkIsolatedSafe,
    canonical_authoritative: descriptor.canonicalAuthoritative,
    sharding: descriptor.sharding,
    go_test_parallelism: descriptor.goTestParallelism,
  };
}

function requireExecutionFamily(entry, label) {
  return {
    family: requireString(entry.execution_family, `${label}.execution_family`),
    label: requireString(entry.execution_label, `${label}.execution_label`),
  };
}

function manifestRows(phase, descriptor, entry) {
  const family = requireExecutionFamily(entry, `manifest entry ${entry.id}`);
  const scenarioSymbols = goEntryScenarioSymbols(entry);
  return {
    ...rowBase(descriptor),
    id: entry.id,
    manifest_phase: phase,
    section: entry.section,
    coverage: entry.coverage,
    execution_dependency: entry.execution_dependency,
    evidence_class: entry.evidence_class,
    layer: entry.layer,
    default_check_required: entry.default_check_required,
    primary_evidence_owner: entry.primary_evidence_owner,
    ...(entry.default_check_reason ? { default_check_reason: entry.default_check_reason } : {}),
    execution_family: family.family,
    execution_label: family.label,
    packages: [entry.package],
    support_only: false,
    support_selector: null,
    raw_selector: null,
    file: entry.file,
    package: entry.package,
    symbols: goEntrySymbols(entry),
    ...(Object.keys(scenarioSymbols).length > 0 ? { scenario_symbols: scenarioSymbols } : {}),
    runtime_binaries: [...(entry.runtime_binaries ?? [])],
    shard_isolation: entry.shard_isolation === true,
    evidence_layer: entry.evidence_layer,
    fixture_policy: {
      postgres: goEntryPostgresFixturePolicy(entry),
    },
    fixture_budget: {
      postgres: goEntryPostgresFixtureBudget(entry),
    },
    symbol_fixture_details: goEntrySymbolFixtureDetails(entry),
  };
}

function supportID(phase, target, file, symbol) {
  const normalized = `${phase}-${target}-${file}-${symbol}`.replace(/[^A-Za-z0-9]+/g, "-");
  return `SUPPORT-${normalized.replace(/^-|-$/g, "")}`;
}

function supportRows(phase, descriptor, entry) {
  const family = requireExecutionFamily(entry, `support_go_target ${entry.target} ${entry.file}`);
  return supportGoEntrySymbols(entry).map((symbol) => ({
    ...rowBase(descriptor),
    id: supportID(phase, entry.target, entry.file, symbol),
    manifest_phase: phase,
    section: entry.section,
    coverage: "support",
    execution_dependency: entry.target,
    evidence_class: entry.evidence_class,
    layer: entry.layer,
    default_check_required: entry.default_check_required,
    primary_evidence_owner: entry.primary_evidence_owner,
    ...(entry.default_check_reason ? { default_check_reason: entry.default_check_reason } : {}),
    execution_family: family.family,
    execution_label: family.label,
    packages: [entry.package],
    support_only: true,
    support_selector: entry.selection_pattern,
    raw_selector: null,
    file: entry.file,
    package: entry.package,
    symbols: [symbol],
    runtime_binaries: [...(entry.runtime_binaries ?? [])],
    shard_isolation: entry.shard_isolation === true,
    evidence_layer: "support",
    fixture_policy: {
      postgres: supportGoEntryPostgresFixturePolicy(entry),
    },
    fixture_budget: {
      postgres: supportGoEntryPostgresFixtureBudget(entry),
    },
    symbol_fixture_details: supportGoEntrySymbolFixtureDetails(entry),
  }));
}

function rawRows(config, aggregate) {
  const descriptor = config.byName.get(aggregate.target);
  return {
    ...rowBase(descriptor),
    id: `RAW-${aggregate.id}`,
    manifest_phase: "",
    section: aggregate.section,
    coverage: "raw",
    execution_dependency: "",
    evidence_class: "diagnostic",
    layer: "raw",
    default_check_required: false,
    execution_family: aggregate.executionFamily,
    execution_label: aggregate.executionLabel,
    packages: [...aggregate.packages],
    support_only: false,
    support_selector: null,
    raw_selector: aggregate.selectionPattern,
    file: "",
    package: "",
    symbols: [],
    runtime_binaries: [],
    shard_isolation: false,
    evidence_layer: "raw",
    label: aggregate.executionLabel,
    fixture_policy: aggregate.fixturePolicy,
    fixture_budget: aggregate.fixtureBudget,
  };
}

export function collectTargetPlanRows(root = process.cwd()) {
  const cacheKey = path.resolve(root);
  const cached = targetPlanRowsCache.get(cacheKey);
  if (cached) {
    return cached;
  }
  const config = loadExecutionTargets(root);
  const rows = [];
  for (const phase of phaseManifestNames(root)) {
    const { manifest } = loadManifest(root, phase);
    for (const entry of collectEntries(manifest)) {
      if (entry.runner !== "go_test") {
        continue;
      }
      if (!entryIsExecutable(entry)) {
        continue;
      }
      const descriptor = config.dependencyTargets.get(entry.execution_dependency);
      if (!descriptor) {
        continue;
      }
      rows.push(manifestRows(phase, descriptor, entry));
    }
    for (const entry of collectSupportGoEntries(manifest)) {
      const descriptor = config.supportTargets.get(entry.target);
      if (!descriptor) {
        continue;
      }
      rows.push(...supportRows(phase, descriptor, entry));
    }
  }
  for (const aggregate of config.rawAggregates) {
    rows.push(rawRows(config, aggregate));
  }
  const result = rows.sort(compareRows);
  targetPlanRowsCache.set(cacheKey, result);
  return result;
}

export function collectTargetNames(root = process.cwd()) {
  return loadExecutionTargets(root).descriptors.map((descriptor) => descriptor.name);
}

export function findTargetDescriptor(target, root = process.cwd()) {
  return loadExecutionTargets(root).byName.get(target) ?? null;
}

export function knownManifestPhases(root = process.cwd()) {
  return phaseManifestNames(root);
}

function rowsForAggregate(root, target, executionFamily) {
  const rows = collectTargetPlanRows(root).filter(
    (row) => row.target === target && row.execution_family === executionFamily,
  );
  if (rows.length === 0) {
    throw new Error(`unknown execution family ${executionFamily} for ${target}`);
  }
  return rows;
}

function aggregateNames(root, target) {
  const names = new Set(
    collectTargetPlanRows(root)
      .filter((row) => row.target === target)
      .map((row) => row.execution_family),
  );
  return Array.from(names).sort(compareStrings);
}

function printLines(lines) {
  process.stdout.write(`${lines.join("\n")}\n`);
}

function main(argv) {
  const [command, target, family, extra] = argv;
  const root = process.cwd();
  switch (command) {
    case "list-targets":
      printLines(collectTargetNames(root));
      return;
    case "target-field": {
      const descriptor = findTargetDescriptor(target, root);
      if (!descriptor) {
        throw new Error(`unknown target ${target}`);
      }
      const field = family;
      const value = descriptor[field] ?? "";
      process.stdout.write(`${String(value)}\n`);
      return;
    }
    case "list-aggregates":
      printLines(aggregateNames(root, target));
      return;
    case "aggregate-spec": {
      const rows = rowsForAggregate(root, target, family);
      printLines([aggregateRegex(rows), ...aggregatePackages(rows)]);
      return;
    }
    case "aggregate-postgres-fixture-policy-tests":
      printLines([fixturePolicyAssignments(rowsForAggregate(root, target, family), "tests").join(",")]);
      return;
    case "aggregate-postgres-fixture-policy-packages":
      printLines([fixturePolicyAssignments(rowsForAggregate(root, target, family), "packages").join(",")]);
      return;
    case "aggregate-postgres-reset-table-tests":
      printLines([resetTableAssignments(rowsForAggregate(root, target, family), "tests").join(",")]);
      return;
    case "aggregate-postgres-reset-table-packages":
      printLines([resetTableAssignments(rowsForAggregate(root, target, family), "packages").join(",")]);
      return;
    case "aggregate-emission-count":
      printLines([String(collectAggregateEmissions(rowsForAggregate(root, target, family)).length)]);
      return;
    case "aggregate-emission-field": {
      const index = Number.parseInt(extra, 10);
      const field = argv[4];
      const emission = collectAggregateEmissions(rowsForAggregate(root, target, family))[index];
      if (!emission) {
        throw new Error(`unknown emission index ${extra} for ${target} ${family}`);
      }
      const value = emission[field] ?? "";
      process.stdout.write(`${String(value)}\n`);
      return;
    }
    case "aggregate-emission-packages": {
      const index = Number.parseInt(extra, 10);
      const emission = collectAggregateEmissions(rowsForAggregate(root, target, family))[index];
      if (!emission) {
        throw new Error(`unknown emission index ${extra} for ${target} ${family}`);
      }
      printLines(emission.packages);
      return;
    }
    default:
      throw new Error(
        "usage: target-plan.mjs <list-targets|target-field|list-aggregates|aggregate-spec|aggregate-postgres-fixture-policy-tests|aggregate-postgres-fixture-policy-packages|aggregate-postgres-reset-table-tests|aggregate-postgres-reset-table-packages|aggregate-emission-count|aggregate-emission-field|aggregate-emission-packages> ...",
      );
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exit(1);
  }
}
