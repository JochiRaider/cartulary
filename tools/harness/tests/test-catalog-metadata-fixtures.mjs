import assert from "node:assert/strict";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { goRunnerBuildContext, loadTestCatalog, resolvePlaywrightSelection, resolveVitestTitles, restrictedExecutableInputRoots } from "../test-catalog/index.mjs";
import { createContractTestContext } from "./contract-test-context.mjs";
import { loadCatalogMetadata, sourceExclusionPolicy } from "../test-support/owner-fixtures.mjs";
import { createScratch } from "../test-support/scratch.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

test("metadata admission rejects linked files and ancestors before reading through them", (t) => {
  const fixture = createScratch(t, "catalog-metadata-boundary-");
  const outside = createScratch(t, "catalog-metadata-external-");
  const read = fs.readFileSync;
  const inputs = new Set();
  let catalog;
  fs.readFileSync = (file, ...args) => {
    if (typeof file === "string" && file.startsWith(`${root}/`) && file.endsWith(".json")) inputs.add(file);
    return read(file, ...args);
  };
  syncBuiltinESMExports();
  try { catalog = loadCatalogMetadata(root); }
  finally { fs.readFileSync = read; syncBuiltinESMExports(); }
  // Copy only admitted machine inputs. Their contents construct the valid
  // fixture; the rejection expectations below are independently authored.
  inputs.add(path.join(root, "tools/performance_fixture_snapshot_owner.json"));
  for (const profile of catalog.fixtureProfiles.registry.profiles) {
    for (const ref of profile.source_contract_refs) inputs.add(path.join(root, ref.path));
  }
  for (const runner of catalog.runners.runners) inputs.add(path.join(root, runner.adapter_path));
  for (const input of inputs) {
    const destination = path.join(fixture, path.relative(root, input));
    fs.mkdirSync(path.dirname(destination), { recursive: true });
    fs.copyFileSync(input, destination);
  }
  assert.ok(loadCatalogMetadata(fixture).rowByID.has("harness.test_catalog.behavior.routing_contract"));

  for (const relative of [
    "tools/test_catalog_owner.json", "tools/test_runner_registry.json",
    "tools/execution_topology_manifest.json", "tools/scheduler_resource_registry.json",
    "tools/harness_work_graph_owner.json", "tools/test_catalog_row_migrations.json",
    "tools/task_surface_owner.json", "contracts/verification/registry.json",
    "tools/postgres_fixture_policy_registry.json",
    "tools/test_families/app.migrate.json", "contracts/verification/owners/app.migrate.json",
    "tools/harness/execution/runners/go.mjs", "tools/test_families",
    "contracts/verification/owners", "contracts/verification", "tools",
  ]) {
    const candidate = path.join(fixture, relative);
    const target = path.join(outside, "moved-input");
    const directory = fs.lstatSync(candidate).isDirectory();
    fs.renameSync(candidate, target);
    const originals = new Map(["readFileSync", "lstatSync", "statSync", "realpathSync"].map((name) => [name, fs[name]]));
    const forbiddenProbes = [];
    try {
      fs.symlinkSync(target, candidate, directory ? "dir" : "file");
      for (const [name, original] of originals) fs[name] = (file, ...args) => {
        if (typeof file === "string" && (
          file === target || file.startsWith(`${target}/`) || file.startsWith(`${candidate}/`) ||
          (file === candidate && name !== "lstatSync")
        )) forbiddenProbes.push(`${name}:${file}`);
        return original(file, ...args);
      };
      syncBuiltinESMExports();
      assert.throws(() => loadCatalogMetadata(fixture), directory ? /ancestor is not a directory/u : /non-symlink regular file/u, relative);
      assert.deepEqual(forbiddenProbes, [], relative);
    } finally {
      for (const [name, original] of originals) fs[name] = original;
      syncBuiltinESMExports();
      fs.rmSync(candidate, { force: true });
      fs.renameSync(target, candidate);
    }
  }

  const registryPath = path.join(fixture, "contracts/verification/registry.json");
  const registryBytes = fs.readFileSync(registryPath);
  const registry = JSON.parse(registryBytes);
  for (const contractPath of [
    "contracts/verification/owners/../app.migrate.json",
    "tools/test_families/app.migrate.json",
    "contracts/verification/owners/readme.json",
  ]) {
    fs.writeFileSync(registryPath, JSON.stringify({ ...registry, owners: [{ ...registry.owners[0], contract_path: contractPath }] }));
    const original = fs.lstatSync;
    const probes = [];
    fs.lstatSync = (file, ...args) => {
      if (String(file).startsWith(path.join(fixture, "contracts/verification/owners"))) probes.push(String(file));
      return original(file, ...args);
    };
    syncBuiltinESMExports();
    try {
      assert.throws(() => loadCatalogMetadata(fixture), /validation failed|documentation|human|prose/iu);
      assert.deepEqual(probes, [], contractPath);
    } finally {
      fs.lstatSync = original;
      syncBuiltinESMExports();
      fs.writeFileSync(registryPath, registryBytes);
    }
  }
  assert.ok(loadCatalogMetadata(fixture).rowByID.has("harness.test_catalog.behavior.routing_contract"));
});

test("support metadata uses the same strict admission as catalog owner inputs", (t) => {
  const fixture = createScratch(t, "catalog-support-metadata-");
  const outside = createScratch(t, "catalog-support-external-");
  fs.mkdirSync(path.join(fixture, "tools"));
  for (const name of ["toolchain_pins", "go_modules", "generated_artifact_policy", "executable_input_policy", "test_runner_registry"]) {
    fs.copyFileSync(path.join(root, `tools/${name}.json`), path.join(fixture, `tools/${name}.json`));
  }
  assert.equal(goRunnerBuildContext(fixture).environment.GOOS, "linux");
  assert.equal(sourceExclusionPolicy(fixture)("internal/gen/example.go"), true);
  assert.ok(restrictedExecutableInputRoots(fixture).length > 0);
  assert.deepEqual(resolveVitestTitles(fixture, []), new Map());
  assert.deepEqual(resolvePlaywrightSelection(fixture, []), []);

  for (const [name, consume] of [
    ["toolchain_pins", () => goRunnerBuildContext(fixture)],
    ["go_modules", () => goRunnerBuildContext(fixture)],
    ["generated_artifact_policy", () => sourceExclusionPolicy(fixture)],
    ["executable_input_policy", () => restrictedExecutableInputRoots(fixture)],
    ["test_runner_registry", () => resolveVitestTitles(fixture, [])],
    ["test_runner_registry", () => resolvePlaywrightSelection(fixture, [])],
  ]) {
    const file = path.join(fixture, `tools/${name}.json`);
    const original = fs.readFileSync(file);
    const target = path.join(outside, "metadata.json");
    fs.renameSync(file, target);
    const read = fs.readFileSync;
    const reads = [];
    try {
      fs.symlinkSync(target, file);
      fs.readFileSync = (input, ...args) => {
        if (input === file || input === target) reads.push(input);
        return read(input, ...args);
      };
      syncBuiltinESMExports();
      assert.throws(consume, /non-symlink regular file/u, name);
      assert.deepEqual(reads, [], name);
    } finally {
      fs.readFileSync = read;
      syncBuiltinESMExports();
      fs.rmSync(file, { force: true });
      fs.renameSync(target, file);
    }
    const [key, value] = Object.entries(JSON.parse(original))[0];
    try {
      fs.writeFileSync(file, original.toString().replace("{", `{${JSON.stringify(key)}:${JSON.stringify(value)},`));
      assert.throws(consume, /duplicate/iu, name);
    } finally { fs.writeFileSync(file, original); }
  }

  // Optional module metadata is absent only when the isolated root has no
  // independent module tree. A dangling link is invalid, never absence.
  const modules = path.join(fixture, "tools/go_modules.json");
  fs.rmSync(modules);
  assert.deepEqual(goRunnerBuildContext(fixture).modules, []);
  fs.symlinkSync(path.join(outside, "missing.json"), modules);
  assert.throws(() => goRunnerBuildContext(fixture), /non-symlink regular file/u);
  fs.rmSync(modules);
  fs.mkdirSync(path.join(fixture, "referencepacks"));
  assert.throws(() => goRunnerBuildContext(fixture), /missing independent Go module registry/u);

  fs.renameSync(path.join(fixture, "tools"), path.join(outside, "tools"));
  fs.symlinkSync(path.join(outside, "tools"), path.join(fixture, "tools"), "dir");
  assert.throws(() => goRunnerBuildContext(fixture), /ancestor is not a directory/u);
  assert.throws(() => sourceExclusionPolicy(fixture), /ancestor is not a directory/u);
  assert.throws(() => restrictedExecutableInputRoots(fixture), /ancestor is not a directory/u);
});

test("renderer context rejects missing browser capacity and unrelated service fixtures before selection", () => {
  const read = fs.readFileSync;
  const manifestPath = path.join(root, "tools/test_families/module.reporting.json");
  let mutation = {};
  fs.readFileSync = (file, ...args) => {
    const value = read(file, ...args);
    if (String(file) !== manifestPath) return value;
    const manifest = JSON.parse(value);
    const row = manifest.rows.find((entry) => entry.row_id === "module.reporting.integration.renderer_slidev");
    Object.assign(row, mutation);
    return JSON.stringify(manifest);
  };
  syncBuiltinESMExports();
  try {
    for (const entry of [
      { resource_profile_id: "standard" },
      { process_isolation: undefined },
      { runtime_profile_id: "default" },
      { evidence_class: "security" },
      { evidence_class: "unit" },
    ]) {
      mutation = entry;
      assert.throws(() => createContractTestContext(root).catalogMetadata, /renderer-node|installed renderer preparation/u);
    }
    mutation = {};
    assert.ok(createContractTestContext(root).catalogMetadata.rows.some((row) => row.row_id === "module.reporting.integration.renderer_slidev" && row.preparation_operations.includes("harness.build_report_renderer")));
  } finally {
    fs.readFileSync = read;
    syncBuiltinESMExports();
  }
});

test("catalog preparation is closed and Go browser preparation requires owned capacity", () => {
  const read = fs.readFileSync;
  const manifestPath = path.join(root, "tools/test_families/module.reporting.json");
  let mutation = {};
  let schemaID;
  fs.readFileSync = (file, ...args) => {
    const value = read(file, ...args);
    if (String(file) !== manifestPath) return value;
    const manifest = JSON.parse(value);
    if (schemaID) manifest.schema_id = schemaID;
    const row = manifest.rows.find((entry) => entry.row_id === "module.reporting.integration.renderer_stage_boundary");
    Object.assign(row, mutation);
    return JSON.stringify(manifest);
  };
  syncBuiltinESMExports();
  try {
    for (const entry of [
      { preparation_operations: ["arbitrary-command"] },
      { preparation_operations: ["harness.frontend_install", "harness.frontend_install"] },
      { preparation_operations: ["harness.playwright_install", "harness.frontend_install"] },
      { preparation_operations: ["harness.playwright_install"] },
      { preparation_operations: ["harness.build_report_renderer"], resource_profile_id: "standard" },
      { preparation_operations: ["harness.build_report_renderer"], process_isolation: undefined },
      { resource_profile_id: "standard" },
      { process_isolation: undefined },
    ]) {
      mutation = entry;
      assert.throws(() => createContractTestContext(root).catalogMetadata);
    }
    mutation = {};
    schemaID = "cartulary.test_family_manifest.v7";
    assert.throws(() => createContractTestContext(root).catalogMetadata);
  } finally {
    fs.readFileSync = read;
    syncBuiltinESMExports();
  }
});

test("metadata fixtures preserve current catalog facts without acquiring source selectors", () => {
  const read = fs.readFileSync;
  const ownerPath = path.join(root, "tools/test_catalog_owner.json");
  const owners = JSON.parse(read(ownerPath, "utf8"));
  const rows = owners.owners.flatMap((owner) => JSON.parse(read(path.join(root, owner.manifest_path), "utf8")).rows);
  const sourceReads = [];
  let rejectMetadata = false;
  fs.readFileSync = (file, ...args) => {
    const name = String(file);
    if (/\.(?:go|tsx?)$/u.test(name) && !name.includes("node_modules")) sourceReads.push(name);
    if (rejectMetadata && name === ownerPath) return '{"schema_id":"invalid"}';
    return read(file, ...args);
  };
  syncBuiltinESMExports();
  try {
    const complete = loadTestCatalog(root);
    assert.ok(sourceReads.some((name) => name.endsWith(".go")));
    assert.ok(sourceReads.some((name) => /\.tsx?$/u.test(name)));
    sourceReads.length = 0;
    const first = createContractTestContext(root);
    const metadata = first.catalogMetadata;
    assert.deepEqual(metadata.registry, owners);
    assert.deepEqual(metadata.rows, rows);
    assert.deepEqual(metadata.postgresFixturePolicy, complete.postgresFixturePolicy);
    assert.equal(metadata.test_catalog_digest, complete.test_catalog_digest);
    assert.equal(Object.hasOwn(metadata, "summary"), false);
    assert.deepEqual(sourceReads, []);
    assert.equal(first.catalogMetadata, metadata);
    assert.notEqual(createContractTestContext(root).catalogMetadata, metadata);

    // A fresh case revalidates current metadata; a completed case owns its
    // existing value. No global cache or successful partial-preflight summary.
    rejectMetadata = true;
    assert.equal(first.catalogMetadata, metadata);
    assert.throws(() => createContractTestContext(root).catalogMetadata, /schema|owners/u);
    rejectMetadata = false;
    assert.deepEqual(createContractTestContext(root).catalogMetadata.rows, rows);
    assert.deepEqual(sourceReads, []);
  } finally {
    fs.readFileSync = read;
    syncBuiltinESMExports();
  }
});

test("catalog package closure rejects a silently removed helper test", () => {
  const read = fs.readFileSync;
  const manifestPath = path.join(root, "tools/test_families/harness.command_surface.json");
  fs.readFileSync = (file, ...args) => {
    const value = read(file, ...args);
    if (String(file) !== manifestPath) return value;
    const manifest = JSON.parse(value);
    const row = manifest.rows.find((entry) => entry.row_id === "harness.command_surface.unit.configtest_helper_contract");
    row.selector.tests = row.selector.tests.filter((name) => name !== "TestLoadFixture");
    return JSON.stringify(manifest);
  };
  syncBuiltinESMExports();
  try {
    assert.throws(() => loadTestCatalog(root), /go:.\/internal\/testutil\/configtest:TestLoadFixture/u);
  } finally {
    fs.readFileSync = read;
    syncBuiltinESMExports();
  }
  assert.equal(loadTestCatalog(root).summary.status, "pass", "TestMain is not a missing catalog assertion");
});

test("catalog entry rejects legacy and hybrid topology before selector work", () => {
  const read = fs.readFileSync;
  const topologyPath = path.join(root, "tools/execution_topology_manifest.json");
  let legacy = true;
  fs.readFileSync = (file, ...args) => {
    const value = read(file, ...args);
    if (String(file) !== topologyPath) return value;
    const topology = JSON.parse(value);
    if (legacy) topology.schema_id = "cartulary.execution_topology.v9";
    else topology.go_targets = { family_runtime_binaries: [] };
    return JSON.stringify(topology);
  };
  syncBuiltinESMExports();
  try {
    assert.throws(() => loadTestCatalog(root), /validation failed/u);
    legacy = false;
    assert.throws(() => loadTestCatalog(root), /validation failed/u);
  } finally {
    fs.readFileSync = read;
    syncBuiltinESMExports();
  }
});


test("runtime binary declarations are row-owned, registered and sorted for process consumers", () => {
  const read = fs.readFileSync;
  const manifestPath = path.join(root, "tools/test_families/module.reporting.json");
  let ids = ["operator"], runner = "go";
  fs.readFileSync = (file, ...args) => {
    const value = read(file, ...args);
    if (String(file) !== manifestPath) return value;
    const manifest = JSON.parse(value);
    const row = manifest.rows.find((entry) => entry.runner === runner);
    assert.ok(row);
    row.runtime_binary_ids = ids;
    return JSON.stringify(manifest);
  };
  syncBuiltinESMExports();
  try {
    assert.ok(createContractTestContext(root).catalogMetadata.rows.some((row) => row.runtime_binary_ids?.includes("operator")));
    for (const invalid of [[], ["unknown"], ["operator", "operator"], ["server-harness", "operator"], ["INVALID"]]) {
      ids = invalid;
      assert.throws(() => createContractTestContext(root).catalogMetadata, /runtime_binary|runtime binary|ASCII|validation/u);
    }
    ids = ["operator"]; runner = "playwright";
    assert.ok(createContractTestContext(root).catalogMetadata.rows.some((row) => row.runner === "playwright" && row.runtime_binary_ids?.includes("operator")));
    runner = "vitest";
    assert.throws(() => createContractTestContext(root).catalogMetadata, /runtime_binary|validation/u);
  } finally {
    fs.readFileSync = read;
    syncBuiltinESMExports();
  }
});
