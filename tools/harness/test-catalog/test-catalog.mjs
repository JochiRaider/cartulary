import {
  existsSync,
  lstatSync,
  readFileSync,
  readdirSync,
  realpathSync,
} from "node:fs";
import path from "node:path";

import {
  canonicalJSONString,
  parseStrictJSON,
  semanticJSONDigest,
  validateSchemaSync,
} from "../contract/index.mjs";
import { loadPerformanceFixtureSnapshotRegistry } from "../performance-fixture/index.mjs";
import { loadVerificationContracts } from "./verification-contracts.mjs";
import { assertSortedUnique, resolveRowSelector } from "./selector-resolution.mjs";
import { assertFixtureServiceDependencies } from "./service-dependencies.mjs";
import { validatePostgresFixturePolicy } from "./postgres-fixture-policy.mjs";

const ownerRegistrySchemaID = "cartulary.test_owner_registry.v1";
const familyManifestSchemaID = "cartulary.test_family_manifest.v5";
const rowMigrationSchemaID = "cartulary.test_catalog_row_migration.v1";
const runnerRegistrySchemaID = "cartulary.test_runner_registry.v1";
export const evidenceEpoch = "cartulary.test_evidence.nlspec.v1";
const expectedRunners = Object.freeze({
  go: "go_exact_tests",
  playwright: "playwright_exact_scenarios",
  shell: "shell_registered_command",
  vitest: "vitest_exact_titles",
});
const expectedRunnerDefinitions = Object.freeze({
  go: {
    runner: "go",
    selector_kind: "go_exact_tests",
    adapter_path: "tools/harness/execution/runners/go.mjs",
    approved_roots: ["cmd", "internal", "tools/testservices"],
  },
  playwright: {
    runner: "playwright",
    selector_kind: "playwright_exact_scenarios",
    adapter_path: "tools/harness/execution/runners/playwright.mjs",
    approved_roots: ["apps/web/e2e"],
    project_ids: ["chromium"],
    stages: ["accessibility", "measurement", "stateful", "support", "visual", "webserver_backed"],
  },
  shell: {
    runner: "shell",
    selector_kind: "shell_registered_command",
    adapter_path: "tools/harness/execution/runners/shell.mjs",
    approved_roots: [],
  },
  vitest: {
    runner: "vitest",
    selector_kind: "vitest_exact_titles",
    adapter_path: "tools/harness/execution/runners/vitest.mjs",
    approved_roots: ["apps/web", "packages"],
  },
});
const expectedProfiles = Object.freeze({
  runtime_profiles: ["default", "network_flow_claimed", "none"],
  resource_profiles: [
    "backend_capacity_isolated",
    "browser_functional",
    "browser_isolated",
    "browser_measurement_quiet",
    "io_heavy",
    "managed_process",
    "performance_fixture_builder",
    "standard",
  ],
});
function readStrictJSON(file) {
  return parseStrictJSON(readFileSync(file, "utf8"), file);
}

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function assertContainedRegularFile(root, relativeFile, expectedRoot, label) {
  if (path.isAbsolute(relativeFile) || relativeFile.includes("\\")) {
    throw new Error(`${label} must be a normalized repository-relative path`);
  }
  const normalized = path.posix.normalize(relativeFile);
  if (normalized !== relativeFile || normalized.startsWith("../")) {
    throw new Error(`${label} contains traversal or normalization drift`);
  }
  const candidate = path.resolve(root, relativeFile);
  if (!existsSync(candidate)) {
    throw new Error(`${label} does not exist: ${relativeFile}`);
  }
  const stat = lstatSync(candidate);
  if (!stat.isFile() || stat.isSymbolicLink()) {
    throw new Error(`${label} must reference a non-symlink regular file`);
  }
  const resolved = realpathSync(candidate);
  const containmentRoot = realpathSync(path.resolve(root, expectedRoot));
  if (
    resolved !== containmentRoot &&
    !resolved.startsWith(`${containmentRoot}${path.sep}`)
  ) {
    throw new Error(`${label} escapes ${expectedRoot}`);
  }
  return candidate;
}

function assertExactIDs(entries, expected, label) {
  const actual = entries.map((entry) => entry.id);
  assertSortedUnique(actual, label);
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(`${label} must declare exactly ${expected.join(", ")}`);
  }
}

function loadTopologyProfiles(root) {
  const topologyPath = path.join(root, "tools/execution_topology_manifest.json");
  const topology = readStrictJSON(topologyPath);
  validateSchemaSync(topology.schema_id, topology);
  for (const [key, ids] of Object.entries(expectedProfiles)) {
    if (!Array.isArray(topology[key])) {
      throw new Error(`${topologyPath}.${key} must be an array`);
    }
    assertExactIDs(topology[key], ids, `${topologyPath}.${key}`);
  }
  const resources = readStrictJSON(path.join(root, "tools/scheduler_resource_registry.json"));
  const knownResources = new Set(resources.resources.map((entry) => entry.name));
  const knownServices = new Set(Object.keys(topology.service_resource_minimums));
  for (const profile of topology.runtime_profiles) {
    assertSortedUnique(
      profile.managed_service_ids,
      `${topologyPath}.runtime_profiles.${profile.id}.managed_service_ids`,
    );
    for (const service of profile.managed_service_ids) {
      if (!knownServices.has(service)) {
        throw new Error(`${topologyPath}.runtime_profiles.${profile.id} has unknown service ${service}`);
      }
    }
  }
  for (const profile of topology.resource_profiles) {
    const claims = Object.keys(profile.resource_claims);
    assertSortedUnique(claims, `${topologyPath}.resource_profiles.${profile.id}.resource_claims`);
    for (const required of ["cpu", "io", "memory_mb", "process"]) {
      if (!Object.hasOwn(profile.resource_claims, required)) {
        throw new Error(
          `${topologyPath}.resource_profiles.${profile.id} must bound executable work with ${required}`,
        );
      }
    }
    for (const [resource, amount] of Object.entries(profile.resource_claims)) {
      if (!knownResources.has(resource) || !Number.isInteger(amount) || amount < 1) {
        throw new Error(`${topologyPath}.resource_profiles.${profile.id} has invalid claim ${resource}`);
      }
    }
  }
  for (const [service, minimums] of Object.entries(topology.service_resource_minimums)) {
    const resourcesForService = Object.keys(minimums);
    assertSortedUnique(resourcesForService, `${topologyPath}.service_resource_minimums.${service}`);
    for (const resource of resourcesForService) {
      if (!knownResources.has(resource)) {
        throw new Error(`${topologyPath}.service_resource_minimums.${service} has unknown resource ${resource}`);
      }
    }
  }
  return {
    runtimeIDs: new Set(topology.runtime_profiles.map((entry) => entry.id)),
    resourceIDs: new Set(topology.resource_profiles.map((entry) => entry.id)),
    semantic: {
      runtime_profiles: topology.runtime_profiles,
      resource_profiles: topology.resource_profiles,
      service_resource_minimums: topology.service_resource_minimums,
    },
  };
}

function loadRunnerRegistry(root) {
  const registryPath = path.join(root, "tools/test_runner_registry.json");
  const registry = readStrictJSON(registryPath);
  validateSchemaSync(runnerRegistrySchemaID, registry);
  const runnerIDs = registry.runners.map((entry) => entry.runner);
  assertSortedUnique(runnerIDs, `${registryPath}.runners.runner`);
  if (JSON.stringify(runnerIDs) !== JSON.stringify(Object.keys(expectedRunners).sort(asciiCompare))) {
    throw new Error(`${registryPath} must declare exactly the closed runner set`);
  }
  const byID = new Map();
  for (const runner of registry.runners) {
    if (runner.selector_kind !== expectedRunners[runner.runner]) {
      throw new Error(`${registryPath}.${runner.runner}.selector_kind is inconsistent`);
    }
    if (
      canonicalJSONString(runner) !==
      canonicalJSONString(expectedRunnerDefinitions[runner.runner])
    ) {
      throw new Error(`${registryPath}.${runner.runner} does not match the closed runner definition`);
    }
    assertSortedUnique(runner.approved_roots, `${registryPath}.${runner.runner}.approved_roots`);
    if (runner.project_ids) {
      assertSortedUnique(runner.project_ids, `${registryPath}.${runner.runner}.project_ids`);
    }
    if (runner.stages) {
      assertSortedUnique(runner.stages, `${registryPath}.${runner.runner}.stages`);
    }
    assertContainedRegularFile(
      root,
      runner.adapter_path,
      "tools/harness/execution/runners",
      `${registryPath}.${runner.runner}.adapter_path`,
    );
    byID.set(runner.runner, runner);
  }
  return { registry, byID };
}

export function validateFixtureProfile({ row, fixtureProfiles, label }) {
  const requiredBindings = row.verification_ids
    .map((verificationID) => fixtureProfiles.verificationBindings.get(verificationID))
    .filter(Boolean);
  const requiredProfileIDs = new Set(requiredBindings.map((entry) => entry.fixture_profile_id));
  if (requiredProfileIDs.size > 1) {
    throw new Error(label + ".verification_ids require incompatible fixture profiles");
  }
  if (requiredProfileIDs.size === 1 && !row.fixture_profile_id) {
    throw new Error(label + ".fixture_profile_id is required by its verification binding");
  }
  if (!row.fixture_profile_id) return;
  const profile = fixtureProfiles.profiles.get(row.fixture_profile_id);
  if (!profile || profile.status !== "active") {
    throw new Error(label + ".fixture_profile_id is unknown or inactive");
  }
  if (requiredProfileIDs.size === 0) {
    throw new Error(label + ".fixture_profile_id has no verification binding");
  }
  if (!requiredProfileIDs.has(row.fixture_profile_id)) {
    throw new Error(label + ".fixture_profile_id diverges from its verification binding");
  }
  const compatibility = profile.compatibility;
  const actual = {
    runner: row.runner,
    evidence_class: row.evidence_class,
    selector_stage: row.selector.stage,
    runtime_profile_id: row.runtime_profile_id,
    resource_profile_id: row.resource_profile_id,
    fixture_capability: row.fixture_capability,
    service_dependencies: row.service_dependencies,
  };
  if (canonicalJSONString(actual) !== canonicalJSONString(compatibility)) {
    throw new Error(label + ".fixture_profile_id is incompatible with the catalog row");
  }
}

function validateRowSemantics({
  row,
  manifest,
  verification,
  runners,
  profiles,
  fixtureProfiles,
  label,
}) {
  if (row.owner_id !== manifest.owner_id) {
    throw new Error(`${label}.owner_id must equal ${manifest.owner_id}`);
  }
  if (!row.family_id.startsWith(`${row.owner_id}.`)) {
    throw new Error(`${label}.family_id must be owner-qualified`);
  }
  if (!row.row_id.startsWith(`${row.family_id}.`)) {
    throw new Error(`${label}.row_id must be family-qualified`);
  }
  assertSortedUnique(row.collaborator_ids, `${label}.collaborator_ids`);
  assertSortedUnique(row.verification_ids, `${label}.verification_ids`);
  assertSortedUnique(row.service_dependencies, `${label}.service_dependencies`);
  for (const [field, values] of [
    ["tests", row.selector.tests],
    ["titles", row.selector.titles],
    ["scenario_ids", row.selector.scenario_ids],
  ]) {
    if (values) assertSortedUnique(values, `${label}.selector.${field}`);
  }
  if (row.collaborator_ids.includes(row.owner_id)) {
    throw new Error(`${label}.collaborator_ids must not repeat the primary owner`);
  }
  for (const verificationID of row.verification_ids) {
    const resolved = verification.verificationByID.get(verificationID);
    if (!resolved) {
      throw new Error(`${label}.verification_ids references unknown ${verificationID}`);
    }
    if (resolved.owner_id !== row.owner_id) {
      throw new Error(`${label}.verification_ids ${verificationID} is not owned by ${row.owner_id}`);
    }
    const expectedEvidence = {
      go: "go_test",
      playwright: "playwright",
      shell: "shell",
      vitest: "vitest",
    }[row.runner];
    if (!resolved.verification.evidence_kinds.includes(expectedEvidence)) {
      throw new Error(`${label}.runner ${row.runner} is unsupported by ${verificationID}`);
    }
    if (row.claim_posture.startsWith("claim.") && resolved.verification.profile !== row.claim_posture) {
      throw new Error(`${label}.claim_posture does not match ${verificationID}`);
    }
    if (row.claim_posture === "implementation" && resolved.verification.behavior_class === "claim_publication") {
      throw new Error(`${label} cannot use claim publication for implementation evidence`);
    }
  }
  if (
    row.evidence_class === "measurement" &&
    row.claim_posture === "informative" &&
    new Set(["fast", "standard"]).has(row.minimum_tier)
  ) {
    throw new Error(`${label} informative measurement must not enter fast or standard tiers`);
  }
  if (!runners.byID.has(row.runner)) {
    throw new Error(`${label}.runner is unsupported`);
  }
  if (!profiles.runtimeIDs.has(row.runtime_profile_id)) {
    throw new Error(`${label}.runtime_profile_id is unresolved`);
  }
  if (!profiles.resourceIDs.has(row.resource_profile_id)) {
    throw new Error(`${label}.resource_profile_id is unresolved`);
  }
  const runtimeProfile = profiles.semantic.runtime_profiles.find(
    (entry) => entry.id === row.runtime_profile_id,
  );
  const resourceProfile = profiles.semantic.resource_profiles.find(
    (entry) => entry.id === row.resource_profile_id,
  );
  for (const dependency of row.service_dependencies) {
    if (!runtimeProfile.managed_service_ids.includes(dependency)) {
      throw new Error(`${label}.service_dependencies ${dependency} is unavailable in ${row.runtime_profile_id}`);
    }
  }
  assertFixtureServiceDependencies(
    row.fixture_capability,
    row.service_dependencies,
    `${label}.fixture_capability`,
  );
  if (
    row.fixture_capability.startsWith("postgres_") &&
    !runtimeProfile.managed_service_ids.includes("postgres")
  ) {
    throw new Error(`${label}.fixture_capability requires a postgres runtime profile`);
  }
  if (row.runner === "playwright" && row.fixture_capability !== "browser_stack") {
    throw new Error(`${label}.fixture_capability must be browser_stack for Playwright`);
  }
  if (row.fixture_capability === "managed_process" && row.runtime_profile_id === "none") {
    throw new Error(`${label}.fixture_capability managed_process requires a runtime profile`);
  }
  if (
    Object.keys(resourceProfile.resource_claims).some((claim) => claim.startsWith("postgres")) &&
    !runtimeProfile.managed_service_ids.includes("postgres")
  ) {
    throw new Error(`${label}.resource_profile_id requires a postgres runtime profile`);
  }
  validateFixtureProfile({ row, fixtureProfiles, label });
}

export function loadTestCatalog(root) {
  const registryPath = path.join(root, "tools/test_catalog_owner.json");
  const registry = readStrictJSON(registryPath);
  validateSchemaSync(ownerRegistrySchemaID, registry);
  const ownerIDs = registry.owners.map((entry) => entry.owner_id);
  const manifestPaths = registry.owners.map((entry) => entry.manifest_path);
  assertSortedUnique(ownerIDs, `${registryPath}.owners.owner_id`);
  assertSortedUnique(manifestPaths, `${registryPath}.owners.manifest_path`);

  const verification = loadVerificationContracts(root);
  const knownOwnerIDs = new Set(verification.registry.owners.map((entry) => entry.owner_id));
  const runners = loadRunnerRegistry(root);
  const profiles = loadTopologyProfiles(root);
  const fixtureProfiles = loadPerformanceFixtureSnapshotRegistry(root);
  const taskSurface = readStrictJSON(path.join(root, "tools/task_surface_owner.json"));
  const taskSurfaceCommandIDs = new Set(taskSurface.targets.map((entry) => entry.command_id));
  const manifests = [];
  const rows = [];
  const familyIDs = new Set();
  const rowIDs = new Set();
  const selectors = new Set();

  for (const owner of registry.owners) {
    if (owner.manifest_path !== `tools/test_families/${owner.owner_id}.json`) {
      throw new Error(`${registryPath}:${owner.owner_id}.manifest_path must match its owner ID`);
    }
    const manifestFile = assertContainedRegularFile(
      root,
      owner.manifest_path,
      "tools/test_families",
      `${registryPath}:${owner.owner_id}.manifest_path`,
    );
    const manifest = readStrictJSON(manifestFile);
    validateSchemaSync(familyManifestSchemaID, manifest);
    if (manifest.owner_id !== owner.owner_id) {
      throw new Error(`${owner.manifest_path}.owner_id must equal ${owner.owner_id}`);
    }
    const manifestRowIDs = manifest.rows.map((row) => row.row_id);
    assertSortedUnique(manifestRowIDs, `${owner.manifest_path}.rows.row_id`);
    for (const [index, row] of manifest.rows.entries()) {
      const label = `${owner.manifest_path}.rows[${index + 1}]`;
      validateRowSemantics({
        row,
        manifest,
        verification,
        runners,
        profiles,
        fixtureProfiles,
        label,
      });
      if (rowIDs.has(row.row_id)) {
        throw new Error(`duplicate row_id ${row.row_id}`);
      }
      rowIDs.add(row.row_id);
      familyIDs.add(row.family_id);
      for (const collaboratorID of row.collaborator_ids) {
        if (!knownOwnerIDs.has(collaboratorID)) {
          throw new Error(`${label}.collaborator_ids references unknown ${collaboratorID}`);
        }
      }
      const resolvedSelectors = resolveRowSelector({
        root,
        row,
        runner: runners.byID.get(row.runner),
        taskSurfaceCommandIDs,
      });
      for (const selector of resolvedSelectors) {
        if (selectors.has(selector)) {
          throw new Error(`${label}.selector overlaps ${selector}`);
        }
        selectors.add(selector);
      }
      rows.push(row);
    }
    manifests.push(manifest);
  }

  const familyRoot = path.join(root, "tools/test_families");
  const registeredBasenames = new Set(manifestPaths.map((entry) => path.basename(entry)));
  const actualBasenames = readdirSync(familyRoot)
    .filter((entry) => entry.endsWith(".json"))
    .sort(asciiCompare);
  if (
    JSON.stringify(actualBasenames) !==
    JSON.stringify([...registeredBasenames].sort(asciiCompare))
  ) {
    throw new Error(`${familyRoot} contains an unregistered or missing owner manifest`);
  }

  const rowMigrations = readStrictJSON(
    path.join(root, "tools/test_catalog_row_migrations.json"),
  );
  validateSchemaSync(rowMigrationSchemaID, rowMigrations);
  const retiredRowIDs = rowMigrations.migrations.map((entry) => entry.retired_row_id);
  assertSortedUnique(retiredRowIDs, "tools/test_catalog_row_migrations.json.migrations.retired_row_id");
  for (const migration of rowMigrations.migrations) {
    assertSortedUnique(
      migration.replacement_row_ids,
      `${migration.retired_row_id}.replacement_row_ids`,
    );
    if (rowIDs.has(migration.retired_row_id)) {
      throw new Error(`retired catalog row remains active: ${migration.retired_row_id}`);
    }
    for (const replacement of migration.replacement_row_ids) {
      if (!rowIDs.has(replacement)) {
        throw new Error(`${migration.retired_row_id} has unknown replacement ${replacement}`);
      }
    }
  }

  const runnerCounts = Object.fromEntries(
    Object.keys(expectedRunners).sort(asciiCompare).map((runner) => [
      runner,
      rows.filter((row) => row.runner === runner).length,
    ]),
  );
  const postgresFixturePolicy = validatePostgresFixturePolicy(root, rows);
  for (const [verificationID, binding] of fixtureProfiles.verificationBindings) {
    const matchingRows = rows.filter((row) =>
      row.verification_ids.includes(verificationID) &&
      row.fixture_profile_id === binding.fixture_profile_id,
    );
    if (matchingRows.length !== 1) {
      throw new Error(
        "fixture verification binding " + verificationID +
        " must resolve to exactly one catalog row",
      );
    }
  }
  const catalogSemanticDigest = semanticJSONDigest({
    schema_id: registry.schema_id,
    owners: registry.owners.map((owner, index) => ({
      owner_id: owner.owner_id,
      manifest_path: owner.manifest_path,
      status: owner.status,
      manifest: {
        schema_id: manifests[index].schema_id,
        owner_id: manifests[index].owner_id,
        rows: manifests[index].rows,
      },
    })),
    runner_registry: runners.registry,
    profiles: profiles.semantic,
    fixture_profiles: fixtureProfiles.semantic_projection,
    postgres_fixture_policy: postgresFixturePolicy.registry,
    row_migrations: rowMigrations,
  });
  const summary = {
    schema_id: "cartulary.test_catalog_check_summary.v2",
    evidence_epoch: evidenceEpoch,
    status: "pass",
    owner_count: ownerIDs.length,
    family_count: familyIDs.size,
    row_count: rows.length,
    selector_count: selectors.size,
    runner_counts: runnerCounts,
    test_catalog_digest: catalogSemanticDigest,
    verification_routing_digest: verification.routing_digest,
  };
  validateSchemaSync(summary.schema_id, summary);
  const coveredVerificationIDs = new Set(
    rows.flatMap((row) => row.verification_ids),
  );
  for (const [verificationID, resolved] of verification.verificationByID) {
    if (
      !coveredVerificationIDs.has(verificationID) &&
      !resolved.verification.public_target
    ) {
      throw new Error(
        `active verification ${verificationID} has neither a test catalog row nor a public target`,
      );
    }
  }
  return {
    registry,
    manifests,
    rows,
    rowByID: new Map(rows.map((row) => [row.row_id, row])),
    verification,
    runners: runners.registry,
    profiles,
    fixtureProfiles,
    postgresFixturePolicy,
    rowMigrations,
    summary,
    test_catalog_digest: catalogSemanticDigest,
  };
}

export function validateTestCatalog(root) {
  return loadTestCatalog(root);
}
