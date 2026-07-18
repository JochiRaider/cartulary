#!/usr/bin/env node

import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../contract/index.mjs";
import {
  buildToolRunSummary,
  fileArtifactRef,
  normalizeOutputMode,
} from "../output/index.mjs";
import {
  auditOwnerEvidence,
  EvidenceAuditUsageError,
} from "./index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");
const target = "test-evidence-audit";

function usage() {
  return "usage: test-evidence-audit --owner <owner-id> --evidence-roots-file <path>";
}

function parseArgs(argv) {
  const options = { ownerID: "", manifestPath: "" };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--owner") options.ownerID = argv[++index] ?? "";
    else if (arg === "--evidence-roots-file") options.manifestPath = argv[++index] ?? "";
    else throw new EvidenceAuditUsageError(usage());
  }
  if (!options.ownerID || !options.manifestPath) throw new EvidenceAuditUsageError(usage());
  return options;
}

function runID() {
  return process.env.CARTULARY_TEST_RUN_ID || `${new Date().toISOString().replaceAll(/[-:]/gu, "").replace(/\.\d{3}Z$/u, "Z")}-p${process.pid}`;
}

function resultsRoot() {
  return path.resolve(root, process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results");
}

function rel(file) {
  return path.relative(root, file).replaceAll("\\", "/");
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const id = runID();
  const startedAt = new Date().toISOString();
  const summary = auditOwnerEvidence(root, {
    ownerID: options.ownerID,
    manifestPath: options.manifestPath,
    timestamp: startedAt,
  });
  validateSchemaSync(summary.schema_id, summary);
  const targetDir = path.join(resultsRoot(), id, target);
  mkdirSync(targetDir, { recursive: true });
  const auditPath = `${target}/test-evidence-audit-summary.json`;
  writeFileSync(path.join(targetDir, "test-evidence-audit-summary.json"), `${JSON.stringify(summary, null, 2)}\n`);
  const status = summary.status;
  const finishedAt = new Date().toISOString();
  const tool = buildToolRunSummary({
    target,
    command: ["make", target],
    status,
    exitCode: status === "pass" ? 0 : 11,
    startedAt,
    completedAt: finishedAt,
    durationMs: summary.duration_ms,
    outputMode: normalizeOutputMode(),
    resultRoot: rel(resultsRoot()),
    runId: id,
    runRoot: rel(path.dirname(targetDir)),
    summaryArtifacts: [
      fileArtifactRef("test_evidence_audit_summary", auditPath),
      fileArtifactRef("tool_run_summary", `${target}/tool-run-summary.json`),
    ],
    counts: {
      tests: summary.counts.active_rows,
      failed: summary.counts.rejected_target_partitions,
    },
    failureClass: status === "pass" ? null : "artifact",
    failureReason: status === "pass" ? null : "artifact_error",
    failures: status === "pass" ? [] : summary.rejected_artifacts.map((artifact) => ({
      target: artifact.target_id,
      failure_class: "artifact",
      failure_reason: "artifact_error",
      headline: artifact.reasons.join(", "),
      artifact: artifact.artifact_path,
    })),
    rerunCommands: [`make ${target} OWNER=${options.ownerID} EVIDENCE_ROOTS_FILE=${options.manifestPath}`],
  });
  validateSchemaSync(tool.schema_id, tool);
  writeFileSync(path.join(targetDir, "tool-run-summary.json"), `${JSON.stringify(tool, null, 2)}\n`);
  if (status === "pass") {
    process.stdout.write(`[RESULT] target=${target} status=pass owner=${options.ownerID} targets=${summary.counts.accepted_target_partitions}/${summary.counts.required_target_partitions} run_root=${rel(path.dirname(targetDir))} summary_json=${auditPath}\n`);
    return 0;
  }
  process.stderr.write(`[FAIL] target=${target} exit_code=11 failure_class=artifact reason=artifact_error owner=${options.ownerID} rejected_targets=${summary.counts.rejected_target_partitions}\n`);
  return 11;
}

try {
  process.exitCode = main();
} catch (error) {
  if (error instanceof EvidenceAuditUsageError || error?.exitCode === 2) {
    process.stderr.write(`${error.message}\n`);
    process.exitCode = 2;
  } else {
    process.stderr.write(`test evidence audit failed: ${error.message}\n`);
    process.exitCode = 11;
  }
}
