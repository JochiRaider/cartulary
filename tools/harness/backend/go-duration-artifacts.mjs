import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import path from "node:path";

import { rawPackageBaselineKey } from "./go-duration-baselines.mjs";
import { collectGoShardPlan } from "./backend-shard-plan.mjs";
import { loadServiceFixtureActivities } from "../diagnostics/fixture-reporting.mjs";

function normalizePositiveInteger(value, fallback = 0) {
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

function walkFiles(root, files = []) {
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    const fullPath = path.join(root, entry.name);
    if (entry.isDirectory()) {
      walkFiles(fullPath, files);
      continue;
    }
    if (entry.isFile()) {
      files.push(fullPath);
    }
  }
  return files;
}

function readJSONArtifact(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`${file} is not readable JSON retained-run evidence: ${message}`);
  }
}

function validateGoDurationRetainedRun(_root, resultsDir) {
  const absoluteResultsDir = path.resolve(resultsDir);
  if (!existsSync(absoluteResultsDir) || !statSync(absoluteResultsDir).isDirectory()) {
    throw new Error(`results directory does not exist: ${resultsDir}`);
  }

  const files = walkFiles(absoluteResultsDir).sort((left, right) => left.localeCompare(right));
  const targetSummaryFiles = files.filter((file) => path.basename(file) === "target-summary.json");
  const schedulerSummaryFiles = files.filter((file) => path.basename(file) === "scheduler-summary.json");
  const schedulerEventFiles = files.filter((file) => path.basename(file) === "scheduler-events.jsonl");
  const executionContextFile = path.join(
    absoluteResultsDir,
    "_shared",
    "harness-observability",
    "execution-context.json",
  );

  if (targetSummaryFiles.length === 0) {
    throw new Error("duration retained run must contain target-summary.json evidence");
  }
  if (schedulerSummaryFiles.length === 0) {
    throw new Error("duration retained run must contain scheduler-summary.json evidence");
  }
  if (schedulerEventFiles.length === 0) {
    throw new Error("duration retained run must contain scheduler-events.jsonl evidence");
  }
  if (!existsSync(executionContextFile)) {
    throw new Error("duration retained run must contain execution-context.json evidence");
  }
  const executionContext = readJSONArtifact(executionContextFile);
  const contaminationReasons = Array.isArray(executionContext.contamination_reasons)
    ? executionContext.contamination_reasons
    : [];
  if (
    executionContext.schema_id !== "cartulary.harness_execution_context.v2" ||
    executionContext.status !== "passed" ||
    executionContext.source_state !== "clean" ||
    executionContext.warm_eligibility !== "eligible" ||
    executionContext.interrupted !== false ||
    executionContext.retry_count !== 0 ||
    contaminationReasons.length > 0
  ) {
    throw new Error(
      `duration retained run execution context is not clean warm evidence: ${contaminationReasons.join(",") || "ineligible_context"}`,
    );
  }

  for (const file of targetSummaryFiles) {
    const summary = readJSONArtifact(file);
    if (summary.status !== "pass") {
      throw new Error(`${file} is not a passing target summary`);
    }
  }
  for (const file of schedulerSummaryFiles) {
    const summary = readJSONArtifact(file);
    if (summary.status !== "pass") {
      throw new Error(`${file} is not a passing scheduler summary`);
    }
    if (
      Number.isInteger(summary.total_work_units) &&
      Number.isInteger(summary.completed_work_units) &&
      summary.completed_work_units < summary.total_work_units
    ) {
      throw new Error(`${file} is incomplete scheduler evidence`);
    }
  }

  return {
    resultsDir: absoluteResultsDir,
    targetSummaryCount: targetSummaryFiles.length,
    schedulerSummaryCount: schedulerSummaryFiles.length,
    schedulerEventCount: schedulerEventFiles.length,
    executionContextFile,
  };
}

function readIntegerFile(file, fallback = 0) {
  if (!existsSync(file)) {
    return fallback;
  }
  const value = Number(readFileSync(file, "utf8").trim());
  return Number.isInteger(value) ? value : fallback;
}

function readTextFile(file) {
  if (!existsSync(file)) {
    return "";
  }
  return readFileSync(file, "utf8");
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
  const targetByTestKey = new Map();
  for (const shard of plan.shards) {
    byShard.set(shard.name, {
      target: shard.target,
      aggregateName: shard.aggregate_name,
      rawAggregateName: shard.has_raw ? shard.aggregate_name : "",
    });
    for (const item of shard.items ?? []) {
      if (item.kind === "raw") {
        continue;
      }
      targetByTestKey.set(item.baseline_key, item.target);
    }
  }
  return { byShard, targetByTestKey };
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

function eventInterval(event) {
  const end = Date.parse(event.Time ?? "");
  const elapsed = Number(event.Elapsed ?? 0) * 1000;
  if (!Number.isFinite(end) || !Number.isFinite(elapsed) || elapsed <= 0) return null;
  return [end - elapsed, end];
}

function packageExecutionUnionMs(events) {
  const intervals = events
    .map(eventInterval)
    .filter((interval) => interval !== null)
    .sort((left, right) => left[0] - right[0] || left[1] - right[1]);
  let total = 0;
  let current = null;
  for (const interval of intervals) {
    if (!current || interval[0] > current[1]) {
      if (current) total += current[1] - current[0];
      current = [...interval];
    } else {
      current[1] = Math.max(current[1], interval[1]);
    }
  }
  if (current) total += current[1] - current[0];
  return Math.round(total);
}

function commandOverhead(target, shardDurationMs, packageEvents, topLevelEvents) {
  const executionEvents = packageEvents.length > 0 ? packageEvents : topLevelEvents;
  const observedElapsedMs = packageExecutionUnionMs(executionEvents);
  if (observedElapsedMs <= 0) {
    throw new Error(`${target} physical Go report has no canonical package timing intervals`);
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

function topLevelTestName(testName) {
  return String(testName ?? "").split("/", 1)[0];
}

function fixtureTimingIndex(root, resultsDir) {
  const runId = path.basename(path.resolve(resultsDir));
  const resultsRoot = path.dirname(path.resolve(resultsDir));
  const byTest = new Map();
  const byPackage = new Map();
  for (const activity of loadServiceFixtureActivities({ resultsRoot, runId, repoRoot: root })) {
    if (!activity.target) {
      continue;
    }
    if (activity.test_name) {
      const key = `${activity.target}::${topLevelTestName(activity.test_name)}`;
      byTest.set(key, (byTest.get(key) ?? 0) + activity.duration_ms);
    }
    if (activity.caller_package) {
      const key = `${activity.target}::${activity.caller_package}`;
      byPackage.set(key, (byPackage.get(key) ?? 0) + activity.duration_ms);
    }
  }
  return { byTest, byPackage };
}

function observedFixtureTests(target, observedTestEntries, fixtureIndex) {
  return observedTestEntries
    .map((test) => ({
      key: test.key,
      target,
      packageName: test.packageName,
      testName: test.testName,
      fixtureMs: fixtureIndex.byTest.get(`${target}::${test.testName}`) ?? 0,
    }))
    .filter((entry) => entry.fixtureMs > 0)
    .sort((left, right) => left.key.localeCompare(right.key));
}

function observedFixturePackages(target, packages, fixtureIndex) {
  return [...packages]
    .map((packageName) => ({
      key: `${target}::${packageName}`,
      target,
      packageName,
      fixtureMs: fixtureIndex.byPackage.get(`${target}::${packageName.replace(/^github\.com\/JochiRaider\/cartulary\//, "")}`) ?? 0,
    }))
    .filter((entry) => entry.fixtureMs > 0)
    .sort((left, right) => left.key.localeCompare(right.key));
}

function observedRawPackages(target, aggregateName, durationMs, topLevelEvents, packageEvents) {
  if (!aggregateName) {
    return [];
  }
  const elapsedByPackage = new Map();
  for (const event of packageEvents) {
    elapsedByPackage.set(
      event.Package,
      Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000)),
    );
  }
  if (elapsedByPackage.size === 0) {
    for (const event of topLevelEvents) {
      const elapsedMs = Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000));
      elapsedByPackage.set(event.Package, (elapsedByPackage.get(event.Package) ?? 0) + elapsedMs);
    }
  }
  if (elapsedByPackage.size === 0) {
    return [];
  }

  const entries = Array.from(elapsedByPackage.entries()).sort(([left], [right]) =>
    left.localeCompare(right),
  );
  const totalElapsedMs = entries.reduce((sum, [, elapsedMs]) => sum + elapsedMs, 0);
  const overheadMs = Math.max(0, durationMs - totalElapsedMs);
  let remainingOverheadMs = overheadMs;
  return entries.map(([packageName, elapsedMs], index) => {
    const overheadShareMs =
      index === entries.length - 1
        ? remainingOverheadMs
        : Math.min(
            remainingOverheadMs,
            Math.round((overheadMs * elapsedMs) / Math.max(1, totalElapsedMs)),
          );
    remainingOverheadMs -= overheadShareMs;
    return {
      key: rawPackageBaselineKey(target, aggregateName, packageName),
      packageName,
      elapsedMs,
      durationMs: elapsedMs + overheadShareMs,
    };
  });
}

function timingContamination(stderrLog) {
  const moduleDownloadCount = readTextFile(stderrLog)
    .split(/\r?\n/)
    .filter((line) => /^go: downloading\s+\S+/.test(line.trim())).length;
  const reasons = [];
  if (moduleDownloadCount > 0) {
    reasons.push("go-module-download");
  }
  return {
    moduleDownloadCount,
    timingContaminationReasons: reasons,
  };
}

export function collectObservedGoShardArtifacts(root, resultsDir) {
  const absoluteResultsDir = path.resolve(resultsDir);
  validateGoDurationRetainedRun(root, absoluteResultsDir);

  const metadata = shardMetadata(root);
  const fixtures = fixtureTimingIndex(root, absoluteResultsDir);
  const shardDirs = walkDirs(absoluteResultsDir).filter((dir) => {
    const shardName = path.basename(dir);
    return (
      path.basename(path.dirname(dir)) === "_shared" &&
      (metadata.byShard.has(shardName) || shardName.includes("-shard-") || shardName.includes("-scn-")) &&
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
    let shardMetadataEntry = metadata.byShard.get(shardName);
    const events = readRunnerEvents(path.join(dir, "runner.jsonl"));
    const topLevelEvents = topLevelTestEvents(events);
    const packageEvents = packagePassEvents(events);
    const observedTestEntries = observedTests(topLevelEvents);
    const contamination = timingContamination(path.join(dir, "stderr.log"));
    if (!shardMetadataEntry) {
      const observedTargets = new Set();
      for (const observedTest of observedTestEntries) {
        const target = metadata.targetByTestKey.get(observedTest.key);
        if (target) {
          observedTargets.add(target);
        }
      }
      if (observedTargets.size === 1) {
        shardMetadataEntry = { target: [...observedTargets][0], rawAggregateName: "" };
      }
    }
    if (!shardMetadataEntry) {
      throw new Error(`no Go shard plan metadata for observed shard ${shardName}`);
    }
    const observedPackages = observedPackageNames(topLevelEvents, packageEvents);
    artifacts.push({
      dir,
      shardName,
      aggregateName: shardMetadataEntry.aggregateName ?? "",
      target: shardMetadataEntry.target,
      durationMs,
      rawAggregateName: shardMetadataEntry.rawAggregateName ?? "",
      observedRawPackages: observedRawPackages(
        shardMetadataEntry.target,
        shardMetadataEntry.rawAggregateName,
        durationMs,
        topLevelEvents,
        packageEvents,
      ),
      observedTests: observedTestEntries,
      observedPackageOverheads: observedPackageOverheads(shardMetadataEntry.target, topLevelEvents, packageEvents),
      observedFixtureTests: observedFixtureTests(shardMetadataEntry.target, observedTestEntries, fixtures),
      observedFixturePackages: observedFixturePackages(shardMetadataEntry.target, observedPackages, fixtures),
      observedPackages,
      commandOverhead: commandOverhead(shardMetadataEntry.target, durationMs, packageEvents, topLevelEvents),
      ...contamination,
    });
  }
  if (artifacts.length === 0) {
    throw new Error("duration retained run has no passing Go shard artifacts");
  }
  return artifacts;
}
