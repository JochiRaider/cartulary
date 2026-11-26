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

test("installation operations make or produce their artifacts before acquisition", () => {
  const owner = readOperationOwner(root);
  const environment = Object.fromEntries(operationBindingNames(owner).map((name) => [name, `/installed/${name}`]));
  Object.assign(environment, { MAKE: "make", CURDIR: root });
  const kind = (invocation) => invocation.command === "make" ? "make"
    : invocation.args.slice(0, 3).join(" ") === "run ./tools/harnessbuild produce" ? "produce" : "command";
  const expected = { node: ["make"], go: ["make", "command"], frontend: ["make"], browser: ["make"], test_images: ["produce", "command"], full: ["make", "produce", "command"] };
  for (const [part, kinds] of Object.entries(expected)) {
    const invocations = buildOperationSequence(owner, `harness.bootstrap_${part}`, environment);
    assert.deepEqual(invocations.map(kind), kinds, part);
    for (const invocation of invocations.filter((entry) => kind(entry) === "make")) {
      assert.ok(invocation.args.slice(3).every((file) => path.isAbsolute(file)));
    }
  }
  const go = buildOperationSequence(owner, "harness.bootstrap_go", environment);
  assert.deepEqual(go[0].args.slice(3), [environment.NODE_BIN]);
  assert.deepEqual(go[1].args.slice(-3), ["bash", "tools/harness/readiness/install-go-dependencies.sh", "--all"]);
  // Registered producers run through the store with the build's own caches;
  // everything else stays a Make target, made first.
  const full = buildOperationSequence(owner, "harness.bootstrap_full", environment);
  assert.deepEqual(full[0].args.slice(3), [environment.FRONTEND_INSTALL_STAMP, environment.PLAYWRIGHT_INSTALL_STAMP]);
  assert.deepEqual(full[1].args.slice(3).sort(), ["cyclonedx-gomod", "goose", "gosec", "govulncheck", "harness-runtime", "shellcheck", "sqlc", "staticcheck", "syft"]);
  assert.equal(full[1].env.GOCACHE, environment.GO_CACHE_DIR);
  assert.deepEqual(full[1].directories, [environment.GO_CACHE_DIR, environment.GO_MOD_CACHE_DIR, environment.GO_TMP_DIR]);
  assert.throws(() => buildOperationSequence(owner, "harness.bootstrap_frontend", { ...environment, FRONTEND_INSTALL_STAMP: "retired-wrapper" }), /absolute file targets/u);
  assert.throws(() => buildOperationSequence(owner, "harness.bootstrap_full", { ...environment, GO_TOOLCHAIN: undefined }), /binding|template/u);
  assert.throws(() => buildOperationSequence(owner, "harness.bootstrap_test_images", { ...environment, GO_CACHE_DIR: "relative" }), /GO_CACHE_DIR/u);
});
