#!/usr/bin/env node

import { writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  defaultTaskSurfaceManifestPath,
  harnessTierChecks,
  loadTaskSurfaceManifest,
} from "./lib/task-surface.mjs";
import {
  durationDriftDescription,
  durationDriftKind,
} from "./lib/duration-drift.mjs";
import { findFilesNamed } from "./lib/result-artifacts.mjs";
import { relToRepo, resolveRepoPath } from "./lib/repo-paths.mjs";
import {
  readJSON,
  readPositiveTargetBaseline,
  sortedObjectByKey,
} from "./lib/target-duration-baselines.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const baselineSchemaID = "cartulary.harness_smoke_duration_baselines.v1";
const defaultBaselineFile = path.join(repoRoot, "tools", "harness_smoke_duration_baselines.json");
const baselineNote =
  "Harness smoke duration weights generated from successful fast-tier harness target summaries. Refresh with make harness-smoke-duration-baselines RESULTS_DIR=<dir>.";

function usage() {
  process.stderr.write(
    [
      "usage:",
      "  harness-smoke-durations.mjs update [--baseline-file <path>] [--manifest <path>] <results-dir>",
      "  harness-smoke-durations.mjs check-drift [--baseline-file <path>] [--manifest <path>] <results-dir>",
    ].join("\n") + "\n",
  );
  process.exit(2);
}

function resolvePath(file) {
  return resolveRepoPath(repoRoot, file);
}

function rel(file) {
  return relToRepo(repoRoot, file);
}

function parseArgs(argv, command) {
  const options = {
    baselineFile: process.env.HARNESS_SMOKE_DURATION_BASELINE
      ? resolvePath(process.env.HARNESS_SMOKE_DURATION_BASELINE)
      : defaultBaselineFile,
    manifestFile: process.env.TASK_SURFACE_MANIFEST
      ? resolvePath(process.env.TASK_SURFACE_MANIFEST)
      : defaultTaskSurfaceManifestPath,
    resultsDir: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--baseline-file") {
      options.baselineFile = resolvePath(argv[index + 1] ?? "");
      index += 1;
      if (!options.baselineFile) {
        usage();
      }
      continue;
    }
    if (arg === "--manifest") {
      options.manifestFile = resolvePath(argv[index + 1] ?? "");
      index += 1;
      if (!options.manifestFile) {
        usage();
      }
      continue;
    }
    if (arg.startsWith("--")) {
      usage();
    }
    if (options.resultsDir) {
      usage();
    }
    options.resultsDir = resolvePath(arg);
  }
  if (!command || !options.resultsDir) {
    usage();
  }
  return options;
}

function readBaseline(file, { allowMissing = false } = {}) {
  return readPositiveTargetBaseline({
    repoRoot,
    file,
    schemaID: baselineSchemaID,
    missingDocument: {
      schema_id: baselineSchemaID,
      note: baselineNote,
      targets: {},
    },
    allowMissing,
  });
}

function targetSummaryFiles(root) {
  return findFilesNamed(root, "target-summary.json", { repoRoot });
}

function positiveDuration(summary) {
  const value =
    summary.critical_path_wall_duration_ms ??
    summary.wall_duration_ms ??
    summary.logical_duration_ms ??
    summary.duration_ms ??
    0;
  return Number.isInteger(value) && value > 0 ? value : 0;
}

function collectObservedHarnessDurations(resultsDir, fastChecks) {
  const observed = new Map();
  for (const file of targetSummaryFiles(resultsDir)) {
    let summary;
    try {
      summary = readJSON(repoRoot, file);
    } catch {
      continue;
    }
    if (!fastChecks.has(summary.target) || summary.status !== "pass") {
      continue;
    }
    const durationMs = positiveDuration(summary);
    if (durationMs > 0) {
      observed.set(summary.target, Math.max(observed.get(summary.target) ?? 0, durationMs));
    }
  }
  return observed;
}

function loadFastChecks(manifestFile) {
  const { manifest } = loadTaskSurfaceManifest(manifestFile);
  return harnessTierChecks(manifest, "fast");
}

function checkDrift(errors, target, actual, planned) {
  const kind = durationDriftKind(actual, planned);
  if (kind) {
    errors.push(
      durationDriftDescription(kind, {
        subject: `target=${target}`,
        plannedMs: planned,
        actualMs: actual,
      }),
    );
  }
}

function updateBaselines(options) {
  const fastChecks = loadFastChecks(options.manifestFile);
  const fastCheckSet = new Set(fastChecks);
  const observed = collectObservedHarnessDurations(options.resultsDir, fastCheckSet);
  const missingObserved = fastChecks.filter((target) => !observed.has(target));
  if (missingObserved.length > 0) {
    throw new Error(
      `missing observed fast-tier harness target summaries: ${missingObserved.join(", ")}`,
    );
  }

  const baseline = readBaseline(options.baselineFile, { allowMissing: true });
  baseline.schema_id = baselineSchemaID;
  baseline.note = baselineNote;
  baseline.updated_at = new Date().toISOString();
  baseline.targets = sortedObjectByKey(fastChecks.map((target) => [target, observed.get(target)]));
  writeFileSync(options.baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  process.stdout.write(`updated ${observed.size} harness smoke duration baselines from ${rel(options.resultsDir)}\n`);
}

function checkBaselineDrift(options) {
  const fastChecks = loadFastChecks(options.manifestFile);
  const fastCheckSet = new Set(fastChecks);
  const observed = collectObservedHarnessDurations(options.resultsDir, fastCheckSet);
  const baseline = readBaseline(options.baselineFile);
  const errors = [];

  for (const target of Object.keys(baseline.targets)) {
    if (!fastCheckSet.has(target)) {
      errors.push(`retired baseline target=${target}`);
    }
  }
  for (const target of fastChecks) {
    const planned = baseline.targets[target];
    const actual = observed.get(target);
    if (!Number.isInteger(planned) || planned <= 0) {
      errors.push(`missing harness smoke baseline target=${target}`);
      continue;
    }
    if (!Number.isInteger(actual) || actual <= 0) {
      errors.push(`missing observed harness smoke target summary target=${target}`);
      continue;
    }
    checkDrift(errors, target, actual, planned);
  }

  if (errors.length > 0) {
    process.stderr.write("Harness smoke duration baseline drift detected:\n");
    for (const error of errors) {
      process.stderr.write(`  - ${error}\n`);
    }
    process.stderr.write(
      `Refresh from a successful fast harness run with: make harness-smoke-duration-baselines RESULTS_DIR=${options.resultsDir}\n`,
    );
    process.exit(1);
  }
  process.stdout.write(`Harness smoke duration baselines match ${observed.size} fast-tier harness check(s)\n`);
}

function main(argv) {
  const [command, ...rest] = argv;
  if (command === "update") {
    updateBaselines(parseArgs(rest, command));
    return;
  }
  if (command === "check-drift") {
    checkBaselineDrift(parseArgs(rest, command));
    return;
  }
  usage();
}

try {
  main(process.argv.slice(2));
} catch (error) {
  console.error(`harness smoke duration baseline failed: ${error.message}`);
  process.exit(1);
}
