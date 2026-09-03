import {
  gridGroupRowSelector,
  gridScrollportClassName,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { type ChangeEvent, createRef, useMemo, useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  assertGridRows,
  type GridEditorRenderContext,
  type SemanticDataGridProps,
} from "./core";
import { SemanticDataGrid as SemanticDataGridDomUnit } from "./domUnitBinding";
import {
  type GridColumn,
  type GridDataRow,
  type GridDensity,
  type GridHandle,
  type GridSortEntry,
  GridViewport,
  SemanticDataGrid,
} from "./index";
import { SemanticDataGrid as SemanticDataGridTestSupport } from "./test-support";

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

const testSurface = { kind: "view_schema", viewSchemaId: "test.view" } as const;

const semanticContractBindings = [
  { Grid: SemanticDataGridDomUnit, name: "production DOM-unit" },
  { Grid: SemanticDataGridTestSupport, name: "test-support" },
] as const;

function decodeTestClipboard(rawText: string) {
  return { kind: "scalar" as const, rawText, value: rawText };
}

function gridAnchor(recordId: string, fieldKey: string) {
  return {
    fieldKey,
    rowIdentity: { kind: "core_record" as const, recordId },
    surface: testSurface,
  };
}

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
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("rejects missing and duplicate saved record identities", () => {
    expect(() =>
      assertGridRows([
        {
          data: { label: "Alpha", state: "open" },
          kind: "data",
          rowIdentity: { kind: "core_record", recordId: " " },
        },
      ]),
    ).toThrow(/invalid semantic identity/i);

    expect(() =>
      assertGridRows([
        {
          data: { label: "Alpha", state: "open" },
          kind: "data",
          rowIdentity: { kind: "core_record", recordId: "record-1" },
        },
        {
          data: { label: "Beta", state: "open" },
          kind: "data",
          rowIdentity: { kind: "core_record", recordId: "record-1" },
        },
      ]),
    ).toThrow(/duplicate semantic row identity/i);
  });

  it("runs identity, state, grouping, sorting, and selection through both semantic bindings", async () => {
    for (const { Grid, name } of semanticContractBindings) {
      const onSelectedRecordIdsChange = vi.fn();
      const onSortChange = vi.fn();
      const contractColumns: readonly GridColumn<HarnessRow>[] = [
        {
          fieldKey: "label",
          headerTestId: `contract-label-header-${name}`,
          label: "Label",
          renderCell: ({ row }) => row.label,
          sortableFieldKey: "label",
        },
        {
          fieldKey: "state",
          label: "State",
          renderCell: ({ row }) => row.state,
        },
      ];
      const contractRows: readonly GridDataRow<HarnessRow>[] = [
        {
          data: { label: "Alpha", state: "open" },
          kind: "data",
          mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
          rowIdentity: { kind: "core_record", recordId: "contract-1" },
          testId: `contract-row-${name}`,
        },
      ];
      const view = render(
        <Grid
          activeRowIdentity={{
            kind: "core_record",
            recordId: "contract-1",
          }}
          columns={contractColumns}
          coreRecordBulkSelection={{
            onSelectedRecordIdsChange,
            selectedRecordIds: new Set(),
          }}
          dataRows={contractRows}
          dataState={{ kind: "stale_error", message: "Refresh failed." }}
          getCellState={({ anchor }) =>
            anchor.fieldKey === "label" ? { conflicted: true } : {}
          }
          grouping={{
            fieldKey: "state",
            formatLabel: (value) => (value === null ? null : String(value)),
            getTestId: () => `contract-group-${name}`,
            getValue: (row) => row.state,
          }}
          onSortChange={onSortChange}
          surface={testSurface}
        />,
      );

      expect(await screen.findByTestId(`contract-group-${name}`)).toBeTruthy();
      const row = screen.getByTestId(`contract-row-${name}`);
      expect(row.getAttribute("data-grid-row-identity-kind")).toBe(
        "core_record",
      );
      expect(row.getAttribute("data-grid-record-id")).toBe("contract-1");
      expect(row.getAttribute("aria-current")).toBe("true");
      const labelCell = row.querySelector<HTMLElement>(
        '[role="gridcell"][data-grid-field-key="label"], [data-grid-field-key="label"]',
      );
      expect(
        labelCell?.closest<HTMLElement>('[role="gridcell"]')?.dataset
          .gridPrimaryState ?? labelCell?.dataset.gridPrimaryState,
      ).toBe("conflicted");
      fireEvent.click(screen.getByTestId(`contract-label-header-${name}`));
      expect(onSortChange).toHaveBeenCalledWith([
        { direction: "asc", fieldKey: "label" },
      ]);
      fireEvent.click(
        screen.getByRole("checkbox", { name: "Select record contract-1" }),
      );
      expect(onSelectedRecordIdsChange).toHaveBeenCalledWith(
        new Set(["contract-1"]),
      );
      expect(
        screen.getByText(
          "Refresh failed. Previously loaded rows may be stale.",
        ),
      ).toBeTruthy();

      view.unmount();
      cleanup();
    }
  });

  it("runs fail-closed capability admission through both semantic bindings", () => {
    for (const { Grid, name } of semanticContractBindings) {
      const onFillCells = vi.fn();
      const onPaste = vi.fn();
      const unsafeExtensionProps = {
        actionsColumn: {
          label: "Unsafe actions",
          renderCell: () => <button type="button">Unsafe action</button>,
        },
        allowPasteCreateRows: true,
        clipboardPaste: { decode: decodeTestClipboard, onPaste },
        columns: [
          {
            contractWritable: true,
            editor: {
              commit: async () => ({ kind: "accepted" as const }),
              initialDraftValue: (row: HarnessRow) => row.label,
              renderEditor: ({
                focusTargetRef,
              }: GridEditorRenderContext<HarnessRow>) => (
                <input
                  aria-label={`Unsafe editor ${name}`}
                  ref={focusTargetRef}
                />
              ),
            },
            fieldKey: "label",
            label: "Label",
            renderCell: ({ row }: { readonly row: HarnessRow }) => row.label,
          },
        ],
        coreRecordBulkSelection: {
          onSelectedRecordIdsChange: vi.fn(),
          selectedRecordIds: new Set<string>(),
        },
        dataRows: [
          {
            data: { label: "Extension row", state: "accepted" },
            kind: "data" as const,
            rowIdentity: {
              extensionProfileId: "network_flow_activity",
              kind: "extension_resource" as const,
              resourceId: "contract-extension-1",
              resourceKind: "accepted_flow_row",
            },
          },
        ],
        draftRow: {
          data: { label: "Unsafe draft", state: "draft" },
          kind: "draft" as const,
        },
        interactionMode: { kind: "editable" as const },
        onFillCells,
        surface: {
          extensionProfileId: "network_flow_activity",
          gridSchemaId: "network_flow.accepted_rows.v1",
          kind: "extension_grid" as const,
          workspaceKey: "contract-workspace",
        },
      } as unknown as SemanticDataGridProps<HarnessRow>;
      expect(() => render(<Grid {...unsafeExtensionProps} />)).toThrow(
        /cannot enable Core mutation/i,
      );
      cleanup();
    }
  });

  it("runs active-cell, navigation, range, copy, paste, fill, and handle behavior through both semantic bindings", async () => {
    for (const { Grid, name } of semanticContractBindings) {
      const handle = createRef<GridHandle>();
      const onActiveCellChange = vi.fn();
      const onCellRangeChange = vi.fn();
      const onCopyCell = vi.fn();
      const onFillCells = vi.fn();
      const onPaste = vi.fn();
      const interactiveColumns: readonly GridColumn<HarnessRow>[] = [
        {
          contractWritable: true,
          editor: {
            commit: async () => ({ kind: "accepted" }),
            initialDraftValue: (row) => row.label,
            renderEditor: ({ focusTargetRef }) => (
              <input
                aria-label={`Contract label editor ${name}`}
                ref={focusTargetRef}
              />
            ),
          },
          fieldKey: "label",
          getClipboardValue: (row) => row.label,
          label: "Label",
          renderCell: ({ row }) => row.label,
        },
        {
          fieldKey: "state",
          getClipboardValue: (row) => row.state,
          label: "State",
          renderCell: ({ row }) => row.state,
        },
      ];
      const interactiveRows: readonly GridDataRow<HarnessRow>[] = [
        {
          data: { label: "Alpha", state: "open" },
          kind: "data",
          mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
          rowIdentity: { kind: "core_record", recordId: "contract-1" },
        },
        {
          data: { label: "Beta", state: "reviewed" },
          kind: "data",
          mutationIdentity: { kind: "core_row_version", baseRowVersion: 2 },
          rowIdentity: { kind: "core_record", recordId: "contract-2" },
        },
      ];
      const view = render(
        <Grid
          ref={handle}
          clipboardPaste={{ decode: decodeTestClipboard, onPaste }}
          columns={interactiveColumns}
          dataRows={interactiveRows}
          onActiveCellChange={onActiveCellChange}
          onCellRangeChange={onCellRangeChange}
          onCopyCell={onCopyCell}
          onFillCells={onFillCells}
          surface={testSurface}
        />,
      );
      const alpha = gridAnchor("contract-1", "label");
      const alphaState = gridAnchor("contract-1", "state");
      const beta = gridAnchor("contract-2", "label");

      expect(handle.current?.focusAnchor(alpha)).toBe(true);
      expect(onActiveCellChange).toHaveBeenCalledTimes(1);
      expect(handle.current?.focusAnchor(alpha)).toBe(true);
      expect(onActiveCellChange).toHaveBeenCalledTimes(1);
      expect(handle.current?.moveFocus(alpha, { key: "ArrowRight" })).toEqual(
        alphaState,
      );
      expect(onActiveCellChange).toHaveBeenCalledTimes(2);
      expect(handle.current?.focusAnchor(alpha)).toBe(true);

      const alphaCell = screen
        .getByText("Alpha")
        .closest<HTMLElement>('[role="gridcell"]');
      if (alphaCell === null) throw new Error(`Missing ${name} Alpha cell`);
      fireEvent.mouseDown(alphaCell);
      fireEvent.keyDown(alphaCell, { key: "ArrowDown", shiftKey: true });
      await waitFor(() =>
        expect(onCellRangeChange).toHaveBeenCalledWith({
          end: beta,
          start: alpha,
        }),
      );
      const betaCell = screen
        .getByText("Beta")
        .closest<HTMLElement>('[role="gridcell"]');
      if (betaCell === null) throw new Error(`Missing ${name} Beta cell`);
      const setClipboardData = vi.fn();
      fireEvent.copy(betaCell, {
        clipboardData: { setData: setClipboardData },
      });
      expect(setClipboardData).toHaveBeenCalledWith(
        "text/plain",
        "Alpha\nBeta",
      );
      expect(onCopyCell).toHaveBeenCalledWith(
        expect.objectContaining({ anchor: beta }),
      );

      fireEvent.keyDown(betaCell, { ctrlKey: true, key: "d" });
      await waitFor(() => expect(onFillCells).toHaveBeenCalledTimes(1));
      expect(onFillCells).toHaveBeenCalledWith(
        expect.objectContaining({
          range: { end: beta, start: alpha },
          targets: [expect.objectContaining(beta)],
        }),
      );
      fireEvent.paste(betaCell, {
        clipboardData: { getData: () => "Gamma" },
      });
      expect(onPaste).toHaveBeenCalledWith(
        expect.objectContaining({ target: expect.objectContaining(beta) }),
      );
      expect(
        handle.current?.planPasteTargets(alpha, {
          columnCount: 1,
          rowCount: 2,
        }),
      ).toEqual(
        expect.objectContaining({
          columns: ["label"],
          rowTargets: expect.any(Array),
        }),
      );
      expect(handle.current?.focusRoot()).toBe(true);
      expect(document.activeElement).toBe(handle.current?.getScrollElement());

      view.unmount();
      cleanup();
    }
  });

  it("runs editor outcomes and draft focus through both semantic bindings", async () => {
    for (const { Grid, name } of semanticContractBindings) {
      const handle = createRef<GridHandle>();
      const commit = vi
        .fn()
        .mockResolvedValueOnce({
          kind: "validation_error",
          message: "Correct the contract value.",
        })
        .mockResolvedValueOnce({ kind: "accepted" });
      const view = render(
        <Grid
          ref={handle}
          columns={[
            {
              contractWritable: true,
              editor: {
                commit,
                initialDraftValue: (row: HarnessRow) => row.label,
                renderEditor: (context) => (
                  <div>
                    <input
                      aria-label={`Contract editor ${name}`}
                      ref={context.focusTargetRef}
                      value={String(context.draftValue)}
                      onChange={(event) =>
                        context.setDraftValue(event.currentTarget.value)
                      }
                    />
                    <button type="button" onClick={() => void context.commit()}>
                      Save contract value
                    </button>
                    {context.outcome?.kind === "validation_error" ? (
                      <span>{context.outcome.message}</span>
                    ) : null}
                  </div>
                ),
              },
              fieldKey: "label",
              label: "Label",
              renderCell: ({ row }) => row.label,
              renderDraftCell: ({ focusTargetRef }) => (
                <input
                  aria-label={`Contract draft ${name}`}
                  ref={focusTargetRef}
                />
              ),
            },
          ]}
          dataRows={[
            {
              data: { label: "Alpha", state: "open" },
              kind: "data",
              mutationIdentity: {
                kind: "core_row_version",
                baseRowVersion: 1,
              },
              rowIdentity: { kind: "core_record", recordId: "contract-1" },
            },
          ]}
          draftRow={{
            data: { label: "", state: "open" },
            kind: "draft",
          }}
          surface={testSurface}
        />,
      );
      const anchor = gridAnchor("contract-1", "label");

      expect(handle.current?.focusDraftCell("label")).toBe(true);
      expect(document.activeElement).toBe(
        screen.getByRole("textbox", { name: `Contract draft ${name}` }),
      );
      expect(handle.current?.activateEdit(anchor)).toBe(true);
      expect(
        handle.current?.activateEdit({
          ...anchor,
          fieldKey: "missing",
        }),
      ).toBe(false);
      expect(
        await screen.findByRole("textbox", { name: `Contract editor ${name}` }),
      ).toBeTruthy();
      fireEvent.click(
        screen.getByRole("button", { name: "Save contract value" }),
      );
      expect(
        (await screen.findAllByText("Correct the contract value.")).length,
      ).toBeGreaterThan(0);
      expect(
        screen.getByRole("textbox", { name: `Contract editor ${name}` }),
      ).toBeTruthy();
      fireEvent.click(
        screen.getByRole("button", { name: "Save contract value" }),
      );
      await waitFor(() =>
        expect(
          screen.queryByRole("textbox", { name: `Contract editor ${name}` }),
        ).toBeNull(),
      );

      view.unmount();
      cleanup();
    }
  });

  it("renders, selects, and focuses extension resources without Core identities", async () => {
    const handle = createRef<GridHandle>();
    const onSelectRow = vi.fn();
    const extensionSurface = {
      kind: "extension_grid",
      extensionProfileId: "network_flow_activity",
      workspaceKey: "incident-1",
      gridSchemaId: "network_flow.accepted_rows.v1",
    } as const;
    const rowIdentity = {
      kind: "extension_resource",
      extensionProfileId: "network_flow_activity",
      resourceKind: "accepted_flow_row",
      resourceId: "flow-row-1",
    } as const;

    render(
      <SemanticDataGrid
        ref={handle}
        accessibleLabel="Accepted Network Flow resources"
        columns={[
          {
            fieldKey: "label",
            getClipboardValue: (row) => row.label,
            label: "Label",
            renderCell: ({ row }) => (
              <span data-testid="extension-live-cell">{row.label}</span>
            ),
          },
        ]}
        dataRows={[
          {
            data: { label: "Flow 1", state: "accepted" },
            kind: "data",
            rowIdentity,
            testId: "extension-live-row",
          },
        ]}
        interactionMode={{ kind: "read_only", label: "Analysis is read-only" }}
        onSelectRow={onSelectRow}
        surface={extensionSurface}
      />,
    );

    expect(
      screen.getByRole("grid", { name: "Accepted Network Flow resources" }),
    ).toBeTruthy();
    const row = await screen.findByTestId("extension-live-row");
    expect(row.getAttribute("data-grid-row-identity-kind")).toBe(
      "extension_resource",
    );
    expect(row.getAttribute("data-grid-record-id")).toBeNull();
    const cell = await screen.findByTestId("extension-live-cell");
    fireEvent.click(cell.closest('[role="gridcell"]') as HTMLElement);
    expect(onSelectRow).toHaveBeenCalledWith(rowIdentity);
    const anchor = {
      fieldKey: "label",
      rowIdentity,
      surface: extensionSurface,
    };
    expect(handle.current?.focusAnchor(anchor)).toBe(true);
  });

  it("groups extension resources in the live grid without changing their identities", async () => {
    render(
      <SemanticDataGrid
        columns={columns}
        dataRows={[
          {
            data: { label: "Flow 1", state: "accepted" },
            kind: "data",
            rowIdentity: {
              kind: "extension_resource",
              extensionProfileId: "network_flow_activity",
              resourceKind: "graph_contributor",
              resourceId: "contributor-1",
            },
            testId: "extension-grouped-row",
          },
        ]}
        grouping={{
          fieldKey: "state",
          formatLabel: (value) => (value === null ? null : String(value)),
          getTestId: () => "extension-group-row",
          getValue: (row) => row.state,
        }}
        surface={{
          kind: "extension_grid",
          extensionProfileId: "network_flow_activity",
          workspaceKey: "incident-1",
          gridSchemaId: "network_flow.graph_contributors.v1",
        }}
      />,
    );

    expect(await screen.findByTestId("extension-group-row")).toBeTruthy();
    const row = await screen.findByTestId("extension-grouped-row");
    expect(row.getAttribute("data-grid-row-identity-kind")).toBe(
      "extension_resource",
    );
    expect(row.getAttribute("data-grid-record-id")).toBeNull();
  });

  it("fails closed when untyped callers attach Core mutation capabilities to extension grids", () => {
    const unsafe = {
      columns,
      dataRows: [],
      interactionMode: { kind: "editable" },
      onFillCells: vi.fn(),
      surface: {
        kind: "extension_grid",
        extensionProfileId: "network_flow_activity",
        workspaceKey: "incident-1",
        gridSchemaId: "network_flow.accepted_rows.v1",
      },
    } as unknown as SemanticDataGridProps<HarnessRow>;

    expect(() => render(<SemanticDataGrid {...unsafe} />)).toThrow(
      /cannot enable Core mutation/i,
    );
  });

  it("renders semantic data states without manufacturing rows and keeps interaction mode independent", async () => {
    const onAction = vi.fn();
    const { container, rerender } = render(
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        columns={columns}
        dataState={{
          generationKey: "records-1",
          kind: "initial_loading",
          surfaceLabel: "Records",
        }}
        interactionMode={{ kind: "read_only", label: "Closed, read-only" }}
        dataRows={[]}
      />,
    );

    expect(screen.getByText("Loading Records…")).toBeTruthy();
    expect(screen.getByText("Closed, read-only")).toBeTruthy();
    await waitFor(() => {
      expect(screen.getByRole("grid").getAttribute("aria-busy")).toBe("true");
      expect(screen.getByRole("grid").getAttribute("aria-readonly")).toBe(
        "true",
      );
    });
    expect(
      container.querySelector('[data-cartulary-grid-row-kind="data"]'),
    ).toBeNull();

    rerender(
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        columns={columns}
        dataState={{
          action: { label: "Clear filters", onInvoke: onAction },
          kind: "filtered_empty",
        }}
        dataRows={[]}
      />,
    );
    expect(screen.getByText("No rows match the current filters.")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
    expect(onAction).toHaveBeenCalledOnce();

    rerender(
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        columns={columns}
        dataState={{
          kind: "stale_error",
          message: "Refresh failed.",
        }}
        dataRows={[
          {
            data: { label: "Retained", state: "saved" },
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 4 },
            rowIdentity: { kind: "core_record", recordId: "record-retained" },
          },
        ]}
      />,
    );
    expect(
      screen.getByText("Refresh failed. Previously loaded rows may be stale."),
    ).toBeTruthy();
    expect(
      container.querySelector('[data-grid-record-id="record-retained"]'),
    ).toBeTruthy();
  });

  it("shows the delayed loading message once per active generation and cancels it on terminal state", () => {
    vi.useFakeTimers();
    const { rerender } = render(
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        columns={columns}
        dataRows={[]}
        dataState={{
          generationKey: "generation-1",
          kind: "initial_loading",
          surfaceLabel: "Records",
        }}
      />,
    );
    act(() => vi.advanceTimersByTime(1999));
    expect(screen.queryByText("Still loading this surface")).toBeNull();
    act(() => vi.advanceTimersByTime(1));
    expect(screen.getByText("Still loading this surface")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Retry" })).toBeNull();

    rerender(
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        columns={columns}
        dataRows={[]}
        dataState={{
          generationKey: "generation-2",
          kind: "initial_loading",
          surfaceLabel: "Records",
        }}
      />,
    );
    expect(screen.getByText("Loading Records…")).toBeTruthy();
    rerender(
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        columns={columns}
        dataRows={[]}
        dataState={{ kind: "ready" }}
      />,
    );
    act(() => vi.advanceTimersByTime(2000));
    expect(screen.queryByText("Still loading this surface")).toBeNull();
    vi.useRealTimers();
  });

  it("compiles semantic row and cell state into private classes, markers, and ARIA", async () => {
    const statefulRow: GridDataRow<HarnessRow> = {
      data: { label: "Alpha", state: "invalid" },
      kind: "data",
      mutationIdentity: { kind: "core_row_version", baseRowVersion: 7 },
      rowIdentity: { kind: "core_record", recordId: "record-stateful" },
      testId: "stateful-row",
    };

    render(
      <SemanticDataGrid
        activeRowIdentity={{ kind: "core_record", recordId: "record-stateful" }}
        coreRecordBulkSelection={{
          onSelectedRecordIdsChange: vi.fn(),
          selectedRecordIds: new Set(["record-stateful"]),
        }}
        columns={columns}
        dataState={{ kind: "stale_error", message: "Refresh failed." }}
        getCellState={({ anchor }) =>
          anchor.fieldKey === "label"
            ? { conflicted: true }
            : { invalid: { message: "Choose an allowed state" } }
        }
        getRowState={() => ({ pending: true })}
        dataRows={[statefulRow]}
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
      />,
    );

    const row = await screen.findByTestId("stateful-row");
    await waitFor(() => {
      expect(row.getAttribute("aria-busy")).toBe("true");
      expect(row.getAttribute("aria-current")).toBe("true");
      expect(row.getAttribute("aria-selected")).toBe("true");
      expect(row.className).toContain("cartulary-grid-row-is-pending");
      expect(row.className).toContain("cartulary-grid-row-is-stale");
    });

    const labelContent = row.querySelector<HTMLElement>(
      '[data-grid-field-key="label"]',
    );
    const stateContent = row.querySelector<HTMLElement>(
      '[data-grid-field-key="state"]',
    );
    const labelCell = labelContent?.closest<HTMLElement>('[role="gridcell"]');
    const stateCell = stateContent?.closest<HTMLElement>('[role="gridcell"]');
    expect(labelCell?.dataset.gridPrimaryState).toBe("conflicted");
    expect(labelCell?.getAttribute("aria-readonly")).toBe("true");
    expect(
      labelContent?.querySelector('[aria-label="Conflict on Label"]'),
    ).toBeTruthy();
    expect(stateCell?.dataset.gridPrimaryState).toBe("invalid");
    expect(stateCell?.getAttribute("aria-invalid")).toBe("true");
    expect(stateCell?.getAttribute("aria-description")).toContain(
      "Choose an allowed state",
    );
    expect(
      stateContent?.querySelector(
        '[aria-label="Invalid State: Choose an allowed state"]',
      ),
    ).toBeTruthy();

    const supportHandle = createRef<GridHandle>();
    const onSupportAction = vi.fn();
    const onSupportActiveCellChange = vi.fn();
    const onSupportSelectRow = vi.fn();
    render(
      <SemanticDataGridTestSupport
        ref={supportHandle}
        actionsColumn={{
          label: "Actions",
          renderCell: () => (
            <button
              data-testid="support-row-action"
              type="button"
              onClick={onSupportAction}
            >
              Inspect
            </button>
          ),
        }}
        activeRowIdentity={{ kind: "core_record", recordId: "record-stateful" }}
        coreRecordBulkSelection={{
          onSelectedRecordIdsChange: vi.fn(),
          selectedRecordIds: new Set(["record-stateful"]),
        }}
        columns={columns}
        dataState={{ kind: "stale_error", message: "Refresh failed." }}
        getCellState={({ anchor }) =>
          anchor.fieldKey === "label"
            ? { conflicted: true }
            : { invalid: { message: "Choose an allowed state" } }
        }
        getRowState={() => ({ pending: true })}
        onActiveCellChange={onSupportActiveCellChange}
        onSelectRow={onSupportSelectRow}
        dataRows={[{ ...statefulRow, testId: "stateful-row-test-support" }]}
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
      />,
    );
    const supportRow = screen.getByTestId("stateful-row-test-support");
    const supportLabelCell = supportRow.querySelector<HTMLElement>(
      '[role="gridcell"][data-grid-field-key="label"]',
    );
    const supportStateCell = supportRow.querySelector<HTMLElement>(
      '[role="gridcell"][data-grid-field-key="state"]',
    );
    expect(supportRow.getAttribute("aria-busy")).toBe("true");
    expect(supportRow.getAttribute("aria-current")).toBe("true");
    expect(supportRow.getAttribute("aria-selected")).toBe("true");
    expect(supportLabelCell?.dataset.gridPrimaryState).toBe("conflicted");
    expect(supportStateCell?.dataset.gridPrimaryState).toBe("invalid");
    const labelAnchor = gridAnchor("record-stateful", "label");
    const stateAnchor = gridAnchor("record-stateful", "state");
    expect(supportHandle.current?.focusAnchor(labelAnchor)).toBe(true);
    expect(onSupportActiveCellChange).toHaveBeenCalledTimes(1);
    expect(onSupportActiveCellChange).toHaveBeenLastCalledWith(labelAnchor);
    expect(supportHandle.current?.focusAnchor(labelAnchor)).toBe(true);
    expect(onSupportActiveCellChange).toHaveBeenCalledTimes(1);
    expect(
      supportHandle.current?.moveFocus(labelAnchor, { key: "ArrowRight" }),
    ).toEqual(stateAnchor);
    expect(onSupportActiveCellChange).toHaveBeenCalledTimes(2);
    expect(onSupportActiveCellChange).toHaveBeenLastCalledWith(stateAnchor);
    expect(
      supportHandle.current?.planPasteTargets(labelAnchor, {
        columnCount: 1,
        rowCount: 1,
      }),
    ).toBeNull();
    const wrongSurfaceAnchor = {
      ...labelAnchor,
      surface: { kind: "view_schema", viewSchemaId: "wrong.view" } as const,
    };
    expect(supportHandle.current?.focusAnchor(wrongSurfaceAnchor)).toBe(false);
    expect(
      supportHandle.current?.planPasteTargets(wrongSurfaceAnchor, {
        columnCount: 1,
        rowCount: 1,
      }),
    ).toBeNull();
    fireEvent.click(screen.getByTestId("support-row-action"));
    expect(onSupportAction).toHaveBeenCalledTimes(1);
    expect(onSupportSelectRow).not.toHaveBeenCalled();
  });

  it("keeps inspector context distinct from opt-in record selection", async () => {
    function SelectionHarness() {
      const [selectedRecordIds, setSelectedRecordIds] = useState<
        ReadonlySet<string>
      >(() => new Set());
      return (
        <SemanticDataGrid
          activeRowIdentity={{ kind: "core_record", recordId: "record-2" }}
          coreRecordBulkSelection={{
            isRecordSelectable: (row) =>
              row.rowIdentity.kind === "core_record" &&
              row.rowIdentity.recordId !== "record-2",
            onSelectedRecordIdsChange: setSelectedRecordIds,
            selectedRecordIds,
          }}
          columns={columns}
          dataRows={[
            {
              kind: "data",
              mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
              rowIdentity: { kind: "core_record", recordId: "record-1" },
              data: { label: "Alpha", state: "open" },
            },
            {
              kind: "data",
              mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
              rowIdentity: { kind: "core_record", recordId: "record-2" },
              data: { label: "Beta", state: "reviewed" },
            },
          ]}
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        />
      );
    }

    render(<SelectionHarness />);
    const first = await screen.findByRole("checkbox", {
      name: "Select record record-1",
    });
    expect(
      screen.queryByRole("checkbox", { name: "Select record record-2" }),
    ).toBeNull();
    fireEvent.click(first);
    expect((first as HTMLInputElement).checked).toBe(true);
    const activeRow = document.querySelector(
      '[data-grid-record-id="record-2"]',
    );
    expect(activeRow?.getAttribute("data-inspector-active")).toBe("true");
    expect(activeRow?.getAttribute("aria-selected")).not.toBe("true");
    fireEvent.click(
      screen.getByRole("checkbox", {
        name: "Select all records on this page",
      }),
    );
    expect((first as HTMLInputElement).checked).toBe(false);
  });

  it("keeps the bottom create draft recordless and usable with zero committed rows", async () => {
    const onPasteCell = vi.fn();
    const handle = createRef<GridHandle>();
    render(
      <SemanticDataGrid
        ref={handle}
        coreRecordBulkSelection={{
          onSelectedRecordIdsChange: vi.fn(),
          selectedRecordIds: new Set(),
        }}
        columns={[
          {
            contractWritable: true,
            fieldKey: "label",
            label: "Label",
            renderCell: ({ row }) => row.label,
            renderDraftCell: ({ focusTargetRef }) => (
              <input
                aria-label="Zero-row create draft"
                defaultValue=""
                ref={focusTargetRef}
              />
            ),
          },
          {
            fieldKey: "state",
            label: "State",
            renderCell: ({ row }) => row.state,
            renderDraftCell: ({ focusTargetRef }) => (
              <input
                aria-label="Disabled create draft"
                disabled
                ref={focusTargetRef}
              />
            ),
          },
        ]}
        draftRow={{
          data: { label: "", state: "open" },
          kind: "draft",
          testId: "zero-row-create-draft",
        }}
        clipboardPaste={{
          decode: decodeTestClipboard,
          onPaste: onPasteCell,
        }}
        dataRows={[]}
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
      />,
    );

    const draftInput = await screen.findByRole("textbox", {
      name: "Zero-row create draft",
    });
    const draftRow = draftInput.closest('[role="row"]');
    expect(draftRow).toBeTruthy();
    expect(draftRow?.getAttribute("data-cartulary-grid-draft-row")).toBe(
      "true",
    );
    expect(draftRow?.getAttribute("data-grid-record-id")).toBeNull();
    expect(draftRow?.querySelectorAll('input[type="checkbox"]')).toHaveLength(
      0,
    );
    expect(document.querySelectorAll("[data-grid-record-id]")).toHaveLength(0);
    expect(handle.current?.focusDraftCell("missing")).toBe(false);
    expect(handle.current?.focusDraftCell("state")).toBe(false);
    expect(handle.current?.focusDraftCell("label")).toBe(true);
    expect(document.activeElement).toBe(draftInput);
    fireEvent.change(draftInput, { target: { value: "Draft remains usable" } });
    expect((draftInput as HTMLInputElement).value).toBe("Draft remains usable");
    fireEvent.paste(draftInput, {
      clipboardData: { getData: () => "must not become a record paste" },
    });
    expect(onPasteCell).not.toHaveBeenCalled();
  });

  it("keeps test-support draft focus and committed measurement semantic", () => {
    const handle = createRef<GridHandle>();
    render(
      <SemanticDataGridTestSupport
        ref={handle}
        columns={[
          {
            fieldKey: "label",
            label: "Label",
            renderCell: ({ row }) => row.label,
            renderDraftCell: ({ focusTargetRef }) => (
              <input aria-label="Support draft" ref={focusTargetRef} />
            ),
          },
        ]}
        dataRows={[
          {
            data: { label: "Alpha", state: "open" },
            kind: "data",
            rowIdentity: { kind: "core_record", recordId: "record-1" },
          },
        ]}
        draftRow={{
          data: { label: "", state: "open" },
          kind: "draft",
        }}
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
      />,
    );

    const input = screen.getByRole("textbox", { name: "Support draft" });
    expect(handle.current?.focusDraftCell("label")).toBe(true);
    expect(document.activeElement).toBe(input);
    expect(
      handle.current?.getAnchorRect(gridAnchor("record-1", "label")),
    ).not.toBeNull();
    expect(
      handle.current?.getAnchorRect({
        ...gridAnchor("record-1", "label"),
        surface: { kind: "view_schema", viewSchemaId: "wrong.view" },
      }),
    ).toBeNull();
  });

  it("translates live cell events and the restricted handle through semantic coordinates", async () => {
    const onActiveCellChange = vi.fn();
    const onCopyCell = vi.fn();
    const onPasteCell = vi.fn();
    const handle = createRef<GridHandle>();
    render(
      <SemanticDataGrid
        ref={handle}
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        columns={[
          {
            contractWritable: true,
            fieldKey: "label",
            label: "Label",
            renderCell: ({ row }) => (
              // biome-ignore lint/a11y/noStaticElementInteractions lint/a11y/useKeyWithClickEvents: this test fixture deliberately models vendor click interception on passive cell content.
              <span
                data-testid="semantic-live-cell"
                onClick={(event) => event.stopPropagation()}
              >
                {row.label}
              </span>
            ),
            editor: {
              commit: async () => ({ kind: "accepted" }),
              initialDraftValue: (row) => row.label,
              renderEditor: ({ cancel, draftValue, focusTargetRef }) => (
                <input
                  aria-label="Semantic editor"
                  defaultValue={String(draftValue)}
                  ref={focusTargetRef}
                  onKeyDown={(event) => {
                    if (event.key === "Escape") cancel();
                  }}
                />
              ),
            },
          },
        ]}
        onActiveCellChange={onActiveCellChange}
        onCopyCell={onCopyCell}
        clipboardPaste={{
          decode: decodeTestClipboard,
          onPaste: onPasteCell,
        }}
        dataRows={[
          {
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 7 },
            rowIdentity: { kind: "core_record", recordId: "record-1" },
            data: { label: "Alpha", state: "open" },
          },
        ]}
      />,
    );
    const cellContent = await screen.findByTestId("semantic-live-cell");
    const cell = cellContent.closest('[role="gridcell"]');
    if (!(cell instanceof HTMLElement))
      throw new Error("Expected live RDG cell");
    // Exercise the production pointer path: users click the rendered static
    // content, not the vendor-owned gridcell wrapper.
    fireEvent.mouseDown(cellContent);
    fireEvent.mouseUp(cellContent);
    fireEvent.click(cellContent);
    await waitFor(() => expect(onActiveCellChange).toHaveBeenCalled());
    const editor = await screen.findByRole("textbox", {
      name: "Semantic editor",
    });
    expect(document.activeElement).toBe(editor);
    expect((editor as HTMLInputElement).selectionStart).toBe(5);
    expect((editor as HTMLInputElement).selectionEnd).toBe(5);
    fireEvent.keyDown(editor, { key: "Escape" });
    const liveGrid = screen.getByRole("grid");
    fireEvent.copy(liveGrid);
    fireEvent.paste(liveGrid);
    const anchor = gridAnchor("record-1", "label");
    const target = {
      ...anchor,
      mutationIdentity: {
        kind: "core_row_version" as const,
        baseRowVersion: 7,
      },
    };
    expect(onActiveCellChange).toHaveBeenCalledWith(anchor);
    expect(onCopyCell).toHaveBeenCalledWith(
      expect.objectContaining({ anchor }),
    );
    expect(onPasteCell).toHaveBeenCalledWith(
      expect.objectContaining({
        input: { kind: "scalar", rawText: "", value: "" },
        target,
      }),
    );
    expect(handle.current?.getScrollElement()).toBeTruthy();
    expect(handle.current?.getAnchorRect(anchor)).not.toBeNull();
    expect(
      handle.current?.getAnchorRect({
        ...anchor,
        surface: { kind: "view_schema", viewSchemaId: "wrong.view" },
      }),
    ).toBeNull();
    expect(
      handle.current?.planPasteTargets(anchor, {
        columnCount: 1,
        rowCount: 1,
      }),
    ).toEqual({
      columns: ["label"],
      rowTargets: [
        {
          kind: "record",
          mutationIdentity: {
            kind: "core_row_version",
            baseRowVersion: 7,
          },
          rowIdentity: { kind: "core_record", recordId: "record-1" },
          surface: testSurface,
        },
      ],
    });
    expect(handle.current?.focusRoot()).toBe(true);
    expect(document.activeElement).toBe(handle.current?.getScrollElement());
    expect(handle.current?.scrollToAnchor(anchor)).toBe(true);
    const externalButton = document.createElement("button");
    document.body.append(externalButton);
    externalButton.focus();
    expect(document.activeElement).toBe(externalButton);
    expect(handle.current?.focusAnchor(anchor)).toBe(true);
    await waitFor(() => {
      const currentContent = document.querySelector(
        '[role="row"][data-grid-record-id="record-1"] [data-grid-field-key="label"]',
      );
      expect(document.activeElement).toBe(
        currentContent?.closest('[role="gridcell"]'),
      );
    });
    expect(handle.current?.focusAnchor(anchor)).toBe(true);
    await new Promise<void>((resolve) => {
      window.setTimeout(() => {
        externalButton.focus();
        resolve();
      }, 0);
    });
    await new Promise<void>((resolve) => window.setTimeout(resolve, 0));
    expect(document.activeElement).toBe(externalButton);
    const interruptingButton = document.createElement("button");
    document.body.append(interruptingButton);
    expect(handle.current?.focusAnchor(anchor)).toBe(true);
    interruptingButton.focus();
    await new Promise<void>((resolve) => window.setTimeout(resolve, 0));
    await new Promise<void>((resolve) => window.setTimeout(resolve, 0));
    expect(document.activeElement).toBe(interruptingButton);
    interruptingButton.remove();
    externalButton.remove();
    expect(
      handle.current?.focusAnchor({
        ...anchor,
        surface: { kind: "view_schema", viewSchemaId: "wrong.view" },
      }),
    ).toBe(false);
  });

  it("owns semantic navigation, range extension, keyboard entry, cancellation, and grid exit", async () => {
    const onActiveCellChange = vi.fn();
    const commit = vi.fn().mockResolvedValue({ kind: "accepted" });
    const keyboardColumns: readonly GridColumn<HarnessRow>[] = [
      {
        contractWritable: true,
        editor: {
          clearDraftValue: "",
          commit,
          initialDraftValue: (row) => row.label,
          renderEditor: (context) => (
            <input
              aria-label="Keyboard editor"
              ref={context.focusTargetRef}
              value={String(context.draftValue)}
              onChange={(event) =>
                context.setDraftValue(event.currentTarget.value)
              }
            />
          ),
        },
        fieldKey: "label",
        label: "Label",
        renderCell: ({ row }) => (
          <span data-testid={`keyboard-${row.label}`}>{row.label}</span>
        ),
      },
      {
        fieldKey: "state",
        label: "State",
        renderCell: ({ row }) => row.state,
      },
    ];
    render(
      <>
        <button type="button">Before grid</button>
        <SemanticDataGrid
          columns={keyboardColumns}
          onActiveCellChange={onActiveCellChange}
          dataRows={[
            {
              data: { label: "Alpha", state: "open" },
              kind: "data",
              mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
              rowIdentity: { kind: "core_record", recordId: "record-1" },
            },
            {
              data: { label: "Beta", state: "open" },
              kind: "data",
              mutationIdentity: { kind: "core_row_version", baseRowVersion: 2 },
              rowIdentity: { kind: "core_record", recordId: "record-2" },
            },
            {
              data: { label: "Gamma", state: "closed" },
              kind: "data",
              mutationIdentity: { kind: "core_row_version", baseRowVersion: 3 },
              rowIdentity: { kind: "core_record", recordId: "record-3" },
            },
          ]}
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
        />
        <button type="button">After grid</button>
      </>,
    );

    const alphaCell = (await screen.findByTestId("keyboard-Alpha")).closest(
      '[role="gridcell"]',
    );
    if (!(alphaCell instanceof HTMLElement)) {
      throw new Error("Expected Alpha grid cell");
    }
    fireEvent.mouseDown(alphaCell);
    fireEvent.keyDown(alphaCell, { key: "ArrowDown", shiftKey: true });
    expect(
      await screen.findByText("Selected 2 rows by 1 columns."),
    ).toBeTruthy();
    const betaCell = screen
      .getByTestId("keyboard-Beta")
      .closest<HTMLElement>('[role="gridcell"]');
    expect(betaCell?.getAttribute("aria-selected")).toBe("true");

    fireEvent.keyDown(betaCell as HTMLElement, {
      ctrlKey: true,
      key: "End",
    });
    await waitFor(() =>
      expect(onActiveCellChange).toHaveBeenLastCalledWith(
        gridAnchor("record-3", "state"),
      ),
    );

    fireEvent.keyDown(
      screen
        .getByTestId("keyboard-Gamma")
        .closest('[role="gridcell"]') as HTMLElement,
      { ctrlKey: true, key: "Home" },
    );
    await waitFor(() =>
      expect(onActiveCellChange).toHaveBeenLastCalledWith(
        gridAnchor("record-1", "label"),
      ),
    );

    fireEvent.keyDown(alphaCell, { key: "Z" });
    const editor = await screen.findByRole("textbox", {
      name: "Keyboard editor",
    });
    expect((editor as HTMLInputElement).value).toBe("Z");
    fireEvent.keyDown(editor, { key: "Escape" });
    await waitFor(() =>
      expect(
        screen.queryByRole("textbox", { name: "Keyboard editor" }),
      ).toBeNull(),
    );
    expect(commit).not.toHaveBeenCalled();

    fireEvent.keyDown(
      screen
        .getByTestId("keyboard-Alpha")
        .closest('[role="gridcell"]') as HTMLElement,
      { key: "Delete" },
    );
    expect(
      (
        await screen.findByRole("textbox", { name: "Keyboard editor" })
      ).getAttribute("value"),
    ).toBe("");
    fireEvent.keyDown(
      screen.getByRole("textbox", { name: "Keyboard editor" }),
      { key: "Escape" },
    );
    fireEvent.keyDown(
      screen
        .getByTestId("keyboard-Alpha")
        .closest('[role="gridcell"]') as HTMLElement,
      { key: "Enter" },
    );
    const acceptedEditor = await screen.findByRole("textbox", {
      name: "Keyboard editor",
    });
    fireEvent.change(acceptedEditor, { target: { value: "Accepted" } });
    fireEvent.keyDown(acceptedEditor, { key: "Enter" });
    await waitFor(() => expect(commit).toHaveBeenCalledTimes(1));
    await waitFor(() =>
      expect(onActiveCellChange).toHaveBeenLastCalledWith(
        gridAnchor("record-2", "label"),
      ),
    );
    fireEvent.keyDown(
      screen
        .getByTestId("keyboard-Beta")
        .closest('[role="gridcell"]') as HTMLElement,
      { key: "Tab" },
    );
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "After grid" }),
    );
  });

  it("keeps embedded actions independent and routes labeled pointer and keyboard fill through semantic targets", async () => {
    const onAction = vi.fn();
    const onFillCells = vi.fn();
    const fillColumns: readonly GridColumn<HarnessRow>[] = [
      {
        contractWritable: true,
        editor: {
          commit: async () => ({ kind: "accepted" }),
          initialDraftValue: (row) => row.label,
          renderEditor: (context) => (
            <input
              aria-label="Fill editor"
              ref={context.focusTargetRef}
              value={String(context.draftValue)}
              onChange={(event) =>
                context.setDraftValue(event.currentTarget.value)
              }
            />
          ),
        },
        fieldKey: "label",
        label: "Label",
        renderCell: ({ row }) => (
          <span data-testid={`fill-${row.label}`}>{row.label}</span>
        ),
      },
      {
        fieldKey: "state",
        label: "State",
        renderCell: ({ row }) => (
          <button type="button" onClick={onAction}>
            Inspect {row.state}
          </button>
        ),
      },
    ];
    render(
      <SemanticDataGrid
        columns={fillColumns}
        onFillCells={onFillCells}
        dataRows={[
          {
            data: { label: "Alpha", state: "open" },
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
            rowIdentity: { kind: "core_record", recordId: "record-1" },
          },
          {
            data: { label: "Beta", state: "open" },
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 2 },
            rowIdentity: { kind: "core_record", recordId: "record-2" },
          },
        ]}
        surface={testSurface}
      />,
    );

    fireEvent.click(
      screen.getAllByRole("button", { name: "Inspect open" })[0] as HTMLElement,
    );
    expect(onAction).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("textbox", { name: "Fill editor" })).toBeNull();

    const alphaCell = screen
      .getByTestId("fill-Alpha")
      .closest<HTMLElement>('[role="gridcell"]');
    if (alphaCell === null) throw new Error("Expected fill source cell");
    fireEvent.mouseDown(alphaCell);
    const fillHandle = await waitFor(() => {
      const handle = document.querySelector<HTMLElement>(
        ".rdg-cell-drag-handle",
      );
      expect(handle).toBeTruthy();
      expect(handle?.getAttribute("aria-label")).toBe(
        "Drag to fill this value",
      );
      return handle as HTMLElement;
    });
    fireEvent.doubleClick(fillHandle);
    expect(onFillCells).not.toHaveBeenCalled();

    fireEvent.keyDown(alphaCell, { key: "ArrowDown", shiftKey: true });
    const betaCell = screen
      .getByTestId("fill-Beta")
      .closest<HTMLElement>('[role="gridcell"]');
    if (betaCell === null) throw new Error("Expected fill target cell");
    fireEvent.keyDown(betaCell, { ctrlKey: true, key: "d" });
    await waitFor(() => expect(onFillCells).toHaveBeenCalledTimes(1));
    expect(onFillCells).toHaveBeenCalledWith({
      range: {
        end: gridAnchor("record-2", "label"),
        start: gridAnchor("record-1", "label"),
      },
      source: {
        ...gridAnchor("record-1", "label"),
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
      },
      target: {
        ...gridAnchor("record-2", "label"),
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 2 },
      },
      targets: [
        {
          ...gridAnchor("record-2", "label"),
          mutationIdentity: { kind: "core_row_version", baseRowVersion: 2 },
        },
      ],
    });
  });

  it("does not republish an unchanged semantic anchor when the active record updates", async () => {
    const onActiveCellChange = vi.fn();

    function UpdatingRecordHarness() {
      const [version, setVersion] = useState(1);
      return (
        <>
          <button type="button" onClick={() => setVersion(2)}>
            Apply record update
          </button>
          <SemanticDataGrid
            surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
            columns={columns}
            onActiveCellChange={onActiveCellChange}
            dataRows={[
              {
                kind: "data",
                mutationIdentity: {
                  kind: "core_row_version",
                  baseRowVersion: version,
                },
                rowIdentity: { kind: "core_record", recordId: "record-1" },
                data: { label: `Alpha ${version}`, state: "open" },
              },
            ]}
          />
        </>
      );
    }

    render(<UpdatingRecordHarness />);
    const cell = screen.getByText("Alpha 1").closest('[role="gridcell"]');
    if (!(cell instanceof HTMLElement)) throw new Error("Expected live cell");
    fireEvent.mouseDown(cell);
    fireEvent.click(cell);
    await waitFor(() => expect(onActiveCellChange).toHaveBeenCalledTimes(1));

    fireEvent.click(
      screen.getByRole("button", { name: "Apply record update" }),
    );

    expect(await screen.findByText("Alpha 2")).toBeTruthy();
    expect(onActiveCellChange).toHaveBeenCalledTimes(1);
  });

  it("retains semantic editor drafts for validation outcomes and closes only after acceptance", async () => {
    const commit = vi
      .fn()
      .mockResolvedValueOnce({
        kind: "validation_error",
        message: "Enter a valid value.",
      })
      .mockResolvedValueOnce({ kind: "accepted" });
    render(
      <SemanticDataGrid
        columns={[
          {
            contractWritable: true,
            editor: {
              commit,
              initialDraftValue: (row: HarnessRow) => row.label,
              renderEditor: (context) => (
                <div>
                  <input
                    aria-label="Retained semantic draft"
                    ref={context.focusTargetRef}
                    value={String(context.draftValue)}
                    onChange={(event) =>
                      context.setDraftValue(event.currentTarget.value)
                    }
                  />
                  <button type="button" onClick={() => void context.commit()}>
                    Commit semantic edit
                  </button>
                </div>
              ),
            },
            fieldKey: "label",
            label: "Label",
            renderCell: ({ row }) => (
              <span data-testid="semantic-edit-cell">{row.label}</span>
            ),
          },
        ]}
        dataRows={[
          {
            data: { label: "Alpha", state: "open" },
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 7 },
            rowIdentity: { kind: "core_record", recordId: "record-1" },
          },
        ]}
        surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
      />,
    );
    const cell = (await screen.findByTestId("semantic-edit-cell")).closest(
      '[role="gridcell"]',
    );
    if (!(cell instanceof HTMLElement)) {
      throw new Error("Expected live RDG cell");
    }
    fireEvent.mouseDown(cell);
    fireEvent.click(cell);
    const input = await screen.findByRole("textbox", {
      name: "Retained semantic draft",
    });
    fireEvent.change(input, { target: { value: "invalid draft" } });
    fireEvent.click(
      screen.getByRole("button", { name: "Commit semantic edit" }),
    );
    expect((await screen.findByRole("alert")).textContent).toContain(
      "Enter a valid value.",
    );
    expect((input as HTMLInputElement).value).toBe("invalid draft");
    fireEvent.click(
      screen.getByRole("button", { name: "Commit semantic edit" }),
    );
    await waitFor(() =>
      expect(
        screen.queryByRole("textbox", { name: "Retained semantic draft" }),
      ).toBeNull(),
    );
    expect(commit).toHaveBeenLastCalledWith({
      draftValue: "invalid draft",
      row: { label: "Alpha", state: "open" },
      target: {
        fieldKey: "label",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 7 },
        rowIdentity: { kind: "core_record", recordId: "record-1" },
        surface: testSurface,
      },
    });
  });

  it("keeps a rejected inter-cell transition on its semantic draft and opens the destination only after acceptance", async () => {
    const commit = vi
      .fn()
      .mockResolvedValueOnce({
        kind: "validation_error",
        message: "Keep the original draft.",
      })
      .mockResolvedValueOnce({ kind: "accepted" });
    const transitionColumns: readonly GridColumn<HarnessRow>[] = [
      {
        contractWritable: true,
        editor: {
          commit,
          initialDraftValue: (row) => row.label,
          renderEditor: (context) => (
            <input
              aria-label={`${context.row.label} transition editor`}
              ref={context.focusTargetRef}
              value={String(context.draftValue)}
              onChange={(event) =>
                context.setDraftValue(event.currentTarget.value)
              }
            />
          ),
        },
        fieldKey: "label",
        label: "Label",
        renderCell: ({ row }) => (
          <span data-testid={`transition-${row.label}`}>{row.label}</span>
        ),
      },
    ];
    render(
      <SemanticDataGrid
        columns={transitionColumns}
        dataRows={[
          {
            data: { label: "Alpha", state: "open" },
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 4 },
            rowIdentity: { kind: "core_record", recordId: "record-1" },
          },
          {
            data: { label: "Beta", state: "open" },
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 9 },
            rowIdentity: { kind: "core_record", recordId: "record-2" },
          },
        ]}
        surface={testSurface}
      />,
    );

    const alphaCell = screen
      .getByTestId("transition-Alpha")
      .closest<HTMLElement>('[role="gridcell"]');
    const betaCell = screen
      .getByTestId("transition-Beta")
      .closest<HTMLElement>('[role="gridcell"]');
    if (alphaCell === null || betaCell === null) {
      throw new Error("Expected transition cells");
    }
    fireEvent.mouseDown(alphaCell);
    fireEvent.click(alphaCell);
    const alphaEditor = await screen.findByRole("textbox", {
      name: "Alpha transition editor",
    });
    fireEvent.change(alphaEditor, { target: { value: "unsaved exact draft" } });

    fireEvent.mouseDown(betaCell);
    fireEvent.click(betaCell);
    expect((await screen.findByRole("alert")).textContent).toContain(
      "Keep the original draft.",
    );
    expect(alphaEditor).toHaveProperty("value", "unsaved exact draft");
    expect(document.activeElement).toBe(alphaEditor);
    expect(commit).toHaveBeenCalledTimes(1);
    expect(
      screen.queryByRole("textbox", { name: "Beta transition editor" }),
    ).toBeNull();

    fireEvent.mouseDown(betaCell);
    fireEvent.click(betaCell);
    const betaEditor = await screen.findByRole("textbox", {
      name: "Beta transition editor",
    });
    expect(document.activeElement).toBe(betaEditor);
    expect(betaEditor).toHaveProperty("value", "Beta");
    expect(commit).toHaveBeenCalledTimes(2);
    expect(commit).toHaveBeenLastCalledWith({
      draftValue: "unsaved exact draft",
      row: { label: "Alpha", state: "open" },
      target: {
        fieldKey: "label",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 4 },
        rowIdentity: { kind: "core_record", recordId: "record-1" },
        surface: testSurface,
      },
    });
  });

  it("renders stable semantic row attributes, sort translation, and presentation-only group rows", async () => {
    const onSortChange = vi.fn();
    const groupedHandle = createRef<GridHandle>();
    const rows: readonly GridDataRow<HarnessRow>[] = [
      {
        gutterLabel: "1",
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-1" },
        data: {
          label: "Alpha",
          state: "open",
        },
        testId: "row-record-1",
      },
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-2" },
        data: {
          label: "Beta",
          state: "reviewed",
        },
        testId: "row-record-2",
      },
    ];

    render(
      <GridViewport testId="grid-shell">
        <SemanticDataGrid
          ref={groupedHandle}
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
          actionsColumn={{
            label: "Actions",
            renderCell: ({ rowIdentity }) => (
              <span>
                {rowIdentity.kind === "core_record"
                  ? rowIdentity.recordId
                  : "extension"}
              </span>
            ),
          }}
          columns={columns}
          grouping={{
            fieldKey: "state",
            formatLabel: (value) => (value === null ? null : String(value)),
            getTestId: (fieldKey, _value, label) =>
              label === null ? undefined : `group-${fieldKey}-${label}`,
            getValue: (row) => row.state,
          }}
          onSortChange={onSortChange}
          dataRows={rows}
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
    expect(gridShell.querySelectorAll(gridGroupRowSelector())).toHaveLength(2);
    expect(
      gridShell.querySelectorAll("[data-grid-record-id][aria-expanded]"),
    ).toHaveLength(0);
    const openGroupToggle = screen.getByTestId("group-state-open");
    expect(screen.getByRole("button", { name: "open" })).toBe(openGroupToggle);
    const openGroupRow = openGroupToggle.closest('[role="row"]');
    expect(openGroupRow).toBeTruthy();
    if (openGroupRow === null) {
      throw new Error("Expected open group toggle to have row ancestor");
    }
    expect(openGroupRow.getAttribute("data-grid-record-id")).toBeNull();
    expect(openGroupRow.matches(gridGroupRowSelector(true))).toBe(true);
    expect(openGroupRow.getAttribute("data-grid-row-kind")).toBeNull();
    expect(openGroupRow.getAttribute("data-grid-primary-state")).toBeNull();
    expect(openGroupRow.classList.contains("cartulary-grid-group-row")).toBe(
      false,
    );
    expect(
      openGroupRow.querySelectorAll("input, textarea, select"),
    ).toHaveLength(0);
    expect(openGroupRow.querySelectorAll("button")).toHaveLength(1);
    expect(openGroupToggle.getAttribute("aria-expanded")).toBe("true");
    fireEvent.click(openGroupToggle);
    expect(openGroupToggle.getAttribute("aria-expanded")).toBe("false");
    expect(openGroupRow.matches(gridGroupRowSelector(false))).toBe(true);
    expect(screen.queryByTestId("row-record-1")).toBeNull();
    expect(
      groupedHandle.current?.focusAnchor(gridAnchor("record-1", "label")),
    ).toBe(false);
    expect(
      groupedHandle.current?.moveFocus(gridAnchor("record-2", "label"), {
        key: "ArrowUp",
      }),
    ).toBeNull();
    fireEvent.click(openGroupToggle);
    expect(openGroupToggle.getAttribute("aria-expanded")).toBe("true");
    expect(screen.getByTestId("row-record-1")).toBeTruthy();
    expect(
      groupedHandle.current?.focusAnchor(gridAnchor("record-2", "label")),
    ).toBe(true);

    const labelHeader = screen.getByTestId("label-header");
    expect(labelHeader.getAttribute("data-grid-field-key")).toBe("label");
    fireEvent.click(labelHeader);
    expect(onSortChange).toHaveBeenCalledWith([
      { fieldKey: "label", direction: "desc" },
    ]);

    const stateHeader = screen.getByTestId("state-header");
    expect(stateHeader.getAttribute("title")).toBe(
      "State sorting disabled in this harness",
    );
    fireEvent.click(stateHeader);
    expect(onSortChange).toHaveBeenCalledTimes(1);
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
    const rows: readonly GridDataRow<GroupRow>[] = [
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "string" },
        data: { group: "1", label: "String" },
      },
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "number" },
        data: { group: 1, label: "Number" },
      },
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "boolean" },
        data: { group: true, label: "Boolean" },
      },
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "null" },
        data: { group: null, label: "Null" },
      },
    ];
    const grid = (fieldKey: string, recordRows = rows) => (
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "typed.view" }}
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
        dataRows={recordRows}
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
    const rows: readonly GridDataRow<HarnessRow>[] = [
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-1" },
        data: {
          label: "Alpha",
          state: "open",
        },
      },
    ];

    const { rerender } = render(
      <GridViewport testId="fill-grid-shell">
        <SemanticDataGrid
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
          actionsColumn={{
            label: "Actions",
            renderCell: ({ rowIdentity }) => (
              <span>
                {rowIdentity.kind === "core_record"
                  ? rowIdentity.recordId
                  : "extension"}
              </span>
            ),
          }}
          columns={columns}
          dataRows={rows}
        />
      </GridViewport>,
    );

    const gridShell = await screen.findByTestId("fill-grid-shell");
    const grid = gridShell.querySelector('[role="grid"]') as HTMLElement;
    rerender(
      <GridViewport testId="fill-grid-shell">
        <SemanticDataGrid
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
          actionsColumn={{
            label: "Actions",
            renderCell: ({ rowIdentity }) => (
              <span>
                {rowIdentity.kind === "core_record"
                  ? rowIdentity.recordId
                  : "extension"}
              </span>
            ),
          }}
          columns={columns}
          fillViewportInline
          dataRows={rows}
        />
      </GridViewport>,
    );

    expect(["0", "0px"]).toContain(grid.style.minWidth);
    expect(grid.style.width).toBe("100%");
  });

  it("renders the production DataGrid with bounded fixed-height row output", async () => {
    const rowCount = 500;
    const handle = createRef<GridHandle>();
    const rows: readonly GridDataRow<HarnessRow>[] = Array.from(
      { length: rowCount },
      (_, index) => ({
        kind: "data" as const,
        mutationIdentity: {
          kind: "core_row_version" as const,
          baseRowVersion: 1,
        },
        rowIdentity: {
          kind: "core_record" as const,
          recordId: `virtual-record-${index}`,
        },
        data: {
          label: `Record ${index}`,
          state: index % 2 === 0 ? "open" : "reviewed",
        },
      }),
    );
    const wideColumns: readonly GridColumn<HarnessRow>[] = Array.from(
      { length: 24 },
      (_, index) => ({
        fieldKey: `wide-${index}`,
        label: `Wide ${index}`,
        renderCell: ({ row }) => `${row.label} ${index}`,
        width: 180,
      }),
    );

    const { rerender } = render(
      <GridViewport testId="virtualized-grid-shell">
        <SemanticDataGrid
          key="ungrouped-virtualized-grid"
          ref={handle}
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
          columns={wideColumns}
          dataRows={rows}
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
    const mountedHeaders = shell.querySelectorAll('[role="columnheader"]');
    expect(mountedHeaders.length).toBeGreaterThan(0);
    expect(mountedHeaders.length).toBeLessThan(wideColumns.length);

    const lastAnchor = gridAnchor("virtual-record-499", "wide-23");
    expect(handle.current?.isAnchorRendered(lastAnchor)).toBe(false);
    expect(handle.current?.scrollToAnchor(lastAnchor)).toBe(true);
    expect(handle.current?.isAnchorRendered(lastAnchor)).toBe(false);

    rerender(
      <GridViewport testId="virtualized-grid-shell">
        <SemanticDataGrid
          key="grouped-virtualized-grid"
          ref={handle}
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
          columns={wideColumns}
          grouping={{
            fieldKey: "state",
            formatLabel: (value) => (value === null ? null : String(value)),
            getTestId: (fieldKey, _value, label) =>
              label === null ? undefined : `virtual-group-${fieldKey}-${label}`,
            getValue: (row) => row.state,
          }}
          dataRows={rows}
        />
      </GridViewport>,
    );

    expect(shell.querySelector('[role="treegrid"]')).toBeTruthy();
    expect(screen.getByTestId("virtual-group-state-open")).toBeTruthy();
    expect(
      shell.querySelectorAll('[role="row"][data-grid-record-id]').length,
    ).toBeLessThan(rowCount);
    expect(handle.current?.scrollToAnchor(lastAnchor)).toBe(true);

    render(
      <GridViewport testId="dom-unit-grid-shell">
        <SemanticDataGridDomUnit
          surface={{ kind: "view_schema", viewSchemaId: "test.dom-unit" }}
          columns={wideColumns}
          dataRows={rows.slice(0, 50)}
        />
      </GridViewport>,
    );
    const domUnitShell = screen.getByTestId("dom-unit-grid-shell");
    expect(
      domUnitShell.querySelectorAll('[role="row"][data-grid-record-id]'),
    ).toHaveLength(50);
    expect(
      shell.querySelectorAll('[role="row"][data-grid-record-id]').length,
    ).toBeLessThan(rowCount);
  });

  it("keeps standalone block sizing by default and supports shell-owned fill block sizing", async () => {
    const rows: readonly GridDataRow<HarnessRow>[] = [
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-1" },
        data: {
          label: "Alpha",
          state: "open",
        },
      },
    ];

    const { rerender } = render(
      <GridViewport testId="block-sizing-grid-shell">
        <SemanticDataGrid
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
          columns={columns}
          dataRows={rows}
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
        <SemanticDataGrid
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
          columns={columns}
          dataRows={rows}
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
    const rows: readonly GridDataRow<HarnessRow>[] = [
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-1" },
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
        <SemanticDataGrid
          surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
          actionsColumn={{
            label: "Actions",
            minWidth: 64,
            renderCell: ({ rowIdentity }) => (
              <span>
                {rowIdentity.kind === "core_record"
                  ? rowIdentity.recordId
                  : "extension"}
              </span>
            ),
            width: 96,
          }}
          columns={sizedColumns}
          dataRows={rows}
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
    const rows: readonly GridDataRow<HarnessRow>[] = [
      {
        kind: "data",
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        rowIdentity: { kind: "core_record", recordId: "record-1" },
        data: {
          label: "Alpha",
          state: "open",
        },
      },
    ];

    function DensityGrid({ density }: { density?: GridDensity }) {
      return (
        <GridViewport testId="density-grid-shell">
          <SemanticDataGrid
            actionsColumn={{
              label: "Actions",
              renderCell: () => <button type="button">Inspect</button>,
            }}
            coreRecordBulkSelection={{
              onSelectedRecordIdsChange: vi.fn(),
              selectedRecordIds: new Set(),
            }}
            surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
            columns={columns}
            density={density}
            dataRows={rows}
            rowGutter={{ label: "Row", width: 48 }}
          />
        </GridViewport>
      );
    }

    const { rerender } = render(<DensityGrid />);

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
    expect(
      grid.style.getPropertyValue("--cartulary-grid-cell-padding-block"),
    ).toBe("3px");
    expect(
      grid.style.getPropertyValue("--cartulary-grid-cell-padding-inline"),
    ).toBe("7px");
    expect(grid.style.getPropertyValue("--cartulary-grid-font-size")).toBe(
      "var(--ct-density-default-fontSize)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-line-height")).toBe(
      "var(--ct-density-default-lineHeight)",
    );
    expect(
      grid.querySelector('[role="columnheader"]')?.getAttribute("style"),
    ).toContain("padding-block-start");
    expect(
      grid.querySelectorAll(".cartulary-grid-header-content"),
    ).toHaveLength(5);
    expect(
      grid.querySelector(".cartulary-grid-selection-content"),
    ).toBeTruthy();
    expect(grid.querySelector(".cartulary-grid-gutter-content")).toBeTruthy();
    expect(grid.querySelector(".cartulary-grid-actions-content")).toBeTruthy();

    rerender(<DensityGrid density="compact" />);

    expect(grid.style.getPropertyValue("--cartulary-grid-density")).toBe(
      "compact",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-row-height")).toBe(
      "var(--ct-density-compact-rowHeight)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-cell-padding")).toBe(
      "var(--ct-density-compact-cellPadding)",
    );
    expect(
      grid.style.getPropertyValue("--cartulary-grid-cell-padding-block"),
    ).toBe("2px");
    expect(
      grid.style.getPropertyValue("--cartulary-grid-cell-padding-inline"),
    ).toBe("5px");
    expect(grid.style.getPropertyValue("--cartulary-grid-font-size")).toBe(
      "var(--ct-density-compact-fontSize)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-line-height")).toBe(
      "var(--ct-density-compact-lineHeight)",
    );

    rerender(<DensityGrid density="comfortable" />);

    expect(grid.style.getPropertyValue("--cartulary-grid-density")).toBe(
      "comfortable",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-row-height")).toBe(
      "var(--ct-density-comfortable-rowHeight)",
    );
    expect(grid.style.getPropertyValue("--cartulary-grid-cell-padding")).toBe(
      "var(--ct-density-comfortable-cellPadding)",
    );
    expect(
      grid.style.getPropertyValue("--cartulary-grid-cell-padding-block"),
    ).toBe("5px");
    expect(
      grid.style.getPropertyValue("--cartulary-grid-cell-padding-inline"),
    ).toBe("9px");
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
      const editableRows = useMemo<readonly GridDataRow<HarnessRow>[]>(
        () => [
          {
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
            rowIdentity: { kind: "core_record", recordId: "record-1" },
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
          renderCell: (row: GridDataRow<HarnessRow>) => (
            <span data-testid="row-action">
              {row.rowIdentity.kind === "core_record"
                ? row.rowIdentity.recordId
                : "extension"}
            </span>
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
          <SemanticDataGrid
            surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
            actionsColumn={actionsColumn}
            columns={editableColumns}
            dataRows={editableRows}
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
      const gridRows = useMemo<readonly GridDataRow<EditableHarnessRow>[]>(
        () =>
          rows.map((row) => ({
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
            rowIdentity: { kind: "core_record", recordId: row.recordId },
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
          <SemanticDataGrid
            surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
            columns={editableColumns}
            onSortChange={(nextSort) => {
              setSort(nextSort);
              setRows((current) =>
                [...current].sort((left, right) =>
                  right.label.localeCompare(left.label),
                ),
              );
            }}
            dataRows={gridRows}
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
      const gridRows = useMemo<readonly GridDataRow<EditableHarnessRow>[]>(
        () =>
          rows.map((row) => ({
            kind: "data",
            mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
            rowIdentity: { kind: "core_record", recordId: row.recordId },
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
          <SemanticDataGrid
            surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
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
            dataRows={gridRows}
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
          renderCell: (row: GridDataRow<HarnessRow>) =>
            row.rowIdentity.kind === "core_record"
              ? row.rowIdentity.recordId
              : "extension",
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
          <SemanticDataGrid
            surface={{ kind: "view_schema", viewSchemaId: "test.view" }}
            actionsColumn={actionsColumn}
            columns={columns}
            draftRow={draftRow}
            dataRows={[]}
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
