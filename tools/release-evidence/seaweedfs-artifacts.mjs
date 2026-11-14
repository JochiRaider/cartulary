import path from "node:path";
import { parseStrictJSON, readRetainedArtifactBytes, redactValue, semanticJSONDigest, validateSchemaSync } from "../harness/contract/index.mjs";
import { validateWorkGraph } from "../harness/scheduler/work-graph/index.mjs";
import { atomicLocalFile } from "../harness/runtime/secure-local-files.mjs";
import { validateReleaseBom } from "./validate-release-sbom.mjs";

const maximumBytes = 16 * 1024 * 1024;
export const seaweedfsChecks = Object.freeze({
  occurrence_inventory: "source_policy",
  release_surface: "source_policy",
  dependency_boundary: "source_policy",
  inventory_classification: "source_policy",
  threat_policy: "contract_inventory",
  storage_ref_policy: "contract_inventory",
  backup_contracts: "contract_inventory",
  compatibility: "runtime",
  recovery: "runtime",
  redaction: "privacy",
});

export function readSeaweedfsJSON(root, relative, schemaID) {
  try {
    const bytes = readRetainedArtifactBytes(root, relative, maximumBytes);
    const value = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes));
    if (schemaID) validateSchemaSync(schemaID, value);
    return value;
  } catch (error) {
    // Neither parser details nor untrusted values/physical paths are diagnostics.
    throw new Error("SeaweedFS evidence admission failed", { cause: error });
  }
}

export function seaweedfsRunContext(environment = process.env) {
  const root = environment.CARTULARY_TEST_RESULTS_DIR;
  const runID = environment.CARTULARY_TEST_RUN_ID;
  const target = environment.CARTULARY_TEST_TARGET;
  if (typeof root !== "string" || !path.isAbsolute(root) ||
      typeof runID !== "string" || !/^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$/u.test(runID) ||
      target !== "seaweedfs-release-gate") {
    throw new Error("SeaweedFS evidence requires explicit current run context");
  }
  const runRoot = path.join(root, runID);
  const manifest = readSeaweedfsJSON(runRoot, "run-manifest.json", "cartulary.harness_run_manifest.v3");
  if (manifest.run_id !== runID) throw new Error("SeaweedFS run identity mismatch");
  return { runRoot, manifest, target };
}

function comparableTimestamp(value) {
  const milliseconds = Date.parse(value);
  if (!Number.isFinite(milliseconds) || new Date(milliseconds).toISOString().slice(0, 19) !== value.slice(0, 19)) return null;
  // Preserve producer nanoseconds when checking order; Date truncates them.
  return `${value.slice(0, 19)}.${(value.slice(19, -1).replace(/^\./u, "")).padEnd(9, "0")}`;
}

export function validateSeaweedfsCompatibility(report, unit) {
  try {
    validateSchemaSync("cartulary.seaweedfs_s3_compatibility_report.v1", report);
    validateSchemaSync("cartulary.harness_unit_result.v1", unit);
  } catch (error) { throw new Error("SeaweedFS compatibility shape invalid", { cause: error }); }
  if (unit.unit_id !== "operation:harness.seaweedfs_compatibility" || unit.status !== "passed" ||
      unit.exit_code !== 0 || unit.signal !== null || unit.failure_class !== null ||
      unit.failure_reason !== null || unit.missing_outputs.length !== 0 ||
      !unit.evidence_outputs.includes("seaweedfs-compatibility/object-store-compatibility-report.json")) {
    throw new Error("SeaweedFS compatibility execution did not pass");
  }
  const start = comparableTimestamp(report.started_at), end = comparableTimestamp(report.completed_at);
  if (start === null || end === null || end < start ||
      report.result !== "pass" || report.forbidden_skip_rows.length !== 0 ||
      new Set(report.cases.map(item => item.case_id)).size !== 14 ||
      report.cases.some(item => item.status !== "pass" || item.reason_code !== null)) {
    throw new Error("SeaweedFS compatibility observations did not pass");
  }
}

export function readCurrentSeaweedfsCompatibility({ runRoot }) {
  const reportRef = "seaweedfs-compatibility/object-store-compatibility-report.json";
  const unitRef = "unit-results/operation-harness.seaweedfs_compatibility.json";
  const report = readSeaweedfsJSON(runRoot, reportRef, "cartulary.seaweedfs_s3_compatibility_report.v1");
  const unit = readSeaweedfsJSON(runRoot, unitRef, "cartulary.harness_unit_result.v1");
  validateSeaweedfsCompatibility(report, unit);
  return { report, unit, evidence_refs: [unitRef, reportRef] };
}

const recoveryRowID = "module.recovery.process.canonical_operator_process_evidence_maps_the_imp_9808fdd4e9";

// The graph is published before execution. Do not wait for a whole-run terminal
// summary from inside one of that run's own dependent operations.
export function readCurrentSeaweedfsRecovery({ runRoot, manifest, target }) {
  try {
    const context = readSeaweedfsJSON(runRoot, "execution-context.json", "cartulary.harness_execution_context.v3");
    const graph = validateWorkGraph(context.graph);
    if (context.manifest_digest !== semanticJSONDigest(manifest) || graph.graph_digest !== manifest.graph_digest) {
      throw new Error("current graph binding differs");
    }
    const gateID = `operation:harness.${target.replaceAll("-", "_")}`;
    const units = new Map(graph.units.map(value => [value.unit_id, value]));
    const gate = units.get(gateID);
    const selected = graph.units.filter(value => value.row_bindings.some(row => row.row_id === recoveryRowID && row.runner === "go"));
    if (!gate || selected.length !== 1) throw new Error("required recovery selection absent");
    const expected = selected[0], ancestors = new Set(), pending = [...gate.needs];
    while (pending.length) {
      const id = pending.pop();
      if (ancestors.has(id)) continue;
      ancestors.add(id); pending.push(...units.get(id).needs);
    }
    if (!ancestors.has(expected.unit_id) || expected.cache_policy !== "none") throw new Error("recovery is not a fresh prerequisite");
    const unitRef = `unit-results/${expected.unit_id.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-")}.json`;
    const rowRef = `rows/${recoveryRowID}.json`;
    const unit = readSeaweedfsJSON(runRoot, unitRef, "cartulary.harness_unit_result.v1");
    const row = readSeaweedfsJSON(runRoot, rowRef, "cartulary.harness_row_result.v6");
    if (unit.unit_id !== expected.unit_id || unit.semantic_digest !== expected.semantic_digest ||
        unit.status !== "passed" || unit.exit_code !== 0 || unit.signal !== null || unit.failure_class !== null ||
        unit.failure_reason !== null || unit.missing_outputs.length ||
        expected.current_run_evidence_outputs.some(ref => !unit.evidence_outputs.includes(ref))) {
      throw new Error("recovery unit incomplete");
    }
    const runStart = comparableTimestamp(manifest.started_at), start = comparableTimestamp(row.started_at), end = comparableTimestamp(row.finished_at);
    if (row.row_id !== recoveryRowID || row.runner !== "go" || row.terminal_state !== "passed" || row.exit_code !== 0 ||
        row.failure_class !== null || row.failure_reason !== null || row.failure_diagnostic !== null ||
        runStart === null || start === null || end === null || start < runStart || end < start) {
      throw new Error("recovery row unsuccessful or stale");
    }
    return { row, unit, evidence_refs: [unitRef, rowRef] };
  } catch (error) { throw new Error("SeaweedFS recovery admission failed", { cause: error }); }
}

export function readSeaweedfsInventories(repositoryRoot) {
  const sbomRef = ".cartulary/release-artifacts/sbom.cyclonedx.json";
  const licenseRef = ".cartulary/release-artifacts/license-report.json";
  try {
    const sbom = readSeaweedfsJSON(repositoryRoot, sbomRef);
    const licenseReport = readSeaweedfsJSON(repositoryRoot, licenseRef, "cartulary.license_report.v2");
    const digest = validateReleaseBom(sbom);
    if (digest !== licenseReport.semantic_input_digest) throw new Error("inventory identity mismatch");
    return { sbom, licenseReport, evidence_refs: [sbomRef, licenseRef] };
  } catch (error) { throw new Error("SeaweedFS inventory admission failed", { cause: error }); }
}

const physicalFields = new Set(["bucket", "source_bucket", "target_bucket", "s3_bucket", "object_key", "source_key", "target_key", "storage_ref", "endpoint", "source_endpoint", "target_endpoint"]);
const secretMarkers = ["seaweedfs-secret", "min" + "io-secret", "supersecret", "secret-access-key"];
const publicMarkers = ["s3://", "private-bucket", "source.internal", "target.internal"];

// Scan the exact typed snapshots already acquired by their current-run reader.
// Boundary classification comes from the owning caller, never filename substrings.
export function scanSeaweedfsPrivacy(artifacts) {
  if (!Array.isArray(artifacts) || artifacts.length === 0 ||
      new Set(artifacts.map(item => item.path)).size !== artifacts.length ||
      artifacts.some(item => !["public_or_shareable", "operator_private"].includes(item.boundary))) {
    throw new Error("SeaweedFS privacy selection invalid");
  }
  const findings = [];
  for (const artifact of artifacts) {
    const add = (code, json_path = null) => findings.push({ check_id: "redaction", code, path: artifact.path, json_path });
    if (artifact.value === null || typeof artifact.value !== "object") { add("current-artifact-missing"); continue; }
    const encoded = JSON.stringify(artifact.value);
    if (encoded !== JSON.stringify(redactValue(artifact.value)) || secretMarkers.some(marker => encoded.includes(marker))) add("raw-secret-value");
    if (artifact.boundary === "operator_private") continue;
    if (publicMarkers.some(marker => encoded.includes(marker))) add("public-storage-identifier");
    const visit = value => {
      if (Array.isArray(value)) { value.forEach(visit); return; }
      if (!value || typeof value !== "object") return;
      for (const [key, child] of Object.entries(value)) {
        // Only the known physical-field name is diagnostic. Arbitrary ancestor
        // keys can themselves contain private identifiers or secret values.
        if (physicalFields.has(key.toLowerCase()) && typeof child === "string" && child.trim()) add("public-raw-storage-field", key.toLowerCase());
        visit(child);
      }
    };
    visit(artifact.value);
  }
  return { status: findings.length ? "fail" : "pass", evidence_refs: artifacts.map(item => item.path), findings };
}

export function validateSeaweedfsSummary(summary) {
  validateSchemaSync("cartulary.seaweedfs_release_gate_summary.v3", summary);
  if (JSON.stringify(summary.checks.map(item => item.check_id)) !== JSON.stringify(Object.keys(seaweedfsChecks)) ||
      summary.checks.some(item => item.kind !== seaweedfsChecks[item.check_id] ||
        (item.kind === "runtime" && item.status === "pass" && item.evidence_refs.length === 0))) {
    throw new Error("SeaweedFS summary check closure invalid");
  }
  const passed = summary.checks.every(item => item.status === "pass") && summary.findings.length === 0;
  if (summary.status !== (passed ? "pass" : "blocked")) throw new Error("SeaweedFS summary outcome invalid");
}

export function buildSeaweedfsSummary({ manifest, target, outcomes, findings = [], occurrences = [] }) {
  const checks = Object.entries(seaweedfsChecks).map(([check_id, kind]) => ({
    check_id, kind, status: outcomes[check_id]?.status ?? "unqualified",
    evidence_refs: outcomes[check_id]?.evidence_refs ?? [],
  }));
  const summary = {
    schema_id: "cartulary.seaweedfs_release_gate_summary.v3",
    run_id: manifest.run_id, target, source_digest: manifest.source_digest,
    toolchain_digest: manifest.toolchain_digest, graph_digest: manifest.graph_digest,
    status: checks.every(item => item.status === "pass") && findings.length === 0 ? "pass" : "blocked",
    checks,
    findings: findings.map(item => ({ check_id: item.check_id, code: item.code,
      path: item.path ?? null, line: item.line ?? null, column: item.column ?? null,
      json_path: item.json_path ?? null })),
    source_occurrences: occurrences.map(item => ({ path: item.path, line: item.line,
      column: item.column, matched_term: item.token, classification: item.classification,
      classification_rule_id: item.classification_rule_id ?? null })),
  };
  validateSeaweedfsSummary(summary);
  return summary;
}

export function publishSeaweedfsSummary({ runRoot, manifest, target }, summary) {
  validateSeaweedfsSummary(summary);
  for (const key of ["run_id", "source_digest", "toolchain_digest", "graph_digest"]) {
    if (summary[key] !== manifest[key]) throw new Error("SeaweedFS publication identity mismatch");
  }
  if (summary.target !== target) throw new Error("SeaweedFS publication target mismatch");
  if (scanSeaweedfsPrivacy([{ path: "release-gate-summary.json", boundary: "public_or_shareable", value: summary }]).status !== "pass") {
    throw new Error("SeaweedFS publication contains private data");
  }
  const bytes = `${JSON.stringify(summary)}\n`;
  if (Buffer.byteLength(bytes) > maximumBytes) throw new Error("SeaweedFS summary exceeds byte limit");
  const relative = `${target}/artifacts/release-gate-summary.json`;
  // A fresh run owns one immutable report. Existing reports are never replaced.
  atomicLocalFile(path.join(runRoot, relative), bytes);
  return relative;
}
