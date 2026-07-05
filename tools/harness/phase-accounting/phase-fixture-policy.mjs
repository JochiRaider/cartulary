import { validPostgresFixtureReasonCodes } from "./phase-manifest-shape.mjs";
import { supportGoEntryLabel } from "./phase-entry-evidence.mjs";

export const postgresFixturePolicyTemplateClone = "template_clone";
export const postgresFixturePolicyPackageReset = "package_reset";
export const postgresFixturePolicyMigrationScratch = "migration_scratch";
export const postgresFixturePolicyTransaction = "transaction";
export const postgresFixturePolicyGroupClone = "group_clone";

const validPostgresFixturePolicies = new Set([
  postgresFixturePolicyTemplateClone,
  postgresFixturePolicyPackageReset,
  postgresFixturePolicyMigrationScratch,
  postgresFixturePolicyTransaction,
  postgresFixturePolicyGroupClone,
]);
const validFixtureBudgetPostgresKeys = new Set([
  "max_template_clones",
  "max_group_clones",
  "max_package_resets",
  "max_transactions",
  "max_migration_scratch",
  "dirty_tables",
  "reset_conformance",
]);
function explicitPostgresFixturePolicy(entry, label) {
  if (entry.fixture_policy === undefined) {
    return "";
  }
  if (
    entry.fixture_policy === null ||
    Array.isArray(entry.fixture_policy) ||
    typeof entry.fixture_policy !== "object"
  ) {
    throw new Error(`${label} fixture_policy must be an object when present`);
  }
  const keys = Object.keys(entry.fixture_policy);
  const unexpected = keys.filter((key) => key !== "postgres");
  if (unexpected.length > 0) {
    throw new Error(`${label} fixture_policy has unsupported keys: ${unexpected.join(",")}`);
  }
  if (entry.fixture_policy.postgres === undefined) {
    return "";
  }
  if (!validPostgresFixturePolicies.has(entry.fixture_policy.postgres)) {
    throw new Error(
      `${label} fixture_policy.postgres must be template_clone|package_reset|migration_scratch|transaction|group_clone`,
    );
  }
  return entry.fixture_policy.postgres;
}

function explicitPostgresFixtureBudget(entry, label) {
  if (entry.fixture_budget === undefined) {
    return {};
  }
  if (
    entry.fixture_budget === null ||
    Array.isArray(entry.fixture_budget) ||
    typeof entry.fixture_budget !== "object"
  ) {
    throw new Error(`${label} fixture_budget must be an object when present`);
  }
  const keys = Object.keys(entry.fixture_budget);
  const unexpected = keys.filter((key) => key !== "postgres");
  if (unexpected.length > 0) {
    throw new Error(`${label} fixture_budget has unsupported keys: ${unexpected.join(",")}`);
  }
  if (entry.fixture_budget.postgres === undefined) {
    return {};
  }
  if (
    entry.fixture_budget.postgres === null ||
    Array.isArray(entry.fixture_budget.postgres) ||
    typeof entry.fixture_budget.postgres !== "object"
  ) {
    throw new Error(`${label} fixture_budget.postgres must be an object when present`);
  }
  const postgresKeys = Object.keys(entry.fixture_budget.postgres);
  const unexpectedPostgres = postgresKeys.filter(
    (key) => !validFixtureBudgetPostgresKeys.has(key),
  );
  if (unexpectedPostgres.length > 0) {
    throw new Error(
      `${label} fixture_budget.postgres has unsupported keys: ${unexpectedPostgres.join(",")}`,
    );
  }
  const budget = {};
  for (const key of [
    "max_template_clones",
    "max_group_clones",
    "max_package_resets",
    "max_transactions",
    "max_migration_scratch",
  ]) {
    if (entry.fixture_budget.postgres[key] === undefined) {
      continue;
    }
    const value = entry.fixture_budget.postgres[key];
    if (!Number.isInteger(value) || value < 0) {
      throw new Error(`${label} fixture_budget.postgres.${key} must be a non-negative integer`);
    }
    budget[key] = value;
  }
  if (entry.fixture_budget.postgres.dirty_tables !== undefined) {
    const dirtyTables = entry.fixture_budget.postgres.dirty_tables;
    if (!Array.isArray(dirtyTables) || dirtyTables.length === 0) {
      throw new Error(`${label} fixture_budget.postgres.dirty_tables must be a non-empty array`);
    }
    const seen = new Set();
    for (const table of dirtyTables) {
      if (typeof table !== "string" || !/^[a-z][a-z0-9_]*$/.test(table)) {
        throw new Error(
          `${label} fixture_budget.postgres.dirty_tables contains invalid table ${JSON.stringify(table)}`,
        );
      }
      if (seen.has(table)) {
        throw new Error(`${label} fixture_budget.postgres.dirty_tables contains duplicate ${table}`);
      }
      seen.add(table);
    }
    budget.dirty_tables = [...dirtyTables].sort();
  }
  if (entry.fixture_budget.postgres.reset_conformance !== undefined) {
    const resetConformance = entry.fixture_budget.postgres.reset_conformance;
    if (typeof resetConformance !== "boolean") {
      throw new Error(`${label} fixture_budget.postgres.reset_conformance must be a boolean`);
    }
    budget.reset_conformance = resetConformance;
  }
  return budget;
}

export function goEntryPostgresFixturePolicy(entry) {
  return explicitPostgresFixturePolicy(entry, `manifest entry ${entry.id}`);
}

export function supportGoEntryPostgresFixturePolicy(entry) {
  return explicitPostgresFixturePolicy(entry, supportGoEntryLabel(entry));
}

export function goEntryPostgresFixtureBudget(entry) {
  return explicitPostgresFixtureBudget(entry, `manifest entry ${entry.id}`);
}

export function supportGoEntryPostgresFixtureBudget(entry) {
  return explicitPostgresFixtureBudget(entry, supportGoEntryLabel(entry));
}

export function validatePostgresFixtureBudget(entry, policy, budget, label) {
  if (policy === postgresFixturePolicyPackageReset) {
    if (
      entry.fixture_policy?.postgres === postgresFixturePolicyPackageReset &&
      entry.fixture_budget?.postgres === undefined
    ) {
      throw new Error(`${label} explicit package_reset must declare fixture_budget.postgres`);
    }
    if (budget.max_package_resets === undefined) {
      throw new Error(`${label} package_reset must declare fixture_budget.postgres.max_package_resets`);
    }
    if (
      entry.fixture_policy?.postgres === postgresFixturePolicyPackageReset &&
      entry.fixture_budget?.postgres !== undefined
    ) {
      if (
        budget.reset_conformance !== true &&
        (!Array.isArray(budget.dirty_tables) || budget.dirty_tables.length === 0)
      ) {
        throw new Error(`${label} explicit package_reset budgets must declare fixture_budget.postgres.dirty_tables`);
      }
      if (
        budget.reset_conformance !== true &&
        (typeof entry.package_reset_reason !== "string" || entry.package_reset_reason.trim() === "")
      ) {
        throw new Error(`${label} explicit package_reset budgets must declare package_reset_reason`);
      }
    }
    return;
  }
  if (
    policy === postgresFixturePolicyTemplateClone &&
    entry.fixture_policy?.postgres === postgresFixturePolicyTemplateClone &&
    entry.fixture_budget?.postgres === undefined
  ) {
    throw new Error(`${label} explicit template_clone must declare fixture_budget.postgres`);
  }
  if (policy === postgresFixturePolicyTemplateClone && budget.max_template_clones === undefined) {
    throw new Error(`${label} template_clone must declare fixture_budget.postgres.max_template_clones`);
  }
  if (policy === postgresFixturePolicyGroupClone && budget.max_group_clones === undefined) {
    throw new Error(`${label} group_clone must declare fixture_budget.postgres.max_group_clones`);
  }
  if (policy === postgresFixturePolicyTransaction && budget.max_transactions === undefined) {
    throw new Error(`${label} transaction must declare fixture_budget.postgres.max_transactions`);
  }
  if (
    policy === postgresFixturePolicyMigrationScratch &&
    entry.fixture_budget?.postgres === undefined
  ) {
    throw new Error(`${label} migration_scratch must declare fixture_budget.postgres`);
  }
  if (
    policy === postgresFixturePolicyMigrationScratch &&
    budget.max_migration_scratch === undefined
  ) {
    throw new Error(`${label} migration_scratch must declare fixture_budget.postgres.max_migration_scratch`);
  }
}

export function validateMigrationScratch(entry, symbols, policy, budget, label) {
  if (policy !== postgresFixturePolicyMigrationScratch) {
    return;
  }
  if (
    typeof entry.migration_scratch_reason_code !== "string" ||
    !validPostgresFixtureReasonCodes.has(entry.migration_scratch_reason_code)
  ) {
    throw new Error(`${label} migration_scratch must declare closed migration_scratch_reason_code`);
  }
  if (
    typeof entry.migration_scratch_reason !== "string" ||
    entry.migration_scratch_reason.trim() === ""
  ) {
    throw new Error(`${label} migration_scratch must declare migration_scratch_reason`);
  }
  if (
    !/\b(backfill|boundary|migration|migrate|replay|upgrade)\b/i.test(
      entry.migration_scratch_reason,
    )
  ) {
    throw new Error(
      `${label} migration_scratch_reason must justify migration, boundary, replay, upgrade, or backfill coverage`,
    );
  }
  if (entry.target !== undefined && budget.max_migration_scratch > symbols.length) {
    throw new Error(
      `${label} migration_scratch budget must not exceed its support symbol count; split multi-database replay coverage into separate support symbols`,
    );
  }
}

export function validateTemplateCloneReason(entry, policy, label) {
  if (policy !== postgresFixturePolicyTemplateClone) {
    return;
  }
  if (entry.fixture_policy?.postgres !== postgresFixturePolicyTemplateClone) {
    return;
  }
  if (
    typeof entry.template_clone_reason_code !== "string" ||
    !validPostgresFixtureReasonCodes.has(entry.template_clone_reason_code)
  ) {
    throw new Error(`${label} explicit template_clone must declare closed template_clone_reason_code`);
  }
  if (entry.execution_dependency === "backend_process") {
    return;
  }
  if (typeof entry.template_clone_reason !== "string" || entry.template_clone_reason.trim() === "") {
    throw new Error(`${label} template_clone outside backend_process must declare template_clone_reason`);
  }
}

export function validateGroupCloneReason(entry, policy, label) {
  if (policy !== postgresFixturePolicyGroupClone) {
    return;
  }
  if (entry.fixture_policy?.postgres !== postgresFixturePolicyGroupClone) {
    return;
  }
  if (
    typeof entry.group_clone_reason_code !== "string" ||
    !validPostgresFixtureReasonCodes.has(entry.group_clone_reason_code)
  ) {
    throw new Error(`${label} explicit group_clone must declare closed group_clone_reason_code`);
  }
  if (typeof entry.group_clone_reason !== "string" || entry.group_clone_reason.trim() === "") {
    throw new Error(`${label} explicit group_clone must declare group_clone_reason`);
  }
}

export function validatePackageResetReasonCode(entry, policy, label) {
  if (policy !== postgresFixturePolicyPackageReset) {
    return;
  }
  if (entry.fixture_policy?.postgres !== postgresFixturePolicyPackageReset) {
    return;
  }
  if (
    typeof entry.package_reset_reason_code !== "string" ||
    !validPostgresFixtureReasonCodes.has(entry.package_reset_reason_code)
  ) {
    throw new Error(`${label} explicit package_reset must declare closed package_reset_reason_code`);
  }
}
