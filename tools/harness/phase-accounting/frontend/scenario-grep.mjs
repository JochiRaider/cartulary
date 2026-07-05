import {
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "./registry.mjs";

function targetDisplayName(target) {
  return target.target_name ? `make ${target.target_name}` : String(target);
}

function targetRefMatches(target, normalizedTarget) {
  return targetDisplayName(target) === normalizedTarget;
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

export function frontendScenarioTitlesForTarget(
  root,
  target,
  layer = "",
  options = {},
) {
  const normalizedTarget = target.startsWith("make ") ? target : `make ${target}`;
  const registry = loadFrontendPhaseRegistry(root);
  const selectedRowIDs =
    options.rowIDs === undefined || options.rowIDs === null
      ? null
      : new Set(
          (Array.isArray(options.rowIDs)
            ? options.rowIDs
            : String(options.rowIDs).split(",")
          )
            .map((rowID) => String(rowID).trim())
            .filter(Boolean),
        );
  const titles = [];
  for (const phase of registry.phases) {
    const { manifest } = loadFrontendPhaseMap(root, phase.phase_id);
    for (const row of manifest.rows) {
      if (selectedRowIDs && !selectedRowIDs.has(row.id)) {
        continue;
      }
      if (layer && row.layer !== layer) {
        continue;
      }
      if (
        !row.targets.some((targetRef) =>
          targetRefMatches(targetRef, normalizedTarget),
        )
      ) {
        continue;
      }
      titles.push(...row.scenario_titles);
    }
  }
  return [...new Set(titles)].sort();
}

export function frontendPlaywrightGrepForTarget(
  root,
  target,
  layer = "",
  options = {},
) {
  const titles = frontendScenarioTitlesForTarget(root, target, layer, options);
  if (titles.length === 0) {
    return "";
  }
  return `(?:${titles.map(escapeRegExp).join("|")})`;
}

export function frontendExactTitleGrepForTarget(
  root,
  target,
  layer = "",
  options = {},
) {
  const titles = frontendScenarioTitlesForTarget(root, target, layer, options);
  if (titles.length === 0) {
    return "";
  }
  if (titles.length === 1) {
    return `${escapeRegExp(titles[0])}$`;
  }
  return `(?:${titles.map(escapeRegExp).join("|")})$`;
}

