import { readFileSync } from "node:fs";
import path from "node:path";

const sectionDefinitions = [
  ["unit", "U-"],
  ["integration", "I-"],
  ["e2e", "E-"],
];

const validCoverage = new Set(["authoritative", "supplemental"]);
const validGoSections = new Set(["unit", "integration", "e2e"]);

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
  const phaseMatch = /^phase(\d+)$/.exec(phase);
  if (!phaseMatch) {
    throw new Error(`invalid phase name ${phase}; expected phase<number>`);
  }
  const phaseNumber = phaseMatch[1];
  const { manifestPath, manifest } = loadManifest(root, phase);

  if (!Array.isArray(manifest.expected_ids) || manifest.expected_ids.length === 0) {
    throw new Error(`manifest ${manifestPath} must define a non-empty expected_ids array`);
  }

  const entries = [];
  for (const [section, prefix] of sectionDefinitions) {
    for (const entry of manifest[section] ?? []) {
      if (typeof entry.id !== "string" || !entry.id.startsWith(prefix)) {
        throw new Error(`manifest entry in ${section} has invalid id: ${JSON.stringify(entry)}`);
      }
      if (!new RegExp(`^${prefix}${phaseNumber}-\\d{2}$`).test(entry.id)) {
        throw new Error(`manifest entry ${entry.id} does not belong to ${phase}`);
      }
      if (typeof entry.coverage !== "string" || !validCoverage.has(entry.coverage)) {
        throw new Error(`manifest entry ${entry.id} must declare coverage=authoritative|supplemental`);
      }
      if (entry.runner === "playwright") {
        if (typeof entry.title !== "string" || entry.title.trim() === "") {
          throw new Error(`manifest entry ${entry.id} is missing a non-empty title`);
        }
        if (typeof entry.file !== "string" || !entry.file.startsWith("apps/web/e2e/")) {
          throw new Error(`manifest entry ${entry.id} must point at an apps/web/e2e file`);
        }
      } else if (entry.runner === "go_test") {
        if (typeof entry.symbol !== "string" || entry.symbol.trim() === "") {
          throw new Error(`manifest entry ${entry.id} is missing a non-empty symbol`);
        }
        if (typeof entry.package !== "string" || !entry.package.startsWith("./")) {
          throw new Error(`manifest entry ${entry.id} must declare a repo-relative Go package owner`);
        }
      } else {
        throw new Error(`manifest entry ${entry.id} must declare runner=go_test|playwright`);
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

  for (const entry of entries) {
    const targetPath = path.join(root, entry.file);
    const source = readFileSync(targetPath, "utf8");
    const needle = entry.symbol ?? entry.title;
    if (!source.includes(needle)) {
      throw new Error(`manifest entry ${entry.id} not found in ${entry.file}: ${needle}`);
    }
  }

  for (const target of manifest.forbidden_id_files ?? []) {
    const source = readFileSync(path.join(root, target), "utf8");
    const hyphenMatches = source.match(new RegExp(String.raw`\b[UIE]-${phaseNumber}-\d{2}\b`, "g")) ?? [];
    const underscoreMatches =
      source.match(new RegExp(String.raw`\b[UIE]_${phaseNumber}_\d{2}\b`, "g")) ?? [];
    const claimedIDs = new Set([
      ...hyphenMatches,
      ...underscoreMatches.map((value) => value.replaceAll("_", "-")),
    ]);
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

function selectGoEntries(root, phase, section, coverage, packagePatterns) {
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
      packagePatterns.some((pattern) => packageMatchesPattern(entry.package, pattern)),
  );
}

function readGoLogRunTests(logFile) {
  const seen = new Set();
  for (const rawLine of readFileSync(logFile, "utf8").split(/\r?\n/)) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    const entry = JSON.parse(line);
    if (entry.Action !== "run" || typeof entry.Test !== "string" || entry.Test.includes("/")) {
      continue;
    }
    seen.add(entry.Test);
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

function selectPlaywrightEntries(root, phase, coverage) {
  const { manifest } = loadManifest(root, phase);
  return collectEntries(manifest).filter(
    (entry) => entry.section === "e2e" && entry.runner === "playwright" && entry.coverage === coverage,
  );
}

function normalizePlaywrightFile(file) {
  if (!file.startsWith("apps/web/")) {
    throw new Error(`playwright manifest file must live under apps/web/: ${file}`);
  }
  return file.slice("apps/web/".length);
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
      const [phase, section, coverage, ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} go tests found for ${phase} ${section} in ${packagePatterns.join(", ")}`);
      }
      printLines([exactRegex(entries.map((entry) => entry.symbol))]);
      return;
    }

    case "go-verify-log": {
      const [phase, section, coverage, logFile, ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} go tests found for ${phase} ${section} in ${packagePatterns.join(", ")}`);
      }
      const expected = entries.map((entry) => entry.symbol).sort();
      const actual = readGoLogRunTests(logFile);
      const missing = expected.filter((symbol) => !actual.has(symbol));
      const unexpected = Array.from(actual)
        .filter((symbol) => expected.includes(symbol))
        .sort();
      if (missing.length > 0) {
        throw new Error(`manifest-go execution mismatch: missing=${missing.join(", ")}`);
      }
      printLines([
        `matched go manifest tests: ${expected.length}`,
        ...unexpected.map((symbol) => `  ${symbol}`),
      ]);
      return;
    }

    case "playwright-files": {
      const [phase, coverage] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      const files = [...new Set(entries.map((entry) => normalizePlaywrightFile(entry.file)))].sort();
      printLines(files);
      return;
    }

    case "playwright-grep": {
      const [phase, coverage] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      printLines([alternationRegex(entries.map((entry) => entry.title))]);
      return;
    }

    case "playwright-verify-list": {
      const [phase, coverage, reportFile] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage);
      const expectedTitles = entries.map((entry) => entry.title).sort();
      verifyPlaywrightSpecSet(reportFile, expectedTitles);
      printLines([
        `listed playwright manifest tests: ${expectedTitles.length}`,
        ...expectedTitles.map((title) => `  ${title}`),
      ]);
      return;
    }

    case "playwright-verify-run": {
      const [phase, coverage, reportFile] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage);
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
