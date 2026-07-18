#!/usr/bin/env node

import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { publicExitCodeForSummary } from "../contract/index.mjs";
import { semanticJSONDigest } from "../test-catalog/semantic-json.mjs";
import { resolveBrowserBatchStage } from "./browser-batch-manifest.mjs";
import { runNormalizedSchedule } from "../scheduler/scheduler-runner.mjs";
import {
  resolveSchedulerResourceLimits,
  schedulerCapacityProfileLimits,
  testSliceDefaultCapacityProfile,
} from "../scheduler/scheduler-resource-policy.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "../../..");

function usage() {
  return "usage: browser-stage-scheduler-cli.mjs <stage> [--mode snapshot_update]";
}

const validationMode = "validation";
const snapshotUpdateMode = "snapshot_update";

function parseArgs(argv) {
  if (argv.length === 1) return { stageName: argv[0], mode: validationMode };
  if (argv.length === 3 && argv[1] === "--mode" && argv[2] === snapshotUpdateMode) {
    return { stageName: argv[0], mode: snapshotUpdateMode };
  }
  throw new Error(usage());
}

function sessions(stage) {
  const grouped = new Map();
  for (const group of stage.groups) {
    const key = `${group.browserSessionGroup}\0${group.runtimeProfileID}`;
    const entry = grouped.get(key) ?? {
      id: group.browserSessionGroup,
      runtimeProfileID: group.runtimeProfileID,
      groupIDs: [],
    };
    entry.groupIDs.push(group.name);
    grouped.set(key, entry);
  }
  return [...grouped.values()]
    .map((entry) => ({ ...entry, groupIDs: entry.groupIDs.sort() }))
    .sort((left, right) =>
      left.id.localeCompare(right.id) || left.runtimeProfileID.localeCompare(right.runtimeProfileID),
    );
}

export function buildBrowserStageSchedule(stage, manifestPath, options = {}) {
  const mode = options.mode ?? validationMode;
  if (!new Set([validationMode, snapshotUpdateMode]).has(mode)) {
    throw new Error(`unsupported browser scheduler mode ${mode}`);
  }
  if (mode === snapshotUpdateMode && stage.name !== "visual") {
    throw new Error("snapshot_update mode is limited to the visual browser stage");
  }
  const target = mode === snapshotUpdateMode
    ? (options.target ?? "browser-e2e-visual-update")
    : stage.target;
  const stageResource = `browser_stage_${target.replaceAll(/[^a-z0-9_]/gu, "_")}`;
  const sessionUnits = sessions(stage).map((session, order) => ({
    id: `browser_session:${session.id}:${session.runtimeProfileID}`,
    label: `browser session ${session.id} (${session.runtimeProfileID})`,
    kind: "browser_stage_session",
    class: "browser",
    target,
    aggregateTarget: target,
    needs: [],
    completionKeys: [`browser_session:${session.id}:${session.runtimeProfileID}`],
    failureKeys: [`browser_session:${session.id}:${session.runtimeProfileID}`],
    resourceClaims: new Map([["browser_stack", 1], [stageResource, 1], ["process", 1]]),
    priority: 0,
    weightMs: 1,
    order,
    timeoutMs: 1_800_000,
    command: {
      command: path.join(root, "tools", "harness", "browser", "start-web-e2e.sh"),
      args: [
        "--",
        path.join(root, "tools", "harness", "browser", "run-browser-e2e-batch.sh"),
        stage.name,
        "--defer-summary",
      ],
      env: {
        ...process.env,
        BROWSER_E2E_BATCH_MANIFEST: manifestPath,
        CARTULARY_TEST_TARGET: target,
        CARTULARY_BROWSER_SESSION_GROUP: session.id,
        CARTULARY_BROWSER_RUNTIME_PROFILE_ID: session.runtimeProfileID,
        CARTULARY_BROWSER_SELECTED_GROUPS: session.groupIDs.join(","),
        CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        ...(mode === snapshotUpdateMode
          ? {
              CARTULARY_BROWSER_MAINTENANCE_MODE: snapshotUpdateMode,
              CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS: "1",
            }
          : {}),
      },
    },
  }));
  const sessionIDs = sessionUnits.map((unit) => unit.id);
  const evidenceTargets = mode === validationMode
    ? [...new Set(stage.groups.map((group) => group.target))].sort()
    : [];
  const node = process.env.NODE_BIN || process.execPath;
  const evidenceUnits = evidenceTargets.map((target, index) => ({
    id: `browser_evidence:${target}`,
    label: `browser evidence ${target}`,
    kind: "finalizer",
    class: "artifact",
    target,
        aggregateTarget: target,
    needs: sessionIDs,
    completionKeys: [`browser_evidence:${target}`],
    failureKeys: [`browser_evidence:${target}`],
    resourceClaims: new Map(),
    priority: 0,
    weightMs: 1,
    order: sessionUnits.length + index,
    timeoutMs: 60_000,
    countInTotal: false,
    command: {
      command: node,
      args: [path.join(root, "tools", "harness", "browser", "browser-evidence-finalize-cli.mjs"), target],
      env: { ...process.env, CARTULARY_TEST_TARGET: target },
    },
  }));
  const targetFinalizer = {
    id: `browser_target_summary:${target}`,
    label: `browser target summary ${target}`,
    kind: "finalizer",
    class: "artifact",
    target,
    aggregateTarget: target,
    needs: evidenceUnits.map((unit) => unit.id),
    completionKeys: [`browser_target_summary:${stage.target}`],
    failureKeys: [`browser_target_summary:${stage.target}`],
    resourceClaims: new Map(),
    priority: 0,
    weightMs: 1,
    order: sessionUnits.length + evidenceUnits.length,
    timeoutMs: 60_000,
    countInTotal: false,
    command: {
      command: node,
      args: [
        path.join(root, "tools", "harness", "browser", "browser-target-finalizer-cli.mjs"),
        "--target", target,
        "--groups", stage.groups.map((group) => group.name).sort().join(","),
        "--group-targets", stage.groups
          .map((group) => `${group.name}=${group.target}`)
          .sort()
          .join(","),
        "--children", stage.summaryChildren.join(","),
      ],
      env: process.env,
    },
  };
  const finalizerUnits = mode === validationMode ? [...evidenceUnits, targetFinalizer] : [];
  const workUnits = [...sessionUnits, ...finalizerUnits];
  const initial = schedulerCapacityProfileLimits(
    "test_slice",
    testSliceDefaultCapacityProfile,
    `${target} browser scheduler`,
  );
  initial.limits.set(stageResource, 1);
  initial.sources.set(stageResource, "generated:browser_stage_serialization");
  const resolved = resolveSchedulerResourceLimits({
    scheduler: "test_slice",
    resourceLimits: initial.limits,
    resourceLimitSources: initial.sources,
    label: `${target} browser scheduler`,
    workUnits,
    pruneToClaims: true,
  });
  const schedulerSemanticDigest = semanticJSONDigest({
    scheduler_kind: "test_slice",
    target,
    stage: stage.name,
    mode,
    sessions: sessions(stage),
    groups: stage.groups.map((group) => ({
      id: group.name,
      target: group.target,
      runtime_profile_id: group.runtimeProfileID,
      browser_session_group: group.browserSessionGroup,
      selected_rows: group.selectedRowIDs,
    })),
    evidence_targets: evidenceTargets,
    resource_limits: Object.fromEntries(resolved.resourceLimits),
  });
  return {
    target,
    kind: "test_slice",
    prefix: "BROWSER-SCHEDULER",
    eventSchemaID: "cartulary.scheduler_event.v6",
    summarySchemaID: "cartulary.service_backed_scheduler_summary.v10",
    resourceScheduler: "test_slice",
    stopOnFirstFailure: false,
    showFinalizing: true,
    summaryTotalWallTime: true,
    validateSummaryTiming: false,
    resourceLimits: resolved.resourceLimits,
    resourceLimitSources: resolved.resourceLimitSources,
    workUnits,
    totalWorkUnits: sessionUnits.length,
    finalizerCount: finalizerUnits.length,
    shouldReplayLog: ({ result }) => result.status !== 0,
    summaryExtra: () => ({
      extensions: {
        "cartulary.test_slice.browser_scheduler.v1": {
          stage_id: stage.name,
          mode,
          scheduler_semantic_digest: schedulerSemanticDigest,
        },
      },
    }),
  };
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const manifestPath = path.resolve(
    root,
    process.env.BROWSER_E2E_BATCH_MANIFEST || "tools/browser_e2e_batch_manifest.json",
  );
  const stage = resolveBrowserBatchStage(manifestPath, options.stageName);
  const result = await runNormalizedSchedule({
    repoRoot: root,
    schedule: buildBrowserStageSchedule(stage, manifestPath, {
      mode: options.mode,
      target: process.env.CARTULARY_TEST_TARGET,
    }),
    testOutputScript: process.env.TEST_OUTPUT_SCRIPT || path.join(root, "tools", "harness", "output", "test-output.mjs"),
  });
  return publicExitCodeForSummary(result.summary);
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main()
    .then((status) => { process.exitCode = status; })
    .catch((error) => {
      process.stderr.write(`${error.message}\n`);
      process.exitCode = error.message === usage() ? 2 : 11;
    });
}
