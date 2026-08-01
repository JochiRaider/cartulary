import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

export const goDurationBaselineSchemaID = "cartulary.go_test_duration_baselines.v5";
const goDurationBaselineFileEnv = "CARTULARY_GO_TEST_DURATION_BASELINE_FILE";
const goDurationBaselineRelativePath = path.join("tools", "go_test_duration_baselines.json");
export const defaultShardTargetMs = 30_000;
const defaultBackendIntegrationShardTargetMs = 18_000;
export const defaultItemWeightMs = 10_000;
export const defaultPackageOverheadMs = 100;
export const defaultCommandOverheadMs = 70_000;
export const baselineNote =
  "Advisory backend service-backed shard weights with explicit test, package, command, and raw package timing components. Refresh with make go-test-duration-baselines RESULTS_DIR=<dir> PRUNE_OBSERVED_PACKAGES=1.";

const defaultShardTargetMsByTargetEntries = [
  ["backend-store", defaultShardTargetMs],
  ["backend-integration", defaultBackendIntegrationShardTargetMs],
  ["backend-integration-support", defaultBackendIntegrationShardTargetMs],
];

export const defaultShardTargetMsByTarget = Object.fromEntries(defaultShardTargetMsByTargetEntries);

export function normalizePositiveInteger(value, fallback = 0) {
  if (Number.isInteger(value) && value > 0) {
    return value;
  }
  return fallback;
}

export function validBaselineValue(value) {
  return Number.isInteger(value) && value > 0;
}

export function resolveGoDurationBaselineFile(repoRoot, file = "") {
  const configured = file || process.env[goDurationBaselineFileEnv] || goDurationBaselineRelativePath;
  return path.isAbsolute(configured) ? configured : path.join(repoRoot, configured);
}

function emptyGoDurationBaseline() {
  return {
    schema_id: goDurationBaselineSchemaID,
    note: baselineNote,
    default_shard_target_ms: defaultShardTargetMs,
    shard_target_ms_by_target: { ...defaultShardTargetMsByTarget },
    default_item_weight_ms: defaultItemWeightMs,
    default_package_overhead_ms: defaultPackageOverheadMs,
    default_command_overhead_ms: defaultCommandOverheadMs,
    command_overheads_by_target: {},
    package_overheads: {},
    fixture_overheads_by_package: {},
    fixture_overheads_by_test: {},
    raw_aggregates: {},
    tests: {},
  };
}

export function readGoDurationBaseline(repoRoot, file = "", options = {}) {
  const baselineFile = resolveGoDurationBaselineFile(repoRoot, file);
  if (!existsSync(baselineFile)) {
    if (options.allowMissing) {
      return { baseline: emptyGoDurationBaseline(), baselineFile };
    }
    throw new Error(`baseline file does not exist: ${path.relative(repoRoot, baselineFile)}`);
  }
  const baseline = JSON.parse(readFileSync(baselineFile, "utf8"));
  if (baseline.schema_id !== goDurationBaselineSchemaID) {
    throw new Error(
      `${path.relative(repoRoot, baselineFile)} must declare schema_id ${goDurationBaselineSchemaID}`,
    );
  }
  for (const [field, value] of [
    ["default_item_weight_ms", baseline.default_item_weight_ms],
    ["default_package_overhead_ms", baseline.default_package_overhead_ms],
    ["default_command_overhead_ms", baseline.default_command_overhead_ms],
  ]) {
    if (!validBaselineValue(value)) {
      throw new Error(`${path.relative(repoRoot, baselineFile)} ${field} must be a positive integer`);
    }
  }
  for (const key of Object.keys(baseline.raw_aggregates ?? {})) {
    const parts = key.split("::");
    if (parts.length !== 3 || parts.some((part) => part.length === 0)) {
      throw new Error(
        `${path.relative(repoRoot, baselineFile)} raw_aggregates key ${key} must be target::aggregate::package`,
      );
    }
  }
  return { baseline, baselineFile };
}

function toGoDurationBaselineMaps(baseline) {
  const shardTargetMsByTarget = new Map(defaultShardTargetMsByTargetEntries);
  for (const [target, targetMs] of Object.entries(baseline.shard_target_ms_by_target ?? {})) {
    shardTargetMsByTarget.set(target, normalizePositiveInteger(targetMs, shardTargetMsByTarget.get(target)));
  }
  return {
    defaultShardTargetMs: normalizePositiveInteger(
      baseline.default_shard_target_ms,
      defaultShardTargetMs,
    ),
    shardTargetMsByTarget,
    defaultItemWeightMs: normalizePositiveInteger(baseline.default_item_weight_ms, defaultItemWeightMs),
    defaultPackageOverheadMs: normalizePositiveInteger(
      baseline.default_package_overhead_ms,
      defaultPackageOverheadMs,
    ),
    defaultCommandOverheadMs: normalizePositiveInteger(
      baseline.default_command_overhead_ms,
      defaultCommandOverheadMs,
    ),
    commandOverheadsByTarget: new Map(Object.entries(baseline.command_overheads_by_target ?? {})),
    packageOverheads: new Map(Object.entries(baseline.package_overheads ?? {})),
    fixtureOverheadsByPackage: new Map(Object.entries(baseline.fixture_overheads_by_package ?? {})),
    fixtureOverheadsByTest: new Map(Object.entries(baseline.fixture_overheads_by_test ?? {})),
    rawAggregates: new Map(Object.entries(baseline.raw_aggregates ?? {})),
    tests: new Map(Object.entries(baseline.tests ?? {})),
  };
}

export function readGoDurationBaselineMaps(repoRoot, file = "", options = {}) {
  const { baseline, baselineFile } = readGoDurationBaseline(repoRoot, file, options);
  return { baselineFile, ...toGoDurationBaselineMaps(baseline) };
}

export function withGoDurationBaselineFile(repoRoot, baselineFile, fn) {
  const previousBaselineOverride = process.env[goDurationBaselineFileEnv];
  process.env[goDurationBaselineFileEnv] = path.isAbsolute(baselineFile)
    ? baselineFile
    : path.join(repoRoot, baselineFile);
  try {
    return fn();
  } finally {
    if (previousBaselineOverride === undefined) {
      delete process.env[goDurationBaselineFileEnv];
    } else {
      process.env[goDurationBaselineFileEnv] = previousBaselineOverride;
    }
  }
}

export function testBaselineKey(importPath, symbol) {
  return `${importPath}::${symbol}`;
}

export function rawPackageBaselineKey(target, aggregateName, packageName) {
  return `${target}::${aggregateName}::${packageName}`;
}

export function packageOverheadBaselineKey(target, importPath) {
  return `${target}::${importPath}`;
}
