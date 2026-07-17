import { describe, expect, it } from "vitest";

import {
  fieldCapability,
  getViewContract,
  getWorkbookSurfaceContract,
  listViewContracts,
  listWorkbookSurfaceContracts,
  normalizeViewRowPatchV1,
  normalizeViewRowV1,
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
        grid_editable: true,
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
          route_binding: {
            kind: "panel_read",
            owner: "current_row_projection",
          },
          seed_bindings: [],
          disabled_when: ["no_row_selected"],
          success_result_behavior: "preserve_selected_row",
          failure_result_behavior: "show_same_shell_error_preserve_selection",
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
            owner: "record_rollback_route",
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
          disabled_when: [
            "no_row_selected",
            "row_version_changed",
            "rollback_target_unavailable",
          ],
          success_result_behavior: "retarget_selected_row",
          failure_result_behavior:
            "show_same_shell_error_invalidate_pending_action",
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
  it("joins workbook identity and status metadata in Core 01 Table 7.4-A order", () => {
    const entries = listWorkbookSurfaceContracts();

    expect(entries.map((entry) => entry.viewSchemaId)).toEqual([
      "cartulary.view.timeline.v2",
      "cartulary.view.hosts.v1",
      "cartulary.view.identities.v1",
      "cartulary.view.evidence.v1",
      "cartulary.view.notes.v1",
      "cartulary.view.indicators.v1",
      "cartulary.view.assessments.v1",
      "cartulary.view.task_requests.v1",
      "cartulary.view.decisions.v1",
      "cartulary.view.parties.v1",
      "cartulary.view.comm_log.v1",
      "cartulary.view.handoff.v1",
      "cartulary.view.status_review.v1",
      "cartulary.view.lesson.v1",
      "cartulary.view.findings.v1",
      "cartulary.view.investigative_queries.v1",
      "cartulary.view.forensic_keywords.v1",
    ]);
    expect(entries.map((entry) => entry.surfaceStatus)).toEqual([
      ...Array.from({ length: 5 }, () => "required_built_in_sheet"),
      ...Array.from({ length: 9 }, () => "required_system_view"),
      ...Array.from(
        { length: 3 },
        () => "standardized_optional_workbook_surface",
      ),
    ]);
    expect(
      getWorkbookSurfaceContract("cartulary.view.missing.v1"),
    ).toBeUndefined();
  });

  it("parses sortable, filterable, and inline-create metadata", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v2");
    const hosts = requireViewContract("cartulary.view.hosts.v1");
    const identities = requireViewContract("cartulary.view.identities.v1");

    expect(timeline.permitsZeroFieldCreate).toBe(true);
    expect(timeline.minimumCreateFieldSets).toEqual([]);
    expect(hosts.minimumCreateFieldSets).toEqual([
      ["host.display_name"],
      ["host.hostname"],
      ["host.fqdn"],
      ["host.aad_device_id"],
    ]);
    expect(identities.minimumCreateFieldSets).toEqual([
      ["identity.display_name"],
      ["identity.aad_object_id"],
      ["identity.sid"],
      ["identity.upn"],
      ["identity.email"],
      ["identity.sam_account_name"],
    ]);
    expect(timeline.sortFields).toContain("timeline.activity_sort_ts");
    expect(timeline.sortNullOrder).toBe("last");
    expect(timeline.filterFields).toContain("timeline.capture_state");
    expect(timeline.groupingFields).toContain("timeline.capture_state");
  });

  it("enforces conditional full-row group_values and sparse patch group_values", () => {
    const grouped = parseFixture();
    const cells = {
      "fixture.editable": { value: "editable" },
      "fixture.queryable": { value: "bucket" },
      "fixture.sort_shadow": { value: "sort" },
    };
    expect(
      normalizeViewRowV1(grouped, {
        record_id: "record-1",
        row_version: 1,
        cells,
        group_values: { "fixture.queryable": "bucket" },
      }).groupValues,
    ).toEqual({ "fixture.queryable": "bucket" });
    expect(() =>
      normalizeViewRowV1(grouped, {
        record_id: "record-1",
        row_version: 1,
        cells,
      }),
    ).toThrow(/group_values is required for grouped schema/);
    expect(() =>
      normalizeViewRowV1(grouped, {
        record_id: "record-1",
        row_version: 1,
        cells,
        group_values: { "fixture.unknown": "bucket" },
      }),
    ).toThrow(/unknown group_values field fixture\.unknown/);

    const multiGrouped = parseFixture({
      ...fixtureRawContract(),
      grouping_fields: ["fixture.editable", "fixture.queryable"],
      fields: fixtureRawContract().fields.map((field) => ({
        ...field,
        groupable:
          field.field_key === "fixture.editable" ||
          field.field_key === "fixture.queryable",
      })),
    });
    expect(() =>
      normalizeViewRowV1(multiGrouped, {
        record_id: "record-1",
        row_version: 1,
        cells,
        group_values: { "fixture.queryable": "bucket" },
      }),
    ).toThrow(/missing group_values field fixture\.editable/);

    const ungrouped = parseFixture({
      ...fixtureRawContract(),
      grouping_fields: [],
      fields: fixtureRawContract().fields.map((field) => ({
        ...field,
        groupable: false,
      })),
    });
    expect(
      normalizeViewRowV1(ungrouped, {
        record_id: "record-1",
        row_version: 1,
        cells,
      }).groupValues,
    ).toBeUndefined();
    expect(() =>
      normalizeViewRowV1(ungrouped, {
        record_id: "record-1",
        row_version: 1,
        cells,
        group_values: {},
      }),
    ).toThrow(/group_values is not allowed for ungrouped schema/);

    expect(
      normalizeViewRowPatchV1(grouped, {
        record_id: "record-1",
        row_version: 2,
        cells: {},
      }).groupValues,
    ).toBeUndefined();
    expect(
      normalizeViewRowPatchV1(grouped, {
        record_id: "record-1",
        row_version: 2,
        cells: {},
        group_values: { "fixture.queryable": "next" },
      }).groupValues,
    ).toEqual({ "fixture.queryable": "next" });
    expect(() =>
      normalizeViewRowPatchV1(grouped, {
        record_id: "record-1",
        row_version: 2,
        cells: {},
        group_values: { "fixture.unknown": "next" },
      }),
    ).toThrow(/unknown group_values field fixture\.unknown/);
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
    expect(
      assessments.fieldMap["assessment.assessment_state"]?.gridEditable,
    ).toBe(false);
    expect(
      requireViewContract("cartulary.view.indicators.v1").fieldMap[
        "indicator.display_value"
      ]?.gridEditable,
    ).toBe(false);
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
    expect(
      requireViewContract("cartulary.view.task_requests.v1").fieldMap[
        "task.owner_user_id"
      ]?.directReferenceContractId,
    ).toBe("incident_member_user_ref_v1");
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
    expect(fixture.inspectorConfig.featureGroups[0]?.routeBinding.owner).toBe(
      "current_row_projection",
    );
    expect(
      fixture.inspectorConfig.featureGroups[1]?.successResultBehavior,
    ).toBe("retarget_selected_row");
    expect(
      fixture.inspectorConfig.featureGroups[1]?.failureResultBehavior,
    ).toBe("show_same_shell_error_invalidate_pending_action");
    expect(
      getViewContract(timeline.inspectorConfig.panels[0]?.label ?? ""),
    ).toBe(undefined);
  });

  it("FE-U-P9-03 Verify inspector panels and feature groups render from active config keys and reject unknown panel IDs, route owners, result behaviors, disabled tokens, duplicate keys, missing required keys, and extra current-profile keys before rendering.", () => {
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
                        route_binding: {
                          kind: "legacy_route",
                          owner: "current_row_projection",
                        },
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
                        route_binding: {
                          kind: "panel_read",
                          owner: "legacy_owner",
                        },
                      }
                    : group,
              ),
          },
        },
        pattern:
          /route_binding\.owner must be one of current_row_projection\|view_query_route\|view_row_create_route\|record_patch_route\|record_mark_reviewed_route\|record_supersede_route\|record_delete_route\|record_restore_route\|record_history_route\|record_rollback_route\|record_merge_route\|entity_mention_resolve_route\|evidence_attach_blob_route\|evidence_preview_handle_route\|evidence_download_handle_route/,
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
          /disabled_when\[1\] must be one of no_row_selected\|incident_closed\|authorization_lost\|row_version_changed\|record_deleted\|record_merged\|evidence_preview_unavailable\|merge_target_unavailable\|record_not_deleted\|rollback_target_unavailable\|party_text_unavailable\|pivot_target_unavailable/,
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
                        success_result_behavior: "legacy_success",
                      }
                    : group,
              ),
          },
        },
        pattern:
          /success_result_behavior must be one of preserve_selected_row\|retarget_selected_row\|clear_to_no_row_selected\|surface_pivot/,
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
                        failure_result_behavior: "legacy_failure",
                      }
                    : group,
              ),
          },
        },
        pattern:
          /failure_result_behavior must be one of show_same_shell_error_preserve_selection\|show_same_shell_error_invalidate_pending_action\|show_same_shell_error_clear_subject/,
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
                    ? { ...group, feature_group_key: "Details Read" }
                    : group,
              ),
          },
        },
        pattern: /feature_group_key must be ASCII lower snake or dotted key/,
      },
    ];

    for (const { pattern, raw } of cases) {
      expectInvariantFailure(raw, pattern);
    }
  });

  it("rejects missing and undeclared feature groups for current-profile surfaces", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v2");
    const raw = {
      ...fixtureRawContract(),
      view_schema_id: timeline.viewSchemaId,
      title: timeline.title,
      inspector_config: {
        ...fixtureRawContract().inspector_config,
        view_schema_id: timeline.viewSchemaId,
        panels: timeline.inspectorConfig.panels.map((panel) => ({
          panel_id: panel.panelId,
          label: panel.label,
        })),
      },
    };

    expectInvariantFailure(
      raw,
      /inspector_config\.feature_groups must contain exactly 27 declared feature groups for cartulary\.view\.timeline\.v2, got 2/,
    );
    expectInvariantFailure(
      {
        ...raw,
        inspector_config: {
          ...raw.inspector_config,
          feature_groups: timeline.inspectorConfig.featureGroups.map(
            (group, index) =>
              index === 0
                ? {
                    feature_group_key: "details.future",
                    panel_id: group.panelId,
                    label: group.label,
                    minimum_incident_role: group.minimumIncidentRole,
                    mutates: group.mutates,
                    requires_confirmation: group.requiresConfirmation,
                    route_binding: {
                      kind: group.routeBinding.kind,
                      owner: group.routeBinding.owner,
                      action_key: group.routeBinding.actionKey,
                      target_view_schema_id:
                        group.routeBinding.targetViewSchemaId,
                    },
                    seed_bindings: group.seedBindings.map((binding) => ({
                      target_field_key: binding.targetFieldKey,
                      source: {
                        kind: binding.source.kind,
                        source_field_key: binding.source.sourceFieldKey,
                        value: binding.source.value,
                      },
                    })),
                    disabled_when: group.disabledWhen,
                    success_result_behavior: group.successResultBehavior,
                    failure_result_behavior: group.failureResultBehavior,
                  }
                : {
                    feature_group_key: group.featureGroupKey,
                    panel_id: group.panelId,
                    label: group.label,
                    minimum_incident_role: group.minimumIncidentRole,
                    mutates: group.mutates,
                    requires_confirmation: group.requiresConfirmation,
                    route_binding: {
                      kind: group.routeBinding.kind,
                      owner: group.routeBinding.owner,
                      action_key: group.routeBinding.actionKey,
                      target_view_schema_id:
                        group.routeBinding.targetViewSchemaId,
                    },
                    seed_bindings: group.seedBindings.map((binding) => ({
                      target_field_key: binding.targetFieldKey,
                      source: {
                        kind: binding.source.kind,
                        source_field_key: binding.source.sourceFieldKey,
                        value: binding.source.value,
                      },
                    })),
                    disabled_when: group.disabledWhen,
                    success_result_behavior: group.successResultBehavior,
                    failure_result_behavior: group.failureResultBehavior,
                  },
          ),
        },
      },
      /inspector_config\.feature_groups missing required feature_group_key details\.read for cartulary\.view\.timeline\.v2/,
    );
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
          grid_editable: true,
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
          grid_editable: true,
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
