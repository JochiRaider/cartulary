import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { renderTaskSurfaceMake } from "../generated-artifacts/task-surface/make-renderer.mjs";
import { validateMakeRecipes } from "../generated-artifacts/task-surface/recipe-validation.mjs";
import { WorkGraphCompiler } from "../scheduler/work-graph/index.mjs";

const root = new URL("../../../", import.meta.url).pathname;
const compiler = new WorkGraphCompiler(root);
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

  // Use the actual transitive build headers, with harmless instrumented recipes.
  const source = readFileSync(path.join(root, "Makefile"), "utf8");
  const bindings = source.slice(source.indexOf("ifeq ($(CARTULARY_HARNESS_GRAPH_CHILD),1)\nFRONTEND_BUILD_PREREQUISITE"), source.indexOf("$(EMBEDDED_WEB_ASSET_STAMP) $(EMBEDDED_WEB_ASSET_ARCHIVE) $(EMBEDDED_CLIENT_ASSET_MANIFEST) $(EMBEDDED_CLIENT_SUPPORT_REGISTRY) $(EMBEDDED_WEB_ASSET_READY_STAMP) &:"));
  assert.ok(bindings.length > 0);
  const headers = source.split("\n").filter((line) => ["$(WEB_DIST_INDEX):", "$(EMBEDDED_WEB_ASSET_STAMP) $(EMBEDDED_WEB_ASSET_ARCHIVE)", "$(SERVER_HARNESS_BIN):"].some((prefix) => line.startsWith(prefix)));
  const names = ["embedded", "server", "web"];
  const ignored = new Set(headers.flatMap((line) => line.split(" ").filter((token) => token.startsWith("tools/") || token === "go.mod")));
  writeFileSync(makefile, `.SECONDEXPANSION:\nWEB_DIST_INDEX := web\nEMBEDDED_WEB_ASSET_STAMP := embedded\nEMBEDDED_WEB_ASSET_ARCHIVE := archive\nEMBEDDED_WEB_ASSET_READY_STAMP := ready\nEMBEDDED_CLIENT_ASSET_MANIFEST := client\nEMBEDDED_CLIENT_SUPPORT_REGISTRY := support\nSERVER_HARNESS_BIN := server\n` + bindings +
    headers.map((line, i) => `${line}\n\t@echo ${names[i]}\n`).join("\n") +
    `\n.PHONY: FORCE\nFORCE go-toolchain-readiness ${[...ignored].join(" ")}:\n`);
  const invoke = (graph) => execFileSync("make", ["--silent", "--no-print-directory", "-f", makefile, "server", `CARTULARY_HARNESS_GRAPH_CHILD=${graph}`], { cwd: scratch, encoding: "utf8" }).trim().split("\n");
  assert.deepEqual(invoke("1"), ["server"], "graph binary consumer never invokes upstream producers");
  assert.deepEqual(invoke("0"), ["web", "embedded", "server"], "standalone file prerequisites still compose production");
} finally { rmSync(scratch, { recursive: true, force: true }); }
