// The command entry (P5-3e) launches a graph target with the environment its
// Make recipe used to assemble: no public input but the target's forwarded
// ones, each with its origin; the resolved machine state; the launch
// bindings; and the target's selection.
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { bootstrapOperation, graphLaunch, publicRecipe, recipeProducers, stepInvocation, targetEnvironment, toolInvocation } from "../command-entry.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
const base = {
  PATH: "/usr/bin", HOME: "/home/fixture", HARNESS_RUNTIME_BIN: "/toolbin/cartulary-harness", MAKE: "make", NODE_BIN: "/node/bin/node",
  TEST_SERVICES_BIN: "/toolbin/cartulary-test-services", GO_CACHE_DIR: "/machine/go/build", CARTULARY_MACHINE_CACHE_DIR: "/machine",
};

test("the entry sets each target's own exports and run context, as its Make rule did", () => {
  const ci = targetEnvironment("ci", { ...base, CI: "0" }, { manifest, root });
  assert.equal(ci.CI, "1");
  assert.equal(ci.CARTULARY_TEST_TARGET, "ci");
  // A target that reports for itself keeps the public target that invoked it.
  assert.equal(targetEnvironment("check", { ...base, CARTULARY_TEST_TARGET: "release-check" }, { manifest, root }).CARTULARY_TEST_TARGET, "release-check");
  const gate = targetEnvironment("seaweedfs-release-gate", { ...base, CARTULARY_TEST_RUN_ID: "run-1" }, { manifest, root });
  // The gate declares its existing run root; it exports only its run identity.
  assert.equal(manifest.make_recipes["seaweedfs-release-gate"].existing_run_root, true);
  assert.deepEqual(Object.keys(manifest.make_recipes["seaweedfs-release-gate"].exports), ["CARTULARY_TEST_RUN_ID"]);
  assert.equal(gate.CARTULARY_TEST_RUN_ID, "run-1");
  // Other targets report for whichever target invoked them.
  assert.equal(Object.hasOwn(targetEnvironment("doctor", base, { manifest, root }), "CARTULARY_TEST_TARGET"), false);
});

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
  assert.deepEqual(recipeProducers(root, gate.recipe), ["frontend-install", "harness-runtime"]);
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
  assert.throws(() => graphLaunch("doctor", base, { manifest, root }), /does not launch a graph/u);
  assert.throws(() => graphLaunch("frontend-install", base, { manifest, root }), /not a public target/u);
  assert.throws(() => graphLaunch("check", { ...base, HARNESS_RUNTIME_BIN: "" }, { manifest, root }), /HARNESS_RUNTIME_BIN/u);
});

test("a step command runs its child under the step runner with the recipe's environments", () => {
  const env = { ...base, RUN_STEP_SCRIPT: "/repo/run-step.sh", GO: "/go/bin/go", SQLC_BIN: "/toolbin/sqlc", RUN_STEP_EXTRA: "kept",
    GO_MOD_CACHE_DIR: "/machine/go/mod", GO_TMP_DIR: "/machine/go/tmp", CARTULARY_MAKE_INPUT_SOURCES: "" };
  const generate = stepInvocation("generate", env, { manifest, root });
  assert.equal(generate.command, env.RUN_STEP_SCRIPT);
  assert.deepEqual(generate.args.slice(0, 3), ["generate", "--", "env"]);
  assert.equal(generate.env.CARTULARY_TEST_TARGET, "generate");
  // The child strips every public input, then sets the machine state and
  // the recipe's own variables, and runs the operation's command.
  const child = generate.args.slice(3);
  assert.ok(child.includes("-u") && child.includes("SQLC_BIN=/toolbin/sqlc") && child.includes("GO_CACHE_DIR=/machine/go/build"));
  assert.deepEqual(child.slice(-2), ["bash", "./tools/harness/generated-artifacts/generate-artifacts.sh"]);
  // A run-step recipe reports under its step label.
  const lockfile = stepInvocation("frontend-lockfile-update", env, { manifest, root });
  assert.equal(lockfile.args[0], manifest.make_recipes["frontend-lockfile-update"].step_label);
  // A summarized command keeps the public target that invoked it.
  assert.equal(stepInvocation("services-up", env, { manifest, root }).env.CARTULARY_TEST_TARGET, "services-up");
  assert.equal(stepInvocation("services-up", { ...env, CARTULARY_TEST_TARGET: "check" }, { manifest, root }).env.CARTULARY_TEST_TARGET, "check");
  // A recipe without a step summary runs its child directly.
  assert.equal(stepInvocation("dev", env, { manifest, root }).command, "env");
});

test("a Node tool runs through the tool launcher with every public input's origin", () => {
  const env = { ...base, OWNER: "harness.command_surface", JSON: "1", CARTULARY_MAKE_INPUT_SOURCES: "OWNER=cli JSON=cli" };
  const guide = toolInvocation("task-guide", env, { manifest, root });
  assert.equal(guide.command, path.join(root, "tools/harness/execution/run-make-node-tool.sh"));
  assert.deepEqual(guide.args, [manifest.make_recipes["task-guide"].tool_id]);
  assert.equal(guide.env.OWNER, env.OWNER);
  assert.equal(guide.env.CARTULARY_TEST_TARGET, "task-guide");
  assert.equal(guide.env.CARTULARY_HARNESS_IDENTITY_PREPARED, "1");
  const sources = new Map(guide.env.CARTULARY_MAKE_INPUT_SOURCES.split(" ").map((entry) => entry.split("=")));
  assert.equal(sources.get("OWNER"), "cli");
  assert.equal(sources.get("CHECK"), "unset");
});

test("bootstrap selects only its command-line selector, and Make keeps its Node-free targets", () => {
  const variants = manifest.targets.find((entry) => entry.name === "bootstrap").selection_binding;
  assert.equal(bootstrapOperation("bootstrap", { PART: "node", CARTULARY_MAKE_INPUT_SOURCES: "PART=cli" }, { manifest }), variants.variants.node);
  assert.equal(bootstrapOperation("bootstrap", { PART: "node", CARTULARY_MAKE_INPUT_SOURCES: "PART=env" }, { manifest }), variants.default_selection);
  for (const target of ["help", "lint-markdown", "test-run-status", "clean", "distclean", "ui-review"]) {
    assert.throws(() => publicRecipe(manifest, target), /does not run through the command entry/u, target);
  }
});
