import assert from "node:assert/strict";
import test from "node:test";
import { execFileSync, spawnSync } from "../workspace/child-process.mjs";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { renderTaskSurfaceMake } from "../generated-artifacts/task-surface/make-renderer.mjs";
import { validateMakeRecipes } from "../generated-artifacts/task-surface/recipe-validation.mjs";
import { readOperationOwner } from "../execution/operations.mjs";
import { loadExecutionTopology } from "../generated-artifacts/execution-topology.mjs";

const root = new URL("../../../", import.meta.url).pathname;
export function registerFrontendProducerGraphTests() {
  test("frontend producers stay private and graph children never rebuild them through Make", () => {
    const owner = readOperationOwner(root);
    const surface = loadExecutionTopology({ root }).taskSurface;
    for (const member of owner.target_members["harness.contract_default"]) {
      const command = typeof member === "string" ? surface.make_recipes[member].command : owner.operations[member.operation_id].execution.command;
      // Evidence keeps its separately reserved four-file context.
      if (command[0] === "$(NODE_BIN)" && member.operation_id !== "harness.harness_evidence_contract") assert.ok(command.includes("--test-concurrency=1"), `${JSON.stringify(member)} bounds Node test-file workers`);
    }
    for (const name of ["protocol-ts-browser-artifact-reachability", "frontend-artifact-consumer-check", "build-web", "build-web-measurement"]) {
      assert.equal(surface.make_recipes[name], undefined, "private producers are not Make bindings");
    }

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
