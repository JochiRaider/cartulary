#!/usr/bin/env node
import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "..");
const manifestSchemaID = "cartulary.backend_module_boundaries.v1";
const summarySchemaID = "cartulary.backend_module_boundary_summary.v1";
const sourceExtensions = new Set([".go", ".mjs", ".js", ".sh"]);
const ignoredDirectoryNames = new Set([
  ".cache",
  ".git",
  "coverage",
  "dist",
  "node_modules",
  "playwright-report",
  "test-results",
  "tmp",
]);

function usage() {
  throw new Error("usage: check-backend-module-boundaries.mjs [--manifest <path>] [--root <path>]");
}

function parseArgs(argv) {
  const options = {
    manifest:
      process.env.BACKEND_MODULE_BOUNDARIES_MANIFEST ??
      "tools/backend_module_boundaries.json",
    root: process.env.CARTULARY_REPO_ROOT ?? defaultRepoRoot,
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--manifest") {
      options.manifest = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--root") {
      options.root = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.manifest || !options.root) {
    usage();
  }
  options.root = path.resolve(options.root);
  options.manifest = resolvePath(options.root, options.manifest);
  return options;
}

function resolvePath(root, value) {
  return path.isAbsolute(value) ? value : path.join(root, value);
}

function normalizePath(value) {
  return value.split(path.sep).join("/");
}

function repoRelative(root, value) {
  return normalizePath(path.relative(root, value));
}

function sha256(value) {
  return `sha256:${createHash("sha256").update(value).digest("hex")}`;
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

function requireArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value;
}

function normalizeManifest(raw) {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw new Error("manifest must be an object");
  }
  if (raw.schema_id !== manifestSchemaID) {
    throw new Error(`manifest must declare schema_id=${manifestSchemaID}`);
  }
  return {
    scanRoots: requireStringArray(raw.scan_roots, "scan_roots").map(normalizePath),
    scanExcludes: requireStringArray(raw.scan_excludes ?? [], "scan_excludes").map(normalizePath),
    ownerPortOnlyImports: requireArray(
      raw.owner_port_only_imports ?? [],
      "owner_port_only_imports",
    ).map((rule, index) => ({
      id: requireString(rule?.id, `owner_port_only_imports[${index + 1}].id`),
      import: requireString(rule?.import, `owner_port_only_imports[${index + 1}].import`),
      allowedImporters: requireStringArray(
        rule?.allowed_importers ?? [],
        `owner_port_only_imports[${index + 1}].allowed_importers`,
      ).map(normalizePath),
    })),
    rawNDJSONTargets: requireArray(raw.raw_ndjson_targets ?? [], "raw_ndjson_targets").map(
      (rule, index) => ({
        id: requireString(rule?.id, `raw_ndjson_targets[${index + 1}].id`),
        functionName: requireString(
          rule?.function,
          `raw_ndjson_targets[${index + 1}].function`,
        ),
        allowedPaths: requireStringArray(
          rule?.allowed_paths ?? [],
          `raw_ndjson_targets[${index + 1}].allowed_paths`,
        ).map(normalizePath),
      }),
    ),
    forbiddenRouteDependencies: requireArray(
      raw.forbidden_route_dependencies ?? [],
      "forbidden_route_dependencies",
    ).map((rule, index) => ({
      id: requireString(rule?.id, `forbidden_route_dependencies[${index + 1}].id`),
      importSuffix: requireString(
        rule?.import_suffix,
        `forbidden_route_dependencies[${index + 1}].import_suffix`,
      ),
      allowedImporters: requireStringArray(
        rule?.allowed_importers ?? [],
        `forbidden_route_dependencies[${index + 1}].allowed_importers`,
      ).map(normalizePath),
    })),
    generatedRootWrites: {
      scanRoots: requireStringArray(
        raw.generated_root_writes?.scan_roots ?? [],
        "generated_root_writes.scan_roots",
      ).map(normalizePath),
      generatedRoots: requireStringArray(
        raw.generated_root_writes?.generated_roots ?? [],
        "generated_root_writes.generated_roots",
      ).map(normalizePath),
      writeSymbols: requireStringArray(
        raw.generated_root_writes?.write_symbols ?? [],
        "generated_root_writes.write_symbols",
      ),
      allowedPaths: requireStringArray(
        raw.generated_root_writes?.allowed_paths ?? [],
        "generated_root_writes.allowed_paths",
      ).map(normalizePath),
    },
  };
}

function collectFiles(root, roots, excludes) {
  const files = [];
  for (const scanRoot of roots) {
    const absolute = resolvePath(root, scanRoot);
    if (!existsSync(absolute)) {
      continue;
    }
    walk(root, absolute, excludes, files);
  }
  return files.sort((left, right) => left.relative.localeCompare(right.relative));
}

function walk(root, directory, excludes, files) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const absolute = path.join(directory, entry.name);
    const relative = repoRelative(root, absolute);
    if (entry.isDirectory()) {
      if (ignoredDirectoryNames.has(entry.name) || isExcluded(relative, excludes)) {
        continue;
      }
      walk(root, absolute, excludes, files);
      continue;
    }
    if (!entry.isFile() || !sourceExtensions.has(path.extname(entry.name))) {
      continue;
    }
    if (isExcluded(relative, excludes)) {
      continue;
    }
    files.push({ absolute, relative, content: readFileSync(absolute, "utf8") });
  }
}

function isExcluded(relative, patterns) {
  return patterns.some((pattern) => matchesPattern(relative, pattern));
}

function matchesPattern(relative, pattern) {
  const normalized = normalizePath(pattern);
  if (normalized.endsWith("/**")) {
    return relative === normalized.slice(0, -3) || relative.startsWith(normalized.slice(0, -2));
  }
  if (normalized.endsWith("/")) {
    return relative.startsWith(normalized);
  }
  return relative === normalized || relative.startsWith(`${normalized}/`);
}

function extractGoImports(content) {
  const imports = [];
  const block = content.match(/import\s*\(([\s\S]*?)\)/m);
  if (block) {
    const quoted = block[1].matchAll(/"([^"]+)"/g);
    for (const match of quoted) {
      imports.push(match[1]);
    }
  }
  for (const match of content.matchAll(/import\s+"([^"]+)"/g)) {
    imports.push(match[1]);
  }
  return imports;
}

function violation(code, file, symbolOrImport) {
  return {
    code,
    path: file.relative,
    symbol_or_import: symbolOrImport,
    result: "fail",
  };
}

function checkOwnerPortOnlyImports(files, rules) {
  const violations = [];
  for (const file of files.filter((entry) => entry.relative.endsWith(".go"))) {
    for (const imported of extractGoImports(file.content)) {
      for (const rule of rules) {
        if (imported !== rule.import && !imported.startsWith(`${rule.import}/`)) {
          continue;
        }
        if (!rule.allowedImporters.some((pattern) => matchesPattern(file.relative, pattern))) {
          violations.push(violation("owner_port_only_import", file, imported));
        }
      }
    }
  }
  return violations;
}

function checkForbiddenRouteDependencies(files, rules) {
  const violations = [];
  for (const file of files.filter((entry) => entry.relative.endsWith(".go"))) {
    for (const imported of extractGoImports(file.content)) {
      for (const rule of rules) {
        if (!imported.endsWith(rule.importSuffix)) {
          continue;
        }
        if (!rule.allowedImporters.some((pattern) => matchesPattern(file.relative, pattern))) {
          violations.push(violation("forbidden_route_dependency", file, imported));
        }
      }
    }
  }
  return violations;
}

function checkRawNDJSONTargets(files, rules) {
  const violations = [];
  for (const file of files.filter((entry) => entry.relative.endsWith(".go"))) {
    for (const rule of rules) {
      if (rule.allowedPaths.some((pattern) => matchesPattern(file.relative, pattern))) {
        continue;
      }
      const escaped = rule.functionName.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
      const rawTargetPattern = new RegExp(`${escaped}\\s*\\([^\\n]*"[^"]+"`, "m");
      if (rawTargetPattern.test(file.content)) {
        violations.push(violation("raw_ndjson_target", file, rule.functionName));
      }
    }
  }
  return violations;
}

function checkGeneratedRootWrites(files, rule) {
  const scanFiles = files.filter((file) =>
    rule.scanRoots.some((scanRoot) => matchesPattern(file.relative, `${scanRoot}/**`)),
  );
  const violations = [];
  for (const file of scanFiles) {
    if (rule.allowedPaths.some((pattern) => matchesPattern(file.relative, pattern))) {
      continue;
    }
    const mentionsGeneratedRoot = rule.generatedRoots.some((root) => file.content.includes(root));
    const mentionsWriteSymbol = rule.writeSymbols.some((symbol) => file.content.includes(symbol));
    if (mentionsGeneratedRoot && mentionsWriteSymbol) {
      violations.push(violation("generated_root_write", file, "generated_root_write"));
    }
  }
  return violations;
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const manifestRawContent = readFileSync(options.manifest, "utf8");
  const manifest = normalizeManifest(JSON.parse(manifestRawContent));
  const files = collectFiles(options.root, manifest.scanRoots, manifest.scanExcludes);
  const violations = [
    ...checkOwnerPortOnlyImports(files, manifest.ownerPortOnlyImports),
    ...checkRawNDJSONTargets(files, manifest.rawNDJSONTargets),
    ...checkForbiddenRouteDependencies(files, manifest.forbiddenRouteDependencies),
    ...checkGeneratedRootWrites(files, manifest.generatedRootWrites),
  ].sort((left, right) => {
    return (
      left.code.localeCompare(right.code) ||
      left.path.localeCompare(right.path) ||
      left.symbol_or_import.localeCompare(right.symbol_or_import)
    );
  });
  const summary = {
    schema_id: summarySchemaID,
    checked_at: new Date().toISOString(),
    repo_root: options.root,
    manifest_digest: sha256(manifestRawContent),
    violations,
    result: violations.length === 0 ? "pass" : "fail",
  };
  console.log(JSON.stringify(summary));
  if (violations.length > 0) {
    process.exit(1);
  }
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`backend module boundary check failed: ${message}`);
  process.exit(1);
}
