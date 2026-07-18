import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadExecutionTopology } from "../generated-artifacts/execution-topology.mjs";
import { loadTestCatalog, targetForCatalogRow } from "../test-catalog/index.mjs";
import {
  aggregatePackages,
  aggregateRegex,
  collectAggregateEmissions,
  fixturePolicyAssignments,
  resetTableAssignments,
} from "./go-target-aggregate.mjs";

const validShardModes = new Set(["none", "go_shards"]);
const validParallelismModes = new Set(["none", "package", "process"]);
const executionTargetsCache = new Map();
const targetPlanRowsCache = new Map();
const fixturePolicyByProfile = Object.freeze({
  none: "",
  object_store_isolated: "",
  postgres_group_clone: "group_clone",
  postgres_migration_scratch: "migration_scratch",
  postgres_package_reset: "package_reset",
  postgres_template_clone: "template_clone",
  postgres_transaction: "transaction",
  service_stack: "",
});

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function compareRows(left, right) {
  return (
    compareStrings(left.target, right.target) ||
    compareStrings(left.execution_family, right.execution_family) ||
    compareStrings(left.id, right.id)
  );
}

function loadExecutionTargets(root) {
  const cacheKey = path.resolve(root);
  const cached = executionTargetsCache.get(cacheKey);
  if (cached) return cached;
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
    runtimeBinariesByFamily: topology.goTargets.runtimeBinariesByFamily,
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

function profileByID(catalog, kind, profileID) {
  const profile = catalog.profiles.semantic[kind].find((entry) => entry.id === profileID);
  if (!profile) throw new Error(`unresolved ${kind} profile ${profileID}`);
  return profile;
}

function catalogRow(config, catalog, row) {
  const target = targetForCatalogRow(row);
  const descriptor = config.byName.get(target);
  if (!descriptor) {
    throw new Error(`catalog Go row ${row.row_id} resolves to unknown target ${target}`);
  }
  const runtimeProfile = profileByID(catalog, "runtime_profiles", row.runtime_profile_id);
  if (descriptor.serviceBacked !== (runtimeProfile.managed_service_ids.length > 0)) {
    throw new Error(
      `catalog Go row ${row.row_id} runtime profile ${row.runtime_profile_id} is incompatible with ${target}`,
    );
  }
  const fixtureProfile = profileByID(catalog, "fixture_profiles", row.fixture_profile_id);
  const resourceProfile = profileByID(catalog, "resource_profiles", row.resource_profile_id);
  const supportOnly = target === "backend-integration-support";
  const fixturePolicy = fixturePolicyByProfile[row.fixture_profile_id];
  if (fixturePolicy === undefined) {
    throw new Error(`catalog Go row ${row.row_id} has unsupported fixture profile ${row.fixture_profile_id}`);
  }
  return {
    ...rowBase(descriptor),
    id: row.row_id,
    owner_id: row.owner_id,
    family_id: row.family_id,
    manifest_phase: row.owner_id,
    section: target === "backend-unit" ? "unit" : "integration",
    coverage: supportOnly ? "support" : "authoritative",
    execution_dependency: target.replaceAll("-", "_"),
    evidence_class: row.evidence_class,
    layer: target,
    default_check_required: row.default_check,
    default_check_kind: row.default_check ? "catalog_default" : "explicit_only",
    default_check_reason_code: row.default_check ? "catalog_selected" : "catalog_explicit_only",
    primary_evidence_owner: row.row_id,
    duplicate_of: null,
    evidence_delta: "Catalog-owned exact selector evidence.",
    warm_local_cost_class: descriptor.serviceBacked ? "service_backed" : "low",
    execution_family: row.family_id,
    execution_label: row.family_id,
    packages: [row.selector.package],
    support_only: supportOnly,
    support_selector: null,
    raw_selector: null,
    file: "",
    package: row.selector.package,
    symbols: [...row.selector.tests],
    runtime_profile_id: row.runtime_profile_id,
    resource_profile_id: row.resource_profile_id,
    resource_claims: { ...resourceProfile.resource_claims },
    fixture_profile_id: row.fixture_profile_id,
    runtime_binaries: [...(config.runtimeBinariesByFamily.get(row.family_id) ?? [])],
    shard_isolation: false,
    evidence_layer: target,
    fixture_policy: { postgres: fixturePolicy },
    fixture_budget: {
      postgres: fixtureProfile.fixture_kind === "postgres" ? { ...fixtureProfile.budget } : {},
    },
  };
}

function rawRow(config, catalog, aggregate) {
  const descriptor = config.byName.get(aggregate.target);
  const fixtureProfileID = aggregate.fixturePolicy?.postgres === "template_clone"
    ? "postgres_template_clone"
    : "none";
  const resourceProfileID = fixtureProfileID === "postgres_template_clone"
    ? "go_clone_heavy"
    : "go_balanced";
  const resourceProfile = profileByID(catalog, "resource_profiles", resourceProfileID);
  return {
    ...rowBase(descriptor),
    id: `RAW-${aggregate.id}`,
    owner_id: "harness.backend",
    family_id: aggregate.executionFamily,
    manifest_phase: "",
    section: aggregate.section,
    coverage: "raw",
    execution_dependency: "",
    evidence_class: "diagnostic",
    layer: "raw",
    default_check_required: false,
    default_check_kind: "explicit_only",
    default_check_reason_code: "explicit_measurement",
    primary_evidence_owner: "raw_diagnostic",
    duplicate_of: null,
    evidence_delta: "Raw aggregate diagnostic coverage is never catalog row evidence.",
    warm_local_cost_class: "explicit_heavy",
    execution_family: aggregate.executionFamily,
    execution_label: aggregate.executionLabel,
    packages: [...aggregate.packages],
    support_only: false,
    support_selector: null,
    raw_selector: aggregate.selectionPattern,
    file: "",
    package: "",
    symbols: [],
    runtime_profile_id: descriptor.serviceBacked ? "default" : "none",
    resource_profile_id: resourceProfileID,
    resource_claims: { ...resourceProfile.resource_claims },
    fixture_profile_id: fixtureProfileID,
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
  if (cached) return cached;
  const config = loadExecutionTargets(root);
  const catalog = loadTestCatalog(root);
  const rows = catalog.rows
    .filter((row) => row.runner === "go")
    .map((row) => catalogRow(config, catalog, row));
  rows.push(...config.rawAggregates.map((aggregate) => rawRow(config, catalog, aggregate)));
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
  return [...new Set(
    collectTargetPlanRows(root)
      .filter((row) => row.target === target)
      .map((row) => row.execution_family),
  )].sort(compareStrings);
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
      if (!descriptor) throw new Error(`unknown target ${target}`);
      process.stdout.write(`${String(descriptor[family] ?? "")}\n`);
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
      const emission = collectAggregateEmissions(rowsForAggregate(root, target, family))[Number.parseInt(extra, 10)];
      if (!emission) throw new Error(`unknown emission index ${extra} for ${target} ${family}`);
      process.stdout.write(`${String(emission[argv[4]] ?? "")}\n`);
      return;
    }
    case "aggregate-emission-packages": {
      const emission = collectAggregateEmissions(rowsForAggregate(root, target, family))[Number.parseInt(extra, 10)];
      if (!emission) throw new Error(`unknown emission index ${extra} for ${target} ${family}`);
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
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    process.exit(1);
  }
}
