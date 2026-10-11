import path from "node:path";
import { canonicalJSONString, parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
import { readLocalFile } from "../../runtime/secure-local-files.mjs";
import { loadRetainedObservability } from "../observability.mjs";
import { approvedPlan, validatePlan, digest, manifestSchema, resultSchema, treatmentOrder } from "./policy.mjs";
import { assessWorkload, eligible, intervalUnion } from "./analysis.mjs";
import { assembleMeasurement, measurementFailure, requireBasicTreatment } from "./measurement.mjs";
import { loadMeasurementTrace } from "./trace.mjs";
const invalid = (message) => Object.assign(Error(message), { exit_code: 11 });
const read = (file, maximum = 1024 * 1024) => { const bytes = readLocalFile(file, { maximum }); return { bytes, value: parseStrictJSON(bytes.toString("utf8")) }; };
function immutable(value) {
  for (const child of Object.values(value)) if (child && typeof child === "object") immutable(child);
  return Object.freeze(value);
}
function relative(root, value) {
  if (path.isAbsolute(value) || path.posix.normalize(value) !== value || value.startsWith("../") || value === ".." || value.includes("\\")) throw invalid("unsafe qualification reference");
  return path.join(root, value);
}
export async function loadAttemptFacts(root, entry, manifest, plan, order, graphByWorkload = new Map()) {
      const { bytes, value } = read(relative(root, entry.result_ref)); validateSchemaSync(resultSchema, value);
      if (digest(bytes) !== entry.result_digest || value.source_digest !== manifest.source_digest || value.plan_digest !== manifest.plan_digest || value.order !== order) throw invalid("qualification attempt identity/digest mismatch");
      const raw = value.measurement_trace ? loadMeasurementTrace(root, value, value.measurement_trace) : null;
      if (!!raw !== !!value.support_publication) throw invalid("measurement publication accounting mismatch");
      if (value.measurement) {
        if (!raw) throw invalid("missing measurement trace");
        const retained = await loadRetainedObservability(relative(root, entry.run_ref)), run = retained.run;
        const reconstructed = assembleMeasurement(raw, retained, value.mode);
        if (canonicalJSONString(reconstructed) !== canonicalJSONString(value.measurement)) throw invalid("measurement reconstruction mismatch");
        if (measurementFailure(raw, reconstructed) !== value.failure) throw invalid("measurement disposition mismatch");
        const workload = plan.workloads.find((entry) => entry.id === value.workload);
        if (!workload || run.manifest.run_id !== value.run_id || run.manifest.target !== workload.target || run.manifest.source_state !== "clean" || run.manifest.source_commit !== manifest.source_commit || run.manifest.instrumentation.mode !== value.mode) throw invalid("qualification run binding mismatch");
        for (const key of ["source_digest", "toolchain_digest", "system_digest"]) if (run.manifest[key] !== manifest[key]) throw invalid("qualification frozen identity mismatch");
        if (run.manifest.instrumentation.policy_digest !== manifest.policy_digest || run.summary.wall_duration_ms !== value.measurement.canonical_ms || run.manifest.cache_mode !== (value.phase === "cold" ? "cold" : "off")) throw invalid("qualification preparation or canonical duration mismatch");
        if (run.manifest.declared_inputs.OWNER !== workload.owner || run.manifest.declared_inputs.ROWS !== workload.rows.join(",")) throw invalid("qualification workload selection mismatch");
        const prior = graphByWorkload.get(value.workload); if (prior && prior !== run.manifest.graph_digest) throw invalid("qualification graph changed"); graphByWorkload.set(value.workload, run.manifest.graph_digest);
        if (value.mode === "basic" && eligible(value)) {
          requireBasicTreatment(retained);
          if (!retained.terminalReceipt) throw invalid("eligible basic attempt lacks receipt");
          const drain = retained.envelope.intervals.filter((entry) => entry.phase === "collector_shutdown").reduce((sum,entry)=>sum+entry.end_ms-entry.start_ms,0);
          const receipt = retained.terminalReceipt;
          if (intervalUnion(value.measurement.finalization_intervals) < drain + Math.max(0,receipt.publication_start_ms-receipt.outer_duration_ms-10)) throw invalid("qualification finalization omits published intervals");
        }
        if (value.exit_code === 0 && run.summary.status !== "pass") throw invalid("qualification command outcome mismatch");
      } else if (value.failure === null) throw invalid("missing measurement without failure");
  return Object.freeze({ result: immutable(value), raw });
}
export async function loadAttemptEvidence(...args) { return (await loadAttemptFacts(...args)).result; }
export async function validateQualification(file) {
  try {
    const { value: manifest } = read(file); validateSchemaSync(manifestSchema, manifest);
    const root = path.dirname(file), { value: plan } = read(path.join(root, "plan.json"));
    const planDigest = validatePlan(plan); if (manifest.plan_digest !== planDigest) throw invalid("qualification plan binding mismatch");
    const attempts = [], identities = new Set(), runIDs = new Set(), references = new Set(), graphByWorkload = new Map();
    for (const [order, entry] of manifest.attempts.entries()) {
      if (references.has(entry.result_ref) || references.has(entry.run_ref)) throw invalid("duplicate qualification reference"); references.add(entry.result_ref); references.add(entry.run_ref);
      const value = await loadAttemptEvidence(root, entry, manifest, plan, order, graphByWorkload);
      if (identities.has(value.attempt_id) || runIDs.has(value.run_id)) throw invalid("duplicate qualification identity");
      identities.add(value.attempt_id); runIDs.add(value.run_id);
      attempts.push(value);
    }
    const exclusionMap = new Map();
    for (const exclusion of manifest.exclusions) { if (!identities.has(exclusion.attempt_id) || exclusionMap.has(exclusion.attempt_id)) throw invalid("invalid qualification exclusion"); exclusionMap.set(exclusion.attempt_id, exclusion.reason); }
    for (const attempt of attempts) {
      const reason = attempt.phase === "cold" || attempt.phase === "warmup" ? attempt.phase : eligible(attempt) ? null : ["command_failed", "measurement_lost", "cancelled"].includes(attempt.failure) ? attempt.failure : "ineligible";
      if ((exclusionMap.get(attempt.attempt_id) ?? null) !== reason) throw invalid("qualification exclusion does not match evidence");
    }
    const failures = [], workloads = [];
    let cursor = 0;
    for (const workload of plan.workloads) {
      const selected = attempts.filter((attempt) => attempt.workload === workload.id);
      const blocks = Math.max(0, ...selected.filter((a) => a.phase === "measured").map((a) => a.block));
      if (![5,6].includes(blocks)) failures.push(`${workload.id}:incomplete_blocks`);
      for (const [phase, block] of [["cold",0],["warmup",0],...Array.from({length:Math.min(blocks,6)},(_,i)=>["measured",i+1])]) {
        for (const treatment of treatmentOrder(plan,workload.id,phase,block)) {
          const actual = attempts[cursor]; if (!actual) continue; cursor++; if ( actual.workload !== workload.id || actual.phase !== phase || actual.block !== block || actual.mode !== treatment.mode || actual.readers !== treatment.readers) throw invalid("qualification treatment sequence differs from frozen order");
        }
      }
      if (blocks === 6 && !assessWorkload(selected.filter((attempt) => attempt.phase !== "measured" || attempt.block <= 5), plan).needs_sixth) throw invalid("unrequested sixth qualification block");
      const assessment = assessWorkload(selected, plan); workloads.push({ workload: workload.id, ...assessment });
      if (!assessment.passed) failures.push(`${workload.id}:${assessment.unstable ? "unstable" : "budget_or_coverage"}`);
      if (selected.some((attempt) => !eligible(attempt))) failures.push(`${workload.id}:ineligible_attempt`);
    }
    if (cursor !== attempts.length) throw invalid("unexpected qualification attempts");
    if (manifest.status !== "complete") failures.push("manifest_not_complete");
    return { manifest, workloads, failures, passed: failures.length === 0 };
  } catch (error) { if (error.exit_code === 2) throw invalid(error.message); error.exit_code ??= 11; throw error; }
}
