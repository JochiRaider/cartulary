import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  requireViewContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  findCellMatches,
  normalizeFindText,
} from "../find/WorkbookFindController";
import {
  entityCellPresentation,
  entityFindText,
} from "./entityCellPresentation";
import { entityRowFromApi } from "./entityWorkbookModel";

const surfaces = [
  ["host", hostsViewSchemaId],
  ["identity", identitiesViewSchemaId],
] as const;
describe("Entity Find presentation", () => {
  it("preserves renderer fallback precedence and eligible literal fields on both schemas", () => {
    for (const [type, view] of surfaces) {
      const contract = requireViewContract(view);
      const row = entityRowFromApi(
        {
          record_id: "internal-record",
          row_version: 77,
          cells: Object.fromEntries(
            Object.entries({
              [`${type}.display_name`]: "Name Café",
              [`${type}.${type === "host" ? "hostname" : "upn"}`]: "primary",
              [`${type}.${type === "host" ? "fqdn" : "email"}`]: "fallback",
              [`${type}.linked_event_count`]: 0,
            }).map(([key, value]) => [key, { value }]),
          ),
        },
        type,
      );
      expect(entityFindText(row, `${type}.display_name`, contract)).toEqual([
        "Name Café",
      ]);
      expect(
        entityFindText(
          row,
          `${type}.${type === "host" ? "hostname" : "upn"}`,
          contract,
        ),
      ).toEqual([type === "host" ? "primary" : "fallback"]);
      expect(
        entityFindText(row, `${type}.linked_event_count`, contract),
      ).toEqual(["0"]);
      expect(
        findCellMatches(
          entityFindText(row, `${type}.display_name`, contract),
          normalizeFindText("CAFE\u0301", false),
          false,
        ),
      ).toBe(true);
      const empty = entityRowFromApi(
        { record_id: "internal-record", row_version: 77, cells: {} },
        type,
      );
      expect(entityCellPresentation(empty, `${type}.display_name`).text).toBe(
        "internal-record",
      );
      for (const field of [
        `${type}.display_name`,
        `${type}.aliases`,
        `${type}.${type === "host" ? "hostname" : "upn"}`,
        "row_version",
        "record_id",
        "metadata",
      ]) {
        expect(entityFindText(empty, field, contract)).toEqual([]);
      }
    }
  });
  it("uses independently displayed collection fragments and excludes fallback metadata and arbitrary JSON", () => {
    for (const [type, view] of surfaces) {
      const contract = requireViewContract(view);
      const field = `${type}.reusable_identifiers`;
      const row = entityRowFromApi(
        {
          record_id: "internal",
          row_version: 2,
          cells: {
            [field]: {
              value: {
                items: [
                  {
                    display_text: "Alpha needle",
                    raw_text: "not displayed",
                    item_ref: "technical",
                  },
                  { display_text: "Beta separate" },
                  { item_ref: "private fallback" },
                ],
              },
            },
            [`${type}.aliases`]: {
              value: {
                items: [
                  {
                    item_kind: "alias",
                    item_ref:
                      "entity_alias:00000000-0000-0000-0000-000000000001",
                    alias_text: "A needle",
                    display_text: "A needle",
                  },
                  {
                    item_kind: "alias",
                    item_ref:
                      "entity_alias:00000000-0000-0000-0000-000000000002",
                    alias_text: "B fragment",
                    display_text: "B fragment",
                  },
                ],
              },
            },
          },
        },
        type,
      );
      expect(entityCellPresentation(row, field).text).toBe(
        "Alpha needle, Beta separate, private fallback",
      );
      const fragments = entityFindText(row, field, contract);
      expect(fragments).toEqual(["Alpha needle", "Beta separate"]);
      expect(findCellMatches(fragments, "needle, Beta", true)).toBe(false);
      expect(entityFindText(row, `${type}.aliases`, contract)).toEqual([
        "A needle",
        "B fragment",
      ]);
      row.rawRow.cells[field] = { value: { private: "arbitrary JSON" } };
      expect(entityFindText(row, field, contract)).toEqual([]);
    }
  });
});
