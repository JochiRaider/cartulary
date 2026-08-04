import { describe, expect, expectTypeOf, it } from "vitest";

import * as publicFacade from "./index";
import {
  assessmentsViewSchemaId,
  buildWorkbookSurfaceContracts,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  fieldCapability,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  getViewContract,
  getWorkbookSurfaceContract,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  listViewContracts,
  listWorkbookSurfaceContracts,
  type NormalizedViewRowPatchV1,
  type NormalizedViewRowV1,
  normalizeViewRowPatchV1,
  normalizeViewRowV1,
  notesViewSchemaId,
  optionalStandardizedWorkbookSurfaceIds,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  requiredSystemWorkbookSurfaceIds,
  requireViewContract,
  requireWorkbookSurfaceContract,
  resolveHeaderSortFieldKey,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
  type ViewContract,
  visibleFields,
} from "./index";
import { parseViewContractJSON } from "./view-contracts";

const fixtureFieldDefaults = {
  default_hidden: false,
  sortable: false,
  header_sort_field_key: null,
  filter_ops: [],
  groupable: false,
  read_kind: "text",
  write_kind: "read_only",
  grid_editable: false,
  conflict_resolution_class: null,
  entity_binding_mode: null,
  string_contract_id: null,
  direct_scalar_contract_id: null,
  direct_reference_contract_id: null,
  clearable: false,
  enum_values: null,
  writable: false,
  read_model: "fixture_value",
} as const;

function fixtureRawContract() {
  return {
    $schema: "https://json-schema.org/draft/2020-12/schema",
    schema_id: "cartulary.view_schema_source.v1",
    view_schema_id: "cartulary.view.fixture.v1",
    title: "Fixture Surface",
    surface_kind: "system_view",
    source_record_types: ["fixture_record"],
    default_visible_fields: ["fixture.editable", "fixture.queryable"],
    default_hidden_fields: ["record_id", "row_version", "fixture.sort_shadow"],
    default_sort: [{ field_key: "fixture.sort_shadow", direction: "asc" }],
    sort_fields: ["fixture.sort_shadow", "fixture.queryable"],
    sort_null_order: "last",
    filter_fields: ["fixture.queryable"],
    grouping_fields: ["fixture.queryable"],
    technical_fields: ["record_id", "row_version"],
    required_reference_pack_keys: [],
    create_capable: false,
    create_inputs: [],
    inline_create: {
      minimum_create_field_sets: [],
      permits_zero_field_create: false,
    },
    fields: [
      {
        ...fixtureFieldDefaults,
        field_key: "fixture.editable",
        label: "Editable Field",
        write_kind: "direct_value",
        writable: true,
        grid_editable: true,
      },
      {
        ...fixtureFieldDefaults,
        field_key: "fixture.queryable",
        label: "Queryable Field",
        filter_ops: ["eq"],
        groupable: true,
        header_sort_field_key: "fixture.sort_shadow",
        sortable: true,
        grid_editable: false,
      },
      {
        ...fixtureFieldDefaults,
        field_key: "fixture.sort_shadow",
        label: "Sort Shadow",
        default_hidden: true,
        sortable: true,
        grid_editable: false,
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

function fixtureCells() {
  return {
    "fixture.editable": { value: "editable" },
    "fixture.queryable": { value: "bucket" },
    "fixture.sort_shadow": { value: "sort" },
  };
}

function sourceInspectorFeatureGroup(
  group: ViewContract["inspectorConfig"]["featureGroups"][number],
) {
  return {
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
      target_view_schema_id: group.routeBinding.targetViewSchemaId,
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
  };
}

describe("view-contracts characterization baseline", () => {
  it("exposes the supported public package facade", () => {
    for (const exportedFunction of [
      buildWorkbookSurfaceContracts,
      fieldCapability,
      getViewContract,
      getWorkbookSurfaceContract,
      listViewContracts,
      listWorkbookSurfaceContracts,
      normalizeViewRowPatchV1,
      normalizeViewRowV1,
      requireViewContract,
      requireWorkbookSurfaceContract,
      resolveHeaderSortFieldKey,
      visibleFields,
    ]) {
      expect(typeof exportedFunction).toBe("function");
    }
    expect(Object.keys(publicFacade).sort()).toEqual(
      [
        "assessmentsViewSchemaId",
        "buildWorkbookSurfaceContracts",
        "commLogViewSchemaId",
        "decisionsViewSchemaId",
        "evidenceViewSchemaId",
        "fieldCapability",
        "findingsViewSchemaId",
        "forensicKeywordsViewSchemaId",
        "getViewContract",
        "getWorkbookSurfaceContract",
        "handoffViewSchemaId",
        "hostsViewSchemaId",
        "identitiesViewSchemaId",
        "indicatorsViewSchemaId",
        "investigativeQueriesViewSchemaId",
        "lessonViewSchemaId",
        "listViewContracts",
        "listWorkbookSurfaceContracts",
        "normalizeViewRowPatchV1",
        "normalizeViewRowV1",
        "notesViewSchemaId",
        "optionalStandardizedWorkbookSurfaceIds",
        "partiesViewSchemaId",
        "requiredBuiltInWorkbookSurfaceIds",
        "requiredSystemWorkbookSurfaceIds",
        "requireViewContract",
        "requireWorkbookSurfaceContract",
        "resolveHeaderSortFieldKey",
        "statusReviewViewSchemaId",
        "taskRequestsViewSchemaId",
        "timelineViewSchemaId",
        "visibleFields",
      ].sort(),
    );
  });

  it("initializes all generated view artifacts by stable identity", () => {
    const contracts = listViewContracts();
    const ids = contracts.map((contract) => contract.viewSchemaId);

    expect(contracts).toHaveLength(17);
    expect(new Set(ids)).toHaveLength(17);
    for (const contract of contracts) {
      expect(getViewContract(contract.viewSchemaId)).toBe(contract);
      expect(requireViewContract(contract.viewSchemaId)).toBe(contract);
      expect(contract.inspectorConfig.viewSchemaId).toBe(contract.viewSchemaId);
    }
  });

  it("derives every schema constant and status partition in registry order", () => {
    const constants = [
      timelineViewSchemaId,
      hostsViewSchemaId,
      identitiesViewSchemaId,
      evidenceViewSchemaId,
      notesViewSchemaId,
      indicatorsViewSchemaId,
      assessmentsViewSchemaId,
      taskRequestsViewSchemaId,
      decisionsViewSchemaId,
      partiesViewSchemaId,
      commLogViewSchemaId,
      handoffViewSchemaId,
      statusReviewViewSchemaId,
      lessonViewSchemaId,
      findingsViewSchemaId,
      investigativeQueriesViewSchemaId,
      forensicKeywordsViewSchemaId,
    ];

    expect(constants).toEqual(
      listWorkbookSurfaceContracts().map((surface) => surface.viewSchemaId),
    );
    expect(requiredBuiltInWorkbookSurfaceIds).toEqual(constants.slice(0, 5));
    expect(requiredSystemWorkbookSurfaceIds).toEqual(constants.slice(5, 14));
    expect(optionalStandardizedWorkbookSurfaceIds).toEqual(constants.slice(14));
    expect(Object.isFrozen(requiredBuiltInWorkbookSurfaceIds)).toBe(true);
    expect(Object.isFrozen(requiredSystemWorkbookSurfaceIds)).toBe(true);
    expect(Object.isFrozen(optionalStandardizedWorkbookSurfaceIds)).toBe(true);
  });

  it("freezes shared contracts, surfaces, inspector metadata, and normalized rows", () => {
    const contract = parseFixture();
    const surfaces = listWorkbookSurfaceContracts();
    const row = normalizeViewRowV1(contract, {
      record_id: "record-1",
      row_version: 1,
      cells: fixtureCells(),
      group_values: { "fixture.queryable": "bucket" },
    });

    expect(Object.isFrozen(listViewContracts())).toBe(true);
    expect(Object.isFrozen(contract)).toBe(true);
    expect(Object.isFrozen(contract.fields)).toBe(true);
    expect(Object.isFrozen(contract.fieldMap)).toBe(true);
    expect(Object.isFrozen(contract.inspectorConfig)).toBe(true);
    expect(Object.isFrozen(contract.inspectorConfig.panels)).toBe(true);
    expect(Object.isFrozen(contract.inspectorConfig.panels[0])).toBe(true);
    expect(Object.isFrozen(contract.inspectorConfig.featureGroups)).toBe(true);
    expect(Object.isFrozen(contract.inspectorConfig.featureGroups[0])).toBe(
      true,
    );
    expect(Object.isFrozen(surfaces)).toBe(true);
    expect(Object.isFrozen(surfaces[0])).toBe(true);
    expect(Object.isFrozen(row)).toBe(true);
    expect(Object.isFrozen(row.cells)).toBe(true);
    expect(Object.isFrozen(row.cells["fixture.editable"])).toBe(true);
    expect(Object.isFrozen(row.groupValues)).toBe(true);
  });

  it("derives the four exact Indicator feature signatures from the owner registry", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v2");
    const indicators = requireViewContract("cartulary.view.indicators.v1");
    const specializedKeys = [
      "indicator.observations.manage",
      "indicator.observations.pivot",
      "indicator.lifecycle.read",
      "indicator.lifecycle.manage",
    ];
    const specialized = [timeline, indicators]
      .flatMap((contract) => contract.inspectorConfig.featureGroups)
      .filter((group) => specializedKeys.includes(group.featureGroupKey));

    expect(
      specialized.map((group) => ({
        actionKey: group.routeBinding.actionKey,
        disabledWhen: group.disabledWhen,
        failure: group.failureResultBehavior,
        featureGroupKey: group.featureGroupKey,
        kind: group.routeBinding.kind,
        mutates: group.mutates,
        owner: group.routeBinding.owner,
        panelId: group.panelId,
        requiresConfirmation: group.requiresConfirmation,
        role: group.minimumIncidentRole,
        seeds: group.seedBindings,
        success: group.successResultBehavior,
        target: group.routeBinding.targetViewSchemaId,
      })),
    ).toEqual([
      {
        actionKey: "indicator.observations.manage",
        disabledWhen: [
          "no_row_selected",
          "incident_closed",
          "authorization_lost",
          "row_version_changed",
          "record_deleted",
        ],
        failure: "show_same_shell_error_invalidate_pending_action",
        featureGroupKey: "indicator.observations.manage",
        kind: "indicator_observations",
        mutates: true,
        owner: "indicator_observations_route",
        panelId: "relationships",
        requiresConfirmation: false,
        role: "editor",
        seeds: [],
        success: "preserve_selected_row",
        target: undefined,
      },
      {
        actionKey: "indicator.observations.pivot",
        disabledWhen: [
          "no_row_selected",
          "authorization_lost",
          "record_deleted",
        ],
        failure: "show_same_shell_error_preserve_selection",
        featureGroupKey: "indicator.observations.pivot",
        kind: "indicator_observations",
        mutates: false,
        owner: "indicator_observations_route",
        panelId: "relationships",
        requiresConfirmation: false,
        role: null,
        seeds: [],
        success: "preserve_selected_row",
        target: undefined,
      },
      {
        actionKey: "indicator.lifecycle.read",
        disabledWhen: [
          "no_row_selected",
          "authorization_lost",
          "record_deleted",
        ],
        failure: "show_same_shell_error_preserve_selection",
        featureGroupKey: "indicator.lifecycle.read",
        kind: "indicator_lifecycle",
        mutates: false,
        owner: "indicator_lifecycle_route",
        panelId: "history",
        requiresConfirmation: false,
        role: null,
        seeds: [],
        success: "preserve_selected_row",
        target: undefined,
      },
      {
        actionKey: "indicator.lifecycle.manage",
        disabledWhen: [
          "no_row_selected",
          "incident_closed",
          "authorization_lost",
          "row_version_changed",
          "record_deleted",
        ],
        failure: "show_same_shell_error_invalidate_pending_action",
        featureGroupKey: "indicator.lifecycle.manage",
        kind: "indicator_lifecycle",
        mutates: true,
        owner: "indicator_lifecycle_route",
        panelId: "history",
        requiresConfirmation: false,
        role: "editor",
        seeds: [],
        success: "preserve_selected_row",
        target: undefined,
      },
    ]);
    expect(specialized).toHaveLength(4);
    expect(
      specialized.some((group) => group.routeBinding.kind === "record_patch"),
    ).toBe(false);
  });

  it("rejects invalid JSON and malformed document roots", () => {
    expect(() => parseViewContractJSON("{", "invalid.json")).toThrow(
      "View contract source validation failed: invalid.json path=$ reason=invalid_json",
    );
    for (const json of ["null", "[]", '"not an object"']) {
      expect(() => parseViewContractJSON(json, "invalid-root.json")).toThrow(
        "View contract source validation failed: invalid-root.json path=$ reason=invalid_type",
      );
    }
  });

  it("rejects missing, mistyped, unknown, and invalid source-schema members", () => {
    const { schema_id: _schemaId, ...missingSchemaId } = fixtureRawContract();
    const { title: _title, ...missingTitle } = fixtureRawContract();
    const cases = [
      {
        raw: missingSchemaId,
        path: "$/schema_id",
        reason: "required_member",
      },
      {
        raw: missingTitle,
        path: "$/title",
        reason: "required_member",
      },
      {
        raw: { ...fixtureRawContract(), schema_id: "legacy.schema.v0" },
        path: "$/schema_id",
        reason: "invalid_value",
      },
      {
        raw: { ...fixtureRawContract(), surface_kind: "legacy_surface" },
        path: "$/surface_kind",
        reason: "invalid_value",
      },
      {
        raw: { ...fixtureRawContract(), fields: "not-an-array" },
        path: "$/fields",
        reason: "invalid_type",
      },
      {
        raw: { ...fixtureRawContract(), legacy_member: true },
        path: "$/legacy_member",
        reason: "unknown_member",
      },
    ] as const;

    for (const { path, raw, reason } of cases) {
      expectInvariantFailure(
        raw,
        new RegExp(
          `^View contract source validation failed: broken-contract\\.json path=${path.replaceAll("$", "\\$")} reason=${reason}$`,
        ),
      );
    }
  });

  it("rejects missing and malformed row identity and version values", () => {
    const contract = parseFixture();
    const base = {
      record_id: "record-1",
      row_version: 1,
      cells: fixtureCells(),
      group_values: { "fixture.queryable": "bucket" },
    };

    for (const recordId of [undefined, null, "", "   ", 1]) {
      expect(() =>
        normalizeViewRowV1(contract, { ...base, record_id: recordId }),
      ).toThrow(/record_id must be a non-empty string/);
    }
    for (const rowVersion of [
      undefined,
      null,
      "1",
      0,
      -1,
      1.5,
      Number.MAX_SAFE_INTEGER + 1,
      Infinity,
    ]) {
      expect(() =>
        normalizeViewRowV1(contract, { ...base, row_version: rowVersion }),
      ).toThrow(/row_version must be a positive safe integer/);
      expect(() =>
        normalizeViewRowPatchV1(contract, {
          ...base,
          cells: {},
          row_version: rowVersion,
        }),
      ).toThrow(/row_version must be a positive safe integer/);
    }
  });

  it("keeps normalized full rows and sparse patches non-assignable", () => {
    expectTypeOf<NormalizedViewRowPatchV1>().not.toMatchTypeOf<NormalizedViewRowV1>();
    expectTypeOf<NormalizedViewRowV1>().not.toMatchTypeOf<NormalizedViewRowPatchV1>();
  });

  it("requires complete full-row cells and rejects technical cells", () => {
    const contract = parseFixture();
    const { "fixture.sort_shadow": _missing, ...incompleteCells } =
      fixtureCells();

    expect(() =>
      normalizeViewRowV1(contract, {
        record_id: "record-1",
        row_version: 1,
        cells: incompleteCells,
        group_values: { "fixture.queryable": "bucket" },
      }),
    ).toThrow(/missing cell fixture\.sort_shadow/);
    expect(() =>
      normalizeViewRowV1(contract, {
        record_id: "record-1",
        row_version: 1,
        cells: { ...fixtureCells(), record_id: { value: "record-1" } },
        group_values: { "fixture.queryable": "bucket" },
      }),
    ).toThrow(/technical cell record_id is not allowed/);
  });

  it("ignores additive row and cell members while preserving wire-derived values", () => {
    const contract = parseFixture();
    const normalized = normalizeViewRowV1(contract, {
      record_id: "record-1",
      row_version: 1,
      view_schema_id: contract.viewSchemaId,
      future_row_member: "ignored",
      cells: {
        ...fixtureCells(),
        "fixture.editable": {
          value: { nested: true },
          future_cell_member: "ignored",
        },
      },
      group_values: { "fixture.queryable": "bucket" },
    });

    expect(normalized).toEqual({
      recordId: "record-1",
      rowVersion: 1,
      viewSchemaId: contract.viewSchemaId,
      cells: {
        "fixture.editable": { value: { nested: true } },
        "fixture.queryable": { value: "bucket" },
        "fixture.sort_shadow": { value: "sort" },
      },
      groupValues: { "fixture.queryable": "bucket" },
    });
    expect(normalized).not.toHaveProperty("future_row_member");
    expect(normalized.cells["fixture.editable"]).not.toHaveProperty(
      "future_cell_member",
    );
  });

  it("accepts sparse patch cells without weakening full-row completeness", () => {
    const contract = parseFixture();
    const patch = normalizeViewRowPatchV1(contract, {
      record_id: "record-1",
      row_version: 2,
      cells: { "fixture.editable": { value: "changed" } },
    });

    expect(patch.cells).toEqual({
      "fixture.editable": { value: "changed" },
    });
    expect(patch.groupValues).toBeUndefined();
    expect(() =>
      normalizeViewRowV1(contract, {
        record_id: "record-1",
        row_version: 2,
        cells: patch.cells,
        group_values: { "fixture.queryable": "bucket" },
      }),
    ).toThrow(/missing cell fixture\.queryable/);
  });

  it("rejects row view identity mismatches deterministically", () => {
    const contract = parseFixture();

    expect(() =>
      normalizeViewRowV1(
        contract,
        {
          record_id: "record-1",
          row_version: 1,
          view_schema_id: "cartulary.view.other.v1",
          cells: fixtureCells(),
          group_values: { "fixture.queryable": "bucket" },
        },
        "fixture-row.json",
      ),
    ).toThrow(
      "View row invariant failed: fixture-row.json view_schema_id must be cartulary.view.fixture.v1",
    );
  });

  it("fails for missing required surfaces and omits missing optional surfaces", () => {
    const allContracts = listViewContracts();
    const requiredId = "cartulary.view.timeline.v2";
    const optionalId = "cartulary.view.findings.v1";

    expect(() =>
      buildWorkbookSurfaceContracts(
        allContracts.filter((contract) => contract.viewSchemaId !== requiredId),
      ),
    ).toThrow(`Missing workbook surface contract: ${requiredId}`);
    expect(
      buildWorkbookSurfaceContracts(
        allContracts.filter((contract) => contract.viewSchemaId !== optionalId),
      ).map((surface) => surface.viewSchemaId),
    ).not.toContain(optionalId);
  });

  it("rejects workbook registry mismatches and preserves registry order", () => {
    const allContracts = listViewContracts();
    const first = allContracts[0];
    if (!first) {
      throw new Error("expected at least one generated view contract");
    }
    const mismatched: ViewContract = {
      ...first,
      surfaceKind:
        first.surfaceKind === "built_in_sheet"
          ? "system_view"
          : "built_in_sheet",
    };

    expect(() =>
      buildWorkbookSurfaceContracts([mismatched, ...allContracts.slice(1)]),
    ).toThrow(
      `Workbook surface ${first.viewSchemaId} has surface_kind ${mismatched.surfaceKind}, expected ${first.surfaceKind}`,
    );
    expect(() =>
      buildWorkbookSurfaceContracts([
        { ...first, requiredReferencePackKeys: ["unexpected_pack"] },
        ...allContracts.slice(1),
      ]),
    ).toThrow(
      `Workbook surface ${first.viewSchemaId} required_reference_pack_keys do not match its registry entry`,
    );
    expect(
      buildWorkbookSurfaceContracts(allContracts).map(
        (surface) => surface.viewSchemaId,
      ),
    ).toEqual(
      listWorkbookSurfaceContracts().map((surface) => surface.viewSchemaId),
    );
  });
});

describe("view-contracts", () => {
  it("requires explicit boolean grid_editable inputs", () => {
    const parsed = parseFixture();
    expect(
      parsed.fields.every((field) => typeof field.gridEditable === "boolean"),
    ).toBe(true);

    for (const gridEditable of [undefined, null, "false"]) {
      expectInvariantFailure(
        {
          ...fixtureRawContract(),
          fields: [
            {
              ...fixtureRawContract().fields[0],
              grid_editable: gridEditable,
            },
          ],
        },
        gridEditable === undefined
          ? /View contract source validation failed: broken-contract\.json path=\$\/fields\/0\/grid_editable reason=required_member/
          : /View contract source validation failed: broken-contract\.json path=\$\/fields\/0\/grid_editable reason=invalid_type/,
      );
    }
  });

  it("requires an exact inline_create policy", () => {
    const { inline_create: _omitted, ...withoutInlineCreate } =
      fixtureRawContract();
    const missingMinimum = {
      permits_zero_field_create: false,
    };
    const missingPermission = {
      minimum_create_field_sets: [],
    };

    expectInvariantFailure(
      withoutInlineCreate,
      /path=\$\/inline_create reason=required_member/,
    );
    expectInvariantFailure(
      { ...fixtureRawContract(), inline_create: null },
      /path=\$\/inline_create reason=invalid_type/,
    );
    expectInvariantFailure(
      { ...fixtureRawContract(), inline_create: missingMinimum },
      /path=\$\/inline_create\/minimum_create_field_sets reason=required_member/,
    );
    expectInvariantFailure(
      { ...fixtureRawContract(), inline_create: missingPermission },
      /path=\$\/inline_create\/permits_zero_field_create reason=required_member/,
    );
    expectInvariantFailure(
      {
        ...fixtureRawContract(),
        inline_create: {
          ...fixtureRawContract().inline_create,
          legacy_default: false,
        },
      },
      /path=\$\/inline_create\/legacy_default reason=unknown_member/,
    );
  });

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
    const cells = fixtureCells();
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

  it("Verify inspector panels and feature groups render from active config keys and reject unknown panel IDs, route owners, result behaviors, disabled tokens, duplicate keys, missing required keys, and extra current-profile keys before rendering.", () => {
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
        pattern: /path=\$\/inspector_config reason=required_member/,
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
          /route_binding\.kind must be one of panel_read\|view_row_create\|record_patch\|record_action\|entity_mention_action\|evidence_access\|surface_pivot\|indicator_observations\|indicator_lifecycle/,
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
          /route_binding\.owner must be one of current_row_projection\|view_query_route\|view_row_create_route\|record_patch_route\|record_mark_reviewed_route\|record_supersede_route\|record_delete_route\|record_restore_route\|record_history_route\|record_rollback_route\|record_merge_route\|entity_mention_resolve_route\|indicator_observations_route\|indicator_lifecycle_route\|evidence_attach_blob_route\|evidence_preview_handle_route\|evidence_download_handle_route/,
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
        pattern:
          /path=\$\/inspector_config\/feature_groups\/0\/feature_group_key reason=constraint_violation/,
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
      /inspector_config\.feature_groups must contain exactly 27 ordered feature groups for cartulary\.view\.timeline\.v2, got 2/,
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
      /inspector_config\.feature_groups\[1\]\.feature_group_key must be details\.read for cartulary\.view\.timeline\.v2/,
    );
    expectInvariantFailure(
      {
        ...raw,
        inspector_config: {
          ...raw.inspector_config,
          feature_groups: timeline.inspectorConfig.featureGroups.map(
            (group) => {
              const sourceGroup = sourceInspectorFeatureGroup(group);
              return group.featureGroupKey === "indicator.observations.manage"
                ? {
                    ...sourceGroup,
                    route_binding: {
                      action_key: group.featureGroupKey,
                      kind: "record_patch",
                      owner: "record_patch_route",
                    },
                  }
                : sourceGroup;
            },
          ),
        },
      },
      /specialized feature_group_key indicator\.observations\.manage does not match the owner registry/,
    );
  });
});

describe("view-schema field-key adapter contract", () => {
  it("selects generated contracts by view_schema_id, not display title", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v2");

    expect(timeline.viewSchemaId).toBe("cartulary.view.timeline.v2");
    expect(getViewContract(timeline.title)).toBeUndefined();
    expect(getViewContract("Timeline")).toBeUndefined();
  });

  it("keeps field identity and capabilities stable when labels change", () => {
    const first = parseFixture();
    const relabeled = parseFixture({
      ...fixtureRawContract(),
      title: "Renamed Surface",
      fields: [
        {
          ...fixtureFieldDefaults,
          field_key: "fixture.editable",
          label: "Editable Field Renamed",
          write_kind: "direct_value",
          writable: true,
          grid_editable: true,
        },
        {
          ...fixtureFieldDefaults,
          field_key: "fixture.queryable",
          label: "Queryable Field Renamed",
          filter_ops: ["eq"],
          groupable: true,
          header_sort_field_key: "fixture.sort_shadow",
          sortable: true,
          grid_editable: false,
        },
        {
          ...fixtureFieldDefaults,
          field_key: "fixture.sort_shadow",
          label: "Sort Shadow Renamed",
          default_hidden: true,
          sortable: true,
          grid_editable: false,
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

  it("treats duplicate labels as display-only with distinct field_key identities", () => {
    const duplicateLabels = parseFixture({
      ...fixtureRawContract(),
      fields: [
        {
          ...fixtureFieldDefaults,
          field_key: "fixture.editable",
          label: "Shared Display Label",
          write_kind: "direct_value",
          writable: true,
          grid_editable: true,
        },
        {
          ...fixtureFieldDefaults,
          field_key: "fixture.queryable",
          label: "Shared Display Label",
          filter_ops: ["eq"],
          groupable: true,
          header_sort_field_key: "fixture.sort_shadow",
          sortable: true,
          grid_editable: false,
        },
        {
          ...fixtureFieldDefaults,
          field_key: "fixture.sort_shadow",
          label: "Sort Shadow",
          default_hidden: true,
          sortable: true,
          grid_editable: false,
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

  it("keeps field identity stable when fields and visible columns reorder", () => {
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

  it("resolves editable, queryable, and synthetic fields by field_key", () => {
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

  it("fails deterministically for duplicate and missing field_key inputs", () => {
    expectInvariantFailure(
      {
        ...fixtureRawContract(),
        fields: [
          ...fixtureRawContract().fields,
          {
            ...fixtureFieldDefaults,
            field_key: "fixture.queryable",
            label: "Duplicate Field",
            grid_editable: false,
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
            ...fixtureFieldDefaults,
            field_key: "",
            label: "Missing Key",
          },
        ],
      },
      /View contract source validation failed: broken-contract\.json path=\$\/fields\/0\/field_key reason=constraint_violation/,
    );

    expectInvariantFailure(
      {
        ...fixtureRawContract(),
        synthetic_filter_predicates: [
          {
            field_key: "",
            label: "Missing Synthetic Key",
            filter_ops: [],
          },
        ],
      },
      /View contract source validation failed: broken-contract\.json path=\$\/synthetic_filter_predicates\/0\/field_key reason=constraint_violation/,
    );
  });

  it("fails deterministically for unknown field_key references", () => {
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

  it("visibleFields fails deterministically when default-visible keys are unresolved", () => {
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
