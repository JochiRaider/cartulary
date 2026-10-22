import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import {
  chmodSync,
  copyFileSync,
  existsSync,
  linkSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  realpathSync,
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
  normalizeExactFileSets,
  resolveExactFileSets,
} from "../static-analysis/exact-file-sets.mjs";
import {
  buildWorkGraph,
  captureCapabilitySnapshot,
  cpuCapacityWithSafetyMargin,
  assertScannerEvidenceParity,
  resolveVulnerabilityDatabaseRevision,
  validateWorkGraph,
  CachePlan,
  resolveCacheDependencyClosure,
} from "../scheduler/work-graph/index.mjs";
import { buildSourceSnapshot } from "../test-catalog/index.mjs";
import { validateFixtureProfile } from "../test-catalog/index.mjs";
import { resolveRowSelector } from "../test-support/owner-fixtures.mjs";
import { createRuntime } from "../runtime/session-client.mjs";
import { validateFrontendAttachment } from "../test-support/owner-fixtures.mjs";
import { resolveFrontendArtifact, sealFrontendArtifact } from "../readiness/frontend-artifact.mjs";
import { resolveBrowserFrontendArtifact } from "../generated-artifacts/execution-topology.mjs";
import { createContractTestContext } from "./contract-test-context.mjs";
import { assertLazyCaseContext, assertImportAndFailureIsolation } from "./contract-initialization-cases.mjs";
import { assertPostgresCatalogClosure, assertPostgresPolicyFixtures, assertFixtureBuilderClosure, assertSelectorResolutionScope, assertPrivateRowSelection, assertSourceSnapshotIdentity } from "./contract-catalog-cases.mjs";

import { assertHarnessTierRouting, assertCatalogTopologyValidation, assertCatalogTierSelection } from "./contract-tier-cases.mjs";

import { assertPrivateChildCaptureBoundary } from "./private-child-capture-cases.mjs";

const root = path.resolve(import.meta.dirname, "../../..");

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}

const tiers = ["fast", "standard", "full", "release"];

function assertPerformanceEvidenceGenerationBoundary() {
  const attachments = readJSON("tools/harness_schema_attachments.json");
  const activeIDs = new Set(attachments.attachments.map((entry) => entry.schema_id));
  for (const schemaID of [
    "cartulary.browser_target_result.v8",
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
  assert.throws(
    () => validateSchemaSync(aggregate.schema_id, {
      ...aggregate,
      summaries: [{ samples: Array.from({ length: 100_000 }, () => 1) }],
    }),
    /additional properties/u,
    "aggregate v3 must grow by immutable summary references, not raw samples",
  );
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
    [profile.fixture_profile_id, syntheticProfile.fixture_profile_id].sort(compareASCII),
  );
  const syntheticVector = syntheticVectors.vectors.find((entry) => entry.name === syntheticProfile.fixture_profile_id);
  assert.ok(syntheticVector);
  validateSchemaSync("cartulary.performance_fixture_snapshot_key.v2", syntheticVector.input);
  assert.equal(
    syntheticVector.input.fixture_profile_id,
    syntheticProfile.fixture_profile_id,
  );
  const syntheticGo = renderPerformanceFixtureProfilesGo([profile, syntheticProfile]);
  assert.equal(
    renderPerformanceFixtureProfilesGo([syntheticProfile, profile]),
    syntheticGo,
    "descriptor output must not depend on registration order",
  );
  assert.equal(
    renderPerformanceFixtureKeyVectors([syntheticProfile, profile]),
    renderPerformanceFixtureKeyVectors([profile, syntheticProfile]),
    "key-vector output must not depend on registration order",
  );
  assert.match(syntheticGo, /synthetic_grid_snapshot_v1/u);
  assert.doesNotMatch(
    syntheticGo,
    /profile\.FixtureProfileID == "timeline_large_grid_snapshot_v2"/u,
  );
  assert.doesNotMatch(
    syntheticGo,
    /contracts\/|RuntimeProfileID|ResourceProfileID|runtime_profile_id|resource_profile_id/u,
    "generated Go descriptors must not contain source paths or runtime routing identities",
  );

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
  assert.deepEqual(
    grouped.map((entry) => entry.fixture_profile_id),
    [profile.fixture_profile_id, syntheticProfile.fixture_profile_id].sort(compareASCII),
  );
  const syntheticGroup = grouped.find((entry) => entry.fixture_profile_id === syntheticProfile.fixture_profile_id);
  assert.ok(syntheticGroup);
  assert.deepEqual(syntheticGroup.predicate_ids, ["perf.synthetic.fixture_profile.v1"]);
  assert.deepEqual(syntheticGroup.row_ids, ["module.synthetic.measurement.fixture_profile"]);
  const sameProfile = groupRowsByPerformanceFixture(root, [row, { ...row, row_id: "module.synthetic.measurement.same_profile" }], { registry: syntheticRegistry });
  assert.equal(sameProfile.length, 1);
  assert.deepEqual(sameProfile[0].row_ids, [row.row_id, "module.synthetic.measurement.same_profile"].sort(compareASCII));
  assert.deepEqual(groupRowsByPerformanceFixture(root, [{ ...row, fixture_profile_id: undefined, verification_ids: [] }], { registry: syntheticRegistry }), []);
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

  const original = context.readJSON("tools/performance_fixture_snapshot_owner.json");
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-fixture-registry."));
  const registryFile = path.join(fixtureRoot, "registry.json");
  const expectRegistryFailure = (mutate, pattern) => {
    const candidate = structuredClone(original);
    const selected = candidate.profiles.find((entry) => entry.fixture_profile_id === profile.fixture_profile_id);
    assert.ok(selected);
    mutate(selected);
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
          candidate.source_contract_refs[0].contract_id.replace(/\.v[1-9][0-9]*$/u, ".v99");
      },
      /contract identity does not match/u,
    );

    const inactive = structuredClone(original);
    inactive.profiles.find((entry) => entry.fixture_profile_id === profile.fixture_profile_id).status = "inactive";
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
    run(context) {
      assert.equal(context.taskSurface.schema_id, "cartulary.task_surface_owner.v4");
      assert.equal(context.topology.schema_id, "cartulary.execution_topology.v15");
    },
  },
  {
    id: "current_target_roster",
    name: "current target roster and public surface are exact",
    acceptance_ids: ["TH-HARNESS-AC-001"],
    run(context) {
      const identity = (entry) => [entry.name, entry.command_id, entry.target_class];
      const targets = context.taskSurface.targets;
      assert.equal(new Set(targets.map((entry) => entry.name)).size, targets.length);
      const publicTargets = targets.filter((entry) => entry.target_class === "public");
      assert.equal(new Set(publicTargets.map((entry) => entry.command_id)).size, publicTargets.length);
      assert.deepEqual(context.readJSON("tools/task_surface_manifest.json").targets.map(identity), targets.map(identity));
    },
  },
  {
    id: "catalog_tier_closure",
    name: "catalog rows use only current minimum tiers",
    acceptance_ids: ["TH-HARNESS-AC-018"],
    run(context) {
      assert.ok(context.catalog.rows.length > 0);
      assert.ok(context.catalog.rows.every((row) => [...tiers, "explicit_only"].includes(row.aggregate_placement)));
    },
  },
  {
    id: "fixture_profile_closure",
    name: "fixture profiles, retired rows, and canonical construction are closed",
    acceptance_ids: ["TH-HARNESS-AC-086", "TH-HARNESS-AC-096", "TH-HARNESS-AC-098"],
    run(context) {
      assert.ok(context.catalog.rows.every((row) => typeof row.fixture_capability === "string"));
      assert.ok(context.catalog.rows.every((row) => !("default_check" in row)));
      assertFixtureBuilderClosure(context);
      const migration = context.rowMigrations.migrations.find((entry) =>
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
    run(context) {
      const retiredTargets = readJSON("tools/harness/tests/retired-make-names.json").retired_names;
      assert.ok(retiredTargets.every((target) =>
        !context.taskSurface.targets.some((entry) => entry.name === target)));
      const results = mkdtempSync(path.join(tmpdir(), "cartulary-retired-commands-"));
      try {
        const env = { ...process.env, CARTULARY_TEST_RESULTS_DIR: results, CARTULARY_TEST_RUN_ID: "retired-rejection" };
        delete env.MAKEFLAGS; delete env.MAKEOVERRIDES; delete env.MFLAGS;
        let failure;
        try { execFileSync("make", ["--no-print-directory", "--keep-going", "--", ...retiredTargets], { cwd: root, env, stdio: "pipe" }); }
        catch (error) { failure = error; }
        assert.equal(failure?.status, 2, "every retired command rejects before dispatch");
        const output = `${failure.stdout}${failure.stderr}`;
        for (const target of retiredTargets) assert.ok(output.includes(`No rule to make target '${target}'`), `${target} must have no callable rule`);
        assert.deepEqual(readdirSync(results), [], "rejected commands must not prepare retained execution");
      } finally { rmSync(results, { recursive: true, force: true }); }
    },
  },
  {
    id: "retired_topology_field_absence",
    name: "retired schedule and fixture fields remain absent",
    acceptance_ids: ["TH-HARNESS-AC-082"],
    run(context) {
      assert.equal("sequence_schedules" in context.topology, false);
      assert.equal("check_schedules" in context.topology, false);
      assert.equal("service_backed_schedules" in context.topology, false);
      assert.equal("fixture_profiles" in context.topology, false);
    },
  },
  {
    id: "cache_registry_identity",
    name: "cache registry exposes the current test-row profile",
    acceptance_ids: ["TH-HARNESS-AC-091"],
    run(context) {
      assert.equal(context.cacheRegistry.schema_id, "cartulary.harness_cache_registry.v2");
      assert.ok(context.cacheRegistry.profiles.some((profile) => profile.profile_id === "test_rows"));
    },
  },
];

async function assertProducerUmaskBoundary() {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-suite-runtime-contract."));
  const repo = path.join(fixtureRoot, "repo");
  const runRoot = path.join(fixtureRoot, "results", "run-runtime-contract");
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

  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

async function assertBoundaryContract(context, kind) {
  switch (kind) {
    case "performance_and_runtime":
      assertPerformanceFixtureSnapshotContract(context);
      assertPerformanceEvidenceGenerationBoundary();
      await assertProducerUmaskBoundary();
      return;
    case "attachment_uniqueness": {
      const attachments = readJSON("tools/harness_schema_attachments.json");
      const ids = attachments.attachments.map((entry) => entry.schema_id);
      assert.equal(new Set(ids).size, ids.length);
      return;
    }
    case "current_schema_presence":
      for (const schema of [
        "cartulary.harness_work_graph.v8.schema.json",
        "cartulary.harness_execution_context.v3.schema.json",
        "cartulary.harness_runtime_plan.v6.schema.json",
        "cartulary.harness_run_manifest.v3.schema.json",
        "cartulary.harness_unit_event.v4.schema.json",
        "cartulary.harness_run_summary.v1.schema.json",
      ]) assert.ok(existsSync(path.join(root, "tools/schemas", schema)));
      return;
    case "retired_schema_absence":
      for (const oldSchema of [
        "cartulary.harness_execution_context.v1.schema.json",
        "cartulary.harness_execution_context.v2.schema.json",
        "cartulary.harness_runtime_plan.v3.schema.json",
        "cartulary.harness_work_graph.v5.schema.json",
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
        "cartulary.harness_work_graph_owner.v8",
        readJSON("tools/harness_work_graph_owner.json"),
      );
      return;
    case "relative_backing_scripts":
      assert.ok(context.taskSurface.targets.every((entry) =>
        (entry.backing_scripts ?? []).every((file) => !path.isAbsolute(file))));
      return;
    case "exact_file_sets":
      assertExactFileSetContract();
      return;
    default:
      throw new Error(`unknown boundary contract ${kind}`);
  }
}

function assertExactFileSetContract() {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-exact-file-set."));
  const importsRoot = path.join(fixtureRoot, "internal/modules/imports");
  const handlerPath = path.join(importsRoot, "http_handlers.go");
  const definition = (paths = ["internal/modules/imports/http_handlers.go"]) => [{
    id: "imports-transport-bindings",
    paths,
    discovery: {
      scan_roots: ["internal/modules/imports"],
      tokens: ["httpapi.BindOwnerRoutes("],
      production_only: true,
    },
  }];
  try {
    mkdirSync(importsRoot, { recursive: true });
    writeFileSync(
      handlerPath,
      "package imports\nfunc bind() { httpapi.BindOwnerRoutes(nil, nil, \"module.imports\", nil) }\n",
    );
    const resolved = resolveExactFileSets(fixtureRoot, normalizeExactFileSets(definition()));
    assert.deepEqual(
      resolved.get("imports-transport-bindings").map((file) => file.relative),
      ["internal/modules/imports/http_handlers.go"],
    );

    for (const invalid of [
      "/absolute.go",
      "C:/absolute.go",
      "internal/modules/imports/../routes.go",
      "internal\\modules\\imports\\routes.go",
      "internal/modules/imports/*.go",
    ]) {
      assert.throws(() => normalizeExactFileSets(definition([invalid])));
    }
    assert.throws(() => normalizeExactFileSets(definition([])), /non-empty array/u);
    assert.throws(
      () => normalizeExactFileSets(definition([
        "internal/modules/imports/http_handlers.go",
        "internal/modules/imports/http_handlers.go",
      ])),
      /unique/u,
    );
    assert.throws(
      () => resolveExactFileSets(
        fixtureRoot,
        normalizeExactFileSets(definition(["internal/modules/imports/missing.go"])),
      ),
      /does not exist/u,
    );

    const stalePath = path.join(importsRoot, "stale.go");
    writeFileSync(stalePath, "package imports\n");
    assert.throws(
      () => resolveExactFileSets(
        fixtureRoot,
        normalizeExactFileSets(definition(["internal/modules/imports/stale.go"])),
      ),
      /missing=\[internal\/modules\/imports\/http_handlers.go\].*stale=\[internal\/modules\/imports\/stale.go\]/u,
    );
    rmSync(stalePath);

    const omittedPath = path.join(importsRoot, "new_routes.go");
    writeFileSync(
      omittedPath,
      "package imports\nfunc bindMore() { httpapi.BindOwnerRoutes(nil, nil, \"module.imports\", nil) }\n",
    );
    assert.throws(
      () => resolveExactFileSets(fixtureRoot, normalizeExactFileSets(definition())),
      /missing=\[internal\/modules\/imports\/new_routes.go\]/u,
    );
    rmSync(omittedPath);

    const directoryPath = path.join(importsRoot, "directory.go");
    mkdirSync(directoryPath);
    assert.throws(
      () => resolveExactFileSets(
        fixtureRoot,
        normalizeExactFileSets(definition(["internal/modules/imports/directory.go"])),
      ),
      /regular file/u,
    );
    rmSync(directoryPath, { recursive: true });

    const symlinkPath = path.join(importsRoot, "binding_link.go");
    symlinkSync("http_handlers.go", symlinkPath);
    assert.throws(
      () => resolveExactFileSets(
        fixtureRoot,
        normalizeExactFileSets(definition(["internal/modules/imports/binding_link.go"])),
      ),
      /symlink/u,
    );
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

function assertCommandSurfaceContract(context, kind) {
  const graphTargets = ["test-slice", "browser-e2e", "test-fast", "check", "ci", "release-check", "lint"];
  switch (kind) {
    case "graph_entrypoints":
      assert.ok(graphTargets.every((target) => context.taskSurface.make_recipes[target]?.type === "work_graph"));
      assert.ok(context.taskSurface.observability_policy.required_targets.every((target) => {
        const recipe = context.taskSurface.make_recipes[target];
        return recipe?.type === "work_graph" || recipe?.graph_entry === true;
      }));
      return;
    case "command_identities":
      assert.ok(context.taskSurface.targets.filter((entry) => entry.target_class === "public").every((entry) =>
        /^cartulary\.harness\.command\.[a-z0-9_]+\.v[1-9][0-9]*$/u.test(entry.command_id)));
      return;
    case "graph_inputs":
      for (const target of context.taskSurface.observability_policy.required_targets) {
        const inputNames = (context.taskSurface.targets.find((entry) => entry.name === target)?.input_contract?.inputs ?? []).map((input) => input.name);
        assert.ok(inputNames.includes("CARTULARY_HARNESS_CACHE_MODE"));
        assert.ok(inputNames.includes("CARTULARY_HARNESS_CAPACITY_OVERRIDE"));
      }
      return;
    case "output_contracts":
      assert.ok(!JSON.stringify(context.taskSurface.make_recipes).includes("nested_scheduler"));
      assert.ok(context.taskSurface.observability_policy.required_targets.every((target) => {
        const entry = context.taskSurface.targets.find((candidate) => candidate.name === target);
        return entry.output_policy.summary_schema === "cartulary.harness_run_summary.v1" &&
          entry.output_policy.artifact_policy === "run_and_target_summaries";
      }));
      return;
    default:
      throw new Error(`unknown command-surface contract ${kind}`);
  }
}

function assertGoSelectorBuildContextContract() {
  const sourceRoot = mkdtempSync(path.join(tmpdir(), "cartulary-selector-build-context."));
  const fixtureRoot = path.join(sourceRoot, "internal/fixture");
  mkdirSync(fixtureRoot, { recursive: true });
  mkdirSync(path.join(sourceRoot, "tools"));
  copyFileSync(path.join(root, "tools/toolchain_pins.json"), path.join(sourceRoot, "tools/toolchain_pins.json"));
  copyFileSync(path.join(root, "tools/generated_artifact_policy.json"), path.join(sourceRoot, "tools/generated_artifact_policy.json"));
  const relativePackage = "./internal/fixture";
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
      resolveRowSelector({ root: sourceRoot, row: row("TestOrdinary"), runner, taskSurfaceCommandIDs: new Set() }),
      [`go:${relativePackage}:TestOrdinary`],
    );
    assert.deepEqual(
      resolveRowSelector({ root: sourceRoot, row: row("TestApplicablePlatform"), runner, taskSurfaceCommandIDs: new Set() }),
      [`go:${relativePackage}:TestApplicablePlatform`],
    );
    assert.throws(
      () => resolveRowSelector({ root: sourceRoot, row: row("TestPrivateProfile"), runner, taskSurfaceCommandIDs: new Set() }),
      /TestPrivateProfile is excluded from the Go runner build context linux\/amd64/u,
    );
    assert.throws(
      () => resolveRowSelector({ root: sourceRoot, row: row("TestExcludedPlatform"), runner, taskSurfaceCommandIDs: new Set() }),
      /TestExcludedPlatform is excluded from the Go runner build context linux\/amd64/u,
    );
  } finally {
    rmSync(sourceRoot, { recursive: true, force: true });
  }
}

function assertEvidenceContract(context, kind) {
  switch (kind) {
    case "selector_build_context":
      assertGoSelectorBuildContextContract();
      return;
    case "tier_partition": {
      const tierCounts = Object.fromEntries([...tiers, "explicit_only"].map((tier) => [tier, context.catalogMetadata.rows.filter((row) => row.aggregate_placement === tier).length]));
      assert.deepEqual(Object.keys(tierCounts), [...tiers, "explicit_only"]);
      assert.equal(Object.values(tierCounts).reduce((sum, count) => sum + count, 0), context.catalogMetadata.rows.length);
      return;
    }
    case "tier_monotonicity": {
      const reached = tiers.map((tier, rank) =>
        context.catalogMetadata.rows.filter((row) => row.aggregate_placement !== "explicit_only" && tiers.indexOf(row.aggregate_placement) <= rank).length);
      assert.ok(reached.every((count, rank) => rank === 0 || count >= reached[rank - 1]));
      assert.equal(reached.at(-1), context.catalogMetadata.rows.filter((row) => row.aggregate_placement !== "explicit_only").length);
      return;
    }
    case "fixture_partition": {
      const fixtureCounts = new Map();
      for (const row of context.catalogMetadata.rows) {
        fixtureCounts.set(row.fixture_capability, (fixtureCounts.get(row.fixture_capability) ?? 0) + 1);
      }
      assert.equal([...fixtureCounts.values()].reduce((sum, count) => sum + count, 0), context.catalogMetadata.rows.length);
      assert.equal(
        fixtureCounts.get("browser_stack"),
        context.catalogMetadata.rows.filter((row) => row.runner === "playwright").length,
      );
      return;
    }
    case "active_owner_coverage":
      assert.ok(context.catalogMetadata.registry.owners.filter((owner) => owner.status === "active").every((owner) =>
        context.catalogMetadata.rows.some((row) => row.owner_id === owner.owner_id)));
      return;
    default:
      throw new Error(`unknown evidence contract ${kind}`);
  }
}

function assertGraphContract(context, kind) {
  const roots = ["test-fast", "check", "ci", "release-check"];
  switch (kind) {
    case "aggregate_determinism":
      for (const target of roots) {
        const graph = context.compiler.compile({ kind: "aggregate", target });
        validateWorkGraph(graph);
        assert.equal(
          graph.graph_digest,
          context.compiler.compile({ kind: "aggregate", target }).graph_digest,
        );
      }
      return;
    case "row_evidence_outputs": {
      const row = context.catalog.rows.find((entry) => entry.runner === "go" && entry.fixture_capability === "none");
      const graph = context.compiler.compile({ kind: "rows", row_ids: [row.row_id] });
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
    case "visual_semantic_partition": {
      const graph = context.compiler.compile({ kind: "target", target: "harness.suite_browser_visual" });
      const groups = graph.units.filter((unit) => unit.unit_id.startsWith("browser_group:visual:"));
      const rows = context.catalog.rows.filter((row) => row.runner === "playwright" && row.selector.stage === "visual");
      const selected = groups.flatMap((unit) => {
        const ids = unit.row_bindings.map((row) => row.row_id);
        assert.ok(ids.length > 0, "visual cohorts retain semantic row identities");
        const selectedRows = ids.map((id) => context.catalog.rowByID.get(id));
        if (ids.length > 1) assert.ok(selectedRows.every((row) => row.browser_state_scope === "test_owned"), "only explicitly test-owned visual rows may share a cohort");
        assert.equal(unit.owner_id, "harness.browser", "visual execution ownership is independent of reporting ownership");
        return ids;
      });
      assert.equal(new Set(selected).size, selected.length, "visual claims execute exactly once");
      assert.deepEqual(selected.sort(), rows.map((row) => row.row_id).sort(), "no active visual claim is omitted");
      for (const unit of graph.units) {
        assert.ok(Buffer.byteLength(`.tmp-${unit.unit_id}.json.2147483647`) <= 255,
          "private visual group/reset artifacts leave room for atomic-write names");
      }
      validateWorkGraph(graph);
      return;
    }
    case "measurement_semantic_identity": {
      const graph = context.compiler.compile({ kind: "target", target: "harness.suite_browser_measurement" });
      const rows = context.catalog.rows.filter((row) => row.selector?.file === "apps/web/e2e/measurement/timeline-grid.spec.ts");
      const ids = new Set(rows.map((row) => row.row_id));
      const groups = graph.units.filter((unit) => unit.unit_id.startsWith("browser_group:measurement:") &&
        unit.row_bindings.some((binding) => ids.has(binding.row_id)));
      assert.equal(groups.length, 4);
      assert.deepEqual(groups.map((unit) => unit.command.environment.CARTULARY_FIXTURE_ROW_ID).sort(), rows.map((row) => row.row_id).sort());
      for (const unit of groups) {
        assert.deepEqual(unit.row_bindings, [{ row_id: unit.command.environment.CARTULARY_FIXTURE_ROW_ID, runner: "playwright" }], "dedicated measurement retains its full catalog identity");
        assert.match(unit.unit_id, /^browser_group:measurement:cohort-[a-f0-9]{64}$/u, "bounded cohort identity binds the exact semantic selection");
      }
      return;
    }
    case "target_graph_validation": {
      const graph = context.compiler.compile({ kind: "target", target: "lint" });
      validateWorkGraph(graph);
      assert.ok(graph.units.length >= 7);
      return;
    }
    default:
      throw new Error(`unknown graph contract ${kind}`);
  }
}

function changedDigest(entry, fill) {
  return { ...entry, byte_digest: `sha256:${fill.repeat(64)}` };
}

function cacheableRowUnit(context, row) {
  return context.compiler.compile({ kind: "rows", row_ids: [row.row_id] }).units.find((unit) =>
    unit.current_run_evidence_outputs.includes(`rows/${row.row_id}.json`),
  );
}

function assertTypeScriptDependencyClosureContract(context, profile, row, unit) {
  const fixtureRoot = mkdtempSync(path.join(tmpdir(), "cartulary-contract-dependencies."));
  const selectedWorkspace = row.selector.file.split("/").slice(0, 2).join("/");
  const importer = "packages/cache-fixture/src/codec.ts";
  const fixture = "contracts/tabularingest/clipboard.v1.json";
  const write = (relative, contents) => {
    const filename = path.join(fixtureRoot, relative);
    mkdirSync(path.dirname(filename), { recursive: true });
    writeFileSync(filename, contents);
  };
  const snapshot = (directory = fixtureRoot) => readdirSync(directory).flatMap((name) => {
    const filename = path.join(directory, name);
    const stat = lstatSync(filename);
    if (stat.isDirectory()) return snapshot(filename);
    return [{
      path: path.relative(fixtureRoot, filename).replaceAll("\\", "/"),
      kind: stat.isSymbolicLink() ? "symlink" : "file",
      mode: "0644",
      byte_digest: stat.isSymbolicLink() ? `sha256:${"0".repeat(64)}` :
        `sha256:${createHash("sha256").update(readFileSync(filename)).digest("hex")}`,
    }];
  });
  const resolve = (entries = snapshot()) => resolveCacheDependencyClosure({
    root: fixtureRoot, entries, profile, unit,
    resolverCache: new Map([["test-catalog", context.catalog]]),
  });
  const importFixture = (specifier) => write(importer, `import fixture from "${specifier}";\nexport default fixture;\n`);
  try {
    write(`${selectedWorkspace}/package.json`, JSON.stringify({ name: "@fixture/selected", dependencies: { "@fixture/codec": "workspace:*" } }));
    write(row.selector.file, 'import fixture from "@fixture/codec";\n');
    write("packages/cache-fixture/package.json", JSON.stringify({ name: "@fixture/codec" }));
    importFixture("../../../" + fixture);
    write(fixture, '{"version":1}\n');
    const initial = resolve();
    assert.equal(initial.strategy, "typescript_workspaces");
    assert.ok(initial.entries.some((entry) => entry.path === fixture));
    assert.ok(initial.entries.some((entry) => entry.path === row.selector.file));
    assert.ok(initial.entries.some((entry) => entry.path === `${selectedWorkspace}/package.json`));
    assert.ok(initial.entries.some((entry) => entry.path === importer), "transitive workspace must contribute executable source");
    assert.deepEqual([...initial.metadata.workspaces].sort(compareASCII), [selectedWorkspace, "packages/cache-fixture"].sort(compareASCII));
    const untracked = `${selectedWorkspace}/src/untracked-cache-input.fixture`;
    write(untracked, "untracked input\n");
    const added = resolve();
    assert.equal(added.strategy, "typescript_workspaces");
    assert.notEqual(added.digest, initial.digest, "untracked workspace content must invalidate the selected row");
    rmSync(path.join(fixtureRoot, untracked));
    assert.equal(resolve().digest, initial.digest, "deleting the untracked input restores the closure");
    const originalImporter = readFileSync(path.join(fixtureRoot, importer), "utf8");
    write(importer, `${originalImporter}// changed transitive workspace source\n`);
    const transitiveChanged = resolve();
    assert.equal(transitiveChanged.strategy, "typescript_workspaces");
    assert.notEqual(transitiveChanged.digest, initial.digest, "a transitive workspace change must invalidate the selected row");
    write(importer, originalImporter);
    assert.equal(resolve().digest, initial.digest);
    const unprovedModule = "tools/harness/fixtures/semantic-fixture/index.mjs";
    write(unprovedModule, "export default {};\n");
    importFixture("../../../" + unprovedModule);
    const fallback = resolve();
    assert.equal(fallback.strategy, "broad_fallback", "existing non-workspace imports still require fail-closed fallback");
    assert.ok(fallback.entries.some((entry) => entry.path === unprovedModule));
    rmSync(path.join(fixtureRoot, unprovedModule));
    write(importer, originalImporter);
    assert.equal(resolve().digest, initial.digest);
    assert.equal(resolve(snapshot().filter((entry) => entry.path !== fixture)).strategy, "broad_fallback", "imports absent from the snapshot must fail closed");
    const captured = snapshot();
    write(fixture, '{"version":2}\n');
    const changed = resolve();
    assert.equal(changed.strategy, "typescript_workspaces");
    assert.notEqual(changed.digest, initial.digest, "contract content must invalidate the workspace closure");
    assert.equal(resolve(captured).strategy, "broad_fallback", "snapshot drift must not produce a narrow key");
    write(fixture, "invalid JSON");
    assert.equal(resolve().strategy, "broad_fallback", "malformed contracts are unsupported");
    rmSync(path.join(fixtureRoot, fixture));
    assert.equal(resolve().strategy, "broad_fallback", "missing imports must fail closed");
    mkdirSync(path.join(fixtureRoot, fixture));
    assert.equal(resolve().strategy, "broad_fallback", "a directory cannot satisfy a JSON import");
    rmSync(path.join(fixtureRoot, fixture), { recursive: true });
    write(fixture, '{"version":1}\n');
    assert.equal(resolve().digest, initial.digest, "restoring the same dependency restores its closure");
    const renamed = "contracts/tabularingest/renamed.json";
    renameSync(path.join(fixtureRoot, fixture), path.join(fixtureRoot, renamed));
    importFixture("../../../" + renamed);
    assert.equal(resolve().strategy, "typescript_workspaces");
    assert.notEqual(resolve().digest, initial.digest);
    symlinkSync(path.join(fixtureRoot, renamed), path.join(fixtureRoot, fixture));
    importFixture("../../../" + fixture);
    assert.equal(resolve().strategy, "broad_fallback", "symlinked contract imports are unsupported");
    rmSync(path.join(fixtureRoot, fixture));
    renameSync(path.join(fixtureRoot, "contracts/tabularingest"), path.join(fixtureRoot, "contract-fixtures"));
    symlinkSync(path.join(fixtureRoot, "contract-fixtures"), path.join(fixtureRoot, "contracts/tabularingest"));
    importFixture("../../../" + renamed);
    assert.equal(resolve().strategy, "broad_fallback", "symlinked ancestors must not escape containment");
    for (const unsupported of ["../../../../outside.json", "../../../contract-fixtures/renamed.json", "../../../contracts/unsupported.txt"]) {
      importFixture(unsupported);
      assert.equal(resolve().strategy, "broad_fallback");
    }
  } finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}

async function assertDependencyClosureContract(context) {
  const source = buildSourceSnapshot(root);
  const profile = context.cacheRegistry.profiles.find((entry) => entry.profile_id === "test_rows");
  const goRow = context.catalog.rows.find((row) => row.runner === "go" && row.fixture_capability === "none");
  const goUnit = cacheableRowUnit(context, goRow);
  const statefulRow = context.catalog.rows.find((row) =>
    new Set(["go", "vitest"]).has(row.runner) &&
    (row.fixture_capability !== "none" || row.service_dependencies.length > 0),
  );
  assert.ok(statefulRow, "the catalog must retain a stateful row cache boundary fixture");
  assert.equal(cacheableRowUnit(context, statefulRow).cache_policy, "none");
  const securityRow = context.catalog.rows.find((row) =>
    new Set(["go", "vitest"]).has(row.runner) && row.evidence_class === "security",
  );
  assert.ok(securityRow, "the catalog must retain a security row cache boundary fixture");
  assert.equal(cacheableRowUnit(context, securityRow).cache_policy, "none");
  const goClosure = resolveCacheDependencyClosure({
    root,
    entries: source.entries,
    profile,
    unit: goUnit,
  });
  assert.equal(goClosure.strategy, "go_packages", JSON.stringify(goClosure.metadata));
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

  const vitestRow = context.catalog.rows.find((row) => row.runner === "vitest");
  const vitestUnit = cacheableRowUnit(context, vitestRow);
  assertTypeScriptDependencyClosureContract(context, profile, vitestRow, vitestUnit);
  let catalogLoads = 0;
  const cache = new CachePlan({ root, registry: context.cacheRegistry,
    toolchainDigest: `sha256:${"8".repeat(64)}`, helperDigest: `sha256:${"9".repeat(64)}`,
    sourceEntries: source.entries, getCatalog: () => { catalogLoads++; return context.catalog; } });
  const first = cache.context(goUnit);
  assert.equal(first.dependencyClosure.strategy, "go_packages");
  assert.equal(cache.context(goUnit), first);
  assert.equal(catalogLoads, 1);

}

function assertReleaseInventoryContract(context) {
  const producerGraph = context.compiler.compile({
    kind: "target",
    target: "harness.release_inventory_artifacts",
  });
  const producer = producerGraph.units.find(
    (unit) => unit.unit_id === "operation:harness.release_inventory_artifacts",
  );
  assert.ok(producer, "release inventory producer must compile as one semantic unit");
  assert.equal(producer.cache_policy, "content_addressed");
  assert.deepEqual(producer.reusable_artifact_outputs, [
    {
      artifact_type: "file",
      relative_path: ".cartulary/release-artifacts/license-report.json",
      destination_class: "repository_artifact",
      mode: "0644",
      producer_identity: "operation:harness.release_inventory_artifacts",
    },
    {
      artifact_type: "file",
      relative_path: ".cartulary/release-artifacts/sbom.cyclonedx.json",
      destination_class: "repository_artifact",
      mode: "0644",
      producer_identity: "operation:harness.release_inventory_artifacts",
    },
  ]);
  const profile = context.cacheRegistry.profiles.find(
    (entry) => entry.profile_id === "release_artifacts",
  );
  assert.deepEqual(profile.operation_ids, ["harness.release_inventory_artifacts"]);
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
  for (const target of ["harness.license_report", "harness.sbom"]) {
    const graph = context.compiler.compile({ kind: "target", target });
    const validator = graph.units.find((unit) => unit.unit_id === `operation:${target}`);
    assert.equal(validator.cache_policy, "none");
    assert.ok(
      validator.needs.includes("operation:harness.release_inventory_artifacts"),
      `${target} must consume the canonical producer`,
    );
    assert.deepEqual(validator.reusable_artifact_outputs, []);
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

async function assertSchedulerContract(context, kind) {
  if (kind === "capacity_snapshot") {
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
    assert.equal(snapshot.object_store_lanes, 2);
    const portBound = captureCapabilitySnapshot({ root,
      override: { schema_id: "cartulary.harness_capacity_override.v1", cpu_tokens: 2 } });
    assert.equal(portBound.port_lanes, 4);
    assert.throws(() => captureCapabilitySnapshot({ root,
      override: { schema_id: "cartulary.harness_capacity_override.v1", cpu_tokens: 2, port_lanes: 5 } }),
    /port_lanes=5 exceeds the detected policy bound 4/u);
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
    assert.ok(context.cacheRegistry.profiles.every((profile) => new Set(profile.operation_ids).size === profile.operation_ids.length));
  } else if (kind === "content_cache") {
    await assertDependencyClosureContract(context);

  } else if (kind === "release_inventory") {
    assertReleaseInventoryContract(context);
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
    semanticCase("private_child_capture_boundary", "private child capture owns allocation and exception-safe cleanup", ["TH-HARNESS-AC-011", "TH-HARNESS-AC-014", "TH-HARNESS-AC-015"], assertPrivateChildCaptureBoundary),
    semanticCase("public_timezone_projection_boundary", "packaged timezone projection excludes backend-only artifacts", ["TH-HARNESS-AC-005"], () => {
      execFileSync(process.execPath, ["tools/harness/generated-artifacts/tests/test-json-shapes.mjs"], { cwd: root, stdio: "pipe" });
    }),
    semanticCase("lazy_case_context", "case contexts load only their own dependencies", ["TH-HARNESS-AC-016"], assertLazyCaseContext),
    semanticCase("import_and_failure_isolation", "support imports are inert and policy failures remain case-local", ["TH-HARNESS-AC-016"], assertImportAndFailureIsolation),
    ...generalCases,
    semanticCase(
      "performance_evidence_and_runtime_boundary",
      "performance evidence generations and suite runtime are isolated",
      ["TH-HARNESS-AC-099", "TH-HARNESS-AC-100"],
      (context) => assertBoundaryContract(context, "performance_and_runtime"),
    ),
    semanticCase("schema_attachment_uniqueness", "active schema attachments are unique", ["TH-HARNESS-AC-005"], (context) => assertBoundaryContract(context, "attachment_uniqueness")),
    semanticCase("current_schema_presence", "current harness schemas are present", ["TH-HARNESS-AC-005"], (context) => assertBoundaryContract(context, "current_schema_presence")),
    semanticCase("retired_schema_absence", "retired harness schemas remain absent", ["TH-HARNESS-AC-005"], (context) => assertBoundaryContract(context, "retired_schema_absence")),
    semanticCase("work_graph_owner_validation", "work graph owner validates against its current schema", ["TH-HARNESS-AC-082"], (context) => assertBoundaryContract(context, "work_graph_owner_validation")),
    semanticCase("relative_backing_script_paths", "task-surface backing scripts are repository-relative", ["TH-HARNESS-AC-016"], (context) => assertBoundaryContract(context, "relative_backing_scripts")),
    semanticCase("exact_boundary_file_sets", "exact boundary file sets reject stale, unsafe, and omitted paths", ["TH-HARNESS-AC-039"], (context) => assertBoundaryContract(context, "exact_file_sets")),
  ],
  command_surface: [
    semanticCase("source_snapshot_identity", "source snapshots preserve byte identity and reject changes during hashing", ["TH-HARNESS-AC-066"], assertSourceSnapshotIdentity),
    semanticCase("private_row_selection_boundary", "private row selection preserves metadata closure and exact source checks", ["TH-HARNESS-AC-063", "TH-HARNESS-AC-065"], assertPrivateRowSelection),
    semanticCase("graph_entrypoint_contract", "graph entrypoints use one work-graph command model", ["TH-HARNESS-AC-082"], (context) => assertCommandSurfaceContract(context, "graph_entrypoints")),
    semanticCase("public_command_identity_contract", "public command identities are closed and versioned", ["TH-HARNESS-AC-001"], (context) => assertCommandSurfaceContract(context, "command_identities")),
    semanticCase("graph_input_contract", "observable graph targets expose cache and capacity inputs", ["TH-HARNESS-AC-083"], (context) => assertCommandSurfaceContract(context, "graph_inputs")),
    semanticCase("public_output_contract", "observable targets publish current run and target summaries", ["TH-HARNESS-AC-004"], (context) => assertCommandSurfaceContract(context, "output_contracts")),
  ],
  evidence: [
    semanticCase("postgres_catalog_closure", "Postgres fixture counts reconcile with active policy partitions", ["TH-HARNESS-AC-086"], assertPostgresCatalogClosure),
    semanticCase("postgres_policy_growth_and_rejection", "Postgres policy admits growth and rejects invalid isolation approvals", ["TH-HARNESS-AC-086"], assertPostgresPolicyFixtures),
    semanticCase("go_selector_build_context", "Go selectors honor the canonical build context", ["TH-HARNESS-AC-018"], (context) => assertEvidenceContract(context, "selector_build_context")),
    semanticCase("catalog_selector_scope", "selector facts refresh between loads and preserve path checks", ["TH-HARNESS-AC-063"], assertSelectorResolutionScope),
    semanticCase("tier_partition_closure", "catalog tier partitions close every row exactly once", ["TH-HARNESS-AC-018"], (context) => assertEvidenceContract(context, "tier_partition")),
    semanticCase("tier_monotonic_closure", "catalog tier reachability is monotonic", ["TH-HARNESS-AC-018"], (context) => assertEvidenceContract(context, "tier_monotonicity")),
    semanticCase("fixture_partition_closure", "fixture capabilities partition current rows", ["TH-HARNESS-AC-086"], (context) => assertEvidenceContract(context, "fixture_partition")),
    semanticCase("active_owner_row_coverage", "every active owner retains current row coverage", ["TH-HARNESS-AC-018"], (context) => assertEvidenceContract(context, "active_owner_coverage")),
  ],
  graph: [
    semanticCase("harness_self_test_tier_routing", "full-tier self-tests preserve direct selection and product coverage", ["TH-HARNESS-AC-082"], assertHarnessTierRouting),
    semanticCase("catalog_topology_input_validation", "catalog topology rejects retired raw selectors and malformed coverage inputs", ["TH-HARNESS-AC-082"], assertCatalogTopologyValidation),
    semanticCase("catalog_independent_tier_selection", "catalog tiers select before readiness without policy coupling", ["TH-HARNESS-AC-082"], assertCatalogTierSelection),
    semanticCase("visual_semantic_partition", "visual groups bound complete catalog claims without duplicate or missing execution", ["TH-HARNESS-AC-011", "TH-HARNESS-AC-082"], (context) => assertGraphContract(context, "visual_semantic_partition")),
    semanticCase("measurement_semantic_identity", "measurement groups preserve complete semantic identities independently of private storage", ["TH-HARNESS-AC-011", "TH-HARNESS-AC-082"], (context) => assertGraphContract(context, "measurement_semantic_identity")),
    semanticCase("aggregate_graph_determinism", "aggregate work graphs are deterministic", ["TH-HARNESS-AC-082"], (context) => assertGraphContract(context, "aggregate_determinism")),
    semanticCase("row_evidence_output_contract", "row graphs declare exact current-run evidence outputs", ["TH-HARNESS-AC-087"], (context) => assertGraphContract(context, "row_evidence_outputs")),
    semanticCase("go_lpt_cpu_budget", "Go LPT planning preserves row closure and CPU budgets", ["TH-HARNESS-AC-087"], (context) => assertGraphContract(context, "go_lpt_budget")),
    semanticCase("target_graph_validation", "target graphs validate before execution", ["TH-HARNESS-AC-082"], (context) => assertGraphContract(context, "target_graph_validation")),
  ],
  scheduler: [
    semanticCase("capacity_snapshot_contract", "capacity snapshots enforce detected resource bounds", ["TH-HARNESS-AC-085"], (context) => assertSchedulerContract(context, "capacity_snapshot")),
    semanticCase("cache_registry_target_uniqueness", "cache registry target bindings are unique", ["TH-HARNESS-AC-091"], (context) => assertSchedulerContract(context, "cache_registry_targets")),
    semanticCase("content_cache_security_and_closure", "cache preparation preserves complete source dependency closure", ["TH-HARNESS-AC-091", "TH-HARNESS-AC-101"], (context) => assertSchedulerContract(context, "content_cache")),
    semanticCase("release_inventory_contract", "release inventory has one deterministic paired cache producer and fresh validators", ["TH-HARNESS-AC-101"], (context) => assertSchedulerContract(context, "release_inventory")),
    semanticCase("vulnerability_revision_contract", "vulnerability database revisions are content-proven", ["TH-HARNESS-AC-098"], (context) => assertSchedulerContract(context, "vulnerability_revision")),
    semanticCase("scanner_evidence_parity", "scanner evidence parity rejects divergent executions", ["TH-HARNESS-AC-098"], (context) => assertSchedulerContract(context, "scanner_parity")),
  ],
};

suiteCases.evidence.push(semanticCase(
  "browser_artifact_admission", "browser attachments bind the selected artifact and current complete build", ["TH-HARNESS-AC-086"], (context) => {
    const temporaryRoot = mkdtempSync(path.join(tmpdir(), "cartulary-browser-artifact-"));
    try {
      const production = resolveBrowserFrontendArtifact(root, "webserver-backed");
      const measurement = resolveBrowserFrontendArtifact(root, "measurement");
      for (const [rowID, expectedProducer] of [
        ["module.auth.browser.browser_sign_in_exposes_the_ordinary_session_sur_2ec8df6229", "build-web"],
        ["web.design.accessibility.account_menu_accessible_keyboard_and_focus_state_3c31f28451", "build-web"],
        ["web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde", "build-web"],
        ["module.networkflow.measurement.saved_graph_dom_ceiling", "build-web-measurement"],
      ]) {
        const graph = context.compiler.compileRows([rowID]);
        assert.ok(graph.units.some((unit) => unit.unit_id === `operation:harness.${expectedProducer.replaceAll("-", "_")}`), `row ${rowID} must build ${expectedProducer}`);
        // Measurement still builds the backend's embedded production assets.
        // Production work must never acquire the measurement producer.
        if (expectedProducer === "build-web") assert.ok(!graph.units.some((unit) => unit.unit_id === "operation:harness.build_web_measurement"), `production row ${rowID} must not build measurement output`);
      }
      assert.notEqual(production.producer_report_key, measurement.producer_report_key);
      assert.notEqual(production.path, measurement.path);
      for (const artifact of [production, measurement]) {
        const runID = artifact.producer_report_key;
        const runRoot = path.join(temporaryRoot, runID);
        mkdirSync(runRoot, { mode: 0o700 });
        writeFileSync(path.join(runRoot, "run-manifest.json"), JSON.stringify({ source_digest: `sha256:${"1".repeat(64)}`, toolchain_digest: `sha256:${"2".repeat(64)}` }));
        const runtime = createRuntime({ repoRoot: root, runRoot, runID, scratchRoot: path.join(temporaryRoot, "private") });
        try {
          const environment = { CARTULARY_TEST_RESULTS_DIR: temporaryRoot, CARTULARY_TEST_RUN_ID: runID, CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID };
          const staging = runtime.privatePath("staging");
          mkdirSync(staging, { mode: 0o700 });
          for (const entry of artifact.entries) writeFileSync(path.join(staging, entry), entry);
          const profile = { ...artifact, id: artifact === production ? "production" : "measurement" };
          const directory = sealFrontendArtifact({ repoRoot: root, runRoot, runtime, profile, staging });
          const sealed = resolveFrontendArtifact(root, artifact.producer_report_key, environment);
          const frontend = { build_artifact_ref: sealed.receiptRef, build_receipt_sha256: sealed.receiptDigest, build_artifact_sha256: sealed.receipt.content_digest };
          validateFrontendAttachment(root, artifact, frontend, environment);
          const other = artifact === production ? measurement : production;
          assert.throws(() => validateFrontendAttachment(root, other, frontend, environment), /does not match selected stage/u);
          assert.throws(() => validateFrontendAttachment(root, artifact, { ...frontend, build_artifact_sha256: `sha256:${"0".repeat(64)}` }, environment), /digest mismatch/u);
          writeFileSync(path.join(directory, "index.html"), "changed after publication");
          assert.throws(() => validateFrontendAttachment(root, artifact, frontend, environment), /digest mismatch/u);
          rmSync(path.join(directory, artifact.entries.at(-1)));
          assert.throws(() => validateFrontendAttachment(root, artifact, frontend, environment));
        } finally { runtime.close(); }
      }
    } finally {
      rmSync(temporaryRoot, { recursive: true, force: true });
    }
  },
));

function assertCaseDefinitions() {
  const allCases = Object.values(suiteCases).flat();
  assert.equal(new Set(allCases.map((entry) => entry.id)).size, allCases.length);
  for (const entry of allCases) {
    assert.match(entry.id, /^[a-z][a-z0-9_]*$/u);
    assert.ok(entry.acceptance_ids.length > 0);
    assert.equal(new Set(entry.acceptance_ids).size, entry.acceptance_ids.length);
    assert.ok(entry.acceptance_ids.every((id) => /^TH-HARNESS-AC-[0-9]{3}$/u.test(id)));
  }
}

export function runContractSuite(suite, { contextFactory = () => createContractTestContext(root) } = {}) {
  const cases = suiteCases[suite];
  assert.ok(cases?.length > 0, `contract suite ${suite} is empty`);
  if (suite === "graph") test("contract case definitions are unique and valid", assertCaseDefinitions);
  for (const entry of cases ?? []) {
    test(`${entry.id}: ${entry.name}`, async () => {
      await entry.run(contextFactory());
    });
  }
}
