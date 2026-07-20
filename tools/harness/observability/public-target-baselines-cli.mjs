#!/usr/bin/env node

import { readFileSync } from "node:fs";
import path from "node:path";

import {
  prettyJSONString,
  repoRoot,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import { buildQualifiedBaseline } from "./performance-evidence.mjs";

function usage() {
  process.stderr.write("usage: public-target-baselines-cli.mjs --evidence-roots-file <manifest.json>\n");
  process.exit(2);
}

try {
  const args = process.argv.slice(2);
  if (args.length !== 2 || args[0] !== "--evidence-roots-file") usage();
  const evidenceFile = path.resolve(repoRoot, args[1]);
  const evidence = JSON.parse(readFileSync(evidenceFile, "utf8"));
  validateSchemaSync("cartulary.harness_performance_evidence_roots.v1", evidence);
  const baseline = buildQualifiedBaseline(evidence.baseline_roots);
  validateSchemaSync(baseline.schema_id, baseline);
  const output = path.join(repoRoot, "tools", "harness_public_target_duration_baselines.json");
  secureWriteFile(output, prettyJSONString(baseline), { allowedRoot: repoRoot });
  process.stdout.write(`harness-public-target-duration-baselines PASS targets=${baseline.targets.length}\n`);
  process.stdout.write(`artifact=${path.relative(repoRoot, output).replaceAll("\\", "/")}\n`);
} catch (error) {
  process.stderr.write(`harness-public-target-duration-baselines FAIL diagnostic=${error instanceof Error ? error.message : String(error)}\n`);
  process.exit(13);
}
