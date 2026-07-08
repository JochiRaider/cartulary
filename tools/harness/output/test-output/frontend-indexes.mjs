import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "../../phase-accounting/index.mjs";

const vitestIndexCache = new Map();
const playwrightIndexCache = new Map();

function cacheKey(root) {
  return root;
}

function frontendEvidenceCoverage(evidenceClass) {
  return evidenceClass === "product_conformance" ? "authoritative" : "support";
}

export function loadFrontendVitestIndex(root) {
  const key = cacheKey(root);
  if (vitestIndexCache.has(key)) {
    return vitestIndexCache.get(key);
  }
  const byTitle = new Map();
  let registry;
  try {
    registry = loadFrontendPhaseRegistry(root);
  } catch {
    const empty = { byTitle };
    vitestIndexCache.set(key, empty);
    return empty;
  }
  for (const phase of registry.phases) {
    const { manifest } = loadFrontendPhaseMap(root, phase.phase_id);
    for (const row of manifest.rows) {
      if (
        !row.targets.some((target) => target.target_name === "frontend-unit") ||
        row.scenario_titles.length === 0
      ) {
        continue;
      }
      for (const title of row.scenario_titles) {
        byTitle.set(title, {
          coverage: frontendEvidenceCoverage(row.evidence_class),
          phase: phase.phase_id,
          id: row.id,
          evidence_class: row.evidence_class,
        });
      }
    }
  }
  const index = { byTitle };
  vitestIndexCache.set(key, index);
  return index;
}

export function loadFrontendPlaywrightIndex(root) {
  const key = cacheKey(root);
  if (playwrightIndexCache.has(key)) {
    return playwrightIndexCache.get(key);
  }
  const byTitle = new Map();
  let registry;
  try {
    registry = loadFrontendPhaseRegistry(root);
  } catch {
    const empty = { byTitle };
    playwrightIndexCache.set(key, empty);
    return empty;
  }
  for (const phase of registry.phases) {
    const { manifest } = loadFrontendPhaseMap(root, phase.phase_id);
    for (const row of manifest.rows) {
      if (
        !row.targets.some((target) =>
          target.target_name.startsWith("browser-e2e"),
        ) ||
        row.scenario_titles.length === 0
      ) {
        continue;
      }
      for (const title of row.scenario_titles) {
        byTitle.set(title, {
          coverage: frontendEvidenceCoverage(row.evidence_class),
          phase: phase.phase_id,
          id: row.id,
          evidence_class: row.evidence_class,
        });
      }
    }
  }
  const index = { byTitle };
  playwrightIndexCache.set(key, index);
  return index;
}
