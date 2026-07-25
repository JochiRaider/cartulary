import { lstatSync, readFileSync, realpathSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import { parseStrictJSON, semanticJSONDigest } from "./semantic-json.mjs";

const registrySchemaID = "cartulary.requirement_registry.v1";
const catalogSchemaID = "cartulary.requirement_catalog.v1";
const contractFamilyRegistrySchemaID =
  "cartulary.contract_family_registry.v2";
const registryPath = "contracts/requirements/registry.json";
const catalogRootPath = "contracts/requirements/owners";
const contractFamilyRegistryPath = "contracts/index.json";

function readJSON(file) {
  return parseStrictJSON(readFileSync(file, "utf8"), file);
}

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function assertSortedUnique(values, label) {
  const sorted = [...values].sort(asciiCompare);
  if (JSON.stringify(values) !== JSON.stringify(sorted)) {
    throw new Error(`${label} must be ASCII-sorted`);
  }
  if (new Set(values).size !== values.length) {
    throw new Error(`${label} must not contain duplicates`);
  }
}

function assertContainedRegularFile(root, relativeFile, expectedRoot, label) {
  if (path.isAbsolute(relativeFile) || relativeFile.includes("\\")) {
    throw new Error(`${label} must be a normalized repository-relative path`);
  }
  const candidate = path.resolve(root, relativeFile);
  const containmentRoot = realpathSync(path.resolve(root, expectedRoot));
  const resolved = realpathSync(candidate);
  if (
    resolved !== containmentRoot &&
    !resolved.startsWith(`${containmentRoot}${path.sep}`)
  ) {
    throw new Error(`${label} escapes ${expectedRoot}`);
  }
  const stat = lstatSync(candidate);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error(`${label} must reference a non-symlink regular file`);
  }
  return candidate;
}

export function loadRequirements(root) {
  const registry = readJSON(path.join(root, registryPath));
  validateSchemaSync(registrySchemaID, registry);
  assertSortedUnique(
    registry.owners.map((entry) => entry.owner_id),
    `${registryPath}.owners.owner_id`,
  );
  assertSortedUnique(
    registry.owners.map((entry) => entry.catalog_path),
    `${registryPath}.owners.catalog_path`,
  );

  const catalogs = [];
  const requirementByID = new Map();
  for (const owner of registry.owners) {
    const file = assertContainedRegularFile(
      root,
      owner.catalog_path,
      catalogRootPath,
      `${registryPath}:${owner.owner_id}.catalog_path`,
    );
    const catalog = readJSON(file);
    validateSchemaSync(catalogSchemaID, catalog);
    if (catalog.owner_id !== owner.owner_id) {
      throw new Error(
        `${owner.catalog_path}.owner_id must equal registry owner ${owner.owner_id}`,
      );
    }
    assertSortedUnique(
      catalog.requirements.map((entry) => entry.requirement_id),
      `${owner.catalog_path}.requirements.requirement_id`,
    );
    for (const requirement of catalog.requirements) {
      if (requirement.contract_ids) {
        assertSortedUnique(
          requirement.contract_ids,
          `${owner.catalog_path}:${requirement.requirement_id}.contract_ids`,
        );
      }
      if (requirementByID.has(requirement.requirement_id)) {
        throw new Error(`duplicate requirement_id ${requirement.requirement_id}`);
      }
      requirementByID.set(requirement.requirement_id, {
        owner_id: owner.owner_id,
        catalog_path: owner.catalog_path,
        requirement,
      });
    }
    catalogs.push(catalog);
  }

  const contractFamilyRegistry = readJSON(
    path.join(root, contractFamilyRegistryPath),
  );
  validateSchemaSync(contractFamilyRegistrySchemaID, contractFamilyRegistry);
  for (const family of contractFamilyRegistry.families) {
    const label = `${contractFamilyRegistryPath}:${family.family_id}`;
    assertSortedUnique(
      family.owner_requirement_ids,
      `${label}.owner_requirement_ids`,
    );
    assertSortedUnique(
      family.owner_contract_ids,
      `${label}.owner_contract_ids`,
    );
    const declaredContractIDs = new Set();
    for (const requirementID of family.owner_requirement_ids) {
      const resolved = requirementByID.get(requirementID);
      if (!resolved) {
        throw new Error(
          `${label}.owner_requirement_ids references unknown ${requirementID}`,
        );
      }
      if (resolved.requirement.status !== family.generation_status) {
        throw new Error(
          `${label}.generation_status ${family.generation_status} does not match requirement ${requirementID} status ${resolved.requirement.status}`,
        );
      }
      for (const contractID of resolved.requirement.contract_ids ?? []) {
        declaredContractIDs.add(contractID);
      }
    }
    for (const contractID of family.owner_contract_ids) {
      if (!declaredContractIDs.has(contractID)) {
        throw new Error(
          `${label}.owner_contract_ids references undeclared ${contractID}`,
        );
      }
    }
  }

  return {
    registry,
    catalogs,
    contractFamilyRegistry,
    requirementByID,
    semantic_digest: semanticJSONDigest({
      schema_id: registry.schema_id,
      owners: registry.owners.map((owner, index) => ({
        ...owner,
        catalog: catalogs[index],
      })),
      contract_family_registry: contractFamilyRegistry,
    }),
  };
}

export function validateRequirements(root) {
  return loadRequirements(root);
}
