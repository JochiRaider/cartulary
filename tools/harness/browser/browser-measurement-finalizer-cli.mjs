#!/usr/bin/env node

import { createHash } from "node:crypto";
import { existsSync, lstatSync, readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import {
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  activePerformanceFixtureProfile,
  loadPerformanceFixtureSnapshotRegistry,
  performanceFixturePredicateIDsForRows,
  validateProfileMeasurementObservation,
} from "../performance-fixture/index.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";
import {
  collectFrontendMeasurementObservations,
  readMeasurementSchedulerEvidence,
} from "./frontend-measurement-evidence.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");
const forbiddenKey = /^(?:bucket_name|credential|credentials|database_name|dsn|email|entered_text|password|payload|record_id|runtime_path|secret|token|transaction_id|txn_id|user_id)$/iu;

function parse(argv) {
  const options = {
    target: "",
    stage: "",
    group: "",
    row: "",
    predicate: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") options.target = argv[++index] ?? "";
    else if (arg === "--stage") options.stage = argv[++index] ?? "";
    else if (arg === "--group") options.group = argv[++index] ?? "";
    else if (arg === "--row") options.row = argv[++index] ?? "";
    else if (arg === "--predicate") options.predicate = argv[++index] ?? "";
    else throw new Error("invalid browser measurement finalizer arguments");
  }
  if (Object.values(options).some((value) => value === "")) {
    throw new Error("browser measurement finalizer requires target, stage, group, row, and predicate");
  }
  return options;
}

function runRoot() {
  const results = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!results || !runID) {
    throw new Error("browser measurement finalizer requires result-root identity");
  }
  return path.resolve(root, results, runID);
}

function contained(base, relative) {
  const resolved = path.resolve(base, relative);
  const checked = path.relative(base, resolved);
  if (!checked || checked.startsWith("../") || path.isAbsolute(checked)) {
    throw new Error(`measurement artifact escapes run root: ${relative}`);
  }
  return resolved;
}

function readArtifact(base, relative, schemaID) {
  const file = contained(base, relative);
  if (!existsSync(file)) {
    throw new Error(`measurement artifact is missing: ${relative}`);
  }
  const stat = lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink() || stat.size > 8 * 1024 * 1024) {
    throw new Error(`measurement artifact exceeds its bounded JSON contract: ${relative}`);
  }
  const bytes = readFileSync(file);
  const artifact = JSON.parse(bytes.toString("utf8"));
  validateSchemaSync(schemaID, artifact);
  rejectSensitiveKeys(artifact, schemaID);
  return { artifact, bytes, file, relative };
}

function rejectSensitiveKeys(value, location) {
  if (Array.isArray(value)) {
    value.forEach((entry, index) => rejectSensitiveKeys(entry, `${location}[${index}]`));
    return;
  }
  if (!value || typeof value !== "object") return;
  for (const [key, entry] of Object.entries(value)) {
    if (forbiddenKey.test(key)) {
      throw new Error(`${location} contains forbidden key ${key}`);
    }
    rejectSensitiveKeys(entry, `${location}.${key}`);
  }
}

function digest(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function artifactRef(role, artifact) {
  return {
    role,
    path_kind: "file",
    format: "json",
    path: artifact.relative,
    sha256: digest(artifact.bytes),
  };
}

function cleanupOutcomes(lease) {
  const outcomes = new Map();
  for (const result of lease.cleanup_results) {
    if (outcomes.has(result.resource_class)) {
      throw new Error(`measurement lease duplicates cleanup class ${result.resource_class}`);
    }
    outcomes.set(result.resource_class, result.outcome);
  }
  return outcomes;
}

async function main() {
  const options = parse(process.argv.slice(2));
  const base = runRoot();
  const groupRelative = path.join(
    options.target,
    "browser-groups",
    options.group.replaceAll(/[^a-zA-Z0-9_.-]+/gu, "-"),
  );
  const groupResultRelative = path.join(
    groupRelative,
    "browser-group-result.json",
  );
  const schedulerEvidence = await readMeasurementSchedulerEvidence(
    path.join(base, "unit-events.ndjson"),
    { group_id: options.group, stage_id: options.stage },
  );
  if (!existsSync(contained(base, groupResultRelative))) {
    if (schedulerEvidence.dependency_skipped) return;
    throw new Error(`measurement artifact is missing: ${groupResultRelative}`);
  }
  if (schedulerEvidence.dependency_skipped) {
    throw new Error("measurement group result exists for a dependency-skipped scheduler unit");
  }
  const group = readArtifact(
    base,
    groupResultRelative,
    "cartulary.browser_group_result.v5",
  );
  if (
    group.artifact.target_id !== options.target ||
    group.artifact.stage_id !== options.stage ||
    group.artifact.group_id !== options.group ||
    JSON.stringify(group.artifact.selected_rows) !== JSON.stringify([options.row])
  ) {
    throw new Error("measurement group result identity is inconsistent");
  }
  const catalog = loadTestCatalog(root);
  const row = catalog.rowByID.get(options.row);
  if (!row || !row.fixture_profile_id) {
    throw new Error("measurement finalizer row lacks an active fixture profile");
  }
  const registry = loadPerformanceFixtureSnapshotRegistry(root);
  const profile = activePerformanceFixtureProfile(registry, row.fixture_profile_id);
  const expectedPredicates = performanceFixturePredicateIDsForRows(root, [row], { registry });
  if (
    expectedPredicates.length !== 1 ||
    expectedPredicates[0] !== options.predicate
  ) {
    throw new Error("measurement finalizer predicate routing is inconsistent");
  }
  const snapshotKey = group.artifact.snapshot_key;
  const profileID = group.artifact.fixture_profile_id;
  if (
    profileID !== row.fixture_profile_id ||
    !/^[a-f0-9]{64}$/u.test(snapshotKey)
  ) {
    throw new Error("measurement group snapshot identity is inconsistent");
  }
  const build = readArtifact(
    base,
    path.join("performance-fixtures", snapshotKey, "snapshot-build.json"),
    profile.artifact_policy.build_schema_id,
  );
  const lease = readArtifact(
    base,
    path.join("performance-fixtures", snapshotKey, "leases", `${options.row}.json`),
    profile.artifact_policy.lease_schema_id,
  );
  const builderUnitID = `fixture_snapshot:default:${profileID}:${snapshotKey}`;
  if (
    build.artifact.state !== "sealed" ||
    build.artifact.fixture_profile_id !== profileID ||
    build.artifact.snapshot_key !== snapshotKey ||
    build.artifact.builder_unit_id !== builderUnitID ||
    lease.artifact.fixture_profile_id !== profileID ||
    lease.artifact.snapshot_key !== snapshotKey ||
    lease.artifact.builder_unit_id !== builderUnitID ||
    lease.artifact.row_id !== options.row ||
    lease.artifact.predicate_id !== options.predicate
  ) {
    throw new Error("measurement build and lease provenance is inconsistent");
  }
  const cleanup = cleanupOutcomes(lease.artifact);
  if (
    lease.artifact.creation_state !== "created" ||
    lease.artifact.isolation_result !== "isolated" ||
    lease.artifact.cleanup_state !== "complete" ||
    cleanup.get("bucket") !== "complete" ||
    cleanup.get("credential_copy") !== "complete" ||
    cleanup.get("database") !== "complete" ||
    cleanup.get("process") !== "complete" ||
    cleanup.get("session") !== "complete"
  ) {
    throw new Error("measurement clone cleanup is incomplete");
  }
  if (
    Date.parse(build.artifact.created_at) > Date.parse(group.artifact.started_at) ||
    Date.parse(lease.artifact.finalized_at) < Date.parse(group.artifact.finished_at)
  ) {
    throw new Error("measurement artifact finalization order is stale");
  }
  const reportRelative = group.artifact.artifacts.playwright_report;
  const report = contained(base, reportRelative);
  const observations = collectFrontendMeasurementObservations({
    expectedPredicateIDs: [options.predicate],
    observationSchemaID: profile.artifact_policy.observation_schema_id,
    reportPaths: [report],
    runRoot: base,
  });
  const observation = observations[0];
  if (!observation) {
    throw new Error("measurement observation is missing");
  }
  validateProfileMeasurementObservation(root, profile, observation);
  const overlapCount = schedulerEvidence.overlap_count;
  if (overlapCount !== 0) {
    throw new Error(`measurement quiet interval overlapped ${overlapCount} ordinary units`);
  }
  const observationOutput = contained(
    base,
    path.join(
      groupRelative,
      `frontend-measurement-observation.${profile.artifact_policy.observation_schema_id.split(".").at(-1)}.json`,
    ),
  );
  if (existsSync(observationOutput)) {
    throw new Error(`measurement observation is immutable: ${observationOutput}`);
  }
  const observationBytes = Buffer.from(
    `${JSON.stringify(observation, null, 2)}\n`,
    "utf8",
  );
  secureWriteFile(observationOutput, observationBytes);
  const observationArtifact = {
    bytes: observationBytes,
    relative: path.relative(base, observationOutput).replaceAll("\\", "/"),
  };
  const summary = {
    schema_id: profile.artifact_policy.summary_schema_id,
    row_id: options.row,
    fixture_profile_id: profileID,
    snapshot_key: snapshotKey,
    observation_artifact: artifactRef(
      "frontend_measurement_observation",
      observationArtifact,
    ),
    build_artifact: artifactRef("performance_fixture_snapshot_build", build),
    lease_artifact: artifactRef("performance_fixture_snapshot_lease", lease),
    clone_ordinal: lease.artifact.clone_ordinal,
    scheduler_overlap_count: overlapCount,
    rollup: {
      predicate_id: observation.predicate_id,
      sample_count: observation.measured_samples,
      p50_ms: observation.p50_ms,
      p95_ms: observation.p95_ms,
      threshold_ms: observation.threshold_ms,
      outcome: observation.outcome,
    },
    qualification_outcome:
      observation.outcome === "passed"
        ? "qualified"
        : observation.outcome === "threshold_failed"
          ? "threshold_failed"
          : "environment_not_qualified",
  };
  validateSchemaSync(summary.schema_id, summary);
  rejectSensitiveKeys(summary, summary.schema_id);
  const output = contained(
    base,
    path.join(
      groupRelative,
      `frontend-measurement-summary.${profile.artifact_policy.summary_schema_id.split(".").at(-1)}.json`,
    ),
  );
  if (existsSync(output)) {
    throw new Error(`measurement summary is immutable: ${output}`);
  }
  secureWriteFile(output, `${JSON.stringify(summary, null, 2)}\n`);
}

main().catch((error) => {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 11;
});
