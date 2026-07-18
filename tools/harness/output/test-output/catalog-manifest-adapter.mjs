import { loadTestCatalog, targetForCatalogRow } from "../../test-catalog/index.mjs";

const manifestIndexCache = new Map();

function catalogEntry(row) {
  const selector = row.selector;
  const target = targetForCatalogRow(row);
  return {
    id: row.row_id,
    row_id: row.row_id,
    owner_id: row.owner_id,
    family_id: row.family_id,
    runner:
      row.runner === "go"
        ? "go_test"
        : row.runner,
    package: selector.package,
    file: selector.file,
    symbol: selector.tests?.length === 1 ? selector.tests[0] : undefined,
    symbols: selector.tests?.length > 1 ? [...selector.tests] : undefined,
    title: selector.titles?.length === 1 ? selector.titles[0] : undefined,
    titles: selector.titles ? [...selector.titles] : undefined,
    coverage: "authoritative",
    step: row.owner_id,
    section: target === "backend-unit" ? "unit" : "integration",
    execution_dependency: target.replaceAll("-", "_"),
    execution_family: row.family_id,
    evidence_class: row.evidence_class,
  };
}

function entriesFor(root) {
  return loadTestCatalog(root).rows
    .filter(
      (row) =>
        row.status === "active" &&
        ["go", "vitest", "playwright"].includes(row.runner),
    )
    .map(catalogEntry);
}

function normalizePackagePattern(pattern) {
  return String(pattern).replace(/^\.\//u, "").replace(/\/\.\.\.$/u, "/...");
}

export function packageMatchesPattern(packageName, pattern) {
  const normalizedPackage = String(packageName).replace(/^\.\//u, "");
  const normalizedPattern = normalizePackagePattern(pattern);
  if (normalizedPattern.endsWith("/...")) {
    const prefix = normalizedPattern.slice(0, -4);
    return normalizedPackage === prefix || normalizedPackage.startsWith(`${prefix}/`);
  }
  return normalizedPackage === normalizedPattern;
}

export function playwrightEntryTitles(entry) {
  return entry.titles ?? (entry.title ? [entry.title] : []);
}

export function vitestEntryTitles(entry) {
  return entry.titles ?? (entry.title ? [entry.title] : []);
}

export function loadManifestIndex(root, { normalizePath, toGoImportPath }) {
  if (manifestIndexCache.has(root)) {
    return manifestIndexCache.get(root);
  }
  const index = {
    authoritativeGo: new Map(),
    authoritativeVitest: new Map(),
    authoritativePlaywright: new Map(),
    manifestVitest: new Map(),
    manifestPlaywright: new Map(),
    forbiddenFilesByStep: new Map(),
  };
  for (const entry of entriesFor(root)) {
    if (entry.runner === "go_test") {
      for (const symbol of entry.symbols ?? [entry.symbol]) {
        index.authoritativeGo.set(
          `${toGoImportPath(entry.package)}::${symbol}`,
          entry,
        );
      }
      continue;
    }
    if (entry.runner === "vitest") {
      for (const title of vitestEntryTitles(entry)) {
        const keyed = { ...entry, title };
        const key = `${normalizePath(entry.file)}::${title}`;
        index.manifestVitest.set(key, keyed);
        index.authoritativeVitest.set(key, keyed);
      }
      continue;
    }
    if (entry.runner === "playwright") {
      for (const title of playwrightEntryTitles(entry)) {
        const keyed = { ...entry, title };
        const key = `${normalizePath(entry.file)}::${title}`;
        index.manifestPlaywright.set(key, keyed);
        index.authoritativePlaywright.set(key, keyed);
      }
    }
  }
  manifestIndexCache.set(root, index);
  return index;
}

function matchesCommon(entry, options) {
  return (
    (!options.step || entry.step === options.step) &&
    (!options.coverage || entry.coverage === options.coverage) &&
    (!options.executionDependency ||
      entry.execution_dependency === options.executionDependency)
  );
}

export function selectManifestEntries(
  root,
  { runner, step = "", coverage = "", executionDependency = "" },
) {
  return entriesFor(root).filter(
    (entry) =>
      entry.runner === runner &&
      matchesCommon(entry, { step, coverage, executionDependency }),
  );
}

export function selectPlaywrightEntries(
  root,
  step = "",
  coverage = "",
  executionDependency = "",
) {
  return selectManifestEntries(root, {
    runner: "playwright",
    step,
    coverage,
    executionDependency,
  });
}

export function selectGoManifestEntries(
  root,
  {
    step = "",
    section = "",
    coverage = "",
    executionDependency = "",
    executionFamily = "",
    packagePatterns = [],
  },
) {
  return entriesFor(root).filter(
    (entry) =>
      entry.runner === "go_test" &&
      matchesCommon(entry, { step, coverage, executionDependency }) &&
      (!section || entry.section === section) &&
      (!executionFamily || entry.execution_family === executionFamily) &&
      (packagePatterns.length === 0 ||
        packagePatterns.some((pattern) =>
          packageMatchesPattern(entry.package, pattern),
        )),
  );
}

export function selectVitestManifestEntries(
  root,
  { step = "", coverage = "", executionDependency = "" },
) {
  return selectManifestEntries(root, {
    runner: "vitest",
    step,
    coverage,
    executionDependency,
  });
}

export function collectVitestManifestEntries(
  root,
  { coverage = "", executionDependency = "" },
) {
  return selectVitestManifestEntries(root, { coverage, executionDependency });
}

export function selectPlaywrightManifestEntries(
  root,
  { step = "", coverage = "", executionDependency = "" },
) {
  return selectPlaywrightEntries(root, step, coverage, executionDependency);
}
