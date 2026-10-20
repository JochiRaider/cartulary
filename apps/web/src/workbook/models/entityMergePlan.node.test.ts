import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import {
  normalizeEntityAlias,
  normalizeEntityIdentifier,
} from "../adapters/entityIdentifierNormalization";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { entityMergeIdentifierFields } from "./entityIdentifierClasses";
import { buildMergePlan } from "./entityMergePlan";
import { entityRowFromApi } from "./entityWorkbookModel";

function fixture(name: string): unknown {
  return JSON.parse(
    readFileSync(
      path.resolve(
        path.dirname(fileURLToPath(import.meta.url)),
        "../../../../../contracts/entities",
        name,
      ),
      "utf8",
    ),
  );
}

type Side = {
  canonical: Record<string, string | null>;
  reusable: {
    identifier_class: string;
    raw_value: string;
    normalized_value: string;
  }[];
  aliases: string[];
};
type PlanningCase = {
  case_id: string;
  entity_type: "host" | "identity";
  survivor: Side;
  loser: Side;
  expected: {
    identifiers: {
      identifier_class: string;
      normalized_value: string;
      action: string;
      source: string;
    }[];
    aliases_to_copy: string[];
    duplicate_aliases: string[];
  };
};

function row(
  type: "host" | "identity",
  id: string,
  side: Side,
  reverse = false,
) {
  const cells: WorkbookQueryRow["cells"] = Object.fromEntries(
    entityMergeIdentifierFields[type].map((field) => [
      field.key,
      { value: side.canonical[field.identifierClass] ?? null },
    ]),
  );
  cells[`${type}.${type}_state`] = { value: "canonical" };
  cells[`${type}.display_name`] = { value: "Same display label" };
  const reusable = side.reusable.map((value, index) => ({
    ...value,
    item_ref: `entity_preserved_identifier:${id}:${index}`,
    item_kind: "reusable_identifier",
    display_text: value.raw_value,
  }));
  const aliases = side.aliases.map((value, index) => ({
    item_ref: `entity_alias:${id}:${index}`,
    item_kind: "alias",
    alias_text: value,
    display_text: value,
  }));
  cells[`${type}.reusable_identifiers`] = {
    value: { items: reverse ? reusable.reverse() : reusable },
  };
  cells[`${type}.aliases`] = {
    value: { items: reverse ? aliases.reverse() : aliases },
  };
  return entityRowFromApi({ record_id: id, row_version: 7, cells }, type);
}

describe("entityMergePlan", () => {
  it("matches the owner identifier normalization corpus and separate alias policy", () => {
    const corpus = fixture("identifier-normalization-corpus.v1.json") as {
      cases: {
        case_id: string;
        identifier_type: string;
        raw_value: string;
        admitted: boolean;
        normalized_value: string | null;
      }[];
    };
    for (const test of corpus.cases)
      expect(
        normalizeEntityIdentifier(test.identifier_type, test.raw_value),
        test.case_id,
      ).toBe(test.admitted ? test.normalized_value : null);
    expect(normalizeEntityAlias("\u0085 Cafe\u0301 \u0085")).toBe("Café");
    expect(normalizeEntityAlias("Shared")).not.toBe(
      normalizeEntityAlias("shared"),
    );
    expect(normalizeEntityAlias("source\u200balias")).toBe("source\u200balias");
    expect(
      normalizeEntityIdentifier("hostname", "source\u200balias"),
    ).toBeNull();
    expect(normalizeEntityIdentifier("hostname", "\ud800")).toBeNull();
  });

  it("explains owner promotion ordering, duplicates and aliases independently of collection positions", () => {
    const corpus = fixture("merge-planning-corpus.v1.json") as {
      cases: PlanningCase[];
    };
    for (const test of corpus.cases)
      for (const reverse of [false, true]) {
        const survivor = row(
          test.entity_type,
          "survivor",
          test.survivor,
          reverse,
        );
        const loser = row(test.entity_type, "loser", test.loser, reverse);
        const before = JSON.stringify([survivor, loser]);
        const plan = buildMergePlan(survivor, loser);
        expect(plan.valid, test.case_id).toBe(true);
        expect(
          plan.identifierOutcomes.map((value) => ({
            identifier_class: value.identifierClass,
            normalized_value: value.normalizedValue,
            action: value.action,
            source: value.source,
          })),
          test.case_id,
        ).toEqual(test.expected.identifiers);
        expect(plan.aliasesToCopy, test.case_id).toEqual(
          test.expected.aliases_to_copy,
        );
        expect(plan.duplicateAliases, test.case_id).toEqual(
          test.expected.duplicate_aliases,
        );
        expect(
          plan.identifierOutcomes.every(
            (value) =>
              value.sourceRecordId === "loser" &&
              value.classification === "exact_match_reuse",
          ),
        ).toBe(true);
        expect(plan.provenanceOnlySummary).toContain("historical loser");
        expect(JSON.stringify([survivor, loser])).toBe(before);
        expect(Object.isFrozen(plan.identifierOutcomes)).toBe(true);
      }
  });

  it("refuses missing, malformed, ineligible or inconsistent loaded inputs", () => {
    const empty: Side = { canonical: {}, reusable: [], aliases: [] };
    for (const type of ["host", "identity"] as const) {
      const survivor = row(type, "survivor", empty);
      const loser = row(type, "loser", empty);
      expect(buildMergePlan(survivor, survivor).valid).toBe(false);
      expect(
        buildMergePlan(survivor, { ...loser, state: "merged" }).valid,
      ).toBe(false);
      delete loser.rawRow.cells[entityMergeIdentifierFields[type][0].key];
      expect(buildMergePlan(survivor, loser).valid).toBe(false);
      const inconsistent = row(type, "loser", {
        ...empty,
        reusable: [
          {
            identifier_class:
              entityMergeIdentifierFields[type][0].identifierClass,
            raw_value: "UPPER",
            normalized_value: "wrong",
          },
        ],
      });
      expect(buildMergePlan(survivor, inconsistent).valid).toBe(false);
      const malformed = row(type, "loser", empty);
      malformed.rawRow.cells[`${type}.aliases`] = { value: { items: [null] } };
      expect(buildMergePlan(survivor, malformed).valid).toBe(false);
    }
  });
});
