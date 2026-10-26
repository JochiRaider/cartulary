import assert from "node:assert/strict";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";
import { validateSchemaSync } from "../../contract/index.mjs";
import { canonicalSnapshotKeyInput, groupRowsByPerformanceFixture, loadPerformanceFixtureSnapshotRegistry, postgresMigrationDigest, snapshotKey, snapshotKeyEnvelope } from "../index.mjs";
import { renderPerformanceFixtureKeyVectors, renderPerformanceFixtureProfilesGo } from "../../generated-artifacts/performance/performance-contracts.mjs";
import { validateFixtureProfile } from "../../test-catalog/index.mjs";
import { createContractTestContext } from "../../tests/contract-test-context.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
function assertPerformanceEvidenceGenerationBoundary() {
  const attachments = readJSON("tools/harness_schema_attachments.json");
  const activeIDs = new Set(attachments.attachments.map((entry) => entry.schema_id));
  for (const schemaID of [
    "cartulary.browser_target_result.v9",
    "cartulary.frontend_measurement_aggregate.v5",
    "cartulary.frontend_measurement_observation.v4",
    "cartulary.frontend_measurement_summary.v5",
    "cartulary.performance_fixture_build_diagnostics.v2",
    "cartulary.performance_fixture_runtime.v2",
    "cartulary.performance_fixture_snapshot.v3",
    "cartulary.performance_fixture_snapshot_key.v2",
    "cartulary.performance_fixture_snapshot_lease.v3",
    "cartulary.performance_fixture_snapshot_owner.v3",
    "cartulary.web_e2e_stack_lease.v2",
  ]) {
    assert.equal(activeIDs.has(schemaID), true);
  }
  // Independent rejection examples require no retired schema bundle or registry.
  for (const schema_id of ["cartulary.performance_fixture_snapshot.v2", "cartulary.performance_fixture_snapshot_lease.v2", "cartulary.performance_fixture_snapshot_owner.v2", "cartulary.performance_fixture_snapshot_key.v1", "cartulary.frontend_measurement_aggregate.v1", "cartulary.web_e2e_stack.v3", "cartulary.test.network_flow_randomness_control.v2"]) {
    assert.throws(() => validateSchemaSync(schema_id, { schema_id }), /unregistered schema attachment/u);
  }
  const reference = (role, name) => ({
    role,
    path_kind: "file",
    format: "json",
    path: `browser-e2e-measurement/${name}.json`,
    sha256: `sha256:${"a".repeat(64)}`,
  });
  const aggregate = {
    schema_id: "cartulary.frontend_measurement_aggregate.v5",
    target_id: "browser-e2e-measurement",
    status: "qualified",
    profile_groups: [{
        fixture_profile_id: "synthetic_grid_snapshot_v1",
        snapshot_key: "b".repeat(64),
        builder_count: 1,
        clone_count: 1,
        scheduler_overlap_count: 0,
        build_artifact: reference("performance_fixture_snapshot_build", "build"),
        summary_artifacts: [reference("frontend_measurement_summary", "summary")],
        rollups: [{
            predicate_id: "perf.synthetic.fixture_profile.v1",
            sample_count: 100,
            statistics: { kind: "percentile", percentile: 95, p50_ms: 10, p95_ms: 20 },
            threshold_ms: 100,
            outcome: "passed",
          }],
      }],
  };
  validateSchemaSync(aggregate.schema_id, aggregate);
  assert.throws(() => validateSchemaSync(aggregate.schema_id, {
    ...aggregate,
    summaries: [{ samples: Array.from({ length: 100000 }, () => 1) }],
  }), /additional properties/u, "aggregate v3 must grow by immutable summary references, not raw samples");
}
function assertPerformanceFixtureSnapshotContract(context) {
  const fixtureProfiles = context.catalog.fixtureProfiles;
  const profile = fixtureProfiles.profiles.get("timeline_large_grid_snapshot_v2");
  assert.ok(profile);
  assert.match(profile.source_contract_digest, /^[a-f0-9]{64}$/u);
  const migrationDigest = postgresMigrationDigest(root);
  const envelope = snapshotKeyEnvelope(profile, migrationDigest);
  validateSchemaSync(envelope.schema_id, envelope);
  assert.match(snapshotKey(profile, migrationDigest), /^[a-f0-9]{64}$/u);
  assert.notEqual(snapshotKey(profile, migrationDigest), snapshotKey(profile, "f".repeat(64)));
  const vectors = readJSON("tools/performance_fixture_snapshot_key_vectors.json");
  for (const vector of vectors.vectors) {
    const selectedProfile = fixtureProfiles.profiles.get(vector.name);
    assert.ok(selectedProfile);
    const vectorProfile = {
      ...selectedProfile,
      status: "active",
      source_contract_digest: vector.input.source_contract_digest,
      fixture_version: vector.input.fixture_version,
      seed: vector.input.seed,
    };
    assert.equal(canonicalSnapshotKeyInput(vectorProfile, vector.input.migration_digest), vector.canonical_json);
    assert.equal(snapshotKey(vectorProfile, vector.input.migration_digest), vector.snapshot_key);
  }
  const syntheticProfile = structuredClone(profile);
  syntheticProfile.fixture_profile_id = "synthetic_grid_snapshot_v1";
  syntheticProfile.fixture_version = "cartulary.perf.synthetic_grid.v1";
  syntheticProfile.seed += 1;
  syntheticProfile.source_contract_digest = "a".repeat(64);
  syntheticProfile.artifact_policy.snapshot_key_schema_id =
    "cartulary.performance_fixture_snapshot_key.v2";
  const syntheticVectors = JSON.parse(renderPerformanceFixtureKeyVectors([profile, syntheticProfile]));
  assert.deepEqual(syntheticVectors.vectors.map((entry) => entry.name), [profile.fixture_profile_id, syntheticProfile.fixture_profile_id].sort(compareASCII));
  const syntheticVector = syntheticVectors.vectors.find((entry) => entry.name === syntheticProfile.fixture_profile_id);
  assert.ok(syntheticVector);
  validateSchemaSync("cartulary.performance_fixture_snapshot_key.v2", syntheticVector.input);
  assert.equal(syntheticVector.input.fixture_profile_id, syntheticProfile.fixture_profile_id);
  const syntheticGo = renderPerformanceFixtureProfilesGo([profile, syntheticProfile]);
  assert.equal(renderPerformanceFixtureProfilesGo([syntheticProfile, profile]), syntheticGo, "descriptor output must not depend on registration order");
  assert.equal(renderPerformanceFixtureKeyVectors([syntheticProfile, profile]), renderPerformanceFixtureKeyVectors([profile, syntheticProfile]), "key-vector output must not depend on registration order");
  assert.match(syntheticGo, /synthetic_grid_snapshot_v1/u);
  assert.doesNotMatch(syntheticGo, /profile\.FixtureProfileID == "timeline_large_grid_snapshot_v2"/u);
  assert.doesNotMatch(syntheticGo, /contracts\/|RuntimeProfileID|ResourceProfileID|runtime_profile_id|resource_profile_id/u, "generated Go descriptors must not contain source paths or runtime routing identities");
  const row = context.catalog.rows.find((entry) => entry.fixture_profile_id === profile.fixture_profile_id);
  assert.ok(row);
  const syntheticVerificationID = "module.synthetic.verification.fixture_profile";
  const syntheticRegistry = {
    profiles: new Map([
      [profile.fixture_profile_id, profile],
      [syntheticProfile.fixture_profile_id, syntheticProfile],
    ]),
    verificationBindings: new Map([
      ...fixtureProfiles.verificationBindings,
      [syntheticVerificationID, {
          fixture_profile_id: syntheticProfile.fixture_profile_id,
          predicate_id: "perf.synthetic.fixture_profile.v1",
        }],
    ]),
  };
  const grouped = groupRowsByPerformanceFixture(root, [
    row,
    {
      ...row,
      fixture_profile_id: syntheticProfile.fixture_profile_id,
      row_id: "module.synthetic.measurement.fixture_profile",
      verification_ids: [syntheticVerificationID],
    },
  ], { registry: syntheticRegistry });
  assert.deepEqual(grouped.map((entry) => entry.fixture_profile_id), [profile.fixture_profile_id, syntheticProfile.fixture_profile_id].sort(compareASCII));
  const syntheticGroup = grouped.find((entry) => entry.fixture_profile_id === syntheticProfile.fixture_profile_id);
  assert.ok(syntheticGroup);
  assert.deepEqual(syntheticGroup.predicate_ids, ["perf.synthetic.fixture_profile.v1"]);
  assert.deepEqual(syntheticGroup.row_ids, ["module.synthetic.measurement.fixture_profile"]);
  const sameProfile = groupRowsByPerformanceFixture(root, [row, { ...row, row_id: "module.synthetic.measurement.same_profile" }], { registry: syntheticRegistry });
  assert.equal(sameProfile.length, 1);
  assert.deepEqual(sameProfile[0].row_ids, [row.row_id, "module.synthetic.measurement.same_profile"].sort(compareASCII));
  assert.deepEqual(groupRowsByPerformanceFixture(root, [{ ...row, fixture_profile_id: undefined, verification_ids: [] }], { registry: syntheticRegistry }), []);
  validateFixtureProfile({ row, fixtureProfiles, label: "valid_row" });
  assert.throws(() => validateFixtureProfile({
    row: { ...row, fixture_profile_id: undefined },
    fixtureProfiles,
    label: "missing_profile",
  }), /fixture_profile_id is required/u);
  assert.throws(() => validateFixtureProfile({
    row: { ...row, fixture_profile_id: "unknown_fixture_profile_v1" },
    fixtureProfiles,
    label: "unknown_profile",
  }), /unknown or inactive/u);
  assert.throws(() => validateFixtureProfile({
    row: { ...row, runtime_profile_id: "none" },
    fixtureProfiles,
    label: "incompatible_profile",
  }), /incompatible/u);
  assert.throws(() => validateSchemaSync("cartulary.performance_fixture_snapshot_key.v2", { ...envelope, schema_id: "cartulary.performance_fixture_snapshot_key.v1" }), /schema_id/u);
  assert.throws(() => validateSchemaSync("cartulary.performance_fixture_snapshot_key.v2", { ...envelope, migration_digest: "sha256:" + migrationDigest }), /migration_digest/u);
  const original = context.readJSON("tools/performance_fixture_snapshot_owner.json");
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-fixture-registry."));
  const registryFile = path.join(fixtureRoot, "registry.json");
  const expectRegistryFailure = (mutate, pattern) => {
    const candidate = structuredClone(original);
    const selected = candidate.profiles.find((entry) => entry.fixture_profile_id === profile.fixture_profile_id);
    assert.ok(selected);
    mutate(selected);
    writeFileSync(registryFile, JSON.stringify(candidate));
    assert.throws(() => loadPerformanceFixtureSnapshotRegistry(root, { registryPath: registryFile }), pattern);
  };
  try {
    expectRegistryFailure((candidate) => candidate.verification_bindings.push(structuredClone(candidate.verification_bindings[0])), /duplicate-free/u);
    expectRegistryFailure((candidate) => candidate.contributions.reverse(), /must precede/u);
    expectRegistryFailure((candidate) => candidate.contributions[0].dependencies.push("unknown.owner.v1"), /unknown dependency/u);
    expectRegistryFailure((candidate) => { candidate.status = "retired"; }, /status/u);
    expectRegistryFailure((candidate) => candidate.contributions[0].expected_receipt_counts.reverse(), /ASCII-sorted/u);
    expectRegistryFailure((candidate) => candidate.semantic_expectations.counts.reverse(), /ASCII-sorted/u);
    expectRegistryFailure((candidate) => candidate.runtime_credential_sets.push(structuredClone(candidate.runtime_credential_sets[0])), /duplicate-free/u);
    expectRegistryFailure((candidate) => {
      candidate.artifact_policy.summary_schema_id =
        "cartulary.frontend_measurement_summary.v99";
    }, /unsupported or mixed evidence generation/u);
    expectRegistryFailure((candidate) => {
      candidate.source_contract_refs[0].contract_id =
        candidate.source_contract_refs[0].contract_id.replace(/\.v[1-9][0-9]*$/u, ".v99");
    }, /contract identity does not match/u);
    const inactive = structuredClone(original);
    inactive.profiles.find((entry) => entry.fixture_profile_id === profile.fixture_profile_id).status = "inactive";
    writeFileSync(registryFile, JSON.stringify(inactive));
    const inactiveProfiles = loadPerformanceFixtureSnapshotRegistry(root, {
      registryPath: registryFile,
    });
    assert.throws(() => validateFixtureProfile({ row, fixtureProfiles: inactiveProfiles, label: "inactive_profile" }), /unknown or inactive/u);
    assert.throws(() => snapshotKey({
      ...profile,
      artifact_policy: {
        ...profile.artifact_policy,
        snapshot_key_schema_id: "cartulary.performance_fixture_snapshot_key.v99",
      },
    }, migrationDigest), /unsupported/u);
  }
  finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}
function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}
function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}
export function registerBoundariesTests() {
  test("performance_evidence: performance fixture snapshots and evidence generations remain isolated", async () => {
    const context = createContractTestContext(root);
    assertPerformanceFixtureSnapshotContract(context);
    assertPerformanceEvidenceGenerationBoundary();
  });
}
