import {
  collectEntries,
  loadManifest,
  loadSubsystemManifests,
  packageMatchesPattern,
  phaseManifestNames,
  subsystemManifestOwner,
  playwrightEntryTitles,
  selectManifestEntries,
  selectPlaywrightEntries,
  vitestEntryTitles,
} from "../../phase-accounting/index.mjs";

const manifestIndexCache = new Map();

function cacheKey(root) {
  return root;
}

export {
  packageMatchesPattern,
  phaseManifestNames,
  playwrightEntryTitles,
  selectManifestEntries,
  selectPlaywrightEntries,
  vitestEntryTitles,
};

export function loadManifestIndex(root, { normalizePath, toGoImportPath }) {
  const key = cacheKey(root);
  if (manifestIndexCache.has(key)) {
    return manifestIndexCache.get(key);
  }

  const index = {
    authoritativeGo: new Map(),
    authoritativeVitest: new Map(),
    authoritativePlaywright: new Map(),
    manifestVitest: new Map(),
    manifestPlaywright: new Map(),
    forbiddenFilesByPhase: new Map(),
  };

  for (const phase of phaseManifestNames(root)) {
    const { manifest } = loadManifest(root, phase);
    const entries = collectEntries(manifest);
    for (const forbidden of manifest.forbidden_id_files ?? []) {
      if (!index.forbiddenFilesByPhase.has(phase)) {
        index.forbiddenFilesByPhase.set(phase, new Set());
      }
      index.forbiddenFilesByPhase.get(phase).add(normalizePath(forbidden));
    }
    for (const entry of entries) {
      if (entry.runner === "vitest") {
        for (const title of vitestEntryTitles(entry)) {
          index.manifestVitest.set(`${normalizePath(entry.file)}::${title}`, {
            ...entry,
            phase,
            title,
          });
        }
      }
      if (entry.runner === "playwright") {
        for (const title of playwrightEntryTitles(entry)) {
          index.manifestPlaywright.set(
            `${normalizePath(entry.file)}::${title}`,
            { ...entry, phase, title },
          );
        }
      }
      if (entry.coverage !== "authoritative") {
        continue;
      }
      if (entry.runner === "go_test") {
        const symbols =
          entry.symbol !== undefined ? [entry.symbol] : entry.symbols;
        for (const symbol of symbols) {
          index.authoritativeGo.set(
            `${toGoImportPath(entry.package)}::${symbol}`,
            { ...entry, phase },
          );
        }
        continue;
      }
      if (entry.runner === "vitest") {
        for (const title of vitestEntryTitles(entry)) {
          index.authoritativeVitest.set(
            `${normalizePath(entry.file)}::${title}`,
            { ...entry, phase, title },
          );
        }
        continue;
      }
      if (entry.runner === "playwright") {
        for (const title of playwrightEntryTitles(entry)) {
          index.authoritativePlaywright.set(
            `${normalizePath(entry.file)}::${title}`,
            { ...entry, phase, title },
          );
        }
      }
    }
  }
  for (const { owner, manifest } of loadSubsystemManifests(root)) {
    const phase = subsystemManifestOwner(owner);
    for (const entry of collectEntries(manifest)) {
      if (entry.coverage !== "authoritative" || entry.runner !== "go_test") {
        continue;
      }
      const symbols = entry.symbol !== undefined ? [entry.symbol] : entry.symbols;
      for (const symbol of symbols) {
        index.authoritativeGo.set(
          `${toGoImportPath(entry.package)}::${symbol}`,
          { ...entry, phase },
        );
      }
    }
  }

  manifestIndexCache.set(key, index);
  return index;
}

export function selectGoManifestEntries(
  root,
  {
    phase,
    section,
    coverage,
    executionDependency,
    executionFamily,
    packagePatterns,
  },
) {
  if (phase.startsWith("subsystem:")) {
    const owner = phase.slice("subsystem:".length);
    const descriptor = loadSubsystemManifests(root).find((entry) => entry.owner === owner);
    if (!descriptor) {
      throw new Error(`unknown subsystem manifest owner ${owner}`);
    }
    return collectEntries(descriptor.manifest).filter(
      (entry) =>
        entry.runner === "go_test" &&
        (section === "" || entry.section === section) &&
        (coverage === "" || entry.coverage === coverage) &&
        (executionDependency === "" || entry.execution_dependency === executionDependency) &&
        (executionFamily === "" || entry.execution_family === executionFamily) &&
        (packagePatterns.length === 0 || packagePatterns.some((pattern) => packageMatchesPattern(entry.package, pattern))),
    );
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

export function selectVitestManifestEntries(
  root,
  { phase, coverage, executionDependency },
) {
  return selectManifestEntries(root, {
    runner: "vitest",
    phase,
    coverage,
    executionDependency,
  });
}

export function collectVitestManifestEntries(
  root,
  { coverage, executionDependency },
) {
  return phaseManifestNames(root).flatMap((phase) =>
    selectVitestManifestEntries(root, {
      phase,
      coverage,
      executionDependency,
    }),
  );
}

export function selectPlaywrightManifestEntries(
  root,
  { phase, coverage, executionDependency },
) {
  return selectPlaywrightEntries(root, phase, coverage, executionDependency);
}
