import { sectionDefinitions } from "./phase-manifest-constants.mjs";

const goScenarioIDPattern = /^SCN-[0-9]{3}(?:-[A-Z0-9]+)*$/;

export function goEntrySymbols(entry) {
  if (entry.symbol !== undefined && entry.symbols !== undefined) {
    throw new Error(`manifest entry ${entry.id} must declare symbol or symbols[], not both`);
  }
  if (entry.symbols !== undefined) {
    if (!Array.isArray(entry.symbols) || entry.symbols.length === 0) {
      throw new Error(`manifest entry ${entry.id} must declare a non-empty symbols[] array`);
    }
    for (const symbol of entry.symbols) {
      if (typeof symbol !== "string" || symbol.trim() === "") {
        throw new Error(`manifest entry ${entry.id} has an invalid symbol in symbols[]`);
      }
    }
    return entry.symbols;
  }
  if (typeof entry.symbol !== "string" || entry.symbol.trim() === "") {
    throw new Error(`manifest entry ${entry.id} is missing a non-empty symbol`);
  }
  return [entry.symbol];
}

export function goEntryScenarioSymbols(entry) {
  if (entry.scenario_symbols === undefined) {
    return {};
  }
  const label = `manifest entry ${entry.id}`;
  if (
    !entry.scenario_symbols ||
    typeof entry.scenario_symbols !== "object" ||
    Array.isArray(entry.scenario_symbols)
  ) {
    throw new Error(`${label} scenario_symbols must be an object mapping scenario IDs to Go symbols`);
  }
  const symbols = goEntrySymbols(entry);
  const allowedSymbols = new Set(symbols);
  const seenSymbols = new Set();
  const result = {};
  for (const [scenarioID, symbol] of Object.entries(entry.scenario_symbols)) {
    if (!goScenarioIDPattern.test(scenarioID)) {
      throw new Error(`${label} scenario_symbols contains invalid scenario id ${scenarioID}`);
    }
    if (typeof symbol !== "string" || symbol.trim() === "") {
      throw new Error(`${label} scenario_symbols.${scenarioID} must be a non-empty Go symbol`);
    }
    if (!allowedSymbols.has(symbol)) {
      throw new Error(`${label} scenario_symbols.${scenarioID} references undeclared symbol ${symbol}`);
    }
    if (seenSymbols.has(symbol)) {
      throw new Error(`${label} scenario_symbols contains duplicate symbol ${symbol}`);
    }
    seenSymbols.add(symbol);
    result[scenarioID] = symbol;
  }
  if (seenSymbols.size !== symbols.length) {
    throw new Error(`${label} scenario_symbols must cover every declared Go symbol`);
  }
  return result;
}

export function vitestEntryTitles(entry) {
  if (entry.title !== undefined && entry.titles !== undefined) {
    throw new Error(`manifest entry ${entry.id} must declare title or titles[], not both`);
  }
  if (entry.titles !== undefined) {
    if (!Array.isArray(entry.titles) || entry.titles.length === 0) {
      throw new Error(`manifest entry ${entry.id} must declare a non-empty titles[] array`);
    }
    for (const title of entry.titles) {
      if (typeof title !== "string" || title.trim() === "") {
        throw new Error(`manifest entry ${entry.id} has an invalid title in titles[]`);
      }
    }
    return entry.titles;
  }
  if (typeof entry.title !== "string" || entry.title.trim() === "") {
    throw new Error(`manifest entry ${entry.id} is missing a non-empty title`);
  }
  return [entry.title];
}

export function playwrightEntryTitles(entry) {
  if (entry.title !== undefined && entry.titles !== undefined) {
    throw new Error(`manifest entry ${entry.id} must declare title or titles[], not both`);
  }
  if (entry.titles !== undefined) {
    if (!Array.isArray(entry.titles) || entry.titles.length === 0) {
      throw new Error(`manifest entry ${entry.id} must declare a non-empty titles[] array`);
    }
    for (const title of entry.titles) {
      if (typeof title !== "string" || title.trim() === "") {
        throw new Error(`manifest entry ${entry.id} has an invalid title in titles[]`);
      }
    }
    return entry.titles;
  }
  if (typeof entry.title !== "string" || entry.title.trim() === "") {
    throw new Error(`manifest entry ${entry.id} is missing a non-empty title`);
  }
  return [entry.title];
}

export function rowIDFragments(id) {
  return [id, id.replaceAll("-", "_")];
}

export function entryEvidenceNames(entry) {
  if (entry.runner === "go_test") {
    return goEntrySymbols(entry);
  }
  if (entry.runner === "vitest") {
    return vitestEntryTitles(entry);
  }
  if (entry.runner === "playwright") {
    return playwrightEntryTitles(entry);
  }
  return [];
}

export function collectEntries(manifest) {
  const entries = [];
  for (const [section] of sectionDefinitions) {
    for (const entry of manifest[section] ?? []) {
      entries.push({ ...entry, section });
    }
  }
  return entries;
}

export function collectSupportGoEntries(manifest) {
  return (manifest.support_go_targets ?? []).map((entry) => ({ ...entry }));
}

export function authoritativeEvidenceNameViolations(manifest, { phase = manifest.phase ?? "" } = {}) {
  const invalid = [];

  for (const entry of collectEntries(manifest)) {
    if (entry.coverage !== "authoritative") {
      continue;
    }
    const fragments = rowIDFragments(entry.id);
    for (const name of entryEvidenceNames(entry)) {
      if (!fragments.some((fragment) => name.includes(fragment))) {
        invalid.push({
          file: entry.file,
          phase,
          symbol: name,
          reason: `authoritative evidence for ${entry.id} must include ${fragments.join(" or ")}`,
        });
      }
    }
  }

  return invalid;
}

export function assertAuthoritativeEvidenceNames(manifest, options = {}) {
  const invalid = authoritativeEvidenceNameViolations(manifest, options);
  if (invalid.length === 0) {
    return;
  }
  throw new Error(
    `authoritative phase evidence names must include manifest-owned row IDs: ${invalid
      .map((entry) => `${entry.file}::${entry.symbol} (${entry.reason})`)
      .join("; ")}`,
  );
}

export function entryClaimStatus(entry) {
  return entry.claim_status ?? "implemented";
}

export function entryIsExecutable(entry) {
  return entryClaimStatus(entry) !== "blocked";
}

export function supportGoEntryLabel(entry) {
  return `support_go_target ${entry.target ?? "(missing target)"} ${entry.file ?? "(missing file)"}`;
}

export function supportGoEntrySymbols(entry) {
  const label = supportGoEntryLabel(entry);
  if (entry.symbol !== undefined && entry.symbols !== undefined) {
    throw new Error(`${label} must declare symbol or symbols[], not both`);
  }
  if (entry.symbols !== undefined) {
    if (!Array.isArray(entry.symbols) || entry.symbols.length === 0) {
      throw new Error(`${label} must declare a non-empty symbols[] array`);
    }
    for (const symbol of entry.symbols) {
      if (typeof symbol !== "string" || symbol.trim() === "") {
        throw new Error(`${label} has an invalid symbol in symbols[]`);
      }
    }
    return entry.symbols;
  }
  if (typeof entry.symbol !== "string" || entry.symbol.trim() === "") {
    throw new Error(`${label} is missing a non-empty symbol`);
  }
  return [entry.symbol];
}
