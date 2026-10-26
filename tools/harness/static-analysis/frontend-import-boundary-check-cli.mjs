#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "../workspace/child-process.mjs";
import { checkFrontendImportBoundaries } from "./frontend-import-boundary-check.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "../../..");

function usage() {
  throw new Error(
    "usage: check-frontend-import-boundaries.mjs [--config <path>] [--root <path>] [--warnings-as-errors]",
  );
}

function parseArgs(argv) {
  const options = {
    config:
      process.env.FRONTEND_IMPORT_BOUNDARIES_CONFIG ??
      "tools/frontend_import_boundaries.json",
    root: process.env.CARTULARY_REPO_ROOT ?? defaultRepoRoot,
    warningsAsErrors: false,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--config") {
      options.config = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--root") {
      options.root = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--warnings-as-errors") {
      options.warningsAsErrors = true;
      continue;
    }
    usage();
  }
  if (!options.config || !options.root) {
    usage();
  }
  return options;
}

function formatDiagnostic(diagnostic, warningsAsErrors) {
  const effectiveLevel =
    warningsAsErrors && diagnostic.level === "warning" ? "error" : diagnostic.level;
  return `${diagnostic.file}:${diagnostic.line}:${diagnostic.column}: ${effectiveLevel}: ${diagnostic.ruleID}: ${diagnostic.message} (${diagnostic.importKind} "${diagnostic.specifier}")`;
}

function emitFrontendTargetSummary(root, status) {
  const target = process.env.CARTULARY_TEST_TARGET ?? "";
  if (target !== "frontend-import-boundary-check") {
    return;
  }
  const nodeBin = process.env.NODE_BIN || process.execPath;
  const helper =
    process.env.TEST_OUTPUT_SCRIPT ?? path.join(root, "tools/harness/output/test-output.mjs");
  const result = spawnSync(
    nodeBin,
    [
      helper,
      "target-summary",
      target,
      status,
      "--quiet-success",
      "--suppress-machine-output",
    ],
    {
      stdio: "inherit",
    },
  );
  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const root = path.resolve(options.root);
  const configPath = path.resolve(root, options.config);
  const { diagnostics, fileCount, errorCount, warningCount } = checkFrontendImportBoundaries({
    root, config: JSON.parse(readFileSync(configPath, "utf8")),
  });
  for (const diagnostic of diagnostics) {
    console.error(formatDiagnostic(diagnostic, options.warningsAsErrors));
  }
  if (errorCount > 0 || (options.warningsAsErrors && warningCount > 0)) {
    console.error(
      `frontend import boundary check failed: ${errorCount} error(s), ${warningCount} warning(s), ${fileCount} file(s) checked`,
    );
    process.exit(1);
  }
  if (warningCount > 0) {
    console.error(
      `frontend import boundary check completed with ${warningCount} warning(s), ${fileCount} file(s) checked`,
    );
    emitFrontendTargetSummary(root, "pass");
    return;
  }
  console.log(`frontend import boundaries verified: ${fileCount} file(s) checked`);
  emitFrontendTargetSummary(root, "pass");
}

try {
  main();
} catch (error) {
  console.error(`frontend import boundary check failed: ${error instanceof Error ? error.message : String(error)}`);
  process.exit(1);
}
