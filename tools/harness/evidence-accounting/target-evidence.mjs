import {
  existsSync,
  readFileSync,
  readdirSync,
} from "node:fs";
import path from "node:path";

import {
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import { buildSourceSnapshot } from "../owner-slice/source-snapshot.mjs";
import { loadTestCatalog, targetForCatalogRow } from "../test-catalog/index.mjs";
import { parseStrictJSON } from "../test-catalog/semantic-json.mjs";
import { accountingRowsForTarget } from "./catalog-accounting.mjs";
import {
  buildTestEvidenceAccounting,
  buildTestOwnerSummary,
} from "./owner-evidence.mjs";

const successfulStates = new Set(["passed", "skipped_authorized"]);
const selectionScopes = new Set(["all", "default_check", "rows"]);

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sortedUnique(values, label) {
  const sorted = [...values].sort(asciiCompare);
  if (new Set(sorted).size !== sorted.length) {
    throw new Error(`${label} contains duplicates`);
  }
  return sorted;
}

function sortedDistinct(values) {
  return [...new Set(values)].sort(asciiCompare);
}

function readJSON(file) {
  return parseStrictJSON(readFileSync(file, "utf8"), file);
}

function commandTargetContext(root) {
  const manifest = readJSON(path.join(root, "tools/task_surface_manifest.json"));
  return {
    commandByTarget: new Map(
      manifest.targets.map((entry) => [entry.name, entry.command_id ?? ""]),
    ),
    targetByCommand: new Map(
      manifest.targets.map((entry) => [entry.command_id, entry.name]),
    ),
  };
}

function rowTarget(row, targetByCommand) {
  return targetForCatalogRow(row, { commandTargetByID: targetByCommand });
}

function rowIDList(value, label) {
  const parsed = String(value ?? "")
    .split(/[\n,]/u)
    .map((entry) => entry.trim())
    .filter(Boolean);
  const normalized = sortedUnique(parsed, label);
  if (JSON.stringify(parsed) !== JSON.stringify(normalized)) {
    throw new Error(`${label} must be ASCII-sorted`);
  }
  return normalized;
}

function selectionTuple(env) {
  const genericScope = String(env.CARTULARY_TARGET_EVIDENCE_SCOPE ?? "").trim();
  const goScope = String(env.CARTULARY_GO_SCHEDULE_SCOPE ?? "").trim();
  const scope = genericScope || goScope || "all";
  if (!selectionScopes.has(scope)) {
    throw new Error(`invalid target evidence selection scope ${scope}`);
  }
  if (genericScope && goScope && genericScope !== goScope) {
    throw new Error("target evidence and Go schedule selection scopes disagree");
  }
  const genericRows = rowIDList(
    env.CARTULARY_TARGET_EVIDENCE_ROW_IDS,
    "target evidence row selection",
  );
  const goRows = rowIDList(
    env.CARTULARY_GO_SCHEDULED_ROW_IDS,
    "Go schedule row selection",
  );
  if (
    genericRows.length > 0 &&
    goRows.length > 0 &&
    JSON.stringify(genericRows) !== JSON.stringify(goRows)
  ) {
    throw new Error("target evidence and Go schedule row selections disagree");
  }
  const rowIDs = genericRows.length > 0 ? genericRows : goRows;
  if (scope === "rows" && rowIDs.length === 0) {
    throw new Error("target evidence rows scope requires row IDs");
  }
  if (scope !== "rows" && rowIDs.length > 0) {
    throw new Error(`target evidence ${scope} scope forbids row IDs`);
  }
  return { scope, rowIDs };
}

function selectedTargetRows(rows, targetID, targetByCommand, tuple) {
  const targetRows = rows
    .filter((row) => rowTarget(row, targetByCommand) === targetID)
    .sort((left, right) => asciiCompare(left.row_id, right.row_id));
  if (tuple.scope === "all") {
    return targetRows;
  }
  if (tuple.scope === "default_check") {
    return targetRows.filter((row) => row.default_check === true);
  }
  const byID = new Map(targetRows.map((row) => [row.row_id, row]));
  return tuple.rowIDs.map((rowID) => {
    const row = byID.get(rowID);
    if (!row) {
      throw new Error(`target evidence row ${rowID} is not selected by ${targetID}`);
    }
    return row;
  });
}

function filesNamed(root, filename) {
  if (!existsSync(root)) {
    return [];
  }
  const files = [];
  const stack = [root];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      if (entry.name === "owners") {
        continue;
      }
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
      } else if (entry.isFile() && entry.name === filename) {
        files.push(next);
      }
    }
  }
  return files.sort(asciiCompare);
}

function normalizeTimestamp(value, fallback) {
  if (typeof value === "number" && Number.isFinite(value)) {
    return new Date(value).toISOString();
  }
  const parsed = Date.parse(String(value ?? ""));
  return Number.isFinite(parsed) ? new Date(parsed).toISOString() : fallback;
}

function timingWindow(records) {
  const fallback = new Date().toISOString();
  const starts = records
    .map((record) => normalizeTimestamp(record.started_at ?? record.start_time, ""))
    .filter(Boolean)
    .sort(asciiCompare);
  const finishes = records
    .map((record) => normalizeTimestamp(record.finished_at ?? record.end_time, ""))
    .filter(Boolean)
    .sort(asciiCompare);
  return {
    startedAt: starts.at(0) ?? fallback,
    finishedAt: finishes.at(-1) ?? starts.at(0) ?? fallback,
  };
}

function passedRow(rowID, durationMs = 0) {
  return {
    row_id: rowID,
    terminal_state: "passed",
    duration_ms: Math.max(0, Math.round(durationMs || 0)),
    exit_code: 0,
    attempt: 1,
  };
}

function goObservations(targetDir, selectedRows, catalogRowIDs) {
  const summaries = filesNamed(targetDir, "step-summary.json").map(readJSON);
  const inventoryByID = new Map();
  for (const summary of summaries) {
    for (const item of summary.inventory ?? []) {
      if (!item.id) continue;
      const values = inventoryByID.get(item.id) ?? [];
      values.push(item);
      inventoryByID.set(item.id, values);
    }
  }
  const selectedIDs = new Set(selectedRows.map((row) => row.row_id));
  const unexpected = [...inventoryByID.keys()]
    .filter((rowID) => catalogRowIDs.has(rowID) && !selectedIDs.has(rowID))
    .sort(asciiCompare);
  if (unexpected.length > 0) {
    throw new Error(
      `target evidence ${selectedRows[0]?.runner ?? "go"} observations contain unexpected rows: ${unexpected.join(", ")}`,
    );
  }
  const results = selectedRows.map((row) => {
    const expected = sortedUnique(row.selector.tests ?? [], `${row.row_id} Go selectors`);
    const observed = sortedUnique(
      (inventoryByID.get(row.row_id) ?? []).map((item) => item.symbol_or_title),
      `${row.row_id} Go observations`,
    );
    if (JSON.stringify(observed) !== JSON.stringify(expected)) {
      throw new Error(
        `target evidence Go selector mismatch for ${row.row_id}: expected=${expected.length} observed=${observed.length}`,
      );
    }
    return passedRow(row.row_id);
  });
  return { results, records: summaries, logs: new Map() };
}

function normalizedReportFile(root, value) {
  const absolute = path.isAbsolute(value) ? value : path.resolve(root, value);
  return path.relative(root, absolute).replaceAll("\\", "/");
}

function vitestObservations(
  root,
  targetDir,
  selectedRows,
  { rejectUnselected = false } = {},
) {
  const reportFile = path.join(targetDir, "raw", "frontend-unit", "runner.json");
  if (!existsSync(reportFile)) {
    throw new Error(`target evidence Vitest report is missing: ${reportFile}`);
  }
  const report = readJSON(reportFile);
  const expectedKeys = new Set(
    selectedRows.flatMap((row) =>
      (row.selector.titles ?? []).map(
        (title) => `${row.selector.file}\u0000${title}`,
      ),
    ),
  );
  const bySelector = new Map();
  const unexpected = [];
  for (const fileResult of report.testResults ?? []) {
    const file = normalizedReportFile(root, fileResult.name);
    for (const assertion of fileResult.assertionResults ?? []) {
      const candidates = [assertion.title, assertion.fullName]
        .filter((title) => typeof title === "string")
        .map((title) => `${file}\u0000${title}`)
        .filter((key) => expectedKeys.has(key));
      const keys = sortedDistinct(candidates);
      if (keys.length === 0) {
        unexpected.push(
          `${file} :: ${assertion.fullName ?? assertion.title ?? "<untitled>"}`,
        );
        continue;
      }
      for (const key of keys) {
        const values = bySelector.get(key) ?? [];
        values.push(assertion);
        bySelector.set(key, values);
      }
    }
  }
  const results = selectedRows.map((row) => {
    let durationMs = 0;
    for (const title of row.selector.titles ?? []) {
      const key = `${row.selector.file}\u0000${title}`;
      const observations = bySelector.get(key) ?? [];
      if (observations.length !== 1 || observations[0].status !== "passed") {
        throw new Error(
          `target evidence Vitest selector mismatch for ${row.row_id}: ${row.selector.file} :: ${title}`,
        );
      }
      durationMs += Number(observations[0].duration ?? 0);
    }
    return passedRow(row.row_id, durationMs);
  });
  if (rejectUnselected && unexpected.length > 0) {
    throw new Error(
      `target evidence Vitest report contains ${unexpected.length} selectors outside the selected catalog partition`,
    );
  }
  return {
    results,
    records: [{ started_at: report.startTime, finished_at: report.testResults?.map((entry) => entry.endTime).filter(Boolean).sort((a, b) => a - b).at(-1) }],
    logs: new Map(
      selectedRows.map((row) => [
        row.row_id,
        {
          stdout_path: path.join(targetDir, "raw", "frontend-unit", "stdout.log"),
          stderr_path: path.join(targetDir, "raw", "frontend-unit", "stderr.log"),
        },
      ]),
    ),
  };
}

function playwrightObservations(targetDir, selectedRows) {
  const groupRoot = path.join(targetDir, "browser-groups");
  const groupFiles = filesNamed(groupRoot, "browser-group-result.json");
  const groups = groupFiles.map((file) => ({ file, result: readJSON(file) }));
  const byRow = new Map();
  for (const group of groups) {
    validateSchemaSync("cartulary.browser_group_result.v1", group.result);
    for (const row of group.result.row_results ?? []) {
      const values = byRow.get(row.row_id) ?? [];
      values.push({ group, row });
      byRow.set(row.row_id, values);
    }
  }
  const selectedIDs = new Set(selectedRows.map((row) => row.row_id));
  const unexpected = [...byRow.keys()]
    .filter((rowID) => !selectedIDs.has(rowID))
    .sort(asciiCompare);
  if (unexpected.length > 0) {
    throw new Error(
      `target evidence Playwright observations contain unexpected rows: ${unexpected.join(", ")}`,
    );
  }
  const logs = new Map();
  const results = selectedRows.map((row) => {
    const observations = byRow.get(row.row_id) ?? [];
    if (
      observations.length !== 1 ||
      !successfulStates.has(observations[0].row.terminal_state)
    ) {
      throw new Error(
        `target evidence Playwright selector mismatch for ${row.row_id}: observations=${observations.length}`,
      );
    }
    logs.set(row.row_id, {
      stdout_path: observations[0].group.result.artifacts?.stdout ?? "",
      stderr_path: observations[0].group.result.artifacts?.stderr ?? "",
    });
    return { ...observations[0].row, attempt: 1 };
  });
  return { results, records: groups.map((entry) => entry.result), logs };
}

function shellObservations(selectedRows, commandID) {
  for (const row of selectedRows) {
    if (row.selector.command_id !== commandID) {
      throw new Error(
        `target evidence shell selector ${row.selector.command_id} does not match ${commandID}`,
      );
    }
  }
  return { results: selectedRows.map((row) => passedRow(row.row_id)), records: [], logs: new Map() };
}

function observationsForRunner(
  root,
  targetDir,
  targetID,
  commandID,
  selectedRows,
  catalogRowIDs,
  selectionScope,
) {
  const runners = sortedDistinct(selectedRows.map((row) => row.runner));
  if (runners.length !== 1) {
    throw new Error(`${targetID} target evidence must use exactly one runner family`);
  }
  if (runners[0] === "go") {
    return goObservations(targetDir, selectedRows, catalogRowIDs);
  }
  if (runners[0] === "vitest") {
    return vitestObservations(root, targetDir, selectedRows, {
      rejectUnselected: selectionScope === "rows",
    });
  }
  if (runners[0] === "playwright") return playwrightObservations(targetDir, selectedRows);
  if (runners[0] === "shell") return shellObservations(selectedRows, commandID);
  throw new Error(`unsupported target evidence runner ${runners[0]}`);
}

function writeValidated(file, value) {
  validateSchemaSync(value.schema_id, value);
  secureWriteFile(file, `${JSON.stringify(value, null, 2)}\n`);
}

function existingOwnerShards(
  targetDir,
  targetID,
  runID,
  currentIdentity,
  selectedRows,
) {
  const ownersDir = path.join(targetDir, "owners");
  if (!existsSync(ownersDir)) return null;
  const shards = [];
  for (const entry of readdirSync(ownersDir, { withFileTypes: true }).sort((left, right) => asciiCompare(left.name, right.name))) {
    if (!entry.isDirectory()) continue;
    const ownerDir = path.join(ownersDir, entry.name);
    const accountingFile = path.join(ownerDir, "test-evidence-accounting.json");
    const summaryFile = path.join(ownerDir, "test-owner-summary.json");
    if (!existsSync(accountingFile) || !existsSync(summaryFile)) {
      throw new Error(`target evidence owner shard ${entry.name} is incomplete`);
    }
    const accounting = readJSON(accountingFile);
    const summary = readJSON(summaryFile);
    validateSchemaSync("cartulary.test_evidence_accounting.v1", accounting);
    validateSchemaSync("cartulary.test_owner_summary.v1", summary);
    for (const [field, expected] of Object.entries({
      run_id: runID,
      owner_id: entry.name,
      target_id: targetID,
      status: "pass",
      ...currentIdentity,
    })) {
      if (accounting[field] !== expected || summary[field] !== expected) {
        throw new Error(`target evidence owner shard ${entry.name} is incompatible on ${field}`);
      }
    }
    if (JSON.stringify(accounting.selected_rows) !== JSON.stringify(summary.selected_rows)) {
      throw new Error(`target evidence owner shard ${entry.name} has mismatched selected rows`);
    }
    shards.push({
      owner_id: entry.name,
      status: accounting.status,
      selected_rows: accounting.selected_rows,
      evidence_accounting: `${targetID}/owners/${entry.name}/test-evidence-accounting.json`,
      owner_summary: `${targetID}/owners/${entry.name}/test-owner-summary.json`,
    });
  }
  if (shards.length === 0) return null;
  const retainedRows = shards
    .flatMap((shard) => shard.selected_rows)
    .sort(asciiCompare);
  const expectedRows = selectedRows.map((row) => row.row_id).sort(asciiCompare);
  if (JSON.stringify(retainedRows) !== JSON.stringify(expectedRows)) {
    throw new Error("target evidence owner shards do not match the selected row scope");
  }
  return shards;
}

export function targetOwnerEvidenceArtifactPaths(
  targetDir,
  targetID,
  { runID = "" } = {},
) {
  const ownersDir = path.join(targetDir, "owners");
  if (!existsSync(ownersDir)) return [];
  let commonIdentity = null;
  return readdirSync(ownersDir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .sort((left, right) => asciiCompare(left.name, right.name))
    .flatMap((entry) => {
      const ownerDir = path.join(ownersDir, entry.name);
      const accountingFile = path.join(
        ownerDir,
        "test-evidence-accounting.json",
      );
      const summaryFile = path.join(ownerDir, "test-owner-summary.json");
      if (!existsSync(accountingFile) || !existsSync(summaryFile)) {
        throw new Error(
          `owner evidence shard ${entry.name} is incomplete for ${targetID}`,
        );
      }
      const accounting = readJSON(accountingFile);
      const summary = readJSON(summaryFile);
      validateSchemaSync("cartulary.test_evidence_accounting.v1", accounting);
      validateSchemaSync("cartulary.test_owner_summary.v1", summary);
      if (
        accounting.target_id !== targetID ||
        summary.target_id !== targetID ||
        accounting.owner_id !== entry.name ||
        summary.owner_id !== entry.name ||
        accounting.status !== summary.status ||
        JSON.stringify(accounting.selected_rows) !==
          JSON.stringify(summary.selected_rows) ||
        (runID && (accounting.run_id !== runID || summary.run_id !== runID))
      ) {
        throw new Error(
          `owner evidence shard ${entry.name} is incompatible for ${targetID}`,
        );
      }
      const identity = {
        run_id: accounting.run_id,
        source_snapshot_digest: accounting.source_snapshot_digest,
        catalog_semantic_digest: accounting.catalog_semantic_digest,
        verification_semantic_digest: accounting.verification_semantic_digest,
      };
      if (commonIdentity === null) {
        commonIdentity = identity;
      } else if (JSON.stringify(commonIdentity) !== JSON.stringify(identity)) {
        throw new Error(`owner evidence shards have mixed identity for ${targetID}`);
      }
      return [
        { role: "test_evidence_accounting", path: accountingFile },
        { role: "test_owner_summary", path: summaryFile },
      ];
    });
}

function writeBrowserIndex(targetDir, targetID, runID, identity, selectedRows, shards) {
  const index = {
    schema_id: "cartulary.browser_owner_index.v1",
    target_id: targetID,
    run_id: runID,
    ...identity,
    status: shards.every((shard) => shard.status === "pass") ? "pass" : "fail",
    selected_rows: selectedRows.map((row) => row.row_id).sort(asciiCompare),
    owner_shards: shards,
  };
  validateSchemaSync(index.schema_id, index);
  secureWriteFile(
    path.join(targetDir, "browser-owner-index.json"),
    `${JSON.stringify(index, null, 2)}\n`,
  );
}

export function finalizeTargetOwnerEvidence(
  root,
  {
    targetID,
    requestedStatus,
    resultsDir,
    runID,
    env = process.env,
  },
) {
  if (requestedStatus !== "pass") return { status: "not_selected", shards: [] };
  const catalog = loadTestCatalog(root);
  const { commandByTarget, targetByCommand } = commandTargetContext(root);
  const commandID = commandByTarget.get(targetID) ?? "";
  const allTargetRows = catalog.rows.filter(
    (row) => rowTarget(row, targetByCommand) === targetID,
  );
  if (allTargetRows.length === 0) return { status: "not_applicable", shards: [] };
  if (!commandID) throw new Error(`target evidence command identity is unavailable for ${targetID}`);

  const tuple = selectionTuple(env);
  const selectedRows = selectedTargetRows(
    catalog.rows,
    targetID,
    targetByCommand,
    tuple,
  );
  if (selectedRows.length === 0) {
    throw new Error(`${targetID} target evidence selection contains zero rows`);
  }

  const targetDir = path.join(resultsDir, runID, targetID);
  secureMkdir(targetDir);
  const sourceSnapshot = buildSourceSnapshot(root);
  const identity = {
    source_snapshot_digest: sourceSnapshot.digest,
    catalog_semantic_digest: catalog.summary.catalog_semantic_digest,
    verification_semantic_digest: catalog.summary.verification_semantic_digest,
  };
  const existing = existingOwnerShards(
    targetDir,
    targetID,
    runID,
    identity,
    selectedRows,
  );
  if (existing) {
    return { status: "pass", shards: existing, reused: true };
  }

  const observed = observationsForRunner(
    root,
    targetDir,
    targetID,
    commandID,
    selectedRows,
    new Set(catalog.rows.map((row) => row.row_id)),
    tuple.scope,
  );
  const resultByID = new Map(observed.results.map((result) => [result.row_id, result]));
  const { startedAt, finishedAt } = timingWindow(observed.records);
  const ownerIDs = sortedDistinct(selectedRows.map((row) => row.owner_id));
  const shards = ownerIDs.map((ownerID) => {
    const ownerRows = selectedRows.filter((row) => row.owner_id === ownerID);
    const selection = accountingRowsForTarget(root, {
      ownerID,
      rowIDs: ownerRows.map((row) => row.row_id),
      targetName: targetID,
    });
    const units = selection.selected_rows.map((rowID) => ({
      work_unit_id: `target:${targetID}:${rowID}`,
      row_ids: [rowID],
    }));
    const plan = {
      command_id:
        ownerRows[0].runner === "playwright"
          ? "cartulary.harness.command.browser_evidence.v1"
          : commandID,
      run_id: runID,
      owner_id: ownerID,
      selected_rows: selection.selected_rows,
      ...identity,
      runtime_profile_digest: selection.runtime_profile_digest,
      resource_profile_digest: selection.resource_profile_digest,
      fixture_profile_digest: selection.fixture_profile_digest,
      target: targetID,
      rows: selection.expected_rows,
      work_units: units,
      selection: { completion_scope: "target_partition" },
      unused_inputs: [],
    };
    const rowResults = selection.selected_rows.map((rowID) => {
      const result = resultByID.get(rowID);
      if (!result) throw new Error(`target evidence is missing observed row ${rowID}`);
      return result;
    });
    const execution = {
      status: rowResults.every((row) => successfulStates.has(row.terminal_state))
        ? "pass"
        : "fail",
      row_results: rowResults,
      duration_ms: rowResults.reduce((total, row) => total + row.duration_ms, 0),
    };
    const logs = units.map((unit) => ({
      work_unit_id: unit.work_unit_id,
      ...(observed.logs.get(unit.row_ids[0]) ?? {}),
    }));
    const accounting = buildTestEvidenceAccounting(
      plan,
      execution,
      logs,
      startedAt,
      finishedAt,
    );
    const prefix = `${targetID}/owners/${ownerID}`;
    const artifacts = {
      evidence_accounting: `${prefix}/test-evidence-accounting.json`,
      owner_summary: `${prefix}/test-owner-summary.json`,
      tool_run_summary: `${targetID}/tool-run-summary.json`,
    };
    const ownerSummary = buildTestOwnerSummary(plan, accounting, artifacts);
    const ownerDir = path.join(targetDir, "owners", ownerID);
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
  });
  if (selectedRows[0].runner === "playwright") {
    writeBrowserIndex(targetDir, targetID, runID, identity, selectedRows, shards);
  }
  return {
    status: shards.every((shard) => shard.status === "pass") ? "pass" : "fail",
    shards,
    reused: false,
  };
}
