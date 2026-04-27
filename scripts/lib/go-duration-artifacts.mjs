import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import path from "node:path";

import { collectGoShardPlan } from "./go-shard-plan.mjs";

const shardNamePattern = /-shard-\d+$/;

export function normalizePositiveInteger(value, fallback = 0) {
  if (Number.isInteger(value) && value > 0) {
    return value;
  }
  return fallback;
}

export function sortedObject(value) {
  return Object.fromEntries(
    Object.entries(value).sort(([left], [right]) => left.localeCompare(right)),
  );
}

function walkDirs(root, dirs = []) {
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    const fullPath = path.join(root, entry.name);
    if (!entry.isDirectory()) {
      continue;
    }
    dirs.push(fullPath);
    walkDirs(fullPath, dirs);
  }
  return dirs;
}

function readIntegerFile(file, fallback = 0) {
  if (!existsSync(file)) {
    return fallback;
  }
  const value = Number(readFileSync(file, "utf8").trim());
  return Number.isInteger(value) ? value : fallback;
}

function readRunnerEvents(runnerLog) {
  const events = [];
  if (!existsSync(runnerLog)) {
    return events;
  }
  for (const rawLine of readFileSync(runnerLog, "utf8").split(/\r?\n/)) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    try {
      events.push(JSON.parse(line));
    } catch {
      // Ignore malformed go test JSON lines; the runner output is advisory here.
    }
  }
  return events;
}

function topLevelTestEvents(events) {
  return events.filter(
    (event) =>
      event.Action === "pass" &&
      typeof event.Package === "string" &&
      typeof event.Test === "string" &&
      event.Test.startsWith("Test") &&
      !event.Test.includes("/"),
  );
}

function rawAggregateTargets(root) {
  const plan = collectGoShardPlan(root);
  const targetsByAggregate = new Map();
  for (const aggregate of plan.aggregates) {
    if (aggregate.coverage === "raw") {
      targetsByAggregate.set(aggregate.name, aggregate.target);
    }
  }
  return targetsByAggregate;
}

function aggregateNameForShard(shardName) {
  return shardName.replace(shardNamePattern, "");
}

function allocateShardWallDuration(topLevelEvents, shardDurationMs) {
  const observed = topLevelEvents
    .map((event) => ({
      key: `${event.Package}::${event.Test}`,
      packageName: event.Package,
      testName: event.Test,
      elapsedMs: Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000)),
    }))
    .sort((left, right) => left.key.localeCompare(right.key));
  if (observed.length === 0) {
    return [];
  }

  const elapsedTotal = observed.reduce((sum, event) => sum + event.elapsedMs, 0);
  const overheadMs = Math.max(0, shardDurationMs - elapsedTotal);
  const allocated = [];
  let allocatedTotal = 0;
  for (let index = 0; index < observed.length; index += 1) {
    const event = observed[index];
    let overheadShare = 0;
    if (overheadMs > 0) {
      if (elapsedTotal > 0) {
        overheadShare = Math.floor((overheadMs * event.elapsedMs) / elapsedTotal);
      } else {
        overheadShare = Math.floor(overheadMs / observed.length);
      }
    }
    const durationMs = event.elapsedMs + overheadShare;
    allocated.push({ ...event, durationMs });
    allocatedTotal += durationMs;
  }

  let remainder = Math.max(0, shardDurationMs - allocatedTotal);
  for (let index = 0; remainder > 0 && allocated.length > 0; index = (index + 1) % allocated.length) {
    allocated[index].durationMs += 1;
    remainder -= 1;
  }
  return allocated;
}

export function collectObservedGoShardArtifacts(root, resultsDir) {
  const absoluteResultsDir = path.resolve(resultsDir);
  if (!existsSync(absoluteResultsDir) || !statSync(absoluteResultsDir).isDirectory()) {
    throw new Error(`results directory does not exist: ${resultsDir}`);
  }

  const rawTargets = rawAggregateTargets(root);
  const shardDirs = walkDirs(absoluteResultsDir).filter((dir) => {
    const shardName = path.basename(dir);
    return (
      path.basename(path.dirname(dir)) === "_shared" &&
      shardNamePattern.test(shardName) &&
      existsSync(path.join(dir, "runner.jsonl"))
    );
  });
  const artifacts = [];
  for (const dir of shardDirs.sort((left, right) => left.localeCompare(right))) {
    const shardName = path.basename(dir);
    const status = readIntegerFile(path.join(dir, "exit_status.txt"), 0);
    if (status !== 0) {
      continue;
    }
    const durationMs = normalizePositiveInteger(readIntegerFile(path.join(dir, "duration_ms.txt"), 0), 1);
    const aggregateName = aggregateNameForShard(shardName);
    const rawTarget = rawTargets.get(aggregateName);
    const events = readRunnerEvents(path.join(dir, "runner.jsonl"));
    const topLevelEvents = topLevelTestEvents(events);
    artifacts.push({
      dir,
      shardName,
      aggregateName,
      durationMs,
      rawAggregateKey: rawTarget ? `${rawTarget}::${aggregateName}` : "",
      observedTests: allocateShardWallDuration(topLevelEvents, durationMs),
      observedPackages: new Set(topLevelEvents.map((event) => event.Package)),
    });
  }
  return artifacts;
}
