import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import ts from "typescript";

const issueArrayKeys = new Set([
  "boundary_violations",
  "boundaryviolations",
  "circular_dependencies",
  "circulardependencies",
  "duplicate_exports",
  "duplicateexports",
  "policy_violations",
  "policyviolations",
  "re_export_cycles",
  "reexportcycles",
  "stale_suppressions",
  "stalesuppressions",
  "test_only_dependencies",
  "testonlydependencies",
  "type_only_dependencies",
  "typeonlydependencies",
  "unlisted_dependencies",
  "unlisteddependencies",
  "unresolved_imports",
  "unresolvedimports",
  "unused_dependencies",
  "unuseddependencies",
  "unused_dev_dependencies",
  "unuseddevdependencies",
  "unused_exports",
  "unusedexports",
  "unused_files",
  "unusedfiles",
  "unused_optional_dependencies",
  "unusedoptionaldependencies",
  "unused_types",
  "unusedtypes",
]);

function normalizeKey(key) {
  return String(key ?? "").replaceAll("-", "_").replaceAll("_", "").toLowerCase();
}

function increment(map, key, count = 1) {
  const normalized = String(key ?? "").trim() || "unknown";
  map[normalized] = (map[normalized] ?? 0) + count;
}

function mergeCounts(left, right) {
  for (const [key, value] of Object.entries(right ?? {})) {
    increment(left, key, value);
  }
}

function classifyRule(item, fallback) {
  return (
    item?.rule_id ??
    item?.ruleId ??
    item?.rule ??
    item?.issue_type ??
    item?.issueType ??
    item?.kind ??
    fallback
  );
}

function classifySeverity(item) {
  return item?.severity ?? item?.level ?? item?.rule_severity ?? "unknown";
}

export function collectIssueCounts(value, result = { issueCount: 0, byRule: {}, bySeverity: {}, rawCountFields: {} }) {
  if (!value || typeof value !== "object") {
    return result;
  }
  if (Array.isArray(value)) {
    for (const item of value) {
      collectIssueCounts(item, result);
    }
    return result;
  }
  for (const [key, nested] of Object.entries(value)) {
    const normalizedKey = normalizeKey(key);
    if (Array.isArray(nested) && issueArrayKeys.has(normalizedKey)) {
      increment(result.rawCountFields, key, nested.length);
      result.issueCount += nested.length;
      for (const item of nested) {
        increment(result.byRule, classifyRule(item, key));
        increment(result.bySeverity, classifySeverity(item));
      }
      continue;
    }
    collectIssueCounts(nested, result);
  }
  return result;
}

export async function runFallowBatch(specs, execute) {
  const settled = await Promise.allSettled(specs.map((spec) => Promise.resolve().then(() => execute(
    spec.reportRoot,
    spec.name,
    spec.args,
    spec.outputFile,
  ))));
  return settled.map((result, index) => {
    if (result.status === "fulfilled") return result.value;
    const spec = specs[index];
    return {
      name: spec.name,
      command: ["fallow", ...spec.args],
      outputFile: spec.outputFile,
      stdoutFile: path.join(spec.reportRoot, `${spec.name}.stdout.log`),
      stderrFile: path.join(spec.reportRoot, `${spec.name}.stderr.log`),
      status: "fail",
      exitCode: 127,
    };
  });
}

// Validate the fields consumed by this profile, allowing unrelated analyzer fields
// to evolve without adding aliases or silently accepting missing owned data.
export function validateFallowReport(report, kind) {
  if (!report || typeof report !== "object" || Array.isArray(report) || report.error || report.kind !== kind) {
    throw new Error(`Fallow report must contain a valid ${kind} result`);
  }
  const arrays = {
    "dead-code": ["unused_files", "unused_exports", "unused_types"],
    dupes: ["clone_groups", "clone_families"],
    health: ["findings"],
  }[kind];
  if (!arrays) throw new Error(`unsupported Fallow report kind ${kind}`);
  for (const key of arrays) {
    if (!Array.isArray(report[key]) || report[key].some((item) => !item || typeof item !== "object" || Array.isArray(item))) {
      throw new Error(`Fallow report has invalid ${key}`);
    }
  }
}

function findingsForRule(report, rule) {
  if (!Array.isArray(report?.[rule])) throw new Error(`Fallow report is missing ${rule}`);
  return report[rule].map((finding) => {
    if (typeof finding?.path !== "string" ||
        finding.path.startsWith("/") || finding.path.includes("\\") ||
        finding.path.split("/").some((part) => !part || part === "." || part === "..")) {
      throw new Error(`Fallow ${rule} finding has an invalid relative path`);
    }
    const symbol = rule === "unused_files" ? null : finding.export_name;
    if (symbol !== null && (typeof symbol !== "string" || !symbol.trim())) {
      throw new Error(`Fallow ${rule} finding has no export name`);
    }
    return { path: finding.path, rule, symbol };
  });
}

function resolveTypeScriptModule(entrypoint, specifier) {
  const base = path.resolve(path.dirname(entrypoint), specifier);
  const candidates = [
    base,
    `${base}.ts`,
    `${base}.tsx`,
    path.join(base, "index.ts"),
    path.join(base, "index.tsx"),
  ];
  return candidates.find((candidate) => existsSync(candidate)) ?? null;
}

export function readPackageSurfaceBindings(root, policy) {
  const bindings = new Map();
  for (const packageSurface of policy.packages) {
    const entrypoint = path.join(root, packageSurface.entrypoint);
    const sourceFile = ts.createSourceFile(
      entrypoint,
      readFileSync(entrypoint, "utf8"),
      ts.ScriptTarget.Latest,
      true,
      entrypoint.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
    );
    for (const statement of sourceFile.statements) {
      if (!ts.isExportDeclaration(statement)) {
        const isExported = ts.canHaveModifiers(statement) &&
          (ts.getModifiers(statement) ?? []).some(
            (modifier) => modifier.kind === ts.SyntaxKind.ExportKeyword,
          );
        if (!isExported) continue;
        const declarationNames = ts.isVariableStatement(statement)
          ? statement.declarationList.declarations.flatMap((declaration) =>
              ts.isIdentifier(declaration.name) ? [declaration.name.text] : [],
            )
          : "name" in statement && statement.name && ts.isIdentifier(statement.name)
            ? [statement.name.text]
            : [];
        for (const symbol of declarationNames) {
          bindings.set(`${packageSurface.entrypoint}\u0000${symbol}`, {
            entrypoint: packageSurface.entrypoint,
            package_name: packageSurface.package_name,
            public_symbol: symbol,
            source_symbol: symbol,
          });
        }
        continue;
      }
      if (!statement.exportClause || !ts.isNamedExports(statement.exportClause)) {
        throw new Error(
          `${packageSurface.entrypoint} must not use wildcard or namespace exports`,
        );
      }
      if (!statement.moduleSpecifier || !ts.isStringLiteral(statement.moduleSpecifier)) {
        continue;
      }
      const specifier = statement.moduleSpecifier.text;
      if (!specifier.startsWith(".")) continue;
      const resolved = resolveTypeScriptModule(entrypoint, specifier);
      if (resolved === null) {
        throw new Error(
          `${packageSurface.entrypoint} cannot resolve re-export ${specifier}`,
        );
      }
      const sourcePath = path.relative(root, resolved).replaceAll("\\", "/");
      if (sourcePath.includes("/generated/")) continue;
      for (const element of statement.exportClause.elements) {
        const sourceSymbol = (element.propertyName ?? element.name).text;
        bindings.set(`${sourcePath}\u0000${sourceSymbol}`, {
          entrypoint: packageSurface.entrypoint,
          package_name: packageSurface.package_name,
          public_symbol: element.name.text,
          source_symbol: sourceSymbol,
        });
      }
    }
  }
  return bindings;
}

export function collectBlockingPackageSurfaceFindings(report, policy, bindings) {
  return policy.rules.flatMap((rule) => findingsForRule(report, rule).flatMap((finding) => {
    const binding = bindings.get(`${finding.path}\u0000${finding.symbol}`);
    return binding === undefined ? [] : [{
      entrypoint: binding.entrypoint,
      package_name: binding.package_name,
      path: finding.path,
      rule,
      source_symbol: binding.source_symbol,
      symbol: binding.public_symbol,
    }];
  }));
}

export function collectBlockingWebSurfaceFindings(report, policy) {
  return policy.rules.flatMap((rule) => findingsForRule(report, rule)
    .filter((finding) => finding.path.startsWith(`${policy.root}/`)));
}

export function summarizeFallowReports(reports) {
  const totals = { reports: reports.length, issue_count: 0, by_rule: {}, by_severity: {} };
  for (const report of reports) {
    totals.issue_count += report.issue_count;
    mergeCounts(totals.by_rule, report.by_rule);
    mergeCounts(totals.by_severity, report.by_severity);
  }
  return totals;
}
