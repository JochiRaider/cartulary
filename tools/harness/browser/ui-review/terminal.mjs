import path from "node:path";
import { scanRetainedRoot } from "../../runtime/suite-runtime.mjs";
import { commandID, failureRecord, ReviewFailure, schemaID } from "./contract.mjs";
import { publishJSON, unregisterSession } from "./session-files.mjs";

// Ordinary shutdown and recovery publish the same terminal evidence. Cleanup
// proof remains registered until both publication and the retained scan succeed.
export async function finishTerminal({ record, locator, runRoot, counts, failures, exitCode, duration, cleanupFailed, forbiddenValues = [] }) {
  failures = [...new Map(failures.map((failure) => [JSON.stringify(failure), failure])).values()];
  let state = failures.length ? "failed" : "closed", receipt = null;
  const artifactFailure = () => {
    state = "failed"; receipt = null; exitCode ||= 11;
    if (!failures.some((failure) => failure.diagnostic_code === "unsafe_artifact")) failures.push(failureRecord(new ReviewFailure("unsafe_artifact")));
  };
  const scan = async () => {
    if ((await scanRetainedRoot(runRoot, { forbiddenValues, removeUnsafe: true })).status !== "pass") throw new ReviewFailure("unsafe_artifact");
  };
  const publishLocator = () => publishJSON(record.locator, "session", { ...locator, state, updated_at: new Date().toISOString(), terminal_receipt: receipt }, { replace: true });
  try {
    await scan();
    const value = { schema_id: schemaID("receipt"), command_id: commandID("ui-review"), session_id: record.session_id, operation_id: null, mode: locator.mode, state, status: failures.length ? "error" : "ok", exit_code: exitCode, started_at: locator.created_at, finished_at: new Date().toISOString(), duration_ms: duration, failures, counts, bundle_id: null, cleanup: cleanupFailed ? "failed" : "complete" };
    receipt = { path: "ui-review/terminal.json", ...publishJSON(path.join(runRoot, "ui-review/terminal.json"), "receipt", value) };
    publishLocator();
    await scan();
    if (!cleanupFailed) unregisterSession(record);
  } catch {
    artifactFailure();
    try { publishLocator(); } catch { /* Keep recovery proof for exact repair. */ }
  }
  return { state, receipt, failures, exitCode };
}
