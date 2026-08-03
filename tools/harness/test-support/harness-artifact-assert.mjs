#!/usr/bin/env node

import { existsSync, readFileSync, statSync } from "node:fs";
import path from "node:path";

function usage() {
  return [
    "usage: harness-artifact-assert.mjs --results-root <dir> --run-id <id> --target <target> --needle <text> --label <label> [--step-label <label>] [--repo-root <dir>]",
  ].join("\n");
}

function parseArgs(argv) {
  const options = {
    stepLabel: "",
    repoRoot: process.cwd(),
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    const value = argv[index + 1] ?? "";
    switch (arg) {
      case "--results-root":
        options.resultsRoot = value;
        index += 1;
        break;
      case "--run-id":
        options.runId = value;
        index += 1;
        break;
      case "--target":
        options.target = value;
        index += 1;
        break;
      case "--step-label":
        options.stepLabel = value;
        index += 1;
        break;
      case "--needle":
        options.needle = value;
        index += 1;
        break;
      case "--label":
        options.label = value;
        index += 1;
        break;
      case "--repo-root":
        options.repoRoot = value;
        index += 1;
        break;
      default:
        throw new Error(`unknown option ${arg}\n${usage()}`);
    }
  }
  for (const key of ["resultsRoot", "runId", "target", "needle", "label"]) {
    if (!options[key]) {
      throw new Error(`missing required --${key.replace(/[A-Z]/g, (char) => `-${char.toLowerCase()}`)}\n${usage()}`);
    }
  }
  return options;
}

function resolvePath(repoRoot, file) {
  if (!file) {
    return "";
  }
  return path.isAbsolute(file) ? file : path.join(repoRoot, file);
}

function readTargetSummary(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch (error) {
    throw new Error(
      `failed to read canonical target summary ${file}: ${error instanceof Error ? error.message : String(error)}`,
    );
  }
}

function safeUnitID(unitID) {
  return unitID.replaceAll(/[^A-Za-z0-9_.-]+/gu, "-");
}

function logPathsForUnit(runRoot, unitID) {
  const logRoot = path.join(runRoot, "unit-logs", safeUnitID(unitID));
  return ["stderr.log", "stdout.log"]
    .map((file) => path.join(logRoot, file))
    .filter((file) => existsSync(file) && statSync(file).isFile());
}

function assertArtifactContains(options) {
  const repoRoot = path.resolve(options.repoRoot);
  const resultsRoot = resolvePath(repoRoot, options.resultsRoot);
  const runRoot = path.join(resultsRoot, options.runId);
  const summaryFile = path.join(runRoot, "target-summaries", `${options.target}.json`);
  if (!existsSync(summaryFile)) {
    throw new Error(
      `${options.label}: canonical target summary is missing for ${options.target} at ${summaryFile}`,
    );
  }

  const summary = readTargetSummary(summaryFile);
  const unitIDs = Array.isArray(summary.unit_ids) ? summary.unit_ids : [];
  const matchingUnitIDs = options.stepLabel
    ? unitIDs.filter((unitID) => unitID === options.stepLabel || unitID.endsWith(`:${options.stepLabel}`))
    : unitIDs;
  if (matchingUnitIDs.length === 0) {
    throw new Error(
      `${options.label}: no canonical unit matched step label ${JSON.stringify(options.stepLabel)}; available unit IDs: ${unitIDs.join(", ") || "(none)"}`,
    );
  }

  const searchedLogs = [];
  for (const unitID of matchingUnitIDs) {
    for (const logPath of logPathsForUnit(runRoot, unitID)) {
      searchedLogs.push(logPath);
      if (readFileSync(logPath, "utf8").includes(options.needle)) {
        return;
      }
    }
  }

  throw new Error(
    `${options.label}: expected target ${options.target} canonical unit logs to contain [${options.needle}]; searched ${searchedLogs.length > 0 ? searchedLogs.join(", ") : "(no unit logs)"}`,
  );
}

try {
  assertArtifactContains(parseArgs(process.argv.slice(2)));
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exit(1);
}
