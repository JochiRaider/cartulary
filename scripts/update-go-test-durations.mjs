#!/usr/bin/env node
import { existsSync, readdirSync, readFileSync, statSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  collectServiceTimingContamination,
  printContaminationReasons,
} from "../tools/harness/scheduler/duration-drift.mjs";
import {
  collectObservedGoShardArtifacts,
  sortedObject,
} from "../tools/harness/backend/go-duration-artifacts.mjs";
import {
  baselineNote,
  defaultItemWeightMs,
  defaultShardTargetMs,
  defaultShardTargetMsByTarget,
  goDurationBaselineSchemaID,
  readGoDurationBaseline,
  resolveGoDurationBaselineFile,
  withGoDurationBaselineFile,
} from "../tools/harness/backend/go-duration-baselines.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const suspiciousCommandOverheadDecreaseRatio = 0.75;
const suspiciousCommandOverheadDecreaseDeltaMs = 500;
const fullServiceBackedTargets = new Set(["test-service-backed", "check-service-backed"]);

function usage() {
  process.stderr.write(
    "usage: update-go-test-durations.mjs [--prune-observed-packages] [--allow-command-overhead-decrease] [--baseline-file <path>] <successful-results-dir>\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    allowCommandOverheadDecrease: false,
    baselineFile: "",
    pruneObservedPackages: false,
    resultsDir: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--prune-observed-packages") {
      options.pruneObservedPackages = true;
      continue;
    }
    if (arg === "--allow-command-overhead-decrease") {
      options.allowCommandOverheadDecrease = true;
      continue;
    }
    if (arg === "--baseline-file") {
      options.baselineFile = argv[index + 1] ?? "";
      index += 1;
      if (!options.baselineFile) {
        usage();
      }
      continue;
    }
    if (!options.resultsDir) {
      options.resultsDir = arg;
      continue;
    }
    usage();
  }
  if (!options.resultsDir) {
    usage();
  }
  return options;
}

function isPositiveInteger(value) {
  return Number.isInteger(value) && value > 0;
}

function isSuspiciousCommandOverheadDecrease(existingMs, observedMs) {
  return (
    isPositiveInteger(existingMs) &&
    isPositiveInteger(observedMs) &&
    observedMs < existingMs * suspiciousCommandOverheadDecreaseRatio &&
    existingMs - observedMs > suspiciousCommandOverheadDecreaseDeltaMs
  );
}

function walkFiles(root) {
  const files = [];
  const stack = [root];
  while (stack.length > 0) {
    const current = stack.pop();
    let entries = [];
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const entry of entries) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (entry.isFile()) {
        files.push(next);
      }
    }
  }
  return files.sort();
}

function readJSONIfPossible(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function hasPassingFullServiceBackedSchedulerSummary(resultsDir) {
  const absoluteResultsDir = path.resolve(resultsDir);
  if (!existsSync(absoluteResultsDir) || !statSync(absoluteResultsDir).isDirectory()) {
    return false;
  }
  for (const file of walkFiles(absoluteResultsDir)) {
    if (path.basename(file) !== "scheduler-summary.json") {
      continue;
    }
    const summary = readJSONIfPossible(file);
    if (
      summary &&
      fullServiceBackedTargets.has(summary.target) &&
      summary.status === "pass" &&
      (summary.scheduler_kind === "service_backed" || summary.scheduler_kind === "service-backed")
    ) {
      return true;
    }
  }
  return false;
}

function assertPruneInputIsFullServiceBackedRun(options) {
  if (!options.pruneObservedPackages) {
    return;
  }
  if (hasPassingFullServiceBackedSchedulerSummary(options.resultsDir)) {
    return;
  }
  process.stderr.write(
    "Refusing to prune Go test duration baselines from partial service-backed timing evidence.\n",
  );
  process.stderr.write(
    "Prune mode requires a successful full service-backed scheduler summary for test-service-backed or check-service-backed.\n",
  );
  process.stderr.write("Run full service-backed evidence with: make test-service-backed\n");
  process.stderr.write(
    `Then refresh with: make go-test-duration-baselines RESULTS_DIR=${options.resultsDir} PRUNE_OBSERVED_PACKAGES=1\n`,
  );
  process.exit(1);
}

function main(argv) {
  const options = parseArgs(argv);
  const baselineFile = resolveGoDurationBaselineFile(repoRoot, options.baselineFile);
  const serviceContamination = collectServiceTimingContamination(repoRoot, options.resultsDir);
  if (serviceContamination.contaminated) {
    process.stderr.write("Refusing to refresh Go test duration baselines from contaminated service timing evidence:\n");
    printContaminationReasons(process.stderr, serviceContamination);
    process.stderr.write(`Inspect fixture costs with: make fixture-report RESULTS_DIR=${options.resultsDir}\n`);
    process.stderr.write("Rerun check-shaped evidence with: make check\n");
    process.exit(1);
  }
  assertPruneInputIsFullServiceBackedRun(options);
  const artifacts = withGoDurationBaselineFile(repoRoot, baselineFile, () =>
    collectObservedGoShardArtifacts(repoRoot, options.resultsDir),
  );

  const { baseline } = readGoDurationBaseline(repoRoot, baselineFile, { allowMissing: true });
  const testDurations = new Map();
  const packageOverheads = new Map();
  const fixtureOverheadsByPackage = new Map();
  const fixtureOverheadsByTest = new Map();
  const commandOverheads = new Map();
  const rawAggregateDurations = new Map();
  const observedRawAggregatePrefixes = new Set();
  const observedPackages = new Set();
  const observedPackagesByTarget = new Map();
  const skippedContaminated = [];
  const keptCommandOverheadDecreases = [];

  for (const artifact of artifacts) {
    if (artifact.timingContaminationReasons?.length > 0) {
      skippedContaminated.push(artifact);
      continue;
    }
    if (artifact.observedRawPackages?.length > 0) {
      for (const observedPackage of artifact.observedRawPackages) {
        rawAggregateDurations.set(
          observedPackage.key,
          Math.max(rawAggregateDurations.get(observedPackage.key) ?? 0, observedPackage.durationMs),
        );
        observedRawAggregatePrefixes.add(
          `${artifact.target}::${artifact.aggregateName}::`,
        );
      }
      continue;
    }
    if (artifact.rawAggregateKey) {
      rawAggregateDurations.set(
        artifact.rawAggregateKey,
        Math.max(rawAggregateDurations.get(artifact.rawAggregateKey) ?? 0, artifact.durationMs),
      );
      continue;
    }
    commandOverheads.set(
      artifact.commandOverhead.target,
      Math.max(commandOverheads.get(artifact.commandOverhead.target) ?? 0, artifact.commandOverhead.overheadMs),
    );
    for (const packageName of artifact.observedPackages) {
      observedPackages.add(packageName);
      if (!observedPackagesByTarget.has(artifact.target)) {
        observedPackagesByTarget.set(artifact.target, new Set());
      }
      observedPackagesByTarget.get(artifact.target).add(packageName);
    }
    for (const observedPackage of artifact.observedPackageOverheads) {
      packageOverheads.set(
        observedPackage.key,
        Math.max(packageOverheads.get(observedPackage.key) ?? 0, observedPackage.overheadMs),
      );
    }
    for (const observedPackage of artifact.observedFixturePackages ?? []) {
      fixtureOverheadsByPackage.set(
        observedPackage.key,
        Math.max(fixtureOverheadsByPackage.get(observedPackage.key) ?? 0, observedPackage.fixtureMs),
      );
    }
    for (const observedTest of artifact.observedTests) {
      testDurations.set(
        observedTest.key,
        Math.max(testDurations.get(observedTest.key) ?? 0, observedTest.elapsedMs),
      );
    }
    for (const observedTest of artifact.observedFixtureTests ?? []) {
      fixtureOverheadsByTest.set(
        observedTest.key,
        Math.max(fixtureOverheadsByTest.get(observedTest.key) ?? 0, observedTest.fixtureMs),
      );
    }
  }

  baseline.schema_id = goDurationBaselineSchemaID;
  baseline.note = baselineNote;
  baseline.default_shard_target_ms ??= defaultShardTargetMs;
  baseline.shard_target_ms_by_target = {
    ...defaultShardTargetMsByTarget,
    ...(baseline.shard_target_ms_by_target ?? {}),
  };
  baseline.default_item_weight_ms ??= defaultItemWeightMs;
  baseline.command_overheads_by_target ??= {};
  baseline.package_overheads ??= {};
  baseline.fixture_overheads_by_package ??= {};
  baseline.fixture_overheads_by_test ??= {};
  baseline.raw_aggregates ??= {};
  baseline.tests ??= {};

  if (options.pruneObservedPackages) {
    for (const key of Object.keys(baseline.tests)) {
      const [packageName] = key.split("::", 1);
      if (observedPackages.has(packageName) && !testDurations.has(key)) {
        delete baseline.tests[key];
      }
    }
    for (const key of Object.keys(baseline.package_overheads)) {
      const [target, packageName] = key.split("::");
      const targetPackages = observedPackagesByTarget.get(target);
      if (targetPackages && !targetPackages.has(packageName)) {
        delete baseline.package_overheads[key];
      }
    }
    for (const key of Object.keys(baseline.fixture_overheads_by_package)) {
      const [target, packageName] = key.split("::");
      const targetPackages = observedPackagesByTarget.get(target);
      if (targetPackages && !targetPackages.has(packageName)) {
        delete baseline.fixture_overheads_by_package[key];
      }
    }
    for (const key of Object.keys(baseline.fixture_overheads_by_test)) {
      const [packageName] = key.split("::", 1);
      if (observedPackages.has(packageName) && !fixtureOverheadsByTest.has(key)) {
        delete baseline.fixture_overheads_by_test[key];
      }
    }
    for (const key of Object.keys(baseline.raw_aggregates)) {
      for (const prefix of observedRawAggregatePrefixes) {
        const legacyKey = prefix.slice(0, -2);
        if (key === legacyKey || (key.startsWith(prefix) && !rawAggregateDurations.has(key))) {
          delete baseline.raw_aggregates[key];
        }
      }
    }
  }

  for (const [key, durationMs] of testDurations) {
    baseline.tests[key] = durationMs;
  }
  for (const [target, overheadMs] of commandOverheads) {
    const existingMs = baseline.command_overheads_by_target[target];
    if (
      !options.allowCommandOverheadDecrease &&
      isSuspiciousCommandOverheadDecrease(existingMs, overheadMs)
    ) {
      keptCommandOverheadDecreases.push({ target, existingMs, observedMs: overheadMs });
      continue;
    }
    baseline.command_overheads_by_target[target] = overheadMs;
  }
  for (const [key, overheadMs] of packageOverheads) {
    baseline.package_overheads[key] = overheadMs;
  }
  for (const [key, fixtureMs] of fixtureOverheadsByPackage) {
    baseline.fixture_overheads_by_package[key] = fixtureMs;
  }
  for (const [key, fixtureMs] of fixtureOverheadsByTest) {
    baseline.fixture_overheads_by_test[key] = fixtureMs;
  }
  for (const [key, durationMs] of rawAggregateDurations) {
    baseline.raw_aggregates[key] = durationMs;
  }

  baseline.updated_at = new Date().toISOString();
  baseline.shard_target_ms_by_target = sortedObject(baseline.shard_target_ms_by_target);
  baseline.command_overheads_by_target = sortedObject(baseline.command_overheads_by_target);
  baseline.package_overheads = sortedObject(baseline.package_overheads);
  baseline.fixture_overheads_by_package = sortedObject(baseline.fixture_overheads_by_package);
  baseline.fixture_overheads_by_test = sortedObject(baseline.fixture_overheads_by_test);
  baseline.raw_aggregates = sortedObject(baseline.raw_aggregates);
  baseline.tests = sortedObject(baseline.tests);

  writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  if (skippedContaminated.length > 0) {
    process.stderr.write("skipped contaminated Go shard timing artifacts:\n");
    for (const artifact of skippedContaminated) {
      process.stderr.write(
        `- shard=${artifact.shardName} go_module_downloads=${artifact.moduleDownloadCount}\n`,
      );
    }
  }
  if (keptCommandOverheadDecreases.length > 0) {
    process.stderr.write("kept existing Go command overhead baselines after suspicious decreases:\n");
    for (const decrease of keptCommandOverheadDecreases) {
      process.stderr.write(
        `- target=${decrease.target} existing_ms=${decrease.existingMs} observed_ms=${decrease.observedMs} override_with=ALLOW_COMMAND_OVERHEAD_DECREASE=1\n`,
      );
    }
  }
  process.stdout.write(
    `updated ${testDurations.size} Go test baselines, ${packageOverheads.size} package overhead baselines, ${commandOverheads.size} command overhead baselines, and ${rawAggregateDurations.size} raw aggregate baselines`,
  );
  if (skippedContaminated.length > 0) {
    process.stdout.write(`; skipped ${skippedContaminated.length} contaminated shard artifacts`);
  }
  process.stdout.write("\n");
}

try {
  main(process.argv.slice(2));
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
