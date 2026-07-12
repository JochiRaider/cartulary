import {
  goEntrySymbols,
  playwrightEntryTitles,
  vitestEntryTitles,
} from "./phase-entry-evidence.mjs";
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
  selectPlaywrightEntriesForFrontendRowIDs,
  selectPlaywrightEntriesForSpecs,
  selectPlaywrightPhases,
  selectVitestEntries,
  selectVitestEntriesForRowIDs,
  selectVitestPhases,
} from "./phase-selection.mjs";

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

  "playwright-phases-for-ids": (rest, root) => {
      const [coverage, executionDependency, rowIDsValue = ""] = rest;
      const rowIDs = rowIDsValue.split(",").map((value) => value.trim()).filter(Boolean);
      const entries = selectPlaywrightEntriesForFrontendRowIDs(root, coverage, executionDependency, rowIDs);
      const phases = [...new Set(entries.map((entry) => entry.phase))].sort();
      if (phases.length > 0) {
        printLines(phases);
      }
  },

  "playwright-grep-for-ids": (rest, root) => {
      const [coverage, executionDependency, rowIDsValue = ""] = rest;
      const rowIDs = rowIDsValue.split(",").map((value) => value.trim()).filter(Boolean);
      const entries = selectPlaywrightEntriesForFrontendRowIDs(root, coverage, executionDependency, rowIDs);
      printLines([alternationRegex(entries.flatMap((entry) => playwrightEntryTitles(entry)))]);
  },

  "playwright-files-for-ids": (rest, root) => {
      const [coverage, executionDependency, rowIDsValue = ""] = rest;
      const rowIDs = rowIDsValue.split(",").map((value) => value.trim()).filter(Boolean);
      const entries = selectPlaywrightEntriesForFrontendRowIDs(root, coverage, executionDependency, rowIDs);
      printLines([...new Set(entries.map((entry) => normalizePlaywrightFile(entry.file)))].sort());
  },

  "playwright-selection-report-for-ids": (rest, root) => {
      const [phase, coverage, executionDependency, rowIDsValue = ""] = rest;
      const rowIDs = rowIDsValue.split(",").map((value) => value.trim()).filter(Boolean);
      const entries = selectPlaywrightEntriesForFrontendRowIDs(root, coverage, executionDependency, rowIDs)
        .filter((entry) => entry.phase === phase);
      const selectedTests = entries.flatMap((entry) =>
        playwrightEntryTitles(entry).map((title) => ({
          id: entry.id,
          file: normalizePlaywrightFile(entry.file),
          title,
          coverage: entry.coverage,
          execution_dependency: entry.execution_dependency ?? "",
        })),
      );
      process.stdout.write(`${JSON.stringify({
        schema_id: "cartulary.playwright_manifest_selection.v1",
        phase,
        coverage,
        execution_dependency: executionDependency,
        expected_count: selectedTests.length,
        selected_tests: selectedTests,
      }, null, 2)}\n`);
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

  "vitest-files-for-ids": (rest, root) => {
      const [phase, coverage, executionDependency = "", rowIDsValue = ""] = rest;
      const rowIDs = rowIDsValue.split(",").map((value) => value.trim()).filter(Boolean);
      const entries = selectVitestEntriesForRowIDs(
        root,
        phase,
        coverage,
        executionDependency,
        rowIDs,
      );
      if (entries.length === 0) {
        throw new Error(`no selected ${coverage} vitest tests found for ${phase}`);
      }
      printLines([...new Set(entries.map((entry) => normalizeVitestFile(entry.file)))].sort());
  },

  "vitest-grep-for-ids": (rest, root) => {
      const [phase, coverage, executionDependency = "", rowIDsValue = ""] = rest;
      const rowIDs = rowIDsValue.split(",").map((value) => value.trim()).filter(Boolean);
      const entries = selectVitestEntriesForRowIDs(
        root,
        phase,
        coverage,
        executionDependency,
        rowIDs,
      );
      if (entries.length === 0) {
        throw new Error(`no selected ${coverage} vitest tests found for ${phase}`);
      }
      printLines([`${alternationRegex(entries.flatMap((entry) => vitestEntryTitles(entry)))}$`]);
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
