import { createHash, randomUUID } from "node:crypto";
import path from "node:path";
import { borrowSuiteRuntime } from "./suite-runtime.mjs";
import { atomicLocalFile, readLocalFile } from "./secure-local-files.mjs";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";

export const executionEnabledKey = "CARTULARY_HARNESS_EXECUTION_ENABLED";
const recordSchema = "cartulary.harness_execution_record.v2";
const digest = (records) => `sha256:${createHash("sha256").update(JSON.stringify(records)).digest("hex")}`;

function encodeObservation(record, maximumBytes, processProof) {
  const observation = processProof ? { record, process_proof: processProof } : record;
  // The private envelope validates the complete record reference as well.
  validateSchemaSync(processProof ? "cartulary.harness_private_launch_observation.v2" : recordSchema, observation);
  const value = `${JSON.stringify(observation)}\n`;
  if (Buffer.byteLength(value) > maximumBytes) throw new Error("execution record exceeds bound");
  return value;
}

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
      processProof ??= proof;
      const value = encodeObservation(record, limits.maximum_record_bytes, processProof);
      atomicLocalFile(file, value, { replace: true, maximumReplacementBytes: limits.maximum_record_bytes });
    } catch { /* Optional loss leaves the previously published incomplete fact. */ }
  };
  try {
    if (environment[executionEnabledKey] !== "1") return { update() {} };
    root = channel(repoRoot, environment);
    limits = parseStrictJSON(readLocalFile(path.join(root, "limits.json")).toString("utf8"));
    if (!Number.isInteger(limits.maximum_execution_records) || limits.maximum_execution_records < 1 || limits.maximum_execution_records > 4096 || limits.maximum_record_bytes !== 65536) throw new Error("invalid execution limits");
    if (!open()) return { update() {} };
    const beginning = encodeObservation(initial, limits.maximum_record_bytes);
    const start = Number.parseInt(randomUUID().slice(0, 8), 16) % limits.maximum_execution_records;
    for (let i = 0; i < limits.maximum_execution_records; i++) {
      const slot = path.join(root, `slot-${(start + i) % limits.maximum_execution_records}`);
      try { atomicLocalFile(slot, beginning); file = slot; break; }
      catch (error) { if (error.code !== "EEXIST") throw error; }
    }
    if (!file) { atomicLocalFile(path.join(root, "truncated"), "1\n"); return { update() {} }; }
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
      atomicLocalFile(file, encodeObservation(record, 65536), { replace: true, maximumReplacementBytes: 65536 });
    } catch { /* Preserve the incomplete beginning on optional publication loss. */ }
  } };
}

export { digest as executionRecordsDigest };
