import { processIdentityAlive } from "../../runtime/host-admission.mjs";

// Native admission persists the nonce and scope before any producer can launch.
export function diagnosticProcessScope(runtime, dependencies) { return runtime.createDiagnosticScope(dependencies); }
// Short-lived CLI commands can finish while the admission transaction waits.
// An expired exact guard owns no remaining process; live bind failures remain
// fatal. The controller lease and pre-spawn nonce still own the daemon window.
export async function bindDiagnosticProcess(lease, proof) {
  try { await lease.bind({ boot: proof.boot, pid: proof.pid, start: proof.start }); return true; }
  catch (error) { if (processIdentityAlive(proof)) throw error; return false; }
}
export function diagnosticProcesses(runtime, scope) {
  return runtime.inspectDiagnosticScope(scope).processes;
}
export async function stopDiagnosticProcesses(runtime, scope) {
  runtime.settleDiagnosticScope(scope);
}
