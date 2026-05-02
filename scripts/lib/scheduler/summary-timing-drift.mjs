import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { schedulerEventFiles } from "./event-order.mjs";

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function readEvents(file) {
  return readFileSync(file, "utf8")
    .split(/\n/)
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => JSON.parse(line));
}

function timestampMs(value) {
  if (typeof value !== "string" || value === "") {
    return NaN;
  }
  return Date.parse(value);
}

function integerField(record, field) {
  const value = record?.[field];
  return Number.isInteger(value) && value >= 0 ? value : null;
}

function schedulerTiming(summary) {
  const startedMonotonicMs = integerField(
    summary,
    "scheduler_started_monotonic_ms",
  );
  const completedMonotonicMs = integerField(
    summary,
    "scheduler_completed_monotonic_ms",
  );
  const totalDurationMs = integerField(summary, "scheduler_total_duration_ms");
  const startedAt = summary?.scheduler_started_at;
  const completedAt = summary?.scheduler_completed_at;
  if (
    startedMonotonicMs === null ||
    completedMonotonicMs === null ||
    totalDurationMs === null ||
    !Number.isFinite(timestampMs(startedAt)) ||
    !Number.isFinite(timestampMs(completedAt))
  ) {
    return null;
  }
  return {
    startedMonotonicMs,
    completedMonotonicMs,
    totalDurationMs,
    startedAt,
    completedAt,
  };
}

function summaryEndMs(summary) {
  return timestampMs(summary?.completed_at ?? summary?.end_time);
}

function summaryDurationMs(summary) {
  return integerField(summary, "duration_ms") ?? integerField(summary, "wall_duration_ms");
}

function checkElapsedSummary({ errors, file, summary, timing, finalEvent }) {
  if (!existsSync(file)) {
    errors.push(`${file}: missing scheduler-backed summary`);
    return;
  }
  const completedAtMs = summaryEndMs(summary);
  const finalEventMs = timestampMs(finalEvent.wall_timestamp);
  if (!Number.isFinite(completedAtMs)) {
    errors.push(`${file}: missing valid completed_at/end_time`);
  } else if (Number.isFinite(finalEventMs) && completedAtMs < finalEventMs) {
    errors.push(
      `${file}: summary completed before final scheduler event ${finalEvent.wall_timestamp}`,
    );
  }

  const durationMs = summaryDurationMs(summary);
  if (durationMs === null) {
    errors.push(`${file}: missing duration_ms/wall_duration_ms`);
  } else if (durationMs < timing.totalDurationMs) {
    errors.push(
      `${file}: duration ${durationMs}ms is below scheduler total ${timing.totalDurationMs}ms`,
    );
  }

  const criticalPathMs = integerField(summary, "critical_path_wall_duration_ms");
  if (criticalPathMs !== null && criticalPathMs < timing.totalDurationMs) {
    errors.push(
      `${file}: critical path duration ${criticalPathMs}ms is below scheduler total ${timing.totalDurationMs}ms`,
    );
  }
}

function checkToolSummaryExtension({ errors, file, summary, timing }) {
  const extension = summary?.extensions?.scheduler_timing;
  if (!extension) {
    errors.push(`${file}: missing extensions.scheduler_timing`);
    return;
  }
  const extensionTotal = integerField(extension, "scheduler_total_duration_ms");
  if (extensionTotal !== timing.totalDurationMs) {
    errors.push(
      `${file}: extensions.scheduler_timing.scheduler_total_duration_ms ${extensionTotal} does not match scheduler total ${timing.totalDurationMs}ms`,
    );
  }
}

function checkSchedulerDirectory(eventsFile) {
  const errors = [];
  const targetDir = path.dirname(eventsFile);
  const runDir = path.dirname(targetDir);
  const target = path.basename(targetDir);
  const schedulerSummaryFile = path.join(targetDir, "scheduler-summary.json");
  if (!existsSync(schedulerSummaryFile)) {
    return {
      files: [eventsFile],
      errors: [`${schedulerSummaryFile}: missing scheduler summary`],
    };
  }
  const events = readEvents(eventsFile);
  if (events.length === 0) {
    return {
      files: [eventsFile, schedulerSummaryFile],
      errors: [`${eventsFile}: scheduler event stream is empty`],
    };
  }
  const finalEvent = events[events.length - 1];
  const schedulerSummary = readJSON(schedulerSummaryFile);
  const timing = schedulerTiming(schedulerSummary);
  if (!timing) {
    return {
      files: [eventsFile, schedulerSummaryFile],
      errors: [`${schedulerSummaryFile}: missing scheduler timing envelope`],
    };
  }
  const expectedTotal = Math.max(
    0,
    timing.completedMonotonicMs - timing.startedMonotonicMs,
  );
  if (timing.totalDurationMs < expectedTotal) {
    errors.push(
      `${schedulerSummaryFile}: scheduler_total_duration_ms ${timing.totalDurationMs}ms is below completed-start monotonic duration ${expectedTotal}ms`,
    );
  }
  if (Number.isInteger(finalEvent.monotonic_ms)) {
    if (timing.completedMonotonicMs < finalEvent.monotonic_ms) {
      errors.push(
        `${schedulerSummaryFile}: scheduler_completed_monotonic_ms ${timing.completedMonotonicMs} is before final event ${finalEvent.monotonic_ms}`,
      );
    }
    const finalEventTotal = Math.max(
      0,
      finalEvent.monotonic_ms - timing.startedMonotonicMs,
    );
    if (timing.totalDurationMs < finalEventTotal) {
      errors.push(
        `${schedulerSummaryFile}: scheduler_total_duration_ms ${timing.totalDurationMs}ms is below final event duration ${finalEventTotal}ms`,
      );
    }
  }

  const schedulerCompletedAtMs = timestampMs(timing.completedAt);
  const finalEventMs = timestampMs(finalEvent.wall_timestamp);
  if (Number.isFinite(finalEventMs) && schedulerCompletedAtMs < finalEventMs) {
    errors.push(
      `${schedulerSummaryFile}: scheduler_completed_at is before final scheduler event ${finalEvent.wall_timestamp}`,
    );
  }

  const targetSummaryFile = path.join(targetDir, "target-summary.json");
  if (existsSync(targetSummaryFile)) {
    checkElapsedSummary({
      errors,
      file: targetSummaryFile,
      summary: readJSON(targetSummaryFile),
      timing,
      finalEvent,
    });
  } else {
    errors.push(`${targetSummaryFile}: missing scheduler-backed target summary`);
  }

  const targetToolSummaryFile = path.join(targetDir, "tool-run-summary.json");
  if (existsSync(targetToolSummaryFile)) {
    const targetToolSummary = readJSON(targetToolSummaryFile);
    checkElapsedSummary({
      errors,
      file: targetToolSummaryFile,
      summary: targetToolSummary,
      timing,
      finalEvent,
    });
    checkToolSummaryExtension({
      errors,
      file: targetToolSummaryFile,
      summary: targetToolSummary,
      timing,
    });
  } else {
    errors.push(`${targetToolSummaryFile}: missing scheduler-backed tool summary`);
  }

  const runSummaryFile = path.join(runDir, "run-summary.json");
  if (existsSync(runSummaryFile)) {
    const runSummary = readJSON(runSummaryFile);
    if (runSummary.label === target) {
      checkElapsedSummary({
        errors,
        file: runSummaryFile,
        summary: runSummary,
        timing,
        finalEvent,
      });
    }
  }

  const runToolSummaryFile = path.join(runDir, "tool-run-summary.json");
  if (existsSync(runToolSummaryFile)) {
    const runToolSummary = readJSON(runToolSummaryFile);
    if (runToolSummary.target === target) {
      checkElapsedSummary({
        errors,
        file: runToolSummaryFile,
        summary: runToolSummary,
        timing,
        finalEvent,
      });
      checkToolSummaryExtension({
        errors,
        file: runToolSummaryFile,
        summary: runToolSummary,
        timing,
      });
    }
  }

  return {
    files: [
      eventsFile,
      schedulerSummaryFile,
      targetSummaryFile,
      targetToolSummaryFile,
      runSummaryFile,
      runToolSummaryFile,
    ].filter((file) => existsSync(file)),
    errors,
  };
}

export function validateSchedulerSummaryTiming(root, options = {}) {
  const eventFiles = schedulerEventFiles(root, options);
  const checkedFiles = new Set();
  const errors = [];
  for (const eventsFile of eventFiles) {
    const result = checkSchedulerDirectory(eventsFile);
    for (const file of result.files) {
      checkedFiles.add(file);
    }
    errors.push(...result.errors);
  }
  return {
    files: Array.from(checkedFiles).sort((left, right) =>
      left.localeCompare(right),
    ),
    schedulerEventFiles: eventFiles,
    errors,
  };
}
