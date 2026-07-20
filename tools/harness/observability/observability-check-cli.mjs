#!/usr/bin/env node

import {
  loadRetainedObservability,
  resolveExactRunDir,
} from "./observability.mjs";

const usage = "usage: observability-check-cli.mjs --results-dir <root|run-dir> [--run-id <id>]";

function parseArgs(argv) {
  const options = { resultsDir: "", runID: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg !== "--results-dir" && arg !== "--run-id") {
      throw new Error(usage);
    }
    const value = argv[index + 1];
    if (!value) throw new Error(usage);
    if (arg === "--results-dir") options.resultsDir = value;
    else options.runID = value;
    index += 1;
  }
  if (!options.resultsDir) throw new Error(usage);
  return options;
}

function configurationFailure() {
  process.stderr.write("harness-observability-check FAIL failure_class=config reason=usage_error diagnostic=invalid-selection\n");
  process.exitCode = 2;
}

function artifactFailure() {
  process.stderr.write("harness-observability-check FAIL failure_class=artifact reason=artifact_error diagnostic=invalid-retained-diagnostics\n");
  process.exitCode = 11;
}

let options;
let runDir;
try {
  options = parseArgs(process.argv.slice(2));
  runDir = resolveExactRunDir(options.resultsDir, options.runID);
} catch {
  configurationFailure();
}

if (runDir) {
  try {
    const retained = loadRetainedObservability(runDir);
    const invocations = retained.index.invocations.length;
    const sources = retained.index.invocations.reduce(
      (total, item) => total + item.source_digests.length,
      0,
    );
    process.stdout.write(
      `harness-observability-check PASS invocations=${invocations} sources=${sources} read_only=1\n`,
    );
  } catch {
    artifactFailure();
  }
}
