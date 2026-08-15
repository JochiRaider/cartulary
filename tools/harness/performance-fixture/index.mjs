import { createHash } from "node:crypto";
import { existsSync, lstatSync, readFileSync, readdirSync, realpathSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import {
  canonicalJSONString,
  parseStrictJSON,
  semanticJSONSHA256,
} from "../contract/index.mjs";
export { validateProfileMeasurementObservation } from "./profile-adapters/index.mjs";

export const registrySchemaID = "cartulary.performance_fixture_snapshot_owner.v2";
export const snapshotKeySchemaID = "cartulary.performance_fixture_snapshot_key.v2";
export const defaultRegistryPath = "tools/performance_fixture_snapshot_owner.json";
const migrationRunnerIdentity = "cartulary-postgres-migrate/goose/v3.27.0";
const supportedArtifactPolicyGenerations = new Set([
  [
    "cartulary.performance_fixture_snapshot_key.v2",
    "cartulary.performance_fixture_snapshot.v2",
    "cartulary.performance_fixture_snapshot_lease.v2",
    "cartulary.performance_fixture_runtime.v2",
    "cartulary.frontend_measurement_observation.v2",
    "cartulary.frontend_measurement_summary.v3",
    "cartulary.frontend_measurement_aggregate.v3",
  ].join("\u0000"),
]);
const snapshotKeyEnvelopeFields = new Map([
  [
    "cartulary.performance_fixture_snapshot_key.v2",
    [
      "fixture_profile_id",
      "migration_digest",
      "source_contract_digest",
      "fixture_version",
      "seed",
    ],
  ],
]);

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function assertSortedUnique(values, label) {
  const sorted = [...values].sort(compareASCII);
  if (new Set(values).size !== values.length || JSON.stringify(values) !== JSON.stringify(sorted)) {
    throw new Error(label + " must be ASCII-sorted and duplicate-free");
  }
}

function assertSortedUniqueBy(entries, key, label) {
  const values = entries.map((entry) => entry[key]);
  assertSortedUnique(values, label);
}

function deepFreeze(value) {
  if (value !== null && typeof value === "object" && !Object.isFrozen(value)) {
    for (const child of Object.values(value)) deepFreeze(child);
    Object.freeze(value);
  }
  return value;
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
    assertSortedUniqueBy(
      contribution.expected_receipt_counts,
      "count_id",
      contributionLabel + ".expected_receipt_counts.count_id",
    );
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

function validateSemanticExpectations(profile, label) {
  assertSortedUniqueBy(
    profile.semantic_expectations.counts,
    "expectation_id",
    label + ".semantic_expectations.counts.expectation_id",
  );
  assertSortedUniqueBy(
    profile.semantic_expectations.conditions,
    "expectation_id",
    label + ".semantic_expectations.conditions.expectation_id",
  );
  assertSortedUniqueBy(
    profile.runtime_credential_sets,
    "set_id",
    label + ".runtime_credential_sets.set_id",
  );
  assertSortedUnique(profile.redaction_policy.forbidden_fields, label + ".redaction_policy.forbidden_fields");
}

function validateArtifactPolicy(profile, label) {
  const policy = profile.artifact_policy;
  const generation = [
    policy.snapshot_key_schema_id,
    policy.build_schema_id,
    policy.lease_schema_id,
    policy.runtime_schema_id,
    policy.observation_schema_id,
    policy.summary_schema_id,
    policy.aggregate_schema_id,
  ].join("\u0000");
  if (!supportedArtifactPolicyGenerations.has(generation)) {
    throw new Error(label + ".artifact_policy selects an unsupported or mixed evidence generation");
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
      expected_receipt_counts: entry.expected_receipt_counts,
    })),
    semantic_expectations: profile.semantic_expectations,
    runtime_credential_sets: profile.runtime_credential_sets,
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
    assertSortedUnique(
      rawProfile.compatibility.service_dependencies,
      label + ".compatibility.service_dependencies",
    );
    validateBindings(rawProfile, label);
    const sourceContracts = validateSourceContracts(root, rawProfile, label);
    validateContributions(rawProfile, sourceContracts, label);
    validateSemanticExpectations(rawProfile, label);
    validateArtifactPolicy(rawProfile, label);
    const profile = deepFreeze({
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

export function activePerformanceFixtureProfile(registry, fixtureProfileID) {
  const profile = registry.profiles.get(fixtureProfileID);
  if (profile === undefined || profile.status !== "active") {
    throw new Error(`performance fixture profile is unknown or inactive: ${fixtureProfileID}`);
  }
  return profile;
}

export function performanceFixtureBindingsForRows(root, rows, options = {}) {
  const registry = options.registry ?? loadPerformanceFixtureSnapshotRegistry(root, options);
  const bindings = [];
  for (const row of rows) {
    const rowBindings = row.verification_ids.flatMap((verificationID) => {
      const binding = registry.verificationBindings.get(verificationID);
      return binding === undefined ? [] : [{ verificationID, ...binding }];
    });
    if (rowBindings.length === 0) {
      if (row.fixture_profile_id) {
        throw new Error(`${row.row_id} declares a fixture profile without a verification binding`);
      }
      continue;
    }
    const profileIDs = new Set(rowBindings.map((binding) => binding.fixture_profile_id));
    if (profileIDs.size !== 1 || !profileIDs.has(row.fixture_profile_id)) {
      throw new Error(`${row.row_id} fixture profile diverges from its verification bindings`);
    }
    const profile = activePerformanceFixtureProfile(registry, row.fixture_profile_id);
    for (const binding of rowBindings) {
      bindings.push(deepFreeze({
        fixture_profile_id: profile.fixture_profile_id,
        predicate_id: binding.predicate_id,
        row_id: row.row_id,
        verification_id: binding.verificationID,
      }));
    }
  }
  return bindings;
}

export function performanceFixturePredicateIDsForRows(root, rows, options = {}) {
  return performanceFixtureBindingsForRows(root, rows, options)
    .map((binding) => binding.predicate_id);
}

export function groupRowsByPerformanceFixture(root, rows, options = {}) {
  const registry = options.registry ?? loadPerformanceFixtureSnapshotRegistry(root, options);
  const bindings = performanceFixtureBindingsForRows(root, rows, { ...options, registry });
  const byProfile = new Map();
  for (const binding of bindings) {
    const group = byProfile.get(binding.fixture_profile_id) ?? {
      fixture_profile_id: binding.fixture_profile_id,
      predicate_ids: new Set(),
      row_ids: new Set(),
      verification_ids: new Set(),
    };
    group.predicate_ids.add(binding.predicate_id);
    group.row_ids.add(binding.row_id);
    group.verification_ids.add(binding.verification_id);
    byProfile.set(binding.fixture_profile_id, group);
  }
  return [...byProfile.values()]
    .map((group) => deepFreeze({
      fixture_profile_id: group.fixture_profile_id,
      predicate_ids: [...group.predicate_ids].sort(compareASCII),
      profile: activePerformanceFixtureProfile(registry, group.fixture_profile_id),
      row_ids: [...group.row_ids].sort(compareASCII),
      verification_ids: [...group.verification_ids].sort(compareASCII),
    }))
    .sort((left, right) => compareASCII(left.fixture_profile_id, right.fixture_profile_id));
}

export function snapshotKeyEnvelope(profile, migrationDigest) {
  if (!/^[a-f0-9]{64}$/u.test(migrationDigest)) {
    throw new Error("migration_digest must be 64 lowercase hexadecimal characters");
  }
  if (!profile || profile.status !== "active") {
    throw new Error("snapshot profile must be active");
  }
  const schemaID = profile.artifact_policy?.snapshot_key_schema_id;
  const fields = snapshotKeyEnvelopeFields.get(schemaID);
  if (fields === undefined) {
    throw new Error("snapshot key schema is unsupported: " + String(schemaID));
  }
  const values = {
    fixture_profile_id: profile.fixture_profile_id,
    migration_digest: migrationDigest,
    source_contract_digest: profile.source_contract_digest,
    fixture_version: profile.fixture_version,
    seed: profile.seed,
  };
  const envelope = { schema_id: schemaID };
  for (const field of fields) envelope[field] = values[field];
  validateSchemaSync(schemaID, envelope);
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
