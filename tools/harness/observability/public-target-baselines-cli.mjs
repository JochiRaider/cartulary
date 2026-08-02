#!/usr/bin/env node

import { lstatSync, readFileSync } from "node:fs";
import path from "node:path";

import {
  prettyJSONString,
  repoRoot,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  buildQualifiedBaseline,
  performanceEvidenceSchemaID,
} from "./performance-evidence.mjs";

const usage = "usage: public-target-baselines-cli.mjs --evidence-roots-file <manifest.json>";

function configurationFailure(message = usage) {
  process.stderr.write(`harness-public-target-duration-baselines FAIL failure_class=config reason=usage_error diagnostic=${message}\n`);
  process.exitCode = 2;
}

function loadManifest(argv) {
  if (argv.length !== 2 || argv[0] !== "--evidence-roots-file" || !argv[1]) {
    throw new Error("invalid-arguments");
  }
  const evidenceFile = path.resolve(repoRoot, argv[1]);
  const stat = lstatSync(evidenceFile);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error("invalid-evidence-manifest");
  }
  const evidence = JSON.parse(readFileSync(evidenceFile, "utf8"));
  validateSchemaSync(performanceEvidenceSchemaID, evidence);
  if (evidence.mode !== "baseline") {
    throw new Error("baseline-mode-required");
  }
  return evidence;
}

let evidence;
try {
  evidence = loadManifest(process.argv.slice(2));
} catch (error) {
  configurationFailure(error instanceof Error ? error.message : "invalid-evidence-manifest");
}

if (evidence) {
  try {
    const baseline = buildQualifiedBaseline(
      evidence.reference_windows,
      evidence.reference_bindings,
      {
        internalWindows: evidence.reference_internal_windows ?? [],
        internalBindings: evidence.reference_internal_bindings ?? [],
        rejectedRoots: evidence.reference_rejected_roots ?? [],
        role: "reference",
      },
    );
    validateSchemaSync(baseline.schema_id, baseline);
    const output = path.join(repoRoot, "tools", "harness_public_target_duration_baselines.json");
    secureWriteFile(output, prettyJSONString(baseline), { allowedRoot: repoRoot });
    process.stdout.write(`harness-public-target-duration-baselines PASS targets=${baseline.targets.length}\n`);
    process.stdout.write(`artifact=${path.relative(repoRoot, output).replaceAll("\\", "/")}\n`);
  } catch (error) {
    process.stderr.write(`harness-public-target-duration-baselines FAIL failure_class=artifact reason=duration_baseline_drift diagnostic=${error instanceof Error ? error.message : String(error)}\n`);
    process.exitCode = 13;
  }
}
