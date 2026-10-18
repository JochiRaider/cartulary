import { readFileSync } from "node:fs";
import path from "node:path";
import { spawn } from "../workspace/child-process.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { processIdentity } from "./host-admission.mjs";
import { parseCleanupDiagnostic, withCleanupDiagnostic } from "./cleanup-diagnostic.mjs";

export function ownedProcess(pid) {
  const proof = processIdentity(pid);
  const fields = readFileSync(`/proc/${pid}/stat`, "utf8").split(") ").at(-1).split(" ");
  return { ...proof, group: Number(fields[2]) === pid, pgid: Number(fields[2]), sid: Number(fields[3]) };
}

// The native physical owner alone verifies identities, signals and confirms
// settlement. An exit notification is not proof that descendants have drained.
export async function stopOwnedProcess(proof, { graceMS = 1000, killMS = 5000, environment = process.env } = {}) {
  const root = path.resolve(import.meta.dirname, "../../..");
  const binary = environment.CARTULARY_HARNESS_RUNTIME_BIN || path.join(root, workspaceLayout.toolbin, "cartulary-harness");
  const child = spawn(binary, ["process"], { cwd: root, stdio: ["pipe", "ignore", "pipe"], env: environment });
  const chunks = [];
  let bytes = 0;
  child.stderr.on("data", (data) => { bytes += data.length; if (bytes <= 4096) chunks.push(data); });
  const completion = new Promise((resolve, reject) => {
    child.once("error", (error) => reject(withCleanupDiagnostic(error, "process_protocol")));
    child.once("close", (code) => {
      if (code === 0) { resolve(); return; }
      const error = new Error("owned process cleanup failed");
      const diagnostic = bytes <= 4096 ? parseCleanupDiagnostic(Buffer.concat(chunks)) : null;
      if (diagnostic) Object.defineProperty(error, "cleanupDiagnostic", { value: diagnostic, configurable: true });
      reject(withCleanupDiagnostic(error, "process_protocol"));
    });
  });
  child.stdin.on("error", () => {});
  child.stdin.end(JSON.stringify({ proof, grace_ms: graceMS, kill_ms: killMS }));
  await completion;
}
