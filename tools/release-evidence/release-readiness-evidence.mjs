#!/usr/bin/env node
import { existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "../harness/frontend/frontend-phase-manifest.mjs";
import { validateSchemaSync } from "../harness/core/public-contract.mjs";
import {
  repoRoot,
  resolveResultsRoot,
  resolveRunId,
} from "../harness/core/test-output/context.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const resolvedRepoRoot = path.resolve(scriptDir, "../..");
const schemaID = "cartulary.release_readiness_evidence.v1";
const frontendRowAccountingSchemaID = "cartulary.frontend_row_accounting.v3";

const requiredTargetEvidence = Object.freeze([
  {
    target: "check",
    evidenceClass: "product_conformance",
    conformanceEffect: "product_conformance",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#release-check"],
  },
  {
    target: "harness-contract",
    evidenceClass: "harness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#harness-contract"],
  },
  {
    target: "go-gosec-audit",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#security-gates"],
  },
  {
    target: "license-report",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#release-artifacts"],
  },
  {
    target: "sbom",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#release-artifacts"],
  },
  {
    target: "seaweedfs-release-gate",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#seaweedfs-release-gate"],
  },
  {
    target: "build-web",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#build"],
  },
  {
    target: "build-server",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#build"],
  },
  {
    target: "build-migrate",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#build"],
  },
  {
    target: "build-operator",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#build"],
  },
  {
    target: "deployable-shape",
    evidenceClass: "release_readiness",
    conformanceEffect: "not_applicable",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/testing-harness-nlspec.md#deployable-shape"],
  },
  {
    target: "browser-e2e-support",
    evidenceClass: "implementation_support",
    conformanceEffect: "no_product_conformance",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/guides/cartulary_frontend_implementation_testing_guide.md#support-evidence"],
  },
  {
    target: "browser-e2e-visual",
    evidenceClass: "design_direction",
    conformanceEffect: "no_product_conformance",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/design.md#visual-direction"],
  },
  {
    target: "browser-e2e-a11y",
    evidenceClass: "design_direction",
    conformanceEffect: "no_product_conformance",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/guides/cartulary_frontend_implementation_testing_guide.md#accessibility-evidence"],
  },
  {
    target: "browser-e2e-a11y-preflight",
    evidenceClass: "implementation_support",
    conformanceEffect: "no_product_conformance",
    claimPublicationEffect: "not_claim_bearing",
    ownerRefs: ["docs/guides/cartulary_frontend_implementation_testing_guide.md#accessibility-preflight"],
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

function artifactRef(role, file, kind = "json") {
  return { role, kind, path: relToRepo(file) };
}

function artifactRefsFromValue(role, value) {
  if (Array.isArray(value)) {
    return value.flatMap((entry, index) =>
      artifactRefsFromValue(`${role}_${index + 1}`, entry),
    );
  }
  if (value && typeof value === "object") {
    if (typeof value.path === "string" && value.path.trim() !== "") {
      return [artifactRef(value.role ?? role, value.path, value.kind ?? "json")];
    }
    return Object.entries(value).flatMap(([key, entry]) =>
      artifactRefsFromValue(`${role}_${key}`, entry),
    );
  }
  if (typeof value !== "string" || value.trim() === "") {
    return [];
  }
  return value
    .split(";")
    .map((entry) => entry.trim())
    .filter(Boolean)
    .map((entry, index) => artifactRef(index === 0 ? role : `${role}_${index + 1}`, entry));
}

function dedupeArtifactRefs(refs) {
  const seen = new Set();
  const result = [];
  for (const ref of refs) {
    const key = `${ref.role}\0${ref.kind}\0${ref.path}`;
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
      left.kind.localeCompare(right.kind),
  );
}

function collectSummaryArtifacts(summary, summaryFile) {
  return dedupeArtifactRefs([
    artifactRef("target_summary", summaryFile),
    ...artifactRefsFromValue("summary_artifact", summary?.summary_artifacts),
    ...artifactRefsFromValue("log_artifact", summary?.log_artifacts),
    ...artifactRefsFromValue("artifact", summary?.artifacts),
    ...artifactRefsFromValue("own_artifact", summary?.own?.artifacts),
    ...artifactRefsFromValue("total_artifact", summary?.totals?.artifacts),
  ]);
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
      artifact_refs: [artifactRef("expected_target_summary", file)],
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
    artifact_refs: collectSummaryArtifacts(summary, file),
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

function ownerRefsForFrontendRow(rowResult, compatRow, manifestRow) {
  const refs = [];
  for (const ownerRef of manifestRow?.owner_refs ?? []) {
    const reqs = (ownerRef.req_ids ?? []).join(",");
    const acs = (ownerRef.ac_ids ?? []).join(",");
    refs.push(
      [
        ownerRef.path,
        ownerRef.section_ref,
        reqs ? `req:${reqs}` : "",
        acs ? `ac:${acs}` : "",
      ]
        .filter(Boolean)
        .join("#"),
    );
  }
  if (refs.length > 0) {
    return sortedUniqueStrings(refs);
  }
  return [`frontend:${rowResult.phase_id}:${rowResult.row_id}:${compatRow?.target ?? "unknown-target"}`];
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

function claimPublicationEffectForRow(rowResult, compatRow, manifestRow) {
  const intent =
    manifestRow?.claim?.claim_publication_intent ??
    compatRow?.claim?.claim_publication_intent ??
    "";
  if (intent === "claim_bearing_publication") {
    return "requires_core05_review";
  }
  if (rowResult.evidence_class === "claim_publication_boundary") {
    return "claim_boundary_only";
  }
  return "not_claim_bearing";
}

function rowReleaseGateEffect(rowResult, compatRow) {
  if (rowResult.closure_status === "not_applicable") {
    return "diagnostic_only";
  }
  if (rowResult.claim_status_at_run === "blocked") {
    return "blocked";
  }
  if (compatRow?.required_for_closure === false) {
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

function artifactRefsForRow(accountingFile, accounting, rowResult) {
  const scenarioRefs = (accounting.scenario_results ?? [])
    .filter((scenario) => (scenario.row_ids ?? []).includes(rowResult.row_id))
    .flatMap((scenario) => artifactRefsFromValue("scenario_artifact", scenario.artifact_refs));
  return dedupeArtifactRefs([
    artifactRef("frontend_row_accounting", accountingFile),
    ...scenarioRefs,
  ]);
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
        owner_refs: ["docs/testing-harness-nlspec.md#frontend-row-accounting"],
        evidence_class: "diagnostic",
        conformance_effect: "not_applicable",
        claim_publication_effect: "not_applicable",
        release_gate_effect: "required",
        run_root: runRootRel,
        artifact_refs: [artifactRef("frontend_row_accounting", accountingFile)],
        status: "failed",
      });
      continue;
    }

    for (const rowResult of accounting.row_results ?? []) {
      const compatRow = (accounting.rows ?? []).find(
        (row) => row.row_id === rowResult.row_id && row.phase_id === rowResult.phase_id,
      );
      const manifestRow = rowsByID.get(rowResult.row_id);
      records.push({
        evidence_id: `frontend-row:${rowResult.row_id}:${accounting.target_name}`,
        source_target: accounting.target_name,
        schema_id: accounting.schema_id,
        owner_refs: ownerRefsForFrontendRow(rowResult, compatRow, manifestRow),
        evidence_class: rowResult.evidence_class,
        conformance_effect: conformanceEffectForRow(rowResult.evidence_class),
        claim_publication_effect: claimPublicationEffectForRow(
          rowResult,
          compatRow,
          manifestRow,
        ),
        release_gate_effect: rowReleaseGateEffect(rowResult, compatRow),
        run_root: accounting.run_root || runRootRel,
        artifact_refs: artifactRefsForRow(accountingFile, accounting, rowResult),
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
        owner_refs: ["docs/testing-harness-nlspec.md#release-readiness-evidence"],
        evidence_class: "harness",
        conformance_effect: "not_applicable",
        claim_publication_effect: "not_applicable",
        release_gate_effect: "required",
        run_root: runRootRel,
        artifact_refs: [artifactRef("release_readiness_evidence", artifactPath)],
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
