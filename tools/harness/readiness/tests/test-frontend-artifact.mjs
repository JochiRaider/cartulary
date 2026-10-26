import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";
import { createRuntime } from "../../runtime/session-client.mjs";
import { validateFrontendAttachment } from "../../test-support/owner-fixtures.mjs";
import { resolveFrontendArtifact, sealFrontendArtifact } from "../frontend-artifact.mjs";
import { resolveBrowserFrontendArtifact } from "../../generated-artifacts/execution-topology.mjs";
import { createContractTestContext } from "../../tests/contract-test-context.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
export function registerEvidenceTests() {
  test("browser_artifact_admission: browser attachments bind the selected artifact and current complete build", async () => {
    const context = createContractTestContext(root);
    const temporaryRoot = mkdtempSync(path.join(tmpdir(), "cartulary-browser-artifact-"));
    try {
      const production = resolveBrowserFrontendArtifact(root, "webserver-backed");
      const measurement = resolveBrowserFrontendArtifact(root, "measurement");
      for (const [rowID, expectedProducer] of [
        ["module.auth.browser.browser_sign_in_exposes_the_ordinary_session_sur_2ec8df6229", "build-web"],
        ["web.design.accessibility.account_menu_accessible_keyboard_and_focus_state_3c31f28451", "build-web"],
        ["web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde", "build-web"],
        ["module.networkflow.measurement.saved_graph_dom_ceiling", "build-web-measurement"],
      ]) {
        const graph = context.compiler.compileRows([rowID]);
        assert.ok(graph.units.some((unit) => unit.unit_id === `operation:harness.${expectedProducer.replaceAll("-", "_")}`), `row ${rowID} must build ${expectedProducer}`);
        // Measurement still builds the backend's embedded production assets.
        // Production work must never acquire the measurement producer.
        if (expectedProducer === "build-web")
          assert.ok(!graph.units.some((unit) => unit.unit_id === "operation:harness.build_web_measurement"), `production row ${rowID} must not build measurement output`);
      }
      assert.notEqual(production.producer_report_key, measurement.producer_report_key);
      assert.notEqual(production.path, measurement.path);
      for (const artifact of [production, measurement]) {
        const runID = artifact.producer_report_key;
        const runRoot = path.join(temporaryRoot, runID);
        mkdirSync(runRoot, { mode: 0o700 });
        writeFileSync(path.join(runRoot, "run-manifest.json"), JSON.stringify({ source_digest: `sha256:${"1".repeat(64)}`, toolchain_digest: `sha256:${"2".repeat(64)}` }));
        const runtime = createRuntime({ repoRoot: root, runRoot, runID, scratchRoot: path.join(temporaryRoot, "private") });
        try {
          const environment = { CARTULARY_TEST_RESULTS_DIR: temporaryRoot, CARTULARY_TEST_RUN_ID: runID, CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID };
          const staging = runtime.privatePath("staging");
          mkdirSync(staging, { mode: 0o700 });
          for (const entry of artifact.entries)
            writeFileSync(path.join(staging, entry), entry);
          const profile = { ...artifact, id: artifact === production ? "production" : "measurement" };
          const directory = sealFrontendArtifact({ repoRoot: root, runRoot, runtime, profile, staging });
          const sealed = resolveFrontendArtifact(root, artifact.producer_report_key, environment);
          const frontend = { build_artifact_ref: sealed.receiptRef, build_receipt_sha256: sealed.receiptDigest, build_artifact_sha256: sealed.receipt.content_digest };
          validateFrontendAttachment(root, artifact, frontend, environment);
          const other = artifact === production ? measurement : production;
          assert.throws(() => validateFrontendAttachment(root, other, frontend, environment), /does not match selected stage/u);
          assert.throws(() => validateFrontendAttachment(root, artifact, { ...frontend, build_artifact_sha256: `sha256:${"0".repeat(64)}` }, environment), /digest mismatch/u);
          writeFileSync(path.join(directory, "index.html"), "changed after publication");
          assert.throws(() => validateFrontendAttachment(root, artifact, frontend, environment), /digest mismatch/u);
          rmSync(path.join(directory, artifact.entries.at(-1)));
          assert.throws(() => validateFrontendAttachment(root, artifact, frontend, environment));
        }
        finally {
          runtime.close();
        }
      }
    }
    finally {
      rmSync(temporaryRoot, { recursive: true, force: true });
    }
  });
}
