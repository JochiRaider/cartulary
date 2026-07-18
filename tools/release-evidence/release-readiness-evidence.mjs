#!/usr/bin/env node
import { existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "../harness/phase-accounting/index.mjs";
import { repoRoot, validateSchemaSync } from "../harness/contract/index.mjs";
import {
  resolveResultsRoot,
  resolveRunId,
} from "../harness/contract/test-output-context.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const resolvedRepoRoot = path.resolve(scriptDir, "../..");
const schemaID = "cartulary.release_readiness_evidence.v2";
const frontendRowAccountingSchemaID = "cartulary.frontend_row_accounting.v5";
const releaseVerificationID =
  "harness.release.verification.current_owner_evidence_only";
const accountingVerificationID =
  "harness.evidence_accounting.verification.semantic_evidence_identity";
const visualVerificationID =
  "harness.visual.verification.stable_fixture_identity";
const designVerificationID =
  "web.design.verification.readiness_direction";

const requiredTargetEvidence = Object.freeze([
  {
    target: "check",
    evidenceClass: "product_conformance",
    conformanceEffect: "product_conformance",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "harness-contract",
    evidenceClass: "harness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "go-gosec-audit",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "license-report",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "sbom",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "seaweedfs-release-gate",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "build-web",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "build-server",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "build-migrate",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "build-operator",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "deployable-shape",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [releaseVerificationID],
  },
  {
    target: "browser-e2e-support",
    evidenceClass: "implementation_support",
    conformanceEffect: "no_product_conformance",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [accountingVerificationID],
  },
  {
    target: "browser-e2e-visual",
    evidenceClass: "design_direction",
    conformanceEffect: "no_product_conformance",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [visualVerificationID],
  },
  {
    target: "browser-e2e-a11y",
    evidenceClass: "design_direction",
    conformanceEffect: "no_product_conformance",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: [designVerificationID],
  },
]);

function normalizePath(value) {
  return String(value ?? "").replaceAll("\\", "/");
}

function relToRepo(file) {
  if (!file) {
    return "";
  }
  const normalized = normalizePath(file);
  if (!path.isAbsolute(file)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(resolvedRepoRoot, file));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function writeJSON(file, value) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function sortedUniqueStrings(values) {
  return [...new Set(values.filter(Boolean).map(String))].sort((left, right) =>
    left.localeCompare(right),
  );
}

function runRelativePath(runRootAbs, file) {
  const relative = normalizePath(path.relative(runRootAbs, path.resolve(file)));
  if (!relative || relative === ".." || relative.startsWith("../")) {
    throw new Error(`retained harness artifact escapes run root: ${file}`);
  }
  return relative;
}

function artifactRef(runRootAbs, role, file, format = "json") {
  return {
    role,
    path_kind: "file",
    format,
    path: runRelativePath(runRootAbs, file),
  };
}

function dedupeArtifactRefs(refs) {
  const seen = new Set();
  const result = [];
  for (const ref of refs) {
    const key = `${ref.role}\0${ref.path_kind}\0${ref.format ?? ""}\0${ref.path}`;
    if (seen.has(key) || !ref.path) {
      continue;
    }
    seen.add(key);
    result.push(ref);
  }
  return result.sort(
    (left, right) =>
      left.role.localeCompare(right.role) ||
      left.path.localeCompare(right.path) ||
      String(left.format ?? "").localeCompare(String(right.format ?? "")),
  );
}

function sourceRefsForScenario(accounting, rowResult) {
  return dedupeSourceRefs(
    (accounting.scenario_results ?? [])
      .filter((scenario) => (scenario.row_ids ?? []).includes(rowResult.row_id))
      .flatMap((scenario) => scenario.source_files ?? [])
      .map((file, index) => ({
        role: `scenario_source_${index + 1}`,
        path: normalizePath(file),
      })),
  );
}

function dedupeSourceRefs(refs) {
  const seen = new Set();
  return refs
    .filter((ref) => ref.path && !path.isAbsolute(ref.path) && !ref.path.split("/").includes(".."))
    .filter((ref) => {
      const key = `${ref.role}\0${ref.path}`;
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    })
    .sort((left, right) => left.role.localeCompare(right.role) || left.path.localeCompare(right.path));
}

function targetSummaryPath(runRootAbs, target) {
  return path.join(runRootAbs, target, "target-summary.json");
}

function readTargetSummary(runRootAbs, target) {
  const file = targetSummaryPath(runRootAbs, target);
  if (!existsSync(file)) {
    return { file, summary: null };
  }
  return { file, summary: readJSON(file) };
}

function statusFromTargetSummary(summary) {
  return summary?.status === "pass" ? "passed" : "failed";
}

function targetEvidenceRecord(runRootAbs, runRootRel, definition) {
  const { file, summary } = readTargetSummary(runRootAbs, definition.target);
  if (!summary) {
    return {
      evidence_id: `target:${definition.target}`,
      source_target: definition.target,
      schema_id: "cartulary.test_target_summary.v4",
      owner_refs: definition.ownerRefs,
      evidence_class: definition.evidenceClass,
      conformance_effect: definition.conformanceEffect,
      claim_publication_effect: definition.claimPublicationEffect,
      release_gate_effect: "required",
      run_root: runRootRel,
      artifact_refs: [artifactRef(runRootAbs, "expected_target_summary", file)],
      source_refs: [],
      status: "missing",
    };
  }
  return {
    evidence_id: `target:${definition.target}`,
    source_target: definition.target,
    schema_id: summary.schema_id ?? "cartulary.test_target_summary.v4",
    owner_refs: definition.ownerRefs,
    evidence_class: definition.evidenceClass,
    conformance_effect: definition.conformanceEffect,
    claim_publication_effect: definition.claimPublicationEffect,
    release_gate_effect: "required",
    run_root: runRootRel,
    artifact_refs: [artifactRef(runRootAbs, "target_summary", file)],
    source_refs: [],
    status: statusFromTargetSummary(summary),
  };
}

function frontendRowsByID() {
  const rows = new Map();
  try {
    const registry = loadFrontendPhaseRegistry(repoRoot);
    for (const phase of registry.phases ?? []) {
      if (!phase?.phase_id) {
        continue;
      }
      const { manifest } = loadFrontendPhaseMap(repoRoot, phase.phase_id);
      for (const row of manifest.rows ?? []) {
        rows.set(row.id, row);
      }
    }
  } catch {
    return rows;
  }
  return rows;
}

function ownerRefsForFrontendRow(rowResult, manifestRow, targetName) {
  void rowResult;
  void manifestRow;
  void targetName;
  return [accountingVerificationID];
}

function conformanceEffectForRow(evidenceClass) {
  if (evidenceClass === "product_conformance") {
    return "product_conformance";
  }
  if (evidenceClass === "TODO_owner_lookup") {
    return "owner_unresolved";
  }
  return "no_product_conformance";
}

function claimPublicationEffectForRow(rowResult, manifestRow) {
  const intent =
    manifestRow?.claim?.claim_publication_intent ??
    "";
  if (intent === "claim_bearing_publication") {
    return "requires_core05_review";
  }
  if (rowResult.evidence_class === "claim_publication_boundary") {
    return "claim_boundary_only";
  }
  return "not_claim_bearing";
}

function rowReleaseGateEffect(rowResult, manifestRow, targetName) {
  if (rowResult.closure_status === "not_applicable") {
    return "diagnostic_only";
  }
  if (rowResult.claim_status_at_run === "blocked") {
    return "blocked";
  }
  const targetRef = (manifestRow?.targets ?? []).find(
    (target) => target.target_name === targetName,
  );
  if (targetRef?.required_for_closure === false) {
    return "supporting";
  }
  return "required";
}

function rowStatus(rowResult) {
  if (rowResult.closure_status === "closed") {
    return "passed";
  }
  if (rowResult.closure_status === "blocked") {
    return "blocked";
  }
  if (rowResult.closure_status === "stale") {
    return "stale";
  }
  if (rowResult.closure_status === "not_applicable") {
    return "diagnostic_only";
  }
  return "failed";
}

function frontendRowEvidenceRecords(runRootAbs, runRootRel) {
  const records = [];
  const rowsByID = frontendRowsByID();
  if (!existsSync(runRootAbs)) {
    return records;
  }
  for (const entry of readdirSync(runRootAbs, { withFileTypes: true })) {
    if (!entry.isDirectory()) {
      continue;
    }
    const accountingFile = path.join(runRootAbs, entry.name, "frontend-row-accounting.json");
    if (!existsSync(accountingFile)) {
      continue;
    }
    let accounting;
    try {
      accounting = readJSON(accountingFile);
      validateSchemaSync(frontendRowAccountingSchemaID, accounting);
    } catch {
      const schema = accounting?.schema_id ?? "unknown";
      records.push({
        evidence_id: `frontend-row-accounting:${entry.name}:schema`,
        source_target: entry.name,
        schema_id: schema,
        owner_refs: [accountingVerificationID],
        evidence_class: "diagnostic",
        conformance_effect: "not_applicable",
        claim_publication_effect: "not_applicable",
        release_gate_effect: "required",
        run_root: runRootRel,
        artifact_refs: [artifactRef(runRootAbs, "frontend_row_accounting", accountingFile)],
        source_refs: [],
        status: "failed",
      });
      continue;
    }

    for (const rowResult of accounting.row_results ?? []) {
      const manifestRow = rowsByID.get(rowResult.row_id);
      records.push({
        evidence_id: `frontend-row:${rowResult.row_id}:${accounting.target_name}`,
        source_target: accounting.target_name,
        schema_id: accounting.schema_id,
        owner_refs: ownerRefsForFrontendRow(rowResult, manifestRow, accounting.target_name),
        evidence_class: rowResult.evidence_class,
        conformance_effect: conformanceEffectForRow(rowResult.evidence_class),
        claim_publication_effect: claimPublicationEffectForRow(
          rowResult,
          manifestRow,
        ),
        release_gate_effect: rowReleaseGateEffect(
          rowResult,
          manifestRow,
          accounting.target_name,
        ),
        run_root: accounting.run_root || runRootRel,
        artifact_refs: [artifactRef(runRootAbs, "frontend_row_accounting", accountingFile)],
        source_refs: sourceRefsForScenario(accounting, rowResult),
        status: rowStatus(rowResult),
      });
    }
  }
  return records;
}

function rollup(records) {
  const totals = {
    total: records.length,
    passed: 0,
    failed: 0,
    missing: 0,
    blocked: 0,
    stale: 0,
    diagnostic_only: 0,
    required_total: 0,
    required_passed: 0,
    required_failed: 0,
  };
  for (const record of records) {
    totals[record.status] += 1;
    if (record.release_gate_effect === "required") {
      totals.required_total += 1;
      if (record.status === "passed") {
        totals.required_passed += 1;
      } else {
        totals.required_failed += 1;
      }
    }
  }
  return totals;
}

function releaseReadinessArtifactPath(runRootAbs) {
  return path.join(runRootAbs, "release-readiness-evidence", "release-readiness-evidence.json");
}

function main() {
  const resultsRoot = resolveResultsRoot();
  const runId = resolveRunId();
  const runRootAbs = path.join(resultsRoot, runId);
  const runRootRel = relToRepo(runRootAbs);
  const records = [
    ...requiredTargetEvidence.map((definition) =>
      targetEvidenceRecord(runRootAbs, runRootRel, definition),
    ),
    ...frontendRowEvidenceRecords(runRootAbs, runRootRel),
  ].sort((left, right) => left.evidence_id.localeCompare(right.evidence_id));
  const summaryRollup = rollup(records);
  const failures = records
    .filter(
      (record) =>
        record.release_gate_effect === "required" && record.status !== "passed",
    )
    .map(
      (record) =>
        `${record.evidence_id} status=${record.status} source_target=${record.source_target}`,
    );
  const artifact = {
    schema_id: schemaID,
    status: failures.length === 0 ? "pass" : "fail",
    generated_at: new Date().toISOString(),
    run_root: runRootRel,
    evidence_records: records,
    rollup: summaryRollup,
    failures,
  };
  const artifactPath = releaseReadinessArtifactPath(runRootAbs);
  validateSchemaSync(schemaID, artifact);
  writeJSON(artifactPath, artifact);
  if (artifact.status === "fail") {
    process.stderr.write(
      `release readiness evidence failed; artifact=${relToRepo(artifactPath)} required_failed=${summaryRollup.required_failed}\n`,
    );
    for (const failure of failures) {
      process.stderr.write(`${failure}\n`);
    }
    process.exit(1);
  }
  process.stdout.write(
    `release readiness evidence passed; artifact=${relToRepo(artifactPath)} records=${records.length}\n`,
  );
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  const resultsRoot = resolveResultsRoot();
  const runRootAbs = path.join(resultsRoot, resolveRunId());
  const runRootRel = relToRepo(runRootAbs);
  const artifactPath = releaseReadinessArtifactPath(runRootAbs);
  const artifact = {
    schema_id: schemaID,
    status: "fail",
    generated_at: new Date().toISOString(),
    run_root: runRootRel,
    evidence_records: [
      {
        evidence_id: "release-readiness-evidence:fatal",
        source_target: "release-readiness-evidence",
        schema_id: schemaID,
        owner_refs: [releaseVerificationID],
        evidence_class: "harness",
        conformance_effect: "not_applicable",
        claim_publication_effect: "not_applicable",
        release_gate_effect: "required",
        run_root: runRootRel,
        artifact_refs: [artifactRef(runRootAbs, "release_readiness_evidence", artifactPath)],
        source_refs: [],
        status: "failed",
      },
    ],
    rollup: {
      total: 1,
      passed: 0,
      failed: 1,
      missing: 0,
      blocked: 0,
      stale: 0,
      diagnostic_only: 0,
      required_total: 1,
      required_passed: 0,
      required_failed: 1,
    },
    failures: [message],
  };
  validateSchemaSync(schemaID, artifact);
  writeJSON(artifactPath, artifact);
  process.stderr.write(
    `release readiness evidence failed; artifact=${relToRepo(artifactPath)}\n${message}\n`,
  );
  process.exit(1);
}
