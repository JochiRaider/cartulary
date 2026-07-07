import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  browserPrivateImportAllowedSourcePaths,
  ownerFacadePathLists,
  unsupportedPrivateHelperRules,
} from "./harness-helper-ownership-registry.mjs";

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
const executionSubsystems = new Set(["backend", "browser", "frontend", "scheduler"]);
const backendOwnerFacadePaths = new Set(ownerFacadePathLists.backend);
const frontendOwnerFacadePaths = new Set(ownerFacadePathLists.frontend);
const browserOwnerFacadePaths = new Set(ownerFacadePathLists.browser);
const durationAccountingOwnerFacadePaths = new Set(ownerFacadePathLists.duration_accounting);
const phaseAccountingOwnerFacadePaths = new Set(ownerFacadePathLists.phase_accounting);
const serviceBackedExecutionOwnerFacadePaths = new Set(
  ownerFacadePathLists.service_backed_execution,
);
const schedulerOwnerFacadePaths = new Set(ownerFacadePathLists.scheduler);
const schedulerDiagnosticsOwnerFacadePaths = new Set(
  ownerFacadePathLists.scheduler_diagnostics,
);
const testOutputOwnerFacadePaths = new Set(ownerFacadePathLists.test_output);
const executionRuntimeOwnerFacadePaths = new Set(ownerFacadePathLists.execution_runtime);
const commandSurfaceOwnerFacadePaths = new Set(ownerFacadePathLists.command_surface);
const browserPrivateImportAllowedSources = new Set(browserPrivateImportAllowedSourcePaths);

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
  const specifiers = [];
  const patterns = [
    /\bimport\s+(?:[^"'()]*?\s+from\s*)?["']([^"']+)["']/gu,
    /\bexport\s+(?:[^"']*?\s+from\s*)["']([^"']+)["']/gu,
    /\bimport\s*\(\s*["']([^"']+)["']\s*\)/gu,
  ];
  for (const pattern of patterns) {
    for (const match of content.matchAll(pattern)) {
      specifiers.push(match[1]);
    }
  }
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

function isExecutionToPlanningEdge(edge) {
  return (
    executionSubsystems.has(subsystemForPath(edge.source)) &&
    edge.target.startsWith("tools/harness/planning/")
  );
}

function planningImportViolation(edge) {
  const sourceSubsystem = subsystemForPath(edge.source);
  return {
    rule: "forbidden_planning_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} ${edgeVerb(edge)} ${edge.target}; ${sourceSubsystem} modules must use an ` +
      "approved planning adapter entrypoint for normalized planning data.",
  };
}

function isPrivateCoreImport(edge) {
  return edge.target.startsWith("tools/harness/core/");
}

function privateCoreImportViolation(edge) {
  return {
    rule: "forbidden_private_core_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} ${edgeVerb(edge)} ${edge.target}; harness code must use the owning ` +
      "contract, output, execution, finalization, diagnostics, smoke, frontend, " +
      "browser, backend, scheduler, planning, or generated-artifact entrypoint.",
  };
}

function unsupportedPrivateHelperRuleForPath(target) {
  return unsupportedPrivateHelperRules.find(
    (rule) =>
      rule.exact.includes(target) ||
      rule.prefixes.some((prefix) => target.startsWith(prefix)),
  );
}

function isUnsupportedPrivateHelperPath(target) {
  return unsupportedPrivateHelperRuleForPath(target) !== undefined;
}

function isUnsupportedPrivateHelperImport(edge) {
  return isUnsupportedPrivateHelperPath(edge.target);
}

function unsupportedPrivateHelperImportViolation(edge) {
  const matchedRule = unsupportedPrivateHelperRuleForPath(edge.target);
  return {
    rule: "forbidden_unsupported_private_helper_import",
    unsupported_private_rule: matchedRule?.id ?? "unknown",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} ${edgeVerb(edge)} ${edge.target}; this helper path is unsupported ` +
      "private compatibility surface and callers must use the declared owner facade.",
  };
}

function isPrivateBackendImplementationImport(edge) {
  if (!edge.target.startsWith("tools/harness/backend/")) {
    return false;
  }
  if (backendOwnerFacadePaths.has(edge.target)) {
    return false;
  }
  if (isUnsupportedPrivateHelperPath(edge.target)) {
    return false;
  }
  return subsystemForPath(edge.source) !== "backend";
}

function privateBackendImplementationImportViolation(edge) {
  return {
    rule: "forbidden_private_backend_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} ${edgeVerb(edge)} ${edge.target}; non-backend harness code must use ` +
      "the backend target, shard, or duration owner facade.",
  };
}

function isPrivateFrontendCatchAllImport(edge) {
  return (
    edge.target.startsWith("tools/harness/frontend/") &&
    !frontendOwnerFacadePaths.has(edge.target) &&
    !isUnsupportedPrivateHelperPath(edge.target)
  );
}

function privateFrontendCatchAllImportViolation(edge) {
  return {
    rule: "forbidden_private_frontend_catch_all_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} ${edgeVerb(edge)} ${edge.target}; frontend harness helpers must use ` +
      "the declared phase-accounting, output, execution, readiness, browser, " +
      "generated-artifact, or static-analysis owner facade.",
  };
}

function isPrivatePhaseAccountingImplementationImport(edge) {
  if (!edge.target.startsWith("tools/harness/phase-accounting/")) {
    return false;
  }
  if (phaseAccountingOwnerFacadePaths.has(edge.target)) {
    return false;
  }
  if (isUnsupportedPrivateHelperPath(edge.target)) {
    return false;
  }
  return subsystemForPath(edge.source) !== "phase-accounting";
}

function privatePhaseAccountingImplementationImportViolation(edge) {
  return {
    rule: "forbidden_private_phase_accounting_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} ${edgeVerb(edge)} ${edge.target}; non-owner harness code must use ` +
      "the declared phase-accounting facade for phase, frontend row, or planner contracts.",
  };
}

function isPrivateBrowserImplementationImport(edge) {
  if (!edge.target.startsWith("tools/harness/browser/")) {
    return false;
  }
  if (browserOwnerFacadePaths.has(edge.target)) {
    return false;
  }
  if (subsystemForPath(edge.source) === "browser") {
    return false;
  }
  if (browserPrivateImportAllowedSources.has(edge.source)) {
    return false;
  }
  return true;
}

function privateBrowserImplementationImportViolation(edge) {
  return {
    rule: "forbidden_private_browser_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} ${edgeVerb(edge)} ${edge.target}; non-owner harness code must use ` +
      "the declared browser owner facade for browser harness contracts.",
  };
}

function isPrivateBrowserImportFromScheduler(edge) {
  return (
    subsystemForPath(edge.source) === "scheduler" &&
    edge.source !== "tools/harness/scheduler/adapters/browser.mjs" &&
    edge.target.startsWith("tools/harness/browser/")
  );
}

function privateBrowserImportFromSchedulerViolation(edge) {
  return {
    rule: "forbidden_scheduler_private_browser_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} ${edgeVerb(edge)} ${edge.target}; scheduler code must use ` +
      "tools/harness/scheduler/adapters/browser.mjs for browser harness contracts.",
  };
}

function adjacencyFromEdges(files, edges) {
  const adjacency = new Map(files.map((file) => [file, []]));
  for (const edge of edges) {
    if (!adjacency.has(edge.source)) {
      adjacency.set(edge.source, []);
    }
    if (!adjacency.has(edge.target)) {
      adjacency.set(edge.target, []);
    }
    adjacency.get(edge.source).push(edge.target);
  }
  for (const targets of adjacency.values()) {
    targets.sort(sortStrings);
  }
  return adjacency;
}

function stronglyConnectedComponents(adjacency) {
  let index = 0;
  const indexes = new Map();
  const lowLinks = new Map();
  const stack = [];
  const onStack = new Set();
  const components = [];

  function visit(node) {
    indexes.set(node, index);
    lowLinks.set(node, index);
    index += 1;
    stack.push(node);
    onStack.add(node);

    for (const target of adjacency.get(node) ?? []) {
      if (!indexes.has(target)) {
        visit(target);
        lowLinks.set(node, Math.min(lowLinks.get(node), lowLinks.get(target)));
        continue;
      }
      if (onStack.has(target)) {
        lowLinks.set(node, Math.min(lowLinks.get(node), indexes.get(target)));
      }
    }

    if (lowLinks.get(node) !== indexes.get(node)) {
      return;
    }

    const component = [];
    while (stack.length > 0) {
      const member = stack.pop();
      onStack.delete(member);
      component.push(member);
      if (member === node) {
        break;
      }
    }
    components.push(component.sort(sortStrings));
  }

  for (const node of [...adjacency.keys()].sort(sortStrings)) {
    if (!indexes.has(node)) {
      visit(node);
    }
  }
  return components;
}

function forbiddenCrossSubsystemSccs(files, edges) {
  const adjacency = adjacencyFromEdges(files, edges);
  const edgeKeys = new Set(
    edges
      .filter((edge) => isExecutionToPlanningEdge(edge))
      .map((edge) => `${edge.source}=>${edge.target}`),
  );
  const byComponent = [];
  for (const component of stronglyConnectedComponents(adjacency)) {
    if (component.length < 2) {
      continue;
    }
    const members = new Set(component);
    const hasPlanning = component.some((file) => subsystemForPath(file) === "planning");
    const hasExecution = component.some((file) => executionSubsystems.has(subsystemForPath(file)));
    if (!hasPlanning || !hasExecution) {
      continue;
    }
    const forbiddenEdges = edges
      .filter((edge) => members.has(edge.source) && members.has(edge.target))
      .filter((edge) => edgeKeys.has(`${edge.source}=>${edge.target}`))
      .map((edge) => ({
        source: edge.source,
        target: edge.target,
      }));
    if (forbiddenEdges.length === 0) {
      continue;
    }
    byComponent.push({
      rule: "forbidden_cross_subsystem_scc",
      files: component,
      forbidden_edges: forbiddenEdges,
      message:
        "forbidden harness import cycle crosses planning and execution subsystems: " +
        component.join(", "),
    });
  }
  return byComponent.sort((left, right) => sortStrings(left.files[0], right.files[0]));
}

function unsupportedPrivateHelperPatterns() {
  return unsupportedPrivateHelperRules
    .flatMap((rule) => [
      ...rule.exact,
      ...rule.prefixes.map((prefix) => `${prefix}**`),
    ])
    .sort(sortStrings);
}

function unsupportedPrivateRuleReport() {
  return unsupportedPrivateHelperRules.map((rule) => ({
    id: rule.id,
    exact: [...rule.exact].sort(sortStrings),
    prefixes: [...rule.prefixes].sort(sortStrings),
  }));
}

function ownerFacadeReport() {
  return Object.fromEntries(
    Object.entries(ownerFacadePathLists)
      .sort(([left], [right]) => sortStrings(left, right))
      .map(([owner, paths]) => [owner, [...paths].sort(sortStrings)]),
  );
}

export function collectHarnessImportBoundaryViolations(
  root = defaultRepoRoot,
  { scanRoot = "tools/harness" } = {},
) {
  const resolvedRoot = path.resolve(root);
  const files = sourceFiles(resolvedRoot, scanRoot);
  const edges = collectEdges(resolvedRoot, files);
  const edgeViolations = edges
    .filter((edge) => isExecutionToPlanningEdge(edge))
    .map(planningImportViolation);
  const privateCoreViolations = edges
    .filter((edge) => isPrivateCoreImport(edge))
    .map(privateCoreImportViolation);
  const unsupportedHelperViolations = edges
    .filter((edge) => isUnsupportedPrivateHelperImport(edge))
    .map(unsupportedPrivateHelperImportViolation);
  const privateBackendViolations = edges
    .filter((edge) => isPrivateBackendImplementationImport(edge))
    .map(privateBackendImplementationImportViolation);
  const privateFrontendViolations = edges
    .filter((edge) => isPrivateFrontendCatchAllImport(edge))
    .map(privateFrontendCatchAllImportViolation);
  const privatePhaseAccountingViolations = edges
    .filter((edge) => isPrivatePhaseAccountingImplementationImport(edge))
    .map(privatePhaseAccountingImplementationImportViolation);
  const privateBrowserViolations = edges
    .filter((edge) => isPrivateBrowserImplementationImport(edge))
    .map(privateBrowserImplementationImportViolation);
  const privateSchedulerBrowserViolations = edges
    .filter((edge) => isPrivateBrowserImportFromScheduler(edge))
    .map(privateBrowserImportFromSchedulerViolation);
  const forbiddenSccs = forbiddenCrossSubsystemSccs(files, edges);
  const sccViolations = forbiddenSccs.map((scc) => ({
    rule: scc.rule,
    source: scc.files[0],
    target: scc.files[scc.files.length - 1],
    message: scc.message,
  }));
  return {
    root: resolvedRoot,
    files,
    edges,
    owner_facades: ownerFacadeReport(),
    unsupported_private_helpers: unsupportedPrivateHelperPatterns(),
    unsupported_private_rules: unsupportedPrivateRuleReport(),
    violations: [
      ...edgeViolations,
      ...privateCoreViolations,
      ...unsupportedHelperViolations,
      ...privateBackendViolations,
      ...privateFrontendViolations,
      ...privatePhaseAccountingViolations,
      ...privateBrowserViolations,
      ...privateSchedulerBrowserViolations,
      ...sccViolations,
    ],
    forbidden_sccs: forbiddenSccs,
  };
}

export function assertHarnessImportBoundary(root = defaultRepoRoot, options = {}) {
  const report = collectHarnessImportBoundaryViolations(root, options);
  if (report.violations.length === 0) {
    return report;
  }
  const details = report.violations.map((violation) => `- ${violation.message}`).join("\n");
  throw new Error(`harness import boundary violations:\n${details}`);
}
