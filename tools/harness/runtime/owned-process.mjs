import { readFileSync } from "node:fs";
import path from "node:path";
import { spawn } from "../workspace/child-process.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { processIdentity } from "./host-admission.mjs";
import { parseCleanupDiagnostic, withCleanupDiagnostic } from "./cleanup-diagnostic.mjs";
import { CommandFailure } from "./command-failure.mjs";

export function ownedProcess(pid) {
  const proof = processIdentity(pid);
  const fields = readFileSync(`/proc/${pid}/stat`, "utf8").split(") ").at(-1).split(" ");
  return { ...proof, group: Number(fields[2]) === pid, pgid: Number(fields[2]), sid: Number(fields[3]) };
}

// The native physical owner alone verifies identities, signals and confirms
// settlement. An exit notification is not proof that descendants have drained.
export async function stopOwnedProcess(proof, { graceMS = 1000, killMS = 5000, environment = process.env } = {}) {
  await processOperation(proof, { graceMS, killMS, environment });
}

export async function inspectOwnedProcess(proof, { environment = process.env } = {}) {
  return processOperation(proof, { graceMS: 0, killMS: 1, environment, inspect: true });
}

export async function drainOwnedProcess(proof, { timeoutMS, environment = process.env } = {}) {
  if (!Number.isSafeInteger(timeoutMS) || timeoutMS < 1 || timeoutMS > 86_400_000) throw new Error("invalid process drain deadline");
  await processOperation(proof, { timeoutMS, environment, drain: true });
}

async function processOperation(proof, { graceMS, killMS, timeoutMS, environment, inspect = false, drain = false }) {
  const root = path.resolve(import.meta.dirname, "../../..");
  const binary = environment.CARTULARY_HARNESS_RUNTIME_BIN || path.join(root, workspaceLayout.toolbin, "cartulary-harness");
  const child = spawn(binary, inspect ? ["process", "inspect"] : drain ? ["process", "drain"] : ["process"], { cwd: root, stdio: ["pipe", "ignore", "pipe"], env: environment });
  const chunks = [];
  let bytes = 0;
  child.stderr.on("data", (data) => { bytes += data.length; if (bytes <= 4096) chunks.push(data); });
  const completion = new Promise((resolve, reject) => {
    child.once("error", (error) => reject(withCleanupDiagnostic(error, "process_protocol")));
    child.once("close", (code) => {
      if (code === 0 || inspect && code === 1) { resolve(code === 0); return; }
      const error = drain && code === 13
        ? new CommandFailure("owned process drain deadline exceeded", { failure_class: "timing", failure_reason: "timeout_failure" })
        : new Error(inspect ? "owned process inspection failed" : "owned process cleanup failed");
      const diagnostic = bytes <= 4096 ? parseCleanupDiagnostic(Buffer.concat(chunks)) : null;
      if (diagnostic) Object.defineProperty(error, "cleanupDiagnostic", { value: diagnostic, configurable: true });
      reject(withCleanupDiagnostic(error, "process_protocol"));
    });
  });
  child.stdin.on("error", () => {});
  child.stdin.end(JSON.stringify(drain ? { proof, timeout_ms: timeoutMS } : { proof, grace_ms: graceMS, kill_ms: killMS }));
  return completion;
}
