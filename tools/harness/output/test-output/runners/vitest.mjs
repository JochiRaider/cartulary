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

function classifyVitestCase(ownerPath, title) {
  const manifestFile = vitestOwnerToSelectionFile(ownerPath);
  const authoritative = loadManifestIndex().authoritativeVitest.get(
    `${manifestFile}::${title}`,
  );
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
  return 1;
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
  stepLabel,
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
        stepLabel,
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
        step: classification.step,
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
        stepLabel,
      );
      owners.add(classification.owner);
      counts.tests += 1;
      addCoverageCount(counts, classification.coverage);
      if (assertion.status === "passed") {
        inventory.push(
          createInventoryItem({
            coverage: classification.coverage,
            step: classification.step,
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
        step: classification.step,
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

  if (!existsSync(reportFile)) {
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
      ? `vitest interrupted${normalizedInterruptSignal ? ` by ${normalizedInterruptSignal}` : ""} before runner.json was written`
      : existsSync(watchdogLog)
        ? "vitest watchdog timed out before runner.json was written"
        : "vitest runner.json was not written";
    const dossier = {
      failure_class: interrupted ? "interrupted" : "artifact",
      failure_reason: interrupted ? "cancelled_or_interrupted" : undefined,
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
    return 1;
  }

  if (context.countingMode === "none") {
    writeStepArtifacts(context, {
      status: "pass",
      step: catalogOwnerFromEnvironment(),
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
    createVitestSelection({ catalogAware }),
    [reportFile, failureDetailsLog, stdoutLog, stderrLog],
    failureDetailsLog,
  );
  const selectedSlicePassed =
    summary.dossiers.length === 0 &&
    (context.exitStatus === 0 || optionalEnv("CARTULARY_REPORT_SLICE") === "1");

  return finalizeManifestAwareRunnerStep(context, {
    catalogAware,
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
