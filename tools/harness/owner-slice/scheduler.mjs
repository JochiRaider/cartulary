import { existsSync, mkdirSync, readFileSync } from "node:fs";
import path from "node:path";

import { loadExecutionTopology } from "../generated-artifacts/execution-topology.mjs";
import {
  runtimeBinaryAbsoluteEnvForIDs,
  runtimeBinaryProducerTargetsForIDs,
} from "../runtime-binary-registry.mjs";
import { runNormalizedSchedule } from "../scheduler/scheduler-runner.mjs";
import {
  resolveSchedulerResourceLimits,
  schedulerCapacityProfileLimits,
  testSliceDefaultCapacityProfile,
} from "../scheduler/scheduler-resource-policy.mjs";

const eventSchemaID = "cartulary.scheduler_event.v7";
const summarySchemaID = "cartulary.service_backed_scheduler_summary.v10";
const ownerSliceOnlyInputs = Object.freeze([
  "OWNER",
  "ROWS",
  "JSON",
  "PLAYWRIGHT_WORKERS",
  "VITEST_MAX_WORKERS",
  "MAKEFLAGS",
  "MFLAGS",
  "GNUMAKEFLAGS",
  "MAKEOVERRIDES",
  "CARTULARY_MAKE_INPUT_SOURCES",
  "CARTULARY_TEST_CATALOG_ROW_IDS",
  "CARTULARY_TEST_OWNER",
  "CARTULARY_TEST_RESULTS_DIR",
  "CARTULARY_TEST_RUN_ID",
  "CARTULARY_TEST_TARGET",
  "CARTULARY_HARNESS_IDENTITY_PREPARED",
]);

export function ownerSliceChildEnvironment(
  environment = process.env,
  { childResultsRoot = "" } = {},
) {
  const child = { ...environment };
  for (const name of ownerSliceOnlyInputs) delete child[name];
  if (childResultsRoot !== "") {
    child.CARTULARY_TEST_RESULTS_DIR = childResultsRoot;
  }
  return child;
}

function terminalStateForStatus(status, skipReason = "") {
  if (skipReason === "dependency_failure") return "skipped_dependency";
  if (skipReason === "cancelled_or_interrupted" || status === 130 || status === 143) return "cancelled";
  return status === 10 ? "failed" : "infrastructure_failed";
}

function failureReasonForStatus(status, skipReason = "") {
  if (skipReason === "dependency_failure") return "dependency_failure";
  if (skipReason === "cancelled_or_interrupted" || status === 130 || status === 143) {
    return "cancelled_or_interrupted";
  }
  if (status === 13) return "timeout_failure";
  if (status === 10) return "test_assertion_failure";
  return "scheduler_accounting_error";
}

function normalizeUnitResult(unit, raw, completed, skipped) {
  if (raw) {
    if (!Array.isArray(raw.unit_results) || raw.unit_results.length !== 1) {
      throw new Error(`${unit.work_unit_id} result must contain exactly one unit result`);
    }
    const result = raw.unit_results[0];
    if (result.work_unit_id !== unit.work_unit_id) {
      throw new Error(`${unit.work_unit_id} result identity mismatch`);
    }
    const observed = [...result.row_results].map((row) => row.row_id).sort();
    if (JSON.stringify(observed) !== JSON.stringify([...unit.row_ids].sort())) {
      throw new Error(`${unit.work_unit_id} result row inventory mismatch`);
    }
    return result;
  }
  const status = completed?.status ?? 11;
  const skipReason = skipped?.reason ?? "";
  const terminalState = terminalStateForStatus(status, skipReason);
  const exitCode = terminalState === "skipped_dependency" ? 11 : status;
  return {
    work_unit_id: unit.work_unit_id,
    runner: unit.runner,
    target_name: unit.target_name,
    row_ids: [...unit.row_ids],
    status: terminalState,
    exit_code: exitCode,
    duration_ms: completed?.duration_ms ?? 0,
    failure_reason: failureReasonForStatus(status, skipReason),
    row_results: unit.row_ids.map((rowID) => ({
      row_id: rowID,
      terminal_state: terminalState,
      duration_ms: 0,
      exit_code: exitCode,
      failure_reason: failureReasonForStatus(status, skipReason),
      attempt: 1,
    })),
    stdout: "",
    stderr: "",
  };
}

export function buildOwnerSliceWorkUnits(root, plan, planPath, targetDir) {
  const node = process.env.NODE_BIN || process.execPath;
  const runner = path.join(root, "tools", "harness", "owner-slice", "owner-unit-runner-cli.mjs");
  const resultDir = path.join(targetDir, "work-unit-results");
  const artifactRoot = path.join(targetDir, "work-unit-artifacts");
  mkdirSync(resultDir, { recursive: true });
  mkdirSync(artifactRoot, { recursive: true });
  const browserReadinessID = "readiness.owner_browser_runtime";
  const needsBrowserReadiness = plan.work_units.some((unit) => unit.runner === "playwright");
  const serviceReadinessID = "readiness.owner_service_runtime";
  const needsServiceReadiness = plan.work_units.some(
    (unit) => unit.runner === "playwright" || unit.managed_service_ids.length > 0,
  );
  const childEnv = ownerSliceChildEnvironment(process.env, {
    childResultsRoot: path.join(targetDir, "child-results"),
  });
  const readinessEnv = { ...childEnv, CARTULARY_SUPPRESS_CHILD_SUCCESS: "1" };
  const units = [];
  const topology = loadExecutionTopology({ root });
  const runtimeBinaryIDs = [...new Set(plan.work_units.flatMap((unit) =>
    unit.runtime_binary_ids,
  ))].sort();
  const runtimeReadinessID = "readiness.owner_runtime_binaries";
  if (runtimeBinaryIDs.length > 0) {
    const producerTargets = runtimeBinaryProducerTargetsForIDs(
      topology.runtimeBinaries,
      runtimeBinaryIDs,
      "owner slice runtime binaries",
    );
    units.push({
      id: runtimeReadinessID,
      label: "owner runtime binary readiness",
      kind: "readiness",
      class: "build_artifact",
      target: "owner-runtime-binary-readiness",
      aggregateTarget: plan.target,
      needs: [],
      completionKeys: [runtimeReadinessID],
      failureKeys: [runtimeReadinessID],
      resourceClaims: new Map([["go_cpu", 1], ["go_io", 1]]),
      priority: 1,
      weightMs: 1,
      order: units.length,
      timeoutMs: 600_000,
      countInTotal: false,
      command: {
        command: process.env.MAKE || "make",
        args: ["--silent", "--no-print-directory", ...producerTargets],
        env: readinessEnv,
      },
    });
  }
  if (needsServiceReadiness) {
    units.push({
      id: serviceReadinessID,
      label: "owner service runtime readiness",
      kind: "readiness",
      class: "build_artifact",
      target: "owner-service-runtime-readiness",
      aggregateTarget: plan.target,
      needs: [],
      completionKeys: [serviceReadinessID],
      failureKeys: [serviceReadinessID],
      resourceClaims: new Map([["process", 1]]),
      priority: 1,
      weightMs: 1,
      order: units.length,
      timeoutMs: 600_000,
      countInTotal: false,
      command: {
        command: process.env.MAKE || "make",
        args: ["--silent", "--no-print-directory", "test-service-images"],
        env: readinessEnv,
      },
    });
  }
  if (needsBrowserReadiness) {
    units.push({
      id: browserReadinessID,
      label: "owner browser runtime readiness",
      kind: "readiness",
      class: "build_artifact",
      target: "owner-browser-runtime-readiness",
      aggregateTarget: plan.target,
      needs: [],
      completionKeys: [browserReadinessID],
      failureKeys: [browserReadinessID],
      resourceClaims: new Map([["process", 1]]),
      priority: 1,
      weightMs: 1,
      order: units.length,
      timeoutMs: 600_000,
      countInTotal: false,
      command: {
        command: process.env.MAKE || "make",
        args: [
          "--silent",
          "--no-print-directory",
          "build-web",
          "build-server-harness",
          "build-migrate",
        ],
        env: readinessEnv,
      },
    });
  }
  units.push(...plan.work_units.map((unit, planOrder) => ({
    id: unit.work_unit_id,
    label: unit.work_unit_id,
    kind: "test_slice_work_unit",
    class: unit.runner,
    target: unit.target_name,
    aggregateTarget: plan.target,
    needs: [
      ...unit.dependencies,
      ...(unit.runner === "playwright" ? [browserReadinessID] : []),
      ...(unit.runner === "playwright" || unit.managed_service_ids.length > 0
        ? [serviceReadinessID]
        : []),
      ...(unit.runtime_binary_ids.length > 0 ? [runtimeReadinessID] : []),
    ],
    completionKeys: [unit.work_unit_id],
    failureKeys: [unit.work_unit_id],
    resourceClaims: new Map(Object.entries(unit.resource_claims)),
    priority: 0,
    weightMs: 1,
    order: planOrder + units.length,
    timeoutMs: unit.timeout_seconds * 1_000,
    command: {
      command: node,
      args: [
        runner,
        "--plan", planPath,
        "--unit", unit.work_unit_id,
        "--result", path.join(resultDir, `${unit.work_unit_id}.json`),
        "--artifact-root", artifactRoot,
      ],
      env: {
        ...childEnv,
        CARTULARY_TEST_OWNER: plan.owner_id,
        CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        ...runtimeBinaryAbsoluteEnvForIDs(
          topology.runtimeBinaries,
          unit.runtime_binary_ids,
          { repoRoot: root, label: `${unit.work_unit_id} runtime binaries` },
        ),
        ...(unit.runner === "playwright" ? {
          PATH: `${path.join(root, "tmp", "node-runtime", "bin")}:${process.env.PATH ?? ""}`,
          NODE_RUNTIME_DIR: path.join(root, "tmp", "node-runtime"),
          NODE_BIN: node,
          PNPM: process.env.PNPM || path.join(root, "tmp", "node-runtime", "bin", "pnpm"),
          CARTULARY_SERVER_HARNESS_BIN: path.join(root, "server-harness"),
          CARTULARY_MIGRATE_BIN: path.join(root, "migrate"),
          CARTULARY_TEST_SERVICES_BIN:
            process.env.TEST_SERVICES_BIN || path.join(root, "tmp", "toolbin", "cartulary-test-services"),
        } : {}),
      },
    },
  })));
  const cleanup = path.join(root, "tools", "harness", "owner-slice", "owner-slice-finalizer-cli.mjs");
  units.push({
    id: plan.finalizers[0].finalizer_id,
    label: plan.finalizers[0].finalizer_id,
    kind: "finalizer",
    class: "cleanup",
    target: plan.target,
    aggregateTarget: plan.target,
    needs: [...plan.finalizers[0].dependencies],
    completionKeys: [plan.finalizers[0].finalizer_id],
    failureKeys: [plan.finalizers[0].finalizer_id],
    resourceClaims: new Map(),
    priority: 0,
    weightMs: 1,
    order: units.length,
    timeoutMs: 30_000,
    countInTotal: false,
    command: {
      command: node,
      args: [cleanup, "--output", path.join(targetDir, "owner-slice-cleanup.json")],
      env: childEnv,
    },
  });
  return units;
}

function genericSchedule(plan, workUnits) {
  const initial = schedulerCapacityProfileLimits(
    "test_slice",
    testSliceDefaultCapacityProfile,
    `${plan.target} owner scheduler`,
  );
  const resolved = resolveSchedulerResourceLimits({
    scheduler: "test_slice",
    resourceLimits: initial.limits,
    resourceLimitSources: initial.sources,
    label: `${plan.target} owner scheduler`,
    workUnits,
    pruneToClaims: true,
  });
  return {
    target: plan.target,
    kind: "test_slice",
    prefix: "TEST-SLICE-SCHEDULER",
    eventSchemaID,
    summarySchemaID,
    resourceScheduler: "test_slice",
    stopOnFirstFailure: false,
    showFinalizing: true,
    summaryTotalWallTime: true,
    validateSummaryTiming: false,
    resourceLimits: resolved.resourceLimits,
    resourceLimitSources: resolved.resourceLimitSources,
    workUnits,
    totalWorkUnits: plan.work_units.length,
    finalizerCount: plan.finalizers.length,
    shouldReplayLog: ({ result }) => result.status !== 0,
    summaryExtra: () => ({
      extensions: {
        "cartulary.test_slice.scheduler.v1": {
          owner_id: plan.owner_id,
          plan_semantic_digest: plan.plan_semantic_digest,
          scheduler_semantic_digest: plan.scheduler_semantic_digest,
        },
      },
    }),
  };
}

export async function executeOwnerSliceSchedule(root, plan, planPath, targetDir) {
  const workUnits = buildOwnerSliceWorkUnits(root, plan, planPath, targetDir);
  const lifecycle = await runNormalizedSchedule({
    repoRoot: root,
    schedule: genericSchedule(plan, workUnits),
    testOutputScript: process.env.TEST_OUTPUT_SCRIPT || path.join(root, "tools", "harness", "output", "test-output.mjs"),
  });
  const completed = new Map(lifecycle.reporter.completedWork.map((entry) => [entry.id, entry]));
  const skipped = new Map(lifecycle.reporter.skippedWork.map((entry) => [entry.id, entry]));
  const unitResults = plan.work_units.map((unit) => {
    const resultPath = path.join(targetDir, "work-unit-results", `${unit.work_unit_id}.json`);
    const raw = existsSync(resultPath) ? JSON.parse(readFileSync(resultPath, "utf8")) : null;
    return normalizeUnitResult(unit, raw, completed.get(unit.work_unit_id), skipped.get(unit.work_unit_id));
  });
  const finalizerRecord = completed.get(plan.finalizers[0].finalizer_id);
  const finalizerFailed = !finalizerRecord || finalizerRecord.status !== 0;
  return {
    duration_ms: Math.round(lifecycle.summary.scheduler_total_duration_ms),
    status:
      lifecycle.status === 0 && !finalizerFailed && unitResults.every((entry) => entry.status === "passed")
        ? "pass"
        : "fail",
    exit_code: lifecycle.status || (finalizerFailed ? 12 : 0),
    unit_results: unitResults,
    row_results: unitResults.flatMap((entry) => entry.row_results),
    lifecycle_summary: lifecycle.summary,
    finalizer_failed: finalizerFailed,
  };
}
