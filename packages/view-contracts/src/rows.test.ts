import { describe, expect, expectTypeOf, it } from "vitest";

import {
  type NormalizedViewRowPatchV1,
  type NormalizedViewRowV1,
  normalizeViewRowPatchV1,
  normalizeViewRowV1,
  type ViewContract,
  type ViewFieldContract,
} from "./index";

const field = (fieldKey: string): ViewFieldContract =>
  Object.freeze({
    clearable: true,
    conflictResolutionClass: null,
    createWritable: false,
    defaultHidden: false,
    directReferenceContractId: null,
    directScalarContractId: null,
    entityBindingMode: null,
    enumValues: null,
    fieldKey,
    filterOps: Object.freeze([]),
    gridEditable: true,
    groupable: fieldKey === "fixture.group",
    headerSortFieldKey: null,
    label: fieldKey,
    readKind: "text",
    sortable: false,
    stringContractId: null,
    writeAction: null,
    writeKind: "direct_value",
  });

const fields = Object.freeze([
  field("fixture.editable"),
  field("fixture.group"),
  field("fixture.sort"),
]);
const contract = {
  fieldMap: Object.freeze(
    Object.fromEntries(fields.map((entry) => [entry.fieldKey, entry])),
  ),
  fields,
  groupingFields: Object.freeze(["fixture.group"]),
  technicalFields: Object.freeze(["record_id", "row_version"]),
  viewSchemaId: "cartulary.view.fixture.v1",
} as unknown as ViewContract;

function fixtureCells() {
  return {
    "fixture.editable": { value: "editable" },
    "fixture.group": { value: "bucket" },
    "fixture.sort": { value: "sort" },
  };
}

describe("view row normalization", () => {
  it("rejects missing and malformed row identity and version values", () => {
    for (const row of [
      { cells: fixtureCells(), row_version: 1 },
      { cells: fixtureCells(), record_id: "record-1" },
      { cells: fixtureCells(), record_id: "record-1", row_version: 0 },
      { cells: fixtureCells(), record_id: "record-1", row_version: 1.5 },
    ]) {
      expect(() => normalizeViewRowV1(contract, row)).toThrow(
        /View row invariant failed/u,
      );
    }
  });

  it("keeps normalized full rows and sparse patches non-assignable", () => {
    expectTypeOf<NormalizedViewRowV1>().not.toMatchTypeOf<NormalizedViewRowPatchV1>();
    expectTypeOf<NormalizedViewRowPatchV1>().not.toMatchTypeOf<NormalizedViewRowV1>();
  });

  it("requires complete full-row cells and rejects technical cells", () => {
    expect(() =>
      normalizeViewRowV1(contract, {
        cells: { "fixture.editable": { value: "only" } },
        group_values: { "fixture.group": "bucket" },
        record_id: "record-1",
        row_version: 1,
      }),
    ).toThrow("missing cell fixture.group");
    expect(() =>
      normalizeViewRowPatchV1(contract, {
        cells: { record_id: { value: "forbidden" } },
        record_id: "record-1",
        row_version: 2,
      }),
    ).toThrow("technical cell record_id is not allowed");
  });

  it("ignores additive row and cell members while preserving wire-derived values", () => {
    const row = normalizeViewRowV1(contract, {
      additive: true,
      cells: Object.fromEntries(
        Object.entries(fixtureCells()).map(([key, cell]) => [
          key,
          { ...cell, additive: true },
        ]),
      ),
      group_values: { "fixture.group": "bucket" },
      record_id: "record-1",
      row_version: 2,
    });
    expect(row.recordId).toBe("record-1");
    expect(row.rowVersion).toBe(2);
    expect(row.cells["fixture.editable"]).toEqual({ value: "editable" });
    expect(Object.isFrozen(row)).toBe(true);
    expect(Object.isFrozen(row.cells)).toBe(true);
    expect(Object.isFrozen(row.groupValues)).toBe(true);
  });

  it("accepts sparse patch cells without weakening full-row completeness", () => {
    const patch = normalizeViewRowPatchV1(contract, {
      cells: { "fixture.editable": { value: "changed" } },
      record_id: "record-1",
      row_version: 3,
    });
    expect(Object.keys(patch.cells)).toEqual(["fixture.editable"]);
    expect(Object.isFrozen(patch)).toBe(true);
    expect(() =>
      normalizeViewRowV1(contract, {
        cells: patch.cells,
        group_values: { "fixture.group": "bucket" },
        record_id: "record-1",
        row_version: 3,
      }),
    ).toThrow("missing cell fixture.group");
  });

  it("rejects row view identity mismatches deterministically", () => {
    expect(() =>
      normalizeViewRowPatchV1(contract, {
        cells: {},
        record_id: "record-1",
        row_version: 2,
        view_schema_id: "cartulary.view.other.v1",
      }),
    ).toThrow("view_schema_id must be cartulary.view.fixture.v1");
  });

  it("enforces conditional full-row group_values and sparse patch group_values", () => {
    expect(() =>
      normalizeViewRowV1(contract, {
        cells: fixtureCells(),
        record_id: "record-1",
        row_version: 1,
      }),
    ).toThrow("group_values is required for grouped schema");
    expect(
      normalizeViewRowPatchV1(contract, {
        cells: {},
        record_id: "record-1",
        row_version: 2,
      }).groupValues,
    ).toBeUndefined();
  });
});
