// Bounded private transactions only. Go owns every runtime/proof/scan decision.
import path from "node:path";
import { spawnSync } from "../workspace/child-process.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { lifecyclePolicy } from "../contract/generated/lifecycle-policy.mjs";

const schema = "cartulary.harness_runtime_session.v2";
const rejected = () => Object.assign(new Error("native private runtime rejected"), { failure_class: "artifact", failure_reason: "artifact_error" });
function transaction(repository, runRoot, runID, operation, fields, environment) {
  const message = { schema_id: schema, request_id: 1, operation, repository, run_root: runRoot, run_id: runID, ...fields };
  validateSchemaSync(schema, message);
  const input = Buffer.from(JSON.stringify(message));
  if (input.length > lifecyclePolicy.limits.frame) throw rejected();
  const binary = environment.CARTULARY_HARNESS_RUNTIME_BIN || path.resolve(import.meta.dirname, "../../..", workspaceLayout.toolbin, "cartulary-harness");
  const result = spawnSync(binary, ["session"], { cwd: repository, env: environment, input, stdio: ["pipe", "pipe", "ignore"], maxBuffer: lifecyclePolicy.limits.frame, timeout: 30000, killSignal: "SIGKILL" });
  if (result.error || result.status !== 0) throw rejected();
  let reply;
  try { reply = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(result.stdout)); validateSchemaSync(schema, reply); }
  catch { throw rejected(); }
  if (reply.request_id !== 1 || reply.operation !== operation || reply.ok !== true) throw rejected();
  return reply.result;
}
function handle(repository, runRoot, identity, environment, owned, admissionToken = () => null) {
  const { root, lease_id: leaseID, run_id: runID } = identity;
  let closed = false;
  const call = (operation, fields = {}) => {
    if (closed) throw rejected();
    return transaction(repository, runRoot, runID, operation, { root, lease_id: leaseID, ...fields }, environment);
  };
  return Object.freeze({ root, leaseID, runID, owned,
    privatePath: (...parts) => call("private_path", { parts }),
    registerSecret: (value) => call("register", { values: [String(value ?? "")], environment: {} }),
    registerEnvironment: (values) => call("register", { values: [], environment: values }),
    record: ({ kind, target, state = "acquired", recovery_files }) => call("record", { ...(admissionToken() ? { admission_token: admissionToken() } : {}), resource: { kind, target, state, ...(recovery_files ? { recovery_files } : {}), run_id: runID, lease_id: leaseID } }),
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
  const identity = transaction(repoRoot, runRoot, runID, "borrow", { root: environment.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT, lease_id: environment.CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID }, environment);
  return handle(repoRoot, runRoot, identity, environment, false, admissionToken ?? (() => environment.CARTULARY_HOST_ADMISSION_LEASE || null));
}

// Exact recovery is admitted by the native owner only after the original owner
// dies (or by that same owner). The caller must supply the bound runtime identity.
export function recoverRuntime({ repoRoot, runRoot, root, runID, leaseID, environment = process.env }) {
  return transaction(repoRoot, runRoot, runID, "recover_close", { root, lease_id: leaseID }, environment);
}
