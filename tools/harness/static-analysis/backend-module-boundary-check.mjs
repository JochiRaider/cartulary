import { readFileSync } from "node:fs";
import { createSourceAdmission } from "./source-admission.mjs";
import path from "node:path";
import { backendRuntimeExcludePatterns, readSupportInventory } from "./support-inventory-profiles.mjs";
import { normalizeExactFileSets, resolveExactFileSets, validateRepositoryRelativePath } from "./exact-file-sets.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import { inspectGoImports, inspectGoSQL } from "./go-import-inspection.mjs";
import { sqlStatements, analyzedSQLStatements } from "./sql-relations.mjs";

const manifestSchemaID = "cartulary.backend_module_boundaries.v4";
const sourceExtensions = new Set([".go", ".mjs", ".js", ".sh", ".sql"]);
function strings(values) {
  return values.map((value) => value.trim());
}

function paths(values) {
  return values.map((value) => {
    validatePathPattern(value);
    return value;
  });
}

function roots(values) {
  return values.map((value) => validateRepositoryRelativePath(value));
}

function prepareImportCapabilities(entries, protectedImports, matchKind) {
  const seen = new Set();
  return (entries ?? []).map((entry) => {
    const imported = validateRepositoryRelativePath(entry.import, "import capability");
    if (!protectedImports.some((prefix) => imported === prefix || (matchKind === "subtree" && imported.startsWith(`${prefix}/`)))) {
      throw new Error("import capability must belong to the protected import scope");
    }
    if (seen.has(imported)) throw new Error("duplicate exact import capability");
    seen.add(imported);
    return { import: imported, productionPackages: roots(entry.production_packages), testPackages: roots(entry.test_packages) };
  });
}

function prepareImportRule(rule) {
  return {
    id: rule.id.trim(), matchKind: rule.match_kind, imports: strings(rule.imports),
    scanPaths: paths(rule.scan_paths), allowedPaths: paths(rule.allowed_paths), productionOnly: rule.production_only,
    importCapabilities: prepareImportCapabilities(rule.import_capabilities, strings(rule.imports), rule.match_kind),
  };
}

function declarationCapability(entry) {
  if (!/^[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)?$/u.test(entry.declaration)) {
    throw new Error("SQL capability requires an exact declaration identity");
  }
  return { package: validateRepositoryRelativePath(entry.package), declaration: entry.declaration };
}

export function prepareBoundaryPolicy(raw) {
  validateSchemaSync(manifestSchemaID, raw);
  const manifest = {
    scanRoots: roots(raw.scan_roots), scanExcludes: paths(raw.scan_excludes),
    sqlControlCapabilities: raw.sql_control_capabilities.map((entry) => {
      const statements = sqlStatements(entry.sql);
      if (statements.length !== 1 || statements[0].dynamic || statements[0].controls.length === 0 || statements[0].accesses.length !== 0) {
        throw new Error("SQL control capability requires a static control-only command");
      }
      return { ...declarationCapability(entry), key: statements[0].key };
    }),
    sqlForwardingCapabilities: raw.sql_forwarding_capabilities.map((entry) => ({ ...declarationCapability(entry), testSupport: entry.test_support })),
    exactFileSets: normalizeExactFileSets(raw.exact_file_sets),
    migrationSourceBoundary: {
      scanRoots: roots(raw.migration_source_boundary.scan_roots),
      restrictedImports: raw.migration_source_boundary.restricted_imports.map(prepareImportRule),
      forbiddenDeclarations: raw.migration_source_boundary.forbidden_declarations.map((rule) => ({
        id: rule.id.trim(), symbols: strings(rule.symbols), scanPaths: paths(rule.scan_paths),
        allowedPaths: paths(rule.allowed_paths), productionOnly: rule.production_only,
      })),
      exactCalls: raw.migration_source_boundary.exact_calls.map((rule) => ({
        id: rule.id.trim(), symbol: rule.symbol.trim(), scanPaths: paths(rule.scan_paths),
        allowedPaths: paths(rule.allowed_paths), requiredPath: validateRepositoryRelativePath(rule.required_path),
        requiredCount: rule.required_count, productionOnly: rule.production_only,
      })),
    },
    ownerPortOnlyImports: raw.owner_port_only_imports.map((rule) => ({
      id: rule.id.trim(), import: rule.import.trim(), allowedImporters: paths(rule.allowed_importers),
      importCapabilities: prepareImportCapabilities(rule.import_capabilities, [rule.import.trim()], "subtree"),
    })),
    rawNDJSONTargets: raw.raw_ndjson_targets.map((rule) => ({
      id: rule.id.trim(), functionName: rule.function.trim(), allowedPaths: paths(rule.allowed_paths),
    })),
    forbiddenRouteDependencies: raw.forbidden_route_dependencies.map((rule) => ({
      id: rule.id.trim(), importSuffix: rule.import_suffix.trim(), allowedImporters: paths(rule.allowed_importers),
    })),
    forbiddenGoImports: raw.forbidden_go_imports.map(prepareImportRule),
    goImportAllowlists: raw.go_import_allowlists.map((rule) => ({
      id: rule.id.trim(), importPrefix: rule.import_prefix.trim(), allowedImports: new Set(strings(rule.allowed_imports)),
      allowedPrefixes: strings(rule.allowed_prefixes), scanPaths: paths(rule.scan_paths), productionOnly: rule.production_only,
    })),
    sourceTableAccess: raw.source_table_access.map((rule) => ({
      id: rule.id.trim(), tables: strings(rule.tables), scanPaths: paths(rule.scan_paths ?? []), allowedPaths: paths(rule.allowed_paths),
    })),
    forbiddenSourceMappings: raw.forbidden_source_mappings.map((rule) => ({
      id: rule.id.trim(), prefixes: strings(rule.prefixes), scanPaths: paths(rule.scan_paths),
    })),
    sqlTableAllowlists: raw.sql_table_allowlists.map((rule) => ({
      id: rule.id.trim(), allowedTables: new Set(strings(rule.allowed_tables)), scanPaths: paths(rule.scan_paths),
    })),
    sqlTableAccess: raw.sql_table_access.map((rule) => ({
      id: rule.id.trim(), table: rule.table.trim().toLowerCase(), scanPaths: paths(rule.scan_paths),
      readAllowedPaths: paths(rule.read_allowed_paths), writeAllowedPaths: paths(rule.write_allowed_paths),
      testReadAllowedPaths: paths(rule.test_read_allowed_paths ?? []), testWriteAllowedPaths: paths(rule.test_write_allowed_paths ?? []),
      readAllowedStatements: (rule.read_allowed_statements ?? []).map((entry) => {
        const statement = sqlStatements(entry.sql);
        if (statement.length !== 1 || statement[0].dynamic || !statement[0].accesses.some((access) => access.table === rule.table && access.operation === "read") || statement[0].accesses.some((access) => access.table === rule.table && access.operation !== "read")) {
          throw new Error(`${rule.id}: statement grant must declare a static read of ${rule.table}`);
        }
        return { ...declarationCapability(entry), key: statement[0].key };
      }),
      lockAllowedPaths: rule.lock_allowed_paths === undefined ? null : paths(rule.lock_allowed_paths),
    })),
    forbiddenGoCalls: raw.forbidden_go_calls.map((rule) => ({
      id: rule.id.trim(), symbols: strings(rule.symbols), allowedPaths: paths(rule.allowed_paths),
    })),
    commandRootShape: {
      root: validateRepositoryRelativePath(raw.command_root_shape.root),
      allowedGoFiles: new Set(roots(raw.command_root_shape.allowed_go_files)),
    },
    forbiddenSourceTokens: raw.forbidden_source_tokens.map((rule, index) => ({
      id: rule.id.trim(), matchKind: requireSourceTokenMatchKind(rule, `forbidden_source_tokens[${index + 1}]`),
      tokens: strings(rule.tokens), scanPaths: paths(rule.scan_paths ?? []), exactFileSetID: rule.exact_file_set_id ?? null,
      allowedPaths: paths(rule.allowed_paths), productionOnly: rule.production_only,
    })),
    forbiddenTestBuildTokens: strings(raw.forbidden_test_build_tokens),
    generatedRootWrites: {
      scanRoots: roots(raw.generated_root_writes.scan_roots), generatedRoots: strings(raw.generated_root_writes.generated_roots),
      writeSymbols: strings(raw.generated_root_writes.write_symbols), allowedPaths: paths(raw.generated_root_writes.allowed_paths),
    },
  };
  const exactFileSetIDs = new Set(manifest.exactFileSets.map((entry) => entry.id));
  for (const rule of manifest.forbiddenSourceTokens) {
    if (rule.exactFileSetID !== null && !exactFileSetIDs.has(rule.exactFileSetID)) {
      throw new Error(`forbidden_source_tokens.${rule.id}.exact_file_set_id references unknown set ${rule.exactFileSetID}`);
    }
  }
  for (const rule of manifest.sqlTableAccess) {
    if (rule.id.endsWith("-projection-storage-access") || rule.id.startsWith("reporting-external-") || ["records-current-envelope-access","timeline-source-owner-access"].includes(rule.id)) {
      for (const grant of [...rule.readAllowedPaths, ...rule.writeAllowedPaths, ...(rule.lockAllowedPaths ?? [])]) {
        if (!grant.endsWith("/*.go") || grant.slice(0, -5).includes("*")) {
          throw new Error(`${rule.id}: production SQL grants require exact package capabilities`);
        }
      }
    }
    const seen = new Set();
    for (const entry of rule.readAllowedStatements) {
      const candidate = `${entry.package}/capability.go`;
      if (entry.package.includes('*') || !pathMatchesAny(candidate, rule.scanPaths) || pathMatchesAny(candidate, rule.readAllowedPaths)) {
        throw new Error(`${rule.id}: statement grant requires an exact scanned package without a blanket read grant`);
      }
      const identity = `${entry.package}:${entry.declaration}:${entry.key}`;
      if (seen.has(identity)) throw new Error(`${rule.id}: duplicate statement grant`);
      seen.add(identity);
    }
  }
  return manifest;
}

function collectFiles(admission, roots, excludes, label) {
  const files = new Map();
  function walk(directory) {
    for (const entry of admission.entries(directory, label)) {
      const relative = `${directory}/${entry.name}`;
      if (isExcluded(relative, excludes)) continue;
      if (entry.isDirectory()) { walk(relative); continue; }
      if (sourceExtensions.has(path.extname(entry.name))) files.set(relative, admission.file(relative, label));
    }
  }
  for (const scanRoot of roots) {
    admission.directory(scanRoot, label);
    walk(scanRoot);
  }
  return [...files.values()].sort((a,b)=>a.relative.localeCompare(b.relative));
}

// Generated Go still defines executable types and SQL. Its exclusion from
// authored-source architecture rules must not remove it from SQL analysis.
export function collectGeneratedSQLSources(root, admission = createSourceAdmission(root)) {
  const files=[];
  function visit(directory) {
    for(const entry of admission.entries(directory, "generated SQL source")) {
      const relative=`${directory}/${entry.name}`;
      if(entry.isDirectory()) {visit(relative);continue;}
      if(!entry.name.endsWith('.go') || entry.name.endsWith('_test.go')) continue;
      const {content}=admission.file(relative, "generated SQL source");
      files.push({relative,content});
    }
  }
  visit('internal/gen');
  return files.sort((a,b)=>a.relative.localeCompare(b.relative));
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

function validatePathPattern(value) {
  const body = value.endsWith("/**") ? value.slice(0, -3) : value;
  if (value.trim() !== value || /^[A-Za-z]:/u.test(body) || /[\\?\[\]{}]/u.test(body) || body.includes("**") ||
      body.split("/").some((segment) => segment === "" || segment === "." || segment === "..")) {
    throw new Error(`unsupported repository-relative path selector: ${value}`);
  }
}

function matchesPattern(relative, pattern) {
  if (!pattern.includes("*")) return relative === pattern;
  const subtree = pattern.endsWith("/**");
  const body = subtree ? pattern.slice(0, -3) : pattern;
  const expression = body.split("*").map((part) => part.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&")).join("[^/]*");
  return new RegExp(`^${expression}${subtree ? "(?:/.*)?" : ""}$`, "u").test(relative);
}

function fileImports(file, importsByPath) {
  const imports = importsByPath.get(file.relative);
  if (!Array.isArray(imports) || imports.some((value) => typeof value !== "string")) throw new Error(`missing parsed imports for ${file.relative}`);
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

function hasImportCapability(file, imported, rule) {
  const receiver = path.posix.dirname(file.relative);
  return (rule.importCapabilities ?? []).some((entry) => entry.import === imported &&
    (file.relative.endsWith("_test.go") ? entry.testPackages : entry.productionPackages).includes(receiver));
}

function checkOwnerPortOnlyImports(files, rules, importsByPath) {
  if (rules.length === 0) return [];
  const violations = [];
  for (const file of files.filter((entry) => entry.relative.endsWith(".go"))) {
    for (const imported of fileImports(file, importsByPath)) {
      for (const rule of rules) {
        if (imported !== rule.import && !imported.startsWith(`${rule.import}/`)) {
          continue;
        }
        if (!hasImportCapability(file, imported, rule) && !rule.allowedImporters.some((pattern) => matchesPattern(file.relative, pattern))) {
          violations.push(violation("owner_port_only_import", file, imported));
        }
      }
    }
  }
  return violations;
}

function checkForbiddenRouteDependencies(files, rules, importsByPath) {
  if (rules.length === 0) return [];
  const violations = [];
  for (const file of files.filter((entry) => entry.relative.endsWith(".go"))) {
    for (const imported of fileImports(file, importsByPath)) {
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

function checkForbiddenGoImports(files, rules, importsByPath) {
  if (rules.length === 0) return [];
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
      for (const imported of fileImports(file, importsByPath)) {
        if (hasImportCapability(file, imported, rule)) continue;
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

function checkGoImportAllowlists(files, rules, importsByPath) {
  if (rules.length === 0) return [];
  const violations = [];
  for (const file of files) {
    for (const rule of rules) {
      if (!isProductionGo(file, rule.productionOnly)) {
        continue;
      }
      if (rule.scanPaths.length > 0 && !pathMatchesAny(file.relative, rule.scanPaths)) {
        continue;
      }
      for (const imported of fileImports(file, importsByPath)) {
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

function checkSourceTableAccess(files, rules, sql) {
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
        if (sql.relations(file).some((access) => access.table === table)) {
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

function checkSQLTableAllowlists(files, rules, sql) {
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
      for (const table of [...new Set(sql.relations(file).map((access) => access.table))].sort()) {
        if (!rule.allowedTables.has(table)) {
          violations.push(violation("sql_table_allowlist", file, table));
        }
      }
    }
  }
  return violations;
}

function checkSQLTableAccess(files, rules, sql, accessScope = "production", completeSourceSet = false) {
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
      const grants = accessScope === "production" ? (rule.readAllowedStatements ?? []).filter((entry) => entry.package === path.posix.dirname(file.relative)) : [];
      const statements = sql.statements(file);
      for (const statement of statements) {
        if (grants.length && statement.accesses.some((access) => access.dynamic)) {
          violations.push(violation("sql_statement_dynamic", file, rule.table, { rule_id: rule.id }));
        }
        const granted = !statement.dynamic && grants.some((entry) => entry.key === statement.key && entry.declaration === statement.declaration);
        for (const access of statement.accesses) {
          if (access.table !== rule.table || access.operation === "lock") continue;
          const allowedPaths = access.operation === "read"
            ? accessScope === "test_fixture" ? rule.testReadAllowedPaths : rule.readAllowedPaths
            : accessScope === "test_fixture" ? rule.testWriteAllowedPaths : rule.writeAllowedPaths;
          if (!pathMatchesAny(file.relative, allowedPaths) && !(access.operation === "read" && granted)) {
            violations.push(violation(access.operation === "read" ? "sql_table_read_access" : "sql_table_write_access", file, rule.table, { rule_id: rule.id, access_scope: accessScope }));
          }
        }
        // Exact grants never authorize a lock, including when legacy path rules
        // elsewhere do not yet constrain locks. Own-table writes remain legal.
        if (grants.length && statement.accesses.some((access) => access.table === rule.table && access.operation === "lock")) {
          violations.push(violation("sql_table_lock_access", file, rule.table, { rule_id: rule.id, access_scope: accessScope }));
        }
      }
      const lockAllowedPaths = accessScope === "test_fixture" ? [] : rule.lockAllowedPaths;
      if (
        lockAllowedPaths !== null &&
        sql.relations(file).some((access) => access.table === rule.table && access.operation === "lock") &&
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
  if (accessScope === "production") {
    if (completeSourceSet) for (const rule of rules) {
      for (const grant of new Set([...rule.readAllowedPaths, ...rule.writeAllowedPaths, ...(rule.lockAllowedPaths ?? [])])) {
        if (grant.endsWith("/*.go") && !sourceFiles.some(file=>matchesPattern(file.relative,grant))) {
          violations.push(violation("sql_package_grant_stale", {relative:grant.slice(0,-5)}, rule.table, {rule_id:rule.id}));
        }
      }
    }
    for (const rule of rules) for (const entry of rule.readAllowedStatements ?? []) {
      const candidates = sourceFiles.filter(file => path.posix.dirname(file.relative) === entry.package);
      if ((completeSourceSet || candidates.length > 0) && !candidates.some(file => sql.statements(file).some(statement => !statement.dynamic && statement.key === entry.key && statement.declaration === entry.declaration))) {
        violations.push(violation("sql_statement_grant_stale", { relative: entry.package }, rule.table, { rule_id: rule.id, declaration: entry.declaration }));
      }
    }
  }
  return violations;
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

function checkCommandRootShape(files, rule) {
  const prefix = `${rule.root}/`;
  return files
    .filter((file) => file.relative.startsWith(prefix) && file.relative.endsWith(".go"))
    .filter((file) => !rule.allowedGoFiles.has(file.relative))
    .map((file) => violation("command_root_shape", file, "unexpected_go_file"));
}

function requireSourceTokenMatchKind(rule, label) {
  if (!["source_text", "go_identifier"].includes(rule?.match_kind)) {
    throw new Error(`${label}.match_kind must be source_text or go_identifier`);
  }
  if (rule.match_kind === "go_identifier" &&
    (!Array.isArray(rule.tokens) || rule.tokens.some((token) =>
      typeof token !== "string" || !/^[\p{L}_][\p{L}\p{Nd}_]*$/u.test(token)))) {
    throw new Error(`${label}.tokens must contain Go identifiers`);
  }
  return rule.match_kind;
}

function checkForbiddenSourceTokens(files, rules, exactFilesByID = new Map()) {
  const violations = [];
  for (const rule of rules) {
    const candidates = rule.exactFileSetID === null
      ? files
      : (exactFilesByID.get(rule.exactFileSetID) ?? []);
    for (const file of candidates) {
      if (!isProductionGo(file, rule.productionOnly)) {
        continue;
      }
      if (rule.exactFileSetID === null && !pathMatchesAny(file.relative, rule.scanPaths)) {
        continue;
      }
      if (pathMatchesAny(file.relative, rule.allowedPaths)) {
        continue;
      }
      const identifiers = rule.matchKind === "go_identifier"
        ? new Set(sanitizeGoStructure(file.content).match(/[\p{L}_][\p{L}\p{Nd}_]*/gu) ?? [])
        : null;
      for (const token of rule.tokens) {
        if (identifiers === null ? file.content.includes(token) : file.relative.endsWith(".go") && identifiers.has(token)) {
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

/** Evaluate explicit source collections; no filesystem or process side effects. */
export function evaluateBoundaryPolicy(manifest, {
  files = [], migrationBoundaryFiles = [], testSupportFiles = [], sqlAdditionalFiles = [], exactFilesByID = new Map(), importsByPath = new Map(), sqlByPath = new Map(), completeSourceSet = false,
} = {}) {
  const sqlFiles=[...files,...sqlAdditionalFiles];
  const cache = new Map();
  const constructionFailures = [];
  const usedForwarders = new Set();
  const usedControls = new Set();
  const sql = {
    statements(file) {
      if (!cache.has(file.relative)) {
        if (file.relative.endsWith(".go")) {
          if (!sqlByPath.has(file.relative)) throw new Error(`missing parsed SQL for ${file.relative}`);
          const statements=[];
          for (const entry of sqlByPath.get(file.relative)) {
            try {
              const grant=manifest.sqlForwardingCapabilities?.find(cap=>path.posix.dirname(file.relative)===cap.package && entry.name===cap.declaration && cap.testSupport===testSupportFiles.some(support=>support.relative===file.relative));
              const qualified=entry.sql.filter(expression=>{
                if(expression.kind!=="forward" || !grant) return true;
                usedForwarders.add(grant);
                return false;
              });
              for(const statement of analyzedSQLStatements([{...entry,sql:qualified}])) {
                if(statement.controls?.length){
                  const authority=manifest.sqlControlCapabilities?.find(cap=>cap.package===path.posix.dirname(file.relative)&&cap.declaration===entry.name&&cap.key===statement.key);
                  if(!authority) constructionFailures.push(violation("sql_control_access",file,entry.name));
                  else usedControls.add(authority);
                }
                statements.push(statement);
              }
            }
            catch(error) { constructionFailures.push(violation("sql_construction_unresolved",file,entry.name,{detail:error.message})); }
          }
          cache.set(file.relative,statements);
        } else cache.set(file.relative,sqlStatements(file.content));
      }
      return cache.get(file.relative);
    },
    relations(file) { return this.statements(file).flatMap(statement=>statement.accesses); },
  };
  if (completeSourceSet || manifest.sqlForwardingCapabilities?.length || sqlAdditionalFiles.length) for (const file of sqlFiles) if (file.relative.endsWith(".go") && !file.relative.endsWith("_test.go")) sql.statements(file);
  const projectionStorageRules = manifest.sqlTableAccess.filter((rule) =>
    rule.id.endsWith("-projection-storage-access"),
  );
  const violations = [
    ...checkOwnerPortOnlyImports(files, manifest.ownerPortOnlyImports, importsByPath),
    ...checkRawNDJSONTargets(files, manifest.rawNDJSONTargets),
    ...checkForbiddenRouteDependencies(files, manifest.forbiddenRouteDependencies, importsByPath),
    ...checkForbiddenGoImports(files, manifest.forbiddenGoImports, importsByPath),
    ...checkGoImportAllowlists(files, manifest.goImportAllowlists, importsByPath),
    ...checkSourceTableAccess(sqlFiles, manifest.sourceTableAccess, sql),
    ...checkForbiddenSourceMappings(files, manifest.forbiddenSourceMappings),
    ...checkSQLTableAllowlists(sqlFiles, manifest.sqlTableAllowlists, sql),
    ...checkSQLTableAccess(sqlFiles, manifest.sqlTableAccess, sql, "production", completeSourceSet),
    ...checkSQLTableAccess(testSupportFiles, projectionStorageRules, sql, "test_fixture"),
    ...checkForbiddenGoCalls(files, manifest.forbiddenGoCalls),
    ...checkCommandRootShape(files, manifest.commandRootShape),
    ...checkForbiddenSourceTokens(files, manifest.forbiddenSourceTokens, exactFilesByID),
    ...checkForbiddenTestBuildTokens(files, manifest.forbiddenTestBuildTokens),
    ...checkGeneratedRootWrites(files, manifest.generatedRootWrites),
    ...checkForbiddenGoImports(
      migrationBoundaryFiles,
      manifest.migrationSourceBoundary.restrictedImports,
      importsByPath,
    ),
    ...checkForbiddenGoDeclarations(
      migrationBoundaryFiles,
      manifest.migrationSourceBoundary.forbiddenDeclarations,
    ),
    ...checkExactGoCalls(migrationBoundaryFiles, manifest.migrationSourceBoundary.exactCalls),
    ...constructionFailures,
    ...(completeSourceSet ? (manifest.sqlControlCapabilities??[]).filter(grant=>!usedControls.has(grant)).map(grant=>({code:"stale_sql_control_grant",path:grant.package,symbol_or_import:grant.declaration})) : []),
    ...(completeSourceSet ? (manifest.sqlForwardingCapabilities??[]).filter(grant=>!usedForwarders.has(grant)).map(grant=>({code:"stale_sql_forwarder",path:grant.package,symbol_or_import:grant.declaration})) : []),
  ].sort((left, right) => {
    return (
      left.code.localeCompare(right.code) ||
      left.path.localeCompare(right.path) ||
      left.symbol_or_import.localeCompare(right.symbol_or_import)
    );
  });
  return {
    violations,
    // Dynamic targets are unproven observations, never physical table names.
    dynamicRelationSources: sqlFiles.filter((file) => cache.has(file.relative))
      .flatMap((file) => sql.relations(file).filter((access) => access.dynamic)
        .map((access) => ({ path: file.relative, operation: access.operation }))),
  };
}

export function inspectBackendBoundaries(options) {
  const manifestRawContent = readFileSync(options.manifest, "utf8");
  const manifest = prepareBoundaryPolicy(JSON.parse(manifestRawContent));
  const supportInventory = readSupportInventory(options.root, options.supportInventory);
  const inventoryScanExcludes = backendRuntimeExcludePatterns(supportInventory);
  const scanExcludes = appendUnique(manifest.scanExcludes, inventoryScanExcludes);
  const admission = createSourceAdmission(options.root);
  const files = collectFiles(admission, manifest.scanRoots, scanExcludes, "backend source root");
  const exactFilesByID = resolveExactFileSets(options.root, manifest.exactFileSets, admission);
  const migrationBoundaryFiles = collectFiles(
    admission,
    manifest.migrationSourceBoundary.scanRoots,
    scanExcludes,
    "migration source root",
  );
  const testSupportRoots = supportInventory.roots
    .filter((entry) => entry.runtimeScan === "excluded" && entry.supportScan === "included")
    .map((entry) => entry.path);
  const testSupportFiles = collectFiles(admission, testSupportRoots, manifest.scanExcludes, "support source root");
  const sqlAdditionalFiles=collectGeneratedSQLSources(options.root, admission);
  return {
    manifestRawContent,
    supportInventoryRaw: supportInventory.raw,
    effectiveScanExcludes: scanExcludes,
    ...evaluateBoundaryPolicy(manifest, { files, migrationBoundaryFiles, testSupportFiles, sqlAdditionalFiles, exactFilesByID,
      importsByPath: inspectGoImports([files, migrationBoundaryFiles]), sqlByPath: inspectGoSQL([[...files,...testSupportFiles,...sqlAdditionalFiles].filter(file=>!file.relative.endsWith("_test.go"))]), completeSourceSet: true,
    }),
  };
}
