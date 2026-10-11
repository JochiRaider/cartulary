// Private transport only. Native fixtures own acquisition, sharing, reuse,
// leases, dependency cleanup and recovery for graph and UI-review callers.
import path from "node:path";
import { spawn } from "../workspace/child-process.mjs";
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { parseStrictJSON, validateSchemaSync, primaryPublicFailure } from "../contract/index.mjs";
import { processIdentity } from "./host-admission.mjs";
import { CommandFailure } from "./command-failure.mjs";
import { ownedProcess, stopOwnedProcess } from "./owned-process.mjs";

const schemaID = "cartulary.harness_fixture_session.v1", maximum = 1048576;
const invalid = (cause) => new CommandFailure("fixture session rejected", { failure_class: "artifact", failure_reason: "artifact_error" }, {cause});
function failure(records) {
  const primary = primaryPublicFailure(records.map((record,index) => ({ ...record, lifecycle_step:"semantic_target_behavior", scheduler_event_sequence:index+1 })));
  const error = new CommandFailure("native fixture operation failed", primary ?? { failure_class:"harness", failure_reason:"fixture_error" });
  error.artifact_refs = [...new Set(records.flatMap((record) => record.artifact_refs))];
  error.cleanupFailures = records.slice(1).map((record) => new CommandFailure("native fixture cleanup failed",record));
  return error;
}
export function createFixtureSession({ root, runRoot, target, runtime, environment, signal, onChildProcess,
  cleanupResults, observeAllocation = () => {}, observeLease = () => {}, onContainerProof = () => {} }) {
  let child, proof, transport, ended, closePromise, allocationClose, sequence = 0, buffer = Buffer.alloc(0), unresolved = false, releasing = () => {}, terminal;
  const pending = new Map();
  const reject = (error) => { error = error instanceof CommandFailure ? error : invalid(error); unresolved = true; terminal ??= error; for (const {reject} of pending.values()) reject(error); pending.clear(); };
  const start = () => {
    if (child) return;
    signal?.throwIfAborted();
    const binary = environment.CARTULARY_HARNESS_RUNTIME_BIN || path.join(root,workspaceLayout.toolbin,"cartulary-harness");
    child = spawn(binary,["fixtures"],{cwd:root,detached:true,env:{ ...environment,
      CARTULARY_TEST_TARGET:target,CARTULARY_TEST_RESULTS_DIR:path.dirname(runRoot),CARTULARY_TEST_RUN_ID:runtime.runID,
      CARTULARY_HARNESS_SUITE_RUNTIME_ROOT:runtime.root,CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID:runtime.leaseID,
      CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID:runtime.runID,CARTULARY_HARNESS_FIXTURE_OWNER:JSON.stringify(processIdentity()) },stdio:["ignore","ignore","ignore","pipe"]});
    transport = child.stdio[3];
    ended = new Promise((resolve) => {
      child.once("error",(error) => reject(error.code==="ENOENT"?new CommandFailure("fixture runtime unavailable",{failure_class:"config",failure_reason:"configuration_error"},{cause:error}):error));
      child.once("close",(code) => { if (pending.size || (code !== 0 && !terminal)) reject(invalid()); resolve(code); });
    });
    if (child.pid) { proof = ownedProcess(child.pid); releasing = onChildProcess?.(child.pid) ?? (() => {}); }
    transport.on("error",reject);
    transport.on("data",(bytes) => {
      try {
        buffer = Buffer.concat([buffer,bytes]);
        while (true) {
          const newline = buffer.indexOf(10);
          if (newline < 0) { if (buffer.length >= maximum) throw invalid(); break; }
          if (newline + 1 > maximum) throw invalid();
          const message = parseStrictJSON(buffer.subarray(0,newline).toString()); buffer = buffer.subarray(newline+1);
          validateSchemaSync(schemaID,message);
          const request = pending.get(message.request_id);
          if (!request || message.kind !== "response" || message.run_id !== runtime.runID || message.runtime_lease_id !== runtime.leaseID || message.operation !== request.operation || !message.result || message.ok !== (message.failures.length === 0)) throw invalid();
          pending.delete(message.request_id);
          const facts = message.result; unresolved = facts.unresolved;
          for (const record of facts.cleanup) cleanupResults?.record(record.operation,{unitID:record.unit_id || null,leaseID:record.lease_id || null,
            artifactRefs:record.published?[`_shared/fixture-leases/${record.lease_id}.json`]:[],error:record.failures.length?failure(record.failures):undefined});
          for (const secret of facts.secrets) runtime.registerSecret(secret);
          // Diagnostic callbacks cannot change physical fixture ownership.
          if (facts.observation_available) for (const record of facts.observations) {
            try { if (record.kind === "lease") observeLease(record); else observeAllocation(record); } catch { /* optional */ }
          }
          try { onContainerProof(facts.container_proof); } catch { /* optional */ }
          if (message.ok) request.resolve(facts); else { const error = failure(message.failures); terminal ??= error; request.reject(error); }
        }
      } catch (error) { reject(error); transport.destroy(); }
    });
    transport.on("end",() => { if (buffer.length || pending.size) reject(invalid()); });
  };
  const request = (operation,values = {}) => {
    start();
    if (sequence >= 65536 || pending.size >= 64) return Promise.reject(invalid());
    const message = {schema_id:schemaID,kind:"request",request_id:++sequence,run_id:runtime.runID,runtime_lease_id:runtime.leaseID,
      operation,request:null,lease_id:null,healthy:false,retain:false,close_scope:null,...values};
    validateSchemaSync(schemaID,message);
    const bytes = Buffer.from(`${JSON.stringify(message)}\n`); if (bytes.length > maximum) return Promise.reject(invalid());
    return new Promise((resolve,reject) => { pending.set(sequence,{operation,resolve,reject}); transport.write(bytes,(error) => { if (error) { pending.delete(message.request_id); reject(invalid(error)); } }); });
  };
  const close = () => closePromise ??= (async () => {
    signal?.removeEventListener("abort",abort);
    if (!child) return;
    let error, timer;
    try {
      await Promise.race([request("close",{close_scope:"session"}),new Promise((_,reject) => { timer = setTimeout(() => reject(new CommandFailure("fixture drain deadline expired",{failure_class:"timing",failure_reason:"timeout_failure"})),125000); })]);
      transport.end(); await ended;
    } catch (cause) { error = cause; unresolved = true; }
    finally {
      clearTimeout(timer); transport.destroy();
      for (const cleanup of [() => proof && stopOwnedProcess(proof),releasing]) try { await cleanup(); } catch (cause) { error ??= cause; unresolved = true; }
    }
    if (error) throw error;
  })();
  const abort = () => { close().catch(() => {}); };
  signal?.addEventListener("abort",abort,{once:true});
  return {
    async acquire(capability,options = {}) {
      if (closePromise || allocationClose) throw invalid();
      const names = {unitID:"unit_id",affinityKey:"affinity_key",digest:"digest",browserStage:"browser_stage",runtimeProfileID:"runtime_profile_id",fixtureProfileID:"fixture_profile_id",snapshotKey:"snapshot_key",builderUnitID:"builder_unit_id",rowID:"row_id",predicateID:"predicate_id"};
      const input = {capability}; for (const [key,name] of Object.entries(names)) if (options[key] !== undefined && options[key] !== null) input[name] = options[key];
      const facts = await request("acquire",{request:input});
      const lease = {record:facts.lease,resource:{environment:facts.environment},entry:{allocationRef:facts.allocation_ref},
        async release({healthy = true,retainWarm = false} = {}) { const result = await request("release",{lease_id:facts.lease.lease_id,healthy,retain:retainWarm}); lease.record=result.lease; return {retained:result.retained}; }};
      return lease;
    },
    recover: () => request("recover"),
    cleanupResults,
    close: () => allocationClose ??= (async () => {
      if (!child) return;
      try { return await request("close",{close_scope:"allocations"}); }
      catch(error) { try{await close();}catch(secondary){(error.cleanupFailures??=[]).push(secondary);}throw error; }
    })(),
    closeSession: close,
    hasUnresolvedCleanup: () => unresolved,
  };
}
