import path from "node:path";
import { repoRoot, validateSchemaSync, assertMachineEvidencePath, assertRetainedArtifactFile } from "../contract/index.mjs";
import { compareRunMetadata } from "./run-comparison.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { parseStrictJSON } from "../contract/strict-json.mjs";
import { resourceProjection } from "./resource-projection.mjs";
import { analyzePerformance } from "./performance-analysis.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";
import { formatPerformanceExplanation } from "./performance-presentation.mjs";
import { readTerminalReceipt, receiptDigest } from "./terminal-receipt.mjs";
import { readExecutionIndex } from "./execution-reader.mjs";

function isCanonicalRun(dir) {
  try {
    for (const name of ["run-manifest.json", "run-summary.json", "unit-events.ndjson"]) assertRetainedArtifactFile(dir, name);
    return true;
  } catch (error) { if (error.code === "ENOENT") return false; throw error; }
}

export function resolveExactRunDir(resultsDir, runID = "") {
  const selected = path.resolve(repoRoot, resultsDir);
  assertMachineEvidencePath(selected);
  if (runID && !/^[A-Za-z0-9][A-Za-z0-9_.-]*$/u.test(runID)) throw new Error("invalid RUN_ID");
  if (runID) {
    const candidate = path.basename(selected) === runID ? selected : path.join(selected, runID);
    if (!isCanonicalRun(candidate)) {
      throw new Error("RUN_ID does not identify canonical retained evidence");
    }
    return candidate;
  }
  if (isCanonicalRun(selected)) return selected;
  throw new Error("RUN_ID is required when RESULTS_DIR names a result root");
}

// Freeze normalized facts only after every published component has closed. Maps
// expose a read-only interface: Object.freeze(Map) alone still permits mutation.
function immutableFacts(value, seen = new WeakMap()) {
  if (value === null || typeof value !== "object") return value;
  if (seen.has(value)) return seen.get(value);
  if (value instanceof Map) {
    const entries = new Map();
    const view = {
      get size() { return entries.size; }, get: (key) => entries.get(key), has: (key) => entries.has(key),
      keys: () => entries.keys(), values: () => entries.values(), entries: () => entries.entries(),
      [Symbol.iterator]: () => entries[Symbol.iterator](),
      forEach: (callback, receiver) => entries.forEach((item, key) => callback.call(receiver, item, key, view)),
    };
    seen.set(value, view);
    for (const [key, item] of value) entries.set(key, immutableFacts(item, seen));
    return Object.freeze(view);
  }
  seen.set(value, value);
  for (const [key, item] of Object.entries(value)) value[key] = immutableFacts(item, seen);
  return Object.freeze(value);
}


function readEnvelope(runDir, run) {
  let envelope, bytes;
  try {
    bytes = readLocalFile(path.join(runDir, "diagnostics/invocation-envelope.json"), { maximum: 65536 });
    envelope = parseStrictJSON(bytes.toString("utf8"), "invocation envelope");
  } catch (error) { if (error.code === "ENOENT") return null; throw error; }
  validateSchemaSync("cartulary.harness_invocation_envelope.v2", envelope);
  if (run.manifest.instrumentation.mode !== "basic") throw new Error("envelope requires basic diagnostics");
  for (const key of ["run_id", "source_digest", "graph_digest"]) if (envelope[key] !== run.manifest[key]) throw new Error("envelope identity mismatch");
  let end = 0;
  for (const interval of envelope.intervals) {
    if (interval.start_ms !== end || interval.end_ms < end) throw new Error("invalid envelope intervals");
    end = interval.end_ms;
  }
  return { value: envelope, digest: receiptDigest(bytes) };
}


export async function loadRetainedObservability(runDir, { canonical } = {}) {
  const run = canonical ?? await validateCanonicalRun(runDir);
  const execution = readExecutionIndex(runDir, run);
  const loadedEnvelope = readEnvelope(runDir, run);
  const envelope = loadedEnvelope?.value ?? null;
  const terminalReceipt = readTerminalReceipt(runDir, run, envelope, loadedEnvelope?.digest);
  let resourceIndex = null;
  const resources = resourceProjection(runDir, run, execution, { onValidatedIndex: (index) => { resourceIndex = index; } });
  const sourceDigests = [run.manifest.source_digest, run.manifest.toolchain_digest,
    run.manifest.system_digest, run.manifest.graph_digest];
  return immutableFacts({
    run, execution, envelope, resources, resourceIndex, terminalReceipt,
    index: {
      schema_id: "cartulary.harness_canonical_observability.v1", status: "complete",
      invocations: [{ target: run.manifest.target, run_id: run.manifest.run_id, source_digests: sourceDigests }],
    },
  });
}

export async function performanceExplanation(runDir, target = "", { comparisonDir = "", comparison = "equivalent" } = {}) {
  const retained = await loadRetainedObservability(runDir);
  const reference = comparisonDir ? (path.resolve(comparisonDir) === path.resolve(runDir)
    ? retained : await loadRetainedObservability(comparisonDir)) : null;
  const comparisonResult = reference ? compareRunMetadata(retained.run, reference.run, comparison) : null;
  return analyzePerformance(retained.run, target, { comparison: comparisonResult,
    resources: retained.resources, envelope: retained.envelope, terminalReceipt: retained.terminalReceipt, execution: retained.execution });
}

export async function printObservabilityPerformance(runDir, target = "", options = {}) {
  const value = await performanceExplanation(runDir, target, options);
  process.stdout.write(formatPerformanceExplanation(value, options));
  return value;
}
