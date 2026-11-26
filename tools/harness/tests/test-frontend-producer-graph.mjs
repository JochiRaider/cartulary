import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import path from "node:path";
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

    // Binaries that embed the web assets wait for their producer in the graph;
    // their store keys cover the embedded files through their import closure.
    for (const operation of ["harness.build_server", "harness.build_server_harness", "harness.build_operator"]) {
      assert.ok(owner.policy_units[operation].needs.includes("harness.embedded_web_assets"), `${operation} waits for the embedded assets`);
    }

    // The web builds and the embedded assets have no Make rule: the graph
    // runs their producers, and packaging reads the run's frontend artifact.
    const producers = JSON.parse(readFileSync(path.join(root, "tools/harness_producers.json"), "utf8")).producers;
    const embedded = producers.find((entry) => entry.kind === "embedded_assets");
    assert.equal(embedded.source, "build-web", "embedded assets package the production artifact");
    assert.ok(owner.policy_units["harness.embedded_web_assets"].needs.includes("harness.build_web"), "packaging waits for the production artifact");
    for (const target of ["build-web", "build-web-measurement"]) {
      assert.equal(producers.filter((entry) => entry.kind === "web_bundle" && entry.target === target).length, 1, `${target} has one web bundle producer`);
    }

  });
}
