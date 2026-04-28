#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  packageOverheadBaselineKey,
  readGoDurationBaselineMaps,
  resolveGoDurationBaselineFile,
  validBaselineValue,
  withGoDurationBaselineFile,
} from "./lib/go-duration-baselines.mjs";
import { collectGoShardPlan } from "./lib/go-shard-plan.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");

function usage() {
  process.stderr.write("usage: check-go-test-duration-baseline-coverage.mjs [--baseline-file <path>]\n");
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    baselineFile: "",
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
    usage();
  }
  return options;
}

function addMissing(missing, label) {
  missing.add(label);
}

function checkBaselineCoverage(plan, baseline) {
  const missing = new Set();
  const nonRawTargets = new Set();
  const nonRawPackageKeys = new Set();

  for (const shard of plan.shards) {
    for (const item of shard.items ?? []) {
      if (item.kind === "raw") {
        if (!validBaselineValue(baseline.rawAggregates.get(item.baseline_key))) {
          addMissing(
            missing,
            `raw aggregate baseline key=${item.baseline_key} shard=${shard.name}`,
          );
        }
        continue;
      }
      if (!validBaselineValue(baseline.tests.get(item.baseline_key))) {
        addMissing(missing, `test baseline key=${item.baseline_key} shard=${shard.name}`);
      }
      nonRawTargets.add(item.target);
      for (const importPath of item.package_import_paths ?? []) {
        nonRawPackageKeys.add(packageOverheadBaselineKey(item.target, importPath));
      }
    }
  }

  for (const key of nonRawPackageKeys) {
    if (!validBaselineValue(baseline.packageOverheads.get(key))) {
      addMissing(missing, `package overhead baseline key=${key}`);
    }
  }
  for (const target of nonRawTargets) {
    if (!validBaselineValue(baseline.commandOverheadsByTarget.get(target))) {
      addMissing(missing, `command overhead baseline target=${target}`);
    }
  }

  return [...missing].sort();
}

function main(argv) {
  const options = parseArgs(argv);
  const baselineFile = resolveGoDurationBaselineFile(repoRoot, options.baselineFile);
  const baseline = readGoDurationBaselineMaps(repoRoot, baselineFile);
  const plan = withGoDurationBaselineFile(repoRoot, baselineFile, () => collectGoShardPlan(repoRoot));
  const missing = checkBaselineCoverage(plan, baseline);

  if (missing.length > 0) {
    process.stderr.write("Go test duration baseline coverage is incomplete:\n");
    for (const error of missing) {
      process.stderr.write(`- missing ${error}\n`);
    }
    process.stderr.write(
      "Refresh from a successful service-backed run with: make go-test-duration-baselines RESULTS_DIR=<dir> PRUNE_OBSERVED_PACKAGES=1\n",
    );
    process.exit(1);
  }

  process.stdout.write(`Go test duration baselines cover ${plan.shards.length} service-backed Go shards\n`);
}

try {
  main(process.argv.slice(2));
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
