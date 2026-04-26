#!/usr/bin/env node
import { existsSync, readFileSync, readdirSync, writeFileSync } from "node:fs";
import path from "node:path";

const baselineFile = path.join("tools", "go_test_duration_baselines.json");
const defaultShardTargetMsByTarget = {
  "backend-integration": 18000,
  "backend-integration-support": 18000,
  "backend-store": 30000,
};

function usage() {
  process.stderr.write("usage: update-go-test-durations.mjs [--prune-observed-packages] <results-dir>\n");
  process.exit(2);
}

function walkFiles(root, predicate, files = []) {
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    const fullPath = path.join(root, entry.name);
    if (entry.isDirectory()) {
      walkFiles(fullPath, predicate, files);
      continue;
    }
    if (entry.isFile() && predicate(fullPath)) {
      files.push(fullPath);
    }
  }
  return files;
}

function readBaseline() {
  if (!existsSync(baselineFile)) {
    return {
      schema_id: "cartulary.go_test_duration_baselines.v2",
      default_shard_target_ms: 30000,
      shard_target_ms_by_target: defaultShardTargetMsByTarget,
      default_integration_weight_ms: 10000,
      tests: {},
    };
  }
  return JSON.parse(readFileSync(baselineFile, "utf8"));
}

function collectDurations(resultsDir) {
  const durations = new Map();
  const observedPackages = new Set();
  const logs = walkFiles(resultsDir, (file) => path.basename(file) === "runner.jsonl");
  for (const log of logs) {
    const statusFile = path.join(path.dirname(log), "exit_status.txt");
    if (existsSync(statusFile) && readFileSync(statusFile, "utf8").trim() !== "0") {
      continue;
    }
    for (const rawLine of readFileSync(log, "utf8").split(/\r?\n/)) {
      const line = rawLine.trim();
      if (line === "") {
        continue;
      }
      let event;
      try {
        event = JSON.parse(line);
      } catch {
        continue;
      }
      if (
        event.Action !== "pass" ||
        typeof event.Package !== "string" ||
        typeof event.Test !== "string" ||
        event.Test.includes("/") ||
        !event.Test.startsWith("Test")
      ) {
        continue;
      }
      observedPackages.add(event.Package);
      const elapsedMs = Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000));
      const key = `${event.Package}::${event.Test}`;
      durations.set(key, Math.max(durations.get(key) ?? 0, elapsedMs));
    }
  }
  return { durations, observedPackages };
}

function sortedObject(value) {
  return Object.fromEntries(Object.entries(value).sort(([left], [right]) => left.localeCompare(right)));
}

function parseArgs(argv) {
  const options = {
    pruneObservedPackages: false,
    resultsDir: "",
  };
  for (const arg of argv) {
    if (arg === "--prune-observed-packages") {
      options.pruneObservedPackages = true;
      continue;
    }
    if (!options.resultsDir) {
      options.resultsDir = arg;
      continue;
    }
    usage();
  }
  return options;
}

function main(argv) {
  const options = parseArgs(argv);
  const resultsDir = options.resultsDir;
  if (!resultsDir) {
    usage();
  }
  const absoluteResultsDir = path.resolve(resultsDir);
  if (!existsSync(absoluteResultsDir)) {
    throw new Error(`results directory does not exist: ${resultsDir}`);
  }

  const baseline = readBaseline();
  const { durations, observedPackages } = collectDurations(absoluteResultsDir);
  baseline.schema_id = "cartulary.go_test_duration_baselines.v2";
  baseline.default_shard_target_ms ??= 30000;
  baseline.shard_target_ms_by_target = {
    ...defaultShardTargetMsByTarget,
    ...(baseline.shard_target_ms_by_target ?? {}),
  };
  baseline.default_integration_weight_ms ??= 10000;
  baseline.tests ??= {};
  if (options.pruneObservedPackages) {
    for (const key of Object.keys(baseline.tests)) {
      const [packageName] = key.split("::", 1);
      if (observedPackages.has(packageName) && !durations.has(key)) {
        delete baseline.tests[key];
      }
    }
  }
  for (const [key, durationMs] of durations) {
    baseline.tests[key] = durationMs;
  }
  baseline.updated_at = new Date().toISOString();
  baseline.shard_target_ms_by_target = sortedObject(baseline.shard_target_ms_by_target);
  baseline.tests = sortedObject(baseline.tests);

  writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  process.stdout.write(`updated ${durations.size} Go test duration baselines\n`);
}

try {
  main(process.argv.slice(2));
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
