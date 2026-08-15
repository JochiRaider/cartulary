#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import { reduceCanonicalUnitIntervals } from "../evidence-accounting/canonical-unit-events.mjs";

function usage() {
  throw new Error("usage: scheduler-summary-timing-drift-cli.mjs [--target <target>] <run-dir>");
}

function parseArgs(argv) {
  const result = { target: "", runDir: "" };
  for (let index = 0; index < argv.length; index += 1) {
    if (argv[index] === "--target") result.target = argv[++index] ?? "";
    else if (argv[index].startsWith("--") || result.runDir) usage();
    else result.runDir = argv[index];
  }
  if (!result.runDir) usage();
  return result;
}

async function validateCanonicalRun(runDir, expectedTarget) {
  const summaryFile = path.join(runDir, "run-summary.json");
  const eventsFile = path.join(runDir, "unit-events.ndjson");
  if (!existsSync(summaryFile) || !existsSync(eventsFile)) {
    throw new Error(`${runDir} is not a canonical harness run (run-summary.json and unit-events.ndjson are required)`);
  }
  const summary = JSON.parse(readFileSync(summaryFile, "utf8"));
  validateSchemaSync(summary.schema_id, summary);
  if (expectedTarget && summary.target !== expectedTarget) {
    throw new Error(`${summaryFile} target ${summary.target} does not match ${expectedTarget}`);
  }
  const eventState = await reduceCanonicalUnitIntervals(eventsFile);
  const counts = summary.unit_counts;
  const terminalCount = counts.passed + counts.failed + counts.skipped + counts.cancelled;
  if (terminalCount !== counts.total || eventState.terminals.size !== counts.total) {
    throw new Error(`${summaryFile} unit roster does not close against canonical events`);
  }
  if (summary.wall_duration_ms < eventState.finalMonotonicMs) {
    throw new Error(`${summaryFile} wall duration is below the final canonical event`);
  }
  return { target: summary.target, units: counts.total, events: eventState.eventCount };
}

try {
  const options = parseArgs(process.argv.slice(2));
  const result = await validateCanonicalRun(path.resolve(options.runDir), options.target);
  process.stdout.write(`[SUMMARY-TIMING] target=${result.target} status=pass units=${result.units} events=${result.events}\n`);
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
}
