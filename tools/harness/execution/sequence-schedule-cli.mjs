#!/usr/bin/env node

import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  compactJSONString,
  createRunnerContext,
  secureWriteFile,
  validateSchemaSync,
} from "../contract/index.mjs";
import {
  loadTaskSurfaceManifest,
  sequenceDefinition,
} from "../generated-artifacts/task-surface/index.mjs";
import {
  loadSummaryTopologyContext,
  resolveSummaryGroups,
  summaryGroupsSpec,
} from "./summary-topology.mjs";
import {
  isDryRunFromMakeFlags,
  makeChildEnv,
  runLifecycle,
  runNormalizedSchedule,
  writeSchedulerDryRun,
} from "../scheduler/scheduler-runner.mjs";
import {
  formatResourceMap,
  normalizeResourceClaims,
  normalizeResourceLimits,
  provisionalResourceLimitsForClaims,
  resolveResourceForwardingProfile,
} from "../scheduler/scheduler-resources.mjs";
import { resolveSchedulerResourceLimits } from "../scheduler/scheduler-resource-policy.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const sequenceEventSchemaID = "cartulary.harness_sequence_event.v1";
const schedulerEventSchemaID = "cartulary.scheduler_event.v7";
const schedulerSummarySchemaID = "cartulary.sequence_scheduler_summary.v1";
const usage = "usage: run-make-sequence.sh --sequence <name> [--manifest <path>]";

function parseArgs(args) {
  const options = { sequence: "", manifest: process.env.TASK_SURFACE_MANIFEST ?? "tools/task_surface_manifest.json" };
  for (let index = 0; index < args.length; index += 1) {
    const arg = args[index];
    if (arg === "--sequence" || arg === "--manifest") {
      const value = args[index + 1];
      if (!value) throw new Error(`${arg} requires a value`);
      options[arg.slice(2)] = value;
      index += 1;
      continue;
    }
    throw new Error(`${usage}\nunknown option ${arg}`);
  }
  if (!options.sequence) {
    throw new Error(usage);
  }
  return options;
}

function positiveInteger(value, label) {
  const parsed = Number.parseInt(String(value ?? ""), 10);
  if (!Number.isInteger(parsed) || parsed < 1 || String(parsed) !== String(value)) {
    throw new Error(`${label} must be a positive integer`);
  }
  return parsed;
}

function stepJobs(step, sequenceName, resourceClaims) {
  if (step.makeJobs !== undefined) {
    const value = typeof step.makeJobs === "string"
      ? resourceClaims.get(step.makeJobs)
      : step.makeJobs;
    return positiveInteger(value, `sequence ${sequenceName} step ${step.target} make_jobs`);
  }
  if (step.type !== "parallel") return 1;
  const value = step.jobs ?? (step.jobsVariable ? process.env[step.jobsVariable] : undefined);
  return positiveInteger(value, `sequence ${sequenceName} step ${step.target} jobs`);
}

function childMakeEnv(skipPrerequisites) {
  const env = makeChildEnv(process.env);
  for (const name of [
    "CARTULARY_HARNESS_IDENTITY_PREPARED",
    "CARTULARY_TEST_TARGET",
    "CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES",
    "CARTULARY_SEQUENCE_PREREQUISITES_SATISFIED",
  ]) {
    delete env[name];
  }
  env.CARTULARY_SUPPRESS_CHILD_SUCCESS = "1";
  if (skipPrerequisites) {
    env.CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES = "1";
    env.CARTULARY_SEQUENCE_PREREQUISITES_SATISFIED = "1";
  }
  return env;
}

function compileSequence(sequence, context, manifestPath) {
  const scheduleLabel = `sequence ${sequence.name}`;
  const normalizedLimits = normalizeResourceLimits(
    sequence.resourceLimits,
    scheduleLabel,
    {
      scheduler: "sequence",
      capacityProfile: sequence.capacityProfile || null,
      allowAuto: true,
      env: null,
    },
  );
  const provisionalLimits = provisionalResourceLimitsForClaims(
    normalizedLimits.limits,
  );
  const provisionalUnits = sequence.steps.map((step) => ({
    resourceClaims: normalizeResourceClaims(
      step.resourceClaims,
      `${scheduleLabel} step ${step.target}`,
      provisionalLimits,
      { scheduler: "sequence" },
    ),
  }));
  const resolvedLimits = resolveSchedulerResourceLimits({
    scheduler: "sequence",
    resourceLimits: normalizedLimits.limits,
    resourceLimitSources: normalizedLimits.sources,
    label: scheduleLabel,
    workUnits: provisionalUnits,
  });
  const resourceLimits = resolvedLimits.resourceLimits;
  const units = sequence.steps.map((step, index) => {
    const declaredNeeds = [...step.needs];
    if (sequence.executionMode === "serial" && index > 0 && !declaredNeeds.includes(sequence.steps[index - 1].target)) {
      declaredNeeds.push(sequence.steps[index - 1].target);
    }
    const resourceClaims = normalizeResourceClaims(
      step.resourceClaims,
      `${scheduleLabel} step ${step.target}`,
      resourceLimits,
      { scheduler: "sequence" },
    );
    const forwarding = step.forwarding === ""
      ? null
      : resolveResourceForwardingProfile(
        step.forwarding,
        resourceClaims,
        `${scheduleLabel} step ${step.target}`,
      );
    const makeJobs = stepJobs(step, sequence.name, resourceClaims);
    const args = ["--no-print-directory"];
    if (makeJobs > 1) args.push("--output-sync=target", `-j${makeJobs}`);
    args.push(step.target);
    return {
      id: step.target,
      label: step.target,
      kind: "make_target",
      type: "make_target",
      class: "sequence",
      target: step.target,
      aggregateTarget: step.target,
      completionKeys: [step.target],
      failureKeys: [step.target],
      runningDependencyKeys: [],
      priority: Number.isInteger(step.priority) ? step.priority : 0,
      weightMs: 1,
      needs: declaredNeeds,
      producesSummaryTargets: [...step.producesSummaryTargets],
      resourceClaims,
      retainedResourceClaims: new Map(),
      releaseRetainedResourceClaims: new Map(),
      makeJobs,
      makePrerequisitePolicy: step.skipPrerequisites ? "skip" : "run",
      env: {},
      runtimeBinaries: [],
      serviceSession: null,
      browserStage: "",
      browserSessionGroup: "",
      browserSessionIsolationReason: "",
      browserGroup: null,
      shard: "",
      shardNames: [],
      schedulerProfile: step.resourceProfile,
      readinessAttribution: null,
      completeOnFailure: false,
      timeoutMs: 0,
      startDetail: {},
      order: index,
      command: {
        command: context.makeBin,
        args,
        env: {
          ...childMakeEnv(step.skipPrerequisites),
          ...Object.fromEntries(forwarding?.resourceLimitEnv ?? []),
        },
      },
      forwardingProfile: forwarding?.profile ?? "",
      forwardingMappings: forwarding?.forwardingMappings ?? [],
      forwardedResourceLimits: forwarding?.forwardedResourceLimits ?? new Map(),
    };
  });
  const ids = new Set(units.map((unit) => unit.id));
  if (ids.size !== units.length) throw new Error(`sequence ${sequence.name} contains duplicate targets`);
  for (const unit of units) {
    for (const dependency of unit.needs) {
      if (!ids.has(dependency)) throw new Error(`sequence ${sequence.name} step ${unit.id} depends on unknown target ${dependency}`);
    }
  }
  const visiting = new Set();
  const visited = new Set();
  const byID = new Map(units.map((unit) => [unit.id, unit]));
  const visit = (unit) => {
    if (visited.has(unit.id)) return;
    if (visiting.has(unit.id)) throw new Error(`sequence ${sequence.name} has a dependency cycle at ${unit.id}`);
    visiting.add(unit.id);
    for (const dependency of unit.needs) visit(byID.get(dependency));
    visiting.delete(unit.id);
    visited.add(unit.id);
  };
  for (const unit of units) visit(unit);
  const summaryTargets = units.flatMap((unit) => unit.producesSummaryTargets);
  const summaryTargetSet = new Set(summaryTargets);
  const helperUnitNames = units.filter((unit) => !summaryTargetSet.has(unit.target)).map((unit) => unit.target);
  const topologyContext = loadSummaryTopologyContext({
    taskSurfaceManifest: context.taskSurfaceManifest,
    schedulerManifestPath: process.env.SCHEDULER_MANIFEST,
    browserBatchManifestPath: process.env.BROWSER_E2E_BATCH_MANIFEST,
  });
  const summaryGroups = summaryGroupsSpec(resolveSummaryGroups(topologyContext, sequence.summaryGroups));
  return {
    target: sequence.name,
    kind: "sequence",
    prefix: "SEQUENCE-SCHEDULER",
    eventSchemaID: schedulerEventSchemaID,
    summarySchemaID: schedulerSummarySchemaID,
    resourceScheduler: "sequence",
    stopOnFirstFailure: true,
    summaryTotalWallTime: true,
    quietHumanOutput: true,
    validateSummaryTiming: true,
    resourceLimits,
    resourceLimitSources: resolvedLimits.resourceLimitSources,
    workUnits: units.sort(
      (left, right) => right.priority - left.priority || left.order - right.order || left.id.localeCompare(right.id),
    ),
    totalWorkUnits: units.length,
    finalizerCount: 0,
    nestedSchedulerLimits: () => units
      .filter((unit) => unit.forwardingProfile !== "")
      .map((unit) => ({
        work_unit_id: unit.id,
        target: unit.target,
        forwarding_profile: unit.forwardingProfile,
        mappings: unit.forwardingMappings,
      })),
    nestedSchedulerObservations: () => [],
    countCompletedUnit: (_unit, result) => result.status === 0,
    shouldReplayLog: ({ result, reporter }) => result.status !== 0 || reporter.verbose,
    summaryTargets,
    summaryTargetSet,
    helperUnitNames,
    summaryGroups,
    executionMode: sequence.executionMode,
    maxJobs: Math.min(
      sequence.maxJobs,
      resourceLimits.get("process") ?? sequence.maxJobs,
    ),
    capacityProfile: sequence.capacityProfile,
    manifestPath,
  };
}

function eventStatusForState(state) {
  if (state.terminal_state === "passed") return { status: "pass", exitCode: 0 };
  if (state.terminal_state === "failed") return { status: "fail", exitCode: 1 };
  return { status: "interrupted", exitCode: 130 };
}

function retainedSequenceEvents(schedule, reporter, requestedStatus, firstFailure) {
  const states = reporter.workUnitStateSnapshot();
  const completedByID = new Map(reporter.completedWork.map((item) => [item.id, item]));
  const skippedByID = new Map(reporter.skippedWork.map((item) => [item.id, item]));
  const events = [{
    order: 0,
    ordinal: 0,
    monotonic: reporter.schedulerStartedMonotonicMs,
    fields: { event: "sequence_started", status: "running" },
  }];
  for (const state of states) {
    const unit = schedule.workUnits.find((candidate) => candidate.id === state.work_unit_id);
    if (!unit || state.eligibility_monotonic_ms === null || state.terminal_monotonic_ms === null) continue;
    const stepIndex = unit.order + 1;
    const common = {
      step_index: stepIndex,
      target: unit.target,
      needs: [...state.dependencies],
      resource_claims: { ...state.resource_claims },
      mode: schedule.executionMode,
      jobs: unit.makeJobs,
    };
    const eligibleMono = state.eligibility_monotonic_ms;
    const eligibleAt = reporter.clock.wallTimestamp(eligibleMono);
    events.push({
      order: 1,
      ordinal: stepIndex,
      monotonic: eligibleMono,
      fields: {
        event: "step_eligible",
        ...common,
        status: "pending",
        eligible_at: eligibleAt,
        eligible_monotonic_ms: eligibleMono,
      },
    });
    if (state.started_monotonic_ms !== null) {
      events.push({
        order: 2,
        ordinal: stepIndex,
        monotonic: state.started_monotonic_ms,
        fields: {
          event: "step_started",
          ...common,
          status: "running",
          eligible_at: eligibleAt,
          started_at: reporter.clock.wallTimestamp(state.started_monotonic_ms),
          eligible_monotonic_ms: eligibleMono,
          started_monotonic_ms: state.started_monotonic_ms,
        },
      });
    }
    const completed = completedByID.get(unit.id);
    const skipped = skippedByID.get(unit.id);
    const terminalStatus = eventStatusForState(state);
    const startedMono = state.started_monotonic_ms ?? state.terminal_monotonic_ms;
    const terminalFields = {
      ...common,
      status: completed?.status === 0 ? "pass" : terminalStatus.status,
      exit_code: completed?.status ?? terminalStatus.exitCode,
      duration_ms: Math.max(0, state.terminal_monotonic_ms - startedMono),
      eligible_at: eligibleAt,
      ended_at: reporter.clock.wallTimestamp(state.terminal_monotonic_ms),
      eligible_monotonic_ms: eligibleMono,
      ended_monotonic_ms: state.terminal_monotonic_ms,
      ...(state.started_monotonic_ms === null ? {} : {
        started_at: reporter.clock.wallTimestamp(state.started_monotonic_ms),
        started_monotonic_ms: state.started_monotonic_ms,
      }),
    };
    if (skipped) {
      terminalFields.skip_reason = skipped.reason === "dependency_failure"
        ? "dependency_failed"
        : state.terminal_state === "interrupted"
          ? "interrupted"
          : "scheduler_stop";
      if (skipped.failed_dependency) terminalFields.failed_dependency = skipped.failed_dependency;
    }
    events.push({
      order: 3,
      ordinal: stepIndex,
      monotonic: state.terminal_monotonic_ms,
      fields: { event: skipped ? "step_skipped" : "step_finished", ...terminalFields },
    });
  }
  const interrupted = firstFailure === 130 || firstFailure === 143 || states.some((state) => state.terminal_state === "interrupted");
  events.push({
    order: 4,
    ordinal: Number.MAX_SAFE_INTEGER,
    monotonic: reporter.schedulerCompletedMonotonicMs,
    fields: {
      event: interrupted ? "sequence_interrupted" : "sequence_finished",
      status: interrupted ? "interrupted" : requestedStatus,
      exit_code: Math.max(0, firstFailure),
      duration_ms: reporter.schedulerTotalDurationMs,
    },
  });
  events.sort((left, right) =>
    left.monotonic - right.monotonic || left.order - right.order || left.ordinal - right.ordinal,
  );
  return events.map((entry, index) => {
    const event = {
      schema_id: sequenceEventSchemaID,
      run_id: process.env.CARTULARY_TEST_RUN_ID ?? "adhoc",
      sequence: schedule.target,
      seq: index + 1,
      emitted_at: reporter.clock.wallTimestamp(entry.monotonic),
      monotonic_ms: entry.monotonic,
      ...entry.fields,
    };
    validateSchemaSync(sequenceEventSchemaID, event);
    return event;
  });
}

function writeSequenceEvents(context, schedule, reporter, requestedStatus, firstFailure) {
  const events = retainedSequenceEvents(schedule, reporter, requestedStatus, firstFailure);
  const file = path.join(context.resultsDir, context.runId, schedule.target, "sequence-events.jsonl");
  secureWriteFile(file, events.map((event) => compactJSONString(event)).join(""), {
    allowedRoot: path.join(context.resultsDir, context.runId),
  });
}

function skippedSummaryTargets(schedule, reporter) {
  const unitsByID = new Map(schedule.workUnits.map((unit) => [unit.id, unit]));
  const skipped = new Set();
  for (const item of reporter.skippedWork) {
    const unit = unitsByID.get(item.id);
    if (!unit) continue;
    if (schedule.summaryTargetSet.has(unit.target)) skipped.add(unit.target);
    for (const target of unit.producesSummaryTargets) skipped.add(target);
  }
  return schedule.summaryTargets.filter((target) => skipped.has(target));
}

function attachSummaryAdapter(schedule, context) {
  return {
    ...schedule,
    afterSummary: async ({ reporter, requestedStatus, completedKeys, firstFailure, firstFailureLabel }) => {
      writeSequenceEvents(context, schedule, reporter, requestedStatus, firstFailure);
      const summaryArgs = [
        "run-summary",
        schedule.target,
        requestedStatus,
        String(reporter.completedCount),
        String(schedule.totalWorkUnits),
        firstFailureLabel ?? "-",
        "--suppress-machine-output",
      ];
      if (schedule.summaryGroups) summaryArgs.push("--summary-groups", schedule.summaryGroups);
      if (schedule.helperUnitNames.length > 0) {
        summaryArgs.push("--helper-units", schedule.helperUnitNames.join(","));
        summaryArgs.push(
          "--completed-helper-units",
          schedule.helperUnitNames.filter((name) => completedKeys.has(name)).join(","),
        );
      }
      const skipped = skippedSummaryTargets(schedule, reporter);
      if (skipped.length > 0) summaryArgs.push("--skipped-after-failure", skipped.join(","));
      summaryArgs.push(...schedule.summaryTargets);
      await runLifecycle(
        repoRoot,
        context.testOutputScript,
        summaryArgs,
        requestedStatus === "pass" ? process.stdout : process.stderr,
      ).catch((error) => {
        if (requestedStatus === "pass") throw error;
      });
      await runLifecycle(repoRoot, context.testOutputScript, [
        "target-summary",
        schedule.target,
        requestedStatus,
        "--children",
        schedule.summaryTargets.join(","),
        "--skipped-from-scheduler",
        schedule.target,
        requestedStatus === "pass" ? "--quiet-success" : "--quiet-failure",
      ]).catch((error) => {
        if (requestedStatus === "pass") throw error;
      });
    },
  };
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const context = createRunnerContext({ repoRoot });
  const { manifest, manifestPath } = loadTaskSurfaceManifest(options.manifest);
  context.taskSurfaceManifest = manifest;
  const sequence = sequenceDefinition(manifest, options.sequence);
  const schedule = attachSummaryAdapter(compileSequence(sequence, context, manifestPath), context);
  if (isDryRunFromMakeFlags()) {
    writeSchedulerDryRun({
      repoRoot,
      schedule,
      manifestPath,
      verboseUnitLine: (unit) =>
        `[DRY-RUN] ${schedule.target} unit ${unit.label} needs=${unit.needs.length ? unit.needs.join(",") : "none"} claims=${formatResourceMap(unit.resourceClaims)} make_jobs=${unit.makeJobs}\n`,
    });
    return;
  }
  const result = await runNormalizedSchedule({ repoRoot, schedule, testOutputScript: context.testOutputScript });
  process.exitCode = result.status;
}

main().catch((error) => {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = Number.isInteger(error?.exitCode) ? error.exitCode : 2;
});
