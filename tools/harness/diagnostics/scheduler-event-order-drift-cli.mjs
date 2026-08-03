#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";

function usage() {
  throw new Error("usage: scheduler-event-order-drift-cli.mjs [--target <target>] <run-dir>");
}

function parseArgs(argv) {
  const options = { target: "", runDir: "" };
  for (let index = 0; index < argv.length; index += 1) {
    if (argv[index] === "--target") options.target = argv[++index] ?? "";
    else if (argv[index].startsWith("--") || options.runDir) usage();
    else options.runDir = argv[index];
  }
  if (!options.runDir) usage();
  return options;
}

try {
  const options = parseArgs(process.argv.slice(2));
  const runDir = path.resolve(options.runDir);
  const manifestFile = path.join(runDir, "run-manifest.json");
  const eventsFile = path.join(runDir, "unit-events.ndjson");
  if (!existsSync(manifestFile) || !existsSync(eventsFile)) {
    throw new Error(`${runDir} is not a canonical harness run`);
  }
  const manifest = JSON.parse(readFileSync(manifestFile, "utf8"));
  validateSchemaSync("cartulary.harness_run_manifest.v1", manifest);
  if (options.target && manifest.target !== options.target) {
    throw new Error(`${manifestFile} target ${manifest.target} does not match ${options.target}`);
  }
  const lines = readFileSync(eventsFile, "utf8").split(/\r?\n/u).filter(Boolean);
  if (lines.length === 0) throw new Error(`${eventsFile} is empty`);
  let monotonicMs = 0;
  for (const [index, line] of lines.entries()) {
    const event = JSON.parse(line);
    validateSchemaSync("cartulary.harness_unit_event.v1", event);
    if (event.seq !== index + 1) throw new Error(`${eventsFile} sequence is not contiguous at line ${index + 1}`);
    if (event.monotonic_ms < monotonicMs) throw new Error(`${eventsFile} monotonic time regresses at line ${index + 1}`);
    monotonicMs = event.monotonic_ms;
  }
  process.stdout.write(`[EVENT-ORDER] target=${manifest.target} status=pass events=${lines.length}\n`);
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
}

