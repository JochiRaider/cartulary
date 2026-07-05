import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import { validSupportTargets } from "../execution/execution-dependencies.mjs";
import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "./frontend-phase-manifest.mjs";
import { validGoSections } from "./phase-manifest-constants.mjs";
import {
  collectEntries,
  collectSupportGoEntries,
  entryIsExecutable,
  supportGoEntrySymbols,
} from "./phase-entry-evidence.mjs";
import { loadManifest, phaseManifestNames } from "./phase-manifest-loader.mjs";

export function packageMatchesPattern(pkg, pattern) {
  if (pattern.endsWith("/...")) {
    const prefix = pattern.slice(0, -4);
    return pkg === prefix || pkg.startsWith(`${prefix}/`);
  }
  return pkg === pattern;
}

function entryMatchesExecutionDependency(entry, executionDependency) {
  return executionDependency === "" || entry.execution_dependency === executionDependency;
}

export function selectManifestEntries(root, {
  phase,
  runner = "",
  section = "",
  coverage = "",
  executionDependency = "",
  executionFamily = "",
  packagePatterns = [],
}) {
  const { manifest } = loadManifest(root, phase);
  return collectEntries(manifest).filter(
    (entry) =>
      entryIsExecutable(entry) &&
      (runner === "" || entry.runner === runner) &&
      (section === "" || entry.section === section) &&
      (coverage === "" || entry.coverage === coverage) &&
      entryMatchesExecutionDependency(entry, executionDependency) &&
      (executionFamily === "" || entry.execution_family === executionFamily) &&
      (packagePatterns.length === 0 ||
        packagePatterns.some((pattern) => packageMatchesPattern(entry.package, pattern))),
  );
}

export function selectGoEntries(
  root,
  phase,
  section,
  coverage,
  executionDependency,
  executionFamily,
  packagePatterns,
) {
  if (!validGoSections.has(section)) {
    throw new Error(`invalid go manifest section ${section}`);
  }
  if (packagePatterns.length === 0) {
    throw new Error("go manifest selection requires at least one package pattern");
  }
  return selectManifestEntries(root, {
    phase,
    runner: "go_test",
    section,
    coverage,
    executionDependency,
    executionFamily,
    packagePatterns,
  });
}

export function selectSupportGoEntries(root, phase, target, executionFamily, packagePatterns) {
  if (!validSupportTargets.has(target)) {
    throw new Error(`invalid support target ${target}`);
  }
  if (packagePatterns.length === 0) {
    throw new Error("support go selection requires at least one package pattern");
  }
  const { manifest } = loadManifest(root, phase);
  return collectSupportGoEntries(manifest).filter(
    (entry) =>
      entry.target === target &&
      (executionFamily === "" || entry.execution_family === executionFamily) &&
      packagePatterns.some((pattern) => packageMatchesPattern(entry.package, pattern)),
  );
}

let cachedGoModulePath;

function loadGoModulePath(root) {
  if (cachedGoModulePath !== undefined) {
    return cachedGoModulePath;
  }
  const goMod = readFileSync(path.join(root, "go.mod"), "utf8");
  const match = goMod.match(/^module\s+(\S+)$/m);
  if (!match) {
    throw new Error("unable to determine Go module path from go.mod");
  }
  cachedGoModulePath = match[1];
  return cachedGoModulePath;
}

export function toGoImportPath(root, repoRelativePackage) {
  if (!repoRelativePackage.startsWith("./")) {
    throw new Error(`manifest Go package must be repo-relative: ${repoRelativePackage}`);
  }
  const suffix = repoRelativePackage.slice(2);
  if (suffix === "") {
    return loadGoModulePath(root);
  }
  return `${loadGoModulePath(root)}/${suffix}`;
}

let cachedPlaywrightSourceFiles = null;

function playwrightSourceFiles(root) {
  if (cachedPlaywrightSourceFiles !== null) {
    return cachedPlaywrightSourceFiles;
  }
  const e2eRoot = path.join(root, "apps", "web", "e2e");
  const files = [];
  const stack = [e2eRoot];
  while (stack.length > 0) {
    const current = stack.pop();
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
        continue;
      }
      if (entry.isFile() && entry.name.endsWith(".spec.ts")) {
        files.push(path.relative(root, next).replaceAll("\\", "/"));
      }
    }
  }
  cachedPlaywrightSourceFiles = files.sort();
  return cachedPlaywrightSourceFiles;
}

function findPlaywrightFileForTitle(root, title) {
  for (const file of playwrightSourceFiles(root)) {
    if (readFileSync(path.join(root, file), "utf8").includes(title)) {
      return file;
    }
  }
  return "";
}

function frontendPhaseToBasePhase(phaseID) {
  const match = /^FE-P([0-9]+)$/u.exec(phaseID);
  return match ? `phase${match[1]}` : "";
}

function frontendTargetForPlaywrightDependency(executionDependency) {
  switch (executionDependency) {
    case "browser_functional":
      return "browser-e2e-webserver-backed";
    case "browser_stateful":
      return "browser-e2e-stateful";
    default:
      return "";
  }
}

function isPlaywrightSupportFile(file) {
  return file.includes(".support.");
}

function selectFrontendPlaywrightEntries(root, phase, coverage, executionDependency) {
  const target = frontendTargetForPlaywrightDependency(executionDependency);
  if (coverage !== "authoritative" || target === "") {
    return [];
  }
  const entries = [];
  let registry;
  try {
    registry = loadFrontendPhaseRegistry(root);
  } catch {
    return entries;
  }
  for (const frontendPhase of registry.phases) {
    const basePhase = frontendPhaseToBasePhase(frontendPhase.phase_id);
    if (basePhase !== phase) {
      continue;
    }
    const { manifest } = loadFrontendPhaseMap(root, frontendPhase.phase_id);
    for (const row of manifest.rows) {
      if (
        row.claim_status !== "implemented" ||
        !row.targets.some((targetRef) => targetRef.target_name === target)
      ) {
        continue;
      }
      for (const title of row.scenario_titles) {
        const file = findPlaywrightFileForTitle(root, title);
        if (file === "") {
          throw new Error(
            `implemented frontend browser row ${row.id} has no Playwright test title: ${title}`,
          );
        }
        if (
          executionDependency === "browser_functional" &&
          isPlaywrightSupportFile(file)
        ) {
          continue;
        }
        entries.push({
          id: row.id,
          phase,
          section: "e2e",
          runner: "playwright",
          coverage: "authoritative",
          execution_dependency: executionDependency,
          file,
          title,
          evidence_class: row.evidence_class,
          layer: row.layer,
          claim_status: row.claim_status,
        });
      }
    }
  }
  return entries.sort((left, right) => {
    if (left.file !== right.file) {
      return left.file.localeCompare(right.file);
    }
    if (left.title !== right.title) {
      return left.title.localeCompare(right.title);
    }
    return left.id.localeCompare(right.id, undefined, { numeric: true });
  });
}

export function selectPlaywrightEntries(root, phase, coverage, executionDependency) {
  return [
    ...selectManifestEntries(root, {
      phase,
      runner: "playwright",
      coverage,
      executionDependency,
    }),
    ...selectFrontendPlaywrightEntries(root, phase, coverage, executionDependency),
  ];
}

function selectBasePlaywrightEntries(root, phase, coverage, executionDependency) {
  return selectManifestEntries(root, {
    phase,
    runner: "playwright",
    coverage,
    executionDependency,
  });
}

export function selectPlaywrightEntriesAll(root, coverage, executionDependency) {
  return phaseManifestNames(root).flatMap((phase) =>
    selectPlaywrightEntries(root, phase, coverage, executionDependency).map((entry) => ({
      ...entry,
      phase,
    })),
  );
}

export function selectPlaywrightPhases(root, coverage, executionDependency) {
  return phaseManifestNames(root).filter(
    (phase) =>
      selectBasePlaywrightEntries(root, phase, coverage, executionDependency)
        .length > 0 ||
      selectFrontendPlaywrightEntries(root, phase, coverage, executionDependency)
        .length > 0,
  );
}

function parsePlaywrightSelectionSpec(spec) {
  const [phase, coverage, executionDependency = ""] = spec.split(":");
  if (!phase || !coverage) {
    throw new Error(
      `invalid playwright selection ${spec}; expected <phase>:<coverage>[:<execution_dependency>]`,
    );
  }
  return { phase, coverage, executionDependency };
}

export function selectPlaywrightEntriesForSpecs(root, specs) {
  if (specs.length === 0) {
    throw new Error("playwright multi-phase selection requires at least one selection spec");
  }
  return specs.flatMap((spec) => {
    const { phase, coverage, executionDependency } = parsePlaywrightSelectionSpec(spec);
    const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
    if (entries.length === 0) {
      throw new Error(`no ${coverage} playwright tests found for ${phase}`);
    }
    return entries;
  });
}

export function normalizePlaywrightFile(file) {
  if (!file.startsWith("apps/web/")) {
    throw new Error(`playwright manifest file must live under apps/web/: ${file}`);
  }
  return file.slice("apps/web/".length);
}

export function selectVitestEntries(root, phase, coverage, executionDependency) {
  return selectManifestEntries(root, {
    phase,
    runner: "vitest",
    coverage,
    executionDependency,
  });
}

export function selectVitestPhases(root, coverage, executionDependency) {
  return phaseManifestNames(root).filter(
    (phase) =>
      selectManifestEntries(root, {
        phase,
        runner: "vitest",
        coverage,
        executionDependency,
      }).length > 0,
  );
}

export function normalizeVitestFile(file) {
  if (!file.startsWith("apps/web/")) {
    throw new Error(`vitest manifest file must live under apps/web/: ${file}`);
  }
  return file.slice("apps/web/".length);
}
