import type {
  GridColumn,
  GridDataRow,
  GridDataState,
  GridFocusResult,
  GridHandle,
} from "@cartulary/grid-adapter";
import { act, fireEvent, waitFor } from "@testing-library/react";
import { forwardRef, useImperativeHandle, useRef } from "react";
import { describe, expect, it, vi } from "vitest";
import { renderWithWorkbookQueryBrowsing as render } from "../../testing/workbookQueryTestSupport";
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
  const root = document.createElement("div");
  return {
    activateEdit: vi.fn(() => false),
    cancelEdit: vi.fn(() => false),
    requestFocus: vi.fn(async (target) =>
      target.kind === "root" ? ("focused" as const) : ("unavailable" as const),
    ),
    getAnchorRect: vi.fn(() => null),
    getScrollElement: vi.fn(() => root),
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
    cancel: vi.fn(),
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
    vi.mocked(handle.requestFocus).mockImplementation(async (target) =>
      target.kind === "draft" && target.fieldKey === "second"
        ? "focused"
        : "unavailable",
    );
    const focusOwner = pendingFocus();
    render(
      <SemanticFocusHarness
        handle={handle}
        focusOwner={focusOwner}
        draftFieldKeys={["first", "second"]}
      />,
    );
    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
    expect(
      vi.mocked(handle.requestFocus).mock.calls.map(([target]) => target),
    ).toEqual([
      { kind: "draft", fieldKey: "first" },
      { kind: "draft", fieldKey: "second" },
    ]);
  });
  it("tries committed rows and visible fields in semantic order", async () => {
    const handle = createGridHandle();
    vi.mocked(handle.requestFocus).mockImplementation(async (target) =>
      target.kind === "cell" &&
      target.anchor.rowIdentity.kind === "core_record" &&
      target.anchor.rowIdentity.recordId === "row-2"
        ? "focused"
        : "unavailable",
    );
    const focusOwner = pendingFocus();
    render(<SemanticFocusHarness handle={handle} focusOwner={focusOwner} />);
    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
    expect(handle.requestFocus).toHaveBeenCalledTimes(3);
    expect(vi.mocked(handle.requestFocus).mock.calls[2]?.[0]).toMatchObject({
      kind: "cell",
      anchor: { fieldKey: "first", rowIdentity: { recordId: "row-2" } },
    });
  });
  it("waits for completed draft focus without acknowledging or falling back", async () => {
    const handle = createGridHandle();
    let complete!: (value: GridFocusResult) => void;
    vi.mocked(handle.requestFocus).mockImplementation(
      () =>
        new Promise((resolve) => {
          complete = resolve;
        }),
    );
    const focusOwner = pendingFocus();
    render(
      <SemanticFocusHarness
        handle={handle}
        focusOwner={focusOwner}
        draftFieldKeys={["first"]}
      />,
    );
    await waitFor(() => expect(handle.requestFocus).toHaveBeenCalledOnce());
    expect(focusOwner.acknowledge).not.toHaveBeenCalled();
    await act(async () => complete("focused"));
    expect(focusOwner.acknowledge).toHaveBeenCalledOnce();
  });
  it("falls back to the root for empty rows", async () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();
    render(
      <SemanticFocusHarness
        handle={handle}
        focusOwner={focusOwner}
        rows={[]}
      />,
    );
    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
    expect(handle.requestFocus).toHaveBeenCalledWith(
      { kind: "root" },
      expect.anything(),
    );
  });
  it("falls back to the root for no visible fields", async () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();
    render(
      <SemanticFocusHarness
        handle={handle}
        focusOwner={focusOwner}
        columns={[]}
      />,
    );
    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
    expect(handle.requestFocus).toHaveBeenCalledWith(
      { kind: "root" },
      expect.anything(),
    );
  });
  it("falls back to the root for presentation-ineligible grouped rows", async () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();
    render(<SemanticFocusHarness handle={handle} focusOwner={focusOwner} />);
    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
    expect(vi.mocked(handle.requestFocus).mock.calls.at(-1)?.[0]).toEqual({
      kind: "root",
    });
  });
  it("waits through busy data and retries from the ready state", async () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();
    const view = render(
      <SemanticFocusHarness
        handle={handle}
        focusOwner={focusOwner}
        dataState={{ kind: "refreshing", surfaceLabel: "Hosts" }}
      />,
    );
    expect(handle.requestFocus).not.toHaveBeenCalled();
    view.rerender(
      <SemanticFocusHarness handle={handle} focusOwner={focusOwner} />,
    );
    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
  });
  it("ignores a wrong surface and retries when the matching surface mounts", async () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus(1, "other");
    const view = render(
      <SemanticFocusHarness handle={handle} focusOwner={focusOwner} />,
    );
    expect(handle.requestFocus).not.toHaveBeenCalled();
    view.rerender(
      <SemanticFocusHarness
        handle={handle}
        focusOwner={focusOwner}
        viewSchemaId="other"
      />,
    );
    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
  });
  it("uses only the newest generation when mounting follows a rapid switch", async () => {
    const old = pendingFocus(1);
    const next = pendingFocus(2);
    const handle = createGridHandle();
    const view = render(
      <SemanticFocusHarness handle={null} focusOwner={old} />,
    );
    view.rerender(<SemanticFocusHarness handle={handle} focusOwner={next} />);
    await waitFor(() =>
      expect(next.acknowledge).toHaveBeenCalledWith({
        generation: 2,
        viewSchemaId:
          next.request.kind === "pending" ? next.request.viewSchemaId : "",
      }),
    );
    expect(old.acknowledge).not.toHaveBeenCalled();
  });
  it("does nothing after unmount and has no deferred retry", async () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();
    let complete!: (value: GridFocusResult) => void;
    vi.mocked(handle.requestFocus).mockImplementation(
      () =>
        new Promise((resolve) => {
          complete = resolve;
        }),
    );
    const view = render(
      <SemanticFocusHarness handle={handle} focusOwner={focusOwner} />,
    );
    await waitFor(() => expect(handle.requestFocus).toHaveBeenCalledOnce());
    const signal = vi.mocked(handle.requestFocus).mock.calls[0]?.[1]?.signal;
    view.unmount();
    expect(signal?.aborted).toBe(true);
    await act(async () => complete("focused"));
    expect(focusOwner.acknowledge).not.toHaveBeenCalled();
  });
  it("cancels pending entry after deliberate keyboard navigation", async () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();
    vi.mocked(handle.requestFocus).mockImplementation(
      () => new Promise(() => {}),
    );
    render(<SemanticFocusHarness handle={handle} focusOwner={focusOwner} />);
    await waitFor(() => expect(handle.requestFocus).toHaveBeenCalledOnce());
    fireEvent.keyDown(document, { key: "Tab" });
    expect(
      vi.mocked(handle.requestFocus).mock.calls[0]?.[1]?.signal?.aborted,
    ).toBe(true);
    expect(focusOwner.acknowledge).not.toHaveBeenCalled();
    expect(focusOwner.cancel).toHaveBeenCalledWith(focusOwner.request);
  });
  it("keeps the request pending when every focus command fails", async () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();
    vi.mocked(handle.requestFocus).mockResolvedValue("unavailable");
    render(<SemanticFocusHarness handle={handle} focusOwner={focusOwner} />);
    await waitFor(() => expect(handle.requestFocus).toHaveBeenCalledTimes(5));
    expect(focusOwner.acknowledge).not.toHaveBeenCalled();
  });
  it("cancels user navigation while loading and cancels entry after authority loss", async () => {
    const handle = createGridHandle();
    const focusOwner = pendingFocus();
    const view = render(
      <SemanticFocusHarness
        handle={handle}
        focusOwner={focusOwner}
        dataState={{ kind: "refreshing", surfaceLabel: "Hosts" }}
      />,
    );
    fireEvent.keyDown(document, { key: "Tab" });
    expect(focusOwner.cancel).toHaveBeenCalledWith(focusOwner.request);
    expect(handle.requestFocus).not.toHaveBeenCalled();
    view.rerender(
      <SemanticFocusHarness
        handle={handle}
        focusOwner={focusOwner}
        dataState={{ kind: "permission_denied", message: "Access lost" }}
      />,
    );
    expect(focusOwner.cancel).toHaveBeenCalledTimes(2);
    expect(focusOwner.acknowledge).not.toHaveBeenCalled();
  });
  it("replaces a handle without acknowledging its late focus completion", async () => {
    const first = createGridHandle();
    const second = createGridHandle();
    let complete!: (result: GridFocusResult) => void;
    vi.mocked(first.requestFocus).mockImplementation(
      () =>
        new Promise((resolve) => {
          complete = resolve;
        }),
    );
    const focusOwner = pendingFocus();
    const view = render(
      <SemanticFocusHarness handle={first} focusOwner={focusOwner} />,
    );
    await waitFor(() => expect(first.requestFocus).toHaveBeenCalledOnce());
    const signal = vi.mocked(first.requestFocus).mock.calls[0]?.[1]?.signal;
    view.rerender(
      <SemanticFocusHarness handle={second} focusOwner={focusOwner} />,
    );
    expect(signal?.aborted).toBe(true);
    await waitFor(() => expect(focusOwner.acknowledge).toHaveBeenCalledOnce());
    await act(async () => complete("focused"));
    expect(focusOwner.acknowledge).toHaveBeenCalledOnce();
  });
});
