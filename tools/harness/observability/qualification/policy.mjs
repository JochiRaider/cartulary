import path from "node:path";
import { readFileSync } from "node:fs";
import { createHash } from "node:crypto";
import { canonicalJSONString, parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
export const repositoryRoot = path.resolve(import.meta.dirname, "../../../..");
export const digest = (bytes) => `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
export const planSchema = "cartulary.harness_instrumentation_qualification_plan.v3";
export const manifestSchema = "cartulary.harness_instrumentation_qualification.v2";
export const resultSchema = "cartulary.harness_instrumentation_qualification_result.v2";
export function approvedPlan(root = repositoryRoot) {
  const value = parseStrictJSON(readFileSync(path.join(root, "tools/harness_instrumentation_qualification_plan.json"), "utf8"));
  validateSchemaSync(planSchema, value); return value;
}
export function validatePlan(value, root = repositoryRoot) {
  validateSchemaSync(planSchema, value);
  if (canonicalJSONString(value) !== canonicalJSONString(approvedPlan(root))) throw Object.assign(Error("qualification plan differs from authored policy"), { exit_code: 2 });
  return digest(canonicalJSONString(value));
}
export function treatmentOrder(plan, workload, phase, block) {
  const bytes = createHash("sha256").update(`${plan.seed}:${workload}:${phase}:${block}`).digest();
  let seed = bytes.readUInt32LE(0);
  const random = () => { seed ^= seed << 13; seed ^= seed >>> 17; seed ^= seed << 5; return (seed >>> 0) / 4294967296; };
  const treatments = plan.modes.flatMap((mode) => plan.reader_counts.map((readers) => ({ mode, readers })));
  for (let i = treatments.length - 1; i > 0; i--) { const j = Math.floor(random() * (i + 1)); [treatments[i], treatments[j]] = [treatments[j], treatments[i]]; }
  return treatments;
}
