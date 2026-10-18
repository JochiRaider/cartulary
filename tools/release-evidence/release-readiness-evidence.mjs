#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");

function currentRunRoot() {
  const results = process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results";
  const runID = process.env.CARTULARY_TEST_RUN_ID || "";
  if (!runID) throw new Error("release readiness validation requires a canonical run ID");
  return path.resolve(root, results, runID);
}

function main() {
  const runRoot = currentRunRoot();
  if (process.env.CARTULARY_HARNESS_GRAPH_CHILD !== "1") {
    throw new Error("release readiness is a private graph operation; inspect retained runs with harness-observability-check CHECK=all");
  }
  // Dependencies are owned by the work graph. The marker checks one package
  // identity after its consumers finish; canonical summaries publish afterward.
  if (!existsSync(path.join(runRoot, "run-manifest.json"))) {
    throw new Error("release readiness marker has no canonical run manifest");
  }
  const producer = JSON.parse(readFileSync(path.join(runRoot, "package-release/artifacts/package.json"), "utf8"));
  for (const consumer of ["standup-package-smoke", "standup-reference-pack-smoke", "standup-operational-recovery-smoke", "credential-capacity-assessment"]) {
    const receipt = JSON.parse(readFileSync(path.join(runRoot, consumer, "artifacts/package.json"), "utf8"));
    for (const key of ["manifest_sha256", "archive_sha256"]) if (receipt[key] !== producer[key]) throw new Error(`${consumer}: package identity mismatch`);
  }
  process.stdout.write("release readiness closure reached; canonical projection pending\n");
}

try {
  main();
} catch (error) {
  process.stderr.write(`release readiness evidence failed; ${error.message}\n`);
  process.exitCode = 1;
}
