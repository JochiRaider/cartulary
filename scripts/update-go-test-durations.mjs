#!/usr/bin/env node
import { writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  collectObservedGoShardArtifacts,
  sortedObject,
} from "./lib/go-duration-artifacts.mjs";
import {
  baselineNote,
  defaultItemWeightMs,
  defaultShardTargetMs,
  defaultShardTargetMsByTarget,
  goDurationBaselineSchemaID,
  readGoDurationBaseline,
  resolveGoDurationBaselineFile,
  withGoDurationBaselineFile,
} from "./lib/go-duration-baselines.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");

function usage() {
  process.stderr.write(
    "usage: update-go-test-durations.mjs [--prune-observed-packages] [--baseline-file <path>] <results-dir>\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
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

function main(argv) {
  const options = parseArgs(argv);
  const baselineFile = resolveGoDurationBaselineFile(repoRoot, options.baselineFile);
  const artifacts = withGoDurationBaselineFile(repoRoot, baselineFile, () =>
    collectObservedGoShardArtifacts(repoRoot, options.resultsDir),
  );

  const { baseline } = readGoDurationBaseline(repoRoot, baselineFile, { allowMissing: true });
  const testDurations = new Map();
  const packageOverheads = new Map();
  const commandOverheads = new Map();
  const rawAggregateDurations = new Map();
  const observedPackages = new Set();
  const observedPackagesByTarget = new Map();

  for (const artifact of artifacts) {
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
    for (const observedTest of artifact.observedTests) {
      testDurations.set(
        observedTest.key,
        Math.max(testDurations.get(observedTest.key) ?? 0, observedTest.elapsedMs),
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
  }

  for (const [key, durationMs] of testDurations) {
    baseline.tests[key] = durationMs;
  }
  for (const [target, overheadMs] of commandOverheads) {
    baseline.command_overheads_by_target[target] = overheadMs;
  }
  for (const [key, overheadMs] of packageOverheads) {
    baseline.package_overheads[key] = overheadMs;
  }
  for (const [key, durationMs] of rawAggregateDurations) {
    baseline.raw_aggregates[key] = durationMs;
  }

  baseline.updated_at = new Date().toISOString();
  baseline.shard_target_ms_by_target = sortedObject(baseline.shard_target_ms_by_target);
  baseline.command_overheads_by_target = sortedObject(baseline.command_overheads_by_target);
  baseline.package_overheads = sortedObject(baseline.package_overheads);
  baseline.raw_aggregates = sortedObject(baseline.raw_aggregates);
  baseline.tests = sortedObject(baseline.tests);

  writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  process.stdout.write(
    `updated ${testDurations.size} Go test baselines, ${packageOverheads.size} package overhead baselines, ${commandOverheads.size} command overhead baselines, and ${rawAggregateDurations.size} raw aggregate baselines\n`,
  );
}

try {
  main(process.argv.slice(2));
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
