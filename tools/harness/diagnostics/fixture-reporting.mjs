import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";

import { loadServiceJournalEvents } from "../services/journal-reader.mjs";

export const fixtureReportSchemaID = "cartulary.fixture_report.v1";
export const defaultFixtureThresholdMS = 30000;
export const defaultFixtureTop = 5;

const fixtureEventTypes = new Set([
  "postgres-db-created",
  "postgres-db-dropped",
  "postgres-db-migrated",
  "postgres-transaction",
  "s3-bucket-created",
  "s3-bucket-cleaned",
  "s3-prefix-cleaned",
]);

export function clampDurationMs(value) {
  if (!Number.isFinite(value) || value < 0) {
    return 0;
  }
  return value;
}

export function normalizeFixtureThreshold(value) {
  const parsed = Number.parseInt(String(value ?? ""), 10);
  return Number.isInteger(parsed) && parsed >= 0 ? parsed : defaultFixtureThresholdMS;
}

export function normalizeFixtureTop(value) {
  const parsed = Number.parseInt(String(value ?? ""), 10);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : defaultFixtureTop;
}

export function formatDuration(durationMs) {
  if (!Number.isFinite(durationMs) || durationMs < 0) {
    return "0ms";
  }
  if (durationMs < 1000) {
    return `${Math.round(durationMs)}ms`;
  }
  const seconds = durationMs / 1000;
  if (seconds < 60) {
    return `${seconds.toFixed(seconds >= 10 ? 1 : 2)}s`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainder = seconds - minutes * 60;
  return `${minutes}m${remainder.toFixed(1)}s`;
}

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

export function relToRepo(repoRoot, value) {
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

export function fixtureOperationForEvent(type) {
  switch (type) {
    case "postgres-db-created":
      return "database-create";
    case "postgres-db-dropped":
      return "database-drop";
    case "postgres-db-migrated":
      return "database-migrate";
    case "postgres-transaction":
      return "transaction";
    case "s3-bucket-created":
      return "bucket-create";
    case "s3-bucket-cleaned":
      return "bucket-clean";
    case "s3-prefix-cleaned":
      return "prefix-clean";
    default:
      return type;
  }
}

export function loadServiceFixtureEvents({
  resultsRoot,
  runId,
  target = "",
  repoRoot = process.cwd(),
} = {}) {
  const events = [];
  for (const { event, suiteRoot, journalPath } of loadServiceJournalEvents({
    resultsRoot,
    runId,
  })) {
    if (!fixtureEventTypes.has(event.type)) continue;
    const details = event.details ?? {};
    if (target && details.target !== target) continue;
    events.push({
      event,
      path: journalPath,
      artifact: relToRepo(repoRoot, suiteRoot),
    });
  }
  events.sort(
    (left, right) =>
      String(left.event.timestamp ?? "").localeCompare(String(right.event.timestamp ?? "")) ||
      String(left.event.name ?? "").localeCompare(String(right.event.name ?? "")),
  );
  return events;
}

export function loadServiceFixtureActivities(options = {}) {
  const activities = [];
  for (const { event, artifact } of loadServiceFixtureEvents(options)) {
    const details = event.details ?? {};
    activities.push({
      service: event.type.startsWith("postgres-") ? "postgres" : "object_store",
      operation: fixtureOperationForEvent(event.type),
      name: event.name ?? "",
      strategy: details.strategy ?? details.preparation_strategy ?? "",
      fixture_policy: details.fixture_policy ?? "",
      fixture_class: details.fixture_class ?? "",
      reuse_scope: details.reuse_scope ?? "per-test",
      reuse_group: details.reuse_group ?? "",
      caller_package: details.caller_package ?? "",
      caller_file: details.caller_file ?? "",
      test_name: details.test_name ?? "",
      target: details.target ?? "",
      duration_ms: clampDurationMs(details.duration_ms ?? 0),
      artifact,
    });
  }
  return activities;
}

export function emptyFixtureSummary(target) {
  return {
    target,
    total_count: 0,
    total_duration_ms: 0,
    by_package: [],
    by_test: [],
    by_strategy: [],
    slowest: [],
  };
}

export function normalizeFixtureSummary(target, fixture) {
  if (!fixture) {
    return emptyFixtureSummary(target);
  }
  return {
    target: fixture.target ?? target,
    total_count: clampDurationMs(fixture.total_count ?? 0),
    total_duration_ms: clampDurationMs(fixture.total_duration_ms ?? 0),
    by_package: fixture.by_package ?? [],
    by_test: fixture.by_test ?? [],
    by_strategy: fixture.by_strategy ?? [],
    slowest: fixture.slowest ?? [],
  };
}

export function summarizeFixtureActivityList(target, activities) {
  const byPackage = new Map();
  const byTest = new Map();
  const byStrategy = new Map();
  let totalDurationMs = 0;

  for (const activity of activities) {
    totalDurationMs += activity.duration_ms;
    addFixtureAggregate(byPackage, [
      activity.service,
      activity.operation,
      activity.reuse_scope,
      activity.fixture_policy,
      activity.fixture_class,
      activity.caller_package,
    ], activity, {
      service: activity.service,
      operation: activity.operation,
      reuse_scope: activity.reuse_scope,
      fixture_policy: activity.fixture_policy,
      fixture_class: activity.fixture_class,
      caller_package: activity.caller_package,
    });
    addFixtureAggregate(byTest, [
      activity.service,
      activity.operation,
      activity.reuse_scope,
      activity.fixture_policy,
      activity.fixture_class,
      activity.test_name,
    ], activity, {
      service: activity.service,
      operation: activity.operation,
      reuse_scope: activity.reuse_scope,
      fixture_policy: activity.fixture_policy,
      fixture_class: activity.fixture_class,
      test_name: activity.test_name,
    });
    addFixtureAggregate(byStrategy, [
      activity.service,
      activity.operation,
      activity.reuse_scope,
      activity.strategy,
      activity.fixture_policy,
      activity.fixture_class,
    ], activity, {
      service: activity.service,
      operation: activity.operation,
      reuse_scope: activity.reuse_scope,
      strategy: activity.strategy,
      fixture_policy: activity.fixture_policy,
      fixture_class: activity.fixture_class,
    });
  }

  const slowest = [...activities]
    .sort(
      (left, right) =>
        right.duration_ms - left.duration_ms ||
        fixtureSortKey(left).localeCompare(fixtureSortKey(right)),
    )
    .slice(0, 10);
  return {
    target,
    total_count: activities.length,
    total_duration_ms: totalDurationMs,
    by_package: sortedFixtureAggregates(byPackage),
    by_test: sortedFixtureAggregates(byTest),
    by_strategy: sortedFixtureAggregates(byStrategy),
    slowest,
  };
}

export function summarizeFixtureActivities(target, options) {
  return summarizeFixtureActivityList(target, loadServiceFixtureActivities({ ...options, target }));
}

export function fixtureSummariesFromActivities(activities) {
  const byTarget = new Map();
  for (const activity of activities) {
    const target = activity.target || "unknown";
    if (!byTarget.has(target)) {
      byTarget.set(target, []);
    }
    byTarget.get(target).push(activity);
  }
  return [...byTarget.entries()]
    .map(([target, targetActivities]) => summarizeFixtureActivityList(target, targetActivities))
    .sort((left, right) => left.target.localeCompare(right.target));
}

function addFixtureAggregate(map, keyParts, activity, base) {
  const key = keyParts.join("\u001f");
  if (!map.has(key)) {
    map.set(key, { ...base, count: 0, total_duration_ms: 0 });
  }
  const aggregate = map.get(key);
  aggregate.count += 1;
  aggregate.total_duration_ms += activity.duration_ms;
}

export function sortedFixtureAggregates(map) {
  return [...map.values()].sort(
    (left, right) =>
      right.total_duration_ms - left.total_duration_ms ||
      right.count - left.count ||
      fixtureSortKey(left).localeCompare(fixtureSortKey(right)),
  );
}

export function fixtureSortKey(value) {
  return [
    value.service ?? "",
    value.operation ?? "",
    value.strategy ?? "",
    value.fixture_policy ?? "",
    value.fixture_class ?? "",
    value.reuse_scope ?? "",
    value.reuse_group ?? "",
    value.caller_package ?? "",
    value.test_name ?? "",
  ].join("\u001f");
}

export function combineFixtureSummaries(target, ownFixture, childTargets) {
  const byPackage = new Map();
  const byTest = new Map();
  const byStrategy = new Map();
  const slowest = [];
  const combined = {
    target,
    total_count: 0,
    total_duration_ms: 0,
    by_package: [],
    by_test: [],
    by_strategy: [],
    slowest,
  };

  for (const fixture of [ownFixture, ...childTargets.map((child) => child.fixture)]) {
    if (!fixture) {
      continue;
    }
    combined.total_count += fixture.total_count ?? 0;
    combined.total_duration_ms += fixture.total_duration_ms ?? 0;
    mergeFixtureAggregateList(byPackage, fixture.by_package ?? []);
    mergeFixtureAggregateList(byTest, fixture.by_test ?? []);
    mergeFixtureAggregateList(byStrategy, fixture.by_strategy ?? []);
    slowest.push(...(fixture.slowest ?? []));
  }

  combined.by_package = sortedFixtureAggregates(byPackage);
  combined.by_test = sortedFixtureAggregates(byTest);
  combined.by_strategy = sortedFixtureAggregates(byStrategy);
  combined.slowest = slowest
    .sort(
      (left, right) =>
        (right.duration_ms ?? right.total_duration_ms ?? 0) -
          (left.duration_ms ?? left.total_duration_ms ?? 0) ||
        fixtureSortKey(left).localeCompare(fixtureSortKey(right)),
    )
    .slice(0, 10);
  return combined;
}

function mergeFixtureAggregateList(map, values) {
  for (const value of values) {
    const key = fixtureSortKey(value);
    if (!map.has(key)) {
      map.set(key, { ...value, count: 0, total_duration_ms: 0 });
    }
    const aggregate = map.get(key);
    aggregate.count += value.count ?? 0;
    aggregate.total_duration_ms += value.total_duration_ms ?? 0;
  }
}

function readJsonIfExists(filePath) {
  if (!existsSync(filePath)) {
    return null;
  }
  return JSON.parse(readFileSync(filePath, "utf8"));
}

function runSummaryPath(resultsRoot, runId) {
  return path.join(resultsRoot, runId, "run-summary.json");
}

function targetSummaryPath(resultsRoot, runId, target) {
  return path.join(resultsRoot, runId, target, "target-summary.json");
}

function hasFixtureSummary(value) {
  return value && typeof value === "object";
}

function loadRunSummary({ resultsRoot, runId } = {}) {
  return readJsonIfExists(runSummaryPath(resultsRoot, runId));
}

function loadTargetFixtureSummary({ resultsRoot, runId, target } = {}) {
  const summary = readJsonIfExists(targetSummaryPath(resultsRoot, runId, target));
  const fixture = summary?.totals?.fixture ?? summary?.fixture;
  return hasFixtureSummary(fixture) ? normalizeFixtureSummary(target, fixture) : null;
}

function loadFixtureSummariesFromTargetSummaries({ resultsRoot, runId, target = "" } = {}) {
  const runRoot = path.join(resultsRoot, runId);
  if (!existsSync(runRoot)) {
    return [];
  }
  const summaries = [];
  for (const entry of readdirSync(runRoot, { withFileTypes: true })) {
    if (!entry.isDirectory() || entry.name === "_shared") {
      continue;
    }
    if (target && entry.name !== target) {
      continue;
    }
    const summaryPath = path.join(runRoot, entry.name, "target-summary.json");
    if (!existsSync(summaryPath)) {
      continue;
    }
    const summary = JSON.parse(readFileSync(summaryPath, "utf8"));
    const fixture = summary.totals?.fixture ?? summary.fixture;
    if (hasFixtureSummary(fixture)) {
      summaries.push(normalizeFixtureSummary(entry.name, fixture));
    }
  }
  return summaries.sort((left, right) => left.target.localeCompare(right.target));
}

export function newestRunID(resultsRoot) {
  if (!existsSync(resultsRoot)) {
    return "";
  }
  const runs = readdirSync(resultsRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => {
      const runPath = path.join(resultsRoot, entry.name);
      return {
        name: entry.name,
        mtimeMs: statSync(runPath).mtimeMs,
        hasRunSummary: existsSync(path.join(runPath, "run-summary.json")),
      };
    })
    .filter((entry) => entry.hasRunSummary)
    .sort((left, right) => right.mtimeMs - left.mtimeMs || right.name.localeCompare(left.name));
  return runs[0]?.name ?? "";
}

export function resolveFixtureResultLocation({
  resultsDir,
  runId = "",
  repoRoot = process.cwd(),
} = {}) {
  const configured = resultsDir || ".cartulary/test-results";
  const absoluteResultsDir = path.isAbsolute(configured)
    ? configured
    : path.join(repoRoot, configured);
  const concreteRunSummaryPath = path.join(absoluteResultsDir, "run-summary.json");

  if (existsSync(concreteRunSummaryPath)) {
    const derivedRunId = path.basename(absoluteResultsDir);
    if (runId && runId !== derivedRunId) {
      throw new Error(
        `RESULTS_DIR points to run ${derivedRunId}, but RUN_ID requested ${runId}`,
      );
    }
    return {
      resultsRoot: path.dirname(absoluteResultsDir),
      runId: derivedRunId,
      runDir: absoluteResultsDir,
    };
  }

  const resolvedRunId = runId || newestRunID(absoluteResultsDir);
  if (!resolvedRunId) {
    throw new Error(`no test runs found under ${absoluteResultsDir}`);
  }

  return {
    resultsRoot: absoluteResultsDir,
    runId: resolvedRunId,
    runDir: path.join(absoluteResultsDir, resolvedRunId),
  };
}

export function buildFixtureReport({
  resultsRoot,
  runId,
  target = "",
  thresholdMs = defaultFixtureThresholdMS,
  repoRoot = process.cwd(),
} = {}) {
  const runSummary = loadRunSummary({ resultsRoot, runId });
  const runFixture = hasFixtureSummary(runSummary?.fixture)
    ? normalizeFixtureSummary((runSummary.label ?? target) || "all", runSummary.fixture)
    : null;
  let targets = [];
  let aggregate = null;

  if (target) {
    if (runSummary?.label === target && runFixture) {
      aggregate = normalizeFixtureSummary(target, runFixture);
    } else {
      const targetFixture = loadTargetFixtureSummary({ resultsRoot, runId, target });
      if (targetFixture) {
        targets = [targetFixture];
        aggregate = targetFixture;
      }
    }
  } else {
    targets = loadFixtureSummariesFromTargetSummaries({ resultsRoot, runId });
    aggregate = runFixture;
  }

  if (!aggregate) {
    const activities = loadServiceFixtureActivities({ resultsRoot, runId, target, repoRoot });
    if (targets.length === 0 && activities.length > 0) {
      targets = fixtureSummariesFromActivities(activities);
    }
    if (targets.length === 0 && !target) {
      targets = loadFixtureSummariesFromTargetSummaries({ resultsRoot, runId });
    }
    aggregate = combineFixtureSummaries(
      target || "all",
      null,
      targets.map((fixture) => ({ fixture })),
    );
  }

  const runDir = path.join(resultsRoot, runId);
  return {
    schema_id: fixtureReportSchemaID,
    results_dir: resultsRoot,
    results_root: resultsRoot,
    run_dir: runDir,
    run_id: runId,
    threshold_ms: thresholdMs,
    targets,
    aggregate,
  };
}

function formatFixtureComponent(value) {
  const text = String(value ?? "").trim();
  return text === "" ? "-" : text.replace(/\s+/g, "_");
}

function fixtureDisplayName(activity) {
  return formatFixtureComponent(
    activity.test_name || activity.caller_package || activity.caller_file || activity.name,
  );
}

export function fixtureHotspotsText(summary, { top = defaultFixtureTop } = {}) {
  const limit = Math.min(normalizeFixtureTop(top), 3);
  const hotspots = [...(summary?.by_package ?? [])]
    .filter((hotspot) => (hotspot.total_duration_ms ?? 0) > 0)
    .sort(
      (left, right) =>
        (right.total_duration_ms ?? 0) - (left.total_duration_ms ?? 0) ||
        (right.count ?? 0) - (left.count ?? 0) ||
        fixtureSortKey(left).localeCompare(fixtureSortKey(right)),
    )
    .slice(0, limit)
    .map((hotspot) => {
      const label = [
        formatFixtureComponent(hotspot.caller_package),
        formatFixtureComponent(hotspot.service),
        formatFixtureComponent(hotspot.operation),
        formatFixtureComponent(hotspot.fixture_policy),
        formatFixtureComponent(hotspot.reuse_scope),
      ].join("/");
      return `${label}(${formatDuration(hotspot.total_duration_ms ?? 0)},count=${hotspot.count ?? 0})`;
    });
  return hotspots.join(",") || "none";
}

export function fixtureSummaryLine(summary, { thresholdMs, top = defaultFixtureTop } = {}) {
  const threshold = normalizeFixtureThreshold(thresholdMs);
  if ((summary?.total_duration_ms ?? 0) < threshold) {
    return "";
  }
  const topStrategy = summary.by_strategy?.[0];
  const topStrategyText = topStrategy
    ? `${formatFixtureComponent(topStrategy.service)}/${formatFixtureComponent(topStrategy.operation)}/${formatFixtureComponent(topStrategy.fixture_policy)}/${formatFixtureComponent(topStrategy.reuse_scope)} count=${topStrategy.count ?? 0} duration=${formatDuration(topStrategy.total_duration_ms ?? 0)}`
    : "none count=0 duration=0ms";
  const slowest = (summary.slowest ?? [])
    .slice(0, normalizeFixtureTop(top))
    .map((activity) => `${fixtureDisplayName(activity)}(${formatDuration(activity.duration_ms ?? activity.total_duration_ms ?? 0)})`)
    .join(",");
  return `[FIXTURE] ${summary.target} total=${formatDuration(summary.total_duration_ms ?? 0)} count=${summary.total_count ?? 0} top_strategy=${topStrategyText} hotspots=${fixtureHotspotsText(summary, { top })} slowest=${slowest || "none"}`;
}

export function fixtureSummaryLines(summaries, options = {}) {
  return summaries
    .map((summary) => fixtureSummaryLine(summary, options))
    .filter((line) => line.length > 0);
}
