#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  buildFixtureReport,
  fixtureReportSchemaID,
  fixtureSummaryLine,
  fixtureSummaryLines,
  normalizeFixtureThreshold,
  normalizeFixtureTop,
  resolveFixtureResultLocation,
} from "./fixture-reporting.mjs";
import {
  prettyJSONString,
  validateSchemaSync,
} from "../contract/harness-contract.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");

function usage() {
  process.stderr.write(
    "usage: print-fixture-report.mjs [--results-dir <dir>] [--run-id <id>] [--target <target>] [--threshold-ms <n>] [--top <n>] [--json]\n",
  );
  process.exit(2);
}

function resolveResultsDir(value) {
  const configured =
    value ||
    process.env.RESULTS_DIR ||
    process.env.CARTULARY_TEST_RESULTS_DIR ||
    ".cartulary/test-results";
  return path.isAbsolute(configured) ? configured : path.join(repoRoot, configured);
}

function parseArgs(argv) {
  const options = {
    resultsRoot: "",
    runId: process.env.RUN_ID || "",
    target: process.env.TARGET || "",
    thresholdMs: normalizeFixtureThreshold(
      process.env.FIXTURE_THRESHOLD_MS ?? process.env.CARTULARY_FIXTURE_THRESHOLD_MS,
    ),
    top: normalizeFixtureTop(process.env.FIXTURE_TOP ?? process.env.CARTULARY_FIXTURE_TOP),
    json: process.env.JSON === "1",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--results-dir") {
      options.resultsRoot = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--run-id") {
      options.runId = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--target") {
      options.target = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--threshold-ms") {
      options.thresholdMs = normalizeFixtureThreshold(argv[index + 1]);
      index += 1;
      continue;
    }
    if (arg === "--top") {
      options.top = normalizeFixtureTop(argv[index + 1]);
      index += 1;
      continue;
    }
    if (arg === "--json") {
      options.json = true;
      continue;
    }
    usage();
  }
  const location = resolveFixtureResultLocation({
    resultsDir: resolveResultsDir(options.resultsRoot),
    runId: options.runId,
    repoRoot,
  });
  options.resultsRoot = location.resultsRoot;
  options.runId = location.runId;
  options.runDir = location.runDir;
  return options;
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const report = buildFixtureReport({
    resultsRoot: options.resultsRoot,
    runId: options.runId,
    target: options.target,
    thresholdMs: options.thresholdMs,
    repoRoot,
  });

  if (options.json) {
    validateSchemaSync(fixtureReportSchemaID, report);
    process.stdout.write(prettyJSONString(report));
    return;
  }

  const aggregateLine = fixtureSummaryLine(report.aggregate, {
    thresholdMs: options.thresholdMs,
    top: options.top,
  });
  const targetLines = options.target
    ? []
    : fixtureSummaryLines(report.targets, {
        thresholdMs: options.thresholdMs,
        top: options.top,
      });
  for (const line of [aggregateLine, ...targetLines].filter(Boolean)) {
    process.stdout.write(`${line}\n`);
  }
}

try {
  main();
} catch (error) {
  process.stderr.write(`fixture report failed: ${error.message}\n`);
  process.exit(1);
}
