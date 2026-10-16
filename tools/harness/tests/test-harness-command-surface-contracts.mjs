import "../contract/tests/test-cleanup.mjs";
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";
import { runContractSuite } from "./contract-suite-support.mjs";
import { registerRunStatusCommandTests } from "../diagnostics/tests/run-observation-cases.mjs";
import { registerFrontendProducerGraphTests } from "./test-frontend-producer-graph.mjs";
import { registerFrontendProducerLifecycleTests } from "./test-frontend-producer-lifecycle.mjs";
import { registerCommandFailureTests } from "./test-command-failure.mjs";
import { renderTaskSurfaceMake, renderTaskSurfaceMakeRuntime } from "../generated-artifacts/task-surface/make-renderer.mjs";
registerFrontendProducerGraphTests();
registerFrontendProducerLifecycleTests();
registerCommandFailureTests();
runContractSuite("command_surface");
registerRunStatusCommandTests();

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
