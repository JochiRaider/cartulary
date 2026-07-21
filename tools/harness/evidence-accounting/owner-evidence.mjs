import {
  lstatSync,
  readFileSync,
  realpathSync,
  statSync,
} from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import { loadTestCatalog } from "../test-catalog/test-catalog.mjs";
import { parseStrictJSON } from "../test-catalog/semantic-json.mjs";
import { buildSourceSnapshot } from "../owner-slice/source-snapshot.mjs";
import { loadOwnerAccountingSelection } from "./catalog-accounting.mjs";

const successfulStates = new Set(["passed", "skipped_authorized"]);
const terminalStates = new Set([
  "passed",
  "failed",
  "infrastructure_failed",
  "skipped_dependency",
  "cancelled",
  "skipped_authorized",
]);

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sortedUnique(values, label) {
  const sorted = [...values].sort(asciiCompare);
  if (new Set(sorted).size !== sorted.length) throw new Error(`${label} contains duplicates`);
  return sorted;
}

function identityFields(plan) {
  return {
    command_id: plan.command_id,
    run_id: plan.run_id,
    owner_id: plan.owner_id,
    selected_rows: [...plan.selected_rows],
    source_snapshot_digest: plan.source_snapshot_digest,
    catalog_semantic_digest: plan.catalog_semantic_digest,
    verification_semantic_digest: plan.verification_semantic_digest,
    runtime_profile_digest: plan.runtime_profile_digest,
    resource_profile_digest: plan.resource_profile_digest,
    fixture_profile_digest: plan.fixture_profile_digest,
  };
}

function countStates(observedRows) {
  const counts = {
    selected: observedRows.length,
    passed: 0,
    failed: 0,
    infrastructure_failed: 0,
    skipped_dependency: 0,
    cancelled: 0,
    skipped_authorized: 0,
  };
  for (const row of observedRows) counts[row.terminal_state] += 1;
  return counts;
}

function primaryFailure(observedRows) {
  const precedence = ["failed", "infrastructure_failed", "cancelled", "skipped_dependency"];
  for (const state of precedence) {
    const row = observedRows.find((entry) => entry.terminal_state === state);
    if (row) {
      return {
        row_id: row.row_id,
        failure_class: state === "failed" ? "product" : "artifact",
        failure_reason: row.failure?.failure_reason ?? state,
        exit_code: row.failure?.exit_code ?? (state === "failed" ? 10 : 11),
      };
    }
  }
  return null;
}

export function buildTestEvidenceAccounting(plan, execution, logs, startedAt, finishedAt) {
  const rowResults = [...execution.row_results].sort((left, right) => asciiCompare(left.row_id, right.row_id));
  const resultIDs = sortedUnique(rowResults.map((row) => row.row_id), "observed row results");
  if (JSON.stringify(resultIDs) !== JSON.stringify(plan.selected_rows)) {
    throw new Error("observed row results do not exactly match the selected row inventory");
  }
  const logByUnit = new Map(logs.map((entry) => [entry.work_unit_id, entry]));
  const unitByRow = new Map();
  for (const unit of plan.work_units) {
    for (const rowID of unit.row_ids) unitByRow.set(rowID, unit);
  }
  const expectedRows = plan.rows.map((row) => ({
    row_id: row.row_id,
    owner_id: row.owner_id,
    family_id: row.family_id,
    verification_ids: [...row.verification_ids],
    runner: row.runner,
    selector_digest: row.selector_digest,
    evidence_class: row.evidence_class,
    evidence_target_id: row.target_name,
    runtime_profile_id: row.runtime_profile_id,
    resource_profile_id: row.resource_profile_id,
    fixture_profile_id: row.fixture_profile_id,
  }));
  const observedRows = rowResults.map((result) => {
    if (!terminalStates.has(result.terminal_state)) {
      throw new Error(`row ${result.row_id} has unsupported terminal state ${result.terminal_state}`);
    }
    const unit = unitByRow.get(result.row_id);
    const log = logByUnit.get(unit.work_unit_id);
    const failure = successfulStates.has(result.terminal_state)
      ? null
      : {
          failure_class: result.terminal_state === "failed" ? "product" : "artifact",
          failure_reason: result.failure_reason ?? result.terminal_state,
          exit_code: result.exit_code || (result.terminal_state === "failed" ? 10 : 11),
        };
    return {
      row_id: result.row_id,
      terminal_state: result.terminal_state,
      logical_duration_ms: result.duration_ms,
      executed_duration_ms: result.duration_ms,
      attempts: [
        {
          attempt: result.attempt,
          terminal_state: result.terminal_state,
          exit_code: result.exit_code,
          duration_ms: result.duration_ms,
          artifact_refs: [log?.stdout_path, log?.stderr_path].filter(Boolean),
        },
      ],
      failure,
    };
  });
  return {
    schema_id: "cartulary.test_evidence_accounting.v1",
    ...identityFields(plan),
    target_id: plan.target,
    started_at: startedAt,
    finished_at: finishedAt,
    duration_ms: execution.duration_ms,
    status:
      execution.status === "pass" && observedRows.every((row) => successfulStates.has(row.terminal_state))
        ? "pass"
        : "fail",
    expected_rows: expectedRows,
    observed_rows: observedRows,
  };
}

export function buildTestOwnerSummary(plan, accounting, artifacts) {
  const counts = countStates(accounting.observed_rows);
  return {
    schema_id: "cartulary.test_owner_summary.v1",
    ...identityFields(plan),
    target_id: plan.target,
    started_at: accounting.started_at,
    finished_at: accounting.finished_at,
    duration_ms: accounting.duration_ms,
    status: accounting.status,
    completion_scope: plan.selection.completion_scope,
    counts,
    unused_inputs: [...plan.unused_inputs],
    primary_failure: primaryFailure(accounting.observed_rows),
    artifacts,
  };
}

export function deriveRequiredEvidencePartitions(root, ownerID) {
  const selection = loadOwnerAccountingSelection(root, { ownerID });
  const partitions = new Map([["test-slice", [...selection.selected_rows]]]);
  for (const row of selection.expected_rows) {
    const values = partitions.get(row.target_name) ?? [];
    values.push(row.row_id);
    partitions.set(row.target_name, values);
  }
  return new Map(
    [...partitions.entries()]
      .map(([targetID, rowIDs]) => [targetID, [...new Set(rowIDs)].sort(asciiCompare)])
      .sort(([left], [right]) => asciiCompare(left, right)),
  );
}

export class EvidenceAuditUsageError extends Error {
  constructor(message) {
    super(message);
    this.name = "EvidenceAuditUsageError";
    this.exitCode = 2;
  }
}

function usage(message) {
  throw new EvidenceAuditUsageError(message);
}

function readManifest(root, ownerID, manifestPath) {
  if (!manifestPath || String(manifestPath).includes("\0")) usage("EVIDENCE_ROOTS_FILE is required");
  const file = path.resolve(root, manifestPath);
  let stat;
  try {
    stat = lstatSync(file);
  } catch {
    usage("EVIDENCE_ROOTS_FILE must reference an existing file");
  }
  if (!stat.isFile() || stat.isSymbolicLink()) usage("EVIDENCE_ROOTS_FILE must be a non-symlink regular file");
  const manifest = parseStrictJSON(readFileSync(file, "utf8"), file);
  try {
    validateSchemaSync("cartulary.test_evidence_root_manifest.v1", manifest);
  } catch (error) {
    usage(`EVIDENCE_ROOTS_FILE is schema-invalid: ${error.message}`);
  }
  if (manifest.schema_id !== "cartulary.test_evidence_root_manifest.v1") usage("EVIDENCE_ROOTS_FILE has an unsupported schema_id");
  if (manifest.owner_id !== ownerID) usage("EVIDENCE_ROOTS_FILE owner_id does not match OWNER");
  if (!Array.isArray(manifest.entries) || manifest.entries.length === 0) usage("EVIDENCE_ROOTS_FILE entries must be non-empty");
  const targets = [];
  for (const entry of manifest.entries) {
    if (!entry || typeof entry.target_id !== "string" || typeof entry.run_root !== "string") {
      usage("EVIDENCE_ROOTS_FILE entries must contain target_id and run_root strings");
    }
    if (!/^[a-z][a-z0-9-]*$/u.test(entry.target_id) || entry.run_root.trim() === "" || entry.run_root.includes("\0")) {
      usage("EVIDENCE_ROOTS_FILE contains an invalid entry");
    }
    targets.push(entry.target_id);
  }
  if (JSON.stringify(targets) !== JSON.stringify([...targets].sort(asciiCompare))) {
    usage("EVIDENCE_ROOTS_FILE entries must be ASCII-sorted by target_id");
  }
  if (new Set(targets).size !== targets.length) usage("EVIDENCE_ROOTS_FILE contains duplicate target_id entries");
  return manifest;
}

function safeRunRoot(root, rawRoot) {
  const candidate = path.resolve(root, rawRoot);
  let stat;
  try {
    stat = lstatSync(candidate);
  } catch {
    throw new Error("run_root does not exist");
  }
  if (!stat.isDirectory() || stat.isSymbolicLink()) throw new Error("run_root must be a non-symlink directory");
  if (realpathSync(candidate) !== candidate) throw new Error("run_root contains a symlinked path component");
  const mode = statSync(candidate).mode;
  if ((mode & 0o002) !== 0 && (mode & 0o1000) === 0) {
    throw new Error("run_root is world-writable without sticky bit");
  }
  return candidate;
}

function readAccountingArtifact(runRoot, targetID, ownerID) {
  const expected = path.join(runRoot, targetID, "owners", ownerID, "test-evidence-accounting.json");
  let resolved;
  try {
    const stat = lstatSync(expected);
    if (!stat.isFile() || stat.isSymbolicLink()) throw new Error("artifact is not a non-symlink regular file");
    resolved = realpathSync(expected);
  } catch (error) {
    throw new Error(`accounting artifact is unavailable: ${error.message}`);
  }
  if (!resolved.startsWith(`${runRoot}${path.sep}`)) throw new Error("accounting artifact escapes run_root");
  return {
    artifact: parseStrictJSON(readFileSync(resolved, "utf8"), resolved),
    path: resolved,
  };
}

function expectedIdentity(root, ownerID, rowIDs) {
  const selection = loadOwnerAccountingSelection(root, { ownerID, rowIDs });
  return {
    selected_rows: selection.selected_rows,
    catalog_semantic_digest: selection.catalog_semantic_digest,
    verification_semantic_digest: selection.verification_semantic_digest,
    runtime_profile_digest: selection.runtime_profile_digest,
    resource_profile_digest: selection.resource_profile_digest,
    fixture_profile_digest: selection.fixture_profile_digest,
  };
}

function authorizedSkip(catalog, rowID, timestamp) {
  const row = catalog.rowByID.get(rowID);
  if (!row) return false;
  return row.verification_ids.some((verificationID) => {
    const policy = catalog.verification.verificationByID.get(verificationID)?.verification?.skip_policy;
    return policy?.mode === "authorize" &&
      policy.owner_id === row.owner_id &&
      Date.parse(policy.expires_at) > Date.parse(timestamp) &&
      String(policy.approval_ref ?? "").trim() !== "";
  });
}

function artifactReasons(artifact, expected, common, targetID, ownerID, catalog, timestamp) {
  try {
    validateSchemaSync("cartulary.test_evidence_accounting.v1", artifact);
  } catch {
    return ["schema_validation_failed"];
  }
  const reasons = [];
  if (artifact.schema_id !== "cartulary.test_evidence_accounting.v1") reasons.push("unsupported_schema");
  if (artifact.owner_id !== ownerID) reasons.push("owner_mismatch");
  if (artifact.target_id !== targetID) reasons.push("target_mismatch");
  for (const field of ["source_snapshot_digest", "catalog_semantic_digest", "verification_semantic_digest"]) {
    if (artifact[field] !== common[field]) reasons.push(`${field}_mismatch`);
  }
  for (const field of ["selected_rows", "runtime_profile_digest", "resource_profile_digest", "fixture_profile_digest"]) {
    if (JSON.stringify(artifact[field]) !== JSON.stringify(expected[field])) reasons.push(`${field}_mismatch`);
  }
  const expectedIDs = artifact.expected_rows?.map((row) => row.row_id) ?? [];
  const observedIDs = artifact.observed_rows?.map((row) => row.row_id) ?? [];
  if (new Set(expectedIDs).size !== expectedIDs.length || JSON.stringify(expectedIDs) !== JSON.stringify(expected.selected_rows)) {
    reasons.push("expected_row_inventory_mismatch");
  }
  if (new Set(observedIDs).size !== observedIDs.length || JSON.stringify(observedIDs) !== JSON.stringify(expected.selected_rows)) {
    reasons.push("observed_row_inventory_mismatch");
  }
  if ((artifact.observed_rows ?? []).some((row) => !successfulStates.has(row.terminal_state))) {
    reasons.push("unsuccessful_terminal_record");
  }
  if (
    (artifact.observed_rows ?? []).some(
      (row) => row.terminal_state === "skipped_authorized" && !authorizedSkip(catalog, row.row_id, timestamp),
    )
  ) {
    reasons.push("invalid_skip_authorization");
  }
  if (artifact.status !== "pass") reasons.push("artifact_status_not_pass");
  return [...new Set(reasons)].sort(asciiCompare);
}

export function auditOwnerEvidence(root, { ownerID, manifestPath, timestamp = new Date().toISOString() }) {
  const catalog = loadTestCatalog(root);
  const owner = catalog.registry.owners.find((entry) => entry.owner_id === ownerID && entry.status === "active");
  if (!owner) usage(`unknown active test owner ${ownerID}`);
  const manifest = readManifest(root, ownerID, manifestPath);
  const required = deriveRequiredEvidencePartitions(root, ownerID);
  const knownTargets = new Set([
    "service-backed-test-slice",
    ...required.keys(),
    ...parseStrictJSON(
      readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"),
      "tools/task_surface_manifest.json",
    ).targets.map((entry) => entry.name),
  ]);
  for (const entry of manifest.entries) {
    if (!knownTargets.has(entry.target_id)) usage(`unknown evidence target ${entry.target_id}`);
  }
  const supplied = new Map(manifest.entries.map((entry) => [entry.target_id, entry]));
  const sliceEntry = supplied.get("test-slice");
  const leafEvidence = new Map(
    [...required].filter(([targetID]) => targetID !== "test-slice"),
  );
  const requiredEvidence = sliceEntry
    ? new Map([["test-slice", required.get("test-slice")]])
    : leafEvidence;
  const missing = [...requiredEvidence.keys()].filter((targetID) => !supplied.has(targetID));
  if (missing.length > 0) usage(`missing required evidence targets: ${missing.join(", ")}`);

  const snapshot = buildSourceSnapshot(root);
  const common = {
    source_snapshot_digest: snapshot.digest,
    catalog_semantic_digest: catalog.summary.catalog_semantic_digest,
    verification_semantic_digest: catalog.summary.verification_semantic_digest,
  };
  const acceptedArtifacts = [];
  const rejectedArtifacts = [];
  const requiredTargets = [];
  const unusedInputs = [];
  for (const entry of manifest.entries) {
    if (!requiredEvidence.has(entry.target_id)) {
      unusedInputs.push({ ...entry });
      continue;
    }
    const rowIDs = requiredEvidence.get(entry.target_id);
    requiredTargets.push({ target_id: entry.target_id, run_root: entry.run_root, row_ids: rowIDs });
    let runRoot;
    let loaded;
    try {
      runRoot = safeRunRoot(root, entry.run_root);
      loaded = readAccountingArtifact(runRoot, entry.target_id, ownerID);
      const expected = expectedIdentity(root, ownerID, rowIDs);
      const reasons = artifactReasons(
        loaded.artifact,
        expected,
        common,
        entry.target_id,
        ownerID,
        catalog,
        timestamp,
      );
      if (reasons.length > 0) {
        rejectedArtifacts.push({ target_id: entry.target_id, run_root: entry.run_root, artifact_path: path.relative(root, loaded.path).replaceAll("\\", "/"), reasons });
      } else {
        acceptedArtifacts.push({ target_id: entry.target_id, run_root: entry.run_root, artifact_path: path.relative(root, loaded.path).replaceAll("\\", "/"), row_ids: rowIDs });
      }
    } catch (error) {
      rejectedArtifacts.push({ target_id: entry.target_id, run_root: entry.run_root, artifact_path: "", reasons: [error.message] });
    }
  }
  const status = rejectedArtifacts.length === 0 ? "pass" : "fail";
  return {
    schema_id: "cartulary.test_evidence_audit_summary.v1",
    command_id: "cartulary.harness.command.test_evidence_audit.v1",
    owner_id: ownerID,
    status,
    started_at: timestamp,
    finished_at: timestamp,
    duration_ms: 0,
    compatibility: common,
    required_targets: requiredTargets,
    unused_inputs: unusedInputs,
    accepted_artifacts: acceptedArtifacts,
    rejected_artifacts: rejectedArtifacts,
    counts: {
      active_rows: catalog.rows.filter((row) => row.owner_id === ownerID).length,
      required_target_partitions: requiredEvidence.size,
      accepted_target_partitions: acceptedArtifacts.length,
      rejected_target_partitions: rejectedArtifacts.length,
    },
  };
}
