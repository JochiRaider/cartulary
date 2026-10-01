import { randomBytes } from "node:crypto";
import { readFileSync, readdirSync, statSync } from "node:fs";
import { processIdentity, processIdentityAlive } from "../../runtime/host-admission.mjs";
import { ownedProcess, stopOwnedProcess, validateDiagnosticScope } from "./ownership.mjs";
import { ReviewFailure } from "./failure.mjs";

// Persisted before spawn: the upstream CLI forks a detached daemon before it
// reports its PID. An operation-specific inherited nonce closes that acquisition
// window, including controller death. No global CLI registry is consulted.
export function diagnosticProcessScope() {
  const { boot, start } = processIdentity();
  return { boot, start, token: randomBytes(32).toString("hex"), uid: process.getuid() };
}
// Short-lived CLI commands can finish while the admission transaction waits.
// An expired exact guard owns no remaining process; live bind failures remain
// fatal. The controller lease and pre-spawn nonce still own the daemon window.
export async function bindDiagnosticProcess(lease, proof) {
  try { await lease.bind({ boot: proof.boot, pid: proof.pid, start: proof.start }); return true; }
  catch (error) { if (processIdentityAlive(proof)) throw error; return false; }
}
export function diagnosticProcesses(scope) {
  validateDiagnosticScope(scope);
  if (scope.boot !== processIdentity().boot) return [];
  const found = [];
  for (const pid of readdirSync("/proc")) {
    if (!/^\d+$/u.test(pid)) continue;
    try {
      if (statSync(`/proc/${pid}`).uid !== scope.uid) continue;
      const identity = ownedProcess(Number(pid));
      if (BigInt(identity.start) < BigInt(scope.start)) continue;
      const environment = readFileSync(`/proc/${pid}/environ`).toString().split("\0");
      if (environment.includes(`CARTULARY_DIAGNOSTIC_TOKEN=${scope.token}`)) found.push(identity);
    } catch (error) { if (!["ENOENT", "ESRCH", "EACCES"].includes(error.code)) throw error; }
  }
  return found;
}
export async function stopDiagnosticProcesses(scope) {
  const deadline = performance.now() + 10000;
  while (performance.now() < deadline) {
    const processes = diagnosticProcesses(scope);
    if (!processes.length) return;
    const results = await Promise.allSettled(processes.map((proof) => stopOwnedProcess(proof, { graceMS: 200, killMS: 1000 })));
    if (results.some((result) => result.status === "rejected")) throw new ReviewFailure("cleanup_failed");
    await new Promise((resolve) => setTimeout(resolve, 20));
  }
  throw new ReviewFailure("cleanup_failed");
}
