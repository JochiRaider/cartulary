import { lstatSync, readFileSync, realpathSync } from "node:fs";
import path from "node:path";

import {
  parseStrictJSON,
  semanticJSONDigest,
  validateSchemaSync,
} from "../contract/index.mjs";

const registrySchemaID = "cartulary.verification_registry.v3";
const contractSchemaID = "cartulary.verification_contract.v3";
const registryPath = "contracts/verification/registry.json";
const contractRootPath = "contracts/verification/owners";

function readJSON(file) {
  return parseStrictJSON(readFileSync(file, "utf8"), file);
}

function assertSortedUnique(values, label) {
  const sorted = [...values].sort((left, right) => left < right ? -1 : left > right ? 1 : 0);
  if (JSON.stringify(values) !== JSON.stringify(sorted)) {
    throw new Error(`${label} must be ASCII-sorted`);
  }
  if (new Set(values).size !== values.length) {
    throw new Error(`${label} must not contain duplicates`);
  }
}

function assertContainedRegularFile(root, relativeFile, expectedRoot, label) {
  if (path.isAbsolute(relativeFile)) {
    throw new Error(`${label} must be repository-relative`);
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
  if (!lstatSync(candidate).isFile()) {
    throw new Error(`${label} must reference a regular file`);
  }
  if (lstatSync(candidate).isSymbolicLink()) {
    throw new Error(`${label} must not reference a symbolic link`);
  }
  return candidate;
}

function validateProfile(verification, label) {
  const { behavior_class: behaviorClass, profile } = verification;
  if (
    ["product", "security"].includes(behaviorClass) &&
    profile !== "base" &&
    !profile.startsWith("extension.")
  ) {
    throw new Error(`${label}.profile must be base or extension-owned`);
  }
  if (["build", "harness"].includes(behaviorClass) && profile !== "support") {
    throw new Error(`${label}.profile must be support`);
  }
  if (behaviorClass === "claim_publication" && !profile.startsWith("claim.")) {
    throw new Error(`${label}.profile must be claim-owned`);
  }
  if (
    behaviorClass === "architecture" &&
    profile !== "base" &&
    profile !== "support" &&
    !profile.startsWith("extension.")
  ) {
    throw new Error(`${label}.profile has no architecture authority`);
  }
}

export function loadVerificationContracts(root) {
  const registryFile = path.join(root, registryPath);
  const registry = readJSON(registryFile);
  validateSchemaSync(registrySchemaID, registry);

  const ownerIDs = registry.owners.map((entry) => entry.owner_id);
  const contractPaths = registry.owners.map((entry) => entry.contract_path);
  assertSortedUnique(ownerIDs, `${registryPath}.owners.owner_id`);
  assertSortedUnique(contractPaths, `${registryPath}.owners.contract_path`);

  const contracts = [];
  const verificationByID = new Map();
  for (const owner of registry.owners) {
    const file = assertContainedRegularFile(
      root,
      owner.contract_path,
      contractRootPath,
      `${registryPath}:${owner.owner_id}.contract_path`,
    );
    const contract = readJSON(file);
    validateSchemaSync(contractSchemaID, contract);
    if (contract.owner_id !== owner.owner_id) {
      throw new Error(
        `${owner.contract_path}.owner_id must equal registry owner ${owner.owner_id}`,
      );
    }
    const verificationIDs = contract.verifications.map(
      (entry) => entry.verification_id,
    );
    assertSortedUnique(
      verificationIDs,
      `${owner.contract_path}.verifications.verification_id`,
    );
    for (const [index, verification] of contract.verifications.entries()) {
      const label = `${owner.contract_path}.verifications[${index + 1}]`;
      if (!verification.verification_id.startsWith(`${owner.owner_id}.verification.`)) {
        throw new Error(`${label}.verification_id must be owner-qualified`);
      }
      assertSortedUnique(verification.evidence_kinds, `${label}.evidence_kinds`);
      validateProfile(verification, label);
      if (verificationByID.has(verification.verification_id)) {
        throw new Error(
          `duplicate verification_id ${verification.verification_id}`,
        );
      }
      verificationByID.set(verification.verification_id, {
        owner_id: owner.owner_id,
        contract_path: owner.contract_path,
        verification,
      });
    }
    contracts.push(contract);
  }

  const routingDigest = semanticJSONDigest({
    schema_id: registry.schema_id,
    owners: registry.owners.map((owner, index) => ({
      owner_id: owner.owner_id,
      contract_path: owner.contract_path,
      contract: contracts[index],
    })),
  });
  return {
    registry,
    contracts,
    verificationByID,
    routing_digest: routingDigest,
  };
}

export function validateVerificationContracts(root) {
  return loadVerificationContracts(root);
}
