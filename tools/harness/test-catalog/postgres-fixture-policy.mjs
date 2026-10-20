import { assertPostgresIsolation } from "./service-dependencies.mjs";
import { readFileSync } from "node:fs";
import path from "node:path";

import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";

const schemaID = "cartulary.postgres_fixture_policy_registry.v5";

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
  assertSortedUnique(registry.capability_policies.map(({ capability }) => capability),
    `${policyPath}.capability_policies`);
  const policies = new Map(registry.capability_policies.map((policy) => {
    assertSortedUnique(policy.cluster_scopes, `${policyPath}.${policy.capability}.cluster_scopes`);
    return [policy.capability, policy];
  }));

  const activeRows = rows.filter((row) => row.status === "active");
  for (const row of activeRows) {
    assertPostgresIsolation(row.postgres_isolation, row.service_dependencies, row.row_id);
    const policy = policies.get(row.fixture_capability);
    if (policy && !policy.cluster_scopes.includes(row.postgres_isolation)) {
      throw new Error(`${row.row_id} requires ${policy.cluster_scopes.join(" or ")} cluster isolation`);
    }
    if (row.fixture_capability === "browser_stack" && row.postgres_isolation !== "exclusive") {
      throw new Error(`${row.row_id} requires exclusive cluster isolation`);
    }
  }
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
    registry.capability_policies.map(({ capability }) => [
      capability,
      postgresRows.filter((row) => row.fixture_capability === capability).length,
    ]),
  );
  return { registry, counts, row_count: postgresRows.length };
}
