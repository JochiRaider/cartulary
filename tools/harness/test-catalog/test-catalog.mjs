import {
  existsSync,
  lstatSync,
  readFileSync,
  readdirSync,
  realpathSync,
} from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../contract/index.mjs";
import { loadVerificationContracts } from "./verification-contracts.mjs";
import { assertSortedUnique, resolveRowSelector } from "./selector-resolution.mjs";
import {
  canonicalJSONString,
  parseStrictJSON,
  semanticJSONDigest,
} from "./semantic-json.mjs";

const ownerRegistrySchemaID = "cartulary.test_owner_registry.v1";
const familyManifestSchemaID = "cartulary.test_family_manifest.v2";
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
    approved_roots: ["cmd", "internal"],
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
    "browser_exclusive",
    "go_balanced",
    "go_clone_heavy",
    "go_cpu_heavy",
    "go_io_heavy",
    "go_reset_heavy",
    "go_transaction_heavy",
    "none",
  ],
  fixture_profiles: [
    "none",
    "object_store_isolated",
    "postgres_group_clone",
    "postgres_migration_scratch",
    "postgres_package_reset",
    "postgres_template_clone",
    "postgres_transaction",
    "service_stack",
  ],
});
const expectedProfileDefinitions = Object.freeze({
  runtime_profiles: [
    { id: "default", managed_service_ids: ["object_store", "postgres"], browser_capable: true, startup_contract: "ordinary_unclaimed_configuration" },
    { id: "network_flow_claimed", managed_service_ids: ["object_store", "postgres"], browser_capable: true, startup_contract: "network_flow_claimed_configuration" },
    { id: "none", managed_service_ids: [], browser_capable: false, startup_contract: "no_managed_runtime" },
  ],
  resource_profiles: [
    { id: "browser_exclusive", resource_claims: { browser_stack: 1, process: 1 } },
    { id: "go_balanced", resource_claims: { go_cpu: 1, go_io: 1, process: 1 } },
    { id: "go_clone_heavy", resource_claims: { go_cpu: 1, go_io: 1, postgres: 1, postgres_clone: 1, process: 1 } },
    { id: "go_cpu_heavy", resource_claims: { go_cpu: 2, go_io: 1, process: 1 } },
    { id: "go_io_heavy", resource_claims: { go_cpu: 1, go_io: 2, process: 1 } },
    { id: "go_reset_heavy", resource_claims: { go_cpu: 1, go_io: 1, postgres: 1, postgres_reset: 1, process: 1 } },
    { id: "go_transaction_heavy", resource_claims: { go_cpu: 1, go_io: 1, postgres: 1, process: 1 } },
    { id: "none", resource_claims: {} },
  ],
  fixture_profiles: [
    { id: "none", fixture_kind: "none", isolation_scope: "none", budget: {} },
    { id: "object_store_isolated", fixture_kind: "object_store", isolation_scope: "row", budget: { max_buckets_or_prefixes: 1 } },
    { id: "postgres_group_clone", fixture_kind: "postgres", isolation_scope: "execution_group", budget: { max_group_clones: 1 } },
    { id: "postgres_migration_scratch", fixture_kind: "postgres", isolation_scope: "row", budget: { max_migration_scratch: 2 } },
    { id: "postgres_package_reset", fixture_kind: "postgres", isolation_scope: "package", budget: { max_resets: 4 } },
    { id: "postgres_template_clone", fixture_kind: "postgres", isolation_scope: "package", budget: { max_template_clones: 20 } },
    { id: "postgres_transaction", fixture_kind: "postgres", isolation_scope: "row", budget: { max_transactions: 32 } },
    { id: "service_stack", fixture_kind: "managed_services", isolation_scope: "execution_group", budget: { max_stacks: 1 } },
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
  for (const [key, ids] of Object.entries(expectedProfiles)) {
    if (!Array.isArray(topology[key])) {
      throw new Error(`${topologyPath}.${key} must be an array`);
    }
    assertExactIDs(topology[key], ids, `${topologyPath}.${key}`);
    if (canonicalProfileBytes(topology[key]) !== canonicalProfileBytes(expectedProfileDefinitions[key])) {
      throw new Error(`${topologyPath}.${key} does not match the closed profile definitions`);
    }
  }
  const resources = readStrictJSON(path.join(root, "tools/scheduler_resource_registry.json"));
  const knownResources = new Set([
    ...resources.resources.map((entry) => entry.name),
    ...resources.templates.map((entry) => entry.prefix),
  ]);
  for (const profile of topology.resource_profiles) {
    const claims = Object.keys(profile.resource_claims);
    assertSortedUnique(claims, `${topologyPath}.resource_profiles.${profile.id}.resource_claims`);
    for (const [resource, amount] of Object.entries(profile.resource_claims)) {
      if (!knownResources.has(resource) || !Number.isInteger(amount) || amount < 1) {
        throw new Error(`${topologyPath}.resource_profiles.${profile.id} has invalid claim ${resource}`);
      }
    }
  }
  return {
    runtimeIDs: new Set(topology.runtime_profiles.map((entry) => entry.id)),
    resourceIDs: new Set(topology.resource_profiles.map((entry) => entry.id)),
    fixtureIDs: new Set(topology.fixture_profiles.map((entry) => entry.id)),
    semantic: {
      runtime_profiles: topology.runtime_profiles,
      resource_profiles: topology.resource_profiles,
      fixture_profiles: topology.fixture_profiles,
    },
  };
}

function canonicalProfileBytes(value) {
  return canonicalJSONString(value);
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

function validateRowSemantics({ row, manifest, verification, runners, profiles, label }) {
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
  if (row.evidence_class === "measurement" && row.claim_posture === "informative" && row.default_check) {
    throw new Error(`${label} informative measurement must set default_check=false`);
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
  if (!profiles.fixtureIDs.has(row.fixture_profile_id)) {
    throw new Error(`${label}.fixture_profile_id is unresolved`);
  }
  const runtimeProfile = profiles.semantic.runtime_profiles.find(
    (entry) => entry.id === row.runtime_profile_id,
  );
  const fixtureProfile = profiles.semantic.fixture_profiles.find(
    (entry) => entry.id === row.fixture_profile_id,
  );
  const resourceProfile = profiles.semantic.resource_profiles.find(
    (entry) => entry.id === row.resource_profile_id,
  );
  if (
    fixtureProfile.fixture_kind === "postgres" &&
    !runtimeProfile.managed_service_ids.includes("postgres")
  ) {
    throw new Error(`${label}.fixture_profile_id requires a postgres runtime profile`);
  }
  if (
    Object.keys(resourceProfile.resource_claims).some((claim) => claim.startsWith("postgres")) &&
    !runtimeProfile.managed_service_ids.includes("postgres")
  ) {
    throw new Error(`${label}.resource_profile_id requires a postgres runtime profile`);
  }
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
      validateRowSemantics({ row, manifest, verification, runners, profiles, label });
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

  const runnerCounts = Object.fromEntries(
    Object.keys(expectedRunners).sort(asciiCompare).map((runner) => [
      runner,
      rows.filter((row) => row.runner === runner).length,
    ]),
  );
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
    summary,
    test_catalog_digest: catalogSemanticDigest,
  };
}

export function validateTestCatalog(root) {
  return loadTestCatalog(root);
}
