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
