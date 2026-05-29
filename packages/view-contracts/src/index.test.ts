import { describe, expect, it } from "vitest";

import {
  fieldCapability,
  getViewContract,
  parseViewContractJSON,
  requireViewContract,
  resolveHeaderSortFieldKey,
  type ViewContract,
  visibleFields,
} from "./index";

function fixtureRawContract() {
  return {
    view_schema_id: "cartulary.view.fixture.v1",
    title: "Fixture Surface",
    surface_kind: "system_view",
    default_visible_fields: ["fixture.editable", "fixture.queryable"],
    default_hidden_fields: ["record_id", "row_version", "fixture.sort_shadow"],
    default_sort: [{ field_key: "fixture.sort_shadow", direction: "asc" }],
    sort_fields: ["fixture.sort_shadow", "fixture.queryable"],
    filter_fields: ["fixture.queryable"],
    grouping_fields: ["fixture.queryable"],
    technical_fields: ["record_id", "row_version"],
    fields: [
      {
        field_key: "fixture.editable",
        label: "Editable Field",
        write_kind: "direct_value",
      },
      {
        field_key: "fixture.queryable",
        label: "Queryable Field",
        filter_ops: ["eq"],
        groupable: true,
        header_sort_field_key: "fixture.sort_shadow",
        sortable: true,
      },
      {
        field_key: "fixture.sort_shadow",
        label: "Sort Shadow",
        default_hidden: true,
        sortable: true,
      },
    ],
    synthetic_filter_predicates: [
      {
        field_key: "fixture.full_text",
        label: "Full Text",
        filter_ops: ["full_text"],
      },
    ],
  };
}

function parseFixture(raw: unknown = fixtureRawContract()) {
  return parseViewContractJSON(JSON.stringify(raw), "fixture-contract.json");
}

function expectInvariantFailure(raw: unknown, pattern: RegExp) {
  expect(() =>
    parseViewContractJSON(JSON.stringify(raw), "broken-contract.json"),
  ).toThrow(pattern);
}

describe("view-contracts", () => {
  it("parses sortable, filterable, and zero-field create metadata", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v1");

    expect(timeline.permitsZeroFieldCreate).toBe(true);
    expect(timeline.sortFields).toContain("timeline.sort_ts");
    expect(timeline.sortNullOrder).toBe("last");
    expect(timeline.filterFields).toContain("timeline.capture_state");
    expect(timeline.groupingFields).toContain("timeline.capture_state");
  });

  it("exposes synthetic filter predicates as filter-only fields", () => {
    const notes = requireViewContract("cartulary.view.notes.v1");

    expect(notes.filterFields).toContain("note.full_text");
    expect(notes.fieldMap["note.full_text"]).toMatchObject({
      fieldKey: "note.full_text",
      filterOps: ["full_text"],
      label: "Full Text",
      sortable: false,
      writeKind: "read_only",
    });
    expect(visibleFields(notes).map((field) => field.fieldKey)).not.toContain(
      "note.full_text",
    );
  });

  it("resolves header sort keys and field capabilities from contract metadata", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v1");

    expect(resolveHeaderSortFieldKey(timeline, "timeline.occurred_at")).toBe(
      "timeline.sort_ts",
    );
    expect(fieldCapability(timeline, "timeline.occurred_at")).toEqual({
      editable: true,
      filterable: false,
      groupable: false,
      sortable: true,
    });
  });

  it("returns the contract-default visible field order", () => {
    const hosts = requireViewContract("cartulary.view.hosts.v1");
    expect(visibleFields(hosts).map((field) => field.fieldKey)).toEqual(
      hosts.defaultVisibleFields,
    );
  });

  it("parses enum values for contract-backed controls", () => {
    const assessments = requireViewContract("cartulary.view.assessments.v1");

    expect(
      assessments.fieldMap["assessment.assessment_state"]?.enumValues,
    ).toEqual(["unknown", "suspected", "confirmed", "disproven", "cleared"]);
    expect(
      assessments.fieldMap["assessment.confidence_band"]?.enumValues,
    ).toEqual(["unset", "low", "medium", "high"]);
  });

  it("exposes mutation metadata needed by workbook controls", () => {
    const evidence = requireViewContract("cartulary.view.evidence.v1");

    expect(evidence.fieldMap["evidence.title"]?.stringContractId).toBe(
      "single_line_title_v1",
    );
    expect(
      evidence.fieldMap["evidence.requested_at"]?.directScalarContractId,
    ).toBe("timestamp_instant_v1");
    expect(
      evidence.fieldMap["evidence.collector_party_id"]
        ?.directReferenceContractId,
    ).toBe("same_incident_party_ref_v1");
    expect(evidence.fieldMap["evidence.lifecycle_state"]?.enumValues).toEqual([
      "requested",
      "pending_receipt",
      "received",
      "available",
      "quarantined",
      "released",
    ]);
  });
});

describe("FE-U-P0-02 view-schema field-key adapter contract", () => {
  it("FE-U-P0-02 selects generated contracts by view_schema_id, not display title", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v1");

    expect(timeline.viewSchemaId).toBe("cartulary.view.timeline.v1");
    expect(getViewContract(timeline.title)).toBeUndefined();
    expect(getViewContract("Timeline")).toBeUndefined();
  });

  it("FE-U-P0-02 keeps field identity and capabilities stable when labels change", () => {
    const first = parseFixture();
    const relabeled = parseFixture({
      ...fixtureRawContract(),
      title: "Renamed Surface",
      fields: [
        {
          field_key: "fixture.editable",
          label: "Editable Field Renamed",
          write_kind: "direct_value",
        },
        {
          field_key: "fixture.queryable",
          label: "Queryable Field Renamed",
          filter_ops: ["eq"],
          groupable: true,
          header_sort_field_key: "fixture.sort_shadow",
          sortable: true,
        },
        {
          field_key: "fixture.sort_shadow",
          label: "Sort Shadow Renamed",
          default_hidden: true,
          sortable: true,
        },
      ],
      synthetic_filter_predicates: [
        {
          field_key: "fixture.full_text",
          label: "Full Text Renamed",
          filter_ops: ["full_text"],
        },
      ],
    });

    expect(first.fieldMap["fixture.editable"]?.fieldKey).toBe(
      relabeled.fieldMap["fixture.editable"]?.fieldKey,
    );
    expect(fieldCapability(first, "fixture.editable")).toEqual(
      fieldCapability(relabeled, "fixture.editable"),
    );
    expect(resolveHeaderSortFieldKey(first, "fixture.queryable")).toBe(
      resolveHeaderSortFieldKey(relabeled, "fixture.queryable"),
    );
    expect(fieldCapability(relabeled, "Editable Field Renamed")).toEqual({
      editable: false,
      filterable: false,
      groupable: false,
      sortable: false,
    });
  });

  it("FE-U-P0-02 treats duplicate labels as display-only with distinct field_key identities", () => {
    const duplicateLabels = parseFixture({
      ...fixtureRawContract(),
      fields: [
        {
          field_key: "fixture.editable",
          label: "Shared Display Label",
          write_kind: "direct_value",
        },
        {
          field_key: "fixture.queryable",
          label: "Shared Display Label",
          filter_ops: ["eq"],
          groupable: true,
          header_sort_field_key: "fixture.sort_shadow",
          sortable: true,
        },
        {
          field_key: "fixture.sort_shadow",
          label: "Sort Shadow",
          default_hidden: true,
          sortable: true,
        },
      ],
    });

    expect(duplicateLabels.fieldMap["fixture.editable"]?.label).toBe(
      "Shared Display Label",
    );
    expect(duplicateLabels.fieldMap["fixture.queryable"]?.label).toBe(
      "Shared Display Label",
    );
    expect(fieldCapability(duplicateLabels, "fixture.editable")).toEqual({
      editable: true,
      filterable: false,
      groupable: false,
      sortable: false,
    });
    expect(fieldCapability(duplicateLabels, "fixture.queryable")).toEqual({
      editable: false,
      filterable: true,
      groupable: true,
      sortable: true,
    });
    expect(fieldCapability(duplicateLabels, "Shared Display Label")).toEqual({
      editable: false,
      filterable: false,
      groupable: false,
      sortable: false,
    });
  });

  it("FE-U-P0-02 keeps field identity stable when fields and visible columns reorder", () => {
    const base = parseFixture();
    const reordered = parseFixture({
      ...fixtureRawContract(),
      default_visible_fields: ["fixture.queryable", "fixture.editable"],
      fields: [...fixtureRawContract().fields].reverse(),
    });

    expect(fieldCapability(base, "fixture.editable")).toEqual(
      fieldCapability(reordered, "fixture.editable"),
    );
    expect(fieldCapability(base, "fixture.queryable")).toEqual(
      fieldCapability(reordered, "fixture.queryable"),
    );
    expect(resolveHeaderSortFieldKey(reordered, "fixture.queryable")).toBe(
      "fixture.sort_shadow",
    );
    expect(
      new Set(visibleFields(reordered).map((field) => field.fieldKey)),
    ).toEqual(new Set(["fixture.editable", "fixture.queryable"]));
  });

  it("FE-U-P0-02 resolves editable, queryable, and synthetic fields by field_key", () => {
    const contract = parseFixture();

    expect(fieldCapability(contract, "fixture.editable")).toEqual({
      editable: true,
      filterable: false,
      groupable: false,
      sortable: false,
    });
    expect(fieldCapability(contract, "fixture.queryable")).toEqual({
      editable: false,
      filterable: true,
      groupable: true,
      sortable: true,
    });
    expect(fieldCapability(contract, "fixture.full_text")).toEqual({
      editable: false,
      filterable: true,
      groupable: false,
      sortable: false,
    });
    expect(resolveHeaderSortFieldKey(contract, "fixture.queryable")).toBe(
      "fixture.sort_shadow",
    );
    expect(contract.filterableFieldMap["fixture.full_text"]).toBe(true);
    expect(fieldCapability(contract, "Queryable Field")).toEqual({
      editable: false,
      filterable: false,
      groupable: false,
      sortable: false,
    });
  });

  it("FE-U-P0-02 fails deterministically for duplicate and missing field_key inputs", () => {
    expectInvariantFailure(
      {
        ...fixtureRawContract(),
        fields: [
          ...fixtureRawContract().fields,
          {
            field_key: "fixture.queryable",
            label: "Duplicate Field",
          },
        ],
      },
      /View contract invariant failed: broken-contract\.json fields duplicate field_key fixture\.queryable/,
    );

    expectInvariantFailure(
      {
        ...fixtureRawContract(),
        fields: [
          {
            field_key: "",
            label: "Missing Key",
          },
        ],
      },
      /View contract invariant failed: broken-contract\.json fields\[1\]\.field_key must be a non-empty string/,
    );

    expectInvariantFailure(
      {
        ...fixtureRawContract(),
        synthetic_filter_predicates: [
          {
            field_key: "",
            label: "Missing Synthetic Key",
          },
        ],
      },
      /View contract invariant failed: broken-contract\.json synthetic_filter_predicates\[1\]\.field_key must be a non-empty string/,
    );
  });

  it("FE-U-P0-02 fails deterministically for unknown field_key references", () => {
    const cases: ReadonlyArray<{
      readonly pattern: RegExp;
      readonly raw: unknown;
    }> = [
      {
        raw: {
          ...fixtureRawContract(),
          default_visible_fields: ["fixture.unknown"],
        },
        pattern: /default_visible_fields references unknown field_key fixture\.unknown/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          default_hidden_fields: ["fixture.unknown"],
        },
        pattern: /default_hidden_fields references unknown field_key fixture\.unknown/,
      },
      {
        raw: { ...fixtureRawContract(), sort_fields: ["fixture.unknown"] },
        pattern: /sort_fields references unknown field_key fixture\.unknown/,
      },
      {
        raw: { ...fixtureRawContract(), filter_fields: ["fixture.unknown"] },
        pattern: /filter_fields references unknown field_key fixture\.unknown/,
      },
      {
        raw: { ...fixtureRawContract(), grouping_fields: ["fixture.unknown"] },
        pattern: /grouping_fields references unknown field_key fixture\.unknown/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          default_sort: [{ field_key: "fixture.unknown", direction: "asc" }],
        },
        pattern: /default_sort\[1\]\.field_key references unknown field_key fixture\.unknown/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          fields: fixtureRawContract().fields.map((field) =>
            field.field_key === "fixture.queryable"
              ? { ...field, header_sort_field_key: "fixture.unknown" }
              : field,
          ),
        },
        pattern: /fields\[2\]\.header_sort_field_key references unknown field_key fixture\.unknown/,
      },
    ];

    for (const { pattern, raw } of cases) {
      expectInvariantFailure(raw, pattern);
    }
  });

  it("FE-U-P0-02 visibleFields fails deterministically when default-visible keys are unresolved", () => {
    const contract = parseFixture();
    const brokenContract: ViewContract = {
      ...contract,
      defaultVisibleFields: ["fixture.missing"],
    };

    expect(() => visibleFields(brokenContract)).toThrow(
      /View contract invariant failed: cartulary\.view\.fixture\.v1 default_visible_fields references unknown field_key fixture\.missing/,
    );
  });
});
