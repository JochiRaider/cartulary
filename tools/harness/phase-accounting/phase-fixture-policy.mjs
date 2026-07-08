import { validPostgresFixtureReasonCodes } from "./phase-manifest-shape.mjs";
import {
  goEntrySymbols,
  supportGoEntryLabel,
  supportGoEntrySymbols,
} from "./phase-entry-evidence.mjs";

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
const validTemplateCloneReasonCodes = new Set([
  "committed_cross_connection_visibility",
  "database_identity",
  "process_lifecycle",
  "schema_mutation",
  "destructive_residue",
]);
const validGroupCloneReasonCodes = new Set(["shared_seeded_state"]);
const validPackageResetReasonCodes = new Set(["bounded_reset_surface"]);
const validMigrationScratchReasonCodes = new Set(["migration_scratch"]);
const postgresReasonFieldsByPolicy = new Map([
  [
    postgresFixturePolicyTemplateClone,
    new Set(["template_clone_reason", "template_clone_reason_code"]),
  ],
  [
    postgresFixturePolicyGroupClone,
    new Set(["group_clone_reason", "group_clone_reason_code"]),
  ],
  [
    postgresFixturePolicyPackageReset,
    new Set(["package_reset_reason", "package_reset_reason_code"]),
  ],
  [
    postgresFixturePolicyMigrationScratch,
    new Set(["migration_scratch_reason", "migration_scratch_reason_code"]),
  ],
  [postgresFixturePolicyTransaction, new Set()],
]);
const allPostgresReasonFields = new Set(
  [...postgresReasonFieldsByPolicy.values()].flatMap((fields) => [...fields]),
);
const fixtureProofKeys = new Set([
  "proof_kind",
  "proof_status",
  "proof_ref",
  "reason",
  "dirty_tables",
]);
const validFixtureProofStatuses = new Set(["accepted", "retained", "blocked"]);
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

function requireFixturePolicy(entry, label) {
  const policy = explicitPostgresFixturePolicy(entry, label);
  if (policy === "") {
    throw new Error(`${label} must declare fixture_policy.postgres`);
  }
  return policy;
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

function normalizeDirtyTables(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} dirty_tables must be an array`);
  }
  const seen = new Set();
  const dirtyTables = [];
  for (const table of value) {
    if (typeof table !== "string" || !/^[a-z][a-z0-9_]*$/.test(table)) {
      throw new Error(`${label} dirty_tables contains invalid table ${JSON.stringify(table)}`);
    }
    if (seen.has(table)) {
      throw new Error(`${label} dirty_tables contains duplicate ${table}`);
    }
    seen.add(table);
    dirtyTables.push(table);
  }
  return dirtyTables.sort();
}

function explicitFixtureProof(entry, policy, label) {
  if (entry.fixture_proof === undefined) {
    return null;
  }
  if (
    entry.fixture_proof === null ||
    Array.isArray(entry.fixture_proof) ||
    typeof entry.fixture_proof !== "object"
  ) {
    throw new Error(`${label} fixture_proof must be an object when present`);
  }
  const unexpected = Object.keys(entry.fixture_proof).filter((key) => !fixtureProofKeys.has(key));
  if (unexpected.length > 0) {
    throw new Error(`${label} fixture_proof has unsupported keys: ${unexpected.join(",")}`);
  }
  const proofKind = entry.fixture_proof.proof_kind;
  if (!validPostgresFixturePolicies.has(proofKind)) {
    throw new Error(`${label} fixture_proof.proof_kind must be a fixture policy token`);
  }
  if (proofKind !== policy) {
    throw new Error(`${label} fixture_proof.proof_kind must match fixture_policy.postgres`);
  }
  const proofStatus = entry.fixture_proof.proof_status;
  if (!validFixtureProofStatuses.has(proofStatus)) {
    throw new Error(`${label} fixture_proof.proof_status must be accepted|retained|blocked`);
  }
  const reason = entry.fixture_proof.reason;
  if (typeof reason !== "string" || reason.trim() === "") {
    throw new Error(`${label} fixture_proof.reason must be a non-empty string`);
  }
  const proof = {
    proof_kind: proofKind,
    proof_status: proofStatus,
    reason: reason.trim(),
  };
  if (entry.fixture_proof.proof_ref !== undefined) {
    if (typeof entry.fixture_proof.proof_ref !== "string" || entry.fixture_proof.proof_ref.trim() === "") {
      throw new Error(`${label} fixture_proof.proof_ref must be a non-empty string`);
    }
    proof.proof_ref = entry.fixture_proof.proof_ref.trim();
  }
  if (entry.fixture_proof.dirty_tables !== undefined) {
    proof.dirty_tables = normalizeDirtyTables(
      entry.fixture_proof.dirty_tables,
      `${label} fixture_proof`,
    );
  }
  return proof;
}

function postgresReasonDetails(entry, policy) {
  switch (policy) {
    case postgresFixturePolicyTemplateClone:
      return {
        reason: entry.template_clone_reason ?? "",
        reason_code: entry.template_clone_reason_code ?? "",
      };
    case postgresFixturePolicyGroupClone:
      return {
        reason: entry.group_clone_reason ?? "",
        reason_code: entry.group_clone_reason_code ?? "",
      };
    case postgresFixturePolicyPackageReset:
      return {
        reason: entry.package_reset_reason ?? "",
        reason_code: entry.package_reset_reason_code ?? "",
      };
    case postgresFixturePolicyMigrationScratch:
      return {
        reason: entry.migration_scratch_reason ?? "",
        reason_code: entry.migration_scratch_reason_code ?? "",
      };
    default:
      return {
        reason: "",
        reason_code: "",
      };
  }
}

function postgresFixtureDetail(policy, budget, entry, label) {
  const reason = postgresReasonDetails(entry, policy);
  const proof = explicitFixtureProof(entry, policy, label);
  return {
    fixture_policy: { postgres: policy },
    fixture_budget: { postgres: { ...budget } },
    ...(reason.reason ? { reason: reason.reason } : {}),
    ...(reason.reason_code ? { reason_code: reason.reason_code } : {}),
    ...(proof?.proof_ref ? { proof_ref: proof.proof_ref } : {}),
    ...(proof?.proof_status ? { proof_status: proof.proof_status } : {}),
    ...(proof?.proof_kind ? { proof_kind: proof.proof_kind } : {}),
    ...(proof?.reason ? { proof_reason: proof.reason } : {}),
    ...(proof?.dirty_tables ? { proof_dirty_tables: proof.dirty_tables } : {}),
  };
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

function validateFixtureContract(entry, symbols, policy, budget, label) {
  validatePostgresFixtureBudget(entry, policy, budget, label);
  validatePostgresFixtureReasonFieldScope(entry, policy, label);
  validateMigrationScratch(entry, symbols, policy, budget, label);
  validateTemplateCloneReason(entry, policy, label);
  validateGroupCloneReason(entry, policy, label);
  validatePackageResetReasonCode(entry, policy, label);
  explicitFixtureProof(entry, policy, label);
}

function overrideEntry(parentEntry, override) {
  const entry = {
    id: parentEntry.id,
    target: parentEntry.target,
    execution_dependency: parentEntry.execution_dependency,
  };
  for (const field of [
    "fixture_policy",
    "fixture_budget",
    "fixture_proof",
    ...allPostgresReasonFields,
  ]) {
    if (override[field] !== undefined) {
      entry[field] = override[field];
    }
  }
  return entry;
}

export function goEntrySymbolFixtureOverrides(entry) {
  if (entry.symbol_fixture_overrides === undefined) {
    return {};
  }
  const label = `manifest entry ${entry.id}`;
  if (
    entry.symbol_fixture_overrides === null ||
    Array.isArray(entry.symbol_fixture_overrides) ||
    typeof entry.symbol_fixture_overrides !== "object"
  ) {
    throw new Error(`${label} symbol_fixture_overrides must be an object`);
  }
  const symbols = goEntrySymbols(entry);
  const allowedSymbols = new Set(symbols);
  const result = {};
  for (const [symbol, override] of Object.entries(entry.symbol_fixture_overrides)) {
    const overrideLabel = `${label} symbol_fixture_overrides.${symbol}`;
    if (!allowedSymbols.has(symbol)) {
      throw new Error(`${overrideLabel} references undeclared symbol`);
    }
    if (override === null || Array.isArray(override) || typeof override !== "object") {
      throw new Error(`${overrideLabel} must be an object`);
    }
    const allowedOverrideKeys = new Set([
      "fixture_policy",
      "fixture_budget",
      "fixture_proof",
      ...allPostgresReasonFields,
    ]);
    const unexpected = Object.keys(override).filter((key) => !allowedOverrideKeys.has(key));
    if (unexpected.length > 0) {
      throw new Error(`${overrideLabel} has unsupported keys: ${unexpected.join(",")}`);
    }
    const normalized = overrideEntry(entry, override);
    const policy = requireFixturePolicy(normalized, overrideLabel);
    if (
      entry.execution_dependency === "backend_store" &&
      policy === postgresFixturePolicyPackageReset
    ) {
      throw new Error(`${overrideLabel} backend_store must not use fixture_policy.postgres=package_reset`);
    }
    const budget = explicitPostgresFixtureBudget(normalized, overrideLabel);
    validateFixtureContract(normalized, [symbol], policy, budget, overrideLabel);
    result[symbol] = { entry: normalized, policy, budget };
  }
  return result;
}

export function goEntrySymbolFixtureDetails(entry) {
  const symbols = goEntrySymbols(entry);
  const rowLabel = `manifest entry ${entry.id}`;
  const rowPolicy = goEntryPostgresFixturePolicy(entry);
  const rowBudget = goEntryPostgresFixtureBudget(entry);
  const overrides = goEntrySymbolFixtureOverrides(entry);
  const details = {};
  for (const symbol of symbols) {
    const override = overrides[symbol];
    if (override) {
      details[symbol] = postgresFixtureDetail(
        override.policy,
        override.budget,
        override.entry,
        `${rowLabel} symbol_fixture_overrides.${symbol}`,
      );
      continue;
    }
    details[symbol] = postgresFixtureDetail(rowPolicy, rowBudget, entry, rowLabel);
  }
  return details;
}

export function supportGoEntrySymbolFixtureDetails(entry) {
  const symbols = supportGoEntrySymbols(entry);
  const label = supportGoEntryLabel(entry);
  const policy = supportGoEntryPostgresFixturePolicy(entry);
  const budget = supportGoEntryPostgresFixtureBudget(entry);
  return Object.fromEntries(
    symbols.map((symbol) => [
      symbol,
      postgresFixtureDetail(policy, budget, entry, label),
    ]),
  );
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

export function validatePostgresFixtureReasonFieldScope(entry, policy, label) {
  const allowedFields = postgresReasonFieldsByPolicy.get(policy) ?? new Set();
  const unexpectedFields = [...allPostgresReasonFields].filter(
    (field) => !allowedFields.has(field) && entry[field] !== undefined,
  );
  if (unexpectedFields.length > 0) {
    const policyLabel = policy || "unspecified";
    throw new Error(
      `${label} fixture_policy.postgres=${policyLabel} must not declare ${unexpectedFields.join(", ")}`,
    );
  }
}

function validateReasonCodeForPolicy({
  label,
  policy,
  field,
  validReasonCodes,
  diagnosticPolicy,
}) {
  if (!validReasonCodes.has(field)) {
    throw new Error(
      `${label} ${diagnosticPolicy ?? policy} reason_code ${field} is not admissible for fixture_policy.postgres=${policy}`,
    );
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
  validateReasonCodeForPolicy({
    label,
    policy,
    field: entry.migration_scratch_reason_code,
    validReasonCodes: validMigrationScratchReasonCodes,
    diagnosticPolicy: "migration_scratch",
  });
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
  validateReasonCodeForPolicy({
    label,
    policy,
    field: entry.template_clone_reason_code,
    validReasonCodes: validTemplateCloneReasonCodes,
    diagnosticPolicy: "template_clone",
  });
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
  validateReasonCodeForPolicy({
    label,
    policy,
    field: entry.group_clone_reason_code,
    validReasonCodes: validGroupCloneReasonCodes,
    diagnosticPolicy: "group_clone",
  });
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
  validateReasonCodeForPolicy({
    label,
    policy,
    field: entry.package_reset_reason_code,
    validReasonCodes: validPackageResetReasonCodes,
    diagnosticPolicy: "package_reset",
  });
}

export function validatePostgresFixtureProof(entry, policy, label) {
  explicitFixtureProof(entry, policy, label);
}
