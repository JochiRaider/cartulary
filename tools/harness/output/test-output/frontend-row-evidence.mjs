import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { flattenPlaywrightSuites } from "./playwright-artifacts.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..", "..");

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function resolveArtifactPath(value) {
  if (!value) {
    return "";
  }
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function collectVitestFileResults(report) {
  if (Array.isArray(report.testResults)) {
    return report.testResults;
  }
  if (Array.isArray(report.testResults?.testResults)) {
    return report.testResults.testResults;
  }
  return [];
}

function normalizeVitestOwnerPath(name) {
  const normalized = normalizePath(name);
  if (normalized.startsWith(repoRoot)) {
    return normalizePath(path.relative(repoRoot, normalized));
  }
  return normalized;
}

export function collectVitestTitleObservations(reportFile) {
  const byTitle = new Map();
  if (!existsSync(reportFile)) {
    return byTitle;
  }
  const report = JSON.parse(readFileSync(reportFile, "utf8"));
  for (const fileResult of collectVitestFileResults(report)) {
    const ownerPath = normalizeVitestOwnerPath(fileResult.name ?? "");
    for (const assertion of fileResult.assertionResults ?? []) {
      const title = assertion.title ?? "";
      if (title === "") {
        continue;
      }
      const observations = byTitle.get(title) ?? [];
      observations.push({
        file: ownerPath,
        status: assertion.status ?? "unknown",
      });
      byTitle.set(title, observations);
    }
  }
  return byTitle;
}

function collectPhaseSummaryRunnerFiles(targetDir) {
  const runnerFiles = new Set();
  if (!existsSync(targetDir)) {
    return [];
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
      if (!entry.isFile() || entry.name !== "phase-summary.json") {
        continue;
      }
      let summary;
      try {
        summary = JSON.parse(readFileSync(next, "utf8"));
      } catch {
        continue;
      }
      const runnerJSON = summary.artifacts?.runner_json;
      if (typeof runnerJSON === "string" && runnerJSON.trim() !== "") {
        runnerFiles.add(resolveArtifactPath(runnerJSON));
      }
    }
  }
  return [...runnerFiles].sort();
}

function normalizePlaywrightFile(file) {
  const normalized = normalizePath(file);
  if (normalized.startsWith("apps/web/")) {
    return normalized;
  }
  return normalizePath(path.join("apps/web", "e2e", normalized));
}

function phaseSummaryTitleObservationFile(value) {
  const file = typeof value === "string" ? value : "";
  if (file === "") {
    return "";
  }
  return file.startsWith("apps/web/")
    ? normalizePath(file)
    : normalizePlaywrightFile(file);
}

function collectPhaseSummaryTitleObservations(targetDir) {
  const byTitle = new Map();
  if (!existsSync(targetDir)) {
    return byTitle;
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
      if (!entry.isFile() || entry.name !== "phase-summary.json") {
        continue;
      }
      let summary;
      try {
        summary = JSON.parse(readFileSync(next, "utf8"));
      } catch {
        continue;
      }
      for (const item of summary.inventory ?? []) {
        const title = item.symbol_or_title ?? "";
        if (title === "") {
          continue;
        }
        const observations = byTitle.get(title) ?? [];
        observations.push({
          file: phaseSummaryTitleObservationFile(item.package_or_file),
          status: "passed",
        });
        byTitle.set(title, observations);
      }
      for (const failure of summary.failures ?? []) {
        const title = failure.symbol_or_title ?? "";
        if (title === "") {
          continue;
        }
        const observations = byTitle.get(title) ?? [];
        observations.push({
          file: phaseSummaryTitleObservationFile(failure.package_or_file),
          status: "failed",
        });
        byTitle.set(title, observations);
      }
    }
  }
  return byTitle;
}

function normalizePlaywrightScenarioStatus(status) {
  switch (status) {
    case "passed":
    case "flaky":
      return "passed";
    case "skipped":
      return "skipped";
    case "failed":
    case "timedOut":
    case "interrupted":
      return "failed";
    default:
      return "unknown";
  }
}

function playwrightSpecScenarioStatus(spec) {
  const statuses = [];
  for (const test of spec.tests ?? []) {
    const results = test.results ?? [];
    if (results.length === 0) {
      const testStatus = normalizePlaywrightScenarioStatus(test.status ?? "");
      if (testStatus !== "unknown") {
        statuses.push(testStatus);
      }
      continue;
    }
    for (const result of results) {
      statuses.push(normalizePlaywrightScenarioStatus(result.status ?? ""));
    }
  }
  if (statuses.length === 0) {
    return "unknown";
  }
  if (statuses.some((status) => status === "failed")) {
    return "failed";
  }
  if (statuses.some((status) => status === "passed")) {
    return "passed";
  }
  if (statuses.every((status) => status === "skipped")) {
    return "skipped";
  }
  return "unknown";
}

function collectPlaywrightTitleObservations(reportFile) {
  const byTitle = new Map();
  if (!existsSync(reportFile)) {
    return byTitle;
  }
  const parsed = JSON.parse(readFileSync(reportFile, "utf8"));
  const reports = Array.isArray(parsed) ? parsed : [parsed];
  for (const report of reports) {
    for (const spec of flattenPlaywrightSuites(report.suites ?? [])) {
      const title = spec.title ?? "";
      if (title === "") {
        continue;
      }
      const observations = byTitle.get(title) ?? [];
      observations.push({
        file: normalizePlaywrightFile(spec.file ?? ""),
        status: playwrightSpecScenarioStatus(spec),
      });
      byTitle.set(title, observations);
    }
  }
  return byTitle;
}

function mergeTitleObservations(target, source) {
  for (const [title, observations] of source.entries()) {
    target.set(title, [...(target.get(title) ?? []), ...observations]);
  }
}

export function collectPlaywrightTitleObservationsForTarget(targetDir) {
  const byTitle = collectPhaseSummaryTitleObservations(targetDir);
  for (const runnerFile of collectPhaseSummaryRunnerFiles(targetDir)) {
    mergeTitleObservations(
      byTitle,
      collectPlaywrightTitleObservations(runnerFile),
    );
  }
  return byTitle;
}

export function frontendScenarioStatus(observations) {
  if (!observations || observations.length === 0) {
    return "missing";
  }
  if (observations.some((entry) => entry.status === "failed")) {
    return "failed";
  }
  if (observations.some((entry) => entry.status === "passed")) {
    return "passed";
  }
  if (observations.every((entry) => entry.status === "skipped")) {
    return "skipped";
  }
  return "unknown";
}
