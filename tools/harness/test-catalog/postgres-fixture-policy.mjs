import { readFileSync } from "node:fs";
import path from "node:path";

import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";

const schemaID = "cartulary.postgres_fixture_policy_registry.v2";
const capabilityContract = Object.freeze([
  {
    capability: "postgres_dedicated",
    isolation_scope: "database",
    rationale: "committed_or_unproven_work_requires_unique_database",
    shared_scope: false,
  },
  {
    capability: "postgres_migration",
    isolation_scope: "migration_database",
    rationale: "schema_lifecycle_requires_fresh_database",
    shared_scope: false,
  },
  {
    capability: "postgres_transaction",
    isolation_scope: "transaction",
    rationale: "single_connection_rollback_proven",
    shared_scope: true,
  },
]);

function compareASCII(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function assertSortedUnique(values, label) {
  const sorted = [...values].sort(compareASCII);
  if (new Set(values).size !== values.length || JSON.stringify(values) !== JSON.stringify(sorted)) {
    throw new Error(`${label} must be ASCII-sorted and duplicate-free`);
  }
}

export function validatePostgresFixturePolicy(root, rows, options = {}) {
  const policyPath = path.resolve(
    root,
    options.policyPath ?? "tools/postgres_fixture_policy_registry.json",
  );
  const registry = parseStrictJSON(readFileSync(policyPath, "utf8"), policyPath);
  validateSchemaSync(schemaID, registry);
  if (JSON.stringify(registry.capability_policies) !== JSON.stringify(capabilityContract)) {
    throw new Error(`${policyPath}.capability_policies diverges from the adopted isolation contract`);
  }

  const activeRows = rows.filter((row) => row.status === "active");
  const postgresRows = activeRows.filter((row) => row.fixture_capability.startsWith("postgres_"));
  for (const row of postgresRows) {
    if (!row.service_dependencies.includes("postgres")) {
      throw new Error(`${row.row_id} PostgreSQL fixture classification lacks the postgres service`);
    }
  }

  const transactionRows = postgresRows
    .filter((row) => row.fixture_capability === "postgres_transaction")
    .map((row) => row.row_id)
    .sort(compareASCII);
  assertSortedUnique(registry.transaction_row_approvals, `${policyPath}.transaction_row_approvals`);
  if (JSON.stringify(registry.transaction_row_approvals) !== JSON.stringify(transactionRows)) {
    throw new Error(
      `${policyPath}.transaction_row_approvals must exactly cover current transaction rows`,
    );
  }

  const counts = Object.fromEntries(
    capabilityContract.map(({ capability }) => [
      capability,
      postgresRows.filter((row) => row.fixture_capability === capability).length,
    ]),
  );
  return { registry, counts, row_count: postgresRows.length };
}
