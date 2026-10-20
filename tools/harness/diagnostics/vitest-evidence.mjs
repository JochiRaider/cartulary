import path from "node:path";
import { parseStrictJSON, redactValue, semanticJSONDigest, validateSchemaSync } from "../contract/index.mjs";
import { atomicLocalFile, listLocalDirectory, readLocalFile } from "../runtime/secure-local-files.mjs";
import { diagnosticFailure, normalizeVitestObservations, reconcileVitestReport } from "./vitest-observations.mjs";

const fileSchema = "cartulary.vitest_file_checkpoint.v1";
const dispositionSchema = "cartulary.vitest_invocation_disposition.v1";
const maxArtifactBytes = 16 * 1024 * 1024;
const maxCheckpointBytes = 2 * maxArtifactBytes;
const fileName = (index) => `file-${index}.json`;
const evidenceDirectory = (reportFile) => path.join(path.dirname(reportFile), "vitest-evidence");
const accountingError = (message) => diagnosticFailure(message, "scheduler_accounting_error");

export function validateVitestRequest(request) {
  validateSchemaSync("cartulary.vitest_invocation.v1", request);
  const files = new Set(), rows = new Set();
  for (const file of request.files) {
    if (files.has(file.file)) throw accountingError("Duplicate Vitest file selection");
    files.add(file.file);
    const titles = new Set();
    for (const row of file.rows) {
      if (rows.has(row.row_id)) throw accountingError("Duplicate Vitest row selection");
      rows.add(row.row_id);
      for (const title of row.selectors) {
        if (titles.has(title)) throw accountingError("Vitest rows select overlapping test identities");
        titles.add(title);
      }
    }
  }
}

function readJSON(file, { optional = false, maximum = maxArtifactBytes, mutable = false } = {}) {
  try {
    const bytes = readLocalFile(file, { maximum, privateFile: false, allowReplacement: mutable });
    return { value: parseStrictJSON(bytes.toString("utf8")), bytes: bytes.length };
  } catch (error) {
    if (optional && error.code === "ENOENT") return undefined;
    throw diagnosticFailure("Vitest evidence is missing, unsafe, oversized, or malformed");
  }
}

function encoded(value, maximum = maxArtifactBytes) {
  validateSchemaSync(value.schema_id, value);
  const bytes = `${JSON.stringify(value)}\n`;
  if (Buffer.byteLength(bytes) > maximum) throw diagnosticFailure("Vitest evidence exceeds its bounded artifact budget");
  return bytes;
}

function binding(root, request, reportFile, runID, invocationID) {
  return { invocation_id: invocationID, run_id: runID, request_digest: semanticJSONDigest(request),
    artifact_root: path.relative(root, path.dirname(reportFile)).replaceAll("\\", "/") };
}

function assertBinding(value, expected) {
  for (const [key, field] of Object.entries(expected)) {
    if (value[key] !== field) throw accountingError("Vitest evidence belongs to a different invocation, run, or selection");
  }
}

function assertNormalized(details) {
  const expected = normalizeVitestObservations({ ...details, failures: [] }, {
    invocationID: details.invocation_id, runID: details.run_id, runnerJSON: details.runner_json,
    stdoutLog: details.stdout_log, stderrLog: details.stderr_log });
  if (semanticJSONDigest(expected) !== semanticJSONDigest(details)) throw accountingError("Vitest diagnostic causes contradict their observations");
}

function assertSelection(details, request) {
  const files = new Map(request.files.map((file) => [file.file, new Set(file.rows.flatMap((row) => row.selectors))]));
  for (const file of details.completed_files) {
    if (file.project !== request.project || !files.has(file.owner_path)) throw accountingError("Vitest completed an unselected file");
  }
  for (const item of details.observations) {
    if (item.scope === "run") continue;
    const titles = files.get(item.owner_path);
    if (item.project !== request.project || !titles || item.scope === "test" && ["passed", "failed"].includes(item.status) && !titles.has(item.title)) {
      throw accountingError("Vitest observed an unselected identity");
    }
  }
}

function assertFile(checkpoint, expected, request, index, root) {
  validateSchemaSync(fileSchema, checkpoint);
  assertBinding(checkpoint, expected);
  const file = request.files[index];
  if (!file || checkpoint.file_index !== index) throw accountingError("Vitest checkpoint ordinal is outside its exact request");
  const { details, report } = checkpoint;
  for (const value of [details, report]) assertBinding(value, { invocation_id: expected.invocation_id, run_id: expected.run_id });
  if (details.final_complete || details.completed_files.length !== 1 ||
      details.completed_files[0].project !== request.project || details.completed_files[0].owner_path !== file.file ||
      details.observations.some((item) => item.scope === "run" || item.project !== request.project || item.owner_path !== file.file)) {
    throw accountingError("Vitest checkpoint does not describe one independently settled selected file");
  }
  assertSelection(details, { ...request, files: [file] });
  assertNormalized(details);
  reconcileVitestReport(details, report, root);
}

function assemble(base, files, finalComplete) {
  const details = { ...base, observations: [], failures: [], completed_files: [], final_complete: finalComplete };
  const report = { schema_id: "cartulary.vitest_runner_reports.v1", invocation_id: base.invocation_id, run_id: base.run_id, reports: [] };
  for (const file of files) {
    details.observations.push(...file.details.observations);
    details.failures.push(...file.details.failures);
    details.completed_files.push(...file.details.completed_files);
    report.reports.push(...file.report.reports);
    if (file.details.run_status === "failed" && details.run_status !== "interrupted") details.run_status = "failed";
  }
  details.observations.push(...base.observations);
  details.failures.push(...base.failures);
  return { details, report };
}

// Owns the one invocation's durable evidence. Each file is reconciled and written
// independently; only final settlement assembles the cumulative terminal views.
export function createVitestEvidence({ root, request, reportFile, detailsFile, runID, invocationID, startedAt,
  publish = atomicLocalFile }) {
  validateVitestRequest(request);
  const identity = binding(root, request, reportFile, runID, invocationID);
  const directory = evidenceDirectory(reportFile);
  const dispositionFile = path.join(directory, "disposition.json");
  const context = { invocationID, runID, runnerJSON: path.relative(root, reportFile),
    stdoutLog: path.relative(root, path.join(path.dirname(reportFile), "stdout.log")),
    stderrLog: path.relative(root, path.join(path.dirname(reportFile), "stderr.log")) };
  const raw = { schema_id: "cartulary.vitest_failure_details.v3", invocation_id: invocationID, run_id: runID,
    run_status: "passed", generated_at: startedAt, runner_json: "", stdout_log: "", stderr_log: "",
    observations: [], failures: [], completed_files: [], final_complete: false };
  const files = [];
  let checkpointBytes = 0, commandFailure = null, globalDetails;
  const disposition = (terminal = null) => {
    globalDetails = redactValue(normalizeVitestObservations(raw, context));
    const value = { schema_id: dispositionSchema, ...identity, details: globalDetails,
      command_failure: commandFailure, terminal_published: Boolean(terminal),
      report_digest: terminal ? semanticJSONDigest(terminal.report) : null,
      details_digest: terminal ? semanticJSONDigest(terminal.details) : null };
    publish(dispositionFile, encoded(value), { replace: true, maximumReplacementBytes: maxArtifactBytes });
  };
  disposition();
  return {
    settleFile(index, capture, nativeReport) {
      if (index !== files.length) throw accountingError("Vitest file settlement is not in request order");
      const observations = capture.observations.filter((item) => item.scope !== "run");
      const details = normalizeVitestObservations({ ...capture, observations, final_complete: false,
        run_status: observations.some((item) => item.status === "failed") ? "failed" : "passed" }, context);
      const report = { schema_id: "cartulary.vitest_runner_reports.v1", invocation_id: invocationID, run_id: runID,
        reports: [{ project: request.project, file: request.files[index].file, report: nativeReport }] };
      const checkpoint = redactValue({ schema_id: fileSchema, ...identity, file_index: index, details, report });
      assertFile(checkpoint, identity, request, index, root);
      const bytes = encoded(checkpoint, maxCheckpointBytes);
      if (checkpointBytes + Buffer.byteLength(bytes) > maxCheckpointBytes) throw diagnosticFailure("Vitest invocation exceeds its checkpoint evidence budget");
      publish(path.join(directory, fileName(index)), bytes);
      checkpointBytes += Buffer.byteLength(bytes);
      files.push(checkpoint);
      return { details: checkpoint.details, report: checkpoint.report };
    },
    recordRunFailures(observations) {
      if (!observations.length) return;
      if (observations.some((item) => item.scope !== "run")) throw accountingError("Vitest invocation disposition contains a file observation");
      raw.observations.push(...observations);
      raw.run_status = "failed"; raw.generated_at = new Date().toISOString();
      disposition();
    },
    fail(failure) {
      commandFailure = { failure_class: failure.failure_class, failure_reason: failure.failure_reason };
      raw.run_status = "failed"; raw.generated_at = new Date().toISOString();
      disposition();
    },
    snapshot() { return { ...assemble(globalDetails, files, false), commandFailure }; },
    finish(signal) {
      raw.run_status = signal ? "interrupted" : raw.run_status;
      raw.generated_at = new Date().toISOString(); raw.final_complete = !signal;
      disposition();
      const terminal = assemble(globalDetails, files, !signal);
      reconcileVitestReport(terminal.details, terminal.report, root);
      publish(reportFile, encoded(terminal.report));
      publish(detailsFile, encoded(terminal.details));
      disposition(terminal);
      return { ...terminal, commandFailure };
    },
  };
}

// Exact-request enumeration is bounded; the filesystem never chooses selection.
// A terminal disposition commits the two unchanged terminal envelopes. Before
// that commit, independently durable file records remain sufficient to recover
// settled outcomes, including a crash before their row files were published.
export function readVitestEvidence({ root, request, reportFile, detailsFile, runID }) {
  validateVitestRequest(request);
  const directory = evidenceDirectory(reportFile);
  const found = readJSON(path.join(directory, "disposition.json"), { optional: true, mutable: true });
  if (!found) return undefined;
  const disposition = found.value;
  validateSchemaSync(dispositionSchema, disposition);
  const identity = binding(root, request, reportFile, runID, disposition.invocation_id);
  assertBinding(disposition, identity);
  assertBinding(disposition.details, { invocation_id: identity.invocation_id, run_id: runID });
  if (disposition.details.completed_files.length || disposition.details.observations.some((item) => item.scope !== "run")) {
    throw accountingError("Vitest invocation disposition contains file evidence");
  }
  assertNormalized(disposition.details);
  if (disposition.terminal_published) {
    const report = readJSON(reportFile).value, details = readJSON(detailsFile).value;
    if (semanticJSONDigest(report) !== disposition.report_digest || semanticJSONDigest(details) !== disposition.details_digest ||
        details.runner_json !== path.relative(root, reportFile) || details.final_complete !== disposition.details.final_complete) {
      throw accountingError("Vitest terminal envelopes disagree with their invocation disposition");
    }
    assertBinding(details, { invocation_id: identity.invocation_id, run_id: runID });
    validateSchemaSync("cartulary.vitest_failure_details.v3", details);
    assertSelection(details, request);
    assertNormalized(details);
    if (semanticJSONDigest(details.observations.filter((item) => item.scope === "run")) !== semanticJSONDigest(disposition.details.observations)) {
      throw accountingError("Vitest terminal observations disagree with their invocation disposition");
    }
    reconcileVitestReport(details, report, root);
    return { details, report, commandFailure: disposition.command_failure };
  }
  const entries = listLocalDirectory(directory);
  if (entries.length > request.files.length + 4) throw accountingError("Vitest checkpoint directory exceeds its selection bound");
  const selected = new Map();
  for (const entry of entries) {
    if (entry.name === "disposition.json" || /^\.publishing-[0-9a-f]{32}$/u.test(entry.name)) continue;
    const match = /^file-(0|[1-9][0-9]*)\.json$/u.exec(entry.name);
    if (!entry.isFile() || !match || Number(match[1]) >= request.files.length) throw accountingError("Unexpected Vitest checkpoint entry");
    selected.set(Number(match[1]), entry.name);
  }
  const files = [];
  let bytes = 0;
  for (let index = 0; index < selected.size; index++) {
    if (!selected.has(index)) throw accountingError("Vitest settled checkpoint sequence has a gap");
    const foundFile = readJSON(path.join(directory, selected.get(index)), { maximum: maxCheckpointBytes });
    bytes += foundFile.bytes;
    if (bytes > maxCheckpointBytes) throw diagnosticFailure("Vitest invocation exceeds its checkpoint evidence budget");
    assertFile(foundFile.value, identity, request, index, root);
    files.push(foundFile.value);
  }
  const result = assemble(disposition.details, files, false);
  validateSchemaSync(result.details.schema_id, result.details);
  reconcileVitestReport(result.details, result.report, root);
  return { ...result, commandFailure: disposition.command_failure };
}
