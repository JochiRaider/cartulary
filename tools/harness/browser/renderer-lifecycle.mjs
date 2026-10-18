// Physical adapter only. The native owner chooses allocation, borrowing,
// accounting and retirement; this process never decides reuse policy.
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parseStrictJSON, validateSchemaSync, secureWriteFile, normalizeFailureRecord } from "../contract/index.mjs";
import { findCleanupDiagnostic } from "../runtime/cleanup-diagnostic.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { borrowRuntime } from "../runtime/session-client.mjs";
import { startVisualRendererLease, rendererDocker } from "./visual-renderer-lease.mjs";
import { settleRenderer } from "./renderer-quiescence.mjs";
import { findOwnedRenderer, removeOwnedRenderer } from "../runtime/owned-renderer.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
export async function rendererLifecycle(operation, file, environment = process.env) {
  const runtime = borrowRuntime({ repoRoot: root, runRoot: path.join(environment.CARTULARY_TEST_RESULTS_DIR, environment.CARTULARY_TEST_RUN_ID), environment });
  if (!path.isAbsolute(file) || !file.startsWith(`${runtime.root}/`) || path.normalize(file) !== file) throw new Error("foreign renderer acquisition");
  if (operation === "acquire") {
    let stage="configuration", lease;
    try {
      lease = await startVisualRendererLease({ root, environment, diagnostics: environment.CARTULARY_RENDERER_DIAGNOSTICS === "1", onStage: value => {stage=value;}, onOwnedResource: resource => runtime.record({ ...resource, depends_on: JSON.parse(environment.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES) }) });
      stage="quiescence";
      await settleRenderer(lease.containerID, rendererDocker);
      stage="publication";
      const physical = { schema_id: "cartulary.harness_renderer_physical.v1", run_id: runtime.runID,
        proof: lease.proof, container_id: lease.containerID, diagnostic_binding: lease.diagnosticBinding, environment: lease.environment,
        attestation: lease.attestation, trust: lease.trustAttestation };
      validateSchemaSync(physical.schema_id, physical);
      secureWriteFile(file, `${JSON.stringify(physical)}\n`);
      return physical;
    } catch (error) {
      try { lease?.cleanup(); } catch (cleanup) { error.cleanupFailures = [...(error.cleanupFailures ?? []), cleanup]; }
      const diagnostic = {schema_id:"cartulary.harness_renderer_startup.v1",stage,container_state:error.rendererState ?? null,cleanup_failed:(error.cleanupFailures?.length ?? 0)>0};
      validateSchemaSync(diagnostic.schema_id,diagnostic);
      secureWriteFile(path.join(path.dirname(file),"startup-diagnostic.json"),`${JSON.stringify(diagnostic)}\n`);
      throw error;
    }
  }
  if (operation === "settle" || operation === "release") {
    const physical = parseStrictJSON(readLocalFile(file, { maximum: 1048576 }).toString());
    validateSchemaSync("cartulary.harness_renderer_physical.v1", physical);
    if (physical.run_id !== runtime.runID) throw new Error("renderer run changed");
    if (operation === "release") { removeOwnedRenderer(physical.proof, rendererDocker); return {}; }
    if (findOwnedRenderer(physical.proof, rendererDocker) !== physical.container_id) throw new Error("renderer generation changed");
    return await settleRenderer(physical.container_id, rendererDocker);
  }
  throw new Error("unknown renderer lifecycle operation");
}
if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try { process.stdout.write(`${JSON.stringify(await rendererLifecycle(process.argv[2], process.argv[3]))}\n`); }
  catch (error) {
    const failure = normalizeFailureRecord({failure_class:error.failure_class ?? "harness",failure_reason:error.failure_reason ?? "fixture_error"});
    process.stdout.write(`${JSON.stringify({failure_class:failure.failure_class,failure_reason:failure.failure_reason})}\n`);
    const diagnostic = findCleanupDiagnostic(error);
    process.stderr.write(diagnostic ? `${JSON.stringify(diagnostic)}\n` : "renderer lifecycle rejected or failed\n");
    process.exitCode = failure.failure_class === "artifact" ? 11 : 12;
  }
}
