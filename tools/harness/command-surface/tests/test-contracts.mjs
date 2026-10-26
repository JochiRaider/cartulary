import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, readdirSync, rmSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";
import { createContractTestContext } from "../../tests/contract-test-context.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}
export function registerBoundariesTests() {
  test("relative_backing_script_paths: task-surface backing scripts are repository-relative", async () => {
    const context = createContractTestContext(root);
    assert.ok(context.taskSurface.targets.every((entry) => (entry.backing_scripts ?? []).every((file) => !path.isAbsolute(file))));
  });
  test("current_owner_and_topology_identity: current task-surface and topology identities are exact", async () => {
    const context = createContractTestContext(root);
    assert.equal(context.taskSurface.schema_id, "cartulary.task_surface_owner.v4");
    assert.equal(context.topology.schema_id, "cartulary.execution_topology.v16");
  });
  test("current_target_roster: current target roster and public surface are exact", async () => {
    const context = createContractTestContext(root);
    const identity = (entry) => [entry.name, entry.command_id, entry.target_class];
    const targets = context.taskSurface.targets;
    assert.equal(new Set(targets.map((entry) => entry.name)).size, targets.length);
    const publicTargets = targets.filter((entry) => entry.target_class === "public");
    assert.equal(new Set(publicTargets.map((entry) => entry.command_id)).size, publicTargets.length);
    assert.deepEqual(context.readJSON("tools/task_surface_manifest.json").targets.map(identity), targets.map(identity));
  });
  test("retired_target_absence: retired public target aliases remain absent", async () => {
    const context = createContractTestContext(root);
    const retiredTargets = readJSON("tools/harness/tests/retired-make-names.json").retired_names;
    assert.ok(retiredTargets.every((target) => !context.taskSurface.targets.some((entry) => entry.name === target)));
    const results = mkdtempSync(path.join(tmpdir(), "cartulary-retired-commands-"));
    try {
      const env = { ...process.env, CARTULARY_TEST_RESULTS_DIR: results, CARTULARY_TEST_RUN_ID: "retired-rejection" };
      delete env.MAKEFLAGS;
      delete env.MAKEOVERRIDES;
      delete env.MFLAGS;
      let failure;
      try {
        execFileSync("make", ["--no-print-directory", "--keep-going", "--", ...retiredTargets], { cwd: root, env, stdio: "pipe" });
      }
      catch (error) {
        failure = error;
      }
      assert.equal(failure?.status, 2, "every retired command rejects before dispatch");
      const output = `${failure.stdout}${failure.stderr}`;
      for (const target of retiredTargets)
        assert.ok(output.includes(`No rule to make target '${target}'`), `${target} must have no callable rule`);
      assert.deepEqual(readdirSync(results), [], "rejected commands must not prepare retained execution");
    }
    finally {
      rmSync(results, { recursive: true, force: true });
    }
  });
  test("retired_topology_field_absence: retired schedule and fixture fields remain absent", async () => {
    const context = createContractTestContext(root);
    assert.equal("sequence_schedules" in context.topology, false);
    assert.equal("check_schedules" in context.topology, false);
    assert.equal("service_backed_schedules" in context.topology, false);
    assert.equal("fixture_profiles" in context.topology, false);
  });
}
export function registerCommandSurfaceTests() {
  test("public_command_identity_contract: public command identities are closed and versioned", async () => {
    const context = createContractTestContext(root);
    assert.ok(context.taskSurface.targets.filter((entry) => entry.target_class === "public").every((entry) => /^cartulary\.harness\.command\.[a-z0-9_]+\.v[1-9][0-9]*$/u.test(entry.command_id)));
  });
  test("public_output_contract: observable targets publish current run and target summaries", async () => {
    const context = createContractTestContext(root);
    assert.ok(!JSON.stringify(context.taskSurface.make_recipes).includes("nested_scheduler"));
    assert.ok(context.taskSurface.observability_policy.required_targets.every((target) => {
      const entry = context.taskSurface.targets.find((candidate) => candidate.name === target);
      return entry.output_policy.summary_schema === "cartulary.harness_run_summary.v1" &&
        entry.output_policy.artifact_policy === "run_and_target_summaries";
    }));
  });
}
