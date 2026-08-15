#!/usr/bin/env node

import { createHash } from "node:crypto";
import { existsSync, lstatSync, readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import {
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  groupRowsByPerformanceFixture,
  validateProfileMeasurementObservation,
} from "../performance-fixture/index.mjs";
import { buildFrontendVisualReconciliation } from "./frontend-visual-reconciliation.mjs";
import {
  collectFinalizedMeasurementSummaries,
  readMeasurementSchedulerEvidenceForGroups,
} from "./frontend-measurement-evidence.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function parse(argv) {
  const options = { target: "", groups: [], groupTargets: new Map(), children: [] };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") options.target = argv[++index] ?? "";
    else if (arg === "--groups") options.groups = (argv[++index] ?? "").split(",").filter(Boolean);
    else if (arg === "--group-targets") {
      for (const entry of (argv[++index] ?? "").split(",").filter(Boolean)) {
        const separator = entry.lastIndexOf("=");
        if (separator < 1) throw new Error("invalid browser group target mapping");
        options.groupTargets.set(entry.slice(0, separator), entry.slice(separator + 1));
      }
    }
    else if (arg === "--children") options.children = (argv[++index] ?? "").split(",").filter(Boolean);
    else throw new Error("invalid browser target finalizer arguments");
  }
  if (!options.target || options.groups.length === 0 || options.groupTargets.size !== options.groups.length) {
    throw new Error("browser target finalizer requires target and exact group targets");
  }
  return options;
}

function runRoot() {
  const results = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!results || !runID) throw new Error("browser target finalizer requires result-root identity");
  return path.resolve(root, results, runID);
}

function groupResult(base, groupID, target) {
  const file = path.join(base, target, "browser-groups", groupID, "browser-group-result.json");
  if (!existsSync(file)) return null;
  const stat = lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink() || stat.size > 16 * 1024 * 1024) {
    throw new Error(`browser group result exceeds its bounded JSON contract: ${file}`);
  }
  const bytes = readFileSync(file);
  const result = JSON.parse(bytes.toString("utf8"));
  validateSchemaSync("cartulary.browser_group_result.v5", result);
  return {
    file,
    bytes,
    result,
  };
}

function sha256(bytes) {
  return `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
}

function relativeToRun(base, file) {
  const relative = path.relative(base, file).replaceAll("\\", "/");
  if (!relative || relative.startsWith("../") || path.isAbsolute(relative)) {
    throw new Error(`browser target artifact escapes run root: ${file}`);
  }
  return relative;
}

async function measurementSchedulerOverlapCount(base, groups) {
  const eventsFile = path.join(base, "unit-events.ndjson");
  if (!existsSync(eventsFile)) {
    throw new Error("measurement qualification requires unit-events.ndjson");
  }
  const states = await readMeasurementSchedulerEvidenceForGroups(
    eventsFile,
    groups.map((group) => group.result),
  );
  return [...states.values()].reduce((total, state) => total + state.overlap_count, 0);
}

async function writeTargetResult(base, options, groups) {
  const targetDirectory = path.join(base, options.target);
  const output = path.join(targetDirectory, "browser-target-result.json");
  if (existsSync(output)) {
    throw new Error(`browser target result is immutable: ${output}`);
  }
  secureMkdir(targetDirectory);
  const artifacts = [];
  let visualReconciliation = null;
  let measurementAggregate = null;
  if (options.target === "browser-e2e-measurement") {
    const catalog = loadTestCatalog(root);
    const selectedRows = groups.flatMap((group) =>
      group.result.selected_rows.map((rowID) => {
        const row = catalog.rowByID.get(rowID);
        if (!row) throw new Error(`measurement target references unknown row ${rowID}`);
        return row;
      }),
    );
    const fixtureGroups = groupRowsByPerformanceFixture(root, selectedRows);
    if (fixtureGroups.length > 0) {
      const aggregateSchemaIDs = new Set(
        fixtureGroups.map(
          (fixtureGroup) => fixtureGroup.profile.artifact_policy.aggregate_schema_id,
        ),
      );
      if (aggregateSchemaIDs.size !== 1) {
        throw new Error("measurement profile groups select mixed aggregate generations");
      }
      const profileGroups = [];
      for (const fixtureGroup of fixtureGroups) {
        const profiledGroups = groups.filter((group) =>
          group.result.selected_rows.some((rowID) =>
            fixtureGroup.row_ids.includes(rowID),
          ),
        );
        const schedulerOverlapCount = await measurementSchedulerOverlapCount(
          base,
          profiledGroups,
        );
        if (schedulerOverlapCount !== 0) {
          throw new Error(
            `measurement quiet interval overlapped ${schedulerOverlapCount} ordinary units`,
          );
        }
        const summaryPaths = profiledGroups.map((group) =>
          path.join(
            path.dirname(group.file),
            `frontend-measurement-summary.${fixtureGroup.profile.artifact_policy.summary_schema_id.split(".").at(-1)}.json`,
          ),
        );
        const summaries = collectFinalizedMeasurementSummaries({
          buildSchemaID: fixtureGroup.profile.artifact_policy.build_schema_id,
          expectedPredicateIDs: fixtureGroup.predicate_ids,
          leaseSchemaID: fixtureGroup.profile.artifact_policy.lease_schema_id,
          observationSchemaID:
            fixtureGroup.profile.artifact_policy.observation_schema_id,
          runRoot: base,
          summarySchemaID: fixtureGroup.profile.artifact_policy.summary_schema_id,
          summaryPaths,
          validateObservation(observation) {
            validateProfileMeasurementObservation(
              root,
              fixtureGroup.profile,
              observation,
            );
          },
        });
        const actualRowIDs = summaries
          .map((entry) => entry.summary.row_id)
          .sort();
        if (JSON.stringify(actualRowIDs) !== JSON.stringify(fixtureGroup.row_ids)) {
          throw new Error("measurement summaries have inconsistent row provenance");
        }
        const fixtureProfileIDs = new Set(
          summaries.map((entry) => entry.summary.fixture_profile_id),
        );
        const snapshotKeys = new Set(
          summaries.map((entry) => entry.summary.snapshot_key),
        );
        const buildArtifacts = new Map(
          summaries.map((entry) => [
            JSON.stringify(entry.summary.build_artifact),
            entry.summary.build_artifact,
          ]),
        );
        const leaseArtifactPaths = new Set(
          summaries.map((entry) => entry.summary.lease_artifact.path),
        );
        const cloneOrdinals = new Set(
          summaries.map((entry) => entry.summary.clone_ordinal),
        );
        if (
          fixtureProfileIDs.size !== 1 ||
          snapshotKeys.size !== 1 ||
          buildArtifacts.size !== 1 ||
          leaseArtifactPaths.size !== summaries.length ||
          cloneOrdinals.size !== summaries.length ||
          summaries.length !== fixtureGroup.row_ids.length
        ) {
          throw new Error("measurement summaries have inconsistent snapshot provenance");
        }
        profileGroups.push({
          fixture_profile_id: [...fixtureProfileIDs][0],
          snapshot_key: [...snapshotKeys][0],
          builder_count: 1,
          clone_count: summaries.length,
          scheduler_overlap_count: schedulerOverlapCount,
          build_artifact: [...buildArtifacts.values()][0],
          summary_artifacts: summaries
            .map((entry) => ({
              role: "frontend_measurement_summary",
              path_kind: "file",
              format: "json",
              path: relativeToRun(base, entry.file),
              sha256: sha256(entry.bytes),
            }))
            .sort((left, right) => left.path.localeCompare(right.path)),
          rollups: summaries
            .map((entry) => ({
              predicate_id: entry.summary.rollup.predicate_id,
              sample_count: entry.summary.rollup.sample_count,
              p95_ms: entry.summary.rollup.p95_ms,
              threshold_ms: entry.summary.rollup.threshold_ms,
              outcome: entry.summary.rollup.outcome,
            }))
            .sort((left, right) =>
              left.predicate_id.localeCompare(right.predicate_id),
            ),
        });
      }
      measurementAggregate = {
        schema_id: [...aggregateSchemaIDs][0],
        target_id: options.target,
        status: profileGroups.some((group) =>
          group.rollups.some((rollup) => rollup.outcome === "threshold_failed"),
        )
          ? "threshold_failed"
          : "qualified",
        profile_groups: profileGroups.sort((left, right) =>
          left.fixture_profile_id.localeCompare(right.fixture_profile_id) ||
          left.snapshot_key.localeCompare(right.snapshot_key),
        ),
      };
      validateSchemaSync(measurementAggregate.schema_id, measurementAggregate);
      const aggregateOutput = path.join(
        targetDirectory,
        "frontend-measurement-aggregate.json",
      );
      const aggregateBytes = Buffer.from(
        `${JSON.stringify(measurementAggregate, null, 2)}\n`,
        "utf8",
      );
      secureWriteFile(aggregateOutput, aggregateBytes);
      artifacts.push({
        kind: `frontend_measurement_aggregate_${measurementAggregate.schema_id.split(".").at(-1)}`,
        ref: relativeToRun(base, aggregateOutput),
        sha256: sha256(aggregateBytes),
      });
    }
  }
  if (options.target === "browser-e2e-visual") {
    const reconciliationOutput = path.join(
      targetDirectory,
      "frontend-visual-reconciliation.json",
    );
    if (existsSync(reconciliationOutput)) {
      throw new Error(
        `frontend visual reconciliation is immutable: ${reconciliationOutput}`,
      );
    }
    visualReconciliation = buildFrontendVisualReconciliation({
      root,
      reportPaths: groups.map((group) =>
        path.join(path.dirname(group.file), "playwright-report.json"),
      ),
      attemptPassed: groups.every((group) => group.result.status === "pass"),
    });
    validateSchemaSync(visualReconciliation.schema_id, visualReconciliation);
    const reconciliationBytes = Buffer.from(
      `${JSON.stringify(visualReconciliation, null, 2)}\n`,
      "utf8",
    );
    secureWriteFile(reconciliationOutput, reconciliationBytes);
    artifacts.push({
      kind: "frontend_visual_reconciliation_v1",
      ref: relativeToRun(base, reconciliationOutput),
      sha256: sha256(reconciliationBytes),
    });
  }
  const sessionsByID = new Map();
  for (const group of groups) {
    const existing = sessionsByID.get(group.result.browser_session_id);
    const session = {
      browser_session_id: group.result.browser_session_id,
      runtime_profile_id: group.result.runtime_profile_id,
      service_requirement: group.result.service_requirement,
      artifacts: group.result.session_artifacts,
    };
    if (existing && JSON.stringify(existing) !== JSON.stringify(session)) {
      throw new Error(
        `browser session ${group.result.browser_session_id} has conflicting target evidence`,
      );
    }
    sessionsByID.set(group.result.browser_session_id, session);
  }
  const payload = {
    schema_id: "cartulary.browser_target_result.v3",
    target_id: options.target,
    status:
      groups.every((group) => group.result.status === "pass") &&
      (visualReconciliation === null || visualReconciliation.status === "pass") &&
      (measurementAggregate === null || measurementAggregate.status === "qualified")
        ? "pass"
        : "fail",
    group_results: groups
      .map((group) => ({
        group_id: group.result.group_id,
        browser_session_id: group.result.browser_session_id,
        ref: relativeToRun(base, group.file),
        sha256: sha256(group.bytes),
      }))
      .sort((left, right) => left.group_id.localeCompare(right.group_id)),
    sessions: [...sessionsByID.values()].sort((left, right) =>
      left.browser_session_id.localeCompare(right.browser_session_id),
    ),
    ...(artifacts.length > 0 ? { artifacts } : {}),
    generated_at: new Date().toISOString(),
  };
  validateSchemaSync(payload.schema_id, payload);
  secureWriteFile(output, `${JSON.stringify(payload, null, 2)}\n`);
  return payload;
}

const options = parse(process.argv.slice(2));
const base = runRoot();
const groups = options.groups
  .map((group) =>
    groupResult(base, group, options.groupTargets.get(group)),
  )
  .filter((group) => group !== null);
const complete = groups.length === options.groups.length;
const targetResult = complete ? await writeTargetResult(base, options, groups) : null;
if (!complete) {
  process.stderr.write(`browser target ${options.target} is missing group results\n`);
  process.exitCode = 11;
} else if (targetResult.status !== "pass") {
  process.stderr.write(
    `browser target ${options.target} contains failed groups or target artifacts\n`,
  );
  process.exitCode = 10;
}
