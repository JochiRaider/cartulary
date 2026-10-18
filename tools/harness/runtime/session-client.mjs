// Bounded private transactions only. Go owns every runtime/proof/scan decision.
import path from "node:path";
import { spawnSync } from "../workspace/child-process.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { lifecyclePolicy } from "../contract/generated/lifecycle-policy.mjs";

const schema = "cartulary.harness_runtime_session.v3";
const rejected = (diagnostic = null) => {
  const error = Object.assign(new Error("native private runtime rejected"), { failure_class: "artifact", failure_reason: "artifact_error" });
  // Deliberate private access only: generic error serialization must not retain
  // process identities or expose native proof diagnostics in public receipts.
  Object.defineProperty(error, "privateDiagnostic", { value: diagnostic });
  return error;
};
function transaction(repository, runRoot, runID, operation, fields, environment) {
  const message = { schema_id: schema, request_id: 1, operation, repository, run_root: runRoot, run_id: runID, ...fields };
  validateSchemaSync(schema, message);
  const input = Buffer.from(JSON.stringify(message));
  if (input.length > lifecyclePolicy.limits.frame) throw rejected();
  const binary = environment.CARTULARY_HARNESS_RUNTIME_BIN || path.resolve(import.meta.dirname, "../../..", workspaceLayout.toolbin, "cartulary-harness");
  const result = spawnSync(binary, ["session"], { cwd: repository, env: environment, input, stdio: ["pipe", "pipe", "ignore"], maxBuffer: lifecyclePolicy.limits.frame, timeout: 30000, killSignal: "SIGKILL" });
  const transport = { operation, reason: "transport", phase: "spawn", status: result.status, signal: result.signal, code: result.error?.code ?? null };
  if (result.error) throw rejected(transport);
  transport.phase = "decode";
  let reply;
  try { reply = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(result.stdout)); validateSchemaSync(schema, reply); }
  catch { throw rejected(transport); }
  transport.phase = "identity";
  if (reply.request_id !== 1 || reply.operation !== operation) throw rejected(transport);
  if (reply.ok === false && result.status === 11) throw rejected({ operation, ...reply.error });
  transport.phase = "outcome";
  if (result.status !== 0 || reply.ok !== true) throw rejected(transport);
  return reply.result;
}
function handle(repository, runRoot, identity, environment, owned, admissionToken = () => null, attempt = null) {
  const { root, lease_id: leaseID, run_id: runID } = identity;
  let closed = false;
  const call = (operation, fields = {}) => {
    if (closed) throw rejected();
    return transaction(repository, runRoot, runID, operation, { root, lease_id: leaseID, ...(attempt && operation !== "private_path" ? { recovery_attempt: attempt } : {}), ...fields }, environment);
  };
  return Object.freeze({ root, leaseID, runID, owned,
    privatePath: (...parts) => call("private_path", { parts }),
    producerPath: (...parts) => call(attempt ? "recovery_detail_path" : "private_path", { parts }),
    producerEnvironment: () => call("producer_environment"),
    recoveryPath: (...parts) => call("recovery_path", { parts }),
    claimRecoveryOutput: () => call("claim_recovery_output"),
    createDiagnosticScope: (depends_on) => call("create_diagnostic_scope", { depends_on }),
    inspectDiagnosticScope: (scope) => call("inspect_diagnostic_scope", { scope }),
    settleDiagnosticScope: (scope) => call("settle_diagnostic_scope", { scope }),
    registerSecret: (value) => call("register", { values: [String(value ?? "")], environment: {} }),
    registerEnvironment: (values) => call("register", { values: [], environment: values }),
    record: ({ kind, target, state = "acquired", recovery_files, depends_on }) => call("record", { ...(admissionToken() ? { admission_token: admissionToken() } : {}), resource: { kind, target, state, ...(depends_on ? { depends_on } : {}), ...(recovery_files ? { recovery_files } : {}), run_id: runID, lease_id: leaseID } }),
    close: () => { if (closed) return; if (owned) call("close"); closed = true; },
  });
}
export function createRuntime({ repoRoot, runRoot, runID, scratchRoot = "", environment = process.env }) {
  const identity = transaction(repoRoot, runRoot, runID, "create", { scratch: scratchRoot }, environment);
  const runtime = handle(repoRoot, runRoot, identity, environment, true);
  try { runtime.registerEnvironment(environment); } catch (error) { runtime.close(); throw error; }
  return runtime;
}
export function borrowRuntime({ repoRoot, runRoot, environment = process.env, admissionToken }) {
  const runID = environment.CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID;
  const attempt=environment.CARTULARY_HARNESS_RECOVERY_ATTEMPT || null;
  const identity = transaction(repoRoot, runRoot, runID, "borrow", { root: environment.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT, lease_id: environment.CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID, ...(attempt ? {recovery_attempt:attempt} : {}) }, environment);
  return handle(repoRoot, runRoot, identity, environment, false, admissionToken ?? (() => environment.CARTULARY_HOST_ADMISSION_LEASE || null), attempt);
}

// Explicit recovery admission. Ordinary borrowed handles cannot mint attempts,
// select output roots, or reopen ordinary detail allocation.
export function openRecoveryAttempt({repoRoot,runRoot,root,runID,leaseID,environment=process.env}) {
  const identity={root,lease_id:leaseID};
  const id=transaction(repoRoot,runRoot,runID,"begin_recovery_attempt",identity,environment);
  const call=(operation,fields={})=>transaction(repoRoot,runRoot,runID,operation,{...identity,recovery_attempt:id,...fields},environment);
  return Object.freeze({id,detailPath:(...parts)=>call("recovery_detail_path",{parts}),allocateOutput:()=>call("allocate_recovery_output"),environment:()=>call("producer_environment"),finish:()=>call("finish_recovery_attempt")});
}

// Exact recovery is admitted by the native owner only after the original owner
// dies (or by that same owner). The caller must supply the bound runtime identity.
export function recoverRuntime({ repoRoot, runRoot, root, runID, leaseID, environment = process.env }) {
  return transaction(repoRoot, runRoot, runID, "recover_close", { root, lease_id: leaseID }, environment);
}
