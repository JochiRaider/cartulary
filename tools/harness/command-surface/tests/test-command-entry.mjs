// The command entry (P5-3e) launches a graph target with the environment its
// Make recipe used to assemble: no public input but the target's forwarded
// ones, each with its origin; the resolved machine state; the launch
// bindings; and the target's selection.
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { graphLaunch, graphProducers } from "../command-entry.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
const base = {
  PATH: "/usr/bin", HOME: "/home/fixture", HARNESS_RUNTIME_BIN: "/toolbin/cartulary-harness", MAKE: "make", NODE_BIN: "/node/bin/node",
  TEST_SERVICES_BIN: "/toolbin/cartulary-test-services", GO_CACHE_DIR: "/machine/go/build", CARTULARY_MACHINE_CACHE_DIR: "/machine",
};

test("a graph launch keeps only the target's forwarded inputs, with their origins", () => {
  const env = { ...base, OWNER: "harness.command_surface", ROWS: "harness.command_surface.unit.go_environment", BROWSER_LANES: "4",
    CHECK: "schema", CARTULARY_MAKE_INPUT_SOURCES: "OWNER=cli ROWS=cli BROWSER_LANES=file CHECK=cli UNDECLARED=cli", TASK_SURFACE_MANIFEST: "/elsewhere.json" };
  const slice = graphLaunch("test-slice", env, { manifest, root });
  assert.equal(slice.command, base.HARNESS_RUNTIME_BIN);
  assert.deepEqual(slice.args, ["graph", "--selection", "command", "--target", "test-slice"]);
  assert.equal(slice.env.OWNER, env.OWNER);
  assert.equal(slice.env.ROWS, env.ROWS);
  // Another target's input never reaches this one; machine state and the
  // launch bindings survive the strip.
  assert.equal(Object.hasOwn(slice.env, "CHECK"), false);
  for (const name of ["GO_CACHE_DIR", "CARTULARY_MACHINE_CACHE_DIR", "MAKE", "NODE_BIN", "TEST_SERVICES_BIN"]) assert.equal(slice.env[name], base[name], name);
  const sources = new Map(slice.env.CARTULARY_MAKE_INPUT_SOURCES.split(" ").map((entry) => entry.split("=")));
  assert.equal(sources.get("OWNER"), "cli");
  assert.equal(sources.get("BROWSER_LANES"), "file");
  assert.equal(sources.get("UNDECLARED"), undefined);
  for (const [name, origin] of sources) assert.ok(Object.hasOwn(slice.env, name), `${name} (${origin}) has a value`);
  // Aggregates and graph-entry operations select themselves.
  assert.deepEqual(graphLaunch("check", env, { manifest, root }).args, ["graph", "--selection", "aggregate", "--target", "check"]);
  const gate = graphLaunch("seaweedfs-release-gate", env, { manifest, root });
  assert.deepEqual(gate.args, ["graph", "--selection", "target", "--target", "seaweedfs-release-gate"]);
  assert.deepEqual(graphProducers(root, gate.recipe), ["frontend-install", "harness-runtime"]);
});

test("a canonical input is the checkout's own file, and only graph targets launch", () => {
  const targets = manifest.targets.filter((entry) => entry.target_class === "public" &&
    (entry.input_contract?.inputs ?? []).some((input) => input.name === "TASK_SURFACE_MANIFEST" && ["runtime_env", "argv_and_runtime_env"].includes(input.child_forwarding)));
  for (const entry of targets) {
    const recipe = manifest.make_recipes[entry.name];
    if (recipe.type !== "work_graph" && recipe.graph_entry !== true) continue;
    const launch = graphLaunch(entry.name, { ...base, TASK_SURFACE_MANIFEST: "/elsewhere.json" }, { manifest, root });
    assert.equal(launch.env.TASK_SURFACE_MANIFEST, path.join(root, "tools/task_surface_manifest.json"));
  }
  assert.throws(() => graphLaunch("help", base, { manifest, root }), /does not launch a graph/u);
  assert.throws(() => graphLaunch("frontend-install", base, { manifest, root }), /not a public target/u);
  assert.throws(() => graphLaunch("check", { ...base, HARNESS_RUNTIME_BIN: "" }, { manifest, root }), /HARNESS_RUNTIME_BIN/u);
});
