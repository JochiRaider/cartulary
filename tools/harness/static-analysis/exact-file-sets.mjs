import path from "node:path";
import { createSourceAdmission, validateRepositoryRelativePath } from "./source-admission.mjs";
export { validateRepositoryRelativePath } from "./source-admission.mjs";

const sourceExtensions = new Set([".go", ".mjs", ".js", ".sh", ".sql"]);

function requireNonemptyString(value, label) {
  if (typeof value !== "string" || value.length === 0 || value.trim() !== value) {
    throw new Error(`${label} must be a non-empty, unpadded string`);
  }
  return value;
}

function requireUniqueStrings(value, label, { paths = false } = {}) {
  if (!Array.isArray(value) || value.length === 0) {
    throw new Error(`${label} must be a non-empty array`);
  }
  const result = value.map((entry, index) => {
    const itemLabel = `${label}[${index + 1}]`;
    return paths
      ? validateRepositoryRelativePath(entry, itemLabel)
      : requireNonemptyString(entry, itemLabel);
  });
  if (new Set(result).size !== result.length) {
    throw new Error(`${label} must contain unique values`);
  }
  return result;
}

export function normalizeExactFileSets(rawSets) {
  if (!Array.isArray(rawSets)) {
    throw new Error("exact_file_sets must be an array");
  }
  const sets = rawSets.map((raw, index) => {
    const label = `exact_file_sets[${index + 1}]`;
    if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
      throw new Error(`${label} must be an object`);
    }
    const id = requireNonemptyString(raw.id, `${label}.id`);
    const paths = requireUniqueStrings(raw.paths, `${label}.paths`, { paths: true });
    const discovery = raw.discovery;
    if (!discovery || typeof discovery !== "object" || Array.isArray(discovery)) {
      throw new Error(`${label}.discovery must be an object`);
    }
    const scanRoots = requireUniqueStrings(
      discovery.scan_roots,
      `${label}.discovery.scan_roots`,
      { paths: true },
    );
    const tokens = requireUniqueStrings(discovery.tokens, `${label}.discovery.tokens`);
    if (
      discovery.production_only !== undefined &&
      typeof discovery.production_only !== "boolean"
    ) {
      throw new Error(`${label}.discovery.production_only must be a boolean`);
    }
    return {
      id,
      paths,
      discovery: {
        scanRoots,
        tokens,
        productionOnly: discovery.production_only ?? true,
      },
    };
  });
  const ids = sets.map((entry) => entry.id);
  if (new Set(ids).size !== ids.length) {
    throw new Error("exact_file_sets ids must be unique");
  }
  return sets;
}

function discoverFiles(admission, exactSet) {
  const discovered = new Set();
  for (let rootIndex = 0; rootIndex < exactSet.discovery.scanRoots.length; rootIndex += 1) {
    const relativeRoot = exactSet.discovery.scanRoots[rootIndex];
    const label = `exact_file_sets.${exactSet.id}.discovery.scan_roots[${rootIndex + 1}]`;
    admission.directory(relativeRoot, label);
    walkDiscoveryRoot(admission, relativeRoot, exactSet, discovered);
  }
  return [...discovered].sort((left, right) => left.localeCompare(right));
}

function walkDiscoveryRoot(admission, directory, exactSet, discovered) {
  const label = `exact_file_sets.${exactSet.id}.discovery`;
  for (const entry of admission.entries(directory, label)) {
    const relative = `${directory}/${entry.name}`;
    if (entry.isDirectory()) { walkDiscoveryRoot(admission, relative, exactSet, discovered); continue; }
    if (!sourceExtensions.has(path.extname(entry.name))) continue;
    if (exactSet.discovery.productionOnly && relative.endsWith("_test.go")) continue;
    const {content} = admission.file(relative, label);
    if (exactSet.discovery.tokens.some(token => content.includes(token))) discovered.add(relative);
  }
}

export function resolveExactFileSets(root, exactSets, admission = createSourceAdmission(root)) {
  const resolved = new Map();
  for (const exactSet of exactSets) {
    const files = exactSet.paths.map((relative, index) => {
      const label = `exact_file_sets.${exactSet.id}.paths[${index + 1}]`;
      return admission.file(relative, label);
    });
    if (files.length === 0) {
      throw new Error(`exact_file_sets.${exactSet.id} resolved no files`);
    }
    const discovered = discoverFiles(admission, exactSet);
    if (discovered.length === 0) {
      throw new Error(`exact_file_sets.${exactSet.id} discovery resolved no files`);
    }
    const declared = files.map((file) => file.relative).sort((left, right) => left.localeCompare(right));
    if (
      declared.length !== discovered.length ||
      declared.some((relative, index) => relative !== discovered[index])
    ) {
      const declaredSet = new Set(declared);
      const discoveredSet = new Set(discovered);
      const missing = discovered.filter((relative) => !declaredSet.has(relative));
      const stale = declared.filter((relative) => !discoveredSet.has(relative));
      throw new Error(
        `exact_file_sets.${exactSet.id} does not equal token discovery; missing=[${missing.join(",")}] stale=[${stale.join(",")}]`,
      );
    }
    resolved.set(exactSet.id, files);
  }
  return resolved;
}
