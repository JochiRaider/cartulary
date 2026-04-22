import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  assertGridRows,
  GridTable,
  GridViewport,
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
    headerTestId: "label-header",
    label: "Label",
    renderCell: (row) => row.label,
    sortableFieldKey: "label",
  },
  {
    fieldKey: "state",
    headerTestId: "state-header",
    label: "State",
    renderCell: (row) => row.state,
    sortableFieldKey: "state",
    sortDisabled: true,
    sortDisabledReason: "State sorting disabled in this harness",
  },
];

describe("grid-adapter", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "ResizeObserver",
      class {
        disconnect() {}
        observe() {}
        unobserve() {}
      },
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

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

  it("renders an RDG-backed grid with stable row attributes, sort translation, and presentation-only group rows", async () => {
    const onToggleSort = vi.fn();
    const rows: readonly GridRow<HarnessRow>[] = [
      {
        key: "record-1",
        recordId: "record-1",
        data: {
          label: "Alpha",
          state: "open",
        },
        testId: "row-record-1",
      },
      {
        key: "record-2",
        recordId: "record-2",
        data: {
          label: "Beta",
          state: "reviewed",
        },
        testId: "row-record-2",
      },
    ];

    render(
      <GridViewport testId="grid-shell">
        <GridTable
          actionsColumn={{
            label: "Actions",
            renderCell: (row) => <span>{row.recordId}</span>,
          }}
          columns={columns}
          getGroupLabel={(row, fieldKey) =>
            fieldKey === "state" ? row.state : null
          }
          getGroupRowTestId={(fieldKey, value) =>
            `group-${fieldKey}-${value}`
          }
          groupBy="state"
          onToggleSort={onToggleSort}
          rows={rows}
          sort={[{ fieldKey: "label", direction: "asc" }]}
        />
      </GridViewport>,
    );

    const gridShell = await screen.findByTestId("grid-shell");
    const grid = gridShell.querySelector('[role="grid"]');
    expect(grid).toBeTruthy();
    expect(gridShell.querySelector('[data-grid-record-id="record-1"]')).toBeTruthy();
    expect(gridShell.querySelector('[data-grid-record-id="record-2"]')).toBeTruthy();
    expect(screen.getByTestId("group-state-open")).toBeTruthy();
    expect(screen.getByTestId("group-state-reviewed")).toBeTruthy();

    const labelHeader = screen.getByTestId("label-header");
    expect(labelHeader.getAttribute("data-grid-field-key")).toBe("label");
    fireEvent.click(labelHeader);
    expect(onToggleSort).toHaveBeenCalledWith("label");

    const stateHeader = screen.getByTestId("state-header");
    expect(stateHeader.getAttribute("title")).toBe(
      "State sorting disabled in this harness",
    );
    fireEvent.click(stateHeader);
    expect(onToggleSort).toHaveBeenCalledTimes(1);
  });
});
