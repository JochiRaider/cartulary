import { createHash } from "node:crypto";
import { existsSync, lstatSync, readFileSync, readdirSync, realpathSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import {
  canonicalJSONString,
  parseStrictJSON,
  semanticJSONSHA256,
} from "./semantic-json.mjs";

export const registrySchemaID = "cartulary.performance_fixture_snapshot_owner.v1";
export const snapshotKeySchemaID = "cartulary.performance_fixture_snapshot_key.v1";
export const defaultRegistryPath = "tools/performance_fixture_snapshot_owner.json";
const migrationRunnerIdentity = "cartulary-postgres-migrate/goose/v3.27.0";

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function assertSortedUnique(values, label) {
  const sorted = [...values].sort(compareASCII);
  if (new Set(values).size !== values.length || JSON.stringify(values) !== JSON.stringify(sorted)) {
    throw new Error(label + " must be ASCII-sorted and duplicate-free");
  }
}

function readStrictJSON(file) {
  return parseStrictJSON(readFileSync(file, "utf8"), file);
}

function resolveContractFile(root, relativePath, label) {
  if (path.isAbsolute(relativePath) || relativePath.includes("\\")) {
    throw new Error(label + " must be a normalized repository-relative path");
  }
  const normalized = path.posix.normalize(relativePath);
  if (normalized !== relativePath || normalized.startsWith("../")) {
    throw new Error(label + " contains traversal or normalization drift");
  }
  const candidate = path.resolve(root, relativePath);
  if (!existsSync(candidate)) throw new Error(label + " does not exist: " + relativePath);
  const stat = lstatSync(candidate);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error(label + " must reference a non-symlink regular file");
  }
  const contractsRoot = realpathSync(path.join(root, "contracts"));
  const resolved = realpathSync(candidate);
  if (!resolved.startsWith(contractsRoot + path.sep)) {
    throw new Error(label + " must remain under contracts");
  }
  return candidate;
}

function validateSourceContracts(root, profile, label) {
  const paths = profile.source_contract_refs.map((entry) => entry.path);
  assertSortedUnique(paths, label + ".source_contract_refs.path");
  const contractIDs = profile.source_contract_refs.map((entry) => entry.contract_id);
  assertSortedUnique(contractIDs, label + ".source_contract_refs.contract_id");
  const byID = new Map();
  for (const ref of profile.source_contract_refs) {
    const file = resolveContractFile(root, ref.path, label + "." + ref.contract_id + ".path");
    const source = readStrictJSON(file);
    if (source.schema_id !== ref.schema_id) {
      throw new Error(label + "." + ref.contract_id + " schema_id does not match " + ref.path);
    }
    const sourceIdentity = source.contract_id ?? source.view_schema_id ?? source.schema_id;
    if (sourceIdentity !== ref.contract_id) {
      throw new Error(label + "." + ref.contract_id + " contract identity does not match " + ref.path);
    }
    byID.set(ref.contract_id, ref);
  }
  return byID;
}

function validateContributions(profile, sourceContracts, label) {
  const contributionIDs = profile.contributions.map((entry) => entry.contribution_id);
  if (new Set(contributionIDs).size !== contributionIDs.length) {
    throw new Error(label + ".contributions contains duplicate contribution_id");
  }
  const known = new Set(contributionIDs);
  const admitted = new Set();
  for (const [index, contribution] of profile.contributions.entries()) {
    const contributionLabel = label + ".contributions[" + (index + 1) + "]";
    assertSortedUnique(contribution.dependencies, contributionLabel + ".dependencies");
    assertSortedUnique(contribution.source_contract_ids, contributionLabel + ".source_contract_ids");
    if (contribution.version !== contribution.contribution_id) {
      throw new Error(contributionLabel + ".version must equal contribution_id");
    }
    for (const dependency of contribution.dependencies) {
      if (!known.has(dependency)) {
        throw new Error(contributionLabel + " references unknown dependency " + dependency);
      }
      if (!admitted.has(dependency)) {
        throw new Error(contributionLabel + " dependency " + dependency + " must precede its dependant");
      }
    }
    for (const contractID of contribution.source_contract_ids) {
      if (!sourceContracts.has(contractID)) {
        throw new Error(contributionLabel + " references unknown source contract " + contractID);
      }
    }
    admitted.add(contribution.contribution_id);
  }
}

function validateBindings(profile, label) {
  const verificationIDs = profile.verification_bindings.map((entry) => entry.verification_id);
  const predicateIDs = profile.verification_bindings.map((entry) => entry.predicate_id);
  assertSortedUnique(verificationIDs, label + ".verification_bindings.verification_id");
  assertSortedUnique(predicateIDs, label + ".verification_bindings.predicate_id");
}

function sourceContractProjection(profile) {
  return {
    schema_id: registrySchemaID,
    fixture_profile_id: profile.fixture_profile_id,
    fixture_version: profile.fixture_version,
    seed: profile.seed,
    source_contract_refs: profile.source_contract_refs.map((entry) => ({
      contract_id: entry.contract_id,
      schema_id: entry.schema_id,
      path: entry.path,
    })),
    contributions: profile.contributions.map((entry) => ({
      contribution_id: entry.contribution_id,
      owner_id: entry.owner_id,
      version: entry.version,
      dependencies: entry.dependencies,
      source_contract_ids: entry.source_contract_ids,
      expected_receipt: entry.expected_receipt,
    })),
    validation_rules: profile.validation_rules,
  };
}

export function loadPerformanceFixtureSnapshotRegistry(root, options = {}) {
  const registryPath = path.resolve(root, options.registryPath ?? defaultRegistryPath);
  const registry = readStrictJSON(registryPath);
  validateSchemaSync(registrySchemaID, registry);
  const profileIDs = registry.profiles.map((entry) => entry.fixture_profile_id);
  assertSortedUnique(profileIDs, registryPath + ".profiles.fixture_profile_id");
  const profiles = new Map();
  const verificationBindings = new Map();
  for (const [index, rawProfile] of registry.profiles.entries()) {
    const label = registryPath + ".profiles[" + (index + 1) + "]";
    validateBindings(rawProfile, label);
    const sourceContracts = validateSourceContracts(root, rawProfile, label);
    validateContributions(rawProfile, sourceContracts, label);
    const profile = Object.freeze({
      ...structuredClone(rawProfile),
      source_contract_digest: semanticJSONSHA256(sourceContractProjection(rawProfile)),
    });
    profiles.set(profile.fixture_profile_id, profile);
    for (const binding of profile.verification_bindings) {
      if (verificationBindings.has(binding.verification_id)) {
        throw new Error("duplicate fixture verification binding " + binding.verification_id);
      }
      verificationBindings.set(binding.verification_id, {
        fixture_profile_id: profile.fixture_profile_id,
        predicate_id: binding.predicate_id,
      });
    }
  }
  return {
    registry,
    profiles,
    verificationBindings,
    semantic_projection: registry.profiles.map(sourceContractProjection),
  };
}

export function snapshotKeyEnvelope(profile, migrationDigest) {
  if (!/^[a-f0-9]{64}$/u.test(migrationDigest)) {
    throw new Error("migration_digest must be 64 lowercase hexadecimal characters");
  }
  if (!profile || profile.status !== "active") {
    throw new Error("snapshot profile must be active");
  }
  const envelope = {
    schema_id: snapshotKeySchemaID,
    migration_digest: migrationDigest,
    source_contract_digest: profile.source_contract_digest,
    fixture_version: profile.fixture_version,
    seed: profile.seed,
  };
  validateSchemaSync(snapshotKeySchemaID, envelope);
  return envelope;
}

export function snapshotKey(profile, migrationDigest) {
  return semanticJSONSHA256(snapshotKeyEnvelope(profile, migrationDigest));
}

export function canonicalSnapshotKeyInput(profile, migrationDigest) {
  return canonicalJSONString(snapshotKeyEnvelope(profile, migrationDigest));
}

export function postgresMigrationDigest(root) {
  const migrationRoot = path.join(root, "db", "migrations");
  const entries = readdirSync(migrationRoot, { withFileTypes: true })
    .filter((entry) => entry.isFile() && /^[0-9]{5}_[a-z0-9]+(?:_[a-z0-9]+)*\.sql$/u.test(entry.name))
    .map((entry) => entry.name)
    .sort(compareASCII);
  if (entries.length === 0) {
    throw new Error("canonical migration source is empty");
  }
  const versions = new Set();
  for (const [index, filename] of entries.entries()) {
    const version = Number(filename.slice(0, 5));
    if (version !== index + 1 || versions.has(version)) {
      throw new Error("canonical migration versions must be contiguous and unique");
    }
    versions.add(version);
  }
  const hash = createHash("sha256");
  hash.update(migrationRunnerIdentity);
  hash.update(Buffer.from([0]));
  for (const filename of entries) {
    hash.update(filename);
    hash.update(Buffer.from([0]));
    hash.update(readFileSync(path.join(migrationRoot, filename)));
    hash.update(Buffer.from([0]));
  }
  return hash.digest("hex");
}
