import "../contract/tests/test-cleanup.mjs";
import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";
import { classifyShellFailureDetails } from "../output/test-output/runners/shell.mjs";
import { runContractSuite } from "./contract-suite-support.mjs";
import { registerRunStatusCommandTests } from "../diagnostics/tests/run-observation-cases.mjs";
import { registerFrontendProducerGraphTests } from "./test-frontend-producer-graph.mjs";
import { registerFrontendProducerLifecycleTests } from "./test-frontend-producer-lifecycle.mjs";
import { registerCommandFailureTests } from "./test-command-failure.mjs";
import { helpProjection, renderTaskSurfaceMake, renderTaskSurfaceMakeRuntime } from "../generated-artifacts/task-surface/make-renderer.mjs";
import { projectOperationBindings, readOperationOwner } from "../scheduler/work-graph/operations.mjs";
import { collectTaskSurfaceManifestErrors } from "../generated-artifacts/task-surface/index.mjs";
registerFrontendProducerGraphTests();
registerFrontendProducerLifecycleTests();
registerCommandFailureTests();
runContractSuite("command_surface");
registerRunStatusCommandTests();

test("native recovery exit retains cleanup classification independently of provider text", () => {
  assert.deepEqual(classifyShellFailureDetails({target:"test-run-recover",exitStatus:12}, [], ["untrusted provider text"], ""),
    {failure_class:"harness",failure_reason:"cleanup_error"});
});

test("real-command assertions forward the published parent diagnostics and reject invalid evidence before Make", (t) => {
  const root = path.resolve(import.meta.dirname, "../../..");
  const parent = mkdtempSync(path.join(tmpdir(), "real-command-diagnostics-"));
  t.after(() => rmSync(parent, { recursive: true, force: true }));
  const runID = "_parent-run";
  const runRoot = path.join(parent, runID);
  const artifacts = path.join(runRoot, "unit-artifacts", "assertion");
  mkdirSync(artifacts, { recursive: true });
  const manifestFile = path.join(runRoot, "run-manifest.json");
  const digest = `sha256:${"a".repeat(64)}`;
  const manifest = {
    schema_id: "cartulary.harness_run_manifest.v3", run_id: runID, target: "harness-contract",
    command_id: "cartulary.harness.command.harness_contract.v1", declared_inputs: {},
    source_commit: "a".repeat(40), source_state: "dirty", source_digest: digest,
    toolchain_digest: digest, system_digest: digest, graph_digest: digest,
    instrumentation: { mode: "basic", policy_digest: digest },
    cache_mode: "normal", started_at: "2026-10-20T00:00:00.000Z", go_worker_budget: 2,
    capability_snapshot: {
      schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 2, memory_bytes: 1024,
      process_slots: 2, io_tokens: 1, postgres_lanes: 1, object_store_lanes: 1,
      writable_volume: true, port_lanes: 1, services: {}, sources: {}, snapshot_digest: digest,
    },
  };
  const receipt = path.join(parent, "make-args");
  const make = path.join(parent, "fake-make");
  // Stop at the child boundary; these are transport tests, not nested run tests.
  writeFileSync(make, '#!/bin/sh\nprintf "%s\\n" "$@" > "$DIAGNOSTICS_TEST_RECEIPT"\nexit 73\n', { mode: 0o700 });
  const invoke = (caseName, ambient) => {
    rmSync(receipt, { force: true });
    const env = { ...process.env, MAKE_BIN: make, NODE_BIN: process.execPath,
      CARTULARY_TEST_RESULTS_DIR: parent, CARTULARY_TEST_RUN_ID: runID,
      CARTULARY_STEP_ARTIFACT_DIR: artifacts, DIAGNOSTICS_TEST_RECEIPT: receipt };
    delete env.HARNESS_DIAGNOSTICS;
    if (ambient !== undefined) env.HARNESS_DIAGNOSTICS = ambient;
    return spawnSync("bash", ["tools/harness/tests/test-tool-output-real-targets.sh", caseName],
      { cwd: root, env, encoding: "utf8", timeout: 10000, maxBuffer: 16384 });
  };
  for (const caseName of ["backend-unit", "backend-unit-machine"]) {
    for (const mode of ["basic", "off"]) {
      manifest.instrumentation.mode = mode;
      writeFileSync(manifestFile, JSON.stringify(manifest));
      for (const ambient of [undefined, mode === "basic" ? "off" : "basic"]) {
        const result = invoke(caseName, ambient);
        assert.equal(result.status, 1, result.stderr);
        assert.match(result.stderr, /failed with status 73/u);
        assert.deepEqual(readFileSync(receipt, "utf8").trim().split("\n"),
          ["--no-print-directory", "test-slice", "SUITE=backend-unit", `HARNESS_DIAGNOSTICS=${mode}`]);
      }
    }
    for (const invalid of [null, "{", JSON.stringify({ ...manifest, run_id: "foreign" }),
      JSON.stringify({ ...manifest, instrumentation: { ...manifest.instrumentation, mode: "unknown" } }),
      JSON.stringify(manifest).replace('{', '{"run_id":"duplicate",'), " ".repeat(1024 * 1024 + 1)]) {
      rmSync(manifestFile, { force: true });
      if (invalid !== null) writeFileSync(manifestFile, invalid);
      const result = invoke(caseName, "basic");
      assert.equal(result.status, 1, result.stderr);
      assert.equal(existsSync(receipt), false, "invalid parent evidence must not start Make");
    }
    rmSync(manifestFile);
    const alias = path.join(parent, "alias-manifest.json");
    writeFileSync(alias, JSON.stringify(manifest));
    symlinkSync(alias, manifestFile);
    assert.equal(invoke(caseName).status, 1);
    assert.equal(existsSync(receipt), false);
    rmSync(manifestFile);
  }
});

test("logical reporting metadata cannot introduce executable commands", () => {
  const root = path.resolve(import.meta.dirname, "../../..");
  const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
  assert.deepEqual(collectTaskSurfaceManifestErrors(manifest), []);
  const check = manifest.harness_checks[0];
  for (const command of [["$(MAKE)", "ci"], ["$(NODE_BIN)", "script.mjs"], null, []]) {
    check.command = command;
    assert.ok(collectTaskSurfaceManifestErrors(manifest).includes(
      `${check.name}.command is forbidden in logical reporting metadata; canonical operations own execution`,
    ));
  }
});

test("complete help exposes contracts without runtime preparation and rejects unsafe inputs", (t) => {
  const root = path.resolve(import.meta.dirname, "../../..");
  const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
  const help = manifest.targets.find((entry) => entry.name === "help");
  help.input_contract.inputs = ["ALL", "JSON"].map((name) => ({ name, type: "exact_1_bool", allowed_sources: ["make_command_line"], required: false, default: null }));
  const scratch = mkdtempSync(path.join(tmpdir(), "help-make-admission-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  mkdirSync(path.join(scratch, "tools"));
  for (const [name, all] of [["help", false], ["help-all", true]]) writeFileSync(path.join(scratch, `tools/task_surface.${name}.generated.json`), JSON.stringify(helpProjection(manifest, { all })));
  const file = path.join(scratch, "Makefile");
  writeFileSync(file, `${renderTaskSurfaceMakeRuntime(manifest)}\n${renderTaskSurfaceMake(manifest)}\nNODE_BIN := /absent/node\nTASK_SURFACE_NODE_READINESS = exit 99\nRUN_PUBLIC_PREFLIGHT = exit 98\n`);
  const env = { ...process.env, CARTULARY_OUTPUT_MODE: "summary" };
  for (const key of ["MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", "GNUMAKEFLAGS", "CARTULARY_TEST_TARGET"]) delete env[key];
  const invoke = (args, overrides = {}) => spawnSync("make", ["--silent", "--no-print-directory", "-f", file, "help", ...args], { cwd: scratch, env: { ...env, ...overrides }, encoding: "utf8", timeout: 10000, maxBuffer: 4 * 1024 * 1024 });
  const complete = invoke(["ALL=1", "JSON=1"]);
  assert.equal(complete.status, 0, complete.stderr);
  assert.deepEqual(JSON.parse(complete.stdout), helpProjection(manifest, { all: true }));
  const compact = invoke(["JSON=1"], { ALL: "1" });
  assert.equal(compact.status, 0, compact.stderr);
  assert.deepEqual(JSON.parse(compact.stdout), helpProjection(manifest));
  assert.equal(invoke(["ALL= 1 ", "JSON=1"]).status, 0);
  for (const args of [["ALL="], ["JSON=0"], ["PART=node"], ["ALL=1 2"], ["JSON=$(shell touch injected)"]]) {
    const result = invoke(args);
    assert.equal(result.status, 2, result.stderr);
    assert.equal(result.stdout, "");
  }
  assert.equal(invoke([], { CARTULARY_OUTPUT_MODE: "machine" }).status, 2);
  assert.equal(existsSync(path.join(scratch, "injected")), false);
});

test("bootstrap admits a closed selector before installing Node and dispatches only its private operation", (t) => {
  const root = path.resolve(import.meta.dirname, "../../..");
  let manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
  const entry = manifest.targets.find((value) => value.name === "bootstrap");
  entry.input_contract.inputs = [{ name: "PART", type: "enum", values: ["node", "frontend", "browser", "test-images"],
    binding: "make_variable", child_forwarding: "argv", allowed_sources: ["make_command_line"], required: false, default: null }];
  entry.selection_binding = { kind: "selector", input: "PART", variants: { node: "harness.bootstrap_node", frontend: "harness.bootstrap_frontend",
    browser: "harness.bootstrap_browser", "test-images": "harness.bootstrap_test_images" }, default_selection: "harness.bootstrap_full" };
  manifest.make_recipes.bootstrap = { type: "bootstrap", prerequisites: [] };
  manifest = projectOperationBindings(manifest, readOperationOwner(root));
  const scratch = mkdtempSync(path.join(tmpdir(), "bootstrap-selector-admission-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  const events = path.join(scratch, "events");
  const node = path.join(scratch, "node");
  const source = path.join(scratch, "fake-node");
  writeFileSync(source, `#!/bin/sh\nprintf '%s\\n' "$*" >> '${events}'\n`, { mode: 0o700 });
  mkdirSync(path.join(scratch, "tools"));
  for (const [name, all] of [["help", false], ["help-all", true]]) writeFileSync(path.join(scratch, `tools/task_surface.${name}.generated.json`), JSON.stringify(helpProjection(manifest, { all })));
  const file = path.join(scratch, "Makefile");
  writeFileSync(file, `${renderTaskSurfaceMakeRuntime(manifest)}\n${renderTaskSurfaceMake(manifest)}\nNODE_BIN := ${node}\nRUN_PUBLIC_PREFLIGHT = printf 'preflight\\n' >> ${events}\nRUN_TARGET_SUMMARY = :\n${node}:\n\tcp ${source} ${node}\n\tprintf 'install\\n' >> ${events}\n`);
  const env = { ...process.env };
  for (const key of ["MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", "GNUMAKEFLAGS", "CARTULARY_TEST_TARGET"]) delete env[key];
  const invoke = (args, overrides = {}) => spawnSync("make", ["--silent", "--no-print-directory", "-f", file, "bootstrap", ...args],
    { cwd: scratch, env: { ...env, ...overrides }, encoding: "utf8", timeout: 10000 });
  for (const args of [["PART="], ["PART=unknown"], ["PART=node frontend"], ["PART=$(shell touch injected)"], ["UNDECLARED=1"]]) {
    const result = invoke(args);
    assert.equal(result.status, 2, result.stderr);
    assert.equal(existsSync(node), false);
    assert.equal(existsSync(events), false);
  }
  assert.equal(existsSync(path.join(scratch, "injected")), false);
  for (const [part, operation] of [...Object.entries(entry.selection_binding.variants), [null, entry.selection_binding.default_selection]]) {
    rmSync(node, { force: true });
    rmSync(events, { force: true });
    const result = invoke(part ? [`PART=${part}`] : [], { PART: "node" });
    assert.equal(result.status, 0, result.stderr);
    assert.deepEqual(readFileSync(events, "utf8").trim().split("\n"), ["install", "preflight", `tools/harness/execution/operation-runner-cli.mjs ${operation}`]);
  }
});

test("required literal recovery inputs reach wrapper admission and its physical child", (t) => {
  const root = path.resolve(import.meta.dirname, "../../..");
  const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
  const scratch = mkdtempSync(path.join(tmpdir(), "recovery-make-transport-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  const probe = path.join(scratch, "probe.mjs");
  writeFileSync(probe, `import {execFileSync} from 'node:child_process';
const stage=process.argv[2];
console.log(JSON.stringify({stage,values:['RESULTS_DIR','RUNTIME_ROOT','RUNTIME_LEASE_ID'].map(key=>process.env[key]),sources:process.env.CARTULARY_MAKE_INPUT_SOURCES}));
if(stage==='wrapper'){const args=process.argv.slice(process.argv.indexOf('--')+1);process.stdout.write(execFileSync(args[0],args.slice(1),{env:process.env}));}
`);
  // Only the physical child and readiness are fixtures. The generated recipe,
  // literal transport, environment stripping and wrapper admission stay real.
  manifest.make_recipes["test-run-recover"].prerequisites = [];
  manifest.make_recipes["test-run-recover"].command = ["$(NODE_BIN)", probe, "child"];
  mkdirSync(path.join(scratch, "tools"));
  for (const [name, all] of [["help", false], ["help-all", true]]) writeFileSync(path.join(scratch, `tools/task_surface.${name}.generated.json`), JSON.stringify(helpProjection(manifest, { all })));
  const file = path.join(scratch, "Makefile");
  writeFileSync(file, `${renderTaskSurfaceMakeRuntime(manifest)}
${renderTaskSurfaceMake(manifest)}
TASK_SURFACE_NODE_READINESS = :
RUN_PUBLIC_PREFLIGHT = :
NODE_BIN := ${process.execPath}
RUN_STEP_SCRIPT := ${process.execPath} ${probe} wrapper
`);
  const values = ["/fixture/result with space", "/fixture/runtime 'quote' $dollar `literal`", "139eb0d9-ba96-4ce6-8f64-a8d5973e20b0"];
  const env = { ...process.env };
  for (const key of ["MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", "GNUMAKEFLAGS", "CARTULARY_TEST_TARGET"]) delete env[key];
  const output = execFileSync("make", ["--silent", "--no-print-directory", "-f", file, "test-run-recover", ...["RESULTS_DIR", "RUNTIME_ROOT", "RUNTIME_LEASE_ID"].map((key,index)=>`${key}=${values[index]}`)], { cwd:scratch, env, encoding:"utf8", timeout:10000 });
  const observations = output.trim().split("\n").map(line=>JSON.parse(line));
  assert.deepEqual(observations.map(value=>value.stage), ["wrapper", "child"]);
  for (const observation of observations) {
    assert.deepEqual(observation.values, values);
    assert.equal(observation.sources, "RESULTS_DIR=cli RUNTIME_LEASE_ID=cli RUNTIME_ROOT=cli");
  }
});

test("build input discovery excludes human material and rejects incomplete or unsafe file sets", () => {
  const root = path.resolve(import.meta.dirname, "../../..");
  execFileSync("bash", ["tools/harness/readiness/tests/test-build-input-discovery.sh"], {
    cwd: root,
    env: { ...process.env, NODE_BIN: process.execPath },
    timeout: 30_000,
    maxBuffer: 8192,
  });
});


test("planning and execution reject oversized exact row inputs before selection", async () => {
  const { resolveCommandInputs, resolveDeclaredTargetInputs } = await import("../contract/index.mjs");
  const root = path.resolve(import.meta.dirname, "../../..");
  const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
  const rows = Array.from({ length: 150 }, (_, index) => `harness.browser.unit.synthetic_selection_${index}`).join(",");
  assert.ok(rows.length > 4096);
  assert.throws(() => resolveCommandInputs("test-slice", { OWNER: "harness.browser", ROWS: rows }, manifest), /exceeds 4096/u);
  assert.throws(() => resolveDeclaredTargetInputs("test-slice", manifest.targets.find((entry) => entry.name === "test-slice"), manifest,
    { OWNER: "harness.browser", ROWS: rows, CARTULARY_MAKE_INPUT_SOURCES: "OWNER=cli ROWS=cli" }), /exceeds 4096/u);
});
