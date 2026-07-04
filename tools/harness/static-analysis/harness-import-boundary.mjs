import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

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
const approvedPlanningImportReasons = new Map([
  [
    "tools/harness/backend/go-duration-artifacts.mjs=>tools/harness/planning/backend-shard-plan.mjs",
    "backend duration artifacts consume planning-owned shard row discovery",
  ],
  [
    "tools/harness/backend/go-shard-plan-cli.mjs=>tools/harness/planning/backend-shard-plan.mjs",
    "backend shard CLI keeps row discovery in the planning adapter",
  ],
  [
    "tools/harness/backend/go-shard-plan.mjs=>tools/harness/planning/target-plan.mjs",
    "executable compatibility path for private shard-plan inspection",
  ],
  [
    "tools/harness/backend/go-target-plan-coverage-cli.mjs=>tools/harness/planning/backend-target-plan.mjs",
    "backend target coverage consumes planning-owned normalized target rows",
  ],
  [
    "tools/harness/backend/go-target-runner.mjs=>tools/harness/planning/backend-target-plan.mjs",
    "backend target runner consumes planning-owned normalized target rows",
  ],
  [
    "tools/harness/backend/go-test-duration-baseline-coverage-cli.mjs=>tools/harness/planning/backend-shard-plan.mjs",
    "backend duration coverage consumes planning-owned shard row discovery",
  ],
  [
    "tools/harness/backend/postgres-fixture-budget-cli.mjs=>tools/harness/planning/backend-target-plan.mjs",
    "postgres fixture budget check consumes planning-owned normalized target rows",
  ],
  [
    "tools/harness/browser/browser-shard-plan.mjs=>tools/harness/planning/phase-manifest.mjs",
    "executable compatibility path for private browser shard discovery",
  ],
  [
    "tools/harness/scheduler/adapters/backend.mjs=>tools/harness/planning/backend-shard-plan.mjs",
    "scheduler backend adapter owns access to planning-backed Go shard discovery",
  ],
  [
    "tools/harness/scheduler/adapters/planning.mjs=>tools/harness/planning/summary-topology.mjs",
    "scheduler planning adapter owns summary topology access",
  ],
  [
    "tools/harness/scheduler/adapters/planning.mjs=>tools/harness/planning/target-plan.mjs",
    "scheduler planning adapter owns target descriptor access",
  ],
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
      if (entry.isFile() && entry.name.endsWith(".mjs")) {
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

function collectEdges(root, files) {
  const edges = [];
  for (const source of files) {
    const content = readFileSync(path.join(root, source), "utf8");
    for (const specifier of importSpecifiers(content)) {
      const target = resolveRelativeImport(root, source, specifier);
      if (!target) {
        continue;
      }
      edges.push({
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

function approvedPlanningImportReason(edge) {
  return approvedPlanningImportReasons.get(`${edge.source}=>${edge.target}`) ?? "";
}

function planningImportViolation(edge) {
  const sourceSubsystem = subsystemForPath(edge.source);
  return {
    rule: "forbidden_planning_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} imports ${edge.target}; ${sourceSubsystem} modules must use an ` +
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
      `${edge.source} imports ${edge.target}; harness code must use the owning ` +
      "contract, output, execution, finalization, diagnostics, smoke, frontend, " +
      "browser, backend, scheduler, planning, or generated-artifact entrypoint.",
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
      .filter((edge) => isExecutionToPlanningEdge(edge) && !approvedPlanningImportReason(edge))
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

export function collectHarnessImportBoundaryViolations(
  root = defaultRepoRoot,
  { scanRoot = "tools/harness" } = {},
) {
  const resolvedRoot = path.resolve(root);
  const files = sourceFiles(resolvedRoot, scanRoot);
  const edges = collectEdges(resolvedRoot, files);
  const edgeViolations = edges
    .filter((edge) => isExecutionToPlanningEdge(edge))
    .filter((edge) => !approvedPlanningImportReason(edge))
    .map(planningImportViolation);
  const privateCoreViolations = edges
    .filter((edge) => isPrivateCoreImport(edge))
    .map(privateCoreImportViolation);
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
    violations: [...edgeViolations, ...privateCoreViolations, ...sccViolations],
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
