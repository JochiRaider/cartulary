// The command entry (P5-3e) launches a graph target with the environment its
// Make recipe used to assemble: no public input but the target's forwarded
// ones, each with its origin; the resolved machine state; the launch
// bindings; and the target's selection.
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { bootstrapOperation, graphLaunch, publicRecipe, recipeProducers, reviewInvocation, stepInvocation, targetEnvironment, toolInvocation } from "../command-entry.mjs";

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

test("a Node tool's environment is default-deny: declared names pass, nothing else does", () => {
  const env = { ...base, OWNER: "harness.command_surface", CARTULARY_MAKE_INPUT_SOURCES: "OWNER=cli",
    NODE_OPTIONS: "--inspect", DEBUG: "*", OTEL_EXPORTER_OTLP_ENDPOINT: "http://collector", UNDECLARED_SHELL_SETTING: "x",
    CARTULARY_OUTPUT_MODE: "machine", TEST_SERVICES_BIN: "/toolbin/cartulary-test-services" };
  const guide = toolInvocation("task-guide", env, { manifest, root });
  for (const name of ["NODE_OPTIONS", "DEBUG", "OTEL_EXPORTER_OTLP_ENDPOINT", "UNDECLARED_SHELL_SETTING"]) assert.equal(Object.hasOwn(guide.env, name), false, name);
  for (const name of ["PATH", "HOME", "CARTULARY_OUTPUT_MODE", "TEST_SERVICES_BIN"]) assert.equal(guide.env[name], env[name], name);
});

test("a UI review command runs its client through the entry with a default-deny environment", async () => {
  const env = { ...base, UI_SESSION: "session-1", CARTULARY_MAKE_INPUT_SOURCES: "UI_SESSION=cli", CARTULARY_OUTPUT_MODE: "machine",
    CARTULARY_TEST_RUN_ID: "run-1", NODE_OPTIONS: "--inspect", PWDEBUG: "1", OTEL_SERVICE_NAME: "x", OWNER: "not-an-input" };
  const status = await reviewInvocation("ui-review-status", env, { manifest, root });
  assert.equal(status.command, base.NODE_BIN);
  assert.deepEqual(status.args, [path.join(root, "tools/harness/browser/ui-review/cli.mjs"), "ui-review-status"]);
  for (const name of ["UI_SESSION", "CARTULARY_MAKE_INPUT_SOURCES", "CARTULARY_OUTPUT_MODE", "CARTULARY_TEST_RUN_ID", "PATH"]) assert.equal(status.env[name], env[name], name);
  for (const name of ["NODE_OPTIONS", "PWDEBUG", "OTEL_SERVICE_NAME", "OWNER"]) assert.equal(Object.hasOwn(status.env, name), false, name);
  await assert.rejects(() => reviewInvocation("task-guide", env, { manifest, root }), /not a UI review command/u);
});

test("an unavailable or incompatible Node is a structured configuration failure in the target's envelope", (t) => {
  const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-entry-node-"));
  t.after(() => rmSync(scratch, { recursive: true, force: true }));
  const wrong = path.join(scratch, "node");
  writeFileSync(wrong, "#!/bin/sh\necho v0.0.1\n", { mode: 0o700 });
  const launcher = path.join(root, "tools/harness/command-surface/command-entry.sh");
  const entry = (nodeBin, ...args) => spawnSync("bash", [launcher, "installed", ...args], { env: { PATH: process.env.PATH, NODE_BIN: nodeBin }, encoding: "utf8" });
  for (const [nodeBin, condition] of [[path.join(scratch, "absent"), "missing"], [wrong, "incompatible"]]) {
    const review = entry(nodeBin, "ui-review-status", "ui_review");
    assert.equal(review.status, 2);
    const result = JSON.parse(review.stdout);
    assert.equal(result.schema_id, "cartulary.ui_review_command_result.v2");
    assert.equal(result.command_id, "cartulary.harness.command.ui_review_status.v1");
    assert.deepEqual([result.failures[0].subject_id, result.failures[0].condition, result.failures[0].recovery_id], ["node", condition, "bootstrap_node"]);
    const harness = entry(nodeBin, "task-guide");
    assert.equal(harness.status, 2);
    assert.equal(harness.stdout, "");
    assert.match(harness.stderr, new RegExp(`^\\[FAIL\\] target=task-guide failure_class=config failure_reason=configuration_error exit_code=2 subject=node condition=${condition} `, "u"));
  }
  assert.equal(entry(wrong, "ui-review-status", "unknown").status, 2, "only the declared envelopes are accepted");
});

test("bootstrap selects only its command-line selector, and Make keeps its Node-free targets", () => {
  const variants = manifest.targets.find((entry) => entry.name === "bootstrap").selection_binding;
  assert.equal(bootstrapOperation("bootstrap", { PART: "node", CARTULARY_MAKE_INPUT_SOURCES: "PART=cli" }, { manifest }), variants.variants.node);
  assert.equal(bootstrapOperation("bootstrap", { PART: "node", CARTULARY_MAKE_INPUT_SOURCES: "PART=env" }, { manifest }), variants.default_selection);
  for (const target of ["help", "lint-markdown", "test-run-status", "clean", "distclean"]) {
    assert.throws(() => publicRecipe(manifest, target), /does not run through the command entry/u, target);
  }
});
