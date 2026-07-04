#!/usr/bin/env node

import {
  existsSync,
  readFileSync,
  rmSync,
  statSync,
} from "node:fs";
import path from "node:path";
import { loadFrontendPlaywrightIndex as loadFrontendPlaywrightIndexAdapter } from "../frontend-indexes.mjs";
import { validateSchemaSync } from "../../../contract/harness-contract.mjs";
import {
  loadManifestIndex as loadManifestIndexAdapter,
  playwrightEntryTitles,
  selectManifestEntries,
  selectPlaywrightManifestEntries as selectPlaywrightManifestEntriesAdapter,
  selectPlaywrightEntries,
  vitestEntryTitles,
} from "../phase-manifest-adapter.mjs";
import {
  flattenPlaywrightSuites,
  summarizePlaywrightErrors,
} from "../../../browser/playwright-report.mjs";
import {
  readPlaywrightSelectionReport as readPlaywrightSelectionReportAdapter,
  selectedPlaywrightEntriesFromReport as selectedPlaywrightEntriesFromReportAdapter,
} from "../../../browser/playwright-selection.mjs";
import { verboseOutput } from "../../tool-output.mjs";
import {
  repoRoot,
  testAccountingClassificationSchemaID,
  testCoverageBucketSet,
  testCoverageBuckets,
} from "../../../contract/test-output-context.mjs";
import {
  createBasePhaseContext,
  writePhaseArtifacts,
} from "../phase-artifacts.mjs";

let cachedGoModulePath;

let cachedTestAccountingClassification;

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

function globToRegExp(pattern) {
  const escaped = String(pattern)
    .replace(/[.+^${}()|[\]\\]/g, "\\$&")
    .replaceAll("**", "\u0000")
    .replaceAll("*", "[^/]*")
    .replaceAll("\u0000", ".*");
  return new RegExp(`^${escaped}$`);
}

function normalizeAccountingRule(rule, label, scope) {
  if (!rule || typeof rule !== "object") {
    throw new Error(`${label} must be an object`);
  }
  const requiresFile = scope === "vitest" || scope === "playwright";
  const requiresPackage = scope === "go_packages" || scope === "go_tests";
  if (
    requiresFile &&
    typeof rule.file !== "string" &&
    typeof rule.file_pattern !== "string"
  ) {
    throw new Error(`${label} must declare file or file_pattern`);
  }
  if (
    requiresPackage &&
    typeof rule.package !== "string" &&
    typeof rule.package_pattern !== "string"
  ) {
    throw new Error(`${label} must declare package or package_pattern`);
  }
  const coverage = normalizeTestCoverage(rule.coverage, "");
  if (coverage === "unmapped") {
    throw new Error(
      `${label}.coverage must be raw|tooling_support|unowned_regression|support|authoritative`,
    );
  }
  if (typeof rule.target !== "string" || rule.target.trim() === "") {
    throw new Error(`${label}.target must declare the owning Make target`);
  }
  if (typeof rule.reason !== "string" || rule.reason.trim().length < 20) {
    throw new Error(`${label}.reason must explain the ownership decision`);
  }
  if (typeof rule.title === "string" && typeof rule.title_pattern === "string") {
    throw new Error(`${label} must not declare both title and title_pattern`);
  }
  return {
    ...rule,
    coverage,
    phase: typeof rule.phase === "string" ? rule.phase : "",
    reason: rule.reason.trim(),
  };
}

function loadTestAccountingClassification() {
  if (cachedTestAccountingClassification) {
    return cachedTestAccountingClassification;
  }
  const file = path.join(
    repoRoot,
    "tools",
    "test_accounting_classification.json",
  );
  const empty = {
    vitest: [],
    go_packages: [],
    go_tests: [],
    playwright: [],
  };
  if (!existsSync(file)) {
    cachedTestAccountingClassification = empty;
    return cachedTestAccountingClassification;
  }
  const manifest = JSON.parse(readFileSync(file, "utf8"));
  if (manifest.schema_id !== testAccountingClassificationSchemaID) {
    throw new Error(
      `${file} must declare schema_id ${testAccountingClassificationSchemaID}`,
    );
  }
  validateSchemaSync(testAccountingClassificationSchemaID, manifest);
  cachedTestAccountingClassification = {
    vitest: (manifest.vitest ?? []).map((rule, index) =>
      normalizeAccountingRule(rule, `${file}.vitest[${index}]`, "vitest"),
    ),
    go_packages: (manifest.go_packages ?? []).map((rule, index) =>
      normalizeAccountingRule(
        rule,
        `${file}.go_packages[${index}]`,
        "go_packages",
      ),
    ),
    go_tests: (manifest.go_tests ?? []).map((rule, index) =>
      normalizeAccountingRule(rule, `${file}.go_tests[${index}]`, "go_tests"),
    ),
    playwright: (manifest.playwright ?? []).map((rule, index) =>
      normalizeAccountingRule(
        rule,
        `${file}.playwright[${index}]`,
        "playwright",
      ),
    ),
  };
  return cachedTestAccountingClassification;
}

function ruleStringMatches(rule, exactKey, patternKey, value) {
  const normalized = normalizePath(value ?? "");
  if (
    typeof rule[exactKey] === "string" &&
    normalizePath(rule[exactKey]) !== normalized
  ) {
    return false;
  }
  if (
    typeof rule[patternKey] === "string" &&
    !globToRegExp(normalizePath(rule[patternKey])).test(normalized)
  ) {
    return false;
  }
  return true;
}

function ruleTitleMatches(rule, title) {
  if (typeof rule.title === "string" && rule.title !== title) {
    return false;
  }
  if (
    typeof rule.title_pattern === "string" &&
    !globToRegExp(rule.title_pattern).test(title)
  ) {
    return false;
  }
  return true;
}

function ruleTargetMatches(rule) {
  const target = optionalEnv("CARTULARY_TEST_TARGET");
  return typeof rule.target !== "string" || rule.target === target;
}

function accountingManifestClassification(
  runner,
  owner,
  title = "",
  fallbackPhase = "",
) {
  const manifest = loadTestAccountingClassification();
  const rules = manifest[runner] ?? [];
  for (const rule of rules) {
    if (!ruleTargetMatches(rule)) {
      continue;
    }
    if (runner === "vitest" || runner === "playwright") {
      if (!ruleStringMatches(rule, "file", "file_pattern", owner)) {
        continue;
      }
      if (!ruleTitleMatches(rule, title)) {
        continue;
      }
    } else if (runner === "go_packages") {
      if (!ruleStringMatches(rule, "package", "package_pattern", owner)) {
        continue;
      }
    } else if (runner === "go_tests") {
      if (!ruleStringMatches(rule, "package", "package_pattern", owner)) {
        continue;
      }
      if (!ruleTitleMatches(rule, title)) {
        continue;
      }
    }
    return {
      coverage: rule.coverage,
      phase: rule.phase || fallbackPhase,
      id: "",
      owner,
    };
  }
  return null;
}

function inferPhaseFromText(value) {
  if (!value) {
    return "";
  }
  const patterns = [
    /\bphase(?:\s|_|-)?(\d+)\b/i,
    /\b[UIE][-_](\d+)-\d+\b/,
    /\b[UIE]_(\d+)_\d+\b/,
  ];
  for (const pattern of patterns) {
    const match = value.match(pattern);
    if (match) {
      return `phase${match[1]}`;
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

function showPhaseDetailOutput(context) {
  return verboseOutput() || context.target === "adhoc";
}

function createInventoryItem({ coverage, phase, id, owner, name }) {
  return {
    coverage,
    phase,
    id: id ?? "",
    package_or_file: owner,
    symbol_or_title: name,
  };
}

function readManifestScopeEnv() {
  return {
    phase: requiredEnv("CARTULARY_MANIFEST_PHASE"),
    coverage: requiredEnv("CARTULARY_MANIFEST_COVERAGE"),
    executionDependency: optionalEnv("CARTULARY_MANIFEST_EXECUTION_DEPENDENCY"),
  };
}

function manifestCoverageToInventoryCoverage(coverage) {
  return coverage === "authoritative" ? "authoritative" : "support";
}

function evaluateFlatTitleManifest(
  summary,
  { phase, entries, inventoryCoverage },
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
    phase,
    missingIDs,
    unexpectedIDs,
    forbiddenIDFiles: Array.from(
      loadManifestIndex().forbiddenFilesByPhase.get(phase) ?? [],
    ).sort(),
  };
}

function finalizeManifestAwareRunnerPhase(
  context,
  {
    manifestAware,
    runner,
    section = "",
    summary,
    selectedSlicePassed,
    artifacts,
    manifestMismatchArtifacts = () => ({}),
    manifestMismatchDetailFields = () => ({}),
    failureDetailFields = (dossier) => dossier,
    extraWritePhaseDetails = {},
  },
) {
  let status = selectedSlicePassed ? "pass" : "fail";
  let manifestSummary = null;
  let manifestMismatch = null;
  let phaseCounts = summary.counts;
  let phaseDossiers = summary.dossiers;
  const emptySelectionAllowed =
    runner === "vitest" &&
    optionalEnv("CARTULARY_VITEST_ALLOW_EMPTY_SELECTION") === "1" &&
    (summary.counts?.tests ?? 0) === 0 &&
    summary.dossiers.length === 0;

  if (manifestAware && selectedSlicePassed && !emptySelectionAllowed) {
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
            repoRoot,
            scope.phase,
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
      phase: scope.phase,
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

  if (status === "fail" && !manifestMismatch && phaseDossiers.length === 0) {
    phaseCounts = {
      ...createCounts(),
      ...(phaseCounts ?? {}),
      failed: (phaseCounts?.failed ?? 0) + 1,
      non_test: (phaseCounts?.non_test ?? 0) + 1,
      non_test_failed: (phaseCounts?.non_test_failed ?? 0) + 1,
    };
    phaseDossiers = [
      {
        failure_class: "harness",
        failure_reason: "tool_diagnostic_failure",
        coverage: "non_test",
        phase: inferPhaseFromText(context.label),
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

  writePhaseArtifacts(context, {
    status,
    phase: inferPhaseFromText(context.label),
    counts: phaseCounts,
    owners: summary.owners,
    inventory: summary.inventory,
    dossiers: phaseDossiers,
    ...extraWritePhaseDetails,
    manifestSummary,
    manifestMismatch,
    artifacts,
  });

  if (status === "pass") {
    return 0;
  }
  if (manifestMismatch) {
    if (showPhaseDetailOutput(context)) {
      printBlock(`manifest mismatch: ${context.label}`, {
        missing_ids: renderList(manifestMismatch.missing_ids),
        unexpected_ids: renderList(manifestMismatch.unexpected_ids),
        forbidden_id_files: renderList(manifestMismatch.forbidden_id_files),
        ...manifestMismatchDetailFields(manifestMismatch),
      });
    }
    return 1;
  }
  if (showPhaseDetailOutput(context)) {
    for (const dossier of phaseDossiers) {
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

function isForbiddenFile(file, phase) {
  if (!phase) {
    return false;
  }
  const files = loadManifestIndex().forbiddenFilesByPhase.get(phase);
  return files ? files.has(file) : false;
}

function loadFrontendPlaywrightIndex() {
  return loadFrontendPlaywrightIndexAdapter(repoRoot);
}

function classifyPlaywrightCase(file, title, phaseLabel) {
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
      phase: authoritative.phase,
      id: authoritative.id,
      owner: normalizedFile,
    };
  }
  if (frontendManifested && /\bauthoritative\b/i.test(phaseLabel)) {
    return {
      coverage: frontendManifested.coverage,
      phase: frontendManifested.phase,
      id: frontendManifested.id,
      owner: normalizedFile,
    };
  }
  if (manifested && manifested.coverage !== "authoritative") {
    return {
      coverage: "support",
      phase: manifested.phase,
      id: manifested.id,
      owner: normalizedFile,
    };
  }
  if (frontendManifested) {
    return {
      coverage: frontendManifested.coverage,
      phase: frontendManifested.phase,
      id: frontendManifested.id,
      owner: normalizedFile,
    };
  }
  const inferredPhase =
    inferPhaseFromText(normalizedFile) ||
    inferPhaseFromText(title) ||
    inferPhaseFromText(phaseLabel);
  if (claimsConformanceRowTitle(title)) {
    return {
      coverage: "unmapped",
      phase: inferredPhase,
      id: "",
      owner: normalizedFile,
    };
  }
  const support =
    normalizedFile.includes(".support.") ||
    isForbiddenFile(normalizedFile, inferredPhase) ||
    /\bsupport\b/i.test(phaseLabel) ||
    /\bsmoke\b/i.test(phaseLabel);
  if (support) {
    return {
      coverage: "support",
      phase: inferredPhase,
      id: "",
      owner: normalizedFile,
    };
  }
  return (
    accountingManifestClassification(
      "playwright",
      normalizedFile,
      title,
      inferredPhase,
    ) ?? {
      coverage: "unmapped",
      phase: inferredPhase,
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

function summarizePlaywrightTiming(specs, phase, phaseLabel, selection = null) {
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
      classifyPlaywrightCase(spec.file ?? "", spec.title ?? "", phaseLabel);
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
          phase: classification.phase,
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
    phase,
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
        : "phase_window_fallback",
  };
}

function createPlaywrightSelection({ manifestAware }) {
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";

  if (manifestAware && reportSlice) {
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
          phase: test.phase,
          id: test.id,
          owner: test.file,
        },
      ]),
    );
    const selected = new Set(
      reportSelection
        ? reportSelection.tests.map((test) => `${test.file}::${test.title}`)
        : selectPlaywrightManifestEntries(
            scope.phase,
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

function summarizePlaywrightRun(reportFile, phaseLabel, selection = null) {
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
        : /\bsupport\b/i.test(phaseLabel)
          ? "support"
          : "unmapped";
    const message = (report.errors ?? [])
      .map((error) => error?.message)
      .find((entry) => typeof entry === "string" && entry.trim() !== "");
    dossiers.push({
      failure_class: "harness",
      coverage,
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "playwright",
      package_or_file: "(playwright setup)",
      symbol_or_title: "(playwright setup)",
      message: (message ?? "playwright setup failure").split("\n")[0],
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
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
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "playwright",
      package_or_file: "(playwright runner)",
      symbol_or_title: "(playwright runner)",
      message: topLevelErrorSummary.split("\n")[0],
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
      raw: relToRepo(reportFile),
    });
    counts.failed += 1;
    counts.non_test_failed += 1;
  }

  for (const spec of specs) {
    const normalizedFile = normalizePlaywrightFile(spec.file ?? "");
    const classification =
      selection?.classify?.(normalizedFile, spec.title ?? "") ??
      classifyPlaywrightCase(spec.file ?? "", spec.title ?? "", phaseLabel);
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
          phase: classification.phase,
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
        phase: classification.phase,
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
        : /\bsupport\b/i.test(phaseLabel)
          ? "support"
          : "unmapped";
    dossiers.push({
      coverage,
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "playwright",
      package_or_file: "(playwright selection)",
      symbol_or_title: "(playwright selection)",
      message: "phase matched zero tests",
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
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
      inferPhaseFromText(phaseLabel),
      phaseLabel,
      selection,
    ),
  };
}

function selectPlaywrightManifestEntries(phase, coverage, executionDependency) {
  return selectPlaywrightManifestEntriesAdapter(repoRoot, {
    phase,
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

export function handlePlaywrightPhase({ manifestAware }) {
  const context = createBasePhaseContext("playwright");
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";
  const reportFile = requiredEnv("CARTULARY_PHASE_RUNNER_LOG");
  const selectionReport = optionalEnv("CARTULARY_PLAYWRIGHT_SELECTION_REPORT");
  const stdoutLog = optionalEnv("CARTULARY_PHASE_STDOUT_LOG");
  const stderrLog = optionalEnv("CARTULARY_PHASE_STDERR_LOG");
  const outputDir = optionalEnv("CARTULARY_PLAYWRIGHT_OUTPUT_DIR");
  const serverLog = optionalEnv("CARTULARY_WEB_E2E_SERVER_LOG");
  const webLog = optionalEnv("CARTULARY_WEB_E2E_WEB_LOG");
  removeEmptyArtifact(stdoutLog);
  removeEmptyArtifact(stderrLog);

  const summary = summarizePlaywrightRun(
    reportFile,
    context.label,
    createPlaywrightSelection({ manifestAware }),
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

  return finalizeManifestAwareRunnerPhase(context, {
    manifestAware,
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
    extraWritePhaseDetails: {
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
