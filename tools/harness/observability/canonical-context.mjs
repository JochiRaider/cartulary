import { createHash } from "node:crypto";
import path from "node:path";

import { canonicalJSONString, parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { assertMachineEvidencePath } from "../contract/artifact-reader.mjs";
import { retainedGoWorkerBudget } from "../scheduler/work-graph/capability.mjs";

const digest = (value) => `sha256:${createHash("sha256").update(canonicalJSONString(value)).digest("hex")}`;
const bytesDigest = (value) => `sha256:${createHash("sha256").update(value).digest("hex")}`;
function same(actual, expected, label) {
  if (canonicalJSONString(actual) !== canonicalJSONString(expected)) throw new Error(`retained ${label} binding does not close`);
}
function ordered(values, label) {
  if (values.some((value, index) => index > 0 && values[index - 1] >= value)) throw new Error(`retained ${label} must be sorted and unique`);
}

// Validate retained preimages only. Historical qualification never reads the
// current checkout to fill in missing source, policy, or execution identities.
export function validateExecutionContext(context, canonical, cleanup) {
  validateSchemaSync("cartulary.harness_execution_context.v3", context);
  validateSchemaSync("cartulary.harness_cleanup_results.v1", cleanup);
  const { manifest, summary, registrations, targetSummaries } = canonical;
  same(context.manifest_digest, digest(manifest), "manifest");
  same(cleanup.run_id, manifest.run_id, "cleanup run");
  if (cleanup.results.some((result, index) => result.sequence !== index + 1 || result.outcome !== "completed")) {
    throw new Error("retained cleanup is incomplete");
  }
  for (const required of ["execution-context.json", "cleanup-results.json"]) {
    if (!summary.artifact_refs.includes(required)) throw new Error(`retained summary omits ${required}`);
  }
  const source = context.source_inputs;
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
  same(bytesDigest(context.instrumentation_policy_bytes), policyInput.byte_digest, "instrumentation bytes");
  const policy = parseStrictJSON(context.instrumentation_policy_bytes);
  validateSchemaSync("cartulary.harness_instrumentation_policy.v3", policy);
  same(digest(policy), manifest.instrumentation.policy_digest, "instrumentation policy");
  const { snapshot_digest, ...capability } = manifest.capability_snapshot;
  same(digest(capability), snapshot_digest, "capability snapshot");
  same(snapshot_digest, manifest.system_digest, "system capacity");
  retainedGoWorkerBudget(manifest);
  ordered(context.host_profile.cpu_models, "CPU models");

  const { graph, projections } = context;
  same(digest({ units: graph.units }), graph.graph_digest, "graph");
  same(graph.graph_digest, manifest.graph_digest, "manifest graph");
  ordered(graph.units.map((unit) => unit.unit_id), "graph units");
  same(graph.units.map((unit) => unit.unit_id), [...registrations.keys()].sort(), "registered units");
  const units = new Map(graph.units.map((unit) => [unit.unit_id, unit]));
  const rowOwners = new Set();
  for (const unit of graph.units) {
    const { semantic_digest, ...preimage } = unit;
    same(digest(preimage), semantic_digest, "unit semantics");
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
  return { systemProfileDigest: digest({ host_profile: context.host_profile, capability_snapshot: manifest.capability_snapshot }) };
}
