#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import { collectObservedGoShardArtifacts } from "./lib/go-duration-artifacts.mjs";
import {
  readGoDurationBaselineMaps,
  resolveGoDurationBaselineFile,
  validBaselineValue,
  withGoDurationBaselineFile,
} from "./lib/go-duration-baselines.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
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

function contaminationDescription(artifact) {
  if (!artifact.timingContaminationReasons?.includes("go-module-download")) {
    return "";
  }
  return `go_module_downloads=${artifact.moduleDownloadCount}`;
}

function checkShardDrift(errors, warnings, artifact, planned) {
  if (artifact.durationMs > planned * underRatio && artifact.durationMs - planned > underDeltaMs) {
    const contamination = contaminationDescription(artifact);
    if (contamination) {
      warnings.push(
        `ignored underplanned contaminated shard=${artifact.shardName} planned_ms=${planned} actual_ms=${artifact.durationMs} ratio=${formatRatio(artifact.durationMs, planned)} ${contamination}`,
      );
      return;
    }
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
  const baselineFile = resolveGoDurationBaselineFile(repoRoot, options.baselineFile);
  const baseline = readGoDurationBaselineMaps(repoRoot, baselineFile);
  const artifacts = withGoDurationBaselineFile(repoRoot, baselineFile, () =>
    collectObservedGoShardArtifacts(repoRoot, options.resultsDir),
  );

  const errors = [];
  const warnings = [];
  for (const artifact of artifacts) {
    if (artifact.observedRawPackages?.length > 0) {
      let planned = 0;
      const missing = [];
      for (const observedPackage of artifact.observedRawPackages) {
        const rawPackageWeight = baseline.rawAggregates.get(observedPackage.key);
        if (!validBaselineValue(rawPackageWeight)) {
          missing.push(`raw package baseline key=${observedPackage.key}`);
          continue;
        }
        planned += rawPackageWeight;
      }
      for (const key of missing) {
        errors.push(`missing ${key} shard=${artifact.shardName}`);
      }
      if (planned <= 0 || missing.length > 0) {
        continue;
      }
      checkShardDrift(errors, warnings, artifact, planned);
      continue;
    }

    if (artifact.rawAggregateKey) {
      const planned = baseline.rawAggregates.get(artifact.rawAggregateKey);
      if (!validBaselineValue(planned)) {
        errors.push(
          `missing raw aggregate baseline key=${artifact.rawAggregateKey} shard=${artifact.shardName}`,
        );
        continue;
      }
      checkShardDrift(errors, warnings, artifact, planned);
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
    checkShardDrift(errors, warnings, artifact, planned);
  }

  if (warnings.length > 0) {
    process.stderr.write("Go test duration baseline drift ignored contaminated timing evidence:\n");
    for (const warning of warnings) {
      process.stderr.write(`- ${warning}\n`);
    }
    process.stderr.write("Rerun after Go module cache warm-up; do not refresh baselines from contaminated timing evidence.\n");
  }

  if (errors.length > 0) {
    process.stderr.write("Go test duration baseline drift detected:\n");
    for (const error of errors) {
      process.stderr.write(`- ${error}\n`);
    }
    process.stderr.write(`Refresh from an uncontaminated successful run with: ${suggestedRefresh(options.resultsDir)}\n`);
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
