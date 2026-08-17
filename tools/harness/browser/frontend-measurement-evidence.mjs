import { createHash } from "node:crypto";
import { existsSync, lstatSync, readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";

import { validateSchemaSync } from "../contract/index.mjs";
import { readCanonicalUnitEvents } from "../evidence-accounting/canonical-unit-events.mjs";

const forbiddenKey = /^(?:bucket_name|credential|credentials|database_name|dsn|email|entered_text|password|payload|record_id|runtime_path|secret|token|transaction_id|txn_id|user_id)$/iu;
const reportByteLimit = 32 * 1024 * 1024;
const attachmentByteLimit = 8 * 1024 * 1024;

export function currentUnitEventFile(runRoot) {
  const configured = String(process.env.CARTULARY_HARNESS_LIVE_UNIT_EVENTS_FILE ?? "").trim();
  if (!configured) return path.join(runRoot, "unit-events.ndjson");
  if (!path.isAbsolute(configured)) {
    throw new Error("live unit-event staging path must be absolute");
  }
  const resolved = path.resolve(configured);
  const relative = path.relative(runRoot, resolved);
  if (
    !relative || relative.startsWith("..") || path.isAbsolute(relative) ||
    !path.basename(resolved).startsWith("unit-events.ndjson.tmp-")
  ) {
    throw new Error("live unit-event staging path escapes its run or has an invalid identity");
  }
  if (!existsSync(resolved)) {
    throw new Error("live unit-event staging path is missing");
  }
  const stat = lstatSync(resolved);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error("live unit-event staging path is not a regular file");
  }
  return resolved;
}

function readBoundedJSON(file, label, maxBytes) {
  const stat = lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink() || stat.size > maxBytes) {
    throw new Error(`${label} must be a non-symlink regular file of at most ${maxBytes} bytes`);
  }
  return JSON.parse(readFileSync(file, "utf8"));
}

function flattenSuites(suites, output = []) {
  for (const suite of suites ?? []) {
    output.push(...(suite.specs ?? []));
    flattenSuites(suite.suites, output);
  }
  return output;
}

function containedAttachment(runRoot, reportPath, attachmentPath) {
  const resolved = path.resolve(
    path.isAbsolute(attachmentPath) ? attachmentPath : path.dirname(reportPath),
    path.isAbsolute(attachmentPath) ? "" : attachmentPath,
  );
  const relative = path.relative(runRoot, resolved);
  if (!relative || relative.startsWith("..") || path.isAbsolute(relative)) {
    throw new Error(`measurement attachment escapes run root: ${attachmentPath}`);
  }
  if (!existsSync(resolved)) {
    throw new Error(`measurement attachment is missing: ${relative}`);
  }
  return resolved;
}

function parseAttachment(runRoot, reportPath, attachment) {
  if (typeof attachment.path === "string" && attachment.path !== "") {
    return readBoundedJSON(
      containedAttachment(runRoot, reportPath, attachment.path),
      "measurement attachment",
      attachmentByteLimit,
    );
  }
  if (typeof attachment.body !== "string" || attachment.body === "") {
    throw new Error(`${attachment.name} has neither a path nor a body`);
  }
  const bytes = Buffer.from(attachment.body, "base64");
  if (bytes.length > attachmentByteLimit) {
    throw new Error(`measurement attachment exceeds ${attachmentByteLimit} bytes`);
  }
  return JSON.parse(bytes.toString("utf8"));
}

function rejectSensitiveKeys(value, location = "summary") {
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

function validateSampleCardinality(summary) {
  const warmupSamples = summary.samples.filter((sample) => sample.warmup).length;
  const measuredSamples = summary.samples.length - warmupSamples;
  if (
    warmupSamples !== summary.warmup_samples ||
    measuredSamples !== summary.measured_samples
  ) {
    throw new Error(
      `${summary.predicate_id} sample counts differ from retained samples`,
    );
  }
  summary.samples.forEach((sample, index) => {
    if (
      sample.sample_index !== index ||
      sample.warmup !== (index < summary.warmup_samples)
    ) {
      throw new Error(
        `${summary.predicate_id} has a non-contiguous sample sequence`,
      );
    }
  });
}

export function collectFrontendMeasurementObservations({
  expectedPredicateIDs,
  observationSchemaID,
  reportPaths,
  runRoot,
}) {
  const attachmentPrefix = `${observationSchemaID}.`;
  const summaries = [];
  for (const reportPath of reportPaths) {
    const report = readBoundedJSON(reportPath, "Playwright JSON report", reportByteLimit);
    for (const spec of flattenSuites(report.suites)) {
      for (const playwrightTest of spec.tests ?? []) {
        for (const result of playwrightTest.results ?? []) {
          for (const attachment of result.attachments ?? []) {
            if (
              typeof attachment?.name !== "string" ||
              !attachment.name.startsWith(attachmentPrefix)
            ) {
              continue;
            }
            const summary = parseAttachment(runRoot, reportPath, attachment);
            validateSchemaSync(observationSchemaID, summary);
            rejectSensitiveKeys(summary);
            validateSampleCardinality(summary);
            summaries.push(summary);
          }
        }
      }
    }
  }
  const actualPredicateIDs = summaries
    .map((summary) => summary.predicate_id)
    .sort();
  const expected = [...expectedPredicateIDs].sort();
  if (JSON.stringify(actualPredicateIDs) !== JSON.stringify(expected)) {
    throw new Error(
      `frontend measurement observations differ: expected=${expected.join(",")} actual=${actualPredicateIDs.join(",")}`,
    );
  }
  return summaries.sort((left, right) =>
    left.predicate_id.localeCompare(right.predicate_id),
  );
}

function digest(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function readReferencedArtifact(runRoot, reference, schemaID, expectedRole) {
  if (
    reference?.role !== expectedRole ||
    reference?.path_kind !== "file" ||
    reference?.format !== "json" ||
    typeof reference.path !== "string" ||
    typeof reference.sha256 !== "string"
  ) {
    throw new Error("measurement artifact reference is malformed");
  }
  const file = containedAttachment(runRoot, path.join(runRoot, "reference.json"), reference.path);
  const stat = lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink() || stat.size > attachmentByteLimit) {
    throw new Error(`measurement artifact is not a bounded regular file: ${reference.path}`);
  }
  const bytes = readFileSync(file);
  if (digest(bytes) !== reference.sha256) {
    throw new Error(`measurement artifact digest or size is invalid: ${reference.path}`);
  }
  const artifact = JSON.parse(bytes.toString("utf8"));
  validateSchemaSync(schemaID, artifact);
  rejectSensitiveKeys(artifact, schemaID);
  return { artifact, bytes, file, reference };
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

export function collectFinalizedMeasurementSummaries({
  buildSchemaID,
  expectedPredicateIDs,
  leaseSchemaID,
  observationSchemaID,
  runRoot,
  summarySchemaID,
  summaryPaths,
  validateObservation,
}) {
  if (new Set(summaryPaths).size !== summaryPaths.length) {
    throw new Error("finalized measurement summary paths are duplicated");
  }
  const summaries = summaryPaths.map((summaryPath) => {
    if (!existsSync(summaryPath)) {
      throw new Error("finalized measurement summary is missing");
    }
    const stat = lstatSync(summaryPath);
    if (!stat.isFile() || stat.isSymbolicLink() || stat.size > attachmentByteLimit) {
      throw new Error("measurement summary is not a bounded regular file");
    }
    const bytes = readFileSync(summaryPath);
    const summary = JSON.parse(bytes.toString("utf8"));
    validateSchemaSync(summarySchemaID, summary);
    rejectSensitiveKeys(summary);
    const observation = readReferencedArtifact(
      runRoot,
      summary.observation_artifact,
      observationSchemaID,
      "frontend_measurement_observation",
    );
    const build = readReferencedArtifact(
      runRoot,
      summary.build_artifact,
      buildSchemaID,
      "performance_fixture_snapshot_build",
    );
    const lease = readReferencedArtifact(
      runRoot,
      summary.lease_artifact,
      leaseSchemaID,
      "performance_fixture_snapshot_lease",
    );
    const cleanup = cleanupOutcomes(lease.artifact);
    validateSampleCardinality(observation.artifact);
    validateObservation?.(observation.artifact);
    if (
      summary.scheduler_overlap_count !== 0 ||
      build.artifact.state !== "sealed" ||
      lease.artifact.isolation_result !== "isolated" ||
      lease.artifact.cleanup_state !== "complete" ||
      cleanup.get("bucket") !== "complete" ||
      cleanup.get("credential_copy") !== "complete" ||
      cleanup.get("database") !== "complete" ||
      cleanup.get("process") !== "complete" ||
      cleanup.get("session") !== "complete"
    ) {
      throw new Error(observation.artifact.predicate_id + " is environment_not_qualified");
    }
    const expectedQualification =
      observation.artifact.outcome === "passed"
        ? "qualified"
        : observation.artifact.outcome === "threshold_failed"
          ? "threshold_failed"
          : "environment_not_qualified";
    if (summary.qualification_outcome !== expectedQualification) {
      throw new Error(
        observation.artifact.predicate_id + " has inconsistent qualification outcome",
      );
    }
    if (expectedQualification === "environment_not_qualified") {
      throw new Error(
        observation.artifact.predicate_id + " is not eligible for active qualification",
      );
    }
    if (
      summary.fixture_profile_id !== observation.artifact.fixture_profile_id ||
      summary.snapshot_key !== build.artifact.snapshot_key ||
      summary.snapshot_key !== lease.artifact.snapshot_key ||
      summary.fixture_profile_id !== build.artifact.fixture_profile_id ||
      summary.fixture_profile_id !== lease.artifact.fixture_profile_id ||
      summary.row_id !== lease.artifact.row_id ||
      summary.clone_ordinal !== lease.artifact.clone_ordinal ||
      summary.rollup.predicate_id !== observation.artifact.predicate_id ||
      summary.rollup.sample_count !== observation.artifact.measured_samples ||
      summary.rollup.p50_ms !== observation.artifact.p50_ms ||
      summary.rollup.p95_ms !== observation.artifact.p95_ms ||
      summary.rollup.threshold_ms !== observation.artifact.threshold_ms ||
      summary.rollup.outcome !== observation.artifact.outcome
    ) {
      throw new Error("measurement summary referenced provenance is inconsistent");
    }
    return { build, bytes, file: summaryPath, lease, observation, summary };
  });
  const actualPredicateIDs = summaries
    .map((entry) => entry.observation.artifact.predicate_id)
    .sort();
  const expected = [...expectedPredicateIDs].sort();
  if (JSON.stringify(actualPredicateIDs) !== JSON.stringify(expected)) {
    throw new Error(
      "frontend measurement summaries differ: expected=" + expected.join(",") +
      " actual=" + actualPredicateIDs.join(","),
    );
  }
  return summaries.sort((left, right) =>
    left.observation.artifact.predicate_id.localeCompare(
      right.observation.artifact.predicate_id,
    ),
  );
}

export async function readMeasurementSchedulerEvidenceForGroups(eventFile, groupResults) {
  if (!existsSync(eventFile)) {
    throw new Error("measurement finalizer requires unit-events.ndjson");
  }
  const states = new Map(groupResults.map((groupResult) => [
    `browser_group:${groupResult.stage_id}:${groupResult.group_id}`,
    {
      dependency_skipped: false,
      end_seq: null,
      group_id: groupResult.group_id,
      overlap_count: 0,
      start_seq: null,
    },
  ]));
  if (states.size !== groupResults.length) {
    throw new Error("measurement scheduler groups are duplicated");
  }
  for await (const event of readCanonicalUnitEvents(eventFile)) {
    const state = states.get(event.unit_id);
    if (state !== undefined) {
        if (event.event === "started") {
          if (state.start_seq !== null) {
            throw new Error(`measurement group ${state.group_id} starts more than once`);
          }
          state.start_seq = event.seq;
          continue;
        }
        if (event.event === "skipped" && event.failure_reason === "dependency_failure") {
          if (state.start_seq !== null) {
            throw new Error(`measurement group ${state.group_id} is skipped after starting`);
          }
          state.dependency_skipped = true;
          state.end_seq = event.seq;
          continue;
        }
        if (["completed", "failed", "cancelled"].includes(event.event)) {
          if (state.start_seq === null) {
            throw new Error(`measurement group ${state.group_id} terminates before starting`);
          }
          if (state.end_seq !== null) {
            throw new Error(`measurement group ${state.group_id} terminates more than once`);
          }
          state.end_seq = event.seq;
          continue;
        }
    }
    if (event.event === "started") {
      for (const [unitID, active] of states) {
        if (event.unit_id !== unitID && active.start_seq !== null && active.end_seq === null) {
          active.overlap_count += 1;
        }
      }
    }
  }
  for (const state of states.values()) {
    if (state.end_seq === null) {
      throw new Error(`measurement group ${state.group_id} lacks a closed scheduler interval`);
    }
  }
  return states;
}

export async function readMeasurementSchedulerEvidence(eventFile, groupResult) {
  const states = await readMeasurementSchedulerEvidenceForGroups(eventFile, [groupResult]);
  return states.get(`browser_group:${groupResult.stage_id}:${groupResult.group_id}`);
}
