#!/usr/bin/env node

import path from "node:path";

import { validateCanonicalRun } from "./canonical-evidence.mjs";

try {
  const value = process.env.RESULTS_DIR || process.argv[2];
  if (!value) throw new Error("RESULTS_DIR is required");
  const run = validateCanonicalRun(path.resolve(value), process.env.TARGET || "");
  process.stdout.write(`[EVIDENCE] target=${run.manifest.target} status=pass units=${run.summary.unit_counts.total} events=${run.events.length}\n`);
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
}
