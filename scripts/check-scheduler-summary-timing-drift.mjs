#!/usr/bin/env node
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

import { collectGoShardPlan } from "./lib/go-shard-plan.mjs";
import { validateSchedulerSummaryTiming } from "./lib/scheduler/summary-timing-drift.mjs";

const warmBalanceMaterialSkewMs = 5000;
const warmPackageResetCountBudget = 30;
const warmPackageResetDurationBudgetMs = 60000;
const warmReadinessThresholds = new Map([
  ["codegen-toolchain", 10000],
  ["go-lint-toolchain", 10000],
  ["govulncheck-toolchain", 10000],
  ["gosec-toolchain", 10000],
  ["shell-lint-toolchain", 10000],
  ["check-frontend-install", 30000],
  ["build-server", 15000],
  ["build-migrate", 15000],
  ["build-operator", 15000],
  ["testservices-build", 15000],
  ["test-service-images", 15000],
]);

function usage() {
  process.stderr.write(
    [
      "usage: check-scheduler-summary-timing-drift.mjs [--target <target>]",
      "  [--warm-check-budget-ms <ms>] [--warm-check-balance-ratio <ratio>]",
      "  <results-dir|run-dir|scheduler-events.jsonl>",
    ].join(" ") + "\n",
  );
  process.exit(2);
}

function parsePositiveInteger(value) {
  const parsed = Number.parseInt(value ?? "", 10);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null;
}

function parsePositiveNumber(value) {
  const parsed = Number.parseFloat(value ?? "");
  return Number.isFinite(parsed) && parsed >= 1 ? parsed : null;
}

function parseArgs(argv) {
  const options = {
    target: "",
    resultsDir: "",
    warmCheckBudgetMs: null,
    warmCheckBalanceRatio: null,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--target") {
      options.target = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--warm-check-budget-ms") {
      options.warmCheckBudgetMs = parsePositiveInteger(argv[index + 1]);
      index += 1;
      if (options.warmCheckBudgetMs === null) {
        usage();
      }
      continue;
    }
    if (arg === "--warm-check-balance-ratio") {
      options.warmCheckBalanceRatio = parsePositiveNumber(argv[index + 1]);
      index += 1;
      if (options.warmCheckBalanceRatio === null) {
        usage();
      }
      continue;
    }
    if (arg.startsWith("--")) {
      usage();
    }
    if (options.resultsDir) {
      usage();
    }
    options.resultsDir = arg;
  }
  if (!options.resultsDir) {
    usage();
  }
  return options;
}

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

function nonNegativeInteger(value) {
  return Number.isInteger(value) && value >= 0 ? value : null;
}

function summaryDurationMs(summary) {
  return (
    nonNegativeInteger(summary?.wall_duration_ms) ??
    nonNegativeInteger(summary?.critical_path_wall_duration_ms) ??
    nonNegativeInteger(summary?.duration_ms)
  );
}

function median(values) {
  if (values.length === 0) {
    return null;
  }
  const sorted = [...values].sort((left, right) => left - right);
  return sorted[Math.floor(sorted.length / 2)];
}

function goShardIsolationIDs() {
  const isolated = new Set();
  const plan = collectGoShardPlan(process.cwd());
  for (const shard of plan.shards ?? []) {
    if (shard?.isolated !== true && shard?.item_count !== 1) {
      continue;
    }
    isolated.add(`${shard.target}:${shard.name}`);
    isolated.add(`check-service-backed:${shard.target}:${shard.name}`);
  }
  return isolated;
}

function goShardFamilyID(target, workUnitID) {
  const name = String(workUnitID).split(":").pop() ?? workUnitID;
  const family = name.replace(/-shard-\d+$/, "");
  return `${target}:${family}`;
}

function pushBalanceError(errors, label, lanes, ratio) {
  if (lanes.length < 2) {
    return;
  }
  const durations = lanes.map((lane) => lane.durationMs);
  const peerMedian = median(durations);
  if (peerMedian === null || peerMedian === 0) {
    return;
  }
  const slowest = lanes
    .slice()
    .sort((left, right) => right.durationMs - left.durationMs || left.id.localeCompare(right.id))[0];
  const allowed = Math.ceil(peerMedian * ratio);
  if (slowest.durationMs > allowed && slowest.durationMs - allowed > warmBalanceMaterialSkewMs) {
    errors.push(
      `${label}: slowest non-isolated lane ${slowest.id} duration ${slowest.durationMs}ms exceeds ${ratio}x peer median ${peerMedian}ms`,
    );
  }
}

function schedulerAccountingExtension(record) {
  const extension = record?.extensions?.["cartulary.scheduler_accounting"];
  return extension && typeof extension === "object" ? extension : null;
}

function inputStampOutcome(record) {
  const extension = schedulerAccountingExtension(record);
  return extension?.input_stamp?.outcome ?? extension?.cache_outcome ?? "";
}

function validateWarmReadinessDurations(eventsFile, events, errors) {
  for (const event of events) {
    if (event.event !== "finish" || typeof event.work_unit_id !== "string") {
      continue;
    }
    const thresholdMs = warmReadinessThresholds.get(event.work_unit_id);
    if (!thresholdMs) {
      continue;
    }
    const durationMs = nonNegativeInteger(event.duration_ms);
    if (durationMs === null) {
      errors.push(`${eventsFile}: warm readiness unit ${event.work_unit_id} is missing duration_ms`);
      continue;
    }
    if (durationMs > thresholdMs) {
      errors.push(
        `${eventsFile}: warm readiness unit ${event.work_unit_id} duration ${durationMs}ms exceeds warm threshold ${thresholdMs}ms`,
      );
    }
  }
}

function validateWarmNoUnexpectedReuse(eventsFile, schedulerSummary, events, errors) {
  const reusedEventIDs = new Set();
  for (const event of events) {
    if (event.event !== "finish" || typeof event.work_unit_id !== "string") {
      continue;
    }
    const accounting = schedulerAccountingExtension(event);
    if (accounting?.accounting_mode !== "reused" && inputStampOutcome(event) !== "hit") {
      continue;
    }
    reusedEventIDs.add(event.work_unit_id);
    errors.push(
      `${eventsFile}: unexpected reused work ${event.work_unit_id} is not allowed in the current check profile`,
    );
  }

  const accounting = schedulerSummary?.extensions?.["cartulary.scheduler_accounting"];
  if (!accounting || typeof accounting !== "object") {
    return;
  }
  const summaryUnits = Array.isArray(accounting.work_unit_accounting)
    ? accounting.work_unit_accounting
    : [];
  for (const unit of summaryUnits) {
    if (unit.accounting_mode !== "reused") {
      continue;
    }
    reusedEventIDs.add(unit.id);
    errors.push(
      `${eventsFile}: scheduler accounting summary records unexpected reused work ${unit.id}`,
    );
  }
  const reusedCount = nonNegativeInteger(accounting.accounting_modes?.reused) ?? 0;
  if (reusedCount > 0) {
    errors.push(
      `${eventsFile}: scheduler accounting reused count ${reusedCount} is not allowed in the current check profile`,
    );
  }
}

function validateWarmFixtureBudget(runDir, errors) {
  const runSummaryFile = path.join(runDir, "run-summary.json");
  if (!existsSync(runSummaryFile)) {
    return;
  }
  const runSummary = readJSON(runSummaryFile);
  const strategies = Array.isArray(runSummary.fixture?.by_strategy)
    ? runSummary.fixture.by_strategy
    : [];
  let count = 0;
  let durationMs = 0;
  for (const strategy of strategies) {
    if (
      strategy?.service !== "postgres" ||
      strategy?.operation !== "database-reset" ||
      strategy?.fixture_policy !== "package_reset"
    ) {
      continue;
    }
    count += nonNegativeInteger(strategy.count) ?? 0;
    durationMs += nonNegativeInteger(strategy.total_duration_ms) ?? 0;
  }
  if (count > warmPackageResetCountBudget) {
    errors.push(
      `${runSummaryFile}: package-reset fixture count ${count} exceeds warm budget ${warmPackageResetCountBudget}`,
    );
  }
  if (durationMs > warmPackageResetDurationBudgetMs) {
    errors.push(
      `${runSummaryFile}: package-reset fixture duration ${durationMs}ms exceeds warm budget ${warmPackageResetDurationBudgetMs}ms`,
    );
  }
}

function validateWarmCheckStream(eventsFile, options) {
  const errors = [];
  const targetDir = path.dirname(eventsFile);
  const runDir = path.dirname(targetDir);
  const events = readEvents(eventsFile);
  const schedulerSummaryFile = path.join(targetDir, "scheduler-summary.json");
  const schedulerSummary = existsSync(schedulerSummaryFile) ? readJSON(schedulerSummaryFile) : null;
  const starts = new Map();
  for (const event of events) {
    if (event.event === "start" && typeof event.work_unit_id === "string") {
      starts.set(event.work_unit_id, event);
    }
  }
  validateWarmReadinessDurations(eventsFile, events, errors);
  validateWarmNoUnexpectedReuse(eventsFile, schedulerSummary, events, errors);
  validateWarmFixtureBudget(runDir, errors);

  const serviceSummaryFile = path.join(runDir, "check-service-backed", "target-summary.json");
  if (!existsSync(serviceSummaryFile)) {
    errors.push(`${serviceSummaryFile}: missing check-service-backed target summary for warm scheduler health`);
  } else if (options.warmCheckBudgetMs !== null) {
    const serviceSummary = readJSON(serviceSummaryFile);
    const durationMs = summaryDurationMs(serviceSummary);
    if (durationMs === null) {
      errors.push(`${serviceSummaryFile}: missing check-service-backed wall duration`);
    } else if (durationMs > options.warmCheckBudgetMs) {
      errors.push(
        `${serviceSummaryFile}: check-service-backed warm duration ${durationMs}ms exceeds budget ${options.warmCheckBudgetMs}ms`,
      );
    }
  }

  const forbiddenBrowserUnit = events.find((event) =>
    typeof event.work_unit_id === "string" &&
    event.work_unit_id.startsWith("check-service-backed:") &&
    (event.work_unit_id.includes("browser-e2e-measurement") ||
      event.work_unit_id.includes("browser-stage-session:measurement") ||
      event.work_unit_id.includes("browser-e2e-visual") ||
      event.work_unit_id.includes("browser-stage-session:visual") ||
      event.work_unit_id.includes("browser-e2e-a11y") ||
      event.work_unit_id.includes("browser-stage-session:a11y")),
  );
  if (forbiddenBrowserUnit) {
    errors.push(
      `${eventsFile}: default warm check includes explicit browser evidence unit ${forbiddenBrowserUnit.work_unit_id}`,
    );
  }

  if (options.warmCheckBalanceRatio === null) {
    return errors;
  }

  const isolatedGoShards = goShardIsolationIDs();
  const backendByFamily = new Map();
  const browserFunctional = [];
  for (const event of events) {
    if (event.event !== "finish" || typeof event.work_unit_id !== "string") {
      continue;
    }
    const durationMs = nonNegativeInteger(event.duration_ms);
    if (durationMs === null) {
      continue;
    }
    const start = starts.get(event.work_unit_id);
    if (!start) {
      continue;
    }
    const type = start.work_unit_type;
    const aggregateTarget = start.aggregate_target ?? "";
    if (type === "go_shard") {
      if (isolatedGoShards.has(event.work_unit_id)) {
        continue;
      }
      const target = aggregateTarget || String(start.work_unit ?? "").split("/")[0];
      if (!target) {
        continue;
      }
      const family = goShardFamilyID(target, event.work_unit_id);
      if (!backendByFamily.has(family)) {
        backendByFamily.set(family, []);
      }
      backendByFamily.get(family).push({ id: event.work_unit_id, durationMs });
      continue;
    }
    if (
      type === "browser_group" &&
      aggregateTarget === "browser-e2e-webserver-backed" &&
      event.work_unit_id.includes("browser-functional-shard")
    ) {
      browserFunctional.push({ id: event.work_unit_id, durationMs });
      continue;
    }
  }

  for (const [family, lanes] of backendByFamily) {
    pushBalanceError(errors, `${eventsFile}: ${family}`, lanes, options.warmCheckBalanceRatio);
  }
  pushBalanceError(
    errors,
    `${eventsFile}: browser-e2e-webserver-backed functional`,
    browserFunctional,
    options.warmCheckBalanceRatio,
  );
  return errors;
}

function validateWarmCheckHealth(eventFiles, options) {
  if (options.warmCheckBudgetMs === null && options.warmCheckBalanceRatio === null) {
    return { checked: 0, errors: [] };
  }
  const checkEventFiles = eventFiles.filter((file) => path.basename(path.dirname(file)) === "check");
  const errors = [];
  for (const eventsFile of checkEventFiles) {
    errors.push(...validateWarmCheckStream(eventsFile, options));
  }
  if (checkEventFiles.length === 0) {
    errors.push("no check scheduler-events.jsonl files found for warm scheduler health");
  }
  return { checked: checkEventFiles.length, errors };
}

const options = parseArgs(process.argv.slice(2));
const { schedulerEventFiles, errors } = validateSchedulerSummaryTiming(
  options.resultsDir,
  { target: options.target },
);
const warmCheck = validateWarmCheckHealth(schedulerEventFiles, options);
errors.push(...warmCheck.errors);
if (schedulerEventFiles.length === 0) {
  process.stderr.write("no scheduler-events.jsonl files found\n");
  process.exit(1);
}
if (errors.length > 0) {
  process.stderr.write("scheduler summary timing drift detected:\n");
  for (const error of errors) {
    process.stderr.write(`  ${error}\n`);
  }
  process.exit(1);
}
process.stdout.write(
  `scheduler summary timing verified for ${schedulerEventFiles.length} scheduler stream(s)\n`,
);
if (warmCheck.checked > 0) {
  process.stdout.write(`warm check scheduler health verified for ${warmCheck.checked} scheduler stream(s)\n`);
}
