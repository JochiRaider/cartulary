#!/usr/bin/env node
import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  backendRuntimeExcludePatterns,
  readSupportInventory,
} from "./support-inventory-profiles.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "../../..");
const manifestSchemaID = "cartulary.backend_module_boundaries.v2";
const summarySchemaID = "cartulary.backend_module_boundary_summary.v1";
const sourceExtensions = new Set([".go", ".mjs", ".js", ".sh", ".sql"]);
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
  throw new Error(
    "usage: check-backend-module-boundaries.mjs [--manifest <path>] [--support-inventory <path>] [--root <path>]",
  );
}

function parseArgs(argv) {
  const options = {
    manifest:
      process.env.BACKEND_MODULE_BOUNDARIES_MANIFEST ??
      "tools/backend_module_boundaries.json",
    supportInventory:
      process.env.TEST_SUPPORT_INVENTORY ??
      process.env.CARTULARY_TEST_SUPPORT_INVENTORY ??
      path.join(defaultRepoRoot, "tools/test_support_inventory.json"),
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
    if (arg === "--support-inventory") {
      options.supportInventory = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.manifest || !options.root || !options.supportInventory) {
    usage();
  }
  options.root = path.resolve(options.root);
  options.manifest = resolvePath(options.root, options.manifest);
  options.supportInventory = resolvePath(options.root, options.supportInventory);
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

function requireBoolean(value, label, defaultValue = false) {
  if (value === undefined) {
    return defaultValue;
  }
  if (typeof value !== "boolean") {
    throw new Error(`${label} must be a boolean`);
  }
  return value;
}

function requireEnum(value, allowed, label) {
  if (typeof value !== "string" || !allowed.has(value)) {
    throw new Error(`${label} must be one of ${[...allowed].join(", ")}`);
  }
  return value;
}

function requireArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value;
}

function requireNonnegativeInteger(value, label) {
  if (!Number.isInteger(value) || value < 0) {
    throw new Error(`${label} must be a nonnegative integer`);
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
    migrationSourceBoundary: {
      scanRoots: requireStringArray(
        raw.migration_source_boundary?.scan_roots ?? [],
        "migration_source_boundary.scan_roots",
      ).map(normalizePath),
      restrictedImports: requireArray(
        raw.migration_source_boundary?.restricted_imports ?? [],
        "migration_source_boundary.restricted_imports",
      ).map((rule, index) => ({
        id: requireString(rule?.id, `migration_source_boundary.restricted_imports[${index + 1}].id`),
        matchKind: requireEnum(
          rule?.match_kind,
          new Set(["exact", "subtree"]),
          `migration_source_boundary.restricted_imports[${index + 1}].match_kind`,
        ),
        imports: requireStringArray(
          rule?.imports ?? [],
          `migration_source_boundary.restricted_imports[${index + 1}].imports`,
        ),
        scanPaths: requireStringArray(
          rule?.scan_paths ?? [],
          `migration_source_boundary.restricted_imports[${index + 1}].scan_paths`,
        ).map(normalizePath),
        allowedPaths: requireStringArray(
          rule?.allowed_paths ?? [],
          `migration_source_boundary.restricted_imports[${index + 1}].allowed_paths`,
        ).map(normalizePath),
        productionOnly: requireBoolean(
          rule?.production_only,
          `migration_source_boundary.restricted_imports[${index + 1}].production_only`,
        ),
      })),
      forbiddenDeclarations: requireArray(
        raw.migration_source_boundary?.forbidden_declarations ?? [],
        "migration_source_boundary.forbidden_declarations",
      ).map((rule, index) => ({
        id: requireString(rule?.id, `migration_source_boundary.forbidden_declarations[${index + 1}].id`),
        symbols: requireStringArray(
          rule?.symbols ?? [],
          `migration_source_boundary.forbidden_declarations[${index + 1}].symbols`,
        ),
        scanPaths: requireStringArray(
          rule?.scan_paths ?? [],
          `migration_source_boundary.forbidden_declarations[${index + 1}].scan_paths`,
        ).map(normalizePath),
        allowedPaths: requireStringArray(
          rule?.allowed_paths ?? [],
          `migration_source_boundary.forbidden_declarations[${index + 1}].allowed_paths`,
        ).map(normalizePath),
        productionOnly: requireBoolean(
          rule?.production_only,
          `migration_source_boundary.forbidden_declarations[${index + 1}].production_only`,
        ),
      })),
      exactCalls: requireArray(
        raw.migration_source_boundary?.exact_calls ?? [],
        "migration_source_boundary.exact_calls",
      ).map((rule, index) => ({
        id: requireString(rule?.id, `migration_source_boundary.exact_calls[${index + 1}].id`),
        symbol: requireString(rule?.symbol, `migration_source_boundary.exact_calls[${index + 1}].symbol`),
        scanPaths: requireStringArray(
          rule?.scan_paths ?? [],
          `migration_source_boundary.exact_calls[${index + 1}].scan_paths`,
        ).map(normalizePath),
        allowedPaths: requireStringArray(
          rule?.allowed_paths ?? [],
          `migration_source_boundary.exact_calls[${index + 1}].allowed_paths`,
        ).map(normalizePath),
        requiredPath: normalizePath(requireString(
          rule?.required_path,
          `migration_source_boundary.exact_calls[${index + 1}].required_path`,
        )),
        requiredCount: requireNonnegativeInteger(
          rule?.required_count,
          `migration_source_boundary.exact_calls[${index + 1}].required_count`,
        ),
        productionOnly: requireBoolean(
          rule?.production_only,
          `migration_source_boundary.exact_calls[${index + 1}].production_only`,
        ),
      })),
    },
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
    forbiddenGoImports: requireArray(raw.forbidden_go_imports ?? [], "forbidden_go_imports").map(
      (rule, index) => ({
        id: requireString(rule?.id, `forbidden_go_imports[${index + 1}].id`),
        matchKind: requireEnum(
          rule?.match_kind,
          new Set(["exact", "subtree"]),
          `forbidden_go_imports[${index + 1}].match_kind`,
        ),
        imports: requireStringArray(
          rule?.imports ?? [],
          `forbidden_go_imports[${index + 1}].imports`,
        ),
        scanPaths: requireStringArray(
          rule?.scan_paths ?? [],
          `forbidden_go_imports[${index + 1}].scan_paths`,
        ).map(normalizePath),
        allowedPaths: requireStringArray(
          rule?.allowed_paths ?? [],
          `forbidden_go_imports[${index + 1}].allowed_paths`,
        ).map(normalizePath),
        productionOnly: requireBoolean(
          rule?.production_only,
          `forbidden_go_imports[${index + 1}].production_only`,
        ),
      }),
    ),
    goImportAllowlists: requireArray(raw.go_import_allowlists ?? [], "go_import_allowlists").map(
      (rule, index) => ({
        id: requireString(rule?.id, `go_import_allowlists[${index + 1}].id`),
        importPrefix: requireString(
          rule?.import_prefix,
          `go_import_allowlists[${index + 1}].import_prefix`,
        ),
        allowedImports: new Set(
          requireStringArray(
            rule?.allowed_imports ?? [],
            `go_import_allowlists[${index + 1}].allowed_imports`,
          ),
        ),
        allowedPrefixes: requireStringArray(
          rule?.allowed_prefixes ?? [],
          `go_import_allowlists[${index + 1}].allowed_prefixes`,
        ),
        scanPaths: requireStringArray(
          rule?.scan_paths ?? [],
          `go_import_allowlists[${index + 1}].scan_paths`,
        ).map(normalizePath),
        productionOnly: requireBoolean(
          rule?.production_only,
          `go_import_allowlists[${index + 1}].production_only`,
        ),
      }),
    ),
    sourceTableAccess: requireArray(raw.source_table_access ?? [], "source_table_access").map(
      (rule, index) => ({
        id: requireString(rule?.id, `source_table_access[${index + 1}].id`),
        tables: requireStringArray(
          rule?.tables ?? [],
          `source_table_access[${index + 1}].tables`,
        ),
        allowedPaths: requireStringArray(
          rule?.allowed_paths ?? [],
          `source_table_access[${index + 1}].allowed_paths`,
        ).map(normalizePath),
        scanPaths: requireStringArray(
          rule?.scan_paths ?? [],
          `source_table_access[${index + 1}].scan_paths`,
        ).map(normalizePath),
      }),
    ),
    forbiddenSourceMappings: requireArray(
      raw.forbidden_source_mappings ?? [],
      "forbidden_source_mappings",
    ).map((rule, index) => ({
      id: requireString(rule?.id, `forbidden_source_mappings[${index + 1}].id`),
      prefixes: requireStringArray(
        rule?.prefixes ?? [],
        `forbidden_source_mappings[${index + 1}].prefixes`,
      ),
      scanPaths: requireStringArray(
        rule?.scan_paths ?? [],
        `forbidden_source_mappings[${index + 1}].scan_paths`,
      ).map(normalizePath),
    })),
    sqlTableAllowlists: requireArray(
      raw.sql_table_allowlists ?? [],
      "sql_table_allowlists",
    ).map((rule, index) => ({
      id: requireString(rule?.id, `sql_table_allowlists[${index + 1}].id`),
      allowedTables: new Set(
        requireStringArray(
          rule?.allowed_tables ?? [],
          `sql_table_allowlists[${index + 1}].allowed_tables`,
        ),
      ),
      scanPaths: requireStringArray(
        rule?.scan_paths ?? [],
        `sql_table_allowlists[${index + 1}].scan_paths`,
        ).map(normalizePath),
      })),
    sqlTableAccess: requireArray(raw.sql_table_access ?? [], "sql_table_access").map(
      (rule, index) => ({
        id: requireString(rule?.id, `sql_table_access[${index + 1}].id`),
        table: requireString(rule?.table, `sql_table_access[${index + 1}].table`).toLowerCase(),
        scanPaths: requireStringArray(
          rule?.scan_paths ?? [],
          `sql_table_access[${index + 1}].scan_paths`,
        ).map(normalizePath),
        readAllowedPaths: requireStringArray(
          rule?.read_allowed_paths ?? [],
          `sql_table_access[${index + 1}].read_allowed_paths`,
        ).map(normalizePath),
        writeAllowedPaths: requireStringArray(
          rule?.write_allowed_paths ?? [],
          `sql_table_access[${index + 1}].write_allowed_paths`,
        ).map(normalizePath),
        testReadAllowedPaths: requireStringArray(
          rule?.test_read_allowed_paths ?? [],
          `sql_table_access[${index + 1}].test_read_allowed_paths`,
        ).map(normalizePath),
        testWriteAllowedPaths: requireStringArray(
          rule?.test_write_allowed_paths ?? [],
          `sql_table_access[${index + 1}].test_write_allowed_paths`,
        ).map(normalizePath),
        lockAllowedPaths: rule?.lock_allowed_paths === undefined
          ? null
          : requireStringArray(
            rule.lock_allowed_paths,
            `sql_table_access[${index + 1}].lock_allowed_paths`,
          ).map(normalizePath),
      }),
    ),
    forbiddenGoCalls: requireArray(raw.forbidden_go_calls ?? [], "forbidden_go_calls").map(
      (rule, index) => ({
        id: requireString(rule?.id, `forbidden_go_calls[${index + 1}].id`),
        symbols: requireStringArray(
          rule?.symbols ?? [],
          `forbidden_go_calls[${index + 1}].symbols`,
        ),
        allowedPaths: requireStringArray(
          rule?.allowed_paths ?? [],
          `forbidden_go_calls[${index + 1}].allowed_paths`,
        ).map(normalizePath),
      }),
    ),
    commandRootShape: {
      root: requireString(raw.command_root_shape?.root ?? "cmd", "command_root_shape.root"),
      allowedGoFiles: new Set(
        requireStringArray(
          raw.command_root_shape?.allowed_go_files ?? [],
          "command_root_shape.allowed_go_files",
        ).map(normalizePath),
      ),
    },
    forbiddenSourceTokens: requireArray(
      raw.forbidden_source_tokens ?? [],
      "forbidden_source_tokens",
    ).map((rule, index) => ({
      id: requireString(rule?.id, `forbidden_source_tokens[${index + 1}].id`),
      tokens: requireStringArray(
        rule?.tokens ?? [],
        `forbidden_source_tokens[${index + 1}].tokens`,
      ),
      scanPaths: requireStringArray(
        rule?.scan_paths ?? [],
        `forbidden_source_tokens[${index + 1}].scan_paths`,
      ).map(normalizePath),
      allowedPaths: requireStringArray(
        rule?.allowed_paths ?? [],
        `forbidden_source_tokens[${index + 1}].allowed_paths`,
      ).map(normalizePath),
      productionOnly: requireBoolean(
        rule?.production_only,
        `forbidden_source_tokens[${index + 1}].production_only`,
      ),
    })),
    forbiddenTestBuildTokens: requireStringArray(
      raw.forbidden_test_build_tokens ?? [],
      "forbidden_test_build_tokens",
    ),
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

function appendUnique(values, additions) {
  const result = [...values];
  const seen = new Set(result);
  for (const addition of additions) {
    if (!seen.has(addition)) {
      seen.add(addition);
      result.push(addition);
    }
  }
  return result;
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
  for (const match of content.matchAll(/import\s+(?:[._A-Za-z][A-Za-z0-9_]*\s+)?"([^"]+)"/g)) {
    imports.push(match[1]);
  }
  return imports;
}

function pathMatchesAny(relative, patterns) {
  return patterns.some((pattern) => matchesPattern(relative, pattern));
}

function isProductionGo(file, productionOnly) {
  return file.relative.endsWith(".go") && (!productionOnly || !file.relative.endsWith("_test.go"));
}

function violation(code, file, symbolOrImport, details = {}) {
  return {
    code,
    path: file.relative,
    symbol_or_import: symbolOrImport,
    ...details,
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

function checkForbiddenGoImports(files, rules) {
  const violations = [];
  for (const file of files) {
    for (const rule of rules) {
      if (!isProductionGo(file, rule.productionOnly)) {
        continue;
      }
      if (rule.scanPaths.length > 0 && !pathMatchesAny(file.relative, rule.scanPaths)) {
        continue;
      }
      if (pathMatchesAny(file.relative, rule.allowedPaths)) {
        continue;
      }
      for (const imported of extractGoImports(file.content)) {
        for (const forbiddenImport of rule.imports) {
          const matches = rule.matchKind === "exact"
            ? imported === forbiddenImport
            : imported === forbiddenImport || imported.startsWith(`${forbiddenImport}/`);
          if (matches) {
            violations.push(violation("forbidden_go_import", file, imported, {
              rule_id: rule.id,
              match_kind: rule.matchKind,
            }));
          }
        }
      }
    }
  }
  return violations;
}

function checkGoImportAllowlists(files, rules) {
  const violations = [];
  for (const file of files) {
    for (const rule of rules) {
      if (!isProductionGo(file, rule.productionOnly)) {
        continue;
      }
      if (rule.scanPaths.length > 0 && !pathMatchesAny(file.relative, rule.scanPaths)) {
        continue;
      }
      for (const imported of extractGoImports(file.content)) {
        if (!imported.startsWith(rule.importPrefix)) {
          continue;
        }
        if (rule.allowedImports.has(imported)) {
          continue;
        }
        const allowedByPrefix = rule.allowedPrefixes.some(
          (prefix) => imported === prefix || imported.startsWith(`${prefix}/`),
        );
        if (!allowedByPrefix) {
          violations.push(violation("go_import_allowlist", file, imported));
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

function checkSourceTableAccess(files, rules) {
  const violations = [];
  const sourceFiles = files.filter((entry) => {
    if (entry.relative.endsWith("_test.go")) {
      return false;
    }
    return entry.relative.endsWith(".go") || entry.relative.endsWith(".sql");
  });
  for (const file of sourceFiles) {
    for (const rule of rules) {
      if (
        rule.scanPaths.length > 0 &&
        !rule.scanPaths.some((pattern) => matchesPattern(file.relative, pattern))
      ) {
        continue;
      }
      if (rule.allowedPaths.some((pattern) => matchesPattern(file.relative, pattern))) {
        continue;
      }
      for (const table of rule.tables) {
        if (mentionsSourceTableAccess(file.content, table)) {
          violations.push(violation("source_table_access", file, table));
        }
      }
    }
  }
  return violations;
}

function checkForbiddenSourceMappings(files, rules) {
  const violations = [];
  const sourceFiles = files.filter(
    (entry) => entry.relative.endsWith(".go") && !entry.relative.endsWith("_test.go"),
  );
  for (const file of sourceFiles) {
    for (const rule of rules) {
      if (!rule.scanPaths.some((pattern) => matchesPattern(file.relative, pattern))) {
        continue;
      }
      for (const prefix of rule.prefixes) {
        const escaped = prefix.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        const mappingPattern = new RegExp(`[\"'\u0060]${escaped}[A-Za-z0-9_]+[\"'\u0060]\\s*:`, "m");
        if (mappingPattern.test(file.content)) {
          violations.push(violation("source_mapping", file, prefix));
        }
      }
    }
  }
  return violations;
}

function checkSQLTableAllowlists(files, rules) {
  const violations = [];
  const sourceFiles = files.filter((entry) => {
    if (entry.relative.endsWith("_test.go")) {
      return false;
    }
    return entry.relative.endsWith(".go") || entry.relative.endsWith(".sql");
  });
  for (const file of sourceFiles) {
    for (const rule of rules) {
      if (!rule.scanPaths.some((pattern) => matchesPattern(file.relative, pattern))) {
        continue;
      }
      for (const table of sqlTableReferences(file.content)) {
        if (!rule.allowedTables.has(table)) {
          violations.push(violation("sql_table_allowlist", file, table));
        }
      }
    }
  }
  return violations;
}

function checkSQLTableAccess(files, rules, accessScope = "production") {
  const violations = [];
  const sourceFiles = files.filter((entry) => {
    if (entry.relative.endsWith("_test.go")) {
      return false;
    }
    return entry.relative.endsWith(".go") || entry.relative.endsWith(".sql");
  });
  for (const file of sourceFiles) {
    for (const rule of rules) {
      if (!pathMatchesAny(file.relative, rule.scanPaths)) {
        continue;
      }
      for (const access of sqlTableAccesses(file.content)) {
        if (access.table !== rule.table) {
          continue;
        }
        const allowedPaths = access.operation === "read"
          ? accessScope === "test_fixture"
            ? rule.testReadAllowedPaths
            : rule.readAllowedPaths
          : accessScope === "test_fixture"
            ? rule.testWriteAllowedPaths
            : rule.writeAllowedPaths;
        if (!pathMatchesAny(file.relative, allowedPaths)) {
          violations.push(
            violation(
              access.operation === "read" ? "sql_table_read_access" : "sql_table_write_access",
              file,
              rule.table,
              { rule_id: rule.id, access_scope: accessScope },
            ),
          );
        }
      }
      const lockAllowedPaths = accessScope === "test_fixture" ? [] : rule.lockAllowedPaths;
      if (
        lockAllowedPaths !== null &&
        mentionsSQLTableLock(file.content, rule.table) &&
        !pathMatchesAny(file.relative, lockAllowedPaths)
      ) {
        violations.push(
          violation("sql_table_lock_access", file, rule.table, {
            rule_id: rule.id,
            access_scope: accessScope,
          }),
        );
      }
    }
  }
  return violations;
}

function mentionsSQLTableLock(content, table) {
  const escaped = table.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const tableReference = `(?:FROM|JOIN)\\s+(?:public\\.)?${escaped}\\b`;
  const lockInStatement = new RegExp(
    `${tableReference}[^;\u0060]{0,4000}\\bFOR\\s+(?:NO\\s+KEY\\s+)?UPDATE\\b`,
    "gis",
  );
  return lockInStatement.test(content);
}

function sqlTableAccesses(content) {
  const accesses = [];
  const pattern =
    /\b(FROM|JOIN|UPDATE|INSERT\s+INTO|DELETE\s+FROM)\s+(?:public\.)?([a-z_][a-z0-9_]*)\b/gi;
  for (const match of content.matchAll(pattern)) {
    const keyword = match[1].toUpperCase().replace(/\s+/g, " ");
    accesses.push({
      operation: keyword === "FROM" || keyword === "JOIN" ? "read" : "write",
      table: match[2].toLowerCase(),
    });
  }
  return accesses;
}

function sqlTableReferences(content) {
  const tables = new Set();
  const pattern = /\b(?:FROM|JOIN|UPDATE|INSERT\s+INTO|DELETE\s+FROM)\s+(?:public\.)?([a-z_][a-z0-9_]*)\b/gi;
  for (const match of content.matchAll(pattern)) {
    const table = match[1].toLowerCase();
    if (table !== "set") {
      tables.add(table);
    }
  }
  return Array.from(tables).sort();
}

function assertBoundaryFixtures(manifest) {
  assertArtifactBoundaryFixtures(manifest);

  const recordsRule = manifest.sqlTableAccess.find(
    (rule) => rule.id === "records-current-envelope-access",
  );
  if (!recordsRule) {
    throw new Error("records-current-envelope-access boundary rule is required");
  }
  const fixtureFiles = [
    {
      label: "records owner write",
      file: {
        relative: "internal/modules/records/fixture_store.go",
        content: "package records\nconst sql = `UPDATE records SET row_version = row_version + 1`",
      },
      wantCode: null,
    },
    {
      label: "revisions write",
      file: {
        relative: "internal/modules/revisions/fixture_store.go",
        content: "package revisions\nconst sql = `UPDATE records SET row_version = 2`",
      },
      wantCode: "sql_table_write_access",
    },
    {
      label: "unapproved source-owner read join",
      file: {
        relative: "internal/modules/entities/fixture_query.go",
        content: "package entities\nconst sql = `SELECT r.record_id FROM records r`",
      },
      wantCode: "sql_table_read_access",
    },
  ];
  for (const fixture of fixtureFiles) {
    const violations = checkSQLTableAccess([fixture.file], [recordsRule]);
    if (fixture.wantCode === null && violations.length !== 0) {
      throw new Error(`${fixture.label} boundary fixture must pass`);
    }
    if (
      fixture.wantCode !== null &&
      (violations.length !== 1 || violations[0].code !== fixture.wantCode)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${fixture.wantCode}`);
    }
  }

  const projectionRules = manifest.sqlTableAccess.filter((rule) =>
    rule.id.endsWith("-projection-storage-access"),
  );
  if (projectionRules.length !== 10 || new Set(projectionRules.map((rule) => rule.table)).size !== 10) {
    throw new Error("exactly ten distinct projection storage access rules are required");
  }
  const timelineProjectionRule = projectionRules.find(
    (rule) => rule.table === "timeline_grid_projection",
  );
  const assessmentProjectionRule = projectionRules.find(
    (rule) => rule.table === "assessment_grid_projection",
  );
  const projectionFixtures = [
    {
      label: "private projection query read",
      scope: "production",
      rule: timelineProjectionRule,
      file: {
        relative: "internal/modules/projections/internal/queryengine/timeline.go",
        content: "package queryengine\nconst sql = `SELECT record_id FROM timeline_grid_projection`",
      },
      wantCode: null,
    },
    {
      label: "source-owner projection read",
      scope: "production",
      rule: timelineProjectionRule,
      file: {
        relative: "internal/modules/timeline/store.go",
        content: "package timeline\nconst sql = `SELECT record_id FROM timeline_grid_projection`",
      },
      wantCode: "sql_table_read_access",
    },
    {
      label: "Timeline test-support projection read",
      scope: "test_fixture",
      rule: timelineProjectionRule,
      file: {
        relative: "internal/modules/timeline/testsupport/asserttest/assertions.go",
        content: "package asserttest\nconst sql = `SELECT COUNT(*) FROM timeline_grid_projection`",
      },
      wantCode: null,
    },
    {
      label: "Projections test-support assessment projection write",
      scope: "test_fixture",
      rule: assessmentProjectionRule,
      file: {
        relative: "internal/modules/projections/testsupport/capability.go",
        content: "package testsupport\nconst sql = `INSERT INTO assessment_grid_projection (record_id) VALUES ($1)`",
      },
      wantCode: null,
    },
    {
      label: "Projections test-support path is not a production permission",
      scope: "production",
      rule: assessmentProjectionRule,
      file: {
        relative: "internal/modules/projections/testsupport/capability.go",
        content: "package testsupport\nconst sql = `INSERT INTO assessment_grid_projection (record_id) VALUES ($1)`",
      },
      wantCode: "sql_table_write_access",
    },
    {
      label: "unapproved projection fixture write",
      scope: "test_fixture",
      rule: assessmentProjectionRule,
      file: {
        relative: "internal/modules/entities/testsupport/fixture.go",
        content: "package testsupport\nconst sql = `DELETE FROM assessment_grid_projection`",
      },
      wantCode: "sql_table_write_access",
    },
  ];
  for (const fixture of projectionFixtures) {
    if (!fixture.rule) {
      throw new Error(`${fixture.label} boundary rule is required`);
    }
    const violations = checkSQLTableAccess([fixture.file], [fixture.rule], fixture.scope);
    if (fixture.wantCode === null && violations.length !== 0) {
      throw new Error(`${fixture.label} boundary fixture must pass`);
    }
    if (
      fixture.wantCode !== null &&
      (violations.length !== 1 || violations[0].code !== fixture.wantCode)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${fixture.wantCode}`);
    }
  }

  const jobsStorageRule = manifest.sqlTableAccess.find(
    (rule) => rule.id === "jobs-storage-owner-write-and-lock-access",
  );
  const jobsForeignStorageRuleIDs = [
    "jobs-no-auth-route-idempotency-storage-access",
    "jobs-no-extension-cancellation-storage-access",
    "jobs-no-collaboration-intent-storage-access",
  ];
  const jobsForeignStorageRules = jobsForeignStorageRuleIDs.map((id) => {
    const rule = manifest.sqlTableAccess.find((candidate) => candidate.id === id);
    if (!rule) {
      throw new Error(`${id} boundary rule is required`);
    }
    return rule;
  });
  if (!jobsStorageRule) {
    throw new Error("jobs-storage-owner-write-and-lock-access boundary rule is required");
  }
  const jobsStorageFixtures = [
    {
      label: "Jobs owner write",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content: "package jobs\nconst sql = `UPDATE jobs SET status = 'running'`",
      },
      rule: jobsStorageRule,
      wantCode: null,
    },
    {
      label: "consumer Jobs read",
      file: {
        relative: "internal/modules/jobapi/fixture.go",
        content: "package jobapi\nconst sql = `SELECT status FROM jobs`",
      },
      rule: jobsStorageRule,
      wantCode: null,
    },
    {
      label: "non-owner Jobs write",
      file: {
        relative: "internal/modules/imports/fixture.go",
        content: "package imports\nconst sql = `UPDATE jobs SET status = 'failed'`",
      },
      rule: jobsStorageRule,
      wantCode: "sql_table_write_access",
    },
    {
      label: "non-owner Jobs transition lock",
      file: {
        relative: "internal/platform/extensionstore/fixture.go",
        content: "package extensionstore\nconst sql = `SELECT status FROM jobs FOR UPDATE`",
      },
      rule: jobsStorageRule,
      wantCode: "sql_table_lock_access",
    },
    {
      label: "auth owner adapter storage access",
      file: {
        relative: "internal/app/server/job_owner_ports.go",
        content: "package server\nconst sql = `SELECT response_json FROM route_idempotency`",
      },
      rule: jobsForeignStorageRules[0],
      wantCode: null,
    },
    {
      label: "Jobs auth storage access",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content: "package jobs\nconst sql = `SELECT response_json FROM route_idempotency`",
      },
      rule: jobsForeignStorageRules[0],
      wantCode: "sql_table_read_access",
    },
    {
      label: "Jobs extension cancellation storage access",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content:
          "package jobs\nconst sql = `INSERT INTO extension_job_cancellation_observations (job_id) VALUES ($1)`",
      },
      rule: jobsForeignStorageRules[1],
      wantCode: "sql_table_write_access",
    },
    {
      label: "Jobs Collaboration storage access",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content: "package jobs\nconst sql = `INSERT INTO collaboration_event_intents (intent_key) VALUES ($1)`",
      },
      rule: jobsForeignStorageRules[2],
      wantCode: "sql_table_write_access",
    },
    {
      label: "Jobs test storage access",
      file: {
        relative: "internal/platform/jobs/fixture_test.go",
        content: "package jobs\nconst sql = `UPDATE route_idempotency SET status_code = 200`",
      },
      rule: jobsForeignStorageRules[0],
      wantCode: null,
    },
    {
      label: "Jobs migration write",
      file: {
        relative: "db/migrations/fixture.sql",
        content: "UPDATE jobs SET job_kind = extension_job_kind;",
      },
      rule: jobsStorageRule,
      wantCode: null,
    },
  ];
  for (const fixture of jobsStorageFixtures) {
    const violations = checkSQLTableAccess([fixture.file], [fixture.rule]);
    if (fixture.wantCode === null && violations.length !== 0) {
      throw new Error(`${fixture.label} boundary fixture must pass`);
    }
    if (
      fixture.wantCode !== null &&
      (violations.length !== 1 ||
        violations[0].code !== fixture.wantCode ||
        violations[0].rule_id !== fixture.rule.id)
    ) {
      throw new Error(
        `${fixture.label} boundary fixture must fail with ${fixture.rule.id}/${fixture.wantCode}`,
      );
    }
  }

  const jobsCapabilityRule = manifest.forbiddenSourceTokens.find(
    (rule) => rule.id === "jobs-consumers-use-narrow-capabilities",
  );
  if (!jobsCapabilityRule) {
    throw new Error("jobs-consumers-use-narrow-capabilities boundary rule is required");
  }
  const jobsCapabilityFixtures = [
    {
      label: "application Jobs composition",
      file: {
        relative: "internal/app/server/fixture.go",
        content: "package server\ntype dependencies struct { Jobs *jobs.Manager }",
      },
      wantViolation: false,
    },
    {
      label: "consumer broad Jobs manager",
      file: {
        relative: "internal/modules/imports/fixture.go",
        content: "package imports\ntype store struct { jobs *jobs.Manager }",
      },
      wantViolation: true,
    },
    {
      label: "consumer raw Jobs transition lock",
      file: {
        relative: "internal/modules/imports/fixture.go",
        content: "package imports\nfunc lock() { jobs.LockTransitionTx(nil, nil, id) }",
      },
      wantViolation: true,
    },
    {
      label: "consumer test broad Jobs manager",
      file: {
        relative: "internal/modules/imports/fixture_test.go",
        content: "package imports\ntype fixture struct { jobs *jobs.Manager }",
      },
      wantViolation: false,
    },
  ];
  for (const fixture of jobsCapabilityFixtures) {
    const violations = checkForbiddenSourceTokens([fixture.file], [jobsCapabilityRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 || violations[0].code !== "forbidden_source_token")
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with the capability rule`);
    }
  }

  const jobsPeerImportRule = manifest.forbiddenGoImports.find(
    (rule) => rule.id === "jobs-use-owner-transaction-ports-not-peer-persistence",
  );
  if (!jobsPeerImportRule) {
    throw new Error(
      "jobs-use-owner-transaction-ports-not-peer-persistence boundary rule is required",
    );
  }
  const jobsPeerImportFixtures = [
    {
      label: "application Auth owner adapter",
      file: {
        relative: "internal/app/server/job_owner_ports.go",
        content:
          'package server\nimport "github.com/JochiRaider/cartulary/internal/platform/authn"',
      },
      wantViolation: false,
    },
    {
      label: "Jobs peer persistence import",
      file: {
        relative: "internal/platform/jobs/fixture.go",
        content:
          'package jobs\nimport "github.com/JochiRaider/cartulary/internal/platform/authn"',
      },
      wantViolation: true,
    },
  ];
  for (const fixture of jobsPeerImportFixtures) {
    const violations = checkForbiddenGoImports([fixture.file], [jobsPeerImportRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 ||
        violations[0].code !== "forbidden_go_import" ||
        violations[0].rule_id !== jobsPeerImportRule.id)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with the peer import rule`);
    }
  }

  const genericDeleteRestoreRule = manifest.forbiddenSourceTokens.find(
    (rule) => rule.id === "delete-restore-sources-no-generic-sql-metadata",
  );
  if (!genericDeleteRestoreRule) {
    throw new Error("delete-restore-sources-no-generic-sql-metadata boundary rule is required");
  }
  const genericProvider = {
    relative: "internal/modules/artifacts/internal/providers/deleterestore/fixture.go",
    content: "package deleterestore\ntype TableProvider struct { SourceTable string }",
  };
  const genericViolations = checkForbiddenSourceTokens(
    [genericProvider],
    [genericDeleteRestoreRule],
  );
  if (
    genericViolations.length !== 2 ||
    genericViolations.some((entry) => entry.code !== "forbidden_source_token")
  ) {
    throw new Error("generic delete/restore provider boundary fixture must fail closed");
  }

  const deleteRestoreContractRule = manifest.ownerPortOnlyImports.find(
    (rule) => rule.id === "revisions-delete-restore-source-contract-import",
  );
  if (!deleteRestoreContractRule) {
    throw new Error("revisions-delete-restore-source-contract-import boundary rule is required");
  }
  const contractImport =
    'github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract';
  const contractFixtures = [
    {
      label: "source owner adapter contract import",
      file: {
        relative: "internal/modules/artifacts/internal/providers/deleterestore/provider.go",
        content: `package deleterestore\nimport "${contractImport}"`,
      },
      wantViolation: false,
    },
    {
      label: "unapproved contract import",
      file: {
        relative: "internal/modules/workbook/fixture.go",
        content: `package workbook\nimport "${contractImport}"`,
      },
      wantViolation: true,
    },
  ];
  for (const fixture of contractFixtures) {
    const violations = checkOwnerPortOnlyImports([fixture.file], [deleteRestoreContractRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
  }

  const configImportRule = manifest.goImportAllowlists.find(
    (rule) => rule.id === "platform-config-kernel-internal-imports",
  );
  if (!configImportRule) {
    throw new Error("platform-config-kernel-internal-imports boundary rule is required");
  }
  const configImportFixtures = [
    {
      label: "platform-neutral external import",
      imported: "github.com/BurntSushi/toml",
      wantViolation: false,
    },
    {
      label: "application import",
      imported: "github.com/JochiRaider/cartulary/internal/app/configassembly",
      wantViolation: true,
    },
    {
      label: "module owner import",
      imported: "github.com/JochiRaider/cartulary/internal/modules/extensions",
      wantViolation: true,
    },
    {
      label: "platform owner import",
      imported: "github.com/JochiRaider/cartulary/internal/platform/telemetry",
      wantViolation: true,
    },
  ];
  for (const fixture of configImportFixtures) {
    const file = {
      relative: "internal/platform/config/fixture.go",
      content: `package config\nimport "${fixture.imported}"`,
    };
    const violations = checkGoImportAllowlists([file], [configImportRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} config-kernel boundary fixture produced an unexpected result`);
    }
  }

  const configRegistryRule = manifest.forbiddenSourceTokens.find(
    (rule) => rule.id === "platform-config-no-owner-profile-registry",
  );
  if (!configRegistryRule) {
    throw new Error("platform-config-no-owner-profile-registry boundary rule is required");
  }
  const profileRegistry = {
    relative: "internal/platform/config/fixture.go",
    content: 'package config\nvar claimPaths = []string{"import.claimed"}',
  };
  const profileRegistryViolations = checkForbiddenSourceTokens(
    [profileRegistry],
    [configRegistryRule],
  );
  if (
    profileRegistryViolations.length !== 1 ||
    profileRegistryViolations[0].code !== "forbidden_source_token"
  ) {
    throw new Error("platform config profile-registry boundary fixture must fail closed");
  }

  const contractedConfigRuleIDs = [
    "platform-config-no-retired-document-snapshot-access",
    "platform-config-no-in-memory-toml-snapshot-admission",
    "configassembly-no-in-memory-admission-definition",
    "application-no-configassembly-in-memory-admission-callers",
    "configassembly-deployment-no-wire-tags",
  ];
  const contractedConfigRules = contractedConfigRuleIDs.map((id) => {
    const rule = manifest.forbiddenSourceTokens.find((candidate) => candidate.id === id);
    if (!rule) {
      throw new Error(`${id} boundary rule is required`);
    }
    return rule;
  });
  const retiredConfigFixtures = [
    {
      label: "retired full-document source",
      relative: "internal/platform/config/source.go",
      content: "package config\ntype Source interface { Decode(path []string, target any) error }",
      ruleID: "platform-config-no-retired-document-snapshot-access",
    },
    {
      label: "retired reflective owner clone",
      relative: "internal/platform/config/snapshot.go",
      content: "package config\nfunc cloneConfig(value any) any { return value }",
      ruleID: "platform-config-no-retired-document-snapshot-access",
    },
    {
      label: "retired in-memory TOML snapshot loading",
      relative: "internal/platform/config/snapshot_test.go",
      content: "package config\nfunc fixture() { LoadSnapshotFromTOML(nil, Catalog{}) }",
      ruleID: "platform-config-no-in-memory-toml-snapshot-admission",
    },
    {
      label: "retired configassembly admission definition",
      relative: "internal/app/configassembly/configuration.go",
      content: "package configassembly\nfunc Admit(value Deployment) (Loaded, error) { return Loaded{}, nil }",
      ruleID: "configassembly-no-in-memory-admission-definition",
    },
    {
      label: "retired configassembly admission caller",
      relative: "internal/app/server/runtime_test.go",
      content: "package server\nfunc fixture() { _, _ = configassembly.Admit(configassembly.Deployment{}) }",
      ruleID: "application-no-configassembly-in-memory-admission-callers",
    },
    {
      label: "wire-tagged application deployment projection",
      relative: "internal/app/configassembly/deployment.go",
      content: 'package configassembly\ntype Deployment struct { Profile string `toml:"deployment_profile"` }',
      ruleID: "configassembly-deployment-no-wire-tags",
    },
  ];
  for (const fixture of retiredConfigFixtures) {
    const rule = contractedConfigRules.find((candidate) => candidate.id === fixture.ruleID);
    const violations = checkForbiddenSourceTokens(
      [{ relative: fixture.relative, content: fixture.content }],
      [rule],
    );
    if (violations.length !== 1 || violations[0].code !== "forbidden_source_token") {
      throw new Error(`${fixture.label} boundary fixture must fail closed`);
    }
  }

  const supportedConfigFixtures = [
    {
      relative: "internal/platform/config/catalog.go",
      content:
        "package config\ntype NamespaceDecoder interface { Decode(target any) error }\ntype NamespacePresence interface { Defined(path ...string) bool }",
    },
    {
      relative: "internal/platform/config/config.go",
      content:
        'package config\nimport "reflect"\nfunc applyOverlay(target reflect.Value) reflect.Value { return target }',
    },
    {
      relative: "internal/app/configassembly/configuration.go",
      content:
        "package configassembly\nvar owner = Definition[Configuration]{Clone: func(value Configuration) Configuration { return value }}",
    },
  ];
  const supportedConfigViolations = checkForbiddenSourceTokens(
    supportedConfigFixtures,
    contractedConfigRules,
  );
  if (supportedConfigViolations.length !== 0) {
    throw new Error("supported config seams must remain accepted by contracted boundary rules");
  }
}

function assertArtifactBoundaryFixtures(manifest) {
  const httpRule = manifest.forbiddenGoImports.find(
    (rule) => rule.id === "artifact-subtree-no-http-or-auth-transport",
  );
  const consumerRule = manifest.forbiddenGoImports.find(
    (rule) => rule.id === "artifact-subtree-no-consumer-implementations",
  );
  const idempotencyRule = manifest.forbiddenSourceTokens.find(
    (rule) => rule.id === "artifact-mutation-idempotency-lookup-is-centralized",
  );
  if (!httpRule || !consumerRule || !idempotencyRule) {
    throw new Error("Artifact subtree boundary rules are required");
  }

  const httpFixtures = [
    {
      label: "Artifact transport-neutral source",
      file: { relative: "internal/modules/artifacts/future_surface.go", content: "package artifacts" },
      wantViolation: false,
    },
    {
      label: "Artifact future HTTP import",
      file: {
        relative: "internal/modules/artifacts/future_surface.go",
        content: 'package artifacts\nimport "net/http"',
      },
      wantViolation: true,
    },
  ];
  for (const fixture of httpFixtures) {
    const violations = checkForbiddenGoImports([fixture.file], [httpRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 ||
        violations[0].code !== "forbidden_go_import" ||
        violations[0].rule_id !== httpRule.id)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${httpRule.id}`);
    }
  }

  const consumerFixtures = [
    {
      label: "Artifact published projection contribution contract",
      file: {
        relative: "internal/modules/artifacts/workbookprojection/contribution.go",
        content:
          'package workbookprojection\nimport "github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"',
      },
      wantViolation: false,
    },
    {
      label: "Artifact future Workbook implementation import",
      file: {
        relative: "internal/modules/artifacts/future_surface.go",
        content:
          'package artifacts\nimport "github.com/JochiRaider/cartulary/internal/modules/workbook"',
      },
      wantViolation: true,
    },
  ];
  for (const fixture of consumerFixtures) {
    const violations = checkForbiddenGoImports([fixture.file], [consumerRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 ||
        violations[0].code !== "forbidden_go_import" ||
        violations[0].rule_id !== consumerRule.id)
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${consumerRule.id}`);
    }
  }

  const idempotencyFixtures = [
    {
      label: "Artifact centralized idempotency lookup",
      file: {
        relative: "internal/modules/artifacts/mutation_idempotency.go",
        content: "package artifacts\nfunc lookup() { f.idempotency.Get(ctx, key, hash) }",
      },
      wantViolation: false,
    },
    {
      label: "Artifact future direct idempotency lookup",
      file: {
        relative: "internal/modules/artifacts/future_surface.go",
        content: "package artifacts\nfunc lookup() { f.idempotency.Get(ctx, key, hash) }",
      },
      wantViolation: true,
    },
  ];
  for (const fixture of idempotencyFixtures) {
    const violations = checkForbiddenSourceTokens([fixture.file], [idempotencyRule]);
    if ((violations.length > 0) !== fixture.wantViolation) {
      throw new Error(`${fixture.label} boundary fixture produced an unexpected result`);
    }
    if (
      fixture.wantViolation &&
      (violations.length !== 1 || violations[0].code !== "forbidden_source_token")
    ) {
      throw new Error(`${fixture.label} boundary fixture must fail with ${idempotencyRule.id}`);
    }
  }
}

function checkForbiddenGoCalls(files, rules) {
  const violations = [];
  const sourceFiles = files.filter(
    (entry) => entry.relative.endsWith(".go") && !entry.relative.endsWith("_test.go"),
  );
  for (const file of sourceFiles) {
    for (const rule of rules) {
      if (rule.allowedPaths.some((pattern) => matchesPattern(file.relative, pattern))) {
        continue;
      }
      for (const symbol of rule.symbols) {
        if (file.content.includes(symbol)) {
          violations.push(violation("forbidden_go_call", file, symbol));
        }
      }
    }
  }
  return violations;
}

function sanitizeGoStructure(content) {
  let result = "";
  let state = "code";
  for (let index = 0; index < content.length; index += 1) {
    const current = content[index];
    const next = content[index + 1] ?? "";
    if (state === "code") {
      if (current === "/" && next === "/") {
        result += "  ";
        index += 1;
        state = "line-comment";
      } else if (current === "/" && next === "*") {
        result += "  ";
        index += 1;
        state = "block-comment";
      } else if (current === '"') {
        result += " ";
        state = "string";
      } else if (current === "'") {
        result += " ";
        state = "rune";
      } else if (current === "`") {
        result += " ";
        state = "raw-string";
      } else {
        result += current;
      }
      continue;
    }
    if (state === "line-comment") {
      result += current === "\n" ? "\n" : " ";
      if (current === "\n") state = "code";
      continue;
    }
    if (state === "block-comment") {
      if (current === "*" && next === "/") {
        result += "  ";
        index += 1;
        state = "code";
      } else {
        result += current === "\n" ? "\n" : " ";
      }
      continue;
    }
    if (state === "raw-string") {
      result += current === "\n" ? "\n" : " ";
      if (current === "`") state = "code";
      continue;
    }
    result += current === "\n" ? "\n" : " ";
    if (current === "\\") {
      if (index + 1 < content.length) {
        result += content[index + 1] === "\n" ? "\n" : " ";
        index += 1;
      }
    } else if ((state === "string" && current === '"') || (state === "rune" && current === "'")) {
      state = "code";
    }
  }
  return result;
}

function goDeclarationNames(content) {
  const source = sanitizeGoStructure(content);
  const names = new Set();
  for (const match of source.matchAll(/\bfunc\s*(?:\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)\s*\(/g)) {
    names.add(match[1]);
  }
  for (const match of source.matchAll(/\b(?:var|const|type)\s+([A-Za-z_][A-Za-z0-9_]*)\b/g)) {
    names.add(match[1]);
  }
  for (const match of source.matchAll(/\b(?:var|const|type)\s*\(([\s\S]*?)\)/g)) {
    for (const line of match[1].split("\n")) {
      const name = line.trim().match(/^([A-Za-z_][A-Za-z0-9_]*)\b/)?.[1];
      if (name) names.add(name);
    }
  }
  return names;
}

function goCallCount(content, symbol) {
  const source = sanitizeGoStructure(content).replace(
    /(\bfunc\s*(?:\([^)]*\)\s*)?)([A-Za-z_][A-Za-z0-9_]*)(\s*\()/g,
    (_match, prefix, name, suffix) => `${prefix}${" ".repeat(name.length)}${suffix}`,
  );
  const escaped = symbol.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  return [...source.matchAll(new RegExp(`\\b${escaped}\\s*\\(`, "g"))].length;
}

function checkForbiddenGoDeclarations(files, rules) {
  const violations = [];
  for (const file of files) {
    for (const rule of rules) {
      if (!isProductionGo(file, rule.productionOnly)) continue;
      if (!pathMatchesAny(file.relative, rule.scanPaths)) continue;
      if (pathMatchesAny(file.relative, rule.allowedPaths)) continue;
      const declarations = goDeclarationNames(file.content);
      for (const symbol of rule.symbols) {
        if (declarations.has(symbol)) {
          violations.push(violation("forbidden_go_declaration", file, symbol, { rule_id: rule.id }));
        }
      }
    }
  }
  return violations;
}

function checkExactGoCalls(files, rules) {
  const violations = [];
  for (const rule of rules) {
    let requiredCount = 0;
    for (const file of files) {
      if (!isProductionGo(file, rule.productionOnly)) continue;
      if (!pathMatchesAny(file.relative, rule.scanPaths)) continue;
      const count = goCallCount(file.content, rule.symbol);
      if (file.relative === rule.requiredPath) requiredCount += count;
      if (count > 0 && !pathMatchesAny(file.relative, rule.allowedPaths)) {
        violations.push(violation("forbidden_go_call", file, rule.symbol, {
          rule_id: rule.id,
          occurrences: count,
        }));
      }
    }
    if (requiredCount !== rule.requiredCount) {
      violations.push(violation(
        "required_go_call_count",
        { relative: rule.requiredPath },
        rule.symbol,
        { rule_id: rule.id, expected: rule.requiredCount, observed: requiredCount },
      ));
    }
  }
  return violations;
}

function assertMigrationSourceBoundaryFixtures(boundary) {
  const approved = [
    {
      relative: "db/migrations/source.go",
      content: "package migrations\nfunc Source() { database_migrations.BuildCanonicalEmbedded(files, embeddedPath) }",
    },
    {
      relative: "internal/modules/database_migrations/migrations.go",
      content: 'package database_migrations\nimport "github.com/JochiRaider/cartulary/internal/modules/database_migrations/sourcecatalog"\nfunc BuildCanonicalEmbedded() {}',
    },
    {
      relative: "internal/testutil/pgtest/pgtest.go",
      content: 'package pgtest\nimport "github.com/JochiRaider/cartulary/internal/modules/database_migrations/sourcecatalog"\nfunc (m *MigrationDatabase) ApplyThrough() {}\nfunc (m *MigrationDatabase) RollbackThrough() {}',
    },
    {
      relative: "internal/app/server/runtime_test.go",
      content: 'package server\nimport "github.com/JochiRaider/cartulary/internal/testutil/pgtest"',
    },
  ];
  if (
    checkForbiddenGoImports(approved, boundary.restrictedImports).length !== 0 ||
    checkForbiddenGoDeclarations(approved, boundary.forbiddenDeclarations).length !== 0 ||
    checkExactGoCalls(approved, boundary.exactCalls).length !== 0
  ) {
    throw new Error("approved database migration source boundary fixtures must pass");
  }

  for (const rule of boundary.restrictedImports) {
    const imported = rule.imports[0];
    const fixture = [{
      relative: "internal/app/server/forbidden.go",
      content: `package server\nimport "${imported}"`,
    }];
    const violations = checkForbiddenGoImports(fixture, [rule]);
    if (violations.length !== 1 || violations[0].rule_id !== rule.id) {
      throw new Error(`${rule.id} negative import fixture must fail closed`);
    }
  }
  for (const rule of boundary.forbiddenDeclarations) {
    for (const symbol of rule.symbols) {
      const fixture = [{
        relative: rule.scanPaths[0].replace("/**", "/forbidden.go"),
        content: `package forbidden\nfunc ${symbol}() {}`,
      }];
      const violations = checkForbiddenGoDeclarations(fixture, [rule]);
      if (violations.length !== 1 || violations[0].symbol_or_import !== symbol) {
        throw new Error(`${rule.id}/${symbol} negative declaration fixture must fail closed`);
      }
    }
  }
  for (const rule of boundary.exactCalls) {
    const fixtures = [
      {
        relative: rule.requiredPath,
        content: `package approved\nfunc Source() { ${rule.symbol}() }`,
      },
      {
        relative: "internal/app/server/forbidden.go",
        content: `package server\nfunc build() { ${rule.symbol}() }`,
      },
    ];
    const violations = checkExactGoCalls(fixtures, [rule]);
    if (violations.length !== 1 || violations[0].code !== "forbidden_go_call") {
      throw new Error(`${rule.id} negative call fixture must reject the extra caller`);
    }
  }
}

function checkCommandRootShape(files, rule) {
  const prefix = `${rule.root}/`;
  return files
    .filter((file) => file.relative.startsWith(prefix) && file.relative.endsWith(".go"))
    .filter((file) => !rule.allowedGoFiles.has(file.relative))
    .map((file) => violation("command_root_shape", file, "unexpected_go_file"));
}

function checkForbiddenSourceTokens(files, rules) {
  const violations = [];
  for (const file of files) {
    for (const rule of rules) {
      if (!isProductionGo(file, rule.productionOnly)) {
        continue;
      }
      if (rule.scanPaths.length > 0 && !pathMatchesAny(file.relative, rule.scanPaths)) {
        continue;
      }
      if (pathMatchesAny(file.relative, rule.allowedPaths)) {
        continue;
      }
      for (const token of rule.tokens) {
        if (file.content.includes(token)) {
          violations.push(violation("forbidden_source_token", file, `${rule.id}:${token}`));
        }
      }
    }
  }
  return violations;
}

function checkForbiddenTestBuildTokens(files, tokens) {
  const violations = [];
  for (const file of files.filter((entry) => entry.relative.endsWith("_test.go"))) {
    for (const token of tokens) {
      if (file.content.includes(token)) {
        violations.push(violation("forbidden_test_build_token", file, token));
      }
    }
  }
  return violations;
}

function mentionsSourceTableAccess(content, table) {
  const escaped = table.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const qualifiedTable = `(?:public\\.)?${escaped}`;
  const patterns = [
    new RegExp(`\\bFROM\\s+${qualifiedTable}\\b`, "i"),
    new RegExp(`\\bJOIN\\s+${qualifiedTable}\\b`, "i"),
    new RegExp(`\\bUPDATE\\s+${qualifiedTable}\\b`, "i"),
    new RegExp(`\\bINSERT\\s+INTO\\s+${qualifiedTable}\\b`, "i"),
    new RegExp(`\\bDELETE\\s+FROM\\s+${qualifiedTable}\\b`, "i"),
  ];
  return patterns.some((pattern) => pattern.test(content));
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const manifestRawContent = readFileSync(options.manifest, "utf8");
  const manifest = normalizeManifest(JSON.parse(manifestRawContent));
  const supportInventory = readSupportInventory(options.root, options.supportInventory);
  const inventoryScanExcludes = backendRuntimeExcludePatterns(supportInventory);
  const scanExcludes = appendUnique(manifest.scanExcludes, inventoryScanExcludes);
  const files = collectFiles(options.root, manifest.scanRoots, scanExcludes);
  const migrationBoundaryFiles = collectFiles(
    options.root,
    manifest.migrationSourceBoundary.scanRoots,
    scanExcludes,
  );
  const testSupportRoots = supportInventory.roots
    .filter((entry) => entry.runtimeScan === "excluded" && entry.supportScan === "included")
    .map((entry) => entry.path);
  const testSupportFiles = collectFiles(options.root, testSupportRoots, manifest.scanExcludes);
  const projectionStorageRules = manifest.sqlTableAccess.filter((rule) =>
    rule.id.endsWith("-projection-storage-access"),
  );
  assertBoundaryFixtures(manifest);
  assertMigrationSourceBoundaryFixtures(manifest.migrationSourceBoundary);
  const violations = [
    ...checkOwnerPortOnlyImports(files, manifest.ownerPortOnlyImports),
    ...checkRawNDJSONTargets(files, manifest.rawNDJSONTargets),
    ...checkForbiddenRouteDependencies(files, manifest.forbiddenRouteDependencies),
    ...checkForbiddenGoImports(files, manifest.forbiddenGoImports),
    ...checkGoImportAllowlists(files, manifest.goImportAllowlists),
    ...checkSourceTableAccess(files, manifest.sourceTableAccess),
    ...checkForbiddenSourceMappings(files, manifest.forbiddenSourceMappings),
    ...checkSQLTableAllowlists(files, manifest.sqlTableAllowlists),
    ...checkSQLTableAccess(files, manifest.sqlTableAccess),
    ...checkSQLTableAccess(testSupportFiles, projectionStorageRules, "test_fixture"),
    ...checkForbiddenGoCalls(files, manifest.forbiddenGoCalls),
    ...checkCommandRootShape(files, manifest.commandRootShape),
    ...checkForbiddenSourceTokens(files, manifest.forbiddenSourceTokens),
    ...checkForbiddenTestBuildTokens(files, manifest.forbiddenTestBuildTokens),
    ...checkGeneratedRootWrites(files, manifest.generatedRootWrites),
    ...checkForbiddenGoImports(
      migrationBoundaryFiles,
      manifest.migrationSourceBoundary.restrictedImports,
    ),
    ...checkForbiddenGoDeclarations(
      migrationBoundaryFiles,
      manifest.migrationSourceBoundary.forbiddenDeclarations,
    ),
    ...checkExactGoCalls(migrationBoundaryFiles, manifest.migrationSourceBoundary.exactCalls),
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
    support_inventory_digest: sha256(supportInventory.raw),
    effective_scan_excludes: scanExcludes,
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
