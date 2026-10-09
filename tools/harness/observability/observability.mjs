import { existsSync } from "node:fs";
import path from "node:path";
import { repoRoot, validateSchemaSync } from "../contract/index.mjs";
import { compareRunMetadata } from "./run-comparison.mjs";
import { resourceProjection } from "./resource-projection.mjs";
import { analyzePerformance } from "./performance-analysis.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";
import { formatPerformanceExplanation } from "./performance-presentation.mjs";

function isCanonicalRun(dir) {
  return ["run-manifest.json", "run-summary.json", "unit-events.ndjson"].every(
    (name) => existsSync(path.join(dir, name)),
  );
}

export function resolveExactRunDir(resultsDir, runID = "") {
  const selected = path.resolve(repoRoot, resultsDir);
  if (!existsSync(selected)) throw new Error("RESULTS_DIR does not exist");
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

export async function loadRetainedObservability(runDir) {
  const run = await validateCanonicalRun(runDir);
  const sourceDigests = [
    run.manifest.source_digest,
    run.manifest.toolchain_digest,
    run.manifest.system_digest,
    run.manifest.graph_digest,
  ];
  return {
    run,
    index: {
      schema_id: "cartulary.harness_canonical_observability.v1",
      status: "complete",
      invocations: [
        {
          target: run.manifest.target,
          run_id: run.manifest.run_id,
          source_digests: sourceDigests,
        },
      ],
    },
  };
}

export async function performanceExplanation(runDir, target = "", { comparisonDir = "", comparison = "equivalent" } = {}) {
  const run = await validateCanonicalRun(runDir);
  const comparisonResult = comparisonDir ? compareRunMetadata(run, await validateCanonicalRun(comparisonDir), comparison) : null;
  let envelope = null;
  try {
    envelope = JSON.parse(readLocalFile(path.join(runDir, "diagnostics/invocation-envelope.json"), { maximum: 65536 }));
    validateSchemaSync("cartulary.harness_invocation_envelope.v1", envelope);
    for (const key of ["run_id", "source_digest", "graph_digest"]) if (envelope[key] !== run.manifest[key]) throw new Error("envelope identity mismatch");
    let end = 0;
    for (const interval of envelope.intervals) { if (interval.start_ms !== end || interval.end_ms < end) throw new Error("invalid envelope intervals"); end = interval.end_ms; }
  } catch (error) { if (error.code !== "ENOENT") throw error; }
  return analyzePerformance(run, target, { comparison: comparisonResult, resources: resourceProjection(runDir, run), envelope });
}

export async function printObservabilityPerformance(runDir, target = "", options = {}) {
  const value = await performanceExplanation(runDir, target, options);
  process.stdout.write(formatPerformanceExplanation(value, options));
  return value;
}
