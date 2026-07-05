import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  validateResultRoot,
  validateRunId,
} from "./harness-contract.mjs";

export const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");

export const phaseSummarySchemaID = "cartulary.test_phase_summary.v3";
export const targetTimingSchemaID = "cartulary.test_target_timing.v1";
export const targetSummarySchemaID = "cartulary.test_target_summary.v4";
export const runSummarySchemaID = "cartulary.test_run_summary.v6";
export const sharedExecutionGroupSchemaID = "cartulary.test_shared_execution_group.v1";
export const testAccountingClassificationSchemaID = "cartulary.test_accounting_classification.v2";
export const frontendRowAccountingSchemaID =
  "cartulary.frontend_row_accounting.v4";
export const vitestFailureDetailsSchemaID =
  "cartulary.vitest_failure_details.v1";

export const testCoverageBuckets = [
  "authoritative",
  "support",
  "raw",
  "tooling_support",
  "unowned_regression",
  "unmapped",
];
export const testCoverageBucketSet = new Set(testCoverageBuckets);

export const timingBucketOrder = [
  "setup",
  "service_wait",
  "migration",
  "server_startup",
  "frontend_startup",
  "test_command",
  "teardown",
  "report_collation",
];
export const timingBucketSet = new Set(timingBucketOrder);

export const validPhaseCountingModes = new Set(["counted", "none"]);

export function resolveResultsRoot(env = process.env) {
  return validateResultRoot(env.CARTULARY_TEST_RESULTS_DIR, { root: repoRoot });
}

export function resolveRunId(env = process.env) {
  if (Object.hasOwn(env, "CARTULARY_TEST_RUN_ID")) {
    return validateRunId(env.CARTULARY_TEST_RUN_ID);
  }
  return "adhoc";
}

export function createTestOutputContext(env = process.env) {
  return {
    repoRoot,
    resultsRoot: resolveResultsRoot(env),
    runId: resolveRunId(env),
    env,
  };
}
