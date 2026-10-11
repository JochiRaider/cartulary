import path from "node:path";
import { readLocalClock, captureClockAnchor } from "../runtime/boot-clock.mjs";
import { atomicLocalFile } from "../runtime/secure-local-files.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import { createResourceCollector } from "./resource-collector.mjs";
import { instrumentationPolicy } from "./resource-collector.mjs";
import { createExecutionSession } from "../runtime/execution-observations.mjs";
import { registerWrapperGraph } from "../runtime/wrapper-observation.mjs";

// The caller may scan retained output only after close resolves. Failure to
// establish quiescence is required cleanup failure, not optional diagnostic loss.
export function createDiagnosticSession({ mode, epoch, elapsed = () => Number((process.hrtime.bigint() - epoch) / 1_000_000n), collectorFactory = createResourceCollector }) {
  const intervals = [];
  const firstAnchor = mode === "basic" ? captureClockAnchor() : null;
  let canonicalEpoch = null;
  let previousBoundary = 0, resources, context, completion, execution, wrapper;
  function mark(phase) {
    const end = elapsed();
    if (mode === "basic") intervals.push({ phase, start_ms: previousBoundary, end_ms: end });
    previousBoundary = end;
  }
  return {
    mark,
    schedulerEpoch(epoch) { canonicalEpoch = Number(epoch) / 1e6; },
    start(options) {
      context = options;
      wrapper = registerWrapperGraph(options);
      if (options.runtime) execution = createExecutionSession({ ...options, policy: instrumentationPolicy().policy,
        onProcessProof: (proof, record) => resources?.registerProof(proof, record.unit_id, { invocationID: record.invocation_id }) });
      resources = collectorFactory({ ...options, epoch, onRelationship: (record) => execution?.record(record) });
      return resources;
    },
    executionEnvironment() { return execution?.environment ?? {}; },
    captureExecution() { execution?.capture(); },
    observeExecution(record) { execution?.record(record); },
    scanStarted() { wrapper?.scanStarted(); },
    scanFinished(passed) { wrapper?.scanFinished(passed); },
    invocationComplete() { wrapper?.complete(); },
    close(stopLive = () => {}) {
      if (completion) return completion;
      completion = (async () => {
        execution?.capture();
        try { await resources?.stop(); } finally { await stopLive(); }
        const lastAnchor = mode === "basic" ? captureClockAnchor() : null;
        execution?.publish(canonicalEpoch === null ? null : { clock_identity: readLocalClock().clock_identity,
          epoch_ms: canonicalEpoch, anchors: [firstAnchor, lastAnchor].filter(Boolean) });
        mark("collector_shutdown");
        if (mode !== "basic" || !context) return;
        const { manifest, runRoot } = context;
        try {
          const receipt = { schema_id: "cartulary.harness_invocation_envelope.v2",
            run_id: manifest.run_id, source_digest: manifest.source_digest, graph_digest: manifest.graph_digest,
            invocation_digest: wrapper?.digest ?? null, boundary: "graph_main_entry_to_pre_scan", clock: "graph_process_monotonic", canonical_ref: "unit-events.ndjson", intervals,
            exclusions: ["make_and_preflight", "module_imports", "retained_secret_scan", "command_return", "receipt_publication"] };
          validateSchemaSync(receipt.schema_id, receipt);
          atomicLocalFile(path.join(runRoot, "diagnostics/invocation-envelope.json"), `${JSON.stringify(receipt)}\n`);
        } catch { /* Optional receipt cannot replace required evidence outcomes. */ }
      })();
      return completion;
    },
  };
}
