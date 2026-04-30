import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";

export function schedulerEventFiles(root, { target = "" } = {}) {
  const resolved = path.resolve(root);
  if (!existsSync(resolved)) {
    throw new Error(`scheduler event path does not exist: ${root}`);
  }
  if (statSync(resolved).isFile()) {
    return path.basename(resolved) === "scheduler-events.jsonl" ? [resolved] : [];
  }
  const files = [];
  const visit = (dir) => {
    for (const entry of readdirSync(dir, { withFileTypes: true })) {
      const child = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        visit(child);
        continue;
      }
      if (entry.isFile() && entry.name === "scheduler-events.jsonl") {
        if (!target || path.basename(path.dirname(child)) === target) {
          files.push(child);
        }
      }
    }
  };
  visit(resolved);
  return files.sort((left, right) => left.localeCompare(right));
}

function parseWallTimestamp(value) {
  if (typeof value !== "string" || value === "") {
    return NaN;
  }
  return Date.parse(value);
}

export function validateSchedulerEventOrderFile(file) {
  const errors = [];
  const lines = readFileSync(file, "utf8").split(/\n/).filter((line) => line.trim() !== "");
  let previousSequence = 0;
  let previousMonotonicMs = -1;
  let previousWallMs = NaN;
  let lastEventWasClockSkew = false;

  lines.forEach((line, index) => {
    const lineNumber = index + 1;
    let event;
    try {
      event = JSON.parse(line);
    } catch (error) {
      errors.push(`${file}:${lineNumber}: invalid JSON: ${error.message}`);
      return;
    }

    if (!Number.isInteger(event.event_sequence) || event.event_sequence < 1) {
      errors.push(`${file}:${lineNumber}: missing positive integer event_sequence`);
    } else if (event.event_sequence !== previousSequence + 1) {
      errors.push(
        `${file}:${lineNumber}: event_sequence got ${event.event_sequence}, want ${previousSequence + 1}`,
      );
    }
    if (!Number.isInteger(event.monotonic_ms) || event.monotonic_ms < 0) {
      errors.push(`${file}:${lineNumber}: missing non-negative integer monotonic_ms`);
    } else if (event.monotonic_ms < previousMonotonicMs) {
      errors.push(
        `${file}:${lineNumber}: monotonic_ms regressed from ${previousMonotonicMs} to ${event.monotonic_ms}`,
      );
    }
    const wallMs = parseWallTimestamp(event.wall_timestamp);
    if (!Number.isFinite(wallMs)) {
      errors.push(`${file}:${lineNumber}: missing valid wall_timestamp`);
    } else if (Number.isFinite(previousWallMs) && wallMs < previousWallMs && !lastEventWasClockSkew) {
      errors.push(`${file}:${lineNumber}: wall_timestamp regressed without preceding clock-skew marker`);
    }

    previousSequence = Number.isInteger(event.event_sequence) ? event.event_sequence : previousSequence;
    previousMonotonicMs = Number.isInteger(event.monotonic_ms)
      ? Math.max(previousMonotonicMs, event.monotonic_ms)
      : previousMonotonicMs;
    previousWallMs = Number.isFinite(wallMs) ? wallMs : previousWallMs;
    lastEventWasClockSkew = event.event === "clock-skew";
  });

  if (lines.length === 0) {
    errors.push(`${file}: scheduler event stream is empty`);
  }
  return errors;
}

export function validateSchedulerEventOrder(root, options = {}) {
  const files = schedulerEventFiles(root, options);
  const errors = files.flatMap((file) => validateSchedulerEventOrderFile(file));
  return { files, errors };
}

