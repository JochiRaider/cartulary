#!/usr/bin/env node

import { lstatSync, readFileSync } from "node:fs";
import path from "node:path";

import { repoRoot, validateSchemaSync } from "../contract/index.mjs";
import {
  comparePerformanceWindows,
  performanceEvidenceSchemaID,
} from "./performance-evidence.mjs";

const usage = "usage: performance-check-cli.mjs --evidence-roots-file <manifest.json>";

function configurationFailure(message = usage) {
  process.stderr.write(`harness-performance-check FAIL failure_class=config reason=usage_error diagnostic=${message}\n`);
  process.exitCode = 2;
}

function acceptanceFailure(message) {
  process.stderr.write(`harness-performance-check FAIL failure_class=timing reason=duration_baseline_drift diagnostic=${message}\n`);
  process.exitCode = 13;
}

function loadManifest(argv) {
  if (argv.length !== 2 || argv[0] !== "--evidence-roots-file" || !argv[1]) {
    throw new Error("invalid-arguments");
  }
  const file = path.resolve(repoRoot, argv[1]);
  const stat = lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error("invalid-evidence-manifest");
  }
  const manifest = JSON.parse(readFileSync(file, "utf8"));
  validateSchemaSync(performanceEvidenceSchemaID, manifest);
  if (manifest.mode !== "comparison") {
    throw new Error("comparison-mode-required");
  }
  return manifest;
}

let manifest;
try {
  manifest = loadManifest(process.argv.slice(2));
} catch (error) {
  configurationFailure(error instanceof Error ? error.message : "invalid-evidence-manifest");
}

if (manifest) {
  try {
    const comparison = comparePerformanceWindows(manifest);
    for (const windowName of ["baseline", "candidate"]) {
      for (const rejected of comparison.rejected_roots[windowName]) {
        process.stdout.write(
          `[PERFORMANCE-REJECTED] window=${windowName} root=${rejected.root} reasons=${rejected.reasons.join(",")}\n`,
        );
      }
    }
    for (const row of comparison.rows) {
      process.stdout.write(
        `[PERFORMANCE] target=${row.target} gate=${row.gate} status=${row.status} baseline_median_ms=${row.baseline_median_ms} baseline_mad_ms=${row.baseline_mad_ms} candidate_median_ms=${row.candidate_median_ms} limit_ms=${row.limit_ms}\n`,
      );
    }
    process.stdout.write(
      `[PERFORMANCE-PORTFOLIO] status=${comparison.portfolio.status} targets=${comparison.portfolio.target_count} baseline_total_ms=${comparison.portfolio.baseline_total_ms} candidate_total_ms=${comparison.portfolio.candidate_total_ms} delta_ms=${comparison.portfolio.delta_ms}\n`,
    );
    if (comparison.failures.length > 0) {
      acceptanceFailure(`failed-targets:${comparison.failures.join(",")}`);
    } else {
      process.stdout.write(`harness-performance-check PASS targets=${comparison.rows.length}\n`);
    }
  } catch (error) {
    acceptanceFailure(error instanceof Error ? error.message : "invalid-performance-evidence");
  }
}
