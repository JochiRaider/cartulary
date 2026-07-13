#!/usr/bin/env node
import { repoRoot } from "../../../contract/index.mjs";

import {
  existsSync,
  readFileSync,
  rmSync,
  statSync,
} from "node:fs";
import path from "node:path";
import { validateSchemaSync } from "../../../contract/harness-contract.mjs";
import {
  loadManifestIndex as loadManifestIndexAdapter,
  packageMatchesPattern,
  selectGoManifestEntries as selectGoManifestEntriesAdapter,
} from "../phase-manifest-adapter.mjs";
import { verboseOutput } from "../../tool-output.mjs";
import {
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

function toRepoRelativePackage(importPath) {
  const modulePath = resolveGoModulePath();
  if (importPath === modulePath) {
    return "./";
  }
  if (importPath.startsWith(`${modulePath}/`)) {
    return `./${importPath.slice(modulePath.length + 1)}`;
  }
  return importPath;
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

function splitLogLines(file) {
  if (!file || !existsSync(file)) {
    return [];
  }
  return readFileSync(file, "utf8").split(/\r?\n/);
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

function accountingOverrideClassification(owner, phase = "") {
  const override = optionalEnv("CARTULARY_ACCOUNTING_COVERAGE");
  if (override === "") {
    return null;
  }
  return {
    coverage: normalizeTestCoverage(override),
    phase,
    id: "",
    owner,
  };
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

function firstGoActionableLine(lines) {
  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    if (
      line.startsWith("=== RUN") ||
      line.startsWith("=== PAUSE") ||
      line.startsWith("=== CONT") ||
      line.startsWith("=== NAME") ||
      line.startsWith("--- PASS") ||
      line.startsWith("--- FAIL") ||
      line.startsWith("--- SKIP") ||
      line === "PASS" ||
      line === "FAIL" ||
      /^ok\s/.test(line) ||
      /^\?\s/.test(line)
    ) {
      continue;
    }
    return line;
  }
  return "";
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

function readGoEvents(logFile) {
  const events = [];
  for (const rawLine of splitLogLines(logFile)) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    events.push(JSON.parse(line));
  }
  return events;
}

function classifyGoTest(importPath, testName, phaseLabel) {
  const manifestIndex = loadManifestIndex();
  const authoritative = manifestIndex.authoritativeGo.get(
    `${importPath}::${testName}`,
  );
  const owner = toRepoRelativePackage(importPath);
  if (authoritative) {
    return {
      coverage: "authoritative",
      phase: authoritative.phase,
      id: authoritative.id,
      owner: authoritative.package,
    };
  }

  const inferredPhase =
    inferPhaseFromText(testName) || inferPhaseFromText(phaseLabel);
  const override = accountingOverrideClassification(owner, inferredPhase);
  if (override) {
    return override;
  }
  const support =
    /^TestSupportPhase\d+_/.test(testName) ||
    /ProcessSmoke/.test(testName) ||
    /\bsupport\b/i.test(phaseLabel) ||
    /\bsmoke\b/i.test(phaseLabel);
  if (support) {
    return {
      coverage: "support",
      phase: inferredPhase,
      id: "",
      owner,
    };
  }

  return (
    accountingManifestClassification(
      "go_tests",
      owner,
      testName,
      inferredPhase,
    ) ?? {
      coverage: "unmapped",
      phase: inferredPhase,
      id: "",
      owner,
    }
  );
}

function classifyGoPackageFailure(importPath, phaseLabel) {
  const owner = toRepoRelativePackage(importPath);
  const manifestPhase = optionalEnv("CARTULARY_MANIFEST_PHASE");
  const manifestCoverage = optionalEnv("CARTULARY_MANIFEST_COVERAGE");
  if (manifestPhase !== "" && manifestCoverage !== "") {
    return {
      coverage: normalizeTestCoverage(
        manifestCoverage === "supplemental" ? "support" : manifestCoverage,
      ),
      phase: manifestPhase,
      id: "",
      owner,
    };
  }

  const inferredPhase = inferPhaseFromText(phaseLabel);
  const override = accountingOverrideClassification(owner, inferredPhase);
  if (override) {
    return override;
  }
  const support =
    /\bsupport\b/i.test(phaseLabel) || /\bsmoke\b/i.test(phaseLabel);
  if (support) {
    return {
      coverage: "support",
      phase: inferredPhase,
      id: "",
      owner,
    };
  }
  return (
    accountingManifestClassification(
      "go_packages",
      owner,
      "",
      inferredPhase,
    ) ?? {
      coverage: "unmapped",
      phase: inferredPhase,
      id: "",
      owner,
    }
  );
}

function createGoSelection({ manifestAware }) {
  const packagePatterns = optionalLines("CARTULARY_GO_PACKAGE_PATTERNS");
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";

  if (manifestAware && reportSlice) {
    const phase = requiredEnv("CARTULARY_MANIFEST_PHASE");
    const section = requiredEnv("CARTULARY_MANIFEST_SECTION");
    const coverage = requiredEnv("CARTULARY_MANIFEST_COVERAGE");
    const executionDependency = optionalEnv(
      "CARTULARY_MANIFEST_EXECUTION_DEPENDENCY",
    );
    const executionFamily = optionalEnv("CARTULARY_EXECUTION_FAMILY");
    const entries = selectGoManifestEntries(
      phase,
      section,
      coverage,
      executionDependency,
      executionFamily,
      packagePatterns,
    );
    const selectedTests = new Set();
    const selectedPackages = new Set();
    for (const entry of entries) {
      const symbols =
        entry.symbol !== undefined ? [entry.symbol] : entry.symbols;
      selectedPackages.add(toGoImportPath(entry.package));
      for (const symbol of symbols) {
        selectedTests.add(`${toGoImportPath(entry.package)}::${symbol}`);
      }
    }
    return {
      matchesTest(importPath, testName) {
        return selectedTests.has(`${importPath}::${testName}`);
      },
      matchesPackage(importPath) {
        return selectedPackages.has(importPath);
      },
    };
  }

  const testRegexSource = optionalEnv("CARTULARY_GO_TEST_REGEX");
  if (testRegexSource === "" && packagePatterns.length === 0) {
    return null;
  }
  const testRegex = testRegexSource === "" ? null : new RegExp(testRegexSource);
  return {
    matchesTest(importPath, testName) {
      if (
        packagePatterns.length > 0 &&
        !packagePatterns.some((pattern) =>
          packageMatchesPattern(toRepoRelativePackage(importPath), pattern),
        )
      ) {
        return false;
      }
      return testRegex === null ? true : testRegex.test(testName);
    },
    matchesPackage(importPath) {
      if (packagePatterns.length === 0) {
        return true;
      }
      return packagePatterns.some((pattern) =>
        packageMatchesPattern(toRepoRelativePackage(importPath), pattern),
      );
    },
  };
}

function summarizeGoRun(logFile, phaseLabel, exitStatus, selection = null) {
  const events = readGoEvents(logFile);
  const topLevel = new Map();
  const packageOutputs = new Map();
  const packageFailures = new Map();

  for (const entry of events) {
    const pkg = typeof entry.Package === "string" ? entry.Package : "";
    const testName = typeof entry.Test === "string" ? entry.Test : "";
    if (pkg !== "" && typeof entry.Output === "string") {
      if (testName === "" || testName.includes("/")) {
        if (!packageOutputs.has(pkg)) {
          packageOutputs.set(pkg, []);
        }
        packageOutputs.get(pkg).push(entry.Output);
      }
    }
    if (testName === "" || testName.includes("/")) {
      if (pkg !== "" && entry.Action === "fail") {
        packageFailures.set(pkg, true);
      }
      continue;
    }
    if (
      !testName.startsWith("Test") &&
      !testName.startsWith("Benchmark") &&
      !testName.startsWith("Fuzz")
    ) {
      continue;
    }
    const key = `${pkg}::${testName}`;
    if (!topLevel.has(key)) {
      topLevel.set(key, {
        package: pkg,
        test: testName,
        status: "",
        outputs: [],
      });
    }
    const current = topLevel.get(key);
    if (typeof entry.Output === "string") {
      current.outputs.push(entry.Output);
    }
    if (entry.Action === "run") {
      if (current.status === "") {
        current.status = "run";
      }
      continue;
    }
    if (["pass", "fail", "skip"].includes(entry.Action)) {
      current.status = entry.Action;
    }
  }

  const cases = [];
  const dossiers = [];
  const owners = new Set();
  const counts = createCounts();
  let passedCount = 0;
  let skippedCount = 0;
  let incompleteCount = 0;

  for (const testCase of topLevel.values()) {
    if (selection && !selection.matchesTest(testCase.package, testCase.test)) {
      continue;
    }
    const classification = classifyGoTest(
      testCase.package,
      testCase.test,
      phaseLabel,
    );
    const owner = classification.owner;
    owners.add(owner);
    if (testCase.status !== "skip") {
      counts.tests += 1;
      addCoverageCount(counts, classification.coverage);
    }
    if (testCase.status === "pass") {
      passedCount += 1;
      cases.push({
        status: "pass",
        classification,
        owner,
        name: testCase.test,
      });
      continue;
    }
    if (testCase.status === "skip") {
      skippedCount += 1;
      continue;
    }
    if (testCase.status === "fail") {
      counts.failed += 1;
      addCoverageFailureCount(counts, classification.coverage);
      dossiers.push({
        coverage: classification.coverage,
        phase: classification.phase,
        id: classification.id,
        runner: "go_test",
        package_or_file: owner,
        symbol_or_title: testCase.test,
        message:
          firstGoActionableLine(testCase.outputs) || `${testCase.test} failed`,
        reproduce: `go test ${owner} -run '^${escapeRegex(testCase.test)}$'`,
        raw: relToRepo(logFile),
      });
      continue;
    }
    incompleteCount += 1;
  }

  for (const pkg of packageFailures.keys()) {
    if (selection && !selection.matchesPackage(pkg)) {
      continue;
    }
    const classification = classifyGoPackageFailure(pkg, phaseLabel);
    const owner = classification.owner;
    owners.add(owner);
    if (
      [...topLevel.values()].some(
        (entry) => entry.package === pkg && entry.status === "fail",
      )
    ) {
      continue;
    }
    counts.failed += 1;
    addCoverageFailureCount(counts, classification.coverage);
    dossiers.push({
      coverage: classification.coverage,
      phase: classification.phase,
      id: "",
      runner: "go_test",
      package_or_file: owner,
      symbol_or_title: "(package setup)",
      message:
        firstGoActionableLine(packageOutputs.get(pkg) ?? []) ||
        `package ${owner} failed before a top-level test was attributed`,
      reproduce: `go test ${owner}`,
      raw: relToRepo(logFile),
    });
  }

  if (
    exitStatus === 0 &&
    (passedCount === 0 || skippedCount > 0 || incompleteCount > 0)
  ) {
    const coverageOverride = optionalEnv("CARTULARY_ACCOUNTING_COVERAGE");
    const coverage =
      coverageOverride !== ""
        ? normalizeTestCoverage(coverageOverride)
        : /\bsupport\b/i.test(phaseLabel)
          ? "support"
          : "unmapped";
    const message =
      passedCount === 0 && skippedCount === 0 && incompleteCount === 0
        ? coverage === "support"
          ? "support phase matched zero tests"
          : "phase matched zero tests"
        : `go test inventory requires top-level pass: skipped=${skippedCount} incomplete=${incompleteCount}`;
    dossiers.push({
      coverage,
      phase: inferPhaseFromText(phaseLabel),
      id: "",
      runner: "go_test",
      package_or_file: "(phase selection)",
      symbol_or_title: "(top-level selection)",
      message,
      reproduce: requiredEnv("CARTULARY_PHASE_COMMAND"),
      raw: relToRepo(logFile),
    });
    counts.failed += 1;
    addCoverageFailureCount(counts, coverage);
  }

  counts.packages = owners.size;

  return {
    counts,
    owners: Array.from(owners).sort(),
    inventory: cases
      .filter(
        (entry) =>
          entry.classification.coverage !== "unmapped" ||
          entry.classification.id,
      )
      .map((entry) =>
        createInventoryItem({
          coverage: entry.classification.coverage,
          phase: entry.classification.phase,
          id: entry.classification.id,
          owner: entry.owner,
          name: entry.name,
        }),
      ),
    dossiers,
  };
}

function selectGoManifestEntries(
  phase,
  section,
  coverage,
  executionDependency,
  executionFamily,
  packagePatterns,
) {
  return selectGoManifestEntriesAdapter(repoRoot, {
    phase,
    section,
    coverage,
    executionDependency,
    executionFamily,
    packagePatterns,
  });
}

function evaluateGoManifest(summary) {
  const phase = requiredEnv("CARTULARY_MANIFEST_PHASE");
  const section = requiredEnv("CARTULARY_MANIFEST_SECTION");
  const coverage = requiredEnv("CARTULARY_MANIFEST_COVERAGE");
  const executionDependency = optionalEnv(
    "CARTULARY_MANIFEST_EXECUTION_DEPENDENCY",
  );
  const executionFamily = optionalEnv("CARTULARY_EXECUTION_FAMILY");
  const packagePatterns = optionalLines("CARTULARY_GO_PACKAGE_PATTERNS");
  const selectedIDs = new Set(
    optionalLines("CARTULARY_MANIFEST_SELECTED_IDS"),
  );
  const entries = selectGoManifestEntries(
    phase,
    section,
    coverage,
    executionDependency,
    executionFamily,
    packagePatterns,
  ).filter((entry) => selectedIDs.size === 0 || selectedIDs.has(entry.id));
  const expectedByID = new Map();
  for (const entry of entries) {
    const symbols = entry.symbol !== undefined ? [entry.symbol] : entry.symbols;
    expectedByID.set(entry.id, {
      id: entry.id,
      symbols,
    });
  }
  const passedTests = new Set(
    summary.inventory.map(
      (item) =>
        `${toGoImportPath(item.package_or_file)}::${item.symbol_or_title}`,
    ),
  );
  const missingIDs = [];
  for (const entry of entries) {
    const symbols = entry.symbol !== undefined ? [entry.symbol] : entry.symbols;
    const missing = symbols.some(
      (symbol) =>
        !passedTests.has(`${toGoImportPath(entry.package)}::${symbol}`),
    );
    if (missing) {
      missingIDs.push(entry.id);
    }
  }
  const expectedIDs = new Set(entries.map((entry) => entry.id));
  const unexpectedIDs = [];
  for (const item of summary.inventory) {
    if (
      item.coverage !== "authoritative" ||
      item.id === "" ||
      expectedIDs.has(item.id)
    ) {
      continue;
    }
    unexpectedIDs.push(item.id);
  }
  return {
    phase,
    missingIDs: [...new Set(missingIDs)].sort(),
    unexpectedIDs: [...new Set(unexpectedIDs)].sort(),
    forbiddenIDFiles: Array.from(
      loadManifestIndex().forbiddenFilesByPhase.get(phase) ?? [],
    ).sort(),
  };
}

export function handleGoPhase({ manifestAware }) {
  const context = createBasePhaseContext("go_test");
  const runnerLog = requiredEnv("CARTULARY_PHASE_RUNNER_LOG");
  const stderrLog = optionalEnv("CARTULARY_PHASE_STDERR_LOG");
  removeEmptyArtifact(stderrLog);

  const summary = summarizeGoRun(
    runnerLog,
    context.label,
    context.exitStatus,
    createGoSelection({ manifestAware }),
  );
  let status =
    context.exitStatus === 0 && summary.dossiers.length === 0 ? "pass" : "fail";
  let manifestMismatch = null;
  let manifestSummary = null;

  if (
    manifestAware &&
    context.exitStatus === 0 &&
    summary.dossiers.length === 0
  ) {
    const verification = evaluateGoManifest(summary);
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
        raw: relToRepo(runnerLog),
      };
    }
  }

  writePhaseArtifacts(context, {
    status,
    phase: inferPhaseFromText(context.label),
    counts: summary.counts,
    owners: summary.owners,
    inventory: summary.inventory,
    dossiers: summary.dossiers,
    manifestSummary,
    manifestMismatch,
    artifacts: {
      runner_jsonl: runnerLog,
      stderr_log: existsSync(stderrLog) ? stderrLog : "",
    },
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
        raw: manifestMismatch.raw,
      });
    }
    return 1;
  }

  if (showPhaseDetailOutput(context)) {
    for (const dossier of summary.dossiers) {
      printBlock(`failure: ${context.label}`, {
        ...dossier,
        raw: renderRawList([runnerLog, stderrLog]),
      });
    }
  }
  return 1;
}

function renderRawList(paths) {
  return paths
    .filter((entry) => entry && existsSync(entry))
    .map((entry) => relToRepo(entry))
    .join(";");
}

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
