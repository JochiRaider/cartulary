#!/usr/bin/env node
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { collectObservedGoShardArtifacts } from "./lib/go-duration-artifacts.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const defaultBaselineFile = path.join(repoRoot, "tools", "go_test_duration_baselines.json");
const baselineFileEnv = "CARTULARY_GO_TEST_DURATION_BASELINE_FILE";
const schemaID = "cartulary.go_test_duration_baselines.v4";
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
  if (baseline.schema_id !== schemaID) {
    throw new Error(`${path.relative(repoRoot, file)} must declare schema_id ${schemaID}`);
  }
  return {
    commandOverheadsByTarget: new Map(Object.entries(baseline.command_overheads_by_target ?? {})),
    packageOverheads: new Map(Object.entries(baseline.package_overheads ?? {})),
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

function validBaselineValue(value) {
  return Number.isInteger(value) && value > 0;
}

function checkShardDrift(errors, artifact, planned) {
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
      if (!validBaselineValue(planned)) {
        errors.push(
          `missing raw aggregate baseline key=${artifact.rawAggregateKey} shard=${artifact.shardName}`,
        );
        continue;
      }
      checkShardDrift(errors, artifact, planned);
      continue;
    }

    let planned = 0;
    const missing = [];
    for (const observedTest of artifact.observedTests) {
      const testWeight = baseline.tests.get(observedTest.key);
      if (!validBaselineValue(testWeight)) {
        missing.push(`test baseline key=${observedTest.key}`);
        continue;
      }
      planned += testWeight;
    }
    for (const observedPackage of artifact.observedPackageOverheads) {
      const packageOverhead = baseline.packageOverheads.get(observedPackage.key);
      if (!validBaselineValue(packageOverhead)) {
        missing.push(`package overhead baseline key=${observedPackage.key}`);
        continue;
      }
      planned += packageOverhead;
    }
    const commandOverhead = baseline.commandOverheadsByTarget.get(artifact.commandOverhead.target);
    if (!validBaselineValue(commandOverhead)) {
      missing.push(`command overhead baseline target=${artifact.commandOverhead.target}`);
    } else {
      planned += commandOverhead;
    }
    for (const key of missing) {
      errors.push(`missing ${key} shard=${artifact.shardName}`);
    }
    if (planned <= 0 || missing.length > 0) {
      continue;
    }
    checkShardDrift(errors, artifact, planned);
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
