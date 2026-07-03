import { createWriteStream } from "node:fs";

const defaultSnapshotLimit = 8;
const defaultSlowestLimit = 5;

function countObject(values) {
  if (values instanceof Map) {
    return Object.fromEntries(Array.from(values.entries()).sort((left, right) => left[0].localeCompare(right[0])));
  }
  if (values && typeof values === "object" && !Array.isArray(values)) {
    return Object.fromEntries(Object.entries(values).sort((left, right) => left[0].localeCompare(right[0])));
  }
  return {};
}

function stringArray(values) {
  return Array.isArray(values)
    ? values.filter((value) => typeof value === "string").sort((left, right) => left.localeCompare(right))
    : [];
}

function slowestRunning(value) {
  if (!value || typeof value !== "object") {
    return null;
  }
  const durationMs = Number(value.duration_ms);
  if (typeof value.label !== "string" || !Number.isFinite(durationMs)) {
    return null;
  }
  return {
    label: value.label,
    duration_ms: Math.max(0, durationMs),
  };
}

function compareSlowest(left, right) {
  return (
    right.duration_ms - left.duration_ms ||
    left.source.localeCompare(right.source) ||
    (left.work_unit ?? "").localeCompare(right.work_unit ?? "") ||
    (left.nested_target ?? "").localeCompare(right.nested_target ?? "") ||
    left.label.localeCompare(right.label)
  );
}

export class SchedulerProgressRecorder {
  constructor(file, { snapshotLimit = defaultSnapshotLimit, slowestLimit = defaultSlowestLimit } = {}) {
    this.file = file;
    this.snapshotLimit = snapshotLimit;
    this.slowestLimit = slowestLimit;
    this.stream = createWriteStream(file, { flags: "w" });
    this.snapshots = [];
    this.slowestByKey = new Map();
  }

  appendLine(line) {
    this.stream.write(line.endsWith("\n") ? line : `${line}\n`);
  }

  recordStart(line) {
    this.appendLine(line);
  }

  recordProgress({
    line,
    seq,
    monotonicMs,
    emittedAt,
    completed,
    total,
    running,
    pending,
    blocked,
    finalizing = null,
    activeGroups,
    runningLabels = [],
    blockedBy = [],
    waitingOn = [],
    unblocksAfter = null,
    slowestRunning: outerSlowest = null,
    nestedProgress = [],
  }) {
    const snapshot = {
      seq,
      monotonic_ms: monotonicMs,
      emitted_at: emittedAt,
      completed,
      total_work_units: total,
      running,
      pending,
      blocked,
      finalizing,
      active_groups: countObject(activeGroups),
      running_units: stringArray(runningLabels),
      blocked_by: stringArray(blockedBy),
      waiting_on: stringArray(waitingOn),
      unblocks_after: unblocksAfter,
      slowest_running: slowestRunning(outerSlowest),
      nested_scheduler_progress: nestedProgress.map((progress) => ({
        work_unit: progress.work_unit ?? "",
        nested_target: progress.nested_target ?? "",
        completed: Number.isInteger(progress.completed) ? progress.completed : 0,
        total_work_units: Number.isInteger(progress.total_work_units) ? progress.total_work_units : 0,
        running: Number.isInteger(progress.running) ? progress.running : 0,
        pending: Number.isInteger(progress.pending) ? progress.pending : 0,
        blocked: Number.isInteger(progress.blocked) ? progress.blocked : 0,
        finalizing: Number.isInteger(progress.finalizing) ? progress.finalizing : null,
        active_groups: countObject(progress.active_groups),
        blocked_by: stringArray(progress.blocked_by),
        waiting_on: stringArray(progress.waiting_on),
        unblocks_after: typeof progress.unblocks_after === "string" ? progress.unblocks_after : null,
        slowest_running: slowestRunning(progress.slowest_running),
        artifacts: typeof progress.artifacts === "string" ? progress.artifacts : "",
        events_jsonl: typeof progress.events_jsonl === "string" ? progress.events_jsonl : "",
      })),
      line: line.trimEnd(),
    };
    this.snapshots.push(snapshot);
    if (this.snapshots.length > this.snapshotLimit) {
      this.snapshots.splice(0, this.snapshots.length - this.snapshotLimit);
    }
    this.recordSlowest({
      source: "outer",
      target: "",
      workUnit: "",
      nestedTarget: "",
      value: snapshot.slowest_running,
      seq,
      monotonicMs,
      emittedAt,
    });
    for (const progress of snapshot.nested_scheduler_progress) {
      this.recordSlowest({
        source: "nested",
        target: progress.nested_target,
        workUnit: progress.work_unit,
        nestedTarget: progress.nested_target,
        value: progress.slowest_running,
        seq,
        monotonicMs,
        emittedAt,
      });
    }
    this.appendLine(line);
  }

  recordSlowest({ source, target, workUnit, nestedTarget, value, seq, monotonicMs, emittedAt }) {
    if (!value) {
      return;
    }
    const key = [source, target, workUnit, nestedTarget, value.label].join("\0");
    const previous = this.slowestByKey.get(key);
    if (previous && previous.duration_ms > value.duration_ms) {
      return;
    }
    this.slowestByKey.set(key, {
      source,
      target,
      work_unit: workUnit || null,
      nested_target: nestedTarget || null,
      label: value.label,
      duration_ms: value.duration_ms,
      seq,
      monotonic_ms: monotonicMs,
      emitted_at: emittedAt,
    });
  }

  recordSummary(line) {
    this.appendLine(line);
  }

  summaryFields() {
    return {
      progress_snapshots: this.snapshots,
      slowest_running_observations: Array.from(this.slowestByKey.values()).sort(compareSlowest).slice(0, this.slowestLimit),
    };
  }

  close() {
    return new Promise((resolve, reject) => {
      this.stream.on("error", reject);
      this.stream.end(resolve);
    });
  }
}
