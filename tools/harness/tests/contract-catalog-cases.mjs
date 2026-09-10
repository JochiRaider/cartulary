import assert from "node:assert/strict";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { validatePostgresFixturePolicy } from "../test-catalog/postgres-fixture-policy.mjs";
import { groupRowsByPerformanceFixture, postgresMigrationDigest, snapshotKey } from "../performance-fixture/index.mjs";

export function assertPostgresCatalogClosure(context) {
  const { rows, postgresFixturePolicy: actual } = context.catalog;
  const active = rows.filter((row) => row.status === "active");
  const classified = new Set();
  for (const policy of actual.registry.capability_policies) {
    const members = active.filter((row) => row.fixture_capability === policy.capability);
    assert.equal(actual.counts[policy.capability], members.length, policy.capability);
    for (const row of members) {
      assert.equal(classified.has(row.row_id), false, row.row_id);
      classified.add(row.row_id);
    }
  }
  assert.deepEqual([...classified].sort(), active
    .filter((row) => row.fixture_capability.startsWith("postgres_"))
    .map((row) => row.row_id).sort());
  assert.equal(actual.row_count, classified.size);
  assert.deepEqual(actual.registry.transaction_row_approvals, active
    .filter((row) => row.fixture_capability === "postgres_transaction")
    .map((row) => row.row_id).sort());
}

export function assertPostgresPolicyFixtures(context) {
  const base = context.readJSON("tools/postgres_fixture_policy_registry.json");
  const directory = mkdtempSync(path.join(tmpdir(), "cartulary-postgres-policy."));
  const policyPath = path.join(directory, "policy.json");
  const row = (name, capability, status = "active") => ({
    row_id: `harness.test_catalog.fixture.${name}`,
    fixture_capability: capability,
    status,
    service_dependencies: capability === "none" ? [] : ["postgres"],
  });
  const transaction = row("transaction", "postgres_transaction");
  const rows = [
    row("dedicated", "postgres_dedicated"),
    row("migration", "postgres_migration"), transaction,
    row("inactive", "postgres_dedicated", "inactive"),
    row("browser", "browser_stack"), row("none", "none"),
  ];
  const validate = (input, approvals = [transaction.row_id]) => {
    writeFileSync(policyPath, JSON.stringify({ ...base, transaction_row_approvals: approvals }));
    return validatePostgresFixturePolicy(context.root, input, { policyPath });
  };
  try {
    assert.deepEqual(validate(rows).counts, {
      postgres_dedicated: 1, postgres_migration: 1, postgres_transaction: 1,
    });
    assert.equal(validate(rows).row_count, 3);
    const grown = [...rows, row("another_dedicated", "postgres_dedicated")];
    assert.deepEqual(validate(grown).counts, {
      postgres_dedicated: 2, postgres_migration: 1, postgres_transaction: 1,
    });
    assert.deepEqual(validate([], []).counts, {
      postgres_dedicated: 0, postgres_migration: 0, postgres_transaction: 0,
    });
    assert.throws(() => validate([{ ...rows[0], service_dependencies: [] }], []),
      /lacks the postgres service/u);
    assert.throws(() => validate(rows, []), /must exactly cover current transaction rows/u);
    assert.throws(() => validate(rows, ["harness.test_catalog.fixture.extra", transaction.row_id]),
      /must exactly cover current transaction rows/u);
    assert.throws(() => validate(rows, [transaction.row_id, transaction.row_id]),
      /duplicate|unique/u);
    const other = row("another_transaction", "postgres_transaction");
    assert.throws(() => validate([...rows, other], [transaction.row_id, other.row_id]),
      /ASCII-sorted/u);
    assert.deepEqual(validate([...rows, other], [other.row_id, transaction.row_id]).counts, {
      postgres_dedicated: 1, postgres_migration: 1, postgres_transaction: 2,
    });
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}

export function assertFixtureBuilderClosure(context) {
  const { catalog, fixtureBuilderPolicy } = context;
  const activeProfiles = [...catalog.fixtureProfiles.profiles.values()]
    .filter((profile) => profile.status === "active");
  assert.deepEqual([...fixtureBuilderPolicy.byFixtureProfileID.keys()].sort(),
    activeProfiles.map((profile) => profile.fixture_profile_id).sort());
  const groups = groupRowsByPerformanceFixture(context.root, catalog.rows, {
    registry: catalog.fixtureProfiles,
  });
  const migrationDigest = postgresMigrationDigest(context.root);
  // Each group and their union exercise O(groups) plans, independent of row count.
  const selections = [...groups.map((group) => [group]), ...(groups.length ? [groups] : [])];
  for (const selected of selections) {
    const graph = context.compiler.compile({ kind: "rows", row_ids: selected.flatMap((group) => group.row_ids) });
    const builders = graph.units.filter((unit) => unit.kind === "fixture_builder");
    assert.deepEqual(builders.map((unit) => unit.snapshot_key).sort(),
      selected.map((group) => snapshotKey(group.profile, migrationDigest)).sort());
    assert.equal(new Set(builders.map((unit) => unit.unit_id)).size, builders.length);
  }
  const unprofiled = catalog.rows.find((row) => !row.fixture_profile_id && row.fixture_capability === "none");
  assert.ok(unprofiled);
  const plain = context.compiler.compile({ kind: "rows", row_ids: [unprofiled.row_id] });
  assert.equal(plain.units.some((unit) => unit.kind === "fixture_builder"), false);
}
