import { createHash } from "node:crypto";
import { existsSync, lstatSync } from "node:fs";
import path from "node:path";
import process from "node:process";

import { canonicalJSONString, parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../contract/index.mjs";
import { readCanonicalUnitEvents } from "../evidence-accounting/index.mjs";

const forbiddenKey = /^(?:bucket_name|cookie|credential|credentials|database_name|dsn|email|entered_text|password|payload|record_id|runtime_path|secret|session|token|transaction_id|txn_id|user_id)$/iu;
const reportByteLimit = 32 * 1024 * 1024;
const attachmentByteLimit = 8 * 1024 * 1024;

export function currentUnitEventSource(runRoot) {
  const configured = process.env.CARTULARY_HARNESS_UNIT_EVENT_PREFIX;
  if (configured === undefined) return { file: path.join(runRoot, "unit-events.ndjson") };
  const prefix = parseStrictJSON(configured, "canonical event prefix");
  validateSchemaSync("cartulary.harness_unit_event_prefix.v1", prefix);
  if (prefix.run_id !== process.env.CARTULARY_TEST_RUN_ID) throw new Error("live unit-event prefix belongs to another run");
  if (!path.isAbsolute(prefix.file)) {
    throw new Error("live unit-event staging path must be absolute");
  }
  const resolved = path.resolve(prefix.file);
  // Native admission supplies the exact stream; its temporary basename is
  // private to the publisher. Validate the scope and inode, not its spelling.
  if (resolved !== prefix.file || path.dirname(resolved) !== path.resolve(runRoot)) {
    throw new Error("live unit-event staging path escapes its run or has an invalid identity");
  }
  if (!existsSync(resolved)) {
    throw new Error("live unit-event staging path is missing");
  }
  const stat = lstatSync(resolved);
  if (!stat.isFile() || stat.isSymbolicLink() || stat.nlink !== 1 ||
      stat.uid !== process.getuid() || (stat.mode & 0o777) !== 0o600) {
    throw new Error("live unit-event staging path is not a private regular file");
  }
  return { file: resolved, prefix };
}

function readBoundedJSON(runRoot, file, label, maxBytes) {
  const bytes = readRetainedArtifactBytes(runRoot, path.relative(runRoot, file), maxBytes);
  return parseStrictJSON(bytes.toString("utf8"), label);
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
  return resolved;
}

function parseAttachment(runRoot, reportPath, attachment) {
  if (typeof attachment.path === "string" && attachment.path !== "") {
    return readBoundedJSON(
      runRoot,
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
  return parseStrictJSON(bytes.toString("utf8"), "measurement attachment");
}

function rejectSensitiveKeys(value) {
  // Only an independently schema-validated observation admits synthetic IDs.
  // Derive the policy here so collection and retained-artifact validation cannot
  // disagree about either the proof kind or its exact location.
  const proof = value?.schema_id === "cartulary.frontend_measurement_observation.v4" ? value.predicate_evidence : null;
  const syntheticProof = value?.predicate_id === "perf.evidence_verification.visible_with_cancel.v1" && proof?.kind === "evidence_verification" ? "evidence_verification" : value?.predicate_id === "perf.evidence_inspector.metadata_shell.v2" && proof?.kind === "evidence_inspector" ? "evidence_inspector" : value?.predicate_id === "perf.anchor_stability_under_live_updates.v2" && proof?.kind === "timeline_anchor_trace" ? "anchor" : value?.predicate_id === "perf.timeline_paste_20x5.v2" && proof?.kind === "timeline_paste_20x5" ? "paste" :
    /^perf\.timeline_presence_(?:surface|row|cell)\.rendered\.v1$/u.test(value?.predicate_id ?? "") && proof?.kind === "timeline_presence" ? "presence" :
      /^perf\.timeline_(?:history_entry_rollback|change_set_rollback|row_restore)\.rendered\.v2$/u.test(value?.predicate_id ?? "") && proof?.kind === "timeline_history" ? "history" :
      /^perf\.import_(?:csv|xlsx)_discovery\.visible_with_cancel\.v1$/u.test(value?.predicate_id ?? "") && proof?.kind === "import_discovery" ? "import" :
      /^perf\.import_(?:csv|xlsx)_apply\.visible_with_cancel\.v1$/u.test(value?.predicate_id ?? "") && proof?.kind === "import_apply" ? "import" :
        /^perf\.reporting_(?:snapshot|render)\.visible_with_cancel\.v1$/u.test(value?.predicate_id ?? "") && proof?.kind === "reporting_progress" ? "reporting" :
        /^perf\.reference_pack_(?:import|reverify|refresh)\.visible_with_cancel\.v1$/u.test(value?.predicate_id ?? "") && proof?.kind === "reference_pack_progress" ? "reference_pack" : null;
  function inspect(entry, location) {
    if (Array.isArray(entry)) {
      entry.forEach((item, index) => inspect(item, `${location}[${index}]`));
      return;
    }
    if (!entry || typeof entry !== "object") return;
    for (const [key, child] of Object.entries(entry)) {
      const admittedSyntheticID = (syntheticProof === "evidence_verification" && ["record_id", "source_record_id"].includes(key) && location === "summary.predicate_evidence") || (syntheticProof === "anchor" && ["record_id", "request_record_id"].includes(key) && location === "summary.predicate_evidence") || key === "record_id" && (
        syntheticProof === "paste" && /^summary\.predicate_evidence\.created_records\[\d+\]$/u.test(location) ||
        ["presence", "history", "import", "reporting", "reference_pack"].includes(syntheticProof) && location === "summary.predicate_evidence" ||
        syntheticProof === "evidence_inspector" && (location === "summary.predicate_evidence.target" || /^summary\.predicate_evidence\.items\[\d+\]$/u.test(location))
      );
      if (forbiddenKey.test(key) && !admittedSyntheticID) {
        throw new Error(`${location} contains forbidden key ${key}`);
      }
      inspect(child, `${location}.${key}`);
    }
  }
  inspect(value, "summary");
}

function validateSampleCardinality(summary) {
  if (summary.statistics.kind === "seeded_trace") {
    if (summary.samples.length !== 0 || summary.measured_samples !== 0 || summary.warmup_samples !== 0 || summary.threshold_ms !== null ||
        summary.statistics.completed_steps !== summary.predicate_evidence?.checkpoints?.length ||
        summary.statistics.completed_steps > summary.statistics.expected_steps) {
      throw new Error("seeded trace cannot contain timing samples or inconsistent step counts");
    }
    return;
  }
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
  if (["passed", "threshold_failed"].includes(summary.outcome)) {
    const values = summary.samples.filter((sample) => !sample.warmup).map((sample) => sample.total_ms).sort((a, b) => a - b);
    const statistics = summary.statistics;
    if (statistics.kind === "single") {
      if (summary.warmup_samples !== 0 || values.length !== 1 || statistics.elapsed_ms !== values[0]) {
        throw new Error("single-observation statistics differ from retained samples");
      }
    } else {
      const percentile = (p) => values[Math.ceil(p / 100 * values.length) - 1];
      if (values.length === 0 || statistics.p50_ms !== percentile(50) || statistics.p95_ms !== percentile(statistics.percentile)) {
        throw new Error("percentile statistics differ from retained samples");
      }
    }
  }
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
    const report = readBoundedJSON(runRoot, reportPath, "Playwright JSON report", reportByteLimit);
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
            if ((playwrightTest.results?.length ?? 0) !== 1 || (result.retry ?? 0) !== 0) {
              throw new Error("retried browser measurement cannot qualify");
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
  const bytes = readRetainedArtifactBytes(runRoot, path.relative(runRoot, file), attachmentByteLimit);
  if (digest(bytes) !== reference.sha256) {
    throw new Error(`measurement artifact digest or size is invalid: ${reference.path}`);
  }
  const artifact = parseStrictJSON(bytes.toString("utf8"), "measurement artifact");
  validateSchemaSync(schemaID, artifact);
  rejectSensitiveKeys(artifact);
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
  objectPopulation,
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
    const bytes = readRetainedArtifactBytes(runRoot, path.relative(runRoot, summaryPath), attachmentByteLimit);
    const summary = parseStrictJSON(bytes.toString("utf8"), "measurement summary");
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
    if (build.artifact.object_snapshot?.object_count !== objectPopulation?.object_count || build.artifact.object_snapshot?.total_bytes !== objectPopulation?.total_bytes) throw new Error("measurement object population diverges from its profile");
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
        : ["threshold_failed", "predicate_failed"].includes(observation.artifact.outcome)
          ? observation.artifact.outcome
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
      canonicalJSONString(build.artifact.object_snapshot) !== canonicalJSONString(lease.artifact.object_snapshot) ||
      summary.snapshot_key !== build.artifact.snapshot_key ||
      summary.snapshot_key !== lease.artifact.snapshot_key ||
      summary.fixture_profile_id !== build.artifact.fixture_profile_id ||
      summary.fixture_profile_id !== lease.artifact.fixture_profile_id ||
      summary.row_id !== lease.artifact.row_id ||
      summary.clone_ordinal !== lease.artifact.clone_ordinal ||
      summary.rollup.predicate_id !== observation.artifact.predicate_id ||
      summary.rollup.sample_count !== observation.artifact.measured_samples ||
      canonicalJSONString(summary.rollup.statistics) !== canonicalJSONString(observation.artifact.statistics) ||
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

export async function readMeasurementSchedulerEvidenceForGroups({ file: eventFile, prefix }, groupResults) {
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
  for await (const event of readCanonicalUnitEvents(eventFile, { prefix })) {
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

export async function readMeasurementSchedulerEvidence(source, groupResult) {
  const states = await readMeasurementSchedulerEvidenceForGroups(source, [groupResult]);
  return states.get(`browser_group:${groupResult.stage_id}:${groupResult.group_id}`);
}
