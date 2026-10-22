import path from "node:path";
import { createHash } from "node:crypto";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";

export const receiptDigest = (value) => `sha256:${createHash("sha256").update(value).digest("hex")}`;

// This reader is installed with the format cutover, before any wrapper can
// publish a receipt. The receipt is structural evidence, never scan authority.
export function readTerminalReceipt(runRoot, run, envelope, envelopeDigest) {
  let receipt, receiptBytes;
  try {
    receiptBytes = readLocalFile(path.join(runRoot, "diagnostics/terminal-receipt.json"), { maximum: 65536 });
    receipt = parseStrictJSON(receiptBytes.toString("utf8"));
  } catch (error) { if (error.code === "ENOENT") return null; throw error; }
  validateSchemaSync("cartulary.harness_terminal_receipt.v1", receipt);
  if (!envelope?.invocation_digest || receipt.invocation_digest !== envelope.invocation_digest || receipt.source_digest !== run.manifest.source_digest) throw new Error("receipt invocation mismatch");
  if (receipt.publication_start_ms < receipt.outer_duration_ms) throw new Error("receipt precedes child completion");
  const seen = new Set();
  for (const member of receipt.graph_memberships) {
    if (seen.has(member.run_digest) || member.scan_interval.end_ms < member.scan_interval.start_ms || member.scan_interval.end_ms > receipt.outer_duration_ms) throw new Error("invalid receipt memberships");
    seen.add(member.run_digest);
  }
  const member = receipt.graph_memberships.find((entry) => entry.run_digest === receiptDigest(run.manifest.run_id));
  if (!member || member.graph_digest !== run.manifest.graph_digest || member.envelope_digest !== envelopeDigest) throw new Error("receipt envelope mismatch");
  const scanBytes = readLocalFile(path.join(runRoot, "retained-secret-scan.json"), { maximum: 65536 });
  const scan = parseStrictJSON(scanBytes.toString("utf8"));
  validateSchemaSync("cartulary.harness_retained_secret_scan.v1", scan);
  if (member.scan_digest !== receiptDigest(scanBytes) || member.scan_outcome !== "passed") throw new Error("receipt scan mismatch");
  if (receipt.outcome === "passed" && (receipt.child_exit_code !== 0 || run.summary.status !== "pass" || receipt.graph_memberships.some((entry) => entry.scan_outcome !== "passed"))) throw new Error("receipt success contradicts retained evidence");
  return { value: receipt, digest: receiptDigest(receiptBytes) };
}
