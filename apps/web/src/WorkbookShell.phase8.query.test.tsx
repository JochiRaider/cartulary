import {
  type GridColumn,
  type GridRow,
  GridTable,
} from "@cartulary/grid-adapter/test-support";
import { gridGroupRowTestId } from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import {
  applyFilterDraft,
  buildQueryRequest,
  emptyWorkbookQueryState,
  toggleSortField,
  updateGroupBy,
} from "./workbookQuery";

describe("Phase 8 workbook query controls", () => {
  it("Phase 8 U-8-GRID-01 emits stable schema keys for sort, filter, and group query controls", () => {
    const contract = requireViewContract("cartulary.view.timeline.v1");

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
      readonly recordId: string;
      readonly state: string;
      readonly summary: string;
    };
    const rows: readonly GridRow<HarnessRow>[] = [
      {
        key: "record-1",
        recordId: "record-1",
        data: { recordId: "record-1", state: "rough", summary: "Alpha" },
        testId: "phase8-row-record-1",
      },
      {
        key: "record-2",
        recordId: "record-2",
        data: { recordId: "record-2", state: "reviewed", summary: "Beta" },
        testId: "phase8-row-record-2",
      },
    ];
    const columns: readonly GridColumn<HarnessRow>[] = [
      {
        fieldKey: "timeline.summary",
        label: "Summary",
        renderCell: (row) => (
          <input
            data-testid={`editable-${row.recordId}`}
            defaultValue={row.summary}
          />
        ),
      },
      {
        fieldKey: "timeline.capture_state",
        label: "State",
        renderCell: (row) => <span>{row.state}</span>,
      },
    ];

    render(
      <GridTable
        actionsColumn={{
          label: "Actions",
          renderCell: (row) => (
            <button
              data-testid={`action-${row.recordId ?? "none"}`}
              type="button"
            >
              Mutate
            </button>
          ),
        }}
        columns={columns}
        getGroupLabel={(row, fieldKey) =>
          fieldKey === "timeline.capture_state" ? row.state : null
        }
        getGroupRowTestId={(fieldKey, value) =>
          gridGroupRowTestId("timeline", fieldKey, value)
        }
        groupBy="timeline.capture_state"
        rows={rows}
      />,
    );

    const groupHeader = screen
      .getByTestId(
        gridGroupRowTestId("timeline", "timeline.capture_state", "rough"),
      )
      .closest("tr");
    expect(groupHeader).not.toBeNull();
    expect(groupHeader?.getAttribute("data-grid-record-id")).toBeNull();
    expect(groupHeader?.querySelector("input,textarea,select")).toBeNull();
    expect(groupHeader?.querySelector("[data-testid^='action-']")).toBeNull();
    expect(groupHeader?.querySelector("[data-grid-field-key]")).toBeNull();
    expect(
      screen
        .getByTestId("phase8-row-record-1")
        .getAttribute("data-grid-record-id"),
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
