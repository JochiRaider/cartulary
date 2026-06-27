import { describe, expect, it } from "vitest";

import {
  fieldCapability,
  getViewContract,
  listViewContracts,
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
    inspector_config: {
      inspector_config_schema_id: "cartulary.inspector_config.v1",
      view_schema_id: "cartulary.view.fixture.v1",
      default_open: false,
      subject_binding: { kind: "selected_record" },
      no_row_state: "no_row_selected",
      unsupported_feature_behavior: "omit_feature",
      panels: [
        { panel_id: "details", label: "Details" },
        { panel_id: "history", label: "History" },
      ],
      feature_groups: [
        {
          feature_group_key: "details.read",
          panel_id: "details",
          label: "Read details",
          minimum_incident_role: null,
          mutates: false,
          requires_confirmation: false,
          route_binding: { kind: "panel_read" },
          seed_bindings: [],
          disabled_when: ["no_row_selected"],
        },
        {
          feature_group_key: "history.rollback",
          panel_id: "history",
          label: "Rollback",
          minimum_incident_role: "editor",
          mutates: true,
          requires_confirmation: true,
          route_binding: {
            kind: "record_action",
            action_key: "history.rollback",
          },
          seed_bindings: [
            {
              target_field_key: "fixture.editable",
              source: {
                kind: "selected_field_value",
                source_field_key: "fixture.queryable",
              },
            },
          ],
          disabled_when: ["no_row_selected", "row_version_changed"],
        },
      ],
    },
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
    const timeline = requireViewContract("cartulary.view.timeline.v2");

    expect(timeline.permitsZeroFieldCreate).toBe(true);
    expect(timeline.sortFields).toContain("timeline.activity_sort_ts");
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
    const timeline = requireViewContract("cartulary.view.timeline.v2");

    expect(
      resolveHeaderSortFieldKey(timeline, "timeline.activity_utc_text"),
    ).toBe("timeline.activity_sort_ts");
    expect(fieldCapability(timeline, "timeline.activity_utc_text")).toEqual({
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

  it("exposes required reference-pack metadata from view contracts", () => {
    const packBound = parseFixture({
      ...fixtureRawContract(),
      required_reference_pack_keys: ["mitre_attack_enterprise"],
    });

    expect(packBound.requiredReferencePackKeys).toEqual([
      "mitre_attack_enterprise",
    ]);
    expect(
      listViewContracts().map((contract) => contract.requiredReferencePackKeys),
    ).toEqual(listViewContracts().map(() => []));
  });

  it("exposes inspector config by stable view_schema_id and semantic keys", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v2");
    const fixture = parseFixture();
    const relabeled = parseFixture({
      ...fixtureRawContract(),
      title: "Relabeled Fixture",
      inspector_config: {
        ...fixtureRawContract().inspector_config,
        panels: [
          { panel_id: "details", label: "Renamed details" },
          { panel_id: "history", label: "Renamed history" },
        ],
      },
    });

    expect(timeline.inspectorConfig.viewSchemaId).toBe(timeline.viewSchemaId);
    expect(timeline.inspectorConfig.defaultOpen).toBe(false);
    expect(timeline.inspectorConfig.noRowState).toBe("no_row_selected");
    expect(
      fixture.inspectorConfig.panels.map((panel) => panel.panelId),
    ).toEqual(["details", "history"]);
    expect(
      relabeled.inspectorConfig.featureGroups.map(
        (group) => group.featureGroupKey,
      ),
    ).toEqual(["details.read", "history.rollback"]);
    expect(
      getViewContract(timeline.inspectorConfig.panels[0]?.label ?? ""),
    ).toBe(undefined);
  });

  it("rejects invalid inspector config vocabulary, bounds, and ownership", () => {
    const tooManyFeatureGroups = Array.from({ length: 65 }, (_, index) => ({
      ...fixtureRawContract().inspector_config.feature_groups[0],
      feature_group_key: `details.read_${index}`,
    }));

    const cases: ReadonlyArray<{
      readonly pattern: RegExp;
      readonly raw: unknown;
    }> = [
      {
        raw: { ...fixtureRawContract(), inspector_config: undefined },
        pattern: /inspector_config must be an object/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          inspector_config: {
            ...fixtureRawContract().inspector_config,
            view_schema_id: "cartulary.view.other.v1",
          },
        },
        pattern:
          /inspector_config\.view_schema_id must match cartulary\.view\.fixture\.v1/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          inspector_config: {
            ...fixtureRawContract().inspector_config,
            panels: [{ panel_id: "legacy", label: "Legacy" }],
          },
        },
        pattern:
          /inspector_config\.panels\[1\]\.panel_id must be one of details\|relationships\|evidence\|history\|workflow/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          inspector_config: {
            ...fixtureRawContract().inspector_config,
            panels: [
              { panel_id: "details", label: "Details" },
              { panel_id: "details", label: "Details again" },
            ],
          },
        },
        pattern: /inspector_config\.panels duplicate panel_id details/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          inspector_config: {
            ...fixtureRawContract().inspector_config,
            feature_groups: [
              fixtureRawContract().inspector_config.feature_groups[0],
              fixtureRawContract().inspector_config.feature_groups[0],
            ],
          },
        },
        pattern:
          /inspector_config\.feature_groups duplicate feature_group_key details\.read/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          inspector_config: {
            ...fixtureRawContract().inspector_config,
            feature_groups: tooManyFeatureGroups,
          },
        },
        pattern:
          /inspector_config\.feature_groups must contain at most 64 entries/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          inspector_config: {
            ...fixtureRawContract().inspector_config,
            feature_groups:
              fixtureRawContract().inspector_config.feature_groups.map(
                (group) =>
                  group.feature_group_key === "details.read"
                    ? {
                        ...group,
                        route_binding: { kind: "legacy_route" },
                      }
                    : group,
              ),
          },
        },
        pattern:
          /route_binding\.kind must be one of panel_read\|view_row_create\|record_patch\|record_action\|entity_mention_action\|evidence_access\|surface_pivot/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          inspector_config: {
            ...fixtureRawContract().inspector_config,
            feature_groups:
              fixtureRawContract().inspector_config.feature_groups.map(
                (group) =>
                  group.feature_group_key === "details.read"
                    ? {
                        ...group,
                        disabled_when: ["stale_legacy_state"],
                      }
                    : group,
              ),
          },
        },
        pattern:
          /disabled_when\[1\] must be one of no_row_selected\|incident_closed\|authorization_lost\|row_version_changed\|record_deleted\|record_merged\|evidence_preview_unavailable\|merge_target_unavailable/,
      },
    ];

    for (const { pattern, raw } of cases) {
      expectInvariantFailure(raw, pattern);
    }
  });
});

describe("FE-U-P0-02 view-schema field-key adapter contract", () => {
  it("FE-U-P0-02 selects generated contracts by view_schema_id, not display title", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v2");

    expect(timeline.viewSchemaId).toBe("cartulary.view.timeline.v2");
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
        pattern:
          /default_visible_fields references unknown field_key fixture\.unknown/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          default_hidden_fields: ["fixture.unknown"],
        },
        pattern:
          /default_hidden_fields references unknown field_key fixture\.unknown/,
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
        pattern:
          /grouping_fields references unknown field_key fixture\.unknown/,
      },
      {
        raw: {
          ...fixtureRawContract(),
          default_sort: [{ field_key: "fixture.unknown", direction: "asc" }],
        },
        pattern:
          /default_sort\[1\]\.field_key references unknown field_key fixture\.unknown/,
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
        pattern:
          /fields\[2\]\.header_sort_field_key references unknown field_key fixture\.unknown/,
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
