import { mkdtempSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import {
  captureNamedSharedReportsParallel,
  captureUnshardedGroup,
} from "./capture.mjs";
import {
  createScheduledAggregateReport,
  createUnshardedFamilyReport,
  parsePhysicalReport,
  parseScheduledPhysicalReports,
} from "./reports.mjs";
import {
  aggregateNames,
  rowsForScheduledAggregate,
  targetAggregates,
  targetShards,
  unshardedCaptureGroups,
} from "./planning.mjs";
import {
  createExecutionFamilyRequest,
  emitExecutionFamilyRequests,
  finishTarget,
  runSettledBounded,
  writeFinalizerFailureStep,
  writeTargetTimingSpan,
} from "./summary-emission.mjs";
import {
  captureFinish,
  captureStart,
} from "./util.mjs";
import { resolveBackendWorkerPool } from "./worker-policy.mjs";

export async function runUnshardedTarget(ctx, target) {
  const groups = unshardedCaptureGroups(ctx, target);
  const pool = resolveBackendWorkerPool(ctx.availableParallelism, groups.length);
  ctx.goMaxProcs = pool.goMaxProcs;
  let status = 0;
  const captures = await runSettledBounded(groups, pool.workers, async (group) => {
    const captured = await captureUnshardedGroup(ctx, group);
    const parseStarted = captureStart();
    try {
      const report = parsePhysicalReport(captured.reportDir);
      const parseWindow = captureFinish(parseStarted);
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `parse ${target}:${group.name}`,
        parseWindow,
        "pass",
      );
      return Object.freeze({ captured, report });
    } catch (error) {
      const parseWindow = captureFinish(parseStarted);
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `parse ${target}:${group.name}`,
        parseWindow,
        "fail",
      );
      throw error;
    }
  });
  for (const result of captures) {
    if (result.error) {
      process.stderr.write(`${result.error.message}\n`);
      if (status === 0) status = Number.isInteger(result.error.exitCode) ? result.error.exitCode : 1;
    }
  }
  const familyRequests = [];
  for (const family of aggregateNames(ctx, target)) {
    const entries = [];
    let incomplete = false;
    for (const [groupIndex, group] of groups.entries()) {
      if (!group.families.includes(family)) continue;
      const result = captures[groupIndex];
      if (result.error) {
        incomplete = true;
        break;
      }
      const ownsPhysicalCapture = group.families[0] === family;
      entries.push({
        name: group.name,
        report: result.value.report,
        usage: ownsPhysicalCapture
          ? result.value.captured.usage
          : result.value.captured.usage === "actual" ? "reused" : result.value.captured.usage,
      });
    }
    if (incomplete) continue;
    familyRequests.push(Object.freeze({
      family,
      entries: Object.freeze(entries.map((entry) => Object.freeze(entry))),
    }));
  }
  const collationResults = await runSettledBounded(
    familyRequests,
    pool.workers,
    async (request) => {
      const collationStarted = captureStart();
      try {
        const report = createUnshardedFamilyReport(
          ctx,
          request.family,
          request.entries,
        );
        const emissionRequest = createExecutionFamilyRequest(
          ctx,
          target,
          request.family,
          report.usage,
          report.reportDir,
        );
        const collationWindow = captureFinish(collationStarted);
        writeTargetTimingSpan(
          ctx,
          "report_collation",
          `collate ${target}:${request.family}`,
          collationWindow,
          "pass",
        );
        return Object.freeze({ request, report, emissionRequest });
      } catch (error) {
        const collationWindow = captureFinish(collationStarted);
        writeTargetTimingSpan(
          ctx,
          "report_collation",
          `collate ${target}:${request.family}`,
          collationWindow,
          "fail",
        );
        writeFinalizerFailureStep(ctx, {
          target,
          label: `collate ${target}:${request.family}`,
          commandArgs: ["run-go-target.mjs", target],
          window: collationWindow,
          error,
          metadataDir: path.join(ctx.resultsRoot, ctx.runId, "_shared"),
        });
        throw error;
      }
    },
  );
  const successfulCollations = collationResults
    .map((result, index) => ({ result, index }))
    .filter(({ result }) => !result.error);
  const emissionResults = await emitExecutionFamilyRequests(
    ctx,
    successfulCollations.map(({ result }) => result.value.emissionRequest),
    pool.workers,
  );
  const emissionsByFamilyIndex = new Map(
    successfulCollations.map(({ index }, resultIndex) => [index, emissionResults[resultIndex]]),
  );
  for (const [index, collationResult] of collationResults.entries()) {
    if (collationResult.error) {
      process.stderr.write(`${collationResult.error.message}\n`);
      if (status === 0) status = Number.isInteger(collationResult.error.exitCode)
        ? collationResult.error.exitCode
        : 1;
      continue;
    }
    const emissionResult = emissionsByFamilyIndex.get(index);
    const emissionWindow = emissionResult.value?.window ?? emissionResult.error?.window;
    if (emissionWindow) {
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `emit ${target}:${collationResult.value.request.family}`,
        emissionWindow,
        emissionResult.error || emissionResult.value.status !== 0 ? "fail" : "pass",
      );
    }
    if (emissionResult.error) {
      writeFinalizerFailureStep(ctx, {
        target,
        label: `emit ${target}:${collationResult.value.request.family}`,
        commandArgs: ["run-go-target.mjs", target],
        window: emissionWindow ?? captureFinish(captureStart()),
        error: emissionResult.error,
        metadataDir: path.join(ctx.resultsRoot, ctx.runId, "_shared"),
      });
      process.stderr.write(`${emissionResult.error.message}\n`);
      if (status === 0) status = Number.isInteger(emissionResult.error.exitCode)
        ? emissionResult.error.exitCode
        : 1;
    } else if (status === 0 && emissionResult.value.status !== 0) {
      status = emissionResult.value.status;
    }
  }
  return await finishTarget(ctx, status);
}

export async function finalizeScheduledShards(
  ctx,
  target,
  metadataDir,
  selectedShardNames = [],
) {
  let status = 0;
  const metadataByShard = new Map();
  const selectedShardSet =
    selectedShardNames.length > 0 ? new Set(selectedShardNames) : null;
  const finalizerCommandArgs = [
    "run-go-target.mjs",
    "finalize-shards",
    target,
    metadataDir,
    ...selectedShardNames,
  ];
  if (selectedShardSet) {
    const validationStarted = captureStart();
    const knownShards = new Set(targetShards(ctx, target).map((shard) => shard.name));
    for (const shardName of selectedShardSet) {
      if (!knownShards.has(shardName)) {
        const error = new Error(`unknown scheduled shard ${shardName} for ${target}`);
        const validationWindow = captureFinish(validationStarted);
        writeTargetTimingSpan(
          ctx,
          "report_collation",
          `validate ${target}:selected-shards`,
          validationWindow,
          "fail",
        );
        writeFinalizerFailureStep(ctx, {
          target,
          label: `validate ${target}:selected-shards`,
          commandArgs: finalizerCommandArgs,
          window: validationWindow,
          error,
          metadataDir,
          shardNames: selectedShardNames,
        });
        process.stderr.write(`${error.message}\n`);
        return await finishTarget(ctx, 1);
      }
    }
  }
  const aggregateRequests = [];
  for (const aggregate of targetAggregates(ctx, target)) {
    const shardNames = selectedShardSet
      ? aggregate.shards.filter((shardName) => selectedShardSet.has(shardName))
      : aggregate.shards;
    if (selectedShardSet && shardNames.length === 0) {
      continue;
    }
    aggregateRequests.push(Object.freeze({
      aggregateName: aggregate.name,
      shardNames: Object.freeze([...shardNames]),
    }));
  }
  const shardNames = [...new Set(aggregateRequests.flatMap((request) => request.shardNames))];
  const pool = resolveBackendWorkerPool(
    ctx.availableParallelism,
    Math.max(shardNames.length, aggregateRequests.length),
  );
  const parsedResults = await runSettledBounded(shardNames, pool.workers, async (shardName) => {
    const parseStarted = captureStart();
    try {
      const parsed = parseScheduledPhysicalReports(
        metadataDir,
        [shardName],
        metadataByShard,
      );
      const parseWindow = captureFinish(parseStarted);
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `parse ${target}:${shardName}`,
        parseWindow,
        "pass",
      );
      return parsed.get(shardName);
    } catch (error) {
      const parseWindow = captureFinish(parseStarted);
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `parse ${target}:${shardName}`,
        parseWindow,
        "fail",
      );
      throw error;
    }
  });
  const parsedByShard = new Map();
  const parseErrorsByShard = new Map();
  for (const [index, shardName] of shardNames.entries()) {
    const result = parsedResults[index];
    if (result.error) parseErrorsByShard.set(shardName, result.error);
    else parsedByShard.set(shardName, result.value);
  }
  const collationResults = await runSettledBounded(
    aggregateRequests,
    pool.workers,
    async (request) => {
      const aggregateStarted = captureStart();
      const aggregateReportDir = path.join(
        metadataDir,
        "aggregate-reports",
        target,
        request.aggregateName,
      );
      try {
        for (const shardName of request.shardNames) {
          if (parseErrorsByShard.has(shardName)) throw parseErrorsByShard.get(shardName);
        }
        const rows = rowsForScheduledAggregate(
          ctx,
          target,
          request.aggregateName,
          request.shardNames,
        );
        const entries = Object.freeze(request.shardNames.map((shardName) => {
          const retained = parsedByShard.get(shardName);
          return Object.freeze({
            name: shardName,
            usage: retained.metadata.usage,
            report: retained.report,
          });
        }));
        const report = createScheduledAggregateReport(
          ctx,
          metadataDir,
          request.aggregateName,
          target,
          entries,
        );
        const emissionRequest = createExecutionFamilyRequest(
          ctx,
          target,
          request.aggregateName,
          report.usage,
          report.reportDir,
          rows,
        );
        const aggregateWindow = captureFinish(aggregateStarted);
        writeTargetTimingSpan(
          ctx,
          "report_collation",
          `collate ${target}:${request.aggregateName}`,
          aggregateWindow,
          "pass",
        );
        return Object.freeze({ request, report, rows, emissionRequest });
      } catch (error) {
        const aggregateWindow = captureFinish(aggregateStarted);
        writeTargetTimingSpan(
          ctx,
          "report_collation",
          `collate ${target}:${request.aggregateName}`,
          aggregateWindow,
          "fail",
        );
        writeFinalizerFailureStep(ctx, {
          target,
          label: `collate ${target}:${request.aggregateName}`,
          commandArgs: finalizerCommandArgs,
          window: aggregateWindow,
          error,
          metadataDir,
          aggregateReportDir,
          shardNames: request.shardNames,
        });
        throw error;
      }
    },
  );
  const successfulCollations = collationResults
    .map((result, index) => ({ result, index }))
    .filter(({ result }) => !result.error);
  const emissionResults = await emitExecutionFamilyRequests(
    ctx,
    successfulCollations.map(({ result }) => result.value.emissionRequest),
    pool.workers,
  );
  const emissionsByAggregateIndex = new Map(
    successfulCollations.map(({ index }, resultIndex) => [index, emissionResults[resultIndex]]),
  );
  for (const [index, collationResult] of collationResults.entries()) {
    if (collationResult.error) {
      process.stderr.write(`${collationResult.error.message}\n`);
      if (status === 0) status = Number.isInteger(collationResult.error.exitCode)
        ? collationResult.error.exitCode
        : 1;
      continue;
    }
    const emissionResult = emissionsByAggregateIndex.get(index);
    const emissionWindow = emissionResult.value?.window ?? emissionResult.error?.window;
    if (emissionWindow) {
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `emit ${target}:${collationResult.value.request.aggregateName}`,
        emissionWindow,
        emissionResult.error || emissionResult.value.status !== 0 ? "fail" : "pass",
      );
    }
    if (emissionResult.error) {
      writeFinalizerFailureStep(ctx, {
        target,
        label: `emit ${target}:${collationResult.value.request.aggregateName}`,
        commandArgs: finalizerCommandArgs,
        window: emissionWindow ?? captureFinish(captureStart()),
        error: emissionResult.error,
        metadataDir,
        aggregateReportDir: collationResult.value.report.reportDir,
        shardNames: collationResult.value.request.shardNames,
      });
      process.stderr.write(`${emissionResult.error.message}\n`);
      if (status === 0) status = Number.isInteger(emissionResult.error.exitCode)
        ? emissionResult.error.exitCode
        : 1;
    } else if (status === 0 && emissionResult.value.status !== 0) {
      status = emissionResult.value.status;
    }
  }
  return await finishTarget(ctx, status);
}

export async function runShardedTarget(ctx, target) {
  const metadataDir = mkdtempSync(
    path.join(os.tmpdir(), `cartulary-${target}-shards.`),
  );
  const shardNames = targetShards(ctx, target).map((shard) => shard.name);
  const pool = resolveBackendWorkerPool(ctx.availableParallelism, shardNames.length);
  ctx.goMaxProcs = pool.goMaxProcs;
  let status = await captureNamedSharedReportsParallel(
    ctx,
    target,
    pool.workers,
    metadataDir,
    shardNames,
  );
  if (status === 0) {
    status = await finalizeScheduledShards(ctx, target, metadataDir, shardNames);
  } else {
    status = await finishTarget(ctx, status);
  }
  rmSync(metadataDir, { recursive: true, force: true });
  return status;
}
