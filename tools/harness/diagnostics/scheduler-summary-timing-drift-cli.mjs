#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";

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

function loadEvents(file) {
  return readFileSync(file, "utf8").split(/\r?\n/u).filter(Boolean).map((line) => JSON.parse(line));
}

function validateCanonicalRun(runDir, expectedTarget) {
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
  const events = loadEvents(eventsFile);
  let previousSeq = 0;
  let previousMs = 0;
  const terminal = new Map();
  for (const entry of events) {
    validateSchemaSync(entry.schema_id, entry);
    if (entry.seq !== previousSeq + 1) throw new Error(`${eventsFile} has non-contiguous sequence at ${entry.seq}`);
    if (entry.monotonic_ms < previousMs) throw new Error(`${eventsFile} monotonic time regresses at ${entry.seq}`);
    previousSeq = entry.seq;
    previousMs = entry.monotonic_ms;
    if (["completed", "failed", "skipped", "cancelled"].includes(entry.event)) {
      if (terminal.has(entry.unit_id)) throw new Error(`${eventsFile} has duplicate terminal event for ${entry.unit_id}`);
      terminal.set(entry.unit_id, entry.status);
    }
  }
  const counts = summary.unit_counts;
  const terminalCount = counts.passed + counts.failed + counts.skipped + counts.cancelled;
  if (terminalCount !== counts.total || terminal.size !== counts.total) {
    throw new Error(`${summaryFile} unit roster does not close against canonical events`);
  }
  if (summary.wall_duration_ms < previousMs) {
    throw new Error(`${summaryFile} wall duration is below the final canonical event`);
  }
  return { target: summary.target, units: counts.total, events: events.length };
}

try {
  const options = parseArgs(process.argv.slice(2));
  const result = validateCanonicalRun(path.resolve(options.runDir), options.target);
  process.stdout.write(`[SUMMARY-TIMING] target=${result.target} status=pass units=${result.units} events=${result.events}\n`);
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
}
