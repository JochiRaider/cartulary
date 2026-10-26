import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, renameSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import path from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";
import { validateSchemaSync } from "../../../contract/index.mjs";
import { captureCapabilitySnapshot, cpuCapacityWithSafetyMargin, assertScannerEvidenceParity, resolveVulnerabilityDatabaseRevision, CachePlan, resolveCacheDependencyClosure } from "../index.mjs";
import { buildSourceSnapshot } from "../../../test-catalog/index.mjs";
import { createContractTestContext } from "../../../tests/contract-test-context.mjs";
const root = path.resolve(import.meta.dirname, "../../../../..");
function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}
async function assertDependencyClosureContract(context) {
  const source = buildSourceSnapshot(root);
  const profile = context.cacheRegistry.profiles.find((entry) => entry.profile_id === "test_rows");
  const goRow = context.catalog.rows.find((row) => row.runner === "go" && row.fixture_capability === "none");
  const goUnit = cacheableRowUnit(context, goRow);
  const statefulRow = context.catalog.rows.find((row) => new Set(["go", "vitest"]).has(row.runner) &&
    (row.fixture_capability !== "none" || row.service_dependencies.length > 0));
  assert.ok(statefulRow, "the catalog must retain a stateful row cache boundary fixture");
  assert.equal(cacheableRowUnit(context, statefulRow).cache_policy, "none");
  const securityRow = context.catalog.rows.find((row) => new Set(["go", "vitest"]).has(row.runner) && row.evidence_class === "security");
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
  assert.ok(goClosure.entries.some((entry) => entry.path.startsWith("tools/harness/generated-artifacts/")), "generator identities must be closed by the package cache key");
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
  assert.equal(resolveCacheDependencyClosure({ root, entries: source.entries, profile, unit: goUnit }).digest, goClosure.digest, "deleting the untracked input must return to the original package closure");
  const renamedClosure = resolveCacheDependencyClosure({
    root,
    entries: [...source.entries, { ...added, path: `${goPackage}/renamed-cache-input.fixture` }]
      .sort((left, right) => compareASCII(left.path, right.path)),
    profile,
    unit: goUnit,
  });
  assert.notEqual(renamedClosure.digest, addedClosure.digest);
  for (const pathUnderTest of ["pnpm-lock.yaml", "tools/toolchain_pins.json"]) {
    const changed = source.entries.map((entry) => entry.path === pathUnderTest ? changedDigest(entry, pathUnderTest === "pnpm-lock.yaml" ? "2" : "3") : entry);
    const closure = resolveCacheDependencyClosure({ root, entries: changed, profile, unit: goUnit });
    assert.notEqual(closure.digest, goClosure.digest);
  }
  const schemaEntry = goClosure.entries.find((entry) => entry.path.startsWith("tools/schemas/"));
  const schemaChanged = source.entries.map((entry) => entry.path === schemaEntry.path ? changedDigest(entry, "4") : entry);
  assert.notEqual(resolveCacheDependencyClosure({ root, entries: schemaChanged, profile, unit: goUnit }).digest, goClosure.digest);
  const selectedGoSource = goClosure.entries.find((entry) => entry.path.startsWith(`${goPackage}/`) && entry.path.endsWith(".go"));
  const incomplete = source.entries.map((entry) => entry.path === selectedGoSource.path ? changedDigest(entry, "5") : entry);
  assert.equal(resolveCacheDependencyClosure({ root, entries: incomplete, profile, unit: goUnit }).strategy, "broad_fallback");
  const unknownUnit = {
    ...goUnit,
    command: {
      ...goUnit.command,
      environment: { ...goUnit.command.environment, CARTULARY_TEST_ROWS: "unknown.row" },
    },
  };
  assert.equal(resolveCacheDependencyClosure({ root, entries: source.entries, profile, unit: unknownUnit }).strategy, "broad_fallback");
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
  const producer = producerGraph.units.find((unit) => unit.unit_id === "operation:harness.release_inventory_artifacts");
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
  const profile = context.cacheRegistry.profiles.find((entry) => entry.profile_id === "release_artifacts");
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
    assert.ok(baselineClosure.entries.some((entry) => entry.path === pathUnderTest), `${pathUnderTest} must occur in the release inventory dependency closure`);
    const changed = source.entries.map((entry) => entry.path === pathUnderTest ? changedDigest(entry, "9") : entry);
    assert.notEqual(resolveCacheDependencyClosure({ root, entries: changed, profile, unit: producer }).digest, baselineClosure.digest, `${pathUnderTest} must change the release inventory cache key`);
  }
  for (const target of ["harness.license_report", "harness.sbom"]) {
    const graph = context.compiler.compile({ kind: "target", target });
    const validator = graph.units.find((unit) => unit.unit_id === `operation:${target}`);
    assert.equal(validator.cache_policy, "none");
    assert.ok(validator.needs.includes("operation:harness.release_inventory_artifacts"), `${target} must consume the canonical producer`);
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
    assert.throws(() => resolveVulnerabilityDatabaseRevision({
      root,
      database: fixtureRoot,
      declaredRevision: `sha256:${"f".repeat(64)}`,
    }), /does not match local content/u);
    assert.equal(resolveVulnerabilityDatabaseRevision({ root }).status, "unknown");
  }
  finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}
function assertScannerParityContract() {
  const evidence = { exitCode: 3, rawOutput: "finding\n", findings: "{\"finding\":1}\n" };
  assert.match(assertScannerEvidenceParity(evidence, evidence), /^sha256:[a-f0-9]{64}$/u);
  assert.throws(() => assertScannerEvidenceParity(evidence, { ...evidence, exitCode: 0 }), /diverge/u);
}
function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}
function changedDigest(entry, fill) {
  return { ...entry, byte_digest: `sha256:${fill.repeat(64)}` };
}
function cacheableRowUnit(context, row) {
  return context.compiler.compile({ kind: "rows", row_ids: [row.row_id] }).units.find((unit) => unit.current_run_evidence_outputs.includes(`rows/${row.row_id}.json`));
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
    if (stat.isDirectory())
      return snapshot(filename);
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
  }
  finally {
    rmSync(fixtureRoot, { recursive: true, force: true });
  }
}
export function registerBoundariesTests() {
  test("cache_registry_identity: cache registry exposes the current test-row profile", async () => {
    const context = createContractTestContext(root);
    assert.equal(context.cacheRegistry.schema_id, "cartulary.harness_cache_registry.v2");
    assert.ok(context.cacheRegistry.profiles.some((profile) => profile.profile_id === "test_rows"));
  });
}
export function registerSchedulerTests() {
  test("capacity_snapshot_contract: capacity snapshots enforce detected resource bounds", async () => {
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
      override: { schema_id: "cartulary.harness_capacity_override.v1", cpu_tokens: 2, port_lanes: 5 } }), /port_lanes=5 exceeds the detected policy bound 4/u);
    assert.throws(() => captureCapabilitySnapshot({
      root,
      override: {
        schema_id: "cartulary.harness_capacity_override.v1",
        cpu_tokens: 2,
        postgres_lanes: 3,
      },
    }), /postgres_lanes=3 exceeds the detected policy bound 2/u);
  });
  test("cache_registry_target_uniqueness: cache registry target bindings are unique", async () => {
    const context = createContractTestContext(root);
    assert.ok(context.cacheRegistry.profiles.every((profile) => new Set(profile.operation_ids).size === profile.operation_ids.length));
  });
  test("content_cache_security_and_closure: cache preparation preserves complete source dependency closure", async () => {
    const context = createContractTestContext(root);
    await assertDependencyClosureContract(context);
  });
  test("release_inventory_contract: release inventory has one deterministic paired cache producer and fresh validators", async () => {
    const context = createContractTestContext(root);
    assertReleaseInventoryContract(context);
  });
  test("vulnerability_revision_contract: vulnerability database revisions are content-proven", async () => {
    assertVulnerabilityRevisionContract();
  });
  test("scanner_evidence_parity: scanner evidence parity rejects divergent executions", async () => {
    assertScannerParityContract();
  });
}
