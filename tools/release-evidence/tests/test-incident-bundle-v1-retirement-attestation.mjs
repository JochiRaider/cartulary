#!/usr/bin/env node

import assert from "node:assert/strict";
import {
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { retirementAttestationSummaryExtension } from "../../harness/output/test-output/step-artifacts.mjs";

import {
  publicationEvidenceDigest,
  readBoundedRegularFile,
  resolveCompatibilitySnapshot,
  resolveRetainedReleaseEvidence,
  RetirementAttestationError,
  sha256Digest,
  validateRetirementAttestationBytes,
  validateRetirementAttestationDocument,
  writeRetirementAttestationResult,
} from "../incident-bundle-v1-retirement-attestation.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const fixturePath = path.join(
  root,
  "tools/release-evidence/fixtures/incident-bundle-v1-retirement-attestation.valid.json",
);
const fixtureBytes = readFileSync(fixturePath);
const fixture = JSON.parse(fixtureBytes.toString("utf8"));
const inputDigest = sha256Digest(fixtureBytes);
const validationClock = new Date("2025-07-10T12:00:00.000Z");

function clone(value = fixture) {
  return structuredClone(value);
}

function qualifyingResolver(release, runKey = release.run_manifest_digest) {
  return Promise.resolve({
    run_key: runKey,
    run_manifest_digest: release.run_manifest_digest,
    run_summary_digest: release.run_summary_digest,
    target_summary_digest: release.target_summary_digest,
    source_digest: release.source_digest,
  });
}

function qualifyingCompatibilityResolver(snapshot) {
  return Promise.resolve({ snapshot_digest: snapshot.snapshot_digest });
}

async function validate(document, options = {}) {
  return validateRetirementAttestationDocument(document, {
    inputDigest,
    now: validationClock,
    compatibilitySnapshotResolver: qualifyingCompatibilityResolver,
    releaseEvidenceResolver: qualifyingResolver,
    ...options,
  });
}

async function assertRejected(name, mutate, expectedCode = "") {
  const document = clone();
  mutate(document);
  await assert.rejects(
    validate(document),
    (error) => {
      assert.ok(error instanceof RetirementAttestationError, `${name}: typed failure`);
      if (expectedCode) assert.equal(error.code, expectedCode, name);
      return true;
    },
    name,
  );
}

const positive = await validateRetirementAttestationBytes(fixtureBytes, {
  now: validationClock,
  compatibilitySnapshotResolver: qualifyingCompatibilityResolver,
  releaseEvidenceResolver: qualifyingResolver,
});
assert.equal(positive.schema_id, "cartulary.incident_bundle_v1_retirement_attestation_result.v1");
assert.equal(positive.eligible, true);
assert.equal(positive.stable_release_count, 3);
assert.equal(positive.minimum_stable_releases, 3);
assert.equal(positive.elapsed_days, 190);
assert.equal(positive.eligibility_date, "2025-06-30");
assert.deepEqual(positive.gates.map((gate) => gate.state), Array(5).fill("pass"));
assert.equal(Object.hasOwn(positive, "result_digest"), false);
assert.equal(Object.hasOwn(positive, "digest"), false);
assert.equal(JSON.stringify(positive), JSON.stringify(await validate(clone())));

await assertRejected("missing subsequent release", (document) => document.releases.pop());
await assertRejected("duplicate release identity", (document) => {
  document.releases[1].release_identity = document.releases[0].release_identity;
}, "release_evidence_not_distinct");
await assertRejected("duplicate retained run", (document) => {
  document.releases[1].run_manifest_digest = document.releases[0].run_manifest_digest;
}, "release_evidence_not_distinct");
await assertRejected("prerelease identity", (document) => {
  document.releases[1].release_identity = "v2.1.0-rc.1";
});
await assertRejected("unstable release", (document) => {
  document.releases[1].stable = false;
});
await assertRejected("non-v2 release", (document) => {
  document.releases[1].bundle_version = 1;
});
await assertRejected("release date ordering", (document) => {
  document.releases[1].published_on = document.releases[0].published_on;
}, "release_dates_not_ordered");
await assertRejected("release identity ordering", (document) => {
  document.releases[1].release_identity = "v1.9.0";
}, "release_identities_not_ordered");
await assertRejected("unpublished release after cutoff", (document) => {
  document.releases[2].published_on = "2025-08-01";
}, "release_after_cutoff");
await assertRejected("invalid calendar date", (document) => {
  document.cutoff_date = "2025-02-30";
}, "invalid_date");
await assert.rejects(
  validateRetirementAttestationDocument(clone(), {
    inputDigest,
    now: new Date("2025-07-11T00:00:00.000Z"),
    compatibilitySnapshotResolver: qualifyingCompatibilityResolver,
    releaseEvidenceResolver: qualifyingResolver,
  }),
  (error) => error instanceof RetirementAttestationError && error.code === "cutoff_not_current",
);

const shortRetention = clone();
shortRetention.cutoff_date = "2025-06-01";
shortRetention.telemetry.interval_start = "2025-05-02";
shortRetention.telemetry.interval_end = "2025-06-01";
shortRetention.inventory.cutoff_date = "2025-06-01";
await assert.rejects(
  validateRetirementAttestationDocument(shortRetention, {
    inputDigest,
    now: new Date("2025-06-01T12:00:00.000Z"),
    compatibilitySnapshotResolver: qualifyingCompatibilityResolver,
    releaseEvidenceResolver: qualifyingResolver,
  }),
  (error) => error instanceof RetirementAttestationError && error.code === "retention_period_incomplete",
);

await assertRejected("short telemetry", (document) => {
  document.telemetry.interval_days = 29;
});
await assertRejected("gapped telemetry", (document) => {
  document.telemetry.missing_intervals = 1;
});
await assertRejected("nonzero telemetry", (document) => {
  document.telemetry.successful_v1_imports = 1;
});
await assertRejected("stale telemetry interval", (document) => {
  document.telemetry.interval_start = "2025-06-09";
}, "telemetry_interval_not_current");
await assertRejected("incomplete inventory", (document) => {
  document.inventory.complete = false;
});
await assertRejected("partial inventory coverage", (document) => {
  document.inventory.inventory_classes_covered = 3;
}, "inventory_coverage_incomplete");
await assertRejected("nonzero inventory", (document) => {
  document.inventory.v1_required_archives = 1;
});
await assertRejected("inventory cutoff mismatch", (document) => {
  document.inventory.cutoff_date = "2025-07-09";
}, "inventory_cutoff_mismatch");
await assertRejected("unknown top-level field", (document) => {
  document.manual_pass = true;
});
for (const field of [
  "incident_id",
  "actor_id",
  "job_id",
  "object_key",
  "archive_name",
  "tenant_id",
  "credential",
  "storage_path",
  "raw_series",
]) {
  await assertRejected(`sensitive field ${field}`, (document) => {
    document.telemetry[field] = "forbidden";
  });
}
await assertRejected("malformed digest", (document) => {
  document.telemetry.evidence_digest = "SHA256:ABC";
});
await assertRejected("publication digest mismatch", (document) => {
  document.releases[0].publication_evidence_digest =
    "sha256:0000000000000000000000000000000000000000000000000000000000000000";
}, "publication_evidence_digest_mismatch");
const missingRetained = clone();
await assert.rejects(
  validateRetirementAttestationDocument(missingRetained, {
    inputDigest,
    now: validationClock,
    compatibilitySnapshotResolver: qualifyingCompatibilityResolver,
    releaseEvidenceResolver: async () => {
      throw new RetirementAttestationError("retained_release_run_missing");
    },
  }),
  (error) => error instanceof RetirementAttestationError && error.code === "retained_release_run_missing",
);

const repeatedResolvedRun = clone();
await assert.rejects(
  validateRetirementAttestationDocument(repeatedResolvedRun, {
    inputDigest,
    now: validationClock,
    compatibilitySnapshotResolver: qualifyingCompatibilityResolver,
    releaseEvidenceResolver: (release) => qualifyingResolver(release, "same-run"),
  }),
  (error) => error instanceof RetirementAttestationError && error.code === "retained_release_runs_not_distinct",
);

for (const release of fixture.releases) {
  assert.equal(publicationEvidenceDigest(release), release.publication_evidence_digest);
}

const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-retirement-attestation-"));
try {
  const regular = path.join(scratch, "input.json");
  writeFileSync(regular, fixtureBytes, { mode: 0o600 });
  assert.deepEqual(readBoundedRegularFile(regular), fixtureBytes);
  const symlink = path.join(scratch, "input-link.json");
  symlinkSync(regular, symlink);
  assert.throws(
    () => readBoundedRegularFile(symlink),
    (error) => error instanceof RetirementAttestationError && error.code === "input_not_bounded_regular_file",
  );
  const oversized = path.join(scratch, "oversized.json");
  writeFileSync(oversized, Buffer.alloc(65, "x"), { mode: 0o600 });
  assert.throws(
    () => readBoundedRegularFile(oversized, 64),
    (error) => error instanceof RetirementAttestationError && error.code === "input_not_bounded_regular_file",
  );
  assert.throws(
    () => readBoundedRegularFile(undefined),
    (error) => error instanceof RetirementAttestationError && error.code === "invalid_input_path",
  );
  const artifactDirectory = path.join(scratch, "artifacts");
  const first = writeRetirementAttestationResult(positive, artifactDirectory);
  const firstBytes = readFileSync(first.file);
  const second = writeRetirementAttestationResult(positive, artifactDirectory);
  assert.equal(second.digest, first.digest);
  assert.deepEqual(readFileSync(second.file), firstBytes);
  assert.equal(JSON.parse(firstBytes).result_digest, undefined);
  assert.deepEqual(retirementAttestationSummaryExtension(first.file), {
    input_digest: positive.input_digest,
    result_digest: first.digest,
  });

  const compatibility = {
    schema_id: "cartulary.incident_bundle_compatibility.v2",
    current_export_version: 2,
    release_lifecycle: "stable_published",
    backward_compatibility_required: true,
    adopting_stable_release: "v2.0.0",
    retained_imports: [
      {
        surface_id: "incident_bundle_v1_import",
        bundle_version: 1,
        owner_id: "module.incidentbundles",
        status: "deprecated",
        export_allowed: false,
        deprecation_started_on: "2025-01-01",
        minimum_stable_releases: 3,
        minimum_retention_days: 180,
        zero_successful_import_days: 30,
        operator_inventory_clear_required: true,
        telemetry_event: "cartulary.incident_bundle.v1_import",
        removal_trigger: "test retirement trigger",
      },
    ],
  };
  const compatibilityBytes = Buffer.from(`${JSON.stringify(compatibility, null, 2)}\n`);
  const compatibilityDirectory = path.join(scratch, "contracts/incident-bundles");
  mkdirSync(compatibilityDirectory, { recursive: true });
  writeFileSync(path.join(compatibilityDirectory, "compatibility.json"), compatibilityBytes);
  const snapshot = {
    ...fixture.compatibility_snapshot,
    snapshot_digest: sha256Digest(compatibilityBytes),
  };
  assert.deepEqual(resolveCompatibilitySnapshot(snapshot, { root: scratch }), {
    snapshot_digest: snapshot.snapshot_digest,
  });
  const preproductionCompatibility = {
    ...compatibility,
    release_lifecycle: "preproduction_unreleased",
    backward_compatibility_required: false,
    adopting_stable_release: null,
    retained_imports: [
      {
        ...compatibility.retained_imports[0],
        status: "development_only",
        deprecation_started_on: null,
      },
    ],
  };
  const preproductionBytes = Buffer.from(
    `${JSON.stringify(preproductionCompatibility, null, 2)}\n`,
  );
  writeFileSync(
    path.join(compatibilityDirectory, "compatibility.json"),
    preproductionBytes,
  );
  assert.throws(
    () =>
      resolveCompatibilitySnapshot(
        {
          ...snapshot,
          snapshot_digest: sha256Digest(preproductionBytes),
        },
        { root: scratch },
      ),
    (error) =>
      error instanceof RetirementAttestationError &&
      error.code === "compatibility_snapshot_not_operationally_gated",
  );
  writeFileSync(path.join(compatibilityDirectory, "compatibility.json"), compatibilityBytes);
  assert.throws(
    () =>
      resolveCompatibilitySnapshot(
        {
          ...snapshot,
          snapshot_digest:
            "sha256:0000000000000000000000000000000000000000000000000000000000000000",
        },
        { root: scratch },
      ),
    (error) =>
      error instanceof RetirementAttestationError &&
      error.code === "compatibility_snapshot_digest_mismatch",
  );

  const retainedRoot = path.join(scratch, "retained-results");
  const releaseRun = path.join(retainedRoot, "release-run-001");
  const targetSummaryDirectory = path.join(releaseRun, "target-summaries");
  mkdirSync(targetSummaryDirectory, { recursive: true });
  const releaseManifestBytes = Buffer.from('{"manifest":"release-run-001"}\n');
  const releaseSummaryBytes = Buffer.from('{"summary":"pass"}\n');
  const releaseTargetBytes = Buffer.from('{"target":"release-check","status":"pass"}\n');
  writeFileSync(path.join(releaseRun, "run-manifest.json"), releaseManifestBytes);
  writeFileSync(path.join(releaseRun, "run-summary.json"), releaseSummaryBytes);
  writeFileSync(
    path.join(targetSummaryDirectory, "release-check.json"),
    releaseTargetBytes,
  );
  const retainedRelease = {
    ...fixture.releases[0],
    run_manifest_digest: sha256Digest(releaseManifestBytes),
    run_summary_digest: sha256Digest(releaseSummaryBytes),
    target_summary_digest: sha256Digest(releaseTargetBytes),
  };
  const canonicalRunValidator = async () => ({
    manifest: {
      command_id: "cartulary.harness.command.release_check.v2",
      target: "release-check",
      source_state: "clean",
      source_digest: retainedRelease.source_digest,
      run_id: "release-run-001",
      started_at: "2024-12-31T23:00:00.000Z",
    },
    summary: { status: "pass" },
    targetSummaries: new Map([["release-check", { status: "pass" }]]),
  });
  assert.deepEqual(
    await resolveRetainedReleaseEvidence(retainedRelease, {
      root: scratch,
      resultsDir: retainedRoot,
      canonicalRunValidator,
    }),
    {
      run_key: "release-run-001",
      run_manifest_digest: retainedRelease.run_manifest_digest,
      run_summary_digest: retainedRelease.run_summary_digest,
      target_summary_digest: retainedRelease.target_summary_digest,
      source_digest: retainedRelease.source_digest,
    },
  );
  await assert.rejects(
    resolveRetainedReleaseEvidence(retainedRelease, {
      root: scratch,
      resultsDir: retainedRoot,
      canonicalRunValidator: async () => {
        throw new Error("noncanonical release run");
      },
    }),
    (error) =>
      error instanceof RetirementAttestationError &&
      error.code === "release_run_not_canonical",
  );
} finally {
  rmSync(scratch, { recursive: true, force: true });
}

process.stdout.write("incident bundle v1 retirement attestation contract checks passed\n");
