import path from "node:path";
import { readdirSync } from "node:fs";
import { borrowSuiteRuntime, closeSuiteRuntimeRoot } from "../../runtime/suite-runtime.mjs";
import { processIdentityAlive } from "../../runtime/host-admission.mjs";
import { readLocalFile } from "../../runtime/secure-local-files.mjs";
import { recoverReviewPreparation } from "../review-preparation.mjs";
import { commandID, emptyCounts, failureRecord, result, ReviewFailure, schemaID, validate } from "./contract.mjs";
import { publishJSON, readLocator, unregisterSession } from "./session-files.mjs";
import { repoRoot } from "./toolchain.mjs";

export async function recoverSession(record, input) {
  if (processIdentityAlive(record.process)) throw new ReviewFailure("session_mismatch");
  const identity = readLocator(record.locator);
  if (identity.locator.session_id !== record.session_id || identity.locator.mode !== input.UI_MODE) throw new ReviewFailure("session_mismatch");
  const runtime = borrowSuiteRuntime({ repoRoot, runRoot: identity.runRoot, environment: { CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: record.runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: record.runtime.lease_id, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: record.runtime.run_id } });
  const failures = [failureRecord(new ReviewFailure("session_lost"))]; let cleaned = true;
  try { if (input.UI_MODE === "seeded") await recoverReviewPreparation({ runtime, profile: input.REVIEW_PROFILE }); }
  catch { failures.push(failureRecord(new ReviewFailure("cleanup_failed"))); cleaned = false; }
  if (cleaned) {
    try { closeSuiteRuntimeRoot({ root: record.runtime.root, expectedLeaseID: record.runtime.lease_id }); }
    catch { failures.push(failureRecord(new ReviewFailure("cleanup_failed"))); cleaned = false; }
  }
  const counts = emptyCounts();
  try {
    const directory = path.join(identity.runRoot, "ui-review/operations");
    for (const operation of readdirSync(directory)) {
      if (!/^[1-9][0-9]*$/u.test(operation)) throw new ReviewFailure("invalid_artifact");
      const receipt = validate("receipt", JSON.parse(readLocalFile(path.join(directory, operation, "receipt.json"))));
      if (receipt.session_id !== record.session_id) throw new ReviewFailure("invalid_artifact");
      if (receipt.status === "ok") for (const key of Object.keys(counts)) counts[key] += receipt.counts[key];
    }
  } catch (error) { if (error.code !== "ENOENT") failures.push(failureRecord(new ReviewFailure("invalid_artifact"))); }
  const receipt = { schema_id: schemaID("receipt"), command_id: commandID("ui-review"), session_id: record.session_id, operation_id: null, mode: input.UI_MODE, state: "failed", status: "error", exit_code: 3, started_at: identity.locator.created_at, finished_at: new Date().toISOString(), duration_ms: Math.floor(Number(process.hrtime.bigint() - BigInt(record.started_tick_ns)) / 1e6), failures, counts, bundle_id: null, cleanup: cleaned ? "complete" : "failed" };
  let reference = null;
  try { reference = { path: "ui-review/terminal.json", ...publishJSON(path.join(identity.runRoot, "ui-review/terminal.json"), "receipt", receipt) }; }
  catch { failures.push(failureRecord(new ReviewFailure("unsafe_artifact"))); }
  publishJSON(identity.file, "session", { ...identity.locator, state: "failed", updated_at: new Date().toISOString(), terminal_receipt: reference }, { replace: true });
  if (cleaned) unregisterSession(record);
  return result("ui-review", { session_id: record.session_id, state: "failed", status: "error", exit_code: 3, failures, receipt: reference });
}
