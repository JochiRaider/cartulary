import type {
  GridColumn,
  GridDataRow,
  GridDataState,
  GridHandle,
} from "@cartulary/grid-adapter";
import { render, waitFor } from "@testing-library/react";
import { forwardRef, useImperativeHandle, useRef } from "react";
import { describe, expect, it, vi } from "vitest";
import type { WorkbookGridEntryFocusOwner } from "../models/workbookGridEntryFocus";
import { useWorkbookSemanticGridFocus } from "./useWorkbookSemanticGridFocus";

type Row = { readonly id: string };

const readyDataState: GridDataState = { kind: "ready" };
const visibleColumns: readonly GridColumn<Row>[] = [
  { fieldKey: "first", label: "First", renderCell: () => null },
  { fieldKey: "second", label: "Second", renderCell: () => null },
];
const dataRows: readonly GridDataRow<Row>[] = [
  {
    data: { id: "row-1" },
    kind: "data",
    rowIdentity: { kind: "core_record", recordId: "row-1" },
  },
  {
    data: { id: "row-2" },
    kind: "data",
    rowIdentity: { kind: "core_record", recordId: "row-2" },
  },
];

function createGridHandle(): GridHandle {
  return {
    activateEdit: vi.fn(() => false),
    cancelEdit: vi.fn(() => false),
    focusAnchor: vi.fn(() => false),
    focusDraftCell: vi.fn(() => false),
    focusRoot: vi.fn(() => true),
    getAnchorRect: vi.fn(() => null),
    getScrollElement: vi.fn(() => null),
    isAnchorRendered: vi.fn(() => false),
    moveFocus: vi.fn(() => null),
    planPasteTargets: vi.fn(() => null),
    scrollToAnchor: vi.fn(() => false),
  };
}

function pendingFocus(
  generation = 1,
  viewSchemaId = "cartulary.view.hosts.v1",
): WorkbookGridEntryFocusOwner {
  return {
    acknowledge: vi.fn(),
    request: { generation, kind: "pending", viewSchemaId },
  };
}

const GridHandleRegistration = forwardRef<
  GridHandle,
  { readonly handle: GridHandle }
>(function GridHandleRegistration({ handle }, ref) {
  useImperativeHandle(ref, () => handle, [handle]);
  return null;
});

function SemanticFocusHarness({
  dataState = readyDataState,
  draftFieldKeys,
  focusOwner,
  handle,
  rows = dataRows,
  columns = visibleColumns,
  viewSchemaId = "cartulary.view.hosts.v1",
}: {
  readonly dataState?: GridDataState | undefined;
  readonly draftFieldKeys?: readonly string[] | undefined;
  readonly focusOwner: WorkbookGridEntryFocusOwner;
  readonly handle: GridHandle | null;
  readonly rows?: readonly GridDataRow<Row>[] | undefined;
  readonly columns?: readonly GridColumn<Row>[] | undefined;
  readonly viewSchemaId?: string | undefined;
}) {
  const handleRef = useRef<GridHandle | null>(null);
  const registerGridHandle = useWorkbookSemanticGridFocus({
    dataRows: rows,
    dataState,
    draftFieldKeys,
    focusOwner,
    gridHandleRef: handleRef,
    visibleColumns: columns,
    viewSchemaId,
  });
  return handle === null ? null : (
    <GridHandleRegistration handle={handle} ref={registerGridHandle} />
  );
}

describe("useWorkbookSemanticGridFocus", () => {
  it("focuses the first registered writable draft field and acknowledges once", async () => {
    const handle = createGridHandle();
    vi.mocked(handle.focusDraftCell).mockImplementation(
      (fieldKey) => fieldKey === "second",
    );
    const focusOwner = pendingFocus();

    render(
      <SemanticFocusHarness
        draftFieldKeys={["first", "second"]}
        focusOwner={focusOwner}
        handle={handle}
      />,
    );

    await waitFor(() =>
      expect(focusOwner.acknowledge).toHaveBeenCalledWith({
        generation: 1,
        viewSchemaId: "cartulary.view.hosts.v1",
      }),
    );
    expect(handle.focusDraftCell).toHaveBeenNthCalledWith(1, "first");
    expect(handle.focusDraftCell).toHaveBeenNthCalledWith(2, "second");
    expect(handle.focusAnchor).not.toHaveBeenCalled();
    expect(focusOwner.acknowledge).toHaveBeenCalledTimes(1);
  });

  it("tries committed rows and visible fields in semantic order", async () => {
    const handle = createGridHandle();
    vi.mocked(handle.focusAnchor).mockImplementation(
      (anchor) =>
        anchor.rowIdentity.kind === "core_record" &&
        anchor.rowIdentity.recordId === "row-2" &&
        anchor.fieldKey === "first",
    );
    const focusOwner = pendingFocus();

    render(<SemanticFocusHarness focusOwner={focusOwner} handle={handle} />);

    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
    expect(handle.focusAnchor).toHaveBeenCalledTimes(3);
    expect(handle.focusAnchor).toHaveBeenNthCalledWith(1, {
      fieldKey: "first",
      rowIdentity: { kind: "core_record", recordId: "row-1" },
      surface: {
        kind: "view_schema",
        viewSchemaId: "cartulary.view.hosts.v1",
      },
    });
    expect(handle.focusAnchor).toHaveBeenNthCalledWith(3, {
      fieldKey: "first",
      rowIdentity: { kind: "core_record", recordId: "row-2" },
      surface: {
        kind: "view_schema",
        viewSchemaId: "cartulary.view.hosts.v1",
      },
    });
    expect(handle.focusRoot).not.toHaveBeenCalled();
  });

  it.each([
    ["empty rows", [], visibleColumns],
    ["no visible fields", dataRows, []],
    ["presentation-ineligible grouped rows", dataRows, visibleColumns],
  ])("falls back to the root for %s", async (_case, rows, columns) => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();

    render(
      <SemanticFocusHarness
        columns={columns}
        focusOwner={focusOwner}
        handle={handle}
        rows={rows}
      />,
    );

    await waitFor(() => expect(handle.focusRoot).toHaveBeenCalledOnce());
    expect(focusOwner.acknowledge).toHaveBeenCalledOnce();
  });

  it("waits through busy data and retries from the ready state", async () => {
    const handle = createGridHandle();
    vi.mocked(handle.focusAnchor).mockReturnValue(true);
    const focusOwner = pendingFocus();
    const { rerender } = render(
      <SemanticFocusHarness
        dataState={{
          generationKey: "load-1",
          kind: "initial_loading",
          surfaceLabel: "Hosts",
        }}
        focusOwner={focusOwner}
        handle={handle}
      />,
    );
    expect(handle.focusAnchor).not.toHaveBeenCalled();

    rerender(
      <SemanticFocusHarness
        dataState={{ kind: "refreshing", surfaceLabel: "Hosts" }}
        focusOwner={focusOwner}
        handle={handle}
      />,
    );
    expect(handle.focusAnchor).not.toHaveBeenCalled();

    rerender(<SemanticFocusHarness focusOwner={focusOwner} handle={handle} />);
    await waitFor(() => expect(handle.focusAnchor).toHaveBeenCalledOnce());
    expect(focusOwner.acknowledge).toHaveBeenCalledOnce();
  });

  it("ignores a wrong surface and retries when the matching surface mounts", async () => {
    const handle = createGridHandle();
    vi.mocked(handle.focusAnchor).mockReturnValue(true);
    const focusOwner = pendingFocus();
    const { rerender } = render(
      <SemanticFocusHarness
        focusOwner={focusOwner}
        handle={handle}
        viewSchemaId="cartulary.view.identities.v1"
      />,
    );
    expect(handle.focusAnchor).not.toHaveBeenCalled();

    rerender(<SemanticFocusHarness focusOwner={focusOwner} handle={handle} />);
    await waitFor(() => expect(handle.focusAnchor).toHaveBeenCalledOnce());
    expect(focusOwner.acknowledge).toHaveBeenCalledWith({
      generation: 1,
      viewSchemaId: "cartulary.view.hosts.v1",
    });
  });

  it("uses only the newest generation when mounting follows a rapid switch", async () => {
    const handle = createGridHandle();
    vi.mocked(handle.focusAnchor).mockReturnValue(true);
    const staleOwner = pendingFocus(1);
    const currentOwner = pendingFocus(2);
    const { rerender } = render(
      <SemanticFocusHarness focusOwner={staleOwner} handle={null} />,
    );

    rerender(
      <SemanticFocusHarness focusOwner={currentOwner} handle={handle} />,
    );
    await waitFor(() =>
      expect(currentOwner.acknowledge).toHaveBeenCalledOnce(),
    );
    expect(staleOwner.acknowledge).not.toHaveBeenCalled();
    expect(currentOwner.acknowledge).toHaveBeenCalledWith({
      generation: 2,
      viewSchemaId: "cartulary.view.hosts.v1",
    });
  });

  it("does nothing after unmount and has no deferred retry", () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();
    const rendered = render(
      <SemanticFocusHarness
        dataState={{
          generationKey: "load-1",
          kind: "initial_loading",
          surfaceLabel: "Hosts",
        }}
        focusOwner={focusOwner}
        handle={handle}
      />,
    );

    rendered.unmount();
    expect(handle.focusAnchor).not.toHaveBeenCalled();
    expect(handle.focusRoot).not.toHaveBeenCalled();
    expect(focusOwner.acknowledge).not.toHaveBeenCalled();
  });

  it("treats a virtualized offscreen anchor as successful through GridHandle", async () => {
    const handle = createGridHandle();
    vi.mocked(handle.focusAnchor).mockReturnValueOnce(true);
    const focusOwner = pendingFocus();

    render(<SemanticFocusHarness focusOwner={focusOwner} handle={handle} />);

    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
    expect(handle.isAnchorRendered).not.toHaveBeenCalled();
    expect(handle.scrollToAnchor).not.toHaveBeenCalled();
    expect(handle.focusRoot).not.toHaveBeenCalled();
  });

  it("keeps the request pending when every focus command fails", async () => {
    const handle = createGridHandle();
    vi.mocked(handle.focusRoot).mockReturnValue(false);
    const focusOwner = pendingFocus();

    render(<SemanticFocusHarness focusOwner={focusOwner} handle={handle} />);

    await waitFor(() => expect(handle.focusRoot).toHaveBeenCalledOnce());
    expect(focusOwner.acknowledge).not.toHaveBeenCalled();
  });
});
