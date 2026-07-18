import { loadTestCatalog } from "../../test-catalog/index.mjs";

const vitestIndexCache = new Map();
const playwrightIndexCache = new Map();

function indexRows(root, runner) {
  const byTitle = new Map();
  for (const row of loadTestCatalog(root).rows) {
    if (row.status !== "active" || row.runner !== runner) {
      continue;
    }
    for (const title of row.selector.titles) {
      byTitle.set(title, {
        coverage: "authoritative",
        owner_id: row.owner_id,
        id: row.row_id,
        evidence_class: row.evidence_class,
      });
    }
  }
  return { byTitle };
}

export function loadFrontendVitestIndex(root) {
  if (!vitestIndexCache.has(root)) {
    vitestIndexCache.set(root, indexRows(root, "vitest"));
  }
  return vitestIndexCache.get(root);
}

export function loadFrontendPlaywrightIndex(root) {
  if (!playwrightIndexCache.has(root)) {
    playwrightIndexCache.set(root, indexRows(root, "playwright"));
  }
  return playwrightIndexCache.get(root);
}
