#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  collectServiceTimingContamination,
  durationDriftDescription,
  durationDriftKind,
  formatContaminationReasons,
  formatRatio,
  formatSignedMs,
  printContaminationReasons,
} from "./lib/duration-drift.mjs";
import { collectObservedGoShardArtifacts } from "./lib/go-duration-artifacts.mjs";
import {
  readGoDurationBaselineMaps,
  resolveGoDurationBaselineFile,
  validBaselineValue,
  withGoDurationBaselineFile,
} from "./lib/go-duration-baselines.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const rawPackageDetailLimit = 5;

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

function suggestedRefresh(resultsDir) {
  return `make go-test-duration-baselines RESULTS_DIR=${resultsDir} PRUNE_OBSERVED_PACKAGES=1`;
}

function contaminationDescription(artifact, serviceContamination) {
  const descriptions = [];
  if (artifact.timingContaminationReasons?.includes("go-module-download")) {
    descriptions.push(`go_module_downloads=${artifact.moduleDownloadCount}`);
  }
  if (serviceContamination.contaminated) {
    descriptions.push(
      `service_timing_contamination=[${formatContaminationReasons(serviceContamination)}]`,
    );
  }
  return descriptions.join(" ");
}

function componentDescription(components) {
  if (!components) {
    return "";
  }
  if (components.rawPackages) {
    return rawPackageComponentDescription(components.rawPackages);
  }
  return [
    `planned_tests_ms=${components.plannedTestsMs}`,
    `actual_tests_ms=${components.actualTestsMs}`,
    `planned_package_overhead_ms=${components.plannedPackageOverheadMs}`,
    `actual_package_overhead_ms=${components.actualPackageOverheadMs}`,
    `planned_command_overhead_ms=${components.plannedCommandOverheadMs}`,
    `actual_command_overhead_ms=${components.actualCommandOverheadMs}`,
  ].join(" ");
}

function rawPackageComponentDescription(components) {
  if (!components || components.length === 0) {
    return "";
  }
  const selected = [...components]
    .sort(
      (left, right) =>
        Math.abs(right.actualMs - right.plannedMs) -
          Math.abs(left.actualMs - left.plannedMs) ||
        right.actualMs - left.actualMs ||
        left.packageName.localeCompare(right.packageName),
    )
    .slice(0, rawPackageDetailLimit);
  const details = selected
    .map((component) =>
      [
        `package=${component.packageName}`,
        `planned_ms=${component.plannedMs}`,
        `actual_ms=${component.actualMs}`,
        `delta_ms=${formatSignedMs(component.actualMs - component.plannedMs)}`,
        `ratio=${formatRatio(component.actualMs, component.plannedMs)}`,
      ].join(" "),
    )
    .join("; ");
  const omitted = components.length - selected.length;
  return `raw_packages=[${details}]${omitted > 0 ? ` raw_package_omitted=${omitted}` : ""}`;
}

function checkShardDrift(driftErrors, warnings, artifact, planned, serviceContamination, components = null) {
  const details = componentDescription(components);
  const kind = durationDriftKind(artifact.durationMs, planned);
  if (!kind) {
    return;
  }
  const description = durationDriftDescription(kind, {
    subject: `shard=${artifact.shardName}`,
    plannedMs: planned,
    actualMs: artifact.durationMs,
    details,
  });
  if (kind === "underplanned") {
    const contamination = contaminationDescription(artifact, serviceContamination);
    if (contamination) {
      warnings.push(`ignored ${description.replace(`${kind} `, `${kind} contaminated `)} ${contamination}`);
      return;
    }
  }
  driftErrors.push(description);
}

function main(argv) {
  const options = parseArgs(argv);
  const baselineFile = resolveGoDurationBaselineFile(repoRoot, options.baselineFile);
  const baseline = readGoDurationBaselineMaps(repoRoot, baselineFile);
  const serviceContamination = collectServiceTimingContamination(repoRoot, options.resultsDir);
  const artifacts = withGoDurationBaselineFile(repoRoot, baselineFile, () =>
    collectObservedGoShardArtifacts(repoRoot, options.resultsDir),
  );

  const missingBaselines = [];
  const driftErrors = [];
  const warnings = [];
  for (const artifact of artifacts) {
    if (artifact.observedRawPackages?.length > 0) {
      let planned = 0;
      const missing = [];
      const rawPackageComponents = [];
      for (const observedPackage of artifact.observedRawPackages) {
        const rawPackageWeight = baseline.rawAggregates.get(observedPackage.key);
        if (!validBaselineValue(rawPackageWeight)) {
          missing.push(`raw package baseline key=${observedPackage.key}`);
          continue;
        }
        planned += rawPackageWeight;
        rawPackageComponents.push({
          packageName: observedPackage.packageName,
          plannedMs: rawPackageWeight,
          actualMs: observedPackage.durationMs,
        });
      }
      for (const key of missing) {
        missingBaselines.push(`missing ${key} shard=${artifact.shardName}`);
      }
      if (planned <= 0 || missing.length > 0) {
        continue;
      }
      checkShardDrift(driftErrors, warnings, artifact, planned, serviceContamination, {
        rawPackages: rawPackageComponents,
      });
      continue;
    }

    if (artifact.rawAggregateKey) {
      const planned = baseline.rawAggregates.get(artifact.rawAggregateKey);
      if (!validBaselineValue(planned)) {
        missingBaselines.push(
          `missing raw aggregate baseline key=${artifact.rawAggregateKey} shard=${artifact.shardName}`,
        );
        continue;
      }
      checkShardDrift(driftErrors, warnings, artifact, planned, serviceContamination);
      continue;
    }

    let planned = 0;
    let plannedTestsMs = 0;
    let plannedPackageOverheadMs = 0;
    let plannedCommandOverheadMs = 0;
    const actualTestsMs = artifact.observedTests.reduce((sum, observedTest) => sum + observedTest.elapsedMs, 0);
    const actualPackageOverheadMs = artifact.observedPackageOverheads.reduce(
      (sum, observedPackage) => sum + observedPackage.overheadMs,
      0,
    );
    const actualCommandOverheadMs = artifact.commandOverhead.overheadMs;
    const missing = [];
    for (const observedTest of artifact.observedTests) {
      const testWeight = baseline.tests.get(observedTest.key);
      if (!validBaselineValue(testWeight)) {
        missing.push(`test baseline key=${observedTest.key}`);
        continue;
      }
      plannedTestsMs += testWeight;
      planned += testWeight;
    }
    for (const observedPackage of artifact.observedPackageOverheads) {
      const packageOverhead = baseline.packageOverheads.get(observedPackage.key);
      if (!validBaselineValue(packageOverhead)) {
        missing.push(`package overhead baseline key=${observedPackage.key}`);
        continue;
      }
      plannedPackageOverheadMs += packageOverhead;
      planned += packageOverhead;
    }
    const commandOverhead = baseline.commandOverheadsByTarget.get(artifact.commandOverhead.target);
    if (!validBaselineValue(commandOverhead)) {
      missing.push(`command overhead baseline target=${artifact.commandOverhead.target}`);
    } else {
      plannedCommandOverheadMs += commandOverhead;
      planned += commandOverhead;
    }
    for (const key of missing) {
      missingBaselines.push(`missing ${key} shard=${artifact.shardName}`);
    }
    if (planned <= 0 || missing.length > 0) {
      continue;
    }
    checkShardDrift(driftErrors, warnings, artifact, planned, serviceContamination, {
      plannedTestsMs,
      actualTestsMs,
      plannedPackageOverheadMs,
      actualPackageOverheadMs,
      plannedCommandOverheadMs,
      actualCommandOverheadMs,
    });
  }

  if (warnings.length > 0) {
    process.stderr.write("Go test duration baseline drift ignored contaminated timing evidence:\n");
    for (const warning of warnings) {
      process.stderr.write(`- ${warning}\n`);
    }
    if (serviceContamination.contaminated) {
      process.stderr.write("Service timing contamination detected:\n");
      printContaminationReasons(process.stderr, serviceContamination);
    }
    process.stderr.write("Rerun after Go module cache warm-up and clean service timing evidence; do not refresh baselines from contaminated timing evidence.\n");
  }

  if (missingBaselines.length > 0 || driftErrors.length > 0) {
    process.stderr.write("Go test duration baseline drift detected:\n");
    if (missingBaselines.length > 0) {
      process.stderr.write("Missing baseline components:\n");
      for (const error of missingBaselines) {
        process.stderr.write(`- ${error}\n`);
      }
    }
    if (driftErrors.length > 0) {
      process.stderr.write("Observed timing drift:\n");
      for (const error of driftErrors) {
        process.stderr.write(`- ${error}\n`);
      }
    }
    process.stderr.write(`Inspect fixture costs with: make fixture-report RESULTS_DIR=${options.resultsDir}\n`);
    process.stderr.write("Rerun check-shaped evidence with: make check\n");
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
