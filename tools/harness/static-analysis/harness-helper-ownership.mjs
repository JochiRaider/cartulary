import { existsSync, readFileSync } from "node:fs";
import path from "node:path";

const harnessHelperOwnershipSchemaID =
  "cartulary.harness_helper_ownership.v1";

export function loadHarnessHelperOwnership(root = process.cwd()) {
  const file = path.join(root, "tools/harness_helper_ownership.json");
  const owner = JSON.parse(readFileSync(file, "utf8"));
  if (owner.schema_id !== harnessHelperOwnershipSchemaID) {
    throw new Error(`${file}.schema_id must be ${harnessHelperOwnershipSchemaID}`);
  }
  const seenKeys = new Set();
  const seenPaths = new Set();
  for (const facade of owner.facades ?? []) {
    if (seenKeys.has(facade.key)) {
      throw new Error(`${file} duplicates facade key ${facade.key}`);
    }
    seenKeys.add(facade.key);
    for (const facadePath of facade.paths ?? []) {
      if (seenPaths.has(facadePath)) {
        throw new Error(`${file} assigns ${facadePath} to more than one facade`);
      }
      if (!existsSync(path.join(root, facadePath))) {
        throw new Error(`${file} references missing facade ${facadePath}`);
      }
      seenPaths.add(facadePath);
    }
  }
  return owner;
}

export function ownerFacadePathLists(owner) {
  const groups = {};
  for (const facade of owner.facades ?? []) {
    groups[facade.boundary_group] ??= [];
    groups[facade.boundary_group].push(...facade.paths);
  }
  for (const paths of Object.values(groups)) {
    paths.sort((left, right) => left.localeCompare(right));
  }
  return groups;
}

export function allowedPrivateImportSources(owner) {
  return new Set(
    (owner.facades ?? []).flatMap((facade) => facade.allowed_consumers ?? []),
  );
}
