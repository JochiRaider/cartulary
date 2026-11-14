import assert from "node:assert/strict";
import fs from "node:fs";
import { createHash } from "node:crypto";
import { syncBuiltinESMExports } from "node:module";
import path from "node:path";
import test from "node:test";
import { createScratch } from "../../harness/test-support/scratch.mjs";
import { buildEvidenceUnit, buildReaderEvidence, buildRowResult, buildUnitResult, writeEvidenceFile } from "../../harness/observability/tests/retained-fixture.mjs";
import { canonicalJSONString } from "../../harness/contract/index.mjs";
import { WorkGraphCompiler } from "../../harness/scheduler/work-graph/index.mjs";
import { buildSeaweedfsSummary, publishSeaweedfsSummary, readCurrentSeaweedfsCompatibility, readCurrentSeaweedfsRecovery, readSeaweedfsInventories, readSeaweedfsJSON, scanSeaweedfsPrivacy, seaweedfsRunContext,
  validateSeaweedfsCompatibility, validateSeaweedfsSummary } from "../seaweedfs-artifacts.mjs";

const target = "seaweedfs-release-gate";
const caseIDs = ["SWFS-COMP-001", "SWFS-COMP-002", "SWFS-COMP-003", "SWFS-COMP-004", "SWFS-COMP-005", "SWFS-COMP-006", "SWFS-COMP-007", "SWFS-COMP-008", "SWFS-COMP-009", "SWFS-COMP-010", "SWFS-COMP-011", "SWFS-COMP-012", "SWFS-COMP-013", "SWFS-COMP-014"];
const checkIDs = ["occurrence_inventory", "release_surface", "dependency_boundary", "inventory_classification", "threat_policy", "storage_ref_policy", "backup_contracts", "compatibility", "recovery", "redaction"];
const report = () => ({ schema_id: "cartulary.seaweedfs_s3_compatibility_report.v1", probe_id: "fixture", object_store_backend: "seaweedfs_s3",
  started_at: "2026-11-11T00:00:00Z", completed_at: "2026-11-11T00:00:01Z", result: "pass", forbidden_skip_rows: [],
  cases: caseIDs.map(case_id => ({ case_id, capability: "fixture", status: "pass", reason_code: null, evidence: { source: "fixture" } })) });
const unit = () => ({ schema_id: "cartulary.harness_unit_result.v1", unit_id: "operation:harness.seaweedfs_compatibility", semantic_digest: `sha256:${"b".repeat(64)}`,
  status: "passed", exit_code: 0, signal: null, failure_class: null, failure_reason: null, evidence_outputs: ["seaweedfs-compatibility/object-store-compatibility-report.json"], missing_outputs: [] });
const manifest = () => buildReaderEvidence({ runID: "seaweed-fixture", target, commandID: "cartulary.harness.command.seaweedfs_release_gate.v2" }).manifest;
const outcomes = () => Object.fromEntries(checkIDs.map(id => [id, { status: "pass", evidence_refs: [`evidence/${id}.json`] }]));
const recoveryRowID = "module.recovery.process.canonical_operator_process_evidence_maps_the_imp_9808fdd4e9";
const hash = value => `sha256:${createHash("sha256").update(canonicalJSONString(value)).digest("hex")}`;

// Independent two-unit graph; no compiler or evidence reader creates the oracle.
function recoveryFixture() {
  const row = { row_id: recoveryRowID, runner: "go" };
  const gate = buildEvidenceUnit("operation:harness.seaweedfs_release_gate");
  gate.needs = ["recovery"];
  gate.current_run_evidence_outputs = ["unit-results/operation-harness.seaweedfs_release_gate.json"];
  const units = [gate, buildEvidenceUnit("recovery", [row])];
  for (const value of units) { value.shared_locks = ["host_activity"]; value.semantic_digest = hash(value); }
  const graph = { schema_id: "cartulary.harness_work_graph.v8", selector: `units:${hash(units.map(value => value.unit_id))}`,
    units, graph_digest: hash({ units }) };
  const m = { ...manifest(), graph_digest: graph.graph_digest };
  const context = { schema_id: "cartulary.harness_execution_context.v3", manifest_digest: hash(m), graph,
    projections: {}, source_inputs: [], host_profile: { platform: "linux", architecture: "x64", kernel_release: "fixture", cpu_models: ["fixture"] }, instrumentation_policy_bytes: "{}" };
  return { manifest: m, context, unit: buildUnitResult(units[1]), row: buildRowResult(row) };
}

function persistRecovery(root, fixture) {
  writeEvidenceFile(root, "execution-context.json", fixture.context);
  writeEvidenceFile(root, "unit-results/recovery.json", fixture.unit);
  writeEvidenceFile(root, `rows/${recoveryRowID}.json`, fixture.row);
}

test("SeaweedFS recovery requires current bound graph, prerequisite execution and actual passing row", t => {
  const root = createScratch(t, "cartulary-seaweed-recovery-");
  const fixture = recoveryFixture(); persistRecovery(root, fixture);
  const context = { runRoot: root, manifest: fixture.manifest, target };
  assert.deepEqual(readCurrentSeaweedfsRecovery(context), { row: fixture.row, unit: fixture.unit,
    evidence_refs: ["unit-results/recovery.json", `rows/${recoveryRowID}.json`] });
  // No terminal run-summary exists while the dependent gate is still executing.
  assert.equal(fs.existsSync(path.join(root, "run-summary.json")), false);
  for (const relative of ["execution-context.json", "unit-results/recovery.json", `rows/${recoveryRowID}.json`]) {
    const file = path.join(root, relative), bytes = fs.readFileSync(file);
    writeEvidenceFile(root, `old-run/${relative}`, bytes.toString("utf8"));
    fs.unlinkSync(file);
    assert.throws(() => readCurrentSeaweedfsRecovery(context), /^Error: SeaweedFS recovery admission failed$/u);
    fs.writeFileSync(file, bytes, { mode: 0o600 });
  }
  for (const mutate of [
    f => { f.manifest.run_id = "prior-run"; },
    f => { f.manifest.graph_digest = `sha256:${"f".repeat(64)}`; },
    f => { f.context.graph.units[0].needs = []; },
    f => { f.context.graph.units[1].semantic_digest = `sha256:${"f".repeat(64)}`; },
    f => { f.unit.semantic_digest = `sha256:${"f".repeat(64)}`; },
    f => { f.unit.unit_id = "unrelated"; }, f => { f.unit.exit_code = 1; },
    f => { f.unit.status = "failed"; }, f => { f.unit.signal = "SIGTERM"; },
    f => { f.unit.failure_class = "artifact"; }, f => { f.unit.failure_reason = "artifact_error"; },
    f => { f.unit.missing_outputs = ["required.json"]; }, f => { f.unit.evidence_outputs = []; },
    f => { f.row.row_id = "module.recovery.process.unrelated"; }, f => { f.row.runner = "shell"; },
    f => { f.row = buildRowResult({ row_id: recoveryRowID, runner: "go" }, "failed"); },
    f => { f.row = buildRowResult({ row_id: recoveryRowID, runner: "go" }, "skipped_authorized"); },
    f => { f.row.started_at = "2026-10-08T00:00:00.000Z"; },
    f => { f.row.finished_at = "2026-10-15T00:00:00.000Z"; },
    f => { f.row.started_at = "2026-02-31T00:00:00.000Z"; },
    f => { f.row.extra = "untrusted"; },
  ]) {
    const changed = structuredClone(fixture); mutate(changed); persistRecovery(root, changed);
    assert.throws(() => readCurrentSeaweedfsRecovery({ ...context, manifest: changed.manifest }), /^Error: SeaweedFS recovery admission failed$/u);
  }
  // Internally consistent graph identities still cannot omit the required edge,
  // select a different row or authorize cached process evidence.
  for (const mutate of [
    units => { units[0].needs = []; },
    units => { units[0].unit_id = "operation:harness.unrelated"; },
    units => { units[1].row_bindings[0].row_id = "module.recovery.process.unrelated";
      units[1].current_run_evidence_outputs = ["rows/module.recovery.process.unrelated.json", "unit-results/recovery.json"]; },
    units => { units[1].cache_policy = "content_addressed"; },
  ]) {
    const changed = structuredClone(fixture); mutate(changed.context.graph.units);
    for (const unit of changed.context.graph.units) { delete unit.semantic_digest; unit.semantic_digest = hash(unit); }
    const graph = changed.context.graph;
    graph.selector = `units:${hash(graph.units.map(value => value.unit_id))}`;
    graph.graph_digest = hash({ units: graph.units }); changed.manifest.graph_digest = graph.graph_digest;
    changed.context.manifest_digest = hash(changed.manifest);
    changed.unit.semantic_digest = graph.units[1].semantic_digest;
    persistRecovery(root, changed);
    assert.throws(() => readCurrentSeaweedfsRecovery({ ...context, manifest: changed.manifest }), /^Error: SeaweedFS recovery admission failed$/u);
  }
});

test("SeaweedFS compatibility separates malformed shape from failed current observations", () => {
  validateSeaweedfsCompatibility(report(), unit());
  for (const change of [
    v => delete v.schema_id, v => { v.schema_id = "old"; }, v => { v.object_store_backend = "unknown"; },
    v => delete v.forbidden_skip_rows, v => { v.extra = "private"; }, v => { v.cases[0].extra = true; },
    v => { v.cases[0].evidence.secret = "private"; }, v => { v.cases[0].status = true; },
    v => { v.cases[0].case_id = "SWFS-COMP-999"; }, v => { v.cases[0].evidence = null; }, v => v.cases.pop(),
  ]) {
    const value = report(); change(value);
    assert.throws(() => validateSeaweedfsCompatibility(value, unit()), /^Error: SeaweedFS compatibility shape invalid$/u);
  }
  for (const change of [v => { v.cases[1].case_id = v.cases[0].case_id; }, v => { v.cases[0].status = "not_run"; },
    v => { v.cases[0].reason_code = "incomplete"; }, v => { v.result = "fail"; }, v => v.forbidden_skip_rows.push("required"),
    v => { v.completed_at = "2026-11-10T00:00:00Z"; }, v => { v.started_at = "2026-02-31T00:00:00Z"; },
    v => { v.started_at = "2026-11-11T00:00:00.000000002Z"; v.completed_at = "2026-11-11T00:00:00.000000001Z"; }]) {
    const value = report(); change(value);
    assert.throws(() => validateSeaweedfsCompatibility(value, unit()), /^Error: SeaweedFS compatibility observations did not pass$/u);
  }
  for (const change of [v => delete v.semantic_digest, v => { v.extra = true; }, v => { v.status = "pass"; }]) {
    const value = unit(); change(value);
    assert.throws(() => validateSeaweedfsCompatibility(report(), value), /shape invalid/u);
  }
  for (const change of [v => { v.unit_id = "operation:harness.other"; }, v => { v.status = "failed"; },
    v => { v.exit_code = 1; }, v => { v.signal = "SIGTERM"; }, v => { v.failure_class = "artifact"; },
    v => { v.failure_reason = "artifact_error"; }, v => v.missing_outputs.push("required.json"), v => { v.evidence_outputs = []; }]) {
    const value = unit(); change(value);
    assert.throws(() => validateSeaweedfsCompatibility(report(), value), /^Error: SeaweedFS compatibility execution did not pass$/u);
  }
});

test("SeaweedFS evidence reads are bounded strict private-path independent admissions", t => {
  const root = createScratch(t, "cartulary-seaweed-read-");
  const file = path.join(root, "input.json");
  fs.writeFileSync(file, JSON.stringify(report()), { mode: 0o600 });
  assert.deepEqual(readSeaweedfsJSON(root, "input.json", "cartulary.seaweedfs_s3_compatibility_report.v1"), report());
  for (const bytes of ['{"result":"fail","result":"pass"}', '{"unfinished":', Buffer.from([0xff]), '{}garbage']) {
    fs.writeFileSync(file, bytes);
    assert.throws(() => readSeaweedfsJSON(root, "input.json"), /^Error: SeaweedFS evidence admission failed$/u);
  }
  fs.truncateSync(file, 16 * 1024 * 1024 + 1);
  assert.throws(() => readSeaweedfsJSON(root, "input.json"), /admission failed/u);
  fs.writeFileSync(file, JSON.stringify(report()));
  fs.symlinkSync("input.json", path.join(root, "alias.json"));
  fs.symlinkSync(root, path.join(root, "alias"));
  for (const relative of ["alias.json", "alias/input.json", "../input.json", "/input.json", "absent.json"]) {
    assert.throws(() => readSeaweedfsJSON(root, relative), /^Error: SeaweedFS evidence admission failed$/u);
  }
  const original = fs.openSync;
  let humanProbes = 0;
  try {
    fs.openSync = () => { humanProbes++; throw new Error("forbidden probe"); };
    syncBuiltinESMExports();
    assert.throws(() => readSeaweedfsJSON(root, "README"), /^Error: SeaweedFS evidence admission failed$/u);
    assert.equal(humanProbes, 0, "human input must reject before any filesystem open");
  } finally { fs.openSync = original; syncBuiltinESMExports(); }
});

test("SeaweedFS compatibility requires both exact current artifacts without historical or services fallback", t => {
  const root = createScratch(t, "cartulary-seaweed-current-");
  const write = (relative, value) => {
    const file = path.join(root, relative); fs.mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
    fs.writeFileSync(file, JSON.stringify(value), { mode: 0o600 });
  };
  write("stable/object-store-compatibility-report.json", report());
  write("services-up/object-store-compatibility-report.json", report());
  write("old-run/seaweedfs-compatibility/object-store-compatibility-report.json", report());
  const context = { runRoot: root };
  assert.throws(() => readCurrentSeaweedfsCompatibility(context), /admission failed/u);
  write("seaweedfs-compatibility/object-store-compatibility-report.json", report());
  assert.throws(() => readCurrentSeaweedfsCompatibility(context), /admission failed/u);
  write("unit-results/operation-harness.seaweedfs_compatibility.json", unit());
  const current = readCurrentSeaweedfsCompatibility(context);
  assert.deepEqual(current.report, report()); assert.deepEqual(current.unit, unit());
  assert.deepEqual(current.evidence_refs, ["unit-results/operation-harness.seaweedfs_compatibility.json", "seaweedfs-compatibility/object-store-compatibility-report.json"]);
  const oldSkip = process.env.CARTULARY_HARNESS_SKIP_PREREQUISITES;
  try {
    process.env.CARTULARY_HARNESS_SKIP_PREREQUISITES = "1";
    assert.deepEqual(readCurrentSeaweedfsCompatibility(context), current, "completed canonical evidence determines qualification");
    fs.unlinkSync(path.join(root, current.evidence_refs[0]));
    assert.throws(() => readCurrentSeaweedfsCompatibility(context), /admission failed/u, "environment cannot invent missing evidence");
  } finally {
    if (oldSkip === undefined) delete process.env.CARTULARY_HARNESS_SKIP_PREREQUISITES;
    else process.env.CARTULARY_HARNESS_SKIP_PREREQUISITES = oldSkip;
  }
});

test("SeaweedFS summary keeps policy inventory distinct and blocks every missing or failed boundary", () => {
  const base = { manifest: manifest(), target, outcomes: outcomes() };
  const passed = buildSeaweedfsSummary(base);
  assert.equal(passed.status, "pass");
  assert.deepEqual(passed.checks.map(item => item.check_id), checkIDs);
  assert.deepEqual(passed.checks.map(item => item.kind), ["source_policy", "source_policy", "source_policy", "source_policy", "contract_inventory", "contract_inventory", "contract_inventory", "runtime", "runtime", "privacy"]);
  for (const id of checkIDs) {
    for (const status of ["fail", "unqualified"]) {
      const changed = outcomes(); changed[id].status = status;
      const result = buildSeaweedfsSummary({ ...base, outcomes: changed });
      assert.equal(result.status, "blocked"); assert.equal(result.checks.find(item => item.check_id === id).status, status);
    }
    const changed = outcomes(); delete changed[id];
    assert.equal(buildSeaweedfsSummary({ ...base, outcomes: changed }).status, "blocked");
  }
  assert.throws(() => buildSeaweedfsSummary({ ...base, target: "seaweedfs-release-evidence" }));
  const changed = outcomes(); changed.compatibility.evidence_refs = [];
  assert.throws(() => buildSeaweedfsSummary({ ...base, outcomes: changed }), /check closure invalid/u);
  const value = buildSeaweedfsSummary({ ...base, findings: [{ check_id: "redaction", code: "public-raw-storage-field", message: "private payload", endpoint: "private endpoint" }],
    occurrences: [{ path: "source/file.go", line: 2, column: 3, token: "fixture", classification: "invalid", line_excerpt: "private payload", owner: "private owner" }] });
  assert.equal(value.status, "blocked");
  assert.deepEqual(value.findings, [{ check_id: "redaction", code: "public-raw-storage-field", path: null, line: null, column: null, json_path: null }]);
  assert.deepEqual(value.source_occurrences, [{ path: "source/file.go", line: 2, column: 3, matched_term: "fixture", classification: "invalid", classification_rule_id: null }]);
  assert.doesNotMatch(JSON.stringify(value), /private/u);
  for (const mutate of [v => v.checks.reverse(), v => { v.checks[0] = v.checks[1]; }, v => { v.checks[4].kind = "runtime"; }, v => { v.status = "blocked"; }]) {
    const invalid = structuredClone(passed); mutate(invalid); assert.throws(() => validateSeaweedfsSummary(invalid));
  }
});

test("SeaweedFS privacy distinguishes public fields from private ownership without exposing values", () => {
  const selected = value => [{ path: "current/artifact.json", boundary: "public_or_shareable", value }];
  const safe = { endpoint: { redacted: true, redaction_class: "endpoint", sha256: "a".repeat(64) }, session_target: "retained" };
  assert.deepEqual(scanSeaweedfsPrivacy(selected(safe)), { status: "pass", evidence_refs: ["current/artifact.json"], findings: [] });
  const raw = { nested: [{ source_bucket: "private-bucket", endpoint: "https://source.internal" }] };
  const publicResult = scanSeaweedfsPrivacy(selected(raw));
  assert.equal(publicResult.status, "fail");
  assert.deepEqual(publicResult.findings.map(item => [item.code, item.json_path]), [["public-storage-identifier", null], ["public-raw-storage-field", "source_bucket"], ["public-raw-storage-field", "endpoint"]]);
  assert.doesNotMatch(JSON.stringify(publicResult), /private-bucket|source\.internal/u);
  assert.doesNotMatch(JSON.stringify(scanSeaweedfsPrivacy(selected({ "sensitive-fixture-value": raw }))), /sensitive-fixture-value/u);
  const privateInput = selected(raw); privateInput[0].boundary = "operator_private";
  assert.equal(scanSeaweedfsPrivacy(privateInput).status, "pass");
  const misleading = selected(raw); misleading[0].path = "object-store-backup-manifest.json";
  assert.equal(scanSeaweedfsPrivacy(misleading).status, "fail", "a filename cannot confer private-artifact authority");
  for (const boundary of ["operator_private", "public_or_shareable"]) for (const value of [
    { password: "sensitive-fixture-value" }, { authorization: "Bearer sensitive-fixture-value" },
    { diagnostic: "CARTULARY_S3_SECRET_KEY=sensitive-fixture-value" }, { diagnostic: "seaweedfs-secret" },
  ]) {
    const result = scanSeaweedfsPrivacy([{ path: "current/input.json", boundary, value }]);
    assert.equal(result.status, "fail"); assert.ok(result.findings.some(item => item.code === "raw-secret-value"));
    assert.doesNotMatch(JSON.stringify(result), /sensitive-fixture-value|seaweedfs-secret/u);
  }
  assert.equal(scanSeaweedfsPrivacy(selected(undefined)).findings[0].code, "current-artifact-missing");
  assert.throws(() => scanSeaweedfsPrivacy([]), /selection invalid/u);
  assert.throws(() => scanSeaweedfsPrivacy([...selected(safe), ...selected(safe)]), /selection invalid/u);
});

test("SeaweedFS inventories require the exact canonical typed pair and one semantic identity", t => {
  const root = createScratch(t, "cartulary-seaweed-inventory-");
  const directory = path.join(root, ".cartulary/release-artifacts");
  fs.mkdirSync(directory, { recursive: true, mode: 0o700 });
  const bom = { bomFormat: "CycloneDX", specVersion: "1.7", version: 1,
    serialNumber: "urn:uuid:aaaaaaaa-aaaa-5aaa-8aaa-aaaaaaaaaaaa", metadata: { component: {
      type: "application", name: "fixture", properties: [{ name: "cartulary:semantic_input_digest", value: `sha256:${"a".repeat(64)}` }],
    } } };
  const license = { schema_id: "cartulary.license_report.v2", semantic_input_digest: `sha256:${"a".repeat(64)}`, entries: [] };
  const write = (name, value) => fs.writeFileSync(path.join(directory, name), JSON.stringify(value), { mode: 0o644 });
  assert.throws(() => readSeaweedfsInventories(root), /inventory admission failed/u);
  write("sbom.cyclonedx.json", bom);
  assert.throws(() => readSeaweedfsInventories(root), /inventory admission failed/u);
  write("license-report.json", license);
  assert.deepEqual(readSeaweedfsInventories(root), { sbom: bom, licenseReport: license, evidence_refs: [".cartulary/release-artifacts/sbom.cyclonedx.json", ".cartulary/release-artifacts/license-report.json"] });
  for (const changed of [{ ...license, semantic_input_digest: `sha256:${"b".repeat(64)}` }, { ...license, schema_id: "cartulary.license_report.v1" }, { ...license, extra: true }]) {
    write("license-report.json", changed); assert.throws(() => readSeaweedfsInventories(root), /^Error: SeaweedFS inventory admission failed$/u);
  }
  write("license-report.json", license);
  for (const change of [v => { v.specVersion = "1.6"; }, v => { v.metadata.timestamp = "2026-11-11T00:00:00Z"; },
    v => { v.serialNumber = "urn:uuid:bbbbbbbb-bbbb-5bbb-8bbb-bbbbbbbbbbbb"; }, v => v.metadata.component.properties.push(v.metadata.component.properties[0])]) {
    const changed = structuredClone(bom); change(changed); write("sbom.cyclonedx.json", changed);
    assert.throws(() => readSeaweedfsInventories(root), /^Error: SeaweedFS inventory admission failed$/u);
  }
});

test("SeaweedFS publication binds explicit current context and keeps one immutable private report", t => {
  const parent = createScratch(t, "cartulary-seaweed-publish-");
  const m = manifest(), runRoot = path.join(parent, m.run_id);
  fs.mkdirSync(runRoot, { mode: 0o700 }); fs.writeFileSync(path.join(runRoot, "run-manifest.json"), JSON.stringify(m), { mode: 0o600 });
  const env = { CARTULARY_TEST_RESULTS_DIR: parent, CARTULARY_TEST_RUN_ID: m.run_id, CARTULARY_TEST_TARGET: target };
  const context = seaweedfsRunContext(env);
  assert.deepEqual(context, { runRoot, manifest: m, target });
  for (const change of [v => delete v.CARTULARY_TEST_RUN_ID, v => { v.CARTULARY_TEST_RUN_ID = "../escape"; },
    v => { v.CARTULARY_TEST_TARGET = "services-up"; }, v => { v.CARTULARY_TEST_TARGET = "seaweedfs-release-evidence"; }, v => { v.CARTULARY_TEST_RESULTS_DIR = "relative"; }]) {
    const value = { ...env }; change(value); assert.throws(() => seaweedfsRunContext(value), /explicit current run context/u);
  }
  const summary = buildSeaweedfsSummary({ manifest: m, target, outcomes: outcomes() });
  const relative = publishSeaweedfsSummary(context, summary);
  assert.equal(relative, "seaweedfs-release-gate/artifacts/release-gate-summary.json");
  assert.deepEqual(readSeaweedfsJSON(runRoot, relative, summary.schema_id), summary);
  const bytes = fs.readFileSync(path.join(runRoot, relative));
  assert.equal(fs.statSync(path.join(runRoot, target, "artifacts")).mode & 0o777, 0o700);
  assert.equal(fs.statSync(path.join(runRoot, relative)).mode & 0o777, 0o600);
  assert.deepEqual(fs.readdirSync(path.join(runRoot, target, "artifacts")), ["release-gate-summary.json"]);
  assert.throws(() => publishSeaweedfsSummary(context, summary));
  assert.deepEqual(fs.readFileSync(path.join(runRoot, relative)), bytes);
  assert.throws(() => publishSeaweedfsSummary(context, { ...summary, source_digest: `sha256:${"f".repeat(64)}` }), /identity mismatch/u);
  const secret = buildSeaweedfsSummary({ manifest: m, target, outcomes: outcomes(), occurrences: [
    { path: "source/seaweedfs-secret.go", line: 1, column: 1, token: "fixture", classification: "sdk_only" },
  ] });
  assert.throws(() => publishSeaweedfsSummary(context, secret), /contains private data/u);
  assert.deepEqual(fs.readFileSync(path.join(runRoot, relative)), bytes);
  const bad = path.join(parent, "linked"); fs.mkdirSync(bad, { mode: 0o700 }); fs.symlinkSync(path.join(runRoot, target), path.join(bad, target));
  assert.throws(() => publishSeaweedfsSummary({ ...context, runRoot: bad }, summary));
  assert.deepEqual(fs.readFileSync(path.join(runRoot, relative)), bytes);
});

test("SeaweedFS graph requires the actual current reports and has no legacy path overrides", () => {
  const compiler = new WorkGraphCompiler(path.resolve(import.meta.dirname, "../../.."));
  const graph = compiler.compileTarget("seaweedfs-release-gate");
  const compatibility = graph.units.find(value => value.unit_id === "operation:harness.seaweedfs_compatibility");
  assert.ok(compatibility.current_run_evidence_outputs.includes("seaweedfs-compatibility/object-store-compatibility-report.json"));
  const gate = graph.units.find(value => value.unit_id === "operation:harness.seaweedfs_release_gate");
  assert.ok(gate.current_run_evidence_outputs.includes("seaweedfs-release-gate/artifacts/release-gate-summary.json"));
  const execution = compiler.owner.operations["harness.seaweedfs_release_gate"].execution;
  assert.deepEqual(execution.command, ["$(NODE_BIN)", "./tools/release-evidence/seaweedfs-release-evidence.mjs"]);
  assert.deepEqual(execution.env, {});
  assert.deepEqual(compiler.owner.policy_units["harness.seaweedfs_release_gate"].required_rows.map(value => value.row_id), [recoveryRowID]);
  const recovery = graph.units.find(value => value.row_bindings.some(row => row.row_id === recoveryRowID));
  assert.ok(recovery, "canonical operator process evidence must execute in this invocation");
  assert.equal(recovery.cache_policy, "none");
  assert.deepEqual(recovery.service_dependencies, ["object_store", "postgres"]);
  assert.ok(gate.needs.includes(recovery.unit_id));
  assert.ok(recovery.current_run_evidence_outputs.includes(`rows/${recoveryRowID}.json`));
  assert.deepEqual(graph.units.flatMap(value => value.row_bindings.map(row => row.row_id)), [recoveryRowID]);
  assert.equal(Object.hasOwn(compiler.owner.operations, "harness.seaweedfs_release_evidence"), false);
  assert.equal(Object.hasOwn(compiler.owner.policy_units, "harness.seaweedfs_release_evidence"), false);
  const row = compiler.catalog.rows.find(value => value.row_id === recoveryRowID);
  assert.deepEqual(row.selector.tests, ["TestCanonicalOperatorBackupCreate_Process", "TestCanonicalOperatorBackupInspectLatest_Process",
    "TestCanonicalOperatorRestoreLatest_Process", "TestCanonicalOperatorRestoreVerifyDue_Process", "TestCanonicalOperatorRestoreVerifyLatest_Process"]);
  assert.equal(row.runtime_profile_id, "default");
  assert.deepEqual(row.runtime_binary_ids, ["operator"]);
});
