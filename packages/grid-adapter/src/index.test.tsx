import { fireEvent, render, screen } from "@testing-library/react";
import { type ChangeEvent, useMemo, useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  assertGridRows,
  type GridColumn,
  type GridRow,
  GridTable,
  GridViewport,
  reconcileRecordRows,
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
      assertGridRows([{ recordId: "record-1" }, { recordId: " " }]),
    ).toThrow(/missing record_id/i);

    expect(() =>
      assertGridRows([{ recordId: "record-1" }, { recordId: "record-1" }]),
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
          getGroupRowTestId={(fieldKey, value) => `group-${fieldKey}-${value}`}
          groupBy="state"
          onToggleSort={onToggleSort}
          rows={rows}
          sort={[{ fieldKey: "label", direction: "asc" }]}
        />
      </GridViewport>,
    );

    const gridShell = await screen.findByTestId("grid-shell");
    const grid = gridShell.querySelector('[role="grid"]') as HTMLElement;
    expect(grid).toBeTruthy();
    expect(grid.style.gridTemplateColumns).toBe("224px 224px 176px");
    expect(
      gridShell.querySelector('[data-grid-record-id="record-1"]'),
    ).toBeTruthy();
    expect(
      gridShell.querySelector('[data-grid-record-id="record-2"]'),
    ).toBeTruthy();
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

  it("honors explicit data and actions column widths", async () => {
    const rows: readonly GridRow<HarnessRow>[] = [
      {
        key: "record-1",
        recordId: "record-1",
        data: {
          label: "Alpha",
          state: "open",
        },
      },
    ];
    const sizedColumns: readonly GridColumn<HarnessRow>[] = [
      {
        fieldKey: "label",
        label: "Label",
        renderCell: (row) => row.label,
        width: 320,
      },
      {
        fieldKey: "state",
        label: "State",
        renderCell: (row) => row.state,
      },
    ];

    render(
      <GridViewport testId="sized-grid-shell">
        <GridTable
          actionsColumn={{
            label: "Actions",
            minWidth: 64,
            renderCell: (row) => <span>{row.recordId}</span>,
            width: 96,
          }}
          columns={sizedColumns}
          rows={rows}
        />
      </GridViewport>,
    );

    const grid = (await screen.findByTestId("sized-grid-shell")).querySelector(
      '[role="grid"]',
    ) as HTMLElement;
    expect(grid.style.gridTemplateColumns).toBe("320px 224px 96px");
  });

  it("keeps editable cells mounted across repeated parent renders with an actions column", async () => {
    function EditableGridHarness() {
      const [label, setLabel] = useState("Alpha");
      const [renderMarker, setRenderMarker] = useState(0);
      const editableColumns = useMemo<readonly GridColumn<HarnessRow>[]>(
        () => [
          {
            fieldKey: "label",
            label: "Label",
            renderCell: (row) => (
              <input
                data-testid="editable-label"
                type="text"
                value={row.label}
                onChange={(event: ChangeEvent<HTMLInputElement>) => {
                  setLabel(event.target.value);
                }}
              />
            ),
          },
          {
            fieldKey: "state",
            label: "State",
            renderCell: (row) => row.state,
          },
        ],
        [],
      );
      const editableRows = useMemo<readonly GridRow<HarnessRow>[]>(
        () => [
          {
            key: "record-1",
            recordId: "record-1",
            data: {
              label,
              state: "open",
            },
            testId: "row-record-1",
          },
        ],
        [label],
      );
      const actionsColumn = useMemo(
        () => ({
          label: "Actions",
          renderCell: (row: GridRow<HarnessRow>) => (
            <span data-testid="row-action">{row.recordId}</span>
          ),
        }),
        [],
      );

      return (
        <GridViewport testId="editable-grid-shell">
          <button
            data-testid="force-rerender"
            type="button"
            onClick={() => {
              setRenderMarker((current) => current + 1);
            }}
          >
            Render {renderMarker}
          </button>
          <GridTable
            actionsColumn={actionsColumn}
            columns={editableColumns}
            rows={editableRows}
          />
        </GridViewport>
      );
    }

    render(<EditableGridHarness />);

    const input = await screen.findByTestId("editable-label");
    fireEvent.change(input, { target: { value: "Beta" } });
    fireEvent.click(screen.getByTestId("force-rerender"));
    fireEvent.change(screen.getByTestId("editable-label"), {
      target: { value: "Gamma" },
    });

    expect(
      (screen.getByTestId("editable-label") as HTMLInputElement).value,
    ).toBe("Gamma");
    expect(screen.getByTestId("row-action").textContent).toBe("record-1");
    expect(screen.getByTestId("editable-grid-shell")).toBeTruthy();
  });

  it("survives jsdom layout measurement when row pending state rerenders the RDG grid", async () => {
    function PendingGridHarness() {
      const [pending, setPending] = useState(false);
      const gridRows = useMemo<readonly GridRow<HarnessRow>[]>(
        () => [
          {
            key: "draft-1",
            recordId: null,
            data: {
              label: pending ? "Pending" : "Ready",
              state: pending ? "pending" : "draft",
            },
            variant: "draft",
          },
        ],
        [pending],
      );
      const actionsColumn = useMemo(
        () => ({
          label: "Actions",
          renderCell: (row: GridRow<HarnessRow>) => (
            <button
              data-testid="pending-row-action"
              disabled={row.data.state === "pending"}
              type="button"
              onMouseDown={(event) => {
                event.preventDefault();
                setPending(true);
              }}
            >
              {row.data.state}
            </button>
          ),
        }),
        [],
      );

      return (
        <GridViewport testId="pending-grid-shell">
          <GridTable
            actionsColumn={actionsColumn}
            columns={columns}
            rows={gridRows}
          />
        </GridViewport>
      );
    }

    render(<PendingGridHarness />);

    expect(() => {
      fireEvent.mouseDown(screen.getByTestId("pending-row-action"));
    }).not.toThrow();

    expect(
      (await screen.findByRole("button", {
        name: "pending",
      })) as HTMLButtonElement,
    ).toHaveProperty("disabled", true);
    expect(screen.getByTestId("pending-grid-shell")).toBeTruthy();
  });
});
