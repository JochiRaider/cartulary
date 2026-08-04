#!/usr/bin/env node

import {
  existsSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import process from "node:process";
import { spawnSync } from "node:child_process";
import { createRequire } from "node:module";
import { fileURLToPath } from "node:url";

const require = createRequire(import.meta.url);
const ts = require("typescript");

const defaultRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../..",
);
const packageRootPath = "packages/protocol-ts";
const generatedRootPath = `${packageRootPath}/src/generated/`;
const defaultEntrypointOwnerPath =
  "contracts/protocol-ts/frontend-entrypoints.v2.json";

function parseArguments(argv) {
  const options = {
    entrypointOwner: defaultEntrypointOwnerPath,
    fallowBin: null,
    output: null,
    root: defaultRoot,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    const value = argv[index + 1];
    if (
      argument === "--root" ||
      argument === "--fallow-bin" ||
      argument === "--output" ||
      argument === "--entrypoint-owner"
    ) {
      if (!value) {
        throw new Error(`${argument} requires a value`);
      }
      if (argument === "--root") {
        options.root = path.resolve(value);
      } else if (argument === "--fallow-bin") {
        options.fallowBin = path.resolve(value);
      } else if (argument === "--output") {
        options.output = path.resolve(value);
      } else {
        options.entrypointOwner = value;
      }
      index += 1;
      continue;
    }
    throw new Error(`unexpected argument ${argument}`);
  }
  options.fallowBin ??= path.join(
    options.root,
    "node_modules/fallow/bin/fallow",
  );
  return options;
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function repoPath(root, relativePath) {
  return path.join(root, relativePath);
}

function repoRelative(root, file) {
  return path.relative(root, file).replaceAll("\\", "/");
}

function collectFiles(root, predicate) {
  const files = [];
  function visit(directory) {
    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      const entryPath = path.join(directory, entry.name);
      if (entry.isDirectory()) {
        visit(entryPath);
      } else if (entry.isFile() && predicate(entryPath)) {
        files.push(repoRelative(root, entryPath));
      }
    }
  }
  visit(repoPath(root, `${packageRootPath}/src`));
  return files.sort((left, right) => left.localeCompare(right));
}

function resolveSourceModule(root, sourcePath, moduleSpecifier) {
  if (!moduleSpecifier.startsWith(".")) {
    return null;
  }
  const unresolved = path.resolve(
    path.dirname(repoPath(root, sourcePath)),
    moduleSpecifier,
  );
  const candidates = [
    unresolved,
    unresolved.replace(/\.js$/u, ".ts"),
    unresolved.replace(/\.js$/u, ".tsx"),
    `${unresolved}.ts`,
    `${unresolved}.tsx`,
    path.join(unresolved, "index.ts"),
    path.join(unresolved, "index.tsx"),
  ];
  const resolved = candidates.find((candidate) => existsSync(candidate));
  return resolved ? repoRelative(root, resolved) : null;
}

function declarationIsTypeOnly(statement) {
  return (
    ts.isInterfaceDeclaration(statement) ||
    ts.isTypeAliasDeclaration(statement)
  );
}

function hasExportModifier(statement) {
  return Boolean(
    statement.modifiers?.some(
      (modifier) => modifier.kind === ts.SyntaxKind.ExportKeyword,
    ),
  );
}

function collectBindingNames(name, names = []) {
  if (ts.isIdentifier(name)) {
    names.push(name.text);
    return names;
  }
  for (const element of name.elements ?? []) {
    if (!ts.isOmittedExpression(element)) {
      collectBindingNames(element.name, names);
    }
  }
  return names;
}

function collectModuleExports(root, sourcePath, cache = new Map(), active = new Set()) {
  if (cache.has(sourcePath)) {
    return cache.get(sourcePath);
  }
  if (active.has(sourcePath) || !existsSync(repoPath(root, sourcePath))) {
    return new Map();
  }
  active.add(sourcePath);
  const source = readFileSync(repoPath(root, sourcePath), "utf8");
  const sourceFile = ts.createSourceFile(
    sourcePath,
    source,
    ts.ScriptTarget.Latest,
    true,
    sourcePath.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
  );
  const exports = new Map();
  const importedBindings = new Map();
  for (const statement of sourceFile.statements) {
    if (!ts.isImportDeclaration(statement)) {
      continue;
    }
    const targetPath = resolveSourceModule(
      root,
      sourcePath,
      statement.moduleSpecifier.text,
    );
    const generated = targetPath?.startsWith(generatedRootPath) === true;
    const clause = statement.importClause;
    if (clause?.name) {
      importedBindings.set(clause.name.text, {
        generated,
        typeOnly: clause.isTypeOnly,
      });
    }
    if (clause?.namedBindings && ts.isNamedImports(clause.namedBindings)) {
      for (const element of clause.namedBindings.elements) {
        importedBindings.set(element.name.text, {
          generated,
          typeOnly: clause.isTypeOnly || element.isTypeOnly,
        });
      }
    }
  }
  for (const statement of sourceFile.statements) {
    if (ts.isExportDeclaration(statement)) {
      const moduleSpecifier = statement.moduleSpecifier?.text;
      const targetPath = moduleSpecifier
        ? resolveSourceModule(root, sourcePath, moduleSpecifier)
        : null;
      const targetExports = targetPath
        ? collectModuleExports(root, targetPath, cache, active)
        : new Map();
      if (!statement.exportClause) {
        for (const [name, metadata] of targetExports) {
          exports.set(name, {
            ...metadata,
            generated: targetPath?.startsWith(generatedRootPath) ||
              metadata.generated === true,
            typeOnly: statement.isTypeOnly || metadata.typeOnly === true,
          });
        }
        continue;
      }
      if (!ts.isNamedExports(statement.exportClause)) {
        continue;
      }
      for (const element of statement.exportClause.elements) {
        const importedName = element.propertyName?.text ?? element.name.text;
        const inherited = targetExports.get(importedName) ??
          importedBindings.get(importedName);
        exports.set(element.name.text, {
          generated: targetPath?.startsWith(generatedRootPath) ||
            inherited?.generated === true,
          typeOnly: statement.isTypeOnly || element.isTypeOnly ||
            inherited?.typeOnly === true,
        });
      }
      continue;
    }
    if (!hasExportModifier(statement)) {
      continue;
    }
    if (ts.isVariableStatement(statement)) {
      for (const declaration of statement.declarationList.declarations) {
        for (const name of collectBindingNames(declaration.name)) {
          exports.set(name, { generated: false, typeOnly: false });
        }
      }
      continue;
    }
    if (statement.name?.text) {
      exports.set(statement.name.text, {
        generated: false,
        typeOnly: declarationIsTypeOnly(statement),
      });
    }
  }
  active.delete(sourcePath);
  cache.set(sourcePath, exports);
  return exports;
}

function collectConsumerImports(root) {
  const imports = new Map();
  const roots = ["apps", "packages"]
    .map((directory) => repoPath(root, directory))
    .filter(existsSync);
  function record(specifier, name) {
    const names = imports.get(specifier) ?? new Set();
    names.add(name);
    imports.set(specifier, names);
  }
  function visit(directory) {
    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      const entryPath = path.join(directory, entry.name);
      const relativePath = repoRelative(root, entryPath);
      if (entry.isDirectory()) {
        if (
          relativePath === packageRootPath ||
          ["node_modules", "dist", "coverage", "test-results"].includes(entry.name)
        ) {
          continue;
        }
        visit(entryPath);
        continue;
      }
      if (!entry.isFile() || !/\.[cm]?[jt]sx?$/u.test(entry.name)) {
        continue;
      }
      const source = readFileSync(entryPath, "utf8");
      const sourceFile = ts.createSourceFile(
        relativePath,
        source,
        ts.ScriptTarget.Latest,
        true,
        entry.name.endsWith("x") ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
      );
      for (const statement of sourceFile.statements) {
        if (ts.isImportDeclaration(statement)) {
          const specifier = statement.moduleSpecifier.text;
          if (!specifier.startsWith("@cartulary/protocol-ts")) {
            continue;
          }
          const clause = statement.importClause;
          if (
            clause?.name ||
            (clause?.namedBindings && ts.isNamespaceImport(clause.namedBindings))
          ) {
            record(specifier, "*");
          }
          if (clause?.namedBindings && ts.isNamedImports(clause.namedBindings)) {
            for (const element of clause.namedBindings.elements) {
              record(specifier, element.propertyName?.text ?? element.name.text);
            }
          }
        } else if (
          ts.isExportDeclaration(statement) &&
          statement.moduleSpecifier?.text?.startsWith("@cartulary/protocol-ts")
        ) {
          const specifier = statement.moduleSpecifier.text;
          if (!statement.exportClause) {
            record(specifier, "*");
          } else if (ts.isNamedExports(statement.exportClause)) {
            for (const element of statement.exportClause.elements) {
              record(specifier, element.propertyName?.text ?? element.name.text);
            }
          }
        }
      }
    }
  }
  for (const directory of roots) {
    visit(directory);
  }
  return imports;
}

function entrypointUnusedExports(root, passThroughs) {
  const packageJSON = readJSON(repoPath(root, `${packageRootPath}/package.json`));
  const consumerImports = collectConsumerImports(root);
  const findings = { unused_exports: [], unused_types: [] };
  const cache = new Map();
  for (const [exportKey, target] of Object.entries(packageJSON.exports ?? {})) {
    if (typeof target !== "string") {
      continue;
    }
    const sourcePath = `${packageRootPath}/${target.replace(/^\.\//u, "")}`;
    if (sourcePath.startsWith(generatedRootPath)) {
      continue;
    }
    const specifier = exportKey === "."
      ? "@cartulary/protocol-ts"
      : `@cartulary/protocol-ts/${exportKey.replace(/^\.\//u, "")}`;
    const used = consumerImports.get(specifier) ?? new Set();
    if (used.has("*")) {
      continue;
    }
    const ownerExemptions = new Set(passThroughs.byFile.get(sourcePath) ?? []);
    for (const [name, metadata] of collectModuleExports(root, sourcePath, cache)) {
      if (used.has(name)) {
        continue;
      }
      if (
        metadata.generated &&
        (!passThroughs.ownerPresent || ownerExemptions.has(name))
      ) {
        continue;
      }
      const finding = { file: sourcePath, name, specifier };
      findings[metadata.typeOnly ? "unused_types" : "unused_exports"].push(
        finding,
      );
    }
  }
  return findings;
}

function packageExportEntries(root) {
  const packageJSON = readJSON(repoPath(root, `${packageRootPath}/package.json`));
  return Object.values(packageJSON.exports ?? {})
    .filter((target) => typeof target === "string")
    .map((target) =>
      `${packageRootPath}/${target.replace(/^\.\//u, "")}`,
    );
}

function normalizeGeneratedImport(root, sourcePath, importPath) {
  if (!importPath.startsWith(".")) {
    return null;
  }
  const resolved = repoRelative(
    root,
    path.resolve(path.dirname(repoPath(root, sourcePath)), importPath),
  ).replace(/\.js$/u, ".ts");
  return resolved.startsWith(generatedRootPath) ? resolved : null;
}

function explicitPassThroughExports(root, ownerPath) {
  const absoluteOwnerPath = repoPath(root, ownerPath);
  if (!existsSync(absoluteOwnerPath)) {
    return { byFile: new Map(), count: 0, ownerPresent: false };
  }
  const owner = readJSON(absoluteOwnerPath);
  const byFile = new Map();
  let count = 0;
  for (const entrypoint of owner.entrypoints ?? []) {
    const sourcePath = entrypoint.authored_path;
    const allowedModules = new Set(
      entrypoint.generated_module_allowlist ?? [],
    );
    if (typeof sourcePath !== "string" || !existsSync(repoPath(root, sourcePath))) {
      continue;
    }
    const source = readFileSync(repoPath(root, sourcePath), "utf8");
    const sourceFile = ts.createSourceFile(
      sourcePath,
      source,
      ts.ScriptTarget.Latest,
      true,
      sourcePath.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
    );
    const exports = new Set();
    for (const statement of sourceFile.statements) {
      if (
        !ts.isExportDeclaration(statement) ||
        !statement.moduleSpecifier ||
        !ts.isStringLiteral(statement.moduleSpecifier)
      ) {
        continue;
      }
      const generatedImport = normalizeGeneratedImport(
        root,
        sourcePath,
        statement.moduleSpecifier.text,
      );
      if (!generatedImport || !allowedModules.has(generatedImport)) {
        continue;
      }
      if (!statement.exportClause) {
        for (const name of collectModuleExports(root, generatedImport).keys()) {
          exports.add(name);
        }
        continue;
      }
      if (ts.isNamedExports(statement.exportClause)) {
        for (const element of statement.exportClause.elements) {
          exports.add(element.name.text);
        }
      }
    }
    if (exports.size > 0) {
      byFile.set(sourcePath, [...exports].sort());
      count += exports.size;
    }
  }
  return { byFile, count, ownerPresent: true };
}

function buildGateConfig(root, entrypointOwner) {
  const base = readJSON(repoPath(root, ".fallowrc.json"));
  const compileFixtures = collectFiles(root, (file) =>
    file.endsWith(".compile.ts") || file.endsWith(".compile.tsx"),
  );
  const passThroughs = explicitPassThroughExports(root, entrypointOwner);
  const ignoreExports = (base.ignoreExports ?? []).filter(
    (rule) => !String(rule?.file ?? "").startsWith(`${packageRootPath}/`),
  );
  for (const file of compileFixtures) {
    ignoreExports.push({ file, exports: ["*"] });
  }
  for (const [file, exports] of passThroughs.byFile) {
    ignoreExports.push({ file, exports });
  }
  const rules = Object.fromEntries(
    Object.keys(base.rules ?? {}).map((rule) => [rule, "off"]),
  );
  rules["unused-files"] = "error";
  rules["unused-exports"] = "error";
  rules["unused-types"] = "error";
  return {
    config: {
      ...base,
      entry: [
        ...(base.entry ?? []).filter(
          (entry) => entry !== `${packageRootPath}/src/index.ts`,
        ),
        ...packageExportEntries(root),
        ...compileFixtures,
      ],
      ignoreExports,
      includeEntryExports: true,
      publicPackages: (base.publicPackages ?? []).filter(
        (packageName) => packageName !== "@cartulary/protocol-ts",
      ),
      rules,
    },
    compileFixtures,
    passThroughs,
  };
}

function findingCount(report) {
  return ["unused_files", "unused_exports", "unused_types"].reduce(
    (count, key) => count + (Array.isArray(report[key]) ? report[key].length : 0),
    0,
  );
}

function findingLabels(report) {
  const labels = [];
  for (const key of ["unused_files", "unused_exports", "unused_types"]) {
    for (const finding of report[key] ?? []) {
      labels.push(`${key}:${JSON.stringify(finding)}`);
    }
  }
  return labels;
}

function run() {
  const options = parseArguments(process.argv.slice(2));
  if (!existsSync(options.fallowBin)) {
    throw new Error(`missing Fallow binary ${options.fallowBin}`);
  }
  for (const required of [
    ".fallowrc.json",
    `${packageRootPath}/package.json`,
    `${packageRootPath}/src`,
  ]) {
    if (!existsSync(repoPath(options.root, required))) {
      throw new Error(`missing required package-gate input ${required}`);
    }
  }

  const temporaryRoot = mkdtempSync(
    path.join(tmpdir(), "cartulary-protocol-ts-dead-code-"),
  );
  try {
    const { config, compileFixtures, passThroughs } = buildGateConfig(
      options.root,
      options.entrypointOwner,
    );
    const configPath = path.join(temporaryRoot, "fallowrc.json");
    writeFileSync(configPath, `${JSON.stringify(config, null, 2)}\n`);
    const result = spawnSync(
      process.execPath,
      [
        options.fallowBin,
        "dead-code",
        "--root",
        options.root,
        "--config",
        configPath,
        "--format",
        "json",
        "--quiet",
        "--no-cache",
        "--workspace",
        "@cartulary/protocol-ts",
        "--production",
        "--include-entry-exports",
        "--fail-on-issues",
      ],
      { cwd: options.root, encoding: "utf8" },
    );
    if (!result.stdout.trim()) {
      throw new Error(
        `Fallow produced no JSON report: ${result.stderr.trim() || `exit ${result.status}`}`,
      );
    }
    const report = JSON.parse(result.stdout);
    const entrypointFindings = entrypointUnusedExports(
      options.root,
      passThroughs,
    );
    report.unused_exports = [
      ...(report.unused_exports ?? []),
      ...entrypointFindings.unused_exports,
    ];
    report.unused_types = [
      ...(report.unused_types ?? []),
      ...entrypointFindings.unused_types,
    ];
    const authoredFindings = findingCount(report);
    const summary = {
      schema_id: "cartulary.protocol_ts_dead_code_summary.v1",
      package: "@cartulary/protocol-ts",
      authored_findings: authoredFindings,
      authored_suppressions: 0,
      automated_compile_fixture_exports: compileFixtures.length,
      automated_generated_pass_through_exports: passThroughs.count,
      entrypoint_owner_present: passThroughs.ownerPresent,
      findings: {
        unused_files: report.unused_files ?? [],
        unused_exports: report.unused_exports ?? [],
        unused_types: report.unused_types ?? [],
      },
    };
    const serialized = `${JSON.stringify(summary, null, 2)}\n`;
    if (options.output) {
      writeFileSync(options.output, serialized);
    }
    process.stdout.write(serialized);
    if (authoredFindings > 0) {
      process.stderr.write(
        `protocol-ts dead-code gate found ${authoredFindings} authored finding(s)\n`,
      );
      for (const label of findingLabels(report).slice(0, 25)) {
        process.stderr.write(`- ${label}\n`);
      }
      return 1;
    }
    return 0;
  } finally {
    rmSync(temporaryRoot, { force: true, recursive: true });
  }
}

try {
  process.exitCode = run();
} catch (error) {
  process.stderr.write(`protocol-ts dead-code check failed: ${error.message}\n`);
  process.exitCode = 2;
}
