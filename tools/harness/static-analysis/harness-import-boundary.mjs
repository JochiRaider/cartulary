import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import ts from "typescript";

import {
  allowedPrivateImportSources,
  loadHarnessHelperOwnership,
  ownerFacadePathLists,
} from "./harness-helper-ownership.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const defaultRepoRoot = path.resolve(scriptDir, "../../..");
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
const knownHarnessOwnerRoots = new Set([
  "backend", "browser", "command-surface", "contract", "diagnostics",
  "evidence-accounting", "execution", "finalization", "fixtures",
  "generated-artifacts", "observability", "output", "performance-fixture",
  "readiness", "runtime", "scheduler", "services", "smoke", "static-analysis",
  "test-catalog", "test-support", "tests", "workspace",
]);
const protectedOwnerGroups = new Map([
  ["backend", "backend"], ["browser", "browser"],
  ["test-catalog", "test_catalog"], ["evidence-accounting", "evidence_accounting"],
]);

function normalizePath(value) {
  return value.split(path.sep).join("/");
}

function repoRelative(root, value) {
  return normalizePath(path.relative(root, value));
}

function sortStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function edgeVerb(edge) {
  return edge.kind === "shell_source" ? "sources" : "imports";
}

function sourceFiles(root, scanRoot = "tools/harness") {
  const absoluteScanRoot = path.join(root, scanRoot);
  if (!existsSync(absoluteScanRoot)) {
    return [];
  }
  const files = [];
  const stack = [absoluteScanRoot];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      if (ignoredDirectoryNames.has(entry.name)) {
        continue;
      }
      const absolutePath = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(absolutePath);
        continue;
      }
      if (entry.isFile() && (entry.name.endsWith(".mjs") || entry.name.endsWith(".sh"))) {
        files.push(repoRelative(root, absolutePath));
      }
    }
  }
  return files.sort(sortStrings);
}

function importSpecifiers(content) {
  const source = ts.createSourceFile("harness.mjs", content, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
  const specifiers = [];
  function visit(node) {
    if ((ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) &&
        node.moduleSpecifier && ts.isStringLiteral(node.moduleSpecifier)) {
      specifiers.push(node.moduleSpecifier.text);
    } else if (ts.isCallExpression(node) && node.expression.kind === ts.SyntaxKind.ImportKeyword &&
               node.arguments.length === 1 && ts.isStringLiteral(node.arguments[0])) {
      specifiers.push(node.arguments[0].text);
    }
    ts.forEachChild(node, visit);
  }
  visit(source);
  return specifiers;
}

function shellSourceSpecifiers(content) {
  const specifiers = [];
  const seen = new Set();
  const patterns = [
    /^[ \t]*(?:source|\.)[ \t]+(["'])([^"']+)\1/gmu,
    /^[ \t]*(?:source|\.)[ \t]+([^ \t\r\n#;]+)/gmu,
  ];
  for (const pattern of patterns) {
    for (const match of content.matchAll(pattern)) {
      const specifier = String(match[2] ?? match[1]).replace(/^["']|["']$/gu, "");
      if (!seen.has(specifier)) {
        seen.add(specifier);
        specifiers.push(specifier);
      }
    }
  }
  return specifiers;
}

function resolveRelativeImport(root, importerRel, specifier) {
  if (!specifier.startsWith(".")) {
    return "";
  }
  const importer = path.join(root, importerRel);
  const rawTarget = path.resolve(path.dirname(importer), specifier);
  const candidates = [
    rawTarget,
    `${rawTarget}.mjs`,
    `${rawTarget}.js`,
    path.join(rawTarget, "index.mjs"),
    path.join(rawTarget, "index.js"),
  ];
  const target = candidates.find((candidate) => {
    if (!existsSync(candidate)) {
      return false;
    }
    return statSync(candidate).isFile();
  }) ?? rawTarget;
  const relative = repoRelative(root, target);
  if (!relative.startsWith("tools/harness/")) {
    return "";
  }
  return relative;
}

function repoLocalHarnessPathFromShellSpecifier(specifier) {
  const marker = "tools/harness/";
  const index = specifier.indexOf(marker);
  if (index < 0) {
    return "";
  }
  return specifier.slice(index).replace(/[)"'`;]+$/gu, "");
}

function resolveShellSource(root, importerRel, specifier) {
  const raw = String(specifier ?? "").trim();
  if (!raw || raw.includes("*")) {
    return "";
  }
  if (raw.startsWith(".")) {
    const importer = path.join(root, importerRel);
    const target = path.resolve(path.dirname(importer), raw);
    const relative = repoRelative(root, target);
    return relative.startsWith("tools/harness/") ? relative : "";
  }
  if (path.isAbsolute(raw)) {
    const relative = repoRelative(root, raw);
    return relative.startsWith("tools/harness/") ? relative : "";
  }
  if (raw.startsWith("tools/harness/")) {
    return raw;
  }
  return repoLocalHarnessPathFromShellSpecifier(raw);
}

function collectEdges(root, files) {
  const edges = [];
  for (const source of files) {
    const content = readFileSync(path.join(root, source), "utf8");
    const isShell = source.endsWith(".sh");
    const specifiers = isShell ? shellSourceSpecifiers(content) : importSpecifiers(content);
    for (const specifier of specifiers) {
      const target = isShell
        ? resolveShellSource(root, source, specifier)
        : resolveRelativeImport(root, source, specifier);
      if (!target) {
        continue;
      }
      edges.push({
        kind: isShell ? "shell_source" : "js_import",
        source,
        specifier,
        target,
      });
    }
  }
  return edges.sort((left, right) => {
    const sourceOrder = sortStrings(left.source, right.source);
    if (sourceOrder !== 0) {
      return sourceOrder;
    }
    return sortStrings(left.target, right.target);
  });
}

function subsystemForPath(repoPath) {
  const parts = repoPath.split("/");
  if (parts[0] !== "tools" || parts[1] !== "harness") {
    return "";
  }
  return parts[2] ?? "";
}

export function collectHarnessImportBoundaryViolations(
  root = defaultRepoRoot,
  { scanRoot = "tools/harness" } = {},
) {
  const resolvedRoot = path.resolve(root);
  const ownership = loadHarnessHelperOwnership(resolvedRoot);
  const ownerFacades = ownerFacadePathLists(ownership);
  const files = sourceFiles(resolvedRoot, scanRoot);
  const edges = collectEdges(resolvedRoot, files);
  const violations = [];
  for (const ownerRoot of new Set(files.filter((file) => file.split("/").length > 3).map(subsystemForPath))) {
    if (!knownHarnessOwnerRoots.has(ownerRoot)) violations.push({
      rule: "forbidden_unknown_harness_owner_root", source: `tools/harness/${ownerRoot}`,
      target: "tools/harness", message: `${ownerRoot} is not a current semantic harness owner root`,
    });
  }
  for (const edge of edges) {
    const sourceOwner = subsystemForPath(edge.source);
    const targetOwner = subsystemForPath(edge.target);
    let rule;
    if (targetOwner === "core" || targetOwner === "frontend") {
      rule = "forbidden_private_catch_all_import";
    } else if (sourceOwner === "scheduler" && targetOwner === "browser" &&
               !allowedPrivateImportSources(ownership, "browser").has(edge.source)) {
      rule = "forbidden_scheduler_private_browser_import";
    } else if (protectedOwnerGroups.has(targetOwner) && sourceOwner !== targetOwner) {
      const group = protectedOwnerGroups.get(targetOwner);
      if (!(ownerFacades[group] ?? []).includes(edge.target) &&
          !allowedPrivateImportSources(ownership, group).has(edge.source)) {
        rule = `forbidden_private_${group}_import`;
      }
    }
    if (rule) violations.push({ rule, source: edge.source, target: edge.target,
      message: `${edge.source} ${edgeVerb(edge)} ${edge.target}; use the declared semantic owner facade` });
  }
  return { root: resolvedRoot, files, edges, owner_facades: ownerFacades, violations };
}
