import {
  compareEntityNormalizedValues,
  normalizeEntityAlias,
  normalizeEntityIdentifier,
} from "../adapters/entityIdentifierNormalization";
import { entityMergeIdentifierFields } from "./entityIdentifierClasses";
import type { EntityRow } from "./entityWorkbookModel";

export type EntityMergeIdentifierOutcome = {
  readonly identifierClass: string;
  readonly label: string;
  readonly normalizedValue: string;
  readonly displayValue: string;
  readonly sourceRecordId: string;
  readonly source: "canonical" | "reusable_identifier";
  readonly sourceItemRef: string | null;
  readonly classification: "exact_match_reuse";
  readonly action: "promotion" | "carry_forward" | "duplicate_noop";
};

export type EntityMergePlan = {
  readonly valid: boolean;
  readonly issues: readonly string[];
  readonly identifierOutcomes: readonly EntityMergeIdentifierOutcome[];
  readonly aliasesToCopy: readonly string[];
  readonly duplicateAliases: readonly string[];
  readonly provenanceOnlySummary: string;
  readonly dependencySummary: string;
};

type Candidate = Omit<EntityMergeIdentifierOutcome, "action">;
type Input = {
  readonly canonical: ReadonlyMap<string, Candidate>;
  readonly secondary: ReadonlyMap<string, readonly Candidate[]>;
  readonly aliases: readonly string[];
};

function object(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function planningInput(row: EntityRow, issues: string[]): Input {
  const canonical = new Map<string, Candidate>();
  const secondary = new Map<string, Candidate[]>();
  const fields = entityMergeIdentifierFields[row.entityType];
  const fail = () =>
    issues.push(
      `The loaded identifiers or aliases for ${row.label} (${row.recordId}) could not be verified. Refresh before reviewing.`,
    );
  for (const field of fields) {
    const cell = row.rawRow.cells[field.key];
    if (
      cell === undefined ||
      (cell.value !== null && typeof cell.value !== "string")
    ) {
      fail();
      continue;
    }
    if (cell.value === null) continue;
    const normalized = normalizeEntityIdentifier(
      field.identifierClass,
      cell.value,
    );
    if (normalized === null) {
      fail();
      continue;
    }
    canonical.set(field.identifierClass, {
      identifierClass: field.identifierClass,
      label: field.label,
      normalizedValue: normalized,
      displayValue: cell.value,
      sourceRecordId: row.recordId,
      source: "canonical",
      sourceItemRef: null,
      classification: "exact_match_reuse",
    });
  }
  const reusable =
    row.rawRow.cells[`${row.entityType}.reusable_identifiers`]?.value;
  if (!object(reusable) || !Array.isArray(reusable.items)) fail();
  else
    for (const item of reusable.items) {
      if (!object(item)) {
        fail();
        continue;
      }
      const field = fields.find(
        (field) => field.identifierClass === item.identifier_class,
      );
      if (
        field === undefined ||
        item.item_kind !== "reusable_identifier" ||
        typeof item.item_ref !== "string" ||
        item.item_ref === "" ||
        typeof item.raw_value !== "string" ||
        typeof item.display_text !== "string" ||
        item.display_text === "" ||
        typeof item.normalized_value !== "string" ||
        normalizeEntityIdentifier(field.identifierClass, item.raw_value) !==
          item.normalized_value
      ) {
        fail();
        continue;
      }
      const values = secondary.get(field.identifierClass) ?? [];
      values.push({
        identifierClass: field.identifierClass,
        label: field.label,
        normalizedValue: item.normalized_value,
        displayValue: item.raw_value,
        sourceRecordId: row.recordId,
        source: "reusable_identifier",
        sourceItemRef: item.item_ref,
        classification: "exact_match_reuse",
      });
      secondary.set(field.identifierClass, values);
    }
  for (const candidates of secondary.values())
    candidates.sort(
      (left, right) =>
        compareEntityNormalizedValues(
          left.normalizedValue,
          right.normalizedValue,
        ) ||
        compareEntityNormalizedValues(
          left.sourceItemRef ?? "",
          right.sourceItemRef ?? "",
        ),
    );
  const aliases: string[] = [];
  const collection = row.rawRow.cells[`${row.entityType}.aliases`]?.value;
  if (!object(collection) || !Array.isArray(collection.items)) fail();
  else
    for (const item of collection.items) {
      if (
        !object(item) ||
        item.item_kind !== "alias" ||
        typeof item.item_ref !== "string" ||
        item.item_ref === "" ||
        typeof item.alias_text !== "string" ||
        normalizeEntityAlias(item.alias_text) !== item.alias_text
      ) {
        fail();
        continue;
      }
      aliases.push(item.alias_text);
    }
  return {
    canonical,
    secondary,
    aliases: [...new Set(aliases)].sort(compareEntityNormalizedValues),
  };
}

export function buildMergePlan(
  survivor: EntityRow,
  loser: EntityRow,
): EntityMergePlan {
  const issues: string[] = [];
  if (
    survivor.recordId === loser.recordId ||
    survivor.entityType !== loser.entityType ||
    [survivor, loser].some(
      (row) =>
        !["stub", "canonical"].includes(row.state) ||
        !Number.isSafeInteger(row.rowVersion) ||
        row.rowVersion < 1,
    )
  ) {
    issues.push(
      "Select two distinct active records of the same entity type and refresh their current versions.",
    );
  }
  const existing = planningInput(survivor, issues);
  const candidates = planningInput(loser, issues);
  const outcomes: EntityMergeIdentifierOutcome[] = [];
  for (const field of entityMergeIdentifierFields[survivor.entityType]) {
    const key = field.identifierClass;
    const populated = existing.canonical.get(key);
    const values = new Set(
      (existing.secondary.get(key) ?? []).map((value) => value.normalizedValue),
    );
    if (populated !== undefined) values.add(populated.normalizedValue);
    let empty = populated === undefined;
    const canonical = candidates.canonical.get(key);
    const ordered = [
      ...(canonical === undefined ? [] : [canonical]),
      ...(candidates.secondary.get(key) ?? []),
    ];
    const admitted = new Set<string>();
    for (const candidate of ordered) {
      if (admitted.has(candidate.normalizedValue)) continue;
      admitted.add(candidate.normalizedValue);
      const action = values.has(candidate.normalizedValue)
        ? "duplicate_noop"
        : empty
          ? "promotion"
          : "carry_forward";
      if (action !== "duplicate_noop") {
        values.add(candidate.normalizedValue);
        empty = false;
      }
      outcomes.push(Object.freeze({ ...candidate, action }));
    }
  }
  const survivorAliases = new Set(existing.aliases);
  return Object.freeze({
    valid: issues.length === 0,
    issues: Object.freeze([...new Set(issues)]),
    identifierOutcomes: Object.freeze(outcomes),
    aliasesToCopy: Object.freeze(
      candidates.aliases.filter((alias) => !survivorAliases.has(alias)),
    ),
    duplicateAliases: Object.freeze(
      candidates.aliases.filter((alias) => survivorAliases.has(alias)),
    ),
    provenanceOnlySummary:
      "The historical loser, merge lineage and source provenance are retained. Ordinary aliases remain suggestion-only; reusable identifiers remain exact-match values.",
    dependencySummary:
      "This review explains loaded identifiers and aliases. The server checks collisions, current versions, dependencies and authorization when the merge is submitted.",
  });
}

export function mergeIdentifierOutcomeText(
  outcome: EntityMergeIdentifierOutcome,
): string {
  const action =
    outcome.action === "promotion"
      ? "Promote"
      : outcome.action === "carry_forward"
        ? "Carry as reusable"
        : "Duplicate no-op";
  return `${action} ${outcome.displayValue}`;
}
