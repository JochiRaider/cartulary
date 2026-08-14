#!/usr/bin/env node

import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import {
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";
import {
  ac043PredicateIDsForRows,
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
  if (!row || row.fixture_profile_id !== "ac043_large_grid_snapshot_v1") {
    throw new Error("measurement finalizer row lacks the active AC-043 fixture profile");
  }
  const expectedPredicates = ac043PredicateIDsForRows(root, [row]);
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
    "cartulary.performance_fixture_snapshot.v1",
  );
  const lease = readArtifact(
    base,
    path.join("performance-fixtures", snapshotKey, "leases", `${options.row}.json`),
    "cartulary.performance_fixture_snapshot_lease.v1",
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
  if (
    lease.artifact.creation_state !== "created" ||
    lease.artifact.isolation_result !== "isolated" ||
    lease.artifact.cleanup_state !== "complete" ||
    !lease.artifact.credential_copy_cleanup ||
    !lease.artifact.database_cleanup ||
    !lease.artifact.bucket_cleanup ||
    !lease.artifact.session_cleanup ||
    !lease.artifact.process_cleanup
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
    reportPaths: [report],
    runRoot: base,
  });
  const observation = observations[0];
  if (!observation) {
    throw new Error("measurement observation is missing");
  }
  const overlapCount = schedulerEvidence.overlap_count;
  if (overlapCount !== 0) {
    throw new Error(`measurement quiet interval overlapped ${overlapCount} ordinary units`);
  }
  const summary = {
    schema_id: "cartulary.frontend_measurement_summary.v2",
    row_id: options.row,
    observation,
    fixture_profile_id: profileID,
    snapshot_key: snapshotKey,
    build_artifact: artifactRef("performance_fixture_snapshot_build", build),
    lease_artifact: artifactRef("performance_fixture_snapshot_lease", lease),
    clone_ordinal: lease.artifact.clone_ordinal,
    isolation_result: lease.artifact.isolation_result,
    credential_copy_cleanup: lease.artifact.credential_copy_cleanup,
    database_cleanup: lease.artifact.database_cleanup,
    bucket_cleanup: lease.artifact.bucket_cleanup,
    scheduler_overlap_count: overlapCount,
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
    path.join(groupRelative, "frontend-measurement-summary.v2.json"),
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
