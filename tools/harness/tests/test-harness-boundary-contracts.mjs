import { runContractSuite } from "./contract-suite-support.mjs";
runContractSuite("boundaries");

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { validateSchemaSync } from "../contract/index.mjs";
import { validateInspectorFieldLayouts } from "../generated-artifacts/design-presentation/design-presentation.mjs";

test("inspector presentation v2 rejects invalid authored field layouts", () => {
  const root = path.resolve(import.meta.dirname, "../../..");
  const presentation = JSON.parse(readFileSync(path.join(root, "contracts/design/presentation.v2.json"), "utf8"));
  const layouts = presentation.inspector.field_layout_overrides;
  const viewSchemas = path.join(root, "contracts/view-schemas");
  validateSchemaSync("cartulary.design_presentation.v2", presentation);
  validateInspectorFieldLayouts(layouts, viewSchemas);
  for (const candidate of [
    [...layouts, { ...layouts[0], layout: "narrative" }],
    [{ ...layouts[0], view_schema_id: "unknown" }],
    [{ ...layouts[0], field_key: "timeline.nonexistent" }],
    [{ ...layouts[0], layout: "unknown" }],
  ]) assert.throws(() => validateInspectorFieldLayouts(candidate, viewSchemas));
  const unknownLayout = structuredClone(presentation);
  unknownLayout.inspector.field_layout_overrides[0].layout = "unknown";
  assert.throws(() => validateSchemaSync("cartulary.design_presentation.v2", unknownLayout));
});

import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import { collectHarnessImportBoundaryViolations } from "../static-analysis/harness-import-boundary.mjs";

test("current harness imports satisfy semantic owner boundaries", () => {
  const report = collectHarnessImportBoundaryViolations(path.resolve(import.meta.dirname, "../../.."));
  assert.deepEqual(report.violations, [], JSON.stringify(report.violations, null, 2));
});

test("harness boundaries reject private imports, unrelated exceptions, unknown owners and invalid facades", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-harness-boundaries-"));
  const write = (name, body) => {
    const file = path.join(root, name);
    mkdirSync(path.dirname(file), { recursive: true });
    writeFileSync(file, body);
  };
  const registry = { schema_id: "cartulary.harness_helper_ownership.v1", facades: [
    { key: "browser_owner_facade", boundary_group: "browser", paths: ["tools/harness/browser/index.mjs"], allowed_consumers: ["tools/harness/scheduler/adapters/browser.mjs"] },
    { key: "test_catalog", boundary_group: "test_catalog", paths: ["tools/harness/test-catalog/index.mjs"], allowed_consumers: ["tools/harness/readiness/check.mjs"] },
  ] };
  const save = () => write("tools/harness_helper_ownership.json", JSON.stringify(registry));
  try {
    for (const file of ["browser/index.mjs", "browser/private.mjs", "test-catalog/index.mjs", "test-catalog/private.mjs", "backend/private.mjs", "evidence-accounting/private.mjs"]) write(`tools/harness/${file}`, "export const value = 1;\n");
    write("tools/harness/scheduler/adapters/browser.mjs", 'export { value } from "../../browser/private.mjs";\n');
    write("tools/harness/readiness/check.mjs", 'import { value } from "../browser/index.mjs";\nconst text = \'import "../browser/private.mjs"\';\n');
    save();
    assert.deepEqual(collectHarnessImportBoundaryViolations(root).violations, []);
    write("tools/harness/readiness/check.mjs", 'import "../browser/private.mjs";\nimport "../backend/private.mjs";\nimport "../evidence-accounting/private.mjs";\nimport "../core/private.mjs";\nimport "../frontend/private.mjs";\n');
    for (const owner of ["core", "frontend"]) write(`tools/harness/${owner}/private.mjs`, "export {};\n");
    write("tools/harness/scheduler/unauthorized.mjs", 'import "../browser/index.mjs";\nimport "../test-catalog/private.mjs";\n');
    write("tools/harness/unrecognized/child.mjs", "export {};\n");
    const rules = new Set(collectHarnessImportBoundaryViolations(root).violations.map((entry) => entry.rule));
    for (const rule of ["forbidden_private_browser_import", "forbidden_private_backend_import", "forbidden_private_test_catalog_import", "forbidden_scheduler_private_browser_import", "forbidden_unknown_harness_owner_root", "forbidden_private_evidence_accounting_import", "forbidden_private_catch_all_import"]) assert.ok(rules.has(rule), rule);
    registry.facades[0].paths.push("tools/harness/browser/missing.mjs"); save();
    assert.throws(() => collectHarnessImportBoundaryViolations(root), /missing facade/);
    registry.facades[0].paths.pop();
    registry.facades.push({ ...registry.facades[0], key: "duplicate" }); save();
    assert.throws(() => collectHarnessImportBoundaryViolations(root), /more than one facade/);
  } finally { rmSync(root, { recursive: true, force: true }); }
});


import { spawn, spawnSync } from "node:child_process";
import { browserAcquisitionLaunchArguments } from "../test-support/owner-fixtures.mjs";

test("browser acquisition launches its owning dispatcher and rejects invalid ownership", () => {
  const command = browserAcquisitionLaunchArguments("missing-acquisition", "invalid-launch", process.execPath, ["--version"]);
  const childEnv = { ...process.env };
  delete childEnv.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
  const child = spawnSync(process.execPath, command, { encoding: "utf8", env: childEnv });
  assert.notEqual(child.status, 0, "an import-only facade must never count as a successful browser producer");
  assert.match(child.stderr, /browser acquisition launch failed/);
});


test("foundation validators preserve bootstrap operation without installed dependencies", async () => {
  await import("../contract/tests/test-foundation-schema-validator.mjs");
});


test("bootstrap artifact cache preserves content identity and failure handling", () => {
  const result = spawnSync("bash", ["tools/harness/readiness/tests/test-cache-artifact.sh"], {
    cwd: path.resolve(import.meta.dirname, "../../.."),
    env: { ...process.env, NODE_BIN: process.execPath }, encoding: "utf8",
  });
  assert.equal(result.status, 0, result.stdout + result.stderr);
});


test("retired private process dispatch rejects without signaling a live owned process", async (t) => {
  const { once } = await import("node:events");
  const { createRuntime } = await import("../runtime/session-client.mjs");
  const { ownedProcess, stopOwnedProcess } = await import("../runtime/owned-process.mjs");
  const repository = path.resolve(import.meta.dirname,"../../..");
  const scratch = mkdtempSync(path.join(os.tmpdir(),"retired-process-dispatch-"));
  const runRoot = path.join(scratch,"results","retired-dispatch");mkdirSync(runRoot,{recursive:true,mode:0o700});
  const runtime = createRuntime({repoRoot:repository,runRoot,runID:"retired-dispatch",scratchRoot:path.join(scratch,"private")});
  const child = spawn(process.execPath,["-e","process.stdout.write('ready');setInterval(()=>{},1000)"],{detached:true,stdio:["ignore","pipe","ignore"]});
  const closed=once(child,"close");
  t.after(async()=>{if(child.exitCode===null){child.kill("SIGKILL");await closed;}});
  await once(child.stdout,"data");
  const proof=ownedProcess(child.pid);
  const resource={kind:"helper_process",target:proof,state:"acquired",run_id:runtime.runID,lease_id:runtime.leaseID,depends_on:[],recovery_files:[]};
  validateSchemaSync("cartulary.harness_runtime_resource.v2",resource);
  runtime.record(resource);
  try {
    assert.throws(()=>validateSchemaSync("cartulary.harness_runtime_resource.v2",{...resource,kind:"preparation_process"}));
    const result=spawnSync(process.execPath,[path.join(repository,"tools/harness/runtime/resource-stop-private.mjs")],{
      cwd:repository,env:{...process.env,...runtime.producerEnvironment(),CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT:runRoot},input:JSON.stringify(resource),timeout:10000,encoding:"utf8"});
    assert.equal(result.error,undefined);assert.equal(result.status,12);
    assert.equal(child.exitCode,null);assert.doesNotThrow(()=>process.kill(child.pid,0),"retired dispatch signaled a process");
  } finally {
    await stopOwnedProcess(proof);await closed;runtime.record({...resource,state:"released"});runtime.close();rmSync(scratch,{recursive:true,force:true});
  }
});


test("frontend owner boundaries survive aliasing extraction and comments", () => {
  const repository = path.resolve(import.meta.dirname, "../../..");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "frontend-owner-boundaries-"));
  const owner = JSON.parse(readFileSync(path.join(repository, "tools/frontend_import_boundaries.json"), "utf8"));
  const identities = ["frontend-grid-vendor-boundary", "frontend-timeline-no-assessment-owner", "frontend-assessment-owner-no-timeline-imports", "frontend-workbook-policy-transport-boundary"];
  const rules = owner.rules.filter((rule) => identities.includes(rule.id));
  assert.equal(rules.length, identities.length);
  const cases = [
    ["workbook/timeline/models/renamed.ts", 'import type { AssessmentCreateDraft as Local } from "../../models/assessmentWorkbookModel"; export type Extracted = Local;', "frontend-timeline-no-assessment-owner"],
    ["workbook/components/AssessmentWorkbookSurface.tsx", 'export { row as Renamed } from "../timeline/models/row";', "frontend-assessment-owner-no-timeline-imports"],
    ["networkFlow/renamed.ts", 'export const load = () => import("react-data-grid");', "frontend-grid-vendor-boundary"],
    ["workbook/policies/renamed.ts", 'import * as local from "../../services/browserApi"; export const capability = local;', "frontend-workbook-policy-transport-boundary"],
    ["workbook/timeline/models/renamed.ts", '// AssessmentCreateDraft react-data-grid authorization\nimport type { Row as Local } from "./row"; export type Extracted = Local;', null],
  ];
  try {
    for (const [index, [file, source, expected]] of cases.entries()) {
      const root = path.join(scratch, String(index));
      const target = path.join(root, "apps/web/src", file);
      mkdirSync(path.dirname(target), { recursive: true });
      writeFileSync(target, source);
      writeFileSync(path.join(root, "boundaries.json"), JSON.stringify({
        schema_id: owner.schema_id, scan_roots: ["apps/web/src"], scan_excludes: [],
        acyclic_import_graphs: [], singleton_imports: [], rules, raw_design_token_literal_checks: [],
      }));
      const childEnv = { ...process.env };
      delete childEnv.CARTULARY_TEST_TARGET;
      const child = spawnSync(process.execPath, [path.join(repository, "tools/harness/static-analysis/frontend-import-boundary-check-cli.mjs"), "--root", root, "--config", "boundaries.json"], { env: childEnv, encoding: "utf8", timeout: 10000 });
      assert.equal(child.error, undefined);
      assert.equal(child.status, expected === null ? 0 : 1, child.stdout + child.stderr);
      if (expected !== null) assert.ok(child.stderr.includes(expected), child.stderr);
    }
  } finally { rmSync(scratch, { recursive: true, force: true }); }
});


test("physical cleanup diagnostics remain bounded and cannot authorize success", async (t) => {
  const { stopOwnedProcess } = await import("../runtime/owned-process.mjs");
  const { findCleanupDiagnostic, withCleanupDiagnostic } = await import("../runtime/cleanup-diagnostic.mjs");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "cleanup-diagnostic-"));
  t.after(() => rmSync(scratch, {recursive:true,force:true}));
  const valid = {schema_id:"cartulary.harness_cleanup_diagnostic.v3",stage:"process_signal",cause:"io",errno:13};
  const cases = [
    [JSON.stringify(valid), "process_signal"],
    [JSON.stringify({...valid,process_id:123}), "process_protocol"],
    [JSON.stringify({...valid,cause:"credential-sentinel"}), "process_protocol"],
    [JSON.stringify(valid)+JSON.stringify(valid), "process_protocol"],
    [" ".repeat(4097), "process_protocol"],
  ];
  for (const [index,[stderr,stage]] of cases.entries()) {
    const binary = path.join(scratch,`physical-${index}`);
    writeFileSync(binary, `#!${process.execPath}\nprocess.stdin.resume();process.stdin.on("end",()=>{process.stderr.write(${JSON.stringify(stderr)});process.exitCode=12;});\n`, {mode:0o700});
    await assert.rejects(stopOwnedProcess({}, {environment:{...process.env,CARTULARY_HARNESS_RUNTIME_BIN:binary}}), error => {
      const diagnostic=findCleanupDiagnostic(new AggregateError([error],"private-path-sentinel"));
      assert.equal(diagnostic.stage,stage);
      assert.equal(JSON.stringify(diagnostic).includes("sentinel"),false);
      return true;
    });
  }
  const privateFailure = new Error("private-path-sentinel");
  Object.defineProperty(privateFailure,"privateDiagnostic",{value:{reason:"unsafe",errno:null,process_id:123,proof_file:"environ"}});
  const projected=findCleanupDiagnostic(withCleanupDiagnostic(privateFailure,"browser_acknowledgement"));
  assert.deepEqual(projected,{schema_id:valid.schema_id,stage:"browser_acknowledgement",cause:"unsafe",errno:0});
  const wrapped = new AggregateError([new Error("private wrapper", {cause:privateFailure})], "private outer");
  assert.deepEqual(findCleanupDiagnostic(withCleanupDiagnostic(wrapped,"recovery_proof")), projected);
  const unprojected = new Error("private-path-sentinel");
  Object.defineProperty(unprojected,"privateDiagnostic",{value:privateFailure.privateDiagnostic});
  const nested = new AggregateError([new Error("private wrapper", {cause:unprojected})], "private outer");
  assert.deepEqual(findCleanupDiagnostic(withCleanupDiagnostic(nested,"recovery_proof")),
    {schema_id:valid.schema_id,stage:"recovery_proof",cause:"unsafe",errno:0});
});


test("cleanup evidence retains each failure without leaking private errors", async (t) => {
  const { CleanupResults } = await import("../runtime/cleanup-results.mjs");
  const { withCleanupDiagnostic } = await import("../runtime/cleanup-diagnostic.mjs");
  const root = mkdtempSync(path.join(os.tmpdir(), "cleanup-results-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  const results = new CleanupResults();
  const cause = new Error("credential-sentinel");
  Object.defineProperty(cause, "privateDiagnostic", { value: { reason: "io", errno: 13, proof_file: "credential-sentinel" } });
  results.record("fixture_release", { error: withCleanupDiagnostic(cause, "browser_namespace") });
  results.record("host_release", { error: new Error("another-private-path") });
  results.record("services_close", { blocked: true });
  results.record("runtime_close");
  results.publish(root, "cleanup-evidence");
  const bytes = readFileSync(path.join(root, "cleanup-results.json"), "utf8");
  assert.equal(bytes.includes("sentinel"), false);
  assert.equal(bytes.includes("private-path"), false);
  const record = JSON.parse(bytes);
  validateSchemaSync("cartulary.harness_cleanup_results.v2", record);
  assert.deepEqual(record.results.map(x => x.diagnostic?.stage ?? null), ["browser_namespace", "host_release", null, null]);
  assert.equal(record.results[0].diagnostic.errno, 13);
  for (const mutate of [
    r => { r.schema_id = "cartulary.harness_cleanup_results.v1"; },
    r => { r.results[0].diagnostic = null; },
    r => { r.results[0].diagnostic.schema_id = "cartulary.harness_cleanup_diagnostic.v2"; },
    r => { r.results[0].diagnostic.secret = "credential-sentinel"; },
    r => { r.results[0].diagnostic.stage = "arbitrary-private-path"; },
    r => { r.results[3].diagnostic = r.results[0].diagnostic; },
  ]) {
    const candidate = structuredClone(record); mutate(candidate);
    assert.throws(() => validateSchemaSync("cartulary.harness_cleanup_results.v2", candidate));
  }
});
