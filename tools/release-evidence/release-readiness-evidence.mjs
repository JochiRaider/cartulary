#!/usr/bin/env node
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { repoRoot, validateSchemaSync } from "../harness/contract/index.mjs";
import {
  resolveResultsRoot,
  resolveRunId,
} from "../harness/contract/test-output-context.mjs";
import {
  loadTestCatalog,
  targetForCatalogRow,
} from "../harness/test-catalog/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const resolvedRepoRoot = path.resolve(scriptDir, "../..");
const schemaID = "cartulary.release_readiness_evidence.v2";
const evidenceAccountingSchemaID = "cartulary.test_evidence_accounting.v2";
const ownerSummarySchemaID = "cartulary.test_owner_summary.v2";
const releaseVerificationID =
  "harness.release.verification.current_owner_evidence_only";
const accountingVerificationID =
  "harness.evidence_accounting.verification.semantic_evidence_identity";
const visualVerificationID =
  "harness.visual.verification.stable_fixture_identity";
const designVerificationID =
  "web.design.verification.readiness_direction";
let catalogContext;

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

function sameArray(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}

function catalogRowsForTarget(targetName) {
  if (!catalogContext) {
    const catalog = loadTestCatalog(repoRoot);
    const taskSurface = readJSON(
      path.join(repoRoot, "tools/task_surface_manifest.json"),
    );
    catalogContext = {
      catalog,
      commandTargetByID: new Map(
        taskSurface.targets.map((entry) => [entry.command_id, entry.name]),
      ),
    };
  }
  const { catalog, commandTargetByID } = catalogContext;
  return catalog.rows
    .filter(
      (row) =>
        targetForCatalogRow(row, { commandTargetByID }) === targetName,
    )
    .sort((left, right) => left.row_id.localeCompare(right.row_id));
}

function ownerPartitionEvidenceRecords(runRootAbs, runRootRel, definition) {
  const rows = catalogRowsForTarget(definition.target);
  if (rows.length === 0) {
    return [];
  }
  const byOwner = new Map();
  for (const row of rows) {
    const ownerRows = byOwner.get(row.owner_id) ?? [];
    ownerRows.push(row);
    byOwner.set(row.owner_id, ownerRows);
  }
  return [...byOwner.entries()]
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([ownerID, ownerRows]) => {
      const ownerDir = path.join(runRootAbs, definition.target, "owners", ownerID);
      const accountingFile = path.join(ownerDir, "test-evidence-accounting.json");
      const ownerSummaryFile = path.join(ownerDir, "test-owner-summary.json");
      const artifactRefs = [
        artifactRef(runRootAbs, "test_evidence_accounting", accountingFile),
        artifactRef(runRootAbs, "test_owner_summary", ownerSummaryFile),
      ];
      const expectedRows = ownerRows.map((row) => row.row_id);
      const ownerRefs = sortedUniqueStrings(
        ownerRows.flatMap((row) => row.verification_ids),
      );
      const base = {
        evidence_id: `owner-partition:${definition.target}:${ownerID}`,
        source_target: definition.target,
        schema_id: evidenceAccountingSchemaID,
        owner_refs: ownerRefs,
        evidence_class: definition.evidenceClass,
        conformance_effect: definition.conformanceEffect,
        claim_publication_effect: definition.claimPublicationEffect,
        release_gate_effect: "required",
        run_root: runRootRel,
        artifact_refs: artifactRefs,
        source_refs: [],
      };
      if (!existsSync(accountingFile) || !existsSync(ownerSummaryFile)) {
        return { ...base, status: "missing" };
      }
      try {
        const accounting = readJSON(accountingFile);
        const ownerSummary = readJSON(ownerSummaryFile);
        validateSchemaSync(evidenceAccountingSchemaID, accounting);
        validateSchemaSync(ownerSummarySchemaID, ownerSummary);
        const currentCatalog = catalogContext.catalog.summary;
        const observedRows = accounting.observed_rows.map((row) => row.row_id).sort();
        const consistent =
          accounting.evidence_epoch === currentCatalog.evidence_epoch &&
          accounting.test_catalog_digest === currentCatalog.test_catalog_digest &&
          accounting.verification_routing_digest ===
            currentCatalog.verification_routing_digest &&
          accounting.owner_id === ownerID &&
          ownerSummary.owner_id === ownerID &&
          accounting.target_id === definition.target &&
          ownerSummary.target_id === definition.target &&
          sameArray(accounting.selected_rows, expectedRows) &&
          sameArray(ownerSummary.selected_rows, expectedRows) &&
          sameArray(observedRows, expectedRows) &&
          [
            "evidence_epoch",
            "run_id",
            "source_snapshot_digest",
            "test_catalog_digest",
            "verification_routing_digest",
            "runtime_profile_digest",
            "resource_profile_digest",
            "fixture_profile_digest",
          ].every((field) => accounting[field] === ownerSummary[field]) &&
          accounting.status === "pass" &&
          ownerSummary.status === "pass" &&
          accounting.observed_rows.every((row) =>
            ["passed", "skipped_authorized"].includes(row.terminal_state),
          );
        return { ...base, status: consistent ? "passed" : "failed" };
      } catch {
        return { ...base, status: "failed" };
      }
    });
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
    ...requiredTargetEvidence.flatMap((definition) => [
      targetEvidenceRecord(runRootAbs, runRootRel, definition),
      ...ownerPartitionEvidenceRecords(runRootAbs, runRootRel, definition),
    ]),
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
