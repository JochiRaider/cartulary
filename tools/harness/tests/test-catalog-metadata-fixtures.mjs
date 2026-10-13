import assert from "node:assert/strict";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { loadTestCatalog } from "../test-catalog/index.mjs";
import { createContractTestContext } from "./contract-test-context.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

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
    assert.throws(() => loadTestCatalog(root), /TestLoadFixture is not cataloged exactly once/u);
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
    else topology.go_targets.raw_go_aggregates = [];
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
