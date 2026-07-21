import {
  existsSync,
  readFileSync,
} from "node:fs";
import path from "node:path";

import {
  secureMkdir,
  secureWriteFile,
} from "../../contract/index.mjs";
import { prepareSharedArtifactDir } from "./context.mjs";
import { isCrossTargetSharedReport } from "./planning.mjs";
import { clampDurationMs, nowUTC } from "./util.mjs";

function readSharedReportMetadata(metadataDir, sharedName) {
  const file = path.join(metadataDir, `${sharedName}.meta`);
  if (!existsSync(file)) throw new Error(`missing shared report metadata for ${sharedName}`);
  const metadata = readFileSync(file, "utf8").trimEnd().split(/\r?\n/u);
  if (metadata.length !== 2) throw new Error(`incomplete shared report metadata for ${sharedName}`);
  return Object.freeze({ reportDir: metadata[0], usage: metadata[1] });
}

function sharedReportMetadata(metadataDir, sharedName, metadataByShard) {
  if (!metadataByShard) return readSharedReportMetadata(metadataDir, sharedName);
  if (!metadataByShard.has(sharedName)) {
    metadataByShard.set(sharedName, readSharedReportMetadata(metadataDir, sharedName));
  }
  return metadataByShard.get(sharedName);
}

function isoWindowDurationMs(start, end) {
  const duration = Date.parse(end) - Date.parse(start);
  return Number.isFinite(duration) && duration > 0 ? duration : 0;
}

function readRequired(file, label) {
  if (!existsSync(file)) throw new Error(`${label} is missing`);
  return readFileSync(file, "utf8");
}

function readNonnegativeInteger(file, label) {
  const value = readRequired(file, label).trim();
  if (!/^\d+$/u.test(value)) throw new Error(`${label} is invalid`);
  const parsed = Number(value);
  if (!Number.isSafeInteger(parsed)) throw new Error(`${label} is invalid`);
  return parsed;
}

export function parsePhysicalReport(reportDir) {
  const runnerJSONL = readRequired(path.join(reportDir, "runner.jsonl"), "physical Go runner report");
  for (const [index, rawLine] of runnerJSONL.split(/\r?\n/u).entries()) {
    if (rawLine.trim() === "") continue;
    try {
      JSON.parse(rawLine);
    } catch {
      throw new Error(`physical Go runner report has malformed JSON at line ${index + 1}`);
    }
  }
  const parsed = {
    reportDir,
    runnerJSONL,
    stderr: existsSync(path.join(reportDir, "stderr.log"))
      ? readFileSync(path.join(reportDir, "stderr.log"), "utf8")
      : "",
    command: readRequired(path.join(reportDir, "command.txt"), "physical Go command").trimEnd(),
    startTime: readRequired(path.join(reportDir, "start_time.txt"), "physical Go start time").trim(),
    endTime: readRequired(path.join(reportDir, "end_time.txt"), "physical Go end time").trim(),
    durationMs: readNonnegativeInteger(
      path.join(reportDir, "duration_ms.txt"),
      "physical Go duration",
    ),
    exitStatus: readNonnegativeInteger(
      path.join(reportDir, "exit_status.txt"),
      "physical Go exit status",
    ),
  };
  const startTimeMs = Date.parse(parsed.startTime);
  const endTimeMs = Date.parse(parsed.endTime);
  if (!Number.isFinite(startTimeMs) || !Number.isFinite(endTimeMs) || endTimeMs < startTimeMs) {
    throw new Error("physical Go report has invalid timing boundaries");
  }
  return Object.freeze(parsed);
}

export function parseScheduledPhysicalReports(metadataDir, shardNames, metadataByShard = new Map()) {
  const result = new Map();
  for (const shardName of shardNames) {
    const metadata = sharedReportMetadata(metadataDir, shardName, metadataByShard);
    result.set(shardName, Object.freeze({
      metadata,
      report: parsePhysicalReport(metadata.reportDir),
    }));
  }
  return result;
}

function appendText(parts, value) {
  if (value === "") return;
  parts.push(value.endsWith("\n") ? value : `${value}\n`);
}

function writeMergedReport(outputDir, entries, aggregateIdentity = "") {
  if (entries.length === 0) throw new Error("cannot merge zero physical Go reports");
  secureMkdir(outputDir);
  const runnerParts = [];
  const stderrParts = [];
  const commandParts = [];
  let startTime = "";
  let endTime = "";
  let durationMs = 0;
  let actualStartTime = "";
  let actualEndTime = "";
  let actualDurationMs = 0;
  let exitStatus = 0;
  let hasActual = false;
  for (const entry of entries) {
    const report = entry.report;
    appendText(runnerParts, report.runnerJSONL);
    appendText(stderrParts, report.stderr);
    commandParts.push(`${entry.name}: ${report.command}`);
    durationMs += report.durationMs;
    if (startTime === "" || report.startTime < startTime) startTime = report.startTime;
    if (endTime === "" || report.endTime > endTime) endTime = report.endTime;
    if (report.exitStatus !== 0 && exitStatus === 0) exitStatus = report.exitStatus;
    if (entry.usage === "actual") {
      hasActual = true;
      actualDurationMs += report.durationMs;
      if (actualStartTime === "" || report.startTime < actualStartTime) actualStartTime = report.startTime;
      if (actualEndTime === "" || report.endTime > actualEndTime) actualEndTime = report.endTime;
    }
  }
  const usage = hasActual ? "actual" : "reused";
  if (hasActual) {
    startTime = actualStartTime;
    endTime = actualEndTime;
    durationMs = actualDurationMs;
  }
  secureWriteFile(path.join(outputDir, "runner.jsonl"), runnerParts.join(""));
  secureWriteFile(path.join(outputDir, "stderr.log"), stderrParts.join(""));
  secureWriteFile(path.join(outputDir, "command.txt"), `${commandParts.join("\n\n")}\n`);
  secureWriteFile(path.join(outputDir, "start_time.txt"), `${startTime}\n`);
  secureWriteFile(path.join(outputDir, "end_time.txt"), `${endTime}\n`);
  secureWriteFile(path.join(outputDir, "duration_ms.txt"), `${clampDurationMs(durationMs)}\n`);
  secureWriteFile(
    path.join(outputDir, "wall_duration_ms.txt"),
    `${clampDurationMs(hasActual ? isoWindowDurationMs(startTime, endTime) : 0)}\n`,
  );
  secureWriteFile(path.join(outputDir, "exit_status.txt"), `${exitStatus}\n`);
  if (aggregateIdentity !== "") {
    secureWriteFile(path.join(outputDir, "aggregate.txt"), `${aggregateIdentity}\n`);
  }
  return Object.freeze({ reportDir: outputDir, usage });
}

export function createAggregateReport(
  ctx,
  metadataDir,
  aggregateName,
  target,
  shardNames,
  metadataByShard = null,
  parsedByShard = null,
) {
  const entries = shardNames.map((shardName) => {
    const retained = parsedByShard?.get(shardName);
    const metadata = retained?.metadata ?? sharedReportMetadata(metadataDir, shardName, metadataByShard);
    return Object.freeze({
      name: shardName,
      usage: metadata.usage,
      report: retained?.report ?? parsePhysicalReport(metadata.reportDir),
    });
  });
  return createScheduledAggregateReport(
    ctx,
    metadataDir,
    aggregateName,
    target,
    entries,
  );
}

export function createScheduledAggregateReport(
  ctx,
  metadataDir,
  aggregateName,
  target,
  entries,
) {
  const normalizedEntries = entries.map((entry) => {
    let usage = entry.usage;
    if (
      usage === "actual" &&
      target === "backend-integration-support" &&
      isCrossTargetSharedReport(ctx, target, entry.name)
    ) usage = "reused";
    return Object.freeze({ ...entry, usage });
  });
  const outputDir = path.join(metadataDir, "aggregate-reports", target, aggregateName);
  return writeMergedReport(outputDir, normalizedEntries, `${target}:${aggregateName}`);
}

export function createUnshardedFamilyReport(ctx, family, entries) {
  if (entries.length === 0) throw new Error(`${family} has no physical capture reports`);
  return writeMergedReport(prepareSharedArtifactDir(ctx, family), entries);
}

export function loadStepWindow(reportDir, mode) {
  const command = readFileSync(path.join(reportDir, "command.txt"), "utf8").trimEnd();
  const exitStatus = Number.parseInt(readFileSync(path.join(reportDir, "exit_status.txt"), "utf8"), 10) || 0;
  const storedDurationMs = clampDurationMs(readFileSync(path.join(reportDir, "duration_ms.txt"), "utf8"));
  const storedWallDurationMs = existsSync(path.join(reportDir, "wall_duration_ms.txt"))
    ? clampDurationMs(readFileSync(path.join(reportDir, "wall_duration_ms.txt"), "utf8"))
    : storedDurationMs;
  if (mode === "actual") {
    return {
      command,
      exitStatus,
      startTime: readFileSync(path.join(reportDir, "start_time.txt"), "utf8").trim(),
      endTime: readFileSync(path.join(reportDir, "end_time.txt"), "utf8").trim(),
      durationMs: storedDurationMs,
      wallDurationMs: storedWallDurationMs,
    };
  }
  const timestamp = nowUTC();
  return {
    command,
    exitStatus,
    startTime: timestamp,
    endTime: timestamp,
    durationMs: mode === "reused" ? storedDurationMs : 0,
    wallDurationMs: 0,
  };
}
