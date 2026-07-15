import { gridScrollportClassName } from "@cartulary/ui-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { type ChangeEvent, createRef, useMemo, useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  assertGridRows,
  type GridColumn,
  type GridHandle,
  type GridRecordRow,
  type GridSortEntry,
  GridViewport,
  WorkbookDataGrid,
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
    renderCell: ({ row }) => row.label,
    sortableFieldKey: "label",
  },
  {
    fieldKey: "state",
    headerTestId: "state-header",
    label: "State",
    renderCell: ({ row }) => row.state,
    sortableFieldKey: "state",
    sortDisabled: true,
    sortDisabledReason: "State sorting disabled in this harness",
  },
];

describe("grid-adapter", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "IntersectionObserver",
      class {
        disconnect() {}
        observe() {}
        unobserve() {}
      },
    );
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

  it("translates live cell events and the restricted handle through semantic coordinates", async () => {
    const onActiveCellChange = vi.fn();
    const onCopyCell = vi.fn();
    const onEditCell = vi.fn();
    const onPasteCell = vi.fn();
    const handle = createRef<GridHandle>();
    render(
      <WorkbookDataGrid
        ref={handle}
        viewSchemaId="test.view"
        columns={[
          {
            contractWritable: true,
            fieldKey: "label",
            label: "Label",
            renderCell: ({ row }) => (
              <span data-testid="semantic-live-cell">{row.label}</span>
            ),
            renderEditCell: ({ row }) => (
              <input aria-label="Semantic editor" defaultValue={row.label} />
            ),
          },
        ]}
        onActiveCellChange={onActiveCellChange}
        onCopyCell={onCopyCell}
        onEditCell={onEditCell}
        onPasteCell={onPasteCell}
        recordRows={[
          {
            kind: "record",
            recordId: "record-1",
            rowVersion: 7,
            data: { label: "Alpha", state: "open" },
          },
        ]}
      />,
    );
    const cell = (await screen.findByTestId("semantic-live-cell")).closest(
      '[role="gridcell"]',
    );
    if (!(cell instanceof HTMLElement))
      throw new Error("Expected live RDG cell");
    fireEvent.mouseDown(cell);
    fireEvent.click(cell);
    await waitFor(() => expect(onActiveCellChange).toHaveBeenCalled());
    const liveGrid = screen.getByRole("grid");
    fireEvent.copy(liveGrid);
    fireEvent.paste(liveGrid);
    fireEvent.doubleClick(cell);
    const anchor = {
      fieldKey: "label",
      recordId: "record-1",
      viewSchemaId: "test.view",
    };
    const target = { ...anchor, baseRowVersion: 7 };
    expect(onActiveCellChange).toHaveBeenCalledWith(anchor);
    expect(onCopyCell).toHaveBeenCalledWith({ anchor });
    expect(onPasteCell).toHaveBeenCalledWith({ target });
    expect(onEditCell).toHaveBeenCalledWith({ target });
    expect(handle.current?.getScrollElement()).toBeTruthy();
    expect(handle.current?.scrollToAnchor(anchor)).toBe(true);
    expect(handle.current?.focusAnchor(anchor)).toBe(true);
    expect(
      handle.current?.focusAnchor({ ...anchor, viewSchemaId: "wrong.view" }),
    ).toBe(false);
  });

  it("renders stable semantic row attributes, sort translation, and presentation-only group rows", async () => {
    const onToggleSort = vi.fn();
    const rows: readonly GridRecordRow<HarnessRow>[] = [
      {
        kind: "record",
        recordId: "record-1",
        rowVersion: 1,
        data: {
          label: "Alpha",
          state: "open",
        },
        testId: "row-record-1",
      },
      {
        kind: "record",
        recordId: "record-2",
        rowVersion: 1,
        data: {
          label: "Beta",
          state: "reviewed",
        },
        testId: "row-record-2",
      },
    ];

    render(
      <GridViewport testId="grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          actionsColumn={{
            label: "Actions",
            renderCell: ({ recordId }) => <span>{recordId}</span>,
          }}
          columns={columns}
          grouping={{
            fieldKey: "state",
            formatLabel: (value) => (value === null ? null : String(value)),
            getTestId: (fieldKey, _value, label) =>
              label === null ? undefined : `group-${fieldKey}-${label}`,
            getValue: (row) => row.state,
          }}
          onToggleSort={onToggleSort}
          recordRows={rows}
          sort={[{ fieldKey: "label", direction: "asc" }]}
        />
      </GridViewport>,
    );

    const gridShell = await screen.findByTestId("grid-shell");
    const grid = gridShell.querySelector('[role="treegrid"]') as HTMLElement;
    expect(grid).toBeTruthy();
    expect(grid.classList.contains(gridScrollportClassName())).toBe(true);
    expect(
      gridShell.querySelector('[data-grid-record-id="record-1"]'),
    ).toBeTruthy();
    expect(
      gridShell.querySelector('[data-grid-record-id="record-2"]'),
    ).toBeTruthy();
    expect(screen.getByTestId("group-state-open")).toBeTruthy();
    expect(screen.getByTestId("group-state-reviewed")).toBeTruthy();
    const openGroupToggle = screen.getByTestId("group-state-open");
    const openGroupRow = openGroupToggle.closest('[role="row"]');
    expect(openGroupRow).toBeTruthy();
    if (openGroupRow === null) {
      throw new Error("Expected open group toggle to have row ancestor");
    }
    expect(openGroupRow.getAttribute("data-grid-record-id")).toBeNull();
    expect(
      openGroupRow.querySelectorAll("input, textarea, select"),
    ).toHaveLength(0);
    expect(openGroupRow.querySelectorAll("button")).toHaveLength(1);
    expect(openGroupToggle.getAttribute("aria-expanded")).toBe("true");
    fireEvent.click(openGroupToggle);
    expect(openGroupToggle.getAttribute("aria-expanded")).toBe("false");
    expect(screen.queryByTestId("row-record-1")).toBeNull();
    fireEvent.click(openGroupToggle);
    expect(openGroupToggle.getAttribute("aria-expanded")).toBe("true");
    expect(screen.getByTestId("row-record-1")).toBeTruthy();

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

  it("preserves typed bucket order, scoped expansion, and draft exclusion", async () => {
    type GroupRow = {
      readonly group: boolean | number | string | null;
      readonly label: string;
    };
    const typedColumns: readonly GridColumn<GroupRow>[] = [
      {
        fieldKey: "label",
        label: "Label",
        renderCell: ({ row }) => row.label,
        renderDraftCell: ({ row }) => (
          <input aria-label="Typed draft" defaultValue={row.label} />
        ),
      },
    ];
    const rows: readonly GridRecordRow<GroupRow>[] = [
      {
        kind: "record",
        recordId: "string",
        rowVersion: 1,
        data: { group: "1", label: "String" },
      },
      {
        kind: "record",
        recordId: "number",
        rowVersion: 1,
        data: { group: 1, label: "Number" },
      },
      {
        kind: "record",
        recordId: "boolean",
        rowVersion: 1,
        data: { group: true, label: "Boolean" },
      },
      {
        kind: "record",
        recordId: "null",
        rowVersion: 1,
        data: { group: null, label: "Null" },
      },
    ];
    const grid = (fieldKey: string, recordRows = rows) => (
      <WorkbookDataGrid
        viewSchemaId="typed.view"
        columns={typedColumns}
        draftRow={{
          kind: "draft",
          data: { group: "draft-only", label: "Draft" },
        }}
        grouping={{
          fieldKey,
          formatLabel: (value) => (value === null ? null : String(value)),
          getTestId: (key, value) =>
            `typed-${key}-${value === null ? "null" : typeof value}-${String(value)}`,
          getValue: (row) => row.group,
        }}
        recordRows={recordRows}
      />
    );
    const { rerender } = render(grid("primary"));
    expect(
      screen
        .getAllByTestId(/^typed-primary-/)
        .map((group) => group.getAttribute("data-cartulary-grid-group-id")),
    ).toEqual(["s:1", "d:1", "b:true", "n:null"]);
    expect(screen.getByLabelText("Typed draft")).toBeTruthy();
    expect(
      document.querySelector('[data-cartulary-grid-group-id="s:draft-only"]'),
    ).toBeNull();

    fireEvent.click(screen.getByTestId("typed-primary-string-1"));
    rerender(grid("secondary"));
    expect(
      screen
        .getByTestId("typed-secondary-string-1")
        .getAttribute("aria-expanded"),
    ).toBe("true");
    rerender(grid("primary"));
    expect(
      screen
        .getByTestId("typed-primary-string-1")
        .getAttribute("aria-expanded"),
    ).toBe("false");
    rerender(grid("primary", rows.slice(1)));
    await waitFor(() =>
      expect(screen.queryByTestId("typed-primary-string-1")).toBeNull(),
    );
    rerender(grid("primary"));
    await waitFor(() =>
      expect(
        screen
          .getByTestId("typed-primary-string-1")
          .getAttribute("aria-expanded"),
      ).toBe("true"),
    );
  });

  it("can fill available inline space on demand", async () => {
    const rows: readonly GridRecordRow<HarnessRow>[] = [
      {
        kind: "record",
        recordId: "record-1",
        rowVersion: 1,
        data: {
          label: "Alpha",
          state: "open",
        },
      },
    ];

    const { rerender } = render(
      <GridViewport testId="fill-grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          actionsColumn={{
            label: "Actions",
            renderCell: ({ recordId }) => <span>{recordId}</span>,
          }}
          columns={columns}
          recordRows={rows}
        />
      </GridViewport>,
    );

    const gridShell = await screen.findByTestId("fill-grid-shell");
    const grid = gridShell.querySelector('[role="grid"]') as HTMLElement;
    rerender(
      <GridViewport testId="fill-grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          actionsColumn={{
            label: "Actions",
            renderCell: ({ recordId }) => <span>{recordId}</span>,
          }}
          columns={columns}
          fillViewportInline
          recordRows={rows}
        />
      </GridViewport>,
    );

    expect(["0", "0px"]).toContain(grid.style.minWidth);
    expect(grid.style.width).toBe("100%");
  });

  it("renders the production DataGrid with bounded fixed-height row output", async () => {
    const rowCount = 500;
    const rows: readonly GridRecordRow<HarnessRow>[] = Array.from(
      { length: rowCount },
      (_, index) => ({
        kind: "record" as const,
        recordId: `virtual-record-${index}`,
        rowVersion: 1,
        data: { label: `Record ${index}`, state: "open" },
      }),
    );

    render(
      <GridViewport testId="virtualized-grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          columns={columns}
          recordRows={rows}
        />
      </GridViewport>,
    );

    const shell = screen.getByTestId("virtualized-grid-shell");
    const grid = shell.querySelector('[role="grid"]');
    const mountedRecordRows = shell.querySelectorAll(
      '[role="row"][data-grid-record-id]',
    );
    expect(grid?.getAttribute("aria-rowcount")).toBe(String(rowCount + 1));
    expect(mountedRecordRows.length).toBeGreaterThan(0);
    expect(mountedRecordRows.length).toBeLessThan(rowCount);
  });

  it("keeps standalone block sizing by default and supports shell-owned fill block sizing", async () => {
    const rows: readonly GridRecordRow<HarnessRow>[] = [
      {
        kind: "record",
        recordId: "record-1",
        rowVersion: 1,
        data: {
          label: "Alpha",
          state: "open",
        },
      },
    ];

    const { rerender } = render(
      <GridViewport testId="block-sizing-grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          columns={columns}
          recordRows={rows}
        />
      </GridViewport>,
    );

    const gridShell = await screen.findByTestId("block-sizing-grid-shell");
    const grid = gridShell.querySelector('[role="grid"]') as HTMLElement;
    expect(gridShell.style.blockSize).toBe("min(70vh, 46rem)");
    expect(gridShell.style.minBlockSize).toBe("18rem");
    expect(grid.style.blockSize).toBe("100%");
    expect(grid.style.overflow).toBe("auto");

    rerender(
      <GridViewport blockSizing="fill" testId="block-sizing-grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          columns={columns}
          recordRows={rows}
        />
      </GridViewport>,
    );

    expect(gridShell.style.blockSize).toBe("100%");
    expect(gridShell.style.boxSizing).toBe("border-box");
    expect(["0", "0px"]).toContain(gridShell.style.minBlockSize);
    expect(grid.style.blockSize).toBe("100%");
    expect(grid.style.overflow).toBe("auto");
  });

  it("renders data and actions columns without freezing implementation geometry", async () => {
    const rows: readonly GridRecordRow<HarnessRow>[] = [
      {
        kind: "record",
        recordId: "record-1",
        rowVersion: 1,
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
        renderCell: ({ row }) => row.label,
        width: 320,
      },
      {
        fieldKey: "state",
        label: "State",
        renderCell: ({ row }) => row.state,
      },
    ];

    render(
      <GridViewport testId="sized-grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          actionsColumn={{
            label: "Actions",
            minWidth: 64,
            renderCell: ({ recordId }) => <span>{recordId}</span>,
            width: 96,
          }}
          columns={sizedColumns}
          recordRows={rows}
        />
      </GridViewport>,
    );

    const grid = (await screen.findByTestId("sized-grid-shell")).querySelector(
      '[role="grid"]',
    );
    expect(grid).toBeTruthy();
    expect(screen.getAllByText("record-1").length).toBeGreaterThan(0);
  });

  it("selects density token variables explicitly for every supported mode", async () => {
    const rows: readonly GridRecordRow<HarnessRow>[] = [
      {
        kind: "record",
        recordId: "record-1",
        rowVersion: 1,
        data: {
          label: "Alpha",
          state: "open",
        },
      },
    ];

    const { rerender } = render(
      <GridViewport testId="density-grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          columns={columns}
          recordRows={rows}
        />
      </GridViewport>,
    );

    const grid = (
      await screen.findByTestId("density-grid-shell")
    ).querySelector('[role="grid"]') as HTMLElement;
    expect(grid.style.getPropertyValue("--cartulary-grid-density")).toBe(
      "default",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-row-height")).toBe(
      "var(--ct-density-default-rowHeight)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-cell-padding")).toBe(
      "var(--ct-density-default-cellPadding)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-font-size")).toBe(
      "var(--ct-density-default-fontSize)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-line-height")).toBe(
      "var(--ct-density-default-lineHeight)",
    );

    rerender(
      <GridViewport testId="density-grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          columns={columns}
          density="compact"
          recordRows={rows}
        />
      </GridViewport>,
    );

    expect(grid.style.getPropertyValue("--cartulary-grid-density")).toBe(
      "compact",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-row-height")).toBe(
      "var(--ct-density-compact-rowHeight)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-cell-padding")).toBe(
      "var(--ct-density-compact-cellPadding)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-font-size")).toBe(
      "var(--ct-density-compact-fontSize)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-line-height")).toBe(
      "var(--ct-density-compact-lineHeight)",
    );

    rerender(
      <GridViewport testId="density-grid-shell">
        <WorkbookDataGrid
          viewSchemaId="test.view"
          columns={columns}
          density="comfortable"
          recordRows={rows}
        />
      </GridViewport>,
    );

    expect(grid.style.getPropertyValue("--cartulary-grid-density")).toBe(
      "comfortable",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-row-height")).toBe(
      "var(--ct-density-comfortable-rowHeight)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-cell-padding")).toBe(
      "var(--ct-density-comfortable-cellPadding)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-font-size")).toBe(
      "var(--ct-density-comfortable-fontSize)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-line-height")).toBe(
      "var(--ct-density-comfortable-lineHeight)",
    );
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
            renderCell: ({ row }) => (
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
            renderCell: ({ row }) => row.state,
          },
        ],
        [],
      );
      const editableRows = useMemo<readonly GridRecordRow<HarnessRow>[]>(
        () => [
          {
            kind: "record",
            recordId: "record-1",
            rowVersion: 1,
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
          renderCell: (row: GridRecordRow<HarnessRow>) => (
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
          <WorkbookDataGrid
            viewSchemaId="test.view"
            actionsColumn={actionsColumn}
            columns={editableColumns}
            recordRows={editableRows}
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
            renderCell: ({ row }) => (
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
            renderCell: ({ row }) => row.state,
          },
        ],
        [],
      );
      const gridRows = useMemo<readonly GridRecordRow<EditableHarnessRow>[]>(
        () =>
          rows.map((row) => ({
            kind: "record",
            recordId: row.recordId,
            rowVersion: 1,
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
          <WorkbookDataGrid
            viewSchemaId="test.view"
            columns={editableColumns}
            onToggleSort={(fieldKey) => {
              setSort([{ fieldKey, direction: "desc" }]);
              setRows((current) =>
                [...current].sort((left, right) =>
                  right.label.localeCompare(left.label),
                ),
              );
            }}
            recordRows={gridRows}
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
            renderCell: ({ row }) => <DraftInputCell row={row} />,
          },
          {
            fieldKey: "state",
            label: "State",
            renderCell: ({ row }) => row.state,
          },
        ],
        [],
      );
      const gridRows = useMemo<readonly GridRecordRow<EditableHarnessRow>[]>(
        () =>
          rows.map((row) => ({
            kind: "record",
            recordId: row.recordId,
            rowVersion: 1,
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
          <WorkbookDataGrid
            viewSchemaId="test.view"
            columns={editableColumns}
            grouping={{
              fieldKey: "state",
              formatLabel: (value) => (value === null ? null : String(value)),
              getTestId: (fieldKey, _value, label) =>
                label === null
                  ? undefined
                  : `grouped-editable-group-${fieldKey}-${label}`,
              getValue: (row) => row.state,
            }}
            recordRows={gridRows}
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
      const draftRow = useMemo(
        () => ({
          kind: "draft" as const,
          data: {
            label: pending ? "Pending" : "Ready",
            state: pending ? "pending" : "draft",
          },
        }),
        [pending],
      );
      const actionsColumn = useMemo(
        () => ({
          label: "Actions",
          renderCell: (row: GridRecordRow<HarnessRow>) => row.recordId,
          renderDraftCell: (row: { readonly data: HarnessRow }) => (
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
          <WorkbookDataGrid
            viewSchemaId="test.view"
            actionsColumn={actionsColumn}
            columns={columns}
            draftRow={draftRow}
            recordRows={[]}
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
