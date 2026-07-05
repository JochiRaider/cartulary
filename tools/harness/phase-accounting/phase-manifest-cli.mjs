import {
  goEntrySymbols,
  playwrightEntryTitles,
  supportGoEntrySymbols,
  vitestEntryTitles,
} from "./phase-entry-evidence.mjs";
import {
  effectiveGoEntryPostgresFixturePolicy,
  effectiveSupportGoEntryPostgresFixturePolicy,
  fixturePolicyAssignments,
  goEntryPostgresFixtureBudget,
  resetTableAssignments,
  supportGoEntryPostgresFixtureBudget,
} from "./phase-fixture-policy.mjs";
import { phaseManifestNames } from "./phase-manifest-loader.mjs";
import {
  emptyGoManifestSelectionAllowed,
  loadPhasePolicyExceptions,
} from "./phase-policy-exceptions.mjs";
import {
  normalizePlaywrightFile,
  normalizeVitestFile,
  selectGoEntries,
  selectPlaywrightEntries,
  selectPlaywrightEntriesAll,
  selectPlaywrightEntriesForSpecs,
  selectPlaywrightPhases,
  selectSupportGoEntries,
  selectVitestEntries,
  selectVitestPhases,
  toGoImportPath,
} from "./phase-selection.mjs";
import {
  describeGoSymbol,
  extractPlaywrightStatuses,
  goLogKey,
  playwrightReportSpecs,
  readGoLogTopLevelStatuses,
  verifyPlaywrightSpecSet,
  verifyVitestRun,
} from "./phase-run-verification.mjs";

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

function printLines(lines) {
  process.stdout.write(`${lines.join("\n")}\n`);
}

const phaseManifestCommandHandlers = {
  "list-phases": (_rest, root) => {
      printLines(phaseManifestNames(root));
  },

  "list-registered-manifest-phases": (_rest, root) => {
      printLines(phaseManifestNames(root, { includePlanned: true }));
  },

  "go-regex": (rest, root) => {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} go tests found for ${phase} ${section} in ${packagePatterns.join(", ")}`);
      }
      printLines([exactRegex(entries.flatMap((entry) => goEntrySymbols(entry)))]);
  },

  "go-count": (rest, root) => {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      printLines([String(entries.length)]);
  },

  "go-postgres-fixture-policy-tests": (rest, root) => {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      printLines([
        fixturePolicyAssignments(
          entries,
          goEntrySymbols,
          effectiveGoEntryPostgresFixturePolicy,
        ).join(","),
      ]);
  },

  "go-postgres-reset-table-tests": (rest, root) => {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      printLines([resetTableAssignments(entries, goEntrySymbols, goEntryPostgresFixtureBudget).join(",")]);
  },

  "go-family-regex": (rest, root) => {
      const [phase, section, coverage, executionDependency = "", executionFamily = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(
        root,
        phase,
        section,
        coverage,
        executionDependency,
        executionFamily,
        packagePatterns,
      );
      if (entries.length === 0) {
        throw new Error(
          `no ${coverage} go tests found for ${phase} ${section} ${executionFamily} in ${packagePatterns.join(", ")}`,
        );
      }
      printLines([exactRegex(entries.flatMap((entry) => goEntrySymbols(entry)))]);
  },

  "go-family-count": (rest, root) => {
      const [phase, section, coverage, executionDependency = "", executionFamily = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(
        root,
        phase,
        section,
        coverage,
        executionDependency,
        executionFamily,
        packagePatterns,
      );
      printLines([String(entries.length)]);
  },

  "support-go-regex": (rest, root) => {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no support go tests found for ${phase} ${target} in ${packagePatterns.join(", ")}`);
      }
      printLines([exactRegex(entries.flatMap((entry) => supportGoEntrySymbols(entry)))]);
  },

  "support-go-count": (rest, root) => {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      printLines([String(entries.length)]);
  },

  "support-go-postgres-fixture-policy-tests": (rest, root) => {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      printLines([
        fixturePolicyAssignments(
          entries,
          supportGoEntrySymbols,
          effectiveSupportGoEntryPostgresFixturePolicy,
        ).join(","),
      ]);
  },

  "support-go-postgres-reset-table-tests": (rest, root) => {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      printLines([
        resetTableAssignments(
          entries,
          supportGoEntrySymbols,
          supportGoEntryPostgresFixtureBudget,
        ).join(","),
      ]);
  },

  "support-go-family-regex": (rest, root) => {
      const [phase, target, executionFamily = "", ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, executionFamily, packagePatterns);
      if (entries.length === 0) {
        throw new Error(
          `no support go tests found for ${phase} ${target} ${executionFamily} in ${packagePatterns.join(", ")}`,
        );
      }
      printLines([exactRegex(entries.flatMap((entry) => supportGoEntrySymbols(entry)))]);
  },

  "support-go-family-count": (rest, root) => {
      const [phase, target, executionFamily = "", ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, executionFamily, packagePatterns);
      printLines([String(entries.length)]);
  },

  "go-verify-log": (rest, root) => {
      const [phase, section, coverage, executionDependency = "", logFile, ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
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
  },

  "playwright-files": (rest, root) => {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      const files = [...new Set(entries.map((entry) => normalizePlaywrightFile(entry.file)))].sort();
      printLines(files);
  },

  "playwright-files-many": (rest, root) => {
      const entries = selectPlaywrightEntriesForSpecs(root, rest);
      const files = [...new Set(entries.map((entry) => normalizePlaywrightFile(entry.file)))].sort();
      printLines(files);
  },

  "playwright-files-all": (rest, root) => {
      const [coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntriesAll(root, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${executionDependency || "all dependencies"}`);
      }
      const files = [...new Set(entries.map((entry) => normalizePlaywrightFile(entry.file)))].sort();
      printLines(files);
  },

  "playwright-grep": (rest, root) => {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      printLines([alternationRegex(entries.flatMap((entry) => playwrightEntryTitles(entry)))]);
  },

  "playwright-count": (rest, root) => {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      printLines([String(entries.flatMap((entry) => playwrightEntryTitles(entry)).length)]);
  },

  "playwright-count-all": (rest, root) => {
      const [coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntriesAll(root, coverage, executionDependency);
      printLines([String(entries.flatMap((entry) => playwrightEntryTitles(entry)).length)]);
  },

  "playwright-phases": (rest, root) => {
      const [coverage, executionDependency = ""] = rest;
      const phases = selectPlaywrightPhases(root, coverage, executionDependency);
      if (phases.length === 0) {
        throw new Error(`no ${coverage} playwright phases found`);
      }
      printLines(phases);
  },

  "playwright-grep-many": (rest, root) => {
      const entries = selectPlaywrightEntriesForSpecs(root, rest);
      printLines([alternationRegex(entries.flatMap((entry) => playwrightEntryTitles(entry)))]);
  },

  "phase-policy-exceptions-validate": (_rest, root) => {
      loadPhasePolicyExceptions(root);
      printLines(["phase policy exceptions verified"]);
  },

  "empty-go-manifest-selection-allowed": (rest, root) => {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      if (
        emptyGoManifestSelectionAllowed(
          root,
          phase,
          section,
          coverage,
          executionDependency,
          packagePatterns,
        )
      ) {
        printLines(["allowed"]);
        return;
      }
      process.exit(1);
  },

  "playwright-selection-report": (rest, root) => {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      const selectedTests = entries.flatMap((entry) =>
        playwrightEntryTitles(entry).map((title) => ({
          id: entry.id,
          file: normalizePlaywrightFile(entry.file),
          title,
          coverage: entry.coverage,
          execution_dependency: entry.execution_dependency ?? "",
        })),
      );
      process.stdout.write(
        `${JSON.stringify(
          {
            schema_id: "cartulary.playwright_manifest_selection.v1",
            phase,
            coverage,
            execution_dependency: executionDependency,
            expected_count: selectedTests.length,
            selected_tests: selectedTests,
          },
          null,
          2,
        )}\n`,
      );
  },

  "playwright-verify-list": (rest, _root) => {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectPlaywrightEntries(_root, phase, coverage, executionDependency);
      const expectedTitles = entries.flatMap((entry) => playwrightEntryTitles(entry)).sort();
      verifyPlaywrightSpecSet(reportFile, expectedTitles);
      printLines([
        `listed playwright manifest tests: ${expectedTitles.length}`,
        ...expectedTitles.map((title) => `  ${title}`),
      ]);
  },

  "playwright-verify-run": (rest, root) => {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.flatMap((entry) => playwrightEntryTitles(entry)).sort();
      const report = verifyPlaywrightSpecSet(reportFile, expectedTitles);
      const specs = playwrightReportSpecs(report);
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
  },

  "vitest-files": (rest, root) => {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} vitest tests found for ${phase}`);
      }
      const files = [...new Set(entries.map((entry) => normalizeVitestFile(entry.file)))].sort();
      printLines(files);
  },

  "vitest-phases": (rest, root) => {
      const [coverage, executionDependency = ""] = rest;
      const phases = selectVitestPhases(root, coverage, executionDependency);
      if (phases.length === 0) {
        throw new Error(`no ${coverage} vitest phases found`);
      }
      printLines(phases);
  },

  "vitest-grep": (rest, root) => {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} vitest tests found for ${phase}`);
      }
      printLines([`${alternationRegex(entries.flatMap((entry) => vitestEntryTitles(entry)))}$`]);
  },

  "vitest-verify-run": (rest, root) => {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.flatMap((entry) => vitestEntryTitles(entry)).sort();
      const result = verifyVitestRun(reportFile, expectedTitles);
      printLines([
        `matched vitest manifest tests: ${result.executed.length}`,
        ...result.files.map((file) => `  file ${file}`),
        ...result.executed.map((title) => `  ${title}`),
      ]);
  },
};

export function runPhaseManifestCLI(argv = process.argv.slice(2), root = process.cwd()) {
  const [command, ...rest] = argv;
  const handler = phaseManifestCommandHandlers[command];
  if (!handler) {
    throw new Error(`unknown phase-manifest command ${command}`);
  }
  handler(rest, root);
}

if (import.meta.url === `file://${process.argv[1]}`) {
  try {
    runPhaseManifestCLI(process.argv.slice(2));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exit(1);
  }
}
