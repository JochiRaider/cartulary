import {
  type GridColumn,
  type GridRow,
  GridTable,
} from "@cartulary/grid-adapter/test-support";
import {
  dataTestIdSelector,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { WorkbookGridControls } from "./WorkbookGridControls";
import {
  applyFilterDraft,
  buildQueryRequest,
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  toggleSortField,
  updateGroupBy,
} from "./workbookQuery";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

describe("Phase 8 workbook query controls", () => {
  it("Phase 8 U-8-GRID-01 emits stable schema keys for sort, filter, and group query controls", () => {
    const contract = requireViewContract(timelineViewSchemaId);

    const sortedByHeader = toggleSortField(
      contract,
      emptyWorkbookQueryState(),
      "timeline.occurred_at",
    );
    expect(buildQueryRequest(contract, sortedByHeader)).toEqual({
      sort: [{ direction: "asc", field_key: "timeline.sort_ts" }],
    });

    const nonSortableCollectionOnly = toggleSortField(
      contract,
      emptyWorkbookQueryState(),
      "timeline.tags",
    );
    expect(buildQueryRequest(contract, nonSortableCollectionOnly)).toEqual({});

    const nonSortableCollectionHeader = toggleSortField(
      contract,
      sortedByHeader,
      "timeline.tags",
    );
    expect(nonSortableCollectionHeader).toBe(sortedByHeader);

    const grouped = updateGroupBy(
      contract,
      nonSortableCollectionHeader,
      "timeline.capture_state",
    );
    const filtered = applyFilterDraft(grouped, {
      booleanValue: "",
      fieldKey: "timeline.capture_state",
      value: "reviewed",
    });

    const request = buildQueryRequest(contract, filtered);
    expect(request).toEqual({
      filters: [
        {
          arg: { value: "reviewed" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [
        { direction: "asc", field_key: "timeline.capture_state" },
        { direction: "asc", field_key: "timeline.sort_ts" },
      ],
    });

    const encoded = JSON.stringify(request);
    expect(encoded).not.toContain("Occurred");
    expect(encoded).not.toContain("Capture State");
    expect(encoded).not.toContain("rdg");
    expect(encoded).not.toContain("column");
    expect(encoded).not.toContain("timeline_items");
    expect(encoded).not.toContain("timeline_grid_projection");
    expect(encoded).not.toContain("timeline_events");
    expect(encoded).not.toContain("record_id");
    expect(encoded).not.toContain("row_version");
  });

  it("FE-U-P8-01 compiles query requests with schema field keys and capability metadata", () => {
    const contract = requireViewContract(timelineViewSchemaId);

    const sortedByHeader = toggleSortField(
      contract,
      emptyWorkbookQueryState(),
      "timeline.occurred_at",
    );
    const ignoredLabelSort = toggleSortField(
      contract,
      sortedByHeader,
      "Occurred",
    );
    expect(ignoredLabelSort).toBe(sortedByHeader);

    const grouped = updateGroupBy(
      contract,
      sortedByHeader,
      "timeline.capture_state",
    );
    const invalidGroup = updateGroupBy(contract, grouped, "Capture State");
    expect(invalidGroup).toBe(grouped);

    const filtered = applyFilterDraft(grouped, {
      booleanValue: "",
      fieldKey: "timeline.capture_state",
      value: "reviewed",
    });

    expect(
      buildQueryRequest(contract, {
        ...filtered,
        filters: [
          ...filtered.filters,
          {
            arg: { value: "visible label" },
            fieldKey: "Capture State",
            op: "eq",
          },
          {
            arg: { value: "unsupported" },
            fieldKey: "timeline.has_evidence",
            op: "prefix",
          },
        ],
        sort: [
          ...filtered.sort,
          { direction: "asc", fieldKey: "timeline.tags" },
          { direction: "asc", fieldKey: "record_id" },
        ],
      }),
    ).toEqual({
      filters: [
        {
          arg: { value: "reviewed" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [
        { direction: "asc", field_key: "timeline.capture_state" },
        { direction: "asc", field_key: "timeline.sort_ts" },
      ],
    });
  });

  it("FE-U-P8-01 serializes saved-view query_json with canonical empty arrays and omitted inactive grouping", () => {
    const contract = requireViewContract(timelineViewSchemaId);

    expect(
      buildSavedViewQueryJson(contract, emptyWorkbookQueryState()),
    ).toEqual({
      filters: [],
      sort: [],
    });
    expect(
      Object.hasOwn(
        buildSavedViewQueryJson(contract, emptyWorkbookQueryState()),
        "group_by",
      ),
    ).toBe(false);

    const withSort = toggleSortField(
      contract,
      emptyWorkbookQueryState(),
      "timeline.occurred_at",
    );
    const withGroup = updateGroupBy(
      contract,
      withSort,
      "timeline.capture_state",
    );
    const withStateFilter = applyFilterDraft(withGroup, {
      booleanValue: "",
      fieldKey: "timeline.capture_state",
      value: "reviewed",
    });
    const withTagFilter = applyFilterDraft(withStateFilter, {
      booleanValue: "",
      fieldKey: "timeline.tags",
      value: "zeta, alpha, alpha",
    });

    expect(buildQueryRequest(contract, withTagFilter)).toMatchObject({
      sort: [
        { direction: "asc", field_key: "timeline.capture_state" },
        { direction: "asc", field_key: "timeline.sort_ts" },
      ],
    });
    expect(buildSavedViewQueryJson(contract, withTagFilter)).toEqual({
      filters: [
        {
          arg: { value: "reviewed" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
        {
          arg: { values: ["alpha", "zeta"] },
          field_key: "timeline.tags",
          op: "contains_any",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [{ direction: "asc", field_key: "timeline.sort_ts" }],
    });
  });

  it("FE-U-P8-01 serializes saved-view layout_json as portable schema field-key state", () => {
    const contract = requireViewContract(timelineViewSchemaId);
    const layout = buildSavedViewLayoutJson(contract, {
      columnOrder: [
        "row_version",
        "timeline.summary",
        "Capture State",
        "timeline.occurred_at",
      ],
      columnWidths: [
        { fieldKey: "timeline.summary", widthPx: 420 },
        { fieldKey: "row_version", widthPx: 55 },
        { fieldKey: "timeline.capture_state", widthPx: 39 },
        { fieldKey: "timeline.occurred_at", widthPx: 96 },
      ],
      hiddenFieldKeys: [
        "row_version",
        "timeline.details",
        "timeline.capture_state",
        "timeline.details",
      ],
    });

    expect(layout.layout_schema_id).toBe("cartulary.layout.v1");
    expect(layout.column_order.slice(0, 2)).toEqual([
      "timeline.summary",
      "timeline.occurred_at",
    ]);
    expect(new Set(layout.column_order)).toEqual(
      new Set(contract.fields.map((field) => field.fieldKey)),
    );
    expect(layout.column_order).not.toContain("row_version");
    expect(layout.column_order).not.toContain("Capture State");
    expect(layout.hidden_field_keys).toEqual([
      "timeline.capture_state",
      "timeline.details",
    ]);
    expect(layout.column_widths).toEqual([
      { field_key: "timeline.occurred_at", width_px: 96 },
      { field_key: "timeline.summary", width_px: 420 },
    ]);
    const encoded = JSON.stringify(layout);
    expect(layout.column_order).not.toContain("record_id");
    expect(layout.hidden_field_keys).not.toContain("record_id");
    expect(layout.hidden_field_keys).not.toContain("row_version");
    expect(layout.column_widths.map((width) => width.field_key)).not.toContain(
      "record_id",
    );
    expect(encoded).not.toContain('"record_id"');
    expect(encoded).not.toContain('"row_version"');
    expect(encoded).not.toContain('"saved_view_id"');
    expect(encoded).not.toContain("scroll");
    expect(encoded).not.toContain("focused");
    expect(encoded).not.toContain("presence");
    expect(encoded).not.toContain("Capture State");
  });

  it("FE-U-P8-01 renders active filter chips and grouping controls by field key", () => {
    const contract = requireViewContract(timelineViewSchemaId);
    const onFilterDraftChange = vi.fn();
    const onGroupByChange = vi.fn();
    const onRemoveFilter = vi.fn();

    render(
      <WorkbookGridControls
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        onApplyFilter={vi.fn()}
        onFilterDraftChange={onFilterDraftChange}
        onGroupByChange={onGroupByChange}
        onRemoveFilter={onRemoveFilter}
        queryState={{
          filters: [
            {
              arg: { value: "reviewed" },
              fieldKey: "timeline.capture_state",
              op: "eq",
            },
          ],
          groupBy: "timeline.capture_state",
          sort: [],
        }}
        surface={timelineViewSchemaId}
      />,
    );

    const grouping = screen.getByTestId(
      gridGroupingSelectTestId(timelineViewSchemaId),
    ) as HTMLSelectElement;
    expect(grouping.value).toBe("timeline.capture_state");
    expect([...grouping.options].map((option) => option.value)).toEqual([
      "",
      ...contract.groupingFields,
    ]);
    fireEvent.change(grouping, {
      target: { value: "timeline.has_evidence" },
    });
    expect(onGroupByChange).toHaveBeenCalledWith("timeline.has_evidence");

    const fieldSelect = screen.getByTestId(
      gridFilterFieldTestId(timelineViewSchemaId),
    ) as HTMLSelectElement;
    expect([...fieldSelect.options].map((option) => option.value)).toEqual(
      contract.filterFields,
    );

    fireEvent.click(
      screen.getByTestId(
        gridFilterChipTestId(timelineViewSchemaId, "timeline.capture_state"),
      ),
    );
    expect(onRemoveFilter).toHaveBeenCalledWith("timeline.capture_state");
  });

  it("Phase 8 U-8-GRID-01 drops non-discovery keys before building query request bodies", () => {
    const contract = requireViewContract("cartulary.view.timeline.v1");
    const request = buildQueryRequest(contract, {
      groupBy: "timeline_grid_projection.capture_state",
      sort: [
        { direction: "asc", fieldKey: "Capture State" },
        { direction: "asc", fieldKey: "record_id" },
      ],
      filters: [
        { arg: { value: "reviewed" }, fieldKey: "Capture State", op: "eq" },
        {
          arg: { value: "rough" },
          fieldKey: "timeline.capture_state",
          op: "eq",
        },
      ],
    });

    expect(request).toEqual({
      filters: [
        {
          arg: { value: "rough" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
    });
    const encoded = JSON.stringify(request);
    for (const forbidden of [
      "Capture State",
      "record_id",
      "row_version",
      "timeline_grid_projection",
      "timeline_events",
      "rdg",
    ]) {
      expect(encoded).not.toContain(forbidden);
    }
  });

  it("Phase 8 U-8-GRID-01 renders group headers as presentation-only rows", () => {
    type HarnessRow = {
      readonly actionTestId: string;
      readonly editableTestId: string;
      readonly recordId: string;
      readonly state: string;
      readonly summary: string;
    };
    const rows: readonly GridRow<HarnessRow>[] = [
      {
        key: "record-1",
        recordId: "record-1",
        data: {
          actionTestId: "action-record-1",
          editableTestId: "editable-record-1",
          recordId: "record-1",
          state: "rough",
          summary: "Alpha",
        },
        testId: "phase8-row-record-1",
      },
      {
        key: "record-2",
        recordId: "record-2",
        data: {
          actionTestId: "action-record-2",
          editableTestId: "editable-record-2",
          recordId: "record-2",
          state: "reviewed",
          summary: "Beta",
        },
        testId: "phase8-row-record-2",
      },
    ];
    const columns: readonly GridColumn<HarnessRow>[] = [
      {
        fieldKey: "timeline.summary",
        label: "Summary",
        renderCell: (row) => (
          <input data-testid={row.editableTestId} defaultValue={row.summary} />
        ),
      },
      {
        fieldKey: "timeline.capture_state",
        label: "State",
        renderCell: (row) => <span>{row.state}</span>,
      },
    ];

    const { container } = render(
      <GridTable
        actionsColumn={{
          label: "Actions",
          renderCell: (row) => (
            <button data-testid={row.data.actionTestId} type="button">
              Mutate
            </button>
          ),
        }}
        columns={columns}
        getGroupLabel={(row, fieldKey) =>
          fieldKey === "timeline.capture_state" ? row.state : null
        }
        getGroupRowTestId={(fieldKey, value) =>
          gridGroupRowTestId(timelineViewSchemaId, fieldKey, value)
        }
        groupBy="timeline.capture_state"
        rows={rows}
      />,
    );

    const groupHeader = container
      .querySelector(
        dataTestIdSelector(
          gridGroupRowTestId(
            timelineViewSchemaId,
            "timeline.capture_state",
            "rough",
          ),
        ),
      )
      ?.closest("tr");
    expect(groupHeader).not.toBeNull();
    expect(groupHeader?.getAttribute("data-grid-record-id")).toBeNull();
    expect(groupHeader?.querySelector("input,textarea,select")).toBeNull();
    expect(
      groupHeader?.querySelector(dataTestIdSelector("action-record-1")),
    ).toBeNull();
    expect(
      groupHeader?.querySelector(dataTestIdSelector("action-record-2")),
    ).toBeNull();
    expect(groupHeader?.querySelector("[data-grid-field-key]")).toBeNull();
    expect(
      container
        .querySelector(dataTestIdSelector("phase8-row-record-1"))
        ?.getAttribute("data-grid-record-id"),
    ).toBe("record-1");
  });

  it("FE-U-P8-01 keeps grouped presentation rows out of mutation-capable anchors", () => {
    type HarnessRow = {
      readonly actionTestId: string;
      readonly editableTestId: string;
      readonly recordId: string;
      readonly state: string;
      readonly summary: string;
    };
    const rows: readonly GridRow<HarnessRow>[] = [
      {
        key: "record-1",
        recordId: "record-1",
        data: {
          actionTestId: "action-record-1",
          editableTestId: "editable-record-1",
          recordId: "record-1",
          state: "rough",
          summary: "Alpha",
        },
        testId: "phase8-row-record-1",
      },
      {
        key: "record-2",
        recordId: "record-2",
        data: {
          actionTestId: "action-record-2",
          editableTestId: "editable-record-2",
          recordId: "record-2",
          state: "reviewed",
          summary: "Beta",
        },
        testId: "phase8-row-record-2",
      },
    ];
    const columns: readonly GridColumn<HarnessRow>[] = [
      {
        fieldKey: "timeline.summary",
        label: "Summary",
        renderCell: (row) => (
          <input data-testid={row.editableTestId} defaultValue={row.summary} />
        ),
      },
      {
        fieldKey: "timeline.capture_state",
        label: "State",
        renderCell: (row) => <span>{row.state}</span>,
      },
    ];

    const { container } = render(
      <GridTable
        actionsColumn={{
          label: "Actions",
          renderCell: (row) => (
            <button data-testid={row.data.actionTestId} type="button">
              Mutate
            </button>
          ),
        }}
        columns={columns}
        getGroupLabel={(row, fieldKey) =>
          fieldKey === "timeline.capture_state" ? row.state : null
        }
        getGroupRowTestId={(fieldKey, value) =>
          gridGroupRowTestId(timelineViewSchemaId, fieldKey, value)
        }
        groupBy="timeline.capture_state"
        rows={rows}
      />,
    );

    const groupHeader = container
      .querySelector(
        dataTestIdSelector(
          gridGroupRowTestId(
            timelineViewSchemaId,
            "timeline.capture_state",
            "rough",
          ),
        ),
      )
      ?.closest("tr");
    expect(groupHeader).not.toBeNull();
    expect(groupHeader?.getAttribute("data-grid-record-id")).toBeNull();
    expect(groupHeader?.querySelector("[data-grid-field-key]")).toBeNull();
    expect(groupHeader?.querySelector("input,textarea,select")).toBeNull();
    expect(
      groupHeader?.querySelector(dataTestIdSelector("action-record-1")),
    ).toBeNull();
    expect(
      groupHeader?.querySelector(dataTestIdSelector("action-record-2")),
    ).toBeNull();
    expect(
      container
        .querySelector(dataTestIdSelector("phase8-row-record-1"))
        ?.getAttribute("data-grid-record-id"),
    ).toBe("record-1");
  });

  it("Phase 8 E-8-04 Notes full_text controls submit the exact-token operator", () => {
    const contract = requireViewContract("cartulary.view.notes.v1");

    const filtered = applyFilterDraft(emptyWorkbookQueryState(), {
      booleanValue: "",
      fieldKey: "note.full_text",
      value: "shell alpha shell",
    });

    expect(buildQueryRequest(contract, filtered)).toEqual({
      filters: [
        {
          arg: { query: "shell alpha shell" },
          field_key: "note.full_text",
          op: "full_text",
        },
      ],
    });
  });
});
