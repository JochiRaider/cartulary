import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import {
  chmodSync,
  existsSync,
  linkSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  renameSync,
  rmSync,
  symlinkSync,
  utimesSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";
import { pathToFileURL } from "node:url";

import { planGoLPTShards } from "../backend/go-lpt-shards.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import {
  loadHistoricalPerformanceSchemaRegistry,
  validateHistoricalPerformanceEvidence,
} from "../diagnostics/historical-performance-evidence.mjs";
import {
  canonicalSnapshotKeyInput,
  groupRowsByPerformanceFixture,
  loadPerformanceFixtureBuilderPolicy,
  loadPerformanceFixtureSnapshotRegistry,
  postgresMigrationDigest,
  snapshotKey,
  snapshotKeyEnvelope,
} from "../performance-fixture/index.mjs";
import {
  renderPerformanceFixtureKeyVectors,
  renderPerformanceFixtureProfilesGo,
} from "../generated-artifacts/performance/performance-contracts.mjs";
import {
  buildWorkGraph,
  captureCapabilitySnapshot,
  cpuCapacityWithSafetyMargin,
  loadCacheRegistry,
  assertScannerEvidenceParity,
  cacheInputRootDigest,
  resolveVulnerabilityDatabaseRevision,
  runWorkGraph,
  simulateWorkGraph,
  validateWorkGraph,
  WorkGraphCache,
  WorkGraphCompiler,
  resolveCacheDependencyClosure,
} from "../scheduler/work-graph/index.mjs";
import { buildSourceSnapshot } from "../test-catalog/source-snapshot.mjs";
import { loadTestCatalog, validateFixtureProfile } from "../test-catalog/index.mjs";
import { validatePostgresFixturePolicy } from "../test-catalog/postgres-fixture-policy.mjs";
import { resolveRowSelector } from "../test-catalog/selector-resolution.mjs";
import { startManagedSuite } from "../scheduler/fixture-broker/providers.mjs";
import {
  cleanupStaleSuiteRuntimeRoots,
  createSuiteRuntime,
  scanRetainedRoot,
} from "../runtime/suite-runtime.mjs";
const root = path.resolve(import.meta.dirname, "../../..");

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}

const taskSurface = readJSON("tools/task_surface_owner.json");
const topology = readJSON("tools/execution_topology_manifest.json");
const rowMigrations = readJSON("tools/test_catalog_row_migrations.json");
const catalog = loadTestCatalog(root);
const compiler = new WorkGraphCompiler(root);
const cacheRegistry = loadCacheRegistry(root).registry;
const fixtureBuilderPolicy = loadPerformanceFixtureBuilderPolicy(root);
assert.deepEqual(
  [...fixtureBuilderPolicy.byFixtureProfileID.keys()],
  ["ac043_large_grid_snapshot_v1"],
);
assert.deepEqual(catalog.postgresFixturePolicy.counts, {
  postgres_dedicated: 256,
  postgres_migration: 8,
  postgres_transaction: 81,
});

function assertPostgresFixturePolicyClosesSharedRows() {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-postgres-policy."));
  const policyPath = path.join(fixtureRoot, "policy.json");
  try {
    const policy = readJSON("tools/postgres_fixture_policy_registry.json");
    policy.transaction_row_approvals.pop();
    writeFileSync(policyPath, JSON.stringify(policy));
    assert.throws(
      () => validatePostgresFixturePolicy(root, catalog.rows, { policyPath }),
      /must exactly cover current transaction rows/u,
    );
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

assertPostgresFixturePolicyClosesSharedRows();
const retiredTargets = [
  "check-service-backed",
  "release-browser-readiness",
  "test-fast-service-backed",
  "test-local",
  "test-service-backed",
];
const tiers = ["fast", "standard", "full", "release"];

function assertPerformanceEvidenceGenerationBoundary() {
  const historical = loadHistoricalPerformanceSchemaRegistry(root);
  const historicalIDs = new Set(historical.schemas.keys());
  const attachments = readJSON("tools/harness_schema_attachments.json");
  const activeIDs = new Set(attachments.attachments.map((entry) => entry.schema_id));
  for (const schemaID of historicalIDs) {
    assert.equal(activeIDs.has(schemaID), false);
    assert.equal(
      existsSync(path.join(root, "tools/schemas", `${schemaID}.schema.json`)),
      false,
    );
  }
  for (const schemaID of [
    "cartulary.browser_target_result.v3",
    "cartulary.frontend_measurement_aggregate.v3",
    "cartulary.frontend_measurement_observation.v2",
    "cartulary.frontend_measurement_summary.v3",
    "cartulary.performance_fixture_build_diagnostics.v2",
    "cartulary.performance_fixture_runtime.v2",
    "cartulary.performance_fixture_snapshot.v2",
    "cartulary.performance_fixture_snapshot_key.v2",
    "cartulary.performance_fixture_snapshot_lease.v2",
    "cartulary.performance_fixture_snapshot_owner.v2",
    "cartulary.web_e2e_stack_lease.v2",
  ]) {
    assert.equal(activeIDs.has(schemaID), true);
    assert.equal(historicalIDs.has(schemaID), false);
  }
  const historicalKey = {
    schema_id: "cartulary.performance_fixture_snapshot_key.v1",
    migration_digest: "e".repeat(64),
    source_contract_digest: "f".repeat(64),
    fixture_version: "cartulary.perf.large_grid.v1",
    seed: 20260405,
  };
  validateHistoricalPerformanceEvidence(
    root,
    historicalKey.schema_id,
    historicalKey,
  );
  assert.throws(
    () => validateSchemaSync(historicalKey.schema_id, historicalKey),
    /missing schema attachment/u,
  );
  const reference = (role, name) => ({
    role,
    path_kind: "file",
    format: "json",
    path: `browser-e2e-measurement/${name}.json`,
    sha256: `sha256:${"a".repeat(64)}`,
  });
  const aggregate = {
    schema_id: "cartulary.frontend_measurement_aggregate.v3",
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
        p95_ms: 20,
        threshold_ms: 100,
        outcome: "passed",
      }],
    }],
  };
  validateSchemaSync(aggregate.schema_id, aggregate);
  assert.throws(
    () => validateSchemaSync(aggregate.schema_id, {
      ...aggregate,
      summaries: [{ samples: Array.from({ length: 100_000 }, () => 1) }],
    }),
    /additional properties/u,
    "aggregate v3 must grow by immutable summary references, not raw samples",
  );
}

function assertPerformanceFixtureSnapshotContract() {
  const fixtureProfiles = catalog.fixtureProfiles;
  const profile = fixtureProfiles.profiles.get("ac043_large_grid_snapshot_v1");
  assert.ok(profile);
  assert.match(profile.source_contract_digest, /^[a-f0-9]{64}$/u);
  const migrationDigest = postgresMigrationDigest(root);
  const envelope = snapshotKeyEnvelope(profile, migrationDigest);
  validateSchemaSync(envelope.schema_id, envelope);
  assert.match(snapshotKey(profile, migrationDigest), /^[a-f0-9]{64}$/u);
  assert.notEqual(
    snapshotKey(profile, migrationDigest),
    snapshotKey(profile, "f".repeat(64)),
  );
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
    assert.equal(
      canonicalSnapshotKeyInput(vectorProfile, vector.input.migration_digest),
      vector.canonical_json,
    );
    assert.equal(
      snapshotKey(vectorProfile, vector.input.migration_digest),
      vector.snapshot_key,
    );
  }

  const syntheticProfile = structuredClone(profile);
  syntheticProfile.fixture_profile_id = "synthetic_grid_snapshot_v1";
  syntheticProfile.fixture_version = "cartulary.perf.synthetic_grid.v1";
  syntheticProfile.seed += 1;
  syntheticProfile.source_contract_digest = "a".repeat(64);
  syntheticProfile.artifact_policy.snapshot_key_schema_id =
    "cartulary.performance_fixture_snapshot_key.v2";
  const syntheticVectors = JSON.parse(
    renderPerformanceFixtureKeyVectors([profile, syntheticProfile]),
  );
  assert.deepEqual(
    syntheticVectors.vectors.map((entry) => entry.name),
    ["ac043_large_grid_snapshot_v1", "synthetic_grid_snapshot_v1"],
  );
  validateSchemaSync(
    "cartulary.performance_fixture_snapshot_key.v2",
    syntheticVectors.vectors[1].input,
  );
  assert.equal(
    syntheticVectors.vectors[1].input.fixture_profile_id,
    syntheticProfile.fixture_profile_id,
  );
  const syntheticGo = renderPerformanceFixtureProfilesGo([profile, syntheticProfile]);
  assert.match(syntheticGo, /synthetic_grid_snapshot_v1/u);
  assert.doesNotMatch(
    syntheticGo,
    /profile\.FixtureProfileID == "ac043_large_grid_snapshot_v1"/u,
  );
  assert.doesNotMatch(
    syntheticGo,
    /contracts\/|RuntimeProfileID|ResourceProfileID|runtime_profile_id|resource_profile_id/u,
    "generated Go descriptors must not contain source paths or runtime routing identities",
  );

  const row = catalog.rows.find((entry) => entry.fixture_profile_id);
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
  assert.deepEqual(
    grouped.map((entry) => entry.fixture_profile_id),
    ["ac043_large_grid_snapshot_v1", "synthetic_grid_snapshot_v1"],
  );
  assert.deepEqual(grouped[1].predicate_ids, ["perf.synthetic.fixture_profile.v1"]);
  assert.deepEqual(grouped[1].row_ids, ["module.synthetic.measurement.fixture_profile"]);
  validateFixtureProfile({ row, fixtureProfiles, label: "valid_row" });
  assert.throws(
    () => validateFixtureProfile({
      row: { ...row, fixture_profile_id: undefined },
      fixtureProfiles,
      label: "missing_profile",
    }),
    /fixture_profile_id is required/u,
  );
  assert.throws(
    () => validateFixtureProfile({
      row: { ...row, fixture_profile_id: "unknown_fixture_profile_v1" },
      fixtureProfiles,
      label: "unknown_profile",
    }),
    /unknown or inactive/u,
  );
  assert.throws(
    () => validateFixtureProfile({
      row: { ...row, runtime_profile_id: "none" },
      fixtureProfiles,
      label: "incompatible_profile",
    }),
    /incompatible/u,
  );
  assert.throws(
    () => validateSchemaSync(
      "cartulary.performance_fixture_snapshot_key.v2",
      { ...envelope, schema_id: "cartulary.performance_fixture_snapshot_key.v1" },
    ),
    /schema_id/u,
  );
  assert.throws(
    () => validateSchemaSync(
      "cartulary.performance_fixture_snapshot_key.v2",
      { ...envelope, migration_digest: "sha256:" + migrationDigest },
    ),
    /migration_digest/u,
  );

  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-fixture-registry."));
  const registryFile = path.join(fixtureRoot, "registry.json");
  const original = readJSON("tools/performance_fixture_snapshot_owner.json");
  const expectRegistryFailure = (mutate, pattern) => {
    const candidate = structuredClone(original);
    mutate(candidate.profiles[0]);
    writeFileSync(registryFile, JSON.stringify(candidate));
    assert.throws(
      () => loadPerformanceFixtureSnapshotRegistry(root, { registryPath: registryFile }),
      pattern,
    );
  };
  try {
    expectRegistryFailure(
      (candidate) => candidate.verification_bindings.push(
        structuredClone(candidate.verification_bindings[0]),
      ),
      /duplicate-free/u,
    );
    expectRegistryFailure(
      (candidate) => candidate.contributions.reverse(),
      /must precede/u,
    );
    expectRegistryFailure(
      (candidate) => candidate.contributions[0].dependencies.push("unknown.owner.v1"),
      /unknown dependency/u,
    );
    expectRegistryFailure(
      (candidate) => { candidate.status = "retired"; },
      /status/u,
    );
    expectRegistryFailure(
      (candidate) => candidate.contributions[0].expected_receipt_counts.reverse(),
      /ASCII-sorted/u,
    );
    expectRegistryFailure(
      (candidate) => candidate.semantic_expectations.counts.reverse(),
      /ASCII-sorted/u,
    );
    expectRegistryFailure(
      (candidate) => candidate.runtime_credential_sets.push(
        structuredClone(candidate.runtime_credential_sets[0]),
      ),
      /duplicate-free/u,
    );
    expectRegistryFailure(
      (candidate) => {
        candidate.artifact_policy.summary_schema_id =
          "cartulary.frontend_measurement_summary.v99";
      },
      /unsupported or mixed evidence generation/u,
    );
    expectRegistryFailure(
      (candidate) => {
        candidate.source_contract_refs[0].contract_id =
          "cartulary.performance.route_divergence.v1";
      },
      /contract identity does not match/u,
    );

    const inactive = structuredClone(original);
    inactive.profiles[0].status = "inactive";
    writeFileSync(registryFile, JSON.stringify(inactive));
    const inactiveProfiles = loadPerformanceFixtureSnapshotRegistry(root, {
      registryPath: registryFile,
    });
    assert.throws(
      () => validateFixtureProfile({ row, fixtureProfiles: inactiveProfiles, label: "inactive_profile" }),
      /unknown or inactive/u,
    );
    assert.throws(
      () => snapshotKey({
        ...profile,
        artifact_policy: {
          ...profile.artifact_policy,
          snapshot_key_schema_id: "cartulary.performance_fixture_snapshot_key.v99",
        },
      }, migrationDigest),
      /unsupported/u,
    );
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

const generalCases = [
  {
    id: "current_owner_and_topology_identity",
    name: "current task-surface and topology identities are exact",
    acceptance_ids: ["TH-HARNESS-AC-082"],
    run() {
      assert.equal(taskSurface.schema_id, "cartulary.task_surface_owner.v2");
      assert.equal(topology.schema_id, "cartulary.execution_topology.v7");
    },
  },
  {
    id: "current_target_roster",
    name: "current target roster and public surface are exact",
    acceptance_ids: ["TH-HARNESS-AC-001"],
    run() {
      assert.equal(taskSurface.targets.length, 149);
      assert.equal(taskSurface.targets.filter((entry) => entry.target_class === "public").length, 101);
    },
  },
  {
    id: "catalog_tier_closure",
    name: "catalog rows use only current minimum tiers",
    acceptance_ids: ["TH-HARNESS-AC-018"],
    run() {
      assert.ok(catalog.rows.length > 0);
      assert.ok(catalog.rows.every((row) => tiers.includes(row.minimum_tier)));
    },
  },
  {
    id: "fixture_profile_closure",
    name: "fixture profiles, retired rows, and canonical construction are closed",
    acceptance_ids: ["TH-HARNESS-AC-086", "TH-HARNESS-AC-096", "TH-HARNESS-AC-098"],
    run() {
      assert.ok(catalog.rows.every((row) => typeof row.fixture_capability === "string"));
      const profiledRows = catalog.rows.filter((row) => row.fixture_profile_id);
      assert.equal(profiledRows.length, 4);
      assert.ok(profiledRows.every((row) =>
        row.fixture_profile_id === "ac043_large_grid_snapshot_v1"));
      assert.ok(catalog.rows.every((row) => !("default_check" in row)));
      for (let mask = 1; mask < (1 << profiledRows.length); mask += 1) {
        const rowIDs = profiledRows
          .filter((_, index) => (mask & (1 << index)) !== 0)
          .map((row) => row.row_id);
        const graph = compiler.compile({ kind: "rows", row_ids: rowIDs });
        assert.equal(
          graph.units.filter((unit) => unit.kind === "fixture_builder").length,
          1,
          `profiled subset ${rowIDs.join(",")} must have one canonical construction`,
        );
      }
      const nonMeasurement = compiler.compile({
        kind: "rows",
        row_ids: ["harness.browser.unit.source_owner_contribution_assembler"],
      });
      assert.equal(nonMeasurement.units.some((unit) => unit.kind === "fixture_builder"), false);
      const migration = rowMigrations.migrations.find((entry) =>
        entry.retired_row_id ===
          "harness.browser.integration.performance_fixture_source_owner_assembly");
      assert.deepEqual(migration?.replacement_row_ids, [
        "harness.browser.integration.performance_fixture_snapshot_lifecycle",
        "harness.browser.unit.source_owner_contribution_assembler",
      ]);
    },
  },
  {
    id: "retired_target_absence",
    name: "retired public target aliases remain absent",
    acceptance_ids: ["TH-HARNESS-AC-001"],
    run() {
      assert.ok(retiredTargets.every((target) =>
        !taskSurface.targets.some((entry) => entry.name === target)));
    },
  },
  {
    id: "retired_topology_field_absence",
    name: "retired schedule and fixture fields remain absent",
    acceptance_ids: ["TH-HARNESS-AC-082"],
    run() {
      assert.equal("sequence_schedules" in topology, false);
      assert.equal("check_schedules" in topology, false);
      assert.equal("service_backed_schedules" in topology, false);
      assert.equal("fixture_profiles" in topology, false);
    },
  },
  {
    id: "cache_registry_identity",
    name: "cache registry exposes the current test-row profile",
    acceptance_ids: ["TH-HARNESS-AC-091"],
    run() {
      assert.equal(cacheRegistry.schema_id, "cartulary.harness_cache_registry.v1");
      assert.ok(cacheRegistry.profiles.some((profile) => profile.profile_id === "test_rows"));
    },
  },
];

async function assertSuiteRuntimeBoundary() {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-suite-runtime-contract."));
  const repo = path.join(fixtureRoot, "repo");
  const runRoot = path.join(fixtureRoot, "results", "run-runtime-contract");
  const scratch = path.join(fixtureRoot, "scratch");
  mkdirSync(repo, { recursive: true, mode: 0o700 });
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  try {
    const descendantRoot = path.join(fixtureRoot, "descendant-output", "nested", "leaf");
    const privateProcessModule = pathToFileURL(
      path.join(root, "tools/harness/runtime/private-process.mjs"),
    ).href;
    execFileSync(process.execPath, [
      "--input-type=module",
      "--eval",
      `
        import { mkdirSync } from "node:fs";
        import { enforcePrivateProcessUmask } from ${JSON.stringify(privateProcessModule)};
        process.umask(0o022);
        enforcePrivateProcessUmask();
        mkdirSync(process.argv[1], { recursive: true });
      `,
      descendantRoot,
    ]);
    for (const directory of [
      path.join(fixtureRoot, "descendant-output"),
      path.join(fixtureRoot, "descendant-output", "nested"),
      descendantRoot,
    ]) {
      assert.equal(lstatSync(directory).mode & 0o777, 0o700);
    }

    const runtime = createSuiteRuntime({
      repoRoot: repo,
      runRoot,
      runID: "run-runtime-contract",
      scratchRoot: scratch,
    });
    assert.equal(lstatSync(runtime.root).mode & 0o777, 0o700);
    runtime.registerSecret("injected-secret-value");
    writeFileSync(path.join(runRoot, "safe.json"), '{"status":"pass"}\n', { mode: 0o600 });
    assert.equal((await scanRetainedRoot(runRoot)).status, "pass");

    const forbiddenFile = path.join(runRoot, "suite-environment.json");
    writeFileSync(forbiddenFile, "{}\n", { mode: 0o600 });
    await assert.rejects(
      () => scanRetainedRoot(runRoot, { removeUnsafe: true }),
      /forbidden runtime filename/u,
    );
    assert.equal(existsSync(forbiddenFile), false);

    const secretFile = path.join(runRoot, "diagnostic.log");
    writeFileSync(secretFile, "injected-secret-value\n", { mode: 0o600 });
    await assert.rejects(
      () => scanRetainedRoot(runRoot, {
        forbiddenValues: runtime.forbiddenValues(),
        removeUnsafe: true,
      }),
      /registered runtime secret/u,
    );
    assert.equal(existsSync(secretFile), false);

    const link = path.join(runRoot, "unsafe-link");
    symlinkSync(path.join(runRoot, "safe.json"), link);
    await assert.rejects(
      () => scanRetainedRoot(runRoot, { removeUnsafe: true }),
      /symlink/u,
    );
    assert.equal(existsSync(link), false);

    const permissive = path.join(runRoot, "permissive.log");
    writeFileSync(permissive, "bounded diagnostic\n", { mode: 0o600 });
    chmodSync(permissive, 0o644);
    await assert.rejects(
      () => scanRetainedRoot(runRoot, { removeUnsafe: true }),
      /not owner-only/u,
    );
    assert.equal(lstatSync(permissive).mode & 0o777, 0o600);
    rmSync(permissive);

    const runtimeBase = path.join(scratch, "suite-runtime");
    const inFlight = mkdtempSync(path.join(runtimeBase, ".creating-suite-race-contract-"));
    chmodSync(inFlight, 0o700);
    const concurrent = createSuiteRuntime({
      repoRoot: repo,
      runRoot,
      runID: "concurrent-runtime-contract",
      scratchRoot: scratch,
    });
    concurrent.close();
    assert.equal(existsSync(inFlight), true);
    const old = new Date("2000-01-01T00:00:00.000Z");
    utimesSync(inFlight, old, old);
    const stagedCleanup = cleanupStaleSuiteRuntimeRoots({
      repoRoot: repo,
      runRoot,
      scratchRoot: scratch,
    });
    assert.equal(stagedCleanup.removed, 1);
    assert.equal(existsSync(inFlight), false);

    const published = runtime.root;
    const disappearing = path.join(runtimeBase, ".creating-suite-publish-contract");
    renameSync(published, disappearing);
    const concurrentPublication = cleanupStaleSuiteRuntimeRoots({
      repoRoot: repo,
      runRoot,
      scratchRoot: scratch,
      beforeCandidateInspection(candidate) {
        if (candidate === disappearing) renameSync(disappearing, published);
      },
    });
    assert.ok(concurrentPublication.scanned >= 1);
    assert.equal(concurrentPublication.removed, 0);
    assert.equal(existsSync(disappearing), false);
    assert.equal(existsSync(published), true);
    runtime.close();
    assert.equal(existsSync(runtime.root), false);

    const stale = createSuiteRuntime({
      repoRoot: repo,
      runRoot,
      runID: "stale-runtime-contract",
      scratchRoot: scratch,
    });
    const ownerPath = path.join(stale.root, "runtime-owner.json");
    const owner = JSON.parse(readFileSync(ownerPath, "utf8"));
    owner.created_at = "2000-01-01T00:00:00.000Z";
    writeFileSync(ownerPath, `${JSON.stringify(owner)}\n`, { mode: 0o600 });
    chmodSync(ownerPath, 0o600);
    const cleanup = cleanupStaleSuiteRuntimeRoots({
      repoRoot: repo,
      runRoot,
      scratchRoot: scratch,
    });
    assert.equal(cleanup.removed, 1);
    assert.equal(existsSync(stale.root), false);

    assert.throws(
      () => createSuiteRuntime({
        repoRoot: repo,
        runRoot,
        runID: "contained-runtime-contract",
        scratchRoot: path.join(repo, "scratch"),
      }),
      /outside repository/u,
    );
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

async function assertBoundaryContract(kind) {
  switch (kind) {
    case "performance_and_runtime":
      assertPerformanceFixtureSnapshotContract();
      assertPerformanceEvidenceGenerationBoundary();
      await assertSuiteRuntimeBoundary();
      return;
    case "attachment_uniqueness": {
      const attachments = readJSON("tools/harness_schema_attachments.json");
      const ids = attachments.attachments.map((entry) => entry.schema_id);
      assert.equal(new Set(ids).size, ids.length);
      return;
    }
    case "current_schema_presence":
      for (const schema of [
        "cartulary.harness_work_graph.v5.schema.json",
        "cartulary.harness_run_manifest.v1.schema.json",
        "cartulary.harness_unit_event.v2.schema.json",
        "cartulary.harness_run_summary.v1.schema.json",
      ]) assert.ok(existsSync(path.join(root, "tools/schemas", schema)));
      return;
    case "retired_schema_absence":
      for (const oldSchema of [
        "cartulary.harness_execution_context.v2.schema.json",
        "cartulary.harness_invocation_start.v1.schema.json",
        "cartulary.harness_observability_index.v1.schema.json",
        "cartulary.harness_sequence_event.v1.schema.json",
        "cartulary.task_surface_owner.v1.schema.json",
        "cartulary.test_family_manifest.v2.schema.json",
        "cartulary.test_family_manifest.v4.schema.json",
        "cartulary.scheduler_manifest.v2.schema.json",
      ]) assert.equal(existsSync(path.join(root, "tools/schemas", oldSchema)), false);
      return;
    case "work_graph_owner_validation":
      validateSchemaSync(
        "cartulary.harness_work_graph_owner.v2",
        readJSON("tools/harness_work_graph_owner.json"),
      );
      return;
    case "relative_backing_scripts":
      assert.ok(taskSurface.targets.every((entry) =>
        (entry.backing_scripts ?? []).every((file) => !path.isAbsolute(file))));
      return;
    default:
      throw new Error(`unknown boundary contract ${kind}`);
  }
}

function assertCommandSurfaceContract(kind) {
  const graphTargets = ["test-slice", "service-backed-test-slice", "test-fast", "test", "check", "ci", "release-check", "lint"];
  switch (kind) {
    case "graph_entrypoints":
      assert.ok(graphTargets.every((target) => taskSurface.make_recipes[target]?.type === "work_graph"));
      assert.ok(taskSurface.observability_policy.required_targets.every((target) => {
        const recipe = taskSurface.make_recipes[target];
        return recipe?.type === "work_graph" || recipe?.graph_entry === true;
      }));
      return;
    case "command_identities":
      assert.ok(taskSurface.targets.filter((entry) => entry.target_class === "public").every((entry) =>
        /^cartulary\.harness\.command\.[a-z0-9_]+\.v[1-9][0-9]*$/u.test(entry.command_id)));
      return;
    case "graph_inputs":
      for (const target of taskSurface.observability_policy.required_targets) {
        const inputNames = (taskSurface.targets.find((entry) => entry.name === target)?.input_contract?.inputs ?? []).map((input) => input.name);
        assert.ok(inputNames.includes("CARTULARY_HARNESS_CACHE_MODE"));
        assert.ok(inputNames.includes("CARTULARY_HARNESS_CAPACITY_OVERRIDE"));
      }
      return;
    case "output_contracts":
      assert.ok(!JSON.stringify(taskSurface.make_recipes).includes("nested_scheduler"));
      assert.ok(taskSurface.observability_policy.required_targets.every((target) => {
        const entry = taskSurface.targets.find((candidate) => candidate.name === target);
        return entry.output_policy.summary_schema === "cartulary.harness_run_summary.v1" &&
          entry.output_policy.artifact_policy === "run_and_target_summaries";
      }));
      return;
    default:
      throw new Error(`unknown command-surface contract ${kind}`);
  }
}

function assertGoSelectorBuildContextContract() {
  const fixtureRoot = mkdtempSync(path.join(root, "internal/selector-resolution-contract."));
  const relativePackage = `./${path.relative(root, fixtureRoot).replaceAll("\\", "/")}`;
  const runner = { approved_roots: ["internal"] };
  const row = (testName) => ({
    row_id: "harness.test_catalog.behavior.fixture",
    runner: "go",
    selector: { package: relativePackage, tests: [testName] },
  });
  try {
    writeFileSync(path.join(fixtureRoot, "ordinary_test.go"), "package fixture\nfunc TestOrdinary() {}\n");
    writeFileSync(
      path.join(fixtureRoot, "applicable_linux_amd64_test.go"),
      "//go:build linux && amd64\n\npackage fixture\nfunc TestApplicablePlatform() {}\n",
    );
    writeFileSync(
      path.join(fixtureRoot, "private_test.go"),
      "//go:build cartulary_harness\n\npackage fixture\nfunc TestPrivateProfile() {}\n",
    );
    writeFileSync(
      path.join(fixtureRoot, "excluded_windows_test.go"),
      "//go:build windows\n\npackage fixture\nfunc TestExcludedPlatform() {}\n",
    );
    assert.deepEqual(
      resolveRowSelector({ root, row: row("TestOrdinary"), runner, taskSurfaceCommandIDs: new Set() }),
      [`go:${relativePackage}:TestOrdinary`],
    );
    assert.deepEqual(
      resolveRowSelector({ root, row: row("TestApplicablePlatform"), runner, taskSurfaceCommandIDs: new Set() }),
      [`go:${relativePackage}:TestApplicablePlatform`],
    );
    assert.throws(
      () => resolveRowSelector({ root, row: row("TestPrivateProfile"), runner, taskSurfaceCommandIDs: new Set() }),
      /TestPrivateProfile is excluded from the Go runner build context linux\/amd64/u,
    );
    assert.throws(
      () => resolveRowSelector({ root, row: row("TestExcludedPlatform"), runner, taskSurfaceCommandIDs: new Set() }),
      /TestExcludedPlatform is excluded from the Go runner build context linux\/amd64/u,
    );
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

function assertEvidenceContract(kind) {
  const tierCounts = Object.fromEntries(tiers.map((tier) => [tier, catalog.rows.filter((row) => row.minimum_tier === tier).length]));
  switch (kind) {
    case "selector_build_context":
      assertGoSelectorBuildContextContract();
      return;
    case "tier_partition":
      assert.deepEqual(Object.keys(tierCounts), tiers);
      assert.equal(Object.values(tierCounts).reduce((sum, count) => sum + count, 0), catalog.rows.length);
      return;
    case "tier_monotonicity": {
      const reached = tiers.map((tier, rank) =>
        catalog.rows.filter((row) => tiers.indexOf(row.minimum_tier) <= rank).length);
      assert.ok(reached.every((count, rank) => rank === 0 || count >= reached[rank - 1]));
      assert.equal(reached.at(-1), catalog.rows.length);
      return;
    }
    case "fixture_partition": {
      const fixtureCounts = new Map();
      for (const row of catalog.rows) {
        fixtureCounts.set(row.fixture_capability, (fixtureCounts.get(row.fixture_capability) ?? 0) + 1);
      }
      assert.equal([...fixtureCounts.values()].reduce((sum, count) => sum + count, 0), catalog.rows.length);
      assert.equal(
        fixtureCounts.get("browser_stack"),
        catalog.rows.filter((row) => row.runner === "playwright").length,
      );
      return;
    }
    case "active_owner_coverage":
      assert.ok(catalog.registry.owners.filter((owner) => owner.status === "active").every((owner) =>
        catalog.rows.some((row) => row.owner_id === owner.owner_id)));
      return;
    default:
      throw new Error(`unknown evidence contract ${kind}`);
  }
}

function assertGraphContract(kind) {
  const roots = ["test-fast", "check", "test", "ci", "release-check"];
  switch (kind) {
    case "aggregate_determinism":
      for (const target of roots) {
        const graph = compiler.compile({ kind: "aggregate", target });
        validateWorkGraph(graph);
        assert.equal(
          graph.graph_digest,
          compiler.compile({ kind: "aggregate", target }).graph_digest,
        );
      }
      return;
    case "row_evidence_outputs": {
      const row = catalog.rows.find((entry) => entry.runner === "go" && entry.fixture_capability === "none");
      const graph = compiler.compile({ kind: "rows", row_ids: [row.row_id] });
      assert.deepEqual(
        graph.units.flatMap((unit) => unit.current_run_evidence_outputs),
        [`rows/${row.row_id}.json`, `unit-results/go-${graph.units[0].unit_id.split(":").slice(1).join("-")}.json`],
      );
      return;
    }
    case "go_lpt_budget": {
      const plan = planGoLPTShards(
        Array.from({ length: 17 }, (_, itemIndex) => ({ id: `row-${itemIndex}`, estimated_work_ms: itemIndex + 1, compatibility: { runtime: "none" } })),
        { availableGoLanes: 4 },
      );
      assert.equal(plan.shards.flatMap((shard) => shard.item_ids).length, 17);
      assert.equal(plan.shards.length, 1);
      assert.equal(plan.worker_count, 1);
      assert.equal(plan.gomaxprocs, 4);
      const hostPlan = planGoLPTShards(
        Array.from({ length: 10 }, (_, itemIndex) => ({
          id: `group-${itemIndex}`,
          estimated_work_ms: itemIndex + 1,
          compatibility: { package: `package-${itemIndex}` },
        })),
        { availableGoLanes: 24 },
      );
      assert.equal(hostPlan.worker_count, 6);
      assert.equal(hostPlan.gomaxprocs, 4);
      assert.ok(hostPlan.shards.every((shard) => shard.cpu_tokens === 4));
      const isolatedPlan = planGoLPTShards(
        Array.from({ length: 3 }, (_, itemIndex) => ({
          id: `isolated-${itemIndex}`,
          estimated_work_ms: itemIndex + 1,
          compatibility: { package: "same-package" },
          isolated: true,
        })),
        { availableGoLanes: 4 },
      );
      assert.equal(isolatedPlan.shards.length, 3);
      assert.ok(isolatedPlan.shards.every((shard) => shard.isolated));
      return;
    }
    case "target_graph_validation": {
      const graph = compiler.compile({ kind: "target", target: "lint" });
      validateWorkGraph(graph);
      assert.ok(graph.units.length >= 7);
      return;
    }
    default:
      throw new Error(`unknown graph contract ${kind}`);
  }
}

function schedulerFixture() {
  const unit = (unitID, needs, claims) => ({
    unit_id: unitID,
    owner_id: "harness.scheduler",
    kind: "runner",
    command: { executable: "true", args: [], environment: {} },
    needs,
    resource_claims: claims,
    fixture_lease: "none",
    service_dependencies: [],
    cache_policy: "none",
    timeout_ms: 1000,
    current_run_evidence_outputs: [`rows/${unitID}.json`],
    failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" },
    estimated_work_ms: 10,
  });
  return buildWorkGraph([
    unit("a", [], { cpu: 1, process: 1 }),
    unit("b", [], { cpu: 1, process: 1 }),
    unit("c", ["a"], { cpu: 1, process: 1 }),
  ]);
}

function assertManagedSuiteFailurePropagation() {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-managed-suite-failure."));
  const resultsRoot = path.join(fixtureRoot, "results");
  const runID = "managed-suite-failure";
  const runRoot = path.join(resultsRoot, runID);
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  const executable = path.join(fixtureRoot, "fake-testservices.cjs");
  writeFileSync(executable, `const fs = require("node:fs");
const path = require("node:path");
const value = (name) => process.argv[process.argv.indexOf(name) + 1];
const suiteID = "0123456789abcdef01234567";
const runID = process.env.CARTULARY_TEST_RUN_ID;
const target = process.env.CARTULARY_TEST_TARGET;
const scopeRef = "_shared/test-services/" + suiteID + "/service-scope.json";
const scopePath = path.join(process.env.CARTULARY_TEST_RESULTS_DIR, runID, scopeRef);
if (process.env.FAKE_START_MODE === "config") process.exit(2);
fs.mkdirSync(path.dirname(scopePath), { recursive: true, mode: 0o700 });
const startup = { attempt_count: 1, retry_count: 0, slowest_attempt_duration_ms: 5, final_attempt: 1, final_status: "pass", final_retryable: false, final_retry_blocked_by_context: false };
const failure = { failure_class: "infra", failure_reason: "service_readiness_timeout", service: "object_store", stage: "object-store-start", operation: "start suite object-store", message: "object-store readiness failed: stage=list attempts=29 cleanup=not_needed reason=deadline_expired", attempts_started: 1, max_attempts: 2, retryable: false, retry_blocked_by_context: false };
const scope = { schema_id: "cartulary.test_services.scope.v2", target, suite_id: suiteID, run_id: runID, artifact_dir: path.dirname(scopePath), readiness_generation: "sha256:" + "a".repeat(64), wrapper: { owned_count: 1, pass_through_count: 0 }, preflight: { docker_ok: true, reaper_ready: true, stale_containers_scanned: 0, stale_containers_removed: 0, stale_containers_deferred: 0, ryuk_disabled_for_suite_startup: true }, failure, failures: { counts: { infra: 1 }, reasons: { service_readiness_timeout: 1 }, exemplars: { infra: [failure] } }, cleanup: { status: "startup_failed", child_exit_status: 3 }, postgres: { started: false, startup, attached_harness_count: 0, created_database_count: 0, migrated_database_count: 0, template_clone_count: 0 }, object_store: { started: false, secure: false, startup, attached_harness_count: 0, bucket_create_count: 0, bucket_cleanup_count: 0 }, browser_e2e: { retired_fixture_count: 0, cleaned_fixture_count: 0, reclaimed_fixture_count: 0 }, fixture: { total_count: 0, total_duration_ms: 0, strategy_aggregate_count: 0 }, started_services: {}, extensions: { "cartulary.object_store_readiness": { schema_id: "cartulary.object_store_readiness_diagnostic.v1", phase: "initial_lane", stage: "list", outcome: "deadline_expired", attempt_count: 29, attempt_timeout_count: 1, elapsed_ms: 120000, cleanup_outcome: "not_needed", cause_counts: { transport_unreachable: 28, operation_timeout: 1 }, container_state: "running", health_state: "none", exit_code: 0, oom_killed: false } } };
fs.writeFileSync(scopePath, JSON.stringify(scope) + "\\n", { mode: 0o600 });
if (process.env.FAKE_START_MODE === "missing") process.exit(3);
const result = { schema_id: "cartulary.test_services.start_result.v1", status: "failed", run_id: runID, target, suite_id: suiteID, service_scope_ref: scopeRef, failure_class: "infra", failure_reason: "service_readiness_timeout" };
if (process.env.FAKE_START_MODE === "foreign") result.run_id = "foreign-run";
if (process.env.FAKE_START_MODE === "escape") result.service_scope_ref = "../foreign.json";
fs.writeFileSync(value("--result-file"), JSON.stringify(result) + "\\n", { mode: 0o600 });
process.exit(3);
`, { mode: 0o700 });
  const suiteRuntime = createSuiteRuntime({
    repoRoot: root,
    runRoot,
    runID,
    scratchRoot: path.join(fixtureRoot, "scratch"),
  });
  try {
    assert.throws(
      () => startManagedSuite({
        root,
        target: "test-slice",
        suiteRuntime,
        executable: process.execPath,
        executableArgs: ["--", executable],
        environment: {
          CARTULARY_TEST_RESULTS_DIR: resultsRoot,
          CARTULARY_TEST_RUN_ID: runID,
        },
      }),
      (error) => {
        assert.equal(error.failure_class, "infra", error.message);
        assert.equal(error.failure_reason, "service_readiness_timeout", error.message);
        assert.deepEqual(error.artifact_refs, [
          "_shared/test-services/0123456789abcdef01234567/service-scope.json",
        ]);
        return true;
      },
    );
    for (const mode of ["escape", "foreign", "missing"]) {
      assert.throws(
        () => startManagedSuite({
          root,
          target: "test-slice",
          suiteRuntime,
          executable: process.execPath,
          executableArgs: ["--", executable],
          environment: {
            CARTULARY_TEST_RESULTS_DIR: resultsRoot,
            CARTULARY_TEST_RUN_ID: runID,
            FAKE_START_MODE: mode,
          },
        }),
        (error) => {
          assert.equal(error.failure_class, "artifact");
          assert.equal(error.failure_reason, "artifact_error");
          return true;
        },
      );
    }
    assert.throws(
      () => startManagedSuite({
        root,
        target: "test-slice",
        suiteRuntime,
        executable: process.execPath,
        executableArgs: ["--", executable],
        environment: {
          CARTULARY_TEST_RESULTS_DIR: resultsRoot,
          CARTULARY_TEST_RUN_ID: runID,
          FAKE_START_MODE: "config",
        },
      }),
      (error) => {
        assert.equal(error.failure_class, "config");
        assert.equal(error.failure_reason, "configuration_error");
        return true;
      },
    );
  } finally {
    suiteRuntime.close();
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

function cacheFixture(policy = "content_addressed") {
  const fixtureRoot = mkdtempSync(path.join(root, "tmp/work-graph-cache-contract."));
  const runRoot = path.join(fixtureRoot, "run");
  const cacheRoot = path.join(fixtureRoot, "cache");
  const input = path.join(fixtureRoot, "input.txt");
  const output = path.join(runRoot, "artifacts/result.txt");
  mkdirSync(path.dirname(output), { recursive: true });
  writeFileSync(input, "input-one\n");
  writeFileSync(output, "output-one\n", { mode: 0o600 });
  const profile = {
    profile_id: "fixture_profile",
    policy,
    dependency_strategy: "broad",
    targets: ["fixture-target"],
    input_roots: [path.relative(root, input).replaceAll("\\", "/")],
    requires_vulnerability_database_revision: false,
  };
  const registry = {
    schema_id: "cartulary.harness_cache_registry.v1",
    profiles: [profile],
  };
  const unit = {
    ...schedulerFixture().units[0],
    unit_id: "target:fixture-target",
    command: {
      executable: "true",
      args: [],
      environment: { CARTULARY_TEST_TARGET: "fixture-target" },
    },
    cache_policy: policy,
    current_run_evidence_outputs: ["unit-results/target-fixture-target.json"],
    reusable_artifact_outputs: [
      {
        artifact_type: "file",
        relative_path: "artifacts/result.txt",
        destination_class: "run_root",
        mode: "0600",
        producer_identity: "target:fixture-target",
      },
    ],
    semantic_digest: `sha256:${"a".repeat(64)}`,
  };
  const create = (options = {}) => new WorkGraphCache({
    root,
    runRoot,
    cacheRoot,
    registry,
    toolchainDigest: options.toolchainDigest ?? `sha256:${"b".repeat(64)}`,
    helperDigest: options.helperDigest ?? `sha256:${"c".repeat(64)}`,
    mode: options.mode ?? "normal",
  });
  return { fixtureRoot, runRoot, cacheRoot, input, output, unit, create };
}

async function assertContentCacheContract() {
  const fixture = cacheFixture();
  try {
    const markdownBoundary = path.join(fixture.fixtureRoot, "markdown-boundary");
    mkdirSync(markdownBoundary);
    writeFileSync(path.join(markdownBoundary, "semantic.txt"), "semantic input\n");
    symlinkSync("missing-markdown-target", path.join(markdownBoundary, "README.md"));
    assert.doesNotThrow(() =>
      cacheInputRootDigest(root, [path.relative(root, markdownBoundary).replaceAll("\\", "/")]),
    );

    const cache = fixture.create();
    assert.equal((await cache.store(fixture.unit)).stored, true);
    rmSync(fixture.output);
    assert.equal((await cache.lookup(fixture.unit)).outcome, "hit");
    assert.equal(readFileSync(fixture.output, "utf8"), "output-one\n");

    const context = cache.context(fixture.unit);
    const directory = cache.entryDirectory(context.profile, context.inputDigest);
    writeFileSync(path.join(directory, "artifacts/0/0"), "corrupt\n");
    assert.deepEqual(
      await cache.lookup(fixture.unit),
      { outcome: "miss", reason: "record_invalid", profile_id: "fixture_profile", write_eligible: true },
    );

    writeFileSync(fixture.input, "input-two\n");
    assert.equal((await fixture.create().lookup(fixture.unit)).reason, "record_missing");
    writeFileSync(fixture.input, "input-one\n");
    assert.equal((await fixture.create({ toolchainDigest: `sha256:${"d".repeat(64)}` }).lookup(fixture.unit)).reason, "record_missing");
    assert.equal((await fixture.create({ helperDigest: `sha256:${"e".repeat(64)}` }).lookup(fixture.unit)).reason, "record_missing");
  } finally {
    rmSync(fixture.fixtureRoot, { recursive: true, force: true });
  }
}

async function assertCacheModeContract() {
  const fixture = cacheFixture("same_run");
  try {
    const cache = fixture.create();
    assert.equal((await cache.lookup(fixture.unit)).outcome, "miss");
    assert.equal((await cache.store(fixture.unit)).stored, true);
    rmSync(fixture.output);
    assert.equal((await cache.lookup(fixture.unit)).outcome, "hit");
    assert.equal((await fixture.create({ mode: "cold" }).lookup(fixture.unit)).reason, "cold_read_bypass");
    assert.equal((await fixture.create({ mode: "off" }).lookup(fixture.unit)).reason, "mode_off");
  } finally {
    rmSync(fixture.fixtureRoot, { recursive: true, force: true });
  }
}

function cacheEntryFor(fixture, cache) {
  const context = cache.context(fixture.unit);
  return {
    context,
    directory: cache.entryDirectory(context.profile, context.inputDigest),
  };
}

async function assertInvalidCacheEntry(mutator) {
  const fixture = cacheFixture();
  try {
    const cache = fixture.create();
    assert.equal((await cache.store(fixture.unit)).stored, true);
    const { directory } = cacheEntryFor(fixture, cache);
    writeFileSync(fixture.output, "preserved-destination\n", { mode: 0o600 });
    mutator({ fixture, cache, directory });
    assert.deepEqual(
      await cache.lookup(fixture.unit),
      {
        outcome: "miss",
        reason: "record_invalid",
        profile_id: "fixture_profile",
        write_eligible: true,
      },
    );
    assert.equal(readFileSync(fixture.output, "utf8"), "preserved-destination\n");
    assert.equal(existsSync(directory), false, "invalid cache entries must be quarantined");
    const quarantine = path.join(fixture.cacheRoot, ".quarantine/fixture_profile");
    assert.equal(readdirSync(quarantine).length, 1);
  } finally {
    rmSync(fixture.fixtureRoot, { recursive: true, force: true });
  }
}

function mutateRecord(directory, mutate) {
  const recordPath = path.join(directory, "record.json");
  const record = JSON.parse(readFileSync(recordPath, "utf8"));
  mutate(record);
  writeFileSync(recordPath, `${JSON.stringify(record, null, 2)}\n`, { mode: 0o600 });
}

async function assertCacheContainmentMatrix() {
  await assertInvalidCacheEntry(({ directory }) => {
    mutateRecord(directory, (record) => {
      record.artifacts[0].relative_path = "/absolute/result.txt";
    });
  });
  await assertInvalidCacheEntry(({ directory }) => {
    mutateRecord(directory, (record) => {
      record.artifacts[0].relative_path = "../traversal.txt";
    });
  });
  await assertInvalidCacheEntry(({ directory }) => {
    mutateRecord(directory, (record) => {
      record.artifacts[0].destination_class = "repository_artifact";
    });
  });
  await assertInvalidCacheEntry(({ directory }) => {
    mutateRecord(directory, (record) => {
      record.producer_identity = "target:wrong-producer";
      record.artifacts[0].producer_identity = "target:wrong-producer";
    });
  });
  await assertInvalidCacheEntry(({ directory }) => {
    mutateRecord(directory, (record) => {
      record.artifacts[0].mode = "0644";
    });
  });
  await assertInvalidCacheEntry(({ directory }) => {
    rmSync(path.join(directory, "artifacts/0/0"));
  });
  await assertInvalidCacheEntry(({ directory }) => {
    writeFileSync(path.join(directory, "artifacts/0/surplus"), "surplus\n", { mode: 0o600 });
  });
  await assertInvalidCacheEntry(({ directory, fixture }) => {
    const payload = path.join(directory, "artifacts/0/0");
    rmSync(payload);
    symlinkSync(fixture.input, payload);
  });
  await assertInvalidCacheEntry(({ directory, fixture }) => {
    const payload = path.join(directory, "artifacts/0/0");
    rmSync(payload);
    linkSync(fixture.input, payload);
  });
  await assertInvalidCacheEntry(({ directory }) => {
    chmodSync(path.join(directory, "artifacts/0/0"), 0o644);
  });
  await assertInvalidCacheEntry(({ directory }) => {
    const payload = path.join(directory, "artifacts/0/0");
    rmSync(payload);
    execFileSync("mkfifo", [payload]);
  });
  await assertInvalidCacheEntry(({ directory, fixture }) => {
    const record = path.join(directory, "record.json");
    rmSync(record);
    symlinkSync(fixture.input, record);
  });
}

function directoryCacheFixture() {
  const fixture = cacheFixture();
  rmSync(fixture.output);
  mkdirSync(fixture.output, { mode: 0o700 });
  mkdirSync(path.join(fixture.output, "nested"), { mode: 0o750 });
  writeFileSync(path.join(fixture.output, "root.txt"), "root\n", { mode: 0o600 });
  writeFileSync(path.join(fixture.output, "nested/child.txt"), "child\n", { mode: 0o640 });
  fixture.unit.reusable_artifact_outputs[0] = {
    ...fixture.unit.reusable_artifact_outputs[0],
    artifact_type: "directory",
    mode: "0700",
  };
  return fixture;
}

async function assertDirectoryAndConcurrentCacheContract() {
  const directoryFixture = directoryCacheFixture();
  try {
    const cache = directoryFixture.create();
    assert.equal((await cache.store(directoryFixture.unit)).stored, true);
    rmSync(directoryFixture.output, { recursive: true });
    assert.equal((await cache.lookup(directoryFixture.unit)).outcome, "hit");
    assert.equal(readFileSync(path.join(directoryFixture.output, "root.txt"), "utf8"), "root\n");
    assert.equal(readFileSync(path.join(directoryFixture.output, "nested/child.txt"), "utf8"), "child\n");
    assert.equal(lstatSync(directoryFixture.output).mode & 0o777, 0o700);
    assert.equal(lstatSync(path.join(directoryFixture.output, "nested")).mode & 0o777, 0o750);
    assert.equal(lstatSync(path.join(directoryFixture.output, "nested/child.txt")).mode & 0o777, 0o640);
  } finally {
    rmSync(directoryFixture.fixtureRoot, { recursive: true, force: true });
  }

  const concurrent = cacheFixture();
  try {
    const first = concurrent.create();
    const second = concurrent.create();
    const firstStore = await first.store(concurrent.unit);
    writeFileSync(concurrent.output, "output-two\n", { mode: 0o600 });
    const secondStore = await second.store(concurrent.unit);
    assert.equal(firstStore.reason, "stored");
    assert.equal(secondStore.reason, "concurrent_entry");
    assert.equal(secondStore.output_digest, firstStore.output_digest);
    rmSync(concurrent.output);
    assert.equal((await second.lookup(concurrent.unit)).outcome, "hit");
    assert.equal(readFileSync(concurrent.output, "utf8"), "output-one\n");
  } finally {
    rmSync(concurrent.fixtureRoot, { recursive: true, force: true });
  }
}

async function assertDestinationRollbackContract() {
  const fixture = cacheFixture();
  try {
    const secondOutput = path.join(fixture.runRoot, "artifacts/second.txt");
    writeFileSync(secondOutput, "second-cached\n", { mode: 0o600 });
    fixture.unit.reusable_artifact_outputs = [
      {
        artifact_type: "file",
        relative_path: "artifacts/result.txt",
        destination_class: "run_root",
        mode: "0600",
        producer_identity: fixture.unit.unit_id,
      },
      {
        artifact_type: "file",
        relative_path: "artifacts/second.txt",
        destination_class: "run_root",
        mode: "0600",
        producer_identity: fixture.unit.unit_id,
      },
    ];
    const cache = fixture.create();
    assert.equal((await cache.store(fixture.unit)).stored, true);
    writeFileSync(fixture.output, "first-preserved\n", { mode: 0o600 });
    rmSync(secondOutput);
    symlinkSync(fixture.input, secondOutput);
    const result = await cache.lookup(fixture.unit);
    assert.equal(result.reason, "restore_rejected");
    assert.equal(readFileSync(fixture.output, "utf8"), "first-preserved\n");
    assert.equal(lstatSync(secondOutput).isSymbolicLink(), true);
  } finally {
    rmSync(fixture.fixtureRoot, { recursive: true, force: true });
  }
}

function changedDigest(entry, fill) {
  return { ...entry, byte_digest: `sha256:${fill.repeat(64)}` };
}

function cacheableRowUnit(row) {
  return compiler.compile({ kind: "rows", row_ids: [row.row_id] }).units.find((unit) =>
    unit.current_run_evidence_outputs.includes(`rows/${row.row_id}.json`),
  );
}

async function assertDependencyClosureContract() {
  const source = buildSourceSnapshot(root);
  const profile = cacheRegistry.profiles.find((entry) => entry.profile_id === "test_rows");
  const goRow = catalog.rows.find((row) => row.runner === "go" && row.fixture_capability === "none");
  const goUnit = cacheableRowUnit(goRow);
  const statefulRow = catalog.rows.find((row) =>
    new Set(["go", "vitest"]).has(row.runner) &&
    (row.fixture_capability !== "none" || row.service_dependencies.length > 0),
  );
  assert.ok(statefulRow, "the catalog must retain a stateful row cache boundary fixture");
  assert.equal(cacheableRowUnit(statefulRow).cache_policy, "none");
  const securityRow = catalog.rows.find((row) =>
    new Set(["go", "vitest"]).has(row.runner) && row.evidence_class === "security",
  );
  assert.ok(securityRow, "the catalog must retain a security row cache boundary fixture");
  assert.equal(cacheableRowUnit(securityRow).cache_policy, "none");
  const goClosure = resolveCacheDependencyClosure({
    root,
    entries: source.entries,
    profile,
    unit: goUnit,
  });
  assert.equal(goClosure.strategy, "go_packages");
  assert.ok(goClosure.entries.some((entry) => entry.path === "go.mod"));
  assert.ok(goClosure.entries.some((entry) => entry.path === "pnpm-lock.yaml"));
  assert.ok(goClosure.entries.some((entry) => entry.path.startsWith("tools/harness/")));
  assert.ok(
    goClosure.entries.some((entry) => entry.path.startsWith("tools/harness/generated-artifacts/")),
    "generator identities must be closed by the package cache key",
  );
  assert.ok(goClosure.entries.some((entry) => entry.path === `tools/test_families/${goRow.owner_id}.json`));
  const goPackage = goRow.selector.package.replace(/^\.\//u, "");
  const added = {
    path: `${goPackage}/untracked-cache-input.fixture`,
    kind: "file",
    mode: "0600",
    byte_digest: `sha256:${"1".repeat(64)}`,
  };
  const addedClosure = resolveCacheDependencyClosure({
    root,
    entries: [...source.entries, added].sort((left, right) => compareASCII(left.path, right.path)),
    profile,
    unit: goUnit,
  });
  assert.equal(addedClosure.strategy, "go_packages");
  assert.notEqual(addedClosure.digest, goClosure.digest);
  assert.equal(
    resolveCacheDependencyClosure({ root, entries: source.entries, profile, unit: goUnit }).digest,
    goClosure.digest,
    "deleting the untracked input must return to the original package closure",
  );
  const renamedClosure = resolveCacheDependencyClosure({
    root,
    entries: [...source.entries, { ...added, path: `${goPackage}/renamed-cache-input.fixture` }]
      .sort((left, right) => compareASCII(left.path, right.path)),
    profile,
    unit: goUnit,
  });
  assert.notEqual(renamedClosure.digest, addedClosure.digest);
  for (const pathUnderTest of ["pnpm-lock.yaml", "tools/toolchain_pins.json"]) {
    const changed = source.entries.map((entry) =>
      entry.path === pathUnderTest ? changedDigest(entry, pathUnderTest === "pnpm-lock.yaml" ? "2" : "3") : entry,
    );
    const closure = resolveCacheDependencyClosure({ root, entries: changed, profile, unit: goUnit });
    assert.notEqual(closure.digest, goClosure.digest);
  }
  const schemaEntry = goClosure.entries.find((entry) => entry.path.startsWith("tools/schemas/"));
  const schemaChanged = source.entries.map((entry) =>
    entry.path === schemaEntry.path ? changedDigest(entry, "4") : entry,
  );
  assert.notEqual(
    resolveCacheDependencyClosure({ root, entries: schemaChanged, profile, unit: goUnit }).digest,
    goClosure.digest,
  );
  const selectedGoSource = goClosure.entries.find((entry) =>
    entry.path.startsWith(`${goPackage}/`) && entry.path.endsWith(".go"),
  );
  const incomplete = source.entries.map((entry) =>
    entry.path === selectedGoSource.path ? changedDigest(entry, "5") : entry,
  );
  assert.equal(
    resolveCacheDependencyClosure({ root, entries: incomplete, profile, unit: goUnit }).strategy,
    "broad_fallback",
  );
  const unknownUnit = {
    ...goUnit,
    command: {
      ...goUnit.command,
      environment: { ...goUnit.command.environment, CARTULARY_TEST_ROWS: "unknown.row" },
    },
  };
  assert.equal(
    resolveCacheDependencyClosure({ root, entries: source.entries, profile, unit: unknownUnit }).strategy,
    "broad_fallback",
  );

  const vitestRow = catalog.rows.find((row) => row.runner === "vitest");
  const vitestUnit = cacheableRowUnit(vitestRow);
  const tsClosure = resolveCacheDependencyClosure({ root, entries: source.entries, profile, unit: vitestUnit });
  assert.equal(tsClosure.strategy, "typescript_workspaces");
  assert.ok(tsClosure.entries.some((entry) => entry.path === vitestRow.selector.file));
  assert.ok(tsClosure.entries.some((entry) => entry.path === "apps/web/package.json"));
  const tsAdded = {
    path: "apps/web/src/untracked-cache-input.fixture",
    kind: "file",
    mode: "0600",
    byte_digest: `sha256:${"6".repeat(64)}`,
  };
  const tsAddedClosure = resolveCacheDependencyClosure({
    root,
    entries: [...source.entries, tsAdded].sort((left, right) => compareASCII(left.path, right.path)),
    profile,
    unit: vitestUnit,
  });
  assert.equal(tsAddedClosure.strategy, "typescript_workspaces");
  assert.notEqual(tsAddedClosure.digest, tsClosure.digest);
  assert.ok(
    tsClosure.metadata.workspaces.length > 1,
    "the web workspace closure must include transitive repository workspaces",
  );
  const transitiveRoot = tsClosure.metadata.workspaces.find((workspace) => workspace !== "apps/web");
  const transitiveEntry = tsClosure.entries.find((entry) =>
    entry.path.startsWith(`${transitiveRoot}/`) &&
    entry.path !== `${transitiveRoot}/package.json`,
  );
  assert.ok(transitiveEntry, "transitive workspace must contribute executable source");
  const transitiveChanged = source.entries.map((entry) =>
    entry.path === transitiveEntry.path ? changedDigest(entry, "7") : entry,
  );
  assert.notEqual(
    resolveCacheDependencyClosure({
      root,
      entries: transitiveChanged,
      profile,
      unit: vitestUnit,
    }).digest,
    tsClosure.digest,
    "a transitive workspace change must invalidate the selected row",
  );

  const fixtureRoot = mkdtempSync(path.join(root, "tmp/unit-aware-cache-contract."));
  try {
    const runRoot = path.join(fixtureRoot, "run");
    const rowOutput = path.join(runRoot, `rows/${goRow.row_id}.json`);
    mkdirSync(path.dirname(rowOutput), { recursive: true, mode: 0o700 });
    const sourceResult = {
      schema_id: "cartulary.harness_row_result.v2",
      row_id: goRow.row_id,
      terminal_state: "passed",
      duration_ms: 12,
      exit_code: 0,
      failure_class: null,
      failure_reason: null,
      failure_diagnostic: null,
      runner: "go",
      started_at: "2026-01-01T00:00:00.000Z",
      finished_at: "2026-01-01T00:00:00.012Z",
      wall_duration_ms: 12,
    };
    validateSchemaSync(sourceResult.schema_id, sourceResult);
    writeFileSync(rowOutput, `${JSON.stringify(sourceResult, null, 2)}\n`, { mode: 0o600 });
    const cache = new WorkGraphCache({
      root,
      runRoot,
      cacheRoot: path.join(fixtureRoot, "cache"),
      registry: cacheRegistry,
      toolchainDigest: `sha256:${"8".repeat(64)}`,
      helperDigest: `sha256:${"9".repeat(64)}`,
      sourceEntries: source.entries,
      clock: () => new Date("2030-01-01T00:00:00.000Z"),
    });
    assert.equal(cache.context(goUnit).dependencyClosure.strategy, "go_packages");
    assert.equal((await cache.store(goUnit)).stored, true);
    rmSync(rowOutput);
    assert.equal((await cache.lookup(goUnit)).outcome, "hit");
    const replay = JSON.parse(readFileSync(rowOutput, "utf8"));
    assert.equal(replay.started_at, "2030-01-01T00:00:00.000Z");
    assert.equal(replay.duration_ms, 0);
    assert.equal(replay.row_id, goRow.row_id);
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

function assertReleaseInventoryContract() {
  const producerGraph = compiler.compile({
    kind: "target",
    target: "release-inventory-artifacts",
  });
  const producer = producerGraph.units.find(
    (unit) => unit.unit_id === "target:release-inventory-artifacts",
  );
  assert.ok(producer, "release inventory producer must compile as one semantic unit");
  assert.equal(producer.cache_policy, "content_addressed");
  assert.deepEqual(producer.reusable_artifact_outputs, [
    {
      artifact_type: "file",
      relative_path: ".cartulary/release-artifacts/license-report.json",
      destination_class: "repository_artifact",
      mode: "0644",
      producer_identity: "target:release-inventory-artifacts",
    },
    {
      artifact_type: "file",
      relative_path: ".cartulary/release-artifacts/sbom.cyclonedx.json",
      destination_class: "repository_artifact",
      mode: "0644",
      producer_identity: "target:release-inventory-artifacts",
    },
  ]);
  const profile = cacheRegistry.profiles.find(
    (entry) => entry.profile_id === "release_artifacts",
  );
  assert.deepEqual(profile.targets, ["release-inventory-artifacts"]);
  for (const requiredInput of [
    "go.mod",
    "go.sum",
    "package.json",
    "pnpm-lock.yaml",
    "pnpm-workspace.yaml",
    "docker-compose.dev.yml",
    "tools/toolchain_pins.json",
    "tools/release-evidence/generate-sbom-license-evidence.mjs",
    "tools/schemas/cartulary.license_report.v2.schema.json",
  ]) {
    assert.ok(profile.input_roots.includes(requiredInput), `${requiredInput} must invalidate release inventory`);
  }
  const source = buildSourceSnapshot(root);
  const baselineClosure = resolveCacheDependencyClosure({
    root,
    entries: source.entries,
    profile,
    unit: producer,
  });
  assert.equal(baselineClosure.strategy, "broad");
  for (const pathUnderTest of [
    "go.mod",
    "go.sum",
    "package.json",
    "pnpm-lock.yaml",
    "pnpm-workspace.yaml",
    "apps/web/package.json",
    "tools/toolchain_pins.json",
    "docker-compose.dev.yml",
    "tools/release-evidence/generate-sbom-license-evidence.mjs",
    "tools/schemas/cartulary.license_report.v2.schema.json",
  ]) {
    assert.ok(
      baselineClosure.entries.some((entry) => entry.path === pathUnderTest),
      `${pathUnderTest} must occur in the release inventory dependency closure`,
    );
    const changed = source.entries.map((entry) =>
      entry.path === pathUnderTest ? changedDigest(entry, "9") : entry,
    );
    assert.notEqual(
      resolveCacheDependencyClosure({ root, entries: changed, profile, unit: producer }).digest,
      baselineClosure.digest,
      `${pathUnderTest} must change the release inventory cache key`,
    );
  }
  for (const target of ["license-report", "sbom"]) {
    const graph = compiler.compile({ kind: "target", target });
    const validator = graph.units.find((unit) => unit.unit_id === `target:${target}`);
    assert.equal(validator.cache_policy, "none");
    assert.ok(
      validator.needs.includes("target:release-inventory-artifacts"),
      `${target} must consume the canonical producer`,
    );
    assert.deepEqual(validator.reusable_artifact_outputs, []);
  }
}

async function assertExtendedCacheContract() {
  await assertCacheContainmentMatrix();
  await assertDirectoryAndConcurrentCacheContract();
  await assertDestinationRollbackContract();
  await assertDependencyClosureContract();
}

function assertVulnerabilityRevisionContract() {
  const fixtureRoot = mkdtempSync(path.join(root, "tmp/vulnerability-revision-contract."));
  try {
    writeFileSync(path.join(fixtureRoot, "index.json"), "{}\n");
    const proven = resolveVulnerabilityDatabaseRevision({ root, database: fixtureRoot });
    assert.equal(proven.status, "proven");
    assert.match(proven.revision, /^sha256:[a-f0-9]{64}$/u);
    assert.throws(
      () => resolveVulnerabilityDatabaseRevision({
        root,
        database: fixtureRoot,
        declaredRevision: `sha256:${"f".repeat(64)}`,
      }),
      /does not match local content/u,
    );
    assert.equal(resolveVulnerabilityDatabaseRevision({ root }).status, "unknown");
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

function assertScannerParityContract() {
  const evidence = { exitCode: 3, rawOutput: "finding\n", findings: "{\"finding\":1}\n" };
  assert.match(assertScannerEvidenceParity(evidence, evidence), /^sha256:[a-f0-9]{64}$/u);
  assert.throws(
    () => assertScannerEvidenceParity(evidence, { ...evidence, exitCode: 0 }),
    /diverge/u,
  );
}

async function assertSchedulerContract(kind) {
  if (kind === "execution_and_cache_admission") {
    const graph = schedulerFixture();
    const result = simulateWorkGraph({ graph, capacities: new Map([["cpu", 2], ["process", 2]]), durations: new Map([["a", 10], ["b", 20], ["c", 5]]) });
    assert.deepEqual(result.admissions.slice(0, 2), ["a", "b"]);
    const dependencyFailure = simulateWorkGraph({
      graph,
      capacities: new Map([["cpu", 2], ["process", 2]]),
      outcomes: new Map([["a", "product"]]),
    });
    assert.equal(dependencyFailure.states.a, "failed");
    assert.equal(dependencyFailure.states.c, "skipped");
    assert.equal(
      dependencyFailure.events.find((event) => event.unit_id === "c" && event.event === "skipped")?.status,
      "skipped",
    );

    const fixtureGraph = buildWorkGraph(graph.units.map((unit) =>
      unit.unit_id === "a"
        ? { ...unit, fixture_lease: "managed_process" }
        : unit,
    ));
    const liveEvents = [];
    const fixtureFailure = await runWorkGraph({
      graph: fixtureGraph,
      capacities: new Map([["cpu", 2], ["process", 2]]),
      cwd: root,
      environment: {},
      fixtureBroker: {
        acquire: async () => { throw new Error("fixture unavailable"); },
        close: async () => {},
      },
      executeUnit: async () => ({ status: "passed", exit_code: 0 }),
      onEvent: (event) => liveEvents.push(event),
    });
    assert.deepEqual(
      liveEvents,
      fixtureFailure.events,
      "scheduler consumers must receive the complete ordered event journal before terminal projection",
    );
    assert.deepEqual(
      Object.fromEntries(Object.entries(fixtureFailure.unit_results).map(([unitID, terminal]) => [unitID, terminal.status])),
      { a: "failed", b: "passed", c: "skipped" },
    );
    assert.equal(fixtureFailure.unit_results.a.failure_class, "harness");
    assert.equal(fixtureFailure.unit_results.a.failure_reason, "fixture_error");
    assert.equal(fixtureFailure.unit_results.a.exit_code, 3);

    const classifiedFailure = await runWorkGraph({
      graph: buildWorkGraph([{
        ...schedulerFixture().units[0],
        fixture_lease: "managed_process",
      }]),
      capacities: new Map([["cpu", 1], ["process", 1]]),
      cwd: root,
      environment: {},
      fixtureBroker: {
        acquire: async () => {
          const error = new Error("classified service readiness failure");
          error.failure_class = "infra";
          error.failure_reason = "service_readiness_timeout";
          error.artifact_refs = ["_shared/test-services/suite/service-scope.json"];
          throw error;
        },
        close: async () => {},
      },
      executeUnit: async () => ({ status: "passed", exit_code: 0 }),
    });
    assert.equal(classifiedFailure.unit_results.a.failure_class, "infra");
    assert.equal(classifiedFailure.unit_results.a.failure_reason, "service_readiness_timeout");
    assert.equal(classifiedFailure.unit_results.a.exit_code, 3);
    assert.deepEqual(classifiedFailure.unit_results.a.artifact_refs, [
      "_shared/test-services/suite/service-scope.json",
    ]);
    assertManagedSuiteFailurePropagation();

    const hitGraph = buildWorkGraph([schedulerFixture().units[0]]);
    let executions = 0;
    const cacheHit = await runWorkGraph({
      graph: hitGraph,
      capacities: new Map([["cpu", 1], ["process", 1]]),
      cwd: root,
      environment: {},
      cache: {
        validateGraph: () => {},
        lookup: async () => ({
          outcome: "hit",
          reason: "record_valid",
          profile_id: "fixture",
          output_digest: `sha256:${"f".repeat(64)}`,
        }),
        store: async () => { throw new Error("cache hit must not be stored again"); },
      },
      executeUnit: async () => {
        executions += 1;
        return { status: "passed", exit_code: 0 };
      },
    });
    assert.equal(executions, 0);
    assert.deepEqual(cacheHit.admissions, []);
    assert.ok(cacheHit.events.some((event) => event.event === "cache_hit"));
    assert.ok(cacheHit.events.some((event) => event.event === "completed"));
    assert.equal(
      cacheHit.events.some((event) => ["admitted", "started"].includes(event.event)),
      false,
      "a cache hit must complete before resource admission",
    );
  } else if (kind === "cancellation") {
    const cancelled = simulateWorkGraph({ graph: schedulerFixture(), capacities: new Map([["cpu", 1], ["process", 1]]), cancelAtMs: 1 });
    assert.ok(cancelled.events.some((entry) => entry.event === "cancelled"));
  } else if (kind === "capacity_snapshot") {
    const schedulerRegistry = readJSON("tools/scheduler_resource_registry.json");
    const cpuPolicy = schedulerRegistry.capacity_policies.cpu_tokens;
    assert.equal(cpuPolicy.safety_margin_percent, 25);
    assert.equal(cpuCapacityWithSafetyMargin(24, cpuPolicy), 19);
    assert.equal(cpuCapacityWithSafetyMargin(4, cpuPolicy), 3);
    assert.equal(cpuCapacityWithSafetyMargin(1, cpuPolicy), 1);
    const snapshot = captureCapabilitySnapshot({
      root,
      override: { schema_id: "cartulary.harness_capacity_override.v1", cpu_tokens: 2, memory_bytes: 1048576, process_slots: 2, io_tokens: 2, port_lanes: 1, writable_volume: true },
      services: { postgres: true },
    });
    validateSchemaSync(snapshot.schema_id, snapshot);
    assert.equal(snapshot.cpu_tokens, 2);
    assert.equal(snapshot.postgres_lanes, 2);
    assert.throws(
      () => captureCapabilitySnapshot({
        root,
        override: {
          schema_id: "cartulary.harness_capacity_override.v1",
          cpu_tokens: 2,
          postgres_lanes: 3,
        },
      }),
      /postgres_lanes=3 exceeds the detected policy bound 2/u,
    );
  } else if (kind === "cache_registry_targets") {
    assert.ok(cacheRegistry.profiles.every((profile) => new Set(profile.targets).size === profile.targets.length));
  } else if (kind === "content_cache") {
    await assertContentCacheContract();
    await assertExtendedCacheContract();
  } else if (kind === "cache_modes") {
    await assertCacheModeContract();
  } else if (kind === "release_inventory") {
    assertReleaseInventoryContract();
  } else if (kind === "vulnerability_revision") {
    assertVulnerabilityRevisionContract();
  } else if (kind === "scanner_parity") {
    assertScannerParityContract();
  } else {
    throw new Error(`unknown scheduler contract ${kind}`);
  }
}

function semanticCase(id, name, acceptanceIDs, run) {
  return { id, name, acceptance_ids: acceptanceIDs, run };
}

const suiteCases = {
  boundaries: [
    ...generalCases,
    semanticCase(
      "performance_evidence_and_runtime_boundary",
      "performance evidence generations and suite runtime are isolated",
      ["TH-HARNESS-AC-099", "TH-HARNESS-AC-100"],
      () => assertBoundaryContract("performance_and_runtime"),
    ),
    semanticCase("schema_attachment_uniqueness", "active schema attachments are unique", ["TH-HARNESS-AC-005"], () => assertBoundaryContract("attachment_uniqueness")),
    semanticCase("current_schema_presence", "current harness schemas are present", ["TH-HARNESS-AC-005"], () => assertBoundaryContract("current_schema_presence")),
    semanticCase("retired_schema_absence", "retired harness schemas remain absent", ["TH-HARNESS-AC-005"], () => assertBoundaryContract("retired_schema_absence")),
    semanticCase("work_graph_owner_validation", "work graph owner validates against its current schema", ["TH-HARNESS-AC-082"], () => assertBoundaryContract("work_graph_owner_validation")),
    semanticCase("relative_backing_script_paths", "task-surface backing scripts are repository-relative", ["TH-HARNESS-AC-016"], () => assertBoundaryContract("relative_backing_scripts")),
  ],
  command_surface: [
    semanticCase("graph_entrypoint_contract", "graph entrypoints use one work-graph command model", ["TH-HARNESS-AC-082"], () => assertCommandSurfaceContract("graph_entrypoints")),
    semanticCase("public_command_identity_contract", "public command identities are closed and versioned", ["TH-HARNESS-AC-001"], () => assertCommandSurfaceContract("command_identities")),
    semanticCase("graph_input_contract", "observable graph targets expose cache and capacity inputs", ["TH-HARNESS-AC-083"], () => assertCommandSurfaceContract("graph_inputs")),
    semanticCase("public_output_contract", "observable targets publish current run and target summaries", ["TH-HARNESS-AC-004"], () => assertCommandSurfaceContract("output_contracts")),
  ],
  evidence: [
    semanticCase("go_selector_build_context", "Go selectors honor the canonical build context", ["TH-HARNESS-AC-018"], () => assertEvidenceContract("selector_build_context")),
    semanticCase("tier_partition_closure", "catalog tier partitions close every row exactly once", ["TH-HARNESS-AC-018"], () => assertEvidenceContract("tier_partition")),
    semanticCase("tier_monotonic_closure", "catalog tier reachability is monotonic", ["TH-HARNESS-AC-018"], () => assertEvidenceContract("tier_monotonicity")),
    semanticCase("fixture_partition_closure", "fixture capabilities partition current rows", ["TH-HARNESS-AC-086"], () => assertEvidenceContract("fixture_partition")),
    semanticCase("active_owner_row_coverage", "every active owner retains current row coverage", ["TH-HARNESS-AC-018"], () => assertEvidenceContract("active_owner_coverage")),
  ],
  graph: [
    semanticCase("aggregate_graph_determinism", "aggregate work graphs are deterministic", ["TH-HARNESS-AC-082"], () => assertGraphContract("aggregate_determinism")),
    semanticCase("row_evidence_output_contract", "row graphs declare exact current-run evidence outputs", ["TH-HARNESS-AC-087"], () => assertGraphContract("row_evidence_outputs")),
    semanticCase("go_lpt_cpu_budget", "Go LPT planning preserves row closure and CPU budgets", ["TH-HARNESS-AC-087"], () => assertGraphContract("go_lpt_budget")),
    semanticCase("target_graph_validation", "target graphs validate before execution", ["TH-HARNESS-AC-082"], () => assertGraphContract("target_graph_validation")),
  ],
  scheduler: [
    semanticCase("scheduler_execution_and_cache_admission", "scheduler dependency, fixture, and cache admission lifecycles are exact", ["TH-HARNESS-AC-089", "TH-HARNESS-AC-091"], () => assertSchedulerContract("execution_and_cache_admission")),
    semanticCase("scheduler_cancellation", "scheduler cancellation emits terminal evidence", ["TH-HARNESS-AC-089"], () => assertSchedulerContract("cancellation")),
    semanticCase("capacity_snapshot_contract", "capacity snapshots enforce detected resource bounds", ["TH-HARNESS-AC-085"], () => assertSchedulerContract("capacity_snapshot")),
    semanticCase("cache_registry_target_uniqueness", "cache registry target bindings are unique", ["TH-HARNESS-AC-091"], () => assertSchedulerContract("cache_registry_targets")),
    semanticCase("content_cache_security_and_closure", "content cache restore is complete, contained, and dependency-closed", ["TH-HARNESS-AC-091", "TH-HARNESS-AC-101"], () => assertSchedulerContract("content_cache")),
    semanticCase("cache_mode_contract", "cache modes preserve explicit read and write behavior", ["TH-HARNESS-AC-091"], () => assertSchedulerContract("cache_modes")),
    semanticCase("release_inventory_contract", "release inventory has one deterministic paired cache producer and fresh validators", ["TH-HARNESS-AC-101"], () => assertSchedulerContract("release_inventory")),
    semanticCase("vulnerability_revision_contract", "vulnerability database revisions are content-proven", ["TH-HARNESS-AC-098"], () => assertSchedulerContract("vulnerability_revision")),
    semanticCase("scanner_evidence_parity", "scanner evidence parity rejects divergent executions", ["TH-HARNESS-AC-098"], () => assertSchedulerContract("scanner_parity")),
  ],
};

const allCases = Object.values(suiteCases).flat();
assert.equal(new Set(allCases.map((entry) => entry.id)).size, allCases.length);
for (const entry of allCases) {
  assert.match(entry.id, /^[a-z][a-z0-9_]*$/u);
  assert.ok(entry.acceptance_ids.length > 0);
  assert.equal(new Set(entry.acceptance_ids).size, entry.acceptance_ids.length);
  assert.ok(entry.acceptance_ids.every((id) => /^TH-HARNESS-AC-[0-9]{3}$/u.test(id)));
}

export function runContractSuite(suite) {
  const cases = suiteCases[suite];
  assert.ok(cases?.length > 0, `contract suite ${suite} is empty`);
  for (const entry of cases) {
    test(`${entry.id}: ${entry.name}`, async () => {
      await entry.run();
    });
  }
}
