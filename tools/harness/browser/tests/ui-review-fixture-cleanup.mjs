import assert from "node:assert/strict";
import { existsSync, mkdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { runPreparedReview, recoverReviewPreparation } from "../review-preparation.mjs";
import { createSuiteRuntime } from "../../runtime/suite-runtime.mjs";
import { recordRuntimeResource, runtimeRecoveryResources } from "../../runtime/resource-recovery.mjs";
import { buildWorkGraph, runWorkGraph } from "../../scheduler/work-graph/index.mjs";
import { repoRoot } from "../ui-review/policy.mjs";

// Exercise the production stack owner, not a duplicate teardown in a mock.
// Inject only work outcome and an actual inability to release the live stack.
export async function realStackCleanupCases() {
  for (const cleanupFailed of [false, true]) {
    const runID = `fixture-cleanup-${Date.now()}-${cleanupFailed ? "unresolved" : "product"}`;
    const runRoot = path.join(repoRoot, ".cartulary/test-results", runID);
    mkdirSync(runRoot, { mode: 0o700 });
    const runtime = createSuiteRuntime({ repoRoot, runRoot, runID });
    let result, origin;
    try {
      const operation = runPreparedReview({ environment: { ...process.env, REVIEW_PROFILE: "default" },
        runID, runRoot, runtime, retainDetail: false, writeOutput: () => {}, hold: async () => {},
        onOwnedResource: (resource) => recordRuntimeResource(runtime, resource),
        onReady: async ({ attached, fixtureLease, fixtureBroker }) => {
          origin = attached.CARTULARY_WEB_E2E_PUBLIC_ORIGIN;
          assert.equal((await fetch(origin)).status, 200);
          if (cleanupFailed) fixtureLease.entry.allocation.release = () => { throw new Error("controlled live-stack cleanup failure"); };
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
      if (cleanupFailed) await assert.rejects(operation, (error) => error.failure_reason === "cleanup_error");
      else await operation;
      assert.equal(result.unit_results["fixture:failed-work"].failure_reason, "test_assertion_failure");
      assert.equal(result.events.at(-1).failure_class, "product");
      const receipt = JSON.parse(readFileSync(path.join(runRoot, "cleanup-results.json"), "utf8"));
      const releases = receipt.results.filter((step) => step.operation === "fixture_release");
      assert.equal(releases.length, 1);
      assert.equal(releases[0].outcome, cleanupFailed ? "failed" : "completed");
      assert.equal(releases[0].failure_reason, cleanupFailed ? "cleanup_error" : null);
      if (cleanupFailed) {
        assert.equal((await fetch(origin)).status, 200, "genuine failure leaves a live owned stack");
        runtime.preserveRecovery();
        const resources = runtimeRecoveryResources(runtime);
        assert.deepEqual(resources.map((resource) => resource.kind).sort(), ["browser_stack", "managed_suite"]);
        for (const resource of resources) assert.ok(existsSync(resource.target));
        assert.equal(existsSync(runtime.privatePath("review-access")), false, "closed login consumers leave no credential detail");
        await recoverReviewPreparation({ runtime, resources,
          onReleased: (resource) => recordRuntimeResource(runtime, { ...resource, state: "released" }) });
      }
      assert.equal(runtimeRecoveryResources(runtime).length, 0);
      await assert.rejects(fetch(origin), "the owned listener is gone after cleanup or exact recovery");
    } finally {
      const resources = runtimeRecoveryResources(runtime);
      if (resources.length) await recoverReviewPreparation({ runtime, resources,
        onReleased: (resource) => recordRuntimeResource(runtime, { ...resource, state: "released" }) });
      runtime.close();
    }
    assert.equal(existsSync(runtime.root), false);
  }
}
