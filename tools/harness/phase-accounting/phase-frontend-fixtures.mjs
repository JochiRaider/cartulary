import { readFileSync } from "node:fs";
import path from "node:path";

import { frontendVisualFixtureIDPattern } from "./frontend/phase-ids.mjs";
import { frontendVisualFixtureRegistryPath } from "./frontend/visual-fixtures.mjs";
import { readJsonObject } from "../contract/json-shape.mjs";

const frontendFixtureRefIDsByRoot = new Map();

export function validateFixtureRefs(entry, label) {
  if (entry.fixture_refs === undefined) {
    return;
  }
  if (!Array.isArray(entry.fixture_refs) || entry.fixture_refs.length === 0) {
    throw new Error(`${label} fixture_refs must be a non-empty string array when present`);
  }
  const seen = new Set();
  for (const [index, ref] of entry.fixture_refs.entries()) {
    if (typeof ref !== "string" || !/^VFIX-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d{2}$/.test(ref)) {
      throw new Error(`${label} fixture_refs[${index + 1}] must be a VFIX-* fixture identifier`);
    }
    if (seen.has(ref)) {
      throw new Error(`${label} fixture_refs contains duplicate ${ref}`);
    }
    seen.add(ref);
  }
}

export function validateFrontendFixtureRefs(root, entry, label) {
  if (entry.frontend_fixture_refs === undefined) {
    return;
  }
  if (!Array.isArray(entry.frontend_fixture_refs) || entry.frontend_fixture_refs.length === 0) {
    throw new Error(`${label} frontend_fixture_refs must be a non-empty string array when present`);
  }
  const validRefs = frontendVisualFixtureRefIDs(root);
  const seen = new Set();
  for (const [index, ref] of entry.frontend_fixture_refs.entries()) {
    if (typeof ref !== "string" || !frontendVisualFixtureIDPattern.test(ref)) {
      throw new Error(
        `${label} frontend_fixture_refs[${index + 1}] must be an FE-VFIX-* fixture identifier`,
      );
    }
    if (!validRefs.has(ref)) {
      throw new Error(
        `${label} frontend_fixture_refs[${index + 1}] references unknown frontend fixture ${ref}`,
      );
    }
    if (seen.has(ref)) {
      throw new Error(`${label} frontend_fixture_refs contains duplicate ${ref}`);
    }
    seen.add(ref);
  }
}

function frontendVisualFixtureRefIDs(root) {
  const normalizedRoot = path.resolve(root);
  const cached = frontendFixtureRefIDsByRoot.get(normalizedRoot);
  if (cached !== undefined) {
    return cached;
  }
  const file = frontendVisualFixtureRegistryPath(normalizedRoot);
  const registry = readJsonObject(file, file);
  if (!Array.isArray(registry.fixtures)) {
    throw new Error(`${file}.fixtures must be an array`);
  }
  const refs = new Set();
  for (const [index, fixture] of registry.fixtures.entries()) {
    const ref = fixture?.fixture_id;
    if (typeof ref !== "string" || !frontendVisualFixtureIDPattern.test(ref)) {
      throw new Error(`${file}.fixtures[${index + 1}].fixture_id must be an FE-VFIX-* fixture identifier`);
    }
    refs.add(ref);
  }
  frontendFixtureRefIDsByRoot.set(normalizedRoot, refs);
  return refs;
}

export function assertAuthoritativeGridRowsUseLiveAdapter(root, entry, label) {
  if (
    entry.coverage !== "authoritative" ||
    entry.runner !== "vitest" ||
    !/^U-\d+-GRID-/.test(entry.id)
  ) {
    return;
  }
  const source = readFileSync(path.join(root, entry.file), "utf8");
  const mocksGridAdapter =
    /vi\s*\.\s*mock\s*\(\s*["']@cartulary\/grid-adapter["']/.test(source) ||
    /mock\s*\(\s*["']@cartulary\/grid-adapter["']/.test(source);
  if (mocksGridAdapter && source.includes("@cartulary/grid-adapter/test-support")) {
    throw new Error(
      `${label} must use the production @cartulary/grid-adapter path, not @cartulary/grid-adapter/test-support`,
    );
  }
}
