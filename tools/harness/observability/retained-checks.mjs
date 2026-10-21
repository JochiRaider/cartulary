import path from "node:path";
import { parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../contract/index.mjs";
import { reduceCanonicalUnitIntervals } from "../evidence-accounting/index.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";
import { validateExecutionContext } from "./canonical-context.mjs";
import { validateRetainedCleanup, validateRetainedSelection } from "./retained-selection.mjs";
import { loadRetainedObservability } from "./observability.mjs";

export const observabilityChecks = Object.freeze(["all", "events", "timing", "evidence", "diagnostics"]);
const read = (root, relative) => parseStrictJSON(readRetainedArtifactBytes(root, relative, 16 * 1024 ** 2).toString("utf8"));

async function eventsOnly(root) {
  const manifest = read(root, "run-manifest.json");
  validateSchemaSync("cartulary.harness_run_manifest.v3", manifest);
  const summary = read(root, "run-summary.json");
  validateSchemaSync("cartulary.harness_run_summary.v1", summary);
  if (manifest.run_id !== summary.run_id || manifest.target !== summary.target) throw new Error("event invocation identity mismatch");
  const events = await reduceCanonicalUnitIntervals(path.join(root, "unit-events.ndjson"));
  if (!events.runStarted || !events.runCompleted || events.registrations.size !== events.terminals.size ||
      events.runCompleted.seq !== events.eventCount || events.runCompleted.monotonic_ms !== summary.wall_duration_ms ||
      ({ pass: "passed", fail: "failed", cancelled: "cancelled" })[summary.status] !== events.runCompleted.status) {
    throw new Error("event invocation is incomplete");
  }
  return manifest;
}

function evidenceClosure(root, run) {
  if (run.summary.status !== "pass" || run.summary.failure_class !== null || run.summary.failure_reason !== null) throw new Error("retained invocation did not pass");
  const context = read(root, "execution-context.json");
  const cleanup = read(root, "cleanup-results.json");
  validateExecutionContext(context, run, cleanup);
  validateRetainedCleanup(run, cleanup);
  const scan = read(root, "retained-secret-scan.json");
  validateSchemaSync("cartulary.harness_retained_secret_scan.v1", scan);
  const units = new Map(context.graph.units.map((unit) => [unit.unit_id, unit]));
  validateRetainedSelection(root, run, units, run.targetSummaries.get(run.manifest.target));
}

// Scoped checks never return the required-evidence closure receipt. Optional
// observation completeness is independent of both validity and run success.
export async function checkRetainedObservability(root, check = "all") {
  if (!observabilityChecks.includes(check)) throw Object.assign(new Error("invalid observability check"), { exit_code: 2 });
  if (check === "events") {
    const manifest = await eventsOnly(root);
    return { check, full_validity: false, run_id: manifest.run_id };
  }
  const run = await validateCanonicalRun(root, "", { evidence: check === "all" || check === "evidence" });
  if (check === "all" || check === "evidence") evidenceClosure(root, run);
  let retained;
  let diagnostics;
  if (check === "all" || check === "diagnostics") {
    retained = await loadRetainedObservability(root, { canonical: run });
    diagnostics = Object.freeze({
      execution: retained.execution?.status ?? "unavailable",
      execution_omissions: retained.execution?.omitted_records ?? null,
      resources: retained.resourceIndex?.status ?? retained.resources.availability,
      resource_omissions: retained.resourceIndex?.omitted_observations ?? null,
      envelope: retained.envelope ? "available" : "unavailable",
      terminal_receipt: retained.terminalReceipt ? "available" : "unavailable",
      postgres_checkpoints: retained.postgresCheckpoints?.status ?? "unavailable",
    });
  }
  return { check, full_validity: check === "all", run_id: run.manifest.run_id, ...(retained ? { retained, diagnostics } : {}) };
}
