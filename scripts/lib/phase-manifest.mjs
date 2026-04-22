import { readFileSync } from "node:fs";
import path from "node:path";

const sectionDefinitions = [
  ["unit", "U-"],
  ["integration", "I-"],
  ["e2e", "E-"],
];
const implementationTestingGuidePath = path.join(
  "docs",
  "guides",
  "cartulary_implementation_testing_guide.md",
);

const validCoverage = new Set(["authoritative", "supplemental"]);
const validGoSections = new Set(["unit", "integration", "e2e"]);
const validExecutionDependencies = new Set([
  "backend_unit",
  "backend_store",
  "backend_integration",
  "backend_process",
  "frontend_unit",
  "browser_functional",
  "browser_stateful",
  "browser_measurement",
]);

function goEntrySymbols(entry) {
  if (entry.symbol !== undefined && entry.symbols !== undefined) {
    throw new Error(`manifest entry ${entry.id} must declare symbol or symbols[], not both`);
  }
  if (entry.symbols !== undefined) {
    if (!Array.isArray(entry.symbols) || entry.symbols.length === 0) {
      throw new Error(`manifest entry ${entry.id} must declare a non-empty symbols[] array`);
    }
    for (const symbol of entry.symbols) {
      if (typeof symbol !== "string" || symbol.trim() === "") {
        throw new Error(`manifest entry ${entry.id} has an invalid symbol in symbols[]`);
      }
    }
    return entry.symbols;
  }
  if (typeof entry.symbol !== "string" || entry.symbol.trim() === "") {
    throw new Error(`manifest entry ${entry.id} is missing a non-empty symbol`);
  }
  return [entry.symbol];
}

function phaseNumberFromPhase(phase) {
  const match = /^phase(\d+)$/.exec(phase);
  if (!match) {
    throw new Error(`invalid phase name ${phase}; expected phase<number>`);
  }
  return match[1];
}

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`);
}

function phaseIDPatternSource(layerPrefix, phaseNumber, separator) {
  const normalizedLayerPrefix = layerPrefix.endsWith(separator)
    ? layerPrefix
    : `${layerPrefix}${separator}`;
  return `${escapeRegex(normalizedLayerPrefix)}${phaseNumber}${escapeRegex(
    separator,
  )}(?:[A-Z0-9]+${escapeRegex(separator)})*\\d{2}`;
}

function phaseIDRegex(layerPrefix, phaseNumber) {
  return new RegExp(`^${phaseIDPatternSource(layerPrefix, phaseNumber, "-")}$`);
}

function claimedPhaseIDRegex(phaseNumber, separator) {
  return new RegExp(
    String.raw`\b[UIE]${escapeRegex(separator)}${phaseNumber}${escapeRegex(
      separator,
    )}(?:[A-Z0-9]+${escapeRegex(separator)})*\d{2}\b`,
    "g",
  );
}

function validateExpectedIDs(expectedIDs, phaseNumber, manifestPath) {
  const seen = new Set();
  for (const id of expectedIDs) {
    if (typeof id !== "string" || id.trim() === "") {
      throw new Error(`manifest ${manifestPath} has an invalid expected_id: ${JSON.stringify(id)}`);
    }
    const layerPrefix = `${id[0] ?? ""}-`;
    if (!phaseIDRegex(layerPrefix, phaseNumber).test(id)) {
      throw new Error(`manifest ${manifestPath} has expected_id ${id} that does not belong to phase${phaseNumber}`);
    }
    if (seen.has(id)) {
      throw new Error(`manifest ${manifestPath} has duplicate expected_id ${id}`);
    }
    seen.add(id);
  }
}

function loadGuideExpectedIDs(root, phaseNumber) {
  const source = readFileSync(path.join(root, implementationTestingGuidePath), "utf8");
  const pattern = new RegExp(
    String.raw`^\|\s*([UIE]-${phaseNumber}(?:-[A-Z0-9]+)*-\d{2})\s*\|`,
  );
  const ids = new Set();
  for (const line of source.split(/\r?\n/)) {
    const match = pattern.exec(line);
    if (match?.[1]) {
      ids.add(match[1]);
    }
  }
  return Array.from(ids).sort();
}

function extractClaimedPhaseIDs(source, phaseNumber) {
  const hyphenMatches = source.match(claimedPhaseIDRegex(phaseNumber, "-")) ?? [];
  const underscoreMatches = source.match(claimedPhaseIDRegex(phaseNumber, "_")) ?? [];
  return new Set([
    ...hyphenMatches,
    ...underscoreMatches.map((value) => value.replaceAll("_", "-")),
  ]);
}

export function loadManifest(root, phase) {
  const manifestPath = path.join(root, "tools", `${phase}_test_map.json`);
  const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
  return { manifestPath, manifest };
}

export function collectEntries(manifest) {
  const entries = [];
  for (const [section] of sectionDefinitions) {
    for (const entry of manifest[section] ?? []) {
      entries.push({ ...entry, section });
    }
  }
  return entries;
}

export function validateManifest(root, phase) {
  const phaseNumber = phaseNumberFromPhase(phase);
  const { manifestPath, manifest } = loadManifest(root, phase);

  if (!Array.isArray(manifest.expected_ids) || manifest.expected_ids.length === 0) {
    throw new Error(`manifest ${manifestPath} must define a non-empty expected_ids array`);
  }
  validateExpectedIDs(manifest.expected_ids, phaseNumber, manifestPath);

  const entries = [];
  for (const [section, prefix] of sectionDefinitions) {
    for (const entry of manifest[section] ?? []) {
      if (typeof entry.id !== "string" || !entry.id.startsWith(prefix)) {
        throw new Error(`manifest entry in ${section} has invalid id: ${JSON.stringify(entry)}`);
      }
      if (!phaseIDRegex(prefix, phaseNumber).test(entry.id)) {
        throw new Error(`manifest entry ${entry.id} does not belong to ${phase}`);
      }
      if (typeof entry.coverage !== "string" || !validCoverage.has(entry.coverage)) {
        throw new Error(`manifest entry ${entry.id} must declare coverage=authoritative|supplemental`);
      }
      if (
        typeof entry.execution_dependency === "string" &&
        !validExecutionDependencies.has(entry.execution_dependency)
      ) {
        throw new Error(
          `manifest entry ${entry.id} has invalid execution_dependency ${entry.execution_dependency}`,
        );
      }
      if (entry.runner === "playwright") {
        if (typeof entry.title !== "string" || entry.title.trim() === "") {
          throw new Error(`manifest entry ${entry.id} is missing a non-empty title`);
        }
        if (typeof entry.file !== "string" || !entry.file.startsWith("apps/web/e2e/")) {
          throw new Error(`manifest entry ${entry.id} must point at an apps/web/e2e file`);
        }
      } else if (entry.runner === "vitest") {
        if (typeof entry.title !== "string" || entry.title.trim() === "") {
          throw new Error(`manifest entry ${entry.id} is missing a non-empty title`);
        }
        if (typeof entry.file !== "string" || !entry.file.startsWith("apps/web/")) {
          throw new Error(`manifest entry ${entry.id} must point at an apps/web file`);
        }
      } else if (entry.runner === "go_test") {
        goEntrySymbols(entry);
        if (typeof entry.package !== "string" || !entry.package.startsWith("./")) {
          throw new Error(`manifest entry ${entry.id} must declare a repo-relative Go package owner`);
        }
      } else {
        throw new Error(`manifest entry ${entry.id} must declare runner=go_test|playwright|vitest`);
      }
      if (typeof entry.file !== "string" || entry.file.trim() === "") {
        throw new Error(`manifest entry ${entry.id} must declare a file`);
      }
      if (typeof entry.evidence_layer !== "string" || entry.evidence_layer.trim() === "") {
        throw new Error(`manifest entry ${entry.id} must declare evidence_layer`);
      }
      entries.push({ ...entry, section });
    }
  }

  const ids = entries.map((entry) => entry.id);
  const uniqueIDs = new Set(ids);
  if (uniqueIDs.size !== ids.length) {
    throw new Error(`duplicate ids in ${manifestPath}`);
  }

  const authoritativeIDs = entries
    .filter((entry) => entry.coverage === "authoritative")
    .map((entry) => entry.id);
  const uniqueAuthoritativeIDs = new Set(authoritativeIDs);
  if (uniqueAuthoritativeIDs.size !== authoritativeIDs.length) {
    throw new Error(`duplicate authoritative ids in ${manifestPath}`);
  }

  const expected = manifest.expected_ids;
  const missing = expected.filter((id) => !uniqueAuthoritativeIDs.has(id));
  const unexpected = authoritativeIDs.filter((id) => !expected.includes(id));
  if (missing.length > 0 || unexpected.length > 0) {
    throw new Error(
      `${phase} manifest mismatch: missing=${missing.join(",") || "none"} unexpected=${unexpected.join(",") || "none"}`,
    );
  }

  const guideExpectedIDs = loadGuideExpectedIDs(root, phaseNumber);
  if (guideExpectedIDs.length > 0) {
    const guideMissing = guideExpectedIDs.filter((id) => !expected.includes(id));
    const guideUnexpected = expected.filter((id) => !guideExpectedIDs.includes(id));
    if (guideMissing.length > 0 || guideUnexpected.length > 0) {
      throw new Error(
        `${phase} guide mismatch: missing=${guideMissing.join(",") || "none"} unexpected=${guideUnexpected.join(",") || "none"}`,
      );
    }
  }

  for (const entry of entries) {
    const targetPath = path.join(root, entry.file);
    const source = readFileSync(targetPath, "utf8");
    const needles = entry.runner === "go_test" ? goEntrySymbols(entry) : [entry.title];
    for (const needle of needles) {
      if (!source.includes(needle)) {
        throw new Error(`manifest entry ${entry.id} not found in ${entry.file}: ${needle}`);
      }
    }
  }

  for (const target of manifest.forbidden_id_files ?? []) {
    const source = readFileSync(path.join(root, target), "utf8");
    const claimedIDs = extractClaimedPhaseIDs(source, phaseNumber);
    if (claimedIDs.size > 0) {
      throw new Error(
        `${target} must not claim ${phase} authoritative ids: ${Array.from(claimedIDs).sort().join(", ")}`,
      );
    }
  }
}

function exactRegex(values) {
  if (values.length === 0) {
    throw new Error("cannot build an exact regex from an empty value list");
  }
  const escaped = values.map((value) => value.replace(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`));
  if (escaped.length === 1) {
    return `^${escaped[0]}$`;
  }
  return `^(${escaped.join("|")})$`;
}

function alternationRegex(values) {
  if (values.length === 0) {
    throw new Error("cannot build a regex from an empty value list");
  }
  const escaped = values.map((value) => value.replace(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`));
  if (escaped.length === 1) {
    return escaped[0];
  }
  return `(${escaped.join("|")})`;
}

function packageMatchesPattern(pkg, pattern) {
  if (pattern.endsWith("/...")) {
    const prefix = pattern.slice(0, -4);
    return pkg === prefix || pkg.startsWith(`${prefix}/`);
  }
  return pkg === pattern;
}

function entryMatchesExecutionDependency(entry, executionDependency) {
  return executionDependency === "" || entry.execution_dependency === executionDependency;
}

function selectGoEntries(root, phase, section, coverage, executionDependency, packagePatterns) {
  if (!validGoSections.has(section)) {
    throw new Error(`invalid go manifest section ${section}`);
  }
  if (packagePatterns.length === 0) {
    throw new Error("go manifest selection requires at least one package pattern");
  }
  const { manifest } = loadManifest(root, phase);
  return collectEntries(manifest).filter(
    (entry) =>
      entry.section === section &&
      entry.runner === "go_test" &&
      entry.coverage === coverage &&
      entryMatchesExecutionDependency(entry, executionDependency) &&
      packagePatterns.some((pattern) => packageMatchesPattern(entry.package, pattern)),
  );
}

let cachedGoModulePath;

function loadGoModulePath(root) {
  if (cachedGoModulePath !== undefined) {
    return cachedGoModulePath;
  }
  const goMod = readFileSync(path.join(root, "go.mod"), "utf8");
  const match = goMod.match(/^module\s+(\S+)$/m);
  if (!match) {
    throw new Error("unable to determine Go module path from go.mod");
  }
  cachedGoModulePath = match[1];
  return cachedGoModulePath;
}

function toGoImportPath(root, repoRelativePackage) {
  if (!repoRelativePackage.startsWith("./")) {
    throw new Error(`manifest Go package must be repo-relative: ${repoRelativePackage}`);
  }
  const suffix = repoRelativePackage.slice(2);
  if (suffix === "") {
    return loadGoModulePath(root);
  }
  return `${loadGoModulePath(root)}/${suffix}`;
}

function goLogKey(pkg, test) {
  return `${pkg}::${test}`;
}

function describeGoSymbol(entry, symbol) {
  return `${symbol} [${entry.package}]`;
}

function readGoLogTopLevelStatuses(logFile) {
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

function flattenPlaywrightSuites(suites, specs = []) {
  for (const suite of suites ?? []) {
    flattenPlaywrightSuites(suite.suites, specs);
    for (const spec of suite.specs ?? []) {
      specs.push(spec);
    }
  }
  return specs;
}

function selectPlaywrightEntries(root, phase, coverage, executionDependency) {
  const { manifest } = loadManifest(root, phase);
  return collectEntries(manifest).filter(
    (entry) =>
      entry.section === "e2e" &&
      entry.runner === "playwright" &&
      entry.coverage === coverage &&
      entryMatchesExecutionDependency(entry, executionDependency),
  );
}

function normalizePlaywrightFile(file) {
  if (!file.startsWith("apps/web/")) {
    throw new Error(`playwright manifest file must live under apps/web/: ${file}`);
  }
  return file.slice("apps/web/".length);
}

function selectVitestEntries(root, phase, coverage, executionDependency) {
  const { manifest } = loadManifest(root, phase);
  return collectEntries(manifest).filter(
    (entry) =>
      entry.section === "unit" &&
      entry.runner === "vitest" &&
      entry.coverage === coverage &&
      entryMatchesExecutionDependency(entry, executionDependency),
  );
}

function normalizeVitestFile(file) {
  if (!file.startsWith("apps/web/")) {
    throw new Error(`vitest manifest file must live under apps/web/: ${file}`);
  }
  return file.slice("apps/web/".length);
}

function readVitestReport(reportFile) {
  return JSON.parse(readFileSync(reportFile, "utf8"));
}

function verifyVitestRun(reportFile, expectedTitles) {
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

function summarizePlaywrightErrors(report) {
  const messages = (report.errors ?? [])
    .map((error) => error?.message)
    .filter((message) => typeof message === "string" && message.trim() !== "");
  return messages.join("; ");
}

function detectPlaywrightSetupFailure(report) {
  const specs = flattenPlaywrightSuites(report.suites);
  if (specs.length > 0 || (report.errors ?? []).length === 0) {
    return null;
  }
  const summary = summarizePlaywrightErrors(report);
  return summary === "" ? "playwright setup failure" : `playwright setup failure: ${summary}`;
}

function verifyPlaywrightSpecSet(reportFile, expectedTitles) {
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

function extractPlaywrightStatuses(spec) {
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

function printLines(lines) {
  process.stdout.write(`${lines.join("\n")}\n`);
}

function main(argv) {
  const [command, ...rest] = argv;
  const root = process.cwd();

  switch (command) {
    case "go-regex": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} go tests found for ${phase} ${section} in ${packagePatterns.join(", ")}`);
      }
      printLines([exactRegex(entries.flatMap((entry) => goEntrySymbols(entry)))]);
      return;
    }

    case "go-count": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, packagePatterns);
      printLines([String(entries.length)]);
      return;
    }

    case "go-verify-log": {
      const [phase, section, coverage, executionDependency = "", logFile, ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} go tests found for ${phase} ${section} in ${packagePatterns.join(", ")}`);
      }
      const actual = readGoLogTopLevelStatuses(logFile);
      const passed = [];
      const missing = [];
      const skipped = [];
      const failed = [];
      const incomplete = [];
      for (const entry of entries) {
        for (const symbol of goEntrySymbols(entry)) {
          const key = goLogKey(toGoImportPath(root, entry.package), symbol);
          const result = actual.get(key);
          if (!result) {
            missing.push(describeGoSymbol(entry, symbol));
            continue;
          }
          switch (result.status) {
            case "pass":
              passed.push(symbol);
              break;
            case "skip":
              skipped.push(describeGoSymbol(entry, symbol));
              break;
            case "fail":
              failed.push(describeGoSymbol(entry, symbol));
              break;
            default:
              incomplete.push(describeGoSymbol(entry, symbol));
              break;
          }
        }
      }
      if (missing.length > 0 || skipped.length > 0 || failed.length > 0 || incomplete.length > 0) {
        throw new Error(
          `manifest-go execution mismatch: missing=${missing.join(",") || "none"} skipped=${skipped.join(",") || "none"} failed=${failed.join(",") || "none"} incomplete=${incomplete.join(",") || "none"}`,
        );
      }
      printLines([
        `matched go manifest tests: ${passed.length}`,
        ...passed.sort().map((symbol) => `  ${symbol}`),
      ]);
      return;
    }

    case "playwright-files": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      const files = [...new Set(entries.map((entry) => normalizePlaywrightFile(entry.file)))].sort();
      printLines(files);
      return;
    }

    case "playwright-grep": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      printLines([alternationRegex(entries.map((entry) => entry.title))]);
      return;
    }

    case "playwright-verify-list": {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.map((entry) => entry.title).sort();
      verifyPlaywrightSpecSet(reportFile, expectedTitles);
      printLines([
        `listed playwright manifest tests: ${expectedTitles.length}`,
        ...expectedTitles.map((title) => `  ${title}`),
      ]);
      return;
    }

    case "playwright-verify-run": {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.map((entry) => entry.title).sort();
      const report = verifyPlaywrightSpecSet(reportFile, expectedTitles);
      const specs = flattenPlaywrightSuites(report.suites);
      const failed = [];
      const executed = [];
      for (const expectedTitle of expectedTitles) {
        const spec = specs.find((candidate) => candidate.title === expectedTitle);
        if (!spec) {
          failed.push(`${expectedTitle} (not found)`);
          continue;
        }
        const statuses = extractPlaywrightStatuses(spec);
        if (statuses.length === 0) {
          failed.push(`${expectedTitle} (not executed)`);
          continue;
        }
        const acceptable = statuses.every((status) => status === "passed" || status === "flaky");
        if (!acceptable) {
          failed.push(`${expectedTitle} (${statuses.join(",")})`);
          continue;
        }
        executed.push(expectedTitle);
      }
      if (failed.length > 0) {
        throw new Error(`playwright execution mismatch: ${failed.join("; ")}`);
      }
      printLines([
        `matched playwright manifest tests: ${executed.length}`,
        ...executed.map((title) => `  ${title}`),
      ]);
      return;
    }

    case "vitest-files": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} vitest tests found for ${phase}`);
      }
      const files = [...new Set(entries.map((entry) => normalizeVitestFile(entry.file)))].sort();
      printLines(files);
      return;
    }

    case "vitest-grep": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} vitest tests found for ${phase}`);
      }
      printLines([`${alternationRegex(entries.map((entry) => entry.title))}$`]);
      return;
    }

    case "vitest-verify-run": {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.map((entry) => entry.title).sort();
      const result = verifyVitestRun(reportFile, expectedTitles);
      printLines([
        `matched vitest manifest tests: ${result.executed.length}`,
        ...result.files.map((file) => `  file ${file}`),
        ...result.executed.map((title) => `  ${title}`),
      ]);
      return;
    }

    default:
      throw new Error(`unknown phase-manifest command ${command}`);
  }
}

if (import.meta.url === `file://${process.argv[1]}`) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exit(1);
  }
}
