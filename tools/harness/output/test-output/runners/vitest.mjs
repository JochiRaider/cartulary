#!/usr/bin/env node

import {
  existsSync,
  readFileSync,
  rmSync,
  statSync,
} from "node:fs";
import path from "node:path";
import { loadFrontendVitestIndex as loadFrontendVitestIndexAdapter } from "../../../frontend/evidence/test-output-indexes.mjs";
import { validateSchemaSync } from "../../../contract/harness-contract.mjs";
import {
  collectVitestManifestEntries as collectVitestManifestEntriesAdapter,
  loadManifestIndex as loadManifestIndexAdapter,
  playwrightEntryTitles,
  selectManifestEntries,
  selectPlaywrightEntries,
  selectVitestManifestEntries as selectVitestManifestEntriesAdapter,
  vitestEntryTitles,
} from "../../../planning/test-output-phase-manifest.mjs";
import { selectedPlaywrightEntriesFromReport as selectedPlaywrightEntriesFromReportAdapter } from "../../../browser/playwright-selection.mjs";
import { verboseOutput } from "../../tool-output.mjs";
import {
  repoRoot,
  testAccountingClassificationSchemaID,
  testCoverageBucketSet,
  testCoverageBuckets,
  vitestFailureDetailsSchemaID,
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

function supportNamedTitle(value) {
  return /^Phase\s+\d+\s+support\b/i.test(value);
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

function classifyVitestCase(ownerPath, title, phaseLabel) {
  const manifestFile = vitestOwnerToSelectionFile(ownerPath);
  const manifested = loadManifestIndex().manifestVitest.get(
    `${manifestFile}::${title}`,
  );
  if (manifested && manifested.coverage !== "authoritative") {
    return {
      coverage: "support",
      phase: manifested.phase,
      id: manifested.id,
      owner: ownerPath,
    };
  }
  const authoritative = loadManifestIndex().authoritativeVitest.get(
    `${manifestFile}::${title}`,
  );
  if (authoritative) {
    return {
      coverage: "authoritative",
      phase: authoritative.phase,
      id: authoritative.id,
      owner: ownerPath,
    };
  }
  const frontendManifested = loadFrontendVitestIndex().byTitle.get(title);
  if (frontendManifested) {
    return {
      coverage: frontendManifested.coverage,
      phase: frontendManifested.phase,
      id: frontendManifested.id,
      owner: ownerPath,
    };
  }
  const inferredPhase =
    inferPhaseFromText(ownerPath) ||
    inferPhaseFromText(title) ||
    inferPhaseFromText(phaseLabel);
  if (claimsConformanceRowTitle(title)) {
    return {
      coverage: "unmapped",
      phase: inferredPhase,
      id: "",
      owner: ownerPath,
    };
  }
  const support =
    ownerPath.includes(".support.") ||
    supportNamedTitle(title) ||
    isForbiddenFile(
      ownerPath,
      inferPhaseFromText(ownerPath) || inferPhaseFromText(title),
    ) ||
    /\bsupport\b/i.test(phaseLabel);
  if (support) {
    return {
      coverage: "support",
      phase: inferredPhase,
      id: "",
      owner: ownerPath,
    };
  }
  return (
    accountingManifestClassification(
      "vitest",
      ownerPath,
      title,
      inferredPhase,
    ) ?? {
      coverage: "unmapped",
      phase: inferredPhase,
      id: "",
      owner: ownerPath,
    }
  );
}

function normalizeVitestOwnerPath(filePath) {
  return relToRepo(filePath);
}

function normalizeVitestSelectionFile(filePath) {
  const relative = relToRepo(filePath);
  if (relative === "") {
    return "";
  }
  if (relative.startsWith("apps/web/")) {
    return relative;
  }
  return normalizePath(path.join("apps/web", relative));
}

function vitestOwnerToSelectionFile(ownerPath) {
  if (ownerPath === "") {
    return "";
  }
  if (ownerPath.startsWith("apps/web/")) {
    return ownerPath;
  }
  return normalizePath(path.join("apps/web", ownerPath));
}

function vitestOwnerToReproducePath(ownerPath) {
  if (ownerPath.startsWith("apps/web/")) {
    return ownerPath.slice("apps/web/".length);
  }
  if (ownerPath.startsWith("packages/")) {
    return `../../${ownerPath}`;
  }
  return ownerPath;
}

function renderVitestReproduceCommand(ownerPath, title = "") {
  const reproducePath = vitestOwnerToReproducePath(ownerPath);
  if (title === "") {
    return `pnpm --dir apps/web exec vitest run ${reproducePath}`;
  }
  return `pnpm --dir apps/web exec vitest run ${reproducePath} -t '${escapeSingleQuotes(title)}$'`;
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

function createVitestSelection({ manifestAware }) {
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";

  if (manifestAware && reportSlice) {
    const { phase, coverage, executionDependency } = readManifestScopeEnv();
    const entries = selectVitestManifestEntries(
      phase,
      coverage,
      executionDependency,
    );
    const selected = new Set(
      entries.flatMap((entry) =>
        vitestEntryTitles(entry).map((title) => `${entry.file}::${title}`),
      ),
    );
    const selectedFiles = new Set(
      entries.map((entry) => normalizePath(entry.file)),
    );
    return {
      matches(ownerPath, title) {
        return selected.has(
          `${vitestOwnerToSelectionFile(ownerPath)}::${title}`,
        );
      },
      matchesFile(ownerPath) {
        return selectedFiles.has(vitestOwnerToSelectionFile(ownerPath));
      },
      classifyFileFailure(ownerPath) {
        if (!selectedFiles.has(vitestOwnerToSelectionFile(ownerPath))) {
          return null;
        }
        return {
          coverage,
          phase,
          id: "",
          owner: ownerPath,
        };
      },
    };
  }

  const excludedManifestDependency = optionalEnv(
    "CARTULARY_VITEST_EXCLUDE_MANIFEST_EXECUTION_DEPENDENCY",
  );
  const excluded = new Set();
  const excludedFiles = new Set();
  if (excludedManifestDependency !== "") {
    for (const entry of collectVitestManifestEntries(
      "authoritative",
      excludedManifestDependency,
    )) {
      for (const title of vitestEntryTitles(entry)) {
        excluded.add(`${normalizePath(entry.file)}::${title}`);
      }
      excludedFiles.add(normalizePath(entry.file));
    }
  }

  const selectedFiles = new Set(
    optionalLines("CARTULARY_VITEST_FILES").map((value) =>
      normalizeVitestSelectionFile(value),
    ),
  );
  const selectedTitles = optionalSetFromLines("CARTULARY_VITEST_TITLES");
  if (
    selectedFiles.size === 0 &&
    selectedTitles.size === 0 &&
    excluded.size === 0
  ) {
    return null;
  }
  return {
    matches(ownerPath, title) {
      if (excluded.has(`${vitestOwnerToSelectionFile(ownerPath)}::${title}`)) {
        return false;
      }
      if (
        selectedFiles.size > 0 &&
        !selectedFiles.has(vitestOwnerToSelectionFile(ownerPath))
      ) {
        return false;
      }
      if (selectedTitles.size > 0 && !selectedTitles.has(title)) {
        return false;
      }
      return true;
    },
    matchesFile(ownerPath) {
      if (excludedFiles.has(vitestOwnerToSelectionFile(ownerPath))) {
        return false;
      }
      if (selectedFiles.size === 0) {
        return selectedTitles.size === 0;
      }
      return selectedFiles.has(vitestOwnerToSelectionFile(ownerPath));
    },
    classifyFileFailure() {
      return null;
    },
  };
}

function isVitestFileResult(value) {
  return Boolean(
    value &&
      typeof value === "object" &&
      typeof value.name === "string" &&
      Array.isArray(value.assertionResults),
  );
}

function collectVitestFileResults(report) {
  const fileResults = [];
  const visited = new Set();
  appendVitestFileResults(report, fileResults, visited);
  return fileResults;
}

function appendVitestFileResults(value, fileResults, visited) {
  if (!value || typeof value !== "object") {
    return;
  }
  if (isVitestFileResult(value)) {
    fileResults.push(value);
    return;
  }
  if (visited.has(value)) {
    return;
  }
  visited.add(value);
  if (Array.isArray(value)) {
    for (const entry of value) {
      appendVitestFileResults(entry, fileResults, visited);
    }
    return;
  }
  for (const key of ["testResults", "projectResults", "projects", "results"]) {
    if (!Array.isArray(value[key])) {
      continue;
    }
    for (const entry of value[key]) {
      appendVitestFileResults(entry, fileResults, visited);
    }
  }
}

function findVitestAuthoritativeFileEntry(ownerPath, phaseLabel) {
  const manifestFile = vitestOwnerToSelectionFile(ownerPath);
  const inferredPhase =
    inferPhaseFromText(ownerPath) || inferPhaseFromText(phaseLabel);
  let fallback = null;

  for (const entry of loadManifestIndex().authoritativeVitest.values()) {
    if (normalizePath(entry.file) !== manifestFile) {
      continue;
    }
    if (!fallback) {
      fallback = entry;
    }
    if (inferredPhase && entry.phase === inferredPhase) {
      return entry;
    }
  }

  return fallback;
}

function classifyVitestFileFailure(ownerPath, phaseLabel, selection = null) {
  const selected = selection?.classifyFileFailure?.(ownerPath);
  if (selected) {
    return selected;
  }

  const authoritative = findVitestAuthoritativeFileEntry(ownerPath, phaseLabel);
  if (authoritative) {
    return {
      coverage: "authoritative",
      phase: authoritative.phase,
      id: "",
      owner: ownerPath,
    };
  }

  const inferredPhase =
    inferPhaseFromText(ownerPath) || inferPhaseFromText(phaseLabel);
  const support =
    ownerPath.includes(".support.") ||
    isForbiddenFile(ownerPath, inferredPhase) ||
    /\bsupport\b/i.test(phaseLabel);
  if (support) {
    return {
      coverage: "support",
      phase: inferredPhase,
      id: "",
      owner: ownerPath,
    };
  }
  return (
    accountingManifestClassification(
      "vitest",
      ownerPath,
      "",
      inferredPhase,
    ) ?? {
      coverage: "unmapped",
      phase: inferredPhase,
      id: "",
      owner: ownerPath,
    }
  );
}

function firstVitestAppFrame(message) {
  const frame = String(message)
    .split("\n")
    .map((line) => line.trim())
    .find((line) =>
      /(?:^at\s+|^\()\/?home\/.*\/cartulary\/apps\/web\/src\//.test(line),
    );
  if (!frame) {
    return "";
  }
  const match = frame.match(/(apps\/web\/src\/[^:)]+):([0-9]+)(?::[0-9]+)?/);
  if (!match) {
    return "";
  }
  return `${match[1]}:${match[2]}`;
}

function normalizeFailureMessages(failureMessage, failureMessages = []) {
  const messages = Array.isArray(failureMessages)
    ? failureMessages.filter((entry) => typeof entry === "string")
    : [];
  if (messages.length > 0) {
    return messages;
  }
  return typeof failureMessage === "string" && failureMessage !== ""
    ? [failureMessage]
    : [];
}

function vitestDiagnosticTags(messageOrMessages) {
  const message = Array.isArray(messageOrMessages)
    ? messageOrMessages.join("\n")
    : String(messageOrMessages ?? "");
  const tags = [];
  if (message.includes("STACK_TRACE_ERROR")) {
    tags.push("vitest_stack_trace_error");
  }
  if (
    message.includes('Unable to find an element by: [data-testid="row-') ||
    message.includes("Expected workbook rows for surface")
  ) {
    tags.push("workbook_row_hydration_wait");
  }
  if (
    message.includes("controlled_input_replacement_mismatch") ||
    message.includes("Expected input value")
  ) {
    tags.push("controlled_input_replacement");
  }
  return tags;
}

function mergeVitestDiagnosticTags(...tagGroups) {
  const tags = new Set();
  for (const group of tagGroups) {
    const entries = Array.isArray(group) ? group : [group];
    for (const entry of entries) {
      if (typeof entry === "string" && entry !== "") {
        tags.add(entry);
      }
    }
  }
  return Array.from(tags).sort();
}

function loadVitestFailureDetails(file) {
  if (!file || !existsSync(file)) {
    return null;
  }
  const details = JSON.parse(readFileSync(file, "utf8"));
  validateSchemaSync(vitestFailureDetailsSchemaID, details);
  return details;
}

function vitestFailureDetailsKey(ownerPath, title) {
  return `${normalizePath(ownerPath)}::${String(title ?? "")}`;
}

function indexVitestFailureDetails(details) {
  const index = new Map();
  for (const failure of details?.failures ?? []) {
    const ownerPath = normalizePath(failure.owner_path ?? "");
    const title = String(failure.title ?? "");
    for (const candidate of [
      ownerPath,
      vitestOwnerToSelectionFile(ownerPath),
      ownerPath.startsWith("apps/web/")
        ? ownerPath.slice("apps/web/".length)
        : "",
    ]) {
      if (candidate) {
        index.set(vitestFailureDetailsKey(candidate, title), failure);
      }
    }
  }
  return index;
}

function vitestFailureDetailsEntry(detailsIndex, ownerPath, title) {
  if (!detailsIndex) {
    return null;
  }
  return (
    detailsIndex.get(vitestFailureDetailsKey(ownerPath, title)) ??
    detailsIndex.get(
      vitestFailureDetailsKey(vitestOwnerToSelectionFile(ownerPath), title),
    ) ??
    null
  );
}

function summarizeVitestFailureMessage({
  fallback,
  failureMessage = "",
  failureMessages = [],
  ownerPath,
  sidecarMessage = "",
  title,
}) {
  const sidecar = String(sidecarMessage ?? "").trim();
  if (sidecar !== "") {
    return sidecar;
  }
  const messages = normalizeFailureMessages(failureMessage, failureMessages);
  for (const message of messages) {
    for (const line of message.split("\n")) {
      const trimmed = line.trim();
      if (
        trimmed &&
        trimmed !== "Error: STACK_TRACE_ERROR" &&
        !trimmed.startsWith("at ")
      ) {
        return trimmed;
      }
    }
  }
  const combinedMessage = messages.join("\n");
  if (combinedMessage.includes("STACK_TRACE_ERROR")) {
    const appFrame = firstVitestAppFrame(combinedMessage);
    return [
      "Vitest reporter emitted STACK_TRACE_ERROR before preserving the assertion message",
      `file=${ownerPath || "(unknown)"}`,
      `title=${title || "(unknown)"}`,
      appFrame ? `first_app_frame=${appFrame}` : "",
    ]
      .filter(Boolean)
      .join("; ");
  }
  return fallback;
}

function summarizeVitestRun(
  reportFile,
  phaseLabel,
  selection = null,
  rawFiles = [reportFile],
  failureDetailsFile = "",
) {
  const report = JSON.parse(readFileSync(reportFile, "utf8"));
  const failureDetails = loadVitestFailureDetails(failureDetailsFile);
  const failureDetailsIndex = indexVitestFailureDetails(failureDetails);
  const owners = new Set();
  const inventory = [];
  const dossiers = [];
  const counts = createCounts();
  const rawArtifacts = renderRawList(rawFiles) || relToRepo(reportFile);

  for (const fileResult of collectVitestFileResults(report)) {
    const ownerPath = normalizeVitestOwnerPath(fileResult.name ?? "");
    const assertions = fileResult.assertionResults ?? [];
    const executedAssertions = assertions.filter(
      (assertion) => assertion.status !== "skipped",
    );
    if (executedAssertions.length === 0 && fileResult.status === "failed") {
      if (selection && !selection.matchesFile(ownerPath)) {
        continue;
      }
      const classification = classifyVitestFileFailure(
        ownerPath,
        phaseLabel,
        selection,
      );
      owners.add(classification.owner);
      counts.failed += 1;
      addCoverageFailureCount(counts, classification.coverage);
      const failureMessage = fileResult.message ?? "";
      const sidecarFailure = vitestFailureDetailsEntry(
        failureDetailsIndex,
        classification.owner,
        "(suite load)",
      );
      const sidecarMessage = sidecarFailure?.message ?? "";
      dossiers.push({
        coverage: classification.coverage,
        phase: classification.phase,
        id: classification.id,
        runner: "vitest",
        package_or_file: classification.owner,
        symbol_or_title: "(suite load)",
        message: summarizeVitestFailureMessage({
          fallback: `test file ${classification.owner} failed before a top-level test was attributed`,
          failureMessage,
          ownerPath: classification.owner,
          sidecarMessage,
          title: "(suite load)",
        }),
        diagnostic_tags: mergeVitestDiagnosticTags(
          vitestDiagnosticTags([failureMessage, sidecarMessage]),
          sidecarFailure?.diagnostic_tags ?? [],
          sidecarFailure ? ["vitest_failure_sidecar"] : [],
        ),
        reproduce: renderVitestReproduceCommand(classification.owner),
        raw: rawArtifacts,
      });
      continue;
    }
    for (const assertion of assertions) {
      if (assertion.status === "skipped") {
        continue;
      }
      if (selection && !selection.matches(ownerPath, assertion.title ?? "")) {
        continue;
      }
      const classification = classifyVitestCase(
        ownerPath,
        assertion.title ?? "",
        phaseLabel,
      );
      owners.add(classification.owner);
      counts.tests += 1;
      addCoverageCount(counts, classification.coverage);
      if (assertion.status === "passed") {
        inventory.push(
          createInventoryItem({
            coverage: classification.coverage,
            phase: classification.phase,
            id: classification.id,
            owner: classification.owner,
            name: assertion.title ?? "(missing title)",
          }),
        );
        continue;
      }
      counts.failed += 1;
      addCoverageFailureCount(counts, classification.coverage);
      const failureMessages = Array.isArray(assertion.failureMessages)
        ? assertion.failureMessages
        : [];
      const failureMessage = failureMessages[0] ?? "";
      const sidecarFailure = vitestFailureDetailsEntry(
        failureDetailsIndex,
        classification.owner,
        assertion.title ?? "(missing title)",
      );
      const sidecarMessage = sidecarFailure?.message ?? "";
      dossiers.push({
        coverage: classification.coverage,
        phase: classification.phase,
        id: classification.id,
        runner: "vitest",
        package_or_file: classification.owner,
        symbol_or_title: assertion.title ?? "(missing title)",
        message: summarizeVitestFailureMessage({
          fallback: `${assertion.title ?? "vitest assertion"} failed`,
          failureMessage,
          failureMessages,
          ownerPath: classification.owner,
          sidecarMessage,
          title: assertion.title ?? "(missing title)",
        }),
        diagnostic_tags: mergeVitestDiagnosticTags(
          vitestDiagnosticTags([...failureMessages, sidecarMessage]),
          sidecarFailure?.diagnostic_tags ?? [],
          sidecarFailure ? ["vitest_failure_sidecar"] : [],
        ),
        reproduce: renderVitestReproduceCommand(
          classification.owner,
          (assertion.title ?? "").trim(),
        ),
        raw: rawArtifacts,
      });
    }
  }

  if (
    counts.tests === 0 &&
    dossiers.length === 0 &&
    optionalEnv("CARTULARY_VITEST_ALLOW_EMPTY_SELECTION") !== "1"
  ) {
    dossiers.push({
      coverage: "unmapped",
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "vitest",
      package_or_file: "(vitest selection)",
      symbol_or_title: "(vitest selection)",
      message: "phase matched zero tests",
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
      raw: rawArtifacts,
    });
    counts.failed += 1;
    addCoverageFailureCount(counts, "unmapped");
  }

  counts.packages = owners.size;

  return {
    report,
    counts,
    owners: Array.from(owners).sort(),
    inventory,
    dossiers,
  };
}

function selectVitestManifestEntries(phase, coverage, executionDependency) {
  return selectVitestManifestEntriesAdapter(repoRoot, {
    phase,
    coverage,
    executionDependency,
  });
}

function collectVitestManifestEntries(coverage, executionDependency) {
  return collectVitestManifestEntriesAdapter(repoRoot, {
    coverage,
    executionDependency,
  });
}

export function handleVitestPhase({ manifestAware }) {
  const context = createBasePhaseContext("vitest");
  const reportFile = requiredEnv("CARTULARY_PHASE_RUNNER_LOG");
  const stderrLog = optionalEnv("CARTULARY_PHASE_STDERR_LOG");
  const stdoutLog = optionalEnv("CARTULARY_PHASE_STDOUT_LOG");
  const watchdogLog = optionalEnv("CARTULARY_PHASE_WATCHDOG_LOG");
  const failureDetailsLog = optionalEnv(
    "CARTULARY_PHASE_VITEST_FAILURE_DETAILS",
  );
  removeEmptyArtifact(stderrLog);
  removeEmptyArtifact(stdoutLog);

  if (!existsSync(reportFile)) {
    const counts = createCounts();
    counts.failed += 1;
    counts.non_test += 1;
    counts.non_test_failed += 1;
    const message = existsSync(watchdogLog)
      ? "vitest watchdog timed out before runner.json was written"
      : "vitest runner.json was not written";
    const dossier = {
      failure_class: "artifact",
      coverage: "non_test",
      phase: inferPhaseFromText(context.label),
      id: "",
      runner: "vitest",
      package_or_file: "(vitest runner)",
      symbol_or_title: "(runner.json)",
      message,
      reproduce: context.command,
      raw: renderRawList([watchdogLog, stdoutLog, stderrLog]),
    };
    writePhaseArtifacts(context, {
      status: "fail",
      phase: inferPhaseFromText(context.label),
      counts,
      owners: [],
      inventory: [],
      dossiers: [dossier],
      artifacts: {
        runner_json: reportFile,
        stdout_log: existsSync(stdoutLog) ? stdoutLog : "",
        stderr_log: existsSync(stderrLog) ? stderrLog : "",
        watchdog_json: existsSync(watchdogLog) ? watchdogLog : "",
        vitest_failure_details_json: existsSync(failureDetailsLog)
          ? failureDetailsLog
          : "",
      },
    });
    if (showPhaseDetailOutput(context)) {
      printBlock(`failure: ${context.label}`, dossier);
    }
    return 1;
  }

  if (context.countingMode === "none") {
    writePhaseArtifacts(context, {
      status: "pass",
      phase: inferPhaseFromText(context.label),
      counts: createCounts(),
      owners: [],
      inventory: [],
      dossiers: [],
      artifacts: {
        runner_json: reportFile,
        stdout_log: existsSync(stdoutLog) ? stdoutLog : "",
        stderr_log: existsSync(stderrLog) ? stderrLog : "",
        watchdog_json: existsSync(watchdogLog) ? watchdogLog : "",
        vitest_failure_details_json: existsSync(failureDetailsLog)
          ? failureDetailsLog
          : "",
      },
    });
    return 0;
  }

  const summary = summarizeVitestRun(
    reportFile,
    context.label,
    createVitestSelection({ manifestAware }),
    [reportFile, failureDetailsLog, stdoutLog, stderrLog],
    failureDetailsLog,
  );
  const selectedSlicePassed =
    summary.dossiers.length === 0 &&
    (context.exitStatus === 0 || optionalEnv("CARTULARY_REPORT_SLICE") === "1");

  return finalizeManifestAwareRunnerPhase(context, {
    manifestAware,
    runner: "vitest",
    summary,
    selectedSlicePassed,
    artifacts: {
      runner_json: reportFile,
      stdout_log: existsSync(stdoutLog) ? stdoutLog : "",
      stderr_log: existsSync(stderrLog) ? stderrLog : "",
      watchdog_json: existsSync(watchdogLog) ? watchdogLog : "",
      vitest_failure_details_json: existsSync(failureDetailsLog)
        ? failureDetailsLog
        : "",
    },
    manifestMismatchArtifacts: () => ({
      raw: relToRepo(reportFile),
    }),
    manifestMismatchDetailFields: (manifestMismatch) => ({
      raw: manifestMismatch.raw,
    }),
    failureDetailFields: (dossier) => ({
      ...dossier,
      raw: renderRawList([reportFile, failureDetailsLog, stdoutLog, stderrLog]),
    }),
  });
}

function isForbiddenFile(file, phase) {
  if (!phase) {
    return false;
  }
  const files = loadManifestIndex().forbiddenFilesByPhase.get(phase);
  return files ? files.has(file) : false;
}

function loadFrontendVitestIndex() {
  return loadFrontendVitestIndexAdapter(repoRoot);
}

function selectedPlaywrightEntriesFromReport(reportFile, scope) {
  return selectedPlaywrightEntriesFromReportAdapter(repoRoot, reportFile, scope);
}

function renderRawList(paths) {
  return paths
    .filter((entry) => entry && existsSync(entry))
    .map((entry) => relToRepo(entry))
    .join(";");
}

function escapeSingleQuotes(value) {
  return value.replaceAll("'", "'\"'\"'");
}
