import {
  type GridColumn,
  type GridDataRow,
  SemanticDataGrid,
} from "@cartulary/grid-adapter/test-support";
import {
  dataTestIdSelector,
  gridFilterFieldTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import {
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { WorkbookGridControls } from "./components/WorkbookGridControls";
import { workbookGridEditorKind } from "./components/WorkbookGridEditorControl";
import { defaultWorkbookLayoutState } from "./layout/workbookColumnLayout";
import {
  applyFilterDraft,
  buildQueryRequest,
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  toggleSortField,
  updateGroupBy,
  workbookLayoutStateFromSavedViewLayoutJson,
  workbookQueryStateFromSavedViewQueryJson,
} from "./models/workbookQuery";

const timelineViewSchemaId = "cartulary.view.timeline.v2";

describe("workbook query controls", () => {
  it("emits stable schema keys for sort, filter, and group query controls", () => {
    const contract = requireViewContract(timelineViewSchemaId);

    const sortedByHeader = toggleSortField(
      contract,
      emptyWorkbookQueryState(),
      "timeline.activity_utc_text",
    );
    expect(buildQueryRequest(contract, sortedByHeader)).toEqual({
      sort: [{ direction: "asc", field_key: "timeline.activity_sort_ts" }],
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
        { direction: "asc", field_key: "timeline.activity_sort_ts" },
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

  it("compiles query requests with schema field keys and capability metadata", () => {
    const contract = requireViewContract(timelineViewSchemaId);

    const sortedByHeader = toggleSortField(
      contract,
      emptyWorkbookQueryState(),
      "timeline.activity_utc_text",
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
        { direction: "asc", field_key: "timeline.activity_sort_ts" },
      ],
    });

    for (const viewContract of listViewContracts()) {
      for (const field of viewContract.fields) {
        if (field.gridEditable) {
          expect(workbookGridEditorKind(field), field.fieldKey).not.toBeNull();
        } else {
          expect(workbookGridEditorKind(field), field.fieldKey).toBeNull();
        }
      }
    }
    for (const viewSchemaId of [
      "cartulary.view.assessments.v1",
      "cartulary.view.indicators.v1",
    ]) {
      expect(
        requireViewContract(viewSchemaId).fields.every(
          (field) => workbookGridEditorKind(field) === null,
        ),
      ).toBe(true);
    }
  });

  it("serializes saved-view query_json with canonical empty arrays and omitted inactive grouping", () => {
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
      "timeline.activity_utc_text",
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
        { direction: "asc", field_key: "timeline.activity_sort_ts" },
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
      sort: [{ direction: "asc", field_key: "timeline.activity_sort_ts" }],
    });
  });

  it("serializes saved-view layout_json as portable schema field-key state", () => {
    const contract = requireViewContract(timelineViewSchemaId);
    const layout = buildSavedViewLayoutJson(contract, {
      columnOrder: [
        "row_version",
        "timeline.activity_synopsis_text",
        "Capture State",
        "timeline.activity_utc_text",
      ],
      columnWidths: [
        { fieldKey: "timeline.activity_synopsis_text", widthPx: 420 },
        { fieldKey: "row_version", widthPx: 55 },
        { fieldKey: "timeline.capture_state", widthPx: 39 },
        { fieldKey: "timeline.activity_utc_text", widthPx: 96 },
      ],
      hiddenFieldKeys: [
        "row_version",
        "timeline.raw_activity_text",
        "timeline.capture_state",
        "timeline.activity_time_pair_state",
      ],
    });

    expect(layout.layout_schema_id).toBe("cartulary.layout.v1");
    expect(layout.column_order.slice(0, 2)).toEqual([
      "timeline.activity_synopsis_text",
      "timeline.activity_utc_text",
    ]);
    expect(new Set(layout.column_order)).toEqual(
      new Set(contract.fields.map((field) => field.fieldKey)),
    );
    expect(layout.column_order).not.toContain("row_version");
    expect(layout.column_order).not.toContain("Capture State");
    expect(layout.hidden_field_keys).toEqual([
      "timeline.activity_time_pair_state",
      "timeline.capture_state",
      "timeline.raw_activity_text",
    ]);
    expect(layout.column_widths).toEqual([
      { field_key: "timeline.activity_synopsis_text", width_px: 420 },
      { field_key: "timeline.activity_utc_text", width_px: 96 },
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

  it("normalizes saved-view query_json and layout_json back to portable field-key state", () => {
    const contract = requireViewContract(timelineViewSchemaId);

    const queryState = workbookQueryStateFromSavedViewQueryJson(contract, {
      filters: [
        {
          field_key: "timeline.tags",
          op: "contains_any",
          arg: { values: ["zeta", "alpha", "alpha", ""] },
        },
        {
          field_key: "Capture State",
          op: "eq",
          arg: { value: "reviewed" },
        },
        {
          field_key: "timeline.capture_state",
          op: "eq",
          arg: { value: "reviewed" },
        },
      ],
      group_by: "timeline.capture_state",
      sort: [
        { field_key: "timeline.activity_sort_ts", direction: "desc" },
        { field_key: "timeline.tags", direction: "asc" },
        { field_key: "timeline.activity_sort_ts", direction: "asc" },
      ],
    });

    expect(buildSavedViewQueryJson(contract, queryState)).toEqual({
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
      sort: [{ direction: "desc", field_key: "timeline.activity_sort_ts" }],
    });

    expect(
      buildSavedViewLayoutJson(
        contract,
        workbookLayoutStateFromSavedViewLayoutJson(contract, {
          layout_schema_id: "cartulary.layout.v1",
          column_order: [
            "timeline.activity_synopsis_text",
            "row_version",
            "timeline.activity_utc_text",
          ],
          hidden_field_keys: ["timeline.raw_activity_text", "record_id"],
          column_widths: [
            { field_key: "timeline.activity_synopsis_text", width_px: 480 },
            { field_key: "row_version", width_px: 80 },
          ],
        }),
      ),
    ).toMatchObject({
      layout_schema_id: "cartulary.layout.v1",
      column_widths: [
        { field_key: "timeline.activity_synopsis_text", width_px: 480 },
      ],
      hidden_field_keys: ["timeline.raw_activity_text"],
    });
  });

  it("renders active filter chips and grouping controls by field key", () => {
    const contract = requireViewContract(timelineViewSchemaId);
    const longTagToken =
      "indicator-without-break-opportunities-0123456789-abcdefghijklmnopqrstuvwxyz";
    const onFilterDraftChange = vi.fn();
    const onGroupByChange = vi.fn();
    const onRemoveFilter = vi.fn();
    const onSortChange = vi.fn();

    const queryState = {
      filters: [
        {
          arg: { value: "reviewed" },
          fieldKey: "timeline.capture_state",
          op: "eq" as const,
        },
        {
          arg: { values: [longTagToken] },
          fieldKey: "timeline.tags",
          op: "contains_any" as const,
        },
      ],
      groupBy: "timeline.capture_state",
      sort: contract.sortFields.slice(0, 8).map((fieldKey, index) => ({
        direction: index % 2 === 0 ? ("asc" as const) : ("desc" as const),
        fieldKey,
      })),
    };
    const { rerender } = render(
      <WorkbookGridControls
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        onApplyFilter={vi.fn()}
        onFilterDraftChange={onFilterDraftChange}
        onGroupByChange={onGroupByChange}
        onColumnHiddenChange={vi.fn()}
        onColumnMove={vi.fn()}
        onResetColumns={vi.fn()}
        onRemoveFilter={onRemoveFilter}
        onSortChange={onSortChange}
        queryState={queryState}
        surface={timelineViewSchemaId}
      />,
    );

    const queryControls = screen.getByTestId(
      workbookViewBarQueryControlsTestId(timelineViewSchemaId),
    );
    expect(queryControls.style.overflow).toBe("visible");
    expect(queryControls.getAttribute("data-query-chip-capacity")).toBe("8");
    expect(queryControls.getAttribute("data-hidden-query-chip-count")).toBe(
      "3",
    );

    const expectedOrderedLabels = [
      "Group: Capture State",
      ...queryState.sort.map(
        (sort) =>
          `Sort: ${contract.fieldMap[sort.fieldKey]?.label ?? sort.fieldKey} ${sort.direction}`,
      ),
      "Capture State: reviewed",
      `Tags: ${longTagToken}`,
    ];
    expect(
      screen
        .getByRole("toolbar", { name: "Active query chips" })
        .querySelectorAll("button[title]").length,
    ).toBe(8);
    expect(
      Array.from(
        screen
          .getByRole("toolbar", { name: "Active query chips" })
          .querySelectorAll<HTMLButtonElement>("button[title]"),
        (button) => button.title,
      ),
    ).toEqual(expectedOrderedLabels.slice(0, 8));

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

    fireEvent.click(
      screen.getByTestId(workbookSortMenuTriggerTestId(timelineViewSchemaId)),
    );
    const sortMenu = screen.getByTestId(
      workbookSortMenuTestId(timelineViewSchemaId),
    );
    expect(sortMenu).toBeInstanceOf(HTMLElement);
    fireEvent.keyDown(sortMenu, { key: "End" });
    const enabledSortActions = sortMenu.querySelectorAll(
      '[role^="menuitem"]:not(:disabled)',
    );
    expect(document.activeElement).toBe(
      enabledSortActions[enabledSortActions.length - 1],
    );
    fireEvent.keyDown(sortMenu, { key: "Escape" });
    expect(
      screen.queryByTestId(workbookSortMenuTestId(timelineViewSchemaId)),
    ).toBeNull();
    expect(document.activeElement).toBe(
      screen.getByTestId(workbookSortMenuTriggerTestId(timelineViewSchemaId)),
    );

    fireEvent.click(
      screen.getByTestId(
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      ),
    );
    expect(
      screen
        .getByTestId(workbookFilterPopoverTriggerTestId(timelineViewSchemaId))
        .getAttribute("aria-label"),
    ).toBe("Filters, 3 hidden");
    expect(
      screen.getByTestId(workbookFilterPopoverTestId(timelineViewSchemaId)),
    ).toBeInstanceOf(HTMLElement);
    const fieldSelect = screen.getByTestId(
      gridFilterFieldTestId(timelineViewSchemaId),
    ) as HTMLSelectElement;
    expect([...fieldSelect.options].map((option) => option.value)).toEqual(
      contract.filterFields,
    );

    fireEvent.click(
      screen.getByRole("button", {
        name: "Remove Capture State: reviewed, hidden from the view bar",
      }),
    );
    expect(onRemoveFilter).toHaveBeenCalledWith("timeline.capture_state");

    expect(
      screen.getByRole("button", {
        name: `Remove Tags: ${longTagToken}, hidden from the view bar`,
      }).textContent,
    ).toContain(longTagToken);

    const groupChip = screen.getByTitle("Group: Capture State");
    expect(groupChip).toBeInstanceOf(HTMLButtonElement);
    const groupChipLabel = groupChip.firstElementChild;
    expect(groupChipLabel).toBeInstanceOf(HTMLElement);
    expect((groupChipLabel as HTMLElement).style.textOverflow).toBe("ellipsis");
    expect((groupChipLabel as HTMLElement).style.overflow).toBe("hidden");

    fireEvent.keyDown(
      screen.getByTestId(workbookFilterPopoverTestId(timelineViewSchemaId)),
      { key: "Escape" },
    );
    expect(
      screen.queryByTestId(workbookFilterPopoverTestId(timelineViewSchemaId)),
    ).toBeNull();
    expect(document.activeElement).toBe(
      screen.getByTestId(
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      ),
    );

    rerender(
      <WorkbookGridControls
        chromeMode="narrow_desktop"
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        onApplyFilter={vi.fn()}
        onFilterDraftChange={onFilterDraftChange}
        onGroupByChange={onGroupByChange}
        onColumnHiddenChange={vi.fn()}
        onColumnMove={vi.fn()}
        onResetColumns={vi.fn()}
        onRemoveFilter={onRemoveFilter}
        onSortChange={onSortChange}
        queryState={queryState}
        surface={timelineViewSchemaId}
      />,
    );
    const narrowQueryControls = screen.getByTestId(
      workbookViewBarQueryControlsTestId(timelineViewSchemaId),
    );
    expect(narrowQueryControls.getAttribute("data-query-chip-capacity")).toBe(
      "6",
    );
    expect(
      narrowQueryControls.getAttribute("data-hidden-query-chip-count"),
    ).toBe("5");
    expect(
      Array.from(
        screen
          .getByRole("toolbar", { name: "Active query chips" })
          .querySelectorAll<HTMLButtonElement>("button[title]"),
        (button) => button.title,
      ),
    ).toEqual(expectedOrderedLabels.slice(0, 6));

    rerender(
      <WorkbookGridControls
        chromeMode="compact_desktop"
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        onApplyFilter={vi.fn()}
        onFilterDraftChange={onFilterDraftChange}
        onGroupByChange={onGroupByChange}
        onColumnHiddenChange={vi.fn()}
        onColumnMove={vi.fn()}
        onResetColumns={vi.fn()}
        onRemoveFilter={onRemoveFilter}
        onSortChange={onSortChange}
        queryState={queryState}
        surface={timelineViewSchemaId}
      />,
    );
    const compactQueryControls = screen.getByTestId(
      workbookViewBarQueryControlsTestId(timelineViewSchemaId),
    );
    expect(compactQueryControls.getAttribute("data-query-chip-capacity")).toBe(
      "0",
    );
    expect(
      compactQueryControls.getAttribute("data-hidden-query-chip-count"),
    ).toBe("11");
    expect(
      screen
        .getByTestId(workbookFilterPopoverTriggerTestId(timelineViewSchemaId))
        .getAttribute("aria-label"),
    ).toBe("Filters, 11 hidden");
    expect(screen.queryByTitle("Group: Capture State")).toBeNull();
    fireEvent.click(
      screen.getByTestId(
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      ),
    );
    expect(
      screen.getAllByRole("button", {
        name: /hidden from the view bar/,
      }),
    ).toHaveLength(11);
  });

  it("drops non-discovery keys before building query request bodies", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
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

  it("renders group headers as presentation-only rows", () => {
    type HarnessRow = {
      readonly actionTestId: string;
      readonly editableTestId: string;
      readonly recordId: string;
      readonly state: string;
      readonly summary: string;
    };
    const rows: readonly GridDataRow<HarnessRow>[] = [
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-1" },
        data: {
          actionTestId: "action-record-1",
          editableTestId: "editable-record-1",
          recordId: "record-1",
          state: "rough",
          summary: "Alpha",
        },
        testId: "saved-view-row-record-1",
      },
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-2" },
        data: {
          actionTestId: "action-record-2",
          editableTestId: "editable-record-2",
          recordId: "record-2",
          state: "reviewed",
          summary: "Beta",
        },
        testId: "saved-view-row-record-2",
      },
    ];
    const columns: readonly GridColumn<HarnessRow>[] = [
      {
        fieldKey: "timeline.activity_synopsis_text",
        label: "Summary",
        renderCell: ({ row }) => (
          <input data-testid={row.editableTestId} defaultValue={row.summary} />
        ),
      },
      {
        fieldKey: "timeline.capture_state",
        label: "State",
        renderCell: ({ row }) => <span>{row.state}</span>,
      },
    ];

    const { container } = render(
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        actionsColumn={{
          label: "Actions",
          renderCell: ({ data: row }) => (
            <button data-testid={row.actionTestId} type="button">
              Mutate
            </button>
          ),
        }}
        columns={columns}
        grouping={{
          fieldKey: "timeline.capture_state",
          formatLabel: (value) => (value === null ? null : String(value)),
          getTestId: (fieldKey, _value, label) =>
            label === null
              ? undefined
              : gridGroupRowTestId(timelineViewSchemaId, fieldKey, label),
          getValue: (row) => row.state,
        }}
        dataRows={rows}
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
        .querySelector(dataTestIdSelector("saved-view-row-record-1"))
        ?.getAttribute("data-grid-record-id"),
    ).toBe("record-1");
  });

  it("keeps grouped presentation rows out of mutation-capable anchors", () => {
    type HarnessRow = {
      readonly actionTestId: string;
      readonly editableTestId: string;
      readonly recordId: string;
      readonly state: string;
      readonly summary: string;
    };
    const rows: readonly GridDataRow<HarnessRow>[] = [
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-1" },
        data: {
          actionTestId: "action-record-1",
          editableTestId: "editable-record-1",
          recordId: "record-1",
          state: "rough",
          summary: "Alpha",
        },
        testId: "saved-view-row-record-1",
      },
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-2" },
        data: {
          actionTestId: "action-record-2",
          editableTestId: "editable-record-2",
          recordId: "record-2",
          state: "reviewed",
          summary: "Beta",
        },
        testId: "saved-view-row-record-2",
      },
    ];
    const columns: readonly GridColumn<HarnessRow>[] = [
      {
        fieldKey: "timeline.activity_synopsis_text",
        label: "Summary",
        renderCell: ({ row }) => (
          <input data-testid={row.editableTestId} defaultValue={row.summary} />
        ),
      },
      {
        fieldKey: "timeline.capture_state",
        label: "State",
        renderCell: ({ row }) => <span>{row.state}</span>,
      },
    ];

    const { container } = render(
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        actionsColumn={{
          label: "Actions",
          renderCell: ({ data: row }) => (
            <button data-testid={row.actionTestId} type="button">
              Mutate
            </button>
          ),
        }}
        columns={columns}
        grouping={{
          fieldKey: "timeline.capture_state",
          formatLabel: (value) => (value === null ? null : String(value)),
          getTestId: (fieldKey, _value, label) =>
            label === null
              ? undefined
              : gridGroupRowTestId(timelineViewSchemaId, fieldKey, label),
          getValue: (row) => row.state,
        }}
        dataRows={rows}
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
        .querySelector(dataTestIdSelector("saved-view-row-record-1"))
        ?.getAttribute("data-grid-record-id"),
    ).toBe("record-1");
  });

  it("Notes full_text controls submit the exact-token operator", () => {
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
