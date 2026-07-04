#!/usr/bin/env node
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { builtinModules } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";

import ts from "typescript";

import { loadDesignTokenDocument } from "./contracts/frontend-design.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "../../..");
const schemaID = "cartulary.frontend_import_boundaries.v2";
const sourceExtensions = new Set([".ts", ".tsx", ".mts", ".cts"]);
const knownNodeBuiltins = new Set(
  builtinModules.flatMap((moduleName) => {
    const withoutProtocol = moduleName.startsWith("node:")
      ? moduleName.slice("node:".length)
      : moduleName;
    return [moduleName, withoutProtocol, withoutProtocol.split("/")[0]];
  }),
);
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

function requirePositiveInteger(value, label) {
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} must be a positive integer`);
  }
  return value;
}

function requireOptionalStringArray(value, label) {
  if (value === undefined) {
    return [];
  }
  return requireStringArray(value, label);
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
    const appliesTo = normalizeAppliesTo(rule?.applies_to, `${label}.applies_to`);
    const normalized = {
      id: requireString(rule?.id, `${label}.id`),
      level: requireString(rule.level, `${label}.level`),
      message: requireString(rule.message, `${label}.message`),
      appliesTo,
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
      if (kind === "node_builtin") {
        normalized.restrictedImports.push({
          kind,
          names: requireOptionalStringArray(restriction.names, `${restrictionLabel}.names`),
        });
        continue;
      }
      if (kind === "workspace_package_facade") {
        normalized.restrictedImports.push({
          kind,
          packageRoots: requireStringArray(
            restriction.package_roots,
            `${restrictionLabel}.package_roots`,
          ).map(normalizePath),
          workspacePackages: [],
        });
        continue;
      }
      throw new Error(
        `${restrictionLabel}.kind must be package, path_prefix, node_builtin, or workspace_package_facade`,
      );
    }
    rules.push(normalized);
  }
  const singletonImports = [];
  for (const [index, entry] of requireArray(
    raw.singleton_imports ?? [],
    "singleton_imports",
  ).entries()) {
    const label = `singleton_imports[${index + 1}]`;
    const normalized = {
      allowedImporters: (entry.allowed_importers ?? []).map((importer, importerIndex) =>
        requireString(importer, `${label}.allowed_importers[${importerIndex + 1}]`),
      ),
      id: requireString(entry?.id, `${label}.id`),
      level: requireString(entry.level, `${label}.level`),
      message: requireString(entry.message, `${label}.message`),
      requiredCount: requirePositiveInteger(
        entry.required_count,
        `${label}.required_count`,
      ),
      specifier: requireString(entry.specifier, `${label}.specifier`),
    };
    if (!["error", "warning"].includes(normalized.level)) {
      throw new Error(`${label}.level must be error or warning`);
    }
    singletonImports.push(normalized);
  }
  return {
    scanRoots: requireStringArray(raw.scan_roots, "scan_roots"),
    scanExcludes: (raw.scan_excludes ?? []).map((entry, index) =>
      requireString(entry, `scan_excludes[${index + 1}]`),
    ),
    rules,
    singletonImports,
    rawDesignTokenLiteralChecks: normalizeRawDesignTokenLiteralChecks(
      raw.raw_design_token_literal_checks ?? [],
    ),
  };
}

function normalizeRawDesignTokenLiteralChecks(checks) {
  return requireArray(checks, "raw_design_token_literal_checks").map((check, index) => {
    const label = `raw_design_token_literal_checks[${index + 1}]`;
    const normalized = {
      appliesTo: normalizeAppliesTo(check?.applies_to, `${label}.applies_to`),
      designDocument: requireString(check?.design_document, `${label}.design_document`),
      id: requireString(check?.id, `${label}.id`),
      level: requireString(check?.level, `${label}.level`),
      message: requireString(check?.message, `${label}.message`),
      tokenNamespaces: new Set(
        requireStringArray(check?.token_namespaces, `${label}.token_namespaces`),
      ),
    };
    if (!["error", "warning"].includes(normalized.level)) {
      throw new Error(`${label}.level must be error or warning`);
    }
    return normalized;
  });
}

function normalizeAppliesTo(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  const include = requireStringArray(value.include, `${label}.include`);
  if (include.length === 0) {
    throw new Error(`${label}.include must not be empty`);
  }
  return {
    exclude: requireOptionalStringArray(value.exclude, `${label}.exclude`),
    include,
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

function isRuleApplicable(rule, importerPath) {
  return (
    rule.appliesTo.include.some((pattern) => matchesGlob(pattern, importerPath)) &&
    !rule.appliesTo.exclude.some((pattern) => matchesGlob(pattern, importerPath))
  );
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

function parsePackageSpecifier(specifier) {
  if (
    specifier.startsWith(".") ||
    specifier.startsWith("/") ||
    specifier.startsWith("node:")
  ) {
    return null;
  }
  const parts = specifier.split("/");
  if (specifier.startsWith("@")) {
    if (parts.length < 2) {
      return null;
    }
    return {
      name: `${parts[0]}/${parts[1]}`,
      subpath: parts.slice(2).join("/"),
    };
  }
  return {
    name: parts[0],
    subpath: parts.slice(1).join("/"),
  };
}

function normalizeExportKeys(exportsValue) {
  if (exportsValue === undefined) {
    return new Set(["."]);
  }
  if (typeof exportsValue === "string" || Array.isArray(exportsValue)) {
    return new Set(["."]);
  }
  if (!exportsValue || typeof exportsValue !== "object") {
    return new Set();
  }
  const keys = Object.keys(exportsValue);
  if (keys.some((key) => key === "." || key.startsWith("./"))) {
    return new Set(keys.filter((key) => key === "." || key.startsWith("./")));
  }
  return new Set(["."]);
}

function exportKeyForSubpath(subpath) {
  return subpath === "" ? "." : `./${subpath}`;
}

function matchesExportPattern(pattern, exportKey) {
  if (!pattern.includes("*")) {
    return pattern === exportKey;
  }
  const source = `^${pattern
    .split("*")
    .map((part) => part.replace(/[.+^${}()|[\]\\]/g, "\\$&"))
    .join(".*")}$`;
  return new RegExp(source).test(exportKey);
}

function isDeclaredPackageExport(workspacePackage, subpath) {
  const exportKey = exportKeyForSubpath(subpath);
  return [...workspacePackage.exportKeys].some((key) =>
    matchesExportPattern(key, exportKey),
  );
}

function hydrateWorkspacePackageRestrictions(root, config) {
  const cache = new Map();
  const readWorkspacePackages = (packageRoots) =>
    packageRoots.map((packageRoot) => {
      if (cache.has(packageRoot)) {
        return cache.get(packageRoot);
      }
      const packageJSONPath = path.join(root, packageRoot, "package.json");
      if (!existsSync(packageJSONPath)) {
        throw new Error(`${packageRoot}/package.json is missing`);
      }
      const manifest = JSON.parse(readFileSync(packageJSONPath, "utf8"));
      const workspacePackage = {
        exportKeys: normalizeExportKeys(manifest.exports),
        name: requireString(manifest.name, `${packageRoot}/package.json.name`),
        root: packageRoot,
        sourceRoot: `${packageRoot}/src`,
      };
      cache.set(packageRoot, workspacePackage);
      return workspacePackage;
    });

  for (const rule of config.rules) {
    for (const restriction of rule.restrictedImports) {
      if (restriction.kind === "workspace_package_facade") {
        restriction.workspacePackages = readWorkspacePackages(restriction.packageRoots);
      }
    }
  }
  return config;
}

function isPathPrefixRestricted(root, importerFile, restriction, specifier) {
  const resolved = resolvedRelativeImport(root, importerFile, specifier);
  return resolved === restriction.path || resolved.startsWith(`${restriction.path}/`);
}

function isNodeBuiltinRestricted(restriction, specifier) {
  const withoutProtocol = specifier.startsWith("node:")
    ? specifier.slice("node:".length)
    : specifier;
  const rootName = withoutProtocol.split("/")[0];
  if (
    !knownNodeBuiltins.has(specifier) &&
    !knownNodeBuiltins.has(withoutProtocol) &&
    !knownNodeBuiltins.has(rootName)
  ) {
    return false;
  }
  if (restriction.names.length === 0 || restriction.names.includes("*")) {
    return true;
  }
  return restriction.names.some((name) => {
    const normalizedName = name.startsWith("node:")
      ? name.slice("node:".length)
      : name;
    return normalizedName === withoutProtocol || normalizedName === rootName;
  });
}

function workspacePackageForPath(workspacePackages, relativePath) {
  return workspacePackages.find(
    (workspacePackage) =>
      relativePath === workspacePackage.root ||
      relativePath.startsWith(`${workspacePackage.root}/`),
  );
}

function isWorkspacePackageFacadeRestricted(root, importerFile, restriction, specifier) {
  const parsedPackage = parsePackageSpecifier(specifier);
  if (parsedPackage !== null) {
    const workspacePackage = restriction.workspacePackages.find(
      (candidate) => candidate.name === parsedPackage.name,
    );
    return (
      workspacePackage !== undefined &&
      !isDeclaredPackageExport(workspacePackage, parsedPackage.subpath)
    );
  }

  const resolved = resolvedRelativeImport(root, importerFile, specifier);
  if (!resolved) {
    return false;
  }
  const importerPath = repoRelative(root, importerFile);
  const importerPackage = workspacePackageForPath(
    restriction.workspacePackages,
    importerPath,
  );
  const targetPackage = workspacePackageForPath(
    restriction.workspacePackages,
    resolved,
  );
  return (
    targetPackage !== undefined &&
    targetPackage !== importerPackage &&
    (resolved === targetPackage.sourceRoot ||
      resolved.startsWith(`${targetPackage.sourceRoot}/`))
  );
}

function matchesRestriction(root, importerFile, restriction, specifier) {
  if (restriction.kind === "package") {
    return isPackageRestricted(restriction, specifier);
  }
  if (restriction.kind === "path_prefix") {
    return isPathPrefixRestricted(root, importerFile, restriction, specifier);
  }
  if (restriction.kind === "node_builtin") {
    return isNodeBuiltinRestricted(restriction, specifier);
  }
  if (restriction.kind === "workspace_package_facade") {
    return isWorkspacePackageFacadeRestricted(root, importerFile, restriction, specifier);
  }
  return false;
}

function fileImports(file) {
  const content = readFileSync(file, "utf8");
  const sourceFile = ts.createSourceFile(
    file,
    content,
    ts.ScriptTarget.Latest,
    true,
    scriptKindFor(file),
  );
  return collectImports(sourceFile);
}

function evaluateFile(root, config, file, imports) {
  const relativeFile = repoRelative(root, file);
  const diagnostics = [];
  for (const imported of imports) {
    for (const rule of config.rules) {
      if (!isRuleApplicable(rule, relativeFile) || isAllowedImporter(rule, relativeFile)) {
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

function evaluateSingletonImports(root, config, fileImportEntries) {
  const diagnostics = [];
  for (const singletonImport of config.singletonImports) {
    const matches = fileImportEntries.flatMap(({ file, imports }) => {
      const relativeFile = repoRelative(root, file);
      return imports
        .filter((imported) => imported.specifier === singletonImport.specifier)
        .map((imported) => ({
          ...imported,
          file: relativeFile,
        }));
    });

    for (const imported of matches) {
      const allowed = singletonImport.allowedImporters.some((pattern) =>
        matchesGlob(pattern, imported.file),
      );
      if (!allowed) {
        diagnostics.push({
          column: imported.column,
          file: imported.file,
          importKind: imported.kind,
          level: singletonImport.level,
          line: imported.line,
          message: singletonImport.message,
          ruleID: singletonImport.id,
          specifier: imported.specifier,
        });
      }
    }

    if (matches.length !== singletonImport.requiredCount) {
      diagnostics.push({
        column: 1,
        file: "tools/frontend_import_boundaries.json",
        importKind: "singleton import",
        level: singletonImport.level,
        line: 1,
        message: `${singletonImport.message}; expected exactly ${singletonImport.requiredCount}, found ${matches.length}`,
        ruleID: singletonImport.id,
        specifier: singletonImport.specifier,
      });
    }
  }
  return diagnostics;
}

function evaluateRawDesignTokenLiterals(root, config, files) {
  const diagnostics = [];
  const documentCache = new Map();
  const tokenValuesForCheck = (check) => {
    const designPath = resolvePath(root, check.designDocument);
    if (!documentCache.has(designPath)) {
      documentCache.set(designPath, loadDesignTokenDocument(designPath));
    }
    const document = documentCache.get(designPath);
    return new Set(
      [...document.tokenMap.values()]
        .filter(
          (entry) =>
            check.tokenNamespaces.has(entry.namespace) &&
            typeof entry.raw === "string" &&
            entry.raw !== "transparent",
        )
        .map((entry) => entry.raw),
    );
  };

  for (const check of config.rawDesignTokenLiteralChecks) {
    const forbiddenValues = tokenValuesForCheck(check);
    for (const file of files) {
      const relativeFile = repoRelative(root, file);
      if (!isRuleApplicable(check, relativeFile)) {
        continue;
      }
      const content = readFileSync(file, "utf8");
      for (const value of forbiddenValues) {
        const index = content.indexOf(value);
        if (index < 0) {
          continue;
        }
        const { column, line } = lineColumnForIndex(content, index);
        diagnostics.push({
          column,
          file: relativeFile,
          importKind: "raw design token literal",
          level: check.level,
          line,
          message: `${check.message} Found ${value}.`,
          ruleID: check.id,
          specifier: value,
        });
      }
    }
  }
  return diagnostics;
}

function lineColumnForIndex(content, index) {
  const prefix = content.slice(0, index);
  const lines = prefix.split(/\r?\n/);
  return {
    column: lines[lines.length - 1].length + 1,
    line: lines.length,
  };
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
  const config = hydrateWorkspacePackageRestrictions(
    root,
    normalizeConfig(readJSON(root, options.config)),
  );
  const files = collectSourceFiles(root, config);
  const fileImportEntries = files.map((file) => ({
    file,
    imports: fileImports(file),
  }));
  const diagnostics = [
    ...fileImportEntries.flatMap(({ file, imports }) =>
      evaluateFile(root, config, file, imports),
    ),
    ...evaluateSingletonImports(root, config, fileImportEntries),
    ...evaluateRawDesignTokenLiterals(root, config, files),
  ];
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
    emitFrontendTargetSummary(root, "pass");
    return;
  }
  console.log(`frontend import boundaries verified: ${files.length} file(s) checked`);
  emitFrontendTargetSummary(root, "pass");
}

try {
  main();
} catch (error) {
  console.error(`frontend import boundary check failed: ${error instanceof Error ? error.message : String(error)}`);
  process.exit(1);
}
