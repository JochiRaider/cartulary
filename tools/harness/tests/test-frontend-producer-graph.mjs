import assert from "node:assert/strict";
import test from "node:test";
import { execFileSync, spawnSync } from "../workspace/child-process.mjs";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { renderTaskSurfaceMake } from "../generated-artifacts/task-surface/make-renderer.mjs";
import { validateMakeRecipes } from "../generated-artifacts/task-surface/recipe-validation.mjs";
import { WorkGraphCompiler } from "../scheduler/work-graph/index.mjs";

const root = new URL("../../../", import.meta.url).pathname;
export function registerFrontendProducerGraphTests() {
  test("frontend producer graph selects canonical dependencies and Make admission", () => {
    const compiler = new WorkGraphCompiler(root);
    for (const name of compiler.owner.target_members["harness-contract"]) {
      const command = compiler.taskSurface.make_recipes[name].command;
      if (name !== "harness-evidence-contract") assert.ok(command.includes("--test-concurrency=1"), `${name} bounds Node test-file workers`);
    }
    const contracts = compiler.compile({ kind: "target", target: "harness-contract" }).units.find((unit) => unit.unit_id === "row:harness.command_surface.behavior.public_registry_parity");
    const contractRow = compiler.compile({ kind: "owner", owner_id: "harness.command_surface" }).units.find((unit) => unit.unit_id === "row:harness.command_surface.behavior.public_registry_parity");
    assert.deepEqual(contractRow.resource_claims, contracts.resource_claims, "direct and aggregate contract reservations agree");
    assert.ok(contracts.resource_claims.process >= 16, "reservation includes measured diagnostic wrappers and bounded producer children");
    const target = "protocol-ts-browser-artifact-reachability";
    const rowID = "package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b";
    const selected = compiler.compile({ kind: "rows", row_ids: [rowID] });
    const combinedTarget = "frontend-artifact-consumer-check";
    const combinedRecipe = compiler.taskSurface.make_recipes[combinedTarget];
    assert.deepEqual(compiler.compileTarget(combinedTarget), compiler.compileRows(combinedRecipe.row_ids), "fixed-row Make binding selects exactly the same canonical graph");
    for (const row_ids of [[], [rowID, rowID], ["invalid/row"], [...combinedRecipe.row_ids].reverse()]) {
      const errors = [];
      validateMakeRecipes(errors, new Map(compiler.taskSurface.targets.map((entry) => [entry.name, entry])), new Map(), { [combinedTarget]: { ...combinedRecipe, row_ids } });
      assert.ok(errors.some((error) => error.includes("row_ids")), JSON.stringify(errors));
    }
    const unknown = new WorkGraphCompiler(root);
    unknown.taskSurface.make_recipes[combinedTarget].row_ids = ["harness.browser.absent"];
    assert.throws(() => unknown.compileTarget(combinedTarget), /unknown active row/u);
    const consumer = selected.units.find((unit) => unit.unit_id === `row:${rowID}`);
    assert.deepEqual(consumer.needs, ["target:build-web"], "bundle-security requires its canonical producer before assertion");
    for (const selection of [
      { kind: "target", target },
      { kind: "owner", owner_id: "package.protocol_ts", row_ids: [rowID] },
      { kind: "aggregate", target: "check" },
    ]) {
      const graph = compiler.compile(selection);
      assert.deepEqual(graph.units.find((unit) => unit.unit_id === consumer.unit_id), consumer, "consumer identity and dependencies are selection independent");
      assert.equal(graph.units.filter((unit) => unit.unit_id === "target:build-web").length, 1);
    }
    assert.equal(compiler.taskSurface.make_recipes[target].graph_entry, true);
    assert.equal(compiler.taskSurface.make_recipes[target].graph_child_skips_prerequisites, true);
    for (const producer of ["build-web", "build-web-measurement"]) {
      assert.equal(compiler.taskSurface.make_recipes[producer].graph_entry, true);
    }

    const synthetic = new WorkGraphCompiler(root);
    synthetic.ensureCatalog();
    const extraID = `${rowID}_synthetic`;
    const extraTarget = "frontend-synthetic-consumer";
    const extraRow = { ...synthetic.catalog.rowByID.get(rowID), row_id: extraID };
    synthetic.catalog.rowByID.set(extraID, extraRow);
    synthetic._rowTargets.set(extraID, extraTarget);
    synthetic.owner.policy_units[extraTarget] = { ...synthetic.owner.policy_units[target], needs: ["build-web-measurement"] };
    const combined = synthetic.compileRows([rowID, extraID]);
    for (const profile of ["build-web", "build-web-measurement"]) {
      assert.equal(combined.units.filter((unit) => unit.unit_id === `target:${profile}`).length, 1);
    }
    synthetic.owner.policy_units[extraTarget].needs = ["absent-producer"];
    assert.throws(() => synthetic.compileRows([extraID]), /unknown task target/u);
    synthetic.owner.policy_units[extraTarget].needs = ["build-web"];
    synthetic.owner.policy_units["build-web"].needs = ["build-web"];
    assert.throws(() => synthetic.compileRows([extraID]), /dependency cycle/u);

    // Exercise generated Make admission using an injected forbidden producer.
    // The fixture changes the assertion command only; its real prerequisite and
    // graph-child rendering must exclude production before that assertion runs.
    const scratch = mkdtempSync(path.join(os.tmpdir(), "frontend-producer-make-"));
    try {
      const surface = structuredClone(compiler.taskSurface);
      surface.make_recipes[target].command = ["true"];
      const makefile = path.join(scratch, "Makefile");
      writeFileSync(path.join(scratch, "step"), '#!/bin/sh\nshift\nshift\nexec "$@"\n', { mode: 0o700 });
      writeFileSync(makefile, `NODE_BIN := true\nWEB_DIST_INDEX := forbidden-build\nRUN_STEP_SCRIPT := ${scratch}/step\nCARTULARY_HARNESS_GRAPH_CHILD := 1\n` +
        renderTaskSurfaceMake(surface) + '\nforbidden-build:\n\t@echo forbidden-producer >&2\n\t@exit 99\n');
      execFileSync("make", ["--silent", "--no-print-directory", "-f", makefile, target], { cwd: scratch });

      // Ask GNU Make for its resolved dependency graph. Comments, whitespace,
      // header ordering and variable spelling are not part of this boundary.
      const artifacts = { WEB_DIST_INDEX: path.join(scratch, "renamed-web"), EMBEDDED_WEB_ASSET_STAMP: path.join(scratch, "renamed-assets"), SERVER_HARNESS_BIN: path.join(scratch, "renamed-server") };
      for (const graph of ["0", "1"]) {
        const env = { ...process.env, MAKEFLAGS: "", MAKEOVERRIDES: "", MFLAGS: "" };
        const result = spawnSync("bash", ["tools/harness/workspace/admission.sh", "run", "--", "make", "--no-print-directory", "-qp", "-f", "Makefile", artifacts.SERVER_HARNESS_BIN, artifacts.EMBEDDED_WEB_ASSET_STAMP, `CARTULARY_HARNESS_GRAPH_CHILD=${graph}`,
          ...Object.entries(artifacts).map(([key, value]) => `${key}=${value}`),
          ...["SERVER", "MIGRATE", "OPERATOR", "WEB", "HARNESS_RUNTIME", "TEST_SERVICES"].map((name) => `${name}_BUILD_INPUTS=`),
        ], { cwd: root, env, encoding: "utf8", maxBuffer: 16 * 1024 * 1024 });
        assert.ok(result.status === 0 || result.status === 1, result.stderr);
        const dependencies = (target) => {
          const line = result.stdout.split("\n").find((entry) => entry.startsWith(`${target}:`));
          assert.ok(line, `Make omitted ${target}`);
          return line.slice(target.length + 1).trim().split(/\s+/u);
        };
        assert.equal(dependencies(artifacts.SERVER_HARNESS_BIN).includes(artifacts.EMBEDDED_WEB_ASSET_STAMP), graph === "0", `standalone server consumes embedded assets; graph child relies on its admitted producer: ${dependencies(artifacts.SERVER_HARNESS_BIN).join(" ")}`);
        assert.equal(dependencies(artifacts.EMBEDDED_WEB_ASSET_STAMP).includes(artifacts.WEB_DIST_INDEX), graph === "0", "standalone assets consume web output; graph child never rebuilds upstream work");
      }

    } finally { rmSync(scratch, { recursive: true, force: true }); }

  });
}
