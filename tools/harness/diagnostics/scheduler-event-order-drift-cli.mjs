#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import { readCanonicalUnitEvents } from "../evidence-accounting/canonical-unit-events.mjs";

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
  let eventCount = 0;
  for await (const _event of readCanonicalUnitEvents(eventsFile)) eventCount += 1;
  process.stdout.write(`[EVENT-ORDER] target=${manifest.target} status=pass events=${eventCount}\n`);
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
}
