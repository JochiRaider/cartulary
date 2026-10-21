#!/usr/bin/env node

import {
  resolveExactRunDir,
} from "./observability.mjs";

import { checkRetainedObservability, observabilityChecks } from "./retained-checks.mjs";

const usage = "usage: observability-check-cli.mjs --results-dir <root|run-dir> [--run-id <id>] [--check all|events|timing|evidence|diagnostics]";

function parseArgs(argv) {
  const options = { resultsDir: "", runID: "", check: "all" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg !== "--results-dir" && arg !== "--run-id" && arg !== "--check") {
      throw new Error(usage);
    }
    const value = argv[index + 1];
    if (!value) throw new Error(usage);
    if (arg === "--results-dir") options.resultsDir = value;
    else if (arg === "--run-id") options.runID = value;
    else options.check = value;
    index += 1;
  }
  if (!options.resultsDir || !observabilityChecks.includes(options.check)) throw new Error(usage);
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
    const result = await checkRetainedObservability(runDir, options.check);
    const diagnostics = result.diagnostics ? Object.entries(result.diagnostics)
      .map(([key, value]) => ` diagnostic_${key}=${value ?? "unavailable"}`).join("") : "";
    process.stdout.write(`harness-observability-check PASS check=${result.check} full_validity=${result.full_validity ? 1 : 0} run_id=${result.run_id} read_only=1${diagnostics}\n`);
  } catch {
    artifactFailure();
  }
}
