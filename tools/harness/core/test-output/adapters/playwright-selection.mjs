import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

const selectionReportCache = new Map();

function normalizePath(value) {
  return String(value ?? "").replaceAll("\\", "/");
}

function relToRoot(root, value) {
  if (!value) {
    return "";
  }
  const normalized = normalizePath(value);
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(root, value));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

function normalizePlaywrightFile(file) {
  const normalized = normalizePath(file);
  if (normalized.startsWith("apps/web/")) {
    return normalized;
  }
  return normalizePath(path.join("apps/web", "e2e", normalized));
}

function normalizePlaywrightSelectionReportFile(file) {
  const normalized = normalizePath(file);
  if (normalized.startsWith("apps/web/")) {
    return normalized;
  }
  if (normalized.startsWith("e2e/")) {
    return `apps/web/${normalized}`;
  }
  return normalizePlaywrightFile(normalized);
}

function frontendPhaseFromRowID(rowID) {
  const match = /^FE-(?:U|I|B|E|V|A11Y|S)-P([0-9]+)-[0-9]{2}$/u.exec(
    rowID,
  );
  return match ? `FE-P${match[1]}` : "";
}

export function readPlaywrightSelectionReport(root, reportFile, scope = null) {
  if (!reportFile || !existsSync(reportFile)) {
    return null;
  }
  const cacheKey = `${root}\u0000${reportFile}\u0000${JSON.stringify(
    scope ?? {},
  )}`;
  if (selectionReportCache.has(cacheKey)) {
    return selectionReportCache.get(cacheKey);
  }
  const report = JSON.parse(readFileSync(reportFile, "utf8"));
  if (report.schema_id !== "cartulary.playwright_manifest_selection.v1") {
    selectionReportCache.set(cacheKey, null);
    return null;
  }
  if (scope) {
    if (report.phase !== scope.phase) {
      throw new Error(
        `${relToRoot(root, reportFile)} phase ${report.phase} does not match ${scope.phase}`,
      );
    }
    if (report.coverage !== scope.coverage) {
      throw new Error(
        `${relToRoot(root, reportFile)} coverage ${report.coverage} does not match ${scope.coverage}`,
      );
    }
    const reportExecutionDependency = report.execution_dependency ?? "";
    if (reportExecutionDependency !== scope.executionDependency) {
      throw new Error(
        `${relToRoot(root, reportFile)} execution_dependency ${reportExecutionDependency} does not match ${scope.executionDependency}`,
      );
    }
  }
  const tests = [];
  for (const [index, test] of (report.selected_tests ?? []).entries()) {
    if (
      typeof test?.id !== "string" ||
      test.id.trim() === "" ||
      typeof test?.file !== "string" ||
      test.file.trim() === "" ||
      typeof test?.title !== "string" ||
      test.title.trim() === ""
    ) {
      throw new Error(
        `${relToRoot(root, reportFile)} selected_tests[${index + 1}] must declare id, file, and title`,
      );
    }
    tests.push({
      id: test.id,
      file: normalizePlaywrightSelectionReportFile(test.file),
      title: test.title,
      coverage: test.coverage ?? scope?.coverage ?? "authoritative",
      execution_dependency:
        test.execution_dependency ?? scope?.executionDependency ?? "",
      phase: frontendPhaseFromRowID(test.id) || scope?.phase || report.phase,
    });
  }
  const selection = { report, tests };
  selectionReportCache.set(cacheKey, selection);
  return selection;
}

export function selectedPlaywrightEntriesFromReport(
  root,
  reportFile,
  scope,
) {
  const selection = readPlaywrightSelectionReport(root, reportFile, scope);
  if (!selection) {
    return null;
  }
  return selection.tests.map((test) => ({
    id: test.id,
    phase: scope.phase,
    runner: "playwright",
    file: test.file,
    title: test.title,
    coverage: test.coverage,
    execution_dependency: test.execution_dependency,
  }));
}
