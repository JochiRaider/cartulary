#!/usr/bin/env node
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../contract/index.mjs";
import {
  buildFrontendEvidenceAuditSummary,
  frontendEvidenceAuditSummarySchemaID,
  relativeFrontendEvidenceAuditPath,
} from "./frontend/evidence-audit.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");

function input(name) {
  return String(process.env[name] ?? "").trim();
}

function requireInput(name, failures) {
  const value = input(name);
  if (value === "") {
    failures.push(`${name} is required`);
  }
  return value;
}

function commandSpecificSummaryPath() {
  const dir =
    process.env.CARTULARY_PHASE_ARTIFACT_DIR ||
    path.join(
      repoRoot,
      ".cartulary",
      "test-results",
      "frontend-evidence-audit-direct",
    );
  mkdirSync(dir, { recursive: true });
  return path.join(dir, "frontend-evidence-audit-summary.json");
}

function writeSummary(summary) {
  const file = commandSpecificSummaryPath();
  validateSchemaSync(frontendEvidenceAuditSummarySchemaID, summary);
  writeFileSync(file, `${JSON.stringify(summary, null, 2)}\n`);
  return file;
}

function resolvedInputs() {
  const failures = [];
  const phaseNamespace = requireInput("PHASE_NAMESPACE", failures);
  const phaseID = requireInput("PHASE", failures);
  const roots = {
    CHECK_RESULTS_DIR: input("CHECK_RESULTS_DIR"),
    BROWSER_SUPPORT_RESULTS_DIR: input("BROWSER_SUPPORT_RESULTS_DIR"),
    BROWSER_VISUAL_RESULTS_DIR: input("BROWSER_VISUAL_RESULTS_DIR"),
    BROWSER_A11Y_RESULTS_DIR: input("BROWSER_A11Y_RESULTS_DIR"),
    BROWSER_A11Y_PREFLIGHT_RESULTS_DIR: input("BROWSER_A11Y_PREFLIGHT_RESULTS_DIR"),
    BROWSER_MEASUREMENT_RESULTS_DIR: input("BROWSER_MEASUREMENT_RESULTS_DIR"),
  };

  if (phaseNamespace && phaseNamespace !== "frontend") {
    failures.push("PHASE_NAMESPACE must be frontend");
  }
  if (phaseID && !/^FE-P(?:0|[1-9][0-9]*)$/.test(phaseID)) {
    failures.push("PHASE must be FE-P<N>");
  }

  return {
    phaseNamespace,
    phaseID,
    roots,
    failures,
  };
}

function main() {
  const { phaseNamespace, phaseID, roots, failures } = resolvedInputs();
  const summary = buildFrontendEvidenceAuditSummary({
    repoRoot,
    phaseNamespace,
    phaseID,
    roots,
    failures,
  });
  const summaryFile = writeSummary(summary);
  if (summary.status === "fail") {
    process.stderr.write(
      `frontend evidence audit failed; summary=${relativeFrontendEvidenceAuditPath(repoRoot, summaryFile)}\n${summary.failures.join("\n")}\n`,
    );
    process.exit(1);
  }
  process.stdout.write(
    `frontend evidence audit passed phase=${phaseID} summary=${relativeFrontendEvidenceAuditPath(repoRoot, summaryFile)}\n`,
  );
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  const summaryFile = writeSummary({
    schema_id: frontendEvidenceAuditSummarySchemaID,
    status: "fail",
    phase_namespace: input("PHASE_NAMESPACE"),
    phase_id: input("PHASE"),
    roots: {},
    digests: {},
    targets: [],
    rows: [],
    failures: [message],
  });
  process.stderr.write(
    `frontend evidence audit failed; summary=${relativeFrontendEvidenceAuditPath(repoRoot, summaryFile)}\n${message}\n`,
  );
  process.exit(1);
}
