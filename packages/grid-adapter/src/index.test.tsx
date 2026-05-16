import { gridScrollportClassName } from "@cartulary/ui-contracts";
import { fireEvent, render, screen } from "@testing-library/react";
import { type ChangeEvent, useMemo, useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  assertGridRows,
  type GridColumn,
  type GridRow,
  type GridSortEntry,
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
    vi.spyOn(HTMLElement.prototype, "clientHeight", "get").mockReturnValue(720);
    vi.spyOn(HTMLElement.prototype, "clientWidth", "get").mockReturnValue(1280);
  });

  afterEach(() => {
    vi.restoreAllMocks();
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
    expect(grid.classList.contains(gridScrollportClassName())).toBe(true);
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

  it("keeps RDG row identity stable across reorder, sort, rerender, and editable cells", async () => {
    type EditableHarnessRow = HarnessRow & {
      readonly recordId: string;
    };

    function ReorderedGridHarness() {
      const [rows, setRows] = useState<readonly EditableHarnessRow[]>([
        { recordId: "record-1", label: "Alpha", state: "open" },
        { recordId: "record-2", label: "Zulu", state: "reviewed" },
      ]);
      const [renderMarker, setRenderMarker] = useState(0);
      const [sort, setSort] = useState<readonly GridSortEntry[]>([
        { fieldKey: "label", direction: "asc" },
      ]);
      const editableColumns = useMemo<
        readonly GridColumn<EditableHarnessRow>[]
      >(
        () => [
          {
            fieldKey: "label",
            headerTestId: "reorder-label-header",
            label: "Label",
            renderCell: (row) => (
              <input
                data-testid={`editable-label-${row.recordId}`}
                type="text"
                value={row.label}
                onChange={(event: ChangeEvent<HTMLInputElement>) => {
                  const nextLabel = event.target.value;
                  setRows((current) =>
                    current.map((candidate) =>
                      candidate.recordId === row.recordId
                        ? { ...candidate, label: nextLabel }
                        : candidate,
                    ),
                  );
                }}
              />
            ),
            sortableFieldKey: "label",
          },
          {
            fieldKey: "state",
            label: "State",
            renderCell: (row) => row.state,
          },
        ],
        [],
      );
      const gridRows = useMemo<readonly GridRow<EditableHarnessRow>[]>(
        () =>
          rows.map((row) => ({
            key: row.recordId,
            recordId: row.recordId,
            data: row,
            testId: `rdg-row-${row.recordId}`,
          })),
        [rows],
      );

      return (
        <GridViewport testId="reordered-grid-shell">
          <button
            data-testid="reverse-rows"
            type="button"
            onClick={() => {
              setRows((current) => [...current].reverse());
            }}
          >
            Reverse
          </button>
          <button
            data-testid="rerender-reordered-grid"
            type="button"
            onClick={() => {
              setRenderMarker((current) => current + 1);
            }}
          >
            Render {renderMarker}
          </button>
          <GridTable
            columns={editableColumns}
            onToggleSort={(fieldKey) => {
              setSort([{ fieldKey, direction: "desc" }]);
              setRows((current) =>
                [...current].sort((left, right) =>
                  right.label.localeCompare(left.label),
                ),
              );
            }}
            rows={gridRows}
            sort={sort}
          />
        </GridViewport>
      );
    }

    render(<ReorderedGridHarness />);

    fireEvent.click(await screen.findByTestId("reverse-rows"));
    fireEvent.click(screen.getByTestId("reorder-label-header"));
    fireEvent.click(screen.getByTestId("rerender-reordered-grid"));
    fireEvent.change(screen.getByTestId("editable-label-record-1"), {
      target: { value: "Alpha edited" },
    });

    const shell = screen.getByTestId("reordered-grid-shell");
    const savedRows = Array.from(
      shell.querySelectorAll('[role="row"][data-grid-record-id]'),
    );
    expect(
      savedRows.map((row) => row.getAttribute("data-grid-record-id")),
    ).toEqual(["record-2", "record-1"]);
    expect(
      (screen.getByTestId("editable-label-record-1") as HTMLInputElement).value,
    ).toBe("Alpha edited");
  });

  it("keeps grouped editable draft cells stable across repeated local edits", async () => {
    type EditableHarnessRow = HarnessRow & {
      readonly recordId: string;
    };

    function DraftInputCell({ row }: { readonly row: EditableHarnessRow }) {
      const [draftValue, setDraftValue] = useState(row.label);
      return (
        <input
          data-testid={`grouped-editable-label-${row.recordId}`}
          type="text"
          value={draftValue}
          onChange={(event: ChangeEvent<HTMLInputElement>) => {
            setDraftValue(event.target.value);
          }}
        />
      );
    }

    function GroupedEditableGridHarness() {
      const [renderMarker, setRenderMarker] = useState(0);
      const rows = useMemo<readonly EditableHarnessRow[]>(
        () => [
          { recordId: "record-1", label: "Alpha", state: "open" },
          { recordId: "record-2", label: "Beta", state: "reviewed" },
          { recordId: "record-3", label: "Gamma", state: "open" },
        ],
        [],
      );
      const editableColumns = useMemo<
        readonly GridColumn<EditableHarnessRow>[]
      >(
        () => [
          {
            fieldKey: "label",
            label: "Label",
            renderCell: (row) => <DraftInputCell row={row} />,
          },
          {
            fieldKey: "state",
            label: "State",
            renderCell: (row) => row.state,
          },
        ],
        [],
      );
      const gridRows = useMemo<readonly GridRow<EditableHarnessRow>[]>(
        () =>
          rows.map((row) => ({
            key: row.recordId,
            recordId: row.recordId,
            data: row,
            testId: `grouped-editable-row-${row.recordId}`,
          })),
        [rows],
      );

      return (
        <GridViewport testId="grouped-editable-grid-shell">
          <button
            data-testid="rerender-grouped-editable-grid"
            type="button"
            onClick={() => {
              setRenderMarker((current) => current + 1);
            }}
          >
            Render {renderMarker}
          </button>
          <GridTable
            columns={editableColumns}
            getGroupLabel={(row, fieldKey) =>
              fieldKey === "state" ? row.state : null
            }
            getGroupRowTestId={(fieldKey, value) =>
              `grouped-editable-group-${fieldKey}-${value}`
            }
            groupBy="state"
            rows={gridRows}
          />
        </GridViewport>
      );
    }

    render(<GroupedEditableGridHarness />);

    const input = await screen.findByTestId("grouped-editable-label-record-3");
    for (const value of ["Gamma draft 1", "Gamma draft 2", "Gamma final"]) {
      fireEvent.change(input, { target: { value } });
      expect((input as HTMLInputElement).value).toBe(value);
    }
    fireEvent.click(screen.getByTestId("rerender-grouped-editable-grid"));

    expect(
      (
        screen.getByTestId(
          "grouped-editable-label-record-3",
        ) as HTMLInputElement
      ).value,
    ).toBe("Gamma final");
    const shell = screen.getByTestId("grouped-editable-grid-shell");
    expect(
      shell.querySelectorAll(
        '[data-testid="grouped-editable-group-state-open"]',
      ),
    ).toHaveLength(1);
    expect(screen.getByTestId("grouped-editable-row-record-3")).toBeTruthy();
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
