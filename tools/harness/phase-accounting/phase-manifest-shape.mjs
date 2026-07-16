import {
  assertObjectKeys,
  assertRequiredKeys,
  assertUnique,
  readJsonObject,
  requireBoolean,
  requireEnum,
  requireObject,
  requireObjectArray,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "../contract/json-shape.mjs";

const phaseNamePattern = /^phase(?:0|[1-9]\d*)$/;
const validCoverage = new Set(["authoritative", "supplemental"]);
const validRunners = new Set(["go_test", "playwright", "vitest"]);
const validClaimStatuses = new Set(["implemented", "blocked", "not_applicable"]);
export const validBackendEvidenceClasses = new Set([
  "product_conformance",
  "implementation_support",
  "harness_policy",
  "readiness",
  "diagnostic",
  "duplicate_regression",
  "measurement",
  "security",
  "release_artifact",
  "visual_readiness",
  "accessibility_readiness",
  "claim_publication_boundary",
]);
export const validBackendLayers = new Set([
  "backend_unit",
  "backend_store",
  "backend_integration",
  "backend_integration_support",
  "backend_process",
  "browser_functional",
  "browser_stateful",
  "browser_measurement",
  "browser_support",
  "browser_visual",
  "frontend_unit",
]);

const phaseTestMapSchemaID = "cartulary.phase_test_map.v2";

export const validDefaultCheckKinds = new Set([
  "primary_local_evidence",
  "default_local_cross_stack_conformance",
  "full_target_equivalent",
  "bounded_readiness",
  "explicit_only",
  "duplicate_regression",
]);

export const validDefaultCheckReasonCodes = new Set([
  "cheapest_authoritative_layer",
  "lower_layer_gap",
  "full_target_equivalent_stateful",
  "bounded_readiness",
  "explicit_full_target",
  "explicit_readiness",
  "explicit_measurement",
  "implementation_support_explicit_only",
  "design_direction_explicit_only",
  "claim_publication_boundary",
  "duplicate_of_primary_owner",
]);

export const validWarmLocalCostClasses = new Set([
  "none",
  "low",
  "medium",
  "service_backed",
  "browser",
  "explicit_heavy",
]);

export const validPostgresFixtureReasonCodes = new Set([
  "committed_cross_connection_visibility",
  "database_identity",
  "process_lifecycle",
  "schema_mutation",
  "destructive_residue",
  "shared_seeded_state",
  "bounded_reset_surface",
  "migration_scratch",
]);

const phaseManifestTopLevelKeys = new Set([
  "schema_id",
  "phase",
  "note",
  "profile_claims",
  "ledger",
  "shared_harnesses",
  "expected_ids",
  "forbidden_id_files",
  "support_go_targets",
  "unit",
  "integration",
  "e2e",
  "visual",
]);

const phaseManifestRequiredKeys = new Set([
  "note",
  "ledger",
  "expected_ids",
  "support_go_targets",
  "unit",
  "integration",
  "e2e",
]);

export const phaseLedgerKeys = new Set([
  "title",
  "notes",
  "authoritative_execution",
  "support_execution_extras",
  "sections",
  "shared_harness",
  "support_only",
]);

export const phaseManifestEntryKeys = new Set([
  "id",
  "coverage",
  "runner",
  "package",
  "file",
  "symbol",
  "symbols",
  "scenario_symbols",
  "title",
  "titles",
  "execution_dependency",
  "evidence_class",
  "layer",
  "default_check_required",
  "default_check_kind",
  "default_check_reason_code",
  "default_check_reason",
  "primary_evidence_owner",
  "duplicate_of",
  "evidence_delta",
  "warm_local_cost_class",
  "evidence_layer",
  "claim_status",
  "claim",
  "out_of_scope",
  "execution_family",
  "execution_label",
  "runtime_binaries",
  "runtime_profile_id",
  "fixture_policy",
  "fixture_budget",
  "fixture_proof",
  "symbol_fixture_overrides",
  "fixture_refs",
  "frontend_fixture_refs",
  "template_clone_reason",
  "template_clone_reason_code",
  "group_clone_reason",
  "group_clone_reason_code",
  "package_reset_reason",
  "package_reset_reason_code",
  "migration_scratch_reason",
  "migration_scratch_reason_code",
]);

export const supportGoEntryKeys = new Set([
  "target",
  "section",
  "package",
  "file",
  "symbol",
  "symbols",
  "selection_pattern",
  "execution_family",
  "execution_label",
  "runtime_binaries",
  "evidence_class",
  "layer",
  "default_check_required",
  "default_check_kind",
  "default_check_reason_code",
  "default_check_reason",
  "primary_evidence_owner",
  "duplicate_of",
  "evidence_delta",
  "warm_local_cost_class",
  "fixture_policy",
  "fixture_budget",
  "fixture_proof",
  "template_clone_reason",
  "template_clone_reason_code",
  "group_clone_reason",
  "group_clone_reason_code",
  "package_reset_reason",
  "package_reset_reason_code",
  "migration_scratch_reason",
  "migration_scratch_reason_code",
]);

const profileClaimKeys = new Set([
  "profile_id",
  "claimed",
  "claim_ac_id",
  "required_ac_ids",
  "direct_evidence_ids",
  "aggregate_ac_ids",
]);

const sharedHarnessEntryKeys = new Set([
  "id",
  "coverage",
  "harnesses",
  "evidence",
  "notes",
]);

export function validatePhaseManifestShape(manifest, label) {
  assertObjectKeys(manifest, phaseManifestTopLevelKeys, label);
  assertRequiredKeys(manifest, phaseManifestRequiredKeys, label);
  requireSchemaID(manifest, phaseTestMapSchemaID, label);
  requireString(manifest.phase, `${label}.phase`, { pattern: phaseNamePattern });
  requireString(manifest.note, `${label}.note`);

  const ledger = requireObject(manifest.ledger, `${label}.ledger`);
  assertObjectKeys(ledger, phaseLedgerKeys, `${label}.ledger`);

  requireStringArray(manifest.expected_ids, `${label}.expected_ids`, {
    nonEmpty: true,
  });
  if (manifest.forbidden_id_files !== undefined) {
    requireStringArray(
      manifest.forbidden_id_files,
      `${label}.forbidden_id_files`,
    );
  }
  const profileClaims = requireObjectArray(
    manifest.profile_claims ?? [],
    `${label}.profile_claims`,
  );
  const profileIDs = [];
  for (const [index, claim] of profileClaims.entries()) {
    const claimLabel = `${label}.profile_claims[${index + 1}]`;
    assertObjectKeys(claim, profileClaimKeys, claimLabel);
    profileIDs.push(requireString(claim.profile_id, `${claimLabel}.profile_id`));
    if (typeof claim.claimed !== "boolean") {
      throw new Error(`${claimLabel}.claimed must be a boolean`);
    }
    requireString(claim.claim_ac_id, `${claimLabel}.claim_ac_id`);
    requireStringArray(claim.required_ac_ids, `${claimLabel}.required_ac_ids`, {
      nonEmpty: true,
    });
    requireStringArray(claim.direct_evidence_ids, `${claimLabel}.direct_evidence_ids`);
    requireStringArray(claim.aggregate_ac_ids, `${claimLabel}.aggregate_ac_ids`);
  }
  assertUnique(profileIDs, `${label}.profile_claims.profile_id`);

  const sharedHarnesses = requireObjectArray(
    manifest.shared_harnesses ?? [],
    `${label}.shared_harnesses`,
  );
  const sharedHarnessIDs = [];
  for (const [index, entry] of sharedHarnesses.entries()) {
    const entryLabel = `${label}.shared_harnesses[${index + 1}]`;
    assertObjectKeys(entry, sharedHarnessEntryKeys, entryLabel);
    sharedHarnessIDs.push(requireString(entry.id, `${entryLabel}.id`));
    requireEnum(entry.coverage, `${entryLabel}.coverage`, validCoverage);
    requireStringArray(entry.harnesses, `${entryLabel}.harnesses`, {
      nonEmpty: true,
    });
    requireStringArray(entry.evidence, `${entryLabel}.evidence`, {
      nonEmpty: true,
    });
    if (entry.notes !== undefined) {
      requireStringArray(entry.notes, `${entryLabel}.notes`);
    }
  }
  assertUnique(sharedHarnessIDs, `${label}.shared_harnesses.id`);

  for (const section of ["unit", "integration", "e2e", "visual"]) {
    const entries = requireObjectArray(
      manifest[section] ?? [],
      `${label}.${section}`,
    );
    const ids = [];
    for (const [index, entry] of entries.entries()) {
      const entryLabel = `${label}.${section}[${index + 1}]`;
      assertObjectKeys(entry, phaseManifestEntryKeys, entryLabel);
      ids.push(requireString(entry.id, `${entryLabel}.id`));
      requireEnum(entry.coverage, `${entryLabel}.coverage`, validCoverage);
      requireEnum(entry.runner, `${entryLabel}.runner`, validRunners);
      requireEnum(entry.evidence_class, `${entryLabel}.evidence_class`, validBackendEvidenceClasses);
      requireEnum(entry.layer, `${entryLabel}.layer`, validBackendLayers);
      requireBoolean(entry.default_check_required, `${entryLabel}.default_check_required`);
      requireEnum(entry.default_check_kind, `${entryLabel}.default_check_kind`, validDefaultCheckKinds);
      requireEnum(
        entry.default_check_reason_code,
        `${entryLabel}.default_check_reason_code`,
        validDefaultCheckReasonCodes,
      );
      requireString(entry.primary_evidence_owner, `${entryLabel}.primary_evidence_owner`);
      if (entry.duplicate_of !== null) {
        requireString(entry.duplicate_of, `${entryLabel}.duplicate_of`);
      }
      requireString(entry.evidence_delta, `${entryLabel}.evidence_delta`);
      requireEnum(
        entry.warm_local_cost_class,
        `${entryLabel}.warm_local_cost_class`,
        validWarmLocalCostClasses,
      );
      if (entry.default_check_reason !== undefined) {
        requireString(entry.default_check_reason, `${entryLabel}.default_check_reason`);
      }
      requireString(entry.file, `${entryLabel}.file`);
      if (entry.title !== undefined && entry.titles !== undefined) {
        throw new Error(`${entryLabel} must declare title or titles[], not both`);
      }
      if (entry.titles !== undefined) {
        requireStringArray(entry.titles, `${entryLabel}.titles`, {
          nonEmpty: true,
        });
      }
      if (entry.fixture_refs !== undefined) {
        requireStringArray(entry.fixture_refs, `${entryLabel}.fixture_refs`, {
          nonEmpty: true,
        });
      }
      if (entry.frontend_fixture_refs !== undefined) {
        requireStringArray(entry.frontend_fixture_refs, `${entryLabel}.frontend_fixture_refs`, {
          nonEmpty: true,
        });
      }
      if (entry.runtime_binaries !== undefined) {
        requireStringArray(entry.runtime_binaries, `${entryLabel}.runtime_binaries`, {
          nonEmpty: true,
        });
      }
      if (entry.runtime_profile_id !== undefined) {
        requireString(entry.runtime_profile_id, `${entryLabel}.runtime_profile_id`, {
          pattern: /^[a-z][a-z0-9_]*$/u,
        });
        if (entry.runner !== "playwright") {
          throw new Error(`${entryLabel}.runtime_profile_id is only valid for Playwright evidence`);
        }
      }
      if (entry.claim_status !== undefined) {
        requireEnum(entry.claim_status, `${entryLabel}.claim_status`, validClaimStatuses);
      }
      requireString(entry.evidence_layer, `${entryLabel}.evidence_layer`);
      if (entry.symbol !== undefined && entry.symbols !== undefined) {
        throw new Error(`${entryLabel} must declare symbol or symbols[], not both`);
      }
    }
    assertUnique(ids, `${label}.${section}.id`);
  }

  const supportEntries = requireObjectArray(
    manifest.support_go_targets ?? [],
    `${label}.support_go_targets`,
  );
  for (const [index, entry] of supportEntries.entries()) {
    const entryLabel = `${label}.support_go_targets[${index + 1}]`;
    assertObjectKeys(entry, supportGoEntryKeys, entryLabel);
    requireString(entry.target, `${entryLabel}.target`);
    requireEnum(entry.evidence_class, `${entryLabel}.evidence_class`, validBackendEvidenceClasses);
    requireEnum(entry.layer, `${entryLabel}.layer`, validBackendLayers);
    requireBoolean(entry.default_check_required, `${entryLabel}.default_check_required`);
    requireEnum(entry.default_check_kind, `${entryLabel}.default_check_kind`, validDefaultCheckKinds);
    requireEnum(
      entry.default_check_reason_code,
      `${entryLabel}.default_check_reason_code`,
      validDefaultCheckReasonCodes,
    );
    requireString(entry.primary_evidence_owner, `${entryLabel}.primary_evidence_owner`);
    if (entry.duplicate_of !== null) {
      requireString(entry.duplicate_of, `${entryLabel}.duplicate_of`);
    }
    requireString(entry.evidence_delta, `${entryLabel}.evidence_delta`);
    requireEnum(
      entry.warm_local_cost_class,
      `${entryLabel}.warm_local_cost_class`,
      validWarmLocalCostClasses,
    );
    if (entry.default_check_reason !== undefined) {
      requireString(entry.default_check_reason, `${entryLabel}.default_check_reason`);
    }
    if (entry.runtime_binaries !== undefined) {
      requireStringArray(entry.runtime_binaries, `${entryLabel}.runtime_binaries`, {
        nonEmpty: true,
      });
    }
    requireString(entry.section, `${entryLabel}.section`);
    requireString(entry.package, `${entryLabel}.package`);
    requireString(entry.file, `${entryLabel}.file`);
    requireString(entry.selection_pattern, `${entryLabel}.selection_pattern`);
  }
}

export function validatePhaseManifestShapeFile(file) {
  validatePhaseManifestShape(readJsonObject(file, file), file);
}
