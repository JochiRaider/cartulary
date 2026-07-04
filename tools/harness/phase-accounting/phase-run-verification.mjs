import { readFileSync } from "node:fs";

import {
  flattenPlaywrightSuites,
  summarizePlaywrightErrors,
} from "../output/test-output/playwright-artifacts.mjs";

export function goLogKey(pkg, test) {
  return `${pkg}::${test}`;
}

export function describeGoSymbol(entry, symbol) {
  return `${symbol} [${entry.package}]`;
}

export function readGoLogTopLevelStatuses(logFile) {
  const seen = new Map();
  for (const rawLine of readFileSync(logFile, "utf8").split(/\r?\n/)) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    const entry = JSON.parse(line);
    if (typeof entry.Package !== "string" || entry.Package === "") {
      continue;
    }
    if (
      !["run", "pass", "fail", "skip"].includes(entry.Action) ||
      typeof entry.Test !== "string" ||
      entry.Test.includes("/")
    ) {
      continue;
    }
    if (
      !entry.Test.startsWith("Test") &&
      !entry.Test.startsWith("Benchmark") &&
      !entry.Test.startsWith("Fuzz")
    ) {
      continue;
    }
    const key = goLogKey(entry.Package, entry.Test);
    if (!seen.has(key)) {
      seen.set(key, { package: entry.Package, test: entry.Test, status: "" });
    }
    const current = seen.get(key);
    if (entry.Action === "run") {
      if (current.status === "") {
        current.status = "run";
      }
      continue;
    }
    current.status = entry.Action;
  }
  return seen;
}

function readVitestReport(reportFile) {
  return JSON.parse(readFileSync(reportFile, "utf8"));
}

export function verifyVitestRun(reportFile, expectedTitles) {
  const report = readVitestReport(reportFile);
  if (report.success !== true) {
    throw new Error("vitest manifest run failed");
  }

  const executed = [];
  const failed = [];
  const files = new Set();
  for (const fileResult of report.testResults ?? []) {
    if (typeof fileResult?.name === "string" && fileResult.name !== "") {
      files.add(fileResult.name);
    }
    for (const assertion of fileResult.assertionResults ?? []) {
      if (assertion.status === "skipped") {
        continue;
      }
      if (typeof assertion.title !== "string" || assertion.title === "") {
        failed.push("(missing title)");
        continue;
      }
      if (assertion.status !== "passed") {
        failed.push(`${assertion.title} (${assertion.status})`);
        continue;
      }
      executed.push(assertion.title);
    }
  }

  const missing = expectedTitles.filter((title) => !executed.includes(title));
  const unexpected = executed.filter((title) => !expectedTitles.includes(title));
  if (missing.length > 0 || unexpected.length > 0 || failed.length > 0) {
    throw new Error(
      `vitest execution mismatch: missing=${missing.join(",") || "none"} unexpected=${unexpected.join(",") || "none"} failed=${failed.join(",") || "none"}`,
    );
  }

  return {
    files: Array.from(files).sort(),
    executed: executed.sort(),
  };
}

function readPlaywrightReport(reportFile) {
  return JSON.parse(readFileSync(reportFile, "utf8"));
}

function detectPlaywrightSetupFailure(report) {
  const specs = flattenPlaywrightSuites(report.suites);
  if (specs.length > 0 || (report.errors ?? []).length === 0) {
    return null;
  }
  const summary = summarizePlaywrightErrors(report);
  return summary === "" ? "playwright setup failure" : `playwright setup failure: ${summary}`;
}

export function verifyPlaywrightSpecSet(reportFile, expectedTitles) {
  const report = readPlaywrightReport(reportFile);
  const setupFailure = detectPlaywrightSetupFailure(report);
  if (setupFailure !== null) {
    throw new Error(setupFailure);
  }
  const actualTitles = flattenPlaywrightSuites(report.suites).map((spec) => spec.title);
  const missing = expectedTitles.filter((title) => !actualTitles.includes(title));
  const unexpected = actualTitles.filter((title) => !expectedTitles.includes(title));
  if (missing.length > 0 || unexpected.length > 0) {
    throw new Error(
      `playwright manifest mismatch: missing=${missing.join(",") || "none"} unexpected=${unexpected.join(",") || "none"}`,
    );
  }
  return report;
}

export function playwrightReportSpecs(report) {
  return flattenPlaywrightSuites(report.suites);
}

export function extractPlaywrightStatuses(spec) {
  const statuses = [];
  for (const test of spec.tests ?? []) {
    if (Array.isArray(test.results) && test.results.length > 0) {
      for (const result of test.results) {
        if (typeof result.status === "string" && result.status !== "") {
          statuses.push(result.status);
        }
      }
      continue;
    }
    if (typeof test.status === "string" && test.status !== "" && test.status !== "skipped") {
      statuses.push(test.status);
    }
  }
  return statuses;
}
