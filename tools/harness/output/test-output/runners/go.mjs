#!/usr/bin/env node
import { repoRoot } from "../../../contract/index.mjs";

import {
  existsSync,
  readFileSync,
  rmSync,
  statSync,
} from "node:fs";
import path from "node:path";
import {
  loadManifestIndex as loadManifestIndexAdapter,
  packageMatchesPattern,
  selectGoManifestEntries as selectGoManifestEntriesAdapter,
} from "../catalog-manifest-adapter.mjs";
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

function accountingOverrideClassification(owner, step = "") {
  const override = optionalEnv("CARTULARY_ACCOUNTING_COVERAGE");
  if (override === "") {
    return null;
  }
  return {
    coverage: normalizeTestCoverage(override),
    step,
    id: "",
    owner,
  };
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

function classifyGoTest(importPath, testName, stepLabel) {
  const manifestIndex = loadManifestIndex();
  const authoritative = manifestIndex.authoritativeGo.get(
    `${importPath}::${testName}`,
  );
  const owner = toRepoRelativePackage(importPath);
  if (authoritative) {
    return {
      coverage: "authoritative",
      step: authoritative.step,
      id: authoritative.id,
      owner: authoritative.package,
    };
  }

  const inferredStep =
    inferStepFromText(testName) || inferStepFromText(stepLabel);
  const override = accountingOverrideClassification(owner, inferredStep);
  if (override) {
    return override;
  }
  const support =
    /^TestSupportStep\d+_/.test(testName) ||
    /ProcessSmoke/.test(testName) ||
    /\bsupport\b/i.test(stepLabel) ||
    /\bsmoke\b/i.test(stepLabel);
  if (support) {
    return {
      coverage: "support",
      step: inferredStep,
      id: "",
      owner,
    };
  }

  return {
    coverage: "unmapped",
    step: inferredStep,
    id: "",
    owner,
  };
}

function classifyGoPackageFailure(importPath, stepLabel) {
  const owner = toRepoRelativePackage(importPath);
  const catalogStep = optionalEnv("CARTULARY_CATALOG_OWNER_ID");
  const manifestCoverage = optionalEnv("CARTULARY_MANIFEST_COVERAGE");
  if (catalogStep !== "" && manifestCoverage !== "") {
    return {
      coverage: normalizeTestCoverage(
        manifestCoverage === "supplemental" ? "support" : manifestCoverage,
      ),
      step: catalogStep,
      id: "",
      owner,
    };
  }

  const inferredStep = inferStepFromText(stepLabel);
  const override = accountingOverrideClassification(owner, inferredStep);
  if (override) {
    return override;
  }
  const support =
    /\bsupport\b/i.test(stepLabel) || /\bsmoke\b/i.test(stepLabel);
  if (support) {
    return {
      coverage: "support",
      step: inferredStep,
      id: "",
      owner,
    };
  }
  return {
    coverage: "unmapped",
    step: inferredStep,
    id: "",
    owner,
  };
}

function createGoSelection({ catalogAware }) {
  const packagePatterns = optionalLines("CARTULARY_GO_PACKAGE_PATTERNS");
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";

  if (catalogAware && reportSlice) {
    const step = requiredEnv("CARTULARY_CATALOG_OWNER_ID");
    const section = requiredEnv("CARTULARY_MANIFEST_SECTION");
    const coverage = requiredEnv("CARTULARY_MANIFEST_COVERAGE");
    const executionDependency = optionalEnv(
      "CARTULARY_MANIFEST_EXECUTION_DEPENDENCY",
    );
    const executionFamily = optionalEnv("CARTULARY_EXECUTION_FAMILY");
    const entries = selectGoManifestEntries(
      step,
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

function summarizeGoRun(logFile, stepLabel, exitStatus, selection = null) {
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
      stepLabel,
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
        step: classification.step,
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
    const classification = classifyGoPackageFailure(pkg, stepLabel);
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
      step: classification.step,
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
        : /\bsupport\b/i.test(stepLabel)
          ? "support"
          : "unmapped";
    const message =
      passedCount === 0 && skippedCount === 0 && incompleteCount === 0
        ? coverage === "support"
          ? "support step matched zero tests"
          : "step matched zero tests"
        : `go test inventory requires top-level pass: skipped=${skippedCount} incomplete=${incompleteCount}`;
    dossiers.push({
      coverage,
      step: inferStepFromText(stepLabel),
      id: "",
      runner: "go_test",
      package_or_file: "(step selection)",
      symbol_or_title: "(top-level selection)",
      message,
      reproduce: requiredEnv("CARTULARY_STEP_COMMAND"),
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
          step: entry.classification.step,
          id: entry.classification.id,
          owner: entry.owner,
          name: entry.name,
        }),
      ),
    dossiers,
  };
}

function selectGoManifestEntries(
  step,
  section,
  coverage,
  executionDependency,
  executionFamily,
  packagePatterns,
) {
  return selectGoManifestEntriesAdapter(repoRoot, {
    step,
    section,
    coverage,
    executionDependency,
    executionFamily,
    packagePatterns,
  });
}

function evaluateGoManifest(summary) {
  const step = requiredEnv("CARTULARY_CATALOG_OWNER_ID");
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
    step,
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
    step,
    missingIDs: [...new Set(missingIDs)].sort(),
    unexpectedIDs: [...new Set(unexpectedIDs)].sort(),
    forbiddenIDFiles: Array.from(
      loadManifestIndex().forbiddenFilesByStep.get(step) ?? [],
    ).sort(),
  };
}

export function handleGoStep({ catalogAware }) {
  const context = createBaseStepContext("go_test");
  const runnerLog = requiredEnv("CARTULARY_STEP_RUNNER_LOG");
  const stderrLog = optionalEnv("CARTULARY_STEP_STDERR_LOG");
  removeEmptyArtifact(stderrLog);

  const summary = summarizeGoRun(
    runnerLog,
    context.label,
    context.exitStatus,
    createGoSelection({ catalogAware }),
  );
  let status =
    context.exitStatus === 0 && summary.dossiers.length === 0 ? "pass" : "fail";
  let manifestMismatch = null;
  let manifestSummary = null;

  if (
    catalogAware &&
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

  writeStepArtifacts(context, {
    status,
    step: inferStepFromText(context.label),
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
    if (showStepDetailOutput(context)) {
      printBlock(`manifest mismatch: ${context.label}`, {
        missing_ids: renderList(manifestMismatch.missing_ids),
        unexpected_ids: renderList(manifestMismatch.unexpected_ids),
        forbidden_id_files: renderList(manifestMismatch.forbidden_id_files),
        raw: manifestMismatch.raw,
      });
    }
    return 1;
  }

  if (showStepDetailOutput(context)) {
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
