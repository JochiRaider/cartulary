import path from "node:path";
import { atomicLocalFile } from "../runtime/secure-local-files.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import { createResourceCollector } from "./resource-collector.mjs";

// The caller may scan retained output only after close resolves. Failure to
// establish quiescence is required cleanup failure, not optional diagnostic loss.
export function createDiagnosticSession({ mode, epoch, elapsed = () => Number((process.hrtime.bigint() - epoch) / 1_000_000n), collectorFactory = createResourceCollector }) {
  const intervals = [];
  let previousBoundary = 0, resources, context, completion;
  function mark(phase) {
    const end = elapsed();
    if (mode === "basic") intervals.push({ phase, start_ms: previousBoundary, end_ms: end });
    previousBoundary = end;
  }
  return {
    mark,
    start(options) { context = options; resources = collectorFactory({ ...options, epoch }); return resources; },
    close(stopLive = () => {}) {
      if (completion) return completion;
      completion = (async () => {
        try { await resources?.stop(); } finally { await stopLive(); }
        mark("collector_shutdown");
        if (mode !== "basic" || !context) return;
        const { manifest, runRoot } = context;
        try {
          const receipt = { schema_id: "cartulary.harness_invocation_envelope.v1",
            run_id: manifest.run_id, source_digest: manifest.source_digest, graph_digest: manifest.graph_digest,
            boundary: "graph_main_entry_to_pre_scan", clock: "graph_process_monotonic", canonical_ref: "unit-events.ndjson", intervals,
            exclusions: ["make_and_preflight", "module_imports", "retained_secret_scan", "command_return", "receipt_publication"] };
          validateSchemaSync(receipt.schema_id, receipt);
          atomicLocalFile(path.join(runRoot, "diagnostics/invocation-envelope.json"), `${JSON.stringify(receipt)}\n`);
        } catch { /* Optional receipt cannot replace required evidence outcomes. */ }
      })();
      return completion;
    },
  };
}
