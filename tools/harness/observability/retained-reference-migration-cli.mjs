#!/usr/bin/env node

import { lstatSync } from "node:fs";
import path from "node:path";

import {
  prettyJSONString,
  repoRoot,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  buildRetainedV1ReferenceManifest,
  performanceEvidenceSchemaID,
} from "./performance-evidence.mjs";

const usage = "usage: retained-reference-migration-cli.mjs --results-root <dir> --window-start <RFC3339> --window-end <RFC3339> --output <file>";

function parseArgs(argv) {
  if (argv.length !== 8) throw new Error(usage);
  const values = {};
  for (let index = 0; index < argv.length; index += 2) values[argv[index]] = argv[index + 1];
  for (const name of ["--results-root", "--window-start", "--window-end", "--output"]) {
    if (!values[name]) throw new Error(usage);
  }
  const resultsRoot = path.resolve(repoRoot, values["--results-root"]);
  if (!lstatSync(resultsRoot).isDirectory()) throw new Error("results root is not a directory");
  return {
    resultsRoot,
    windowStartedAt: values["--window-start"],
    windowEndedAt: values["--window-end"],
    output: path.resolve(repoRoot, values["--output"]),
  };
}

try {
  const args = parseArgs(process.argv.slice(2));
  const manifest = buildRetainedV1ReferenceManifest(args.resultsRoot, args);
  validateSchemaSync(performanceEvidenceSchemaID, manifest);
  secureWriteFile(args.output, prettyJSONString(manifest), { allowedRoot: repoRoot });
  process.stdout.write(
    `retained-v1-reference-migration PASS contexts=${manifest.reference_audit.inspected_contexts} nominal=${manifest.reference_audit.nominally_eligible_contexts} strict_accepted=${manifest.reference_audit.strict_accepted_contexts} strict_rejected=${manifest.reference_audit.strict_rejected_contexts} windows=${manifest.reference_windows.length} targets=${manifest.reference_bindings.length}\n`,
  );
  process.stdout.write(`artifact=${path.relative(repoRoot, args.output).replaceAll("\\", "/")}\n`);
} catch (error) {
  process.stderr.write(`retained-v1-reference-migration FAIL ${error instanceof Error ? error.message : String(error)}\n`);
  process.exitCode = 13;
}
