// Closed native diagnostic entrypoints. Each owns public input preflight and
// requires an installed binary; read-only invocations never provision tools.
const nativeTools = Object.freeze({
  "test-run-status": Object.freeze({ binary: "$(HARNESS_RUNTIME_BIN)", arguments: "diagnostics --make-inputs" }),
});
export const nativeTool = (target) => nativeTools[target] ?? null;
