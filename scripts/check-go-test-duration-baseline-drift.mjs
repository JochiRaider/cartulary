#!/usr/bin/env node
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { collectObservedGoShardArtifacts } from "./lib/go-duration-artifacts.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultBaselineFile = path.join(repoRoot, "tools", "go_test_duration_baselines.json");
const baselineFileEnv = "CARTULARY_GO_TEST_DURATION_BASELINE_FILE";
const underRatio = 1.75;
const underDeltaMs = 5000;
const overRatio = 3;
const overDeltaMs = 15000;

function usage() {
  process.stderr.write(
    "usage: check-go-test-duration-baseline-drift.mjs [--baseline-file <path>] <results-dir>\n",
  );
  process.exit(2);
}

function resolveBaselineFile(file) {
  const configured = file || process.env[baselineFileEnv] || defaultBaselineFile;
  return path.isAbsolute(configured) ? configured : path.join(repoRoot, configured);
}

function readBaseline(file) {
  if (!existsSync(file)) {
    throw new Error(`baseline file does not exist: ${path.relative(repoRoot, file)}`);
  }
  const baseline = JSON.parse(readFileSync(file, "utf8"));
  if (baseline.schema_id !== "cartulary.go_test_duration_baselines.v3") {
    throw new Error(`${path.relative(repoRoot, file)} must declare schema_id cartulary.go_test_duration_baselines.v3`);
  }
  return {
    rawAggregates: new Map(Object.entries(baseline.raw_aggregates ?? {})),
    tests: new Map(Object.entries(baseline.tests ?? {})),
  };
}

function parseArgs(argv) {
  const options = {
    baselineFile: "",
    resultsDir: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
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

function formatRatio(actual, planned) {
  if (planned <= 0) {
    return "inf";
  }
  return (actual / planned).toFixed(2);
}

function suggestedRefresh(resultsDir) {
  return `make go-test-duration-baselines RESULTS_DIR=${resultsDir} PRUNE_OBSERVED_PACKAGES=1`;
}

function main(argv) {
  const options = parseArgs(argv);
  const baselineFile = resolveBaselineFile(options.baselineFile);
  const baseline = readBaseline(baselineFile);
  const previousBaselineOverride = process.env[baselineFileEnv];
  process.env[baselineFileEnv] = baselineFile;
  const artifacts = collectObservedGoShardArtifacts(repoRoot, options.resultsDir);
  if (previousBaselineOverride === undefined) {
    delete process.env[baselineFileEnv];
  } else {
    process.env[baselineFileEnv] = previousBaselineOverride;
  }

  const errors = [];
  for (const artifact of artifacts) {
    if (artifact.rawAggregateKey) {
      const planned = baseline.rawAggregates.get(artifact.rawAggregateKey);
      if (!Number.isInteger(planned) || planned <= 0) {
        errors.push(
          `missing raw aggregate baseline key=${artifact.rawAggregateKey} shard=${artifact.shardName}`,
        );
        continue;
      }
      if (artifact.durationMs > planned * underRatio && artifact.durationMs - planned > underDeltaMs) {
        errors.push(
          `underplanned shard=${artifact.shardName} planned_ms=${planned} actual_ms=${artifact.durationMs} ratio=${formatRatio(artifact.durationMs, planned)}`,
        );
      }
      if (planned > artifact.durationMs * overRatio && planned - artifact.durationMs > overDeltaMs) {
        errors.push(
          `overplanned shard=${artifact.shardName} planned_ms=${planned} actual_ms=${artifact.durationMs} ratio=${formatRatio(artifact.durationMs, planned)}`,
        );
      }
      continue;
    }

    let planned = 0;
    const missing = [];
    for (const observedTest of artifact.observedTests) {
      const testWeight = baseline.tests.get(observedTest.key);
      if (!Number.isInteger(testWeight) || testWeight <= 0) {
        missing.push(observedTest.key);
        continue;
      }
      planned += testWeight;
    }
    for (const key of missing) {
      errors.push(`missing test baseline key=${key} shard=${artifact.shardName}`);
    }
    if (planned <= 0 || missing.length > 0) {
      continue;
    }
    if (artifact.durationMs > planned * underRatio && artifact.durationMs - planned > underDeltaMs) {
      errors.push(
        `underplanned shard=${artifact.shardName} planned_ms=${planned} actual_ms=${artifact.durationMs} ratio=${formatRatio(artifact.durationMs, planned)}`,
      );
    }
    if (planned > artifact.durationMs * overRatio && planned - artifact.durationMs > overDeltaMs) {
      errors.push(
        `overplanned shard=${artifact.shardName} planned_ms=${planned} actual_ms=${artifact.durationMs} ratio=${formatRatio(artifact.durationMs, planned)}`,
      );
    }
  }

  if (errors.length > 0) {
    process.stderr.write("Go test duration baseline drift detected:\n");
    for (const error of errors) {
      process.stderr.write(`- ${error}\n`);
    }
    process.stderr.write(`Refresh with: ${suggestedRefresh(options.resultsDir)}\n`);
    process.exit(1);
  }

  process.stdout.write(`Go test duration baselines match ${artifacts.length} observed service-backed shards\n`);
}

try {
  main(process.argv.slice(2));
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
