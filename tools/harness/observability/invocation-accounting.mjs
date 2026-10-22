import path from "node:path";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { receiptDigest } from "./terminal-receipt.mjs";

export function readInvocationAccounting(runRoot, run, receipt, terminalDigest) {
  let value;
  try { value = parseStrictJSON(readLocalFile(path.join(runRoot, "diagnostics/invocation-accounting.json"), { maximum: 65536 }).toString("utf8")); }
  catch (error) {
    if (error.code === "ENOENT" && run.manifest.declared_inputs?.HARNESS_ACCOUNTING !== "required") return null;
    throw error;
  }
  validateSchemaSync("cartulary.harness_invocation_accounting.v1", value);
  const cleanup = parseStrictJSON(readLocalFile(path.join(runRoot, "cleanup-results.json"), { maximum: 16 * 1024 ** 2 }).toString("utf8"));
  validateSchemaSync("cartulary.harness_cleanup_results.v2", cleanup);
  if (cleanup.run_id !== run.manifest.run_id || cleanup.results.some((result) => result.outcome !== "completed")) throw new Error("accounting cleanup did not settle");
  if (value.metrics.memory_peak_bytes < value.metrics.initial_memory_peak_bytes) throw new Error("accounting memory peak decreased");
  if (!receipt || value.invocation_digest !== receipt.invocation_digest || value.source_digest !== run.manifest.source_digest || value.metrics.child_exit_code !== receipt.child_exit_code) throw new Error("accounting invocation mismatch");
  if (value.metrics.duration_ms + 1 < run.summary.wall_duration_ms || value.metrics.duration_ms > receipt.outer_duration_ms + 10) throw new Error("accounting duration contradicts invocation boundaries");
  if (value.metrics.signal !== 0 && receipt.outcome !== "cancelled") throw new Error("accounting signal contradicts receipt");
  const seen = new Set();
  for (const member of value.graph_memberships) {
    if (seen.has(member.run_digest)) throw new Error("duplicate accounting membership");
    seen.add(member.run_digest);
    if (!receipt.graph_memberships.some((other) => other.run_digest === member.run_digest && other.graph_digest === member.graph_digest)) throw new Error("foreign accounting membership");
  }
  if (value.graph_memberships.length !== receipt.graph_memberships.length) throw new Error("incomplete accounting membership");
  const own = value.graph_memberships.find((member) => member.run_digest === receiptDigest(run.manifest.run_id));
  if (!own || own.graph_digest !== run.manifest.graph_digest || own.receipt_digest !== terminalDigest) throw new Error("accounting receipt mismatch");
  return value;
}
