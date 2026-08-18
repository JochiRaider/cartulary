import { createHash } from "node:crypto";
import {
  chmodSync,
  closeSync,
  constants,
  existsSync,
  fstatSync,
  lstatSync,
  openSync,
  readFileSync,
  readdirSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { validateSchemaSync } from "../harness/contract/index.mjs";
import { secureMkdir } from "../harness/contract/artifact-writer.mjs";
import { validateCanonicalRun } from "../harness/observability/canonical-evidence.mjs";

const moduleRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const inputSchemaID = "cartulary.incident_bundle_v1_retirement_attestation.v1";
const resultSchemaID = "cartulary.incident_bundle_v1_retirement_attestation_result.v1";
const compatibilitySchemaID = "cartulary.incident_bundle_compatibility.v2";
const releaseCheckCommandID = "cartulary.harness.command.release_check.v2";
const maxAttestationBytes = 1024 * 1024;
const maxManifestBytes = 1024 * 1024;
const maxSummaryBytes = 16 * 1024 * 1024;
const maxRetainedRuns = 10_000;
const millisecondsPerDay = 24 * 60 * 60 * 1000;

export class RetirementAttestationError extends Error {
  constructor(code, message = code) {
    super(message);
    this.name = "RetirementAttestationError";
    this.code = code;
  }
}

function fail(code, message = code) {
  throw new RetirementAttestationError(code, message);
}

export function sha256Digest(value) {
  return `sha256:${createHash("sha256").update(value).digest("hex")}`;
}

function canonicalJSON(value) {
  return `${JSON.stringify(value)}\n`;
}

function safeLstat(file, code) {
  try {
    return lstatSync(file);
  } catch {
    fail(code);
  }
}

export function readBoundedRegularFile(file, maxBytes = maxAttestationBytes) {
  if (typeof file !== "string" || file === "" || file.includes("\0")) {
    fail("invalid_input_path");
  }
  const before = safeLstat(file, "input_not_found");
  if (
    !before.isFile() ||
    before.isSymbolicLink() ||
    before.size <= 0 ||
    before.size > maxBytes
  ) {
    fail("input_not_bounded_regular_file");
  }
  let descriptor;
  try {
    descriptor = openSync(file, constants.O_RDONLY | (constants.O_NOFOLLOW ?? 0));
    const opened = fstatSync(descriptor);
    if (
      !opened.isFile() ||
      opened.size !== before.size ||
      opened.dev !== before.dev ||
      opened.ino !== before.ino
    ) {
      fail("input_changed_during_read");
    }
    const content = readFileSync(descriptor);
    if (content.length !== opened.size || content.length > maxBytes) {
      fail("input_changed_during_read");
    }
    return content;
  } catch (error) {
    if (error instanceof RetirementAttestationError) throw error;
    fail("input_read_failed");
  } finally {
    if (descriptor !== undefined) closeSync(descriptor);
  }
}

function parseDate(value, label) {
  if (typeof value !== "string" || !/^[0-9]{4}-[0-9]{2}-[0-9]{2}$/u.test(value)) {
    fail("invalid_date", `${label} is not a UTC date`);
  }
  const instant = new Date(`${value}T00:00:00.000Z`);
  if (Number.isNaN(instant.valueOf()) || instant.toISOString().slice(0, 10) !== value) {
    fail("invalid_date", `${label} is not a real UTC date`);
  }
  return instant;
}

function dateString(instant) {
  return instant.toISOString().slice(0, 10);
}

function addDays(value, days) {
  const instant = parseDate(value, "date");
  instant.setUTCDate(instant.getUTCDate() + days);
  return dateString(instant);
}

function elapsedDays(start, end) {
  return Math.floor((parseDate(end, "end date") - parseDate(start, "start date")) / millisecondsPerDay);
}

function currentUTCDate(now = new Date()) {
  if (!(now instanceof Date) || Number.isNaN(now.valueOf())) fail("invalid_validator_clock");
  return dateString(now);
}

function semverTuple(identity) {
  const match = /^v?([0-9]+)\.([0-9]+)\.([0-9]+)$/u.exec(identity);
  if (!match) fail("unstable_release_identity");
  return match.slice(1).map((part) => BigInt(part));
}

function compareTuples(left, right) {
  for (let index = 0; index < left.length; index += 1) {
    if (left[index] < right[index]) return -1;
    if (left[index] > right[index]) return 1;
  }
  return 0;
}

function publicationProjection(release) {
  return {
    release_identity: release.release_identity,
    published_on: release.published_on,
    stable: release.stable,
    bundle_version: release.bundle_version,
    run_manifest_digest: release.run_manifest_digest,
    run_summary_digest: release.run_summary_digest,
    target_summary_digest: release.target_summary_digest,
    source_digest: release.source_digest,
  };
}

export function publicationEvidenceDigest(release) {
  return sha256Digest(canonicalJSON(publicationProjection(release)));
}

function readDigest(file, limit) {
  return sha256Digest(readBoundedRegularFile(file, limit));
}

function retainedResultsRoot(configured, root) {
  const token = configured || ".cartulary/test-results";
  if (token.includes("\0")) fail("unsafe_retained_results_root");
  const resolved = path.resolve(root, token);
  const info = safeLstat(resolved, "retained_results_root_missing");
  if (!info.isDirectory() || info.isSymbolicLink()) fail("unsafe_retained_results_root");
  return resolved;
}

export function resolveCompatibilitySnapshot(
  snapshot,
  { root = moduleRoot } = {},
) {
  const file = path.join(root, "contracts/incident-bundles/compatibility.json");
  const bytes = readBoundedRegularFile(file, maxAttestationBytes);
  if (sha256Digest(bytes) !== snapshot.snapshot_digest) {
    fail("compatibility_snapshot_digest_mismatch");
  }
  let compatibility;
  try {
    compatibility = JSON.parse(bytes.toString("utf8"));
    validateSchemaSync(compatibilitySchemaID, compatibility);
  } catch {
    fail("compatibility_snapshot_invalid");
  }
  const retained = compatibility.retained_imports[0];
  if (
    compatibility.release_lifecycle !== "stable_published" ||
    compatibility.backward_compatibility_required !== true ||
    retained.status !== "deprecated"
  ) {
    fail("compatibility_snapshot_not_operationally_gated");
  }
  if (
    compatibility.adopting_stable_release !==
      snapshot.adopting_release_identity ||
    retained.deprecation_started_on !== snapshot.deprecation_started_on ||
    retained.minimum_stable_releases !== snapshot.minimum_stable_releases ||
    retained.minimum_retention_days !== snapshot.minimum_retention_days ||
    retained.zero_successful_import_days !== 30 ||
    retained.operator_inventory_clear_required !== true
  ) {
    fail("compatibility_snapshot_semantics_mismatch");
  }
  return { snapshot_digest: snapshot.snapshot_digest };
}

function findRunRootByManifestDigest(expectedDigest, resultsRoot) {
  const entries = readdirSync(resultsRoot, { withFileTypes: true });
  if (entries.length > maxRetainedRuns) fail("retained_results_root_unbounded");
  const matches = [];
  for (const entry of entries) {
    if (!entry.isDirectory() || entry.isSymbolicLink()) continue;
    const runRoot = path.join(resultsRoot, entry.name);
    const manifest = path.join(runRoot, "run-manifest.json");
    if (!existsSync(manifest)) continue;
    const info = lstatSync(manifest);
    if (!info.isFile() || info.isSymbolicLink() || info.size <= 0 || info.size > maxManifestBytes) {
      continue;
    }
    if (readDigest(manifest, maxManifestBytes) === expectedDigest) matches.push(runRoot);
  }
  if (matches.length === 0) fail("retained_release_run_missing");
  if (matches.length !== 1) fail("retained_release_run_ambiguous");
  return matches[0];
}

export async function resolveRetainedReleaseEvidence(
  release,
  {
    root = moduleRoot,
    resultsDir = process.env.CARTULARY_TEST_RESULTS_DIR,
    canonicalRunValidator = validateCanonicalRun,
  } = {},
) {
  const resultsRoot = retainedResultsRoot(resultsDir, root);
  const runRoot = findRunRootByManifestDigest(release.run_manifest_digest, resultsRoot);
  const manifestPath = path.join(runRoot, "run-manifest.json");
  const summaryPath = path.join(runRoot, "run-summary.json");
  const targetSummaryPath = path.join(runRoot, "target-summaries", "release-check.json");
  if (readDigest(summaryPath, maxSummaryBytes) !== release.run_summary_digest) {
    fail("release_run_summary_digest_mismatch");
  }
  if (readDigest(targetSummaryPath, maxSummaryBytes) !== release.target_summary_digest) {
    fail("release_target_summary_digest_mismatch");
  }
  let retained;
  try {
    retained = await canonicalRunValidator(runRoot, "release-check");
  } catch {
    fail("release_run_not_canonical");
  }
  const targetSummary = retained.targetSummaries.get("release-check");
  if (
    retained.manifest.command_id !== releaseCheckCommandID ||
    retained.manifest.target !== "release-check" ||
    retained.manifest.source_state !== "clean" ||
    retained.manifest.source_digest !== release.source_digest ||
    retained.summary.status !== "pass" ||
    targetSummary?.status !== "pass"
  ) {
    fail("release_run_not_qualifying");
  }
  const startedOn = String(retained.manifest.started_at).slice(0, 10);
  parseDate(startedOn, "release run start");
  if (startedOn > release.published_on) fail("release_run_after_publication");
  return {
    run_key: retained.manifest.run_id,
    run_manifest_digest: readDigest(manifestPath, maxManifestBytes),
    run_summary_digest: release.run_summary_digest,
    target_summary_digest: release.target_summary_digest,
    source_digest: retained.manifest.source_digest,
  };
}

function assertInputSchema(document) {
  try {
    validateSchemaSync(inputSchemaID, document);
  } catch {
    fail("attestation_schema_invalid");
  }
}

function assertResultSchema(result) {
  try {
    validateSchemaSync(resultSchemaID, result);
  } catch {
    fail("attestation_result_invalid");
  }
}

export async function validateRetirementAttestationDocument(
  document,
  {
    inputDigest,
    now = new Date(),
    compatibilitySnapshotResolver = resolveCompatibilitySnapshot,
    releaseEvidenceResolver = resolveRetainedReleaseEvidence,
  } = {},
) {
  assertInputSchema(document);
  if (typeof inputDigest !== "string" || !/^sha256:[a-f0-9]{64}$/u.test(inputDigest)) {
    fail("input_digest_missing");
  }
  try {
    await compatibilitySnapshotResolver(document.compatibility_snapshot);
  } catch (error) {
    if (error instanceof RetirementAttestationError) throw error;
    fail("compatibility_snapshot_invalid");
  }
  const expectedRoles = ["adopting", "subsequent_1", "subsequent_2"];
  if (document.releases.some((release, index) => release.role !== expectedRoles[index])) {
    fail("release_roles_not_exact");
  }
  const identities = document.releases.map((release) => release.release_identity);
  const manifestDigests = document.releases.map((release) => release.run_manifest_digest);
  if (new Set(identities).size !== 3 || new Set(manifestDigests).size !== 3) {
    fail("release_evidence_not_distinct");
  }
  if (document.compatibility_snapshot.adopting_release_identity !== identities[0]) {
    fail("adopting_release_mismatch");
  }
  if (
    document.compatibility_snapshot.deprecation_started_on !==
    document.releases[0].published_on
  ) {
    fail("deprecation_start_mismatch");
  }
  for (let index = 1; index < document.releases.length; index += 1) {
    const previous = document.releases[index - 1];
    const current = document.releases[index];
    if (previous.published_on >= current.published_on) fail("release_dates_not_ordered");
    if (compareTuples(semverTuple(previous.release_identity), semverTuple(current.release_identity)) >= 0) {
      fail("release_identities_not_ordered");
    }
  }
  const cutoff = document.cutoff_date;
  parseDate(cutoff, "cutoff date");
  if (cutoff !== currentUTCDate(now)) fail("cutoff_not_current");
  if (document.releases[2].published_on > cutoff) fail("release_after_cutoff");
  const eligibilityDate = addDays(
    document.compatibility_snapshot.deprecation_started_on,
    document.compatibility_snapshot.minimum_retention_days,
  );
  const elapsed = elapsedDays(document.compatibility_snapshot.deprecation_started_on, cutoff);
  if (cutoff < eligibilityDate || elapsed < 180) fail("retention_period_incomplete");
  if (
    document.telemetry.interval_start !== addDays(cutoff, -30) ||
    document.telemetry.interval_end !== cutoff
  ) {
    fail("telemetry_interval_not_current");
  }
  if (document.inventory.cutoff_date !== cutoff) fail("inventory_cutoff_mismatch");
  if (
    document.inventory.inventory_classes_expected !==
    document.inventory.inventory_classes_covered
  ) {
    fail("inventory_coverage_incomplete");
  }
  const resolvedReleases = [];
  for (const release of document.releases) {
    if (publicationEvidenceDigest(release) !== release.publication_evidence_digest) {
      fail("publication_evidence_digest_mismatch");
    }
    const resolved = await releaseEvidenceResolver(release);
    for (const field of [
      "run_manifest_digest",
      "run_summary_digest",
      "target_summary_digest",
      "source_digest",
    ]) {
      if (resolved[field] !== release[field]) fail("retained_release_digest_mismatch");
    }
    resolvedReleases.push(resolved);
  }
  if (new Set(resolvedReleases.map((release) => release.run_key)).size !== 3) {
    fail("retained_release_runs_not_distinct");
  }
  const result = {
    schema_id: resultSchemaID,
    eligible: true,
    cutoff_date: cutoff,
    input_digest: inputDigest,
    compatibility_snapshot_digest: document.compatibility_snapshot.snapshot_digest,
    stable_release_count: 3,
    minimum_stable_releases: 3,
    eligibility_date: eligibilityDate,
    elapsed_days: elapsed,
    minimum_retention_days: 180,
    telemetry_interval_days: 30,
    successful_v1_imports: 0,
    inventory_classes_expected: document.inventory.inventory_classes_expected,
    inventory_classes_covered: document.inventory.inventory_classes_covered,
    v1_required_archives: 0,
    release_evidence: document.releases.map((release) => ({
      role: release.role,
      release_identity: release.release_identity,
      published_on: release.published_on,
      run_manifest_digest: release.run_manifest_digest,
      run_summary_digest: release.run_summary_digest,
      target_summary_digest: release.target_summary_digest,
      source_digest: release.source_digest,
      publication_evidence_digest: release.publication_evidence_digest,
    })),
    telemetry_evidence_digest: document.telemetry.evidence_digest,
    inventory_evidence_digest: document.inventory.evidence_digest,
    gates: [
      { gate_id: "IB-R3-GATE-001", state: "pass" },
      { gate_id: "IB-R3-GATE-002", state: "pass" },
      { gate_id: "IB-R3-GATE-003", state: "pass" },
      { gate_id: "IB-R3-GATE-004", state: "pass" },
      { gate_id: "IB-R3-GATE-005", state: "pass" },
    ],
  };
  assertResultSchema(result);
  return result;
}

export async function validateRetirementAttestationBytes(bytes, options = {}) {
  let document;
  try {
    document = JSON.parse(bytes.toString("utf8"));
  } catch {
    fail("attestation_json_invalid");
  }
  return validateRetirementAttestationDocument(document, {
    ...options,
    inputDigest: sha256Digest(bytes),
  });
}

export function writeRetirementAttestationResult(result, artifactDirectory) {
  if (!artifactDirectory || artifactDirectory.includes("\0")) fail("artifact_directory_missing");
  const directory = path.resolve(artifactDirectory);
  secureMkdir(directory);
  const destination = path.join(directory, "retirement-attestation-result.json");
  const temporary = path.join(directory, `.retirement-attestation-result.${process.pid}.tmp`);
  const bytes = Buffer.from(`${JSON.stringify(result, null, 2)}\n`, "utf8");
  try {
    writeFileSync(temporary, bytes, { flag: "wx", mode: 0o600 });
    renameSync(temporary, destination);
    chmodSync(destination, 0o600);
  } finally {
    if (existsSync(temporary)) rmSync(temporary, { force: true });
  }
  return { digest: sha256Digest(bytes), file: destination };
}
