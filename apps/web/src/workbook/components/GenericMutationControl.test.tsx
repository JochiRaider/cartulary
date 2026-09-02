import type { ViewFieldContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { emptyGenericReferenceOptions } from "../models/workbookReferenceOptions";
import { resolveGenericMutationControl } from "./genericMutationControlModel";

const baseField: ViewFieldContract = {
  clearable: true,
  conflictResolutionClass: null,
  createWritable: true,
  defaultHidden: false,
  directReferenceContractId: null,
  directScalarContractId: null,
  entityBindingMode: null,
  enumValues: null,
  fieldKey: "test.value",
  filterOps: [],
  gridEditable: true,
  groupable: false,
  headerSortFieldKey: null,
  label: "Test value",
  readKind: "text",
  sortable: false,
  stringContractId: null,
  writeAction: null,
  writeKind: "direct_value",
};

const referenceOptions = {
  ...emptyGenericReferenceOptions(),
  allRecords: [
    { label: "Record A", recordId: "record-a", viewSchemaId: "view-a" },
  ],
  incidentMembers: [
    { label: "Analyst A", recordId: "user-a", viewSchemaId: "members" },
  ],
};

describe("GenericMutationControl descriptor", () => {
  it("resolves every field-control variant in form and grid modes", () => {
    const cases = [
      {
        collectionItems: [{ displayText: "Alias A", itemRef: "alias-a" }],
        collectionMode: "remove" as const,
        field: {
          ...baseField,
          fieldKey: "host.aliases",
          readKind: "collection",
          writeKind: "action_payload" as const,
        },
        kind: "collection_removal",
      },
      {
        collectionMode: "add" as const,
        field: {
          ...baseField,
          fieldKey: "decision.support_refs",
          readKind: "collection",
          writeKind: "action_payload" as const,
        },
        kind: "collection_reference",
      },
      {
        collectionMode: "add" as const,
        field: {
          ...baseField,
          directReferenceContractId: "incident_member_user_ref_v1",
        },
        kind: "direct_reference",
      },
      {
        collectionMode: "add" as const,
        field: { ...baseField, enumValues: ["open", "closed"] },
        kind: "enumerated_value",
      },
      {
        collectionMode: "add" as const,
        field: { ...baseField, readKind: "boolean" },
        kind: "boolean",
      },
      {
        collectionMode: "add" as const,
        field: { ...baseField, readKind: "number" },
        kind: "number",
      },
      {
        collectionMode: "add" as const,
        field: { ...baseField, fieldKey: "assessment.rationale" },
        kind: "multiline_text",
      },
      {
        collectionMode: "add" as const,
        field: {
          ...baseField,
          directScalarContractId: "timestamp_instant_v1",
        },
        kind: "text",
      },
    ] as const;

    for (const testCase of cases) {
      for (const surface of ["form", "grid"] as const) {
        expect(
          resolveGenericMutationControl({
            collectionItems:
              "collectionItems" in testCase ? testCase.collectionItems : [],
            collectionMode: testCase.collectionMode,
            field: testCase.field,
            referenceOptions,
            surface,
          }).kind,
          `${testCase.kind} on ${surface}`,
        ).toBe(testCase.kind);
      }
    }
  });

  it("derives options, sizing, clear labels, rows, number types, and timestamp hints", () => {
    const collectionField = {
      ...baseField,
      fieldKey: "decision.support_refs",
      readKind: "collection",
      writeKind: "action_payload" as const,
    };
    expect(
      resolveGenericMutationControl({
        collectionItems: [],
        collectionMode: "add",
        field: collectionField,
        referenceOptions,
        surface: "form",
      }),
    ).toMatchObject({
      kind: "collection_reference",
      options: [{ label: "Record A", value: "record-a" }],
      size: 2,
    });
    expect(
      resolveGenericMutationControl({
        collectionItems: [],
        collectionMode: "add",
        field: collectionField,
        referenceOptions,
        surface: "grid",
      }),
    ).toMatchObject({ kind: "collection_reference", size: 1 });
    expect(
      resolveGenericMutationControl({
        collectionItems: [],
        collectionMode: "add",
        field: {
          ...baseField,
          clearable: false,
          directReferenceContractId: "incident_member_user_ref_v1",
        },
        referenceOptions,
        surface: "form",
      }),
    ).toMatchObject({ emptyLabel: "Select", kind: "direct_reference" });
    expect(
      resolveGenericMutationControl({
        collectionItems: [],
        collectionMode: "add",
        field: { ...baseField, readKind: "number" },
        referenceOptions,
        surface: "grid",
      }),
    ).toMatchObject({ inputType: "text", kind: "number" });
    expect(
      resolveGenericMutationControl({
        collectionItems: [],
        collectionMode: "add",
        field: { ...baseField, fieldKey: "assessment.rationale" },
        referenceOptions,
        surface: "form",
      }),
    ).toMatchObject({ kind: "multiline_text", rows: 3 });
    expect(
      resolveGenericMutationControl({
        collectionItems: [],
        collectionMode: "add",
        field: {
          ...baseField,
          directScalarContractId: "timestamp_instant_v1",
        },
        referenceOptions,
        surface: "form",
      }),
    ).toMatchObject({ kind: "text", placeholder: "RFC3339 timestamp" });
  });
});
