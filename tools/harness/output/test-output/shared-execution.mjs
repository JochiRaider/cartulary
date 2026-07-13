#!/usr/bin/env node
import { repoRoot } from "../../contract/index.mjs";

import {
  existsSync,
  readdirSync,
  readFileSync,
} from "node:fs";
import path from "node:path";
import { failureFieldsForJSON } from "../../contract/failure-taxonomy.mjs";
import {
  prettyJSONString,
  secureMkdir,
  secureWriteFile,
} from "../../contract/harness-contract.mjs";
import {
  resolveResultsRoot,
  resolveRunId,
  sharedExecutionGroupSchemaID,
} from "../../contract/test-output-context.mjs";
import {
  disjointSpanDurationMs,
  timingStatusFailed,
} from "./timing.mjs";

const resultsRoot = resolveResultsRoot();

const runId = resolveRunId();

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function relToRepo(value) {
  if (!value) {
    return "";
  }
  const normalized = normalizePath(value);
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(repoRoot, value));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

function ensureDir(dir) {
  secureMkdir(dir);
}

function writeJson(file, value) {
  ensureDir(path.dirname(file));
  secureWriteFile(file, prettyJSONString(value));
}

function clampDurationMs(value) {
  if (!Number.isFinite(value) || value < 0) {
    return 0;
  }
  return value;
}

export function handleSharedExecution(args) {
  const [
    group,
    sharedReport,
    status,
    startTime,
    endTime,
    durationText,
    exitStatusText,
    outputPath,
  ] = args;
  if (
    !group ||
    !sharedReport ||
    !status ||
    !startTime ||
    !endTime ||
    durationText === undefined ||
    exitStatusText === undefined ||
    !outputPath
  ) {
    throw new Error(
      "usage: test-output.mjs shared-execution <group> <shared-report> <status> <start-time> <end-time> <duration-ms> <exit-status> <output-path>",
    );
  }
  const durationMs = clampDurationMs(Number.parseInt(durationText, 10));
  writeJson(outputPath, {
    schema_id: sharedExecutionGroupSchemaID,
    execution_group: group,
    shared_report: sharedReport,
    status,
    start_time: startTime,
    end_time: endTime,
    duration_ms: durationMs,
    wall_duration_ms: durationMs,
    executed_duration_ms: durationMs,
    exit_status: Number.parseInt(exitStatusText, 10) || 0,
    artifact: relToRepo(path.dirname(outputPath)),
  });
  return 0;
}

function loadSharedExecutionRecords() {
  const sharedRoot = path.join(resultsRoot, runId, "_shared");
  if (!existsSync(sharedRoot)) {
    return [];
  }
  const records = [];
  const stack = [sharedRoot];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (!entry.isFile() || entry.name !== "shared-execution.json") {
        continue;
      }
      let record;
      try {
        record = JSON.parse(readFileSync(next, "utf8"));
      } catch {
        continue;
      }
      if (
        record?.schema_id !== sharedExecutionGroupSchemaID ||
        !record.execution_group
      ) {
        continue;
      }
      records.push({
        execution_group: record.execution_group,
        shared_report: record.shared_report ?? "",
        status: record.status ?? "",
        start_time: record.start_time ?? "",
        end_time: record.end_time ?? "",
        duration_ms: clampDurationMs(record.duration_ms ?? 0),
        wall_duration_ms: clampDurationMs(
          record.wall_duration_ms ?? record.duration_ms ?? 0,
        ),
        executed_duration_ms: clampDurationMs(
          record.executed_duration_ms ?? record.duration_ms ?? 0,
        ),
        exit_status: clampDurationMs(record.exit_status ?? 0),
        artifact: record.artifact ?? relToRepo(path.dirname(next)),
      });
    }
  }
  return records.sort((left, right) =>
    `${left.execution_group}:${left.shared_report}`.localeCompare(
      `${right.execution_group}:${right.shared_report}`,
    ),
  );
}

export function buildSharedExecutionGroups() {
  const byGroup = new Map();
  for (const record of loadSharedExecutionRecords()) {
    if (!byGroup.has(record.execution_group)) {
      byGroup.set(record.execution_group, []);
    }
    byGroup.get(record.execution_group).push(record);
  }

  return [...byGroup.entries()]
    .sort((left, right) => left[0].localeCompare(right[0]))
    .map(([name, records]) => {
      const startTime =
        records
          .map((record) => record.start_time)
          .filter((value) => Number.isFinite(Date.parse(value)))
          .sort((left, right) => Date.parse(left) - Date.parse(right))[0] ?? "";
      const endTimes = records
        .map((record) => record.end_time)
        .filter((value) => Number.isFinite(Date.parse(value)))
        .sort((left, right) => Date.parse(left) - Date.parse(right));
      const endTime = endTimes[endTimes.length - 1] ?? "";
      const wallDurationMs = disjointSpanDurationMs(
        records.map((record) => ({
          start_time: record.start_time,
          end_time: record.end_time,
          duration_ms: record.wall_duration_ms,
        })),
      );
      const executedDurationMs = records.reduce(
        (total, record) => total + clampDurationMs(record.executed_duration_ms),
        0,
      );
      const failed = records.some((record) =>
        timingStatusFailed(record.status),
      );
      const failureFields = failureFieldsForJSON(
        failed
          ? [
              {
                failure_class: "harness",
                kind: "shared",
                source: "shared-execution",
                label: name,
                message: `shared execution group failed: ${name}`,
              },
            ]
          : [],
      );
      return {
        schema_id: sharedExecutionGroupSchemaID,
        name,
        status: failed ? "fail" : "pass",
        start_time: startTime,
        end_time: endTime,
        wall_duration_ms: wallDurationMs,
        critical_path_wall_duration_ms: wallDurationMs,
        executed_duration_ms: executedDurationMs,
        shared_reports: records
          .map((record) => record.shared_report)
          .filter(Boolean)
          .sort(),
        reports: records.length,
        ...failureFields,
      };
    });
}
