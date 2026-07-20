import { mkdtempSync, rmSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import {
  assignExecutionFamily,
  captureNamedSharedReportsParallel,
} from "./capture.mjs";
import { createAggregateReport } from "./reports.mjs";
import {
  aggregateNames,
  rowsForScheduledAggregate,
  targetAggregates,
  targetShards,
} from "./planning.mjs";
import {
  emitExecutionFamily,
  finishTarget,
  runBounded,
  writeFinalizerFailureStep,
  writeTargetTimingSpan,
} from "./summary-emission.mjs";
import {
  captureFinish,
  captureStart,
} from "./util.mjs";

export async function runUnshardedTarget(ctx, target) {
  let status = 0;
  for (const aggregateName of aggregateNames(ctx, target)) {
    try {
      const captured = await assignExecutionFamily(ctx, target, aggregateName);
      if (status === 0) {
        status = await emitExecutionFamily(
          ctx,
          target,
          aggregateName,
          captured.usage,
          captured.reportDir,
        );
      }
    } catch (error) {
      process.stderr.write(`${error.message}\n`);
      status = Number.isInteger(error.exitCode) ? error.exitCode : 1;
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
  const aggregateReports = [];
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
  for (const aggregate of targetAggregates(ctx, target)) {
    const shardNames = selectedShardSet
      ? aggregate.shards.filter((shardName) => selectedShardSet.has(shardName))
      : aggregate.shards;
    if (selectedShardSet && shardNames.length === 0) {
      continue;
    }
    const aggregateStarted = captureStart();
    let report = null;
    let rows = null;
    const aggregateRoot = path.join(metadataDir, "aggregate-reports", target);
    const aggregateReportDir = path.join(aggregateRoot, aggregate.name);
    try {
      rows = rowsForScheduledAggregate(ctx, target, aggregate.name, shardNames);
      report = createAggregateReport(
        ctx,
        metadataDir,
        aggregate.name,
        target,
        shardNames,
        metadataByShard,
      );
      const aggregateWindow = captureFinish(aggregateStarted);
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `collate ${target}:${aggregate.name}`,
        aggregateWindow,
        "pass",
      );
    } catch (error) {
      const aggregateWindow = captureFinish(aggregateStarted);
      writeTargetTimingSpan(
        ctx,
        "report_collation",
        `collate ${target}:${aggregate.name}`,
        aggregateWindow,
        "fail",
      );
      writeFinalizerFailureStep(ctx, {
        target,
        label: `collate ${target}:${aggregate.name}`,
        commandArgs: finalizerCommandArgs,
        window: aggregateWindow,
        error,
        metadataDir,
        aggregateReportDir,
        shardNames,
      });
      process.stderr.write(`${error.message}\n`);
      status = 1;
      continue;
    }
    aggregateReports.push({ aggregate, report, rows });
  }
  if (status === 0) {
    const emitStatuses = await runBounded(
      aggregateReports,
      1,
      async ({ aggregate, report, rows }) => {
        let emitStatus = 0;
        const emitStarted = captureStart();
        try {
          emitStatus = await emitExecutionFamily(
            ctx,
            target,
            aggregate.name,
            report.usage,
            report.reportDir,
            rows,
          );
        } catch (error) {
          process.stderr.write(`${error.message}\n`);
          emitStatus = 1;
        }
        const emitWindow = captureFinish(emitStarted);
        writeTargetTimingSpan(
          ctx,
          "report_collation",
          `emit ${target}:${aggregate.name}`,
          emitWindow,
          emitStatus === 0 ? "pass" : "fail",
        );
        return emitStatus;
      },
    );
    for (const emitStatus of emitStatuses) {
      if (emitStatus !== 0) {
        status = emitStatus;
        break;
      }
    }
  }
  return await finishTarget(ctx, status);
}

export async function runShardedTarget(ctx, target) {
  const metadataDir = mkdtempSync(
    path.join(os.tmpdir(), `cartulary-${target}-shards.`),
  );
  const shardNames = targetShards(ctx, target).map((shard) => shard.name);
  let status = await captureNamedSharedReportsParallel(
    ctx,
    target,
    ctx.backendIntegrationShardJobs,
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
