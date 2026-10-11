import assert from "node:assert/strict";
import { existsSync, mkdirSync, readFileSync, chmodSync, mkdtempSync, rmSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import { get } from "node:https";
import { admittedBrowserTrust } from "../browser-trust.mjs";
import { runPreparedReview, recoverReviewPreparation } from "../review-preparation-native.mjs";
import { createSuiteRuntime } from "../../runtime/suite-runtime.mjs";
import { recordRuntimeResource, runtimeRecoveryResources } from "../../runtime/resource-recovery.mjs";
import { buildWorkGraph, runWorkGraph } from "../../scheduler/work-graph/index.mjs";
import { cleanupErrors } from "../../scheduler/fixture-broker/cleanup-lifecycle.mjs";
import { repoRoot } from "../ui-review/policy.mjs";
import { workspaceLayout } from "../../../workspace_layout.generated.mjs";
import { createFixtureSession } from "../../runtime/fixture-session.mjs";
import { createLocalSession, localSessionStatus, stopLocalSession } from "../../services/local-session.mjs";

async function realBorrowedSuiteCase() {
  const scratch = mkdtempSync(path.join(tmpdir(), "cartulary-native-borrowed-"));
  const sessionFile = path.join(scratch, "owner", "session.json");
  const binary = path.join(repoRoot, workspaceLayout.toolbin, "cartulary-test-services");
  const owner = { root: repoRoot, binary, sessionFile };
  const runID = `borrowed-fixture-${Date.now()}`;
  const runRoot = path.join(repoRoot, ".cartulary/test-results", runID);
  mkdirSync(runRoot, { mode: 0o700 });
  const runtime = createSuiteRuntime({ repoRoot, runRoot, runID });
  const client = createFixtureSession({ root: repoRoot, runRoot, runtime,
    target: "service-backed-test-slice", environment: { ...process.env,
      NODE_BIN: process.execPath, CARTULARY_TEST_SERVICES_BIN: binary,
      CARTULARY_TEST_SERVICES_MODE: "attach", CARTULARY_TEST_SERVICES_SESSION_FILE: sessionFile,
    } });
  try {
    createLocalSession({ ...owner, environment: {} });
    const before = localSessionStatus(owner);
    assert.equal(before.state, "active");
    const lease = await client.acquire("postgres_dedicated", { unitID: "fixture:borrowed" });
    assert.equal(lease.record.ownership, "borrowed");
    assert.equal(localSessionStatus(owner).borrower_count, 1);
    assert.throws(() => stopLocalSession(owner), /live borrower/u);
    await lease.release();
    await client.closeSession();
    await client.closeSession();
    const after = localSessionStatus(owner);
    assert.equal(after.state, "active", "native close preserves the physical borrowed suite");
    assert.equal(after.borrower_count, 0);
    assert.deepEqual(after.services, before.services);
    assert.equal(runtimeRecoveryResources(runtime).length, 0);
  } finally {
    // The native borrower drains before the distinct physical owner may stop.
    await client.closeSession();
    stopLocalSession(owner);
    runtime.close();
    rmSync(scratch, { recursive: true, force: true });
  }
}

// Exercise the production stack owner, not a duplicate teardown in a mock.
// Inject only work outcome and an actual inability to release the live stack.
export async function realStackCleanupCases() {
  await realBorrowedSuiteCase();
  for (const cleanupFailed of [false, true]) {
    const runID = `fixture-cleanup-${Date.now()}-${cleanupFailed ? "unresolved" : "product"}`;
    const runRoot = path.join(repoRoot, ".cartulary/test-results", runID);
    mkdirSync(runRoot, { mode: 0o700 });
    const runtime = createSuiteRuntime({ repoRoot, runRoot, runID });
    let result, origin, primary, probe, blockedProof;
    try {
      const operation = runPreparedReview({ environment: { ...process.env, REVIEW_PROFILE: "default" },
        runID, runRoot, runtime, retainDetail: false, writeOutput: () => {}, hold: async () => {},
        onReady: async ({ attached, fixtureLease, fixtureBroker }) => {
          origin = attached.CARTULARY_WEB_E2E_PUBLIC_ORIGIN;
          const ca = admittedBrowserTrust(attached).bytes;
          probe = () => new Promise((resolve, reject) => {
            const request = get(origin, { ca, minVersion: "TLSv1.3", maxVersion: "TLSv1.3", agent: false, signal: AbortSignal.timeout(3000) }, (response) => { response.resume(); response.once("end", () => resolve(response.statusCode)); response.once("error", reject); });
            request.once("error", reject);
          });
          assert.equal(await probe(), 200);
          if (cleanupFailed) {
            const resource=runtimeRecoveryResources(runtime).find((value)=>value.kind==="browser_stack");
            assert.ok(resource,"live stack has durable ownership proof");
            blockedProof=resource.target;
            // Withhold safe access to the exact ownership proof. Native cleanup
            // must refuse it and preserve the live dependency for exact recovery.
            chmodSync(blockedProof,0o644);
          }
          const unit = { unit_id: "fixture:failed-work", owner_id: "harness.browser", kind: "runner",
            command: { executable: "true", args: [], environment: { CARTULARY_BROWSER_RELEASE_AFFINITY: "1" } },
            needs: [], resource_claims: { cpu: 1 }, fixture_lease: "browser_stack", service_dependencies: [],
            cache_policy: "none", timeout_ms: 1000, current_run_evidence_outputs: [], estimated_work_ms: 1,
            failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" } };
          result = await runWorkGraph({ graph: buildWorkGraph([unit]), capacities: new Map([["cpu", 1]]),
            cwd: repoRoot, environment: {}, fixtureBroker: {
              acquire: async () => fixtureLease, close: () => fixtureBroker.close(),
              hasUnresolvedCleanup: () => fixtureBroker.hasUnresolvedCleanup(), cleanupResults: fixtureBroker.cleanupResults,
            }, executeUnit: async () => ({ status: "failed", failure_class: "product", failure_reason: "test_assertion_failure", exit_code: 10 }),
          });
        },
      });
      if (cleanupFailed) await assert.rejects(operation, (error) => error.failure_reason === "artifact_error" && cleanupErrors(error).some((failure)=>failure.failure_reason==="artifact_error"));
      else await operation;
      assert.equal(result.unit_results["fixture:failed-work"].failure_reason, "test_assertion_failure");
      assert.equal(result.events.at(-1).failure_class, "product");
      const receipt = JSON.parse(readFileSync(path.join(runRoot, "cleanup-results.json"), "utf8"));
      const releases = receipt.results.filter((step) => step.operation === "fixture_release");
      assert.equal(releases.length, 1);
      assert.equal(releases[0].outcome, cleanupFailed ? "failed" : "completed");
      assert.equal(releases[0].failure_reason, cleanupFailed ? "artifact_error" : null);
      if (cleanupFailed) {
        assert.equal(await probe(), 200, "genuine failure leaves a live owned stack");
        chmodSync(blockedProof,0o600); blockedProof=null;
        runtime.preserveRecovery();
        const resources = runtimeRecoveryResources(runtime);
        assert.deepEqual(resources.map((resource) => resource.kind).sort(), ["browser_stack", "managed_suite"]);
        for (const resource of resources) assert.ok(existsSync(resource.target));
        assert.equal(existsSync(runtime.privatePath("review-access")), false, "closed login consumers leave no credential detail");
        await assert.rejects(recoverReviewPreparation({ runtime: { ...runtime, runID: "wrong-recovery-run" }, resources }), (error) => error.failure_reason === "artifact_error");
        await recoverReviewPreparation({ runtime, resources,
          environment: { ...process.env, CARTULARY_TEST_SUITE_ID: "unrelated-ambient-suite", CARTULARY_TEST_SERVICES_CALL_MODE: "none" },
          onReleased: (resource) => recordRuntimeResource(runtime, { ...resource, state: "released" }) });
      }
      assert.equal(runtimeRecoveryResources(runtime).length, 0);
      await assert.rejects(probe(), "the owned listener is gone after cleanup or exact recovery");
    } catch (error) { primary = error; }
    finally {
      try {
        if(blockedProof) chmodSync(blockedProof,0o600);
        const resources = runtimeRecoveryResources(runtime);
        if (resources.length) await recoverReviewPreparation({ runtime, resources,
          onReleased: (resource) => recordRuntimeResource(runtime, { ...resource, state: "released" }) });
        runtime.close();
      } catch (error) {
        if (primary) (primary.cleanupFailures ??= []).push(error); else primary = error;
        try { runtime.preserveRecovery(); } catch (failure) { (primary.cleanupFailures ??= []).push(failure); }
      }
    }
    if (primary) throw primary;
    assert.equal(existsSync(runtime.root), false);
  }
}
