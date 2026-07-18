#!/usr/bin/env node

import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import {
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  accountingRowsForTarget,
  buildTestEvidenceAccounting,
  buildTestOwnerSummary,
  evidenceTargetForCatalogRow,
} from "../evidence-accounting/index.mjs";
import { buildSourceSnapshot } from "../owner-slice/source-snapshot.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";
import { parseStrictJSON } from "../test-catalog/semantic-json.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function usage() {
  return "usage: browser-evidence-finalize-cli.mjs <target-id>";
}

function runRoot() {
  const results = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  if (!results || !runID) throw new Error("CARTULARY_TEST_RESULTS_DIR and CARTULARY_TEST_RUN_ID are required");
  return path.resolve(root, results, runID);
}

function readGroupResults(targetID) {
  const directory = path.join(runRoot(), targetID, "browser-groups");
  if (!existsSync(directory)) return [];
  return readdirSync(directory, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(directory, entry.name, "browser-group-result.json"))
    .filter(existsSync)
    .map((file) => ({
      file,
      result: parseStrictJSON(readFileSync(file, "utf8"), file),
    }))
    .filter(({ result }) => result.schema_id === "cartulary.browser_group_result.v1" && result.target_id === targetID)
    .sort((left, right) => left.result.group_id.localeCompare(right.result.group_id));
}

function resultByRow(groupResults) {
  const byRow = new Map();
  for (const group of groupResults) {
    for (const row of group.result.row_results ?? []) {
      const values = byRow.get(row.row_id) ?? [];
      values.push({ group, row });
      byRow.set(row.row_id, values);
    }
  }
  return byRow;
}

function normalizedRowResult(rowID, observations) {
  if (observations.length !== 1) {
    return {
      row_id: rowID,
      terminal_state: "infrastructure_failed",
      duration_ms: 0,
      exit_code: 11,
      failure_reason: observations.length === 0 ? "missing_selector_result" : "duplicate_selector_result",
      attempt: 1,
    };
  }
  return { ...observations[0].row, attempt: 1 };
}

function workUnitsForOwner(ownerID, selectedRows, observations) {
  return selectedRows.map((rowID) => ({
    work_unit_id: `browser:${ownerID}:${rowID}`,
    row_ids: [rowID],
    observation: observations.get(rowID)?.[0] ?? null,
  }));
}

function writeValidated(file, value) {
  validateSchemaSync(value.schema_id, value);
  secureWriteFile(file, `${JSON.stringify(value, null, 2)}\n`);
}

function finalizeOwner(targetID, selection, observations, snapshot, startedAt, finishedAt) {
  const ownerID = selection.owner_id;
  const units = workUnitsForOwner(ownerID, selection.selected_rows, observations);
  const plan = {
    command_id: "cartulary.harness.command.browser_evidence.v1",
    run_id: process.env.CARTULARY_TEST_RUN_ID,
    owner_id: ownerID,
    selected_rows: selection.selected_rows,
    source_snapshot_digest: snapshot.digest,
    catalog_semantic_digest: selection.catalog_semantic_digest,
    verification_semantic_digest: selection.verification_semantic_digest,
    runtime_profile_digest: selection.runtime_profile_digest,
    resource_profile_digest: selection.resource_profile_digest,
    fixture_profile_digest: selection.fixture_profile_digest,
    target: targetID,
    rows: selection.expected_rows,
    work_units: units,
    selection: { completion_scope: "target_partition" },
    unused_inputs: [],
  };
  const rowResults = selection.selected_rows.map((rowID) =>
    normalizedRowResult(rowID, observations.get(rowID) ?? []),
  );
  const execution = {
    status: rowResults.every((row) => row.terminal_state === "passed" || row.terminal_state === "skipped_authorized")
      ? "pass"
      : "fail",
    row_results: rowResults,
    duration_ms: rowResults.reduce((total, row) => total + row.duration_ms, 0),
  };
  const logs = units.map((unit) => ({
    work_unit_id: unit.work_unit_id,
    stdout_path: unit.observation?.group.result.artifacts?.stdout ?? "",
    stderr_path: unit.observation?.group.result.artifacts?.stderr ?? "",
  }));
  const accounting = buildTestEvidenceAccounting(plan, execution, logs, startedAt, finishedAt);
  const prefix = `${targetID}/owners/${ownerID}`;
  const artifacts = {
    evidence_accounting: `${prefix}/test-evidence-accounting.json`,
    owner_summary: `${prefix}/test-owner-summary.json`,
    tool_run_summary: `${targetID}/tool-run-summary.json`,
  };
  const ownerSummary = buildTestOwnerSummary(plan, accounting, artifacts);
  const ownerDir = path.join(runRoot(), targetID, "owners", ownerID);
  secureMkdir(ownerDir);
  writeValidated(path.join(ownerDir, "test-evidence-accounting.json"), accounting);
  writeValidated(path.join(ownerDir, "test-owner-summary.json"), ownerSummary);
  return {
    owner_id: ownerID,
    status: accounting.status,
    selected_rows: selection.selected_rows,
    evidence_accounting: artifacts.evidence_accounting,
    owner_summary: artifacts.owner_summary,
  };
}

function main() {
  if (process.argv.length !== 3) throw new Error(usage());
  const targetID = process.argv[2];
  if (!/^[a-z][a-z0-9-]*$/u.test(targetID)) throw new Error(usage());
  const catalog = loadTestCatalog(root);
  const targetRows = catalog.rows.filter(
    (row) => row.runner === "playwright" && evidenceTargetForCatalogRow(row) === targetID,
  );
  if (targetRows.length === 0) return 0;
  const expectedIDs = new Set(targetRows.map((row) => row.row_id));
  const groupResults = readGroupResults(targetID);
  const observations = resultByRow(groupResults);
  const unexpected = [...observations.keys()].filter((rowID) => !expectedIDs.has(rowID)).sort();
  if (unexpected.length > 0) throw new Error(`target ${targetID} observed unexpected rows: ${unexpected.join(", ")}`);
  const startedAt = groupResults.map(({ result }) => result.started_at).sort().at(0) ?? new Date().toISOString();
  const finishedAt = groupResults.map(({ result }) => result.finished_at).sort().at(-1) ?? startedAt;
  const snapshot = buildSourceSnapshot(root);
  const ownerIDs = [...new Set(targetRows.map((row) => row.owner_id))].sort();
  const shards = ownerIDs.map((ownerID) => {
    const selection = accountingRowsForTarget(root, { ownerID, targetName: targetID });
    return finalizeOwner(targetID, selection, observations, snapshot, startedAt, finishedAt);
  });
  const index = {
    schema_id: "cartulary.browser_owner_index.v1",
    target_id: targetID,
    run_id: process.env.CARTULARY_TEST_RUN_ID,
    source_snapshot_digest: snapshot.digest,
    catalog_semantic_digest: catalog.summary.catalog_semantic_digest,
    verification_semantic_digest: catalog.summary.verification_semantic_digest,
    status: shards.every((shard) => shard.status === "pass") ? "pass" : "fail",
    selected_rows: [...expectedIDs].sort(),
    owner_shards: shards,
  };
  validateSchemaSync(index.schema_id, index);
  secureWriteFile(
    path.join(runRoot(), targetID, "browser-owner-index.json"),
    `${JSON.stringify(index, null, 2)}\n`,
  );
  return index.status === "pass" ? 0 : 11;
}

try {
  process.exitCode = main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = error.message === usage() ? 2 : 11;
}
