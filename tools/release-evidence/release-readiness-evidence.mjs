#!/usr/bin/env node

import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { validateCanonicalRun } from "../harness/observability/canonical-evidence.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const requiredProjections = [
  "browser-e2e-a11y",
  "browser-e2e-support",
  "browser-e2e-visual",
  "build-migrate",
  "build-operator",
  "build-server",
  "build-web",
  "credential-capacity-assessment",
  "cryptographic-policy-assessment",
  "deployable-shape",
  "go-gosec-audit",
  "harness-contract",
  "license-report",
  "package-inspect",
  "package-release",
  "release-readiness-evidence",
  "sbom",
  "seaweedfs-release-gate",
  "standup-operational-recovery-smoke",
  "standup-package-smoke",
  "standup-reference-pack-smoke",
];

function currentRunRoot() {
  const results = process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results";
  const runID = process.env.CARTULARY_TEST_RUN_ID || "";
  if (!runID) throw new Error("release readiness validation requires a canonical run ID");
  return path.resolve(root, results, runID);
}

function main() {
  const runRoot = currentRunRoot();
  if (process.env.CARTULARY_HARNESS_GRAPH_CHILD === "1") {
    // The scheduler publishes target projections atomically after every unit is
    // terminal. This unit is the dependency-closure marker; it must not read a
    // partial run or write a parallel release-evidence format.
    if (!existsSync(path.join(runRoot, "run-manifest.json"))) {
      throw new Error("release readiness marker has no canonical run manifest");
    }
    const producer = JSON.parse(readFileSync(path.join(runRoot,"package-release/artifacts/package.json"),"utf8"));
    for (const consumer of ["standup-package-smoke","standup-reference-pack-smoke","standup-operational-recovery-smoke","credential-capacity-assessment"]) {
      const receipt = JSON.parse(readFileSync(path.join(runRoot,consumer,"artifacts/package.json"),"utf8"));
      for (const key of ["manifest_sha256","archive_sha256"]) if (receipt[key] !== producer[key]) throw new Error(`${consumer}: package identity mismatch`);
    }
    process.stdout.write("release readiness closure reached; canonical projection pending\n");
    return;
  }

  const retained = validateCanonicalRun(runRoot, "release-check");
  const failures = [];
  for (const target of requiredProjections) {
    const summary = retained.targetSummaries.get(target);
    if (!summary) failures.push(`${target}:missing`);
    else if (summary.status !== "pass") failures.push(`${target}:${summary.status}`);
  }
  if (retained.summary.status !== "pass") failures.push(`release-check:${retained.summary.status}`);
  if (failures.length > 0) {
    throw new Error(`canonical release readiness failed: ${failures.join(",")}`);
  }
  process.stdout.write(
    `release readiness evidence passed; source=canonical target_projections=${requiredProjections.length}\n`,
  );
}

try {
  main();
} catch (error) {
  process.stderr.write(`release readiness evidence failed; ${error.message}\n`);
  process.exitCode = 1;
}
