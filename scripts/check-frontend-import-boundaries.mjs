#!/usr/bin/env node
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import ts from "typescript";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "..");
const schemaID = "cartulary.frontend_import_boundaries.v1";
const sourceExtensions = new Set([".ts", ".tsx", ".mts", ".cts"]);
const ignoredDirectoryNames = new Set([
  ".cache",
  ".git",
  ".pnpm-store",
  "coverage",
  "dist",
  "node_modules",
  "playwright-report",
  "test-results",
  "tmp",
]);

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

function resolvePath(root, value) {
  return path.isAbsolute(value) ? value : path.join(root, value);
}

function repoRelative(root, value) {
  return normalizePath(path.relative(root, value));
}

function normalizePath(value) {
  return value.split(path.sep).join("/");
}

function readJSON(root, value) {
  return JSON.parse(readFileSync(resolvePath(root, value), "utf8"));
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function requireStringArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value.map((entry, index) => requireString(entry, `${label}[${index + 1}]`));
}

function normalizeConfig(raw) {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw new Error("config must be an object");
  }
  if (raw.schema_id !== schemaID) {
    throw new Error(`config must declare schema_id=${schemaID}`);
  }
  const rules = [];
  for (const [ruleIndex, rule] of requireArray(raw.rules, "rules").entries()) {
    const label = `rules[${ruleIndex + 1}]`;
    const normalized = {
      id: requireString(rule?.id, `${label}.id`),
      level: requireString(rule.level, `${label}.level`),
      message: requireString(rule.message, `${label}.message`),
      allowedImporters: (rule.allowed_importers ?? []).map((entry, index) =>
        requireString(entry, `${label}.allowed_importers[${index + 1}]`),
      ),
      restrictedImports: [],
    };
    if (!["error", "warning"].includes(normalized.level)) {
      throw new Error(`${label}.level must be error or warning`);
    }
    for (const [restrictionIndex, restriction] of requireArray(
      rule.restricted_imports,
      `${label}.restricted_imports`,
    ).entries()) {
      const restrictionLabel = `${label}.restricted_imports[${restrictionIndex + 1}]`;
      const kind = requireString(restriction?.kind, `${restrictionLabel}.kind`);
      if (kind === "package") {
        normalized.restrictedImports.push({
          kind,
          includeSubpaths: restriction.include_subpaths === true,
          name: requireString(restriction.name, `${restrictionLabel}.name`),
        });
        continue;
      }
      if (kind === "path_prefix") {
        normalized.restrictedImports.push({
          kind,
          path: normalizePath(requireString(restriction.path, `${restrictionLabel}.path`)),
        });
        continue;
      }
      throw new Error(`${restrictionLabel}.kind must be package or path_prefix`);
    }
    rules.push(normalized);
  }
  return {
    scanRoots: requireStringArray(raw.scan_roots, "scan_roots"),
    scanExcludes: (raw.scan_excludes ?? []).map((entry, index) =>
      requireString(entry, `scan_excludes[${index + 1}]`),
    ),
    rules,
  };
}

function requireArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value;
}

function matchesGlob(pattern, value) {
  const normalizedPattern = normalizePath(pattern);
  if (normalizedPattern.endsWith("/**")) {
    const prefix = normalizedPattern.slice(0, -3);
    return value === prefix || value.startsWith(`${prefix}/`);
  }
  let source = "^";
  for (let index = 0; index < normalizedPattern.length; index += 1) {
    const char = normalizedPattern[index];
    if (char === "*" && normalizedPattern[index + 1] === "*") {
      source += ".*";
      index += 1;
      continue;
    }
    if (char === "*") {
      source += "[^/]*";
      continue;
    }
    source += char.replace(/[.+^${}()|[\]\\]/g, "\\$&");
  }
  source += "$";
  const regex = new RegExp(source);
  return regex.test(value);
}

function isAllowedImporter(rule, importerPath) {
  return rule.allowedImporters.some((pattern) => matchesGlob(pattern, importerPath));
}

function shouldExclude(config, relativePath) {
  return config.scanExcludes.some((pattern) => matchesGlob(pattern, relativePath));
}

function collectSourceFiles(root, config) {
  const files = [];
  const visit = (absolutePath) => {
    if (!existsSync(absolutePath)) {
      throw new Error(`scan root does not exist: ${repoRelative(root, absolutePath)}`);
    }
    const relativePath = repoRelative(root, absolutePath);
    if (relativePath && shouldExclude(config, relativePath)) {
      return;
    }
    const stats = statSync(absolutePath);
    if (stats.isDirectory()) {
      if (ignoredDirectoryNames.has(path.basename(absolutePath))) {
        return;
      }
      for (const entry of readdirSync(absolutePath).sort()) {
        visit(path.join(absolutePath, entry));
      }
      return;
    }
    if (!stats.isFile()) {
      return;
    }
    if (sourceExtensions.has(path.extname(absolutePath))) {
      files.push(absolutePath);
    }
  };
  for (const scanRoot of config.scanRoots) {
    visit(resolvePath(root, scanRoot));
  }
  return [...new Set(files)].sort();
}

function scriptKindFor(file) {
  if (file.endsWith(".tsx")) {
    return ts.ScriptKind.TSX;
  }
  return ts.ScriptKind.TS;
}

function collectImports(sourceFile) {
  const imports = [];
  const record = (specifier, node, kind) => {
    const { line, character } = sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile));
    imports.push({
      column: character + 1,
      kind,
      line: line + 1,
      specifier,
    });
  };
  const visit = (node) => {
    if (
      (ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) &&
      node.moduleSpecifier &&
      ts.isStringLiteralLike(node.moduleSpecifier)
    ) {
      record(node.moduleSpecifier.text, node.moduleSpecifier, ts.isImportDeclaration(node) ? "import" : "export");
    }
    if (
      ts.isCallExpression(node) &&
      node.expression.kind === ts.SyntaxKind.ImportKeyword &&
      node.arguments.length === 1 &&
      ts.isStringLiteralLike(node.arguments[0])
    ) {
      record(node.arguments[0].text, node.arguments[0], "dynamic import");
    }
    if (
      ts.isCallExpression(node) &&
      ts.isIdentifier(node.expression) &&
      node.expression.text === "require" &&
      node.arguments.length === 1 &&
      ts.isStringLiteralLike(node.arguments[0])
    ) {
      record(node.arguments[0].text, node.arguments[0], "require");
    }
    if (
      ts.isImportTypeNode(node) &&
      ts.isLiteralTypeNode(node.argument) &&
      ts.isStringLiteralLike(node.argument.literal)
    ) {
      record(node.argument.literal.text, node.argument.literal, "import type");
    }
    ts.forEachChild(node, visit);
  };
  visit(sourceFile);
  return imports;
}

function isPackageRestricted(restriction, specifier) {
  if (restriction.includeSubpaths) {
    return specifier === restriction.name || specifier.startsWith(`${restriction.name}/`);
  }
  return specifier === restriction.name;
}

function resolvedRelativeImport(root, importerFile, specifier) {
  if (specifier.startsWith(".")) {
    return repoRelative(root, path.resolve(path.dirname(importerFile), specifier));
  }
  if (path.isAbsolute(specifier)) {
    return repoRelative(root, specifier);
  }
  return "";
}

function isPathPrefixRestricted(root, importerFile, restriction, specifier) {
  const resolved = resolvedRelativeImport(root, importerFile, specifier);
  return resolved === restriction.path || resolved.startsWith(`${restriction.path}/`);
}

function matchesRestriction(root, importerFile, restriction, specifier) {
  if (restriction.kind === "package") {
    return isPackageRestricted(restriction, specifier);
  }
  if (restriction.kind === "path_prefix") {
    return isPathPrefixRestricted(root, importerFile, restriction, specifier);
  }
  return false;
}

function evaluateFile(root, config, file) {
  const relativeFile = repoRelative(root, file);
  const content = readFileSync(file, "utf8");
  const sourceFile = ts.createSourceFile(
    file,
    content,
    ts.ScriptTarget.Latest,
    true,
    scriptKindFor(file),
  );
  const diagnostics = [];
  for (const imported of collectImports(sourceFile)) {
    for (const rule of config.rules) {
      if (isAllowedImporter(rule, relativeFile)) {
        continue;
      }
      const restricted = rule.restrictedImports.some((restriction) =>
        matchesRestriction(root, file, restriction, imported.specifier),
      );
      if (!restricted) {
        continue;
      }
      diagnostics.push({
        column: imported.column,
        file: relativeFile,
        importKind: imported.kind,
        level: rule.level,
        line: imported.line,
        message: rule.message,
        ruleID: rule.id,
        specifier: imported.specifier,
      });
    }
  }
  return diagnostics;
}

function formatDiagnostic(diagnostic, warningsAsErrors) {
  const effectiveLevel =
    warningsAsErrors && diagnostic.level === "warning" ? "error" : diagnostic.level;
  return `${diagnostic.file}:${diagnostic.line}:${diagnostic.column}: ${effectiveLevel}: ${diagnostic.ruleID}: ${diagnostic.message} (${diagnostic.importKind} "${diagnostic.specifier}")`;
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const root = path.resolve(options.root);
  const config = normalizeConfig(readJSON(root, options.config));
  const files = collectSourceFiles(root, config);
  const diagnostics = files.flatMap((file) => evaluateFile(root, config, file));
  const errorCount = diagnostics.filter((diagnostic) => diagnostic.level === "error").length;
  const warningCount = diagnostics.filter((diagnostic) => diagnostic.level === "warning").length;
  for (const diagnostic of diagnostics) {
    console.error(formatDiagnostic(diagnostic, options.warningsAsErrors));
  }
  if (errorCount > 0 || (options.warningsAsErrors && warningCount > 0)) {
    console.error(
      `frontend import boundary check failed: ${errorCount} error(s), ${warningCount} warning(s), ${files.length} file(s) checked`,
    );
    process.exit(1);
  }
  if (warningCount > 0) {
    console.error(
      `frontend import boundary check completed with ${warningCount} warning(s), ${files.length} file(s) checked`,
    );
    return;
  }
  console.log(`frontend import boundaries verified: ${files.length} file(s) checked`);
}

try {
  main();
} catch (error) {
  console.error(`frontend import boundary check failed: ${error instanceof Error ? error.message : String(error)}`);
  process.exit(1);
}
