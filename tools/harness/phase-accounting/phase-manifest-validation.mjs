import { readFileSync } from "node:fs";
import path from "node:path";

import {
  serviceBackedGoExecutionDependencies,
  serviceBackedSupportTargets,
  validExecutionDependencies,
  validSupportTargets,
} from "../execution/execution-dependencies.mjs";
import { assertObjectKeys } from "../contract/json-shape.mjs";
import {
  sectionDefinitions,
  supportTargetSections,
  validClaimStatuses,
  validCoverage,
  validGoSections,
} from "./phase-manifest-constants.mjs";
import {
  assertAuthoritativeEvidenceNames,
  collectEntries,
  collectSupportGoEntries,
  goEntryScenarioSymbols,
  goEntrySymbols,
  playwrightEntryTitles,
  supportGoEntryLabel,
  supportGoEntrySymbols,
  vitestEntryTitles,
} from "./phase-entry-evidence.mjs";
import {
  goEntryPostgresFixtureBudget,
  goEntryPostgresFixturePolicy,
  postgresFixturePolicyPackageReset,
  supportGoEntryPostgresFixtureBudget,
  supportGoEntryPostgresFixturePolicy,
  validateGroupCloneReason,
  validateMigrationScratch,
  validatePackageResetReasonCode,
  validatePostgresFixtureReasonFieldScope,
  validatePostgresFixtureBudget,
  validateTemplateCloneReason,
} from "./phase-fixture-policy.mjs";
import {
  extractClaimedPhaseIDs,
  loadGuideExpectedIDs,
  phaseIDRegex,
  validateExpectedIDs,
} from "./phase-manifest-id-validation.mjs";
import { validateProfileClaims } from "./phase-manifest-profile-claims.mjs";
import {
  runtimeBinaries,
  validateEvidencePlacement,
  validateExecutionFamily,
} from "./phase-manifest-runner-validation.mjs";
import {
  collectAuthoritativeGoTestFunctions,
} from "./phase-manifest-source-scan.mjs";
import {
  phaseLedgerKeys,
  phaseManifestEntryKeys,
  supportGoEntryKeys,
} from "./phase-manifest-shape.mjs";
import {
  assertAuthoritativeGridRowsUseLiveAdapter,
  validateFixtureRefs,
  validateFrontendFixtureRefs,
} from "./phase-frontend-fixtures.mjs";
import { loadManifest, phaseNumberFromPhase } from "./phase-manifest-loader.mjs";

export function validateManifest(root, phase, { allowPlanned = false } = {}) {
  const { manifestPath, manifest } = loadManifest(root, phase, { allowPlanned });
  const phaseName = manifest.phase;
  const phaseNumber = phaseNumberFromPhase(phaseName);

  if (!Array.isArray(manifest.expected_ids) || manifest.expected_ids.length === 0) {
    throw new Error(`manifest ${manifestPath} must define a non-empty expected_ids array`);
  }
  validateExpectedIDs(manifest.expected_ids, phaseNumber, manifestPath);
  if (manifest.support_go_targets !== undefined && !Array.isArray(manifest.support_go_targets)) {
    throw new Error(`manifest ${manifestPath} support_go_targets must be an array when present`);
  }

  const entries = [];
  const supportEntries = [];
  if (manifest.ledger !== undefined) {
    if (!manifest.ledger || typeof manifest.ledger !== "object" || Array.isArray(manifest.ledger)) {
      throw new Error(`manifest ${manifestPath} ledger must be an object`);
    }
    assertObjectKeys(manifest.ledger, phaseLedgerKeys, `manifest ${manifestPath}.ledger`);
  }
  for (const [section, prefix] of sectionDefinitions) {
    for (const entry of manifest[section] ?? []) {
      assertObjectKeys(entry, phaseManifestEntryKeys, `manifest entry ${entry.id ?? "(missing id)"}`);
      if (typeof entry.id !== "string" || !entry.id.startsWith(prefix)) {
        throw new Error(`manifest entry in ${section} has invalid id: ${JSON.stringify(entry)}`);
      }
      if (!phaseIDRegex(prefix, phaseNumber).test(entry.id)) {
        throw new Error(`manifest entry ${entry.id} does not belong to ${phaseName}`);
      }
      if (typeof entry.coverage !== "string" || !validCoverage.has(entry.coverage)) {
        throw new Error(`manifest entry ${entry.id} must declare coverage=authoritative|supplemental`);
      }
      validateEvidencePlacement(entry, `manifest entry ${entry.id}`);
      if (
        typeof entry.execution_dependency === "string" &&
        !validExecutionDependencies.has(entry.execution_dependency)
      ) {
        throw new Error(
          `manifest entry ${entry.id} has invalid execution_dependency ${entry.execution_dependency}`,
        );
      }
      if (entry.claim_status !== undefined && !validClaimStatuses.has(entry.claim_status)) {
        throw new Error(
          `manifest entry ${entry.id} claim_status must be implemented|blocked|not_applicable`,
        );
      }
      if (entry.runner === "playwright") {
        playwrightEntryTitles(entry);
        if (typeof entry.file !== "string" || !entry.file.startsWith("apps/web/e2e/")) {
          throw new Error(`manifest entry ${entry.id} must point at an apps/web/e2e file`);
        }
      } else if (entry.runner === "vitest") {
        vitestEntryTitles(entry);
        if (typeof entry.file !== "string" || !entry.file.startsWith("apps/web/")) {
          throw new Error(`manifest entry ${entry.id} must point at an apps/web file`);
        }
      } else if (entry.runner === "go_test") {
        if (entry.title !== undefined || entry.titles !== undefined) {
          throw new Error(`manifest entry ${entry.id} must declare symbol or symbols[], not title metadata`);
        }
        validateExecutionFamily(entry, `manifest entry ${entry.id}`);
        runtimeBinaries(entry, `manifest entry ${entry.id}`);
        goEntrySymbols(entry);
        goEntryScenarioSymbols(entry);
        const postgresFixturePolicy = goEntryPostgresFixturePolicy(entry);
        const postgresFixtureBudget = goEntryPostgresFixtureBudget(entry);
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
          postgresFixtureBudget,
          `manifest entry ${entry.id}`,
        );
        validatePostgresFixtureReasonFieldScope(
          entry,
          postgresFixturePolicy,
          `manifest entry ${entry.id}`,
        );
        validateMigrationScratch(
          entry,
          goEntrySymbols(entry),
          postgresFixturePolicy,
          postgresFixtureBudget,
          `manifest entry ${entry.id}`,
        );
        validateTemplateCloneReason(entry, postgresFixturePolicy, `manifest entry ${entry.id}`);
        validateGroupCloneReason(entry, postgresFixturePolicy, `manifest entry ${entry.id}`);
        validatePackageResetReasonCode(entry, postgresFixturePolicy, `manifest entry ${entry.id}`);
        if (typeof entry.package !== "string" || !entry.package.startsWith("./")) {
          throw new Error(`manifest entry ${entry.id} must declare a repo-relative Go package owner`);
        }
      } else {
        if (entry.runtime_binaries !== undefined) {
          throw new Error(`manifest entry ${entry.id} runtime_binaries are valid only for runner=go_test`);
        }
        throw new Error(`manifest entry ${entry.id} must declare runner=go_test|playwright|vitest`);
      }
      if (entry.runner !== "go_test" && entry.runtime_binaries !== undefined) {
        throw new Error(`manifest entry ${entry.id} runtime_binaries are valid only for runner=go_test`);
      }
      if (typeof entry.file !== "string" || entry.file.trim() === "") {
        throw new Error(`manifest entry ${entry.id} must declare a file`);
      }
      if (typeof entry.evidence_layer !== "string" || entry.evidence_layer.trim() === "") {
        throw new Error(`manifest entry ${entry.id} must declare evidence_layer`);
      }
      if (entry.coverage === "authoritative") {
        if (typeof entry.claim !== "string" || entry.claim.trim() === "") {
          throw new Error(`manifest entry ${entry.id} must declare a non-empty claim`);
        }
        if (typeof entry.out_of_scope !== "string" || entry.out_of_scope.trim() === "") {
          throw new Error(`manifest entry ${entry.id} must declare a non-empty out_of_scope`);
        }
      }
      validateFixtureRefs(entry, `manifest entry ${entry.id}`);
      validateFrontendFixtureRefs(root, entry, `manifest entry ${entry.id}`);
      assertAuthoritativeGridRowsUseLiveAdapter(root, entry, `manifest entry ${entry.id}`);
      entries.push({ ...entry, section });
    }
  }

  for (const entry of collectSupportGoEntries(manifest)) {
    assertObjectKeys(entry, supportGoEntryKeys, supportGoEntryLabel(entry));
    if (typeof entry.target !== "string" || !validSupportTargets.has(entry.target)) {
      throw new Error(
        `${supportGoEntryLabel(entry)} must declare target=backend_unit|backend_integration_support`,
      );
    }
    validateEvidencePlacement(entry, supportGoEntryLabel(entry));
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
    } catch {
      throw new Error(
        `${supportGoEntryLabel(entry)} has invalid selection_pattern ${JSON.stringify(entry.selection_pattern)}`,
      );
    }
    const symbols = supportGoEntrySymbols(entry);
    validateExecutionFamily(entry, supportGoEntryLabel(entry));
    runtimeBinaries(entry, supportGoEntryLabel(entry));
    const postgresFixturePolicy = supportGoEntryPostgresFixturePolicy(entry);
    if (serviceBackedSupportTargets.has(entry.target) && postgresFixturePolicy === "") {
      throw new Error(
        `${supportGoEntryLabel(entry)} must declare fixture_policy.postgres for service-backed support target ${entry.target}`,
      );
    }
    const postgresFixtureBudget = supportGoEntryPostgresFixtureBudget(entry);
    validatePostgresFixtureBudget(
      entry,
      postgresFixturePolicy,
      postgresFixtureBudget,
      supportGoEntryLabel(entry),
    );
    validatePostgresFixtureReasonFieldScope(
      entry,
      postgresFixturePolicy,
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
    validateGroupCloneReason(entry, postgresFixturePolicy, supportGoEntryLabel(entry));
    validatePackageResetReasonCode(entry, postgresFixturePolicy, supportGoEntryLabel(entry));
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
      `${phaseName} manifest mismatch: missing=${missing.join(",") || "none"} unexpected=${unexpected.join(",") || "none"}`,
    );
  }
  validateProfileClaims(manifest, entries, manifestPath);
  assertAuthoritativeEvidenceNames(manifest, { phase: phaseName });

  const guideExpectedIDs = process.env.CARTULARY_PHASE_MANIFEST_ROOT
    ? []
    : loadGuideExpectedIDs(root, phaseNumber);
  if (guideExpectedIDs.length > 0) {
    const guideMissing = guideExpectedIDs.filter((id) => !expected.includes(id));
    const guideUnexpected = expected.filter((id) => !guideExpectedIDs.includes(id));
    if (guideMissing.length > 0 || guideUnexpected.length > 0) {
      throw new Error(
        `${phaseName} guide mismatch: missing=${guideMissing.join(",") || "none"} unexpected=${guideUnexpected.join(",") || "none"}`,
      );
    }
  }

  for (const entry of entries) {
    const targetPath = path.join(root, entry.file);
    const source = readFileSync(targetPath, "utf8");
    const needles =
      entry.runner === "go_test"
        ? goEntrySymbols(entry)
        : entry.runner === "vitest"
          ? vitestEntryTitles(entry)
          : playwrightEntryTitles(entry);
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
      `${phaseName} has authoritative-looking Go tests missing from ${manifestPath}: ${unmanifestedAuthoritativeGoTests
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
        `${target} must not claim ${phaseName} authoritative ids: ${Array.from(claimedIDs).sort().join(", ")}`,
      );
    }
  }
}
