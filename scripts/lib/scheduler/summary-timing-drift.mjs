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
  const finalEventMs = timestampMs(finalEvent.emitted_at);
  if (!Number.isFinite(completedAtMs)) {
    errors.push(`${file}: missing valid completed_at/end_time`);
  } else if (Number.isFinite(finalEventMs) && completedAtMs < finalEventMs) {
    errors.push(
      `${file}: summary completed before final scheduler event `,
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
  const extension = summary?.scheduler_timing;
  if (!extension) {
    errors.push(`${file}: missing scheduler_timing`);
    return;
  }
  const extensionTotal = integerField(extension, "scheduler_total_duration_ms");
  if (extensionTotal !== timing.totalDurationMs) {
    errors.push(
      `${file}: scheduler_timing.scheduler_total_duration_ms ${extensionTotal} does not match scheduler total ${timing.totalDurationMs}ms`,
    );
  }
}

function checkSchedulerCriticalPath({ errors, file, summary, timing, events }) {
  const criticalPathMs = integerField(summary, "critical_path_wall_duration_ms");
  if (criticalPathMs === null) {
    errors.push(`${file}: missing critical_path_wall_duration_ms`);
    return;
  }
  if (criticalPathMs !== timing.totalDurationMs) {
    errors.push(
      `${file}: critical_path_wall_duration_ms ${criticalPathMs}ms does not match scheduler total ${timing.totalDurationMs}ms`,
    );
  }
  const units = summary.critical_path_units;
  if (!Array.isArray(units)) {
    errors.push(`${file}: critical_path_units must be an array`);
    return;
  }
  const terminal = summary.critical_path_terminal_unit;
  if (terminal !== null && (!terminal || typeof terminal !== "object" || Array.isArray(terminal))) {
    errors.push(`${file}: critical_path_terminal_unit must be null or an object`);
    return;
  }
  if (!terminal) {
    const failedOnlySummary =
      summary.status === "fail" &&
      Array.isArray(summary.observed_failed_work_units) &&
      summary.observed_failed_work_units.length > 0 &&
      units.length === 0;
    if ((summary.completed_work_units ?? 0) > 0 && !failedOnlySummary) {
      errors.push(`${file}: critical_path_terminal_unit is missing despite completed work`);
    }
    return;
  }
  const finishEvents = events.filter(
    (event) => event.event === "finish" || event.event === "finalize-finish",
  );
  const startEvents = events.filter(
    (event) => event.event === "start" || event.event === "finalize-start",
  );
  const finishByID = new Map(
    finishEvents.map((event) => [event.work_unit_id ?? event.finalizer_id, event]),
  );
  const startByID = new Map(
    startEvents.map((event) => [event.work_unit_id ?? event.finalizer_id, event]),
  );
  const matchingFinish = finishEvents.find((event) =>
    event.work_unit_id === terminal.id || event.finalizer_id === terminal.id,
  );
  if (!matchingFinish) {
    errors.push(`${file}: critical_path_terminal_unit ${terminal.id} has no finish event`);
  }
  const finishedMs = integerField(terminal, "finished_monotonic_ms");
  if (finishedMs === null) {
    errors.push(`${file}: critical_path_terminal_unit.finished_monotonic_ms is missing`);
    return;
  }
  const terminalWallMs = Math.max(0, finishedMs - timing.startedMonotonicMs);
  if (terminalWallMs > criticalPathMs) {
    errors.push(
      `${file}: terminal critical path wall ${terminalWallMs}ms exceeds critical_path_wall_duration_ms ${criticalPathMs}ms`,
    );
  }
  if (units.length === 0) {
    errors.push(`${file}: critical_path_units is empty despite terminal unit ${terminal.id}`);
    return;
  }
  if (units[units.length - 1]?.id !== terminal.id) {
    errors.push(`${file}: critical_path_terminal_unit ${terminal.id} does not match last critical_path_units entry`);
  }
  let previous = null;
  for (const [index, unit] of units.entries()) {
    if (!unit || typeof unit !== "object" || Array.isArray(unit)) {
      errors.push(`${file}: critical_path_units[${index}] must be an object`);
      continue;
    }
    if (typeof unit.id !== "string" || unit.id.trim() === "") {
      errors.push(`${file}: critical_path_units[${index}].id must be a non-empty string`);
      continue;
    }
    const startMs = integerField(unit, "started_monotonic_ms");
    const finishMs = integerField(unit, "finished_monotonic_ms");
    const durationMs = integerField(unit, "duration_ms");
    if (startMs === null || finishMs === null || durationMs === null) {
      errors.push(`${file}: critical_path_units[${index}] is missing monotonic timing fields`);
      continue;
    }
    if (finishMs < startMs) {
      errors.push(`${file}: critical_path_units[${index}] finishes before it starts`);
    }
    if (durationMs !== Math.max(0, finishMs - startMs)) {
      errors.push(`${file}: critical_path_units[${index}] duration does not match monotonic start/finish`);
    }
    const startEvent = startByID.get(unit.id);
    const finishEvent = finishByID.get(unit.id);
    if (!startEvent) {
      errors.push(`${file}: critical_path_units[${index}] ${unit.id} has no start event`);
    }
    if (!finishEvent) {
      errors.push(`${file}: critical_path_units[${index}] ${unit.id} has no finish event`);
    }
    if (startEvent && finishEvent) {
      const startEventMs = Number.isInteger(startEvent.monotonic_ms) ? startEvent.monotonic_ms : null;
      const finishEventMs = Number.isInteger(finishEvent.monotonic_ms) ? finishEvent.monotonic_ms : null;
      if (startEventMs !== null && finishEventMs !== null && finishEventMs < startEventMs) {
        errors.push(`${file}: critical_path_units[${index}] finish event precedes start event`);
      }
      if (startEventMs !== null && startEventMs > finishMs) {
        errors.push(`${file}: critical_path_units[${index}] retained start event is after unit finish`);
      }
      if (finishEventMs !== null && finishEventMs < startMs) {
        errors.push(`${file}: critical_path_units[${index}] retained finish event is before unit start`);
      }
    }
    if (previous) {
      const previousKeys = new Set(
        Array.isArray(previous.completion_keys) ? previous.completion_keys : [],
      );
      const needs = Array.isArray(unit.needs) ? unit.needs : [];
      if (!needs.some((need) => previousKeys.has(need))) {
        errors.push(`${file}: critical_path_units[${index}] ${unit.id} is not linked to previous unit ${previous.id}`);
      }
      const previousFinish = integerField(previous, "finished_monotonic_ms");
      if (previousFinish !== null && startMs < previousFinish) {
        errors.push(`${file}: critical_path_units[${index}] starts before previous critical-path unit finishes`);
      }
    }
    previous = unit;
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
  checkSchedulerCriticalPath({
    errors,
    file: schedulerSummaryFile,
    summary: schedulerSummary,
    timing,
    events,
  });
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
  const finalEventMs = timestampMs(finalEvent.emitted_at);
  if (Number.isFinite(finalEventMs) && schedulerCompletedAtMs < finalEventMs) {
    errors.push(
      `${schedulerSummaryFile}: scheduler_completed_at is before final scheduler event `,
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

function checkParentWorkUnitSummaries(eventsFile) {
  const errors = [];
  const targetDir = path.dirname(eventsFile);
  const runDir = path.dirname(targetDir);
  const events = readEvents(eventsFile);
  const starts = new Map();
  for (const event of events) {
    if (event.event === "start" && typeof event.work_unit_id === "string") {
      starts.set(event.work_unit_id, event);
    }
  }
  for (const event of events) {
    if (event.event !== "finish" || typeof event.work_unit_id !== "string") {
      continue;
    }
    const start = starts.get(event.work_unit_id);
    if (start?.work_unit_type !== "make_target") {
      continue;
    }
    if (start?.nested_scheduler?.type !== "service_backed") {
      continue;
    }
    const target = start.aggregate_target || start.work_unit || event.work_unit;
    if (typeof target !== "string" || target === "") {
      continue;
    }
    const nestedSchedulerSummaryFile = path.join(
      runDir,
      target,
      "scheduler-summary.json",
    );
    if (!existsSync(nestedSchedulerSummaryFile)) {
      continue;
    }
    const nestedSchedulerSummary = readJSON(nestedSchedulerSummaryFile);
    if (nestedSchedulerSummary?.scheduler_kind !== "service_backed") {
      continue;
    }
    const durationMs = integerField(event, "duration_ms");
    if (durationMs === null) {
      continue;
    }
    const targetSummaryFile = path.join(runDir, target, "target-summary.json");
    if (!existsSync(targetSummaryFile)) {
      continue;
    }
    const summary = readJSON(targetSummaryFile);
    const elapsedMs = summaryDurationMs(summary);
    if (elapsedMs !== null && elapsedMs < durationMs) {
      errors.push(
        `${targetSummaryFile}: duration ${elapsedMs}ms is below parent scheduler work-unit ${event.work_unit_id} duration ${durationMs}ms`,
      );
    }
    const criticalPathMs = integerField(summary, "critical_path_wall_duration_ms");
    if (criticalPathMs !== null && criticalPathMs < durationMs) {
      errors.push(
        `${targetSummaryFile}: critical path duration ${criticalPathMs}ms is below parent scheduler work-unit ${event.work_unit_id} duration ${durationMs}ms`,
      );
    }
  }
  return errors;
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
    errors.push(...checkParentWorkUnitSummaries(eventsFile));
  }
  return {
    files: Array.from(checkedFiles).sort((left, right) =>
      left.localeCompare(right),
    ),
    schedulerEventFiles: eventFiles,
    errors,
  };
}
