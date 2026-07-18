import { createHash } from "node:crypto";
import path from "node:path";

import { loadOwnerAccountingSelection } from "../evidence-accounting/catalog-accounting.mjs";
import { loadTestCatalog } from "../test-catalog/test-catalog.mjs";
import { semanticJSONDigest } from "../test-catalog/semantic-json.mjs";
import { buildSourceSnapshot } from "./source-snapshot.mjs";

const ownerIDPattern = /^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}$/u;
const rowIDPattern = /^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,127}_[0-9a-f]{10}$/u;

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

export class OwnerSliceUsageError extends Error {
  constructor(message) {
    super(message);
    this.name = "OwnerSliceUsageError";
    this.exitCode = 2;
  }
}

function usage(message) {
  throw new OwnerSliceUsageError(message);
}

function positiveWorker(value, name, fallback) {
  const normalized = value === undefined ? String(fallback) : String(value).trim();
  if (!/^(?:[1-9]|1[0-6])$/u.test(normalized)) usage(`${name} must be an integer from 1 through 16`);
  return Number.parseInt(normalized, 10);
}

function parseRows(raw, provided) {
  if (!provided) return null;
  const source = String(raw ?? "");
  if (source.trim() === "") usage("ROWS must not be empty when supplied");
  const tokens = source.split(",").map((token) => token.trim());
  if (tokens.some((token) => token === "")) usage("ROWS must not contain empty tokens");
  if (tokens.some((token) => !rowIDPattern.test(token))) usage("ROWS contains a malformed row ID");
  if (new Set(tokens).size !== tokens.length) usage("ROWS must not contain duplicate row IDs");
  return [...tokens].sort(asciiCompare);
}

function profileByID(catalog, kind, profileID) {
  const profile = catalog.profiles.semantic[kind].find((entry) => entry.id === profileID);
  if (!profile) throw new Error(`unresolved ${kind} profile ${profileID}`);
  return profile;
}

function targetForGoFamily(familyID) {
  const family = familyID.split(".").at(-1);
  if (["engine", "fixtures", "support_unit", "unit"].includes(family)) return "backend-unit";
  if (family === "store") return "backend-store";
  if (family === "process") return "backend-process";
  return "backend-integration";
}

function targetForRow(row, expectedRow) {
  if (row.runner === "go") return targetForGoFamily(row.family_id);
  if (expectedRow.target_name) return expectedRow.target_name;
  throw new Error(`catalog row ${row.row_id} has no execution target`);
}

function unitID(key, rowIDs) {
  const digest = createHash("sha256").update(`${key}\0${rowIDs.join("\0")}`).digest("hex").slice(0, 12);
  const slug = key.replaceAll(/[^a-z0-9]+/gu, "_").replaceAll(/^_|_$/gu, "").slice(0, 48);
  return `unit.${slug}.${digest}`;
}

function runnerTimeoutSeconds(runner) {
  if (runner === "playwright") return 900;
  if (runner === "go" || runner === "vitest") return 600;
  return 300;
}

function rowRecord(row, expectedRow, targetName) {
  return {
    ...expectedRow,
    target_name: targetName,
    selector: row.selector,
  };
}

export function resolveOwnerSliceSelection(root, options) {
  const ownerID = String(options.ownerID ?? "").trim();
  if (!ownerID) usage("OWNER is required");
  if (!ownerIDPattern.test(ownerID)) usage("OWNER is malformed");
  if (!new Set(["all", "service_backed"]).has(options.dependencyScope)) {
    throw new Error(`unsupported owner dependency scope ${options.dependencyScope}`);
  }
  const requestedRows = parseRows(options.rows, options.rowsProvided === true);
  const vitestWorkers = positiveWorker(options.vitestWorkers, "VITEST_MAX_WORKERS", 4);
  const playwrightWorkers = positiveWorker(options.playwrightWorkers, "PLAYWRIGHT_WORKERS", 3);
  if (options.jsonValue !== undefined && !["", "1"].includes(String(options.jsonValue))) {
    usage("JSON accepts only exact 1 or an empty value");
  }
  if (String(options.jsonValue ?? "") === "1" && process.env.CARTULARY_OUTPUT_MODE === "machine") {
    usage("JSON=1 cannot be combined with CARTULARY_OUTPUT_MODE=machine");
  }

  let accounting;
  try {
    accounting = loadOwnerAccountingSelection(root, { ownerID, rowIDs: requestedRows });
  } catch (error) {
    usage(error.message);
  }
  const catalog = loadTestCatalog(root);
  let selectedRows = accounting.selected_rows.map((rowID) => catalog.rowByID.get(rowID));
  if (options.dependencyScope === "service_backed") {
    const nonService = selectedRows.filter((row) => {
      const runtime = profileByID(catalog, "runtime_profiles", row.runtime_profile_id);
      return runtime.managed_service_ids.length === 0;
    });
    if (requestedRows !== null && nonService.length > 0) {
      usage(`service-backed-test-slice row does not require managed services: ${nonService[0].row_id}`);
    }
    selectedRows = selectedRows.filter((row) => {
      const runtime = profileByID(catalog, "runtime_profiles", row.runtime_profile_id);
      return runtime.managed_service_ids.length > 0;
    });
  }
  if (selectedRows.length === 0) usage(`owner ${ownerID} resolves to zero executable rows`);
  selectedRows.sort((left, right) => asciiCompare(left.row_id, right.row_id));
  const selectedSet = new Set(selectedRows.map((row) => row.row_id));
  const expectedByID = new Map(accounting.expected_rows.map((row) => [row.row_id, row]));
  const records = selectedRows.map((row) => rowRecord(row, expectedByID.get(row.row_id), targetForRow(row, expectedByID.get(row.row_id))));

  const groups = new Map();
  for (const record of records) {
    const key = [
      record.runner,
      record.target_name,
      record.runtime_profile_id,
      record.resource_profile_id,
      record.fixture_profile_id,
    ].join("\0");
    const values = groups.get(key) ?? [];
    values.push(record);
    groups.set(key, values);
  }
  const workUnits = [...groups.entries()].map(([key, rows]) => {
    rows.sort((left, right) => asciiCompare(left.row_id, right.row_id));
    const runtime = profileByID(catalog, "runtime_profiles", rows[0].runtime_profile_id);
    const resource = profileByID(catalog, "resource_profiles", rows[0].resource_profile_id);
    return {
      work_unit_id: unitID(key, rows.map((row) => row.row_id)),
      runner: rows[0].runner,
      target_name: rows[0].target_name,
      row_ids: rows.map((row) => row.row_id),
      runtime_profile_id: rows[0].runtime_profile_id,
      resource_profile_id: rows[0].resource_profile_id,
      fixture_profile_id: rows[0].fixture_profile_id,
      managed_service_ids: [...runtime.managed_service_ids],
      resource_claims: resource.resource_claims,
      dependencies: [],
      expected_artifacts: ["cartulary.test_evidence_accounting.v1"],
      timeout_seconds: runnerTimeoutSeconds(rows[0].runner),
    };
  }).sort((left, right) => asciiCompare(left.work_unit_id, right.work_unit_id));

  const resolvedRows = records.map((row) => row.row_id);
  return {
    catalog,
    accounting,
    selection: {
      selection_mode: requestedRows === null ? "default_owner" : "exact_rows",
      owner_id: ownerID,
      dependency_scope: options.dependencyScope,
      completion_scope: requestedRows === null ? "full_owner" : "selected_subset",
      requested_row_ids: requestedRows ?? [],
      resolved_row_ids: resolvedRows,
    },
    rows: records,
    workUnits,
    workers: { vitest: vitestWorkers, playwright: playwrightWorkers },
    unusedInputs: [
      ...(records.some((row) => row.runner === "vitest") ? [] : ["VITEST_MAX_WORKERS"]),
      ...(records.some((row) => row.runner === "playwright") ? [] : ["PLAYWRIGHT_WORKERS"]),
    ].sort(asciiCompare),
    selectedSet,
  };
}

export function buildOwnerSlicePlan(root, options) {
  const resolvedRoot = path.resolve(root);
  const resolved = resolveOwnerSliceSelection(resolvedRoot, options);
  const snapshot = buildSourceSnapshot(resolvedRoot);
  const timestamp = options.timestamp ?? new Date().toISOString();
  const runtimeProfiles = [...new Set(resolved.rows.map((row) => row.runtime_profile_id))]
    .sort(asciiCompare)
    .map((id) => profileByID(resolved.catalog, "runtime_profiles", id));
  const resourceProfiles = [...new Set(resolved.rows.map((row) => row.resource_profile_id))]
    .sort(asciiCompare)
    .map((id) => profileByID(resolved.catalog, "resource_profiles", id));
  const fixtureProfiles = [...new Set(resolved.rows.map((row) => row.fixture_profile_id))]
    .sort(asciiCompare)
    .map((id) => profileByID(resolved.catalog, "fixture_profiles", id));
  const finalizers = [
    {
      finalizer_id: "finalizer.owner_slice_cleanup",
      kind: "cleanup",
      dependencies: resolved.workUnits.map((unit) => unit.work_unit_id).sort(asciiCompare),
    },
  ];
  const expectedArtifacts = [
    "cartulary.test_evidence_accounting.v1",
    "cartulary.test_owner_summary.v1",
    "cartulary.test_slice_scheduler_summary.v1",
    "cartulary.tool_run_summary.v4",
  ];
  const planSemanticDigest = semanticJSONDigest({
    command_id: options.commandID,
    target: options.target,
    owner_id: resolved.selection.owner_id,
    selection: resolved.selection,
    workers: resolved.workers,
    rows: resolved.rows,
    work_units: resolved.workUnits,
    finalizers,
    expected_artifacts: expectedArtifacts,
  });
  const schedulerSemanticDigest = semanticJSONDigest({
    scheduler_kind: "test_slice",
    capacity_profile: "test_slice_default",
    stop_on_first_failure: false,
    work_units: resolved.workUnits.map((unit) => ({
      id: unit.work_unit_id,
      needs: unit.dependencies,
      resource_claims: unit.resource_claims,
      timeout_seconds: unit.timeout_seconds,
    })),
    finalizers,
  });
  return {
    schema_id: "cartulary.test_slice_plan.v1",
    command_id: options.commandID,
    target: options.target,
    run_id: options.runID,
    owner_id: resolved.selection.owner_id,
    selected_rows: [...resolved.selection.resolved_row_ids],
    source_snapshot_digest: snapshot.digest,
    catalog_semantic_digest: resolved.catalog.summary.catalog_semantic_digest,
    verification_semantic_digest: resolved.catalog.summary.verification_semantic_digest,
    runtime_profile_digest: semanticJSONDigest(runtimeProfiles),
    resource_profile_digest: semanticJSONDigest(resourceProfiles),
    fixture_profile_digest: semanticJSONDigest(fixtureProfiles),
    plan_semantic_digest: planSemanticDigest,
    scheduler_semantic_digest: schedulerSemanticDigest,
    started_at: timestamp,
    finished_at: timestamp,
    duration_ms: 0,
    source_file_count: snapshot.file_count,
    selection: resolved.selection,
    workers: resolved.workers,
    unused_inputs: resolved.unusedInputs,
    rows: resolved.rows,
    work_units: resolved.workUnits,
    finalizers,
    expected_artifacts: expectedArtifacts,
  };
}
