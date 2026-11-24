import assert from "node:assert/strict";
import test from "node:test";
import { resolveCommandInputs } from "../contract/index.mjs";
import path from "node:path";
import { buildOperationSequence, operationBindingNames, readOperationOwner } from "../execution/operations.mjs";
import { loadExecutionTopology } from "../generated-artifacts/execution-topology.mjs";

const root = new URL("../../../", import.meta.url).pathname;

test("selection inputs resolve canonically and selectors require their declared inputs", () => {
  const surface = loadExecutionTopology({ root }).taskSurface;
  const rowID = "harness.command_surface.unit.scheduler_smoke";
  const equivalent = { OWNER: "harness.command_surface", ROWS: `${rowID},harness.command_surface.unit.native_runtime` };
  const reverse = { ...equivalent, ROWS: equivalent.ROWS.split(",").reverse().join(",") };
  assert.deepEqual(resolveCommandInputs("test-slice", equivalent, surface), resolveCommandInputs("test-slice", reverse, surface));
  assert.throws(() => resolveCommandInputs("browser-e2e", {}, surface), /required/u);
  for (const invalid of [{ COMPONENT: "unknown" }, { COMPONENT: "" }, { COMPONENT: 1 }, { MAKE: "false" }, { GO_CACHE_DIR: "/tmp/foreign" }, { ARGS: "--override" }]) {
    assert.throws(() => resolveCommandInputs("build", invalid, surface), /must (?:not )?be|not declared/u);
  }
});

test("generated-source validation cannot refresh drift", () => {
  assert.equal(readOperationOwner(root).operations["harness.source_generated"].execution.env.CARTULARY_GENERATE_DRIFT_REFRESH, "0");
});

test("installation operations validate all file producers before acquisition", () => {
  const owner = readOperationOwner(root);
  const environment = Object.fromEntries(operationBindingNames(owner).map((name) => [name, `/installed/${name}`]));
  environment.MAKE = "make";
  for (const part of ["node", "go", "frontend", "browser", "test_images", "full"]) {
    const invocations = buildOperationSequence(owner, `harness.bootstrap_${part}`, environment);
    assert.equal(invocations[0].command, "make");
    assert.ok(invocations[0].args.slice(3).every((file) => path.isAbsolute(file)));
    assert.equal(invocations.length, ["go", "test_images", "full"].includes(part) ? 2 : 1);
  }
  const go = buildOperationSequence(owner, "harness.bootstrap_go", environment);
  assert.deepEqual(go[0].args.slice(3), [environment.NODE_BIN]);
  assert.deepEqual(go[1].args.slice(-3), ["bash", "tools/harness/readiness/install-go-dependencies.sh", "--all"]);
  assert.throws(() => buildOperationSequence(owner, "harness.bootstrap_frontend", { ...environment, FRONTEND_INSTALL_STAMP: "retired-wrapper" }), /absolute file targets/u);
  assert.throws(() => buildOperationSequence(owner, "harness.bootstrap_full", { ...environment, GO_TOOLCHAIN: undefined }), /binding|template/u);
});
