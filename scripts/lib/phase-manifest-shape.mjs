import {
  assertObjectKeys,
  assertRequiredKeys,
  assertUnique,
  readJsonObject,
  requireEnum,
  requireObject,
  requireObjectArray,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "./json-shape.mjs";

const phaseNamePattern = /^phase(?:0|[1-9]\d*)$/;
const validCoverage = new Set(["authoritative", "supplemental"]);
const validRunners = new Set(["go_test", "playwright", "vitest"]);

export const phaseTestMapSchemaID = "cartulary.phase_test_map.v1";

export const phaseManifestTopLevelKeys = new Set([
  "schema_id",
  "phase",
  "note",
  "ledger",
  "expected_ids",
  "forbidden_id_files",
  "support_go_targets",
  "unit",
  "integration",
  "e2e",
  "visual",
]);

export const phaseManifestRequiredKeys = new Set([
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
  "title",
  "titles",
  "execution_dependency",
  "evidence_layer",
  "claim",
  "out_of_scope",
  "execution_family",
  "execution_label",
  "fixture_policy",
  "fixture_budget",
  "fixture_refs",
  "template_clone_reason",
  "migration_scratch_reason",
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
  "fixture_policy",
  "fixture_budget",
  "migration_scratch_reason",
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
    requireString(entry.section, `${entryLabel}.section`);
    requireString(entry.package, `${entryLabel}.package`);
    requireString(entry.file, `${entryLabel}.file`);
    requireString(entry.selection_pattern, `${entryLabel}.selection_pattern`);
  }
}

export function validatePhaseManifestShapeFile(file) {
  validatePhaseManifestShape(readJsonObject(file, file), file);
}
