import { randomUUID } from "node:crypto";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { processProof } from "./process-proof.mjs";
import { executionObservation } from "./execution-observations.mjs";

export const launchContextKey = "CARTULARY_HARNESS_LAUNCH_CONTEXT";

export function readLaunchContext(environment) {
  if (!environment[launchContextKey]) return null;
  const context = parseStrictJSON(environment[launchContextKey]);
  validateSchemaSync("cartulary.harness_launch_context.v1", context);
  if (context.run_id !== (environment.CARTULARY_TEST_RUN_ID || "standalone")) throw new Error("launch context run mismatch");
  return context;
}

export function createLaunchContext({ repoRoot, environment, unitID = null, commandID = null, rowIDs = [], producer = "private_capture" }) {
  let parent = null;
  try { parent = readLaunchContext(environment); }
  catch { /* Reject stale or malformed inheritance; this attempt gets a fresh scope. */ }
  const context = Object.freeze({ schema_id: "cartulary.harness_launch_context.v1",
    run_id: environment.CARTULARY_TEST_RUN_ID || "standalone", invocation_id: randomUUID(),
    parent_invocation_id: parent?.invocation_id ?? null,
    unit_id: unitID ?? parent?.unit_id ?? null, command_id: commandID ?? parent?.command_id ?? null,
    row_ids: [...new Set(rowIDs.length ? rowIDs : parent?.row_ids ?? [])].sort(), producer });
  validateSchemaSync(context.schema_id, context);
  const { schema_id: _schema, ...identity } = context;
  let record = { kind: "launch", ref: `launch:${context.invocation_id}`, ...identity, launched: false, outcome: "incomplete" };
  let settled = false;
  const observation = executionObservation({ repoRoot, environment }, record);
  return {
    identity: context,
    environment: { [launchContextKey]: JSON.stringify(context) },
    spawned(pid) {
      if (settled) return;
      record = { ...record, launched: true };
      let proof;
      // Unit and allocation owners register directly at spawn. Nested captures
      // carry their proof privately so short-lived children remain explainable.
      if (environment.CARTULARY_HARNESS_EXECUTION_ENABLED === "1" && !["unit", "browser_acquisition", "browser_review"].includes(producer)) {
        try { proof = processProof(pid); } catch { /* A launch remains known without physical proof. */ }
      }
      observation.update(record, proof);
    },
    settle(outcome) {
      if (settled) return;
      settled = true; record = { ...record, outcome }; observation.update(record);
    },
  };
}
