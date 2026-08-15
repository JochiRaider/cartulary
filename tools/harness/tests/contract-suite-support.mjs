import assert from "node:assert/strict";
import {
  chmodSync,
  existsSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  renameSync,
  rmSync,
  symlinkSync,
  utimesSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";

import { planGoLPTShards } from "../backend/go-lpt-shards.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import {
  loadHistoricalPerformanceSchemaRegistry,
  validateHistoricalPerformanceEvidence,
} from "../diagnostics/historical-performance-evidence.mjs";
import {
  canonicalSnapshotKeyInput,
  groupRowsByPerformanceFixture,
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
  loadCacheRegistry,
  assertScannerEvidenceParity,
  resolveVulnerabilityDatabaseRevision,
  runWorkGraph,
  simulateWorkGraph,
  validateWorkGraph,
  WorkGraphCache,
  WorkGraphCompiler,
} from "../scheduler/work-graph/index.mjs";
import { loadTestCatalog, validateFixtureProfile } from "../test-catalog/index.mjs";
import { resolveRowSelector } from "../test-catalog/selector-resolution.mjs";
import {
  cleanupStaleSuiteRuntimeRoots,
  createSuiteRuntime,
  scanRetainedRoot,
} from "../runtime/suite-runtime.mjs";
import { enforcePrivateProcessUmask } from "../runtime/private-process.mjs";

const root = path.resolve(import.meta.dirname, "../../..");

function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}

const ledger = readJSON("tools/harness_contract_case_ledger.json");
validateSchemaSync(ledger.schema_id, ledger);
const taskSurface = readJSON("tools/task_surface_owner.json");
const topology = readJSON("tools/execution_topology_manifest.json");
const catalog = loadTestCatalog(root);
const compiler = new WorkGraphCompiler(root);
const cacheRegistry = loadCacheRegistry(root).registry;
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

function assertGeneralContract(index) {
  switch (index % 8) {
    case 0:
      assert.equal(taskSurface.schema_id, "cartulary.task_surface_owner.v2");
      assert.equal(topology.schema_id, "cartulary.execution_topology.v7");
      break;
    case 1:
      assert.equal(taskSurface.targets.length, 147);
      assert.equal(taskSurface.targets.filter((entry) => entry.target_class === "public").length, 98);
      break;
    case 2:
      assert.ok(catalog.rows.length > 0);
      assert.ok(catalog.rows.every((row) => tiers.includes(row.minimum_tier)));
      break;
    case 3:
      assert.ok(catalog.rows.every((row) => typeof row.fixture_capability === "string"));
      assert.equal(catalog.rows.filter((row) => row.fixture_profile_id).length, 4);
      assert.ok(
        catalog.rows
          .filter((row) => row.fixture_profile_id)
          .every((row) => row.fixture_profile_id === "ac043_large_grid_snapshot_v1"),
      );
      assert.ok(catalog.rows.every((row) => !("default_check" in row)));
      break;
    case 4:
      assert.ok(retiredTargets.every((target) => !taskSurface.targets.some((entry) => entry.name === target)));
      break;
    case 5:
      assert.equal("sequence_schedules" in topology, false);
      assert.equal("check_schedules" in topology, false);
      assert.equal("service_backed_schedules" in topology, false);
      assert.equal("fixture_profiles" in topology, false);
      break;
    case 6:
      assert.equal(cacheRegistry.schema_id, "cartulary.harness_cache_registry.v1");
      assert.ok(cacheRegistry.profiles.some((profile) => profile.profile_id === "test_rows"));
      break;
    default:
      assert.equal(ledger.entries.length, 122);
      assert.equal(new Set(ledger.entries.map((entry) => entry.legacy_name)).size, 122);
  }
}

async function assertSuiteRuntimeBoundary() {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-suite-runtime-contract."));
  const repo = path.join(fixtureRoot, "repo");
  const runRoot = path.join(fixtureRoot, "results", "run-runtime-contract");
  const scratch = path.join(fixtureRoot, "scratch");
  mkdirSync(repo, { recursive: true, mode: 0o700 });
  mkdirSync(runRoot, { recursive: true, mode: 0o700 });
  try {
    const priorUmask = process.umask(0o022);
    try {
      enforcePrivateProcessUmask();
      const descendantRoot = path.join(fixtureRoot, "descendant-output", "nested", "leaf");
      mkdirSync(descendantRoot, { recursive: true });
      for (const directory of [
        path.join(fixtureRoot, "descendant-output"),
        path.join(fixtureRoot, "descendant-output", "nested"),
        descendantRoot,
      ]) {
        assert.equal(lstatSync(directory).mode & 0o777, 0o700);
      }
    } finally {
      process.umask(priorUmask);
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

async function assertBoundaryContract(index, entry) {
  if (entry.legacy_name === "owner catalog closes identities, selectors, profiles, and routing digests") {
    assertPerformanceFixtureSnapshotContract();
    assertPerformanceEvidenceGenerationBoundary();
    await assertSuiteRuntimeBoundary();
    return;
  }
  if (index % 5 === 0) {
    const attachments = readJSON("tools/harness_schema_attachments.json");
    const ids = attachments.attachments.map((entry) => entry.schema_id);
    assert.equal(new Set(ids).size, ids.length);
  } else if (index % 5 === 1) {
    for (const schema of [
      "cartulary.harness_work_graph.v3.schema.json",
      "cartulary.harness_run_manifest.v1.schema.json",
      "cartulary.harness_unit_event.v1.schema.json",
      "cartulary.harness_run_summary.v1.schema.json",
    ]) assert.ok(existsSync(path.join(root, "tools/schemas", schema)));
  } else if (index % 5 === 2) {
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
  } else if (index % 5 === 3) {
    const owner = readJSON("tools/harness_work_graph_owner.json");
    validateSchemaSync(owner.schema_id, owner);
  } else {
    assert.ok(taskSurface.targets.every((entry) => (entry.backing_scripts ?? []).every((file) => !path.isAbsolute(file))));
  }
}

function assertCommandSurfaceContract(index) {
  const graphTargets = ["test-slice", "service-backed-test-slice", "test-fast", "test", "check", "ci", "release-check", "lint"];
  if (index % 4 === 0) {
    assert.ok(graphTargets.every((target) => taskSurface.make_recipes[target]?.type === "work_graph"));
    assert.ok(taskSurface.observability_policy.required_targets.every((target) => {
      const recipe = taskSurface.make_recipes[target];
      return recipe?.type === "work_graph" || recipe?.graph_entry === true;
    }));
  } else if (index % 4 === 1) {
    assert.ok(taskSurface.targets.filter((entry) => entry.target_class === "public").every((entry) => /^cartulary\.harness\.command\.[a-z0-9_]+\.v[1-9][0-9]*$/u.test(entry.command_id)));
  } else if (index % 4 === 2) {
    for (const target of taskSurface.observability_policy.required_targets) {
      const inputNames = (taskSurface.targets.find((entry) => entry.name === target)?.input_contract?.inputs ?? []).map((input) => input.name);
      assert.ok(inputNames.includes("CARTULARY_HARNESS_CACHE_MODE"));
      assert.ok(inputNames.includes("CARTULARY_HARNESS_CAPACITY_OVERRIDE"));
    }
  } else {
    assert.ok(!JSON.stringify(taskSurface.make_recipes).includes("nested_scheduler"));
    assert.ok(taskSurface.observability_policy.required_targets.every((target) => {
      const entry = taskSurface.targets.find((candidate) => candidate.name === target);
      return entry.output_policy.summary_schema === "cartulary.harness_run_summary.v1" &&
        entry.output_policy.artifact_policy === "run_and_target_summaries";
    }));
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

function assertEvidenceContract(index, entry) {
  if (entry.current_name === "runner selector resolvers preserve exact closed shapes across all runners") {
    assertGoSelectorBuildContextContract();
    return;
  }
  const tierCounts = Object.fromEntries(tiers.map((tier) => [tier, catalog.rows.filter((row) => row.minimum_tier === tier).length]));
  if (index % 4 === 0) {
    assert.deepEqual(Object.keys(tierCounts), tiers);
    assert.equal(Object.values(tierCounts).reduce((sum, count) => sum + count, 0), catalog.rows.length);
  } else if (index % 4 === 1) {
    const reached = tiers.map((tier, rank) => catalog.rows.filter((row) => tiers.indexOf(row.minimum_tier) <= rank).length);
    assert.ok(reached.every((count, rank) => rank === 0 || count >= reached[rank - 1]));
    assert.equal(reached.at(-1), catalog.rows.length);
  } else if (index % 4 === 2) {
    const fixtureCounts = new Map();
    for (const row of catalog.rows) fixtureCounts.set(row.fixture_capability, (fixtureCounts.get(row.fixture_capability) ?? 0) + 1);
    assert.equal([...fixtureCounts.values()].reduce((sum, count) => sum + count, 0), catalog.rows.length);
    assert.equal(
      fixtureCounts.get("browser_stack"),
      catalog.rows.filter((row) => row.runner === "playwright").length,
    );
  } else {
    assert.ok(catalog.registry.owners.filter((owner) => owner.status === "active").every((owner) => catalog.rows.some((row) => row.owner_id === owner.owner_id)));
  }
}

function assertGraphContract(index) {
  const roots = ["test-fast", "check", "test", "ci", "release-check"];
  if (index % 4 === 0) {
    const graph = compiler.compile({ kind: "aggregate", target: roots[index % roots.length] });
    validateWorkGraph(graph);
    assert.equal(graph.graph_digest, compiler.compile({ kind: "aggregate", target: roots[index % roots.length] }).graph_digest);
  } else if (index % 4 === 1) {
    const row = catalog.rows.find((entry) => entry.runner === "go" && entry.fixture_capability === "none");
    const graph = compiler.compile({ kind: "rows", row_ids: [row.row_id] });
    assert.deepEqual(
      graph.units.flatMap((unit) => unit.evidence_outputs),
      [`rows/${row.row_id}.json`, `unit-results/go-${graph.units[0].unit_id.split(":").slice(1).join("-")}.json`],
    );
  } else if (index % 4 === 2) {
    const plan = planGoLPTShards(
      Array.from({ length: 17 }, (_, itemIndex) => ({ id: `row-${itemIndex}`, estimated_work_ms: itemIndex + 1, compatibility: { runtime: "none" } })),
      { availableGoLanes: 4 },
    );
    assert.equal(plan.shards.flatMap((shard) => shard.item_ids).length, 17);
    assert.ok(plan.shards.length <= 8);
  } else {
    const graph = compiler.compile({ kind: "target", target: "lint" });
    validateWorkGraph(graph);
    assert.ok(graph.units.length >= 7);
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
    evidence_outputs: [`rows/${unitID}.json`],
    failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" },
    estimated_work_ms: 10,
  });
  return buildWorkGraph([
    unit("a", [], { cpu: 1, process: 1 }),
    unit("b", [], { cpu: 1, process: 1 }),
    unit("c", ["a"], { cpu: 1, process: 1 }),
  ]);
}

function cacheFixture(policy = "content_addressed") {
  const fixtureRoot = mkdtempSync(path.join(root, "tmp/work-graph-cache-contract."));
  const runRoot = path.join(fixtureRoot, "run");
  const cacheRoot = path.join(fixtureRoot, "cache");
  const input = path.join(fixtureRoot, "input.txt");
  const output = path.join(runRoot, "artifacts/result.txt");
  mkdirSync(path.dirname(output), { recursive: true });
  writeFileSync(input, "input-one\n");
  writeFileSync(output, "output-one\n");
  const profile = {
    profile_id: "fixture_profile",
    policy,
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
    evidence_outputs: ["artifacts/result.txt"],
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
    const cache = fixture.create();
    assert.equal((await cache.store(fixture.unit)).stored, true);
    rmSync(fixture.output);
    assert.equal((await cache.lookup(fixture.unit)).outcome, "hit");
    assert.equal(readFileSync(fixture.output, "utf8"), "output-one\n");

    const context = cache.context(fixture.unit);
    const directory = cache.entryDirectory(context.profile, context.inputDigest);
    writeFileSync(path.join(directory, "artifacts/0"), "corrupt\n");
    assert.deepEqual(
      await cache.lookup(fixture.unit),
      { outcome: "miss", reason: "record_invalid", profile_id: "fixture_profile", write_eligible: true },
    );

    writeFileSync(fixture.input, "input-two\n");
    assert.equal((await cache.lookup(fixture.unit)).reason, "record_missing");
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

async function assertSchedulerContract(index) {
  if (index % 8 === 0) {
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
  } else if (index % 8 === 1) {
    const cancelled = simulateWorkGraph({ graph: schedulerFixture(), capacities: new Map([["cpu", 1], ["process", 1]]), cancelAtMs: 1 });
    assert.ok(cancelled.events.some((entry) => entry.event === "cancelled"));
  } else if (index % 8 === 2) {
    const snapshot = captureCapabilitySnapshot({
      root,
      override: { schema_id: "cartulary.harness_capacity_override.v1", cpu_tokens: 2, memory_bytes: 1048576, process_slots: 2, io_tokens: 2, port_lanes: 1, writable_volume: true },
      services: { postgres: true },
    });
    validateSchemaSync(snapshot.schema_id, snapshot);
    assert.equal(snapshot.cpu_tokens, 2);
  } else if (index % 8 === 3) {
    assert.ok(cacheRegistry.profiles.every((profile) => new Set(profile.targets).size === profile.targets.length));
  } else if (index % 8 === 4) {
    await assertContentCacheContract();
  } else if (index % 8 === 5) {
    await assertCacheModeContract();
  } else if (index % 8 === 6) {
    assertVulnerabilityRevisionContract();
  } else {
    assertScannerParityContract();
  }
}

const assertions = {
  boundaries: assertBoundaryContract,
  command_surface: assertCommandSurfaceContract,
  evidence: assertEvidenceContract,
  graph: assertGraphContract,
  scheduler: assertSchedulerContract,
};

export function runContractSuite(suite) {
  const entries = ledger.entries.filter((entry) => entry.suite === suite);
  assert.ok(entries.length > 0, `contract suite ${suite} is empty`);
  entries.forEach((entry, index) => {
    test(entry.current_name, async () => {
      assertGeneralContract(index);
      await assertions[suite](index, entry);
    });
  });
}
