#!/usr/bin/env node

import process from "node:process";

import { deriveTestRowID } from "./row-id-authoring.mjs";

function parseArgs(argv) {
  const options = { familyID: "", claim: "", selectorKey: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--family-id") options.familyID = argv[++index] ?? "";
    else if (arg === "--claim") options.claim = argv[++index] ?? "";
    else if (arg === "--selector-key") options.selectorKey = argv[++index] ?? "";
    else throw new Error("invalid test row ID authoring usage");
  }
  return options;
}

try {
  process.stdout.write(`${deriveTestRowID(parseArgs(process.argv.slice(2)))}\n`);
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 2;
}
