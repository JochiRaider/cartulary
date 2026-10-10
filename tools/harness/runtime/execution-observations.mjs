import { createHash, randomUUID } from "node:crypto";
import { readdirSync } from "node:fs";
import path from "node:path";
import { borrowSuiteRuntime } from "./suite-runtime.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile } from "./secure-local-files.mjs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";

export const executionEnabledKey = "CARTULARY_HARNESS_EXECUTION_ENABLED";
const recordSchema = "cartulary.harness_execution_record.v1";
const digest = (records) => `sha256:${createHash("sha256").update(JSON.stringify(records)).digest("hex")}`;

function channel(repoRoot, environment) {
  const runRoot = path.resolve(repoRoot, environment.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results", environment.CARTULARY_TEST_RUN_ID);
  const runtime = borrowSuiteRuntime({ repoRoot, runRoot, environment });
  return runtime.privatePath("execution-observations");
}

// Each reserved slot has one producer. Fixed slot names bound concurrent writers
// without a shared append stream, daemon, lock recovery or unbounded directory.
export function executionObservation({ repoRoot, environment }, initial) {
  let file, root, limits, processProof;
  const open = () => readLocalFile(path.join(root, "open"), { maximum: 8 }).toString("utf8") === "1\n";
  const write = (record, proof) => {
    if (!file) return;
    try {
      if (!open()) return;
      validateSchemaSync(recordSchema, record);
      processProof ??= proof;
      const observation = processProof ? { record, process_proof: processProof } : record;
      if (processProof) validateSchemaSync("cartulary.harness_private_launch_observation.v1", observation);
      const value = `${JSON.stringify(observation)}\n`;
      if (Buffer.byteLength(value) > limits.maximum_record_bytes) throw new Error("execution record exceeds bound");
      atomicLocalFile(file, value, { replace: true, maximumReplacementBytes: limits.maximum_record_bytes });
    } catch { /* Optional loss leaves the previously published incomplete fact. */ }
  };
  try {
    if (environment[executionEnabledKey] !== "1") return { update() {} };
    root = channel(repoRoot, environment);
    limits = parseStrictJSON(readLocalFile(path.join(root, "limits.json")).toString("utf8"));
    if (!Number.isInteger(limits.maximum_execution_records) || limits.maximum_execution_records < 1 || limits.maximum_execution_records > 4096 || limits.maximum_record_bytes !== 65536) throw new Error("invalid execution limits");
    if (!open()) return { update() {} };
    const start = Number.parseInt(randomUUID().slice(0, 8), 16) % limits.maximum_execution_records;
    for (let i = 0; i < limits.maximum_execution_records; i++) {
      const slot = path.join(root, `slot-${(start + i) % limits.maximum_execution_records}`);
      try { validateSchemaSync(recordSchema, initial); atomicLocalFile(slot, `${JSON.stringify(initial)}\n`); file = slot; break; }
      catch (error) { if (error.code !== "EEXIST") throw error; }
    }
    if (!file) { atomicLocalFile(path.join(root, "truncated"), "1\n"); return { update() {} }; }
    write(initial);
  } catch { /* Observation is independent of execution and required channels. */ }
  return { update: write, token: file ? { slot: path.basename(file), ref: initial.ref } : null };
}

// A logical producer may span helper processes. Its private token names only its
// reserved slot, and the retained identity must still match before replacement.
export function resumeExecutionObservation({ repoRoot, environment }, token) {
  return { update(record) {
    try {
      if (environment[executionEnabledKey] !== "1" || !/^slot-[0-9]{1,4}$/u.test(token?.slot) || token.ref !== record.ref) return;
      const root = channel(repoRoot, environment), file = path.join(root, token.slot);
      if (readLocalFile(path.join(root, "open")).toString("utf8") !== "1\n") return;
      const previous = parseStrictJSON(readLocalFile(file).toString("utf8"));
      if (previous.ref !== token.ref || previous.outcome !== "incomplete") return;
      validateSchemaSync(recordSchema, record);
      atomicLocalFile(file, `${JSON.stringify(record)}\n`, { replace: true, maximumReplacementBytes: 65536 });
    } catch { /* Preserve the incomplete beginning on optional publication loss. */ }
  } };
}

export function createExecutionSession({ runtime, manifest, policy, runRoot, onProcessProof = () => {} }) {
  const records = [], omittedRefs = [];
  let captured = false, published = false, partial = false, truncated = false, omissions = 0, recordBytes = 0;
  const root = runtime.privatePath("execution-observations");
  const enabled = manifest.instrumentation.mode === "basic";
  if (enabled) {
    try {
      privateDirectory(root);
      atomicLocalFile(path.join(root, "limits.json"), JSON.stringify(policy));
      atomicLocalFile(path.join(root, "open"), "1\n");
    } catch { partial = true; }
  }
  function record(value) {
    if (!enabled || published) return;
    try {
      validateSchemaSync(recordSchema, value);
      const previous = records.findIndex((entry) => entry.ref === value.ref);
      if (previous !== -1) {
        const old = records[previous];
        if (JSON.stringify(old) === JSON.stringify(value)) return;
        if (old.kind !== "allocation" || value.kind !== "allocation" || old.allocation_ref !== value.allocation_ref || old.unit_id !== value.unit_id || old.capability !== value.capability || old.outcome !== "incomplete") throw new Error("conflicting execution observation");
        recordBytes -= Buffer.byteLength(JSON.stringify(old));
        records.splice(previous, 1);
      }
      const bytes = Buffer.byteLength(JSON.stringify(value));
      if (records.length >= policy.maximum_execution_records || recordBytes + bytes > policy.maximum_metadata_bytes - 512 * 1024) {
        truncated = true; omissions += 1;
        if (omittedRefs.length < policy.maximum_execution_records) omittedRefs.push(value.ref);
        return;
      }
      records.push(value);
      recordBytes += bytes;
    } catch { partial = true; omissions += 1; }
  }
  return {
    environment: enabled ? { [executionEnabledKey]: "1" } : {},
    record,
    capture() {
      if (!enabled || captured) return;
      captured = true;
      try {
        atomicLocalFile(path.join(root, "open"), "0\n", { replace: true });
        const entries = readdirSync(root).sort();
        truncated ||= entries.includes("truncated");
        for (const entry of entries) {
          if (!/^slot-[0-9]+$/u.test(entry)) continue;
          try {
            const value = parseStrictJSON(readLocalFile(path.join(root, entry), { maximum: policy.maximum_record_bytes }).toString("utf8"));
            if (value.process_proof) {
              validateSchemaSync("cartulary.harness_private_launch_observation.v1", value);
              if (value.record.kind !== "launch" || !value.record.launched || value.process_proof.start !== value.process_proof.stat.start) throw new Error("invalid private launch proof");
              record(value.record);
              onProcessProof(value.process_proof, value.record);
            } else record(value);
          }
          catch { partial = true; omissions += 1; }
        }
      } catch { partial = true; }
    },
    publish() {
      if (!enabled || published) return;
      this.capture();
      published = true;
      try {
        records.sort((a, b) => a.ref.localeCompare(b.ref));
        let bytes = 0;
        const retained = [];
        for (const value of records) {
          const size = Buffer.byteLength(JSON.stringify(value)) + 1;
          // Reserve bounded space for identities of records omitted by the byte cap.
          if (bytes + size > policy.maximum_metadata_bytes - 512 * 1024) { omittedRefs.push(value.ref); omissions += 1; truncated = true; }
          else { retained.push(value); bytes += size; }
        }
        // Corrupt/lost/capped private beginnings must not make retained children
        // unresolvable. Declare unavailable references, never fabricate records.
        const refs = new Set(retained.map((value) => value.ref));
        const missing = new Set(omittedRefs.filter((ref) => !refs.has(ref)));
        for (let i = retained.length - 1; i >= 0; i--) {
          const value = retained[i];
          const dependencies = [value.parent_invocation_id && `launch:${value.parent_invocation_id}`,
            value.kind !== "launch" && value.invocation_id && `launch:${value.invocation_id}`,
            value.kind !== "allocation" && value.allocation_ref].filter(Boolean);
          for (const ref of dependencies) if (!refs.has(ref)) missing.add(ref);
        }
        if (missing.size) { partial = true; omissions = Math.max(omissions, missing.size); }
        if (missing.size > policy.maximum_execution_records) throw new Error("execution reference closure exceeds bound");
        const index = { schema_id: "cartulary.harness_execution_index.v1", run_id: manifest.run_id,
          source_digest: manifest.source_digest, graph_digest: manifest.graph_digest, policy_digest: manifest.instrumentation.policy_digest,
          status: truncated ? "truncated" : partial || retained.some((record) => record.outcome === "incomplete") ? "partial" : "complete", omitted_records: truncated ? null : omissions,
          omitted_refs: [...missing].sort(), records: retained, records_digest: digest(retained) };
        validateSchemaSync(index.schema_id, index);
        const value = `${JSON.stringify(index)}\n`;
        if (Buffer.byteLength(value) > policy.maximum_metadata_bytes) throw new Error("execution index exceeds bound");
        atomicLocalFile(path.join(runRoot, "diagnostics/execution-index.json"), value);
      } catch { /* Absent terminal publication is explicitly unavailable. */ }
    },
  };
}

export { digest as executionRecordsDigest };
