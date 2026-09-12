#!/usr/bin/env node
import { publicExitCodeForFailure, publicExitCodeForFailures, repoRoot } from "../../../contract/index.mjs";
import { diagnosticFailure, readVitestJSON, reconcileVitestReport } from "../../../diagnostics/vitest-failure-details.mjs";
import { readCommandFailure } from "../../../runtime/command-failure.mjs";

import {
  existsSync,
  readFileSync,
  rmSync,
  statSync,
} from "node:fs";
import path from "node:path";
import { validateSchemaSync } from "../../../contract/harness-contract.mjs";
import {
  collectVitestManifestEntries as collectVitestManifestEntriesAdapter,
  loadManifestIndex as loadManifestIndexAdapter,
  playwrightEntryTitles,
  selectManifestEntries,
  selectPlaywrightEntries,
  selectVitestManifestEntries as selectVitestManifestEntriesAdapter,
  vitestEntryTitles,
} from "../catalog-manifest-adapter.mjs";
import { selectedPlaywrightEntriesFromReport as selectedPlaywrightEntriesFromReportAdapter } from "../playwright-artifacts.mjs";
import { verboseOutput } from "../../tool-output.mjs";
import {
  testCoverageBucketSet,
  testCoverageBuckets,
  vitestFailureDetailsSchemaID,
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

function catalogOwnerFromEnvironment() {
  return optionalEnv("CARTULARY_CATALOG_OWNER_ID");
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

function classifyVitestCase(ownerPath, title, leafTitle = title) {
  const manifestFile = vitestOwnerToSelectionFile(ownerPath);
  const authoritative = loadManifestIndex().authoritativeVitest.get(
    `${manifestFile}::${title}`,
  ) ?? loadManifestIndex().authoritativeVitest.get(`${manifestFile}::${leafTitle}`);
  if (authoritative) {
    return {
      coverage: "authoritative",
      step: authoritative.step,
      id: authoritative.id,
      owner: ownerPath,
    };
  }
  return {
    coverage: "unmapped",
    step: catalogOwnerFromEnvironment(),
    id: "",
    owner: ownerPath,
  };
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
    forbiddenIDFiles: [],
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
        step: catalogOwnerFromEnvironment(),
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
    step: catalogOwnerFromEnvironment(),
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
  return publicExitCodeForFailures(stepDossiers, { exit_status: context.exitStatus }) || 1;
}

function createVitestSelection({ catalogAware }) {
  const reportSlice = optionalEnv("CARTULARY_REPORT_SLICE") === "1";

  if (catalogAware && reportSlice) {
    const { step, coverage, executionDependency } = readManifestScopeEnv();
    const entries = selectVitestManifestEntries(
      step,
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
          step,
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

function findVitestAuthoritativeFileEntry(ownerPath) {
  const manifestFile = vitestOwnerToSelectionFile(ownerPath);
  const entries = [...loadManifestIndex().authoritativeVitest.values()].filter(
    (entry) => normalizePath(entry.file) === manifestFile,
  );
  const ownerIDs = new Set(entries.map((entry) => entry.step));
  return ownerIDs.size === 1 ? entries[0] : null;
}

function classifyVitestFileFailure(ownerPath, _stepLabel, selection = null) {
  const selected = selection?.classifyFileFailure?.(ownerPath);
  if (selected) {
    return selected;
  }

  const authoritative = findVitestAuthoritativeFileEntry(ownerPath);
  if (authoritative) {
    return {
      coverage: "authoritative",
      step: authoritative.step,
      id: "",
      owner: ownerPath,
    };
  }

  return {
    coverage: "unmapped",
    step: catalogOwnerFromEnvironment(),
    id: "",
    owner: ownerPath,
  };
}

function summarizeVitestRun(
  reportFile,
  stepLabel,
  selection = null,
  rawFiles = [reportFile],
  failureDetailsFile = "",
) {
  const report = readVitestJSON(reportFile);
  const failureDetails = readVitestJSON(failureDetailsFile);
  validateSchemaSync(vitestFailureDetailsSchemaID, failureDetails);
  reconcileVitestReport(failureDetails, report, repoRoot);
  const owners = new Set();
  const inventory = [];
  const dossiers = [];
  const counts = createCounts();
  const rawArtifacts = renderRawList(rawFiles) || relToRepo(reportFile);

  for (const observation of failureDetails.observations) {
    const ownerPath = normalizeVitestOwnerPath(observation.owner_path);
    if (selection && !(observation.scope === "test"
      ? (selection.matches(ownerPath, observation.title) || selection.matches(ownerPath, observation.test_name))
      : selection.matchesFile(ownerPath))) continue;
    if (observation.status === "skipped") continue;
    const classification = observation.scope === "test"
      ? classifyVitestCase(ownerPath, observation.title, observation.test_name)
      : classifyVitestFileFailure(ownerPath, stepLabel, selection);
    owners.add(classification.owner);
    if (observation.scope === "test") {
      counts.tests += 1;
      addCoverageCount(counts, classification.coverage);
    }
    if (observation.status === "passed") {
      inventory.push(createInventoryItem({ coverage: classification.coverage, step: classification.step,
        id: classification.id, owner: classification.owner, name: observation.title }));
      continue;
    }
    counts.failed += 1;
    addCoverageFailureCount(counts, classification.coverage);
    const failures = failureDetails.failures.filter((failure) =>
      failure.project === observation.project && failure.owner_path === observation.owner_path && failure.title === observation.title);
    if (!failures.length) throw diagnosticFailure("Failed Vitest observation has no cause", "scheduler_accounting_error");
    for (const failure of failures) dossiers.push({
      coverage: classification.coverage, step: classification.step, id: classification.id,
      runner: "vitest", package_or_file: classification.owner, symbol_or_title: observation.title,
      failure_class: failure.failure_class, failure_reason: failure.failure_reason,
      message: `${failure.original_error.name}: ${failure.message}`,
      diagnostic_tags: [...failure.diagnostic_tags, "vitest_failure_sidecar"],
      reproduce: renderVitestReproduceCommand(classification.owner, observation.scope === "test" ? observation.title : ""),
      raw: rawArtifacts,
    });
  }

  if (
    counts.tests === 0 &&
    dossiers.length === 0 &&
    optionalEnv("CARTULARY_VITEST_ALLOW_EMPTY_SELECTION") !== "1"
  ) {
    dossiers.push({
      failure_class: "harness",
      failure_reason: "scheduler_accounting_error",
      coverage: "unmapped",
      step: catalogOwnerFromEnvironment(),
      id: "",
      runner: "vitest",
      package_or_file: "(vitest selection)",
      symbol_or_title: "(vitest selection)",
      message: "step matched zero tests",
      reproduce: requiredEnv("CARTULARY_STEP_COMMAND"),
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

function selectVitestManifestEntries(step, coverage, executionDependency) {
  return selectVitestManifestEntriesAdapter(repoRoot, {
    step,
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

export function handleVitestStep({ catalogAware }) {
  const context = createBaseStepContext("vitest");
  const reportFile = requiredEnv("CARTULARY_STEP_RUNNER_LOG");
  const stderrLog = optionalEnv("CARTULARY_STEP_STDERR_LOG");
  const stdoutLog = optionalEnv("CARTULARY_STEP_STDOUT_LOG");
  const watchdogLog = optionalEnv("CARTULARY_STEP_WATCHDOG_LOG");
  const interruptSignal = optionalEnv("CARTULARY_STEP_INTERRUPT_SIGNAL");
  const failureDetailsLog = optionalEnv(
    "CARTULARY_STEP_VITEST_FAILURE_DETAILS",
  );
  removeEmptyArtifact(stderrLog);
  removeEmptyArtifact(stdoutLog);

  if (!existsSync(reportFile) || !existsSync(failureDetailsLog) || interruptSignal !== "" || [130, 143].includes(context.exitStatus)) {
    const commandFailure = readCommandFailure(repoRoot);
    const interrupted =
      interruptSignal !== "" ||
      context.exitStatus === 130 ||
      context.exitStatus === 143;
    const normalizedInterruptSignal =
      interruptSignal ||
      (context.exitStatus === 130
        ? "SIGINT"
        : context.exitStatus === 143
          ? "SIGTERM"
          : "");
    const counts = createCounts();
    counts.failed += 1;
    counts.non_test += 1;
    counts.non_test_failed += 1;
    const message = interrupted
      ? `vitest interrupted${normalizedInterruptSignal ? ` by ${normalizedInterruptSignal}` : ""}${existsSync(reportFile) ? " after diagnostic publication" : " before runner.json was written"}`
      : existsSync(watchdogLog)
        ? "vitest watchdog timed out before runner.json was written"
        : commandFailure?.failure_reason === "scheduler_accounting_error"
          ? "Vitest invocation rejected contradictory diagnostic observations"
          : "required Vitest runner JSON or diagnostic sidecar was not written";
    const dossier = {
      failure_class: interrupted ? "interrupted" : existsSync(watchdogLog) ? "timing" : commandFailure?.failure_class ?? "artifact",
      failure_reason: interrupted ? "cancelled_or_interrupted" : existsSync(watchdogLog) ? "timeout_failure" : commandFailure?.failure_reason ?? "artifact_error",
      coverage: "non_test",
      step: catalogOwnerFromEnvironment(),
      id: "",
      runner: "vitest",
      package_or_file: "(vitest runner)",
      symbol_or_title: "(runner.json)",
      message,
      reproduce: context.command,
      raw: renderRawList([watchdogLog, stdoutLog, stderrLog]),
    };
    writeStepArtifacts(context, {
      status: "fail",
      step: catalogOwnerFromEnvironment(),
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
    if (showStepDetailOutput(context)) {
      printBlock(`failure: ${context.label}`, dossier);
    }
    return publicExitCodeForFailure(dossier, { exit_status: context.exitStatus });
  }

  let summary;
  try {
    summary = summarizeVitestRun(reportFile, context.label,
      createVitestSelection({ catalogAware }),
      [reportFile, failureDetailsLog, stdoutLog, stderrLog], failureDetailsLog);
  } catch (error) {
    const failure = error.failure_reason ? error : diagnosticFailure("Invalid required Vitest diagnostic evidence");
    const counts = createCounts();
    counts.failed = 1;
    counts.non_test = 1;
    counts.non_test_failed = 1;
    const dossier = { failure_class: failure.failure_class, failure_reason: failure.failure_reason,
      coverage: "non_test", step: catalogOwnerFromEnvironment(), id: "", runner: "vitest",
      package_or_file: "(vitest diagnostics)", symbol_or_title: "(vitest diagnostics)",
      message: failure.message, reproduce: context.command,
      raw: renderRawList([reportFile, failureDetailsLog, stdoutLog, stderrLog]) };
    writeStepArtifacts(context, { status: "fail", step: catalogOwnerFromEnvironment(), counts, owners: [], inventory: [], dossiers: [dossier],
      artifacts: { runner_json: reportFile, vitest_failure_details_json: failureDetailsLog } });
    if (showStepDetailOutput(context)) printBlock(`failure: ${context.label}`, dossier);
    return publicExitCodeForFailure(failure);
  }
  const projectedSummary =
    context.countingMode === "none"
      ? {
          ...summary,
          counts: createCounts(),
          owners: [],
          inventory: [],
        }
      : summary;
  const selectedSlicePassed =
    summary.dossiers.length === 0 &&
    (context.exitStatus === 0 || optionalEnv("CARTULARY_REPORT_SLICE") === "1");

  return finalizeManifestAwareRunnerStep(context, {
    catalogAware,
    runner: "vitest",
    summary: projectedSummary,
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
