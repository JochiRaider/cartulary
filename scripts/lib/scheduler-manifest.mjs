import { readFile } from "node:fs/promises";
import path from "node:path";

export function parseResourceLimitOverride(value) {
  const [resource, amountText, extra] = value.split("=");
  if (!resource || !amountText || extra !== undefined) {
    throw new Error(`--resource-limit expects <name=value>, got ${value}`);
  }
  const amount = Number.parseInt(amountText, 10);
  if (!Number.isInteger(amount) || amount < 1) {
    throw new Error(`--resource-limit ${resource} must be a positive integer`);
  }
  return [resource.trim(), amount];
}

export async function loadScheduleManifest(file, { repoRoot, schemaID }) {
  const manifestPath = path.isAbsolute(file) ? file : path.join(repoRoot, file);
  const manifest = JSON.parse(await readFile(manifestPath, "utf8"));
  if (manifest.schema_id !== schemaID) {
    throw new Error(`${manifestPath} must declare schema_id ${schemaID}`);
  }
  if (!Array.isArray(manifest.schedules)) {
    throw new Error(`${manifestPath} must declare schedules[]`);
  }
  return { manifest, manifestPath };
}

export function selectSingleSchedule(manifest, target, { label = "schedule" } = {}) {
  const matches = manifest.schedules.filter((schedule) => schedule?.target === target);
  if (matches.length !== 1) {
    throw new Error(`expected exactly one ${label} for ${target}, found ${matches.length}`);
  }
  return matches[0];
}

export function normalizeStringList(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label} entries must be non-empty strings`);
    }
    return entry.trim();
  });
}

export function normalizeNeeds(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} needs must be an array`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label} needs entries must be non-empty strings`);
    }
    return entry.trim();
  });
}

export function validateTargetDependencyGraph(
  nodes,
  { scheduleLabel, nodeKind, duplicateTargetsMessage = null },
) {
  const targets = new Set();
  const duplicateTargets = [];
  for (const node of nodes) {
    if (targets.has(node.target)) {
      duplicateTargets.push(node.target);
    }
    targets.add(node.target);
  }
  if (duplicateTargets.length > 0) {
    if (duplicateTargetsMessage) {
      throw new Error(duplicateTargetsMessage(duplicateTargets));
    }
    throw new Error(`${scheduleLabel} contains duplicate ${nodeKind} target ${duplicateTargets[0]}`);
  }

  for (const node of nodes) {
    for (const need of node.needs) {
      if (!targets.has(need)) {
        throw new Error(`${scheduleLabel} ${nodeKind} ${node.target} depends on unknown target ${need}`);
      }
      if (need === node.target) {
        throw new Error(`${scheduleLabel} ${nodeKind} ${node.target} cannot depend on itself`);
      }
    }
  }

  const byTarget = new Map(nodes.map((node) => [node.target, node]));
  const visiting = new Set();
  const visited = new Set();
  const visit = (node) => {
    if (visited.has(node.target)) {
      return;
    }
    if (visiting.has(node.target)) {
      throw new Error(`${scheduleLabel} has a dependency cycle at ${node.target}`);
    }
    visiting.add(node.target);
    for (const need of node.needs) {
      visit(byTarget.get(need));
    }
    visiting.delete(node.target);
    visited.add(node.target);
  };
  for (const node of nodes) {
    visit(node);
  }
}
