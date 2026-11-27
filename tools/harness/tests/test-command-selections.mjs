import assert from "node:assert/strict";
import test from "node:test";
import { resolveCommandInputs } from "../contract/index.mjs";
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
  const { operations } = readOperationOwner(root);
  assert.equal(operations["harness.source_generated"].execution.command.includes("--refresh"), false);
  assert.equal(operations["harness.generated_refresh"].execution.command.at(-1), "--refresh");
});

test("installation operations produce their installations and artifacts before acquisition", () => {
  const owner = readOperationOwner(root);
  const environment = Object.fromEntries(operationBindingNames(owner).map((name) => [name, `/installed/${name}`]));
  Object.assign(environment, { MAKE: "make", CURDIR: root });
  // A producer call runs the published launcher, or `go run` of its sources
  // before it is published (P6-5).
  const produced = (invocation) => invocation.args[0] === "produce" ? invocation.args.slice(1)
    : invocation.args.slice(0, 3).join(" ") === "run ./tools/harnessbuild produce" ? invocation.args.slice(3) : null;
  const kind = (invocation) => invocation.command === "make" ? "make" : produced(invocation) ? "produce" : "command";
  // No installation re-enters Make; only explicit Go dependency maintenance
  // runs a command after its producers.
  const expected = { node: ["produce"], go: ["produce", "command"], frontend: ["produce"], browser: ["produce"], test_images: ["produce"], full: ["produce", "command"] };
  for (const [part, kinds] of Object.entries(expected)) {
    const invocations = buildOperationSequence(owner, `harness.bootstrap_${part}`, environment);
    assert.deepEqual(invocations.map(kind), kinds, part);
  }
  const go = buildOperationSequence(owner, "harness.bootstrap_go", environment);
  assert.deepEqual(produced(go[0]), ["node-runtime"]);
  assert.deepEqual(go[1].args.slice(-3), ["bash", "tools/harness/readiness/install-go-dependencies.sh", "--all"]);
  // Registered producers, installations among them, run through the store
  // with the build's own caches.
  const full = buildOperationSequence(owner, "harness.bootstrap_full", environment);
  assert.deepEqual(produced(full[0]).sort(), ["cyclonedx-gomod", "frontend-install", "goose", "gosec", "govulncheck", "harness-runtime", "playwright", "shellcheck", "sqlc", "staticcheck", "syft"]);
  assert.equal(full[0].env.GOCACHE, environment.GO_CACHE_DIR);
  assert.deepEqual(full[0].directories, [environment.GO_CACHE_DIR, environment.GO_MOD_CACHE_DIR, environment.GO_TMP_DIR]);
  // A prerequisite that names no registered producer is refused: nothing is
  // made by Make.
  const unproduced = structuredClone(owner);
  unproduced.operations["harness.release_inventory_artifacts"].execution.prerequisites = ["$(LICENSE_REPORT_ARTIFACT)"];
  assert.throws(() => buildOperationSequence(unproduced, "harness.release_inventory_artifacts", environment), /name no registered producer/u);
  assert.throws(() => buildOperationSequence(owner, "harness.bootstrap_full", { ...environment, GO_TOOLCHAIN: undefined }), /binding|template/u);
  assert.throws(() => buildOperationSequence(owner, "harness.bootstrap_test_images", { ...environment, GO_CACHE_DIR: "relative" }), /GO_CACHE_DIR/u);
});
