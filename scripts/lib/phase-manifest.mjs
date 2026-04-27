import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

const sectionDefinitions = [
  ["unit", "U-"],
  ["integration", "I-"],
  ["e2e", "E-"],
];
const implementationTestingGuidePath = path.join(
  "docs",
  "guides",
  "cartulary_implementation_testing_guide.md",
);

const validCoverage = new Set(["authoritative", "supplemental"]);
const validGoSections = new Set(["unit", "integration", "e2e"]);
const validExecutionDependencies = new Set([
  "backend_unit",
  "backend_store",
  "backend_integration",
  "backend_process",
  "frontend_unit",
  "browser_functional",
  "browser_stateful",
  "browser_measurement",
]);
const validSupportTargets = new Set(["backend_unit", "backend_integration_support"]);
const postgresFixturePolicyTemplateClone = "template_clone";
const postgresFixturePolicyPackageReset = "package_reset";
const postgresFixturePolicyMigrationScratch = "migration_scratch";
const postgresFixturePolicyTransaction = "transaction";
const postgresFixturePolicyGroupClone = "group_clone";
const validPostgresFixturePolicies = new Set([
  postgresFixturePolicyTemplateClone,
  postgresFixturePolicyPackageReset,
  postgresFixturePolicyMigrationScratch,
  postgresFixturePolicyTransaction,
  postgresFixturePolicyGroupClone,
]);
const postgresFixturePolicyEnvAssignable = new Set([
  postgresFixturePolicyTemplateClone,
  postgresFixturePolicyPackageReset,
  postgresFixturePolicyTransaction,
  postgresFixturePolicyGroupClone,
]);
const validFixtureBudgetPostgresKeys = new Set([
  "max_template_clones",
  "max_group_clones",
  "max_package_resets",
  "max_reset_duration_ms",
  "max_transactions",
  "max_migration_scratch",
  "dirty_tables",
]);
const defaultPackageResetBudget = Object.freeze({
  max_package_resets_per_symbol: 8,
  max_reset_duration_ms_per_symbol: 10000,
});
const defaultProcessTemplateCloneBudget = Object.freeze({
  max_template_clones_per_symbol: 4,
});
const defaultTransactionBudget = Object.freeze({
  max_transactions_per_symbol: 8,
});
const serviceBackedGoExecutionDependencies = new Set([
  "backend_store",
  "backend_integration",
  "backend_process",
]);
const serviceBackedSupportTargets = new Set(["backend_integration_support"]);
const supportTargetSections = new Map([
  ["backend_unit", "unit"],
  ["backend_integration_support", "integration"],
]);

export function goEntrySymbols(entry) {
  if (entry.symbol !== undefined && entry.symbols !== undefined) {
    throw new Error(`manifest entry ${entry.id} must declare symbol or symbols[], not both`);
  }
  if (entry.symbols !== undefined) {
    if (!Array.isArray(entry.symbols) || entry.symbols.length === 0) {
      throw new Error(`manifest entry ${entry.id} must declare a non-empty symbols[] array`);
    }
    for (const symbol of entry.symbols) {
      if (typeof symbol !== "string" || symbol.trim() === "") {
        throw new Error(`manifest entry ${entry.id} has an invalid symbol in symbols[]`);
      }
    }
    return entry.symbols;
  }
  if (typeof entry.symbol !== "string" || entry.symbol.trim() === "") {
    throw new Error(`manifest entry ${entry.id} is missing a non-empty symbol`);
  }
  return [entry.symbol];
}

function supportGoEntryLabel(entry) {
  return `support_go_target ${entry.target ?? "(missing target)"} ${entry.file ?? "(missing file)"}`;
}

function validateExecutionFamily(entry, label) {
  if (typeof entry.execution_family !== "string" || entry.execution_family.trim() === "") {
    throw new Error(`${label} must declare execution_family`);
  }
  if (!/^[a-z][a-z0-9-]*$/.test(entry.execution_family)) {
    throw new Error(`${label} execution_family must be a lowercase hyphenated identifier`);
  }
  if (typeof entry.execution_label !== "string" || entry.execution_label.trim() === "") {
    throw new Error(`${label} must declare execution_label`);
  }
}

export function supportGoEntrySymbols(entry) {
  const label = supportGoEntryLabel(entry);
  if (entry.symbol !== undefined && entry.symbols !== undefined) {
    throw new Error(`${label} must declare symbol or symbols[], not both`);
  }
  if (entry.symbols !== undefined) {
    if (!Array.isArray(entry.symbols) || entry.symbols.length === 0) {
      throw new Error(`${label} must declare a non-empty symbols[] array`);
    }
    for (const symbol of entry.symbols) {
      if (typeof symbol !== "string" || symbol.trim() === "") {
        throw new Error(`${label} has an invalid symbol in symbols[]`);
      }
    }
    return entry.symbols;
  }
  if (typeof entry.symbol !== "string" || entry.symbol.trim() === "") {
    throw new Error(`${label} is missing a non-empty symbol`);
  }
  return [entry.symbol];
}

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
    "max_reset_duration_ms",
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

function defaultGoPostgresFixturePolicy(entry) {
  if (entry.execution_dependency === "backend_store") {
    return postgresFixturePolicyTransaction;
  }
  if (entry.execution_dependency === "backend_integration") {
    return postgresFixturePolicyTemplateClone;
  }
  if (entry.execution_dependency === "backend_process") {
    return postgresFixturePolicyTemplateClone;
  }
  return "";
}

function defaultSupportPostgresFixturePolicy(entry) {
  if (entry.target === "backend_integration_support") {
    return postgresFixturePolicyTemplateClone;
  }
  return "";
}

export function effectiveGoEntryPostgresFixturePolicy(entry) {
  return goEntryPostgresFixturePolicy(entry) || defaultGoPostgresFixturePolicy(entry);
}

export function effectiveSupportGoEntryPostgresFixturePolicy(entry) {
  return supportGoEntryPostgresFixturePolicy(entry) || defaultSupportPostgresFixturePolicy(entry);
}

function symbolCount(symbols) {
  return Math.max(1, symbols.length);
}

function mergeDefaultPostgresBudget(explicitBudget, defaults) {
  return {
    ...defaults,
    ...explicitBudget,
  };
}

function defaultPostgresFixtureBudget(policy, symbols) {
  const count = symbolCount(symbols);
  switch (policy) {
    case postgresFixturePolicyPackageReset:
      return {
        max_package_resets: count * defaultPackageResetBudget.max_package_resets_per_symbol,
        max_reset_duration_ms:
          count * defaultPackageResetBudget.max_reset_duration_ms_per_symbol,
      };
    case postgresFixturePolicyTemplateClone:
      return {
        max_template_clones:
          count * defaultProcessTemplateCloneBudget.max_template_clones_per_symbol,
      };
    case postgresFixturePolicyTransaction:
      return {
        max_transactions: count * defaultTransactionBudget.max_transactions_per_symbol,
      };
    default:
      return {};
  }
}

export function effectiveGoEntryPostgresFixtureBudget(entry) {
  const policy = effectiveGoEntryPostgresFixturePolicy(entry);
  return mergeDefaultPostgresBudget(
    goEntryPostgresFixtureBudget(entry),
    defaultPostgresFixtureBudget(policy, goEntrySymbols(entry)),
  );
}

export function effectiveSupportGoEntryPostgresFixtureBudget(entry) {
  const policy = effectiveSupportGoEntryPostgresFixturePolicy(entry);
  return mergeDefaultPostgresBudget(
    supportGoEntryPostgresFixtureBudget(entry),
    defaultPostgresFixtureBudget(policy, supportGoEntrySymbols(entry)),
  );
}

function resetTableAssignments(entries, symbolsForEntry, budgetForEntry) {
  const assignments = [];
  for (const entry of entries) {
    const dirtyTables = budgetForEntry(entry).dirty_tables ?? [];
    if (dirtyTables.length === 0) {
      continue;
    }
    for (const symbol of symbolsForEntry(entry)) {
      assignments.push(`${symbol}=${dirtyTables.join("|")}`);
    }
  }
  return assignments.sort();
}

function validatePostgresFixtureBudget(entry, policy, budget, label) {
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
    if (budget.max_reset_duration_ms === undefined) {
      throw new Error(`${label} package_reset must declare fixture_budget.postgres.max_reset_duration_ms`);
    }
    if (
      entry.fixture_policy?.postgres === postgresFixturePolicyPackageReset &&
      entry.fixture_budget?.postgres !== undefined
    ) {
      if (!Array.isArray(budget.dirty_tables) || budget.dirty_tables.length === 0) {
        throw new Error(`${label} explicit package_reset budgets must declare fixture_budget.postgres.dirty_tables`);
      }
      if (typeof entry.package_reset_reason !== "string" || entry.package_reset_reason.trim() === "") {
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

function validateMigrationScratch(entry, symbols, policy, budget, label) {
  if (policy !== postgresFixturePolicyMigrationScratch) {
    return;
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

function validateTemplateCloneReason(entry, policy, label) {
  if (policy !== postgresFixturePolicyTemplateClone) {
    return;
  }
  if (entry.fixture_policy?.postgres !== postgresFixturePolicyTemplateClone) {
    return;
  }
  if (entry.execution_dependency === "backend_process") {
    return;
  }
  if (typeof entry.template_clone_reason !== "string" || entry.template_clone_reason.trim() === "") {
    throw new Error(`${label} template_clone outside backend_process must declare template_clone_reason`);
  }
}

function fixturePolicyAssignments(entries, symbolsForEntry, policyForEntry) {
  const assignments = [];
  for (const entry of entries) {
    const policy = policyForEntry(entry);
    if (!policy) {
      continue;
    }
    for (const symbol of symbolsForEntry(entry)) {
      if (!postgresFixturePolicyEnvAssignable.has(policy)) {
        continue;
      }
      assignments.push(`${symbol}=${policy}`);
    }
  }
  return assignments.sort();
}

function phaseNumberFromPhase(phase) {
  const match = /^phase(\d+)$/.exec(phase);
  if (!match) {
    throw new Error(`invalid phase name ${phase}; expected phase<number>`);
  }
  return match[1];
}

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`);
}

function phaseIDPatternSource(layerPrefix, phaseNumber, separator) {
  const normalizedLayerPrefix = layerPrefix.endsWith(separator)
    ? layerPrefix
    : `${layerPrefix}${separator}`;
  return `${escapeRegex(normalizedLayerPrefix)}${phaseNumber}${escapeRegex(
    separator,
  )}(?:[A-Z0-9]+${escapeRegex(separator)})*\\d{2}`;
}

function phaseIDRegex(layerPrefix, phaseNumber) {
  return new RegExp(`^${phaseIDPatternSource(layerPrefix, phaseNumber, "-")}$`);
}

function claimedPhaseIDRegex(phaseNumber, separator) {
  return new RegExp(
    String.raw`\b[UIE]${escapeRegex(separator)}${phaseNumber}${escapeRegex(
      separator,
    )}(?:[A-Z0-9]+${escapeRegex(separator)})*\d{2}\b`,
    "g",
  );
}

function validateExpectedIDs(expectedIDs, phaseNumber, manifestPath) {
  const seen = new Set();
  for (const id of expectedIDs) {
    if (typeof id !== "string" || id.trim() === "") {
      throw new Error(`manifest ${manifestPath} has an invalid expected_id: ${JSON.stringify(id)}`);
    }
    const layerPrefix = `${id[0] ?? ""}-`;
    if (!phaseIDRegex(layerPrefix, phaseNumber).test(id)) {
      throw new Error(`manifest ${manifestPath} has expected_id ${id} that does not belong to phase${phaseNumber}`);
    }
    if (seen.has(id)) {
      throw new Error(`manifest ${manifestPath} has duplicate expected_id ${id}`);
    }
    seen.add(id);
  }
}

function loadGuideExpectedIDs(root, phaseNumber) {
  const source = readFileSync(path.join(root, implementationTestingGuidePath), "utf8");
  const pattern = new RegExp(
    String.raw`^\|\s*([UIE]-${phaseNumber}(?:-[A-Z0-9]+)*-\d{2})\s*\|`,
  );
  const ids = new Set();
  for (const line of source.split(/\r?\n/)) {
    const match = pattern.exec(line);
    if (match?.[1]) {
      ids.add(match[1]);
    }
  }
  return Array.from(ids).sort();
}

function extractClaimedPhaseIDs(source, phaseNumber) {
  const hyphenMatches = source.match(claimedPhaseIDRegex(phaseNumber, "-")) ?? [];
  const underscoreMatches = source.match(claimedPhaseIDRegex(phaseNumber, "_")) ?? [];
  return new Set([
    ...hyphenMatches,
    ...underscoreMatches.map((value) => value.replaceAll("_", "-")),
  ]);
}

function collectGoTestFiles(root, relativeRoot) {
  const absoluteRoot = path.join(root, relativeRoot);
  const files = [];
  for (const entry of readdirSync(absoluteRoot, { withFileTypes: true })) {
    const relativePath = path.posix.join(relativeRoot, entry.name);
    if (entry.isDirectory()) {
      files.push(...collectGoTestFiles(root, relativePath));
      continue;
    }
    if (entry.isFile() && entry.name.endsWith("_test.go")) {
      files.push(relativePath);
    }
  }
  return files;
}

function collectAuthoritativeGoTestFunctions(root, phaseNumber) {
  const testFunctionPattern = new RegExp(
    String.raw`\bfunc\s+(TestPhase${phaseNumber}[A-Za-z0-9_]*_[UIE]_${phaseNumber}_(?:[A-Z0-9]+_)*\d{2})\s*\(`,
    "g",
  );
  const functions = [];
  for (const searchRoot of ["internal", path.posix.join("cmd", "server")]) {
    for (const file of collectGoTestFiles(root, searchRoot)) {
      const source = readFileSync(path.join(root, file), "utf8");
      for (const match of source.matchAll(testFunctionPattern)) {
        functions.push({ file, symbol: match[1] });
      }
    }
  }
  return functions.sort((left, right) => {
    if (left.symbol !== right.symbol) {
      return left.symbol.localeCompare(right.symbol);
    }
    return left.file.localeCompare(right.file);
  });
}

export function loadManifest(root, phase) {
  const manifestRoot = process.env.CARTULARY_PHASE_MANIFEST_ROOT
    ? path.resolve(process.env.CARTULARY_PHASE_MANIFEST_ROOT)
    : root;
  const manifestPath = path.join(manifestRoot, "tools", `${phase}_test_map.json`);
  const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
  return { manifestPath, manifest };
}

export function phaseManifestNames(root) {
  const manifestRoot = process.env.CARTULARY_PHASE_MANIFEST_ROOT
    ? path.resolve(process.env.CARTULARY_PHASE_MANIFEST_ROOT)
    : root;
  return readdirSync(path.join(manifestRoot, "tools"))
    .filter((entry) => /^phase\d+_test_map\.json$/.test(entry))
    .map((entry) => entry.replace(/_test_map\.json$/, ""))
    .sort((left, right) => {
      const leftNumber = Number.parseInt(left.replace(/^phase/, ""), 10);
      const rightNumber = Number.parseInt(right.replace(/^phase/, ""), 10);
      return leftNumber - rightNumber || left.localeCompare(right);
    });
}

export function collectEntries(manifest) {
  const entries = [];
  for (const [section] of sectionDefinitions) {
    for (const entry of manifest[section] ?? []) {
      entries.push({ ...entry, section });
    }
  }
  return entries;
}

export function collectSupportGoEntries(manifest) {
  return (manifest.support_go_targets ?? []).map((entry) => ({ ...entry }));
}

export function validateManifest(root, phase) {
  const phaseNumber = phaseNumberFromPhase(phase);
  const { manifestPath, manifest } = loadManifest(root, phase);
  const requireLedgerClaims =
    phase === "phase0" ||
    phase === "phase1" ||
    phase === "phase2" ||
    phase === "phase3" ||
    phase === "phase4";

  if (!Array.isArray(manifest.expected_ids) || manifest.expected_ids.length === 0) {
    throw new Error(`manifest ${manifestPath} must define a non-empty expected_ids array`);
  }
  validateExpectedIDs(manifest.expected_ids, phaseNumber, manifestPath);
  if (manifest.support_go_targets !== undefined && !Array.isArray(manifest.support_go_targets)) {
    throw new Error(`manifest ${manifestPath} support_go_targets must be an array when present`);
  }

  const entries = [];
  const supportEntries = [];
  for (const [section, prefix] of sectionDefinitions) {
    for (const entry of manifest[section] ?? []) {
      if (typeof entry.id !== "string" || !entry.id.startsWith(prefix)) {
        throw new Error(`manifest entry in ${section} has invalid id: ${JSON.stringify(entry)}`);
      }
      if (!phaseIDRegex(prefix, phaseNumber).test(entry.id)) {
        throw new Error(`manifest entry ${entry.id} does not belong to ${phase}`);
      }
      if (typeof entry.coverage !== "string" || !validCoverage.has(entry.coverage)) {
        throw new Error(`manifest entry ${entry.id} must declare coverage=authoritative|supplemental`);
      }
      if (
        typeof entry.execution_dependency === "string" &&
        !validExecutionDependencies.has(entry.execution_dependency)
      ) {
        throw new Error(
          `manifest entry ${entry.id} has invalid execution_dependency ${entry.execution_dependency}`,
        );
      }
      if (entry.runner === "playwright") {
        if (typeof entry.title !== "string" || entry.title.trim() === "") {
          throw new Error(`manifest entry ${entry.id} is missing a non-empty title`);
        }
        if (typeof entry.file !== "string" || !entry.file.startsWith("apps/web/e2e/")) {
          throw new Error(`manifest entry ${entry.id} must point at an apps/web/e2e file`);
        }
      } else if (entry.runner === "vitest") {
        if (typeof entry.title !== "string" || entry.title.trim() === "") {
          throw new Error(`manifest entry ${entry.id} is missing a non-empty title`);
        }
        if (typeof entry.file !== "string" || !entry.file.startsWith("apps/web/")) {
          throw new Error(`manifest entry ${entry.id} must point at an apps/web file`);
        }
      } else if (entry.runner === "go_test") {
        validateExecutionFamily(entry, `manifest entry ${entry.id}`);
        goEntrySymbols(entry);
        const postgresFixturePolicy = effectiveGoEntryPostgresFixturePolicy(entry);
        if (
          typeof entry.execution_dependency === "string" &&
          serviceBackedGoExecutionDependencies.has(entry.execution_dependency) &&
          postgresFixturePolicy === ""
        ) {
          throw new Error(
            `manifest entry ${entry.id} must declare fixture_policy.postgres for service-backed execution_dependency ${entry.execution_dependency}`,
          );
        }
        if (
          entry.execution_dependency === "backend_store" &&
          postgresFixturePolicy === postgresFixturePolicyPackageReset
        ) {
          throw new Error(
            `manifest entry ${entry.id} backend_store must not use fixture_policy.postgres=package_reset`,
          );
        }
        validatePostgresFixtureBudget(
          entry,
          postgresFixturePolicy,
          effectiveGoEntryPostgresFixtureBudget(entry),
          `manifest entry ${entry.id}`,
        );
        validateMigrationScratch(
          entry,
          goEntrySymbols(entry),
          postgresFixturePolicy,
          effectiveGoEntryPostgresFixtureBudget(entry),
          `manifest entry ${entry.id}`,
        );
        validateTemplateCloneReason(entry, postgresFixturePolicy, `manifest entry ${entry.id}`);
        if (typeof entry.package !== "string" || !entry.package.startsWith("./")) {
          throw new Error(`manifest entry ${entry.id} must declare a repo-relative Go package owner`);
        }
      } else {
        throw new Error(`manifest entry ${entry.id} must declare runner=go_test|playwright|vitest`);
      }
      if (typeof entry.file !== "string" || entry.file.trim() === "") {
        throw new Error(`manifest entry ${entry.id} must declare a file`);
      }
      if (typeof entry.evidence_layer !== "string" || entry.evidence_layer.trim() === "") {
        throw new Error(`manifest entry ${entry.id} must declare evidence_layer`);
      }
      if (requireLedgerClaims) {
        if (typeof entry.claim !== "string" || entry.claim.trim() === "") {
          throw new Error(`manifest entry ${entry.id} must declare a non-empty claim`);
        }
        if (typeof entry.out_of_scope !== "string" || entry.out_of_scope.trim() === "") {
          throw new Error(`manifest entry ${entry.id} must declare a non-empty out_of_scope`);
        }
      }
      entries.push({ ...entry, section });
    }
  }

  for (const entry of collectSupportGoEntries(manifest)) {
    if (typeof entry.target !== "string" || !validSupportTargets.has(entry.target)) {
      throw new Error(
        `${supportGoEntryLabel(entry)} must declare target=backend_unit|backend_integration_support`,
      );
    }
    if (typeof entry.section !== "string" || !validGoSections.has(entry.section)) {
      throw new Error(`${supportGoEntryLabel(entry)} must declare section=unit|integration|e2e`);
    }
    const expectedSection = supportTargetSections.get(entry.target);
    if (entry.section !== expectedSection) {
      throw new Error(
        `${supportGoEntryLabel(entry)} must declare section=${expectedSection} for target=${entry.target}`,
      );
    }
    if (typeof entry.package !== "string" || !entry.package.startsWith("./")) {
      throw new Error(`${supportGoEntryLabel(entry)} must declare a repo-relative Go package owner`);
    }
    if (typeof entry.file !== "string" || entry.file.trim() === "") {
      throw new Error(`${supportGoEntryLabel(entry)} must declare a file`);
    }
    if (typeof entry.selection_pattern !== "string" || entry.selection_pattern.trim() === "") {
      throw new Error(`${supportGoEntryLabel(entry)} must declare a non-empty selection_pattern`);
    }
    let selectionPattern;
    try {
      selectionPattern = new RegExp(entry.selection_pattern);
    } catch (error) {
      throw new Error(
        `${supportGoEntryLabel(entry)} has invalid selection_pattern ${JSON.stringify(entry.selection_pattern)}`,
      );
    }
    const symbols = supportGoEntrySymbols(entry);
    validateExecutionFamily(entry, supportGoEntryLabel(entry));
    const postgresFixturePolicy = effectiveSupportGoEntryPostgresFixturePolicy(entry);
    if (serviceBackedSupportTargets.has(entry.target) && postgresFixturePolicy === "") {
      throw new Error(
        `${supportGoEntryLabel(entry)} must declare fixture_policy.postgres for service-backed support target ${entry.target}`,
      );
    }
    const postgresFixtureBudget = effectiveSupportGoEntryPostgresFixtureBudget(entry);
    validatePostgresFixtureBudget(
      entry,
      postgresFixturePolicy,
      postgresFixtureBudget,
      supportGoEntryLabel(entry),
    );
    validateMigrationScratch(
      entry,
      symbols,
      postgresFixturePolicy,
      postgresFixtureBudget,
      supportGoEntryLabel(entry),
    );
    validateTemplateCloneReason(entry, postgresFixturePolicy, supportGoEntryLabel(entry));
    for (const symbol of symbols) {
      if (!selectionPattern.test(symbol)) {
        throw new Error(
          `${supportGoEntryLabel(entry)} selection_pattern does not match symbol ${symbol}`,
        );
      }
    }
    const packageRoot = entry.package.startsWith("./") ? entry.package.slice(2) : entry.package;
    const normalizedPackageRoot = packageRoot.endsWith("/...")
      ? packageRoot.slice(0, -4)
      : packageRoot;
    const fileDir = path.posix.dirname(entry.file);
    const packageOwnsFile = packageRoot.endsWith("/...")
      ? fileDir === normalizedPackageRoot || fileDir.startsWith(`${normalizedPackageRoot}/`)
      : fileDir === normalizedPackageRoot;
    if (!packageOwnsFile) {
      throw new Error(
        `${supportGoEntryLabel(entry)} file ${entry.file} does not belong to package ${entry.package}`,
      );
    }
    supportEntries.push(entry);
  }

  const ids = entries.map((entry) => entry.id);
  const uniqueIDs = new Set(ids);
  if (uniqueIDs.size !== ids.length) {
    throw new Error(`duplicate ids in ${manifestPath}`);
  }

  const authoritativeIDs = entries
    .filter((entry) => entry.coverage === "authoritative")
    .map((entry) => entry.id);
  const uniqueAuthoritativeIDs = new Set(authoritativeIDs);
  if (uniqueAuthoritativeIDs.size !== authoritativeIDs.length) {
    throw new Error(`duplicate authoritative ids in ${manifestPath}`);
  }

  const expected = manifest.expected_ids;
  const missing = expected.filter((id) => !uniqueAuthoritativeIDs.has(id));
  const unexpected = authoritativeIDs.filter((id) => !expected.includes(id));
  if (missing.length > 0 || unexpected.length > 0) {
    throw new Error(
      `${phase} manifest mismatch: missing=${missing.join(",") || "none"} unexpected=${unexpected.join(",") || "none"}`,
    );
  }

  const guideExpectedIDs = loadGuideExpectedIDs(root, phaseNumber);
  if (guideExpectedIDs.length > 0) {
    const guideMissing = guideExpectedIDs.filter((id) => !expected.includes(id));
    const guideUnexpected = expected.filter((id) => !guideExpectedIDs.includes(id));
    if (guideMissing.length > 0 || guideUnexpected.length > 0) {
      throw new Error(
        `${phase} guide mismatch: missing=${guideMissing.join(",") || "none"} unexpected=${guideUnexpected.join(",") || "none"}`,
      );
    }
  }

  for (const entry of entries) {
    const targetPath = path.join(root, entry.file);
    const source = readFileSync(targetPath, "utf8");
    const needles = entry.runner === "go_test" ? goEntrySymbols(entry) : [entry.title];
    for (const needle of needles) {
      if (!source.includes(needle)) {
        throw new Error(`manifest entry ${entry.id} not found in ${entry.file}: ${needle}`);
      }
    }
  }

  const manifestedGoSymbols = new Set();
  for (const entry of entries) {
    if (entry.runner !== "go_test") {
      continue;
    }
    for (const symbol of goEntrySymbols(entry)) {
      manifestedGoSymbols.add(symbol);
    }
  }
  const unmanifestedAuthoritativeGoTests = collectAuthoritativeGoTestFunctions(
    root,
    phaseNumber,
  ).filter((test) => !manifestedGoSymbols.has(test.symbol));
  if (unmanifestedAuthoritativeGoTests.length > 0) {
    throw new Error(
      `${phase} has authoritative-looking Go tests missing from ${manifestPath}: ${unmanifestedAuthoritativeGoTests
        .map((test) => `${test.file}::${test.symbol}`)
        .join(", ")}`,
    );
  }

  for (const entry of supportEntries) {
    const targetPath = path.join(root, entry.file);
    const source = readFileSync(targetPath, "utf8");
    for (const needle of supportGoEntrySymbols(entry)) {
      if (!source.includes(needle)) {
        throw new Error(`${supportGoEntryLabel(entry)} not found in ${entry.file}: ${needle}`);
      }
    }
  }

  for (const target of manifest.forbidden_id_files ?? []) {
    const source = readFileSync(path.join(root, target), "utf8");
    const claimedIDs = extractClaimedPhaseIDs(source, phaseNumber);
    if (claimedIDs.size > 0) {
      throw new Error(
        `${target} must not claim ${phase} authoritative ids: ${Array.from(claimedIDs).sort().join(", ")}`,
      );
    }
  }
}

function exactRegex(values) {
  if (values.length === 0) {
    throw new Error("cannot build an exact regex from an empty value list");
  }
  const escaped = values.map((value) => value.replace(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`));
  if (escaped.length === 1) {
    return `^${escaped[0]}$`;
  }
  return `^(${escaped.join("|")})$`;
}

function alternationRegex(values) {
  if (values.length === 0) {
    throw new Error("cannot build a regex from an empty value list");
  }
  const escaped = values.map((value) => value.replace(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`));
  if (escaped.length === 1) {
    return escaped[0];
  }
  return `(${escaped.join("|")})`;
}

export function packageMatchesPattern(pkg, pattern) {
  if (pattern.endsWith("/...")) {
    const prefix = pattern.slice(0, -4);
    return pkg === prefix || pkg.startsWith(`${prefix}/`);
  }
  return pkg === pattern;
}

function entryMatchesExecutionDependency(entry, executionDependency) {
  return executionDependency === "" || entry.execution_dependency === executionDependency;
}

function selectGoEntries(
  root,
  phase,
  section,
  coverage,
  executionDependency,
  executionFamily,
  packagePatterns,
) {
  if (!validGoSections.has(section)) {
    throw new Error(`invalid go manifest section ${section}`);
  }
  if (packagePatterns.length === 0) {
    throw new Error("go manifest selection requires at least one package pattern");
  }
  const { manifest } = loadManifest(root, phase);
  return collectEntries(manifest).filter(
    (entry) =>
      entry.section === section &&
      entry.runner === "go_test" &&
      entry.coverage === coverage &&
      entryMatchesExecutionDependency(entry, executionDependency) &&
      (executionFamily === "" || entry.execution_family === executionFamily) &&
      packagePatterns.some((pattern) => packageMatchesPattern(entry.package, pattern)),
  );
}

function selectSupportGoEntries(root, phase, target, executionFamily, packagePatterns) {
  if (!validSupportTargets.has(target)) {
    throw new Error(`invalid support target ${target}`);
  }
  if (packagePatterns.length === 0) {
    throw new Error("support go selection requires at least one package pattern");
  }
  const { manifest } = loadManifest(root, phase);
  return collectSupportGoEntries(manifest).filter(
    (entry) =>
      entry.target === target &&
      (executionFamily === "" || entry.execution_family === executionFamily) &&
      packagePatterns.some((pattern) => packageMatchesPattern(entry.package, pattern)),
  );
}

let cachedGoModulePath;

function loadGoModulePath(root) {
  if (cachedGoModulePath !== undefined) {
    return cachedGoModulePath;
  }
  const goMod = readFileSync(path.join(root, "go.mod"), "utf8");
  const match = goMod.match(/^module\s+(\S+)$/m);
  if (!match) {
    throw new Error("unable to determine Go module path from go.mod");
  }
  cachedGoModulePath = match[1];
  return cachedGoModulePath;
}

function toGoImportPath(root, repoRelativePackage) {
  if (!repoRelativePackage.startsWith("./")) {
    throw new Error(`manifest Go package must be repo-relative: ${repoRelativePackage}`);
  }
  const suffix = repoRelativePackage.slice(2);
  if (suffix === "") {
    return loadGoModulePath(root);
  }
  return `${loadGoModulePath(root)}/${suffix}`;
}

function goLogKey(pkg, test) {
  return `${pkg}::${test}`;
}

function describeGoSymbol(entry, symbol) {
  return `${symbol} [${entry.package}]`;
}

function readGoLogTopLevelStatuses(logFile) {
  const seen = new Map();
  for (const rawLine of readFileSync(logFile, "utf8").split(/\r?\n/)) {
    const line = rawLine.trim();
    if (line === "") {
      continue;
    }
    const entry = JSON.parse(line);
    if (typeof entry.Package !== "string" || entry.Package === "") {
      continue;
    }
    if (
      !["run", "pass", "fail", "skip"].includes(entry.Action) ||
      typeof entry.Test !== "string" ||
      entry.Test.includes("/")
    ) {
      continue;
    }
    if (
      !entry.Test.startsWith("Test") &&
      !entry.Test.startsWith("Benchmark") &&
      !entry.Test.startsWith("Fuzz")
    ) {
      continue;
    }
    const key = goLogKey(entry.Package, entry.Test);
    if (!seen.has(key)) {
      seen.set(key, { package: entry.Package, test: entry.Test, status: "" });
    }
    const current = seen.get(key);
    if (entry.Action === "run") {
      if (current.status === "") {
        current.status = "run";
      }
      continue;
    }
    current.status = entry.Action;
  }
  return seen;
}

function flattenPlaywrightSuites(suites, specs = []) {
  for (const suite of suites ?? []) {
    flattenPlaywrightSuites(suite.suites, specs);
    for (const spec of suite.specs ?? []) {
      specs.push(spec);
    }
  }
  return specs;
}

function selectPlaywrightEntries(root, phase, coverage, executionDependency) {
  const { manifest } = loadManifest(root, phase);
  return collectEntries(manifest).filter(
    (entry) =>
      entry.section === "e2e" &&
      entry.runner === "playwright" &&
      entry.coverage === coverage &&
      entryMatchesExecutionDependency(entry, executionDependency),
  );
}

function parsePlaywrightSelectionSpec(spec) {
  const [phase, coverage, executionDependency = ""] = spec.split(":");
  if (!phase || !coverage) {
    throw new Error(
      `invalid playwright selection ${spec}; expected <phase>:<coverage>[:<execution_dependency>]`,
    );
  }
  return { phase, coverage, executionDependency };
}

function selectPlaywrightEntriesForSpecs(root, specs) {
  if (specs.length === 0) {
    throw new Error("playwright multi-phase selection requires at least one selection spec");
  }
  return specs.flatMap((spec) => {
    const { phase, coverage, executionDependency } = parsePlaywrightSelectionSpec(spec);
    const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
    if (entries.length === 0) {
      throw new Error(`no ${coverage} playwright tests found for ${phase}`);
    }
    return entries;
  });
}

function normalizePlaywrightFile(file) {
  if (!file.startsWith("apps/web/")) {
    throw new Error(`playwright manifest file must live under apps/web/: ${file}`);
  }
  return file.slice("apps/web/".length);
}

function selectVitestEntries(root, phase, coverage, executionDependency) {
  const { manifest } = loadManifest(root, phase);
  return collectEntries(manifest).filter(
    (entry) =>
      entry.section === "unit" &&
      entry.runner === "vitest" &&
      entry.coverage === coverage &&
      entryMatchesExecutionDependency(entry, executionDependency),
  );
}

function selectVitestPhases(root, coverage, executionDependency) {
  return phaseManifestNames(root).filter((phase) => {
    const { manifest } = loadManifest(root, phase);
    return collectEntries(manifest).some(
      (entry) =>
        entry.section === "unit" &&
        entry.runner === "vitest" &&
        entry.coverage === coverage &&
        entryMatchesExecutionDependency(entry, executionDependency),
    );
  });
}

function normalizeVitestFile(file) {
  if (!file.startsWith("apps/web/")) {
    throw new Error(`vitest manifest file must live under apps/web/: ${file}`);
  }
  return file.slice("apps/web/".length);
}

function readVitestReport(reportFile) {
  return JSON.parse(readFileSync(reportFile, "utf8"));
}

function verifyVitestRun(reportFile, expectedTitles) {
  const report = readVitestReport(reportFile);
  if (report.success !== true) {
    throw new Error("vitest manifest run failed");
  }

  const executed = [];
  const failed = [];
  const files = new Set();
  for (const fileResult of report.testResults ?? []) {
    if (typeof fileResult?.name === "string" && fileResult.name !== "") {
      files.add(fileResult.name);
    }
    for (const assertion of fileResult.assertionResults ?? []) {
      if (assertion.status === "skipped") {
        continue;
      }
      if (typeof assertion.title !== "string" || assertion.title === "") {
        failed.push("(missing title)");
        continue;
      }
      if (assertion.status !== "passed") {
        failed.push(`${assertion.title} (${assertion.status})`);
        continue;
      }
      executed.push(assertion.title);
    }
  }

  const missing = expectedTitles.filter((title) => !executed.includes(title));
  const unexpected = executed.filter((title) => !expectedTitles.includes(title));
  if (missing.length > 0 || unexpected.length > 0 || failed.length > 0) {
    throw new Error(
      `vitest execution mismatch: missing=${missing.join(",") || "none"} unexpected=${unexpected.join(",") || "none"} failed=${failed.join(",") || "none"}`,
    );
  }

  return {
    files: Array.from(files).sort(),
    executed: executed.sort(),
  };
}

function readPlaywrightReport(reportFile) {
  return JSON.parse(readFileSync(reportFile, "utf8"));
}

function summarizePlaywrightErrors(report) {
  const messages = (report.errors ?? [])
    .map((error) => error?.message)
    .filter((message) => typeof message === "string" && message.trim() !== "");
  return messages.join("; ");
}

function detectPlaywrightSetupFailure(report) {
  const specs = flattenPlaywrightSuites(report.suites);
  if (specs.length > 0 || (report.errors ?? []).length === 0) {
    return null;
  }
  const summary = summarizePlaywrightErrors(report);
  return summary === "" ? "playwright setup failure" : `playwright setup failure: ${summary}`;
}

function verifyPlaywrightSpecSet(reportFile, expectedTitles) {
  const report = readPlaywrightReport(reportFile);
  const setupFailure = detectPlaywrightSetupFailure(report);
  if (setupFailure !== null) {
    throw new Error(setupFailure);
  }
  const actualTitles = flattenPlaywrightSuites(report.suites).map((spec) => spec.title);
  const missing = expectedTitles.filter((title) => !actualTitles.includes(title));
  const unexpected = actualTitles.filter((title) => !expectedTitles.includes(title));
  if (missing.length > 0 || unexpected.length > 0) {
    throw new Error(
      `playwright manifest mismatch: missing=${missing.join(",") || "none"} unexpected=${unexpected.join(",") || "none"}`,
    );
  }
  return report;
}

function extractPlaywrightStatuses(spec) {
  const statuses = [];
  for (const test of spec.tests ?? []) {
    if (Array.isArray(test.results) && test.results.length > 0) {
      for (const result of test.results) {
        if (typeof result.status === "string" && result.status !== "") {
          statuses.push(result.status);
        }
      }
      continue;
    }
    if (typeof test.status === "string" && test.status !== "" && test.status !== "skipped") {
      statuses.push(test.status);
    }
  }
  return statuses;
}

function printLines(lines) {
  process.stdout.write(`${lines.join("\n")}\n`);
}

function main(argv) {
  const [command, ...rest] = argv;
  const root = process.cwd();

  switch (command) {
    case "list-phases": {
      printLines(phaseManifestNames(root));
      return;
    }

    case "go-regex": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} go tests found for ${phase} ${section} in ${packagePatterns.join(", ")}`);
      }
      printLines([exactRegex(entries.flatMap((entry) => goEntrySymbols(entry)))]);
      return;
    }

    case "go-count": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      printLines([String(entries.length)]);
      return;
    }

    case "go-postgres-fixture-policy-tests": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      printLines([
        fixturePolicyAssignments(
          entries,
          goEntrySymbols,
          effectiveGoEntryPostgresFixturePolicy,
        ).join(","),
      ]);
      return;
    }

    case "go-postgres-reset-table-tests": {
      const [phase, section, coverage, executionDependency = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      printLines([resetTableAssignments(entries, goEntrySymbols, goEntryPostgresFixtureBudget).join(",")]);
      return;
    }

    case "go-family-regex": {
      const [phase, section, coverage, executionDependency = "", executionFamily = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(
        root,
        phase,
        section,
        coverage,
        executionDependency,
        executionFamily,
        packagePatterns,
      );
      if (entries.length === 0) {
        throw new Error(
          `no ${coverage} go tests found for ${phase} ${section} ${executionFamily} in ${packagePatterns.join(", ")}`,
        );
      }
      printLines([exactRegex(entries.flatMap((entry) => goEntrySymbols(entry)))]);
      return;
    }

    case "go-family-count": {
      const [phase, section, coverage, executionDependency = "", executionFamily = "", ...packagePatterns] = rest;
      const entries = selectGoEntries(
        root,
        phase,
        section,
        coverage,
        executionDependency,
        executionFamily,
        packagePatterns,
      );
      printLines([String(entries.length)]);
      return;
    }

    case "support-go-regex": {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no support go tests found for ${phase} ${target} in ${packagePatterns.join(", ")}`);
      }
      printLines([exactRegex(entries.flatMap((entry) => supportGoEntrySymbols(entry)))]);
      return;
    }

    case "support-go-count": {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      printLines([String(entries.length)]);
      return;
    }

    case "support-go-postgres-fixture-policy-tests": {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      printLines([
        fixturePolicyAssignments(
          entries,
          supportGoEntrySymbols,
          effectiveSupportGoEntryPostgresFixturePolicy,
        ).join(","),
      ]);
      return;
    }

    case "support-go-postgres-reset-table-tests": {
      const [phase, target, ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, "", packagePatterns);
      printLines([
        resetTableAssignments(
          entries,
          supportGoEntrySymbols,
          supportGoEntryPostgresFixtureBudget,
        ).join(","),
      ]);
      return;
    }

    case "support-go-family-regex": {
      const [phase, target, executionFamily = "", ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, executionFamily, packagePatterns);
      if (entries.length === 0) {
        throw new Error(
          `no support go tests found for ${phase} ${target} ${executionFamily} in ${packagePatterns.join(", ")}`,
        );
      }
      printLines([exactRegex(entries.flatMap((entry) => supportGoEntrySymbols(entry)))]);
      return;
    }

    case "support-go-family-count": {
      const [phase, target, executionFamily = "", ...packagePatterns] = rest;
      const entries = selectSupportGoEntries(root, phase, target, executionFamily, packagePatterns);
      printLines([String(entries.length)]);
      return;
    }

    case "go-verify-log": {
      const [phase, section, coverage, executionDependency = "", logFile, ...packagePatterns] = rest;
      const entries = selectGoEntries(root, phase, section, coverage, executionDependency, "", packagePatterns);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} go tests found for ${phase} ${section} in ${packagePatterns.join(", ")}`);
      }
      const actual = readGoLogTopLevelStatuses(logFile);
      const passed = [];
      const missing = [];
      const skipped = [];
      const failed = [];
      const incomplete = [];
      for (const entry of entries) {
        for (const symbol of goEntrySymbols(entry)) {
          const key = goLogKey(toGoImportPath(root, entry.package), symbol);
          const result = actual.get(key);
          if (!result) {
            missing.push(describeGoSymbol(entry, symbol));
            continue;
          }
          switch (result.status) {
            case "pass":
              passed.push(symbol);
              break;
            case "skip":
              skipped.push(describeGoSymbol(entry, symbol));
              break;
            case "fail":
              failed.push(describeGoSymbol(entry, symbol));
              break;
            default:
              incomplete.push(describeGoSymbol(entry, symbol));
              break;
          }
        }
      }
      if (missing.length > 0 || skipped.length > 0 || failed.length > 0 || incomplete.length > 0) {
        throw new Error(
          `manifest-go execution mismatch: missing=${missing.join(",") || "none"} skipped=${skipped.join(",") || "none"} failed=${failed.join(",") || "none"} incomplete=${incomplete.join(",") || "none"}`,
        );
      }
      printLines([
        `matched go manifest tests: ${passed.length}`,
        ...passed.sort().map((symbol) => `  ${symbol}`),
      ]);
      return;
    }

    case "playwright-files": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      const files = [...new Set(entries.map((entry) => normalizePlaywrightFile(entry.file)))].sort();
      printLines(files);
      return;
    }

    case "playwright-files-many": {
      const entries = selectPlaywrightEntriesForSpecs(root, rest);
      const files = [...new Set(entries.map((entry) => normalizePlaywrightFile(entry.file)))].sort();
      printLines(files);
      return;
    }

    case "playwright-grep": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      printLines([alternationRegex(entries.map((entry) => entry.title))]);
      return;
    }

    case "playwright-count": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      printLines([String(entries.length)]);
      return;
    }

    case "playwright-grep-many": {
      const entries = selectPlaywrightEntriesForSpecs(root, rest);
      printLines([alternationRegex(entries.map((entry) => entry.title))]);
      return;
    }

    case "playwright-selection-report": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} playwright tests found for ${phase}`);
      }
      const specs = entries.map((entry) => ({
        title: entry.title,
        file: normalizePlaywrightFile(entry.file),
        tests: [{ results: [], status: "skipped" }],
      }));
      process.stdout.write(
        `${JSON.stringify({ suites: [{ specs, suites: [] }], errors: [] }, null, 2)}\n`,
      );
      return;
    }

    case "playwright-verify-list": {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.map((entry) => entry.title).sort();
      verifyPlaywrightSpecSet(reportFile, expectedTitles);
      printLines([
        `listed playwright manifest tests: ${expectedTitles.length}`,
        ...expectedTitles.map((title) => `  ${title}`),
      ]);
      return;
    }

    case "playwright-verify-run": {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectPlaywrightEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.map((entry) => entry.title).sort();
      const report = verifyPlaywrightSpecSet(reportFile, expectedTitles);
      const specs = flattenPlaywrightSuites(report.suites);
      const failed = [];
      const executed = [];
      for (const expectedTitle of expectedTitles) {
        const spec = specs.find((candidate) => candidate.title === expectedTitle);
        if (!spec) {
          failed.push(`${expectedTitle} (not found)`);
          continue;
        }
        const statuses = extractPlaywrightStatuses(spec);
        if (statuses.length === 0) {
          failed.push(`${expectedTitle} (not executed)`);
          continue;
        }
        const acceptable = statuses.every((status) => status === "passed" || status === "flaky");
        if (!acceptable) {
          failed.push(`${expectedTitle} (${statuses.join(",")})`);
          continue;
        }
        executed.push(expectedTitle);
      }
      if (failed.length > 0) {
        throw new Error(`playwright execution mismatch: ${failed.join("; ")}`);
      }
      printLines([
        `matched playwright manifest tests: ${executed.length}`,
        ...executed.map((title) => `  ${title}`),
      ]);
      return;
    }

    case "vitest-files": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} vitest tests found for ${phase}`);
      }
      const files = [...new Set(entries.map((entry) => normalizeVitestFile(entry.file)))].sort();
      printLines(files);
      return;
    }

    case "vitest-phases": {
      const [coverage, executionDependency = ""] = rest;
      const phases = selectVitestPhases(root, coverage, executionDependency);
      if (phases.length === 0) {
        throw new Error(`no ${coverage} vitest phases found`);
      }
      printLines(phases);
      return;
    }

    case "vitest-grep": {
      const [phase, coverage, executionDependency = ""] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      if (entries.length === 0) {
        throw new Error(`no ${coverage} vitest tests found for ${phase}`);
      }
      printLines([`${alternationRegex(entries.map((entry) => entry.title))}$`]);
      return;
    }

    case "vitest-verify-run": {
      const [phase, coverage, executionDependency = "", reportFile] = rest;
      const entries = selectVitestEntries(root, phase, coverage, executionDependency);
      const expectedTitles = entries.map((entry) => entry.title).sort();
      const result = verifyVitestRun(reportFile, expectedTitles);
      printLines([
        `matched vitest manifest tests: ${result.executed.length}`,
        ...result.files.map((file) => `  file ${file}`),
        ...result.executed.map((title) => `  ${title}`),
      ]);
      return;
    }

    default:
      throw new Error(`unknown phase-manifest command ${command}`);
  }
}

if (import.meta.url === `file://${process.argv[1]}`) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exit(1);
  }
}
