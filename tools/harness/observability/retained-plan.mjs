import { createHash } from "node:crypto";
import path from "node:path";

import { lifecyclePolicy } from "../contract/generated/lifecycle-policy.mjs";
import { assertMachineEvidencePath, canonicalJSONString, parseStrictJSON, readRetainedArtifactBytes, semanticJSONDigest, validateSchemaSync } from "../contract/index.mjs";
import { retainedGoWorkerBudget } from "../scheduler/work-graph/capability.mjs";

// A run publishes the planner's document, without launch bindings, as
// plan.json before any child starts. Readers take the graph, projections and
// provenance from it and never recompile a selection.
export const planDocument = "plan.json";
export const planSchemaID = "cartulary.harness_runtime_plan.v9";

const digest = (value) => semanticJSONDigest(value);
const bytesDigest = (value) => `sha256:${createHash("sha256").update(value).digest("hex")}`;
function same(actual, expected, label) {
  if (canonicalJSONString(actual) !== canonicalJSONString(expected)) throw new Error(`retained ${label} binding does not close`);
}
function ordered(values, label) {
  if (values.some((value, index) => index > 0 && values[index - 1] >= value)) throw new Error(`retained ${label} must be sorted and unique`);
}

/** Read a run's published plan within the plan byte bound and admit its schema. */
export function readRunPlan(runRoot) {
  const bytes = readRetainedArtifactBytes(runRoot, planDocument, lifecyclePolicy.limits.plan);
  const plan = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes));
  validateSchemaSync(planSchemaID, plan);
  return plan;
}

/**
 * Bind a plan to its run: the plan's manifest is the run manifest and its
 * graph recomputes to the manifest's graph digest, unit by unit.
 */
export function bindRunPlan(plan, manifest) {
  validateSchemaSync(planSchemaID, plan);
  same(plan.manifest, manifest, "plan manifest");
  const { graph } = plan;
  ordered(graph.units.map((unit) => unit.unit_id), "graph units");
  for (const unit of graph.units) {
    const { semantic_digest, ...preimage } = unit;
    same(digest(preimage), semantic_digest, "unit semantics");
  }
  same(digest({ units: graph.units }), graph.graph_digest, "graph");
  same(graph.graph_digest, manifest.graph_digest, "manifest graph");
  return plan;
}

/** Read, bind and validate a run's retained plan: the reader every consumer shares. */
export function loadRetainedPlan(runRoot, canonical, cleanup) {
  const plan = readRunPlan(runRoot);
  return { plan, ...validateRetainedPlan(plan, canonical, cleanup) };
}

// Validate retained preimages only. Historical qualification never reads the
// current checkout to fill in missing source, policy, or execution identities.
export function validateRetainedPlan(plan, canonical, cleanup) {
  validateSchemaSync("cartulary.harness_cleanup_results.v2", cleanup);
  const { manifest, summary, registrations, targetSummaries } = canonical;
  bindRunPlan(plan, manifest);
  same(cleanup.run_id, manifest.run_id, "cleanup run");
  if (cleanup.results.some((result, index) => result.sequence !== index + 1 || result.outcome !== "completed")) {
    throw new Error("retained cleanup is incomplete");
  }
  for (const required of [planDocument, "cleanup-results.json"]) {
    if (!summary.artifact_refs.includes(required)) throw new Error(`retained summary omits ${required}`);
  }
  const source = plan.source_inputs;
  ordered(source.map((entry) => entry.path), "source entries");
  for (const entry of source) {
    if (entry.path.includes("\\") || entry.path.includes("\0") || path.posix.isAbsolute(entry.path) ||
        entry.path.split("/").some((part) => !part || part === "." || part === "..")) throw new Error("unsafe retained source path");
    assertMachineEvidencePath(entry.path);
  }
  same(digest({ schema_id: "cartulary.source_snapshot_digest.v2", entries: source }), manifest.source_digest, "source snapshot");
  const sourceByPath = new Map(source.map((entry) => [entry.path, entry]));
  const toolchain = sourceByPath.get("tools/toolchain_pins.json");
  if (toolchain?.kind !== "file") throw new Error("retained toolchain input is missing");
  same(toolchain.byte_digest, manifest.toolchain_digest, "toolchain");
  const policyInput = sourceByPath.get("tools/harness_instrumentation_policy.json");
  if (policyInput?.kind !== "file") throw new Error("retained instrumentation input is missing");
  same(bytesDigest(plan.instrumentation_policy_bytes), policyInput.byte_digest, "instrumentation bytes");
  const policy = parseStrictJSON(plan.instrumentation_policy_bytes);
  validateSchemaSync("cartulary.harness_instrumentation_policy.v5", policy);
  same(digest(policy), manifest.instrumentation.policy_digest, "instrumentation policy");
  const { snapshot_digest, ...capability } = manifest.capability_snapshot;
  same(digest(capability), snapshot_digest, "capability snapshot");
  same(snapshot_digest, manifest.system_digest, "system capacity");
  retainedGoWorkerBudget(manifest);
  ordered(plan.host_profile.cpu_models, "CPU models");

  const { graph, projections } = plan;
  same(graph.units.map((unit) => unit.unit_id), [...registrations.keys()].sort(), "registered units");
  const units = new Map(graph.units.map((unit) => [unit.unit_id, unit]));
  const rowOwners = new Set();
  for (const unit of graph.units) {
    const registered = registrations.get(unit.unit_id);
    for (const key of ["kind", "needs", "resource_claims", "service_dependencies"]) same(registered[key], unit[key], `unit ${key}`);
    const rows = unit.row_bindings.map((row) => row.row_id);
    ordered(rows, "unit rows");
    same(unit.current_run_evidence_outputs.filter((output) => output.startsWith("rows/")), rows.map((id) => `rows/${id}.json`), "row outputs");
    same(registered.row_ids, rows, "registered rows");
    for (const id of rows) {
      if (rowOwners.has(id)) throw new Error("retained row has multiple graph owners");
      rowOwners.add(id);
    }
  }
  same(Object.keys(projections).sort(), [...targetSummaries.keys()].sort(), "target roster");
  for (const [target, projection] of Object.entries(projections)) {
    ordered(projection.unit_ids, "projection units");
    const selected = new Set(projection.unit_ids);
    const evidence = new Set();
    for (const id of selected) {
      const unit = units.get(id);
      if (!unit || unit.needs.some((dependency) => !selected.has(dependency))) throw new Error("retained projection dependency closure is incomplete");
      for (const output of unit.current_run_evidence_outputs) if (!output.startsWith("unit-results/")) evidence.add(output);
    }
    const result = targetSummaries.get(target);
    same(projection.unit_ids, result.unit_ids, "target units");
    same(projection.command_id, result.command_id, "target command");
    same(digest({ target, evidence_outputs: [...evidence].sort() }), result.workload_digest, "target workload");
  }
  const provider = projections[manifest.target];
  if (!provider) throw new Error("retained provider projection is missing");
  same(provider.command_id, manifest.command_id, "provider command");
  same(provider.unit_ids, graph.units.map((unit) => unit.unit_id), "provider graph");
  return { units, systemProfileDigest: digest({ host_profile: plan.host_profile, capability_snapshot: manifest.capability_snapshot }) };
}
