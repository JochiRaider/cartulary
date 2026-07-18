#!/usr/bin/env node
import { repoRoot } from "../../../contract/index.mjs";

import {
  existsSync,
  readFileSync,
  rmSync,
  statSync,
} from "node:fs";
import path from "node:path";
import { loadFrontendPlaywrightIndex as loadFrontendPlaywrightIndexAdapter } from "../frontend-indexes.mjs";
import {
  loadManifestIndex as loadManifestIndexAdapter,
  playwrightEntryTitles,
  selectManifestEntries,
  selectPlaywrightManifestEntries as selectPlaywrightManifestEntriesAdapter,
  selectPlaywrightEntries,
  vitestEntryTitles,
} from "../catalog-manifest-adapter.mjs";
import {
  flattenPlaywrightSuites,
  readPlaywrightSelectionReport as readPlaywrightSelectionReportAdapter,
  summarizePlaywrightErrors,
  selectedPlaywrightEntriesFromReport as selectedPlaywrightEntriesFromReportAdapter,
} from "../playwright-artifacts.mjs";
import { verboseOutput } from "../../tool-output.mjs";
import {
  testCoverageBucketSet,
  testCoverageBuckets,
} from "../../../contract/test-output-context.mjs";
import {
  createBaseStepContext,
  writeStepArtifacts,
} from "../step-artifacts.mjs";

let cachedGoModulePath;


function resolveGoModulePath() {
  if (cachedGoModulePath !== undefined) {
    return cachedGoModulePath;
  }
  const goMod = readFileSync(path.join(repoRoot, "go.mod"), "utf8");
  const match = goMod.match(/^module\s+(\S+)$/m);
  if (!match) {
    throw new Error("unable to determine Go module path from go.mod");
  }
  cachedGoModulePath = match[1];
  return cachedGoModulePath;
}

function toGoImportPath(repoRelativePackage) {
  if (!repoRelativePackage.startsWith("./")) {
    return repoRelativePackage;
  }
  const suffix = repoRelativePackage.slice(2);
  if (suffix === "") {
    return resolveGoModulePath();
  }
  return `${resolveGoModulePath()}/${suffix}`;
}

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function relToRepo(value) {
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

function createCounts() {
  const counts = {
    tests: 0,
    failed: 0,
    non_test: 0,
    non_test_failed: 0,
    packages: 0,
  };
  for (const coverage of testCoverageBuckets) {
    counts[coverage] = 0;
    counts[`${coverage}_failed`] = 0;
  }
  return counts;
}

function normalizeTestCoverage(value, fallback = "unmapped") {
  const normalized = String(value ?? "").trim();
  if (testCoverageBucketSet.has(normalized)) {
    return normalized;
  }
  return testCoverageBucketSet.has(fallback) ? fallback : "unmapped";
}

function addCoverageCount(counts, coverage, amount = 1) {
  counts[normalizeTestCoverage(coverage)] += amount;
}

function addCoverageFailureCount(counts, coverage, amount = 1) {
  counts[`${normalizeTestCoverage(coverage)}_failed`] += amount;
}

function clampDurationMs(value) {
  if (!Number.isFinite(value) || value < 0) {
    return 0;
  }
  return value;
}

function requiredEnv(name) {
  const value = process.env[name];
  if (value === undefined || value === "") {
    throw new Error(`missing required environment variable ${name}`);
  }
  return value;
}

function optionalEnv(name, fallback = "") {
  return process.env[name] ?? fallback;
}

function optionalLines(name) {
  const value = optionalEnv(name);
  if (value === "") {
    return [];
  }
  return value
    .split("\n")
    .map((entry) => entry.trim())
    .filter(Boolean);
}

function optionalSetFromLines(name) {
  return new Set(optionalLines(name));
}

function removeEmptyArtifact(file) {
  if (!file) {
    return;
  }
  let stat;
  try {
    stat = statSync(file);
  } catch (error) {
    if (error?.code === "ENOENT") {
      return;
    }
    throw error;
  }
  if (stat.size === 0) {
    rmSync(file, { force: true });
  }
}

function loadManifestIndex() {
  return loadManifestIndexAdapter(repoRoot, { normalizePath, toGoImportPath });
}

function catalogFallbackClassification() {
  return null;
}

function inferStepFromText(value) {
  if (!value) {
    return "";
  }
  const patterns = [
    /\bstep(?:\s|_|-)?(\d+)\b/i,
    /\b[UIE][-_](\d+)-\d+\b/,
    /\b[UIE]_(\d+)_\d+\b/,
  ];
  for (const pattern of patterns) {
    const match = value.match(pattern);
    if (match) {
      return `step${match[1]}`;
    }
  }
  return "";
}

function claimsConformanceRowTitle(value) {
  const text = String(value ?? "");
  return (
    /\bFE-[A-Z]+-P\d+-\d+\b/.test(text) ||
    /\b[UIE]-\d+(?:-[A-Z0-9]+)*-\d+\b/.test(text)
  );
}

function renderList(values) {
  if (values.length === 0) {
    return "none";
  }
  return values.join(",");
}

function printBlock(header, fields) {
  const lines = [header];
  for (const [key, value] of Object.entries(fields)) {
    lines.push(`${key}=${value === "" ? "-" : value}`);
  }
  process.stderr.write(`${lines.join("\n")}\n`);
}

function showStepDetailOutput(context) {
  return verboseOutput() || context.target === "adhoc";
}

function createInventoryItem({ coverage, step, id, owner, name }) {
  return {
    coverage,
    step,
    id: id ?? "",
    package_or_file: owner,
    symbol_or_title: name,
  };
}

function readManifestScopeEnv() {
  return {
    step: requiredEnv("CARTULARY_CATALOG_OWNER_ID"),
    coverage: requiredEnv("CARTULARY_MANIFEST_COVERAGE"),
    executionDependency: optionalEnv("CARTULARY_MANIFEST_EXECUTION_DEPENDENCY"),
  };
}

function manifestCoverageToInventoryCoverage(coverage) {
  return coverage === "authoritative" ? "authoritative" : "support";
}

function evaluateFlatTitleManifest(
  summary,
  { step, entries, inventoryCoverage },
) {
  const executedKeys = new Set(
    summary.inventory
      .filter((item) => item.coverage === inventoryCoverage)
      .map((item) => `${item.package_or_file}::${item.symbol_or_title}`),
  );
  const missingIDs = [
    ...new Set(
      entries
        .flatMap((entry) =>
          (entry.runner === "vitest"
            ? vitestEntryTitles(entry)
            : entry.runner === "playwright"
              ? playwrightEntryTitles(entry)
              : [entry.title]
          ).map((title) => ({
            entry,
            title,
          })),
        )
        .filter(
          ({ entry, title }) => !executedKeys.has(`${entry.file}::${title}`),
        )
        .map(({ entry }) => entry.id),
    ),
  ].sort();
  const expectedIDs = new Set(entries.map((entry) => entry.id));
  const unexpectedIDs = summary.inventory
    .filter(
      (item) =>
        item.coverage === inventoryCoverage &&
        item.id &&
        !expectedIDs.has(item.id),
    )
    .map((item) => item.id)
    .sort();

  return {
    step,
    missingIDs,
    unexpectedIDs,
    forbiddenIDFiles: Array.from(
      loadManifestIndex().forbiddenFilesByStep.get(step) ?? [],
    ).sort(),
  };
}

function finalizeManifestAwareRunnerStep(
  context,
  {
    catalogAware,
    runner,
    section = "",
    summary,
    selectedSlicePassed,
    artifacts,
    manifestMismatchArtifacts = () => ({}),
    manifestMismatchDetailFields = () => ({}),
    failureDetailFields = (dossier) => dossier,
    extraWriteStepDetails = {},
  },
) {
  let status = selectedSlicePassed ? "pass" : "fail";
  let manifestSummary = null;
  let manifestMismatch = null;
  let stepCounts = summary.counts;
  let stepDossiers = summary.dossiers;
  const emptySelectionAllowed =
    runner === "vitest" &&
    optionalEnv("CARTULARY_VITEST_ALLOW_EMPTY_SELECTION") === "1" &&
    (summary.counts?.tests ?? 0) === 0 &&
    summary.dossiers.length === 0;

  if (catalogAware && selectedSlicePassed && !emptySelectionAllowed) {
    const scope = readManifestScopeEnv();
    const selectedIDs =
      runner === "playwright"
        ? new Set()
        : optionalSetFromLines("CARTULARY_MANIFEST_SELECTED_IDS");
    const selectedPlaywrightEntries =
      runner === "playwright"
        ? selectedPlaywrightEntriesFromReport(
            optionalEnv("CARTULARY_PLAYWRIGHT_SELECTION_REPORT"),
            scope,
          )
        : null;
    const entries =
      selectedPlaywrightEntries ??
      (runner === "playwright"
        ? selectPlaywrightEntries(
            scope.step,
            scope.coverage,
            scope.executionDependency,
          )
        : selectManifestEntries(repoRoot, {
            runner,
            section,
            ...scope,
          })
      ).filter((entry) => selectedIDs.size === 0 || selectedIDs.has(entry.id));
    const verification = evaluateFlatTitleManifest(summary, {
      step: scope.step,
      entries,
      inventoryCoverage: manifestCoverageToInventoryCoverage(scope.coverage),
    });
    manifestSummary = {
      missing_ids: verification.missingIDs,
      unexpected_ids: verification.unexpectedIDs,
    };
    if (
      verification.missingIDs.length > 0 ||
      verification.unexpectedIDs.length > 0
    ) {
      status = "fail";
      manifestMismatch = {
        missing_ids: verification.missingIDs,
        unexpected_ids: verification.unexpectedIDs,
        forbidden_id_files: verification.forbiddenIDFiles,
        ...manifestMismatchArtifacts(verification, scope),
      };
    }
  }

  if (status === "fail" && !manifestMismatch && stepDossiers.length === 0) {
    stepCounts = {
      ...createCounts(),
      ...(stepCounts ?? {}),
      failed: (stepCounts?.failed ?? 0) + 1,
      non_test: (stepCounts?.non_test ?? 0) + 1,
      non_test_failed: (stepCounts?.non_test_failed ?? 0) + 1,
    };
    stepDossiers = [
      {
        failure_class: "harness",
        failure_reason: "tool_diagnostic_failure",
        coverage: "non_test",
        step: inferStepFromText(context.label),
        id: "",
        runner,
        package_or_file: `(${runner} runner)`,
        symbol_or_title: "(runner status)",
        message: `${runner} runner exited with status ${context.exitStatus} without selected test failures`,
        reproduce: context.command,
        raw: renderRawList(Object.values(artifacts ?? {})),
      },
    ];
  }

  writeStepArtifacts(context, {
    status,
    step: inferStepFromText(context.label),
    counts: stepCounts,
    owners: summary.owners,
    inventory: summary.inventory,
    dossiers: stepDossiers,
    ...extraWriteStepDetails,
    manifestSummary,
    manifestMismatch,
    artifacts,
  });

  if (status === "pass") {
    return 0;
  }
  if (manifestMismatch) {
    if (showStepDetailOutput(context)) {
      printBlock(`manifest mismatch: ${context.label}`, {
        missing_ids: renderList(manifestMismatch.missing_ids),
        unexpected_ids: renderList(manifestMismatch.unexpected_ids),
        forbidden_id_files: renderList(manifestMismatch.forbidden_id_files),
        ...manifestMismatchDetailFields(manifestMismatch),
      });
    }
    return 1;
  }
  if (showStepDetailOutput(context)) {
    for (const dossier of stepDossiers) {
      printBlock(`failure: ${context.label}`, failureDetailFields(dossier));
    }
  }
  return 1;
}

function normalizePlaywrightFile(file) {
  const normalized = normalizePath(file);
  if (normalized.startsWith("apps/web/")) {
    return normalized;
  }
  return normalizePath(path.join("apps/web", "e2e", normalized));
}

function isForbiddenFile(file, step) {
  if (!step) {
    return false;
  }
  const files = loadManifestIndex().forbiddenFilesByStep.get(step);
  return files ? files.has(file) : false;
}

function loadFrontendPlaywrightIndex() {
  return loadFrontendPlaywrightIndexAdapter(repoRoot);
}

function classifyPlaywrightCase(file, title, stepLabel) {
  const normalizedFile = normalizePlaywrightFile(file);
  const manifested = loadManifestIndex().manifestPlaywright.get(
    `${normalizedFile}::${title}`,
  );
  const frontendManifested = loadFrontendPlaywrightIndex().byTitle.get(title);
  const authoritative = loadManifestIndex().authoritativePlaywright.get(
    `${normalizedFile}::${title}`,
  );
  if (authoritative) {
    return {
      coverage: "authoritative",
      step: authoritative.step,
      id: authoritative.id,
      owner: normalizedFile,
    };
  }
  if (frontendManifested && /\bauthoritative\b/i.test(stepLabel)) {
    return {
      coverage: frontendManifested.coverage,
      step: frontendManifested.step,
      id: frontendManifested.id,
      owner: normalizedFile,
    };
  }
  if (manifested && manifested.coverage !== "authoritative") {
    return {
      coverage: "support",
      step: manifested.step,
      id: manifested.id,
      owner: normalizedFile,
    };
  }
  if (frontendManifested) {
    return {
      coverage: frontendManifested.coverage,
      step: frontendManifested.step,
      id: frontendManifested.id,
      owner: normalizedFile,
    };
  }
  const inferredStep =
    inferStepFromText(normalizedFile) ||
    inferStepFromText(title) ||
    inferStepFromText(stepLabel);
  if (claimsConformanceRowTitle(title)) {
    return {
      coverage: "unmapped",
      step: inferredStep,
      id: "",
      owner: normalizedFile,
    };
  }
  const support =
    normalizedFile.includes(".support.") ||
    isForbiddenFile(normalizedFile, inferredStep) ||
    /\bsupport\b/i.test(stepLabel) ||
    /\bsmoke\b/i.test(stepLabel);
  if (support) {
    return {
      coverage: "support",
      step: inferredStep,
      id: "",
      owner: normalizedFile,
    };
  }
  return (
    catalogFallbackClassification(
      "playwright",
      normalizedFile,
      title,
      inferredStep,
    ) ?? {
      coverage: "unmapped",
      step: inferredStep,
      id: "",
      owner: normalizedFile,
    }
  );
}

function parsePlaywrightStartTime(value) {
  const parsed = Date.parse(value ?? "");
  return Number.isFinite(parsed) ? parsed : null;
}

function updateTimingWindow(window, startMs, durationMs) {
  if (startMs === null) {
    return;
  }
  const endMs = startMs + clampDurationMs(durationMs);
  if (window.startMs === null || startMs < window.startMs) {
    window.startMs = startMs;
  }
  if (window.endMs === null || endMs > window.endMs) {
    window.endMs = endMs;
  }
}

function timingWindowDurationMs(window) {
  if (
    window.startMs === null ||
    window.endMs === null ||
    window.endMs < window.startMs
  ) {
    return 0;
  }
  return window.endMs - window.startMs;
}

function summarizePlaywrightTiming(specs, step, stepLabel, selection = null) {
  const files = new Map();
  const entryTimings = [];
  const totalWindow = { startMs: null, endMs: null };
  let tests = 0;
  let executedDurationMs = 0;
  let hasDuration = false;
  let hasTimestamp = false;

  for (const spec of specs) {
    const file = normalizePlaywrightFile(spec.file ?? "");
    const classification =
      selection?.classify?.(file, spec.title ?? "") ??
      classifyPlaywrightCase(spec.file ?? "", spec.title ?? "", stepLabel);
    if (!files.has(file)) {
      files.set(file, {
        file,
        tests: 0,
        executed_duration_ms: 0,
        wall_duration_ms: 0,
        window: { startMs: null, endMs: null },
        hasTimestamp: false,
      });
    }
    const fileTiming = files.get(file);
    const entryWindow = { startMs: null, endMs: null };
    let entryExecutedDurationMs = 0;
    let entryHasTimestamp = false;
    let executed = false;

    for (const test of spec.tests ?? []) {
      for (const result of test.results ?? []) {
        if (!result.status || result.status === "skipped") {
          continue;
        }
        executed = true;
        const durationMs = clampDurationMs(result.duration ?? 0);
        if (durationMs > 0) {
          hasDuration = true;
        }
        executedDurationMs += durationMs;
        fileTiming.executed_duration_ms += durationMs;
        entryExecutedDurationMs += durationMs;

        const startMs = parsePlaywrightStartTime(result.startTime);
        if (startMs !== null) {
          hasTimestamp = true;
          fileTiming.hasTimestamp = true;
          entryHasTimestamp = true;
          updateTimingWindow(totalWindow, startMs, durationMs);
          updateTimingWindow(fileTiming.window, startMs, durationMs);
          updateTimingWindow(entryWindow, startMs, durationMs);
        }
      }
    }

    if (executed) {
      tests += 1;
      fileTiming.tests += 1;
      if (classification.id) {
        entryTimings.push({
          id: classification.id,
          step: classification.step,
          file,
          title: spec.title ?? "(missing title)",
          executed_duration_ms: entryExecutedDurationMs,
          wall_duration_ms: entryHasTimestamp
            ? timingWindowDurationMs(entryWindow)
            : entryExecutedDurationMs,
        });
      }
    }
  }

  const fileSummaries = Array.from(files.values())
    .map((fileTiming) => ({
      file: fileTiming.file,
      tests: fileTiming.tests,
      executed_duration_ms: fileTiming.executed_duration_ms,
      wall_duration_ms: fileTiming.hasTimestamp
        ? timingWindowDurationMs(fileTiming.window)
        : fileTiming.executed_duration_ms,
    }))
    .sort((left, right) => left.file.localeCompare(right.file));

  return {
    step,
    files: fileSummaries,
    entries: entryTimings.sort((left, right) =>
      left.id.localeCompare(right.id, undefined, { numeric: true }),
    ),
    tests,
    executed_duration_ms: executedDurationMs,
    wall_duration_ms: hasTimestamp
      ? timingWindowDurationMs(totalWindow)
      : executedDurationMs,
    start_time:
      hasTimestamp && totalWindow.startMs !== null
        ? new Date(totalWindow.startMs).toISOString()
        : "",
    end_time:
      hasTimestamp && totalWindow.endMs !== null
        ? new Date(totalWindow.endMs).toISOString()
        : "",
    source: hasTimestamp
      ? "playwright_result_timestamps"
      : hasDuration
        ? "playwright_result_durations"
        : "step_window_fallback",
  };
}

function createPlaywrightSelection({ catalogAware }) {
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";

  if (catalogAware && reportSlice) {
    const scope = readManifestScopeEnv();
    const selectionReport = optionalEnv("CARTULARY_PLAYWRIGHT_SELECTION_REPORT");
    const reportSelection = readPlaywrightSelectionReport(
      selectionReport,
      scope,
    );
    const reportClassificationByKey = new Map(
      (reportSelection?.tests ?? []).map((test) => [
        `${test.file}::${test.title}`,
        {
          coverage: manifestCoverageToInventoryCoverage(
            test.coverage ?? scope.coverage,
          ),
          step: test.step,
          id: test.id,
          owner: test.file,
        },
      ]),
    );
    const selected = new Set(
      reportSelection
        ? reportSelection.tests.map((test) => `${test.file}::${test.title}`)
        : selectPlaywrightManifestEntries(
            scope.step,
            scope.coverage,
            scope.executionDependency,
          ).flatMap((entry) =>
            playwrightEntryTitles(entry).map(
              (title) => `${entry.file}::${title}`,
            ),
          ),
    );
    return {
      matches(normalizedFile, title) {
        return selected.has(`${normalizedFile}::${title}`);
      },
      classify(normalizedFile, title) {
        return reportClassificationByKey.get(`${normalizedFile}::${title}`) ?? null;
      },
    };
  }

  const selectedFiles = new Set(
    optionalLines("CARTULARY_PLAYWRIGHT_FILES").map((value) =>
      normalizePlaywrightFile(value),
    ),
  );
  const selectedTitles = optionalSetFromLines("CARTULARY_PLAYWRIGHT_TITLES");
  if (selectedFiles.size === 0 && selectedTitles.size === 0) {
    return null;
  }
  return {
    matches(normalizedFile, title) {
      if (selectedFiles.size > 0 && !selectedFiles.has(normalizedFile)) {
        return false;
      }
      if (selectedTitles.size > 0 && !selectedTitles.has(title)) {
        return false;
      }
      return true;
    },
  };
}

function summarizePlaywrightRun(reportFile, stepLabel, selection = null) {
  const report = JSON.parse(readFileSync(reportFile, "utf8"));
  const owners = new Set();
  const inventory = [];
  const dossiers = [];
  const counts = createCounts();

  const specs = flattenPlaywrightSuites(report.suites).filter((spec) => {
    if (!selection) {
      return true;
    }
    return selection.matches(
      normalizePlaywrightFile(spec.file ?? ""),
      spec.title ?? "",
    );
  });
  if (specs.length === 0 && (report.errors ?? []).length > 0) {
    const coverageOverride = optionalEnv("CARTULARY_ACCOUNTING_COVERAGE");
    const coverage =
      coverageOverride !== ""
        ? normalizeTestCoverage(coverageOverride)
        : /\bsupport\b/i.test(stepLabel)
          ? "support"
          : "unmapped";
    const message = (report.errors ?? [])
      .map((error) => error?.message)
      .find((entry) => typeof entry === "string" && entry.trim() !== "");
    dossiers.push({
      failure_class: "harness",
      coverage,
      step: inferStepFromText(stepLabel),
      id: "",
      runner: "playwright",
      package_or_file: "(playwright setup)",
      symbol_or_title: "(playwright setup)",
      message: (message ?? "playwright setup failure").split("\n")[0],
      reproduce: requiredEnv("CARTULARY_STEP_COMMAND"),
      raw: relToRepo(reportFile),
    });
    counts.failed += 1;
    addCoverageFailureCount(counts, coverage);
    return {
      report,
      counts,
      owners: [],
      inventory,
      dossiers,
    };
  }

  const topLevelErrorSummary = summarizePlaywrightErrors(report);
  if (topLevelErrorSummary !== "") {
    dossiers.push({
      failure_class: "harness",
      coverage: "non_test",
      step: inferStepFromText(stepLabel),
      id: "",
      runner: "playwright",
      package_or_file: "(playwright runner)",
      symbol_or_title: "(playwright runner)",
      message: topLevelErrorSummary.split("\n")[0],
      reproduce: requiredEnv("CARTULARY_STEP_COMMAND"),
      raw: relToRepo(reportFile),
    });
    counts.failed += 1;
    counts.non_test_failed += 1;
  }

  for (const spec of specs) {
    const normalizedFile = normalizePlaywrightFile(spec.file ?? "");
    const classification =
      selection?.classify?.(normalizedFile, spec.title ?? "") ??
      classifyPlaywrightCase(spec.file ?? "", spec.title ?? "", stepLabel);
    owners.add(classification.owner);
    const executedResults = [];
    for (const test of spec.tests ?? []) {
      for (const result of test.results ?? []) {
        if (result.status && result.status !== "skipped") {
          executedResults.push({
            retry: result.retry ?? 0,
            status: result.status,
            result,
          });
        }
      }
    }
    if (executedResults.length === 0) {
      continue;
    }

    counts.tests += 1;
    addCoverageCount(counts, classification.coverage);
    const successful = executedResults.some(
      (entry) => entry.status === "passed" || entry.status === "flaky",
    );
    if (successful) {
      inventory.push(
        createInventoryItem({
          coverage: classification.coverage,
          step: classification.step,
          id: classification.id,
          owner: classification.owner,
          name: spec.title ?? "(missing title)",
        }),
      );
    }

    if (
      successful &&
      executedResults.every(
        (entry) => entry.status === "passed" || entry.status === "flaky",
      )
    ) {
      continue;
    }

    counts.failed += 1;
    addCoverageFailureCount(counts, classification.coverage);
    for (const attempt of executedResults.filter(
      (entry) => entry.status !== "passed" && entry.status !== "flaky",
    )) {
      const attachments = (attempt.result.attachments ?? [])
        .map((attachment) => attachment?.path)
        .filter((entry) => typeof entry === "string" && entry !== "")
        .map((entry) => relToRepo(entry));
      const message =
        attempt.result.error?.message?.split("\n")[0] ||
        (attempt.result.errors ?? [])
          .map((error) => error?.message)
          .find((entry) => typeof entry === "string" && entry.trim() !== "")
          ?.split("\n")[0] ||
        `${spec.title ?? "playwright spec"} failed`;
      dossiers.push({
        coverage: classification.coverage,
        step: classification.step,
        id: classification.id,
        runner: "playwright",
        package_or_file: classification.owner,
        symbol_or_title: spec.title ?? "(missing title)",
        message: message.trim(),
        reproduce: `pnpm --dir apps/web exec playwright test ${classification.owner.replace(/^apps\/web\//, "")} -g '^${escapeSingleQuotes(spec.title ?? "")}$'`,
        raw:
          attachments.length > 0
            ? `${relToRepo(reportFile)};${attachments.join(";")}`
            : relToRepo(reportFile),
        retry: String(attempt.retry ?? 0),
      });
    }
  }

  counts.packages = owners.size;
  if (counts.tests === 0 && dossiers.length === 0) {
    const coverageOverride = optionalEnv("CARTULARY_ACCOUNTING_COVERAGE");
    const coverage =
      coverageOverride !== ""
        ? normalizeTestCoverage(coverageOverride)
        : /\bsupport\b/i.test(stepLabel)
          ? "support"
          : "unmapped";
    dossiers.push({
      coverage,
      step: inferStepFromText(stepLabel),
      id: "",
      runner: "playwright",
      package_or_file: "(playwright selection)",
      symbol_or_title: "(playwright selection)",
      message: "step matched zero tests",
      reproduce: requiredEnv("CARTULARY_STEP_COMMAND"),
      raw: relToRepo(reportFile),
    });
    counts.failed += 1;
    addCoverageFailureCount(counts, coverage);
  }
  return {
    report,
    counts,
    owners: Array.from(owners).sort(),
    inventory,
    dossiers,
    playwrightTiming: summarizePlaywrightTiming(
      specs,
      inferStepFromText(stepLabel),
      stepLabel,
      selection,
    ),
  };
}

function selectPlaywrightManifestEntries(step, coverage, executionDependency) {
  return selectPlaywrightManifestEntriesAdapter(repoRoot, {
    step,
    coverage,
    executionDependency,
  });
}

function readPlaywrightSelectionReport(reportFile, scope = null) {
  return readPlaywrightSelectionReportAdapter(repoRoot, reportFile, scope);
}

function selectedPlaywrightEntriesFromReport(reportFile, scope) {
  return selectedPlaywrightEntriesFromReportAdapter(repoRoot, reportFile, scope);
}

export function handlePlaywrightStep({ catalogAware }) {
  const context = createBaseStepContext("playwright");
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";
  const reportFile = requiredEnv("CARTULARY_STEP_RUNNER_LOG");
  const selectionReport = optionalEnv("CARTULARY_PLAYWRIGHT_SELECTION_REPORT");
  const stdoutLog = optionalEnv("CARTULARY_STEP_STDOUT_LOG");
  const stderrLog = optionalEnv("CARTULARY_STEP_STDERR_LOG");
  const outputDir = optionalEnv("CARTULARY_PLAYWRIGHT_OUTPUT_DIR");
  const serverLog = optionalEnv("CARTULARY_WEB_E2E_SERVER_LOG");
  const webLog = optionalEnv("CARTULARY_WEB_E2E_WEB_LOG");
  removeEmptyArtifact(stdoutLog);
  removeEmptyArtifact(stderrLog);

  const summary = summarizePlaywrightRun(
    reportFile,
    context.label,
    createPlaywrightSelection({ catalogAware }),
  );
  if (reportSlice && summary.playwrightTiming) {
    const timing = summary.playwrightTiming;
    const executedDurationMs = clampDurationMs(
      timing.executed_duration_ms ?? 0,
    );
    const wallDurationMs = clampDurationMs(timing.wall_duration_ms ?? 0);
    if (executedDurationMs > 0) {
      context.executedDurationMs = executedDurationMs;
      context.logicalDurationMs = executedDurationMs;
      context.reusedDurationMs =
        context.accountingMode === "reused" ? executedDurationMs : 0;
      context.derivedDurationMs =
        context.accountingMode === "derived" ? executedDurationMs : 0;
    }
    if (wallDurationMs > 0) {
      context.wallDurationMs = wallDurationMs;
    }
    if (timing.start_time && timing.end_time) {
      context.startTime = timing.start_time;
      context.endTime = timing.end_time;
    }
  }
  const selectedSlicePassed =
    summary.dossiers.length === 0 && (context.exitStatus === 0 || reportSlice);

  return finalizeManifestAwareRunnerStep(context, {
    catalogAware,
    runner: "playwright",
    summary,
    selectedSlicePassed,
    artifacts: {
      selected_tests_json: selectionReport,
      runner_json: reportFile,
      stdout_log: existsSync(stdoutLog) ? stdoutLog : "",
      stderr_log: existsSync(stderrLog) ? stderrLog : "",
      playwright_output: outputDir,
      server_log: serverLog,
      web_log: webLog,
    },
    extraWriteStepDetails: {
      playwrightTiming: summary.playwrightTiming,
    },
    manifestMismatchArtifacts: () => ({
      selection:
        selectionReport && existsSync(selectionReport)
          ? relToRepo(selectionReport)
          : "",
      runner: relToRepo(reportFile),
    }),
    manifestMismatchDetailFields: (manifestMismatch) => ({
      selection: manifestMismatch.selection,
      runner: manifestMismatch.runner,
    }),
    failureDetailFields: (dossier) => {
      const { runner, raw, ...rest } = dossier;
      return {
        ...rest,
        test_runner: runner,
        selection:
          selectionReport && existsSync(selectionReport)
            ? relToRepo(selectionReport)
            : "",
        runner: relToRepo(reportFile),
        raw: mergeRawPaths(
          raw,
          renderRawList([reportFile, outputDir, serverLog, webLog]),
        ),
      };
    },
  });
}

function renderRawList(paths) {
  return paths
    .filter((entry) => entry && existsSync(entry))
    .map((entry) => relToRepo(entry))
    .join(";");
}

function mergeRawPaths(...values) {
  return values
    .filter(Boolean)
    .join(";")
    .split(";")
    .map((entry) => entry.trim())
    .filter(Boolean)
    .filter((entry, index, array) => array.indexOf(entry) === index)
    .join(";");
}

function escapeSingleQuotes(value) {
  return value.replaceAll("'", "'\"'\"'");
}
