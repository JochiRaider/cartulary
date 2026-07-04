import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  checkBaselineDriftFromEntries,
  createPlanFromEntries,
  updateBaselinesFromEntries,
} from "./browser-shard-plan.mjs";
import {
  collectEntries,
  entryIsExecutable,
  loadManifest,
  phaseManifestNames,
  playwrightEntryTitles,
} from "../phase-accounting/phase-manifest.mjs";
import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "../frontend/evidence/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");

function compareEntries(left, right) {
  if (left.phase !== right.phase) {
    return left.phase.localeCompare(right.phase, undefined, { numeric: true });
  }
  if (left.file !== right.file) {
    return left.file.localeCompare(right.file);
  }
  if (left.title !== right.title) {
    return left.title.localeCompare(right.title);
  }
  return left.id.localeCompare(right.id, undefined, { numeric: true });
}

function normalizeManifestFile(file) {
  const normalized = String(file ?? "").replaceAll("\\", "/");
  if (normalized.startsWith("apps/web/e2e/")) {
    return normalized;
  }
  if (normalized.startsWith("e2e/")) {
    return `apps/web/${normalized}`;
  }
  return normalized;
}

export function browserFunctionalEntries(root = repoRoot, { phase: phaseFilter = "" } = {}) {
  const entries = [];
  const seenIDs = new Set();
  for (const phase of phaseManifestNames(root)) {
    if (phaseFilter && phase !== phaseFilter) {
      continue;
    }
    const { manifest } = loadManifest(root, phase);
    for (const entry of collectEntries(manifest)) {
      if (
        entry.section === "e2e" &&
        entry.runner === "playwright" &&
        entry.coverage === "authoritative" &&
        entry.execution_dependency === "browser_functional" &&
        entryIsExecutable(entry)
      ) {
        if (seenIDs.has(entry.id)) {
          throw new Error(`duplicate browser functional manifest ID ${entry.id}`);
        }
        seenIDs.add(entry.id);
        const titles = playwrightEntryTitles(entry);
        entries.push({
          id: entry.id,
          phase,
          file: normalizeManifestFile(entry.file),
          title: titles[0],
          titles,
        });
      }
    }
  }
  return entries.sort(compareEntries);
}

const playwrightSourceFileCache = new Map();

function playwrightSourceFiles(root) {
  const cacheKey = path.resolve(root);
  if (playwrightSourceFileCache.has(cacheKey)) {
    return playwrightSourceFileCache.get(cacheKey);
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
  const sortedFiles = files.sort();
  playwrightSourceFileCache.set(cacheKey, sortedFiles);
  return sortedFiles;
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

function isPlaywrightSupportFile(file) {
  return file.includes(".support.");
}

export function frontendBrowserReadinessEntries(
  root = repoRoot,
  { baseEntries, phase = "", frontendRowIDs = new Set() } = {},
) {
  if (process.env.CARTULARY_PHASE_MANIFEST_ROOT && frontendRowIDs.size === 0) {
    return [];
  }
  const baseTitles = new Set(
    baseEntries.flatMap((entry) => entry.titles ?? [entry.title]),
  );
  const seenTitles = new Set();
  const knownSelectedIDs = new Set();
  const entries = [];
  const activeBasePhases = new Set(phaseManifestNames(root));
  const registry = loadFrontendPhaseRegistry(root);
  for (const frontendPhase of registry.phases) {
    const basePhase = frontendPhaseToBasePhase(frontendPhase.phase_id);
    if (basePhase === "" || (phase && basePhase !== phase)) {
      continue;
    }
    if (frontendRowIDs.size === 0 && !activeBasePhases.has(basePhase)) {
      continue;
    }
    const { manifest } = loadFrontendPhaseMap(root, frontendPhase.phase_id);
    for (const row of manifest.rows) {
      if (frontendRowIDs.size > 0) {
        if (!frontendRowIDs.has(row.id)) {
          continue;
        }
        knownSelectedIDs.add(row.id);
      }
      if (
        row.claim_status !== "implemented" ||
        !row.targets.some(
          (target) => target.target_name === "browser-e2e-webserver-backed",
        )
      ) {
        continue;
      }
      for (const title of row.scenario_titles) {
        if (baseTitles.has(title) || seenTitles.has(title)) {
          continue;
        }
        const file = findPlaywrightFileForTitle(root, title);
        if (file === "") {
          throw new Error(
            `implemented frontend browser row ${row.id} has no Playwright test title: ${title}`,
          );
        }
        if (isPlaywrightSupportFile(file)) {
          continue;
        }
        seenTitles.add(title);
        entries.push({
          id: row.id,
          phase: basePhase,
          file,
          title,
          titles: [title],
          frontend_phase: frontendPhase.phase_id,
        });
      }
    }
  }
  if (frontendRowIDs.size > 0) {
    const unknown = [...frontendRowIDs]
      .filter((rowID) => !knownSelectedIDs.has(rowID))
      .sort((left, right) => left.localeCompare(right));
    if (unknown.length > 0) {
      throw new Error(
        `selected frontend browser row id(s) not found: ${unknown.join(",")}`,
      );
    }
  }
  return entries.sort(compareEntries);
}

export function browserDurationBaselineEntries(
  root = repoRoot,
  { phase = "", frontendRowIDs = new Set() } = {},
) {
  const baseEntries = browserFunctionalEntries(root, { phase });
  const allBaseEntries =
    phase || frontendRowIDs.size > 0 ? browserFunctionalEntries(root) : baseEntries;
  const frontendEntries = frontendBrowserReadinessEntries(root, {
    baseEntries: allBaseEntries,
    phase,
    frontendRowIDs,
  });
  return [...baseEntries, ...frontendEntries].sort(compareEntries);
}

function selectedEntriesForPlan(root, { phase = "", frontendRowIDs = new Set() } = {}) {
  const baseEntries = browserFunctionalEntries(root);
  const frontendEntries = frontendBrowserReadinessEntries(root, {
    baseEntries: frontendRowIDs.size > 0 ? [] : baseEntries,
    phase,
    frontendRowIDs,
  });
  return [
    ...browserFunctionalEntries(root, { phase }),
    ...frontendEntries.filter((entry) => !phase || entry.phase === phase),
  ].sort(compareEntries);
}

export function createPlan(options) {
  return createPlanFromEntries({
    ...options,
    baselineEntries: browserDurationBaselineEntries(repoRoot),
    selectedEntries: selectedEntriesForPlan(repoRoot, options),
  });
}

export function updateBaselines(argv) {
  updateBaselinesFromEntries(argv, browserDurationBaselineEntries(repoRoot));
}

export function checkBaselineDrift(argv) {
  checkBaselineDriftFromEntries(argv, browserDurationBaselineEntries(repoRoot));
}
