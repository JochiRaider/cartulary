import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import {
  serviceBackedGoExecutionDependencies,
  serviceBackedSupportTargets,
  validExecutionDependencies,
  validSupportTargets,
} from "../execution/execution-dependencies.mjs";
import { assertObjectKeys } from "../contract/json-shape.mjs";
import {
  defaultReasonRequiredLayers,
  sectionDefinitions,
  supportTargetSections,
  validClaimStatuses,
  validCoverage,
  validGoSections,
  validRuntimeBinaries,
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
  effectiveGoEntryPostgresFixtureBudget,
  effectiveGoEntryPostgresFixturePolicy,
  effectiveSupportGoEntryPostgresFixtureBudget,
  effectiveSupportGoEntryPostgresFixturePolicy,
  postgresFixturePolicyPackageReset,
  validateGroupCloneReason,
  validateMigrationScratch,
  validatePackageResetReasonCode,
  validatePostgresFixtureBudget,
  validateTemplateCloneReason,
} from "./phase-fixture-policy.mjs";
import {
  phaseLedgerKeys,
  phaseManifestEntryKeys,
  supportGoEntryKeys,
  validBackendEvidenceClasses,
  validBackendLayers,
  validDefaultCheckKinds,
  validDefaultCheckReasonCodes,
  validWarmLocalCostClasses,
} from "./phase-manifest-shape.mjs";
import {
  assertAuthoritativeGridRowsUseLiveAdapter,
  validateFixtureRefs,
  validateFrontendFixtureRefs,
} from "./phase-frontend-fixtures.mjs";
import { loadManifest, phaseNumberFromPhase } from "./phase-manifest-loader.mjs";

const implementationTestingGuidePath = path.join(
  "docs",
  "guides",
  "cartulary_implementation_testing_guide.md",
);

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

function runtimeBinaries(entry, label) {
  if (entry.runtime_binaries === undefined) {
    return [];
  }
  if (!Array.isArray(entry.runtime_binaries) || entry.runtime_binaries.length === 0) {
    throw new Error(`${label} runtime_binaries must be a non-empty string array when present`);
  }
  const seen = new Set();
  const result = [];
  for (const [index, raw] of entry.runtime_binaries.entries()) {
    if (typeof raw !== "string" || raw.trim() === "") {
      throw new Error(`${label} runtime_binaries[${index + 1}] must be a non-empty string`);
    }
    const id = raw.trim();
    if (!validRuntimeBinaries.has(id)) {
      throw new Error(`${label} runtime_binaries[${index + 1}] has unknown runtime binary ${id}`);
    }
    if (seen.has(id)) {
      throw new Error(`${label} runtime_binaries contains duplicate ${id}`);
    }
    seen.add(id);
    result.push(id);
  }
  return result;
}

function validateEvidencePlacement(entry, label) {
  if (typeof entry.evidence_class !== "string" || !validBackendEvidenceClasses.has(entry.evidence_class)) {
    throw new Error(`${label} must declare closed evidence_class`);
  }
  if (typeof entry.layer !== "string" || !validBackendLayers.has(entry.layer)) {
    throw new Error(`${label} must declare closed layer`);
  }
  if (typeof entry.default_check_required !== "boolean") {
    throw new Error(`${label} must declare default_check_required as a boolean`);
  }
  if (
    typeof entry.default_check_kind !== "string" ||
    !validDefaultCheckKinds.has(entry.default_check_kind)
  ) {
    throw new Error(`${label} must declare closed default_check_kind`);
  }
  if (
    typeof entry.default_check_reason_code !== "string" ||
    !validDefaultCheckReasonCodes.has(entry.default_check_reason_code)
  ) {
    throw new Error(`${label} must declare closed default_check_reason_code`);
  }
  if (
    typeof entry.primary_evidence_owner !== "string" ||
    entry.primary_evidence_owner.trim() === ""
  ) {
    throw new Error(`${label} must declare primary_evidence_owner`);
  }
  if (
    entry.duplicate_of !== null &&
    (typeof entry.duplicate_of !== "string" || entry.duplicate_of.trim() === "")
  ) {
    throw new Error(`${label} must declare duplicate_of as null or a non-empty string`);
  }
  if (typeof entry.evidence_delta !== "string" || entry.evidence_delta.trim() === "") {
    throw new Error(`${label} must declare evidence_delta`);
  }
  if (
    typeof entry.warm_local_cost_class !== "string" ||
    !validWarmLocalCostClasses.has(entry.warm_local_cost_class)
  ) {
    throw new Error(`${label} must declare closed warm_local_cost_class`);
  }
  if (entry.default_check_required === true && entry.default_check_kind === "explicit_only") {
    throw new Error(`${label} default_check_required=true cannot use default_check_kind=explicit_only`);
  }
  if (entry.default_check_required === false && entry.default_check_kind === "primary_local_evidence") {
    throw new Error(`${label} default_check_required=false cannot use primary_local_evidence`);
  }
  const requiresReason =
    entry.default_check_required === true &&
    (entry.evidence_class !== "product_conformance" ||
      defaultReasonRequiredLayers.has(entry.layer));
  if (requiresReason) {
    if (typeof entry.default_check_reason !== "string" || entry.default_check_reason.trim() === "") {
      throw new Error(`${label} must declare default_check_reason for non-obvious default check inclusion`);
    }
  }
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
    String.raw`\b[UIEV]${escapeRegex(separator)}${phaseNumber}${escapeRegex(
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
    String.raw`^\|\s*([UIEV]-${phaseNumber}(?:-[A-Z0-9]+)*-\d{2})\s*\|`,
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

function validateProfileClaims(manifest, entries, manifestPath) {
  const entryByID = new Map(entries.map((entry) => [entry.id, entry]));
  for (const claim of manifest.profile_claims ?? []) {
    const label = `manifest ${manifestPath} profile_claims.${claim.profile_id}`;
    if (!claim.required_ac_ids.includes(claim.claim_ac_id)) {
      throw new Error(`${label} claim_ac_id must be listed in required_ac_ids`);
    }
    for (const [field, values] of [
      ["required_ac_ids", claim.required_ac_ids],
      ["direct_evidence_ids", claim.direct_evidence_ids],
      ["aggregate_ac_ids", claim.aggregate_ac_ids],
    ]) {
      const uniqueValues = new Set(values);
      if (uniqueValues.size !== values.length) {
        throw new Error(`${label} ${field} must not contain duplicates`);
      }
    }
    for (const aggregateAC of claim.aggregate_ac_ids) {
      if (!claim.required_ac_ids.includes(aggregateAC)) {
        throw new Error(`${label} aggregate_ac_ids must be a subset of required_ac_ids`);
      }
    }
    if (!claim.claimed) {
      continue;
    }
    if (claim.direct_evidence_ids.length === 0) {
      throw new Error(`${label} claimed profiles must declare direct_evidence_ids`);
    }
    for (const evidenceID of claim.direct_evidence_ids) {
      const entry = entryByID.get(evidenceID);
      if (!entry || entry.coverage !== "authoritative") {
        throw new Error(`${label} direct_evidence_id ${evidenceID} must name an authoritative row`);
      }
      const status = entry.claim_status ?? "implemented";
      if (status !== "implemented") {
        throw new Error(
          `${label} direct_evidence_id ${evidenceID} must have claim_status=implemented`,
        );
      }
    }
  }
}
