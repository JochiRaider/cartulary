#!/usr/bin/env node

import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { finalizeTargetOwnerEvidence } from "../evidence-accounting/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function usage() {
  return "usage: browser-evidence-finalize-cli.mjs <target-id>";
}

function main() {
  if (process.argv.length !== 3) throw new Error(usage());
  const targetID = process.argv[2];
  if (!/^[a-z][a-z0-9-]*$/u.test(targetID)) throw new Error(usage());
  const resultsDir = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!resultsDir || !runID) {
    throw new Error(
      "CARTULARY_TEST_RESULTS_DIR and CARTULARY_TEST_RUN_ID are required",
    );
  }
  const result = finalizeTargetOwnerEvidence(root, {
    targetID,
    requestedStatus: "pass",
    resultsDir: path.resolve(root, resultsDir),
    runID,
    env: {
      ...process.env,
      CARTULARY_TARGET_EVIDENCE_SCOPE: "all",
    },
  });
  return result.status === "pass" || result.status === "not_applicable" ? 0 : 11;
}

try {
  process.exitCode = main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = error.message === usage() ? 2 : 11;
}
