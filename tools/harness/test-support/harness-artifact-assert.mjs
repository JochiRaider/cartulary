#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
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

function collectStepSummaries(targetDir) {
  const summaries = [];
  if (!existsSync(targetDir)) {
    return summaries;
  }
  const stack = [targetDir];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (entry.isFile() && entry.name === "step-summary.json") {
        summaries.push(next);
      }
    }
  }
  summaries.sort();
  return summaries;
}

function readSummary(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch (error) {
    throw new Error(
      `failed to read step summary ${file}: ${error instanceof Error ? error.message : String(error)}`,
    );
  }
}

function logPathsForSummary(repoRoot, summary) {
  return ["stderr_log", "stdout_log"]
    .map((key) => resolvePath(repoRoot, summary.artifacts?.[key] ?? ""))
    .filter((file) => file && existsSync(file) && statSync(file).isFile());
}

function assertArtifactContains(options) {
  const repoRoot = path.resolve(options.repoRoot);
  const resultsRoot = resolvePath(repoRoot, options.resultsRoot);
  const targetDir = path.join(resultsRoot, options.runId, options.target);
  const summaryFiles = collectStepSummaries(targetDir);
  if (summaryFiles.length === 0) {
    throw new Error(
      `${options.label}: no step summaries found for target ${options.target} under ${targetDir}`,
    );
  }

  const summaries = summaryFiles.map((file) => ({
    file,
    summary: readSummary(file),
  }));
  const matchingSummaries = options.stepLabel
    ? summaries.filter(({ summary }) => summary.label === options.stepLabel)
    : summaries;
  if (matchingSummaries.length === 0) {
    const available = summaries
      .map(({ summary }) => summary.label)
      .filter(Boolean)
      .join(", ");
    throw new Error(
      `${options.label}: no step summary matched step label ${JSON.stringify(options.stepLabel)}; available labels: ${available || "(none)"}`,
    );
  }

  const searchedLogs = [];
  for (const { summary } of matchingSummaries) {
    for (const logPath of logPathsForSummary(repoRoot, summary)) {
      searchedLogs.push(logPath);
      if (readFileSync(logPath, "utf8").includes(options.needle)) {
        return;
      }
    }
  }

  throw new Error(
    `${options.label}: expected target ${options.target} artifact logs to contain [${options.needle}]; searched ${searchedLogs.length > 0 ? searchedLogs.join(", ") : "(no declared logs)"}`,
  );
}

try {
  assertArtifactContains(parseArgs(process.argv.slice(2)));
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exit(1);
}
