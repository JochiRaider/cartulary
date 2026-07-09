import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  collectEntries,
  entryIsExecutable,
  frontendPhaseBaseJoin,
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
  loadManifest,
  phaseManifestNames,
  playwrightEntryTitles,
} from "../phase-accounting/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const explicitDefaultReasonCodes = new Set([
  "explicit_full_target",
  "explicit_measurement",
  "design_direction_explicit_only",
  "claim_publication_boundary",
]);

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

export function browserDefaultCheckRowIsAdmissible(row) {
  return (
    row.default_check_required === true &&
    row.default_check_kind !== "explicit_only" &&
    !explicitDefaultReasonCodes.has(row.default_check_reason_code) &&
    row.warm_local_cost_class !== "explicit_heavy"
  );
}

function browserTargetsForExecutionDependency(executionDependency) {
  switch (executionDependency) {
    case "browser_functional":
      return ["browser-e2e-functional", "browser-e2e-webserver-backed"];
    case "browser_stateful":
      return ["browser-e2e-stateful"];
    case "browser_measurement":
      return ["browser-e2e-measurement"];
    case "browser_a11y":
      return ["browser-e2e-a11y"];
    case "browser_visual":
      return ["browser-e2e-visual"];
    default:
      return [];
  }
}

export function browserFunctionalEntries(
  root = repoRoot,
  { phase: phaseFilter = "", defaultCheckOnly = false } = {},
) {
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
        if (defaultCheckOnly && !browserDefaultCheckRowIsAdmissible(entry)) {
          continue;
        }
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
          execution_dependency: entry.execution_dependency,
          default_check_required: entry.default_check_required,
          default_check_kind: entry.default_check_kind,
          default_check_reason_code: entry.default_check_reason_code,
          default_check_reason: entry.default_check_reason,
          primary_evidence_owner: entry.primary_evidence_owner,
          duplicate_of: entry.duplicate_of,
          evidence_delta: entry.evidence_delta,
          warm_local_cost_class: entry.warm_local_cost_class,
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

function isPlaywrightSupportFile(file) {
  return file.includes(".support.");
}

export function frontendBrowserReadinessEntries(
  root = repoRoot,
  { baseEntries, phase = "", frontendRowIDs = new Set(), defaultCheckOnly = false } = {},
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
    const basePhase = frontendPhaseBaseJoin(frontendPhase);
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
        (defaultCheckOnly && !browserDefaultCheckRowIsAdmissible(row)) ||
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
          target_names: row.targets.map((target) => target.target_name),
          default_check_required: row.default_check_required,
          default_check_kind: row.default_check_kind,
          default_check_reason_code: row.default_check_reason_code,
          default_check_reason: row.default_check_reason,
          primary_evidence_owner: row.primary_evidence_owner,
          duplicate_of: row.duplicate_of,
          evidence_delta: row.evidence_delta,
          warm_local_cost_class: row.warm_local_cost_class,
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
  { phase = "", frontendRowIDs = new Set(), defaultCheckOnly = false } = {},
) {
  const baseEntries = browserFunctionalEntries(root, { phase, defaultCheckOnly });
  const allBaseEntries =
    phase || frontendRowIDs.size > 0
      ? browserFunctionalEntries(root, { defaultCheckOnly })
      : baseEntries;
  const frontendEntries = frontendBrowserReadinessEntries(root, {
    baseEntries: allBaseEntries,
    phase,
    frontendRowIDs,
    defaultCheckOnly,
  });
  return [...baseEntries, ...frontendEntries].sort(compareEntries);
}

export function selectedEntriesForPlan(
  root = repoRoot,
  { phase = "", frontendRowIDs = new Set(), defaultCheckOnly = false } = {},
) {
  const baseEntries = browserFunctionalEntries(root, { defaultCheckOnly });
  const frontendEntries = frontendBrowserReadinessEntries(root, {
    baseEntries: frontendRowIDs.size > 0 ? [] : baseEntries,
    phase,
    frontendRowIDs,
    defaultCheckOnly,
  });
  return [
    ...browserFunctionalEntries(root, { phase, defaultCheckOnly }),
    ...frontendEntries.filter((entry) => !phase || entry.phase === phase),
  ].sort(compareEntries);
}

function addBrowserRowRecord(records, id, record) {
  if (records.has(id)) {
    throw new Error(`duplicate browser row id ${id}`);
  }
  records.set(id, record);
}

export function browserDefaultCheckRowIndex(root = repoRoot) {
  const records = new Map();
  for (const phase of phaseManifestNames(root)) {
    const { manifest } = loadManifest(root, phase);
    for (const entry of collectEntries(manifest)) {
      if (
        entry.runner !== "playwright" ||
        !String(entry.execution_dependency ?? "").startsWith("browser_")
      ) {
        continue;
      }
      addBrowserRowRecord(records, entry.id, {
        id: entry.id,
        source_family: "browser",
        phase,
        execution_dependency: entry.execution_dependency,
        targets: browserTargetsForExecutionDependency(entry.execution_dependency),
        implemented: entryIsExecutable(entry),
        admissible:
          entryIsExecutable(entry) && browserDefaultCheckRowIsAdmissible(entry),
        default_check_required: entry.default_check_required,
        default_check_kind: entry.default_check_kind,
        default_check_reason_code: entry.default_check_reason_code,
        warm_local_cost_class: entry.warm_local_cost_class,
      });
    }
  }

  const registry = loadFrontendPhaseRegistry(root);
  for (const frontendPhase of registry.phases) {
    const basePhase = frontendPhaseBaseJoin(frontendPhase);
    if (basePhase === "") {
      continue;
    }
    const { manifest } = loadFrontendPhaseMap(root, frontendPhase.phase_id);
    for (const row of manifest.rows) {
      const targets = row.targets
        .map((target) => target.target_name)
        .filter((target) => target.startsWith("browser-e2e"));
      if (targets.length === 0) {
        continue;
      }
      addBrowserRowRecord(records, row.id, {
        id: row.id,
        source_family: "frontend",
        phase: basePhase,
        frontend_phase: frontendPhase.phase_id,
        targets,
        implemented: row.claim_status === "implemented",
        admissible:
          row.claim_status === "implemented" &&
          browserDefaultCheckRowIsAdmissible(row),
        default_check_required: row.default_check_required,
        default_check_kind: row.default_check_kind,
        default_check_reason_code: row.default_check_reason_code,
        warm_local_cost_class: row.warm_local_cost_class,
      });
    }
  }
  return records;
}
