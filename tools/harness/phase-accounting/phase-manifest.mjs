import {
  assertAuthoritativeEvidenceNames,
  authoritativeEvidenceNameViolations,
  collectEntries,
  collectSupportGoEntries,
  entryClaimStatus,
  entryEvidenceNames,
  entryIsExecutable,
  goEntryScenarioSymbols,
  goEntrySymbols,
  playwrightEntryTitles,
  rowIDFragments,
  supportGoEntrySymbols,
  vitestEntryTitles,
} from "./phase-entry-evidence.mjs";
import {
  effectiveGoEntryPostgresFixtureBudget,
  effectiveGoEntryPostgresFixturePolicy,
  effectiveSupportGoEntryPostgresFixtureBudget,
  effectiveSupportGoEntryPostgresFixturePolicy,
  fixturePolicyAssignments,
  goEntryPostgresFixtureBudget,
  goEntryPostgresFixturePolicy,
  resetTableAssignments,
  supportGoEntryPostgresFixtureBudget,
  supportGoEntryPostgresFixturePolicy,
} from "./phase-fixture-policy.mjs";
import { loadManifest, phaseManifestNames } from "./phase-manifest-loader.mjs";
import { validateManifest } from "./phase-manifest-validation.mjs";
import {
  emptyGoManifestSelectionAllowed,
  loadPhasePolicyExceptions,
} from "./phase-policy-exceptions.mjs";
import {
  normalizePlaywrightFile,
  normalizeVitestFile,
  packageMatchesPattern,
  selectGoEntries,
  selectManifestEntries,
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

export {
  assertAuthoritativeEvidenceNames,
  authoritativeEvidenceNameViolations,
  collectEntries,
  collectSupportGoEntries,
  effectiveGoEntryPostgresFixturePolicy,
  effectiveGoEntryPostgresFixtureBudget,
  effectiveSupportGoEntryPostgresFixturePolicy,
  effectiveSupportGoEntryPostgresFixtureBudget,
  entryClaimStatus,
  entryEvidenceNames,
  entryIsExecutable,
  goEntryPostgresFixturePolicy,
  goEntryPostgresFixtureBudget,
  goEntryScenarioSymbols,
  goEntrySymbols,
  loadManifest,
  loadPhasePolicyExceptions,
  packageMatchesPattern,
  phaseManifestNames,
  playwrightEntryTitles,
  rowIDFragments,
  selectManifestEntries,
  selectPlaywrightEntries,
  supportGoEntryPostgresFixturePolicy,
  supportGoEntryPostgresFixtureBudget,
  supportGoEntrySymbols,
  validateManifest,
  vitestEntryTitles,
};

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

function main(argv) {
  const [command, ...rest] = argv;
  const root = process.cwd();

  switch (command) {
    case "list-phases": {
      printLines(phaseManifestNames(root));
      return;
    }

    case "list-registered-manifest-phases": {
      printLines(phaseManifestNames(root, { includePlanned: true }));
      return;
    }

    case "go-regex": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} go tests found for ${phase} ${section} in ${packagePatterns.join(", ")}`);
      }
      printLines([exactRegex(entries.flatMap((entry) => goEntrySymbols(entry)))]);
      return;
    }

    case "go-count": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      printLines([String(entries.length)]);
      return;
    }

    case "go-postgres-fixture-policy-tests": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      printLines([
        fixturePolicyAssignments(
          entries,
          goEntrySymbols,
          effectiveGoEntryPostgresFixturePolicy,
        ).join(","),
      ]);
      return;
    }

    case "go-postgres-reset-table-tests": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      printLines([resetTableAssignments(entries, goEntrySymbols, goEntryPostgresFixtureBudget).join(",")]);
      return;
    }

    case "go-family-regex": {
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
      return;
    }

    case "go-family-count": {
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
      return;
    }

    case "support-go-regex": {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no support go tests found for ${phase} ${target} in ${packagePatterns.join(", ")}`);
      }
      printLines([exactRegex(entries.flatMap((entry) => supportGoEntrySymbols(entry)))]);
      return;
    }

    case "support-go-count": {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      printLines([String(entries.length)]);
      return;
    }

    case "support-go-postgres-fixture-policy-tests": {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      printLines([
        fixturePolicyAssignments(
          entries,
          supportGoEntrySymbols,
          effectiveSupportGoEntryPostgresFixturePolicy,
        ).join(","),
      ]);
      return;
    }

    case "support-go-postgres-reset-table-tests": {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      printLines([
        resetTableAssignments(
          entries,
          supportGoEntrySymbols,
          supportGoEntryPostgresFixtureBudget,
        ).join(","),
      ]);
      return;
    }

    case "support-go-family-regex": {
      const [phase, target, executionFamily = "", ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, executionFamily, packagePatterns);
      if (entries.length === 0) {
        throw new Error(
          `no support go tests found for ${phase} ${target} ${executionFamily} in ${packagePatterns.join(", ")}`,
        );
      }
      printLines([exactRegex(entries.flatMap((entry) => supportGoEntrySymbols(entry)))]);
      return;
    }

    case "support-go-family-count": {
      const [phase, target, executionFamily = "", ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, executionFamily, packagePatterns);
      printLines([String(entries.length)]);
      return;
    }

    case "go-verify-log": {
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

    case "playwright-files-many": {
      const entries = selectPlaywrightEntriesForSpecs(root, rest);
      const files = [...new Set(entries.map((entry) => normalizePlaywrightFile(entry.file)))].sort();
      printLines(files);
      return;
    }

    case "playwright-files-all": {
      const [coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntriesAll(root, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${executionDependency || "all dependencies"}`);
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
      printLines([alternationRegex(entries.flatMap((entry) => playwrightEntryTitles(entry)))]);
      return;
    }

    case "playwright-count": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      printLines([String(entries.flatMap((entry) => playwrightEntryTitles(entry)).length)]);
      return;
    }

    case "playwright-count-all": {
      const [coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntriesAll(root, coverage, executionDependency);
      printLines([String(entries.flatMap((entry) => playwrightEntryTitles(entry)).length)]);
      return;
    }

    case "playwright-phases": {
      const [coverage, executionDependency = ""] = rest;
      const phases = selectPlaywrightPhases(root, coverage, executionDependency);
      if (phases.length === 0) {
        throw new Error(`no ${coverage} playwright phases found`);
      }
      printLines(phases);
      return;
    }

    case "playwright-grep-many": {
      const entries = selectPlaywrightEntriesForSpecs(root, rest);
      printLines([alternationRegex(entries.flatMap((entry) => playwrightEntryTitles(entry)))]);
      return;
    }

    case "phase-policy-exceptions-validate": {
      loadPhasePolicyExceptions(root);
      printLines(["phase policy exceptions verified"]);
      return;
    }

    case "empty-go-manifest-selection-allowed": {
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
      return;
    }

    case "playwright-selection-report": {
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
      return;
    }

    case "playwright-verify-list": {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.flatMap((entry) => playwrightEntryTitles(entry)).sort();
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

    case "vitest-phases": {
      const [coverage, executionDependency = ""] = rest;
      const phases = selectVitestPhases(root, coverage, executionDependency);
      if (phases.length === 0) {
        throw new Error(`no ${coverage} vitest phases found`);
      }
      printLines(phases);
      return;
    }

    case "vitest-grep": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} vitest tests found for ${phase}`);
      }
      printLines([`${alternationRegex(entries.flatMap((entry) => vitestEntryTitles(entry)))}$`]);
      return;
    }

    case "vitest-verify-run": {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.flatMap((entry) => vitestEntryTitles(entry)).sort();
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
