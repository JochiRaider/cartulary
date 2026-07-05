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
const backendOwnerFacadePaths = new Set([
  "tools/harness/backend/backend-duration-accounting.mjs",
  "tools/harness/backend/backend-shard-plan.mjs",
  "tools/harness/backend/backend-target-execution.mjs",
  "tools/harness/backend/backend-target-plan.mjs",
]);
const frontendOwnerFacadePaths = new Set([
  "tools/harness/browser/accessibility-summary-cli.mjs",
  "tools/harness/execution/run-frontend-unit.sh",
  "tools/harness/execution/run-vitest-manifest-phase.sh",
  "tools/harness/execution/run-vitest-phase.sh",
  "tools/harness/generated-artifacts/design-tokens/index.mjs",
  "tools/harness/phase-accounting/frontend/index.mjs",
  "tools/harness/phase-accounting/frontend-readiness.mjs",
  "tools/harness/readiness/build-web-artifact.sh",
  "tools/harness/readiness/embed-web-assets.sh",
  "tools/harness/readiness/frontend-install.sh",
  "tools/harness/readiness/frontend-toolchain.sh",
  "tools/harness/static-analysis/font-bundle-check-cli.mjs",
]);
const browserOwnerFacadePaths = new Set([
  "tools/harness/browser/accessibility-summary-cli.mjs",
  "tools/harness/browser/browser-batch-manifest.mjs",
  "tools/harness/browser/browser-duration-accounting.mjs",
  "tools/harness/browser/browser-shard-plan.mjs",
  "tools/harness/output/test-output/playwright-artifacts.mjs",
  "tools/harness/scheduler/adapters/browser.mjs",
]);
const durationAccountingOwnerFacadePaths = new Set([
  "tools/harness/duration-accounting/index.mjs",
  "tools/harness/duration-accounting/duration-drift.mjs",
  "tools/harness/duration-accounting/duration-baseline-cli.mjs",
  "tools/harness/duration-accounting/target-duration-baselines.mjs",
]);
const phaseAccountingOwnerFacadePaths = new Set([
  "tools/harness/phase-accounting/phase-slice-plan.mjs",
]);
const serviceBackedExecutionOwnerFacadePaths = new Set([
  "tools/harness/execution/service-backed/schedule-planning.mjs",
]);
const schedulerOwnerFacadePaths = new Set([
  "tools/harness/scheduler/phase-slice-execution.mjs",
  "tools/harness/scheduler/scheduler-runner.mjs",
  "tools/harness/scheduler/scheduler-family-contract.mjs",
  "tools/harness/scheduler/scheduler-manifest.mjs",
  "tools/harness/scheduler/scheduler-resource-policy.mjs",
  "tools/harness/scheduler/scheduler-reporting.mjs",
  "tools/harness/scheduler/scheduler-resources.mjs",
  "tools/harness/scheduler/process-executor.mjs",
]);
const schedulerDiagnosticsOwnerFacadePaths = new Set([
  "tools/harness/scheduler/scheduler/event-order.mjs",
  "tools/harness/scheduler/scheduler/summary-timing-drift.mjs",
]);
const testOutputOwnerFacadePaths = new Set([
  "tools/harness/output/test-output/frontend-indexes.mjs",
  "tools/harness/output/test-output/frontend-row-evidence.mjs",
  "tools/harness/output/test-output/playwright-artifacts.mjs",
]);
const browserPrivateImportAllowedSources = new Set([
  "tools/harness/output/test-output/playwright-artifacts.mjs",
  "tools/harness/scheduler/adapters/browser.mjs",
]);
const unsupportedPrivateHelperPaths = new Set([
  "tools/harness/backend/drift/manifests.mjs",
  "tools/harness/backend/duration/baselines.mjs",
  "tools/harness/backend/govulncheck-findings.mjs",
  "tools/harness/backend/migration-history-cli.mjs",
  "tools/harness/backend/migration-history.mjs",
  "tools/harness/backend/runner/go-shards.mjs",
  "tools/harness/backend/schema-object-ownership-cli.mjs",
  "tools/harness/backend/schema-object-ownership.mjs",
  "tools/harness/frontend/accessibility-summary-cli.mjs",
  "tools/harness/frontend/build-web-artifact.sh",
  "tools/harness/frontend/design-token-cli.mjs",
  "tools/harness/frontend/design-tokens.mjs",
  "tools/harness/frontend/design/index.mjs",
  "tools/harness/frontend/embed-web-assets.sh",
  "tools/harness/frontend/evidence/index.mjs",
  "tools/harness/frontend/evidence/test-output-indexes.mjs",
  "tools/harness/frontend/font-bundle-check-cli.mjs",
  "tools/harness/frontend/frontend-evidence-audit-cli.mjs",
  "tools/harness/frontend/frontend-install.sh",
  "tools/harness/frontend/frontend-phase-manifest.mjs",
  "tools/harness/frontend/frontend-phase-slice-plan.mjs",
  "tools/harness/frontend/frontend-row-accounting.mjs",
  "tools/harness/frontend/frontend-toolchain.sh",
  "tools/harness/frontend/readiness/index.mjs",
  "tools/harness/frontend/run-frontend-unit.sh",
  "tools/harness/frontend/run-vitest-manifest-phase.sh",
  "tools/harness/frontend/run-vitest-phase.sh",
  "tools/harness/frontend/vitest-failure-details.mjs",
  "tools/harness/scheduler/adapters/backend.mjs",
  "tools/harness/scheduler/adapters/schedule-context.mjs",
  "tools/harness/scheduler/check-service-backed-expansion.mjs",
  "tools/harness/scheduler/duration-baseline-cli.mjs",
  "tools/harness/scheduler/duration-baseline-drift-suite.sh",
  "tools/harness/scheduler/duration-drift.mjs",
  "tools/harness/scheduler/execution-dependencies.mjs",
  "tools/harness/scheduler/harness-smoke-durations-cli.mjs",
  "tools/harness/scheduler/phase-slice-cli.mjs",
  "tools/harness/scheduler/phase-slice-plan.mjs",
  "tools/harness/scheduler/scheduler/process-executor.mjs",
  "tools/harness/scheduler/scheduler-event-order-drift-cli.mjs",
  "tools/harness/scheduler/scheduler-summary-timing-drift-cli.mjs",
  "tools/harness/scheduler/service-backed-make-target-durations-cli.mjs",
  "tools/harness/scheduler/service-backed-schedule-manifest.mjs",
  "tools/harness/scheduler/service-backed-schedule-topology.mjs",
  "tools/harness/scheduler/target-duration-baselines.mjs",
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

function isUnsupportedPrivateHelperImport(edge) {
  return unsupportedPrivateHelperPaths.has(edge.target);
}

function unsupportedPrivateHelperImportViolation(edge) {
  return {
    rule: "forbidden_unsupported_private_helper_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} imports ${edge.target}; this helper path is unsupported ` +
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
  if (unsupportedPrivateHelperPaths.has(edge.target)) {
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
      `${edge.source} imports ${edge.target}; non-backend harness code must use ` +
      "the backend target, shard, or duration owner facade.",
  };
}

function isPrivateFrontendCatchAllImport(edge) {
  return (
    edge.target.startsWith("tools/harness/frontend/") &&
    !frontendOwnerFacadePaths.has(edge.target) &&
    !unsupportedPrivateHelperPaths.has(edge.target)
  );
}

function privateFrontendCatchAllImportViolation(edge) {
  return {
    rule: "forbidden_private_frontend_catch_all_import",
    source: edge.source,
    target: edge.target,
    message:
      `${edge.source} imports ${edge.target}; frontend harness helpers must use ` +
      "the declared phase-accounting, output, execution, readiness, browser, " +
      "generated-artifact, or static-analysis owner facade.",
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
      `${edge.source} imports ${edge.target}; non-owner harness code must use ` +
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
      `${edge.source} imports ${edge.target}; scheduler code must use ` +
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
    owner_facades: {
      backend: Array.from(backendOwnerFacadePaths).sort(sortStrings),
      browser: Array.from(browserOwnerFacadePaths).sort(sortStrings),
      duration_accounting: Array.from(durationAccountingOwnerFacadePaths).sort(sortStrings),
      frontend: Array.from(frontendOwnerFacadePaths).sort(sortStrings),
      phase_accounting: Array.from(phaseAccountingOwnerFacadePaths).sort(sortStrings),
      scheduler: Array.from(schedulerOwnerFacadePaths).sort(sortStrings),
      scheduler_diagnostics: Array.from(schedulerDiagnosticsOwnerFacadePaths).sort(sortStrings),
      service_backed_execution: Array.from(serviceBackedExecutionOwnerFacadePaths).sort(sortStrings),
      test_output: Array.from(testOutputOwnerFacadePaths).sort(sortStrings),
    },
    unsupported_private_helpers: Array.from(unsupportedPrivateHelperPaths).sort(sortStrings),
    violations: [
      ...edgeViolations,
      ...privateCoreViolations,
      ...unsupportedHelperViolations,
      ...privateBackendViolations,
      ...privateFrontendViolations,
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
