import { createReadStream, existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { createInterface } from "node:readline";

import { validateSchemaSync } from "../contract/index.mjs";
import { loadPerformanceFixtureSnapshotRegistry } from "../performance-fixture/index.mjs";

const attachmentPrefix = "cartulary.frontend_measurement_observation.v1.";
const forbiddenKey = /^(?:bucket_name|credential|credentials|database_name|dsn|email|entered_text|password|payload|record_id|runtime_path|secret|token|transaction_id|txn_id|user_id)$/iu;

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
    return JSON.parse(
      readFileSync(containedAttachment(runRoot, reportPath, attachment.path), "utf8"),
    );
  }
  if (typeof attachment.body !== "string" || attachment.body === "") {
    throw new Error(`${attachment.name} has neither a path nor a body`);
  }
  return JSON.parse(Buffer.from(attachment.body, "base64").toString("utf8"));
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
  reportPaths,
  runRoot,
}) {
  const summaries = [];
  for (const reportPath of reportPaths) {
    const report = JSON.parse(readFileSync(reportPath, "utf8"));
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
            validateSchemaSync("cartulary.frontend_measurement_observation.v1", summary);
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

export function collectFinalizedMeasurementSummaries({
  expectedPredicateIDs,
  summaryPaths,
}) {
  if (new Set(summaryPaths).size !== summaryPaths.length) {
    throw new Error("finalized measurement summary paths are duplicated");
  }
  const summaries = summaryPaths.map((summaryPath) => {
    if (!existsSync(summaryPath)) {
      throw new Error("finalized measurement summary is missing");
    }
    const summary = JSON.parse(readFileSync(summaryPath, "utf8"));
    validateSchemaSync("cartulary.frontend_measurement_summary.v2", summary);
    rejectSensitiveKeys(summary);
    validateSampleCardinality(summary.observation);
    if (
      summary.scheduler_overlap_count !== 0 ||
      summary.isolation_result !== "isolated" ||
      !summary.credential_copy_cleanup ||
      !summary.database_cleanup ||
      !summary.bucket_cleanup
    ) {
      throw new Error(summary.observation.predicate_id + " is environment_not_qualified");
    }
    const expectedQualification =
      summary.observation.outcome === "passed"
        ? "qualified"
        : summary.observation.outcome === "threshold_failed"
          ? "threshold_failed"
          : "environment_not_qualified";
    if (summary.qualification_outcome !== expectedQualification) {
      throw new Error(
        summary.observation.predicate_id + " has inconsistent qualification outcome",
      );
    }
    if (expectedQualification === "environment_not_qualified") {
      throw new Error(
        summary.observation.predicate_id + " is not eligible for active qualification",
      );
    }
    return summary;
  });
  const actualPredicateIDs = summaries
    .map((summary) => summary.observation.predicate_id)
    .sort();
  const expected = [...expectedPredicateIDs].sort();
  if (JSON.stringify(actualPredicateIDs) !== JSON.stringify(expected)) {
    throw new Error(
      "frontend measurement summaries differ: expected=" + expected.join(",") +
      " actual=" + actualPredicateIDs.join(","),
    );
  }
  return summaries.sort((left, right) =>
    left.observation.predicate_id.localeCompare(right.observation.predicate_id),
  );
}

export function ac043PredicateIDsForRows(root, rows) {
  const contract = JSON.parse(
    readFileSync(path.join(root, "contracts/performance/ac043.v1.json"), "utf8"),
  );
  const fixtureProfiles = loadPerformanceFixtureSnapshotRegistry(root);
  const contractIDs = new Set(contract.predicates.map((entry) => entry.predicate_id));
  return rows.flatMap((row) =>
    row.verification_ids.flatMap((verificationID) => {
      const binding = fixtureProfiles.verificationBindings.get(verificationID);
      if (binding === undefined) return [];
      const predicateID = binding.predicate_id;
      if (!contractIDs.has(predicateID)) {
        throw new Error(`${verificationID} maps to a predicate absent from the AC-043 contract`);
      }
      return [predicateID];
    }),
  );
}

export function measurementSchedulerOverlapCount(events, groupResults) {
  let overlaps = 0;
  for (const result of groupResults) {
    const groupUnitID = `browser_group:${result.stage_id}:${result.group_id}`;
    const groupStart = events.find(
      (event) => event.event === "started" && event.unit_id === groupUnitID,
    );
    const groupEnd = events.find(
      (event) =>
        ["completed", "failed", "cancelled"].includes(event.event) &&
        event.unit_id === groupUnitID,
    );
    if (!groupStart || !groupEnd) {
      throw new Error(`measurement group ${result.group_id} lacks a closed scheduler interval`);
    }
    // The admitted snapshot builder completes before this interval. The group
    // unit then owns clone preparation, traffic stabilization, warm-up,
    // samples, observation attachment, and lease cleanup. Its exclusive
    // host_activity claim is therefore the interval that must be proven quiet.
    overlaps += events.filter(
      (event) =>
        event.event === "started" &&
        event.seq > groupStart.seq &&
        event.seq < groupEnd.seq &&
        event.unit_id !== groupUnitID,
    ).length;
  }
  return overlaps;
}

export async function readMeasurementSchedulerEvidence(eventFile, groupResult) {
  if (!existsSync(eventFile)) {
    throw new Error("measurement finalizer requires unit-events.ndjson");
  }
  const groupUnitID = `browser_group:${groupResult.stage_id}:${groupResult.group_id}`;
  const input = createReadStream(eventFile, { encoding: "utf8" });
  const lines = createInterface({ input, crlfDelay: Number.POSITIVE_INFINITY });
  let previousSeq = 0;
  let groupStartSeq = null;
  let overlapCount = 0;
  let lineNumber = 0;
  try {
    for await (const line of lines) {
      lineNumber += 1;
      if (line === "") continue;
      let event;
      try {
        event = JSON.parse(line);
      } catch (error) {
        throw new Error(
          `unit-events.ndjson line ${lineNumber} is invalid JSON: ${error.message}`,
        );
      }
      if (!Number.isSafeInteger(event.seq) || event.seq !== previousSeq + 1) {
        throw new Error(
          `unit-events.ndjson sequence ${event.seq} is not contiguous at line ${lineNumber}`,
        );
      }
      previousSeq = event.seq;
      if (event.unit_id === groupUnitID) {
        if (event.event === "started") {
          if (groupStartSeq !== null) {
            throw new Error(`measurement group ${groupResult.group_id} starts more than once`);
          }
          groupStartSeq = event.seq;
          continue;
        }
        if (event.event === "skipped" && event.failure_reason === "dependency_failure") {
          if (groupStartSeq !== null) {
            throw new Error(`measurement group ${groupResult.group_id} is skipped after starting`);
          }
          return {
            dependency_skipped: true,
            overlap_count: 0,
            start_seq: null,
            end_seq: event.seq,
          };
        }
        if (["completed", "failed", "cancelled"].includes(event.event)) {
          if (groupStartSeq === null) {
            throw new Error(`measurement group ${groupResult.group_id} terminates before starting`);
          }
          return {
            dependency_skipped: false,
            overlap_count: overlapCount,
            start_seq: groupStartSeq,
            end_seq: event.seq,
          };
        }
      }
      if (
        groupStartSeq !== null &&
        event.event === "started" &&
        event.unit_id !== groupUnitID
      ) {
        overlapCount += 1;
      }
    }
  } finally {
    lines.close();
    input.destroy();
  }
  throw new Error(`measurement group ${groupResult.group_id} lacks a closed scheduler interval`);
}
