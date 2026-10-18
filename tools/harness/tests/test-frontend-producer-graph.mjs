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
    for (const member of compiler.owner.target_members["harness.contract_default"]) {
      const command = typeof member === "string" ? compiler.taskSurface.make_recipes[member].command : compiler.owner.operations[member.operation_id].execution.command;
      // Evidence keeps its separately reserved four-file context.
      if (command[0] === "$(NODE_BIN)" && member.operation_id !== "harness.harness_evidence_contract") assert.ok(command.includes("--test-concurrency=1"), `${JSON.stringify(member)} bounds Node test-file workers`);
    }
    const contracts = compiler.compile({ kind: "target", target: "harness-contract" }).units.find((unit) => unit.unit_id === "row:harness.command_surface.behavior.public_registry_parity");
    const contractRow = compiler.compile({ kind: "owner", owner_id: "harness.command_surface" }).units.find((unit) => unit.unit_id === "row:harness.command_surface.behavior.public_registry_parity");
    assert.deepEqual(contractRow.resource_claims, contracts.resource_claims, "direct and aggregate contract reservations agree");
    assert.ok(contracts.resource_claims.process >= 16, "reservation includes measured diagnostic wrappers and bounded producer children");
    const target = "harness.protocol_ts_browser_artifact_reachability";
    const rowID = "package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b";
    const selected = compiler.compile({ kind: "rows", row_ids: [rowID] });
    const combinedTarget = "harness.lint_protocol";
    assert.ok(compiler.compilePrivateSelection(combinedTarget).units.some((unit) => unit.current_run_evidence_outputs.includes(`rows/${rowID}.json`)));
    const unknown = new WorkGraphCompiler(root);
    unknown.owner.target_members[combinedTarget] = [{ row_id: "harness.browser.absent" }];
    assert.throws(() => unknown.compilePrivateSelection(combinedTarget), /unknown active row/u);
    const consumer = selected.units.find((unit) => unit.unit_id === `row:${rowID}`);
    assert.deepEqual(consumer.needs, ["operation:harness.build_web"], "bundle-security requires its canonical producer before assertion");
    for (const selection of [
      { kind: "target", target },
      { kind: "owner", owner_id: "package.protocol_ts", row_ids: [rowID] },
      { kind: "aggregate", target: "check" },
    ]) {
      const graph = compiler.compile(selection);
      assert.deepEqual(graph.units.find((unit) => unit.unit_id === consumer.unit_id), consumer, "consumer identity and dependencies are selection independent");
      assert.equal(graph.units.filter((unit) => unit.unit_id === "operation:harness.build_web").length, 1);
    }
    for (const name of ["protocol-ts-browser-artifact-reachability", "frontend-artifact-consumer-check", "build-web", "build-web-measurement"]) {
      assert.equal(compiler.taskSurface.make_recipes[name], undefined, "private producers are not Make bindings");
    }

    const synthetic = new WorkGraphCompiler(root);
    synthetic.ensureCatalog();
    const extraID = "package.protocol_ts.boundary_support.synthetic_consumer";
    const extraTarget = "harness.frontend_synthetic_consumer";
    const extraRow = { ...synthetic.catalog.rowByID.get(rowID), row_id: extraID };
    synthetic.catalog.rowByID.set(extraID, extraRow);
    synthetic._rowTargets.set(extraID, extraTarget);
    synthetic.owner.policy_units[extraTarget] = { ...synthetic.owner.policy_units[target], needs: ["harness.build_web_measurement"] };
    const combined = synthetic.compileRows([rowID, extraID]);
    for (const profile of ["build-web", "build-web-measurement"]) {
      assert.equal(combined.units.filter((unit) => unit.unit_id === `operation:harness.${profile.replaceAll("-", "_")}`).length, 1);
    }
    synthetic.owner.policy_units[extraTarget].needs = ["absent-producer"];
    assert.throws(() => synthetic.compileRows([extraID]), /unknown harness operation/u);
    synthetic.owner.policy_units[extraTarget].needs = ["harness.build_web"];
    synthetic.owner.policy_units["harness.build_web"].needs = ["harness.build_web"];
    assert.throws(() => synthetic.compileRows([extraID]), /dependency cycle/u);

    // Exercise generated Make admission using an injected forbidden producer.
    // The fixture changes the assertion command only; its real prerequisite and
    // graph-child rendering must exclude production before that assertion runs.
    const scratch = mkdtempSync(path.join(os.tmpdir(), "frontend-producer-make-"));
    try {
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
