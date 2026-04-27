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

function packagePassEvents(events) {
  return events.filter(
    (event) =>
      event.Action === "pass" &&
      typeof event.Package === "string" &&
      event.Test === undefined,
  );
}

function shardMetadata(root) {
  const plan = collectGoShardPlan(root);
  const byShard = new Map();
  for (const shard of plan.shards) {
    byShard.set(shard.name, {
      target: shard.target,
      rawAggregateKey: shard.has_raw ? `${shard.target}::${shard.aggregate_name}` : "",
    });
  }
  return byShard;
}

function aggregateNameForShard(shardName) {
  return shardName.replace(shardNamePattern, "");
}

function observedTests(topLevelEvents) {
  return topLevelEvents
    .map((event) => ({
      key: `${event.Package}::${event.Test}`,
      packageName: event.Package,
      testName: event.Test,
      elapsedMs: Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000)),
    }))
    .sort((left, right) => left.key.localeCompare(right.key));
}

function observedPackageOverheads(target, topLevelEvents, packageEvents) {
  const testElapsedByPackage = new Map();
  for (const event of topLevelEvents) {
    const elapsedMs = Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000));
    testElapsedByPackage.set(event.Package, (testElapsedByPackage.get(event.Package) ?? 0) + elapsedMs);
  }

  return packageEvents
    .map((event) => {
      const packageElapsedMs = Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000));
      const testElapsedMs = testElapsedByPackage.get(event.Package) ?? 0;
      return {
        key: `${target}::${event.Package}`,
        target,
        packageName: event.Package,
        elapsedMs: packageElapsedMs,
        overheadMs: Math.max(1, packageElapsedMs - testElapsedMs),
      };
    })
    .sort((left, right) => left.key.localeCompare(right.key));
}

function commandOverhead(target, shardDurationMs, packageEvents, topLevelEvents) {
  let observedElapsedMs = 0;
  if (packageEvents.length > 0) {
    observedElapsedMs = packageEvents.reduce(
      (sum, event) => sum + Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000)),
      0,
    );
  } else {
    observedElapsedMs = topLevelEvents.reduce(
      (sum, event) => sum + Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000)),
      0,
    );
  }
  return {
    target,
    overheadMs: Math.max(1, shardDurationMs - observedElapsedMs),
  };
}

function observedPackageNames(topLevelEvents, packageEvents) {
  const packages = new Set();
  for (const event of topLevelEvents) {
    packages.add(event.Package);
  }
  for (const event of packageEvents) {
    packages.add(event.Package);
  }
  return packages;
}

export function collectObservedGoShardArtifacts(root, resultsDir) {
  const absoluteResultsDir = path.resolve(resultsDir);
  if (!existsSync(absoluteResultsDir) || !statSync(absoluteResultsDir).isDirectory()) {
    throw new Error(`results directory does not exist: ${resultsDir}`);
  }

  const metadataByShard = shardMetadata(root);
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
    const metadata = metadataByShard.get(shardName);
    if (!metadata) {
      throw new Error(`no Go shard plan metadata for observed shard ${shardName}`);
    }
    const events = readRunnerEvents(path.join(dir, "runner.jsonl"));
    const topLevelEvents = topLevelTestEvents(events);
    const packageEvents = packagePassEvents(events);
    artifacts.push({
      dir,
      shardName,
      aggregateName,
      target: metadata.target,
      durationMs,
      rawAggregateKey: metadata.rawAggregateKey,
      observedTests: observedTests(topLevelEvents),
      observedPackageOverheads: observedPackageOverheads(metadata.target, topLevelEvents, packageEvents),
      observedPackages: observedPackageNames(topLevelEvents, packageEvents),
      commandOverhead: commandOverhead(metadata.target, durationMs, packageEvents, topLevelEvents),
    });
  }
  return artifacts;
}
