#!/usr/bin/env node
import { existsSync, readFileSync, readdirSync, writeFileSync } from "node:fs";
import path from "node:path";

const baselineFile = path.join("tools", "go_test_duration_baselines.json");

function usage() {
  process.stderr.write("usage: update-go-test-durations.mjs <results-dir>\n");
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
      schema_id: "cartulary.go_test_duration_baselines.v1",
      default_shard_target_ms: 30000,
      default_integration_weight_ms: 10000,
      tests: {},
    };
  }
  return JSON.parse(readFileSync(baselineFile, "utf8"));
}

function collectDurations(resultsDir) {
  const durations = new Map();
  const logs = walkFiles(resultsDir, (file) => path.basename(file) === "runner.jsonl");
  for (const log of logs) {
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
      const elapsedMs = Math.max(1, Math.round(Number(event.Elapsed ?? 0) * 1000));
      const key = `${event.Package}::${event.Test}`;
      durations.set(key, Math.max(durations.get(key) ?? 0, elapsedMs));
    }
  }
  return durations;
}

function main(argv) {
  const [resultsDir] = argv;
  if (!resultsDir) {
    usage();
  }
  const absoluteResultsDir = path.resolve(resultsDir);
  if (!existsSync(absoluteResultsDir)) {
    throw new Error(`results directory does not exist: ${resultsDir}`);
  }

  const baseline = readBaseline();
  const collected = collectDurations(absoluteResultsDir);
  baseline.schema_id = "cartulary.go_test_duration_baselines.v1";
  baseline.default_shard_target_ms ??= 30000;
  baseline.default_integration_weight_ms ??= 10000;
  baseline.tests ??= {};
  for (const [key, durationMs] of collected) {
    baseline.tests[key] = Math.max(Number(baseline.tests[key] ?? 0), durationMs);
  }
  baseline.updated_at = new Date().toISOString();
  baseline.tests = Object.fromEntries(Object.entries(baseline.tests).sort());

  writeFileSync(baselineFile, `${JSON.stringify(baseline, null, 2)}\n`);
  process.stdout.write(`updated ${collected.size} Go test duration baselines\n`);
}

try {
  main(process.argv.slice(2));
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
