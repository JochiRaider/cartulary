import {
  lstatSync,
  readdirSync,
  readFileSync,
  realpathSync,
} from "node:fs";
import path from "node:path";

const sourceExtensions = new Set([".go", ".mjs", ".js", ".sh", ".sql"]);
const globSyntax = /[*?[\]{}]/u;
const windowsDrive = /^[A-Za-z]:\//u;

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

export function validateRepositoryRelativePath(value, label = "path") {
  const candidate = requireNonemptyString(value, label);
  if (candidate.includes("\\")) {
    throw new Error(`${label} must use repository-relative forward slashes`);
  }
  if (path.posix.isAbsolute(candidate) || windowsDrive.test(candidate)) {
    throw new Error(`${label} must be repository-relative`);
  }
  if (globSyntax.test(candidate)) {
    throw new Error(`${label} must be an exact path without glob syntax`);
  }
  const segments = candidate.split("/");
  if (segments.some((segment) => segment === "" || segment === "." || segment === "..")) {
    throw new Error(`${label} must be normalized without empty, '.' or '..' segments`);
  }
  if (path.posix.normalize(candidate) !== candidate) {
    throw new Error(`${label} must be normalized`);
  }
  return candidate;
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

function containedRelative(root, absolute, label) {
  const relative = path.relative(root, absolute);
  if (relative === ".." || relative.startsWith(`..${path.sep}`) || path.isAbsolute(relative)) {
    throw new Error(`${label} escapes the repository root`);
  }
}

function resolveWithoutSymlinks(root, relative, expectedKind, label) {
  const rootReal = realpathSync(root);
  let current = rootReal;
  const segments = relative.split("/");
  for (let index = 0; index < segments.length; index += 1) {
    current = path.join(current, segments[index]);
    let stat;
    try {
      stat = lstatSync(current);
    } catch (error) {
      if (error?.code === "ENOENT") {
        throw new Error(`${label} does not exist`);
      }
      throw error;
    }
    if (stat.isSymbolicLink()) {
      throw new Error(`${label} must not contain a symlink component`);
    }
    if (index < segments.length - 1 && !stat.isDirectory()) {
      throw new Error(`${label} has a non-directory parent component`);
    }
    if (index === segments.length - 1) {
      const expected = expectedKind === "file" ? stat.isFile() : stat.isDirectory();
      if (!expected) {
        throw new Error(`${label} must resolve to a regular ${expectedKind}`);
      }
    }
  }
  containedRelative(rootReal, realpathSync(current), label);
  return current;
}

function discoverFiles(root, exactSet) {
  const discovered = new Set();
  for (let rootIndex = 0; rootIndex < exactSet.discovery.scanRoots.length; rootIndex += 1) {
    const relativeRoot = exactSet.discovery.scanRoots[rootIndex];
    const label = `exact_file_sets.${exactSet.id}.discovery.scan_roots[${rootIndex + 1}]`;
    const absoluteRoot = resolveWithoutSymlinks(root, relativeRoot, "directory", label);
    walkDiscoveryRoot(root, absoluteRoot, exactSet, discovered);
  }
  return [...discovered].sort((left, right) => left.localeCompare(right));
}

function walkDiscoveryRoot(root, directory, exactSet, discovered) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const absolute = path.join(directory, entry.name);
    const relative = path.relative(root, absolute).split(path.sep).join("/");
    if (entry.isSymbolicLink()) {
      throw new Error(`exact_file_sets.${exactSet.id}.discovery found symlink ${relative}`);
    }
    if (entry.isDirectory()) {
      walkDiscoveryRoot(root, absolute, exactSet, discovered);
      continue;
    }
    if (!entry.isFile()) {
      throw new Error(`exact_file_sets.${exactSet.id}.discovery found non-regular path ${relative}`);
    }
    if (!sourceExtensions.has(path.extname(entry.name))) {
      continue;
    }
    if (exactSet.discovery.productionOnly && relative.endsWith("_test.go")) {
      continue;
    }
    const content = readFileSync(absolute, "utf8");
    if (exactSet.discovery.tokens.some((token) => content.includes(token))) {
      discovered.add(relative);
    }
  }
}

export function resolveExactFileSets(root, exactSets) {
  const rootAbsolute = path.resolve(root);
  const rootReal = realpathSync(rootAbsolute);
  const resolved = new Map();
  for (const exactSet of exactSets) {
    const files = exactSet.paths.map((relative, index) => {
      const label = `exact_file_sets.${exactSet.id}.paths[${index + 1}]`;
      const absolute = resolveWithoutSymlinks(rootReal, relative, "file", label);
      return { absolute, relative, content: readFileSync(absolute, "utf8") };
    });
    if (files.length === 0) {
      throw new Error(`exact_file_sets.${exactSet.id} resolved no files`);
    }
    const discovered = discoverFiles(rootReal, exactSet);
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
