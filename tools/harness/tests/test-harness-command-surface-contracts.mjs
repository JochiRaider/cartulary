import "../contract/tests/test-cleanup.mjs";
import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";
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
