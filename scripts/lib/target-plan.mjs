import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadExecutionTopology } from "./execution-topology.mjs";
import {
  collectEntries,
  collectSupportGoEntries,
  effectiveGoEntryPostgresFixtureBudget,
  effectiveGoEntryPostgresFixturePolicy,
  effectiveSupportGoEntryPostgresFixtureBudget,
  effectiveSupportGoEntryPostgresFixturePolicy,
  goEntrySymbols,
  loadManifest,
  phaseManifestNames,
  supportGoEntrySymbols,
} from "./phase-manifest.mjs";

const validShardModes = new Set(["none", "go_shards"]);
const validParallelismModes = new Set(["none", "package", "process"]);
const postgresFixturePolicyEnvAssignable = new Set([
  "template_clone",
  "package_reset",
  "transaction",
  "group_clone",
]);

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
  return {
    descriptors: topology.goTargets.targets,
    byName: topology.goTargets.byName,
    dependencyTargets: topology.goTargets.dependencyTargets,
    supportTargets: topology.goTargets.supportTargets,
    rawAggregates: topology.goTargets.rawAggregates,
  };
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
  return {
    ...rowBase(descriptor),
    id: entry.id,
    manifest_phase: phase,
    section: entry.section,
    coverage: entry.coverage,
    execution_dependency: entry.execution_dependency,
    execution_family: family.family,
    execution_label: family.label,
    packages: [entry.package],
    support_only: false,
    support_selector: null,
    raw_selector: null,
    file: entry.file,
    package: entry.package,
    symbols: goEntrySymbols(entry),
    shard_isolation: entry.shard_isolation === true,
    evidence_layer: entry.evidence_layer,
    fixture_policy: {
      postgres: effectiveGoEntryPostgresFixturePolicy(entry),
    },
    fixture_budget: {
      postgres: effectiveGoEntryPostgresFixtureBudget(entry),
    },
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
    execution_family: family.family,
    execution_label: family.label,
    packages: [entry.package],
    support_only: true,
    support_selector: entry.selection_pattern,
    raw_selector: null,
    file: entry.file,
    package: entry.package,
    symbols: [symbol],
    shard_isolation: entry.shard_isolation === true,
    evidence_layer: "support",
    fixture_policy: {
      postgres: effectiveSupportGoEntryPostgresFixturePolicy(entry),
    },
    fixture_budget: {
      postgres: effectiveSupportGoEntryPostgresFixtureBudget(entry),
    },
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
    execution_family: aggregate.executionFamily,
    execution_label: aggregate.executionLabel,
    packages: [...aggregate.packages],
    support_only: false,
    support_selector: null,
    raw_selector: aggregate.selectionPattern,
    file: "",
    package: "",
    symbols: [],
    shard_isolation: false,
    evidence_layer: "raw",
    label: aggregate.executionLabel,
    fixture_policy: aggregate.fixturePolicy,
    fixture_budget: aggregate.fixtureBudget,
  };
}

export function collectTargetPlanRows(root = process.cwd()) {
  const config = loadExecutionTargets(root);
  const rows = [];
  for (const phase of phaseManifestNames(root)) {
    const { manifest } = loadManifest(root, phase);
    for (const entry of collectEntries(manifest)) {
      if (entry.runner !== "go_test") {
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
  return rows.sort(compareRows);
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

function rowPackages(row) {
  if (row.package) {
    return [row.package];
  }
  return [...(row.packages ?? [])];
}

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`);
}

function exactRegex(values) {
  if (values.length === 0) {
    throw new Error("cannot build an exact regex from an empty value list");
  }
  const escaped = values.map(escapeRegex);
  if (escaped.length === 1) {
    return `^${escaped[0]}$`;
  }
  return `^(${escaped.join("|")})$`;
}

function buildUnionRegex(components) {
  const values = components.filter((component) => component !== "");
  if (values.length === 0) {
    throw new Error("cannot build aggregate regex from an empty selection");
  }
  if (values.length === 1) {
    return values[0];
  }
  return values.map((component) => `(${component})`).join("|");
}

function aggregateRegex(rows) {
  const symbols = rows.flatMap((row) => row.symbols ?? []);
  const components = [];
  if (symbols.length > 0) {
    components.push(exactRegex(symbols.sort(compareStrings)));
  }
  for (const row of rows) {
    if (row.raw_selector) {
      components.push(row.raw_selector);
    }
  }
  return buildUnionRegex(components);
}

function aggregatePackages(rows) {
  return Array.from(new Set(rows.flatMap(rowPackages))).sort(compareStrings);
}

function fixturePolicyAssignments(rows, mode) {
  const assignments = [];
  for (const row of rows) {
    const policy = row.fixture_policy?.postgres ?? "";
    if (!postgresFixturePolicyEnvAssignable.has(policy)) {
      continue;
    }
    if (mode === "tests" && row.coverage !== "raw") {
      for (const symbol of row.symbols ?? []) {
        assignments.push(`${symbol}=${policy}`);
      }
    }
    if (mode === "packages" && row.coverage === "raw") {
      for (const pkg of row.packages ?? []) {
        assignments.push(`${pkg}=${policy}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

function resetTableAssignments(rows, mode) {
  const assignments = [];
  for (const row of rows) {
    const dirtyTables = row.fixture_budget?.postgres?.dirty_tables ?? [];
    if (dirtyTables.length === 0) {
      continue;
    }
    if (mode === "tests" && row.coverage !== "raw") {
      for (const symbol of row.symbols ?? []) {
        assignments.push(`${symbol}=${dirtyTables.join("|")}`);
      }
    }
    if (mode === "packages" && row.coverage === "raw") {
      for (const pkg of row.packages ?? []) {
        assignments.push(`${pkg}=${dirtyTables.join("|")}`);
      }
    }
  }
  return assignments.sort(compareStrings);
}

function aggregateKey(row) {
  if (row.coverage === "raw") {
    return `raw:${row.id}`;
  }
  if (row.support_only) {
    return [
      "support",
      row.manifest_phase,
      row.execution_dependency,
      row.execution_family,
      row.execution_label,
    ].join("\u001f");
  }
  return [
    "manifest",
    row.manifest_phase,
    row.section,
    row.coverage,
    row.execution_dependency,
    row.execution_family,
    row.execution_label,
  ].join("\u001f");
}

function collectAggregateEmissions(rows) {
  const groups = new Map();
  for (const row of rows) {
    const key = aggregateKey(row);
    if (!groups.has(key)) {
      groups.set(key, {
        mode: row.coverage === "raw" ? "raw" : row.support_only ? "support" : "manifest",
        label: row.execution_label,
        phase: row.manifest_phase,
        section: row.section,
        coverage: row.coverage,
        execution_dependency: row.execution_dependency,
        execution_family: row.execution_family,
        support_target: row.support_only ? row.execution_dependency : "",
        regex: row.raw_selector ?? "",
        packages: new Set(),
        symbols: [],
      });
    }
    const group = groups.get(key);
    for (const pkg of rowPackages(row)) {
      group.packages.add(pkg);
    }
    if (row.support_only) {
      group.symbols.push(...(row.symbols ?? []));
    }
  }

  return Array.from(groups.values()).map((group) => {
    const symbols = group.symbols.sort(compareStrings);
    return {
      ...group,
      regex: group.mode === "support" ? exactRegex(symbols) : group.regex,
      packages: Array.from(group.packages).sort(compareStrings),
      symbols,
    };
  });
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
