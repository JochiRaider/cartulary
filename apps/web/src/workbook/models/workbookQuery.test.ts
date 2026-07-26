import type { GridColumn } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  applyWorkbookLayoutToColumns,
  defaultWorkbookLayoutState,
  moveWorkbookColumn,
  reorderWorkbookColumns,
  setWorkbookColumnHidden,
  setWorkbookColumnWidth,
} from "./workbookLayout";
import {
  applyFilterDraft,
  buildQueryRequest,
  cycleWorkbookSortField,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  filterChipLabel,
  filterInputMode,
  toggleSortField,
  updateGroupBy,
} from "./workbookQuery";
import {
  workbookContractForViewSchemaId,
  workbookQuerySurfaceSlot,
} from "./workbookSurfaceQueryRuntime";

describe("workbookQuery", () => {
  it("builds tag and boolean filters from the client-local draft state", () => {
    const state = applyFilterDraft(emptyWorkbookQueryState(), {
      booleanValue: "",
      fieldKey: "timeline.tags",
      value: "phish, c2",
    });

    expect(state.filters).toEqual([
      {
        fieldKey: "timeline.tags",
        op: "contains_any",
        arg: {
          values: ["c2", "phish"],
        },
      },
    ]);

    expect(
      applyFilterDraft(emptyWorkbookQueryState(), {
        booleanValue: "true",
        fieldKey: "timeline.has_evidence",
        value: "",
      }).filters,
    ).toEqual([
      {
        fieldKey: "timeline.has_evidence",
        op: "eq",
        arg: {
          value: true,
        },
      },
    ]);
  });

  it("prepends group-by sorting when the user sort does not already cluster grouped rows", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const next = updateGroupBy(
      contract,
      toggleSortField(
        contract,
        emptyWorkbookQueryState(),
        "timeline.activity_synopsis_text",
      ),
      "timeline.capture_state",
    );

    expect(buildQueryRequest(contract, next)).toEqual({
      group_by: "timeline.capture_state",
      sort: [
        { field_key: "timeline.capture_state", direction: "asc" },
        { field_key: "timeline.activity_synopsis_text", direction: "asc" },
      ],
    });
  });

  it("supports ordered additive sort cycles and enforces the owner limit", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const sortable = contract.fields
      .map((field) => field.fieldKey)
      .filter((fieldKey) => contract.sortableFieldMap[fieldKey])
      .slice(0, 9);
    let state = emptyWorkbookQueryState();
    for (const fieldKey of sortable) {
      state = cycleWorkbookSortField(contract, state, fieldKey, true);
    }
    expect(state.sort).toHaveLength(Math.min(8, sortable.length));
    const first = state.sort[0];
    expect(first).toBeDefined();
    if (first === undefined) return;
    const descending = cycleWorkbookSortField(
      contract,
      state,
      first.fieldKey,
      true,
    );
    expect(descending.sort[0]).toEqual({ ...first, direction: "desc" });
    const removed = cycleWorkbookSortField(
      contract,
      descending,
      first.fieldKey,
      true,
    );
    expect(
      removed.sort.some((entry) => entry.fieldKey === first.fieldKey),
    ).toBe(false);
  });

  it("describes filter chip labels from contract metadata", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    expect(
      filterChipLabel(contract, {
        fieldKey: "timeline.capture_state",
        op: "eq",
        arg: { value: "reviewed" },
      }),
    ).toBe("Capture State: reviewed");
  });

  it("initializes filter drafts from the contract and resolves input modes", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    expect(defaultFilterDraft(contract).fieldKey).toBe(
      "timeline.date_entered_sort_day",
    );
    expect(filterInputMode("timeline.date_entered_sort_day")).toBe("date");
    expect(filterInputMode("timeline.tags")).toBe("tagset");
    expect(workbookQuerySurfaceSlot("cartulary.view.timeline.v2")).toBe(
      "timeline",
    );
    expect(workbookQuerySurfaceSlot("cartulary.view.hosts.v1")).toBe("hosts");
    expect(workbookQuerySurfaceSlot("cartulary.view.assessments.v1")).toBe(
      "assessments",
    );
    expect(workbookQuerySurfaceSlot("cartulary.view.notes.v1")).toBe("generic");
    expect(
      workbookContractForViewSchemaId("cartulary.view.notes.v1").viewSchemaId,
    ).toBe("cartulary.view.notes.v1");
    expect(() =>
      workbookContractForViewSchemaId("cartulary.view.unknown.v1"),
    ).toThrow("Unknown workbook view schema");
  });

  it("keeps a full semantic layout permutation while compiling visible columns", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const initial = defaultWorkbookLayoutState(contract);
    const [first, second] = initial.columnOrder;
    expect(first).toBeDefined();
    expect(second).toBeDefined();
    if (first === undefined || second === undefined) return;
    const reordered = reorderWorkbookColumns(contract, initial, second, first);
    const hidden = setWorkbookColumnHidden(contract, reordered, first, true);
    const sized = setWorkbookColumnWidth(contract, hidden, second, 480);
    const columns = contract.fields.map(
      (field): GridColumn<Record<string, never>> => ({
        fieldKey: field.fieldKey,
        label: field.label,
        renderCell: () => null,
      }),
    );

    expect(sized.columnOrder).toHaveLength(contract.fields.length);
    expect(sized.columnOrder.slice(0, 2)).toEqual([second, first]);
    expect(
      applyWorkbookLayoutToColumns(contract, columns, sized)
        .slice(0, 1)
        .map((column) => ({ key: column.fieldKey, width: column.width })),
    ).toEqual([{ key: second, width: 480 }]);
    expect(moveWorkbookColumn(contract, sized, second, "earlier")).toEqual(
      sized,
    );
  });

  it("rejects structural layout keys and out-of-range widths", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const initial = defaultWorkbookLayoutState(contract);
    expect(
      setWorkbookColumnWidth(contract, initial, "__selection__", 96),
    ).toEqual(initial);
    const fieldKey = initial.columnOrder[0] ?? "";
    expect(setWorkbookColumnWidth(contract, initial, fieldKey, 39)).toEqual(
      initial,
    );
    expect(setWorkbookColumnWidth(contract, initial, fieldKey, 4097)).toEqual(
      initial,
    );
  });
});
