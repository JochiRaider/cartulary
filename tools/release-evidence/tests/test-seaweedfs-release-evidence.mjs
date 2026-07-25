#!/usr/bin/env node
import assert from "node:assert/strict";
import { Buffer } from "node:buffer";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import {
  buildDependencyBoundary,
  buildOccurrenceInventoryFromEntries,
  buildMigrationPreservationEvidence,
  buildRedactionLeakageScan,
  buildReleaseManifestExposure,
  buildReleaseGateSummary,
  buildSeaweedFSCompatibilityEvidence,
  buildSbomLicenseClassification,
  buildStorageRefOwnerCoverage,
  buildThreatModelCoverage,
  globToRegExp,
  isDefaultExcluded,
  isTextBuffer,
  occurrenceTokens,
} from "../seaweedfs-release-evidence.mjs";

const legacyStem = "min" + "io";
const legacyTitle = "Min" + "IO";
const upperLegacyTitle = "MIN" + "IO";
const bucketToken = `${legacyStem}_bucket`;
const endpointToken = `${legacyStem}_endpoint`;
const sdkModule = `github.com/${legacyStem}/${legacyStem}-go/v7`;
const serverImage = `${legacyStem}/${legacyStem}`;
const generatedPolicy = {
  roots: ["internal/gen", "packages/protocol-ts/src/generated"],
  files: ["tools/task_surface.generated.mk"],
};

const rules = [
  {
    id: "sdk",
    pathRegexps: [globToRegExp("adapter/**")],
    classification: "sdk_only",
    owner: "fixture sdk owner",
    rationale: "fixture SDK boundary",
  },
  {
    id: "legacy",
    pathRegexps: [globToRegExp("testdata/history/**")],
    classification: "historical_changelog",
    owner: "fixture archive owner",
    rationale: "fixture archive boundary",
  },
  {
    id: "invalid",
    pathRegexps: [globToRegExp("bad/**")],
    classification: "invalid",
  },
];

assert.deepEqual(occurrenceTokens, [
  legacyTitle,
  legacyStem,
  upperLegacyTitle,
  bucketToken,
  endpointToken,
]);

assert.equal(isDefaultExcluded(".cartulary/test-results/run.json", generatedPolicy), true);
assert.equal(isDefaultExcluded("internal/gen/contracts/contracts_gen.go", generatedPolicy), true);
assert.equal(isDefaultExcluded("tools/task_surface.generated.mk", generatedPolicy), true);
assert.equal(isDefaultExcluded("internal/platform/objectstore/objectstore.go", generatedPolicy), false);
assert.equal(isTextBuffer(Buffer.from("plain text")), true);
assert.equal(isTextBuffer(Buffer.from([0x61, 0x00, 0x62])), false);

const inventory = buildOccurrenceInventoryFromEntries({
  rules,
  repoCommitValue: "abc123",
  scannedAt: "2026-06-04T00:00:00.000Z",
  entries: [
    {
      path: "testdata/history/legacy.txt",
      text: `${legacyTitle} history\n`,
    },
    {
      path: "adapter/client.go",
      text: `import "${sdkModule}"\nconst a = "${bucketToken}"\n`,
    },
    {
      path: "bad/default.txt",
      text: `${upperLegacyTitle} server default\n`,
    },
    {
      path: "unclassified.txt",
      text: `${endpointToken}\n`,
    },
  ],
});
assert.equal(inventory.schema_id, "cartulary.seaweedfs_migration_occurrence_inventory.v1");
assert.equal(inventory.result, "fail");
assert.equal(inventory.occurrences.length, 8);
assert.deepEqual(
  inventory.occurrences.map((entry) => [entry.path, entry.line, entry.column, entry.token, entry.classification]),
  [
    ["adapter/client.go", 1, 20, legacyStem, "sdk_only"],
    ["adapter/client.go", 1, 26, legacyStem, "sdk_only"],
    ["adapter/client.go", 2, 12, legacyStem, "sdk_only"],
    ["adapter/client.go", 2, 12, bucketToken, "sdk_only"],
    ["bad/default.txt", 1, 1, upperLegacyTitle, "invalid"],
    ["testdata/history/legacy.txt", 1, 1, legacyTitle, "historical_changelog"],
    ["unclassified.txt", 1, 1, legacyStem, "unclassified"],
    ["unclassified.txt", 1, 1, endpointToken, "unclassified"],
  ].sort((a, b) => a[0].localeCompare(b[0]) || a[1] - b[1] || a[2] - b[2] || a[3].localeCompare(b[3])),
);
assert.equal(
  inventory.occurrences.every((entry) => !Object.hasOwn(entry, "line_excerpt")),
  true,
);

const passingInventory = buildOccurrenceInventoryFromEntries({
  rules,
  entries: [
    {
      path: "adapter/client.go",
      text: `import "${sdkModule}"\n`,
    },
  ],
});
assert.equal(passingInventory.result, "pass");

const releaseFixture = new Map([
  ["docker-compose.release-test.yml", `services:\n  object:\n    image: ${serverImage}:latest\n`],
  ["configs/release-test.toml", "cors_allowed_origins = \"*\"\n"],
  ["internal/modules/evidence/leak.go", 'const api = "/cluster/status"\n'],
]);
const exposure = buildReleaseManifestExposure({
  files: [...releaseFixture.keys()],
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  readFile: (rel) => releaseFixture.get(rel),
});
assert.equal(exposure.result, "fail");
assert.deepEqual(
  new Set(exposure.findings.map((finding) => finding.check_id)),
  new Set(["legacy-server-image", "wildcard-direct-upload-cors", "runtime-seaweedfs-admin-api"]),
);

const cleanExposure = buildReleaseManifestExposure({
  files: ["docker-compose.dev.yml"],
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  readFile: () =>
    "services:\n  object:\n    image: docker.io/chrislusf/seaweedfs:4.17@sha256:abc\n    ports:\n      - \"8333:8333\"\n    command: [\"server\", \"-s3\", \"-webdav=false\"]\n",
});
assert.equal(cleanExposure.result, "pass");

const dependencyFixture = new Map([
  ["go.mod", `module fixture\nrequire ${sdkModule} v7.0.100\n`],
  ["internal/platform/objectstore/objectstore.go", `import "${sdkModule}"\n`],
  ["internal/modules/evidence/evidence.go", `import "${sdkModule}"\n`],
]);
const dependency = buildDependencyBoundary({
  files: [...dependencyFixture.keys()],
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  goModText: dependencyFixture.get("go.mod"),
  readFile: (rel) => dependencyFixture.get(rel),
});
assert.equal(dependency.result, "fail");
assert.equal(dependency.sdk_dependency.disallowed_imports.length, 1);

const dependencyServer = buildDependencyBoundary({
  files: ["go.mod", "internal/platform/objectstore/objectstore.go"],
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  goModText: dependencyFixture.get("go.mod"),
  readFile: (rel) => dependencyFixture.get(rel) ?? `import "${sdkModule}"\n`,
  sbom: { components: [{ name: serverImage }] },
  licenseReport: { dependencies: [{ package: sdkModule }] },
});
assert.equal(dependencyServer.result, "fail");
assert.equal(dependencyServer.release_server_artifacts.length, 1);

const sbomLicense = buildSbomLicenseClassification({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  sbomPath: "fixture-sbom.json",
  licensePath: "fixture-license.json",
  sbom: { components: [{ name: sdkModule, purl: `pkg:golang/${sdkModule}` }] },
  licenseReport: { dependencies: [{ package_name: sdkModule, license: "Apache-2.0" }] },
});
assert.equal(sbomLicense.result, "pass");
assert.equal(sbomLicense.sdk_dependency.classification, "sdk_only");

const sbomLicenseServer = buildSbomLicenseClassification({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  sbomPath: "fixture-sbom.json",
  licensePath: "fixture-license.json",
  sbom: { components: [{ name: serverImage }] },
  licenseReport: { dependencies: [] },
});
assert.equal(sbomLicenseServer.result, "fail");
assert.equal(sbomLicenseServer.release_server_artifacts.length, 1);

const threat = buildThreatModelCoverage({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
});
assert.equal(threat.result, "pass");

const storageRefOwner = buildStorageRefOwnerCoverage({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
});
assert.equal(storageRefOwner.result, "pass");

function compatibilityReport(statusByCase = {}) {
  return {
    schema_id: "cartulary.seaweedfs_s3_compatibility_report.v1",
    probe_id: "fixture",
    object_store_backend: "seaweedfs_s3",
    started_at: "2026-06-04T00:00:00Z",
    completed_at: "2026-06-04T00:00:01Z",
    result: Object.values(statusByCase).some((status) => status !== "pass") ? "fail" : "pass",
    cases: Array.from({ length: 14 }, (_, index) => {
      const caseID = `SWFS-COMP-${String(index + 1).padStart(3, "0")}`;
      return {
        case_id: caseID,
        capability: "fixture",
        status: statusByCase[caseID] ?? "pass",
        reason_code: null,
        evidence: { source: "fixture" },
      };
    }),
    forbidden_skip_rows: [],
  };
}

const currentResultsDir = ".cartulary/test-results";
const currentRunID = "unit-current-run";
const currentCompatibilityReportPath = `${currentResultsDir}/${currentRunID}/seaweedfs-compatibility/object-store-compatibility-report.json`;
const passingCompatibilitySummary = {
  schema_id: "cartulary.tool_run_summary.v5",
  target: "seaweedfs-compatibility",
  status: "pass",
};
const compatibility = buildSeaweedFSCompatibilityEvidence({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  reportPath: currentCompatibilityReportPath,
  report: compatibilityReport(),
  requireCurrentRun: true,
  currentResultsDir,
  currentRunId: currentRunID,
  targetSummary: passingCompatibilitySummary,
});
assert.equal(compatibility.result, "pass");

const partialCompatibility = buildSeaweedFSCompatibilityEvidence({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  reportPath: currentCompatibilityReportPath,
  report: compatibilityReport({ "SWFS-COMP-007": "not_run" }),
  requireCurrentRun: true,
  currentResultsDir,
  currentRunId: currentRunID,
  targetSummary: passingCompatibilitySummary,
});
assert.equal(partialCompatibility.result, "fail");
assert.equal(
  partialCompatibility.findings.some((finding) => finding.check_id === "compatibility-case-status"),
  true,
);

const stableCompatibility = buildSeaweedFSCompatibilityEvidence({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  reportPath: ".cartulary/release-artifacts/seaweedfs/object-store-compatibility-report.json",
  report: compatibilityReport(),
  requireCurrentRun: true,
  currentResultsDir,
  currentRunId: currentRunID,
  targetSummary: passingCompatibilitySummary,
});
assert.equal(stableCompatibility.result, "fail");
assert.equal(
  stableCompatibility.findings.some((finding) => finding.check_id === "compatibility-stable-report-source"),
  true,
);

const servicesUpCompatibility = buildSeaweedFSCompatibilityEvidence({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  reportPath:
    ".cartulary/test-results/20260603T224410Z-p820796/services-up/object-store-compatibility-report.json",
  report: compatibilityReport(),
  requireCurrentRun: true,
  currentResultsDir,
  currentRunId: currentRunID,
  targetSummary: passingCompatibilitySummary,
});
assert.equal(servicesUpCompatibility.result, "fail");
assert.equal(
  servicesUpCompatibility.findings.some((finding) => finding.check_id === "compatibility-services-up-source"),
  true,
);

const missingSummaryCompatibility = buildSeaweedFSCompatibilityEvidence({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  reportPath:
    ".cartulary/test-results/unit-current-run-missing-summary/seaweedfs-compatibility/object-store-compatibility-report.json",
  report: compatibilityReport(),
  requireCurrentRun: true,
  currentResultsDir,
  currentRunId: "unit-current-run-missing-summary",
});
assert.equal(missingSummaryCompatibility.result, "fail");
assert.equal(
  missingSummaryCompatibility.findings.some(
    (finding) => finding.check_id === "compatibility-target-summary-present",
  ),
  true,
);

const failedSummaryCompatibility = buildSeaweedFSCompatibilityEvidence({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  reportPath: currentCompatibilityReportPath,
  report: compatibilityReport(),
  requireCurrentRun: true,
  currentResultsDir,
  currentRunId: currentRunID,
  targetSummary: {
    schema_id: "cartulary.tool_run_summary.v5",
    target: "seaweedfs-compatibility",
    status: "fail",
  },
});
assert.equal(failedSummaryCompatibility.result, "fail");
assert.equal(
  failedSummaryCompatibility.findings.some(
    (finding) => finding.check_id === "compatibility-target-summary-status",
  ),
  true,
);

const skippedPrerequisiteCompatibility = buildSeaweedFSCompatibilityEvidence({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  reportPath: currentCompatibilityReportPath,
  report: compatibilityReport(),
  requireCurrentRun: true,
  currentResultsDir,
  currentRunId: currentRunID,
  targetSummary: passingCompatibilitySummary,
  prerequisitesSkipped: true,
});
assert.equal(skippedPrerequisiteCompatibility.result, "fail");
assert.equal(
  skippedPrerequisiteCompatibility.findings.some(
    (finding) => finding.check_id === "compatibility-current-run-prerequisites",
  ),
  true,
);

const previousSkipPrerequisites = process.env.CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES;
const previousSequencePrerequisites = process.env.CARTULARY_SEQUENCE_PREREQUISITES_SATISFIED;
try {
  process.env.CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES = "1";
  process.env.CARTULARY_SEQUENCE_PREREQUISITES_SATISFIED = "1";
  const sequenceOwnedPrerequisiteCompatibility = buildSeaweedFSCompatibilityEvidence({
    repoCommitValue: "abc123",
    generatedAt: "2026-06-04T00:00:00.000Z",
    reportPath: currentCompatibilityReportPath,
    report: compatibilityReport(),
    requireCurrentRun: true,
    currentResultsDir,
    currentRunId: currentRunID,
    targetSummary: passingCompatibilitySummary,
  });
  assert.equal(sequenceOwnedPrerequisiteCompatibility.result, "pass");
} finally {
  if (previousSkipPrerequisites === undefined) {
    delete process.env.CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES;
  } else {
    process.env.CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES = previousSkipPrerequisites;
  }
  if (previousSequencePrerequisites === undefined) {
    delete process.env.CARTULARY_SEQUENCE_PREREQUISITES_SATISFIED;
  } else {
    process.env.CARTULARY_SEQUENCE_PREREQUISITES_SATISFIED = previousSequencePrerequisites;
  }
}

const migration = buildMigrationPreservationEvidence({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  migrationPassDir: ".cartulary/test-results/fixture/seaweedfs-migration-preservation/object-store-migration/pass",
  validation: {
    schema_id: "cartulary.object_store_migration_validation.v1",
    result: "pass",
    source_backend: "minio_s3",
    target_backend: "seaweedfs_s3",
    source_bucket: "fixture-bucket",
    target_bucket: "fixture-bucket",
    object_blob_count: 1,
    blocking_diagnostics: [],
    objects_checked: [
      {
        object_blob_id: "00000000-0000-4000-8000-000000000001",
        status: "pass",
        storage_ref_sha256: "a".repeat(64),
        source_size_bytes: 4,
        target_size_bytes: 4,
        source_sha256: "b".repeat(64),
        target_sha256: "b".repeat(64),
      },
    ],
  },
  copyLedger: {
    schema_id: "cartulary.object_store_migration_copy_ledger.v1",
    result: "pass",
    object_count: 1,
    items: [
      {
        object_blob_id: "00000000-0000-4000-8000-000000000001",
        status: "copied",
        source_bucket_ref: { sha256: "c".repeat(64) },
        target_bucket_ref: { sha256: "c".repeat(64) },
        source_key_ref: { sha256: "d".repeat(64) },
        target_key_ref: { sha256: "d".repeat(64) },
        source_size_bytes: 4,
        target_size_bytes: 4,
        source_sha256: "b".repeat(64),
        target_sha256: "b".repeat(64),
      },
    ],
  },
  migrationRun: {
    schema_id: "cartulary.object_store_migration_run.v1",
    current_state: "cutover_ready",
    events: [{ event: "validation_passed" }],
  },
});
assert.equal(migration.result, "pass");
assert.equal(migration.schema_id, "cartulary.seaweedfs_migration_preservation_evidence.v2");
assert.equal(migration.preservation_checks.bucket_preserved, true);
assert.equal(migration.preservation_checks.source_bucket, undefined);
assert.equal(migration.preservation_checks.target_bucket, undefined);
assert.equal(migration.preservation_checks.source_bucket_ref.redaction_class, "bucket");
assert.match(migration.preservation_checks.source_bucket_ref.sha256, /^[a-f0-9]{64}$/);

const missingMigration = buildMigrationPreservationEvidence({
  repoCommitValue: "abc123",
  generatedAt: "2026-06-04T00:00:00.000Z",
  migrationPassDir: ".cartulary/test-results/missing/seaweedfs-migration-preservation/object-store-migration/pass",
});
assert.equal(missingMigration.result, "fail");

const redactionMissingCurrent = buildRedactionLeakageScan({
  generatedAt: "2026-06-04T00:00:00.000Z",
  repoCommitValue: "abc123",
  selectedArtifactPaths: [".cartulary/release-artifacts/seaweedfs/fixture/migration-preservation-evidence.json"],
  compatibilityReportPath: currentCompatibilityReportPath,
  migrationPassDir: ".cartulary/test-results/fixture/seaweedfs-migration-preservation/object-store-migration/pass",
  requireBackendProcessArtifacts: true,
});
assert.equal(redactionMissingCurrent.result, "fail");
assert.equal(
  redactionMissingCurrent.scanned_artifacts.some(
    (artifact) => artifact.path === currentCompatibilityReportPath,
  ),
  true,
);
assert.equal(
  redactionMissingCurrent.scanned_artifacts.some((artifact) =>
    artifact.path.endsWith("/backend-process/backup-restore/object-store-backup-manifest.json"),
  ),
  true,
);
assert.equal(
  redactionMissingCurrent.scanned_artifacts.some((artifact) =>
    artifact.path.endsWith("/seaweedfs-migration-preservation/object-store-migration/mismatch/target-probe.json"),
  ),
  true,
);

const redactionFixtureDir = mkdtempSync(path.join(os.tmpdir(), "cartulary-redaction-fixture-"));
try {
  const rawPublicArtifact = path.join(redactionFixtureDir, "raw-public.json");
  writeFileSync(rawPublicArtifact, JSON.stringify({ source_bucket: "fixture-bucket" }) + "\n");
  const rawPublicScan = buildRedactionLeakageScan({
    generatedAt: "2026-06-04T00:00:00.000Z",
    repoCommitValue: "abc123",
    selectedArtifactPaths: [rawPublicArtifact],
  });
  assert.equal(rawPublicScan.result, "fail");
  assert.equal(
    rawPublicScan.findings.some((finding) => finding.check_id === "public-raw-storage-field"),
    true,
  );
} finally {
  rmSync(redactionFixtureDir, { recursive: true, force: true });
}

const pathMap = {
  "occurrence-inventory.json": ".cartulary/release-artifacts/seaweedfs/fixture/occurrence-inventory.json",
  "release-manifest-exposure.json": ".cartulary/release-artifacts/seaweedfs/fixture/release-manifest-exposure.json",
  "dependency-boundary.json": ".cartulary/release-artifacts/seaweedfs/fixture/dependency-boundary.json",
  "sbom-license-classification.json": ".cartulary/release-artifacts/seaweedfs/fixture/sbom-license-classification.json",
  "threat-model-coverage.json": ".cartulary/release-artifacts/seaweedfs/fixture/threat-model-coverage.json",
  "storage-ref-owner-coverage.json": ".cartulary/release-artifacts/seaweedfs/fixture/storage-ref-owner-coverage.json",
  "seaweedfs-compatibility-evidence.json": ".cartulary/release-artifacts/seaweedfs/fixture/seaweedfs-compatibility-evidence.json",
  "migration-preservation-evidence.json": ".cartulary/release-artifacts/seaweedfs/fixture/migration-preservation-evidence.json",
  "redaction-leakage-scan.json": ".cartulary/release-artifacts/seaweedfs/fixture/redaction-leakage-scan.json",
  "release-gate-summary.json": ".cartulary/release-artifacts/seaweedfs/fixture/release-gate-summary.json",
};

function assertSwfsAc024SummaryInvariant(summary) {
  const claim = summary.claims.find((entry) => entry.row === "SWFS-AC-024");
  assert.ok(claim);
  const claimable = claim.status === "claimable";
  const noBlockingRows = summary.blocking_rows.length === 0;
  const noBlockingChecks = summary.blocking_checks.length === 0;
  const aggregatePasses =
    summary.evidence_result === "pass" && summary.release_gate_result === "pass";
  assert.equal(claimable, aggregatePasses && noBlockingRows && noBlockingChecks);
}

const passingSummary = buildReleaseGateSummary({
  generatedAt: "2026-06-04T00:00:00.000Z",
  repoCommitValue: "abc123",
  pathMap,
  artifacts: {
    occurrence: { result: "pass" },
    exposure: { result: "pass" },
    dependency: { result: "pass" },
    sbomLicense: { result: "pass" },
    threat: { result: "pass" },
    storageRefOwner,
    compatibility,
    migration,
    redaction: { result: "pass" },
  },
});
assert.equal(passingSummary.release_gate_result, "pass");
assert.deepEqual(passingSummary.blocking_rows, []);
assert.deepEqual(passingSummary.blocking_checks, []);
assert.equal(passingSummary.claims.find((claim) => claim.row === "SWFS-AC-015")?.status, "claimable");
assert.equal(passingSummary.claims.find((claim) => claim.row === "SWFS-AC-018")?.status, "claimable");
assert.equal(passingSummary.claims.find((claim) => claim.row === "SWFS-AC-024")?.status, "claimable");
assert.equal(
  passingSummary.claims
    .find((claim) => claim.row === "SWFS-AC-024")
    ?.evidence_paths.includes(pathMap["redaction-leakage-scan.json"]),
  true,
);
assertSwfsAc024SummaryInvariant(passingSummary);

const redactionBlockedSummary = buildReleaseGateSummary({
  generatedAt: "2026-06-04T00:00:00.000Z",
  repoCommitValue: "abc123",
  pathMap,
  artifacts: {
    occurrence: { result: "pass" },
    exposure: { result: "pass" },
    dependency: { result: "pass" },
    sbomLicense: { result: "pass" },
    threat: { result: "pass" },
    storageRefOwner,
    compatibility,
    migration,
    redaction: { result: "fail" },
  },
});
assert.equal(redactionBlockedSummary.evidence_result, "fail");
assert.equal(redactionBlockedSummary.release_gate_result, "blocked");
assert.deepEqual(redactionBlockedSummary.blocking_rows, []);
assert.equal(redactionBlockedSummary.blocking_checks.length, 1);
assert.equal(redactionBlockedSummary.blocking_checks[0].check_id, "redaction-leakage-scan");
assert.equal(redactionBlockedSummary.blocking_checks[0].result, "fail");
assert.deepEqual(redactionBlockedSummary.blocking_checks[0].evidence_paths, [
  pathMap["redaction-leakage-scan.json"],
]);
assert.equal(redactionBlockedSummary.claims.find((claim) => claim.row === "SWFS-AC-024")?.status, "blocked");
assertSwfsAc024SummaryInvariant(redactionBlockedSummary);

const blockedSummary = buildReleaseGateSummary({
  generatedAt: "2026-06-04T00:00:00.000Z",
  repoCommitValue: "abc123",
  pathMap,
  artifacts: {
    occurrence: { result: "pass" },
    exposure: { result: "pass" },
    dependency: { result: "pass" },
    sbomLicense: { result: "pass" },
    threat: { result: "pass" },
    storageRefOwner,
    compatibility: partialCompatibility,
    migration: missingMigration,
    redaction: { result: "pass" },
  },
});
assert.equal(blockedSummary.release_gate_result, "blocked");
assert.deepEqual(new Set(blockedSummary.blocking_rows), new Set(["SWFS-AC-015", "SWFS-AC-018"]));
assert.deepEqual(blockedSummary.blocking_checks, []);
assertSwfsAc024SummaryInvariant(blockedSummary);
