import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import {
  assertGridRows,
  GridViewport,
  GridTable,
  reconcileRecordRows,
  type GridColumn,
  type GridRow,
} from "./index";

type HarnessRow = {
  readonly label: string;
  readonly state: string;
};

const columns: readonly GridColumn<HarnessRow>[] = [
  {
    fieldKey: "label",
    label: "Label",
    renderCell: (row) => row.label,
    sortableFieldKey: "label",
  },
  {
    fieldKey: "state",
    label: "State",
    renderCell: (row) => row.state,
    sortableFieldKey: "state",
    sortDisabled: true,
    sortDisabledReason: "State sorting disabled in this harness",
  },
];

describe("grid-adapter", () => {
  it("rejects missing and duplicate saved record identities", () => {
    expect(() =>
      assertGridRows([
        { recordId: "record-1" },
        { recordId: " " },
      ]),
    ).toThrow(/missing record_id/i);

    expect(() =>
      assertGridRows([
        { recordId: "record-1" },
        { recordId: "record-1" },
      ]),
    ).toThrow(/duplicate record_id/i);
  });

  it("preserves unchanged saved-row references by record_id", () => {
    const stable = {
      recordId: "record-1",
      label: "Alpha",
      state: "open",
    };
    const previous = [stable];
    const next = [
      {
        recordId: "record-1",
        label: "Alpha",
        state: "open",
      },
      {
        recordId: "record-2",
        label: "Beta",
        state: "closed",
      },
    ];

    const reconciled = reconcileRecordRows(previous, next);
    expect(reconciled[0]).toBe(stable);
    expect(reconciled[1]).not.toBe(next[0]);
  });

  it("renders sortable headers, disabled capability blocking, and grouped rows", async () => {
    const onToggleSort = vi.fn();
    const rows: readonly GridRow<HarnessRow>[] = [
      {
        key: "record-1",
        recordId: "record-1",
        data: {
          label: "Alpha",
          state: "open",
        },
      },
      {
        key: "record-2",
        recordId: "record-2",
        data: {
          label: "Beta",
          state: "reviewed",
        },
      },
    ];

    render(
      <GridViewport>
        <GridTable
          columns={columns}
          getGroupLabel={(row, fieldKey) =>
            fieldKey === "state" ? row.state : null
          }
          groupBy="state"
          onToggleSort={onToggleSort}
          rows={rows}
          sort={[{ fieldKey: "label", direction: "asc" }]}
        />
      </GridViewport>,
    );

    await screen.findAllByText("open");
    expect(screen.getByRole("button", { name: "Sort Label" })).toBeTruthy();
    expect(screen.getByText("State").getAttribute("title")).toBe(
      "State sorting disabled in this harness",
    );

    screen.getByRole("button", { name: "Sort Label" }).click();
    expect(onToggleSort).toHaveBeenCalledWith("label");
    expect(screen.getAllByText("reviewed")).toHaveLength(2);
  });
});
