import {
  appendFileSync,
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
  if (!existsSync(file)) {
    throw new Error(`missing shared report metadata for ${sharedName}`);
  }
  const metadata = readFileSync(file, "utf8").trimEnd().split(/\r?\n/u);
  if (metadata.length !== 2) {
    throw new Error(`incomplete shared report metadata for ${sharedName}`);
  }
  return { reportDir: metadata[0], usage: metadata[1] };
}

function sharedReportMetadata(metadataDir, sharedName, metadataByShard) {
  if (!metadataByShard) {
    return readSharedReportMetadata(metadataDir, sharedName);
  }
  if (!metadataByShard.has(sharedName)) {
    metadataByShard.set(sharedName, readSharedReportMetadata(metadataDir, sharedName));
  }
  return metadataByShard.get(sharedName);
}

function isoWindowDurationMs(start, end) {
  const duration = Date.parse(end) - Date.parse(start);
  return Number.isFinite(duration) && duration > 0 ? duration : 0;
}

export function createAggregateReport(
  ctx,
  metadataDir,
  aggregateName,
  target,
  shardNames,
  metadataByShard = null,
) {
  const aggregateRoot = path.join(metadataDir, "aggregate-reports", target);
  const outputDir = path.join(aggregateRoot, aggregateName);
  const runnerLog = path.join(outputDir, "runner.jsonl");
  const stderrLog = path.join(outputDir, "stderr.log");
  const commandFile = path.join(outputDir, "command.txt");
  secureMkdir(outputDir);
  secureWriteFile(runnerLog, "");
  secureWriteFile(stderrLog, "");
  secureWriteFile(commandFile, "");

  let startTime = "";
  let endTime = "";
  let durationMs = 0;
  let actualStartTime = "";
  let actualEndTime = "";
  let actualDurationMs = 0;
  let wallDurationMs = 0;
  let exitStatus = 0;
  let hasActual = false;
  let wroteCommand = false;

  for (const shardName of shardNames) {
    const metadata = sharedReportMetadata(metadataDir, shardName, metadataByShard);
    let usage = metadata.usage;
    if (
      usage === "actual" &&
      target === "backend-integration-support" &&
      isCrossTargetSharedReport(ctx, target, shardName)
    ) {
      usage = "reused";
    }
    if (existsSync(path.join(metadata.reportDir, "runner.jsonl"))) {
      appendFileSync(
        runnerLog,
        readFileSync(path.join(metadata.reportDir, "runner.jsonl")),
      );
    }
    if (existsSync(path.join(metadata.reportDir, "stderr.log"))) {
      appendFileSync(
        stderrLog,
        readFileSync(path.join(metadata.reportDir, "stderr.log")),
      );
    }
    if (wroteCommand) {
      appendFileSync(commandFile, "\n");
    }
    appendFileSync(
      commandFile,
      `${shardName}: ${readFileSync(path.join(metadata.reportDir, "command.txt"), "utf8").trimEnd()}\n`,
    );
    wroteCommand = true;

    const shardDuration = clampDurationMs(
      readFileSync(path.join(metadata.reportDir, "duration_ms.txt"), "utf8"),
    );
    durationMs += shardDuration;
    const shardStatus =
      Number.parseInt(
        readFileSync(path.join(metadata.reportDir, "exit_status.txt"), "utf8"),
        10,
      ) || 0;
    if (shardStatus !== 0) {
      exitStatus = shardStatus;
    }
    const shardStart = readFileSync(
      path.join(metadata.reportDir, "start_time.txt"),
      "utf8",
    ).trim();
    const shardEnd = readFileSync(
      path.join(metadata.reportDir, "end_time.txt"),
      "utf8",
    ).trim();
    if (startTime === "" || shardStart < startTime) {
      startTime = shardStart;
    }
    if (endTime === "" || shardEnd > endTime) {
      endTime = shardEnd;
    }
    if (usage === "actual") {
      hasActual = true;
      actualDurationMs += shardDuration;
      if (actualStartTime === "" || shardStart < actualStartTime) {
        actualStartTime = shardStart;
      }
      if (actualEndTime === "" || shardEnd > actualEndTime) {
        actualEndTime = shardEnd;
      }
    }
  }

  const usage = hasActual ? "actual" : "reused";
  if (hasActual) {
    startTime = actualStartTime;
    endTime = actualEndTime;
    durationMs = actualDurationMs;
    wallDurationMs = isoWindowDurationMs(startTime, endTime);
  }
  secureWriteFile(path.join(outputDir, "start_time.txt"), `${startTime}\n`);
  secureWriteFile(path.join(outputDir, "end_time.txt"), `${endTime}\n`);
  secureWriteFile(
    path.join(outputDir, "duration_ms.txt"),
    `${clampDurationMs(durationMs)}\n`,
  );
  secureWriteFile(
    path.join(outputDir, "wall_duration_ms.txt"),
    `${clampDurationMs(wallDurationMs)}\n`,
  );
  secureWriteFile(path.join(outputDir, "exit_status.txt"), `${exitStatus}\n`);
  secureWriteFile(
    path.join(outputDir, "aggregate.txt"),
    `${target}:${aggregateName}\n`,
  );
  return { reportDir: outputDir, usage };
}

export function createUnshardedFamilyReport(ctx, family, entries) {
  if (entries.length === 0) throw new Error(`${family} has no physical capture reports`);
  const outputDir = prepareSharedArtifactDir(ctx, family);
  const runnerLog = path.join(outputDir, "runner.jsonl");
  const stderrLog = path.join(outputDir, "stderr.log");
  const commandFile = path.join(outputDir, "command.txt");
  secureWriteFile(runnerLog, "");
  secureWriteFile(stderrLog, "");
  secureWriteFile(commandFile, "");
  let startTime = "";
  let endTime = "";
  let durationMs = 0;
  let actualStartTime = "";
  let actualEndTime = "";
  let actualDurationMs = 0;
  let exitStatus = 0;
  let hasActual = false;
  for (const [index, entry] of entries.entries()) {
    const reportDir = entry.reportDir;
    if (existsSync(path.join(reportDir, "runner.jsonl"))) appendFileSync(runnerLog, readFileSync(path.join(reportDir, "runner.jsonl")));
    if (existsSync(path.join(reportDir, "stderr.log"))) appendFileSync(stderrLog, readFileSync(path.join(reportDir, "stderr.log")));
    if (index > 0) appendFileSync(commandFile, "\n");
    appendFileSync(
      commandFile,
      `${entry.name}: ${readFileSync(path.join(reportDir, "command.txt"), "utf8").trimEnd()}\n`,
    );
    const reportDuration = clampDurationMs(readFileSync(path.join(reportDir, "duration_ms.txt"), "utf8"));
    const reportStart = readFileSync(path.join(reportDir, "start_time.txt"), "utf8").trim();
    const reportEnd = readFileSync(path.join(reportDir, "end_time.txt"), "utf8").trim();
    const reportStatus = Number.parseInt(readFileSync(path.join(reportDir, "exit_status.txt"), "utf8"), 10) || 0;
    durationMs += reportDuration;
    if (startTime === "" || reportStart < startTime) startTime = reportStart;
    if (endTime === "" || reportEnd > endTime) endTime = reportEnd;
    if (reportStatus !== 0 && exitStatus === 0) exitStatus = reportStatus;
    if (entry.usage === "actual") {
      hasActual = true;
      actualDurationMs += reportDuration;
      if (actualStartTime === "" || reportStart < actualStartTime) actualStartTime = reportStart;
      if (actualEndTime === "" || reportEnd > actualEndTime) actualEndTime = reportEnd;
    }
  }
  const usage = hasActual ? "actual" : "reused";
  if (hasActual) {
    startTime = actualStartTime;
    endTime = actualEndTime;
    durationMs = actualDurationMs;
  }
  secureWriteFile(path.join(outputDir, "start_time.txt"), `${startTime}\n`);
  secureWriteFile(path.join(outputDir, "end_time.txt"), `${endTime}\n`);
  secureWriteFile(path.join(outputDir, "duration_ms.txt"), `${clampDurationMs(durationMs)}\n`);
  secureWriteFile(
    path.join(outputDir, "wall_duration_ms.txt"),
    `${clampDurationMs(hasActual ? isoWindowDurationMs(startTime, endTime) : 0)}\n`,
  );
  secureWriteFile(path.join(outputDir, "exit_status.txt"), `${exitStatus}\n`);
  return { reportDir: outputDir, usage };
}

export function loadStepWindow(reportDir, mode) {
  const command = readFileSync(
    path.join(reportDir, "command.txt"),
    "utf8",
  ).trimEnd();
  const exitStatus =
    Number.parseInt(
      readFileSync(path.join(reportDir, "exit_status.txt"), "utf8"),
      10,
    ) || 0;
  const storedDurationMs = clampDurationMs(
    readFileSync(path.join(reportDir, "duration_ms.txt"), "utf8"),
  );
  const storedWallDurationMs = existsSync(
    path.join(reportDir, "wall_duration_ms.txt"),
  )
    ? clampDurationMs(
        readFileSync(path.join(reportDir, "wall_duration_ms.txt"), "utf8"),
      )
    : storedDurationMs;
  if (mode === "actual") {
    return {
      command,
      exitStatus,
      startTime: readFileSync(
        path.join(reportDir, "start_time.txt"),
        "utf8",
      ).trim(),
      endTime: readFileSync(
        path.join(reportDir, "end_time.txt"),
        "utf8",
      ).trim(),
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
